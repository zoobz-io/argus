package stores

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// Documents provides database access for documents.
type Documents struct {
	*sum.Database[models.Document]
}

// NewDocuments creates a new documents store.
func NewDocuments(db *sqlx.DB, renderer astql.Renderer) *Documents {
	return &Documents{
		Database: sum.NewDatabase[models.Document](db, "documents", renderer),
	}
}

// GetDocument retrieves a document by ID.
func (s *Documents) GetDocument(ctx context.Context, id string) (*models.Document, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// DeleteDocument removes a document.
func (s *Documents) DeleteDocument(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

// ListDocuments retrieves a paginated list of all documents.
func (s *Documents) ListDocuments(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Document], error) {
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
	return &models.OffsetResult[models.Document]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}

// ListDocumentsByTenant retrieves documents for a specific tenant using offset/limit pagination.
func (s *Documents) ListDocumentsByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Document], error) {
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
	return &models.OffsetResult[models.Document]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}
