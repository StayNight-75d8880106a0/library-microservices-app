package repository

import (
	"activity-logbook-services/internal/models"
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LogbookRepositoryInterface interface {
	Create(ctx context.Context, logbook *models.Logbook) error
	GetAll(ctx context.Context, limit int, cursorOccurredAt *time.Time, cursorID *string) ([]models.Logbook, error)
	GetByTraceID(ctx context.Context, traceID string) ([]models.Logbook, error)
	GetByID(ctx context.Context, ID string, traceID string) (*models.Logbook, error)
}

type LogbookRepository struct {
	DB *gorm.DB
}

func NewLogbookRepository(db *gorm.DB) *LogbookRepository {
	return &LogbookRepository{
		DB: db,
	}
}

func (repo *LogbookRepository) Create(ctx context.Context, logbook *models.Logbook) error {

	errCreate := repo.DB.WithContext(ctx).Table("logbook_audits").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "event_id"}, {Name: "occurred_at"}},
		DoNothing: true,
	}).Create(&logbook).Error

	return errCreate

}

func (repo *LogbookRepository) GetAll(ctx context.Context, limit int, cursorOccurredAt *time.Time, cursorID *string) ([]models.Logbook, error) {

	var logbooks []models.Logbook

	query := repo.DB.WithContext(ctx).Table("logbook_audits").Select("id, trace_id, user_id, occurred_at").Where("is_root = ?", true)

	if cursorOccurredAt != nil && cursorID != nil {
		query = query.Where("(occurred_at, id) < (?, ?)", *cursorOccurredAt, *cursorID)
	}

	errGet := query.Order("occurred_at DESC, id DESC").Limit(limit).Find(&logbooks).Error

	return logbooks, errGet
}

func (repo *LogbookRepository) GetByTraceID(ctx context.Context, traceID string) ([]models.Logbook, error) {

	var logbooks []models.Logbook

	errGet := repo.DB.WithContext(ctx).Table("logbook_audits").Where("trace_id = ?", traceID).Order("occurred_at ASC").Find(&logbooks).Error

	return logbooks, errGet

}

func (repo *LogbookRepository) GetByID(ctx context.Context, ID string, traceID string) (*models.Logbook, error) {

	var logbook models.Logbook

	errGet := repo.DB.WithContext(ctx).Table("logbook_audits").Where("id = ? AND trace_id = ?", ID, traceID).First(&logbook).Error

	return &logbook, errGet

}
