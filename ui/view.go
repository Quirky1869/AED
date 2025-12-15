// package ui

// import (
// 	"fmt"
// 	"path/filepath"
// 	"strings"
// 	"sync/atomic"

// 	"aed/scanner"
// )

// func (m Model) View() string {
// 	if m.state == StateInputPath {
// 		title := titleStyle.Render("AED - Analyseur d'Espace Disque")

// 		var pathLabel, excludeLabel string
		
// 		// Logique de couleur pour savoir quel champ est actif
// 		if m.focusIndex == 0 {
// 			pathLabel = helpStyle.Render("Entrez le dossier à analyser :")
// 			excludeLabel = inactiveStyle.Render("Exclure (fichiers/dossiers, sép. par virgules) :")
// 		} else {
// 			pathLabel = inactiveStyle.Render("Entrez le dossier à analyser :")
// 			excludeLabel = helpStyle.Render("Exclure (fichiers/dossiers, sép. par virgules) :")
// 		}

// 		inputView := m.pathInput.View()
// 		excludeView := m.excludeInput.View()

// 		return fmt.Sprintf(
// 			"\n  %s\n\n  %s\n  %s\n\n  %s\n  %s\n\n  %s",
// 			title,
// 			pathLabel,
// 			inputView,
// 			excludeLabel,
// 			excludeView,
// 			helpStyle.Render("(tab: suivant • enter: valider • esc: quitter)"),
// 		)
// 	}

// 	if m.state == StateScanning {
// 		count := atomic.LoadInt64(m.filesScanned)
// 		return fmt.Sprintf(
// 			"\n  %s Analyse en cours...\n\n%s fichiers scannés\n\n  %s",
// 			m.spinner.View(),
// 			countStyle.Render(fmt.Sprintf("%d", count)),
// 			helpStyle.Render("Appuyer sur q pour quitter"),
// 		)
// 	}

// 	if m.state == StateBrowsing {
// 		if m.currentNode == nil {
// 			return "Erreur: Node vide"
// 		}

// 		title := titleStyle.Render("AED")
// 		path := pathStyle.Render(m.currentNode.Path)
// 		totalSize := infoStyle.Render(fmt.Sprintf("Total: %s", formatBytes(m.currentNode.Size)))

// 		var diskSizeStr string
// 		if m.diskTotalSize > 0 {
// 			diskSizeStr = infoStyle.Render(fmt.Sprintf("Disque: %s", formatBytes(m.diskTotalSize)))
// 		}

// 		header := fmt.Sprintf("  %s  %s  (%s)  (%s)\n", title, path, totalSize, diskSizeStr)

// 		footerHeight := 2
// 		if !m.showHelp {
// 			footerHeight = 0
// 		}

// 		visibleHeight := m.height - 5 - footerHeight
// 		if visibleHeight < 1 {
// 			visibleHeight = 1
// 		}

// 		var rows []string
// 		items := m.getDisplayItems()

// 		start := m.yOffset
// 		end := start + visibleHeight
// 		if end > len(items) {
// 			end = len(items)
// 		}

// 		barWidth := 20

// 		for i := start; i < end; i++ {
// 			item := items[i]

// 			var sizeStr, bar, name string

// 			if item.Name == "." || item.Name == ".." {
// 				sizeStr = fmt.Sprintf("%8s", "")
// 				if item.Name == "." {
// 					sizeStr = fmt.Sprintf("%8s", formatBytes(item.Size))
// 				}
// 				bar = strings.Repeat(" ", barWidth)
// 				name = item.Name
// 			} else {
// 				sizeStr = fmt.Sprintf("%8s", formatBytes(item.Size))

// 				percent := 0.0
// 				if m.currentNode.Size > 0 {
// 					percent = float64(item.Size) / float64(m.currentNode.Size)
// 				}
// 				filledLen := int(percent * float64(barWidth))
// 				emptyLen := barWidth - filledLen
// 				bar = barFull.Render(strings.Repeat("■", filledLen)) + barEmpty.Render(strings.Repeat("-", emptyLen))

// 				name = item.Name
// 				if item.IsDir {
// 					name += "/"
// 				}
// 			}

// 			row := fmt.Sprintf("%s  %s  %s", sizeStr, bar, name)

// 			if i == m.cursor {
// 				row = selectedStyle.Render(fmt.Sprintf("%-*s", m.width-4, row))
// 			} else {
// 				row = fmt.Sprintf("  %s", row)
// 			}
// 			rows = append(rows, row)
// 		}

// 		content := strings.Join(rows, "\n")

// 		var footer string
// 		if m.showHelp {
// 			footer = helpStyle.Render("\n ?: aide • ↑/↓/←/→: naviguer • enter: sélectionner • q: quitter")
// 		} else {
// 			footer = helpStyle.Render("\n ?: réduire aide • ↑/↓/←/→: naviguer • enter: sélectionner • q: quitter\n g: explorer • b: shell • r: recalculer")
// 		}

// 		return fmt.Sprintf("\n%s\n%s\n%s", header, content, footer)
// 	}

// 	return ""
// }

// func (m Model) getDisplayItems() []*scanner.FileNode {
// 	var items []*scanner.FileNode

// 	if m.currentNode == nil {
// 		return items
// 	}

// 	dot := &scanner.FileNode{
// 		Name:  ".",
// 		Path:  m.currentNode.Path,
// 		Size:  m.currentNode.Size,
// 		IsDir: true,
// 	}
// 	items = append(items, dot)

// 	if m.currentNode.Parent != nil {
// 		parentPath := filepath.Dir(m.currentNode.Path)
// 		dotdot := &scanner.FileNode{
// 			Name:  "..",
// 			Path:  parentPath,
// 			Size:  0,
// 			IsDir: true,
// 		}
// 		items = append(items, dotdot)
// 	}

// 	items = append(items, m.currentNode.Children...)

// 	return items
// }

// func formatBytes(b int64) string {
// 	const unit = 1024
// 	if b < unit {
// 		return fmt.Sprintf("%d B", b)
// 	}
// 	div, exp := int64(unit), 0
// 	for n := b / unit; n >= unit; n /= unit {
// 		div *= unit
// 		exp++
// 	}
// 	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
// }

package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"

	"aed/scanner"
)

func (m Model) View() string {
	
	// --- VUE INPUT ---
	if m.state == StateInputPath {
		title := titleStyle.Render(m.lang.Title)

		var pathLabel, excludeLabel string
		
		if m.focusIndex == 0 {
			pathLabel = helpStyle.Render(m.lang.PathLabelActive)
			excludeLabel = inactiveStyle.Render(m.lang.ExcludeLabelInactive)
		} else {
			pathLabel = inactiveStyle.Render(m.lang.PathLabelInactive)
			excludeLabel = helpStyle.Render(m.lang.ExcludeLabelActive)
		}

		inputView := m.pathInput.View()
		excludeView := m.excludeInput.View()

		return fmt.Sprintf(
			"\n  %s\n\n  %s\n  %s\n\n  %s\n  %s\n\n  %s",
			title,
			pathLabel,
			inputView,
			excludeLabel,
			excludeView,
			helpStyle.Render(m.lang.HelpInput),
		)
	}

	// --- VUE SCANNING ---
	if m.state == StateScanning {
		count := atomic.LoadInt64(m.filesScanned)
		return fmt.Sprintf(
			"\n  %s %s\n\n%s %s\n\n  %s",
			m.spinner.View(),
			m.lang.ScanningTitle,
			countStyle.Render(fmt.Sprintf("%d", count)),
			m.lang.FilesScanned,
			helpStyle.Render(m.lang.HelpScanning),
		)
	}

	// --- VUE BROWSING ---
	if m.state == StateBrowsing {
		if m.currentNode == nil {
			return m.lang.ErrorEmpty
		}

		title := titleStyle.Render(m.lang.Title) // Ou juste "AED" si vous préférez court
		path := pathStyle.Render(m.currentNode.Path)
		
		// Total: XX GB
		totalSize := infoStyle.Render(fmt.Sprintf("%s: %s", m.lang.TotalLabel, formatBytes(m.currentNode.Size)))

		// Disque: XX GB
		var diskSizeStr string
		if m.diskTotalSize > 0 {
			diskSizeStr = infoStyle.Render(fmt.Sprintf("%s: %s", m.lang.DiskLabel, formatBytes(m.diskTotalSize)))
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
			footer = helpStyle.Render(m.lang.HelpFooterShort)
		} else {
			footer = helpStyle.Render(m.lang.HelpFooterFull)
		}

		return fmt.Sprintf("\n%s\n%s\n%s", header, content, footer)
	}

	return ""
}

func (m Model) getDisplayItems() []*scanner.FileNode {
	var items []*scanner.FileNode

	if m.currentNode == nil {
		return items
	}

	dot := &scanner.FileNode{
		Name:  ".",
		Path:  m.currentNode.Path,
		Size:  m.currentNode.Size,
		IsDir: true,
	}
	items = append(items, dot)

	if m.currentNode.Parent != nil {
		parentPath := filepath.Dir(m.currentNode.Path)
		dotdot := &scanner.FileNode{
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