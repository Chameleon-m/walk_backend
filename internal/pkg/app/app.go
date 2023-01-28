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
	"walk_backend/internal/pkg/component/cache"
	"walk_backend/internal/pkg/component/env"
	httpserver "walk_backend/internal/pkg/component/http_server"
	rabbitmqLog "walk_backend/internal/pkg/rabbitmqcustom"

	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	rabbitmq "github.com/wagslane/go-rabbitmq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/sync/errgroup"
)

type App struct {
	// handlers []*gin.HandlerFunc
	handlers handlers.HandlersInterface
	engine   *gin.Engine
	env      env.EnvInterface

	ctxSignal     context.Context
	ctxSignalStop context.CancelFunc
	ctx           context.Context
	ctxCancel     context.CancelFunc

	redisClient    *redis.Client
	redisCtx       context.Context
	redisCtxCancel context.CancelFunc

	mongoClient    *mongo.Client
	mongoCtx       context.Context
	mongoCtxCancel context.CancelFunc
	mongoDefaultDb string

	rabbitmqPublisher *rabbitmq.Publisher

	sessionStore sessions.Store
}

var _ httpserver.ServerInterface = (*App)(nil)

func New() (*App, error) {

	var err error

	app := &App{}
	app.env = env.New()

	gin.DisableConsoleColor()
	gin.SetMode(app.env.GetMust(gin.EnvGinMode)) // GIN_MODE

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if app.env.GetMust(gin.EnvGinMode) == gin.DebugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Print("Started")

	// Create context that listens for the interrupt signal from the OS.
	app.ctxSignal, app.ctxSignalStop = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	app.ctx, app.ctxCancel = context.WithCancel(app.ctxSignal)

	app.handlers = handlers.New(app)

	// TODO gCtx // or components New, Start, defer Stop on Run
	g, _ := errgroup.WithContext(app.ctx)

	// Redis
	g.Go(func() error { return app.initRedis() })

	g.Go(func() error {
		// DB mongo
		if err := app.initMongoDB(); err != nil {
			return err
		}
		// Session // MongoDB store
		if err = app.initSession(); err != nil {
			return err
		}

		return nil
	})

	// RabbitMQ
	g.Go(func() error { return app.initRabbitMQ() })

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// auth
	collectionUsers := app.GetMongoDB().Collection("users")
	userMongoRepository := repository.NewUserMongoRepository(app.ctx, collectionUsers)
	authService := service.NewDefaultAuthService(userMongoRepository)
	tokenPresenter := presenter.NewTokenPresenter()
	app.handlers.SetAuthHandler(handlers.NewAuthHandler(app.ctx, authService, tokenPresenter))

	// category
	collectionCategories := app.GetMongoDB().Collection("categories")
	categoryMongoRepository := repository.NewCategoryMongoRepository(app.ctx, collectionCategories)
	categoryService := service.NewDefaultCategoryService(categoryMongoRepository)
	categoryPresenter := presenter.NewCategoryPresenter()
	app.handlers.SetCategoriesHandler(handlers.NewCategoriesHandler(app.ctx, categoryService, categoryPresenter))

	// place
	collectionPlaces := app.GetMongoDB().Collection("places")
	placeMongoRepository := repository.NewPlaceMongoRepository(app.ctx, collectionPlaces)
	placeCacheRedisRepository := repository.NewPlaceCacheRedisRepository(app.ctx, app.redisClient)
	placeQueueRabbitRepository := repository.NewPlaceQueueRabbitRepository(
		app.rabbitmqPublisher,
		app.rabbitmqPublisher.NotifyReturn(),
		app.rabbitmqPublisher.NotifyPublish(),
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
		return nil, err
	}

	defaultLevel, err := zerolog.ParseLevel(app.env.GetMust("LOG_DEFAULT_LEVEL"))
	if err != nil {
		return nil, err
	}
	clientLevel, err := zerolog.ParseLevel(app.env.GetMust("LOG_CLIENT_LEVEL"))
	if err != nil {
		return nil, err
	}
	serverLevel, err := zerolog.ParseLevel(app.env.GetMust("LOG_SERVER_LEVEL"))
	if err != nil {
		return nil, err
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
	sessionMidlleware := middleware.Session(sessionName, app.sessionStore)

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

	return app, nil
}

func (app *App) Run() error {

	log.Print("Run started")

	defer app.ctxSignalStop()
	defer app.ctxCancel()
	defer app.redisCtxCancel()
	defer app.mongoCtxCancel()
	defer app.initRedisDefer()
	defer app.initMongoDBDefer()
	defer app.initRabbitMQDefer()

	done := make(chan bool, 1)
	go func() {
		select {
		// Listen for the interrupt signal.
		case <-app.ctxSignal.Done():
			log.Print("os signal done")
		case <-app.ctx.Done():
			log.Print("ctx done")
		case <-app.mongoCtx.Done():
			log.Print("mongo done")
		case <-app.redisCtx.Done():
			log.Print("redis done")
		}

		done <- true
	}()

	// build server
	addr := fmt.Sprintf(":%s", app.env.GetMust("API_PORT"))
	server := &http.Server{
		Addr:    addr,
		Handler: app.engine,
		// TODO
	}
	// srv.RegisterOnShutdown()

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
	app.ctxSignalStop()
	log.Print("Shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Panic().Err(err).Msg("server forced to shutdown")
	}

	log.Print("Server exiting")

	return nil
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

func (app *App) initRedis() error {

	log.Print("Start connected to Redis")

	app.redisCtx, app.redisCtxCancel = context.WithCancel(app.ctx)

	addrRedis := fmt.Sprintf("%s:%s", app.env.GetMust("REDIS_HOST"), app.env.GetMust("REDIS_PORT"))
	app.redisClient = redis.NewClient(&redis.Options{
		Addr:     addrRedis,
		Username: app.env.GetMust("REDIS_USERNAME"),
		Password: app.env.GetMust("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})

	log.Print("Connected to Redis")

	log.Print("Redis status: PING")
	status, err := app.redisClient.Ping(app.redisCtx).Result()
	if err != nil {
		return err
	}
	log.Print("Redis status: " + status)

	return nil
}

func (app *App) initRedisDefer() func() {
	return func() {
		if app.redisClient != nil {
			if statusCmd, err := app.redisClient.ShutdownSave(app.redisCtx).Result(); err != nil {
				log.Error().Err(err).Msg(statusCmd)
			}
		}
	}
}

func (app *App) initMongoDB() error {

	log.Print("Start connected to MongoDB")

	var err error
	mongoURI := app.env.GetMust("MONGO_URI")
	app.mongoDefaultDb = app.env.GetMust("MONGO_INITDB_DATABASE")
	app.mongoCtx, app.mongoCtxCancel = context.WithCancel(app.ctx)

	mongoClientOptions := options.Client()
	mongoClientOptions.ApplyURI(mongoURI)
	app.mongoClient, err = mongo.Connect(app.mongoCtx, mongoClientOptions)
	if err != nil {
		return err
	}

	log.Print("Connected to MongoDB")

	log.Print("MongoDB PING")
	if err := app.initMongoDBPing(); err != nil {
		return err
	}
	log.Print("MongoDB PONG")

	return nil
}

func (app *App) initMongoDBDefer() func() {
	return func() {
		if app.mongoClient != nil {
			if err := app.mongoClient.Disconnect(app.mongoCtx); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}
}

func (app *App) initMongoDBPing() error {
	return app.mongoClient.Ping(app.mongoCtx, readpref.Primary())
}

func (app *App) GetMongoDB() *mongo.Database {
	return app.mongoClient.Database(app.mongoDefaultDb)
}

func (app *App) initRabbitMQ() error {

	log.Print("Start connected to RabbitMQ")

	var err error
	rabbitMQURL := app.env.GetMust("RABBITMQ_URI")

	app.rabbitmqPublisher, err = rabbitmq.NewPublisher(
		rabbitMQURL,
		rabbitmq.Config{
			Dial: amqp091.DefaultDial(30 * time.Second),
		},
		rabbitmq.WithPublisherOptionsLogger(rabbitmqLog.NewZerologLogger(log.Logger, log.Logger)),
	)
	if err != nil {
		return err
	}

	log.Print("Connected to RabbitMQ")

	return nil
}

func (app *App) initRabbitMQDefer() func() {
	return func() {
		if app.rabbitmqPublisher != nil {
			if err := app.rabbitmqPublisher.Close(); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}
}

func (app *App) initSession() error {

	log.Print("Start connected to Session store")

	sessionSecret := app.env.GetMust("SESSION_SECRET")
	sessionPath := app.env.GetMust("SESSION_PATH")
	sessionDomain := app.env.GetMust("SESSION_DOMAIN")
	sessionMaxAge := app.env.GetMustInt("SESSION_MAX_AGE")

	// session store
	colectionSessions := app.GetMongoDB().Collection("sessions")
	app.sessionStore = mongodriver.NewStore(colectionSessions, sessionMaxAge, false, []byte(sessionSecret))
	app.sessionStore.Options(sessions.Options{
		Path:     sessionPath,
		Domain:   sessionDomain,
		MaxAge:   sessionMaxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	log.Print("Connected to Session store")

	return nil
}
