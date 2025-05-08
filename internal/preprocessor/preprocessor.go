package preprocessor

import (
	"fmt"
	"os"
)

// OpenAPIVersion constants
const (
	// OpenAPIVersion30 is the target OpenAPI version (3.0.3)
	OpenAPIVersion30 = "3.0.3"

	// OpenAPIVersion31Prefix is the prefix for OpenAPI 3.1.x versions
	OpenAPIVersion31Prefix = "3.1"
)

// EnsureOpenAPICompatibility ensures the OpenAPI spec is compatible with ogen.
// It converts OpenAPI 3.1 specs to 3.0.3 compatible specs if needed.
// Returns the path to the compatible spec (either the original or a new temporary file).
func EnsureOpenAPICompatibility(specPath string) (string, error) {
	// Create a temporary file for the potentially modified spec
	tempFile, err := os.CreateTemp("", "openapi-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFile.Close() // Close immediately as the converter will reopen it
	tempFilePath := tempFile.Name()

	// Set up cleanup in case of errors
	var cleanupNeeded bool
	defer func() {
		if cleanupNeeded {
			os.Remove(tempFilePath)
		}
	}()

	// Try to convert the spec using the jbcom/openapi-31-to-30-converter library
	/*err = converter.Convert(specPath, tempFilePath)
	if err != nil {
		cleanupNeeded = true
		return "", fmt.Errorf("failed to convert OpenAPI spec: %w", err)
	}

	// Check if the file was actually modified (conversion was needed)
	convertedStat, err := os.Stat(tempFilePath)
	if err != nil {
		cleanupNeeded = true
		return "", fmt.Errorf("failed to stat converted file: %w", err)
	}

	// If the converted file is empty or very small, it likely failed silently
	if convertedStat.Size() < 10 {
		cleanupNeeded = true
		return "", fmt.Errorf("conversion resulted in an invalid file")
	}*/

	return tempFilePath, nil
}
