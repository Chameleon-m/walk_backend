package components

import (
	"context"
	"flag"
	"fmt"
	"net"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type RedisConfig struct {
	Host     string `yaml:"host"     env:"REDIS_HOST"     env-description:"Redis host"`
	Port     string `yaml:"port"     env:"REDIS_PORT"     env-description:"Redis port"`
	Username string `yaml:"username" env:"REDIS_USERNAME" env-description:"Redis username"`
	Password string `yaml:"password" env:"REDIS_PASSWORD" env-description:"Redis password"`
}

func (cfg *RedisConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.Host, "redis-host", cfg.Host, "Redis host")
	fs.StringVar(&cfg.Port, "redis-port", cfg.Port, "Redis port")
	fs.StringVar(&cfg.Username, "redis-username", cfg.Username, "Redis username")
	fs.StringVar(&cfg.Password, "redis-password", cfg.Password, "Redis password")
}

func (cfg *RedisConfig) Validate() error {
	// TODO
	if cfg.Host == "" {
		return fmt.Errorf("invalid host")
	} else if cfg.Port == "" {
		return fmt.Errorf("invalid port")
	} else if cfg.Username == "" {
		return fmt.Errorf("invalid username")
	} else if cfg.Password == "" {
		return fmt.Errorf("invalid password")
	}
	return nil
}

var _ ComponentInterface = (*redisComponent)(nil)

type redisComponent struct {
	redisClient *redis.Client
	host        string
	port        string
	username    string
	password    string
	m           sync.Mutex
	stop        bool
	ready       chan struct{}
	log         zerolog.Logger
}

func NewRedis(name string, log zerolog.Logger, host, port, username, password string) *redisComponent {
	return &redisComponent{
		host:     host,
		port:     port,
		username: username,
		password: password,
		stop:     false,
		ready:    make(chan struct{}, 1),
		log:      log.With().Str("component", name).Logger(),
	}
}

func (c *redisComponent) GetClient() *redis.Client {
	return c.redisClient
}

func (c *redisComponent) Start(ctx context.Context) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.log.Print("Start connected to Redis")

	c.redisClient = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(c.host, c.port),
		Username: c.username,
		Password: c.password,
		DB:       0, // use default DB
	})

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

func (c *redisComponent) Stop(ctx context.Context) error {
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

func (c *redisComponent) Ready() <-chan struct{} {
	return c.ready
}
