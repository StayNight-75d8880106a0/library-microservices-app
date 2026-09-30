package usecase

import (
	"activity-logbook-services/internal/dto"
	"activity-logbook-services/internal/helper"
	"activity-logbook-services/internal/models"
	"activity-logbook-services/internal/repository"
	"context"
	"errors"

	"gorm.io/gorm"
)

type LogbookUsecaseInterface interface {
	GetALL(ctx context.Context, limit int, cursor string) ([]dto.LogbookResponseAll, helper.PaginationMeta, error)
	GetByTraceID(ctx context.Context, traceID string) ([]dto.LogbookResponseByTraceID, error)
	GetByID(ctx context.Context, ID string, traceID string) (*dto.LogbookResponseByID, error)
	Create(ctx context.Context, request *dto.LogbookRequest) error
}

type LogbookUsecase struct {
	repository repository.LogbookRepositoryInterface
}

func NewLogbookUsecase(logbookRepository repository.LogbookRepositoryInterface) *LogbookUsecase {
	return &LogbookUsecase{
		repository: logbookRepository,
	}
}

func (u *LogbookUsecase) GetALL(ctx context.Context, limit int, cursor string) ([]dto.LogbookResponseAll, helper.PaginationMeta, error) {

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	cursorOccurredAt, cursorID, errCursor := helper.DecodeCursor(cursor)

	if errCursor != nil {
		return nil, helper.PaginationMeta{}, helper.NewBadRequestError("Invalid Cursor!", helper.ErrorDetail{Detail: errCursor.Error()})
	}

	logbooks, errGet := u.repository.GetAll(ctx, limit+1, cursorOccurredAt, cursorID)
	if errGet != nil {
		return nil, helper.PaginationMeta{}, helper.NewInternalServerError("An Error During Get All Logbook!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	hasNext := len(logbooks) > limit

	if hasNext {
		logbooks = logbooks[:limit]
	}

	result := make([]dto.LogbookResponseAll, 0, len(logbooks))

	for _, value := range logbooks {
		result = append(result, dto.LogbookResponseAll{
			TraceID:    *value.TraceID,
			UserID:     value.UserID,
			OccurredAt: helper.FormatTimeRFC3339Jakarta(value.OccurredAt),
		})
	}

	var nextCursor *string

	if hasNext && len(logbooks) > 0 {
		lastData := logbooks[len(logbooks)-1]
		encoded := helper.EncodeCursor(lastData.OccurredAt, *lastData.ID)
		nextCursor = &encoded
	}

	pagination := helper.PaginationMeta{
		Limit:      limit,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}

	return result, pagination, nil
}

func (u *LogbookUsecase) GetByTraceID(ctx context.Context, traceID string) ([]dto.LogbookResponseByTraceID, error) {

	logbook, errGet := u.repository.GetByTraceID(ctx, traceID)

	if errGet != nil {
		return nil, helper.NewInternalServerError("An Error During Get Logbook By TraceID!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	if len(logbook) == 0 {
		return nil, helper.NewNotFoundError("Logbook Not Found!", helper.ErrorDetail{Detail: "Logbook Not Found!"})
	}

	result := make([]dto.LogbookResponseByTraceID, 0, len(logbook))

	for _, value := range logbook {
		result = append(result, dto.LogbookResponseByTraceID{
			ID:              value.ID,
			ServiceName:     value.ServiceName,
			Method:          value.Method,
			Endpoint:        value.Endpoint,
			HTTPStatus:      value.HTTPStatus,
			HTTPCode:        value.HTTPCode,
			Kind:            value.Kind,
			ExecutionTimeMs: value.ExecutionTimeMs,
			OccurredAt:      helper.FormatTimeRFC3339Jakarta(value.OccurredAt),
		})
	}

	return result, nil

}

func (u *LogbookUsecase) GetByID(ctx context.Context, ID string, traceID string) (*dto.LogbookResponseByID, error) {

	logbook, errGet := u.repository.GetByID(ctx, ID, traceID)

	if errGet != nil {
		if errors.Is(errGet, gorm.ErrRecordNotFound) {
			return nil, helper.NewNotFoundError("Logbook Not Found!", helper.ErrorDetail{Detail: errGet.Error()})
		}
		return nil, helper.NewInternalServerError("An Error During Get Logbook By ID!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	result := &dto.LogbookResponseByID{
		ID:              logbook.ID,
		EventID:         logbook.EventID,
		TraceID:         logbook.TraceID,
		UserID:          logbook.UserID,
		ServiceName:     logbook.ServiceName,
		Method:          logbook.Method,
		Endpoint:        logbook.Endpoint,
		HTTPStatus:      logbook.HTTPStatus,
		HTTPCode:        logbook.HTTPCode,
		Kind:            logbook.Kind,
		IPAddress:       logbook.IPAddress,
		RequestBody:     logbook.RequestBody,
		ResponseBody:    logbook.ResponseBody,
		ExecutionTimeMs: logbook.ExecutionTimeMs,
		OccurredAt:      helper.FormatTimeRFC3339Jakarta(logbook.OccurredAt),
		CreatedAt:       helper.FormatTimeRFC3339Jakarta(logbook.CreatedAt),
	}

	return result, nil

}

func (u *LogbookUsecase) Create(ctx context.Context, request *dto.LogbookRequest) error {

	if request.OccurredAt.IsZero() {
		return helper.NewUnprocessableEntityError("Invalid Logbook Event!", helper.ErrorDetail{Detail: "occurredAt is required"})
	}

	logbook := &models.Logbook{
		EventID:         request.EventID,
		TraceID:         request.TraceID,
		UserID:          request.UserID,
		ServiceName:     request.ServiceName,
		Method:          request.Method,
		Endpoint:        request.Endpoint,
		HTTPStatus:      request.HTTPStatus,
		HTTPCode:        request.HTTPCode,
		Kind:            request.Kind,
		IPAddress:       request.IPAddress,
		RequestBody:     request.RequestBody,
		ResponseBody:    request.ResponseBody,
		ExecutionTimeMs: request.ExecutionTimeMs,
		IsRoot:          request.IsRoot,
		OccurredAt:      request.OccurredAt,
	}

	errCreate := u.repository.Create(ctx, logbook)

	if errCreate != nil {
		return helper.NewInternalServerError("An Error During Create Logbook!", helper.ErrorDetail{Detail: errCreate.Error()})
	}

	return nil

}
