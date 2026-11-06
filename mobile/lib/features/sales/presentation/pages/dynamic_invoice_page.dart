import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:share_plus/share_plus.dart';
import 'dart:io' show File, Directory;
import 'package:flutter/foundation.dart' show kIsWeb;
import 'return_selection_page.dart';

/// DynamicInvoicePage — Faktur Penjualan Dinamis yang menyesuaikan
/// tampilan berdasarkan tipe retur barang:
///  - POTONG_NOTA: Tampilkan baris Kredit Nota / Pengurangan
///  - TUKAR_LANGSUNG: Tampilkan tabel tukar dengan produk pengganti
///  - TUKAR_SUSULAN: Tampilkan tabel retur + catatan "pengiriman susulan"
///  - TUKAR_PRODUK: Tampilkan tabel substitusi + selisih harga
///  - GARANSI: Tampilkan bagian klaim garansi, nomor klaim, produk pengganti
class DynamicInvoicePage extends StatefulWidget {
  final Map<String, dynamic> orderData;
  final List<ReturnItemResult> returnResults;
  final String paymentMethod;
  final double payAmount;
  final DateTime? paymentDueDate;
  final String receivedBy;

  const DynamicInvoicePage({
    super.key,
    required this.orderData,
    required this.returnResults,
    required this.paymentMethod,
    required this.payAmount,
    this.paymentDueDate,
    this.receivedBy = '',
  });

  @override
  State<DynamicInvoicePage> createState() => _DynamicInvoicePageState();
}

class _DynamicInvoicePageState extends State<DynamicInvoicePage> {
  bool _isGeneratingPdf = false;

  static const _bgColor = Color(0xFFF8FAFC);
  static const _cardColor = Colors.white;
  static const _cardBorder = Color(0xFFE2E8F0);
  static const _accentColor = Colors.blueAccent;

  final _currFmt = NumberFormat.currency(
    locale: 'id_ID',
    symbol: 'Rp ',
    decimalDigits: 0,
  );

  // ─────────────────── helpers ───────────────────────────────

  List<dynamic> get _items => widget.orderData['items'] as List<dynamic>? ?? [];
  List<dynamic> get _bonusItems =>
      widget.orderData['bonus_items'] as List<dynamic>? ?? [];

  Map<String, ReturnItemResult> get _returnMap {
    final map = <String, ReturnItemResult>{};
    for (final r in widget.returnResults) {
      map[r.productId] = r;
    }
    return map;
  }

  double get _baseTotal => (widget.orderData['total_amount'] ?? 0.0).toDouble();

  double get _totalReturnDeduction {
    double d = 0;
    for (final r in widget.returnResults) {
      if (r.returnType == 'POTONG_NOTA') {
        d += r.returnedQuantity * r.piecesPerUnit * r.price;
      }
    }
    return d;
  }

  double get _totalReplacementAddition {
    // TUKAR_PRODUK does not affect invoice, so return 0
    return 0;
  }

  double get _finalTotal =>
      (_baseTotal - _totalReturnDeduction + _totalReplacementAddition).clamp(
        0,
        double.infinity,
      );

  static const _paymentLabels = {
    'CASH': 'Tunai (Cash)',
    'VA': 'Transfer Bank',
    'TEMPO': 'Tempo / Kredit',
    'QRIS': 'QRIS / Scan QR',
  };

  static const _returnTypeColors = {
    'POTONG_NOTA': Color(0xFFEF4444),
    'TUKAR_LANGSUNG': Color(0xFF06B6D4),
    'TUKAR_PRODUK': Color(0xFF06B6D4),
  };

  static const _returnTypeLabels = {
    'POTONG_NOTA': 'Potong Nota',
    'TUKAR_LANGSUNG': 'Tukar Langsung',
    'TUKAR_PRODUK': 'Tukar Produk Berbeda',
  };

  static const _returnTypeIcons = {
    'POTONG_NOTA': Icons.money_off_rounded,
    'TUKAR_LANGSUNG': Icons.swap_horizontal_circle_rounded,
    'TUKAR_PRODUK': Icons.compare_arrows_rounded,
  };

  // ─────────────────────── BUILD ──────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _bgColor,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_rounded, color: Colors.black87),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Faktur Penjualan Dinamis',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w800,
                color: const Color(0xFF1E293B),
                fontSize: 16,
              ),
            ),
            Text(
              widget.orderData['so_number'] ??
                  widget.orderData['receipt_no'] ??
                  '',
              style: GoogleFonts.outfit(color: Colors.grey, fontSize: 11),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: _isGeneratingPdf
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: Colors.blueAccent,
                    ),
                  )
                : const Icon(
                    Icons.picture_as_pdf_rounded,
                    color: Colors.black87,
                  ),
            tooltip: 'Bagikan PDF',
            onPressed: _isGeneratingPdf ? null : _generateAndSharePdf,
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildInvoiceHeader(),
            const SizedBox(height: 16),
            _buildParties(),
            const SizedBox(height: 16),
            _buildProductTable(),
            const SizedBox(height: 16),
            if (widget.returnResults.isNotEmpty) ...[
              _buildReturnSections(),
              const SizedBox(height: 16),
            ],
            _buildTotals(),
            const SizedBox(height: 16),
            _buildPaymentSection(),
            const SizedBox(height: 24),
            _buildFooter(),
            const SizedBox(height: 12),
            ElevatedButton.icon(
              onPressed: () {
                Navigator.pop(context);
              },
              icon: const Icon(
                Icons.arrow_back_rounded,
                color: Colors.white,
                size: 18,
              ),
              label: Text(
                'KEMBALI KE DETAIL PENGIRIMAN',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                  fontSize: 13,
                ),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: _accentColor,
                padding: const EdgeInsets.symmetric(vertical: 16),
                minimumSize: const Size(double.infinity, 50),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
              ),
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  // ── Invoice Header ────────────────────────────────────────────────────────

  Widget _buildInvoiceHeader() {
    final invoiceNo =
        widget.orderData['so_number'] ??
        widget.orderData['invoice_no'] ??
        'INV-?';
    final dateStr = DateFormat('dd MMMM yyyy', 'id_ID').format(DateTime.now());
    final companyName =
        widget.orderData['employee']?['company']?['name'] ??
        widget.orderData['company']?['name'] ??
        'Perusahaan Demo';

    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            _accentColor.withOpacity(0.06),
            const Color(0xFF8B5CF6).withOpacity(0.03),
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _accentColor.withOpacity(0.15)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  companyName.toUpperCase(),
                  style: GoogleFonts.outfit(
                    color: const Color(0xFF1E293B),
                    fontWeight: FontWeight.w900,
                    fontSize: 13,
                    letterSpacing: 0.5,
                  ),
                ),
                Text(
                  widget.orderData['company']?['address'] ?? '',
                  style: GoogleFonts.outfit(
                    color: Colors.black54,
                    fontSize: 10,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                if ((widget.orderData['company']?['tax_number'] ?? '')
                    .toString()
                    .isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Text(
                    'NPWP: ${widget.orderData['company']?['tax_number']}',
                    style: GoogleFonts.outfit(
                      color: Colors.black54,
                      fontSize: 10,
                    ),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: _accentColor,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  'FAKTUR PENJUALAN',
                  style: GoogleFonts.outfit(
                    color: Colors.white,
                    fontWeight: FontWeight.w900,
                    fontSize: 10,
                    letterSpacing: 1,
                  ),
                ),
              ),
              const SizedBox(height: 6),
              Text(
                invoiceNo,
                style: GoogleFonts.outfit(
                  color: const Color(0xFF1E293B),
                  fontWeight: FontWeight.w700,
                  fontSize: 12,
                ),
              ),
              Text(
                dateStr,
                style: GoogleFonts.outfit(color: Colors.black54, fontSize: 10),
              ),
              if (widget.returnResults.isNotEmpty) ...[
                const SizedBox(height: 8),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xFFEF4444).withOpacity(0.15),
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(
                      color: const Color(0xFFEF4444).withOpacity(0.5),
                    ),
                  ),
                  child: Text(
                    'ADA RETUR',
                    style: GoogleFonts.outfit(
                      color: const Color(0xFFEF4444),
                      fontSize: 9,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 1,
                    ),
                  ),
                ),
              ],
            ],
          ),
        ],
      ),
    );
  }

  // ── Parties ───────────────────────────────────────────────────────────────

  Widget _buildParties() {
    final storeName = widget.orderData['store']?['name'] ?? 'Toko';
    final storeAddress = widget.orderData['store']?['address'] ?? '';
    final payLabel =
        _paymentLabels[widget.paymentMethod] ?? widget.paymentMethod;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _cardColor,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _cardBorder),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'DITAGIHKAN KEPADA',
                  style: GoogleFonts.outfit(
                    color: Colors.black45,
                    fontSize: 9,
                    fontWeight: FontWeight.w700,
                    letterSpacing: 1,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  storeName,
                  style: GoogleFonts.outfit(
                    color: const Color(0xFF1E293B),
                    fontWeight: FontWeight.w700,
                    fontSize: 13,
                  ),
                ),
                if (storeAddress.isNotEmpty)
                  Text(
                    storeAddress,
                    style: GoogleFonts.outfit(
                      color: Colors.black54,
                      fontSize: 11,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                if (widget.receivedBy.isNotEmpty) ...[
                  const SizedBox(height: 4),
                  Text(
                    'Diterima: ${widget.receivedBy}',
                    style: GoogleFonts.outfit(
                      color: Colors.black54,
                      fontSize: 11,
                    ),
                  ),
                ],
              ],
            ),
          ),
          Container(
            width: 1,
            height: 60,
            color: _cardBorder,
            margin: const EdgeInsets.symmetric(horizontal: 12),
          ),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  'PEMBAYARAN',
                  style: GoogleFonts.outfit(
                    color: Colors.black45,
                    fontSize: 9,
                    fontWeight: FontWeight.w700,
                    letterSpacing: 1,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  payLabel.toUpperCase(),
                  style: GoogleFonts.outfit(
                    color: const Color(0xFF1E293B),
                    fontWeight: FontWeight.w700,
                    fontSize: 13,
                  ),
                ),
                if (widget.paymentMethod == 'TEMPO' &&
                    widget.paymentDueDate != null) ...[
                  const SizedBox(height: 2),
                  Text(
                    'JT: ${DateFormat('dd MMM yyyy').format(widget.paymentDueDate!)}',
                    style: GoogleFonts.outfit(
                      color: const Color(0xFFF59E0B),
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
                const SizedBox(height: 2),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 3,
                  ),
                  decoration: BoxDecoration(
                    color: widget.paymentMethod == 'TEMPO'
                        ? const Color(0xFFF59E0B).withOpacity(0.15)
                        : const Color(0xFF10B981).withOpacity(0.15),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Text(
                    widget.paymentMethod == 'TEMPO'
                        ? 'BELUM LUNAS'
                        : (widget.paymentMethod == 'CASH'
                              ? 'LUNAS'
                              : 'MENUNGGU'),
                    style: GoogleFonts.outfit(
                      color: widget.paymentMethod == 'TEMPO'
                          ? const Color(0xFFF59E0B)
                          : const Color(0xFF10B981),
                      fontSize: 9,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 1,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ── Product Table ──────────────────────────────────────────────────────────

  Widget _buildProductTable() {
    final retMap = _returnMap;
    return Container(
      decoration: BoxDecoration(
        color: _cardColor,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _cardBorder),
      ),
      child: Column(
        children: [
          // Header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            decoration: BoxDecoration(
              color: Colors.black.withOpacity(0.02),
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(14),
              ),
            ),
            child: Row(
              children: [
                Expanded(
                  flex: 3,
                  child: Text(
                    'PRODUK',
                    style: GoogleFonts.outfit(
                      color: Colors.black45,
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                      letterSpacing: 0.5,
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Text(
                    'QTY',
                    textAlign: TextAlign.center,
                    style: GoogleFonts.outfit(
                      color: Colors.black45,
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Text(
                    'HARGA',
                    textAlign: TextAlign.right,
                    style: GoogleFonts.outfit(
                      color: Colors.black45,
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Text(
                    'SUBTOTAL',
                    textAlign: TextAlign.right,
                    style: GoogleFonts.outfit(
                      color: Colors.black45,
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ],
            ),
          ),
          // Items
          ..._items.map((item) {
            final pId =
                item['product_id']?.toString() ?? item['id']?.toString() ?? '';
            final ret = retMap[pId];
            final name = item['product']?['name'] ?? 'Produk';
            final orderedQty = (item['ordered_quantity'] ?? 0) as int;
            final ppu = (item['pieces_per_unit'] ?? 1) as int;
            final price = (item['price'] ?? 0.0).toDouble();
            final retQty = ret?.returnedQuantity ?? 0;
            final actualQty = (ret?.returnType == 'POTONG_NOTA')
                ? orderedQty - retQty
                : orderedQty;
            final subtotal = actualQty * ppu * price;
            final hasReturn = retQty > 0;
            final retColor = hasReturn
                ? (_returnTypeColors[ret!.returnType] ?? Colors.grey)
                : Colors.transparent;

            return Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
              decoration: BoxDecoration(
                border: const Border(bottom: BorderSide(color: _cardBorder)),
                color: hasReturn
                    ? retColor.withOpacity(0.03)
                    : Colors.transparent,
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        flex: 3,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              name,
                              style: GoogleFonts.outfit(
                                color: const Color(0xFF1E293B),
                                fontSize: 12,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            if (hasReturn)
                              Padding(
                                padding: const EdgeInsets.only(top: 2),
                                child: Row(
                                  children: [
                                    Icon(
                                      _returnTypeIcons[ret!.returnType] ??
                                          Icons.undo_rounded,
                                      color: retColor,
                                      size: 11,
                                    ),
                                    const SizedBox(width: 4),
                                    Text(
                                      _returnTypeLabels[ret.returnType] ??
                                          ret.returnType,
                                      style: GoogleFonts.outfit(
                                        color: retColor,
                                        fontSize: 9.5,
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                                    Text(
                                      ' · ${ret.returnedQuantity} retur',
                                      style: GoogleFonts.outfit(
                                        color: Colors.black45,
                                        fontSize: 9.5,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                          ],
                        ),
                      ),
                      Expanded(
                        flex: 1,
                        child: Text(
                          (ret?.returnType == 'POTONG_NOTA')
                              ? '$actualQty/$orderedQty'
                              : '$orderedQty',
                          textAlign: TextAlign.center,
                          style: GoogleFonts.outfit(
                            color: (ret?.returnType == 'POTONG_NOTA')
                                ? const Color(0xFFEF4444)
                                : const Color(0xFF1E293B),
                            fontSize: 12,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                      ),
                      Expanded(
                        flex: 2,
                        child: Text(
                          _currFmt.format(price),
                          textAlign: TextAlign.right,
                          style: GoogleFonts.outfit(
                            color: Colors.black54,
                            fontSize: 11,
                          ),
                        ),
                      ),
                      Expanded(
                        flex: 2,
                        child: Text(
                          _currFmt.format(subtotal),
                          textAlign: TextAlign.right,
                          style: GoogleFonts.outfit(
                            color: const Color(0xFF1E293B),
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            );
          }),

          if (_bonusItems.isNotEmpty) ...[
            const Divider(color: _cardBorder, height: 1),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              color: const Color(0xFFF8FAFC),
              child: Row(
                children: [
                  const Icon(
                    Icons.card_giftcard,
                    size: 14,
                    color: Color(0xFF0EA5E9),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    'Bonus / Hadiah',
                    style: GoogleFonts.outfit(
                      color: const Color(0xFF0EA5E9),
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
            ),
            const Divider(color: _cardBorder, height: 1),
            ..._bonusItems.map((item) {
              final prodName =
                  item['product_name'] ??
                  item['bonus_inventory']?['name'] ??
                  item['name'] ??
                  'Bonus';
              final qty = (item['quantity'] ?? item['qty'] ?? 0) as int;
              final unit =
                  item['unit'] ?? item['bonus_inventory']?['unit'] ?? 'PCS';

              return Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 12,
                ),
                decoration: const BoxDecoration(
                  border: Border(bottom: BorderSide(color: _cardBorder)),
                ),
                child: Row(
                  children: [
                    Expanded(
                      flex: 3,
                      child: Text(
                        prodName,
                        style: GoogleFonts.outfit(
                          color: const Color(0xFF1E293B),
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Text(
                        '$qty',
                        textAlign: TextAlign.center,
                        style: GoogleFonts.outfit(
                          color: const Color(0xFF1E293B),
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                    Expanded(
                      flex: 2,
                      child: Text(
                        '-',
                        textAlign: TextAlign.right,
                        style: GoogleFonts.outfit(
                          color: Colors.black54,
                          fontSize: 11,
                        ),
                      ),
                    ),
                    Expanded(
                      flex: 2,
                      child: Text(
                        'Gratis',
                        textAlign: TextAlign.right,
                        style: GoogleFonts.outfit(
                          color: const Color(0xFF0EA5E9),
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ),
              );
            }),
          ],
        ],
      ),
    );
  }

  // ── Return Sections (per type) ────────────────────────────────────────────

  Widget _buildReturnSections() {
    final byType = <String, List<ReturnItemResult>>{};
    for (final r in widget.returnResults) {
      byType.putIfAbsent(r.returnType, () => []).add(r);
    }

    return Column(
      children: byType.entries.map((entry) {
        switch (entry.key) {
          case 'POTONG_NOTA':
            return _buildPotongNotaSection(entry.value);
          case 'TUKAR_LANGSUNG':
            return _buildTukarSection(
              entry.value,
              'TUKAR_LANGSUNG',
              'Tukar Langsung di Tempat',
            );
          case 'TUKAR_SUSULAN':
            return _buildTukarSusulanSection(entry.value);
          case 'TUKAR_PRODUK':
            return _buildTukarProdukSection(entry.value);
          default:
            return const SizedBox.shrink();
        }
      }).toList(),
    );
  }

  Widget _buildReturnSectionCard({
    required String typeKey,
    required String title,
    required Widget child,
    String? subtitle,
  }) {
    final color = _returnTypeColors[typeKey] ?? Colors.grey;
    final icon = _returnTypeIcons[typeKey] ?? Icons.undo_rounded;
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: color.withOpacity(0.04),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: color.withOpacity(0.2)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.fromLTRB(14, 12, 14, 10),
            decoration: BoxDecoration(
              color: color.withOpacity(0.08),
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(13),
              ),
              border: Border(
                bottom: BorderSide(color: color.withOpacity(0.12)),
              ),
            ),
            child: Row(
              children: [
                Icon(icon, color: color, size: 16),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title,
                        style: GoogleFonts.outfit(
                          color: color,
                          fontWeight: FontWeight.w800,
                          fontSize: 12,
                          letterSpacing: 0.5,
                        ),
                      ),
                      if (subtitle != null)
                        Text(
                          subtitle,
                          style: GoogleFonts.outfit(
                            color: color.withOpacity(0.7),
                            fontSize: 10,
                          ),
                        ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          Padding(padding: const EdgeInsets.all(14), child: child),
        ],
      ),
    );
  }

  Widget _buildPotongNotaSection(List<ReturnItemResult> items) {
    double totalCut = 0;
    for (final r in items) {
      totalCut += r.returnedQuantity * r.piecesPerUnit * r.price;
    }
    return _buildReturnSectionCard(
      typeKey: 'POTONG_NOTA',
      title: 'KREDIT NOTA / PENGURANGAN TAGIHAN',
      subtitle: 'Barang dikembalikan, nilai nota dipotong',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ...items.map(
            (r) => Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Row(
                children: [
                  const Icon(
                    Icons.remove_circle_rounded,
                    color: Color(0xFFEF4444),
                    size: 14,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      r.productName,
                      style: GoogleFonts.outfit(
                        color: const Color(0xFF1E293B),
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                  Text(
                    '${r.returnedQuantity} ${r.unit} × ${_currFmt.format(r.price)}',
                    style: GoogleFonts.outfit(
                      color: Colors.black54,
                      fontSize: 11,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(
                    '-${_currFmt.format(r.returnedQuantity * r.piecesPerUnit * r.price)}',
                    style: GoogleFonts.outfit(
                      color: const Color(0xFFEF4444),
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const Divider(color: Color(0xFFEF4444), height: 16),
          Row(
            children: [
              const Spacer(),
              Text(
                'Total Pengurangan:',
                style: GoogleFonts.outfit(color: Colors.black54, fontSize: 12),
              ),
              const SizedBox(width: 8),
              Text(
                '-${_currFmt.format(totalCut)}',
                style: GoogleFonts.outfit(
                  color: const Color(0xFFEF4444),
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildTukarSection(
    List<ReturnItemResult> items,
    String typeKey,
    String title,
  ) {
    return _buildReturnSectionCard(
      typeKey: typeKey,
      title: title.toUpperCase(),
      subtitle: 'Tagihan tetap penuh. Barang rusak kembali ke gudang.',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ...items.map(
            (r) => Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(
                    Icons.swap_horiz_rounded,
                    color: Color(0xFF06B6D4),
                    size: 14,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          r.productName,
                          style: GoogleFonts.outfit(
                            color: const Color(0xFF1E293B),
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        Text(
                          '${r.returnedQuantity} ${r.unit} · Alasan: ${r.returnReason}',
                          style: GoogleFonts.outfit(
                            color: Colors.black45,
                            fontSize: 10.5,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 3,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFF06B6D4).withOpacity(0.15),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      'TAGIHAN TETAP',
                      style: GoogleFonts.outfit(
                        color: const Color(0xFF06B6D4),
                        fontSize: 9,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTukarSusulanSection(List<ReturnItemResult> items) {
    return _buildReturnSectionCard(
      typeKey: 'TUKAR_SUSULAN',
      title: 'RETUR TUKAR SUSULAN',
      subtitle:
          'Tagihan dipotong. Pengganti dikirim & ditagih pada SJ berikutnya.',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ...items.map(
            (r) => Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(
                    Icons.schedule_send_rounded,
                    color: Color(0xFF8B5CF6),
                    size: 14,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          r.productName,
                          style: GoogleFonts.outfit(
                            color: const Color(0xFF1E293B),
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        Text(
                          '${r.returnedQuantity} ${r.unit} dipotong dari pengiriman ini',
                          style: GoogleFonts.outfit(
                            color: Colors.black45,
                            fontSize: 10.5,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          Container(
            margin: const EdgeInsets.only(top: 8),
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: const Color(0xFF8B5CF6).withOpacity(0.08),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: const Color(0xFF8B5CF6).withOpacity(0.2),
              ),
            ),
            child: Row(
              children: [
                const Icon(
                  Icons.info_outline_rounded,
                  color: Color(0xFF8B5CF6),
                  size: 14,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Pengiriman susulan & tagihan barunya akan dijadwalkan oleh admin.',
                    style: GoogleFonts.outfit(
                      color: const Color(0xFF8B5CF6),
                      fontSize: 10.5,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTukarProdukSection(List<ReturnItemResult> items) {
    return _buildReturnSectionCard(
      typeKey: 'TUKAR_PRODUK',
      title: 'SUBSTITUSI PRODUK',
      subtitle:
          'Produk diganti langsung dengan item berbeda dari muatan armada.',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ...items.map(
            (r) => Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(
                        Icons.arrow_upward_rounded,
                        color: Color(0xFFEF4444),
                        size: 13,
                      ),
                      const SizedBox(width: 6),
                      Text(
                        'Diretur: ',
                        style: GoogleFonts.outfit(
                          color: Colors.black45,
                          fontSize: 11,
                        ),
                      ),
                      Text(
                        '${r.productName} (${r.returnedQuantity} ${r.unit})',
                        style: GoogleFonts.outfit(
                          color: const Color(0xFF1E293B),
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      const Icon(
                        Icons.arrow_downward_rounded,
                        color: Color(0xFF10B981),
                        size: 13,
                      ),
                      const SizedBox(width: 6),
                      Text(
                        'Pengganti: ',
                        style: GoogleFonts.outfit(
                          color: Colors.black45,
                          fontSize: 11,
                        ),
                      ),
                      Expanded(
                        child: Text(
                          r.replacementProductName != null
                              ? '${r.replacementProductName} (${r.replacementQuantity ?? r.returnedQuantity} ${r.unit})'
                              : '— belum dipilih —',
                          style: GoogleFonts.outfit(
                            color: const Color(0xFF06B6D4),
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ],
                  ),
                  if (r.replacementPrice != null) ...[
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        const SizedBox(width: 19),
                        Text(
                          'Harga Original: ',
                          style: GoogleFonts.outfit(
                            color: Colors.black45,
                            fontSize: 10,
                          ),
                        ),
                        Text(
                          _currFmt.format(r.price),
                          style: GoogleFonts.outfit(
                            color: const Color(0xFF1E293B),
                            fontSize: 10,
                          ),
                        ),
                        const SizedBox(width: 10),
                        Text(
                          'Harga Baru: ',
                          style: GoogleFonts.outfit(
                            color: Colors.black45,
                            fontSize: 10,
                          ),
                        ),
                        Text(
                          _currFmt.format(r.replacementPrice),
                          style: GoogleFonts.outfit(
                            color: const Color(0xFF10B981),
                            fontSize: 10,
                          ),
                        ),
                      ],
                    ),
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ── Totals ────────────────────────────────────────────────────────────────

  Widget _buildTotals() {
    final discount = (widget.orderData['discount_amount'] ?? 0.0).toDouble();
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _cardColor,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _cardBorder),
      ),
      child: Column(
        children: [
          _totalRow('Subtotal', _baseTotal, Colors.black54, fontSize: 12),
          if (discount > 0) ...[
            const SizedBox(height: 6),
            _totalRow(
              'Diskon',
              -discount,
              const Color(0xFF10B981),
              fontSize: 12,
            ),
          ],
          if (_totalReturnDeduction > 0) ...[
            const SizedBox(height: 6),
            _totalRow(
              'Total Pengurangan Retur',
              -_totalReturnDeduction,
              const Color(0xFFEF4444),
              fontSize: 12,
            ),
          ],
          if (_totalReplacementAddition > 0) ...[
            const SizedBox(height: 6),
            _totalRow(
              'Tambahan Tukar Langsung',
              _totalReplacementAddition,
              const Color(0xFF10B981),
              fontSize: 12,
            ),
          ],
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 10),
            child: Divider(color: _cardBorder),
          ),
          _totalRow(
            'TOTAL TAGIHAN',
            _finalTotal - discount,
            const Color(0xFF1E293B),
            fontSize: 16,
            bold: true,
          ),
          if (widget.payAmount > 0 && widget.paymentMethod != 'TEMPO') ...[
            const SizedBox(height: 8),
            _totalRow(
              'Dibayar',
              widget.payAmount,
              const Color(0xFF10B981),
              fontSize: 13,
            ),
            const SizedBox(height: 4),
            _totalRow(
              'Sisa Tagihan',
              (_finalTotal - discount - widget.payAmount).clamp(
                0,
                double.infinity,
              ),
              const Color(0xFFF59E0B),
              fontSize: 13,
            ),
          ],
        ],
      ),
    );
  }

  Widget _totalRow(
    String label,
    double amount,
    Color color, {
    double fontSize = 12,
    bool bold = false,
  }) {
    return Row(
      children: [
        Text(
          label,
          style: GoogleFonts.outfit(color: Colors.black54, fontSize: fontSize),
        ),
        const Spacer(),
        Text(
          '${amount < 0 ? '-' : ''}${_currFmt.format(amount.abs())}',
          style: GoogleFonts.outfit(
            color: color,
            fontSize: fontSize,
            fontWeight: bold ? FontWeight.w900 : FontWeight.w600,
          ),
        ),
      ],
    );
  }

  // ── Payment Section ───────────────────────────────────────────────────────

  Widget _buildPaymentSection() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF10B981).withOpacity(0.08),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFF10B981).withOpacity(0.2)),
      ),
      child: Row(
        children: [
          const Icon(
            Icons.payments_rounded,
            color: Color(0xFF10B981),
            size: 20,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Metode Pembayaran',
                  style: GoogleFonts.outfit(
                    color: Colors.black45,
                    fontSize: 11,
                  ),
                ),
                Text(
                  _paymentLabels[widget.paymentMethod] ?? widget.paymentMethod,
                  style: GoogleFonts.outfit(
                    color: const Color(0xFF10B981),
                    fontWeight: FontWeight.w700,
                    fontSize: 14,
                  ),
                ),
                if (widget.paymentMethod == 'TEMPO' &&
                    widget.paymentDueDate != null)
                  Text(
                    'Jatuh Tempo: ${DateFormat('dd MMMM yyyy').format(widget.paymentDueDate!)}',
                    style: GoogleFonts.outfit(
                      color: const Color(0xFFF59E0B),
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
              ],
            ),
          ),
          Text(
            _currFmt.format(widget.payAmount),
            style: GoogleFonts.outfit(
              color: const Color(0xFF10B981),
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }

  // ── Footer ────────────────────────────────────────────────────────────────

  Widget _buildFooter() {
    return Column(
      children: [
        const Divider(color: _cardBorder),
        const SizedBox(height: 8),
        Text(
          'Terima kasih atas kepercayaan dan kerja sama Anda.',
          textAlign: TextAlign.center,
          style: GoogleFonts.outfit(
            color: Colors.black45,
            fontSize: 11,
            fontStyle: FontStyle.italic,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          'Dokumen ini sah tanpa tanda tangan basah.',
          textAlign: TextAlign.center,
          style: GoogleFonts.outfit(color: Colors.black38, fontSize: 10),
        ),
      ],
    );
  }

  // ── PDF Generation ────────────────────────────────────────────────────────

  Future<void> _generateAndSharePdf() async {
    setState(() => _isGeneratingPdf = true);
    try {
      final pdf = pw.Document();
      final invoiceNo =
          widget.orderData['so_number'] ??
          widget.orderData['invoice_no'] ??
          'INV';
      final companyName =
          widget.orderData['employee']?['company']?['name'] ??
          widget.orderData['company']?['name'] ??
          'Perusahaan Demo';
      final storeName = widget.orderData['store']?['name'] ?? 'Toko';
      final dateStr = DateFormat('dd MMMM yyyy').format(DateTime.now());
      final retMap = _returnMap;
      final discount = (widget.orderData['discount_amount'] ?? 0.0).toDouble();

      pdf.addPage(
        pw.Page(
          pageFormat: PdfPageFormat.a4,
          margin: const pw.EdgeInsets.symmetric(vertical: 20, horizontal: 28),
          build: (ctx) => pw.Column(
            crossAxisAlignment: pw.CrossAxisAlignment.start,
            children: [
              // Header
              pw.Row(
                mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                children: [
                  pw.Column(
                    crossAxisAlignment: pw.CrossAxisAlignment.start,
                    children: [
                      pw.Text(
                        companyName.toUpperCase(),
                        style: pw.TextStyle(
                          fontSize: 9,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                      pw.Text(
                        widget.orderData['company']?['address'] ?? '',
                        style: const pw.TextStyle(fontSize: 7),
                      ),
                    ],
                  ),
                  pw.Column(
                    crossAxisAlignment: pw.CrossAxisAlignment.end,
                    children: [
                      pw.Text(
                        'FAKTUR PENJUALAN',
                        style: pw.TextStyle(
                          fontSize: 11,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                      pw.Text(
                        invoiceNo,
                        style: pw.TextStyle(
                          fontSize: 8,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                      pw.Text(
                        'Tanggal: $dateStr',
                        style: const pw.TextStyle(fontSize: 7),
                      ),
                      if (widget.returnResults.isNotEmpty)
                        pw.Text(
                          '** ADA RETUR BARANG **',
                          style: pw.TextStyle(
                            fontSize: 7,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.red,
                          ),
                        ),
                    ],
                  ),
                ],
              ),
              pw.Divider(thickness: 1.0, color: PdfColors.grey400, height: 14),

              // Parties
              pw.Row(
                mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                children: [
                  pw.Column(
                    crossAxisAlignment: pw.CrossAxisAlignment.start,
                    children: [
                      pw.Text(
                        'DITAGIHKAN KEPADA:',
                        style: const pw.TextStyle(
                          fontSize: 7,
                          color: PdfColors.grey600,
                        ),
                      ),
                      pw.Text(
                        storeName,
                        style: pw.TextStyle(
                          fontSize: 9,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                      pw.Text(
                        widget.orderData['store']?['address'] ?? '',
                        style: const pw.TextStyle(fontSize: 7),
                      ),
                    ],
                  ),
                  pw.Column(
                    crossAxisAlignment: pw.CrossAxisAlignment.end,
                    children: [
                      pw.Text(
                        'PEMBAYARAN:',
                        style: const pw.TextStyle(
                          fontSize: 7,
                          color: PdfColors.grey600,
                        ),
                      ),
                      pw.Text(
                        (_paymentLabels[widget.paymentMethod] ??
                                widget.paymentMethod)
                            .toUpperCase(),
                        style: pw.TextStyle(
                          fontSize: 9,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                      if (widget.paymentMethod == 'TEMPO' &&
                          widget.paymentDueDate != null)
                        pw.Text(
                          'JT: ${DateFormat('dd MMM yyyy').format(widget.paymentDueDate!)}',
                          style: const pw.TextStyle(fontSize: 7),
                        ),
                    ],
                  ),
                ],
              ),
              pw.SizedBox(height: 10),

              // Product table
              pw.Container(
                color: PdfColors.grey200,
                padding: const pw.EdgeInsets.symmetric(
                  vertical: 4,
                  horizontal: 6,
                ),
                child: pw.Row(
                  children: [
                    pw.Expanded(
                      flex: 3,
                      child: pw.Text(
                        'PRODUK',
                        style: pw.TextStyle(
                          fontSize: 7,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                    ),
                    pw.Expanded(
                      flex: 1,
                      child: pw.Text(
                        'QTY',
                        textAlign: pw.TextAlign.center,
                        style: pw.TextStyle(
                          fontSize: 7,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                    ),
                    pw.Expanded(
                      flex: 2,
                      child: pw.Text(
                        'HARGA',
                        textAlign: pw.TextAlign.right,
                        style: pw.TextStyle(
                          fontSize: 7,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                    ),
                    pw.Expanded(
                      flex: 2,
                      child: pw.Text(
                        'SUBTOTAL',
                        textAlign: pw.TextAlign.right,
                        style: pw.TextStyle(
                          fontSize: 7,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              ..._items.map((item) {
                final pId =
                    item['product_id']?.toString() ??
                    item['id']?.toString() ??
                    '';
                final ret = retMap[pId];
                final name = item['product']?['name'] ?? 'Produk';
                final orderedQty = (item['ordered_quantity'] ?? 0) as int;
                final ppu = (item['pieces_per_unit'] ?? 1) as int;
                final price = (item['price'] ?? 0.0).toDouble();
                final retQty = ret?.returnedQuantity ?? 0;
                final actualQty = (ret?.returnType == 'POTONG_NOTA')
                    ? orderedQty - retQty
                    : orderedQty;
                final subtotal = actualQty * ppu * price;
                return pw.Container(
                  padding: const pw.EdgeInsets.symmetric(
                    vertical: 3,
                    horizontal: 6,
                  ),
                  decoration: const pw.BoxDecoration(
                    border: pw.Border(
                      bottom: pw.BorderSide(color: PdfColors.grey300),
                    ),
                  ),
                  child: pw.Row(
                    children: [
                      pw.Expanded(
                        flex: 3,
                        child: pw.Column(
                          crossAxisAlignment: pw.CrossAxisAlignment.start,
                          children: [
                            pw.Text(
                              name,
                              style: const pw.TextStyle(fontSize: 7),
                            ),
                            if (retQty > 0)
                              pw.Text(
                                '[${_returnTypeLabels[ret!.returnType] ?? ret.returnType}: $retQty retur]',
                                style: const pw.TextStyle(
                                  fontSize: 6,
                                  color: PdfColors.red,
                                ),
                              ),
                          ],
                        ),
                      ),
                      pw.Expanded(
                        flex: 1,
                        child: pw.Text(
                          (ret?.returnType == 'POTONG_NOTA')
                              ? '$actualQty/$orderedQty'
                              : '$orderedQty',
                          textAlign: pw.TextAlign.center,
                          style: const pw.TextStyle(fontSize: 7),
                        ),
                      ),
                      pw.Expanded(
                        flex: 2,
                        child: pw.Text(
                          _currFmt.format(price),
                          textAlign: pw.TextAlign.right,
                          style: const pw.TextStyle(fontSize: 7),
                        ),
                      ),
                      pw.Expanded(
                        flex: 2,
                        child: pw.Text(
                          _currFmt.format(subtotal),
                          textAlign: pw.TextAlign.right,
                          style: const pw.TextStyle(fontSize: 7),
                        ),
                      ),
                    ],
                  ),
                );
              }),
              pw.SizedBox(height: 8),

              // Return notes
              if (widget.returnResults.isNotEmpty) ...[
                pw.Text(
                  'CATATAN RETUR:',
                  style: pw.TextStyle(
                    fontSize: 7,
                    fontWeight: pw.FontWeight.bold,
                    color: PdfColors.red,
                  ),
                ),
                ...widget.returnResults.map(
                  (r) => pw.Text(
                    '• ${r.productName} (${r.returnedQuantity} ${r.unit}): ${_returnTypeLabels[r.returnType] ?? r.returnType} — ${r.returnReason}',
                    style: const pw.TextStyle(
                      fontSize: 6.5,
                      color: PdfColors.red700,
                    ),
                  ),
                ),
                pw.SizedBox(height: 6),
              ],

              // Totals
              pw.Row(
                mainAxisAlignment: pw.MainAxisAlignment.end,
                children: [
                  pw.Column(
                    crossAxisAlignment: pw.CrossAxisAlignment.end,
                    children: [
                      pw.Text(
                        'Subtotal: ${_currFmt.format(_baseTotal)}',
                        style: const pw.TextStyle(fontSize: 8),
                      ),
                      if (discount > 0)
                        pw.Text(
                          'Diskon: -${_currFmt.format(discount)}',
                          style: const pw.TextStyle(
                            fontSize: 8,
                            color: PdfColors.green700,
                          ),
                        ),
                      if (_totalReturnDeduction > 0)
                        pw.Text(
                          'Retur (POTONG_NOTA): -${_currFmt.format(_totalReturnDeduction)}',
                          style: const pw.TextStyle(
                            fontSize: 8,
                            color: PdfColors.red,
                          ),
                        ),
                      pw.Text(
                        'TOTAL: ${_currFmt.format(_finalTotal - discount)}',
                        style: pw.TextStyle(
                          fontSize: 10,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              pw.Spacer(),
              pw.Center(
                child: pw.Text(
                  'Terima kasih atas kepercayaan dan kerja sama Anda.',
                  style: pw.TextStyle(
                    fontSize: 7,
                    fontStyle: pw.FontStyle.italic,
                    color: PdfColors.grey600,
                  ),
                ),
              ),
            ],
          ),
        ),
      );

      final pdfBytes = await pdf.save();
      if (kIsWeb) {
        final xFile = XFile.fromData(
          pdfBytes,
          mimeType: 'application/pdf',
          name: 'Faktur_$invoiceNo.pdf',
        );
        await Share.shareXFiles([
          xFile,
        ], text: 'Faktur Penjualan $invoiceNo dari $companyName');
      } else {
        final dir = Directory.systemTemp;
        final file = File('${dir.path}/Faktur_$invoiceNo.pdf');
        await file.writeAsBytes(pdfBytes);
        final xFile = XFile(file.path, mimeType: 'application/pdf');
        await Share.shareXFiles([
          xFile,
        ], text: 'Faktur Penjualan $invoiceNo dari $companyName');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal membuat PDF: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isGeneratingPdf = false);
    }
  }
}
