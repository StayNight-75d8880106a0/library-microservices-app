package cache

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

	config := config.NewRedisCacheConfig()

	client := redis.NewClient(&redis.Options{
		Addr:            config.RedisHost + ":" + config.RedisPort,
		Password:        config.RedisPassword,
		DB:              config.RedisDB,
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

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)

	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		log.Println("⚠️  Redis is unavailable — the application is running WITHOUT caching : " + err.Error())
		return err
	}

	log.Println("Success Connect To Redis Cache ✅")

	RDS = client

	return nil

}
