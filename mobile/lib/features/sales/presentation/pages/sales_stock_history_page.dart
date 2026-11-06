import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';

class SalesStockHistoryPage extends StatefulWidget {
  const SalesStockHistoryPage({super.key});

  @override
  State<SalesStockHistoryPage> createState() => _SalesStockHistoryPageState();
}

class _SalesStockHistoryPageState extends State<SalesStockHistoryPage>
    with SingleTickerProviderStateMixin {
  final ApiClient apiClient = di.sl<ApiClient>();
  bool _isLoading = false;
  List<dynamic> _transfers = [];
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _fetchHistory();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _fetchHistory() async {
    setState(() => _isLoading = true);
    try {
      final profileRes = await apiClient.client.get('employees/profile');
      final employeeId = profileRes.data['data']['id'];

      final response =
          await apiClient.client.get('sales-transfers/history/$employeeId');
      if (mounted) {
        setState(() {
          _transfers = response.data['data'] ?? [];
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Gagal memuat riwayat: $e')));
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final requests = _transfers.where((t) => t['type'] == 'TRANSFER').toList();
    final returns = _transfers.where((t) => t['type'] == 'RETURN').toList();

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: Text(
          'Riwayat Ambil Barang',
          style: GoogleFonts.outfit(
              color: const Color(0xFF0F172A), fontWeight: FontWeight.bold),
        ),
        iconTheme: const IconThemeData(color: Color(0xFF0F172A)),
        bottom: TabBar(
          controller: _tabController,
          labelColor: AppColors.blueSales,
          unselectedLabelColor: const Color(0xFF64748B),
          indicatorColor: AppColors.blueSales,
          labelStyle: GoogleFonts.outfit(fontWeight: FontWeight.bold),
          tabs: const [
            Tab(text: 'Request Barang'),
            Tab(text: 'Retur Barang'),
          ],
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tabController,
              children: [
                _buildList(requests),
                _buildList(returns),
              ],
            ),
    );
  }

  Widget _buildList(List<dynamic> items) {
    if (items.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.history_rounded,
                size: 64, color: Color(0xFFCBD5E1)),
            const SizedBox(height: 16),
            Text(
              'Belum ada riwayat',
              style: GoogleFonts.outfit(
                  color: const Color(0xFF64748B), fontSize: 16),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchHistory,
      child: ListView.separated(
        padding: const EdgeInsets.all(16),
        itemCount: items.length,
        separatorBuilder: (context, index) => const SizedBox(height: 12),
        itemBuilder: (context, index) {
          final item = items[index];
          final product = item['product'] ?? {};
          final status = item['status'] ?? 'PENDING';
          final qty = item['quantity'] ?? 0;
          final unit = item['unit'] ?? 'pcs';
          final ref = item['reference_no'] ?? '-';
          final dateStr = item['transfer_date'] ?? item['created_at'];
          DateTime? date;
          if (dateStr != null) {
            date = DateTime.tryParse(dateStr)?.toLocal();
          }

          Color statusColor;
          String statusText;
          if (status == 'COMPLETED' || status == 'APPROVED') {
            statusColor = const Color(0xFF10B981);
            statusText = 'Disetujui';
          } else if (status == 'CANCELLED' || status == 'REJECTED') {
            statusColor = const Color(0xFFEF4444);
            statusText = 'Ditolak';
          } else if (status == 'WAITING_SUPERVISOR') {
            statusColor = const Color(0xFF3B82F6);
            statusText = 'Menunggu SPV';
          } else if (status == 'PENDING') {
            statusColor = const Color(0xFFF59E0B);
            statusText = 'Menunggu Gudang';
          } else {
            statusColor = const Color(0xFFF59E0B);
            statusText = 'Menunggu';
          }

          return Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: const Color(0xFFE2E8F0)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      ref,
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w600,
                        color: const Color(0xFF64748B),
                        fontSize: 12,
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: statusColor.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        statusText,
                        style: GoogleFonts.outfit(
                          color: statusColor,
                          fontSize: 10,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Container(
                      width: 48,
                      height: 48,
                      decoration: BoxDecoration(
                        color: const Color(0xFFF1F5F9),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Icon(Icons.inventory_2_rounded,
                          color: Color(0xFF94A3B8)),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            product['name'] ?? 'Produk',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.bold,
                              fontSize: 16,
                              color: const Color(0xFF0F172A),
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            '$qty $unit',
                            style: GoogleFonts.outfit(
                              color: AppColors.blueSales,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                if (date != null) ...[
                  const Divider(height: 24, color: Color(0xFFF1F5F9)),
                  Row(
                    children: [
                      const Icon(Icons.calendar_today_rounded,
                          size: 14, color: Color(0xFF94A3B8)),
                      const SizedBox(width: 4),
                      Text(
                        DateFormat('dd MMM yyyy HH:mm').format(date),
                        style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: const Color(0xFF64748B),
                        ),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          );
        },
      ),
    );
  }
}
