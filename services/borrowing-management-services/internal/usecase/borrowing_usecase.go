package usecase

import (
	"borrowing-management-services/internal/client"
	"borrowing-management-services/internal/config"
	"borrowing-management-services/internal/dto"
	"borrowing-management-services/internal/helper"
	"borrowing-management-services/internal/infrastructure/kafka/event"
	"borrowing-management-services/internal/infrastructure/kafka/producer"
	"borrowing-management-services/internal/models"
	"borrowing-management-services/internal/repository"
	"context"
	"log"
	"math"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type BorrowingUsecaseInterface interface {
	CreateBorrowing(ctx context.Context, request *dto.CreateBorrowingRequest, userID string) (*dto.BorrowingResponse, error)
	GetAllBorrowings(ctx context.Context, limit int, page int, role bool, userID string) ([]dto.BorrowingResponse, helper.PaginationMeta, error)
	GetBorrowingByID(ctx context.Context, ID string, role bool, userID string) (*dto.BorrowingResponse, error)
	UpdateBorrowingStatus(ctx context.Context, ID string, request *dto.BorrowingUpdateRequest) error
}

type BorrowingUsecase struct {
	repository            repository.BorrowingUserRepositoryInterface
	userCache             repository.UserRedisCacheInterface
	kafkaProducer         *producer.KafkaProducer
	cfg                   *config.AppConfig
	userGrpc              client.UserGrpcClientInterface
	waitingListRepository repository.WaitingListRepositoryInterface
}

func NewBorrowingUsecase(borrowingRepository repository.BorrowingUserRepositoryInterface, userCache repository.UserRedisCacheInterface, kafkaProducer *producer.KafkaProducer, cfg *config.AppConfig, userGrpc client.UserGrpcClientInterface, waitingListRepository repository.WaitingListRepositoryInterface) *BorrowingUsecase {
	return &BorrowingUsecase{
		repository:            borrowingRepository,
		userCache:             userCache,
		kafkaProducer:         kafkaProducer,
		cfg:                   cfg,
		userGrpc:              userGrpc,
		waitingListRepository: waitingListRepository,
	}
}

func (u *BorrowingUsecase) CreateBorrowing(ctx context.Context, request *dto.CreateBorrowingRequest, userID string) (*dto.BorrowingResponse, error) {

	if request.BookID == nil || *request.BookID == "" {
		return nil, helper.NewUnprocessableEntityError("Book Cannot Be Empty!", helper.ErrorDetail{Detail: "Book ID is required!"})
	}

	grpcStatus, errGrpc := u.userGrpc.GetUserStatus(ctx, userID)

	if errGrpc != nil {
		switch status.Code(errGrpc) {

		case codes.NotFound:
			return nil, helper.NewForbiddenError(
				"Account Not Registered!",
				helper.ErrorDetail{Detail: "Your account has not been registered as a library member, which automatically invalidates your claim to access the lending facilities."},
			)

		case codes.Unavailable, codes.DeadlineExceeded:
			return nil, helper.NewServiceUnavailableError(
				"User Service Unavailable!",
				helper.ErrorDetail{Detail: "Cannot verify your account status right now. Please try again in a moment."},
			)

		default:
			log.Printf("[gRPC Error] get user status %s: %v", userID, errGrpc)
			return nil, helper.NewInternalServerError(
				"An Error During Get User Status!",
				helper.ErrorDetail{Detail: errGrpc.Error()},
			)
		}
	}

	if strings.ToUpper(grpcStatus) == "" {
		return nil, helper.NewForbiddenError("User Status Not Found!", helper.ErrorDetail{Detail: "Your account status is not found. Please contact the administrator!"})
	}

	if strings.ToUpper(grpcStatus) != "ACTIVE" {
		return nil, helper.NewForbiddenError("User Status Not Active!", helper.ErrorDetail{Detail: "Your account status is INCOMPLETED. Please fill in and complete your profile details so that you can borrow books!"})
	}

	var calculateDueDate time.Time

	if request.Duedate == nil || request.Duedate.IsZero() {
		calculateDueDate = time.Now().AddDate(0, 0, 14)
	} else {
		userDueDate := *request.Duedate

		if userDueDate.Before(time.Now()) {
			return nil, helper.NewUnprocessableEntityError("Due Date Cannot Be Before Today!", helper.ErrorDetail{Detail: "Due Date must be after today!"})
		}

		maxAllowedDueDate := time.Now().AddDate(0, 0, 30)
		if userDueDate.After(maxAllowedDueDate) {
			return nil, helper.NewUnprocessableEntityError("Due Date Cannot Be More Than 30 Days!", helper.ErrorDetail{Detail: "Due Date must be within 30 days from today!"})
		}

		calculateDueDate = userDueDate
	}

	claimed, errClaim := u.waitingListRepository.MarkFulfilledWaitingList(ctx, userID, *request.BookID)

	if errClaim != nil {
		return nil, helper.NewInternalServerError("An Error During Mark Fulfilled Waiting List!", helper.ErrorDetail{Detail: errClaim.Error()})
	}

	if claimed == 0 {
		notifiedCount, errCount := u.waitingListRepository.CountNotifiedWaitingListsByBookID(ctx, *request.BookID)

		if errCount != nil {
			return nil, helper.NewInternalServerError("An Error During Count Notified Waiting Lists By Book ID!", helper.ErrorDetail{Detail: errCount.Error()})
		}

		if notifiedCount > 0 {
			return nil, helper.NewForbiddenError("Waiting List Not Fulfilled!", helper.ErrorDetail{Detail: "You have not fulfilled your waiting list for this book. Please wait for your turn to borrow the book!"})
		}
	}

	borrowing := &models.Borrowing{
		UserID:     userID,
		BookID:     *request.BookID,
		BorrowCode: helper.GenerateBorrowingCode(),
		BorrowedAt: time.Now(),
		DueDate:    calculateDueDate,
	}

	errCreate := u.repository.CreateBorrowing(ctx, borrowing)

	if errCreate != nil {
		if claimed > 0 {
			errRevert := u.waitingListRepository.RevertFulfilledWaitingList(ctx, userID, *request.BookID)

			if errRevert != nil {
				log.Printf("[WaitingList] Gagal mengembalikan status antrean buku %s untuk user %s: %v", *request.BookID, userID, errRevert)
			}
		}

		if helper.IsDuplicateEntryError(errCreate, "uq_active_borrowing") {
			return nil, helper.NewConflictError("Active Borrowing Already Exists!", helper.ErrorDetail{Detail: "You already have an active borrowing for this book. Please return the book before borrowing it again!"})
		}

		return nil, helper.NewInternalServerError("An Error Durng Create Borrowing!", helper.ErrorDetail{Detail: errCreate.Error()})
	}

	result := &dto.BorrowingResponse{
		ID:         borrowing.ID,
		UserID:     borrowing.UserID,
		BookID:     borrowing.BookID,
		BorrowCode: borrowing.BorrowCode,
		BorrowedAt: helper.FormatTimeRFC3339Jakarta(borrowing.BorrowedAt),
		DueDate:    helper.FormatTimeRFC3339Jakarta(borrowing.DueDate),
		ReturnedAt: helper.FormatTimeRFC3339JakartaPTR(borrowing.ReturnedAt),
		Status:     string(borrowing.Status),
		CreatedAt:  helper.FormatTimeRFC3339Jakarta(borrowing.CreatedAt),
		UpdatedAt:  helper.FormatTimeRFC3339Jakarta(borrowing.UpdatedAt),
	}

	return result, nil

}

func (u *BorrowingUsecase) GetAllBorrowings(ctx context.Context, limit int, page int, role bool, userID string) ([]dto.BorrowingResponse, helper.PaginationMeta, error) {

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	if role {

		borrowings, totalData, errGet := u.repository.GetAllBorrowings(ctx, limit, offset)

		result := make([]dto.BorrowingResponse, 0, len(borrowings))

		if errGet != nil {
			return nil, helper.PaginationMeta{}, helper.NewInternalServerError("An Error During Get All Borrowings!", helper.ErrorDetail{Detail: errGet.Error()})
		}

		for _, value := range borrowings {
			result = append(result, dto.BorrowingResponse{
				ID:         value.ID,
				UserID:     value.UserID,
				BookID:     value.BookID,
				BorrowCode: value.BorrowCode,
				BorrowedAt: helper.FormatTimeRFC3339Jakarta(value.BorrowedAt),
				DueDate:    helper.FormatTimeRFC3339Jakarta(value.DueDate),
				ReturnedAt: helper.FormatTimeRFC3339JakartaPTR(value.ReturnedAt),
				Status:     string(value.Status),
				CreatedAt:  helper.FormatTimeRFC3339Jakarta(value.CreatedAt),
				UpdatedAt:  helper.FormatTimeRFC3339Jakarta(value.UpdatedAt),
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

	borrowings, totalData, errGet := u.repository.GetAllMyBorrowings(ctx, userID, limit, offset)

	result := make([]dto.BorrowingResponse, 0, len(borrowings))

	if errGet != nil {
		return nil, helper.PaginationMeta{}, helper.NewInternalServerError("An Error During Get All My Borrowings!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	for _, value := range borrowings {
		result = append(result, dto.BorrowingResponse{
			ID:         value.ID,
			UserID:     value.UserID,
			BookID:     value.BookID,
			BorrowCode: value.BorrowCode,
			BorrowedAt: helper.FormatTimeRFC3339Jakarta(value.BorrowedAt),
			DueDate:    helper.FormatTimeRFC3339Jakarta(value.DueDate),
			ReturnedAt: helper.FormatTimeRFC3339JakartaPTR(value.ReturnedAt),
			Status:     string(value.Status),
			CreatedAt:  helper.FormatTimeRFC3339Jakarta(value.CreatedAt),
			UpdatedAt:  helper.FormatTimeRFC3339Jakarta(value.UpdatedAt),
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

func (u *BorrowingUsecase) GetBorrowingByID(ctx context.Context, ID string, role bool, userID string) (*dto.BorrowingResponse, error) {

	borrowing, errGet := u.repository.GetBorrowingByID(ctx, ID)

	if errGet != nil {
		if errGet == gorm.ErrRecordNotFound {
			return nil, helper.NewNotFoundError("Borrowing Not Found!", helper.ErrorDetail{Detail: "Borrowing with the given ID does not exist!"})
		}
		return nil, helper.NewInternalServerError("An Error During Get Borrowing By ID!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	if role {

		result := &dto.BorrowingResponse{
			ID:         borrowing.ID,
			UserID:     borrowing.UserID,
			BookID:     borrowing.BookID,
			BorrowCode: borrowing.BorrowCode,
			BorrowedAt: helper.FormatTimeRFC3339Jakarta(borrowing.BorrowedAt),
			DueDate:    helper.FormatTimeRFC3339Jakarta(borrowing.DueDate),
			ReturnedAt: helper.FormatTimeRFC3339JakartaPTR(borrowing.ReturnedAt),
			Status:     string(borrowing.Status),
			CreatedAt:  helper.FormatTimeRFC3339Jakarta(borrowing.CreatedAt),
			UpdatedAt:  helper.FormatTimeRFC3339Jakarta(borrowing.UpdatedAt),
		}

		return result, nil
	}

	if borrowing.UserID != userID {
		return nil, helper.NewForbiddenError("Access Denied!", helper.ErrorDetail{Detail: "You do not have access to this borrowing!"})
	}

	result := &dto.BorrowingResponse{
		ID:         borrowing.ID,
		UserID:     borrowing.UserID,
		BookID:     borrowing.BookID,
		BorrowCode: borrowing.BorrowCode,
		BorrowedAt: helper.FormatTimeRFC3339Jakarta(borrowing.BorrowedAt),
		DueDate:    helper.FormatTimeRFC3339Jakarta(borrowing.DueDate),
		ReturnedAt: helper.FormatTimeRFC3339JakartaPTR(borrowing.ReturnedAt),
		Status:     string(borrowing.Status),
		CreatedAt:  helper.FormatTimeRFC3339Jakarta(borrowing.CreatedAt),
		UpdatedAt:  helper.FormatTimeRFC3339Jakarta(borrowing.UpdatedAt),
	}

	return result, nil

}

func (u *BorrowingUsecase) UpdateBorrowingStatus(ctx context.Context, ID string, request *dto.BorrowingUpdateRequest) error {

	if request.Status == nil || *request.Status == "" {
		return helper.NewUnprocessableEntityError("Status Cannot Be Empty!", helper.ErrorDetail{Detail: "Status is required!"})
	}

	reqStatus := strings.ToUpper(*request.Status)

	statusEnum := models.BorrowingStatus(reqStatus)

	switch statusEnum {
	case models.BorrowingStatusPending, models.BorrowingStatusBorrowing, models.BorrowingStatusReturned:
	default:
		return helper.NewUnprocessableEntityError("Invalid Status!", helper.ErrorDetail{Detail: "Status must be PENDING, BORROWED, or RETURNED!"})
	}

	borrowing, errGet := u.repository.GetBorrowingByID(ctx, ID)

	if errGet != nil {
		if errGet == gorm.ErrRecordNotFound {
			return helper.NewNotFoundError("Borrowing Not Found!", helper.ErrorDetail{Detail: "Borrowing with the given ID does not exist!"})
		}
		return helper.NewInternalServerError("An Error During Get Borrowing By ID!", helper.ErrorDetail{Detail: errGet.Error()})
	}

	if borrowing.Status == statusEnum {
		return helper.NewUnprocessableEntityError("Status Already Set!", helper.ErrorDetail{Detail: "The borrowing status is already set to the requested status!"})
	}

	errUpdate := u.repository.UpdateStatus(ctx, ID, statusEnum)

	if errUpdate != nil {
		return helper.NewInternalServerError("An Error During Update Borrowing Status!", helper.ErrorDetail{Detail: errUpdate.Error()})
	}

	var borrowingEvent *event.BorrowingCreatedEvent

	if statusEnum == models.BorrowingStatusReturned {
		borrowingEvent = &event.BorrowingCreatedEvent{
			BookID:    borrowing.BookID,
			Quantity:  1,
			Action:    "RETURNED",
			CreatedAt: time.Now(),
		}

		go func(bookID string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PANIC] processNextWaitingUser: %v", r)
				}
			}()

			if err := u.processNextWaitingUser(context.Background(), bookID); err != nil {
				log.Printf("[WaitingList] Gagal promosi antrean buku %s: %v", bookID, err)
			}
		}(borrowing.BookID)
	} else if statusEnum == models.BorrowingStatusBorrowing {
		borrowingEvent = &event.BorrowingCreatedEvent{
			BookID:    borrowing.BookID,
			Quantity:  1,
			Action:    "BORROWED",
			CreatedAt: time.Now(),
		}
	}

	if borrowingEvent != nil {
		go func(event *event.BorrowingCreatedEvent) {
			errPublish := u.kafkaProducer.PublishBorrowingCreatedEvent(context.Background(), event, ID, u.cfg.Kafka.TopicBorrowingCreated)
			if errPublish != nil {
				log.Printf("[Kafka Publish Error] Failed to send event for bookID in borrowing service %s: %v", ID, errPublish)
			}
		}(borrowingEvent)
	}

	return nil

}

func (u *BorrowingUsecase) processNextWaitingUser(ctx context.Context, bookID string) error {

	const maxAttempts = 5

	for attempt := 0; attempt < maxAttempts; attempt++ {

		nextUser, errGet := u.waitingListRepository.GetFirstWaitingListByBookID(ctx, bookID)

		if errGet != nil {
			if errGet == gorm.ErrRecordNotFound {
				return nil
			}
			return helper.NewInternalServerError("An Error During Get First Waiting List By Book ID!", helper.ErrorDetail{Detail: errGet.Error()})
		}

		rowsAffected, errUpdate := u.waitingListRepository.UpdateStatusWaitingList(ctx, nextUser.ID, models.WaitingListStatusWaiting, models.WaitingListStatusNotified)

		if errUpdate != nil {
			return helper.NewInternalServerError("An Error During Update Waiting List Status!", helper.ErrorDetail{Detail: errUpdate.Error()})
		}

		if rowsAffected > 0 {
			return nil
		}

		log.Printf("[WaitingList] Kandidat %s sudah diambil proses lain, lanjut ke antrean berikutnya", nextUser.ID)
	}

	return nil

}
