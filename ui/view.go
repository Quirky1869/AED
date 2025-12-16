package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"

	"aed/scanner"
)

func (m Model) View() string {

	// Vue 1 : Saisie du chemin et des exclusions
	if m.state == StateInputPath {
		title := titleStyle.Render(m.lang.Title)
		var pathLabel, excludeLabel string
		// Note: helpDescStyle remplace helpStyle ici
		if m.focusIndex == 0 {
			pathLabel = helpDescStyle.Render(m.lang.PathLabelActive)
			excludeLabel = inactiveStyle.Render(m.lang.ExcludeLabelInactive)
		} else {
			pathLabel = inactiveStyle.Render(m.lang.PathLabelInactive)
			excludeLabel = helpDescStyle.Render(m.lang.ExcludeLabelActive)
		}
		inputView := m.pathInput.View()
		excludeView := m.excludeInput.View()
		return fmt.Sprintf(
			"\n  %s\n\n  %s\n  %s\n\n  %s\n  %s\n\n  %s",
			title, pathLabel, inputView, excludeLabel, excludeView, helpDescStyle.Render(m.lang.HelpInput),
		)
	}

	// Vue 2 : Spinner pendant le chargement
	if m.state == StateScanning {
		count := atomic.LoadInt64(m.filesScanned)
		return fmt.Sprintf(
			"\n  %s %s\n\n%s %s\n\n  %s",
			m.spinner.View(), m.lang.ScanningTitle,
			countStyle.Render(fmt.Sprintf("%d", count)), m.lang.FilesScanned,
			helpDescStyle.Render(m.lang.HelpScanning),
		)
	}

	// Vue 3 : Explorateur de fichiers
	if m.state == StateBrowsing {
		if m.currentNode == nil {
			return m.lang.ErrorEmpty
		}

		title := titleStyle.Render(m.lang.Title)
		path := pathStyle.Render(m.currentNode.Path)
		totalSize := infoStyle.Render(fmt.Sprintf("%s: %s", m.lang.TotalLabel, formatBytes(m.currentNode.Size)))

		var diskSizeStr string
		if m.diskTotalSize > 0 {
			diskSizeStr = infoStyle.Render(fmt.Sprintf("%s: %s", m.lang.DiskLabel, formatBytes(m.diskTotalSize)))
		}

		var sortName string
		switch m.sortMode {
		case SortByName:
			sortName = m.lang.SortName
		case SortByCount:
			sortName = m.lang.SortCount
		default:
			sortName = m.lang.SortSize
		}

		arrow := "↓"
		if !m.sortDesc {
			arrow = "↑"
		}

		sortStr := sortStyle.Render(fmt.Sprintf("%s: %s %s", m.lang.SortLabel, sortName, arrow))

		header := fmt.Sprintf("  %s  %s  (%s)  (%s)  [%s]\n", title, path, totalSize, diskSizeStr, sortStr)

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

		// Construction du footer bicolore ---
		var footer string
		if m.showHelp {
			footer = renderFooter(m.lang.HelpFooterShort)
		} else {
			footer = renderFooter(m.lang.HelpFooterFull)
		}

		return fmt.Sprintf("\n%s\n%s\n%s", header, content, footer)
	}

	return ""
}

// Fonction utilitaire pour générer le footer avec deux couleurs
func renderFooter(lines [][]HelpItem) string {
	var sb strings.Builder
	sep := helpDescStyle.Render(" • ") // Le séparateur en rose

	for _, line := range lines {
		sb.WriteString("\n ") // Saut de ligne et marge gauche
		var parts []string

		for _, item := range line {
			var part string
			// Si la clé est vide (ex: "Trier par ="), on n'affiche que la description
			if item.Key == "" {
				part = helpDescStyle.Render(item.Desc)
			} else {
				// Clé en Cyan, Description en Rose
				key := helpKeyStyle.Render(item.Key)
				desc := helpDescStyle.Render(": " + item.Desc)
				part = key + desc
			}
			parts = append(parts, part)
		}
		sb.WriteString(strings.Join(parts, sep))
	}
	return sb.String()
}

// Récupère les éléments à afficher en filtrant les fichiers cachés si nécessaire
func (m Model) getDisplayItems() []*scanner.FileNode {
	var items []*scanner.FileNode
	if m.currentNode == nil {
		return items
	}
	dot := &scanner.FileNode{Name: ".", Path: m.currentNode.Path, Size: m.currentNode.Size, IsDir: true}
	items = append(items, dot)
	if m.currentNode.Parent != nil {
		parentPath := filepath.Dir(m.currentNode.Path)
		dotdot := &scanner.FileNode{Name: "..", Path: parentPath, Size: 0, IsDir: true}
		items = append(items, dotdot)
	}

	for _, child := range m.currentNode.Children {
		if !m.showHidden && strings.HasPrefix(child.Name, ".") {
			continue
		}
		items = append(items, child)
	}
	return items
}

// Formateur de taille (Bytes -> TiB)
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