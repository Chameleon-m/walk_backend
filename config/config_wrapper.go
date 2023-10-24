package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"walk_backend/internal/pkg/app"

	"github.com/ilyakaznacheev/cleanenv"
)

type ConfigWrapper struct {
	App app.Config `yaml:"app"` // `yaml:",inline"`

	envDescription bool
	verifyConfig   bool
	printConfig    bool
	configFile     string
	envFile        string
}

// New returns app config.
func New() (*ConfigWrapper, error) {

	cfg := &ConfigWrapper{}

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = cleanenv.FUsage(fs.Output(), cfg, nil, fs.Usage)
	fs.StringVar(&cfg.configFile, "config-file", "config/config.yaml", "configuration file to load")
	fs.StringVar(&cfg.envFile, "env-file", ".env", "env file to load")
	fs.BoolVar(&cfg.envDescription, "env-desc", false, "Descriptions of all environment variables")
	cfg.RegisterFlags(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil, err
		}
		return nil, fmt.Errorf("flag error: %w", err)
	}

	if err := cleanenv.ReadConfig(cfg.configFile, cfg); err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}

	if cfg.envFile != "" {
		if file, err := os.Stat(cfg.envFile); err == nil {
			if file.IsDir() {
				return nil, fmt.Errorf("env file error: the file is a directory")
			} else if err := cleanenv.ReadConfig(cfg.envFile, cfg); err != nil {
				return nil, fmt.Errorf("env file error: %w", err)
			}
		} else if errors.Is(err, os.ErrNotExist) {
			// Read Env
		} else {
			return nil, fmt.Errorf("env file error: %w", err)
		}
	} else if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("env error: %w", err)
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("flag error: %w", err)
	}

	return cfg, nil
}

// RegisterFlags ...
func (cfg *ConfigWrapper) RegisterFlags(fs *flag.FlagSet) {

	fs.BoolVar(&cfg.verifyConfig, "verify-config", false, "Verify config file and exits")
	fs.BoolVar(&cfg.printConfig, "print-config-stderr", false, "Dump the entire config object to stderr")

	cfg.App.RegisterFlags(fs)
}

// Validate ...
func (cfg *ConfigWrapper) Validate() error {

	if err := cfg.App.Validate(); err != nil {
		return err
	}
	return nil
}

// GetEnvDescription ...
func (cfg *ConfigWrapper) GetEnvDescription() (string, error) {
	return cleanenv.GetDescription(cfg, nil)
}

// UpdateEnv ...
func (cfg *ConfigWrapper) UpdateEnv() error {
	return cleanenv.UpdateEnv(cfg)
}

// IsEnvDescription ...
func (cfg *ConfigWrapper) IsEnvDescription() bool {
	return cfg.envDescription
}

// IsVerifyConfig ...
func (cfg *ConfigWrapper) IsVerifyConfig() bool {
	return cfg.verifyConfig
}

// IsPrintConfig ...
func (cfg *ConfigWrapper) IsPrintConfig() bool {
	return cfg.printConfig
}
