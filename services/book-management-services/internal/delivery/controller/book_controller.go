package controller

import (
	"book-management-services/internal/dto"
	"book-management-services/internal/helper"
	"book-management-services/internal/usecase"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type BookController struct {
	usecase usecase.BookUsecaseInterface
}

func NewBookController(bookUsecase usecase.BookUsecaseInterface) *BookController {
	return &BookController{
		usecase: bookUsecase,
	}
}

func (c *BookController) Create(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	request := new(dto.CreateBookRequest)

	errRequest := ctx.ShouldBind(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, helper.NewBadRequestError("Invalid request body", helper.ErrorDetail{Detail: errRequest.Error()}))
		return
	}

	book, errCreate := c.usecase.CreateBook(contextVariable, request)

	if errCreate != nil {
		helper.NewErrorResponse(ctx, errCreate)
		return
	}

	helper.NewResponseGlobal(ctx, 201, "Success Create Book!", book, nil, nil)

}

func (c *BookController) GetAll(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	query := new(dto.GetBooksQuery)

	errQuery := ctx.ShouldBindQuery(query)

	if errQuery != nil {
		helper.NewErrorResponse(ctx, helper.NewBadRequestError("Invalid query parameter", helper.ErrorDetail{Detail: errQuery.Error()}))
		return
	}

	books, pagination, errGet := c.usecase.GetAllBooks(contextVariable, *query)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get All Books!", books, nil, pagination)
}

func (c *BookController) GetByID(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	ID := ctx.Param("id")

	book, errGet := c.usecase.GetBookByID(contextVariable, ID)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get Book By ID!", book, nil, nil)

}

func (c *BookController) Delete(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	ID := ctx.Param("id")

	errDelete := c.usecase.DeleteBook(contextVariable, ID)

	if errDelete != nil {
		helper.NewErrorResponse(ctx, errDelete)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Delete Book!", nil, nil, nil)

}

func (c *BookController) Update(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	ID := ctx.Param("id")

	request := new(dto.UpdateBookRequest)

	errRequest := ctx.ShouldBind(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, helper.NewBadRequestError("Invalid request body", helper.ErrorDetail{Detail: errRequest.Error()}))
		return
	}

	book, errUpdate := c.usecase.UpdateBook(contextVariable, request, ID)

	if errUpdate != nil {
		helper.NewErrorResponse(ctx, errUpdate)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Update Book!", book, nil, nil)

}
