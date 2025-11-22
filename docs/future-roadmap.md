# OpenAPI Go Generator - Future Roadmap

## Overview

This document outlines the planned enhancements and future direction for the OpenAPI Go Generator project. Our goal is to maintain a production-grade, extensible, and high-performance code generation tool while expanding support for newer OpenAPI versions and additional features.

## Table of Contents

1. [Short-Term Goals (Q1-Q2 2026)](#short-term-goals-q1-q2-2026)
2. [Medium-Term Goals (Q3-Q4 2026)](#medium-term-goals-q3-q4-2026)
3. [Long-Term Vision (2027+)](#long-term-vision-2027)
4. [OpenAPI 3.1 Support](#openapi-31-support)
5. [Community Requests](#community-requests)

---

## Short-Term Goals (Q1-Q2 2026)

### 1. OpenAPI 3.1 Support ⭐ **HIGH PRIORITY**

**Status**: ❌ **BLOCKED** - Research Complete (2025-11-22)

**Objective**: Add full support for OpenAPI 3.1.x specifications.

**Blocker Identified**:
The ogen code generator (v1.14.0 current, v1.16.0 latest) does **NOT support OpenAPI 3.1** specifications. Both versions fail to parse 3.1 syntax due to breaking changes:
- `nullable: true` → `type: ["string", "null"]` (type arrays not supported)
- `exclusiveMinimum: true` → `exclusiveMinimum: 0` (numeric values not supported)
- No roadmap or tracking issue found for 3.1 support in ogen

**Research Report**: See [OpenAPI 3.1 Research Report](./openapi-3.1-research-report.md) for full analysis.

**Current Action**: **Monitor & Wait**
- Monitor ogen releases monthly: https://github.com/ogen-go/ogen/releases
- Continue with other roadmap items (Enhanced Error Handling, Enhanced Metrics)
- Revisit when ogen adds 3.1 support

**Alternative Options** (if ogen never supports 3.1):
1. **Spec Downgrader**: Build 3.1 → 3.0 converter (2-3 weeks effort)
2. **Multi-Generator**: Add alternative generator with 3.1 support (4-6 weeks, but no viable generator exists yet)
3. **Contribute to ogen**: Add 3.1 support upstream (2-3 months effort)

**Key Features** (when unblocked):
- JSON Schema 2020-12 support
- Webhook support
- Improved discriminator handling
- New JSON Schema keywords (`prefixItems`, `unevaluatedProperties`, etc.)
- `$ref` siblings support
- Enhanced anyOf/oneOf/allOf handling

**Timeline**: Unknown - depends on ogen library

**Dependencies**:
- ❌ **BLOCKER**: ogen library OpenAPI 3.1 support
- Monitoring: https://github.com/ogen-go/ogen/issues

**Next Review**: 2026-01-01 (monthly check)

---

### 2. Enhanced Validation & Linting

**Status**: ✅ **COMPLETED** - v2.1.0 (2025-11-22)

**Objective**: Add built-in spec validation and linting before generation.

**Features**:
- OpenAPI spec validation using standard validators
- Custom linting rules for common issues
- Pre-generation validation to catch errors early
- Configurable validation strictness
- Detailed validation reports

**Implementation**:
```go
// internal/validator/validator.go
type Validator interface {
    Validate(specPath string) (*ValidationResult, error)
}

type ValidationResult struct {
    Valid      bool
    Errors     []ValidationError
    Warnings   []ValidationWarning
    Spec       SpecInfo
}

// Configuration
type ValidatorConfig struct {
    Enabled          bool     `yaml:"enabled"`
    FailOnWarnings   bool     `yaml:"fail_on_warnings"`
    CustomRules      []string `yaml:"custom_rules"`
    IgnoredRules     []string `yaml:"ignored_rules"`
}
```

**Benefits**:
- Catch spec issues before generation
- Better error messages
- Reduced generation failures
- Improved developer experience

---

### 3. Incremental Generation

**Status**: ✅ **COMPLETED** - v2.2.0 (2025-11-22)

**Objective**: Support incremental generation to regenerate only changed endpoints.

**Features**:
- Fine-grained change detection (endpoint-level)
- Partial regeneration of affected files
- Faster iteration during development
- Preserve manual customizations where possible

**Implementation Approach**:
- Hash individual endpoints/operations
- Track changes at path/method level
- Regenerate only affected files
- Merge with existing generated code

**Performance Benefits**:
- 5-10x faster for small changes
- Reduced compilation time
- Better development workflow

---

## Medium-Term Goals (Q3-Q4 2026)

### 4. Multiple Generator Support

**Status**: Planned

**Objective**: Support alternative code generators beyond ogen.

**Target Generators**:
1. **ogen** (current): Fast, type-safe, OpenAPI 3.0/3.1
2. **oapi-codegen**: Popular alternative with different features
3. **go-swagger**: Mature, feature-rich, wide adoption
4. **Custom generators**: Plugin support for custom implementations

**Configuration**:
```yaml
generator:
  name: "ogen"          # ogen, oapi-codegen, go-swagger, custom
  version: "v1.14.0"
  options:
    strict: true
    generate_tests: false
```

**Benefits**:
- Flexibility to choose best generator for use case
- Fallback options if primary generator fails
- Compare generated code quality across generators
- Support legacy projects using different generators

---

### 5. Plugin System for Post-Processors

**Status**: Planned

**Objective**: Create a plugin system for custom post-processors.

**Features**:
- Dynamic plugin loading
- Plugin configuration
- Plugin marketplace/registry
- Built-in plugins for common tasks

**Plugin Interface**:
```go
// Plugin interface
type Plugin interface {
    Name() string
    Version() string
    Process(ctx context.Context, input ProcessInput) error
    Configure(config map[string]interface{}) error
}

// Plugin configuration
type PluginConfig struct {
    Name    string
    Enabled bool
    Config  map[string]interface{}
}
```

**Example Plugins**:
- **Mock generator**: Generate mock implementations for testing
- **Documentation generator**: Generate API documentation from specs
- **Client wrappers**: Generate higher-level wrapper clients
- **Custom transformations**: Project-specific code modifications

---

### 6. Enhanced Metrics & Observability

**Status**: Planned

**Objective**: Expand metrics collection and add monitoring integrations.

**Features**:
- Prometheus metrics exporter
- OpenTelemetry tracing
- Performance profiling
- Resource usage tracking
- Historical metrics and trends

**Metrics Exporters**:
```yaml
metrics:
  enabled: true
  exporters:
    - type: "json"
      path: ".openapi-metrics.json"
    - type: "prometheus"
      port: 9090
      path: "/metrics"
    - type: "opentelemetry"
      endpoint: "http://otel-collector:4318"
```

**New Metrics**:
- Memory usage per spec
- CPU time per operation
- Network I/O for remote specs
- Generator execution time breakdown
- Cache efficiency metrics

---

## Long-Term Vision (2027+)

### 7. Multi-Language Support

**Status**: Future

**Objective**: Generate clients for multiple programming languages.

**Target Languages**:
- **Go** (current)
- **TypeScript/JavaScript**
- **Python**
- **Rust**
- **Java/Kotlin**

**Architecture**:
- Abstract generator interface
- Language-specific implementations
- Shared spec parsing and validation
- Language-specific post-processors

**Configuration**:
```yaml
languages:
  - name: "go"
    generator: "ogen"
    output_dir: "./generated/go"
  - name: "typescript"
    generator: "openapi-typescript"
    output_dir: "./generated/ts"
```

---

### 8. WebSocket & gRPC Support

**Status**: Future

**Objective**: Support WebSocket operations and gRPC transcoding.

**WebSocket Support**:
- OpenAPI 3.1 webhook definitions
- Bidirectional streaming clients
- Connection management
- Reconnection logic

**gRPC Support**:
- gRPC transcoding from OpenAPI specs
- Protocol buffer generation
- Streaming RPC support
- gRPC-gateway integration

---

### 9. Cloud-Native Features

**Status**: Future

**Objective**: Add cloud-native capabilities for modern deployments.

**Features**:
- **Service mesh integration**: Istio, Linkerd configuration
- **Kubernetes operators**: CRDs for spec management
- **Serverless support**: Lambda, Cloud Functions adapters
- **Distributed tracing**: Auto-instrumented clients
- **Circuit breakers**: Built-in resilience patterns

---

### 10. Interactive CLI & UI

**Status**: Future

**Objective**: Create interactive tools for easier usage.

**CLI Features**:
- Interactive configuration wizard
- Progress bars and rich output
- Interactive spec selection
- Diff view for changes
- Live reload during development

**Web UI**:
- Visual spec browser
- Generation dashboard
- Metrics visualization
- Configuration editor
- Spec comparison tool

---

## OpenAPI 3.1 Support

### Detailed Feature Breakdown

#### 1. JSON Schema 2020-12 Support

OpenAPI 3.1 aligns with JSON Schema 2020-12, bringing new keywords and enhanced validation:

**New Keywords**:
- `prefixItems`: Tuple validation for arrays
- `unevaluatedProperties`: Stricter object validation
- `unevaluatedItems`: Stricter array validation
- `dependentSchemas`: Conditional schema application
- `$dynamicRef`/`$dynamicAnchor`: Dynamic references

**Example**:
```json
{
  "type": "object",
  "properties": {
    "type": {"const": "person"},
    "name": {"type": "string"}
  },
  "if": {
    "properties": {"type": {"const": "person"}}
  },
  "then": {
    "required": ["name"]
  }
}
```

**Code Generation Impact**:
- More precise type generation
- Better validation in generated clients
- Improved error messages

---

#### 2. Webhooks

OpenAPI 3.1 adds first-class webhook support:

**Specification**:
```yaml
webhooks:
  newUser:
    post:
      requestBody:
        description: New user webhook
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '200':
          description: Success
```

**Generated Code**:
```go
// Webhook handler interface
type WebhookHandler interface {
    HandleNewUser(ctx context.Context, user User) error
}

// Webhook server
type WebhookServer struct {
    handler WebhookHandler
}

func (s *WebhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Handle webhook
}
```

---

#### 3. Enhanced Discriminators

OpenAPI 3.1 improves discriminator handling:

**Specification**:
```yaml
components:
  schemas:
    Pet:
      oneOf:
        - $ref: '#/components/schemas/Dog'
        - $ref: '#/components/schemas/Cat'
      discriminator:
        propertyName: petType
        mapping:
          dog: '#/components/schemas/Dog'
          cat: '#/components/schemas/Cat'
```

**Generated Code**:
```go
// More precise discriminated unions
type Pet interface {
    isPet()
}

type Dog struct { /* ... */ }
func (Dog) isPet() {}

type Cat struct { /* ... */ }
func (Cat) isPet() {}

// Type-safe deserialization
func DecodePet(data []byte) (Pet, error) {
    // Discriminator-based routing
}
```

---

#### 4. $ref Siblings

OpenAPI 3.1 allows `$ref` to have siblings:

**Before (OpenAPI 3.0)**:
```yaml
# Not allowed - description is ignored
schema:
  $ref: '#/components/schemas/User'
  description: "User object"
```

**After (OpenAPI 3.1)**:
```yaml
# Allowed - description applies to referenced schema
schema:
  $ref: '#/components/schemas/User'
  description: "User object"
  examples:
    - name: "John Doe"
```

**Impact**: More flexible schema composition and documentation.

---

### Migration Strategy (3.0 → 3.1)

#### 1. Backward Compatibility

**Approach**:
- Maintain OpenAPI 3.0 support indefinitely
- Auto-detect spec version
- Provide migration tools

**Configuration**:
```yaml
openapi_version: "auto"  # auto, 3.0, 3.1
strict_validation: false  # Allow 3.0 specs with 3.1 features
```

#### 2. Migration Guide

**Step 1**: Validate 3.0 specs
```bash
openapi-go validate --version 3.0
```

**Step 2**: Preview 3.1 conversion
```bash
openapi-go migrate --from 3.0 --to 3.1 --dry-run
```

**Step 3**: Migrate specs
```bash
openapi-go migrate --from 3.0 --to 3.1 --output ./specs-3.1/
```

**Step 4**: Regenerate clients
```bash
openapi-go generate --openapi-version 3.1
```

#### 3. Breaking Changes

**Potential Breaking Changes**:
1. Stricter validation (may reject previously valid specs)
2. Different type generation for some schemas
3. New required fields in generated structs
4. Changes to discriminator handling

**Mitigation**:
- Gradual migration path
- Compatibility flags
- Detailed migration documentation
- Automated migration tools

---

## Community Requests

### Frequently Requested Features

#### 1. Custom Template Support
- Allow custom Go templates for code generation
- Override default templates selectively
- Template marketplace for sharing

#### 2. Mock Server Generation
- Generate mock servers from OpenAPI specs
- Configurable response generation
- Testing and development support

#### 3. API Versioning Support
- Handle multiple API versions
- Version-specific client generation
- Version negotiation in clients

#### 4. Enhanced Error Handling
- Structured error types
- Retry policies
- Circuit breaker patterns
- Timeout configuration

#### 5. Request/Response Middleware
- Interceptor support
- Logging middleware
- Authentication middleware
- Custom transformations

---

## Implementation Priorities

### Priority Matrix

| Feature | Impact | Effort | Priority | Timeline |
|---------|--------|--------|----------|----------|
| OpenAPI 3.1 Support | ⭐⭐⭐⭐⭐ | High | **P0** | Q1-Q2 2026 |
| Enhanced Validation | ⭐⭐⭐⭐ | Medium | **P1** | Q2 2026 |
| Incremental Generation | ⭐⭐⭐⭐ | Medium | **P1** | Q2 2026 |
| Multiple Generators | ⭐⭐⭐ | High | **P2** | Q3-Q4 2026 |
| Plugin System | ⭐⭐⭐ | High | **P2** | Q3-Q4 2026 |
| Enhanced Metrics | ⭐⭐⭐ | Low | **P2** | Q4 2026 |
| Multi-Language | ⭐⭐⭐⭐⭐ | Very High | **P3** | 2027+ |
| WebSocket/gRPC | ⭐⭐⭐ | High | **P3** | 2027+ |
| Cloud-Native | ⭐⭐⭐ | Medium | **P3** | 2027+ |
| Interactive UI | ⭐⭐ | Medium | **P4** | TBD |

---

## Contributing to the Roadmap

We welcome community input on our roadmap! Here's how you can contribute:

### 1. Feature Requests

Open an issue with the "enhancement" label describing:
- **Use case**: Why is this feature needed?
- **Proposed solution**: How should it work?
- **Alternatives**: Other approaches considered
- **Priority**: How important is this to you?

### 2. Roadmap Discussions

Join roadmap discussions in:
- GitHub Discussions
- Issue tracker
- Community meetings (TBD)

### 3. Implementation Help

Want to help implement a feature?
1. Comment on the relevant issue
2. Review the implementation plan
3. Submit a PR with tests and documentation
4. Coordinate with maintainers

---

## Success Metrics

We track these metrics to measure roadmap success:

### Technical Metrics
- **OpenAPI Coverage**: % of OpenAPI 3.0/3.1 features supported
- **Generation Speed**: Time to generate clients
- **Cache Hit Rate**: % of generations using cache
- **Test Coverage**: % of code covered by tests

### User Experience Metrics
- **Adoption Rate**: # of projects using the generator
- **Issue Resolution Time**: Average time to fix bugs
- **Documentation Quality**: User satisfaction with docs
- **Community Engagement**: # of contributors and discussions

### Performance Metrics
- **Generation Time**: Trend over time
- **Memory Usage**: Peak memory during generation
- **CPU Efficiency**: CPU time per spec
- **Build Size**: Size of generated clients

---

## Feedback & Updates

This roadmap is a living document and will be updated based on:
- Community feedback
- Changing requirements
- Technical discoveries
- Resource availability

**Last Updated**: 2025-11-22

**Next Review**: Q1 2026

---

## Resources

- **OpenAPI 3.1 Specification**: https://spec.openapis.org/oas/v3.1.0
- **JSON Schema 2020-12**: https://json-schema.org/draft/2020-12/release-notes
- **Ogen Generator**: https://github.com/ogen-go/ogen
- **Project Repository**: https://gitlab.stashaway.com/vladimir.semashko/openapi-go
- **Issue Tracker**: https://gitlab.stashaway.com/vladimir.semashko/openapi-go/-/issues

---

**Questions?**
- Open a discussion in the issue tracker
- Reach out to the platform team
- Join community meetings (TBD)

**Want to contribute?**
- See [Contributing Guide](../CONTRIBUTING.md)
- Check [open issues](https://gitlab.stashaway.com/vladimir.semashko/openapi-go/-/issues)
- Review the [Architecture Documentation](./architecture.md)
