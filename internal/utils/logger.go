package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Log levels
const (
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	DEBUG = "DEBUG"
)

var (
	instance *Logger
	once     sync.Once
)

// Logger struct
type Logger struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// getDefaultLogFilePath returns the default log file path
func getDefaultLogFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	logDir := filepath.Join(homeDir, ".geomys")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}
	return filepath.Join(logDir, "geomys.log")
}

// NewLogger creates a new logger instance (singleton)
func NewLogger(logFilePath string, debugMode bool) *Logger {
	once.Do(func() {
		if logFilePath == "" {
			logFilePath = getDefaultLogFilePath()
		}

		// Open the log file
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}

		// Create a multi-writer to log to both the file and console
		multiWriter := io.MultiWriter(file, os.Stdout)

		// Create loggers
		infoLogger := log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime)
		warnLogger := log.New(multiWriter, "[WARN] ", log.Ldate|log.Ltime)
		errorLogger := log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime)

		var debugWriter io.Writer
		if debugMode {
			debugWriter = multiWriter
		} else {
			debugWriter = file
		}
		debugLogger := log.New(debugWriter, "[DEBUG] ", log.Ldate|log.Ltime)

		instance = &Logger{
			infoLogger:  infoLogger,
			warnLogger:  warnLogger,
			errorLogger: errorLogger,
			debugLogger: debugLogger,
		}
	})
	return instance
}

// GetLogger retrieves the singleton logger instance
func GetLogger() *Logger {
	if instance == nil {
		log.Fatalf("Logger has not been initialized. Call NewLogger() first.")
	}
	return instance
}

// Logging methods
func (l *Logger) Info(message string) {
	l.infoLogger.Println(message)
}

func (l *Logger) Warn(message string) {
	l.warnLogger.Println(message)
}

func (l *Logger) Error(message string) {
	l.errorLogger.Println(message)
}

func (l *Logger) Debug(message string) {
	l.debugLogger.Println(message)
}
