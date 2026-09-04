package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"golang.org/x/sys/unix"
)

// Identifiant unique pour dédupliquer les fichiers (hardlinks)
type FileID struct {
	Dev uint64
	Ino uint64
}

// Structure allégée (Sans le champ "Path" pour économiser la RAM)
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
	// Hack pour garder le "/" initial si on est sous Unix et que la racine était "/"
	if len(parts) > 0 && parts[0] == "/" && !strings.HasPrefix(full, "/") {
		full = "/" + full
	}
	return full
}

// Sémaphore pour limiter le parallélisme (évite de saturer l'OS)
var maxWorkers = runtime.NumCPU() * 2
var semaphore = make(chan struct{}, maxWorkers)

// Wrapper principal
func Scan(path string, exclusions []string, counter *int64) (*FileNode, int64, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, 0, err
	}

	// Map partagée pour les hardlinks
	visited := make(map[FileID]struct{})
	var visitedMu sync.Mutex

	// Lancement du scan optimisé
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

	// On garde le File (et son fd) ouvert pour pouvoir faire des Fstatat
	// relatifs au dossier plutôt que des Lstat qui reparcourent tout le
	// chemin depuis la racine à chaque fichier.
	dirFile, err := os.Open(path)
	if err != nil {
		return node, nil
	}

	entries, err := dirFile.ReadDir(-1)
	if err != nil {
		dirFile.Close()
		return node, nil
	}
	dfd := int(dirFile.Fd())

	// Pré-allocation
	node.Children = make([]*FileNode, 0, len(entries))

	var (
		totalSize  int64
		totalCount int64
		wg         sync.WaitGroup
		mu         sync.Mutex
	)

	localFiles := make([]*FileNode, 0, len(entries))
	var localSize int64
	var localCount int64

	for _, entry := range entries {
		entryName := entry.Name()
		isDir := entry.IsDir()

		// Le chemin complet ne sert que pour les sous-dossiers ou les
		// exclusions : on évite de l'allouer pour chaque fichier.
		var childPath string
		if isDir || len(exclusions) > 0 {
			childPath = filepath.Join(path, entryName)
		}

		if len(exclusions) > 0 {
			if isExcluded(entryName, exclusions) || isExcluded(childPath, exclusions) {
				continue
			}
		}

		if isDir {
			if path == "/" && (entryName == "proc" || entryName == "sys" || entryName == "dev" || entryName == "run") {
				continue
			}

			wg.Add(1)

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

			var stat unix.Stat_t
			if err := unix.Fstatat(dfd, entryName, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				continue
			}

			// Optimisation hardlinks
			size := stat.Blocks * 512
			if stat.Nlink > 1 {
				id := FileID{Dev: uint64(stat.Dev), Ino: stat.Ino}
				visitedMu.Lock()
				if _, seen := visited[id]; !seen {
					visited[id] = struct{}{}
					localSize += size
				}
				visitedMu.Unlock()
			} else {
				localSize += size
			}

			child := &FileNode{
				Name:      entryName,
				Size:      size,
				FileCount: 1,
				IsDir:     false,
				Parent:    node,
			}
			localFiles = append(localFiles, child)
			localCount++
		}
	}

	// On n'a plus besoin du fd une fois les fichiers de ce dossier statés :
	// on le libère avant d'attendre les sous-dossiers pour ne pas garder
	// trop de fd ouverts simultanément lors d'une récursion profonde.
	dirFile.Close()

	wg.Wait()

	mu.Lock()
	node.Children = append(node.Children, localFiles...)
	totalSize += localSize
	totalCount += localCount
	mu.Unlock()

	node.Size = totalSize
	node.FileCount = totalCount

	if len(node.Children) > 1 {
		slices.SortFunc(node.Children, func(a, b *FileNode) int {
			switch {
			case a.Size > b.Size:
				return -1
			case a.Size < b.Size:
				return 1
			default:
				return 0
			}
		})
	}

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
