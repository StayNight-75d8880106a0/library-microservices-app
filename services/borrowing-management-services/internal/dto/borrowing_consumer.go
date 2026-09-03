package dto

import "time"

type UserAuthKafkaPayloadConsumer struct {
	EventType  string    `json:"eventType"`
	KeycloakID string    `json:"keycloakId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type UserStatusUpdatedKafkaPayloadConsumer struct {
	EventType string    `json:"eventType"`
	UserID    string    `json:"userId"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
}
