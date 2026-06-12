import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../shared/widgets/app_error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../data/booking_repository.dart';
import '../domain/booking.dart';

class BookingListScreen extends ConsumerStatefulWidget {
  const BookingListScreen({super.key});

  @override
  ConsumerState<BookingListScreen> createState() => _BookingListScreenState();
}

class _BookingListScreenState extends ConsumerState<BookingListScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabCtrl;

  @override
  void initState() {
    super.initState();
    _tabCtrl = TabController(length: 3, vsync: this);
  }

  @override
  void dispose() {
    _tabCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('マイ予約'),
        bottom: TabBar(
          controller: _tabCtrl,
          tabs: const [
            Tab(text: '予定中'),
            Tab(text: '完了'),
            Tab(text: 'キャンセル'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabCtrl,
        children: const [
          _BookingTab(status: 'confirmed'),
          _BookingTab(status: 'completed'),
          _BookingTab(status: 'cancelled'),
        ],
      ),
    );
  }
}

class _BookingTab extends ConsumerWidget {
  const _BookingTab({required this.status});
  final String status;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final bookingsAsync = ref.watch(myBookingsProvider(status: status));
    return bookingsAsync.when(
      data: (bookings) => bookings.isEmpty
          ? Center(child: Text('${_statusLabel(status)}の予約はありません'))
          : ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: bookings.length,
              separatorBuilder: (_, __) => const SizedBox(height: 8),
              itemBuilder: (context, i) => _BookingCard(
                booking: bookings[i],
                onTap: () => context.push('/bookings/${bookings[i].id}'),
              ),
            ),
      loading: () => const LoadingWidget(),
      error: (e, _) => AppErrorWidget(
        message: '予約の読み込みに失敗しました',
        onRetry: () => ref.refresh(myBookingsProvider(status: status)),
      ),
    );
  }

  String _statusLabel(String status) => switch (status) {
        'confirmed' => '予定中',
        'completed' => '完了',
        'cancelled' => 'キャンセル',
        _ => status,
      };
}

class _BookingCard extends StatelessWidget {
  const _BookingCard({required this.booking, required this.onTap});
  final Booking booking;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheduledAt = DateTime.tryParse(booking.scheduledAt);
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              // Date block
              Container(
                width: 56,
                height: 56,
                decoration: BoxDecoration(
                  color: theme.colorScheme.primaryContainer,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: scheduledAt != null
                    ? Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(DateFormat('M/d').format(scheduledAt),
                              style: theme.textTheme.labelLarge?.copyWith(
                                color: theme.colorScheme.onPrimaryContainer,
                                fontWeight: FontWeight.bold,
                              )),
                          Text(DateFormat('HH:mm').format(scheduledAt),
                              style: theme.textTheme.labelSmall?.copyWith(
                                color: theme.colorScheme.onPrimaryContainer,
                              )),
                        ],
                      )
                    : const Icon(Icons.event),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      booking.salonName ?? 'サロン',
                      style: theme.textTheme.bodyMedium
                          ?.copyWith(fontWeight: FontWeight.w600),
                    ),
                    if (booking.designTitle != null)
                      Text(booking.designTitle!,
                          style: theme.textTheme.bodySmall
                              ?.copyWith(color: theme.colorScheme.outline)),
                    if (booking.totalAmountYen != null)
                      Text(
                        '¥${NumberFormat('#,###').format(booking.totalAmountYen)}',
                        style: theme.textTheme.labelMedium
                            ?.copyWith(color: theme.colorScheme.primary),
                      ),
                  ],
                ),
              ),
              const Icon(Icons.chevron_right),
            ],
          ),
        ),
      ),
    );
  }
}
