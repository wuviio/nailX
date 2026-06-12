import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../shared/widgets/loading_widget.dart';
import '../../auth/data/auth_repository.dart';

class IPWalletScreen extends ConsumerWidget {
  const IPWalletScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userAsync = ref.watch(currentUserProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('IPウォレット')),
      body: userAsync.when(
        data: (user) {
          if (user == null) return const Center(child: Text('ログインしてください'));
          return _WalletView(user: user);
        },
        loading: () => const LoadingWidget(),
        error: (_, __) => const Center(child: Text('読み込みに失敗しました')),
      ),
    );
  }
}

class _WalletView extends StatelessWidget {
  const _WalletView({required this.user});
  final dynamic user;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final balance = user.pointBalance as int;

    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Balance card
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  theme.colorScheme.primary,
                  theme.colorScheme.secondary
                ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(16),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'ポイント残高',
                  style: theme.textTheme.labelLarge
                      ?.copyWith(color: Colors.white70),
                ),
                const SizedBox(height: 8),
                Text(
                  '${NumberFormat('#,###').format(balance)} pt',
                  style: theme.textTheme.displaySmall?.copyWith(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 16),
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton(
                        onPressed: () => _showExchangeInfo(context),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.white,
                          side: const BorderSide(color: Colors.white54),
                        ),
                        child: const Text('交換申請'),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),

          // Info section
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerHighest.withOpacity(0.5),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(children: [
                  Icon(Icons.info_outline,
                      size: 16, color: theme.colorScheme.outline),
                  const SizedBox(width: 8),
                  Text('ポイントについて', style: theme.textTheme.labelMedium),
                ]),
                const SizedBox(height: 8),
                Text(
                  '・デザインIPが施術に使われるたびにロイヤリティが付与されます\n'
                  '・IPフォーク時は最大3段階の親IPにも分配されます\n'
                  '・ポイント現金化は法務確認完了後に対応予定です\n'
                  '・現在はアプリ内クーポンとして使用できます',
                  style: theme.textTheme.bodySmall,
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),

          Text('収益履歴', style: theme.textTheme.titleMedium),
          const SizedBox(height: 12),

          // Placeholder list
          ...List.generate(
              3,
              (i) => Card(
                    margin: const EdgeInsets.only(bottom: 8),
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Row(
                        children: [
                          Container(
                            width: 36,
                            height: 36,
                            decoration: BoxDecoration(
                              color: theme.colorScheme.primaryContainer,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Icon(Icons.auto_awesome,
                                size: 18,
                                color: theme.colorScheme.onPrimaryContainer),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text('ロイヤリティ付与 (Depth ${i == 0 ? 0 : i})',
                                    style: theme.textTheme.bodySmall),
                                Text('デザインID: xxxxxx',
                                    style: theme.textTheme.labelSmall?.copyWith(
                                        color: theme.colorScheme.outline)),
                              ],
                            ),
                          ),
                          Text(
                            '+${NumberFormat('#,###').format((3 - i) * 350)} pt',
                            style: theme.textTheme.bodyMedium?.copyWith(
                              color: Colors.green,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ),
                  )),

          const SizedBox(height: 8),
          Center(
            child: Text(
              '収益履歴の詳細はPhase 5（決済実装）で表示されます',
              style: theme.textTheme.labelSmall
                  ?.copyWith(color: theme.colorScheme.outline),
            ),
          ),
        ],
      ),
    );
  }

  void _showExchangeInfo(BuildContext context) {
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('ポイント交換申請'),
        content: const Text(
          'ポイントの現金化機能は現在法務確認中です。\n'
          '対応完了後にお知らせします。\n\n'
          '現時点では、ポイントをアプリ内クーポンとして使用できます。',
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('閉じる')),
        ],
      ),
    );
  }
}
