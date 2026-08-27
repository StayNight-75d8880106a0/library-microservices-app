package repository

import (
	"book-management-services/internal/dto"
	"book-management-services/internal/models"
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

type BookRepositoryInterface interface {
	Create(ctx context.Context, book *models.Books) error
	GetAll(ctx context.Context, param dto.GetBooksQuery) ([]models.Books, int64, error)
	GetById(ctx context.Context, ID string) (*models.Books, error)
	Delete(ctx context.Context, ID string) error
	Update(ctx context.Context, book *models.Books, ID string) error
}

type BookRepository struct {
	DB *gorm.DB
	es *elasticsearch.Client
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{
		DB: db,
	}
}

func (repo *BookRepository) Create(ctx context.Context, book *models.Books) error {

	errCreate := repo.DB.WithContext(ctx).Table("books").Create(book).Error

	return errCreate

}

func (repo *BookRepository) GetAll(ctx context.Context, param dto.GetBooksQuery) ([]models.Books, int64, error) {

	var books []models.Books
	var totalData int64

	query := repo.DB.WithContext(ctx).Table("books").Where("deleted_at IS NULL")

	if param.Category != "" {
		query = query.Where("category = ?", param.Category)
	}

	errCount := query.Count(&totalData)

	if errCount.Error != nil {
		return nil, 0, errCount.Error
	}

	offset := (param.Page - 1) * param.Limit

	errFind := query.Order("created_at DESC").Limit(param.Limit).Offset(offset).Find(&books).Error

	return books, totalData, errFind

}

func (repo *BookRepository) GetById(ctx context.Context, ID string) (*models.Books, error) {

	var book models.Books

	errFind := repo.DB.WithContext(ctx).Table("books").Where("id = ? AND deleted_at IS NULL", ID).First(&book).Error

	return &book, errFind

}

func (repo *BookRepository) Delete(ctx context.Context, ID string) error {

	errDelete := repo.DB.WithContext(ctx).Table("books").Where("id = ?", ID).Delete(&models.Books{}).Error

	return errDelete

}

func (repo *BookRepository) Update(ctx context.Context, book *models.Books, ID string) error {

	errUpdate := repo.DB.WithContext(ctx).Table("books").Where("id = ?", ID).Updates(book).Error

	return errUpdate

}
