package producer

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaProducer{
		writer: writer,
	}
}

func (kp *KafkaProducer) PublishBorrowingCreatedEvent(ctx context.Context, event interface{}, key string, topic string) error {

	payload, errPayload := json.Marshal(event)

	if errPayload != nil {
		return errors.New("An Error During Marhal Json In Publish Borrowing Event! :" + errPayload.Error())
	}

	err := kp.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: payload,
		Topic: topic,
	})

	if err != nil {
		return errors.New("An Error During Marhal Json In Publish Borrowing Event! :" + err.Error())
	}

	return nil

}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}
