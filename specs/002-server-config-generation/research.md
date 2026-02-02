# Research: Operator-Generated Server Configuration

**Feature Branch**: `002-server-config-generation`  
**Date**: 2026-02-02

---

## Overview

This document captures research findings and design decisions for the server configuration generation feature.

---

## 1. Base Configuration Extraction

### Question
How does the operator access the distribution's embedded `config.yaml` to use as a base?

### Research Findings

**Option A: Extract from container image at build time**
- Pre-extract config.yaml from each distribution image
- Embed in operator binary or ConfigMap
- **Pros**: Fast access, no runtime dependency
- **Cons**: Requires operator rebuild for new distributions, version coupling

**Option B: Extract at runtime via init container**
- Use init container to copy config.yaml from distribution image
- Mount via emptyDir volume
- **Pros**: Always uses actual distribution config
- **Cons**: Slower startup, adds complexity

**Option C: Use well-known distribution defaults embedded in operator**
- Operator contains default configs for known distributions
- Override only user-specified fields
- **Pros**: Predictable, fast, testable
- **Cons**: Must update operator for new distributions

**Option D: Read from distribution image metadata**
- Store config.yaml in image labels or annotations
- Operator reads via container registry API
- **Pros**: No runtime extraction needed
- **Cons**: Requires registry access, non-standard

### Decision
**Option C: Use well-known distribution defaults embedded in operator**

**Rationale**:
- The operator already maintains a `distributionImages` map for supported distributions
- Embedding default configs follows the same pattern
- Provides clear version support policy (n and n-1)
- Enables validation at apply time without running containers
- Can be extended to Option B in the future if needed

### Implementation
```go
// pkg/config/distributions.go
var DistributionConfigs = map[string][]byte{
    "starter":    starterConfigYAML,
    "rh-dev":     rhDevConfigYAML,
    // ... embedded via go:embed
}

//go:embed configs/starter.yaml
var starterConfigYAML []byte
```

---

## 2. Deep Merge Strategy

### Question
How should the operator merge `global.raw` over expanded `global.core` configuration?

### Research Findings

**Standard YAML merge behaviors**:
1. **Replace**: New value completely replaces old value
2. **Append**: Arrays are concatenated
3. **Merge**: Maps are recursively merged

**Kubernetes precedents**:
- Strategic Merge Patch: Uses merge keys for arrays
- JSON Merge Patch (RFC 7386): Replace semantics for arrays
- Kustomize: Various strategies per field

**llama-stack config.yaml specifics**:
- `providers` is a map of arrays (e.g., `providers.inference: [...]`)
- `registered_resources` contains arrays of registrations
- Most fields are simple scalar or map types

### Decision
**Use RFC 7386 JSON Merge Patch semantics with provider_id merge keys**

**Merge Rules**:
1. Scalars: Raw replaces Core
2. Maps: Recursive merge
3. Arrays: Replace by default
4. Provider arrays: Merge by `provider_id` (if present)
5. Resource arrays: Merge by `*_id` field (model_id, shield_id, etc.)

**Rationale**:
- Consistent with Kubernetes patterns
- Provider/resource ID merge enables partial updates
- Replace for arrays is safest default
- Matches user expectations from Kustomize experience

### Implementation
```go
// controllers/deep_merge.go
func DeepMerge(base, overlay map[string]interface{}, opts MergeOptions) map[string]interface{} {
    // For each key in overlay:
    // 1. If not in base, add
    // 2. If both maps, recurse
    // 3. If both slices with merge key, merge by key
    // 4. Otherwise, replace
}

type MergeOptions struct {
    // Map of path patterns to merge keys
    // e.g., "providers.*" -> "provider_id"
    ArrayMergeKeys map[string]string
}
```

---

## 3. Secret Reference Detection

### Question
How should the operator detect and resolve `secretKeyRef` references throughout the configuration?

### Research Findings

**Secret reference patterns in config.yaml**:
```yaml
# Pattern 1: Direct in config value
api_key: "${env.OPENAI_API_KEY}"

# Pattern 2: Kubernetes-style valueFrom (proposed for CR)
apiKey:
  secretKeyRef:
    name: openai-secret
    key: api-key
```

**Environment variable naming conventions**:
- Kubernetes convention: `<PREFIX>_<PATH>_<KEY>`
- Must be valid shell variable names (alphanumeric + underscore)
- Should be deterministic for debugging

### Decision
**Use structured secretKeyRef in CR, convert to ${env.VAR} in generated config**

**Naming Convention**: `LLSD_<SECTION>_<SUBSECTION>_<KEY>`

Examples:
- `global.core.inference.apiKey` → `LLSD_INFERENCE_API_KEY`
- `global.core.storage.connectionString` → `LLSD_STORAGE_CONNECTION_STRING`
- `global.raw.providers.inference[0].config.api_key` → `LLSD_PROVIDER_INFERENCE_0_API_KEY`

**Rationale**:
- Structured `secretKeyRef` is Kubernetes-native
- Deterministic naming aids debugging
- Prefix prevents collision with other env vars
- Uppercase is shell convention

### Implementation
```go
// controllers/secret_resolver.go
type SecretResolver struct {
    client client.Client
    namespace string
}

type ResolvedSecret struct {
    EnvVarName  string
    SecretName  string
    SecretKey   string
    Placeholder string  // ${env.VAR}
}

func (r *SecretResolver) Resolve(config map[string]interface{}) ([]ResolvedSecret, map[string]interface{}, error) {
    // Walk tree, find secretKeyRef, replace with placeholder, return list
}
```

---

## 4. Provider Type Mapping

### Question
What provider types should be supported and how do simple names map to full types?

### Research Findings

**llama-stack provider types** (from config.yaml schema):
- `remote::ollama` - Ollama inference
- `remote::vllm` - vLLM inference
- `remote::openai` - OpenAI-compatible APIs
- `remote::bedrock` - AWS Bedrock
- `remote::azure` - Azure OpenAI
- `remote::anthropic` - Anthropic Claude
- `remote::gemini` - Google Gemini
- `remote::together` - Together AI
- `remote::fireworks` - Fireworks AI
- `remote::groq` - Groq
- `remote::nvidia` - NVIDIA NIM

**Other provider categories**:
- `memory` storage backends
- `vector_io` providers (chromadb, weaviate, pgvector)
- `tool_runtime` providers
- `safety` providers

### Decision
**Support common inference providers in core, full types in raw**

**core.inference.provider enum**:
```
ollama, vllm, openai, bedrock, azure, anthropic, gemini, together, fireworks, groq, nvidia
```

**Expansion mapping**:
```go
var InferenceProviderMap = map[string]string{
    "ollama":    "remote::ollama",
    "vllm":      "remote::vllm",
    "openai":    "remote::openai",
    "bedrock":   "remote::bedrock",
    "azure":     "remote::azure",
    "anthropic": "remote::anthropic",
    "gemini":    "remote::gemini",
    "together":  "remote::together",
    "fireworks": "remote::fireworks",
    "groq":      "remote::groq",
    "nvidia":    "remote::nvidia",
}
```

**Rationale**:
- Covers most common use cases
- Simple names reduce configuration errors
- Full type access via `global.raw` for advanced users

---

## 5. ConfigMap Naming and Lifecycle

### Question
How should generated ConfigMaps be named and managed?

### Research Findings

**Naming patterns in Kubernetes**:
- Hash suffix: `myapp-config-abc123` (like Deployment pod template hash)
- Generation suffix: `myapp-config-gen-1`
- Timestamp suffix: `myapp-config-20240215`

**Lifecycle considerations**:
- Owner reference enables automatic cleanup
- Immutable ConfigMaps prevent accidental modification
- Need to track which ConfigMap is current
- May want to keep previous for rollback

### Decision
**Use generation-based naming with immutable ConfigMaps**

**Naming**: `<cr-name>-config-gen-<generation>`

**Lifecycle**:
1. Store `configGeneration` in CR status
2. Increment on each config change
3. Create new ConfigMap with new name
4. Update Deployment to reference new ConfigMap
5. Old ConfigMaps cleaned up by owner reference (keep n-1 for 1 reconcile cycle)

**ConfigMap spec**:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-llsd-config-gen-5
  ownerReferences:
    - apiVersion: llama.ai/v1alpha1
      kind: LlamaStackDistribution
      name: my-llsd
immutable: true
data:
  config.yaml: |
    version: 2
    ...
```

**Rationale**:
- Generation number provides clear audit trail
- Immutable prevents drift
- Owner reference ensures cleanup
- Matches existing patterns (PVC, Service)

---

## 6. Validation Strategy

### Question
What validation should happen at CRD admission vs runtime?

### Research Findings

**Kubebuilder validation capabilities**:
- `// +kubebuilder:validation:Enum=...` - Field enum
- `// +kubebuilder:validation:Required` - Required field
- `// +kubebuilder:validation:Pattern=...` - Regex pattern
- `// +kubebuilder:validation:XValidation:rule=...` - CEL expressions

**Runtime validation needs**:
- Secret existence check
- Cross-field validation (e.g., postgres requires connectionString)
- ConfigMap size limits
- Schema version compatibility

### Decision
**Maximize CRD validation, runtime validation for external references**

**CRD Validation (admission time)**:
- Provider enum values
- Required fields
- Mutual exclusivity (config vs userConfig)
- Field patterns (endpoint URLs)
- Cross-field rules via XValidation

**Runtime Validation (reconciliation time)**:
- Secret existence
- ConfigMap size limits
- Base config.yaml version compatibility
- Generated config.yaml validity

**Rationale**:
- Fail fast at admission for most errors
- Runtime validation for external dependencies
- Matches constitution §2.1 preference for CRD validation

---

## 7. Status Reporting

### Question
What status conditions should be added for configuration generation?

### Research Findings

**Existing conditions**:
- `DeploymentReady`
- `StorageReady`
- `ServiceReady`
- `HealthCheck`

**New conditions needed**:
- Configuration validation status
- ConfigMap generation status
- Secret resolution status

### Decision
**Add ConfigReady condition**

**Condition Definition**:
```go
const (
    ConditionTypeConfigReady = "ConfigReady"
    
    ReasonConfigGenerated     = "ConfigGenerated"
    ReasonConfigValidationFailed = "ValidationFailed"
    ReasonSecretNotFound      = "SecretNotFound"
    ReasonVersionNotSupported = "VersionNotSupported"
    
    MessageConfigReady = "Server configuration generated successfully"
)
```

**Status Fields**:
```go
type LlamaStackDistributionStatus struct {
    // ... existing fields
    
    // ConfigGeneration tracks the current config generation number
    ConfigGeneration int64 `json:"configGeneration,omitempty"`
    
    // ConfigHash is the hash of the current configuration
    ConfigHash string `json:"configHash,omitempty"`
}
```

**Rationale**:
- Single condition covers config lifecycle
- Reason/message provides specific failure info
- Generation/hash in status aids debugging

---

## 8. Backward Compatibility

### Question
How to maintain backward compatibility with existing `userConfig.configMapName` approach?

### Research Findings

**Current behavior**:
- `spec.server.userConfig.configMapName` references external ConfigMap
- ConfigMap mounted at `/etc/llama-stack/`
- User owns ConfigMap lifecycle

**New behavior**:
- `spec.server.config` triggers operator-generated ConfigMap
- Operator manages ConfigMap lifecycle

**Conflict scenarios**:
- Both fields specified
- Migration from userConfig to config
- Rollback from config to userConfig

### Decision
**Mutual exclusivity with clear migration path**

**XValidation Rule**:
```go
// +kubebuilder:validation:XValidation:rule="!(has(self.config) && has(self.userConfig) && self.userConfig.configMapName != '')",message="config and userConfig.configMapName are mutually exclusive"
```

**Behavior Matrix**:

| config | userConfig.configMapName | Behavior |
|--------|--------------------------|----------|
| not set | not set | Use distribution defaults |
| set | not set | Generate ConfigMap from config |
| not set | set | Mount user ConfigMap (legacy) |
| set | set | Validation error |

**Migration Path**:
1. User removes `userConfig.configMapName`
2. User adds `config` section
3. Operator generates ConfigMap
4. Old user ConfigMap can be deleted

**Rationale**:
- Clear separation of concerns
- No ambiguity about which config applies
- Simple migration (one field change)
- Legacy path continues to work

---

## Summary

| Topic | Decision | Key Rationale |
|-------|----------|---------------|
| Base Config | Embedded in operator | Fast, testable, version controlled |
| Deep Merge | RFC 7386 with merge keys | Kubernetes-consistent, predictable |
| Secret Detection | Structured secretKeyRef | Kubernetes-native, debuggable |
| Provider Mapping | Simple → full type | User-friendly, extensible |
| ConfigMap Naming | Generation suffix | Audit trail, immutable |
| Validation | CRD + runtime | Fail fast, external checks at runtime |
| Status | ConfigReady condition | Single condition, clear reasons |
| Backward Compat | Mutual exclusivity | Clear behavior, simple migration |
