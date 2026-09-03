package initregistry

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/registry"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Module struct {
	KafkaCacheRegistry *registry.KafkaCacheRegistryModule
	BorrowingRegistry  *registry.BorrowingModule
}

func NewInitRegistry(rds *redis.Client, cfg *config.AppConfig, db *gorm.DB) *Module {
	kafkaCacheRegistry := registry.NewKafkaCacheRegistryModule(rds, cfg)
	borrowingRegistry := registry.NewBorrowingRegistryModule(db, rds, cfg)

	return &Module{
		KafkaCacheRegistry: kafkaCacheRegistry,
		BorrowingRegistry:  borrowingRegistry,
	}
}
