package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// Deliveries provides database access for webhook delivery records.
type Deliveries struct {
	*sum.Database[models.Delivery]
}

// NewDeliveries creates a new deliveries store.
func NewDeliveries(db *sqlx.DB, renderer astql.Renderer) *Deliveries {
	return &Deliveries{
		Database: sum.NewDatabase[models.Delivery](db, "deliveries", renderer),
	}
}

// CreateDelivery records a webhook delivery attempt.
func (s *Deliveries) CreateDelivery(ctx context.Context, hookID, eventID, tenantID string, statusCode, attempt int, deliveryErr *string) error {
	d := &models.Delivery{
		ID:         uuid.New().String(),
		HookID:     hookID,
		EventID:    eventID,
		TenantID:   tenantID,
		StatusCode: statusCode,
		Attempt:    attempt,
		Error:      deliveryErr,
	}
	if err := s.Set(ctx, "", d); err != nil {
		return fmt.Errorf("creating delivery: %w", err)
	}
	return nil
}

// ListByHook retrieves a paginated list of deliveries for a specific hook.
func (s *Deliveries) ListByHook(ctx context.Context, tenantID, hookID string, page models.OffsetPage) (*models.OffsetResult[models.Delivery], error) {
	params := map[string]any{"hook_id": hookID, "tenant_id": tenantID}
	items, err := s.Query().
		Where("hook_id", "=", "hook_id").
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "DESC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, params)
	if err != nil {
		return nil, err
	}
	total, countErr := s.Count().
		Where("hook_id", "=", "hook_id").
		Where("tenant_id", "=", "tenant_id").
		Exec(ctx, params)
	if countErr != nil {
		return nil, countErr
	}
	return &models.OffsetResult[models.Delivery]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}

// ListDeliveries retrieves a paginated list of all deliveries (cross-tenant).
func (s *Deliveries) ListDeliveries(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Delivery], error) {
	items, err := s.Query().
		OrderBy("created_at", "DESC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, nil)
	if err != nil {
		return nil, err
	}
	total, countErr := s.Count().Exec(ctx, nil)
	if countErr != nil {
		return nil, countErr
	}
	return &models.OffsetResult[models.Delivery]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}
