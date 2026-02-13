package utils

import (
	"log"
	"os"
	"sync"
)

type Logger struct {
	*log.Logger
}

var (
	instance *Logger
	once     sync.Once
)

// GetLogger returns singleton logger instance
func GetLogger() *Logger {
	once.Do(func() {
		logLevel := os.Getenv("LOG_LEVEL")
		if logLevel == "" {
			logLevel = "info"
		}

		instance = &Logger{
			Logger: log.New(os.Stdout, "[Kloset] ", log.LstdFlags|log.Lshortfile),
		}
	})
	return instance
}

// Info logs info level message
func (l *Logger) Info(format string, args ...interface{}) {
	l.Printf("[INFO] "+format, args...)
}

// Error logs error level message
func (l *Logger) Error(format string, args ...interface{}) {
	l.Printf("[ERROR] "+format, args...)
}

// Debug logs debug level message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.Printf("[DEBUG] "+format, args...)
}

// Fatal logs fatal level message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.Fatalf("[FATAL] "+format, args...)
}
