import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../shared/widgets/app_error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../data/auction_repository.dart';
import '../domain/auction.dart';

class SalonDashboardScreen extends ConsumerWidget {
  const SalonDashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final requestsAsync = ref.watch(openRequestsProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('入札ダッシュボード'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.refresh(openRequestsProvider),
          ),
        ],
      ),
      body: requestsAsync.when(
        data: (requests) => requests.isEmpty
            ? const Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.inbox_outlined, size: 48),
                    SizedBox(height: 12),
                    Text('現在マッチするリクエストはありません'),
                  ],
                ),
              )
            : ListView.separated(
                padding: const EdgeInsets.all(12),
                itemCount: requests.length,
                separatorBuilder: (_, __) => const SizedBox(height: 8),
                itemBuilder: (context, i) => _RequestCard(
                  request: requests[i],
                  onBid: () => context.push('/salon/bid/${requests[i].id}'),
                ),
              ),
        loading: () => const LoadingWidget(),
        error: (e, _) => AppErrorWidget(
          message: 'リクエストの読み込みに失敗しました',
          onRetry: () => ref.refresh(openRequestsProvider),
        ),
      ),
    );
  }
}

class _RequestCard extends StatelessWidget {
  const _RequestCard({required this.request, required this.onBid});
  final BookingRequest request;
  final VoidCallback onBid;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final expiresAt = request.expiresAt != null
        ? DateTime.tryParse(request.expiresAt!)
        : null;
    final remaining = expiresAt?.difference(DateTime.now());

    return Card(
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Container(
            color: theme.colorScheme.primaryContainer.withOpacity(0.3),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            child: Row(
              children: [
                Icon(Icons.open_in_browser,
                    size: 16, color: theme.colorScheme.primary),
                const SizedBox(width: 6),
                Text('新着リクエスト',
                    style: theme.textTheme.labelMedium
                        ?.copyWith(color: theme.colorScheme.primary)),
                const Spacer(),
                if (remaining != null && !remaining.isNegative) ...[
                  Icon(Icons.timer_outlined,
                      size: 14, color: theme.colorScheme.outline),
                  const SizedBox(width: 4),
                  Text(
                    '残り ${remaining.inHours}h ${remaining.inMinutes % 60}m',
                    style: theme.textTheme.labelSmall
                        ?.copyWith(color: theme.colorScheme.outline),
                  ),
                ],
              ],
            ),
          ),

          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Budget & Area
                Row(
                  children: [
                    _InfoChip(
                      icon: Icons.payments_outlined,
                      label:
                          '上限 ¥${NumberFormat('#,###').format(request.budgetMaxYen)}',
                      color: theme.colorScheme.primaryContainer,
                    ),
                    const SizedBox(width: 8),
                    _InfoChip(
                      icon: Icons.location_on_outlined,
                      label:
                          '${request.areaPrefecture}${request.areaCity != null ? " ${request.areaCity}" : ""}',
                      color: theme.colorScheme.secondaryContainer,
                    ),
                  ],
                ),
                const SizedBox(height: 12),

                // Date range
                Row(
                  children: [
                    Icon(Icons.date_range,
                        size: 16, color: theme.colorScheme.outline),
                    const SizedBox(width: 6),
                    Text(
                      '${DateFormat('M/d').format(DateTime.parse(request.desiredDateFrom))} 〜 '
                      '${DateFormat('M/d').format(DateTime.parse(request.desiredDateTo))}',
                      style: theme.textTheme.bodySmall,
                    ),
                  ],
                ),

                // Nail data
                if (request.nailDataSnapshot != null) ...[
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Icon(Icons.spa_outlined,
                          size: 16, color: theme.colorScheme.outline),
                      const SizedBox(width: 6),
                      Text(
                        _nailSummary(request.nailDataSnapshot!),
                        style: theme.textTheme.bodySmall
                            ?.copyWith(color: theme.colorScheme.outline),
                      ),
                    ],
                  ),
                ],

                const SizedBox(height: 16),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton.icon(
                    onPressed: onBid,
                    icon: const Icon(Icons.gavel, size: 18),
                    label: const Text('入札する'),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  String _nailSummary(NailDataSnapshot snap) {
    final parts = <String>[];
    if (snap.lengthMm != null)
      parts.add('爪長さ ${snap.lengthMm!.toStringAsFixed(1)}mm');
    if (snap.hasExistingGel) parts.add('既存ジェルあり');
    if (snap.shape != null) parts.add(snap.shape!);
    if (snap.estimatedTreatmentMin != null)
      parts.add('推定 ${snap.estimatedTreatmentMin}分');
    return parts.join(' / ');
  }
}

class _InfoChip extends StatelessWidget {
  const _InfoChip(
      {required this.icon, required this.label, required this.color});
  final IconData icon;
  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14),
          const SizedBox(width: 4),
          Text(label, style: Theme.of(context).textTheme.labelSmall),
        ],
      ),
    );
  }
}
