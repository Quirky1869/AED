package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func GetAutocompleteSuggestions(currentPath string, onlyDirs bool) (string, []string) {
	// 1. Gestion du tilde (~)
	expanded := currentPath
	home, _ := os.UserHomeDir()
	usedTilde := false

	if strings.HasPrefix(currentPath, "~") {
		usedTilde = true
		if currentPath == "~" {
			expanded = home
		} else if len(currentPath) > 1 {
			expanded = filepath.Join(home, currentPath[2:])
		}
	}

	// 2. Séparer le dossier parent et le début du fichier en cours de frappe
	dir, filePrefix := filepath.Split(expanded)
	
	// Si on tape juste un nom sans /, on est dans le dossier courant
	if dir == "" {
		dir = "."
	}

	// 3. Lire le contenu du dossier
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", nil
	}

	// 4. Trouver les candidats
	var candidates []string
	for _, entry := range entries {
		// Ignore les fichiers cachés sauf si on a tapé un "."
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

	// On trie pour que l'ordre soit logique (alphabétique)
	sort.Strings(candidates)

	// 5. Reconstruction du préfixe de retour
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