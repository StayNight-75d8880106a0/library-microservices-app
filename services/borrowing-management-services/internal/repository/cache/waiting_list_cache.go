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

type WaitingListCacheRepository struct {
	base repository.WaitingListRepositoryInterface
	rds  *redis.Client
	cfg  *config.RedisCacheConfig
}

func NewWaitingListCacheRepository(base repository.WaitingListRepositoryInterface, rds *redis.Client, cfg *config.RedisCacheConfig) *WaitingListCacheRepository {
	return &WaitingListCacheRepository{
		base: base,
		rds:  rds,
		cfg:  cfg,
	}
}

func (repo *WaitingListCacheRepository) Create(ctx context.Context, data *models.WaitingList) error {
	return repo.base.Create(ctx, data)
}

func (repo *WaitingListCacheRepository) GetALLWaitingList(ctx context.Context, limit int, offset int) ([]models.WaitingList, int64, error) {
	return repo.base.GetALLWaitingList(ctx, limit, offset)
}

func (repo *WaitingListCacheRepository) GetALLMyWaitingList(ctx context.Context, limit int, offset int, userID string) ([]models.WaitingList, int64, error) {
	return repo.base.GetALLMyWaitingList(ctx, limit, offset, userID)
}

func (repo *WaitingListCacheRepository) GetWaitingListByID(ctx context.Context, ID string) (*models.WaitingList, error) {

	cacheKey := "waitinglist:" + ID

	cachedData, errCache := repo.rds.Get(ctx, cacheKey).Result()

	if errCache == nil {
		var waitingList models.WaitingList

		errJson := json.Unmarshal([]byte(cachedData), &waitingList)

		if errJson != nil {
			return nil, errJson
		} else {
			return &waitingList, nil
		}
	}

	log.Println("Cache MISS or Redis Down. Fetching from DB for ID:", ID)

	waitingList, errDB := repo.base.GetWaitingListByID(ctx, ID)

	if errDB != nil {
		return nil, errDB
	}

	if repo.rds != nil {
		go func(data *models.WaitingList, key string) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Redis Set Panic:", r)
				}
			}()
			bgContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			bytes, _ := json.Marshal(data)
			repo.rds.Set(bgContext, key, bytes, repo.cfg.RedisCacheTTL)
		}(waitingList, cacheKey)
	}

	return waitingList, nil
}

func (repo *WaitingListCacheRepository) UpdateStatusWaitingList(ctx context.Context, ID string, from models.WaitingListStatus, to models.WaitingListStatus) (int64, error) {

	rowsAffected, err := repo.base.UpdateStatusWaitingList(ctx, ID, from, to)

	if err == nil && repo.rds != nil {
		cacheKey := "waitinglist:" + ID
		repo.rds.Del(ctx, cacheKey)
	}

	return rowsAffected, err

}

// func (repo *WaitingListCacheRepository) GetLastQueueNumber(ctx context.Context, bookID string) (int, error) {
// 	return repo.base.GetLastQueueNumber(ctx, bookID)
// }

func (repo *WaitingListCacheRepository) GetFirstWaitingListByBookID(ctx context.Context, bookID string) (*models.WaitingList, error) {
	return repo.base.GetFirstWaitingListByBookID(ctx, bookID)
}

func (repo *WaitingListCacheRepository) CancelWaitingListByUser(ctx context.Context, ID string, userID string) (int64, error) {
	rowsAffected, err := repo.base.CancelWaitingListByUser(ctx, ID, userID)

	if err == nil && repo.rds != nil {
		cacheKey := "waitinglist:" + ID
		repo.rds.Del(ctx, cacheKey)
	}

	return rowsAffected, err
}

func (repo *WaitingListCacheRepository) GetExpiredWaitingLists(ctx context.Context, expiredHours int) ([]models.WaitingList, error) {
	return repo.base.GetExpiredWaitingLists(ctx, expiredHours)
}

func (repo *WaitingListCacheRepository) MarkFulfilledWaitingList(ctx context.Context, userID string, bookID string) (int64, error) {
	return repo.base.MarkFulfilledWaitingList(ctx, userID, bookID)
}

func (repo *WaitingListCacheRepository) RevertFulfilledWaitingList(ctx context.Context, userID string, bookID string) error {
	return repo.base.RevertFulfilledWaitingList(ctx, userID, bookID)
}

func (repo *WaitingListCacheRepository) CountNotifiedWaitingListsByBookID(ctx context.Context, bookID string) (int64, error) {
	return repo.base.CountNotifiedWaitingListsByBookID(ctx, bookID)
}
