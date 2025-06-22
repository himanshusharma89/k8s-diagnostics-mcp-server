#!/bin/bash

echo "🚀 Setting up K8s Diagnostics Demo Environment"

# Start minikube if not running
if ! minikube status | grep -q "Running"; then
    echo "Starting minikube..."
    minikube start
else
    echo "Minikube is already running"
fi

# Create demo namespace
echo "Creating demo namespace..."
kubectl create namespace demo --dry-run=client -o yaml | kubectl apply -f -

# Deploy some healthy pods
echo "Deploying healthy demo pods..."
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: healthy-app
  namespace: demo
spec:
  replicas: 2
  selector:
    matchLabels:
      app: healthy-app
  template:
    metadata:
      labels:
        app: healthy-app
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
EOF

# Deploy a pod that will fail (wrong image)
echo "Deploying problematic pods for demo..."
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: failing-app
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: failing-app
  template:
    metadata:
      labels:
        app: failing-app
    spec:
      containers:
      - name: app
        image: nonexistent-image:latest
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
EOF

# Deploy a pod with resource issues
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: resource-heavy-app
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: resource-heavy-app
  template:
    metadata:
      labels:
        app: resource-heavy-app
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
        # No resource limits - will cause issues
EOF

# Deploy a pod that will restart frequently
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: restarting-app
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: restarting-app
  template:
    metadata:
      labels:
        app: restarting-app
    spec:
      containers:
      - name: app
        image: busybox
        command: ["/bin/sh"]
        args: ["-c", "echo 'Starting...'; sleep 10; exit 1"]
        resources:
          requests:
            memory: "32Mi"
            cpu: "25m"
          limits:
            memory: "64Mi"
            cpu: "50m"
EOF

echo "Waiting for pods to be scheduled..."
sleep 30

echo "📊 Current pod status:"
kubectl get pods -n demo

echo "🔍 Demo setup complete!"
echo ""
echo "You now have:"
echo "✅ 2 healthy nginx pods"
echo "❌ 1 failing pod (wrong image)"
echo "⚠️  1 pod without resource limits"
echo "🔄 1 pod that restarts frequently"
echo ""
echo "To run the demo server locally:"
echo "1. make build-http"
echo "2. DEMO_MODE=false ./bin/k8s-diagnostics-mcp-server-http"
echo ""
echo "To test endpoints:"
echo "curl -X POST http://localhost:8080/analyze_cluster_health"
echo "curl -X POST http://localhost:8080/find_problematic_pods" 