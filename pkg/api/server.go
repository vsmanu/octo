package api

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/manu/octo/pkg/config"
	"github.com/manu/octo/pkg/storage"
)

type Server struct {
	configManager *config.Manager
	storage       storage.Provider
	frontendFS    fs.FS
}

func NewServer(cfgMgr *config.Manager, store storage.Provider, frontendFS fs.FS) *Server {
	return &Server{
		configManager: cfgMgr,
		storage:       store,
		frontendFS:    frontendFS,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /metrics", s.handleMetrics) // Placeholder for Prometheus
	mux.HandleFunc("GET /api/v1/config", s.handleGetConfig)
	mux.HandleFunc("POST /api/v1/config", s.handleUpdateConfig)
	mux.HandleFunc("POST /api/v1/config/endpoints", s.handleCreateEndpoint)
	mux.HandleFunc("PUT /api/v1/config/endpoints/{id}", s.handleUpdateEndpoint)
	mux.HandleFunc("DELETE /api/v1/config/endpoints/{id}", s.handleDeleteEndpoint)
	mux.HandleFunc("GET /api/v1/endpoints/{id}/history", s.handleGetEndpointHistory)

	// Frontend (SPA)
	if s.frontendFS != nil {
		fileServer := http.FileServer(http.FS(s.frontendFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			// If it's an API route that wasn't matched above (shouldn't happen if patterns are correct, but good safety)
			if strings.HasPrefix(path, "/api/") {
				http.NotFound(w, r)
				return
			}

			// Check if file exists in FS
			f, err := s.frontendFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				defer f.Close()
				stat, _ := f.Stat()
				if !stat.IsDir() {
					fileServer.ServeHTTP(w, r)
					return
				}
			}

			// Fallback to index.html for SPA
			index, err := s.frontendFS.Open("index.html")
			if err != nil {
				http.Error(w, "Frontend not found", http.StatusInternalServerError)
				return
			}
			defer index.Close()
			http.ServeContent(w, r, "index.html", s.getModTime(index), index.(io.ReadSeeker))
		})
	}

	return mux
}

// Helper to get ModTime, avoiding Stat if possible or handling errors
func (s *Server) getModTime(f fs.File) (t time.Time) {
	stat, err := f.Stat()
	if err == nil {
		return stat.ModTime()
	}
	return time.Now()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Integrate Prometheus exporter
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("# Metrics placeholder\n"))
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := s.configManager.GetConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newCfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := s.configManager.UpdateConfig(func(current *config.Config) error {
		*current = newCfg // Replace config
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to update config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"updated"}`))
}
