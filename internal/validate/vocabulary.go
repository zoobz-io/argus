// Package validate provides shared validation rules for user-supplied content
// that will be injected into LLM prompts.
package validate

import (
	"regexp"

	"github.com/zoobz-io/check"
)

// Vocabulary name: alphanumeric, spaces, hyphens. No special chars that could
// act as prompt delimiters or injection vectors.
var namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 \-]*$`)

// Vocabulary description: printable characters, basic punctuation. No control
// characters, code fences, or structural markers.
var descPattern = regexp.MustCompile(`^[a-zA-Z0-9 ,.\-;:'"!?()\/&]+$`)

const (
	// MaxNameLen is the maximum length for a vocabulary name.
	MaxNameLen = 50
	// MaxDescLen is the maximum length for a vocabulary description.
	MaxDescLen = 200
)

// Name validates a vocabulary name (topic name, tag name, etc.) for safe
// inclusion in LLM prompts.
func Name(value string, field string) *check.Validation {
	return check.Str(value, field).
		Required().
		MaxLen(MaxNameLen).
		SingleLine().
		Match(namePattern).
		V()
}

// Description validates a vocabulary description for safe inclusion in LLM prompts.
func Description(value string, field string) *check.Validation {
	return check.Str(value, field).
		Required().
		MaxLen(MaxDescLen).
		SingleLine().
		Match(descPattern).
		V()
}
