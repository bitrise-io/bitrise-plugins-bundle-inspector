// Package logger provides structured logging capabilities
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Level represents the severity level of a log message
type Level int

const (
	// LevelDebug for debug messages
	LevelDebug Level = iota
	// LevelInfo for informational messages
	LevelInfo
	// LevelWarn for warning messages
	LevelWarn
	// LevelError for error messages
	LevelError
)

// Logger defines the interface for structured logging
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// DefaultLogger is a standard implementation of Logger
type DefaultLogger struct {
	output io.Writer
	level  Level
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(output io.Writer, level Level) *DefaultLogger {
	return &DefaultLogger{
		output: output,
		level:  level,
	}
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.log("DEBUG", format, args...)
	}
}

// Info logs an informational message
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.log("INFO", format, args...)
	}
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.log("WARN", format, args...)
	}
}

// Error logs an error message
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.log("ERROR", format, args...)
	}
}

func (l *DefaultLogger) log(level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.New(l.output, fmt.Sprintf("[%s] ", level), log.LstdFlags).Print(msg)
}

// SilentLogger is a logger that discards all output (useful for testing)
type SilentLogger struct{}

// NewSilentLogger creates a new silent logger
func NewSilentLogger() *SilentLogger {
	return &SilentLogger{}
}

// Debug does nothing
func (l *SilentLogger) Debug(format string, args ...interface{}) {}

// Info does nothing
func (l *SilentLogger) Info(format string, args ...interface{}) {}

// Warn does nothing
func (l *SilentLogger) Warn(format string, args ...interface{}) {}

// Error does nothing
func (l *SilentLogger) Error(format string, args ...interface{}) {}

// Global logger instance
var global Logger = NewDefaultLogger(os.Stderr, LevelInfo)

// SetGlobal sets the global logger
func SetGlobal(logger Logger) {
	global = logger
}

// Debug logs a debug message using the global logger
func Debug(format string, args ...interface{}) {
	global.Debug(format, args...)
}

// Info logs an informational message using the global logger
func Info(format string, args ...interface{}) {
	global.Info(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(format string, args ...interface{}) {
	global.Warn(format, args...)
}

// Error logs an error message using the global logger
func Error(format string, args ...interface{}) {
	global.Error(format, args...)
}
