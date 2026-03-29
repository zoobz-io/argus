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
	ErrUserNotFound         = rocco.ErrNotFound.WithMessage("user not found")
	ErrSubscriptionNotFound = rocco.ErrNotFound.WithMessage("subscription not found")
	ErrNotificationNotFound = rocco.ErrNotFound.WithMessage("notification not found")

	ErrMissingRedirectURI      = rocco.ErrBadRequest.WithMessage("redirect_uri query parameter is required")
	ErrInvalidRedirectURI      = rocco.ErrBadRequest.WithMessage("redirect_uri is not in the allowed list")
	ErrProviderTypeNotSupported = rocco.ErrBadRequest.WithMessage("provider type not supported")
	ErrInvalidState            = rocco.ErrBadRequest.WithMessage("invalid or expired state token")
	ErrOAuthExchangeFailed     = rocco.ErrBadRequest.WithMessage("failed to exchange authorization code")
)
