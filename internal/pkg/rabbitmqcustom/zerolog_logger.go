package rabbitmqcustom

import (
	"github.com/rs/zerolog"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

var _ rabbitmq.Logger = (*zerologLogger)(nil)

// zerologLogger logs to stdout up to the `DebugF` level
type zerologLogger struct {
	log    zerolog.Logger
	logErr zerolog.Logger
}

func NewZerologLogger(log, logErr zerolog.Logger) zerologLogger {
	return zerologLogger{log: log, logErr: logErr}
}

func (l zerologLogger) Fatalf(format string, v ...interface{}) {
	l.logErr.Fatal().Msgf(format, v...)
}

func (l zerologLogger) Errorf(format string, v ...interface{}) {
	l.logErr.Error().Msgf(format, v...)
}

func (l zerologLogger) Warnf(format string, v ...interface{}) {
	l.logErr.Warn().Msgf(format, v...)
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
