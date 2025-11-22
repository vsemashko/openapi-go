# OpenAPI Go SDK Generator - Comprehensive Testing Strategy

**Version:** 1.0
**Date:** 2025-11-22
**Purpose:** Define testing approach for robust, reliable SDK generation

---

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Pyramid](#test-pyramid)
3. [Unit Tests](#unit-tests)
4. [Integration Tests](#integration-tests)
5. [End-to-End Tests](#end-to-end-tests)
6. [Contract Tests](#contract-tests)
7. [Regression Tests](#regression-tests)
8. [Performance Tests](#performance-tests)
9. [CI/CD Integration](#cicd-integration)
10. [Test Data Management](#test-data-management)
11. [Coverage Goals](#coverage-goals)

---

## Testing Philosophy

### Core Principles

1. **Fast Feedback** - Most tests should run in < 1 second
2. **Deterministic** - Same input always produces same output
3. **Isolated** - Tests don't depend on external services or filesystem state
4. **Comprehensive** - Cover happy paths, error paths, and edge cases
5. **Maintainable** - Tests are easy to understand and update

### Testing Goals

- ✅ **Zero regressions** - Catch breaking changes before merge
- ✅ **Confidence in refactoring** - Change internals without fear
- ✅ **Documentation** - Tests serve as usage examples
- ✅ **Quality gate** - Failing tests block deployment

---

## Test Pyramid

```
                    ┌─────────────────┐
                    │   E2E Tests     │  ~10 tests (5%)
                    │  (Full workflow)│
                    └─────────────────┘
                  ┌─────────────────────┐
                  │ Integration Tests   │  ~30 tests (15%)
                  │ (Component combos)  │
                  └─────────────────────┘
              ┌───────────────────────────┐
              │      Unit Tests           │  ~160 tests (80%)
              │  (Individual functions)   │
              └───────────────────────────┘
```

**Target Distribution:**
- 80% Unit Tests - Fast, focused, isolated
- 15% Integration Tests - Component interactions
- 5% E2E Tests - Full generation workflow

**Current State:** ~0% coverage → **Target: 85% code coverage**

---

## Unit Tests

### 1. Configuration Tests

**File:** `internal/config/config_test.go`

```go
package config_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
)

func TestLoadConfig(t *testing.T) {
    t.Run("loads valid config", func(t *testing.T) {
        // Create temp config file
        tmpDir := t.TempDir()
        configPath := filepath.Join(tmpDir, "application.yml")

        configContent := `
specs_dir: "./specs"
output_dir: "./output"
target_services: "(service1|service2)"
`
        err := os.WriteFile(configPath, []byte(configContent), 0644)
        require.NoError(t, err)

        // Set config path via env var
        os.Setenv("CONFIG_PATH", tmpDir)
        defer os.Unsetenv("CONFIG_PATH")

        cfg, err := config.LoadConfig()
        require.NoError(t, err)

        assert.Equal(t, "./specs", cfg.SpecsDir)
        assert.Equal(t, "./output", cfg.OutputDir)
        assert.Equal(t, "(service1|service2)", cfg.TargetServices)
    })

    t.Run("environment variables override config file", func(t *testing.T) {
        os.Setenv("SPECS_DIR", "/custom/specs")
        defer os.Unsetenv("SPECS_DIR")

        cfg, err := config.LoadConfig()
        require.NoError(t, err)

        assert.Equal(t, "/custom/specs", cfg.SpecsDir)
    })

    t.Run("fails on missing config file", func(t *testing.T) {
        os.Setenv("CONFIG_PATH", "/nonexistent")
        defer os.Unsetenv("CONFIG_PATH")

        _, err := config.LoadConfig()
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "config file")
    })
}

func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  config.Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid config",
            config: config.Config{
                SpecsDir:       t.TempDir(),
                OutputDir:      t.TempDir(),
                TargetServices: "(service.*)",
            },
            wantErr: false,
        },
        {
            name: "missing specs_dir",
            config: config.Config{
                OutputDir: t.TempDir(),
            },
            wantErr: true,
            errMsg:  "specs_dir is required",
        },
        {
            name: "nonexistent specs_dir",
            config: config.Config{
                SpecsDir:  "/nonexistent/path",
                OutputDir: t.TempDir(),
            },
            wantErr: true,
            errMsg:  "specs_dir validation failed",
        },
        {
            name: "invalid regex",
            config: config.Config{
                SpecsDir:       t.TempDir(),
                OutputDir:      t.TempDir(),
                TargetServices: "[invalid(regex",
            },
            wantErr: true,
            errMsg:  "not a valid regex",
        },
        {
            name: "output_dir not writable",
            config: config.Config{
                SpecsDir:  t.TempDir(),
                OutputDir: "/root/readonly",
            },
            wantErr: true,
            errMsg:  "output_dir validation failed",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

---

### 2. Service Name Normalization Tests

**File:** `internal/processor/utils_test.go`

```go
package processor

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestNormalizeServiceName(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        // Basic cases
        {"funding-server-sdk", "funding"},
        {"holidays-server-sdk", "holidays"},
        {"simple-sdk", "simple"},

        // Multiple hyphens
        {"user-management-service-sdk", "userManagementService"},
        {"api-gateway-server-sdk", "APIGateway"},

        // Abbreviation handling
        {"user-api-sdk", "userAPI"},
        {"get-id-service-sdk", "getIDService"},
        {"my-sdk-api-sdk", "mySDKAPI"},

        // Edge cases
        {"single", "single"},
        {"double-word-sdk", "doubleWord"},
        {"", ""},

        // Complex cases
        {"payment-processing-api-server-sdk", "paymentProcessingAPI"},
        {"user-id-verification-service-sdk", "userIDVerificationService"},

        // Special characters (should be handled gracefully)
        {"service_with_underscores-sdk", "serviceWithUnderscores"},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := normalizeServiceName(tt.input)
            assert.Equal(t, tt.expected, result,
                "normalizeServiceName(%q) = %q, want %q", tt.input, result, tt.expected)
        })
    }
}

func TestNormalizeServiceNameConsistency(t *testing.T) {
    // Same input should always produce same output
    input := "funding-server-sdk"

    results := make(map[string]int)
    for i := 0; i < 100; i++ {
        result := normalizeServiceName(input)
        results[result]++
    }

    // Should have exactly one unique result
    assert.Len(t, results, 1, "normalizeServiceName should be deterministic")
}

func TestNormalizeServiceNameIsValidGoIdentifier(t *testing.T) {
    inputs := []string{
        "funding-server-sdk",
        "user-api-sdk",
        "my-complex-service-name-sdk",
    }

    for _, input := range inputs {
        result := normalizeServiceName(input)

        // Check if result is a valid Go identifier
        assert.NotEmpty(t, result, "result should not be empty")

        // First character should be lowercase letter
        assert.True(t, isLowerLetter(rune(result[0])),
            "first character should be lowercase: %s", result)

        // Rest should be alphanumeric
        for _, ch := range result[1:] {
            assert.True(t, isAlphaNumeric(ch),
                "all characters should be alphanumeric: %s", result)
        }
    }
}

func isLowerLetter(r rune) bool {
    return r >= 'a' && r <= 'z'
}

func isAlphaNumeric(r rune) bool {
    return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
```

---

### 3. Spec Discovery Tests

**File:** `internal/processor/processor_test.go`

```go
package processor

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFindOpenAPISpecs(t *testing.T) {
    t.Run("finds matching specs", func(t *testing.T) {
        // Create test directory structure
        tmpDir := t.TempDir()

        // Create matching specs
        createSpec(t, tmpDir, "funding-server-sdk/openapi.json")
        createSpec(t, tmpDir, "holidays-server-sdk/openapi.json")

        // Create non-matching spec (should be ignored)
        createSpec(t, tmpDir, "other-service-sdk/openapi.json")

        // Find specs with filter
        specs, err := findOpenAPISpecs(tmpDir, "(funding|holidays).*", nil)
        require.NoError(t, err)

        assert.Len(t, specs, 2)
    })

    t.Run("supports multiple file formats", func(t *testing.T) {
        tmpDir := t.TempDir()

        createSpec(t, tmpDir, "service1-sdk/openapi.json")
        createSpec(t, tmpDir, "service2-sdk/openapi.yaml")
        createSpec(t, tmpDir, "service3-sdk/openapi.yml")
        createSpec(t, tmpDir, "service4-sdk/swagger.json")

        patterns := []string{"openapi.json", "openapi.yaml", "openapi.yml", "swagger.json"}
        specs, err := findOpenAPISpecs(tmpDir, ".*", patterns)
        require.NoError(t, err)

        assert.Len(t, specs, 4)
    })

    t.Run("returns error on invalid regex", func(t *testing.T) {
        tmpDir := t.TempDir()

        _, err := findOpenAPISpecs(tmpDir, "[invalid(regex", nil)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid")
    })

    t.Run("returns error when no specs found", func(t *testing.T) {
        tmpDir := t.TempDir()

        _, err := findOpenAPISpecs(tmpDir, ".*", nil)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "no OpenAPI specs found")
    })
}

func createSpec(t *testing.T, baseDir, relativePath string) {
    t.Helper()

    fullPath := filepath.Join(baseDir, relativePath)
    dir := filepath.Dir(fullPath)

    err := os.MkdirAll(dir, 0755)
    require.NoError(t, err)

    err = os.WriteFile(fullPath, []byte(`{"openapi":"3.0.0"}`), 0644)
    require.NoError(t, err)
}
```

---

### 4. Security Detection Tests

**File:** `internal/spec/parser_test.go`

```go
package spec

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHasSecurity(t *testing.T) {
    tests := []struct {
        name     string
        spec     string
        expected bool
    }{
        {
            name: "has bearer auth security scheme",
            spec: `{
                "openapi": "3.0.0",
                "info": {"title": "Test", "version": "1.0"},
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
            name: "has global security requirement",
            spec: `{
                "openapi": "3.0.0",
                "info": {"title": "Test", "version": "1.0"},
                "security": [{"apiKey": []}],
                "components": {
                    "securitySchemes": {
                        "apiKey": {
                            "type": "apiKey",
                            "in": "header",
                            "name": "X-API-Key"
                        }
                    }
                }
            }`,
            expected: true,
        },
        {
            name: "no security",
            spec: `{
                "openapi": "3.0.0",
                "info": {"title": "Test", "version": "1.0"},
                "paths": {}
            }`,
            expected: false,
        },
        {
            name: "empty security schemes",
            spec: `{
                "openapi": "3.0.0",
                "info": {"title": "Test", "version": "1.0"},
                "components": {
                    "securitySchemes": {}
                }
            }`,
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Write spec to temp file
            tmpFile := filepath.Join(t.TempDir(), "spec.json")
            err := os.WriteFile(tmpFile, []byte(tt.spec), 0644)
            require.NoError(t, err)

            // Parse spec
            spec, err := ParseSpecFile(tmpFile)
            require.NoError(t, err)

            // Check security
            assert.Equal(t, tt.expected, spec.HasSecurity())
        })
    }
}

func TestParseInvalidSpec(t *testing.T) {
    tmpFile := filepath.Join(t.TempDir(), "invalid.json")
    err := os.WriteFile(tmpFile, []byte(`{invalid json}`), 0644)
    require.NoError(t, err)

    _, err = ParseSpecFile(tmpFile)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to parse")
}

func TestParseNonexistentFile(t *testing.T) {
    _, err := ParseSpecFile("/nonexistent/file.json")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to read")
}
```

---

### 5. Path Utilities Tests

**File:** `internal/paths/paths_test.go`

```go
package paths

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestEnsurePathExists(t *testing.T) {
    t.Run("existing path", func(t *testing.T) {
        tmpFile := filepath.Join(t.TempDir(), "test.txt")
        err := os.WriteFile(tmpFile, []byte("test"), 0644)
        require.NoError(t, err)

        err = EnsurePathExists(tmpFile)
        assert.NoError(t, err)
    })

    t.Run("nonexistent path", func(t *testing.T) {
        err := EnsurePathExists("/nonexistent/path")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "does not exist")
    })
}

func TestEnsureDirectoryWritable(t *testing.T) {
    t.Run("writable directory", func(t *testing.T) {
        tmpDir := t.TempDir()
        err := EnsureDirectoryWritable(tmpDir)
        assert.NoError(t, err)
    })

    t.Run("non-writable directory", func(t *testing.T) {
        // Skip on systems where we can't test this
        if os.Getuid() == 0 {
            t.Skip("Cannot test non-writable dir as root")
        }

        tmpDir := t.TempDir()
        err := os.Chmod(tmpDir, 0444) // Read-only
        require.NoError(t, err)
        defer os.Chmod(tmpDir, 0755)

        err = EnsureDirectoryWritable(tmpDir)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "not writable")
    })
}

func TestFindRepositoryRoot(t *testing.T) {
    // This test assumes it runs within the repository
    root, err := findRepositoryRoot(".")
    require.NoError(t, err)

    // Should contain go.mod
    goModPath := filepath.Join(root, "go.mod")
    assert.FileExists(t, goModPath)
}
```

---

## Integration Tests

### 1. Full Generation Flow Test

**File:** `test/integration/generation_test.go`

```go
package integration_test

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFullGenerationFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup test environment
    tmpDir := t.TempDir()
    specsDir := filepath.Join(tmpDir, "specs")
    outputDir := filepath.Join(tmpDir, "output")

    // Create test spec
    createTestSpec(t, specsDir, "test-service-sdk", testOpenAPISpec)

    // Create config
    configContent := `
specs_dir: ` + specsDir + `
output_dir: ` + outputDir + `
target_services: ".*"
`
    configPath := filepath.Join(tmpDir, "application.yml")
    err := os.WriteFile(configPath, []byte(configContent), 0644)
    require.NoError(t, err)

    // Run generator
    cmd := exec.Command("go", "run", "../../main.go")
    cmd.Env = append(os.Environ(), "CONFIG_PATH="+tmpDir)
    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "Generator failed: %s", string(output))

    // Verify output
    clientDir := filepath.Join(outputDir, "clients", "testservicesdk")
    assert.DirExists(t, clientDir)

    // Check for expected generated files
    expectedFiles := []string{
        "oas_client_gen.go",
        "oas_schemas_gen.go",
        "oas_json_gen.go",
        "oas_internal_client_gen.go",
    }

    for _, file := range expectedFiles {
        filePath := filepath.Join(clientDir, file)
        assert.FileExists(t, filePath, "Expected file not generated: %s", file)
    }

    // Verify generated code compiles
    cmd = exec.Command("go", "build", "./...")
    cmd.Dir = clientDir
    output, err = cmd.CombinedOutput()
    assert.NoError(t, err, "Generated code does not compile: %s", string(output))
}

const testOpenAPISpec = `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test Service",
    "version": "1.0.0"
  },
  "paths": {
    "/health": {
      "get": {
        "operationId": "getHealth",
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": {"type": "string"}
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

func createTestSpec(t *testing.T, baseDir, serviceName, content string) {
    t.Helper()

    serviceDir := filepath.Join(baseDir, serviceName)
    err := os.MkdirAll(serviceDir, 0755)
    require.NoError(t, err)

    specPath := filepath.Join(serviceDir, "openapi.json")
    err = os.WriteFile(specPath, []byte(content), 0644)
    require.NoError(t, err)
}
```

---

### 2. Error Handling Integration Tests

**File:** `test/integration/error_handling_test.go`

```go
package integration_test

import (
    "os/exec"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestInvalidSpecHandling(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    tmpDir := t.TempDir()
    specsDir := filepath.Join(tmpDir, "specs")

    // Create invalid spec
    createTestSpec(t, specsDir, "invalid-sdk", `{invalid json}`)

    // Run generator
    cmd := exec.Command("go", "run", "../../main.go")
    cmd.Env = append(os.Environ(),
        "SPECS_DIR="+specsDir,
        "OUTPUT_DIR="+filepath.Join(tmpDir, "output"),
        "CONTINUE_ON_ERROR=false",
    )

    output, err := cmd.CombinedOutput()

    // Should fail
    assert.Error(t, err)
    assert.Contains(t, string(output), "failed")
}

func TestContinueOnErrorMode(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    tmpDir := t.TempDir()
    specsDir := filepath.Join(tmpDir, "specs")

    // Create one valid and one invalid spec
    createTestSpec(t, specsDir, "valid-sdk", testOpenAPISpec)
    createTestSpec(t, specsDir, "invalid-sdk", `{invalid}`)

    // Run with continue-on-error
    cmd := exec.Command("go", "run", "../../main.go")
    cmd.Env = append(os.Environ(),
        "SPECS_DIR="+specsDir,
        "OUTPUT_DIR="+filepath.Join(tmpDir, "output"),
        "CONTINUE_ON_ERROR=true",
    )

    output, err := cmd.CombinedOutput()

    // Should succeed (partial success)
    assert.NoError(t, err)
    assert.Contains(t, string(output), "Successfully processed 1/2")
}
```

---

## End-to-End Tests

### 1. Real Spec Generation Test

**File:** `test/e2e/real_specs_test.go`

```go
package e2e_test

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGenerateFromRealSpecs(t *testing.T) {
    if os.Getenv("E2E_TEST") != "true" {
        t.Skip("Set E2E_TEST=true to run end-to-end tests")
    }

    // Use actual submodule specs
    cmd := exec.Command("task", "generate")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "Generation failed: %s", string(output))

    // Verify all expected SDKs were generated
    expectedSDKs := []string{
        "fundingsdk",
        "holidayssdk",
    }

    for _, sdk := range expectedSDKs {
        sdkPath := filepath.Join("generated", "clients", sdk)
        assert.DirExists(t, sdkPath, "SDK not generated: %s", sdk)

        // Verify SDK compiles
        cmd := exec.Command("go", "build", "./...")
        cmd.Dir = sdkPath
        output, err := cmd.CombinedOutput()
        assert.NoError(t, err, "SDK %s does not compile: %s", sdk, string(output))
    }
}

func TestGeneratedClientUsage(t *testing.T) {
    if os.Getenv("E2E_TEST") != "true" {
        t.Skip("Set E2E_TEST=true to run end-to-end tests")
    }

    // Test that generated client can be instantiated
    testCode := `
package main

import (
    "fmt"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/generated/clients/holidayssdk"
)

func main() {
    client, err := holidayssdk.NewInternalClient("http://localhost:8080")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Client created: %v\n", client)
}
`

    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "main.go")
    err := os.WriteFile(testFile, []byte(testCode), 0644)
    require.NoError(t, err)

    // Initialize go module
    cmd := exec.Command("go", "mod", "init", "test")
    cmd.Dir = tmpDir
    cmd.Run()

    // Try to build
    cmd = exec.Command("go", "build", ".")
    cmd.Dir = tmpDir
    output, err := cmd.CombinedOutput()
    assert.NoError(t, err, "Client usage test failed: %s", string(output))
}
```

---

## Contract Tests

### OpenAPI Spec Contract Tests

**File:** `test/contract/spec_contract_test.go`

```go
package contract_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/spec"
)

// TestSpecParserContract ensures our spec parser handles
// all required OpenAPI 3.0 features
func TestSpecParserContract(t *testing.T) {
    tests := []struct {
        name     string
        feature  string
        spec     string
        validate func(*testing.T, *spec.OpenAPISpec)
    }{
        {
            name:    "bearer authentication",
            feature: "HTTP Bearer auth security scheme",
            spec: `{
                "openapi": "3.0.0",
                "components": {
                    "securitySchemes": {
                        "bearerAuth": {
                            "type": "http",
                            "scheme": "bearer",
                            "bearerFormat": "JWT"
                        }
                    }
                }
            }`,
            validate: func(t *testing.T, s *spec.OpenAPISpec) {
                assert.True(t, s.HasSecurity())
                schemes := s.GetSecuritySchemes()
                assert.Contains(t, schemes, "bearerAuth")
                assert.Equal(t, "http", schemes["bearerAuth"].Type)
                assert.Equal(t, "bearer", schemes["bearerAuth"].Scheme)
            },
        },
        {
            name:    "api key authentication",
            feature: "API Key security scheme",
            spec: `{
                "openapi": "3.0.0",
                "components": {
                    "securitySchemes": {
                        "apiKey": {
                            "type": "apiKey",
                            "in": "header",
                            "name": "X-API-Key"
                        }
                    }
                }
            }`,
            validate: func(t *testing.T, s *spec.OpenAPISpec) {
                assert.True(t, s.HasSecurity())
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpFile := filepath.Join(t.TempDir(), "spec.json")
            err := os.WriteFile(tmpFile, []byte(tt.spec), 0644)
            require.NoError(t, err)

            spec, err := spec.ParseSpecFile(tmpFile)
            require.NoError(t, err)

            tt.validate(t, spec)
        })
    }
}
```

---

## Regression Tests

### Golden File Tests

**File:** `test/regression/golden_test.go`

```go
package regression_test

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGoldenFiles(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping regression test")
    }

    goldenDir := "testdata/golden"
    tmpDir := t.TempDir()

    // Generate from golden spec
    specPath := filepath.Join(goldenDir, "input", "openapi.json")
    outputDir := filepath.Join(tmpDir, "output")

    generateClient(t, specPath, outputDir)

    // Compare with golden files
    goldenOutput := filepath.Join(goldenDir, "expected")

    compareDirectories(t, goldenOutput, outputDir)
}

func generateClient(t *testing.T, specPath, outputDir string) {
    t.Helper()

    cmd := exec.Command("ogen",
        "--target", outputDir,
        "--package", "testclient",
        "--clean",
        specPath)

    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "Generation failed: %s", string(output))
}

func compareDirectories(t *testing.T, expected, actual string) {
    t.Helper()

    // Walk expected directory
    filepath.Walk(expected, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return err
        }

        // Get relative path
        relPath, _ := filepath.Rel(expected, path)
        actualPath := filepath.Join(actual, relPath)

        // Compare files
        expectedContent, err := os.ReadFile(path)
        require.NoError(t, err)

        actualContent, err := os.ReadFile(actualPath)
        require.NoError(t, err)

        assert.Equal(t, string(expectedContent), string(actualContent),
            "File mismatch: %s", relPath)

        return nil
    })
}

// To update golden files:
// UPDATE_GOLDEN=true go test ./test/regression/...
func TestUpdateGoldenFiles(t *testing.T) {
    if os.Getenv("UPDATE_GOLDEN") != "true" {
        t.Skip("Set UPDATE_GOLDEN=true to update golden files")
    }

    // Generate new output
    // Copy to golden directory
    // This is run manually when we intentionally change output format
}
```

---

## Performance Tests

**File:** `test/performance/benchmark_test.go`

```go
package performance_test

import (
    "testing"

    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/processor"
)

func BenchmarkNormalizeServiceName(b *testing.B) {
    for i := 0; i < b.N; i++ {
        processor.normalizeServiceName("funding-server-sdk")
    }
}

func BenchmarkFindOpenAPISpecs(b *testing.B) {
    // Setup test directory with many specs
    tmpDir := setupLargeSpecDirectory(b)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        processor.findOpenAPISpecs(tmpDir, ".*", nil)
    }
}

func BenchmarkFullGeneration(b *testing.B) {
    if testing.Short() {
        b.Skip("Skipping benchmark in short mode")
    }

    // Benchmark full generation cycle
    for i := 0; i < b.N; i++ {
        // Run full generation
    }
}
```

---

## CI/CD Integration

### GitHub Actions / GitLab CI Configuration

**File:** `.gitlab-ci.yml` (updated)

```yaml
stages:
  - test
  - integration
  - update

variables:
  GOLANG_VERSION: "1.24"

# Unit tests - run on every commit
test:unit:
  stage: test
  image: golang:$GOLANG_VERSION
  script:
    - go test -v -race -coverprofile=coverage.txt -covermode=atomic ./internal/...
    - go tool cover -func=coverage.txt
  coverage: '/total:.*?(\d+.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.txt

# Integration tests - run on merge requests
test:integration:
  stage: integration
  image: golang:$GOLANG_VERSION
  script:
    - go test -v ./test/integration/...
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'

# E2E tests - run on main branch
test:e2e:
  stage: integration
  image: golang:$GOLANG_VERSION
  before_script:
    - apk add --no-cache git task
  script:
    - E2E_TEST=true go test -v ./test/e2e/...
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'

# Existing update job (keep as is)
update-sdk:
  stage: update
  # ... existing configuration ...
```

---

## Test Data Management

### Test Fixtures Structure

```
test/
├── fixtures/
│   ├── specs/
│   │   ├── minimal.json          # Minimal valid OpenAPI spec
│   │   ├── with_security.json    # Spec with security schemes
│   │   ├── complex.json          # Complex spec with many operations
│   │   └── invalid.json          # Invalid spec for error testing
│   ├── configs/
│   │   ├── default.yml
│   │   ├── custom_output.yml
│   │   └── invalid.yml
│   └── golden/
│       ├── input/
│       │   └── openapi.json
│       └── expected/
│           ├── oas_client_gen.go
│           ├── oas_schemas_gen.go
│           └── ...
├── integration/
├── e2e/
└── helpers/
    └── test_helpers.go
```

### Test Helpers

**File:** `test/helpers/test_helpers.go`

```go
package helpers

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/require"
)

// LoadFixture loads a test fixture file
func LoadFixture(t *testing.T, name string) string {
    t.Helper()

    path := filepath.Join("../fixtures", name)
    content, err := os.ReadFile(path)
    require.NoError(t, err, "Failed to load fixture: %s", name)

    return string(content)
}

// CreateTempSpec creates a temporary OpenAPI spec file
func CreateTempSpec(t *testing.T, content string) string {
    t.Helper()

    tmpFile := filepath.Join(t.TempDir(), "openapi.json")
    err := os.WriteFile(tmpFile, []byte(content), 0644)
    require.NoError(t, err)

    return tmpFile
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
    t.Helper()

    _, err := os.Stat(path)
    require.NoError(t, err, "File does not exist: %s", path)
}
```

---

## Coverage Goals

### Target Coverage

| Component | Target Coverage | Priority |
|-----------|----------------|----------|
| `internal/config` | 90% | High |
| `internal/processor` | 85% | High |
| `internal/spec` | 95% | High |
| `internal/paths` | 80% | Medium |
| Overall | 85% | - |

### Coverage Commands

```bash
# Run all tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out | grep total

# Fail if coverage below threshold
go test -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out | \
  grep total | \
  awk '{if ($3+0 < 85) {print "Coverage below 85%"; exit 1}}'
```

---

## Test Execution Guide

### Local Development

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run only unit tests (fast)
go test -short ./...

# Run specific test
go test -v ./internal/config -run TestLoadConfig

# Run with coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### CI/CD

```bash
# Run in CI mode (with proper exit codes)
go test -v -race ./... || exit 1

# Integration tests
go test -v ./test/integration/...

# E2E tests
E2E_TEST=true go test -v ./test/e2e/...
```

---

## Summary

This testing strategy provides:

✅ **Comprehensive coverage** - Unit, integration, E2E, contract, regression
✅ **Fast feedback** - Most tests run in < 1 second
✅ **Confidence** - 85% code coverage target
✅ **CI/CD integration** - Automated testing on every commit
✅ **Regression prevention** - Golden file tests catch unexpected changes
✅ **Performance monitoring** - Benchmarks track performance over time

**Next Steps:**
1. Implement unit tests (Priority P0 functions first)
2. Set up CI/CD pipeline with test stages
3. Create test fixtures and golden files
4. Reach 85% coverage target
5. Add E2E tests for real-world scenarios
