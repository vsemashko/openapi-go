package errors

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts     int           // Maximum number of retry attempts
	InitialBackoff  time.Duration // Initial backoff duration
	MaxBackoff      time.Duration // Maximum backoff duration
	BackoffMultiple float64       // Multiplier for exponential backoff
	RetryableErrors []ErrorCode   // List of error codes that should trigger retries
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialBackoff:  1 * time.Second,
		MaxBackoff:      30 * time.Second,
		BackoffMultiple: 2.0,
		RetryableErrors: []ErrorCode{
			ErrCodeNetworkTimeout,
			ErrCodeNetworkUnavailable,
			ErrCodeGeneratorInstall, // May fail due to network
			ErrCodeCacheReadFailed,  // May be locked temporarily
			ErrCodeCacheWriteFailed, // May be locked temporarily
		},
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// RetryContext contains information about the current retry attempt
type RetryContext struct {
	Attempt      int
	MaxAttempts  int
	LastError    error
	TotalElapsed time.Duration
}

// Retry executes a function with exponential backoff retry logic
func Retry(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	var lastErr error
	startTime := time.Now()

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()

		// Success case
		if err == nil {
			if attempt > 1 {
				log.Printf("Operation succeeded after %d attempt(s)", attempt)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err, config.RetryableErrors) {
			return err // Not retryable, fail immediately
		}

		// Check if we've exhausted attempts
		if attempt >= config.MaxAttempts {
			break
		}

		// Check context cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}

		// Calculate backoff duration
		backoff := calculateBackoff(attempt, config)

		// Log retry attempt
		log.Printf("Operation failed (attempt %d/%d), retrying in %v: %v",
			attempt, config.MaxAttempts, backoff, err)

		// Wait for backoff duration (with context cancellation support)
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	// All attempts exhausted
	elapsed := time.Since(startTime)
	return fmt.Errorf("operation failed after %d attempts (%v): %w",
		config.MaxAttempts, elapsed, lastErr)
}

// RetryWithCallback executes a function with retry logic and calls a callback for each attempt
func RetryWithCallback(
	ctx context.Context,
	config RetryConfig,
	fn RetryableFunc,
	onRetry func(ctx RetryContext),
) error {
	var lastErr error
	startTime := time.Now()

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()

		// Update retry context
		retryCtx := RetryContext{
			Attempt:      attempt,
			MaxAttempts:  config.MaxAttempts,
			LastError:    err,
			TotalElapsed: time.Since(startTime),
		}

		// Success case
		if err == nil {
			if attempt > 1 && onRetry != nil {
				onRetry(retryCtx)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err, config.RetryableErrors) {
			return err // Not retryable, fail immediately
		}

		// Check if we've exhausted attempts
		if attempt >= config.MaxAttempts {
			break
		}

		// Check context cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}

		// Calculate backoff duration
		backoff := calculateBackoff(attempt, config)

		// Call retry callback
		if onRetry != nil {
			onRetry(retryCtx)
		}

		// Wait for backoff duration
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	// All attempts exhausted
	elapsed := time.Since(startTime)
	return fmt.Errorf("operation failed after %d attempts (%v): %w",
		config.MaxAttempts, elapsed, lastErr)
}

// isRetryable checks if an error should trigger a retry
func isRetryable(err error, retryableErrors []ErrorCode) bool {
	if err == nil {
		return false
	}

	// Check if it's a GenerationError with a retryable code
	var genErr *GenerationError
	if As(err, &genErr) {
		for _, code := range retryableErrors {
			if genErr.Code == code {
				return true
			}
		}
	}

	return false
}

// calculateBackoff calculates the backoff duration for a given attempt
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	// Exponential backoff: initialBackoff * (multiplier ^ (attempt - 1))
	backoff := float64(config.InitialBackoff) * math.Pow(config.BackoffMultiple, float64(attempt-1))

	// Cap at max backoff
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}

	return time.Duration(backoff)
}

// IsRetryableError checks if a specific error code is retryable in the default config
func IsRetryableError(code ErrorCode) bool {
	config := DefaultRetryConfig()
	for _, retryable := range config.RetryableErrors {
		if code == retryable {
			return true
		}
	}
	return false
}

// As is a helper that wraps errors.As for GenerationError
func As(err error, target **GenerationError) bool {
	var genErr *GenerationError
	if ok := asGenerationError(err, &genErr); ok {
		*target = genErr
		return true
	}
	return false
}

func asGenerationError(err error, target **GenerationError) bool {
	if err == nil {
		return false
	}

	// Direct type assertion
	if genErr, ok := err.(*GenerationError); ok {
		*target = genErr
		return true
	}

	// Check if error wraps a GenerationError
	type unwrapper interface {
		Unwrap() error
	}

	if u, ok := err.(unwrapper); ok {
		return asGenerationError(u.Unwrap(), target)
	}

	return false
}

// RetryableOperation wraps a function with default retry logic
func RetryableOperation(ctx context.Context, operation string, fn RetryableFunc) error {
	config := DefaultRetryConfig()

	return RetryWithCallback(ctx, config, fn, func(retryCtx RetryContext) {
		if retryCtx.LastError != nil {
			backoff := calculateBackoff(retryCtx.Attempt, config)
			log.Printf("[%s] Attempt %d/%d failed, retrying in %v: %v",
				operation, retryCtx.Attempt, retryCtx.MaxAttempts, backoff, retryCtx.LastError)
		} else {
			log.Printf("[%s] Succeeded after %d attempt(s)", operation, retryCtx.Attempt)
		}
	})
}
