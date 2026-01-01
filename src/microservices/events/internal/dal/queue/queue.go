package queue

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Queue struct {
	producer     *kafka.Producer
	consumer     *kafka.Consumer
	kafkaBrokers string
}

func (q *Queue) Connect() error {
	var err error
	kafkaBrokers := getEnv("KAFKA_BROKERS", "kafka:9092")
	q.producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBrokers,
	})
	if err != nil {
		return fmt.Errorf("producer connection error: %w", err)
	}

	q.consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBrokers,
		"group.id":          "eventsGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return fmt.Errorf("consumer connection error: %w", err)
	}

	return nil
}

func (q *Queue) Close() {
	q.producer.Close()
	q.consumer.Close()
}

func (q *Queue) SendEvents(ctx context.Context, eventType string, eventPayload []byte) error {
	topic := eventType
	err := q.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          eventPayload,
		Timestamp:      time.Now(),
		TimestampType:  kafka.TimestampCreateTime,
	}, nil)
	if err != nil {
		return err
	}

	// Дождаться доставки сообщений перед закрытием продюсера
	q.producer.Flush(15 * 1000)

	return nil
}

func (q *Queue) ReceiveEvents(ctx context.Context) error {
	topics := []string{"movie-events", "user-events", "payment-events"}
	err := q.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	for {
		msg, err := q.consumer.ReadMessage(-1)
		if err == nil {
			log.Printf("Received message: %s\n", string(msg.Value))
		} else {
			// Обработка ошибок при чтении сообщений
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func New(kafkaBrokers string) *Queue {
	return &Queue{kafkaBrokers: kafkaBrokers}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
