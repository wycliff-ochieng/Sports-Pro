package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

func (c *CreateUser) DeliveryReportHandler() {
	for e := range c.deliverych {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				//message was not delivered
				log.Printf("CRITICAL , Delivery failed for message in topic %s : %v\n", *&ev.TopicPartition, ev.TopicPartition.Error)
			} else {
				log.Printf("successfuly delivered message to topic %s, partition %d, and offset %v\n", *ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
			}
		}
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
	if err != nil {
		log.Printf("failed to enqueueu message: %v", err)
		return fmt.Errorf("failed to enqueue message for kafka:%v", err)
	}
	log.Printf(">>successfully published event to the topic :%v", c.topic)
	fmt.Println("Published event onto the topic")
	return nil
}

func InitKafkaProducer() (*kafka.Producer, error) {

	kafkaURL := os.Getenv("KAFKA_BROKER")
	if kafkaURL == "" {
		kafkaURL = "localhost:9092"
	}

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaURL, //"localhost:9092",
		"client.id":         "wyckie",
		"acks":              "all",
	})
	fmt.Println("initialized successfully....")
	if err != nil {
		log.Fatalf("failed to produce user created event: %s", err)
	}
	return p, nil
}
