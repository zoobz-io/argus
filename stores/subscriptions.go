package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/soy"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// Subscriptions provides database access for notification subscriptions.
type Subscriptions struct {
	*sum.Database[models.Subscription]
}

// NewSubscriptions creates a new subscriptions store.
func NewSubscriptions(db *sqlx.DB, renderer astql.Renderer) *Subscriptions {
	return &Subscriptions{
		Database: sum.NewDatabase[models.Subscription](db, "subscriptions", renderer),
	}
}

// CreateSubscription creates a new subscription.
func (s *Subscriptions) CreateSubscription(ctx context.Context, tenantID, userID, eventType string, channel models.SubscriptionChannel, webhookEndpointID string) (*models.Subscription, error) {
	sub := &models.Subscription{
		ID:        uuid.New().String(),
		UserID:    userID,
		TenantID:  tenantID,
		EventType: eventType,
		Channel:   channel,
	}
	if webhookEndpointID != "" {
		sub.WebhookEndpointID = &webhookEndpointID
	}
	if err := s.Set(ctx, "", sub); err != nil {
		return nil, fmt.Errorf("creating subscription: %w", err)
	}
	return sub, nil
}

// GetSubscription retrieves a subscription by ID.
func (s *Subscriptions) GetSubscription(ctx context.Context, id string) (*models.Subscription, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// GetSubscriptionByTenant retrieves a subscription by ID scoped to a tenant.
func (s *Subscriptions) GetSubscriptionByTenant(ctx context.Context, tenantID, id string) (*models.Subscription, error) {
	params := map[string]any{"id": id, "tenant_id": tenantID}
	q := s.Query().
		Where("id", "=", "id").
		Where("tenant_id", "=", "tenant_id").
		Limit(1)
	results, err := q.Exec(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("getting subscription by tenant: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("subscription not found")
	}
	return results[0], nil
}

// DeleteSubscription removes a subscription by ID after verifying tenant and user ownership.
func (s *Subscriptions) DeleteSubscription(ctx context.Context, tenantID, userID, id string) error {
	sub, err := s.GetSubscriptionByTenant(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if sub.UserID != userID {
		return fmt.Errorf("subscription not found")
	}
	return s.Delete(ctx, id)
}

// ListSubscriptionsByUser retrieves subscriptions for a specific user within a tenant.
func (s *Subscriptions) ListSubscriptionsByUser(ctx context.Context, tenantID, userID string, page models.OffsetPage) (*models.OffsetResult[models.Subscription], error) {
	items, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		Where("user_id", "=", "user_id").
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, map[string]any{"tenant_id": tenantID, "user_id": userID})
	if err != nil {
		return nil, err
	}
	return &models.OffsetResult[models.Subscription]{Items: items, Offset: page.Offset}, nil
}

// ListSubscriptions retrieves a paginated list of all subscriptions (cross-tenant).
func (s *Subscriptions) ListSubscriptions(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Subscription], error) {
	items, err := s.Query().
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &models.OffsetResult[models.Subscription]{Items: items, Offset: page.Offset}, nil
}

// FindByTenantAndEventType retrieves subscriptions matching a tenant and event type.
// Also includes subscriptions with a wildcard event type ("*").
func (s *Subscriptions) FindByTenantAndEventType(ctx context.Context, tenantID, eventType string) ([]*models.Subscription, error) {
	params := map[string]any{"tenant_id": tenantID, "event_type": eventType, "wildcard": "*"}
	items, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		WhereOr(
			soy.C("event_type", "=", "event_type"),
			soy.C("event_type", "=", "wildcard"),
		).
		Exec(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("finding subscriptions by tenant and event type: %w", err)
	}
	return items, nil
}

// CreateDefaultSubscriptions creates default inbox subscriptions for ingest.completed and ingest.failed.
// Conflicts are silently ignored.
func (s *Subscriptions) CreateDefaultSubscriptions(ctx context.Context, userID, tenantID string) error {
	eventTypes := []string{
		string(models.NotificationIngestCompleted),
		string(models.NotificationIngestFailed),
	}
	for _, et := range eventTypes {
		sub := &models.Subscription{
			ID:        uuid.New().String(),
			UserID:    userID,
			TenantID:  tenantID,
			EventType: et,
			Channel:   models.SubscriptionChannelInbox,
		}
		_, err := s.Insert().
			OnConflict("user_id", "event_type", "channel").
			DoNothing().
			Exec(ctx, sub)
		if err != nil {
			return fmt.Errorf("creating default subscription for %s: %w", et, err)
		}
	}
	return nil
}

// AdminSubscriptions wraps Subscriptions to satisfy admin/contracts.Subscriptions.
type AdminSubscriptions struct {
	*Subscriptions
}

// DeleteSubscription removes a subscription by ID (admin — no tenant scope).
func (a *AdminSubscriptions) DeleteSubscription(ctx context.Context, id string) error {
	return a.Delete(ctx, id)
}
