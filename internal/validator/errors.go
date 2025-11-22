package validator

import (
	"fmt"
	"strings"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/errors"
)

// ToGenerationError converts a ValidationError to a structured GenerationError with suggestions
func ToGenerationError(ve ValidationError, specPath string) *errors.GenerationError {
	// Map validation codes to error codes
	errorCode := mapValidationCodeToErrorCode(ve.Code)

	err := errors.New(errorCode, ve.Message)

	// Add file location if we have field information
	if ve.Field != "" {
		err = err.WithLocation(specPath, 0, 0)
		err = err.WithContext("field", ve.Field)
	}

	// Get contextual suggestion
	suggestionProvider := errors.NewSuggestionProvider()
	context := map[string]interface{}{
		"file":  specPath,
		"field": ve.Field,
	}
	suggestion := suggestionProvider.GetSuggestion(errorCode, context)
	if suggestion != "" {
		err = err.WithSuggestion(suggestion)
	}

	return err
}

// mapValidationCodeToErrorCode maps validator error codes to error package codes
func mapValidationCodeToErrorCode(code string) errors.ErrorCode {
	codeMap := map[string]errors.ErrorCode{
		"FILE_NOT_FOUND":          errors.ErrCodeFileNotFound,
		"FILE_ACCESS_ERROR":       errors.ErrCodeFileAccessDenied,
		"NOT_A_FILE":              errors.ErrCodeFileIsDirectory,
		"PARSE_ERROR":             errors.ErrCodeSpecParseError,
		"INVALID_FORMAT":          errors.ErrCodeSpecInvalidFormat,
		"UNSUPPORTED_VERSION":     errors.ErrCodeSpecUnsupportedVer,
		"MISSING_OPENAPI_VERSION": errors.ErrCodeSpecMissingField,
		"MISSING_INFO":            errors.ErrCodeSpecMissingField,
		"MISSING_TITLE":           errors.ErrCodeSpecMissingField,
		"MISSING_VERSION":         errors.ErrCodeSpecMissingField,
		"INVALID_VERSION_FORMAT":  errors.ErrCodeSpecInvalidField,
		"MISSING_OPERATION_ID":    errors.ErrCodeSpecMissingOpID,
		"DUPLICATE_OPERATION_ID":  errors.ErrCodeSpecDuplicateOpID,
		"INVALID_REF":             errors.ErrCodeSpecInvalidRef,
		"MISSING_SCHEMA":          errors.ErrCodeSpecMissingSchema,
	}

	if errorCode, exists := codeMap[code]; exists {
		return errorCode
	}

	return errors.ErrCodeSpecInvalidField
}

// FormatValidationResultEnhanced formats a validation result with enhanced error messages
func FormatValidationResultEnhanced(result *ValidationResult) string {
	var lines []string

	// Header
	specName := result.SpecInfo.Path
	if result.Valid && len(result.Warnings) == 0 {
		lines = append(lines, fmt.Sprintf("✅ %s - Valid", specName))
		return strings.Join(lines, "\n")
	}

	if !result.Valid {
		lines = append(lines, fmt.Sprintf("❌ Validation failed for: %s", specName))
	} else {
		lines = append(lines, fmt.Sprintf("⚠️  Validation warnings for: %s", specName))
	}

	// Spec info
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("   Spec Info:"))
	lines = append(lines, fmt.Sprintf("   - Version: %s", result.SpecInfo.Version))
	lines = append(lines, fmt.Sprintf("   - Format: %s", result.SpecInfo.Format))
	if result.SpecInfo.Title != "" {
		lines = append(lines, fmt.Sprintf("   - Title: %s", result.SpecInfo.Title))
	}
	if result.SpecInfo.HasSecurity {
		lines = append(lines, fmt.Sprintf("   - Security: %d scheme(s) - %s",
			result.SpecInfo.SchemeCount,
			strings.Join(result.SpecInfo.SchemeNames, ", ")))
	}

	// Errors
	if len(result.Errors) > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("   Errors (%d):", len(result.Errors)))

		for i, err := range result.Errors {
			genErr := ToGenerationError(err, specName)
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("   %d. %s", i+1, genErr.Format()))
		}
	}

	// Warnings
	if len(result.Warnings) > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("   Warnings (%d):", len(result.Warnings)))

		for i, warning := range result.Warnings {
			lines = append(lines, fmt.Sprintf("   %d. [%s] %s: %s",
				i+1, warning.Code, warning.Field, warning.Message))
		}
	}

	return strings.Join(lines, "\n")
}

// FormatValidationErrors formats multiple validation errors into a single error list
func FormatValidationErrors(results []*ValidationResult) *errors.ErrorList {
	errorList := &errors.ErrorList{}

	for _, result := range results {
		if !result.Valid {
			for _, ve := range result.Errors {
				genErr := ToGenerationError(ve, result.SpecInfo.Path)
				errorList.Add(genErr)
			}
		}
	}

	return errorList
}
