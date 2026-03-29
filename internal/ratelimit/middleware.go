// Package ratelimit provides Redis-backed sliding window rate limiting middleware.
//
// Rate limiting is keyed by client IP address (extracted from X-Forwarded-For
// or RemoteAddr) since engine-level middleware runs before authentication.
package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Limiter implements a Redis sliding window rate limiter.
type Limiter struct {
	redis *goredis.Client
	rpm   int
}

// New creates a rate limiter backed by Redis.
func New(redis *goredis.Client, rpm int) *Limiter {
	return &Limiter{
		redis: redis,
		rpm:   rpm,
	}
}

// Middleware returns HTTP middleware that enforces rate limits per client IP.
//
// The limiter uses a per-minute bucket in Redis: INCR the counter and EXPIRE
// after 60 seconds on first touch. If the count exceeds the configured RPM,
// the request is rejected with 429 Too Many Requests.
func (l *Limiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clientIP := extractClientIP(r)
			bucket := time.Now().UTC().Format("2006-01-02T15:04")
			key := fmt.Sprintf("ratelimit:%s:%s", clientIP, bucket)

			count, err := l.redis.Incr(ctx, key).Result()
			if err != nil {
				// Redis failure — let the request through rather than block traffic.
				next.ServeHTTP(w, r)
				return
			}

			// Set expiry on first increment so the key auto-cleans.
			if count == 1 {
				l.redis.Expire(ctx, key, 60*time.Second)
			}

			remaining := l.rpm - int(count)
			if remaining < 0 {
				remaining = 0
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(l.rpm))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if int(count) > l.rpm {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIP returns the client IP from X-Forwarded-For (first entry)
// or falls back to RemoteAddr.
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can be comma-separated; take the first (client) IP.
		if idx := strings.IndexByte(xff, ','); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	// RemoteAddr is host:port — strip the port.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
