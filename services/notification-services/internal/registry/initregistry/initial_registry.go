package initregistry

import (
	"notification-services/internal/config"
	"notification-services/internal/registry"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Module struct {
	EmailLog *registry.EmailModule
}

func NewInitRegistry(db *gorm.DB, rds *redis.Client, cfg *config.AppConfig) *Module {

	emailLogModule := registry.NewEmailLogModuleRegistry(db, rds, cfg)

	return &Module{
		EmailLog: emailLogModule,
	}

}
