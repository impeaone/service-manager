package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"time"
)

type OTPRepository struct {
	client *redis.Client
}

func NewOTPRepository(client *redis.Client) *OTPRepository {
	return &OTPRepository{client: client}
}

func (r *OTPRepository) GenerateOTP(ctx context.Context, email, purpose string, ttl time.Duration) (string, error) {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	key := fmt.Sprintf("otp:%s:%s", email, purpose)

	err := r.client.Set(ctx, key, code, ttl).Err()
	if err != nil {
		return "", fmt.Errorf("failed to save OTP: %w", err)
	}

	return code, nil
}

func (r *OTPRepository) VerifyOTP(ctx context.Context, email, purpose, code string) (bool, error) {
	key := fmt.Sprintf("otp:%s:%s", email, purpose)

	savedCode, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get OTP: %w", err)
	}

	if savedCode != code {
		return false, nil
	}

	r.client.Del(ctx, key)

	return true, nil
}

func (r *OTPRepository) DeleteOTP(ctx context.Context, email, purpose string) error {
	key := fmt.Sprintf("otp:%s:%s", email, purpose)
	return r.client.Del(ctx, key).Err()
}
