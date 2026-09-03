package repository

import (
	"borrowing-management-services/internal/models"
	"context"

	"gorm.io/gorm"
)

type BorrowingUserRepositoryInterface interface {
	CreateBorrowing(ctx context.Context, borrowing *models.Borrowing) error
	GetAllMyBorrowings(ctx context.Context, userID string, limit int, offset int) ([]models.Borrowing, int64, error)
	GetAllBorrowings(ctx context.Context, limit int, offset int) ([]models.Borrowing, int64, error)
	GetBorrowingByID(ctx context.Context, ID string) (*models.Borrowing, error)
	UpdateStatus(ctx context.Context, ID string, status models.BorrowingStatus) error
}

type BorrowingUserRepository struct {
	db *gorm.DB
}

func NewBorrowingUserRepository(db *gorm.DB) *BorrowingUserRepository {
	return &BorrowingUserRepository{
		db: db,
	}
}

func (repo *BorrowingUserRepository) CreateBorrowing(ctx context.Context, borrowing *models.Borrowing) error {

	errCreate := repo.db.WithContext(ctx).Table("borrowings").Create(borrowing).Error

	return errCreate

}

func (repo *BorrowingUserRepository) GetAllMyBorrowings(ctx context.Context, userID string, limit int, offset int) ([]models.Borrowing, int64, error) {

	var borrowings []models.Borrowing

	var total int64

	errCount := repo.db.WithContext(ctx).Table("borrowings").Where("user_id = ?", userID).Count(&total).Error

	if errCount != nil {
		return borrowings, 0, errCount
	}

	errGet := repo.db.WithContext(ctx).Table("borrowings").Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&borrowings).Error

	return borrowings, total, errGet

}

func (repo *BorrowingUserRepository) GetAllBorrowings(ctx context.Context, limit int, offset int) ([]models.Borrowing, int64, error) {

	var borrowings []models.Borrowing

	var total int64

	errCount := repo.db.WithContext(ctx).Table("borrowings").Count(&total).Error

	if errCount != nil {
		return borrowings, 0, errCount
	}

	errGet := repo.db.WithContext(ctx).Table("borrowings").Limit(limit).Offset(offset).Find(&borrowings).Error

	return borrowings, total, errGet

}

func (repo *BorrowingUserRepository) GetBorrowingByID(ctx context.Context, ID string) (*models.Borrowing, error) {

	var borrowing models.Borrowing

	errGet := repo.db.WithContext(ctx).Table("borrowings").Where("id = ?", ID).First(&borrowing).Error

	return &borrowing, errGet

}

func (repo *BorrowingUserRepository) UpdateStatus(ctx context.Context, ID string, status models.BorrowingStatus) error {

	errUpdate := repo.db.WithContext(ctx).Table("borrowings").Where("id = ?", ID).Updates(map[string]interface{}{
		"status":      status,
		"returned_at": gorm.Expr("NOW()"),
		"updated_at":  gorm.Expr("NOW()"),
	}).Error

	return errUpdate

}
