package redisdb

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrLockNotAcquired = errors.New("lock not acquired")

var releaseScript = redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	end
	return 0
`)

type LockerInterface interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (func(), error)
	AcquireWithRetry(ctx context.Context, key string, ttl time.Duration, attempts int, backoff time.Duration) (func(), error)
}

type Locker struct {
	rds *redis.Client
}

func NewLocker(rds *redis.Client) *Locker {
	return &Locker{
		rds: rds,
	}
}

func (loc *Locker) Acquire(ctx context.Context, key string, ttl time.Duration) (func(), error) {

	token := uuid.NewString()

	ok, err := loc.rds.SetNX(ctx, key, token, ttl).Result()

	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrLockNotAcquired
	}

	releaseFunc := func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := releaseScript.Run(releaseCtx, loc.rds, []string{key}, token).Err(); err != nil {
			log.Printf("[Lock] Failed to release %s: %v", key, err)
		}
	}

	return releaseFunc, nil
}

func (loc *Locker) AcquireWithRetry(ctx context.Context, key string, ttl time.Duration, attempts int, backoff time.Duration) (func(), error) {

	if backoff <= 0 {
		backoff = 51 * time.Millisecond
	}

	var lastError error

	for i := 0; i < attempts; i++ {

		release, errReleases := loc.Acquire(ctx, key, ttl)

		if errReleases == nil {
			return release, nil
		}

		if !errors.Is(errReleases, ErrLockNotAcquired) {
			return nil, errReleases
		}

		lastError = errReleases

		if i == attempts-1 {
			break
		}

		jitter := time.Duration(rand.Int63n(int64(backoff)))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff + jitter):
		}

	}

	return nil, lastError

}
