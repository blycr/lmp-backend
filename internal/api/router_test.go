package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"
	"testing"

	"lms/backend/internal/config"
	"lms/backend/internal/auth"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
	}
	h := SetupRouter(func() *config.Config { return cfg })
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestFilesPlaceholder(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
	}
	h := SetupRouter(func() *config.Config { return cfg })
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRateLimitBlocksExcess(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: true, RequestsPerSecond: 1.0, Burst: 1},
	}
	h := SetupRouter(func() *config.Config { return cfg })
	req1 := httptest.NewRequest(http.MethodGet, "/health", nil)
	req1.RemoteAddr = "192.168.1.2:12345"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr1.Code)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr2.Code)
	}
}

func TestDynamicLANOnlyToggle(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: true},
		Rate:   config.RateLimitConfig{Enabled: false},
	}
	h := SetupRouter(func() *config.Config { return cfg })
	reqBlocked := httptest.NewRequest(http.MethodGet, "/health", nil)
	reqBlocked.RemoteAddr = "1.2.3.4:5678"
	rrB := httptest.NewRecorder()
	h.ServeHTTP(rrB, reqBlocked)
	if rrB.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rrB.Code)
	}
	cfg.Server.LANOnly = false
	reqAllowed := httptest.NewRequest(http.MethodGet, "/health", nil)
	reqAllowed.RemoteAddr = "1.2.3.4:5678"
	rrA := httptest.NewRecorder()
	h.ServeHTTP(rrA, reqAllowed)
	if rrA.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrA.Code)
	}
}

func TestListFilesEndpoint(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("bb"), 0644)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
		Auth:   config.AuthConfig{EnableDeviceAuth: true, SessionTimeout: time.Minute},
	}
	auth.SetDefaultManager(auth.NewManager(time.Minute))
	h := SetupRouter(func() *config.Config { return cfg })
	// 未授权应 401
	reqUnauthorized := httptest.NewRequest(http.MethodGet, "/api/v1/files?sort=name&order=asc&page=1&page_size=1", nil)
	reqUnauthorized.RemoteAddr = "192.168.1.3:12345"
	rrU := httptest.NewRecorder()
	h.ServeHTTP(rrU, reqUnauthorized)
	if rrU.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rrU.Code)
	}
	// 设备认证获取 token
	body := []byte(`{"device_name":"test"}`)
	reqCode := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/code", bytes.NewReader(body))
	reqCode.RemoteAddr = "192.168.1.3:12345"
	rrC := httptest.NewRecorder()
	h.ServeHTTP(rrC, reqCode)
	if rrC.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrC.Code)
	}
	var cResp struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(rrC.Body.Bytes(), &cResp)
	reqVerify := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/verify", bytes.NewReader([]byte(`{"code":"`+cResp.Code+`"}`)))
	reqVerify.RemoteAddr = "192.168.1.3:12345"
	rrV := httptest.NewRecorder()
	h.ServeHTTP(rrV, reqVerify)
	if rrV.Code != http.StatusOK {
		t.Fatalf("expected 200 verify, got %d", rrV.Code)
	}
	var vResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(rrV.Body.Bytes(), &vResp)
	// 使用 token 访问
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files?sort=name&order=asc&page=1&page_size=1", nil)
	req.RemoteAddr = "192.168.1.3:12345"
	req.Header.Set("Authorization", "Bearer "+vResp.Token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSearchEndpoint(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "alpha.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "beta.txt"), []byte("b"), 0644)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
	}
	h := SetupRouter(func() *config.Config { return cfg })
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=a&page=1&page_size=10", nil)
	req.RemoteAddr = "192.168.1.4:12345"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSearchFiltersTypeAndSize(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "small.txt"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "large.txt"), []byte(strings.Repeat("y", 1024)), 0644)
	_ = os.Mkdir(filepath.Join(tmp, "folder"), 0755)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
	}
	h := SetupRouter(func() *config.Config { return cfg })
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/search?type=dir", nil)
	req1.RemoteAddr = "192.168.1.5:12345"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr1.Code)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/search?type=file&min_size=512", nil)
	req2.RemoteAddr = "192.168.1.5:12345"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
}

func TestSearchFiltersTime(t *testing.T) {
	tmp := t.TempDir()
	old := filepath.Join(tmp, "old.txt")
	newf := filepath.Join(tmp, "new.txt")
	_ = os.WriteFile(old, []byte("o"), 0644)
	time.Sleep(1100 * time.Millisecond)
	_ = os.WriteFile(newf, []byte("n"), 0644)
	after := time.Now().Add(-500 * time.Millisecond).Unix()
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
	}
	h := SetupRouter(func() *config.Config { return cfg })
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?modified_after="+strconv.FormatInt(after, 10), nil)
	req.RemoteAddr = "192.168.1.6:12345"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSearchSortNameDesc(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "c.txt"), []byte("c"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("b"), 0644)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
	}
	h := SetupRouter(func() *config.Config { return cfg })
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?sort=name&order=desc&page_size=2", nil)
	req.RemoteAddr = "192.168.1.7:12345"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp.Items) < 2 {
		t.Fatalf("expected at least 2 items")
	}
	if !(resp.Items[0].Name >= resp.Items[1].Name) {
		t.Fatalf("expected desc name order")
	}
}
