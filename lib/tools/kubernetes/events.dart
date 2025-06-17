import 'package:k8s/k8s.dart';
import 'package:logging/logging.dart';

final _logger = Logger('Events');

Future<Map<String, dynamic>> summarizeEvents(
    Map<String, dynamic> params) async {
  try {
    final client = K8sClient();
    final namespace = params['namespace'] as String?;
    final duration = params['duration'] as String? ?? '1h';

    final events = await client.listEvents(namespace: namespace);
    final now = DateTime.now();

    // Filter events based on duration
    final filteredEvents = events.where((event) {
      final eventTime = event.lastTimestamp ?? event.firstTimestamp;
      if (eventTime == null) return false;

      final difference = now.difference(eventTime);
      return difference.inHours <= int.parse(duration.replaceAll('h', ''));
    }).toList();

    // Group events by type and reason
    final eventSummary = <String, Map<String, int>>{};
    for (final event in filteredEvents) {
      final type = event.type ?? 'Unknown';
      final reason = event.reason ?? 'Unknown';

      eventSummary.putIfAbsent(type, () => {});
      eventSummary[type]!
          .update(reason, (count) => count + 1, ifAbsent: () => 1);
    }

    // Get most recent events
    final recentEvents = filteredEvents
        .take(10)
        .map((e) => {
              'type': e.type,
              'reason': e.reason,
              'message': e.message,
              'timestamp': e.lastTimestamp?.toIso8601String(),
              'involved_object': {
                'kind': e.involvedObject?.kind,
                'name': e.involvedObject?.name,
                'namespace': e.involvedObject?.namespace
              }
            })
        .toList();

    return {
      'status': 'success',
      'data': {
        'event_summary': eventSummary,
        'recent_events': recentEvents,
        'total_events': filteredEvents.length,
        'duration': duration,
        'timestamp': now.toIso8601String()
      }
    };
  } catch (e) {
    _logger.severe('Error summarizing events: $e');
    return {'status': 'error', 'error': e.toString()};
  }
}
