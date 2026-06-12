import 'package:freezed_annotation/freezed_annotation.dart';

part 'user.freezed.dart';
part 'user.g.dart';

@freezed
class AppUser with _$AppUser {
  const factory AppUser({
    required String id,
    required String displayName,
    required String role,
    String? avatarUrl,
    String? gender,
    @Default([]) List<String> lifestyleTags,
    @Default(0) int pointBalance,
  }) = _AppUser;

  factory AppUser.fromJson(Map<String, dynamic> json) => _$AppUserFromJson(json);
}

@freezed
class NailProfile with _$NailProfile {
  const factory NailProfile({
    String? nailShape,
    double? avgNailLengthMm,
    String? gelLiftTendency,
    String? allergyNotes,
  }) = _NailProfile;

  factory NailProfile.fromJson(Map<String, dynamic> json) => _$NailProfileFromJson(json);
}
