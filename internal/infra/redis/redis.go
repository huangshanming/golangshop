package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"goshop/internal/config"
)

func NewClient(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}

func Ping(client *redis.Client) error {
	return client.Ping(context.Background()).Err()
}
