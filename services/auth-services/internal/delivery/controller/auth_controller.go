package controller

import (
	"auth-services/internal/dto"
	"auth-services/internal/helper"
	"auth-services/internal/usecase"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthController struct {
	usecase usecase.AuthUsecaseInterface
}

func NewAuthControllerRegistry(authUsecase usecase.AuthUsecaseInterface) *AuthController {
	return &AuthController{
		usecase: authUsecase,
	}
}

func (c *AuthController) Login(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	request := new(dto.LoginRequest)

	errRequest := ctx.ShouldBindJSON(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	login, err := c.usecase.Login(contextVariable, request)

	if err != nil {
		if errors.Is(contextVariable.Err(), context.DeadlineExceeded) {
			helper.NewErrorResponse(ctx, helper.NewStatusGatewayTimeoutError("Request timeout, upstream service took too long!", helper.ErrorDetail{Detail: contextVariable.Err().Error()}))
			return
		}
		helper.NewErrorResponse(ctx, err)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Login!", login, nil)

}

func (c *AuthController) Register(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	request := new(dto.RegisterUser)

	errRequest := ctx.ShouldBindJSON(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	err := c.usecase.RegisterUser(contextVariable, request)

	if err != nil {
		if errors.Is(contextVariable.Err(), context.DeadlineExceeded) {
			helper.NewErrorResponse(ctx, helper.NewStatusGatewayTimeoutError("Request timeout, upstream service took too long!", helper.ErrorDetail{Detail: contextVariable.Err().Error()}))
			return
		}
		helper.NewErrorResponse(ctx, err)
		return
	}

	helper.NewResponseGlobal(ctx, 201, "Success Create User!", nil, nil)

}

func (c *AuthController) Logout(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
	defer cancel()

	request := new(dto.RefreshTokenRequest)

	errRequest := ctx.ShouldBindJSON(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	authHeader := ctx.GetHeader("Authorization")

	token := strings.Replace(authHeader, "Bearer ", "", 1)

	err := c.usecase.Logout(contextVariable, token, request)

	if err != nil {
		if errors.Is(contextVariable.Err(), context.DeadlineExceeded) {
			helper.NewErrorResponse(ctx, helper.NewStatusGatewayTimeoutError("Request timeout, upstream service took too long!", helper.ErrorDetail{Detail: contextVariable.Err().Error()}))
			return
		}
		helper.NewErrorResponse(ctx, err)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Logout!", nil, nil)

}

func (c *AuthController) Me(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
	defer cancel()

	claims, exists := ctx.Get("claims")

	if !exists {
		helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Unauthorized!", helper.ErrorDetail{Detail: "Claims not found!"}))
		return
	}

	mapClaims, ok := claims.(jwt.MapClaims)

	if !ok {
		helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Unauthorized!", helper.ErrorDetail{Detail: "Invalid claims format!"}))
		return
	}

	profile, err := c.usecase.Me(contextVariable, mapClaims)

	if err != nil {
		if errors.Is(contextVariable.Err(), context.DeadlineExceeded) {
			helper.NewErrorResponse(ctx, helper.NewStatusGatewayTimeoutError("Request timeout, upstream service took too long!", helper.ErrorDetail{Detail: contextVariable.Err().Error()}))
			return
		}
		helper.NewErrorResponse(ctx, err)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get Profile!", profile, nil)
}

func (c *AuthController) RefreshToken(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	request := new(dto.RefreshTokenRequest)

	errRequest := ctx.ShouldBindJSON(request)

	if errRequest != nil {
		helper.NewErrorResponse(ctx, errRequest)
		return
	}

	refresh, err := c.usecase.RefreshToken(contextVariable, request)

	if err != nil {
		if errors.Is(contextVariable.Err(), context.DeadlineExceeded) {
			helper.NewErrorResponse(ctx, helper.NewStatusGatewayTimeoutError("Request timeout, upstream service took too long!", helper.ErrorDetail{Detail: contextVariable.Err().Error()}))
			return
		}
		helper.NewErrorResponse(ctx, err)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Refresh Token!", refresh, nil)
}
