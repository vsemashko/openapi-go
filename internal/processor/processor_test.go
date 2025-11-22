package processor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
)

func TestFindOpenAPISpecs(t *testing.T) {
	tests := []struct {
		name             string
		setupSpecs       func(string) error
		targetServices   string
		specFilePatterns []string
		expectedCount    int
		wantErr          bool
		errContains      string
	}{
		{
			name: "find all specs (empty filter)",
			setupSpecs: func(dir string) error {
				// Create multiple service directories with openapi.json
				services := []string{"funding-server-sdk", "holidays-server-sdk", "auth-service-sdk"}
				for _, svc := range services {
					svcDir := filepath.Join(dir, svc)
					if err := os.MkdirAll(svcDir, 0755); err != nil {
						return err
					}
					specPath := filepath.Join(svcDir, "openapi.json")
					if err := os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
						return err
					}
				}
				return nil
			},
			targetServices: "",
			expectedCount:  3,
			wantErr:        false,
		},
		{
			name: "filter specific services",
			setupSpecs: func(dir string) error {
				services := []string{"funding-server-sdk", "holidays-server-sdk", "auth-service-sdk"}
				for _, svc := range services {
					svcDir := filepath.Join(dir, svc)
					if err := os.MkdirAll(svcDir, 0755); err != nil {
						return err
					}
					specPath := filepath.Join(svcDir, "openapi.json")
					if err := os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
						return err
					}
				}
				return nil
			},
			targetServices: "(funding-server-sdk|holidays-server-sdk)",
			expectedCount:  2,
			wantErr:        false,
		},
		{
			name: "no specs found",
			setupSpecs: func(dir string) error {
				// Create empty directory
				return nil
			},
			targetServices: "",
			expectedCount:  0,
			wantErr:        true,
			errContains:    "no OpenAPI specs found",
		},
		{
			name: "nested directory structure",
			setupSpecs: func(dir string) error {
				// Create nested structure
				svcDir := filepath.Join(dir, "services", "funding-server-sdk")
				if err := os.MkdirAll(svcDir, 0755); err != nil {
					return err
				}
				specPath := filepath.Join(svcDir, "openapi.json")
				return os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0644)
			},
			targetServices: "",
			expectedCount:  1,
			wantErr:        false,
		},
		{
			name: "ignore non-openapi.json files",
			setupSpecs: func(dir string) error {
				svcDir := filepath.Join(dir, "funding-server-sdk")
				if err := os.MkdirAll(svcDir, 0755); err != nil {
					return err
				}
				// Create openapi.json (should be found)
				if err := os.WriteFile(filepath.Join(svcDir, "openapi.json"), []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
					return err
				}
				// Create other files (should be ignored)
				if err := os.WriteFile(filepath.Join(svcDir, "spec.json"), []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(svcDir, "README.md"), []byte("readme"), 0644); err != nil {
					return err
				}
				return nil
			},
			targetServices: "",
			expectedCount:  1,
			wantErr:        false,
		},
		{
			name: "find YAML specs",
			setupSpecs: func(dir string) error {
				// Create service with openapi.yaml
				svcDir := filepath.Join(dir, "yaml-service-sdk")
				if err := os.MkdirAll(svcDir, 0755); err != nil {
					return err
				}
				yamlContent := `openapi: 3.0.0
info:
  title: YAML API
  version: 1.0.0
paths: {}`
				return os.WriteFile(filepath.Join(svcDir, "openapi.yaml"), []byte(yamlContent), 0644)
			},
			targetServices:   "",
			specFilePatterns: []string{"openapi.yaml"},
			expectedCount:    1,
			wantErr:          false,
		},
		{
			name: "find YML specs",
			setupSpecs: func(dir string) error {
				// Create service with openapi.yml
				svcDir := filepath.Join(dir, "yml-service-sdk")
				if err := os.MkdirAll(svcDir, 0755); err != nil {
					return err
				}
				ymlContent := `openapi: 3.0.0
info:
  title: YML API
  version: 1.0.0
paths: {}`
				return os.WriteFile(filepath.Join(svcDir, "openapi.yml"), []byte(ymlContent), 0644)
			},
			targetServices:   "",
			specFilePatterns: []string{"openapi.yml"},
			expectedCount:    1,
			wantErr:          false,
		},
		{
			name: "find mixed JSON and YAML specs",
			setupSpecs: func(dir string) error {
				// Create JSON service
				jsonSvcDir := filepath.Join(dir, "json-service-sdk")
				if err := os.MkdirAll(jsonSvcDir, 0755); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(jsonSvcDir, "openapi.json"), []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
					return err
				}

				// Create YAML service
				yamlSvcDir := filepath.Join(dir, "yaml-service-sdk")
				if err := os.MkdirAll(yamlSvcDir, 0755); err != nil {
					return err
				}
				yamlContent := `openapi: 3.0.0
info:
  title: YAML API
  version: 1.0.0
paths: {}`
				if err := os.WriteFile(filepath.Join(yamlSvcDir, "openapi.yaml"), []byte(yamlContent), 0644); err != nil {
					return err
				}

				// Create YML service
				ymlSvcDir := filepath.Join(dir, "yml-service-sdk")
				if err := os.MkdirAll(ymlSvcDir, 0755); err != nil {
					return err
				}
				ymlContent := `openapi: 3.0.0
info:
  title: YML API
  version: 1.0.0
paths: {}`
				return os.WriteFile(filepath.Join(ymlSvcDir, "openapi.yml"), []byte(ymlContent), 0644)
			},
			targetServices:   "",
			specFilePatterns: []string{"openapi.json", "openapi.yaml", "openapi.yml"},
			expectedCount:    3,
			wantErr:          false,
		},
		{
			name: "YAML patterns only ignore JSON files",
			setupSpecs: func(dir string) error {
				// Create services with both JSON and YAML
				jsonSvcDir := filepath.Join(dir, "json-service-sdk")
				if err := os.MkdirAll(jsonSvcDir, 0755); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(jsonSvcDir, "openapi.json"), []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
					return err
				}

				yamlSvcDir := filepath.Join(dir, "yaml-service-sdk")
				if err := os.MkdirAll(yamlSvcDir, 0755); err != nil {
					return err
				}
				yamlContent := `openapi: 3.0.0
info:
  title: YAML API
  version: 1.0.0
paths: {}`
				return os.WriteFile(filepath.Join(yamlSvcDir, "openapi.yaml"), []byte(yamlContent), 0644)
			},
			targetServices:   "",
			specFilePatterns: []string{"openapi.yaml", "openapi.yml"}, // only YAML patterns
			expectedCount:    1,                                         // should find only YAML, not JSON
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Setup test specs
			if tt.setupSpecs != nil {
				if err := tt.setupSpecs(tmpDir); err != nil {
					t.Fatalf("Failed to setup specs: %v", err)
				}
			}

			// Run findOpenAPISpecs
			patterns := tt.specFilePatterns
			if patterns == nil {
				patterns = []string{"openapi.json"} // default for existing tests
			}
			specs, err := findOpenAPISpecs(tmpDir, tt.targetServices, patterns)

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("findOpenAPISpecs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("findOpenAPISpecs() error = %q, should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			// Check spec count
			if len(specs) != tt.expectedCount {
				t.Errorf("findOpenAPISpecs() found %d specs, expected %d", len(specs), tt.expectedCount)
			}

			// Verify all found specs exist and match one of the expected patterns
			for _, specPath := range specs {
				filename := filepath.Base(specPath)
				validName := false
				for _, pattern := range patterns {
					if filename == pattern {
						validName = true
						break
					}
				}
				if !validName {
					t.Errorf("Expected spec file to match patterns %v, got %s", patterns, filename)
				}
				if _, err := os.Stat(specPath); os.IsNotExist(err) {
					t.Errorf("Spec file does not exist: %s", specPath)
				}
			}
		})
	}
}

func TestGenerateClients(t *testing.T) {
	tests := []struct {
		name            string
		setupSpecs      func(string) ([]string, error)
		continueOnError bool
		expectedSuccess int
		expectedFailed  int
		wantErr         bool
	}{
		{
			name: "empty specs list",
			setupSpecs: func(dir string) ([]string, error) {
				return []string{}, nil
			},
			continueOnError: false,
			expectedSuccess: 0,
			expectedFailed:  0,
			wantErr:         false,
		},
		{
			name: "valid specs with continue on error",
			setupSpecs: func(dir string) ([]string, error) {
				// Create a valid spec
				svcDir := filepath.Join(dir, "funding-server-sdk")
				if err := os.MkdirAll(svcDir, 0755); err != nil {
					return nil, err
				}
				specPath := filepath.Join(svcDir, "openapi.json")
				validSpec := `{
					"openapi": "3.0.0",
					"info": {"title": "Test", "version": "1.0"},
					"paths": {}
				}`
				if err := os.WriteFile(specPath, []byte(validSpec), 0644); err != nil {
					return nil, err
				}
				return []string{specPath}, nil
			},
			continueOnError: true,
			expectedSuccess: 0, // Will fail because ogen won't actually run successfully
			expectedFailed:  1,
			wantErr:         false, // continue-on-error enabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputDir := filepath.Join(tmpDir, "output")

			// Setup specs
			specs, err := tt.setupSpecs(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup specs: %v", err)
			}

			// Run generateClients with context
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := generateClients(ctx, specs, outputDir, tt.continueOnError, 4, nil)

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("generateClients() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify result structure
			if result == nil {
				t.Fatal("generateClients() returned nil result")
			}

			// Check totals
			if result.TotalSpecs != len(specs) {
				t.Errorf("TotalSpecs = %d, expected %d", result.TotalSpecs, len(specs))
			}

			// For empty specs, just verify structure
			if len(specs) == 0 {
				if result.SuccessCount != 0 {
					t.Errorf("SuccessCount = %d, expected 0 for empty specs", result.SuccessCount)
				}
				if len(result.FailedSpecs) != 0 {
					t.Errorf("FailedSpecs = %d, expected 0 for empty specs", len(result.FailedSpecs))
				}
			}
		})
	}
}

func TestLogProcessingResult(t *testing.T) {
	tests := []struct {
		name   string
		result *ProcessingResult
	}{
		{
			name: "all successful",
			result: &ProcessingResult{
				TotalSpecs:   3,
				SuccessCount: 3,
				FailedSpecs:  []SpecFailure{},
			},
		},
		{
			name: "some failures",
			result: &ProcessingResult{
				TotalSpecs:   3,
				SuccessCount: 1,
				FailedSpecs: []SpecFailure{
					{
						SpecPath:    "/path/to/spec1.json",
						ServiceName: "service1",
						Error:       os.ErrNotExist,
					},
					{
						SpecPath:    "/path/to/spec2.json",
						ServiceName: "service2",
						Error:       os.ErrPermission,
					},
				},
			},
		},
		{
			name: "all failures",
			result: &ProcessingResult{
				TotalSpecs:   2,
				SuccessCount: 0,
				FailedSpecs: []SpecFailure{
					{
						SpecPath:    "/path/to/spec1.json",
						ServiceName: "service1",
						Error:       os.ErrNotExist,
					},
					{
						SpecPath:    "/path/to/spec2.json",
						ServiceName: "service2",
						Error:       os.ErrPermission,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function only logs, so we just verify it doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("logProcessingResult() panicked: %v", r)
				}
			}()

			logProcessingResult(tt.result)
		})
	}
}

func TestProcessOpenAPISpecsValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(string) config.Config
		wantErr     bool
		errContains string
	}{
		{
			name: "missing specs directory",
			setupConfig: func(tmpDir string) config.Config {
				return config.Config{
					SpecsDir:  "/nonexistent/directory",
					OutputDir: tmpDir,
				}
			},
			wantErr:     true,
			errContains: "no OpenAPI specs found",
		},
		{
			name: "valid empty directory",
			setupConfig: func(tmpDir string) config.Config {
				specsDir := filepath.Join(tmpDir, "specs")
				os.MkdirAll(specsDir, 0755)
				return config.Config{
					SpecsDir:  specsDir,
					OutputDir: filepath.Join(tmpDir, "output"),
				}
			},
			wantErr:     true,
			errContains: "no OpenAPI specs found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := tt.setupConfig(tmpDir)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err := ProcessOpenAPISpecs(ctx, cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessOpenAPISpecs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("ProcessOpenAPISpecs() error = %q, should contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestGeneratorIsInstalled(t *testing.T) {
	// This test just verifies the generator check doesn't panic
	// Actual result depends on whether the generator is installed in test environment
	result := defaultGenerator.IsInstalled()
	t.Logf("defaultGenerator.IsInstalled() = %v (generator: %s)", result, defaultGenerator.Name())

	// No assertions - this is environment-dependent
	// The function should at least not panic
}

// Helper function to check string contains substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
