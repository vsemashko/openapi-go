package processor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/postprocessor"
)

func TestApplyPostProcessors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(string) (clientPath, serviceName, specPath string)
		wantErr bool
	}{
		{
			name: "valid spec with security",
			setup: func(tmpDir string) (string, string, string) {
				clientPath := filepath.Join(tmpDir, "client")
				os.MkdirAll(clientPath, 0755)

				specPath := filepath.Join(tmpDir, "spec.json")
				spec := `{
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
				}`
				os.WriteFile(specPath, []byte(spec), 0644)

				// Create a sample Go file for formatting
				os.WriteFile(filepath.Join(clientPath, "test.go"), []byte("package test\n\nfunc Test() {}\n"), 0644)

				return clientPath, "testservice", specPath
			},
			wantErr: false,
		},
		{
			name: "valid spec without security",
			setup: func(tmpDir string) (string, string, string) {
				clientPath := filepath.Join(tmpDir, "client")
				os.MkdirAll(clientPath, 0755)

				specPath := filepath.Join(tmpDir, "spec.json")
				spec := `{
					"openapi": "3.0.0",
					"info": {"title": "Test", "version": "1.0"},
					"paths": {}
				}`
				os.WriteFile(specPath, []byte(spec), 0644)

				// Create a sample Go file for formatting
				os.WriteFile(filepath.Join(clientPath, "test.go"), []byte("package test\n\nfunc Test() {}\n"), 0644)

				return clientPath, "testservice", specPath
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			clientPath, serviceName, specPath := tt.setup(tmpDir)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := ApplyPostProcessors(ctx, clientPath, serviceName, specPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyPostProcessors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If successful, verify internal client file was created
			if err == nil {
				internalClientPath := filepath.Join(clientPath, "oas_internal_client_gen.go")
				if _, err := os.Stat(internalClientPath); os.IsNotExist(err) {
					t.Errorf("Expected internal client file not created: %s", internalClientPath)
				}
			}
		})
	}
}

func TestSetPostProcessorChain(t *testing.T) {
	// Save original chain
	originalChain := GetPostProcessorChain()
	defer SetPostProcessorChain(originalChain)

	// Create a new chain
	newChain := postprocessor.NewChain()
	newChain.Add(postprocessor.NewInternalClientProcessor())

	SetPostProcessorChain(newChain)

	retrievedChain := GetPostProcessorChain()
	if retrievedChain != newChain {
		t.Error("SetPostProcessorChain() did not update the chain")
	}

	// Test setting nil (should not update)
	SetPostProcessorChain(nil)
	if GetPostProcessorChain() != newChain {
		t.Error("SetPostProcessorChain(nil) should not update the chain")
	}
}

func TestGetPostProcessorChain(t *testing.T) {
	chain := GetPostProcessorChain()

	if chain == nil {
		t.Fatal("GetPostProcessorChain() returned nil")
	}

	// Verify default chain has expected processors
	list := chain.List()
	if len(list) < 2 {
		t.Errorf("Default chain has %d processors, expected at least 2", len(list))
	}

	// Verify it includes InternalClientGenerator and GoFormatter
	foundInternal := false
	foundFormatter := false

	for _, name := range list {
		if name == "InternalClientGenerator" {
			foundInternal = true
		}
		if name == "GoFormatter" {
			foundFormatter = true
		}
	}

	if !foundInternal {
		t.Error("Default chain should include InternalClientGenerator")
	}

	if !foundFormatter {
		t.Error("Default chain should include GoFormatter")
	}
}

func TestApplyPostProcessorsWithCustomChain(t *testing.T) {
	// Save original chain
	originalChain := GetPostProcessorChain()
	defer SetPostProcessorChain(originalChain)

	// Create a custom chain with only the internal client processor
	customChain := postprocessor.NewChain()
	customChain.Add(postprocessor.NewInternalClientProcessor())
	SetPostProcessorChain(customChain)

	tmpDir := t.TempDir()
	clientPath := filepath.Join(tmpDir, "client")
	os.MkdirAll(clientPath, 0755)

	specPath := filepath.Join(tmpDir, "spec.json")
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0"},
		"paths": {}
	}`
	os.WriteFile(specPath, []byte(spec), 0644)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ApplyPostProcessors(ctx, clientPath, "testservice", specPath)
	if err != nil {
		t.Errorf("ApplyPostProcessors() with custom chain error = %v", err)
	}

	// Verify internal client file was created
	internalClientPath := filepath.Join(clientPath, "oas_internal_client_gen.go")
	if _, err := os.Stat(internalClientPath); os.IsNotExist(err) {
		t.Error("Expected internal client file was not created")
	}
}
