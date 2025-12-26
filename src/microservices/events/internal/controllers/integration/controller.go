package integration

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Integration struct {
	queue        Queue
	kafkaBrokers string
}

type Queue interface {
	Connect() error
	Close()
	ReceiveEvents(ctx context.Context) error
}

func New(queue Queue) *Integration {
	return &Integration{
		queue: queue,
	}
}

func (i *Integration) Start(ctx context.Context) error {
	defer i.queue.Close()

	go func() {
		err := i.queue.Connect()
		if err != nil {
			log.Println(err)
			return
		}

		{
			err = i.queue.ReceiveEvents(ctx)
			time.Sleep(10 * time.Second)
		}
	}()

	<-ctx.Done()
	fmt.Println(ctx.Err())
	return nil
}
