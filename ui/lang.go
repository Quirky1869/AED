package ui

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

	HelpFooterShort string
	HelpFooterFull  string
}

// Textes en Français
var fr = Language{
	Code: "FR",

	Title:                "AED - Analyseur d'Espace Disque",
	PathLabelActive:      "Entrez le dossier à analyser :",
	PathLabelInactive:    "Entrez le dossier à analyser :",
	PathPlaceholder:      "/home/user (ou ~)",
	ExcludeLabelActive:   "Exclure (fichiers/dossiers, sép. par virgules) :",
	ExcludeLabelInactive: "Exclure (fichiers/dossiers, sép. par virgules) :",
	ExcludePlaceholder:   "node_modules, .git, *.tmp",
	HelpInput:            "(tab: suivant • enter: valider • ctrl+l: langue • esc: quitter)",

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

	HelpFooterShort: "\n ?: aide • ↑/↓/←/→: naviguer • enter: sélectionner • q: quitter",
	HelpFooterFull:  "\n ?: réduire aide • ↑/↓/←/→: naviguer • enter: sélectionner • q: quitter\n g: explorer • b: shell • r: recalculer • h: fichiers cachés • ctrl+l: langue\n Trier par = s: taille • n: nom • C: éléments",
}

// Textes en Anglais
var en = Language{
	Code: "EN",

	Title:                "DSA - Disk Space Analyzer",
	PathLabelActive:      "Enter directory to analyze:",
	PathLabelInactive:    "Enter directory to analyze:",
	PathPlaceholder:      "/home/user (or ~)",
	ExcludeLabelActive:   "Exclude (files/folders, comma sep.):",
	ExcludeLabelInactive: "Exclude (files/folders, comma sep.):",
	ExcludePlaceholder:   "node_modules, .git, *.tmp",
	HelpInput:            "(tab: next • enter: confirm • ctrl+l: lang • esc: quit)",

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

	HelpFooterShort: "\n ?: help • ↑/↓/←/→: nav • enter: select • q: quit",
	HelpFooterFull:  "\n ?: less help • ↑/↓/←/→: nav • enter: select • q: quit\n g: explore • b: shell • r: refresh • h: hidden files • ctrl+l: lang\n Sort by = s: size • n: name • C: items",
}