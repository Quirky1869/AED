package ui

// Structure pour découper l'aide
type HelpItem struct {
	Key  string
	Desc string
}

// Structure contenant tous les textes de l'interface
type Language struct {
	Code string

	Title                string
	PathLabelActive      string
	PathLabelInactive    string
	PathPlaceholder      string
	ExcludeLabelActive   string
	ExcludeLabelInactive string
	ExcludePlaceholder   string
	HelpInput            string

	ScanningTitle string
	FilesScanned  string
	HelpScanning  string

	ErrorEmpty string
	TotalLabel string
	DiskLabel  string

	SortLabel string
	SortSize  string
	SortName  string
	SortCount string

	HiddenFilesLabel string

	HelpFooterShort [][]HelpItem
	HelpFooterFull  [][]HelpItem
}

// Textes en Français
var fr = Language{
	Code: "FR",

	Title:             "AED - Analyseur d'Espace Disque",
	PathLabelActive:   "Entrez le dossier à analyser :",
	PathLabelInactive: "Entrez le dossier à analyser :",
	PathPlaceholder:   "/home/user (ou ~) (ou $HOME)",

	ExcludeLabelActive:   "Exclure (fichiers/dossiers, sép. par virgules) :",
	ExcludeLabelInactive: "Exclure (fichiers/dossiers, sép. par virgules) :",
	ExcludePlaceholder:   "/home, node_modules, .git, *.tmp",
	HiddenFilesLabel:     "Fichiers cachés",

	HelpInput: "(tab: compléter • ↑/↓: options • enter: valider • ctrl+l: langue • esc: quitter)",

	ScanningTitle: "Analyse en cours...",
	FilesScanned:  "fichiers scannés",
	HelpScanning:  "Appuyer sur q pour quitter",

	ErrorEmpty: "Erreur: Node vide",
	TotalLabel: "Total",
	DiskLabel:  "Disque",

	SortLabel: "Tri",
	SortSize:  "Taille",
	SortName:  "Nom",
	SortCount: "Éléments",

	HelpFooterShort: [][]HelpItem{
		{
			{"?", "aide"},
			{"↑/↓/←/→", "naviguer"},
			{"enter", "sélectionner"},
			{"q", "quitter"},
		},
	},

	HelpFooterFull: [][]HelpItem{
		{
			{"?", "réduire aide"},
			{"↑j/↓k/←h/→l", "naviguer"},
			{"enter", "sélectionner"},
			{"esc", "revenir menu"},
			{"q", "quitter"},
		},
		{
			{"g", "explorer"},
			{"b", "shell"},
			{"r", "recalculer"},
			{"e", "fichiers cachés"},
			{"ctrl+l", "langue"},
		},
		{
			{"", "Trier par ="},
			{"s", "taille"},
			{"n", "nom"},
			{"C", "éléments"},
		},
	},
}

// Textes en Anglais
var en = Language{
	Code: "EN",

	Title:                "DSA - Disk Space Analyzer",
	PathLabelActive:      "Enter directory to analyze:",
	PathLabelInactive:    "Enter directory to analyze:",
	PathPlaceholder:      "/home/user (or ~) (or $HOME)",
	ExcludeLabelActive:   "Exclude (files/folders, comma sep.):",
	ExcludeLabelInactive: "Exclude (files/folders, comma sep.):",
	ExcludePlaceholder:   "/home, node_modules, .git, *.tmp",
	HiddenFilesLabel:     "Hidden files",

	HelpInput: "(tab: autocomplete • ↑/↓: options • enter: confirm • ctrl+l: lang • esc: quit)",

	ScanningTitle: "Scanning in progress...",
	FilesScanned:  "files scanned",
	HelpScanning:  "Press q to quit",

	ErrorEmpty: "Error: Empty Node",
	TotalLabel: "Total",
	DiskLabel:  "Disk",

	SortLabel: "Sort",
	SortSize:  "Size",
	SortName:  "Name",
	SortCount: "Items",

	HelpFooterShort: [][]HelpItem{
		{
			{"?", "help"},
			{"↑/↓/←/→", "nav"},
			{"enter", "select"},
			{"q", "quit"},
		},
	},

	HelpFooterFull: [][]HelpItem{
		{
			{"?", "less help"},
			{"↑j/↓k/←h/→l", "nav"},
			{"enter", "select"},
			{"esc", "back menu"},
			{"q", "quit"},
		},
		{
			{"g", "explore"},
			{"b", "shell"},
			{"r", "refresh"},
			{"e", "hidden files"},
			{"ctrl+l", "lang"},
		},
		{
			{"", "Sort by ="},
			{"s", "size"},
			{"n", "name"},
			{"C", "items"},
		},
	},
}
