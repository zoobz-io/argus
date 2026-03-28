// Package main is the entry point for the notification sidecar.
//
// The notifier subscribes to a single notification stream via herald and indexes
// notifications into OpenSearch for per-tenant notification feeds.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v4"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zoobz-io/capitan"
	grubopensearch "github.com/zoobz-io/grub/opensearch"
	"github.com/zoobz-io/herald"
	heraldredis "github.com/zoobz-io/herald/redis"
	osrenderer "github.com/zoobz-io/lucene/opensearch"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/stores"
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

	// Redis
	redisCfg := sum.MustUse[config.Redis](ctx)
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: redisCfg.Addr,
	})
	defer func() { _ = redisClient.Close() }()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
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

	// =========================================================================
	// 3. Create Notification Store
	// =========================================================================

	notifStore := stores.NewNotifications(searchProvider)

	// =========================================================================
	// 4. Freeze Registry
	// =========================================================================

	sum.Freeze(k)

	// =========================================================================
	// 5. Herald Subscriber (single notification stream)
	// =========================================================================

	notifCfg := sum.MustUse[config.Notifier](ctx)
	hostname, _ := os.Hostname()

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
	// 6. Notification Indexer
	// =========================================================================

	capitan.Hook(events.NotificationSignal, func(ctx context.Context, e *capitan.Event) {
		notif, ok := events.NotificationKey.From(e)
		if !ok {
			return
		}
		notif.ID = uuid.NewString()
		notif.CreatedAt = time.Now().UTC()
		notif.Status = models.NotificationUnread

		if err := notifStore.Index(ctx, &notif); err != nil {
			capitan.Error(ctx, events.NotifierIndexError,
				events.NotifierTypeKey.Field(string(notif.Type)),
				events.NotifierErrorKey.Field(err),
			)
			return
		}
		capitan.Info(ctx, events.NotifierIndexed,
			events.NotifierTypeKey.Field(string(notif.Type)),
		)
	})

	log.Println("notification indexer registered")

	// =========================================================================
	// 7. Block Until Shutdown
	// =========================================================================

	log.Println("notifier ready")
	<-ctx.Done()
	log.Println("shutting down...")
	return nil
}
