// Places API
//
// This is a places API.
//
//		Schemes: http
//	 Host: localhost:8080
//		BasePath: /v1
//		Version: 0.0.1
//		Contact: Dmitry Korolev <korolev.d.l@yandex.ru> https://github.com/Chameleon-m
//
//		Consumes:
//		- application/json
//
//		Produces:
//		- application/json
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
	"syscall"
	"time"

	"walk_backend/cmd/api/handlers"
	"walk_backend/cmd/api/middleware"
	"walk_backend/cmd/api/presenter"
	"walk_backend/repository"
	"walk_backend/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var placesHandler *handlers.PlacesHandler
var categoriesHandler *handlers.CategoriesHandler

func init() {

}

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	// category
	collectionCategories := client.Database(mongoDB).Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(ctx, collectionCategories)
	categoryService := service.NewDefaultCategoryService(categoryMongoRepository)
	categoryPresenter := presenter.NewCategoryPresenter()
	categoriesHandler = handlers.NewCategoriesHandler(ctx, categoryService, categoryPresenter)

	// place
	collectionPlaces := client.Database(mongoDB).Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(ctx, collectionPlaces)
	placeService := service.NewDefaultPlaceService(placeMongoRepository, categoryMongoRepository)
	placePresenter := presenter.NewPlacePresenter()
	placesHandler = handlers.NewPlacesHandler(ctx, placeService, placePresenter)

	// log
	fileLog, _ := os.Create("debug.log")
	gin.DefaultWriter = io.MultiWriter(fileLog)

	router := gin.Default()
	// router.SetTrustedProxies([]string{"192.168.1.2"})
	// router.UseH2C = true

	// midelleware
	router.Use(middleware.RequestAbsUrl())

	// routes for version 1
	apiV1 := router.Group("/v1")

	placesHandler.MakeHandlers(apiV1)
	placesHandler.MakeRequestValidation()
	categoriesHandler.MakeHandlers(apiV1)

	router.GET("/version", VersionHandler)

	// build server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	// srv.RegisterOnShutdown()

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
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
	if err := srv.Shutdown(ctxSignal); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": os.Getenv("API_VERSION")})
}
