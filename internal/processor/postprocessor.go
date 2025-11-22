package processor

import (
	"context"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/postprocessor"
)

var (
	// defaultPostProcessorChain is the default chain of post-processors
	// Can be overridden for testing or customization
	defaultPostProcessorChain *postprocessor.Chain
)

func init() {
	// Initialize default post-processor chain
	defaultPostProcessorChain = postprocessor.NewChain()

	// Add internal client generator
	defaultPostProcessorChain.Add(postprocessor.NewInternalClientProcessor())

	// Add Go formatter (without simplify for compatibility)
	defaultPostProcessorChain.Add(postprocessor.NewFormatterProcessor(false))
}

// ApplyPostProcessors applies post-processing steps to the generated client code.
// This uses the configured post-processor chain.
func ApplyPostProcessors(ctx context.Context, clientPath, serviceName, specPath string) error {
	spec := postprocessor.ProcessSpec{
		ClientPath:  clientPath,
		ServiceName: serviceName,
		SpecPath:    specPath,
		PackageName: serviceName,
	}

	return defaultPostProcessorChain.Process(ctx, spec)
}

// SetPostProcessorChain allows overriding the default post-processor chain
func SetPostProcessorChain(chain *postprocessor.Chain) {
	if chain != nil {
		defaultPostProcessorChain = chain
	}
}

// GetPostProcessorChain returns the current post-processor chain
func GetPostProcessorChain() *postprocessor.Chain {
	return defaultPostProcessorChain
}
