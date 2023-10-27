package rabbitmq

import (
	"flag"
	"fmt"
)

type Config struct {
	URL string `yaml:"url" env:"RABBITMQ_URL" env-default:"amqp://guest:guest@localhost:5672/" env-description:"RabbitMQ connection URL"`
}

func (cfg *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.URL, "rabbitmq-url", cfg.URL, "RabbitMQ connection URL")
}

func (cfg *Config) Validate() error {
	// TODO
	if cfg.URL == "" {
		return fmt.Errorf("invalid URL")
	}
	return nil
}
