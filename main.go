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
	CreatedAt    time.Time         `json:"created_at"`
}

type ClusterHealth struct {
	NodeCount       int                    `json:"node_count"`
	HealthyNodes    int                    `json:"healthy_nodes"`
	NamespaceCount  int                    `json:"namespace_count"`
	PodIssues       []PodDiagnostic        `json:"pod_issues"`
	ResourceUsage   map[string]interface{} `json:"resource_usage"`
	Recommendations []string               `json:"recommendations"`
	Timestamp       time.Time              `json:"timestamp"`
}

type LogAnalysis struct {
	PodName      string   `json:"pod_name"`
	Namespace    string   `json:"namespace"`
	LogLines     int      `json:"log_lines"`
	ErrorsFound  []string `json:"errors_found"`
	Suggestions  []string `json:"suggestions"`
	ErrorCount   int      `json:"error_count"`
	WarningCount int      `json:"warning_count"`
}

// PodResourceInfo holds resource usage and status info for a pod
type PodResourceInfo struct {
	Name            string  `json:"name"`
	Namespace       string  `json:"namespace"`
	CPURequest      string  `json:"cpu_request"`
	MemoryRequest   string  `json:"memory_request"`
	CPULimit        string  `json:"cpu_limit"`
	MemoryLimit     string  `json:"memory_limit"`
	RestartCount    int32   `json:"restart_count"`
	Status          string  `json:"status"`
	HasResourceIssues bool  `json:"has_resource_issues"`
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
		Name:        pod.Name,
		Namespace:   pod.Namespace,
		Status:      string(pod.Status.Phase),
		Issues:      []string{},
		Suggestions: []string{},
		Resources:   make(map[string]string),
		CreatedAt:   time.Now(),
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

		// Store resource information
		if container.Resources.Requests != nil {
			if cpu := container.Resources.Requests.Cpu(); cpu != nil {
				diagnostic.Resources[container.Name+"_cpu_request"] = cpu.String()
			}
			if memory := container.Resources.Requests.Memory(); memory != nil {
				diagnostic.Resources[container.Name+"_memory_request"] = memory.String()
			}
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
					fmt.Sprintf("%s: %s (%s)", event.Reason, event.Message, event.LastTimestamp.Format(time.RFC3339)))
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
		Timestamp:       time.Now(),
	}

	// Get node information
	nodes, err := s.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	health.NodeCount = len(nodes.Items)
	unhealthyNodes := []string{}
	
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
		} else {
			unhealthyNodes = append(unhealthyNodes, node.Name)
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
	totalPods := 0
	
	for _, pod := range pods.Items {
		// Skip completed jobs and succeeded pods
		if pod.Status.Phase == "Succeeded" {
			continue
		}
		
		totalPods++
		hasIssues := false
		
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

	// Add resource usage statistics
	health.ResourceUsage["total_pods"] = totalPods
	health.ResourceUsage["problem_pods"] = problemPods
	health.ResourceUsage["problem_percentage"] = float64(problemPods) / float64(totalPods) * 100

	// Generate recommendations
	if float64(health.HealthyNodes)/float64(health.NodeCount) < 0.8 {
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Cluster has %d unhealthy nodes: %s", 
				len(unhealthyNodes), strings.Join(unhealthyNodes, ", ")))
	}
	
	if problemPods > 10 {
		health.Recommendations = append(health.Recommendations,
			"High number of problematic pods detected - investigate cluster resource constraints")
	}

	if float64(problemPods)/float64(totalPods) > 0.2 {
		health.Recommendations = append(health.Recommendations,
			"More than 20% of pods have issues - consider cluster-wide investigation")
	}

	return health, nil
}

func (s *K8sDiagnosticsServer) analyzePodLogs(ctx context.Context, namespace, podName, container string, lines int64) (*LogAnalysis, error) {
	logOptions := &corev1.PodLogOptions{
		TailLines: &lines,
	}
	if container != "" {
		logOptions.Container = container
	}

	logs, err := s.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions).Do(ctx).Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	logText := string(logs)
	logLines := strings.Split(logText, "\n")

	analysis := &LogAnalysis{
		PodName:     podName,
		Namespace:   namespace,
		LogLines:    len(logLines),
		ErrorsFound: []string{},
		Suggestions: []string{},
	}

	// Analyze logs for common error patterns
	errorPatterns := []string{
		"error", "fatal", "exception", "panic", "failed", "timeout",
		"connection refused", "permission denied", "out of memory",
		"killed", "segmentation fault", "stack overflow",
	}

	warningPatterns := []string{
		"warning", "warn", "deprecated", "retry", "fallback",
	}

	errorMap := make(map[string]bool)
	suggestionMap := make(map[string]bool)

	for _, line := range logLines {
		lowerLine := strings.ToLower(line)
		
		// Check for errors
		for _, pattern := range errorPatterns {
			if strings.Contains(lowerLine, pattern) {
				if !errorMap[line] {
					analysis.ErrorsFound = append(analysis.ErrorsFound, line)
					analysis.ErrorCount++
					errorMap[line] = true
				}

				// Add specific suggestions
				var suggestion string
				switch pattern {
				case "out of memory":
					suggestion = "Consider increasing memory limits or optimizing application memory usage"
				case "connection refused":
					suggestion = "Check network policies, service configurations, and target service availability"
				case "permission denied":
					suggestion = "Review RBAC permissions and file system permissions"
				case "timeout":
					suggestion = "Check network connectivity and increase timeout values if appropriate"
				case "killed":
					suggestion = "Pod may have been killed due to resource limits (OOMKilled) - check resource usage"
				case "segmentation fault":
					suggestion = "Application crash detected - review application code and dependencies"
				}
				
				if suggestion != "" && !suggestionMap[suggestion] {
					analysis.Suggestions = append(analysis.Suggestions, suggestion)
					suggestionMap[suggestion] = true
				}
				break
			}
		}

		// Check for warnings
		for _, pattern := range warningPatterns {
			if strings.Contains(lowerLine, pattern) {
				analysis.WarningCount++
				break
			}
		}
	}

	// Add general suggestions based on error count
	if analysis.ErrorCount > 10 {
		analysis.Suggestions = append(analysis.Suggestions, 
			"High error rate detected - consider reviewing application stability")
	}
	
	if analysis.WarningCount > 5 {
		analysis.Suggestions = append(analysis.Suggestions,
			"Multiple warnings detected - review application configuration")
	}

	return analysis, nil
}

func (s *K8sDiagnosticsServer) getWorkloadRecommendations(ctx context.Context, namespace string) ([]string, error) {
	recommendations := []string{}

	// Check deployments
	deployments, err := s.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, deployment := range deployments.Items {
		hasLimits := false
		hasRequests := false
		
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Resources.Limits != nil {
				hasLimits = true
			}
			if container.Resources.Requests != nil {
				hasRequests = true
			}
		}
		
		if !hasLimits {
			recommendations = append(recommendations,
				fmt.Sprintf("Deployment %s/%s should have resource limits", 
					deployment.Namespace, deployment.Name))
		}
		
		if !hasRequests {
			recommendations = append(recommendations,
				fmt.Sprintf("Deployment %s/%s should have resource requests", 
					deployment.Namespace, deployment.Name))
		}

		// Check replica count
		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 1 {
			recommendations = append(recommendations,
				fmt.Sprintf("Deployment %s/%s has only 1 replica - consider scaling for HA", 
					deployment.Namespace, deployment.Name))
		}

		// Check for missing probes
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.LivenessProbe == nil {
				recommendations = append(recommendations,
					fmt.Sprintf("Container %s in deployment %s/%s missing liveness probe",
						container.Name, deployment.Namespace, deployment.Name))
			}
			if container.ReadinessProbe == nil {
				recommendations = append(recommendations,
					fmt.Sprintf("Container %s in deployment %s/%s missing readiness probe",
						container.Name, deployment.Namespace, deployment.Name))
			}
		}
	}

	return recommendations, nil
}

// Additional exploratory tools to add to the existing K8s diagnostics MCP server

// Tool: Find and diagnose problematic pods
func (s *K8sDiagnosticsServer) findProblematicPods(ctx context.Context, namespace string, criteria string) ([]PodDiagnostic, error) {
	var pods *corev1.PodList
	var err error

	if namespace == "" || namespace == "all" {
		pods, err = s.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	} else {
		pods, err = s.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, err
	}

	var problematicPods []PodDiagnostic
	
	for _, pod := range pods.Items {
		// Skip system namespaces unless specifically requested
		if namespace == "" && (strings.HasPrefix(pod.Namespace, "kube-") || 
			pod.Namespace == "kube-system" || pod.Namespace == "kube-public" || 
			pod.Namespace == "kube-node-lease") {
			continue
		}

		isProblem := false
		
		switch criteria {
		case "failing", "failed", "error":
			isProblem = pod.Status.Phase == "Failed" || pod.Status.Phase == "Pending"
		case "restarting", "restart":
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.RestartCount > 3 {
					isProblem = true
					break
				}
			}
		case "not-ready", "unready":
			for _, cs := range pod.Status.ContainerStatuses {
				if !cs.Ready {
					isProblem = true
					break
				}
			}
		case "resource-issues":
			// Check for resource-related issues
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Waiting != nil && 
					(strings.Contains(cs.State.Waiting.Reason, "Memory") || 
					 strings.Contains(cs.State.Waiting.Reason, "CPU")) {
					isProblem = true
					break
				}
			}
		case "image-issues":
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Waiting != nil && 
					(cs.State.Waiting.Reason == "ImagePullBackOff" || 
					 cs.State.Waiting.Reason == "ErrImagePull") {
					isProblem = true
					break
				}
			}
		default: // "all" or any other criteria - find any problematic pods
			isProblem = pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded"
			if !isProblem {
				for _, cs := range pod.Status.ContainerStatuses {
					if cs.RestartCount > 3 || !cs.Ready {
						isProblem = true
						break
					}
				}
			}
		}

		if isProblem {
			diagnostic, err := s.diagnosePod(ctx, pod.Namespace, pod.Name)
			if err == nil {
				problematicPods = append(problematicPods, *diagnostic)
			}
		}
	}

	return problematicPods, nil
}

// Tool: Search pods by name pattern
func (s *K8sDiagnosticsServer) searchPods(ctx context.Context, namePattern string, namespace string) ([]PodDiagnostic, error) {
	var pods *corev1.PodList
	var err error

	if namespace == "" || namespace == "all" {
		pods, err = s.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	} else {
		pods, err = s.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, err
	}

	var matchingPods []PodDiagnostic
	pattern := strings.ToLower(namePattern)
	
	for _, pod := range pods.Items {
		// Skip system namespaces unless specifically requested
		if namespace == "" && (strings.HasPrefix(pod.Namespace, "kube-") || 
			pod.Namespace == "kube-system" || pod.Namespace == "kube-public" || 
			pod.Namespace == "kube-node-lease") {
			continue
		}

		podName := strings.ToLower(pod.Name)
		podNamespace := strings.ToLower(pod.Namespace)
		
		// Match against pod name, namespace, or labels
		if strings.Contains(podName, pattern) || 
		   strings.Contains(podNamespace, pattern) {
			diagnostic, err := s.diagnosePod(ctx, pod.Namespace, pod.Name)
			if err == nil {
				matchingPods = append(matchingPods, *diagnostic)
			}
		} else {
			// Check labels
			for key, value := range pod.Labels {
				if strings.Contains(strings.ToLower(key), pattern) || 
				   strings.Contains(strings.ToLower(value), pattern) {
					diagnostic, err := s.diagnosePod(ctx, pod.Namespace, pod.Name)
					if err == nil {
						matchingPods = append(matchingPods, *diagnostic)
					}
					break
				}
			}
		}
	}

	return matchingPods, nil
}

// Tool: Get resource usage across pods
func (s *K8sDiagnosticsServer) getResourceUsage(ctx context.Context, namespace string, sortBy string) ([]PodResourceInfo, error) {
	var pods *corev1.PodList
	var err error

	if namespace == "" || namespace == "all" {
		pods, err = s.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	} else {
		pods, err = s.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, err
	}

	var resourceInfo []PodResourceInfo
	
	for _, pod := range pods.Items {
		// Skip system namespaces unless specifically requested
		if namespace == "" && (strings.HasPrefix(pod.Namespace, "kube-") || 
			pod.Namespace == "kube-system" || pod.Namespace == "kube-public" || 
			pod.Namespace == "kube-node-lease") {
			continue
		}

		info := PodResourceInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
		}

		totalRestarts := int32(0)
		hasResourceIssues := false

		// Aggregate resource information from all containers
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				if cpu := container.Resources.Requests.Cpu(); cpu != nil {
					info.CPURequest += cpu.String() + " "
				}
				if memory := container.Resources.Requests.Memory(); memory != nil {
					info.MemoryRequest += memory.String() + " "
				}
			}
			if container.Resources.Limits != nil {
				if cpu := container.Resources.Limits.Cpu(); cpu != nil {
					info.CPULimit += cpu.String() + " "
				}
				if memory := container.Resources.Limits.Memory(); memory != nil {
					info.MemoryLimit += memory.String() + " "
				}
			}

			// Check if container has no resource configuration
			if container.Resources.Requests == nil && container.Resources.Limits == nil {
				hasResourceIssues = true
			}
		}

		// Get restart count
		for _, cs := range pod.Status.ContainerStatuses {
			totalRestarts += cs.RestartCount
			
			// Check for resource-related waiting states
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				if strings.Contains(reason, "Memory") || strings.Contains(reason, "CPU") || 
				   reason == "OOMKilled" {
					hasResourceIssues = true
				}
			}
		}

		info.RestartCount = totalRestarts
		info.HasResourceIssues = hasResourceIssues
		
		resourceInfo = append(resourceInfo, info)
	}

	// TODO: Implement sorting based on sortBy parameter
	// Could sort by restart count, resource usage, etc.

	return resourceInfo, nil
}

func main() {
	s := server.NewMCPServer(
		"K8s Diagnostics MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	diagnostics, err := NewK8sDiagnosticsServer()
	if err != nil {
		log.Fatalf("Failed to create diagnostics server: %v", err)
	}

	// Tool: Diagnose Pod
	diagnosePodTool := mcp.NewTool("diagnose_pod",
		mcp.WithDescription("Diagnose issues with a specific Kubernetes pod"),
		mcp.WithString("namespace", mcp.Description("Namespace of the pod (default: default)")),
		mcp.WithString("pod_name", mcp.Required(), mcp.Description("Name of the pod")),
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

		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
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
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// Tool: Get workload recommendations
	recommendTool := mcp.NewTool("get_workload_recommendations",
		mcp.WithDescription("Get optimization recommendations for workloads in a namespace"),
		mcp.WithString("namespace", mcp.Description("Namespace to scan (default: default)")),
	)

	s.AddTool(recommendTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := req.GetString("namespace", "default")
		result, err := diagnostics.getWorkloadRecommendations(ctx, namespace)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("recommendation generation failed", err), nil
		}
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// Tool: Analyze pod logs
	analyzePodLogsTool := mcp.NewTool("analyze_pod_logs",
		mcp.WithDescription("Get and analyze pod logs for common error patterns"),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)")),
		mcp.WithString("pod_name", mcp.Required(), mcp.Description("Name of the pod")),
		mcp.WithString("container", mcp.Description("Container name (optional)")),
		mcp.WithNumber("lines", mcp.Description("Number of log lines to retrieve (default: 100)")),
	)

	s.AddTool(analyzePodLogsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := req.GetString("namespace", "default")
		podName, err := req.RequireString("pod_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		container := req.GetString("container", "")
		lines := int64(req.GetInt("lines", 100))
		if lines <= 0 {
			lines = 100
		}

		result, err := diagnostics.analyzePodLogs(ctx, namespace, podName, container, lines)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("log analysis failed", err), nil
		}

		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// Tool: List pods in namespace
	listPodsTool := mcp.NewTool("list_pods",
		mcp.WithDescription("List all pods in a namespace with their status"),
		mcp.WithString("namespace", mcp.Description("Namespace to list pods from (default: default)")),
		mcp.WithBoolean("show_system", mcp.Description("Include system namespaces (default: false)")),
	)

	s.AddTool(listPodsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := req.GetString("namespace", "default")
		showSystem := req.GetBool("show_system", false)

		var pods *corev1.PodList
		var err error

		if namespace == "all" || showSystem {
			pods, err = diagnostics.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		} else {
			pods, err = diagnostics.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		}

		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to list pods", err), nil
		}

		type PodInfo struct {
			Name         string `json:"name"`
			Namespace    string `json:"namespace"`
			Status       string `json:"status"`
			Ready        string `json:"ready"`
			Restarts     int32  `json:"restarts"`
			Age          string `json:"age"`
		}

		var podList []PodInfo
		for _, pod := range pods.Items {
			// Skip system namespaces unless explicitly requested
			if !showSystem && (strings.HasPrefix(pod.Namespace, "kube-") || 
				pod.Namespace == "kube-system" || pod.Namespace == "kube-public" || 
				pod.Namespace == "kube-node-lease") {
				continue
			}

			readyCount := 0
			totalCount := len(pod.Status.ContainerStatuses)
			restarts := int32(0)

			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Ready {
					readyCount++
				}
				restarts += cs.RestartCount
			}

			age := time.Since(pod.CreationTimestamp.Time).Truncate(time.Second).String()

			podList = append(podList, PodInfo{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Status:    string(pod.Status.Phase),
				Ready:     fmt.Sprintf("%d/%d", readyCount, totalCount),
				Restarts:  restarts,
				Age:       age,
			})
		}

		result := map[string]interface{}{
			"namespace": namespace,
			"pod_count": len(podList),
			"pods":      podList,
		}

		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// Tool: Find problematic pods
	findProblematicTool := mcp.NewTool("find_problematic_pods",
		mcp.WithDescription("Find and diagnose pods with issues (failing, restarting, not ready, etc.)"),
		mcp.WithString("namespace", mcp.Description("Namespace to search (default: all non-system namespaces)")),
		mcp.WithString("criteria", mcp.Description("Type of problems to find: failing, restarting, not-ready, resource-issues, image-issues, or all (default: all)")),
	)

	s.AddTool(findProblematicTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := req.GetString("namespace", "")
		criteria := req.GetString("criteria", "all")

		result, err := diagnostics.findProblematicPods(ctx, namespace, criteria)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to find problematic pods", err), nil
		}

		response := map[string]interface{}{
			"search_criteria": criteria,
			"namespace": namespace,
			"problem_count": len(result),
			"problematic_pods": result,
		}

		jsonBytes, _ := json.MarshalIndent(response, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// Tool: Search pods by pattern
	searchPodsTool := mcp.NewTool("search_pods",
		mcp.WithDescription("Search for pods by name pattern, namespace, or labels and get their diagnostics"),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("Search pattern (pod name, namespace, or label value)")),
		mcp.WithString("namespace", mcp.Description("Namespace to search (default: all non-system namespaces)")),
	)

	s.AddTool(searchPodsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pattern, err := req.RequireString("pattern")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		
		namespace := req.GetString("namespace", "")

		result, err := diagnostics.searchPods(ctx, pattern, namespace)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("pod search failed", err), nil
		}

		response := map[string]interface{}{
			"search_pattern": pattern,
			"namespace": namespace,
			"matches_found": len(result),
			"matching_pods": result,
		}

		jsonBytes, _ := json.MarshalIndent(response, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// Tool: Get resource usage overview
	resourceUsageTool := mcp.NewTool("get_resource_usage",
		mcp.WithDescription("Get resource usage overview for pods to identify resource-related issues"),
		mcp.WithString("namespace", mcp.Description("Namespace to analyze (default: all non-system namespaces)")),
		mcp.WithString("sort_by", mcp.Description("Sort results by: restarts, cpu, memory (default: restarts)")),
	)

	s.AddTool(resourceUsageTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := req.GetString("namespace", "")
		sortBy := req.GetString("sort_by", "restarts")

		result, err := diagnostics.getResourceUsage(ctx, namespace, sortBy)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("resource usage analysis failed", err), nil
		}

		response := map[string]interface{}{
			"namespace": namespace,
			"sort_by": sortBy,
			"pod_count": len(result),
			"resource_usage": result,
		}

		jsonBytes, _ := json.MarshalIndent(response, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// Tool: Quick cluster triage
	triageTool := mcp.NewTool("quick_triage",
		mcp.WithDescription("Perform quick cluster triage to identify immediate issues across all namespaces"),
	)

	s.AddTool(triageTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get cluster health
		clusterHealth, err := diagnostics.analyzeClusterHealth(ctx)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("cluster health check failed", err), nil
		}

		// Find critical issues
		criticalPods, err := diagnostics.findProblematicPods(ctx, "", "failing")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to find critical pods", err), nil
		}

		// Find high restart pods  
		restartingPods, err := diagnostics.findProblematicPods(ctx, "", "restarting")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to find restarting pods", err), nil
		}

		response := map[string]interface{}{
			"timestamp": time.Now(),
			"cluster_health": clusterHealth,
			"critical_pods": criticalPods,
			"restarting_pods": restartingPods,
			"immediate_actions": []string{
				"Check critical/failing pods first",
				"Investigate high restart count pods", 
				"Review cluster resource availability",
				"Check node health status",
			},
		}

		jsonBytes, _ := json.MarshalIndent(response, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	// Add comprehensive troubleshooting guide resource
	troubleshootingGuide := `# Kubernetes Diagnostics MCP Server Guide

This MCP server provides comprehensive Kubernetes cluster diagnostics and troubleshooting capabilities.

## Available Tools

### 1. diagnose_pod
Performs detailed analysis of a specific pod including:
- Container status and restart counts
- Resource configuration
- Recent events
- Common issues and suggestions

**Usage:** Provide namespace and pod_name

### 2. analyze_cluster_health
Provides cluster-wide health analysis including:
- Node status and health
- Namespace and pod statistics
- Problem pod identification
- Resource usage metrics
- Recommendations for cluster improvements

### 3. get_workload_recommendations
Analyzes workloads (deployments) for best practices:
- Resource requests and limits
- Replica counts for HA
- Health probe configurations
- Security and reliability recommendations

### 4. analyze_pod_logs
Advanced log analysis with pattern detection:
- Error pattern identification
- Warning detection
- Contextual suggestions
- Statistical analysis of log issues

### 5. list_pods
Lists pods with status information:
- Pod status and readiness
- Restart counts
- Age information
- Namespace filtering

### 6. find_problematic_pods
Finds pods with specific issues:
- Failing pods
- Restarting pods
- Not ready pods
- Resource-related issues
- Image pull issues
- All problematic pods
- Provides diagnostics for identified pods

### 7. search_pods
Searches for pods by name pattern, namespace, or labels:
- Find pods matching a specific name or label
- Get diagnostics for matching pods

### 8. get_resource_usage
Analyzes resource usage across pods:
- CPU and memory requests/limits
- Restart counts
- Resource issues detection
- Sorting options for resource usage

### 9. quick_triage
Performs quick cluster triage:
- Cluster health overview
- Critical/failing pod identification
- High restart pod detection
- Immediate actions for cluster issues

## Integration with Other MCP Servers

This server is designed to work alongside:
- **Filesystem MCP Server**: For reading issue reports and saving diagnostic results
- **GitHub MCP Server**: For creating issues from diagnostic reports

## Common Troubleshooting Patterns

### High Restart Count
1. Use diagnose_pod to identify the problematic pod
2. Use analyze_pod_logs to examine error patterns
3. Check resource limits and requests
4. Review recent events for context

### Cluster Performance Issues
1. Use analyze_cluster_health for overview
2. Use get_workload_recommendations for optimization
3. Check node health and resource allocation
4. Identify problematic workloads

### Application Errors
1. Use analyze_pod_logs for detailed error analysis
2. Cross-reference with pod diagnostic information
3. Check networking and permissions
4. Review resource constraints

## Best Practices

- Always start with cluster health analysis
- Use pod diagnostics for specific issues
- Combine log analysis with diagnostic data
- Save diagnostic reports for tracking
- Create GitHub issues for follow-up actions
`

	guideResource := mcp.NewResource("k8s://diagnostics/guide", 
		"Comprehensive Kubernetes diagnostics guide", 
		mcp.WithMIMEType("text/markdown"))
	
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