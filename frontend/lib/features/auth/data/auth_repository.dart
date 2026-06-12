import 'package:dio/dio.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../../core/api/client.dart';
import '../domain/user.dart';

part 'auth_repository.g.dart';

@riverpod
AuthRepository authRepository(AuthRepositoryRef ref) {
  return AuthRepository(ref.watch(apiClientProvider));
}

@riverpod
Stream<User?> authStateChanges(AuthStateChangesRef ref) {
  return FirebaseAuth.instance.authStateChanges();
}

@riverpod
Future<AppUser?> currentUser(CurrentUserRef ref) async {
  final authState = await ref.watch(authStateChangesProvider.future);
  if (authState == null) return null;
  return ref.watch(authRepositoryProvider).getMe();
}

class AuthRepository {
  AuthRepository(this._dio);
  final Dio _dio;

  Future<AppUser> register({
    required String firebaseToken,
    required String displayName,
    String? gender,
  }) async {
    final res = await _dio.post('/auth/register', data: {
      'firebase_token': firebaseToken,
      'display_name': displayName,
      if (gender != null) 'gender': gender,
    });
    return AppUser.fromJson(res.data['user'] as Map<String, dynamic>);
  }

  Future<AppUser> getMe() async {
    final res = await _dio.get('/users/me');
    return AppUser.fromJson(res.data['user'] as Map<String, dynamic>);
  }

  Future<AppUser> updateMe({
    String? displayName,
    String? avatarUrl,
    String? gender,
    List<String>? lifestyleTags,
  }) async {
    final res = await _dio.patch('/users/me', data: {
      if (displayName != null) 'display_name': displayName,
      if (avatarUrl != null) 'avatar_url': avatarUrl,
      if (gender != null) 'gender': gender,
      if (lifestyleTags != null) 'lifestyle_tags': lifestyleTags,
    });
    return AppUser.fromJson(res.data['user'] as Map<String, dynamic>);
  }

  Future<AppUser> getUserById(String userId) async {
    final res = await _dio.get('/users/$userId');
    return AppUser.fromJson(res.data['user'] as Map<String, dynamic>);
  }

  /// Google Sign-In.
  /// Web: signInWithPopup (popup window)
  /// Android/iOS: signInWithProvider (native federated sign-in)
  /// Desktop: signInWithProvider (throws UnimplementedError if not supported —
  ///          run the app with `flutter run -d web-server` in the container instead)
  Future<void> signInWithGoogle() async {
    final provider = GoogleAuthProvider();
    if (kIsWeb) {
      await FirebaseAuth.instance.signInWithPopup(provider);
    } else {
      await FirebaseAuth.instance.signInWithProvider(provider);
    }
  }

  Future<void> signInWithApple() async {
    final provider = AppleAuthProvider();
    if (kIsWeb) {
      await FirebaseAuth.instance.signInWithPopup(provider);
    } else {
      await FirebaseAuth.instance.signInWithProvider(provider);
    }
  }

  Future<void> signOut() async {
    await FirebaseAuth.instance.signOut();
  }
}
