# Kubernetes Diagnostics MCP Server

A Model Context Protocol (MCP) server specifically designed for **intelligent Kubernetes troubleshooting and diagnostics**. Built by InfraCloud to complement existing K8s management tools with advanced debugging capabilities.

## 🚀 Features

### Advanced Diagnostics
- **Pod Health Analysis**: Deep inspection of pod issues with intelligent suggestions
- **Cluster Health Overview**: Comprehensive cluster health assessment
- **Log Analysis**: Automated error pattern detection in pod logs
- **Workload Recommendations**: Best practice suggestions for deployments and services

### Intelligent Troubleshooting
- Identifies common issues: CrashLoopBackOff, ImagePullBackOff, resource constraints
- Provides contextual suggestions based on error patterns
- Analyzes restart counts, resource usage, and configuration issues
- Correlates events with pod problems

### Built for Production
- Works with any Kubernetes cluster (in-cluster or external)
- Supports both kubeconfig and in-cluster authentication
- Comprehensive error handling and logging
- Follows Kubernetes client-go best practices

## 🛠 Installation

### Prerequisites
- Go 1.21+
- Access to a Kubernetes cluster
- kubectl configured or running inside a K8s cluster

### Build from Source
```bash
git clone <this-repo>
cd k8s-diagnostics-mcp
go mod tidy
go build -o k8s-diagnostics-mcp
```

### Configuration
The server automatically detects your Kubernetes configuration:
1. **In-cluster**: Uses service account when running inside K8s
2. **Local**: Uses `~/.kube/config` or `$KUBECONFIG` environment variable

## 🔧 Available Tools

### `diagnose_pod`
Diagnose issues with a specific Kubernetes pod.

**Parameters:**
- `pod_name` (required): Name of the pod to diagnose
- `namespace` (optional): Kubernetes namespace (default: "default")

**Returns:**
- Pod status and phase information
- Container restart counts and ready status
- Identified issues and intelligent suggestions
- Recent events related to the pod
- Resource configuration analysis

### `analyze_cluster_health`
Analyze overall Kubernetes cluster health and identify issues.

**Returns:**
- Node count and health status
- Namespace count
- List of problematic pods with diagnostics
- Resource usage overview
- Cluster-wide recommendations

### `get_workload_recommendations`
Get recommendations for improving workload configurations.

**Parameters:**
- `namespace` (optional): Kubernetes namespace to analyze (default: "default")

**Returns:**
- Best practice recommendations for deployments
- Resource limit suggestions
- High availability recommendations

### `analyze_pod_logs`
Get and analyze pod logs for common error patterns.

**Parameters:**
- `pod_name` (required): Name of the pod
- `namespace` (optional): Kubernetes namespace (default: "default")
- `container` (optional): Specific container name
- `lines` (optional): Number of log lines to retrieve (default: 100)

**Returns:**
- Raw log output
- Detected error patterns
- Contextual suggestions based on errors
- Log analysis summary

## 📚 Resources

### `k8s://troubleshooting/common-issues`
Provides a comprehensive troubleshooting guide for common Kubernetes issues including:
- CrashLoopBackOff debugging
- ImagePullBackOff resolution
- Pending pod issues
- Performance troubleshooting
- Networking problem diagnosis

## 🎯 Use Cases

### For DevOps Engineers
- Quick pod issue diagnosis during incidents
- Cluster health monitoring and alerting
- Automated troubleshooting workflows
- Best practice compliance checking

### For Platform Teams
- Standardized troubleshooting procedures
- Knowledge sharing through AI-assisted debugging
- Proactive issue identification
- Developer self-service debugging

### For AI-Assisted Operations
- Integration with ChatGPT, Claude, or other LLMs
- Natural language troubleshooting queries
- Automated incident response
- Intelligent alert correlation

## 🌟 What Makes This Unique

Unlike existing Kubernetes MCP servers that focus on basic cluster management, this server specializes in:

1. **Intelligent Issue Detection**: Goes beyond simple status checks to identify root causes
2. **Contextual Suggestions**: Provides actionable recommendations based on specific error patterns
3. **Log Analysis**: Automatically parses and analyzes logs for common issues
4. **Production-Ready**: Built with enterprise Kubernetes environments in mind
5. **CNCF Ecosystem Integration**: Designed to work with other CNCF tools and best practices

## 🤝 Contributing

We welcome contributions! This project aligns with InfraCloud's commitment to the CNCF ecosystem and open-source community.

### Development
```bash
# Run locally
go run main.go

# Run tests
go test ./...

# Build for production
go build -ldflags="-s -w" -o k8s-diagnostics-mcp
```

## 📞 Support

- **InfraCloud Blog**: Check our [technical blog](https://www.infracloud.io/blogs/) for K8s best practices
- **GitHub Issues**: Report bugs and feature requests
- **CNCF Community**: Engage with the broader Kubernetes community

## 🏷 License

[TODO]

---

**Built with ❤️ by [InfraCloud](https://www.infracloud.io) for the CNCF community**