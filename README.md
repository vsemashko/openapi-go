# OpenAPI Go Generator

[![Go Version](https://img.shields.io/badge/Go-1.24.0+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Coverage](https://img.shields.io/badge/coverage-75.6%25-brightgreen.svg)](https://gitlab.com)

> A production-grade tool for generating Go client code from OpenAPI 3.0 specifications with smart caching, parallel processing, and comprehensive observability.

## ✨ Features

- **🚀 OpenAPI 3.0 Support**: Full support for OpenAPI 3.0.x specifications (JSON, YAML, YML)
- **⚡ Smart Caching**: SHA256-based caching to skip regeneration of unchanged specs
- **🔄 Parallel Processing**: Configurable worker pool for concurrent spec processing
- **📊 Metrics Collection**: Comprehensive performance tracking with JSON export
- **📝 Structured Logging**: Production-grade logging with configurable levels and formats
- **🔌 Extensible Architecture**: Plugin-based post-processor chain for custom transformations
- **🔒 Internal Client Support**: Auto-generate clients without security requirements
- **🎯 Service Filtering**: Regex-based filtering to process specific services
- **⚙️ Flexible Configuration**: YAML config with environment variable overrides
- **🏗️ CI/CD Ready**: Seamless integration with GitLab CI, GitHub Actions, and more

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Usage](#-usage)
- [Configuration](#-configuration)
- [Generated Clients](#-generated-clients)
- [Documentation](#-documentation)
- [Architecture](#-architecture)
- [Contributing](#-contributing)
- [License](#-license)

## 🚀 Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd openapi-go

# Install dependencies
go mod download

# (Optional) Use asdf for version management
asdf install

# Run the generator
go run main.go
```

That's it! Generated clients will be available in `generated/clients/`.

## 📦 Installation

### Prerequisites

- **Go 1.24.0+**: Required for development
- **Git**: For version control and submodules
- **asdf** (optional): For managing tool versions

### Using Go

```bash
# Build from source
go install gitlab.stashaway.com/vladimir.semashko/openapi-go@latest

# Or build locally
git clone <repository-url>
cd openapi-go
go build -o openapi-go
```

### Using Task

```bash
# Install Task (optional)
go install github.com/go-task/task/v3/cmd/task@latest

# Run with Task
task generate-clients
```

## 💻 Usage

### Basic Usage

```bash
# Run with default configuration
go run main.go

# Or use the compiled binary
./openapi-go
```

### Advanced Usage

```bash
# Override configuration with environment variables
export SPECS_DIR="./custom/specs"
export OUTPUT_DIR="./custom/output"
export WORKER_COUNT="8"
export LOG_LEVEL="debug"
go run main.go

# Using Task
task generate-clients

# With custom configuration
WORKER_COUNT=8 task generate-clients
```

### Output

The generator provides real-time feedback:

```
2025-11-22T10:00:00Z INFO Starting OpenAPI client generator
2025-11-22T10:00:00Z INFO Configuration loaded specs_directory=./specs
2025-11-22T10:00:01Z INFO Found 5 OpenAPI specs matching the criteria
2025-11-22T10:00:01Z INFO Processing 5 specs with 4 parallel workers
2025-11-22T10:00:02Z INFO ⚡ Using cached client for fundingsdk (spec unchanged)
2025-11-22T10:00:05Z INFO ✅ Successfully generated client for holidayssdk
...
2025-11-22T10:00:15Z INFO Metrics exported to: ./generated/.openapi-metrics.json
2025-11-22T10:00:15Z INFO Success rate: 100.0%
2025-11-22T10:00:15Z INFO Cache hit rate: 20.0%
```

## ⚙️ Configuration

Configuration is managed via `resources/application.yml` with environment variable overrides.

### Basic Configuration

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
log_level: "info"      # debug, info, warn, error
log_format: "json"     # json, text

# Error handling
continue_on_error: false
```

### Environment Variables

All options can be overridden:

```bash
export SPECS_DIR="./specs"
export OUTPUT_DIR="./output"
export TARGET_SERVICES=".*"
export WORKER_COUNT="8"
export ENABLE_CACHE="true"
export LOG_LEVEL="debug"
export LOG_FORMAT="text"
```

See [Configuration Guide](./docs/configuration.md) for complete details.

## 📦 Generated Clients

### Project Structure

```
generated/
├── clients/                      # Generated client code
│   ├── fundingsdk/              # Funding service client
│   │   ├── oas_client_gen.go
│   │   ├── oas_json_gen.go
│   │   ├── oas_schemas_gen.go
│   │   └── oas_internal_client_gen.go  # Auto-generated
│   ├── holidayssdk/             # Holidays service client
│   └── ...
└── .openapi-metrics.json        # Performance metrics (gitignored)
```

### Using Generated Clients

#### Standard Client (with authentication)

```go
import "yourmodule/generated/clients/fundingsdk"

// For endpoints requiring authentication
client, err := fundingsdk.NewClient(
    "https://api.example.com",
    securitySource,  // Your auth implementation
)
if err != nil {
    log.Fatal(err)
}

// Make API calls
response, err := client.CreateWithdrawal(ctx, req)
```

#### Internal Client (without authentication)

```go
import "yourmodule/generated/clients/fundingsdk"

// For internal-only endpoints (no auth required)
client, err := fundingsdk.NewInternalClient(
    "https://internal-api.example.com",
)
if err != nil {
    log.Fatal(err)
}

// Make API calls
response, err := client.GetUserData(ctx, params)
```

See [Usage Guide](./docs/usage-guide.md) for complete examples.

## 📚 Documentation

### User Documentation

- **[Usage Guide](./docs/usage-guide.md)**: Comprehensive usage instructions with examples
- **[Configuration Guide](./docs/configuration.md)**: Detailed configuration options and examples
- **[Troubleshooting Guide](./docs/troubleshooting.md)**: Common issues and solutions

### Developer Documentation

- **[Contributing Guide](./CONTRIBUTING.md)**: Guidelines for contributing to the project
- **[Architecture Documentation](./docs/architecture.md)**: System design and component overview
- **[Changelog](./CHANGELOG.md)**: Project history and versioned releases

### Additional Resources

- **[Submodule Management](./docs/submodule-management.md)**: Guide for managing git submodules
- **[Future Roadmap](./docs/future-roadmap.md)**: Upcoming features and improvements

## 🏗️ Architecture

### High-Level Overview

```
┌─────────────┐
│   main.go   │  ← Entry point
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  processor  │  ← Orchestration
└──────┬──────┘
       │
   ┌───┴────┬─────────┬──────────┐
   ▼        ▼         ▼          ▼
┌──────┐ ┌─────┐  ┌──────┐  ┌────────┐
│cache │ │ gen │  │worker│  │metrics │
└──────┘ └─────┘  └──────┘  └────────┘
```

### Core Components

- **Processor**: Main orchestration logic
- **Generator**: OpenAPI code generation (ogen v1.14.0)
- **Cache**: SHA256-based spec caching
- **Worker Pool**: Parallel processing
- **Metrics**: Performance tracking
- **Logger**: Structured logging
- **Post-Processors**: Code transformations (internal client, formatting)

See [Architecture Documentation](./docs/architecture.md) for complete details.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](./CONTRIBUTING.md) for:

- Development setup
- Testing guidelines
- Code style conventions
- Pull request process

### Quick Contributor Setup

```bash
# Fork and clone
git clone git@gitlab.stashaway.com:your-username/openapi-go.git
cd openapi-go

# Install dependencies
go mod download
asdf install

# Run tests
go test ./...

# Run linter
golangci-lint run

# Make your changes and submit a PR!
```

## 🎯 CI/CD Integration

### GitLab CI Example

```yaml
stages:
  - generate
  - test

generate-clients:
  stage: generate
  image: golang:1.24
  script:
    - go run main.go
  artifacts:
    paths:
      - generated/

test-clients:
  stage: test
  image: golang:1.24
  script:
    - go test ./generated/clients/...
```

### GitHub Actions Example

```yaml
name: Generate Clients

on: [push]

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - run: go run main.go
```

See [Usage Guide](./docs/usage-guide.md#cicd-integration) for complete examples.

## 📊 Performance & Metrics

The generator automatically tracks and exports metrics:

```json
{
  "total_specs": 10,
  "successful_specs": 10,
  "failed_specs": 0,
  "cached_specs": 2,
  "total_duration_ms": 25340,
  "average_duration_ms": 2534,
  "success_rate": 100.0,
  "cache_hit_rate": 20.0
}
```

**Typical Performance**:
- **First run**: ~3-5 seconds per spec
- **Cached run**: ~1ms per spec (cache hit)
- **Parallel (4 workers)**: ~4x throughput improvement

## 🔐 Security

- **No secrets in config**: Use environment variables
- **Safe file operations**: Validates all paths
- **No arbitrary code execution**: Controlled generator execution
- **Secure defaults**: Fail-fast mode, permission checks

Report security issues privately to the maintainers.

## 📝 License

[Your License Here]

## 👥 Authors & Acknowledgments

- **Original Author**: Vladimir Semashko
- **Contributors**: [List of contributors]
- **Powered by**: [ogen](https://github.com/ogen-go/ogen) - OpenAPI code generator

## 🔗 Links

- **Repository**: [GitLab](https://gitlab.stashaway.com/vladimir.semashko/openapi-go)
- **Documentation**: [docs/](./docs/)
- **Issues**: [Issue Tracker](https://gitlab.stashaway.com/vladimir.semashko/openapi-go/-/issues)
- **Changelog**: [CHANGELOG.md](./CHANGELOG.md)

---

**Need Help?**
- 📖 Read the [Usage Guide](./docs/usage-guide.md)
- 🔍 Check the [Troubleshooting Guide](./docs/troubleshooting.md)
- 💬 Open an [issue](https://gitlab.stashaway.com/vladimir.semashko/openapi-go/-/issues)
- 👥 Reach out to the platform team

**Want to Contribute?**
- 📝 Read the [Contributing Guide](./CONTRIBUTING.md)
- 🏗️ Review the [Architecture Documentation](./docs/architecture.md)
- 🚀 Submit a [pull request](https://gitlab.stashaway.com/vladimir.semashko/openapi-go/-/merge_requests)
