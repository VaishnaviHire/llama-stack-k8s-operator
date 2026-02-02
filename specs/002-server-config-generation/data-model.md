# Data Model: Operator-Generated Server Configuration

**Feature Branch**: `002-server-config-generation`  
**Date**: 2026-02-02

---

## Overview

This document defines the CRD type definitions for the server configuration generation feature.

---

## Type Definitions

### ServerSpec Extension

The existing `ServerSpec` type is extended with a new `Config` field:

```go
// ServerSpec defines the desired state of llama server.
type ServerSpec struct {
    Distribution  DistributionType `json:"distribution"`
    ContainerSpec ContainerSpec    `json:"containerSpec,omitempty"`
    // ... existing fields ...
    
    // Config defines operator-managed server configuration.
    // When specified, the operator generates a ConfigMap from this configuration.
    // Mutually exclusive with userConfig.configMapName.
    // +optional
    Config *ServerConfigSpec `json:"config,omitempty"`
    
    // UserConfig defines the user configuration for the llama-stack server
    // +optional
    UserConfig *UserConfigSpec `json:"userConfig,omitempty"`
}
```

### ServerConfigSpec

Top-level configuration structure:

```go
// ServerConfigSpec defines the server configuration that the operator
// will use to generate a config.yaml ConfigMap.
// +kubebuilder:validation:XValidation:rule="!(has(self.global) && has(self.global.core) && has(self.global.core.storage) && self.global.core.storage.type == 'postgres' && !has(self.global.core.storage.connectionString))",message="storage.connectionString is required when storage.type is postgres"
type ServerConfigSpec struct {
    // Global defines global server configuration (providers, storage, server settings).
    // +optional
    Global *GlobalConfigSpec `json:"global,omitempty"`
    
    // Resources defines resource registrations (models, tools, shields, etc.).
    // +optional
    Resources *ResourceConfigSpec `json:"resources,omitempty"`
}
```

### GlobalConfigSpec

Global configuration with core and raw sections:

```go
// GlobalConfigSpec defines global server configuration.
type GlobalConfigSpec struct {
    // Core provides simplified, strongly-typed configuration options
    // suitable for 80% of use cases.
    // +optional
    Core *GlobalCoreSpec `json:"core,omitempty"`
    
    // Raw provides full passthrough to config.yaml sections for advanced use cases.
    // Field names use exact config.yaml naming (snake_case).
    // Raw values are deep-merged over expanded Core values.
    // +optional
    Raw *GlobalRawSpec `json:"raw,omitempty"`
}
```

### GlobalCoreSpec

Simplified configuration for common use cases:

```go
// GlobalCoreSpec defines simplified global configuration options.
type GlobalCoreSpec struct {
    // Inference configures the inference provider.
    // +optional
    Inference *InferenceConfig `json:"inference,omitempty"`
    
    // Storage configures the storage backend.
    // +optional
    Storage *StorageConfig `json:"storage,omitempty"`
    
    // Safety configures safety features.
    // +optional
    Safety *SafetyConfig `json:"safety,omitempty"`
    
    // Telemetry configures telemetry/tracing.
    // +optional
    Telemetry *TelemetryConfig `json:"telemetry,omitempty"`
    
    // DisableApis lists APIs to disable.
    // +optional
    // +kubebuilder:validation:Items:Enum=inference;agents;safety;memory;datasetio;scoring;eval;post_training;tool_runtime;inspect;providers;routes;telemetry;vector_io;files
    DisableApis []string `json:"disableApis,omitempty"`
}
```

### InferenceConfig

Inference provider configuration:

```go
// InferenceConfig defines the inference provider configuration.
type InferenceConfig struct {
    // Provider is the inference provider type.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Enum=ollama;vllm;openai;bedrock;azure;anthropic;gemini;together;fireworks;groq;nvidia
    Provider string `json:"provider"`
    
    // Endpoint is the URL to the inference provider.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Pattern=`^https?://.*`
    Endpoint string `json:"endpoint"`
    
    // ApiKey is a reference to a Secret containing the API key.
    // +optional
    ApiKey *SecretValueSource `json:"apiKey,omitempty"`
}
```

### StorageConfig

Storage backend configuration:

```go
// StorageConfig defines the storage backend configuration.
type StorageConfig struct {
    // Type is the storage backend type.
    // +optional
    // +kubebuilder:default:="sqlite"
    // +kubebuilder:validation:Enum=sqlite;postgres
    Type string `json:"type,omitempty"`
    
    // ConnectionString is a reference to a Secret containing the database connection string.
    // Required when Type is postgres.
    // +optional
    ConnectionString *SecretValueSource `json:"connectionString,omitempty"`
}
```

### SafetyConfig

Safety feature configuration:

```go
// SafetyConfig defines safety feature configuration.
type SafetyConfig struct {
    // Enabled controls whether safety features are enabled.
    // +optional
    // +kubebuilder:default:=true
    Enabled *bool `json:"enabled,omitempty"`
}
```

### TelemetryConfig

Telemetry configuration:

```go
// TelemetryConfig defines telemetry configuration.
type TelemetryConfig struct {
    // Enabled controls whether telemetry is enabled.
    // +optional
    // +kubebuilder:default:=true
    Enabled *bool `json:"enabled,omitempty"`
}
```

### SecretValueSource

Reference to a Kubernetes Secret:

```go
// SecretValueSource references a value from a Kubernetes Secret.
type SecretValueSource struct {
    // SecretKeyRef references a specific key in a Secret.
    // +kubebuilder:validation:Required
    SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef"`
}
```

### GlobalRawSpec

Advanced passthrough configuration:

```go
// GlobalRawSpec provides full passthrough to config.yaml sections.
// Field names use exact config.yaml naming (snake_case).
type GlobalRawSpec struct {
    // Providers is the full providers configuration.
    // Structure: {"inference": [...], "safety": [...], "vector_io": [...], ...}
    // +optional
    Providers *apiextensionsv1.JSON `json:"providers,omitempty"`
    
    // Storage is the full storage configuration.
    // Structure: {"backends": {...}, "stores": {...}}
    // +optional
    Storage *apiextensionsv1.JSON `json:"storage,omitempty"`
    
    // Server is the server configuration (port, TLS, auth).
    // +optional
    Server *apiextensionsv1.JSON `json:"server,omitempty"`
    
    // VectorStores is the vector store configuration.
    // +optional
    VectorStores *apiextensionsv1.JSON `json:"vector_stores,omitempty"`
    
    // Safety is the safety defaults configuration.
    // +optional
    Safety *apiextensionsv1.JSON `json:"safety,omitempty"`
    
    // Connectors is the external connectors configuration.
    // +optional
    Connectors *apiextensionsv1.JSON `json:"connectors,omitempty"`
}
```

### ResourceConfigSpec

Resource registrations with core and raw sections:

```go
// ResourceConfigSpec defines resource registrations.
type ResourceConfigSpec struct {
    // Core provides simplified resource registration.
    // +optional
    Core *ResourceCoreSpec `json:"core,omitempty"`
    
    // Raw provides full resource definitions.
    // Field names use exact config.yaml naming (snake_case).
    // Raw values are deep-merged over expanded Core values.
    // +optional
    Raw *ResourceRawSpec `json:"raw,omitempty"`
}
```

### ResourceCoreSpec

Simplified resource registration:

```go
// ResourceCoreSpec defines simplified resource registration.
type ResourceCoreSpec struct {
    // Models is a list of model names to register.
    // Models are registered with the default inference provider.
    // +optional
    Models []string `json:"models,omitempty"`
    
    // Tools is a list of tool groups to enable.
    // +optional
    // +kubebuilder:validation:Items:Enum=websearch;rag;code_interpreter;memory;wolfram_alpha
    Tools []string `json:"tools,omitempty"`
}
```

### ResourceRawSpec

Full resource definitions:

```go
// ResourceRawSpec provides full resource definitions.
// Field names use exact config.yaml naming (snake_case) from registered_resources.
type ResourceRawSpec struct {
    // Models is the full model registration list.
    // +optional
    Models *apiextensionsv1.JSON `json:"models,omitempty"`
    
    // Shields is the shield registration list.
    // +optional
    Shields *apiextensionsv1.JSON `json:"shields,omitempty"`
    
    // VectorDbs is the vector database registration list.
    // +optional
    VectorDbs *apiextensionsv1.JSON `json:"vector_dbs,omitempty"`
    
    // ToolGroups is the tool group registration list.
    // +optional
    ToolGroups *apiextensionsv1.JSON `json:"tool_groups,omitempty"`
    
    // Datasets is the dataset registration list.
    // +optional
    Datasets *apiextensionsv1.JSON `json:"datasets,omitempty"`
    
    // ScoringFns is the scoring function registration list.
    // +optional
    ScoringFns *apiextensionsv1.JSON `json:"scoring_fns,omitempty"`
    
    // Benchmarks is the benchmark registration list.
    // +optional
    Benchmarks *apiextensionsv1.JSON `json:"benchmarks,omitempty"`
}
```

---

## Status Extensions

### LlamaStackDistributionStatus Extension

```go
// LlamaStackDistributionStatus defines the observed state of LlamaStackDistribution.
type LlamaStackDistributionStatus struct {
    // ... existing fields ...
    
    // ConfigGeneration tracks the current config generation number.
    // Incremented each time the generated ConfigMap changes.
    // +optional
    ConfigGeneration int64 `json:"configGeneration,omitempty"`
    
    // ConfigHash is the hash of the current generated configuration.
    // Used to detect configuration changes.
    // +optional
    ConfigHash string `json:"configHash,omitempty"`
    
    // GeneratedConfigMapName is the name of the currently active generated ConfigMap.
    // +optional
    GeneratedConfigMapName string `json:"generatedConfigMapName,omitempty"`
}
```

---

## Constants

```go
const (
    // DefaultStorageType is the default storage backend type
    DefaultStorageType = "sqlite"
    
    // DefaultSafetyEnabled is the default value for safety.enabled
    DefaultSafetyEnabled = true
    
    // DefaultTelemetryEnabled is the default value for telemetry.enabled
    DefaultTelemetryEnabled = true
    
    // ConfigMapGenerationPrefix is the prefix for generated ConfigMap names
    ConfigMapGenerationPrefix = "-config-gen-"
    
    // ConfigYAMLKey is the key in the ConfigMap for the config.yaml data
    ConfigYAMLKey = "config.yaml"
    
    // GeneratedConfigMountPath is where generated config is mounted
    GeneratedConfigMountPath = "/etc/llama-stack"
    
    // Condition types
    ConditionTypeConfigReady = "ConfigReady"
    
    // Condition reasons
    ReasonConfigGenerated        = "ConfigGenerated"
    ReasonConfigValidationFailed = "ValidationFailed"
    ReasonSecretNotFound         = "SecretNotFound"
    ReasonVersionNotSupported    = "VersionNotSupported"
    
    // Condition messages
    MessageConfigReady = "Server configuration generated successfully"
)
```

---

## Validation Rules

### CRD-Level Validation (Kubebuilder Markers)

```go
// Mutual exclusivity between config and userConfig
// +kubebuilder:validation:XValidation:rule="!(has(self.config) && has(self.userConfig) && self.userConfig.configMapName != '')",message="config and userConfig.configMapName are mutually exclusive"

// PostgreSQL requires connectionString
// +kubebuilder:validation:XValidation:rule="!(has(self.global) && has(self.global.core) && has(self.global.core.storage) && self.global.core.storage.type == 'postgres' && !has(self.global.core.storage.connectionString))",message="storage.connectionString is required when storage.type is postgres"

// Inference requires both provider and endpoint
// +kubebuilder:validation:XValidation:rule="!has(self.provider) || has(self.endpoint)",message="endpoint is required when provider is specified"
```

### Runtime Validation (Controller)

1. **Secret Existence**: Verify all referenced Secrets exist in namespace
2. **Provider ID Uniqueness**: Ensure no duplicate provider_id in raw config
3. **Config Size**: Validate generated config doesn't exceed ConfigMap limits
4. **Schema Version**: Check base config.yaml version is supported

---

## Entity Relationships

```text
LlamaStackDistribution
└── spec
    └── server
        ├── distribution (DistributionType)
        ├── config (ServerConfigSpec)           ← NEW
        │   ├── global (GlobalConfigSpec)
        │   │   ├── core (GlobalCoreSpec)
        │   │   │   ├── inference (InferenceConfig)
        │   │   │   │   ├── provider: string (enum)
        │   │   │   │   ├── endpoint: string
        │   │   │   │   └── apiKey: SecretValueSource
        │   │   │   ├── storage (StorageConfig)
        │   │   │   │   ├── type: string (enum)
        │   │   │   │   └── connectionString: SecretValueSource
        │   │   │   ├── safety (SafetyConfig)
        │   │   │   ├── telemetry (TelemetryConfig)
        │   │   │   └── disableApis: []string
        │   │   └── raw (GlobalRawSpec)
        │   │       ├── providers: JSON
        │   │       ├── storage: JSON
        │   │       ├── server: JSON
        │   │       ├── vector_stores: JSON
        │   │       ├── safety: JSON
        │   │       └── connectors: JSON
        │   └── resources (ResourceConfigSpec)
        │       ├── core (ResourceCoreSpec)
        │       │   ├── models: []string
        │       │   └── tools: []string
        │       └── raw (ResourceRawSpec)
        │           ├── models: JSON
        │           ├── shields: JSON
        │           ├── vector_dbs: JSON
        │           ├── tool_groups: JSON
        │           ├── datasets: JSON
        │           ├── scoring_fns: JSON
        │           └── benchmarks: JSON
        └── userConfig (UserConfigSpec)    ← Existing, mutually exclusive
```

---

## Example CR

```yaml
apiVersion: llama.ai/v1alpha1
kind: LlamaStackDistribution
metadata:
  name: my-llsd
  namespace: llama-stack
spec:
  replicas: 2
  server:
    distribution:
      name: starter
    
    config:
      global:
        core:
          inference:
            provider: vllm
            endpoint: "http://vllm:8000/v1"
            apiKey:
              secretKeyRef:
                name: vllm-creds
                key: api-key
          storage:
            type: postgres
            connectionString:
              secretKeyRef:
                name: pg-creds
                key: connection-url
          safety:
            enabled: true
          disableApis:
            - post_training
            - eval
        
        raw:
          server:
            port: 8321
      
      resources:
        core:
          models:
            - "llama3.2-8b"
            - "llama3.2-70b"
          tools:
            - websearch
            - rag
        
        raw:
          models:
            - model_id: "custom-embed"
              provider_id: vllm
              model_type: embedding
              metadata:
                embedding_dimension: 768
```
