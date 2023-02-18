package httpserver

import "context"

type ServerInterface interface {
	Run()
	GetEnvironment() string
	IsDebug() bool
	GetContext() context.Context
}
