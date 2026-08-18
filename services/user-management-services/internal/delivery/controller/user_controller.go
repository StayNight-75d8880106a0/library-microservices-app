package controller

import (
	"context"
	"time"
	"user-management-services/internal/dto"
	"user-management-services/internal/helper"
	"user-management-services/internal/usecase"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	usecase usecase.UserProfileUsecaseInterface
}

func NewUserControllerRegistry(userUsecase usecase.UserProfileUsecaseInterface) *UserController {
	return &UserController{
		usecase: userUsecase,
	}
}

func (c *UserController) UpdateMyProfile(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	request := new(dto.UserUpdateProfileRequest)

	errRequest := ctx.ShouldBind(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	userID := ctx.GetString("userID")

	if userID == "" {
		helper.NewErrorResponse(ctx, helper.NewInternalServerError("Missing User ID!", helper.ErrorDetail{Detail: "User ID in Context Is Required!"}))
		return
	}

	errUpdateProfile := c.usecase.UpdateMyProfile(contextVariable, userID, request)

	if errUpdateProfile != nil {
		helper.NewErrorResponse(ctx, errUpdateProfile)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Update My Profile", nil, nil, nil)
}

func (c *UserController) UpdateMyPassword(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	request := new(dto.UserUpdatePasswordRequest)

	errRequest := ctx.ShouldBind(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	userID := ctx.GetString("userID")

	if userID == "" {
		helper.NewErrorResponse(ctx, helper.NewInternalServerError("Missing User ID!", helper.ErrorDetail{Detail: "User ID in Context Is Required!"}))
		return
	}

	errUpdatePassword := c.usecase.UpdateMyPassword(contextVariable, userID, request)

	if errUpdatePassword != nil {
		helper.NewErrorResponse(ctx, errUpdatePassword)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Update My Password", nil, nil, nil)

}

func (c *UserController) GetMyProfile(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	userID := ctx.GetString("userID")

	if userID == "" {
		helper.NewErrorResponse(ctx, helper.NewInternalServerError("Missing User ID!", helper.ErrorDetail{Detail: "User ID in Context Is Required!"}))
		return
	}

	profile, errGetProfile := c.usecase.GetMyProfile(contextVariable, userID)

	if errGetProfile != nil {
		helper.NewErrorResponse(ctx, errGetProfile)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get My Profile", profile, nil, nil)

}
