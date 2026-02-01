package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides structured logging with levels
type Logger struct {
	debug  bool
	logger *log.Logger
}

// New creates a new Logger instance
func New(debug bool) *Logger {
	return &Logger{
		debug:  debug,
		logger: log.New(os.Stderr, "", 0),
	}
}

// formatMessage formats a log message with timestamp and level
func (l *Logger) formatMessage(level, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s - %s - %s", timestamp, level, message)
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.logger.Println(l.formatMessage("INFO", message))
}

// Debug logs a debug message (only if debug mode is enabled)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.debug {
		message := fmt.Sprintf(format, args...)
		l.logger.Println(l.formatMessage("DEBUG", message))
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.logger.Println(l.formatMessage("WARNING", message))
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.logger.Println(l.formatMessage("ERROR", message))
}

// SetDebug enables or disables debug logging
func (l *Logger) SetDebug(debug bool) {
	l.debug = debug
}

// IsDebug returns whether debug logging is enabled
func (l *Logger) IsDebug() bool {
	return l.debug
}

// Made with help from Bob
