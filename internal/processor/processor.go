package processor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/preprocessor"
)

// ProcessOpenAPISpecs processes OpenAPI specifications and generates client code.
// It searches for OpenAPI specs in the specified directory that match the targetServices pattern,
// then generates Go client code for each spec using the ogen tool.
//
// Parameters:
// - config: Configuration containing specs directory, output directory, and target services pattern
//
// Returns an error if the process fails at any stage.
func ProcessOpenAPISpecs(cfg config.Config) error {
	// Setup the client output directory
	clientOutputDir := filepath.Join(cfg.OutputDir, "clients")
	if err := os.MkdirAll(clientOutputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create client output directory: %w", err)
	}

	// Find OpenAPI specs
	specs, err := findOpenAPISpecs(cfg.SpecsDir, cfg.TargetServices)
	if err != nil {
		return err
	}

	// Generate clients
	successCount, err := generateClients(specs, cfg.OutputDir)
	if err != nil {
		return err
	}

	log.Printf("Successfully processed %d/%d OpenAPI specs", successCount, len(specs))
	return nil
}

// findOpenAPISpecs searches for OpenAPI specs in the given directory.
func findOpenAPISpecs(specsDir string, targetServices string) ([]string, error) {
	// Compile service regex for filtering
	serviceRegex, err := compileServiceRegex(targetServices)
	if err != nil {
		return nil, err
	}

	var specs []string

	err = filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		// Skip directories and non-OpenAPI files
		if err != nil || info.IsDir() || filepath.Base(path) != "openapi.json" {
			return nil
		}

		// Check if service name matches the filter
		serviceDir := filepath.Base(filepath.Dir(path))
		if !serviceRegex.MatchString(serviceDir) {
			return nil
		}

		specs = append(specs, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find OpenAPI specs: %w", err)
	}

	if len(specs) == 0 {
		return nil, fmt.Errorf("no OpenAPI specs found for target services")
	}

	log.Printf("Found %d OpenAPI specs matching the criteria", len(specs))
	return specs, nil
}

// generateClients generates clients for all found OpenAPI specs.
func generateClients(specs []string, outputDir string) (int, error) {
	successCount := 0

	for _, specPath := range specs {
		serviceDir := filepath.Base(filepath.Dir(specPath))
		serviceName := normalizeServiceName(serviceDir)
		folderName := serviceName + "sdk"

		log.Printf("Processing service: %s (spec: %s)", serviceName, specPath)

		if err := generateClientForSpec(specPath, serviceName, folderName, outputDir); err != nil {
			log.Printf("Warning: Failed to generate client for %s: %v", folderName, err)
		} else {
			successCount++
		}
	}

	return successCount, nil
}

// generateClientForSpec generates a client for a single OpenAPI spec.
func generateClientForSpec(specPath, serviceName, folderName, outputDir string) error {
	// Create the client directory
	clientPath := filepath.Join(outputDir, "clients", folderName)
	if err := os.MkdirAll(clientPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create client directory for %s: %w", serviceName, err)
	}

	// Convert OpenAPI 3.1 to 3.0.3 if needed
	compatibleSpecPath, err := preprocessor.EnsureOpenAPICompatibility(specPath)
	if err != nil {
		return fmt.Errorf("failed to ensure OpenAPI compatibility for %s: %w", serviceName, err)
	}
	defer func() {
		// Clean up temporary file if one was created
		if compatibleSpecPath != specPath {
			os.Remove(compatibleSpecPath)
		}
	}()

	// Run the client generator
	if err := runOgenGenerator(folderName, compatibleSpecPath, clientPath); err != nil {
		return err
	}

	log.Printf("Successfully generated client for %s", folderName)
	return nil
}

// runOgenGenerator executes the ogen tool to generate a client from an OpenAPI spec.
func runOgenGenerator(serviceName, specPath, outputDir string) error {
	log.Printf("Generating client for %s...", serviceName)

	// Step 1: Ensure ogen CLI is installed
	if err := installOgenCLI(); err != nil {
		return fmt.Errorf("failed to install ogen CLI: %w", err)
	}

	// Step 2: Run ogen to generate the client
	cmd := exec.Command("ogen",
		"--target", outputDir,
		"--package", serviceName,
		"--clean",
		"--config", "ogen.yml",
		specPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ogen for %s: %w", serviceName, err)
	}

	return nil
}

// installOgenCLI ensures the ogen CLI tool is installed.
func installOgenCLI() error {
	cmd := exec.Command("go", "install", "github.com/ogen-go/ogen/cmd/ogen@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
