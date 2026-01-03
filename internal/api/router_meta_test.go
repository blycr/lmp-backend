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

func TestMetaEndpoint(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "meta.txt")
	_ = os.WriteFile(fp, []byte("meta"), 0644)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
		Auth:   config.AuthConfig{EnableDeviceAuth: true, SessionTimeout: time.Minute},
	}
	auth.SetDefaultManager(auth.NewManager(time.Minute))
	h := SetupRouter(func() *config.Config { return cfg })
	// 获取 code
	rrC := httptest.NewRecorder()
	reqC := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/code", bytes.NewReader([]byte(`{"device_name":"t"}`)))
	reqC.RemoteAddr = "192.168.1.8:12345"
	h.ServeHTTP(rrC, reqC)
	var codeResp struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(rrC.Body.Bytes(), &codeResp)
	// 验证获取 token
	rrV := httptest.NewRecorder()
	reqV := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/verify", bytes.NewReader([]byte(`{"code":"`+codeResp.Code+`"}`)))
	reqV.RemoteAddr = "192.168.1.8:12345"
	h.ServeHTTP(rrV, reqV)
	var tokenResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(rrV.Body.Bytes(), &tokenResp)
	// meta 正常
	reqMeta := httptest.NewRequest(http.MethodGet, "/api/v1/meta?path="+urlEncode(fp), nil)
	reqMeta.RemoteAddr = "192.168.1.8:12345"
	reqMeta.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	rrM := httptest.NewRecorder()
	h.ServeHTTP(rrM, reqMeta)
	if rrM.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrM.Code)
	}
	// meta 非共享目录禁止
	reqF := httptest.NewRequest(http.MethodGet, "/api/v1/meta?path="+urlEncode(filepath.Join(os.TempDir(), "x.txt")), nil)
	reqF.RemoteAddr = "192.168.1.8:12345"
	reqF.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	rrF := httptest.NewRecorder()
	h.ServeHTTP(rrF, reqF)
	if rrF.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rrF.Code)
	}
}
