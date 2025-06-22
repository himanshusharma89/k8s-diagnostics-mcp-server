# Deployment Guide for K8s Diagnostics MCP Server

## Quick Deploy on Render

### Prerequisites
- GitHub repository with your code
- Render account (free tier available)

### Step 1: Prepare Your Repository
Make sure your repository contains:
- `Dockerfile` (updated with Go 1.24)
- `main.go` (MCP server)
- `http_server.go` (HTTP wrapper)
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
   - `KUBECONFIG`: If you need to connect to a specific cluster
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
  "time": "2024-01-01T12:00:00Z"
}
```

## Available Endpoints

All endpoints accept POST requests with JSON bodies:

- `/health` - Health check
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

### Diagnose a Pod
```bash
curl -X POST https://your-app-name.onrender.com/diagnose_pod \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "default",
    "pod_name": "my-app-pod-123"
  }'
```

### Analyze Cluster Health
```bash
curl -X POST https://your-app-name.onrender.com/analyze_cluster_health \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Lyzr AI Integration

1. **Provide OpenAPI Spec**: Give the updated `openapi-spec.json` to Lyzr AI
2. **Test Endpoints**: Make sure all endpoints work as expected
3. **Monitor Usage**: Check Render dashboard for usage and logs

## Troubleshooting

### Build Fails
- Check that Go version in Dockerfile matches `go.mod` (should be 1.24)
- Ensure all dependencies are in `go.mod`

### Runtime Errors
- Check Render logs for Kubernetes connection issues
- Verify environment variables are set correctly
- Ensure the cluster is accessible from Render's servers

### API Not Responding
- Check if the service is running in Render dashboard
- Verify the port is set to 8080
- Test the `/health` endpoint first

## Cost Considerations

- **Free Tier**: 750 hours/month, auto-sleeps after 15 minutes of inactivity
- **Paid Plans**: Start at $7/month for always-on service
- **Bandwidth**: Free tier includes 100GB/month

## Security Notes

- No authentication implemented (as requested for hackathon)
- Consider adding auth for production use
- Monitor access logs in Render dashboard
- Use HTTPS (automatically provided by Render) 