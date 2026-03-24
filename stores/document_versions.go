package stores

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersions provides database access for document versions.
type DocumentVersions struct {
	*sum.Database[models.DocumentVersion]
}

// NewDocumentVersions creates a new document versions store.
func NewDocumentVersions(db *sqlx.DB, renderer astql.Renderer) (*DocumentVersions, error) {
	database, err := sum.NewDatabase[models.DocumentVersion](db, "document_versions", renderer)
	if err != nil {
		return nil, err
	}
	return &DocumentVersions{Database: database}, nil
}

// GetDocumentVersion retrieves a document version by ID.
func (s *DocumentVersions) GetDocumentVersion(ctx context.Context, id int64) (*models.DocumentVersion, error) {
	return s.Select().
		Where("id", "=", ":id").
		Exec(ctx, map[string]any{"id": id})
}

// DeleteDocumentVersion removes a document version.
func (s *DocumentVersions) DeleteDocumentVersion(ctx context.Context, id int64) error {
	return s.Delete(ctx, fmt.Sprintf("%d", id))
}

// ListDocumentVersions retrieves a paginated list of all document versions.
func (s *DocumentVersions) ListDocumentVersions(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.DocumentVersion], error) {
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

// ListVersionsByDocument retrieves versions for a specific document using cursor pagination.
func (s *DocumentVersions) ListVersionsByDocument(ctx context.Context, documentID int64, page models.CursorPage) (*models.CursorResult[models.DocumentVersion], error) {
	limit := page.PageSize()
	q := s.Query().Where("document_id", "=", ":document_id").OrderBy("id", "ASC").Limit(limit + 1)
	params := map[string]any{"document_id": documentID}
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
