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
	if err := sum.Config[config.Redis](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load redis config: %w", err)
	}
	if err := sum.Config[config.OpenSearch](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load opensearch config: %w", err)
	}
	if err := sum.Config[config.Notifier](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load notifier config: %w", err)
	}

	// =========================================================================
	// 2. Connect to Infrastructure
	// =========================================================================

	db, err := boot.Database(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	redisClient, err := boot.Redis(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = redisClient.Close() }()

	searchProvider, err := boot.OpenSearch(ctx)
	if err != nil {
		return err
	}

	// =========================================================================
	// 3. Create Stores and Register Internal Contracts
	// =========================================================================

	renderer := astqlpg.New()
	subStore := stores.NewSubscriptions(db, renderer)
	notifStore := stores.NewNotifications(searchProvider)
	auditStore := stores.NewAudit(searchProvider)

	sum.Register[intcontracts.NotifySubscriptions](k, subStore)
	sum.Register[intcontracts.NotifyIndexer](k, notifStore)
	sum.Register[intcontracts.AuditIndexer](k, auditStore)

	// =========================================================================
	// 4. Create Pipeline and Freeze
	// =========================================================================

	pipeline := notify.New()
	sum.Freeze(k)

	// =========================================================================
	// 5. Herald Subscriber (notification stream)
	// =========================================================================

	notifCfg := sum.MustUse[config.Notifier](ctx)
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
	notifSub.Start(ctx)
	defer func() { _ = notifSub.Close() }()
	log.Println("herald subscriber started")

	// =========================================================================
	// 6. streamz: notification → []*FanOutItem expansion
	// =========================================================================

	inputCh := make(chan streamz.Result[models.Notification])

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

	expandedCh := expander.Process(ctx, inputCh)

	// =========================================================================
	// 7. Pipeline runner: process each FanOutItem via streamz.Tap
	// =========================================================================

	runner := streamz.NewTap(func(result streamz.Result[[]*notify.FanOutItem]) {
		if result.IsError() {
			capitan.Error(ctx, events.NotifierFanOutError,
				events.NotifierErrorKey.Field(result.Error().Unwrap()),
			)
			return
		}
		for _, item := range result.Value() {
			if _, err := pipeline.Process(ctx, item); err != nil {
				capitan.Error(ctx, events.NotifierFanOutError,
					events.NotifierTypeKey.Field(string(item.Notification.Type)),
					events.NotifierErrorKey.Field(err),
				)
				continue
			}
			capitan.Info(ctx, events.NotifierFanOutCompleted,
				events.NotifierTypeKey.Field(string(item.Notification.Type)),
			)
		}
	}).WithName("pipeline-runner")
	_ = runner.Process(ctx, expandedCh)

	// =========================================================================
	// 8. Herald hook: feed notifications into streamz input channel
	// =========================================================================

	capitan.Hook(events.NotificationSignal, func(ctx context.Context, e *capitan.Event) {
		notif, ok := events.NotificationKey.From(e)
		if !ok {
			return
		}
		select {
		case inputCh <- streamz.NewSuccess(notif):
		case <-ctx.Done():
		}
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
	auditSub.Start(ctx)
	defer func() { _ = auditSub.Close() }()
	log.Println("audit subscriber started")

	capitan.Hook(events.AuditSignal, func(ctx context.Context, e *capitan.Event) {
		entry, ok := events.AuditKey.From(e)
		if !ok {
			return
		}
		if err := auditStore.Index(ctx, &entry); err != nil {
			capitan.Error(ctx, events.AuditIndexError,
				events.AuditActionKey.Field(entry.Action),
				events.AuditErrorKey.Field(err),
			)
			return
		}
		capitan.Info(ctx, events.AuditIndexed,
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
	<-ctx.Done()
	log.Println("shutting down...")
	return nil
}
