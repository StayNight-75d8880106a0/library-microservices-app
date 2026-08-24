package event

import "time"

type UserCreatedEvent struct {
	KeycloakID string    `json:"keycloakId"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	CreatedAt  time.Time `json:"createdAt"`
}
