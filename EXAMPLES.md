# Kubernetes Diagnostics MCP Server - Usage Examples

This document provides practical examples of how to use the Kubernetes Diagnostics MCP Server for common troubleshooting scenarios.

## 🔍 Basic Diagnostics

### Diagnosing a Problematic Pod

```bash
# Example: Diagnose a pod that's in CrashLoopBackOff
mcp-client call diagnose_pod '{"pod_name": "nginx-deployment-67d4f7d4c8-xyz", "namespace": "production"}'
```

**Sample Response:**
```json
{
  "name": "nginx-deployment-67d4f7d4c8-xyz",
  "namespace": "production",
  "status": "Running",
  "restart_count": 15,
  "issues": [
    "Container nginx has high restart count: 15",
    "Container nginx is waiting: CrashLoopBackOff"
  ],
  "suggestions": [
    "Check container logs and resource limits",
    "Check application logs and startup configuration"
  ],
  "recent_events": [
    "BackOff: Back-off restarting failed container",
    "Killing: Container nginx failed liveness probe"
  ],
  "resources": {}
}
```

### Analyzing Cluster Health

```bash
# Get overall cluster health assessment
mcp-client call analyze_cluster_health '{}'
```

**Sample Response:**
```json
{
  "node_count": 5,
  "healthy_nodes": 4,
  "namespace_count": 12,
  "pod_issues": [
    {
      "name": "redis-master-0",
      "namespace": "database",
      "status": "Pending",
      "issues": ["Pod has been pending for over 5 minutes"],
      "suggestions": ["Check node resources and scheduling constraints"]
    }
  ],
  "resource_usage": {},
  "recommendations": [
    "Consider investigating unhealthy nodes",
    "Review pending pods for resource constraints"
  ]
}
```

## 🚨 Incident Response Scenarios

### Scenario 1: Application Not Starting

**Problem**: New deployment fails to start, pods in ImagePullBackOff

```bash
# Diagnose the failing pod
mcp-client call diagnose_pod '{"pod_name": "myapp-deployment-abc123", "namespace": "staging"}'

# Analyze logs for more context
mcp-client call analyze_pod_logs '{"pod_name": "myapp-deployment-abc123", "namespace": "staging"}'
```

**Expected Issues Detected:**
- ImagePullBackOff status
- Suggestions to check image name and registry credentials
- Event history showing pull failures

### Scenario 2: High Memory Usage

**Problem**: Application consuming too much memory, getting OOMKilled

```bash
# Get workload recommendations
mcp-client call get_workload_recommendations '{"namespace": "production"}'

# Analyze specific pod
mcp-client call diagnose_pod '{"pod_name": "memory-hungry-app-xyz", "namespace": "production"}'

# Check logs for memory-related errors
mcp-client call analyze_pod_logs '{"pod_name": "memory-hungry-app-xyz", "namespace": "production", "lines": 200}'
```

**Expected Recommendations:**
- Set appropriate memory limits
- Review memory usage patterns
- Consider horizontal pod autoscaling

### Scenario 3: Network Connectivity Issues

**Problem**: Service-to-service communication failing

```bash
# Diagnose all pods in the namespace
mcp-client call analyze_cluster_health '{}'

# Check specific service pods
mcp-client call diagnose_pod '{"pod_name": "frontend-service-123", "namespace": "default"}'
mcp-client call diagnose_pod '{"pod_name": "backend-service-456", "namespace": "default"}'

# Analyze logs for network errors
mcp-client call analyze_pod_logs '{"pod_name": "frontend-service-123", "namespace": "default"}'
```

## 🤖 AI-Assisted Troubleshooting

### Integration with ChatGPT/Claude

Here's how you might use this MCP server with an AI assistant:

**Prompt Example:**
```
"I have a Kubernetes pod that keeps restarting. Can you help me diagnose it? 
The pod name is 'web-server-deployment-abc123' in the 'production' namespace."
```

**AI Response using MCP:**
The AI would use the `diagnose_pod` tool and then provide human-readable analysis:

"I've analyzed your pod and found several issues:

1. **High Restart Count**: The pod has restarted 15 times, indicating a persistent problem
2. **CrashLoopBackOff**: The container is failing to start successfully
3. **Recent Events**: The logs show liveness probe failures

**Immediate Actions:**
- Check your application's startup process
- Review resource limits (CPU/Memory)
- Examine application logs for startup errors
- Verify environment variables and configuration

Let me also analyze the logs for specific error patterns..."

### Advanced Troubleshooting Workflow

```bash
# 1. Get cluster overview
mcp-client call analyze_cluster_health '{}'

# 2. Focus on problematic namespace
mcp-client call get_workload_recommendations '{"namespace": "problematic-ns"}'

# 3. Deep dive into specific pods
mcp-client call diagnose_pod '{"pod_name": "failing-pod", "namespace": "problematic-ns"}'

# 4. Analyze logs with context
mcp-client call analyze_pod_logs '{"pod_name": "failing-pod", "namespace": "problematic-ns", "lines": 500}'

# 5. Get troubleshooting guide
mcp-client get-resource "k8s://troubleshooting/common-issues"
```

## 📊 Production Use Cases

### Daily Health Checks

Create a script that runs daily health checks:

```bash
#!/bin/bash
echo "=== Daily Kubernetes Health Check ==="
echo "Timestamp: $(date)"

# Overall cluster health
echo "## Cluster Health"
mcp-client call analyze_cluster_health '{}'

# Check critical namespaces
for ns in production staging database monitoring; do
    echo "## Namespace: $ns"
    mcp-client call get_workload_recommendations "{\"namespace\": \"$ns\"}"
done
```

### Automated Incident Detection

Use the MCP server in monitoring pipelines:

```yaml
# Example Prometheus AlertManager integration
groups:
- name: k8s-diagnostics
  rules:
  - alert: PodCrashLooping
    expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
    for: 5m
    annotations:
      summary: "Pod {{ $labels.pod }} is crash looping"
      description: "Use MCP diagnostics: diagnose_pod with pod={{ $labels.pod }} namespace={{ $labels.namespace }}"
```

### Self-Service Debugging

Enable developers to debug their own applications:

```bash
# Developer toolkit script
#!/bin/bash
POD_NAME=$1
NAMESPACE=${2:-default}

echo "🔍 Diagnosing pod: $POD_NAME in namespace: $NAMESPACE"

# Quick diagnosis
mcp-client call diagnose_pod "{\"pod_name\": \"$POD_NAME\", \"namespace\": \"$NAMESPACE\"}"

echo "📝 Analyzing recent logs..."
mcp-client call analyze_pod_logs "{\"pod_name\": \"$POD_NAME\", \"namespace\": \"$NAMESPACE\", \"lines\": 100}"

echo "💡 Getting troubleshooting guide..."
mcp-client get-resource "k8s://troubleshooting/common-issues"
```

## 🎯 Best Practices

### 1. Regular Health Assessments
- Run `analyze_cluster_health` daily during maintenance windows
- Set up alerts based on cluster health metrics
- Track trends in pod restart counts and failure patterns

### 2. Proactive Issue Detection
- Use `get_workload_recommendations` during deployment reviews
- Implement resource limit checks in CI/CD pipelines
- Regular audit of security and configuration best practices

### 3. Incident Response Integration
- Include MCP diagnostics in runbooks
- Train teams on using diagnostic tools effectively
- Create templates for common troubleshooting scenarios

### 4. Documentation and Knowledge Sharing
- Use AI assistants with MCP integration for team training
- Create searchable knowledge base from diagnostic results
- Share common patterns and solutions across teams

## 🔧 Advanced Configuration

### Custom Error Patterns

Extend the log analysis with custom patterns:

```go
// Add to your custom build
customPatterns := []string{
    "database connection failed",
    "authentication error",
    "rate limit exceeded",
}
```

### Integration with Observability Stack

Combine with Prometheus, Grafana, and Jaeger:

```bash
# Get pod metrics context
kubectl top pod $POD_NAME -n $NAMESPACE

# Run MCP diagnostics
mcp-client call diagnose_pod "{\"pod_name\": \"$POD_NAME\", \"namespace\": \"$NAMESPACE\"}"

# Check traces in Jaeger
# (integrate with tracing data)
```

---

This MCP server transforms Kubernetes troubleshooting from reactive debugging to proactive, AI-assisted operations management.