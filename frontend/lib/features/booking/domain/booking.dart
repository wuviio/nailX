import 'package:freezed_annotation/freezed_annotation.dart';

part 'booking.freezed.dart';
part 'booking.g.dart';

@freezed
class Booking with _$Booking {
  const factory Booking({
    required String id,
    required String bidId,
    required String salonId,
    String? salonName,
    String? designIpId,
    String? designTitle,
    String? designPreviewUrl,
    required String scheduledAt,
    required String status,
    int? totalAmountYen,
    String? cancelReason,
    String? createdAt,
  }) = _Booking;

  factory Booking.fromJson(Map<String, dynamic> json) => _$BookingFromJson(json);
}

@freezed
class Review with _$Review {
  const factory Review({
    required String id,
    required String bookingId,
    required int reproductionScore,
    required int overallScore,
    String? comment,
    String? beforePhotoUrl,
    String? afterPhotoUrl,
    String? createdAt,
  }) = _Review;

  factory Review.fromJson(Map<String, dynamic> json) => _$ReviewFromJson(json);
}
