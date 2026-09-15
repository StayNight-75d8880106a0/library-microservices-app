package models

import "time"

type EmailStatus string

const (
	EmailStatusFailed  EmailStatus = "FAILED"
	EmailStatusSuccess EmailStatus = "SUCCESS"
)

type EmailLogs struct {
	ID           *string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       string      `json:"user_id" gorm:"type:varchar(100);not null"`
	Email        string      `json:"email" gorm:"type:varchar(255);not null"`
	FirstName    string      `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName     string      `json:"last_name" gorm:"type:varchar(100);not null"`
	Subject      string      `json:"subject" gorm:"type:varchar(255);not null"`
	Status       EmailStatus `json:"status" gorm:"type:email_status;not null"`
	ErrorMessage *string     `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time   `json:"created_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}
