import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../data/auction_repository.dart';

class BidFormScreen extends ConsumerStatefulWidget {
  const BidFormScreen({super.key, required this.requestId});
  final String requestId;

  @override
  ConsumerState<BidFormScreen> createState() => _BidFormScreenState();
}

class _BidFormScreenState extends ConsumerState<BidFormScreen> {
  final _priceCtrl = TextEditingController();
  final _discountReasonCtrl = TextEditingController();
  final _messageCtrl = TextEditingController();
  bool _includesRemoval = false;
  int _removalFee = 0;
  DateTime? _slotAt;
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _priceCtrl.dispose();
    _discountReasonCtrl.dispose();
    _messageCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickSlot() async {
    final date = await showDatePicker(
      context: context,
      initialDate: DateTime.now().add(const Duration(days: 1)),
      firstDate: DateTime.now().add(const Duration(days: 1)),
      lastDate: DateTime.now().add(const Duration(days: 60)),
    );
    if (date == null || !mounted) return;
    final time = await showTimePicker(
      context: context,
      initialTime: const TimeOfDay(hour: 10, minute: 0),
    );
    if (time == null) return;
    setState(() {
      _slotAt =
          DateTime(date.year, date.month, date.day, time.hour, time.minute);
    });
  }

  Future<void> _submit() async {
    final price = int.tryParse(_priceCtrl.text.replaceAll(',', ''));
    if (price == null || price <= 0) {
      setState(() => _error = '正しい価格を入力してください');
      return;
    }
    if (_slotAt == null) {
      setState(() => _error = '提案日時を選択してください');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref.read(auctionRepositoryProvider).placeBid(
            requestId: widget.requestId,
            priceYen: price,
            includesRemoval: _includesRemoval,
            removalFeeYen: _includesRemoval ? 0 : _removalFee,
            availableSlotAt: _slotAt!.toIso8601String(),
            dynamicDiscountReason: _discountReasonCtrl.text.trim().isEmpty
                ? null
                : _discountReasonCtrl.text.trim(),
            message: _messageCtrl.text.trim().isEmpty
                ? null
                : _messageCtrl.text.trim(),
          );
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('入札しました！')));
        context.pop();
      }
    } catch (e) {
      setState(() => _error = '入札に失敗しました。既に入札済みの場合は PATCH で修正してください。');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final fmt = DateFormat('M/d(E) HH:mm', 'ja');
    return Scaffold(
      appBar: AppBar(title: const Text('入札フォーム')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Price
            TextFormField(
              controller: _priceCtrl,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(
                labelText: '提案価格（円）*',
                hintText: '6500',
                prefixIcon: Icon(Icons.currency_yen),
              ),
            ),
            const SizedBox(height: 16),

            // Removal
            Card(
              child: Column(
                children: [
                  SwitchListTile(
                    title: const Text('ジェルオフ込み'),
                    value: _includesRemoval,
                    onChanged: (v) => setState(() => _includesRemoval = v),
                  ),
                  if (!_includesRemoval) ...[
                    const Divider(height: 1),
                    Padding(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 16, vertical: 8),
                      child: TextFormField(
                        initialValue: _removalFee > 0 ? '$_removalFee' : '',
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'オフ別途料金（円）',
                          hintText: '0',
                          isDense: true,
                        ),
                        onChanged: (v) => _removalFee = int.tryParse(v) ?? 0,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(height: 16),

            // Slot
            OutlinedButton.icon(
              onPressed: _pickSlot,
              icon: const Icon(Icons.schedule),
              label: Text(
                _slotAt != null ? fmt.format(_slotAt!) : '提案スロット（日時）を選択 *',
              ),
              style: OutlinedButton.styleFrom(
                  minimumSize: const Size.fromHeight(52)),
            ),
            const SizedBox(height: 16),

            // Discount reason
            TextFormField(
              controller: _discountReasonCtrl,
              decoration: const InputDecoration(
                labelText: 'ダイナミック割引理由（任意）',
                hintText: '直前空き枠割引',
                prefixIcon: Icon(Icons.local_offer_outlined),
              ),
            ),
            const SizedBox(height: 16),

            // Message
            TextFormField(
              controller: _messageCtrl,
              maxLines: 3,
              decoration: const InputDecoration(
                labelText: '一言メッセージ（任意）',
                hintText: '3Dパーツの再現度には自信があります',
                alignLabelWithHint: true,
              ),
            ),
            const SizedBox(height: 16),

            // Info
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: theme.colorScheme.surfaceContainerHighest,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                '・入札は一発提示（チャットなし）\n'
                '・再入札は価格引き下げのみ、1回まで可能\n'
                '・ユーザーがこの提案を選択すると予約確定します',
                style: theme.textTheme.bodySmall,
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
            FilledButton.icon(
              onPressed: _loading ? null : _submit,
              icon: const Icon(Icons.gavel),
              label: _loading
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                          strokeWidth: 2, color: Colors.white))
                  : const Text('入札する'),
              style: FilledButton.styleFrom(
                  minimumSize: const Size.fromHeight(52)),
            ),
          ],
        ),
      ),
    );
  }
}
