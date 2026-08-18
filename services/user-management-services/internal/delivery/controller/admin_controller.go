package controller

import (
	"context"
	"strconv"
	"time"
	"user-management-services/internal/dto"
	"user-management-services/internal/helper"
	"user-management-services/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	usecase usecase.AdminUsecaseInterface
}

func NewAdminControllerRegistry(adminUsecase usecase.AdminUsecaseInterface) *AdminController {
	return &AdminController{
		usecase: adminUsecase,
	}
}

func (c *AdminController) CreateAdmin(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	request := new(dto.CreateAdminRequest)

	errRequest := ctx.ShouldBind(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	errCreateAdmin := c.usecase.CreateAdmin(contextVariable, request)

	if errCreateAdmin != nil {
		helper.NewErrorResponse(ctx, errCreateAdmin)
		return
	}

	helper.NewResponseGlobal(ctx, 201, "Success Create Admin", nil, nil, nil)

}

func (c *AdminController) GetAllUsers(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	users, pagination, errGetUsers := c.usecase.GetAllUsers(contextVariable, page, limit)

	if errGetUsers != nil {
		helper.NewErrorResponse(ctx, errGetUsers)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get All Users", users, nil, pagination)

}

func (c *AdminController) UpdateAdmin(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	request := new(dto.UserUpdatePasswordRequest)

	errRequest := ctx.ShouldBind(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	adminID := ctx.Param("adminID")

	if adminID == "" {
		helper.NewErrorResponse(ctx, helper.NewInternalServerError("Missing Admin ID Param!", helper.ErrorDetail{Detail: "Admin ID In Param Is Required!"}))
		return
	}

	errUpdateAdmin := c.usecase.UpdateAdmin(contextVariable, adminID, request)

	if errUpdateAdmin != nil {
		helper.NewErrorResponse(ctx, errUpdateAdmin)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Update Admin", nil, nil, nil)

}

func (c *AdminController) DeleteAdmin(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	adminID := ctx.Param("adminID")

	if adminID == "" {
		helper.NewErrorResponse(ctx, helper.NewInternalServerError("Missing Admin ID Param!", helper.ErrorDetail{Detail: "Admin ID In Param Is Required!"}))
		return
	}

	errDeleteAdmin := c.usecase.DeleteAdmin(contextVariable, adminID)

	if errDeleteAdmin != nil {
		helper.NewErrorResponse(ctx, errDeleteAdmin)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Delete Admin", nil, nil, nil)

}
