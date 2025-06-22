package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HTTPServer struct {
	diagnostics *K8sDiagnosticsServer
	demoMode    bool
}

func NewHTTPServer() (*HTTPServer, error) {
	demoMode := os.Getenv("DEMO_MODE") == "true"

	var diagnostics *K8sDiagnosticsServer
	var err error

	if demoMode {
		log.Println("Running in DEMO mode with mock data")
		diagnostics = &K8sDiagnosticsServer{} // Empty for demo mode
	} else {
		log.Println("Running in REAL mode with Kubernetes cluster")
		diagnostics, err = NewK8sDiagnosticsServer()
		if err != nil {
			return nil, err
		}
	}

	return &HTTPServer{
		diagnostics: diagnostics,
		demoMode:    demoMode,
	}, nil
}

// Mock data for demo mode
func (s *HTTPServer) getMockPodDiagnostic() *PodDiagnostic {
	return &PodDiagnostic{
		Name:         "demo-app-pod",
		Namespace:    "default",
		Status:       "Running",
		RestartCount: 2,
		Issues: []string{
			"Container has restarted 2 times in the last hour",
			"Memory usage is at 85% of limit",
		},
		Suggestions: []string{
			"Check application logs for errors",
			"Consider increasing memory limits",
			"Review resource requests",
		},
		Events: []string{
			"Pod scheduled successfully",
			"Container started",
			"Container restarted due to OOM",
		},
		Resources: map[string]string{
			"app-container_cpu_request":    "100m",
			"app-container_memory_request": "128Mi",
			"app-container_cpu_limit":      "200m",
			"app-container_memory_limit":   "256Mi",
		},
		CreatedAt: time.Now(),
	}
}

func (s *HTTPServer) getMockClusterHealth() *ClusterHealth {
	return &ClusterHealth{
		NodeCount:      3,
		HealthyNodes:   2,
		NamespaceCount: 5,
		PodIssues: []PodDiagnostic{
			*s.getMockPodDiagnostic(),
		},
		ResourceUsage: map[string]interface{}{
			"cpu_usage_percent":    65,
			"memory_usage_percent": 78,
			"disk_usage_percent":   45,
		},
		Recommendations: []string{
			"Consider scaling up nodes for better resource distribution",
			"Review pod resource requests and limits",
			"Monitor node health more frequently",
		},
		Timestamp: time.Now(),
	}
}

func (s *HTTPServer) getMockLogAnalysis() *LogAnalysis {
	return &LogAnalysis{
		PodName:   "demo-app-pod",
		Namespace: "default",
		LogLines:  150,
		ErrorsFound: []string{
			"ERROR: Database connection timeout",
			"ERROR: Failed to process request",
			"WARN: High memory usage detected",
		},
		Suggestions: []string{
			"Check database connectivity",
			"Review application error handling",
			"Monitor memory usage patterns",
		},
		ErrorCount:   2,
		WarningCount: 1,
	}
}

func (s *HTTPServer) handleDiagnosePod(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Namespace string `json:"namespace"`
		PodName   string `json:"pod_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.PodName == "" {
		http.Error(w, "pod_name is required", http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	var result *PodDiagnostic
	var err error

	if s.demoMode {
		result = s.getMockPodDiagnostic()
		result.Name = req.PodName
		result.Namespace = req.Namespace
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err = s.diagnostics.diagnosePod(ctx, req.Namespace, req.PodName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to diagnose pod: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) handleAnalyzeClusterHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var result *ClusterHealth
	var err error

	if s.demoMode {
		result = s.getMockClusterHealth()
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err = s.diagnostics.analyzeClusterHealth(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cluster health analysis failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) handleAnalyzePodLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Namespace string `json:"namespace"`
		PodName   string `json:"pod_name"`
		Container string `json:"container"`
		Lines     int    `json:"lines"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.PodName == "" {
		http.Error(w, "pod_name is required", http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	if req.Lines <= 0 {
		req.Lines = 100
	}

	var result *LogAnalysis
	var err error

	if s.demoMode {
		result = s.getMockLogAnalysis()
		result.PodName = req.PodName
		result.Namespace = req.Namespace
		result.LogLines = req.Lines
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err = s.diagnostics.analyzePodLogs(ctx, req.Namespace, req.PodName, req.Container, int64(req.Lines))
		if err != nil {
			http.Error(w, fmt.Sprintf("Log analysis failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) handleListPods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Namespace  string `json:"namespace"`
		ShowSystem bool   `json:"show_system"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	var result map[string]interface{}

	if s.demoMode {
		// Mock pod list
		pods := []map[string]interface{}{
			{
				"name":      "demo-app-pod",
				"namespace": "default",
				"status":    "Running",
				"ready":     "1/1",
				"restarts":  2,
				"age":       "2h30m",
			},
			{
				"name":      "demo-db-pod",
				"namespace": "default",
				"status":    "Running",
				"ready":     "1/1",
				"restarts":  0,
				"age":       "1h45m",
			},
			{
				"name":      "demo-web-pod",
				"namespace": "default",
				"status":    "Pending",
				"ready":     "0/1",
				"restarts":  0,
				"age":       "5m",
			},
		}

		result = map[string]interface{}{
			"namespace": req.Namespace,
			"pod_count": len(pods),
			"pods":      pods,
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var pods *corev1.PodList
		var err error

		if req.Namespace == "all" || req.ShowSystem {
			pods, err = s.diagnostics.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		} else {
			pods, err = s.diagnostics.clientset.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{})
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list pods: %v", err), http.StatusInternalServerError)
			return
		}

		type PodInfo struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Status    string `json:"status"`
			Ready     string `json:"ready"`
			Restarts  int32  `json:"restarts"`
			Age       string `json:"age"`
		}

		var podList []PodInfo
		for _, pod := range pods.Items {
			// Skip system namespaces unless explicitly requested
			if !req.ShowSystem && (filepath.HasPrefix(pod.Namespace, "kube-") ||
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

		result = map[string]interface{}{
			"namespace": req.Namespace,
			"pod_count": len(podList),
			"pods":      podList,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) handleFindProblematicPods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Namespace string `json:"namespace"`
		Criteria  string `json:"criteria"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Criteria == "" {
		req.Criteria = "all"
	}

	var result []PodDiagnostic
	var err error

	if s.demoMode {
		// Mock problematic pods
		result = []PodDiagnostic{
			*s.getMockPodDiagnostic(),
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err = s.diagnostics.findProblematicPods(ctx, req.Namespace, req.Criteria)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to find problematic pods: %v", err), http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"search_criteria":  req.Criteria,
		"namespace":        req.Namespace,
		"problem_count":    len(result),
		"problematic_pods": result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleGetResourceUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Namespace string `json:"namespace"`
		SortBy    string `json:"sort_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.SortBy == "" {
		req.SortBy = "restarts"
	}

	var result []PodResourceInfo
	var err error

	if s.demoMode {
		// Mock resource usage
		result = []PodResourceInfo{
			{
				Name:              "demo-app-pod",
				Namespace:         "default",
				CPURequest:        "100m",
				MemoryRequest:     "128Mi",
				CPULimit:          "200m",
				MemoryLimit:       "256Mi",
				RestartCount:      2,
				Status:            "Running",
				HasResourceIssues: true,
			},
			{
				Name:              "demo-db-pod",
				Namespace:         "default",
				CPURequest:        "500m",
				MemoryRequest:     "512Mi",
				CPULimit:          "1000m",
				MemoryLimit:       "1Gi",
				RestartCount:      0,
				Status:            "Running",
				HasResourceIssues: false,
			},
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err = s.diagnostics.getResourceUsage(ctx, req.Namespace, req.SortBy)
		if err != nil {
			http.Error(w, fmt.Sprintf("Resource usage analysis failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"namespace":      req.Namespace,
		"sort_by":        req.SortBy,
		"pod_count":      len(result),
		"resource_usage": result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleQuickTriage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var response map[string]interface{}

	if s.demoMode {
		// Mock triage data
		response = map[string]interface{}{
			"timestamp":      time.Now(),
			"cluster_health": s.getMockClusterHealth(),
			"critical_pods": []PodDiagnostic{
				*s.getMockPodDiagnostic(),
			},
			"restarting_pods": []PodDiagnostic{
				*s.getMockPodDiagnostic(),
			},
			"immediate_actions": []string{
				"Check critical/failing pods first",
				"Investigate high restart count pods",
				"Review cluster resource availability",
				"Check node health status",
			},
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get cluster health
		clusterHealth, err := s.diagnostics.analyzeClusterHealth(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cluster health check failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Find critical issues
		criticalPods, err := s.diagnostics.findProblematicPods(ctx, "", "failing")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to find critical pods: %v", err), http.StatusInternalServerError)
			return
		}

		// Find high restart pods
		restartingPods, err := s.diagnostics.findProblematicPods(ctx, "", "restarting")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to find restarting pods: %v", err), http.StatusInternalServerError)
			return
		}

		response = map[string]interface{}{
			"timestamp":       time.Now(),
			"cluster_health":  clusterHealth,
			"critical_pods":   criticalPods,
			"restarting_pods": restartingPods,
			"immediate_actions": []string{
				"Check critical/failing pods first",
				"Investigate high restart count pods",
				"Review cluster resource availability",
				"Check node health status",
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleGetWorkloadRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Namespace string `json:"namespace"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	var result []string
	var err error

	if s.demoMode {
		// Mock recommendations
		result = []string{
			"Set resource requests and limits for all containers",
			"Configure health checks (liveness and readiness probes)",
			"Use multiple replicas for high availability",
			"Implement proper logging and monitoring",
			"Review security policies and RBAC",
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err = s.diagnostics.getWorkloadRecommendations(ctx, req.Namespace)
		if err != nil {
			http.Error(w, fmt.Sprintf("Recommendation generation failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) handleSearchPods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Pattern   string `json:"pattern"`
		Namespace string `json:"namespace"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Pattern == "" {
		http.Error(w, "pattern is required", http.StatusBadRequest)
		return
	}

	var result []PodDiagnostic
	var err error

	if s.demoMode {
		// Mock search results
		if req.Pattern == "demo" || req.Pattern == "app" {
			result = []PodDiagnostic{
				*s.getMockPodDiagnostic(),
			}
		} else {
			result = []PodDiagnostic{}
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err = s.diagnostics.searchPods(ctx, req.Pattern, req.Namespace)
		if err != nil {
			http.Error(w, fmt.Sprintf("Pod search failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"search_pattern": req.Pattern,
		"namespace":      req.Namespace,
		"matches_found":  len(result),
		"matching_pods":  result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "healthy",
		"time":      time.Now().Format(time.RFC3339),
		"demo_mode": s.demoMode,
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	fmt.Println(">>> Starting k8s-diagnostics-mcp-server-http main()")
	// Check if we should run as HTTP server or MCP server
	if os.Getenv("HTTP_MODE") == "true" {
		fmt.Println(">>> Starting http mode")
		runHTTPServer()
	} else {
		fmt.Println(">>> Starting mcp mode")
		runMCPServer()
	}
}

func runHTTPServer() {
	server, err := NewHTTPServer()
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

	// Set up routes
	http.HandleFunc("/diagnose_pod", server.handleDiagnosePod)
	http.HandleFunc("/analyze_cluster_health", server.handleAnalyzeClusterHealth)
	http.HandleFunc("/analyze_pod_logs", server.handleAnalyzePodLogs)
	http.HandleFunc("/list_pods", server.handleListPods)
	http.HandleFunc("/find_problematic_pods", server.handleFindProblematicPods)
	http.HandleFunc("/get_resource_usage", server.handleGetResourceUsage)
	http.HandleFunc("/quick_triage", server.handleQuickTriage)
	http.HandleFunc("/get_workload_recommendations", server.handleGetWorkloadRecommendations)
	http.HandleFunc("/search_pods", server.handleSearchPods)
	http.HandleFunc("/health", server.handleHealth)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HTTP server on port %s (Demo Mode: %v)", port, server.demoMode)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
