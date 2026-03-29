package connector

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/grub"

	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
)

// FetchStore defines the database operations needed by the fetcher.
type FetchStore interface {
	GetProvider(ctx context.Context, id string) (*models.Provider, error)
	CreateJob(ctx context.Context, versionID, documentID, tenantID string) (*models.Job, error)
}

// Fetcher downloads document content from providers and stages it in MinIO.
type Fetcher struct {
	store    FetchStore
	creds    *CredentialManager
	registry *provider.Registry
	bucket   grub.BucketProvider
}

// NewFetcher creates a fetcher backed by the given store, credential manager,
// provider registry, and bucket provider.
func NewFetcher(store FetchStore, creds *CredentialManager, registry *provider.Registry, bucket grub.BucketProvider) *Fetcher {
	return &Fetcher{
		store:    store,
		creds:    creds,
		registry: registry,
		bucket:   bucket,
	}
}

// HandleFetch downloads content from the provider, writes it to MinIO,
// creates an ingestion job, and emits IngestQueueSignal.
func (f *Fetcher) HandleFetch(ctx context.Context, msg events.FetchMessage) error {
	// Load the provider model to get the type.
	prov, err := f.store.GetProvider(ctx, msg.ProviderID)
	if err != nil {
		return fmt.Errorf("loading provider %s: %w", msg.ProviderID, err)
	}

	// Look up the provider implementation.
	impl, err := f.registry.Get(string(prov.Type))
	if err != nil {
		return fmt.Errorf("registry lookup %s: %w", prov.Type, err)
	}

	// Load credentials.
	creds, err := f.creds.Get(ctx, msg.ProviderID)
	if err != nil {
		return fmt.Errorf("loading credentials: %w", err)
	}

	// Fetch content from the provider.
	rc, meta, updatedCreds, err := impl.Fetch(ctx, creds, msg.Ref)
	if err != nil {
		return fmt.Errorf("provider fetch: %w", err)
	}
	defer func() { _ = rc.Close() }()

	// Read all bytes.
	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("reading fetch response: %w", err)
	}

	// Write to MinIO.
	objInfo := &grub.ObjectInfo{
		Key:         msg.ObjectKey,
		ContentType: meta.MimeType,
		Size:        int64(len(data)),
	}
	err = f.bucket.Put(ctx, msg.ObjectKey, data, objInfo)
	if err != nil {
		return fmt.Errorf("writing to storage: %w", err)
	}
	log.Printf("fetcher: wrote %d bytes to %s", len(data), msg.ObjectKey)

	// Persist refreshed credentials if returned.
	if updatedCreds != nil {
		err = f.creds.Update(ctx, msg.ProviderID, updatedCreds)
		if err != nil {
			return fmt.Errorf("updating credentials: %w", err)
		}
	}

	// Create an ingestion job.
	job, err := f.store.CreateJob(ctx, msg.VersionID, msg.DocumentID, msg.TenantID)
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}
	log.Printf("fetcher: created job %s for version %s", job.ID, msg.VersionID)

	// Emit to the ingest queue so the worker picks it up.
	capitan.Emit(ctx, events.IngestQueueSignal, events.IngestQueueKey.Field(events.IngestMessage{
		JobID:     job.ID,
		VersionID: msg.VersionID,
	}))

	return nil
}
