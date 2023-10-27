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
	"walk_backend/internal/pkg/env"
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

	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true}).With().Timestamp().Logger()
	logErr := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}).With().Timestamp().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if env.GetMust("GIN_MODE") == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Create context that listens for the interrupt signal from the OS.
	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ENV
	workersCount := env.GetMustInt("RABBITMQ_CONSUMERS_PLACE_REINDEX_COUNT")

	mongoURL := env.GetMust("MONGO_URL")
	mongoDB := env.GetMust("MONGO_INITDB_NAME")

	rabbitmqURL := env.GetMust("RABBITMQ_URL")
	consumerTag := env.GetMust("RABBITMQ_CONSUMERS_PLACE_REINDEX_TAG")
	exchange := env.GetMust("RABBITMQ_EXCHANGE_REINDEX")
	exchangeType := env.GetMust("RABBITMQ_EXCHANGE_TYPE")
	queue := env.GetMust("RABBITMQ_QUEUE_PLACE_REINDEX")
	routingKey := env.GetMust("RABBITMQ_ROUTING_PLACE_KEY")

	// DB
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		logErr.Fatal().Err(err).Caller().Send()
	}
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			log.Info().Err(err).Caller().Send()
		}
	}()
	if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		logErr.Fatal().Err(err).Caller().Send()
	}
	log.Print("Сonnected to MongoDB")

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
	defer func() {
		if err = consumer.Close(); err != nil {
			log.Info().Err(err).Caller().Send()
		}
	}()

	publisher, err := rabbitmq.NewPublisher(
		rabbitmqURL,
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		logErr.Fatal().Err(err).Caller().Send()
	}
	defer func() {
		if err = publisher.Close(); err != nil {
			log.Info().Err(err).Caller().Send()
		}
	}()

	log.Print("Connected to RabbitMQ")

	// category
	collectionCategories := mongoClient.Database(mongoDB).Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(collectionCategories)

	// place
	collectionPlaces := mongoClient.Database(mongoDB).Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(collectionPlaces)
	placeQueueRabbitRepository := repository.NewPlaceQueueRabbitRepository(ctx, publisher, exchange, routingKey)
	placeService := service.NewDefaultPlaceService(placeMongoRepository, categoryMongoRepository, placeQueueRabbitRepository, nil, nil)

	done := make(chan struct{}, 1)
	go func() {
		select {
		case <-ctxSignal.Done():
			log.Print("os signal done")
		case <-ctx.Done():
			log.Print("ctx done")
		}

		done <- struct{}{}
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

			place, err := placeService.Find(ctx, id)
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
