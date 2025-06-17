import 'package:json_annotation/json_annotation.dart';

part 'k8s_models.g.dart';

@JsonSerializable()
class K8sClient {
  K8sClient();

  Future<List<Node>> listNodes() async {
    // TODO: Implement actual Kubernetes API call
    return [];
  }

  Future<List<Pod>> listPods() async {
    // TODO: Implement actual Kubernetes API call
    return [];
  }

  Future<String> getPodLogs(String namespace, String podName) async {
    // TODO: Implement actual Kubernetes API call
    return '';
  }

  Future<List<Event>> listEvents() async {
    // TODO: Implement actual Kubernetes API call
    return [];
  }
}

@JsonSerializable()
class Node {
  final Metadata metadata;
  final NodeStatus status;

  Node({required this.metadata, required this.status});

  factory Node.fromJson(Map<String, dynamic> json) => _$NodeFromJson(json);
  Map<String, dynamic> toJson() => _$NodeToJson(this);
}

@JsonSerializable()
class Pod {
  final Metadata metadata;
  final PodStatus status;

  Pod({required this.metadata, required this.status});

  factory Pod.fromJson(Map<String, dynamic> json) => _$PodFromJson(json);
  Map<String, dynamic> toJson() => _$PodToJson(this);
}

@JsonSerializable()
class Event {
  final String type;
  final String reason;
  final String message;
  @JsonKey(name: 'lastTimestamp')
  final DateTime? lastTimestamp;

  Event({
    required this.type,
    required this.reason,
    required this.message,
    this.lastTimestamp,
  });

  factory Event.fromJson(Map<String, dynamic> json) => _$EventFromJson(json);
  Map<String, dynamic> toJson() => _$EventToJson(this);
}

@JsonSerializable()
class Metadata {
  final String name;
  final String namespace;

  Metadata({required this.name, required this.namespace});

  factory Metadata.fromJson(Map<String, dynamic> json) =>
      _$MetadataFromJson(json);
  Map<String, dynamic> toJson() => _$MetadataToJson(this);
}

@JsonSerializable()
class NodeStatus {
  final List<NodeCondition> conditions;

  NodeStatus({required this.conditions});

  factory NodeStatus.fromJson(Map<String, dynamic> json) =>
      _$NodeStatusFromJson(json);
  Map<String, dynamic> toJson() => _$NodeStatusToJson(this);
}

@JsonSerializable()
class PodStatus {
  final List<ContainerStatus>? containerStatuses;

  PodStatus({this.containerStatuses});

  factory PodStatus.fromJson(Map<String, dynamic> json) =>
      _$PodStatusFromJson(json);
  Map<String, dynamic> toJson() => _$PodStatusToJson(this);
}

@JsonSerializable()
class NodeCondition {
  final String type;
  final String status;

  NodeCondition({required this.type, required this.status});

  factory NodeCondition.fromJson(Map<String, dynamic> json) =>
      _$NodeConditionFromJson(json);
  Map<String, dynamic> toJson() => _$NodeConditionToJson(this);
}

@JsonSerializable()
class ContainerStatus {
  final int restartCount;

  ContainerStatus({required this.restartCount});

  factory ContainerStatus.fromJson(Map<String, dynamic> json) =>
      _$ContainerStatusFromJson(json);
  Map<String, dynamic> toJson() => _$ContainerStatusToJson(this);
}
