package sandbox

import (
	"context"
	"fmt"
	"log"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	Ctx(ctx context.Context) Logger
}

// ANSI color
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

type NoOpLogger struct{}

func (n NoOpLogger) Ctx(ctx context.Context) Logger {
	return n
}

func (n NoOpLogger) Debugf(format string, args ...interface{}) {
	log.Printf(fmt.Sprintf("%s[DEBUG] %s%s", colorBlue, fmt.Sprintf(format, args...), colorReset))
}

func (n NoOpLogger) Infof(format string, args ...interface{}) {
	log.Printf(fmt.Sprintf("%s[INFO] %s%s", colorGreen, fmt.Sprintf(format, args...), colorReset))
}

func (n NoOpLogger) Warnf(format string, args ...interface{}) {
	log.Printf(fmt.Sprintf("%s[WARN] %s%s", colorYellow, fmt.Sprintf(format, args...), colorReset))
}

func (n NoOpLogger) Errorf(format string, args ...interface{}) {
	log.Printf(fmt.Sprintf("%s[ERROR] %s%s", colorRed, fmt.Sprintf(format, args...), colorReset))
}

func (n NoOpLogger) WithField(key string, value interface{}) Logger {
	return n
}

var InternalLogger Logger = &NoOpLogger{}

func SetLogger(logger Logger) {
	if logger == nil {
		InternalLogger = &NoOpLogger{}
		return
	}
	InternalLogger = logger
}
