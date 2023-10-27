package redis

import (
	"context"
	"crypto/tls"
	"sync"
	"walk_backend/internal/pkg/components"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var _ components.ComponentInterface = (*component)(nil)

type component struct {
	redisClient           *redis.Client
	config                Config
	onConnect             func(ctx context.Context, cn *redis.Conn) error
	credentialsProvider   func() (username string, password string)
	contextTimeoutEnabled bool
	maxActiveConns        int
	TLSConfig             *tls.Config
	limiter               redis.Limiter
	disableIndentity      bool
	m                     sync.Mutex
	stop                  bool
	ready                 chan struct{}
	log                   zerolog.Logger
}

type options struct {
	onConnect             func(ctx context.Context, cn *redis.Conn) error
	credentialsProvider   func() (username string, password string)
	contextTimeoutEnabled *bool
	maxActiveConns        *int
	TLSConfig             *tls.Config
	limiter               redis.Limiter
	disableIndentity      *bool
}

type Option func(options *options) error

func WithOnConnect(onConnect func(ctx context.Context, cn *redis.Conn) error) Option {
	return func(options *options) error {
		if onConnect == nil {
			return ErrInvalidOnConnect
		}
		options.onConnect = onConnect
		return nil
	}
}

func WithCredentialsProvider(credentialsProvider func() (username string, password string)) Option {
	return func(options *options) error {
		if credentialsProvider == nil {
			return ErrInvalidCredentialsProvider
		}
		options.credentialsProvider = credentialsProvider
		return nil
	}
}

func WithContextTimeoutEnabled(contextTimeoutEnabled bool) Option {
	return func(options *options) error {
		options.contextTimeoutEnabled = &contextTimeoutEnabled
		return nil
	}
}

func WithMaxActiveConns(maxActiveConns int) Option {
	return func(options *options) error {
		if maxActiveConns < 0 {
			return ErrInvalidMaxActiveConns
		}
		options.maxActiveConns = &maxActiveConns
		return nil
	}
}

func WithTLSConfig(config *tls.Config) Option {
	return func(options *options) error {
		if config == nil {
			return ErrInvalidTLSConfig
		}
		options.TLSConfig = config
		return nil
	}
}

func WithLimiter(limiter redis.Limiter) Option {
	return func(options *options) error {
		if limiter == nil {
			return ErrInvalidLimiter
		}
		options.limiter = limiter
		return nil
	}
}

func WithDisableIndentity(disableIndentity bool) Option {
	return func(options *options) error {
		options.disableIndentity = &disableIndentity
		return nil
	}
}

func New(name string, log zerolog.Logger, config Config, opts ...Option) (*component, error) {
	var options options
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	// OnConnect
	onConnect := options.onConnect

	// CredentialsProvider
	credentialsProvider := options.credentialsProvider

	// ContextTimeoutEnabled
	var contextTimeoutEnabled bool
	if options.contextTimeoutEnabled == nil {
		contextTimeoutEnabled = false
	} else {
		contextTimeoutEnabled = *options.contextTimeoutEnabled
	}

	// MaxActiveConns
	var maxActiveConns int
	if options.maxActiveConns == nil {
		maxActiveConns = 0
	} else {
		maxActiveConns = *options.maxActiveConns
	}

	// TLSConfig
	var TLSConfig *tls.Config
	if options.TLSConfig != nil {
		TLSConfig = options.TLSConfig
	}

	// Limiter
	var limiter redis.Limiter
	if options.limiter != nil {
		limiter = options.limiter
	}

	// DisableIndentity
	var disableIndentity bool
	if options.disableIndentity == nil {
		disableIndentity = false
	} else {
		disableIndentity = *options.disableIndentity
	}

	return &component{
		config:                config,
		onConnect:             onConnect,
		credentialsProvider:   credentialsProvider,
		contextTimeoutEnabled: contextTimeoutEnabled,
		maxActiveConns:        maxActiveConns,
		TLSConfig:             TLSConfig,
		limiter:               limiter,
		disableIndentity:      disableIndentity,
		stop:                  false,
		ready:                 make(chan struct{}, 1),
		log:                   log.With().Str("component", name).Logger(),
	}, nil
}

func (c *component) GetClient() *redis.Client {
	return c.redisClient
}

func (c *component) Start(ctx context.Context) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.log.Print("Start connected to Redis")

	opt, err := redis.ParseURL(c.config.URL)
	if err != nil {
		return err
	}
	if c.onConnect != nil {
		opt.OnConnect = c.onConnect
	}
	if c.credentialsProvider != nil {
		opt.CredentialsProvider = c.credentialsProvider
	}
	opt.ContextTimeoutEnabled = c.contextTimeoutEnabled
	opt.MaxActiveConns = c.maxActiveConns
	if c.TLSConfig != nil {
		opt.TLSConfig = c.TLSConfig
	}
	if c.limiter != nil {
		opt.Limiter = c.limiter
	}
	opt.DisableIndentity = c.disableIndentity
	c.redisClient = redis.NewClient(opt)

	c.log.Print("Connected to Redis")

	c.log.Print("Redis status: PING")
	status, err := c.redisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}
	c.log.Print("Redis status: " + status)

	c.stop = false

	c.ready <- struct{}{}
	close(c.ready)
	c.log.Print("Redis READY")

	return nil
}

func (c *component) Stop(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.stop {
		c.log.Print("Stop Redis component")
		if err := c.redisClient.Close(); err != nil {
			return err
		}
		c.stop = true
		c.log.Print("Stopped Redis component")
	}
	return nil
}

func (c *component) Ready() <-chan struct{} {
	return c.ready
}
