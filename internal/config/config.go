package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds the application configuration
type Config struct {
	WatchPaths []string
	Port       int
}

// ManifestEntry represents a single file entry in the manifest
type ManifestEntry struct {
	Name     string
	Path     string
	Category string
	Metadata map[string]string
}

// Manifest is the top-level manifest structure
type Manifest struct {
	Version string
	Files   []ManifestEntry
}

// NewManifest creates a new Manifest instance
func NewManifest() *Manifest {
	return &Manifest{
		Version: "1.0.0",
		Files:   []ManifestEntry{},
	}
}

// AddFile adds a file to the manifest
func (m *Manifest) AddFile(file ManifestEntry) {
	m.Files = append(m.Files, file)
}

// Project represents a user-defined project with named directories
type Project struct {
	Name      string   `json:"name"`
	Paths     []string `json:"paths"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// ProjectStore manages projects persisted to disk
type ProjectStore struct {
	Projects []Project `json:"projects"`
	Active   string    `json:"active"`
}

// NetworkConfig holds network-related configuration
type NetworkConfig struct {
	BindAddr string // "host:port" or ":port" for all interfaces
	Port     int    // shortcut: port only, bind addr derived
	ShowInfo bool   // print Tailscale IP on startup
}

// DefaultProjectStorePath returns the default path for the project store
func DefaultProjectStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".archterm_projects.json"
	}
	return filepath.Join(home, ".archterm_projects.json")
}

// LoadProjectStore loads projects from disk
func LoadProjectStore(path string) (*ProjectStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProjectStore{
				Projects: []Project{},
				Active:   "",
			}, nil
		}
		return nil, err
	}
	var store ProjectStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

// SaveProjectStore saves projects to disk
func (s *ProjectStore) SaveProjectStore(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// AddProject adds a new project to the store
func (s *ProjectStore) AddProject(name string, paths []string) error {
	for _, p := range s.Projects {
		if p.Name == name {
			return fmt.Errorf("project %s already exists", name)
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	s.Projects = append(s.Projects, Project{
		Name:      name,
		Paths:     paths,
		Active:    false,
		CreatedAt: now,
		UpdatedAt: now,
	})
	return s.SaveProjectStore(DefaultProjectStorePath())
}

// DeleteProject removes a project by name
func (s *ProjectStore) DeleteProject(name string) error {
	for i, p := range s.Projects {
		if p.Name == name {
			s.Projects = append(s.Projects[:i], s.Projects[i+1:]...)
			if s.Active == name {
				s.Active = ""
			}
			return s.SaveProjectStore(DefaultProjectStorePath())
		}
	}
	return fmt.Errorf("project %s not found", name)
}

// ActivateProject sets the active project
func (s *ProjectStore) ActivateProject(name string) error {
	found := false
	for _, p := range s.Projects {
		if p.Name == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("project %s not found", name)
	}
	s.Active = name
	return s.SaveProjectStore(DefaultProjectStorePath())
}

// GetActiveProject returns the currently active project
func (s *ProjectStore) GetActiveProject() *Project {
	for _, p := range s.Projects {
		if p.Name == s.Active {
			return &p
		}
	}
	return nil
}

// GetProject returns a project by name
func (s *ProjectStore) GetProject(name string) *Project {
	for _, p := range s.Projects {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// ListProjects returns all projects
func (s *ProjectStore) ListProjects() []Project {
	return s.Projects
}

// GetActivePaths returns the paths of the active project
func (s *ProjectStore) GetActivePaths() []string {
	if s.Active == "" {
		return []string{}
	}
	for _, p := range s.Projects {
		if p.Name == s.Active {
			return p.Paths
		}
	}
	return []string{}
}

// UpdateProjectPaths updates the paths of an existing project
func (s *ProjectStore) UpdateProjectPaths(name string, paths []string) error {
	for i, p := range s.Projects {
		if p.Name == name {
			s.Projects[i].Paths = paths
			s.Projects[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			return s.SaveProjectStore(DefaultProjectStorePath())
		}
	}
	return fmt.Errorf("project %s not found", name)
}
