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

type WaitingListModule struct {
	WaitingListController *controller.WaitingListController
}

func NewWaitingListRegistryModule(db *gorm.DB, rds *redis.Client, cfg *config.AppConfig) *WaitingListModule {

	repository := repository.NewWaitingListRepository(db)

	cacheRepository := cache.NewWaitingListCacheRepository(repository, rds, cfg.RedisCacheConfig)

	usecase := usecase.NewWaitingListUsecase(cacheRepository)

	controller := controller.NewWaitingListController(usecase)

	return &WaitingListModule{
		WaitingListController: controller,
	}

}
