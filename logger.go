package querycraft

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogLevel represents the level of logging
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogFormat represents the format of log messages
type LogFormat int

const (
	LogFormatText LogFormat = iota
	LogFormatJSON
)

// LoggerOptions represents the options for the logger
type LoggerOptions struct {
	Enabled        bool
	Level          LogLevel
	Format         LogFormat
	SaveToFile     bool
	PrintToConsole bool
	LogDir         string
	AutoCleanDays  int
}

// DefaultLoggerOptions returns the default logger options
func DefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Enabled:        false,
		Level:          LogLevelInfo,
		Format:         LogFormatText,
		SaveToFile:     true,
		PrintToConsole: false,
		LogDir:         "./storage/logs/sql/",
		AutoCleanDays:  7,
	}
}

// FileLogger is a logger that writes to files
type FileLogger struct {
	options LoggerOptions
}

// NewFileLogger creates a new file logger
func NewFileLogger(options LoggerOptions) *FileLogger {
	logger := &FileLogger{
		options: options,
	}

	// Create log directory if it doesn't exist
	if options.Enabled {
		err := os.MkdirAll(options.LogDir, 0755)
		if err != nil {
			fmt.Printf("Error creating log directory: %v\n", err)
		}

		// Clean old log files
		if options.AutoCleanDays > 0 {
			logger.cleanOldLogs()
		}
	}

	return logger
}

// LogQuery logs a query
func (l *FileLogger) LogQuery(ctx context.Context, query string, args []any, duration time.Duration, err error) {
	if !l.options.Enabled {
		return
	}

	// Format the query with arguments
	formattedQuery := query
	for _, arg := range args {
		formattedQuery = strings.Replace(formattedQuery, "?", fmt.Sprintf("'%v'", arg), 1)
	}

	// Create log entry
	timestamp := time.Now()
	logEntry := fmt.Sprintf(
		"[%s] [QUERY] Duration: %v, Query: %s, Error: %v\n",
		timestamp.Format("2006-01-02 15:04:05"),
		duration,
		formattedQuery,
		err,
	)

	// Print to console if PrintToConsole is enabled
	if l.options.PrintToConsole {
		fmt.Print(logEntry)
	}

	// Write to file if SaveToFile is enabled
	if l.options.SaveToFile && l.options.LogDir != "" {
		filename := filepath.Join(l.options.LogDir, timestamp.Format("2006_01_02")+".log")
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("Error opening log file: %v\n", err)
			return
		}
		defer file.Close()

		_, err = file.WriteString(logEntry)
		if err != nil {
			fmt.Printf("Error writing to log file: %v\n", err)
		}
	}
}

// cleanOldLogs removes log files older than the specified number of days
func (l *FileLogger) cleanOldLogs() {
	// Get all log files in the directory
	files, err := os.ReadDir(l.options.LogDir)
	if err != nil {
		return
	}

	// Calculate the cutoff time
	cutoff := time.Now().AddDate(0, 0, -l.options.AutoCleanDays)

	// Remove old files
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Parse the date from the filename
		filename := file.Name()
		if !strings.HasSuffix(filename, ".log") {
			continue
		}

		// Extract date from filename (format: 2006_01_02.log)
		dateStr := strings.TrimSuffix(filename, ".log")
		date, err := time.Parse("2006_01_02", dateStr)
		if err != nil {
			continue
		}

		// Remove if older than cutoff
		if date.Before(cutoff) {
			err := os.Remove(filepath.Join(l.options.LogDir, filename))
			if err != nil {
				fmt.Printf("Error removing old log file %s: %v\n", filename, err)
			}
		}
	}
}
