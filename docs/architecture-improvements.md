# OpenAPI Go SDK Generator - Architecture Improvements

**Version:** 1.0
**Date:** 2025-11-22
**Purpose:** Long-term architectural recommendations for scalability and maintainability

---

## Table of Contents

1. [Current Architecture Assessment](#current-architecture-assessment)
2. [Proposed Architecture](#proposed-architecture)
3. [Modularity Improvements](#modularity-improvements)
4. [Scalability Enhancements](#scalability-enhancements)
5. [Plugin System](#plugin-system)
6. [Observability & Monitoring](#observability--monitoring)
7. [Error Handling Strategy](#error-handling-strategy)
8. [Configuration Management](#configuration-management)
9. [Future Enhancements](#future-enhancements)

---

## Current Architecture Assessment

### Strengths ✅

1. **Simple & Linear** - Easy to understand the flow
2. **Minimal Dependencies** - Only essential packages
3. **Clear Separation** - Config, processor, postprocessor are separated
4. **Template-Based** - Easy to customize generated output

### Weaknesses ❌

1. **Tight Coupling** - Processor directly calls exec commands
2. **No Abstraction** - Hardcoded dependency on ogen
3. **Limited Extensibility** - Hard to add new generators or plugins
4. **No Observability** - No metrics, tracing, or structured logging
5. **Synchronous Processing** - Generates one spec at a time
6. **No Caching** - Regenerates everything every time

---

## Proposed Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                                │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐      │
│  │ generate │ validate │ list     │ version  │ watch    │      │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘      │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────────┐
│                      Application Layer                           │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐     │
│  │  Orchestrator│   Pipeline  │  Validator  │   Cache     │     │
│  └─────────────┴─────────────┴─────────────┴─────────────┘     │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────────┐
│                       Domain Layer                               │
│  ┌──────────────────────────────────────────────────────┐       │
│  │              Generator Interface                      │       │
│  │  ┌──────────┬──────────┬──────────┬──────────┐      │       │
│  │  │  Ogen    │  Custom  │  Swagger │  Future  │      │       │
│  │  │Generator │Generator │Codegen   │Generator │      │       │
│  │  └──────────┴──────────┴──────────┴──────────┘      │       │
│  └──────────────────────────────────────────────────────┘       │
│  ┌──────────────────────────────────────────────────────┐       │
│  │            Post-Processor Chain                       │       │
│  │  ┌──────────┬──────────┬──────────┬──────────┐      │       │
│  │  │Internal  │  Linter  │ Formatter│  Custom  │      │       │
│  │  │Client    │          │          │          │      │       │
│  │  └──────────┴──────────┴──────────┴──────────┘      │       │
│  └──────────────────────────────────────────────────────┘       │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────────┐
│                    Infrastructure Layer                          │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────┐          │
│  │ Spec    │ File    │ Git     │ HTTP    │Template │          │
│  │ Parser  │ System  │ Client  │ Client  │ Engine  │          │
│  └─────────┴─────────┴─────────┴─────────┴─────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

---

## Modularity Improvements

### 1. Generator Interface

**Problem:** Currently tightly coupled to ogen

**Solution:** Abstract generator behind interface

**File:** `internal/generator/generator.go` (new)

```go
package generator

import (
    "context"
)

// Generator defines the interface for code generators
type Generator interface {
    // Name returns the generator identifier
    Name() string

    // Generate generates client code from an OpenAPI spec
    Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error)

    // Validate checks if the spec is compatible with this generator
    Validate(ctx context.Context, specPath string) error
}

// GenerateRequest contains parameters for generation
type GenerateRequest struct {
    SpecPath     string
    OutputDir    string
    PackageName  string
    ConfigPath   string
    Options      map[string]interface{}
}

// GenerateResult contains the output of generation
type GenerateResult struct {
    OutputDir     string
    GeneratedFiles []string
    Warnings      []string
}

// Registry manages available generators
type Registry struct {
    generators map[string]Generator
}

// NewRegistry creates a new generator registry
func NewRegistry() *Registry {
    return &Registry{
        generators: make(map[string]Generator),
    }
}

// Register adds a generator to the registry
func (r *Registry) Register(gen Generator) {
    r.generators[gen.Name()] = gen
}

// Get retrieves a generator by name
func (r *Registry) Get(name string) (Generator, error) {
    gen, ok := r.generators[name]
    if !ok {
        return nil, fmt.Errorf("generator not found: %s", name)
    }
    return gen, nil
}

// List returns all registered generators
func (r *Registry) List() []string {
    names := make([]string, 0, len(r.generators))
    for name := range r.generators {
        names = append(names, name)
    }
    return names
}
```

**File:** `internal/generator/ogen/ogen.go` (new)

```go
package ogen

import (
    "context"
    "fmt"
    "os/exec"

    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/generator"
)

// OgenGenerator implements the Generator interface for ogen
type OgenGenerator struct {
    version string
}

// NewOgenGenerator creates a new ogen generator
func NewOgenGenerator(version string) *OgenGenerator {
    return &OgenGenerator{
        version: version,
    }
}

// Name returns the generator name
func (g *OgenGenerator) Name() string {
    return "ogen"
}

// Generate generates code using ogen
func (g *OgenGenerator) Generate(ctx context.Context, req generator.GenerateRequest) (*generator.GenerateResult, error) {
    // Ensure ogen is installed
    if err := g.ensureInstalled(ctx); err != nil {
        return nil, fmt.Errorf("failed to ensure ogen installation: %w", err)
    }

    // Build command
    cmd := exec.CommandContext(ctx, "ogen",
        "--target", req.OutputDir,
        "--package", req.PackageName,
        "--clean",
        "--config", req.ConfigPath,
        req.SpecPath)

    // Execute
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("ogen failed: %w\nOutput: %s", err, string(output))
    }

    // Collect generated files
    files, err := collectGeneratedFiles(req.OutputDir)
    if err != nil {
        return nil, err
    }

    return &generator.GenerateResult{
        OutputDir:      req.OutputDir,
        GeneratedFiles: files,
        Warnings:       parseWarnings(output),
    }, nil
}

// Validate checks if spec is valid for ogen
func (g *OgenGenerator) Validate(ctx context.Context, specPath string) error {
    // Parse spec and check for ogen compatibility
    // Return error if incompatible
    return nil
}

// ensureInstalled ensures ogen is installed with correct version
func (g *OgenGenerator) ensureInstalled(ctx context.Context) error {
    // Check if installed
    // Install if needed
    return nil
}
```

**Usage:**

```go
// In main.go
registry := generator.NewRegistry()
registry.Register(ogen.NewOgenGenerator("v1.14.0"))

// Future: Add more generators
// registry.Register(swagger.NewSwaggerCodegen())
// registry.Register(custom.NewCustomGenerator())

// Use generator
gen, _ := registry.Get("ogen")
result, err := gen.Generate(ctx, req)
```

**Benefits:**
- ✅ Easy to swap generators
- ✅ Can support multiple generators
- ✅ Testable with mocks
- ✅ Clear contract

---

### 2. Post-Processor Chain

**Problem:** Only one hardcoded post-processor (internal client)

**Solution:** Chain of responsibility pattern

**File:** `internal/postprocessor/postprocessor.go` (refactored)

```go
package postprocessor

import (
    "context"
)

// PostProcessor defines the interface for post-processing steps
type PostProcessor interface {
    // Name returns the processor identifier
    Name() string

    // Process applies post-processing to generated code
    Process(ctx context.Context, req ProcessRequest) error

    // ShouldRun determines if this processor should run
    ShouldRun(ctx context.Context, req ProcessRequest) bool
}

// ProcessRequest contains context for post-processing
type ProcessRequest struct {
    OutputDir   string
    PackageName string
    SpecPath    string
    Metadata    map[string]interface{}
}

// Chain manages a sequence of post-processors
type Chain struct {
    processors []PostProcessor
}

// NewChain creates a new post-processor chain
func NewChain(processors ...PostProcessor) *Chain {
    return &Chain{
        processors: processors,
    }
}

// Add adds a processor to the chain
func (c *Chain) Add(p PostProcessor) {
    c.processors = append(c.processors, p)
}

// Execute runs all processors in order
func (c *Chain) Execute(ctx context.Context, req ProcessRequest) error {
    for _, processor := range c.processors {
        if !processor.ShouldRun(ctx, req) {
            continue
        }

        if err := processor.Process(ctx, req); err != nil {
            return fmt.Errorf("processor %s failed: %w", processor.Name(), err)
        }
    }
    return nil
}
```

**File:** `internal/postprocessor/internal_client.go` (new)

```go
package postprocessor

import (
    "context"
)

// InternalClientProcessor adds NewInternalClient helper
type InternalClientProcessor struct {
    templatePath string
}

func NewInternalClientProcessor(templatePath string) *InternalClientProcessor {
    return &InternalClientProcessor{
        templatePath: templatePath,
    }
}

func (p *InternalClientProcessor) Name() string {
    return "internal_client"
}

func (p *InternalClientProcessor) Process(ctx context.Context, req ProcessRequest) error {
    // Generate internal client file
    // (existing logic from postprocessor.go)
    return nil
}

func (p *InternalClientProcessor) ShouldRun(ctx context.Context, req ProcessRequest) bool {
    // Always run
    return true
}
```

**File:** `internal/postprocessor/linter.go` (new - example)

```go
package postprocessor

// LinterProcessor runs linter on generated code
type LinterProcessor struct{}

func (p *LinterProcessor) Name() string {
    return "linter"
}

func (p *LinterProcessor) Process(ctx context.Context, req ProcessRequest) error {
    // Run golangci-lint on generated code
    return nil
}

func (p *LinterProcessor) ShouldRun(ctx context.Context, req ProcessRequest) bool {
    // Check if linting is enabled in config
    return req.Metadata["enable_linting"].(bool)
}
```

**Usage:**

```go
chain := postprocessor.NewChain(
    postprocessor.NewInternalClientProcessor("templates/internal_client.tmpl"),
    postprocessor.NewFormatterProcessor(), // gofmt
    postprocessor.NewLinterProcessor(),    // golangci-lint
    postprocessor.NewCustomProcessor(),    // User-defined
)

err := chain.Execute(ctx, req)
```

---

### 3. Pipeline Architecture

**File:** `internal/pipeline/pipeline.go` (new)

```go
package pipeline

import (
    "context"
    "sync"

    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/generator"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/postprocessor"
)

// Stage represents a processing stage
type Stage interface {
    Execute(ctx context.Context, input StageInput) (StageOutput, error)
}

// Pipeline orchestrates the generation workflow
type Pipeline struct {
    stages []Stage
}

// StageInput contains input for a stage
type StageInput struct {
    Spec         Spec
    OutputDir    string
    Config       Config
    PreviousData map[string]interface{}
}

// StageOutput contains output from a stage
type StageOutput struct {
    Success bool
    Data    map[string]interface{}
    Error   error
}

// NewPipeline creates a generation pipeline
func NewPipeline(stages ...Stage) *Pipeline {
    return &Pipeline{stages: stages}
}

// Execute runs the pipeline
func (p *Pipeline) Execute(ctx context.Context, input StageInput) error {
    stageInput := input

    for _, stage := range p.stages {
        output, err := stage.Execute(ctx, stageInput)
        if err != nil {
            return err
        }

        // Pass output to next stage
        stageInput.PreviousData = output.Data
    }

    return nil
}

// Stages:

type DiscoveryStage struct{}
type ValidationStage struct{}
type GenerationStage struct {
    generator generator.Generator
}
type PostProcessStage struct {
    chain *postprocessor.Chain
}
type CompilationStage struct{}

// Example pipeline:
// Discovery -> Validation -> Generation -> PostProcess -> Compilation
```

---

## Scalability Enhancements

### 1. Parallel Processing

**Problem:** Generates specs sequentially, slow for many specs

**Solution:** Concurrent generation with worker pool

**File:** `internal/worker/pool.go` (new)

```go
package worker

import (
    "context"
    "sync"
)

// Pool manages concurrent workers
type Pool struct {
    workers   int
    taskQueue chan Task
    results   chan Result
    wg        sync.WaitGroup
}

// Task represents a generation task
type Task struct {
    ID       string
    SpecPath string
    Config   interface{}
}

// Result contains task result
type Result struct {
    TaskID  string
    Success bool
    Error   error
}

// NewPool creates a worker pool
func NewPool(workers int) *Pool {
    return &Pool{
        workers:   workers,
        taskQueue: make(chan Task, workers*2),
        results:   make(chan Result, workers*2),
    }
}

// Start starts the worker pool
func (p *Pool) Start(ctx context.Context, processor TaskProcessor) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx, processor)
    }
}

// Submit submits a task to the pool
func (p *Pool) Submit(task Task) {
    p.taskQueue <- task
}

// Wait waits for all tasks to complete
func (p *Pool) Wait() []Result {
    close(p.taskQueue)
    p.wg.Wait()
    close(p.results)

    var results []Result
    for result := range p.results {
        results = append(results, result)
    }
    return results
}

// worker processes tasks
func (p *Pool) worker(ctx context.Context, processor TaskProcessor) {
    defer p.wg.Done()

    for task := range p.taskQueue {
        err := processor.Process(ctx, task)
        p.results <- Result{
            TaskID:  task.ID,
            Success: err == nil,
            Error:   err,
        }
    }
}

// TaskProcessor defines task processing interface
type TaskProcessor interface {
    Process(ctx context.Context, task Task) error
}
```

**Usage:**

```go
// Create pool with 4 workers
pool := worker.NewPool(4)
pool.Start(ctx, generationProcessor)

// Submit specs
for _, spec := range specs {
    pool.Submit(worker.Task{
        ID:       spec.Name,
        SpecPath: spec.Path,
        Config:   config,
    })
}

// Wait for completion
results := pool.Wait()

// Check results
for _, result := range results {
    if !result.Success {
        log.Printf("Failed: %s: %v", result.TaskID, result.Error)
    }
}
```

**Performance Impact:**
- Current: 10 specs × 10s each = **100 seconds**
- With 4 workers: 10 specs ÷ 4 × 10s = **25 seconds** (4x speedup)

---

### 2. Caching & Incremental Generation

**Problem:** Regenerates everything even if specs haven't changed

**Solution:** Hash-based caching

**File:** `internal/cache/cache.go` (new)

```go
package cache

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "os"
    "path/filepath"
)

// Cache manages generation cache
type Cache struct {
    cacheDir string
    index    map[string]CacheEntry
}

// CacheEntry represents cached generation metadata
type CacheEntry struct {
    SpecHash    string
    ConfigHash  string
    GeneratedAt time.Time
    OutputDir   string
    Files       []string
}

// NewCache creates a new cache
func NewCache(cacheDir string) (*Cache, error) {
    cache := &Cache{
        cacheDir: cacheDir,
        index:    make(map[string]CacheEntry),
    }

    if err := cache.load(); err != nil {
        return nil, err
    }

    return cache, nil
}

// NeedsRegeneration checks if spec needs to be regenerated
func (c *Cache) NeedsRegeneration(specPath, configPath string) bool {
    // Calculate hashes
    specHash, _ := hashFile(specPath)
    configHash, _ := hashFile(configPath)

    // Check cache
    entry, exists := c.index[specPath]
    if !exists {
        return true // Never generated
    }

    // Check if hashes match
    if entry.SpecHash != specHash || entry.ConfigHash != configHash {
        return true // Changed
    }

    // Check if output still exists
    for _, file := range entry.Files {
        if _, err := os.Stat(file); os.IsNotExist(err) {
            return true // Output missing
        }
    }

    return false // Cache valid
}

// Update updates cache after generation
func (c *Cache) Update(specPath, configPath, outputDir string, files []string) error {
    specHash, _ := hashFile(specPath)
    configHash, _ := hashFile(configPath)

    c.index[specPath] = CacheEntry{
        SpecHash:    specHash,
        ConfigHash:  configHash,
        GeneratedAt: time.Now(),
        OutputDir:   outputDir,
        Files:       files,
    }

    return c.save()
}

// hashFile calculates SHA256 hash of file
func hashFile(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }

    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:]), nil
}

// load loads cache index from disk
func (c *Cache) load() error {
    indexPath := filepath.Join(c.cacheDir, "index.json")

    data, err := os.ReadFile(indexPath)
    if os.IsNotExist(err) {
        return nil // No cache yet
    }
    if err != nil {
        return err
    }

    return json.Unmarshal(data, &c.index)
}

// save saves cache index to disk
func (c *Cache) save() error {
    os.MkdirAll(c.cacheDir, 0755)

    data, err := json.MarshalIndent(c.index, "", "  ")
    if err != nil {
        return err
    }

    indexPath := filepath.Join(c.cacheDir, "index.json")
    return os.WriteFile(indexPath, data, 0644)
}
```

**Usage:**

```go
cache, err := cache.NewCache(".cache")

for _, spec := range specs {
    if !cache.NeedsRegeneration(spec.Path, configPath) {
        log.Printf("Skipping %s (unchanged)", spec.Name)
        continue
    }

    // Generate
    result, err := generate(spec)

    // Update cache
    cache.Update(spec.Path, configPath, result.OutputDir, result.Files)
}
```

---

## Plugin System

**File:** `internal/plugin/plugin.go` (new)

```go
package plugin

import (
    "context"
    "plugin"
)

// Plugin defines the plugin interface
type Plugin interface {
    Name() string
    Version() string
    Init(config map[string]interface{}) error
    Execute(ctx context.Context, event Event) error
}

// Event represents a plugin event
type Event struct {
    Type    string
    Payload interface{}
}

// Loader loads plugins from .so files
type Loader struct {
    pluginDir string
    plugins   map[string]Plugin
}

// LoadPlugin loads a plugin from file
func (l *Loader) LoadPlugin(path string) error {
    // Load .so file
    p, err := plugin.Open(path)
    if err != nil {
        return err
    }

    // Look up symbol
    symbol, err := p.Lookup("Plugin")
    if err != nil {
        return err
    }

    // Assert type
    plug, ok := symbol.(Plugin)
    if !ok {
        return fmt.Errorf("invalid plugin type")
    }

    l.plugins[plug.Name()] = plug
    return nil
}
```

**Example Plugin:**

```go
// plugins/custom/custom.go
package main

import (
    "context"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/plugin"
)

type CustomPlugin struct{}

func (p *CustomPlugin) Name() string { return "custom" }
func (p *CustomPlugin) Version() string { return "1.0.0" }

func (p *CustomPlugin) Init(config map[string]interface{}) error {
    return nil
}

func (p *CustomPlugin) Execute(ctx context.Context, event plugin.Event) error {
    // Custom logic
    return nil
}

var Plugin CustomPlugin
```

---

## Observability & Monitoring

### 1. Structured Logging

**File:** `internal/logger/logger.go` (new)

```go
package logger

import (
    "log/slog"
    "os"
)

// Logger wraps slog with custom configuration
type Logger struct {
    *slog.Logger
}

// New creates a new structured logger
func New(level slog.Level) *Logger {
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: level,
    })

    return &Logger{
        Logger: slog.New(handler),
    }
}

// WithSpec adds spec context to logger
func (l *Logger) WithSpec(specPath, serviceName string) *Logger {
    return &Logger{
        Logger: l.With(
            "spec_path", specPath,
            "service", serviceName,
        ),
    }
}
```

**Usage:**

```go
logger := logger.New(slog.LevelInfo)

logger.Info("starting generation",
    "total_specs", len(specs),
    "output_dir", outputDir)

specLogger := logger.WithSpec(specPath, serviceName)
specLogger.Info("generating client")

specLogger.Error("generation failed",
    "error", err,
    "duration_ms", duration.Milliseconds())
```

---

### 2. Metrics

**File:** `internal/metrics/metrics.go` (new)

```go
package metrics

import (
    "time"
)

// Collector collects generation metrics
type Collector struct {
    metrics map[string]Metric
}

// Metric represents a metric value
type Metric struct {
    Name      string
    Value     float64
    Labels    map[string]string
    Timestamp time.Time
}

// RecordGeneration records generation metrics
func (c *Collector) RecordGeneration(serviceName string, duration time.Duration, success bool) {
    c.metrics[serviceName] = Metric{
        Name:  "generation_duration_seconds",
        Value: duration.Seconds(),
        Labels: map[string]string{
            "service": serviceName,
            "success": fmt.Sprintf("%v", success),
        },
        Timestamp: time.Now(),
    }
}

// Export exports metrics (Prometheus format, JSON, etc.)
func (c *Collector) Export() string {
    // Format metrics for export
    return ""
}
```

---

## Error Handling Strategy

### Improved Error Types

**File:** `internal/errors/errors.go` (new)

```go
package errors

import (
    "fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
    ErrorTypeConfig     ErrorType = "config"
    ErrorTypeValidation ErrorType = "validation"
    ErrorTypeGeneration ErrorType = "generation"
    ErrorTypeIO         ErrorType = "io"
)

// Error represents a structured error
type Error struct {
    Type    ErrorType
    Message string
    Cause   error
    Context map[string]interface{}
}

// Error implements error interface
func (e *Error) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
    return e.Cause
}

// New creates a new structured error
func New(errType ErrorType, message string) *Error {
    return &Error{
        Type:    errType,
        Message: message,
        Context: make(map[string]interface{}),
    }
}

// Wrap wraps an error with context
func Wrap(err error, errType ErrorType, message string) *Error {
    return &Error{
        Type:    errType,
        Message: message,
        Cause:   err,
        Context: make(map[string]interface{}),
    }
}

// WithContext adds context to error
func (e *Error) WithContext(key string, value interface{}) *Error {
    e.Context[key] = value
    return e
}
```

**Usage:**

```go
func generateClient(spec string) error {
    if _, err := os.Stat(spec); err != nil {
        return errors.Wrap(err, errors.ErrorTypeIO, "spec file not found").
            WithContext("spec_path", spec)
    }

    // ...

    if err := runOgen(); err != nil {
        return errors.Wrap(err, errors.ErrorTypeGeneration, "ogen failed").
            WithContext("spec", spec).
            WithContext("output_dir", outputDir)
    }

    return nil
}
```

---

## Configuration Management

### Layered Configuration

**File:** `internal/config/config.go` (enhanced)

```go
package config

// Config with validation and defaults
type Config struct {
    // Core settings
    SpecsDir       string `mapstructure:"specs_dir" validate:"required,dir"`
    OutputDir      string `mapstructure:"output_dir" validate:"required"`
    TargetServices string `mapstructure:"target_services" validate:"regexp"`

    // Generation settings
    Generator struct {
        Name    string                 `mapstructure:"name" default:"ogen"`
        Version string                 `mapstructure:"version" default:"v1.14.0"`
        Options map[string]interface{} `mapstructure:"options"`
    } `mapstructure:"generator"`

    // Processing settings
    Processing struct {
        Workers         int  `mapstructure:"workers" default:"4"`
        ContinueOnError bool `mapstructure:"continue_on_error" default:"false"`
        EnableCache     bool `mapstructure:"enable_cache" default:"true"`
        CacheDir        string `mapstructure:"cache_dir" default:".cache"`
    } `mapstructure:"processing"`

    // Post-processing
    PostProcessors []string `mapstructure:"post_processors"`

    // Observability
    Logging struct {
        Level  string `mapstructure:"level" default:"info"`
        Format string `mapstructure:"format" default:"json"`
    } `mapstructure:"logging"`

    // Plugins
    Plugins []PluginConfig `mapstructure:"plugins"`
}

type PluginConfig struct {
    Name    string                 `mapstructure:"name"`
    Path    string                 `mapstructure:"path"`
    Enabled bool                   `mapstructure:"enabled"`
    Config  map[string]interface{} `mapstructure:"config"`
}
```

**File:** `resources/application.yml` (enhanced)

```yaml
# Core settings
specs_dir: "./external/sdk/sdk-packages"
output_dir: "./generated"
target_services: "(funding-server-sdk|holidays-server-sdk)"

# Generator settings
generator:
  name: ogen
  version: v1.14.0
  options:
    enable_validation: true

# Processing settings
processing:
  workers: 4
  continue_on_error: false
  enable_cache: true
  cache_dir: .cache

# Post-processors (executed in order)
post_processors:
  - internal_client
  - formatter
  - linter

# Logging
logging:
  level: info
  format: json

# Plugins
plugins:
  - name: custom_validator
    path: ./plugins/validator.so
    enabled: true
    config:
      strict_mode: true
```

---

## Future Enhancements

### 1. Watch Mode

Monitor specs for changes and auto-regenerate

```go
// internal/watcher/watcher.go
package watcher

import (
    "github.com/fsnotify/fsnotify"
)

type Watcher struct {
    specsDir string
    onChange func(path string)
}

func (w *Watcher) Watch() error {
    watcher, err := fsnotify.NewWatcher()
    // Watch for changes
    // Trigger regeneration
}
```

**Usage:**

```bash
openapi-go watch --specs ./specs
```

---

### 2. REST API

Expose generation as an HTTP service

```go
// cmd/server/main.go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    r.POST("/generate", handleGenerate)
    r.GET("/status/:jobId", handleStatus)
    r.GET("/download/:jobId", handleDownload)

    r.Run(":8080")
}
```

---

### 3. Web UI

Dashboard for managing SDK generation

- Upload OpenAPI specs
- Configure generation settings
- View generation history
- Download generated SDKs

---

### 4. Multi-Language Support

Generate SDKs for multiple languages

```go
type MultiLangConfig struct {
    Languages []LanguageConfig
}

type LanguageConfig struct {
    Name      string // go, typescript, python, java
    Generator string
    OutputDir string
}
```

---

## Migration Path

### Phase 1: Foundation (Weeks 1-2)
- ✅ Implement Generator interface
- ✅ Add structured logging
- ✅ Improve error handling

### Phase 2: Scalability (Weeks 3-4)
- ✅ Implement worker pool
- ✅ Add caching system
- ✅ Implement post-processor chain

### Phase 3: Extensibility (Weeks 5-6)
- ✅ Implement plugin system
- ✅ Add metrics collection
- ✅ Enhanced configuration

### Phase 4: Advanced Features (Weeks 7-8)
- ✅ Watch mode
- ✅ REST API
- ✅ Web UI (optional)

---

## Summary

These architectural improvements will transform the SDK generator from a simple script into a **robust, scalable, enterprise-grade tool** with:

✅ **Modularity** - Pluggable generators and post-processors
✅ **Scalability** - Parallel processing and caching
✅ **Extensibility** - Plugin system for custom logic
✅ **Observability** - Structured logging and metrics
✅ **Reliability** - Better error handling and testing
✅ **Maintainability** - Clear architecture and separation of concerns

**Next Steps:**
- Review and prioritize features
- See `implementation-roadmap.md` for detailed timeline
- Start with P0 fixes, then gradually adopt architectural improvements
