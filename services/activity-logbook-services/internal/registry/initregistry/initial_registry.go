package initregistry

import (
	"activity-logbook-services/internal/config"
	"activity-logbook-services/internal/registry"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Module struct {
	Logbook *registry.LogbookModule
}

func NewInitRegistry(db *gorm.DB, rds *redis.Client, cfg *config.AppConfig) *Module {
	logbookModule := registry.NewLogbookModuleRegistry(db, rds, cfg)

	return &Module{
		Logbook: logbookModule,
	}
}
