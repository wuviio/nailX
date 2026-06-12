import 'package:freezed_annotation/freezed_annotation.dart';

part 'design.freezed.dart';
part 'design.g.dart';

@freezed
class DesignIP with _$DesignIP {
  const factory DesignIP({
    required String id,
    required String title,
    required String previewImageUrl,
    String? description,
    String? creatorId,
    String? creatorName,
    String? parentIpId,
    String? genderTag,
    @Default([]) List<String> styleTags,
    @Default('active') String status,
    @Default(0) int usageCount,
    @Default(true) bool isPublic,
    String? createdAt,
  }) = _DesignIP;

  factory DesignIP.fromJson(Map<String, dynamic> json) => _$DesignIPFromJson(json);
}

@freezed
class RoyaltyNode with _$RoyaltyNode {
  const factory RoyaltyNode({
    required String userId,
    required double sharePercent,
    required int depthLevel,
  }) = _RoyaltyNode;

  factory RoyaltyNode.fromJson(Map<String, dynamic> json) => _$RoyaltyNodeFromJson(json);
}

@freezed
class DesignDetail with _$DesignDetail {
  const factory DesignDetail({
    required DesignIP design,
    AppCreator? creator,
    DesignIP? parentIp,
    @Default([]) List<RoyaltyNode> royaltyNodes,
  }) = _DesignDetail;

  factory DesignDetail.fromJson(Map<String, dynamic> json) => _$DesignDetailFromJson(json);
}

@freezed
class AppCreator with _$AppCreator {
  const factory AppCreator({
    required String id,
    required String displayName,
    String? avatarUrl,
  }) = _AppCreator;

  factory AppCreator.fromJson(Map<String, dynamic> json) => _$AppCreatorFromJson(json);
}

@freezed
class DesignFeedFilter with _$DesignFeedFilter {
  const factory DesignFeedFilter({
    String? genderTag,
    String? styleTag,
    @Default('latest') String sort,
    String? cursor,
    @Default(20) int limit,
  }) = _DesignFeedFilter;
}
