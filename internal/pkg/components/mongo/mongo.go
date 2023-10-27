package mongo

import (
	"context"
	"sync"

	"walk_backend/internal/pkg/components"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var _ components.ComponentInterface = (*component)(nil)

type component struct {
	mongoClient *mongo.Client
	config      Config
	m           sync.Mutex
	stop        bool
	ready       chan struct{}
	log         zerolog.Logger
}

func New(name string, log zerolog.Logger, config Config) *component {
	return &component{
		ready:  make(chan struct{}, 1),
		config: config,
		stop:   false,
		log:    log.With().Str("component", name).Logger(),
	}
}

func (c *component) GetClient() *mongo.Client {
	return c.mongoClient
}

func (c *component) initMongoDBPing(ctx context.Context) error {
	return c.mongoClient.Ping(ctx, readpref.Primary())
}

func (c *component) Start(ctx context.Context) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.log.Print("Start connected to MongoDB")

	var err error

	mongoClientOptions := options.Client()
	mongoClientOptions.ApplyURI(c.config.URL)
	c.mongoClient, err = mongo.Connect(ctx, mongoClientOptions)
	if err != nil {
		return err
	}

	c.log.Print("Connected to MongoDB")

	c.log.Print("MongoDB PING")
	if err := c.initMongoDBPing(ctx); err != nil {
		return err
	}
	c.log.Print("MongoDB PONG")

	c.stop = false

	c.ready <- struct{}{}
	close(c.ready)
	c.log.Print("MongoDB READY")

	return nil
}

func (c *component) Stop(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.stop {
		c.log.Print("Stop MongoDB component")
		if err := c.mongoClient.Disconnect(ctx); err != nil {
			return err
		}
		c.stop = true
		c.log.Print("Stopped MongoDB component")
	}
	return nil
}

func (c *component) Ready() <-chan struct{} {
	return c.ready
}
