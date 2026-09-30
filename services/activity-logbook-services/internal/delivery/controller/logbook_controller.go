package controller

import (
	"activity-logbook-services/internal/helper"
	"activity-logbook-services/internal/usecase"
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type LogbookController struct {
	usecase usecase.LogbookUsecaseInterface
}

func NewLogbookController(logbookUsecase usecase.LogbookUsecaseInterface) *LogbookController {
	return &LogbookController{
		usecase: logbookUsecase,
	}
}

func (c *LogbookController) GetAll(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 31*time.Second)
	defer cancel()

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	cursor := ctx.Query("cursor")

	logbooks, pagination, errGet := c.usecase.GetALL(contextVariable, limit, cursor)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get All Logbook!", logbooks, nil, pagination)

}

func (c *LogbookController) GetByTraceID(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	traceID := ctx.Param("traceID")

	logbooks, errGet := c.usecase.GetByTraceID(contextVariable, traceID)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get Logbook By Trace ID!", logbooks, nil, nil)

}

func (c *LogbookController) GetByID(ctx *gin.Context) {

	contextVariable, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	traceID := ctx.Param("traceID")
	ID := ctx.Param("id")

	logbook, errGet := c.usecase.GetByID(contextVariable, ID, traceID)

	if errGet != nil {
		helper.NewErrorResponse(ctx, errGet)
		return
	}

	helper.NewResponseGlobal(ctx, 200, "Success Get Logbook By ID!", logbook, nil, nil)

}
