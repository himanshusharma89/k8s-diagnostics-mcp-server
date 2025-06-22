# Dual Mode Setup: MCP Server + HTTP Server

This K8s Diagnostics MCP Server now supports two modes:

## 🖥️ **MCP Mode (Claude Desktop)**
- **Purpose**: Run as a local MCP server for Claude Desktop
- **Protocol**: Uses stdio for communication
- **Binary**: `bin/k8s-diagnostics-mcp-server`
- **Usage**: Direct integration with Claude Desktop

## 🌐 **HTTP Mode (Web Hosting)**
- **Purpose**: Run as a web API server for Lyzr AI and other HTTP clients
- **Protocol**: HTTP REST API
- **Binary**: `bin/k8s-diagnostics-mcp-server-http` (or same binary with env var)
- **Usage**: Deployed on Render/Railway for web access

---

## How It Works

The server automatically switches modes based on the `HTTP_MODE` environment variable:

```go
func main() {
    if os.Getenv("HTTP_MODE") == "true" {
        runHTTPServer()  // Web API mode
    } else {
        runMCPServer()   // Claude Desktop mode
    }
}
```

---

## Building and Running

### For Claude Desktop (MCP Mode)
```bash
# Build for Claude Desktop
make build

# Binary location
./bin/k8s-diagnostics-mcp-server

# Add to Claude Desktop config
{
  "mcpServers": {
    "k8s-diagnostics": {
      "command": "/path/to/bin/k8s-diagnostics-mcp-server",
      "args": []
    }
  }
}
```

### For Web Hosting (HTTP Mode)
```bash
# Build for HTTP hosting
make build-http

# Run locally for testing
./bin/k8s-diagnostics-mcp-server-http

# Or set environment variable
HTTP_MODE=true ./bin/k8s-diagnostics-mcp-server
```

---

## Available Endpoints (HTTP Mode)

All endpoints accept POST requests with JSON:

| Endpoint | Description | Example Request |
|----------|-------------|-----------------|
| `/health` | Health check | `{}` |
| `/diagnose_pod` | Diagnose specific pod | `{"namespace": "default", "pod_name": "my-pod"}` |
| `/analyze_cluster_health` | Cluster health analysis | `{}` |
| `/analyze_pod_logs` | Analyze pod logs | `{"pod_name": "my-pod", "lines": 100}` |
| `/list_pods` | List pods in namespace | `{"namespace": "default"}` |
| `/find_problematic_pods` | Find problematic pods | `{"criteria": "all"}` |
| `/get_resource_usage` | Resource usage analysis | `{"sort_by": "restarts"}` |
| `/quick_triage` | Quick cluster triage | `{}` |
| `/get_workload_recommendations` | Get recommendations | `{"namespace": "default"}` |
| `/search_pods` | Search pods by pattern | `{"pattern": "my-app"}` |

---

## Deployment Options

### 1. Local Development
```bash
# MCP mode (Claude Desktop)
make build
./bin/k8s-diagnostics-mcp-server

# HTTP mode (local testing)
make build-http
./bin/k8s-diagnostics-mcp-server-http
```

### 2. Web Hosting (Render/Railway)
- Use the `Dockerfile` (automatically sets `HTTP_MODE=true`)
- Deploy to Render/Railway
- Update `openapi-spec.json` with your URL
- Provide OpenAPI spec to Lyzr AI

### 3. Docker Local Testing
```bash
# Build Docker image
docker build -t k8s-diagnostics-mcp .

# Run in HTTP mode
docker run -p 8080:8080 k8s-diagnostics-mcp

# Run in MCP mode
docker run -e HTTP_MODE=false k8s-diagnostics-mcp
```

---

## Testing

### Test MCP Mode
```bash
make test
```

### Test HTTP Mode
```bash
make build-http
./bin/k8s-diagnostics-mcp-server-http &
curl -X POST http://localhost:8080/health
```

---

## Configuration Files

### Claude Desktop Config
```json
{
  "mcpServers": {
    "k8s-diagnostics": {
      "command": "/path/to/k8s-diagnostics-mcp-server/bin/k8s-diagnostics-mcp-server",
      "args": []
    }
  }
}
```

### OpenAPI Spec (for Lyzr AI)
- Use `openapi-spec.json`
- Update the `servers` section with your deployment URL
- All endpoints are documented with request/response schemas

---

## Benefits of This Setup

1. **Single Codebase**: Same logic for both modes
2. **Flexible Deployment**: Local MCP or web API
3. **Easy Testing**: Can test both modes locally
4. **Production Ready**: HTTP mode for web hosting
5. **Development Friendly**: MCP mode for local development

---

## Troubleshooting

### MCP Mode Issues
- Check Claude Desktop config path
- Ensure binary has execute permissions
- Verify Kubernetes cluster access

### HTTP Mode Issues
- Check if `HTTP_MODE=true` is set
- Verify port 8080 is available
- Check Kubernetes cluster connectivity
- Review Render/Railway logs

### Common Issues
- **Go version mismatch**: Ensure Dockerfile uses Go 1.24
- **Kubernetes access**: Verify cluster connectivity
- **Port conflicts**: Check if port 8080 is in use 