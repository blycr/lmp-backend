package middleware

import (
	"lms/backend/internal/config"
	"net/http"
	"sync"
	"time"
)

type tokenBucket struct {
	tokens float64
	last   time.Time
}

type limiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

func newLimiter() *limiter {
	return &limiter{buckets: make(map[string]*tokenBucket)}
}

func (l *limiter) allow(key string, rps float64, burst int) bool {
	now := time.Now()
	l.mu.Lock()
	b, ok := l.buckets[key]
	if !ok {
		b = &tokenBucket{tokens: float64(burst), last: now}
		l.buckets[key] = b
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * rps
	if b.tokens > float64(burst) {
		b.tokens = float64(burst)
	}
	b.last = now
	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		l.mu.Unlock()
		return true
	}
	l.mu.Unlock()
	return false
}

func RateLimit(provider func() *config.Config) func(http.Handler) http.Handler {
	l := newLimiter()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg := provider()
			if !cfg.Rate.Enabled || cfg.Rate.RequestsPerSecond <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			ip := clientIP(r)
			if ip == "" {
				http.Error(w, "rate limit", http.StatusTooManyRequests)
				return
			}
			if l.allow(ip, cfg.Rate.RequestsPerSecond, cfg.Rate.Burst) {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit", http.StatusTooManyRequests)
		})
	}
}

