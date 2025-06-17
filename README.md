# Kubernetes Diagnostics MCP Server

A Dart-based Model Context Protocol (MCP) server for Kubernetes diagnostics. This server provides tools for monitoring and troubleshooting Kubernetes clusters.

## Features

- Get cluster status and health metrics
- Retrieve pod logs
- Detect crash loop pods
- Summarize cluster events

## Getting Started

### Prerequisites

- Dart SDK (>=3.0.0)
- Kubernetes cluster access
- kubectl configured

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
dart run bin/server.dart
```

The server will start on port 8080 by default. You can change the port by setting the `PORT` environment variable.

## API Endpoints

### Get Cluster Status
```http
POST /tools/get_cluster_status
```

### Get Pod Logs
```http
POST /tools/get_pod_logs
{
  "namespace": "default",
  "pod_name": "my-pod",
  "container_name": "main",
  "tail_lines": 100
}
```

### Detect Crash Loop Pods
```http
POST /tools/detect_crash_loop_pods
{
  "namespace": "default"
}
```

### Summarize Events
```http
POST /tools/summarize_events
{
  "namespace": "default",
  "duration": "1h"
}
```

## Development

### Running Tests
```bash
dart test
```

### Building
```bash
dart compile exe bin/server.dart -o k8s-diagnostics-mcp-server
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
