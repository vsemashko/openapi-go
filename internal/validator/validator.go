package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/spec"
)

// Validator validates OpenAPI specifications
type Validator interface {
	Validate(specPath string) (*ValidationResult, error)
}

// ValidationResult contains the results of validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationWarning
	SpecInfo SpecInfo
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string
	Message string
	Code    string
}

// SpecInfo contains information about the spec
type SpecInfo struct {
	Path         string
	Format       string // json or yaml
	Version      string // OpenAPI version
	Title        string
	HasSecurity  bool
	SchemeCount  int
	SchemeNames  []string
}

// Config configures the validator behavior
type Config struct {
	Enabled        bool     `yaml:"enabled"`
	FailOnWarnings bool     `yaml:"fail_on_warnings"`
	CustomRules    []string `yaml:"custom_rules"`
	IgnoredRules   []string `yaml:"ignored_rules"`
	StrictMode     bool     `yaml:"strict_mode"`
}

// DefaultValidator is the standard OpenAPI validator
type DefaultValidator struct {
	config Config
}

// NewValidator creates a new validator with the given configuration
func NewValidator(config Config) *DefaultValidator {
	return &DefaultValidator{
		config: config,
	}
}

// Validate validates an OpenAPI specification file
func (v *DefaultValidator) Validate(specPath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		SpecInfo: SpecInfo{
			Path: specPath,
		},
	}

	// 1. Check file exists and is readable
	if err := v.validateFileExists(specPath, result); err != nil {
		return result, err
	}

	// 2. Detect format
	v.detectFormat(specPath, result)

	// 3. Parse spec
	parsedSpec, err := spec.ParseSpecFile(specPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "spec",
			Message: fmt.Sprintf("Failed to parse spec: %v", err),
			Code:    "PARSE_ERROR",
		})
		return result, nil
	}

	// 4. Validate OpenAPI version
	v.validateOpenAPIVersion(parsedSpec, result)

	// 5. Validate info section
	v.validateInfo(parsedSpec, result)

	// 6. Extract security information
	v.extractSecurityInfo(parsedSpec, result)

	// 7. Run custom rules if configured
	v.applyCustomRules(specPath, parsedSpec, result)

	// 8. Apply ignored rules
	v.filterIgnoredRules(result)

	// Determine final validity
	result.Valid = len(result.Errors) == 0
	if v.config.FailOnWarnings && len(result.Warnings) > 0 {
		result.Valid = false
	}

	return result, nil
}

// validateFileExists checks if the file exists and is readable
func (v *DefaultValidator) validateFileExists(specPath string, result *ValidationResult) error {
	info, err := os.Stat(specPath)
	if os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "file",
			Message: fmt.Sprintf("Spec file does not exist: %s", specPath),
			Code:    "FILE_NOT_FOUND",
		})
		return fmt.Errorf("spec file not found: %s", specPath)
	}
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "file",
			Message: fmt.Sprintf("Cannot access spec file: %v", err),
			Code:    "FILE_ACCESS_ERROR",
		})
		return err
	}

	if info.IsDir() {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "file",
			Message: "Path is a directory, not a file",
			Code:    "NOT_A_FILE",
		})
		return fmt.Errorf("path is a directory: %s", specPath)
	}

	return nil
}

// detectFormat detects the spec format (JSON or YAML)
func (v *DefaultValidator) detectFormat(specPath string, result *ValidationResult) {
	ext := strings.ToLower(filepath.Ext(specPath))
	switch ext {
	case ".json":
		result.SpecInfo.Format = "json"
	case ".yaml", ".yml":
		result.SpecInfo.Format = "yaml"
	default:
		result.SpecInfo.Format = "unknown"
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "format",
			Message: fmt.Sprintf("Unknown file extension: %s. Will attempt auto-detection.", ext),
			Code:    "UNKNOWN_FORMAT",
		})
	}
}

// validateOpenAPIVersion validates the OpenAPI version
func (v *DefaultValidator) validateOpenAPIVersion(parsedSpec *spec.OpenAPISpec, result *ValidationResult) {
	result.SpecInfo.Version = parsedSpec.OpenAPI

	if parsedSpec.OpenAPI == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "openapi",
			Message: "Missing required 'openapi' field",
			Code:    "MISSING_OPENAPI_VERSION",
		})
		return
	}

	// Check version format (should be semantic version)
	versionPattern := regexp.MustCompile(`^3\.\d+\.\d+$`)
	if !versionPattern.MatchString(parsedSpec.OpenAPI) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "openapi",
			Message: fmt.Sprintf("OpenAPI version format may be invalid: %s", parsedSpec.OpenAPI),
			Code:    "INVALID_VERSION_FORMAT",
		})
	}

	// Check if version is supported
	if strings.HasPrefix(parsedSpec.OpenAPI, "3.0") {
		// OpenAPI 3.0.x - fully supported
		result.SpecInfo.Version = parsedSpec.OpenAPI
	} else if strings.HasPrefix(parsedSpec.OpenAPI, "3.1") {
		// OpenAPI 3.1.x - not fully supported yet
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "openapi",
			Message: "OpenAPI 3.1 is not fully supported. Some features may not work correctly.",
			Code:    "UNSUPPORTED_VERSION",
		})
	} else if strings.HasPrefix(parsedSpec.OpenAPI, "2.") {
		// Swagger 2.0 - not supported
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "openapi",
			Message: "Swagger 2.0 is not supported. Please use OpenAPI 3.0 or convert your spec.",
			Code:    "UNSUPPORTED_VERSION",
		})
	} else {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "openapi",
			Message: fmt.Sprintf("Unknown OpenAPI version: %s", parsedSpec.OpenAPI),
			Code:    "UNKNOWN_VERSION",
		})
	}
}

// validateInfo validates the info section
func (v *DefaultValidator) validateInfo(parsedSpec *spec.OpenAPISpec, result *ValidationResult) {
	if parsedSpec.Info == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "info",
			Message: "Missing required 'info' section",
			Code:    "MISSING_INFO",
		})
		return
	}

	// Extract title
	if title, ok := parsedSpec.Info["title"].(string); ok {
		result.SpecInfo.Title = title
		if title == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "info.title",
				Message: "Title is empty",
				Code:    "EMPTY_TITLE",
			})
		}
	} else {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "info.title",
			Message: "Missing required 'title' field in info section",
			Code:    "MISSING_TITLE",
		})
	}

	// Extract version
	if version, ok := parsedSpec.Info["version"].(string); ok {
		if version == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "info.version",
				Message: "Version is empty",
				Code:    "EMPTY_VERSION",
			})
		}
	} else {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "info.version",
			Message: "Missing required 'version' field in info section",
			Code:    "MISSING_VERSION",
		})
	}
}

// extractSecurityInfo extracts security information from the spec
func (v *DefaultValidator) extractSecurityInfo(parsedSpec *spec.OpenAPISpec, result *ValidationResult) {
	result.SpecInfo.HasSecurity = parsedSpec.HasSecurity()

	schemes := parsedSpec.GetSecuritySchemes()
	if schemes != nil {
		result.SpecInfo.SchemeCount = len(schemes)
		result.SpecInfo.SchemeNames = make([]string, 0, len(schemes))
		for name := range schemes {
			result.SpecInfo.SchemeNames = append(result.SpecInfo.SchemeNames, name)
		}
	}

	// Warn if no security is defined (in strict mode)
	if v.config.StrictMode && !result.SpecInfo.HasSecurity {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "security",
			Message: "No security schemes defined. Consider adding authentication.",
			Code:    "NO_SECURITY",
		})
	}
}

// applyCustomRules applies custom validation rules
func (v *DefaultValidator) applyCustomRules(specPath string, parsedSpec *spec.OpenAPISpec, result *ValidationResult) {
	for _, rule := range v.config.CustomRules {
		switch rule {
		case "require-description":
			if desc, ok := parsedSpec.Info["description"].(string); !ok || desc == "" {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:   "info.description",
					Message: "Description is recommended but missing",
					Code:    "MISSING_DESCRIPTION",
				})
			}
		case "require-contact":
			if _, ok := parsedSpec.Info["contact"]; !ok {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:   "info.contact",
					Message: "Contact information is recommended but missing",
					Code:    "MISSING_CONTACT",
				})
			}
		case "require-license":
			if _, ok := parsedSpec.Info["license"]; !ok {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:   "info.license",
					Message: "License information is recommended but missing",
					Code:    "MISSING_LICENSE",
				})
			}
		}
	}
}

// filterIgnoredRules removes errors/warnings for ignored rules
func (v *DefaultValidator) filterIgnoredRules(result *ValidationResult) {
	if len(v.config.IgnoredRules) == 0 {
		return
	}

	ignoredSet := make(map[string]bool)
	for _, rule := range v.config.IgnoredRules {
		ignoredSet[rule] = true
	}

	// Filter errors
	filteredErrors := []ValidationError{}
	for _, err := range result.Errors {
		if !ignoredSet[err.Code] {
			filteredErrors = append(filteredErrors, err)
		}
	}
	result.Errors = filteredErrors

	// Filter warnings
	filteredWarnings := []ValidationWarning{}
	for _, warn := range result.Warnings {
		if !ignoredSet[warn.Code] {
			filteredWarnings = append(filteredWarnings, warn)
		}
	}
	result.Warnings = filteredWarnings
}

// ValidateMultiple validates multiple spec files
func ValidateMultiple(validator Validator, specPaths []string) ([]*ValidationResult, error) {
	results := make([]*ValidationResult, 0, len(specPaths))

	for _, specPath := range specPaths {
		result, err := validator.Validate(specPath)
		if err != nil {
			// Continue validation for other specs even if one fails
			if result != nil {
				results = append(results, result)
			}
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// HasErrors checks if any validation result has errors
func HasErrors(results []*ValidationResult) bool {
	for _, result := range results {
		if !result.Valid {
			return true
		}
	}
	return false
}

// FormatValidationResult formats a validation result for display
func FormatValidationResult(result *ValidationResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Validation Result for: %s\n", result.SpecInfo.Path))
	sb.WriteString(fmt.Sprintf("  Format: %s\n", result.SpecInfo.Format))
	sb.WriteString(fmt.Sprintf("  OpenAPI Version: %s\n", result.SpecInfo.Version))
	if result.SpecInfo.Title != "" {
		sb.WriteString(fmt.Sprintf("  Title: %s\n", result.SpecInfo.Title))
	}
	sb.WriteString(fmt.Sprintf("  Has Security: %v\n", result.SpecInfo.HasSecurity))
	if result.SpecInfo.HasSecurity {
		sb.WriteString(fmt.Sprintf("  Security Schemes: %d (%v)\n", result.SpecInfo.SchemeCount, result.SpecInfo.SchemeNames))
	}
	sb.WriteString("\n")

	if len(result.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("❌ Errors (%d):\n", len(result.Errors)))
		for i, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s: %s\n", i+1, err.Code, err.Field, err.Message))
		}
		sb.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("⚠️  Warnings (%d):\n", len(result.Warnings)))
		for i, warn := range result.Warnings {
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s: %s\n", i+1, warn.Code, warn.Field, warn.Message))
		}
		sb.WriteString("\n")
	}

	if result.Valid {
		sb.WriteString("✅ Validation: PASSED\n")
	} else {
		sb.WriteString("❌ Validation: FAILED\n")
	}

	return sb.String()
}
