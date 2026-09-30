package registry

import (
	"activity-logbook-services/internal/config"
	"activity-logbook-services/internal/delivery/controller"
	"activity-logbook-services/internal/repository"
	"activity-logbook-services/internal/usecase"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type LogbookModule struct {
	LogbookController *controller.LogbookController
}

func NewLogbookModuleRegistry(db *gorm.DB, rds *redis.Client, cfg *config.AppConfig) *LogbookModule {

	logbookRepository := repository.NewLogbookRepository(db)

	logbookCacheRepository := repository.NewLogbookCacheRepository(logbookRepository, rds, cfg.RedisConfig)

	logbookUsecase := usecase.NewLogbookUsecase(logbookCacheRepository)

	logbookController := controller.NewLogbookController(logbookUsecase)

	return &LogbookModule{
		LogbookController: logbookController,
	}

}
