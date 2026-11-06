import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';

class CustomerReturnHistoryPage extends StatefulWidget {
  const CustomerReturnHistoryPage({super.key});

  @override
  State<CustomerReturnHistoryPage> createState() =>
      _CustomerReturnHistoryPageState();
}

class _CustomerReturnHistoryPageState extends State<CustomerReturnHistoryPage> {
  // Dummy data for now, ideally fetched from API
  final List<Map<String, dynamic>> _returns = [
    {
      'id': 'RTR-20231015-01',
      'date': '2023-10-15 14:30',
      'customer': 'Toko Maju Jaya',
      'delivery_order': 'SJ-20231015-001',
      'type': 'TUKAR BARANG', // Tukar Barang or Potong Nota
      'status': 'MENUNGGU GUDANG',
      'items': [
        {
          'name': 'Produk A - Varian X',
          'quantity': 2,
          'reason': 'Barang Cacat'
        },
      ],
    },
    {
      'id': 'RTR-20231014-02',
      'date': '2023-10-14 10:15',
      'customer': 'Apotek Sehat',
      'delivery_order': 'SJ-20231014-012',
      'type': 'POTONG NOTA',
      'status': 'SELESAI',
      'items': [
        {
          'name': 'Produk B - Varian Y',
          'quantity': 1,
          'reason': 'Kemasan Rusak'
        },
        {
          'name': 'Produk C - Varian Z',
          'quantity': 3,
          'reason': 'Expired Dekat'
        },
      ],
    },
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'RIWAYAT RETUR PELANGGAN',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.w900,
            color: Colors.white,
            fontSize: 14,
            letterSpacing: 1,
          ),
        ),
        backgroundColor: AppColors.primaryGreenDark,
        centerTitle: true,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded,
              color: Colors.white, size: 20),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: _returns.isEmpty
          ? _buildEmptyState()
          : ListView.builder(
              padding: const EdgeInsets.all(20),
              itemCount: _returns.length,
              physics: const BouncingScrollPhysics(),
              itemBuilder: (context, index) {
                return _buildReturnCard(_returns[index]);
              },
            ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.assignment_return_rounded,
              size: 64, color: Colors.blueGrey.withOpacity(0.2)),
          const SizedBox(height: 16),
          Text(
            'Belum ada riwayat retur',
            style: GoogleFonts.outfit(
                color: const Color(0xFF0F172A),
                fontWeight: FontWeight.w800,
                fontSize: 16),
          ),
          const SizedBox(height: 8),
          Text(
            'Data retur dari pelanggan akan tampil di sini',
            style: GoogleFonts.outfit(color: Colors.blueGrey, fontSize: 12),
          ),
        ],
      ),
    );
  }

  Widget _buildReturnCard(Map<String, dynamic> returnData) {
    final bool isTukar = returnData['type'] == 'TUKAR BARANG';
    final Color typeColor =
        isTukar ? const Color(0xFF0EA5E9) : const Color(0xFFF43F5E);
    final Color statusColor = returnData['status'] == 'SELESAI'
        ? AppColors.primaryGreen
        : const Color(0xFFF59E0B);

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.03),
              blurRadius: 10,
              offset: const Offset(0, 4)),
        ],
        border: Border.all(color: const Color(0xFFF1F5F9)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header Card
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: const BorderRadius.only(
                  topLeft: Radius.circular(20), topRight: Radius.circular(20)),
              border: Border(bottom: BorderSide(color: Colors.grey.shade200)),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  returnData['date'],
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 11,
                      color: Colors.blueGrey,
                      fontWeight: FontWeight.w600),
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: statusColor.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    returnData['status'],
                    style: GoogleFonts.outfit(
                        fontSize: 9,
                        fontWeight: FontWeight.w800,
                        color: statusColor,
                        letterSpacing: 0.5),
                  ),
                ),
              ],
            ),
          ),

          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            returnData['customer'],
                            style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w900,
                                fontSize: 16,
                                color: const Color(0xFF1E293B)),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'Ref: ${returnData['delivery_order']}',
                            style: GoogleFonts.outfit(
                                fontSize: 12,
                                color: Colors.blueGrey,
                                fontWeight: FontWeight.w500),
                          ),
                        ],
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 10, vertical: 6),
                      decoration: BoxDecoration(
                        color: typeColor,
                        borderRadius: BorderRadius.circular(10),
                        boxShadow: [
                          BoxShadow(
                              color: typeColor.withOpacity(0.3),
                              blurRadius: 8,
                              offset: const Offset(0, 4)),
                        ],
                      ),
                      child: Row(
                        children: [
                          Icon(
                              isTukar
                                  ? Icons.swap_horiz_rounded
                                  : Icons.content_cut_rounded,
                              color: Colors.white,
                              size: 14),
                          const SizedBox(width: 4),
                          Text(
                            returnData['type'],
                            style: GoogleFonts.outfit(
                                fontSize: 10,
                                fontWeight: FontWeight.w800,
                                color: Colors.white,
                                letterSpacing: 0.5),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                const Divider(height: 1),
                const SizedBox(height: 16),
                Text(
                  'DETAIL BARANG:',
                  style: GoogleFonts.outfit(
                      fontSize: 10,
                      fontWeight: FontWeight.w800,
                      color: Colors.blueGrey,
                      letterSpacing: 1),
                ),
                const SizedBox(height: 8),
                ...((returnData['items'] as List).map((item) {
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          margin: const EdgeInsets.only(top: 2),
                          padding: const EdgeInsets.all(4),
                          decoration: BoxDecoration(
                              color: Colors.red.shade50,
                              borderRadius: BorderRadius.circular(4)),
                          child: Text(
                            '${item['quantity']}x',
                            style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                                color: Colors.red.shade700),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                item['name'],
                                style: GoogleFonts.plusJakartaSans(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w700,
                                    color: const Color(0xFF1E293B)),
                              ),
                              const SizedBox(height: 2),
                              Text(
                                'Alasan: ${item['reason']}',
                                style: GoogleFonts.plusJakartaSans(
                                    fontSize: 11,
                                    color: Colors.blueGrey,
                                    fontWeight: FontWeight.w500),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  );
                }).toList()),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
