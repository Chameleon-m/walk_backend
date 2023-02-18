package components

import (
	"context"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ ComponentStartInterface = (*sessionGinMongoDBComponent)(nil)
var _ ComponentReadyInterface = (*sessionGinMongoDBComponent)(nil)

type sessionGinMongoDBComponent struct {
	mongoClient    *mongo.Client
	sessionMongoDB string
	sessionSecret  string
	sessionPath    string
	sessionDomain  string
	sessionMaxAge  int
	sessionStore   sessions.Store
	ready          chan struct{}
}

func NewSessionGinMongoDB(
	sessionSecret string,
	sessionPath string,
	sessionDomain string,
	sessionMaxAge int,
	sessionMongoDB string,
	mongoClient *mongo.Client,
) *sessionGinMongoDBComponent {
	return &sessionGinMongoDBComponent{
		mongoClient:    mongoClient,
		sessionMongoDB: sessionMongoDB,
		sessionSecret:  sessionSecret,
		sessionPath:    sessionPath,
		sessionDomain:  sessionDomain,
		sessionMaxAge:  sessionMaxAge,
		ready:          make(chan struct{}, 1),
	}
}

func (c *sessionGinMongoDBComponent) Start(ctx context.Context) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	log.Print("Start connected to Session store")

	// session store
	colectionSessions := c.mongoClient.Database(c.sessionMongoDB).Collection("sessions")
	c.sessionStore = mongodriver.NewStore(colectionSessions, c.sessionMaxAge, false, []byte(c.sessionSecret))
	c.sessionStore.Options(sessions.Options{
		Path:     c.sessionPath,
		Domain:   c.sessionDomain,
		MaxAge:   c.sessionMaxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	log.Print("Connected to Session store")

	c.ready <- struct{}{}
	close(c.ready)
	log.Print("Session READY")

	return nil
}

func (c *sessionGinMongoDBComponent) GetClient() sessions.Store {
	return c.sessionStore
}

func (c *sessionGinMongoDBComponent) Ready() <-chan struct{} {
	return c.ready
}
