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
	redis      *goredis.Client
	rpm        int
	trustProxy bool // If false, ignore X-Forwarded-For (prevents spoofing)
}

// New creates a rate limiter backed by Redis.
// trustProxy controls whether X-Forwarded-For is used for client IP extraction.
// Set to true only when behind a trusted reverse proxy.
func New(redis *goredis.Client, rpm int, trustProxy bool) *Limiter {
	return &Limiter{
		redis:      redis,
		rpm:        rpm,
		trustProxy: trustProxy,
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
			clientIP := l.extractClientIP(r)
			bucket := time.Now().UTC().Format("2006-01-02T15:04")
			key := fmt.Sprintf("ratelimit:%s:%s", clientIP, bucket)

			// Atomic INCR + EXPIRE via Lua to prevent the race where a crash
			// between INCR and EXPIRE leaves a key with no TTL.
			count, err := incrWithExpire.Run(ctx, l.redis, []string{key}, 60).Int64()
			if err != nil {
				// Redis failure — let the request through rather than block traffic.
				next.ServeHTTP(w, r)
				return
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

// incrWithExpire atomically increments a key and sets its TTL.
// Prevents the race where a crash between INCR and EXPIRE leaves
// a key with no expiry, accumulating forever.
var incrWithExpire = goredis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return count
`)

// extractClientIP returns the client IP. If trustProxy is true, uses
// X-Forwarded-For (first entry). Otherwise uses RemoteAddr only.
func (l *Limiter) extractClientIP(r *http.Request) string {
	if l.trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.IndexByte(xff, ','); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
