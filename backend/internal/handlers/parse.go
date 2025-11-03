// Package handlers provides HTTP request handlers for the API endpoints.
// It defines the routing logic, response formatting, and error handling mechanisms.
package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/terrascope/core/internal/parser"
)

func ParseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	state, err := parser.ParseTfstate(body)
	if err != nil {
		http.Error(w, "Invalid tfstate: "+err.Error(), http.StatusBadRequest)
		return
	}

	graph := parser.BuildGraph(state)

	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	if r.URL.Query().Get("pretty") == "true" {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(graph); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
