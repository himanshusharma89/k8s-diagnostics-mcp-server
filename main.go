package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8sDiagnosticsServer struct {
	clientset *kubernetes.Clientset
}

type PodDiagnostic struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Status       string            `json:"status"`
	RestartCount int32             `json:"restart_count"`
	Issues       []string          `json:"issues"`
	Suggestions  []string          `json:"suggestions"`
	Events       []string          `json:"recent_events"`
	Resources    map[string]string `json:"resources"`
}

type ClusterHealth struct {
	NodeCount       int                    `json:"node_count"`
	HealthyNodes    int                    `json:"healthy_nodes"`
	NamespaceCount  int                    `json:"namespace_count"`
	PodIssues       []PodDiagnostic        `json:"pod_issues"`
	ResourceUsage   map[string]interface{} `json:"resource_usage"`
	Recommendations []string               `json:"recommendations"`
}

func NewK8sDiagnosticsServer() (*K8sDiagnosticsServer, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
			kubeconfig = kubeconfigEnv
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &K8sDiagnosticsServer{clientset: clientset}, nil
}

func (s *K8sDiagnosticsServer) diagnosePod(ctx context.Context, namespace, podName string) (*PodDiagnostic, error) {
	pod, err := s.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	diagnostic := &PodDiagnostic{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		Issues:    []string{},
		Suggestions: []string{},
		Resources: make(map[string]string),
	}

	// Analyze container statuses
	for _, containerStatus := range pod.Status.ContainerStatuses {
		diagnostic.RestartCount += containerStatus.RestartCount
		
		if containerStatus.RestartCount > 5 {
			diagnostic.Issues = append(diagnostic.Issues, 
				fmt.Sprintf("Container %s has high restart count: %d", 
					containerStatus.Name, containerStatus.RestartCount))
			diagnostic.Suggestions = append(diagnostic.Suggestions,
				"Check container logs and resource limits")
		}

		if !containerStatus.Ready {
			diagnostic.Issues = append(diagnostic.Issues,
				fmt.Sprintf("Container %s is not ready", containerStatus.Name))
		}

		// Check waiting state
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			diagnostic.Issues = append(diagnostic.Issues,
				fmt.Sprintf("Container %s is waiting: %s", containerStatus.Name, reason))
			
			switch reason {
			case "ImagePullBackOff", "ErrImagePull":
				diagnostic.Suggestions = append(diagnostic.Suggestions,
					"Check image name, registry credentials, and network connectivity")
			case "CrashLoopBackOff":
				diagnostic.Suggestions = append(diagnostic.Suggestions,
					"Check application logs and startup configuration")
			}
		}
	}

	// Check resource requests/limits
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests == nil && container.Resources.Limits == nil {
			diagnostic.Issues = append(diagnostic.Issues,
				fmt.Sprintf("Container %s has no resource requests/limits", container.Name))
			diagnostic.Suggestions = append(diagnostic.Suggestions,
				"Set appropriate resource requests and limits")
		}
	}

	// Get recent events
	events, err := s.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", podName),
	})
	if err == nil {
		for _, event := range events.Items {
			if time.Since(event.LastTimestamp.Time) < time.Hour*24 {
				diagnostic.Events = append(diagnostic.Events,
					fmt.Sprintf("%s: %s", event.Reason, event.Message))
			}
		}
	}

	return diagnostic, nil
}

func (s *K8sDiagnosticsServer) analyzeClusterHealth(ctx context.Context) (*ClusterHealth, error) {
	health := &ClusterHealth{
		PodIssues:       []PodDiagnostic{},
		ResourceUsage:   make(map[string]interface{}),
		Recommendations: []string{},
	}

	// Get node information
	nodes, err := s.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	health.NodeCount = len(nodes.Items)
	for _, node := range nodes.Items {
		ready := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				ready = true
				break
			}
		}
		if ready {
			health.HealthyNodes++
		}
	}

	// Get namespace count
	namespaces, err := s.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err == nil {
		health.NamespaceCount = len(namespaces.Items)
	}

	// Find problematic pods across all namespaces
	pods, err := s.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	problemPods := 0
	for _, pod := range pods.Items {
		hasIssues := false
		
		// Skip completed jobs
		if pod.Status.Phase == "Succeeded" {
			continue
		}

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.RestartCount > 3 || !containerStatus.Ready {
				hasIssues = true
				break
			}
		}

		if pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded" {
			hasIssues = true
		}

		if hasIssues {
			diagnostic, err := s.diagnosePod(ctx, pod.Namespace, pod.Name)
			if err == nil {
				health.PodIssues = append(health.PodIssues, *diagnostic)
				problemPods++
			}
		}
	}

	// Generate recommendations
	if float64(health.HealthyNodes)/float64(health.NodeCount) < 0.8 {
		health.Recommendations = append(health.Recommendations,
			"Consider investigating unhealthy nodes")
	}
	
	if problemPods > 10 {
		health.Recommendations = append(health.Recommendations,
			"High number of problematic pods detected - investigate cluster resource constraints")
	}

	return health, nil
}

func (s *K8sDiagnosticsServer) getWorkloadRecommendations(ctx context.Context, namespace string) ([]string, error) {
	recommendations := []string{}

	// Check deployments without resource limits
	deployments, err := s.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, deployment := range deployments.Items {
		hasLimits := false
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Resources.Limits != nil {
				hasLimits = true
				break
			}
		}
		if !hasLimits {
			recommendations = append(recommendations,
				fmt.Sprintf("Deployment %s/%s should have resource limits", 
					deployment.Namespace, deployment.Name))
		}

		// Check replica count
		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 1 {
			recommendations = append(recommendations,
				fmt.Sprintf("Deployment %s/%s has only 1 replica - consider scaling for HA", 
					deployment.Namespace, deployment.Name))
		}
	}

	return recommendations, nil
}

func main() {
	s := server.NewMCPServer(
		"K8s Diagnostics MCP",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	diagnostics, err := NewK8sDiagnosticsServer()
	if err != nil {
		log.Fatalf("Failed to create diagnostics server: %v", err)
	}

	// Tool 1: Diagnose Pod
	diagnosePodTool := mcp.NewTool("diagnose_pod",
		mcp.WithDescription("Diagnose issues with a specific Kubernetes pod"),
		mcp.WithString("namespace",
			mcp.Description("Namespace of the pod"),
		),
		mcp.WithString("pod_name",
			mcp.Required(),
			mcp.Description("Name of the pod"),
		),
	)

	s.AddTool(diagnosePodTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := req.GetString("namespace", "default")
		podName, err := req.RequireString("pod_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result, err := diagnostics.diagnosePod(ctx, namespace, podName)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to diagnose pod", err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("%+v", result)), nil
	})

	// Tool: Analyze cluster health
	analyzeClusterTool := mcp.NewTool("analyze_cluster_health",
		mcp.WithDescription("Analyze overall cluster health and identify issues"),
	)

	s.AddTool(analyzeClusterTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := diagnostics.analyzeClusterHealth(ctx)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("cluster health analysis failed", err), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", result)), nil
	})

	// Tool: Get workload recommendations
	recommendTool := mcp.NewTool("get_workload_recommendations",
		mcp.WithDescription("Get optimization recommendations for workloads"),
		mcp.WithString("namespace",
			mcp.Description("Namespace to scan"),
		),
	)

	s.AddTool(recommendTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := req.GetString("namespace", "default")
		result, err := diagnostics.getWorkloadRecommendations(ctx, namespace)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("recommendation generation failed", err), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("%+v", result)), nil
	})

	// Tool: Get pod logs with error analysis
	analyzePodLogsTool := mcp.NewTool("analyze_pod_logs",
		mcp.WithDescription("Get and analyze pod logs for common error patterns"),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
		),
		mcp.WithString("pod_name",
			mcp.Required(),
			mcp.Description("Name of the pod"),
		),
		mcp.WithString("container",
			mcp.Description("Container name (optional)"),
		),
		mcp.WithNumber("lines",
			mcp.Description("Number of log lines to retrieve"),
		),
	)

	s.AddTool(analyzePodLogsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := req.GetString("namespace", "default")
		podName, err := req.RequireString("pod_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		container := req.GetString("container", "")
		lines := int64(100)
		if l := req.GetInt("lines", 100); l > 0 {
			lines = int64(l)
		}

		logOptions := &corev1.PodLogOptions{
			TailLines: &lines,
		}
		if container != "" {
			logOptions.Container = container
		}

		// Assume server.clientset is accessible here; you need to inject or define it accordingly
		logs, err := diagnostics.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions).Do(ctx).Raw()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get logs: %v", err)), nil
		}

		logText := string(logs)

		// Analyze logs for common error patterns
		errorPatterns := []string{
			"error", "fatal", "exception", "panic", "failed", "timeout",
			"connection refused", "permission denied", "out of memory",
		}

		foundErrors := []string{}
		suggestions := []string{}

		logLines := strings.Split(logText, "\n")
		for _, line := range logLines {
			for _, pattern := range errorPatterns {
				if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
					foundErrors = append(foundErrors, line)

					switch pattern {
					case "out of memory":
						suggestions = append(suggestions, "Consider increasing memory limits or optimizing application memory usage")
					case "connection refused":
						suggestions = append(suggestions, "Check network policies, service configurations, and target service availability")
					case "permission denied":
						suggestions = append(suggestions, "Review RBAC permissions and file system permissions")
					case "timeout":
						suggestions = append(suggestions, "Check network connectivity and increase timeout values if appropriate")
					}
					break
				}
			}
		}

		result := map[string]interface{}{
			"pod_name":     podName,
			"namespace":    namespace,
			"log_lines":    len(logLines),
			"errors_found": foundErrors,
			"suggestions":  suggestions,
			"raw_logs":     logText,
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to serialize result to JSON"), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// Add resource to provide troubleshooting guides
	troubleshootingGuide := `# Kubernetes Troubleshooting Guide

	## Common Pod Issues

	### CrashLoopBackOff
	- **Cause**: Container keeps crashing after startup
	- **Debug**: Check logs, resource limits, startup probes
	- **Solutions**: Fix application code, adjust resource limits, review configuration

	### ImagePullBackOff
	- **Cause**: Cannot pull container image
	- **Debug**: Check image name, registry access, network connectivity
	- **Solutions**: Verify image exists, check registry credentials, review network policies

	### Pending Pods
	- **Cause**: Pod cannot be scheduled
	- **Debug**: Check node resources, taints/tolerations, node selectors
	- **Solutions**: Scale cluster, adjust resource requests, fix scheduling constraints

	## Performance Issues

	### High Memory Usage
	- Monitor with: kubectl top pods
	- Set appropriate memory limits
	- Use memory profiling tools

	### High CPU Usage  
	- Check CPU limits and requests
	- Profile application performance
	- Consider horizontal pod autoscaling

	## Networking Issues

	### Service Discovery
	- Verify service exists and has endpoints
	- Check DNS resolution
	- Review network policies

	### Inter-pod Communication
	- Test connectivity with kubectl exec
	- Check security contexts
	- Review firewall rules
	`

	guideResource := mcp.NewResource("k8s://troubleshooting/common-issues", "Common Kubernetes troubleshooting guide", mcp.WithMIMEType("text/markdown"))
	s.AddResource(guideResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
    return []mcp.ResourceContents{
        mcp.TextResourceContents{
            URI:      request.Params.URI,
            MIMEType: "text/markdown",
            Text:     troubleshootingGuide,
        },
    }, nil
})

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}