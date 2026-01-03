package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lms/backend/internal/config"
)

func TestUIServesIndex(t *testing.T) {
	provider := func() *config.Config {
		return &config.Config{
			Server: config.ServerConfig{LANOnly: false},
			Auth:   config.AuthConfig{EnableDeviceAuth: true},
			Files:  config.FilesConfig{ShareDirs: []string{}},
		}
	}
	h := SetupRouter(provider)
	s := httptest.NewServer(h)
	defer s.Close()
	resp, err := http.Get(s.URL + "/ui/")
	if err != nil {
		t.Fatalf("http get ui: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("content-type=%s", ct)
	}
	b, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(b), "<title>LMP 前端</title>") {
		t.Fatalf("unexpected body")
	}
}
