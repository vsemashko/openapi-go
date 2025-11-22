package postprocessor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewInternalClientProcessor(t *testing.T) {
	processor := NewInternalClientProcessor()

	if processor == nil {
		t.Fatal("NewInternalClientProcessor() returned nil")
	}

	if processor.Name() != "InternalClientGenerator" {
		t.Errorf("Name() = %q, want %q", processor.Name(), "InternalClientGenerator")
	}
}

func TestInternalClientProcessorName(t *testing.T) {
	processor := NewInternalClientProcessor()
	name := processor.Name()

	if name != "InternalClientGenerator" {
		t.Errorf("Name() = %q, want %q", name, "InternalClientGenerator")
	}
}

func TestInternalClientProcessorProcess(t *testing.T) {
	tests := []struct {
		name        string
		setupSpec   func(string) ProcessSpec
		wantErr     bool
		errContains string
	}{
		{
			name: "valid spec with security",
			setupSpec: func(tmpDir string) ProcessSpec {
				clientPath := filepath.Join(tmpDir, "client")
				os.MkdirAll(clientPath, 0755)

				specPath := filepath.Join(tmpDir, "spec.json")
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
				os.WriteFile(specPath, []byte(spec), 0644)

				return ProcessSpec{
					ClientPath:  clientPath,
					ServiceName: "testservice",
					SpecPath:    specPath,
					PackageName: "testpkg",
				}
			},
			wantErr: false,
		},
		{
			name: "valid spec without security",
			setupSpec: func(tmpDir string) ProcessSpec {
				clientPath := filepath.Join(tmpDir, "client")
				os.MkdirAll(clientPath, 0755)

				specPath := filepath.Join(tmpDir, "spec.json")
				spec := `{
					"openapi": "3.0.0",
					"info": {"title": "Test", "version": "1.0"},
					"paths": {}
				}`
				os.WriteFile(specPath, []byte(spec), 0644)

				return ProcessSpec{
					ClientPath:  clientPath,
					ServiceName: "testservice",
					SpecPath:    specPath,
					PackageName: "testpkg",
				}
			},
			wantErr: false,
		},
		{
			name: "missing spec file (fallback to file check)",
			setupSpec: func(tmpDir string) ProcessSpec {
				clientPath := filepath.Join(tmpDir, "client")
				os.MkdirAll(clientPath, 0755)

				return ProcessSpec{
					ClientPath:  clientPath,
					ServiceName: "testservice",
					SpecPath:    "/nonexistent/spec.json",
					PackageName: "testpkg",
				}
			},
			wantErr: false, // Should use fallback detection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			spec := tt.setupSpec(tmpDir)

			processor := NewInternalClientProcessor()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := processor.Process(ctx, spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Process() error = %q, should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			// If successful, verify output file was created
			if err == nil {
				outputPath := filepath.Join(spec.ClientPath, "oas_internal_client_gen.go")
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("Expected output file not created: %s", outputPath)
				}

				// Verify file has content
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
				} else if len(content) == 0 {
					t.Error("Output file is empty")
				}
			}
		})
	}
}

func TestInternalClientProcessorDetectSecurity(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		expected bool
		wantErr  bool
	}{
		{
			name: "spec with bearer auth",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"components": {
					"securitySchemes": {
						"bearerAuth": {"type": "http", "scheme": "bearer"}
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
			name:     "invalid spec",
			spec:     `{invalid json}`,
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "spec.json")
			os.WriteFile(tmpFile, []byte(tt.spec), 0644)

			processor := NewInternalClientProcessor()
			hasSecurity, err := processor.detectSecurityFromSpec(tmpFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("detectSecurityFromSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if hasSecurity != tt.expected {
				t.Errorf("detectSecurityFromSpec() = %v, want %v", hasSecurity, tt.expected)
			}
		})
	}
}

func TestInternalClientProcessorDetectSecurityFromFiles(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(string) error
		expected bool
	}{
		{
			name: "security file exists",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "oas_security_gen.go"), []byte("package test"), 0644)
			},
			expected: true,
		},
		{
			name: "security file does not exist",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "oas_client_gen.go"), []byte("package test"), 0644)
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

			processor := NewInternalClientProcessor()
			result := processor.detectSecurityFromGeneratedFiles(tmpDir)

			if result != tt.expected {
				t.Errorf("detectSecurityFromGeneratedFiles() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInternalClientProcessorImplementsInterface(t *testing.T) {
	// Verify InternalClientProcessor implements PostProcessor interface
	var _ PostProcessor = (*InternalClientProcessor)(nil)
}
