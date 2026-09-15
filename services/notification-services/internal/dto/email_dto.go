package dto

import "time"

type UserCreatedConsumer struct {
	KeycloakID       string    `json:"keycloakId"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Email            string    `json:"email"`
	VerificationLink string    `json:"verificationLink"`
	CreatedAt        time.Time `json:"createdAt"`
}

type EmailLogAllResponse struct {
	ID        *string `json:"id"`
	Email     string  `json:"email"`
	Subject   string  `json:"subject"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
}

type EmailLogDetailResponse struct {
	ID           *string `json:"id"`
	UserID       string  `json:"userID"`
	Email        string  `json:"email"`
	FirstName    string  `json:"firstName"`
	LastName     string  `json:"lastName"`
	Subject      string  `json:"subject"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"errorMessage"`
	CreatedAt    string  `json:"createdAt"`
}
