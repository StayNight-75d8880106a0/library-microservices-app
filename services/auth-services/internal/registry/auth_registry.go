package registry

import (
	"auth-services/internal/config"
	"auth-services/internal/delivery/controller"
	"auth-services/internal/infrastructure/keycloak"
	"auth-services/internal/repository"
	"auth-services/internal/usecase"

	"github.com/redis/go-redis/v9"
)

type AuthModule struct {
	AuthController *controller.AuthController
	AuthRepository *repository.AuthRepository
}

func NewAuthModuleRegistry(appConfig *config.AppConfig, rds *redis.Client) *AuthModule {

	keyCloakClient := keycloak.NewKeycloakClientRegistry(appConfig.Keycloak)

	repository := repository.NewAuthRepositoryRegistry(rds)

	usecase := usecase.NewAuthUsecaseRegistry(keyCloakClient, *repository)

	controller := controller.NewAuthControllerRegistry(usecase)

	return &AuthModule{
		AuthController: controller,
		AuthRepository: repository,
	}
}
