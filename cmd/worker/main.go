// Package main is the entry point for the ingestion worker sidecar.
//
// The worker subscribes to the ingestion queue via herald and runs the
// pipeline for each job. Pipeline signals are bridged to a job-status
// herald stream for SSE fanout to app instances.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/opensearch-project/opensearch-go/v4"
	goredis "github.com/redis/go-redis/v9"
	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/capitan"
	grubminio "github.com/zoobz-io/grub/minio"
	grubopensearch "github.com/zoobz-io/grub/opensearch"
	"github.com/zoobz-io/herald"
	heraldredis "github.com/zoobz-io/herald/redis"
	osrenderer "github.com/zoobz-io/lucene/opensearch"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/vex"
	vexopenai "github.com/zoobz-io/vex/openai"
	"github.com/zoobz-io/zyn"
	zynopenai "github.com/zoobz-io/zyn/openai"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/internal/ingest"
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
	log.Println("starting worker...")
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
	if err := sum.Config[config.OCR](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load ocr config: %w", err)
	}
	if err := sum.Config[config.Convert](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load convert config: %w", err)
	}
	if err := sum.Config[config.Classify](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load classify config: %w", err)
	}
	if err := sum.Config[config.LLM](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load llm config: %w", err)
	}
	if err := sum.Config[config.Embedding](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load embedding config: %w", err)
	}
	if err := sum.Config[config.Worker](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load worker config: %w", err)
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

	// Redis
	redisCfg := sum.MustUse[config.Redis](ctx)
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: redisCfg.Addr,
	})
	defer func() { _ = redisClient.Close() }()
	if pingErr := redisClient.Ping(ctx).Err(); pingErr != nil {
		return fmt.Errorf("failed to connect to redis: %w", pingErr)
	}
	log.Println("redis connected")

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

	// OCR (gRPC)
	ocrCfg := sum.MustUse[config.OCR](ctx)
	ocrConn, err := grpc.NewClient(ocrCfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to ocr service: %w", err)
	}
	defer func() { _ = ocrConn.Close() }()
	ocrClient := proto.NewOCRServiceClient(ocrConn)
	log.Println("ocr service connected")

	// Convert (gRPC)
	convertCfg := sum.MustUse[config.Convert](ctx)
	convertConn, err := grpc.NewClient(convertCfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to convert service: %w", err)
	}
	defer func() { _ = convertConn.Close() }()
	convertClient := proto.NewConvertServiceClient(convertConn)
	log.Println("convert service connected")

	// Classify (gRPC)
	classifyCfg := sum.MustUse[config.Classify](ctx)
	classifyConn, err := grpc.NewClient(classifyCfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to classify service: %w", err)
	}
	defer func() { _ = classifyConn.Close() }()
	classifyClient := proto.NewClassifyServiceClient(classifyConn)
	log.Println("classify service connected")

	// LLM (zyn)
	llmCfg := sum.MustUse[config.LLM](ctx)
	llmProvider := zynopenai.New(zynopenai.Config{
		APIKey:  llmCfg.APIKey,
		Model:   llmCfg.Model,
		BaseURL: llmCfg.BaseURL,
	})
	analyzerSynapse, err := zyn.Extract[models.DocumentAnalysis](
		"Analyze the document content and extract: a concise summary paragraph, the ISO 639-1 language code, and any matching topics and tags from the provided vocabulary lists. Only select topics and tags that clearly apply to the content.",
		llmProvider,
		zyn.WithRetry(3),
		zyn.WithTimeout(60*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to create analyzer synapse: %w", err)
	}
	analyzer := ingest.NewSynapseAnalyzer(analyzerSynapse)
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
	// 3. Create and Register Internal Contracts
	// =========================================================================

	renderer := astqlpg.New()
	allStores := stores.New(db, renderer, bucketProvider, searchProvider)

	sum.Register[intcontracts.IngestVersions](k, allStores.DocumentVersions)
	sum.Register[intcontracts.IngestDocuments](k, allStores.Documents)
	sum.Register[intcontracts.IngestSearch](k, allStores.DocumentVersionSearch)
	sum.Register[intcontracts.IngestJobs](k, allStores.Jobs)
	sum.Register[intcontracts.IngestTopics](k, allStores.Topics)
	sum.Register[intcontracts.IngestTags](k, allStores.Tags)
	sum.Register[intcontracts.OCR](k, ocrClient)
	sum.Register[intcontracts.Converter](k, convertClient)
	sum.Register[intcontracts.Classifier](k, classifyClient)
	sum.Register[intcontracts.Analyzer](k, analyzer)
	sum.Register[intcontracts.Embedder](k, embedService)

	// =========================================================================
	// 4. Create Pipeline and Freeze
	// =========================================================================

	pipeline := ingest.New()
	sum.Freeze(k)

	// =========================================================================
	// 5. Signal Bridge: pipeline signals → JobStatusSignal
	// =========================================================================

	bridge := func(stage string) func(context.Context, *capitan.Event) {
		return func(ctx context.Context, e *capitan.Event) {
			jobID, _ := events.IngestJobIDKey.From(e)
			versionID, _ := events.IngestVersionIDKey.From(e)
			documentID, _ := events.IngestDocumentIDKey.From(e)
			tenantID, _ := events.IngestTenantIDKey.From(e)
			var errMsg string
			if ingestErr, ok := events.IngestErrorKey.From(e); ok && ingestErr != nil {
				errMsg = ingestErr.Error()
			}
			capitan.Emit(ctx, events.JobStatusSignal, events.JobStatusKey.Field(events.JobStatusEvent{
				JobID:      jobID,
				VersionID:  versionID,
				DocumentID: documentID,
				TenantID:   tenantID,
				Stage:      stage,
				Error:      errMsg,
			}))
		}
	}

	capitan.Hook(events.IngestStarted, bridge("started"))
	capitan.Hook(events.IngestExtracted, bridge("extracted"))
	capitan.Hook(events.IngestSummarized, bridge("analyzed"))
	capitan.Hook(events.IngestEmbedded, bridge("embedded"))
	capitan.Hook(events.IngestIndexed, bridge("indexed"))
	capitan.Hook(events.IngestCompleted, bridge("completed"))
	capitan.Hook(events.IngestFailed, bridge("failed"))

	// =========================================================================
	// 6. Herald Publisher: JobStatusSignal → argus:job-status stream
	// =========================================================================

	jobStatusStream := heraldredis.New("argus:job-status", heraldredis.WithClient(redisClient))
	jobStatusPub := herald.NewPublisher(
		jobStatusStream,
		events.JobStatusSignal,
		events.JobStatusKey,
		[]herald.Option[events.JobStatusEvent]{
			herald.WithRetry[events.JobStatusEvent](3),
			herald.WithBackoff[events.JobStatusEvent](3, 500*time.Millisecond),
		},
	)
	jobStatusPub.Start()
	defer func() { _ = jobStatusPub.Close() }()
	log.Println("job status publisher initialized")

	// =========================================================================
	// 7. Notification Hooks + Herald Publisher
	// =========================================================================

	capitan.Hook(events.IngestCompleted, func(ctx context.Context, e *capitan.Event) {
		versionID, _ := events.IngestVersionIDKey.From(e)
		documentID, _ := events.IngestDocumentIDKey.From(e)
		tenantID, _ := events.IngestTenantIDKey.From(e)
		capitan.Emit(ctx, events.NotificationSignal, events.NotificationKey.Field(models.Notification{
			TenantID:   tenantID,
			DocumentID: documentID,
			VersionID:  versionID,
			Type:       models.NotificationIngestCompleted,
			Message:    "Document ingestion completed",
		}))
	})

	capitan.Hook(events.IngestFailed, func(ctx context.Context, e *capitan.Event) {
		versionID, _ := events.IngestVersionIDKey.From(e)
		documentID, _ := events.IngestDocumentIDKey.From(e)
		tenantID, _ := events.IngestTenantIDKey.From(e)
		ingestErr, ok := events.IngestErrorKey.From(e)
		var errMsg string
		if ok && ingestErr != nil {
			errMsg = ingestErr.Error()
		}
		capitan.Emit(ctx, events.NotificationSignal, events.NotificationKey.Field(models.Notification{
			TenantID:   tenantID,
			DocumentID: documentID,
			VersionID:  versionID,
			Type:       models.NotificationIngestFailed,
			Message:    "Document ingestion failed",
			Error:      errMsg,
		}))
	})

	notifStream := heraldredis.New("argus:notifications", heraldredis.WithClient(redisClient))
	notifPub := herald.NewPublisher(
		notifStream,
		events.NotificationSignal,
		events.NotificationKey,
		[]herald.Option[models.Notification]{
			herald.WithRetry[models.Notification](3),
			herald.WithBackoff[models.Notification](3, 500*time.Millisecond),
		},
	)
	notifPub.Start()
	defer func() { _ = notifPub.Close() }()
	log.Println("notification hooks and publisher initialized")

	// =========================================================================
	// 8. Herald Subscriber: argus:ingestion → run pipeline
	// =========================================================================

	workerCfg := sum.MustUse[config.Worker](ctx)
	hostname, _ := os.Hostname()

	ingestStream := heraldredis.New("argus:ingestion",
		heraldredis.WithClient(redisClient),
		heraldredis.WithGroup(workerCfg.ConsumerGroup),
		heraldredis.WithConsumer(hostname),
	)
	ingestSub := herald.NewSubscriber(
		ingestStream,
		events.IngestQueueSignal,
		events.IngestQueueKey,
		[]herald.Option[events.IngestMessage]{
			herald.WithRetry[events.IngestMessage](3),
		},
	)
	ingestSub.Start(ctx)
	defer func() { _ = ingestSub.Close() }()
	log.Println("ingestion queue subscriber started")

	capitan.Hook(events.IngestQueueSignal, func(ctx context.Context, e *capitan.Event) {
		msg, ok := events.IngestQueueKey.From(e)
		if !ok {
			return
		}
		log.Printf("processing job %s (version %s)", msg.JobID, msg.VersionID)
		if err := pipeline.Ingest(ctx, msg.JobID, msg.VersionID); err != nil {
			log.Printf("pipeline error for job %s: %v", msg.JobID, err)
		}
	})

	// =========================================================================
	// 9. Block Until Shutdown
	// =========================================================================

	log.Println("worker ready")
	<-ctx.Done()
	log.Println("shutting down...")
	return nil
}
