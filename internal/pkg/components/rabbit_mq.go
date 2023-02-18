package components

import (
	"context"
	"sync"
	"time"

	rabbitmqLog "walk_backend/internal/pkg/rabbitmqcustom"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

var _ ComponentInterface = (*rabbitMQComponent)(nil)

type rabbitMQComponent struct {
	rabbitmqPublisher *rabbitmq.Publisher
	rabbitMQURL       string
	m                 sync.Mutex
	stop              bool
	ready             chan struct{}
}

func NewRabbitMQ(uri string) *rabbitMQComponent {
	return &rabbitMQComponent{
		rabbitMQURL: uri,
		ready:       make(chan struct{}, 1),
		stop:        false,
	}
}

func (c *rabbitMQComponent) GetClient() *rabbitmq.Publisher {
	return c.rabbitmqPublisher
}

func (c *rabbitMQComponent) Start(ctx context.Context) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.m.Lock()
	defer c.m.Unlock()

	log.Print("Start connected to RabbitMQ")

	var err error

	c.rabbitmqPublisher, err = rabbitmq.NewPublisher(
		c.rabbitMQURL,
		rabbitmq.Config{
			Dial: amqp091.DefaultDial(30 * time.Second),
		},
		rabbitmq.WithPublisherOptionsLogger(rabbitmqLog.NewZerologLogger(log.Logger, log.Logger)),
	)
	if err != nil {
		return err
	}

	log.Print("Connected to RabbitMQ")

	c.stop = false

	c.ready <- struct{}{}
	close(c.ready)
	log.Print("Rabbit READY")

	return nil
}

func (c *rabbitMQComponent) Stop(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.stop {
		log.Print("Stop Rabbit component")
		// TODO Data race
		// github.com/wagslane/go-rabbitmq@v0.10.0/publish_flow_block.go:10
		// github.com/wagslane/go-rabbitmq@v0.10.0/publish_flow_block.go:29

		if err := c.rabbitmqPublisher.Close(); err != nil {
			log.Error().Err(err).Send()
			return err
		}
		c.stop = true
		log.Print("Stopped Rabbit component")
	}
	return nil
}

func (c *rabbitMQComponent) Ready() <-chan struct{} {
	return c.ready
}
