package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateSpecFingerprint(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"summary": "Get all users",
					"responses": {
						"200": {"description": "Success"}
					}
				},
				"post": {
					"operationId": "createUser",
					"summary": "Create a user",
					"responses": {
						"201": {"description": "Created"}
					}
				}
			},
			"/users/{id}": {
				"get": {
					"operationId": "getUser",
					"parameters": [{"name": "id", "in": "path"}],
					"responses": {
						"200": {"description": "Success"}
					}
				}
			}
		}
	}`

	tmpFile := filepath.Join(t.TempDir(), "spec.json")
	err := os.WriteFile(tmpFile, []byte(spec), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parsedSpec, err := ParseSpecFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	fingerprint, err := CreateSpecFingerprint(tmpFile, parsedSpec)
	if err != nil {
		t.Fatalf("CreateSpecFingerprint() error = %v", err)
	}

	// Check spec path
	if fingerprint.SpecPath != tmpFile {
		t.Errorf("SpecPath = %v, want %v", fingerprint.SpecPath, tmpFile)
	}

	// Check spec hash exists
	if fingerprint.SpecHash == "" {
		t.Error("SpecHash should not be empty")
	}

	// Check operations count
	if len(fingerprint.Operations) != 3 {
		t.Errorf("Expected 3 operations, got %d", len(fingerprint.Operations))
	}

	// Check specific operations exist
	expectedOps := []string{
		"GET /users",
		"POST /users",
		"GET /users/{id}",
	}

	for _, op := range expectedOps {
		if _, exists := fingerprint.Operations[op]; !exists {
			t.Errorf("Operation %s not found", op)
		}
	}

	// Check operation IDs mapped
	if len(fingerprint.OperationIDs) != 3 {
		t.Errorf("Expected 3 operation ID mappings, got %d", len(fingerprint.OperationIDs))
	}

	// Check operation hashes are unique
	hashes := make(map[string]bool)
	for _, op := range fingerprint.Operations {
		if op.Hash == "" {
			t.Error("Operation hash should not be empty")
		}
		hashes[op.Hash] = true
	}
	if len(hashes) != 3 {
		t.Error("Each operation should have a unique hash")
	}
}

func TestCompareFingerprints_NoChanges(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"responses": {"200": {"description": "Success"}}
				}
			}
		}
	}`

	tmpFile := filepath.Join(t.TempDir(), "spec.json")
	os.WriteFile(tmpFile, []byte(spec), 0644)

	parsedSpec, _ := ParseSpecFile(tmpFile)
	fp1, _ := CreateSpecFingerprint(tmpFile, parsedSpec)
	fp2, _ := CreateSpecFingerprint(tmpFile, parsedSpec)

	comparison := CompareFingerprints(fp1, fp2)

	if comparison.HasChanges() {
		t.Error("Should detect no changes for identical specs")
	}

	if len(comparison.Unchanged) != 1 {
		t.Errorf("Expected 1 unchanged operation, got %d", len(comparison.Unchanged))
	}

	expected := "No changes detected"
	if comparison.Summary() != expected {
		t.Errorf("Summary = %q, want %q", comparison.Summary(), expected)
	}
}

func TestCompareFingerprints_AddedOperation(t *testing.T) {
	spec1 := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"responses": {"200": {"description": "Success"}}
				}
			}
		}
	}`

	spec2 := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"responses": {"200": {"description": "Success"}}
				},
				"post": {
					"operationId": "createUser",
					"responses": {"201": {"description": "Created"}}
				}
			}
		}
	}`

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "spec1.json")
	file2 := filepath.Join(tmpDir, "spec2.json")

	os.WriteFile(file1, []byte(spec1), 0644)
	os.WriteFile(file2, []byte(spec2), 0644)

	parsed1, _ := ParseSpecFile(file1)
	parsed2, _ := ParseSpecFile(file2)

	fp1, _ := CreateSpecFingerprint(file1, parsed1)
	fp2, _ := CreateSpecFingerprint(file2, parsed2)

	comparison := CompareFingerprints(fp1, fp2)

	if !comparison.HasChanges() {
		t.Error("Should detect changes")
	}

	if len(comparison.Added) != 1 {
		t.Errorf("Expected 1 added operation, got %d", len(comparison.Added))
	}

	if len(comparison.Unchanged) != 1 {
		t.Errorf("Expected 1 unchanged operation, got %d", len(comparison.Unchanged))
	}

	if comparison.Added[0] != "POST /users" {
		t.Errorf("Expected POST /users to be added, got %s", comparison.Added[0])
	}
}

func TestCompareFingerprints_ModifiedOperation(t *testing.T) {
	spec1 := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"responses": {"200": {"description": "Success"}}
				}
			}
		}
	}`

	spec2 := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"parameters": [{"name": "limit", "in": "query"}],
					"responses": {"200": {"description": "Success"}}
				}
			}
		}
	}`

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "spec1.json")
	file2 := filepath.Join(tmpDir, "spec2.json")

	os.WriteFile(file1, []byte(spec1), 0644)
	os.WriteFile(file2, []byte(spec2), 0644)

	parsed1, _ := ParseSpecFile(file1)
	parsed2, _ := ParseSpecFile(file2)

	fp1, _ := CreateSpecFingerprint(file1, parsed1)
	fp2, _ := CreateSpecFingerprint(file2, parsed2)

	comparison := CompareFingerprints(fp1, fp2)

	if !comparison.HasChanges() {
		t.Error("Should detect changes")
	}

	if len(comparison.Modified) != 1 {
		t.Errorf("Expected 1 modified operation, got %d", len(comparison.Modified))
	}

	if comparison.Modified[0] != "GET /users" {
		t.Errorf("Expected GET /users to be modified, got %s", comparison.Modified[0])
	}
}

func TestCompareFingerprints_DeletedOperation(t *testing.T) {
	spec1 := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {"responses": {"200": {"description": "Success"}}},
				"post": {"responses": {"201": {"description": "Created"}}}
			}
		}
	}`

	spec2 := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {"responses": {"200": {"description": "Success"}}}
			}
		}
	}`

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "spec1.json")
	file2 := filepath.Join(tmpDir, "spec2.json")

	os.WriteFile(file1, []byte(spec1), 0644)
	os.WriteFile(file2, []byte(spec2), 0644)

	parsed1, _ := ParseSpecFile(file1)
	parsed2, _ := ParseSpecFile(file2)

	fp1, _ := CreateSpecFingerprint(file1, parsed1)
	fp2, _ := CreateSpecFingerprint(file2, parsed2)

	comparison := CompareFingerprints(fp1, fp2)

	if !comparison.HasChanges() {
		t.Error("Should detect changes")
	}

	if len(comparison.Deleted) != 1 {
		t.Errorf("Expected 1 deleted operation, got %d", len(comparison.Deleted))
	}

	if comparison.Deleted[0] != "POST /users" {
		t.Errorf("Expected POST /users to be deleted, got %s", comparison.Deleted[0])
	}
}

func TestCompareFingerprints_MultipleChanges(t *testing.T) {
	spec1 := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {"responses": {"200": {"description": "Success"}}},
				"post": {"responses": {"201": {"description": "Created"}}}
			},
			"/products": {
				"get": {"responses": {"200": {"description": "Success"}}}
			}
		}
	}`

	spec2 := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/users": {
				"get": {
					"parameters": [{"name": "id"}],
					"responses": {"200": {"description": "Success"}}
				}
			},
			"/products": {
				"get": {"responses": {"200": {"description": "Success"}}},
				"post": {"responses": {"201": {"description": "Created"}}}
			}
		}
	}`

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "spec1.json")
	file2 := filepath.Join(tmpDir, "spec2.json")

	os.WriteFile(file1, []byte(spec1), 0644)
	os.WriteFile(file2, []byte(spec2), 0644)

	parsed1, _ := ParseSpecFile(file1)
	parsed2, _ := ParseSpecFile(file2)

	fp1, _ := CreateSpecFingerprint(file1, parsed1)
	fp2, _ := CreateSpecFingerprint(file2, parsed2)

	comparison := CompareFingerprints(fp1, fp2)

	if !comparison.HasChanges() {
		t.Error("Should detect changes")
	}

	// Expect:
	// - 1 modified: GET /users (parameters added)
	// - 1 deleted: POST /users
	// - 1 added: POST /products
	// - 1 unchanged: GET /products

	if len(comparison.Modified) != 1 {
		t.Errorf("Expected 1 modified operation, got %d", len(comparison.Modified))
	}

	if len(comparison.Deleted) != 1 {
		t.Errorf("Expected 1 deleted operation, got %d", len(comparison.Deleted))
	}

	if len(comparison.Added) != 1 {
		t.Errorf("Expected 1 added operation, got %d", len(comparison.Added))
	}

	if len(comparison.Unchanged) != 1 {
		t.Errorf("Expected 1 unchanged operation, got %d", len(comparison.Unchanged))
	}
}

func TestHashOperation_Deterministic(t *testing.T) {
	op := OperationInfo{
		Path:   "/users",
		Method: "GET",
		Operation: &Operation{
			OperationID: "getUsers",
			Summary:     "Get users",
			Responses: map[string]interface{}{
				"200": map[string]interface{}{"description": "Success"},
			},
		},
	}

	// Hash same operation multiple times
	hash1, err1 := hashOperation(op)
	hash2, err2 := hashOperation(op)
	hash3, err3 := hashOperation(op)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("hashOperation() errors: %v, %v, %v", err1, err2, err3)
	}

	// All hashes should be identical
	if hash1 != hash2 || hash2 != hash3 {
		t.Error("hashOperation() should be deterministic")
	}
}

func TestHashOperation_IgnoresDescription(t *testing.T) {
	op1 := OperationInfo{
		Path:   "/users",
		Method: "GET",
		Operation: &Operation{
			OperationID: "getUsers",
			Summary:     "Get all users",
			Description: "This is a long description",
			Responses: map[string]interface{}{
				"200": map[string]interface{}{"description": "Success"},
			},
		},
	}

	op2 := OperationInfo{
		Path:   "/users",
		Method: "GET",
		Operation: &Operation{
			OperationID: "getUsers",
			Summary:     "Get all users",
			// Description changed/removed - should not affect hash
			Responses: map[string]interface{}{
				"200": map[string]interface{}{"description": "Success"},
			},
		},
	}

	hash1, _ := hashOperation(op1)
	hash2, _ := hashOperation(op2)

	// Hashes should be different because description affects signature
	// Actually, looking at the code, description is NOT included in canonical
	// So they should be the same
	if hash1 != hash2 {
		t.Error("Changing description should not affect hash")
	}
}

func TestFingerprintComparison_Summary(t *testing.T) {
	tests := []struct {
		name     string
		comp     *FingerprintComparison
		expected string
	}{
		{
			name: "no changes",
			comp: &FingerprintComparison{
				Unchanged: []string{"GET /users"},
			},
			expected: "No changes detected",
		},
		{
			name: "only additions",
			comp: &FingerprintComparison{
				Added:     []string{"POST /users", "GET /products"},
				Unchanged: []string{"GET /users"},
			},
			expected: "Changes: +2 added, ~0 modified, -0 deleted (1 unchanged)",
		},
		{
			name: "mixed changes",
			comp: &FingerprintComparison{
				Added:     []string{"POST /users"},
				Modified:  []string{"GET /users"},
				Deleted:   []string{"DELETE /users"},
				Unchanged: []string{"GET /products"},
			},
			expected: "Changes: +1 added, ~1 modified, -1 deleted (1 unchanged)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.comp.Summary()
			if result != tt.expected {
				t.Errorf("Summary() = %q, want %q", result, tt.expected)
			}
		})
	}
}
