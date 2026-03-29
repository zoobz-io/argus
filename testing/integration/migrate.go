//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// migrationsDir returns the absolute path to the project's migrations directory.
func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
}

// RunMigrations applies all SQL migrations to the test database.
func RunMigrations(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("opening db for migrations: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir()); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

// CreateOpenSearchIndex creates the document_versions index from the mapping file.
func CreateOpenSearchIndex(ctx context.Context, addr string) error {
	mappingPath := filepath.Join(migrationsDir(), "opensearch", "001_document_versions.json")
	mapping, err := os.ReadFile(mappingPath)
	if err != nil {
		return fmt.Errorf("reading index mapping: %w", err)
	}

	// Inject index.knn=true into the settings (required for knn_vector fields).
	var parsed map[string]any
	if err := json.Unmarshal(mapping, &parsed); err != nil {
		return fmt.Errorf("parsing mapping JSON: %w", err)
	}
	settings, _ := parsed["settings"].(map[string]any)
	if settings == nil {
		settings = make(map[string]any)
		parsed["settings"] = settings
	}
	if _, ok := settings["index.knn"]; !ok {
		settings["index.knn"] = true
	}
	mapping, err = json.Marshal(parsed)
	if err != nil {
		return fmt.Errorf("marshaling mapping: %w", err)
	}

	url := fmt.Sprintf("%s/document_versions", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(mapping))
	if err != nil {
		return fmt.Errorf("creating index request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("creating index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("creating index: status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CreateNotificationsIndex creates the notifications index from the mapping file.
func CreateNotificationsIndex(ctx context.Context, addr string) error {
	mappingPath := filepath.Join(migrationsDir(), "opensearch", "002_notifications.json")
	mapping, err := os.ReadFile(mappingPath)
	if err != nil {
		return fmt.Errorf("reading notifications index mapping: %w", err)
	}

	url := fmt.Sprintf("%s/notifications", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(mapping))
	if err != nil {
		return fmt.Errorf("creating notifications index request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("creating notifications index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("creating notifications index: status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RefreshNotificationsIndex forces OpenSearch to make all indexed notifications searchable.
func RefreshNotificationsIndex(ctx context.Context, t *testing.T, addr string) {
	t.Helper()
	url := fmt.Sprintf("%s/notifications/_refresh", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		t.Fatalf("creating notifications refresh request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("refreshing notifications index: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("notifications refresh failed: status %d", resp.StatusCode)
	}
}

// RefreshOpenSearchIndex forces OpenSearch to make all indexed data searchable.
func RefreshOpenSearchIndex(ctx context.Context, t *testing.T, addr string) {
	t.Helper()
	url := fmt.Sprintf("%s/document_versions/_refresh", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		t.Fatalf("creating refresh request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("refreshing index: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("refresh failed: status %d", resp.StatusCode)
	}
}
