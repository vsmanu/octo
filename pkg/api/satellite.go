package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/manu/octo/pkg/checker"
	"github.com/manu/octo/pkg/config"
)

// handleSatelliteHeartbeat updates the satellite status
func (s *Server) handleSatelliteHeartbeat(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth check (API Key)

	// Get ID from context or request (for now assumes ID is passed in header or query,
	// but strictly it should be via auth token/key)
	// Simplified: query param or header
	satelliteID := r.Header.Get("X-Satellite-ID")
	if satelliteID == "" {
		http.Error(w, "Missing Satellite ID", http.StatusBadRequest)
		return
	}

	s.satelliteManager.RegisterHeartbeat(satelliteID)
	w.WriteHeader(http.StatusOK)
}

// handleSatelliteConfig returns the configuration for a specific satellite
func (s *Server) handleSatelliteConfig(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth check

	satelliteID := r.Header.Get("X-Satellite-ID")
	if satelliteID == "" {
		http.Error(w, "Missing Satellite ID", http.StatusBadRequest)
		return
	}

	// Filter endpoints for this satellite
	// Logic: If endpoint has no satellites defined -> Master only (or default)
	// If endpoint has satellites -> Check if this ID is in the list

	cfg := s.configManager.GetConfig()
	var satelliteEndpoints []config.EndpointConfig

	for _, ep := range cfg.Endpoints {
		if shouldRunOnSatellite(ep, satelliteID) {
			satelliteEndpoints = append(satelliteEndpoints, ep)
		}
	}

	log.Printf("Serving config for satellite '%s': %d endpoints found", satelliteID, len(satelliteEndpoints))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(satelliteEndpoints)
}

func shouldRunOnSatellite(endpoint config.EndpointConfig, satelliteID string) bool {
	// If empty, Master only (default)
	if len(endpoint.Satellites) == 0 {
		return false
	}

	for _, s := range endpoint.Satellites {
		if s == satelliteID || s == "all" {
			return true
		}
	}

	return false
}

// handleSatelliteResults receives check results from satellites
func (s *Server) handleSatelliteResults(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth check

	var results []checker.Result
	if err := json.NewDecoder(r.Body).Decode(&results); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Debug Log
	if len(results) > 0 {
		log.Printf("Received %d results. First result SatelliteID: '%s'", len(results), results[0].SatelliteID)
	}

	// Store results
	for _, res := range results {
		// Enriched with satellite ID?
		// Storage needs to support satellite ID.
		// For now, just write.
		if err := s.storage.WriteResult(res); err != nil {
			// Log error but continue
		}
	}

	w.WriteHeader(http.StatusOK)
}
