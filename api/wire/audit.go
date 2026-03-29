package wire

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zoobz-io/sum"
)

// AuditEntryResponse is the public API response for an audit log entry.
type AuditEntryResponse struct {
	Timestamp    time.Time       `json:"timestamp" description:"When the action occurred"`
	Action       string          `json:"action" description:"Action performed" example:"provider.created" discriminator:"metadata"`
	ResourceType string          `json:"resource_type" description:"Resource type" example:"provider"`
	ResourceID   string          `json:"resource_id" description:"Resource ID"`
	ID           string          `json:"id" description:"Audit entry ID"`
	ActorID      string          `json:"actor_id" description:"Actor who performed the action"`
	Metadata     json.RawMessage `json:"metadata,omitempty" description:"Action-specific metadata" discriminate:"ProviderCreatedMeta,ProviderUpdatedMeta,ProviderConnectedMeta,ProviderDeletedMeta,DocumentIngestedMeta,WatchedPathCreatedMeta,WatchedPathUpdatedMeta,TopicCreatedMeta,TopicUpdatedMeta,TagCreatedMeta,TagUpdatedMeta,TenantCreatedMeta,TenantUpdatedMeta,TenantDeletedMeta"`
}

// OnSend applies boundary masking before the response is marshaled.
func (a *AuditEntryResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[AuditEntryResponse]](ctx)
	masked, err := b.Send(ctx, *a)
	if err != nil {
		return err
	}
	*a = masked
	return nil
}

// Clone returns a copy of the response.
func (a AuditEntryResponse) Clone() AuditEntryResponse {
	c := a
	if a.Metadata != nil {
		c.Metadata = make(json.RawMessage, len(a.Metadata))
		copy(c.Metadata, a.Metadata)
	}
	return c
}

// AuditListResponse is the public API response for a paginated audit log.
type AuditListResponse struct {
	Entries []AuditEntryResponse `json:"entries" description:"List of audit entries"`
	Offset  int                  `json:"offset" description:"Number of results skipped"`
	Limit   int                  `json:"limit" description:"Page size" example:"20"`
	Total   int64                `json:"total" description:"Total number of results"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *AuditListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[AuditListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r AuditListResponse) Clone() AuditListResponse {
	c := r
	if r.Entries != nil {
		c.Entries = make([]AuditEntryResponse, len(r.Entries))
		copy(c.Entries, r.Entries)
	}
	return c
}

// Variant metadata types for discriminated union on AuditEntryResponse.Metadata.

// ProviderCreatedMeta carries metadata for provider.created actions.
type ProviderCreatedMeta struct {
	ProviderType string `json:"provider_type"`
	ProviderName string `json:"provider_name"`
}

// Clone returns a copy.
func (m ProviderCreatedMeta) Clone() ProviderCreatedMeta { return m }

// ProviderUpdatedMeta carries metadata for provider.updated actions.
type ProviderUpdatedMeta struct {
	ProviderType string `json:"provider_type"`
	ProviderName string `json:"provider_name"`
}

// Clone returns a copy.
func (m ProviderUpdatedMeta) Clone() ProviderUpdatedMeta { return m }

// ProviderConnectedMeta carries metadata for provider.connected actions.
type ProviderConnectedMeta struct {
	ProviderType string `json:"provider_type"`
}

// Clone returns a copy.
func (m ProviderConnectedMeta) Clone() ProviderConnectedMeta { return m }

// ProviderDeletedMeta carries metadata for provider.deleted actions.
type ProviderDeletedMeta struct{}

// Clone returns a copy.
func (m ProviderDeletedMeta) Clone() ProviderDeletedMeta { return m }

// DocumentIngestedMeta carries metadata for document.ingested actions.
type DocumentIngestedMeta struct {
	VersionID string `json:"version_id"`
}

// Clone returns a copy.
func (m DocumentIngestedMeta) Clone() DocumentIngestedMeta { return m }

// WatchedPathCreatedMeta carries metadata for watched_path.created actions.
type WatchedPathCreatedMeta struct {
	ProviderID string `json:"provider_id"`
	Path       string `json:"path"`
}

// Clone returns a copy.
func (m WatchedPathCreatedMeta) Clone() WatchedPathCreatedMeta { return m }

// WatchedPathUpdatedMeta carries metadata for watched_path.updated actions.
type WatchedPathUpdatedMeta struct {
	Path string `json:"path"`
}

// Clone returns a copy.
func (m WatchedPathUpdatedMeta) Clone() WatchedPathUpdatedMeta { return m }

// TopicCreatedMeta carries metadata for topic.created actions.
type TopicCreatedMeta struct {
	Name string `json:"name"`
}

// Clone returns a copy.
func (m TopicCreatedMeta) Clone() TopicCreatedMeta { return m }

// TopicUpdatedMeta carries metadata for topic.updated actions.
type TopicUpdatedMeta struct {
	Name string `json:"name"`
}

// Clone returns a copy.
func (m TopicUpdatedMeta) Clone() TopicUpdatedMeta { return m }

// TagCreatedMeta carries metadata for tag.created actions.
type TagCreatedMeta struct {
	Name string `json:"name"`
}

// Clone returns a copy.
func (m TagCreatedMeta) Clone() TagCreatedMeta { return m }

// TagUpdatedMeta carries metadata for tag.updated actions.
type TagUpdatedMeta struct {
	Name string `json:"name"`
}

// Clone returns a copy.
func (m TagUpdatedMeta) Clone() TagUpdatedMeta { return m }

// TenantCreatedMeta carries metadata for tenant.created actions.
type TenantCreatedMeta struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Clone returns a copy.
func (m TenantCreatedMeta) Clone() TenantCreatedMeta { return m }

// TenantUpdatedMeta carries metadata for tenant.updated actions.
type TenantUpdatedMeta struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Clone returns a copy.
func (m TenantUpdatedMeta) Clone() TenantUpdatedMeta { return m }

// TenantDeletedMeta carries metadata for tenant.deleted actions.
type TenantDeletedMeta struct{}

// Clone returns a copy.
func (m TenantDeletedMeta) Clone() TenantDeletedMeta { return m }
