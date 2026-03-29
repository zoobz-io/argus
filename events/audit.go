package events

import (
	"github.com/zoobz-io/capitan"

	"github.com/zoobz-io/argus/models"
)

// AuditSignal is emitted when an auditable action occurs.
// Herald publishes this to the audit stream for the notifier sidecar to consume.
var AuditSignal = capitan.NewSignal("argus.audit", "Audit event")

// AuditKey carries the AuditEntry payload on the signal.
var AuditKey = capitan.NewKey[models.AuditEntry]("audit", "models.AuditEntry")

// Audit sidecar operational signals.
var (
	AuditIndexed    = capitan.NewSignal("argus.audit.indexed", "Audit entry indexed in search")
	AuditIndexError = capitan.NewSignal("argus.audit.index.error", "Failed to index audit entry")
)

// Audit field keys for signal emission.
var (
	AuditActionKey = capitan.NewStringKey("audit_action")
	AuditErrorKey  = capitan.NewErrorKey("error")
)
