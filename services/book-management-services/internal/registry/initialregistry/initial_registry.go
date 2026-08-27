package initialregistry

import (
	"book-management-services/internal/config"
	"book-management-services/internal/registry"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Module struct {
	Book *registry.BookModule
}

func NewInitRegistry(db *gorm.DB, rds *redis.Client, elasticsearch *elasticsearch.Client, cfg *config.AppConfig) *Module {

	BookModule := registry.NewBookModuleRegistry(db, rds, elasticsearch, cfg)

	return &Module{
		Book: BookModule,
	}

}
