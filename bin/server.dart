import 'dart:io';
import 'package:shelf/shelf.dart';
import 'package:shelf/shelf_io.dart' as shelf_io;
import 'package:logging/logging.dart';
import '../lib/server/mcp_server.dart';

void main() async {
  // Configure logging
  Logger.root.level = Level.ALL;
  Logger.root.onRecord.listen((record) {
    print('${record.level.name}: ${record.time}: ${record.message}');
  });

  final server = MCP();
  final handler =
      const Pipeline().addMiddleware(logRequests()).addHandler(server.handler);

  final port = int.parse(Platform.environment['PORT'] ?? '8080');
  final ip = InternetAddress.anyIPv4;

  final serverInstance = await shelf_io.serve(handler, ip, port);
  print('Server listening on port ${serverInstance.port}');
}
