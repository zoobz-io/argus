package extract

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

// Archive safety limits.
const (
	// MaxArchiveFiles is the maximum number of entries allowed in an archive.
	MaxArchiveFiles = 10_000

	// MaxEntrySize is the maximum decompressed size per archive entry (256 MB).
	MaxEntrySize int64 = 256 << 20

	// MaxArchiveInputSize is the maximum raw input size for archive-based formats (512 MB).
	MaxArchiveInputSize int64 = 512 << 20
)

// safeZIPReader opens a zip archive with entry count validation.
// Returns an error if the archive exceeds MaxArchiveFiles.
func safeZIPReader(data []byte) (*zip.Reader, error) {
	if int64(len(data)) > MaxArchiveInputSize {
		return nil, fmt.Errorf("archive exceeds maximum input size (%d bytes)", MaxArchiveInputSize)
	}
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	if len(r.File) > MaxArchiveFiles {
		return nil, fmt.Errorf("archive contains %d entries (max %d)", len(r.File), MaxArchiveFiles)
	}
	return r, nil
}

// safeOpen opens a zip entry with a decompressed size cap.
// Returns an error if decompressed content exceeds MaxEntrySize.
//
// NOTE: f.UncompressedSize64 is attacker-controlled (set by the archive
// creator). The pre-check catches honest archives early. The read-time
// limit is the real security boundary.
func safeOpen(f *zip.File) (io.ReadCloser, error) {
	if f.UncompressedSize64 > uint64(MaxEntrySize) {
		return nil, fmt.Errorf("entry %s declares size %d bytes (max %d)", f.Name, f.UncompressedSize64, MaxEntrySize)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	return &guardedReadCloser{
		rc:    rc,
		limit: io.LimitReader(rc, MaxEntrySize),
		name:  f.Name,
	}, nil
}

// guardedReadCloser wraps a ReadCloser with a size limit that errors
// on truncation instead of silently returning partial content.
type guardedReadCloser struct {
	rc    io.ReadCloser
	limit io.Reader
	name  string
	total int64
}

func (g *guardedReadCloser) Read(p []byte) (int, error) {
	n, err := g.limit.Read(p)
	g.total += int64(n)
	if err == io.EOF && g.total >= MaxEntrySize {
		// LimitReader returned EOF at the cap — probe for more data.
		var probe [1]byte
		if pn, _ := g.rc.Read(probe[:]); pn > 0 {
			return n, fmt.Errorf("entry %s exceeds maximum decompressed size (%d bytes)", g.name, MaxEntrySize)
		}
	}
	return n, err
}

func (g *guardedReadCloser) Close() error {
	return g.rc.Close()
}
