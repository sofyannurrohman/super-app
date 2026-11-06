import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:barcode_widget/barcode_widget.dart';

class CashSettlementPage extends StatefulWidget {
  final List<dynamic> batches;
  const CashSettlementPage({super.key, required this.batches});

  @override
  State<CashSettlementPage> createState() => _CashSettlementPageState();
}

class _CashSettlementPageState extends State<CashSettlementPage> {
  final formatter =
      NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);

  @override
  Widget build(BuildContext context) {
    List<dynamic> settlementItems = [];
    double totalCash = 0;
    double totalDigital = 0;

    for (var batch in widget.batches) {
      final items = batch['items'] as List? ?? [];
      for (var item in items) {
        if ((item['payment_amount'] ?? 0) > 0) {
          settlementItems.add({
            ...item,
            'batch_no': batch['delivery_order_no'],
          });

          final amount = (item['payment_amount'] ?? 0.0).toDouble();
          if (item['payment_method'] == 'CASH') {
            totalCash += amount;
          } else {
            totalDigital += amount;
          }
        }
      }
    }

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'VERIFIKASI SETORAN',
          style: GoogleFonts.outfit(
              fontWeight: FontWeight.w900,
              color: const Color(0xFF1E293B),
              fontSize: 14,
              letterSpacing: 1),
        ),
        backgroundColor: Colors.white,
        elevation: 0,
        centerTitle: true,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded,
              color: Color(0xFF1E293B), size: 18),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: Column(
        children: [
          // Header Summary
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(24),
            decoration: const BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.only(
                  bottomLeft: Radius.circular(32),
                  bottomRight: Radius.circular(32)),
            ),
            child: Column(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: _buildSummaryMini(
                          'TUNAI (CASH)', totalCash, Colors.green),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: _buildSummaryMini(
                          'DIGITAL (TRF)', totalDigital, Colors.blueAccent),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                const Divider(height: 1),
                const SizedBox(height: 16),
                Text(
                  'TOTAL SEMUA SETORAN',
                  style: GoogleFonts.outfit(
                      color: Colors.blueGrey,
                      fontSize: 10,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 1.5),
                ),
                Text(
                  formatter.format(totalCash + totalDigital),
                  style: GoogleFonts.outfit(
                      color: const Color(0xFF0F172A),
                      fontWeight: FontWeight.w900,
                      fontSize: 24),
                ),
              ],
            ),
          ),

          const SizedBox(height: 16),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: Row(
              children: [
                Text(
                  'DAFTAR NOTA MASUK (${settlementItems.length})',
                  style: GoogleFonts.outfit(
                      fontSize: 11,
                      fontWeight: FontWeight.w900,
                      color: Colors.blueGrey,
                      letterSpacing: 1),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),

          Expanded(
            child: settlementItems.isEmpty
                ? _buildEmptyState()
                : ListView.builder(
                    padding: const EdgeInsets.symmetric(horizontal: 20),
                    itemCount: settlementItems.length,
                    itemBuilder: (context, index) {
                      final item = settlementItems[index];
                      final store = (item['sales_order'] != null)
                          ? item['sales_order']['store']
                          : item['sales_transaction']['store'];
                      final amount = (item['payment_amount'] ?? 0.0).toDouble();
                      final String method = item['payment_method'] ?? 'UNK';
                      final isCash = method == 'CASH';
                      final String storeName = store['name'] ?? 'Toko';

                      return Container(
                        margin: const EdgeInsets.only(bottom: 12),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(24),
                          border: Border.all(color: const Color(0xFFF1F5F9)),
                        ),
                        child: Material(
                          color: Colors.transparent,
                          borderRadius: BorderRadius.circular(24),
                          child: InkWell(
                            borderRadius: BorderRadius.circular(24),
                            onTap: () =>
                                _showBarcodeModal(context, item, storeName),
                            child: Padding(
                              padding: const EdgeInsets.all(20),
                              child: Row(
                                children: [
                                  Container(
                                    padding: const EdgeInsets.all(12),
                                    decoration: BoxDecoration(
                                        color: (isCash
                                                ? Colors.green
                                                : Colors.blueAccent)
                                            .withOpacity(0.1),
                                        borderRadius:
                                            BorderRadius.circular(16)),
                                    child: Icon(
                                        isCash
                                            ? Icons.payments_rounded
                                            : Icons
                                                .account_balance_wallet_rounded,
                                        color: isCash
                                            ? Colors.green
                                            : Colors.blueAccent,
                                        size: 24),
                                  ),
                                  const SizedBox(width: 16),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(storeName,
                                            style: GoogleFonts.outfit(
                                                fontWeight: FontWeight.w800,
                                                fontSize: 15)),
                                        Row(
                                          children: [
                                            Container(
                                              padding:
                                                  const EdgeInsets.symmetric(
                                                      horizontal: 6,
                                                      vertical: 2),
                                              decoration: BoxDecoration(
                                                color: (isCash
                                                        ? Colors.green
                                                        : Colors.blueAccent)
                                                    .withOpacity(0.1),
                                                borderRadius:
                                                    BorderRadius.circular(4),
                                              ),
                                              child: Text(
                                                method,
                                                style: GoogleFonts.outfit(
                                                    fontSize: 8,
                                                    fontWeight: FontWeight.bold,
                                                    color: isCash
                                                        ? Colors.green
                                                        : Colors.blueAccent),
                                              ),
                                            ),
                                            const SizedBox(width: 6),
                                            Text('SJ: ${item['batch_no']}',
                                                style: GoogleFonts.outfit(
                                                    fontSize: 10,
                                                    color: Colors.blueGrey,
                                                    fontWeight:
                                                        FontWeight.w600)),
                                          ],
                                        ),
                                      ],
                                    ),
                                  ),
                                  Column(
                                    crossAxisAlignment: CrossAxisAlignment.end,
                                    children: [
                                      Text(
                                        formatter.format(amount),
                                        style: GoogleFonts.outfit(
                                            fontWeight: FontWeight.w900,
                                            fontSize: 14,
                                            color: const Color(0xFF1E293B)),
                                      ),
                                      Text(
                                        'Verifikasi Kasir',
                                        style: GoogleFonts.outfit(
                                            fontSize: 9,
                                            color: Colors.blueAccent,
                                            fontWeight: FontWeight.bold),
                                      ),
                                    ],
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryMini(String label, double amount, Color color) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: GoogleFonts.outfit(
              color: Colors.blueGrey,
              fontSize: 9,
              fontWeight: FontWeight.w800,
              letterSpacing: 0.5),
        ),
        Text(
          formatter.format(amount),
          style: GoogleFonts.outfit(
              color: color, fontWeight: FontWeight.w900, fontSize: 16),
        ),
      ],
    );
  }

  void _showBarcodeModal(BuildContext context, dynamic item, String storeName) {
    final amount = (item['payment_amount'] ?? 0.0).toDouble();
    final String method = item['payment_method'] ?? 'UNK';
    final String barcodeData = 'SETTLE|ITEM|${item['id']}';

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        padding: const EdgeInsets.all(32),
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(40)),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(2))),
            const SizedBox(height: 32),
            Text(
              'SCAN UNTUK VERIFIKASI',
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  fontSize: 14,
                  color: Colors.blueGrey,
                  letterSpacing: 1),
            ),
            const SizedBox(height: 8),
            Text(storeName,
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w900,
                    fontSize: 22,
                    color: const Color(0xFF1E293B))),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                color: (method == 'CASH' ? Colors.green : Colors.blueAccent)
                    .withOpacity(0.1),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Text(
                'METODE: $method',
                style: GoogleFonts.outfit(
                    fontSize: 10,
                    fontWeight: FontWeight.w900,
                    color: method == 'CASH' ? Colors.green : Colors.blueAccent),
              ),
            ),
            const SizedBox(height: 12),
            Text(
              formatter.format(amount),
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  fontSize: 28,
                  color: const Color(0xFF1E293B)),
            ),
            const SizedBox(height: 40),

            // Barcode
            Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                color: const Color(0xFFF8FAFC),
                borderRadius: BorderRadius.circular(24),
                border: Border.all(color: const Color(0xFFF1F5F9)),
              ),
              child: BarcodeWidget(
                barcode: Barcode.code128(),
                data: barcodeData,
                width: double.infinity,
                height: 120,
                drawText: false,
              ),
            ),
            const SizedBox(height: 16),
            Text(
              item['id'].toString().substring(0, 8).toUpperCase(),
              style: GoogleFonts.outfit(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color: Colors.blueGrey,
                  letterSpacing: 2),
            ),

            const SizedBox(height: 40),
            Text(
              'Minta Admin Kasir men-scan barcode ini untuk memverifikasi pembayaran $method ini.',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(
                  fontSize: 13,
                  color: Colors.blueGrey,
                  fontWeight: FontWeight.w500),
            ),
            const SizedBox(height: 40),
            ElevatedButton(
              onPressed: () => Navigator.pop(context),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF0F172A),
                foregroundColor: Colors.white,
                minimumSize: const Size(double.infinity, 54),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16)),
              ),
              child: const Text('TUTUP'),
            ),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.payments_outlined,
              size: 80, color: Colors.blueGrey.withOpacity(0.1)),
          const SizedBox(height: 24),
          Text('Tidak Ada Setoran',
              style: GoogleFonts.outfit(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Colors.blueGrey)),
          const SizedBox(height: 8),
          Text('Belum ada pembayaran yang tertagih hari ini.',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(fontSize: 14, color: Colors.blueGrey)),
        ],
      ),
    );
  }
}
