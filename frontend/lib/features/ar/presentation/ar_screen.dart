import 'package:flutter/material.dart';

/// AR試着画面
/// DeepAR SDK は公式 Flutter プラグインが存在しないため、
/// Phase 3 で Method Channel を使ったネイティブブリッジを実装予定。
/// 現時点では UI プレースホルダーを表示する。
class ARScreen extends StatelessWidget {
  const ARScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(title: const Text('AR試着')),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Container(
                width: 120,
                height: 120,
                decoration: BoxDecoration(
                  color: theme.colorScheme.primaryContainer,
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  Icons.camera_alt_outlined,
                  size: 60,
                  color: theme.colorScheme.onPrimaryContainer,
                ),
              ),
              const SizedBox(height: 24),
              Text('AR試着', style: theme.textTheme.headlineSmall),
              const SizedBox(height: 12),
              Text(
                'リアルタイムAR試着機能は Phase 3 で実装予定です。\n'
                'DeepAR SDK との Method Channel ブリッジを実装後、\n'
                '爪の自動計測も有効になります。',
                textAlign: TextAlign.center,
                style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.outline),
              ),
              const SizedBox(height: 32),
              OutlinedButton.icon(
                onPressed: () => Navigator.pop(context),
                icon: const Icon(Icons.arrow_back),
                label: const Text('戻る'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
