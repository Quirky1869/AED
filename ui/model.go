// package ui

// import (
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"sort"
// 	"strings"
// 	"sync/atomic"

// 	"github.com/charmbracelet/bubbles/spinner"
// 	"github.com/charmbracelet/bubbles/textinput"
// 	tea "github.com/charmbracelet/bubbletea"

// 	"github.com/quirky1869/aed/scanner"
// )

// // Messages internes pour la communication asynchrone (fin de scan, refresh)
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

// // États de la machine à états de l'interface
// type SessionState int

// const (
// 	StateInputPath SessionState = iota
// 	StateScanning
// 	StateBrowsing
// )

// // Modes de tri disponibles pour l'affichage des fichiers
// type SortMode int

// const (
// 	SortBySize SortMode = iota
// 	SortByName
// 	SortByCount
// )

// // Modèle principal contenant l'état de l'application, les données et la configuration UI
// type Model struct {
// 	state        SessionState
// 	pathInput    textinput.Model
// 	excludeInput textinput.Model
// 	focusIndex   int

// 	spinner spinner.Model

// 	root        *scanner.FileNode
// 	currentNode *scanner.FileNode
// 	cursor      int
// 	yOffset     int

// 	filesScanned  *int64
// 	showHelp      bool
// 	diskTotalSize int64

// 	currentExclusions []string
// 	lang              Language

// 	sortMode   SortMode
// 	sortDesc   bool
// 	showHidden bool

// 	suggestions     []string
// 	suggestionIndex int
// 	suggestionBase  string

// 	width, height int
// 	err           error
// }

// // Initialisation du modèle avec les valeurs par défaut
// func New(w, h int) Model {
// 	currentLang := fr

// 	ti := textinput.New()
// 	ti.Placeholder = currentLang.PathPlaceholder
// 	ti.Focus()
// 	ti.CharLimit = 256
// 	ti.Width = 50
// 	ti.SetValue(".")

// 	ei := textinput.New()
// 	ei.Placeholder = currentLang.ExcludePlaceholder
// 	ei.CharLimit = 256
// 	ei.Width = 50

// 	s := spinner.New()
// 	s.Spinner = spinner.Dot
// 	s.Style = spinnerStyle

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
// 		lang:         currentLang,
// 		sortMode:     SortBySize,
// 		sortDesc:     true,
// 		showHidden:   true,
// 		suggestions:  []string{},
// 	}
// }

// func (m Model) Init() tea.Cmd {
// 	return textinput.Blink
// }

// // Trie les enfants du nœud courant selon le mode sélectionné (Taille, Nom, Nombre)
// func (m *Model) applySort() {
// 	if m.currentNode == nil || len(m.currentNode.Children) == 0 {
// 		return
// 	}

// 	sort.Slice(m.currentNode.Children, func(i, j int) bool {
// 		a := m.currentNode.Children[i]
// 		b := m.currentNode.Children[j]

// 		isLess := false
// 		switch m.sortMode {
// 		case SortByName:
// 			isLess = strings.ToLower(a.Name) < strings.ToLower(b.Name)
// 		case SortByCount:
// 			isLess = a.FileCount < b.FileCount
// 		default:
// 			isLess = a.Size < b.Size
// 		}

// 		if m.sortDesc {
// 			return !isLess
// 		}
// 		return isLess
// 	})
// }

// // Boucle principale de gestion des événements
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

// 		// Bascule de la langue (FR <-> EN)
// 		if msg.String() == "ctrl+l" {
// 			if m.lang.Code == "FR" {
// 				m.lang = en
// 			} else {
// 				m.lang = fr
// 			}
// 			m.pathInput.Placeholder = m.lang.PathPlaceholder
// 			m.excludeInput.Placeholder = m.lang.ExcludePlaceholder
// 			return m, nil
// 		}

// 		// Gestion de l'écran de saisie (Input)
// 		if m.state == StateInputPath {
// 			switch msg.String() {

// 			// Gestion de l'autocomplétion (Tab)
// 			case "tab":
// 				var activeInput *textinput.Model
// 				isExcludeField := false

// 				if m.focusIndex == 0 {
// 					activeInput = &m.pathInput
// 				} else {
// 					activeInput = &m.excludeInput
// 					isExcludeField = true
// 				}

// 				// Initialisation du cycle de suggestions
// 				if len(m.suggestions) == 0 {
// 					fullText := activeInput.Value()
// 					searchStr := fullText
// 					prefix := ""

// 					if isExcludeField {
// 						lastComma := strings.LastIndex(fullText, ",")
// 						if lastComma != -1 {
// 							prefix = fullText[:lastComma+1] + " "
// 							searchStr = strings.TrimSpace(fullText[lastComma+1:])
// 						}
// 					}

// 					baseDir, candidates := GetAutocompleteSuggestions(searchStr, true)

// 					if len(candidates) > 0 {
// 						m.suggestions = candidates
// 						m.suggestionBase = prefix + baseDir
// 						m.suggestionIndex = 0

// 						candidate := candidates[0]
// 						newValue := m.suggestionBase + candidate
// 						if isDirOrShouldSlash(candidate) && !strings.HasSuffix(newValue, string(os.PathSeparator)) {
// 							newValue += string(os.PathSeparator)
// 						}

// 						activeInput.SetValue(newValue)
// 						activeInput.SetCursor(len(newValue))
// 					}
// 				} else {
// 					// Cycle suivant
// 					m.suggestionIndex = (m.suggestionIndex + 1) % len(m.suggestions)
// 					candidate := m.suggestions[m.suggestionIndex]
// 					newValue := m.suggestionBase + candidate
// 					if isDirOrShouldSlash(candidate) && !strings.HasSuffix(newValue, string(os.PathSeparator)) {
// 						newValue += string(os.PathSeparator)
// 					}
// 					activeInput.SetValue(newValue)
// 					activeInput.SetCursor(len(newValue))
// 				}
// 				return m, nil

// 			// Gestion de l'autocomplétion inverse (Shift+Tab)
// 			case "shift+tab":
// 				if len(m.suggestions) > 0 {
// 					var activeInput *textinput.Model
// 					if m.focusIndex == 0 {
// 						activeInput = &m.pathInput
// 					} else {
// 						activeInput = &m.excludeInput
// 					}
// 					m.suggestionIndex--
// 					if m.suggestionIndex < 0 {
// 						m.suggestionIndex = len(m.suggestions) - 1
// 					}
// 					candidate := m.suggestions[m.suggestionIndex]
// 					newValue := m.suggestionBase + candidate
// 					if isDirOrShouldSlash(candidate) && !strings.HasSuffix(newValue, string(os.PathSeparator)) {
// 						newValue += string(os.PathSeparator)
// 					}
// 					activeInput.SetValue(newValue)
// 					activeInput.SetCursor(len(newValue))
// 					return m, nil
// 				}
// 				// Sinon navigation champ précédent
// 				m.suggestions = nil
// 				m.focusIndex = 0
// 				m.excludeInput.Blur()
// 				m.pathInput.Focus()
// 				return m, textinput.Blink

// 			// Navigation entre les champs
// 			case "up":
// 				m.suggestions = nil
// 				m.focusIndex = 0
// 				m.excludeInput.Blur()
// 				m.pathInput.Focus()
// 				return m, textinput.Blink

// 			case "down":
// 				m.suggestions = nil
// 				m.focusIndex = 1
// 				m.pathInput.Blur()
// 				m.excludeInput.Focus()
// 				return m, textinput.Blink

// 			// Validation et lancement du scan
// 			case "enter":
// 				m.suggestions = nil
// 				rawInput := m.pathInput.Value()
// 				path := rawInput
// 				if path == "." {
// 					path, _ = os.Getwd()
// 				}
// 				path = scanner.ExpandPath(path)

// 				rawExcludes := m.excludeInput.Value()
// 				var exclusions []string
// 				if strings.TrimSpace(rawExcludes) != "" {
// 					parts := strings.Split(rawExcludes, ",")
// 					for _, p := range parts {
// 						trimmed := strings.TrimSpace(p)
// 						expanded := scanner.ExpandPath(trimmed)
// 						expanded = strings.TrimRight(expanded, string(os.PathSeparator))
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

// 			case "esc":
// 				return m, func() tea.Msg { return BackMsg{} }

// 			default:
// 				m.suggestions = nil
// 			}

// 			if m.focusIndex == 0 {
// 				m.pathInput, cmd = m.pathInput.Update(msg)
// 			} else {
// 				m.excludeInput, cmd = m.excludeInput.Update(msg)
// 			}
// 			return m, cmd
// 		}

// 		// Gestion de l'état Scanning
// 		if m.state == StateScanning {
// 			if msg.String() == "q" || msg.String() == "esc" {
// 				return m, func() tea.Msg { return BackMsg{} }
// 			}
// 		}

// 		// Gestion de l'état Browsing (Navigation dans les résultats)
// 		if m.state == StateBrowsing {
// 			items := m.getDisplayItems()

// 			footerHeight := 2
// 			if !m.showHelp {
// 				footerHeight = 0
// 			}
// 			visibleHeight := m.height - 6 - footerHeight
// 			if visibleHeight < 1 {
// 				visibleHeight = 1
// 			}

// 			switch msg.String() {
// 			case "q":
// 				return m, func() tea.Msg { return BackMsg{} }
// 			case ",", "?":
// 				m.showHelp = !m.showHelp
// 				return m, nil

// 			// Affichage / Masquage des fichiers cachés
// 			case "e":
// 				m.showHidden = !m.showHidden
// 				newItems := m.getDisplayItems()
// 				if m.cursor >= len(newItems) {
// 					m.cursor = len(newItems) - 1
// 					if m.cursor < 0 {
// 						m.cursor = 0
// 					}
// 				}
// 				if m.cursor < m.yOffset {
// 					m.yOffset = m.cursor
// 				}
// 				return m, nil

// 			// Options de tri (Taille, Nom, Nombre)
// 			case "s":
// 				if m.sortMode == SortBySize {
// 					m.sortDesc = !m.sortDesc
// 				} else {
// 					m.sortMode = SortBySize
// 					m.sortDesc = true
// 				}
// 				m.applySort()
// 				return m, nil

// 			case "n":
// 				if m.sortMode == SortByName {
// 					m.sortDesc = !m.sortDesc
// 				} else {
// 					m.sortMode = SortByName
// 					m.sortDesc = false
// 				}
// 				m.applySort()
// 				return m, nil

// 			case "C":
// 				if m.sortMode == SortByCount {
// 					m.sortDesc = !m.sortDesc
// 				} else {
// 					m.sortMode = SortByCount
// 					m.sortDesc = true
// 				}
// 				m.applySort()
// 				return m, nil

// 			// Rafraîchir le dossier courant
// 			case "r":
// 				if m.currentNode != nil {
// 					m.state = StateScanning
// 					atomic.StoreInt64(m.filesScanned, 0)
// 					visitedInodes := make(map[scanner.FileID]struct{})
// 					return m, tea.Batch(
// 						m.spinner.Tick,
// 						refreshDirectoryCmd(m.currentNode.FullPath(), m.filesScanned, visitedInodes, m.currentExclusions),
// 					)
// 				}
// 				return m, nil

// 			// Navigation hiérarchique (Parent/Enfant)
// 			case "left", "h":
// 				if m.currentNode.Parent != nil {
// 					m.currentNode = m.currentNode.Parent
// 					m.cursor = 0
// 					m.yOffset = 0
// 					m.applySort()
// 				}
// 				return m, nil

// 			case "backspace", "esc":
// 				m.state = StateInputPath
// 				m.focusIndex = 0
// 				m.pathInput.Focus()
// 				m.excludeInput.Blur()
// 				return m, nil

// 			case "enter", "right", "l":
// 				if len(items) > 0 && m.cursor < len(items) {
// 					selected := items[m.cursor]
// 					if selected.Name == ".." {
// 						if m.currentNode.Parent != nil {
// 							m.currentNode = m.currentNode.Parent
// 							m.cursor = 0
// 							m.yOffset = 0
// 							m.applySort()
// 						}
// 						return m, nil
// 					}
// 					if selected.Name == "." {
// 						return m, nil
// 					}
// 					if selected.IsDir {
// 						for _, child := range m.currentNode.Children {
// 							if child.Name == selected.Name {
// 								m.currentNode = child
// 								m.cursor = 0
// 								m.yOffset = 0
// 								m.applySort()
// 								break
// 							}
// 						}
// 					}
// 				}

// 			// Actions système (Ouvrir, Shell)
// 			case "g":
// 				if len(items) > 0 && m.cursor < len(items) {
// 					selected := items[m.cursor]
// 					cmd := exec.Command("xdg-open", selected.FullPath())
// 					cmd.Start()
// 				}
// 				return m, nil

// 			case "b":
// 				if len(items) > 0 && m.cursor < len(items) {
// 					selected := items[m.cursor]
// 					targetPath := selected.FullPath()
// 					if !selected.IsDir {
// 						targetPath = filepath.Dir(targetPath)
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

// 			// Navigation verticale (Curseur)
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
// 					if m.cursor >= m.yOffset+visibleHeight {
// 						m.yOffset = m.cursor - visibleHeight + 1
// 					}
// 				}

// 			// Navigation rapide (Home, End, PageUp, PageDown)
// 			case "home":
// 				m.cursor = 0
// 				m.yOffset = 0

// 			case "end":
// 				m.cursor = len(items) - 1
// 				if m.cursor < 0 {
// 					m.cursor = 0
// 				}
// 				if m.cursor >= visibleHeight {
// 					m.yOffset = m.cursor - visibleHeight + 1
// 				} else {
// 					m.yOffset = 0
// 				}

// 			case "pgup":
// 				m.cursor -= 47
// 				if m.cursor < 0 {
// 					m.cursor = 0
// 				}
// 				if m.cursor < m.yOffset {
// 					m.yOffset = m.cursor
// 				}

// 			case "pgdown":
// 				m.cursor += 47
// 				if m.cursor >= len(items) {
// 					m.cursor = len(items) - 1
// 				}
// 				if m.cursor >= m.yOffset+visibleHeight {
// 					m.yOffset = m.cursor - visibleHeight + 1
// 				}
// 			}
// 		}

// 	// Gestion des messages asynchrones (Fin de scan, Refresh)
// 	case scanFinishedMsg:
// 		if msg.err != nil {
// 			m.err = msg.err
// 			m.state = StateInputPath
// 		} else {
// 			m.root = msg.root
// 			m.currentNode = msg.root
// 			m.diskTotalSize = msg.diskSize
// 			m.state = StateBrowsing
// 			m.applySort()
// 		}

// 	case refreshFinishedMsg:
// 		if msg.err != nil {
// 			m.err = msg.err
// 			m.state = StateBrowsing
// 		} else {
// 			newNode := msg.newNode
// 			oldNode := m.currentNode

// 			// Greffe du nouveau nœud scanné dans l'arbre existant
// 			if oldNode.Parent != nil {
// 				newNode.Parent = oldNode.Parent
// 				newNode.Name = filepath.Base(newNode.FullPath())
// 				for i, child := range oldNode.Parent.Children {
// 					if child.Name == oldNode.Name {
// 						oldNode.Parent.Children[i] = newNode
// 						break
// 					}
// 				}
// 				// Mise à jour de la taille totale des parents
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
// 			m.applySort()
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

// // Helper pour déterminer si un slash final est nécessaire pour l'autocomplétion
// func isDirOrShouldSlash(name string) bool {
// 	return strings.HasSuffix(name, string(os.PathSeparator))
// }

// // Commande pour lancer le scan complet d'un répertoire
// func scanDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
// 	return func() tea.Msg {
// 		root, diskSize, err := scanner.Scan(path, exclusions, counter)
// 		return scanFinishedMsg{root: root, diskSize: diskSize, err: err}
// 	}
// }

// // Commande pour rafraîchir un sous-répertoire spécifique
// func refreshDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
// 	return func() tea.Msg {
// 		root, _, err := scanner.Scan(path, exclusions, counter)
// 		return refreshFinishedMsg{newNode: root, err: err}
// 	}
// }

///////////////////////////////////////////::

// package ui

// import (
//     "os"
//     "os/exec"
//     "path/filepath"
//     "sort"
//     "strings"
//     "sync/atomic"

//     "github.com/charmbracelet/bubbles/spinner"
//     "github.com/charmbracelet/bubbles/textinput"
//     tea "github.com/charmbracelet/bubbletea"

//     "github.com/quirky1869/aed/scanner"
// )

// // Messages internes pour la communication asynchrone (fin de scan, refresh)
// type BackMsg struct{}

// type refreshFinishedMsg struct {
//     newNode *scanner.FileNode
//     err     error
// }

// type scanFinishedMsg struct {
//     root     *scanner.FileNode
//     diskSize int64
//     err      error
// }

// // États de la machine à états de l'interface
// type SessionState int

// const (
//     StateInputPath SessionState = iota
//     StateScanning
//     StateBrowsing
// )

// // Modes de tri disponibles pour l'affichage des fichiers
// type SortMode int

// const (
//     SortBySize SortMode = iota
//     SortByName
//     SortByCount
// )

// // Modèle principal contenant l'état de l'application, les données et la configuration UI
// type Model struct {
//     state        SessionState
//     pathInput    textinput.Model
//     excludeInput textinput.Model
//     focusIndex   int

//     spinner spinner.Model

//     root        *scanner.FileNode
//     currentNode *scanner.FileNode
//     cursor      int
//     yOffset     int

//     filesScanned  *int64
//     showHelp      bool
//     diskTotalSize int64

//     currentExclusions []string
//     lang              Language

//     sortMode   SortMode
//     sortDesc   bool
//     showHidden bool

//     suggestions     []string
//     suggestionIndex int
//     suggestionBase  string

//     width, height int
//     err           error
// }

// // Initialisation du modèle avec les valeurs par défaut
// func New(w, h int) Model {
//     currentLang := fr

//     ti := textinput.New()
//     ti.Placeholder = currentLang.PathPlaceholder
//     ti.Focus()
//     ti.CharLimit = 256
//     ti.Width = 50
//     ti.SetValue(".")

//     ei := textinput.New()
//     ei.Placeholder = currentLang.ExcludePlaceholder
//     ei.CharLimit = 256
//     ei.Width = 50

//     s := spinner.New()
//     s.Spinner = spinner.Dot
//     s.Style = spinnerStyle

//     zero := int64(0)

//     return Model{
//         state:        StateInputPath,
//         pathInput:    ti,
//         excludeInput: ei,
//         focusIndex:   0,
//         spinner:      s,
//         filesScanned: &zero,
//         width:        w,
//         height:       h,
//         showHelp:     true,
//         lang:         currentLang,
//         sortMode:     SortBySize,
//         sortDesc:     true,
//         showHidden:   true,
//         suggestions:  []string{},
//     }
// }

// func (m Model) Init() tea.Cmd {
//     return textinput.Blink
// }

// // Trie les enfants du nœud courant selon le mode sélectionné (Taille, Nom, Nombre)
// func (m *Model) applySort() {
//     if m.currentNode == nil || len(m.currentNode.Children) == 0 {
//         return
//     }

//     sort.Slice(m.currentNode.Children, func(i, j int) bool {
//         a := m.currentNode.Children[i]
//         b := m.currentNode.Children[j]

//         isLess := false
//         switch m.sortMode {
//         case SortByName:
//             isLess = strings.ToLower(a.Name) < strings.ToLower(b.Name)
//         case SortByCount:
//             isLess = a.FileCount < b.FileCount
//         default:
//             isLess = a.Size < b.Size
//         }

//         if m.sortDesc {
//             return !isLess
//         }
//         return isLess
//     })
// }

// // Boucle principale de gestion des événements
// func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//     var cmd tea.Cmd

//     switch msg := msg.(type) {

//     case BackMsg:
//         return m, tea.Quit

//     case tea.WindowSizeMsg:
//         m.width = msg.Width
//         m.height = msg.Height

//     case tea.KeyMsg:
//         if msg.String() == "ctrl+c" {
//             return m, func() tea.Msg { return BackMsg{} }
//         }

//         // Bascule de la langue (FR <-> EN)
//         if msg.String() == "ctrl+l" {
//             if m.lang.Code == "FR" {
//                 m.lang = en
//             } else {
//                 m.lang = fr
//             }
//             m.pathInput.Placeholder = m.lang.PathPlaceholder
//             m.excludeInput.Placeholder = m.lang.ExcludePlaceholder
//             return m, nil
//         }

//         // Gestion de l'écran de saisie (Input)
//         if m.state == StateInputPath {
//             switch msg.String() {

//             // Gestion de l'autocomplétion (Tab)
//             case "tab":
//                 var activeInput *textinput.Model
//                 isExcludeField := false

//                 if m.focusIndex == 0 {
//                     activeInput = &m.pathInput
//                 } else {
//                     activeInput = &m.excludeInput
//                     isExcludeField = true
//                 }

//                 // Initialisation du cycle de suggestions
//                 if len(m.suggestions) == 0 {
//                     fullText := activeInput.Value()
//                     searchStr := fullText
//                     prefix := ""

//                     if isExcludeField {
//                         lastComma := strings.LastIndex(fullText, ",")
//                         if lastComma != -1 {
//                             prefix = fullText[:lastComma+1] + " "
//                             searchStr = strings.TrimSpace(fullText[lastComma+1:])
//                         }
//                     }

//                     baseDir, candidates := GetAutocompleteSuggestions(searchStr, true)

//                     if len(candidates) > 0 {
//                         m.suggestions = candidates
//                         m.suggestionBase = prefix + baseDir
//                         m.suggestionIndex = 0

//                         candidate := candidates[0]
//                         newValue := m.suggestionBase + candidate
//                         if isDirOrShouldSlash(candidate) && !strings.HasSuffix(newValue, string(os.PathSeparator)) {
//                             newValue += string(os.PathSeparator)
//                         }

//                         activeInput.SetValue(newValue)
//                         activeInput.SetCursor(len(newValue))
//                     }
//                 } else {
//                     // Cycle suivant
//                     m.suggestionIndex = (m.suggestionIndex + 1) % len(m.suggestions)
//                     candidate := m.suggestions[m.suggestionIndex]
//                     newValue := m.suggestionBase + candidate
//                     if isDirOrShouldSlash(candidate) && !strings.HasSuffix(newValue, string(os.PathSeparator)) {
//                         newValue += string(os.PathSeparator)
//                     }
//                     activeInput.SetValue(newValue)
//                     activeInput.SetCursor(len(newValue))
//                 }
//                 return m, nil

//             // Gestion de l'autocomplétion inverse (Shift+Tab)
//             case "shift+tab":
//                 if len(m.suggestions) > 0 {
//                     var activeInput *textinput.Model
//                     if m.focusIndex == 0 {
//                         activeInput = &m.pathInput
//                     } else {
//                         activeInput = &m.excludeInput
//                     }
//                     m.suggestionIndex--
//                     if m.suggestionIndex < 0 {
//                         m.suggestionIndex = len(m.suggestions) - 1
//                     }
//                     candidate := m.suggestions[m.suggestionIndex]
//                     newValue := m.suggestionBase + candidate
//                     if isDirOrShouldSlash(candidate) && !strings.HasSuffix(newValue, string(os.PathSeparator)) {
//                         newValue += string(os.PathSeparator)
//                     }
//                     activeInput.SetValue(newValue)
//                     activeInput.SetCursor(len(newValue))
//                     return m, nil
//                 }
//                 // Sinon navigation champ précédent
//                 m.suggestions = nil
//                 m.focusIndex = 0
//                 m.excludeInput.Blur()
//                 m.pathInput.Focus()
//                 return m, textinput.Blink

//             // Navigation entre les champs
//             case "up":
//                 m.suggestions = nil
//                 m.focusIndex = 0
//                 m.excludeInput.Blur()
//                 m.pathInput.Focus()
//                 return m, textinput.Blink

//             case "down":
//                 m.suggestions = nil
//                 m.focusIndex = 1
//                 m.pathInput.Blur()
//                 m.excludeInput.Focus()
//                 return m, textinput.Blink

//             // Validation et lancement du scan
//             case "enter":
//                 m.suggestions = nil
//                 rawInput := m.pathInput.Value()
//                 path := rawInput
//                 if path == "." {
//                     path, _ = os.Getwd()
//                 }
//                 path = scanner.ExpandPath(path)

//                 rawExcludes := m.excludeInput.Value()
//                 var exclusions []string
//                 if strings.TrimSpace(rawExcludes) != "" {
//                     parts := strings.Split(rawExcludes, ",")
//                     for _, p := range parts {
//                         trimmed := strings.TrimSpace(p)
//                         expanded := scanner.ExpandPath(trimmed)
//                         expanded = strings.TrimRight(expanded, string(os.PathSeparator))
//                         exclusions = append(exclusions, expanded)
//                     }
//                 }
//                 m.currentExclusions = exclusions

//                 m.state = StateScanning
//                 atomic.StoreInt64(m.filesScanned, 0)
//                 visitedInodes := make(map[scanner.FileID]struct{})

//                 return m, tea.Batch(
//                     m.spinner.Tick,
//                     scanDirectoryCmd(path, m.filesScanned, visitedInodes, exclusions),
//                 )

//             case "esc":
//                 return m, func() tea.Msg { return BackMsg{} }

//             default:
//                 m.suggestions = nil
//             }

//             if m.focusIndex == 0 {
//                 m.pathInput, cmd = m.pathInput.Update(msg)
//             } else {
//                 m.excludeInput, cmd = m.excludeInput.Update(msg)
//             }
//             return m, cmd
//         }

//         // Gestion de l'état Scanning
//         if m.state == StateScanning {
//             if msg.String() == "q" || msg.String() == "esc" {
//                 return m, func() tea.Msg { return BackMsg{} }
//             }
//         }

//         // Gestion de l'état Browsing (Navigation dans les résultats)
//         if m.state == StateBrowsing {
//             items := m.getDisplayItems()

//             footerHeight := 2
//             if !m.showHelp {
//                 footerHeight = 0
//             }
//             visibleHeight := m.height - 6 - footerHeight
//             if visibleHeight < 1 {
//                 visibleHeight = 1
//             }

//             switch msg.String() {
//             case "q":
//                 return m, func() tea.Msg { return BackMsg{} }
//             case ",", "?":
//                 m.showHelp = !m.showHelp
//                 return m, nil

//             // Affichage / Masquage des fichiers cachés
//             case "e":
//                 m.showHidden = !m.showHidden
//                 newItems := m.getDisplayItems()
//                 if m.cursor >= len(newItems) {
//                     m.cursor = len(newItems) - 1
//                     if m.cursor < 0 {
//                         m.cursor = 0
//                     }
//                 }
//                 if m.cursor < m.yOffset {
//                     m.yOffset = m.cursor
//                 }
//                 return m, nil

//             // Options de tri (Taille, Nom, Nombre)
//             case "s":
//                 if m.sortMode == SortBySize {
//                     m.sortDesc = !m.sortDesc
//                 } else {
//                     m.sortMode = SortBySize
//                     m.sortDesc = true
//                 }
//                 m.applySort()
//                 return m, nil

//             case "n":
//                 if m.sortMode == SortByName {
//                     m.sortDesc = !m.sortDesc
//                 } else {
//                     m.sortMode = SortByName
//                     m.sortDesc = false
//                 }
//                 m.applySort()
//                 return m, nil

//             case "C":
//                 if m.sortMode == SortByCount {
//                     m.sortDesc = !m.sortDesc
//                 } else {
//                     m.sortMode = SortByCount
//                     m.sortDesc = true
//                 }
//                 m.applySort()
//                 return m, nil

//             // Rafraîchir le dossier courant
//             case "r":
//                 if m.currentNode != nil {
//                     m.state = StateScanning
//                     atomic.StoreInt64(m.filesScanned, 0)
//                     visitedInodes := make(map[scanner.FileID]struct{})
//                     return m, tea.Batch(
//                         m.spinner.Tick,
//                         refreshDirectoryCmd(m.currentNode.FullPath(), m.filesScanned, visitedInodes, m.currentExclusions),
//                     )
//                 }
//                 return m, nil

//             // Navigation hiérarchique (Parent/Enfant)
//             case "left", "h":
//                 if m.currentNode.Parent != nil {
//                     m.currentNode = m.currentNode.Parent
//                     m.cursor = 0
//                     m.yOffset = 0
//                     m.applySort()
//                 }
//                 return m, nil

//             case "backspace", "esc":
//                 m.state = StateInputPath
//                 m.focusIndex = 0
//                 m.pathInput.Focus()
//                 m.excludeInput.Blur()
//                 return m, nil

//             case "enter", "right", "l":
//                 if len(items) > 0 && m.cursor < len(items) {
//                     selected := items[m.cursor]
//                     if selected.Name == ".." {
//                         if m.currentNode.Parent != nil {
//                             m.currentNode = m.currentNode.Parent
//                             m.cursor = 0
//                             m.yOffset = 0
//                             m.applySort()
//                         }
//                         return m, nil
//                     }
//                     if selected.Name == "." {
//                         return m, nil
//                     }
//                     if selected.IsDir {
//                         for _, child := range m.currentNode.Children {
//                             if child.Name == selected.Name {
//                                 m.currentNode = child
//                                 m.cursor = 0
//                                 m.yOffset = 0
//                                 m.applySort()
//                                 break
//                             }
//                         }
//                     }
//                 }

//             // Actions système (Ouvrir, Shell)
//             case "g":
//                 if len(items) > 0 && m.cursor < len(items) {
//                     selected := items[m.cursor]
                    
//                     // --- MODIFICATION ICI : GESTION DU CAS "." POUR XDG-OPEN ---
//                     var pathOpen string
//                     if selected.Name == "." {
//                         // Si c'est ".", on veut ouvrir le dossier COURANT, pas remonter
//                         pathOpen = m.currentNode.FullPath()
//                     } else {
//                         pathOpen = selected.FullPath()
//                     }
//                     

//                     cmd := exec.Command("xdg-open", pathOpen)
//                     cmd.Start()
//                 }
//                 return m, nil

//             case "b":
//                 if len(items) > 0 && m.cursor < len(items) {
//                     selected := items[m.cursor]
                    
//                     // --- MODIFICATION ICI : GESTION DU CAS "." POUR LE SHELL ---
//                     var targetPath string
//                     if selected.Name == "." {
//                         // Si c'est ".", le target est le dossier COURANT
//                         targetPath = m.currentNode.FullPath()
//                     } else {
//                         targetPath = selected.FullPath()
//                         if !selected.IsDir {
//                             targetPath = filepath.Dir(targetPath)
//                         }
//                     }
//                     

//                     shell := os.Getenv("SHELL")
//                     if shell == "" {
//                         shell = "/bin/bash"
//                     }
//                     c := exec.Command(shell)
//                     c.Dir = targetPath
//                     return m, tea.ExecProcess(c, func(err error) tea.Msg { return nil })
//                 }
//                 return m, nil

//             // Navigation verticale (Curseur)
//             case "up", "k":
//                 if m.cursor > 0 {
//                     m.cursor--
//                     if m.cursor < m.yOffset {
//                         m.yOffset = m.cursor
//                     }
//                 }
//             case "down", "j":
//                 if m.cursor < len(items)-1 {
//                     m.cursor++
//                     if m.cursor >= m.yOffset+visibleHeight {
//                         m.yOffset = m.cursor - visibleHeight + 1
//                     }
//                 }

//             // Navigation rapide (Home, End, PageUp, PageDown)
//             case "home":
//                 m.cursor = 0
//                 m.yOffset = 0

//             case "end":
//                 m.cursor = len(items) - 1
//                 if m.cursor < 0 {
//                     m.cursor = 0
//                 }
//                 if m.cursor >= visibleHeight {
//                     m.yOffset = m.cursor - visibleHeight + 1
//                 } else {
//                     m.yOffset = 0
//                 }

//             case "pgup":
//                 m.cursor -= 47
//                 if m.cursor < 0 {
//                     m.cursor = 0
//                 }
//                 if m.cursor < m.yOffset {
//                     m.yOffset = m.cursor
//                 }

//             case "pgdown":
//                 m.cursor += 47
//                 if m.cursor >= len(items) {
//                     m.cursor = len(items) - 1
//                 }
//                 if m.cursor >= m.yOffset+visibleHeight {
//                     m.yOffset = m.cursor - visibleHeight + 1
//                 }
//             }
//         }

//     // Gestion des messages asynchrones (Fin de scan, Refresh)
//     case scanFinishedMsg:
//         if msg.err != nil {
//             m.err = msg.err
//             m.state = StateInputPath
//         } else {
//             m.root = msg.root
//             m.currentNode = msg.root
//             m.diskTotalSize = msg.diskSize
//             m.state = StateBrowsing
//             m.applySort()
//         }

//     case refreshFinishedMsg:
//         if msg.err != nil {
//             m.err = msg.err
//             m.state = StateBrowsing
//         } else {
//             newNode := msg.newNode
//             oldNode := m.currentNode

//             // Greffe du nouveau nœud scanné dans l'arbre existant
//             if oldNode.Parent != nil {
//                 newNode.Parent = oldNode.Parent
//                 newNode.Name = filepath.Base(newNode.FullPath())
//                 for i, child := range oldNode.Parent.Children {
//                     if child.Name == oldNode.Name {
//                         oldNode.Parent.Children[i] = newNode
//                         break
//                     }
//                 }
//                 // Mise à jour de la taille totale des parents
//                 diff := newNode.Size - oldNode.Size
//                 parent := newNode.Parent
//                 for parent != nil {
//                     parent.Size += diff
//                     parent = parent.Parent
//                 }
//             } else {
//                 m.root = newNode
//             }

//             m.currentNode = newNode
//             if m.cursor >= len(m.getDisplayItems()) {
//                 m.cursor = 0
//                 m.yOffset = 0
//             }
//             m.state = StateBrowsing
//             m.applySort()
//         }

//     case spinner.TickMsg:
//         if m.state == StateScanning {
//             var cmdSpinner tea.Cmd
//             m.spinner, cmdSpinner = m.spinner.Update(msg)
//             return m, cmdSpinner
//         }
//     }

//     return m, nil
// }

// // Helper pour déterminer si un slash final est nécessaire pour l'autocomplétion
// func isDirOrShouldSlash(name string) bool {
//     return strings.HasSuffix(name, string(os.PathSeparator))
// }

// // Commande pour lancer le scan complet d'un répertoire
// func scanDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
//     return func() tea.Msg {
//         root, diskSize, err := scanner.Scan(path, exclusions, counter)
//         return scanFinishedMsg{root: root, diskSize: diskSize, err: err}
//     }
// }

// // Commande pour rafraîchir un sous-répertoire spécifique
// func refreshDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
//     return func() tea.Msg {
//         root, _, err := scanner.Scan(path, exclusions, counter)
//         return refreshFinishedMsg{newNode: root, err: err}
//     }
// }

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

    "github.com/quirky1869/aed/scanner"
)

// Messages internes pour la communication asynchrone (fin de scan, refresh)
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

// États de la machine à états de l'interface
type SessionState int

const (
    StateInputPath SessionState = iota
    StateScanning
    StateBrowsing
)

// Modes de tri disponibles pour l'affichage des fichiers
type SortMode int

const (
    SortBySize SortMode = iota
    SortByName
    SortByCount
)

// Modèle principal contenant l'état de l'application, les données et la configuration UI
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

    sortMode   SortMode
    sortDesc   bool
    showHidden bool

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

// Trie les enfants du nœud courant selon le mode sélectionné (Taille, Nom, Nombre)
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

// Boucle principale de gestion des événements
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

        // Bascule de la langue (FR <-> EN)
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

        // Gestion de l'écran de saisie (Input)
        if m.state == StateInputPath {
            switch msg.String() {

            // Gestion de l'autocomplétion (Tab)
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
                    // Cycle suivant
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

            // Gestion de l'autocomplétion inverse (Shift+Tab)
            case "shift+tab":
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
                // Sinon navigation champ précédent
                m.suggestions = nil
                m.focusIndex = 0
                m.excludeInput.Blur()
                m.pathInput.Focus()
                return m, textinput.Blink

            // Navigation entre les champs
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

            // Validation et lancement du scan
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

        // Gestion de l'état Scanning
        if m.state == StateScanning {
            if msg.String() == "q" || msg.String() == "esc" {
                return m, func() tea.Msg { return BackMsg{} }
            }
        }

        // Gestion de l'état Browsing (Navigation dans les résultats)
        if m.state == StateBrowsing {
            items := m.getDisplayItems()

            footerHeight := 2
            if !m.showHelp {
                footerHeight = 0
            }
            visibleHeight := m.height - 6 - footerHeight
            if visibleHeight < 1 {
                visibleHeight = 1
            }

            switch msg.String() {
            case "q":
                return m, func() tea.Msg { return BackMsg{} }
            case ",", "?":
                m.showHelp = !m.showHelp
                return m, nil

            // Affichage / Masquage des fichiers cachés
            case "e":
                m.showHidden = !m.showHidden
                newItems := m.getDisplayItems()
                if m.cursor >= len(newItems) {
                    m.cursor = len(newItems) - 1
                    if m.cursor < 0 {
                        m.cursor = 0
                    }
                }
                if m.cursor < m.yOffset {
                    m.yOffset = m.cursor
                }
                return m, nil

            // Options de tri (Taille, Nom, Nombre)
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

            // Rafraîchir le dossier courant
            case "r":
                if m.currentNode != nil {
                    m.state = StateScanning
                    atomic.StoreInt64(m.filesScanned, 0)
                    visitedInodes := make(map[scanner.FileID]struct{})
                    return m, tea.Batch(
                        m.spinner.Tick,
                        refreshDirectoryCmd(m.currentNode.FullPath(), m.filesScanned, visitedInodes, m.currentExclusions),
                    )
                }
                return m, nil

            // Navigation hiérarchique (Parent/Enfant)
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
                            if child.Name == selected.Name {
                                m.currentNode = child
                                m.cursor = 0
                                m.yOffset = 0
                                m.applySort()
                                break
                            }
                        }
                    }
                }

            // Actions système (Ouvrir, Shell)
            case "g":
                if len(items) > 0 && m.cursor < len(items) {
                    selected := items[m.cursor]

                    
                    var pathOpen string
                    if selected.Name == "." {
                        pathOpen = m.currentNode.FullPath()
                    } else if selected.Name == ".." {
                        // Si c'est "..", on ouvre le parent du dossier courant
                        if m.currentNode.Parent != nil {
                            pathOpen = m.currentNode.Parent.FullPath()
                        } else {
                            // Sécurité si on est à la racine
                            pathOpen = m.currentNode.FullPath()
                        }
                    } else {
                        pathOpen = selected.FullPath()
                    }
                    

                    cmd := exec.Command("xdg-open", pathOpen)
                    cmd.Start()
                }
                return m, nil

            case "b":
                if len(items) > 0 && m.cursor < len(items) {
                    selected := items[m.cursor]

                    var targetPath string
                    if selected.Name == "." {
                        targetPath = m.currentNode.FullPath()
                    } else if selected.Name == ".." {
                        // Si c'est "..", on ouvre le terminal dans le parent
                        if m.currentNode.Parent != nil {
                            targetPath = m.currentNode.Parent.FullPath()
                        } else {
                            targetPath = m.currentNode.FullPath()
                        }
                    } else {
                        targetPath = selected.FullPath()
                        if !selected.IsDir {
                            targetPath = filepath.Dir(targetPath)
                        }
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

            // Navigation verticale (Curseur)
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
                    if m.cursor >= m.yOffset+visibleHeight {
                        m.yOffset = m.cursor - visibleHeight + 1
                    }
                }

            // Navigation rapide (Home, End, PageUp, PageDown)
            case "home":
                m.cursor = 0
                m.yOffset = 0

            case "end":
                m.cursor = len(items) - 1
                if m.cursor < 0 {
                    m.cursor = 0
                }
                if m.cursor >= visibleHeight {
                    m.yOffset = m.cursor - visibleHeight + 1
                } else {
                    m.yOffset = 0
                }

            case "pgup":
                m.cursor -= 47
                if m.cursor < 0 {
                    m.cursor = 0
                }
                if m.cursor < m.yOffset {
                    m.yOffset = m.cursor
                }

            case "pgdown":
                m.cursor += 47
                if m.cursor >= len(items) {
                    m.cursor = len(items) - 1
                }
                if m.cursor >= m.yOffset+visibleHeight {
                    m.yOffset = m.cursor - visibleHeight + 1
                }
            }
        }

    // Gestion des messages asynchrones (Fin de scan, Refresh)
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

            // Greffe du nouveau nœud scanné dans l'arbre existant
            if oldNode.Parent != nil {
                newNode.Parent = oldNode.Parent
                newNode.Name = filepath.Base(newNode.FullPath())
                for i, child := range oldNode.Parent.Children {
                    if child.Name == oldNode.Name {
                        oldNode.Parent.Children[i] = newNode
                        break
                    }
                }
                // Mise à jour de la taille totale des parents
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

// Helper pour déterminer si un slash final est nécessaire pour l'autocomplétion
func isDirOrShouldSlash(name string) bool {
    return strings.HasSuffix(name, string(os.PathSeparator))
}

// Commande pour lancer le scan complet d'un répertoire
func scanDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
    return func() tea.Msg {
        root, diskSize, err := scanner.Scan(path, exclusions, counter)
        return scanFinishedMsg{root: root, diskSize: diskSize, err: err}
    }
}

// Commande pour rafraîchir un sous-répertoire spécifique
func refreshDirectoryCmd(path string, counter *int64, visited map[scanner.FileID]struct{}, exclusions []string) tea.Cmd {
    return func() tea.Msg {
        root, _, err := scanner.Scan(path, exclusions, counter)
        return refreshFinishedMsg{newNode: root, err: err}
    }
}