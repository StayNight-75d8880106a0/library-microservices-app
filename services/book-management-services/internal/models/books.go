package models

import (
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type Books struct {
	ID             string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	ISBN           string         `gorm:"type:varchar(20);unique;not null" json:"isbn"`
	Title          string         `gorm:"type:varchar(255);not null" json:"title"`
	Authors        string         `gorm:"type:text" json:"authors"`
	Publisher      string         `gorm:"type:varchar(150)" json:"publisher"`
	PublishedDate  string         `gorm:"type:varchar(50)" json:"published_date"`
	Page           int            `gorm:"type:int" json:"page"`
	Description    string         `gorm:"type:text" json:"description"`
	CoverURL       string         `gorm:"type:text" json:"cover_url"`
	Category       string         `gorm:"type:varchar(100);default:'General'" json:"category"`
	TotalStock     int            `gorm:"type:int;not null;default:1" json:"total_stock"`
	AvailableStock int            `gorm:"type:int;not null;default:1" json:"available_stock"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (b *Books) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return
}
