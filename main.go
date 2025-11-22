package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/logger"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/processor"
)

func main() {
	// Step 1: Load configuration (before logger so we can configure it)
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use default logger for config load errors
		defaultLog := logger.NewDefault()
		defaultLog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Step 2: Initialize structured logger with config
	structuredLog := logger.New(logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
		Output: os.Stdout,
	})

	structuredLog.Info("Starting OpenAPI client generator")
	config.LogConfiguration(cfg, structuredLog)

	// Step 3: Set up context with cancellation on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown on SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		structuredLog.Warn("Received interrupt signal, cancelling operations...")
		cancel()
	}()

	// Step 4: Process OpenAPI specs to generate clients
	if err := processor.ProcessOpenAPISpecs(ctx, cfg, structuredLog); err != nil {
		structuredLog.Error("Error processing OpenAPI specs", "error", err)
		os.Exit(1)
	}

	structuredLog.Info("Client generation completed successfully")
}
