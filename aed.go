package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BackMsg struct{}

type refreshFinishedMsg struct {
	newNode *FileNode
	err     error
}

type scanFinishedMsg struct {
	root     *FileNode
	diskSize int64
	err      error
}

type fileID struct {
	dev uint64
	ino uint64
}

type FileNode struct {
	Name     string
	Path     string
	Size     int64
	IsDir    bool
	Children []*FileNode
	Parent   *FileNode
}

type SessionState int

const (
	StateInputPath SessionState = iota
	StateScanning
	StateBrowsing
)

// --- STYLES ---

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4")).Bold(true)
	pathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff")).Bold(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))

	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Bold(true)
	barFull       = lipgloss.NewStyle().Foreground(lipgloss.Color("#9900ff"))
	barEmpty      = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	countStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff")).Bold(true).PaddingLeft(2)
)

type Model struct {
	state     SessionState
	textInput textinput.Model
	spinner   spinner.Model

	root        *FileNode
	currentNode *FileNode
	cursor      int
	yOffset     int

	filesScanned  *int64
	showHelp      bool
	diskTotalSize int64

	width, height int
	err           error
}

func New(w, h int) Model {
	ti := textinput.New()
	ti.Placeholder = "/home/user (ou ~)"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.SetValue(".")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D"))

	zero := int64(0)

	return Model{
		state:        StateInputPath,
		textInput:    ti,
		spinner:      s,
		filesScanned: &zero,
		width:        w,
		height:       h,
		showHelp:     true,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// --- LOGIQUE D'AFFICHAGE ---

func (m Model) getDisplayItems() []*FileNode {
	var items []*FileNode

	if m.currentNode == nil {
		return items
	}

	dot := &FileNode{
		Name:  ".",
		Path:  m.currentNode.Path,
		Size:  m.currentNode.Size,
		IsDir: true,
	}
	items = append(items, dot)

	if m.currentNode.Parent != nil {
		parentPath := filepath.Dir(m.currentNode.Path)
		dotdot := &FileNode{
			Name:  "..",
			Path:  parentPath,
			Size:  0,
			IsDir: true,
		}
		items = append(items, dotdot)
	}

	items = append(items, m.currentNode.Children...)

	return items
}

// --- UPDATE ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case BackMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, func() tea.Msg { return BackMsg{} }
		}

		// --- ETAT: SAISIE DU CHEMIN ---
		if m.state == StateInputPath {
			switch msg.String() {
			case "enter":
				rawInput := m.textInput.Value()

				// 1. Gestion du raccourci "."
				path := rawInput
				if path == "." {
					path, _ = os.Getwd()
				}

				// 2. Gestion du raccourci "~" (NOUVEAU)
				path = expandPath(path)

				m.state = StateScanning
				atomic.StoreInt64(m.filesScanned, 0)
				visitedInodes := make(map[fileID]struct{})

				return m, tea.Batch(
					m.spinner.Tick,
					scanDirectoryCmd(path, m.filesScanned, visitedInodes),
				)
			case "esc", "q":
				return m, func() tea.Msg { return BackMsg{} }
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// --- ETAT: SCAN EN COURS ---
		if m.state == StateScanning {
			if msg.String() == "q" || msg.String() == "esc" {
				return m, func() tea.Msg { return BackMsg{} }
			}
		}

		// --- ETAT: NAVIGATION ---
		if m.state == StateBrowsing {
			items := m.getDisplayItems()

			switch msg.String() {

			case "q":
				return m, func() tea.Msg { return BackMsg{} }

			case ",", "?":
				m.showHelp = !m.showHelp
				return m, nil

			case "g":
				if len(items) > 0 && m.cursor < len(items) {
					selected := items[m.cursor]
					cmd := exec.Command("xdg-open", selected.Path)
					cmd.Start()
				}
				return m, nil

			case "b":
				if len(items) > 0 && m.cursor < len(items) {
					selected := items[m.cursor]
					targetPath := selected.Path
					if !selected.IsDir {
						targetPath = filepath.Dir(selected.Path)
					}
					shell := os.Getenv("SHELL")
					if shell == "" {
						shell = "/bin/bash"
					}
					c := exec.Command(shell)
					c.Dir = targetPath
					return m, tea.ExecProcess(c, func(err error) tea.Msg { return nil })
				}
				return m, nil

			case "r":
				if m.currentNode != nil {
					m.state = StateScanning
					atomic.StoreInt64(m.filesScanned, 0)
					visitedInodes := make(map[fileID]struct{})

					return m, tea.Batch(
						m.spinner.Tick,
						refreshDirectoryCmd(m.currentNode.Path, m.filesScanned, visitedInodes),
					)
				}
				return m, nil

			case "backspace", "left", "h", "esc":
				if m.currentNode.Parent != nil {
					m.currentNode = m.currentNode.Parent
					m.cursor = 0
					m.yOffset = 0
				}

			case "enter", "right", "l":
				if len(items) > 0 && m.cursor < len(items) {
					selected := items[m.cursor]
					if selected.Name == ".." {
						if m.currentNode.Parent != nil {
							m.currentNode = m.currentNode.Parent
							m.cursor = 0
							m.yOffset = 0
						}
						return m, nil
					}
					if selected.Name == "." {
						return m, nil
					}
					if selected.IsDir {
						for _, child := range m.currentNode.Children {
							if child.Path == selected.Path {
								m.currentNode = child
								m.cursor = 0
								m.yOffset = 0
								break
							}
						}
					}
				}

			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					if m.cursor < m.yOffset {
						m.yOffset = m.cursor
					}
				}
			case "down", "j":
				if m.cursor < len(items)-1 {
					m.cursor++
					footerHeight := 2
					if !m.showHelp {
						footerHeight = 0
					}
					visibleHeight := m.height - 5 - footerHeight
					if m.cursor >= m.yOffset+visibleHeight {
						m.yOffset = m.cursor - visibleHeight + 1
					}
				}
			}
		}

	case scanFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = StateInputPath
		} else {
			m.root = msg.root
			m.currentNode = msg.root
			m.diskTotalSize = msg.diskSize
			m.state = StateBrowsing
		}

	case refreshFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = StateBrowsing
		} else {
			newNode := msg.newNode
			oldNode := m.currentNode

			if oldNode.Parent != nil {
				newNode.Parent = oldNode.Parent
				newNode.Name = filepath.Base(newNode.Path)

				for i, child := range oldNode.Parent.Children {
					if child.Path == oldNode.Path {
						oldNode.Parent.Children[i] = newNode
						break
					}
				}

				diff := newNode.Size - oldNode.Size
				parent := newNode.Parent
				for parent != nil {
					parent.Size += diff
					parent = parent.Parent
				}
			} else {
				m.root = newNode
			}

			m.currentNode = newNode

			if m.cursor >= len(m.getDisplayItems()) {
				m.cursor = 0
				m.yOffset = 0
			}

			m.state = StateBrowsing
		}

	case spinner.TickMsg:
		if m.state == StateScanning {
			var cmdSpinner tea.Cmd
			m.spinner, cmdSpinner = m.spinner.Update(msg)
			return m, cmdSpinner
		}
	}

	return m, nil
}

// --- VIEW ---

func (m Model) View() string {
	if m.state == StateInputPath {
		title := titleStyle.Render("AED - Analyseur d'Espace Disque")
		input := m.textInput.View()
		return fmt.Sprintf("\n  %s\n\n  Entrez le dossier à analyser :\n  %s\n\n  %s", title, input, helpStyle.Render("(enter: valider • esc: quitter)"))
	}

	if m.state == StateScanning {
		count := atomic.LoadInt64(m.filesScanned)
		return fmt.Sprintf(
			"\n  %s Analyse en cours...\n\n%s fichiers scannés\n\n  %s",
			m.spinner.View(),
			countStyle.Render(fmt.Sprintf("%d", count)),
			helpStyle.Render("Appuyer sur q pour quitter"),
		)
	}

	if m.state == StateBrowsing {
		if m.currentNode == nil {
			return "Erreur: Node vide"
		}

		title := titleStyle.Render("AED")
		path := pathStyle.Render(m.currentNode.Path)
		totalSize := infoStyle.Render(fmt.Sprintf("Total: %s", formatBytes(m.currentNode.Size)))

		var diskSizeStr string
		if m.diskTotalSize > 0 {
			diskSizeStr = infoStyle.Render(fmt.Sprintf("Disque: %s", formatBytes(m.diskTotalSize)))
		}

		header := fmt.Sprintf("  %s  %s  (%s)  (%s)\n", title, path, totalSize, diskSizeStr)

		footerHeight := 2
		if !m.showHelp {
			footerHeight = 0
		}

		visibleHeight := m.height - 5 - footerHeight
		if visibleHeight < 1 {
			visibleHeight = 1
		}

		var rows []string
		items := m.getDisplayItems()

		start := m.yOffset
		end := start + visibleHeight
		if end > len(items) {
			end = len(items)
		}

		barWidth := 20

		for i := start; i < end; i++ {
			item := items[i]

			var sizeStr, bar, name string

			if item.Name == "." || item.Name == ".." {
				sizeStr = fmt.Sprintf("%8s", "")
				if item.Name == "." {
					sizeStr = fmt.Sprintf("%8s", formatBytes(item.Size))
				}
				bar = strings.Repeat(" ", barWidth)
				name = item.Name
			} else {
				sizeStr = fmt.Sprintf("%8s", formatBytes(item.Size))

				percent := 0.0
				if m.currentNode.Size > 0 {
					percent = float64(item.Size) / float64(m.currentNode.Size)
				}
				filledLen := int(percent * float64(barWidth))
				emptyLen := barWidth - filledLen
				bar = barFull.Render(strings.Repeat("■", filledLen)) + barEmpty.Render(strings.Repeat("-", emptyLen))

				name = item.Name
				if item.IsDir {
					name += "/"
				}
			}

			row := fmt.Sprintf("%s  %s  %s", sizeStr, bar, name)

			if i == m.cursor {
				row = selectedStyle.Render(fmt.Sprintf("%-*s", m.width-4, row))
			} else {
				row = fmt.Sprintf("  %s", row)
			}
			rows = append(rows, row)
		}

		content := strings.Join(rows, "\n")

		var footer string
		if m.showHelp {
			footer = helpStyle.Render("\n ?: aide • ↑/↓/←/→: naviguer • enter: sélectionner • q: quitter")
		} else {
			footer = helpStyle.Render("\n ?: réduire aide • ↑/↓/←/→: naviguer • enter: sélectionner • q: quitter\n g: explorer • b: shell • r: recalculer")
		}

		return fmt.Sprintf("\n%s\n%s\n%s", header, content, footer)
	}

	return ""
}

// --- COMMANDES DE SCAN ---

func scanDirectoryCmd(path string, counter *int64, visited map[fileID]struct{}) tea.Cmd {
	return func() tea.Msg {
		diskSize := getPartitionSize(path)
		root, err := scanRecursively(path, nil, counter, visited)
		return scanFinishedMsg{root: root, diskSize: diskSize, err: err}
	}
}

func refreshDirectoryCmd(path string, counter *int64, visited map[fileID]struct{}) tea.Cmd {
	return func() tea.Msg {
		root, err := scanRecursively(path, nil, counter, visited)
		return refreshFinishedMsg{newNode: root, err: err}
	}
}

// --- FONCTION DE SCAN RECURSIVE ---

func scanRecursively(path string, parent *FileNode, counter *int64, visited map[fileID]struct{}) (*FileNode, error) {
	atomic.AddInt64(counter, 1)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(absPath)
	if parent == nil {
		name = absPath
	}

	node := &FileNode{
		Name:   name,
		Path:   absPath,
		IsDir:  true,
		Parent: parent,
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return node, nil
	}

	var totalSize int64

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if node.Path == "/" && (entry.Name() == "proc" || entry.Name() == "sys" || entry.Name() == "dev" || entry.Name() == "run") {
			continue
		}

		childPath := filepath.Join(absPath, entry.Name())

		if entry.IsDir() {
			childNode, _ := scanRecursively(childPath, node, counter, visited)
			if childNode != nil {
				node.Children = append(node.Children, childNode)
				totalSize += childNode.Size
			}
		} else {
			atomic.AddInt64(counter, 1)

			var size int64
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				size = stat.Blocks * 512
				id := fileID{dev: stat.Dev, ino: stat.Ino}
				if _, seen := visited[id]; !seen {
					visited[id] = struct{}{}
					totalSize += size
				}
			} else {
				size = info.Size()
				totalSize += size
			}

			child := &FileNode{
				Name:   entry.Name(),
				Path:   childPath,
				Size:   size,
				IsDir:  false,
				Parent: node,
			}
			node.Children = append(node.Children, child)
		}
	}

	node.Size = totalSize

	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})

	return node, nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func getPartitionSize(path string) int64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0
	}
	return int64(stat.Blocks) * int64(stat.Bsize)
}

// --- NOUVELLE FONCTION POUR GERER LE "~" ---
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path // En cas d'erreur, on laisse le path tel quel
		}
		if path == "~" {
			return home
		}
		// Si c'est "~/Documents" par exemple
		if len(path) > 1 && path[1] == '/' {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func main() {
	p := tea.NewProgram(New(80, 24), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur lors de l'exécution : %v\n", err)
		os.Exit(1)
	}
}
