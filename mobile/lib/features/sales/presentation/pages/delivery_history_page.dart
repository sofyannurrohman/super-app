import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';
import './delivery_tracking_page.dart';

class DeliveryHistoryPage extends StatefulWidget {
  const DeliveryHistoryPage({super.key});

  @override
  State<DeliveryHistoryPage> createState() => _DeliveryHistoryPageState();
}

class _DeliveryHistoryPageState extends State<DeliveryHistoryPage> {
  final ApiClient apiClient = di.sl<ApiClient>();
  bool _isLoading = true;
  List<dynamic> _history = [];

  @override
  void initState() {
    super.initState();
    _fetchHistory();
  }

  Future<void> _fetchHistory() async {
    if (!mounted) return;
    setState(() => _isLoading = true);
    try {
      final response = await apiClient.client.get('delivery/history');
      if (mounted) {
        setState(() {
          _history = response.data;
        });
      }
    } catch (e) {
      debugPrint('Error fetching delivery history: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Column(
          children: [
            Text(
              'RIWAYAT PENGIRIMAN',
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  color: const Color(0xFF1E293B),
                  fontSize: 14,
                  letterSpacing: 1),
            ),
            Text(
              'Tugas Selesai',
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w700,
                  color: Colors.blueAccent,
                  fontSize: 8,
                  letterSpacing: 1),
            ),
          ],
        ),
        backgroundColor: Colors.white,
        centerTitle: true,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded,
              color: Color(0xFF1E293B), size: 18),
          onPressed: () => Navigator.pop(context),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh_rounded, color: Colors.blueAccent),
            onPressed: _fetchHistory,
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _fetchHistory,
        child: _isLoading
            ? const Center(child: CircularProgressIndicator())
            : _history.isEmpty
                ? _buildEmptyState()
                : ListView.builder(
                    padding: const EdgeInsets.all(20),
                    itemCount: _history.length,
                    itemBuilder: (context, index) =>
                        _buildHistoryCard(_history[index]),
                  ),
      ),
    );
  }

  Widget _buildHistoryCard(dynamic task) {
    final String doNo = task['delivery_order_no'] ?? 'SJ-COMPLETED';
    final String finishedAtStr = task['finished_at'] ?? task['created_at'];
    final DateTime finishedAt = DateTime.parse(finishedAtStr).toLocal();
    final String dateStr = DateFormat('dd MMM yyyy, HH:mm').format(finishedAt);
    final int itemsCount = (task['items'] as List? ?? []).length;
    final double totalCash = (task['total_cash_collected'] ?? 0.0).toDouble();
    final double totalTransfer =
        (task['total_transfer_collected'] ?? 0.0).toDouble();
    final double totalTempo = (task['total_tempo_collected'] ?? 0.0).toDouble();
    final formatter =
        NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);

    int totalReturnQty = 0;
    final List itemsList = task['items'] as List? ?? [];
    for (var item in itemsList) {
      if (item['has_return'] == true) {
        totalReturnQty += (item['return_quantity'] ?? 0) as int;
      }
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 10,
              offset: const Offset(0, 4)),
        ],
        border: Border.all(color: const Color(0xFFF1F5F9)),
      ),
      child: InkWell(
        onTap: () {
          // You could navigate to a detail page here
          // For now, let's just show a snackbar or the tracking page (which handles detail)
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (_) => DeliveryTrackingPage(deliveryOrderNo: doNo),
            ),
          );
        },
        borderRadius: BorderRadius.circular(24),
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                      color: Colors.green.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.check_circle_rounded,
                            color: Colors.green, size: 12),
                        const SizedBox(width: 4),
                        Text(
                          'SELESAI',
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.bold,
                              color: Colors.green,
                              fontSize: 10),
                        ),
                      ],
                    ),
                  ),
                  Text(
                    dateStr,
                    style: GoogleFonts.outfit(
                        fontSize: 11,
                        color: Colors.blueGrey,
                        fontWeight: FontWeight.w600),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Text(
                'SJ #$doNo',
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w900,
                    fontSize: 16,
                    color: const Color(0xFF1E293B)),
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'OUTLET',
                          style: GoogleFonts.outfit(
                              fontSize: 9,
                              color: Colors.blueGrey,
                              fontWeight: FontWeight.w800,
                              letterSpacing: 0.5),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '$itemsCount Lokasi',
                          style: GoogleFonts.outfit(
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                              color: const Color(0xFF334155)),
                        ),
                        if (totalReturnQty > 0) ...[
                          const SizedBox(height: 4),
                          Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(
                              color: Colors.red.withOpacity(0.1),
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: Text(
                              'Retur: $totalReturnQty Unit',
                              style: GoogleFonts.outfit(
                                  fontSize: 9,
                                  color: Colors.red,
                                  fontWeight: FontWeight.bold),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                  Container(
                    width: 1,
                    height: 30,
                    color: const Color(0xFFF1F5F9),
                  ),
                  const SizedBox(width: 20),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'TOTAL REKONSILIASI PENGIRIMAN',
                          style: GoogleFonts.outfit(
                              fontSize: 9,
                              color: Colors.blueGrey,
                              fontWeight: FontWeight.w800,
                              letterSpacing: 0.5),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          formatter
                              .format(totalCash + totalTransfer + totalTempo),
                          style: GoogleFonts.outfit(
                              fontSize: 13,
                              fontWeight: FontWeight.w900,
                              color: const Color(0xFF1E293B)),
                        ),
                        Text(
                          'Cash: ${formatter.format(totalCash)} • Trf: ${formatter.format(totalTransfer)}${totalTempo > 0 ? " • Tempo: ${formatter.format(totalTempo)}" : ""}',
                          style: GoogleFonts.outfit(
                              fontSize: 8,
                              fontWeight: FontWeight.bold,
                              color: Colors.blueGrey),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              const Divider(height: 1, color: Color(0xFFF1F5F9)),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  Text(
                    'Lihat Detail',
                    style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: Colors.blueAccent,
                        fontWeight: FontWeight.w700),
                  ),
                  const SizedBox(width: 4),
                  const Icon(Icons.arrow_forward_ios_rounded,
                      size: 10, color: Colors.blueAccent),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.history_rounded,
              size: 80, color: Colors.blueGrey.withOpacity(0.1)),
          const SizedBox(height: 24),
          Text(
            'Belum Ada Riwayat',
            style: GoogleFonts.outfit(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: const Color(0xFF94A3B8)),
          ),
          const SizedBox(height: 8),
          Text(
            'Semua pengiriman yang telah Anda selesaikan\nakan muncul di sini.',
            textAlign: TextAlign.center,
            style: GoogleFonts.outfit(fontSize: 14, color: Colors.blueGrey),
          ),
        ],
      ),
    );
  }
}
