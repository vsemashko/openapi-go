package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// OpenAPISpec represents a minimal OpenAPI specification structure
// We only parse the parts we need for security detection
type OpenAPISpec struct {
	OpenAPI    string                    `json:"openapi" yaml:"openapi"`
	Info       map[string]interface{}    `json:"info" yaml:"info"`
	Security   []map[string][]string     `json:"security,omitempty" yaml:"security,omitempty"`
	Components *Components               `json:"components,omitempty" yaml:"components,omitempty"`
}

// Components represents the components section of OpenAPI spec
type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme definition
type SecurityScheme struct {
	Type         string `json:"type" yaml:"type"`
	Scheme       string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	In           string `json:"in,omitempty" yaml:"in,omitempty"`
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
}

// ParseSpecFile parses an OpenAPI specification file (JSON or YAML)
func ParseSpecFile(specPath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec OpenAPISpec

	// Detect format by file extension
	ext := strings.ToLower(filepath.Ext(specPath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse spec YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse spec JSON: %w", err)
		}
	default:
		// Try JSON first, then YAML as fallback
		if err := json.Unmarshal(data, &spec); err != nil {
			if yamlErr := yaml.Unmarshal(data, &spec); yamlErr != nil {
				return nil, fmt.Errorf("failed to parse spec (tried JSON and YAML): JSON error: %w, YAML error: %v", err, yamlErr)
			}
		}
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
