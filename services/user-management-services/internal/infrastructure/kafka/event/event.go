package event

import "time"

type UserUpdatedEvent struct {
	EventType string    `json:"eventType"`
	UserID    string    `json:"userID"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
}
