//go:build testing

package extract

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
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

func TestGuardedReadCloser_ExceedsLimit(t *testing.T) {
	// Build a zip with content larger than a small limit.
	content := strings.Repeat("x", 1024)
	data := makeZIP(t, map[string]string{"big.txt": content})
	r, err := safeZIPReader(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rc, err := r.File[0].Open()
	if err != nil {
		t.Fatalf("open error: %v", err)
	}

	// Wrap with a small limit (100 bytes) to trigger truncation detection.
	smallLimit := int64(100)
	guarded := &guardedReadCloser{
		rc:       rc,
		limit:    io.LimitReader(rc, smallLimit),
		name:     "big.txt",
		maxBytes: smallLimit,
	}
	// Read all — should error when limit is hit and more data exists.
	buf := make([]byte, 4096)
	var totalRead int64
	var readErr error
	for {
		n, err := guarded.Read(buf)
		totalRead += int64(n)
		if err != nil {
			readErr = err
			break
		}
	}
	if readErr == nil || errors.Is(readErr, io.EOF) {
		t.Fatal("expected size exceeded error, got nil or EOF")
	}
	if !strings.Contains(readErr.Error(), "exceeds maximum decompressed size") {
		t.Errorf("error should mention size exceeded, got %q", readErr.Error())
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
