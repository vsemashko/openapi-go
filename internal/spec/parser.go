package spec

import (
	"encoding/json"
	"fmt"
	"os"
)

// OpenAPISpec represents a minimal OpenAPI specification structure
// We only parse the parts we need for security detection
type OpenAPISpec struct {
	OpenAPI    string                    `json:"openapi"`
	Info       map[string]interface{}    `json:"info"`
	Security   []map[string][]string     `json:"security,omitempty"`
	Components *Components               `json:"components,omitempty"`
}

// Components represents the components section of OpenAPI spec
type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme definition
type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	In           string `json:"in,omitempty"`
	Name         string `json:"name,omitempty"`
}

// ParseSpecFile parses an OpenAPI specification file
func ParseSpecFile(specPath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse spec JSON: %w", err)
	}

	return &spec, nil
}

// HasSecurity checks if the spec defines any security requirements
func (s *OpenAPISpec) HasSecurity() bool {
	// Check global security requirements
	if len(s.Security) > 0 {
		return true
	}

	// Check if security schemes are defined
	if s.Components != nil && len(s.Components.SecuritySchemes) > 0 {
		return true
	}

	return false
}

// GetSecuritySchemes returns all defined security schemes
func (s *OpenAPISpec) GetSecuritySchemes() map[string]SecurityScheme {
	if s.Components == nil {
		return nil
	}
	return s.Components.SecuritySchemes
}
