package repository

import (
	"context"
	"notification-services/internal/models"

	"gorm.io/gorm"
)

type EmailRepositoryInterface interface {
	SaveEmailLog(ctx context.Context, emailLog *models.EmailLogs) error
	GetAllEmailLogs(ctx context.Context, limit int, offset int) ([]models.EmailLogs, int64, error)
	GetEmailLogByID(ctx context.Context, ID string) (*models.EmailLogs, error)
}

type EmailRepository struct {
	DB *gorm.DB
}

func NewEmailRepository(db *gorm.DB) *EmailRepository {
	return &EmailRepository{
		DB: db,
	}
}

func (repo *EmailRepository) SaveEmailLog(ctx context.Context, emailLog *models.EmailLogs) error {

	errCreate := repo.DB.WithContext(ctx).Table("email_logs").Create(emailLog).Error

	return errCreate

}

func (repo *EmailRepository) GetAllEmailLogs(ctx context.Context, limit int, offset int) ([]models.EmailLogs, int64, error) {

	emailLogs := []models.EmailLogs{}
	var total int64

	errCount := repo.DB.WithContext(ctx).Table("email_logs").Count(&total).Error

	if errCount != nil {
		return emailLogs, 0, errCount
	}

	errGet := repo.DB.WithContext(ctx).Table("email_logs").Limit(limit).Offset(offset).Find(&emailLogs).Error

	return emailLogs, total, errGet

}

func (repo *EmailRepository) GetEmailLogByID(ctx context.Context, ID string) (*models.EmailLogs, error) {

	var emailLog models.EmailLogs

	errGet := repo.DB.WithContext(ctx).Table("email_logs").Where("id = ?", ID).First(&emailLog).Error

	return &emailLog, errGet

}
