package registry

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/delivery/controller"
	redisdb "borrowing-management-services/internal/infrastructure/redis"
	"borrowing-management-services/internal/repository"
	"borrowing-management-services/internal/repository/cache"
	"borrowing-management-services/internal/usecase"
	"borrowing-management-services/internal/usecase/cronjob"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type WaitingListModule struct {
	WaitingListController *controller.WaitingListController
	WaitingListCronJob    *cronjob.CronService
}

func NewWaitingListRegistryModule(db *gorm.DB, rds *redis.Client, cfg *config.AppConfig, cron *cron.Cron) *WaitingListModule {

	repository := repository.NewWaitingListRepository(db)

	cacheRepository := cache.NewWaitingListCacheRepository(repository, rds, cfg.RedisCacheConfig)

	locker := redisdb.NewLocker(rds)

	usecase := usecase.NewWaitingListUsecase(cacheRepository)

	cronJob := cronjob.NewCronService(usecase, cron, locker)

	controller := controller.NewWaitingListController(usecase)

	return &WaitingListModule{
		WaitingListController: controller,
		WaitingListCronJob:    cronJob,
	}

}
