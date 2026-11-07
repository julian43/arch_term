package filesystem

import (
	"io/fs"
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
}

// NewWatcher creates a new Watcher instance
func NewWatcher(m *config.Manifest) *Watcher {
	return &Watcher{
		manifest: m,
		paths:    []string{"./docs", "./plans"},
		changes:  make(chan string, 100),
	}
}

// Start begins watching for file changes
func (w *Watcher) Start() error {
	for _, path := range w.paths {
		if err := filepath.Walk(path, w.walkFn); err != nil {
			return err
		}
	}
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
	category := "documents"
	switch ext {
	case ".pdf":
		category = "pdfs"
	case ".md", ".markdown":
		category = "markdown"
	case ".go":
		category = "code"
	case ".json":
		category = "config"
	case ".txt":
		category = "documents"
	}

	entry := config.ManifestEntry{
		Name:     info.Name(),
		Path:     relPath,
		Category: category,
		Metadata: map[string]string{
			"size":     string(rune(info.Size())),
			"modified": info.ModTime().Format("2006-01-02T15:04:05"),
		},
	}
	w.manifest.AddFile(entry)
}
