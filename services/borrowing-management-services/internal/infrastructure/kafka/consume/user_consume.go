package consume

import (
	"borrowing-management-services/internal/infrastructure/kafka/event"
	"context"
	"fmt"
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
		PartitionWatchInterval: 5 * time.Millisecond,
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
	backoffMin = 1 * time.Second
	backoffMax = 30 * time.Second
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
				log.Printf("kafka consumer error, topic: %s, group_id: %s, error: %v", kfk.reader.Config().Topic, kfk.reader.Config().GroupID, errMsg)
				slog.Error("kafka consumer error",
					"topic", kfk.reader.Config().Topic,
					"group_id", kfk.reader.Config().GroupID,
					"error", errMsg,
				)
				continue
			}

			errProcess := processFunc(ctx, message.Value)
			if errProcess != nil {
				log.Printf("kafka consumer processing error, topic: %s, group_id: %s, error: %v", kfk.reader.Config().Topic, kfk.reader.Config().GroupID, errProcess)
				slog.Error("kafka consumer processing error",
					"topic", kfk.reader.Config().Topic,
					"group_id", kfk.reader.Config().GroupID,
					"error", errProcess,
				)
				continue
			}

			if errCommit := kfk.reader.CommitMessages(ctx, message); errCommit != nil {
				slog.Error("kafka consumer commit error",
					"topic", kfk.reader.Config().Topic,
					"group_id", kfk.reader.Config().GroupID,
					"error", errCommit,
				)
			}
		}
	}

}

func (kfk *KafkaConsumer) Close() error {
	return kfk.reader.Close()
}
