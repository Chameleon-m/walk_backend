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

	"walk_backend/internal/app/api/handlers"
	"walk_backend/internal/app/api/middleware"
	"walk_backend/internal/app/api/presenter"
	"walk_backend/internal/app/repository"
	"walk_backend/internal/app/service"
	rabbitmqLog "walk_backend/internal/pkg/rabbitmqcustom"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	rabbitmq "github.com/wagslane/go-rabbitmq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var authHandler *handlers.AuthHandler
var placesHandler *handlers.PlacesHandler
var categoriesHandler *handlers.CategoriesHandler

func init() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("[MAIN] ")
}

func main() {

	// Create context that listens for the interrupt signal from the OS.
	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var err error

	// DB
	ctxMongo, cancel := context.WithCancel(ctxSignal)
	defer cancel()
	mongoClientOptions := options.Client()
	mongoClientOptions.ApplyURI(os.Getenv("MONGO_URI"))
	mongoClient, err := mongo.Connect(ctxMongo, mongoClientOptions)
	failOnError(err, "MongoDB")
	defer func() {
		if err := mongoClient.Disconnect(ctxMongo); err != nil {
			log.Printf("MongoDB: %s", err)
		}
	}()

	failOnError(mongoClient.Ping(ctxMongo, readpref.Primary()), "MongoDB | Ping")
	log.Println("Connected to MongoDB")

	mongoDB := os.Getenv("MONGO_INITDB_DATABASE")

	// RabbitMQ

	rabbitmqLoggger := log.New(log.Writer(), "[RABBITMQ] ", log.LstdFlags|log.Lmsgprefix)
	publisher, err := rabbitmq.NewPublisher(
		os.Getenv("RABBITMQ_URI"),
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogger(rabbitmqLog.NewLogger(rabbitmqLoggger, rabbitmqLoggger)),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	log.Println("Connected to RabbitMQ")

	notifyReturn := publisher.NotifyReturn()
	notifyPublish := publisher.NotifyPublish()

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
	colectionSessions := mongoClient.Database(mongoDB).Collection("sessions")
	sessionStore := mongodriver.NewStore(colectionSessions, sessionMaxAge, false, []byte(sessionSecret))
	sessionStore.Options(sessions.Options{
		Path:     sessionPath,
		Domain:   sessionDomain,
		MaxAge:   sessionMaxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	ctx, cancel := context.WithCancel(ctxSignal)
	defer cancel()

	// auth
	collectionUsers := mongoClient.Database(mongoDB).Collection("users")
	userMongoRepository := repository.NewUserMongoRepository(ctx, collectionUsers)
	authService := service.NewDefaultAuthService(userMongoRepository)
	tokenPresenter := presenter.NewTokenPresenter()
	authHandler = handlers.NewAuthHandler(ctx, authService, tokenPresenter)

	// category
	collectionCategories := mongoClient.Database(mongoDB).Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(ctx, collectionCategories)
	categoryService := service.NewDefaultCategoryService(categoryMongoRepository)
	categoryPresenter := presenter.NewCategoryPresenter()
	categoriesHandler = handlers.NewCategoriesHandler(ctx, categoryService, categoryPresenter)

	// place
	collectionPlaces := mongoClient.Database(mongoDB).Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(ctx, collectionPlaces)
	placeQueueRabbitRepository := repository.NewPlaceQueueRabbitRepository(publisher, notifyReturn, notifyPublish)
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
	router.Use(middleware.Prometheus())
	router.Use(middleware.Cors(siteSchema, siteHost, sitePort))
	router.Use(middleware.RequestAbsURL())

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
	router.GET("/prometheus", gin.WrapH(promhttp.Handler()))

	// build server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
		// TODO
	}
	// srv.RegisterOnShutdown()

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

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			failOnError(err, "listen")
		}
	}()

	// Awaiting done chan
	<-done

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	failOnError(srv.Shutdown(ctxShutdown), "Server forced to shutdown")

	log.Println("Server exiting")
}

// VersionHandler api version
func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": os.Getenv("API_VERSION")})
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
