library k8s_diagnostics_mcp_server;

import 'package:shelf/shelf.dart';
import 'package:shelf/shelf_io.dart' as shelf_io;
import 'src/mcp_server.dart';
import 'src/models/k8s_models.dart';

export 'src/mcp_server.dart';
export 'src/models/k8s_models.dart';

class K8sDiagnosticsMcpServer {
  final McpServer _server;
  final K8sClient _k8sClient;

  K8sDiagnosticsMcpServer()
      : _server = McpServer(),
        _k8sClient = K8sClient() {
    _registerTools();
  }

  void _registerTools() {
    _server.registerTool('get_cluster_status', getClusterStatus);
    _server.registerTool('get_pod_logs', getPodLogs);
    _server.registerTool('detect_crash_loop_pods', detectCrashLoopPods);
    _server.registerTool('summarize_events', summarizeEvents);
  }

  Future<void> serve({int port = 8080}) async {
    final handler = const Pipeline()
        .addMiddleware(logRequests())
        .addHandler(_server.handler);

    final server = await shelf_io.serve(handler, 'localhost', port);
    print('Server running on port ${server.port}');
  }

  Future<Map<String, dynamic>> getClusterStatus(
      Map<String, dynamic> params) async {
    try {
      final nodes = await _k8sClient.listNodes();
      final pods = await _k8sClient.listPods();

      return {
        'status': 'healthy',
        'nodes': nodes.length,
        'pods': pods.length,
        'details': {
          'nodes': nodes.map((n) => n.metadata.name).toList(),
          'pods': pods.map((p) => p.metadata.name).toList(),
        }
      };
    } catch (e) {
      return {
        'status': 'error',
        'message': e.toString(),
      };
    }
  }

  Future<Map<String, dynamic>> getPodLogs(Map<String, dynamic> params) async {
    final namespace = params['namespace'] as String? ?? 'default';
    final podName = params['pod_name'] as String;

    try {
      final logs = await _k8sClient.getPodLogs(namespace, podName);
      return {
        'status': 'success',
        'logs': logs,
      };
    } catch (e) {
      return {
        'status': 'error',
        'message': e.toString(),
      };
    }
  }

  Future<Map<String, dynamic>> detectCrashLoopPods(
      Map<String, dynamic> params) async {
    try {
      final pods = await _k8sClient.listPods();
      final crashLoopPods = pods.where((pod) {
        return pod.status.containerStatuses?.any((status) {
              return status.restartCount > 3;
            }) ??
            false;
      }).toList();

      return {
        'status': 'success',
        'crash_loop_pods': crashLoopPods
            .map((pod) => {
                  'name': pod.metadata.name,
                  'namespace': pod.metadata.namespace,
                  'restart_count':
                      pod.status.containerStatuses?.first.restartCount,
                })
            .toList(),
      };
    } catch (e) {
      return {
        'status': 'error',
        'message': e.toString(),
      };
    }
  }

  Future<Map<String, dynamic>> summarizeEvents(
      Map<String, dynamic> params) async {
    try {
      final events = await _k8sClient.listEvents();
      final recentEvents = events.where((event) {
        final eventTime = event.lastTimestamp;
        return eventTime != null &&
            DateTime.now().difference(eventTime).inHours < 24;
      }).toList();

      return {
        'status': 'success',
        'events': recentEvents
            .map((event) => {
                  'type': event.type,
                  'reason': event.reason,
                  'message': event.message,
                  'timestamp': event.lastTimestamp?.toIso8601String(),
                })
            .toList(),
      };
    } catch (e) {
      return {
        'status': 'error',
        'message': e.toString(),
      };
    }
  }
}
