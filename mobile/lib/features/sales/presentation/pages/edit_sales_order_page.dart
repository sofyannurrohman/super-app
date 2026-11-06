import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';

/// Halaman edit SO bagi Sales.
/// Hanya bisa digunakan saat SO berstatus DRAFT (sebelum diverifikasi Admin Nota).
class EditSalesOrderPage extends StatefulWidget {
  final Map<String, dynamic> salesOrder;

  const EditSalesOrderPage({super.key, required this.salesOrder});

  @override
  State<EditSalesOrderPage> createState() => _EditSalesOrderPageState();
}

class _EditSalesOrderPageState extends State<EditSalesOrderPage> {
  final _apiService = SalesApiService(apiClient: di.sl<ApiClient>());
  final _currency =
      NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);

  late List<_EditableItem> _items;
  final TextEditingController _adjustNotesCtrl = TextEditingController();
  bool _isSaving = false;

  // Products master untuk autocomplete / tambah item baru
  List<dynamic> _allProducts = [];
  bool _loadingProducts = false;

  @override
  void initState() {
    super.initState();
    _initItems();
    _fetchProducts();
  }

  void _initItems() {
    final raw = widget.salesOrder['items'] as List<dynamic>? ?? [];
    _items = raw.map((item) {
      final productName =
          item['product']?['name'] ?? item['product_name'] ?? 'Produk';
      final productId = item['product_id'] as String? ?? '';
      final orderedQty = (item['ordered_quantity'] as num?)?.toInt() ??
          (item['quantity'] as num?)?.toInt() ??
          0;
      final ppu = (item['pieces_per_unit'] as num?)?.toInt() ?? 1;
      final unit = item['unit']?.toString() ?? 'PCS';
      final price = (item['price'] as num?)?.toDouble() ?? 0.0;
      final discountPercent =
          (item['discount_percent'] as num?)?.toDouble() ?? 0.0;
      return _EditableItem(
        productId: productId,
        productName: productName,
        orderedQtyCtrl: TextEditingController(text: orderedQty.toString()),
        piecesPerUnitCtrl: TextEditingController(text: ppu.toString()),
        priceCtrl: TextEditingController(text: price.toStringAsFixed(0)),
        discountCtrl:
            TextEditingController(text: discountPercent.toStringAsFixed(1)),
        unit: unit,
      );
    }).toList();
  }

  Widget _buildImage(String? rawUrl) {
    if (rawUrl == null || rawUrl.isEmpty) {
      return const Icon(Icons.inventory_2_outlined,
          size: 18, color: Colors.grey);
    }
    String imageUrl = rawUrl;
    if (!rawUrl.startsWith('http') && !rawUrl.startsWith('assets/')) {
      imageUrl = '${AppConstants.baseUrl.replaceAll('/api/v1/', '')}$rawUrl';
    }

    if (imageUrl.startsWith('http')) {
      return CachedNetworkImage(imageUrl: imageUrl, fit: BoxFit.cover, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,);
    } else if (imageUrl.startsWith('assets/')) {
      return Image.asset(
        imageUrl,
        fit: BoxFit.cover,
        errorBuilder: (_, __, ___) => const Icon(Icons.inventory_2_outlined,
            size: 18, color: Colors.grey),
      );
    }
    return const Icon(Icons.inventory_2_outlined, size: 18, color: Colors.grey);
  }

  Future<void> _fetchProducts() async {
    setState(() => _loadingProducts = true);
    try {
      final api = di.sl<ApiClient>();
      final res = await api.client.get('factory/products');
      List? raw;
      if (res.data is Map && res.data.containsKey('data')) {
        raw = res.data['data'] as List?;
      } else if (res.data is List) {
        raw = res.data as List?;
      }
      if (mounted && raw != null) {
        setState(() => _allProducts = raw!);
      }
    } catch (e) {
      debugPrint('EditSO: gagal fetch produk: $e');
    } finally {
      if (mounted) setState(() => _loadingProducts = false);
    }
  }

  double _calcSubtotal(_EditableItem item) {
    final qty = (int.tryParse(item.orderedQtyCtrl.text) ?? 0) *
        (int.tryParse(item.piecesPerUnitCtrl.text) ?? 1);
    final price = double.tryParse(item.priceCtrl.text) ?? 0;
    final disc = double.tryParse(item.discountCtrl.text) ?? 0;
    return qty * price * (1 - disc / 100);
  }

  double get _grandTotal =>
      _items.fold(0.0, (sum, item) => sum + _calcSubtotal(item));

  Future<void> _saveChanges() async {
    final soId = widget.salesOrder['id'] as String?;
    if (soId == null) return;

    // Validasi
    for (final item in _items) {
      final qty = int.tryParse(item.orderedQtyCtrl.text) ?? 0;
      if (qty < 0) {
        _showError(
            'Jumlah pesanan tidak boleh negatif untuk produk ${item.productName}.');
        return;
      }
    }

    setState(() => _isSaving = true);
    try {
      final itemsPayload = _items.where((item) {
        // Hanya kirim item yang qty-nya > 0 (yang 0 dianggap hapus)
        return (int.tryParse(item.orderedQtyCtrl.text) ?? 0) > 0;
      }).map((item) {
        return {
          'product_id': item.productId,
          'ordered_quantity': int.tryParse(item.orderedQtyCtrl.text) ?? 0,
          'pieces_per_unit': int.tryParse(item.piecesPerUnitCtrl.text) ?? 1,
          'unit': item.unit,
          'price': double.tryParse(item.priceCtrl.text) ?? 0.0,
          'discount_percent': double.tryParse(item.discountCtrl.text) ?? 0.0,
        };
      }).toList();

      if (itemsPayload.isEmpty) {
        _showError('Minimal harus ada satu item dalam pesanan.');
        setState(() => _isSaving = false);
        return;
      }

      final payload = {
        'items': itemsPayload,
        'discount_amount': 0.0,
        'adjust_notes': _adjustNotesCtrl.text.trim().isNotEmpty
            ? '[KOREKSI SALES] ${_adjustNotesCtrl.text.trim()}'
            : '[KOREKSI SALES]',
      };

      await _apiService.editSalesOrder(soId, payload);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
          content: Text('✓ Pesanan berhasil diperbarui!'),
          backgroundColor: Color(0xFF10B981),
          duration: Duration(seconds: 2),
        ));
        Navigator.pop(context, true); // true = berhasil, supaya parent refresh
      }
    } catch (e) {
      _showError(e.toString().replaceFirst('Exception: ', ''));
    } finally {
      if (mounted) setState(() => _isSaving = false);
    }
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(
      content: Text(msg),
      backgroundColor: Colors.redAccent,
      duration: const Duration(seconds: 4),
    ));
  }

  void _removeItem(int index) {
    setState(() => _items.removeAt(index));
  }

  void _addProductDialog() {
    if (_allProducts.isEmpty) {
      _showError('Daftar produk masih dimuat atau kosong.');
      return;
    }
    setState(() {
      final defaultProduct = _allProducts.first;
      _items.add(_EditableItem(
        productId: defaultProduct['id'].toString(),
        productName: defaultProduct['name'].toString(),
        orderedQtyCtrl: TextEditingController(text: '1'),
        piecesPerUnitCtrl: TextEditingController(
            text: (defaultProduct['pcs_per_unit']?.toString() ?? '1')),
        priceCtrl: TextEditingController(
            text: (defaultProduct['selling_price']?.toString() ?? '0')),
        discountCtrl: TextEditingController(text: '0'),
        unit: defaultProduct['unit']?.toString() ?? 'PCS',
      ));
    });
  }

  Widget _buildModalField(String label, TextEditingController ctrl,
      {bool isNumber = false, bool readOnly = false}) {
    return TextField(
      controller: ctrl,
      readOnly: readOnly,
      keyboardType: isNumber
          ? const TextInputType.numberWithOptions(decimal: true)
          : TextInputType.text,
      decoration: InputDecoration(
        labelText: label,
        labelStyle: GoogleFonts.outfit(fontSize: 13, color: Colors.blueGrey),
        filled: true,
        fillColor: const Color(0xFFF8FAFC),
        border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide.none),
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      ),
    );
  }

  @override
  void dispose() {
    _adjustNotesCtrl.dispose();
    for (final item in _items) {
      item.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final soNumber =
        widget.salesOrder['so_number'] ?? widget.salesOrder['SONumber'] ?? 'SO';
    final storeName = widget.salesOrder['store']?['name'] ?? 'Toko';

    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded,
              color: Color(0xFF334155)),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('KOREKSI PESANAN',
                style: GoogleFonts.outfit(
                    fontSize: 10,
                    color: const Color(0xFFF59E0B),
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1)),
            Text(soNumber,
                style: GoogleFonts.outfit(
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B))),
          ],
        ),
        actions: [
          if (!_loadingProducts)
            TextButton.icon(
              onPressed: _addProductDialog,
              icon: const Icon(Icons.add_circle_outline_rounded,
                  color: Color(0xFF10B981), size: 20),
              label: Text('TAMBAH',
                  style: GoogleFonts.outfit(
                      fontSize: 11,
                      fontWeight: FontWeight.w800,
                      color: const Color(0xFF10B981))),
            ),
          const SizedBox(width: 4),
        ],
      ),
      body: Column(
        children: [
          // Info banner
          Container(
            color: const Color(0xFFFFF7ED),
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
            child: Row(
              children: [
                const Icon(Icons.warning_amber_rounded,
                    color: Color(0xFFF59E0B), size: 20),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(storeName,
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800, fontSize: 13)),
                      Text(
                          'Pesanan hanya bisa diedit sebelum diverifikasi Admin Nota (status DRAFT).',
                          style: GoogleFonts.outfit(
                              fontSize: 11, color: const Color(0xFF92400E))),
                    ],
                  ),
                ),
              ],
            ),
          ),

          // Items list
          Expanded(
            child: _items.isEmpty
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.remove_shopping_cart_rounded,
                            size: 64, color: Colors.grey.shade300),
                        const SizedBox(height: 16),
                        Text('Belum ada item',
                            style: GoogleFonts.outfit(
                                fontSize: 16, color: Colors.blueGrey)),
                        const SizedBox(height: 8),
                        Text('Tap TAMBAH untuk menambahkan produk',
                            style: GoogleFonts.outfit(
                                fontSize: 13, color: Colors.grey)),
                      ],
                    ),
                  )
                : ListView.builder(
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                    itemCount: _items.length,
                    itemBuilder: (ctx, index) =>
                        _buildItemCard(_items[index], index),
                  ),
          ),

          // Catatan Penyesuaian
          Container(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
            color: Colors.white,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Divider(height: 1),
                const SizedBox(height: 12),
                Text('Alasan / Catatan Koreksi',
                    style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w700,
                        fontSize: 13,
                        color: const Color(0xFF334155))),
                const SizedBox(height: 8),
                TextField(
                  controller: _adjustNotesCtrl,
                  maxLines: 2,
                  decoration: InputDecoration(
                    hintText:
                        'Contoh: Salah input qty produk A, koreksi dari 10 menjadi 8',
                    hintStyle:
                        GoogleFonts.outfit(fontSize: 12, color: Colors.grey),
                    filled: true,
                    fillColor: const Color(0xFFF8FAFC),
                    border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: BorderSide.none),
                    contentPadding: const EdgeInsets.all(12),
                  ),
                ),
              ],
            ),
          ),

          // Bottom summary bar
          _buildBottomBar(),
        ],
      ),
    );
  }

  Widget _buildItemCard(_EditableItem item, int index) {
    return AnimatedContainer(
      duration: const Duration(milliseconds: 200),
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 12,
              offset: const Offset(0, 4))
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header: nama produk + hapus
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: const Color(0xFF10B981).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Icon(Icons.inventory_2_rounded,
                      color: Color(0xFF10B981), size: 20),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: _loadingProducts
                      ? Text(item.productName,
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800,
                              fontSize: 14,
                              color: const Color(0xFF1E293B)))
                      : DropdownButtonFormField<String>(
                          value: _allProducts.any((p) =>
                                  p['id'].toString() ==
                                  item.productId.toString())
                              ? item.productId.toString()
                              : null,
                          hint: Text(item.productName,
                              style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w800,
                                  fontSize: 14,
                                  color: const Color(0xFF1E293B))),
                          decoration: InputDecoration(
                            contentPadding: const EdgeInsets.symmetric(
                                horizontal: 10, vertical: 8),
                            border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(10),
                                borderSide:
                                    BorderSide(color: Colors.grey.shade300)),
                            enabledBorder: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(10),
                                borderSide:
                                    BorderSide(color: Colors.grey.shade300)),
                            filled: true,
                            fillColor: const Color(0xFFF8FAFC),
                          ),
                          isExpanded: true,
                          icon: const Icon(Icons.arrow_drop_down_rounded,
                              color: Color(0xFF10B981)),
                          items: _allProducts.map((p) {
                            return DropdownMenuItem<String>(
                              value: p['id'].toString(),
                              child: Row(
                                children: [
                                  Container(
                                    width: 32,
                                    height: 32,
                                    margin: const EdgeInsets.only(right: 12),
                                    decoration: BoxDecoration(
                                      color: Colors.white,
                                      borderRadius: BorderRadius.circular(6),
                                      border: Border.all(
                                          color: Colors.grey.shade200),
                                    ),
                                    child: ClipRRect(
                                      borderRadius: BorderRadius.circular(6),
                                      child: _buildImage(
                                          p['image_url']?.toString()),
                                    ),
                                  ),
                                  ConstrainedBox(
                                    constraints: BoxConstraints(
                                      maxWidth:
                                          MediaQuery.of(context).size.width -
                                              130,
                                    ),
                                    child: Text(
                                      p['name']?.toString() ?? '',
                                      style: GoogleFonts.outfit(
                                          fontWeight: FontWeight.w700,
                                          fontSize: 13,
                                          color: const Color(0xFF1E293B)),
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ),
                                ],
                              ),
                            );
                          }).toList(),
                          onChanged: (newId) {
                            if (newId != null) {
                              final selectedProduct = _allProducts.firstWhere(
                                  (p) => p['id'].toString() == newId);
                              setState(() {
                                item.productId = newId;
                                item.productName = selectedProduct['name'];
                                item.priceCtrl.text =
                                    ((selectedProduct['selling_price'] as num?)
                                                ?.toDouble() ??
                                            0.0)
                                        .toStringAsFixed(0);
                                item.piecesPerUnitCtrl.text =
                                    ((selectedProduct['pcs_per_unit'] as num?)
                                                ?.toInt() ??
                                            1)
                                        .toString();
                                item.unit =
                                    selectedProduct['unit']?.toString() ??
                                        'PCS';
                              });
                            }
                          },
                        ),
                ),
                IconButton(
                  icon: Icon(Icons.delete_outline_rounded,
                      color: Colors.red.shade400, size: 22),
                  onPressed: () => _confirmRemove(index),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                ),
              ],
            ),
            const SizedBox(height: 12),

            // Fields: Jumlah, PPU, Harga, Diskon
            Row(
              children: [
                Expanded(
                  child: _buildInlineField(
                    label: 'Jumlah (${item.unit})',
                    ctrl: item.orderedQtyCtrl,
                    onChanged: (_) => setState(() {}),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: _buildInlineField(
                    label: 'Isi per Satuan',
                    ctrl: item.piecesPerUnitCtrl,
                    onChanged: (_) => setState(() {}),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: _buildInlineField(
                    label: 'Harga Satuan',
                    ctrl: item.priceCtrl,
                    prefix: 'Rp ',
                    onChanged: (_) => setState(() {}),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: _buildInlineField(
                    label: 'Diskon (%)',
                    ctrl: item.discountCtrl,
                    suffix: '%',
                    onChanged: (_) => setState(() {}),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),

            // Subtotal
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: const Color(0xFFF0FDF4),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Subtotal',
                      style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: const Color(0xFF059669),
                          fontWeight: FontWeight.w600)),
                  Text(_currency.format(_calcSubtotal(item)),
                      style: GoogleFonts.outfit(
                          fontSize: 14,
                          color: const Color(0xFF059669),
                          fontWeight: FontWeight.w800)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInlineField({
    required String label,
    required TextEditingController ctrl,
    String? prefix,
    String? suffix,
    ValueChanged<String>? onChanged,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label,
            style: GoogleFonts.outfit(
                fontSize: 10,
                color: Colors.blueGrey,
                fontWeight: FontWeight.w600)),
        const SizedBox(height: 4),
        TextField(
          controller: ctrl,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          onChanged: onChanged,
          style: GoogleFonts.outfit(fontWeight: FontWeight.w700, fontSize: 14),
          decoration: InputDecoration(
            prefixText: prefix,
            suffixText: suffix,
            prefixStyle:
                GoogleFonts.outfit(fontSize: 13, color: Colors.blueGrey),
            suffixStyle:
                GoogleFonts.outfit(fontSize: 13, color: Colors.blueGrey),
            filled: true,
            fillColor: const Color(0xFFF8FAFC),
            border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(10),
                borderSide: BorderSide.none),
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 10, vertical: 10),
          ),
        ),
      ],
    );
  }

  void _confirmRemove(int index) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Hapus Item?',
            style: GoogleFonts.outfit(fontWeight: FontWeight.w800)),
        content: Text(
            'Produk "${_items[index].productName}" akan dihapus dari pesanan ini.',
            style: GoogleFonts.outfit()),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: Text('Batal', style: GoogleFonts.outfit())),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
                backgroundColor: Colors.red,
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12))),
            onPressed: () {
              Navigator.pop(ctx);
              _removeItem(index);
            },
            child: Text('Hapus',
                style: GoogleFonts.outfit(
                    color: Colors.white, fontWeight: FontWeight.w700)),
          ),
        ],
      ),
    );
  }

  Widget _buildBottomBar() {
    return Container(
      padding: const EdgeInsets.fromLTRB(20, 12, 20, 20),
      decoration: const BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
              color: Colors.black12, blurRadius: 12, offset: Offset(0, -4))
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text('Total Pesanan',
                    style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: Colors.blueGrey,
                        fontWeight: FontWeight.w600)),
                Text(_currency.format(_grandTotal),
                    style: GoogleFonts.outfit(
                        fontSize: 22,
                        fontWeight: FontWeight.w900,
                        color: const Color(0xFF10B981))),
              ],
            ),
          ),
          const SizedBox(width: 16),
          SizedBox(
            height: 52,
            child: ElevatedButton.icon(
              style: ElevatedButton.styleFrom(
                minimumSize: const Size(120, 52),
                backgroundColor: const Color(0xFF10B981),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16)),
                padding: const EdgeInsets.symmetric(horizontal: 24),
              ),
              onPressed: _isSaving ? null : _confirmSave,
              icon: _isSaving
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(
                          strokeWidth: 2, color: Colors.white))
                  : const Icon(Icons.check_circle_rounded,
                      color: Colors.white, size: 22),
              label: Text(
                _isSaving ? 'Menyimpan...' : 'SIMPAN KOREKSI',
                style: GoogleFonts.outfit(
                    color: Colors.white,
                    fontWeight: FontWeight.w800,
                    fontSize: 14),
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _confirmSave() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Simpan Koreksi?',
            style: GoogleFonts.outfit(fontWeight: FontWeight.w800)),
        content: Text(
            'Data pesanan akan diperbarui sesuai koreksi Anda. Total baru: ${_currency.format(_grandTotal)}',
            style: GoogleFonts.outfit()),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: Text('Batal', style: GoogleFonts.outfit())),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF10B981),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12))),
            onPressed: () {
              Navigator.pop(ctx);
              _saveChanges();
            },
            child: Text('Simpan',
                style: GoogleFonts.outfit(
                    color: Colors.white, fontWeight: FontWeight.w700)),
          ),
        ],
      ),
    );
  }
}

/// Model helper untuk setiap item yang sedang diedit.
class _EditableItem {
  String productId;
  String productName;
  final TextEditingController orderedQtyCtrl;
  final TextEditingController piecesPerUnitCtrl;
  final TextEditingController priceCtrl;
  final TextEditingController discountCtrl;
  String unit;

  _EditableItem({
    required this.productId,
    required this.productName,
    required this.orderedQtyCtrl,
    required this.piecesPerUnitCtrl,
    required this.priceCtrl,
    required this.discountCtrl,
    required this.unit,
  });

  void dispose() {
    orderedQtyCtrl.dispose();
    piecesPerUnitCtrl.dispose();
    priceCtrl.dispose();
    discountCtrl.dispose();
  }
}
