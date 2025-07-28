package service

import (
	"academy/internal/service/jwt"
	"academy/internal/storage/cache"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type JWTService struct {
	jwtAuth *jwt.JWTAuthenticator
	storage *cache.JWTStorage
}

func NewJWTService(storage *cache.JWTStorage, jwtAuth *jwt.JWTAuthenticator) *JWTService {
	return &JWTService{
		storage: storage,
		jwtAuth: jwtAuth,
	}
}

func (s *JWTService) GenerateTokenPair(ctx context.Context, claims *jwt.TokenClaims) (*jwt.TokenPair, error) {
	tokenPair, err := s.jwtAuth.GenerateTokenPair(claims)
	if err != nil {
		return nil, err
	}

	err = s.storage.Save(ctx, claims.UserID, tokenPair.RefreshToken)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (s *JWTService) GetRefreshByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {

	refreshToken, err := s.storage.GetRefreshToken(ctx, userID)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *JWTService) ParseAccessToken(authToken string) (*jwt.TokenClaims, error) {
	claims, err := s.jwtAuth.ParseAccessToken(authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	return claims, nil
}

func (s *JWTService) ParseRefreshToken(authToken string) (*jwt.TokenClaims, error) {
	claims, err := s.jwtAuth.ParseRefreshToken(authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	return claims, nil
}
