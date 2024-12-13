package postprocess

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
)

// AddInternalClientsToAll adds the NewInternalClient function to all generated clients in the outputDir
func AddInternalClientsToAll(cfg config.Config) error {
	clientsDir := filepath.Join(cfg.OutputDir, "clients")

	// Get all client directories
	entries, err := os.ReadDir(clientsDir)
	if err != nil {
		return fmt.Errorf("error reading clients directory: %w", err)
	}

	// Track successful additions
	successCount := 0

	// Process each client directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		clientName := entry.Name()
		if err := AddInternalClientToSDK(cfg, clientName); err != nil {
			fmt.Printf("Warning: Failed to add NewInternalClient to %s: %v\n", clientName, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("Successfully added NewInternalClient function to %d clients\n", successCount)
	return nil
}

// AddInternalClient is kept for backward compatibility, now calls AddInternalClientToSDK
func AddInternalClient(outputDir string) error {
	cfg := config.Config{
		OutputDir: outputDir,
	}
	return AddInternalClientToSDK(cfg, "fundingsdk")
}

// AddInternalClientToSDK adds the NewInternalClient function to a specific SDK client
func AddInternalClientToSDK(cfg config.Config, clientName string) error {
	// Path to the generated client file
	clientFile := filepath.Join(cfg.OutputDir, "clients", clientName, "oas_client_gen.go")

	// Verify the file exists
	if _, err := os.Stat(clientFile); os.IsNotExist(err) {
		return fmt.Errorf("client file %s does not exist", clientFile)
	}

	// Read the original file content
	content, err := os.ReadFile(clientFile)
	if err != nil {
		return fmt.Errorf("error reading client file: %w", err)
	}

	// Check if NewInternalClient already exists
	if bytes.Contains(content, []byte("func NewInternalClient(")) {
		fmt.Printf("NewInternalClient function already exists in %s, skipping injection\n", clientName)
		return nil
	}

	// Check if the client has a security parameter in NewClient
	hasSecurityParam := bytes.Contains(content, []byte("func NewClient(serverURL string, sec SecuritySource"))

	// Determine the appropriate internal client function to inject
	var internalClientFunc string

	if hasSecurityParam {
		// Client has security parameter, add SecuritySourceOptional implementation
		internalClientFunc = `
// SecuritySourceOptional represents an optional security source implementation
// that returns empty security settings
type SecuritySourceOptional struct{}

// Bearer provides an empty bearer token for internal use
func (s *SecuritySourceOptional) Bearer(ctx context.Context, operationName string) (Bearer, error) {
	return Bearer{}, nil
}

// NewInternalClient initializes a Client that can be used for internal endpoints without security
// This is specifically for internal endpoints that don't require authentication
func NewInternalClient(serverURL string, opts ...ClientOption) (*Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	trimTrailingSlashes(u)

	c, err := newClientConfig(opts...).baseClient()
	if err != nil {
		return nil, err
	}
	
	return &Client{
		serverURL:  u,
		sec:        &SecuritySourceOptional{},
		baseClient: c,
	}, nil
}`
	} else {
		// Client doesn't have security parameter, create a simpler wrapper
		internalClientFunc = `
// NewInternalClient is an alias for NewClient for consistency with other SDKs
// This client doesn't require authentication for internal endpoints
func NewInternalClient(serverURL string, opts ...ClientOption) (*Client, error) {
	return NewClient(serverURL, opts...)
}`
	}

	// Find the insertion point - right after the NewClient function
	insertionPoint := []byte("func NewClient(")
	insertionPointIdx := 0

	// Find the end of the NewClient function
	if idx := findFunctionEnd(content, insertionPoint); idx > 0 {
		insertionPointIdx = idx
	} else {
		return fmt.Errorf("could not find insertion point")
	}

	// Insert our new function after the NewClient function
	newContent := append(
		content[:insertionPointIdx],
		append(
			[]byte(internalClientFunc),
			content[insertionPointIdx:]...,
		)...,
	)

	// Write the updated content back to the file
	if err := os.WriteFile(clientFile, newContent, 0644); err != nil {
		return fmt.Errorf("error writing to client file: %w", err)
	}

	fmt.Printf("Successfully added NewInternalClient function to %s\n", clientName)
	return nil
}

// findFunctionEnd finds the end of a function, starting from the position of the function declaration
func findFunctionEnd(content []byte, declaration []byte) int {
	// Find the declaration
	idx := 0
	for i := 0; i <= len(content)-len(declaration); i++ {
		match := true
		for j := 0; j < len(declaration); j++ {
			if content[i+j] != declaration[j] {
				match = false
				break
			}
		}
		if match {
			idx = i
			break
		}
	}

	if idx == 0 {
		return -1
	}

	// Find the opening bracket {
	openBracketIdx := idx
	for openBracketIdx < len(content) {
		if content[openBracketIdx] == '{' {
			break
		}
		openBracketIdx++
	}

	// Count brackets to find the end of the function
	bracketCount := 1
	endIdx := openBracketIdx + 1

	for endIdx < len(content) && bracketCount > 0 {
		if content[endIdx] == '{' {
			bracketCount++
		} else if content[endIdx] == '}' {
			bracketCount--
		}
		endIdx++
	}

	return endIdx
}
