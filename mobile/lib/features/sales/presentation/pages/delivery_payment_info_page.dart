import 'package:cached_network_image/cached_network_image.dart';
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'visit_success_page.dart';
import 'package:intl/intl.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:drift/drift.dart' hide Column;
import 'package:dio/dio.dart';
import 'package:hris_app/core/widgets/qris_sandbox_warning.dart';

class DeliveryPaymentInfoPage extends StatefulWidget {
  final Map<String, dynamic> paymentData;
  final String storeName;
  final double amount;
  final Map<String, dynamic>? transaction;
  final String? localId;
  final Uint8List? invoicePdfBytes;

  const DeliveryPaymentInfoPage({
    super.key,
    required this.paymentData,
    required this.storeName,
    required this.amount,
    this.transaction,
    this.localId,
    this.invoicePdfBytes,
  });

  @override
  State<DeliveryPaymentInfoPage> createState() =>
      _DeliveryPaymentInfoPageState();
}

class _DeliveryPaymentInfoPageState extends State<DeliveryPaymentInfoPage> {
  late Map<String, dynamic> _currentPaymentData;
  Timer? _pollingTimer;
  final AppDatabase _db = di.sl<AppDatabase>();
  final ApiClient _api = di.sl<ApiClient>();
  bool _isPaid = false;

  @override
  void initState() {
    super.initState();
    _currentPaymentData = Map<String, dynamic>.from(widget.paymentData);
    _startPolling();
  }

  @override
  void dispose() {
    _pollingTimer?.cancel();
    super.dispose();
  }

  void _startPolling() {
    _pollingTimer = Timer.periodic(const Duration(seconds: 3), (timer) async {
      if (!mounted) return;

      // 1. Poll local DB for updated payment info or official receipt no
      String? officialReceiptNo;
      if (widget.localId != null) {
        final checkTx = await (_db.select(_db.localTransactions)
              ..where((t) => t.localId.equals(widget.localId!)))
            .getSingleOrNull();
        if (checkTx != null) {
          officialReceiptNo = checkTx.receiptNo;

          final qrisUrl = _currentPaymentData['midtrans_qris_url'];
          final vaNumber = _currentPaymentData['midtrans_va_number'] ??
              _currentPaymentData['midtrans_bill_key'];

          if ((_currentPaymentData['payment_method'] == 'QRIS' &&
                  qrisUrl == null) ||
              (_currentPaymentData['payment_method'] == 'VA' &&
                  vaNumber == null)) {
            if (checkTx.midtransQrisUrl != null ||
                checkTx.midtransVaNumber != null ||
                checkTx.midtransBillKey != null) {
              setState(() {
                _currentPaymentData = {
                  'payment_method': _currentPaymentData['payment_method'],
                  'midtrans_qris_url': checkTx.midtransQrisUrl,
                  'midtrans_va_number': checkTx.midtransVaNumber,
                  'midtrans_bank': checkTx.midtransBank,
                  'midtrans_bill_key': checkTx.midtransBillKey,
                  'midtrans_biller_code': checkTx.midtransBillerCode,
                };
              });
            }
          }
        }
      }

      // 2. Poll server for payment status using the official receipt number
      final String? receiptNo =
          officialReceiptNo ?? widget.transaction?['receipt_no'];
      if (receiptNo != null && receiptNo.isNotEmpty && !_isPaid) {
        try {
          final response =
              await _api.client.get('sales/transactions/receipt/$receiptNo');
          if (response.statusCode == 200 && response.data != null) {
            final data =
                response.data is Map && response.data.containsKey('data')
                    ? response.data['data']
                    : response.data;
            final String? payStatus =
                data['payment_status']?.toString().toUpperCase();
            final String? status = data['status']?.toString().toUpperCase();

            if (payStatus == 'PAID' ||
                payStatus == 'LUNAS' ||
                status == 'VERIFIED') {
              _isPaid = true;
              _pollingTimer?.cancel();

              // Update local DB status to synced
              if (widget.localId != null) {
                await (_db.update(_db.localTransactions)
                      ..where((t) => t.localId.equals(widget.localId!)))
                    .write(const LocalTransactionsCompanion(
                  syncStatus: Value('synced'),
                ));
              }

              // Update the passed transaction object
              final updatedTx = Map<String, dynamic>.from(widget.transaction!);
              updatedTx['status'] = 'LUNAS';
              updatedTx['receipt_no'] =
                  receiptNo; // Update to the official server-generated receipt number

              if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(SnackBar(
                  content: Text('✓ Pembayaran Berhasil Diterima!',
                      style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
                  backgroundColor: Colors.green,
                ));
                Navigator.of(context).pushReplacement(
                  MaterialPageRoute(
                    builder: (_) => VisitSuccessPage(
                      transaction: updatedTx,
                      localId: widget.localId,
                      invoicePdfBytes: widget.invoicePdfBytes,
                    ),
                  ),
                );
              }
            }
          }
        } catch (e) {
          // Gracefully ignore 404/network errors while transaction is syncing in background
          if (e is DioException && e.response?.statusCode == 404) {
            // Normal: transaction not yet synced to the server, skip printing stacktrace
          } else {
            print('DEBUG PAYMENT INFO POLLING ERROR: $e');
          }
        }
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final currencyFormat =
        NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);
    final paymentMethod = _currentPaymentData['payment_method'] ?? 'PENDING';
    final qrisUrl = _currentPaymentData['midtrans_qris_url'];
    final vaNumber = _currentPaymentData['midtrans_va_number'] ??
        _currentPaymentData['midtrans_bill_key'];
    final bank = (_currentPaymentData['midtrans_bank'] ?? 'Bank')
        .toString()
        .toUpperCase();
    final billerCode = _currentPaymentData['midtrans_biller_code'];

    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.close_rounded, color: Color(0xFF334155)),
          onPressed: () {
            if (widget.transaction != null && widget.localId != null) {
              Navigator.of(context).pushReplacement(
                MaterialPageRoute(
                  builder: (_) => VisitSuccessPage(
                    transaction: widget.transaction!,
                    localId: widget.localId!,
                    invoicePdfBytes: widget.invoicePdfBytes,
                  ),
                ),
              );
            } else {
              Navigator.pop(context);
            }
          },
        ),
        title: Text('INFO PEMBAYARAN',
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w900, color: const Color(0xFF1E293B))),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Container(
              width: double.infinity,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(32),
                boxShadow: [
                  BoxShadow(
                      color: Colors.black.withOpacity(0.05),
                      blurRadius: 20,
                      offset: const Offset(0, 10))
                ],
              ),
              child: Column(
                children: [
                  const SizedBox(height: 32),
                  if (paymentMethod == 'QRIS' && qrisUrl != null) ...[
                    _buildQRISSection(context, qrisUrl),
                  ] else if (paymentMethod == 'VA' && vaNumber != null) ...[
                    _buildVASection(context, bank, vaNumber, billerCode),
                  ] else ...[
                    const SizedBox(height: 40),
                    const CircularProgressIndicator(color: Colors.blueAccent),
                    const SizedBox(height: 24),
                    Text('Menyiapkan Barcode / Info Pembayaran...',
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.bold,
                            color: Colors.blueGrey)),
                    Text('Mohon tunggu sebentar...',
                        style: GoogleFonts.outfit(
                            fontSize: 12, color: Colors.grey)),
                    const SizedBox(height: 40),
                  ],
                  const SizedBox(height: 32),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 32),
                    child: Column(
                      children: [
                        _buildRow('Toko', widget.storeName, isBold: true),
                        const Divider(height: 32),
                        _buildRow('Metode Bayar', paymentMethod),
                        const SizedBox(height: 12),
                        _buildRow('Total Tagihan',
                            currencyFormat.format(widget.amount),
                            isPrice: true),
                      ],
                    ),
                  ),
                  const SizedBox(height: 40),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      color: Colors.blueAccent.withOpacity(0.05),
                      borderRadius: const BorderRadius.only(
                          bottomLeft: Radius.circular(32),
                          bottomRight: Radius.circular(32)),
                    ),
                    child: Column(
                      children: [
                        Text(
                          paymentMethod == 'QRIS'
                              ? 'SCAN QRIS DI ATAS'
                              : 'TRANSFER KE NOMOR DI ATAS',
                          style: GoogleFonts.outfit(
                              fontSize: 12,
                              fontWeight: FontWeight.w900,
                              color: Colors.blueAccent,
                              letterSpacing: 1),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          paymentMethod == 'QRIS'
                              ? 'Pelanggan dapat memindai kode QR di atas menggunakan aplikasi perbankan atau e-wallet.'
                              : 'Silakan lakukan transfer ke nomor Virtual Account di atas sebelum meninggalkan toko.',
                          textAlign: TextAlign.center,
                          style: GoogleFonts.outfit(
                              fontSize: 12,
                              color: Colors.blueGrey,
                              fontWeight: FontWeight.w600),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),
            SizedBox(
              width: double.infinity,
              height: 64,
              child: ElevatedButton.icon(
                onPressed: () => _shareToWhatsApp(
                    context, paymentMethod, vaNumber, bank, billerCode),
                icon: const Icon(Icons.share_rounded, color: Colors.white),
                label: Text('KIRIM KE WHATSAPP',
                    style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w900, color: Colors.white)),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF25D366),
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20)),
                ),
              ),
            ),
            const SizedBox(height: 16),
            SizedBox(
              width: double.infinity,
              height: 64,
              child: OutlinedButton(
                onPressed: () {
                  if (widget.transaction != null && widget.localId != null) {
                    Navigator.of(context).pushReplacement(
                      MaterialPageRoute(
                        builder: (_) => VisitSuccessPage(
                          transaction: widget.transaction!,
                          localId: widget.localId!,
                          invoicePdfBytes: widget.invoicePdfBytes,
                        ),
                      ),
                    );
                  } else {
                    Navigator.pop(context);
                  }
                },
                style: OutlinedButton.styleFrom(
                  side: const BorderSide(color: Color(0xFFE2E8F0), width: 2),
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20)),
                ),
                child: Text('TUTUP',
                    style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w900,
                        color: const Color(0xFF1E293B))),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildQRISSection(BuildContext context, String url) {
    return Container(
      padding: const EdgeInsets.all(24),
      margin: const EdgeInsets.symmetric(horizontal: 24),
      decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: Colors.grey.shade200)),
      child: Column(children: [
        if (url.startsWith('http')) ...[
          QrisSandboxWarning(
              qrisUrl: url, margin: const EdgeInsets.only(bottom: 16)),
        ],
        Builder(builder: (ctx) {
          final String displayUrl = url.startsWith('http')
              ? '${AppConstants.baseUrl}sales/qris-image?url=${Uri.encodeComponent(url)}'
              : url;
          return url.startsWith('http')
              ? CachedNetworkImage(imageUrl: displayUrl, height: 240,
                  width: 240,
                  fit: BoxFit.contain, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,)
              : QrImageView(data: url, version: QrVersions.auto, size: 240.0);
        }),
        const SizedBox(height: 12),
        Text('SCAN QRIS UNTUK BAYAR',
            style: GoogleFonts.outfit(
                fontSize: 14,
                fontWeight: FontWeight.w900,
                color: Colors.blueAccent)),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            color: const Color(0xFFF8FAFC),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Colors.blueGrey.withOpacity(0.1)),
          ),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  url.toString(),
                  style:
                      GoogleFonts.outfit(fontSize: 11, color: Colors.blueGrey),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              IconButton(
                icon: const Icon(Icons.copy_rounded,
                    size: 18, color: Colors.blueAccent),
                onPressed: () {
                  Clipboard.setData(ClipboardData(text: url.toString()));
                  ScaffoldMessenger.of(context).showSnackBar(SnackBar(
                    content: Text('Tautan QRIS Disalin!',
                        style: GoogleFonts.outfit()),
                    backgroundColor: Colors.green,
                  ));
                },
              ),
            ],
          ),
        ),
      ]),
    );
  }

  Widget _buildVASection(
      BuildContext context, String bank, String va, String? biller) {
    return Container(
      padding: const EdgeInsets.all(24),
      margin: const EdgeInsets.symmetric(horizontal: 24),
      decoration: BoxDecoration(
          color: Colors.blueAccent.withOpacity(0.05),
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: Colors.blueAccent.withOpacity(0.2))),
      child: Column(children: [
        Text('NOMOR VIRTUAL ACCOUNT ($bank)',
            style: GoogleFonts.outfit(
                fontSize: 10,
                fontWeight: FontWeight.w900,
                color: Colors.blueAccent,
                letterSpacing: 1)),
        const SizedBox(height: 12),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(va,
                style: GoogleFonts.outfit(
                    fontSize: 24,
                    fontWeight: FontWeight.w900,
                    color: const Color(0xFF1E293B),
                    letterSpacing: 1)),
            IconButton(
              icon: const Icon(Icons.copy_rounded,
                  color: Colors.blueGrey, size: 20),
              onPressed: () {
                Clipboard.setData(ClipboardData(text: va));
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
                    content: Text('Nomor VA disalin!'),
                    backgroundColor: Colors.green));
              },
            ),
          ],
        ),
        if (biller != null) ...[
          const SizedBox(height: 4),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text('Kode Biller: $biller',
                  style: GoogleFonts.outfit(
                      fontSize: 14,
                      fontWeight: FontWeight.w700,
                      color: Colors.blueGrey)),
              IconButton(
                icon: const Icon(Icons.copy_rounded,
                    color: Colors.blueGrey, size: 16),
                onPressed: () {
                  Clipboard.setData(ClipboardData(text: biller));
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
                      content: Text('Kode Biller disalin!'),
                      backgroundColor: Colors.green));
                },
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
              ),
            ],
          ),
        ],
      ]),
    );
  }

  Widget _buildRow(String label, String value,
      {bool isBold = false, bool isPrice = false}) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label,
            style: GoogleFonts.outfit(
                fontSize: 14,
                color: Colors.blueGrey,
                fontWeight: FontWeight.w600)),
        Text(
          value,
          style: GoogleFonts.outfit(
            fontSize: isPrice ? 18 : 14,
            fontWeight: isBold || isPrice ? FontWeight.w900 : FontWeight.w700,
            color: isPrice ? Colors.blueAccent : const Color(0xFF1E293B),
          ),
        ),
      ],
    );
  }

  Future<void> _shareToWhatsApp(BuildContext context, String method, String? va,
      String bank, String? biller) async {
    final currencyFormat =
        NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);
    String message =
        "Halo *${widget.storeName}*,\n\nBerikut informasi pembayaran untuk pengiriman hari ini:\n";
    message += "💰 *Total Tagihan:* ${currencyFormat.format(widget.amount)}\n";
    message += "💳 *Metode Pembayaran:* $method\n";

    if (method == 'VA' && va != null) {
      message += "\n🏦 *Informasi Pembayaran ($bank):*\n";
      message += "Nomor: `$va`";
      if (biller != null) message += "\nKode Biller: `$biller`";
    }

    message += "\n\nTerima kasih!";

    final whatsappUrl =
        Uri.parse("https://wa.me/?text=${Uri.encodeComponent(message)}");
    if (await canLaunchUrl(whatsappUrl)) {
      await launchUrl(whatsappUrl, mode: LaunchMode.externalApplication);
    }
  }
}
