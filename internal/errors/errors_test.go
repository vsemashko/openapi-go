package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestGenerationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *GenerationError
		contains []string
	}{
		{
			name: "simple error with code",
			err: &GenerationError{
				Code:    ErrCodeFileNotFound,
				Message: "file not found",
			},
			contains: []string{"[FS_FILE_NOT_FOUND]", "file not found"},
		},
		{
			name: "error with location",
			err: &GenerationError{
				Code:     ErrCodeSpecParseError,
				Message:  "invalid syntax",
				Location: Location{File: "openapi.yaml", Line: 42, Column: 10},
			},
			contains: []string{"[SPEC_PARSE_ERROR]", "openapi.yaml:42:10", "invalid syntax"},
		},
		{
			name: "error with cause",
			err: &GenerationError{
				Code:    ErrCodeGeneratorFailed,
				Message: "generation failed",
				Cause:   fmt.Errorf("ogen error"),
			},
			contains: []string{"[GEN_FAILED]", "generation failed", "ogen error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Error() = %q, should contain %q", result, substr)
				}
			}
		})
	}
}

func TestGenerationError_Category(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected ErrorCategory
	}{
		{"file system error", ErrCodeFileNotFound, CategoryFileSystem},
		{"spec error", ErrCodeSpecParseError, CategoryValidation},
		{"generation error", ErrCodeGeneratorFailed, CategoryGeneration},
		{"post-processing error", ErrCodePostProcessFailed, CategoryPostProcessing},
		{"config error", ErrCodeConfigInvalid, CategoryConfiguration},
		{"cache error", ErrCodeCacheReadFailed, CategoryCache},
		{"network error", ErrCodeNetworkTimeout, CategoryNetwork},
		{"unknown error", ErrorCode("UNKNOWN"), CategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &GenerationError{Code: tt.code}
			if got := err.Category(); got != tt.expected {
				t.Errorf("Category() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	err := New(ErrCodeFileNotFound, "test file not found")

	if err.Code != ErrCodeFileNotFound {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeFileNotFound)
	}
	if err.Message != "test file not found" {
		t.Errorf("Message = %v, want %v", err.Message, "test file not found")
	}
	if err.Context == nil {
		t.Error("Context should be initialized")
	}
}

func TestWrap(t *testing.T) {
	t.Run("wrap standard error", func(t *testing.T) {
		original := fmt.Errorf("original error")
		wrapped := Wrap(original, ErrCodeGeneratorFailed, "wrapped message")

		if wrapped.Code != ErrCodeGeneratorFailed {
			t.Errorf("Code = %v, want %v", wrapped.Code, ErrCodeGeneratorFailed)
		}
		if wrapped.Message != "wrapped message" {
			t.Errorf("Message = %v, want %v", wrapped.Message, "wrapped message")
		}
		if !errors.Is(wrapped, original) {
			t.Error("Wrapped error should unwrap to original")
		}
	})

	t.Run("wrap GenerationError", func(t *testing.T) {
		original := New(ErrCodeFileNotFound, "file missing").
			WithLocation("test.yaml", 10, 5).
			WithSuggestion("check file path")

		wrapped := Wrap(original, ErrCodeGeneratorFailed, "generation failed")

		if wrapped.Location.File != "test.yaml" {
			t.Error("Location should be preserved from original error")
		}
		if wrapped.Suggestion != "check file path" {
			t.Error("Suggestion should be preserved from original error")
		}
	})

	t.Run("wrap nil error", func(t *testing.T) {
		wrapped := Wrap(nil, ErrCodeGeneratorFailed, "test")
		if wrapped != nil {
			t.Error("Wrapping nil should return nil")
		}
	})
}

func TestGenerationError_WithLocation(t *testing.T) {
	err := New(ErrCodeSpecParseError, "parse failed").
		WithLocation("openapi.yaml", 42, 10)

	if err.Location.File != "openapi.yaml" {
		t.Errorf("File = %v, want %v", err.Location.File, "openapi.yaml")
	}
	if err.Location.Line != 42 {
		t.Errorf("Line = %v, want %v", err.Location.Line, 42)
	}
	if err.Location.Column != 10 {
		t.Errorf("Column = %v, want %v", err.Location.Column, 10)
	}
}

func TestGenerationError_WithSuggestion(t *testing.T) {
	suggestion := "try using valid YAML syntax"
	err := New(ErrCodeSpecParseError, "parse failed").
		WithSuggestion(suggestion)

	if err.Suggestion != suggestion {
		t.Errorf("Suggestion = %v, want %v", err.Suggestion, suggestion)
	}
}

func TestGenerationError_WithContext(t *testing.T) {
	err := New(ErrCodeGeneratorFailed, "generation failed").
		WithContext("spec", "users-api").
		WithContext("service", "usersdk")

	if err.Context["spec"] != "users-api" {
		t.Errorf("Context[spec] = %v, want %v", err.Context["spec"], "users-api")
	}
	if err.Context["service"] != "usersdk" {
		t.Errorf("Context[service] = %v, want %v", err.Context["service"], "usersdk")
	}
}

func TestGenerationError_Format(t *testing.T) {
	err := New(ErrCodeSpecMissingOpID, "operationId is required").
		WithLocation("openapi.yaml", 42, 10).
		WithSuggestion("Add operationId field to the operation").
		WithContext("path", "/users").
		WithContext("method", "GET")

	formatted := err.Format()

	// Check that formatted output contains all elements
	requiredElements := []string{
		"❌",
		"SPEC_MISSING_OPERATION_ID",
		"openapi.yaml:42:10",
		"operationId is required",
		"💡 Suggestion:",
		"Add operationId field",
		"path:",
		"method:",
	}

	for _, element := range requiredElements {
		if !strings.Contains(formatted, element) {
			t.Errorf("Format() should contain %q, got:\n%s", element, formatted)
		}
	}
}

func TestErrorList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		el := &ErrorList{}
		if el.HasErrors() {
			t.Error("Empty list should not have errors")
		}
		if el.ToError() != nil {
			t.Error("Empty list ToError() should return nil")
		}
	})

	t.Run("single error", func(t *testing.T) {
		el := &ErrorList{}
		el.Add(New(ErrCodeFileNotFound, "file not found"))

		if !el.HasErrors() {
			t.Error("List should have errors")
		}
		if el.ToError() == nil {
			t.Error("ToError() should return error")
		}
		if !strings.Contains(el.Error(), "file not found") {
			t.Errorf("Error() = %q, should contain 'file not found'", el.Error())
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		el := &ErrorList{}
		el.Add(New(ErrCodeFileNotFound, "error 1"))
		el.Add(New(ErrCodeSpecParseError, "error 2"))
		el.Add(New(ErrCodeGeneratorFailed, "error 3"))

		if !el.HasErrors() {
			t.Error("List should have errors")
		}
		if len(el.Errors) != 3 {
			t.Errorf("Errors count = %d, want 3", len(el.Errors))
		}
		if !strings.Contains(el.Error(), "3 errors") {
			t.Errorf("Error() = %q, should mention '3 errors'", el.Error())
		}
	})
}

func TestErrorList_GroupByCategory(t *testing.T) {
	el := &ErrorList{}
	el.Add(New(ErrCodeFileNotFound, "file error 1"))
	el.Add(New(ErrCodeFileAccessDenied, "file error 2"))
	el.Add(New(ErrCodeSpecParseError, "spec error 1"))
	el.Add(New(ErrCodeGeneratorFailed, "gen error 1"))

	grouped := el.GroupByCategory()

	if len(grouped[CategoryFileSystem]) != 2 {
		t.Errorf("FileSystem errors = %d, want 2", len(grouped[CategoryFileSystem]))
	}
	if len(grouped[CategoryValidation]) != 1 {
		t.Errorf("Validation errors = %d, want 1", len(grouped[CategoryValidation]))
	}
	if len(grouped[CategoryGeneration]) != 1 {
		t.Errorf("Generation errors = %d, want 1", len(grouped[CategoryGeneration]))
	}
}

func TestLocation_String(t *testing.T) {
	tests := []struct {
		name     string
		location Location
		expected string
	}{
		{
			name:     "empty location",
			location: Location{},
			expected: "",
		},
		{
			name:     "file only",
			location: Location{File: "test.yaml"},
			expected: "test.yaml",
		},
		{
			name:     "file and line",
			location: Location{File: "test.yaml", Line: 42},
			expected: "test.yaml:42",
		},
		{
			name:     "file, line, and column",
			location: Location{File: "test.yaml", Line: 42, Column: 10},
			expected: "test.yaml:42:10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.location.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFormatList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		result := FormatList([]*GenerationError{})
		if result != "" {
			t.Errorf("FormatList([]) should return empty string, got %q", result)
		}
	})

	t.Run("single error", func(t *testing.T) {
		errors := []*GenerationError{
			New(ErrCodeFileNotFound, "file not found"),
		}
		result := FormatList(errors)
		if !strings.Contains(result, "1.") {
			t.Error("Should contain numbering")
		}
		if !strings.Contains(result, "file not found") {
			t.Error("Should contain error message")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		errors := []*GenerationError{
			New(ErrCodeFileNotFound, "error 1"),
			New(ErrCodeSpecParseError, "error 2"),
		}
		result := FormatList(errors)
		if !strings.Contains(result, "1.") || !strings.Contains(result, "2.") {
			t.Error("Should contain numbering for all errors")
		}
	})
}
