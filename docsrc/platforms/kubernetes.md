# Kubernetes Deployment

Deploy AgentKit agents to Kubernetes using Helm charts.

## Supported Distributions

AgentKit works with any Kubernetes distribution:

| Type | Distributions |
|------|---------------|
| **Cloud** | AWS EKS, Google GKE, Azure AKS, DigitalOcean DOKS |
| **Local** | Minikube, kind, k3s, Docker Desktop |
| **On-prem** | Rancher, OpenShift, Tanzu |

The Helm charts and values validation are distribution-agnostic.

## Overview

AgentKit provides:

- **Helm values validation** - Go structs with type-safe validation
- **Reusable templates** - Common deployment patterns
- **Multi-agent support** - Deploy multiple agents in one chart

## Helm Values Validation

Use the `platforms/kubernetes` package to validate your Helm values:

```go
import "github.com/agentplexus/agentkit/platforms/kubernetes"

// Load and validate
values, errs := kubernetes.LoadAndValidate("values.yaml")
if len(errs) > 0 {
    for _, err := range errs {
        log.Printf("Validation error: %v", err)
    }
}

// Merge base and overlay
values, err := kubernetes.LoadAndMerge("values.yaml", "values-prod.yaml")
```

## Example values.yaml

```yaml
global:
  image:
    registry: ghcr.io/myorg
    pullPolicy: IfNotPresent
    tag: "latest"

namespace:
  create: true
  name: my-agents

llm:
  provider: gemini
  geminiModel: "gemini-2.0-flash-exp"

agents:
  research:
    enabled: true
    replicaCount: 2
    image:
      repository: research-agent
    service:
      type: ClusterIP
      port: 8001
      a2aPort: 9001
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
    autoscaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 10
      targetCPUUtilizationPercentage: 70

  synthesis:
    enabled: true
    replicaCount: 1
    image:
      repository: synthesis-agent
    service:
      port: 8002
      a2aPort: 9002

secrets:
  create: true
  geminiApiKey: ""  # Set via --set or external secret manager

vaultguard:
  enabled: true
  minSecurityScore: 50
  requireEncryption: true

ingress:
  enabled: true
  className: nginx
  host: agents.example.com
  tls:
    - secretName: agents-tls
      hosts:
        - agents.example.com
```

## Helm Templates

AgentKit provides reusable templates in `platforms/kubernetes/templates/`:

### _helpers.tpl

Common template functions:

```yaml
# Agent labels
{{- include "agentkit.agentLabels" (dict "context" . "agent" "research") }}

# Image name
{{- include "agentkit.image" (dict "global" .Values.global "agent" .Values.agents.research) }}

# Namespace
{{- include "agentkit.namespace" . }}
```

### deployment.yaml.tpl

Generic deployment template:

```yaml
# In your chart's templates/research-deployment.yaml
{{- include "agentkit.deployment" (dict "agent" .Values.agents.research "name" "research" "values" .) }}
```

## Validation Rules

The Kubernetes package validates:

| Field | Validation |
|-------|------------|
| `namespace.name` | Required, 1-63 chars |
| `llm.provider` | One of: gemini, claude, openai, ollama, xai |
| `agents.*.replicaCount` | 0-100 |
| `agents.*.service.port` | 1-65535, no conflicts |
| `resources.*.cpu` | Valid K8s quantity (e.g., "100m", "1") |
| `resources.*.memory` | Valid K8s quantity (e.g., "128Mi", "1Gi") |
| `ingress.host` | Valid hostname when ingress enabled |

## Port Conflict Detection

```go
values, errs := kubernetes.LoadAndValidate("values.yaml")
// Returns error if two agents use the same port:
// "port conflict: research and synthesis both use port 8001"
```

## Deployment

```bash
# Install
helm install my-agents ./chart -f values.yaml

# Upgrade
helm upgrade my-agents ./chart -f values.yaml -f values-prod.yaml

# With secrets
helm install my-agents ./chart \
  -f values.yaml \
  --set secrets.geminiApiKey=$GEMINI_API_KEY
```

## Security Configuration

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL

vaultguard:
  enabled: true
  minSecurityScore: 50
  requireEncryption: true
  requireIam: true
  deniedNamespaces:
    - kube-system
    - default
```

## Resource Management

```yaml
agents:
  research:
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    autoscaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 20
      targetCPUUtilizationPercentage: 70
      targetMemoryUtilizationPercentage: 80
      behavior:
        scaleDown:
          stabilizationWindowSeconds: 300
        scaleUp:
          stabilizationWindowSeconds: 60
    pdb:
      enabled: true
      minAvailable: "50%"
```

## Next Steps

- [AWS AgentCore](agentcore.md) - Alternative serverless deployment
- [Local Development](../getting-started/local-development.md) - Test before deploying
