package middleware

import (
	"lms/backend/internal/auth"
	"lms/backend/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthRequiredPassesWithValidToken(t *testing.T) {
	m := auth.NewManager(1 * time.Minute)
	auth.SetDefaultManager(m)
	code := m.GenerateDeviceCode("phone")
	sess, ok := m.VerifyDeviceCode(code.Code)
	if !ok {
		t.Fatalf("verify failed")
	}
	provider := func() *config.Config {
		return &config.Config{
			Auth: config.AuthConfig{EnableDeviceAuth: true},
		}
	}
	h := AuthRequired(provider)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+sess.Token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuthRequiredBlocksWithoutToken(t *testing.T) {
	m := auth.NewManager(1 * time.Minute)
	auth.SetDefaultManager(m)
	provider := func() *config.Config {
		return &config.Config{
			Auth: config.AuthConfig{EnableDeviceAuth: true},
		}
	}
	h := AuthRequired(provider)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthRequiredDisabledSkips(t *testing.T) {
	auth.SetDefaultManager(auth.NewManager(1 * time.Minute))
	provider := func() *config.Config {
		return &config.Config{
			Auth: config.AuthConfig{EnableDeviceAuth: false},
		}
	}
	h := AuthRequired(provider)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuthRequiredNoManagerBlocks(t *testing.T) {
	provider := func() *config.Config {
		return &config.Config{
			Auth: config.AuthConfig{EnableDeviceAuth: true},
		}
	}
	auth.SetDefaultManager(nil)
	h := AuthRequired(provider)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
