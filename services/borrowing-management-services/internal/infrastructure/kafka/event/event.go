package event

import "time"

type BorrowingCreatedEvent struct {
	BookID    string    `json:"bookID"`
	Quantity  int       `json:"quantity"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"createdAt"`
}
