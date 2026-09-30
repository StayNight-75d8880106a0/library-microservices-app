package models

import (
	"encoding/json"
	"time"
)

type Logbook struct {
	ID              *string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EventID         *string         `json:"event_id" gorm:"type:uuid;not null"`
	TraceID         *string         `json:"trace_id" gorm:"type:varchar(225);not null"`
	UserID          *string         `json:"user_id" gorm:"type:uuid;not null"`
	ServiceName     *string         `json:"service_name" gorm:"type:varchar(100);not null"`
	Method          *string         `json:"method" gorm:"type:varchar(10);not null"`
	Endpoint        *string         `json:"endpoint" gorm:"type:varchar(200);not null"`
	HTTPStatus      *string         `json:"http_status" gorm:"type:varchar(200)"`
	HTTPCode        *int            `json:"http_code" gorm:"type:int"`
	Kind            *string         `json:"kind" gorm:"type:varchar(100)"`
	IPAddress       *string         `json:"ip_address" gorm:"type:varchar(45)"`
	RequestBody     json.RawMessage `json:"request_body" gorm:"type:jsonb"`
	ResponseBody    json.RawMessage `json:"response_body" gorm:"type:jsonb"`
	ExecutionTimeMs *int            `json:"execution_time_ms" gorm:"type:int;not null"`
	IsRoot          bool            `json:"is_root" gorm:"type:boolean;not null;default:false"`
	OccurredAt      time.Time       `json:"occurred_at" gorm:"type:timestamptz;not null"`
	CreatedAt       time.Time       `json:"created_at" gorm:"type:timestamptz;not null;default:now()"`
}
