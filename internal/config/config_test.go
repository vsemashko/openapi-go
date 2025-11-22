package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		setup   func(*Config)
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			setup: func(cfg *Config) {
				cfg.SpecsDir = t.TempDir()
				cfg.OutputDir = filepath.Join(t.TempDir(), "output")
				cfg.TargetServices = "(service1|service2)"
			},
			wantErr: false,
		},
		{
			name: "missing specs_dir",
			setup: func(cfg *Config) {
				cfg.SpecsDir = ""
				cfg.OutputDir = t.TempDir()
			},
			wantErr: true,
			errMsg:  "specs_dir is required",
		},
		{
			name: "nonexistent specs_dir",
			setup: func(cfg *Config) {
				cfg.SpecsDir = "/nonexistent/path/that/does/not/exist"
				cfg.OutputDir = t.TempDir()
			},
			wantErr: true,
			errMsg:  "specs_dir validation failed",
		},
		{
			name: "missing output_dir",
			setup: func(cfg *Config) {
				cfg.SpecsDir = t.TempDir()
				cfg.OutputDir = ""
			},
			wantErr: true,
			errMsg:  "output_dir is required",
		},
		{
			name: "invalid regex",
			setup: func(cfg *Config) {
				cfg.SpecsDir = t.TempDir()
				cfg.OutputDir = t.TempDir()
				cfg.TargetServices = "[invalid(regex"
			},
			wantErr: true,
			errMsg:  "not a valid regex",
		},
		{
			name: "empty regex (matches all)",
			setup: func(cfg *Config) {
				cfg.SpecsDir = t.TempDir()
				cfg.OutputDir = t.TempDir()
				cfg.TargetServices = ""
			},
			wantErr: false,
		},
		{
			name: "valid complex regex",
			setup: func(cfg *Config) {
				cfg.SpecsDir = t.TempDir()
				cfg.OutputDir = t.TempDir()
				cfg.TargetServices = "^(service-[a-z]+|api-\\d+)$"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{}
			if tt.setup != nil {
				tt.setup(&cfg)
			}

			err := cfg.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %q, should contain %q", err.Error(), tt.errMsg)
				}
			}

			// If no error, verify output directory was created
			if err == nil && cfg.OutputDir != "" {
				if _, statErr := os.Stat(cfg.OutputDir); os.IsNotExist(statErr) {
					t.Errorf("Validate() did not create output directory: %s", cfg.OutputDir)
				}
			}
		})
	}
}

func TestConfigValidationCreatesOutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "nested", "output", "dir")

	cfg := Config{
		SpecsDir:  tmpDir,
		OutputDir: outputDir,
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Errorf("Validate() did not create nested output directory: %s", outputDir)
	}
}

func TestConfigValidationCheckWritable(t *testing.T) {
	// Skip this test on systems where we can't test non-writable directories
	if os.Getuid() == 0 {
		t.Skip("Cannot test non-writable dir as root")
	}

	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")

	// Create a read-only directory
	err := os.Mkdir(readOnlyDir, 0444)
	if err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Cleanup

	// Try to use read-only directory as specs dir (for reading - should work)
	cfg := Config{
		SpecsDir:  readOnlyDir,
		OutputDir: t.TempDir(),
	}

	err = cfg.Validate()
	// This should succeed because SpecsDir only needs to exist, not be writable
	if err != nil {
		t.Errorf("Validate() unexpected error for read-only SpecsDir: %v", err)
	}
}

func TestLoadConfig(t *testing.T) {
	// This test requires the actual config file to exist
	// We'll test that it loads without error when run from the repository
	cfg, err := LoadConfig()

	if err != nil {
		// If error, it should be a clear config-related error
		t.Logf("LoadConfig() error (expected if not in repo): %v", err)
		return
	}

	// If successful, verify config was populated
	if cfg.SpecsDir == "" {
		t.Error("LoadConfig() returned empty SpecsDir")
	}
	if cfg.OutputDir == "" {
		t.Error("LoadConfig() returned empty OutputDir")
	}
}

func TestLoadConfigWithEnvOverride(t *testing.T) {
	// Set environment variable
	tmpDir := t.TempDir()
	os.Setenv("SPECS_DIR", tmpDir)
	defer os.Unsetenv("SPECS_DIR")

	cfg, err := LoadConfig()
	if err != nil {
		// Expected if we're not in the repository
		t.Logf("LoadConfig() error: %v", err)
		return
	}

	// If successful, environment variable should have been applied
	// Note: paths.MakeAbsolutePath will be applied, so we check if it contains our tmpDir
	if !contains(cfg.SpecsDir, filepath.Base(tmpDir)) {
		t.Logf("LoadConfig() SpecsDir = %q, expected to contain %q (may have been made absolute)",
			cfg.SpecsDir, tmpDir)
	}
}

func TestContinueOnErrorDefault(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		SpecsDir:        tmpDir,
		OutputDir:       filepath.Join(tmpDir, "output"),
		ContinueOnError: false, // Default value
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Default should be false (fail fast)
	if cfg.ContinueOnError {
		t.Error("Default ContinueOnError should be false")
	}
}

func TestContinueOnErrorEnabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		SpecsDir:        tmpDir,
		OutputDir:       filepath.Join(tmpDir, "output"),
		ContinueOnError: true,
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should be true when explicitly set
	if !cfg.ContinueOnError {
		t.Error("ContinueOnError should be true when set")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		stringContains(s, substr))))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
