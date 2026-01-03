package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lms/backend/internal/auth"
	"lms/backend/internal/config"
)

func TestDownloadEndpointFullAndRange(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "file.txt")
	_ = os.WriteFile(fp, []byte(strings.Repeat("x", 100)), 0644)
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
	// 全量下载
	reqFull := httptest.NewRequest(http.MethodGet, "/api/v1/download?path="+urlEncode(fp), nil)
	reqFull.RemoteAddr = "192.168.1.8:12345"
	reqFull.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	rrF := httptest.NewRecorder()
	h.ServeHTTP(rrF, reqFull)
	if rrF.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrF.Code)
	}
	data, _ := io.ReadAll(rrF.Body)
	if len(data) != 100 {
		t.Fatalf("expected 100 bytes, got %d", len(data))
	}
	// Range 下载
	reqR := httptest.NewRequest(http.MethodGet, "/api/v1/download?path="+urlEncode(fp), nil)
	reqR.Header.Set("Range", "bytes=0-9")
	reqR.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	reqR.RemoteAddr = "192.168.1.8:12345"
	rrR := httptest.NewRecorder()
	h.ServeHTTP(rrR, reqR)
	if rrR.Code != http.StatusPartialContent {
		t.Fatalf("expected 206, got %d", rrR.Code)
	}
	dataR, _ := io.ReadAll(rrR.Body)
	if len(dataR) != 10 {
		t.Fatalf("expected 10 bytes, got %d", len(dataR))
	}
	// 非共享目录路径应禁止
	reqFbd := httptest.NewRequest(http.MethodGet, "/api/v1/download?path="+urlEncode(filepath.Join(os.TempDir(), "nope.txt")), nil)
	reqFbd.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	reqFbd.RemoteAddr = "192.168.1.8:12345"
	rrFbd := httptest.NewRecorder()
	h.ServeHTTP(rrFbd, reqFbd)
	if rrFbd.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden path, got %d", rrFbd.Code)
	}
	// 无效 Range
	reqInv := httptest.NewRequest(http.MethodGet, "/api/v1/download?path="+urlEncode(fp), nil)
	reqInv.Header.Set("Range", "bytes=-")
	reqInv.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	reqInv.RemoteAddr = "192.168.1.8:12345"
	rrInv := httptest.NewRecorder()
	h.ServeHTTP(rrInv, reqInv)
	if rrInv.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("expected 416, got %d", rrInv.Code)
	}
}

func urlEncode(s string) string {
	r := strings.NewReplacer(":", "%3A", "\\", "%5C")
	return r.Replace(s)
}
