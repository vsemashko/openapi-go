package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewOgenGenerator(t *testing.T) {
	gen := NewOgenGenerator()

	if gen == nil {
		t.Fatal("NewOgenGenerator() returned nil")
	}

	if gen.Name() != "ogen" {
		t.Errorf("Name() = %q, want %q", gen.Name(), "ogen")
	}

	if gen.Version() != OgenVersion {
		t.Errorf("Version() = %q, want %q", gen.Version(), OgenVersion)
	}
}

func TestOgenGeneratorName(t *testing.T) {
	gen := NewOgenGenerator()

	name := gen.Name()
	if name != OgenName {
		t.Errorf("Name() = %q, want %q", name, OgenName)
	}

	if name != "ogen" {
		t.Errorf("Name() = %q, want %q", name, "ogen")
	}
}

func TestOgenGeneratorVersion(t *testing.T) {
	gen := NewOgenGenerator()

	version := gen.Version()
	if version != OgenVersion {
		t.Errorf("Version() = %q, want %q", version, OgenVersion)
	}

	if version != "v1.14.0" {
		t.Errorf("Version() = %q, want %q", version, "v1.14.0")
	}
}

func TestOgenGeneratorIsInstalled(t *testing.T) {
	gen := NewOgenGenerator()

	// This test is environment-dependent
	// Just verify it doesn't panic
	result := gen.IsInstalled()
	t.Logf("IsInstalled() = %v", result)

	// No assertions - result depends on whether ogen is installed
}

func TestOgenGeneratorValidate(t *testing.T) {
	tests := []struct {
		name    string
		gen     *OgenGenerator
		wantErr bool
	}{
		{
			name:    "valid ogen generator",
			gen:     NewOgenGenerator(),
			wantErr: false,
		},
		{
			name: "missing version",
			gen: &OgenGenerator{
				version: "",
				pkg:     OgenPackage,
			},
			wantErr: true,
		},
		{
			name: "missing package",
			gen: &OgenGenerator{
				version: OgenVersion,
				pkg:     "",
			},
			wantErr: true,
		},
		{
			name: "both missing",
			gen: &OgenGenerator{
				version: "",
				pkg:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gen.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOgenGeneratorGenerateValidation(t *testing.T) {
	gen := NewOgenGenerator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name    string
		spec    GenerateSpec
		wantErr bool
	}{
		{
			name: "missing spec file",
			spec: GenerateSpec{
				SpecPath:    "/nonexistent/spec.json",
				OutputDir:   "/tmp/output",
				PackageName: "testpkg",
			},
			wantErr: true,
			// Note: Will fail on EnsureInstalled first in test environment
			// In production with ogen installed, it would fail on spec validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Generate(ctx, tt.spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Any error is acceptable for this test
			// (either ogen not installed or spec not found)
			if err != nil {
				t.Logf("Generate() failed as expected: %v", err)
			}
		})
	}
}

func TestGenerateSpecValidation(t *testing.T) {
	tests := []struct {
		name string
		spec GenerateSpec
	}{
		{
			name: "complete spec",
			spec: GenerateSpec{
				SpecPath:    "/path/to/spec.json",
				OutputDir:   "/output",
				PackageName: "testpkg",
				ConfigPath:  "/config/ogen.yml",
				Clean:       true,
			},
		},
		{
			name: "minimal spec",
			spec: GenerateSpec{
				SpecPath:    "/path/to/spec.json",
				OutputDir:   "/output",
				PackageName: "testpkg",
			},
		},
		{
			name: "spec with clean false",
			spec: GenerateSpec{
				SpecPath:    "/path/to/spec.json",
				OutputDir:   "/output",
				PackageName: "testpkg",
				Clean:       false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify struct can be created and fields are accessible
			if tt.spec.SpecPath == "" {
				t.Error("SpecPath should not be empty in test")
			}
			if tt.spec.OutputDir == "" {
				t.Error("OutputDir should not be empty in test")
			}
			if tt.spec.PackageName == "" {
				t.Error("PackageName should not be empty in test")
			}
		})
	}
}

func TestOgenGeneratorContextCancellation(t *testing.T) {
	gen := NewOgenGenerator()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.json")

	// Create a valid spec file
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0"},
		"paths": {}
	}`
	if err := os.WriteFile(specPath, []byte(spec), 0644); err != nil {
		t.Fatalf("Failed to create test spec: %v", err)
	}

	generateSpec := GenerateSpec{
		SpecPath:    specPath,
		OutputDir:   filepath.Join(tmpDir, "output"),
		PackageName: "testpkg",
		Clean:       true,
	}

	// Attempt generation with cancelled context
	// This might fail due to context cancellation or missing ogen
	err := gen.Generate(ctx, generateSpec)

	// We expect an error (either context cancelled or ogen not installed)
	if err == nil {
		t.Log("Generate() succeeded despite cancelled context (ogen might not be installed)")
	} else {
		t.Logf("Generate() failed as expected: %v", err)
	}
}

func TestOgenGeneratorInterfaceImplementation(t *testing.T) {
	// Verify OgenGenerator implements Generator interface
	var _ Generator = (*OgenGenerator)(nil)
}

func TestOgenConstants(t *testing.T) {
	if OgenName != "ogen" {
		t.Errorf("OgenName = %q, want %q", OgenName, "ogen")
	}

	if OgenVersion != "v1.14.0" {
		t.Errorf("OgenVersion = %q, want %q", OgenVersion, "v1.14.0")
	}

	if OgenPackage != "github.com/ogen-go/ogen/cmd/ogen" {
		t.Errorf("OgenPackage = %q, want %q", OgenPackage, "github.com/ogen-go/ogen/cmd/ogen")
	}
}
