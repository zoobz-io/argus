//go:build testing

package argustest

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// ZipFile represents a single file entry in a test zip archive.
type ZipFile struct {
	Name    string
	Content string
}

// testdataDir returns the absolute path to the testing/testdata directory.
func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

// ReadTestdata reads a file from the testing/testdata directory.
func ReadTestdata(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join(testdataDir(), name)
	p = filepath.Clean(p)
	data, err := os.ReadFile(p) //nolint:gosec // test-only, path is from testdata dir
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	return data
}

// BuildZip creates a minimal zip archive in memory from the given files.
func BuildZip(t *testing.T, files ...ZipFile) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, f := range files {
		fw, err := w.Create(f.Name)
		if err != nil {
			t.Fatalf("creating zip entry %s: %v", f.Name, err)
		}
		if _, err := fw.Write([]byte(f.Content)); err != nil {
			t.Fatalf("writing zip entry %s: %v", f.Name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	return buf.Bytes()
}
