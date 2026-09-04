package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/quirky1869/aed/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Le scan génère énormément de petites allocations (un *FileNode par
	// fichier/dossier). On relâche le seuil du GC pour limiter le nombre
	// de cycles pendant le scan, quitte à consommer un peu plus de RAM.
	// Respecte GOGC si l'utilisateur l'a explicitement défini.
	if _, set := os.LookupEnv("GOGC"); !set {
		debug.SetGCPercent(400)
	}

	// Initialisation du programme avec le modèle UI et activation du mode plein écran
	p := tea.NewProgram(ui.New(80, 24), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur lors de l'exécution : %v\n", err)
		os.Exit(1)
	}
}
