import 'package:dio/dio.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../../core/api/client.dart';
import '../domain/booking.dart';

part 'booking_repository.g.dart';

@riverpod
BookingRepository bookingRepository(BookingRepositoryRef ref) {
  return BookingRepository(ref.watch(apiClientProvider));
}

@riverpod
Future<List<Booking>> myBookings(MyBookingsRef ref, {String? status}) async {
  return ref.watch(bookingRepositoryProvider).listBookings(status: status);
}

@riverpod
Future<Booking> bookingDetail(BookingDetailRef ref, String id) async {
  return ref.watch(bookingRepositoryProvider).getBooking(id);
}

class BookingRepository {
  BookingRepository(this._dio);
  final Dio _dio;

  Future<({Booking booking, String stripeClientSecret})> createBooking({
    required String bidId,
    required String paymentMethodId,
  }) async {
    final res = await _dio.post('/bookings', data: {
      'bid_id': bidId,
      'payment_method_id': paymentMethodId,
    });
    return (
      booking: Booking.fromJson(res.data['booking'] as Map<String, dynamic>),
      stripeClientSecret: res.data['stripe_client_secret'] as String,
    );
  }

  Future<List<Booking>> listBookings({String? status, String? cursor}) async {
    final res = await _dio.get('/bookings', queryParameters: {
      if (status != null) 'status': status,
      if (cursor != null) 'cursor': cursor,
      'limit': 20,
    });
    return (res.data['bookings'] as List)
        .map((e) => Booking.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Booking> getBooking(String id) async {
    final res = await _dio.get('/bookings/$id');
    return Booking.fromJson(res.data['booking'] as Map<String, dynamic>);
  }

  Future<void> cancelBooking(String id, {String? reason}) async {
    await _dio.post('/bookings/$id/cancel',
        data: {if (reason != null) 'reason': reason});
  }

  Future<Review> postReview({
    required String bookingId,
    required int reproductionScore,
    required int overallScore,
    String? comment,
    String? beforePhotoUrl,
    String? afterPhotoUrl,
  }) async {
    final res = await _dio.post('/reviews', data: {
      'booking_id': bookingId,
      'reproduction_score': reproductionScore,
      'overall_score': overallScore,
      if (comment != null) 'comment': comment,
      if (beforePhotoUrl != null) 'before_photo_url': beforePhotoUrl,
      if (afterPhotoUrl != null) 'after_photo_url': afterPhotoUrl,
    });
    return Review.fromJson(res.data['review'] as Map<String, dynamic>);
  }
}
