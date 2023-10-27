package mongo

import (
	"flag"
	"fmt"
)

type Config struct {
	InitDBName string `yaml:"init_db_name" env:"MONGO_INITDB_NAME" env-description:"Mongo init database"`
	URL        string `yaml:"url"          env:"MONGO_URL"         env-description:"Mongo connection URL"`
}

func (cfg *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.InitDBName, "mongo-init-db", cfg.InitDBName, "Mongo init database")
	fs.StringVar(&cfg.URL, "mongo-url", cfg.URL, "Mongo connection URL")
}

func (cfg *Config) Validate() error {
	// TODO
	if cfg.InitDBName == "" {
		return fmt.Errorf("invalid init DB name")
	} else if cfg.URL == "" {
		return fmt.Errorf("invalid URL")
	}
	return nil
}
