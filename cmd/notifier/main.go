// Package main is the entry point for the notification sidecar.
//
// The notifier subscribes to a single notification stream via herald, expands
// each notification into per-subscriber fan-out items via streamz, and runs
// each item through the notify pipeline (assign → index → hint).
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
	"github.com/zoobz-io/streamz"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/internal/boot"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/internal/notify"
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
	log.Println("starting notifier...")

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
	if err := sum.Config[config.Redis](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load redis config: %w", err)
	}
	if err := sum.Config[config.OpenSearch](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load opensearch config: %w", err)
	}
	if err := sum.Config[config.Notifier](sigCtx, k, nil); err != nil {
		return fmt.Errorf("failed to load notifier config: %w", err)
	}

	// =========================================================================
	// 2. Connect to Infrastructure
	// =========================================================================

	db, err := boot.Database(sigCtx)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	redisClient, err := boot.Redis(sigCtx)
	if err != nil {
		return err
	}
	defer func() { _ = redisClient.Close() }()

	searchProvider, err := boot.OpenSearch(sigCtx)
	if err != nil {
		return err
	}

	osCfg := sum.MustUse[config.OpenSearch](sigCtx)
	if err := boot.EnsureIndices(sigCtx, osCfg.Addr); err != nil {
		return fmt.Errorf("ensuring opensearch indices: %w", err)
	}

	// =========================================================================
	// 3. Create Stores and Register Internal Contracts
	// =========================================================================

	renderer := astqlpg.New()
	subStore := stores.NewSubscriptions(db, renderer)
	notifStore := stores.NewNotifications(searchProvider)
	auditStore := stores.NewAudit(searchProvider)

	deliveryStore := stores.NewDeliveries(db, renderer)
	hookStore := stores.NewHooks(db, renderer)

	sum.Register[intcontracts.NotifySubscriptions](k, subStore)
	sum.Register[intcontracts.NotifyIndexer](k, notifStore)
	sum.Register[intcontracts.AuditIndexer](k, auditStore)
	sum.Register[intcontracts.NotifyHookLoader](k, hookStore)
	sum.Register[intcontracts.NotifyDeliveryLogger](k, deliveryStore)

	// Model boundaries for cereal encryption (hooks have encrypted secrets).
	sum.NewBoundary[models.Hook](k)

	// =========================================================================
	// 4. Create Pipelines and Freeze
	// =========================================================================

	inboxPipeline := notify.New()
	webhookPipeline := notify.NewWebhookPipeline()
	sum.Freeze(k)

	// =========================================================================
	// 5. Herald Subscriber (notification stream)
	// =========================================================================

	notifCfg := sum.MustUse[config.Notifier](sigCtx)
	hostname := boot.Hostname()

	notifStream := heraldredis.New("argus:notifications",
		heraldredis.WithClient(redisClient),
		heraldredis.WithGroup(notifCfg.ConsumerGroup),
		heraldredis.WithConsumer(hostname),
	)
	notifSub := herald.NewSubscriber(
		notifStream,
		events.NotificationSignal,
		events.NotificationKey,
		[]herald.Option[models.Notification]{
			herald.WithRetry[models.Notification](3),
		},
	)
	notifSub.Start(sigCtx)
	defer func() { _ = notifSub.Close() }()
	log.Println("herald subscriber started")

	// =========================================================================
	// 6. streamz: notification → []*FanOutItem expansion
	// =========================================================================

	inputCh := make(chan streamz.Result[models.Notification])

	// Work context: independent of signal — allows in-flight notifications
	// to finish processing through the pipeline.
	workCtx, workCancel := context.WithCancel(context.Background())
	defer workCancel()

	var drainer shutdown.Drainer

	expander := streamz.NewAsyncMapper(func(ctx context.Context, notif models.Notification) ([]*notify.FanOutItem, error) {
		subs, err := subStore.FindByTenantAndEventType(ctx, notif.TenantID, string(notif.Type))
		if err != nil {
			return nil, fmt.Errorf("finding subscriptions: %w", err)
		}
		items := make([]*notify.FanOutItem, len(subs))
		for i, sub := range subs {
			n := notif.Clone()
			items[i] = &notify.FanOutItem{
				Notification: &n,
				Subscription: sub,
				EventID:      notif.ID,
			}
		}
		return items, nil
	}).WithWorkers(4).WithName("notification-expander")

	expandedCh := expander.Process(workCtx, inputCh)

	// =========================================================================
	// 7. Pipeline runner: process each FanOutItem via streamz.Tap
	// =========================================================================

	runner := streamz.NewTap(func(result streamz.Result[[]*notify.FanOutItem]) {
		if result.IsError() {
			capitan.Error(workCtx, events.NotifierFanOutError,
				events.NotifierErrorKey.Field(result.Error().Unwrap()),
			)
			return
		}
		for _, item := range result.Value() {
			var err error
			switch item.Subscription.Channel {
			case models.SubscriptionChannelWebhook:
				_, err = webhookPipeline.Process(workCtx, item)
			default:
				_, err = inboxPipeline.Process(workCtx, item)
			}
			if err != nil {
				capitan.Error(workCtx, events.NotifierFanOutError,
					events.NotifierTypeKey.Field(string(item.Notification.Type)),
					events.NotifierErrorKey.Field(err),
				)
				continue
			}
			capitan.Info(workCtx, events.NotifierFanOutCompleted,
				events.NotifierTypeKey.Field(string(item.Notification.Type)),
			)
		}
	}).WithName("pipeline-runner")
	_ = runner.Process(workCtx, expandedCh)

	// =========================================================================
	// 8. Herald hook: feed notifications into streamz input channel
	// =========================================================================

	capitan.Hook(events.NotificationSignal, func(_ context.Context, e *capitan.Event) {
		notif, ok := events.NotificationKey.From(e)
		if !ok {
			return
		}
		done := drainer.Track(notif.ID)
		go func() {
			defer done()
			select {
			case inputCh <- streamz.NewSuccess(notif):
			case <-workCtx.Done():
			}
		}()
	})

	log.Println("notification fan-out pipeline registered")

	// =========================================================================
	// 8b. Herald Subscriber: audit stream → direct index
	// =========================================================================

	auditStream := heraldredis.New("argus:audit",
		heraldredis.WithClient(redisClient),
		heraldredis.WithGroup(notifCfg.ConsumerGroup),
		heraldredis.WithConsumer(hostname),
	)
	auditSub := herald.NewSubscriber(
		auditStream,
		events.AuditSignal,
		events.AuditKey,
		[]herald.Option[models.AuditEntry]{
			herald.WithRetry[models.AuditEntry](3),
		},
	)
	auditSub.Start(sigCtx)
	defer func() { _ = auditSub.Close() }()
	log.Println("audit subscriber started")

	capitan.Hook(events.AuditSignal, func(_ context.Context, e *capitan.Event) {
		entry, ok := events.AuditKey.From(e)
		if !ok {
			return
		}
		if err := auditStore.Index(workCtx, &entry); err != nil {
			capitan.Error(workCtx, events.AuditIndexError,
				events.AuditActionKey.Field(entry.Action),
				events.AuditErrorKey.Field(err),
			)
			return
		}
		capitan.Info(workCtx, events.AuditIndexed,
			events.AuditActionKey.Field(entry.Action),
		)
	})
	log.Println("audit indexer hook registered")

	// =========================================================================
	// 9. Herald Publisher: notify hints → argus:notify-hints stream
	// =========================================================================

	hintStream := heraldredis.New("argus:notify-hints", heraldredis.WithClient(redisClient))
	hintPub := herald.NewPublisher(
		hintStream,
		events.NotifyHintSignal,
		events.NotifyHintKey,
		[]herald.Option[events.NotifyHint]{
			herald.WithRetry[events.NotifyHint](3),
			herald.WithBackoff[events.NotifyHint](3, 500*time.Millisecond),
		},
	)
	hintPub.Start()
	defer func() { _ = hintPub.Close() }()
	log.Println("notify-hints publisher initialized")

	// =========================================================================
	// 10. Block Until Shutdown
	// =========================================================================

	log.Println("notifier ready")
	<-sigCtx.Done()
	log.Println("shutting down — draining in-flight notifications...")

	// Phase 1: Drain in-flight notification processing.
	drainer.Drain(notifCfg.DrainTimeout)

	// Phase 2: Close input channel to signal streamz pipeline to drain.
	close(inputCh)

	// Phase 3: Cancel work context.
	workCancel()

	// Phase 4: Remove consumers from Redis groups.
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanupCancel()
	shutdown.RemoveConsumer(cleanupCtx, redisClient, "argus:notifications", notifCfg.ConsumerGroup, hostname)
	shutdown.RemoveConsumer(cleanupCtx, redisClient, "argus:audit", notifCfg.ConsumerGroup, hostname)

	log.Println("notifier stopped")
	return nil
}
