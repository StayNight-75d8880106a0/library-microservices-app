package registry

import (
	"user-management-services/internal/config"
	"user-management-services/internal/delivery/controller"
	"user-management-services/internal/infrastructure/kafka/consumer"
	"user-management-services/internal/infrastructure/kafka/producer"
	"user-management-services/internal/infrastructure/keycloak"
	"user-management-services/internal/repository"
	"user-management-services/internal/usecase"

	"gorm.io/gorm"
)

type UserModule struct {
	UserController  *controller.UserController
	AdminController *controller.AdminController
	UserConsumer    *consumer.KafkaConsumer
	UserUsecase     *usecase.UserProfileUsecase
}

func NewUserModuleRegistry(appConfig *config.AppConfig, db *gorm.DB) *UserModule {

	keycloak := keycloak.NewKeycloakClientUserRegistry(appConfig.Keycloak)

	repository := repository.NewUserRepositoryRegistry(db)

	kafkaProducer := producer.NewKafkaProducer(appConfig.Kafka.Brokers)

	userUsecase := usecase.NewUserProfileUsecaseRegistry(keycloak, repository, kafkaProducer, appConfig)

	adminUsecase := usecase.NewAdminUsecaseRegistry(keycloak)

	kafkaReader := consumer.NewKafkaConsumer(appConfig.Kafka.Brokers, appConfig.Kafka.TopicUserCreated, appConfig.Kafka.GroupID, userUsecase)

	userControllerRegistry := controller.NewUserControllerRegistry(userUsecase)

	adminControllerRegistry := controller.NewAdminControllerRegistry(adminUsecase)

	return &UserModule{
		UserController:  userControllerRegistry,
		AdminController: adminControllerRegistry,
		UserConsumer:    kafkaReader,
		UserUsecase:     userUsecase,
	}

}
