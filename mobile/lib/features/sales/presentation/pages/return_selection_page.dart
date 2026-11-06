import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';

// ─────────────────────────────────────────────
// DATA MODEL
// ─────────────────────────────────────────────

class ReturnItemResult {
  final String productId;
  final String productName;
  final int returnedQuantity;
  final String returnReason;
  final String returnType;
  final int orderedQuantity;
  final double price;
  final String unit;
  final int piecesPerUnit;
  final String? replacementProductId;
  final String? replacementProductName;
  final int? replacementQuantity;
  final double? replacementPrice;
  final String? claimReferenceNo;
  final String? damageEvidenceUrl;

  const ReturnItemResult({
    required this.productId,
    required this.productName,
    required this.returnedQuantity,
    required this.returnReason,
    required this.returnType,
    required this.orderedQuantity,
    required this.price,
    required this.unit,
    required this.piecesPerUnit,
    this.replacementProductId,
    this.replacementProductName,
    this.replacementQuantity,
    this.replacementPrice,
    this.claimReferenceNo,
    this.damageEvidenceUrl,
  });

  Map<String, dynamic> toJson() {
    final map = <String, dynamic>{
      'product_id': productId,
      'returned_quantity': returnedQuantity,
      'return_reason': returnReason,
      'return_type': returnType,
    };
    if (returnType == 'TUKAR_PRODUK' && replacementProductId != null) {
      map['replacement_product_id'] = replacementProductId;
      map['replacement_quantity'] = replacementQuantity ?? returnedQuantity;
    }
    if (returnType == 'GARANSI') {
      if (claimReferenceNo != null && claimReferenceNo!.isNotEmpty) {
        map['claim_reference_no'] = claimReferenceNo;
      }
      if (damageEvidenceUrl != null && damageEvidenceUrl!.isNotEmpty) {
        map['damage_evidence_url'] = damageEvidenceUrl;
      }
    }
    return map;
  }
}

// ─────────────────────────────────────────────
// PAGE
// ─────────────────────────────────────────────

class ReturnSelectionPage extends StatefulWidget {
  final List<dynamic> items;
  final Map<String, int> initialReturnedQties;
  final Map<String, String> initialReturnReasons;
  final List<dynamic> availableProducts;

  const ReturnSelectionPage({
    super.key,
    required this.items,
    required this.initialReturnedQties,
    required this.initialReturnReasons,
    this.availableProducts = const [],
  });

  @override
  State<ReturnSelectionPage> createState() => _ReturnSelectionPageState();
}

class _ReturnSelectionPageState extends State<ReturnSelectionPage>
    with SingleTickerProviderStateMixin {
  final Map<String, _ReturnItemState> _states = {};
  late TabController _tabController;

  static const _primaryColor = Color(0xFF6366F1);
  static const _bgColor = Color(0xFFF8FAFC); // Latar belakang cerah
  static const _cardColor = Colors.white; // Card putih bersih
  static const _cardBorder = Color(0xFFE2E8F0); // Border abu-abu terang
  static const _textDark = Color(0xFF1E293B); // Teks utama gelap
  static const _textMuted = Color(0xFF64748B); // Teks sekunder
  static const _textLight = Color(0xFF94A3B8); // Teks petunjuk halus

  static const _returnTypeConfigs = <String, _ReturnTypeConfig>{
    'POTONG_NOTA': _ReturnTypeConfig(
      label: 'Potong Nota',
      description: 'Mengurangi nilai total nota saat ini sesuai jumlah retur.',
      icon: Icons.money_off_rounded,
      colorHex: 0xFFEF4444, // Red
      hint:
          'Pilihan terbaik jika pelanggan ingin segera mengurangi tagihan belanja mereka.',
    ),
    'TUKAR_PRODUK': _ReturnTypeConfig(
      label: 'Tukar Produk Langsung',
      description: 'Tukar dengan produk lain saat ini dari stok muatan armada.',
      icon: Icons.swap_horizontal_circle_rounded,
      colorHex: 0xFF06B6D4, // Cyan
      hint:
          'Pastikan stok produk pengganti tersedia di armada kendaraan Anda saat ini.',
    ),
    'GARANSI': _ReturnTypeConfig(
      label: 'Klaim Garansi / Retur Supplier',
      description:
          'Nota tetap penuh. Klaim diproses secara terpisah ke produsen/supplier.',
      icon: Icons.verified_user_rounded,
      colorHex: 0xFF10B981, // Emerald
      hint:
          'Gunakan jika produk memiliki cacat produksi yang ditanggung oleh supplier.',
    ),
  };

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: widget.items.length, vsync: this);

    for (final item in widget.items) {
      final pId =
          item['product_id']?.toString() ?? item['id']?.toString() ?? '';
      final orderedQty = item['ordered_quantity'] ?? 0;
      final retQty = widget.initialReturnedQties[pId] ?? 0;
      final retReason = widget.initialReturnReasons[pId] ?? 'RUSAK';

      // Pre-select return type based on existing database item if any
      String retType = 'POTONG_NOTA';
      if (item['return_type'] != null) {
        retType = item['return_type'].toString();
      }

      String claimRef = '';
      if (item['claim_reference_no'] != null) {
        claimRef = item['claim_reference_no'].toString();
      }

      final controller = TextEditingController(text: claimRef);
      _states[pId] = _ReturnItemState(
        orderedQty: orderedQty,
        returnedQty: retQty,
        returnType: retType,
        returnReason: retReason,
        claimRefController: controller,
      );
      _states[pId]!.claimRefNo = claimRef;

      if (item['replacement_product_id'] != null) {
        _states[pId]!.replacementProductId =
            item['replacement_product_id'].toString();
        _states[pId]!.replacementProductName =
            item['replacement_product']?['name']?.toString();
        _states[pId]!.replacementPrice =
            (item['replacement_price'] as num? ?? 0.0).toDouble();
      }
    }
  }

  @override
  void dispose() {
    _tabController.dispose();
    for (final s in _states.values) {
      s.claimRefController.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final currencyFmt =
        NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);
    return Scaffold(
      backgroundColor: _bgColor,
      appBar: _buildAppBar(),
      body: TabBarView(
        controller: _tabController,
        children: widget.items.map((item) {
          final pId =
              item['product_id']?.toString() ?? item['id']?.toString() ?? '';
          return _buildItemTab(item, pId, _states[pId]!, currencyFmt);
        }).toList(),
      ),
      bottomNavigationBar: _buildSummaryBar(currencyFmt),
    );
  }

  AppBar _buildAppBar() {
    return AppBar(
      backgroundColor: _cardColor,
      elevation: 0,
      iconTheme: const IconThemeData(color: _textDark),
      leading: IconButton(
        icon: const Icon(Icons.close_rounded, color: _textMuted),
        onPressed: () => Navigator.pop(context),
      ),
      title: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Mekanisme Retur Barang',
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800, color: _textDark, fontSize: 16)),
          Text('${widget.items.length} item dalam nota ini',
              style: GoogleFonts.outfit(color: _textMuted, fontSize: 11)),
        ],
      ),
      actions: [
        TextButton(
          onPressed: _onConfirm,
          child: Text('SIMPAN',
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  color: _primaryColor,
                  fontSize: 13,
                  letterSpacing: 1)),
        ),
      ],
      bottom: PreferredSize(
        preferredSize: const Size.fromHeight(44),
        child: Container(
          decoration: const BoxDecoration(
            color: _cardColor,
            border: Border(bottom: BorderSide(color: _cardBorder)),
          ),
          child: TabBar(
            controller: _tabController,
            isScrollable: true,
            indicatorColor: _primaryColor,
            indicatorWeight: 3,
            labelColor: _primaryColor,
            unselectedLabelColor: _textMuted,
            labelStyle:
                GoogleFonts.outfit(fontWeight: FontWeight.w700, fontSize: 12),
            unselectedLabelStyle:
                GoogleFonts.outfit(fontWeight: FontWeight.w500, fontSize: 12),
            tabs: widget.items.map((item) {
              final pId = item['product_id']?.toString() ??
                  item['id']?.toString() ??
                  '';
              final name = item['product']?['name'] ?? 'Produk';
              final short =
                  name.length > 14 ? '${name.substring(0, 14)}…' : name;
              final hasRet = (_states[pId]?.returnedQty ?? 0) > 0;
              return Tab(
                child: Row(mainAxisSize: MainAxisSize.min, children: [
                  if (hasRet)
                    Container(
                      width: 8,
                      height: 8,
                      margin: const EdgeInsets.only(right: 6),
                      decoration: const BoxDecoration(
                          color: Color(0xFFEF4444), shape: BoxShape.circle),
                    ),
                  Text(short),
                ]),
              );
            }).toList(),
          ),
        ),
      ),
    );
  }

  Widget _buildItemTab(Map<String, dynamic> item, String pId,
      _ReturnItemState state, NumberFormat currFmt) {
    final name = item['product']?['name'] ?? 'Produk';
    final orderedQty = (item['ordered_quantity'] ?? 0) as int;
    final ppu = (item['pieces_per_unit'] ?? 1) as int;
    final price = (item['price'] as num? ?? 0.0).toDouble();
    final unit = item['unit'] ?? 'PCS';
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 100),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildProductHeader(name, orderedQty, ppu, price, unit, currFmt),
          const SizedBox(height: 20),
          _buildQtySelector(pId, state, orderedQty, unit),
          const SizedBox(height: 24),
          if (state.returnedQty > 0) ...[
            _buildReturnTypeSelector(pId, state),
            const SizedBox(height: 20),
            _buildReturnReasonSelector(pId, state),
            const SizedBox(height: 20),
            _buildExtraFields(pId, state),
            const SizedBox(height: 20),
            _buildImpactPreview(state, orderedQty, ppu, price, unit, currFmt),
          ] else
            _buildNoReturnCard(name),
        ],
      ),
    );
  }

  // ── Product Header ──────────────────────────────────────────────────────

  Widget _buildProductHeader(String name, int orderedQty, int ppu, double price,
      String unit, NumberFormat currFmt) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _cardBorder),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 8,
              offset: const Offset(0, 4))
        ],
      ),
      child: Row(children: [
        Container(
          width: 52,
          height: 52,
          decoration: BoxDecoration(
            gradient: const LinearGradient(
                colors: [Color(0xFF6366F1), Color(0xFF8B5CF6)]),
            borderRadius: BorderRadius.circular(14),
          ),
          child: const Icon(Icons.inventory_2_rounded,
              color: Colors.white, size: 26),
        ),
        const SizedBox(width: 14),
        Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(name,
              style: GoogleFonts.outfit(
                  color: _textDark, fontWeight: FontWeight.w700, fontSize: 15)),
          const SizedBox(height: 6),
          Wrap(spacing: 8, children: [
            _chip('$orderedQty $unit', const Color(0xFF6366F1)),
            _chip('${currFmt.format(price)}/pcs', _textMuted),
            if (ppu > 1) _chip('$ppu pcs/$unit', _textMuted),
          ]),
        ])),
      ]),
    );
  }

  Widget _chip(String label, Color bg) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: bg.withOpacity(0.08),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: bg.withOpacity(0.2)),
      ),
      child: Text(label,
          style: GoogleFonts.outfit(
              color: bg, fontSize: 10, fontWeight: FontWeight.w700)),
    );
  }

  // ── Qty Selector ────────────────────────────────────────────────────────

  Widget _buildQtySelector(
      String pId, _ReturnItemState state, int maxQty, String unit) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _cardBorder),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 8,
              offset: const Offset(0, 4))
        ],
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text('JUMLAH YANG DIRETUR',
            style: GoogleFonts.outfit(
                color: _textMuted,
                fontSize: 11,
                fontWeight: FontWeight.w700,
                letterSpacing: 1)),
        const SizedBox(height: 16),
        Row(children: [
          _qtyBtn(Icons.remove_rounded, state.returnedQty > 0,
              () => setState(() => _states[pId]!.returnedQty--)),
          Expanded(
              child: Column(children: [
            Text('${state.returnedQty}',
                textAlign: TextAlign.center,
                style: GoogleFonts.outfit(
                  color: state.returnedQty > 0
                      ? const Color(0xFFEF4444)
                      : _textDark,
                  fontSize: 36,
                  fontWeight: FontWeight.w900,
                )),
            Text('dari $maxQty $unit',
                style: GoogleFonts.outfit(color: _textMuted, fontSize: 11),
                textAlign: TextAlign.center),
          ])),
          _qtyBtn(Icons.add_rounded, state.returnedQty < maxQty,
              () => setState(() => _states[pId]!.returnedQty++)),
        ]),
        const SizedBox(height: 12),
        Slider(
          value: state.returnedQty.toDouble(),
          min: 0,
          max: maxQty.toDouble(),
          divisions: maxQty > 0 ? maxQty : 1,
          activeColor:
              state.returnedQty > 0 ? const Color(0xFFEF4444) : _primaryColor,
          inactiveColor: const Color(0xFFE2E8F0),
          onChanged: (v) =>
              setState(() => _states[pId]!.returnedQty = v.toInt()),
        ),
        const SizedBox(height: 4),
        Row(children: [
          _quickBtn('BATAL RETUR', 0, state.returnedQty,
              () => setState(() => _states[pId]!.returnedQty = 0)),
          const SizedBox(width: 8),
          _quickBtn(
              'SETENGAH',
              (maxQty / 2).ceil(),
              state.returnedQty,
              () => setState(
                  () => _states[pId]!.returnedQty = (maxQty / 2).ceil())),
          const SizedBox(width: 8),
          _quickBtn('SEMUA', maxQty, state.returnedQty,
              () => setState(() => _states[pId]!.returnedQty = maxQty)),
        ]),
      ]),
    );
  }

  Widget _qtyBtn(IconData icon, bool enabled, VoidCallback onTap) {
    return GestureDetector(
      onTap: enabled ? onTap : null,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        width: 48,
        height: 48,
        decoration: BoxDecoration(
          color: enabled ? const Color(0xFFF1F5F9) : const Color(0xFFF8FAFC),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: const Color(0xFFE2E8F0)),
        ),
        child: Icon(icon, color: enabled ? _textDark : _textLight, size: 20),
      ),
    );
  }

  Widget _quickBtn(String label, int target, int current, VoidCallback onTap) {
    final isActive = current == target;
    return Expanded(
      child: GestureDetector(
        onTap: onTap,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 150),
          padding: const EdgeInsets.symmetric(vertical: 8),
          decoration: BoxDecoration(
            color: isActive
                ? _primaryColor.withOpacity(0.08)
                : const Color(0xFFF1F5F9),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
                color: isActive ? _primaryColor : const Color(0xFFE2E8F0)),
          ),
          child: Text(label,
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(
                color: isActive ? _primaryColor : _textMuted,
                fontSize: 9,
                fontWeight: FontWeight.w700,
                letterSpacing: 0.5,
              )),
        ),
      ),
    );
  }

  // ── Return Type Selector ────────────────────────────────────────────────

  Widget _buildReturnTypeSelector(String pId, _ReturnItemState state) {
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text('PILIH MEKANISME RETUR',
          style: GoogleFonts.outfit(
              color: _textMuted,
              fontSize: 11,
              fontWeight: FontWeight.w700,
              letterSpacing: 1)),
      const SizedBox(height: 10),
      ..._returnTypeConfigs.entries
          .map((e) => _buildReturnTypeCard(pId, state, e.key, e.value)),
    ]);
  }

  Widget _buildReturnTypeCard(String pId, _ReturnItemState state,
      String typeKey, _ReturnTypeConfig cfg) {
    final isSelected = state.returnType == typeKey;
    final color = Color(cfg.colorHex);
    return GestureDetector(
      onTap: () => setState(() => _states[pId]!.returnType = typeKey),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        margin: const EdgeInsets.only(bottom: 10),
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: isSelected ? color.withOpacity(0.06) : _cardColor,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(
              color: isSelected ? color : _cardBorder,
              width: isSelected ? 2 : 1),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                      color: color.withOpacity(0.08),
                      blurRadius: 12,
                      spreadRadius: 1)
                ]
              : [],
        ),
        child: Row(children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: isSelected ? color : const Color(0xFFF1F5F9),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(cfg.icon,
                color: isSelected ? Colors.white : _textMuted, size: 22),
          ),
          const SizedBox(width: 12),
          Expanded(
              child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                Text(cfg.label.toUpperCase(),
                    style: GoogleFonts.outfit(
                        color: isSelected ? color : _textDark,
                        fontWeight: FontWeight.w800,
                        fontSize: 12,
                        letterSpacing: 0.5)),
                const SizedBox(height: 2),
                Text(cfg.description,
                    style:
                        GoogleFonts.outfit(color: _textMuted, fontSize: 10.5)),
              ])),
          const SizedBox(width: 10),
          AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            width: 20,
            height: 20,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              border: Border.all(
                  color: isSelected ? color : const Color(0xFFCBD5E1),
                  width: 2),
              color: isSelected ? color : Colors.transparent,
            ),
            child: isSelected
                ? const Icon(Icons.check, color: Colors.white, size: 12)
                : null,
          ),
        ]),
      ),
    );
  }

  // ── Return Reason ────────────────────────────────────────────────────────

  static const _returnReasons = {
    'RUSAK': '🔴 Barang Rusak / Penyok',
    'TOLAK_CUSTOMER': '🚫 Ditolak Pelanggan',
    'KADALUARSA': '⏰ Kadaluarsa / Mendekati ED',
    'CACAT_PABRIK': '🏭 Cacat Produksi / Garansi',
    'LAINNYA': '📝 Lainnya',
  };

  Widget _buildReturnReasonSelector(String pId, _ReturnItemState state) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: _cardColor,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _cardBorder),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 8,
              offset: const Offset(0, 4))
        ],
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text('ALASAN RETUR',
            style: GoogleFonts.outfit(
                color: _textMuted,
                fontSize: 11,
                fontWeight: FontWeight.w700,
                letterSpacing: 1)),
        const SizedBox(height: 12),
        Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _returnReasons.entries.map((e) {
              final isSelected = state.returnReason == e.key;
              return GestureDetector(
                onTap: () => setState(() => _states[pId]!.returnReason = e.key),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 150),
                  padding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  decoration: BoxDecoration(
                    color: isSelected
                        ? _primaryColor.withOpacity(0.08)
                        : const Color(0xFFF1F5F9),
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(
                        color: isSelected
                            ? _primaryColor
                            : const Color(0xFFE2E8F0),
                        width: isSelected ? 1.5 : 1),
                  ),
                  child: Text(e.value,
                      style: GoogleFonts.outfit(
                        color: isSelected ? _primaryColor : _textDark,
                        fontSize: 11,
                        fontWeight:
                            isSelected ? FontWeight.w700 : FontWeight.w500,
                      )),
                ),
              );
            }).toList()),
      ]),
    );
  }

  Widget _buildExtraFields(String pId, _ReturnItemState state) {
    switch (state.returnType) {
      case 'TUKAR_PRODUK':
        return _buildTukarProdukFields(pId, state);
      case 'GARANSI':
        return _buildGaransiFields(pId, state);

      case 'TUKAR_SUSULAN':
        return _buildInfoCard(
          icon: Icons.info_outline_rounded,
          color: const Color(0xFF8B5CF6),
          title: 'Pengiriman Susulan Akan Dijadwalkan',
          body:
              'Barang pengganti akan dikirim pada Surat Jalan berikutnya. Admin gudang akan menjadwalkan setelah retur disetujui.',
        );
      case 'TUKAR_LANGSUNG':
        return _buildInfoCard(
          icon: Icons.local_shipping_rounded,
          color: const Color(0xFFF59E0B),
          title: 'Tukar di Tempat dari Stok Truk',
          body:
              'Pastikan ada stok pengganti produk sejenis di truk Anda. Barang rusak dibawa kembali to gudang karantina.',
        );
      default:
        return const SizedBox.shrink();
    }
  }

  Widget _buildTukarProdukFields(String pId, _ReturnItemState state) {
    final products = widget.availableProducts.where((p) {
      final stock =
          (p['warehouse_stock'] ?? p['quantity'] ?? p['stock'] ?? 0) as int;
      return stock > 0;
    }).toList();
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: _cardColor,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFF06B6D4).withOpacity(0.5)),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          const Icon(Icons.compare_arrows_rounded,
              color: Color(0xFF06B6D4), size: 18),
          const SizedBox(width: 8),
          Text('PILIH PRODUK PENGGANTI',
              style: GoogleFonts.outfit(
                  color: const Color(0xFF06B6D4),
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                  letterSpacing: 0.5)),
        ]),
        const SizedBox(height: 12),
        if (products.isEmpty)
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.orange.withOpacity(0.08),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: Colors.orange.withOpacity(0.3)),
            ),
            child: Row(children: [
              const Icon(Icons.warning_amber_rounded,
                  color: Colors.orange, size: 16),
              const SizedBox(width: 8),
              Expanded(
                  child: Text(
                      'Data produk tidak tersedia. Hubungi admin untuk pengalihan manual.',
                      style: GoogleFonts.outfit(
                          color: Colors.orange, fontSize: 11))),
            ]),
          )
        else ...[
          DropdownButtonFormField<String>(
            initialValue: state.replacementProductId,
            dropdownColor: _cardColor,
            decoration: InputDecoration(
              labelText: 'Produk Pengganti',
              labelStyle: GoogleFonts.outfit(color: _textMuted, fontSize: 12),
              filled: true,
              fillColor: const Color(0xFFF8FAFC),
              enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(10),
                  borderSide: const BorderSide(color: const Color(0xFFE2E8F0))),
              focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(10),
                  borderSide: const BorderSide(color: Color(0xFF06B6D4))),
            ),
            style: GoogleFonts.outfit(color: _textDark, fontSize: 13),
            hint: Text('Pilih produk…',
                style: GoogleFonts.outfit(color: _textLight, fontSize: 13)),
            items: products.map<DropdownMenuItem<String>>((p) {
              final pId2 = p['id']?.toString() ?? '';
              final pName = p['dropdown_label']?.toString() ??
                  p['name']?.toString() ??
                  'Produk';
              return DropdownMenuItem<String>(
                value: pId2,
                child: Text(pName,
                    style: GoogleFonts.outfit(color: _textDark, fontSize: 13)),
              );
            }).toList(),
            onChanged: (v) => setState(() {
              _states[pId]!.replacementProductId = v;
              final found = products.firstWhere((p) => p['id']?.toString() == v,
                  orElse: () => {});
              _states[pId]!.replacementProductName = found['name']?.toString();
              _states[pId]!.replacementPrice = (found['price'] ?? 0.0) is num
                  ? (found['price'] ?? 0.0).toDouble()
                  : 0.0;
            }),
          ),
          if (state.replacementProductId != null) ...[
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              decoration: BoxDecoration(
                color: const Color(0xFF06B6D4).withOpacity(0.06),
                borderRadius: BorderRadius.circular(10),
                border:
                    Border.all(color: const Color(0xFF06B6D4).withOpacity(0.2)),
              ),
              child: Text(
                  'Selisih harga akan dihitung otomatis saat retur disetujui admin.',
                  style: GoogleFonts.outfit(
                      color: const Color(0xFF06B6D4), fontSize: 11)),
            ),
          ],
        ],
      ]),
    );
  }

  Widget _buildGaransiFields(String pId, _ReturnItemState state) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: _cardColor,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFF10B981).withOpacity(0.5)),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          const Icon(Icons.verified_user_rounded,
              color: Color(0xFF10B981), size: 18),
          const SizedBox(width: 8),
          Text('INFORMASI KLAIM GARANSI',
              style: GoogleFonts.outfit(
                  color: const Color(0xFF10B981),
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                  letterSpacing: 0.5)),
        ]),
        const SizedBox(height: 12),
        TextField(
          controller: state.claimRefController,
          style: GoogleFonts.outfit(color: _textDark, fontSize: 13),
          decoration: InputDecoration(
            labelText: 'No. Referensi Klaim (opsional)',
            labelStyle: GoogleFonts.outfit(color: _textMuted, fontSize: 12),
            hintText: 'Contoh: WRN-2024-001',
            hintStyle: GoogleFonts.outfit(color: _textLight, fontSize: 12),
            filled: true,
            fillColor: const Color(0xFFF8FAFC),
            prefixIcon:
                const Icon(Icons.tag_rounded, color: _textMuted, size: 18),
            enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(10),
                borderSide: const BorderSide(color: const Color(0xFFE2E8F0))),
            focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(10),
                borderSide: const BorderSide(color: Color(0xFF10B981))),
          ),
          onChanged: (v) => _states[pId]!.claimRefNo = v,
        ),
        const SizedBox(height: 12),
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: const Color(0xFF10B981).withOpacity(0.06),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: const Color(0xFF10B981).withOpacity(0.2)),
          ),
          child: Column(children: [
            Row(children: [
              const Icon(Icons.shield_rounded,
                  color: Color(0xFF10B981), size: 16),
              const SizedBox(width: 8),
              Expanded(
                  child: Text(
                      'Nota TIDAK dipotong. Produk pengganti dikirim terpisah oleh tim support.',
                      style: GoogleFonts.outfit(
                          color: const Color(0xFF10B981), fontSize: 11))),
            ]),
            const SizedBox(height: 8),
            Row(children: [
              const Icon(Icons.photo_camera_rounded,
                  color: _textMuted, size: 16),
              const SizedBox(width: 8),
              Expanded(
                  child: Text(
                      'Foto bukti kerusakan dapat dilampirkan saat konfirmasi pengiriman (POD photo).',
                      style: GoogleFonts.outfit(
                          color: _textMuted, fontSize: 10.5))),
            ]),
          ]),
        ),
      ]),
    );
  }

  Widget _buildInfoCard(
      {required IconData icon,
      required Color color,
      required String title,
      required String body}) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withOpacity(0.06),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: color.withOpacity(0.2)),
      ),
      child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Icon(icon, color: color, size: 20),
        const SizedBox(width: 10),
        Expanded(
            child:
                Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(title,
              style: GoogleFonts.outfit(
                  color: color, fontWeight: FontWeight.w700, fontSize: 12)),
          const SizedBox(height: 4),
          Text(body,
              style: GoogleFonts.outfit(color: _textMuted, fontSize: 11)),
        ])),
      ]),
    );
  }

  // ── Impact Preview ────────────────────────────────────────────────────────

  Widget _buildImpactPreview(_ReturnItemState state, int orderedQty, int ppu,
      double price, String unit, NumberFormat currFmt) {
    final cfg = _returnTypeConfigs[state.returnType]!;
    final color = Color(cfg.colorHex);
    final retValue = state.returnedQty * ppu * price;
    final remainingQty = orderedQty - state.returnedQty;

    String impactText;
    Color impactColor;
    IconData impactIcon;

    switch (state.returnType) {
      case 'POTONG_NOTA':
        impactText = 'Nota dikurangi: ${currFmt.format(retValue)}';
        impactColor = const Color(0xFFEF4444);
        impactIcon = Icons.remove_circle_outline_rounded;
        break;
      case 'GARANSI':
        impactText = 'Nota TIDAK dipotong. Klaim garansi diajukan ke supplier.';
        impactColor = const Color(0xFF10B981);
        impactIcon = Icons.verified_user_outlined;
        break;
      default:
        impactText =
            'Nota TETAP penuh. ${state.returnType == 'TUKAR_PRODUK' ? 'Selisih dihitung saat approval.' : ''}';
        impactColor = const Color(0xFFF59E0B);
        impactIcon = Icons.swap_vert_circle_outlined;
    }

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withOpacity(0.06),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: color.withOpacity(0.2)),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text('RINGKASAN DAMPAK',
            style: GoogleFonts.outfit(
                color: _textMuted,
                fontSize: 10,
                fontWeight: FontWeight.w700,
                letterSpacing: 1)),
        const SizedBox(height: 10),
        _impactRow('Qty Diretur', '${state.returnedQty} $unit',
            const Color(0xFFEF4444)),
        const SizedBox(height: 6),
        _impactRow(
            'Qty Terkirim', '$remainingQty $unit', const Color(0xFF10B981)),
        const SizedBox(height: 6),
        _impactRow('Tipe Retur', cfg.label, color),
        const SizedBox(height: 10),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          decoration: BoxDecoration(
              color: impactColor.withOpacity(0.08),
              borderRadius: BorderRadius.circular(8)),
          child: Row(children: [
            Icon(impactIcon, color: impactColor, size: 16),
            const SizedBox(width: 8),
            Expanded(
                child: Text(impactText,
                    style: GoogleFonts.outfit(
                        color: impactColor,
                        fontWeight: FontWeight.w600,
                        fontSize: 11.5))),
          ]),
        ),
        const SizedBox(height: 8),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
          decoration: BoxDecoration(
              color: const Color(0xFFF1F5F9),
              borderRadius: BorderRadius.circular(8)),
          child: Text('💡 ${cfg.hint}',
              style: GoogleFonts.outfit(color: _textMuted, fontSize: 10)),
        ),
      ]),
    );
  }

  Widget _impactRow(String label, String value, Color valueColor) {
    return Row(children: [
      Text('$label:',
          style: GoogleFonts.outfit(color: _textMuted, fontSize: 11)),
      const Spacer(),
      Text(value,
          style: GoogleFonts.outfit(
              color: valueColor, fontSize: 12, fontWeight: FontWeight.w700)),
    ]);
  }

  Widget _buildNoReturnCard(String name) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: const Color(0xFF10B981).withOpacity(0.06),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFF10B981).withOpacity(0.2)),
      ),
      child: Column(children: [
        const Icon(Icons.check_circle_rounded,
            color: Color(0xFF10B981), size: 40),
        const SizedBox(height: 12),
        Text('Semua Terkirim ✓',
            style: GoogleFonts.outfit(
                color: const Color(0xFF10B981),
                fontWeight: FontWeight.w800,
                fontSize: 16)),
        const SizedBox(height: 6),
        Text('$name diterima penuh tanpa retur.',
            textAlign: TextAlign.center,
            style: GoogleFonts.outfit(color: _textMuted, fontSize: 12)),
      ]),
    );
  }

  // ── Summary Bar ───────────────────────────────────────────────────────────

  Widget _buildSummaryBar(NumberFormat currFmt) {
    int totalReturnedItems = 0;
    int totalDeductions = 0;
    for (final s in _states.values) {
      if (s.returnedQty > 0) {
        totalReturnedItems++;
        if (s.returnType == 'POTONG_NOTA') totalDeductions++;
      }
    }
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
      decoration: const BoxDecoration(
        color: _cardColor,
        border: Border(top: BorderSide(color: _cardBorder)),
      ),
      child: Row(children: [
        Expanded(
            child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
              if (totalReturnedItems == 0)
                Text('Tidak ada item retur',
                    style: GoogleFonts.outfit(color: _textMuted, fontSize: 12))
              else ...[
                Text('$totalReturnedItems item diretur',
                    style: GoogleFonts.outfit(
                        color: _textDark,
                        fontWeight: FontWeight.w700,
                        fontSize: 13)),
                Text(
                    '$totalDeductions potong nota · ${totalReturnedItems - totalDeductions} tanpa potongan',
                    style:
                        GoogleFonts.outfit(color: _textMuted, fontSize: 10.5)),
              ],
            ])),
        const SizedBox(width: 12),
        ElevatedButton.icon(
          style: ElevatedButton.styleFrom(
            backgroundColor: _primaryColor, foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
            minimumSize: const Size(
                0, 48), // Overrides global double.infinity to prevent Row crash
            shape:
                RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          ),
          icon: const Icon(Icons.check_rounded, size: 18),
          label: Text('Konfirmasi',
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w700, fontSize: 13)),
          onPressed: _onConfirm,
        ),
      ]),
    );
  }

  // ── Confirm ───────────────────────────────────────────────────────────────

  void _onConfirm() {
    for (final entry in _states.entries) {
      final s = entry.value;
      if (s.returnedQty > 0 &&
          s.returnType == 'TUKAR_PRODUK' &&
          s.replacementProductId == null) {
        final item = widget.items.firstWhere(
          (i) =>
              (i['product_id']?.toString() ?? i['id']?.toString()) == entry.key,
          orElse: () => <String, dynamic>{},
        );
        final name = item['product']?['name'] ?? 'Produk';
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(
          content: Text('Pilih produk pengganti untuk "$name"'),
          backgroundColor: const Color(0xFFEF4444),
        ));
        return;
      }
    }

    final results = <ReturnItemResult>[];
    for (final item in widget.items) {
      final pId =
          item['product_id']?.toString() ?? item['id']?.toString() ?? '';
      final s = _states[pId]!;
      if (s.returnedQty <= 0) continue;
      results.add(ReturnItemResult(
        productId: pId,
        productName: item['product']?['name'] ?? 'Produk',
        returnedQuantity: s.returnedQty,
        returnReason: s.returnReason,
        returnType: s.returnType,
        orderedQuantity: item['ordered_quantity'] ?? 0,
        price: (item['price'] as num? ?? 0.0).toDouble(),
        unit: item['unit'] ?? 'PCS',
        piecesPerUnit: item['pieces_per_unit'] ?? 1,
        replacementProductId: s.replacementProductId,
        replacementProductName: s.replacementProductName,
        replacementQuantity:
            s.replacementProductId != null ? s.returnedQty : null,
        replacementPrice: s.replacementPrice,
        claimReferenceNo: s.claimRefNo.isNotEmpty ? s.claimRefNo : null,
      ));
    }
    Navigator.pop(context, results);
  }
}

// ─────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────

class _ReturnItemState {
  int orderedQty;
  int returnedQty;
  String returnType;
  String returnReason;
  String? replacementProductId;
  String? replacementProductName;
  double? replacementPrice;
  String claimRefNo = '';
  final TextEditingController claimRefController;

  _ReturnItemState({
    required this.orderedQty,
    required this.returnedQty,
    required this.returnType,
    required this.returnReason,
    required this.claimRefController,
  });
}

class _ReturnTypeConfig {
  final String label;
  final String description;
  final IconData icon;
  final int colorHex;
  final String hint;

  const _ReturnTypeConfig({
    required this.label,
    required this.description,
    required this.icon,
    required this.colorHex,
    required this.hint,
  });
}
