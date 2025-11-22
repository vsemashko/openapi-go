# Contributing to OpenAPI Go Generator

Thank you for your interest in contributing to the OpenAPI Go Generator! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Setup](#development-setup)
4. [Development Workflow](#development-workflow)
5. [Testing](#testing)
6. [Code Style](#code-style)
7. [Commit Messages](#commit-messages)
8. [Pull Request Process](#pull-request-process)
9. [Architecture Overview](#architecture-overview)
10. [Adding Features](#adding-features)

## Code of Conduct

This project adheres to a code of professional conduct. By participating, you are expected to:

- Be respectful and inclusive
- Focus on what is best for the project
- Show empathy towards other contributors
- Accept constructive criticism gracefully
- Collaborate openly and transparently

## Getting Started

### Prerequisites

- **Go 1.24.0+**: Required for development
- **Git**: For version control
- **asdf** (optional): For managing tool versions
- **Task** (optional): For running predefined tasks

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd openapi-go

# Install dependencies
go mod download

# Install asdf tools (optional)
asdf install

# Run tests
go test ./...

# Build
go build
```

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitLab/GitHub

# Clone your fork
git clone git@gitlab.stashaway.com:your-username/openapi-go.git
cd openapi-go

# Add upstream remote
git remote add upstream git@gitlab.stashaway.com:vladimir.semashko/openapi-go.git
```

### 2. Install Development Tools

```bash
# Install asdf (if not already installed)
# macOS
brew install asdf

# Ubuntu/Debian
apt install asdf

# Install tools via asdf
asdf install

# Verify Go version
go version  # Should be 1.24.0 or later

# Install Task (optional)
go install github.com/go-task/task/v3/cmd/task@latest

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. Initialize Submodules

```bash
# Initialize git submodules
git submodule update --init --recursive

# Verify submodules
git submodule status
```

See [Submodule Management Guide](./docs/submodule-management.md) for details.

### 4. Verify Setup

```bash
# Run tests
go test ./...

# Run linter
golangci-lint run

# Build binary
go build

# Generate clients (integration test)
./openapi-go
```

## Development Workflow

### 1. Create a Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name

# Or bug fix branch
git checkout -b fix/issue-description
```

**Branch naming conventions**:
- `feature/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation updates
- `test/` - Test improvements
- `chore/` - Maintenance tasks

### 2. Make Changes

```bash
# Edit code
vim internal/processor/processor.go

# Run tests frequently
go test ./internal/processor -v

# Check linting
golangci-lint run ./internal/processor

# Build to verify
go build
```

### 3. Commit Changes

```bash
# Stage changes
git add internal/processor/processor.go

# Commit with descriptive message
git commit -m "feat: Add support for OpenAPI 3.1 specs"

# Push to your fork
git push origin feature/your-feature-name
```

See [Commit Messages](#commit-messages) for conventions.

### 4. Keep Branch Updated

```bash
# Fetch upstream changes
git fetch upstream

# Rebase on main
git rebase upstream/main

# Resolve conflicts if any
git rebase --continue

# Force push (if already pushed)
git push --force-with-lease origin feature/your-feature-name
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/processor -v

# Run specific test
go test ./internal/processor -run TestProcessOpenAPISpecs -v

# Run tests in parallel
go test ./... -parallel 4
```

### Writing Tests

**Test file naming**: `*_test.go`

**Test function naming**: `Test<FunctionName>`

**Example test**:
```go
package processor

import (
    "context"
    "testing"
)

func TestProcessOpenAPISpecs(t *testing.T) {
    tests := []struct {
        name    string
        cfg     config.Config
        wantErr bool
    }{
        {
            name: "valid config",
            cfg: config.Config{
                SpecsDir:  "./testdata/specs",
                OutputDir: "./testdata/output",
            },
            wantErr: false,
        },
        {
            name: "missing specs dir",
            cfg: config.Config{
                SpecsDir:  "./nonexistent",
                OutputDir: "./testdata/output",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            err := ProcessOpenAPISpecs(ctx, tt.cfg)
            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessOpenAPISpecs() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Coverage Requirements

- **Minimum coverage**: 75% for new code
- **Target coverage**: 85%+ for critical paths
- Run coverage check: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`

### Integration Tests

```bash
# Run full integration test
task generate-clients

# Verify generated code compiles
go build ./generated/clients/...

# Clean up
rm -rf generated/
```

## Code Style

### Go Formatting

```bash
# Format all code
gofmt -w .

# Or use goimports
goimports -w .

# Verify formatting
gofmt -d .
```

### Linting

```bash
# Run golangci-lint
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix

# Run on specific package
golangci-lint run ./internal/processor
```

### Style Guidelines

1. **Follow Go conventions**:
   - Use `gofmt` for formatting
   - Use meaningful variable names
   - Keep functions small and focused
   - Add comments for exported functions

2. **Error handling**:
   ```go
   // Good: wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to process spec: %w", err)
   }

   // Bad: lose error context
   if err != nil {
       return err
   }
   ```

3. **Logging**:
   ```go
   // Use structured logging
   logger.Info("Processing service",
       "service", serviceName,
       "spec", specPath,
   )

   // Avoid log.Printf in new code
   ```

4. **Comments**:
   ```go
   // ProcessOpenAPISpecs processes OpenAPI specifications and generates client code.
   // It searches for OpenAPI specs in the specified directory that match the targetServices pattern,
   // then generates Go client code for each spec using the configured generator.
   //
   // Parameters:
   // - ctx: Context for cancellation and timeouts
   // - cfg: Configuration containing specs directory, output directory, and target services pattern
   //
   // Returns an error if the process fails at any stage.
   func ProcessOpenAPISpecs(ctx context.Context, cfg config.Config) error {
       // Implementation
   }
   ```

5. **Package organization**:
   - One package per directory
   - Keep packages focused and cohesive
   - Minimize inter-package dependencies
   - Use internal/ for private packages

## Commit Messages

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks
- **perf**: Performance improvements

### Examples

```
feat(processor): Add support for OpenAPI 3.1 specs

Implement OpenAPI 3.1 parsing and generation using updated ogen library.
This adds support for webhooks, discriminator enhancements, and new
JSON Schema features.

Closes #123
```

```
fix(cache): Prevent cache corruption on concurrent writes

Add mutex locks around cache write operations to prevent race conditions
when multiple workers attempt to update cache entries simultaneously.

Fixes #456
```

```
docs: Add comprehensive usage guide

Create detailed usage guide covering:
- Quick start and installation
- Configuration options
- CI/CD integration examples
- Best practices

Related to #789
```

### Best Practices

1. **Use imperative mood**: "Add feature" not "Added feature"
2. **First line < 72 characters**: Keep subject line concise
3. **Explain why, not what**: Body explains motivation
4. **Reference issues**: Include issue numbers
5. **Break changes**: Note breaking changes in footer

## Pull Request Process

### 1. Prepare Your PR

Before creating a PR:

```bash
# Update from main
git fetch upstream
git rebase upstream/main

# Run tests
go test ./...

# Run linter
golangci-lint run

# Check coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total

# Verify build
go build
```

### 2. Create Pull Request

- Go to your fork on GitLab/GitHub
- Click "New Pull Request"
- Select your branch
- Fill in PR template:

```markdown
## Description
Brief description of changes

## Motivation
Why these changes are needed

## Changes
- Change 1
- Change 2
- Change 3

## Testing
How changes were tested

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Linter passes
- [ ] Coverage maintained/improved
- [ ] Breaking changes documented
```

### 3. PR Review Process

1. **Automated checks**: CI pipeline must pass
2. **Code review**: At least one approval required
3. **Address feedback**: Make requested changes
4. **Merge**: Maintainer will merge when approved

### 4. After Merge

```bash
# Update your local main
git checkout main
git pull upstream main

# Delete feature branch
git branch -d feature/your-feature-name
git push origin --delete feature/your-feature-name
```

## Architecture Overview

See [Architecture Documentation](./docs/architecture.md) for detailed architecture.

### Key Components

```
internal/
├── cache/          # Caching system (SHA256-based)
├── config/         # Configuration management
├── generator/      # Generator interface and implementations
├── logger/         # Structured logging
├── metrics/        # Metrics collection
├── paths/          # Path utilities
├── postprocessor/  # Post-processing chain
├── processor/      # Main processing logic
├── spec/           # OpenAPI spec utilities
└── worker/         # Parallel worker pool
```

### Design Principles

1. **Separation of Concerns**: Each package has a single responsibility
2. **Dependency Injection**: Pass dependencies explicitly
3. **Testability**: Write testable code with minimal mocks
4. **Error Handling**: Use wrapped errors with context
5. **Concurrency**: Use goroutines and channels safely
6. **Extensibility**: Use interfaces and plugin patterns

## Adding Features

### Adding a New Post-Processor

1. **Create processor**:
```go
// internal/postprocessor/myprocessor.go
package postprocessor

type MyProcessor struct{}

func (p *MyProcessor) Name() string {
    return "MyProcessor"
}

func (p *MyProcessor) Process(ctx context.Context, clientPath, serviceName, specPath string) error {
    // Implementation
    return nil
}
```

2. **Register processor**:
```go
// internal/postprocessor/chain.go
func init() {
    defaultChain = []PostProcessor{
        &InternalClientGenerator{},
        &GoFormatter{},
        &MyProcessor{},  // Add your processor
    }
}
```

3. **Add tests**:
```go
// internal/postprocessor/myprocessor_test.go
func TestMyProcessor_Process(t *testing.T) {
    // Test implementation
}
```

### Adding Configuration Options

1. **Add to Config struct**:
```go
// internal/config/config.go
type Config struct {
    // ...
    MyNewOption string `yaml:"my_new_option"`
}
```

2. **Add default value**:
```go
func LoadConfig() (Config, error) {
    cfg := Config{
        // ...
        MyNewOption: "default-value",
    }
    // ...
}
```

3. **Add environment variable**:
```go
if env := os.Getenv("MY_NEW_OPTION"); env != "" {
    cfg.MyNewOption = env
}
```

4. **Update documentation**:
- `docs/configuration.md`
- `resources/application.yml`

### Adding Generator Support

1. **Implement Generator interface**:
```go
// internal/generator/mygen.go
type MyGenerator struct{}

func (g *MyGenerator) Name() string {
    return "mygen"
}

func (g *MyGenerator) Version() string {
    return "v1.0.0"
}

func (g *MyGenerator) Generate(ctx context.Context, spec GenerateSpec) error {
    // Implementation
    return nil
}

func (g *MyGenerator) IsInstalled() bool {
    // Check if generator is available
    return true
}

func (g *MyGenerator) EnsureInstalled(ctx context.Context) error {
    // Install if needed
    return nil
}
```

2. **Add tests**:
```go
// internal/generator/mygen_test.go
func TestMyGenerator_Generate(t *testing.T) {
    // Test implementation
}
```

3. **Update processor**:
```go
// internal/processor/processor.go
var defaultGenerator = generator.NewMyGenerator()
```

## Questions and Support

- **Questions**: Open a discussion or reach out to the team
- **Bugs**: Open an issue with reproduction steps
- **Features**: Open an issue describing the use case
- **Security**: Report security issues privately to the team

## Resources

- [Usage Guide](./docs/usage-guide.md)
- [Configuration Guide](./docs/configuration.md)
- [Troubleshooting Guide](./docs/troubleshooting.md)
- [Architecture Documentation](./docs/architecture.md)
- [Submodule Management](./docs/submodule-management.md)

Thank you for contributing! 🎉
