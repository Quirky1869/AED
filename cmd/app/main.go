package main

import (
	"fmt"
	"os"

	"github.com/Quirky1869/aed/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialisation du programme avec le modèle UI et activation du mode plein écran
	p := tea.NewProgram(ui.New(80, 24), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur lors de l'exécution : %v\n", err)
		os.Exit(1)
	}
}