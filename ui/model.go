// package ui

// import (
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"strings"
// 	"sync/atomic"

// 	"github.com/charmbracelet/bubbles/spinner"
// 	"github.com/charmbracelet/bubbles/textinput"
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/charmbracelet/lipgloss"

// 	"aed/scanner"
// )

// // --- MESSAGES ---

// type BackMsg struct{}

// type refreshFinishedMsg struct {
// 	newNode *scanner.FileNode
// 	err     error
// }

// type scanFinishedMsg struct {
// 	root     *scanner.FileNode
// 	diskSize int64
// 	err      error
// }

// // --- ETATS ---

// type SessionState int

// const (
// 	StateInputPath SessionState = iota
// 	StateScanning
// 	StateBrowsing
// )

// // --- MODEL ---

// type Model struct {
// 	state        SessionState
// 	pathInput    textinput.Model
// 	excludeInput textinput.Model
// 	focusIndex   int // 0: path, 1: exclude

// 	spinner spinner.Model

// 	root        *scanner.FileNode
// 	currentNode *scanner.FileNode
// 	cursor      int
// 	yOffset     int

// 	filesScanned  *int64
// 	showHelp      bool
// 	diskTotalSize int64

// 	currentExclusions []string

// 	width, height int
// 	err           error
// }

// func New(w, h int) Model {
// 	// Input Path
// 	ti := textinput.New()
// 	ti.Placeholder = "/home/user (ou ~)"
// 	ti.Focus()
// 	ti.CharLimit = 256
// 	ti.Width = 50
// 	ti.SetValue(".")

// 	// Input Exclusions
// 	ei := textinput.New()
// 	ei.Placeholder = "node_modules, .git, *.tmp"
// 	ei.CharLimit = 256
// 	ei.Width = 50

// 	// Spinner (Rose #FF2A6D)
// 	s := spinner.New()
// 	s.Spinner = spinner.Dot
// 	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))

// 	zero := int64(0)

// 	return Model{
// 		state:        StateInputPath,
// 		pathInput:    ti,
// 		excludeInput: ei,
// 		focusIndex:   0,
// 		spinner:      s,
// 		filesScanned: &zero,
// 		width:        w,
// 		height:       h,
// 		showHelp:     true,
// 	}
// }

// func (m Model) Init() tea.Cmd {
// 	return textinput.Blink
// }

// func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmd tea.Cmd

// 	switch msg := msg.(type) {

// 	case BackMsg:
// 		return m, tea.Quit

// 	case tea.WindowSizeMsg:
// 		m.width = msg.Width
// 		m.height = msg.Height

// 	case tea.KeyMsg:
// 		if msg.String() == "ctrl+c" {
// 			return m, func() tea.Msg { return BackMsg{} }
// 		}

// 		// --- PAGE D'ACCUEIL (INPUTS) ---
// 		if m.state == StateInputPath {
// 			switch msg.String() {

// 			// Navigation entre les champs
// 			case "tab", "shift+tab", "up", "down":
// 				if m.focusIndex == 0 {
// 					m.focusIndex = 1
// 					m.pathInput.Blur()
// 					m.excludeInput.Focus()
// 				} else {
// 					m.focusIndex = 0
// 					m.excludeInput.Blur()
// 					m.pathInput.Focus()
// 				}
// 				return m, textinput.Blink

// 			// Validation
// 			case "enter":
// 				// 1. Gestion du chemin principal
// 				rawInput := m.pathInput.Value()
// 				path := rawInput
// 				if path == "." {
// 					path, _ = os.Getwd()
// 				}
// 				path = scanner.ExpandPath(path)

// 				// 2. Gestion des exclusions (CORRIGÉE)
// 				rawExcludes := m.excludeInput.Value()
// 				var exclusions []string
// 				if strings.TrimSpace(rawExcludes) != "" {
// 					parts := strings.Split(rawExcludes, ",")
// 					for _, p := range parts {
// 						trimmed := strings.TrimSpace(p)
// 						// On applique l'expansion du tilde (~) sur chaque exclusion
// 						expanded := scanner.ExpandPath(trimmed)
// 						exclusions = append(exclusions, expanded)
// 					}
// 				}
// 				m.currentExclusions = exclusions

// 				m.state = StateScanning
// 				atomic.StoreInt64(m.filesScanned, 0)
// 				visitedInodes := make(map[scanner.FileID]struct{})

// 				return m, tea.Batch(
// 					m.spinner.Tick,
// 					scanDirectoryCmd(path, m.filesScanned, visitedInodes, exclusions),
// 				)

// 			case "esc", "q":
// 				return m, func() tea.Msg { return BackMsg{} }
// 			}

// 			if m.focusIndex == 0 {
// 				m.pathInput, cmd = m.pathInput.Update(msg)
// 			} else {
// 				m.excludeInput, cmd = m.excludeInput.Update(msg)
// 			}
// 			return m, cmd
// 		}

// 		// --- SCAN EN COURS ---
// 		if m.state == StateScanning {
// 			if msg.String() == "q" || msg.String() == "esc" {
// 				return m, func() tea.Msg { return BackMsg{} }
// 			}
// 		}

// 		// --- NAVIGATION FICHIERS ---
// 		if m.state == StateBrowsing {
// 			items := m.getDisplayItems()

// 			switch msg.String() {
// 			case "q":
// 				return m, func() tea.Msg { return BackMsg{} }
// 			case ",", "?":
// 				m.showHelp = !m.showHelp
// 				return m, nil

// 			// Recalculer (Refresh)
// 			case "r":
// 				if m.currentNode != nil {
// 					m.state = StateScanning
// 					atomic.StoreInt64(m.filesScanned, 0)
// 					visitedInodes := make(map[scanner.FileID]struct{})

// 					return m, tea.Batch(
// 						m.spinner.Tick,
// 						refreshDirectoryCmd(m.currentNode.Path, m.filesScanned, visitedInodes, m.currentExclusions),
// 					)
// 				}
// 				return m, nil

// 			case "backspace", "left", "h", "esc":
// 				if m.currentNode.Parent != nil {
// 					m.currentNode = m.currentNode.Parent
// 					m.cursor = 0
// 					m.yOffset = 0
// 				}

// 			case "enter", "right", "l":
// 				if len(items) > 0 && m.cursor < len(items) {
// 					selected := items[m.cursor]
// 					if selected.Name == ".." {
// 						if m.currentNode.Parent != nil {
// 							m.currentNode = m.currentNode.Parent
// 							m.cursor = 0
// 							m.yOffset = 0
// 						}
// 						return m, nil
// 					}
// 					if selected.Name == "." {
// 						return m, nil
// 					}
// 					if selected.IsDir {
// 						for _, child := range m.currentNode.Children {
// 							if child.Path == selected.Path {
// 								m.currentNode = child
// 								m.cursor = 0
// 								m.yOffset = 0
// 								break
// 							}
// 						}
// 					}
// 				}

// 			case "g":
// 				if len(items) > 0 && m.cursor < len(items) {
// 					selected := items[m.cursor]
// 					cmd := exec.Command("xdg-open", selected.Path)
// 					cmd.Start()
// 				}
// 				return m, nil

// 			case "b":
// 				if len(items) > 0 && m.cursor < len(items) {
// 					selected := items[m.cursor]
// 					targetPath := selected.Path
// 					if !selected.IsDir {
// 						targetPath = filepath.Dir(selected.Path)
// 					}
// 					shell := os.Getenv("SHELL")
// 					if shell == "" {
// 						shell = "/bin/bash"
// 					}
// 					c := exec.Command(shell)
// 					c.Dir = targetPath
// 					return m, tea.ExecProcess(c, func(err error) tea.Msg { return nil })
// 				}
// 				return m, nil

// 			case "up", "k":
// 				if m.cursor > 0 {
// 					m.cursor--
// 					if m.cursor < m.yOffset {
// 						m.yOffset = m.cursor
// 					}
// 				}
// 			case "down", "j":
// 				if m.cursor < len(items)-1 {
// 					m.cursor++
// 					footerHeight := 2
// 					if !m.showHelp {
// 						footerHeight = 0
// 					}
// 					visibleHeight := m.height - 5 - footerHeight
// 					if m.cursor >= m.yOffset+visibleHeight {
// 						m.yOffset = m.cursor - visibleHeight + 1
// 					}
// 				}
// 			}
// 		}

// 	case scanFinishedMsg:
// 		if msg.err != nil {
// 			m.err = msg.err
// 			m.state = StateInputPath
// 		} else {
// 			m.root = msg.root
// 			m.currentNode = msg.root
// 			m.diskTotalSize = msg.diskSize
// 			m.state = StateBrowsing
// 		}

// 	case refreshFinishedMsg:
// 		if msg.err != nil {
// 			m.err = msg.err
// 			m.state = StateBrowsing
// 		} else {
// 			newNode := msg.newNode
// 			oldNode := m.currentNode

// 			// La greffe chirurgicale
// 			if oldNode.Parent != nil {
// 				newNode.Parent = oldNode.Parent
// 				newNode.Name = filepath.Base(newNode.Path)

// 				for i, child := range oldNode.Parent.Children {
// 					if child.Path == oldNode.Path {
// 						oldNode.Parent.Children[i] = newNode
// 						break
// 					}
// 				}

// 				diff := newNode.Size - oldNode.Size
// 				parent := newNode.Parent
// 				for parent != nil {
// 					parent.Size += diff
// 					parent = parent.Parent
// 				}
// 			} else {
// 				m.root = newNode
// 			}

// 			m.currentNode = newNode
// 			if m.cursor >= len(m.getDisplayItems()) {
// 				m.cursor = 0
// 				m.yOffset = 0
// 			}
// 			m.state = StateBrowsing
// 		}

// 	case spinner.TickMsg:
// 		if m.state == StateScanning {
// 			var cmdSpinner tea.Cmd
// 			m.spinner, cmdSpinner = m.spinner.Update(msg)
// 			return m, cmdSpinner
// 		}
// 	}

// 	return m, nil
// }

// // --- COMMANDES ---

// func scanDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
// 	return func() tea.Msg {
// 		diskSize := scanner.GetPartitionSize(path)
// 		root, err := scanner.ScanRecursively(path, nil, counter, visited, exclusions)
// 		return scanFinishedMsg{root: root, diskSize: diskSize, err: err}
// 	}
// }

// func refreshDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
// 	return func() tea.Msg {
// 		root, err := scanner.ScanRecursively(path, nil, counter, visited, exclusions)
// 		return refreshFinishedMsg{newNode: root, err: err}
// 	}
// }

package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"aed/scanner"
)

// --- MESSAGES ---

type BackMsg struct{}

type refreshFinishedMsg struct {
	newNode *scanner.FileNode
	err     error
}

type scanFinishedMsg struct {
	root     *scanner.FileNode
	diskSize int64
	err      error
}

// --- ETATS ---

type SessionState int

const (
	StateInputPath SessionState = iota
	StateScanning
	StateBrowsing
)

// --- MODEL ---

type Model struct {
	state        SessionState
	pathInput    textinput.Model
	excludeInput textinput.Model
	focusIndex   int 

	spinner spinner.Model

	root        *scanner.FileNode
	currentNode *scanner.FileNode
	cursor      int
	yOffset     int

	filesScanned  *int64
	showHelp      bool
	diskTotalSize int64
	
	currentExclusions []string

	// NOUVEAU : Langue active
	lang Language

	width, height int
	err           error
}

func New(w, h int) Model {
	// Par défaut on part sur le Français
	currentLang := fr

	// Input Path
	ti := textinput.New()
	ti.Placeholder = currentLang.PathPlaceholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.SetValue(".")

	// Input Exclusions
	ei := textinput.New()
	ei.Placeholder = currentLang.ExcludePlaceholder
	ei.CharLimit = 256
	ei.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))

	zero := int64(0)

	return Model{
		state:        StateInputPath,
		pathInput:    ti,
		excludeInput: ei,
		focusIndex:   0,
		spinner:      s,
		filesScanned: &zero,
		width:        w,
		height:       h,
		showHelp:     true,
		lang:         currentLang, // Initialisation
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

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

		// --- GESTION CHANGEMENT DE LANGUE (GLOBAL) ---
		if msg.String() == "l" || msg.String() == "L" {
			if m.lang.Code == "FR" {
				m.lang = en
			} else {
				m.lang = fr
			}
			// Mise à jour des placeholders si on change de langue
			m.pathInput.Placeholder = m.lang.PathPlaceholder
			m.excludeInput.Placeholder = m.lang.ExcludePlaceholder
			return m, nil
		}
		// ---------------------------------------------

		// --- PAGE D'ACCUEIL ---
		if m.state == StateInputPath {
			switch msg.String() {
			
			case "tab", "shift+tab", "up", "down":
				if m.focusIndex == 0 {
					m.focusIndex = 1
					m.pathInput.Blur()
					m.excludeInput.Focus()
				} else {
					m.focusIndex = 0
					m.excludeInput.Blur()
					m.pathInput.Focus()
				}
				return m, textinput.Blink

			case "enter":
				rawInput := m.pathInput.Value()
				path := rawInput
				if path == "." {
					path, _ = os.Getwd()
				}
				path = scanner.ExpandPath(path)

				rawExcludes := m.excludeInput.Value()
				var exclusions []string
				if strings.TrimSpace(rawExcludes) != "" {
					parts := strings.Split(rawExcludes, ",")
					for _, p := range parts {
						trimmed := strings.TrimSpace(p)
						expanded := scanner.ExpandPath(trimmed)
						exclusions = append(exclusions, expanded)
					}
				}
				m.currentExclusions = exclusions

				m.state = StateScanning
				atomic.StoreInt64(m.filesScanned, 0)
				visitedInodes := make(map[scanner.FileID]struct{})

				return m, tea.Batch(
					m.spinner.Tick,
					scanDirectoryCmd(path, m.filesScanned, visitedInodes, exclusions),
				)

			case "esc", "q":
				return m, func() tea.Msg { return BackMsg{} }
			}

			if m.focusIndex == 0 {
				m.pathInput, cmd = m.pathInput.Update(msg)
			} else {
				m.excludeInput, cmd = m.excludeInput.Update(msg)
			}
			return m, cmd
		}

		// --- SCAN EN COURS ---
		if m.state == StateScanning {
			if msg.String() == "q" || msg.String() == "esc" {
				return m, func() tea.Msg { return BackMsg{} }
			}
		}

		// --- BROWSING ---
		if m.state == StateBrowsing {
			items := m.getDisplayItems()

			switch msg.String() {
			case "q":
				return m, func() tea.Msg { return BackMsg{} }
			case ",", "?":
				m.showHelp = !m.showHelp
				return m, nil

			case "r":
				if m.currentNode != nil {
					m.state = StateScanning
					atomic.StoreInt64(m.filesScanned, 0)
					visitedInodes := make(map[scanner.FileID]struct{})

					return m, tea.Batch(
						m.spinner.Tick,
						refreshDirectoryCmd(m.currentNode.Path, m.filesScanned, visitedInodes, m.currentExclusions),
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

// --- COMMANDES ---

func scanDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
	return func() tea.Msg {
		diskSize := scanner.GetPartitionSize(path)
		root, err := scanner.ScanRecursively(path, nil, counter, visited, exclusions)
		return scanFinishedMsg{root: root, diskSize: diskSize, err: err}
	}
}

func refreshDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
	return func() tea.Msg {
		root, err := scanner.ScanRecursively(path, nil, counter, visited, exclusions)
		return refreshFinishedMsg{newNode: root, err: err}
	}
}