package filesystem

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"archterm/internal/config"
)

// Watcher watches for file changes
type Watcher struct {
	manifest *config.Manifest
	paths    []string
	changes  chan string
	// categoryMap maps file extensions to categories
	categoryMap map[string]string
}

// NewWatcher creates a new Watcher instance
func NewWatcher(m *config.Manifest) *Watcher {
	w := &Watcher{
		manifest: m,
		paths:    []string{"./docs", "./plans", "./cmd", "./internal", "./webapp", "./python_parser"},
		changes:  make(chan string, 100),
		categoryMap: map[string]string{
			".pdf":      "documents",
			".md":       "documents",
			".markdown": "documents",
			".txt":      "documents",
			".doc":      "documents",
			".docx":     "documents",
			".go":       "code",
			".js":       "code",
			".ts":       "code",
			".py":       "code",
			".rs":       "code",
			".java":     "code",
			".json":     "config",
			".yaml":     "config",
			".yml":      "config",
			".toml":     "config",
			".ini":      "config",
			".cfg":      "config",
			".png":      "images",
			".jpg":      "images",
			".jpeg":     "images",
			".gif":      "images",
			".svg":      "images",
			".ico":      "images",
			".csv":      "data",
			".xml":      "data",
			".sql":      "data",
			".db":       "data",
		},
	}
	return w
}

// Start begins watching for file changes
func (w *Watcher) Start() error {
	log.Println("Starting file watcher...")
	for _, path := range w.paths {
		log.Printf("Watching path: %s", path)
		if err := filepath.Walk(path, w.walkFn); err != nil {
			log.Printf("Warning: error walking %s: %v", path, err)
		}
	}
	log.Printf("Initial scan complete. Indexed %d files.", len(w.manifest.Files))
	return nil
}

// walkFn is the filepath.WalkFunc for walking the directory tree
func (w *Watcher) walkFn(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return filepath.SkipDir
	}
	if !info.IsDir() {
		w.registerFile(path, info)
	}
	return nil
}

// registerFile registers a file in the manifest
func (w *Watcher) registerFile(path string, info os.FileInfo) {
	relPath, err := filepath.Rel(".", path)
	if err != nil {
		relPath = path
	}

	ext := strings.ToLower(filepath.Ext(path))
	category := w.getCategory(ext)

	// Skip binary files and hidden directories
	if strings.HasPrefix(info.Name(), ".") {
		return
	}
	if isBinary(path, info) {
		return
	}

	entry := config.ManifestEntry{
		Name:     info.Name(),
		Path:     relPath,
		Category: category,
		Metadata: map[string]string{
			"size":     fmt.Sprintf("%d", info.Size()),
			"modified": info.ModTime().Format("2006-01-02T15:04:05"),
			"mode":     info.Mode().String(),
		},
	}
	w.manifest.AddFile(entry)
}

// getCategory returns the category for a file extension
func (w *Watcher) getCategory(ext string) string {
	if cat, ok := w.categoryMap[ext]; ok {
		return cat
	}
	return "other"
}

// isBinary checks if a file is likely binary by reading the first 512 bytes
func isBinary(path string, info os.FileInfo) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return true
	}

	// Check for null bytes (indicator of binary file)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}
