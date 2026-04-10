package redis_client

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/retail-core/sales-service/internal/logger"
	"go.uber.org/zap"
)

func InitRedisClient(redisAddr string) (*redis.Client, error) {
	options := &redis.Options{
		Addr: redisAddr,
	}

	client := redis.NewClient(options)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.L().Info("Successfully connected to Redis ✅", zap.String("address", redisAddr))

	return client, nil
}