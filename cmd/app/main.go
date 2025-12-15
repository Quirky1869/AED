package main

import (
	"fmt"
	"os"

	"aed/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(ui.New(80, 24), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur lors de l'exécution : %v\n", err)
		os.Exit(1)
	}
}
