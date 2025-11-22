package postprocessor

import (
	"context"
	"fmt"
	"log"
)

// PostProcessor defines the interface for post-processing generated client code.
// Post-processors can add additional files, modify generated code, format output, etc.
type PostProcessor interface {
	// Name returns the name of the post-processor
	Name() string

	// Process applies the post-processing step
	// Parameters:
	//   - ctx: Context for cancellation
	//   - spec: ProcessSpec containing all necessary information
	// Returns an error if processing fails
	Process(ctx context.Context, spec ProcessSpec) error
}

// ProcessSpec contains all parameters needed for post-processing
type ProcessSpec struct {
	// ClientPath is the directory containing the generated client code
	ClientPath string

	// ServiceName is the name of the service (e.g., "funding", "holidays")
	ServiceName string

	// SpecPath is the path to the original OpenAPI specification
	SpecPath string

	// PackageName is the Go package name for the generated client
	PackageName string
}

// Chain manages an ordered list of post-processors and executes them sequentially
type Chain struct {
	processors []PostProcessor
}

// NewChain creates a new post-processor chain
func NewChain() *Chain {
	return &Chain{
		processors: make([]PostProcessor, 0),
	}
}

// Add appends a post-processor to the chain
func (c *Chain) Add(processor PostProcessor) error {
	if processor == nil {
		return fmt.Errorf("cannot add nil post-processor")
	}

	c.processors = append(c.processors, processor)
	return nil
}

// Process executes all post-processors in the chain sequentially
func (c *Chain) Process(ctx context.Context, spec ProcessSpec) error {
	if len(c.processors) == 0 {
		log.Printf("No post-processors configured, skipping post-processing")
		return nil
	}

	log.Printf("Running %d post-processor(s) for %s...", len(c.processors), spec.ServiceName)

	for i, processor := range c.processors {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("post-processing cancelled: %w", ctx.Err())
		default:
		}

		log.Printf("  [%d/%d] Running %s...", i+1, len(c.processors), processor.Name())

		if err := processor.Process(ctx, spec); err != nil {
			return fmt.Errorf("post-processor %q failed: %w", processor.Name(), err)
		}

		log.Printf("  [%d/%d] ✓ %s completed", i+1, len(c.processors), processor.Name())
	}

	log.Printf("All post-processors completed successfully for %s", spec.ServiceName)
	return nil
}

// Count returns the number of post-processors in the chain
func (c *Chain) Count() int {
	return len(c.processors)
}

// Clear removes all post-processors from the chain
func (c *Chain) Clear() {
	c.processors = make([]PostProcessor, 0)
}

// List returns the names of all post-processors in the chain
func (c *Chain) List() []string {
	names := make([]string, len(c.processors))
	for i, p := range c.processors {
		names[i] = p.Name()
	}
	return names
}
