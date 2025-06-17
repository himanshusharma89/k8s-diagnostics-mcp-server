// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'k8s_models.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

K8sClient _$K8sClientFromJson(Map<String, dynamic> json) => K8sClient();

Map<String, dynamic> _$K8sClientToJson(K8sClient instance) =>
    <String, dynamic>{};

Node _$NodeFromJson(Map<String, dynamic> json) => Node(
      metadata: Metadata.fromJson(json['metadata'] as Map<String, dynamic>),
      status: NodeStatus.fromJson(json['status'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$NodeToJson(Node instance) => <String, dynamic>{
      'metadata': instance.metadata,
      'status': instance.status,
    };

Pod _$PodFromJson(Map<String, dynamic> json) => Pod(
      metadata: Metadata.fromJson(json['metadata'] as Map<String, dynamic>),
      status: PodStatus.fromJson(json['status'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$PodToJson(Pod instance) => <String, dynamic>{
      'metadata': instance.metadata,
      'status': instance.status,
    };

Event _$EventFromJson(Map<String, dynamic> json) => Event(
      type: json['type'] as String,
      reason: json['reason'] as String,
      message: json['message'] as String,
      lastTimestamp: json['lastTimestamp'] == null
          ? null
          : DateTime.parse(json['lastTimestamp'] as String),
    );

Map<String, dynamic> _$EventToJson(Event instance) => <String, dynamic>{
      'type': instance.type,
      'reason': instance.reason,
      'message': instance.message,
      'lastTimestamp': instance.lastTimestamp?.toIso8601String(),
    };

Metadata _$MetadataFromJson(Map<String, dynamic> json) => Metadata(
      name: json['name'] as String,
      namespace: json['namespace'] as String,
    );

Map<String, dynamic> _$MetadataToJson(Metadata instance) => <String, dynamic>{
      'name': instance.name,
      'namespace': instance.namespace,
    };

NodeStatus _$NodeStatusFromJson(Map<String, dynamic> json) => NodeStatus(
      conditions: (json['conditions'] as List<dynamic>)
          .map((e) => NodeCondition.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$NodeStatusToJson(NodeStatus instance) =>
    <String, dynamic>{
      'conditions': instance.conditions,
    };

PodStatus _$PodStatusFromJson(Map<String, dynamic> json) => PodStatus(
      containerStatuses: (json['containerStatuses'] as List<dynamic>?)
          ?.map((e) => ContainerStatus.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$PodStatusToJson(PodStatus instance) => <String, dynamic>{
      'containerStatuses': instance.containerStatuses,
    };

NodeCondition _$NodeConditionFromJson(Map<String, dynamic> json) =>
    NodeCondition(
      type: json['type'] as String,
      status: json['status'] as String,
    );

Map<String, dynamic> _$NodeConditionToJson(NodeCondition instance) =>
    <String, dynamic>{
      'type': instance.type,
      'status': instance.status,
    };

ContainerStatus _$ContainerStatusFromJson(Map<String, dynamic> json) =>
    ContainerStatus(
      restartCount: (json['restartCount'] as num).toInt(),
    );

Map<String, dynamic> _$ContainerStatusToJson(ContainerStatus instance) =>
    <String, dynamic>{
      'restartCount': instance.restartCount,
    };
