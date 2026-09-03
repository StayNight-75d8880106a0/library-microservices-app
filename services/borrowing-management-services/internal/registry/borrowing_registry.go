package registry

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/delivery/controller"
	"borrowing-management-services/internal/repository"
	"borrowing-management-services/internal/repository/cache"
	"borrowing-management-services/internal/usecase"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type BorrowingModule struct {
	BorrowingController *controller.BorrowingController
}

func NewBorrowingRegistryModule(db *gorm.DB, rds *redis.Client, cfg *config.AppConfig) *BorrowingModule {

	repository := repository.NewBorrowingUserRepository(db)

	cacheRepository := cache.NewBorrowingUserCacheRepository(repository, rds, cfg.RedisCacheConfig)

	usecase := usecase.NewBorrowingUsecase(cacheRepository)

	controller := controller.NewBorrowingController(usecase)

	return &BorrowingModule{
		BorrowingController: controller,
	}

}
