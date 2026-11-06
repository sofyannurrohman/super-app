import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';
import './delivery_tracking_page.dart';
import './delivery_history_page.dart';
import './cash_settlement_page.dart';
import './delivery_invoice_list_page.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_state.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:shimmer/shimmer.dart';
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/core/services/offline_delivery_manager.dart';
import './sales_stock_request_page.dart';
import './customer_return_history_page.dart';

class DeliveryDashboardPage extends StatefulWidget {
  const DeliveryDashboardPage({super.key});

  @override
  State<DeliveryDashboardPage> createState() => _DeliveryDashboardPageState();
}

class _DeliveryDashboardPageState extends State<DeliveryDashboardPage> {
  final ApiClient apiClient = di.sl<ApiClient>();
  bool _isLoading = true;
  List<dynamic> _tasks = [];
  List<dynamic> _driverLoad = [];

  @override
  void initState() {
    super.initState();
    _fetchDashboardData();
  }

  Future<void> _fetchDashboardData() async {
    if (!mounted) return;
    setState(() => _isLoading = true);

    // Load local cache first
    try {
      final prefs = await SharedPreferences.getInstance();
      final cachedTasks = prefs.getString('offline_delivery_tasks');
      final cachedLoad = prefs.getString('offline_driver_load');
      if (cachedTasks != null) {
        if (mounted) {
          setState(() {
            _tasks = jsonDecode(cachedTasks);
            if (cachedLoad != null) {
              _driverLoad = jsonDecode(cachedLoad);
            }
          });
        }
      }
    } catch (e) {
      debugPrint('Error loading cached tasks: $e');
    }

    // Fetch from backend
    try {
      final response = await apiClient.client.get('delivery/tasks');
      final loadResponse = await apiClient.client.get('delivery/driver-load');
      if (mounted) {
        setState(() {
          _tasks = response.data;
          _driverLoad = loadResponse.data['data'] ?? [];
        });
      }

      // Save to local cache
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(
        'offline_delivery_tasks',
        jsonEncode(response.data),
      );
      await prefs.setString('offline_driver_load', jsonEncode(_driverLoad));
    } catch (e) {
      debugPrint('Error fetching delivery data: $e');
      if (mounted && _tasks.isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Offline mode: Menampilkan data tersimpan.'),
            backgroundColor: Colors.orange,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _syncData() async {
    if (!mounted) return;
    setState(() => _isLoading = true);
    try {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Sinkronisasi muatan dan tugas...'),
          duration: Duration(seconds: 1),
        ),
      );

      // 1. Sync pending offline tasks
      await OfflineDeliveryManager().syncPendingDeliveries(apiClient);

      // 2. Fetch fresh data
      await _fetchDashboardData();

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Sinkronisasi berhasil!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal sinkronisasi: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    // Calculate stats
    int totalItems = 0;
    int deliveredItems = 0;
    double totalCash = 0;
    double totalTransfer = 0;
    double totalTempo = 0;

    for (var task in _tasks) {
      final List items = task['items'] as List? ?? [];
      totalItems += items.length;
      deliveredItems += items
          .where(
            (i) => (i['status'] ?? '').toString().toUpperCase() == 'DELIVERED',
          )
          .length;
      totalCash += (task['total_cash_collected'] ?? 0.0).toDouble();
      totalTransfer += (task['total_transfer_collected'] ?? 0.0).toDouble();
      totalTempo += (task['total_tempo_collected'] ?? 0.0).toDouble();
    }

    final int totalBatches = _tasks.length;
    final double overallProgress = totalItems > 0
        ? deliveredItems / totalItems
        : 0.0;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Column(
          children: [
            Text(
              'DASHBOARD PENGIRIMAN',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w900,
                color: Colors.white,
                fontSize: 14,
                letterSpacing: 1,
              ),
            ),
            BlocBuilder<ProfileBloc, ProfileState>(
              builder: (context, state) {
                String company = 'Perusahaan Demo';
                if (state is ProfileLoaded) {
                  company =
                      state.profile['company']?['name'] ?? 'Perusahaan Demo';
                }
                return Text(
                  company.toUpperCase(),
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w700,
                    color: AppColors.primaryGreenLight,
                    fontSize: 8,
                    letterSpacing: 1,
                  ),
                );
              },
            ),
          ],
        ),
        backgroundColor:
            AppColors.primaryGreenDark, // Premium Emerald Green Dark
        centerTitle: true,
        elevation: 0,
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
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
              ),
              child: Text(
                'KE HRIS',
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
            icon: const Icon(Icons.sync_rounded, color: Colors.white),
            onPressed: _syncData,
            tooltip: 'Sinkronisasi Muatan',
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _syncData,
        backgroundColor: Colors.white,
        color: AppColors.primaryGreen,
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildTasksHeaderHorizontal(),
              const SizedBox(height: 12),
              if (_isLoading)
                _buildHorizontalSkeleton()
              else if (_tasks.isEmpty)
                _buildHorizontalEmptyState()
              else
                _buildHorizontalTaskList(),
              const SizedBox(height: 24),
              _buildCarriedGoodsHeader(),
              const SizedBox(height: 24),
              _buildQuickStats(
                totalBatches,
                totalCash,
                totalTransfer,
                totalTempo,
              ),
              const SizedBox(height: 32),
              _buildQuickActions(context),
              const SizedBox(height: 40),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildHorizontalSkeleton() {
    return SizedBox(
      height: 130,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        itemCount: 3,
        itemBuilder: (context, index) => Shimmer.fromColors(
          baseColor: Colors.grey.shade300,
          highlightColor: Colors.grey.shade100,
          child: Container(
            width: 280,
            margin: const EdgeInsets.only(right: 16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildTasksHeaderHorizontal() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          'TUGAS PENGIRIMAN',
          style: GoogleFonts.outfit(
            fontSize: 12,
            fontWeight: FontWeight.w900,
            color: Colors.blueGrey,
            letterSpacing: 1.5,
          ),
        ),
        TextButton(
          onPressed: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const DeliveryTrackingPage()),
          ),
          child: Text(
            'Lihat Detail',
            style: GoogleFonts.outfit(
              fontSize: 11,
              fontWeight: FontWeight.w800,
              color: AppColors.primaryGreen,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildHorizontalTaskList() {
    return SizedBox(
      height: 130,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        physics: const BouncingScrollPhysics(),
        itemCount: _tasks.length,
        itemBuilder: (context, index) => _buildHorizontalTaskCard(
          context,
          _tasks[index],
          index == _tasks.length - 1,
        ),
      ),
    );
  }

  Widget _buildCarriedGoodsHeader() {
    int remainingItems = 0;

    for (var load in _driverLoad) {
      final qty = (load['stock'] ?? 0) as int;
      remainingItems += qty;
    }

    final hasCargo = _driverLoad.isNotEmpty;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [AppColors.primaryGreen, AppColors.primaryGreenDark],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(32),
        boxShadow: [
          BoxShadow(
            color: AppColors.primaryGreen.withOpacity(0.3),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'MUATAN ARMADA SAAT INI',
                      style: GoogleFonts.outfit(
                        color: Colors.white.withOpacity(0.7),
                        fontWeight: FontWeight.w800,
                        fontSize: 10,
                        letterSpacing: 1.5,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      hasCargo ? '$remainingItems Item Tersedia' : 'Kosong',
                      style: GoogleFonts.outfit(
                        color: Colors.white,
                        fontWeight: FontWeight.w900,
                        fontSize: 24,
                      ),
                    ),
                    Text(
                      'Telah disinkronisasi',
                      style: GoogleFonts.outfit(
                        color: Colors.white70,
                        fontSize: 11,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(16),
                ),
                child: const Icon(
                  Icons.local_shipping_rounded,
                  color: Colors.white,
                  size: 28,
                ),
              ),
            ],
          ),
          if (hasCargo) ...[
            const SizedBox(height: 16),
            const Divider(color: Colors.white24, height: 1),
            const SizedBox(height: 16),
            ConstrainedBox(
              constraints: const BoxConstraints(maxHeight: 180),
              child: ListView(
                shrinkWrap: true,
                physics: const BouncingScrollPhysics(),
                children: _driverLoad.map((load) {
                  final name = load['name'] ?? 'Produk';
                  final onTruck = load['stock'] ?? 0;
                  final unit = 'PCS';

                  return Padding(
                    padding: const EdgeInsets.only(bottom: 10),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Expanded(
                          child: Text(
                            name,
                            style: GoogleFonts.outfit(
                              color: Colors.white,
                              fontWeight: FontWeight.w700,
                              fontSize: 13,
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          '$onTruck $unit',
                          style: GoogleFonts.outfit(
                            color: onTruck > 0
                                ? Colors.white
                                : Colors.white.withOpacity(0.5),
                            fontWeight: FontWeight.w800,
                            fontSize: 13,
                          ),
                        ),
                      ],
                    ),
                  );
                }).toList(),
              ),
            ),
          ] else ...[
            const SizedBox(height: 16),
            Text(
              'Belum ada muatan aktif. Silakan ambil barang di gudang sesuai Surat Jalan.',
              style: GoogleFonts.outfit(
                color: Colors.white70,
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildQuickStats(
    int batches,
    double cash,
    double transfer,
    double tempo,
  ) {
    final formatter = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    return Column(
      children: [
        Row(
          children: [
            Expanded(
              child: _buildStatCard(
                'TOTAL SJ',
                batches.toString(),
                Icons.assignment_rounded,
                Colors.orange,
                'Surat Jalan',
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => const DeliveryTrackingPage(),
                    ),
                  );
                },
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: _buildStatCard(
                'KAS COD',
                formatter.format(cash),
                Icons.payments_rounded,
                Colors.green,
                'Uang Tunai',
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => DeliveryInvoiceListPage(
                        paymentMethod: 'CASH',
                        tasks: _tasks,
                      ),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        Row(
          children: [
            Expanded(
              child: _buildStatCard(
                'TRANSFER MASUK',
                formatter.format(transfer),
                Icons.account_balance_wallet_rounded,
                AppColors.primaryGreen,
                'Digital & QRIS',
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => DeliveryInvoiceListPage(
                        paymentMethod: 'TRANSFER',
                        tasks: _tasks,
                      ),
                    ),
                  );
                },
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: _buildStatCard(
                'TAGIHAN TEMPO',
                formatter.format(tempo),
                Icons.calendar_month_rounded,
                Colors.deepOrange,
                'Tempo / Kredit',
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => DeliveryInvoiceListPage(
                        paymentMethod: 'TEMPO',
                        tasks: _tasks,
                      ),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildStatCardFull(
    String label,
    String value,
    IconData icon,
    Color color,
    String subLabel,
  ) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(28),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
        border: Border.all(color: color.withOpacity(0.05)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: color.withOpacity(0.1),
              borderRadius: BorderRadius.circular(16),
            ),
            child: Icon(icon, color: color, size: 24),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: GoogleFonts.outfit(
                    color: Colors.blueGrey,
                    fontSize: 10,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 0.5,
                  ),
                ),
                Text(
                  value,
                  style: GoogleFonts.outfit(
                    color: const Color(0xFF0F172A),
                    fontWeight: FontWeight.w900,
                    fontSize: 20,
                  ),
                ),
                Text(
                  subLabel,
                  style: GoogleFonts.outfit(
                    color: Colors.blueGrey.withOpacity(0.6),
                    fontSize: 10,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatCard(
    String label,
    String value,
    IconData icon,
    Color color,
    String subLabel, {
    VoidCallback? onTap,
  }) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(28),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
        border: Border.all(color: color.withOpacity(0.05)),
      ),
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(28),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(28),
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: color.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Icon(icon, color: color, size: 18),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        label,
                        style: GoogleFonts.outfit(
                          color: Colors.blueGrey,
                          fontSize: 10,
                          fontWeight: FontWeight.w800,
                          letterSpacing: 0.5,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                FittedBox(
                  fit: BoxFit.scaleDown,
                  child: Text(
                    value,
                    style: GoogleFonts.outfit(
                      color: const Color(0xFF0F172A),
                      fontWeight: FontWeight.w900,
                      fontSize: 20,
                    ),
                  ),
                ),
                Text(
                  subLabel,
                  style: GoogleFonts.outfit(
                    color: Colors.blueGrey.withOpacity(0.6),
                    fontSize: 10,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildQuickActions(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'AKSES OPERASIONAL',
          style: GoogleFonts.outfit(
            fontSize: 11,
            fontWeight: FontWeight.w900,
            color: Colors.blueGrey,
            letterSpacing: 1.5,
          ),
        ),
        const SizedBox(height: 16),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            _buildActionItem(
              context,
              Icons.qr_code_scanner_rounded,
              'Scan SJ',
              const Color(0xFF6366F1),
              () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) =>
                        const DeliveryTrackingPage(initialScan: true),
                  ),
                );
              },
            ),
            _buildActionItem(
              context,
              Icons.payments_rounded,
              'Setor Uang',
              const Color(0xFFF59E0B),
              () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => CashSettlementPage(batches: _tasks),
                  ),
                );
              },
            ),
            _buildActionItem(
              context,
              Icons.local_shipping_rounded,
              'Tugas Saya',
              const Color(0xFF0EA5E9),
              () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const DeliveryTrackingPage(),
                  ),
                );
              },
            ),
          ],
        ),
        const SizedBox(height: 16),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            _buildActionItem(
              context,
              Icons.keyboard_return_rounded,
              'Retur Barang',
              const Color(0xFFEF4444),
              () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const SalesStockRequestPage(isReturn: true),
                  ),
                );
              },
            ),
            _buildActionItem(
              context,
              Icons.assignment_return_rounded,
              'Retur Pelanggan',
              const Color(0xFFE11D48),
              () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const CustomerReturnHistoryPage(),
                  ),
                );
              },
            ),
            _buildActionItem(
              context,
              Icons.history_rounded,
              'Riwayat',
              const Color(0xFF64748B),
              () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const DeliveryHistoryPage(),
                  ),
                );
              },
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildActionItem(
    BuildContext context,
    IconData icon,
    String label,
    Color color,
    VoidCallback onTap,
  ) {
    return Expanded(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(24),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(18),
              decoration: BoxDecoration(
                color: color.withOpacity(0.1),
                borderRadius: BorderRadius.circular(24),
              ),
              child: Icon(icon, color: color, size: 28),
            ),
            const SizedBox(height: 10),
            Text(
              label,
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(
                fontSize: 11,
                fontWeight: FontWeight.w700,
                color: const Color(0xFF1E293B),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHorizontalTaskCard(
    BuildContext context,
    dynamic task,
    bool isLast,
  ) {
    final String doNo = task['delivery_order_no'] ?? 'SJ-PENDING';
    final String status = task['status'] ?? 'WAITING';
    final List items = task['items'] as List? ?? [];
    final int itemsCount = items.length;
    final int deliveredCount = items
        .where(
          (i) => (i['status'] ?? '').toString().toUpperCase() == 'DELIVERED',
        )
        .length;

    Color statusColor = Colors.grey;
    String statusLabel = status;
    if (status == 'ON_DELIVERY') {
      statusColor = AppColors.primaryGreen;
      statusLabel = 'ON_DELIVERY';
    }
    if (status == 'COMPLETED') {
      statusColor = Colors.green;
      statusLabel = 'SELESAI';
    }
    if (status == 'WAITING_ASSIGNMENT') {
      statusColor = Colors.orange;
      statusLabel = 'MENUNGGU';
    }
    if (status == 'PICKING') {
      statusColor = const Color(0xFFF59E0B);
      statusLabel = 'PICKING';
    }
    if (status == 'PENDING') {
      statusColor = Colors.blueGrey;
      statusLabel = 'PENDING';
    }

    return Container(
      width: 280,
      margin: EdgeInsets.only(right: isLast ? 0 : 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 15,
            offset: const Offset(0, 8),
          ),
        ],
        border: Border.all(color: const Color(0xFFF1F5F9)),
      ),
      child: InkWell(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const DeliveryTrackingPage()),
          );
        },
        borderRadius: BorderRadius.circular(24),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      Icons.local_shipping_rounded,
                      color: statusColor,
                      size: 20,
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.12),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text(
                      statusLabel,
                      style: GoogleFonts.outfit(
                        fontSize: 9,
                        fontWeight: FontWeight.w800,
                        color: statusColor,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Text(
                'SJ #$doNo',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  fontSize: 16,
                  color: const Color(0xFF1E293B),
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 4),
              Row(
                children: [
                  const Icon(
                    Icons.store_rounded,
                    size: 12,
                    color: Colors.blueGrey,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    '$deliveredCount / $itemsCount Outlet Selesai',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      color: Colors.blueGrey,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildHorizontalEmptyState() {
    return Container(
      width: double.infinity,
      height: 130,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: const Color(0xFFF1F5F9)),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.assignment_turned_in_outlined,
            size: 32,
            color: Colors.blueGrey.withOpacity(0.2),
          ),
          const SizedBox(height: 10),
          Text(
            'Tidak ada tugas hari ini',
            style: GoogleFonts.outfit(
              color: const Color(0xFF0F172A),
              fontWeight: FontWeight.w800,
              fontSize: 14,
            ),
          ),
        ],
      ),
    );
  }
}
