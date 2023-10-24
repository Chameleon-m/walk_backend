package rabbitmq

import (
	"context"
	"sync"
	"time"

	"walk_backend/internal/pkg/components"
	rabbitmqLog "walk_backend/internal/pkg/rabbitmqcustom"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/wagslane/go-rabbitmq"
)

const (
	maxChannelMax = (2 << 15) - 1

	defaultHeartbeat         = 10 * time.Second
	defaultConnectionTimeout = 30 * time.Second
	defaultChannelMax        = (2 << 10) - 1
)

var _ components.ComponentInterface = (*component)(nil)

type component struct {
	rabbitmqPublisher *rabbitmq.Publisher
	rabbitMQURL       string
	vhost             string
	channelMax        int
	frameSize         int
	heartbeat         time.Duration
	dial              time.Duration
	m                 sync.Mutex
	stop              bool
	ready             chan struct{}
	log               zerolog.Logger
}

type options struct {
	vhost      *string
	channelMax *int
	frameSize  *int
	heartbeat  *time.Duration
	dial       *time.Duration
}

type Option func(options *options) error

func WithVhost(vhost string) Option {
	return func(options *options) error {
		options.vhost = &vhost
		return nil
	}
}

func WithChannelMax(channelMax int) Option {
	return func(options *options) error {
		if channelMax < 0 {
			return ErrInvalidChannelMax
		} else if channelMax > maxChannelMax {
			return ErrInvalidChannelMaxLimit
		}
		options.channelMax = &channelMax
		return nil
	}
}

func WithFrameSize(frameSize int) Option {
	return func(options *options) error {
		if frameSize < 0 {
			return ErrInvalidFrameSize
		}
		options.frameSize = &frameSize
		return nil
	}
}

func WithHeartbeat(heartbeat time.Duration) Option {
	return func(options *options) error {
		if heartbeat < 0 {
			return ErrInvalidHeartbeat
		}
		options.heartbeat = &heartbeat
		return nil
	}
}

func WithDial(dial time.Duration) Option {
	return func(options *options) error {
		if dial < 0 {
			return ErrInvalidDial
		}
		options.dial = &dial
		return nil
	}
}

func New(name string, log zerolog.Logger, uri string, opts ...Option) (*component, error) {
	var options options
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	// Vhost
	var vhost string
	if options.vhost == nil {
		vhost = "/"
	} else {
		vhost = *options.vhost
	}

	// ChannelMax
	var channelMax int
	if options.channelMax == nil {
		channelMax = defaultChannelMax
	} else {
		if *options.channelMax == 0 {
			channelMax = maxChannelMax
		} else {
			channelMax = *options.channelMax
		}
	}

	// FrameSize
	var frameSize int
	if options.frameSize == nil {
		frameSize = 0 // 0 max bytes means unlimited
	} else {
		frameSize = *options.frameSize
	}

	// Heartbeat
	var heartbeat time.Duration
	if options.heartbeat == nil {
		heartbeat = defaultHeartbeat
	} else {
		heartbeat = *options.heartbeat
	}

	// Dial
	var dial time.Duration
	if options.dial == nil {
		dial = defaultConnectionTimeout
	} else {
		dial = *options.dial
	}

	return &component{
		rabbitMQURL: uri,
		vhost:       vhost,
		channelMax:  channelMax,
		frameSize:   frameSize,
		heartbeat:   heartbeat,
		dial:        dial,
		ready:       make(chan struct{}, 1),
		stop:        false,
		log:         log.With().Str("component", name).Logger(),
	}, nil
}

func (c *component) GetClient() *rabbitmq.Publisher {
	return c.rabbitmqPublisher
}

func (c *component) Start(ctx context.Context) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.log.Print("Start connected to RabbitMQ")

	var err error

	c.rabbitmqPublisher, err = rabbitmq.NewPublisher(
		c.rabbitMQURL,
		rabbitmq.Config{
			Vhost:      c.vhost,
			ChannelMax: c.channelMax,
			FrameSize:  c.frameSize,
			Heartbeat:  c.heartbeat,
			Dial:       amqp091.DefaultDial(c.dial),
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

func (c *component) Stop(ctx context.Context) error {
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

func (c *component) Ready() <-chan struct{} {
	return c.ready
}
