package components

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var _ ComponentInterface = (*mongoDBComponent)(nil)

type mongoDBComponent struct {
	mongoClient *mongo.Client
	mongoURI    string
	m           sync.Mutex
	stop        bool
	ready       chan struct{}
}

func NewMongoDB(mongoURI string) *mongoDBComponent {
	return &mongoDBComponent{
		ready:    make(chan struct{}, 1),
		mongoURI: mongoURI,
		stop:     false,
	}
}

func (c *mongoDBComponent) GetClient() *mongo.Client {
	return c.mongoClient
}

func (c *mongoDBComponent) initMongoDBPing(ctx context.Context) error {
	return c.mongoClient.Ping(ctx, readpref.Primary())
}

func (c *mongoDBComponent) Start(ctx context.Context) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.m.Lock()
	defer c.m.Unlock()

	log.Print("Start connected to MongoDB")

	var err error

	mongoClientOptions := options.Client()
	mongoClientOptions.ApplyURI(c.mongoURI)
	c.mongoClient, err = mongo.Connect(ctx, mongoClientOptions)
	if err != nil {
		return err
	}

	log.Print("Connected to MongoDB")

	log.Print("MongoDB PING")
	if err := c.initMongoDBPing(ctx); err != nil {
		return err
	}
	log.Print("MongoDB PONG")

	c.stop = false

	c.ready <- struct{}{}
	close(c.ready)
	log.Print("MongoDB READY")

	return nil
}

func (c *mongoDBComponent) Stop(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.stop {
		log.Print("Stop MongoDB component")
		if err := c.mongoClient.Disconnect(ctx); err != nil {
			log.Error().Err(err).Send()
			return err
		}
		c.stop = true
		log.Print("Stopped MongoDB component")
	}
	return nil
}

func (c *mongoDBComponent) Ready() <-chan struct{} {
	return c.ready
}
