package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"walk_backend/internal/app/api/handlers/auth"
	"walk_backend/internal/app/api/handlers/category"
	"walk_backend/internal/app/api/handlers/place"
	"walk_backend/internal/app/api/middleware"
	"walk_backend/internal/app/api/presenter"
	"walk_backend/internal/app/repository"
	"walk_backend/internal/app/service"
	"walk_backend/internal/pkg/cache"
	"walk_backend/internal/pkg/components"
	"walk_backend/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// HandlersInterface ...
type HandlersInterface interface {
	// Make routes, validation, etc
	Make()
}

type App struct {
	engine    *gin.Engine
	cfg       *Config
	ctx       context.Context
	ctxCancel context.CancelFunc
	logger    zerolog.Logger
}

func New(cfg *Config) *App {

	app := &App{}
	app.cfg = cfg

	gin.DisableConsoleColor()
	gin.SetMode(app.cfg.GinMode)

	zerolog.TimeFieldFormat = time.RFC3339Nano
	if app.cfg.Log.UTC {
		zerolog.TimestampFunc = func() time.Time {
			return time.Now().UTC()
		}
	}
	app.logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	app.logger.Printf("Started mod: %s", gin.Mode())
	logLevel := zerolog.Level(app.cfg.Log.Level)
	app.logger.Level(logLevel)
	zerolog.SetGlobalLevel(logLevel)
	if app.cfg.GinMode == gin.DebugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	app.logger.Printf("Log level: %s", logLevel.String())

	return app
}

func (app *App) Run() {

	log := app.logger
	log.Print("Run started")
	defer log.Print("Server exiting")

	// Create context that listens for the interrupt signal from the OS.
	ctxSignal, ctxSignalStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer ctxSignalStop()

	// Creater app context with logger
	app.ctx, app.ctxCancel = context.WithCancel(logger.ContextWithLogger(context.Background(), &log))
	defer app.ctxCancel()

	g, gCtx := errgroup.WithContext(app.ctx)

	// TODO create logger interface and inject logger

	// Redis
	redisComponent := components.NewRedis(
		"redis",
		app.logger,
		app.cfg.Redis.Host,
		app.cfg.Redis.Port,
		app.cfg.Redis.Username,
		app.cfg.Redis.Password,
	)
	g.Go(func() error { return redisComponent.Start(gCtx) })
	defer func() {
		if err := redisComponent.Stop(context.TODO()); err != nil {
			log.Error().Err(err).Caller(0).Send()
		}
	}()

	// RabbitMQ
	rabbitMqPublisherComponent := components.NewRabbitMQ("rabbitMQ", app.logger, app.cfg.RabbitMQ.URI)
	g.Go(func() error { return rabbitMqPublisherComponent.Start(gCtx) })
	defer func() {
		if err := rabbitMqPublisherComponent.Stop(context.TODO()); err != nil {
			log.Error().Err(err).Caller(0).Send()
		}
	}()

	// DB mongo
	mongoDefaultDB := app.cfg.MongoDB.InitDBName
	mongoDBComponent := components.NewMongoDB("mongoDB", app.logger, app.cfg.MongoDB.URI)
	g.Go(func() error { return mongoDBComponent.Start(gCtx) })
	defer func() {
		if err := mongoDBComponent.Stop(context.TODO()); err != nil {
			log.Error().Err(err).Caller(0).Send()
		}
	}()

	<-mongoDBComponent.Ready()
	// Session // MongoDB store
	sessionName := app.cfg.Session.Name
	sessionComponent := components.NewSessionGinMongoDB(
		"session",
		app.logger,
		app.cfg.Session.Secret,
		app.cfg.Session.Path,
		app.cfg.Session.Domain,
		app.cfg.Session.MaxAge,
		app.cfg.Session.DBName,
		mongoDBComponent.GetClient(),
	)
	g.Go(func() error { return sessionComponent.Start(gCtx) })

	if err := g.Wait(); err != nil {
		log.Fatal().Err(err).Caller(0).Send()
	}

	mongoClient := mongoDBComponent.GetClient()
	rabbitMQClient := rabbitMqPublisherComponent.GetClient()
	redisClient := redisComponent.GetClient()

	// Engine
	app.engine = gin.New()
	app.engine.Use(gin.Recovery())

	// common middleware

	// request logger middleware
	if app.cfg.RequestLog.Enable {
		log.Printf("Request logger skip path: %v", app.cfg.RequestLog.SkipPath)
		app.engine.Use(middleware.RequestLogger(&app.logger, app.cfg.RequestLog.SkipPath))
	}

	// Prometheus middleware
	app.engine.Use(middleware.Prometheus())

	// CORS middleware
	app.engine.Use(middleware.Cors(app.cfg.Site.Schema, app.cfg.Site.Host, app.cfg.Site.Port))

	// RequestAbsURL middleware
	app.engine.Use(middleware.RequestAbsURL())

	// session middleware
	sessionMidlleware := middleware.Session(sessionName, sessionComponent.GetClient())

	// auth middleware
	authMiddleware := middleware.Auth()

	// routes for version 1
	apiV1 := app.engine.Group("/api/v1")
	apiV1.Use(sessionMidlleware)

	apiV1auth := apiV1.Group("")
	apiV1auth.Use(authMiddleware)

	// Build handlers
	var authHandlers, categoryHandlers, placeHandlers HandlersInterface

	// auth
	collectionUsers := mongoClient.Database(mongoDefaultDB).Collection("users")
	userMongoRepository := repository.NewUserMongoRepository(collectionUsers)
	authService := service.NewDefaultAuthService(userMongoRepository)
	tokenPresenter := presenter.NewTokenPresenter()
	authHandlers = auth.NewHandler(app.ctx, apiV1, authService, tokenPresenter)
	authHandlers.Make()

	// category
	collectionCategories := mongoClient.Database(mongoDefaultDB).Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(collectionCategories)
	categoryService := service.NewDefaultCategoryService(categoryMongoRepository)
	categoryPresenter := presenter.NewCategoryPresenter()
	categoryHandlers = category.NewHandler(app.ctx, apiV1, apiV1auth, categoryService, categoryPresenter)
	categoryHandlers.Make()

	// place
	collectionPlaces := mongoClient.Database(mongoDefaultDB).Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(collectionPlaces)
	placeCacheRedisRepository := repository.NewPlaceCacheRedisRepository(redisClient)
	placeQueueRabbitRepository := repository.NewPlaceQueueRabbitRepository(
		app.ctx,
		rabbitMQClient,
		app.cfg.Queue.ReIndex.Exchange,
		app.cfg.Queue.ReIndex.Place.RoutingKey,
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
	placeHandlers = place.NewHandler(app.ctx, apiV1, apiV1auth, placeService, placePresenter)
	placeHandlers.Make()

	app.engine.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": app.cfg.Version})
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
	server := &http.Server{
		Addr:    ":" + app.cfg.Api.Port, //net.JoinHostPort(app.cfg.Api.Host, app.cfg.Api.Port),
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

// GetContext return app context
func (app *App) GetContext() context.Context {
	return app.ctx
}
