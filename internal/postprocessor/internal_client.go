package postprocessor

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/paths"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/spec"
)

// InternalClientProcessor generates an internal client file with convenience functions
// for initializing clients with base security for internal endpoints.
type InternalClientProcessor struct {
	templatePath string
}

// NewInternalClientProcessor creates a new internal client processor
func NewInternalClientProcessor() *InternalClientProcessor {
	return &InternalClientProcessor{
		templatePath: paths.GetInternalClientTemplatePath(),
	}
}

// Name returns the processor name
func (p *InternalClientProcessor) Name() string {
	return "InternalClientGenerator"
}

// Process generates the internal client file
func (p *InternalClientProcessor) Process(ctx context.Context, spec ProcessSpec) error {
	// Verify template exists
	if err := paths.EnsurePathExists(p.templatePath); err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// Parse OpenAPI spec to detect security requirements
	hasSecurity, err := p.detectSecurityFromSpec(spec.SpecPath)
	if err != nil {
		// Fall back to file-based detection if spec parsing fails
		log.Printf("Warning: Failed to parse spec for security detection, falling back to file check: %v", err)
		hasSecurity = p.detectSecurityFromGeneratedFiles(spec.ClientPath)
	}

	log.Printf("Security detection for %s: hasSecurity=%v", spec.ServiceName, hasSecurity)

	// Create the template data
	data := struct {
		PackageName string
		HasSecurity bool
	}{
		PackageName: spec.ServiceName,
		HasSecurity: hasSecurity,
	}

	// Parse the template from file
	tmpl, err := template.ParseFiles(p.templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", p.templatePath, err)
	}

	// Create the output file
	outputPath := filepath.Join(spec.ClientPath, "oas_internal_client_gen.go")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Execute the template
	if err := tmpl.ExecuteTemplate(file, filepath.Base(p.templatePath), data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Printf("Generated internal client file: %s", outputPath)
	return nil
}

// detectSecurityFromSpec parses the OpenAPI spec to check for security schemes
func (p *InternalClientProcessor) detectSecurityFromSpec(specPath string) (bool, error) {
	openAPISpec, err := spec.ParseSpecFile(specPath)
	if err != nil {
		return false, err
	}

	return openAPISpec.HasSecurity(), nil
}

// detectSecurityFromGeneratedFiles checks for security file (fallback method)
func (p *InternalClientProcessor) detectSecurityFromGeneratedFiles(clientPath string) bool {
	securityFilePath := filepath.Join(clientPath, "oas_security_gen.go")
	_, err := os.Stat(securityFilePath)
	return err == nil
}
