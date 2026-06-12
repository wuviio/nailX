import 'package:dio/dio.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../../core/api/client.dart';
import '../domain/auction.dart';

part 'auction_repository.g.dart';

@riverpod
AuctionRepository auctionRepository(AuctionRepositoryRef ref) {
  return AuctionRepository(ref.watch(apiClientProvider));
}

@riverpod
Future<BookingRequestWithBids> bookingRequestDetail(
  BookingRequestDetailRef ref,
  String requestId,
) async {
  return ref.watch(auctionRepositoryProvider).getRequestDetail(requestId);
}

@riverpod
Future<List<BookingRequest>> openRequests(OpenRequestsRef ref) async {
  return ref.watch(auctionRepositoryProvider).listOpenRequests();
}

class AuctionRepository {
  AuctionRepository(this._dio);
  final Dio _dio;

  Future<BookingRequest> createRequest({
    required String designIpId,
    String? arSessionId,
    NailDataSnapshot? nailDataSnapshot,
    required int budgetMaxYen,
    required String desiredDateFrom,
    required String desiredDateTo,
    required String areaPrefecture,
    String? areaCity,
  }) async {
    final res = await _dio.post('/auctions/requests', data: {
      'design_ip_id': designIpId,
      if (arSessionId != null) 'ar_session_id': arSessionId,
      if (nailDataSnapshot != null)
        'nail_data_snapshot': nailDataSnapshot.toJson(),
      'budget_max_yen': budgetMaxYen,
      'desired_date_from': desiredDateFrom,
      'desired_date_to': desiredDateTo,
      'area_prefecture': areaPrefecture,
      if (areaCity != null) 'area_city': areaCity,
    });
    return BookingRequest.fromJson(
        res.data['booking_request'] as Map<String, dynamic>);
  }

  Future<BookingRequestWithBids> getRequestDetail(String requestId) async {
    final res = await _dio.get('/auctions/requests/$requestId');
    return BookingRequestWithBids.fromJson(res.data as Map<String, dynamic>);
  }

  Future<void> cancelRequest(String requestId) async {
    await _dio.delete('/auctions/requests/$requestId');
  }

  Future<List<BookingRequest>> listOpenRequests({String? cursor}) async {
    final res = await _dio.get('/auctions/requests', queryParameters: {
      'sort': 'expires_at',
      'limit': 20,
      if (cursor != null) 'cursor': cursor,
    });
    return (res.data['requests'] as List)
        .map((e) => BookingRequest.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<Bid>> listBids(String requestId, {String? sort}) async {
    final res = await _dio.get(
      '/auctions/requests/$requestId/bids',
      queryParameters: {if (sort != null) 'sort': sort},
    );
    return (res.data['bids'] as List)
        .map((e) => Bid.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Bid> placeBid({
    required String requestId,
    required int priceYen,
    required bool includesRemoval,
    int removalFeeYen = 0,
    required String availableSlotAt,
    String? dynamicDiscountReason,
    String? message,
  }) async {
    final res = await _dio.post('/auctions/requests/$requestId/bids', data: {
      'price_yen': priceYen,
      'includes_removal': includesRemoval,
      'removal_fee_yen': removalFeeYen,
      'available_slot_at': availableSlotAt,
      if (dynamicDiscountReason != null)
        'dynamic_discount_reason': dynamicDiscountReason,
      if (message != null) 'message': message,
    });
    return Bid.fromJson(res.data['bid'] as Map<String, dynamic>);
  }
}
