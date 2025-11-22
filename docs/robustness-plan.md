# OpenAPI Go SDK Generator - Robustness Improvement Plan

**Version:** 1.0
**Date:** 2025-11-22
**Author:** SDK Team

## Table of Contents

1. [Overview](#overview)
2. [Phase 1: Critical Fixes (P0)](#phase-1-critical-fixes-p0)
3. [Phase 2: High Priority (P1)](#phase-2-high-priority-p1)
4. [Phase 3: Medium Priority (P2)](#phase-3-medium-priority-p2)
5. [Phase 4: Enhancements (P3-P4)](#phase-4-enhancements-p3-p4)
6. [Implementation Guidelines](#implementation-guidelines)

---

## Overview

This document provides detailed, actionable solutions for each issue identified in `analysis.md`. Each section includes:
- Problem description
- Proposed solution with code examples
- Testing recommendations
- Migration considerations

**Guiding Principles:**
1. ✅ No hardcoded values - everything configurable
2. ✅ Fail fast with clear error messages
3. ✅ Deterministic builds - same input = same output
4. ✅ Comprehensive testing at all levels
5. ✅ Backward compatibility where possible

---

## Phase 1: Critical Fixes (P0)

### 1.1 Pin Ogen Version (P0)

**Problem:** `go install github.com/ogen-go/ogen/cmd/ogen@latest` is non-deterministic

**Solution:** Use exact version from `go.mod`

#### Implementation

**File:** `internal/processor/processor.go`

```go
package processor

import (
    "fmt"
    "log"
    "os/exec"
    "strings"
)

const (
    // OgenVersion should match the version in go.mod
    // This ensures deterministic builds
    OgenVersion = "v1.14.0"
    OgenPackage = "github.com/ogen-go/ogen/cmd/ogen"
)

// installOgenCLI ensures the ogen CLI tool is installed with the correct version.
// It checks if ogen is already installed and verifies the version before installing.
func installOgenCLI() error {
    // Check if ogen is already installed
    if isOgenInstalled() {
        log.Printf("ogen CLI already installed, skipping installation")
        return nil
    }

    log.Printf("Installing ogen CLI version %s...", OgenVersion)

    cmd := exec.Command("go", "install", fmt.Sprintf("%s@%s", OgenPackage, OgenVersion))
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("failed to install ogen: %w\nOutput: %s", err, string(output))
    }

    // Verify installation
    if !isOgenInstalled() {
        return fmt.Errorf("ogen installation verification failed")
    }

    log.Printf("ogen CLI version %s installed successfully", OgenVersion)
    return nil
}

// isOgenInstalled checks if ogen is available in PATH with correct version
func isOgenInstalled() bool {
    cmd := exec.Command("ogen", "--version")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return false
    }

    // Parse version from output
    // Expected format: "ogen version v1.14.0"
    version := strings.TrimSpace(string(output))
    expectedVersion := fmt.Sprintf("ogen version %s", OgenVersion)

    return strings.Contains(version, OgenVersion)
}
```

**Alternative: Use tools.go pattern**

**File:** `tools/tools.go` (new file)

```go
//go:build tools

package tools

import (
    _ "github.com/ogen-go/ogen/cmd/ogen"
)
```

**File:** `go.mod`

```go
// Add to go.mod to ensure ogen is available as a dependency
require (
    github.com/ogen-go/ogen v1.14.0
)
```

**File:** `Taskfile.yml`

```yaml
tasks:
  install-tools:
    desc: Install required tools
    cmds:
      - go install github.com/ogen-go/ogen/cmd/ogen@v1.14.0
    status:
      - command -v ogen
```

**Benefits:**
- ✅ Deterministic builds
- ✅ Version controlled
- ✅ No redundant installations
- ✅ Fast build times

**Testing:**
```go
func TestOgenVersionConsistency(t *testing.T) {
    // Ensure OgenVersion constant matches go.mod
    goModVersion := getOgenVersionFromGoMod(t)
    assert.Equal(t, goModVersion, OgenVersion)
}

func TestOgenInstallation(t *testing.T) {
    err := installOgenCLI()
    require.NoError(t, err)

    // Verify ogen is available
    cmd := exec.Command("ogen", "--version")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)
    assert.Contains(t, string(output), OgenVersion)
}
```

---

### 1.2 Fix Hardcoded Paths (P0)

**Problem:** All paths relative to CWD, breaks when run from different directories

**Solution:** Calculate absolute paths based on executable or repository root

#### Implementation

**File:** `internal/paths/paths.go` (new package)

```go
package paths

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
)

var (
    // Cached paths calculated once at startup
    executableDir string
    repositoryRoot string
)

func init() {
    var err error

    // Get executable directory
    executableDir, err = getExecutableDir()
    if err != nil {
        panic(fmt.Sprintf("failed to determine executable directory: %v", err))
    }

    // Find repository root by looking for go.mod
    repositoryRoot, err = findRepositoryRoot(executableDir)
    if err != nil {
        panic(fmt.Sprintf("failed to determine repository root: %v", err))
    }
}

// getExecutableDir returns the directory containing the current executable
func getExecutableDir() (string, error) {
    ex, err := os.Executable()
    if err != nil {
        return "", err
    }
    return filepath.Dir(ex), nil
}

// findRepositoryRoot walks up the directory tree to find go.mod
func findRepositoryRoot(startDir string) (string, error) {
    dir := startDir

    for {
        // Check if go.mod exists in current directory
        goModPath := filepath.Join(dir, "go.mod")
        if _, err := os.Stat(goModPath); err == nil {
            return dir, nil
        }

        // Move up one directory
        parent := filepath.Dir(dir)
        if parent == dir {
            // Reached filesystem root without finding go.mod
            return "", fmt.Errorf("repository root not found (no go.mod)")
        }
        dir = parent
    }
}

// GetRepositoryRoot returns the absolute path to repository root
func GetRepositoryRoot() string {
    return repositoryRoot
}

// GetOgenConfigPath returns the absolute path to ogen.yml
func GetOgenConfigPath() string {
    return filepath.Join(repositoryRoot, "ogen.yml")
}

// GetTemplatesDir returns the absolute path to templates directory
func GetTemplatesDir() string {
    return filepath.Join(repositoryRoot, "resources", "templates")
}

// GetInternalClientTemplatePath returns path to internal client template
func GetInternalClientTemplatePath() string {
    return filepath.Join(GetTemplatesDir(), "internal_client.tmpl")
}

// GetConfigPath returns the absolute path to application.yml
func GetConfigPath() string {
    return filepath.Join(repositoryRoot, "resources", "application.yml")
}

// EnsurePathExists verifies that a path exists and is accessible
func EnsurePathExists(path string) error {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return fmt.Errorf("path does not exist: %s", path)
    } else if err != nil {
        return fmt.Errorf("cannot access path %s: %w", path, err)
    }
    return nil
}

// EnsureDirectoryWritable checks if directory is writable
func EnsureDirectoryWritable(dir string) error {
    // Try to create a temporary file
    testFile := filepath.Join(dir, ".write_test_"+randomString())
    f, err := os.Create(testFile)
    if err != nil {
        return fmt.Errorf("directory not writable: %s: %w", dir, err)
    }
    f.Close()
    os.Remove(testFile)
    return nil
}

// randomString generates a random string for test files
func randomString() string {
    return fmt.Sprintf("%d", os.Getpid())
}
```

**File:** `internal/config/config.go` (updated)

```go
package config

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "regexp"
    "strings"

    "github.com/spf13/viper"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/paths"
)

// Config holds all configuration parameters for the application
type Config struct {
    // SpecsDir is the directory containing OpenAPI specification files
    SpecsDir string `mapstructure:"specs_dir"`

    // OutputDir is the base directory where generated clients will be stored
    OutputDir string `mapstructure:"output_dir"`

    // TargetServices is a regular expression pattern to filter services
    TargetServices string `mapstructure:"target_services"`
}

// LoadConfig initializes Viper and loads configuration from application.yml
// with the ability to override via environment variables
func LoadConfig() (Config, error) {
    v := viper.New()

    // Set up config file support with absolute paths
    configPath := paths.GetConfigPath()
    configDir := filepath.Dir(configPath)

    v.SetConfigName("application")
    v.SetConfigType("yml")
    v.AddConfigPath(configDir)

    // Also check user home directory
    if home, err := os.UserHomeDir(); err == nil {
        v.AddConfigPath(filepath.Join(home, ".openapi-go"))
    }

    // Enable automatic environment variable binding
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

    // Try to read config file
    if err := v.ReadInConfig(); err != nil {
        return Config{}, fmt.Errorf("error reading config file: %w", err)
    }

    log.Printf("Using config file: %s", v.ConfigFileUsed())

    // Unmarshal config into struct
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return Config{}, fmt.Errorf("unable to decode config into struct: %w", err)
    }

    // Convert relative paths to absolute paths
    cfg.SpecsDir = makeAbsolutePath(cfg.SpecsDir)
    cfg.OutputDir = makeAbsolutePath(cfg.OutputDir)

    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return Config{}, fmt.Errorf("invalid configuration: %w", err)
    }

    return cfg, nil
}

// makeAbsolutePath converts a relative path to absolute based on repository root
func makeAbsolutePath(p string) string {
    if filepath.IsAbs(p) {
        return p
    }
    return filepath.Join(paths.GetRepositoryRoot(), p)
}

// Validate checks if the configuration is valid
func Validate(cfg *Config) error {
    // Validate SpecsDir exists
    if cfg.SpecsDir == "" {
        return fmt.Errorf("specs_dir is required")
    }
    if err := paths.EnsurePathExists(cfg.SpecsDir); err != nil {
        return fmt.Errorf("specs_dir validation failed: %w", err)
    }

    // Validate OutputDir
    if cfg.OutputDir == "" {
        return fmt.Errorf("output_dir is required")
    }

    // Create output directory if it doesn't exist
    if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
        return fmt.Errorf("failed to create output_dir: %w", err)
    }

    // Check if output directory is writable
    if err := paths.EnsureDirectoryWritable(cfg.OutputDir); err != nil {
        return fmt.Errorf("output_dir validation failed: %w", err)
    }

    // Validate TargetServices regex
    if cfg.TargetServices != "" {
        if _, err := regexp.Compile(cfg.TargetServices); err != nil {
            return fmt.Errorf("target_services is not a valid regex: %w", err)
        }
    }

    return nil
}

// LogConfiguration logs the current configuration parameters
func LogConfiguration(cfg Config) {
    log.Printf("Configuration loaded:")
    log.Printf("  Repository root: %s", paths.GetRepositoryRoot())
    log.Printf("  Specs directory: %s", cfg.SpecsDir)
    log.Printf("  Output directory: %s", cfg.OutputDir)
    log.Printf("  Target services: %s", cfg.TargetServices)
    log.Printf("  Ogen config: %s", paths.GetOgenConfigPath())
}
```

**File:** `internal/processor/processor.go` (updated)

```go
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
        "--config", ogenConfigPath,  // ✅ Now uses absolute path
        specPath)

    // Capture output for better error messages
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("ogen failed for %s: %w\nOutput: %s", serviceName, err, string(output))
    }

    return nil
}
```

**File:** `internal/processor/postprocessor.go` (updated)

```go
// generateInternalClientFile creates a file with the NewInternalClient function
func generateInternalClientFile(clientPath, serviceName string) error {
    // Use absolute path to template
    templatePath := paths.GetInternalClientTemplatePath()

    // Verify template exists
    if err := paths.EnsurePathExists(templatePath); err != nil {
        return fmt.Errorf("template not found: %w", err)
    }

    // Rest of function remains the same...
}
```

**Benefits:**
- ✅ Works from any directory
- ✅ Can be run as installed binary
- ✅ Fails fast with clear errors if paths are wrong
- ✅ Portable across environments

---

### 1.3 Fail on Any Spec Failure (P0)

**Problem:** Partial failures are silent, CI shows green even when SDKs missing

**Solution:** Fail fast by default, add flag for continue-on-error behavior

#### Implementation

**File:** `internal/config/config.go` (updated)

```go
type Config struct {
    SpecsDir       string `mapstructure:"specs_dir"`
    OutputDir      string `mapstructure:"output_dir"`
    TargetServices string `mapstructure:"target_services"`

    // ContinueOnError allows generation to continue even if some specs fail
    // Default: false (fail fast on first error)
    ContinueOnError bool `mapstructure:"continue_on_error"`
}
```

**File:** `internal/processor/processor.go` (updated)

```go
// ProcessingResult contains the results of processing OpenAPI specs
type ProcessingResult struct {
    TotalSpecs    int
    SuccessCount  int
    FailedSpecs   []SpecFailure
}

// SpecFailure represents a failed spec generation
type SpecFailure struct {
    SpecPath    string
    ServiceName string
    Error       error
}

// ProcessOpenAPISpecs processes OpenAPI specifications and generates client code.
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
```

**File:** `resources/application.yml` (updated)

```yaml
# OpenAPI Go Generator Configuration

# Directory containing OpenAPI specs
specs_dir: "./external/sdk/sdk-packages"

# Output directory for generated clients
output_dir: "./generated"

# Regex pattern to filter services
target_services: "(funding-server-sdk|holidays-server-sdk)"

# Continue processing even if some specs fail (default: false)
# Set to true for development, keep false for CI/CD
continue_on_error: false
```

**Environment variable override:**

```bash
# Allow partial failures in development
CONTINUE_ON_ERROR=true go run main.go

# Strict mode for CI (default)
go run main.go
```

**File:** `main.go` (updated)

```go
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
        // Exit with non-zero status code
        os.Exit(1)
    }

    log.Println("✅ Client generation completed successfully!")
}
```

**Benefits:**
- ✅ CI/CD detects failures immediately
- ✅ Clear error reporting showing which specs failed
- ✅ Non-zero exit code on failure
- ✅ Flexibility for development with continue-on-error flag

**Testing:**

```go
func TestFailFastBehavior(t *testing.T) {
    cfg := config.Config{
        SpecsDir:        "./testdata/specs",
        OutputDir:       t.TempDir(),
        TargetServices:  ".*",
        ContinueOnError: false, // Fail fast mode
    }

    err := processor.ProcessOpenAPISpecs(cfg)

    // Should fail because testdata contains invalid spec
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to generate")
}

func TestContinueOnError(t *testing.T) {
    cfg := config.Config{
        SpecsDir:        "./testdata/specs",
        OutputDir:       t.TempDir(),
        TargetServices:  ".*",
        ContinueOnError: true, // Continue mode
    }

    err := processor.ProcessOpenAPISpecs(cfg)

    // Should succeed but report failures
    assert.NoError(t, err)
}
```

---

## Phase 2: High Priority (P1)

### 2.1 Implement Proper Security Detection (P1)

**Problem:** Security detection relies on file existence check, brittle

**Solution:** Parse OpenAPI spec directly to detect security schemes

#### Implementation

**File:** `internal/spec/parser.go` (new package)

```go
package spec

import (
    "encoding/json"
    "fmt"
    "os"
)

// OpenAPISpec represents a minimal OpenAPI specification structure
// We only parse the parts we need for security detection
type OpenAPISpec struct {
    OpenAPI  string                    `json:"openapi"`
    Info     map[string]interface{}    `json:"info"`
    Security []map[string][]string     `json:"security,omitempty"`
    Components *Components             `json:"components,omitempty"`
}

// Components represents the components section of OpenAPI spec
type Components struct {
    SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme definition
type SecurityScheme struct {
    Type        string `json:"type"`
    Scheme      string `json:"scheme,omitempty"`
    BearerFormat string `json:"bearerFormat,omitempty"`
}

// ParseSpecFile parses an OpenAPI specification file
func ParseSpecFile(specPath string) (*OpenAPISpec, error) {
    data, err := os.ReadFile(specPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read spec file: %w", err)
    }

    var spec OpenAPISpec
    if err := json.Unmarshal(data, &spec); err != nil {
        return nil, fmt.Errorf("failed to parse spec JSON: %w", err)
    }

    return &spec, nil
}

// HasSecurity checks if the spec defines any security requirements
func (s *OpenAPISpec) HasSecurity() bool {
    // Check global security requirements
    if len(s.Security) > 0 {
        return true
    }

    // Check if security schemes are defined
    if s.Components != nil && len(s.Components.SecuritySchemes) > 0 {
        return true
    }

    return false
}

// GetSecuritySchemes returns all defined security schemes
func (s *OpenAPISpec) GetSecuritySchemes() map[string]SecurityScheme {
    if s.Components == nil {
        return nil
    }
    return s.Components.SecuritySchemes
}
```

**File:** `internal/processor/postprocessor.go` (updated)

```go
package processor

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "text/template"

    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/paths"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/spec"
)

// ApplyPostProcessors applies post-processing steps to the generated client code.
func ApplyPostProcessors(clientPath, serviceName, specPath string) error {
    // Generate the internal client file
    if err := generateInternalClientFile(clientPath, serviceName, specPath); err != nil {
        return fmt.Errorf("failed to generate internal client file: %w", err)
    }

    return nil
}

// generateInternalClientFile creates a file with the NewInternalClient function
func generateInternalClientFile(clientPath, serviceName, specPath string) error {
    // Use absolute path to template
    templatePath := paths.GetInternalClientTemplatePath()

    // Verify template exists
    if err := paths.EnsurePathExists(templatePath); err != nil {
        return fmt.Errorf("template not found: %w", err)
    }

    // Parse OpenAPI spec to detect security requirements
    hasSecurity, err := detectSecurityFromSpec(specPath)
    if err != nil {
        // Fall back to file-based detection if spec parsing fails
        log.Printf("Warning: Failed to parse spec for security detection, falling back to file check: %v", err)
        hasSecurity = detectSecurityFromGeneratedFiles(clientPath)
    }

    log.Printf("Security detection for %s: hasSecurity=%v", serviceName, hasSecurity)

    // Create the template data
    data := struct {
        PackageName string
        HasSecurity bool
    }{
        PackageName: serviceName,
        HasSecurity: hasSecurity,
    }

    // Parse the template from file
    tmpl, err := template.ParseFiles(templatePath)
    if err != nil {
        return fmt.Errorf("failed to parse template file %s: %w", templatePath, err)
    }

    // Create the output file
    outputPath := filepath.Join(clientPath, "oas_internal_client_gen.go")
    file, err := os.Create(outputPath)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    // Execute the template
    if err := tmpl.ExecuteTemplate(file, filepath.Base(templatePath), data); err != nil {
        return fmt.Errorf("failed to execute template: %w", err)
    }

    return nil
}

// detectSecurityFromSpec parses the OpenAPI spec to check for security schemes
func detectSecurityFromSpec(specPath string) (bool, error) {
    openAPISpec, err := spec.ParseSpecFile(specPath)
    if err != nil {
        return false, err
    }

    return openAPISpec.HasSecurity(), nil
}

// detectSecurityFromGeneratedFiles checks for security file (fallback method)
func detectSecurityFromGeneratedFiles(clientPath string) bool {
    securityFilePath := filepath.Join(clientPath, "oas_security_gen.go")
    _, err := os.Stat(securityFilePath)
    return err == nil
}
```

**Benefits:**
- ✅ Robust: Based on spec content, not generated file names
- ✅ Predictable: Works even if ogen changes file naming
- ✅ Testable: Easy to verify behavior
- ✅ Fallback: Still works if spec parsing fails

**Testing:**

```go
func TestSecurityDetection(t *testing.T) {
    tests := []struct {
        name     string
        spec     string
        expected bool
    }{
        {
            name: "spec with bearer auth",
            spec: `{
                "openapi": "3.0.0",
                "components": {
                    "securitySchemes": {
                        "bearerAuth": {
                            "type": "http",
                            "scheme": "bearer"
                        }
                    }
                }
            }`,
            expected: true,
        },
        {
            name: "spec without security",
            spec: `{
                "openapi": "3.0.0",
                "info": {"title": "Test", "version": "1.0"}
            }`,
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Write spec to temp file
            tmpFile := filepath.Join(t.TempDir(), "openapi.json")
            err := os.WriteFile(tmpFile, []byte(tt.spec), 0644)
            require.NoError(t, err)

            // Detect security
            hasSecurity, err := detectSecurityFromSpec(tmpFile)
            require.NoError(t, err)
            assert.Equal(t, tt.expected, hasSecurity)
        })
    }
}
```

---

### 2.2 Add Configuration Validation (P1)

**Already covered in section 1.2** ✅

See the `Validate()` function in the updated `config.go`

---

## Phase 3: Medium Priority (P2)

### 3.1 Support Multiple Spec File Formats (P2)

**Problem:** Only supports `openapi.json`, ignores YAML and other names

**Solution:** Support multiple file patterns and formats

#### Implementation

**File:** `internal/config/config.go` (updated)

```go
type Config struct {
    SpecsDir       string   `mapstructure:"specs_dir"`
    OutputDir      string   `mapstructure:"output_dir"`
    TargetServices string   `mapstructure:"target_services"`
    ContinueOnError bool    `mapstructure:"continue_on_error"`

    // SpecFilePatterns defines which files to search for
    // Default: ["openapi.json", "openapi.yaml", "openapi.yml", "swagger.json"]
    SpecFilePatterns []string `mapstructure:"spec_file_patterns"`
}

// Default patterns if not specified in config
var DefaultSpecFilePatterns = []string{
    "openapi.json",
    "openapi.yaml",
    "openapi.yml",
    "swagger.json",
}
```

**File:** `internal/processor/processor.go` (updated)

```go
// findOpenAPISpecs searches for OpenAPI specs in the given directory.
func findOpenAPISpecs(specsDir string, targetServices string, filePatterns []string) ([]string, error) {
    // Use default patterns if none specified
    if len(filePatterns) == 0 {
        filePatterns = config.DefaultSpecFilePatterns
    }

    // Compile service regex for filtering
    serviceRegex, err := compileServiceRegex(targetServices)
    if err != nil {
        return nil, err
    }

    var specs []string

    err = filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
        // Skip directories and errors
        if err != nil || info.IsDir() {
            return nil
        }

        // Check if filename matches any of the patterns
        filename := filepath.Base(path)
        if !matchesAnyPattern(filename, filePatterns) {
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

// matchesAnyPattern checks if filename matches any of the given patterns
func matchesAnyPattern(filename string, patterns []string) bool {
    for _, pattern := range patterns {
        if filename == pattern {
            return true
        }
        // Support glob patterns if needed
        if matched, _ := filepath.Match(pattern, filename); matched {
            return true
        }
    }
    return false
}
```

**File:** `resources/application.yml` (updated)

```yaml
# Spec file patterns to search for (optional)
# Default: ["openapi.json", "openapi.yaml", "openapi.yml", "swagger.json"]
spec_file_patterns:
  - "openapi.json"
  - "openapi.yaml"
  - "openapi.yml"
  - "swagger.json"
```

---

### 3.2 Fix Submodule Behavior (P2)

**Problem:** Always pulls latest, ignores commit pins

**Solution:** Respect submodule pins, add flag for updating

#### Implementation

**File:** `Taskfile.yml` (updated)

```yaml
tasks:
  generate:
    desc: Generate clients from OpenAPI specs
    summary: |
      Generates Go clients from OpenAPI specs in the submodule.
      By default, uses the pinned submodule commit.
      Set UPDATE_SUBMODULE=true to pull latest changes.
    cmds:
      - task: init
      - task: generate-code

  init:
    internal: true
    desc: Initialize git submodules
    cmds:
      # Always initialize submodules first
      - git submodule update --init --recursive

      # Conditionally update to latest if UPDATE_SUBMODULE is set
      - |
        if [ "${UPDATE_SUBMODULE:-false}" = "true" ]; then
          echo "🔄 Updating SDK submodule to latest commit on main branch..."
          cd external/sdk && git checkout main && git pull origin main && cd ../..
          echo "✅ Submodule updated to latest"
          echo "⚠️  Remember to commit the new submodule reference if you want to pin this version"
        else
          echo "📌 Using pinned submodule commit (set UPDATE_SUBMODULE=true to update)"
          git submodule status
        fi
    status:
      - test -f external/sdk/sdk-packages/funding-server-sdk/openapi.json

  update-submodule:
    desc: Update submodule to latest and commit the new reference
    cmds:
      - git submodule update --init --recursive
      - cd external/sdk && git checkout main && git pull origin main && cd ../..
      - git add external/sdk
      - |
        if git diff --cached --quiet; then
          echo "Submodule already up to date"
        else
          git commit -m "chore: update SDK submodule to latest"
          echo "✅ Submodule updated and committed"
        fi
```

**Usage:**

```bash
# Use pinned commit (default, deterministic)
task generate

# Update to latest (for development)
UPDATE_SUBMODULE=true task generate

# Update and commit the new pin
task update-submodule
```

**File:** `.gitlab-ci.yml` (updated)

```yaml
variables:
  GOLANG_VERSION: "1.24"
  # Set to true for scheduled updates, false for manual builds
  UPDATE_SUBMODULE: "true"

update-sdk:
  stage: update
  image: golang:$GOLANG_VERSION
  before_script:
    - apk add --no-cache git task
  script:
    # Generate with updated submodule
    - task generate

    # Stage all changes (generated code AND submodule reference)
    - git add generated/ external/sdk

    # Commit if changes exist
    - |
      if [[ -n "$(git diff --cached --name-only)" ]]; then
        git commit -m "chore: update generated clients and SDK submodule"
        git pull --rebase origin ${CI_COMMIT_REF_NAME}
        git push "https://gitlab-ci-token:${CI_JOB_TOKEN}@${CI_REPOSITORY_URL#*@}" "HEAD:${CI_COMMIT_REF_NAME}"
        echo "Changes committed and pushed successfully"
      else
        echo "No changes to commit"
      fi
  rules:
    - if: '$CI_PIPELINE_SOURCE == "schedule"'
      when: always
      variables:
        UPDATE_SUBMODULE: "true"
    - if: '$CI_PIPELINE_SOURCE == "web"'
      when: manual
      variables:
        UPDATE_SUBMODULE: "false"  # Use pinned version for manual runs
```

**Benefits:**
- ✅ Deterministic builds by default
- ✅ Easy to update when needed
- ✅ Clear commit history showing submodule updates
- ✅ Can reproduce old builds

---

### 3.3 Remove or Implement Preprocessor (P2)

**Problem:** Unused commented-out code for OpenAPI 3.1 → 3.0 conversion

**Decision Point:** Does the project need OpenAPI 3.1 support?

#### Option A: Remove (Recommended if not needed)

```bash
rm internal/preprocessor/preprocessor.go
rmdir internal/preprocessor
```

Update imports in other files if referenced.

#### Option B: Implement (If OpenAPI 3.1 support is needed)

**File:** `go.mod` (add dependency)

```go
require (
    github.com/pb33f/libopenapi v0.15.0 // OpenAPI parser with 3.1 support
)
```

**File:** `internal/preprocessor/preprocessor.go` (reimplemented)

```go
package preprocessor

import (
    "fmt"
    "os"

    "github.com/pb33f/libopenapi"
    "github.com/pb33f/libopenapi/datamodel"
)

// ConvertOpenAPI31To30 converts an OpenAPI 3.1 spec to 3.0.3
func ConvertOpenAPI31To30(specPath string) (string, error) {
    // Read the spec
    specBytes, err := os.ReadFile(specPath)
    if err != nil {
        return "", fmt.Errorf("failed to read spec: %w", err)
    }

    // Parse the document
    doc, err := libopenapi.NewDocument(specBytes)
    if err != nil {
        return "", fmt.Errorf("failed to parse OpenAPI document: %w", err)
    }

    // Check version
    version := doc.GetVersion()
    if version != "3.1.0" && version != "3.1" {
        // Already 3.0.x, return original path
        return specPath, nil
    }

    // Convert to 3.0
    converted, errs := doc.RenderAndReload()
    if len(errs) > 0 {
        return "", fmt.Errorf("conversion errors: %v", errs)
    }

    // Write to temp file
    tmpFile, err := os.CreateTemp("", "openapi-30-*.json")
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }

    if _, err := tmpFile.Write(converted); err != nil {
        tmpFile.Close()
        os.Remove(tmpFile.Name())
        return "", fmt.Errorf("failed to write converted spec: %w", err)
    }

    tmpFile.Close()
    return tmpFile.Name(), nil
}
```

**Decision:** Choose based on actual requirements. If no 3.1 specs exist, remove the code.

---

## Phase 4: Enhancements (P3-P4)

### 4.1 Add Dry-Run Mode (P3)

**File:** `internal/config/config.go`

```go
type Config struct {
    // ... existing fields ...

    // DryRun simulates generation without making changes
    DryRun bool `mapstructure:"dry_run"`
}
```

**File:** `internal/processor/processor.go`

```go
func generateClientForSpec(specPath, serviceName, folderName, outputDir string, dryRun bool) error {
    clientPath := filepath.Join(outputDir, "clients", folderName)

    if dryRun {
        log.Printf("[DRY RUN] Would generate client: %s", clientPath)
        log.Printf("[DRY RUN]   Spec: %s", specPath)
        log.Printf("[DRY RUN]   Service: %s", serviceName)
        return nil
    }

    // Actual generation...
}
```

**Usage:**

```bash
DRY_RUN=true task generate
```

---

## Implementation Guidelines

### Backward Compatibility

- Default values should maintain current behavior
- New config options are optional
- Fail fast mode should be opt-in initially, then become default

### Testing Requirements

Each change must include:
1. Unit tests for new functions
2. Integration tests for end-to-end flows
3. Error path testing
4. Documentation updates

### Migration Path

1. **Week 1:** Implement all P0 fixes with feature flags off by default
2. **Week 2:** Enable P0 fixes by default, monitor for issues
3. **Week 3:** Implement P1 fixes
4. **Week 4:** Implement P2 fixes
5. **Week 5:** Testing and documentation

### Review Checklist

Before merging each change:
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Documentation updated
- [ ] Config schema updated
- [ ] Migration guide updated
- [ ] CI/CD pipeline tested
- [ ] Code review approved

---

## Next Steps

1. Review this plan with team
2. Prioritize fixes based on current pain points
3. See `testing-strategy.md` for comprehensive testing approach
4. See `implementation-roadmap.md` for detailed timeline
