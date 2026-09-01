package registry

import (
	"borrowing-management-services/internal/client"
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/infrastructure/kafka/consume"
	"borrowing-management-services/internal/infrastructure/kafka/event"
	"borrowing-management-services/internal/repository"

	"github.com/redis/go-redis/v9"
)

type KafkaCacheRegistryModule struct {
	UserCacheRepo  repository.UserRedisCacheInterface
	UserGrpcClient client.UserGrpcClientInterface
	AuthConsumer   *consume.KafkaConsumer
	StatusConsumer *consume.KafkaConsumer
	EventHandler   *event.EventHandler
}

func NewKafkaCacheRegistryModule(rds *redis.Client, cfg *config.AppConfig) *KafkaCacheRegistryModule {

	userCacheRepo := repository.NewUserRedisCache(rds)

	userGrpcClient, _ := client.NewUserGrpcClient(cfg.PortConfig.GRPC)

	eventHandler := event.NewEventHandler(userCacheRepo, userGrpcClient)

	authConsumer := consume.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.TopicUserAuthenticated, cfg.Kafka.GroupID, eventHandler)

	statusConsumer := consume.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.TopicUserStatusUpdated, cfg.Kafka.GroupID, eventHandler)

	return &KafkaCacheRegistryModule{
		UserCacheRepo:  userCacheRepo,
		UserGrpcClient: userGrpcClient,
		AuthConsumer:   authConsumer,
		StatusConsumer: statusConsumer,
		EventHandler:   eventHandler,
	}
}
