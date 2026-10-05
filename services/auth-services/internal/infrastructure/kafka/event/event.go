package event

import (
	"encoding/json"
	"time"
)

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

type UserAuditLogEvent struct {
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
