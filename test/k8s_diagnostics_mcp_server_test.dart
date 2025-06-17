import 'package:test/test.dart';
import 'package:k8s_diagnostics_mcp_server/k8s_diagnostics_mcp_server.dart';
import 'package:k8s_diagnostics_mcp_server/src/models/k8s_models.dart';

void main() {
  group('K8sDiagnosticsMcpServer', () {
    late K8sDiagnosticsMcpServer server;

    setUp(() {
      server = K8sDiagnosticsMcpServer();
    });

    test('get_cluster_status returns correct format', () async {
      final result = await server.getClusterStatus({});
      expect(result, contains('status'));
      expect(result, contains('nodes'));
      expect(result, contains('pods'));
      expect(result, contains('details'));
    });

    test('get_pod_logs requires pod_name parameter', () async {
      final result = await server.getPodLogs({});
      expect(result['status'], equals('error'));
      expect(result, contains('message'));
    });

    test('detect_crash_loop_pods returns correct format', () async {
      final result = await server.detectCrashLoopPods({});
      expect(result, contains('status'));
      expect(result, contains('crash_loop_pods'));
    });

    test('summarize_events returns correct format', () async {
      final result = await server.summarizeEvents({});
      expect(result, contains('status'));
      expect(result, contains('events'));
    });
  });
}
