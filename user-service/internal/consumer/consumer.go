package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/wycliff-ochieng/internal/service"
)

type UserEventCreated struct {
	UserID    int    `json:"userId"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
}

type UserEventConsumer struct {
	l        *log.Logger
	u        *service.UserService
	consumer *kafka.Consumer
}

func NewUserEventConsumer(l *log.Logger, u *service.UserService, bootstrapServers string, groupID string) (*UserEventConsumer, error) {

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupID,
		"auto.offset.reset": "smallest",
	})
	if err != nil {
		log.Fatalf("setting up consumer hitch:%v", err)
	}

	return &UserEventConsumer{
		l:        l,
		u:        u,
		consumer: consumer,
	}, nil
}

func (c *UserEventConsumer) StartEventConsumer(ctx context.Context, topic string) {

	err := c.consumer.Subscribe(topic, nil)
	if err != nil {
		log.Fatalf("error subscribing to topic: %v", err)
	}

	for {
		select {
		//check if context has been cancelled, to enable shutting down gracefull
		case <-ctx.Done():
			c.l.Println("cancelled, shutting down")
			return

		//if context is not cancelled ,continue with polling
		default:
			ev := c.consumer.Poll(100)
			if ev == nil {
				continue //no event poll again
			}
			switch e := ev.(type) {
			case *kafka.Message:
				c.l.Printf("consumed messge from topic %s [%d] at offset %v", *e.TopicPartition.Topic, e.TopicPartition.Partition, e.TopicPartition.Offset)
				log.Printf("RAW MESSAGE CONSU?MED: %s", string(e.Value))

				var event UserEventCreated

				if err := json.Unmarshal(e.Value, &event); err != nil {
					c.l.Printf("error decoding event data:%v", err)

					c.consumer.CommitMessage(e)

					continue
				}

				log.Printf("MESSSAGE AFTERR UNMARSHALING: %+v", event)
				//call database /create profile service
				//database operation context
				opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err := c.u.CreateUserProfile(opCtx, event.UserID, event.FirstName, event.LastName, event.Email)
				if err != nil {
					log.Fatalf("Error creating userprofile: %v", err)
				}
				defer cancel()

				if err != nil {
					c.l.Printf("some error when creating user profile %d:%v", event.UserID, err)
				} else {
					c.l.Printf("successfully created user:%d", event.UserID)

					_, err := c.consumer.CommitMessage(e)
					if err != nil {
						c.l.Printf("error commit event message:%v", err)
					}
				}

			case *kafka.Error:
				//handling errors from the kafka brokers
				c.l.Printf("Kafka Error: %v(code:%d)", e, e.Code())
				if e.IsFatal() {
					c.l.Println("something Fatal")
					return
				}
			default:
				c.l.Println("event ignored")
			}

		}

	}
}

/*
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

		//handler

		var event *UserEventCreated

		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.l.Printf("failed to unmarshal the event message: %v", err)

			consumer.CommitMessage(message)
		}

		//call user service to create profile >> TODO tomorrow connect to database
		//profile,err := c.up.CreateUserProfile(ctx, event.userID,event.Email)

	}

}

//kafka reader

//func (c *UserEventConsumer) StartReader(ctx context.Context, reader *kafka.Reader)
/*
func InitKafkaConsumer(topic string)error {
	//topic := "profiles"
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "foo",
		"auto.offset.reset": "smallest",
	})
	if err != nil {
		log.Fatalf("something unexpected: %s", err)
	}

	err = consumer.Subscribe(topic, nil)
	if err != nil {
		return err
	}

	for {
		ev := consumer.Poll(100)
		switch e := ev.(type) {
		case *kafka.Message:
			fmt.Printf("consumed user created message from%v:",e.Value)
		case *kafka.Error:
			fmt.Printf("something happenned while consuming data:%v",e)
		}
	}
}
*/
