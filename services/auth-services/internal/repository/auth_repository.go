package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

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

	errGet := repo.rds.Set(ctx, "blacklist:"+token, "revoked", expiration).Err()

	return errGet

}

func (repo *AuthRepository) IsTokenBlacklisted(ctx context.Context, token string) bool {

	result, _ := repo.rds.Get(ctx, "blacklist:"+token).Result()

	return result == "revoked"

}
