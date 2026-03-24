// Package main is the entry point for the application.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/opensearch-project/opensearch-go/v4"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zoobz-io/aperture"
	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/capitan"
	grubminio "github.com/zoobz-io/grub/minio"
	grubopensearch "github.com/zoobz-io/grub/opensearch"
	grubredis "github.com/zoobz-io/grub/redis"
	osrenderer "github.com/zoobz-io/lucene/opensearch"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/vex"
	vexopenai "github.com/zoobz-io/vex/openai"
	zynopenai "github.com/zoobz-io/zyn/openai"
	"github.com/zoobz-io/zyn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	admincontracts "github.com/zoobz-io/argus/admin/contracts"
	adminhandlers "github.com/zoobz-io/argus/admin/handlers"
	apicontracts "github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/handlers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/internal/ingest"
	intotel "github.com/zoobz-io/argus/internal/otel"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/proto"
	"github.com/zoobz-io/argus/stores"

	_ "github.com/zoobz-io/grub/postgres"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("starting...")
	ctx := context.Background()

	// Initialize sum service and registry.
	svc := sum.New()
	k := sum.Start()

	// =========================================================================
	// 1. Load Configuration
	// =========================================================================

	if err := sum.Config[config.App](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}
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
	if err := sum.Config[config.OCR](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load ocr config: %w", err)
	}
	if err := sum.Config[config.Encryption](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load encryption config: %w", err)
	}
	if err := sum.Config[config.LLM](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load llm config: %w", err)
	}
	if err := sum.Config[config.Embedding](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load embedding config: %w", err)
	}

	// =========================================================================
	// 2. Connect to Infrastructure
	// =========================================================================

	// Database
	dbCfg := sum.MustUse[config.Database](ctx)
	db, err := sqlx.Connect("postgres", dbCfg.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()
	log.Println("database connected")
	capitan.Emit(ctx, events.StartupDatabaseConnected)

	// Storage (MinIO)
	storageCfg := sum.MustUse[config.Storage](ctx)
	minioClient, err := minio.New(storageCfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(storageCfg.AccessKey, storageCfg.SecretKey, ""),
		Secure: storageCfg.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to storage: %w", err)
	}
	bucketProvider := grubminio.New(minioClient, storageCfg.Bucket)
	log.Println("storage connected")
	capitan.Emit(ctx, events.StartupStorageConnected)

	// Redis
	redisCfg := sum.MustUse[config.Redis](ctx)
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: redisCfg.Addr,
	})
	defer func() { _ = redisClient.Close() }()
	if pingErr := redisClient.Ping(ctx).Err(); pingErr != nil {
		return fmt.Errorf("failed to connect to redis: %w", pingErr)
	}
	redisProvider := grubredis.New(redisClient)
	_ = redisProvider // Available for cache stores.
	log.Println("redis connected")
	capitan.Emit(ctx, events.StartupRedisConnected)

	// OpenSearch
	osCfg := sum.MustUse[config.OpenSearch](ctx)
	osClient, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{osCfg.Addr},
		Username:  osCfg.Username,
		Password:  osCfg.Password,
	})
	if err != nil {
		return fmt.Errorf("failed to create opensearch client: %w", err)
	}
	searchProvider := grubopensearch.New(osClient, grubopensearch.Config{
		Version: osrenderer.V2,
	})
	log.Println("opensearch connected")
	capitan.Emit(ctx, events.StartupOpenSearchConnected)

	// OCR (gRPC)
	ocrCfg := sum.MustUse[config.OCR](ctx)
	ocrConn, err := grpc.NewClient(ocrCfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to ocr service: %w", err)
	}
	defer func() { _ = ocrConn.Close() }()
	ocrClient := proto.NewOCRServiceClient(ocrConn)
	log.Println("ocr service connected")
	capitan.Emit(ctx, events.StartupOCRConnected)

	// LLM (zyn)
	llmCfg := sum.MustUse[config.LLM](ctx)
	llmProvider := zynopenai.New(zynopenai.Config{
		APIKey:  llmCfg.APIKey,
		Model:   llmCfg.Model,
		BaseURL: llmCfg.BaseURL,
	})
	summarySynapse, err := zyn.Transform(
		"Summarize the following document content into a concise paragraph that captures the key points and themes.",
		llmProvider,
		zyn.WithRetry(3),
		zyn.WithTimeout(60*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to create summary synapse: %w", err)
	}
	summarizer := ingest.NewSynapseSummarizer(summarySynapse)

	extractionSynapse, err := zyn.Transform(
		"Extract and return the readable text content from the following raw document data. Preserve structure (paragraphs, headings, lists, tables) as plain text. Remove markup, control codes, and formatting artifacts.",
		llmProvider,
		zyn.WithRetry(3),
		zyn.WithTimeout(60*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to create extraction synapse: %w", err)
	}
	textExtractor := ingest.NewSynapseExtractor(extractionSynapse)
	log.Println("llm provider initialized")

	// Embeddings (vex)
	embeddingCfg := sum.MustUse[config.Embedding](ctx)
	embeddingProvider := vexopenai.New(vexopenai.Config{
		APIKey:     embeddingCfg.APIKey,
		Model:      embeddingCfg.Model,
		BaseURL:    embeddingCfg.BaseURL,
		Dimensions: embeddingCfg.Dimensions,
	})
	embedService := vex.NewService(embeddingProvider,
		vex.WithRetry(3),
		vex.WithTimeout(30*time.Second),
	)
	log.Println("embedding provider initialized")

	// =========================================================================
	// 3. Create and Register Stores
	// =========================================================================

	renderer := astqlpg.New()

	allStores, err := stores.New(db, renderer, bucketProvider, searchProvider)
	if err != nil {
		return fmt.Errorf("failed to create stores: %w", err)
	}

	// Public API contracts
	sum.Register[apicontracts.Tenants](k, allStores.Tenants)
	sum.Register[apicontracts.Providers](k, allStores.Providers)
	sum.Register[apicontracts.WatchedPaths](k, allStores.WatchedPaths)
	sum.Register[apicontracts.Documents](k, allStores.Documents)
	sum.Register[apicontracts.DocumentVersions](k, allStores.DocumentVersions)
	sum.Register[apicontracts.DocumentVersionSearch](k, allStores.DocumentVersionSearch)

	// Admin API contracts
	sum.Register[admincontracts.Tenants](k, allStores.Tenants)
	sum.Register[admincontracts.Providers](k, allStores.Providers)
	sum.Register[admincontracts.WatchedPaths](k, allStores.WatchedPaths)
	sum.Register[admincontracts.Documents](k, allStores.Documents)
	sum.Register[admincontracts.DocumentVersions](k, allStores.DocumentVersions)
	sum.Register[admincontracts.DocumentVersionSearch](k, allStores.DocumentVersionSearch)

	// Internal contracts (ingestion pipeline)
	sum.Register[intcontracts.IngestVersions](k, allStores.DocumentVersions)
	sum.Register[intcontracts.IngestDocuments](k, allStores.Documents)
	sum.Register[intcontracts.IngestSearch](k, allStores.DocumentVersionSearch)
	sum.Register[intcontracts.OCR](k, ocrClient)
	sum.Register[intcontracts.Summarizer](k, summarizer)
	sum.Register[intcontracts.TextExtractor](k, textExtractor)
	sum.Register[intcontracts.Embedder](k, embedService)

	// Pipeline contract (public API)
	pipeline := ingest.New()
	sum.Register[apicontracts.Ingest](k, pipeline)

	// =========================================================================
	// 4. Register Boundaries
	// =========================================================================

	// Model boundaries
	_, err = sum.NewBoundary[models.Tenant](k)
	if err != nil {
		return fmt.Errorf("failed to create tenant boundary: %w", err)
	}
	_, err = sum.NewBoundary[models.Provider](k)
	if err != nil {
		return fmt.Errorf("failed to create provider boundary: %w", err)
	}
	_, err = sum.NewBoundary[models.WatchedPath](k)
	if err != nil {
		return fmt.Errorf("failed to create watched path boundary: %w", err)
	}
	_, err = sum.NewBoundary[models.Document](k)
	if err != nil {
		return fmt.Errorf("failed to create document boundary: %w", err)
	}
	_, err = sum.NewBoundary[models.DocumentVersion](k)
	if err != nil {
		return fmt.Errorf("failed to create document version boundary: %w", err)
	}

	// Wire boundaries
	err = wire.RegisterBoundaries(k)
	if err != nil {
		return fmt.Errorf("failed to register wire boundaries: %w", err)
	}

	// =========================================================================
	// 5. Freeze Registry
	// =========================================================================

	sum.Freeze(k)
	capitan.Emit(ctx, events.StartupServicesReady)

	// =========================================================================
	// 6. Initialize Observability (OTEL + Aperture)
	// =========================================================================

	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "localhost:4318"
	}
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "argus"
	}

	otelProviders, err := intotel.New(ctx, intotel.Config{
		Endpoint:    otelEndpoint,
		ServiceName: serviceName,
	})
	if err != nil {
		return fmt.Errorf("failed to create otel providers: %w", err)
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()
	log.Println("observability initialized")
	capitan.Emit(ctx, events.StartupOTELReady)

	// Initialize aperture to bridge capitan events → OTEL.
	ap, err := aperture.New(
		capitan.Default(),
		otelProviders.Log,
		otelProviders.Metric,
		otelProviders.Trace,
	)
	if err != nil {
		return fmt.Errorf("failed to create aperture: %w", err)
	}
	defer ap.Close()
	capitan.Emit(ctx, events.StartupApertureReady)

	// =========================================================================
	// 7. Initialize Pipeline and Register Handlers
	// =========================================================================

	svc.Handle(handlers.All()...)
	svc.Handle(adminhandlers.All()...)

	appCfg := sum.MustUse[config.App](ctx)
	capitan.Emit(ctx, events.StartupServerListening, events.StartupPortKey.Field(appCfg.Port))
	log.Printf("starting server on port %d...", appCfg.Port)

	_ = ap // Remove when using ap.Apply() above.

	return svc.Run("", appCfg.Port)
}
