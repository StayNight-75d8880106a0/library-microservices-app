package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserRedisCacheInterface interface {
	SetUserStatus(ctx context.Context, userID string, status string, ttl time.Duration) error
	GetUserStatus(ctx context.Context, userID string) (string, error)
	DeleteUserStatus(ctx context.Context, userID string) error
}

type UserRedisCache struct {
	rds *redis.Client
}

func NewUserRedisCache(redis *redis.Client) *UserRedisCache {
	return &UserRedisCache{
		rds: redis,
	}
}

var ErrUserStatusNotFound = errors.New("user status not found in cache")

func (r *UserRedisCache) getKey(userID string) string {
	return fmt.Sprintf("user:status:%s", userID)
}

func (r *UserRedisCache) SetUserStatus(ctx context.Context, userID string, status string, ttl time.Duration) error {

	key := r.getKey(userID)

	errRedis := r.rds.Set(ctx, key, status, ttl).Err()

	return errRedis
}

func (r *UserRedisCache) GetUserStatus(ctx context.Context, userID string) (string, error) {

	key := r.getKey(userID)

	status, errRedis := r.rds.Get(ctx, key).Result()

	if errRedis != nil {
		if errors.Is(errRedis, redis.Nil) {
			return "", ErrUserStatusNotFound
		}
		return "", errRedis
	}

	return status, nil

}

func (r *UserRedisCache) DeleteUserStatus(ctx context.Context, userID string) error {

	key := r.getKey(userID)

	errRedis := r.rds.Del(ctx, key).Err()

	return errRedis

}
