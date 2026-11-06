import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/finance/data/services/finance_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/core/utils/dialog_utils.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:shimmer/shimmer.dart';

class CashierDashboardPage extends StatefulWidget {
  const CashierDashboardPage({super.key});

  @override
  State<CashierDashboardPage> createState() => _CashierDashboardPageState();
}

class _CashierDashboardPageState extends State<CashierDashboardPage> {
  final FinanceApiService _financeService = di.sl<FinanceApiService>();
  final NumberFormat _currencyFormat =
      NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);

  bool _isLoading = false;
  List<dynamic> _pendingPayments = [];

  @override
  void initState() {
    super.initState();
    _fetchPendingPayments();
  }

  Future<void> _fetchPendingPayments() async {
    setState(() => _isLoading = true);
    try {
      final payments = await _financeService.getPendingPayments();
      setState(() => _pendingPayments = payments);
    } catch (e) {
      if (mounted)
        SnackBarUtils.showError(context, 'Gagal memuat pembayaran pending');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _approvePayment(String id) async {
    DialogUtils.showLoading(context);
    try {
      await _financeService.approvePayment(id);
      if (mounted) Navigator.pop(context); // close loading
      if (mounted)
        SnackBarUtils.showSuccess(context, 'Pembayaran berhasil disetujui');
      _fetchPendingPayments();
    } catch (e) {
      if (mounted) Navigator.pop(context); // close loading
      if (mounted)
        SnackBarUtils.showError(context, 'Gagal menyetujui pembayaran');
    }
  }

  void _scanBarcode() async {
    final result = await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => const BarcodeScannerPage(),
      ),
    );

    if (result != null && result is String) {
      if (result.startsWith('SETTLE|ITEM|')) {
        final paymentId = result.split('|')[2];
        _showApprovalConfirmDialog(paymentId);
      } else {
        SnackBarUtils.showError(
            context, 'Format barcode tidak valid atau bukan QR Setoran Kasir');
      }
    }
  }

  void _showApprovalConfirmDialog(String paymentId) {
    final payment = _pendingPayments.firstWhere((p) => p['id'] == paymentId,
        orElse: () => null);

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Konfirmasi Verifikasi',
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w900, color: AppColors.blueHRDark)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
                'Apakah setoran tunai ini sudah divalidasi dan diterima kasir?',
                style: GoogleFonts.outfit(fontSize: 14)),
            if (payment != null) ...[
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                    color: Colors.green.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12)),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Nominal:',
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w700, color: Colors.green)),
                    Text(
                        _currencyFormat
                            .format((payment['amount'] ?? 0).toDouble()),
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w900,
                            color: Colors.green,
                            fontSize: 16)),
                  ],
                ),
              ),
            ] else ...[
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                    color: Colors.orange.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12)),
                child: Text(
                    'Data tidak ditemukan di antrean Anda saat ini. Verifikasi tetap akan diteruskan ke server.',
                    style:
                        GoogleFonts.outfit(color: Colors.orange, fontSize: 11)),
              ),
            ]
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text('BATAL',
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w800, color: Colors.grey)),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _approvePayment(paymentId);
            },
            style: ElevatedButton.styleFrom(
                backgroundColor: Colors.green,
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12))),
            child: Text('SETUJUI',
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w800, color: Colors.white)),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      body: CustomScrollView(
        slivers: [
          _buildAppBar(),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildQuickAccess(),
                  const SizedBox(height: 24),
                  Text(
                    'ANTREAN SETORAN',
                    style: GoogleFonts.outfit(
                        fontSize: 12,
                        fontWeight: FontWeight.w900,
                        color: Colors.blueGrey,
                        letterSpacing: 1),
                  ),
                  const SizedBox(height: 16),
                  _buildPendingPaymentsList(),
                ],
              ),
            ),
          )
        ],
      ),
    );
  }

  Widget _buildAppBar() {
    return SliverAppBar(
      expandedHeight: 110,
      floating: false,
      pinned: true,
      backgroundColor: Colors.blueAccent,
      elevation: 0,
      iconTheme: const IconThemeData(color: Colors.white),
      flexibleSpace: FlexibleSpaceBar(
        background: Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [Colors.blueAccent, Colors.blue.shade800],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
        ),
        titlePadding: const EdgeInsets.only(left: 60, bottom: 16),
        title: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'CASHIER DASHBOARD',
              style: GoogleFonts.outfit(
                  fontSize: 10,
                  fontWeight: FontWeight.w900,
                  color: Colors.white70,
                  letterSpacing: 2),
            ),
            Text(
              'Validasi Setoran',
              style: GoogleFonts.outfit(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Colors.white),
            ),
          ],
        ),
      ),
      actions: [
        IconButton(
          icon: const Icon(Icons.qr_code_scanner_rounded, color: Colors.white),
          onPressed: _scanBarcode,
          tooltip: 'Scan Barcode',
        ),
        const SizedBox(width: 8),
      ],
    );
  }

  Widget _buildQuickAccess() {
    return InkWell(
      onTap: _scanBarcode,
      borderRadius: BorderRadius.circular(24),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          border:
              Border.all(color: Colors.blueAccent.withOpacity(0.3), width: 2),
          boxShadow: [
            BoxShadow(
                color: Colors.blueAccent.withOpacity(0.1),
                blurRadius: 20,
                offset: const Offset(0, 10)),
          ],
        ),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.blueAccent.withOpacity(0.1),
                shape: BoxShape.circle,
              ),
              child: const Icon(Icons.qr_code_scanner_rounded,
                  size: 40, color: Colors.blueAccent),
            ),
            const SizedBox(height: 16),
            Text(
              'SCAN BARCODE SETORAN',
              style: GoogleFonts.outfit(
                  fontSize: 14,
                  fontWeight: FontWeight.w900,
                  color: const Color(0xFF1E293B)),
            ),
            const SizedBox(height: 4),
            Text(
              'Gunakan kamera untuk validasi instan',
              style: GoogleFonts.outfit(
                  fontSize: 11,
                  color: Colors.blueGrey,
                  fontWeight: FontWeight.w500),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSkeleton() {
    return Shimmer.fromColors(
      baseColor: Colors.grey.shade300,
      highlightColor: Colors.grey.shade100,
      child: Column(
        children: List.generate(
            3,
            (index) => Container(
                  margin: const EdgeInsets.only(bottom: 12),
                  height: 100,
                  decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(20)),
                )),
      ),
    );
  }

  Widget _buildPendingPaymentsList() {
    if (_isLoading) {
      return _buildSkeleton();
    }

    if (_pendingPayments.isEmpty) {
      return Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(vertical: 60),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: const Color(0xFFE2E8F0)),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.check_circle_outline_rounded,
                size: 60, color: Colors.green.withOpacity(0.5)),
            const SizedBox(height: 16),
            Text('Semua Setoran Tervalidasi',
                style: GoogleFonts.outfit(
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                    color: Colors.blueGrey)),
            const SizedBox(height: 4),
            Text('Tidak ada antrean pending saat ini',
                style: GoogleFonts.outfit(fontSize: 12, color: Colors.grey)),
          ],
        ),
      );
    }

    return ListView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: _pendingPayments.length,
      itemBuilder: (context, index) {
        final payment = _pendingPayments[index];
        final amount = (payment['amount'] ?? 0.0).toDouble();
        final String method =
            (payment['payment_method'] ?? 'UNK').toUpperCase();
        final isCash = method == 'CASH';
        final salesOrder = payment['sales_order'];
        final salesTransaction = payment['sales_transaction'];
        final String receiptNo = payment['receipt_no'] ??
            (salesOrder != null
                ? salesOrder['order_no']
                : (salesTransaction != null
                    ? salesTransaction['receipt_no']
                    : 'Transaksi'));

        // Extracting employee name
        String employeeName = 'Salesman';
        if (salesOrder != null && salesOrder['employee'] != null) {
          employeeName =
              "${salesOrder['employee']['first_name']} ${salesOrder['employee']['last_name']}"
                  .trim();
        } else if (salesTransaction != null &&
            salesTransaction['employee'] != null) {
          employeeName =
              "${salesTransaction['employee']['first_name']} ${salesTransaction['employee']['last_name']}"
                  .trim();
        }

        return Container(
          margin: const EdgeInsets.only(bottom: 12),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: const Color(0xFFE2E8F0)),
            boxShadow: [
              BoxShadow(
                  color: Colors.black.withOpacity(0.02),
                  blurRadius: 10,
                  offset: const Offset(0, 4)),
            ],
          ),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: isCash
                        ? Colors.green.withOpacity(0.1)
                        : Colors.blue.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Icon(
                      isCash
                          ? Icons.payments_rounded
                          : Icons.account_balance_wallet_rounded,
                      color: isCash ? Colors.green : Colors.blueAccent),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(receiptNo,
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800,
                              fontSize: 13,
                              color: const Color(0xFF1E293B))),
                      const SizedBox(height: 2),
                      Text(employeeName,
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w600,
                              fontSize: 11,
                              color: Colors.blueGrey)),
                      const SizedBox(height: 6),
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 6, vertical: 2),
                        decoration: BoxDecoration(
                            color: const Color(0xFFF1F5F9),
                            borderRadius: BorderRadius.circular(4)),
                        child: Text(method,
                            style: GoogleFonts.outfit(
                                fontSize: 9,
                                fontWeight: FontWeight.w900,
                                color: Colors.blueGrey)),
                      ),
                    ],
                  ),
                ),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text(_currencyFormat.format(amount),
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w900,
                            fontSize: 14,
                            color: isCash ? Colors.green : Colors.blueAccent)),
                    const SizedBox(height: 8),
                    if (method.contains('VA') ||
                        method.contains('TRANSFER') ||
                        method.contains('MIDTRANS'))
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 12, vertical: 6),
                        decoration: BoxDecoration(
                          color: Colors.blue.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(8),
                          border:
                              Border.all(color: Colors.blue.withOpacity(0.3)),
                        ),
                        child: Text('OTOMATIS',
                            style: GoogleFonts.outfit(
                                fontSize: 10,
                                fontWeight: FontWeight.w900,
                                color: Colors.blueAccent)),
                      )
                    else
                      InkWell(
                        onTap: () => _showApprovalConfirmDialog(payment['id']),
                        borderRadius: BorderRadius.circular(8),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 16, vertical: 6),
                          decoration: BoxDecoration(
                            color: Colors.green,
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text('VERIFIKASI',
                              style: GoogleFonts.outfit(
                                  fontSize: 10,
                                  fontWeight: FontWeight.w900,
                                  color: Colors.white)),
                        ),
                      ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

class BarcodeScannerPage extends StatefulWidget {
  const BarcodeScannerPage({super.key});

  @override
  State<BarcodeScannerPage> createState() => _BarcodeScannerPageState();
}

class _BarcodeScannerPageState extends State<BarcodeScannerPage> {
  final MobileScannerController controller = MobileScannerController(
    detectionSpeed: DetectionSpeed.noDuplicates,
  );
  bool _isScanned = false;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        title: Text('Scan Barcode Setoran',
            style: GoogleFonts.outfit(
                color: Colors.white, fontWeight: FontWeight.bold)),
        backgroundColor: Colors.transparent,
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
      ),
      body: Stack(
        children: [
          MobileScanner(
            controller: controller,
            onDetect: (capture) {
              if (_isScanned) return;
              final List<Barcode> barcodes = capture.barcodes;
              for (final barcode in barcodes) {
                if (barcode.rawValue != null) {
                  _isScanned = true;
                  Navigator.pop(context, barcode.rawValue);
                  break;
                }
              }
            },
          ),
          Center(
            child: Container(
              width: 250,
              height: 150,
              decoration: BoxDecoration(
                border: Border.all(color: Colors.green, width: 3),
                borderRadius: BorderRadius.circular(16),
              ),
            ),
          ),
          Positioned(
            bottom: 40,
            left: 0,
            right: 0,
            child: Text(
              'Arahkan kamera ke barcode setoran dari aplikasi salesman',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(color: Colors.white, fontSize: 12),
            ),
          ),
        ],
      ),
    );
  }
}
