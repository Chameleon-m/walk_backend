package redis

import (
	"flag"
	"fmt"
)

type Config struct {
	URL string `yaml:"url"  env:"REDIS_URL"  env-default:"redis://username:password@localhost:6379/0" env-description:"Redis connection URL"`
}

func (cfg *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.URL, "redis-url", cfg.URL, "Redis connection URL")
}

func (cfg *Config) Validate() error {
	// TODO
	if cfg.URL == "" {
		return fmt.Errorf("invalid URL")
	}
	return nil
}
