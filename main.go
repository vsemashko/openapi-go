package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/processor"
)

func main() {
	// Step 1: Set up context with cancellation on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown on SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal, cancelling operations...")
		cancel()
	}()

	// Step 2: Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	config.LogConfiguration(cfg)

	// Step 3: Process OpenAPI specs to generate clients
	if err := processor.ProcessOpenAPISpecs(ctx, cfg); err != nil {
		log.Fatalf("Error processing OpenAPI specs: %v", err)
	}

	log.Println("✅ Client generation completed successfully!")
}
