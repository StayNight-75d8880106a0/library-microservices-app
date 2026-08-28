package helper

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaginationMeta struct {
	TotalData int64   `json:"totalData"`
	TotalPage int     `json:"totalPage"`
	Page      int     `json:"page"`
	Limit     int     `json:"limit"`
	Keywords  *string `json:"keywords"`
}

type ResponseGlobal struct {
	ResponseCode    int         `json:"responseCode"`
	ResponseMessage string      `json:"responseMessage"`
	Data            interface{} `json:"data"`
	Error           interface{} `json:"error"`
	Pagination      interface{} `json:"pagination"`
}

func NewResponseGlobal(ctx *gin.Context, responseCode int, responseMessage string, data interface{}, err interface{}, pagination interface{}) {
	ctx.JSON(responseCode, ResponseGlobal{
		ResponseCode:    responseCode,
		ResponseMessage: responseMessage,
		Data:            data,
		Error:           err,
		Pagination:      pagination,
	})
}

func NewErrorResponse(ctx *gin.Context, err error) {

	var appErr *AppError
	if errors.As(err, &appErr) {
		NewResponseGlobal(ctx, appErr.Code, appErr.ErrorMessage, nil, appErr.Detail, nil)
		return
	}

	log.Printf("[ERROR] %v", err)
	NewResponseGlobal(ctx, http.StatusInternalServerError, "An error has occurred on the server!", nil, nil, nil)
}
