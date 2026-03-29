package models

import "time"

// AuditEntry is the OpenSearch document type for compliance audit logging.
type AuditEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	ActorID      string    `json:"actor_id"`
	Metadata     string    `json:"metadata"`
}

// Clone returns a deep copy of the audit entry.
func (a AuditEntry) Clone() AuditEntry {
	return a
}

// AuditSearchParams holds filter parameters for audit log queries.
type AuditSearchParams struct {
	From         *time.Time
	To           *time.Time
	Action       string
	ResourceType string
	ActorID      string
	TenantID     string
	Offset       int
	Limit        int
}

// PageSize returns the effective page size, clamped to bounds.
func (p AuditSearchParams) PageSize() int {
	if p.Limit <= 0 {
		return DefaultPageSize
	}
	if p.Limit > MaxPageSize {
		return MaxPageSize
	}
	return p.Limit
}
