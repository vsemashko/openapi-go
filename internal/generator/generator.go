package generator

import (
	"context"
	"fmt"
)

// Generator defines the interface for OpenAPI client code generators.
// Implementations of this interface are responsible for generating Go client
// code from OpenAPI specifications.
type Generator interface {
	// Name returns the name of the generator (e.g., "ogen", "oapi-codegen")
	Name() string

	// Version returns the version of the generator being used
	Version() string

	// EnsureInstalled checks if the generator is available and installs it if needed
	EnsureInstalled(ctx context.Context) error

	// Generate generates client code from an OpenAPI spec
	// Parameters:
	//   - ctx: Context for cancellation
	//   - spec: GenerateSpec containing all generation parameters
	// Returns an error if generation fails
	Generate(ctx context.Context, spec GenerateSpec) error

	// IsInstalled checks if the generator is currently installed and ready to use
	IsInstalled() bool
}

// GenerateSpec contains all parameters needed for code generation
type GenerateSpec struct {
	// SpecPath is the absolute path to the OpenAPI specification file
	SpecPath string

	// OutputDir is the directory where generated code should be written
	OutputDir string

	// PackageName is the Go package name for the generated code
	PackageName string

	// ConfigPath is the optional path to generator-specific configuration
	ConfigPath string

	// Clean indicates whether to clean the output directory before generation
	Clean bool
}

// Registry manages available generators and provides a way to select and use them
type Registry struct {
	generators       map[string]Generator
	defaultGenerator string
}

// NewRegistry creates a new generator registry
func NewRegistry() *Registry {
	return &Registry{
		generators:       make(map[string]Generator),
		defaultGenerator: "",
	}
}

// Register adds a generator to the registry
func (r *Registry) Register(gen Generator) error {
	if gen == nil {
		return fmt.Errorf("cannot register nil generator")
	}

	name := gen.Name()
	if name == "" {
		return fmt.Errorf("generator name cannot be empty")
	}

	if _, exists := r.generators[name]; exists {
		return fmt.Errorf("generator %q is already registered", name)
	}

	r.generators[name] = gen

	// Set as default if it's the first generator
	if r.defaultGenerator == "" {
		r.defaultGenerator = name
	}

	return nil
}

// Get retrieves a generator by name
func (r *Registry) Get(name string) (Generator, error) {
	gen, exists := r.generators[name]
	if !exists {
		return nil, fmt.Errorf("generator %q not found", name)
	}
	return gen, nil
}

// GetDefault returns the default generator
func (r *Registry) GetDefault() (Generator, error) {
	if r.defaultGenerator == "" {
		return nil, fmt.Errorf("no default generator set")
	}
	return r.Get(r.defaultGenerator)
}

// SetDefault sets the default generator by name
func (r *Registry) SetDefault(name string) error {
	if _, exists := r.generators[name]; !exists {
		return fmt.Errorf("generator %q not found", name)
	}
	r.defaultGenerator = name
	return nil
}

// List returns the names of all registered generators
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.generators))
	for name := range r.generators {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered generators
func (r *Registry) Count() int {
	return len(r.generators)
}

// Clear removes all registered generators
func (r *Registry) Clear() {
	r.generators = make(map[string]Generator)
	r.defaultGenerator = ""
}
