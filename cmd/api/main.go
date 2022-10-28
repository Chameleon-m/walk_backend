// Places API
//
// This is a places API.
//
//	Schemes: http
//	Host: localhost:8080
//	BasePath: /v1/api
//	Version: 0.0.1
//	Contact: Dmitry Korolev <korolev.d.l@yandex.ru> https://github.com/Chameleon-m
//
//	SecurityDefinitions:
//	    api_key:
//	      type: apiKey
//	      name: Authorization
//	      in: header
//
//	Consumes:
//	  - application/json
//
//	Produces:
//	  - application/json
//
// swagger:meta
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"walk_backend/cmd/api/handlers"
	"walk_backend/cmd/api/middleware"
	"walk_backend/cmd/api/presenter"
	"walk_backend/repository"
	"walk_backend/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var authHandler *handlers.AuthHandler
var placesHandler *handlers.PlacesHandler
var categoriesHandler *handlers.CategoriesHandler

func init() {

}

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var err error

	// DB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	failOnError(err, "MongoDB")
	defer func() {
		failOnError(client.Disconnect(ctx), "MongoDB")
	}()

	failOnError(client.Ping(context.TODO(), readpref.Primary()), "MongoDB | Ping")
	log.Println("Connected to MongoDB")

	mongoDB := os.Getenv("MONGO_INITDB_DATABASE")

	// RabbitMQ
	amqpConnection, err := amqp.Dial(os.Getenv("RABBITMQ_URI"))
	failOnError(err, "RabbitMQ")
	defer func() {
		if !amqpConnection.IsClosed() {
			failOnError(amqpConnection.Close(), "RabbitMQ | Close")
		}
	}()
	log.Println("Connected to RabbitMQ")

	channelAmqp, err := amqpConnection.Channel()
	failOnError(err, "RabbitMQ | Channel")
	defer func() {
		if !channelAmqp.IsClosed() {
			failOnError(channelAmqp.Close(), "RabbitMQ | Chanel close error")
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

	err = channelAmqp.ExchangeDeclare(os.Getenv("RABBITMQ_EXCHANGE_REINDEX"), amqp.ExchangeDirect, true, false, false, false, nil)
	failOnError(err, "RabbitMQ | Failed to declare a exchange")
	placeQueue, err := channelAmqp.QueueDeclare(os.Getenv("RABBITMQ_QUEUE_PLACE_REINDEX"), true, false, false, false, nil)
	failOnError(err, "RabbitMQ | Failed to declare a queue")
	err = channelAmqp.QueueBind(placeQueue.Name, os.Getenv("RABBITMQ_ROUTING_PLACE_KEY"), os.Getenv("RABBITMQ_EXCHANGE_REINDEX"), false, nil)
	failOnError(err, "RabbitMQ | Failed to bind a queue")

	// SITE
	siteSchema := os.Getenv("SITE_SCHEMA")
	siteHost := os.Getenv("SITE_HOST")
	sitePort := os.Getenv("SITE_PORT")

	// session
	sessionSecret := os.Getenv("SESSION_SECRET")
	sessionName := os.Getenv("SESSION_NAME")
	sessionPath := os.Getenv("SESSION_PATH")
	sessionDomain := os.Getenv("SESSION_DOMAIN")
	sessionMaxAge, err := strconv.Atoi(os.Getenv("SESSION_MAX_AGE"))
	failOnError(err, "SESSION | SESSION_MAX_AGE")

	// session store // TODO Remake
	colectionSessions := client.Database(mongoDB).Collection("sessions")
	sessionStore := mongodriver.NewStore(colectionSessions, sessionMaxAge, false, []byte(sessionSecret))
	sessionStore.Options(sessions.Options{
		Path:     sessionPath,
		Domain:   sessionDomain,
		MaxAge:   sessionMaxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// auth // TODO Remake
	collectionUsers := client.Database(mongoDB).Collection("users")
	userMongoRepository := repository.NewUserMongoRepository(ctx, collectionUsers)
	authService := service.NewDefaultAuthService(userMongoRepository)
	tokenPresenter := presenter.NewTokenPresenter()
	authHandler = handlers.NewAuthHandler(ctx, authService, tokenPresenter)

	// category
	collectionCategories := client.Database(mongoDB).Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(ctx, collectionCategories)
	categoryService := service.NewDefaultCategoryService(categoryMongoRepository)
	categoryPresenter := presenter.NewCategoryPresenter()
	categoriesHandler = handlers.NewCategoriesHandler(ctx, categoryService, categoryPresenter)

	// place
	collectionPlaces := client.Database(mongoDB).Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(ctx, collectionPlaces)
	placeQueueRabbitRepository := repository.NewPlaceQueueRabbitRepository(channelAmqp)
	placeService := service.NewDefaultPlaceService(placeMongoRepository, categoryMongoRepository, placeQueueRabbitRepository)
	placePresenter := presenter.NewPlacePresenter()
	placesHandler = handlers.NewPlacesHandler(ctx, placeService, placePresenter)

	// log
	fileLog, _ := os.Create("debug.log")
	gin.DefaultWriter = io.MultiWriter(fileLog)

	router := gin.Default()
	// router.SetTrustedProxies([]string{"192.168.1.2"})
	// router.UseH2C = true

	// common midelleware
	router.Use(middleware.Cors(siteSchema, siteHost, sitePort))
	router.Use(middleware.RequestAbsUrl())

	// routes for version 1
	apiV1 := router.Group("/v1")
	apiV1auth := apiV1.Group("")

	// session midelleware
	sessionMidlleware := middleware.Session(sessionName, sessionStore)
	apiV1.Use(sessionMidlleware)
	apiV1auth.Use(sessionMidlleware)
	// auth midelleware
	apiV1auth.Use(middleware.Auth())

	authHandler.MakeHandlers(apiV1)
	placesHandler.MakeHandlers(apiV1, apiV1auth)
	placesHandler.MakeRequestValidation()
	categoriesHandler.MakeHandlers(apiV1, apiV1auth)

	router.GET("/version", VersionHandler)

	// build server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
		// TODO
	}
	// srv.RegisterOnShutdown()

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			failOnError(err, "listen")
		}
	}()

	// Listen for the interrupt signal.
	<-ctxSignal.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctxSignal, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	failOnError(srv.Shutdown(ctxSignal), "Server forced to shutdown")

	log.Println("Server exiting")
}

func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": os.Getenv("API_VERSION")})
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
