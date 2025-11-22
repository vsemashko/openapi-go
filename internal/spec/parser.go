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
// We only parse the parts we need for security detection and operation tracking
type OpenAPISpec struct {
	OpenAPI    string                    `json:"openapi" yaml:"openapi"`
	Info       map[string]interface{}    `json:"info" yaml:"info"`
	Security   []map[string][]string     `json:"security,omitempty" yaml:"security,omitempty"`
	Components *Components               `json:"components,omitempty" yaml:"components,omitempty"`
	Paths      map[string]PathItem       `json:"paths,omitempty" yaml:"paths,omitempty"`
}

// PathItem represents an OpenAPI path item with operations
type PathItem struct {
	Get     *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post    *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put     *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Patch   *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	Delete  *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options *Operation `json:"options,omitempty" yaml:"options,omitempty"`
	Head    *Operation `json:"head,omitempty" yaml:"head,omitempty"`
	Trace   *Operation `json:"trace,omitempty" yaml:"trace,omitempty"`
}

// Operation represents an OpenAPI operation
type Operation struct {
	OperationID string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Summary     string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty" yaml:"tags,omitempty"`
	Parameters  []interface{}          `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody interface{}            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]interface{} `json:"responses,omitempty" yaml:"responses,omitempty"`
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

// OperationInfo contains information about a single operation
type OperationInfo struct {
	Path        string
	Method      string
	OperationID string
	Summary     string
	Operation   *Operation
}

// GetOperations extracts all operations from the spec
func (s *OpenAPISpec) GetOperations() []OperationInfo {
	var operations []OperationInfo

	if s.Paths == nil {
		return operations
	}

	for path, pathItem := range s.Paths {
		// Check each HTTP method
		if pathItem.Get != nil {
			operations = append(operations, OperationInfo{
				Path:        path,
				Method:      "GET",
				OperationID: pathItem.Get.OperationID,
				Summary:     pathItem.Get.Summary,
				Operation:   pathItem.Get,
			})
		}
		if pathItem.Post != nil {
			operations = append(operations, OperationInfo{
				Path:        path,
				Method:      "POST",
				OperationID: pathItem.Post.OperationID,
				Summary:     pathItem.Post.Summary,
				Operation:   pathItem.Post,
			})
		}
		if pathItem.Put != nil {
			operations = append(operations, OperationInfo{
				Path:        path,
				Method:      "PUT",
				OperationID: pathItem.Put.OperationID,
				Summary:     pathItem.Put.Summary,
				Operation:   pathItem.Put,
			})
		}
		if pathItem.Patch != nil {
			operations = append(operations, OperationInfo{
				Path:        path,
				Method:      "PATCH",
				OperationID: pathItem.Patch.OperationID,
				Summary:     pathItem.Patch.Summary,
				Operation:   pathItem.Patch,
			})
		}
		if pathItem.Delete != nil {
			operations = append(operations, OperationInfo{
				Path:        path,
				Method:      "DELETE",
				OperationID: pathItem.Delete.OperationID,
				Summary:     pathItem.Delete.Summary,
				Operation:   pathItem.Delete,
			})
		}
		if pathItem.Options != nil {
			operations = append(operations, OperationInfo{
				Path:        path,
				Method:      "OPTIONS",
				OperationID: pathItem.Options.OperationID,
				Summary:     pathItem.Options.Summary,
				Operation:   pathItem.Options,
			})
		}
		if pathItem.Head != nil {
			operations = append(operations, OperationInfo{
				Path:        path,
				Method:      "HEAD",
				OperationID: pathItem.Head.OperationID,
				Summary:     pathItem.Head.Summary,
				Operation:   pathItem.Head,
			})
		}
		if pathItem.Trace != nil {
			operations = append(operations, OperationInfo{
				Path:        path,
				Method:      "TRACE",
				OperationID: pathItem.Trace.OperationID,
				Summary:     pathItem.Trace.Summary,
				Operation:   pathItem.Trace,
			})
		}
	}

	return operations
}

// GetOperationCount returns the total number of operations in the spec
func (s *OpenAPISpec) GetOperationCount() int {
	return len(s.GetOperations())
}
