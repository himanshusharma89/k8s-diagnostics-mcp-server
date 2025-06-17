# Kubernetes Diagnostics MCP Server

A Model Context Protocol (MCP) server built with Dart that provides Kubernetes diagnostic tools. This server exposes various Kubernetes diagnostic capabilities through a standardized MCP interface.

## Features

- `get_cluster_status`: Get overall cluster health and status
- `get_pod_logs`: Retrieve logs from specific pods
- `detect_crash_loop_pods`: Identify pods stuck in crash loop
- `summarize_events`: Get a summary of Kubernetes events

## Prerequisites

- Dart SDK >= 3.0.0
- Kubernetes cluster access
- kubeconfig file

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/k8s-diagnostics-mcp-server.git
cd k8s-diagnostics-mcp-server
```

2. Install dependencies:
```bash
dart pub get
```

## Usage

Run the server:
```bash
dart run bin/k8s_diagnostics_server.dart
```

Options:
- `--port`: Server port (default: 8080)
- `--kubeconfig`: Path to kubeconfig file (default: ~/.kube/config)
- `--help`: Show help message

## MCP Tools

### get_cluster_status
Returns the current status of the Kubernetes cluster, including node and pod information.

### get_pod_logs
Retrieves logs from a specific pod.

Parameters:
- `namespace`: Pod namespace
- `pod_name`: Name of the pod
- `container`: (Optional) Container name

### detect_crash_loop_pods
Identifies pods that are stuck in a crash loop state.

### summarize_events
Provides a summary of Kubernetes events, optionally filtered by namespace.

Parameters:
- `namespace`: (Optional) Filter events by namespace

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License
