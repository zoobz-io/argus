package stores

import (
	"context"
	"fmt"

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
func NewDocuments(db *sqlx.DB, renderer astql.Renderer) (*Documents, error) {
	database, err := sum.NewDatabase[models.Document](db, "documents", renderer)
	if err != nil {
		return nil, err
	}
	return &Documents{Database: database}, nil
}

// GetDocument retrieves a document by ID.
func (s *Documents) GetDocument(ctx context.Context, id int64) (*models.Document, error) {
	return s.Select().
		Where("id", "=", ":id").
		Exec(ctx, map[string]any{"id": id})
}

// DeleteDocument removes a document.
func (s *Documents) DeleteDocument(ctx context.Context, id int64) error {
	return s.Delete(ctx, fmt.Sprintf("%d", id))
}

// ListDocuments retrieves a paginated list of all documents.
func (s *Documents) ListDocuments(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.Document], error) {
	limit := page.PageSize()
	q := s.Query().OrderBy("id", "ASC").Limit(limit + 1)
	params := map[string]any{}
	if page.Cursor != nil {
		q = q.Where("id", ">", ":cursor")
		params["cursor"] = *page.Cursor
	}
	items, err := q.Exec(ctx, params)
	if err != nil {
		return nil, err
	}
	return cursorResult(items, limit), nil
}

// ListDocumentsByTenant retrieves documents for a specific tenant using cursor pagination.
func (s *Documents) ListDocumentsByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.Document], error) {
	limit := page.PageSize()
	q := s.Query().Where("tenant_id", "=", ":tenant_id").OrderBy("id", "ASC").Limit(limit + 1)
	params := map[string]any{"tenant_id": tenantID}
	if page.Cursor != nil {
		q = q.Where("id", ">", ":cursor")
		params["cursor"] = *page.Cursor
	}
	items, err := q.Exec(ctx, params)
	if err != nil {
		return nil, err
	}
	return cursorResult(items, limit), nil
}
