package registry

import (
	"user-management-services/internal/config"
	"user-management-services/internal/delivery/controller"
	"user-management-services/internal/infrastructure/keycloak"
	"user-management-services/internal/usecase"
)

type UserModule struct {
	UserController  *controller.UserController
	AdminController *controller.AdminController
}

func NewUserModuleRegistry(appConfig *config.AppConfig) *UserModule {

	keycloak := keycloak.NewKeycloakClientUserRegistry(appConfig.Keycloak)

	userUsecase := usecase.NewUserProfileUsecaseRegistry(keycloak)

	adminUsecase := usecase.NewAdminUsecaseRegistry(keycloak)

	userControllerRegistry := controller.NewUserControllerRegistry(userUsecase)

	adminControllerRegistry := controller.NewAdminControllerRegistry(adminUsecase)

	return &UserModule{
		UserController:  userControllerRegistry,
		AdminController: adminControllerRegistry,
	}

}
