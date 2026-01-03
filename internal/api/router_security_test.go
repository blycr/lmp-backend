package api

import (
	"net/http/httptest"
	"testing"

	"lms/backend/internal/config"
)

func TestSecurityHeadersCSP(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Auth:   config.AuthConfig{EnableDeviceAuth: false},
	}
	r := SetupRouter(func() *config.Config { return cfg })
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	if w.Header().Get("Content-Security-Policy") == "" {
		t.Fatalf("missing CSP header")
	}
}

