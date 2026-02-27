package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/munchies/platform/backend/internal/config"
)

func newTestServer() *Server {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:           8080,
			Environment:    config.EnvLocal,
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}
	return New(cfg)
}

func TestHealthz(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /healthz = %d, want %d", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestReadyz(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /readyz = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestNotFound(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /nonexistent = %d, want %d", w.Code, http.StatusNotFound)
	}
}
