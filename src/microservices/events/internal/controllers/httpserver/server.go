package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"

	"github.com/voitenkov-courses/architecture-pro-cinemaabyss/src/microservices/events/api"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ api.ServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}
}

// Проверка работоспособности микросервиса событий
// (GET /api/events/health)
func (Server) GetEventsServiceHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

// Создание события фильма
// (POST /api/events/movie)
func (Server) CreateMovieEvent(c *gin.Context) {
	var movieEvent api.MovieEvent
	if err := json.NewDecoder(c.Request.Body).Decode(&movieEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := json.Marshal(movieEvent)
	if err == nil {
		sendEvents("movie-events", event)
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
	})
}

// Создание события платежа
// (POST /api/events/payment)
func (Server) CreatePaymentEvent(c *gin.Context) {
	var paymentEvent api.PaymentEvent
	if err := json.NewDecoder(c.Request.Body).Decode(&paymentEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := json.Marshal(paymentEvent)
	if err == nil {
		sendEvents("payment-events", event)
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
	})
}

// Создание события пользователя
// (POST /api/events/user)
func (Server) CreateUserEvent(c *gin.Context) {
	var userEvent api.UserEvent
	if err := json.NewDecoder(c.Request.Body).Decode(&userEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := json.Marshal(userEvent)
	if err == nil {
		sendEvents("user-events", event)
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
	})
}

func sendEvents(eventType string, eventPayload []byte) error {
	kafkaBrokers := getEnv("KAFKA_BROKERS", "kafka:9092")

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBrokers,
	})
	if err != nil {
		return fmt.Errorf("producer connection error: %w", err)
	}

	defer producer.Close()

	topic := eventType
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          eventPayload,
		Timestamp:      time.Now(),
		TimestampType:  kafka.TimestampCreateTime,
	}, nil)
	if err != nil {
		return err
	}

	// Дождаться доставки сообщений перед закрытием продюсера
	producer.Flush(1 * 1000)

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
