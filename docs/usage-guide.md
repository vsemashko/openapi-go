# OpenAPI Go Generator - Usage Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Start](#quick-start)
3. [Basic Usage](#basic-usage)
4. [Configuration](#configuration)
5. [Advanced Features](#advanced-features)
6. [Using Generated Clients](#using-generated-clients)
7. [CI/CD Integration](#cicd-integration)
8. [Best Practices](#best-practices)

## Introduction

The OpenAPI Go Generator automates the process of generating Go client code from OpenAPI 3.0 specifications. It provides:

- **Multi-spec processing**: Process multiple OpenAPI specs in parallel
- **Smart caching**: Skip regeneration when specs haven't changed
- **Internal client support**: Auto-generate internal clients without security requirements
- **YAML support**: Process JSON, YAML, and YML spec formats
- **Structured logging**: Production-grade logging with configurable levels and formats
- **Metrics tracking**: Comprehensive performance and success metrics

## Quick Start

### Prerequisites

1. **Go 1.24.0+**: The project requires Go 1.24.0 or later
2. **asdf** (optional): For version management
3. **Task** (optional): For running predefined tasks

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd openapi-go

# Install Go dependencies
go mod download

# (Optional) Use asdf for version management
asdf install
```

### First Run

```bash
# Using Task (recommended)
task generate-clients

# Or run directly
go run main.go
```

This will:
1. Find all OpenAPI specs matching your configuration
2. Generate Go client code for each spec
3. Apply post-processors (internal client generation, formatting)
4. Export metrics to `.openapi-metrics.json`

## Basic Usage

### Running the Generator

**Using Task** (recommended):
```bash
task generate-clients
```

**Using Go directly**:
```bash
go run main.go
```

**Using the compiled binary**:
```bash
# Build first
go build -o openapi-go

# Run
./openapi-go
```

### Command-Line Output

The generator provides real-time feedback:

```
2025-11-22T10:00:00Z INFO Starting OpenAPI client generator
2025-11-22T10:00:00Z INFO Configuration loaded specs_directory=./external/sdk/sdk-packages
2025-11-22T10:00:01Z INFO Found 5 OpenAPI specs matching the criteria
2025-11-22T10:00:01Z INFO Processing 5 specs with 4 parallel workers
2025-11-22T10:00:02Z INFO Processing service: funding (spec: ./external/sdk/.../openapi.json)
2025-11-22T10:00:05Z INFO ✅ Successfully generated client for fundingsdk
...
2025-11-22T10:00:15Z INFO Metrics exported to: ./generated/.openapi-metrics.json
2025-11-22T10:00:15Z INFO Generation Summary: 5 total, 5 successful, 0 failed, 0 cached
2025-11-22T10:00:15Z INFO Success rate: 100.0%
2025-11-22T10:00:15Z INFO Cache hit rate: 0.0%
```

### Generated Output Structure

```
generated/
├── clients/                    # Generated client code
│   ├── fundingsdk/            # Funding service client
│   │   ├── oas_client_gen.go
│   │   ├── oas_json_gen.go
│   │   ├── oas_internal_client_gen.go  # Auto-generated
│   │   └── ...
│   ├── holidayssdk/           # Holidays service client
│   └── ...
└── .openapi-metrics.json      # Performance metrics (gitignored)
```

## Configuration

The generator is configured via `resources/application.yml`. See [Configuration Guide](./configuration.md) for detailed options.

### Basic Configuration Example

```yaml
# Specs location
specs_dir: "./external/sdk/sdk-packages"

# Output location
output_dir: "./generated"

# Service filter (regex)
target_services: "(funding-server-sdk|holidays-server-sdk)"

# Performance options
worker_count: 4
enable_cache: true
cache_dir: ".openapi-cache"

# Spec file patterns
spec_file_patterns:
  - "openapi.json"
  - "openapi.yaml"
  - "openapi.yml"

# Logging
log_level: "info"
log_format: "json"

# Error handling
continue_on_error: false
```

### Environment Variables

Override configuration with environment variables:

```bash
export SPECS_DIR="./custom/specs"
export OUTPUT_DIR="./custom/output"
export TARGET_SERVICES="my-service-sdk"
export WORKER_COUNT="8"
export LOG_LEVEL="debug"
export LOG_FORMAT="text"
```

## Advanced Features

### Parallel Processing

The generator processes multiple specs concurrently for performance:

```yaml
# Use 8 workers for parallel generation
worker_count: 8
```

**When to use more workers:**
- You have many specs to process (10+)
- Your machine has sufficient CPU cores
- Network/disk I/O is not a bottleneck

**When to use fewer workers:**
- CI/CD environments with limited resources
- Few specs to process (1-3)
- Memory constraints

### Smart Caching

The generator uses SHA256-based caching to skip regeneration:

```yaml
enable_cache: true
cache_dir: ".openapi-cache"
```

**Cache invalidation happens when:**
- The OpenAPI spec file changes
- The generator version changes
- The cache directory is deleted

**Cache benefits:**
- Significantly faster subsequent runs
- Reduced CPU usage
- Metrics track cache hit rate

### Error Handling

**Fail-fast mode** (default):
```yaml
continue_on_error: false
```
Stops on first failure. Recommended for CI/CD.

**Continue-on-error mode**:
```yaml
continue_on_error: true
```
Processes all specs even if some fail. Useful for debugging.

### Service Filtering

Use regex patterns to filter which services to generate:

```yaml
# Generate specific services
target_services: "(funding-server-sdk|holidays-server-sdk)"

# Generate all services ending with -sdk
target_services: ".*-sdk$"

# Generate all services (match everything)
target_services: ".*"
```

### Multiple Spec Formats

Support for JSON, YAML, and YML formats:

```yaml
spec_file_patterns:
  - "openapi.json"   # JSON format
  - "openapi.yaml"   # YAML format
  - "openapi.yml"    # YML format
```

### Logging Configuration

**JSON format** (recommended for production):
```yaml
log_level: "info"
log_format: "json"
```

Output:
```json
{"time":"2025-11-22T10:00:00Z","level":"INFO","msg":"Processing service","service":"funding"}
```

**Text format** (human-readable):
```yaml
log_level: "debug"
log_format: "text"
```

Output:
```
time=2025-11-22T10:00:00Z level=INFO msg="Processing service" service=funding
```

**Log levels**: `debug`, `info`, `warn`, `error`

### Metrics Collection

Metrics are automatically exported to `.openapi-metrics.json`:

```json
{
  "total_specs": 5,
  "successful_specs": 5,
  "failed_specs": 0,
  "cached_specs": 0,
  "total_duration_ms": 15234,
  "average_duration_ms": 3046,
  "start_time": "2025-11-22T10:00:00Z",
  "end_time": "2025-11-22T10:00:15Z",
  "spec_metrics": [
    {
      "spec_path": "./external/sdk/.../openapi.json",
      "service_name": "funding",
      "success": true,
      "cached": false,
      "duration_ms": 4523,
      "generated_at": "2025-11-22T10:00:05Z"
    }
  ]
}
```

## Using Generated Clients

### Client Initialization

The generator creates two client types for each service:

#### 1. Standard Client (with security)

```go
import "yourmodule/generated/clients/fundingsdk"

// For endpoints requiring authentication
client, err := fundingsdk.NewClient(
    "https://api.example.com",
    securitySource, // Your auth implementation
)
if err != nil {
    log.Fatal(err)
}
```

#### 2. Internal Client (without security)

```go
import "yourmodule/generated/clients/fundingsdk"

// For internal-only endpoints (no auth required)
client, err := fundingsdk.NewInternalClient(
    "https://internal-api.example.com",
)
if err != nil {
    log.Fatal(err)
}
```

### Making API Calls

```go
ctx := context.Background()

// Call an endpoint
response, err := client.GetUser(ctx, fundingsdk.GetUserParams{
    UserID: "12345",
})
if err != nil {
    log.Fatalf("API call failed: %v", err)
}

// Handle response
switch r := response.(type) {
case *fundingsdk.GetUserOK:
    fmt.Printf("User: %+v\n", r.User)
case *fundingsdk.GetUserNotFound:
    fmt.Println("User not found")
default:
    fmt.Printf("Unexpected response: %T\n", r)
}
```

### Error Handling

```go
// Type-safe error handling
response, err := client.CreateWithdrawal(ctx, req)
if err != nil {
    // Network or client error
    log.Printf("Request failed: %v", err)
    return err
}

// Handle different response types
switch r := response.(type) {
case *fundingsdk.CreateWithdrawalCreated:
    fmt.Printf("Success: %s\n", r.WithdrawalID)
case *fundingsdk.CreateWithdrawalBadRequest:
    fmt.Printf("Validation error: %s\n", r.Message)
case *fundingsdk.CreateWithdrawalUnauthorized:
    fmt.Println("Authentication required")
case *fundingsdk.CreateWithdrawalInternalServerError:
    fmt.Println("Server error")
}
```

## CI/CD Integration

### GitLab CI Example

```yaml
stages:
  - generate
  - test
  - deploy

generate-clients:
  stage: generate
  image: golang:1.24
  script:
    - go mod download
    - go run main.go
  artifacts:
    paths:
      - generated/
    expire_in: 1 day
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'
```

### GitHub Actions Example

```yaml
name: Generate Clients

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Generate clients
        run: |
          go mod download
          go run main.go

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: generated-clients
          path: generated/
```

### Automated Updates

For automated spec updates and regeneration:

```yaml
regenerate-on-spec-change:
  stage: generate
  script:
    - cd external/sdk
    - git pull origin main  # Update specs
    - cd ../..
    - go run main.go        # Regenerate clients
    - |
      if [ -n "$(git status --porcelain)" ]; then
        git config user.name "CI Bot"
        git config user.email "ci@example.com"
        git add generated/
        git commit -m "chore: Regenerate clients from updated specs"
        git push
      fi
  only:
    - schedules  # Run on schedule (e.g., daily)
```

## Best Practices

### 1. Use Version Control for Specs

Store OpenAPI specs in a git submodule:

```bash
git submodule add <specs-repo-url> external/sdk
git submodule update --init --recursive
```

See [Submodule Management Guide](./submodule-management.md) for details.

### 2. Enable Caching

Always enable caching for better performance:

```yaml
enable_cache: true
cache_dir: ".openapi-cache"
```

Add to `.gitignore`:
```
.openapi-cache/
.openapi-metrics.json
```

### 3. Use Fail-Fast in CI/CD

For CI/CD, use fail-fast mode to catch errors early:

```yaml
continue_on_error: false
```

### 4. Monitor Metrics

Track generation metrics over time:

```bash
# View metrics
cat generated/.openapi-metrics.json | jq .

# Check success rate
cat generated/.openapi-metrics.json | jq '.successful_specs / .total_specs * 100'

# Find slow specs
cat generated/.openapi-metrics.json | jq '.spec_metrics | sort_by(.duration_ms) | reverse | .[0:5]'
```

### 5. Use Structured Logging in Production

```yaml
log_level: "info"
log_format: "json"
```

Parse logs with tools like `jq`, `graylog`, or `datadog`.

### 6. Optimize Worker Count

Benchmark to find optimal worker count:

```bash
# Test with different worker counts
for workers in 1 2 4 8 16; do
  echo "Testing with $workers workers"
  export WORKER_COUNT=$workers
  time go run main.go
done
```

### 7. Validate Specs Before Generation

Use spec validation tools:

```bash
# Install swagger-cli
npm install -g @apidevtools/swagger-cli

# Validate spec
swagger-cli validate path/to/openapi.json
```

### 8. Keep Generated Code Separate

Never manually edit generated code:

```
generated/        # Generated (gitignored or committed as-is)
internal/         # Your application code
clients/          # Your custom client wrappers
```

### 9. Version Your Generated Clients

If committing generated code:

```bash
# Tag releases
git tag -a v1.0.0 -m "Release v1.0.0 with funding + holidays clients"
git push origin v1.0.0
```

### 10. Document Client Usage

Create examples for common use cases:

```go
// examples/funding/main.go
package main

import (
    "context"
    "log"
    "yourmodule/generated/clients/fundingsdk"
)

func main() {
    client, err := fundingsdk.NewInternalClient("https://api.example.com")
    if err != nil {
        log.Fatal(err)
    }

    // Example: Create withdrawal
    response, err := client.CreateWithdrawal(context.Background(), ...)
    // ... handle response
}
```

## Next Steps

- Read the [Configuration Guide](./configuration.md) for detailed configuration options
- Check the [Troubleshooting Guide](./troubleshooting.md) if you encounter issues
- Review the [Architecture Documentation](./architecture.md) to understand internals
- See [Contributing Guide](../CONTRIBUTING.md) to contribute improvements
