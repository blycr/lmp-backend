package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"lms/backend/internal/config"
)

func TestCORSPreflightAllowsLocalhost(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Auth:   config.AuthConfig{EnableDeviceAuth: false},
	}
	r := SetupRouter(func() *config.Config { return cfg })

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/device/code", nil)
	req.Header.Set("Origin", "http://localhost:8455")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8455" {
		t.Fatalf("missing allow origin")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("missing allow methods")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Fatalf("missing allow headers")
	}
}

func TestCORSOnPOSTEchoesOrigin(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Auth:   config.AuthConfig{EnableDeviceAuth: false},
	}
	r := SetupRouter(func() *config.Config { return cfg })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/code", bytes.NewReader([]byte(`{"device_name":"t"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:8455")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8455" {
		t.Fatalf("missing allow origin")
	}
}

