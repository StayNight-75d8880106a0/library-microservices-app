package initregistry

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/registry"

	"github.com/redis/go-redis/v9"
)

type Module struct {
	KafkaCacheRegistry *registry.KafkaCacheRegistryModule
}

func NewInitRegistry(rds *redis.Client, cfg *config.AppConfig) *Module {
	kafkaCacheRegistry := registry.NewKafkaCacheRegistryModule(rds, cfg)

	return &Module{
		KafkaCacheRegistry: kafkaCacheRegistry,
	}
}
