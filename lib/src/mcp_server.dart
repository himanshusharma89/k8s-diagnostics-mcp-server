import 'dart:async';
import 'dart:convert';
import 'package:shelf/shelf.dart';

typedef ToolHandler = Future<Map<String, dynamic>> Function(
    Map<String, dynamic> params);

class McpServer {
  final Map<String, ToolHandler> _tools = {};
  final _controller = StreamController<Map<String, dynamic>>.broadcast();

  McpServer();

  void registerTool(String name, ToolHandler handler) {
    _tools[name] = handler;
  }

  Handler get handler => (Request request) async {
        if (request.method != 'POST') {
          return Response(405, body: 'Method not allowed');
        }

        try {
          final body = await request.readAsString();
          final data = json.decode(body) as Map<String, dynamic>;

          final toolName = data['tool'] as String?;
          if (toolName == null) {
            return Response(400, body: 'Missing tool name');
          }

          final handler = _tools[toolName];
          if (handler == null) {
            return Response(404, body: 'Tool not found: $toolName');
          }

          final params = data['params'] as Map<String, dynamic>? ?? {};
          final result = await handler(params);

          return Response(
            200,
            headers: {'content-type': 'application/json'},
            body: json.encode(result),
          );
        } catch (e) {
          return Response(
            500,
            headers: {'content-type': 'application/json'},
            body: json.encode({
              'error': e.toString(),
            }),
          );
        }
      };

  void dispose() {
    _controller.close();
  }
}
