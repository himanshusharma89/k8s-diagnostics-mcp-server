# Kubernetes Diagnostics MCP Server

A powerful Kubernetes diagnostics server built with Dart that provides real-time insights into your Kubernetes cluster through an MCP (Model-Controller-Provider) interface.

## Features

- **Cluster Status Monitoring**: Get real-time status of your Kubernetes cluster
- **Pod Logs**: Retrieve logs from any pod in your cluster
- **Crash Loop Detection**: Automatically detect pods stuck in crash loops
- **Event Summarization**: Get a summary of recent cluster events

## Getting Started

### Prerequisites

- Dart SDK (>=3.0.0)
- Kubernetes cluster access
- kubectl configured with cluster access

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/k8s-diagnostics-mcp-server.git
cd k8s-diagnostics-mcp-server
```

2. Install dependencies:
```bash
dart pub get
```

3. Run the server:
```bash
dart run bin/k8s_diagnostics_mcp_server.dart
```

The server will start on port 8080 by default.

## Available Tools

### get_cluster_status
Retrieves the current status of your Kubernetes cluster, including node and pod information.

### get_pod_logs
Fetches logs from a specific pod in your cluster.

Parameters:
- `namespace` (optional): The namespace containing the pod (defaults to 'default')
- `pod_name`: The name of the pod to fetch logs from

### detect_crash_loop_pods
Identifies pods that are stuck in crash loops by analyzing their restart counts.

### summarize_events
Provides a summary of recent cluster events from the last 24 hours.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
