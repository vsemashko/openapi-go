# OpenAPI Go Generator - Troubleshooting Guide

## Table of Contents

1. [Common Issues](#common-issues)
2. [Error Messages](#error-messages)
3. [Performance Issues](#performance-issues)
4. [Cache Issues](#cache-issues)
5. [Generated Code Issues](#generated-code-issues)
6. [Debugging Tips](#debugging-tips)
7. [Getting Help](#getting-help)

## Common Issues

### No OpenAPI Specs Found

**Symptom**:
```
Error: no OpenAPI specs found for target services
```

**Causes**:
1. Incorrect `specs_dir` configuration
2. Target services filter too restrictive
3. Spec files not named correctly
4. Directory structure doesn't match expectations

**Solutions**:

1. **Verify specs directory exists**:
```bash
ls -la ./external/sdk/sdk-packages
```

2. **Check directory structure**:
```
specs/
├── funding-server-sdk/
│   └── openapi.json      # ✓ Correct
├── holidays-server/
│   └── api-spec.json     # ✗ Wrong filename
└── payment/
    └── openapi.json      # May be filtered out
```

3. **Test with relaxed filter**:
```yaml
target_services: ".*"  # Match everything
```

4. **Enable debug logging**:
```yaml
log_level: "debug"
log_format: "text"
```

5. **Check spec file patterns**:
```yaml
spec_file_patterns:
  - "openapi.json"
  - "openapi.yaml"
  - "openapi.yml"
  - "api-spec.json"  # Add custom patterns
```

---

### Failed to Create Output Directory

**Symptom**:
```
Error: failed to create client output directory: permission denied
```

**Causes**:
1. Insufficient permissions
2. Parent directory doesn't exist
3. File exists with same name

**Solutions**:

1. **Check permissions**:
```bash
ls -ld ./generated
chmod 755 ./generated
```

2. **Create directory manually**:
```bash
mkdir -p ./generated/clients
```

3. **Use writable location**:
```yaml
output_dir: "/tmp/openapi-generated"
```

4. **Check for file conflicts**:
```bash
# Remove file if it exists
rm ./generated
mkdir -p ./generated
```

---

### Generator Installation Failed

**Symptom**:
```
Error: failed to ensure ogen is installed: ogen installation verification failed
```

**Causes**:
1. Network connectivity issues
2. Go toolchain not properly configured
3. Insufficient disk space
4. Proxy configuration

**Solutions**:

1. **Verify Go installation**:
```bash
go version  # Should be 1.24.0+
go env GOPATH
go env GOMODCACHE
```

2. **Install ogen manually**:
```bash
go install github.com/ogen-go/ogen/cmd/ogen@v1.14.0
```

3. **Check network connectivity**:
```bash
curl -I https://proxy.golang.org
go get -v github.com/ogen-go/ogen/cmd/ogen@v1.14.0
```

4. **Configure proxy if needed**:
```bash
export GOPROXY=https://proxy.golang.org,direct
export GOPRIVATE=gitlab.stashaway.com
```

5. **Clear Go cache and retry**:
```bash
go clean -modcache
go mod download
go run main.go
```

---

### Generation Failed for Spec

**Symptom**:
```
Error: generation failed for funding: failed to generate client
```

**Causes**:
1. Invalid OpenAPI spec
2. Unsupported OpenAPI features
3. Malformed JSON/YAML
4. Incompatible OpenAPI version

**Solutions**:

1. **Validate the spec**:
```bash
# Install validator
npm install -g @apidevtools/swagger-cli

# Validate spec
swagger-cli validate path/to/openapi.json
```

2. **Check OpenAPI version**:
```yaml
# Spec must be OpenAPI 3.0.x
openapi: 3.0.3  # ✓ Supported
openapi: 3.1.0  # ✗ Not supported yet
openapi: 2.0    # ✗ Swagger 2.0 not supported
```

3. **Test with simplified spec**:
```json
{
  "openapi": "3.0.3",
  "info": {
    "title": "Test API",
    "version": "1.0.0"
  },
  "paths": {
    "/health": {
      "get": {
        "responses": {
          "200": {
            "description": "OK"
          }
        }
      }
    }
  }
}
```

4. **Check for common spec issues**:
- Missing `openapi` version field
- Invalid JSON/YAML syntax
- Circular references
- Missing required fields
- Unsupported extensions

5. **Enable continue-on-error to see all failures**:
```yaml
continue_on_error: true
log_level: "debug"
```

---

### Context Cancelled

**Symptom**:
```
Error: generation cancelled: context canceled
```

**Causes**:
1. Manual interruption (Ctrl+C)
2. Timeout in CI/CD
3. System resource exhaustion

**Solutions**:

1. **Increase CI/CD timeout**:
```yaml
# .gitlab-ci.yml
generate:
  timeout: 30m  # Increase from default
```

2. **Reduce parallel workers**:
```yaml
worker_count: 2  # Lower for resource-constrained environments
```

3. **Check system resources**:
```bash
# CPU usage
top

# Memory usage
free -h

# Disk space
df -h
```

4. **Process specs in batches**:
```yaml
# First run
target_services: "(service1|service2|service3)"

# Second run
target_services: "(service4|service5|service6)"
```

---

## Error Messages

### "cache check failed"

**Full Message**:
```
Warning: Cache check failed for funding: failed to read spec file
```

**Meaning**: The cache system couldn't validate if the cached version is still valid.

**Impact**: Non-critical. Generation will proceed without cache.

**Solutions**:
1. Verify spec file still exists and is readable
2. Clear cache: `rm -rf .openapi-cache`
3. Disable cache if persistent: `enable_cache: false`

---

### "failed to update cache"

**Full Message**:
```
Warning: Failed to update cache for funding: permission denied
```

**Meaning**: Cache entry couldn't be written after successful generation.

**Impact**: Non-critical. Client is generated but not cached.

**Solutions**:
1. Check cache directory permissions: `chmod 755 .openapi-cache`
2. Use writable cache location: `cache_dir: "/tmp/openapi-cache"`
3. Disable cache if needed: `enable_cache: false`

---

### "failed to apply post-processors"

**Full Message**:
```
Error: failed to apply post-processors for fundingsdk: internal client generation failed
```

**Meaning**: Generated code couldn't be post-processed.

**Impact**: Critical. Client is generated but missing features (internal client, formatting).

**Solutions**:
1. Check generated code is valid Go
2. Verify gofmt is available: `which gofmt`
3. Check for disk space: `df -h`
4. Review post-processor logs for details

---

### "parallel processing failed"

**Full Message**:
```
Error: parallel processing failed: worker pool initialization failed
```

**Meaning**: Worker pool couldn't be created or failed during execution.

**Impact**: Critical. No clients generated.

**Solutions**:
1. Reduce worker count: `worker_count: 1`
2. Check system resources
3. Enable debug logging: `log_level: "debug"`
4. Try sequential processing: `worker_count: 1`

---

### "failed to export metrics"

**Full Message**:
```
Warning: Failed to export metrics: permission denied
```

**Meaning**: Metrics file couldn't be written.

**Impact**: Non-critical. Generation succeeds but metrics not saved.

**Solutions**:
1. Check output directory permissions
2. Create metrics directory: `mkdir -p ./generated`
3. Use temporary location: `OUTPUT_DIR=/tmp/generated`

---

## Performance Issues

### Slow Generation

**Symptoms**:
- Generation takes several minutes
- High CPU usage
- Slow progress

**Diagnosis**:
```bash
# Check metrics
cat generated/.openapi-metrics.json | jq '.spec_metrics | sort_by(.duration_ms) | reverse | .[0:5]'

# Check cache hit rate
cat generated/.openapi-metrics.json | jq '(.cached_specs / .total_specs) * 100'
```

**Solutions**:

1. **Enable caching**:
```yaml
enable_cache: true
cache_dir: ".openapi-cache"
```

2. **Increase workers** (if CPU available):
```yaml
worker_count: 8
```

3. **Optimize specs**:
- Remove unused definitions
- Split large specs
- Use `$ref` for reusability

4. **Profile generation**:
```bash
# Enable profiling
go run -cpuprofile=cpu.prof main.go

# Analyze
go tool pprof cpu.prof
```

---

### High Memory Usage

**Symptoms**:
- System becomes unresponsive
- OOM (Out of Memory) errors
- Swap usage increases

**Diagnosis**:
```bash
# Monitor memory
watch -n 1 'free -h'

# Check per-process memory
ps aux | grep openapi-go
```

**Solutions**:

1. **Reduce parallel workers**:
```yaml
worker_count: 2  # Lower for memory-constrained systems
```

2. **Process in batches**:
```bash
# Batch 1
export TARGET_SERVICES="(service1|service2)"
go run main.go

# Batch 2
export TARGET_SERVICES="(service3|service4)"
go run main.go
```

3. **Increase system swap**:
```bash
# Add swap space (Linux)
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

4. **Use 64-bit Go**:
```bash
go version
# Should show amd64 or arm64, not 386
```

---

### Cache Not Working

**Symptoms**:
- Cache hit rate always 0%
- Specs regenerated every time
- No speed improvement

**Diagnosis**:
```bash
# Check cache directory
ls -la .openapi-cache

# Check cache is enabled
grep enable_cache resources/application.yml

# View cache entries
cat .openapi-cache/*.json | jq .
```

**Solutions**:

1. **Verify cache is enabled**:
```yaml
enable_cache: true
```

2. **Check cache directory permissions**:
```bash
chmod 755 .openapi-cache
```

3. **Verify specs aren't changing**:
```bash
# Check spec modification time
stat external/sdk/*/openapi.json

# Compute hash
sha256sum external/sdk/*/openapi.json
```

4. **Clear and rebuild cache**:
```bash
rm -rf .openapi-cache
go run main.go  # First run: builds cache
go run main.go  # Second run: uses cache
```

5. **Check for git changes**:
```bash
# Git might be modifying files (line endings, etc.)
git config core.autocrlf false
```

---

## Cache Issues

### Cache Entries Invalid

**Symptom**:
```
Pruned 5 invalid cache entries
```

**Meaning**: Cache contained entries that couldn't be validated or were corrupted.

**Impact**: Non-critical. Invalid entries are automatically removed.

**Actions**:
- No action needed (automatic cleanup)
- If frequent, investigate disk integrity: `fsck`

---

### Cache Directory Permission Denied

**Symptom**:
```
Warning: Failed to initialize cache, proceeding without caching: permission denied
```

**Solutions**:
```bash
# Fix permissions
chmod 755 .openapi-cache

# Or use different location
export CACHE_DIR=/tmp/openapi-cache
```

---

### Cache Taking Too Much Space

**Symptom**:
```bash
$ du -sh .openapi-cache
2.5G    .openapi-cache
```

**Solutions**:

1. **Clear cache**:
```bash
rm -rf .openapi-cache
```

2. **Prune old entries** (manual):
```bash
# Remove entries older than 30 days
find .openapi-cache -type f -mtime +30 -delete
```

3. **Disable cache**:
```yaml
enable_cache: false
```

---

## Generated Code Issues

### Compilation Errors

**Symptom**:
```bash
$ go build ./generated/clients/fundingsdk
# errors about undefined types, missing methods
```

**Solutions**:

1. **Regenerate the client**:
```bash
rm -rf ./generated/clients/fundingsdk
go run main.go
```

2. **Check Go module**:
```bash
cd generated/clients/fundingsdk
go mod tidy
go build
```

3. **Verify spec is valid**:
```bash
swagger-cli validate path/to/openapi.json
```

4. **Check ogen version**:
```bash
go list -m github.com/ogen-go/ogen
# Should be v1.14.0
```

---

### Missing Internal Client

**Symptom**:
```go
// NewInternalClient not found
client, err := fundingsdk.NewInternalClient(url)
// undefined: NewInternalClient
```

**Causes**:
- Post-processor failed
- Spec has no security definitions

**Solutions**:

1. **Check logs for post-processor errors**:
```bash
go run main.go 2>&1 | grep "post-processor"
```

2. **Verify file exists**:
```bash
ls -la generated/clients/fundingsdk/oas_internal_client_gen.go
```

3. **Regenerate with debug logging**:
```yaml
log_level: "debug"
log_format: "text"
```

4. **Check spec has security**:
```yaml
# Spec needs security for internal client generation
security:
  - bearerAuth: []
```

---

### Type Mismatches

**Symptom**:
```go
cannot use response (type *fundingsdk.GetUserOK) as type User
```

**Cause**:
Generated types don't match your expectations.

**Solutions**:

1. **Use correct types from generated code**:
```go
switch r := response.(type) {
case *fundingsdk.GetUserOK:
    user := r.User  // Use generated types
}
```

2. **Check generated documentation**:
```bash
godoc -http=:6060
# Visit http://localhost:6060/pkg/yourmodule/generated/clients/fundingsdk
```

3. **Review spec definitions**:
```json
{
  "components": {
    "schemas": {
      "User": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "name": {"type": "string"}
        }
      }
    }
  }
}
```

---

## Debugging Tips

### Enable Debug Logging

```yaml
log_level: "debug"
log_format: "text"
```

Or:
```bash
export LOG_LEVEL=debug
export LOG_FORMAT=text
go run main.go 2>&1 | tee debug.log
```

---

### Verbose Ogen Output

```bash
# Run ogen directly to see detailed output
cd /tmp
ogen --target ./output \
     --package testpkg \
     --clean \
     --config /path/to/ogen.yaml \
     /path/to/openapi.json
```

---

### Test with Minimal Spec

Create a minimal test spec:

```json
{
  "openapi": "3.0.3",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "paths": {
    "/health": {
      "get": {
        "operationId": "getHealth",
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": {"type": "string"}
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
```

Test generation:
```bash
mkdir -p test/testservice
cp test-spec.json test/testservice/openapi.json

export SPECS_DIR=./test
export OUTPUT_DIR=./test-output
export TARGET_SERVICES="testservice"
go run main.go
```

---

### Check System Resources

```bash
# CPU
top
htop

# Memory
free -h
vmstat 1

# Disk
df -h
du -sh ./*

# I/O
iostat 1

# Network
netstat -an | grep ESTABLISHED
```

---

### Trace Execution

```bash
# Go execution trace
go run -trace=trace.out main.go
go tool trace trace.out

# System calls trace (Linux)
strace -o trace.txt go run main.go

# Time breakdown
time go run main.go
```

---

### Inspect Metrics

```bash
# View all metrics
cat generated/.openapi-metrics.json | jq .

# Success rate
cat generated/.openapi-metrics.json | jq '(.successful_specs / .total_specs) * 100'

# Average duration
cat generated/.openapi-metrics.json | jq '.average_duration_ms / 1000'

# Slowest specs
cat generated/.openapi-metrics.json | jq '.spec_metrics | sort_by(.duration_ms) | reverse | .[0:5]'

# Failed specs
cat generated/.openapi-metrics.json | jq '.spec_metrics[] | select(.success == false)'

# Cache effectiveness
cat generated/.openapi-metrics.json | jq '{
  total: .total_specs,
  cached: .cached_specs,
  rate: ((.cached_specs / .total_specs) * 100)
}'
```

---

### Compare Generations

```bash
# Save old generated code
cp -r generated generated.old

# Regenerate
go run main.go

# Compare
diff -r generated.old generated
```

---

### Test Individual Components

**Test spec finding**:
```go
package main

import (
    "fmt"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/processor"
)

func main() {
    cfg, _ := config.LoadConfig()
    specs, err := processor.FindOpenAPISpecs(cfg.SpecsDir, cfg.TargetServices, cfg.SpecFilePatterns)
    fmt.Printf("Found %d specs: %v\n", len(specs), specs)
    fmt.Printf("Error: %v\n", err)
}
```

**Test cache**:
```go
package main

import (
    "fmt"
    "gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/cache"
)

func main() {
    c, _ := cache.NewCache(cache.Config{CacheDir: ".openapi-cache"})
    valid, err := c.IsValid("path/to/spec.json", "v1.14.0")
    fmt.Printf("Valid: %v, Error: %v\n", valid, err)
}
```

---

## Getting Help

### Before Asking for Help

1. **Check this guide** for your specific issue
2. **Review logs** with debug level enabled
3. **Validate your OpenAPI spec** using standard tools
4. **Test with minimal configuration** to isolate the problem
5. **Check metrics** for performance insights

### Information to Include

When reporting issues, include:

1. **Version information**:
```bash
go version
git rev-parse HEAD
go list -m github.com/ogen-go/ogen
```

2. **Configuration**:
```bash
cat resources/application.yml
```

3. **Error logs**:
```bash
export LOG_LEVEL=debug
go run main.go 2>&1 | tee error.log
```

4. **System information**:
```bash
uname -a
go env
df -h
free -h
```

5. **Minimal reproduction**:
- Simplified spec
- Exact steps to reproduce
- Expected vs actual behavior

### Where to Get Help

1. **Documentation**:
   - [Usage Guide](./usage-guide.md)
   - [Configuration Guide](./configuration.md)
   - [Architecture Documentation](./architecture.md)

2. **Issue Tracker**:
   - Create detailed bug reports
   - Include reproduction steps
   - Attach relevant logs

3. **Internal Teams**:
   - Reach out to platform team
   - Check internal wiki
   - Ask in team channels

### Common Questions

**Q: Why is generation so slow?**
A: Enable caching, increase workers (if resources allow), optimize specs.

**Q: Can I use OpenAPI 3.1?**
A: Not yet. Currently only OpenAPI 3.0.x is supported. See [Roadmap](./future-roadmap.md).

**Q: How do I debug generated code issues?**
A: Regenerate with debug logging, validate the spec, check ogen version.

**Q: Can I customize the generated code?**
A: Not directly. Use wrapper types or contribute to the post-processor chain.

**Q: Why are some specs skipped?**
A: Check the `target_services` filter and ensure spec filenames match `spec_file_patterns`.

**Q: How do I update to a new ogen version?**
A: Update `go.mod`, clear cache, regenerate all clients.

---

## Next Steps

- Review [Usage Guide](./usage-guide.md) for best practices
- Check [Configuration Guide](./configuration.md) for optimization options
- Read [Architecture Documentation](./architecture.md) to understand internals
- See [Future Roadmap](./future-roadmap.md) for upcoming features
