package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "json format",
			config: Config{
				Level:  "info",
				Format: "json",
			},
		},
		{
			name: "text format",
			config: Config{
				Level:  "debug",
				Format: "text",
			},
		},
		{
			name: "default format (json)",
			config: Config{
				Level:  "info",
				Format: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tt.config.Output = buf

			logger := New(tt.config)
			if logger == nil {
				t.Fatal("Expected logger to be created")
			}

			// Test that logger can write
			logger.Info("test message")
			if buf.Len() == 0 {
				t.Error("Expected logger to write output")
			}
		})
	}
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault()
	if logger == nil {
		t.Fatal("Expected default logger to be created")
	}

	// Test logging
	buf := &bytes.Buffer{}
	logger = New(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	logger.Info("test message", "key", "value")
	output := buf.String()

	// Verify JSON format
	var logEntry map[string]any
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	// Verify message
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg='test message', got '%v'", logEntry["msg"])
	}

	// Verify level
	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level='INFO', got '%v'", logEntry["level"])
	}

	// Verify custom field
	if logEntry["key"] != "value" {
		t.Errorf("Expected key='value', got '%v'", logEntry["key"])
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected string // We'll check the level name
	}{
		{"debug", "DEBUG"},
		{"info", "INFO"},
		{"warn", "WARN"},
		{"warning", "WARN"},
		{"error", "ERROR"},
		{"invalid", "INFO"}, // defaults to info
		{"", "INFO"},        // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLevel(tt.input)
			if level.String() != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, level, tt.expected)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     string
		logFunc      func(*Logger, string)
		shouldOutput bool
	}{
		{
			name:         "debug message at debug level",
			logLevel:     "debug",
			logFunc:      func(l *Logger, msg string) { l.Debug(msg) },
			shouldOutput: true,
		},
		{
			name:         "debug message at info level",
			logLevel:     "info",
			logFunc:      func(l *Logger, msg string) { l.Debug(msg) },
			shouldOutput: false,
		},
		{
			name:         "info message at info level",
			logLevel:     "info",
			logFunc:      func(l *Logger, msg string) { l.Info(msg) },
			shouldOutput: true,
		},
		{
			name:         "info message at error level",
			logLevel:     "error",
			logFunc:      func(l *Logger, msg string) { l.Info(msg) },
			shouldOutput: false,
		},
		{
			name:         "error message at info level",
			logLevel:     "info",
			logFunc:      func(l *Logger, msg string) { l.Error(msg) },
			shouldOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := New(Config{
				Level:  tt.logLevel,
				Format: "json",
				Output: buf,
			})

			tt.logFunc(logger, "test message")

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldOutput {
				t.Errorf("Expected output=%v, got output=%v", tt.shouldOutput, hasOutput)
			}
		})
	}
}

func TestWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	fields := map[string]any{
		"user_id":  "123",
		"action":   "create",
		"duration": 42,
	}

	logger.WithFields(fields).Info("test message")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify all fields are present
	if logEntry["user_id"] != "123" {
		t.Errorf("Expected user_id='123', got '%v'", logEntry["user_id"])
	}
	if logEntry["action"] != "create" {
		t.Errorf("Expected action='create', got '%v'", logEntry["action"])
	}
	if logEntry["duration"] != float64(42) {
		t.Errorf("Expected duration=42, got '%v'", logEntry["duration"])
	}
}

func TestWithField(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	logger.WithField("request_id", "abc-123").Info("test message")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if logEntry["request_id"] != "abc-123" {
		t.Errorf("Expected request_id='abc-123', got '%v'", logEntry["request_id"])
	}
}

func TestWithError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	testErr := errors.New("test error")
	logger.WithError(testErr).Error("operation failed")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if logEntry["error"] != "test error" {
		t.Errorf("Expected error='test error', got '%v'", logEntry["error"])
	}
}

func TestWithErrorNil(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	// WithError(nil) should not add error field
	logger.WithError(nil).Info("test message")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if _, hasError := logEntry["error"]; hasError {
		t.Error("Expected no error field when error is nil")
	}
}

func TestWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	ctx := context.Background()
	logger.WithContext(ctx).Info("test message")

	if buf.Len() == 0 {
		t.Error("Expected logger to write output")
	}
}

func TestTextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:  "info",
		Format: "text",
		Output: buf,
	})

	logger.Info("test message", "key", "value")
	output := buf.String()

	// Text format should contain the message and key=value
	if !strings.Contains(output, "test message") {
		t.Error("Expected output to contain message")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("Expected output to contain key=value")
	}
}

func TestChainedWithMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	testErr := errors.New("test error")
	logger.
		WithField("request_id", "req-123").
		WithField("user_id", "user-456").
		WithError(testErr).
		Error("operation failed")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify all chained fields
	if logEntry["request_id"] != "req-123" {
		t.Errorf("Expected request_id='req-123', got '%v'", logEntry["request_id"])
	}
	if logEntry["user_id"] != "user-456" {
		t.Errorf("Expected user_id='user-456', got '%v'", logEntry["user_id"])
	}
	if logEntry["error"] != "test error" {
		t.Errorf("Expected error='test error', got '%v'", logEntry["error"])
	}
	if logEntry["msg"] != "operation failed" {
		t.Errorf("Expected msg='operation failed', got '%v'", logEntry["msg"])
	}
}
