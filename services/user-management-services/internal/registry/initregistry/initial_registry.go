package initregistry

import (
	"user-management-services/internal/config"
	"user-management-services/internal/registry"

	"gorm.io/gorm"
)

type Modules struct {
	User *registry.UserModule
}

func NewInitRegistry(appConfig *config.AppConfig, db *gorm.DB) *Modules {

	UserModule := registry.NewUserModuleRegistry(appConfig, db)

	return &Modules{
		User: UserModule,
	}

}
