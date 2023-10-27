package session

import (
	"flag"
	"fmt"
)

type GinMongoDBConfig struct {
	Name   string `yaml:"name"    env:"SESSION_NAME"    env-description:"Session name"`
	Secret string `yaml:"secret"  env:"SESSION_SECRET"  env-description:"Session secret"`
	Path   string `yaml:"path"    env:"SESSION_PATH"    env-description:"Session path"`
	Domain string `yaml:"domain"  env:"SESSION_DOMAIN"  env-description:"Session domain"`
	MaxAge int    `yaml:"max_age" env:"SESSION_MAX_AGE" env-description:"Session max age"`
	DBName string `yaml:"db_name" env:"SESSION_DB_NAME" env-description:"Session table/collection ... name"`
}

func (cfg *GinMongoDBConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.Name, "session-name", cfg.Name, "Session name")
	fs.StringVar(&cfg.Secret, "session-secret", cfg.Secret, "Session secret")
	fs.StringVar(&cfg.Path, "session-path", cfg.Path, "Session path")
	fs.StringVar(&cfg.Domain, "session-domain", cfg.Domain, "Session domain")
	fs.IntVar(&cfg.MaxAge, "session-max-age", cfg.MaxAge, "Session max age")
	fs.StringVar(&cfg.DBName, "session-db-name", cfg.DBName, "Session table/collection ... name")
}

func (cfg *GinMongoDBConfig) Validate() error {
	// TODO
	if cfg.Name == "" {
		return fmt.Errorf("invalid name")
	}
	return nil
}
