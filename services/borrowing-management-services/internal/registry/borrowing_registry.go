package registry

import (
	"borrowing-management-services/internal/client"
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/delivery/controller"
	"borrowing-management-services/internal/infrastructure/kafka/producer"
	redisdb "borrowing-management-services/internal/infrastructure/redis"
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

	waitingListRepository := repository.NewWaitingListRepository(db)

	cacheRepository := cache.NewBorrowingUserCacheRepository(borrowingRepository, rds, cfg.RedisCacheConfig)

	waitingListCacheRepository := cache.NewWaitingListCacheRepository(waitingListRepository, rds, cfg.RedisCacheConfig)

	userCache := repository.NewUserRedisCache(rds)

	kafkaProducer := producer.NewKafkaProducer(cfg.Kafka.Brokers)

	userGrpc, errUserGRPC := client.NewUserGrpcClient(cfg.PortConfig.GRPCHOST + ":" + cfg.PortConfig.GRPC)

	if errUserGRPC != nil {
		panic(errUserGRPC)
	}

	bookGrpc, errBookGRPC := client.NewBookGrpcClient(cfg.PortConfig.BOOKGRPCHOST + ":" + cfg.PortConfig.BOOKGRPC)

	if errBookGRPC != nil {
		panic(errBookGRPC)
	}

	locker := redisdb.NewLocker(rds)

	waitingListUsecase := usecase.NewWaitingListUsecase(waitingListCacheRepository)

	usecase := usecase.NewBorrowingUsecase(cacheRepository, userCache, kafkaProducer, cfg, userGrpc, waitingListCacheRepository, bookGrpc, waitingListUsecase, locker)

	controller := controller.NewBorrowingController(usecase)

	return &BorrowingModule{
		BorrowingController: controller,
	}

}
