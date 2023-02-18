package components

import (
	"context"
	"net"
	"sync"

	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog/log"
)

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
}

func NewRedis(host, port, username, password string) *redisComponent {
	return &redisComponent{
		host:     host,
		port:     port,
		username: username,
		password: password,
		stop:     false,
		ready:    make(chan struct{}, 1),
	}
}

func (c *redisComponent) GetClient() *redis.Client {
	return c.redisClient
}

func (c *redisComponent) Start(ctx context.Context) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.m.Lock()
	defer c.m.Unlock()

	log.Print("Start connected to Redis")

	c.redisClient = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(c.host, c.port),
		Username: c.username,
		Password: c.password,
		DB:       0, // use default DB
	})

	log.Print("Connected to Redis")

	log.Print("Redis status: PING")
	status, err := c.redisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}
	log.Print("Redis status: " + status)

	c.stop = false

	c.ready <- struct{}{}
	close(c.ready)
	log.Print("Redis READY")

	return nil
}

func (c *redisComponent) Stop(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.stop {
		log.Print("Stop Redis component")
		if err := c.redisClient.Close(); err != nil {
			log.Error().Err(err).Send()
			return err
		}
		c.stop = true
		log.Print("Stopped Redis component")
	}
	return nil
}

func (c *redisComponent) Ready() <-chan struct{} {
	return c.ready
}
