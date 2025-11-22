package processor

import (
	"os"
	"path/filepath"
	"testing"
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

		// Abbreviation handling (API, SDK, ID are uppercased)
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

		// Just -server-sdk suffix
		{"test-server-sdk", "test"},

		// Just -sdk suffix
		{"test-sdk", "test"},

		// No suffix
		{"test-service", "testService"},

		// Multiple parts
		{"one-two-three-four-sdk", "oneTwoThreeFour"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeServiceName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeServiceName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
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
	if len(results) != 1 {
		t.Errorf("normalizeServiceName not deterministic, got %d different results: %v",
			len(results), results)
	}

	// Verify expected result
	expected := "funding"
	for result := range results {
		if result != expected {
			t.Errorf("normalizeServiceName(%q) = %q, want %q", input, result, expected)
		}
	}
}

func TestNormalizeServiceNameIsValidGoIdentifier(t *testing.T) {
	inputs := []string{
		"funding-server-sdk",
		"user-api-sdk",
		"my-complex-service-name-sdk",
		"simple",
	}

	for _, input := range inputs {
		result := normalizeServiceName(input)

		// Skip empty results
		if result == "" {
			continue
		}

		// First character should be lowercase letter
		if result[0] < 'a' || result[0] > 'z' {
			t.Errorf("normalizeServiceName(%q) = %q, first char should be lowercase", input, result)
		}

		// Rest should be alphanumeric
		for i, ch := range result {
			if !isAlphaNumeric(ch) {
				t.Errorf("normalizeServiceName(%q) = %q, char at position %d (%c) is not alphanumeric",
					input, result, i, ch)
			}
		}
	}
}

func isAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func TestCompileServiceRegex(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{
			name:    "valid regex",
			pattern: "(service1|service2)",
			wantErr: false,
		},
		{
			name:    "valid complex regex",
			pattern: "^service-[a-z]+$",
			wantErr: false,
		},
		{
			name:    "empty pattern (matches all)",
			pattern: "",
			wantErr: false,
		},
		{
			name:    "match all",
			pattern: ".*",
			wantErr: false,
		},
		{
			name:    "invalid regex",
			pattern: "[invalid(regex",
			wantErr: true,
		},
		{
			name:    "unclosed group",
			pattern: "(unclosed",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := compileServiceRegex(tt.pattern)

			if (err != nil) != tt.wantErr {
				t.Errorf("compileServiceRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && regex == nil {
				t.Error("compileServiceRegex() returned nil regex without error")
			}

			// For valid regex, test that it can be used
			if err == nil {
				// Test with some sample input
				testInput := "test-service"
				matches := regex.MatchString(testInput)
				// Just verify it doesn't panic
				t.Logf("Pattern %q matches %q: %v", tt.pattern, testInput, matches)
			}
		})
	}
}

func TestCompileServiceRegexEmptyPattern(t *testing.T) {
	// Empty pattern should match everything
	regex, err := compileServiceRegex("")
	if err != nil {
		t.Fatalf("compileServiceRegex(\"\") unexpected error: %v", err)
	}

	testCases := []string{
		"anything",
		"test-service",
		"",
		"123",
		"special-chars-!@#",
	}

	for _, tc := range testCases {
		if !regex.MatchString(tc) {
			t.Errorf("Empty pattern should match %q", tc)
		}
	}
}

func TestCleanDirectory(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(string) error
		wantErr bool
	}{
		{
			name: "empty directory",
			setup: func(dir string) error {
				return nil // Directory is already empty
			},
			wantErr: false,
		},
		{
			name: "directory with files",
			setup: func(dir string) error {
				// Create some files
				files := []string{"file1.txt", "file2.go", "file3.json"}
				for _, f := range files {
					if err := os.WriteFile(filepath.Join(dir, f), []byte("test"), 0644); err != nil {
						return err
					}
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "directory with subdirectories",
			setup: func(dir string) error {
				// Create subdirectories with files
				subdir1 := filepath.Join(dir, "subdir1")
				subdir2 := filepath.Join(dir, "subdir2")

				if err := os.MkdirAll(subdir1, 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(subdir2, 0755); err != nil {
					return err
				}

				// Add files to subdirectories
				if err := os.WriteFile(filepath.Join(subdir1, "file1.txt"), []byte("test"), 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(subdir2, "file2.txt"), []byte("test"), 0644); err != nil {
					return err
				}

				return nil
			},
			wantErr: false,
		},
		{
			name: "nested subdirectories",
			setup: func(dir string) error {
				// Create deeply nested structure
				nested := filepath.Join(dir, "level1", "level2", "level3")
				if err := os.MkdirAll(nested, 0755); err != nil {
					return err
				}

				// Add files at various levels
				if err := os.WriteFile(filepath.Join(dir, "level1", "file.txt"), []byte("test"), 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, "level1", "level2", "file.txt"), []byte("test"), 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(nested, "file.txt"), []byte("test"), 0644); err != nil {
					return err
				}

				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			// Setup test directory
			if tt.setup != nil {
				if err := tt.setup(dir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Clean directory
			err := cleanDirectory(dir)

			if (err != nil) != tt.wantErr {
				t.Errorf("cleanDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify directory is empty
				entries, err := os.ReadDir(dir)
				if err != nil {
					t.Fatalf("Failed to read directory after cleaning: %v", err)
				}

				if len(entries) > 0 {
					t.Errorf("cleanDirectory() did not clean directory, %d entries remain", len(entries))
					for _, entry := range entries {
						t.Logf("  Remaining: %s (IsDir: %v)", entry.Name(), entry.IsDir())
					}
				}
			}
		})
	}
}

func TestCleanDirectoryNonexistent(t *testing.T) {
	// Cleaning nonexistent directory should not error (already clean)
	err := cleanDirectory("/nonexistent/directory")
	if err != nil {
		t.Errorf("cleanDirectory() should not error for nonexistent directory, got: %v", err)
	}
}

func TestCleanDirectoryPreservesDirectory(t *testing.T) {
	dir := t.TempDir()

	// Create some files
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644)

	// Clean directory
	err := cleanDirectory(dir)
	if err != nil {
		t.Fatalf("cleanDirectory() error = %v", err)
	}

	// Verify directory still exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("cleanDirectory() removed the directory itself")
	}

	// Verify directory is empty
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(entries) > 0 {
		t.Errorf("Directory not empty after cleaning, has %d entries", len(entries))
	}
}

func TestProcessingResult(t *testing.T) {
	// Test the ProcessingResult struct
	result := &ProcessingResult{
		TotalSpecs:   5,
		SuccessCount: 3,
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
	}

	// Verify fields
	if result.TotalSpecs != 5 {
		t.Errorf("TotalSpecs = %d, want 5", result.TotalSpecs)
	}
	if result.SuccessCount != 3 {
		t.Errorf("SuccessCount = %d, want 3", result.SuccessCount)
	}
	if len(result.FailedSpecs) != 2 {
		t.Errorf("FailedSpecs count = %d, want 2", len(result.FailedSpecs))
	}

	// Verify failure details
	if result.FailedSpecs[0].ServiceName != "service1" {
		t.Errorf("First failure ServiceName = %q, want %q",
			result.FailedSpecs[0].ServiceName, "service1")
	}
}
