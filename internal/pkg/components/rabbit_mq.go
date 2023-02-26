package components

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	rabbitmqLog "walk_backend/internal/pkg/rabbitmqcustom"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

type RabbitMQConfig struct {
	URI string `yaml:"uri" env:"RABBITMQ_URI" env-description:"RabbitMQ connection URI"`
}

func (cfg *RabbitMQConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.URI, "rabbitmq-uri", cfg.URI, "RabbitMQ connection URI")
}

func (cfg *RabbitMQConfig) Validate() error {
	// TODO
	if cfg.URI == "" {
		return fmt.Errorf("invalid URI")
	}
	return nil
}

var _ ComponentInterface = (*rabbitMQComponent)(nil)

type rabbitMQComponent struct {
	rabbitmqPublisher *rabbitmq.Publisher
	rabbitMQURL       string
	m                 sync.Mutex
	stop              bool
	ready             chan struct{}
	log               zerolog.Logger
}

func NewRabbitMQ(name string, log zerolog.Logger, uri string) *rabbitMQComponent {
	return &rabbitMQComponent{
		rabbitMQURL: uri,
		ready:       make(chan struct{}, 1),
		stop:        false,
		log:         log.With().Str("component", name).Logger(),
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

	c.log.Print("Start connected to RabbitMQ")

	var err error

	c.rabbitmqPublisher, err = rabbitmq.NewPublisher(
		c.rabbitMQURL,
		rabbitmq.Config{
			Dial: amqp091.DefaultDial(30 * time.Second),
		},
		rabbitmq.WithPublisherOptionsLogger(rabbitmqLog.NewZerologLogger(c.log, c.log)),
	)
	if err != nil {
		return err
	}

	c.log.Print("Connected to RabbitMQ")

	c.stop = false

	c.ready <- struct{}{}
	close(c.ready)
	c.log.Print("Rabbit READY")

	return nil
}

func (c *rabbitMQComponent) Stop(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.stop {
		c.log.Print("Stop Rabbit component")
		// TODO Data race
		// github.com/wagslane/go-rabbitmq@v0.10.0/publish_flow_block.go:10
		// github.com/wagslane/go-rabbitmq@v0.10.0/publish_flow_block.go:29

		if err := c.rabbitmqPublisher.Close(); err != nil {
			return err
		}
		c.stop = true
		c.log.Print("Stopped Rabbit component")
	}
	return nil
}

func (c *rabbitMQComponent) Ready() <-chan struct{} {
	return c.ready
}
