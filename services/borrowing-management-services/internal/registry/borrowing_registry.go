package registry

import (
	"borrowing-management-services/internal/client"
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/delivery/controller"
	"borrowing-management-services/internal/infrastructure/kafka/producer"
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

	borrowingRepository := repository.NewBorrowingUserRepository(db)

	cacheRepository := cache.NewBorrowingUserCacheRepository(borrowingRepository, rds, cfg.RedisCacheConfig)

	userCache := repository.NewUserRedisCache(rds)

	kafkaProducer := producer.NewKafkaProducer(cfg.Kafka.Brokers)

	userGrpc, _ := client.NewUserGrpcClient(cfg.PortConfig.GRPCHOST + ":" + cfg.PortConfig.GRPC)

	usecase := usecase.NewBorrowingUsecase(cacheRepository, userCache, kafkaProducer, cfg, userGrpc)

	controller := controller.NewBorrowingController(usecase)

	return &BorrowingModule{
		BorrowingController: controller,
	}

}
