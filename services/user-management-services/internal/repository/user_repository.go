package repository

import (
	"context"
	"user-management-services/internal/models"

	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	CreateUserFromEvent(ctx context.Context, user *models.Users) error
	GetUserByKeycloakID(ctx context.Context, keycloakID string) (*models.Users, error)
	UpdateUserProfileDB(ctx context.Context, user *models.Users, keycloakID string) error
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

func (repo *UserRepository) GetUserByKeycloakID(ctx context.Context, keycloakID string) (*models.Users, error) {

	var user models.Users

	errGet := repo.DB.WithContext(ctx).Table("user_profiles").Where("keycloak_user_id = ?", keycloakID).First(&user).Error

	return &user, errGet

}

func (repo *UserRepository) UpdateUserProfileDB(ctx context.Context, user *models.Users, keycloakID string) error {

	errUpdate := repo.DB.WithContext(ctx).Table("user_profiles").Where("keycloak_user_id = ?", keycloakID).Updates(&user).Error

	return errUpdate
}
