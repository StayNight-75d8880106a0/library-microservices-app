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

type WaitingListController struct {
	usecase usecase.WaitingListUsecaseInterface
}

func NewWaitingListController(waitingListUsecase usecase.WaitingListUsecaseInterface) *WaitingListController {
	return &WaitingListController{
		usecase: waitingListUsecase,
	}
}

func (c *WaitingListController) Create(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 11*time.Second)
	defer cancel()

	userID := ctx.GetString("userID")

	request := new(dto.CreateWaitingListRequest)

	errRequest := ctx.ShouldBindJSON(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	waitingList, errCreate := c.usecase.JoinWaitingList(contextVariable, userID, request)

	if errCreate != nil {
		helper.NewErrorResponse(ctx, errCreate)
		return
	}

	helper.NewResponseGlobal(ctx, 201, "Success Waiting For Book!", waitingList, nil, nil)

}

func (c *WaitingListController) GetAll(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	userID := ctx.GetString("userID")
	isAdmin := ctx.GetBool("isAdmin")

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	waitingLists, pagination, errGet := c.usecase.GetAllWaitingLists(contextVariable, limit, page, isAdmin, userID)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get All Waiting Lists", waitingLists, nil, pagination)
}

func (c *WaitingListController) GetByID(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	ID := ctx.Param("id")
	userID := ctx.GetString("userID")
	isAdmin := ctx.GetBool("isAdmin")

	waitingList, errGet := c.usecase.GetWaitingListByID(contextVariable, ID, isAdmin, userID)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get Detail Waiting Lists", waitingList, nil, nil)

}

func (c *WaitingListController) Cancel(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	ID := ctx.Param("id")
	userID := ctx.GetString("userID")

	errUpdate := c.usecase.CancelWaitingList(contextVariable, ID, userID)

	if errUpdate != nil {
		helper.NewErrorResponse(ctx, errUpdate)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Cancel Waiting List", nil, nil, nil)

}
