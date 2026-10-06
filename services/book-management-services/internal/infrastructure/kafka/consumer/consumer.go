package consumer

import (
	"book-management-services/internal/dto"
	"book-management-services/internal/helper"
	"book-management-services/internal/usecase"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader  *kafka.Reader
	usecase usecase.BookUsecaseInterface
}

func NewKafkaConsumer(brokerAddress []string, topic string, groupID string, bookUsecase usecase.BookUsecaseInterface) *KafkaConsumer {
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
		usecase: bookUsecase,
	}
}

const (
	backoffMin  = 1 * time.Second
	backoffMax  = 30 * time.Second
	maxAttempts = 6
)

func (kfk *KafkaConsumer) StartConsuming(ctx context.Context) {

	slog.Info("kafka consumer started",
		"topic", kfk.reader.Config().Topic,
		"group_id", kfk.reader.Config().GroupID,
	)

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("kafka consumer stopping",
					"topic", kfk.reader.Config().Topic,
					"group_id", kfk.reader.Config().GroupID,
				)
				return
			default:
				msg, errMsg := kfk.reader.FetchMessage(ctx)

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

				if !kfk.processMessage(ctx, msg) {
					return
				}

				comitCTX, cancelConmmit := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
				errCommit := kfk.reader.CommitMessages(comitCTX, msg)
				cancelConmmit()

				if errCommit != nil {
					slog.Error("kafka consumer error",
						"topic", kfk.reader.Config().Topic,
						"group_id", kfk.reader.Config().GroupID,
						"error", errCommit,
					)
				}
			}
		}
	}()

}

func (kfk *KafkaConsumer) processMessage(ctx context.Context, msg kafka.Message) bool {

	var eventUser dto.BorrowingCreatedEvent

	errUnmarshal := json.Unmarshal(msg.Value, &eventUser)

	if errUnmarshal != nil {
		slog.Error("kafka consumer error",
			"topic", kfk.reader.Config().Topic,
			"group_id", kfk.reader.Config().GroupID,
			"error", errUnmarshal,
		)
		return true
	}

	backoff := backoffMin

	for attempt := 1; ; attempt++ {

		errProcess := kfk.usecase.UpdateAvaliableStock(ctx, &eventUser)

		if errProcess == nil {
			return true
		}

		if ctx.Err() != nil {
			return false
		}

		var appErr *helper.AppError

		isClientError := errors.As(errProcess, &appErr) && appErr.Code >= 400 && appErr.Code < 500

		if isClientError || attempt >= maxAttempts {
			slog.Error("kafka consumer error",
				"topic", kfk.reader.Config().Topic,
				"group_id", kfk.reader.Config().GroupID,
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
