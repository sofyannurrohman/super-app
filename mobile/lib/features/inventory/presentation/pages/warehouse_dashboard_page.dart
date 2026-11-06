import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:intl/intl.dart';
import 'package:shimmer/shimmer.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:hris_app/core/utils/constants.dart';

class WarehouseDashboardPage extends StatefulWidget {
  const WarehouseDashboardPage({super.key});

  @override
  State<WarehouseDashboardPage> createState() => _WarehouseDashboardPageState();
}

class _WarehouseDashboardPageState extends State<WarehouseDashboardPage>
    with SingleTickerProviderStateMixin {
  final ApiClient apiClient = di.sl<ApiClient>();
  late TabController _tabController;
  bool _isLoading = false;
  final TextEditingController _searchController = TextEditingController();
  final PageController _pageController = PageController();
  int _currentPage = 0;

  List<dynamic> _inventory = [];
  List<dynamic> _filteredInventory = [];
  List<dynamic> _pendingInbound = [];
  List<dynamic> _pendingOutbound = [];
  List<dynamic> _salesRequests = [];
  List<dynamic> _logs = [];
  List<dynamic> _filteredLogs = [];
  final TextEditingController _logSearchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _fetchAllData();
  }

  @override
  void dispose() {
    _tabController.dispose();
    _pageController.dispose();
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _fetchAllData() async {
    setState(() => _isLoading = true);
    try {
      await Future.wait([
        _fetchInventory(),
        _fetchPendingInbound(),
        _fetchPendingOutbound(),
        _fetchSalesRequests(),
        _fetchLogs(),
      ]);
    } catch (e) {
      _showError('Gagal mengambil data: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _fetchInventory() async {
    final response = await apiClient.client.get('warehouse/stock');
    setState(() {
      List<dynamic> rawInventory = [];
      if (response.data is Map && response.data.containsKey('data')) {
        rawInventory = (response.data['data'] as List<dynamic>?) ?? [];
      } else {
        rawInventory = (response.data as List<dynamic>?) ?? [];
      }

      Map<String, dynamic> aggregated = {};
      for (var item in rawInventory) {
        final productId = item['product_id']?.toString() ?? '';
        if (productId.isEmpty) continue;

        final stockType = item['stock_type']?.toString() ?? 'GOOD';
        if (stockType != 'GOOD') continue;

        final qty = (item['quantity'] as num?)?.toInt() ?? 0;
        final res = (item['reserved_quantity'] as num?)?.toInt() ?? 0;
        final minLimit = (item['min_limit'] as num?)?.toInt() ?? 0;
        final batchNo = item['batch_no']?.toString() ?? '';
        final expiry = item['expiry_date']?.toString() ?? '';

        if (!aggregated.containsKey(productId)) {
          aggregated[productId] = {
            'product_id': productId,
            'product': item['product'],
            'quantity': 0,
            'reserved_quantity': 0,
            'min_limit': minLimit,
            'backorder': 0,
            'batches': [],
          };
        }

        aggregated[productId]['reserved_quantity'] =
            (aggregated[productId]['reserved_quantity'] as int) + res;

        if (minLimit > (aggregated[productId]['min_limit'] as int)) {
          aggregated[productId]['min_limit'] = minLimit;
        }

        if (batchNo == 'BACKORDER') {
          aggregated[productId]['backorder'] =
              (aggregated[productId]['backorder'] as int) + qty;
        } else {
          aggregated[productId]['quantity'] =
              (aggregated[productId]['quantity'] as int) + qty;

          if (qty > 0 || res > 0) {
            (aggregated[productId]['batches'] as List).add({
              'batch_no': batchNo,
              'quantity': qty,
              'reserved_quantity': res,
              'expiry_date': expiry,
            });
          }
        }
      }

      _inventory = aggregated.values.toList();

      _filteredInventory = _inventory.where((item) {
        final qty = (item['quantity'] as num?)?.toInt() ?? 0;
        final res = (item['reserved_quantity'] as num?)?.toInt() ?? 0;
        final min = (item['min_limit'] as num?)?.toInt() ?? 0;
        final bo = (item['backorder'] as num?)?.toInt() ?? 0;
        return qty > 0 || res > 0 || min > 0 || bo > 0;
      }).toList();
    });
  }

  Future<void> _fetchPendingInbound() async {
    final response = await apiClient.client.get('warehouse/transfers/pending');
    setState(() {
      if (response.data is Map && response.data.containsKey('data')) {
        _pendingInbound = (response.data['data'] as List<dynamic>?) ?? [];
      } else {
        _pendingInbound = (response.data as List<dynamic>?) ?? [];
      }
    });
  }

  Future<void> _fetchPendingOutbound() async {
    final response = await apiClient.client.get('delivery/batches/warehouse');
    setState(() {
      if (response.data is Map && response.data.containsKey('data')) {
        _pendingOutbound = (response.data['data'] as List<dynamic>?) ?? [];
      } else {
        _pendingOutbound = (response.data as List<dynamic>?) ?? [];
      }
    });
  }

  Future<void> _fetchSalesRequests() async {
    final response = await apiClient.client.get('sales-transfers/pending');
    setState(() {
      if (response.data is Map && response.data.containsKey('data')) {
        _salesRequests = (response.data['data'] as List<dynamic>?) ?? [];
      } else {
        _salesRequests = (response.data as List<dynamic>?) ?? [];
      }
    });
  }

  Future<void> _fetchLogs() async {
    final response = await apiClient.client.get('warehouse/logs');
    setState(() {
      if (response.data is Map && response.data.containsKey('data')) {
        _logs = (response.data['data'] as List<dynamic>?) ?? [];
      } else {
        _logs = (response.data as List<dynamic>?) ?? [];
      }
      _filteredLogs = _logs;
    });
  }

  void _filterLogs(String query) {
    setState(() {
      if (query.isEmpty) {
        _filteredLogs = _logs;
      } else {
        _filteredLogs = _logs.where((log) {
          final productName = (log['product']?['name'] ?? '')
              .toString()
              .toLowerCase();
          return productName.contains(query.toLowerCase());
        }).toList();
      }
    });
  }

  void _filterInventory(String query) {
    setState(() {
      if (query.isEmpty) {
        _filteredInventory = _inventory.where((item) {
          final qty = (item['quantity'] ?? 0);
          final res = (item['reserved_quantity'] ?? 0);
          final min = (item['min_limit'] ?? 0);
          return qty > 0 || res > 0 || min > 0;
        }).toList();
      } else {
        _filteredInventory = _inventory.where((item) {
          final productName = (item['product']['name'] ?? '')
              .toString()
              .toLowerCase();
          final sku = (item['product']['sku'] ?? '').toString().toLowerCase();
          final matchesSearch =
              productName.contains(query.toLowerCase()) ||
              sku.contains(query.toLowerCase());

          if (!matchesSearch) return false;

          final qty = (item['quantity'] ?? 0);
          final res = (item['reserved_quantity'] ?? 0);
          final min = (item['min_limit'] ?? 0);
          return qty > 0 || res > 0 || min > 0;
        }).toList();
      }
    });
  }

  void _showError(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          msg,
          style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w600),
        ),
        backgroundColor: AppColors.error,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),
    );
  }

  void _showSuccess(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          msg,
          style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w600),
        ),
        backgroundColor: AppColors.success,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),
    );
  }

  // --- LOGIC: RECEIVE FROM FACTORY ---
  Future<void> _processInboundScan(String doNo) async {
    setState(() => _isLoading = true);
    try {
      final response = await apiClient.client.get(
        'warehouse/transfers/do/$doNo',
      );
      final transfer = response.data;
      if (mounted) _showReceiveDialog(transfer, doNo);
    } catch (e) {
      _showError('Surat Jalan tidak ditemukan.');
    } finally {
      setState(() => _isLoading = false);
    }
  }

  // --- UTILS ---
  String _formatStock(dynamic qty, dynamic product) {
    if (qty == null) return '0';
    final q = qty is num ? qty : double.tryParse(qty.toString()) ?? 0;
    final p = product ?? {};
    final unit = (p['unit'] ?? 'Unit').toString();
    final itemsPerUnit = p['items_per_unit'] ?? p['pcs_per_unit'] ?? 0;

    if (itemsPerUnit > 1) {
      final major = (q / itemsPerUnit).floor();
      final pcs = (q % itemsPerUnit).toInt();

      List<String> parts = [];
      if (major > 0) parts.add('$major $unit');
      if (pcs > 0) parts.add('$pcs Pcs');

      return parts.isEmpty ? '0 Pcs' : parts.join(' ');
    }

    return '$q $unit';
  }

  void _showReceiveDialog(Map<String, dynamic> tr, String doNo) {
    int totalQty = (tr['quantity'] ?? 0) as int;
    final goodController = TextEditingController(text: totalQty.toString());
    final damagedController = TextEditingController(text: '0');

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) {
          return AlertDialog(
            backgroundColor: Colors.white,
            surfaceTintColor: Colors.white,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(28),
            ),
            title: Row(
              children: [
                const Icon(
                  Icons.download_for_offline_rounded,
                  color: AppColors.success,
                  size: 28,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    'Terima Barang',
                    style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.w800,
                      fontSize: 18,
                    ),
                  ),
                ),
              ],
            ),
            content: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Penerimaan dari pabrik dengan Surat Jalan:',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      color: AppColors.textSecondary,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: AppColors.success.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: AppColors.success.withOpacity(0.2),
                      ),
                    ),
                    child: Text(
                      doNo,
                      textAlign: TextAlign.center,
                      style: GoogleFonts.plusJakartaSans(
                        color: AppColors.success,
                        fontWeight: FontWeight.w800,
                        fontSize: 16,
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  _buildDetailBox(
                    tr['product']?['name'] ?? 'Produk Tidak Diketahui',
                    'Total Dikirim: $totalQty ${tr['product']?['unit'] ?? 'Unit'}',
                  ),
                  const SizedBox(height: 20),
                  Text(
                    'Kondisi Stok (Bisa disesuaikan):',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 10),
                  Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Kondisi Baik',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 11,
                                color: AppColors.textSecondary,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            const SizedBox(height: 4),
                            TextField(
                              controller: goodController,
                              keyboardType: TextInputType.number,
                              decoration: InputDecoration(
                                isDense: true,
                                contentPadding: const EdgeInsets.symmetric(
                                  horizontal: 12,
                                  vertical: 12,
                                ),
                                border: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(8),
                                ),
                              ),
                              onChanged: (val) {
                                if (val.isEmpty) return;
                                int g = int.tryParse(val) ?? 0;
                                int d = totalQty - g;
                                if (d < 0) d = 0;
                                if (damagedController.text != d.toString()) {
                                  damagedController.text = d.toString();
                                }
                              },
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Kondisi Rusak',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 11,
                                color: AppColors.textSecondary,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            const SizedBox(height: 4),
                            TextField(
                              controller: damagedController,
                              keyboardType: TextInputType.number,
                              decoration: InputDecoration(
                                isDense: true,
                                contentPadding: const EdgeInsets.symmetric(
                                  horizontal: 12,
                                  vertical: 12,
                                ),
                                border: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(8),
                                ),
                              ),
                              onChanged: (val) {
                                if (val.isEmpty) return;
                                int d = int.tryParse(val) ?? 0;
                                int g = totalQty - d;
                                if (g < 0) g = 0;
                                if (goodController.text != g.toString()) {
                                  goodController.text = g.toString();
                                }
                              },
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context),
                child: Text(
                  'BATAL',
                  style: GoogleFonts.plusJakartaSans(
                    fontWeight: FontWeight.w700,
                    color: AppColors.textTertiary,
                  ),
                ),
              ),
              ElevatedButton(
                onPressed: () {
                  int finalGood = int.tryParse(goodController.text) ?? 0;
                  int finalDamaged = int.tryParse(damagedController.text) ?? 0;
                  Navigator.pop(context);
                  _confirmReceive(doNo, finalGood, finalDamaged);
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.success,
                  minimumSize: const Size(100, 44),
                ),
                child: const Text('TERIMA'),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<void> _confirmReceive(String doNo, int goodQty, int damagedQty) async {
    try {
      await apiClient.client.post(
        'warehouse/transfers/do/$doNo/receive',
        data: {
          'actual_received_quantity': goodQty,
          'damaged_quantity': damagedQty,
        },
      );
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('✓ Barang berhasil masuk stok!'),
          backgroundColor: Colors.green,
        ),
      );
      _fetchAllData();
    } catch (e) {
      _showError('Gagal: $e');
    }
  }

  // --- LOGIC: DISPATCH FOR SALES (INVOICE) ---
  Future<void> _processOutboundScan(String code) async {
    setState(() => _isLoading = true);
    try {
      // 1. Try fetching as a Delivery Batch (Manifest/Surat Jalan)
      try {
        final batchResp = await apiClient.client.get('delivery/batch/$code');
        if (batchResp.data != null) {
          final batch =
              batchResp.data is Map &&
                  (batchResp.data as Map).containsKey('data')
              ? batchResp.data['data']
              : batchResp.data;
          if (mounted) {
            setState(() => _isLoading = false);
            _showBatchDetailSheet(batch);
            return;
          }
        }
      } catch (_) {
        // Fallback to single Sales Order
      }

      // 2. Try fetching as a single Sales Order (Nota)
      final response = await apiClient.client.get(
        'sales/transactions/receipt/$code',
      );
      Map<String, dynamic> transaction;
      if (response.data is Map && response.data.containsKey('data')) {
        transaction = response.data['data'];
      } else {
        transaction = response.data;
      }
      if (mounted) _showDispatchSheet(transaction);
    } catch (e) {
      _showError('Barcode tidak dikenali atau data tidak ditemukan.');
    } finally {
      setState(() => _isLoading = false);
    }
  }

  void _showBatchDetailSheet(Map<String, dynamic> batch) {
    final items = (batch['items'] as List<dynamic>?) ?? [];
    final driver = batch['driver']?['first_name'] ?? 'Driver';
    final vehicle =
        batch['vehicle']?['name'] ??
        batch['vehicle']?['license_plate'] ??
        'Armada';
    final doNo = batch['delivery_order_no'] ?? '-';
    final status = batch['status'] ?? '';
    final batchId = batch['id']?.toString() ?? '';

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.9,
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
        ),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: AppColors.info.withOpacity(0.05),
                borderRadius: const BorderRadius.vertical(
                  top: Radius.circular(32),
                ),
                border: Border(
                  bottom: BorderSide(color: AppColors.info.withOpacity(0.1)),
                ),
              ),
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: AppColors.info,
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: const Icon(
                      Icons.local_shipping_rounded,
                      color: Colors.white,
                      size: 28,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'DETAIL MANIFEST',
                          style: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w900,
                            fontSize: 16,
                            color: AppColors.textPrimary,
                          ),
                        ),
                        Text(
                          'No. SJ: $doNo',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 12,
                            color: AppColors.textSecondary,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(Icons.close_rounded),
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(20),
              child: Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: AppColors.grayBorder),
                ),
                child: Row(
                  children: [
                    _buildInfoColumn('DRIVER', driver, Icons.person_rounded),
                    const SizedBox(width: 24),
                    _buildInfoColumn(
                      'ARMADA',
                      vehicle,
                      Icons.directions_car_rounded,
                    ),
                    const SizedBox(width: 24),
                    _buildInfoColumn(
                      'TOTAL NOTA',
                      '${items.length}',
                      Icons.description_rounded,
                    ),
                  ],
                ),
              ),
            ),
            Expanded(
              child: ListView.builder(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                itemCount: items.length,
                itemBuilder: (context, index) {
                  final it = items[index];
                  final so = it['sales_order'] ?? it['sales_transaction'] ?? {};
                  final store = so['store']?['name'] ?? 'Toko';
                  final receipt =
                      so['receipt_no'] ?? so['receipt_number'] ?? 'No Nota';
                  final status = it['status'] ?? 'PENDING';

                  return Container(
                    margin: const EdgeInsets.only(bottom: 12),
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: AppColors.grayBorder.withOpacity(0.5),
                      ),
                    ),
                    child: Row(
                      children: [
                        Container(
                          width: 40,
                          height: 40,
                          decoration: BoxDecoration(
                            color: AppColors.info.withOpacity(0.1),
                            shape: BoxShape.circle,
                          ),
                          child: Center(
                            child: Text(
                              '${index + 1}',
                              style: GoogleFonts.plusJakartaSans(
                                fontWeight: FontWeight.w800,
                                color: AppColors.info,
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                store,
                                style: GoogleFonts.plusJakartaSans(
                                  fontWeight: FontWeight.w800,
                                  fontSize: 14,
                                ),
                              ),
                              Text(
                                'Nota: $receipt',
                                style: GoogleFonts.plusJakartaSans(
                                  fontSize: 11,
                                  color: AppColors.textSecondary,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ],
                          ),
                        ),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: status == 'PENDING'
                                ? Colors.orange.shade50
                                : Colors.green.shade50,
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            status,
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 9,
                              fontWeight: FontWeight.w900,
                              color: status == 'PENDING'
                                  ? Colors.orange.shade700
                                  : Colors.green.shade700,
                            ),
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
            if (status == 'SUPERVISOR_APPROVED')
              Padding(
                padding: const EdgeInsets.all(20),
                child: ElevatedButton(
                  onPressed: () {
                    Navigator.pop(context);
                    _confirmWarehouseReady(batchId);
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.success,
                    minimumSize: const Size(double.infinity, 56),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20),
                    ),
                    elevation: 8,
                    shadowColor: AppColors.success.withOpacity(0.4),
                  ),
                  child: Text(
                    'KONFIRMASI SIAP PENGIRIMAN',
                    style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.w800,
                      color: Colors.white,
                      fontSize: 15,
                      letterSpacing: 1,
                    ),
                  ),
                ),
              )
            else if (status == 'PICKING')
              Padding(
                padding: const EdgeInsets.all(20),
                child: ElevatedButton(
                  onPressed: () {
                    Navigator.pop(context);
                    _confirmBatchStart(doNo);
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.info,
                    minimumSize: const Size(double.infinity, 56),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20),
                    ),
                    elevation: 8,
                    shadowColor: AppColors.info.withOpacity(0.4),
                  ),
                  child: Text(
                    'KONFIRMASI KEBERANGKATAN',
                    style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.w800,
                      color: Colors.white,
                      fontSize: 15,
                      letterSpacing: 1,
                    ),
                  ),
                ),
              )
            else
              const SizedBox.shrink(),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoColumn(String label, String value, IconData icon) {
    return Expanded(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 10, color: AppColors.textTertiary),
              const SizedBox(width: 4),
              Text(
                label,
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 8,
                  fontWeight: FontWeight.w800,
                  color: AppColors.textTertiary,
                  letterSpacing: 0.5,
                ),
              ),
            ],
          ),
          const SizedBox(height: 4),
          Text(
            value,
            style: GoogleFonts.plusJakartaSans(
              fontSize: 11,
              fontWeight: FontWeight.w900,
              color: AppColors.textPrimary,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ),
    );
  }

  Future<void> _confirmBatchStart(String batchNo) async {
    setState(() => _isLoading = true);
    try {
      await apiClient.client.post('delivery/batch/$batchNo/start');
      _showSuccess('Batch pengiriman berhasil dimulai. Status: ON DELIVERY');
      _fetchAllData();
    } catch (e) {
      _showError('Gagal memproses keberangkatan batch.');
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _confirmWarehouseReady(String batchId) async {
    setState(() => _isLoading = true);
    try {
      await apiClient.client.post('delivery/batch/$batchId/warehouse-ready');
      _showSuccess('Surat Jalan dikonfirmasi siap. Status: PICKING');
      _fetchAllData();
    } catch (e) {
      _showError('Gagal mengonfirmasi persiapan Surat Jalan.');
    } finally {
      setState(() => _isLoading = false);
    }
  }

  void _showDispatchSheet(Map<String, dynamic> tr) {
    final items = tr['items'] as List<dynamic>;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.8,
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.only(
            topLeft: Radius.circular(32),
            topRight: Radius.circular(32),
          ),
        ),
        child: Column(
          children: [
            const SizedBox(height: 12),
            Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: AppColors.grayBorder,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            const SizedBox(height: 20),
            Text(
              'Kirim ke Sales (Invoice)',
              style: GoogleFonts.plusJakartaSans(
                fontSize: 20,
                fontWeight: FontWeight.w800,
              ),
            ),
            Text(
              'NOTA: ${tr['receipt_no']}',
              style: GoogleFonts.plusJakartaSans(
                color: AppColors.info,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 20),
            Expanded(
              child: ListView.separated(
                padding: const EdgeInsets.symmetric(horizontal: 24),
                itemCount: items.length,
                separatorBuilder: (_, __) =>
                    const Divider(height: 1, color: AppColors.grayLight),
                itemBuilder: (context, index) {
                  final it = items[index];
                  return Padding(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    child: Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.grayLight,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: const Icon(
                            Icons.inventory_2_outlined,
                            size: 18,
                            color: AppColors.textSecondary,
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Text(
                            it['product']?['name'] ?? 'Produk',
                            style: GoogleFonts.plusJakartaSans(
                              fontWeight: FontWeight.w700,
                              color: AppColors.textPrimary,
                            ),
                          ),
                        ),
                        Text(
                          '${it['quantity'] ?? 0} ${it['product']?['unit'] ?? 'Unit'}',
                          style: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w800,
                            fontSize: 16,
                            color: AppColors.textPrimary,
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(24),
              child: ElevatedButton(
                onPressed: () {
                  Navigator.pop(context);
                  _confirmDispatch(tr['receipt_no']);
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.success,
                  minimumSize: const Size(double.infinity, 64),
                ),
                child: const Text('KONFIRMASI PENGELUARAN'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _confirmDispatch(String receiptNo) async {
    try {
      await apiClient.client.post('warehouse/dispatch/invoice/$receiptNo');
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('✓ Barang berhasil dikirim!'),
          backgroundColor: Colors.green,
        ),
      );
      _fetchAllData();
    } catch (e) {
      _showError('Gagal: $e');
    }
  }

  // --- LOGIC: APPROVE SALES REQUEST ---
  Future<void> _processSalesRequestApproval(String id) async {
    setState(() => _isLoading = true);
    try {
      await apiClient.client.patch('sales-transfers/$id/complete');
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('✓ Permintaan stok disetujui!'),
          backgroundColor: Colors.green,
        ),
      );
      _fetchAllData();
    } catch (e) {
      _showError('Gagal menyetujui: $e');
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Widget _buildSkeleton() {
    return Shimmer.fromColors(
      baseColor: Colors.grey.shade300,
      highlightColor: Colors.grey.shade100,
      child: SingleChildScrollView(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Column(
          children: [
            const SizedBox(height: 10),
            Container(
              height: 200,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
              ),
            ),
            const SizedBox(height: 24),
            Row(
              children: [
                Expanded(
                  child: Container(
                    height: 100,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Container(
                    height: 100,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 32),
            Container(
              height: 150,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC), // Clean Light Ice Blue
      appBar: AppBar(
        title: BlocBuilder<AuthBloc, AuthState>(
          builder: (context, state) {
            String companyName = 'Perusahaan Demo';
            String branchName = '';

            if (state is Authenticated) {
              final profile = state.userProfile;
              if (profile != null) {
                final company = profile['company'];
                if (company != null && company['name'] != null) {
                  companyName = company['name'].toString().toUpperCase();
                }
                final branch = profile['branch'];
                if (branch != null && branch['name'] != null) {
                  branchName = branch['name'].toString().toUpperCase();
                }
              }
            }

            final subtitle = branchName.isNotEmpty
                ? '$companyName - $branchName'
                : companyName;

            return Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  'GUDANG & LOGISTIK',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w900,
                    color: Colors.white,
                    fontSize: 14,
                    letterSpacing: 1,
                  ),
                ),
                Text(
                  subtitle,
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w700,
                    color: AppColors.blueWarehouseLight,
                    fontSize: 8,
                    letterSpacing: 1,
                  ),
                ),
              ],
            );
          },
        ),
        backgroundColor: AppColors.blueWarehouseDark, // Elegant Dark Steel Blue
        centerTitle: true,
        elevation: 0,
        automaticallyImplyLeading: false,
        leadingWidth: 100,
        leading: Center(
          child: Padding(
            padding: const EdgeInsets.only(left: 12),
            child: TextButton(
              onPressed: () => Navigator.pop(context),
              style: TextButton.styleFrom(
                backgroundColor: Colors.white.withOpacity(0.1),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              ),
              child: Text(
                'KEMBALI',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800,
                  fontSize: 11,
                  color: Colors.white,
                ),
              ),
            ),
          ),
        ),
        actions: [
          IconButton(
            onPressed: _fetchAllData,
            icon: const Icon(Icons.refresh_rounded, color: Colors.white),
          ),
          IconButton(
            onPressed: _showScanChoiceMenu,
            icon: const Icon(
              Icons.qr_code_scanner_rounded,
              color: Colors.white,
            ),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: _isLoading
          ? _buildSkeleton()
          : RefreshIndicator(
              onRefresh: _fetchAllData,
              child: SingleChildScrollView(
                physics: const BouncingScrollPhysics(),
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 12,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const SizedBox(height: 10),
                    _buildSwipeableHeader(),
                    const SizedBox(height: 24),
                    _buildOperationalAccess(),
                    const SizedBox(height: 32),
                    _buildRecentMutations(),
                    const SizedBox(height: 40),
                  ],
                ),
              ),
            ),
    );
  }

  Widget _buildSwipeableHeader() {
    return Column(
      children: [
        SizedBox(
          height: 380,
          child: PageView(
            controller: _pageController,
            physics: const BouncingScrollPhysics(),
            onPageChanged: (v) => setState(() => _currentPage = v),
            children: [_buildInventorySummaryCard(), _buildTaskSummaryCard()],
          ),
        ),
        const SizedBox(height: 12),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: List.generate(
            2,
            (index) => Container(
              margin: const EdgeInsets.symmetric(horizontal: 4),
              width: _currentPage == index ? 24 : 8,
              height: 8,
              decoration: BoxDecoration(
                color: _currentPage == index
                    ? AppColors.blueWarehouse
                    : const Color(0xFFCBD5E1),
                borderRadius: BorderRadius.circular(4),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildInventorySummaryCard() {
    int lowStock = _inventory
        .where((item) => (item['quantity'] ?? 0) <= (item['min_limit'] ?? 0))
        .length;

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 2),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: AppColors.blueWarehouseDark.withOpacity(0.08),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
        border: Border.all(
          color: AppColors.blueWarehouse.withOpacity(0.1),
          width: 1.5,
        ),
      ),
      child: Stack(
        children: [
          Positioned(
            right: -10,
            top: -10,
            child: Icon(
              Icons.inventory_2_rounded,
              size: 100,
              color: AppColors.blueWarehouseLight.withOpacity(0.04),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: AppColors.blueWarehouseLight.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Icon(
                        Icons.analytics_rounded,
                        color: AppColors.blueWarehouse,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Text(
                      'RINGKASAN INVENTORY',
                      style: GoogleFonts.outfit(
                        color: AppColors.blueWarehouseDark,
                        fontWeight: FontWeight.w800,
                        fontSize: 11,
                        letterSpacing: 1.5,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Expanded(
                  child: _inventory.isEmpty
                      ? Center(
                          child: Text(
                            'Tidak ada data produk',
                            style: GoogleFonts.outfit(
                              color: Colors.blueGrey,
                              fontSize: 12,
                            ),
                          ),
                        )
                      : ListView.separated(
                          physics: const BouncingScrollPhysics(),
                          itemCount: _inventory.length,
                          separatorBuilder: (context, index) => const Divider(
                            height: 16,
                            color: Color(0xFFF1F5F9),
                          ),
                          itemBuilder: (context, index) {
                            final item = _inventory[index];
                            final product = item['product'];
                            final productName = product != null
                                ? product['name']
                                : 'Produk tidak diketahui';
                            final photoUrl = product != null
                                ? product['image_url']
                                : null;
                            final qty = item['quantity'] ?? 0;
                            final limit = item['min_limit'] ?? 0;
                            final backorder = item['backorder'] ?? 0;
                            final List batches = item['batches'] ?? [];

                            Color qtyColor = AppColors.blueWarehouse;
                            if (qty <= 0) {
                              qtyColor = const Color(0xFFB91C1C);
                            } else if (qty <= limit) {
                              qtyColor = Colors.orange;
                            }

                            return InkWell(
                              onTap: () => _showOpnameDialog(item),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Row(
                                    mainAxisAlignment:
                                        MainAxisAlignment.spaceBetween,
                                    children: [
                                      if (photoUrl != null &&
                                          photoUrl.toString().isNotEmpty) ...[
                                        ClipRRect(
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                          child: CachedNetworkImage(
                                            imageUrl:
                                                photoUrl.toString().startsWith(
                                                  'http',
                                                )
                                                ? photoUrl
                                                : '${Uri.parse(AppConstants.baseUrl).origin}${photoUrl.toString().startsWith('/') ? photoUrl : '/$photoUrl'}',
                                            width: 44,
                                            height: 44,
                                            fit: BoxFit.cover,
                                            placeholder: (context, url) =>
                                                const Center(
                                                  child:
                                                      CircularProgressIndicator(),
                                                ),
                                            errorWidget:
                                                (context, url, error) =>
                                                    const Icon(Icons.error),
                                            memCacheWidth: 800,
                                          ),
                                        ),
                                        const SizedBox(width: 12),
                                      ] else ...[
                                        Container(
                                          width: 44,
                                          height: 44,
                                          decoration: BoxDecoration(
                                            color: Colors.grey.shade200,
                                            borderRadius: BorderRadius.circular(
                                              8,
                                            ),
                                          ),
                                          child: const Icon(
                                            Icons.inventory_2,
                                            color: Colors.grey,
                                            size: 20,
                                          ),
                                        ),
                                        const SizedBox(width: 12),
                                      ],
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            Text(
                                              productName,
                                              style: GoogleFonts.outfit(
                                                fontWeight: FontWeight.w800,
                                                fontSize: 13,
                                                color: const Color(0xFF0F172A),
                                              ),
                                              maxLines: 1,
                                              overflow: TextOverflow.ellipsis,
                                            ),
                                            const SizedBox(height: 4),
                                            Row(
                                              children: [
                                                Text(
                                                  'Limit: $limit',
                                                  style: GoogleFonts.outfit(
                                                    fontSize: 10,
                                                    color: Colors.blueGrey,
                                                    fontWeight: FontWeight.w600,
                                                  ),
                                                ),
                                                if (backorder > 0) ...[
                                                  const SizedBox(width: 8),
                                                  Container(
                                                    padding:
                                                        const EdgeInsets.symmetric(
                                                          horizontal: 6,
                                                          vertical: 2,
                                                        ),
                                                    decoration: BoxDecoration(
                                                      color: Colors.red
                                                          .withOpacity(0.1),
                                                      borderRadius:
                                                          BorderRadius.circular(
                                                            4,
                                                          ),
                                                    ),
                                                    child: Text(
                                                      'Backorder: $backorder',
                                                      style: GoogleFonts.outfit(
                                                        fontSize: 9,
                                                        color: Colors.red,
                                                        fontWeight:
                                                            FontWeight.w800,
                                                      ),
                                                    ),
                                                  ),
                                                ],
                                              ],
                                            ),
                                          ],
                                        ),
                                      ),
                                      Container(
                                        padding: const EdgeInsets.symmetric(
                                          horizontal: 10,
                                          vertical: 6,
                                        ),
                                        decoration: BoxDecoration(
                                          color: qtyColor.withOpacity(0.1),
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                        ),
                                        child: Column(
                                          children: [
                                            Text(
                                              '$qty',
                                              style: GoogleFonts.outfit(
                                                fontWeight: FontWeight.w900,
                                                fontSize: 14,
                                                color: qtyColor,
                                              ),
                                            ),
                                            Text(
                                              'Stok',
                                              style: GoogleFonts.outfit(
                                                fontSize: 8,
                                                color: qtyColor,
                                                fontWeight: FontWeight.w700,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ),
                                  if (batches.isNotEmpty)
                                    Container(
                                      margin: const EdgeInsets.only(top: 12),
                                      padding: const EdgeInsets.all(10),
                                      decoration: BoxDecoration(
                                        color: const Color(0xFFF8FAFC),
                                        borderRadius: BorderRadius.circular(8),
                                        border: Border.all(
                                          color: const Color(0xFFE2E8F0),
                                        ),
                                      ),
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: batches.map<Widget>((b) {
                                          String exp =
                                              (b['expiry_date'] != null &&
                                                  b['expiry_date']
                                                      .toString()
                                                      .isNotEmpty)
                                              ? DateFormat(
                                                  'dd MMM yyyy',
                                                ).format(
                                                  DateTime.parse(
                                                    b['expiry_date'],
                                                  ),
                                                )
                                              : '-';
                                          return Padding(
                                            padding: const EdgeInsets.only(
                                              bottom: 6,
                                            ),
                                            child: Row(
                                              mainAxisAlignment:
                                                  MainAxisAlignment
                                                      .spaceBetween,
                                              children: [
                                                Text(
                                                  'Batch: ${b['batch_no']}',
                                                  style: GoogleFonts.outfit(
                                                    fontSize: 10,
                                                    fontWeight: FontWeight.w600,
                                                    color: Colors.blueGrey,
                                                  ),
                                                ),
                                                Text(
                                                  'Exp: $exp  |  Qty: ${b['quantity']}',
                                                  style: GoogleFonts.outfit(
                                                    fontSize: 10,
                                                    fontWeight: FontWeight.w800,
                                                    color: const Color(
                                                      0xFF0F172A,
                                                    ),
                                                  ),
                                                ),
                                              ],
                                            ),
                                          );
                                        }).toList(),
                                      ),
                                    ),
                                ],
                              ),
                            );
                          },
                        ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryItem(String label, String value, Color valueColor) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          value,
          style: GoogleFonts.outfit(
            color: valueColor,
            fontWeight: FontWeight.w900,
            fontSize: 24,
          ),
        ),
        Text(
          label,
          style: GoogleFonts.outfit(
            color: Colors.blueGrey,
            fontSize: 11,
            fontWeight: FontWeight.w500,
          ),
        ),
      ],
    );
  }

  Widget _buildTaskSummaryCard() {
    int pendingInbound = _pendingInbound.length;
    int pendingOutbound = _pendingOutbound.length + _salesRequests.length;

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 2),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.blue.withOpacity(0.08),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
        border: Border.all(color: Colors.blue.withOpacity(0.1), width: 1.5),
      ),
      child: Stack(
        children: [
          Positioned(
            right: -10,
            top: -10,
            child: Icon(
              Icons.assignment_rounded,
              size: 100,
              color: Colors.blue.withOpacity(0.04),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: Colors.blue.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Icon(
                        Icons.pending_actions_rounded,
                        color: Colors.blue,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Text(
                      'TUGAS PENDING',
                      style: GoogleFonts.outfit(
                        color: Colors.blue.shade900,
                        fontWeight: FontWeight.w800,
                        fontSize: 11,
                        letterSpacing: 1.5,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    _buildSummaryItem(
                      'Inbound',
                      pendingInbound.toString(),
                      Colors.blue.shade800,
                    ),
                    _buildSummaryItem(
                      'Outbound',
                      pendingOutbound.toString(),
                      Colors.orange.shade800,
                    ),
                    _buildSummaryItem(
                      'Total',
                      (pendingInbound + pendingOutbound).toString(),
                      const Color(0xFF0F172A),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                Expanded(
                  child:
                      (_pendingInbound.isEmpty &&
                          _pendingOutbound.isEmpty &&
                          _salesRequests.isEmpty)
                      ? Center(
                          child: Text(
                            'Tidak ada tugas pending',
                            style: GoogleFonts.outfit(
                              color: Colors.blueGrey,
                              fontSize: 12,
                            ),
                          ),
                        )
                      : ListView(
                          physics: const BouncingScrollPhysics(),
                          children: [
                            ..._pendingInbound.map(
                              (it) => _buildMiniTaskItem(
                                title: it['delivery_order_no'] ?? 'NO DO',
                                subtitle:
                                    'Dari: ${it['from_factory']?['name'] ?? 'Pabrik'}',
                                type: 'INBOUND',
                                date:
                                    it['created_at'] ??
                                    DateTime.now().toIso8601String(),
                                color: Colors.blue,
                                onTap: () {
                                  _showInboundModal();
                                },
                              ),
                            ),
                            ..._salesRequests.map(
                              (it) => _buildMiniTaskItem(
                                title: 'Permintaan Cabang',
                                subtitle:
                                    'DO: ${it['delivery_order_no'] ?? '-'}',
                                type: 'REQUEST',
                                date:
                                    it['created_at'] ??
                                    DateTime.now().toIso8601String(),
                                color: Colors.purple,
                                onTap: () {
                                  _showSalesApprovalModal();
                                },
                              ),
                            ),
                            ..._pendingOutbound.map((it) {
                              final driverName =
                                  it['driver']?['first_name'] ?? 'Belum ada';
                              final itemsCount =
                                  (it['items'] as List<dynamic>?)?.length ?? 0;
                              return _buildMiniTaskItem(
                                title:
                                    'Surat Jalan: ${it['delivery_order_no'] ?? '-'}',
                                subtitle:
                                    'Driver: $driverName • $itemsCount Nota',
                                type: 'OUTBOUND',
                                date:
                                    it['created_at'] ??
                                    DateTime.now().toIso8601String(),
                                color: Colors.orange,
                                onTap: () {
                                  _showOutboundOrdersModal();
                                },
                              );
                            }),
                          ],
                        ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMiniTaskItem({
    required String title,
    required String subtitle,
    required String type,
    required String date,
    required Color color,
    required VoidCallback onTap,
  }) {
    final formattedDate = DateFormat(
      'dd MMM, HH:mm',
    ).format(DateTime.parse(date));
    return InkWell(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: Row(
          children: [
            Container(
              width: 36,
              height: 36,
              decoration: BoxDecoration(
                color: color.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(Icons.assignment_rounded, color: color, size: 18),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          title,
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w800,
                            fontSize: 12,
                            color: const Color(0xFF0F172A),
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        formattedDate,
                        style: GoogleFonts.outfit(
                          fontSize: 9,
                          color: Colors.blueGrey,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 4,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: color,
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          type,
                          style: GoogleFonts.outfit(
                            color: Colors.white,
                            fontSize: 8,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ),
                      const SizedBox(width: 6),
                      Expanded(
                        child: Text(
                          subtitle,
                          style: GoogleFonts.outfit(
                            fontSize: 10,
                            color: const Color(0xFF64748B),
                            fontWeight: FontWeight.w500,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
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
    );
  }

  Widget _buildOperationalAccess() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'AKSES OPERASIONAL',
              style: GoogleFonts.outfit(
                fontSize: 11,
                fontWeight: FontWeight.w900,
                color: const Color(0xFF64748B),
                letterSpacing: 1.5,
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        GridView.count(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          crossAxisCount: 3,
          mainAxisSpacing: 12,
          crossAxisSpacing: 12,
          childAspectRatio: 1.1,
          children: [
            _buildActionItem(
              Icons.inventory_2_rounded,
              'Inventory',
              const Color(0xFF0D9488),
              'Cek Stok',
              () => _showInventoryModal(),
            ),
            _buildActionItem(
              Icons.download_for_offline_rounded,
              'Terima',
              const Color(0xFF0891B2),
              'Dari Pabrik',
              () => _showInboundModal(),
            ),
            _buildActionItem(
              Icons.upload_file_rounded,
              'Kirim',
              const Color(0xFF4F46E5),
              'Pesanan Toko',
              () => _showOutboundOrdersModal(),
            ),
            _buildActionItem(
              Icons.fact_check_rounded,
              'Approve',
              const Color(0xFFB45309),
              'Stok Sales',
              () => _showSalesApprovalModal(),
            ),
            _buildActionItem(
              Icons.history_rounded,
              'Riwayat',
              const Color(0xFF6366F1),
              'Mutasi Barang',
              () => _showLogsModal(),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildActionItem(
    IconData icon,
    String label,
    Color color,
    String subLabel,
    VoidCallback onTap,
  ) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: color.withOpacity(0.08),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(24),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        color.withOpacity(0.15),
                        color.withOpacity(0.05),
                      ],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(icon, color: color, size: 24),
                ),
                const SizedBox(height: 10),
                Text(
                  label,
                  textAlign: TextAlign.center,
                  style: GoogleFonts.outfit(
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B),
                  ),
                ),
                Text(
                  subLabel,
                  textAlign: TextAlign.center,
                  style: GoogleFonts.outfit(
                    fontSize: 9,
                    fontWeight: FontWeight.w500,
                    color: const Color(0xFF94A3B8),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildRecentMutations() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'MUTASI TERKINI',
          style: GoogleFonts.outfit(
            fontSize: 11,
            fontWeight: FontWeight.w900,
            color: const Color(0xFF64748B),
            letterSpacing: 1.5,
          ),
        ),
        const SizedBox(height: 16),
        if (_logs.isEmpty)
          Container(
            padding: const EdgeInsets.all(24),
            width: double.infinity,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(20),
              border: Border.all(color: const Color(0xFFE2E8F0)),
            ),
            child: Center(
              child: Text(
                'Belum ada mutasi barang',
                style: GoogleFonts.outfit(color: Colors.blueGrey, fontSize: 13),
              ),
            ),
          )
        else
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _logs.length > 5 ? 5 : _logs.length,
            itemBuilder: (context, index) {
              final log = _logs[index];
              final type = log['type'] ?? 'ADJ';
              final isPositive =
                  log['quantity'] != null && (log['quantity'] as num) > 0;
              final Color typeColor = type == 'IN'
                  ? Colors.green
                  : (type == 'OUT' ? Colors.red : Colors.orange);

              return Container(
                margin: const EdgeInsets.only(bottom: 12),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: const Color(0xFFE2E8F0)),
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: typeColor.withOpacity(0.1),
                        shape: BoxShape.circle,
                      ),
                      child: Icon(
                        type == 'IN'
                            ? Icons.add_box_rounded
                            : (type == 'OUT'
                                  ? Icons.indeterminate_check_box_rounded
                                  : Icons.edit_note_rounded),
                        color: typeColor,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            log['product']?['name'] ?? 'Produk',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800,
                              fontSize: 14,
                            ),
                          ),
                          Text(
                            DateFormat(
                              'dd MMM yyyy, HH:mm',
                            ).format(DateTime.parse(log['created_at'])),
                            style: GoogleFonts.outfit(
                              fontSize: 10,
                              color: Colors.blueGrey,
                            ),
                          ),
                          if (log['processed_by'] != null ||
                              log['approved_by'] != null)
                            Padding(
                              padding: const EdgeInsets.only(top: 2.0),
                              child: Text(
                                'Oleh: ${log['processed_by'] ?? '-'}${log['approved_by'] != null ? ' • Aprv: ${log['approved_by']}' : ''}',
                                style: GoogleFonts.outfit(
                                  fontSize: 9,
                                  color: Colors.blueGrey.shade400,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                        ],
                      ),
                    ),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          '${isPositive ? "+" : ""}${log["quantity"]}',
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w900,
                            color: typeColor,
                            fontSize: 16,
                          ),
                        ),
                        Text(
                          'PCS',
                          style: GoogleFonts.outfit(
                            fontSize: 8,
                            fontWeight: FontWeight.w800,
                            color: Colors.blueGrey,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              );
            },
          ),
      ],
    );
  }

  int _getDistinctProductCount() {
    final ids = _inventory.map((e) => e['product_id']).toSet();
    return ids.length;
  }

  void _showInventoryModal() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.9,
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
        ),
        child: _buildInventoryTab(),
      ),
    );
  }

  void _showInboundModal() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.9,
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
        ),
        child: _buildInboundTab(),
      ),
    );
  }

  void _showOutboundOrdersModal() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.9,
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
        ),
        child: _buildOutboundOrdersTab(),
      ),
    );
  }

  void _showSalesApprovalModal() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.9,
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
        ),
        child: _buildSalesApprovalTab(),
      ),
    );
  }

  Widget _buildInventoryTab() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 10),
          child: TextField(
            controller: _searchController,
            onChanged: _filterInventory,
            decoration: const InputDecoration(
              hintText: 'Cari barang di gudang...',
              prefixIcon: Icon(Icons.search_rounded),
              contentPadding: EdgeInsets.symmetric(vertical: 16),
            ),
          ),
        ),
        Expanded(
          child: RefreshIndicator(
            onRefresh: _fetchAllData,
            color: AppColors.primaryRed,
            child: _filteredInventory.isEmpty
                ? _buildEmptyState(
                    Icons.inventory_2_outlined,
                    'Belum ada data barang',
                  )
                : ListView.builder(
                    padding: const EdgeInsets.all(20),
                    itemCount: _filteredInventory.length,
                    itemBuilder: (context, index) {
                      final item = _filteredInventory[index];
                      final qty = (item['quantity'] ?? 0);
                      final min = (item['min_limit'] ?? 0);
                      final isLow = qty <= min;
                      final isOut = qty <= 0;

                      return _buildInventoryCard(item, isLow, isOut);
                    },
                  ),
          ),
        ),
      ],
    );
  }

  Widget _buildInventoryCard(
    Map<String, dynamic> item,
    bool isLow,
    bool isOut,
  ) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(
          color: isOut
              ? AppColors.error.withOpacity(0.2)
              : (isLow ? Colors.orange.withOpacity(0.2) : AppColors.grayBorder),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(24),
          onTap: () => _showOpnameDialog(item),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: isOut
                        ? AppColors.error.withOpacity(0.1)
                        : (isLow
                              ? Colors.orange.withOpacity(0.1)
                              : AppColors.grayLight),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    isOut
                        ? Icons.block_flipped
                        : (isLow
                              ? Icons.warning_amber_rounded
                              : Icons.inventory_2_rounded),
                    color: isOut
                        ? AppColors.error
                        : (isLow ? Colors.orange : AppColors.textSecondary),
                    size: 20,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        item['product']?['name'] ?? 'Produk Tanpa Nama',
                        style: GoogleFonts.plusJakartaSans(
                          fontWeight: FontWeight.w800,
                          fontSize: 15,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Row(
                        children: [
                          Text(
                            'SKU: ${item['product']?['sku'] ?? '-'}',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 11,
                              color: AppColors.textTertiary,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          if (item['product'] != null &&
                              (item['product']['items_per_unit'] != null ||
                                  item['product']['pcs_per_unit'] != null)) ...[
                            const SizedBox(width: 8),
                            Container(
                              width: 4,
                              height: 4,
                              decoration: const BoxDecoration(
                                color: AppColors.grayBorder,
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              '1 ${item['product']['unit'] ?? 'Unit'} = ${item['product']['items_per_unit'] ?? item['product']['pcs_per_unit']} Pcs',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 10,
                                color: AppColors.textSecondary,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ],
                        ],
                      ),
                    ],
                  ),
                ),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text(
                      _formatStock(item['quantity'], item['product']),
                      style: GoogleFonts.plusJakartaSans(
                        fontWeight: FontWeight.w900,
                        fontSize: 16,
                        color: isOut
                            ? AppColors.error
                            : (isLow ? Colors.orange : AppColors.textPrimary),
                      ),
                    ),
                    Text(
                      'STOK FISIK',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 8,
                        fontWeight: FontWeight.w800,
                        color: AppColors.textTertiary,
                      ),
                    ),
                    if ((item['reserved_quantity'] ?? 0) > 0) ...[
                      const SizedBox(height: 2),
                      if ((item['quantity'] ?? 0) -
                              (item['reserved_quantity'] ?? 0) <
                          0) ...[
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.red.shade50,
                            borderRadius: BorderRadius.circular(4),
                            border: Border.all(color: Colors.red.shade100),
                          ),
                          child: Text(
                            'BACK ORDER',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 7,
                              fontWeight: FontWeight.w900,
                              color: Colors.red.shade700,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ),
                        const SizedBox(height: 1),
                        Text(
                          _formatStock(
                            (item['reserved_quantity'] ?? 0) -
                                (item['quantity'] ?? 0),
                            item['product'],
                          ),
                          style: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w800,
                            fontSize: 11,
                            color: Colors.red.shade700,
                          ),
                        ),
                      ] else ...[
                        Text(
                          _formatStock(
                            (item['quantity'] ?? 0) -
                                (item['reserved_quantity'] ?? 0),
                            item['product'],
                          ),
                          style: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w800,
                            fontSize: 11,
                            color: AppColors.primaryGreenLight,
                          ),
                        ),
                        Text(
                          'TERSEDIA',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 7,
                            fontWeight: FontWeight.w800,
                            color: AppColors.primaryGreenLight.withOpacity(0.7),
                          ),
                        ),
                      ],
                    ],
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void _showOpnameDialog(Map<String, dynamic> item) {
    final int currentQty = item['quantity'] ?? 0;
    final String productId = item['product_id'] ?? item['product']?['id'] ?? '';

    // Ambil batch_no dari batch pertama yang tersedia jika ada
    final List batches = item['batches'] ?? [];
    String batchNo = 'DEFAULT';
    int firstBatchQty = 0;
    if (batches.isNotEmpty) {
      batchNo = batches.first['batch_no'] ?? 'DEFAULT';
      firstBatchQty = (batches.first['quantity'] as num?)?.toInt() ?? 0;
    } else if (item['batch_no'] != null) {
      batchNo = item['batch_no'];
      firstBatchQty = currentQty;
    } else {
      firstBatchQty = currentQty;
    }

    final qtyController = TextEditingController(text: currentQty.toString());
    final reasonController = TextEditingController();

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Padding(
        padding: EdgeInsets.only(
          bottom: MediaQuery.of(context).viewInsets.bottom,
        ),
        child: Container(
          decoration: const BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
          ),
          child: SingleChildScrollView(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(24, 20, 24, 24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Center(
                    child: Container(
                      width: 48,
                      height: 5,
                      decoration: BoxDecoration(
                        color: AppColors.grayBorder,
                        borderRadius: BorderRadius.circular(10),
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: const Color(0xFF059669).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: const Icon(
                          Icons.edit_note_rounded,
                          color: Color(0xFF059669),
                          size: 24,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          'Stock Opname',
                          style: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w800,
                            fontSize: 18,
                            color: const Color(0xFF065F46),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  _buildDetailBox(
                    item['product']?['name'] ?? 'Produk Tidak Diketahui',
                    'Stok Sistem: $currentQty ${item['product']?['unit'] ?? 'Unit'}',
                  ),
                  const SizedBox(height: 20),
                  Text(
                    'Stok Fisik Aktual:',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: qtyController,
                    keyboardType: TextInputType.number,
                    decoration: InputDecoration(
                      hintText: 'Masukkan jumlah fisik...',
                      isDense: true,
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      filled: true,
                      fillColor: AppColors.grayLight,
                    ),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'Alasan / Keterangan:',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: reasonController,
                    maxLines: 2,
                    decoration: InputDecoration(
                      hintText: 'Contoh: Barang rusak, hilang, dsb...',
                      isDense: true,
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      filled: true,
                      fillColor: AppColors.grayLight,
                    ),
                  ),
                  const SizedBox(height: 24),
                  Row(
                    children: [
                      Expanded(
                        child: OutlinedButton(
                          onPressed: () => Navigator.pop(context),
                          style: OutlinedButton.styleFrom(
                            side: const BorderSide(color: AppColors.grayBorder),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16),
                            ),
                            padding: const EdgeInsets.symmetric(vertical: 14),
                          ),
                          child: Text(
                            'BATAL',
                            style: GoogleFonts.plusJakartaSans(
                              fontWeight: FontWeight.w700,
                              color: AppColors.textTertiary,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: ElevatedButton(
                          onPressed: () {
                            int finalTotalQty =
                                int.tryParse(qtyController.text) ?? currentQty;
                            if (finalTotalQty == currentQty &&
                                reasonController.text.isEmpty) {
                              Navigator.pop(context);
                              return;
                            }

                            int diff = finalTotalQty - currentQty;
                            List<Map<String, dynamic>> adjustments = [];

                            if (diff > 0) {
                              adjustments.add({
                                'batch_no': batchNo,
                                'target_qty': firstBatchQty + diff,
                              });
                            } else if (diff < 0) {
                              int remainingDecrease = -diff;
                              for (var b in batches) {
                                final bNo =
                                    b['batch_no']?.toString() ?? 'DEFAULT';
                                final bQty =
                                    (b['quantity'] as num?)?.toInt() ?? 0;

                                if (bQty >= remainingDecrease) {
                                  adjustments.add({
                                    'batch_no': bNo,
                                    'target_qty': bQty - remainingDecrease,
                                  });
                                  remainingDecrease = 0;
                                  break;
                                } else {
                                  adjustments.add({
                                    'batch_no': bNo,
                                    'target_qty': 0,
                                  });
                                  remainingDecrease -= bQty;
                                }
                              }
                            }

                            if (adjustments.isEmpty) {
                              adjustments.add({
                                'batch_no': batchNo,
                                'target_qty': finalTotalQty,
                              });
                            }

                            Navigator.pop(context);
                            _processOpname(
                              adjustments,
                              productId,
                              reasonController.text,
                            );
                          },
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFF059669),
                            foregroundColor: Colors.white,
                            elevation: 0,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16),
                            ),
                            padding: const EdgeInsets.symmetric(vertical: 14),
                          ),
                          child: Text(
                            'SIMPAN',
                            style: GoogleFonts.plusJakartaSans(
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _processOpname(
    List<Map<String, dynamic>> adjustments,
    String productId,
    String reason,
  ) async {
    setState(() => _isLoading = true);
    try {
      for (var adj in adjustments) {
        await apiClient.client.post(
          'warehouse/adjust',
          data: {
            'product_id': productId,
            'quantity': adj['target_qty'],
            'reason': reason,
            'batch_no': adj['batch_no'],
          },
        );
      }
      if (mounted)
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('✓ Stok berhasil disesuaikan!'),
            backgroundColor: Colors.blue,
          ),
        );
      _fetchAllData();
    } catch (e) {
      if (mounted) _showError('Gagal melakukan opname: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Widget _buildInboundTab() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
          child: InkWell(
            onTap: () => _openScanner((v) => _processInboundScan(v)),
            borderRadius: BorderRadius.circular(24),
            child: Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    AppColors.success.withOpacity(0.12),
                    AppColors.success.withOpacity(0.04),
                  ],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(24),
                border: Border.all(
                  color: AppColors.success.withOpacity(0.15),
                  width: 1.5,
                ),
              ),
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: AppColors.success,
                      borderRadius: BorderRadius.circular(16),
                      boxShadow: [
                        BoxShadow(
                          color: AppColors.success.withOpacity(0.3),
                          blurRadius: 12,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: const Icon(
                      Icons.qr_code_scanner_rounded,
                      color: Colors.white,
                      size: 28,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'SCAN SURAT JALAN',
                          style: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w900,
                            fontSize: 15,
                            color: AppColors.textPrimary,
                            letterSpacing: 0.5,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          'Sentuh untuk scan barcode di surat jalan',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 11,
                            color: AppColors.textSecondary,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
        Expanded(
          child: RefreshIndicator(
            onRefresh: _fetchAllData,
            color: AppColors.primaryRed,
            child: _pendingInbound.isEmpty
                ? _buildEmptyState(
                    Icons.download_for_offline_outlined,
                    'Semua kiriman pabrik sudah diterima',
                  )
                : ListView.builder(
                    padding: const EdgeInsets.all(20),
                    itemCount: _pendingInbound.length,
                    itemBuilder: (context, index) {
                      final it = _pendingInbound[index];
                      return _buildTaskCard(
                        title: it['delivery_order_no'] ?? 'NO DO',
                        subtitle:
                            'Dari: ${it['from_factory']?['name'] ?? 'Pabrik'}',
                        type: 'PABRIK',
                        date:
                            it['created_at'] ??
                            DateTime.now().toIso8601String(),
                        color: AppColors.success,
                        onTap: () =>
                            _processInboundScan(it['delivery_order_no'] ?? ''),
                      );
                    },
                  ),
          ),
        ),
      ],
    );
  }

  Widget _buildOutboundOrdersTab() {
    return RefreshIndicator(
      onRefresh: _fetchAllData,
      color: AppColors.primaryRed,
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
            child: InkWell(
              onTap: () => _openScanner((v) => _processOutboundScan(v)),
              borderRadius: BorderRadius.circular(24),
              child: Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      AppColors.info.withOpacity(0.12),
                      AppColors.info.withOpacity(0.04),
                    ],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(24),
                  border: Border.all(
                    color: AppColors.info.withOpacity(0.15),
                    width: 1.5,
                  ),
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: AppColors.info,
                        borderRadius: BorderRadius.circular(16),
                        boxShadow: [
                          BoxShadow(
                            color: AppColors.info.withOpacity(0.3),
                            blurRadius: 12,
                            offset: const Offset(0, 4),
                          ),
                        ],
                      ),
                      child: const Icon(
                        Icons.qr_code_scanner_rounded,
                        color: Colors.white,
                        size: 28,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'SCAN NOTA PESANAN',
                            style: GoogleFonts.plusJakartaSans(
                              fontWeight: FontWeight.w900,
                              fontSize: 15,
                              color: AppColors.textPrimary,
                              letterSpacing: 0.5,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            'Scan barcode nota untuk kirim barang',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 11,
                              color: AppColors.textSecondary,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          Expanded(
            child: _pendingOutbound.isEmpty
                ? _buildEmptyState(
                    Icons.upload_file_rounded,
                    'Tidak ada pengiriman toko pending',
                  )
                : ListView.builder(
                    padding: const EdgeInsets.all(20),
                    itemCount: _pendingOutbound.length,
                    itemBuilder: (context, index) {
                      final it = _pendingOutbound[index];
                      final statusVal = it['status'] ?? 'PENDING';
                      final statusLabel = statusVal == 'SUPERVISOR_APPROVED'
                          ? 'SIAP'
                          : (statusVal == 'PICKING' ? 'PICKING' : statusVal);
                      final driverName =
                          it['driver']?['first_name'] ?? 'Belum ada';
                      final licensePlate =
                          it['vehicle']?['license_plate'] ?? 'No Plat';
                      final itemsCount =
                          (it['items'] as List<dynamic>?)?.length ?? 0;

                      return _buildTaskCard(
                        title: 'SURAT JALAN: ${it['delivery_order_no'] ?? '-'}',
                        subtitle:
                            'Driver: $driverName ($licensePlate) • $itemsCount Nota\nStatus: $statusLabel',
                        type: 'OUTBOUND',
                        date:
                            it['created_at'] ??
                            DateTime.now().toIso8601String(),
                        color: statusVal == 'SUPERVISOR_APPROVED'
                            ? AppColors.success
                            : AppColors.info,
                        onTap: () => _showBatchDetailSheet(it),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildSalesApprovalTab() {
    return RefreshIndicator(
      onRefresh: _fetchAllData,
      color: AppColors.primaryRed,
      child: _salesRequests.isEmpty
          ? _buildEmptyState(
              Icons.fact_check_rounded,
              'Tidak ada permintaan stok sales pending',
            )
          : ListView.builder(
              padding: const EdgeInsets.all(20),
              itemCount: _salesRequests.length,
              itemBuilder: (context, index) {
                final it = _salesRequests[index];
                return _buildTaskCard(
                  title: (it['employee']?['first_name'] ?? 'Sales')
                      .toUpperCase(),
                  subtitle: 'Request: ${it['product']?['name'] ?? 'Produk'}',
                  type: 'STOK SALES',
                  date: it['created_at'] ?? DateTime.now().toIso8601String(),
                  color: Colors.deepOrange,
                  trailing:
                      '${it['quantity'] ?? 0} ${it['unit'] ?? it['product']?['unit'] ?? 'PCS'}',
                  onTap: () => _processSalesRequestApproval(it['id']),
                );
              },
            ),
    );
  }

  Widget _buildTaskCard({
    required String title,
    required String subtitle,
    required String type,
    required String date,
    required Color color,
    String? trailing,
    required VoidCallback onTap,
  }) {
    final formattedDate = DateFormat(
      'dd MMM, HH:mm',
    ).format(DateTime.parse(date));

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: AppColors.grayBorder),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(24),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: color.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(6),
                          ),
                          child: Text(
                            type,
                            style: GoogleFonts.plusJakartaSans(
                              color: color,
                              fontSize: 9,
                              fontWeight: FontWeight.w900,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ),
                        const SizedBox(width: 10),
                        Text(
                          formattedDate,
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 10,
                            color: AppColors.textTertiary,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 10),
                    Text(
                      title,
                      style: GoogleFonts.plusJakartaSans(
                        fontWeight: FontWeight.w800,
                        fontSize: 16,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      subtitle,
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 12,
                        color: AppColors.textSecondary,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
              ),
              if (trailing != null)
                Text(
                  trailing,
                  style: GoogleFonts.plusJakartaSans(
                    fontWeight: FontWeight.w900,
                    fontSize: 15,
                    color: color,
                  ),
                )
              else
                Icon(
                  Icons.arrow_forward_ios_rounded,
                  size: 16,
                  color: AppColors.textTertiary.withOpacity(0.5),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState(IconData icon, String message) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, size: 64, color: AppColors.grayBorder),
          const SizedBox(height: 16),
          Text(
            message,
            style: GoogleFonts.plusJakartaSans(
              color: AppColors.textTertiary,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailBox(String name, String qty) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.grayLight,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Expanded(
            child: Text(
              name,
              style: GoogleFonts.plusJakartaSans(
                fontWeight: FontWeight.w800,
                color: AppColors.textPrimary,
              ),
            ),
          ),
          Text(
            qty,
            style: GoogleFonts.plusJakartaSans(
              fontWeight: FontWeight.w900,
              color: AppColors.info,
              fontSize: 16,
            ),
          ),
        ],
      ),
    );
  }

  void _showScanChoiceMenu() {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.only(
            topLeft: Radius.circular(32),
            topRight: Radius.circular(32),
          ),
        ),
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const SizedBox(height: 12),
              Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: AppColors.grayBorder,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              const SizedBox(height: 24),
              Text(
                'OPERASI GUDANG',
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                  color: AppColors.textTertiary,
                  letterSpacing: 1,
                ),
              ),
              const SizedBox(height: 20),
              _buildScanItem(
                'Terima Surat Jalan Pabrik',
                AppColors.success,
                Icons.download_for_offline_rounded,
                () {
                  Navigator.pop(context);
                  _openScanner((v) => _processInboundScan(v));
                },
              ),
              const SizedBox(height: 12),
              _buildScanItem(
                'Kirim Nota Sales Motoris',
                AppColors.info,
                Icons.upload_file_rounded,
                () {
                  Navigator.pop(context);
                  _openScanner((v) => _processOutboundScan(v));
                },
              ),
              const SizedBox(height: 12),
              _buildScanItem(
                'Approve Stok Sales',
                Colors.orange,
                Icons.fact_check_rounded,
                () {
                  Navigator.pop(context);
                  _openScanner((v) => _processSalesRequestApproval(v));
                },
              ),
              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildScanItem(
    String label,
    Color color,
    IconData icon,
    VoidCallback onTap,
  ) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: color.withOpacity(0.05),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: color.withOpacity(0.1)),
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: color,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: Colors.white, size: 20),
            ),
            const SizedBox(width: 16),
            Text(
              label,
              style: GoogleFonts.plusJakartaSans(
                fontWeight: FontWeight.w800,
                fontSize: 15,
                color: AppColors.textPrimary,
              ),
            ),
            const Spacer(),
            Icon(
              Icons.arrow_forward_ios_rounded,
              size: 14,
              color: color.withOpacity(0.5),
            ),
          ],
        ),
      ),
    );
  }

  void _showLogsModal() {
    _filterLogs(''); // Reset filter before showing
    _logSearchController.clear();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.9,
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
        ),
        child: _buildLogsTab(),
      ),
    );
  }

  Widget _buildLogsTab() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 10),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _logSearchController,
                  onChanged: _filterLogs,
                  decoration: const InputDecoration(
                    hintText: 'Cari mutasi barang...',
                    prefixIcon: Icon(Icons.search_rounded),
                    contentPadding: EdgeInsets.symmetric(vertical: 16),
                  ),
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: StatefulBuilder(
            builder: (context, setModalState) {
              return RefreshIndicator(
                onRefresh: _fetchAllData,
                color: AppColors.primaryRed,
                child: _filteredLogs.isEmpty
                    ? _buildEmptyState(
                        Icons.history_rounded,
                        'Belum ada riwayat mutasi',
                      )
                    : ListView.builder(
                        padding: const EdgeInsets.all(20),
                        itemCount: _filteredLogs.length,
                        itemBuilder: (context, index) {
                          final log = _filteredLogs[index];
                          final type = log['type'] ?? 'ADJ';
                          final isPositive =
                              log['quantity'] != null &&
                              (log['quantity'] as num) > 0;
                          final Color typeColor = type == 'IN'
                              ? Colors.green
                              : (type == 'OUT' ? Colors.red : Colors.orange);

                          return _buildLogItem(
                            log,
                            type,
                            isPositive,
                            typeColor,
                          );
                        },
                      ),
              );
            },
          ),
        ),
      ],
    );
  }

  Widget _buildLogItem(
    dynamic log,
    String type,
    bool isPositive,
    Color typeColor,
  ) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: typeColor.withOpacity(0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(
              type == 'IN'
                  ? Icons.add_box_rounded
                  : (type == 'OUT'
                        ? Icons.indeterminate_check_box_rounded
                        : Icons.edit_note_rounded),
              color: typeColor,
              size: 20,
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  log['product']?['name'] ?? 'Produk',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w800,
                    fontSize: 14,
                  ),
                ),
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        color: AppColors.grayLight,
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Text(
                        log['source'] ?? 'ADJUSTMENT',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 8,
                          fontWeight: FontWeight.w800,
                          color: AppColors.textSecondary,
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      DateFormat(
                        'dd MMM yyyy, HH:mm',
                      ).format(DateTime.parse(log['created_at'])),
                      style: GoogleFonts.outfit(
                        fontSize: 10,
                        color: Colors.blueGrey,
                      ),
                    ),
                  ],
                ),
                if (log['processed_by'] != null || log['approved_by'] != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 2.0),
                    child: Text(
                      'Oleh: ${log['processed_by'] ?? '-'}${log['approved_by'] != null ? ' • Aprv: ${log['approved_by']}' : ''}',
                      style: GoogleFonts.outfit(
                        fontSize: 9,
                        color: Colors.blueGrey.shade400,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                '${isPositive ? "+" : ""}${log["quantity"]}',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  color: typeColor,
                  fontSize: 16,
                ),
              ),
              Text(
                'PCS',
                style: GoogleFonts.outfit(
                  fontSize: 8,
                  fontWeight: FontWeight.w800,
                  color: Colors.blueGrey,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  void _openScanner(Function(String) onResult) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.black,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.8,
        clipBehavior: Clip.antiAlias,
        decoration: const BoxDecoration(
          borderRadius: BorderRadius.only(
            topLeft: Radius.circular(32),
            topRight: Radius.circular(32),
          ),
        ),
        child: Stack(
          children: [
            MobileScanner(
              onDetect: (cap) {
                final bc = cap.barcodes.first.rawValue;
                if (bc != null) {
                  Navigator.pop(context);
                  onResult(bc);
                }
              },
            ),
            Positioned(
              top: 20,
              right: 20,
              child: IconButton(
                icon: const Icon(
                  Icons.close_rounded,
                  color: Colors.white,
                  size: 32,
                ),
                onPressed: () => Navigator.pop(context),
              ),
            ),
            Center(
              child: Container(
                width: 250,
                height: 250,
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.white, width: 2),
                  borderRadius: BorderRadius.circular(32),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
