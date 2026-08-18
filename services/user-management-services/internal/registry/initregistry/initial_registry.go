package initregistry

import (
	"user-management-services/internal/config"
	"user-management-services/internal/registry"
)

type Modules struct {
	User *registry.UserModule
}

func NewInitRegistry(appConfig *config.AppConfig) *Modules {

	UserModule := registry.NewUserModuleRegistry(appConfig)

	return &Modules{
		User: UserModule,
	}

}
