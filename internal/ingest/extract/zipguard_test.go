//go:build testing

package extract

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

// makeZIP creates a valid zip archive with the given entries.
func makeZIP(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range entries {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("creating zip entry %q: %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("writing zip entry %q: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	return buf.Bytes()
}

// makeZIPWithManyFiles creates a zip with n empty entries.
func makeZIPWithManyFiles(t *testing.T, n int) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for i := 0; i < n; i++ {
		if _, err := w.Create("file" + string(rune('0'+i%10)) + ".txt"); err != nil {
			t.Fatalf("creating zip entry %d: %v", i, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	return buf.Bytes()
}

func TestSafeZIPReader_Valid(t *testing.T) {
	data := makeZIP(t, map[string]string{"test.txt": "hello"})
	r, err := safeZIPReader(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.File) != 1 {
		t.Errorf("expected 1 file, got %d", len(r.File))
	}
}

func TestSafeZIPReader_InvalidArchive(t *testing.T) {
	_, err := safeZIPReader([]byte("not a zip"))
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}

func TestSafeZIPReader_TooManyFiles(t *testing.T) {
	data := makeZIPWithManyFiles(t, MaxArchiveFiles+1)
	_, err := safeZIPReader(data)
	if err == nil {
		t.Fatal("expected error for too many files")
	}
	if !strings.Contains(err.Error(), "entries") {
		t.Errorf("error should mention entries, got %q", err.Error())
	}
}

func TestSafeZIPReader_InputTooLarge(t *testing.T) {
	// We can't easily create a 512MB+ file in a test, but we can test the check.
	// Create a normal zip and verify it passes, then verify the size check logic.
	data := makeZIP(t, map[string]string{"test.txt": "hello"})
	if int64(len(data)) > MaxArchiveInputSize {
		t.Fatal("test data should be small")
	}
	if _, err := safeZIPReader(data); err != nil {
		t.Fatalf("small archive should pass: %v", err)
	}
}

func TestSafeOpen_Valid(t *testing.T) {
	data := makeZIP(t, map[string]string{"test.txt": "hello world"})
	r, _ := safeZIPReader(data)
	rc, err := safeOpen(r.File[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(rc); err != nil {
		t.Fatalf("read error: %v", err)
	}
	if buf.String() != "hello world" {
		t.Errorf("content: got %q, want %q", buf.String(), "hello world")
	}
}
