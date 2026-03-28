// Package handlers provides HTTP endpoint handlers for the public API.
package handlers

import "github.com/zoobz-io/rocco"

// Sentinel errors for API handler responses.
var (
	ErrTenantNotFound      = rocco.ErrNotFound.WithMessage("tenant not found")
	ErrProviderNotFound    = rocco.ErrNotFound.WithMessage("provider not found")
	ErrWatchedPathNotFound = rocco.ErrNotFound.WithMessage("watched path not found")
	ErrDocumentNotFound    = rocco.ErrNotFound.WithMessage("document not found")
	ErrVersionNotFound = rocco.ErrNotFound.WithMessage("document version not found")
	ErrTopicNotFound   = rocco.ErrNotFound.WithMessage("topic not found")
	ErrTagNotFound     = rocco.ErrNotFound.WithMessage("tag not found")
	ErrJobNotFound     = rocco.ErrNotFound.WithMessage("job not found")
)
