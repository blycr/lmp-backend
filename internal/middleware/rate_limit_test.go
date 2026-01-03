package middleware

import (
	"lms/backend/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitBlocksBurst(t *testing.T) {
	provider := func() *config.Config {
		return &config.Config{
			Server: config.ServerConfig{LANOnly: false},
			Rate:   config.RateLimitConfig{Enabled: true, RequestsPerSecond: 1, Burst: 1},
		}
	}
	h := RateLimit(provider)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.5:8080"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Fatalf("first request should pass, got %d", rr1.Code)
	}
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request should be limited, got %d", rr2.Code)
	}
}

func TestRateLimitDisabledPass(t *testing.T) {
	provider := func() *config.Config {
		return &config.Config{
			Rate: config.RateLimitConfig{Enabled: false},
		}
	}
	h := RateLimit(provider)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.1.1.1:80"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRateLimitEmptyIPBlocks(t *testing.T) {
	provider := func() *config.Config {
		return &config.Config{
			Rate: config.RateLimitConfig{Enabled: true, RequestsPerSecond: 1, Burst: 1},
		}
	}
	h := RateLimit(provider)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ""
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
}
