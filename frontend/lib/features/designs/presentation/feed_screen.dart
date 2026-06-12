import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../shared/widgets/app_error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../../../shared/widgets/nail_design_card.dart';
import '../domain/design.dart';
import '../data/design_repository.dart';

class FeedScreen extends ConsumerStatefulWidget {
  const FeedScreen({super.key});

  @override
  ConsumerState<FeedScreen> createState() => _FeedScreenState();
}

class _FeedScreenState extends ConsumerState<FeedScreen> {
  DesignFeedFilter _filter = const DesignFeedFilter();

  @override
  Widget build(BuildContext context) {
    final feedAsync = ref.watch(designFeedProvider(filter: _filter));
    return Scaffold(
      appBar: AppBar(
        title: const Text('nailX'),
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () => _showFilterSheet(context),
          ),
        ],
      ),
      body: Column(
        children: [
          _FilterChips(
            filter: _filter,
            onChanged: (f) => setState(() => _filter = f),
          ),
          Expanded(
            child: feedAsync.when(
              data: (designs) => designs.isEmpty
                  ? const Center(child: Text('デザインがまだありません'))
                  : GridView.builder(
                      padding: const EdgeInsets.all(8),
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 2,
                        childAspectRatio: 0.72,
                        crossAxisSpacing: 8,
                        mainAxisSpacing: 8,
                      ),
                      itemCount: designs.length,
                      itemBuilder: (context, index) => NailDesignCard(
                        design: designs[index],
                        onTap: () => context.push('/designs/${designs[index].id}'),
                      ),
                    ),
              loading: () => const LoadingWidget(),
              error: (e, _) => AppErrorWidget(
                message: 'デザインの読み込みに失敗しました',
                onRetry: () => ref.refresh(designFeedProvider(filter: _filter)),
              ),
            ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.push('/designs/new'),
        icon: const Icon(Icons.add),
        label: const Text('デザイン作成'),
      ),
    );
  }

  void _showFilterSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (_) => _FilterSheet(
        filter: _filter,
        onApply: (f) {
          setState(() => _filter = f);
          Navigator.pop(context);
        },
      ),
    );
  }
}

class _FilterChips extends StatelessWidget {
  const _FilterChips({required this.filter, required this.onChanged});
  final DesignFeedFilter filter;
  final ValueChanged<DesignFeedFilter> onChanged;

  @override
  Widget build(BuildContext context) {
    final genderOptions = [
      (null, 'すべて'),
      ('feminine', '女性向け'),
      ('masculine', '男性向け'),
      ('neutral', 'ユニセックス'),
    ];
    return SizedBox(
      height: 48,
      child: ListView(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 8),
        children: genderOptions.map((opt) {
          final (value, label) = opt;
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: FilterChip(
              label: Text(label),
              selected: filter.genderTag == value,
              onSelected: (_) => onChanged(filter.copyWith(genderTag: value)),
            ),
          );
        }).toList(),
      ),
    );
  }
}

class _FilterSheet extends StatefulWidget {
  const _FilterSheet({required this.filter, required this.onApply});
  final DesignFeedFilter filter;
  final ValueChanged<DesignFeedFilter> onApply;

  @override
  State<_FilterSheet> createState() => _FilterSheetState();
}

class _FilterSheetState extends State<_FilterSheet> {
  late DesignFeedFilter _filter;

  @override
  void initState() {
    super.initState();
    _filter = widget.filter;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: EdgeInsets.only(
        left: 24, right: 24, top: 24,
        bottom: MediaQuery.of(context).viewInsets.bottom + 24,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text('並び順', style: theme.textTheme.labelLarge),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            children: [
              ('latest', '新着順'),
              ('usage_count', '人気順'),
            ].map((s) {
              final (value, label) = s;
              return ChoiceChip(
                label: Text(label),
                selected: _filter.sort == value,
                onSelected: (_) => setState(() => _filter = _filter.copyWith(sort: value)),
              );
            }).toList(),
          ),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: () => widget.onApply(_filter),
            child: const Text('適用'),
          ),
        ],
      ),
    );
  }
}
