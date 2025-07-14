package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaProducer interface {
	PublishUserCreation(ctx context.Context, userData interface{}) error
}

type CreateUser struct {
	producer   *kafka.Producer
	topic      string
	deliverych chan (kafka.Event)
}

func NewCreateUser(p *kafka.Producer, topic string) *CreateUser {
	return &CreateUser{
		producer:   p,
		topic:      topic,
		deliverych: make(chan kafka.Event, 1000),
	}
}

func (c *CreateUser) PublishUserCreation(ctx context.Context, userData interface{}) error {

	data, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %s", err)
	}

	err = c.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &c.topic,
			Partition: kafka.PartitionAny,
		},
		Value: data,
	}, c.deliverych)
	fmt.Println("Published event onto the queue")
	return fmt.Errorf("failed to publish user creation data: %s", err)
}

func InitKafkaProducer() (*kafka.Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "wyckie",
		"acks":              "all",
	})
	if err != nil {
		log.Fatalf("failed to produce user created event: %s", err)
	}
	return p, nil
}
