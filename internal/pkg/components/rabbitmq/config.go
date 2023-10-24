package rabbitmq

import (
	"flag"
	"fmt"
)

type Config struct {
	URI string `yaml:"uri" env:"RABBITMQ_URI" env-default:"amqp://guest:guest@localhost:5672/" env-description:"RabbitMQ connection URI"`
}

func (cfg *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.URI, "rabbitmq-uri", cfg.URI, "RabbitMQ connection URI")
}

func (cfg *Config) Validate() error {
	// TODO
	if cfg.URI == "" {
		return fmt.Errorf("invalid URI")
	}
	return nil
}
