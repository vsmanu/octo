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

	// Parse time range
	var from, to time.Time
	now := time.Now()

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	durationStr := r.URL.Query().Get("duration")

	if fromStr != "" && toStr != "" {
		// Parse explicit range
		var err error
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			http.Error(w, "Invalid 'from' time format (RFC3339 required)", http.StatusBadRequest)
			return
		}
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			http.Error(w, "Invalid 'to' time format (RFC3339 required)", http.StatusBadRequest)
			return
		}
	} else {
		// Fallback to duration (default 1h)
		if durationStr == "" {
			durationStr = "1h"
		}
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			http.Error(w, "Invalid duration", http.StatusBadRequest)
			return
		}
		to = now
		from = now.Add(-duration)
	}

	metrics, err := s.storage.QueryHistory(r.Context(), id, from, to)
	if err != nil {
		http.Error(w, "Failed to query history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
