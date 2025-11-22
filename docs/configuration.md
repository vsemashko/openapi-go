# OpenAPI Go Generator - Configuration Guide

## Table of Contents

1. [Overview](#overview)
2. [Configuration File](#configuration-file)
3. [Configuration Options](#configuration-options)
4. [Environment Variables](#environment-variables)
5. [Configuration Examples](#configuration-examples)
6. [Advanced Configuration](#advanced-configuration)

## Overview

The OpenAPI Go Generator is configured using a YAML configuration file (`resources/application.yml`) with support for environment variable overrides. This allows for flexible configuration across different environments (development, staging, production).

## Configuration File

### Location

The default configuration file is located at:
```
resources/application.yml
```

### Format

The configuration uses YAML format:

```yaml
# Directory containing OpenAPI specs
specs_dir: "./external/sdk/sdk-packages"

# Output directory for generated clients
output_dir: "./generated"

# Regex pattern to filter services
target_services: ".*"

# Performance and behavior options
worker_count: 4
continue_on_error: false

# Caching configuration
enable_cache: true
cache_dir: ".openapi-cache"

# Spec file patterns
spec_file_patterns:
  - "openapi.json"
  - "openapi.yaml"
  - "openapi.yml"

# Logging configuration
log_level: "info"
log_format: "json"
```

## Configuration Options

### Specs Directory

**Option**: `specs_dir`
**Type**: String
**Default**: `"./external/sdk/sdk-packages"`
**Environment Variable**: `SPECS_DIR`

The directory where OpenAPI specification files are located. The generator will recursively search this directory for spec files.

```yaml
specs_dir: "./external/sdk/sdk-packages"
```

**Examples**:
```yaml
# Relative path
specs_dir: "./specs"

# Absolute path
specs_dir: "/home/user/project/specs"

# Specs in subdirectories
specs_dir: "./services/api-specs"
```

### Output Directory

**Option**: `output_dir`
**Type**: String
**Default**: `"./generated"`
**Environment Variable**: `OUTPUT_DIR`

The directory where generated client code will be written. The generator creates a `clients/` subdirectory within this path.

```yaml
output_dir: "./generated"
```

Generated structure:
```
{output_dir}/
├── clients/
│   ├── fundingsdk/
│   ├── holidayssdk/
│   └── ...
└── .openapi-metrics.json
```

**Examples**:
```yaml
# Development environment
output_dir: "./generated"

# Production build
output_dir: "./build/clients"

# Custom location
output_dir: "/tmp/openapi-clients"
```

### Target Services

**Option**: `target_services`
**Type**: String (Regex Pattern)
**Default**: `".*"` (matches all)
**Environment Variable**: `TARGET_SERVICES`

A regular expression pattern to filter which services to generate clients for. Only specs in directories matching this pattern will be processed.

```yaml
target_services: "(funding-server-sdk|holidays-server-sdk)"
```

**Pattern Matching**:
The pattern matches against the **directory name** containing the OpenAPI spec file.

For example, with structure:
```
specs/
├── funding-server-sdk/
│   └── openapi.json
├── holidays-server-sdk/
│   └── openapi.json
└── internal-api-sdk/
    └── openapi.json
```

**Examples**:
```yaml
# Generate all services (default)
target_services: ".*"

# Generate specific services
target_services: "(funding-server-sdk|holidays-server-sdk)"

# Generate all services ending with -sdk
target_services: ".*-sdk$"

# Generate all services starting with api-
target_services: "^api-.*"

# Generate services with "payment" in name
target_services: ".*payment.*"

# Generate services matching complex pattern
target_services: "^(prod|staging)-.*-server-sdk$"
```

**Common Patterns**:
```yaml
# All services
target_services: ".*"

# Production services only
target_services: "^prod-.*"

# Specific service types
target_services: ".*(api|server|gateway)-sdk$"

# Exclude test services
target_services: "^(?!test-).*"
```

### Worker Count

**Option**: `worker_count`
**Type**: Integer
**Default**: `4`
**Environment Variable**: `WORKER_COUNT`

The number of parallel workers to use for spec processing. Higher values increase throughput but consume more resources.

```yaml
worker_count: 4
```

**Guidelines**:
- **1 worker**: Sequential processing (debugging)
- **2-4 workers**: Good for small projects (< 10 specs)
- **4-8 workers**: Recommended for most projects (10-50 specs)
- **8-16 workers**: Large projects (50+ specs) with sufficient resources

**Considerations**:
- More workers = more CPU and memory usage
- Optimal count ≈ number of CPU cores
- Network/disk I/O may be bottleneck
- CI/CD environments may have resource limits

**Examples**:
```yaml
# Single worker (debugging)
worker_count: 1

# Optimal for quad-core CPU
worker_count: 4

# High throughput on powerful machine
worker_count: 16

# CI/CD with limited resources
worker_count: 2
```

### Continue on Error

**Option**: `continue_on_error`
**Type**: Boolean
**Default**: `false`
**Environment Variable**: `CONTINUE_ON_ERROR` (values: `true`, `false`)

Controls whether generation continues after encountering an error. When `false`, the generator stops at the first failure (fail-fast mode).

```yaml
continue_on_error: false
```

**Behavior**:

**Fail-fast mode** (`false`):
- Stops at first error
- Returns non-zero exit code
- Recommended for CI/CD
- Ensures all-or-nothing generation

**Continue mode** (`true`):
- Processes all specs even if some fail
- Collects all errors
- Useful for debugging
- Returns success if at least one spec succeeds

**Examples**:
```yaml
# CI/CD: Stop on first failure
continue_on_error: false

# Development: Process all specs to see all errors
continue_on_error: true
```

### Enable Cache

**Option**: `enable_cache`
**Type**: Boolean
**Default**: `true`
**Environment Variable**: `ENABLE_CACHE` (values: `true`, `false`)

Enables SHA256-based caching to skip regeneration of unchanged specs.

```yaml
enable_cache: true
```

**Benefits**:
- Significantly faster subsequent runs
- Reduced CPU usage
- Metrics track cache hit rate

**Cache Invalidation**:
Cache entries are invalidated when:
- The OpenAPI spec file content changes
- The generator version changes
- Cache entries are manually deleted

**Examples**:
```yaml
# Enable caching (recommended)
enable_cache: true

# Disable caching (force regeneration)
enable_cache: false
```

### Cache Directory

**Option**: `cache_dir`
**Type**: String
**Default**: `".openapi-cache"`
**Environment Variable**: `CACHE_DIR`

The directory where cache entries are stored.

```yaml
cache_dir: ".openapi-cache"
```

**Cache Structure**:
```
.openapi-cache/
└── {sha256-hash}.json  # Metadata for each cached spec
```

**Management**:
```bash
# Clear cache
rm -rf .openapi-cache

# View cache entries
ls -la .openapi-cache

# Check cache size
du -sh .openapi-cache
```

**Examples**:
```yaml
# Default location (project-local)
cache_dir: ".openapi-cache"

# System-wide cache
cache_dir: "/var/cache/openapi-go"

# User cache directory
cache_dir: "~/.cache/openapi-go"

# Temporary cache
cache_dir: "/tmp/openapi-cache"
```

### Spec File Patterns

**Option**: `spec_file_patterns`
**Type**: Array of Strings
**Default**: `["openapi.json", "openapi.yaml", "openapi.yml"]`
**Environment Variable**: Not supported (use config file)

List of filename patterns to identify OpenAPI spec files. The generator matches exact filenames (not glob patterns).

```yaml
spec_file_patterns:
  - "openapi.json"
  - "openapi.yaml"
  - "openapi.yml"
```

**Matching**:
- Exact filename match
- Case-sensitive
- Searches recursively in `specs_dir`

**Examples**:
```yaml
# JSON only
spec_file_patterns:
  - "openapi.json"

# YAML only
spec_file_patterns:
  - "openapi.yaml"
  - "openapi.yml"

# All formats (default)
spec_file_patterns:
  - "openapi.json"
  - "openapi.yaml"
  - "openapi.yml"

# Custom naming conventions
spec_file_patterns:
  - "api-spec.json"
  - "swagger.json"
  - "openapi.json"

# Versioned specs
spec_file_patterns:
  - "openapi-v1.json"
  - "openapi-v2.json"
  - "openapi.json"
```

### Log Level

**Option**: `log_level`
**Type**: String
**Default**: `"info"`
**Environment Variable**: `LOG_LEVEL`

Controls the verbosity of log output.

```yaml
log_level: "info"
```

**Available Levels**:

1. **debug**: Detailed diagnostic information
   - All messages including debug details
   - Useful for troubleshooting
   - Very verbose

2. **info**: General informational messages
   - Normal operation logs
   - Progress updates
   - Recommended for production

3. **warn**: Warning messages
   - Potential issues
   - Degraded functionality
   - Non-critical errors

4. **error**: Error messages only
   - Critical failures
   - Minimal output

**Examples**:
```yaml
# Development: verbose logging
log_level: "debug"

# Production: standard logging
log_level: "info"

# Production: warnings only
log_level: "warn"

# Production: errors only
log_level: "error"
```

**Output Comparison**:
```bash
# DEBUG level
2025-11-22T10:00:00Z DEBUG Cache check for fundingsdk
2025-11-22T10:00:00Z DEBUG SHA256: abc123...
2025-11-22T10:00:01Z INFO Processing service: funding
2025-11-22T10:00:05Z INFO Successfully generated client

# INFO level
2025-11-22T10:00:01Z INFO Processing service: funding
2025-11-22T10:00:05Z INFO Successfully generated client

# WARN level
(only warnings and errors)

# ERROR level
(only errors)
```

### Log Format

**Option**: `log_format`
**Type**: String
**Default**: `"json"`
**Environment Variable**: `LOG_FORMAT`

Controls the format of log output.

```yaml
log_format: "json"
```

**Available Formats**:

1. **json**: Structured JSON format
   - Machine-readable
   - Recommended for production
   - Easy to parse and index
   - Compatible with log aggregation tools

2. **text**: Human-readable text format
   - Key-value pairs
   - Good for development
   - Easier to read in terminal

**Examples**:
```yaml
# Production: JSON format
log_format: "json"

# Development: text format
log_format: "text"
```

**Output Comparison**:

**JSON format**:
```json
{"time":"2025-11-22T10:00:00Z","level":"INFO","msg":"Processing service","service":"funding","spec":"./specs/funding/openapi.json"}
{"time":"2025-11-22T10:00:05Z","level":"INFO","msg":"Successfully generated client","service":"funding","duration_ms":4523}
```

**Text format**:
```
time=2025-11-22T10:00:00Z level=INFO msg="Processing service" service=funding spec=./specs/funding/openapi.json
time=2025-11-22T10:00:05Z level=INFO msg="Successfully generated client" service=funding duration_ms=4523
```

## Environment Variables

All configuration options can be overridden using environment variables. This is useful for CI/CD pipelines and different deployment environments.

### Available Variables

| Config Option | Environment Variable | Type | Example |
|--------------|---------------------|------|---------|
| `specs_dir` | `SPECS_DIR` | String | `./specs` |
| `output_dir` | `OUTPUT_DIR` | String | `./output` |
| `target_services` | `TARGET_SERVICES` | String | `funding-.*` |
| `worker_count` | `WORKER_COUNT` | Integer | `8` |
| `continue_on_error` | `CONTINUE_ON_ERROR` | Boolean | `true` |
| `enable_cache` | `ENABLE_CACHE` | Boolean | `false` |
| `cache_dir` | `CACHE_DIR` | String | `/tmp/cache` |
| `log_level` | `LOG_LEVEL` | String | `debug` |
| `log_format` | `LOG_FORMAT` | String | `text` |

### Usage Examples

```bash
# Override single option
export SPECS_DIR="./custom/specs"
go run main.go

# Override multiple options
export SPECS_DIR="./specs"
export OUTPUT_DIR="./build"
export WORKER_COUNT="8"
export LOG_LEVEL="debug"
go run main.go

# One-liner for CI/CD
WORKER_COUNT=2 LOG_LEVEL=info go run main.go

# Load from .env file
export $(cat .env | xargs) && go run main.go
```

### Priority Order

Configuration sources are applied in this order (later overrides earlier):

1. **Default values** (hardcoded)
2. **Configuration file** (`resources/application.yml`)
3. **Environment variables** (highest priority)

Example:
```yaml
# application.yml
worker_count: 4
```

```bash
# Override with environment variable
export WORKER_COUNT=8
go run main.go  # Uses 8 workers, not 4
```

## Configuration Examples

### Development Environment

```yaml
# resources/application.yml (development)
specs_dir: "./external/sdk/sdk-packages"
output_dir: "./generated"
target_services: ".*"
worker_count: 4
continue_on_error: true  # Continue to see all errors
enable_cache: true
cache_dir: ".openapi-cache"
spec_file_patterns:
  - "openapi.json"
  - "openapi.yaml"
  - "openapi.yml"
log_level: "debug"  # Verbose logging
log_format: "text"  # Human-readable
```

### Production Environment

```yaml
# resources/application.yml (production)
specs_dir: "/var/lib/openapi/specs"
output_dir: "/var/lib/openapi/generated"
target_services: "^prod-.*-sdk$"  # Production services only
worker_count: 8
continue_on_error: false  # Fail-fast
enable_cache: true
cache_dir: "/var/cache/openapi"
spec_file_patterns:
  - "openapi.json"
log_level: "info"  # Standard logging
log_format: "json"  # Machine-readable
```

### CI/CD Environment

```yaml
# resources/application.yml (CI/CD)
specs_dir: "./external/sdk/sdk-packages"
output_dir: "./generated"
target_services: ".*"
worker_count: 2  # Limited resources
continue_on_error: false  # Fail-fast
enable_cache: true  # Speed up repeated runs
cache_dir: ".openapi-cache"
spec_file_patterns:
  - "openapi.json"
log_level: "info"
log_format: "json"
```

### Testing Environment

```yaml
# resources/application.yml (testing)
specs_dir: "./test/fixtures/specs"
output_dir: "./test/output"
target_services: "test-.*"  # Test services only
worker_count: 1  # Sequential for reproducibility
continue_on_error: false
enable_cache: false  # Always regenerate for tests
cache_dir: "/tmp/openapi-cache"
spec_file_patterns:
  - "openapi.json"
log_level: "debug"
log_format: "text"
```

### Minimal Configuration

```yaml
# resources/application.yml (minimal)
specs_dir: "./specs"
output_dir: "./generated"
# All other options use defaults
```

Equivalent to:
```yaml
specs_dir: "./specs"
output_dir: "./generated"
target_services: ".*"
worker_count: 4
continue_on_error: false
enable_cache: true
cache_dir: ".openapi-cache"
spec_file_patterns:
  - "openapi.json"
  - "openapi.yaml"
  - "openapi.yml"
log_level: "info"
log_format: "json"
```

## Advanced Configuration

### Multi-Environment Setup

Use separate config files for different environments:

```bash
# Directory structure
configs/
├── development.yml
├── staging.yml
└── production.yml
```

```bash
# Load specific config
ln -sf configs/development.yml resources/application.yml
go run main.go
```

Or use environment-specific overrides:

```bash
# development.env
SPECS_DIR=./specs
LOG_LEVEL=debug
LOG_FORMAT=text
WORKER_COUNT=4

# production.env
SPECS_DIR=/var/lib/specs
LOG_LEVEL=info
LOG_FORMAT=json
WORKER_COUNT=8

# Load and run
set -a && source production.env && set +a
go run main.go
```

### Dynamic Configuration

Use environment variables with defaults:

```bash
#!/bin/bash
# run.sh

# Detect environment
ENV=${ENVIRONMENT:-development}

# Set configuration
export SPECS_DIR=${SPECS_DIR:-"./specs"}
export OUTPUT_DIR=${OUTPUT_DIR:-"./generated"}
export WORKER_COUNT=${WORKER_COUNT:-4}
export LOG_LEVEL=${LOG_LEVEL:-"info"}

# Run generator
go run main.go
```

### Configuration Validation

Validate configuration before running:

```bash
#!/bin/bash
# validate-config.sh

# Check specs directory exists
if [ ! -d "$SPECS_DIR" ]; then
  echo "Error: SPECS_DIR '$SPECS_DIR' does not exist"
  exit 1
fi

# Check worker count is valid
if [ "$WORKER_COUNT" -lt 1 ] || [ "$WORKER_COUNT" -gt 32 ]; then
  echo "Error: WORKER_COUNT must be between 1 and 32"
  exit 1
fi

# Check log level is valid
case "$LOG_LEVEL" in
  debug|info|warn|error) ;;
  *) echo "Error: Invalid LOG_LEVEL '$LOG_LEVEL'"; exit 1 ;;
esac

echo "Configuration valid"
```

### Performance Tuning

Optimize for your workload:

```yaml
# Small projects (< 5 specs)
worker_count: 1
enable_cache: false

# Medium projects (5-20 specs)
worker_count: 4
enable_cache: true

# Large projects (20-100 specs)
worker_count: 8
enable_cache: true

# Very large projects (100+ specs)
worker_count: 16
enable_cache: true
```

Benchmark to find optimal settings:

```bash
#!/bin/bash
# benchmark.sh

for workers in 1 2 4 8 16; do
  echo "Testing with $workers workers"
  export WORKER_COUNT=$workers
  time go run main.go
  echo "---"
done
```

### Security Considerations

**Sensitive configurations**:
```yaml
# Do NOT commit sensitive values
database_url: "${DATABASE_URL}"  # Use env var
api_key: "${API_KEY}"             # Use env var
```

**File permissions**:
```bash
# Restrict access to config files
chmod 600 resources/application.yml

# Restrict cache directory
chmod 700 .openapi-cache
```

**Environment variables in CI/CD**:
```yaml
# .gitlab-ci.yml
generate:
  variables:
    SPECS_DIR: "$CI_PROJECT_DIR/specs"
    OUTPUT_DIR: "$CI_PROJECT_DIR/generated"
  script:
    - go run main.go
```

## Next Steps

- Review [Usage Guide](./usage-guide.md) for practical examples
- Check [Troubleshooting Guide](./troubleshooting.md) for common issues
- Read [Architecture Documentation](./architecture.md) for internals
