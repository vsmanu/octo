package api

import (
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) handleGetEndpointHistory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing endpoint ID", http.StatusBadRequest)
		return
	}

	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "1h"
	}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		http.Error(w, "Invalid duration", http.StatusBadRequest)
		return
	}

	metrics, err := s.storage.QueryHistory(r.Context(), id, duration)
	if err != nil {
		http.Error(w, "Failed to query history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
