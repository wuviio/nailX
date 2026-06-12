import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../shared/widgets/app_error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../data/design_repository.dart';

class DesignDetailScreen extends ConsumerWidget {
  const DesignDetailScreen({super.key, required this.id});
  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailAsync = ref.watch(designDetailProvider(id));
    return detailAsync.when(
      data: (detail) => _DesignDetailView(detail: detail),
      loading: () => const Scaffold(body: LoadingWidget()),
      error: (e, _) => Scaffold(
        appBar: AppBar(),
        body: AppErrorWidget(
          message: 'デザインの読み込みに失敗しました',
          onRetry: () => ref.refresh(designDetailProvider(id)),
        ),
      ),
    );
  }
}

class _DesignDetailView extends StatelessWidget {
  const _DesignDetailView({required this.detail});
  final dynamic detail; // DesignDetail

  @override
  Widget build(BuildContext context) {
    final design = detail.design;
    final creator = detail.creator;
    final parentIp = detail.parentIp;
    final royaltyNodes = detail.royaltyNodes as List;
    final theme = Theme.of(context);

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 320,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              background: CachedNetworkImage(
                imageUrl: design.previewImageUrl as String,
                fit: BoxFit.cover,
                placeholder: (_, __) =>
                    Container(color: theme.colorScheme.surfaceContainerHighest),
                errorWidget: (_, __, ___) => Container(
                  color: theme.colorScheme.surfaceContainerHighest,
                  child:
                      const Icon(Icons.image_not_supported_outlined, size: 48),
                ),
              ),
            ),
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Title & stats
                  Text(design.title as String,
                      style: theme.textTheme.headlineSmall),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Icon(Icons.favorite_border,
                          size: 16, color: theme.colorScheme.primary),
                      const SizedBox(width: 4),
                      Text('${design.usageCount} 回使用',
                          style: theme.textTheme.bodySmall),
                      const SizedBox(width: 16),
                      if (design.genderTag != null) ...[
                        Icon(Icons.label_outline,
                            size: 16, color: theme.colorScheme.outline),
                        const SizedBox(width: 4),
                        Text(design.genderTag as String,
                            style: theme.textTheme.bodySmall),
                      ],
                    ],
                  ),

                  // Style tags
                  if ((design.styleTags as List).isNotEmpty) ...[
                    const SizedBox(height: 12),
                    Wrap(
                      spacing: 6,
                      children: (design.styleTags as List).map<Widget>((tag) {
                        return Chip(
                          label: Text(tag as String),
                          materialTapTargetSize:
                              MaterialTapTargetSize.shrinkWrap,
                          visualDensity: VisualDensity.compact,
                        );
                      }).toList(),
                    ),
                  ],

                  // Creator
                  if (creator != null) ...[
                    const SizedBox(height: 20),
                    const Divider(),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        CircleAvatar(
                          radius: 20,
                          child: Text(
                            (creator.displayName as String).isNotEmpty
                                ? (creator.displayName as String)[0]
                                : '?',
                          ),
                        ),
                        const SizedBox(width: 12),
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('クリエイター',
                                style: theme.textTheme.labelSmall?.copyWith(
                                    color: theme.colorScheme.outline)),
                            Text(creator.displayName as String,
                                style: theme.textTheme.bodyMedium
                                    ?.copyWith(fontWeight: FontWeight.w600)),
                          ],
                        ),
                        const Spacer(),
                        TextButton(
                          onPressed: () => context.push('/users/${creator.id}'),
                          child: const Text('プロフィール'),
                        ),
                      ],
                    ),
                  ],

                  // Fork origin
                  if (parentIp != null) ...[
                    const SizedBox(height: 16),
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.secondaryContainer
                            .withOpacity(0.5),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Row(
                        children: [
                          Icon(Icons.call_split,
                              size: 18, color: theme.colorScheme.secondary),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text('フォーク元',
                                    style: theme.textTheme.labelSmall),
                                Text(parentIp.title as String,
                                    style: theme.textTheme.bodySmall),
                              ],
                            ),
                          ),
                          TextButton(
                            onPressed: () =>
                                context.push('/designs/${parentIp.id}'),
                            child: const Text('見る'),
                          ),
                        ],
                      ),
                    ),
                  ],

                  // Royalty nodes
                  if (royaltyNodes.isNotEmpty) ...[
                    const SizedBox(height: 20),
                    Text('ロイヤリティ分配', style: theme.textTheme.titleSmall),
                    const SizedBox(height: 8),
                    ...royaltyNodes.map((node) => Padding(
                          padding: const EdgeInsets.only(bottom: 4),
                          child: Row(
                            children: [
                              Icon(
                                node.depthLevel == 0
                                    ? Icons.star
                                    : Icons.subdirectory_arrow_right,
                                size: 16,
                                color: theme.colorScheme.outline,
                              ),
                              const SizedBox(width: 8),
                              Text(
                                'Depth ${node.depthLevel}',
                                style: theme.textTheme.bodySmall,
                              ),
                              const Spacer(),
                              Text(
                                '${node.sharePercent}%',
                                style: theme.textTheme.bodySmall
                                    ?.copyWith(fontWeight: FontWeight.w600),
                              ),
                            ],
                          ),
                        )),
                  ],

                  // Description
                  if (design.description != null &&
                      (design.description as String).isNotEmpty) ...[
                    const SizedBox(height: 20),
                    Text('説明', style: theme.textTheme.titleSmall),
                    const SizedBox(height: 8),
                    Text(design.description as String,
                        style: theme.textTheme.bodyMedium),
                  ],

                  const SizedBox(height: 100),
                ],
              ),
            ),
          ),
        ],
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: () => context.push('/ar'),
                  icon: const Icon(Icons.camera_alt_outlined),
                  label: const Text('AR試着'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: FilledButton.icon(
                  onPressed: () => context.push('/auctions/new'),
                  icon: const Icon(Icons.event_available),
                  label: const Text('このデザインで予約'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
