package stores

import (
	"context"
	"fmt"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// Audit wraps a search index for compliance audit entries.
type Audit struct {
	index *sum.Search[models.AuditEntry]
	qb    *lucene.Builder[models.AuditEntry]
}

// NewAudit creates a new audit search store.
func NewAudit(provider grub.SearchProvider) *Audit {
	return &Audit{
		index: sum.NewSearch[models.AuditEntry](provider, "audit"),
		qb:    lucene.New[models.AuditEntry](),
	}
}

// Index writes an audit entry to the search index.
func (s *Audit) Index(ctx context.Context, entry *models.AuditEntry) error {
	return s.index.Index(ctx, entry.ID, entry)
}

// Search queries audit entries with optional filters.
func (s *Audit) Search(ctx context.Context, params models.AuditSearchParams) (*models.OffsetResult[models.AuditEntry], error) {
	var filters []lucene.Query

	if params.TenantID != "" {
		filters = append(filters, s.qb.Term("tenant_id", params.TenantID))
	}
	if params.Action != "" {
		filters = append(filters, s.qb.Term("action", params.Action))
	}
	if params.ResourceType != "" {
		filters = append(filters, s.qb.Term("resource_type", params.ResourceType))
	}
	if params.ActorID != "" {
		filters = append(filters, s.qb.Term("actor_id", params.ActorID))
	}
	if params.From != nil {
		filters = append(filters, s.qb.Range("timestamp").Gte(params.From.Format("2006-01-02T15:04:05Z")))
	}
	if params.To != nil {
		filters = append(filters, s.qb.Range("timestamp").Lte(params.To.Format("2006-01-02T15:04:05Z")))
	}

	search := lucene.NewSearch().
		Query(s.qb.Bool().Filter(filters...)).
		Sort(lucene.SortField{Field: "timestamp", Order: "desc"}).
		Size(params.PageSize()).
		From(params.Offset)

	result, err := s.index.Execute(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("searching audit entries: %w", err)
	}

	items := make([]*models.AuditEntry, len(result.Hits))
	for i, hit := range result.Hits {
		e := hit.Content
		items[i] = &e
	}
	return &models.OffsetResult[models.AuditEntry]{
		Items:  items,
		Total:  result.Total,
		Offset: params.Offset,
	}, nil
}
