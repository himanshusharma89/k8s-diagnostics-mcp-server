import 'package:k8s/k8s.dart';
import 'package:logging/logging.dart';

final _logger = Logger('PodLogs');

Future<Map<String, dynamic>> getPodLogs(Map<String, dynamic> params) async {
  try {
    final client = K8sClient();
    final namespace = params['namespace'] as String? ?? 'default';
    final podName = params['pod_name'] as String;
    final containerName = params['container_name'] as String?;
    final tailLines = params['tail_lines'] as int? ?? 100;

    final logs = await client.getPodLogs(
      namespace: namespace,
      podName: podName,
      containerName: containerName,
      tailLines: tailLines,
    );

    return {
      'status': 'success',
      'data': {
        'pod_name': podName,
        'namespace': namespace,
        'container_name': containerName,
        'logs': logs,
        'timestamp': DateTime.now().toIso8601String()
      }
    };
  } catch (e) {
    _logger.severe('Error getting pod logs: $e');
    return {'status': 'error', 'error': e.toString()};
  }
}
