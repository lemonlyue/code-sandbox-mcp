package sandbox

import "context"

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	Ctx(ctx context.Context) Logger
}

type NoOpLogger struct{}

func (n NoOpLogger) Ctx(ctx context.Context) Logger {
	return n
}

func (n NoOpLogger) Debugf(format string, args ...interface{}) {}

func (n NoOpLogger) Infof(format string, args ...interface{}) {}

func (n NoOpLogger) Warnf(format string, args ...interface{}) {}

func (n NoOpLogger) Errorf(format string, args ...interface{}) {}

func (n NoOpLogger) WithField(key string, value interface{}) Logger {
	return n
}

var internalLogger Logger = &NoOpLogger{}

func SetLogger(logger Logger) {
	if logger == nil {
		internalLogger = &NoOpLogger{}
		return
	}
	internalLogger = logger
}
