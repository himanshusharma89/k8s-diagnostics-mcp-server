import 'package:k8s/k8s.dart';
import 'package:logging/logging.dart';

final _logger = Logger('ClusterStatus');

Future<Map<String, dynamic>> getClusterStatus(
    Map<String, dynamic> params) async {
  try {
    final client = K8sClient();

    // Get cluster info
    final nodes = await client.listNodes();
    final pods = await client.listPods();
    final deployments = await client.listDeployments();

    // Calculate cluster health metrics
    final totalNodes = nodes.length;
    final readyNodes = nodes
        .where((node) =>
            node.status?.conditions
                ?.any((c) => c.type == 'Ready' && c.status == 'True') ??
            false)
        .length;

    final totalPods = pods.length;
    final runningPods =
        pods.where((pod) => pod.status?.phase == 'Running').length;

    final totalDeployments = deployments.length;
    final availableDeployments = deployments
        .where((deployment) =>
            deployment.status?.availableReplicas == deployment.status?.replicas)
        .length;

    return {
      'status': 'success',
      'data': {
        'cluster_health': {
          'nodes': {
            'total': totalNodes,
            'ready': readyNodes,
            'health_percentage':
                (readyNodes / totalNodes * 100).toStringAsFixed(2)
          },
          'pods': {
            'total': totalPods,
            'running': runningPods,
            'health_percentage':
                (runningPods / totalPods * 100).toStringAsFixed(2)
          },
          'deployments': {
            'total': totalDeployments,
            'available': availableDeployments,
            'health_percentage': (availableDeployments / totalDeployments * 100)
                .toStringAsFixed(2)
          }
        },
        'timestamp': DateTime.now().toIso8601String()
      }
    };
  } catch (e) {
    _logger.severe('Error getting cluster status: $e');
    return {'status': 'error', 'error': e.toString()};
  }
}
