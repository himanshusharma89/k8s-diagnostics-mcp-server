# Makefile - Build and test automation
.PHONY: build test setup-kind deploy-test-pods clean

# Build the K8s diagnostics server
build:
	go mod tidy
	go build -o bin/k8s-diagnostics-server main.go

# Set up kind cluster for testing
setup-kind:
	kind create cluster --name k8s-diagnostics --config kind-config.yaml
	kubectl cluster-info --context kind-k8s-diagnostics

# Deploy test pods with various issues
deploy-test-pods:
	kubectl apply -f test-manifests/

# Test the MCP server locally
test-local: build
	echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./bin/k8s-diagnostics-server

# Clean up
clean:
	kind delete cluster --name k8s-diagnostics
	docker-compose down -v

# Full test setup
test-setup: setup-kind deploy-test-pods build
