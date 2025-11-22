package processor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/paths"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/spec"
)

// ApplyPostProcessors applies post-processing steps to the generated client code.
// This includes creating additional client files with convenience functions.
func ApplyPostProcessors(clientPath, serviceName, specPath string) error {
	// Generate the internal client file
	if err := generateInternalClientFile(clientPath, serviceName, specPath); err != nil {
		return fmt.Errorf("failed to generate internal client file: %w", err)
	}

	return nil
}

// generateInternalClientFile creates a file with the NewInternalClient function
// that initializes a client with base security for internal endpoints.
func generateInternalClientFile(clientPath, serviceName, specPath string) error {
	// Use absolute path to template
	templatePath := paths.GetInternalClientTemplatePath()

	// Verify template exists
	if err := paths.EnsurePathExists(templatePath); err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// Parse OpenAPI spec to detect security requirements
	hasSecurity, err := detectSecurityFromSpec(specPath)
	if err != nil {
		// Fall back to file-based detection if spec parsing fails
		log.Printf("Warning: Failed to parse spec for security detection, falling back to file check: %v", err)
		hasSecurity = detectSecurityFromGeneratedFiles(clientPath)
	}

	log.Printf("Security detection for %s: hasSecurity=%v", serviceName, hasSecurity)

	// Create the template data
	data := struct {
		PackageName string
		HasSecurity bool
	}{
		PackageName: serviceName,
		HasSecurity: hasSecurity,
	}

	// Parse the template from file
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", templatePath, err)
	}

	// Create the output file
	outputPath := filepath.Join(clientPath, "oas_internal_client_gen.go")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Execute the template
	if err := tmpl.ExecuteTemplate(file, filepath.Base(templatePath), data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// detectSecurityFromSpec parses the OpenAPI spec to check for security schemes
func detectSecurityFromSpec(specPath string) (bool, error) {
	openAPISpec, err := spec.ParseSpecFile(specPath)
	if err != nil {
		return false, err
	}

	return openAPISpec.HasSecurity(), nil
}

// detectSecurityFromGeneratedFiles checks for security file (fallback method)
func detectSecurityFromGeneratedFiles(clientPath string) bool {
	securityFilePath := filepath.Join(clientPath, "oas_security_gen.go")
	_, err := os.Stat(securityFilePath)
	return err == nil
}
