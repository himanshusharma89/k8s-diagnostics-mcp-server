# Deployment Guide for K8s Diagnostics MCP Server

## Quick Deploy on Render

### Prerequisites
- GitHub repository with your code
- Render account (free tier available)

### Step 1: Prepare Your Repository
Make sure your repository contains:
- `Dockerfile` (updated with Go 1.24)
- `main.go` (MCP server)
- `http_server.go` (HTTP wrapper with demo mode)
- `go.mod` and `go.sum`
- `openapi-spec.json`

### Step 2: Deploy on Render

1. **Go to Render Dashboard**
   - Visit [https://dashboard.render.com/](https://dashboard.render.com/)
   - Sign up or log in

2. **Create New Web Service**
   - Click **"New +"** → **"Web Service"**
   - Connect your GitHub repository
   - Select your repository

3. **Configure the Service**
   - **Name**: `k8s-diagnostics-mcp-server` (or your preferred name)
   - **Environment**: `Docker`
   - **Region**: Choose closest to you
   - **Branch**: `main` (or your default branch)
   - **Build Command**: Leave empty (uses Dockerfile)
   - **Start Command**: Leave empty (uses Dockerfile ENTRYPOINT)
   - **Port**: `8080`

4. **Environment Variables** (Optional)
   - `HTTP_MODE`: `true` (already set in Dockerfile)
   - `DEMO_MODE`: `true` (already set in Dockerfile for cloud deployment)
   - `KUBECONFIG`: Only needed for real cluster access (not for demo mode)
   - `PORT`: `8080` (Render will set this automatically)

5. **Deploy**
   - Click **"Create Web Service"**
   - Wait for build and deployment (usually 2-5 minutes)

### Step 3: Get Your URL
Once deployed, you'll get a URL like:
```
https://your-app-name.onrender.com
```

### Step 4: Update OpenAPI Spec
Update your `openapi-spec.json` file:
```json
{
  "servers": [
    {
      "url": "https://your-app-name.onrender.com",
      "description": "Production server"
    }
  ]
}
```

### Step 5: Test Your API
Test the health endpoint:
```bash
curl https://your-app-name.onrender.com/health
```

You should get:
```json
{
  "status": "healthy",
  "time": "2024-01-01T12:00:00Z",
  "demo_mode": true
}
```

## Demo Mode vs Real Mode

### 🌐 **Demo Mode (Cloud Deployment)**
- **Purpose**: For Lyzr AI integration and demos
- **Data**: Uses realistic mock data
- **No Kubernetes**: Doesn't require cluster access
- **Environment**: `DEMO_MODE=true`

### 🖥️ **Real Mode (Local Development)**
- **Purpose**: For actual Kubernetes diagnostics
- **Data**: Real cluster data
- **Requires**: Kubernetes cluster access
- **Environment**: `DEMO_MODE=false` (default)

## Available Endpoints

All endpoints accept POST requests with JSON bodies:

- `/health` - Health check (shows demo mode status)
- `/diagnose_pod` - Diagnose a specific pod
- `/analyze_cluster_health` - Analyze cluster health
- `/analyze_pod_logs` - Analyze pod logs
- `/list_pods` - List pods in namespace
- `/find_problematic_pods` - Find problematic pods
- `/get_resource_usage` - Get resource usage
- `/quick_triage` - Quick cluster triage
- `/get_workload_recommendations` - Get recommendations
- `/search_pods` - Search pods by pattern

## Example Usage

### Diagnose a Pod (Demo Mode)
```bash
curl -X POST https://your-app-name.onrender.com/diagnose_pod \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "default",
    "pod_name": "my-app-pod-123"
  }'
```

**Response (Demo Mode):**
```json
{
  "name": "my-app-pod-123",
  "namespace": "default",
  "status": "Running",
  "restart_count": 2,
  "issues": [
    "Container has restarted 2 times in the last hour",
    "Memory usage is at 85% of limit"
  ],
  "suggestions": [
    "Check application logs for errors",
    "Consider increasing memory limits",
    "Review resource requests"
  ],
  "recent_events": [
    "Pod scheduled successfully",
    "Container started",
    "Container restarted due to OOM"
  ],
  "resources": {
    "app-container_cpu_request": "100m",
    "app-container_memory_request": "128Mi",
    "app-container_cpu_limit": "200m",
    "app-container_memory_limit": "256Mi"
  },
  "created_at": "2024-01-01T12:00:00Z"
}
```

### Analyze Cluster Health (Demo Mode)
```bash
curl -X POST https://your-app-name.onrender.com/analyze_cluster_health \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Local Development Setup

### For Real Kubernetes Access
```bash
# Build for real cluster access
make build

# Run with real cluster
./bin/k8s-diagnostics-mcp-server

# Or set environment explicitly
DEMO_MODE=false ./bin/k8s-diagnostics-mcp-server
```

### For Demo Mode Testing
```bash
# Build for demo mode
make build-http

# Run in demo mode
DEMO_MODE=true ./bin/k8s-diagnostics-mcp-server-http

# Test locally
curl -X POST http://localhost:8080/health
```

## Lyzr AI Integration

1. **Provide OpenAPI Spec**: Give the updated `openapi-spec.json` to Lyzr AI
2. **Demo Mode**: The server will return realistic mock data
3. **Test Endpoints**: All endpoints work with mock data
4. **Monitor Usage**: Check Render dashboard for usage and logs

## Troubleshooting

### Build Fails
- Check that Go version in Dockerfile matches `go.mod` (should be 1.24)
- Ensure all dependencies are in `go.mod`

### Runtime Errors
- **Demo Mode**: Should work without any Kubernetes cluster
- **Real Mode**: Check Kubernetes cluster connectivity
- Verify environment variables are set correctly

### API Not Responding
- Check if the service is running in Render dashboard
- Verify the port is set to 8080
- Test the `/health` endpoint first
- Check if `DEMO_MODE` is set correctly

## Cost Considerations

- **Free Tier**: 750 hours/month, auto-sleeps after 15 minutes of inactivity
- **Paid Plans**: Start at $7/month for always-on service
- **Bandwidth**: Free tier includes 100GB/month

## Security Notes

- No authentication implemented (as requested for hackathon)
- Demo mode doesn't require any cluster access
- Consider adding auth for production use
- Monitor access logs in Render dashboard
- Use HTTPS (automatically provided by Render)

## Future Enhancements

### For Production Deployment
1. **Add Authentication**: Implement API keys or OAuth
2. **Real Cluster Access**: Configure for actual Kubernetes clusters
3. **Database Integration**: Store diagnostic history
4. **Monitoring**: Add metrics and alerting
5. **Auto-mitigation**: Implement automatic issue resolution

### For Demo Mode
1. **More Realistic Data**: Expand mock data scenarios
2. **Interactive Demos**: Add guided troubleshooting flows
3. **Scenario Builder**: Create custom demo scenarios 