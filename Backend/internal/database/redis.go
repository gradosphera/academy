package database

import (
	"academy/internal/config"
	"context"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

func CreateRedisClient(cfg *config.Config) (*redis.Client, error) {
	addr := net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return client, nil
}
