package generator

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/errors"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/paths"
)

const (
	// OgenName is the name identifier for the ogen generator
	OgenName = "ogen"

	// OgenVersion defines the exact ogen version to use for generation
	// This ensures deterministic builds - same spec always generates same code
	// IMPORTANT: This version must match the version in go.mod
	OgenVersion = "v1.14.0"

	// OgenPackage is the full Go package path for the ogen CLI
	OgenPackage = "github.com/ogen-go/ogen/cmd/ogen"
)

// OgenGenerator implements the Generator interface for the ogen code generator
type OgenGenerator struct {
	version string
	pkg     string
}

// NewOgenGenerator creates a new ogen generator instance
func NewOgenGenerator() *OgenGenerator {
	return &OgenGenerator{
		version: OgenVersion,
		pkg:     OgenPackage,
	}
}

// Name returns the generator name
func (g *OgenGenerator) Name() string {
	return OgenName
}

// Version returns the generator version
func (g *OgenGenerator) Version() string {
	return g.version
}

// IsInstalled checks if ogen is available in PATH with the correct version
func (g *OgenGenerator) IsInstalled() bool {
	cmd := exec.Command("ogen", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	// Parse version from output
	// Expected format: "ogen version v1.14.0" or similar
	versionOutput := strings.TrimSpace(string(output))

	// Check if the output contains our expected version
	return strings.Contains(versionOutput, g.version)
}

// EnsureInstalled ensures the ogen CLI is installed with the correct version
// Uses retry logic with exponential backoff for network failures
func (g *OgenGenerator) EnsureInstalled(ctx context.Context) error {
	// Check if already installed with correct version
	if g.IsInstalled() {
		log.Printf("ogen CLI %s already installed, skipping installation", g.version)
		return nil
	}

	log.Printf("Installing ogen CLI %s...", g.version)

	// Install with retry logic for transient failures (network issues)
	err := errors.RetryableOperation(ctx, "install ogen", func() error {
		// Install specific version (not @latest for deterministic builds)
		cmd := exec.CommandContext(ctx, "go", "install", fmt.Sprintf("%s@%s", g.pkg, g.version))
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Wrap error with structured error type
			return errors.Wrap(err, errors.ErrCodeGeneratorInstall,
				fmt.Sprintf("failed to install ogen %s", g.version)).
				WithContext("output", string(output)).
				WithSuggestion("Check your network connection and Go installation")
		}

		// Verify installation succeeded
		if !g.IsInstalled() {
			return errors.New(errors.ErrCodeGeneratorInstall,
				"ogen installation verification failed").
				WithSuggestion("Try running: go install "+g.pkg+"@"+g.version)
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Printf("✅ ogen CLI %s installed successfully", g.version)
	return nil
}

// Generate generates client code using ogen
func (g *OgenGenerator) Generate(ctx context.Context, spec GenerateSpec) error {
	// Ensure ogen is installed
	if err := g.EnsureInstalled(ctx); err != nil {
		return errors.Wrap(err, errors.ErrCodeGeneratorNotFound, "ogen CLI not available")
	}

	// Validate spec path
	if err := paths.EnsurePathExists(spec.SpecPath); err != nil {
		return errors.Wrap(err, errors.ErrCodeFileNotFound, "spec file not found").
			WithContext("spec", spec.SpecPath).
			WithSuggestion("Check if the OpenAPI spec file exists at the specified path")
	}

	// Validate config path if provided
	configPath := spec.ConfigPath
	if configPath == "" {
		configPath = paths.GetOgenConfigPath()
	}
	if err := paths.EnsurePathExists(configPath); err != nil {
		return errors.Wrap(err, errors.ErrCodeConfigMissing, "ogen config not found").
			WithContext("config", configPath).
			WithSuggestion("Create ogen config file or check the path")
	}

	// Build command arguments
	args := []string{
		"--target", spec.OutputDir,
		"--package", spec.PackageName,
		"--config", configPath,
	}

	if spec.Clean {
		args = append(args, "--clean")
	}

	args = append(args, spec.SpecPath)

	// Execute ogen
	log.Printf("Generating client with ogen for package %s...", spec.PackageName)
	cmd := exec.CommandContext(ctx, "ogen", args...)

	// Capture output for better error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Create structured error with ogen output in context
		return errors.Wrap(err, errors.ErrCodeGeneratorFailed,
			fmt.Sprintf("ogen failed for package %s", spec.PackageName)).
			WithContext("package", spec.PackageName).
			WithContext("spec", spec.SpecPath).
			WithContext("ogen_error", string(output)).
			WithSuggestion("Check the ogen error message above for specific issues")
	}

	// Log ogen output
	if len(output) > 0 {
		log.Printf("ogen output for %s:\n%s", spec.PackageName, string(output))
	}

	return nil
}

// Validate checks if the generator configuration is valid
func (g *OgenGenerator) Validate() error {
	if g.version == "" {
		return fmt.Errorf("ogen version not set")
	}
	if g.pkg == "" {
		return fmt.Errorf("ogen package path not set")
	}
	return nil
}
