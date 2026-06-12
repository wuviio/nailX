import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../shared/widgets/app_error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../data/auction_repository.dart';
import '../domain/auction.dart';

class AuctionDetailScreen extends ConsumerWidget {
  const AuctionDetailScreen({super.key, required this.id});
  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailAsync = ref.watch(bookingRequestDetailProvider(id));
    return detailAsync.when(
      data: (detail) => _AuctionDetailView(detail: detail, requestId: id),
      loading: () => const Scaffold(body: LoadingWidget()),
      error: (e, _) => Scaffold(
        appBar: AppBar(title: const Text('入札一覧')),
        body: AppErrorWidget(
          message: 'データの読み込みに失敗しました',
          onRetry: () => ref.refresh(bookingRequestDetailProvider(id)),
        ),
      ),
    );
  }
}

class _AuctionDetailView extends ConsumerWidget {
  const _AuctionDetailView({required this.detail, required this.requestId});
  final BookingRequestWithBids detail;
  final String requestId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final request = detail.request;
    final bids = detail.bids;
    final theme = Theme.of(context);
    final statusLabel = _statusLabel(request.status);
    final statusColor = _statusColor(context, request.status);

    return Scaffold(
      appBar: AppBar(
        title: const Text('入札一覧'),
        actions: [
          if (request.status == 'open' || request.status == 'bidding')
            TextButton(
              onPressed: () => _confirmCancel(context, ref),
              child: const Text('キャンセル', style: TextStyle(color: Colors.red)),
            ),
        ],
      ),
      body: CustomScrollView(
        slivers: [
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Status badge
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 12, vertical: 4),
                        decoration: BoxDecoration(
                          color: statusColor.withOpacity(0.15),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(color: statusColor),
                        ),
                        child: Text(statusLabel,
                            style: TextStyle(
                                color: statusColor,
                                fontWeight: FontWeight.w600)),
                      ),
                      const Spacer(),
                      if (request.expiresAt != null)
                        _ExpiryTimer(expiresAt: request.expiresAt!),
                    ],
                  ),
                  const SizedBox(height: 16),

                  // Request summary card
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        children: [
                          _InfoRow('予算上限',
                              '¥${NumberFormat('#,###').format(request.budgetMaxYen)}'),
                          _InfoRow('エリア',
                              '${request.areaPrefecture}${request.areaCity != null ? " ${request.areaCity}" : ""}'),
                          if (request.desiredDateFrom.isNotEmpty)
                            _InfoRow(
                              '希望日程',
                              '${DateFormat('M/d').format(DateTime.parse(request.desiredDateFrom))} 〜 '
                                  '${DateFormat('M/d').format(DateTime.parse(request.desiredDateTo))}',
                            ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),

                  Row(
                    children: [
                      Text('入札 ${bids.length} 件',
                          style: theme.textTheme.titleMedium),
                      const Spacer(),
                      if (bids.isNotEmpty)
                        Text(
                            '最安値: ¥${NumberFormat('#,###').format(bids.map((b) => b.priceYen).reduce((a, b) => a < b ? a : b))}',
                            style: theme.textTheme.bodySmall
                                ?.copyWith(color: theme.colorScheme.primary)),
                    ],
                  ),
                  const SizedBox(height: 8),
                ],
              ),
            ),
          ),

          // Bid list
          if (bids.isEmpty)
            const SliverFillRemaining(
              child: Center(child: Text('まだ入札がありません\nサロンからのオファーをお待ちください')),
            )
          else
            SliverList(
              delegate: SliverChildBuilderDelegate(
                (context, i) => _BidCard(
                  bid: bids[i],
                  onAccept: request.status == 'bidding'
                      ? () => _acceptBid(context, bids[i])
                      : null,
                ),
                childCount: bids.length,
              ),
            ),
        ],
      ),
    );
  }

  String _statusLabel(String status) => switch (status) {
        'open' => '入札募集中',
        'bidding' => '入札あり',
        'confirmed' => '予約確定',
        'cancelled' => 'キャンセル済み',
        'expired' => '期限切れ',
        _ => status,
      };

  Color _statusColor(BuildContext context, String status) {
    final cs = Theme.of(context).colorScheme;
    return switch (status) {
      'open' => cs.primary,
      'bidding' => Colors.orange,
      'confirmed' => Colors.green,
      'cancelled' || 'expired' => cs.outline,
      _ => cs.outline,
    };
  }

  void _acceptBid(BuildContext context, Bid bid) {
    context.push('/bookings/new?bid_id=${bid.id}');
  }

  Future<void> _confirmCancel(BuildContext context, WidgetRef ref) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('リクエストをキャンセル'),
        content: const Text('このリクエストをキャンセルしますか？この操作は取り消せません。'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('戻る')),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('キャンセルする'),
          ),
        ],
      ),
    );
    if (confirmed == true && context.mounted) {
      try {
        await ref.read(auctionRepositoryProvider).cancelRequest(requestId);
        if (context.mounted) context.pop();
      } catch (e) {
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('キャンセルに失敗しました')),
          );
        }
      }
    }
  }
}

class _BidCard extends StatelessWidget {
  const _BidCard({required this.bid, this.onAccept});
  final Bid bid;
  final VoidCallback? onAccept;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                CircleAvatar(
                  radius: 20,
                  child: Text(bid.salonName?.isNotEmpty == true
                      ? bid.salonName![0]
                      : 'S'),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(bid.salonName ?? 'サロン',
                          style: theme.textTheme.bodyMedium
                              ?.copyWith(fontWeight: FontWeight.w600)),
                      if (bid.salonReproductionScore != null)
                        Row(children: [
                          const Icon(Icons.star, size: 14, color: Colors.amber),
                          Text(
                              ' 再現度 ${bid.salonReproductionScore!.toStringAsFixed(1)}',
                              style: theme.textTheme.labelSmall),
                        ]),
                    ],
                  ),
                ),
                Text(
                  '¥${NumberFormat('#,###').format(bid.priceYen)}',
                  style: theme.textTheme.titleLarge?.copyWith(
                    color: theme.colorScheme.primary,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            if (bid.availableSlotAt.isNotEmpty)
              _InfoRow(
                  '提案日時',
                  DateFormat('M/d(E) HH:mm', 'ja')
                      .format(DateTime.parse(bid.availableSlotAt))),
            _InfoRow(
                'オフ',
                bid.includesRemoval
                    ? '込み'
                    : '別途 ¥${NumberFormat('#,###').format(bid.removalFeeYen)}'),
            if (bid.dynamicDiscountReason != null)
              _InfoRow('割引理由', bid.dynamicDiscountReason!),
            if (bid.message != null && bid.message!.isNotEmpty) ...[
              const SizedBox(height: 8),
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(bid.message!, style: theme.textTheme.bodySmall),
              ),
            ],
            if (onAccept != null) ...[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: FilledButton(
                  onPressed: onAccept,
                  child: const Text('この提案で予約する'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  const _InfoRow(this.label, this.value);
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        children: [
          SizedBox(
            width: 72,
            child: Text(label,
                style: Theme.of(context)
                    .textTheme
                    .labelSmall
                    ?.copyWith(color: Theme.of(context).colorScheme.outline)),
          ),
          Expanded(
              child: Text(value, style: Theme.of(context).textTheme.bodySmall)),
        ],
      ),
    );
  }
}

class _ExpiryTimer extends StatelessWidget {
  const _ExpiryTimer({required this.expiresAt});
  final String expiresAt;

  @override
  Widget build(BuildContext context) {
    final expiry = DateTime.tryParse(expiresAt);
    if (expiry == null) return const SizedBox.shrink();
    final remaining = expiry.difference(DateTime.now());
    if (remaining.isNegative) {
      return Text('期限切れ',
          style: TextStyle(color: Theme.of(context).colorScheme.error));
    }
    final hours = remaining.inHours;
    final minutes = remaining.inMinutes % 60;
    return Row(
      children: [
        Icon(Icons.timer_outlined,
            size: 16, color: Theme.of(context).colorScheme.outline),
        const SizedBox(width: 4),
        Text('残り ${hours}h ${minutes}m',
            style: Theme.of(context).textTheme.labelSmall),
      ],
    );
  }
}
