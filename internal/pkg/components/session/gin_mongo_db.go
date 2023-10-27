package session

import (
	"context"
	"net/http"

	"walk_backend/internal/pkg/components"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ components.ComponentStartInterface = (*ginMongoDBComponent)(nil)
var _ components.ComponentReadyInterface = (*ginMongoDBComponent)(nil)

type ginMongoDBComponent struct {
	mongoClient  *mongo.Client
	config       GinMongoDBConfig
	sessionStore sessions.Store
	ready        chan struct{}
	log          zerolog.Logger
}

func NewGinMongoDB(name string, log zerolog.Logger, config GinMongoDBConfig, mongoClient *mongo.Client) *ginMongoDBComponent {
	return &ginMongoDBComponent{
		mongoClient: mongoClient,
		config:      config,
		ready:       make(chan struct{}, 1),
		log:         log.With().Str("component", name).Logger(),
	}
}

func (c *ginMongoDBComponent) Start(ctx context.Context) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	c.log.Print("Start connected to Session store")

	// session store
	colectionSessions := c.mongoClient.Database(c.config.DBName).Collection("sessions")
	c.sessionStore = mongodriver.NewStore(colectionSessions, c.config.MaxAge, false, []byte(c.config.Secret))
	c.sessionStore.Options(sessions.Options{
		Path:     c.config.Path,
		Domain:   c.config.Domain,
		MaxAge:   c.config.MaxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	c.log.Print("Connected to Session store")

	c.ready <- struct{}{}
	close(c.ready)
	c.log.Print("Session READY")

	return nil
}

func (c *ginMongoDBComponent) GetClient() sessions.Store {
	return c.sessionStore
}

func (c *ginMongoDBComponent) Ready() <-chan struct{} {
	return c.ready
}
