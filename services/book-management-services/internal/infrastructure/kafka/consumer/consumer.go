package consumer

import (
	"book-management-services/internal/dto"
	"book-management-services/internal/usecase"
	"context"
	"encoding/json"
	"fmt"
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
		PartitionWatchInterval: 5 * time.Millisecond,
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
	backoffMin = 1 * time.Second
	backoffMax = 30 * time.Second
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
					slog.Error("kafka consumer error",
						"topic", kfk.reader.Config().Topic,
						"group_id", kfk.reader.Config().GroupID,
						"error", errMsg,
					)
					continue
				}

				var eventUser dto.BorrowingCreatedEvent

				errUnmarshal := json.Unmarshal(msg.Value, &eventUser)

				if errUnmarshal != nil {
					slog.Error("kafka consumer error",
						"topic", kfk.reader.Config().Topic,
						"group_id", kfk.reader.Config().GroupID,
						"error", errUnmarshal,
					)
					continue
				}

				errCreate := kfk.usecase.UpdateAvaliableStock(ctx, &eventUser)

				if errCreate != nil {
					slog.Error("kafka consumer error",
						"topic", kfk.reader.Config().Topic,
						"group_id", kfk.reader.Config().GroupID,
						"error", errCreate,
					)
					continue
				}

				errCommit := kfk.reader.CommitMessages(ctx, msg)

				if errCommit != nil {
					slog.Error("kafka consumer error",
						"topic", kfk.reader.Config().Topic,
						"group_id", kfk.reader.Config().GroupID,
						"error", errCommit,
					)
					continue
				}
			}
		}
	}()

}

func (kfk *KafkaConsumer) Close() error {
	return kfk.reader.Close()
}
