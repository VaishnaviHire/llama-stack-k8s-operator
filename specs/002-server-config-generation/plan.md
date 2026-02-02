# Implementation Plan: Operator-Generated Server Configuration

**Branch**: `002-server-config-generation` | **Date**: 2026-02-02 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/002-server-config-generation/spec.md`

---

## Summary

Enable the llama-stack Kubernetes operator to generate server configuration (`config.yaml`) from a high-level, abstracted specification in the `LlamaStackDistribution` CR. The operator will support a layered configuration model with `global.core` (simplified), `global.raw` (advanced), `resources.core` (simplified), and `resources.raw` (advanced) sections, deep-merging user configuration over distribution defaults while securely handling secrets via environment variable injection.

---

## Technical Context

**Language/Version**: Go 1.23+  
**Primary Dependencies**: controller-runtime, kubebuilder, sigs.k8s.io/yaml, gopkg.in/yaml.v3  
**Storage**: Kubernetes ConfigMaps (generated), Secrets (referenced)  
**Testing**: Ginkgo/Gomega (controller tests), testify (unit tests), envtest (integration)  
**Target Platform**: Kubernetes 1.28+  
**Project Type**: Kubernetes Operator (kubebuilder-based)  
**Performance Goals**: Configuration generation <100ms, reconciliation cycle <5s  
**Constraints**: Namespace-scoped resources only, no cluster-admin privileges required  
**Scale/Scope**: Support 100+ concurrent LlamaStackDistribution CRs per cluster

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Rule | Section | Status | Notes |
|------|---------|--------|-------|
| Reconciliation is Idempotent | §1.2 | ✅ PASS | Config generation produces same output for same input |
| Separate Reconciliation from Status | §1.2 | ✅ PASS | Will follow existing `reconcileResources`/`updateStatus` pattern |
| Wrap Errors with Context | §4.1 | ✅ PASS | All errors will use `fmt.Errorf("failed to X: %w", err)` |
| Use Kubebuilder Validation | §2.1 | ✅ PASS | Will use validation tags for provider enum, required fields |
| Status Has Phase + Conditions | §3 | ✅ PASS | Will add ConfigGenerated condition type |
| Logger in Context | §5.1 | ✅ PASS | Will use existing context logger pattern |
| Table-Driven Tests | §6.1 | ✅ PASS | All test cases will be table-driven |
| Builder Pattern for Tests | §6.4 | ✅ PASS | Will extend existing DistributionBuilder |
| Owner References for Cleanup | §1.3 | ✅ PASS | Generated ConfigMaps will have owner references |
| Namespace-Scoped Resources | §1.1 | ✅ PASS | No cluster-scoped resources |
| Use Pointers for Optional Structs | §2.2 | ✅ PASS | Config will be `*ServerConfigSpec` |
| Define Constants for Defaults | §2.3 | ✅ PASS | Will define defaults for storage type, etc. |

**All gates pass. Proceeding with implementation.**

---

## Project Structure

### Documentation (this feature)

```text
specs/002-server-config-generation/
├── plan.md              # This file
├── research.md          # Phase 0 research findings
├── data-model.md        # CRD type definitions
├── quickstart.md        # User getting started guide
├── checklists/          # Quality checklists
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Implementation tasks (created by /speckit.tasks)
```

### Source Code (repository root)

```text
api/v1alpha1/
├── llamastackdistribution_types.go   # Extended with ServerConfigSpec
├── config_types.go                   # NEW: Config type definitions
├── config_validation.go              # NEW: Validation logic
└── zz_generated.deepcopy.go          # Auto-generated

controllers/
├── llamastackdistribution_controller.go  # Extended reconciliation
├── config_generator.go                   # NEW: Config generation logic
├── config_generator_test.go              # NEW: Config generation tests
├── secret_resolver.go                    # NEW: Secret reference resolution
├── secret_resolver_test.go               # NEW: Secret resolver tests
├── deep_merge.go                         # NEW: Deep merge implementation
├── deep_merge_test.go                    # NEW: Deep merge tests
├── resource_helper.go                    # Extended for config mount
├── status.go                             # Extended with new conditions
└── testing_support_test.go               # Extended test builders

pkg/
├── config/
│   ├── expander.go                       # NEW: Core to raw expansion
│   ├── expander_test.go                  # NEW: Expander tests
│   ├── merger.go                         # NEW: Config merge logic
│   ├── merger_test.go                    # NEW: Merger tests
│   ├── schema.go                         # NEW: Config schema definitions
│   └── version.go                        # NEW: Version detection/validation
└── deploy/
    └── deploy.go                         # Extended for config ConfigMap
```

**Structure Decision**: Single project structure following existing operator patterns. New functionality added to existing packages (`api/v1alpha1`, `controllers`) with a new `pkg/config` package for configuration-specific logic.

---

## Architecture Overview

### Component Diagram

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                    LlamaStackDistribution CR                            │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ spec.server.config                                               │    │
│  │  ├── global.core (InferenceConfig, StorageConfig, etc.)         │    │
│  │  ├── global.raw  (providers, storage, server - unstructured)    │    │
│  │  ├── resources.core (models[], tools[])                         │    │
│  │  └── resources.raw  (models[], shields[], vector_dbs[], etc.)   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Reconciler                                      │
│  ┌────────────────┐    ┌────────────────┐    ┌────────────────────┐    │
│  │ Validator      │───▶│ ConfigGenerator │───▶│ Secret Resolver    │    │
│  │ (CRD + Runtime)│    │ (Expand + Merge)│    │ (valueFrom → env)  │    │
│  └────────────────┘    └────────────────┘    └────────────────────┘    │
│           │                     │                      │                │
│           ▼                     ▼                      ▼                │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    ConfigMap Generator                           │    │
│  │  - Read base config.yaml from distribution                       │    │
│  │  - Expand core → raw                                            │    │
│  │  - Deep merge raw over expanded                                  │    │
│  │  - Replace secrets with ${env.VAR} placeholders                  │    │
│  │  - Generate <name>-config-gen-<N> ConfigMap                      │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Generated Resources                             │
│  ┌──────────────────┐    ┌─────────────────────────────────────────┐   │
│  │ ConfigMap        │    │ Deployment                              │   │
│  │ <name>-config-   │    │  - config-hash annotation               │   │
│  │ gen-<generation> │    │  - env vars from secret refs            │   │
│  │                  │    │  - ConfigMap volume mount               │   │
│  └──────────────────┘    └─────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Configuration Flow

```text
1. User applies CR with spec.server.config
                    │
                    ▼
2. Validator checks:
   - Mutual exclusivity (config vs userConfig)
   - Provider values (enum validation)
   - Required fields (storage.connectionString when postgres)
   - Secret references exist
                    │
                    ▼
3. Config Generator:
   a. Load distribution's embedded config.yaml
   b. Expand global.core → full provider config
   c. Deep merge global.raw over expanded
   d. Expand resources.core → full registrations
   e. Deep merge resources.raw over expanded
                    │
                    ▼
4. Secret Resolver:
   - Find all secretKeyRef in config
   - Generate env var names (LLSD_*)
   - Replace with ${env.VAR} placeholders
   - Build env var list for Deployment
                    │
                    ▼
5. ConfigMap Generator:
   - Serialize final config to YAML
   - Create immutable ConfigMap with owner ref
   - Name: <cr-name>-config-gen-<generation>
                    │
                    ▼
6. Deployment Update:
   - Mount ConfigMap as volume
   - Inject environment variables
   - Add config-hash annotation
   - Kubernetes triggers rollout
```

---

## Key Design Decisions

### 1. Core vs Raw Layering

**Decision**: Two-tier configuration with `core` (simplified) and `raw` (advanced) options.

**Rationale**:
- 80% of users need simple options (core)
- 20% of users need full control (raw)
- Core provides guardrails and validation
- Raw provides escape hatch without breaking changes

**Implementation**:
- `core` fields are strongly-typed Go structs with kubebuilder validation
- `raw` fields use `apiextensionsv1.JSON` for unstructured data
- Expansion logic converts core → equivalent raw structure
- Deep merge applies raw over expanded core

### 2. Secret Handling via Environment Variables

**Decision**: Replace secret references with `${env.LLSD_*}` placeholders, inject as Deployment env vars.

**Rationale**:
- Secrets never appear in ConfigMaps
- Standard Kubernetes pattern (envFrom/valueFrom)
- Works with existing secret rotation mechanisms
- Deterministic naming enables debugging

**Implementation**:
- Walk config tree finding `secretKeyRef` objects
- Generate env var name: `LLSD_<PATH>_<KEY>` (uppercase, underscores)
- Replace in-tree with `${env.LLSD_<PATH>_<KEY>}`
- Build `corev1.EnvVar` list for Deployment

### 3. Immutable ConfigMaps with Generation Suffix

**Decision**: Create immutable ConfigMaps named `<cr-name>-config-gen-<N>`.

**Rationale**:
- Immutability prevents accidental modification
- Generation suffix provides audit trail
- Hash annotation ensures pod restart on change
- Old ConfigMaps cleaned up by owner reference

**Implementation**:
- Increment generation on each config change
- Store generation in CR status
- Set `immutable: true` on ConfigMap
- Clean up old ConfigMaps (keep N-1 for rollback)

### 4. Provider Type Mapping

**Decision**: Simple provider names (vllm, ollama) map to full types (remote::vllm).

**Rationale**:
- Users don't need to know internal type format
- Reduces configuration errors
- Centralizes provider knowledge in operator
- Easy to extend for new providers

**Implementation**:
```go
var ProviderTypeMap = map[string]string{
    "ollama":    "remote::ollama",
    "vllm":      "remote::vllm",
    "openai":    "remote::openai",
    // ... etc
}
```

### 5. Version Support Policy

**Decision**: Support current (n) and previous (n-1) config schema versions.

**Rationale**:
- Allows gradual migration between versions
- Clear deprecation path
- Prevents accumulation of legacy code

**Implementation**:
- Read `version` field from base config.yaml
- Define supported versions as constants
- Log deprecation warning for n-1
- Reject n-2 and older with clear error

---

## Implementation Phases

### Phase 1: CRD Types and Validation (PR 1)

**Scope**:
- Add `ServerConfigSpec` type to CRD
- Add `GlobalConfig`, `ResourceConfig`, and sub-types
- Add kubebuilder validation markers
- Add XValidation for mutual exclusivity
- Regenerate CRD manifests

**Files Changed**:
- `api/v1alpha1/llamastackdistribution_types.go`
- `api/v1alpha1/config_types.go` (new)
- `api/v1alpha1/zz_generated.deepcopy.go`
- `config/crd/bases/*.yaml`

**Tests**:
- Unit tests for type construction
- Validation tests (valid/invalid scenarios)
- CRD schema validation

### Phase 2: Core-to-Raw Expansion (PR 2)

**Scope**:
- Implement expansion logic for `global.core` → provider config
- Implement expansion logic for `resources.core` → resource registrations
- Provider type mapping
- Default value application

**Files Changed**:
- `pkg/config/expander.go` (new)
- `pkg/config/expander_test.go` (new)
- `pkg/config/schema.go` (new)

**Tests**:
- Table-driven tests for each provider type
- Default value tests
- Edge case tests (empty values, partial config)

### Phase 3: Deep Merge Implementation (PR 3)

**Scope**:
- Implement deep merge for maps and slices
- Handle merge strategies (replace vs append)
- Support both typed and unstructured data

**Files Changed**:
- `controllers/deep_merge.go` (new)
- `controllers/deep_merge_test.go` (new)

**Tests**:
- Nested map merge tests
- Slice merge tests
- Typed struct merge tests
- Edge cases (nil values, empty maps)

### Phase 4: Secret Resolution (PR 4)

**Scope**:
- Implement secret reference detection
- Generate deterministic env var names
- Replace references with placeholders
- Build env var list for Deployment

**Files Changed**:
- `controllers/secret_resolver.go` (new)
- `controllers/secret_resolver_test.go` (new)

**Tests**:
- Secret reference detection tests
- Env var naming tests
- Placeholder replacement tests
- Multiple secrets tests

### Phase 5: ConfigMap Generation (PR 5)

**Scope**:
- Read base config.yaml from distribution image
- Combine expansion, merge, and secret resolution
- Generate immutable ConfigMap
- Add owner reference

**Files Changed**:
- `controllers/config_generator.go` (new)
- `controllers/config_generator_test.go` (new)
- `pkg/config/version.go` (new)

**Tests**:
- End-to-end generation tests
- ConfigMap naming tests
- Owner reference tests
- Version compatibility tests

### Phase 6: Controller Integration (PR 6)

**Scope**:
- Integrate config generation into reconciliation loop
- Update Deployment with env vars and mount
- Add hash annotation for rollout
- Update status with ConfigGenerated condition

**Files Changed**:
- `controllers/llamastackdistribution_controller.go`
- `controllers/resource_helper.go`
- `controllers/status.go`

**Tests**:
- Controller reconciliation tests
- Deployment update tests
- Status condition tests
- Rollout trigger tests

### Phase 7: Backward Compatibility (PR 7)

**Scope**:
- Ensure userConfig.configMapName still works
- Add mutual exclusivity validation
- Handle migration scenarios

**Files Changed**:
- `controllers/llamastackdistribution_controller.go`
- `api/v1alpha1/llamastackdistribution_types.go`

**Tests**:
- Legacy config tests
- Mutual exclusivity tests
- Migration scenario tests

### Phase 8: E2E Tests and Documentation (PR 8)

**Scope**:
- Add E2E tests for new configuration
- Update API documentation
- Add sample CRs
- Update README

**Files Changed**:
- `tests/e2e/config_test.go` (new)
- `docs/api-overview.md`
- `config/samples/` (new samples)
- `README.md`

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Base config.yaml extraction from image fails | Medium | High | Fallback to embedded defaults, clear error message |
| Deep merge produces unexpected results | Medium | Medium | Comprehensive test suite, document merge behavior |
| Secret resolution performance at scale | Low | Medium | Cache resolved secrets within reconciliation |
| Schema version detection fails | Low | High | Default to latest version with warning |
| ConfigMap size limit exceeded | Low | Medium | Validate size before creation, clear error |

---

## Dependencies

### External Dependencies (no changes needed)

- `sigs.k8s.io/yaml` - YAML marshaling (already used)
- `gopkg.in/yaml.v3` - YAML parsing (already used)
- `k8s.io/apiextensions-apiserver` - For `apiextensionsv1.JSON` type

### Internal Dependencies

- Existing `deploy` package for manifest rendering
- Existing `status` helpers for conditions
- Existing test infrastructure (builders, envtest)

---

## Complexity Tracking

> No constitution violations requiring justification.

| Component | Complexity | Justification |
|-----------|------------|---------------|
| Deep Merge | Medium | Required for layered config model |
| Secret Resolution | Medium | Security requirement |
| Core Expansion | Low | Straightforward mapping |
| Version Detection | Low | Simple string comparison |
