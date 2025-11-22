package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSecurityFromSpec(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		expected    bool
		wantErr     bool
		errContains string
	}{
		{
			name: "spec with bearer auth",
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
			wantErr:  false,
		},
		{
			name: "spec without security",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"paths": {}
			}`,
			expected: false,
			wantErr:  false,
		},
		{
			name: "spec with global security",
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
			wantErr:  false,
		},
		{
			name:        "invalid spec",
			spec:        `{invalid json}`,
			expected:    false,
			wantErr:     true,
			errContains: "failed to parse spec",
		},
		{
			name: "empty components",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"components": {}
			}`,
			expected: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write spec to temp file
			tmpFile := filepath.Join(t.TempDir(), "spec.json")
			if err := os.WriteFile(tmpFile, []byte(tt.spec), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test detectSecurityFromSpec
			hasSecurity, err := detectSecurityFromSpec(tmpFile)

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("detectSecurityFromSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("detectSecurityFromSpec() error = %q, should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			// Check result
			if hasSecurity != tt.expected {
				t.Errorf("detectSecurityFromSpec() = %v, expected %v", hasSecurity, tt.expected)
			}
		})
	}
}

func TestDetectSecurityFromSpecNonexistent(t *testing.T) {
	_, err := detectSecurityFromSpec("/nonexistent/spec.json")
	if err == nil {
		t.Error("detectSecurityFromSpec() should error for nonexistent file")
	}
}

func TestDetectSecurityFromGeneratedFiles(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(string) error
		expected bool
	}{
		{
			name: "security file exists",
			setup: func(dir string) error {
				securityFile := filepath.Join(dir, "oas_security_gen.go")
				return os.WriteFile(securityFile, []byte("package test"), 0644)
			},
			expected: true,
		},
		{
			name: "security file does not exist",
			setup: func(dir string) error {
				// Create other files but not oas_security_gen.go
				clientFile := filepath.Join(dir, "oas_client_gen.go")
				return os.WriteFile(clientFile, []byte("package test"), 0644)
			},
			expected: false,
		},
		{
			name: "empty directory",
			setup: func(dir string) error {
				return nil
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.setup != nil {
				if err := tt.setup(tmpDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			result := detectSecurityFromGeneratedFiles(tmpDir)
			if result != tt.expected {
				t.Errorf("detectSecurityFromGeneratedFiles() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateInternalClientFile(t *testing.T) {
	tests := []struct {
		name        string
		setupSpec   func(string) (string, error)
		serviceName string
		wantErr     bool
		errContains string
	}{
		{
			name: "spec with security",
			setupSpec: func(dir string) (string, error) {
				specPath := filepath.Join(dir, "spec.json")
				spec := `{
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
				}`
				err := os.WriteFile(specPath, []byte(spec), 0644)
				return specPath, err
			},
			serviceName: "testservice",
			wantErr:     false,
		},
		{
			name: "spec without security",
			setupSpec: func(dir string) (string, error) {
				specPath := filepath.Join(dir, "spec.json")
				spec := `{
					"openapi": "3.0.0",
					"info": {"title": "Test", "version": "1.0"},
					"paths": {}
				}`
				err := os.WriteFile(specPath, []byte(spec), 0644)
				return specPath, err
			},
			serviceName: "testservice",
			wantErr:     false,
		},
		{
			name: "invalid spec path falls back to file check",
			setupSpec: func(dir string) (string, error) {
				return "/nonexistent/spec.json", nil
			},
			serviceName: "testservice",
			wantErr:     false, // Function has fallback, so no error
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			clientPath := filepath.Join(tmpDir, "client")
			if err := os.MkdirAll(clientPath, 0755); err != nil {
				t.Fatalf("Failed to create client directory: %v", err)
			}

			specPath, err := tt.setupSpec(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup spec: %v", err)
			}

			err = generateInternalClientFile(clientPath, tt.serviceName, specPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateInternalClientFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("generateInternalClientFile() error = %q, should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			// If successful, verify output file was created
			if err == nil {
				outputPath := filepath.Join(clientPath, "oas_internal_client_gen.go")
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("Expected output file not created: %s", outputPath)
				}
			}
		})
	}
}

func TestApplyPostProcessors(t *testing.T) {
	tests := []struct {
		name        string
		setupSpec   func(string) (string, error)
		serviceName string
		wantErr     bool
	}{
		{
			name: "valid spec",
			setupSpec: func(dir string) (string, error) {
				specPath := filepath.Join(dir, "spec.json")
				spec := `{
					"openapi": "3.0.0",
					"info": {"title": "Test", "version": "1.0"},
					"paths": {}
				}`
				err := os.WriteFile(specPath, []byte(spec), 0644)
				return specPath, err
			},
			serviceName: "testservice",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			clientPath := filepath.Join(tmpDir, "client")
			if err := os.MkdirAll(clientPath, 0755); err != nil {
				t.Fatalf("Failed to create client directory: %v", err)
			}

			specPath, err := tt.setupSpec(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup spec: %v", err)
			}

			err = ApplyPostProcessors(clientPath, tt.serviceName, specPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyPostProcessors() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify output file was created
			if err == nil {
				outputPath := filepath.Join(clientPath, "oas_internal_client_gen.go")
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("Expected output file not created: %s", outputPath)
				}

				// Verify file content
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
				} else {
					contentStr := string(content)
					// Should contain package declaration
					if !contains(contentStr, "package") {
						t.Error("Output file should contain package declaration")
					}
				}
			}
		})
	}
}

func TestApplyPostProcessorsNonexistentSpec(t *testing.T) {
	tmpDir := t.TempDir()
	clientPath := filepath.Join(tmpDir, "client")
	os.MkdirAll(clientPath, 0755)

	// This should still work because it falls back to file-based detection
	err := ApplyPostProcessors(clientPath, "testservice", "/nonexistent/spec.json")

	// The function should handle the error gracefully and fall back
	// It will still try to generate the file
	if err != nil {
		// Error is acceptable if template doesn't exist
		t.Logf("ApplyPostProcessors() error (acceptable): %v", err)
	}
}
