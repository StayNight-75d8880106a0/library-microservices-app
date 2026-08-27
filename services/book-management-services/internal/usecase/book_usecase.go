package usecase

import (
	"book-management-services/internal/client"
	"book-management-services/internal/dto"
	"book-management-services/internal/helper"
	"book-management-services/internal/models"
	"book-management-services/internal/repository"
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type BookUsecaseInterface interface {
	CreateBook(ctx context.Context, request *dto.CreateBookRequest) (*dto.BookResponse, error)
	GetAllBooks(ctx context.Context, param dto.GetBooksQuery) ([]dto.BookResponse, helper.PaginationMeta, error)
	GetBookByID(ctx context.Context, ID string) (*dto.BookResponse, error)
	DeleteBook(ctx context.Context, ID string) error
	UpdateBook(ctx context.Context, request *dto.UpdateBookRequest, ID string) (*dto.BookResponse, error)
}

type BookUsecase struct {
	repository    repository.BookRepositoryInterface
	elasticsearch repository.ElasticsearchRepositoryInterface
	openLibrary   client.OpenLibraryClientInterface
}

func NewBookUsecase(bookRepository repository.BookRepositoryInterface, es repository.ElasticsearchRepositoryInterface, ol client.OpenLibraryClientInterface) *BookUsecase {
	return &BookUsecase{
		repository:    bookRepository,
		elasticsearch: es,
		openLibrary:   ol,
	}
}

func (u *BookUsecase) CreateBook(ctx context.Context, request *dto.CreateBookRequest) (*dto.BookResponse, error) {

	if request.ISBN == nil || *request.ISBN == "" {
		return nil, helper.NewUnprocessableEntityError("ISBN Cannot Be Empty!", helper.ErrorDetail{Detail: "ISBN is required!"})
	}

	if request.TotalStock == nil || *request.TotalStock <= 0 {
		return nil, helper.NewUnprocessableEntityError("Total Stock Cannot Be Empty!", helper.ErrorDetail{Detail: "Total Stock is required!"})
	}

	if request.AvailableStock == nil || *request.AvailableStock < 0 {
		return nil, helper.NewUnprocessableEntityError("Available Stock Cannot Be Empty!", helper.ErrorDetail{Detail: "Available Stock is required!"})
	}

	olBook, errOL := u.openLibrary.FetchByISBN(ctx, *request.ISBN)

	if errOL != nil {
		return nil, errOL
	}

	rawDate := ""
	if request.PublishedDate != "" && request.PublishedDate != "0000-00-00" {
		rawDate = helper.NormalizeDate(request.PublishedDate)
	} else if olBook != nil && olBook.PublishedDate != "" {
		rawDate = helper.NormalizeDate(olBook.PublishedDate)
	}

	book := &models.Books{
		ISBN:           *request.ISBN,
		Title:          request.Title,
		Authors:        request.Authors,
		Publisher:      request.Publisher,
		PublishedDate:  rawDate,
		Page:           request.Page,
		Description:    request.Description,
		CoverURL:       request.CoverURL,
		Category:       request.Category,
		TotalStock:     *request.TotalStock,
		AvailableStock: *request.AvailableStock,
	}

	if olBook != nil {
		if book.Title == "" {
			book.Title = olBook.Title
		}
		if book.Publisher == "" {
			book.Publisher = olBook.Publisher
		}
		if book.PublishedDate == "" {
			book.PublishedDate = olBook.PublishedDate
		}
		if book.Page == 0 {
			book.Page = olBook.Page
		}
		if book.CoverURL == "" {
			book.CoverURL = olBook.CoverURL
		}
		if book.Authors == "" && len(olBook.Authors) > 0 {
			book.Authors = strings.Join(olBook.Authors, ", ")
		}
		if book.Category == "" && len(olBook.Subjects) > 0 {
			book.Category = helper.ResolveCategory(olBook.Subjects)
		}
	}

	if book.Title == "" {
		return nil, helper.NewUnprocessableEntityError("Book Title Cannot Be Empty!", helper.ErrorDetail{Detail: "Title is required"})
	}

	if book.Category == "" {
		book.Category = "General"
	}

	errCreate := u.repository.Create(ctx, book)

	if errCreate != nil {
		return nil, errCreate
	}

	_ = u.elasticsearch.IndexToElasticsearch(ctx, book)

	result := &dto.BookResponse{
		ID:             &book.ID,
		ISBN:           &book.ISBN,
		Title:          &book.Title,
		Authors:        &book.Authors,
		Publisher:      &book.Publisher,
		PublishedDate:  &book.PublishedDate,
		Page:           &book.Page,
		Description:    &book.Description,
		CoverURL:       &book.CoverURL,
		Category:       &book.Category,
		TotalStock:     &book.TotalStock,
		AvailableStock: &book.AvailableStock,
		CreatedAt:      helper.FormatTimeRFC3339Jakarta(book.CreatedAt),
		UpdatedAt:      helper.FormatTimeRFC3339Jakarta(book.UpdatedAt),
	}

	return result, nil

}

func (u *BookUsecase) GetAllBooks(ctx context.Context, param dto.GetBooksQuery) ([]dto.BookResponse, helper.PaginationMeta, error) {

	if param.Page < 1 {
		param.Page = 1
	}

	if param.Limit < 1 {
		param.Limit = 10
	}

	var books []models.Books
	var totalData int64
	var err error

	if param.Keywords != "" {
		books, totalData, err = u.elasticsearch.Search(ctx, param)
	} else {
		books, totalData, err = u.repository.GetAll(ctx, param)
	}

	if err != nil {
		return nil, helper.PaginationMeta{}, helper.NewInternalServerError("An Error During Get Books Data!", helper.ErrorDetail{Detail: err.Error()})
	}

	totalPage := int((totalData + int64(param.Limit) - 1) / int64(param.Limit))

	result := make([]dto.BookResponse, 0, len(books))

	for _, value := range books {
		result = append(result, dto.BookResponse{
			ID:             &value.ID,
			ISBN:           &value.ISBN,
			Title:          &value.Title,
			Authors:        &value.Authors,
			Publisher:      &value.Publisher,
			PublishedDate:  &value.PublishedDate,
			Page:           &value.Page,
			Description:    &value.Description,
			CoverURL:       &value.CoverURL,
			Category:       &value.Category,
			TotalStock:     &value.TotalStock,
			AvailableStock: &value.AvailableStock,
			CreatedAt:      helper.FormatTimeRFC3339Jakarta(value.CreatedAt),
			UpdatedAt:      helper.FormatTimeRFC3339Jakarta(value.UpdatedAt),
		})
	}

	var kw *string

	if param.Keywords != "" {
		kw = &param.Keywords
	}

	pagination := &helper.PaginationMeta{
		Page:      param.Page,
		Limit:     param.Limit,
		TotalData: totalData,
		TotalPage: totalPage,
		Keywords:  kw,
		Category:  &param.Category,
	}

	return result, *pagination, nil
}

func (u *BookUsecase) GetBookByID(ctx context.Context, ID string) (*dto.BookResponse, error) {

	book, errGet := u.repository.GetById(ctx, ID)

	if errGet != nil {
		if errors.Is(errGet, gorm.ErrRecordNotFound) {
			return nil, helper.NewNotFoundError("Mahasiswa Not Found!", helper.ErrorDetail{})
		}
		return nil, helper.NewInternalServerError("An Error During Get Mahasiswa By ID!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	result := &dto.BookResponse{
		ID:             &book.ID,
		ISBN:           &book.ISBN,
		Title:          &book.Title,
		Authors:        &book.Authors,
		Publisher:      &book.Publisher,
		PublishedDate:  &book.PublishedDate,
		Page:           &book.Page,
		Description:    &book.Description,
		CoverURL:       &book.CoverURL,
		Category:       &book.Category,
		TotalStock:     &book.TotalStock,
		AvailableStock: &book.AvailableStock,
		CreatedAt:      helper.FormatTimeRFC3339Jakarta(book.CreatedAt),
		UpdatedAt:      helper.FormatTimeRFC3339Jakarta(book.UpdatedAt),
	}

	return result, nil
}

func (u *BookUsecase) DeleteBook(ctx context.Context, ID string) error {

	_, errGet := u.repository.GetById(ctx, ID)

	if errGet != nil {
		if errors.Is(errGet, gorm.ErrRecordNotFound) {
			return helper.NewNotFoundError("Book Not Found!", helper.ErrorDetail{})
		}
		return helper.NewInternalServerError("An Error During Get Book By ID!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	errDelete := u.repository.Delete(ctx, ID)

	if errDelete != nil {
		return helper.NewInternalServerError("An Error During Delete Book!", helper.ErrorDetail{Detail: errDelete.Error()})
	}

	errESDelete := u.elasticsearch.DeleteFromElasticsearch(ctx, ID)

	if errESDelete != nil {
		return helper.NewInternalServerError("An Error During Delete Book From Elasticsearch!", helper.ErrorDetail{Detail: errESDelete.Error()})
	}

	return nil

}

func (u *BookUsecase) UpdateBook(ctx context.Context, request *dto.UpdateBookRequest, ID string) (*dto.BookResponse, error) {

	book, errGet := u.repository.GetById(ctx, ID)

	if errGet != nil {
		if errors.Is(errGet, gorm.ErrRecordNotFound) {
			return nil, helper.NewNotFoundError("Book Not Found!", helper.ErrorDetail{})
		}
		return nil, helper.NewInternalServerError("An Error During Get Book By ID!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	if request.ISBN == nil || *request.ISBN == "" {
		return nil, helper.NewUnprocessableEntityError("ISBN Cannot Be Empty!", helper.ErrorDetail{Detail: "ISBN is required!"})
	}

	if request.TotalStock == nil || *request.TotalStock <= 0 {
		return nil, helper.NewUnprocessableEntityError("Total Stock Cannot Be Empty!", helper.ErrorDetail{Detail: "Total Stock is required!"})
	}

	if request.AvailableStock == nil || *request.AvailableStock < 0 {
		return nil, helper.NewUnprocessableEntityError("Available Stock Cannot Be Empty!", helper.ErrorDetail{Detail: "Available Stock is required!"})
	}

	if request.Title == nil || *request.Title == "" {
		return nil, helper.NewUnprocessableEntityError("Book Title Cannot Be Empty!", helper.ErrorDetail{Detail: "Title is required"})
	}

	if request.Category == nil || *request.Category == "" {
		return nil, helper.NewUnprocessableEntityError("Book Category Cannot Be Empty!", helper.ErrorDetail{Detail: "Category is required"})
	}

	if request.PublishedDate == nil || *request.PublishedDate == "" {
		return nil, helper.NewUnprocessableEntityError("Book Published Date Cannot Be Empty!", helper.ErrorDetail{Detail: "Published Date is required"})
	}

	if request.Page == nil || *request.Page <= 0 {
		return nil, helper.NewUnprocessableEntityError("Book Page Cannot Be Empty!", helper.ErrorDetail{Detail: "Page is required"})
	}

	if request.Authors == nil || *request.Authors == "" {
		return nil, helper.NewUnprocessableEntityError("Book Authors Cannot Be Empty!", helper.ErrorDetail{Detail: "Authors is required"})
	}

	if request.Publisher == nil || *request.Publisher == "" {
		return nil, helper.NewUnprocessableEntityError("Book Publisher Cannot Be Empty!", helper.ErrorDetail{Detail: "Publisher is required"})
	}

	if request.Description == nil || *request.Description == "" {
		return nil, helper.NewUnprocessableEntityError("Book Description Cannot Be Empty!", helper.ErrorDetail{Detail: "Description is required"})
	}

	if request.CoverURL == nil || *request.CoverURL == "" {
		return nil, helper.NewUnprocessableEntityError("Book Cover URL Cannot Be Empty!", helper.ErrorDetail{Detail: "Cover URL is required"})
	}

	book.ISBN = *request.ISBN
	book.Title = *request.Title
	book.Authors = *request.Authors
	book.Publisher = *request.Publisher
	book.PublishedDate = *request.PublishedDate
	book.Page = *request.Page
	book.Description = *request.Description
	book.CoverURL = *request.CoverURL
	book.Category = *request.Category
	book.TotalStock = *request.TotalStock
	book.AvailableStock = *request.AvailableStock
	book.UpdatedAt = time.Now()

	errUpdate := u.repository.Update(ctx, book, ID)

	if errUpdate != nil {
		return nil, helper.NewInternalServerError("An Error During Update Book!", helper.ErrorDetail{Detail: errUpdate.Error()})
	}

	errESUpdate := u.elasticsearch.UpdateInElasticsearch(ctx, book)

	if errESUpdate != nil {
		return nil, helper.NewInternalServerError("An Error During Update Book In Elasticsearch!", helper.ErrorDetail{Detail: errESUpdate.Error()})
	}

	result := &dto.BookResponse{
		ID:             &book.ID,
		ISBN:           &book.ISBN,
		Title:          &book.Title,
		Authors:        &book.Authors,
		Publisher:      &book.Publisher,
		PublishedDate:  &book.PublishedDate,
		Page:           &book.Page,
		Description:    &book.Description,
		CoverURL:       &book.CoverURL,
		Category:       &book.Category,
		TotalStock:     &book.TotalStock,
		AvailableStock: &book.AvailableStock,
		CreatedAt:      helper.FormatTimeRFC3339Jakarta(book.CreatedAt),
		UpdatedAt:      helper.FormatTimeRFC3339Jakarta(book.UpdatedAt),
	}

	return result, nil

}
