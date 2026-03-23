// Package handlers provides HTTP endpoint handlers for the public API.
package handlers

import "github.com/zoobz-io/rocco"

// Sentinel errors for API handler responses.
var (
	ErrTenantNotFound   = rocco.ErrNotFound.WithMessage("tenant not found")
	ErrProviderNotFound = rocco.ErrNotFound.WithMessage("provider not found")
	ErrDocumentNotFound = rocco.ErrNotFound.WithMessage("document not found")
	ErrVersionNotFound  = rocco.ErrNotFound.WithMessage("document version not found")
	ErrMissingQuery     = rocco.ErrBadRequest.WithMessage("query parameter required")
)
