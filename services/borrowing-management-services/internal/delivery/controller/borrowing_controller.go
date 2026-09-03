package controller

import (
	"borrowing-management-services/internal/dto"
	"borrowing-management-services/internal/helper"
	"borrowing-management-services/internal/usecase"
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type BorrowingController struct {
	usecase usecase.BorrowingUsecaseInterface
}

func NewBorrowingController(borrowingUsecase usecase.BorrowingUsecaseInterface) *BorrowingController {
	return &BorrowingController{
		usecase: borrowingUsecase,
	}
}

func (c *BorrowingController) Create(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	userID := ctx.GetString("userID")

	request := new(dto.CreateBorrowingRequest)

	errRequest := ctx.ShouldBindJSON(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	borrowing, errCreate := c.usecase.CreateBorrowing(contextVariable, request, userID)

	if errCreate != nil {
		helper.NewErrorResponse(ctx, errCreate)
		return
	}

	helper.NewResponseGlobal(ctx, 201, "Success Create Borrowing!", borrowing, nil, nil)

}

func (c *BorrowingController) GetALL(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	userID := ctx.GetString("userID")
	isAdmin := ctx.GetBool("isAdmin")

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	borrowings, pagination, errGet := c.usecase.GetAllBorrowings(contextVariable, limit, page, isAdmin, userID)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get Borrowing!", borrowings, nil, pagination)

}

func (c *BorrowingController) GetMyByID(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	userID := ctx.GetString("userID")
	ID := ctx.Param("id")

	borrowing, errGet := c.usecase.GetMyBorrowingByID(contextVariable, ID, userID)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get Borrowing!", borrowing, nil, nil)

}

func (c *BorrowingController) GetByID(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	ID := ctx.Param("id")

	borrowing, errGet := c.usecase.GetBorrowingByID(contextVariable, ID)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get Borrowing!", borrowing, nil, nil)

}

func (c *BorrowingController) UpdateStatus(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	ID := ctx.Param("id")

	request := new(dto.BorrowingUpdateRequest)

	errRequest := ctx.ShouldBindJSON(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	errUpdate := c.usecase.UpdateBorrowingStatus(contextVariable, ID, request)

	if errUpdate != nil {
		helper.NewErrorResponse(ctx, errUpdate)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Update Borrowing Status!", nil, nil, nil)

}
