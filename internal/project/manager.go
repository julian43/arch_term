package project

import (
	"fmt"
	"os"
	"path/filepath"

	"archterm/internal/config"
)

// Manager handles project operations
type Manager struct {
	store *config.ProjectStore
}

// NewManager creates a new project manager
func NewManager() (*Manager, error) {
	store, err := config.LoadProjectStore(config.DefaultProjectStorePath())
	if err != nil {
		return nil, err
	}
	return &Manager{store: store}, nil
}

// List returns all projects
func (m *Manager) List() []config.Project {
	return m.store.ListProjects()
}

// GetActive returns the active project
func (m *Manager) GetActive() *config.Project {
	return m.store.GetActiveProject()
}

// Get returns a project by name
func (m *Manager) Get(name string) *config.Project {
	return m.store.GetProject(name)
}

// Create creates a new project with the given name and paths
func (m *Manager) Create(name string, paths []string) error {
	// Validate paths exist
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("invalid path %s: %v", p, err)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absPath)
		}
	}
	return m.store.AddProject(name, paths)
}

// Delete removes a project by name
func (m *Manager) Delete(name string) error {
	return m.store.DeleteProject(name)
}

// Activate sets the active project
func (m *Manager) Activate(name string) error {
	return m.store.ActivateProject(name)
}

// UpdatePaths updates the paths of an existing project
func (m *Manager) UpdatePaths(name string, paths []string) error {
	// Validate paths exist
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("invalid path %s: %v", p, err)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absPath)
		}
	}
	return m.store.UpdateProjectPaths(name, paths)
}

// GetActivePaths returns the paths of the active project
func (m *Manager) GetActivePaths() []string {
	return m.store.GetActivePaths()
}

// ScanPaths scans the given paths and returns file entries for the manifest
func ScanPaths(paths []string) ([]config.ManifestEntry, error) {
	var entries []config.ManifestEntry
	categoryMap := map[string]string{
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
		".csv":      "data",
		".xml":      "data",
		".sql":      "data",
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Skip hidden directories
				if info.Name()[0] == '.' && info.Name() != "." {
					return filepath.SkipDir
				}
				return nil
			}
			ext := filepath.Ext(filePath)
			category := categoryMap[ext]
			if category == "" {
				category = "other"
			}
			relPath, err := filepath.Rel(absPath, filePath)
			if err != nil {
				relPath = filePath
			}
			entries = append(entries, config.ManifestEntry{
				Name:     info.Name(),
				Path:     relPath,
				Category: category,
				Metadata: map[string]string{
					"size":         fmt.Sprintf("%d", info.Size()),
					"modified":     info.ModTime().Format("2006-01-02T15:04:05Z"),
					"project_path": absPath,
				},
			})
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error scanning path %s: %v", path, err)
		}
	}
	return entries, nil
}
