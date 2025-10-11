package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaProducer interface {
	PublishTeamUpdate(ctx context.Context, teamData interface{}) error
}

type UpdateTeam struct {
	producer   *kafka.Producer
	topic      string
	deliverych chan (kafka.Event)
}

func NewUpdateTeam(p *kafka.Producer, topic string) *UpdateTeam {
	return &UpdateTeam{
		producer:   p,
		topic:      topic,
		deliverych: make(chan kafka.Event, 1000),
	}
}

func (c *UpdateTeam) PublishTeamUpdate(ctx context.Context, teamData interface{}) error {

	data, err := json.Marshal(teamData)
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
	log.Println(">>successfully published event to the queue")
	fmt.Println("Published updateuserprofile event onto the queue")
	return err //fmt.Errorf("failed to publish user creation data: %s", err)
}

func InitKafkaProducer() (*kafka.Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "wyckie",
		"acks":              "all",
	})
	fmt.Println("initialized successfully....")
	if err != nil {
		log.Fatalf("failed to produce user created event: %s", err)
	}
	return p, nil
}
