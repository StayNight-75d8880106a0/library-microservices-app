package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const prefixBlacklistKey = "auth:blacklist:"

type AuthRepositoryInterface interface {
	BlacklistToken(ctx context.Context, token string, expiration time.Duration) error
	IsTokenBlacklisted(ctx context.Context, token string) bool
}

type AuthRepository struct {
	rds *redis.Client
}

func NewAuthRepositoryRegistry(redisClient *redis.Client) *AuthRepository {
	return &AuthRepository{
		rds: redisClient,
	}
}

func (repo *AuthRepository) BlacklistToken(ctx context.Context, token string, expiration time.Duration) error {

	errExists := repo.rds.Set(ctx, prefixBlacklistKey+token, "revoked", expiration).Err()

	return errExists

}

func (repo *AuthRepository) IsTokenBlacklisted(ctx context.Context, token string) bool {

	result, _ := repo.rds.Get(ctx, prefixBlacklistKey+token).Result()

	return result == "revoked"

}
