package rabbitmqcustom

import (
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

var _ rabbitmq.Logger = (*nullLogging)(nil)

// nullLogging
type nullLogging struct {
}

func NewNullLogger() nullLogging {
	return nullLogging{}
}

func (l *nullLogging) Fatalf(format string, v ...interface{}) {}

func (l *nullLogging) Errorf(format string, v ...interface{}) {}

func (l *nullLogging) Warnf(format string, v ...interface{}) {}

func (l *nullLogging) Infof(format string, v ...interface{}) {}

func (l *nullLogging) Debugf(format string, v ...interface{}) {}

func (l *nullLogging) Tracef(format string, v ...interface{}) {}
