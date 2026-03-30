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
	"os/signal"
	"syscall"
	"time"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/herald"
	heraldredis "github.com/zoobz-io/herald/redis"
	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/internal/boot"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/internal/ingest"
	"github.com/zoobz-io/argus/internal/shutdown"
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
	log.Println("starting worker...")

	// Signal context: cancelled on SIGTERM/SIGINT. Stops accepting new work.
	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

	// Initialize sum service and registry.
	_ = sum.New()
	k := sum.Start()

	// =========================================================================
	// 1. Load Configuration
	// =========================================================================

	if err := sum.Config[config.Database](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}
	if err := sum.Config[config.Storage](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load storage config: %w", err)
	}
	if err := sum.Config[config.Redis](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load redis config: %w", err)
	}
	if err := sum.Config[config.OpenSearch](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load opensearch config: %w", err)
	}
	if err := sum.Config[config.OCR](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load ocr config: %w", err)
	}
	if err := sum.Config[config.Convert](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load convert config: %w", err)
	}
	if err := sum.Config[config.Classify](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load classify config: %w", err)
	}
	if err := sum.Config[config.LLM](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load llm config: %w", err)
	}
	if err := sum.Config[config.Embedding](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load embedding config: %w", err)
	}
	if err := sum.Config[config.OTEL](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load otel config: %w", err)
	}
	if err := sum.Config[config.Worker](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load worker config: %w", err)
	}

	// =========================================================================
	// 2. Connect to Infrastructure
	// =========================================================================

	db, err := boot.Database(sigCtx)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	bucketProvider, err := boot.Storage(sigCtx)
	if err != nil {
		return err
	}

	redisClient, err := boot.Redis(sigCtx)
	if err != nil {
		return err
	}
	defer func() { _ = redisClient.Close() }()

	searchProvider, err := boot.OpenSearch(sigCtx)
	if err != nil {
		return err
	}

	ocrConn, ocrClient, err := boot.OCR(sigCtx)
	if err != nil {
		return err
	}
	defer func() { _ = ocrConn.Close() }()

	convertConn, convertClient, err := boot.Convert(sigCtx)
	if err != nil {
		return err
	}
	defer func() { _ = convertConn.Close() }()

	classifyConn, classifyClient, err := boot.Classify(sigCtx)
	if err != nil {
		return err
	}
	defer func() { _ = classifyConn.Close() }()

	analyzer, err := boot.LLM(sigCtx)
	if err != nil {
		return err
	}

	embedService, err := boot.Embedding(sigCtx)
	if err != nil {
		return err
	}

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
	// 5. Initialize Observability (OTEL + Aperture)
	// =========================================================================

	otelProviders, err := boot.OTEL(sigCtx, "argus-worker")
	if err != nil {
		return err
	}
	defer func() { _ = otelProviders.Shutdown(sigCtx) }()

	ap, err := boot.Aperture(sigCtx, otelProviders)
	if err != nil {
		return err
	}
	defer ap.Close()
	_ = ap

	// =========================================================================
	// 6. Signal Bridge: pipeline signals → JobStatusSignal
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
	// 7. Herald Publisher: JobStatusSignal → argus:job-status stream
	// =========================================================================

	jobStatusStream := heraldredis.NewPubSub("argus:job-status", heraldredis.WithPubSubClient(redisClient))
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
	// 8. Notification Hooks + Herald Publisher
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

	// Audit Publisher: AuditSignal → argus:audit stream
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
	// 9. Herald Subscriber: argus:ingestion → run pipeline
	// =========================================================================

	workerCfg := sum.MustUse[config.Worker](sigCtx)
	hostname := boot.Hostname()

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
	ingestSub.Start(sigCtx)
	defer func() { _ = ingestSub.Close() }()
	log.Println("ingestion queue subscriber started")

	// WorkerPool: bounded pipeline concurrency.
	// Work context: independent of signal — allows in-flight jobs to finish.
	workCtx, workCancel := context.WithCancel(context.Background())
	defer workCancel()

	var drainer shutdown.Drainer

	processor := pipz.Apply(
		pipz.NewIdentity("ingest-worker", "Process ingestion job"),
		func(ctx context.Context, job ingestJob) (ingestJob, error) {
			if err := pipeline.Ingest(ctx, job.JobID, job.VersionID); err != nil {
				log.Printf("pipeline error for job %s: %v", job.JobID, err)
				// Job failure is logged, not fatal to the pool.
			}
			return job, nil
		},
	)

	pool := pipz.NewWorkerPool(
		pipz.NewIdentity("ingest-pool", "Bounded pipeline concurrency"),
		workerCfg.WorkerCount,
		processor,
	)
	defer func() { _ = pool.Close() }()

	capitan.Hook(events.IngestQueueSignal, func(_ context.Context, e *capitan.Event) {
		msg, ok := events.IngestQueueKey.From(e)
		if !ok {
			return
		}
		log.Printf("processing job %s (version %s)", msg.JobID, msg.VersionID)
		done := drainer.Track(msg.JobID)
		go func() {
			defer done()
			_, _ = pool.Process(workCtx, ingestJob{JobID: msg.JobID, VersionID: msg.VersionID})
		}()
	})

	// =========================================================================
	// 10. Block Until Shutdown
	// =========================================================================

	log.Println("worker ready")
	<-sigCtx.Done()
	log.Println("shutting down — draining in-flight jobs...")

	// Phase 1: Drain in-flight work.
	interrupted := drainer.Drain(workerCfg.DrainTimeout)

	// Mark interrupted jobs as failed.
	if len(interrupted) > 0 {
		markCtx, markCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer markCancel()
		shutdownErr := "interrupted by shutdown"
		for _, jobID := range interrupted {
			if err := allStores.Jobs.UpdateJobStatus(markCtx, jobID, models.JobFailed, &shutdownErr); err != nil {
				log.Printf("shutdown: failed to mark job %s as failed: %v", jobID, err)
			}
		}
	}

	// Phase 2: Cancel work context (stops any lingering operations).
	workCancel()

	// Phase 3: Remove consumer from Redis group.
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanupCancel()
	shutdown.RemoveConsumer(cleanupCtx, redisClient, "argus:ingestion", workerCfg.ConsumerGroup, hostname)

	log.Println("worker stopped")
	return nil
}

// ingestJob carries job and version identifiers through the worker pool.
type ingestJob struct {
	JobID     string
	VersionID string
}

// Clone implements pipz.Cloner for parallel dispatch.
func (j ingestJob) Clone() ingestJob { return j }
