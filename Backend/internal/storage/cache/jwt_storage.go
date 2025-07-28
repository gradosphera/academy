package cache

import (
	"academy/internal/config"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type JWTStorage struct {
	client          *redis.Client
	conf            *config.Config
	refreshTokenTTL time.Duration
}

func NewJWTCacheStorage(
	client *redis.Client,
	cfg *config.Config,
) *JWTStorage {

	return &JWTStorage{
		client:          client,
		conf:            cfg,
		refreshTokenTTL: cfg.Auth.RefreshTokenTTL,
	}
}

func (s *JWTStorage) Save(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	err := s.client.Set(ctx, userID.String(), refreshToken, s.refreshTokenTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to save user token: %w", err)
	}

	return nil
}

func (s *JWTStorage) GetRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	result, err := s.client.Get(ctx, userID.String()).Result()
	if err != nil {
		return "", fmt.Errorf("failed to find token by user_id: %w", err)
	}

	if result == "" {
		return "", fmt.Errorf("token not found")
	}

	return result, nil
}
