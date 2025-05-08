package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ApplyPostProcessors applies post-processing steps to the generated client code.
// This includes creating additional client files with convenience functions.
func ApplyPostProcessors(clientPath, serviceName string) error {
	// Generate the internal client file
	if err := generateInternalClientFile(clientPath, serviceName); err != nil {
		return fmt.Errorf("failed to generate internal client file: %w", err)
	}

	return nil
}

// generateInternalClientFile creates a file with the NewInternalClient function
// that initializes a client with base security for internal endpoints.
func generateInternalClientFile(clientPath, serviceName string) error {
	// Path to the template file
	templatePath := "resources/templates/internal_client.tmpl"

	// Check if the client has security by looking for the security file
	securityFilePath := filepath.Join(clientPath, "oas_security_gen.go")
	hasSecurity := false
	if _, err := os.Stat(securityFilePath); err == nil {
		hasSecurity = true
	}

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
