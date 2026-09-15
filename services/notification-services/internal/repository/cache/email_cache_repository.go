package cache

import (
	"context"
	"encoding/json"
	"log"
	"notification-services/internal/config"
	"notification-services/internal/models"
	"notification-services/internal/repository"
	"time"

	"github.com/redis/go-redis/v9"
)

type EmailCacheRepository struct {
	base repository.EmailRepositoryInterface
	rds  *redis.Client
	cfg  *config.RedisConfig
}

func NewEmailCacheRepository(base repository.EmailRepositoryInterface, rds *redis.Client, cfg *config.RedisConfig) *EmailCacheRepository {
	return &EmailCacheRepository{
		base: base,
		rds:  rds,
		cfg:  cfg,
	}
}

func (repo *EmailCacheRepository) SaveEmailLog(ctx context.Context, emailLog *models.EmailLogs) error {
	return repo.base.SaveEmailLog(ctx, emailLog)
}

func (repo *EmailCacheRepository) GetAllEmailLogs(ctx context.Context, limit int, offset int) ([]models.EmailLogs, int64, error) {
	return repo.base.GetAllEmailLogs(ctx, limit, offset)
}

func (repo *EmailCacheRepository) GetEmailLogByID(ctx context.Context, ID string) (*models.EmailLogs, error) {

	cacheKey := "emailLog:" + ID

	cachedData, errCache := repo.rds.Get(ctx, cacheKey).Result()

	if errCache == nil {
		var emailLog models.EmailLogs

		errJson := json.Unmarshal([]byte(cachedData), &emailLog)

		if errJson != nil {
			return nil, errJson
		} else {
			return &emailLog, nil
		}
	}

	log.Println("Cache MISS or Redis Down. Fetching from DB for ID:", ID)

	emailLog, errGet := repo.base.GetEmailLogByID(ctx, ID)

	if errGet != nil {
		return nil, errGet
	}

	if repo.rds != nil {
		go func(data *models.EmailLogs, key string) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Redis Set Panic:", r)
				}
			}()
			bgContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			bytes, _ := json.Marshal(data)
			repo.rds.Set(bgContext, key, bytes, repo.cfg.RedisCacheTTL)
		}(emailLog, cacheKey)
	}

	return emailLog, nil
}
