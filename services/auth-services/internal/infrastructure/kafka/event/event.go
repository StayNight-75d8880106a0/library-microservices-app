package event

import "time"

type UserCreatedEvent struct {
	EventID          string    `json:"eventId"`
	KeycloakID       string    `json:"keycloakId"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Email            string    `json:"email"`
	VerificationLink string    `json:"verificationLink"`
	CreatedAt        time.Time `json:"createdAt"`
}

type UserAuthenticatedEvent struct {
	EventType  string    `json:"eventType"`
	KeycloakID string    `json:"keycloakId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type VerificationEmailRequestedEvent struct {
	EventID          string    `json:"eventId"`
	KeycloakID       string    `json:"keycloakId"`
	Email            string    `json:"email"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	VerificationLink string    `json:"verificationLink"`
	CreatedAt        time.Time `json:"createdAt"`
}
