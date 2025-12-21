package ui

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/lipgloss"

	"aed/scanner"
)

func (m Model) View() string {

	if m.state == StateInputPath {
		title := titleStyle.Render(m.lang.Title)
		var pathLabel, excludeLabel string
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

	if m.state == StateScanning {
		count := atomic.LoadInt64(m.filesScanned)
		return fmt.Sprintf(
			"\n  %s %s\n\n%s %s\n\n  %s",
			m.spinner.View(), m.lang.ScanningTitle,
			countStyle.Render(fmt.Sprintf("%d", count)), m.lang.FilesScanned,
			helpDescStyle.Render(m.lang.HelpScanning),
		)
	}

	if m.state == StateBrowsing {
		if m.currentNode == nil {
			return m.lang.ErrorEmpty
		}

		title := titleStyle.Render(m.lang.Title)
		path := pathStyle.Render(m.currentNode.FullPath())

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

		sortText := fmt.Sprintf("%s: %s %s", m.lang.SortLabel, sortName, arrow)
		if m.sortMode == SortByCount {
			sortText = fmt.Sprintf("%s: %s %s - %d", m.lang.SortLabel, sortName, arrow, m.currentNode.FileCount)
		}
		sortStr := sortStyle.Render(fmt.Sprintf("[%s]", sortText))

		// Indicateur Fichiers Cachés
		hiddenIcon := "︶"
		if m.showHidden {
			hiddenIcon = "👁 "
		}
		hiddenStr := hiddenStyle.Render(fmt.Sprintf("[%s : %s]", m.lang.HiddenFilesLabel, hiddenIcon))

		totalSize := infoStyle.Render(fmt.Sprintf("(%s: %s)", m.lang.TotalLabel, formatBytes(m.currentNode.Size)))
		var diskSizeStr string
		if m.diskTotalSize > 0 {
			diskSizeStr = infoStyle.Render(fmt.Sprintf("(%s: %s)", m.lang.DiskLabel, formatBytes(m.diskTotalSize)))
		}

		paddingLen := 2 + lipgloss.Width(title) + 2
		padding := strings.Repeat(" ", paddingLen)

		headerLine1 := fmt.Sprintf("  %s  %s  %s  %s", title, path, totalSize, diskSizeStr)
		headerLine2 := fmt.Sprintf("%s%s   %s", padding, sortStr, hiddenStr)

		header := headerLine1 + "\n" + headerLine2 + "\n"

		footerHeight := 2
		if !m.showHelp {
			footerHeight = 0
		}

		visibleHeight := m.height - 6 - footerHeight
		if visibleHeight < 1 {
			visibleHeight = 1
		}

		var rows []string
		items := m.getDisplayItems()

		// Calcul de l'alignement
		maxNameLen := 0
		for _, item := range items {
			length := len(item.Name)
			if item.IsDir {
				length++
			}
			if length > maxNameLen {
				maxNameLen = length
			}
		}

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

				if m.sortMode == SortByCount && item.IsDir {
					currentLen := len(item.Name) + 1
					paddingNeeded := (maxNameLen - currentLen) + 4
					padding := strings.Repeat(" ", paddingNeeded)
					countStr := fileCountStyle.Render(fmt.Sprintf("%d", item.FileCount))
					name += padding + countStr
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
			footer = renderFooter(m.lang.HelpFooterShort)
		} else {
			footer = renderFooter(m.lang.HelpFooterFull)
		}

		return fmt.Sprintf("\n%s\n%s\n%s", header, content, footer)
	}

	return ""
}

func renderFooter(lines [][]HelpItem) string {
	var sb strings.Builder
	sep := helpDescStyle.Render(" • ")

	for _, line := range lines {
		sb.WriteString("\n ")
		var parts []string

		for _, item := range line {
			var part string
			if item.Key == "" {
				part = helpDescStyle.Render(item.Desc)
			} else {
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

func (m Model) getDisplayItems() []*scanner.FileNode {
	var items []*scanner.FileNode
	if m.currentNode == nil {
		return items
	}
	dot := &scanner.FileNode{Name: ".", Size: m.currentNode.Size, FileCount: m.currentNode.FileCount, IsDir: true, Parent: m.currentNode.Parent}
	items = append(items, dot)
	if m.currentNode.Parent != nil {
		dotdot := &scanner.FileNode{Name: "..", Size: 0, IsDir: true}
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