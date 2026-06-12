import 'package:dio/dio.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/auth_repository.dart';

class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _nameCtrl = TextEditingController();
  String? _gender;
  bool _loading = false;
  String? _error;

  final _genders = [
    ('female', '女性'),
    ('male', '男性'),
    ('neutral', 'その他 / 未選択'),
  ];

  @override
  void dispose() {
    _nameCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final name = _nameCtrl.text.trim();
    if (name.isEmpty) {
      setState(() => _error = '名前を入力してください');
      return;
    }

    // Firebase ログインが完了していないと backend への登録ができない
    final firebaseUser = FirebaseAuth.instance.currentUser;
    if (firebaseUser == null) {
      setState(() => _error =
          'Googleアカウントでのサインインが完了していません。\n'
          'ログイン画面に戻って「Googleでログイン」を先に行ってください。');
      return;
    }

    setState(() { _loading = true; _error = null; });
    try {
      final token = await firebaseUser.getIdToken(true); // force refresh
      if (token == null) throw Exception('Firebase ID token の取得に失敗しました');

      // POST /auth/register — 新規ユーザーを作成（冪等: すでに存在する場合は 409）
      await ref.read(authRepositoryProvider).register(
        firebaseToken: token,
        displayName: name,
        gender: _gender,
      );
      // Router の authStateChanges が反応してホーム画面へ遷移する
    } on DioException catch (e) {
      final status = e.response?.statusCode;
      if (status == 409) {
        // すでに登録済み → 名前だけ更新して続行
        try {
          await ref.read(authRepositoryProvider).updateMe(
            displayName: name,
            gender: _gender,
          );
          // updateMe 成功 → router が home へ遷移
          return;
        } catch (updateErr) {
          if (mounted) {
            setState(() => _error =
                'プロフィール更新に失敗しました: ${updateErr.runtimeType}: $updateErr');
          }
          return;
        }
      }
      if (mounted) {
        setState(() => _error =
            'プロフィール設定に失敗しました [$status]\n'
            '原因: ${e.response?.data ?? e.message}');
      }
    } catch (e) {
      if (mounted) {
        setState(() => _error =
            'プロフィール設定に失敗しました\n${e.runtimeType}: $e');
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final firebaseUser = FirebaseAuth.instance.currentUser;

    return Scaffold(
      appBar: AppBar(title: const Text('プロフィール設定')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text('ようこそ！', style: theme.textTheme.headlineSmall),
              const SizedBox(height: 8),
              Text(
                '始める前に基本情報を設定しましょう',
                style: theme.textTheme.bodyMedium
                    ?.copyWith(color: theme.colorScheme.outline),
              ),

              // Firebase 未ログイン時の案内バナー
              if (firebaseUser == null) ...[
                const SizedBox(height: 16),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.tertiaryContainer,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.info_outline,
                          color: theme.colorScheme.onTertiaryContainer),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          'まず「Googleでログイン」でサインインしてください。\nサインイン完了後にこの画面が表示されます。',
                          style: TextStyle(
                              color: theme.colorScheme.onTertiaryContainer),
                        ),
                      ),
                    ],
                  ),
                ),
              ],

              const SizedBox(height: 32),
              // Name
              TextFormField(
                controller: _nameCtrl,
                enabled: firebaseUser != null,
                decoration: const InputDecoration(
                  labelText: '表示名',
                  hintText: '田中 美優',
                  prefixIcon: Icon(Icons.person_outline),
                ),
                textInputAction: TextInputAction.done,
              ),
              const SizedBox(height: 24),
              // Gender
              Text('性別（任意）', style: theme.textTheme.labelLarge),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                children: _genders.map((g) {
                  final (value, label) = g;
                  return ChoiceChip(
                    label: Text(label),
                    selected: _gender == value,
                    onSelected: firebaseUser == null
                        ? null
                        : (_) => setState(() => _gender = value),
                  );
                }).toList(),
              ),
              const SizedBox(height: 32),
              if (_error != null) ...[
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.errorContainer,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    _error!,
                    style:
                        TextStyle(color: theme.colorScheme.onErrorContainer),
                  ),
                ),
                const SizedBox(height: 16),
              ],
              FilledButton(
                onPressed: (firebaseUser == null || _loading) ? null : _submit,
                style: FilledButton.styleFrom(
                    minimumSize: const Size.fromHeight(52)),
                child: _loading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                            strokeWidth: 2, color: Colors.white))
                    : const Text('はじめる'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
