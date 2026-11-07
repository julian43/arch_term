package config

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
