package awssecretmanagerrotationlambda

import "log"

type LeveledLogger interface {
	Error(format string, a ...any)
	Warn(format string, a ...any)
	Info(format string, a ...any)
	Debug(format string, a ...any)
	Trace(format string, a ...any)
}

// LeveledLoggerStandard is a simple and stupid implementation. It will silent Debug + Trace levels
type LeveledLoggerStandard struct {
	*log.Logger
	logLevel LogLevel
}

type LogLevel uint8

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

func NewLeveledLoggerStandard(logLevel LogLevel) LeveledLoggerStandard {
	return LeveledLoggerStandard{
		Logger: log.Default(),
	}
}

func (l LeveledLoggerStandard) Error(format string, a ...any) {
	if LogLevelError <= l.logLevel {
		return
	}
	l.Logger.Printf("Error: "+format, a)
}
func (l LeveledLoggerStandard) Warn(format string, a ...any) {
	if LogLevelWarn <= l.logLevel {
		return
	}
	l.Logger.Printf("Warn: "+format, a)
}
func (l LeveledLoggerStandard) Info(format string, a ...any) {
	if LogLevelInfo <= l.logLevel {
		return
	}
	l.Logger.Printf("Info: "+format, a)
}
func (l LeveledLoggerStandard) Debug(format string, a ...any) {
	if LogLevelDebug <= l.logLevel {
		return
	}
	l.Logger.Printf("Debug: "+format, a)
}
func (l LeveledLoggerStandard) Trace(format string, a ...any) {
	if LogLevelTrace <= l.logLevel {
		return
	}
	l.Logger.Printf("Trace: "+format, a)
}
