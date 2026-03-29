package stores

import (
	"context"
	"fmt"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// Notifications wraps a search index for application notifications.
type Notifications struct {
	index *sum.Search[models.Notification]
	qb    *lucene.Builder[models.Notification]
}

// NewNotifications creates a new notification search store.
func NewNotifications(provider grub.SearchProvider) *Notifications {
	return &Notifications{
		index: sum.NewSearch[models.Notification](provider, "notifications"),
		qb:    lucene.New[models.Notification](),
	}
}

// Index writes a notification to the search index.
func (s *Notifications) Index(ctx context.Context, n *models.Notification) error {
	return s.index.Index(ctx, n.ID, n)
}

// SearchByUser queries notifications for a user within a tenant.
func (s *Notifications) SearchByUser(ctx context.Context, tenantID, userID string, page models.OffsetPage) (*models.OffsetResult[models.Notification], error) {
	filters := []lucene.Query{
		s.qb.Term("user_id", userID),
		s.qb.Term("tenant_id", tenantID),
	}

	search := lucene.NewSearch().
		Query(s.qb.Bool().Filter(filters...)).
		Sort(lucene.SortField{Field: "created_at", Order: "desc"}).
		Size(page.PageSize()).
		From(page.Offset)

	result, err := s.index.Execute(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("searching notifications: %w", err)
	}

	items := make([]*models.Notification, len(result.Hits))
	for i, hit := range result.Hits {
		n := hit.Content
		items[i] = &n
	}
	return &models.OffsetResult[models.Notification]{
		Items:  items,
		Total:  result.Total,
		Offset: page.Offset,
	}, nil
}

// UpdateStatus updates the status of a single notification and returns it.
func (s *Notifications) UpdateStatus(ctx context.Context, _, id string, status models.NotificationStatus) (*models.Notification, error) {
	doc, err := s.index.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching notification for status update: %w", err)
	}
	doc.Content.Status = status
	if err := s.index.Index(ctx, id, &doc.Content); err != nil {
		return nil, fmt.Errorf("indexing notification after status update: %w", err)
	}
	return &doc.Content, nil
}

// BulkUpdateStatus marks all notifications for a user within a tenant as the given status.
func (s *Notifications) BulkUpdateStatus(ctx context.Context, tenantID, userID string, status models.NotificationStatus) error {
	// Find all notifications for this user+tenant.
	filters := []lucene.Query{
		s.qb.Term("user_id", userID),
		s.qb.Term("tenant_id", tenantID),
	}
	search := lucene.NewSearch().
		Query(s.qb.Bool().Filter(filters...)).
		Size(1000)

	result, err := s.index.Execute(ctx, search)
	if err != nil {
		return fmt.Errorf("searching notifications for bulk update: %w", err)
	}

	if len(result.Hits) == 0 {
		return nil
	}

	docs := make(map[string]*models.Notification, len(result.Hits))
	for _, hit := range result.Hits {
		n := hit.Content
		n.Status = status
		docs[n.ID] = &n
	}
	return s.index.IndexBatch(ctx, docs)
}
