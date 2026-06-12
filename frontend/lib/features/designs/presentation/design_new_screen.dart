import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../data/design_repository.dart';

class DesignNewScreen extends ConsumerStatefulWidget {
  const DesignNewScreen({super.key});

  @override
  ConsumerState<DesignNewScreen> createState() => _DesignNewScreenState();
}

class _DesignNewScreenState extends ConsumerState<DesignNewScreen> {
  final _titleCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  String _genderTag = 'neutral';
  final List<String> _styleTags = [];
  bool _isPublic = true;
  bool _loading = false;
  String? _error;

  final _availableStyleTags = [
    'simple',
    'mirror',
    '3d_parts',
    'art',
    'french',
    'gel',
    'mens',
    'seasonal',
    'floral',
    'abstract',
  ];

  @override
  void dispose() {
    _titleCtrl.dispose();
    _descCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final title = _titleCtrl.text.trim();
    if (title.isEmpty) {
      setState(() => _error = 'タイトルを入力してください');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref.read(designRepositoryProvider).createDesign(
            title: title,
            description:
                _descCtrl.text.trim().isEmpty ? null : _descCtrl.text.trim(),
            previewImageUrl:
                'https://cdn.nailx.jp/designs/preview/placeholder.jpg', // TODO: real upload
            designData: {'fingers': {}}, // TODO: 3D editor data
            genderTag: _genderTag,
            styleTags: _styleTags,
            isPublic: _isPublic,
          );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('デザインIP申請を送信しました。審査中...')),
        );
        context.pop();
      }
    } catch (e) {
      setState(() => _error = 'デザインの登録に失敗しました');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(
        title: const Text('デザイン作成'),
        actions: [
          TextButton(
            onPressed: _loading ? null : _submit,
            child: const Text('申請'),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // 3D editor placeholder
            Container(
              height: 240,
              decoration: BoxDecoration(
                color: theme.colorScheme.surfaceContainerHighest,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.view_in_ar,
                      size: 48, color: theme.colorScheme.outline),
                  const SizedBox(height: 8),
                  Text('3Dエディター', style: theme.textTheme.titleMedium),
                  Text('（Phase 2 で実装予定）',
                      style: theme.textTheme.bodySmall
                          ?.copyWith(color: theme.colorScheme.outline)),
                ],
              ),
            ),
            const SizedBox(height: 24),
            TextFormField(
              controller: _titleCtrl,
              decoration: const InputDecoration(
                  labelText: 'タイトル *', hintText: 'シンプルミラーネイル'),
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: _descCtrl,
              decoration: const InputDecoration(labelText: '説明（任意）'),
              maxLines: 3,
            ),
            const SizedBox(height: 24),
            Text('性別タグ', style: theme.textTheme.labelLarge),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              children: [
                ('feminine', '女性向け'),
                ('masculine', '男性向け'),
                ('neutral', 'ユニセックス'),
              ].map((g) {
                final (value, label) = g;
                return ChoiceChip(
                  label: Text(label),
                  selected: _genderTag == value,
                  onSelected: (_) => setState(() => _genderTag = value),
                );
              }).toList(),
            ),
            const SizedBox(height: 24),
            Text('スタイルタグ', style: theme.textTheme.labelLarge),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              children: _availableStyleTags.map((tag) {
                return FilterChip(
                  label: Text(tag),
                  selected: _styleTags.contains(tag),
                  onSelected: (selected) {
                    setState(() {
                      if (selected) {
                        _styleTags.add(tag);
                      } else {
                        _styleTags.remove(tag);
                      }
                    });
                  },
                );
              }).toList(),
            ),
            const SizedBox(height: 24),
            SwitchListTile(
              title: const Text('公開する'),
              subtitle: const Text('オフにすると自分だけ閲覧可能'),
              value: _isPublic,
              onChanged: (v) => setState(() => _isPublic = v),
            ),
            if (_error != null) ...[
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.errorContainer,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(_error!,
                    style:
                        TextStyle(color: theme.colorScheme.onErrorContainer)),
              ),
            ],
            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }
}
