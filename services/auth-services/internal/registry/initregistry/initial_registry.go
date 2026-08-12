package initregistry

import (
	"auth-services/internal/config"
	"auth-services/internal/helper"
	"auth-services/internal/registry"

	"github.com/redis/go-redis/v9"
)

type Modules struct {
	Dependencies *helper.Depedencies
	Auth         *registry.AuthModule
}

func NewInitRegistry(appConfig *config.AppConfig, rds *redis.Client) *Modules {
	AuthModule := registry.NewAuthModuleRegistry(appConfig, rds)
	deps := &helper.Depedencies{
		IsblacklistToken: AuthModule.AuthRepository,
	}
	return &Modules{
		Dependencies: deps,
		Auth:         AuthModule,
	}
}
