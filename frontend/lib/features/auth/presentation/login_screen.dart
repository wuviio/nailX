import 'package:dio/dio.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/router/router.dart';
import '../data/auth_repository.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  bool _loading = false;
  String? _error;

  Future<void> _signIn(Future<void> Function() providerSignIn) async {
    if (!kIsWeb) {
      // Desktop platform: Firebase social sign-in is not supported.
      // Run the app via `flutter run -d web-server` in the Dev Container.
      setState(() => _error =
          'このプラットフォームではソーシャルログインに未対応です。\n'
          'Dev Container 内で flutter run -d web-server を実行し、\n'
          'ブラウザで http://localhost:3000 を開いてください。');
      return;
    }
    setState(() { _loading = true; _error = null; });
    try {
      await providerSignIn();

      // Firebase auth succeeded — check backend user exists.
      // If not registered yet, navigate to register screen.
      try {
        await ref.read(authRepositoryProvider).getMe();
        // Backend user exists → router redirect handles navigation to home
      } on DioException catch (e) {
        if ((e.response?.statusCode == 404 || e.response?.statusCode == 401) && mounted) {
          // New user: backend record not created yet
          context.go(AppRoutes.register);
        } else {
          rethrow;
        }
      }
    } on FirebaseAuthException catch (e) {
      setState(() => _error = '[${e.code}] ${e.message ?? 'ログインに失敗しました'}');
    } on UnimplementedError catch (e) {
      setState(() => _error =
          'このプラットフォームは未対応です。\n'
          'Dev Container 内で web-server として起動してください。\n'
          '詳細: $e');
    } catch (e) {
      setState(() => _error = '${e.runtimeType}: $e');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _signInWithGoogle() =>
      _signIn(ref.read(authRepositoryProvider).signInWithGoogle);

  Future<void> _signInWithApple() =>
      _signIn(ref.read(authRepositoryProvider).signInWithApple);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Spacer(flex: 2),
              // Logo
              Center(
                child: Column(
                  children: [
                    Container(
                      width: 80,
                      height: 80,
                      decoration: BoxDecoration(
                        color: theme.colorScheme.primary,
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: const Icon(Icons.auto_awesome,
                          color: Colors.white, size: 40),
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'nailX',
                      style: theme.textTheme.headlineLarge?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '次世代UGCネイルプラットフォーム',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.outline,
                      ),
                    ),
                  ],
                ),
              ),
              const Spacer(flex: 2),
              // Error message
              if (_error != null) ...[
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.errorContainer,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    _error!,
                    style: TextStyle(color: theme.colorScheme.onErrorContainer),
                  ),
                ),
                const SizedBox(height: 16),
              ],
              // Google Sign-In
              _SocialLoginButton(
                onPressed: _loading ? null : _signInWithGoogle,
                icon: const Icon(Icons.g_mobiledata, size: 24),
                label: 'Googleでログイン',
                loading: _loading,
              ),
              const SizedBox(height: 12),
              // Apple Sign-In
              _SocialLoginButton(
                onPressed: _loading ? null : _signInWithApple,
                icon: const Icon(Icons.apple, size: 24),
                label: 'Appleでログイン',
                loading: false,
                filled: true,
              ),
              const SizedBox(height: 24),
              // New user note: sign-up uses the same Google/Apple buttons above
              Center(
                child: Text(
                  '初めての方も上のボタンでアカウント作成できます',
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Theme.of(context).colorScheme.outline,
                  ),
                ),
              ),
              const Spacer(),
              Text(
                'ログインすることで利用規約・プライバシーポリシーに同意したことになります',
                textAlign: TextAlign.center,
                style: theme.textTheme.labelSmall
                    ?.copyWith(color: theme.colorScheme.outline),
              ),
              const SizedBox(height: 24),
            ],
          ),
        ),
      ),
    );
  }
}

class _SocialLoginButton extends StatelessWidget {
  const _SocialLoginButton({
    required this.onPressed,
    required this.icon,
    required this.label,
    this.loading = false,
    this.filled = false,
  });

  final VoidCallback? onPressed;
  final Widget icon;
  final String label;
  final bool loading;
  final bool filled;

  @override
  Widget build(BuildContext context) {
    final child = loading
        ? const SizedBox(
            width: 20,
            height: 20,
            child: CircularProgressIndicator(strokeWidth: 2))
        : Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [icon, const SizedBox(width: 12), Text(label)],
          );

    if (filled) {
      return FilledButton(
        onPressed: onPressed,
        style: FilledButton.styleFrom(
          backgroundColor: Colors.black,
          foregroundColor: Colors.white,
          minimumSize: const Size.fromHeight(52),
        ),
        child: child,
      );
    }
    return OutlinedButton(
      onPressed: onPressed,
      style: OutlinedButton.styleFrom(minimumSize: const Size.fromHeight(52)),
      child: child,
    );
  }
}
