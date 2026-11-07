package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"archterm/internal/config"
	"archterm/internal/markdown"
	"archterm/internal/project"
	"archterm/internal/search"
)

// Server handles HTTP serving
type Server struct {
	port     int
	manifest *config.Manifest
	searcher *search.SemanticSearcher
	projMgr  *project.Manager
	mux      *http.ServeMux
	srv      *http.Server
}

// NewServer creates a new Server instance
func NewServer(port int, m *config.Manifest) *Server {
	// Create semantic search indexer
	searcher := search.NewSemanticSearcher()

	// Create project manager
	projMgr, err := project.NewManager()
	if err != nil {
		log.Printf("Warning: Could not load project store: %v", err)
		// Create empty store
		emptyStore := &config.ProjectStore{
			Projects: []config.Project{},
			Active:   "",
		}
		projMgr = &project.Manager{}
		_ = emptyStore
	}

	s := &Server{
		port:     port,
		manifest: m,
		searcher: searcher,
		projMgr:  projMgr,
		mux:      http.NewServeMux(),
	}

	// Index manifest files (read actual file content for search)
	for _, f := range m.Files {
		content := ""
		// Only index text-based files (skip binary/images)
		ext := strings.ToLower(filepath.Ext(f.Path))
		textExts := map[string]bool{
			".go": true, ".js": true, ".ts": true, ".py": true, ".rs": true, ".java": true,
			".md": true, ".markdown": true, ".txt": true, ".doc": true, ".docx": true,
			".json": true, ".yaml": true, ".yml": true, ".toml": true, ".ini": true, ".cfg": true,
			".csv": true, ".xml": true, ".sql": true, ".db": true, ".pdf": true,
		}
		if textExts[ext] {
			if data, err := os.ReadFile(filepath.Join(".", f.Path)); err == nil {
				content = string(data)
			}
		}

		searcher.AddDocument(search.Document{
			ID:       f.Path,
			Title:    f.Name,
			Content:  content,
			Source:   f.Path,
			Category: f.Category,
		})
	}

	s.registerRoutes()
	return s
}

// Start begins serving HTTP requests
func (s *Server) Start() error {
	log.Printf("Starting server on port %d...", s.port)
	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
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
	s.mux.HandleFunc("/api/related", s.handleRelated)
	s.mux.HandleFunc("/api/concepts", s.handleConcepts)

	// Project routes
	s.mux.HandleFunc("/api/projects", s.handleProjects)
	s.mux.HandleFunc("/api/projects/", s.handleProjectAction)

	// Serve static files
	s.mux.HandleFunc("/", s.handleStatic)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Handle root path - serve index.html
	if r.URL.Path == "/" || r.URL.Path == "" {
		http.ServeFile(w, r, "./webapp/index.html")
		return
	}

	// Build the file path from the webapp directory
	filePath := filepath.Join("./webapp", r.URL.Path)

	// Prevent directory traversal attacks
	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, filepath.Clean("./webapp")) {
		http.NotFound(w, r)
		return
	}

	// Check if file exists
	if _, err := os.Stat(cleanPath); err != nil {
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, cleanPath)
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

	// Check if rendered output is requested
	render := r.URL.Query().Get("render")
	if render == "true" || render == "1" {
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".md" || ext == ".markdown" {
			renderer := markdown.NewRenderer()
			html := renderer.Render(string(data))
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(html))
			return
		}
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

	// Use semantic search
	semanticResults := s.searcher.Search(query, 10)

	type SearchResult struct {
		Document search.Document `json:"document"`
		Score    float64         `json:"score"`
	}
	results := make([]SearchResult, 0)
	for _, r := range semanticResults {
		results = append(results, SearchResult{
			Document: r.Document,
			Score:    r.Score,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"results": results,
	})
}

func (s *Server) handleRelated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	docID := r.URL.Query().Get("doc")
	if docID == "" {
		http.Error(w, "Missing required parameter 'doc'", http.StatusBadRequest)
		return
	}

	maxResults := 10
	if maxStr := r.URL.Query().Get("max"); maxStr != "" {
		fmt.Sscanf(maxStr, "%d", &maxResults)
		if maxResults <= 0 {
			maxResults = 10
		}
	}

	relationships := s.searcher.GetRelatedDocuments(docID, maxResults)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"source":        docID,
		"relationships": relationships,
	})
}

func (s *Server) handleConcepts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	docID := r.URL.Query().Get("doc")
	if docID == "" {
		http.Error(w, "Missing required parameter 'doc'", http.StatusBadRequest)
		return
	}

	concepts := s.searcher.GetConceptsForDocument(docID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"source":   docID,
		"concepts": concepts,
	})
}

// Project handlers

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listProjects(w, r)
	case http.MethodPost:
		s.createProject(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleProjectAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Missing project name", http.StatusBadRequest)
		return
	}

	projectName := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "activate":
		if r.Method == http.MethodPost {
			s.activateProject(w, r, projectName)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "paths":
		if r.Method == http.MethodPut {
			s.updateProjectPaths(w, r, projectName)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "scan":
		if r.Method == http.MethodPost {
			s.scanProject(w, r, projectName)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		if r.Method == http.MethodDelete {
			s.deleteProject(w, r, projectName)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	projects := s.projMgr.List()
	active := s.projMgr.GetActive()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"projects": projects,
		"active":   active,
	})
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string   `json:"name"`
		Paths []string `json:"paths"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	if len(req.Paths) == 0 {
		http.Error(w, "At least one path is required", http.StatusBadRequest)
		return
	}

	if err := s.projMgr.Create(req.Name, req.Paths); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Project '%s' created successfully", req.Name),
	})
}

func (s *Server) activateProject(w http.ResponseWriter, r *http.Request, name string) {
	if err := s.projMgr.Activate(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Project '%s' activated", name),
	})
}

func (s *Server) updateProjectPaths(w http.ResponseWriter, r *http.Request, name string) {
	var req struct {
		Paths []string `json:"paths"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(req.Paths) == 0 {
		http.Error(w, "At least one path is required", http.StatusBadRequest)
		return
	}

	if err := s.projMgr.UpdatePaths(name, req.Paths); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Project '%s' paths updated", name),
	})
}

func (s *Server) scanProject(w http.ResponseWriter, r *http.Request, name string) {
	proj := s.projMgr.Get(name)
	if proj == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	entries, err := project.ScanPaths(proj.Paths)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"project": name,
		"files":   entries,
		"count":   len(entries),
	})
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request, name string) {
	if err := s.projMgr.Delete(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Project '%s' deleted", name),
	})
}
