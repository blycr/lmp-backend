package middleware

import (
	"lms/backend/internal/config"
	"net"
	"net/http"
	"strings"
)

func LANOnly(enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}
			ip := clientIP(r)
			if ip == "" || !isPrivateIP(net.ParseIP(ip)) {
				http.Error(w, "LAN access only", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func DynamicLANOnly(provider func() *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg := provider()
			if !cfg.Server.LANOnly {
				next.ServeHTTP(w, r)
				return
			}
			ip := clientIP(r)
			if ip == "" || !isPrivateIP(net.ParseIP(ip)) {
				http.Error(w, "LAN access only", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	// Prefer X-Forwarded-For if present; otherwise use RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		return strings.Trim(host[:idx], "[]")
	}
	// Trim IPv6 brackets if present
	return strings.Trim(host, "[]")
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	// IPv4 RFC1918 + CGNAT
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10",
	}
	for _, cidr := range privateBlocks {
		_, block, _ := net.ParseCIDR(cidr)
		if block != nil && block.Contains(ip) {
			return true
		}
	}
	// IPv6 ULA fc00::/7
	_, ula, _ := net.ParseCIDR("fc00::/7")
	if ula != nil && ula.Contains(ip) {
		return true
	}
	return false
}
