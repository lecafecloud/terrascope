// Package handlers provides HTTP request handlers for the API endpoints.
// It defines the routing logic, response formatting, and error handling mechanisms.
package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	t.Run("returns 200 OK for GET request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns correct content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("returns valid JSON response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response.Status)
		assert.Equal(t, "terrascope-api", response.Service)
		assert.NotEmpty(t, response.Timestamp)
		assert.NotEmpty(t, response.Uptime)
	})

	t.Run("timestamp is in RFC3339 format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		_, err = time.Parse(time.RFC3339, response.Timestamp)
		assert.NoError(t, err, "Timestamp should be valid RFC3339 format")
	})

	t.Run("includes Go version in details", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotNil(t, response.Details)
		assert.Equal(t, runtime.Version(), response.Details["go_version"])
	})

	t.Run("includes CPU count in details", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Details["num_cpu"])
	})

	t.Run("uptime increases over time", func(t *testing.T) {
		req1 := httptest.NewRequest(http.MethodGet, "/health", nil)
		w1 := httptest.NewRecorder()
		HealthHandler(w1, req1)

		var response1 HealthResponse
		err := json.NewDecoder(w1.Body).Decode(&response1)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
		w2 := httptest.NewRecorder()
		HealthHandler(w2, req2)

		var response2 HealthResponse
		err = json.NewDecoder(w2.Body).Decode(&response2)
		require.NoError(t, err)

		assert.NotEqual(t, response1.Uptime, response2.Uptime)
	})

	t.Run("returns 405 for POST request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Contains(t, w.Body.String(), "Method not allowed")
	})

	t.Run("returns 405 for PUT request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("returns 405 for DELETE request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("returns 405 for PATCH request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("response structure matches HealthResponse", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Status)
		assert.NotEmpty(t, response.Timestamp)
		assert.NotEmpty(t, response.Service)
		assert.NotEmpty(t, response.Uptime)
		assert.NotNil(t, response.Details)
	})

	t.Run("handles multiple concurrent requests", func(t *testing.T) {
		numRequests := 10
		results := make(chan int, numRequests)

		for range numRequests {
			go func() {
				req := httptest.NewRequest(http.MethodGet, "/health", nil)
				w := httptest.NewRecorder()
				HealthHandler(w, req)
				results <- w.Code
			}()
		}

		for range numRequests {
			code := <-results
			assert.Equal(t, http.StatusOK, code)
		}
	})

	t.Run("timestamp is recent", func(t *testing.T) {
		before := time.Now().UTC().Add(-1 * time.Second)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		HealthHandler(w, req)

		after := time.Now().UTC().Add(1 * time.Second)

		var response HealthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		timestamp, err := time.Parse(time.RFC3339, response.Timestamp)
		require.NoError(t, err)

		assert.True(t, timestamp.After(before) || timestamp.Equal(before))
		assert.True(t, timestamp.Before(after) || timestamp.Equal(after))
	})

	t.Run("handles HEAD request as not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodHead, "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHealthResponse_JSONMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal preserves data", func(t *testing.T) {
		original := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Service:   "terrascope-api",
			Uptime:    "1h30m",
			Details: map[string]string{
				"go_version": runtime.Version(),
				"num_cpu":    "8",
			},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded HealthResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, original.Status, decoded.Status)
		assert.Equal(t, original.Timestamp, decoded.Timestamp)
		assert.Equal(t, original.Service, decoded.Service)
		assert.Equal(t, original.Uptime, decoded.Uptime)
		assert.Equal(t, original.Details["go_version"], decoded.Details["go_version"])
	})

	t.Run("omitempty fields are omitted when empty", func(t *testing.T) {
		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Service:   "terrascope-api",
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		jsonString := string(data)
		assert.NotContains(t, jsonString, "uptime")
		assert.NotContains(t, jsonString, "details")
	})
}

func TestStartTime(t *testing.T) {
	t.Run("startTime is initialized", func(t *testing.T) {
		assert.False(t, startTime.IsZero())
	})

	t.Run("startTime is in the past", func(t *testing.T) {
		assert.True(t, startTime.Before(time.Now()))
	})
}
