import 'package:dio/dio.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'client.g.dart';

const _baseUrl = String.fromEnvironment('API_BASE_URL', defaultValue: 'http://localhost:8080/api/v1');

@riverpod
Dio apiClient(ApiClientRef ref) {
  final dio = Dio(BaseOptions(
    baseUrl: _baseUrl,
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 30),
    headers: {'Content-Type': 'application/json'},
  ));

  dio.interceptors.add(_AuthInterceptor());
  dio.interceptors.add(LogInterceptor(
    request: false,
    responseBody: false,
    error: true,
  ));

  return dio;
}

class _AuthInterceptor extends Interceptor {
  @override
  Future<void> onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    final user = FirebaseAuth.instance.currentUser;
    if (user != null) {
      final token = await user.getIdToken();
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    if (err.response?.statusCode == 401) {
      // TODO: 認証切れ → ログイン画面へリダイレクト
    }
    handler.next(err);
  }
}

/// SSE 接続用の生 Dio インスタンス（タイムアウト無し）
Dio sseClient() {
  final dio = Dio(BaseOptions(
    baseUrl: _baseUrl,
    receiveTimeout: Duration.zero, // SSE は接続を維持し続ける
    headers: {
      'Accept': 'text/event-stream',
      'Cache-Control': 'no-cache',
    },
  ));
  dio.interceptors.add(_AuthInterceptor());
  return dio;
}
