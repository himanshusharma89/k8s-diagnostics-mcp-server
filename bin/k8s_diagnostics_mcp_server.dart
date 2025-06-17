import 'package:k8s_diagnostics_mcp_server/k8s_diagnostics_mcp_server.dart';

void main() async {
  final server = K8sDiagnosticsMcpServer();
  await server.serve(port: 8080);
}
