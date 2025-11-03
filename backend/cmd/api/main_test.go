// Package main starts an HTTP server that provides endpoints for health checks
// and Terraform state parsing. It uses the internal handlers package to process
// incoming requests and return JSON responses.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terrascope/core/internal/handlers"
	"github.com/terrascope/core/internal/models"
)

func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/parse", handlers.ParseHandler)
	return mux
}

func TestMainRoutes(t *testing.T) {
	router := setupRouter()

	t.Run("health endpoint is accessible", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("parse endpoint is accessible", func(t *testing.T) {
		validTfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": []
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(validTfstate))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("non-existent route returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("root path returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHealthEndpointIntegration(t *testing.T) {
	router := setupRouter()

	t.Run("health returns valid response structure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response handlers.HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response.Status)
		assert.Equal(t, "terrascope-api", response.Service)
		assert.NotEmpty(t, response.Timestamp)
		assert.NotEmpty(t, response.Uptime)
	})

	t.Run("health endpoint rejects POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestParseEndpointIntegration(t *testing.T) {
	router := setupRouter()

	t.Run("parse returns valid graph", func(t *testing.T) {
		tfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": [
				{
					"mode": "managed",
					"type": "aws_vpc",
					"name": "main",
					"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
					"instances": [{
						"schema_version": 1,
						"attributes": {"id": "vpc-123"}
					}]
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(tfstate))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var graph models.Graph
		err := json.NewDecoder(w.Body).Decode(&graph)
		require.NoError(t, err)

		assert.Len(t, graph.Nodes, 1)
		assert.Equal(t, "aws_vpc.main", graph.Nodes[0].ID)
	})

	t.Run("parse rejects GET requests", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/parse", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("parse rejects invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader("invalid"))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEndToEndFlow(t *testing.T) {
	router := setupRouter()

	t.Run("complete workflow: health check then parse", func(t *testing.T) {
		healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
		healthW := httptest.NewRecorder()
		router.ServeHTTP(healthW, healthReq)
		assert.Equal(t, http.StatusOK, healthW.Code)

		tfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": [
				{
					"mode": "managed",
					"type": "aws_s3_bucket",
					"name": "assets",
					"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
					"instances": [{
						"schema_version": 0,
						"attributes": {"id": "my-bucket"}
					}]
				},
				{
					"mode": "managed",
					"type": "aws_s3_bucket_policy",
					"name": "assets_policy",
					"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
					"instances": [{
						"schema_version": 0,
						"attributes": {"id": "my-bucket-policy"},
						"dependencies": ["aws_s3_bucket.assets"]
					}]
				}
			]
		}`

		parseReq := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(tfstate))
		parseW := httptest.NewRecorder()
		router.ServeHTTP(parseW, parseReq)

		assert.Equal(t, http.StatusOK, parseW.Code)

		var graph models.Graph
		err := json.NewDecoder(parseW.Body).Decode(&graph)
		require.NoError(t, err)

		assert.Len(t, graph.Nodes, 2)
		assert.Len(t, graph.Edges, 1)
	})
}

func TestRoutePaths(t *testing.T) {
	router := setupRouter()

	testCases := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{"health with GET", "/health", http.MethodGet, http.StatusOK},
		{"health with POST", "/health", http.MethodPost, http.StatusMethodNotAllowed},
		{"parse with POST", "/parse", http.MethodPost, http.StatusBadRequest},
		{"parse with GET", "/parse", http.MethodGet, http.StatusMethodNotAllowed},
		{"unknown path", "/unknown", http.MethodGet, http.StatusNotFound},
		{"root path", "/", http.MethodGet, http.StatusNotFound},
		{"health with trailing slash", "/health/", http.MethodGet, http.StatusNotFound},
		{"parse with trailing slash", "/parse/", http.MethodPost, http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestConcurrentRequests(t *testing.T) {
	router := setupRouter()

	t.Run("handles concurrent health checks", func(t *testing.T) {
		numRequests := 50
		results := make(chan int, numRequests)

		for range numRequests {
			go func() {
				req := httptest.NewRequest(http.MethodGet, "/health", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				results <- w.Code
			}()
		}

		for range numRequests {
			code := <-results
			assert.Equal(t, http.StatusOK, code)
		}
	})

	t.Run("handles concurrent parse requests", func(t *testing.T) {
		tfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": []
		}`

		numRequests := 50
		results := make(chan int, numRequests)

		for range numRequests {
			go func() {
				req := httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(tfstate))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				results <- w.Code
			}()
		}

		for range numRequests {
			code := <-results
			assert.Equal(t, http.StatusOK, code)
		}
	})

	t.Run("handles mixed concurrent requests", func(t *testing.T) {
		numRequests := 100
		results := make(chan int, numRequests)

		tfstate := `{
			"version": 4,
			"terraform_version": "1.5.0",
			"serial": 1,
			"lineage": "abc-123",
			"resources": []
		}`

		for i := range numRequests {
			go func(index int) {
				var req *http.Request
				if index%2 == 0 {
					req = httptest.NewRequest(http.MethodGet, "/health", nil)
				} else {
					req = httptest.NewRequest(http.MethodPost, "/parse", strings.NewReader(tfstate))
				}
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				results <- w.Code
			}(i)
		}

		for range numRequests {
			code := <-results
			assert.Equal(t, http.StatusOK, code)
		}
	})
}

func TestContentTypeHeaders(t *testing.T) {
	router := setupRouter()

	t.Run("all successful responses return application/json", func(t *testing.T) {
		tests := []struct {
			name   string
			method string
			path   string
			body   string
		}{
			{"health endpoint", http.MethodGet, "/health", ""},
			{"parse endpoint", http.MethodPost, "/parse", `{"version":4,"terraform_version":"1.5.0","serial":1,"lineage":"abc","resources":[]}`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var req *http.Request
				if tt.body != "" {
					req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				} else {
					req = httptest.NewRequest(tt.method, tt.path, nil)
				}

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				}
			})
		}
	})
}
