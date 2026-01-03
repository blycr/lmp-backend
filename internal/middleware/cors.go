package middleware

import (
	"lms/backend/internal/config"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func CORS(provider func() *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" {
				if allowOrigin(origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,Range")
					w.Header().Set("Access-Control-Expose-Headers", "Content-Length,Content-Range,Accept-Ranges")
				}
				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			_ = provider
			next.ServeHTTP(w, r)
		})
	}
}

func allowOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	if ip.IsPrivate() {
		return true
	}
	return false
}
