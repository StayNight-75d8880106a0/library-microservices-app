package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"notification-services/internal/dto"
	"notification-services/internal/usecase"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader  *kafka.Reader
	usecase usecase.EmailUsecaseInterface
}

func NewKafkaConsumer(brokerAddress []string, topic string, groupID string, emailUsecase usecase.EmailUsecaseInterface) *KafkaConsumer {
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
		usecase: emailUsecase,
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

				var eventUser dto.UserCreatedConsumer

				errUnmarshal := json.Unmarshal(msg.Value, &eventUser)

				if errUnmarshal != nil {
					slog.Error("kafka consumer error",
						"topic", kfk.reader.Config().Topic,
						"group_id", kfk.reader.Config().GroupID,
						"error", errUnmarshal,
					)
					continue
				}

				_, errCreate := kfk.usecase.SaveEmailLog(ctx, &eventUser)

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
