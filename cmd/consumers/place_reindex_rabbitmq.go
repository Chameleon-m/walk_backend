package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"walk_backend/model"
	"walk_backend/repository"
	"walk_backend/service"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var channelAmqp *amqp.Channel

func init() {

}

func main() {

	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// DB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	mongoDB := os.Getenv("MONGO_INITDB_DATABASE")

	// QUEUE
	amqpConnection, err := amqp.Dial(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if !amqpConnection.IsClosed() {
			if err = amqpConnection.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}()
	log.Println("Connected to RabbitMQ")

	channelAmqp, err := amqpConnection.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if !channelAmqp.IsClosed() {
			if err = channelAmqp.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}()

	go func() {
		for {
			select {
			case err := <-amqpConnection.NotifyClose(make(chan *amqp.Error)):
				if err != nil {
					log.Printf("RabbitMQ | Connection closing: %s", err)
					stop()
				}
			case err := <-channelAmqp.NotifyClose(make(chan *amqp.Error)):
				if err != nil {
					log.Printf("RabbitMQ | Channel closing: %s", err)
					stop()
				}
			}
		}
	}()

	// category
	collectionCategories := client.Database(mongoDB).Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(ctx, collectionCategories)

	// place
	collectionPlaces := client.Database(mongoDB).Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(ctx, collectionPlaces)
	placeQueueRabbitRepository := repository.NewPlaceQueueRabbitRepository(channelAmqp)
	placeService := service.NewDefaultPlaceService(placeMongoRepository, categoryMongoRepository, placeQueueRabbitRepository)

	err = channelAmqp.Qos(1, 0, false)
	if err != nil {
		log.Fatal(err)
	}
	msgs, err := channelAmqp.Consume(os.Getenv("RABBITMQ_QUEUE_PLACE_REINDEX"), "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for delivery := range msgs {

			if delivery.ContentType != "text/plain" {
				log.Printf("Error ContentType: %s", delivery.ContentType)
				continue
			}

			log.Printf("Received a place: %s ", delivery.Body)

			// TODO TO SERVICE

			id, err := model.StringToID(string(delivery.Body))
			if err != nil {
				log.Printf("Error string to id: %s", err)
				continue
			}

			log.Printf("Received a place id: %s", id)

			place, err := placeService.Find(id)
			if err != nil {
				if !delivery.Redelivered {
					delivery.Reject(true)
				}
				log.Printf("Error Find place with id: %s", id)
				continue
			}

			log.Printf("TODO send to elastic: %s", place.ID)
			delivery.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-ctxSignal.Done()
	stop()
	log.Println("Consumer exiting")
}
