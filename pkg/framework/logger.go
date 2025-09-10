package framework

import "log"

// Logger defines logging interface
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// stdLogger basic implementation using log package
type stdLogger struct{}

func (l *stdLogger) Debugf(f string, a ...interface{}) { log.Printf("DEBUG "+f, a...) }
func (l *stdLogger) Infof(f string, a ...interface{})  { log.Printf("INFO  "+f, a...) }
func (l *stdLogger) Warnf(f string, a ...interface{})  { log.Printf("WARN  "+f, a...) }
func (l *stdLogger) Errorf(f string, a ...interface{}) { log.Printf("ERROR "+f, a...) }

func newLogger(enabled bool) Logger {
	if !enabled { return &noopLogger{} }
	return &stdLogger{}
}

type noopLogger struct{}
func (n *noopLogger) Debugf(string, ...interface{}) {}
func (n *noopLogger) Infof(string, ...interface{})  {}
func (n *noopLogger) Warnf(string, ...interface{})  {}
func (n *noopLogger) Errorf(string, ...interface{}) {}
