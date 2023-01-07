package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"walk_backend/internal/app/model"
	"walk_backend/internal/app/repository"
	"walk_backend/internal/app/service"
	rabbitmqLog "walk_backend/internal/pkg/rabbitmqcustom"

	rabbitmq "github.com/wagslane/go-rabbitmq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	prefetchCount  = flag.Int("prefetch-count", 0, "Qos prefetch count")
	reconnectDelay = flag.Duration("reconnect-delay", 5*time.Second, "Reconnect delay")
	errLog         = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lmsgprefix)
	infLog         = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lmsgprefix)
)

func init() {
	flag.Parse()
}

func main() {

	// Create context that listens for the interrupt signal from the OS.
	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ENV
	workersCount, err := strconv.Atoi(os.Getenv("RABBITMQ_CONSUMERS_PLACE_REINDEX_COUNT"))
	if err != nil {
		errLog.Fatalf("ENV RABBITMQ_CONSUMERS_PLACE_REINDEX_COUNT: %s", err)
	}

	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_INITDB_DATABASE")

	rabbitmqURL := os.Getenv("RABBITMQ_URI")
	consumerTag := os.Getenv("RABBITMQ_CONSUMERS_PLACE_REINDEX_TAG")
	exchange := os.Getenv("RABBITMQ_EXCHANGE_REINDEX")
	exchangeType := os.Getenv("RABBITMQ_EXCHANGE_TYPE")
	queue := os.Getenv("RABBITMQ_QUEUE_PLACE_REINDEX")
	routingKey := os.Getenv("RABBITMQ_ROUTING_PLACE_KEY")

	// DB
	ctxMongo, cancel := context.WithCancel(ctxSignal)
	defer cancel()
	mongoClient, err := mongo.Connect(ctxMongo, options.Client().ApplyURI(mongoURI))
	if err != nil {
		errLog.Fatal(err)
	}
	defer func() {
		if err = mongoClient.Disconnect(ctxMongo); err != nil {
			errLog.Printf("error disconect client : %s\n", err)
		}
	}()
	if err = mongoClient.Ping(ctxMongo, readpref.Primary()); err != nil {
		errLog.Fatal(err)
	}
	infLog.Println("connected to MongoDB")

	// Consumer configuration
	consumer, err := rabbitmq.NewConsumer(
		rabbitmqURL,
		rabbitmq.Config{},
		rabbitmq.WithConsumerOptionsLogger(rabbitmqLog.NewLogger(infLog, errLog)),
		rabbitmq.WithConsumerOptionsReconnectInterval(*reconnectDelay),
	)
	if err != nil {
		errLog.Fatal(err)
	}
	defer consumer.Close()

	publisher, err := rabbitmq.NewPublisher(
		rabbitmqURL,
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		errLog.Fatal(err)
	}
	defer publisher.Close()

	notifyReturn := publisher.NotifyReturn()
	notifyPublish := publisher.NotifyPublish()

	infLog.Println("Connected to RabbitMQ")

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
				errLog.Printf("error string to id: %s", err)
				return rabbitmq.NackRequeue
			}

			infLog.Printf("received a place id: %s", id)

			place, err := placeService.Find(id)
			if err != nil {
				if !d.Redelivered {
					errLog.Printf("error Find place with id: %s, discard", id)
					return rabbitmq.NackDiscard
				}
				errLog.Printf("error Find place with id: %s", id)
				return rabbitmq.NackRequeue
			}

			infLog.Printf("TODO send to elastic: %s", place.ID)
			return rabbitmq.Ack
		},

		queue,
		[]string{routingKey},
		rabbitmq.WithConsumeOptionsConcurrency(workersCount),
		rabbitmq.WithConsumeOptionsQueueDurable,
		// rabbitmq.WithConsumeOptionsQuorum,
		rabbitmq.WithConsumeOptionsBindingExchangeName(exchange),
		rabbitmq.WithConsumeOptionsBindingExchangeKind(exchangeType),
		rabbitmq.WithConsumeOptionsBindingExchangeDurable,
		rabbitmq.WithConsumeOptionsConsumerName(consumerTag),
		rabbitmq.WithConsumeOptionsQOSPrefetch(*prefetchCount),
	)
	if err != nil {
		errLog.Fatal(err)
	}

	// Awaiting done chan
	<-done

	infLog.Printf("consumer %s exiting", consumerTag)
}
