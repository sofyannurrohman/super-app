import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:intl/intl.dart';

class StoreOrderHistoryPage extends StatefulWidget {
  final StoreModel store;

  const StoreOrderHistoryPage({super.key, required this.store});

  @override
  State<StoreOrderHistoryPage> createState() => _StoreOrderHistoryPageState();
}

class _StoreOrderHistoryPageState extends State<StoreOrderHistoryPage> {
  final _apiService = StoreApiService(apiClient: di.sl<ApiClient>(), prefs: di.sl<SharedPreferences>());
  bool _isLoading = true;
  List<dynamic> _historyOrders = [];
  int _currentPage = 1;
  int _totalPages = 1;
  int _totalItems = 0;
  Set<String> _expandedOrderIds = {};

  @override
  void initState() {
    super.initState();
    _fetchHistory();
  }

  Future<void> _fetchHistory({int page = 1}) async {
    setState(() => _isLoading = true);
    try {
      final response = await _apiService.getStoreOrders(widget.store.id, page: page, limit: 10);
      setState(() {
        _historyOrders = response['data'] ?? [];
        final meta = response['meta'] ?? {};
        _currentPage = meta['page'] ?? 1;
        _totalPages = meta['total_pages'] ?? 1;
        _totalItems = meta['total'] ?? _historyOrders.length;
      });
    } catch (e) {
      debugPrint('Error fetching store orders history: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal memuat riwayat pesanan: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _toggleExpand(String orderId) {
    setState(() {
      if (_expandedOrderIds.contains(orderId)) {
        _expandedOrderIds.remove(orderId);
      } else {
        _expandedOrderIds.add(orderId);
      }
    });
  }

  String _formatCurrency(dynamic amount) {
    if (amount == null) return 'Rp 0';
    final number = double.tryParse(amount.toString()) ?? 0;
    return NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0).format(number);
  }

  String _formatDate(String? dateStr) {
    if (dateStr == null || dateStr.isEmpty) return '-';
    try {
      final dt = DateTime.parse(dateStr).toLocal();
      return DateFormat('dd MMM yyyy').format(dt);
    } catch (e) {
      return dateStr;
    }
  }

  Color _getStatusColor(String status) {
    switch (status.toUpperCase()) {
      case 'DELIVERED':
      case 'CONFIRMED':
        return Colors.green;
      case 'WAITING_STOCK':
      case 'PENDING':
        return Colors.orange;
      case 'CANCELLED':
      case 'REJECTED':
        return Colors.red;
      default:
        return Colors.blueGrey;
    }
  }

  Widget _buildOrderItem(dynamic item) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: Colors.indigo.shade50,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(Icons.inventory_2_outlined, size: 20, color: Colors.indigo.shade400),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  item['product']?['name'] ?? 'Produk Tidak Diketahui',
                  style: GoogleFonts.outfit(fontWeight: FontWeight.w700, fontSize: 13, color: Colors.black87),
                ),
                const SizedBox(height: 2),
                Text(
                  '${item['quantity']} x ${_formatCurrency(item['price'])}',
                  style: GoogleFonts.outfit(fontWeight: FontWeight.w500, fontSize: 11, color: Colors.black54),
                ),
              ],
            ),
          ),
          Text(
            _formatCurrency(item['subtotal']),
            style: GoogleFonts.outfit(fontWeight: FontWeight.w800, fontSize: 13, color: Colors.black87),
          ),
        ],
      ),
    );
  }

  Widget _buildOrderCard(dynamic order) {
    final orderId = order['id'] ?? '';
    final isExpanded = _expandedOrderIds.contains(orderId);
    final status = order['status'] ?? '-';
    final items = (order['items'] as List<dynamic>?) ?? [];
    
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InkWell(
            onTap: () => _toggleExpand(orderId),
            borderRadius: BorderRadius.circular(20),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    width: 48,
                    height: 48,
                    decoration: BoxDecoration(
                      color: Colors.blueAccent.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: const Icon(Icons.shopping_bag_outlined, color: Colors.blueAccent),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          order['so_number'] ?? 'SO Baru',
                          style: GoogleFonts.outfit(fontWeight: FontWeight.w800, fontSize: 12, color: const Color(0xFF1E293B)),
                        ),
                        const SizedBox(height: 4),
                        Row(
                          children: [
                            const Icon(Icons.calendar_today_outlined, size: 12, color: Colors.grey),
                            const SizedBox(width: 4),
                            Text(
                              _formatDate(order['created_at']),
                              style: GoogleFonts.outfit(fontWeight: FontWeight.w500, fontSize: 11, color: Colors.grey.shade600),
                            ),
                          ],
                        ),
                        const SizedBox(height: 4),
                        Row(
                          children: [
                            const Icon(Icons.person_outline, size: 12, color: Colors.grey),
                            const SizedBox(width: 4),
                            Text(
                              order['employee']?['first_name'] ?? '-',
                              style: GoogleFonts.outfit(fontWeight: FontWeight.w500, fontSize: 11, color: Colors.grey.shade600),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        _formatCurrency(order['total_amount']),
                        style: GoogleFonts.outfit(fontWeight: FontWeight.w800, fontSize: 14, color: Colors.green),
                      ),
                      const SizedBox(height: 6),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                        decoration: BoxDecoration(
                          color: _getStatusColor(status).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Text(
                          status,
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w800,
                            fontSize: 10,
                            color: _getStatusColor(status),
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          
          if (isExpanded)
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.grey.shade50,
                borderRadius: const BorderRadius.vertical(bottom: Radius.circular(20)),
                border: Border(top: BorderSide(color: Colors.grey.shade200)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Rincian Barang',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w800,
                      fontSize: 12,
                      color: Colors.blueGrey,
                      letterSpacing: 0.5,
                    ),
                  ),
                  const SizedBox(height: 12),
                  if (items.isEmpty)
                    Padding(
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      child: Text('Tidak ada item.', style: GoogleFonts.outfit(fontSize: 12, color: Colors.grey)),
                    ),
                  ...items.map((item) => _buildOrderItem(item)).toList(),
                ],
              ),
            ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Riwayat Pesanan',
              style: GoogleFonts.outfit(
                color: const Color(0xFF1E293B),
                fontWeight: FontWeight.w800,
                fontSize: 18,
              ),
            ),
            Text(
              widget.store.name,
              style: GoogleFonts.outfit(
                color: Colors.grey.shade600,
                fontWeight: FontWeight.w500,
                fontSize: 12,
              ),
            ),
          ],
        ),
        iconTheme: const IconThemeData(color: Color(0xFF1E293B)),
      ),
      body: _isLoading && _historyOrders.isEmpty
          ? const Center(child: CircularProgressIndicator(color: Colors.blueAccent))
          : RefreshIndicator(
              onRefresh: () => _fetchHistory(page: 1),
              child: _historyOrders.isEmpty
                  ? SingleChildScrollView(
                      physics: const AlwaysScrollableScrollPhysics(),
                      child: Container(
                        height: MediaQuery.of(context).size.height * 0.7,
                        alignment: Alignment.center,
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.history_rounded, size: 80, color: Colors.grey.shade300),
                            const SizedBox(height: 16),
                            Text(
                              'Belum ada riwayat pesanan.',
                              style: GoogleFonts.outfit(
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                                color: Colors.blueGrey,
                              ),
                            ),
                          ],
                        ),
                      ),
                    )
                  : ListView.builder(
                      padding: const EdgeInsets.all(16),
                      itemCount: _historyOrders.length + 1,
                      itemBuilder: (context, index) {
                        if (index == _historyOrders.length) {
                          // Pagination Controls
                          if (_totalPages > 1) {
                            return Padding(
                              padding: const EdgeInsets.symmetric(vertical: 16),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  IconButton(
                                    icon: const Icon(Icons.chevron_left),
                                    onPressed: _currentPage > 1
                                        ? () => _fetchHistory(page: _currentPage - 1)
                                        : null,
                                  ),
                                  Text(
                                    'Halaman $_currentPage dari $_totalPages',
                                    style: GoogleFonts.outfit(fontWeight: FontWeight.w600),
                                  ),
                                  IconButton(
                                    icon: const Icon(Icons.chevron_right),
                                    onPressed: _currentPage < _totalPages
                                        ? () => _fetchHistory(page: _currentPage + 1)
                                        : null,
                                  ),
                                ],
                              ),
                            );
                          }
                          return const SizedBox(height: 20);
                        }
                        return _buildOrderCard(_historyOrders[index]);
                      },
                    ),
            ),
    );
  }
}
