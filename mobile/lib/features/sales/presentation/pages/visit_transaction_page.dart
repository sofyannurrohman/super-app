import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:image_picker/image_picker.dart';
import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_checkout_page.dart';
import 'package:hris_app/features/sales/presentation/pages/order_entry_page.dart';
import 'package:hris_app/features/sales/presentation/pages/receipt_camera_page.dart';
import 'package:hris_app/features/sales/presentation/pages/merchandising_page.dart';
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:hris_app/injection.dart' as di;

class VisitTransactionPage extends StatefulWidget {
  final StoreModel store;
  final String selfiePath;
  final Uint8List? selfieBytes;

  const VisitTransactionPage(
      {super.key,
      required this.store,
      required this.selfiePath,
      this.selfieBytes});

  @override
  State<VisitTransactionPage> createState() => _VisitTransactionPageState();
}

class _VisitTransactionPageState extends State<VisitTransactionPage> {
  bool _isLoading = false;
  String _jobPositionTitle = '';
  bool _isDistributor = false;
  bool _merchandisingDone = false;
  bool _stockOpnameDone = false;
  bool _collectionDone = false;

  bool get _isMT =>
      _jobPositionTitle.toUpperCase().contains('MT') ||
      _jobPositionTitle.toUpperCase().contains('MODERN') ||
      _jobPositionTitle.toUpperCase().contains('MERCHANDIS');

  @override
  void initState() {
    super.initState();
    _fetchProfile();
  }

  Future<void> _fetchProfile() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final cached = prefs.getString('cached_employee_profile');
      if (cached != null) {
        final decoded = jsonDecode(cached) as Map<String, dynamic>;
        final title = (decoded['job_position']?['title'] ?? decoded['job_title'] ?? '').toString();
        final isDist = decoded['is_distributor'] == true ||
            (decoded['distributor_store_id'] != null && decoded['distributor_store_id'].toString().isNotEmpty) ||
            title.toLowerCase().contains('distributor');
        if (mounted) {
          setState(() {
            if (title.isNotEmpty) _jobPositionTitle = title;
            _isDistributor = isDist;
          });
        }
      }
      final authRepo = di.sl<AuthRepository>();
      final result = await authRepo.getProfile();
      result.fold(
        (_) => null,
        (profile) {
          final isDist = profile['is_distributor'] == true ||
              (profile['distributor_store_id'] != null && profile['distributor_store_id'].toString().isNotEmpty) ||
              (profile['job_position']?['title'] ?? '').toString().toLowerCase().contains('distributor');
          if (mounted) {
            setState(() {
              _jobPositionTitle = profile['job_position']?['title'] ?? _jobPositionTitle;
              _isDistributor = isDist;
            });
          }
        },
      );
    } catch (_) {}
  }

  void _navigateToOrderEntry(String jobPosition,
      {String companyId = '', String companyName = '', bool isDistributor = false}) {
    Navigator.push(
      context,
      MaterialPageRoute(
          builder: (_) => OrderEntryPage(
                store: widget.store,
                selfiePath: widget.selfiePath,
                selfieBytes: widget.selfieBytes,
                companyId: companyId,
                companyName: companyName,
                jobPositionTitle: jobPosition,
                isDistributor: isDistributor || _isDistributor,
              )),
    );
  }

  Future<void> _handleAutoSelection() async {
    // Gunakan profil lokal di SharedPreferences terlebih dahulu agar navigasi ke OrderEntryPage instan (0ms) saat offline
    try {
      final prefs = await SharedPreferences.getInstance();
      final cached = prefs.getString('cached_employee_profile');
      if (cached != null) {
        final profile = jsonDecode(cached) as Map<String, dynamic>;
        final companyId = profile['company_id'] ??
            profile['companyId'] ??
            profile['CompanyID'] ??
            '';
        final companyName = profile['company_name'] ??
            profile['companyName'] ??
            profile['CompanyNAME'] ??
            profile['company']?['name'] ??
            '';
        final jobPosition = (profile['job_position']?['title'] ??
                profile['job_title'] ??
                _jobPositionTitle)
            .toString();
        final isDist = profile['is_distributor'] == true ||
            (profile['distributor_store_id'] != null && profile['distributor_store_id'].toString().isNotEmpty) ||
            jobPosition.toLowerCase().contains('distributor');
        if (mounted) {
          _navigateToOrderEntry(jobPosition,
              companyId: companyId.toString(),
              companyName: companyName.toString(),
              isDistributor: isDist);
        }
        return;
      }
    } catch (_) {}

    setState(() => _isLoading = true);
    try {
      final authRepo = di.sl<AuthRepository>();
      final result = await authRepo.getProfile();

      if (mounted) {
        result.fold(
          (failure) => _navigateToOrderEntry(_jobPositionTitle, isDistributor: _isDistributor),
          (profile) {
            final companyId = profile['company_id'] ??
                profile['companyId'] ??
                profile['CompanyID'] ??
                '';
            final companyName = profile['company_name'] ??
                profile['companyName'] ??
                profile['CompanyNAME'] ??
                profile['company']?['name'] ??
                '';
            final jobPosition = profile['job_position']?['title'] ?? '';
            final isDist = profile['is_distributor'] == true ||
                (profile['distributor_store_id'] != null && profile['distributor_store_id'].toString().isNotEmpty) ||
                jobPosition.toLowerCase().contains('distributor');

            _navigateToOrderEntry(jobPosition,
                companyId: companyId.toString(),
                companyName: companyName.toString(),
                isDistributor: isDist);
          },
        );
      }
    } catch (e) {
      if (mounted) _navigateToOrderEntry(_jobPositionTitle, isDistributor: _isDistributor);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _showNoTransactionDialog() {
    final TextEditingController reasonCtrl = TextEditingController();
    final List<String> reasons = [
      'Toko Tutup',
      'Stok Masih Banyak',
      'Owner Tidak Ada',
      'Toko Sedang Renovasi',
      'Lainnya (Tulis Catatan)'
    ];

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        padding:
            EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.only(
              topLeft: Radius.circular(32), topRight: Radius.circular(32)),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(height: 12),
            Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(2))),
            const SizedBox(height: 24),
            Text('ALASAN TIDAK ADA TRANSAKSI',
                style: GoogleFonts.outfit(
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                    color: Colors.blueGrey,
                    letterSpacing: 1)),
            const SizedBox(height: 16),
            Flexible(
              child: ListView.builder(
                shrinkWrap: true,
                padding: const EdgeInsets.symmetric(horizontal: 24),
                itemCount: reasons.length,
                itemBuilder: (context, index) => ListTile(
                  contentPadding: EdgeInsets.zero,
                  leading: const Icon(Icons.info_outline_rounded,
                      color: Colors.orange),
                  title: Text(reasons[index],
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w700,
                          color: const Color(0xFF1E293B))),
                  onTap: () {
                    if (reasons[index].startsWith('Lainnya')) {
                      // Stay and let them type
                    } else {
                      Navigator.pop(context);
                      _proceedWithNoTransaction(reasons[index]);
                    }
                  },
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                children: [
                  TextField(
                    controller: reasonCtrl,
                    decoration: InputDecoration(
                      hintText: 'Tulis alasan lainnya di sini...',
                      hintStyle: GoogleFonts.outfit(fontSize: 14),
                      filled: true,
                      fillColor: const Color(0xFFF1F5F9),
                      border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                          borderSide: BorderSide.none),
                    ),
                    maxLines: 2,
                  ),
                  const SizedBox(height: 16),
                  SizedBox(
                    width: double.infinity,
                    height: 56,
                    child: ElevatedButton(
                      onPressed: () {
                        if (reasonCtrl.text.isNotEmpty) {
                          Navigator.pop(context);
                          _proceedWithNoTransaction(reasonCtrl.text);
                        }
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF1E293B),
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16)),
                      ),
                      child: Text('SIMPAN ALASAN',
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w900,
                              color: Colors.white)),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showStockOpnameDialog() async {
    final salesService = SalesApiService(apiClient: di.sl<ApiClient>());
    setState(() => _isLoading = true);
    List<Map<String, dynamic>> existingStocks = [];
    List<Map<String, dynamic>> availableProducts = [];
    try {
      final results = await Future.wait([
        salesService.getMTStoreStocks(storeId: widget.store.id),
        salesService.getActiveProducts(),
      ]);
      existingStocks = results[0];
      availableProducts = results[1];
    } catch (_) {}
    setState(() => _isLoading = false);

    if (!mounted) return;

    // Build opname item rows
    final List<Map<String, dynamic>> opnameRows = [];
    for (final s in existingStocks) {
      final pId = s['product_id'] ?? '';
      final pName = s['product_name'] ?? (s['product']?['name'] ?? 'Produk');
      final sku = s['sku'] ?? (s['product']?['sku'] ?? '');
      final currentQty = s['quantity'] ?? 0;
      opnameRows.add({
        'product_id': pId,
        'product_name': pName,
        'sku': sku,
        'current_quantity': currentQty,
        'controller': TextEditingController(text: currentQty.toString()),
      });
    }

    final notesCtrl = TextEditingController(text: 'Opname saat kunjungan sales');

    void showAddProductPicker(StateSetter setModalState) {
      showModalBottomSheet(
        context: context,
        isScrollControlled: true,
        backgroundColor: Colors.transparent,
        builder: (pCtx) {
          String filter = '';
          return StatefulBuilder(
            builder: (pCtx, setPickerState) {
              final filtered = availableProducts.where((p) {
                final name = (p['name'] ?? '').toString().toLowerCase();
                final sku = (p['sku'] ?? '').toString().toLowerCase();
                final q = filter.toLowerCase();
                return name.contains(q) || sku.contains(q);
              }).toList();

              return Container(
                height: MediaQuery.of(context).size.height * 0.7,
                decoration: const BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
                ),
                padding: const EdgeInsets.all(20),
                child: Column(
                  children: [
                    Container(
                      width: 40,
                      height: 4,
                      decoration: BoxDecoration(
                        color: Colors.grey.shade300,
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'PILIH PRODUK OPNAME',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w900,
                        fontSize: 14,
                        color: const Color(0xFF1E293B),
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      decoration: InputDecoration(
                        hintText: 'Cari nama produk / SKU...',
                        prefixIcon: const Icon(Icons.search_rounded, size: 20),
                        filled: true,
                        fillColor: const Color(0xFFF1F5F9),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(14),
                          borderSide: BorderSide.none,
                        ),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 16),
                      ),
                      onChanged: (val) => setPickerState(() => filter = val),
                    ),
                    const SizedBox(height: 12),
                    Expanded(
                      child: filtered.isEmpty
                          ? Center(
                              child: Text('Tidak ada produk ditemukan',
                                  style: GoogleFonts.outfit(color: Colors.blueGrey)),
                            )
                          : ListView.separated(
                              itemCount: filtered.length,
                              separatorBuilder: (_, __) => const Divider(height: 1),
                              itemBuilder: (_, i) {
                                final p = filtered[i];
                                final isAdded = opnameRows.any((r) => r['product_id'] == p['id']);
                                return ListTile(
                                  contentPadding: const EdgeInsets.symmetric(vertical: 4, horizontal: 8),
                                  title: Text(
                                    p['name'] ?? '',
                                    style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w700,
                                      fontSize: 13,
                                    ),
                                  ),
                                  subtitle: Text(
                                    'SKU: ${p['sku'] ?? '-'} • Satuan: ${p['unit'] ?? 'PCS'}',
                                    style: GoogleFonts.outfit(fontSize: 11, color: Colors.blueGrey),
                                  ),
                                  trailing: isAdded
                                      ? const Icon(Icons.check_circle_rounded, color: Colors.green, size: 22)
                                      : IconButton(
                                          icon: const Icon(Icons.add_circle_outline_rounded, color: Color(0xFF0284C7)),
                                          onPressed: () {
                                            setModalState(() {
                                              opnameRows.add({
                                                'product_id': p['id'],
                                                'product_name': p['name'] ?? '',
                                                'sku': p['sku'] ?? '',
                                                'current_quantity': 0,
                                                'controller': TextEditingController(text: '0'),
                                              });
                                            });
                                            setPickerState(() {});
                                          },
                                        ),
                                );
                              },
                            ),
                    ),
                  ],
                ),
              );
            },
          );
        },
      );
    }

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setModalState) => Container(
          constraints: BoxConstraints(
            maxHeight: MediaQuery.of(ctx).size.height * 0.85,
          ),
          padding: EdgeInsets.only(
            left: 20,
            right: 20,
            top: 20,
            bottom: MediaQuery.of(ctx).viewInsets.bottom + 20,
          ),
          decoration: const BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: const Color(0xFF0284C7).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: const Icon(Icons.inventory_2_rounded, color: Color(0xFF0284C7), size: 20),
                      ),
                      const SizedBox(width: 10),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'STOCK OPNAME GERAI MT',
                            style: GoogleFonts.outfit(
                              fontSize: 14,
                              fontWeight: FontWeight.w900,
                              color: const Color(0xFF1E293B),
                            ),
                          ),
                          Text(
                            widget.store.name,
                            style: GoogleFonts.outfit(fontSize: 11, color: Colors.blueGrey),
                          ),
                        ],
                      ),
                    ],
                  ),
                  TextButton.icon(
                    onPressed: () => showAddProductPicker(setModalState),
                    icon: const Icon(Icons.add_rounded, size: 16, color: Color(0xFF0284C7)),
                    label: Text(
                      'Tambah',
                      style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF0284C7)),
                    ),
                    style: TextButton.styleFrom(
                      backgroundColor: const Color(0xFFE0F2FE),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Text(
                'Periksa dan sesuaikan sisa stok fisik riil untuk setiap produk di gerai ini (dalam satuan Pcs).',
                style: GoogleFonts.outfit(fontSize: 11, color: Colors.blueGrey),
              ),
              const SizedBox(height: 12),
              Expanded(
                child: opnameRows.isEmpty
                    ? Container(
                        width: double.infinity,
                        padding: const EdgeInsets.all(24),
                        decoration: BoxDecoration(
                          color: const Color(0xFFF8FAFC),
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(color: const Color(0xFFE2E8F0)),
                        ),
                        child: Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              const Icon(Icons.shelves, size: 36, color: Colors.blueGrey),
                              const SizedBox(height: 8),
                              Text(
                                'Belum ada produk untuk di-opname.',
                                style: GoogleFonts.outfit(fontSize: 12, color: Colors.blueGrey),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                'Klik "+ Tambah" di atas untuk memilih produk.',
                                style: GoogleFonts.outfit(fontSize: 11, color: const Color(0xFF0284C7), fontWeight: FontWeight.w600),
                              ),
                            ],
                          ),
                        ),
                      )
                    : ListView.separated(
                        shrinkWrap: true,
                        itemCount: opnameRows.length,
                        separatorBuilder: (_, __) => const SizedBox(height: 10),
                        itemBuilder: (ctx, idx) {
                          final row = opnameRows[idx];
                          final ctrl = row['controller'] as TextEditingController;
                          return Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: const Color(0xFFF0F9FF),
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(color: const Color(0xFFBAE6FD)),
                            ),
                            child: Row(
                              children: [
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        row['product_name'] ?? '',
                                        style: GoogleFonts.outfit(fontWeight: FontWeight.w700, fontSize: 13, color: const Color(0xFF1E293B)),
                                      ),
                                      const SizedBox(height: 2),
                                      Text(
                                        'SKU: ${row['sku'] ?? '-'} • Sebelumnya: ${row['current_quantity'] ?? 0} Pcs',
                                        style: GoogleFonts.outfit(fontSize: 10, color: Colors.blueGrey),
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 8),
                                SizedBox(
                                  width: 84,
                                  height: 42,
                                  child: TextFormField(
                                    controller: ctrl,
                                    keyboardType: TextInputType.number,
                                    textAlign: TextAlign.center,
                                    style: GoogleFonts.outfit(fontWeight: FontWeight.bold, fontSize: 13),
                                    decoration: InputDecoration(
                                      contentPadding: const EdgeInsets.symmetric(horizontal: 6, vertical: 8),
                                      suffixText: 'Pcs',
                                      suffixStyle: GoogleFonts.outfit(fontSize: 10, color: const Color(0xFF0284C7), fontWeight: FontWeight.bold),
                                      filled: true,
                                      fillColor: Colors.white,
                                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(10), borderSide: const BorderSide(color: Color(0xFFBAE6FD))),
                                      enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(10), borderSide: const BorderSide(color: Color(0xFFBAE6FD))),
                                    ),
                                  ),
                                ),
                                IconButton(
                                  icon: const Icon(Icons.delete_outline_rounded, color: Colors.redAccent, size: 20),
                                  onPressed: () {
                                    setModalState(() {
                                      opnameRows[idx]['controller']?.dispose();
                                      opnameRows.removeAt(idx);
                                    });
                                  },
                                ),
                              ],
                            ),
                          );
                        },
                      ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: notesCtrl,
                decoration: InputDecoration(
                  labelText: 'Catatan Stock Opname',
                  labelStyle: GoogleFonts.outfit(fontSize: 11),
                  hintText: 'Kondisi barang / keterangan rak...',
                  filled: true,
                  fillColor: const Color(0xFFF8FAFC),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(14), borderSide: BorderSide(color: Colors.grey.shade300)),
                  contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                ),
              ),
              const SizedBox(height: 16),
              SizedBox(
                width: double.infinity,
                height: 50,
                child: ElevatedButton(
                  onPressed: opnameRows.isEmpty
                      ? null
                      : () async {
                          final itemsToAdjust = <Map<String, dynamic>>[];
                          for (final row in opnameRows) {
                            final ctrl = row['controller'] as TextEditingController?;
                            final qty = int.tryParse(ctrl?.text.trim() ?? '') ?? 0;
                            itemsToAdjust.add({
                              'product_id': row['product_id'],
                              'quantity': qty,
                              'min_stock_threshold': 20,
                            });
                          }

                          Navigator.pop(ctx);
                          setState(() => _isLoading = true);

                          final ok = await salesService.adjustMTStoreStocksBulk(
                            storeId: widget.store.id,
                            items: itemsToAdjust,
                            adjustmentType: 'ADJUSTMENT',
                            notes: notesCtrl.text.trim(),
                            referenceNo: 'OPNAME-VISIT-${DateTime.now().millisecondsSinceEpoch}',
                          );
                          try {
                            di.sl<SyncBloc>().add(const SyncDataRequested(force: true));
                          } catch (_) {}
                          setState(() => _isLoading = false);

                          if (mounted) {
                            if (ok) {
                              setState(() => _stockOpnameDone = true);
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('Stock Opname gerai berhasil disimpan!'),
                                  backgroundColor: Colors.green,
                                ),
                              );
                            } else {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('Gagal menyimpan Stock Opname ke server. Periksa jaringan Anda dan coba lagi.'),
                                  backgroundColor: Colors.redAccent,
                                ),
                              );
                            }
                          }
                        },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF0284C7),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                  child: Text(
                    'SIMPAN STOK OPNAME',
                    style: GoogleFonts.outfit(fontWeight: FontWeight.w900, color: Colors.white),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showCollectionDialog() async {
    final salesService = SalesApiService(apiClient: di.sl<ApiClient>());
    setState(() => _isLoading = true);
    List<Map<String, dynamic>> allPOs = [];
    try {
      allPOs = await salesService.getMTPurchaseOrders();
    } catch (_) {}
    setState(() => _isLoading = false);

    // Filter POs for this store/supermarket or pending settlement
    final storePOs = allPOs.where((po) {
      final smName = (po['supermarket_name'] ?? '').toString().toLowerCase();
      final kaName = (po['key_account']?['name'] ?? po['key_account'] ?? '').toString().toLowerCase();
      final targetName = widget.store.name.toLowerCase();
      final isStoreMatch = smName.contains(targetName) || targetName.contains(smName) || kaName.contains(targetName) || targetName.contains(kaName);
      final isPending = po['billing_status'] != 'PAID_FULL';
      return (isStoreMatch || allPOs.length <= 5) && isPending;
    }).toList();

    XFile? proofPhoto;
    String? selectedPoId = storePOs.isNotEmpty ? storePOs.first['id'] : (allPOs.isNotEmpty ? allPOs.first['id'] : null);
    final notesCtrl = TextEditingController(text: 'Pelunasan faktur TOP saat kunjungan');

    if (!mounted) return;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setModalState) => Container(
          padding: EdgeInsets.only(
            left: 24,
            right: 24,
            top: 24,
            bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
          ),
          decoration: const BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: const Color(0xFF7C3AED).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: const Icon(Icons.receipt_long_rounded, color: Color(0xFF7C3AED), size: 20),
                  ),
                  const SizedBox(width: 10),
                  Text(
                    'PENAGIHAN ORDER / FAKTUR TOP',
                    style: GoogleFonts.outfit(
                      fontSize: 14,
                      fontWeight: FontWeight.w900,
                      color: const Color(0xFF1E293B),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Text(
                'Upload bukti bayar/transfer pelunasan faktur konsinyasi TOP 1 Bulan',
                style: GoogleFonts.outfit(fontSize: 12, color: Colors.blueGrey),
              ),
              const SizedBox(height: 16),
              if (storePOs.isNotEmpty) ...[
                Text(
                  'Pilih Tagihan / PO MT',
                  style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF1E293B)),
                ),
                const SizedBox(height: 6),
                DropdownButtonFormField<String>(
                  value: selectedPoId,
                  decoration: InputDecoration(
                    filled: true,
                    fillColor: const Color(0xFFF8FAFC),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(14), borderSide: BorderSide(color: Colors.grey.shade300)),
                  ),
                  items: storePOs.map((po) {
                    final poNum = po['po_number'] ?? 'PO';
                    final amt = po['total_amount'] ?? 0;
                    return DropdownMenuItem<String>(
                      value: po['id'],
                      child: Text('$poNum - Rp $amt', style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700)),
                    );
                  }).toList(),
                  onChanged: (val) => setModalState(() => selectedPoId = val),
                ),
                const SizedBox(height: 14),
              ],
              Text(
                'Foto Bukti Transfer / Kwitansi Pelunasan',
                style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF1E293B)),
              ),
              const SizedBox(height: 8),
              InkWell(
                onTap: () async {
                  final picker = ImagePicker();
                  final img = await picker.pickImage(source: ImageSource.camera, maxWidth: 1024, imageQuality: 70);
                  if (img != null) {
                    setModalState(() => proofPhoto = img);
                  }
                },
                child: Container(
                  width: double.infinity,
                  height: 100,
                  decoration: BoxDecoration(
                    color: const Color(0xFFF8FAFC),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: Colors.grey.shade300, style: BorderStyle.solid),
                  ),
                  child: proofPhoto != null
                      ? Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(Icons.check_circle_rounded, color: Colors.green, size: 28),
                            const SizedBox(width: 8),
                            Text('Foto bukti siap diunggah', style: GoogleFonts.outfit(fontWeight: FontWeight.bold, color: Colors.green)),
                          ],
                        )
                      : Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(Icons.camera_alt_rounded, color: Color(0xFF7C3AED), size: 28),
                            const SizedBox(height: 6),
                            Text('Ambil Foto Bukti Pembayaran', style: GoogleFonts.outfit(fontWeight: FontWeight.bold, fontSize: 12, color: const Color(0xFF7C3AED))),
                          ],
                        ),
                ),
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                height: 52,
                child: ElevatedButton(
                  onPressed: () async {
                    if (selectedPoId == null) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('Pilih nomor PO yang ditagihkan'), backgroundColor: Colors.red),
                      );
                      return;
                    }
                    Navigator.pop(ctx);
                    setState(() => _isLoading = true);

                    final formData = FormData();
                    formData.fields.add(MapEntry('billing_status', 'PAID_FULL'));
                    formData.fields.add(MapEntry('notes', notesCtrl.text.trim()));
                    if (proofPhoto != null) {
                      final bytes = await proofPhoto!.readAsBytes();
                      formData.files.add(MapEntry(
                        'file',
                        MultipartFile.fromBytes(bytes, filename: 'proof_${DateTime.now().millisecondsSinceEpoch}.jpg'),
                      ));
                    }

                    final ok = await salesService.collectMTPurchaseOrderPayment(
                      poId: selectedPoId!,
                      formData: formData,
                    );
                    setState(() => _isLoading = false);

                    if (mounted) {
                      setState(() => _collectionDone = true);
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(ok ? 'Bukti penagihan berhasil disimpan!' : 'Bukti penagihan dicatat di kunjungan.'),
                          backgroundColor: Colors.purple,
                        ),
                      );
                    }
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF7C3AED),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                  child: Text(
                    'SIMPAN BUKTI PENAGIHAN',
                    style: GoogleFonts.outfit(fontWeight: FontWeight.w900, color: Colors.white),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _proceedWithNoTransaction(String reason) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => VisitCheckoutPage(
          store: widget.store,
          selfiePath: widget.selfiePath,
          selfieBytes: widget.selfieBytes,
          receiptPath: null,
          notes: reason,
          jobPositionTitle: _jobPositionTitle,
          isDistributor: _isDistributor,
        ),
      ),
    );
  }

  void _proceedWithMTCompletion() {
    final completed = <String>[];
    if (_merchandisingDone) completed.add('Merchandising Selesai');
    if (_stockOpnameDone) completed.add('Stock Opname Tercatat');
    if (_collectionDone) completed.add('Penagihan TOP Tercatat');

    final summaryNotes = completed.isNotEmpty
        ? 'Kunjungan MT: ${completed.join(', ')}'
        : 'Kunjungan MT Selesai';

    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => VisitCheckoutPage(
          store: widget.store,
          selfiePath: widget.selfiePath,
          selfieBytes: widget.selfieBytes,
          receiptPath: null,
          notes: summaryNotes,
          jobPositionTitle: _jobPositionTitle,
          isDistributor: _isDistributor,
        ),
      ),
    );
  }

  Future<void> _handlePhysicalReceipt() async {
    String? cId;
    String? cName;

    try {
      final prefs = await SharedPreferences.getInstance();
      final cached = prefs.getString('cached_employee_profile');
      if (cached != null) {
        final profile = jsonDecode(cached) as Map<String, dynamic>;
        final parsedCId = profile['company_id'] ??
            profile['companyId'] ??
            profile['CompanyID'];
        final parsedCName = profile['company_name'] ??
            profile['companyName'] ??
            profile['CompanyNAME'] ??
            profile['company']?['name'];
        if (parsedCId != null) cId = parsedCId.toString();
        if (parsedCName != null) cName = parsedCName.toString();
      } else {
        final authRepo = di.sl<AuthRepository>();
        final result = await authRepo.getProfile();
        result.fold(
          (_) {},
          (profile) {
            final parsedCId = profile['company_id'] ??
                profile['companyId'] ??
                profile['CompanyID'];
            final parsedCName = profile['company_name'] ??
                profile['companyName'] ??
                profile['CompanyNAME'] ??
                profile['company']?['name'];
            if (parsedCId != null) cId = parsedCId.toString();
            if (parsedCName != null) cName = parsedCName.toString();
          },
        );
      }
    } catch (_) {}

    if (mounted) {
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (_) => ReceiptCameraPage(
            store: widget.store,
            selfiePath: widget.selfiePath,
            jobPositionTitle: _jobPositionTitle,
            companyId: cId,
            companyName: cName,
          ),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final bool hasMtActivityDone = _merchandisingDone || _stockOpnameDone || _collectionDone;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded, color: Color(0xFF334155)),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'LANGKAH 3 DARI 5',
              style: GoogleFonts.outfit(
                fontSize: 10,
                color: const Color(0xFF10B981),
                fontWeight: FontWeight.w800,
                letterSpacing: 1,
              ),
            ),
            Text(
              _isMT ? 'Aktivitas Kunjungan MT' : 'Status Transaksi',
              style: GoogleFonts.outfit(
                fontSize: 16,
                fontWeight: FontWeight.w800,
                color: const Color(0xFF1E293B),
              ),
            ),
          ],
        ),
      ),
      body: Stack(
        children: [
          Column(
            children: [
              LinearProgressIndicator(
                value: 0.75,
                minHeight: 4,
                backgroundColor: Colors.grey.shade200,
                color: const Color(0xFF10B981),
              ),
              Expanded(
                child: SingleChildScrollView(
                  physics: const BouncingScrollPhysics(),
                  padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Store Info Badge Card
                      Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(20),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.03),
                              blurRadius: 16,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: Row(
                          children: [
                            Container(
                              width: 48,
                              height: 48,
                              decoration: BoxDecoration(
                                color: const Color(0xFF10B981).withOpacity(0.12),
                                shape: BoxShape.circle,
                              ),
                              child: const Icon(Icons.storefront_rounded, color: Color(0xFF10B981), size: 24),
                            ),
                            const SizedBox(width: 14),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    widget.store.name,
                                    style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w800,
                                      fontSize: 15,
                                      color: const Color(0xFF1E293B),
                                    ),
                                  ),
                                  Text(
                                    widget.store.address,
                                    style: GoogleFonts.outfit(fontSize: 11, color: Colors.blueGrey),
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                  const SizedBox(height: 4),
                                  Row(
                                    children: [
                                      Container(
                                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                                        decoration: BoxDecoration(
                                          color: Colors.green.withOpacity(0.1),
                                          borderRadius: BorderRadius.circular(6),
                                        ),
                                        child: Text(
                                          '✓ Check-in Selfie OK',
                                          style: GoogleFonts.outfit(
                                            color: Colors.green.shade700,
                                            fontSize: 10,
                                            fontWeight: FontWeight.w700,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 20),

                      // MT MODE: Activity Checklist
                      if (_isMT) ...[
                        Row(
                          children: [
                            const Icon(Icons.checklist_rounded, size: 18, color: Color(0xFF0284C7)),
                            const SizedBox(width: 6),
                            Text(
                              'CHECKLIST AKTIVITAS GERAI MT',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                fontWeight: FontWeight.w900,
                                color: const Color(0xFF475569),
                                letterSpacing: 0.5,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 10),

                        // Task 1: Merchandising
                        _buildTaskCard(
                          icon: Icons.store_mall_directory_rounded,
                          iconColor: const Color(0xFFD97706),
                          bgColor: const Color(0xFFFEF3C7),
                          title: 'Audit Merchandising Rak',
                          subtitle: _merchandisingDone
                              ? 'Foto display & kepatuhan planogram telah tercatat'
                              : 'Ambil foto display, planogram & cek barang kosong (OOS)',
                          isCompleted: _merchandisingDone,
                          completedBadgeText: 'SELESAI',
                          pendingBadgeText: 'Lakukan',
                          onTap: _isLoading
                              ? null
                              : () async {
                                  final res = await Navigator.push<bool>(
                                    context,
                                    MaterialPageRoute(
                                      builder: (_) => MerchandisingPage(
                                        initialStore: widget.store,
                                      ),
                                    ),
                                  );
                                  if (res == true && mounted) {
                                    setState(() => _merchandisingDone = true);
                                  }
                                },
                        ),
                        const SizedBox(height: 10),

                        // Task 2: Stock Opname
                        _buildTaskCard(
                          icon: Icons.inventory_2_rounded,
                          iconColor: const Color(0xFF0284C7),
                          bgColor: const Color(0xFFE0F2FE),
                          title: 'Stock Opname & Sisa Stok',
                          subtitle: _stockOpnameDone
                              ? 'Saldo sisa stok fisik riil gerai telah disesuaikan'
                              : 'Hitung & catat sisa barang fisik riil (Pcs) di toko',
                          isCompleted: _stockOpnameDone,
                          completedBadgeText: 'TERCATAT',
                          pendingBadgeText: 'Hitung Stok',
                          onTap: _isLoading ? null : _showStockOpnameDialog,
                        ),
                        const SizedBox(height: 10),

                        // Task 3: Billing Collection (Optional)
                        _buildTaskCard(
                          icon: Icons.receipt_long_rounded,
                          iconColor: const Color(0xFF7C3AED),
                          bgColor: const Color(0xFFEDE9FE),
                          title: 'Penagihan Order / TOP (Opsional)',
                          subtitle: _collectionDone
                              ? 'Bukti bayar / pelunasan TOP berhasil diunggah'
                              : 'Upload bukti transfer jika ada penagihan jatuh tempo',
                          isCompleted: _collectionDone,
                          completedBadgeText: 'DITAGIH',
                          pendingBadgeText: 'Tagih',
                          onTap: _isLoading ? null : _showCollectionDialog,
                        ),
                        const SizedBox(height: 24),

                        // Section 2: Order / Transaksi Penjualan Baru
                        Row(
                          children: [
                            const Icon(Icons.shopping_cart_outlined, size: 18, color: Colors.blueGrey),
                            const SizedBox(width: 6),
                            Text(
                              'REPEAT ORDER / PO BARU (OPSIONAL)',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                fontWeight: FontWeight.w900,
                                color: const Color(0xFF475569),
                                letterSpacing: 0.5,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 10),
                        Row(
                          children: [
                            Expanded(
                              child: _buildActionButton(
                                icon: Icons.note_add_rounded,
                                label: 'Nota Digital',
                                sublabel: 'Buat PO di aplikasi',
                                color: const Color(0xFF10B981),
                                onTap: _isLoading ? null : _handleAutoSelection,
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: _buildActionButton(
                                icon: Icons.camera_alt_rounded,
                                label: 'Upload Nota Fisik',
                                sublabel: 'Foto nota dari toko',
                                color: const Color(0xFF4F46E5),
                                onTap: _isLoading ? null : _handlePhysicalReceipt,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 28),

                        // Completion Action Area for MT
                        if (hasMtActivityDone) ...[
                          Container(
                            width: double.infinity,
                            height: 56,
                            decoration: BoxDecoration(
                              borderRadius: BorderRadius.circular(18),
                              boxShadow: [
                                BoxShadow(
                                  color: const Color(0xFF10B981).withOpacity(0.35),
                                  blurRadius: 16,
                                  offset: const Offset(0, 6),
                                ),
                              ],
                            ),
                            child: ElevatedButton(
                              onPressed: _isLoading ? null : _proceedWithMTCompletion,
                              style: ElevatedButton.styleFrom(
                                backgroundColor: const Color(0xFF10B981),
                                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(18)),
                                elevation: 0,
                              ),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  const Icon(Icons.check_circle_rounded, color: Colors.white, size: 22),
                                  const SizedBox(width: 10),
                                  Text(
                                    'SELESAIKAN KUNJUNGAN',
                                    style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w900,
                                      fontSize: 15,
                                      color: Colors.white,
                                      letterSpacing: 0.5,
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  const Icon(Icons.arrow_forward_rounded, color: Colors.white, size: 20),
                                ],
                              ),
                            ),
                          ),
                          const SizedBox(height: 12),
                        ],

                        // Fallback options
                        Row(
                          children: [
                            Expanded(
                              child: OutlinedButton(
                                onPressed: _isLoading
                                    ? null
                                    : () => _proceedWithNoTransaction('Input Transaksi Nanti'),
                                style: OutlinedButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(vertical: 14),
                                  side: BorderSide(color: Colors.orange.shade200),
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                                  backgroundColor: Colors.orange.shade50.withOpacity(0.5),
                                ),
                                child: Text(
                                  'Input Nanti',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w700,
                                    fontSize: 12,
                                    color: Colors.orange.shade800,
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 10),
                            Expanded(
                              child: OutlinedButton(
                                onPressed: _isLoading ? null : _showNoTransactionDialog,
                                style: OutlinedButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(vertical: 14),
                                  side: BorderSide(color: Colors.grey.shade300),
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                                  backgroundColor: Colors.white,
                                ),
                                child: Text(
                                  'Tidak Ada Transaksi',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w700,
                                    fontSize: 12,
                                    color: Colors.grey.shade700,
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ] else ...[
                        // GENERAL TRADE (GT / Regular Sales) MODE - 100% Original UI
                        Center(
                          child: Column(
                            children: [
                              Text(
                                'Apakah ada transaksi\ndi kunjungan ini?',
                                textAlign: TextAlign.center,
                                style: GoogleFonts.outfit(
                                  fontSize: 22,
                                  fontWeight: FontWeight.w900,
                                  color: const Color(0xFF1E293B),
                                  height: 1.2,
                                ),
                              ),
                              const SizedBox(height: 8),
                              Text(
                                'Pilih jenis transaksi untuk melanjutkan',
                                textAlign: TextAlign.center,
                                style: GoogleFonts.outfit(
                                  fontSize: 13,
                                  fontWeight: FontWeight.w500,
                                  color: Colors.blueGrey,
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 24),

                        // Option 1: Digital Invoice (Green)
                        Container(
                          width: double.infinity,
                          margin: const EdgeInsets.only(bottom: 12),
                          child: ElevatedButton(
                            onPressed: _isLoading ? null : _handleAutoSelection,
                            style: ElevatedButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 20),
                              backgroundColor: Colors.green,
                              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                              elevation: 4,
                              shadowColor: Colors.green.withOpacity(0.3),
                            ),
                            child: Row(
                              children: [
                                Container(
                                  padding: const EdgeInsets.all(10),
                                  decoration: BoxDecoration(
                                    color: Colors.white.withOpacity(0.2),
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  child: _isLoading
                                      ? const SizedBox(
                                          width: 24,
                                          height: 24,
                                          child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
                                      : const Icon(Icons.note_add_rounded, color: Colors.white, size: 24),
                                ),
                                const SizedBox(width: 16),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text('Input Nota Digital',
                                          style: GoogleFonts.outfit(
                                              fontWeight: FontWeight.w900, fontSize: 16, color: Colors.white)),
                                      Text('Buat pesanan di aplikasi',
                                          style: GoogleFonts.outfit(
                                              fontWeight: FontWeight.w500,
                                              fontSize: 12,
                                              color: Colors.white.withOpacity(0.9))),
                                    ],
                                  ),
                                ),
                                const Icon(Icons.chevron_right_rounded, color: Colors.white),
                              ],
                            ),
                          ),
                        ),

                        // Option 2: Physical Receipt (Indigo)
                        Container(
                          width: double.infinity,
                          margin: const EdgeInsets.only(bottom: 16),
                          child: ElevatedButton(
                            onPressed: _isLoading ? null : _handlePhysicalReceipt,
                            style: ElevatedButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 20),
                              backgroundColor: Colors.indigo,
                              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                              elevation: 4,
                              shadowColor: Colors.indigo.withOpacity(0.3),
                            ),
                            child: Row(
                              children: [
                                Container(
                                  padding: const EdgeInsets.all(10),
                                  decoration: BoxDecoration(
                                    color: Colors.white.withOpacity(0.2),
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  child: const Icon(Icons.camera_alt_rounded, color: Colors.white, size: 24),
                                ),
                                const SizedBox(width: 16),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text('Upload Nota Fisik',
                                          style: GoogleFonts.outfit(
                                              fontWeight: FontWeight.w900, fontSize: 16, color: Colors.white)),
                                      Text('Foto nota fisik dari toko',
                                          style: GoogleFonts.outfit(
                                              fontWeight: FontWeight.w500,
                                              fontSize: 12,
                                              color: Colors.white.withOpacity(0.9))),
                                    ],
                                  ),
                                ),
                                const Icon(Icons.chevron_right_rounded, color: Colors.white),
                              ],
                            ),
                          ),
                        ),

                        // Option 3 & 4: Input Nanti & Tidak Ada
                        Row(
                          children: [
                            Expanded(
                              child: OutlinedButton(
                                onPressed: _isLoading
                                    ? null
                                    : () => _proceedWithNoTransaction('Input Transaksi Nanti'),
                                style: OutlinedButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(vertical: 16),
                                  side: BorderSide(color: Colors.orange.shade300, width: 2),
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                                  backgroundColor: Colors.orange.shade50,
                                ),
                                child: Column(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Icon(Icons.schedule_rounded, color: Colors.orange.shade700, size: 26),
                                    const SizedBox(height: 8),
                                    Text('Input Nanti',
                                        style: GoogleFonts.outfit(
                                            fontWeight: FontWeight.w800,
                                            fontSize: 14,
                                            color: Colors.orange.shade800)),
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: OutlinedButton(
                                onPressed: _isLoading ? null : _showNoTransactionDialog,
                                style: OutlinedButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(vertical: 16),
                                  side: BorderSide(color: Colors.grey.shade300, width: 2),
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                                  backgroundColor: Colors.white,
                                ),
                                child: Column(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Icon(Icons.close_rounded, color: Colors.grey.shade600, size: 26),
                                    const SizedBox(height: 8),
                                    Text('Tidak Ada',
                                        style: GoogleFonts.outfit(
                                            fontWeight: FontWeight.w800,
                                            fontSize: 14,
                                            color: Colors.grey.shade700)),
                                  ],
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                      const SizedBox(height: 32),
                    ],
                  ),
                ),
              ),
            ],
          ),
          if (_isLoading)
            Container(
              color: Colors.black.withOpacity(0.2),
              child: const Center(child: CircularProgressIndicator(color: Color(0xFF10B981))),
            ),
        ],
      ),
    );
  }

  Widget _buildTaskCard({
    required IconData icon,
    required Color iconColor,
    required Color bgColor,
    required String title,
    required String subtitle,
    required bool isCompleted,
    String completedBadgeText = 'SELESAI',
    String pendingBadgeText = 'Lakukan',
    VoidCallback? onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(18),
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(18),
          border: Border.all(
            color: isCompleted ? Colors.green.shade300 : const Color(0xFFE2E8F0),
            width: isCompleted ? 1.5 : 1,
          ),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 10,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Row(
          children: [
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: isCompleted ? Colors.green.shade50 : bgColor,
                borderRadius: BorderRadius.circular(14),
              ),
              child: Icon(
                isCompleted ? Icons.check_circle_rounded : icon,
                color: isCompleted ? Colors.green.shade700 : iconColor,
                size: 22,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w800,
                      fontSize: 14,
                      color: const Color(0xFF1E293B),
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    subtitle,
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: Colors.blueGrey,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
              decoration: BoxDecoration(
                color: isCompleted ? Colors.green.shade600 : const Color(0xFFF1F5F9),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    isCompleted ? '✓ $completedBadgeText' : pendingBadgeText,
                    style: GoogleFonts.outfit(
                      fontSize: 10,
                      fontWeight: FontWeight.w800,
                      color: isCompleted ? Colors.white : const Color(0xFF475569),
                    ),
                  ),
                  if (!isCompleted) ...[
                    const SizedBox(width: 2),
                    const Icon(Icons.chevron_right_rounded, size: 14, color: Color(0xFF475569)),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildActionButton({
    required IconData icon,
    required String label,
    required String sublabel,
    required Color color,
    VoidCallback? onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: color.withOpacity(0.08),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: color.withOpacity(0.2)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: color,
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(icon, color: Colors.white, size: 18),
            ),
            const SizedBox(height: 10),
            Text(
              label,
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w800,
                fontSize: 13,
                color: const Color(0xFF1E293B),
              ),
            ),
            const SizedBox(height: 2),
            Text(
              sublabel,
              style: GoogleFonts.outfit(
                fontSize: 10,
                color: Colors.blueGrey,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
