package errors

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRetry_Success(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  10 * time.Millisecond,
		MaxBackoff:      100 * time.Millisecond,
		BackoffMultiple: 2.0,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return nil // Success on first attempt
	}

	err := Retry(ctx, config, fn)
	if err != nil {
		t.Errorf("Retry() returned error: %v", err)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  10 * time.Millisecond,
		MaxBackoff:      100 * time.Millisecond,
		BackoffMultiple: 2.0,
		RetryableErrors: []ErrorCode{ErrCodeNetworkTimeout},
	}

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return New(ErrCodeNetworkTimeout, "timeout")
		}
		return nil // Success on third attempt
	}

	startTime := time.Now()
	err := Retry(ctx, config, fn)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Errorf("Retry() returned error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	// Should have waited: 10ms + 20ms = 30ms (plus some overhead)
	if elapsed < 30*time.Millisecond {
		t.Errorf("Expected at least 30ms elapsed, got %v", elapsed)
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  10 * time.Millisecond,
		MaxBackoff:      100 * time.Millisecond,
		BackoffMultiple: 2.0,
		RetryableErrors: []ErrorCode{ErrCodeNetworkTimeout},
	}

	attempts := 0
	fn := func() error {
		attempts++
		return New(ErrCodeFileNotFound, "file not found") // Non-retryable
	}

	err := Retry(ctx, config, fn)
	if err == nil {
		t.Error("Retry() should return error for non-retryable error")
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retries), got %d", attempts)
	}
}

func TestRetry_ExhaustedAttempts(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  10 * time.Millisecond,
		MaxBackoff:      100 * time.Millisecond,
		BackoffMultiple: 2.0,
		RetryableErrors: []ErrorCode{ErrCodeNetworkTimeout},
	}

	attempts := 0
	fn := func() error {
		attempts++
		return New(ErrCodeNetworkTimeout, "always fails")
	}

	err := Retry(ctx, config, fn)
	if err == nil {
		t.Error("Retry() should return error after exhausting attempts")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := RetryConfig{
		MaxAttempts:     10,
		InitialBackoff:  100 * time.Millisecond,
		MaxBackoff:      1 * time.Second,
		BackoffMultiple: 2.0,
		RetryableErrors: []ErrorCode{ErrCodeNetworkTimeout},
	}

	attempts := 0
	fn := func() error {
		attempts++
		if attempts == 2 {
			cancel() // Cancel after second attempt
		}
		return New(ErrCodeNetworkTimeout, "timeout")
	}

	err := Retry(ctx, config, fn)
	if err == nil {
		t.Error("Retry() should return error when context is cancelled")
	}
	if attempts < 2 {
		t.Errorf("Expected at least 2 attempts, got %d", attempts)
	}
}

func TestCalculateBackoff(t *testing.T) {
	config := RetryConfig{
		InitialBackoff:  1 * time.Second,
		MaxBackoff:      30 * time.Second,
		BackoffMultiple: 2.0,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 1 * time.Second},  // 1 * 2^0 = 1
		{2, 2 * time.Second},  // 1 * 2^1 = 2
		{3, 4 * time.Second},  // 1 * 2^2 = 4
		{4, 8 * time.Second},  // 1 * 2^3 = 8
		{5, 16 * time.Second}, // 1 * 2^4 = 16
		{6, 30 * time.Second}, // 1 * 2^5 = 32, capped at 30
		{7, 30 * time.Second}, // Capped at max
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt %d", tt.attempt), func(t *testing.T) {
			backoff := calculateBackoff(tt.attempt, config)
			if backoff != tt.expected {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, backoff, tt.expected)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected bool
	}{
		{"network timeout", ErrCodeNetworkTimeout, true},
		{"network unavailable", ErrCodeNetworkUnavailable, true},
		{"generator install", ErrCodeGeneratorInstall, true},
		{"cache read", ErrCodeCacheReadFailed, true},
		{"cache write", ErrCodeCacheWriteFailed, true},
		{"file not found", ErrCodeFileNotFound, false},
		{"spec parse error", ErrCodeSpecParseError, false},
		{"generator failed", ErrCodeGeneratorFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.code); got != tt.expected {
				t.Errorf("IsRetryableError(%v) = %v, want %v", tt.code, got, tt.expected)
			}
		})
	}
}

func TestRetryWithCallback(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  10 * time.Millisecond,
		MaxBackoff:      100 * time.Millisecond,
		BackoffMultiple: 2.0,
		RetryableErrors: []ErrorCode{ErrCodeNetworkTimeout},
	}

	attempts := 0
	callbackCount := 0

	fn := func() error {
		attempts++
		if attempts < 2 {
			return New(ErrCodeNetworkTimeout, "timeout")
		}
		return nil
	}

	onRetry := func(ctx RetryContext) {
		callbackCount++
		if ctx.Attempt < 1 || ctx.Attempt > config.MaxAttempts {
			t.Errorf("Invalid attempt number in callback: %d", ctx.Attempt)
		}
		if ctx.MaxAttempts != config.MaxAttempts {
			t.Errorf("MaxAttempts = %d, want %d", ctx.MaxAttempts, config.MaxAttempts)
		}
	}

	err := RetryWithCallback(ctx, config, fn, onRetry)
	if err != nil {
		t.Errorf("RetryWithCallback() returned error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
	if callbackCount != 2 {
		// Called once after the 1st failure, and once after success
		t.Errorf("Expected 2 callback calls, got %d", callbackCount)
	}
}

func TestRetryableOperation(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		attempts := 0
		err := RetryableOperation(ctx, "test operation", func() error {
			attempts++
			return nil
		})
		if err != nil {
			t.Errorf("RetryableOperation() returned error: %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("retries then succeeds", func(t *testing.T) {
		attempts := 0
		err := RetryableOperation(ctx, "test operation", func() error {
			attempts++
			if attempts < 2 {
				return New(ErrCodeNetworkTimeout, "timeout")
			}
			return nil
		})
		if err != nil {
			t.Errorf("RetryableOperation() returned error: %v", err)
		}
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		attempts := 0
		err := RetryableOperation(ctx, "test operation", func() error {
			attempts++
			return New(ErrCodeFileNotFound, "not found")
		})
		if err == nil {
			t.Error("RetryableOperation() should return error")
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})
}

func TestAs(t *testing.T) {
	t.Run("direct GenerationError", func(t *testing.T) {
		err := New(ErrCodeFileNotFound, "not found")
		var genErr *GenerationError
		if !As(err, &genErr) {
			t.Error("As() should return true for GenerationError")
		}
		if genErr.Code != ErrCodeFileNotFound {
			t.Errorf("Code = %v, want %v", genErr.Code, ErrCodeFileNotFound)
		}
	})

	t.Run("wrapped GenerationError", func(t *testing.T) {
		inner := New(ErrCodeFileNotFound, "not found")
		wrapped := Wrap(inner, ErrCodeGeneratorFailed, "generation failed")
		var genErr *GenerationError
		if !As(wrapped, &genErr) {
			t.Error("As() should return true for wrapped GenerationError")
		}
	})

	t.Run("standard error", func(t *testing.T) {
		err := fmt.Errorf("standard error")
		var genErr *GenerationError
		if As(err, &genErr) {
			t.Error("As() should return false for standard error")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		var genErr *GenerationError
		if As(nil, &genErr) {
			t.Error("As() should return false for nil error")
		}
	})
}
