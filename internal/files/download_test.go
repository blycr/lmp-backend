package files

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestServeFileFullAndRange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	data := []byte("abcdef")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	req := httptest.NewRequest("GET", "/download", nil)
	rr := httptest.NewRecorder()
	err := ServeFile(rr, req, path)
	if err != nil {
		t.Fatalf("serve file error: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("Accept-Ranges") != "bytes" {
		t.Fatalf("missing accept-ranges")
	}
	req2 := httptest.NewRequest("GET", "/download", nil)
	req2.Header.Set("Range", "bytes=0-2")
	rr2 := httptest.NewRecorder()
	err = ServeFile(rr2, req2, path)
	if err != nil {
		t.Fatalf("serve range error: %v", err)
	}
	if rr2.Code != http.StatusPartialContent {
		t.Fatalf("expected 206, got %d", rr2.Code)
	}
	if rr2.Body.String() != "abc" {
		t.Fatalf("unexpected body: %s", rr2.Body.String())
	}
}
