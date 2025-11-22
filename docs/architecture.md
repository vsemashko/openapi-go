# OpenAPI Go Generator - Architecture Documentation

## Table of Contents

1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Core Components](#core-components)
4. [Data Flow](#data-flow)
5. [Design Patterns](#design-patterns)
6. [Concurrency Model](#concurrency-model)
7. [Caching Strategy](#caching-strategy)
8. [Error Handling](#error-handling)
9. [Extension Points](#extension-points)

## Overview

The OpenAPI Go Generator is a tool that automates the generation of Go client code from OpenAPI 3.0 specifications. The architecture follows Go best practices with a focus on:

- **Modularity**: Clear separation of concerns with focused packages
- **Testability**: High test coverage with minimal mocking
- **Performance**: Parallel processing with smart caching
- **Extensibility**: Plugin-based post-processor chain
- **Observability**: Structured logging and metrics collection

### High-Level Architecture

```
┌────────────────────────────────────────────────────────────┐
│                         main.go                            │
│  - Load configuration                                      │
│  - Initialize structured logger                            │
│  - Set up context with cancellation                        │
│  - Call processor.ProcessOpenAPISpecs()                    │
└────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌────────────────────────────────────────────────────────────┐
│                  internal/processor                        │
│  - Find OpenAPI specs                                      │
│  - Initialize cache and metrics                            │
│  - Generate clients (parallel or sequential)               │
│  - Apply post-processors                                   │
│  - Export metrics                                          │
└────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   cache     │     │   generator  │     │ post-processor  │
│  - SHA256   │     │  - ogen      │     │  - internal     │
│  - IsValid  │     │  - Generate  │     │    client gen   │
│  - Set      │     │  - Install   │     │  - Go formatter │
└─────────────┘     └──────────────┘     └─────────────────┘
```

## System Architecture

### Package Structure

```
openapi-go/
├── main.go                  # Entry point
├── internal/                # Internal packages
│   ├── cache/              # Caching system
│   │   ├── cache.go        # Cache implementation
│   │   └── cache_test.go   # Cache tests
│   ├── config/             # Configuration management
│   │   ├── config.go       # Config loading
│   │   ├── config_logging.go  # Config logging
│   │   └── config_test.go  # Config tests
│   ├── generator/          # Generator interface
│   │   ├── generator.go    # Generator interface
│   │   ├── ogen.go         # Ogen implementation
│   │   └── ogen_test.go    # Ogen tests
│   ├── logger/             # Structured logging
│   │   ├── logger.go       # Logger implementation
│   │   └── logger_test.go  # Logger tests
│   ├── metrics/            # Metrics collection
│   │   ├── metrics.go      # Metrics collector
│   │   └── metrics_test.go # Metrics tests
│   ├── paths/              # Path utilities
│   │   ├── paths.go        # Path resolution
│   │   └── paths_test.go   # Path tests
│   ├── postprocessor/      # Post-processing chain
│   │   ├── chain.go        # Chain management
│   │   ├── formatter.go    # Code formatter
│   │   ├── internal_client.go  # Internal client gen
│   │   └── *_test.go       # Tests
│   ├── processor/          # Main processing logic
│   │   ├── processor.go    # Orchestration
│   │   ├── utils.go        # Helper functions
│   │   └── *_test.go       # Tests
│   ├── spec/               # OpenAPI spec utilities
│   │   ├── spec.go         # Spec operations
│   │   └── spec_test.go    # Spec tests
│   └── worker/             # Parallel worker pool
│       ├── pool.go         # Worker pool
│       └── pool_test.go    # Pool tests
├── generated/              # Generated client code
│   └── clients/            # SDK clients
├── resources/              # Configuration files
│   └── application.yml     # Default config
└── docs/                   # Documentation
```

### Dependency Graph

```
main.go
  ├──> config
  ├──> logger
  └──> processor
         ├──> config
         ├──> cache
         ├──> generator
         │      └──> spec
         ├──> metrics
         ├──> worker
         ├──> postprocessor
         └──> paths
```

**Key principles**:
- No circular dependencies
- Internal packages not exported
- Clear dependency direction (top to bottom)

## Core Components

### 1. Configuration (internal/config)

**Purpose**: Load and manage configuration from YAML and environment variables.

**Key Files**:
- `config.go`: Configuration struct and loading logic
- `config_logging.go`: Configuration logging with structured/fallback support

**Interface**:
```go
type Config struct {
    SpecsDir         string   // Directory containing OpenAPI specs
    OutputDir        string   // Output directory for generated clients
    TargetServices   string   // Regex pattern to filter services
    WorkerCount      int      // Number of parallel workers
    ContinueOnError  bool     // Continue processing on error
    EnableCache      bool     // Enable/disable caching
    CacheDir         string   // Cache directory location
    SpecFilePatterns []string // Spec file patterns to match
    LogLevel         string   // Logging level
    LogFormat        string   // Logging format
}

func LoadConfig() (Config, error)
func LogConfiguration(cfg Config, optionalLogger ...interface{})
```

**Configuration precedence**:
1. Default values (hardcoded)
2. YAML file (`resources/application.yml`)
3. Environment variables (highest priority)

### 2. Logger (internal/logger)

**Purpose**: Provide structured logging with configurable levels and formats.

**Key Features**:
- Built on Go's standard `log/slog`
- JSON and text formats
- Debug, Info, Warn, Error levels
- Method chaining for context

**Interface**:
```go
type Logger struct {
    *slog.Logger
}

type Config struct {
    Level  string      // Log level: debug, info, warn, error
    Format string      // Format: json, text
    Output io.Writer   // Output destination
}

func New(cfg Config) *Logger
func (l *Logger) Debug(msg string, args ...any)
func (l *Logger) Info(msg string, args ...any)
func (l *Logger) Warn(msg string, args ...any)
func (l *Logger) Error(msg string, args ...any)
func (l *Logger) WithFields(fields map[string]any) *Logger
```

### 3. Metrics (internal/metrics)

**Purpose**: Collect and export generation performance metrics.

**Key Features**:
- Per-spec timing and success tracking
- Aggregated statistics (success rate, cache hit rate)
- JSON export for analysis
- Thread-safe concurrent recording

**Interface**:
```go
type Metrics struct {
    TotalSpecs        int
    SuccessfulSpecs   int
    FailedSpecs       int
    CachedSpecs       int
    TotalDurationMs   int64
    AverageDurationMs int64
    StartTime         time.Time
    EndTime           time.Time
    SpecMetrics       []SpecMetric
}

type SpecMetric struct {
    SpecPath    string
    ServiceName string
    Success     bool
    Cached      bool
    DurationMs  int64
    Error       string
    GeneratedAt time.Time
}

type Collector struct { /* private */ }

func NewCollector() *Collector
func (c *Collector) RecordSpec(metric SpecMetric)
func (c *Collector) Finalize()
func (c *Collector) Export(path string) error
func (c *Collector) Summary() string
func (c *Collector) SuccessRate() float64
func (c *Collector) CacheHitRate() float64
```

### 4. Cache (internal/cache)

**Purpose**: Implement SHA256-based caching to skip regeneration of unchanged specs.

**Key Features**:
- SHA256 hashing of spec files
- Generator version tracking
- Automatic cache invalidation
- Cache pruning of invalid entries

**Interface**:
```go
type Cache struct { /* private */ }

type Config struct {
    CacheDir string  // Cache directory location
}

func NewCache(cfg Config) (*Cache, error)
func (c *Cache) IsValid(specPath, generatorVersion string) (bool, error)
func (c *Cache) Set(specPath, clientPath, serviceName, generatorVersion string) error
func (c *Cache) PruneInvalid() (int, error)
```

**Cache entry format** (`.openapi-cache/{sha256}.json`):
```json
{
  "spec_path": "/path/to/spec.json",
  "spec_hash": "abc123...",
  "client_path": "/path/to/generated/client",
  "service_name": "funding",
  "generator_version": "v1.14.0",
  "cached_at": "2025-11-22T10:00:00Z"
}
```

### 5. Generator (internal/generator)

**Purpose**: Abstract interface for different OpenAPI code generators.

**Key Features**:
- Generator abstraction (currently: ogen)
- Version management
- Automatic installation
- Configurable generation options

**Interface**:
```go
type Generator interface {
    Name() string
    Version() string
    Generate(ctx context.Context, spec GenerateSpec) error
    IsInstalled() bool
    EnsureInstalled(ctx context.Context) error
}

type GenerateSpec struct {
    SpecPath    string  // Path to OpenAPI spec
    OutputDir   string  // Output directory
    PackageName string  // Go package name
    ConfigPath  string  // Generator config path
    Clean       bool    // Clean output directory
}
```

**Current implementation**: Ogen (v1.14.0)
- OpenAPI 3.0.x support
- High-performance code generation
- Type-safe client code

### 6. Worker Pool (internal/worker)

**Purpose**: Manage parallel processing of multiple specs.

**Key Features**:
- Configurable worker count
- Task queue with buffering
- Context-aware cancellation
- Result collection

**Interface**:
```go
type Pool struct { /* private */ }

type Config struct {
    WorkerCount   int  // Number of parallel workers
    TaskQueueSize int  // Task queue buffer size
}

type Task struct {
    ID      string  // Task identifier
    Execute func(context.Context) error  // Task function
}

type Result struct {
    TaskID string  // Task identifier
    Error  error   // Execution error (nil if success)
}

func NewPool(cfg Config) *Pool
func (p *Pool) ProcessBatch(ctx context.Context, tasks []Task) ([]Result, error)
```

### 7. Post-Processor Chain (internal/postprocessor)

**Purpose**: Apply post-processing transformations to generated code.

**Key Features**:
- Chain of Responsibility pattern
- Pluggable processors
- Sequential execution
- Error propagation

**Interface**:
```go
type PostProcessor interface {
    Name() string
    Process(ctx context.Context, clientPath, serviceName, specPath string) error
}

func ApplyPostProcessors(ctx context.Context, clientPath, serviceName, specPath string) error
func SetPostProcessorChain(chain []PostProcessor)
func GetPostProcessorChain() []PostProcessor
```

**Built-in processors**:
1. **InternalClientGenerator**: Generates `NewInternalClient` function for specs with security
2. **GoFormatter**: Formats generated code with `gofmt`

### 8. Processor (internal/processor)

**Purpose**: Orchestrate the entire generation process.

**Key Responsibilities**:
- Find OpenAPI specs
- Initialize subsystems (cache, metrics)
- Coordinate parallel/sequential generation
- Apply post-processors
- Export results

**Main Functions**:
```go
func ProcessOpenAPISpecs(ctx context.Context, cfg config.Config, optionalLogger ...interface{}) error
func findOpenAPISpecs(specsDir, targetServices string, specFilePatterns []string) ([]string, error)
func generateClients(ctx context.Context, specs []string, outputDir string, continueOnError bool, workerCount int, specCache *cache.Cache, metricsCollector *metrics.Collector) (*ProcessingResult, error)
func generateClientForSpec(ctx context.Context, specPath, serviceName, folderName, outputDir string) error
```

## Data Flow

### Complete Generation Flow

```
┌────────────────────┐
│  1. Load Config    │
│  - YAML file       │
│  - Env vars        │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│  2. Init Logger    │
│  - Level & Format  │
│  - Output stream   │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│  3. Find Specs     │
│  - Walk directory  │
│  - Match patterns  │
│  - Filter services │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│  4. Init Systems   │
│  - Cache           │
│  - Metrics         │
│  - Worker pool     │
└─────────┬──────────┘
          │
          ▼
┌─────────────────────────┐
│  5. Process Specs       │
│  ┌─────────────────┐    │
│  │ For each spec:  │    │
│  │ - Check cache   │    │
│  │ - Generate      │    │
│  │ - Post-process  │    │
│  │ - Update cache  │    │
│  │ - Record metric │    │
│  └─────────────────┘    │
└─────────┬───────────────┘
          │
          ▼
┌────────────────────┐
│  6. Export Results │
│  - Metrics JSON    │
│  - Summary logs    │
└────────────────────┘
```

### Parallel Processing Flow

```
                        ┌─────────────┐
                        │ Spec Queue  │
                        └──────┬──────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
         ▼                     ▼                     ▼
    ┌─────────┐          ┌─────────┐          ┌─────────┐
    │Worker 1 │          │Worker 2 │          │Worker N │
    │         │          │         │          │         │
    │- Cache? │          │- Cache? │          │- Cache? │
    │- Gen    │          │- Gen    │          │- Gen    │
    │- Post   │          │- Post   │          │- Post   │
    │- Metric │          │- Metric │          │- Metric │
    └────┬────┘          └────┬────┘          └────┬────┘
         │                    │                     │
         └────────────────────┼─────────────────────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │ Results Channel │
                     └─────────────────┘
```

## Design Patterns

### 1. Dependency Injection

Dependencies are passed explicitly rather than using globals.

```go
// Good: Dependencies injected
func ProcessOpenAPISpecs(ctx context.Context, cfg config.Config, logger *logger.Logger) error {
    cache, _ := cache.NewCache(cache.Config{CacheDir: cfg.CacheDir})
    metrics := metrics.NewCollector()
    // ...
}

// Avoid: Global dependencies
var globalCache *cache.Cache
```

### 2. Chain of Responsibility

Post-processors form a chain where each can process or pass to the next.

```go
type PostProcessor interface {
    Name() string
    Process(ctx context.Context, clientPath, serviceName, specPath string) error
}

var defaultChain = []PostProcessor{
    &InternalClientGenerator{},
    &GoFormatter{},
}

func ApplyPostProcessors(...) error {
    for _, processor := range defaultChain {
        if err := processor.Process(...); err != nil {
            return fmt.Errorf("%s failed: %w", processor.Name(), err)
        }
    }
    return nil
}
```

### 3. Strategy Pattern

Different generators can be swapped via the Generator interface.

```go
type Generator interface {
    Generate(ctx context.Context, spec GenerateSpec) error
}

var defaultGenerator Generator = &OgenGenerator{}

// Can be swapped for testing or alternative generators
func SetGenerator(gen Generator) {
    defaultGenerator = gen
}
```

### 4. Worker Pool Pattern

Parallel processing using a pool of workers.

```go
type Pool struct {
    workerCount int
    taskQueue   chan Task
    resultQueue chan Result
}

func (p *Pool) ProcessBatch(ctx context.Context, tasks []Task) ([]Result, error) {
    // Create workers
    for i := 0; i < p.workerCount; i++ {
        go p.worker(ctx)
    }

    // Submit tasks
    for _, task := range tasks {
        p.taskQueue <- task
    }

    // Collect results
    // ...
}
```

### 5. Template Method Pattern

Base generation flow with customizable steps.

```go
func generateClientForSpec(...) error {
    // 1. Setup
    if err := os.MkdirAll(clientPath, os.ModePerm); err != nil {
        return err
    }

    // 2. Clean
    if err := cleanDirectory(clientPath); err != nil {
        return err
    }

    // 3. Generate (customizable via generator interface)
    if err := runGenerator(ctx, serviceName, specPath, clientPath); err != nil {
        return err
    }

    // 4. Post-process (customizable via post-processor chain)
    if err := ApplyPostProcessors(ctx, clientPath, serviceName, specPath); err != nil {
        return err
    }

    return nil
}
```

## Concurrency Model

### Thread Safety

**Thread-safe components**:
- `metrics.Collector`: Uses `sync.RWMutex`
- `cache.Cache`: Uses `sync.RWMutex`
- `worker.Pool`: Uses channels for communication

**Not thread-safe** (by design):
- `config.Config`: Immutable after loading
- `generator.Generator`: Used per-task, not shared

### Synchronization

**Metrics recording**:
```go
type Metrics struct {
    mu sync.RWMutex
    // fields
}

func (m *Metrics) RecordSpec(metric SpecMetric) {
    m.mu.Lock()
    defer m.mu.Unlock()
    // Safe concurrent access
}
```

**Worker pool results**:
```go
// Use channels for goroutine communication
resultQueue := make(chan Result, len(tasks))

// Workers send results
resultQueue <- Result{TaskID: task.ID, Error: err}

// Main goroutine collects results
for i := 0; i < len(tasks); i++ {
    result := <-resultQueue
    results = append(results, result)
}
```

## Caching Strategy

### Cache Key

```
SHA256(spec file content) + generator version
```

### Cache Validation

1. **Compute spec hash**: `SHA256(spec content)`
2. **Check cache entry exists**: `.openapi-cache/{hash}.json`
3. **Validate generator version**: Must match current version
4. **Validate client output exists**: Check generated client directory

### Cache Invalidation

Automatically invalidated when:
- Spec content changes (different hash)
- Generator version changes
- Generated client directory missing
- Cache entry corrupted

### Cache Performance

**Cache hit** (typical):
- Check cache: ~1ms
- Skip generation: ~3-5 seconds saved per spec

**Cache miss**:
- Generate client: ~3-5 seconds
- Update cache: ~1ms

## Error Handling

### Error Wrapping

All errors are wrapped with context:

```go
if err != nil {
    return fmt.Errorf("failed to process spec %s: %w", specPath, err)
}
```

This creates error chains:
```
failed to process spec ./funding/openapi.json:
  generation failed for funding:
    failed to execute ogen:
      exit status 1
```

### Error Recovery

**Non-critical errors** (warnings):
- Cache check failures
- Cache update failures
- Metrics export failures

**Critical errors** (fatal):
- Spec not found
- Generation failures
- Output directory creation failures

### Context Cancellation

All long-running operations respect context:

```go
func ProcessOpenAPISpecs(ctx context.Context, ...) error {
    select {
    case <-ctx.Done():
        return fmt.Errorf("cancelled: %w", ctx.Err())
    default:
        // Continue processing
    }
}
```

## Extension Points

### Adding a New Generator

1. Implement `Generator` interface
2. Register in processor
3. Add configuration option

### Adding a Post-Processor

1. Implement `PostProcessor` interface
2. Add to default chain
3. Add tests

### Adding Configuration Options

1. Add field to `Config` struct
2. Add YAML parsing
3. Add environment variable support
4. Update documentation

### Adding Metrics

1. Add field to `SpecMetric` or `Metrics`
2. Record in appropriate locations
3. Update export format
4. Update summary output

---

## Future Enhancements

See [Future Roadmap](./future-roadmap.md) for planned improvements:

- OpenAPI 3.1 support
- Multiple generator support
- Plugin system for post-processors
- WebSocket/gRPC support
- Incremental generation
- Multi-language support

---

## Resources

- [Usage Guide](./usage-guide.md)
- [Configuration Guide](./configuration.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Troubleshooting Guide](./troubleshooting.md)
