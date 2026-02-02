# Feature Specification: Operator-Generated Server Configuration

**Feature Branch**: `002-server-config-generation`  
**Created**: 2026-02-02  
**Status**: Draft  
**Input**: User description: "Enable the llama-stack Kubernetes operator to generate server configuration from a high-level abstracted specification in the LlamaStackDistribution CR"  
**Related Issue**: [RHAISTRAT-1061](https://issues.redhat.com/browse/RHAISTRAT-1061), [GitHub #234](https://github.com/llamastack/llama-stack-k8s-operator/issues/234)

---

## Overview

Enable the `llama-stack` Kubernetes operator to generate the server configuration (`config.yaml`) from a high-level, abstracted specification in the `LlamaStackDistribution` CR, rather than requiring users to provide a complete ConfigMap. This provides a simplified, validated, and version-resilient configuration experience.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Simple Inference Configuration (Priority: P1)

A platform engineer wants to deploy LlamaStack with a vLLM inference backend and PostgreSQL storage using minimal configuration. They want the operator to handle configuration complexity while they specify only the essential connection details.

**Why this priority**: This represents the most common use case (80% of users) who need basic inference capabilities without understanding the full config.yaml schema. Delivers immediate value by dramatically reducing configuration complexity.

**Independent Test**: Can be fully tested by deploying a LlamaStackDistribution CR with only `global.core` settings and verifying the server starts successfully with the correct provider configuration.

**Acceptance Scenarios**:

1. **Given** a cluster with the operator installed, **When** user applies a CR with `global.core.inference.provider: vllm` and `global.core.inference.endpoint`, **Then** the operator generates a complete config.yaml with the vLLM provider configured and creates a ConfigMap mounted to the deployment.

2. **Given** a CR with `global.core.storage.type: postgres` and a secret reference for connection string, **When** the CR is applied, **Then** the operator resolves the secret reference to an environment variable and configures PostgreSQL storage in the generated config.

3. **Given** a CR with `global.core.disableApis: [post_training, eval]`, **When** the CR is applied, **Then** the generated config.yaml has those APIs disabled.

---

### User Story 2 - Distribution Defaults Without Configuration (Priority: P1)

A developer wants to quickly deploy LlamaStack using distribution defaults without providing any custom configuration, relying on environment variables for runtime overrides.

**Why this priority**: Enables zero-configuration deployments for evaluation and development scenarios. Critical for onboarding experience.

**Independent Test**: Can be tested by deploying a CR with no `spec.server.config` field and verifying the distribution's embedded config.yaml is used as-is.

**Acceptance Scenarios**:

1. **Given** a CR with only `spec.server.distribution.name` specified (no `config` field), **When** the CR is applied, **Then** the operator uses the distribution's embedded config.yaml without modification.

2. **Given** an existing deployment using distribution defaults, **When** the user later adds `spec.server.config`, **Then** the operator generates a new ConfigMap merging user config over defaults and triggers a rollout.

---

### User Story 3 - Backward Compatibility with Existing ConfigMap (Priority: P1)

An existing user has a complete ConfigMap-based configuration and wants to continue using it without changes after upgrading to the new operator version.

**Why this priority**: Critical for production deployments. Breaking existing workflows would block operator adoption.

**Independent Test**: Can be tested by applying a CR with `spec.server.userConfig.configMapName` and verifying the operator mounts the referenced ConfigMap without generating its own.

**Acceptance Scenarios**:

1. **Given** a CR with `spec.server.userConfig.configMapName` pointing to an existing ConfigMap, **When** the CR is applied, **Then** the operator mounts the user-provided ConfigMap and does not generate a new one.

2. **Given** a CR with both `userConfig.configMapName` and `config` fields, **When** the CR is applied, **Then** validation fails with a clear error indicating mutual exclusivity.

---

### User Story 4 - Advanced Multi-Provider Configuration (Priority: P2)

A power user needs to configure multiple inference providers (e.g., vLLM for LLM, a separate vLLM for embeddings, OpenAI as fallback) with custom parameters that the simplified `core` options don't expose.

**Why this priority**: Supports the 20% of advanced users who need full control. Essential for production deployments with complex requirements.

**Independent Test**: Can be tested by applying a CR with `global.raw.providers` containing multiple inference providers and verifying all providers appear in the generated config.yaml.

**Acceptance Scenarios**:

1. **Given** a CR with `global.raw.providers.inference` containing multiple provider definitions, **When** the CR is applied, **Then** all providers are included in the generated config.yaml with exact configurations specified.

2. **Given** a CR with both `global.core.inference` and `global.raw.providers.inference`, **When** the CR is applied, **Then** the raw configuration is deep-merged over the expanded core configuration, with raw values taking precedence.

3. **Given** a CR with `global.raw.server.tls.enabled: true` and TLS certificate paths, **When** the CR is applied, **Then** the generated config.yaml includes TLS configuration in the server section.

---

### User Story 5 - Secret Reference Resolution (Priority: P2)

A platform engineer needs to provide API keys and database credentials securely through Kubernetes Secrets, with the operator automatically wiring them as environment variables.

**Why this priority**: Security requirement for production deployments. Secrets must never appear in ConfigMaps or logs.

**Independent Test**: Can be tested by creating a Secret, referencing it in a CR via `secretKeyRef`, and verifying the deployment has the corresponding environment variable with `${env.VAR}` placeholder in config.yaml.

**Acceptance Scenarios**:

1. **Given** a CR with `global.core.inference.apiKey.secretKeyRef` pointing to a valid Secret, **When** the CR is applied, **Then** the operator creates an environment variable entry in the Deployment and the config.yaml contains `${env.LLSD_*}` placeholder.

2. **Given** a CR with a secret reference to a non-existent Secret, **When** the CR is applied, **Then** validation fails with a clear error identifying the missing Secret.

3. **Given** multiple secret references across the configuration, **When** the CR is applied, **Then** each secret gets a unique, deterministic environment variable name (e.g., `LLSD_INFERENCE_API_KEY`).

---

### User Story 6 - Simple Resource Registration (Priority: P2)

A user wants to register models and tools without understanding the full resource registration schema, just by listing model names and tool groups.

**Why this priority**: Simplifies the common task of making models and tools available. Reduces cognitive load for typical deployments.

**Independent Test**: Can be tested by specifying `resources.core.models` list and verifying the generated config.yaml contains proper `registered_resources.models` entries.

**Acceptance Scenarios**:

1. **Given** a CR with `resources.core.models: ["llama3.2-8b", "llama3.2-70b"]`, **When** the CR is applied, **Then** the generated config.yaml contains model registrations with the default provider and standard metadata.

2. **Given** a CR with `resources.core.tools: ["websearch", "rag"]`, **When** the CR is applied, **Then** the generated config.yaml enables the corresponding tool groups.

---

### User Story 7 - Advanced Resource Registration (Priority: P3)

A power user needs to register models with custom metadata, specific provider assignments, and embedding configurations.

**Why this priority**: Supports advanced use cases like custom model configurations, multiple embedding models, and vector databases.

**Independent Test**: Can be tested by specifying `resources.raw.models` with custom metadata and verifying exact reproduction in generated config.yaml.

**Acceptance Scenarios**:

1. **Given** a CR with `resources.raw.models` including `model_type`, `provider_id`, and `metadata`, **When** the CR is applied, **Then** the generated config.yaml contains the exact model registration with all specified fields.

2. **Given** a CR with both `resources.core.models` and `resources.raw.models`, **When** the CR is applied, **Then** raw models are deep-merged over expanded core models.

---

### User Story 8 - Configuration Change Triggers Rollout (Priority: P2)

A user updates the configuration in their CR and expects the changes to be applied to the running deployment without manual intervention.

**Why this priority**: Essential for operational workflows. Configuration changes must be applied reliably.

**Independent Test**: Can be tested by modifying a CR's config section and verifying a new pod rollout occurs with the updated ConfigMap.

**Acceptance Scenarios**:

1. **Given** a running LlamaStack deployment, **When** user modifies `spec.server.config` and re-applies the CR, **Then** the operator generates a new ConfigMap with updated content and triggers a deployment rollout via hash annotation.

2. **Given** a configuration change, **When** the new ConfigMap is generated, **Then** the ConfigMap name includes a generation suffix (e.g., `<name>-config-gen-<N>`) for traceability.

---

### User Story 9 - Configuration Validation at Apply Time (Priority: P2)

A user wants to know immediately if their configuration is invalid, rather than discovering issues when the server fails to start.

**Why this priority**: Improves user experience and reduces troubleshooting time. Validation catches errors early.

**Independent Test**: Can be tested by applying an invalid CR and verifying rejection with a descriptive error message.

**Acceptance Scenarios**:

1. **Given** a CR with an invalid `inference.provider` value, **When** the CR is applied, **Then** admission validation rejects the CR with an error identifying valid provider options.

2. **Given** a CR with `storage.type: postgres` but missing `connectionString`, **When** the CR is applied, **Then** validation fails with an error indicating the required field.

3. **Given** a CR with duplicate `provider_id` values in `global.raw.providers`, **When** the CR is applied, **Then** validation fails with an error identifying the duplicate.

---

### User Story 10 - Version Schema Compatibility (Priority: P3)

The operator must handle config.yaml schema versions gracefully, supporting current and previous versions while rejecting unsupported older versions.

**Why this priority**: Ensures smooth upgrades and clear messaging when versions are incompatible.

**Independent Test**: Can be tested by using a distribution with different config.yaml versions and verifying appropriate behavior.

**Acceptance Scenarios**:

1. **Given** a distribution with current config.yaml schema version (n), **When** the CR is applied, **Then** the operator generates configuration normally.

2. **Given** a distribution with previous schema version (n-1), **When** the CR is applied, **Then** the operator logs a deprecation warning but generates configuration successfully.

3. **Given** a distribution with unsupported schema version (n-2 or older), **When** the CR is applied, **Then** the operator rejects the CR with an error indicating the version is not supported and suggesting upgrade.

---

### Edge Cases

- What happens when the referenced Secret is deleted after deployment?
  - The deployment continues running with cached environment variables. Next reconciliation logs a warning in the CR status condition.

- How does the system handle empty `config` field vs. absent `config` field?
  - Both are treated identically: use distribution defaults.

- What happens if `global.raw` specifies a provider that conflicts with `global.core`?
  - `global.raw` takes precedence (deep merge behavior).

- How does the system handle very large configurations that exceed ConfigMap size limits?
  - Validation fails with a descriptive error before creating the ConfigMap.

- What happens if the distribution's embedded config.yaml is malformed?
  - Operator logs an error and sets a Failed condition on the CR with details.

---

## Requirements *(mandatory)*

### Functional Requirements

#### CRD Schema & Validation

- **FR-001**: System MUST extend `LlamaStackDistributionSpec.Server` with a new `config` field containing `global` and `resources` sub-structures.

- **FR-002**: System MUST support `global.core` with strongly-typed fields for common configuration options: `inference` (provider, endpoint, apiKey), `storage` (type, connectionString), `safety` (enabled), `telemetry` (enabled), and `disableApis`.

- **FR-003**: System MUST support `global.raw` as an unstructured passthrough for advanced config.yaml sections: `providers`, `storage`, `server`, `vector_stores`, `safety`, `connectors`.

- **FR-004**: System MUST support `resources.core` with simplified registration: `models` (list of strings) and `tools` (list of strings).

- **FR-005**: System MUST support `resources.raw` for full resource definitions: `models`, `shields`, `vector_dbs`, `tool_groups`, `datasets`, `scoring_fns`, `benchmarks`.

- **FR-006**: System MUST validate that `spec.server.config` and `spec.server.userConfig.configMapName` are mutually exclusive.

- **FR-007**: System MUST validate inference provider values against supported list: `ollama`, `vllm`, `openai`, `bedrock`, `azure`, `anthropic`, `gemini`, `together`, `fireworks`, `groq`, `nvidia`.

- **FR-008**: System MUST validate that `storage.connectionString` is required when `storage.type` is `postgres`.

- **FR-009**: System MUST validate that all `provider_id` values are unique across provider configurations.

- **FR-010**: System MUST validate that referenced Secrets exist in the same namespace.

#### Configuration Generation

- **FR-011**: System MUST read the distribution's embedded `config.yaml` as the base configuration when `spec.server.config` is specified.

- **FR-012**: System MUST expand `global.core` settings to full provider/storage configuration using documented provider type mappings.

- **FR-013**: System MUST deep-merge `global.raw` over the expanded `global.core` configuration.

- **FR-014**: System MUST expand `resources.core` to full resource registrations with default provider assignment.

- **FR-015**: System MUST deep-merge `resources.raw` over the expanded `resources.core` configuration.

- **FR-016**: System MUST preserve all fields from the base config.yaml that are not explicitly overridden.

#### Secret Handling

- **FR-017**: System MUST resolve `valueFrom.secretKeyRef` references to environment variables in the Deployment.

- **FR-018**: System MUST generate deterministic, namespaced environment variable names (e.g., `LLSD_INFERENCE_API_KEY`).

- **FR-019**: System MUST replace secret references in the generated config.yaml with `${env.VAR}` placeholders.

- **FR-020**: System MUST NOT include secret values in ConfigMaps or logs.

#### ConfigMap Management

- **FR-021**: System MUST generate ConfigMap with name format `<cr-name>-config-gen-<generation>`.

- **FR-022**: System MUST set owner reference on generated ConfigMap for garbage collection.

- **FR-023**: System MUST generate an immutable ConfigMap for the configuration.

- **FR-024**: System MUST update Deployment annotation with configuration hash to trigger rollouts on changes.

#### Backward Compatibility

- **FR-025**: System MUST continue to support `spec.server.userConfig.configMapName` for user-provided ConfigMaps.

- **FR-026**: System MUST use distribution defaults when neither `config` nor `userConfig` is specified.

#### Version Support

- **FR-027**: System MUST support current (n) and previous (n-1) config.yaml schema versions.

- **FR-028**: System MUST log deprecation warnings for n-1 schema version.

- **FR-029**: System MUST reject unsupported schema versions (n-2 and older) with clear error messages.

- **FR-030**: System MUST read the `version` field from the distribution's base config.yaml to determine schema version.

#### Status Reporting

- **FR-031**: System MUST report configuration generation status in CR conditions.

- **FR-032**: System MUST report validation errors with specific field paths and corrective guidance.

---

### Key Entities

- **Config**: Top-level configuration structure containing `global` and `resources` sections.

- **GlobalConfig**: Configuration for providers, storage, server settings. Contains `core` (simplified) and `raw` (advanced) sub-structures.

- **ResourceConfig**: Configuration for registered resources (models, tools, etc.). Contains `core` (simplified) and `raw` (advanced) sub-structures.

- **SecretRef**: Reference to a Kubernetes Secret for secure value injection. Uses standard `secretKeyRef` structure.

- **InferenceConfig**: Simplified inference provider configuration with `provider`, `endpoint`, and `apiKey` fields.

- **StorageConfig**: Simplified storage configuration with `type` and `connectionString` fields.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can deploy a functional LlamaStack with inference and storage in under 20 lines of YAML (compared to 100+ lines for full ConfigMap approach).

- **SC-002**: Configuration changes applied to the CR result in deployment updates within 60 seconds.

- **SC-003**: 100% of existing deployments using `userConfig.configMapName` continue to work without modification after operator upgrade.

- **SC-004**: All validation errors provide specific field paths and actionable guidance, enabling users to fix issues without consulting documentation.

- **SC-005**: Zero secrets appear in ConfigMaps, logs, or status conditions.

- **SC-006**: Distribution upgrades that change config.yaml schema do not require user CR modifications for supported versions (n and n-1).

- **SC-007**: Users can configure 80% of common use cases using only `global.core` and `resources.core` without touching `raw` sections.

---

## Assumptions

- The `/admin` API is disabled for llama-stack server (per feature document).
- The distribution images contain an embedded `config.yaml` that can be read by the operator.
- The llama-stack project increments the config.yaml `version` field for any breaking changes.
- Users deploying with this feature have already provisioned any required external services (PostgreSQL, vLLM, etc.).
- Standard Kubernetes RBAC allows the operator to read Secrets in the same namespace as the CR.

---

## Out of Scope

- 1:1 mapping of every config.yaml field to LLSD spec (use `raw` for advanced fields).
- Support for config.yaml version 1 schema (only v2 and later supported).
- Runtime configuration changes (all changes require pod restart).
- Automatic provisioning of external services (PostgreSQL, inference backends, etc.).
