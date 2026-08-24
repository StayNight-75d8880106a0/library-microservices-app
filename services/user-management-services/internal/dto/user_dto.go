package dto

import "time"

type UserUpdatePasswordRequest struct {
	NewPassword     *string `json:"newPassword"`
	ConfirmPassword *string `json:"confirmPassword"`
}

type UserUpdateProfileRequest struct {
	FirstName   *string `json:"firstName"`
	LastName    *string `json:"lastName"`
	PhoneNumber *string `json:"phoneNumber"`
	Address     *string `json:"address"`
}

type CreateAdminRequest struct {
	Username  *string `json:"username"`
	Email     *string `json:"email" binding:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Password  *string `json:"password"`
}

type AdminResponse struct {
	ID            *string `json:"id"`
	Username      *string `json:"username"`
	FirstName     *string `json:"firstName"`
	LastName      *string `json:"lastName"`
	Email         *string `json:"email"`
	EmailVerified *bool   `json:"emailVerified"`
	CreatedAt     string  `json:"createdAt"`
}

type UserResponse struct {
	ID                *string `json:"id"`
	Username          *string `json:"username"`
	FirstName         *string `json:"firstName"`
	LastName          *string `json:"lastName"`
	Email             *string `json:"email"`
	EmailVerified     *bool   `json:"emailVerified"`
	PhoneNumber       *string `json:"phoneNumber"`
	Address           *string `json:"address"`
	ProfileStatus     string  `json:"profileStatus"`
	LibraryCardNumber *string `json:"libraryCardNumber"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type UserCreatedKafkaPayloadConsumer struct {
	KeycloakID string    `json:"keycloakId"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	CreatedAt  time.Time `json:"createdAt"`
}
