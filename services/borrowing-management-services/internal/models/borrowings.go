package models

import (
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type BorrowingStatus string

const (
	BorrowingStatusPending   BorrowingStatus = "PENDING"
	BorrowingStatusBorrowing BorrowingStatus = "BORROWED"
	BorrowingStatusReturned  BorrowingStatus = "RETURNED"
)

type Borrowing struct {
	ID         string          `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID     string          `gorm:"type:varchar(36);not null" json:"user_id"`
	BookID     string          `gorm:"type:varchar(36);not null" json:"book_id"`
	BorrowCode string          `gorm:"type:varchar(255);unique;not null" json:"borrow_code"`
	BorrowedAt time.Time       `gorm:"type:datetime;not null" json:"borrowed_at"`
	DueDate    time.Time       `gorm:"type:datetime;not null" json:"due_date"`
	ReturnedAt *time.Time      `gorm:"type:datetime" json:"returned_at"`
	Status     BorrowingStatus `gorm:"type:enum('PENDING', 'BORROWED', 'RETURNED');not null;default:'PENDING'" json:"status"`
	CreatedAt  time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (b *Borrowing) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return
}
