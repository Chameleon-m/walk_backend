package rabbitmqcustom

import (
	"fmt"
	"log"

	rabbitmq "github.com/wagslane/go-rabbitmq"
)

var _ rabbitmq.Logger = (*logLogging)(nil)

// logLogging ...
type logLogging struct {
	logger      *log.Logger
	loggerError *log.Logger
}

// NewLogLogger create new logger
func NewLogLogger(log, errLog *log.Logger) logLogging {
	return logLogging{
		logger:      log,
		loggerError: errLog,
	}
}

// Fatalf ...
func (l logLogging) Fatalf(format string, v ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("FATAL: %s", format), v...)
}

// Errorf ...
func (l logLogging) Errorf(format string, v ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("ERROR: %s", format), v...)
}

// Warnf ...
func (l logLogging) Warnf(format string, v ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("WARN: %s", format), v...)
}

// Infof ...
func (l logLogging) Infof(format string, v ...interface{}) {
	l.logger.Printf(fmt.Sprintf("INFO: %s", format), v...)
}

// Debugf ...
func (l logLogging) Debugf(format string, v ...interface{}) {
	l.logger.Printf(fmt.Sprintf("DEBUG: %s", format), v...)
}

// Tracef ...
func (l logLogging) Tracef(format string, v ...interface{}) {
	l.logger.Printf(fmt.Sprintf("TRACE: %s", format), v...)
}
