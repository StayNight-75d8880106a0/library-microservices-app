package repository

import (
	"book-management-services/internal/config"
	"book-management-services/internal/dto"
	"book-management-services/internal/models"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type BookCacheRepository struct {
	base BookRepositoryInterface
	rds  *redis.Client
	cfg  *config.RedisConfig
}

func NewBookCacheRepository(base BookRepositoryInterface, rds *redis.Client, cfg *config.RedisConfig) *BookCacheRepository {
	return &BookCacheRepository{
		base: base,
		rds:  rds,
		cfg:  cfg,
	}
}

func (repo *BookCacheRepository) Create(ctx context.Context, book *models.Books) error {
	return repo.base.Create(ctx, book)
}

func (repo *BookCacheRepository) GetById(ctx context.Context, ID string) (*models.Books, error) {

	cacheKey := "book:" + ID

	cachedData, errCache := repo.rds.Get(ctx, cacheKey).Result()

	if errCache == nil {
		var book models.Books

		errJson := json.Unmarshal([]byte(cachedData), &book)

		if errJson != nil {
			return nil, errJson
		} else {
			return &book, nil
		}
	}

	log.Println("Cache MISS or Redis Down. Fetching from DB for ID:", ID)

	book, errDB := repo.base.GetById(ctx, ID)

	if errDB != nil {
		return nil, errDB
	}

	if repo.rds != nil {
		go func(data *models.Books, key string) {
			bgContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			bytes, _ := json.Marshal(data)
			repo.rds.Set(bgContext, key, bytes, repo.cfg.RedisCacheTTL)
		}(book, cacheKey)
	}

	return book, nil

}

func (repo *BookCacheRepository) GetAll(ctx context.Context, param dto.GetBooksQuery) ([]models.Books, int64, error) {
	return repo.base.GetAll(ctx, param)
}

func (repo *BookCacheRepository) Delete(ctx context.Context, ID string) error {
	err := repo.base.Delete(ctx, ID)

	if err == nil && repo.rds != nil {
		cacheKey := "book:" + ID

		repo.rds.Del(ctx, cacheKey)
	}

	return err
}

func (repo *BookCacheRepository) Update(ctx context.Context, book *models.Books, ID string) error {

	err := repo.base.Update(ctx, book, ID)

	if err == nil && repo.rds != nil {
		cacheKey := "book:" + ID

		repo.rds.Del(ctx, cacheKey)
	}

	return err

}
