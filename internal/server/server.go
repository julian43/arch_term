package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"archterm/internal/config"
)

// Server handles HTTP serving
type Server struct {
	port     int
	manifest *config.Manifest
	mux      *http.ServeMux
	srv      *http.Server
}

// NewServer creates a new Server instance
func NewServer(port int, m *config.Manifest) *Server {
	s := &Server{
		port:     port,
		manifest: m,
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Start begins serving HTTP requests
func (s *Server) Start() error {
	addr := s.port
	log.Printf("Starting server on port %d...", addr)
	s.srv = &http.Server{
		Addr:         `:8080`,
		Handler:      s.mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) registerRoutes() {
	// API routes
	s.mux.HandleFunc("/api/manifest", s.handleManifest)
	s.mux.HandleFunc("/api/files/", s.handleFiles)
	s.mux.HandleFunc("/api/ls", s.handleLS)
	s.mux.HandleFunc("/api/search", s.handleSearch)

	// Serve static files
	s.mux.HandleFunc("/", s.handleStatic)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("./webapp", r.URL.Path)
	if path == "/" {
		path = "/webapp/index.html"
	}
	if _, err := os.Stat(path); err == nil {
		http.StripPrefix("/webapp/", http.FileServer(http.Dir(filepath.Dir(path)))).ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.manifestResponse(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) manifestResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.manifest)
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/files/")
	if path == "" {
		http.Error(w, "Missing file path", http.StatusBadRequest)
		return
	}

	// Read the file
	data, err := os.ReadFile(filepath.Join(".", path))
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Return the file content
	w.Header().Set("Content-Type", "text/plain")
	w.Write(data)
}

func (s *Server) handleLS(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		category = "all"
	}

	// Return the manifest
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"category": category,
		"files":    s.manifest.Files,
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	// Simple search implementation
	results := make([]config.ManifestEntry, 0)
	lowerQuery := strings.ToLower(query)
	for _, f := range s.manifest.Files {
		if strings.Contains(strings.ToLower(f.Name), lowerQuery) {
			results = append(results, f)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"results": results,
	})
}
