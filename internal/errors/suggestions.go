package errors

import (
	"fmt"
	"strings"
)

// SuggestionProvider provides contextual suggestions for errors
type SuggestionProvider struct{}

// NewSuggestionProvider creates a new suggestion provider
func NewSuggestionProvider() *SuggestionProvider {
	return &SuggestionProvider{}
}

// GetSuggestion returns a helpful suggestion based on error code and context
func (sp *SuggestionProvider) GetSuggestion(code ErrorCode, context map[string]interface{}) string {
	switch code {
	// File System Errors
	case ErrCodeFileNotFound:
		return sp.suggestFileNotFound(context)
	case ErrCodeFileAccessDenied:
		return sp.suggestFileAccessDenied(context)
	case ErrCodeFileIsDirectory:
		return "The path points to a directory. Provide a path to an OpenAPI spec file (e.g., openapi.yaml)"
	case ErrCodeDirectoryNotFound:
		return sp.suggestDirectoryNotFound(context)

	// Spec Validation Errors
	case ErrCodeSpecParseError:
		return sp.suggestSpecParseError(context)
	case ErrCodeSpecInvalidFormat:
		return "Ensure the file is valid JSON or YAML. Check for syntax errors like unclosed brackets or invalid indentation"
	case ErrCodeSpecUnsupportedVer:
		return "This tool supports OpenAPI 3.0.x. Consider downgrading your spec or check if the version field is correct"
	case ErrCodeSpecMissingField:
		return sp.suggestMissingField(context)
	case ErrCodeSpecMissingOpID:
		return "Add an 'operationId' field to the operation. Example: operationId: getUsers"
	case ErrCodeSpecDuplicateOpID:
		return "Each operation must have a unique operationId. Check for duplicates in your spec"
	case ErrCodeSpecInvalidRef:
		return sp.suggestInvalidRef(context)
	case ErrCodeSpecMissingSchema:
		return "The referenced schema doesn't exist. Check the components.schemas section"

	// Generation Errors
	case ErrCodeGeneratorNotFound:
		return "Install ogen using: go install github.com/ogen-go/ogen/cmd/ogen@v1.14.0"
	case ErrCodeGeneratorFailed:
		return sp.suggestGeneratorFailed(context)
	case ErrCodeGeneratorInstall:
		return "Check your network connection and Go installation. Try: go install github.com/ogen-go/ogen/cmd/ogen@v1.14.0"
	case ErrCodeGeneratorVersion:
		return "Run: go install github.com/ogen-go/ogen/cmd/ogen@v1.14.0 to install the correct version"
	case ErrCodeGeneratorTimeout:
		return "The generation process took too long. Try with a smaller spec or increase the timeout"

	// Post-Processing Errors
	case ErrCodePostProcessFailed:
		return "Check if the generated code is valid Go code. There may be an issue with the spec"
	case ErrCodeFormattingFailed:
		return "Install gofmt or check if the generated code has syntax errors"

	// Configuration Errors
	case ErrCodeConfigInvalid:
		return sp.suggestConfigInvalid(context)
	case ErrCodeConfigMissing:
		return "Create a configuration file at resources/application.yml or use environment variables"
	case ErrCodeConfigLoadFailed:
		return "Check if the config file exists and is valid YAML"

	// Cache Errors
	case ErrCodeCacheReadFailed:
		return "The cache file may be corrupted. Delete .openapi-cache.json and try again"
	case ErrCodeCacheWriteFailed:
		return "Check write permissions for the output directory"

	// Network Errors
	case ErrCodeNetworkTimeout:
		return "Check your network connection and try again"
	case ErrCodeNetworkUnavailable:
		return "Ensure you have network access. If using a proxy, configure Go's proxy settings"

	default:
		return "Check the error message for more details"
	}
}

func (sp *SuggestionProvider) suggestFileNotFound(context map[string]interface{}) string {
	file, ok := context["file"].(string)
	if !ok {
		return "Check if the file path is correct and the file exists"
	}

	suggestions := []string{
		fmt.Sprintf("Check if '%s' exists", file),
	}

	// Suggest common file name patterns
	if strings.Contains(file, "openapi") {
		suggestions = append(suggestions, "Common OpenAPI spec names: openapi.yaml, openapi.json, swagger.json")
	}

	// Suggest checking directory
	if strings.Contains(file, "/") {
		suggestions = append(suggestions, "Verify the directory path is correct")
	}

	return strings.Join(suggestions, ". ")
}

func (sp *SuggestionProvider) suggestFileAccessDenied(context map[string]interface{}) string {
	file, ok := context["file"].(string)
	if !ok {
		return "Check file permissions. Use 'chmod +r' to add read permissions"
	}
	return fmt.Sprintf("Check permissions for '%s'. Use 'chmod +r %s' to add read permissions", file, file)
}

func (sp *SuggestionProvider) suggestDirectoryNotFound(context map[string]interface{}) string {
	dir, ok := context["directory"].(string)
	if !ok {
		return "Create the directory or check the path in your configuration"
	}
	return fmt.Sprintf("Create the directory: mkdir -p %s", dir)
}

func (sp *SuggestionProvider) suggestSpecParseError(context map[string]interface{}) string {
	format, ok := context["format"].(string)
	if !ok {
		return "Check for syntax errors in your OpenAPI spec (JSON or YAML)"
	}

	if format == "yaml" || format == "yml" {
		return "Check YAML syntax: ensure proper indentation (use spaces, not tabs), close all quotes, and avoid special characters without quotes"
	}
	return "Check JSON syntax: ensure all brackets are closed, use double quotes for strings, and remove trailing commas"
}

func (sp *SuggestionProvider) suggestMissingField(context map[string]interface{}) string {
	field, ok := context["field"].(string)
	if !ok {
		return "Add the required field to your OpenAPI spec"
	}

	examples := map[string]string{
		"info":    "Add info section: info:\n  title: My API\n  version: 1.0.0",
		"title":   "Add title to info: info:\n  title: My API Name",
		"version": "Add version to info: info:\n  version: 1.0.0",
		"paths":   "Add paths section with at least one endpoint",
	}

	if example, exists := examples[field]; exists {
		return fmt.Sprintf("Add the '%s' field. Example:\n%s", field, example)
	}

	return fmt.Sprintf("Add the required '%s' field to your OpenAPI spec", field)
}

func (sp *SuggestionProvider) suggestInvalidRef(context map[string]interface{}) string {
	ref, ok := context["ref"].(string)
	if !ok {
		return "Check that all $ref paths point to existing definitions in components"
	}

	// Parse the ref to understand what's missing
	parts := strings.Split(ref, "/")
	if len(parts) >= 3 {
		section := parts[len(parts)-2] // e.g., "schemas", "parameters"
		name := parts[len(parts)-1]    // e.g., "User", "PageParam"

		return fmt.Sprintf("The reference '%s' doesn't exist. Add it to components.%s or fix the $ref path", name, section)
	}

	return fmt.Sprintf("Check the $ref path '%s' and ensure the referenced component exists", ref)
}

func (sp *SuggestionProvider) suggestGeneratorFailed(context map[string]interface{}) string {
	// Check if we have ogen-specific error information
	if ogenErr, ok := context["ogen_error"].(string); ok {
		if strings.Contains(ogenErr, "nullable") {
			return "Remove 'nullable: true' and use type arrays for OpenAPI 3.1: type: [\"string\", \"null\"]"
		}
		if strings.Contains(ogenErr, "exclusiveMinimum") || strings.Contains(ogenErr, "exclusiveMaximum") {
			return "Use boolean syntax for exclusive min/max in OpenAPI 3.0: exclusiveMinimum: true (not numeric values)"
		}
	}

	return "Check the ogen error message above for specific issues. Common problems: invalid schema types, missing references, unsupported OpenAPI features"
}

func (sp *SuggestionProvider) suggestConfigInvalid(context map[string]interface{}) string {
	field, ok := context["field"].(string)
	if !ok {
		return "Check your configuration file (resources/application.yml) for syntax errors"
	}

	examples := map[string]string{
		"specs_dir":      "Set specs_dir to the directory containing OpenAPI specs, e.g., specs_dir: ./specs",
		"output_dir":     "Set output_dir to where generated code should go, e.g., output_dir: ./generated",
		"worker_count":   "Set worker_count to a positive integer, e.g., worker_count: 4",
		"target_services": "Set target_services pattern, e.g., target_services: \".*\" for all services",
	}

	if example, exists := examples[field]; exists {
		return example
	}

	return fmt.Sprintf("Fix the '%s' configuration field. Check the configuration guide for valid values", field)
}

// GetCommonMistakes returns common mistakes for a given error code
func (sp *SuggestionProvider) GetCommonMistakes(code ErrorCode) []string {
	switch code {
	case ErrCodeSpecParseError:
		return []string{
			"Using tabs instead of spaces in YAML",
			"Missing quotes around special characters",
			"Unclosed brackets or braces",
			"Trailing commas in JSON",
		}
	case ErrCodeSpecMissingOpID:
		return []string{
			"Forgetting to add operationId to operations",
			"Using duplicate operationId values",
			"Using invalid characters in operationId",
		}
	case ErrCodeGeneratorFailed:
		return []string{
			"Using OpenAPI 3.1 syntax with 3.0 generators",
			"Invalid schema references",
			"Circular dependencies in schemas",
			"Missing required fields in schemas",
		}
	default:
		return nil
	}
}
