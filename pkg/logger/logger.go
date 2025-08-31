package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Initialize sets up the logger with the specified level
func Initialize(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)
}

// GetLogger returns a logger instance with context fields
func GetLogger() *logrus.Logger {
	return logrus.StandardLogger()
}

// WithField returns a logger with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return logrus.WithField(key, value)
}

// WithFields returns a logger with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return logrus.WithFields(fields)
}