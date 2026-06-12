import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../data/auction_repository.dart';

class AuctionNewScreen extends ConsumerStatefulWidget {
  const AuctionNewScreen({super.key, this.designIpId});
  final String? designIpId;

  @override
  ConsumerState<AuctionNewScreen> createState() => _AuctionNewScreenState();
}

class _AuctionNewScreenState extends ConsumerState<AuctionNewScreen> {
  int _budgetMaxYen = 8000;
  DateTimeRange? _desiredDateRange;
  String _areaPrefecture = '東京';
  String? _areaCity;
  bool _loading = false;
  String? _error;

  final _prefectures = [
    '東京',
    '大阪',
    '神奈川',
    '愛知',
    '福岡',
    '北海道',
    '京都',
    '兵庫',
    'その他'
  ];

  Future<void> _pickDateRange() async {
    final now = DateTime.now();
    final picked = await showDateRangePicker(
      context: context,
      firstDate: now.add(const Duration(days: 1)),
      lastDate: now.add(const Duration(days: 60)),
      initialDateRange: _desiredDateRange,
    );
    if (picked != null) setState(() => _desiredDateRange = picked);
  }

  Future<void> _submit() async {
    if (widget.designIpId == null) {
      setState(() => _error = 'デザインが選択されていません');
      return;
    }
    if (_desiredDateRange == null) {
      setState(() => _error = '希望日程を選択してください');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final request = await ref.read(auctionRepositoryProvider).createRequest(
            designIpId: widget.designIpId!,
            budgetMaxYen: _budgetMaxYen,
            desiredDateFrom: _desiredDateRange!.start.toIso8601String(),
            desiredDateTo: _desiredDateRange!.end.toIso8601String(),
            areaPrefecture: _areaPrefecture,
            areaCity: _areaCity,
          );
      if (mounted) {
        context.pushReplacement('/auctions/${request.id}');
      }
    } catch (e) {
      setState(() => _error = '予約リクエストの作成に失敗しました。\n同時オープン可能なリクエストは3件までです。');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final fmt = DateFormat('M/d(E)', 'ja');
    return Scaffold(
      appBar: AppBar(title: const Text('予約リクエスト作成')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Design card
            if (widget.designIpId != null)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.primaryContainer.withOpacity(0.4),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    Icon(Icons.palette_outlined,
                        color: theme.colorScheme.primary),
                    const SizedBox(width: 8),
                    Expanded(
                        child: Text('デザイン ID: ${widget.designIpId}',
                            style: theme.textTheme.bodySmall)),
                  ],
                ),
              )
            else
              OutlinedButton.icon(
                onPressed: () => context.push('/home'),
                icon: const Icon(Icons.search),
                label: const Text('デザインを選ぶ'),
              ),
            const SizedBox(height: 24),

            // Budget slider
            Text('予算上限', style: theme.textTheme.titleSmall),
            const SizedBox(height: 4),
            Text(
              '¥${NumberFormat('#,###').format(_budgetMaxYen)}',
              style: theme.textTheme.headlineSmall?.copyWith(
                  color: theme.colorScheme.primary,
                  fontWeight: FontWeight.bold),
            ),
            Slider(
              value: _budgetMaxYen.toDouble(),
              min: 3000,
              max: 30000,
              divisions: 54,
              label: '¥${NumberFormat('#,###').format(_budgetMaxYen)}',
              onChanged: (v) => setState(() => _budgetMaxYen = v.round()),
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('¥3,000', style: theme.textTheme.labelSmall),
                Text('¥30,000', style: theme.textTheme.labelSmall),
              ],
            ),
            const SizedBox(height: 24),

            // Date range
            Text('希望日程', style: theme.textTheme.titleSmall),
            const SizedBox(height: 8),
            OutlinedButton.icon(
              onPressed: _pickDateRange,
              icon: const Icon(Icons.date_range),
              label: Text(
                _desiredDateRange == null
                    ? '日程を選択'
                    : '${fmt.format(_desiredDateRange!.start)} 〜 ${fmt.format(_desiredDateRange!.end)}',
              ),
              style: OutlinedButton.styleFrom(
                  minimumSize: const Size.fromHeight(52)),
            ),
            const SizedBox(height: 24),

            // Area
            Text('エリア', style: theme.textTheme.titleSmall),
            const SizedBox(height: 8),
            DropdownButtonFormField<String>(
              value: _areaPrefecture,
              decoration: const InputDecoration(labelText: '都道府県'),
              items: _prefectures
                  .map((p) => DropdownMenuItem(value: p, child: Text(p)))
                  .toList(),
              onChanged: (v) => setState(() => _areaPrefecture = v!),
            ),
            const SizedBox(height: 12),
            TextFormField(
              decoration: const InputDecoration(
                labelText: '市区町村（任意）',
                hintText: '渋谷区',
              ),
              onChanged: (v) => _areaCity = v.isEmpty ? null : v,
            ),
            const SizedBox(height: 24),

            // Info box
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: theme.colorScheme.surfaceContainerHighest,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(children: [
                    Icon(Icons.info_outline,
                        size: 16, color: theme.colorScheme.outline),
                    const SizedBox(width: 8),
                    Text('逆オークションの仕組み', style: theme.textTheme.labelMedium),
                  ]),
                  const SizedBox(height: 8),
                  Text(
                    '・リクエストを出品すると、条件に合うサロンが入札します\n'
                    '・サロンの一発提示（チャットなし）\n'
                    '・入札期限内に気に入った提案を選んで予約確定\n'
                    '・同時オープン可能なリクエストは最大3件',
                    style: theme.textTheme.bodySmall,
                  ),
                ],
              ),
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
            FilledButton(
              onPressed: _loading ? null : _submit,
              style: FilledButton.styleFrom(
                  minimumSize: const Size.fromHeight(52)),
              child: _loading
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                          strokeWidth: 2, color: Colors.white))
                  : const Text('リクエストを出品する'),
            ),
          ],
        ),
      ),
    );
  }
}
