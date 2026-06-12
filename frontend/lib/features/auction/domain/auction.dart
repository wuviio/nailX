import 'package:freezed_annotation/freezed_annotation.dart';

part 'auction.freezed.dart';
part 'auction.g.dart';

@freezed
class BookingRequest with _$BookingRequest {
  const factory BookingRequest({
    required String id,
    required String designIpId,
    required String status,
    required int budgetMaxYen,
    required String desiredDateFrom,
    required String desiredDateTo,
    required String areaPrefecture,
    String? areaCity,
    String? arSessionId,
    NailDataSnapshot? nailDataSnapshot,
    String? expiresAt,
    String? createdAt,
  }) = _BookingRequest;

  factory BookingRequest.fromJson(Map<String, dynamic> json) =>
      _$BookingRequestFromJson(json);
}

@freezed
class NailDataSnapshot with _$NailDataSnapshot {
  const factory NailDataSnapshot({
    double? lengthMm,
    @Default(false) bool hasExistingGel,
    String? shape,
    int? estimatedTreatmentMin,
    double? estimatedGelAmountMl,
  }) = _NailDataSnapshot;

  factory NailDataSnapshot.fromJson(Map<String, dynamic> json) =>
      _$NailDataSnapshotFromJson(json);
}

@freezed
class Bid with _$Bid {
  const factory Bid({
    required String id,
    required String requestId,
    required String salonId,
    String? salonName,
    String? salonAvatarUrl,
    required int priceYen,
    @Default(false) bool includesRemoval,
    @Default(0) int removalFeeYen,
    required String availableSlotAt,
    String? dynamicDiscountReason,
    String? message,
    required String status,
    String? expiresAt,
    double? salonReproductionScore,
    String? createdAt,
  }) = _Bid;

  factory Bid.fromJson(Map<String, dynamic> json) => _$BidFromJson(json);
}

@freezed
class BookingRequestWithBids with _$BookingRequestWithBids {
  const factory BookingRequestWithBids({
    required BookingRequest request,
    @Default([]) List<Bid> bids,
  }) = _BookingRequestWithBids;

  factory BookingRequestWithBids.fromJson(Map<String, dynamic> json) =>
      _$BookingRequestWithBidsFromJson(json);
}
