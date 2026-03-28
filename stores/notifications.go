package stores

import (
	"context"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// Notifications wraps a search index for application notifications.
type Notifications struct {
	index *sum.Search[models.Notification]
}

// NewNotifications creates a new notification search store.
func NewNotifications(provider grub.SearchProvider) *Notifications {
	return &Notifications{
		index: sum.NewSearch[models.Notification](provider, "notifications"),
	}
}

// Index writes a notification to the search index.
func (s *Notifications) Index(ctx context.Context, n *models.Notification) error {
	return s.index.Index(ctx, n.ID, n)
}
