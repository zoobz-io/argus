// Package events provides event definitions for the application.
package events

import "github.com/zoobz-io/capitan"

// Startup signals for server lifecycle.
// These are direct capitan signals (not sum.Event) since they're
// operational events, not domain lifecycle events for consumers.
var (
	StartupDatabaseConnected   = capitan.NewSignal("argus.startup.database.connected", "Database connection established")
	StartupStorageConnected    = capitan.NewSignal("argus.startup.storage.connected", "Object storage connection established")
	StartupRedisConnected      = capitan.NewSignal("argus.startup.redis.connected", "Redis connection established")
	StartupOpenSearchConnected = capitan.NewSignal("argus.startup.opensearch.connected", "OpenSearch connection established")
	StartupOCRConnected        = capitan.NewSignal("argus.startup.ocr.connected", "OCR service connection established")
	StartupAuthReady           = capitan.NewSignal("argus.startup.auth.ready", "OIDC authenticator initialized")
	StartupServicesReady       = capitan.NewSignal("argus.startup.services.ready", "All services registered")
	StartupOTELReady           = capitan.NewSignal("argus.startup.otel.ready", "OpenTelemetry providers initialized")
	StartupApertureReady       = capitan.NewSignal("argus.startup.aperture.ready", "Aperture observability bridge initialized")
	StartupServerListening     = capitan.NewSignal("argus.startup.server.listening", "HTTP server listening")
	StartupFailed              = capitan.NewSignal("argus.startup.failed", "Server startup failed")
)

// Startup field keys for direct emission.
var (
	StartupPortKey    = capitan.NewIntKey("port")
	StartupWorkersKey = capitan.NewIntKey("workers")
	StartupErrorKey   = capitan.NewErrorKey("error")
)
