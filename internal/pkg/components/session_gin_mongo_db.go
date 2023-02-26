package components

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

type SessionGinMongoDBConfig struct {
	Name   string `yaml:"name"    env:"SESSION_NAME"    env-description:"Session name"`
	Secret string `yaml:"secret"  env:"SESSION_SECRET"  env-description:"Session secret"`
	Path   string `yaml:"path"    env:"SESSION_PATH"    env-description:"Session path"`
	Domain string `yaml:"domain"  env:"SESSION_DOMAIN"  env-description:"Session domain"`
	MaxAge int    `yaml:"max_age" env:"SESSION_MAX_AGE" env-description:"Session max age"`
	DBName string `yaml:"db_name" env:"SESSION_DB_NAME" env-description:"Session table/collection ... name"`
}

func (cfg *SessionGinMongoDBConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.Name, "session-name", cfg.Name, "Session name")
	fs.StringVar(&cfg.Secret, "session-secret", cfg.Secret, "Session secret")
	fs.StringVar(&cfg.Path, "session-path", cfg.Path, "Session path")
	fs.StringVar(&cfg.Domain, "session-domain", cfg.Domain, "Session domain")
	fs.IntVar(&cfg.MaxAge, "session-max-age", cfg.MaxAge, "Session max age")
	fs.StringVar(&cfg.DBName, "session-db-name", cfg.DBName, "Session table/collection ... name")
}

func (cfg *SessionGinMongoDBConfig) Validate() error {
	// TODO
	if cfg.Name == "" {
		return fmt.Errorf("invalid name")
	}
	return nil
}

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
	log            zerolog.Logger
}

func NewSessionGinMongoDB(
	name string,
	log zerolog.Logger,
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
		log:            log.With().Str("component", name).Logger(),
	}
}

func (c *sessionGinMongoDBComponent) Start(ctx context.Context) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.log.Print("Start connected to Session store")

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

	c.log.Print("Connected to Session store")

	c.ready <- struct{}{}
	close(c.ready)
	c.log.Print("Session READY")

	return nil
}

func (c *sessionGinMongoDBComponent) GetClient() sessions.Store {
	return c.sessionStore
}

func (c *sessionGinMongoDBComponent) Ready() <-chan struct{} {
	return c.ready
}
