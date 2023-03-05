package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type ctxLogger struct{}

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, logger)
}

// LoggerFromContext returns logger from context
func LoggerFromContext(ctx context.Context) *zerolog.Logger {
	if logger, ok := ctx.Value(ctxLogger{}).(*zerolog.Logger); ok {
		return logger
	}
	panic("Context does not contain a logger")
}
