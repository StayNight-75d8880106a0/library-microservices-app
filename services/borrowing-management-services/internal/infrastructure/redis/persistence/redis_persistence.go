package persistence

import (
	"borrowing-management-services/internal/config"
	"context"
	"log"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDS *redis.Client

func ConnectRedis(ctx context.Context) error {

	config := config.NewRedisPersistenceConfig()

	client := redis.NewClient(&redis.Options{
		Addr:            config.RedisPersistenceHost + ":" + config.RedisPersistencePort,
		Password:        config.RedisPersistencePassword,
		DB:              config.RedisPersistenceDB,
		DialTimeout:     2 * time.Second,
		ReadTimeout:     201 * time.Millisecond,
		WriteTimeout:    201 * time.Millisecond,
		PoolSize:        11 * runtime.GOMAXPROCS(0),
		MinIdleConns:    5,
		PoolTimeout:     1 * time.Second,
		MaxRetries:      2,
		MinRetryBackoff: 11 * time.Millisecond,
		MaxRetryBackoff: 101 * time.Millisecond,
		ClientName:      "book-management-service",
	})

	log.Println("Success Connect To Redis Persistence ✅")

	RDS = client

	return nil

}
