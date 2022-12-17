package rabbitmqcustom

import (
	"fmt"
)

// Logging test
type Logging interface {
	Printf(format string, v ...any)
}

// NewNullLogger create new null logger
func NewNullLogger() Logger {
	return Logger{
		logger:      nullLogging{},
		loggerError: nullLogging{},
	}
}

// NewLogger create new logger
func NewLogger(log, errLog Logging) Logger {
	return Logger{
		logger:      log,
		loggerError: errLog,
	}
}

// Logger logs to stdout up to the `DebugF` level
type Logger struct {
	logger      Logging
	loggerError Logging
}

// SetLogger Enables logging using a custom Logging instance. Note that this is
// not thread safe and should be called at application start
func (l *Logger) SetLogger(log Logging) {
	l.logger = log
}

// SetLoggerError ...
func (l *Logger) SetLoggerError(log Logging) {
	l.loggerError = log
}

// Fatalf ...
func (l Logger) Fatalf(format string, v ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("FATAL: %s", format), v...)
}

// Errorf ...
func (l Logger) Errorf(format string, v ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("ERROR: %s", format), v...)
}

// Warnf ...
func (l Logger) Warnf(format string, v ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("WARN: %s", format), v...)
}

// Infof ...
func (l Logger) Infof(format string, v ...interface{}) {
	l.logger.Printf(fmt.Sprintf("INFO: %s", format), v...)
}

// Debugf ...
func (l Logger) Debugf(format string, v ...interface{}) {
	l.logger.Printf(fmt.Sprintf("DEBUG: %s", format), v...)
}

// Tracef ...
func (l Logger) Tracef(format string, v ...interface{}) {}

// nullLogging
type nullLogging struct {
}

// Printf ...
func (l nullLogging) Printf(format string, v ...any) {
}
