import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../shared/widgets/app_error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../data/booking_repository.dart';
import '../domain/booking.dart';

class BookingDetailScreen extends ConsumerWidget {
  const BookingDetailScreen({super.key, required this.id});
  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final bookingAsync = ref.watch(bookingDetailProvider(id));
    return bookingAsync.when(
      data: (booking) => _BookingDetailView(booking: booking),
      loading: () => const Scaffold(body: LoadingWidget()),
      error: (e, _) => Scaffold(
        appBar: AppBar(),
        body: AppErrorWidget(
          message: '予約情報の読み込みに失敗しました',
          onRetry: () => ref.refresh(bookingDetailProvider(id)),
        ),
      ),
    );
  }
}

class _BookingDetailView extends ConsumerWidget {
  const _BookingDetailView({required this.booking});
  final Booking booking;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final scheduledAt = DateTime.tryParse(booking.scheduledAt);

    return Scaffold(
      appBar: AppBar(title: const Text('予約詳細')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Status
            _StatusBadge(status: booking.status),
            const SizedBox(height: 20),

            // Main info card
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    if (scheduledAt != null)
                      _InfoRow(
                        icon: Icons.event,
                        label: '施術日時',
                        value: DateFormat('yyyy年M月d日(E) HH:mm', 'ja')
                            .format(scheduledAt),
                      ),
                    _InfoRow(
                        icon: Icons.storefront,
                        label: 'サロン',
                        value: booking.salonName ?? '-'),
                    if (booking.designTitle != null)
                      _InfoRow(
                          icon: Icons.palette,
                          label: 'デザイン',
                          value: booking.designTitle!),
                    if (booking.totalAmountYen != null)
                      _InfoRow(
                        icon: Icons.payments_outlined,
                        label: '支払い金額',
                        value:
                            '¥${NumberFormat('#,###').format(booking.totalAmountYen)}',
                        valueColor: theme.colorScheme.primary,
                      ),
                  ],
                ),
              ),
            ),

            // Review button (completed bookings)
            if (booking.status == 'completed') ...[
              const SizedBox(height: 16),
              FilledButton.icon(
                onPressed: () => _showReviewSheet(context, ref),
                icon: const Icon(Icons.rate_review),
                label: const Text('レビューを投稿する'),
              ),
            ],

            // Cancel button (confirmed only)
            if (booking.status == 'confirmed') ...[
              const SizedBox(height: 16),
              OutlinedButton(
                onPressed: () => _confirmCancel(context, ref),
                style: OutlinedButton.styleFrom(
                    foregroundColor: theme.colorScheme.error),
                child: const Text('予約をキャンセルする'),
              ),
            ],

            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }

  void _showReviewSheet(BuildContext context, WidgetRef ref) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (ctx) => _ReviewSheet(bookingId: booking.id, ref: ref),
    );
  }

  Future<void> _confirmCancel(BuildContext context, WidgetRef ref) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('予約キャンセル'),
        content: const Text('この予約をキャンセルしますか？'),
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
        await ref.read(bookingRepositoryProvider).cancelBooking(booking.id);
        ref.invalidate(bookingDetailProvider(booking.id));
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('予約をキャンセルしました')),
          );
        }
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

class _StatusBadge extends StatelessWidget {
  const _StatusBadge({required this.status});
  final String status;

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (status) {
      'confirmed' => ('予約確定', Colors.green),
      'completed' => ('施術完了', Theme.of(context).colorScheme.primary),
      'cancelled' => ('キャンセル', Colors.red),
      _ => (status, Theme.of(context).colorScheme.outline),
    };
    return Center(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
        decoration: BoxDecoration(
          color: color.withOpacity(0.12),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: color),
        ),
        child: Text(label,
            style: TextStyle(
                color: color, fontWeight: FontWeight.w700, fontSize: 16)),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  const _InfoRow(
      {required this.icon,
      required this.label,
      required this.value,
      this.valueColor});
  final IconData icon;
  final String label;
  final String value;
  final Color? valueColor;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Icon(icon, size: 20, color: Theme.of(context).colorScheme.outline),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label,
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                        color: Theme.of(context).colorScheme.outline)),
                Text(value,
                    style: Theme.of(context)
                        .textTheme
                        .bodyMedium
                        ?.copyWith(color: valueColor)),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ReviewSheet extends StatefulWidget {
  const _ReviewSheet({required this.bookingId, required this.ref});
  final String bookingId;
  final WidgetRef ref;

  @override
  State<_ReviewSheet> createState() => _ReviewSheetState();
}

class _ReviewSheetState extends State<_ReviewSheet> {
  int _reproductionScore = 5;
  int _overallScore = 5;
  final _commentCtrl = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _commentCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() => _loading = true);
    try {
      await widget.ref.read(bookingRepositoryProvider).postReview(
            bookingId: widget.bookingId,
            reproductionScore: _reproductionScore,
            overallScore: _overallScore,
            comment: _commentCtrl.text.trim().isEmpty
                ? null
                : _commentCtrl.text.trim(),
          );
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('レビューを投稿しました')));
      }
    } catch (e) {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(
        left: 24,
        right: 24,
        top: 24,
        bottom: MediaQuery.of(context).viewInsets.bottom + 24,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text('レビュー投稿', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 20),
          Text('再現度スコア', style: Theme.of(context).textTheme.labelLarge),
          _StarRating(
              score: _reproductionScore,
              onChanged: (s) => setState(() => _reproductionScore = s)),
          const SizedBox(height: 16),
          Text('総合スコア', style: Theme.of(context).textTheme.labelLarge),
          _StarRating(
              score: _overallScore,
              onChanged: (s) => setState(() => _overallScore = s)),
          const SizedBox(height: 16),
          TextFormField(
            controller: _commentCtrl,
            decoration: const InputDecoration(
                labelText: 'コメント（任意）', hintText: 'デザイン通りに再現してくれました！'),
            maxLines: 3,
          ),
          const SizedBox(height: 20),
          FilledButton(
            onPressed: _loading ? null : _submit,
            child: _loading
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                        strokeWidth: 2, color: Colors.white))
                : const Text('投稿する'),
          ),
        ],
      ),
    );
  }
}

class _StarRating extends StatelessWidget {
  const _StarRating({required this.score, required this.onChanged});
  final int score;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: List.generate(5, (i) {
        return IconButton(
          icon: Icon(i < score ? Icons.star : Icons.star_border,
              color: Colors.amber),
          onPressed: () => onChanged(i + 1),
        );
      }),
    );
  }
}
