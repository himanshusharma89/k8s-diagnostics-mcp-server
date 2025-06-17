import 'dart:convert';
import 'package:shelf/shelf.dart';
import 'package:shelf_router/shelf_router.dart';
import 'package:logging/logging.dart';
import '../tools/kubernetes/cluster_status.dart';
import '../tools/kubernetes/pod_logs.dart';
import '../tools/kubernetes/crash_loop.dart';
import '../tools/kubernetes/events.dart';

class MCP {
  final _logger = Logger('MCP');
  final _router = Router();

  MCP() {
    _setupRoutes();
  }

  void _setupRoutes() {
    _router.post('/tools/get_cluster_status', _handleClusterStatus);
    _router.post('/tools/get_pod_logs', _handlePodLogs);
    _router.post('/tools/detect_crash_loop_pods', _handleCrashLoop);
    _router.post('/tools/summarize_events', _handleEvents);
  }

  Future<Response> _handleClusterStatus(Request request) async {
    try {
      final params = await request.readAsString();
      final result = await getClusterStatus(jsonDecode(params));
      return Response.ok(jsonEncode(result));
    } catch (e) {
      _logger.severe('Error in get_cluster_status: $e');
      return Response.internalServerError(
        body: jsonEncode({'error': e.toString()}),
      );
    }
  }

  Future<Response> _handlePodLogs(Request request) async {
    try {
      final params = await request.readAsString();
      final result = await getPodLogs(jsonDecode(params));
      return Response.ok(jsonEncode(result));
    } catch (e) {
      _logger.severe('Error in get_pod_logs: $e');
      return Response.internalServerError(
        body: jsonEncode({'error': e.toString()}),
      );
    }
  }

  Future<Response> _handleCrashLoop(Request request) async {
    try {
      final params = await request.readAsString();
      final result = await detectCrashLoopPods(jsonDecode(params));
      return Response.ok(jsonEncode(result));
    } catch (e) {
      _logger.severe('Error in detect_crash_loop_pods: $e');
      return Response.internalServerError(
        body: jsonEncode({'error': e.toString()}),
      );
    }
  }

  Future<Response> _handleEvents(Request request) async {
    try {
      final params = await request.readAsString();
      final result = await summarizeEvents(jsonDecode(params));
      return Response.ok(jsonEncode(result));
    } catch (e) {
      _logger.severe('Error in summarize_events: $e');
      return Response.internalServerError(
        body: jsonEncode({'error': e.toString()}),
      );
    }
  }

  Handler get handler => _router;
}
