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
// The returned reader will return io.EOF after MaxEntrySize bytes.
func safeOpen(f *zip.File) (io.ReadCloser, error) {
	if f.UncompressedSize64 > uint64(MaxEntrySize) {
		return nil, fmt.Errorf("entry %s exceeds maximum size (%d bytes)", f.Name, MaxEntrySize)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	return &limitedReadCloser{
		rc:    rc,
		limit: io.LimitReader(rc, MaxEntrySize),
	}, nil
}

// limitedReadCloser wraps a ReadCloser with a size limit.
type limitedReadCloser struct {
	rc    io.ReadCloser
	limit io.Reader
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
	return l.limit.Read(p)
}

func (l *limitedReadCloser) Close() error {
	return l.rc.Close()
}
