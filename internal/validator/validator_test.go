package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate_ValidJSON(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0",
			"description": "A test API"
		},
		"paths": {}
	}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{Enabled: true})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("Validate() expected valid result")
	}

	if len(result.Errors) > 0 {
		t.Errorf("Validate() expected no errors, got %d", len(result.Errors))
	}

	if result.SpecInfo.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", result.SpecInfo.Format)
	}

	if result.SpecInfo.Version != "3.0.0" {
		t.Errorf("Expected version '3.0.0', got '%s'", result.SpecInfo.Version)
	}

	if result.SpecInfo.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", result.SpecInfo.Title)
	}
}

func TestValidate_ValidYAML(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
  description: A test API
paths: {}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.yaml")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{Enabled: true})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("Validate() expected valid result")
	}

	if result.SpecInfo.Format != "yaml" {
		t.Errorf("Expected format 'yaml', got '%s'", result.SpecInfo.Format)
	}
}

func TestValidate_WithSecurity(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Secure API",
			"version": "1.0.0"
		},
		"components": {
			"securitySchemes": {
				"bearerAuth": {
					"type": "http",
					"scheme": "bearer",
					"bearerFormat": "JWT"
				},
				"apiKey": {
					"type": "apiKey",
					"in": "header",
					"name": "X-API-Key"
				}
			}
		},
		"paths": {}
	}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{Enabled: true})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("Validate() expected valid result")
	}

	if !result.SpecInfo.HasSecurity {
		t.Error("Expected HasSecurity to be true")
	}

	if result.SpecInfo.SchemeCount != 2 {
		t.Errorf("Expected 2 security schemes, got %d", result.SpecInfo.SchemeCount)
	}
}

func TestValidate_MissingFile(t *testing.T) {
	validator := NewValidator(Config{Enabled: true})
	result, err := validator.Validate("/nonexistent/file.json")

	if err == nil {
		t.Error("Validate() expected error for missing file")
	}

	if result.Valid {
		t.Error("Validate() expected invalid result for missing file")
	}

	if len(result.Errors) == 0 {
		t.Error("Validate() expected errors for missing file")
	}

	// Check for FILE_NOT_FOUND error
	foundError := false
	for _, e := range result.Errors {
		if e.Code == "FILE_NOT_FOUND" {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("Expected FILE_NOT_FOUND error")
	}
}

func TestValidate_InvalidJSON(t *testing.T) {
	spec := `{invalid json}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{Enabled: true})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Validate() expected invalid result")
	}

	if len(result.Errors) == 0 {
		t.Error("Validate() expected parse errors")
	}

	// Check for PARSE_ERROR
	foundError := false
	for _, e := range result.Errors {
		if e.Code == "PARSE_ERROR" {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("Expected PARSE_ERROR")
	}
}

func TestValidate_MissingOpenAPIVersion(t *testing.T) {
	spec := `{
		"info": {
			"title": "Test",
			"version": "1.0"
		}
	}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{Enabled: true})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Validate() expected invalid result")
	}

	// Check for MISSING_OPENAPI_VERSION error
	foundError := false
	for _, e := range result.Errors {
		if e.Code == "MISSING_OPENAPI_VERSION" {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("Expected MISSING_OPENAPI_VERSION error")
	}
}

func TestValidate_MissingInfo(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"paths": {}
	}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{Enabled: true})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Validate() expected invalid result")
	}

	// Check for MISSING_INFO error
	foundError := false
	for _, e := range result.Errors {
		if e.Code == "MISSING_INFO" {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("Expected MISSING_INFO error")
	}
}

func TestValidate_Swagger20NotSupported(t *testing.T) {
	spec := `{
		"swagger": "2.0",
		"info": {
			"title": "Old API",
			"version": "1.0"
		}
	}`

	tmpFile := filepath.Join(t.TempDir(), "swagger.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{Enabled: true})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Validate() expected invalid result for Swagger 2.0")
	}
}

func TestValidate_OpenAPI31Warning(t *testing.T) {
	spec := `{
		"openapi": "3.1.0",
		"info": {
			"title": "New API",
			"version": "1.0"
		},
		"paths": {}
	}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{Enabled: true})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Should have warnings about 3.1 support
	foundWarning := false
	for _, w := range result.Warnings {
		if w.Code == "UNSUPPORTED_VERSION" {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("Expected UNSUPPORTED_VERSION warning for OpenAPI 3.1")
	}
}

func TestValidate_StrictMode(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "No Security API",
			"version": "1.0"
		},
		"paths": {}
	}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{
		Enabled:    true,
		StrictMode: true,
	})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Should have warning about no security in strict mode
	foundWarning := false
	for _, w := range result.Warnings {
		if w.Code == "NO_SECURITY" {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("Expected NO_SECURITY warning in strict mode")
	}
}

func TestValidate_CustomRules(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Minimal API",
			"version": "1.0"
		},
		"paths": {}
	}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{
		Enabled:     true,
		CustomRules: []string{"require-description", "require-contact", "require-license"},
	})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Should have warnings for missing description, contact, and license
	expectedWarnings := map[string]bool{
		"MISSING_DESCRIPTION": false,
		"MISSING_CONTACT":     false,
		"MISSING_LICENSE":     false,
	}

	for _, w := range result.Warnings {
		if _, exists := expectedWarnings[w.Code]; exists {
			expectedWarnings[w.Code] = true
		}
	}

	for code, found := range expectedWarnings {
		if !found {
			t.Errorf("Expected warning %s not found", code)
		}
	}
}

func TestValidate_IgnoredRules(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Minimal API",
			"version": "1.0"
		},
		"paths": {}
	}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{
		Enabled:      true,
		StrictMode:   true,
		IgnoredRules: []string{"NO_SECURITY"},
	})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// NO_SECURITY warning should be filtered out
	for _, w := range result.Warnings {
		if w.Code == "NO_SECURITY" {
			t.Error("NO_SECURITY warning should have been ignored")
		}
	}
}

func TestValidate_FailOnWarnings(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "",
			"version": "1.0"
		},
		"paths": {}
	}`

	tmpFile := filepath.Join(t.TempDir(), "openapi.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewValidator(Config{
		Enabled:        true,
		FailOnWarnings: true,
	})
	result, err := validator.Validate(tmpFile)

	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Should be invalid due to FailOnWarnings and empty title warning
	if result.Valid {
		t.Error("Expected invalid result when FailOnWarnings is true and warnings exist")
	}
}

func TestValidateMultiple(t *testing.T) {
	// Create multiple spec files
	spec1 := `{
		"openapi": "3.0.0",
		"info": {"title": "API 1", "version": "1.0"},
		"paths": {}
	}`

	spec2 := `{
		"openapi": "3.0.0",
		"info": {"title": "API 2", "version": "1.0"},
		"paths": {}
	}`

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "api1.json")
	file2 := filepath.Join(tmpDir, "api2.json")

	os.WriteFile(file1, []byte(spec1), 0644)
	os.WriteFile(file2, []byte(spec2), 0644)

	validator := NewValidator(Config{Enabled: true})
	results, err := ValidateMultiple(validator, []string{file1, file2})

	if err != nil {
		t.Errorf("ValidateMultiple() unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for i, result := range results {
		if !result.Valid {
			t.Errorf("Result %d expected to be valid", i)
		}
	}
}

func TestHasErrors(t *testing.T) {
	results := []*ValidationResult{
		{Valid: true, Errors: []ValidationError{}},
		{Valid: false, Errors: []ValidationError{{Code: "TEST_ERROR"}}},
		{Valid: true, Errors: []ValidationError{}},
	}

	if !HasErrors(results) {
		t.Error("HasErrors() should return true when any result is invalid")
	}

	validResults := []*ValidationResult{
		{Valid: true, Errors: []ValidationError{}},
		{Valid: true, Errors: []ValidationError{}},
	}

	if HasErrors(validResults) {
		t.Error("HasErrors() should return false when all results are valid")
	}
}

func TestFormatValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Field: "test", Message: "Test error", Code: "TEST_ERROR"},
		},
		Warnings: []ValidationWarning{
			{Field: "test", Message: "Test warning", Code: "TEST_WARNING"},
		},
		SpecInfo: SpecInfo{
			Path:        "/test/spec.json",
			Format:      "json",
			Version:     "3.0.0",
			Title:       "Test API",
			HasSecurity: true,
			SchemeCount: 1,
			SchemeNames: []string{"bearerAuth"},
		},
	}

	formatted := FormatValidationResult(result)

	// Check that formatted output contains key information
	if !contains(formatted, "/test/spec.json") {
		t.Error("Formatted output should contain spec path")
	}
	if !contains(formatted, "json") {
		t.Error("Formatted output should contain format")
	}
	if !contains(formatted, "3.0.0") {
		t.Error("Formatted output should contain version")
	}
	if !contains(formatted, "Test API") {
		t.Error("Formatted output should contain title")
	}
	if !contains(formatted, "TEST_ERROR") {
		t.Error("Formatted output should contain error code")
	}
	if !contains(formatted, "TEST_WARNING") {
		t.Error("Formatted output should contain warning code")
	}
	if !contains(formatted, "FAILED") {
		t.Error("Formatted output should indicate FAILED")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
