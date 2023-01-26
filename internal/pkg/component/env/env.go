package env

import (
	"os"
	"strconv"
)

var _ EnvInterface = (*Env)(nil)

// Env ...
type Env struct{}

func New() *Env {
	return &Env{}
}

// GetEnv ...
func (e *Env) Get(name string) string {
	return os.Getenv(name)
}

// GetMust panic if env not isset or empty string
func (e *Env) GetMust(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic(NewError(ErrEnvNotIsset, name, value))
	}
	return value
}

// GetMustInt convert string env to int
func (e *Env) GetMustInt(name string) int {
	value := os.Getenv(name)
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		panic(NewError(ErrEnvConvert, name, value).AddParentErr(err))
	}
	return valueInt
}
