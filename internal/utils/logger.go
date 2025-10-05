package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/nareshkumar-microsoft/notificationService/pkg/interfaces"
)

// SimpleLogger is a basic implementation of the Logger interface
type SimpleLogger struct {
	logger *log.Logger
	level  LogLevel
	fields map[string]interface{}
}

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// NewSimpleLogger creates a new simple logger
func NewSimpleLogger(level string) *SimpleLogger {
	return &SimpleLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		level:  parseLogLevel(level),
		fields: make(map[string]interface{}),
	}
}

func (l *SimpleLogger) Debug(args ...interface{}) {
	if l.level <= LevelDebug {
		l.logger.Printf("[DEBUG] %s", fmt.Sprint(args...))
	}
}

func (l *SimpleLogger) Info(args ...interface{}) {
	if l.level <= LevelInfo {
		l.logger.Printf("[INFO] %s", fmt.Sprint(args...))
	}
}

func (l *SimpleLogger) Warn(args ...interface{}) {
	if l.level <= LevelWarn {
		l.logger.Printf("[WARN] %s", fmt.Sprint(args...))
	}
}

func (l *SimpleLogger) Error(args ...interface{}) {
	if l.level <= LevelError {
		l.logger.Printf("[ERROR] %s", fmt.Sprint(args...))
	}
}

func (l *SimpleLogger) Debugf(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

func (l *SimpleLogger) Warnf(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.logger.Printf("[WARN] "+format, args...)
	}
}

func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

func (l *SimpleLogger) WithField(key string, value interface{}) interfaces.Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &SimpleLogger{
		logger: l.logger,
		level:  l.level,
		fields: newFields,
	}
}

func (l *SimpleLogger) WithFields(fields map[string]interface{}) interfaces.Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &SimpleLogger{
		logger: l.logger,
		level:  l.level,
		fields: newFields,
	}
}

func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}
