package models

import "time"

type UserStatus string

const (
	UserStatusActive     UserStatus = "ACTIVE"
	UserStatusIncomplete UserStatus = "INCOMPLETE"
	UserStatusSuspended  UserStatus = "SUSPENDED"
)

type Users struct {
	KeycloakUserID    string     `json:"keycloak_user_id" gorm:"column:keycloak_user_id;primaryKey"`
	FirstName         string     `json:"first_name"`
	LastName          string     `json:"last_name"`
	PhoneNumber       string     `json:"phone_number"`
	Address           string     `json:"address"`
	ProfileStatus     UserStatus `json:"profile_status"`
	LibraryCardNumber string     `json:"library_card_number"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}
