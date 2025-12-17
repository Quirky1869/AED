package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"aed/scanner"
)

// Messages internes pour le traitement asynchrone
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

// États de la machine à états principale
type SessionState int

const (
	StateInputPath SessionState = iota
	StateScanning
	StateBrowsing
)

// Modes de tri des fichiers
type SortMode int

const (
	SortBySize SortMode = iota
	SortByName
	SortByCount
)

// Modèle principal contenant l'état de l'application et les données
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
	lang              Language

	// Options de tri et d'affichage
	sortMode   SortMode
	sortDesc   bool
	showHidden bool

	// État pour l'autocomplétion cyclique
	suggestions     []string
	suggestionIndex int
	suggestionBase  string

	width, height int
	err           error
}

// Initialisation du modèle avec les valeurs par défaut
func New(w, h int) Model {
	currentLang := fr

	ti := textinput.New()
	ti.Placeholder = currentLang.PathPlaceholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.SetValue(".")

	ei := textinput.New()
	ei.Placeholder = currentLang.ExcludePlaceholder
	ei.CharLimit = 256
	ei.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

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
		lang:         currentLang,
		sortMode:     SortBySize,
		sortDesc:     true,
		showHidden:   true,
		suggestions:  []string{},
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Applique le tri sur le dossier courant selon le mode (Taille, Nom, Nombre)
func (m *Model) applySort() {
	if m.currentNode == nil || len(m.currentNode.Children) == 0 {
		return
	}

	sort.Slice(m.currentNode.Children, func(i, j int) bool {
		a := m.currentNode.Children[i]
		b := m.currentNode.Children[j]

		isLess := false
		switch m.sortMode {
		case SortByName:
			isLess = strings.ToLower(a.Name) < strings.ToLower(b.Name)
		case SortByCount:
			isLess = a.FileCount < b.FileCount
		default:
			isLess = a.Size < b.Size
		}

		if m.sortDesc {
			return !isLess
		}
		return isLess
	})
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

		// Bascule de la langue
		if msg.String() == "ctrl+l" {
			if m.lang.Code == "FR" {
				m.lang = en
			} else {
				m.lang = fr
			}
			m.pathInput.Placeholder = m.lang.PathPlaceholder
			m.excludeInput.Placeholder = m.lang.ExcludePlaceholder
			return m, nil
		}

		// Gestion de la vue de Saisie (Input)
		if m.state == StateInputPath {
			switch msg.String() {

			// Gestion de l'autocomplétion cyclique (Tab / Shift+Tab)
			case "tab":
				var activeInput *textinput.Model
				isExcludeField := false

				if m.focusIndex == 0 {
					activeInput = &m.pathInput
				} else {
					activeInput = &m.excludeInput
					isExcludeField = true
				}

				// Initialisation du cycle de suggestions
				if len(m.suggestions) == 0 {
					fullText := activeInput.Value()
					searchStr := fullText
					prefix := ""

					if isExcludeField {
						lastComma := strings.LastIndex(fullText, ",")
						if lastComma != -1 {
							prefix = fullText[:lastComma+1] + " "
							searchStr = strings.TrimSpace(fullText[lastComma+1:])
						}
					}

					baseDir, candidates := GetAutocompleteSuggestions(searchStr, true)

					if len(candidates) > 0 {
						m.suggestions = candidates
						m.suggestionBase = prefix + baseDir
						m.suggestionIndex = 0

						candidate := candidates[0]
						newValue := m.suggestionBase + candidate
						if isDirOrShouldSlash(candidate) && !strings.HasSuffix(newValue, string(os.PathSeparator)) {
							newValue += string(os.PathSeparator)
						}

						activeInput.SetValue(newValue)
						activeInput.SetCursor(len(newValue))
					}
				} else {
					// Passage au candidat suivant dans le cycle
					m.suggestionIndex = (m.suggestionIndex + 1) % len(m.suggestions)

					candidate := m.suggestions[m.suggestionIndex]
					newValue := m.suggestionBase + candidate

					if isDirOrShouldSlash(candidate) && !strings.HasSuffix(newValue, string(os.PathSeparator)) {
						newValue += string(os.PathSeparator)
					}

					activeInput.SetValue(newValue)
					activeInput.SetCursor(len(newValue))
				}
				return m, nil

			case "shift+tab":
				// Navigation arrière dans le cycle de suggestions
				if len(m.suggestions) > 0 {
					var activeInput *textinput.Model
					if m.focusIndex == 0 {
						activeInput = &m.pathInput
					} else {
						activeInput = &m.excludeInput
					}

					m.suggestionIndex--
					if m.suggestionIndex < 0 {
						m.suggestionIndex = len(m.suggestions) - 1
					}

					candidate := m.suggestions[m.suggestionIndex]
					newValue := m.suggestionBase + candidate

					if isDirOrShouldSlash(candidate) && !strings.HasSuffix(newValue, string(os.PathSeparator)) {
						newValue += string(os.PathSeparator)
					}

					activeInput.SetValue(newValue)
					activeInput.SetCursor(len(newValue))
					return m, nil
				}

				// Sinon, changement de focus standard
				m.suggestions = nil
				m.focusIndex = 0
				m.excludeInput.Blur()
				m.pathInput.Focus()
				return m, textinput.Blink

			case "up":
				m.suggestions = nil
				m.focusIndex = 0
				m.excludeInput.Blur()
				m.pathInput.Focus()
				return m, textinput.Blink

			case "down":
				m.suggestions = nil
				m.focusIndex = 1
				m.pathInput.Blur()
				m.excludeInput.Focus()
				return m, textinput.Blink

			case "enter":
				m.suggestions = nil
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
						expanded = strings.TrimRight(expanded, string(os.PathSeparator))
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

			case "esc":
				return m, func() tea.Msg { return BackMsg{} }

			default:
				m.suggestions = nil
			}

			if m.focusIndex == 0 {
				m.pathInput, cmd = m.pathInput.Update(msg)
			} else {
				m.excludeInput, cmd = m.excludeInput.Update(msg)
			}
			return m, cmd
        }

        // Gestion de la vue Scan
        if m.state == StateScanning {
            if msg.String() == "q" || msg.String() == "esc" {
                return m, func() tea.Msg { return BackMsg{} }
            }
        }

        // Gestion de la navigation (Browsing)
        if m.state == StateBrowsing {
            items := m.getDisplayItems()

            switch msg.String() {
            case "q":
                return m, func() tea.Msg { return BackMsg{} }
            case ",", "?":
                m.showHelp = !m.showHelp
                return m, nil

            case "e":
                m.showHidden = !m.showHidden
                newItems := m.getDisplayItems()
                if m.cursor >= len(newItems) {
                    m.cursor = len(newItems) - 1
                    if m.cursor < 0 {
                        m.cursor = 0
                    }
                }
                visibleHeight := m.height - 7
                if visibleHeight > 0 {
                    if m.cursor < m.yOffset {
                        m.yOffset = m.cursor
                    }
                }
                return m, nil

            // Options de Tri
            case "s":
                if m.sortMode == SortBySize {
                    m.sortDesc = !m.sortDesc
                } else {
                    m.sortMode = SortBySize
                    m.sortDesc = true
                }
                m.applySort()
                return m, nil

            case "n":
                if m.sortMode == SortByName {
                    m.sortDesc = !m.sortDesc
                } else {
                    m.sortMode = SortByName
                    m.sortDesc = false
                }
                m.applySort()
                return m, nil

            case "C":
                if m.sortMode == SortByCount {
                    m.sortDesc = !m.sortDesc
                } else {
                    m.sortMode = SortByCount
                    m.sortDesc = true
                }
                m.applySort()
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

            case "left", "h":
                if m.currentNode.Parent != nil {
                    m.currentNode = m.currentNode.Parent
                    m.cursor = 0
                    m.yOffset = 0
                    m.applySort()
                }
                return m, nil

            case "backspace", "esc":
                m.state = StateInputPath
                m.focusIndex = 0
                m.pathInput.Focus()
                m.excludeInput.Blur()
                return m, nil

            case "enter", "right", "l":
                if len(items) > 0 && m.cursor < len(items) {
                    selected := items[m.cursor]
                    if selected.Name == ".." {
                        if m.currentNode.Parent != nil {
                            m.currentNode = m.currentNode.Parent
                            m.cursor = 0
                            m.yOffset = 0
                            m.applySort()
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
                                m.applySort()
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
            m.applySort()
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
            m.applySort()
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

// Helper pour déterminer si un slash doit être ajouté au chemin (pour les dossiers)
func isDirOrShouldSlash(name string) bool {
    return strings.HasSuffix(name, string(os.PathSeparator))
}

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