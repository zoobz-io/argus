package handlers

import "github.com/zoobz-io/rocco"

// Sentinel errors for admin handler responses.
var (
	ErrTenantNotFound      = rocco.ErrNotFound.WithMessage("tenant not found")
	ErrProviderNotFound    = rocco.ErrNotFound.WithMessage("provider not found")
	ErrWatchedPathNotFound = rocco.ErrNotFound.WithMessage("watched path not found")
	ErrDocumentNotFound    = rocco.ErrNotFound.WithMessage("document not found")
	ErrVersionNotFound     = rocco.ErrNotFound.WithMessage("document version not found")
	ErrTopicNotFound       = rocco.ErrNotFound.WithMessage("topic not found")
	ErrTagNotFound         = rocco.ErrNotFound.WithMessage("tag not found")
	ErrUserNotFound        = rocco.ErrNotFound.WithMessage("user not found")
)
