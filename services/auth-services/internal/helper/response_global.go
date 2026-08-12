package helper

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseGlobal struct {
	ReponseCode     int         `json:"responseCode"`
	ResponseMessage string      `json:"responseMessage"`
	Data            interface{} `json:"data"`
	Error           interface{} `json:"error"`
}

func NewResponseGlobal(ctx *gin.Context, responseCode int, responseMessage string, data interface{}, err interface{}) {
	ctx.JSON(responseCode, ResponseGlobal{
		ReponseCode:     responseCode,
		ResponseMessage: responseMessage,
		Data:            data,
		Error:           err,
	})
}

func NewErrorResponse(ctx *gin.Context, err error) {

	var appErr *AppError
	if errors.As(err, &appErr) {
		NewResponseGlobal(ctx, appErr.Code, appErr.ErrorMessage, nil, appErr.Detail)
		return
	}

	log.Printf("[ERROR] %v", err)
	NewResponseGlobal(ctx, http.StatusInternalServerError, "An error has occurred on the server!", nil, nil)
}
