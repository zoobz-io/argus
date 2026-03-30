// Package main is the entry point for the notification sidecar.
//
// The notifier subscribes to a single domain events stream via herald and
// routes each event to: (1) the domain_events audit index (always),
// (2) notification fan-out via streamz if subscriptions match, and
// (3) webhook delivery for webhook subscriptions.
package main

import (
	"context"
	"encoding/json"
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
	domainEventsStore := stores.NewDomainEvents(searchProvider)

	deliveryStore := stores.NewDeliveries(db, renderer)
	hookStore := stores.NewHooks(db, renderer)

	sum.Register[intcontracts.NotifySubscriptions](k, subStore)
	sum.Register[intcontracts.NotifyIndexer](k, notifStore)
	sum.Register[intcontracts.DomainEventIndexer](k, domainEventsStore)
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
	// 5. Herald Subscriber (unified domain events stream)
	// =========================================================================

	notifCfg := sum.MustUse[config.Notifier](sigCtx)
	hostname := boot.Hostname()

	eventsStream := heraldredis.New("argus:events",
		heraldredis.WithClient(redisClient),
		heraldredis.WithGroup(notifCfg.ConsumerGroup),
		heraldredis.WithConsumer(hostname),
	)
	eventsSub := herald.NewSubscriber(
		eventsStream,
		events.DomainEventSignal,
		events.DomainEventKey,
		[]herald.Option[models.DomainEvent]{
			herald.WithRetry[models.DomainEvent](3),
		},
	)
	eventsSub.Start(sigCtx)
	defer func() { _ = eventsSub.Close() }()
	log.Println("domain events subscriber started")

	// =========================================================================
	// 6. streamz: DomainEvent → []*FanOutItem expansion
	// =========================================================================

	inputCh := make(chan streamz.Result[models.DomainEvent])

	// Work context: independent of signal — allows in-flight processing to finish.
	workCtx, workCancel := context.WithCancel(context.Background())
	defer workCancel()

	var drainer shutdown.Drainer

	expander := streamz.NewAsyncMapper(func(ctx context.Context, evt models.DomainEvent) ([]*notify.FanOutItem, error) {
		// Always index the domain event (audit log).
		if err := domainEventsStore.Index(ctx, &evt); err != nil {
			capitan.Error(ctx, events.DomainEventIndexError,
				events.DomainEventActionKey.Field(evt.Action),
				events.DomainEventErrorKey.Field(err),
			)
			// Continue — indexing failure should not prevent notification delivery.
		} else {
			capitan.Info(ctx, events.DomainEventIndexed,
				events.DomainEventActionKey.Field(evt.Action),
			)
		}

		// Fan out to matching subscriptions.
		subs, err := subStore.FindByTenantAndEventType(ctx, evt.TenantID, evt.Action)
		if err != nil {
			return nil, fmt.Errorf("finding subscriptions: %w", err)
		}
		if len(subs) == 0 {
			return nil, nil
		}

		items := make([]*notify.FanOutItem, len(subs))
		for i, sub := range subs {
			n := materializeNotification(evt)
			e := evt.Clone()
			items[i] = &notify.FanOutItem{
				Notification: &n,
				Subscription: sub,
				DomainEvent:  &e,
				EventID:      evt.ID,
			}
		}
		return items, nil
	}).WithWorkers(notifCfg.FanOutWorkers).WithName("domain-event-expander")

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
			// Always run inbox pipeline first (assign + index + hint).
			// This ensures Notification.ID, UserID, Status are set for all items.
			processed, err := inboxPipeline.Process(workCtx, item)
			if err != nil {
				capitan.Error(workCtx, events.NotifierFanOutError,
					events.NotifierTypeKey.Field(string(item.Notification.Type)),
					events.NotifierErrorKey.Field(err),
				)
				continue
			}
			// For webhook subscriptions, additionally deliver via webhook.
			if processed.Subscription.Channel == models.SubscriptionChannelWebhook {
				if _, err := webhookPipeline.Process(workCtx, processed); err != nil {
					capitan.Error(workCtx, events.NotifierFanOutError,
						events.NotifierTypeKey.Field(string(item.Notification.Type)),
						events.NotifierErrorKey.Field(err),
					)
					continue
				}
			}
			capitan.Info(workCtx, events.NotifierFanOutCompleted,
				events.NotifierTypeKey.Field(string(item.Notification.Type)),
			)
		}
	}).WithName("pipeline-runner")
	_ = runner.Process(workCtx, expandedCh)

	// =========================================================================
	// 8. Herald hook: feed domain events into streamz input channel
	// =========================================================================

	capitan.Hook(events.DomainEventSignal, func(_ context.Context, e *capitan.Event) {
		evt, ok := events.DomainEventKey.From(e)
		if !ok {
			return
		}
		done := drainer.Track(evt.ID)
		go func() {
			defer done()
			select {
			case inputCh <- streamz.NewSuccess(evt):
			case <-workCtx.Done():
			}
		}()
	})

	log.Println("domain event routing pipeline registered")

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
	log.Println("shutting down — draining in-flight events...")

	// Phase 1: Drain in-flight event processing.
	drainer.Drain(notifCfg.DrainTimeout)

	// Phase 2: Close input channel to signal streamz pipeline to drain.
	close(inputCh)

	// Phase 3: Cancel work context.
	workCancel()

	// Phase 4: Remove consumer from Redis group.
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanupCancel()
	shutdown.RemoveConsumer(cleanupCtx, redisClient, "argus:events", notifCfg.ConsumerGroup, hostname)

	log.Println("notifier stopped")
	return nil
}

// materializeNotification builds a per-user Notification from a DomainEvent.
// Convenience fields (DocumentID, VersionID, Message, Error) are extracted from
// metadata for events that carry them; Metadata is passed through for the
// discriminated union on the wire type.
func materializeNotification(evt models.DomainEvent) models.Notification {
	n := models.Notification{
		TenantID: evt.TenantID,
		Type:     models.NotificationType(evt.Action),
		Message:  evt.Message,
		Metadata: append(json.RawMessage(nil), evt.Metadata...),
	}

	// Extract convenience fields from metadata if present.
	if evt.Metadata != nil {
		var m struct {
			DocumentID string `json:"document_id"`
			VersionID  string `json:"version_id"`
			Error      string `json:"error"`
		}
		if err := json.Unmarshal(evt.Metadata, &m); err == nil {
			n.DocumentID = m.DocumentID
			n.VersionID = m.VersionID
			n.Error = m.Error
		}
	}
	return n
}

