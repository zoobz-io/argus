// Package main is the entry point for the connector service.
//
// The connector syncs documents from cloud storage providers into the
// ingestion pipeline. It watches for changes via provider polling,
// downloads content, stages it in MinIO, and publishes to the ingest queue.
//
// This PR: skeleton with infrastructure setup and credential management.
// Sync and fetch will be added in subsequent PRs.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/internal/boot"
	"github.com/zoobz-io/argus/internal/connector"
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
	_ = redisClient // Used by herald (PR 2).

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
	_ = credManager // Used by sync loop (PR 2).

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
	// 7. Block Until Shutdown
	// =========================================================================

	log.Println("connector ready")
	<-ctx.Done()
	log.Println("shutting down...")
	return nil
}
