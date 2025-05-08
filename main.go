package main

import (
	"log"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/processor"
)

func main() {
	// Step 1: Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	config.LogConfiguration(cfg)

	// Step 2: Process OpenAPI specs to generate clients
	if err := processor.ProcessOpenAPISpecs(cfg); err != nil {
		log.Fatalf("Error processing OpenAPI specs: %v", err)
	}

	log.Println("Client generation completed successfully!")
}
