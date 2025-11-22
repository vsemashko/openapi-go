# Changelog

All notable changes to the OpenAPI Go Generator project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- OpenAPI 3.1 support (monitoring ogen library status)
- Multiple generator support (alternative to ogen)
- Plugin system for post-processors
- WebSocket/gRPC support
- Incremental generation optimization

## [2.1.0] - 2025-11-22

### Added - Validation & YAML Support

#### YAML Spec Parsing Fix (Critical Bug Fix)
- **Fixed critical bug** in `internal/spec/parser.go` where YAML specs were discovered but not parsed correctly
- Added proper YAML unmarshaling using `gopkg.in/yaml.v3`
- Added dual `json` and `yaml` struct tags to all OpenAPI spec types
- Implemented format detection by file extension (.json, .yaml, .yml)
- Added fallback parsing for unknown extensions (tries JSON first, then YAML)
- **Impact**: Security detection now works correctly for YAML specs

#### Comprehensive Spec Validator (New Feature)
- Added new `internal/validator` package with full validation framework
- Validates OpenAPI 3.0.x specifications before code generation
- Supports both JSON and YAML formats
- Detects and reports security scheme configuration
- Configurable validation rules with custom linting
- Error codes for all validation failures
- Warning system with optional fail-on-warnings mode
- Strict mode for enhanced validation checks
- Integrated into generation pipeline with early failure detection
- Test coverage: 100% for validator package

#### Validator Features
- **File validation**: Checks file exists, is readable, and is not a directory
- **Format detection**: Auto-detects JSON vs YAML from extension
- **Version validation**: Validates OpenAPI version (3.0.x fully supported, 3.1.x partial with warnings)
- **Info validation**: Validates required info section (title, version)
- **Security detection**: Extracts and reports security scheme information
- **Custom rules**: Support for custom validation rules (require-description, require-contact, require-license)
- **Rule ignoring**: Ability to ignore specific validation rules by code
- **Detailed reporting**: Formatted validation results with error/warning codes and messages

#### Configuration Updates
- Added `validator` configuration section to `Config` struct
- Added `ValidatorConfig` with options: enabled, fail_on_warnings, strict_mode, custom_rules, ignored_rules
- Default: validator enabled, fail-on-warnings disabled, strict mode disabled
- Environment variable support: `VALIDATOR_ENABLED`, `VALIDATOR_FAIL_ON_WARNINGS`, `VALIDATOR_STRICT_MODE`

#### Testing
- Added 5 new YAML parsing tests in `internal/spec/parser_test.go`
- Added comprehensive validator tests (16 test cases) in `internal/validator/validator_test.go`
- All tests passing (100% success rate)
- Test coverage increased for spec package

#### Documentation
- Added comprehensive **Validator Guide** (`docs/validator-guide.md`)
- Documented all validation rules and error codes
- Added configuration examples and best practices
- Included troubleshooting section for common validation issues
- Added examples of valid and invalid specs

### Fixed
- **CRITICAL**: YAML spec parsing now works correctly (was only doing JSON unmarshaling)
- **CRITICAL**: Security detection now works for YAML/YML spec files
- NewInternalClient generation now correct for YAML specs with security

### Research
- **OpenAPI 3.1 Support Status** in ogen-go (as of Nov 2025):
  - ogen-go primarily supports OpenAPI 3.0.x
  - Partial OpenAPI 3.1 support (some specs work, some features not supported)
  - Full 3.1 support not yet complete
  - Recommendation: Continue using OpenAPI 3.0.x for best compatibility
  - Monitor: https://github.com/ogen-go/ogen/issues for updates

## [2.0.0] - 2025-11-22

### Added - Documentation & Polish (Phase 3)

#### User Documentation
- **Usage Guide** (`docs/usage-guide.md`): Comprehensive guide covering quick start, basic/advanced usage, generated client usage, CI/CD integration, and best practices
- **Configuration Guide** (`docs/configuration.md`): Detailed documentation of all configuration options, environment variables, and configuration examples for different environments
- **Troubleshooting Guide** (`docs/troubleshooting.md`): Common issues with solutions, error message explanations, performance troubleshooting, debugging tips, and metrics analysis

#### Developer Documentation
- **Contributing Guide** (`CONTRIBUTING.md`): Guidelines for contributing, development setup, workflow, testing, code style, commit conventions, and PR process
- **Architecture Documentation** (`docs/architecture.md`): System architecture overview, core components, data flow, design patterns, concurrency model, and extension points
- **Changelog** (`CHANGELOG.md`): Complete project history with versioned releases

### Added - Enhanced Observability (Phase 2)

#### Structured Logging (Phase 2.1)
- Implemented production-grade structured logging using Go's standard `log/slog`
- Created `internal/logger` package with comprehensive test coverage (93.1%)
- Added configurable log levels: debug, info, warn, error
- Added configurable log formats: JSON (machine-readable) and text (human-readable)
- Smart configuration logging with fallback support (`config_logging.go`)
- Updated `main.go` to use structured logger throughout
- Environment variable overrides: `LOG_LEVEL`, `LOG_FORMAT`

#### Metrics Collection (Phase 2.2)
- Created `internal/metrics` package for comprehensive performance tracking
- Implemented `Collector` with thread-safe concurrent recording using `sync.RWMutex`
- Added per-spec metrics tracking: duration, success/failure, cached status, errors
- Aggregated metrics: total/successful/failed/cached counts, average duration
- Calculated metrics: success rate %, cache hit rate %
- JSON export to `.openapi-metrics.json` with detailed per-spec data
- Integrated metrics into both parallel and sequential generation paths
- Added metrics to `.gitignore`
- Test coverage: 95.7%

### Added - Technical Debt Cleanup (Phase 1)

#### Phase 1.1: Dead Code Removal
- Removed unused `internal/preprocessor` module (59 lines of commented code)
- Cleaned up 0% test coverage dead code
- Simplified codebase and reduced confusion

#### Phase 1.2: YAML Spec Support
- Added `SpecFilePatterns` configuration option (array of strings)
- Updated `findOpenAPISpecs()` to accept multiple file patterns
- Support for JSON, YAML, and YML formats: `openapi.json`, `openapi.yaml`, `openapi.yml`
- Created comprehensive test fixtures for YAML formats
- Added 5 new test cases covering YAML/YML matching
- Updated `resources/application.yml` with default patterns

#### Phase 1.3: Git Submodule Deterministic Builds
- Fixed `Taskfile.yml` to respect pinned git submodule commits
- Changed `init` task from `git submodule update --remote` to `git submodule update --init --recursive`
- Created separate `update-submodule` task for explicit updates
- Added comprehensive `docs/submodule-management.md` guide
- Ensures reproducible builds by respecting commit pins

#### Phase 1.4: CI/CD Test Execution
- Added test, build, and update stages to `.gitlab-ci.yml`
- Implemented `test:unit` job with coverage reporting
- Added coverage threshold enforcement (75% minimum)
- Implemented `test:lint` job with golangci-lint
- Added `build:verify` job depending on tests passing
- Coverage artifacts exported to GitLab

#### Phase 1.5: Go Version Synchronization
- Updated `.tool-versions` to Go 1.24.0
- Updated `.gitlab-ci.yml` to use Go 1.24.0 image
- Ensured consistency across all version management files

### Changed
- Logging now uses structured `slog` instead of standard `log.Printf`
- Configuration logging is now environment-aware (structured vs fallback)
- Generation process now exports comprehensive metrics automatically
- Tests updated to use metrics collector
- Build process enforces code coverage and linting

## [1.0.0] - 2025-11-15

### Added - Robustness & Testing (Pre-Phase Improvements)

#### Phase 3.4: Caching System
- Implemented SHA256-based caching system in `internal/cache`
- Cache validation based on spec hash and generator version
- Automatic cache invalidation when specs change
- Cache pruning to remove invalid entries
- Significant performance improvements for repeated runs
- Test coverage: 81.1%

#### Phase 3.3: Parallel Processing
- Implemented worker pool pattern in `internal/worker`
- Configurable worker count via `worker_count` configuration
- Context-aware cancellation support
- Concurrent spec processing with result collection
- Performance improvements for multi-spec generation
- Test coverage: 91.5%

#### Phase 3.2: Post-Processor Chain
- Implemented Chain of Responsibility pattern in `internal/postprocessor`
- Created pluggable post-processor architecture
- Added `InternalClientGenerator` processor (generates `NewInternalClient`)
- Added `GoFormatter` processor (formats generated code)
- Extensible chain for future processors
- Test coverage: 89.5%

#### Phase 3.1: Generator Interface
- Created `internal/generator` interface abstraction
- Implemented Ogen generator with version management
- Automatic generator installation with verification
- Configurable generation options
- Support for multiple generators (extensible)
- Test coverage: 66.2%

#### Phase 2: Comprehensive Testing
- Achieved 75.6% overall test coverage
- Added comprehensive unit tests for all internal packages
- Added integration tests for full generation workflow
- Added test fixtures for realistic test scenarios
- Improved error handling and edge case coverage
- Added coverage reporting to CI/CD

#### Phase 1: Critical Robustness Improvements
- Enhanced error handling with wrapped errors and context
- Added input validation for all configuration options
- Improved spec finding with better error messages
- Added service name normalization and validation
- Enhanced directory operations with safety checks
- Better logging throughout the codebase

### Added - Core Features (Initial Release)

#### OpenAPI Client Generation
- Support for OpenAPI 3.0.x specifications (JSON format initially)
- Automatic Go client code generation using ogen v1.14.0
- Service filtering with regex patterns
- Configurable output directories
- Clean generated code with proper formatting

#### Internal Client Support
- Automatic generation of `NewInternalClient` function
- Support for internal-only endpoints without security requirements
- Security detection in OpenAPI specs
- Seamless integration with post-processor chain

#### Configuration Management
- YAML configuration file support (`resources/application.yml`)
- Environment variable overrides for all options
- Sensible defaults for all configuration values
- Flexible configuration precedence

#### CI/CD Integration
- GitLab CI/CD pipeline support
- Automated client generation
- Git integration for tracking changes
- Support for automated updates

### Documentation
- Initial README with basic usage
- Implementation roadmap documentation
- Testing strategy documentation
- Analysis and architectural improvements documentation

## Version History

### [2.0.0] - 2025-11-22
**Focus**: Documentation, Observability, and Polish
- Comprehensive user and developer documentation
- Structured logging with slog
- Metrics collection and export
- Technical debt cleanup (dead code, YAML support, submodule fixes)

### [1.0.0] - 2025-11-15
**Focus**: Robustness, Testing, and Architecture
- Caching system for performance
- Parallel processing with worker pool
- Post-processor chain for extensibility
- Generator interface abstraction
- Comprehensive test coverage (75.6%)
- Critical robustness improvements

### [0.1.0] - Initial Development
**Focus**: Core functionality
- Basic OpenAPI 3.0 client generation
- Ogen integration
- Internal client support
- Configuration management
- GitLab CI/CD integration

## Migration Guides

### Migrating from 1.x to 2.x

#### Configuration Changes
```yaml
# NEW: Add spec file patterns for YAML support
spec_file_patterns:
  - "openapi.json"
  - "openapi.yaml"
  - "openapi.yml"

# NEW: Add logging configuration
log_level: "info"
log_format: "json"
```

#### Code Changes
- No breaking changes to generated client code
- Metrics are now automatically exported to `.openapi-metrics.json`
- Logging format changed to structured (can be configured)

#### CI/CD Changes
```yaml
# NEW: CI now enforces test coverage
test:unit:
  script:
    - go test ./... -coverprofile=coverage.out
    - go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//' | awk '{if ($1 < 75) exit 1}'
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on:
- Development setup
- Testing requirements
- Code style
- Commit conventions
- Pull request process

## Support

- **Documentation**: [docs/](./docs/)
- **Issues**: GitLab issue tracker
- **Questions**: Reach out to the platform team

## License

[Your License Here]

## Authors

- Original implementation: Vladimir Semashko
- Contributors: [List of contributors]

---

For detailed information about specific changes, see:
- [Usage Guide](./docs/usage-guide.md)
- [Configuration Guide](./docs/configuration.md)
- [Architecture Documentation](./docs/architecture.md)
- [Troubleshooting Guide](./docs/troubleshooting.md)
