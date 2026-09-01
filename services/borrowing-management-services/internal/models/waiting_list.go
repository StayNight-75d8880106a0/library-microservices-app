package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WaitingListStatus string

const (
	WaitingListStatusWaiting   WaitingListStatus = "WAITING"
	WaitingListStatusNotified  WaitingListStatus = "NOTIFIED"
	WaitingListStatusCancelled WaitingListStatus = "CANCELLED"
	WaitingListStatusFulFilled WaitingListStatus = "FULFILLED"
)

type WaitingList struct {
	ID             string            `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID         string            `gorm:"type:varchar(36);not null" json:"user_id"`
	BookID         string            `gorm:"type:varchar(36);not null" json:"book_id"`
	ArrivalRate    float64           `gorm:"type:decimal(10,4);not null" json:"arrival_rate"`  // λ (request/jam)
	ServiceRate    float64           `gorm:"type:decimal(10,4);not null" json:"service_rate"`  // μ (request/jam per server)
	NumServers     int               `gorm:"type:int;not null" json:"num_servers"`             // c
	Utilization    float64           `gorm:"type:decimal(6,4);not null" json:"utilization"`    // ρ
	ProbWait       float64           `gorm:"type:decimal(6,4);not null" json:"prob_wait"`      // P(wait)
	AvgQueueLen    float64           `gorm:"type:decimal(10,4);not null" json:"avg_queue_len"` // Lq
	AvgWaitMin     float64           `gorm:"type:decimal(10,4);not null" json:"avg_wait_min"`  // Wq (menit)
	OptimalServers *int              `gorm:"type:int;null" json:"optimal_servers"`             // c*
	QueueNumber    int               `gorm:"type:int;not null" json:"queue_number"`
	Status         WaitingListStatus `gorm:"type:enum('WAITING', 'NOTIFIED', 'CANCELLED', 'FULFILLED');not null;default:'WAITING'" json:"status"`
	CreatedAt      time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
}

func (b *WaitingList) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return
}
