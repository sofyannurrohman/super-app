import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:hris_app/features/sales/presentation/pages/delivery_checkout_page.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:hris_app/features/sales/presentation/pages/optimal_route_map_page.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/models/visit_plan_model.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/services/offline_delivery_manager.dart';

class DeliveryTrackingPage extends StatefulWidget {
  final bool initialScan;
  final bool autoShowMap;
  final String? deliveryOrderNo;

  const DeliveryTrackingPage({
    super.key,
    this.initialScan = false,
    this.autoShowMap = false,
    this.deliveryOrderNo,
  });

  @override
  State<DeliveryTrackingPage> createState() => _DeliveryTrackingPageState();
}

class _DeliveryTrackingPageState extends State<DeliveryTrackingPage> {
  bool _isScanning = false;
  bool _isLoading = false;
  bool _showTasks = true;
  List<dynamic> _tasks = [];
  Map<String, dynamic>? _batchData;
  final MobileScannerController _scannerController = MobileScannerController();
  final ApiClient apiClient = di.sl<ApiClient>();

  @override
  void initState() {
    super.initState();
    if (widget.deliveryOrderNo != null) {
      _fetchBatch(widget.deliveryOrderNo!);
    } else if (widget.initialScan) {
      _isScanning = true;
      _showTasks = false;
    } else {
      _fetchTasks();
    }
  }

  Future<void> _fetchTasks() async {
    setState(() {
      _isLoading = true;
      _showTasks = true;
      _isScanning = false;
    });

    try {
      final response = await apiClient.client.get('delivery/tasks');

      setState(() {
        _tasks = response.data;
      });
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
            content: Text('Gagal mengambil tugas: $e'),
            backgroundColor: Colors.red),
      );
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _fetchBatch(String doNo) async {
    setState(() {
      _isLoading = true;
      _isScanning = false;
      _showTasks = false;
    });

    try {
      final response = await apiClient.client.get('delivery/batch/$doNo');

      setState(() {
        _batchData = response.data;
      });

      if (widget.autoShowMap && mounted) {
        final items = _batchData!['items'] as List? ?? [];
        final List<VisitPlanModel> plans = items.map((i) {
          final storeData = (i['sales_order'] != null)
              ? i['sales_order']['store']
              : i['sales_transaction']['store'];
          final s = StoreModel.fromJson(storeData);
          return VisitPlanModel(
            storeId: s.id,
            storeName: s.name,
            address: s.address,
            status: 'PLANNED',
            isExtra: false,
            latitude: s.latitude,
            longitude: s.longitude,
          );
        }).toList();
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (_) => OptimalRouteMapPage(plans: plans)),
        );
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
            content: Text('Gagal mengambil data SJ: $e'),
            backgroundColor: Colors.red),
      );
      setState(() => _showTasks = true);
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _updateItemStatus(String itemID, String status) async {
    try {
      await apiClient.client.post(
        'delivery/items/$itemID/confirm',
        data: {'status': status, 'notes': 'Diterima oleh toko'},
      );

      // Refresh data
      _fetchBatch(_batchData!['delivery_order_no']);
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
            content: Text('Gagal update status: $e'),
            backgroundColor: Colors.red),
      );
    }
  }

  Future<void> _syncData() async {
    if (!mounted) return;
    setState(() => _isLoading = true);
    try {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
            content: Text('Sinkronisasi muatan dan tugas...'),
            duration: Duration(seconds: 1)),
      );

      // 1. Sync pending offline tasks
      await OfflineDeliveryManager().syncPendingDeliveries(apiClient);

      // 2. Fetch fresh data depending on state
      if (_batchData != null) {
        await _fetchBatch(_batchData!['delivery_order_no']);
      } else if (_showTasks) {
        await _fetchTasks();
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
              content: Text('Sinkronisasi berhasil!'),
              backgroundColor: Colors.green),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text('Gagal sinkronisasi: $e'),
              backgroundColor: Colors.red),
        );
      }
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: Text(
          _showTasks
              ? 'Tugas Pengiriman'
              : (_isScanning ? 'Scan Surat Jalan' : 'Detail Pengiriman'),
          style: GoogleFonts.outfit(
              fontWeight: FontWeight.w800, color: const Color(0xFF1E293B)),
        ),
        actions: [
          if (!_showTasks)
            IconButton(
              icon:
                  const Icon(Icons.list_alt_rounded, color: Colors.blueAccent),
              onPressed: () => setState(() {
                _showTasks = true;
                _isScanning = false;
                _batchData = null;
              }),
            ),
          if (!_isScanning)
            IconButton(
              icon: const Icon(Icons.qr_code_scanner_rounded,
                  color: Colors.blueAccent),
              onPressed: () => setState(() {
                _isScanning = true;
                _showTasks = false;
              }),
            ),
          IconButton(
            icon: const Icon(Icons.sync_rounded, color: Colors.blueAccent),
            onPressed: _syncData,
            tooltip: 'Sinkronisasi Data',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _showTasks
              ? _buildTaskList()
              : _isScanning
                  ? _buildScanner()
                  : _batchData == null
                      ? _buildEmptyState()
                      : _buildBatchDetail(),
      floatingActionButton: (!_isLoading &&
              !_showTasks &&
              !_isScanning &&
              _batchData != null)
          ? FloatingActionButton.extended(
              onPressed: () {
                final items = _batchData!['items'] as List? ?? [];
                final List<VisitPlanModel> stores = items.map((i) {
                  final storeData = (i['sales_order'] != null)
                      ? i['sales_order']['store']
                      : i['sales_transaction']['store'];
                  final s = StoreModel.fromJson(storeData);
                  return VisitPlanModel(
                    storeId: s.id,
                    storeName: s.name,
                    address: s.address,
                    status:
                        i['status'] == 'DELIVERED' ? 'COMPLETED' : 'PLANNED',
                    isExtra: false,
                    latitude: s.latitude,
                    longitude: s.longitude,
                  );
                }).toList();
                Navigator.push(
                    context,
                    MaterialPageRoute(
                        builder: (_) => OptimalRouteMapPage(plans: stores)));
              },
              backgroundColor: const Color(0xFF1E293B),
              icon: const Icon(Icons.map_rounded, color: Colors.white),
              label: Text('Rute Optimal',
                  style: GoogleFonts.outfit(
                      color: Colors.white, fontWeight: FontWeight.bold)),
            )
          : null,
    );
  }

  Widget _buildTaskList() {
    if (_tasks.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.assignment_turned_in_outlined,
                size: 80, color: Color(0xFFE2E8F0)),
            const SizedBox(height: 16),
            Text(
              'Tidak ada tugas pengiriman',
              style: GoogleFonts.outfit(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: const Color(0xFF94A3B8)),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: _fetchTasks,
              icon: const Icon(Icons.refresh),
              label: const Text('Refresh'),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchTasks,
      child: ListView.builder(
        padding: const EdgeInsets.all(20),
        itemCount: _tasks.length,
        itemBuilder: (context, index) {
          final task = _tasks[index];
          final status = task['status'];
          final doNo = task['delivery_order_no'];
          final List itemsList = task['items'] as List? ?? [];
          final itemCount = itemsList.length;

          return Card(
            margin: const EdgeInsets.only(bottom: 16),
            shape:
                RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
            elevation: 2,
            shadowColor: Colors.black12,
            child: InkWell(
              onTap: () => _fetchBatch(doNo),
              borderRadius: BorderRadius.circular(20),
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 10, vertical: 4),
                          decoration: BoxDecoration(
                            color: Colors.blueAccent.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            doNo,
                            style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w900,
                                color: Colors.blueAccent,
                                fontSize: 12),
                          ),
                        ),
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 10, vertical: 4),
                          decoration: BoxDecoration(
                            color: _getStatusColor(status).withOpacity(0.1),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            status,
                            style: GoogleFonts.outfit(
                                fontWeight: FontWeight.bold,
                                color: _getStatusColor(status),
                                fontSize: 10),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'Tugas Pengiriman #$doNo',
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w800,
                          fontSize: 16,
                          color: const Color(0xFF1E293B)),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        const Icon(Icons.inventory_2_outlined,
                            size: 14, color: Color(0xFF64748B)),
                        const SizedBox(width: 6),
                        Text(
                          '$itemCount Toko Tujuan',
                          style: GoogleFonts.outfit(
                              fontSize: 12, color: const Color(0xFF64748B)),
                        ),
                        const SizedBox(width: 16),
                        const Icon(Icons.calendar_today_outlined,
                            size: 14, color: Color(0xFF64748B)),
                        const SizedBox(width: 6),
                        Text(
                          'Hari Ini',
                          style: GoogleFonts.outfit(
                              fontSize: 12, color: const Color(0xFF64748B)),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'WAITING':
        return Colors.orange;
      case 'PICKING':
        return const Color(0xFFF59E0B); // amber — sedang disiapkan gudang
      case 'IN_TRANSIT':
        return Colors.blue;
      case 'ON_DELIVERY':
        return Colors.blue;
      case 'DELIVERED':
        return Colors.green;
      case 'COMPLETED':
        return Colors.green;
      default:
        return Colors.grey;
    }
  }

  /// Dipanggil driver saat SJ berstatus PICKING — konfirmasi barang sudah diambil
  /// lalu stok gudang langsung dipotong dan status berubah ke ON_DELIVERY.
  Future<void> _confirmPicking() async {
    if (_batchData == null) return;
    final doNo = _batchData!['delivery_order_no'] as String?;
    if (doNo == null) return;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Row(
          children: [
            const Icon(Icons.inventory_2_rounded, color: Color(0xFFF59E0B)),
            const SizedBox(width: 8),
            Text('Konfirmasi Pengambilan',
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w800, fontSize: 16)),
          ],
        ),
        content: Text(
          'Pastikan semua barang dari Surat Jalan $doNo sudah ada di truk.\n\nStok gudang akan otomatis dipotong sesuai pesanan.',
          style:
              GoogleFonts.outfit(fontSize: 13, color: const Color(0xFF475569)),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text('BATAL',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFFF59E0B),
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12)),
            ),
            onPressed: () => Navigator.pop(ctx, true),
            child: Text('YA, SUDAH DIAMBIL',
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w800, color: Colors.white)),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    setState(() => _isLoading = true);
    try {
      await apiClient.client.post('delivery/batch/$doNo/start');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Pengiriman $doNo dimulai! Stok gudang sudah dipotong.',
              style: GoogleFonts.outfit()),
          backgroundColor: Colors.green,
          behavior: SnackBarBehavior.floating,
        ),
      );
      _fetchBatch(doNo);
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
            content: Text('Gagal konfirmasi: $e'), backgroundColor: Colors.red),
      );
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Widget _buildScanner() {
    return Stack(
      children: [
        MobileScanner(
          controller: _scannerController,
          onDetect: (capture) {
            if (_isLoading) return; // Use existing _isLoading as a lock
            final List<Barcode> barcodes = capture.barcodes;
            for (final barcode in barcodes) {
              final String? code = barcode.rawValue;
              if (code != null) {
                if (code.startsWith('SJ-')) {
                  _fetchBatch(code);
                } else {
                  _openCheckoutForCode(code);
                }
                break;
              }
            }
          },
        ),
        Center(
          child: Container(
            width: 250,
            height: 250,
            decoration: BoxDecoration(
              border: Border.all(color: Colors.blueAccent, width: 4),
              borderRadius: BorderRadius.circular(24),
            ),
          ),
        ),
        Positioned(
          bottom: 100,
          left: 0,
          right: 0,
          child: Center(
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
              decoration: BoxDecoration(
                color: Colors.black54,
                borderRadius: BorderRadius.circular(30),
              ),
              child: Text(
                'Arahkan kamera ke barcode SJ atau Nota',
                style: GoogleFonts.outfit(color: Colors.white, fontSize: 14),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Future<void> _openCheckoutForCode(String code) async {
    setState(() => _isLoading = true);

    // If we have a batch open, try to find the item in the batch
    if (_batchData != null) {
      final items = _batchData!['items'] as List? ?? [];
      final matchedItem = items.cast<Map<String, dynamic>>().firstWhere(
        (item) {
          final soNumber = item['sales_order']?['so_number'];
          final invoiceNo = item['sales_transaction']?['invoice_no'];
          final itemId = item['id'];
          return soNumber == code || invoiceNo == code || itemId == code;
        },
        orElse: () => <String, dynamic>{},
      );

      if (matchedItem.isNotEmpty) {
        setState(() {
          _isLoading = false;
          _isScanning = false;
        });

        final result = await Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => DeliveryCheckoutPage(
              deliveryItem: matchedItem,
              batchData: _batchData,
              onSuccess: () => _fetchBatch(_batchData!['delivery_order_no']),
            ),
          ),
        );
        if (result == true) {
          _fetchBatch(_batchData!['delivery_order_no']);
        }
        return;
      }
    }

    setState(() => _isLoading = false);
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text(
            'Data Nota tidak ditemukan di Surat Jalan ini. Silakan scan Surat Jalan utamanya terlebih dahulu.'),
        backgroundColor: Colors.red,
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  void _showUpdateCashDialog(double currentCash) {
    final TextEditingController controller =
        TextEditingController(text: currentCash.toInt().toString());
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Update Uang Dibawa',
            style: GoogleFonts.outfit(fontWeight: FontWeight.bold)),
        content: TextField(
          controller: controller,
          keyboardType: TextInputType.number,
          decoration: const InputDecoration(
            labelText: 'Jumlah Uang (Rp)',
            prefixText: 'Rp ',
            border: OutlineInputBorder(),
          ),
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('BATAL')),
          ElevatedButton(
            onPressed: () async {
              final amount = double.tryParse(controller.text) ?? 0.0;
              Navigator.pop(context);
              _updateCash(amount);
            },
            child: const Text('SIMPAN'),
          ),
        ],
      ),
    );
  }

  Future<void> _updateCash(double amount) async {
    if (_batchData == null) return;
    setState(() => _isLoading = true);
    try {
      await apiClient.client.post(
        'delivery/batch/${_batchData!['id']}/cash',
        data: {'amount': amount},
      );
      _fetchBatch(_batchData!['delivery_order_no']);
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
            content: Text('Gagal update uang: $e'),
            backgroundColor: Colors.red),
      );
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.local_shipping_outlined,
              size: 80, color: Color(0xFFE2E8F0)),
          const SizedBox(height: 16),
          Text(
            'Belum ada data pengiriman',
            style: GoogleFonts.outfit(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: const Color(0xFF94A3B8)),
          ),
        ],
      ),
    );
  }

  Widget _buildBatchDetail() {
    final items = _batchData!['items'] as List? ?? [];
    final List<VisitPlanModel> stores = items.map((i) {
      final storeData = (i['sales_order'] != null)
          ? i['sales_order']['store']
          : i['sales_transaction']['store'];
      final s = StoreModel.fromJson(storeData);
      final isDelivered = i['status'] == 'DELIVERED';
      return VisitPlanModel(
        storeId: s.id,
        storeName: s.name,
        address: s.address,
        status: isDelivered ? 'COMPLETED' : 'PLANNED',
        isExtra: false,
        latitude: s.latitude,
        longitude: s.longitude,
      );
    }).toList();
    final deliveredCount =
        items.where((i) => i['status'] == 'DELIVERED').length;
    final progress = items.isEmpty ? 0.0 : deliveredCount / items.length;
    final totalCash = (_batchData!['total_cash_collected'] ?? 0.0).toDouble();
    final totalTransfer =
        (_batchData!['total_transfer_collected'] ?? 0.0).toDouble();
    final totalTempo = (_batchData!['total_tempo_collected'] ?? 0.0).toDouble();

    int totalReturnQty = 0;
    for (var item in items) {
      if (item['has_return'] == true) {
        totalReturnQty += (item['return_quantity'] ?? 0) as int;
      }
    }

    return RefreshIndicator(
      onRefresh: () => _fetchBatch(_batchData!['delivery_order_no']),
      child: ListView(
        physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics()),
        children: [
          // Header Info
          Container(
            padding: const EdgeInsets.all(20),
            color: Colors.white,
            child: Column(
              children: [
                // Progress Section
                Row(
                  children: [
                    Expanded(
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(10),
                        child: LinearProgressIndicator(
                          value: progress,
                          minHeight: 10,
                          backgroundColor: const Color(0xFFF1F5F9),
                          color: Colors.green,
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Text(
                      '${(progress * 100).toInt()}%',
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w900,
                          color: Colors.green,
                          fontSize: 14),
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('NOMOR SURAT JALAN',
                            style: GoogleFonts.outfit(
                                fontSize: 10,
                                color: const Color(0xFF94A3B8),
                                fontWeight: FontWeight.w800,
                                letterSpacing: 1)),
                        Text(_batchData!['delivery_order_no'],
                            style: GoogleFonts.outfit(
                                fontSize: 18,
                                fontWeight: FontWeight.w900,
                                color: const Color(0xFF1E293B))),
                      ],
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 12, vertical: 6),
                      decoration: BoxDecoration(
                        color: Colors.blueAccent.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Text(
                        _batchData!['status'],
                        style: GoogleFonts.outfit(
                            color: Colors.blueAccent,
                            fontWeight: FontWeight.bold,
                            fontSize: 12),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),

                // Armada Info
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF1F5F9),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.people_alt_rounded,
                          color: Colors.blueGrey, size: 20),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'TIM PENGIRIMAN',
                              style: GoogleFonts.outfit(
                                  fontSize: 9,
                                  color: Colors.blueGrey,
                                  fontWeight: FontWeight.w800,
                                  letterSpacing: 0.5),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              (_batchData!['driver'] != null
                                      ? _batchData!['driver']['first_name']
                                      : 'Tanpa Supir') +
                                  (_batchData!['helper'] != null
                                      ? ' & ${_batchData!['helper']['first_name']} (Helper)'
                                      : ''),
                              style: GoogleFonts.outfit(
                                  fontSize: 13,
                                  fontWeight: FontWeight.w700,
                                  color: const Color(0xFF1E293B)),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
                // Collection Stats Grid
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF8FAFC),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: const Color(0xFFE2E8F0)),
                  ),
                  child: Column(
                    children: [
                      Row(
                        children: [
                          Expanded(
                              child: _buildCollectionMiniCard(
                                  Icons.payments_rounded,
                                  Colors.green,
                                  'TUNAI',
                                  totalCash,
                                  onEdit: () =>
                                      _showUpdateCashDialog(totalCash))),
                          const SizedBox(width: 8),
                          Expanded(
                              child: _buildCollectionMiniCard(
                                  Icons.account_balance_wallet_rounded,
                                  Colors.blueAccent,
                                  'TRANSFER',
                                  totalTransfer)),
                        ],
                      ),
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          Expanded(
                              child: _buildCollectionMiniCard(
                                  Icons.calendar_month_rounded,
                                  Colors.orange,
                                  'TEMPO',
                                  totalTempo)),
                          const SizedBox(width: 8),
                          if (totalReturnQty > 0)
                            Expanded(
                              child: Container(
                                padding: const EdgeInsets.all(10),
                                decoration: BoxDecoration(
                                    color: Colors.white,
                                    borderRadius: BorderRadius.circular(12),
                                    border: Border.all(
                                        color: Colors.red.withOpacity(0.3))),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      children: [
                                        const Icon(
                                            Icons.assignment_return_rounded,
                                            color: Colors.red,
                                            size: 16),
                                        const SizedBox(width: 4),
                                        Text('RETUR',
                                            style: GoogleFonts.outfit(
                                                fontSize: 9,
                                                color: Colors.blueGrey,
                                                fontWeight: FontWeight.w800)),
                                      ],
                                    ),
                                    const SizedBox(height: 4),
                                    Text('$totalReturnQty Unit',
                                        style: GoogleFonts.outfit(
                                            fontSize: 14,
                                            fontWeight: FontWeight.w900,
                                            color: Colors.red)),
                                  ],
                                ),
                              ),
                            )
                          else
                            const Spacer(),
                        ],
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
                _buildManifestSummary(items),
                const SizedBox(height: 16),
                // Aksi PICKING: driver konfirmasi sudah ambil barang
                if ((_batchData!['status'] ?? '') == 'PICKING')
                  Container(
                    width: double.infinity,
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [Color(0xFFF59E0B), Color(0xFFD97706)],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                      borderRadius: BorderRadius.circular(20),
                      boxShadow: [
                        BoxShadow(
                            color: const Color(0xFFF59E0B).withOpacity(0.4),
                            blurRadius: 12,
                            offset: const Offset(0, 6)),
                      ],
                    ),
                    child: Material(
                      color: Colors.transparent,
                      borderRadius: BorderRadius.circular(20),
                      child: InkWell(
                        borderRadius: BorderRadius.circular(20),
                        onTap: _confirmPicking,
                        child: Padding(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 20, vertical: 18),
                          child: Row(
                            children: [
                              Container(
                                padding: const EdgeInsets.all(10),
                                decoration: BoxDecoration(
                                    color: Colors.white.withOpacity(0.2),
                                    shape: BoxShape.circle),
                                child: const Icon(Icons.inventory_2_rounded,
                                    color: Colors.white, size: 26),
                              ),
                              const SizedBox(width: 16),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'KONFIRMASI AMBIL BARANG',
                                      style: GoogleFonts.outfit(
                                          color: Colors.white,
                                          fontWeight: FontWeight.w900,
                                          fontSize: 13,
                                          letterSpacing: 0.5),
                                    ),
                                    Text(
                                      'Semua barang sudah ada di truk? Tekan untuk mulai pengiriman & potong stok gudang.',
                                      style: GoogleFonts.outfit(
                                          color: Colors.white.withOpacity(0.85),
                                          fontSize: 10,
                                          fontWeight: FontWeight.w600),
                                    ),
                                  ],
                                ),
                              ),
                              const Icon(Icons.arrow_forward_ios_rounded,
                                  color: Colors.white, size: 16),
                            ],
                          ),
                        ),
                      ),
                    ),
                  )
                else if ((_batchData!['status'] ?? '') == 'PENDING' ||
                    (_batchData!['status'] ?? '') == 'SUPERVISOR_APPROVED' ||
                    (_batchData!['status'] ?? '') == 'WAITING_ASSIGNMENT')
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(
                        horizontal: 20, vertical: 16),
                    decoration: BoxDecoration(
                      color: Colors.blueGrey.withOpacity(0.05),
                      borderRadius: BorderRadius.circular(20),
                      border:
                          Border.all(color: Colors.blueGrey.withOpacity(0.2)),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.hourglass_empty_rounded,
                            color: Colors.blueGrey, size: 24),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                'MENUNGGU GUDANG',
                                style: GoogleFonts.outfit(
                                    color: Colors.blueGrey,
                                    fontWeight: FontWeight.w900,
                                    fontSize: 12,
                                    letterSpacing: 0.5),
                              ),
                              Text(
                                'Barang sedang disiapkan oleh tim gudang. Tombol konfirmasi akan muncul setelah gudang selesai menyiapkannya.',
                                style: GoogleFonts.outfit(
                                    color: Colors.blueGrey,
                                    fontSize: 11,
                                    fontWeight: FontWeight.w500),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),

          // Section Title
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 16, 20, 8),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'DAFTAR TOKO TUJUAN',
                  style: GoogleFonts.outfit(
                      fontSize: 11,
                      fontWeight: FontWeight.w900,
                      color: Colors.blueGrey,
                      letterSpacing: 1.5),
                ),
                Text(
                  '${items.length} Outlet',
                  style: GoogleFonts.outfit(
                      fontSize: 11,
                      fontWeight: FontWeight.w800,
                      color: Colors.blueAccent),
                ),
              ],
            ),
          ),

          // List of Stores
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            child: Column(
              children: List.generate(items.length, (index) {
                final item = items[index];
                final store = (item['sales_order'] != null)
                    ? item['sales_order']['store']
                    : item['sales_transaction']['store'];
                final isDelivered = item['status'] == 'DELIVERED';
                final lat = store['latitude'];
                final lng = store['longitude'];

                return Container(
                  margin: const EdgeInsets.only(bottom: 16),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(20),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.05),
                        blurRadius: 10,
                        offset: const Offset(0, 4),
                      ),
                    ],
                    border: Border.all(
                        color: isDelivered
                            ? Colors.green.withOpacity(0.2)
                            : Colors.transparent),
                  ),
                  child: Material(
                    color: Colors.transparent,
                    borderRadius: BorderRadius.circular(20),
                    child: InkWell(
                      borderRadius: BorderRadius.circular(20),
                      onTap: () async {
                        final result = await Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (_) => DeliveryCheckoutPage(
                              deliveryItem: item,
                              batchData: _batchData,
                              onSuccess: () =>
                                  _fetchBatch(_batchData!['delivery_order_no']),
                            ),
                          ),
                        );
                        if (result == true) {
                          _fetchBatch(_batchData!['delivery_order_no']);
                        }
                      },
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Row(
                          children: [
                            Container(
                              width: 40,
                              height: 40,
                              decoration: BoxDecoration(
                                color: isDelivered
                                    ? Colors.green.withOpacity(0.1)
                                    : const Color(0xFFF1F5F9),
                                shape: BoxShape.circle,
                              ),
                              child: Center(
                                child: Text(
                                  '${index + 1}',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w900,
                                    color: isDelivered
                                        ? Colors.green
                                        : const Color(0xFF64748B),
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(store['name'] ?? 'Toko Tanpa Nama',
                                      style: GoogleFonts.outfit(
                                          fontWeight: FontWeight.w800,
                                          fontSize: 16)),
                                  Text(
                                      store['address'] ??
                                          'Alamat tidak tersedia',
                                      style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          color: const Color(0xFF64748B)),
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis),
                                  if (item['has_return'] == true) ...[
                                    const SizedBox(height: 4),
                                    Container(
                                      width: double.infinity,
                                      padding: const EdgeInsets.symmetric(
                                          horizontal: 8, vertical: 6),
                                      decoration: BoxDecoration(
                                        color: Colors.red.withOpacity(0.1),
                                        borderRadius: BorderRadius.circular(6),
                                      ),
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            'Barang Diretur (${item['return_quantity'] ?? '-'} unit):',
                                            style: GoogleFonts.outfit(
                                                fontSize: 10,
                                                color: Colors.red,
                                                fontWeight: FontWeight.bold),
                                          ),
                                          Builder(builder: (context) {
                                            List<Widget> returnWidgets = [];
                                            try {
                                              if (item['sales_order'] != null &&
                                                  item['sales_order']
                                                          ['returns'] !=
                                                      null) {
                                                for (var ret
                                                    in (item['sales_order']
                                                        ['returns'] as List)) {
                                                  if (ret['items'] != null) {
                                                    for (var retItem
                                                        in (ret['items']
                                                            as List)) {
                                                      final prodName =
                                                          retItem['product']
                                                                  ?['name'] ??
                                                              'Item';
                                                      final qty =
                                                          retItem['quantity'] ??
                                                              0;
                                                      returnWidgets.add(Padding(
                                                        padding:
                                                            const EdgeInsets
                                                                .only(top: 2),
                                                        child: Text(
                                                          '• $prodName ($qty PCS)',
                                                          style: GoogleFonts
                                                              .outfit(
                                                                  fontSize: 10,
                                                                  color: Colors
                                                                      .red[800],
                                                                  fontWeight:
                                                                      FontWeight
                                                                          .w600),
                                                        ),
                                                      ));
                                                    }
                                                  }
                                                }
                                              }
                                            } catch (e) {
                                              // Ignore parse errors safely
                                            }
                                            return Column(
                                              crossAxisAlignment:
                                                  CrossAxisAlignment.start,
                                              children: returnWidgets,
                                            );
                                          }),
                                          if (item['return_reason'] != null &&
                                              item['return_reason']
                                                  .toString()
                                                  .isNotEmpty)
                                            Padding(
                                              padding:
                                                  const EdgeInsets.only(top: 4),
                                              child: Text(
                                                  'Alasan: ${item['return_reason']}',
                                                  style: GoogleFonts.outfit(
                                                      fontSize: 9,
                                                      color: Colors.red[900],
                                                      fontStyle:
                                                          FontStyle.italic)),
                                            )
                                        ],
                                      ),
                                    ),
                                  ],
                                ],
                              ),
                            ),
                            if (lat != null && lng != null)
                              IconButton(
                                icon: const Icon(Icons.near_me_rounded,
                                    color: Colors.blueAccent, size: 24),
                                onPressed: () async {
                                  final url =
                                      'https://www.google.com/maps/search/?api=1&query=$lat,$lng';
                                  if (await canLaunchUrl(Uri.parse(url))) {
                                    await launchUrl(Uri.parse(url),
                                        mode: LaunchMode.externalApplication);
                                  }
                                },
                                tooltip: 'Navigasi ke Toko',
                              ),
                            if (!isDelivered)
                              const Icon(Icons.arrow_forward_ios_rounded,
                                  color: Color(0xFFCBD5E1), size: 16)
                            else
                              const Icon(Icons.check_circle_rounded,
                                  color: Colors.green, size: 32),
                          ],
                        ),
                      ),
                    ),
                  ),
                );
              }),
            ),
          ),
          const SizedBox(height: 40),
        ],
      ),
    );
  }

  Widget _buildCollectionMiniCard(
      IconData icon, Color color, String label, double amount,
      {VoidCallback? onEdit}) {
    final formatter =
        NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Icon(icon, color: color, size: 16),
                  const SizedBox(width: 4),
                  Text(label,
                      style: GoogleFonts.outfit(
                          fontSize: 9,
                          color: const Color(0xFF64748B),
                          fontWeight: FontWeight.w800)),
                ],
              ),
              if (onEdit != null)
                InkWell(
                    onTap: onEdit,
                    child: const Icon(Icons.edit_note_rounded,
                        color: Colors.blueAccent, size: 16)),
            ],
          ),
          const SizedBox(height: 4),
          FittedBox(
            fit: BoxFit.scaleDown,
            child: Text(
              formatter.format(amount),
              style: GoogleFonts.outfit(
                  fontSize: 14,
                  fontWeight: FontWeight.w900,
                  color: const Color(0xFF1E293B)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildManifestSummary(List<dynamic> items) {
    Map<String, Map<String, dynamic>> manifest = {};
    int totalInitial = 0;
    int totalOnTruck = 0;

    for (var item in items) {
      final soItems = item['sales_order']?['items'] ?? [];
      final bool isDelivered = item['status'] == 'DELIVERED';

      for (var soItem in soItems) {
        final productName = soItem['product']?['name'] ?? 'Produk';
        final ordered = (soItem['ordered_quantity'] ?? 0) as int;
        final actual = (soItem['actual_quantity'] ?? ordered) as int;
        final ret = isDelivered ? (ordered - actual) : 0;

        // Stock on Truck calculation:
        // If pending: the whole 'ordered' quantity is still on truck.
        // If delivered: only the 'ret' (returned) quantity is back on truck.
        final currentOnTruck = isDelivered ? ret : ordered;

        if (!manifest.containsKey(productName)) {
          manifest[productName] = {
            'initial': 0,
            'onTruck': 0,
            'returned': 0,
            'uom': soItem['product']?['uom'] ?? 'Unit'
          };
        }
        manifest[productName]!['initial'] =
            (manifest[productName]!['initial'] as int) + ordered;
        manifest[productName]!['onTruck'] =
            (manifest[productName]!['onTruck'] as int) + currentOnTruck;
        manifest[productName]!['returned'] =
            (manifest[productName]!['returned'] as int) + ret;

        totalInitial += ordered;
        totalOnTruck += currentOnTruck;
      }
    }

    // Include bonus items (hadiah)
    for (var item in items) {
      final soBonusItems = item['sales_order']?['bonus_items'] ?? [];
      final bool isDelivered = item['status'] == 'DELIVERED';

      for (var bonusItem in soBonusItems) {
        if (bonusItem['bonus_inventory'] != null) {
          final productName =
              '${bonusItem['bonus_inventory']['name'] ?? 'Bonus'} (Hadiah)';
          final qty = (bonusItem['quantity'] ?? 0) as int;
          final unit = bonusItem['bonus_inventory']['unit'] ?? 'PCS';

          // Jika status delivered, barang hadiah sudah diserahkan (tidak di truk lagi).
          final currentOnTruck = isDelivered ? 0 : qty;

          if (!manifest.containsKey(productName)) {
            manifest[productName] = {
              'initial': 0,
              'onTruck': 0,
              'returned': 0,
              'uom': unit,
            };
          }
          manifest[productName]!['initial'] =
              (manifest[productName]!['initial'] as int) + qty;
          manifest[productName]!['onTruck'] =
              (manifest[productName]!['onTruck'] as int) + currentOnTruck;

          totalInitial += qty;
          totalOnTruck += currentOnTruck;
        }
      }
    }

    // Include spares (barang cadangan) from batch data
    final List spares = _batchData?['spares'] as List? ?? [];
    for (var spare in spares) {
      if (spare['product'] != null) {
        final productName =
            '${spare['product']['name'] ?? 'Produk'} (Cadangan)';
        final qty = (spare['quantity'] ?? 0) as int;
        final unit = spare['product']['unit'] ?? 'PCS';

        if (!manifest.containsKey(productName)) {
          manifest[productName] = {
            'initial': 0,
            'onTruck': 0,
            'returned': 0,
            'uom': unit,
          };
        }
        manifest[productName]!['initial'] += qty;
        manifest[productName]!['onTruck'] += qty;

        totalInitial += qty;
        totalOnTruck += qty;
      }
    }

    // Process TUKAR_PRODUK returns to adjust the manifest
    for (var item in items) {
      final isDelivered = item['status'] == 'DELIVERED';
      if (!isDelivered) continue;

      final so = item['sales_order'];
      if (so != null && so['returns'] != null) {
        for (var ret in (so['returns'] as List)) {
          if (ret['return_type'] == 'TUKAR_PRODUK' && ret['items'] != null) {
            for (var retItem in (ret['items'] as List)) {
              final productName = retItem['product']?['name'] ?? 'Produk';
              final retQty = (retItem['quantity'] ?? 0) as int;

              // 1. Add to Damaged stock
              final damagedName = '$productName (Rusak)';
              if (!manifest.containsKey(damagedName)) {
                manifest[damagedName] = {
                  'initial': 0,
                  'onTruck': 0,
                  'returned': 0,
                  'uom': retItem['product']?['unit'] ?? 'Unit'
                };
              }
              manifest[damagedName]!['onTruck'] =
                  (manifest[damagedName]!['onTruck'] as int) + retQty;
              totalOnTruck += retQty;

              // 2. Decrease Replacement stock
              final repName =
                  retItem['replacement_product']?['name'] ?? productName;
              final repQty = (retItem['replacement_quantity'] ?? retQty) as int;

              // Try to deduct from normal stock first
              if (manifest.containsKey(repName)) {
                int current = manifest[repName]!['onTruck'] as int;
                if (current >= repQty) {
                  manifest[repName]!['onTruck'] = current - repQty;
                  totalOnTruck -= repQty;
                } else {
                  manifest[repName]!['onTruck'] = 0;
                  totalOnTruck -= current;
                  int remaining = repQty - current;

                  final cadanganName = '$repName (Cadangan)';
                  if (manifest.containsKey(cadanganName)) {
                    int cadangan = manifest[cadanganName]!['onTruck'] as int;
                    int deduct = cadangan >= remaining ? remaining : cadangan;
                    manifest[cadanganName]!['onTruck'] = cadangan - deduct;
                    totalOnTruck -= deduct;
                  }
                }
              } else {
                final cadanganName = '$repName (Cadangan)';
                if (manifest.containsKey(cadanganName)) {
                  int cadangan = manifest[cadanganName]!['onTruck'] as int;
                  int deduct = cadangan >= repQty ? repQty : cadangan;
                  manifest[cadanganName]!['onTruck'] = cadangan - deduct;
                  totalOnTruck -= deduct;
                }
              }
            }
          }
        }
      }
    }

    if (manifest.isEmpty) return const SizedBox.shrink();

    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.blueAccent.withOpacity(0.05),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.blueAccent.withOpacity(0.1)),
      ),
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          title: Text(
            'GUDANG BERJALAN',
            style: GoogleFonts.outfit(
                fontSize: 11,
                fontWeight: FontWeight.w900,
                color: Colors.blueAccent,
                letterSpacing: 1),
          ),
          subtitle: Text(
            'Sisa di Truk: $totalOnTruck item (Awal: $totalInitial)',
            style: GoogleFonts.outfit(
                fontSize: 10,
                fontWeight: FontWeight.w700,
                color: Colors.blueGrey),
          ),
          leading: Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
                color: Colors.blueAccent.withOpacity(0.1),
                shape: BoxShape.circle),
            child: const Icon(Icons.local_shipping_rounded,
                color: Colors.blueAccent, size: 18),
          ),
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Column(
                children: [
                  const Divider(height: 1),
                  ...manifest.entries.map((entry) {
                    final data = entry.value;
                    final hasRetur = data['returned'] > 0;
                    return Padding(
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      child: Row(
                        children: [
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(entry.key,
                                    style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w700,
                                        fontSize: 13,
                                        color: const Color(0xFF1E293B))),
                                Text(
                                  'Awal: ${data['initial']} ${data['uom']} ${hasRetur ? "• Retur: ${data['returned']}" : ""}',
                                  style: GoogleFonts.outfit(
                                      fontSize: 9,
                                      color: Colors.blueGrey,
                                      fontWeight: FontWeight.w600),
                                ),
                              ],
                            ),
                          ),
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Text(
                                '${data['onTruck']} ${data['uom']}',
                                style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w900,
                                    fontSize: 15,
                                    color: data['onTruck'] > 0
                                        ? Colors.blueAccent
                                        : Colors.grey),
                              ),
                              Text('DI TRUK',
                                  style: GoogleFonts.outfit(
                                      fontSize: 9,
                                      color: Colors.blueGrey,
                                      fontWeight: FontWeight.w800,
                                      letterSpacing: 0.5)),
                            ],
                          ),
                        ],
                      ),
                    );
                  }),
                  const SizedBox(height: 8),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
