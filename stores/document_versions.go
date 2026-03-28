package stores

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersions provides database access for document versions.
type DocumentVersions struct {
	*sum.Database[models.DocumentVersion]
	bucket grub.BucketProvider
}

// NewDocumentVersions creates a new document versions store.
func NewDocumentVersions(db *sqlx.DB, renderer astql.Renderer, bucket grub.BucketProvider) *DocumentVersions {
	return &DocumentVersions{
		Database: sum.NewDatabase[models.DocumentVersion](db, "document_versions", renderer),
		bucket:   bucket,
	}
}

// GetDocumentVersion retrieves a document version by ID.
func (s *DocumentVersions) GetDocumentVersion(ctx context.Context, id string) (*models.DocumentVersion, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// DeleteDocumentVersion removes a document version.
func (s *DocumentVersions) DeleteDocumentVersion(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

// ListDocumentVersions retrieves a paginated list of all document versions.
func (s *DocumentVersions) ListDocumentVersions(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error) {
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
	return &models.OffsetResult[models.DocumentVersion]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}

// GetVersionContent retrieves the raw bytes for a document version from object storage.
func (s *DocumentVersions) GetVersionContent(ctx context.Context, objectKey string) ([]byte, error) {
	data, _, err := s.bucket.Get(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("fetching object %s: %w", objectKey, err)
	}
	return data, nil
}

// ListVersionsByDocument retrieves versions for a specific document using offset/limit pagination.
func (s *DocumentVersions) ListVersionsByDocument(ctx context.Context, documentID string, page models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error) {
	params := map[string]any{"document_id": documentID}
	items, err := s.Query().
		Where("document_id", "=", "document_id").
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, params)
	if err != nil {
		return nil, err
	}
	total, countErr := s.Count().
		Where("document_id", "=", "document_id").
		Exec(ctx, params)
	if countErr != nil {
		return nil, countErr
	}
	return &models.OffsetResult[models.DocumentVersion]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}
