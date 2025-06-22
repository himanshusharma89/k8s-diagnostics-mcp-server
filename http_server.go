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
}

func NewHTTPServer() (*HTTPServer, error) {
	diagnostics, err := NewK8sDiagnosticsServer()
	if err != nil {
		return nil, err
	}
	return &HTTPServer{diagnostics: diagnostics}, nil
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.diagnostics.diagnosePod(ctx, req.Namespace, req.PodName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to diagnose pod: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) handleAnalyzeClusterHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.diagnostics.analyzeClusterHealth(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cluster health analysis failed: %v", err), http.StatusInternalServerError)
		return
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.diagnostics.analyzePodLogs(ctx, req.Namespace, req.PodName, req.Container, int64(req.Lines))
	if err != nil {
		http.Error(w, fmt.Sprintf("Log analysis failed: %v", err), http.StatusInternalServerError)
		return
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

	result := map[string]interface{}{
		"namespace": req.Namespace,
		"pod_count": len(podList),
		"pods":      podList,
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.diagnostics.findProblematicPods(ctx, req.Namespace, req.Criteria)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to find problematic pods: %v", err), http.StatusInternalServerError)
		return
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.diagnostics.getResourceUsage(ctx, req.Namespace, req.SortBy)
	if err != nil {
		http.Error(w, fmt.Sprintf("Resource usage analysis failed: %v", err), http.StatusInternalServerError)
		return
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

	response := map[string]interface{}{
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.diagnostics.getWorkloadRecommendations(ctx, req.Namespace)
	if err != nil {
		http.Error(w, fmt.Sprintf("Recommendation generation failed: %v", err), http.StatusInternalServerError)
		return
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.diagnostics.searchPods(ctx, req.Pattern, req.Namespace)
	if err != nil {
		http.Error(w, fmt.Sprintf("Pod search failed: %v", err), http.StatusInternalServerError)
		return
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
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func main() {
	// Check if we should run as HTTP server or MCP server
	if os.Getenv("HTTP_MODE") == "true" {
		runHTTPServer()
	} else {
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

	log.Printf("Starting HTTP server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
