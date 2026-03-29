package connector

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
)

// SyncStore defines the database operations needed by the syncer.
type SyncStore interface {
	ListActiveWatchedPaths(ctx context.Context) ([]*models.WatchedPath, error)
	UpdateSyncState(ctx context.Context, id string, syncState *string) error
	GetDocumentByExternalID(ctx context.Context, tenantID, externalID string) (*models.Document, error)
	CreateDocument(ctx context.Context, doc *models.Document) (*models.Document, error)
	CreateDocumentVersion(ctx context.Context, ver *models.DocumentVersion) (*models.DocumentVersion, error)
	GetLatestVersion(ctx context.Context, documentID string) (*models.DocumentVersion, error)
	GetProvider(ctx context.Context, id string) (*models.Provider, error)
}

// Syncer polls cloud storage providers for changes to watched paths.
type Syncer struct {
	store    SyncStore
	creds    *CredentialManager
	registry *provider.Registry
	interval time.Duration
}

// NewSyncer creates a syncer that polls at the given interval.
func NewSyncer(store SyncStore, creds *CredentialManager, registry *provider.Registry, interval time.Duration) *Syncer {
	return &Syncer{
		store:    store,
		creds:    creds,
		registry: registry,
		interval: interval,
	}
}

// Run starts the polling loop. Blocks until ctx is cancelled.
func (s *Syncer) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run once immediately on start.
	s.pollAll(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("syncer: shutting down")
			return ctx.Err()
		case <-ticker.C:
			s.pollAll(ctx)
		}
	}
}

// pollAll lists active watched paths, groups by provider, and polls each.
func (s *Syncer) pollAll(ctx context.Context) {
	paths, err := s.store.ListActiveWatchedPaths(ctx)
	if err != nil {
		log.Printf("syncer: listing active paths: %v", err)
		return
	}

	if len(paths) == 0 {
		return
	}

	for _, wp := range paths {
		if err := s.pollPath(ctx, wp); err != nil {
			log.Printf("syncer: poll path %s (%s): %v", wp.ID, wp.Path, err)
		}
	}
}

// pollPath syncs a single watched path against its provider.
func (s *Syncer) pollPath(ctx context.Context, wp *models.WatchedPath) error {
	// Load the provider model to get the type.
	prov, err := s.store.GetProvider(ctx, wp.ProviderID)
	if err != nil {
		return fmt.Errorf("loading provider %s: %w", wp.ProviderID, err)
	}

	// Look up the provider implementation by type.
	impl, err := s.registry.Get(string(prov.Type))
	if err != nil {
		return fmt.Errorf("registry lookup %s: %w", prov.Type, err)
	}

	// Load credentials.
	creds, err := s.creds.Get(ctx, wp.ProviderID)
	if err != nil {
		return fmt.Errorf("loading credentials: %w", err)
	}

	// Determine sync token from current state.
	var syncToken string
	if wp.SyncState != nil {
		syncToken = *wp.SyncState
	}

	// Call provider for changes.
	changes, newToken, updatedCreds, err := impl.Changes(ctx, creds, wp.Path, syncToken)
	if err != nil {
		return fmt.Errorf("provider changes: %w", err)
	}

	// Persist updated sync state.
	if newToken != "" {
		if err := s.store.UpdateSyncState(ctx, wp.ID, &newToken); err != nil {
			return fmt.Errorf("updating sync state: %w", err)
		}
	}

	// Persist refreshed credentials if the provider returned them.
	if updatedCreds != nil {
		if err := s.creds.Update(ctx, wp.ProviderID, updatedCreds); err != nil {
			return fmt.Errorf("updating credentials: %w", err)
		}
	}

	// Process each change.
	for _, change := range changes {
		if change.Type == provider.ChangeDeleted {
			log.Printf("syncer: path %s: deleted ref %s (skipping)", wp.Path, change.Ref)
			continue
		}

		if change.Entry == nil {
			continue
		}

		if err := s.processChange(ctx, wp, prov, change); err != nil {
			log.Printf("syncer: path %s: processing change ref %s: %v", wp.Path, change.Ref, err)
		}
	}

	if len(changes) > 0 {
		log.Printf("syncer: path %s: processed %d changes", wp.Path, len(changes))
	}

	return nil
}

// processChange finds or creates a document and creates a new version for a change.
func (s *Syncer) processChange(ctx context.Context, wp *models.WatchedPath, prov *models.Provider, change provider.Change) error {
	entry := change.Entry

	// Find existing document by external ID.
	doc, err := s.store.GetDocumentByExternalID(ctx, wp.TenantID, change.Ref)
	if err != nil {
		return fmt.Errorf("looking up document: %w", err)
	}

	if doc == nil {
		// Create new document.
		docID := uuid.New().String()
		doc = &models.Document{
			ID:            docID,
			TenantID:      wp.TenantID,
			ProviderID:    prov.ID,
			WatchedPathID: wp.ID,
			ExternalID:    change.Ref,
			Name:          entry.Name,
			MimeType:      entry.MimeType,
			ObjectKey:     fmt.Sprintf("objects/%s/%s", wp.TenantID, docID),
		}
		doc, err = s.store.CreateDocument(ctx, doc)
		if err != nil {
			return fmt.Errorf("creating document: %w", err)
		}
		log.Printf("syncer: created document %s for %s", doc.ID, entry.Name)
	}

	// Dedup: skip if latest version already has this content hash.
	if doc.ID != "" && entry.ContentHash != "" {
		latest, _ := s.store.GetLatestVersion(ctx, doc.ID)
		if latest != nil && latest.ContentHash == entry.ContentHash {
			return nil // Already processed.
		}
	}

	// Create a new version.
	ver := &models.DocumentVersion{
		ID:            uuid.New().String(),
		DocumentID:    doc.ID,
		TenantID:      wp.TenantID,
		ContentHash:   entry.ContentHash,
		VersionNumber: 1, // Will be properly sequenced by ingest pipeline.
	}
	if _, err := s.store.CreateDocumentVersion(ctx, ver); err != nil {
		return fmt.Errorf("creating document version: %w", err)
	}
	log.Printf("syncer: created version %s for document %s", ver.ID, doc.ID)

	return nil
}
