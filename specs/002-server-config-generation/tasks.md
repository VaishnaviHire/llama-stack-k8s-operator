# Tasks: Operator-Generated Server Configuration

**Feature Branch**: `002-server-config-generation`  
**Input**: Design documents from `/specs/002-server-config-generation/`  
**Prerequisites**: plan.md (required), spec.md (required), data-model.md, research.md, quickstart.md

**Tests**: Unit tests included per constitution requirement (§6.1 Table-Driven Tests)

**Organization**: Tasks are grouped to support incremental delivery. User stories are grouped into logical phases based on shared infrastructure requirements.

---

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Project Infrastructure)

**Purpose**: Create new files and package structure for the feature

- [ ] T001 Create pkg/config/ package directory structure
- [ ] T002 [P] Create api/v1alpha1/config_types.go with package declaration and imports
- [ ] T003 [P] Create pkg/config/schema.go with provider type mapping constants
- [ ] T004 [P] Create controllers/deep_merge.go with package declaration
- [ ] T005 [P] Create controllers/secret_resolver.go with package declaration
- [ ] T006 [P] Create controllers/config_generator.go with package declaration

**Checkpoint**: Package structure ready for implementation

---

## Phase 2: Foundational - CRD Types (Blocking Prerequisites)

**Purpose**: Define all CRD types that ALL user stories depend on. MUST complete before user story work.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Core Type Definitions

- [ ] T007 Define SecretValueSource type with secretKeyRef field in api/v1alpha1/config_types.go
- [ ] T008 [P] Define InferenceConfig type with provider, endpoint, apiKey fields in api/v1alpha1/config_types.go
- [ ] T009 [P] Define StorageConfig type with type, connectionString fields in api/v1alpha1/config_types.go
- [ ] T010 [P] Define SafetyConfig type with enabled field in api/v1alpha1/config_types.go
- [ ] T011 [P] Define TelemetryConfig type with enabled field in api/v1alpha1/config_types.go
- [ ] T012 Define GlobalCoreSpec type combining inference, storage, safety, telemetry, disableApis in api/v1alpha1/config_types.go
- [ ] T013 [P] Define GlobalRawSpec type with providers, storage, server, vector_stores, safety, connectors JSON fields in api/v1alpha1/config_types.go
- [ ] T014 Define GlobalConfigSpec type with core and raw fields in api/v1alpha1/config_types.go
- [ ] T015 [P] Define ResourceCoreSpec type with models and tools string arrays in api/v1alpha1/config_types.go
- [ ] T016 [P] Define ResourceRawSpec type with models, shields, vector_dbs, tool_groups JSON fields in api/v1alpha1/config_types.go
- [ ] T017 Define ResourceConfigSpec type with core and raw fields in api/v1alpha1/config_types.go
- [ ] T018 Define ServerConfigSpec type with global and resources fields in api/v1alpha1/config_types.go

### CRD Integration

- [ ] T019 Add Config *ServerConfigSpec field to ServerSpec in api/v1alpha1/llamastackdistribution_types.go
- [ ] T020 Add kubebuilder validation markers for provider enum (ollama, vllm, openai, etc.) in api/v1alpha1/config_types.go
- [ ] T021 Add kubebuilder validation markers for storage type enum (sqlite, postgres) in api/v1alpha1/config_types.go
- [ ] T022 Add kubebuilder validation markers for tools enum (websearch, rag, etc.) in api/v1alpha1/config_types.go
- [ ] T023 Add XValidation for mutual exclusivity of config and userConfig.configMapName in api/v1alpha1/llamastackdistribution_types.go
- [ ] T024 Add XValidation for postgres requiring connectionString in api/v1alpha1/config_types.go

### Status Extensions

- [ ] T025 Add ConfigGeneration int64 field to LlamaStackDistributionStatus in api/v1alpha1/llamastackdistribution_types.go
- [ ] T026 Add ConfigHash string field to LlamaStackDistributionStatus in api/v1alpha1/llamastackdistribution_types.go
- [ ] T027 Add GeneratedConfigMapName string field to LlamaStackDistributionStatus in api/v1alpha1/llamastackdistribution_types.go

### Constants and Condition Helpers

- [ ] T028 Add config-related constants (DefaultStorageType, ConfigMapGenerationPrefix, etc.) in api/v1alpha1/config_types.go
- [ ] T029 Add ConditionTypeConfigReady and related reason/message constants in controllers/status.go
- [ ] T030 Add SetConfigReadyCondition helper function in controllers/status.go

### Code Generation

- [ ] T031 Run make generate to regenerate zz_generated.deepcopy.go
- [ ] T032 Run make manifests to regenerate CRD YAML files in config/crd/bases/

**Checkpoint**: CRD types complete, manifests generated. User story implementation can begin.

---

## Phase 3: Core Infrastructure - Deep Merge & Expander

**Purpose**: Implement core algorithms required by multiple user stories

### Deep Merge Implementation (Required by US1, US4, US7)

- [ ] T033 Implement DeepMerge function for map[string]interface{} in controllers/deep_merge.go
- [ ] T034 Add MergeOptions struct with ArrayMergeKeys map in controllers/deep_merge.go
- [ ] T035 Implement slice merge with provider_id/model_id merge key support in controllers/deep_merge.go
- [ ] T036 Add unit tests for DeepMerge with nested maps in controllers/deep_merge_test.go
- [ ] T037 [P] Add unit tests for DeepMerge with slice replacement in controllers/deep_merge_test.go
- [ ] T038 [P] Add unit tests for DeepMerge with merge key matching in controllers/deep_merge_test.go
- [ ] T039 Add unit tests for DeepMerge edge cases (nil, empty) in controllers/deep_merge_test.go

### Core-to-Raw Expansion (Required by US1, US6)

- [ ] T040 Define ProviderTypeMap constant in pkg/config/schema.go (ollama→remote::ollama, etc.)
- [ ] T041 Implement ExpandGlobalCore function to convert GlobalCoreSpec to providers config in pkg/config/expander.go
- [ ] T042 Implement ExpandResourceCore function to convert ResourceCoreSpec to registered_resources in pkg/config/expander.go
- [ ] T043 Add unit tests for ExpandGlobalCore with each provider type in pkg/config/expander_test.go
- [ ] T044 [P] Add unit tests for ExpandResourceCore with models list in pkg/config/expander_test.go
- [ ] T045 [P] Add unit tests for ExpandResourceCore with tools list in pkg/config/expander_test.go

**Checkpoint**: Core algorithms ready for user story integration

---

## Phase 4: User Story 1 - Simple Inference Configuration (Priority: P1) 🎯 MVP

**Goal**: Deploy LlamaStack with vLLM/PostgreSQL using minimal config (global.core)

**Independent Test**: Apply CR with global.core.inference settings, verify ConfigMap generated with correct provider config

### Implementation for User Story 1

- [ ] T046 [US1] Create base distribution config loader in pkg/config/distributions.go
- [ ] T047 [US1] Embed starter distribution config.yaml using go:embed in pkg/config/distributions.go
- [ ] T048 [US1] Implement loadBaseConfig function to parse distribution config in controllers/config_generator.go
- [ ] T049 [US1] Implement generateConfig function combining expand + merge in controllers/config_generator.go
- [ ] T050 [US1] Implement createGeneratedConfigMap function with immutable flag in controllers/config_generator.go
- [ ] T051 [US1] Add owner reference to generated ConfigMap in controllers/config_generator.go
- [ ] T052 [US1] Implement reconcileGeneratedConfig function in controllers/llamastackdistribution_controller.go
- [ ] T053 [US1] Call reconcileGeneratedConfig from reconcileResources in controllers/llamastackdistribution_controller.go
- [ ] T054 [US1] Add generated ConfigMap mount to Deployment in controllers/resource_helper.go
- [ ] T055 [US1] Update buildManifestContext to include generated config hash in controllers/llamastackdistribution_controller.go

### Tests for User Story 1

- [ ] T056 [P] [US1] Add unit test for generateConfig with inference provider in controllers/config_generator_test.go
- [ ] T057 [P] [US1] Add unit test for generateConfig with postgres storage in controllers/config_generator_test.go
- [ ] T058 [P] [US1] Add unit test for generateConfig with disableApis in controllers/config_generator_test.go
- [ ] T059 [US1] Add controller test for CR with global.core creating ConfigMap in controllers/llamastackdistribution_controller_test.go

**Checkpoint**: User Story 1 complete - simple inference configuration works

---

## Phase 5: User Story 5 - Secret Reference Resolution (Priority: P2)

**Goal**: Securely inject API keys and credentials via environment variables

**Independent Test**: Apply CR with secretKeyRef, verify Deployment has env var and config.yaml has ${env.VAR} placeholder

### Implementation for User Story 5

- [ ] T060 [US5] Define ResolvedSecret struct with EnvVarName, SecretName, SecretKey, Placeholder in controllers/secret_resolver.go
- [ ] T061 [US5] Implement detectSecretRefs function to walk config tree finding secretKeyRef in controllers/secret_resolver.go
- [ ] T062 [US5] Implement generateEnvVarName function with LLSD_ prefix in controllers/secret_resolver.go
- [ ] T063 [US5] Implement replaceSecretRefs function to substitute ${env.VAR} placeholders in controllers/secret_resolver.go
- [ ] T064 [US5] Implement buildEnvVars function to create corev1.EnvVar list in controllers/secret_resolver.go
- [ ] T065 [US5] Implement validateSecretsExist function to check secrets in namespace in controllers/secret_resolver.go
- [ ] T066 [US5] Integrate secret resolution into generateConfig flow in controllers/config_generator.go
- [ ] T067 [US5] Add resolved env vars to container spec in controllers/resource_helper.go

### Tests for User Story 5

- [ ] T068 [P] [US5] Add unit test for detectSecretRefs in nested config in controllers/secret_resolver_test.go
- [ ] T069 [P] [US5] Add unit test for generateEnvVarName determinism in controllers/secret_resolver_test.go
- [ ] T070 [P] [US5] Add unit test for replaceSecretRefs placeholder format in controllers/secret_resolver_test.go
- [ ] T071 [US5] Add unit test for buildEnvVars with multiple secrets in controllers/secret_resolver_test.go

**Checkpoint**: User Story 5 complete - secrets securely resolved to env vars

---

## Phase 6: User Story 2 - Distribution Defaults (Priority: P1)

**Goal**: Deploy without config field, use distribution's embedded config.yaml

**Independent Test**: Apply CR with only distribution.name, verify server starts with distribution defaults

### Implementation for User Story 2

- [ ] T072 [US2] Add hasServerConfig helper function in controllers/llamastackdistribution_controller.go
- [ ] T073 [US2] Skip config generation when spec.server.config is nil in reconcileGeneratedConfig in controllers/llamastackdistribution_controller.go
- [ ] T074 [US2] Ensure deployment uses distribution default config when no config specified in controllers/resource_helper.go

### Tests for User Story 2

- [ ] T075 [US2] Add controller test for CR without config field using defaults in controllers/llamastackdistribution_controller_test.go

**Checkpoint**: User Story 2 complete - distribution defaults work

---

## Phase 7: User Story 3 - Backward Compatibility (Priority: P1)

**Goal**: Existing userConfig.configMapName continues to work

**Independent Test**: Apply CR with userConfig.configMapName, verify user ConfigMap mounted (no generated ConfigMap)

### Implementation for User Story 3

- [ ] T076 [US3] Ensure reconcileGeneratedConfig skips when userConfig.configMapName set in controllers/llamastackdistribution_controller.go
- [ ] T077 [US3] Add runtime validation for mutual exclusivity in reconcileResources in controllers/llamastackdistribution_controller.go

### Tests for User Story 3

- [ ] T078 [US3] Add controller test for CR with userConfig.configMapName mounting user config in controllers/llamastackdistribution_controller_test.go
- [ ] T079 [US3] Add validation test for CR with both config and userConfig rejecting in controllers/llamastackdistribution_controller_test.go

**Checkpoint**: User Story 3 complete - legacy config path preserved

---

## Phase 8: User Story 4 - Advanced Multi-Provider Configuration (Priority: P2)

**Goal**: Configure multiple providers using global.raw

**Independent Test**: Apply CR with global.raw.providers, verify all providers in generated config.yaml

### Implementation for User Story 4

- [ ] T080 [US4] Ensure GlobalRawSpec fields properly merge over expanded core in controllers/config_generator.go
- [ ] T081 [US4] Handle multiple providers in raw.providers.inference array in pkg/config/expander.go

### Tests for User Story 4

- [ ] T082 [P] [US4] Add unit test for raw providers overriding core inference in controllers/config_generator_test.go
- [ ] T083 [US4] Add unit test for multiple inference providers from raw in controllers/config_generator_test.go

**Checkpoint**: User Story 4 complete - advanced multi-provider works

---

## Phase 9: User Story 6 - Simple Resource Registration (Priority: P2)

**Goal**: Register models and tools with simple string lists

**Independent Test**: Apply CR with resources.core.models list, verify registered_resources in config.yaml

### Implementation for User Story 6

- [ ] T084 [US6] Implement model expansion in ExpandResourceCore assigning default provider in pkg/config/expander.go
- [ ] T085 [US6] Implement tools expansion in ExpandResourceCore enabling tool groups in pkg/config/expander.go
- [ ] T086 [US6] Merge expanded resources into final config in controllers/config_generator.go

### Tests for User Story 6

- [ ] T087 [P] [US6] Add unit test for model list expansion with default provider in pkg/config/expander_test.go
- [ ] T088 [US6] Add unit test for tools list expansion in pkg/config/expander_test.go

**Checkpoint**: User Story 6 complete - simple resource registration works

---

## Phase 10: User Story 8 - Configuration Change Triggers Rollout (Priority: P2)

**Goal**: Config changes trigger deployment rollout via hash annotation

**Independent Test**: Modify CR config, verify new ConfigMap created and pod restarts

### Implementation for User Story 8

- [ ] T089 [US8] Implement calculateConfigHash function in controllers/config_generator.go
- [ ] T090 [US8] Update ConfigGeneration in status on each config change in controllers/llamastackdistribution_controller.go
- [ ] T091 [US8] Add config hash annotation to deployment pod template in controllers/resource_helper.go
- [ ] T092 [US8] Implement ConfigMap naming with generation suffix (<name>-config-gen-<N>) in controllers/config_generator.go

### Tests for User Story 8

- [ ] T093 [US8] Add controller test for config change incrementing generation in controllers/llamastackdistribution_controller_test.go
- [ ] T094 [US8] Add controller test for hash annotation update triggering rollout in controllers/llamastackdistribution_controller_test.go

**Checkpoint**: User Story 8 complete - config changes trigger rollout

---

## Phase 11: User Story 9 - Configuration Validation (Priority: P2)

**Goal**: Invalid configurations fail fast with clear error messages

**Independent Test**: Apply CR with invalid provider, verify admission rejection with helpful message

### Implementation for User Story 9

- [ ] T095 [US9] Add runtime validation for secret existence in reconcileGeneratedConfig in controllers/llamastackdistribution_controller.go
- [ ] T096 [US9] Add runtime validation for duplicate provider_id in raw config in controllers/config_generator.go
- [ ] T097 [US9] Add ConfigMap size validation before creation in controllers/config_generator.go
- [ ] T098 [US9] Set ConfigReady condition to False with specific error on validation failure in controllers/llamastackdistribution_controller.go

### Tests for User Story 9

- [ ] T099 [P] [US9] Add test for missing secret validation error in controllers/config_generator_test.go
- [ ] T100 [US9] Add test for duplicate provider_id validation error in controllers/config_generator_test.go

**Checkpoint**: User Story 9 complete - validation catches errors early

---

## Phase 12: User Story 7 - Advanced Resource Registration (Priority: P3)

**Goal**: Register resources with custom metadata via resources.raw

**Independent Test**: Apply CR with resources.raw.models, verify exact fields in config.yaml

### Implementation for User Story 7

- [ ] T101 [US7] Ensure ResourceRawSpec fields merge over expanded core resources in controllers/config_generator.go
- [ ] T102 [US7] Support all resource types (shields, vector_dbs, tool_groups, etc.) in raw merge in controllers/config_generator.go

### Tests for User Story 7

- [ ] T103 [US7] Add unit test for raw models with custom metadata in controllers/config_generator_test.go
- [ ] T104 [US7] Add unit test for raw resources merging with core resources in controllers/config_generator_test.go

**Checkpoint**: User Story 7 complete - advanced resource registration works

---

## Phase 13: User Story 10 - Version Schema Compatibility (Priority: P3)

**Goal**: Support current (n) and previous (n-1) config schema versions

**Independent Test**: Use distribution with n-1 version, verify deprecation warning logged

### Implementation for User Story 10

- [ ] T105 [US10] Define supported schema versions as constants in pkg/config/version.go
- [ ] T106 [US10] Implement detectConfigVersion function to read version from base config in pkg/config/version.go
- [ ] T107 [US10] Implement validateSchemaVersion function with deprecation warning for n-1 in pkg/config/version.go
- [ ] T108 [US10] Integrate version validation into config generation flow in controllers/config_generator.go
- [ ] T109 [US10] Reject unsupported versions (n-2) with clear error in pkg/config/version.go

### Tests for User Story 10

- [ ] T110 [P] [US10] Add unit test for current version (n) passing in pkg/config/version_test.go
- [ ] T111 [P] [US10] Add unit test for previous version (n-1) with warning in pkg/config/version_test.go
- [ ] T112 [US10] Add unit test for unsupported version (n-2) rejection in pkg/config/version_test.go

**Checkpoint**: User Story 10 complete - version compatibility works

---

## Phase 14: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, samples, and final cleanup

### Test Builder Extensions

- [ ] T113 [P] Add WithConfig() method to DistributionBuilder in controllers/testing_support_test.go
- [ ] T114 [P] Add WithGlobalCore() method to DistributionBuilder in controllers/testing_support_test.go
- [ ] T115 [P] Add WithResources() method to DistributionBuilder in controllers/testing_support_test.go

### Sample CRs

- [ ] T116 [P] Create sample CR with global.core inference in config/samples/llama_v1alpha1_simple_config.yaml
- [ ] T117 [P] Create sample CR with global.raw providers in config/samples/llama_v1alpha1_advanced_config.yaml
- [ ] T118 [P] Create sample CR with resources.core models in config/samples/llama_v1alpha1_models_config.yaml

### Documentation

- [ ] T119 Update API documentation in docs/api-overview.md with new config fields
- [ ] T120 Update README.md with server configuration examples
- [ ] T121 Validate quickstart.md examples work end-to-end

### Final Validation

- [ ] T122 Run make lint to ensure all new code passes linting
- [ ] T123 Run make test to ensure all tests pass
- [ ] T124 Run make manifests to ensure CRD is up to date

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Phase 1: Setup
    ↓
Phase 2: Foundational (CRD Types) ──── BLOCKS ALL USER STORIES
    ↓
Phase 3: Core Infrastructure (Deep Merge, Expander)
    ↓
┌───────────────────────────────────────────────────────────────┐
│ User Stories can proceed in priority order or in parallel     │
├───────────────────────────────────────────────────────────────┤
│ Phase 4:  US1 (Simple Inference) ─────────── MVP 🎯           │
│ Phase 5:  US5 (Secret Resolution) ──────────┐                 │
│ Phase 6:  US2 (Distribution Defaults) ──────┤                 │
│ Phase 7:  US3 (Backward Compatibility) ─────┤ P1 Stories      │
│ Phase 8:  US4 (Advanced Multi-Provider) ────┤                 │
│ Phase 9:  US6 (Simple Resources) ───────────┤                 │
│ Phase 10: US8 (Rollout Triggers) ───────────┘                 │
│ Phase 11: US9 (Validation) ─────────────────┐                 │
│ Phase 12: US7 (Advanced Resources) ─────────┤ P2/P3 Stories   │
│ Phase 13: US10 (Version Compat) ────────────┘                 │
└───────────────────────────────────────────────────────────────┘
    ↓
Phase 14: Polish
```

### Critical Path (MVP)

1. **Phase 1**: Setup (T001-T006)
2. **Phase 2**: CRD Types (T007-T032) - ~26 tasks
3. **Phase 3**: Core Infrastructure (T033-T045) - ~13 tasks
4. **Phase 4**: User Story 1 (T046-T059) - ~14 tasks

**MVP Complete after Phase 4** - Users can deploy with simple inference config

### User Story Dependencies

| Story | Depends On | Can Start After |
|-------|------------|-----------------|
| US1 (P1) | Phase 3 | Core Infrastructure complete |
| US2 (P1) | US1 | Simple Inference works |
| US3 (P1) | US1 | Simple Inference works |
| US4 (P2) | US1 | Simple Inference works |
| US5 (P2) | Phase 3 | Core Infrastructure complete |
| US6 (P2) | US1 | Simple Inference works |
| US7 (P3) | US6 | Simple Resources works |
| US8 (P2) | US1 | Simple Inference works |
| US9 (P2) | US1 | Simple Inference works |
| US10 (P3) | US1 | Simple Inference works |

### Parallel Opportunities

Within Phase 2 (CRD Types):
```bash
# These can run in parallel - different type definitions:
T008 [P] InferenceConfig
T009 [P] StorageConfig
T010 [P] SafetyConfig
T011 [P] TelemetryConfig
T013 [P] GlobalRawSpec
T015 [P] ResourceCoreSpec
T016 [P] ResourceRawSpec
```

Within Phase 3 (Core Infrastructure):
```bash
# These can run in parallel - different test files:
T036, T037, T038, T039 - DeepMerge tests
T043, T044, T045 - Expander tests
```

User Stories (with multiple developers):
```bash
# After Phase 3 completes, these can run in parallel:
Developer A: US1 (Simple Inference)
Developer B: US5 (Secret Resolution)

# After US1 completes:
Developer A: US2, US3 (other P1 stories)
Developer B: US4, US8, US9 (P2 stories)
```

---

## Implementation Strategy

### MVP First (Phase 1-4 Only)

1. Complete Phase 1: Setup (~6 tasks)
2. Complete Phase 2: CRD Types (~26 tasks)
3. Complete Phase 3: Core Infrastructure (~13 tasks)
4. Complete Phase 4: User Story 1 (~14 tasks)
5. **STOP and VALIDATE**: Test simple inference configuration end-to-end
6. **MVP READY**: Users can deploy with minimal YAML

**MVP Task Count**: ~59 tasks

### Incremental Delivery

| Increment | User Stories | Cumulative Value |
|-----------|-------------|------------------|
| MVP | US1 | Simple inference with core config |
| +P1 | US2, US3 | Distribution defaults + backward compat |
| +P2 | US4, US5, US6, US8, US9 | Full production features |
| +P3 | US7, US10 | Advanced features + version compat |
| Polish | - | Documentation, samples |

### PR Strategy (from plan.md)

| PR | Phases | Tasks |
|----|--------|-------|
| PR 1 | Phase 1-2 | T001-T032 (CRD Types) |
| PR 2 | Phase 3 (partial) | T040-T045 (Expander) |
| PR 3 | Phase 3 (partial) | T033-T039 (Deep Merge) |
| PR 4 | Phase 5 | T060-T071 (Secret Resolution) |
| PR 5 | Phase 4 | T046-T059 (ConfigMap Generation) |
| PR 6 | Phase 6-10 | Controller Integration |
| PR 7 | Phase 7 | T076-T079 (Backward Compat) |
| PR 8 | Phase 14 | T113-T124 (Polish) |

---

## Summary

| Metric | Count |
|--------|-------|
| **Total Tasks** | 124 |
| **Setup Tasks** | 6 |
| **Foundational Tasks** | 26 |
| **User Story Tasks** | 79 |
| **Polish Tasks** | 13 |
| **Parallel Opportunities** | 47 tasks marked [P] |

### Tasks per User Story

| Story | Priority | Tasks | Notes |
|-------|----------|-------|-------|
| US1 | P1 | 14 | MVP - Simple Inference 🎯 |
| US2 | P1 | 4 | Distribution Defaults |
| US3 | P1 | 4 | Backward Compatibility |
| US4 | P2 | 4 | Advanced Multi-Provider |
| US5 | P2 | 12 | Secret Resolution |
| US6 | P2 | 5 | Simple Resources |
| US7 | P3 | 4 | Advanced Resources |
| US8 | P2 | 6 | Rollout Triggers |
| US9 | P2 | 6 | Validation |
| US10 | P3 | 8 | Version Compatibility |

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Constitution §6.1 requires table-driven tests - included
- Constitution §6.4 requires builder pattern - T113-T115 extend builders
- Each user story is independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
