package ui

import "github.com/charmbracelet/lipgloss"

var (

	// Couleur du titre
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4")).Bold(true)

	// Couleur du logo pendant le chargement ( ⡿ )
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))

	// Couleur du spinner pendant le chargement (fichiers scannés)
	countStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff")).Bold(true).PaddingLeft(2)

	// Couleur texte non sélectionné
	inactiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	// Couleur du chemin
	pathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff")).Bold(true)

	// Couleur des informations (Total et Disque)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))

	// Couleur du type de tri
	sortStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffe600"))

	// Couleur de l'indicateur des fichiers cachés
	hiddenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffe600"))

	// Couleur de la ligne sélectionnée
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#2b2b2b")).Bold(true)

	// Couleur de la barre représentant la taille du dossier ( ■■■■■ )
	barFull = lipgloss.NewStyle().Foreground(lipgloss.Color("#500AFF"))

	// Couleur si la taille du dossier est vide/petite ( ----- )
	barEmpty = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	// Couleur du nombre d'éléments/items présent dans un dossier
	fileCountStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))

	// Couleur pour la description de l'aide
	helpKeyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))

	// Couleur pour les touches d'aide
	helpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))

)
