package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/manu/octo/pkg/config"
)

// generateID creates a random ID for new endpoints
func generateID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if random fails (unlikely)
		return fmt.Sprintf("ep-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func (s *Server) handleCreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var newEndpoint config.EndpointConfig
	if err := json.NewDecoder(r.Body).Decode(&newEndpoint); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate basic fields
	if newEndpoint.Name == "" || newEndpoint.URL == "" {
		http.Error(w, "Name and URL are required", http.StatusBadRequest)
		return
	}

	// Generate ID if missing
	if newEndpoint.ID == "" {
		newEndpoint.ID = generateID()
	}

	// Set default interval if missing or zero
	if newEndpoint.Interval == 0 {
		newEndpoint.Interval = 60 * time.Second
	}

	err := s.configManager.UpdateConfig(func(cfg *config.Config) error {
		// Check for duplicate ID
		for _, ep := range cfg.Endpoints {
			if ep.ID == newEndpoint.ID {
				return fmt.Errorf("endpoint with ID %s already exists", newEndpoint.ID)
			}
		}
		cfg.Endpoints = append(cfg.Endpoints, newEndpoint)
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to create endpoint: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newEndpoint)
}

func (s *Server) handleUpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Endpoint ID is required", http.StatusBadRequest)
		return
	}

	var updatedEndpoint config.EndpointConfig
	if err := json.NewDecoder(r.Body).Decode(&updatedEndpoint); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure ID matches path (prevent changing ID via body)
	updatedEndpoint.ID = id

	err := s.configManager.UpdateConfig(func(cfg *config.Config) error {
		found := false
		for i, ep := range cfg.Endpoints {
			if ep.ID == id {
				// Update fields, preserving ID
				cfg.Endpoints[i] = updatedEndpoint
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("endpoint with ID %s not found", id)
		}
		return nil
	})

	if err != nil {
		if err.Error() == fmt.Sprintf("endpoint with ID %s not found", id) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update endpoint: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedEndpoint)
}

func (s *Server) handleDeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Endpoint ID is required", http.StatusBadRequest)
		return
	}

	err := s.configManager.UpdateConfig(func(cfg *config.Config) error {
		found := false
		newEndpoints := make([]config.EndpointConfig, 0, len(cfg.Endpoints))
		for _, ep := range cfg.Endpoints {
			if ep.ID == id {
				found = true
				continue // Skip this one to delete
			}
			newEndpoints = append(newEndpoints, ep)
		}
		if !found {
			return fmt.Errorf("endpoint with ID %s not found", id)
		}
		cfg.Endpoints = newEndpoints
		return nil
	})

	if err != nil {
		if err.Error() == fmt.Sprintf("endpoint with ID %s not found", id) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete endpoint: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
