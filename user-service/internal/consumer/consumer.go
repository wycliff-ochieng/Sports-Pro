package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type UserEventCreated struct {
	UserID int    `json:"userid"`
	Email  string `json:"email"`
}

type UserEventConsumer struct {
	l *log.Logger
}

func NewUserEventConsumer(l *log.Logger) *UserEventConsumer {
	return &UserEventConsumer{l}
}

func (c *UserEventConsumer) StartMessageConsume(ctx context.Context, timeout time.Duration, consumer *kafka.Consumer) {
	c.l.Println(">>> consuming topic messages started successsfully")

	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	//timeout := time.Second * 60

	for {
		message, err := consumer.ReadMessage(timeout)
		if err != nil {
			log.Fatalf("something unexoected happened while trying to fetch messages: %s", err)
			break
		}
		c.l.Printf("message has been received successfully from: %v", message.TopicPartition.Offset)

		var event *UserEventCreated

		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.l.Printf("failed to unmarshal the event message: %v",err)

			consumer.CommitMessage(message)
		}

		//call user service to create profile >> TODO tomorrow connect to database 
		//profile,err := c.up.CreateUserProfile(ctx, event.userID,event.Email)

	}

}

//kafka reader

//func (c *UserEventConsumer) StartReader(ctx context.Context, reader *kafka.Reader)

func InitKafkaConsumer() (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "foo",
		"auto.offset.reset": "smallest",
	})
	if err != nil {
		log.Fatalf("something unexpected: %s", err)
	}
	return consumer, nil
}
