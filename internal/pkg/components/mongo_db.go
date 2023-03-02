package components

import (
	"context"
	"flag"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDBConfig struct {
	InitDBName string `yaml:"init_db_name" env:"MONGO_INITDB_NAME" env-description:"Mongo init database"`
	URI        string `yaml:"uri"          env:"MONGO_URI"         env-description:"Mongo connection URI"`
}

func (cfg *MongoDBConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.InitDBName, "mongo-init-db", cfg.InitDBName, "Mongo init database")
	fs.StringVar(&cfg.URI, "mongo-uri", cfg.URI, "Mongo connection URI")
}

func (cfg *MongoDBConfig) Validate() error {
	// TODO
	if cfg.InitDBName == "" {
		return fmt.Errorf("invalid init DB name")
	} else if cfg.URI == "" {
		return fmt.Errorf("invalid URI")
	}
	return nil
}

var _ ComponentInterface = (*mongoDBComponent)(nil)

type mongoDBComponent struct {
	mongoClient *mongo.Client
	mongoURI    string
	m           sync.Mutex
	stop        bool
	ready       chan struct{}
	log         zerolog.Logger
}

func NewMongoDB(name string, log zerolog.Logger, mongoURI string) *mongoDBComponent {
	return &mongoDBComponent{
		ready:    make(chan struct{}, 1),
		mongoURI: mongoURI,
		stop:     false,
		log:      log.With().Str("component", name).Logger(),
	}
}

func (c *mongoDBComponent) GetClient() *mongo.Client {
	return c.mongoClient
}

func (c *mongoDBComponent) initMongoDBPing(ctx context.Context) error {
	return c.mongoClient.Ping(ctx, readpref.Primary())
}

func (c *mongoDBComponent) Start(ctx context.Context) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.log.Print("Start connected to MongoDB")

	var err error

	mongoClientOptions := options.Client()
	mongoClientOptions.ApplyURI(c.mongoURI)
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

func (c *mongoDBComponent) Stop(ctx context.Context) error {
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

func (c *mongoDBComponent) Ready() <-chan struct{} {
	return c.ready
}
