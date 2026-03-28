// Package boot provides per-concern infrastructure connection functions.
//
// Each function reads config via sum.MustUse, builds the client, and returns it.
// Callers own lifecycle — defer Close on returned clients. No cleanup, no lifecycle
// management. Context-agnostic — callers pass their own context.
package boot

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/opensearch-project/opensearch-go/v4"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zoobz-io/aperture"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/grub"
	grubminio "github.com/zoobz-io/grub/minio"
	grubopensearch "github.com/zoobz-io/grub/opensearch"
	osrenderer "github.com/zoobz-io/lucene/opensearch"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/config"
	intotel "github.com/zoobz-io/argus/internal/otel"
)

// Database creates a PostgreSQL connection from config.
func Database(ctx context.Context) (*sqlx.DB, error) {
	cfg := sum.MustUse[config.Database](ctx)
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	log.Println("database connected")
	return db, nil
}

// Redis creates a Redis client from config and verifies connectivity.
func Redis(ctx context.Context) (*goredis.Client, error) {
	cfg := sum.MustUse[config.Redis](ctx)
	client := goredis.NewClient(&goredis.Options{
		Addr: cfg.Addr,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}
	log.Println("redis connected")
	return client, nil
}

// OpenSearch creates an OpenSearch search provider from config.
func OpenSearch(ctx context.Context) (grub.SearchProvider, error) {
	cfg := sum.MustUse[config.OpenSearch](ctx)
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{cfg.Addr},
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("creating opensearch client: %w", err)
	}
	provider := grubopensearch.New(client, grubopensearch.Config{
		Version: osrenderer.V2,
	})
	log.Println("opensearch connected")
	return provider, nil
}

// Storage creates a MinIO bucket provider from config.
func Storage(ctx context.Context) (grub.BucketProvider, error) {
	cfg := sum.MustUse[config.Storage](ctx)
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to storage: %w", err)
	}
	provider := grubminio.New(client, cfg.Bucket)
	log.Println("storage connected")
	return provider, nil
}

// OTEL creates OpenTelemetry providers. serviceName identifies the process
// in traces and metrics (e.g., "argus", "argus-worker", "argus-notifier").
// Reads endpoint from config.OTEL via sum.
func OTEL(ctx context.Context, serviceName string) (*intotel.Providers, error) {
	cfg := sum.MustUse[config.OTEL](ctx)
	providers, err := intotel.New(ctx, intotel.Config{
		Endpoint:    cfg.Endpoint,
		ServiceName: serviceName,
	})
	if err != nil {
		return nil, fmt.Errorf("creating otel providers: %w", err)
	}
	log.Println("observability initialized")
	return providers, nil
}

// Aperture creates an aperture bridge from capitan events to OTEL providers.
func Aperture(_ context.Context, providers *intotel.Providers) (*aperture.Aperture, error) {
	ap, err := aperture.New(
		capitan.Default(),
		providers.Log,
		providers.Metric,
		providers.Trace,
	)
	if err != nil {
		return nil, fmt.Errorf("creating aperture: %w", err)
	}
	log.Println("aperture initialized")
	return ap, nil
}

// Hostname returns the system hostname with a UUID fallback.
// Avoids empty consumer names in ephemeral environments.
func Hostname() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "unknown-" + uuid.NewString()[:8]
	}
	return h
}
