import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';

class PurchaseRequestPage extends StatefulWidget {
  const PurchaseRequestPage({super.key});

  @override
  State<PurchaseRequestPage> createState() => _PurchaseRequestPageState();
}

class _PurchaseRequestPageState extends State<PurchaseRequestPage> {
  final ApiClient _api = di.sl<ApiClient>();
  List<Map<String, dynamic>> _requests = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  List<Map<String, dynamic>> _rows(dynamic response) {
    final data = response is Map ? (response['data'] ?? []) : response;
    if (data is! List) return [];
    return data
        .whereType<Map>()
        .map((row) => Map<String, dynamic>.from(row))
        .toList();
  }

  String _message(Object error) {
    if (error is DioException) {
      final data = error.response?.data;
      if (data is Map && data['error'] != null) return data['error'].toString();
    }
    return 'Gagal memuat data. Periksa koneksi dan coba lagi.';
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final response = await _api.client.get('/factory/raw-materials/requests',
          queryParameters: {'mine': 'true'});
      if (!mounted) return;
      setState(() => _requests = _rows(response.data));
    } catch (error) {
      if (mounted) setState(() => _error = _message(error));
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _create() async {
    if (_loading) return;
    final created = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (_) => _PurchaseRequestForm(api: _api),
      ),
    );
    if (created == true && mounted) _load();
  }

  String _status(String? status) => switch (status) {
        'PENDING' => 'Menunggu Purchasing',
        'PROCESSED' => 'Sudah ditindaklanjuti',
        'REJECTED' => 'Ditolak',
        'CANCELLED' => 'Dibatalkan',
        _ => status ?? '-',
      };

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Purchase Request'),
        backgroundColor: AppColors.primaryGreen,
        foregroundColor: Colors.white,
        actions: [
          IconButton(onPressed: _load, icon: const Icon(Icons.refresh))
        ],
      ),
      floatingActionButton: _loading || _error != null
          ? null
          : FloatingActionButton.extended(
              onPressed: _create,
              backgroundColor: AppColors.primaryGreen,
              foregroundColor: Colors.white,
              icon: const Icon(Icons.add),
              label: const Text('Buat Request'),
            ),
      body: Column(
        children: [
          const Padding(
              padding: EdgeInsets.all(16),
              child: Text('Permintaan pembelian saya',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold))),
          if (_loading) const LinearProgressIndicator(),
          if (_error != null)
            Padding(
              padding: const EdgeInsets.all(16),
              child: Text(_error!, style: const TextStyle(color: Colors.red)),
            ),
          Expanded(
            child: _requests.isEmpty && !_loading
                ? const Center(child: Text('Belum ada purchase request.'))
                : RefreshIndicator(
                    onRefresh: _load,
                    child: ListView.builder(
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 88),
                      itemCount: _requests.length,
                      itemBuilder: (context, index) {
                        final pr = _requests[index];
                        final items =
                            pr['items'] is List ? pr['items'] as List : [];
                        final po = pr['purchase_order'];
                        final date = DateTime.tryParse(
                            pr['request_date']?.toString() ?? '');
                        return Card(
                          child: Padding(
                            padding: const EdgeInsets.all(14),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(children: [
                                  Expanded(
                                      child: Text(
                                          pr['pr_number']?.toString() ?? '-',
                                          style: const TextStyle(
                                              fontWeight: FontWeight.bold,
                                              fontSize: 16))),
                                  Text(_status(pr['status']?.toString()),
                                      style: TextStyle(
                                          color: pr['status'] == 'REJECTED'
                                              ? Colors.red
                                              : AppColors.primaryGreen,
                                          fontWeight: FontWeight.w600)),
                                ]),
                                const SizedBox(height: 6),
                                Text(
                                    '${date == null ? '-' : DateFormat('dd MMM yyyy').format(date)} • ${pr['created_by_name'] ?? '-'}'),
                                if (items.isNotEmpty) ...[
                                  const SizedBox(height: 8),
                                  ...items.map((item) {
                                    final row = item is Map
                                        ? item
                                        : <String, dynamic>{};
                                    final product = row['product'];
                                    final name = row['item_name'] ??
                                        (product is Map
                                            ? product['name']
                                            : null);
                                    return Text(
                                        '• ${name ?? '-'}: ${row['requested_qty'] ?? '-'} ${row['unit'] ?? ''}');
                                  }),
                                ],
                                if (po is Map && po['po_number'] != null)
                                  Text('PO: ${po['po_number']}',
                                      style: const TextStyle(
                                          fontWeight: FontWeight.w600)),
                                if (pr['review_notes'] != null &&
                                    pr['review_notes'].toString().isNotEmpty)
                                  Text('Tindak lanjut: ${pr['review_notes']}'),
                                if (pr['status'] == 'REJECTED' &&
                                    pr['rejection_reason'] != null)
                                  Text('Alasan: ${pr['rejection_reason']}',
                                      style:
                                          const TextStyle(color: Colors.red)),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
                  ),
          ),
        ],
      ),
    );
  }
}

class _RequestItem {
  final name = TextEditingController();
  final quantity = TextEditingController(text: '1');
  final unit = TextEditingController(text: 'PCS');
  final notes = TextEditingController();

  void dispose() {
    name.dispose();
    quantity.dispose();
    unit.dispose();
    notes.dispose();
  }
}

class _PurchaseRequestForm extends StatefulWidget {
  const _PurchaseRequestForm({required this.api});

  final ApiClient api;

  @override
  State<_PurchaseRequestForm> createState() => _PurchaseRequestFormState();
}

class _PurchaseRequestFormState extends State<_PurchaseRequestForm> {
  final _formKey = GlobalKey<FormState>();
  final _notes = TextEditingController();
  final List<_RequestItem> _items = [_RequestItem()];
  DateTime _requestDate = DateTime.now();
  DateTime? _requiredDate;
  bool _submitting = false;

  @override
  void dispose() {
    _notes.dispose();
    for (final item in _items) {
      item.dispose();
    }
    super.dispose();
  }

  Future<void> _pickDate(bool required) async {
    final date = await showDatePicker(
      context: context,
      initialDate: required ? (_requiredDate ?? _requestDate) : _requestDate,
      firstDate: DateTime(2020),
      lastDate: DateTime(2100),
    );
    if (date != null) {
      setState(() {
        if (required) {
          _requiredDate = date;
        } else {
          _requestDate = date;
        }
      });
    }
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate() || _submitting) return;
    if (_requiredDate != null && _requiredDate!.isBefore(_requestDate)) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
          content:
              Text('Tanggal dibutuhkan harus setelah tanggal pengajuan.')));
      return;
    }
    setState(() => _submitting = true);
    try {
      final date = DateFormat('yyyy-MM-dd');
      await widget.api.client.post('/factory/raw-materials/requests', data: {
        'request_date': '${date.format(_requestDate)}T00:00:00Z',
        if (_requiredDate != null)
          'required_date': '${date.format(_requiredDate!)}T00:00:00Z',
        'notes': _notes.text.trim(),
        'items': _items
            .map((item) => {
                  'item_name': item.name.text.trim(),
                  'requested_qty': int.parse(item.quantity.text.trim()),
                  'unit': item.unit.text.trim().toUpperCase(),
                  'notes': item.notes.text.trim(),
                })
            .toList(),
      });
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
          content: Text('Purchase Request berhasil dikirim ke Purchasing.')));
      Navigator.pop(context, true);
    } catch (error) {
      if (!mounted) return;
      final data = error is DioException ? error.response?.data : null;
      final message = data is Map && data['error'] != null
          ? data['error'].toString()
          : 'Gagal mengirim Purchase Request.';
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(message)));
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
          title: const Text('Buat Purchase Request'),
          backgroundColor: AppColors.primaryGreen,
          foregroundColor: Colors.white),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            ListTile(
                title: const Text('Tanggal pengajuan'),
                subtitle: Text(DateFormat('dd MMM yyyy').format(_requestDate)),
                trailing: const Icon(Icons.calendar_today),
                onTap: () => _pickDate(false)),
            ListTile(
                title: const Text('Tanggal dibutuhkan (opsional)'),
                subtitle: Text(_requiredDate == null
                    ? 'Belum dipilih'
                    : DateFormat('dd MMM yyyy').format(_requiredDate!)),
                trailing: const Icon(Icons.calendar_today),
                onTap: () => _pickDate(true)),
            const SizedBox(height: 8),
            TextFormField(
                controller: _notes,
                maxLines: 2,
                decoration: const InputDecoration(
                    labelText: 'Keterangan / keperluan',
                    border: OutlineInputBorder())),
            const SizedBox(height: 20),
            const Text('Daftar barang yang diminta',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
            const SizedBox(height: 8),
            for (var index = 0; index < _items.length; index++)
              _itemCard(index),
            OutlinedButton.icon(
                onPressed: () => setState(() => _items.add(_RequestItem())),
                icon: const Icon(Icons.add),
                label: const Text('Tambah barang')),
            const SizedBox(height: 16),
            FilledButton.icon(
                onPressed: _submitting ? null : _submit,
                icon: _submitting
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(strokeWidth: 2))
                    : const Icon(Icons.send),
                label: const Text('Kirim ke Purchasing')),
          ],
        ),
      ),
    );
  }

  Widget _itemCard(int index) {
    final item = _items[index];
    return Card(
      key: ObjectKey(item),
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(children: [
          Row(children: [
            Expanded(
                child: Text('Barang ${index + 1}',
                    style: const TextStyle(fontWeight: FontWeight.bold))),
            if (_items.length > 1)
              IconButton(
                  tooltip: 'Hapus barang',
                  icon: const Icon(Icons.delete_outline, color: Colors.red),
                  onPressed: () => setState(() {
                        _items.removeAt(index);
                        item.dispose();
                      })),
          ]),
          TextFormField(
              controller: item.name,
              decoration: const InputDecoration(
                  labelText: 'Nama barang *', border: OutlineInputBorder()),
              validator: (value) =>
                  (value?.trim().isEmpty ?? true) ? 'Isi nama barang' : null),
          const SizedBox(height: 12),
          Row(children: [
            Expanded(
                child: TextFormField(
                    controller: item.quantity,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(
                        labelText: 'Jumlah *', border: OutlineInputBorder()),
                    validator: (value) =>
                        (int.tryParse(value?.trim() ?? '') ?? 0) < 1
                            ? 'Minimal 1'
                            : null)),
            const SizedBox(width: 12),
            Expanded(
                child: TextFormField(
                    controller: item.unit,
                    decoration: const InputDecoration(
                        labelText: 'Satuan *', border: OutlineInputBorder()),
                    validator: (value) =>
                        (value?.trim().isEmpty ?? true) ? 'Isi satuan' : null)),
          ]),
          const SizedBox(height: 12),
          TextFormField(
              controller: item.notes,
              decoration: const InputDecoration(
                  labelText: 'Catatan barang', border: OutlineInputBorder())),
        ]),
      ),
    );
  }
}
