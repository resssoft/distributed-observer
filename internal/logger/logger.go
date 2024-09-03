package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger is a wrapper around slog.Logger with additional functionality
type Logger struct {
	logger *slog.Logger
}

// New creates a new Logger instance
func New(level slog.Leveler, output *os.File) *Logger {
	if output == nil {
		output = os.Stdout
	}
	if level == nil {
		level = &slog.LevelVar{}
	}
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(output, opts)
	logger := slog.New(handler).With()
	return &Logger{
		logger: logger,
	}
}

// Info logs an informational message with optional context
func (l *Logger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.logger.InfoContext(ctx, msg, args...)
}

// Debug logs a debug message with optional context
func (l *Logger) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.logger.DebugContext(ctx, msg, args...)
}

// Error logs an error message with optional context
func (l *Logger) Error(ctx context.Context, err error, msg string, args ...interface{}) {
	l.logger.ErrorContext(ctx, msg, append([]interface{}{"error", err}, args...)...)
}

// Warn logs a warning message with optional context
func (l *Logger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.logger.WarnContext(ctx, msg, args...)
}

// With adds a context key-value pair to the logger
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
	}
}

// Example usage:
func example() {
	// Create a new logger instance
	log := New(slog.LevelInfo, os.Stdout)

	// Use the logger
	log.Info(context.Background(), "Starting the application")
	log.Debug(context.Background(), "This is a debug message")
	log.Warn(context.Background(), "This is a warning message")
	log.Error(context.Background(), nil, "An error occurred", "code", 500)

	// Using context with the logger
	ctx := context.WithValue(context.Background(), "request_id", "12345")
	log.With("request_id", "12345").Info(ctx, "Processing request")
}
