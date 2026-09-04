package registry

import (
	"book-management-services/internal/client"
	"book-management-services/internal/config"
	"book-management-services/internal/delivery/controller"
	"book-management-services/internal/infrastructure/kafka/consumer"
	"book-management-services/internal/repository"
	"book-management-services/internal/usecase"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type BookModule struct {
	BookController *controller.BookController
	BookConsumer   *consumer.KafkaConsumer
}

func NewBookModuleRegistry(db *gorm.DB, rds *redis.Client, elasticsearch *elasticsearch.Client, cfg *config.AppConfig) *BookModule {

	bookRepository := repository.NewBookRepository(db)

	bookRepositoryCache := repository.NewBookCacheRepository(bookRepository, rds, cfg.RedisConfig)

	elasticsearchRepository := repository.NewElasticSearchRepository(elasticsearch)

	openLibraryClient := client.NewOpenLibraryClient(*cfg.OpenLibraryAPI)

	usecase := usecase.NewBookUsecase(bookRepositoryCache, elasticsearchRepository, openLibraryClient)

	kafkaConsumer := consumer.NewKafkaConsumer(cfg.KafkaConfig.Brokers, cfg.KafkaConfig.TopicBorrowingCreated, cfg.KafkaConfig.GroupID, usecase)

	controller := controller.NewBookController(usecase)

	return &BookModule{
		BookController: controller,
		BookConsumer:   kafkaConsumer,
	}

}
