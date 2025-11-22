package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpecFile(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{
			name: "valid minimal spec",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {}
			}`,
			wantErr: false,
		},
		{
			name: "spec with security schemes",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"components": {
					"securitySchemes": {
						"bearerAuth": {
							"type": "http",
							"scheme": "bearer",
							"bearerFormat": "JWT"
						}
					}
				},
				"paths": {}
			}`,
			wantErr: false,
		},
		{
			name: "invalid JSON",
			spec: `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary spec file
			tmpFile := filepath.Join(t.TempDir(), "openapi.json")
			err := os.WriteFile(tmpFile, []byte(tt.spec), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse spec
			spec, err := ParseSpecFile(tmpFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSpecFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify spec was parsed
				if spec.OpenAPI == "" {
					t.Error("ParseSpecFile() returned spec with empty openapi field")
				}
			}
		})
	}
}

func TestParseSpecFileNonexistent(t *testing.T) {
	_, err := ParseSpecFile("/nonexistent/file.json")
	if err == nil {
		t.Error("ParseSpecFile() should error for nonexistent file")
	}
}

func TestHasSecurity(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		expected bool
	}{
		{
			name: "has bearer auth security scheme",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"components": {
					"securitySchemes": {
						"bearerAuth": {
							"type": "http",
							"scheme": "bearer"
						}
					}
				}
			}`,
			expected: true,
		},
		{
			name: "has api key security scheme",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"components": {
					"securitySchemes": {
						"apiKey": {
							"type": "apiKey",
							"in": "header",
							"name": "X-API-Key"
						}
					}
				}
			}`,
			expected: true,
		},
		{
			name: "has global security requirement",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"security": [{"apiKey": []}],
				"components": {
					"securitySchemes": {
						"apiKey": {
							"type": "apiKey",
							"in": "header",
							"name": "X-API-Key"
						}
					}
				}
			}`,
			expected: true,
		},
		{
			name: "no security",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"paths": {}
			}`,
			expected: false,
		},
		{
			name: "empty security schemes",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"components": {
					"securitySchemes": {}
				}
			}`,
			expected: false,
		},
		{
			name: "components but no security schemes",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"components": {
					"schemas": {
						"User": {
							"type": "object",
							"properties": {
								"id": {"type": "string"}
							}
						}
					}
				}
			}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write spec to temp file
			tmpFile := filepath.Join(t.TempDir(), "spec.json")
			err := os.WriteFile(tmpFile, []byte(tt.spec), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse spec
			spec, err := ParseSpecFile(tmpFile)
			if err != nil {
				t.Fatalf("ParseSpecFile() error = %v", err)
			}

			// Check security
			result := spec.HasSecurity()
			if result != tt.expected {
				t.Errorf("HasSecurity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetSecuritySchemes(t *testing.T) {
	tests := []struct {
		name          string
		spec          string
		expectedCount int
		expectedNames []string
	}{
		{
			name: "single bearer auth",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"components": {
					"securitySchemes": {
						"bearerAuth": {
							"type": "http",
							"scheme": "bearer"
						}
					}
				}
			}`,
			expectedCount: 1,
			expectedNames: []string{"bearerAuth"},
		},
		{
			name: "multiple security schemes",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"},
				"components": {
					"securitySchemes": {
						"bearerAuth": {
							"type": "http",
							"scheme": "bearer"
						},
						"apiKey": {
							"type": "apiKey",
							"in": "header",
							"name": "X-API-Key"
						}
					}
				}
			}`,
			expectedCount: 2,
			expectedNames: []string{"bearerAuth", "apiKey"},
		},
		{
			name: "no security schemes",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"}
			}`,
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write spec to temp file
			tmpFile := filepath.Join(t.TempDir(), "spec.json")
			err := os.WriteFile(tmpFile, []byte(tt.spec), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse spec
			spec, err := ParseSpecFile(tmpFile)
			if err != nil {
				t.Fatalf("ParseSpecFile() error = %v", err)
			}

			// Get security schemes
			schemes := spec.GetSecuritySchemes()

			// Check count
			if len(schemes) != tt.expectedCount {
				t.Errorf("GetSecuritySchemes() count = %d, want %d", len(schemes), tt.expectedCount)
			}

			// Check expected names are present
			for _, name := range tt.expectedNames {
				if _, ok := schemes[name]; !ok {
					t.Errorf("GetSecuritySchemes() missing expected scheme: %s", name)
				}
			}
		})
	}
}

func TestSecuritySchemeTypes(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0"},
		"components": {
			"securitySchemes": {
				"bearerAuth": {
					"type": "http",
					"scheme": "bearer",
					"bearerFormat": "JWT"
				},
				"apiKey": {
					"type": "apiKey",
					"in": "header",
					"name": "X-API-Key"
				}
			}
		}
	}`

	tmpFile := filepath.Join(t.TempDir(), "spec.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parsed, err := ParseSpecFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseSpecFile() error = %v", err)
	}

	schemes := parsed.GetSecuritySchemes()

	// Verify bearer auth properties
	if bearer, ok := schemes["bearerAuth"]; ok {
		if bearer.Type != "http" {
			t.Errorf("bearerAuth.Type = %q, want %q", bearer.Type, "http")
		}
		if bearer.Scheme != "bearer" {
			t.Errorf("bearerAuth.Scheme = %q, want %q", bearer.Scheme, "bearer")
		}
		if bearer.BearerFormat != "JWT" {
			t.Errorf("bearerAuth.BearerFormat = %q, want %q", bearer.BearerFormat, "JWT")
		}
	} else {
		t.Error("bearerAuth scheme not found")
	}

	// Verify API key properties
	if apiKey, ok := schemes["apiKey"]; ok {
		if apiKey.Type != "apiKey" {
			t.Errorf("apiKey.Type = %q, want %q", apiKey.Type, "apiKey")
		}
		if apiKey.In != "header" {
			t.Errorf("apiKey.In = %q, want %q", apiKey.In, "header")
		}
		if apiKey.Name != "X-API-Key" {
			t.Errorf("apiKey.Name = %q, want %q", apiKey.Name, "X-API-Key")
		}
	} else {
		t.Error("apiKey scheme not found")
	}
}

func TestParseYAMLSpecFile(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		filename string
		wantErr  bool
	}{
		{
			name: "valid YAML spec with .yaml extension",
			spec: `openapi: "3.0.0"
info:
  title: Test API
  version: 1.0.0
paths: {}`,
			filename: "openapi.yaml",
			wantErr:  false,
		},
		{
			name: "valid YAML spec with .yml extension",
			spec: `openapi: "3.0.0"
info:
  title: Test API
  version: 1.0.0
paths: {}`,
			filename: "openapi.yml",
			wantErr:  false,
		},
		{
			name: "YAML spec with security schemes",
			spec: `openapi: "3.0.0"
info:
  title: Test API
  version: 1.0.0
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
paths: {}`,
			filename: "openapi.yaml",
			wantErr:  false,
		},
		{
			name: "YAML spec with global security",
			spec: `openapi: "3.0.0"
info:
  title: Test API
  version: 1.0.0
security:
  - bearerAuth: []
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
paths: {}`,
			filename: "openapi.yaml",
			wantErr:  false,
		},
		{
			name:     "invalid YAML",
			spec:     `{invalid yaml: [}`,
			filename: "openapi.yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary spec file
			tmpFile := filepath.Join(t.TempDir(), tt.filename)
			err := os.WriteFile(tmpFile, []byte(tt.spec), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse spec
			spec, err := ParseSpecFile(tmpFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSpecFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify spec was parsed
				if spec.OpenAPI == "" {
					t.Error("ParseSpecFile() returned spec with empty openapi field")
				}
			}
		})
	}
}

func TestYAMLSecurityDetection(t *testing.T) {
	tests := []struct {
		name         string
		spec         string
		hasSecurity  bool
		schemeCount  int
		schemeNames  []string
	}{
		{
			name: "YAML with bearer auth",
			spec: `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT`,
			hasSecurity: true,
			schemeCount: 1,
			schemeNames: []string{"bearerAuth"},
		},
		{
			name: "YAML with multiple security schemes",
			spec: `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
    apiKey:
      type: apiKey
      in: header
      name: X-API-Key`,
			hasSecurity: true,
			schemeCount: 2,
			schemeNames: []string{"bearerAuth", "apiKey"},
		},
		{
			name: "YAML with no security",
			spec: `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths: {}`,
			hasSecurity: false,
			schemeCount: 0,
			schemeNames: []string{},
		},
		{
			name: "YAML with global security requirement",
			spec: `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
security:
  - apiKey: []
components:
  securitySchemes:
    apiKey:
      type: apiKey
      in: header
      name: X-API-Key`,
			hasSecurity: true,
			schemeCount: 1,
			schemeNames: []string{"apiKey"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write spec to temp file with YAML extension
			tmpFile := filepath.Join(t.TempDir(), "spec.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.spec), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse spec
			spec, err := ParseSpecFile(tmpFile)
			if err != nil {
				t.Fatalf("ParseSpecFile() error = %v", err)
			}

			// Check HasSecurity
			if spec.HasSecurity() != tt.hasSecurity {
				t.Errorf("HasSecurity() = %v, want %v", spec.HasSecurity(), tt.hasSecurity)
			}

			// Check security schemes count
			schemes := spec.GetSecuritySchemes()
			if len(schemes) != tt.schemeCount {
				t.Errorf("GetSecuritySchemes() count = %d, want %d", len(schemes), tt.schemeCount)
			}

			// Check expected scheme names
			for _, name := range tt.schemeNames {
				if _, ok := schemes[name]; !ok {
					t.Errorf("GetSecuritySchemes() missing expected scheme: %s", name)
				}
			}
		})
	}
}

func TestMixedFormatParsing(t *testing.T) {
	// Test that both JSON and YAML can be parsed correctly
	jsonSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "JSON Test", "version": "1.0"},
		"components": {
			"securitySchemes": {
				"bearerAuth": {
					"type": "http",
					"scheme": "bearer"
				}
			}
		}
	}`

	yamlSpec := `openapi: "3.0.0"
info:
  title: YAML Test
  version: "1.0"
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer`

	// Parse JSON
	jsonFile := filepath.Join(t.TempDir(), "spec.json")
	err := os.WriteFile(jsonFile, []byte(jsonSpec), 0644)
	if err != nil {
		t.Fatalf("Failed to create JSON test file: %v", err)
	}

	jsonParsed, err := ParseSpecFile(jsonFile)
	if err != nil {
		t.Fatalf("ParseSpecFile() JSON error = %v", err)
	}

	// Parse YAML
	yamlFile := filepath.Join(t.TempDir(), "spec.yaml")
	err = os.WriteFile(yamlFile, []byte(yamlSpec), 0644)
	if err != nil {
		t.Fatalf("Failed to create YAML test file: %v", err)
	}

	yamlParsed, err := ParseSpecFile(yamlFile)
	if err != nil {
		t.Fatalf("ParseSpecFile() YAML error = %v", err)
	}

	// Both should have security
	if !jsonParsed.HasSecurity() {
		t.Error("JSON spec should have security")
	}
	if !yamlParsed.HasSecurity() {
		t.Error("YAML spec should have security")
	}

	// Both should have bearerAuth scheme
	jsonSchemes := jsonParsed.GetSecuritySchemes()
	yamlSchemes := yamlParsed.GetSecuritySchemes()

	if _, ok := jsonSchemes["bearerAuth"]; !ok {
		t.Error("JSON spec missing bearerAuth scheme")
	}
	if _, ok := yamlSchemes["bearerAuth"]; !ok {
		t.Error("YAML spec missing bearerAuth scheme")
	}
}

func TestUnknownExtensionFallback(t *testing.T) {
	// Test that files with unknown extensions try JSON then YAML
	tests := []struct {
		name     string
		content  string
		filename string
		wantErr  bool
	}{
		{
			name: "unknown extension with JSON content",
			content: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0"}
			}`,
			filename: "spec.txt",
			wantErr:  false,
		},
		{
			name: "unknown extension with YAML content",
			content: `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"`,
			filename: "spec.txt",
			wantErr:  false,
		},
		{
			name:     "unknown extension with invalid content",
			content:  `not valid json or yaml {[}]`,
			filename: "spec.txt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), tt.filename)
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			spec, err := ParseSpecFile(tmpFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSpecFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && spec.OpenAPI == "" {
				t.Error("ParseSpecFile() returned spec with empty openapi field")
			}
		})
	}
}
