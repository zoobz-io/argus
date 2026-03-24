package extract

import (
	"context"
	"strings"
)

// Plaintext extracts content from text-based formats by reading bytes directly.
// Suitable for TXT, Markdown, CSV, and other plain text MIME types.
func Plaintext(_ context.Context, data []byte) (string, error) {
	return strings.TrimSpace(string(data)), nil
}
