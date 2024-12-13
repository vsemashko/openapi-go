# OpenAPI Go Generator

## Overview

OpenAPI Go Generator is a tool that automates the process of fetching OpenAPI specifications and generating Go stubs, services, and activities from them. It integrates seamlessly into a CI/CD pipeline, ensuring that the generated files are always up to date.

## Features

- Fetch OpenAPI specs from remote URLs.
- Generate Go stubs for OpenAPI clients.
- Create Go service and activity implementations.
- Integrate with GitLab CI/CD for automated updates.
- Automatically commit and push generated files to the repository.

## Repository Structure

```text
.
├── .gitlab-ci.yml            # GitLab CI/CD configuration
├── README.md                # Project documentation
├── go.mod                   # Go module file
├── go.sum                   # Go dependencies
├── main.go                  # Entry point for CLI
├── openapi_spec_urls.json   # JSON file with OpenAPI spec URLs (optional)
├── internal/                # Internal modules
│   ├── processor/           # Combines fetching and generation logic
│   │   └── processor.go
├── generated/               # Generated files
│   ├── stubs/               # Generated Go stubs
│   └── services/            # Generated Go services and activities
```

## Getting Started

### Prerequisites

Install asdf version manager, then run:

```bash
asdf install
```

### Installation

Clone the repository:

```bash
git clone https://github.com/your-org/openapi-go-gen.git
cd openapi-go-gen
```

Install dependencies:

```bash
asdf install
go mod tidy
```

### Configuration

The tool is configured using the `resources/application.yml` file which is included in the repository. The default configuration includes:

```yaml
# Directory containing OpenAPI specs
specs_dir: "./external/sdk/sdk-packages"

# Output directory for generated clients
output_dir: "./generated"

# Regex pattern to filter services
target_services: "(funding-server-sdk|holidays-server-sdk|customer-data-sdk)"
```

#### Overriding Configuration

You can override the configuration using environment variables:

```bash
export SPECS_DIR="./path/to/specs"
export OUTPUT_DIR="./path/to/output"
export TARGET_SERVICES="my-service-sdk"
```

### Usage

Run the generator locally:

```bash
go run main.go
```

This will fetch OpenAPI specs from the URLs defined in `openapi_spec_urls.json` and generate the stubs, services, and activities in the `generated/` directory.

### CI/CD Integration

The project integrates with GitLab CI/CD for automation.

#### GitLab CI/CD Pipeline

1. **Process Specs**:
    - Fetch specs and generate stubs/services.
2. **Commit Changes**:
    - Commit and push the generated files back to the repository.

## Code Generation Process

The code generation process now involves two main steps, all integrated within the Go code:

1. **Filtering External Endpoints**: When generating clients, external endpoints (paths starting with `/external`) are filtered out, keeping only internal endpoints. This is done in the `internal/processor/processor.go` file.

2. **Adding Internal Client Support**: After the clients are generated, the `NewInternalClient` function is automatically added to all SDK clients. This allows internal services to use the clients without requiring a security source. This is handled by the `internal/postprocess/postprocess.go` module.

## Task Commands

You can run the following task commands to generate and process clients:

```bash
# Generate clients (includes filtering and adding NewInternalClient)
task generate-clients

# Run the full process (generate clients, go mod tidy)
task
```

## Generated Client Features

### Internal-Only Endpoints

When generating the SDKs, the code generator filters out external endpoints and only includes internal endpoints. This is useful for services that only need to access other services' internal endpoints without exposing external endpoints.

### Client Initialization Options

The generated clients provide the following initialization methods:

#### For all SDK clients

```go
// For all endpoints with security 
client, err := sdkName.NewClient("https://service-url.example.com", securitySource)

// For internal endpoints only without security
client, err := sdkName.NewInternalClient("https://service-url.example.com")
```

The `NewInternalClient` function provides a way to initialize the client without requiring a security source, which is useful for services that only need to access internal endpoints.
