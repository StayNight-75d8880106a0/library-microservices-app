package helper

import "net/http"

type AppError struct {
	Code         int         `json:"code"`
	ErrorMessage string      `json:"errorMessage"`
	ErrorCode    string      `json:"errorCode"`
	Detail       ErrorDetail `json:"detail"`
}

type ErrorDetail struct {
	Detail string `json:"detail"`
}

func (e *AppError) Error() string {
	return e.ErrorMessage
}

func NewBadRequestError(message string, details ErrorDetail) *AppError {
	return &AppError{
		Code:         http.StatusBadRequest,
		ErrorCode:    "BAD_REQUEST",
		ErrorMessage: message,
		Detail:       details,
	}
}

func NewInternalServerError(message string, details ErrorDetail) *AppError {
	return &AppError{
		Code:         http.StatusInternalServerError,
		ErrorCode:    "INTERNAL_SERVER_ERROR",
		ErrorMessage: message,
		Detail:       details,
	}
}

func NewUnprocessableEntityError(message string, details ErrorDetail) *AppError {
	return &AppError{
		Code:         http.StatusUnprocessableEntity,
		ErrorCode:    "UNPROCESSABLE_ENTITY",
		ErrorMessage: message,
		Detail:       details,
	}
}

func NewUnauthorizedError(message string, details ErrorDetail) *AppError {
	return &AppError{
		Code:         http.StatusUnauthorized,
		ErrorCode:    "UNAUTHORIZED",
		ErrorMessage: message,
		Detail:       details,
	}
}

func NewStatusGatewayTimeoutError(message string, details ErrorDetail) *AppError {
	return &AppError{
		Code:         http.StatusGatewayTimeout,
		ErrorCode:    "GATEWAY_TIMEOUT",
		ErrorMessage: message,
		Detail:       details,
	}
}

func NewServiceUnavailableError(message string, details ErrorDetail) *AppError {
	return &AppError{
		Code:         http.StatusServiceUnavailable,
		ErrorCode:    "SERVICE_UNAVAILABLE",
		ErrorMessage: message,
		Detail:       details,
	}
}
