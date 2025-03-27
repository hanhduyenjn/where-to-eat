package util

import (
	"fmt"
	"log"
	"os"
)

// Logger provides structured logging
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// NewLogger creates a new logger
func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
	}
}

// Info logs information messages
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.log(l.infoLogger, msg, keyvals...)
}

// Error logs error messages
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.log(l.errorLogger, msg, keyvals...)
}

// Debug logs debug messages
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.log(l.debugLogger, msg, keyvals...)
}

// log formats and writes the log message
func (l *Logger) log(logger *log.Logger, msg string, keyvals ...interface{}) {
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "MISSING")
	}

	details := ""
	for i := 0; i < len(keyvals); i += 2 {
		details += fmt.Sprintf(" %v=%v", keyvals[i], keyvals[i+1])
	}

	logger.Println(msg + details)
}
