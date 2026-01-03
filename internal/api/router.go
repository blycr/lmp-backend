package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fmt"
	"lms/backend/internal/auth"
	"lms/backend/internal/config"
	"lms/backend/internal/files"
	"lms/backend/internal/middleware"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

type healthResp struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

func SetupRouter(provider func() *config.Config) http.Handler {
	r := mux.NewRouter()

	r.Use(middleware.CORS(provider))
	r.Use(middleware.DynamicLANOnly(provider))
	r.Use(middleware.RateLimit(provider))
	r.Use(securityHeaders)

	api := r.PathPrefix("/api/v1").Subrouter()
	api.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	public := api.PathPrefix("").Subrouter()
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthRequired(provider))

	public.HandleFunc("/auth/device/code", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cfg := provider()
		if !cfg.Auth.EnableDeviceAuth {
			writeJSON(w, http.StatusOK, map[string]any{"disabled": true})
			return
		}
		var body struct {
			DeviceName string `json:"device_name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		m := auth.DefaultManager()
		if m == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "auth not initialized"})
			return
		}
		code := m.GenerateDeviceCode(body.DeviceName)
		writeJSON(w, http.StatusOK, map[string]any{"code": code.Code, "expires_in": int(time.Until(code.ExpiresAt).Seconds())})
	}).Methods(http.MethodPost)

	public.HandleFunc("/auth/device/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cfg := provider()
		if !cfg.Auth.EnableDeviceAuth {
			writeJSON(w, http.StatusOK, map[string]any{"disabled": true})
			return
		}
		var body struct {
			Code string `json:"code"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		m := auth.DefaultManager()
		if m == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "auth not initialized"})
			return
		}
		if sess, ok := m.VerifyDeviceCode(body.Code); ok {
			writeJSON(w, http.StatusOK, map[string]any{
				"token":      sess.Token,
				"expires_in": int(time.Until(sess.ExpiresAt).Seconds()),
			})
			return
		}
		http.Error(w, "invalid code", http.StatusUnauthorized)
	}).Methods(http.MethodPost)

	protected.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		cfg := provider()
		q := r.URL.Query()
		filter := q.Get("filter")
		sortBy := q.Get("sort")
		order := strings.ToLower(q.Get("order")) == "desc"
		opt := files.ListOptions{Filter: filter, SortBy: sortBy, OrderDesc: order}
		items, _ := files.ListTopLevel(cfg.Files.ShareDirs, opt)
		page, size := parsePage(q.Get("page")), parsePageSize(q.Get("page_size"))
		if size <= 0 {
			size = 50
		}
		total := len(items)
		start := (page - 1) * size
		if start < 0 {
			start = 0
		}
		if start > total {
			start = total
		}
		end := start + size
		if end > total {
			end = total
		}
		slice := items[start:end]
		if slice == nil {
			slice = []files.FileItem{}
		}
		resp := map[string]any{
			"items": slice,
			"page":  page,
			"total": total,
		}
		writeJSON(w, http.StatusOK, resp)
	}).Methods(http.MethodGet)

	protected.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		cfg := provider()
		ix := files.GetDefaultIndexer()
		if ix == nil {
			tmp := files.NewIndexer()
			tmp.SetRoots(cfg.Files.ShareDirs)
			_ = tmp.Rebuild()
			ix = tmp
		}
		q := r.URL.Query()
		query := q.Get("q")
		page, size := parsePage(q.Get("page")), parsePageSize(q.Get("page_size"))
		var opt files.SearchOptions
		tp := strings.ToLower(strings.TrimSpace(q.Get("type")))
		if tp == "file" || tp == "dir" {
			opt.Type = tp
		}
		sortBy := strings.ToLower(strings.TrimSpace(q.Get("sort")))
		if sortBy == "size" || sortBy == "time" || sortBy == "type" || sortBy == "name" {
			opt.SortBy = sortBy
		} else {
			opt.SortBy = "name"
		}
		opt.OrderDesc = strings.ToLower(strings.TrimSpace(q.Get("order"))) == "desc"
		if ms := q.Get("min_size"); ms != "" {
			if v, err := strconv.ParseInt(ms, 10, 64); err == nil && v >= 0 {
				opt.MinSize = v
			}
		}
		if ms := q.Get("max_size"); ms != "" {
			if v, err := strconv.ParseInt(ms, 10, 64); err == nil && v >= 0 {
				opt.MaxSize = v
			}
		}
		if s := q.Get("modified_after"); s != "" {
			if t, ok := parseTime(s); ok {
				opt.ModifiedAfter = t
			}
		}
		if s := q.Get("modified_before"); s != "" {
			if t, ok := parseTime(s); ok {
				opt.ModifiedBefore = t
			}
		}
		items, total := ix.SearchAdvanced(query, opt, page, size)
		if items == nil {
			items = []files.FileItem{}
		}
		resp := map[string]any{
			"items": items,
			"page":  page,
			"total": total,
		}
		writeJSON(w, http.StatusOK, resp)
	}).Methods(http.MethodGet)

	protected.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		cfg := provider()
		q := r.URL.Query()
		raw := q.Get("path")
		if raw == "" {
			http.Error(w, "path required", http.StatusBadRequest)
			return
		}
		p, err := url.QueryUnescape(raw)
		if err != nil {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		if safe, ok := files.ResolveSafePath(cfg.Files.ShareDirs, p); ok {
			if err := files.ServeFile(w, r, safe); err != nil {
				http.Error(w, fmt.Sprintf("download error: %v", err), http.StatusInternalServerError)
			}
			return
		}
		http.Error(w, "forbidden path", http.StatusForbidden)
	}).Methods(http.MethodGet)

	protected.HandleFunc("/browse", func(w http.ResponseWriter, r *http.Request) {
		cfg := provider()
		q := r.URL.Query()
		filter := q.Get("filter")
		sortBy := q.Get("sort")
		order := strings.ToLower(q.Get("order")) == "desc"
		page, size := parsePage(q.Get("page")), parsePageSize(q.Get("page_size"))
		opt := files.ListOptions{Filter: filter, SortBy: sortBy, OrderDesc: order}
		raw := q.Get("path")
		if raw == "" {
			items, _ := files.ListTopLevel(cfg.Files.ShareDirs, opt)
			total := len(items)
			start := (page - 1) * size
			if start < 0 {
				start = 0
			}
			if start > total {
				start = total
			}
			end := start + size
			if end > total {
				end = total
			}
			slice := items[start:end]
			if slice == nil {
				slice = []files.FileItem{}
			}
			resp := map[string]any{
				"items":   slice,
				"page":    page,
				"total":   total,
				"current": "",
				"parent":  "",
			}
			writeJSON(w, http.StatusOK, resp)
			return
		}
		p, err := url.QueryUnescape(raw)
		if err != nil {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		if safe, ok := files.ResolveSafePath(cfg.Files.ShareDirs, p); ok {
			items, _ := files.ListDir(safe, opt)
			total := len(items)
			start := (page - 1) * size
			if start < 0 {
				start = 0
			}
			if start > total {
				start = total
			}
			end := start + size
			if end > total {
				end = total
			}
			slice := items[start:end]
			if slice == nil {
				slice = []files.FileItem{}
			}
			parent := filepath.Dir(safe)
			// If parent escapes roots, clear it
			if _, ok := files.ResolveSafePath(cfg.Files.ShareDirs, parent); !ok {
				parent = ""
			}
			resp := map[string]any{
				"items":   slice,
				"page":    page,
				"total":   total,
				"current": safe,
				"parent":  parent,
			}
			writeJSON(w, http.StatusOK, resp)
			return
		}
		http.Error(w, "forbidden path", http.StatusForbidden)
	}).Methods(http.MethodGet)

	protected.HandleFunc("/meta", func(w http.ResponseWriter, r *http.Request) {
		cfg := provider()
		q := r.URL.Query()
		raw := q.Get("path")
		if raw == "" {
			http.Error(w, "path required", http.StatusBadRequest)
			return
		}
		p, err := url.QueryUnescape(raw)
		if err != nil {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		if safe, ok := files.ResolveSafePath(cfg.Files.ShareDirs, p); ok {
			it, err := files.StatItem(safe)
			if err != nil {
				http.Error(w, "stat error", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, it)
			return
		}
		http.Error(w, "forbidden path", http.StatusForbidden)
	}).Methods(http.MethodGet)

	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResp{Status: "ok", Timestamp: time.Now().Unix()})
	})

	if dir := findFrontendDir(); dir != "" {
		r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir(dir))))
	}

	return r
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}

func parsePage(s string) int {
	if s == "" {
		return 1
	}
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return 1
}

func parsePageSize(s string) int {
	if s == "" {
		return 50
	}
	if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 500 {
		return n
	}
	return 50
}

func parseTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(n, 0), true
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func findFrontendDir() string {
	candidates := []string{
		filepath.Join("..", "..", "..", "dist", "ui"),
		filepath.Join("..", "..", "dist", "ui"),
		filepath.Join("..", "dist", "ui"),
		filepath.Join("dist", "ui"),
		filepath.Join("..", "..", "..", "frontend", "public"),
		filepath.Join("..", "..", "frontend", "public"),
		filepath.Join("..", "frontend", "public"),
		filepath.Join("frontend", "public"),
		filepath.Join(".", "frontend", "public"),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return ""
}
