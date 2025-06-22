.PHONY: setup build build-http test clean cluster-up cluster-down deploy-test-pods

# Setup development environment
setup:
	@echo "Setting up K8s MCP Server development environment..."
	go mod tidy
	kind --version || (echo "Please install kind: https://kind.sigs.k8s.io/docs/user/quick-start/" && exit 1)
	kubectl version --client || (echo "Please install kubectl" && exit 1)

# Build the MCP server (for Claude Desktop)
build:
	@echo "Building K8s MCP Server for Claude Desktop..."
	go build -o bin/k8s-diagnostics-mcp-server .

# Build the HTTP server (for web hosting)
build-http:
	@echo "Building K8s MCP Server for HTTP hosting..."
	HTTP_MODE=true go build -o bin/k8s-diagnostics-mcp-server-http .

# Create Kind cluster
cluster-up:
	@echo "Creating Kind cluster..."
	kind create cluster --config kind-config.yaml --name mcp-test-cluster
	kubectl cluster-info --context kind-mcp-test-cluster

# Deploy test scenarios
deploy-test-pods:
	@echo "Deploying test scenarios..."
	kubectl apply -f test-scenarios/problematic-pods.yaml
	kubectl apply -f test-scenarios/healthy-workloads.yaml
	@echo "Waiting for pods to be scheduled..."
	sleep 30
	kubectl get pods --all-namespaces

# Test all MCP server functions
test: build
	@echo "Testing MCP Server functions..."
	@echo "1. Testing cluster health analysis..."
	echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "analyze_cluster_health", "arguments": {}}}' | ./bin/k8s-diagnostics-mcp-server
	
	@echo "2. Testing problematic pod detection..."
	echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "find_problematic_pods", "arguments": {"criteria": "all"}}}' | ./bin/k8s-diagnostics-mcp-server
	
	@echo "3. Testing pod search..."
	echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "search_pods", "arguments": {"pattern": "elasticsearch"}}}' | ./bin/k8s-diagnostics-mcp-server

# Clean up
clean:
	kind delete cluster --name mcp-test-cluster
	rm -f bin/k8s-diagnostics-mcp-server
	rm -f bin/k8s-diagnostics-mcp-server-http

# Full test cycle
full-test: cluster-up deploy-test-pods test

# Claude Desktop integration test
claude-test: build
	@echo "Testing Claude Desktop integration..."
	@echo "Make sure to add the server config to Claude Desktop first!"
	@echo "Server binary ready at: $(PWD)/bin/k8s-diagnostics-mcp-server"

# HTTP server test
http-test: build-http
	@echo "Testing HTTP server..."
	@echo "Server binary ready at: $(PWD)/bin/k8s-diagnostics-mcp-server-http"
	@echo "Run with: ./bin/k8s-diagnostics-mcp-server-http"