package initregistry

import (
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/registry"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type Module struct {
	KafkaCacheRegistry  *registry.KafkaCacheRegistryModule
	BorrowingRegistry   *registry.BorrowingModule
	WaitingListRegistry *registry.WaitingListModule
}

func NewInitRegistry(rds *redis.Client, cfg *config.AppConfig, db *gorm.DB, cron *cron.Cron) *Module {
	kafkaCacheRegistry := registry.NewKafkaCacheRegistryModule(rds, cfg)
	borrowingRegistry := registry.NewBorrowingRegistryModule(db, rds, cfg)
	waitingListRegistry := registry.NewWaitingListRegistryModule(db, rds, cfg, cron)

	return &Module{
		KafkaCacheRegistry:  kafkaCacheRegistry,
		BorrowingRegistry:   borrowingRegistry,
		WaitingListRegistry: waitingListRegistry,
	}
}
