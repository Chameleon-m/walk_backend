package rabbitmq_custom

import (
	"fmt"
)

type Logging interface {
	Printf(format string, v ...any)
}

func NewNullLogger() logger {
	return logger{
		logger:      nullLogging{},
		loggerError: nullLogging{},
	}
}

func NewLogger(log, errLog Logging) logger {
	return logger{
		logger:      log,
		loggerError: errLog,
	}
}

// logger logs to stdout up to the `DebugF` level
type logger struct {
	logger      Logging
	loggerError Logging
}

// Enables logging using a custom Logging instance. Note that this is
// not thread safe and should be called at application start
func (l *logger) SetLogger(log Logging) {
	l.logger = log
}

func (l *logger) SetLoggerError(log Logging) {
	l.loggerError = log
}

func (l logger) Fatalf(format string, v ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("FATAL: %s", format), v...)
}

func (l logger) Errorf(format string, v ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("ERROR: %s", format), v...)
}

func (l logger) Warnf(format string, v ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("WARN: %s", format), v...)
}

func (l logger) Infof(format string, v ...interface{}) {
	l.logger.Printf(fmt.Sprintf("INFO: %s", format), v...)
}

func (l logger) Debugf(format string, v ...interface{}) {
	l.logger.Printf(fmt.Sprintf("DEBUG: %s", format), v...)
}

func (l logger) Tracef(format string, v ...interface{}) {}

// nullLogging
type nullLogging struct {
}

func (l nullLogging) Printf(format string, v ...any) {
}
