package processor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/paths"
)

const (
	// OgenVersion defines the exact ogen version to use for generation
	// This ensures deterministic builds - same spec always generates same code
	// IMPORTANT: This version must match the version in go.mod
	OgenVersion = "v1.14.0"
	OgenPackage = "github.com/ogen-go/ogen/cmd/ogen"
)

// ProcessingResult contains the results of processing OpenAPI specs
type ProcessingResult struct {
	TotalSpecs   int
	SuccessCount int
	FailedSpecs  []SpecFailure
}

// SpecFailure represents a failed spec generation
type SpecFailure struct {
	SpecPath    string
	ServiceName string
	Error       error
}

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
	result, err := generateClients(specs, cfg.OutputDir, cfg.ContinueOnError)
	if err != nil {
		return err
	}

	// Log results
	logProcessingResult(result)

	// Return error if any specs failed (unless continue-on-error is enabled)
	if !cfg.ContinueOnError && result.SuccessCount < result.TotalSpecs {
		return fmt.Errorf("failed to generate %d/%d clients",
			len(result.FailedSpecs), result.TotalSpecs)
	}

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
func generateClients(specs []string, outputDir string, continueOnError bool) (*ProcessingResult, error) {
	result := &ProcessingResult{
		TotalSpecs:   len(specs),
		SuccessCount: 0,
		FailedSpecs:  []SpecFailure{},
	}

	for _, specPath := range specs {
		serviceDir := filepath.Base(filepath.Dir(specPath))
		serviceName := normalizeServiceName(serviceDir)
		folderName := serviceName + "sdk"

		log.Printf("Processing service: %s (spec: %s)", serviceName, specPath)

		err := generateClientForSpec(specPath, serviceName, folderName, outputDir)
		if err != nil {
			failure := SpecFailure{
				SpecPath:    specPath,
				ServiceName: serviceName,
				Error:       err,
			}
			result.FailedSpecs = append(result.FailedSpecs, failure)

			log.Printf("❌ Failed to generate client for %s: %v", folderName, err)

			// Fail fast unless continue-on-error is enabled
			if !continueOnError {
				return result, fmt.Errorf("generation failed for %s: %w", serviceName, err)
			}
		} else {
			result.SuccessCount++
			log.Printf("✅ Successfully generated client for %s", folderName)
		}
	}

	return result, nil
}

// logProcessingResult logs a summary of the processing results
func logProcessingResult(result *ProcessingResult) {
	log.Printf("=====================================")
	log.Printf("SDK Generation Summary")
	log.Printf("=====================================")
	log.Printf("Total specs:    %d", result.TotalSpecs)
	log.Printf("Successful:     %d", result.SuccessCount)
	log.Printf("Failed:         %d", len(result.FailedSpecs))

	if len(result.FailedSpecs) > 0 {
		log.Printf("-------------------------------------")
		log.Printf("Failed specs:")
		for _, failure := range result.FailedSpecs {
			log.Printf("  - %s: %v", failure.ServiceName, failure.Error)
		}
	}
	log.Printf("=====================================")
}

// generateClientForSpec generates a client for a single OpenAPI spec.
func generateClientForSpec(specPath, serviceName, folderName, outputDir string) error {
	// Create the client directory
	clientPath := filepath.Join(outputDir, "clients", folderName)
	if err := os.MkdirAll(clientPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create client directory for %s: %w", serviceName, err)
	}

	// Clean existing files in the client directory
	log.Printf("Cleaning existing files for %s...", folderName)
	if err := cleanDirectory(clientPath); err != nil {
		return fmt.Errorf("failed to clean client directory for %s: %w", serviceName, err)
	}

	// Run the client generator
	if err := runOgenGenerator(folderName, specPath, clientPath); err != nil {
		return err
	}

	// Apply post-processors to the generated client
	log.Printf("Applying post-processors for %s...", folderName)
	if err := ApplyPostProcessors(clientPath, folderName, specPath); err != nil {
		return fmt.Errorf("failed to apply post-processors for %s: %w", folderName, err)
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

	// Step 2: Get absolute path to ogen config
	ogenConfigPath := paths.GetOgenConfigPath()
	if err := paths.EnsurePathExists(ogenConfigPath); err != nil {
		return fmt.Errorf("ogen config not found: %w", err)
	}

	// Step 3: Run ogen to generate the client
	cmd := exec.Command("ogen",
		"--target", outputDir,
		"--package", serviceName,
		"--clean",
		"--config", ogenConfigPath,
		specPath)

	// Capture output for better error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ogen failed for %s: %w\nOutput: %s", serviceName, err, string(output))
	}

	// Log ogen output
	if len(output) > 0 {
		log.Printf("ogen output for %s:\n%s", serviceName, string(output))
	}

	return nil
}

// installOgenCLI ensures the ogen CLI tool is installed with the correct version.
// It checks if ogen is already installed and verifies the version before installing.
func installOgenCLI() error {
	// Check if ogen is already installed with the correct version
	if isOgenInstalled() {
		log.Printf("ogen CLI %s already installed, skipping installation", OgenVersion)
		return nil
	}

	log.Printf("Installing ogen CLI %s...", OgenVersion)

	// Install specific version (not @latest for deterministic builds)
	cmd := exec.Command("go", "install", fmt.Sprintf("%s@%s", OgenPackage, OgenVersion))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install ogen: %w\nOutput: %s", err, string(output))
	}

	// Verify installation succeeded
	if !isOgenInstalled() {
		return fmt.Errorf("ogen installation verification failed")
	}

	log.Printf("ogen CLI %s installed successfully", OgenVersion)
	return nil
}

// isOgenInstalled checks if ogen is available in PATH with the correct version
func isOgenInstalled() bool {
	cmd := exec.Command("ogen", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	// Parse version from output
	// Expected format: "ogen version v1.14.0" or similar
	versionOutput := strings.TrimSpace(string(output))

	// Check if the output contains our expected version
	return strings.Contains(versionOutput, OgenVersion)
}
