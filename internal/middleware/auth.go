package middleware

import (
	"lms/backend/internal/auth"
	"lms/backend/internal/config"
	"net/http"
	"strings"
)

func AuthRequired(provider func() *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg := provider()
			if !cfg.Auth.EnableDeviceAuth {
				next.ServeHTTP(w, r)
				return
			}
			m := auth.DefaultManager()
			if m == nil {
				http.Error(w, "auth not initialized", http.StatusUnauthorized)
				return
			}
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			token := strings.TrimSpace(h[len("Bearer "):])
			if _, ok := m.ValidateToken(token); !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
