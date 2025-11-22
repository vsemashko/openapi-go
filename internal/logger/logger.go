package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with additional convenience methods
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output io.Writer
}

// New creates a new structured logger with the specified configuration
func New(cfg Config) *Logger {
	// Parse log level
	level := parseLevel(cfg.Level)

	// Set output writer (default to stdout)
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		handler = slog.NewTextHandler(output, opts)
	default:
		// Default to JSON for production
		handler = slog.NewJSONHandler(output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewDefault creates a logger with default settings (INFO level, JSON format)
func NewDefault() *Logger {
	return New(Config{
		Level:  "info",
		Format: "json",
		Output: os.Stdout,
	})
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext returns a logger with context values
// This can be extended to extract values from context (request ID, user ID, etc.)
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// In the future, extract context values here
	// For now, return the same logger
	return l
}

// WithFields returns a logger with additional structured fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	if len(fields) == 0 {
		return l
	}

	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		Logger: l.With(args...),
	}
}

// WithField returns a logger with a single additional field
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{
		Logger: l.With(key, value),
	}
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return &Logger{
		Logger: l.With("error", err.Error()),
	}
}
