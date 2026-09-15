package controller

import (
	"context"
	"notification-services/internal/helper"
	"notification-services/internal/usecase"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type EmailController struct {
	usecase usecase.EmailUsecaseInterface
}

func NewEmailController(emailUsecase usecase.EmailUsecaseInterface) *EmailController {
	return &EmailController{
		usecase: emailUsecase,
	}
}

func (c *EmailController) GetAllEmailLogs(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	emailLogs, pagination, errGet := c.usecase.GetAllEmailLogs(contextVariable, page, limit)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get All Email Logs!", emailLogs, nil, pagination)

}

func (c *EmailController) GetEmailLogByID(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	ID := ctx.Param("id")

	emailLog, errGet := c.usecase.GetEmailLogByID(contextVariable, ID)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get All Email Logs!", emailLog, nil, nil)

}
