package producer

import (
	"auth-services/internal/helper"
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaProducer{
		writer: writer,
	}
}

func (kp *KafkaProducer) PublishUserCreatedEvent(ctx context.Context, event interface{}, key string) error {

	payload, errPayload := json.Marshal(event)

	if errPayload != nil {
		return helper.NewInternalServerError("An Error During Marhal Json In Publish User Event! :", helper.ErrorDetail{Detail: errPayload.Error()})
	}

	err := kp.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: payload,
	})

	if err != nil {
		return helper.NewInternalServerError("An Error During Publish User Event! :", helper.ErrorDetail{Detail: err.Error()})
	}

	return nil

}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}
