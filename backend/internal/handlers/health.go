// Package handlers provides HTTP request handlers for the API endpoints.
// It defines the routing logic, response formatting, and error handling mechanisms.
package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Service   string            `json:"service"`
	Uptime    string            `json:"uptime,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

var startTime = time.Now()

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "terrascope-api",
		Uptime:    time.Since(startTime).String(),
		Details: map[string]string{
			"go_version": runtime.Version(),
			"num_cpu":    string(rune(runtime.NumCPU())),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
