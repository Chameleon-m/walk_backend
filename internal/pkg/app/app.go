package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"walk_backend/internal/app/api/handlers"
	"walk_backend/internal/app/api/middleware"
	"walk_backend/internal/app/api/presenter"
	"walk_backend/internal/app/repository"
	"walk_backend/internal/app/service"
	"walk_backend/internal/pkg/cache"
	"walk_backend/internal/pkg/components"
	"walk_backend/internal/pkg/env"
	"walk_backend/internal/pkg/httpserver"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type App struct {
	// handlers []*gin.HandlerFunc
	handlers  handlers.HandlersInterface
	engine    *gin.Engine
	env       env.EnvInterface
	ctx       context.Context
	ctxCancel context.CancelFunc
}

var _ httpserver.ServerInterface = (*App)(nil)

func New() *App {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true})
	log.Print("Started")

	app := &App{}
	app.env = env.New()

	gin.DisableConsoleColor()
	gin.SetMode(app.env.GetMust(gin.EnvGinMode)) // GIN_MODE

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if app.env.GetMust(gin.EnvGinMode) == gin.DebugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return app
}

func (app *App) Run() {

	log.Print("Run started")
	defer log.Print("Server exiting")

	var err error

	// Create context that listens for the interrupt signal from the OS.
	ctxSignal, ctxSignalStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer ctxSignalStop()

	app.ctx, app.ctxCancel = context.WithCancel(context.Background())
	defer app.ctxCancel()

	app.handlers = handlers.New(app)

	g, gCtx := errgroup.WithContext(app.ctx)

	// TODO create logger interface and inject logger

	// Redis
	redisComponent := components.NewRedis(
		app.env.GetMust("REDIS_HOST"),
		app.env.GetMust("REDIS_PORT"),
		app.env.GetMust("REDIS_USERNAME"),
		app.env.GetMust("REDIS_PASSWORD"),
	)
	g.Go(func() error { return redisComponent.Start(gCtx) })
	defer func() {
		if err := redisComponent.Stop(context.TODO()); err != nil {
			log.Error().Err(err).Caller(0).Send()
		}
	}()

	// RabbitMQ
	rabbitMqPublisherComponent := components.NewRabbitMQ(app.env.GetMust("RABBITMQ_URI"))
	g.Go(func() error { return rabbitMqPublisherComponent.Start(gCtx) })
	defer func() {
		if err := rabbitMqPublisherComponent.Stop(context.TODO()); err != nil {
			log.Error().Err(err).Caller(0).Send()
		}
	}()

	// DB mongo
	mongoDefaultDB := app.env.GetMust("MONGO_INITDB_DATABASE")
	mongoDBComponent := components.NewMongoDB(app.env.GetMust("MONGO_URI"))
	g.Go(func() error { return mongoDBComponent.Start(gCtx) })
	defer func() {
		if err := mongoDBComponent.Stop(context.TODO()); err != nil {
			log.Error().Err(err).Caller(0).Send()
		}
	}()

	<-mongoDBComponent.Ready()
	// Session // MongoDB store
	sessionComponent := components.NewSessionGinMongoDB(
		app.env.GetMust("SESSION_SECRET"),
		app.env.GetMust("SESSION_PATH"),
		app.env.GetMust("SESSION_DOMAIN"),
		app.env.GetMustInt("SESSION_MAX_AGE"),
		app.env.GetMust("SESSION_MONGO_DATABASE"),
		mongoDBComponent.GetClient(),
	)
	g.Go(func() error { return sessionComponent.Start(gCtx) })

	if err := g.Wait(); err != nil {
		log.Fatal().Err(err).Caller(0).Send()
	}

	mongoClient := mongoDBComponent.GetClient()
	rabbitMQClient := rabbitMqPublisherComponent.GetClient()
	redisClient := redisComponent.GetClient()

	// auth
	collectionUsers := mongoClient.Database(mongoDefaultDB).Collection("users")
	userMongoRepository := repository.NewUserMongoRepository(app.ctx, collectionUsers)
	authService := service.NewDefaultAuthService(userMongoRepository)
	tokenPresenter := presenter.NewTokenPresenter()
	app.handlers.SetAuthHandler(handlers.NewAuthHandler(app.ctx, authService, tokenPresenter))

	// category
	collectionCategories := mongoClient.Database(mongoDefaultDB).Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(app.ctx, collectionCategories)
	categoryService := service.NewDefaultCategoryService(categoryMongoRepository)
	categoryPresenter := presenter.NewCategoryPresenter()
	app.handlers.SetCategoriesHandler(handlers.NewCategoriesHandler(app.ctx, categoryService, categoryPresenter))

	// place
	collectionPlaces := mongoClient.Database(mongoDefaultDB).Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(app.ctx, collectionPlaces)
	placeCacheRedisRepository := repository.NewPlaceCacheRedisRepository(app.ctx, redisClient)
	placeQueueRabbitRepository := repository.NewPlaceQueueRabbitRepository(
		rabbitMQClient,
		rabbitMQClient.NotifyReturn(),
		rabbitMQClient.NotifyPublish(),
	)
	keyBuilder := cache.NewKeyBuilderDefault()
	placeService := service.NewDefaultPlaceService(
		placeMongoRepository,
		categoryMongoRepository,
		placeQueueRabbitRepository,
		placeCacheRedisRepository,
		keyBuilder,
	)
	placePresenter := presenter.NewPlacePresenter()
	app.handlers.SetPlacesHandler(handlers.NewPlacesHandler(app.ctx, placeService, placePresenter))

	// Engine
	app.engine = gin.New()
	app.engine.Use(gin.Recovery())

	// common midelleware

	// zerolog midelleware
	fileLog, err := os.Create("debug.log")
	if err != nil {
		log.Fatal().Err(err).Caller(0).Send()
	}

	defaultLevel, err := zerolog.ParseLevel(app.env.GetMust("LOG_DEFAULT_LEVEL"))
	if err != nil {
		log.Fatal().Err(err).Caller(0).Send()
	}
	clientLevel, err := zerolog.ParseLevel(app.env.GetMust("LOG_CLIENT_LEVEL"))
	if err != nil {
		log.Fatal().Err(err).Caller(0).Send()
	}
	serverLevel, err := zerolog.ParseLevel(app.env.GetMust("LOG_SERVER_LEVEL"))
	if err != nil {
		log.Fatal().Err(err).Caller(0).Send()
	}

	app.engine.Use(logger.SetLogger(
		logger.WithSkipPath([]string{"/version", "/prometheus"}),
		logger.WithUTC(true),
		logger.WithWriter(io.MultiWriter(fileLog)),
		logger.WithDefaultLevel(defaultLevel),
		logger.WithClientErrorLevel(clientLevel),
		logger.WithServerErrorLevel(serverLevel),
		logger.WithLogger(func(c *gin.Context, l zerolog.Logger) zerolog.Logger {
			return l.With().
				Str("id", c.GetHeader("X-Request-ID")).
				Logger()
		}),
	))

	// Prometheus middleware
	app.engine.Use(middleware.Prometheus())

	// CORS middleware
	siteSchema := os.Getenv("SITE_SCHEMA")
	siteHost := os.Getenv("SITE_HOST")
	sitePort := os.Getenv("SITE_PORT")
	app.engine.Use(middleware.Cors(siteSchema, siteHost, sitePort))

	// RequestAbsURL middleware
	app.engine.Use(middleware.RequestAbsURL())

	// session midelleware
	sessionName := app.env.GetMust("SESSION_NAME")
	sessionMidlleware := middleware.Session(sessionName, sessionComponent.GetClient())

	// auth midelleware
	authMiddleware := middleware.Auth()

	// routes for version 1
	apiV1 := app.engine.Group("/api/v1")
	apiV1auth := apiV1.Group("")
	apiV1.Use(sessionMidlleware)
	apiV1auth.Use(sessionMidlleware)
	apiV1auth.Use(authMiddleware)

	// app.handlers.make() // make interface
	app.handlers.GetAuthHandler().MakeHandlers(apiV1)
	app.handlers.GetPlacesHandler().MakeHandlers(apiV1, apiV1auth)
	app.handlers.GetPlacesHandler().MakeRequestValidation()
	app.handlers.GetCategoriesHandler().MakeHandlers(apiV1, apiV1auth)

	app.engine.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": os.Getenv("API_VERSION")})
	})
	app.engine.GET("/prometheus", gin.WrapH(promhttp.Handler()))

	done := make(chan struct{}, 1)
	go func() {
		select {
		case <-ctxSignal.Done():
			log.Print("os signal done")
		case <-app.ctx.Done():
			log.Print("app ctx done")
		}

		done <- struct{}{}
	}()

	// build server
	addr := fmt.Sprintf(":%s", app.env.GetMust("API_PORT"))
	server := &http.Server{
		Addr:    addr,
		Handler: app.engine,
		// TODO
	}
	// server.RegisterOnShutdown(func() {})

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if err != nil {
				log.Panic().Err(err).Send()
			}
		}
	}()

	// Awaiting done chan
	<-done

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	log.Print("Shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Panic().Err(err).Msg("server forced to shutdown")
	}
}

// GetEnvironment return debug release test
func (app *App) GetEnvironment() string {
	return gin.Mode()
}

// IsDebug return bool
func (app *App) IsDebug() bool {
	return gin.IsDebugging()
}

func (app *App) GetContext() context.Context {
	return app.ctx
}
