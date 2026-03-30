package shutdown

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// RemoveConsumer removes a consumer from a Redis Stream consumer group.
// This prevents orphaned consumers from accumulating when instances with
// ephemeral hostnames (e.g. Kubernetes pods) terminate.
func RemoveConsumer(ctx context.Context, client *redis.Client, stream, group, consumer string) {
	n, err := client.XGroupDelConsumer(ctx, stream, group, consumer).Result()
	if err != nil {
		log.Printf("shutdown: failed to remove consumer %s from %s/%s: %v", consumer, stream, group, err)
		return
	}
	log.Printf("shutdown: removed consumer %s from %s/%s (%d pending entries cleared)", consumer, stream, group, n)
}
