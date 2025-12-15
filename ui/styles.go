package ui

import "github.com/charmbracelet/lipgloss"

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