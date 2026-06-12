import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../../core/api/client.dart';
import '../../../shared/widgets/app_error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';

part 'notification_screen.g.dart';

// NOTE: run `dart run build_runner build` to generate notification_screen.g.dart

// ---- Domain model ----

class AppNotification {
  const AppNotification({
    required this.id,
    required this.type,
    required this.title,
    this.body,
    required this.isRead,
    required this.createdAt,
  });

  final String id;
  final String type;
  final String title;
  final String? body;
  final bool isRead;
  final DateTime createdAt;

  factory AppNotification.fromJson(Map<String, dynamic> json) =>
      AppNotification(
        id: json['id'] as String,
        type: json['type'] as String,
        title: json['title'] as String,
        body: json['body'] as String?,
        isRead: json['is_read'] as bool,
        createdAt: DateTime.parse(json['created_at'] as String),
      );
}

// ---- Repository ----

class NotificationRepository {
  NotificationRepository(this._dio);
  final Dio _dio;

  Future<List<AppNotification>> list({String? cursor}) async {
    final res = await _dio.get('/notifications', queryParameters: {
      'limit': 30,
      if (cursor != null) 'cursor': cursor,
    });
    return (res.data['notifications'] as List)
        .map((e) => AppNotification.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> markRead(String id) async {
    await _dio.patch('/notifications/$id/read');
  }
}

// ---- Provider ----

@riverpod
NotificationRepository notificationRepository(NotificationRepositoryRef ref) {
  return NotificationRepository(ref.watch(apiClientProvider));
}

@riverpod
Future<List<AppNotification>> notifications(NotificationsRef ref) async {
  return ref.watch(notificationRepositoryProvider).list();
}

// ---- Screen ----

class NotificationScreen extends ConsumerWidget {
  const NotificationScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifAsync = ref.watch(notificationsProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('通知'),
        actions: [
          TextButton(
            onPressed: () => ref.refresh(notificationsProvider),
            child: const Text('更新'),
          ),
        ],
      ),
      body: notifAsync.when(
        data: (notifs) => notifs.isEmpty
            ? const Center(child: Text('通知はありません'))
            : ListView.separated(
                itemCount: notifs.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) => _NotificationTile(
                  notification: notifs[i],
                  onTap: () async {
                    if (!notifs[i].isRead) {
                      await ref
                          .read(notificationRepositoryProvider)
                          .markRead(notifs[i].id);
                      ref.invalidate(notificationsProvider);
                    }
                  },
                ),
              ),
        loading: () => const LoadingWidget(),
        error: (e, _) => AppErrorWidget(
          message: '通知の読み込みに失敗しました',
          onRetry: () => ref.refresh(notificationsProvider),
        ),
      ),
    );
  }
}

class _NotificationTile extends StatelessWidget {
  const _NotificationTile({required this.notification, required this.onTap});
  final AppNotification notification;
  final VoidCallback onTap;

  IconData _iconForType(String type) => switch (type) {
        'new_bid' => Icons.gavel,
        'booking_confirmed' => Icons.event_available,
        'booking_completed' => Icons.check_circle_outline,
        'review_requested' => Icons.rate_review_outlined,
        'royalty_credited' => Icons.payments_outlined,
        _ => Icons.notifications_outlined,
      };

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final unread = !notification.isRead;
    return InkWell(
      onTap: onTap,
      child: Container(
        color: unread
            ? theme.colorScheme.primaryContainer.withOpacity(0.15)
            : null,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: unread
                    ? theme.colorScheme.primaryContainer
                    : theme.colorScheme.surfaceContainerHighest,
                shape: BoxShape.circle,
              ),
              child: Icon(
                _iconForType(notification.type),
                size: 20,
                color: unread
                    ? theme.colorScheme.onPrimaryContainer
                    : theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    notification.title,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      fontWeight: unread ? FontWeight.w700 : FontWeight.normal,
                    ),
                  ),
                  if (notification.body != null) ...[
                    const SizedBox(height: 2),
                    Text(
                      notification.body!,
                      style: theme.textTheme.bodySmall
                          ?.copyWith(color: theme.colorScheme.outline),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                  const SizedBox(height: 4),
                  Text(
                    DateFormat('M月d日 HH:mm')
                        .format(notification.createdAt.toLocal()),
                    style: theme.textTheme.labelSmall
                        ?.copyWith(color: theme.colorScheme.outline),
                  ),
                ],
              ),
            ),
            if (unread)
              Container(
                width: 8,
                height: 8,
                margin: const EdgeInsets.only(top: 4),
                decoration: BoxDecoration(
                  color: theme.colorScheme.primary,
                  shape: BoxShape.circle,
                ),
              ),
          ],
        ),
      ),
    );
  }
}
