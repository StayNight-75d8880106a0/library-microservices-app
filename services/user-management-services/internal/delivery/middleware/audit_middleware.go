package middleware

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"
	"user-management-services/internal/config"
	"user-management-services/internal/helper"
	"user-management-services/internal/infrastructure/kafka/event"
	"user-management-services/internal/infrastructure/kafka/producer"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type auditBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (writer *auditBodyWriter) Write(data []byte) (int, error) {

	writer.body.Write(data)

	return writer.ResponseWriter.Write(data)

}

func (writer *auditBodyWriter) WriteString(data string) (int, error) {

	writer.body.WriteString(data)

	return writer.ResponseWriter.WriteString(data)

}

func AuditMiddleware(kafkaProducer *producer.KafkaProducer, cfg *config.AppConfig, serviceName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		traceID := ctx.GetHeader("X-Trace-Id")

		isRoot := false

		_, errParse := uuid.Parse(traceID)

		if errParse != nil {
			traceID = uuid.NewString()
			isRoot = true
		}

		ctx.Request = ctx.Request.WithContext(helper.WithTraceID(ctx.Request.Context(), traceID))

		ctx.Header("X-Trace-Id", traceID)

		var requestBody []byte

		if ctx.Request.Body != nil {
			requestBody, _ = io.ReadAll(ctx.Request.Body)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		bodyWriter := &auditBodyWriter{
			ResponseWriter: ctx.Writer,
			body:           bytes.NewBuffer(nil),
		}

		ctx.Writer = bodyWriter

		occurredAt := time.Now()

		ctx.Next()

		eventID := uuid.NewString()
		executionTimeMs := int(time.Since(occurredAt).Milliseconds())
		httpCode := ctx.Writer.Status()
		httpStatus := http.StatusText(httpCode)
		method := ctx.Request.Method
		endpoint := ctx.FullPath()
		ipAddress := ctx.ClientIP()
		kind := "HTTP"

		if endpoint == "" {
			endpoint = ctx.Request.URL.Path
		}

		var userID *string

		if value := ctx.GetString("userID"); value != "" {
			userID = &value
		}

		auditEvent := event.UserAuditLogEvent{
			EventID:         &eventID,
			TraceID:         &traceID,
			UserID:          userID,
			ServiceName:     &serviceName,
			Method:          &method,
			Endpoint:        &endpoint,
			HTTPStatus:      &httpStatus,
			HTTPCode:        &httpCode,
			Kind:            &kind,
			IPAddress:       &ipAddress,
			RequestBody:     helper.SafeJSONBody(helper.RedactSensitiveData(requestBody)),
			ResponseBody:    helper.SafeJSONBody(helper.RedactSensitiveData(bodyWriter.body.Bytes())),
			ExecutionTimeMs: &executionTimeMs,
			IsRoot:          isRoot,
			OccurredAt:      occurredAt,
		}

		publishContext, cancel := context.WithTimeout(context.WithoutCancel(ctx.Request.Context()), 5*time.Second)
		defer cancel()

		errPublish := kafkaProducer.PublishEvent(publishContext, auditEvent, traceID, cfg.Kafka.TopicAuditLog)

		if errPublish != nil {
			log.Printf("Failed to publish audit log event: %v", errPublish)
		}

	}
}
