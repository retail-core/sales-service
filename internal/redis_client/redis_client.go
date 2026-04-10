package redis_client

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type QueueStore interface {
	GetNextQueueNumber(ctx context.Context, inventoryID string) (int64, error)
}

type RedisQueueStore struct {
	client *redis.Client
}

func NewRedisQueueStore(client *redis.Client) *RedisQueueStore {
	return &RedisQueueStore{client: client}
}

func (r *RedisQueueStore) GetNextQueueNumber(ctx context.Context, inventoryID string) (int64, error) {
	now := time.Now()
	key := fmt.Sprintf("queue:%s:%s", inventoryID, now.Format("2006-01-02"))

	queueNumber, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// set expiry only if first time
	if queueNumber == 1 {
		r.client.Expire(ctx, key, getTTLUntilMidnight())
	}

	return queueNumber, nil
}

func getTTLUntilMidnight() time.Duration {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)

	midnight := time.Date(
		tomorrow.Year(),
		tomorrow.Month(),
		tomorrow.Day(),
		0, 0, 0, 0,
		now.Location(),
	)

	return time.Until(midnight)
}