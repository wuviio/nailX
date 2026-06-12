import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../shared/widgets/app_error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../../designs/domain/design.dart';
import '../../designs/data/design_repository.dart';
import '../../auth/data/auth_repository.dart';
import '../../../shared/widgets/nail_design_card.dart';

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userAsync = ref.watch(currentUserProvider);
    return userAsync.when(
      data: (user) {
        if (user == null) {
          return const Scaffold(body: Center(child: Text('ログインしてください')));
        }
        return _ProfileView(user: user);
      },
      loading: () => const Scaffold(body: LoadingWidget()),
      error: (e, _) => Scaffold(
        appBar: AppBar(),
        body: AppErrorWidget(
          message: 'プロフィールの読み込みに失敗しました',
          onRetry: () => ref.refresh(currentUserProvider),
        ),
      ),
    );
  }
}

class _ProfileView extends ConsumerWidget {
  const _ProfileView({required this.user});
  final dynamic user;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final myDesignsAsync = ref.watch(
      designFeedProvider(filter: const DesignFeedFilter()),
    );

    return Scaffold(
      appBar: AppBar(
        title: const Text('マイページ'),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings_outlined),
            onPressed: () => context.push('/settings'),
          ),
        ],
      ),
      body: CustomScrollView(
        slivers: [
          // Profile header
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                children: [
                  CircleAvatar(
                    radius: 40,
                    child: Text(
                      (user.displayName as String).isNotEmpty
                          ? (user.displayName as String)[0]
                          : '?',
                      style: theme.textTheme.headlineMedium,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    user.displayName as String,
                    style: theme.textTheme.titleLarge
                        ?.copyWith(fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    user.role as String,
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: theme.colorScheme.outline),
                  ),
                  const SizedBox(height: 16),
                  // Stats row
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                    children: [
                      _StatBlock(
                          label: 'ポイント',
                          value:
                              '${NumberFormat('#,###').format(user.pointBalance as int)} pt'),
                      Container(
                          width: 1,
                          height: 40,
                          color: theme.colorScheme.outlineVariant),
                      const _StatBlock(label: 'デザイン', value: '-'),
                      Container(
                          width: 1,
                          height: 40,
                          color: theme.colorScheme.outlineVariant),
                      const _StatBlock(label: '収益', value: '-'),
                    ],
                  ),
                  const SizedBox(height: 16),
                  // Action buttons
                  Row(
                    children: [
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: () => context.push('/profile/edit'),
                          icon: const Icon(Icons.edit_outlined, size: 18),
                          label: const Text('プロフィール編集'),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: () => context.push('/bookings'),
                          icon: const Icon(Icons.calendar_today_outlined,
                              size: 18),
                          label: const Text('マイ予約'),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),

          // My designs section
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Row(
                children: [
                  Text('マイデザイン', style: theme.textTheme.titleMedium),
                  const Spacer(),
                  TextButton(
                    onPressed: () => context.push('/designs/new'),
                    child: const Text('+ 作成'),
                  ),
                ],
              ),
            ),
          ),

          myDesignsAsync.when(
            data: (designs) => designs.isEmpty
                ? const SliverToBoxAdapter(
                    child: Padding(
                      padding: EdgeInsets.all(32),
                      child:
                          Center(child: Text('まだデザインがありません\n最初のデザインを作成しましょう')),
                    ),
                  )
                : SliverGrid(
                    delegate: SliverChildBuilderDelegate(
                      (context, i) => NailDesignCard(
                        design: designs[i],
                        onTap: () => context.push('/designs/${designs[i].id}'),
                      ),
                      childCount: designs.length,
                    ),
                    gridDelegate:
                        const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 2,
                      childAspectRatio: 0.72,
                      crossAxisSpacing: 8,
                      mainAxisSpacing: 8,
                    ),
                  ),
            loading: () => const SliverToBoxAdapter(child: LoadingWidget()),
            error: (_, __) => const SliverToBoxAdapter(
              child: Padding(
                  padding: EdgeInsets.all(16), child: Text('デザインの読み込みに失敗しました')),
            ),
          ),

          // Sign out
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: TextButton(
                onPressed: () async {
                  await ref.read(authRepositoryProvider).signOut();
                },
                style: TextButton.styleFrom(
                    foregroundColor: theme.colorScheme.error),
                child: const Text('ログアウト'),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _StatBlock extends StatelessWidget {
  const _StatBlock({required this.label, required this.value});
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(value,
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(fontWeight: FontWeight.bold)),
        const SizedBox(height: 2),
        Text(label,
            style: Theme.of(context)
                .textTheme
                .labelSmall
                ?.copyWith(color: Theme.of(context).colorScheme.outline)),
      ],
    );
  }
}
