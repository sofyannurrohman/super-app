import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:share_plus/share_plus.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:hris_app/injection.dart' as di;
import './digital_receipt_page.dart';

class VisitSuccessPage extends StatefulWidget {
  final Map<String, dynamic> transaction;
  final String? localId;
  final Uint8List? invoicePdfBytes;

  const VisitSuccessPage({
    super.key,
    required this.transaction,
    this.localId,
    this.invoicePdfBytes,
  });

  @override
  State<VisitSuccessPage> createState() => _VisitSuccessPageState();
}

class _VisitSuccessPageState extends State<VisitSuccessPage>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _scaleAnimation;
  late Animation<double> _opacityAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1000),
    );
    _scaleAnimation = CurvedAnimation(
      parent: _controller,
      curve: Curves.elasticOut,
    );
    _opacityAnimation = CurvedAnimation(
      parent: _controller,
      curve: const Interval(0.0, 0.5, curve: Curves.easeIn),
    );
    _controller.forward();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  String _formatItemQty(Map<String, dynamic> item) {
    final breakdown = item['breakdown'];
    if (breakdown != null && breakdown.toString().isNotEmpty) {
      return breakdown.toString();
    }

    final qty = item['quantity'] ?? 0;
    final unit = item['unit'] ?? 'PCS';
    final piecesPerUnit = item['pieces_per_unit'] ?? 1;

    if (piecesPerUnit > 1) {
      final bulkQty = qty ~/ piecesPerUnit;
      final looseQty = qty % piecesPerUnit;
      final bulkUnit = unit;
      const baseUnit = 'PCS';

      if (bulkQty > 0 && looseQty > 0) {
        return '$bulkQty $bulkUnit + $looseQty $baseUnit';
      } else if (bulkQty > 0) {
        return '$bulkQty $bulkUnit';
      } else {
        return '$looseQty $baseUnit';
      }
    }

    return '$qty $unit';
  }

  Future<void> _sharePdfToWhatsApp() async {
    if (widget.invoicePdfBytes == null) return;

    final receiptNo = widget.transaction['receipt_no'] ?? 'NO-DATA';
    final companyName =
        widget.transaction['employee']?['company']?['name'] ??
        widget.transaction['company']?['name'] ??
        'Perusahaan Demo';

    try {
      if (kIsWeb) {
        final xFile = XFile.fromData(
          widget.invoicePdfBytes!,
          mimeType: 'application/pdf',
          name: 'Faktur_$receiptNo.pdf',
        );
        await Share.shareXFiles(
          [xFile],
          text:
              'Halo! Berikut terlampir salinan digital Faktur Penjualan resmi Anda ($receiptNo) dari $companyName. Terima kasih!',
        );
      } else {
        final tempDir = Directory.systemTemp;
        final file = File('${tempDir.path}/Faktur_$receiptNo.pdf');
        await file.writeAsBytes(widget.invoicePdfBytes!);

        final xFile = XFile(file.path, mimeType: 'application/pdf');
        await Share.shareXFiles(
          [xFile],
          text:
              'Halo! Berikut terlampir salinan digital Faktur Penjualan resmi Anda ($receiptNo) dari $companyName. Terima kasih!',
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal berbagi PDF: $e', style: GoogleFonts.outfit()),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final currency = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    final amount = (widget.transaction['total_amount'] ?? 0.0).toDouble();
    final storeName = widget.transaction['store']?['name'] ?? 'Toko';
    final receiptNo = widget.transaction['receipt_no'] ?? 'NO-DATA';
    final isSO = widget.transaction['payment_method'] == 'SALES_ORDER';
    final items = widget.transaction['items'] as List?;
    final hasItems = items != null && items.isNotEmpty;
    final isPhysicalReceipt = widget.transaction['receipt_path'] != null;

    return Scaffold(
      backgroundColor: Colors.white,
      body: Stack(
        children: [
          // Background decorations
          Positioned(
            top: -50,
            right: -50,
            child: Container(
              width: 200,
              height: 200,
              decoration: BoxDecoration(
                color: Colors.green.withOpacity(0.05),
                shape: BoxShape.circle,
              ),
            ),
          ),
          Positioned(
            bottom: 100,
            left: -30,
            child: Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                color: Colors.blueAccent.withOpacity(0.05),
                shape: BoxShape.circle,
              ),
            ),
          ),

          // Random decorative circles for confetti feel
          Positioned(top: 100, left: 40, child: _dot(Colors.green, 12)),
          Positioned(top: 150, right: 60, child: _dot(Colors.blueAccent, 8)),
          Positioned(bottom: 250, right: 40, child: _dot(Colors.orange, 10)),
          Positioned(bottom: 200, left: 80, child: _dot(Colors.pinkAccent, 6)),

          SafeArea(
            child: LayoutBuilder(
              builder: (context, constraints) {
                return SingleChildScrollView(
                  physics: const BouncingScrollPhysics(),
                  child: ConstrainedBox(
                    constraints: BoxConstraints(
                      minHeight: constraints.maxHeight,
                    ),
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 32),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const SizedBox(height: 40),

                          // Top notification style badge
                          ScaleTransition(
                            scale: _scaleAnimation,
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 24,
                                vertical: 12,
                              ),
                              decoration: BoxDecoration(
                                color: isPhysicalReceipt
                                    ? Colors.indigo
                                    : Colors.green,
                                borderRadius: BorderRadius.circular(100),
                                boxShadow: [
                                  BoxShadow(
                                    color:
                                        (isPhysicalReceipt
                                                ? Colors.indigo
                                                : Colors.green)
                                            .withOpacity(0.3),
                                    blurRadius: 20,
                                    offset: const Offset(0, 10),
                                  ),
                                ],
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  const Icon(
                                    Icons.notifications_active_rounded,
                                    color: Colors.white,
                                    size: 18,
                                  ),
                                  const SizedBox(width: 8),
                                  Text(
                                    isPhysicalReceipt
                                        ? 'NOTA FISIK UPLOADED!'
                                        : (hasItems
                                              ? (isSO
                                                    ? 'SALES ORDER TERSIMPAN!'
                                                    : 'PEMBAYARAN BERHASIL!')
                                              : 'KUNJUNGAN BERHASIL!'),
                                    style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w900,
                                      color: Colors.white,
                                      fontSize: 12,
                                      letterSpacing: 1,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),

                          const SizedBox(height: 48),

                          // Main Icon
                          ScaleTransition(
                            scale: _scaleAnimation,
                            child: Container(
                              width: 120,
                              height: 120,
                              decoration: BoxDecoration(
                                color: Colors.green.withOpacity(0.1),
                                shape: BoxShape.circle,
                              ),
                              child: const Center(
                                child: Icon(
                                  Icons.check_circle_rounded,
                                  color: Colors.green,
                                  size: 80,
                                ),
                              ),
                            ),
                          ),

                          const SizedBox(height: 32),

                          FadeTransition(
                            opacity: _opacityAnimation,
                            child: Column(
                              children: [
                                Text(
                                  isPhysicalReceipt
                                      ? 'Upload Nota Berhasil'
                                      : (hasItems
                                            ? (isSO
                                                  ? 'Order Berhasil Dibuat'
                                                  : 'Pembayaran Berhasil')
                                            : (widget.transaction['notes'] ==
                                                      'Input Transaksi Nanti'
                                                  ? 'Kunjungan Disimpan'
                                                  : 'Kunjungan Selesai')),
                                  style: GoogleFonts.outfit(
                                    fontSize: 28,
                                    fontWeight: FontWeight.w900,
                                    color: const Color(0xFF1E293B),
                                  ),
                                ),
                                const SizedBox(height: 12),
                                Text(
                                  isPhysicalReceipt
                                      ? 'Foto nota fisik dan nominal transaksi untuk $storeName telah berhasil disimpan dan akan diproses oleh admin.'
                                      : (hasItems
                                            ? (isSO
                                                  ? 'Pesanan untuk $storeName telah berhasil disimpan and menunggu verifikasi admin.'
                                                  : 'Kunjungan dan pembayaran untuk $storeName telah berhasil diselesaikan.')
                                            : (widget.transaction['notes'] ==
                                                      'Input Transaksi Nanti'
                                                  ? 'Kunjungan untuk $storeName telah berhasil disimpan. Jangan lupa untuk menginput transaksi di kemudian waktu.'
                                                  : 'Kunjungan untuk $storeName telah berhasil diselesaikan.')),
                                  textAlign: TextAlign.center,
                                  style: GoogleFonts.outfit(
                                    fontSize: 15,
                                    color: Colors.blueGrey,
                                    fontWeight: FontWeight.w600,
                                    height: 1.5,
                                  ),
                                ),
                              ],
                            ),
                          ),

                          const SizedBox(height: 48),

                          // Summary Card
                          FadeTransition(
                            opacity: _opacityAnimation,
                            child: Container(
                              padding: const EdgeInsets.all(24),
                              decoration: BoxDecoration(
                                color: const Color(0xFFF8FAFC),
                                borderRadius: BorderRadius.circular(24),
                                border: Border.all(
                                  color: const Color(0xFFE2E8F0),
                                ),
                              ),
                              child: Column(
                                children: [
                                  if (hasItems) ...[
                                    _summaryRow('ID Transaksi', receiptNo),
                                    const Padding(
                                      padding: EdgeInsets.symmetric(
                                        vertical: 12,
                                      ),
                                      child: Divider(),
                                    ),
                                  ],

                                  // Item Details
                                  if (widget.transaction['items'] != null) ...[
                                    ...((widget.transaction['items'] as List).map((
                                      item,
                                    ) {
                                      final itemMap = Map<String, dynamic>.from(
                                        item as Map,
                                      );
                                      return Padding(
                                        padding: const EdgeInsets.only(
                                          bottom: 8,
                                        ),
                                        child: Row(
                                          mainAxisAlignment:
                                              MainAxisAlignment.spaceBetween,
                                          children: [
                                            Expanded(
                                              child: Column(
                                                crossAxisAlignment:
                                                    CrossAxisAlignment.start,
                                                children: [
                                                  Text(
                                                    '${itemMap['product_name']}',
                                                    style: GoogleFonts.outfit(
                                                      fontSize: 13,
                                                      color: const Color(
                                                        0xFF1E293B,
                                                      ),
                                                      fontWeight:
                                                          FontWeight.w600,
                                                    ),
                                                  ),
                                                  if (itemMap['price_tier'] !=
                                                          null &&
                                                      itemMap['price_tier'] !=
                                                          'NORMAL')
                                                    Text(
                                                      itemMap['price_tier'] ==
                                                              'DBP'
                                                          ? 'Distributor (DBP)'
                                                          : (itemMap['price_tier'] ==
                                                                    'WBP'
                                                                ? 'Grosir (WBP)'
                                                                : 'Custom Booster'),
                                                      style: GoogleFonts.outfit(
                                                        fontSize: 10,
                                                        color: Colors
                                                            .purple
                                                            .shade700,
                                                        fontWeight:
                                                            FontWeight.w700,
                                                      ),
                                                    ),
                                                ],
                                              ),
                                            ),
                                            Text(
                                              'x${_formatItemQty(itemMap)}',
                                              style: GoogleFonts.outfit(
                                                fontSize: 13,
                                                color: Colors.blueGrey,
                                                fontWeight: FontWeight.w800,
                                              ),
                                            ),
                                          ],
                                        ),
                                      );
                                    })),
                                    const Divider(height: 16),
                                  ],

                                  // Return Items
                                  if (widget.transaction['return_items'] !=
                                          null &&
                                      (widget.transaction['return_items']
                                              as List)
                                          .isNotEmpty) ...[
                                    Padding(
                                      padding: const EdgeInsets.only(
                                        bottom: 12,
                                      ),
                                      child: Text(
                                        'Retur Barang Sampingan',
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          color: Colors.red.shade700,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                    ),
                                    ...((widget.transaction['return_items']
                                            as List)
                                        .map((item) {
                                          final itemMap =
                                              Map<String, dynamic>.from(
                                                item as Map,
                                              );
                                          return Padding(
                                            padding: const EdgeInsets.only(
                                              bottom: 8,
                                            ),
                                            child: Row(
                                              mainAxisAlignment:
                                                  MainAxisAlignment
                                                      .spaceBetween,
                                              children: [
                                                Expanded(
                                                  child: Text(
                                                    '${itemMap['product_name']}',
                                                    style: GoogleFonts.outfit(
                                                      fontSize: 13,
                                                      color:
                                                          Colors.red.shade900,
                                                      fontWeight:
                                                          FontWeight.w600,
                                                    ),
                                                  ),
                                                ),
                                                Text(
                                                  'x${_formatItemQty(itemMap)}',
                                                  style: GoogleFonts.outfit(
                                                    fontSize: 13,
                                                    color: Colors.red.shade400,
                                                    fontWeight: FontWeight.w800,
                                                  ),
                                                ),
                                              ],
                                            ),
                                          );
                                        })),
                                    const Divider(height: 24),
                                  ],

                                  if (hasItems) ...[
                                    _summaryRow(
                                      'Metode Pembayaran',
                                      widget.transaction['payment_method'] ==
                                              'SALES_ORDER'
                                          ? 'Sales Order'
                                          : (widget.transaction['payment_method'] ??
                                                '-'),
                                    ),
                                    if (widget.transaction['payment_method'] ==
                                            'TEMPO' &&
                                        widget.transaction['payment_due_date'] !=
                                            null) ...[
                                      const SizedBox(height: 8),
                                      _summaryRow(
                                        'Jatuh Tempo',
                                        DateFormat('dd MMM yyyy').format(
                                          DateTime.parse(
                                            widget
                                                .transaction['payment_due_date'],
                                          ).toLocal(),
                                        ),
                                      ),
                                    ],
                                    if ((widget.transaction['dp_amount']
                                                    as num?)
                                                ?.toDouble() !=
                                            null &&
                                        (widget.transaction['dp_amount'] as num)
                                                .toDouble() >
                                            0) ...[
                                      const SizedBox(height: 8),
                                      _summaryRow(
                                        'Uang Muka (DP)',
                                        currency.format(
                                          widget.transaction['dp_amount'],
                                        ),
                                      ),
                                    ],
                                    const Padding(
                                      padding: EdgeInsets.symmetric(
                                        vertical: 12,
                                      ),
                                      child: Divider(),
                                    ),
                                    if (widget.transaction['return_items'] !=
                                            null &&
                                        (widget.transaction['return_items']
                                                as List)
                                            .isNotEmpty) ...[
                                      _summaryRow(
                                        'Total Belanja',
                                        currency.format(
                                          widget.transaction['total_amount'] ??
                                              amount,
                                        ),
                                      ),
                                      _summaryRow(
                                        'Total Potongan Retur',
                                        '- ${currency.format((widget.transaction['return_items'] as List).fold(0.0, (sum, item) => sum + (item['quantity'] ?? 0) * ((item['price'] ?? 0) as num).toDouble()))}',
                                        valueColor: Colors.red.shade700,
                                      ),
                                      const SizedBox(height: 8),
                                    ],
                                    _summaryRow(
                                      'Total Nominal (Bayar)',
                                      currency.format(amount),
                                      isBold: true,
                                    ),
                                  ] else if (isPhysicalReceipt) ...[
                                    _summaryRow('Nomor Nota Manual', receiptNo),
                                    const Padding(
                                      padding: EdgeInsets.symmetric(
                                        vertical: 12,
                                      ),
                                      child: Divider(),
                                    ),
                                    _summaryRow('Toko / Customer', storeName),
                                    const Padding(
                                      padding: EdgeInsets.symmetric(
                                        vertical: 12,
                                      ),
                                      child: Divider(),
                                    ),
                                    _summaryRow(
                                      'Status Nota',
                                      'Menunggu Verifikasi Admin',
                                      isBold: true,
                                    ),
                                    const Padding(
                                      padding: EdgeInsets.symmetric(
                                        vertical: 12,
                                      ),
                                      child: Divider(),
                                    ),
                                    _summaryRow(
                                      'Total Nominal',
                                      currency.format(amount),
                                      isBold: true,
                                    ),
                                  ] else ...[
                                    _summaryRow('Toko / Customer', storeName),
                                    const Padding(
                                      padding: EdgeInsets.symmetric(
                                        vertical: 12,
                                      ),
                                      child: Divider(),
                                    ),
                                    _summaryRow(
                                      'Status Kunjungan',
                                      widget.transaction['notes'] ==
                                              'Input Transaksi Nanti'
                                          ? 'Menunggu Input Transaksi'
                                          : 'Selesai (Tanpa Transaksi)',
                                      isBold: true,
                                    ),
                                    if (widget.transaction['notes'] != null &&
                                        widget.transaction['notes']
                                            .toString()
                                            .isNotEmpty) ...[
                                      const Padding(
                                        padding: EdgeInsets.symmetric(
                                          vertical: 12,
                                        ),
                                        child: Divider(),
                                      ),
                                      _summaryRow(
                                        'Alasan / Catatan',
                                        widget.transaction['notes'].toString(),
                                      ),
                                    ],
                                  ],
                                ],
                              ),
                            ),
                          ),

                          const SizedBox(height: 40),

                          // Buttons
                          FadeTransition(
                            opacity: _opacityAnimation,
                            child: Column(
                              children: [
                                if (hasItems && !isPhysicalReceipt) ...[
                                  SizedBox(
                                    width: double.infinity,
                                    height: 64,
                                    child: ElevatedButton(
                                      onPressed: () {
                                        Navigator.of(context).push(
                                          MaterialPageRoute(
                                            builder: (_) => DigitalReceiptPage(
                                              transaction: widget.transaction,
                                              localId: widget.localId,
                                            ),
                                          ),
                                        );
                                      },
                                      style: ElevatedButton.styleFrom(
                                        backgroundColor: const Color(
                                          0xFF1E293B,
                                        ),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(
                                            20,
                                          ),
                                        ),
                                        elevation: 0,
                                      ),
                                      child: Text(
                                        'LIHAT NOTA DIGITAL',
                                        style: GoogleFonts.outfit(
                                          fontWeight: FontWeight.w900,
                                          fontSize: 16,
                                          color: Colors.white,
                                          letterSpacing: 0.5,
                                        ),
                                      ),
                                    ),
                                  ),
                                  if (widget.invoicePdfBytes != null) ...[
                                    const SizedBox(height: 16),
                                    SizedBox(
                                      width: double.infinity,
                                      height: 60,
                                      child: ElevatedButton.icon(
                                        onPressed: _sharePdfToWhatsApp,
                                        icon: const Icon(
                                          Icons.share_rounded,
                                          color: Colors.white,
                                        ),
                                        label: Text(
                                          'KIRIM FAKTUR VIA WHATSAPP',
                                          style: GoogleFonts.outfit(
                                            fontWeight: FontWeight.w900,
                                            fontSize: 14,
                                            color: Colors.white,
                                            letterSpacing: 0.5,
                                          ),
                                        ),
                                        style: ElevatedButton.styleFrom(
                                          backgroundColor:
                                              Colors.green.shade600,
                                          shape: RoundedRectangleBorder(
                                            borderRadius: BorderRadius.circular(
                                              20,
                                            ),
                                          ),
                                          elevation: 0,
                                        ),
                                      ),
                                    ),
                                  ],
                                  const SizedBox(height: 16),
                                  TextButton(
                                    onPressed: () {
                                      di.sl<SyncBloc>().add(
                                        const SyncDataRequested(force: true),
                                      );
                                      Navigator.of(context).popUntil((route) {
                                        final name = route.settings.name;
                                        return (name != null &&
                                                name.endsWith('_dashboard')) ||
                                            route.isFirst;
                                      });
                                    },
                                    child: Text(
                                      'KEMBALI KE DASHBOARD',
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w800,
                                        color: Colors.blueAccent,
                                        fontSize: 14,
                                      ),
                                    ),
                                  ),
                                ] else ...[
                                  SizedBox(
                                    width: double.infinity,
                                    height: 64,
                                    child: ElevatedButton(
                                      onPressed: () {
                                        di.sl<SyncBloc>().add(
                                          const SyncDataRequested(force: true),
                                        );
                                        Navigator.of(context).popUntil((route) {
                                          final name = route.settings.name;
                                          return (name != null &&
                                                  name.endsWith(
                                                    '_dashboard',
                                                  )) ||
                                              route.isFirst;
                                        });
                                      },
                                      style: ElevatedButton.styleFrom(
                                        backgroundColor: const Color(
                                          0xFF1E293B,
                                        ),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(
                                            20,
                                          ),
                                        ),
                                        elevation: 0,
                                      ),
                                      child: Text(
                                        'KEMBALI KE DASHBOARD',
                                        style: GoogleFonts.outfit(
                                          fontWeight: FontWeight.w900,
                                          fontSize: 16,
                                          color: Colors.white,
                                          letterSpacing: 0.5,
                                        ),
                                      ),
                                    ),
                                  ),
                                ],
                                const SizedBox(height: 40),
                              ],
                            ),
                          ),
                        ],
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

  Widget _dot(Color color, double size) {
    return FadeTransition(
      opacity: _opacityAnimation,
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          color: color.withOpacity(0.2),
          shape: BoxShape.circle,
        ),
      ),
    );
  }

  Widget _summaryRow(
    String label,
    String value, {
    bool isBold = false,
    Color? valueColor,
  }) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: GoogleFonts.outfit(
            color: Colors.blueGrey,
            fontWeight: FontWeight.w600,
            fontSize: 14,
          ),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Text(
            value,
            textAlign: TextAlign.end,
            style: GoogleFonts.outfit(
              fontWeight: isBold ? FontWeight.w900 : FontWeight.w700,
              color: valueColor ?? const Color(0xFF1E293B),
              fontSize: isBold ? 18 : 14,
            ),
          ),
        ),
      ],
    );
  }
}
