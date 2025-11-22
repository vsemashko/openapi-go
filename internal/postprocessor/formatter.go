package postprocessor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// FormatterProcessor formats generated Go code using gofmt
type FormatterProcessor struct {
	// If true, will use gofmt -s (simplify code)
	simplify bool
}

// NewFormatterProcessor creates a new formatter processor
func NewFormatterProcessor(simplify bool) *FormatterProcessor {
	return &FormatterProcessor{
		simplify: simplify,
	}
}

// Name returns the processor name
func (p *FormatterProcessor) Name() string {
	return "GoFormatter"
}

// Process formats all Go files in the client directory
func (p *FormatterProcessor) Process(ctx context.Context, spec ProcessSpec) error {
	// Find all .go files in the client directory
	goFiles, err := p.findGoFiles(spec.ClientPath)
	if err != nil {
		return fmt.Errorf("failed to find Go files: %w", err)
	}

	if len(goFiles) == 0 {
		log.Printf("No Go files found to format in %s", spec.ClientPath)
		return nil
	}

	log.Printf("Formatting %d Go file(s) in %s...", len(goFiles), spec.ClientPath)

	// Build gofmt command
	args := []string{"-w"}
	if p.simplify {
		args = append(args, "-s")
	}
	args = append(args, goFiles...)

	// Run gofmt
	cmd := exec.CommandContext(ctx, "gofmt", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gofmt failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		log.Printf("gofmt output: %s", string(output))
	}

	log.Printf("Successfully formatted %d Go file(s)", len(goFiles))
	return nil
}

// findGoFiles recursively finds all .go files in the directory
func (p *FormatterProcessor) findGoFiles(dir string) ([]string, error) {
	var goFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a Go file
		if filepath.Ext(path) == ".go" {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return goFiles, nil
}
