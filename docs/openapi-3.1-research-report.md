# OpenAPI 3.1 Support - Research Report

**Date**: 2025-11-22
**Version**: v1.0
**Status**: Research Complete
**Priority**: P0 (Highest)

## Executive Summary

This report documents research into adding OpenAPI 3.1 support to the openapi-go generator. After comprehensive testing and analysis, **we cannot proceed with OpenAPI 3.1 support at this time** due to a critical dependency: the ogen code generator does not support OpenAPI 3.1 specifications.

### Key Finding

❌ **Blocker Identified**: ogen (our code generator) does NOT support OpenAPI 3.1
- Tested: v1.14.0 (current) and v1.16.0 (latest as of Oct 2024)
- Both versions fail to parse OpenAPI 3.1 specs
- No roadmap or tracking issue found for 3.1 support in ogen repository

---

## Research Findings

### 1. Ogen Library Status

#### Current Usage
- **Our Version**: v1.14.0 (June 2024)
- **Latest Version**: v1.16.0 (October 2024)
- **Location**: `internal/generator/ogen.go:20`

#### OpenAPI 3.1 Support Status

**Official Documentation**: No mention of OpenAPI 3.1 support

**Testing Results**:
```bash
# Test with v1.14.0 (our version)
$ ogen generate test-openapi-3.1.yaml
ERROR: cannot unmarshal !!int `0` into bool (exclusiveMinimum)
ERROR: cannot unmarshal !!seq into string (type array)

# Test with v1.16.0 (latest)
$ ogen generate test-openapi-3.1.yaml
ERROR: cannot unmarshal !!int `0` into bool (exclusiveMinimum)
ERROR: cannot unmarshal !!seq into string (type array)

# Test with v1.14.0 using OpenAPI 3.0
$ ogen generate test-openapi-3.0.yaml
SUCCESS: Generated 12 Go files
```

**Conclusion**: Neither version supports OpenAPI 3.1 syntax.

#### GitHub Analysis

**Issues Found**:
- [Issue #1309](https://github.com/ogen-go/ogen/issues/1309) - User reported errors with OpenAPI 3.1 spec (Sep 2024)
- [Issue #1320](https://github.com/ogen-go/ogen/issues/1320) - CORS issue with 3.1 spec (Oct 2024)

**Missing**:
- No feature request or tracking issue for OpenAPI 3.1 support
- No mention in roadmap or release notes (v1.10.0 - v1.16.0)
- No discussions about 3.1 implementation

---

### 2. OpenAPI 3.0 vs 3.1 Differences

#### Breaking Changes in 3.1

| Feature | OpenAPI 3.0 | OpenAPI 3.1 | Impact |
|---------|-------------|-------------|--------|
| **Nullable** | `nullable: true` | `type: ["string", "null"]` | **HIGH** - Schema parsing |
| **exclusiveMinimum** | `minimum: 7` + `exclusiveMinimum: true` | `exclusiveMinimum: 7` | **HIGH** - Schema validation |
| **exclusiveMaximum** | `maximum: 100` + `exclusiveMaximum: true` | `exclusiveMaximum: 100` | **HIGH** - Schema validation |
| **example** | `example: "value"` | `examples: ["value"]` | **MEDIUM** - Documentation |
| **File uploads** | `format: base64` | `contentEncoding: base64` | **MEDIUM** - Binary data |
| **$ref siblings** | Not allowed | Allowed | **LOW** - Schema composition |

#### New Features in 3.1

1. **JSON Schema 2020-12 Alignment**: Full compatibility with modern JSON Schema
2. **Webhooks**: First-class webhook support
3. **Enhanced discriminators**: Better polymorphism handling
4. **$schema support**: Explicit schema version declaration
5. **Advanced validation**: New keywords like `prefixItems`, `unevaluatedProperties`

#### Error Examples from Testing

```yaml
# This OpenAPI 3.1 spec FAILS with ogen
openapi: 3.1.0
paths:
  /users/{id}:
    parameters:
      - schema:
          type: integer
          exclusiveMinimum: 0  # ERROR: "cannot unmarshal !!int `0` into bool"
components:
  schemas:
    User:
      properties:
        email:
          type: ["string", "null"]  # ERROR: "cannot unmarshal !!seq into string"
```

---

### 3. Alternative Go Generators Analysis

| Generator | OpenAPI 3.1 Support | Status | Notes |
|-----------|---------------------|--------|-------|
| **ogen** | ❌ No | Active | Fast, type-safe, no 3.1 support |
| **oapi-codegen** | ❌ No | Active | Popular, awaiting upstream support |
| **go-swagger** | ⚠️ Partial | Maintenance | Mature, limited 3.1 features |
| **libopenapi** | ✅ Yes | Active | Parser only, not a generator |

**Key Finding**: Most Go OpenAPI generators do NOT support 3.1 yet (as of Nov 2025).

---

### 4. Ecosystem Analysis

#### Industry Adoption (as of 2025)

- **OpenAPI 3.1 Release**: February 2021 (nearly 4 years old)
- **Go Ecosystem**: Slow adoption of 3.1 support
- **Tooling Gap**: Most Go generators still targeting 3.0.x
- **Trend**: Increasing interest in 3.1 over past 2 years

#### Why is 3.1 Adoption Slow?

1. **Breaking Changes**: The `nullable`, `exclusiveMin/Max` changes require parser rewrites
2. **JSON Schema Complexity**: 2020-12 adds significant complexity
3. **Backward Compatibility**: Many existing specs are still 3.0
4. **Library Dependencies**: Code generators depend on parsing libraries

---

## Impact Analysis

### What This Means for openapi-go

#### Current State
- ✅ Full OpenAPI 3.0.x support (tested and working)
- ✅ YAML and JSON parsing
- ✅ Comprehensive validation
- ✅ Incremental generation
- ❌ Cannot parse OpenAPI 3.1 specs

#### User Impact

**Who is Blocked?**
- Users with modern APIs using OpenAPI 3.1 features
- Teams wanting to use JSON Schema 2020-12 features
- Projects requiring webhook support

**Who is NOT Blocked?**
- Users with OpenAPI 3.0.x specs (vast majority)
- Teams using standard REST APIs
- Existing openapi-go users

**Migration Workaround**: Users can downgrade 3.1 specs to 3.0 manually:
```yaml
# Change from 3.1
type: ["string", "null"]
exclusiveMinimum: 5

# To 3.0
type: string
nullable: true
minimum: 5
exclusiveMinimum: true
```

---

## Options & Recommendations

### Option 1: Wait for ogen Support ⏳ **RECOMMENDED**

**Approach**: Monitor ogen repository and upgrade when 3.1 support is added

**Pros**:
- Zero development effort
- Maintains compatibility with existing architecture
- No risk of breaking changes
- ogen is actively maintained

**Cons**:
- Unknown timeline (could be months or years)
- No tracking issue to follow
- Uncertain if 3.1 is even on their roadmap

**Effort**: 1-2 hours/month monitoring
**Risk**: Low
**Timeline**: Unknown

**Recommendation**: Continue monitoring ogen for 3.1 support while implementing other roadmap items.

---

### Option 2: Contribute to ogen ⚙️

**Approach**: Contribute OpenAPI 3.1 support to ogen library

**Pros**:
- Benefits entire Go ecosystem
- Full control over implementation
- Upstream solution

**Cons**:
- Very high effort (estimated 4-6 weeks full-time)
- Requires deep knowledge of ogen internals
- Needs ogen maintainer buy-in
- Ongoing maintenance burden

**Effort**: 160-240 hours
**Risk**: High (maintainer rejection, scope creep)
**Timeline**: 2-3 months

**Recommendation**: Only if OpenAPI 3.1 becomes critical business requirement.

---

### Option 3: Implement Spec Downgrader 🔄

**Approach**: Build a 3.1 → 3.0 converter in openapi-go

**Pros**:
- Unblocks 3.1 users immediately
- Full control over conversion logic
- Transparent to end users
- Medium effort

**Cons**:
- Lossy conversion (some 3.1 features can't be represented in 3.0)
- Additional complexity in codebase
- Maintenance burden
- Webhooks cannot be converted

**Effort**: 40-60 hours (2-3 weeks)
**Risk**: Medium (conversion edge cases)
**Timeline**: 2-3 weeks

**Implementation**:
```go
// internal/spec/downgrader.go
type SpecDowngrader struct{}

func (d *SpecDowngrader) DowngradeToV30(spec *OpenAPISpec) (*OpenAPISpec, error) {
    // Convert type arrays to nullable
    // Convert exclusive min/max syntax
    // Convert examples to example
    // Warn about webhooks (not supported)
}
```

**Recommendation**: Good medium-term solution if 3.1 demand increases.

---

### Option 4: Add Multi-Generator Support 🔌

**Approach**: Support multiple code generators, add one with 3.1 support

**Pros**:
- Flexibility for users
- Can support both 3.0 and 3.1
- Aligns with roadmap (P2 priority)

**Cons**:
- No Go generator currently supports 3.1 fully
- High implementation effort
- Increased maintenance complexity

**Effort**: 80-120 hours (4-6 weeks)
**Risk**: High (no viable 3.1 generator exists)
**Timeline**: 4-6 weeks (after finding suitable generator)

**Recommendation**: Good long-term strategy, but blocked by same issue (no 3.1 generators available).

---

### Option 5: Custom Code Generator 🏗️

**Approach**: Build our own OpenAPI 3.1 code generator from scratch

**Pros**:
- Full control over features
- Can support 3.1 from day one
- No dependency on external libraries

**Cons**:
- **EXTREMELY high effort** (estimated 3-6 months)
- Massive maintenance burden
- High complexity
- Risky and unrealistic

**Effort**: 480-960 hours (3-6 months full-time)
**Risk**: Very High
**Timeline**: 6-12 months

**Recommendation**: ❌ Not recommended - unrealistic scope.

---

## Recommended Action Plan

### Phase 1: Continue with Current Roadmap ✅ **IMMEDIATE**

**Status**: Proceed with other high-priority items

**Rationale**:
- OpenAPI 3.1 is blocked by external dependency
- Other roadmap items provide immediate value
- OpenAPI 3.0 serves vast majority of use cases

**Next Steps**:
1. ✅ Enhanced Validation (COMPLETED - v2.1.0)
2. ✅ Incremental Generation (COMPLETED - v2.2.0)
3. ➡️ **Enhanced Error Handling** (Option B from roadmap)
4. ➡️ **Enhanced Metrics** (Option C from roadmap)

---

### Phase 2: Monitor & Document 📊 **ONGOING**

**Actions**:
1. **Monthly Check**: Monitor ogen releases and issues
2. **Update CHANGELOG**: Keep "Planned" section current
3. **User Communication**: Document 3.1 status in README
4. **Issue Tracking**: Create ogen issue if none exists

**Monitoring Checklist**:
- [ ] Check ogen releases monthly: https://github.com/ogen-go/ogen/releases
- [ ] Search for "OpenAPI 3.1" in ogen issues
- [ ] Review alternative generators (oapi-codegen, go-swagger)
- [ ] Track JSON Schema ecosystem progress

---

### Phase 3: Prepare for Future Implementation 🎯 **WHEN READY**

**Trigger**: ogen adds OpenAPI 3.1 support

**Preparation**:
1. **Upgrade Path**: Document ogen upgrade process
2. **Testing**: Create comprehensive 3.1 test suite
3. **Validation**: Update validator for 3.1 features
4. **Migration Guide**: Help users transition to 3.1

**Estimated Effort** (once ogen supports 3.1):
- Update ogen version: 1 hour
- Test suite updates: 8-16 hours
- Validator updates: 8-16 hours
- Documentation: 8 hours
- **Total**: 25-41 hours (1-2 weeks)

---

## Technical Details

### Required Changes to openapi-go (when ogen supports 3.1)

#### 1. Spec Parser Updates
```go
// internal/spec/parser.go

// Add 3.1-specific fields
type OpenAPISpec struct {
    OpenAPI string  // "3.1.0" instead of "3.0.3"
    // ... existing fields

    // NEW: 3.1-specific fields
    Webhooks map[string]PathItem `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`
    JSONSchemaDialect string `json:"jsonSchemaDialect,omitempty" yaml:"jsonSchemaDialect,omitempty"`
}

// Support type arrays for nullable
type Schema struct {
    Type interface{} `json:"type,omitempty" yaml:"type,omitempty"` // string or []string
    // ... other fields
}
```

#### 2. Validator Updates
```go
// internal/validator/validator.go

func (v *SpecValidator) validateVersion(spec *OpenAPISpec) []ValidationError {
    switch {
    case strings.HasPrefix(spec.OpenAPI, "3.0"):
        return v.validate30(spec)
    case strings.HasPrefix(spec.OpenAPI, "3.1"):
        return v.validate31(spec)  // NEW
    default:
        return []ValidationError{{Code: "UNSUPPORTED_VERSION"}}
    }
}
```

#### 3. Fingerprinting Updates
```go
// internal/spec/fingerprint.go

// Include webhooks in fingerprint
func CreateSpecFingerprint(spec *OpenAPISpec) (*SpecFingerprint, error) {
    operations := spec.GetOperations()
    webhooks := spec.GetWebhooks()  // NEW

    fp := &SpecFingerprint{
        Operations: make(map[string]OperationFingerprint),
        Webhooks:   make(map[string]WebhookFingerprint),  // NEW
    }
    // ...
}
```

#### 4. Generator Version Updates
```go
// internal/generator/ogen.go

const (
    // Update to version that supports 3.1
    OgenVersion = "v2.0.0"  // Example future version
)
```

#### 5. Test Suite Additions
- Add OpenAPI 3.1 test fixtures
- Test type arrays (`type: ["string", "null"]`)
- Test new exclusive min/max syntax
- Test webhooks support
- Test JSON Schema 2020-12 keywords

---

## Metrics & Success Criteria

### How to Measure Success (when implemented)

1. **Spec Compatibility**:
   - ✅ Parse 100% of valid OpenAPI 3.1 specs
   - ✅ Support all 3.1-specific keywords
   - ✅ Maintain backward compatibility with 3.0

2. **Code Generation**:
   - ✅ Generate working clients from 3.1 specs
   - ✅ Type safety for nullable fields (type arrays)
   - ✅ Proper handling of exclusive min/max
   - ✅ Webhook support (if ogen supports it)

3. **Performance**:
   - ✅ No performance degradation vs 3.0
   - ✅ Incremental generation works with 3.1
   - ✅ Cache invalidation accounts for 3.1 features

4. **User Experience**:
   - ✅ Clear error messages for 3.1-specific issues
   - ✅ Migration guide from 3.0 to 3.1
   - ✅ Examples and documentation

---

## Conclusion

### Summary

OpenAPI 3.1 support **cannot be implemented at this time** due to the ogen code generator lacking 3.1 support. After testing both our current version (v1.14.0) and the latest version (v1.16.0), neither can parse OpenAPI 3.1 specifications.

### Recommendation

**Proceed with Option 1**: Continue with current roadmap items while monitoring ogen for 3.1 support.

**Next Immediate Actions**:
1. ✅ Document findings (this report)
2. ✅ Update roadmap with current status
3. ➡️ Proceed with Enhanced Error Handling (Option B)
4. ➡️ Set up monthly monitoring of ogen releases
5. ➡️ Consider creating issue on ogen repository to request 3.1 support

### Timeline Estimate

- **If ogen adds 3.1 support in Q1 2026**: Implementation in Q1-Q2 2026 (4-6 weeks)
- **If ogen adds 3.1 support in H2 2026**: Implementation in H2 2026
- **If ogen never adds 3.1 support**: Reconsider multi-generator approach or spec downgrader in 2026

### Risk Assessment

**Risk Level**: **LOW**
**Rationale**: Most users still use OpenAPI 3.0.x. The lack of 3.1 support is not currently blocking any known use cases.

---

## References

### Sources Consulted

1. **ogen Repository**:
   - [GitHub - ogen-go/ogen](https://github.com/ogen-go/ogen)
   - [Issue #1309](https://github.com/ogen-go/ogen/issues/1309) - OpenAPI 3.1 parsing error
   - [Issue #1320](https://github.com/ogen-go/ogen/issues/1320) - CORS with 3.1 spec

2. **OpenAPI Specification**:
   - [Upgrading from OpenAPI 3.0 to 3.1](https://learn.openapis.org/upgrading/v3.0-to-v3.1.html)
   - [Migrating from OpenAPI 3.0 to 3.1.0](https://www.openapis.org/blog/2021/02/16/migrating-from-openapi-3-0-to-3-1-0)
   - [OpenAPI 3.1.0 Release Notes](https://www.openapis.org/blog/2021/02/18/openapi-specification-3-1-released)

3. **Comparison Articles**:
   - [OpenAPI 3.1.0 Compared to 3.0.3 | Beeceptor](https://beeceptor.com/docs/concepts/openapi-what-is-new-3.1.0/)
   - [Why Upgrade to OpenAPI 3.1 | Document360](https://document360.com/blog/openapi-3-0-vs-openapi-3-1/)
   - [Difference Between OpenAPI 2.0, 3.0, and 3.1 | Stoplight](https://blog.stoplight.io/difference-between-open-v2-v3-v31)

### Related Documentation

- [Future Roadmap](./future-roadmap.md) - Phase 1: OpenAPI 3.1 Support (P0)
- [Architecture Documentation](./architecture.md) - Generator Interface
- [Validator Guide](./validator-guide.md) - Spec validation system

---

**Report Authors**: Claude Code
**Review Date**: 2025-11-22
**Next Review**: 2026-01-01 (monthly ogen check)
**Status**: Final - Ready for Decision
