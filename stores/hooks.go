package stores

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// Hooks provides database access for webhook endpoints.
type Hooks struct {
	*sum.Database[models.Hook]
}

// NewHooks creates a new hooks store.
func NewHooks(db *sqlx.DB, renderer astql.Renderer) *Hooks {
	return &Hooks{
		Database: sum.NewDatabase[models.Hook](db, "hooks", renderer),
	}
}

// CreateHook creates a new webhook endpoint with a generated signing secret.
func (s *Hooks) CreateHook(ctx context.Context, tenantID, userID, url string) (*models.Hook, error) {
	secret, err := generateSecret()
	if err != nil {
		return nil, fmt.Errorf("generating hook secret: %w", err)
	}
	h := &models.Hook{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		UserID:   userID,
		URL:      url,
		Secret:   secret,
		Active:   true,
	}
	if err := s.Set(ctx, "", h); err != nil {
		return nil, fmt.Errorf("creating hook: %w", err)
	}
	return h, nil
}

// GetHook retrieves a hook by ID.
func (s *Hooks) GetHook(ctx context.Context, id string) (*models.Hook, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// GetHookByTenant retrieves a hook by ID scoped to a tenant.
func (s *Hooks) GetHookByTenant(ctx context.Context, tenantID, id string) (*models.Hook, error) {
	params := map[string]any{"id": id, "tenant_id": tenantID}
	q := s.Query().
		Where("id", "=", "id").
		Where("tenant_id", "=", "tenant_id").
		Limit(1)
	results, err := q.Exec(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("getting hook by tenant: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("hook not found")
	}
	return results[0], nil
}

// GetWithSecret retrieves a hook by ID scoped to a tenant, with its secret decrypted (for HMAC signing).
func (s *Hooks) GetWithSecret(ctx context.Context, tenantID, id string) (*models.Hook, error) {
	params := map[string]any{"id": id, "tenant_id": tenantID}
	results, err := s.Query().
		Where("id", "=", "id").
		Where("tenant_id", "=", "tenant_id").
		Limit(1).
		Exec(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("loading hook with secret: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("hook not found")
	}
	return results[0], nil
}

// ListHooksByTenant retrieves a paginated list of hooks for a tenant.
func (s *Hooks) ListHooksByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Hook], error) {
	params := map[string]any{"tenant_id": tenantID}
	items, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, params)
	if err != nil {
		return nil, err
	}
	total, countErr := s.Count().
		Where("tenant_id", "=", "tenant_id").
		Exec(ctx, params)
	if countErr != nil {
		return nil, countErr
	}
	return &models.OffsetResult[models.Hook]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}

// ListHooks retrieves a paginated list of all hooks (cross-tenant).
func (s *Hooks) ListHooks(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Hook], error) {
	items, err := s.Query().
		OrderBy("created_at", "ASC").
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
	return &models.OffsetResult[models.Hook]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}

// DeleteHook removes a hook by ID scoped to a tenant.
func (s *Hooks) DeleteHook(ctx context.Context, tenantID, id string) error {
	_, err := s.GetHookByTenant(ctx, tenantID, id)
	if err != nil {
		return err
	}
	return s.Delete(ctx, id)
}

// AdminHooks wraps Hooks to satisfy admin/contracts.Hooks.
type AdminHooks struct {
	*Hooks
	deliveries *Deliveries
}

// DeleteHook removes a hook by ID (admin — no tenant scope).
func (a *AdminHooks) DeleteHook(ctx context.Context, id string) error {
	return a.Delete(ctx, id)
}

// ListDeliveries retrieves a paginated list of all deliveries (cross-tenant).
func (a *AdminHooks) ListDeliveries(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Delivery], error) {
	return a.deliveries.ListDeliveries(ctx, page)
}

// ListHooks retrieves a paginated list of all hooks — admin override to avoid ambiguity.
func (a *AdminHooks) ListHooks(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Hook], error) {
	return a.Hooks.ListHooks(ctx, page)
}

// generateSecret generates a 32-byte hex-encoded signing secret.
func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
