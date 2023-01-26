package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"walk_backend/internal/app/model"
	"walk_backend/internal/app/repository"
	"walk_backend/internal/app/service"
	"walk_backend/internal/pkg/component/env"
	rabbitmqLog "walk_backend/internal/pkg/rabbitmqcustom"

	"github.com/rs/zerolog"
	rabbitmq "github.com/wagslane/go-rabbitmq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	prefetchCount  = flag.Int("prefetch-count", 0, "Qos prefetch count")
	reconnectDelay = flag.Duration("reconnect-delay", 5*time.Second, "Reconnect delay")
)

func init() {
	flag.Parse()
}

func main() {

	env := env.New()

	// zerolog.TimestampFieldName = "t"
	// zerolog.LevelFieldName = "l"
	// zerolog.MessageFieldName = "m"
	// zerolog.ErrorFieldName = "e"
	// zerolog.CallerFieldName = "c"
	// zerolog.ErrorStackFieldName = "s"
	// zerolog.DisableSampling(true)

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logErr := zerolog.New(os.Stderr).With().Timestamp().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if env.GetMust("GIN_MODE") == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Create context that listens for the interrupt signal from the OS.
	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ENV
	workersCount := env.GetMustInt("RABBITMQ_CONSUMERS_PLACE_REINDEX_COUNT")

	mongoURI := env.GetMust("MONGO_URI")
	mongoDB := env.GetMust("MONGO_INITDB_DATABASE")

	rabbitmqURL := env.GetMust("RABBITMQ_URI")
	consumerTag := env.GetMust("RABBITMQ_CONSUMERS_PLACE_REINDEX_TAG")
	exchange := env.GetMust("RABBITMQ_EXCHANGE_REINDEX")
	exchangeType := env.GetMust("RABBITMQ_EXCHANGE_TYPE")
	queue := env.GetMust("RABBITMQ_QUEUE_PLACE_REINDEX")
	routingKey := env.GetMust("RABBITMQ_ROUTING_PLACE_KEY")

	// DB
	ctxMongo, cancel := context.WithCancel(ctxSignal)
	defer cancel()
	mongoClient, err := mongo.Connect(ctxMongo, options.Client().ApplyURI(mongoURI))
	if err != nil {
		logErr.Fatal().Err(err).Caller().Send()
	}
	defer func() {
		if err = mongoClient.Disconnect(ctxMongo); err != nil {
			log.Info().Err(err).Caller().Send()
		}
	}()
	if err = mongoClient.Ping(ctxMongo, readpref.Primary()); err != nil {
		logErr.Fatal().Err(err).Caller().Send()
	}
	log.Print("connected to MongoDB")

	// Consumer configuration
	consumer, err := rabbitmq.NewConsumer(
		rabbitmqURL,
		rabbitmq.Config{},
		rabbitmq.WithConsumerOptionsLogger(rabbitmqLog.NewZerologLogger(log, logErr)),
		rabbitmq.WithConsumerOptionsReconnectInterval(*reconnectDelay),
	)
	if err != nil {
		logErr.Fatal().Err(err).Caller().Send()
	}
	defer consumer.Close()

	publisher, err := rabbitmq.NewPublisher(
		rabbitmqURL,
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		logErr.Fatal().Err(err).Caller().Send()
	}
	defer publisher.Close()

	notifyReturn := publisher.NotifyReturn()
	notifyPublish := publisher.NotifyPublish()

	log.Print("Connected to RabbitMQ")

	ctx, cancel := context.WithCancel(ctxSignal)
	defer cancel()

	// category
	collectionCategories := mongoClient.Database(mongoDB).Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(ctx, collectionCategories)

	// place
	collectionPlaces := mongoClient.Database(mongoDB).Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(ctx, collectionPlaces)
	placeQueueRabbitRepository := repository.NewPlaceQueueRabbitRepository(publisher, notifyReturn, notifyPublish)
	placeService := service.NewDefaultPlaceService(placeMongoRepository, categoryMongoRepository, placeQueueRabbitRepository, nil, nil)

	done := make(chan bool, 1)
	go func() {
		select {
		// Listen for the interrupt signal.
		case <-ctxSignal.Done():
			log.Print("os signal done")
		case <-ctx.Done():
			log.Print("ctx done")
		case <-ctxMongo.Done():
			log.Print("mongo done")
		}

		done <- true
	}()

	err = consumer.StartConsuming(
		func(d rabbitmq.Delivery) rabbitmq.Action {

			// Consumer == Handler // !command Command->execute(dto DTO): bool | analog service with NewService(ctx,...)
			id, err := model.StringToID(string(d.Body))
			if err != nil {
				logErr.Error().Err(err).Caller().Send()
				return rabbitmq.NackRequeue
			}

			log.Printf("received a place id: %s", id)

			place, err := placeService.Find(id)
			if err != nil {
				if !d.Redelivered {
					logErr.Error().Err(err).Caller().Str("id", id.String()).Msg("discard")
					return rabbitmq.NackDiscard
				}
				logErr.Error().Err(err).Caller().Str("id", id.String()).Msg("requeue")
				return rabbitmq.NackRequeue
			}

			log.Printf("TODO send to elastic: %s", place.ID)
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
		logErr.Fatal().Err(err).Caller().Send()
	}

	// Awaiting done chan
	<-done

	log.Printf("consumer %s exiting", consumerTag)
}
