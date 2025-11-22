package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	// Cached paths calculated once at startup
	repositoryRoot string
)

func init() {
	var err error

	// Find repository root by looking for go.mod
	repositoryRoot, err = findRepositoryRoot()
	if err != nil {
		// Fallback: try to use current working directory
		if cwd, cwdErr := os.Getwd(); cwdErr == nil {
			repositoryRoot = cwd
		} else {
			panic(fmt.Sprintf("failed to determine repository root: %v (cwd error: %v)", err, cwdErr))
		}
	}
}

// findRepositoryRoot walks up the directory tree to find go.mod
func findRepositoryRoot() (string, error) {
	// Start from the directory of this source file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}

	dir := filepath.Dir(filename)

	// Walk up the directory tree
	for {
		// Check if go.mod exists in current directory
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			return "", fmt.Errorf("repository root not found (no go.mod)")
		}
		dir = parent
	}
}

// GetRepositoryRoot returns the absolute path to repository root
func GetRepositoryRoot() string {
	return repositoryRoot
}

// GetOgenConfigPath returns the absolute path to ogen.yml
func GetOgenConfigPath() string {
	return filepath.Join(repositoryRoot, "ogen.yml")
}

// GetTemplatesDir returns the absolute path to templates directory
func GetTemplatesDir() string {
	return filepath.Join(repositoryRoot, "resources", "templates")
}

// GetInternalClientTemplatePath returns path to internal client template
func GetInternalClientTemplatePath() string {
	return filepath.Join(GetTemplatesDir(), "internal_client.tmpl")
}

// GetConfigPath returns the absolute path to application.yml
func GetConfigPath() string {
	return filepath.Join(repositoryRoot, "resources", "application.yml")
}

// GetResourcesDir returns the absolute path to resources directory
func GetResourcesDir() string {
	return filepath.Join(repositoryRoot, "resources")
}

// EnsurePathExists verifies that a path exists and is accessible
func EnsurePathExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	} else if err != nil {
		return fmt.Errorf("cannot access path %s: %w", path, err)
	}
	return nil
}

// EnsureDirectoryWritable checks if directory is writable
func EnsureDirectoryWritable(dir string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Try to create a temporary file
	testFile := filepath.Join(dir, fmt.Sprintf(".write_test_%d", os.Getpid()))
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("directory not writable: %s: %w", dir, err)
	}
	f.Close()
	os.Remove(testFile)
	return nil
}

// MakeAbsolutePath converts a relative path to absolute based on repository root
// If the path is already absolute, returns it unchanged
func MakeAbsolutePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(repositoryRoot, p)
}
