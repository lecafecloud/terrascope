package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCors(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("handles OPTIONS preflight request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/parse", nil)
		rec := httptest.NewRecorder()

		Cors(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}

		if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin == "" {
			t.Error("expected Access-Control-Allow-Origin header to be set")
		}

		if methods := rec.Header().Get("Access-Control-Allow-Methods"); methods == "" {
			t.Error("expected Access-Control-Allow-Methods header to be set")
		}
	})

	t.Run("passes POST request to next handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/parse", nil)
		rec := httptest.NewRecorder()

		Cors(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin == "" {
			t.Error("expected Access-Control-Allow-Origin header to be set")
		}
	})

	t.Run("sets CORS headers on all requests", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		Cors(handler).ServeHTTP(rec, req)

		headers := map[string]bool{
			"Access-Control-Allow-Origin":  false,
			"Access-Control-Allow-Methods": false,
			"Access-Control-Allow-Headers": false,
			"Access-Control-Max-Age":       false,
		}

		for header := range headers {
			if rec.Header().Get(header) != "" {
				headers[header] = true
			}
		}

		for header, set := range headers {
			if !set {
				t.Errorf("expected %s header to be set", header)
			}
		}
	})
}
