package repository

import (
	"context"
	"user-management-services/internal/models"

	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	CreateUserFromEvent(ctx context.Context, user *models.Users) error
}

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepositoryRegistry(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (repo *UserRepository) CreateUserFromEvent(ctx context.Context, user *models.Users) error {

	errCreate := repo.DB.WithContext(ctx).Table("user_profiles").Create(&user).Error

	return errCreate

}
