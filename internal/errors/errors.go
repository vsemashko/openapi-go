package errors

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCode represents a unique error identifier
type ErrorCode string

// Error codes organized by category
const (
	// File System Errors (FS_*)
	ErrCodeFileNotFound      ErrorCode = "FS_FILE_NOT_FOUND"
	ErrCodeFileAccessDenied  ErrorCode = "FS_ACCESS_DENIED"
	ErrCodeFileIsDirectory   ErrorCode = "FS_IS_DIRECTORY"
	ErrCodeDirectoryNotFound ErrorCode = "FS_DIR_NOT_FOUND"
	ErrCodeFileReadError     ErrorCode = "FS_READ_ERROR"
	ErrCodeFileWriteError    ErrorCode = "FS_WRITE_ERROR"

	// Spec Validation Errors (SPEC_*)
	ErrCodeSpecParseError       ErrorCode = "SPEC_PARSE_ERROR"
	ErrCodeSpecInvalidFormat    ErrorCode = "SPEC_INVALID_FORMAT"
	ErrCodeSpecUnsupportedVer   ErrorCode = "SPEC_UNSUPPORTED_VERSION"
	ErrCodeSpecMissingField     ErrorCode = "SPEC_MISSING_FIELD"
	ErrCodeSpecInvalidField     ErrorCode = "SPEC_INVALID_FIELD"
	ErrCodeSpecMissingOpID      ErrorCode = "SPEC_MISSING_OPERATION_ID"
	ErrCodeSpecDuplicateOpID    ErrorCode = "SPEC_DUPLICATE_OPERATION_ID"
	ErrCodeSpecInvalidRef       ErrorCode = "SPEC_INVALID_REFERENCE"
	ErrCodeSpecMissingSchema    ErrorCode = "SPEC_MISSING_SCHEMA"
	ErrCodeSpecInvalidSecurity  ErrorCode = "SPEC_INVALID_SECURITY"

	// Generation Errors (GEN_*)
	ErrCodeGeneratorNotFound    ErrorCode = "GEN_NOT_FOUND"
	ErrCodeGeneratorFailed      ErrorCode = "GEN_FAILED"
	ErrCodeGeneratorInstall     ErrorCode = "GEN_INSTALL_FAILED"
	ErrCodeGeneratorVersion     ErrorCode = "GEN_VERSION_MISMATCH"
	ErrCodeGeneratorTimeout     ErrorCode = "GEN_TIMEOUT"

	// Post-Processing Errors (POST_*)
	ErrCodePostProcessFailed    ErrorCode = "POST_PROCESS_FAILED"
	ErrCodeFormattingFailed     ErrorCode = "POST_FORMAT_FAILED"
	ErrCodeInternalClientFailed ErrorCode = "POST_INTERNAL_CLIENT_FAILED"

	// Configuration Errors (CFG_*)
	ErrCodeConfigInvalid        ErrorCode = "CFG_INVALID"
	ErrCodeConfigMissing        ErrorCode = "CFG_MISSING"
	ErrCodeConfigLoadFailed     ErrorCode = "CFG_LOAD_FAILED"

	// Cache Errors (CACHE_*)
	ErrCodeCacheReadFailed      ErrorCode = "CACHE_READ_FAILED"
	ErrCodeCacheWriteFailed     ErrorCode = "CACHE_WRITE_FAILED"
	ErrCodeCacheInvalidFormat   ErrorCode = "CACHE_INVALID_FORMAT"

	// Network Errors (NET_*)
	ErrCodeNetworkTimeout       ErrorCode = "NET_TIMEOUT"
	ErrCodeNetworkUnavailable   ErrorCode = "NET_UNAVAILABLE"
)

// ErrorCategory represents the type of error
type ErrorCategory string

const (
	CategoryFileSystem     ErrorCategory = "File System"
	CategoryValidation     ErrorCategory = "Validation"
	CategoryGeneration     ErrorCategory = "Generation"
	CategoryPostProcessing ErrorCategory = "Post-Processing"
	CategoryConfiguration  ErrorCategory = "Configuration"
	CategoryCache          ErrorCategory = "Cache"
	CategoryNetwork        ErrorCategory = "Network"
	CategoryUnknown        ErrorCategory = "Unknown"
)

// Location represents a location in a file
type Location struct {
	File   string
	Line   int
	Column int
}

// String returns a human-readable location string
func (l Location) String() string {
	if l.File == "" {
		return ""
	}
	if l.Line > 0 && l.Column > 0 {
		return fmt.Sprintf("%s:%d:%d", l.File, l.Line, l.Column)
	}
	if l.Line > 0 {
		return fmt.Sprintf("%s:%d", l.File, l.Line)
	}
	return l.File
}

// GenerationError is a structured error with context, location, and suggestions
type GenerationError struct {
	Code       ErrorCode
	Message    string
	Suggestion string
	Location   Location
	Cause      error
	Context    map[string]interface{}
}

// Error implements the error interface
func (e *GenerationError) Error() string {
	var parts []string

	// Add error code
	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Code))
	}

	// Add location
	if loc := e.Location.String(); loc != "" {
		parts = append(parts, loc)
	}

	// Add message
	parts = append(parts, e.Message)

	// Build error string
	errorStr := strings.Join(parts, " ")

	// Add cause if present
	if e.Cause != nil {
		errorStr = fmt.Sprintf("%s: %v", errorStr, e.Cause)
	}

	return errorStr
}

// Unwrap implements the unwrap interface for errors.Is and errors.As
func (e *GenerationError) Unwrap() error {
	return e.Cause
}

// Category returns the error category based on the error code
func (e *GenerationError) Category() ErrorCategory {
	if e.Code == "" {
		return CategoryUnknown
	}

	prefix := strings.Split(string(e.Code), "_")[0]
	switch prefix {
	case "FS":
		return CategoryFileSystem
	case "SPEC":
		return CategoryValidation
	case "GEN":
		return CategoryGeneration
	case "POST":
		return CategoryPostProcessing
	case "CFG":
		return CategoryConfiguration
	case "CACHE":
		return CategoryCache
	case "NET":
		return CategoryNetwork
	default:
		return CategoryUnknown
	}
}

// ErrorList is a collection of errors
type ErrorList struct {
	Errors []*GenerationError
}

// Error implements the error interface
func (el *ErrorList) Error() string {
	if len(el.Errors) == 0 {
		return "no errors"
	}
	if len(el.Errors) == 1 {
		return el.Errors[0].Error()
	}
	return fmt.Sprintf("%d errors occurred", len(el.Errors))
}

// Add adds an error to the list
func (el *ErrorList) Add(err *GenerationError) {
	el.Errors = append(el.Errors, err)
}

// HasErrors returns true if there are any errors
func (el *ErrorList) HasErrors() bool {
	return len(el.Errors) > 0
}

// ToError returns the error list as an error, or nil if empty
func (el *ErrorList) ToError() error {
	if !el.HasErrors() {
		return nil
	}
	return el
}

// GroupByCategory groups errors by category
func (el *ErrorList) GroupByCategory() map[ErrorCategory][]*GenerationError {
	grouped := make(map[ErrorCategory][]*GenerationError)
	for _, err := range el.Errors {
		category := err.Category()
		grouped[category] = append(grouped[category], err)
	}
	return grouped
}

// New creates a new GenerationError
func New(code ErrorCode, message string) *GenerationError {
	return &GenerationError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *GenerationError {
	if err == nil {
		return nil
	}

	// If already a GenerationError, preserve it
	var genErr *GenerationError
	if errors.As(err, &genErr) {
		return &GenerationError{
			Code:       code,
			Message:    message,
			Cause:      genErr,
			Context:    make(map[string]interface{}),
			Location:   genErr.Location,
			Suggestion: genErr.Suggestion,
		}
	}

	return &GenerationError{
		Code:    code,
		Message: message,
		Cause:   err,
		Context: make(map[string]interface{}),
	}
}

// WithLocation adds location information to an error
func (e *GenerationError) WithLocation(file string, line, column int) *GenerationError {
	e.Location = Location{
		File:   file,
		Line:   line,
		Column: column,
	}
	return e
}

// WithSuggestion adds a helpful suggestion to an error
func (e *GenerationError) WithSuggestion(suggestion string) *GenerationError {
	e.Suggestion = suggestion
	return e
}

// WithContext adds context key-value pairs to an error
func (e *GenerationError) WithContext(key string, value interface{}) *GenerationError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Format formats the error with full details including suggestions
func (e *GenerationError) Format() string {
	var lines []string

	// Error header
	header := e.Error()
	lines = append(lines, fmt.Sprintf("❌ %s", header))

	// Add suggestion if present
	if e.Suggestion != "" {
		lines = append(lines, fmt.Sprintf("   💡 Suggestion: %s", e.Suggestion))
	}

	// Add context if present
	if len(e.Context) > 0 {
		for key, value := range e.Context {
			lines = append(lines, fmt.Sprintf("   %s: %v", key, value))
		}
	}

	return strings.Join(lines, "\n")
}

// FormatList formats a list of errors with numbering
func FormatList(errors []*GenerationError) string {
	if len(errors) == 0 {
		return ""
	}

	var lines []string
	for i, err := range errors {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, err.Format()))
	}
	return strings.Join(lines, "\n\n")
}
