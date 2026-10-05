package event

import (
	"encoding/json"
	"time"
)

type BorrowingCreatedEvent struct {
	BookID    string    `json:"bookID"`
	Quantity  int       `json:"quantity"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"createdAt"`
}

type BorrowingAuditLogEvent struct {
	EventID         *string         `json:"eventID"`
	TraceID         *string         `json:"traceID"`
	UserID          *string         `json:"userID"`
	ServiceName     *string         `json:"serviceName"`
	Method          *string         `json:"method"`
	Endpoint        *string         `json:"endpoint"`
	HTTPStatus      *string         `json:"httpStatus"`
	HTTPCode        *int            `json:"httpCode"`
	Kind            *string         `json:"kind"`
	IPAddress       *string         `json:"ipAddress"`
	RequestBody     json.RawMessage `json:"requestBody"`
	ResponseBody    json.RawMessage `json:"responseBody"`
	ExecutionTimeMs *int            `json:"executionTime"`
	IsRoot          bool            `json:"isRoot"`
	OccurredAt      time.Time       `json:"occurredAt"`
}
