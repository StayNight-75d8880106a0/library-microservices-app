package repository

import (
	"activity-logbook-services/internal/config"
	"activity-logbook-services/internal/models"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type LogbookCacheRepository struct {
	base LogbookRepositoryInterface
	rds  *redis.Client
	cfg  *config.RedisConfig
}

func NewLogbookCacheRepository(base LogbookRepositoryInterface, rds *redis.Client, cfg *config.RedisConfig) *LogbookCacheRepository {
	return &LogbookCacheRepository{
		base: base,
		rds:  rds,
		cfg:  cfg,
	}
}

func (repo *LogbookCacheRepository) Create(ctx context.Context, logbook *models.Logbook) error {
	return repo.base.Create(ctx, logbook)
}

func (repo *LogbookCacheRepository) GetAll(ctx context.Context, limit int, cursorOccurredAt *time.Time, cursorID *string) ([]models.Logbook, error) {
	return repo.base.GetAll(ctx, limit, cursorOccurredAt, cursorID)
}

func (repo *LogbookCacheRepository) GetByTraceID(ctx context.Context, traceID string) ([]models.Logbook, error) {

	cacheKey := "logbook:trace:" + traceID

	cachedData, errCache := repo.rds.Get(ctx, cacheKey).Result()

	if errCache == nil {
		var logbooks []models.Logbook

		errJson := json.Unmarshal([]byte(cachedData), &logbooks)

		if errJson != nil {
			return nil, errJson
		} else {
			return logbooks, nil
		}
	}

	log.Println("Cache MISS or Redis Down. Fetching from DB for Trace ID:", traceID)

	logbook, errDB := repo.base.GetByTraceID(ctx, traceID)

	if errDB != nil {
		return nil, errDB
	}

	if repo.rds != nil {
		go func(data []models.Logbook, key string) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Redis Set Panic:", r)
				}
			}()
			bgContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			bytes, _ := json.Marshal(data)
			repo.rds.Set(bgContext, key, bytes, repo.cfg.RedisCacheTTL)
		}(logbook, cacheKey)
	}

	return logbook, nil

}

func (repo *LogbookCacheRepository) GetByID(ctx context.Context, ID string, traceID string) (*models.Logbook, error) {

	cacheKey := "logbook:trace:" + traceID + ":id:" + ID

	cachedData, errCache := repo.rds.Get(ctx, cacheKey).Result()

	if errCache == nil {
		var logbook models.Logbook

		errJson := json.Unmarshal([]byte(cachedData), &logbook)

		if errJson != nil {
			return nil, errJson
		} else {
			return &logbook, nil
		}
	}

	log.Println("Cache MISS or Redis Down. Fetching from DB for ID:", ID)

	logbook, errDB := repo.base.GetByID(ctx, ID, traceID)

	if errDB != nil {
		return nil, errDB
	}

	if repo.rds != nil {
		go func(data *models.Logbook, key string) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Redis Set Panic:", r)
				}
			}()
			bgContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			bytes, _ := json.Marshal(data)
			repo.rds.Set(bgContext, key, bytes, repo.cfg.RedisCacheTTL)
		}(logbook, cacheKey)
	}

	return logbook, nil

}
