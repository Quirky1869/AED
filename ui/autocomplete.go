package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Analyse le chemin en cours de saisie pour proposer des complétions basées sur le contenu du disque
// Gère l'expansion du tilde (~) et filtre selon le préfixe déjà tapé
func GetAutocompleteSuggestions(currentPath string, onlyDirs bool) (string, []string) {
	expanded := os.ExpandEnv(currentPath)
	home, _ := os.UserHomeDir()
	usedTilde := false

	// Gestion de l'expansion du répertoire home (~)
	if strings.HasPrefix(currentPath, "~") {
		usedTilde = true
		if currentPath == "~" {
			expanded = home
		} else if len(currentPath) > 1 {
			expanded = filepath.Join(home, currentPath[2:])
		}
	}

	// Séparation du dossier parent et du début du fichier en cours de frappe
	dir, filePrefix := filepath.Split(expanded)

	if dir == "" {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", nil
	}

	// Filtrage des candidats correspondant au préfixe
	var candidates []string
	for _, entry := range entries {
		// On ignore les fichiers cachés sauf si l'utilisateur a explicitement tapé un point
		if strings.HasPrefix(entry.Name(), ".") && !strings.HasPrefix(filePrefix, ".") {
			continue
		}

		if onlyDirs && !entry.IsDir() {
			continue
		}

		if strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(filePrefix)) {
			name := entry.Name()
			if entry.IsDir() {
				name += string(os.PathSeparator)
			}
			candidates = append(candidates, name)
		}
	}

	if len(candidates) == 0 {
		return "", nil
	}

	sort.Strings(candidates)

	// Reconstruction du chemin de base pour l'affichage (réintégration du tilde si utilisé)
	baseToReturn := dir

	if usedTilde {
		if strings.HasPrefix(dir, home) {
			rel, err := filepath.Rel(home, dir)
			if err == nil {
				if rel == "." {
					baseToReturn = "~" + string(os.PathSeparator)
				} else {
					baseToReturn = "~" + string(os.PathSeparator) + rel + string(os.PathSeparator)
				}
			}
		}
	}

	if baseToReturn == "." {
		baseToReturn = ""
	}

	return baseToReturn, candidates
}