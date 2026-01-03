package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lms/backend/internal/auth"
	"lms/backend/internal/config"
)

func TestProtectedEndpointsUnauthorized(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
		Auth:   config.AuthConfig{EnableDeviceAuth: true, SessionTimeout: time.Minute},
	}
	auth.SetDefaultManager(auth.NewManager(time.Minute))
	h := SetupRouter(func() *config.Config { return cfg })
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=a", nil)
	req1.RemoteAddr = "192.168.1.9:12345"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for search, got %d", rr1.Code)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/download?path="+urlEncode(filepath.Join(tmp, "a.txt")), nil)
	req2.RemoteAddr = "192.168.1.9:12345"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for download, got %d", rr2.Code)
	}
}

func TestInvalidDeviceCode(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Auth:   config.AuthConfig{EnableDeviceAuth: true, SessionTimeout: time.Minute},
	}
	auth.SetDefaultManager(auth.NewManager(time.Minute))
	h := SetupRouter(func() *config.Config { return cfg })
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/verify", bytes.NewReader([]byte(`{"code":"ZZZZZZ"}`)))
	req.RemoteAddr = "192.168.1.10:12345"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 invalid code, got %d", rr.Code)
	}
}

func TestAuthDisabledAllowsAccess(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
		Auth:   config.AuthConfig{EnableDeviceAuth: false, SessionTimeout: time.Minute},
	}
	h := SetupRouter(func() *config.Config { return cfg })
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
	req.RemoteAddr = "192.168.1.11:12345"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with auth disabled, got %d", rr.Code)
	}
	// 验证公共端点返回 disabled 标记
	rrC := httptest.NewRecorder()
	reqC := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/code", bytes.NewReader([]byte(`{"device_name":"t"}`)))
	reqC.RemoteAddr = "192.168.1.11:12345"
	h.ServeHTTP(rrC, reqC)
	var resp map[string]any
	_ = json.Unmarshal(rrC.Body.Bytes(), &resp)
	if _, ok := resp["disabled"]; !ok {
		t.Fatalf("expected disabled flag when auth disabled")
	}
}
