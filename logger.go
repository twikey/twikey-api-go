package twikey

import "log"

type Logger interface {
	// Debugf is the level for basic functions and arguments
	Debugf(format string, v ...interface{})
	// Tracef is the level for http calls
	Tracef(format string, v ...interface{})
}

// NullLogger is the (default) logger that ignores all logging
type NullLogger struct{}

func (n NullLogger) Debugf(format string, v ...interface{}) {}
func (n NullLogger) Tracef(format string, v ...interface{}) {}

// DebugLogger allows the logging of both basic payloads as well as responses
type DebugLogger struct {
	logger *log.Logger
}

func (t DebugLogger) Debugf(format string, v ...interface{}) {
	t.logger.Printf(format, v...)
}

func (t DebugLogger) Tracef(format string, v ...interface{}) {
	t.logger.Printf(format, v...)
}

func NewDebugLogger(logger *log.Logger) Logger {
	return DebugLogger{logger: logger}
}
