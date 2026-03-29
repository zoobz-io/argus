// Package main is the entry point for the connector service.
//
// The connector syncs documents from cloud storage providers into the
// ingestion pipeline. It watches for changes via provider polling,
// downloads content, stages it in MinIO, and publishes to the ingest queue.
//
// Infrastructure setup, credential management, and polling-based change
// detection are in place. Content fetch and ingest queue publishing follow.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/internal/boot"
	"github.com/zoobz-io/argus/internal/connector"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
	"github.com/zoobz-io/argus/provider/googledrive"
	"github.com/zoobz-io/argus/provider/onedrive"
	"github.com/zoobz-io/argus/provider/dropbox"
	"github.com/zoobz-io/argus/provider/s3"
	"github.com/zoobz-io/argus/provider/gcs"
	"github.com/zoobz-io/argus/provider/azureblob"
	"github.com/zoobz-io/argus/stores"

	_ "github.com/zoobz-io/grub/postgres"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("starting connector...")
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize sum service and registry.
	_ = sum.New()
	k := sum.Start()

	// =========================================================================
	// 1. Load Configuration
	// =========================================================================

	if err := sum.Config[config.Database](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}
	if err := sum.Config[config.Storage](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load storage config: %w", err)
	}
	if err := sum.Config[config.Redis](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load redis config: %w", err)
	}
	if err := sum.Config[config.OpenSearch](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load opensearch config: %w", err)
	}
	if err := sum.Config[config.OTEL](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load otel config: %w", err)
	}
	if err := sum.Config[config.Providers](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load providers config: %w", err)
	}
	if err := sum.Config[config.Connector](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load connector config: %w", err)
	}

	// =========================================================================
	// 2. Connect to Infrastructure
	// =========================================================================

	db, err := boot.Database(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	bucketProvider, err := boot.Storage(ctx)
	if err != nil {
		return err
	}
	_ = bucketProvider // Used by fetch (PR 3).

	redisClient, err := boot.Redis(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = redisClient.Close() }()
	_ = redisClient // Used by herald (PR 3).

	searchProvider, err := boot.OpenSearch(ctx)
	if err != nil {
		return err
	}

	// =========================================================================
	// 3. Create Stores
	// =========================================================================

	renderer := astqlpg.New()
	allStores := stores.New(db, renderer, bucketProvider, searchProvider)

	// =========================================================================
	// 4. Register Provider Implementations
	// =========================================================================

	providersCfg := sum.MustUse[config.Providers](ctx)
	registry := provider.NewRegistry()

	if providersCfg.GoogleClientID != "" {
		registry.Register(googledrive.New(googledrive.Config{
			ClientID:     providersCfg.GoogleClientID,
			ClientSecret: providersCfg.GoogleClientSecret,
		}))
	}
	if providersCfg.MicrosoftClientID != "" {
		registry.Register(onedrive.New(onedrive.Config{
			ClientID:     providersCfg.MicrosoftClientID,
			ClientSecret: providersCfg.MicrosoftClientSecret,
		}))
	}
	if providersCfg.DropboxClientID != "" {
		registry.Register(dropbox.New(dropbox.Config{
			ClientID:     providersCfg.DropboxClientID,
			ClientSecret: providersCfg.DropboxClientSecret,
		}))
	}
	// Static credential providers — always registered, no app credentials needed.
	registry.Register(s3.New())
	registry.Register(gcs.New())
	registry.Register(azureblob.New())

	log.Printf("provider registry: %v", registry.Types())

	// =========================================================================
	// 5. Create Credential Manager
	// =========================================================================

	credManager := connector.NewCredentialManager(allStores.Providers)

	// =========================================================================
	// 6. Freeze and Initialize Observability
	// =========================================================================

	sum.Freeze(k)

	otelProviders, err := boot.OTEL(ctx, "argus-connector")
	if err != nil {
		return err
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()

	ap, err := boot.Aperture(ctx, otelProviders)
	if err != nil {
		return err
	}
	defer ap.Close()
	_ = ap

	// =========================================================================
	// 7. Start Syncer
	// =========================================================================

	connectorCfg := sum.MustUse[config.Connector](ctx)
	syncStore := &syncStoreAdapter{
		wp:   allStores.WatchedPaths,
		docs: allStores.Documents,
		vers: allStores.DocumentVersions,
		prov: allStores.Providers,
	}
	syncer := connector.NewSyncer(syncStore, credManager, registry, connectorCfg.PollInterval)

	syncErr := make(chan error, 1)
	go func() {
		syncErr <- syncer.Run(ctx)
	}()

	// =========================================================================
	// 8. Block Until Shutdown
	// =========================================================================

	log.Println("connector ready")
	select {
	case <-ctx.Done():
	case err := <-syncErr:
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("syncer error: %w", err)
		}
	}
	log.Println("shutting down...")
	return nil
}

// syncStoreAdapter composes individual stores to satisfy connector.SyncStore.
type syncStoreAdapter struct {
	wp   *stores.WatchedPaths
	docs *stores.Documents
	vers *stores.DocumentVersions
	prov *stores.Providers
}

func (a *syncStoreAdapter) ListActiveWatchedPaths(ctx context.Context) ([]*models.WatchedPath, error) {
	return a.wp.ListActiveWatchedPaths(ctx)
}

func (a *syncStoreAdapter) UpdateSyncState(ctx context.Context, id string, syncState *string) error {
	return a.wp.UpdateSyncState(ctx, id, syncState)
}

func (a *syncStoreAdapter) GetDocumentByExternalID(ctx context.Context, tenantID, externalID string) (*models.Document, error) {
	return a.docs.GetDocumentByExternalID(ctx, tenantID, externalID)
}

func (a *syncStoreAdapter) CreateDocument(ctx context.Context, doc *models.Document) (*models.Document, error) {
	return a.docs.CreateDocument(ctx, doc)
}

func (a *syncStoreAdapter) CreateDocumentVersion(ctx context.Context, ver *models.DocumentVersion) (*models.DocumentVersion, error) {
	return a.vers.CreateDocumentVersion(ctx, ver)
}

func (a *syncStoreAdapter) GetLatestVersion(ctx context.Context, documentID string) (*models.DocumentVersion, error) {
	return a.vers.GetLatestVersion(ctx, documentID)
}

func (a *syncStoreAdapter) GetProvider(ctx context.Context, id string) (*models.Provider, error) {
	return a.prov.GetProvider(ctx, id)
}
