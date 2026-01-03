package api

import (
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

func TestFilesEndpointBasic(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "x.txt"), []byte("x"), 0644)
	_ = os.Mkdir(filepath.Join(tmp, "d"), 0755)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, LANOnly: false},
		Rate:   config.RateLimitConfig{Enabled: false},
		Files:  config.FilesConfig{ShareDirs: []string{tmp}},
		Auth:   config.AuthConfig{EnableDeviceAuth: true, SessionTimeout: time.Minute},
	}
	auth.SetDefaultManager(auth.NewManager(time.Minute))
	h := SetupRouter(func() *config.Config { return cfg })
	// get token
	rrC := httptest.NewRecorder()
	reqC := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/code", nil)
	rrV := httptest.NewRecorder()
	reqV := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/verify", rrC.Body)
	h.ServeHTTP(rrC, reqC)
	h.ServeHTTP(rrV, reqV)
	var token struct{ Token string }
	_ = json.Unmarshal(rrV.Body.Bytes(), &token)
	// files
	reqF := httptest.NewRequest(http.MethodGet, "/api/v1/files?page=1&page_size=50&sort=name&order=asc", nil)
	reqF.Header.Set("Authorization", "Bearer "+token.Token)
	rrF := httptest.NewRecorder()
	h.ServeHTTP(rrF, reqF)
	if rrF.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrF.Code)
	}
}
