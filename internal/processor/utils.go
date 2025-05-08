package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// compileServiceRegex creates a regex for filtering services.
func compileServiceRegex(targetServices string) (*regexp.Regexp, error) {
	if targetServices == "" {
		return regexp.MustCompile(".*"), nil
	}

	regex, err := regexp.Compile(targetServices)
	if err != nil {
		return nil, fmt.Errorf("invalid target services pattern: %w", err)
	}

	return regex, nil
}

// normalizeServiceName converts a service directory name to a valid Go package name.
// For example: "funding-server-sdk" -> "funding"
func normalizeServiceName(service string) string {
	// Remove common suffixes in a single pass
	suffixes := []string{"-server-sdk", "-sdk"}
	name := service
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
			break // Only remove one suffix
		}
	}

	// Split by hyphens and process each part
	parts := strings.Split(name, "-")
	for i, part := range parts {
		part = strings.ToLower(part)

		// Special handling for abbreviations
		switch part {
		case "api", "sdk", "id":
			parts[i] = strings.ToUpper(part)
		case "": // Handle empty parts that might result from splitting
			continue
		default:
			if i == 0 {
				// Keep the first part lowercase for package name conventions
				parts[i] = part
			} else if len(part) > 0 {
				// Title case for non-first parts (capitalize first letter)
				parts[i] = strings.ToUpper(part[:1]) + part[1:]
			}
		}
	}

	return strings.Join(parts, "")
}

// cleanDirectory removes all files in the specified directory.
// It returns an error if the directory doesn't exist or if there's an issue removing files.
func cleanDirectory(dir string) error {
	// Check if directory exists
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}
	if err != nil {
		return fmt.Errorf("failed to access directory %s: %w", dir, err)
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Remove each file
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			// Recursively clean subdirectories
			if err := cleanDirectory(path); err != nil {
				return err
			}
			// Remove the now-empty directory
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove directory %s: %w", path, err)
			}
		} else {
			// Remove file
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove file %s: %w", path, err)
			}
		}
	}

	return nil
}
