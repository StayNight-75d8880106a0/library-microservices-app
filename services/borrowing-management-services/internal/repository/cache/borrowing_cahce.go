package cache

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/models"
	"borrowing-management-services/internal/repository"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type BorrowingUserCacheRepository struct {
	base repository.BorrowingUserRepositoryInterface
	rds  *redis.Client
	cfg  *config.RedisCacheConfig
}

func NewBorrowingUserCacheRepository(base repository.BorrowingUserRepositoryInterface, rds *redis.Client, cfg *config.RedisCacheConfig) *BorrowingUserCacheRepository {
	return &BorrowingUserCacheRepository{
		base: base,
		rds:  rds,
		cfg:  cfg,
	}
}

func (repo *BorrowingUserCacheRepository) CreateBorrowing(ctx context.Context, borrowing *models.Borrowing) error {
	return repo.base.CreateBorrowing(ctx, borrowing)
}

func (repo *BorrowingUserCacheRepository) GetAllMyBorrowings(ctx context.Context, userID string, limit int, offset int) ([]models.Borrowing, int64, error) {
	return repo.base.GetAllMyBorrowings(ctx, userID, limit, offset)
}

func (repo *BorrowingUserCacheRepository) GetAllBorrowings(ctx context.Context, limit int, offset int) ([]models.Borrowing, int64, error) {
	return repo.base.GetAllBorrowings(ctx, limit, offset)
}

func (repo *BorrowingUserCacheRepository) GetBorrowingByID(ctx context.Context, ID string) (*models.Borrowing, error) {

	cacheKey := "borrowing:" + ID

	cachedData, errCache := repo.rds.Get(ctx, cacheKey).Result()

	if errCache == nil {
		var borrowing models.Borrowing

		errJson := json.Unmarshal([]byte(cachedData), &borrowing)

		if errJson != nil {
			return nil, errJson
		} else {
			return &borrowing, nil
		}
	}

	log.Println("Cache MISS or Redis Down. Fetching from DB for ID:", ID)

	borrowing, errDB := repo.base.GetBorrowingByID(ctx, ID)

	if errDB != nil {
		return nil, errDB
	}

	if repo.rds != nil {
		go func(data *models.Borrowing, key string) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Redis Set Panic:", r)
				}
			}()
			bgContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			bytes, _ := json.Marshal(data)
			repo.rds.Set(bgContext, key, bytes, repo.cfg.RedisCacheTTL)
		}(borrowing, cacheKey)
	}

	return borrowing, nil

}

func (repo *BorrowingUserCacheRepository) UpdateStatus(ctx context.Context, ID string, status models.BorrowingStatus) error {

	err := repo.base.UpdateStatus(ctx, ID, status)

	if err == nil && repo.rds != nil {
		cacheKey := "borrowing:" + ID
		repo.rds.Del(ctx, cacheKey)
	}

	return err

}
