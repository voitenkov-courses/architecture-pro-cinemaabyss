package queue

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	smarthomeintegration "github.com/voitenkov-courses/architecture-pro-warmhouse/apps/device-management/internal/controllers/smarthome-integration"
)

type Queue struct {
	conn *amqp.Connection
}

var rmqqueue amqp.Queue

func (q *Queue) Connect() error {
	var err error
	q.conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	return nil
}

func (q *Queue) Close() error {
	return q.conn.Close()
}

func (q *Queue) ReceiveAndProcessSensorUpdates(ctx context.Context, fn smarthomeintegration.CallbackFunc) error {
	queueName := "sensorupdates"
	channel, err := q.conn.Channel()
	if err != nil {
		return fmt.Errorf("channel error: %w", err)
	}

	if rmqqueue, err = channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	); err != nil {
		return fmt.Errorf("queue declaration error: %w", err)
	}

	telemetry, err := channel.ConsumeWithContext(
		ctx,
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed open channel: %w", err)
	}

	go func() {
		for telemetryMessage := range telemetry {
			fn(ctx, telemetryMessage.Body)
		}
	}()

	<-ctx.Done()
	return nil
}

func New() *Queue {
	return &Queue{}
}
