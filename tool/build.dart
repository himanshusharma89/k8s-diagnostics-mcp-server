import 'dart:io';

void main() async {
  final result = await Process.run('dart', ['run', 'build_runner', 'build']);
  print(result.stdout);
  if (result.stderr.isNotEmpty) {
    print(result.stderr);
  }
  exit(result.exitCode);
}
