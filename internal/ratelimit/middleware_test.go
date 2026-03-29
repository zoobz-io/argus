//go:build testing

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func setupTestLimiter(t *testing.T, rpm int) (*Limiter, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return New(client, rpm), mr
}

func TestMiddleware_UnderLimit_RequestPasses(t *testing.T) {
	limiter, _ := setupTestLimiter(t, 10)
	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	remaining := rec.Header().Get("X-RateLimit-Remaining")
	if remaining != "9" {
		t.Errorf("expected X-RateLimit-Remaining=9, got %s", remaining)
	}
	limit := rec.Header().Get("X-RateLimit-Limit")
	if limit != "10" {
		t.Errorf("expected X-RateLimit-Limit=10, got %s", limit)
	}
}

func TestMiddleware_OverLimit_Returns429(t *testing.T) {
	limiter, _ := setupTestLimiter(t, 3)
	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Send 3 requests (at limit) — all should pass.
	for i := range 3 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// 4th request exceeds limit.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
	if retry := rec.Header().Get("Retry-After"); retry != "60" {
		t.Errorf("expected Retry-After=60, got %s", retry)
	}
	if remaining := rec.Header().Get("X-RateLimit-Remaining"); remaining != "0" {
		t.Errorf("expected X-RateLimit-Remaining=0, got %s", remaining)
	}
}

func TestMiddleware_HeadersSetCorrectly(t *testing.T) {
	limiter, _ := setupTestLimiter(t, 5)
	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := range 5 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "172.16.0.1:8080"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		expected := 5 - (i + 1)
		remaining := rec.Header().Get("X-RateLimit-Remaining")
		if remaining != strconv.Itoa(expected) {
			t.Errorf("request %d: expected remaining=%d, got %s", i+1, expected, remaining)
		}
	}
}

func TestMiddleware_XForwardedFor_UsedAsKey(t *testing.T) {
	limiter, _ := setupTestLimiter(t, 1)
	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request from forwarded IP — passes.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18")
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Second request from same forwarded IP — blocked.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.50")
	req2.RemoteAddr = "127.0.0.1:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for same forwarded IP, got %d", rec2.Code)
	}

	// Third request from different forwarded IP — passes.
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set("X-Forwarded-For", "198.51.100.10")
	req3.RemoteAddr = "127.0.0.1:1234"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Errorf("expected 200 for different forwarded IP, got %d", rec3.Code)
	}
}
