# Quickstart: Operator-Generated Server Configuration

**Feature Branch**: `002-server-config-generation`  
**Date**: 2026-02-02

---

## Overview

This guide shows how to use the operator-generated server configuration feature to deploy LlamaStack with minimal YAML.

---

## Prerequisites

- Kubernetes cluster with the llama-stack operator installed
- kubectl configured to access the cluster
- (Optional) vLLM or other inference backend running

---

## Quick Start

### Example 1: Minimal Configuration with vLLM

Deploy LlamaStack with a vLLM inference backend using only 15 lines of YAML:

```yaml
apiVersion: llama.ai/v1alpha1
kind: LlamaStackDistribution
metadata:
  name: my-llsd
  namespace: llama-stack
spec:
  replicas: 1
  server:
    distribution:
      name: starter
    config:
      global:
        core:
          inference:
            provider: vllm
            endpoint: "http://vllm-service:8000/v1"
```

Apply and verify:

```bash
kubectl apply -f my-llsd.yaml
kubectl get llsd my-llsd -n llama-stack
```

### Example 2: With PostgreSQL Storage

Add persistent PostgreSQL storage:

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
            endpoint: "http://vllm-service:8000/v1"
          storage:
            type: postgres
            connectionString:
              secretKeyRef:
                name: pg-creds
                key: url
```

First, create the PostgreSQL credentials secret:

```bash
kubectl create secret generic pg-creds \
  --from-literal=url="postgresql://user:pass@postgres:5432/llama" \
  -n llama-stack
```

### Example 3: With API Key Authentication

Connect to an authenticated inference endpoint:

```yaml
apiVersion: llama.ai/v1alpha1
kind: LlamaStackDistribution
metadata:
  name: my-llsd
  namespace: llama-stack
spec:
  server:
    distribution:
      name: starter
    config:
      global:
        core:
          inference:
            provider: openai
            endpoint: "https://api.openai.com/v1"
            apiKey:
              secretKeyRef:
                name: openai-creds
                key: api-key
```

Create the API key secret:

```bash
kubectl create secret generic openai-creds \
  --from-literal=api-key="sk-..." \
  -n llama-stack
```

### Example 4: Register Models

Register specific models to be available via the API:

```yaml
apiVersion: llama.ai/v1alpha1
kind: LlamaStackDistribution
metadata:
  name: my-llsd
spec:
  server:
    distribution:
      name: starter
    config:
      global:
        core:
          inference:
            provider: vllm
            endpoint: "http://vllm:8000/v1"
      resources:
        core:
          models:
            - "llama3.2-8b"
            - "llama3.2-70b"
          tools:
            - websearch
            - rag
```

### Example 5: Advanced Configuration (Raw)

For power users who need full control, use the `raw` sections:

```yaml
apiVersion: llama.ai/v1alpha1
kind: LlamaStackDistribution
metadata:
  name: advanced-llsd
spec:
  server:
    distribution:
      name: starter
    config:
      global:
        core:
          storage:
            type: postgres
            connectionString:
              secretKeyRef:
                name: pg-creds
                key: url
        
        raw:
          providers:
            inference:
              - provider_id: vllm-llm
                provider_type: remote::vllm
                config:
                  base_url: "http://vllm-llm:8000/v1"
                  max_tokens: 8192
              - provider_id: vllm-embed
                provider_type: remote::vllm
                config:
                  base_url: "http://vllm-embed:8000/v1"
              - provider_id: openai-fallback
                provider_type: remote::openai
                config:
                  api_key: ${env.LLSD_OPENAI_API_KEY}
          
          server:
            port: 8321
      
      resources:
        raw:
          models:
            - model_id: "llama3.2-8b-custom"
              provider_id: vllm-llm
              model_type: llm
              metadata:
                context_length: 128000
            - model_id: "nomic-embed"
              provider_id: vllm-embed
              model_type: embedding
              metadata:
                embedding_dimension: 768
```

---

## Verifying Configuration

### Check Generated ConfigMap

The operator creates a ConfigMap with the generated configuration:

```bash
kubectl get configmaps -n llama-stack | grep config-gen
```

View the generated config:

```bash
kubectl get configmap my-llsd-config-gen-1 -n llama-stack -o yaml
```

### Check Status

View the configuration status in the CR:

```bash
kubectl get llsd my-llsd -n llama-stack -o jsonpath='{.status.conditions}'
```

Look for the `ConfigReady` condition:

```json
{
  "type": "ConfigReady",
  "status": "True",
  "reason": "ConfigGenerated",
  "message": "Server configuration generated successfully"
}
```

### Check Environment Variables

Verify secrets are injected as environment variables:

```bash
kubectl get deployment my-llsd -n llama-stack -o jsonpath='{.spec.template.spec.containers[0].env}'
```

You should see entries like:

```json
[
  {"name": "LLSD_INFERENCE_API_KEY", "valueFrom": {"secretKeyRef": {"name": "openai-creds", "key": "api-key"}}},
  {"name": "LLSD_STORAGE_CONNECTION_STRING", "valueFrom": {"secretKeyRef": {"name": "pg-creds", "key": "url"}}}
]
```

---

## Migration from userConfig

If you're currently using `userConfig.configMapName`, migration is straightforward:

### Before (Legacy)

```yaml
spec:
  server:
    userConfig:
      configMapName: my-config
```

### After (Operator-Generated)

```yaml
spec:
  server:
    config:
      global:
        core:
          inference:
            provider: vllm
            endpoint: "http://vllm:8000/v1"
```

Note: `config` and `userConfig.configMapName` are mutually exclusive. Remove one before adding the other.

---

## Supported Providers

### Inference Providers

| Provider | Description | Typical Endpoint |
|----------|-------------|------------------|
| `ollama` | Ollama | `http://ollama:11434` |
| `vllm` | vLLM | `http://vllm:8000/v1` |
| `openai` | OpenAI API | `https://api.openai.com/v1` |
| `bedrock` | AWS Bedrock | N/A (uses AWS credentials) |
| `azure` | Azure OpenAI | Deployment-specific |
| `anthropic` | Anthropic Claude | `https://api.anthropic.com` |
| `gemini` | Google Gemini | N/A (uses Google credentials) |
| `together` | Together AI | `https://api.together.xyz/v1` |
| `fireworks` | Fireworks AI | `https://api.fireworks.ai/inference/v1` |
| `groq` | Groq | `https://api.groq.com/openai/v1` |
| `nvidia` | NVIDIA NIM | `https://integrate.api.nvidia.com/v1` |

### Storage Types

| Type | Description |
|------|-------------|
| `sqlite` | SQLite (default, non-persistent) |
| `postgres` | PostgreSQL (requires connectionString) |

---

## Troubleshooting

### Configuration Validation Failed

Check the status condition:

```bash
kubectl get llsd my-llsd -o jsonpath='{.status.conditions[?(@.type=="ConfigReady")]}'
```

Common issues:
- Missing required field (e.g., `endpoint` when `provider` is set)
- Invalid provider name
- Missing `connectionString` for PostgreSQL

### Secret Not Found

Ensure the referenced secret exists in the same namespace:

```bash
kubectl get secret -n llama-stack
```

### ConfigMap Not Created

Check operator logs:

```bash
kubectl logs -l app=llama-stack-operator -n llama-stack-operator-system
```

---

## Reference

- [Full Specification](./spec.md)
- [Data Model](./data-model.md)
- [Research Notes](./research.md)
