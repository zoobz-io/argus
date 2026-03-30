// Package main is the entry point for the application.
package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/cereal"
	"github.com/zoobz-io/herald"
	heraldredis "github.com/zoobz-io/herald/redis"
	astqlpg "github.com/zoobz-io/astql/postgres"
	grubredis "github.com/zoobz-io/grub/redis"
	"github.com/zoobz-io/sum"

	admincontracts "github.com/zoobz-io/argus/admin/contracts"
	adminhandlers "github.com/zoobz-io/argus/admin/handlers"
	apicontracts "github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/handlers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/internal/auth"
	"github.com/zoobz-io/argus/internal/boot"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/internal/ratelimit"
	"github.com/zoobz-io/argus/internal/ingest"
	"github.com/zoobz-io/argus/internal/vocabulary"
	"github.com/zoobz-io/argus/models"
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
	if err := sum.Config[config.Classify](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load classify config: %w", err)
	}
	if err := sum.Config[config.Embedding](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load embedding config: %w", err)
	}
	if err := sum.Config[config.OTEL](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load otel config: %w", err)
	}
	if err := sum.Config[config.Encryption](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load encryption config: %w", err)
	}
	if err := sum.Config[config.Auth](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load auth config: %w", err)
	}
	if err := sum.Config[config.RateLimit](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load rate limit config: %w", err)
	}
	encCfg := sum.MustUse[config.Encryption](ctx)
	encKey, err := hex.DecodeString(encCfg.Key)
	if err != nil {
		return fmt.Errorf("failed to decode encryption key: %w", err)
	}
	aesEncryptor, err := cereal.AES(encKey)
	if err != nil {
		return fmt.Errorf("failed to create AES encryptor: %w", err)
	}
	svc.WithEncryptor(cereal.EncryptAES, aesEncryptor)

	// =========================================================================
	// 2. Connect to Infrastructure
	// =========================================================================

	db, err := boot.Database(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	capitan.Emit(ctx, events.StartupDatabaseConnected)

	bucketProvider, err := boot.Storage(ctx)
	if err != nil {
		return err
	}
	capitan.Emit(ctx, events.StartupStorageConnected)

	redisClient, err := boot.Redis(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = redisClient.Close() }()
	redisProvider := grubredis.New(redisClient)
	_ = redisProvider // Available for cache stores.
	capitan.Emit(ctx, events.StartupRedisConnected)

	searchProvider, err := boot.OpenSearch(ctx)
	if err != nil {
		return err
	}
	capitan.Emit(ctx, events.StartupOpenSearchConnected)

	osCfg := sum.MustUse[config.OpenSearch](ctx)
	if err = boot.EnsureIndices(ctx, osCfg.Addr); err != nil {
		return fmt.Errorf("ensuring opensearch indices: %w", err)
	}

	classifyConn, classifyClient, err := boot.Classify(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = classifyConn.Close() }()

	// Authentication (OIDC) — upserter wired after stores are created (section 3).
	authCfg := sum.MustUse[config.Auth](ctx)
	embedService, err := boot.Embedding(ctx)
	if err != nil {
		return err
	}

	// =========================================================================
	// 3. Create and Register Stores
	// =========================================================================

	renderer := astqlpg.New()
	allStores := stores.New(db, renderer, bucketProvider, searchProvider)

	// Public API contracts
	sum.Register[apicontracts.Tenants](k, allStores.Tenants)
	sum.Register[apicontracts.Providers](k, allStores.Providers)
	sum.Register[apicontracts.WatchedPaths](k, allStores.WatchedPaths)
	sum.Register[apicontracts.Documents](k, allStores.Documents)
	sum.Register[apicontracts.DocumentVersions](k, allStores.DocumentVersions)
	sum.Register[apicontracts.DocumentVersionSearch](k, allStores.DocumentVersionSearch)
	sum.Register[apicontracts.QueryEmbedder](k, embedService)
	sum.Register[apicontracts.Topics](k, allStores.Topics)
	sum.Register[apicontracts.Tags](k, allStores.Tags)
	sum.Register[apicontracts.JobReader](k, allStores.Jobs)
	sum.Register[apicontracts.Users](k, allStores.Users)
	sum.Register[apicontracts.Subscriptions](k, allStores.Subscriptions)
	sum.Register[apicontracts.Notifications](k, allStores.Notifications)
	sum.Register[apicontracts.AuditLog](k, allStores.Audit)
	sum.Register[apicontracts.Hooks](k, allStores.Hooks)
	sum.Register[apicontracts.Deliveries](k, allStores.Deliveries)

	// Admin API contracts
	sum.Register[admincontracts.Tenants](k, allStores.Tenants)
	sum.Register[admincontracts.Providers](k, allStores.Providers)
	sum.Register[admincontracts.WatchedPaths](k, allStores.WatchedPaths)
	sum.Register[admincontracts.Documents](k, allStores.Documents)
	sum.Register[admincontracts.DocumentVersions](k, allStores.DocumentVersions)
	sum.Register[admincontracts.DocumentVersionSearch](k, allStores.DocumentVersionSearch)
	sum.Register[admincontracts.Topics](k, allStores.Topics)
	sum.Register[admincontracts.Tags](k, allStores.Tags)
	sum.Register[admincontracts.Users](k, allStores.Users)
	sum.Register[admincontracts.Subscriptions](k, allStores.AdminSubscriptions)
	sum.Register[admincontracts.AuditLog](k, allStores.Audit)
	sum.Register[admincontracts.Hooks](k, allStores.AdminHooks)

	// Internal contracts — enqueuer needs versions, documents, jobs.
	// Classifier needed by vocabulary pipeline.
	sum.Register[intcontracts.IngestVersions](k, allStores.DocumentVersions)
	sum.Register[intcontracts.IngestDocuments](k, allStores.Documents)
	sum.Register[intcontracts.IngestJobs](k, allStores.Jobs)
	sum.Register[intcontracts.Classifier](k, classifyClient)

	// Async ingestion enqueuer
	enqueuer := ingest.NewEnqueuer()
	sum.Register[apicontracts.IngestEnqueuer](k, enqueuer)

	vocabPipeline := vocabulary.New()
	sum.Register[apicontracts.Vocabulary](k, vocabPipeline)
	sum.Register[admincontracts.Vocabulary](k, vocabPipeline)

	// Authentication (OIDC) — now that stores exist, wire the user upserter.
	authenticator, err := auth.NewAuthenticator(ctx, authCfg.Issuer, authCfg.Audience, &userUpserterAdapter{store: allStores.Users, subs: allStores.Subscriptions})
	if err != nil {
		return fmt.Errorf("failed to create oidc authenticator: %w", err)
	}
	svc.Engine().WithAuthenticator(authenticator)
	log.Println("oidc authenticator initialized")
	capitan.Emit(ctx, events.StartupAuthReady)

	// =========================================================================
	// 4. Register Boundaries
	// =========================================================================

	sum.NewBoundary[models.Tenant](k)
	sum.NewBoundary[models.Provider](k)
	sum.NewBoundary[models.WatchedPath](k)
	sum.NewBoundary[models.Document](k)
	sum.NewBoundary[models.DocumentVersion](k)
	sum.NewBoundary[models.User](k)
	sum.NewBoundary[models.Subscription](k)
	sum.NewBoundary[models.Hook](k)
	sum.NewBoundary[models.Delivery](k)
	wire.RegisterBoundaries(k)

	// =========================================================================
	// 5. Freeze Registry
	// =========================================================================

	sum.Freeze(k)
	capitan.Emit(ctx, events.StartupServicesReady)

	// =========================================================================
	// 6. Initialize Observability (OTEL + Aperture)
	// =========================================================================

	otelProviders, err := boot.OTEL(ctx, "argus")
	if err != nil {
		return err
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()
	capitan.Emit(ctx, events.StartupOTELReady)

	ap, err := boot.Aperture(ctx, otelProviders)
	if err != nil {
		return err
	}
	defer ap.Close()
	capitan.Emit(ctx, events.StartupApertureReady)

	// =========================================================================
	// 7. Herald: Ingestion Queue Publisher + Job Status Subscriber
	// =========================================================================

	ingestStream := heraldredis.New("argus:ingestion", heraldredis.WithClient(redisClient))
	ingestPub := herald.NewPublisher(
		ingestStream,
		events.IngestQueueSignal,
		events.IngestQueueKey,
		[]herald.Option[events.IngestMessage]{
			herald.WithRetry[events.IngestMessage](3),
		},
	)
	ingestPub.Start()
	defer func() { _ = ingestPub.Close() }()
	log.Println("ingestion queue publisher initialized")

	// Subscriber: job-status via Pub/Sub for broadcast to all app instances.
	// No consumer groups — Pub/Sub is fire-and-forget fanout.
	jobStatusStream := heraldredis.NewPubSub("argus:job-status",
		heraldredis.WithPubSubClient(redisClient),
	)
	jobStatusSub := herald.NewSubscriber(
		jobStatusStream,
		events.JobStatusSignal,
		events.JobStatusKey,
		[]herald.Option[events.JobStatusEvent]{},
	)
	jobStatusSub.Start(ctx)
	defer func() { _ = jobStatusSub.Close() }()
	log.Println("job status subscriber initialized")

	notifyHintStream := heraldredis.NewPubSub("argus:notify-hints",
		heraldredis.WithPubSubClient(redisClient),
	)
	notifyHintSub := herald.NewSubscriber(
		notifyHintStream,
		events.NotifyHintSignal,
		events.NotifyHintKey,
		[]herald.Option[events.NotifyHint]{},
	)
	notifyHintSub.Start(ctx)
	defer func() { _ = notifyHintSub.Close() }()
	log.Println("notify hints subscriber initialized")

	// Herald Publisher: AuditSignal → argus:audit stream
	auditStream := heraldredis.New("argus:audit", heraldredis.WithClient(redisClient))
	auditPub := herald.NewPublisher(
		auditStream,
		events.AuditSignal,
		events.AuditKey,
		[]herald.Option[models.AuditEntry]{
			herald.WithRetry[models.AuditEntry](3),
			herald.WithBackoff[models.AuditEntry](3, 500*time.Millisecond),
		},
	)
	auditPub.Start()
	defer func() { _ = auditPub.Close() }()
	log.Println("audit publisher initialized")

	// =========================================================================
	// 8. Register Handlers and Start Server
	// =========================================================================

	// Rate limiting middleware — runs before auth, keyed by client IP.
	rlCfg := sum.MustUse[config.RateLimit](ctx)
	rateLimiter := ratelimit.New(redisClient, rlCfg.RequestsPerMinute, rlCfg.TrustProxy)
	svc.Engine().WithMiddleware(rateLimiter.Middleware())

	svc.Handle(handlers.All()...)
	svc.Handle(adminhandlers.All()...)

	appCfg := sum.MustUse[config.App](ctx)
	capitan.Emit(ctx, events.StartupServerListening, events.StartupPortKey.Field(appCfg.Port))
	log.Printf("starting server on port %d...", appCfg.Port)

	_ = ap // Remove when using ap.Apply() above.

	return svc.Run("", appCfg.Port)
}

// userUpserterAdapter adapts the Users store to the auth.UserUpserter interface.
type userUpserterAdapter struct {
	store interface {
		UpsertFromClaims(ctx context.Context, externalID, tenantID, email, displayName string) (*models.User, error)
	}
	subs interface {
		CreateDefaultSubscriptions(ctx context.Context, userID, tenantID string) error
	}
}

func (a *userUpserterAdapter) UpsertFromClaims(ctx context.Context, externalID, tenantID, email, displayName string) error {
	user, err := a.store.UpsertFromClaims(ctx, externalID, tenantID, email, displayName)
	if err != nil {
		return err
	}
	if err := a.subs.CreateDefaultSubscriptions(ctx, user.ID, user.TenantID); err != nil {
		log.Printf("warning: failed to create default subscriptions for user %s: %v", user.ID, err)
	}
	return nil
}
