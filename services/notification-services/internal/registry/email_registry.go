package registry

import (
	"notification-services/internal/config"
	"notification-services/internal/delivery/controller"
	"notification-services/internal/infrastructure/kafka/consumer"
	"notification-services/internal/repository"
	"notification-services/internal/repository/cache"
	"notification-services/internal/usecase"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type EmailModule struct {
	EmailLogController *controller.EmailController
	EmailLogConsumer   *consumer.KafkaConsumer
}

func NewEmailLogModuleRegistry(db *gorm.DB, rds *redis.Client, cfg *config.AppConfig) *EmailModule {

	repository := repository.NewEmailRepository(db)

	cacheRepository := cache.NewEmailCacheRepository(repository, rds, cfg.RedisConfig)

	usecase := usecase.NewEmailUsecase(cacheRepository, cfg.SMTPConfig)

	kafkaConsumer := consumer.NewKafkaConsumer(cfg.KafkaConfig.Brokers, cfg.KafkaConfig.TopicUserCreated, cfg.KafkaConfig.GroupID, usecase)

	controller := controller.NewEmailController(usecase)

	return &EmailModule{
		EmailLogController: controller,
		EmailLogConsumer:   kafkaConsumer,
	}

}
