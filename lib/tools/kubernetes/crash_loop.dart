import 'package:k8s/k8s.dart';
import 'package:logging/logging.dart';

final _logger = Logger('CrashLoop');

Future<Map<String, dynamic>> detectCrashLoopPods(
    Map<String, dynamic> params) async {
  try {
    final client = K8sClient();
    final namespace = params['namespace'] as String?;

    final pods = await client.listPods(namespace: namespace);
    final crashLoopPods = <Map<String, dynamic>>[];

    for (final pod in pods) {
      final containerStatuses = pod.status?.containerStatuses ?? [];
      for (final container in containerStatuses) {
        final restartCount = container.restartCount ?? 0;
        final lastState = container.lastState;

        if (restartCount > 0 && lastState != null) {
          final terminated = lastState.terminated;
          if (terminated != null && terminated.reason == 'CrashLoopBackOff') {
            crashLoopPods.add({
              'pod_name': pod.metadata?.name,
              'namespace': pod.metadata?.namespace,
              'container_name': container.name,
              'restart_count': restartCount,
              'last_state': {
                'reason': terminated.reason,
                'exit_code': terminated.exitCode,
                'message': terminated.message,
                'finished_at': terminated.finishedAt?.toIso8601String()
              }
            });
          }
        }
      }
    }

    return {
      'status': 'success',
      'data': {
        'crash_loop_pods': crashLoopPods,
        'total_pods_checked': pods.length,
        'timestamp': DateTime.now().toIso8601String()
      }
    };
  } catch (e) {
    _logger.severe('Error detecting crash loop pods: $e');
    return {'status': 'error', 'error': e.toString()};
  }
}
