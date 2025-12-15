package ui

// Language définit tous les textes dynamiques de l'interface
type Language struct {
	Code string

	// Input View
	Title                string
	PathLabelActive      string
	PathLabelInactive    string
	PathPlaceholder      string
	ExcludeLabelActive   string
	ExcludeLabelInactive string
	ExcludePlaceholder   string
	HelpInput            string

	// Scanning View
	ScanningTitle string
	FilesScanned  string
	HelpScanning  string

	// Browsing View
	ErrorEmpty string
	TotalLabel string
	DiskLabel  string
	SortLabel  string // NOUVEAU : Label "Tri :"

	// NOUVEAU : Noms des modes de tri
	SortSize  string
	SortName  string
	SortCount string

	// Footers
	HelpFooterShort string
	HelpFooterFull  string
}

// Dictionnaire Français
var fr = Language{
	Code: "FR",

	Title:                "AED - Analyseur d'Espace Disque",
	PathLabelActive:      "Entrez le dossier à analyser :",
	PathLabelInactive:    "Entrez le dossier à analyser :",
	PathPlaceholder:      "/home/user (ou ~)",
	ExcludeLabelActive:   "Exclure (fichiers/dossiers, sép. par virgules) :",
	ExcludeLabelInactive: "Exclure (fichiers/dossiers, sép. par virgules) :",
	ExcludePlaceholder:   "node_modules, .git, *.tmp",
	HelpInput:            "(tab: suivant • enter: valider • L: langue • esc: quitter)",

	ScanningTitle: "Analyse en cours...",
	FilesScanned:  "fichiers scannés",
	HelpScanning:  "Appuyer sur q pour quitter",

	ErrorEmpty: "Erreur: Node vide",
	TotalLabel: "Total",
	DiskLabel:  "Disque",
	SortLabel:  "Tri", // NOUVEAU

	SortSize:  "Taille",
	SortName:  "Nom",
	SortCount: "Éléments",

	HelpFooterShort: "\n ?: aide • ↑/↓/←/→: naviguer • enter: sélectionner • q: quitter",
	HelpFooterFull: "\n ?: réduire aide • ↑/↓/←/→: naviguer • enter: sélectionner • q: quitter\n g: explorer • b: shell • r: recalculer • L: langue\n Trier par = s: taille • n: nom • C: éléments",
}

// Dictionnaire Anglais
var en = Language{
	Code: "EN",

	Title:                "DSA - Disk Space Analyzer",
	PathLabelActive:      "Enter directory to analyze:",
	PathLabelInactive:    "Enter directory to analyze:",
	PathPlaceholder:      "/home/user (or ~)",
	ExcludeLabelActive:   "Exclude (files/folders, comma sep.):",
	ExcludeLabelInactive: "Exclude (files/folders, comma sep.):",
	ExcludePlaceholder:   "node_modules, .git, *.tmp",
	HelpInput:            "(tab: next • enter: confirm • L: lang • esc: quit)",

	ScanningTitle: "Scanning in progress...",
	FilesScanned:  "files scanned",
	HelpScanning:  "Press q to quit",

	ErrorEmpty: "Error: Empty Node",
	TotalLabel: "Total",
	DiskLabel:  "Disk",
	SortLabel:  "Sort",

	SortSize:  "Size",
	SortName:  "Name",
	SortCount: "Items",

	HelpFooterShort: "\n ?: help • ↑/↓/←/→: nav • enter: select • q: quit",
	HelpFooterFull:  "\n ?: less help • ↑/↓/←/→: nav • enter: select • q: quit\n g: explore • b: shell • r: refresh • L: lang\n Sort by = s: size • n: name • C: items",
}
