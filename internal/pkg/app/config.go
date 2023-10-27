package app

import (
	"flag"
	"fmt"
	mongoComponent "walk_backend/internal/pkg/components/mongo"
	rabitmqComponent "walk_backend/internal/pkg/components/rabbitmq"
	redisComponent "walk_backend/internal/pkg/components/redis"
	sessionComponent "walk_backend/internal/pkg/components/session"
	"walk_backend/internal/pkg/util"
)

type Config struct {
	Version string `yaml:"version"  env:"APP_VERSION" env-required:"true"   env-description:"Application version"`
	GinMode string `yaml:"gin_mode" env:"GIN_MODE"    env-default:"release" env-description:"Set gin mode"`
	Log     struct {
		Level int  `yaml:"log_level" env:"LOG_LEVEL" env-default:"1"    env-description:"Log level"`
		UTC   bool `yaml:"utc"       env:"LOG_UTC"   env-default:"true" env-description:"Use UTC for log timestamp"`
	} `yaml:"log"`
	RequestLog struct {
		Enable   bool                 `yaml:"enable"    env:"REQUEST_LOG_ENABLE"    env-default:"true" env-description:"Request log on/off"`
		SkipPath util.StringSliceFlag `yaml:"skip_path" env:"REQUEST_LOG_SKIP_PATH" env-default:""     env-description:"Request log skip path" env-separator:","`
	} `yaml:"request_log"`
	Api struct {
		Schema string `yaml:"schema" env:"API_SCHEMA" env-default:"http" env-description:"API schema"`
		Host   string `yaml:"host"   env:"API_HOST"   env-default:""     env-description:"API host"`
		Port   string `yaml:"port"   env:"API_PORT"   env-default:"8080" env-description:"API port"`
	} `yaml:"api"`
	Site struct {
		Schema string `yaml:"schema" env:"SITE_SCHEMA" env-default:"https"     env-description:"SITE schema"`
		Host   string `yaml:"host"   env:"SITE_HOST"   env-default:"localhost" env-description:"SITE host"`
		Port   string `yaml:"port"   env:"SITE_PORT"   env-default:"443"       env-description:"SITE port"`
	} `yaml:"site"`
	Queue struct {
		ReIndex struct {
			Exchange string `yaml:"exchange" env:"RABBITMQ_EXCHANGE_REINDEX" env-default:"reindex_exchange" env-description:"Exchange for reindex"`
			Place    struct {
				RoutingKey        string `yaml:"routing_key"   env:"RABBITMQ_ROUTING_PLACE_KEY"   env-default:"place_routing_key"   env-description:"Routing key for place"`
				QueuePlaceReindex string `yaml:"queue"         env:"RABBITMQ_QUEUE_PLACE_REINDEX" env-default:"place_reindex_queue" env-description:"Queue name for place reindex"`
			} `yaml:"place"`
		} `yaml:"reindex"`
	} `yaml:"queue"`
	Redis    redisComponent.Config             `yaml:"redis_component"`
	RabbitMQ rabitmqComponent.Config           `yaml:"rabbit_mq_component"`
	MongoDB  mongoComponent.Config             `yaml:"mongo_db_component"`
	Session  sessionComponent.GinMongoDBConfig `yaml:"session_component"`
}

func (cfg *Config) RegisterFlags(fs *flag.FlagSet) {

	fs.StringVar(&cfg.GinMode, "gin-mode", cfg.GinMode, "Gin mode")
	fs.IntVar(&cfg.Log.Level, "log-level", cfg.Log.Level, "Log level")
	fs.BoolVar(&cfg.Log.UTC, "log-utc", cfg.Log.UTC, "Use UTC for log timestamp")
	fs.BoolVar(&cfg.RequestLog.Enable, "request-log", cfg.RequestLog.Enable, "Request log on/off")
	fs.Var(&cfg.RequestLog.SkipPath, "request-log-skip", "Request log skip path, use , for list")
	fs.StringVar(&cfg.Api.Schema, "api-schema", cfg.Api.Schema, "API schema")
	fs.StringVar(&cfg.Api.Host, "api-host", cfg.Api.Host, "API host")
	fs.StringVar(&cfg.Api.Port, "api-port", cfg.Api.Port, "API port")
	fs.StringVar(&cfg.Site.Schema, "site-schema", cfg.Site.Schema, "Site schema")
	fs.StringVar(&cfg.Site.Host, "site-host", cfg.Site.Host, "Site host")
	fs.StringVar(&cfg.Site.Port, "site-port", cfg.Site.Port, "Site port")
	fs.StringVar(&cfg.Queue.ReIndex.Exchange, "queue-reindex-exchange", cfg.Queue.ReIndex.Exchange, "Queue exchange for reindex")
	fs.StringVar(&cfg.Queue.ReIndex.Place.RoutingKey, "queue-routing-place-key", cfg.Queue.ReIndex.Exchange, "Queue routing key for place")
	fs.StringVar(&cfg.Queue.ReIndex.Place.QueuePlaceReindex, "queue-name-place-reindex", cfg.Queue.ReIndex.Exchange, "Queue name for place reindex")

	cfg.Redis.RegisterFlags(fs)
	cfg.RabbitMQ.RegisterFlags(fs)
	cfg.MongoDB.RegisterFlags(fs)
	cfg.Session.RegisterFlags(fs)
}

func (cfg *Config) Validate() error {
	if cfg.GinMode != "debug" && cfg.GinMode != "release" && cfg.GinMode != "test" {
		return fmt.Errorf("invalid gin mode")
	}
	// TODO
	if err := cfg.Redis.Validate(); err != nil {
		return fmt.Errorf("config redis_component error: %w", err)
	}
	if err := cfg.RabbitMQ.Validate(); err != nil {
		return fmt.Errorf("config rabbit_mq_component error: %w", err)
	}
	if err := cfg.MongoDB.Validate(); err != nil {
		return fmt.Errorf("config mongo_db_component error: %w", err)
	}
	if err := cfg.Session.Validate(); err != nil {
		return fmt.Errorf("config session_component error: %w", err)
	}
	return nil
}
