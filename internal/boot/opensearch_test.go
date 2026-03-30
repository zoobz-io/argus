package boot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestEnsureIndices_CreatesWhenMissing(t *testing.T) {
	var mu sync.Mutex
	created := map[string]bool{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		index := r.URL.Path[1:] // strip leading /

		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusNotFound)
		case http.MethodPut:
			mu.Lock()
			created[index] = true
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	if err := EnsureIndices(context.Background(), srv.URL); err != nil {
		t.Fatalf("EnsureIndices returned error: %v", err)
	}

	expected := []string{"document_versions", "notifications", "audit"}
	for _, idx := range expected {
		if !created[idx] {
			t.Errorf("expected index %q to be created", idx)
		}
	}
}

func TestEnsureIndices_SkipsExisting(t *testing.T) {
	var mu sync.Mutex
	created := map[string]bool{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusOK) // index exists
		case http.MethodPut:
			index := r.URL.Path[1:]
			mu.Lock()
			created[index] = true
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	if err := EnsureIndices(context.Background(), srv.URL); err != nil {
		t.Fatalf("EnsureIndices returned error: %v", err)
	}

	if len(created) != 0 {
		t.Errorf("expected no indices to be created, but got: %v", created)
	}
}

func TestEnsureIndices_CreationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusNotFound)
		case http.MethodPut:
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"cluster_block_exception"}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	err := EnsureIndices(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error from EnsureIndices, got nil")
	}
}
