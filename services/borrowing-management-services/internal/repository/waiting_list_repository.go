package repository

import (
	"borrowing-management-services/internal/models"
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type WaitingListRepositoryInterface interface {
	Create(ctx context.Context, data *models.WaitingList) error
	GetALLWaitingList(ctx context.Context, limit int, offset int) ([]models.WaitingList, int64, error)
	GetALLMyWaitingList(ctx context.Context, limit int, offset int, userID string) ([]models.WaitingList, int64, error)
	GetWaitingListByID(ctx context.Context, ID string) (*models.WaitingList, error)
	UpdateStatusWaitingList(ctx context.Context, ID string, status models.WaitingListStatus) (int64, error)
	// GetLastQueueNumber(ctx context.Context, bookID string) (int, error)
	GetFirstWaitingListByBookID(ctx context.Context, bookID string) (*models.WaitingList, error)
	CancelWaitingListByUser(ctx context.Context, ID string, userID string) (int64, error)
}

type WaitingListRepository struct {
	DB *gorm.DB
}

func NewWaitingListRepository(db *gorm.DB) *WaitingListRepository {
	return &WaitingListRepository{
		DB: db,
	}
}

const maxCreateRetries = 3

func (repo *WaitingListRepository) Create(ctx context.Context, data *models.WaitingList) error {

	var lastErr error

	for attempt := 0; attempt < maxCreateRetries; attempt++ {

		data.ID = ""

		lastErr = repo.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var lastQueueNumber int

			err := tx.Raw(
				"SELECT COALESCE(MAX(queue_number), 0) FROM waiting_lists WHERE book_id = ? FOR UPDATE",
				data.BookID,
			).Scan(&lastQueueNumber).Error
			if err != nil {
				return err
			}

			data.QueueNumber = lastQueueNumber + 1

			return tx.Table("waiting_lists").Create(data).Error
		})

		if lastErr == nil {
			return nil
		}

		if ctx.Err() != nil {
			return lastErr
		}

		if !isRetryableMySQLError(lastErr) {
			return lastErr
		}
	}

	return lastErr

}

func (repo *WaitingListRepository) GetALLWaitingList(ctx context.Context, limit int, offset int) ([]models.WaitingList, int64, error) {

	var waitingLists []models.WaitingList
	var total int64

	errCount := repo.DB.WithContext(ctx).Table("waiting_lists").Count(&total).Error

	if errCount != nil {
		return waitingLists, 0, errCount
	}

	errGet := repo.DB.WithContext(ctx).Table("waiting_lists").Order("created_at DESC").Limit(limit).Offset(offset).Find(&waitingLists).Error

	return waitingLists, total, errGet

}

func (repo *WaitingListRepository) GetALLMyWaitingList(ctx context.Context, limit int, offset int, userID string) ([]models.WaitingList, int64, error) {

	var waitingLists []models.WaitingList
	var total int64

	errCount := repo.DB.WithContext(ctx).Table("waiting_lists").Where("user_id = ?", userID).Count(&total).Error

	if errCount != nil {
		return waitingLists, 0, errCount
	}

	errGet := repo.DB.WithContext(ctx).Table("waiting_lists").Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&waitingLists).Error

	return waitingLists, total, errGet

}

func (repo *WaitingListRepository) GetWaitingListByID(ctx context.Context, ID string) (*models.WaitingList, error) {

	var waitingList models.WaitingList

	errGet := repo.DB.WithContext(ctx).Table("waiting_lists").Where("id = ?", ID).First(&waitingList).Error

	return &waitingList, errGet

}

func (repo *WaitingListRepository) UpdateStatusWaitingList(ctx context.Context, ID string, status models.WaitingListStatus) (int64, error) {

	result := repo.DB.WithContext(ctx).Table("waiting_lists").
		Where("id = ? AND status = ?", ID, models.WaitingListStatusWaiting).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": gorm.Expr("NOW()"),
		})

	return result.RowsAffected, result.Error

}

// func (repo *WaitingListRepository) GetLastQueueNumber(ctx context.Context, bookID string) (int, error) {

// 	var lastQueueNumber int

// 	errGet := repo.DB.WithContext(ctx).Table("waiting_lists").Where("book_id = ?", bookID).Select("COALESCE(MAX(queue_number), 0)").Scan(&lastQueueNumber).Error

// 	return lastQueueNumber, errGet

// }

func (repo *WaitingListRepository) GetFirstWaitingListByBookID(ctx context.Context, bookID string) (*models.WaitingList, error) {

	var waitingList models.WaitingList

	errGet := repo.DB.WithContext(ctx).Table("waiting_lists").Where("book_id = ? AND status = ?", bookID, models.WaitingListStatusWaiting).Order("queue_number ASC").First(&waitingList).Error

	return &waitingList, errGet

}

func isRetryableMySQLError(err error) bool {
	var mysqlErr *mysql.MySQLError

	if !errors.As(err, &mysqlErr) {
		return false
	}

	switch mysqlErr.Number {
	case 1062, // Duplicate entry (uq_book_queue)
		1213, // Deadlock found
		1205: // Lock wait timeout exceeded
		return true
	}

	return false
}

func (repo *WaitingListRepository) CancelWaitingListByUser(ctx context.Context, ID string, userID string) (int64, error) {

	result := repo.DB.WithContext(ctx).Table("waiting_lists").
		Where("id = ? AND user_id = ? AND status = ?", ID, userID, models.WaitingListStatusWaiting).
		Updates(map[string]interface{}{
			"status":     models.WaitingListStatusCancelled,
			"updated_at": gorm.Expr("NOW()"),
		})

	return result.RowsAffected, result.Error

}
