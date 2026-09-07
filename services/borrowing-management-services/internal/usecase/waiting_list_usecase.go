package usecase

import (
	"borrowing-management-services/internal/dto"
	"borrowing-management-services/internal/helper"
	"borrowing-management-services/internal/models"
	"borrowing-management-services/internal/repository"
	"context"
	"math"

	"gorm.io/gorm"
)

type WaitingListUsecaseInterface interface {
	JoinWaitingList(ctx context.Context, userID string, request *dto.CreateWaitingListRequest) (*dto.WaitingListResponse, error)
	GetAllWaitingLists(ctx context.Context, limit int, page int, role bool, userID string) ([]dto.WaitingListResponse, helper.PaginationMeta, error)
	GetWaitingListByID(ctx context.Context, ID string, role bool, userID string) (*dto.WaitingListResponse, error)
}

type WaitingListUsecase struct {
	repository repository.WaitingListRepositoryInterface
}

func NewWaitingListUsecase(waitingListRepository repository.WaitingListRepositoryInterface) *WaitingListUsecase {
	return &WaitingListUsecase{
		repository: waitingListRepository,
	}
}

func (u *WaitingListUsecase) JoinWaitingList(ctx context.Context, userID string, request *dto.CreateWaitingListRequest) (*dto.WaitingListResponse, error) {

	if request.BookID == nil || *request.BookID == "" {
		return nil, helper.NewUnprocessableEntityError("Book Id Cannot Be Empty!", helper.ErrorDetail{Detail: "Book Is Required!"})
	}

	lambda := 2.0
	mu := 1.0 / 168.0
	numServers := 1

	erlangRes := helper.CalculateErlangC(lambda, mu, numServers)

	optimalServers := erlangRes.OptimalServers

	waitingTerm := &models.WaitingList{
		UserID:         userID,
		BookID:         *request.BookID,
		ArrivalRate:    lambda,
		ServiceRate:    mu,
		NumServers:     numServers,
		Utilization:    erlangRes.Utilization,
		ProbWait:       erlangRes.ProbWait,
		AvgQueueLen:    erlangRes.AvgQueueLen,
		AvgWaitMin:     erlangRes.AvgWaitMin,
		OptimalServers: &optimalServers,
		Status:         models.WaitingListStatusWaiting,
	}

	errCreate := u.repository.Create(ctx, waitingTerm)

	if errCreate != nil {
		return nil, helper.NewInternalServerError("Failed to create waiting list", helper.ErrorDetail{Detail: errCreate.Error()})
	}

	result := &dto.WaitingListResponse{
		ID:             waitingTerm.ID,
		UserID:         waitingTerm.UserID,
		BookID:         waitingTerm.BookID,
		ArrivalRate:    waitingTerm.ArrivalRate,
		ServiceRate:    waitingTerm.ServiceRate,
		NumServers:     waitingTerm.NumServers,
		Utilization:    waitingTerm.Utilization,
		ProbWait:       waitingTerm.ProbWait,
		AvgQueueLen:    waitingTerm.AvgQueueLen,
		AvgWaitMin:     waitingTerm.AvgWaitMin,
		OptimalServers: waitingTerm.OptimalServers,
		QueueNumber:    waitingTerm.QueueNumber,
		Status:         string(waitingTerm.Status),
		CreatedAt:      helper.FormatTimeRFC3339Jakarta(waitingTerm.CreatedAt),
		UpdatedAt:      helper.FormatTimeRFC3339Jakarta(waitingTerm.UpdatedAt),
	}

	return result, nil

}

func (u *WaitingListUsecase) GetAllWaitingLists(ctx context.Context, limit int, page int, role bool, userID string) ([]dto.WaitingListResponse, helper.PaginationMeta, error) {

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	if role {
		waitingLists, totalData, errGet := u.repository.GetALLWaitingList(ctx, limit, offset)

		if errGet != nil {
			return nil, helper.PaginationMeta{}, helper.NewInternalServerError("Failed to get waiting lists", helper.ErrorDetail{Detail: errGet.Error()})
		}

		result := make([]dto.WaitingListResponse, 0, len(waitingLists))

		for _, waitingList := range waitingLists {
			result = append(result, dto.WaitingListResponse{
				ID:             waitingList.ID,
				UserID:         waitingList.UserID,
				BookID:         waitingList.BookID,
				ArrivalRate:    waitingList.ArrivalRate,
				ServiceRate:    waitingList.ServiceRate,
				NumServers:     waitingList.NumServers,
				Utilization:    waitingList.Utilization,
				ProbWait:       waitingList.ProbWait,
				AvgQueueLen:    waitingList.AvgQueueLen,
				AvgWaitMin:     waitingList.AvgWaitMin,
				OptimalServers: waitingList.OptimalServers,
				QueueNumber:    waitingList.QueueNumber,
				Status:         string(waitingList.Status),
				CreatedAt:      helper.FormatTimeRFC3339Jakarta(waitingList.CreatedAt),
				UpdatedAt:      helper.FormatTimeRFC3339Jakarta(waitingList.UpdatedAt),
			})
		}

		totalPage := int(math.Ceil(float64(totalData) / float64(limit)))

		pagination := helper.PaginationMeta{
			TotalData: totalData,
			TotalPage: totalPage,
			Page:      page,
			Limit:     limit,
			Keywords:  nil,
		}

		return result, pagination, nil

	}

	waitingLists, totalData, errGet := u.repository.GetALLMyWaitingList(ctx, limit, offset, userID)

	if errGet != nil {
		return nil, helper.PaginationMeta{}, helper.NewInternalServerError("Failed to get waiting lists", helper.ErrorDetail{Detail: errGet.Error()})
	}

	result := make([]dto.WaitingListResponse, 0, len(waitingLists))

	for _, waitingList := range waitingLists {
		result = append(result, dto.WaitingListResponse{
			ID:             waitingList.ID,
			UserID:         waitingList.UserID,
			BookID:         waitingList.BookID,
			ArrivalRate:    waitingList.ArrivalRate,
			ServiceRate:    waitingList.ServiceRate,
			NumServers:     waitingList.NumServers,
			Utilization:    waitingList.Utilization,
			ProbWait:       waitingList.ProbWait,
			AvgQueueLen:    waitingList.AvgQueueLen,
			AvgWaitMin:     waitingList.AvgWaitMin,
			OptimalServers: waitingList.OptimalServers,
			QueueNumber:    waitingList.QueueNumber,
			Status:         string(waitingList.Status),
			CreatedAt:      helper.FormatTimeRFC3339Jakarta(waitingList.CreatedAt),
			UpdatedAt:      helper.FormatTimeRFC3339Jakarta(waitingList.UpdatedAt),
		})
	}

	totalPage := int(math.Ceil(float64(totalData) / float64(limit)))

	pagination := helper.PaginationMeta{
		TotalData: totalData,
		TotalPage: totalPage,
		Page:      page,
		Limit:     limit,
		Keywords:  nil,
	}

	return result, pagination, nil

}

func (u *WaitingListUsecase) GetWaitingListByID(ctx context.Context, ID string, role bool, userID string) (*dto.WaitingListResponse, error) {

	waitingList, errGet := u.repository.GetWaitingListByID(ctx, ID)

	if errGet != nil {
		if errGet == gorm.ErrRecordNotFound {
			return nil, helper.NewNotFoundError("Waiting List Not Found!", helper.ErrorDetail{Detail: "Waiting List with the given ID does not exist!"})
		}
		return nil, helper.NewInternalServerError("An Error During Get Waiting List By ID!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	if role {
		result := &dto.WaitingListResponse{
			ID:             waitingList.ID,
			UserID:         waitingList.UserID,
			BookID:         waitingList.BookID,
			ArrivalRate:    waitingList.ArrivalRate,
			ServiceRate:    waitingList.ServiceRate,
			NumServers:     waitingList.NumServers,
			Utilization:    waitingList.Utilization,
			ProbWait:       waitingList.ProbWait,
			AvgQueueLen:    waitingList.AvgQueueLen,
			AvgWaitMin:     waitingList.AvgWaitMin,
			OptimalServers: waitingList.OptimalServers,
			QueueNumber:    waitingList.QueueNumber,
			Status:         string(waitingList.Status),
			CreatedAt:      helper.FormatTimeRFC3339Jakarta(waitingList.CreatedAt),
			UpdatedAt:      helper.FormatTimeRFC3339Jakarta(waitingList.UpdatedAt),
		}

		return result, nil
	}

	if waitingList.UserID != userID {
		return nil, helper.NewForbiddenError("Access Denied!", helper.ErrorDetail{Detail: "You do not have access to this Waiting List!"})
	}

	result := &dto.WaitingListResponse{
		ID:             waitingList.ID,
		UserID:         waitingList.UserID,
		BookID:         waitingList.BookID,
		ArrivalRate:    waitingList.ArrivalRate,
		ServiceRate:    waitingList.ServiceRate,
		NumServers:     waitingList.NumServers,
		Utilization:    waitingList.Utilization,
		ProbWait:       waitingList.ProbWait,
		AvgQueueLen:    waitingList.AvgQueueLen,
		AvgWaitMin:     waitingList.AvgWaitMin,
		OptimalServers: waitingList.OptimalServers,
		QueueNumber:    waitingList.QueueNumber,
		Status:         string(waitingList.Status),
		CreatedAt:      helper.FormatTimeRFC3339Jakarta(waitingList.CreatedAt),
		UpdatedAt:      helper.FormatTimeRFC3339Jakarta(waitingList.UpdatedAt),
	}

	return result, nil

}
