package httpserver

import "context"

type ServerInterface interface {
	Run() error
	GetEnvironment() string
	IsDebug() bool
	GetContext() context.Context
}
