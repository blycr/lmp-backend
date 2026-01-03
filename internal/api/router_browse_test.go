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

func TestBrowseForbiddenAndTopLevel(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "x.txt"), []byte("x"), 0644)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
		Auth:   config.AuthConfig{EnableDeviceAuth: true, SessionTimeout: time.Minute},
	}
	auth.SetDefaultManager(auth.NewManager(time.Minute))
	h := SetupRouter(func() *config.Config { return cfg })
	// 获取 token
	rrC := httptest.NewRecorder()
	reqC := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/code", bytes.NewReader([]byte(`{"device_name":"t"}`)))
	h.ServeHTTP(rrC, reqC)
	var codeResp struct{ Code string }
	_ = json.Unmarshal(rrC.Body.Bytes(), &codeResp)
	rrV := httptest.NewRecorder()
	reqV := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/verify", bytes.NewReader([]byte(`{"code":"`+codeResp.Code+`"}`)))
	h.ServeHTTP(rrV, reqV)
	var tokenResp struct{ Token string }
	_ = json.Unmarshal(rrV.Body.Bytes(), &tokenResp)
	// 顶层
	reqTop := httptest.NewRequest(http.MethodGet, "/api/v1/browse?page=1&page_size=50&sort=name&order=asc", nil)
	reqTop.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	rrTop := httptest.NewRecorder()
	h.ServeHTTP(rrTop, reqTop)
	if rrTop.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrTop.Code)
	}
	// 非共享目录
	reqF := httptest.NewRequest(http.MethodGet, "/api/v1/browse?page=1&page_size=50&sort=name&order=asc&path="+urlEncode(filepath.Join(os.TempDir(), "nope")), nil)
	reqF.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	rrF := httptest.NewRecorder()
	h.ServeHTTP(rrF, reqF)
	if rrF.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rrF.Code)
	}
}
