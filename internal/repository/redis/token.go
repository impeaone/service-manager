package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenRepository struct {
	client *redis.Client
}

func NewTokenRepository(client *redis.Client) *TokenRepository {
	return &TokenRepository{client: client}
}

func (r *TokenRepository) SaveRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", token)

	err := r.client.Set(ctx, key, userID, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	userKey := fmt.Sprintf("user_tokens:%s", userID)

	err = r.client.SAdd(ctx, userKey, token).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to user index: %w", err)
	}
	r.client.Expire(ctx, userKey, ttl)

	return nil
}

func (r *TokenRepository) GetUserIDByRefreshToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", token)

	userID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("refresh token not found or expired")
		}
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	return userID, nil
}

func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	userID, err := r.client.Get(ctx, key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to get token: %w", err)
	}

	if err = r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	if userID != "" {
		userKey := fmt.Sprintf("user_tokens:%s", userID)
		r.client.SRem(ctx, userKey, token)
	}

	return nil
}

func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	userKey := fmt.Sprintf("user_tokens:%s", userID)

	tokens, err := r.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	for _, token := range tokens {
		tokenKey := fmt.Sprintf("refresh_token:%s", token)
		r.client.Del(ctx, tokenKey)
	}

	r.client.Del(ctx, userKey)

	return nil
}
