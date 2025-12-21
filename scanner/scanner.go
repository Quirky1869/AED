package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

// Identifiant unique pour dédupliquer les fichiers (hardlinks)
type FileID struct {
	Dev uint64
	Ino uint64
}

// Structure allégée (sans Path)
type FileNode struct {
	Name      string
	Size      int64
	FileCount int64
	IsDir     bool
	Children  []*FileNode
	Parent    *FileNode
}

// Reconstruit le chemin complet à la volée (pour l'affichage)
func (n *FileNode) FullPath() string {
	if n.Parent == nil {
		return n.Name
	}
	parts := []string{}
	curr := n
	for curr != nil {
		parts = append([]string{curr.Name}, parts...)
		curr = curr.Parent
	}
	full := filepath.Join(parts...)
	if len(parts) > 0 && parts[0] == "/" && !strings.HasPrefix(full, "/") {
		full = "/" + full
	}
	return full
}

// Sémaphore pour limiter le parallélisme au nombre de cœurs
var maxWorkers = runtime.NumCPU() * 2
var semaphore = make(chan struct{}, maxWorkers)

func Scan(path string, exclusions []string, counter *int64) (*FileNode, int64, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, 0, err
	}

	// Map partagée pour les hardlinks
	visited := make(map[FileID]struct{})
	var visitedMu sync.Mutex

	// Lancement du scan
	root, err := scanFast(absPath, nil, counter, visited, &visitedMu, exclusions)
	
	return root, GetPartitionSize(absPath), err
}

func scanFast(path string, parent *FileNode, counter *int64, visited map[FileID]struct{}, visitedMu *sync.Mutex, exclusions []string) (*FileNode, error) {
	atomic.AddInt64(counter, 1)

	name := filepath.Base(path)
	if parent == nil {
		name = path
	}

	node := &FileNode{
		Name:   name,
		IsDir:  true,
		Parent: parent,
	}

	// 1. Lecture du dossier (Syscall opendir/readdir)
	entries, err := os.ReadDir(path)
	if err != nil {
		return node, nil
	}

	// Pré-allocation mémoire
	node.Children = make([]*FileNode, 0, len(entries))

	var (
		totalSize  int64
		totalCount int64
		wg         sync.WaitGroup
		mu         sync.Mutex // Protège l'écriture dans node.Children
	)

	// Listes temporaires pour éviter de verrouiller 'mu' à chaque fichier
	var localFiles []*FileNode
	var localSize int64
	var localCount int64

	for _, entry := range entries {
		// Exclusion rapide
		if len(exclusions) > 0 {
			if isExcluded(entry.Name(), exclusions) {
				continue
			}
		}

		childPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			// Protection boucles infinies
			if path == "/" && (entry.Name() == "proc" || entry.Name() == "sys" || entry.Name() == "dev" || entry.Name() == "run") {
				continue
			}

			wg.Add(1)
			
			// Fonction de traitement d'un sous-dossier
			scanSubDir := func(cp string) {
				defer wg.Done()
				childNode, _ := scanFast(cp, node, counter, visited, visitedMu, exclusions)
				if childNode != nil {
					mu.Lock()
					node.Children = append(node.Children, childNode)
					totalSize += childNode.Size
					totalCount += (1 + childNode.FileCount)
					mu.Unlock()
				}
			}

			select {
			case semaphore <- struct{}{}:
				go func(cp string) {
					defer func() { <-semaphore }()
					scanSubDir(cp)
				}(childPath)
			default:
				scanSubDir(childPath)
			}

		} else {
			atomic.AddInt64(counter, 1)
			
			info, err := entry.Info()
			if err != nil {
				continue
			}

			var size int64
			
			// On ne vérifie les hardlinks que si Nlink > 1
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				size = stat.Blocks * 512
				
				// Si Nlink > 1, c'est potentiellement un hardlink partagé, il faut verrouiller
				if stat.Nlink > 1 {
					id := FileID{Dev: stat.Dev, Ino: stat.Ino}
					visitedMu.Lock()
					if _, seen := visited[id]; !seen {
						visited[id] = struct{}{}
						localSize += size // On compte la taille
					}
					visitedMu.Unlock()
				} else {
					// Cas normal (99% des fichiers) : pas de verrou, accès direct !
					localSize += size
				}
			} else {
				size = info.Size()
				localSize += size
			}

			child := &FileNode{
				Name:      entry.Name(),
				Size:      size,
				FileCount: 1,
				IsDir:     false,
				Parent:    node,
			}
			localFiles = append(localFiles, child)
			localCount++
		}
	}

	// Attendre que les sous-dossiers lancés en parallèle aient fini
	wg.Wait()

	// Fusionner les résultats locaux
	mu.Lock()
	node.Children = append(node.Children, localFiles...)
	totalSize += localSize
	totalCount += localCount
	mu.Unlock()
	
	node.Size = totalSize
	node.FileCount = totalCount

	// Tri final
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})

	return node, nil
}

func isExcluded(name string, patterns []string) bool {
	for _, p := range patterns {
		if matched, _ := filepath.Match(p, name); matched {
			return true
		}
	}
	return false
}

func GetPartitionSize(path string) int64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0
	}
	return int64(stat.Blocks) * int64(stat.Bsize)
}

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
		if len(path) > 1 && (path[1] == '/' || path[1] == os.PathSeparator) {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}