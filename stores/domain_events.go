package stores

import (
	"context"
	"fmt"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// DomainEvents wraps a search index for domain events (unified audit + notification source).
type DomainEvents struct {
	index *sum.Search[models.DomainEvent]
	qb    *lucene.Builder[models.DomainEvent]
}

// NewDomainEvents creates a new domain events search store.
func NewDomainEvents(provider grub.SearchProvider) *DomainEvents {
	return &DomainEvents{
		index: sum.NewSearch[models.DomainEvent](provider, "domain_events"),
		qb:    lucene.New[models.DomainEvent](),
	}
}

// Index writes a domain event to the search index.
func (s *DomainEvents) Index(ctx context.Context, event *models.DomainEvent) error {
	return s.index.Index(ctx, event.ID, event)
}

// Search queries domain events scoped to a tenant. TenantID is required.
func (s *DomainEvents) Search(ctx context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
	if params.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required for scoped search")
	}
	return s.search(ctx, params)
}

// SearchAll queries domain events across all tenants (admin use only).
func (s *DomainEvents) SearchAll(ctx context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
	return s.search(ctx, params)
}

func (s *DomainEvents) search(ctx context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
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
		return nil, fmt.Errorf("searching domain events: %w", err)
	}

	items := make([]*models.DomainEvent, len(result.Hits))
	for i, hit := range result.Hits {
		e := hit.Content
		items[i] = &e
	}
	return &models.OffsetResult[models.DomainEvent]{
		Items:  items,
		Total:  result.Total,
		Offset: params.Offset,
	}, nil
}
