package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"
)

// Identifiant unique pour dédupliquer les fichiers (gestion des hardlinks)
type FileID struct {
	Dev uint64
	Ino uint64
}

// Représentation d'un nœud dans l'arborescence (Fichier ou Dossier)
type FileNode struct {
	Name      string
	Path      string
	Size      int64
	FileCount int64 // Nombre total de fichiers contenus (pour les dossiers)
	IsDir     bool
	Children  []*FileNode
	Parent    *FileNode
}

// Parcourt le disque récursivement, calcule les tailles et le nombre de fichiers
func ScanRecursively(path string, parent *FileNode, counter *int64, visited map[FileID]struct{}, exclusions []string) (*FileNode, error) {
	atomic.AddInt64(counter, 1)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(absPath)
	if parent == nil {
		name = absPath
	}

	node := &FileNode{
		Name:   name,
		Path:   absPath,
		IsDir:  true,
		Parent: parent,
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return node, nil
	}

	var totalSize int64
	var totalCount int64

	for _, entry := range entries {
		childPath := filepath.Join(absPath, entry.Name())

		// Vérification des motifs d'exclusion
		isExcluded := false
		for _, pattern := range exclusions {
			if matched, _ := filepath.Match(pattern, entry.Name()); matched {
				isExcluded = true
				break
			}
			if matched, _ := filepath.Match(pattern, childPath); matched {
				isExcluded = true
				break
			}
		}
		if isExcluded {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			// Appel récursif pour les sous-dossiers
			childNode, _ := ScanRecursively(childPath, node, counter, visited, exclusions)
			if childNode != nil {
				node.Children = append(node.Children, childNode)
				totalSize += childNode.Size
				totalCount += (1 + childNode.FileCount) // On compte le dossier + son contenu
			}
		} else {
			atomic.AddInt64(counter, 1)

			var size int64
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				size = stat.Blocks * 512
				id := FileID{Dev: stat.Dev, Ino: stat.Ino}
				if _, seen := visited[id]; !seen {
					visited[id] = struct{}{}
					totalSize += size
				}
			} else {
				size = info.Size()
				totalSize += size
			}

			child := &FileNode{
				Name:      entry.Name(),
				Path:      childPath,
				Size:      size,
				FileCount: 1,
				IsDir:     false,
				Parent:    node,
			}
			node.Children = append(node.Children, child)
			totalCount++
		}
	}

	node.Size = totalSize
	node.FileCount = totalCount

	// Tri par défaut : Taille décroissante
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})

	return node, nil
}

// Retourne la taille totale de la partition disque
func GetPartitionSize(path string) int64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0
	}
	return int64(stat.Blocks) * int64(stat.Bsize)
}

// Remplace le tilde (~) par le chemin complet du dossier utilisateur
func ExpandPath(path string) string {
	path = os.ExpandEnv(path)
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if path == "~" {
			return home
		}
		if len(path) > 1 && path[1] == '/' {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
