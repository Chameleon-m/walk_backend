package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"walk_backend/internal/app/model"
	"walk_backend/internal/app/repository"
	"walk_backend/internal/app/service"
	rabbitmqLog "walk_backend/internal/pkg/go_rabbitmq"

	rabbitmq "github.com/wagslane/go-rabbitmq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	uri            = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchange       = flag.String("exchange", "test-exchange", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType   = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	queue          = flag.String("queue", "test-queue", "Ephemeral AMQP queue name")
	bindingKey     = flag.String("binding-key", "test-key", "AMQP binding key")
	consumerTag    = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
	prefetchCount  = flag.Int("prefetch-count", 0, "Qos prefetch count")
	reconnectDelay = flag.Duration("reconnect-delay", 5*time.Second, "Reconnect delay")
	ErrLog         = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lmsgprefix)
	Log            = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lmsgprefix)
	workersCount   = flag.Int("workers-count", runtime.NumCPU(), "Workers count")
)

func init() {
	flag.Parse()
}

func main() {

	// Create context that listens for the interrupt signal from the OS.
	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// DB
	ctxMongo, cancel := context.WithCancel(ctxSignal)
	defer cancel()
	mongoClient, err := mongo.Connect(ctxMongo, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		ErrLog.Fatal(err)
	}
	defer func() {
		if err = mongoClient.Disconnect(ctxMongo); err != nil {
			ErrLog.Printf("error disconect client : %s\n", err)
		}
	}()
	if err = mongoClient.Ping(ctxMongo, readpref.Primary()); err != nil {
		ErrLog.Fatal(err)
	}
	Log.Println("connected to MongoDB")

	mongoDB := os.Getenv("MONGO_INITDB_DATABASE")

	// Consumer configuration
	consumer, err := rabbitmq.NewConsumer(
		*uri,
		rabbitmq.Config{},
		rabbitmq.WithConsumerOptionsLogger(rabbitmqLog.NewLogger(Log, ErrLog)),
		rabbitmq.WithConsumerOptionsReconnectInterval(*reconnectDelay),
	)
	if err != nil {
		ErrLog.Fatal(err)
	}
	defer consumer.Close()

	publisher, err := rabbitmq.NewPublisher(
		*uri,
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		ErrLog.Fatal(err)
	}
	defer publisher.Close()

	notifyReturn := publisher.NotifyReturn()
	notifyPublish := publisher.NotifyPublish()

	log.Println("Connected to RabbitMQ")

	ctx, cancel := context.WithCancel(ctxSignal)
	defer cancel()

	// category
	collectionCategories := mongoClient.Database(mongoDB).Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(ctx, collectionCategories)

	// place
	collectionPlaces := mongoClient.Database(mongoDB).Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(ctx, collectionPlaces)
	placeQueueRabbitRepository := repository.NewPlaceQueueRabbitRepository(publisher, notifyReturn, notifyPublish)
	placeService := service.NewDefaultPlaceService(placeMongoRepository, categoryMongoRepository, placeQueueRabbitRepository)

	done := make(chan bool, 1)
	go func() {
		select {
		case <-ctx.Done():
			log.Println("ctx done")
		// Listen for the interrupt signal.
		case <-ctxSignal.Done():
			log.Println("os signal done")
		case <-ctxMongo.Done():
			log.Println("mongo done")
		}

		done <- true
	}()

	err = consumer.StartConsuming(
		func(d rabbitmq.Delivery) rabbitmq.Action {

			// Consumer == Handler // !command Command->execute(dto DTO): bool | analog service with NewService(ctx,...)
			id, err := model.StringToID(string(d.Body))
			if err != nil {
				ErrLog.Printf("error string to id: %s", err)
				return rabbitmq.NackRequeue
			}

			Log.Printf("received a place id: %s", id)

			place, err := placeService.Find(id)
			if err != nil {
				if !d.Redelivered {
					ErrLog.Printf("error Find place with id: %s, discard", id)
					return rabbitmq.NackDiscard
				}
				ErrLog.Printf("error Find place with id: %s", id)
				return rabbitmq.NackRequeue
			}

			Log.Printf("TODO send to elastic: %s", place.ID)
			return rabbitmq.Ack
		},

		*queue,
		[]string{*bindingKey},
		rabbitmq.WithConsumeOptionsConcurrency(*workersCount),
		rabbitmq.WithConsumeOptionsQueueDurable,
		// rabbitmq.WithConsumeOptionsQuorum,
		rabbitmq.WithConsumeOptionsBindingExchangeName(*exchange),
		rabbitmq.WithConsumeOptionsBindingExchangeKind(*exchangeType),
		rabbitmq.WithConsumeOptionsBindingExchangeDurable,
		rabbitmq.WithConsumeOptionsConsumerName(*consumerTag),
		rabbitmq.WithConsumeOptionsQOSPrefetch(*prefetchCount),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Awaiting done chan
	<-done

	Log.Printf("consumer %s exiting", *consumerTag)
}