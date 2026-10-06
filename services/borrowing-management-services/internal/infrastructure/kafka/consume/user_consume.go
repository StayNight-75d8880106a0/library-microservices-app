package consume

import (
	"borrowing-management-services/internal/infrastructure/kafka/event"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader  *kafka.Reader
	handler *event.EventHandler
}

func NewKafkaConsumer(brokerAddress []string, topic string, groupID string, handler *event.EventHandler) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:                brokerAddress,
		Topic:                  topic,
		GroupID:                groupID,
		WatchPartitionChanges:  true,
		PartitionWatchInterval: 5 * time.Second,
		StartOffset:            kafka.FirstOffset,
		MaxWait:                501 * time.Millisecond,
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			slog.Error("kafka reader", "msg", fmt.Sprintf(msg, args...))
		}),
	})
	return &KafkaConsumer{
		reader:  reader,
		handler: handler,
	}
}

const (
	backoffMin  = 1 * time.Second
	backoffMax  = 30 * time.Second
	maxAttempts = 6
)

func (kfk *KafkaConsumer) StartConsuming(ctx context.Context, processFunc func(ctx context.Context, msg []byte) error) {

	log.Printf("kafka consumer started, topic: %s, group_id: %s", kfk.reader.Config().Topic, kfk.reader.Config().GroupID)

	defer kfk.reader.Close()

	for {
		select {
		case <-ctx.Done():
			log.Printf("kafka consumer stopping, topic: %s, group_id: %s", kfk.reader.Config().Topic, kfk.reader.Config().GroupID)
			return
		default:
			message, errMsg := kfk.reader.FetchMessage(ctx)

			if errMsg != nil {
				if ctx.Err() != nil || errors.Is(errMsg, io.EOF) {
					return
				}

				slog.Error("kafka consumer error",
					"topic", kfk.reader.Config().Topic,
					"group_id", kfk.reader.Config().GroupID,
					"error", errMsg,
				)

				sleepWithContext(ctx, backoffMin)
				continue
			}

			if !kfk.processMessage(ctx, message, processFunc) {
				return
			}

			commitCtx, cancelCommit := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			errCommit := kfk.reader.CommitMessages(commitCtx, message)
			cancelCommit()

			if errCommit != nil {
				slog.Error("kafka consumer commit error",
					"topic", kfk.reader.Config().Topic,
					"group_id", kfk.reader.Config().GroupID,
					"error", errCommit,
				)
			}
		}
	}

}

func (kfk *KafkaConsumer) processMessage(ctx context.Context, msg kafka.Message, processFunc func(ctx context.Context, msg []byte) error) bool {

	backoff := backoffMin

	for attempt := 1; ; attempt++ {

		errProcess := processFunc(ctx, msg.Value)

		if errProcess == nil {
			return true
		}

		if ctx.Err() != nil {
			return false
		}

		var syntaxErr *json.SyntaxError

		isInvalidPayload := errors.As(errProcess, &syntaxErr)

		if isInvalidPayload || attempt >= maxAttempts {
			slog.Error("kafka consumer processing error",
				"topic", kfk.reader.Config().Topic,
				"group_id", kfk.reader.Config().GroupID,
				"offset", msg.Offset,
				"error", errProcess,
			)
			return true
		}

		slog.Warn("kafka message retry", "offset", msg.Offset, "attempt", attempt, "error", errProcess)

		if !sleepWithContext(ctx, backoff) {
			return false
		}

		backoff = min(backoff*2, backoffMax)
	}
}

func sleepWithContext(ctx context.Context, duration time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(duration):
		return true
	}
}

func (kfk *KafkaConsumer) Close() error {
	return kfk.reader.Close()
}
