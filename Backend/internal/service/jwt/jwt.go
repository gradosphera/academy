package jwt

import (
	"academy/internal/config"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	MiniAppID uuid.UUID `json:"mini_app_id"`
	IsOwner   bool      `json:"is_owner"`
	IsMod     bool      `json:"is_mod"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenJWTClaims struct {
	jwt.RegisteredClaims

	TokenClaims
}

type JWTAuthenticator struct {
	refreshSignKey  []byte
	accessSignKey   []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTAuth(cfg *config.Config) *JWTAuthenticator {
	return &JWTAuthenticator{
		accessSignKey:   []byte(cfg.Auth.AccessTokenSignKey),
		refreshSignKey:  []byte(cfg.Auth.RefreshTokenSignKey),
		accessTokenTTL:  cfg.Auth.AccessTokenTTL,
		refreshTokenTTL: cfg.Auth.RefreshTokenTTL,
	}
}

func (a *JWTAuthenticator) GenerateTokenPair(claims *TokenClaims) (*TokenPair, error) {
	refreshToken, err := a.generateToken(claims, a.refreshSignKey, a.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	accessToken, err := a.generateToken(claims, a.accessSignKey, a.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (a *JWTAuthenticator) ParseAccessToken(authToken string) (*TokenClaims, error) {
	return a.parseToken(authToken, a.getAccessKey)
}

func (a *JWTAuthenticator) ParseRefreshToken(authToken string) (*TokenClaims, error) {
	return a.parseToken(authToken, a.getRefreshKey)
}

func (a *JWTAuthenticator) parseToken(
	authToken string,
	keyFunc jwt.Keyfunc,
) (*TokenClaims, error) {

	authToken = strings.TrimPrefix(authToken, "Bearer ")

	token, err := jwt.Parse(
		authToken,
		keyFunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) || errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token is not valid: %w", err)
		}

		return nil, fmt.Errorf("failed to parse jwt token: %w", err)
	}

	if token == nil || !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	parsed, ok := extractTokenClaims(claims)
	if !ok {
		return nil, fmt.Errorf("invalid claim")
	}

	return parsed, nil
}

func (a *JWTAuthenticator) getAccessKey(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method")
	}

	return []byte(a.accessSignKey), nil
}

func (a *JWTAuthenticator) getRefreshKey(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method")
	}

	return []byte(a.refreshSignKey), nil
}

func (a *JWTAuthenticator) generateToken(
	claims *TokenClaims,
	signKey []byte,
	tokenLifetime time.Duration,
) (string, error) {

	if claims.UserID == uuid.Nil {
		return "", fmt.Errorf("invalid user_id claim")
	}

	now := time.Now().UTC()
	jwtClaims := TokenJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "academy-api",
			Subject:   "client",
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
		TokenClaims: *claims,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	signedToken, err := token.SignedString(signKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return signedToken, nil
}

func extractTokenClaims(claims jwt.MapClaims) (*TokenClaims, bool) {
	v, ok := claims["user_id"]
	if !ok {
		return nil, false
	}
	stringUUID, ok := v.(string)
	if !ok {
		return nil, false
	}
	userID, err := uuid.Parse(stringUUID)
	if err != nil {
		return nil, false
	}

	v, ok = claims["mini_app_id"]
	if !ok {
		return nil, false
	}
	stringUUID, ok = v.(string)
	if !ok {
		return nil, false
	}
	miniAppID, err := uuid.Parse(stringUUID)
	if err != nil {
		return nil, false
	}

	v, ok = claims["is_owner"]
	if !ok {
		return nil, false
	}
	isOwner, ok := v.(bool)
	if !ok {
		return nil, false
	}

	v, ok = claims["is_mod"]
	if !ok {
		return nil, false
	}
	isMod, ok := v.(bool)
	if !ok {
		return nil, false
	}

	return &TokenClaims{
		UserID:    userID,
		MiniAppID: miniAppID,
		IsOwner:   isOwner,
		IsMod:     isMod,
	}, true
}
