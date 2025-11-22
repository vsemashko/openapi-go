# OpenAPI Spec Validator Guide

## Overview

The OpenAPI Spec Validator is a built-in feature that validates OpenAPI specifications before code generation. This helps catch errors early, ensures spec quality, and provides detailed feedback about potential issues.

## Features

- ✅ Validates OpenAPI 3.0.x specifications
- ✅ Supports both JSON and YAML formats
- ✅ Detects security scheme configuration
- ✅ Configurable validation rules
- ✅ Custom linting rules
- ✅ Warning and error reporting
- ✅ Strict mode for enhanced validation

## Configuration

### Basic Configuration

Add validator settings to your `resources/application.yml`:

```yaml
validator:
  enabled: true                    # Enable/disable validation (default: true)
  fail_on_warnings: false          # Treat warnings as errors (default: false)
  strict_mode: false               # Enable strict validation (default: false)
  custom_rules: []                 # List of custom rules to apply
  ignored_rules: []                # List of rule codes to ignore
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `true` | Enable or disable spec validation |
| `fail_on_warnings` | boolean | `false` | Treat warnings as validation failures |
| `strict_mode` | boolean | `false` | Enable additional strict validation checks |
| `custom_rules` | array | `[]` | Custom validation rules to apply |
| `ignored_rules` | array | `[]` | Validation error/warning codes to ignore |

### Environment Variables

You can override configuration using environment variables:

```bash
export VALIDATOR_ENABLED=true
export VALIDATOR_FAIL_ON_WARNINGS=false
export VALIDATOR_STRICT_MODE=true
```

## Validation Rules

### Built-in Validation

The validator automatically checks:

1. **File Accessibility**
   - File exists and is readable
   - Path is a file (not a directory)
   - Code: `FILE_NOT_FOUND`, `FILE_ACCESS_ERROR`, `NOT_A_FILE`

2. **Spec Format**
   - Valid JSON or YAML syntax
   - Auto-detects format from file extension
   - Code: `PARSE_ERROR`, `UNKNOWN_FORMAT`

3. **OpenAPI Version**
   - Required `openapi` field is present
   - Version format is valid (e.g., `3.0.0`)
   - Version is supported (3.0.x fully supported, 3.1.x partial)
   - Code: `MISSING_OPENAPI_VERSION`, `INVALID_VERSION_FORMAT`, `UNSUPPORTED_VERSION`, `UNKNOWN_VERSION`

4. **Info Section**
   - Required `info` section is present
   - Required `title` and `version` fields exist
   - Fields are not empty
   - Code: `MISSING_INFO`, `MISSING_TITLE`, `MISSING_VERSION`, `EMPTY_TITLE`, `EMPTY_VERSION`

5. **Security Detection**
   - Extracts security scheme information
   - Identifies security requirements
   - Reports security scheme count and names

### Custom Rules

Enable additional validation checks with custom rules:

```yaml
validator:
  custom_rules:
    - require-description    # Require info.description field
    - require-contact        # Require info.contact field
    - require-license        # Require info.license field
```

Available custom rules:
- `require-description`: API description is recommended
- `require-contact`: Contact information is recommended
- `require-license`: License information is recommended

### Strict Mode

When strict mode is enabled, additional warnings are generated:

```yaml
validator:
  strict_mode: true
```

Strict mode checks:
- **NO_SECURITY**: Warns if no security schemes are defined

### Ignoring Rules

Suppress specific warnings or errors by adding their codes to `ignored_rules`:

```yaml
validator:
  ignored_rules:
    - NO_SECURITY           # Don't warn about missing security
    - EMPTY_TITLE           # Allow empty titles
    - MISSING_DESCRIPTION   # Don't require descriptions
```

## Validation Output

### Success Example

```
2025-11-22 10:00:00 Validating 3 OpenAPI specs...
2025-11-22 10:00:00 All specs validated successfully
```

### Error Example

```
Validation Result for: /path/to/spec.json
  Format: json
  OpenAPI Version: 2.0
  Title: Old API
  Has Security: false

❌ Errors (1):
  1. [UNSUPPORTED_VERSION] openapi: Swagger 2.0 is not supported. Please use OpenAPI 3.0 or convert your spec.

❌ Validation: FAILED
```

### Warning Example

```
Validation Result for: /path/to/spec.yaml
  Format: yaml
  OpenAPI Version: 3.0.0
  Title: Test API
  Has Security: false

⚠️  Warnings (2):
  1. [NO_SECURITY] security: No security schemes defined. Consider adding authentication.
  2. [MISSING_DESCRIPTION] info.description: Description is recommended but missing

✅ Validation: PASSED
```

## Integration with Generation Pipeline

Validation runs automatically before code generation:

```
1. Find OpenAPI specs
2. Validate specs (if enabled)  ← You are here
3. Initialize cache
4. Generate clients in parallel
5. Export metrics
```

### Behavior

- **Validation enabled**: All specs are validated before generation
- **Validation fails**: Generation stops (unless `continue_on_error: true`)
- **Warnings only**: Generation continues (unless `fail_on_warnings: true`)
- **Validation disabled**: Skip directly to generation

## Error Codes Reference

### Critical Errors (Block Generation)

| Code | Field | Description |
|------|-------|-------------|
| `FILE_NOT_FOUND` | file | Spec file does not exist |
| `FILE_ACCESS_ERROR` | file | Cannot read spec file |
| `NOT_A_FILE` | file | Path is a directory, not a file |
| `PARSE_ERROR` | spec | Invalid JSON or YAML syntax |
| `MISSING_OPENAPI_VERSION` | openapi | Required 'openapi' field missing |
| `UNSUPPORTED_VERSION` | openapi | OpenAPI version not supported (e.g., Swagger 2.0, OpenAPI 3.1 with limitations) |
| `MISSING_INFO` | info | Required 'info' section missing |
| `MISSING_TITLE` | info.title | Required 'title' field missing |
| `MISSING_VERSION` | info.version | Required 'version' field missing |

### Warnings (Don't Block Generation)

| Code | Field | Description |
|------|-------|-------------|
| `UNKNOWN_FORMAT` | format | Unknown file extension, auto-detection will be used |
| `INVALID_VERSION_FORMAT` | openapi | Version format may be invalid |
| `UNKNOWN_VERSION` | openapi | Unrecognized OpenAPI version |
| `EMPTY_TITLE` | info.title | Title field is empty |
| `EMPTY_VERSION` | info.version | Version field is empty |
| `NO_SECURITY` | security | No security schemes defined (strict mode only) |
| `MISSING_DESCRIPTION` | info.description | Description recommended but missing |
| `MISSING_CONTACT` | info.contact | Contact info recommended but missing |
| `MISSING_LICENSE` | info.license | License info recommended but missing |

## Best Practices

### 1. Enable Validation in CI/CD

Always validate specs in your CI/CD pipeline:

```yaml
# .gitlab-ci.yml
validate:
  stage: validate
  script:
    - export VALIDATOR_ENABLED=true
    - export VALIDATOR_FAIL_ON_WARNINGS=true
    - task generate
```

### 2. Use Strict Mode for New Projects

For new projects, enable strict mode to enforce best practices:

```yaml
validator:
  enabled: true
  strict_mode: true
  fail_on_warnings: true
  custom_rules:
    - require-description
    - require-contact
    - require-license
```

### 3. Gradual Migration for Legacy Projects

For existing projects with many specs, use ignored rules:

```yaml
validator:
  enabled: true
  ignored_rules:
    - NO_SECURITY        # We'll add security later
    - EMPTY_TITLE        # Some old specs have this
```

### 4. Test Validation Locally

Before committing, test validation locally:

```bash
# Enable validation
export VALIDATOR_ENABLED=true

# Run generation
task generate

# Check for validation errors in output
```

### 5. Document Ignored Rules

If you ignore specific rules, document why:

```yaml
validator:
  # Temporarily ignore NO_SECURITY while we migrate to auth
  # TODO: Remove this after Q2 2026 auth migration
  ignored_rules:
    - NO_SECURITY
```

## Troubleshooting

### Validation Taking Too Long

If validation is slow for many specs:

```yaml
# Disable validation (not recommended for production)
validator:
  enabled: false
```

Or use selective validation:

```yaml
# Only validate specs matching pattern
target_services: "critical-.*"
```

### Too Many Warnings

If getting too many warnings:

1. **Gradual approach**: Start with errors only
   ```yaml
   validator:
     fail_on_warnings: false
   ```

2. **Ignore specific warnings**:
   ```yaml
   validator:
     ignored_rules:
       - MISSING_DESCRIPTION
       - EMPTY_TITLE
   ```

3. **Fix specs**: Update specs to address warnings

### Validation Passes but Generation Fails

Validation checks spec format and structure, not all OpenAPI features. Some valid specs may use features not supported by ogen:

- OpenAPI 3.1 specific features (partial support)
- Complex schema compositions
- Certain callback patterns

Check the ogen documentation for supported features: https://ogen.dev/docs/features/

## Examples

### Example 1: Minimal Valid Spec

```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "My API",
    "version": "1.0.0"
  },
  "paths": {}
}
```

**Validation**: ✅ PASSED

### Example 2: Spec with Warnings (Strict Mode)

```yaml
openapi: "3.0.0"
info:
  title: My API
  version: "1.0.0"
paths: {}
```

**Validation** (strict mode): ⚠️ PASSED with warnings
- Warning: NO_SECURITY - No security schemes defined

### Example 3: Invalid Spec

```json
{
  "swagger": "2.0",
  "info": {
    "title": "Old API",
    "version": "1.0"
  }
}
```

**Validation**: ❌ FAILED
- Error: UNSUPPORTED_VERSION - Swagger 2.0 is not supported

### Example 4: Comprehensive Spec

```yaml
openapi: "3.0.0"
info:
  title: Secure API
  version: "1.0.0"
  description: A well-documented API
  contact:
    name: API Support
    email: support@example.com
  license:
    name: MIT
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
security:
  - bearerAuth: []
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success
```

**Validation** (strict mode with all custom rules): ✅ PASSED
- No errors
- No warnings
- Has security: true
- Security schemes: 1 (bearerAuth)

## See Also

- [Configuration Guide](./configuration.md)
- [Troubleshooting Guide](./troubleshooting.md)
- [OpenAPI 3.0 Specification](https://spec.openapis.org/oas/v3.0.3)
- [ogen Generator Features](https://ogen.dev/docs/features/)
