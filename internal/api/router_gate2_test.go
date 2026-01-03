package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lms/backend/internal/auth"
	"lms/backend/internal/config"
)

func TestGate2_EndToEnd(t *testing.T) {
	// prepare temp share dir
	root := t.TempDir()
	sub := filepath.Join(root, "subdir")
	_ = os.Mkdir(sub, 0o755)
	fileA := filepath.Join(root, "a.txt")
	fileB := filepath.Join(sub, "b.txt")
	_ = os.WriteFile(fileA, []byte("hello A"), 0o644)
	_ = os.WriteFile(fileB, []byte("hello B"), 0o644)

	// config provider
	provider := func() *config.Config {
		return &config.Config{
			Server: config.ServerConfig{Port: 8080, LANOnly: true},
			Files:  config.FilesConfig{ShareDirs: []string{root}},
			Auth:   config.AuthConfig{EnableDeviceAuth: true, SessionTimeout: time.Minute},
			Rate:   config.RateLimitConfig{Enabled: false},
		}
	}
	// auth manager
	auth.SetDefaultManager(auth.NewManager(time.Minute))

	srv := httptest.NewServer(SetupRouter(provider))
	defer srv.Close()

	// 1) get device code
	res, err := http.Post(srv.URL+"/api/v1/auth/device/code", "application/json", io.NopCloser(
		readString(`{"device_name":"tester"}`),
	))
	if err != nil {
		t.Fatalf("code post err: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", res.StatusCode)
	}
	var codeResp struct {
		Code      string `json:"code"`
		ExpiresIn int    `json:"expires_in"`
		Disabled  bool   `json:"disabled"`
	}
	_ = json.NewDecoder(res.Body).Decode(&codeResp)
	if codeResp.Disabled {
		t.Fatalf("device auth disabled")
	}
	if codeResp.Code == "" {
		t.Fatalf("empty device code")
	}

	// 2) verify device code -> token
	reqBody := readString(`{"code":"` + codeResp.Code + `"}`)
	res2, err := http.Post(srv.URL+"/api/v1/auth/device/verify", "application/json", io.NopCloser(reqBody))
	if err != nil {
		t.Fatalf("verify post err: %v", err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", res2.StatusCode)
	}
	var verifyResp struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
		Disabled  bool   `json:"disabled"`
	}
	_ = json.NewDecoder(res2.Body).Decode(&verifyResp)
	if verifyResp.Disabled {
		t.Fatalf("device auth disabled")
	}
	if verifyResp.Token == "" {
		t.Fatalf("empty token")
	}

	// helper client with bearer
	client := &http.Client{}
	authReq := func(method, url string, body io.Reader) (*http.Response, error) {
		req, _ := http.NewRequest(method, url, body)
		req.Header.Set("Authorization", "Bearer "+verifyResp.Token)
		return client.Do(req)
	}

	// 3) browse top-level
	res3, err := authReq("GET", srv.URL+"/api/v1/browse?page=1&page_size=50&sort=name&order=asc", nil)
	if err != nil {
		t.Fatalf("browse top err: %v", err)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", res3.StatusCode)
	}
	var browseTop struct {
		Items   []map[string]any `json:"items"`
		Total   int              `json:"total"`
		Current string           `json:"current"`
		Parent  string           `json:"parent"`
	}
	_ = json.NewDecoder(res3.Body).Decode(&browseTop)
	if browseTop.Total < 2 {
		t.Fatalf("expected at least 2 items, got %d", browseTop.Total)
	}

	// 4) browse subdir
	res4, err := authReq("GET", srv.URL+"/api/v1/browse?page=1&page_size=50&sort=name&order=asc&path="+urlEncode(sub), nil)
	if err != nil {
		t.Fatalf("browse sub err: %v", err)
	}
	defer res4.Body.Close()
	if res4.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", res4.StatusCode)
	}
	var browseSub struct {
		Items   []map[string]any `json:"items"`
		Total   int              `json:"total"`
		Current string           `json:"current"`
		Parent  string           `json:"parent"`
	}
	_ = json.NewDecoder(res4.Body).Decode(&browseSub)
	if browseSub.Current != sub {
		t.Fatalf("current mismatch: %s", browseSub.Current)
	}
	if browseSub.Total < 1 {
		t.Fatalf("expected items in subdir")
	}

	// 5) search
	res5, err := authReq("GET", srv.URL+"/api/v1/search?q=.txt&page=1&page_size=50&sort=name&order=asc", nil)
	if err != nil {
		t.Fatalf("search err: %v", err)
	}
	defer res5.Body.Close()
	if res5.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", res5.StatusCode)
	}
	var searchResp struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}
	_ = json.NewDecoder(res5.Body).Decode(&searchResp)
	if searchResp.Total < 2 {
		t.Fatalf("expected at least 2 search hits, got %d", searchResp.Total)
	}

	// 6) download
	res6, err := authReq("GET", srv.URL+"/api/v1/download?path="+urlEncode(fileA), nil)
	if err != nil {
		t.Fatalf("download err: %v", err)
	}
	defer res6.Body.Close()
	if res6.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", res6.StatusCode)
	}
	data, _ := io.ReadAll(res6.Body)
	if string(data) != "hello A" {
		t.Fatalf("download content mismatch: %s", string(data))
	}

	// 7) ui smoke
	res7, err := http.Get(srv.URL + "/ui/")
	if err != nil {
		t.Fatalf("ui get err: %v", err)
	}
	defer res7.Body.Close()
	if res7.StatusCode != http.StatusOK {
		t.Fatalf("ui unexpected status %d", res7.StatusCode)
	}
}

func readString(s string) *stringReader {
	return &stringReader{S: s, I: 0}
}

type stringReader struct {
	S string
	I int
}

func (r *stringReader) Read(p []byte) (int, error) {
	if r.I >= len(r.S) {
		return 0, io.EOF
	}
	n := copy(p, []byte(r.S[r.I:]))
	r.I += n
	return n, nil
}
