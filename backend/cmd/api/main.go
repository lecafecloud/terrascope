// Package main starts an HTTP server that provides endpoints for health checks
// and Terraform state parsing. It uses the internal handlers package to process
// incoming requests and return JSON responses.
package main

import (
	"log"
	"net/http"

	"github.com/terrascope/core/internal/handlers"
)

func main() {
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/parse", handlers.ParseHandler)

	log.Printf("🚀 Server starting on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
