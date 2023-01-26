package rabbitmqcustom

import (
	"github.com/rs/zerolog"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

var _ rabbitmq.Logger = (*zerologLogger)(nil)

// zerologLogger logs to stdout up to the `DebugF` level
type zerologLogger struct {
	log zerolog.Logger
}

func NewZerologLogger(log zerolog.Logger) zerologLogger {
	return zerologLogger{log: log}
}

func (l zerologLogger) Fatalf(format string, v ...interface{}) {
	l.log.Fatal().Msgf(format, v...)
}

func (l zerologLogger) Errorf(format string, v ...interface{}) {
	l.log.Error().Msgf(format, v...)
}

func (l zerologLogger) Warnf(format string, v ...interface{}) {
	l.log.Warn().Msgf(format, v...)
}

func (l zerologLogger) Infof(format string, v ...interface{}) {
	l.log.Info().Msgf(format, v...)
}

func (l zerologLogger) Debugf(format string, v ...interface{}) {
	l.log.Debug().Msgf(format, v...)
}

func (l zerologLogger) Tracef(format string, v ...interface{}) {
	l.log.Trace().Msgf(format, v...)
}
