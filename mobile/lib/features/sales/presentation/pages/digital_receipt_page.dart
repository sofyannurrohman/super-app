import 'package:cached_network_image/cached_network_image.dart';
import 'package:barcode_widget/barcode_widget.dart';
import 'package:flutter/material.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:intl/intl.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:url_launcher/url_launcher.dart';
import 'dart:async';
import 'dart:convert';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:share_plus/share_plus.dart';
import 'dart:io' show File, Directory;
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:dio/dio.dart';
import 'package:hris_app/core/widgets/qris_sandbox_warning.dart';
import './tempo_payment_page.dart';

class DigitalReceiptPage extends StatefulWidget {
  final Map<String, dynamic> transaction;
  final String? localId;

  const DigitalReceiptPage({
    super.key,
    required this.transaction,
    this.localId,
  });

  @override
  State<DigitalReceiptPage> createState() => _DigitalReceiptPageState();
}

class _DigitalReceiptPageState extends State<DigitalReceiptPage> {
  late Map<String, dynamic> _currentTrx;
  Timer? _refreshTimer;
  final AppDatabase _db = di.sl<AppDatabase>();

  @override
  void initState() {
    super.initState();
    _currentTrx = widget.transaction;
    if (widget.localId != null) {
      _fetchLatestData();
      // Poll every 5 seconds to catch sync updates (QRIS/VA)
      _refreshTimer = Timer.periodic(
        const Duration(seconds: 5),
        (_) => _fetchLatestData(),
      );
    }
  }

  @override
  void dispose() {
    _refreshTimer?.cancel();
    super.dispose();
  }

  Future<void> _fetchLatestData() async {
    if (widget.localId == null) return;

    final updated = await (_db.select(
      _db.localTransactions,
    )..where((t) => t.localId.equals(widget.localId!))).getSingleOrNull();

    final updatedItems = await (_db.select(
      _db.localTransactionItems,
    )..where((t) => t.transactionLocalId.equals(widget.localId!))).get();

    if (updated != null && mounted) {
      setState(() {
        _currentTrx = {
          ..._currentTrx,
          'receipt_no': updated.receiptNo ?? _currentTrx['receipt_no'],
          'status': updated.syncStatus == 'synced'
              ? 'BERHASIL (TERSIMPAN)'
              : (_currentTrx['payment_method'] == 'CASH'
                    ? 'LUNAS (CASH)'
                    : 'PENDING (MENUNGGU SINYAL)'),
          'midtrans_id': updated.midtransId,
          'midtrans_qris_url': updated.midtransQrisUrl,
          'midtrans_va_number': updated.midtransVaNumber,
          'midtrans_bank': updated.midtransBank,
          'midtrans_bill_key': updated.midtransBillKey,
          'midtrans_biller_code': updated.midtransBillerCode,
          'dp_payment_method':
              updated.dpPaymentMethod ?? _currentTrx['dp_payment_method'],
          'dp_amount': updated.paidAmount ?? _currentTrx['dp_amount'],
          // Preserve signature — carried from initial widget.transaction
          'signature_image_url': _currentTrx['signature_image_url'],
          'receipt_path': updated.receiptPath ?? _currentTrx['receipt_path'],
          'items': updatedItems
              .map(
                (i) => {
                  'product_name': i.productName,
                  'quantity': i.quantity,
                  'unit': i.unit,
                  'pieces_per_unit': i.piecesPerUnit,
                  'price': i.price,
                  'subtotal': (i.quantity ?? 0) * (i.price ?? 0),
                },
              )
              .toList(),
        };
      });
    }
  }

  Widget _buildReceiptImage(
    String path, {
    double? height = 200,
    double? width = double.infinity,
    BoxFit fit = BoxFit.cover,
  }) {
    if (path.startsWith('data:')) {
      try {
        final bytes = base64Decode(path.split(',').last);
        return Image.memory(bytes, height: height, width: width, fit: fit);
      } catch (e) {
        print('Error decoding base64 in receipt preview: $e');
      }
    }

    String fullUrl = path;
    if (!path.startsWith('http') &&
        !path.startsWith('data:') &&
        !path.startsWith('blob:')) {
      const baseUrl = AppConstants.baseUrl;
      final cleanBase = baseUrl.endsWith('/')
          ? baseUrl.substring(0, baseUrl.length - 1)
          : baseUrl;
      final cleanUrl = path.startsWith('/') ? path : '/$path';
      final hostOnly = cleanBase
          .replaceAll(RegExp(r'/api/v1$'), '')
          .replaceAll(RegExp(r'/api$'), '');
      fullUrl = '$hostOnly$cleanUrl';
    }

    if (!kIsWeb && !path.startsWith('http') && File(path).existsSync()) {
      return Image.file(File(path), height: height, width: width, fit: fit);
    }

    return CachedNetworkImage(
      imageUrl: fullUrl,
      height: height,
      width: width,
      fit: fit,
      placeholder: (context, url) =>
          const Center(child: CircularProgressIndicator()),
      errorWidget: (context, url, error) => const Icon(Icons.error),
      memCacheWidth: 800,
    );
  }

  void _showImagePreviewDialog(String path) {
    showDialog(
      context: context,
      builder: (_) => Dialog(
        backgroundColor: Colors.black87,
        insetPadding: EdgeInsets.zero,
        child: Stack(
          alignment: Alignment.center,
          children: [
            SizedBox(
              width: double.infinity,
              height: double.infinity,
              child: InteractiveViewer(
                panEnabled: true,
                minScale: 0.5,
                maxScale: 4.0,
                child: _buildReceiptImage(
                  path,
                  height: null,
                  width: null,
                  fit: BoxFit.contain,
                ),
              ),
            ),
            Positioned(
              top: 40,
              right: 20,
              child: Container(
                decoration: const BoxDecoration(
                  color: Colors.black54,
                  shape: BoxShape.circle,
                ),
                child: IconButton(
                  icon: const Icon(
                    Icons.close_rounded,
                    color: Colors.white,
                    size: 28,
                  ),
                  onPressed: () => Navigator.pop(context),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showTempoPaymentDialog() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => TempoPaymentPage(
          transactionData: _currentTrx,
          onSuccess: (method, amount, signatureUrl) async {
            Navigator.pop(context); // Close the payment page

            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Pembayaran berhasil dicatat!'),
                backgroundColor: Colors.green,
              ),
            );

            setState(() {
              _currentTrx['payment_method'] = method;
              _currentTrx['payment_status'] = 'PAID';
              if (signatureUrl != null) {
                _currentTrx['signature_image_url'] = signatureUrl;
              }
            });

            // Launch WhatsApp automatically to share receipt
            await _shareToWhatsApp(context);
          },
        ),
      ),
    );
  }

  Future<void> _shareToWhatsApp(BuildContext context) async {
    final storeName = _currentTrx['store']?['name'] ?? 'Toko';
    final receiptNo = _currentTrx['receipt_no'] ?? 'NO-DATA';
    final amount = (_currentTrx['total_amount'] ?? 0.0).toDouble();
    final paymentMethod = _currentTrx['payment_method'] ?? 'CASH';
    final vaNumber =
        (_currentTrx['midtrans_va_number'] != null &&
            _currentTrx['midtrans_va_number'].toString().isNotEmpty)
        ? _currentTrx['midtrans_va_number']
        : (_currentTrx['midtrans_bill_key'] ?? '');
    final bank = _currentTrx['midtrans_bank'] ?? '';

    final currencyFormat = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    String message =
        "Halo *$storeName*,\n\nTerima kasih telah berbelanja di *Perusahaan Demo*.\n\nBerikut ringkasan pesanan Anda:\n";
    message += "📌 *No. Nota:* $receiptNo\n";

    if (_currentTrx['items'] != null) {
      message += "📦 *Item Pesanan:*\n";
      for (var item in (_currentTrx['items'] as List)) {
        final itemMap = Map<String, dynamic>.from(item as Map);
        message +=
            "- ${itemMap['product_name'] ?? (itemMap['product'] as Map?)?['name'] ?? itemMap['name'] ?? 'Produk'} (x${_formatItemQty(itemMap)})\n";
      }
    }

    message += "💰 *Total Tagihan:* ${currencyFormat.format(amount)}\n";
    message += "💳 *Metode Pembayaran:* $paymentMethod\n";

    if (vaNumber.isNotEmpty) {
      message += "\n🏦 *Informasi Pembayaran (${bank.toUpperCase()}):*\n";
      message += "Nomor: `$vaNumber`";
      if (_currentTrx['midtrans_biller_code'] != null) {
        message += "\nKode Biller: `${_currentTrx['midtrans_biller_code']}`";
      }
      message += "\n\n_Silakan lakukan pembayaran sesuai nominal di atas._\n";
    }

    if (amount == 0 && _currentTrx['notes'] != null) {
      message += "\n📝 *Catatan Kunjungan:* ${_currentTrx['notes']}\n";
    }

    message += "\nTerima kasih atas kerja samanya!";

    final whatsappUrl = Uri.parse(
      "https://wa.me/?text=${Uri.encodeComponent(message)}",
    );

    if (await canLaunchUrl(whatsappUrl)) {
      await launchUrl(whatsappUrl, mode: LaunchMode.externalApplication);
    } else {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Gagal membuka WhatsApp.')),
        );
      }
    }
  }

  Future<pw.ImageProvider?> _buildPdfImageProvider(String url) async {
    if (url.isEmpty) return null;
    try {
      if (url.startsWith('data:image')) {
        final base64String = url.split(',').last;
        final bytes = base64Decode(base64String);
        return pw.MemoryImage(bytes);
      } else if (!kIsWeb && !url.startsWith('http') && File(url).existsSync()) {
        final bytes = await File(url).readAsBytes();
        return pw.MemoryImage(bytes);
      } else {
        // Determine full URL
        String fullUrl = url;
        if (!url.startsWith('http')) {
          const baseUrl = AppConstants.baseUrl;
          final cleanBase = baseUrl.endsWith('/')
              ? baseUrl.substring(0, baseUrl.length - 1)
              : baseUrl;
          final cleanUrl = url.startsWith('/') ? url : '/$url';
          final hostOnly = cleanBase
              .replaceAll(RegExp(r'/api/v1$'), '')
              .replaceAll(RegExp(r'/api$'), '');
          fullUrl = '$hostOnly$cleanUrl';
        }

        // Use Dio to fetch image bytes directly to avoid networkImage silent failures
        final dio = di.sl<ApiClient>().client;
        final response = await dio.get(
          fullUrl,
          options: Options(responseType: ResponseType.bytes),
        );

        if (response.statusCode == 200 && response.data != null) {
          return pw.MemoryImage(response.data);
        }
        return null;
      }
    } catch (e) {
      print('Failed to load PDF image via Dio: $e');
      return null;
    }
  }

  Future<void> _generateAndSharePdf(BuildContext context) async {
    // Show a loading indicator
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext context) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const CircularProgressIndicator(),
              const SizedBox(height: 16),
              Text(
                'Menyiapkan Faktur Penjualan...',
                style: GoogleFonts.outfit(fontWeight: FontWeight.bold),
              ),
            ],
          ),
        );
      },
    );

    try {
      final pdf = pw.Document();
      final currencyFormat = NumberFormat.currency(
        locale: 'id_ID',
        symbol: 'Rp ',
        decimalDigits: 0,
      );

      // Extract Signature URLs - must split on 'data:image/' boundary, NOT on commas
      // because Base64 strings themselves contain commas after the header.
      final signatureUrlsStr =
          _currentTrx['signature_image_url'] as String? ?? '';
      String courierSigUrl = '';
      String customerSigUrl = '';
      if (signatureUrlsStr.isNotEmpty) {
        if (signatureUrlsStr.contains('data:image/')) {
          // Split safely by re-splitting on the 'data:image/' prefix
          final rawParts = signatureUrlsStr.split('data:image/');
          final validParts = rawParts
              .where((p) => p.trim().isNotEmpty)
              .toList();
          if (validParts.isNotEmpty)
            courierSigUrl =
                'data:image/${validParts[0].trimRight().replaceAll(RegExp(r',$'), '')}';
          if (validParts.length > 1)
            customerSigUrl =
                'data:image/${validParts[1].trimRight().replaceAll(RegExp(r',$'), '')}';
        } else {
          // Plain URLs separated by comma
          final parts = signatureUrlsStr.split(',');
          if (parts.isNotEmpty) courierSigUrl = parts[0].trim();
          if (parts.length > 1) customerSigUrl = parts[1].trim();
        }
      }

      print(
        'DEBUG PDF: courierSigUrl type=${courierSigUrl.startsWith('data:') ? 'base64' : 'url'} len=${courierSigUrl.length}',
      );
      print(
        'DEBUG PDF: customerSigUrl type=${customerSigUrl.startsWith('data:') ? 'base64' : 'url'} len=${customerSigUrl.length}',
      );

      // Load Images
      final pw.ImageProvider? courierSigImg = await _buildPdfImageProvider(
        courierSigUrl,
      );
      final pw.ImageProvider? customerSigImg = await _buildPdfImageProvider(
        customerSigUrl,
      );

      print('DEBUG PDF: courierSigImg loaded=${courierSigImg != null}');
      print('DEBUG PDF: customerSigImg loaded=${customerSigImg != null}');

      final storeName = _currentTrx['store']?['name'] ?? 'Toko';
      final storeAddress = _currentTrx['store']?['address'] ?? 'Alamat Toko';
      final receiptNo = _currentTrx['receipt_no'] ?? 'NO-DATA';
      final date =
          DateTime.tryParse(
            (_currentTrx['transaction_date'] ?? _currentTrx['created_at'] ?? '')
                .toString(),
          ) ??
          DateTime.now();
      final dateStr = DateFormat('dd MMMM yyyy').format(date);
      final paymentMethod = _currentTrx['payment_method'] ?? 'CASH';
      final totalAmount = (_currentTrx['total_amount'] ?? 0.0).toDouble();

      String pdfPaymentStatusText = 'LUNAS';
      if (paymentMethod.toString().toUpperCase() == 'TEMPO') {
        pdfPaymentStatusText = 'BELUM LUNAS (TEMPO)';
      } else if (paymentMethod.toString().toUpperCase() == 'SALES_ORDER') {
        pdfPaymentStatusText = 'BELUM LUNAS';
      } else if (paymentMethod.toString().toUpperCase() == 'VA' ||
          paymentMethod.toString().toUpperCase() == 'QRIS') {
        final payStatus =
            (_currentTrx['payment_status'] ?? _currentTrx['status'] ?? '')
                .toString()
                .toUpperCase();
        pdfPaymentStatusText = (payStatus == 'PAID' || payStatus == 'SUCCESS')
            ? 'LUNAS'
            : 'MENUNGGU PEMBAYARAN';
      }

      final companyName =
          _currentTrx['employee']?['company']?['name'] ??
          _currentTrx['company']?['name'] ??
          'Perusahaan Demo';
      final companyAddress =
          _currentTrx['company']?['address'] ??
          'Kawasan Industri & Distribusi Sentra Prima';

      final items = _currentTrx['items'] as List? ?? [];
      final bonusItems = _currentTrx['bonus_items'] as List? ?? [];
      print('DEBUG PDF: items count=${items.length}');
      if (items.isNotEmpty)
        print(
          'DEBUG PDF: first item keys=${Map<String, dynamic>.from(items.first as Map).keys.toList()}',
        );
      if (items.isNotEmpty)
        print(
          'DEBUG PDF: first item=${Map<String, dynamic>.from(items.first as Map)}',
        );

      pdf.addPage(
        pw.Page(
          pageFormat: PdfPageFormat.a4,
          margin: const pw.EdgeInsets.symmetric(vertical: 16, horizontal: 24),
          build: (pw.Context context) {
            return pw.Column(
              crossAxisAlignment: pw.CrossAxisAlignment.start,
              children: [
                // Header
                pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                  children: [
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.start,
                      children: [
                        pw.Text(
                          companyName.toUpperCase(),
                          style: pw.TextStyle(
                            fontSize: 9,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.Container(
                          width: 250,
                          child: pw.Text(
                            companyAddress,
                            style: const pw.TextStyle(fontSize: 7),
                          ),
                        ),
                      ],
                    ),
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.end,
                      children: [
                        pw.Text(
                          'FAKTUR PENJUALAN',
                          style: pw.TextStyle(
                            fontSize: 11,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.black,
                          ),
                        ),
                        pw.Text(
                          receiptNo,
                          style: pw.TextStyle(
                            fontSize: 8,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.Text(
                          'Tanggal: $dateStr',
                          style: const pw.TextStyle(fontSize: 7),
                        ),
                      ],
                    ),
                  ],
                ),
                pw.Divider(
                  thickness: 1.0,
                  color: PdfColors.grey400,
                  height: 10,
                ),
                pw.SizedBox(height: 4),

                // Billing Details
                pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                  crossAxisAlignment: pw.CrossAxisAlignment.start,
                  children: [
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.start,
                      children: [
                        pw.Text(
                          'DITAGIHKAN KEPADA:',
                          style: pw.TextStyle(
                            fontSize: 7,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.grey700,
                          ),
                        ),
                        pw.Text(
                          storeName,
                          style: pw.TextStyle(
                            fontSize: 8.5,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.Container(
                          width: 220,
                          child: pw.Text(
                            storeAddress,
                            style: const pw.TextStyle(fontSize: 7),
                            maxLines: 3,
                          ),
                        ),
                      ],
                    ),
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.end,
                      children: [
                        pw.Text(
                          'METODE PEMBAYARAN:',
                          style: pw.TextStyle(
                            fontSize: 7,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.grey700,
                          ),
                        ),
                        pw.Text(
                          paymentMethod.toString().toUpperCase(),
                          style: pw.TextStyle(
                            fontSize: 8.5,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        if (paymentMethod.toString().toUpperCase() == 'TEMPO' &&
                            (_currentTrx['payment_due_date'] != null ||
                                _currentTrx['due_date'] != null))
                          pw.Text(
                            'Jatuh Tempo: ${DateFormat('dd MMM yyyy').format((DateTime.tryParse((_currentTrx['payment_due_date'] ?? _currentTrx['due_date'] ?? '').toString()) ?? DateTime.now()).toLocal())}',
                            style: pw.TextStyle(
                              fontSize: 7,
                              fontWeight: pw.FontWeight.bold,
                              color: PdfColors.black,
                            ),
                          ),
                        pw.Text(
                          'Status: $pdfPaymentStatusText',
                          style: pw.TextStyle(
                            fontSize: 7.5,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.black,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                pw.SizedBox(height: 6),

                // Product Table Headers
                pw.Container(
                  color: PdfColors.grey200,
                  padding: const pw.EdgeInsets.symmetric(
                    vertical: 3,
                    horizontal: 6,
                  ),
                  child: pw.Row(
                    children: [
                      pw.Expanded(
                        flex: 3,
                        child: pw.Text(
                          'Nama Produk',
                          style: pw.TextStyle(
                            color: PdfColors.black,
                            fontWeight: pw.FontWeight.bold,
                            fontSize: 7,
                          ),
                        ),
                      ),
                      pw.Expanded(
                        flex: 1,
                        child: pw.Text(
                          'Qty',
                          style: pw.TextStyle(
                            color: PdfColors.black,
                            fontWeight: pw.FontWeight.bold,
                            fontSize: 7,
                          ),
                          textAlign: pw.TextAlign.center,
                        ),
                      ),
                      pw.Expanded(
                        flex: 2,
                        child: pw.Text(
                          'Harga',
                          style: pw.TextStyle(
                            color: PdfColors.black,
                            fontWeight: pw.FontWeight.bold,
                            fontSize: 7,
                          ),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                      pw.Expanded(
                        flex: 2,
                        child: pw.Text(
                          'Subtotal',
                          style: pw.TextStyle(
                            color: PdfColors.black,
                            fontWeight: pw.FontWeight.bold,
                            fontSize: 7,
                          ),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                ),

                // Product Table Items
                ...items.map((item) {
                  final itemMap = Map<String, dynamic>.from(item as Map);

                  // Name: try product_name -> product.name -> name
                  final prodName =
                      itemMap['product_name'] ??
                      (itemMap['product'] as Map?)?['name'] ??
                      itemMap['name'] ??
                      'Produk';

                  // Price: try price -> price_at_transaction -> selling_price from nested product
                  final price =
                      ((itemMap['price'] ??
                                  itemMap['price_at_transaction'] ??
                                  (itemMap['product']
                                      as Map?)?['selling_price'] ??
                                  0)
                              as num)
                          .toDouble();

                  // Qty
                  final qty =
                      ((itemMap['quantity'] ?? itemMap['qty'] ?? 0) as num)
                          .toDouble();

                  // Subtotal: prefer stored subtotal, fallback to calc
                  final subtotal =
                      ((itemMap['subtotal'] ?? (qty * price)) as num)
                          .toDouble();

                  final unit =
                      itemMap['unit'] ??
                      (itemMap['product'] as Map?)?['unit'] ??
                      'PCS';

                  return pw.Container(
                    padding: const pw.EdgeInsets.symmetric(
                      vertical: 2,
                      horizontal: 6,
                    ),
                    decoration: const pw.BoxDecoration(
                      border: pw.Border(
                        bottom: pw.BorderSide(color: PdfColors.grey300),
                      ),
                    ),
                    child: pw.Row(
                      children: [
                        pw.Expanded(
                          flex: 3,
                          child: pw.Text(
                            prodName,
                            style: const pw.TextStyle(fontSize: 7),
                          ),
                        ),
                        pw.Expanded(
                          flex: 1,
                          child: pw.Text(
                            '${qty.toInt()} $unit',
                            style: const pw.TextStyle(fontSize: 7),
                            textAlign: pw.TextAlign.center,
                          ),
                        ),
                        pw.Expanded(
                          flex: 2,
                          child: pw.Text(
                            price > 0 ? currencyFormat.format(price) : '-',
                            style: const pw.TextStyle(fontSize: 7),
                            textAlign: pw.TextAlign.right,
                          ),
                        ),
                        pw.Expanded(
                          flex: 2,
                          child: pw.Text(
                            subtotal > 0
                                ? currencyFormat.format(subtotal)
                                : '-',
                            style: const pw.TextStyle(fontSize: 7),
                            textAlign: pw.TextAlign.right,
                          ),
                        ),
                      ],
                    ),
                  );
                }),

                if (bonusItems.isNotEmpty) ...[
                  pw.SizedBox(height: 4),
                  pw.Text(
                    'Bonus / Hadiah:',
                    style: pw.TextStyle(
                      fontSize: 7,
                      fontWeight: pw.FontWeight.bold,
                    ),
                  ),
                  ...bonusItems.map((item) {
                    final itemMap = Map<String, dynamic>.from(item as Map);
                    final prodName =
                        itemMap['product_name'] ??
                        (itemMap['bonus_inventory'] as Map?)?['name'] ??
                        itemMap['name'] ??
                        'Bonus';
                    final qty =
                        ((itemMap['quantity'] ?? itemMap['qty'] ?? 0) as num)
                            .toDouble();
                    final unit =
                        itemMap['unit'] ??
                        (itemMap['bonus_inventory'] as Map?)?['unit'] ??
                        'PCS';

                    return pw.Container(
                      padding: const pw.EdgeInsets.symmetric(
                        vertical: 2,
                        horizontal: 6,
                      ),
                      decoration: const pw.BoxDecoration(
                        border: pw.Border(
                          bottom: pw.BorderSide(color: PdfColors.grey300),
                        ),
                      ),
                      child: pw.Row(
                        children: [
                          pw.Expanded(
                            flex: 3,
                            child: pw.Text(
                              '$prodName (Hadiah)',
                              style: const pw.TextStyle(
                                fontSize: 7,
                                color: PdfColors.black,
                              ),
                            ),
                          ),
                          pw.Expanded(
                            flex: 1,
                            child: pw.Text(
                              '${qty.toInt()} $unit',
                              style: const pw.TextStyle(fontSize: 7),
                              textAlign: pw.TextAlign.center,
                            ),
                          ),
                          pw.Expanded(
                            flex: 2,
                            child: pw.Text(
                              '-',
                              style: const pw.TextStyle(fontSize: 7),
                              textAlign: pw.TextAlign.right,
                            ),
                          ),
                          pw.Expanded(
                            flex: 2,
                            child: pw.Text(
                              '-',
                              style: const pw.TextStyle(fontSize: 7),
                              textAlign: pw.TextAlign.right,
                            ),
                          ),
                        ],
                      ),
                    );
                  }),
                ],

                pw.SizedBox(height: 6),

                // Totals
                pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.end,
                  children: [
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.end,
                      children: [
                        pw.Row(
                          children: [
                            pw.Text(
                              'Subtotal: ',
                              style: const pw.TextStyle(fontSize: 8),
                            ),
                            pw.Text(
                              currencyFormat.format(totalAmount),
                              style: pw.TextStyle(
                                fontSize: 8,
                                fontWeight: pw.FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                        pw.Row(
                          children: [
                            pw.Text(
                              'Total Bayar: ',
                              style: const pw.TextStyle(fontSize: 8),
                            ),
                            pw.Text(
                              currencyFormat.format(totalAmount),
                              style: pw.TextStyle(
                                fontSize: 8,
                                fontWeight: pw.FontWeight.bold,
                                color: PdfColors.black,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ],
                ),
                pw.SizedBox(height: 10),

                // Signatures
                pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.spaceAround,
                  children: [
                    pw.Column(
                      children: [
                        pw.Text(
                          'Penerima / Toko',
                          style: pw.TextStyle(
                            fontSize: 8,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.SizedBox(height: 4),
                        if (customerSigImg != null)
                          pw.Image(
                            customerSigImg,
                            width: 70,
                            height: 35,
                            fit: pw.BoxFit.contain,
                          )
                        else
                          pw.SizedBox(height: 20),
                        pw.SizedBox(height: 4),
                        pw.Text(
                          'Pelanggan',
                          style: const pw.TextStyle(fontSize: 7),
                        ),
                      ],
                    ),
                    pw.Column(
                      children: [
                        pw.Text(
                          'Salesman / Motoris',
                          style: pw.TextStyle(
                            fontSize: 8,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.SizedBox(height: 4),
                        if (courierSigImg != null)
                          pw.Image(
                            courierSigImg,
                            width: 70,
                            height: 35,
                            fit: pw.BoxFit.contain,
                          )
                        else
                          pw.SizedBox(height: 20),
                        pw.SizedBox(height: 4),
                        pw.Text(
                          'Sales Motoris',
                          style: const pw.TextStyle(fontSize: 8),
                        ),
                      ],
                    ),
                  ],
                ),

                pw.Spacer(),
                pw.Center(
                  child: pw.Text(
                    'Terima kasih atas kerja sama dan kepercayaan Anda.',
                    style: pw.TextStyle(
                      fontSize: 7,
                      fontStyle: pw.FontStyle.italic,
                      color: PdfColors.grey600,
                    ),
                  ),
                ),
              ],
            );
          },
        ),
      );

      final pdfBytes = await pdf.save();

      // Close the loading dialog
      if (context.mounted) {
        Navigator.pop(context);
      }

      if (kIsWeb) {
        final xFile = XFile.fromData(
          pdfBytes,
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
        await file.writeAsBytes(pdfBytes);

        // Launch Native Share Sheet
        final xFile = XFile(file.path, mimeType: 'application/pdf');
        await Share.shareXFiles(
          [xFile],
          text:
              'Halo! Berikut terlampir salinan digital Faktur Penjualan resmi Anda ($receiptNo) dari $companyName. Terima kasih!',
        );
      }
    } catch (e) {
      // Close the loading dialog if it is still open
      if (context.mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal mengunduh faktur: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
      print('Gagal menghasilkan / membagikan PDF: $e');
    }
  }

  Future<void> _printReceipt(BuildContext context) async {
    // Simulasi koneksi ke Printer Thermal Bluetooth
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext context) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const CircularProgressIndicator(),
              const SizedBox(height: 16),
              Text(
                'Mencari Printer Bluetooth...',
                style: GoogleFonts.outfit(fontWeight: FontWeight.bold),
              ),
            ],
          ),
        );
      },
    );

    // Mock delay untuk pencarian dan proses print
    await Future.delayed(const Duration(seconds: 2));

    if (context.mounted) {
      Navigator.pop(context); // Tutup dialog loading
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Row(
            children: [
              const Icon(Icons.print, color: Colors.white),
              const SizedBox(width: 8),
              Text(
                'Mencetak struk ke printer thermal...',
                style: GoogleFonts.outfit(),
              ),
            ],
          ),
          backgroundColor: Colors.blueAccent,
        ),
      );
    }
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

  @override
  Widget build(BuildContext context) {
    final storeName = _currentTrx['store']?['name'] ?? 'Toko';
    final receiptNo = _currentTrx['receipt_no'] ?? 'NO-DATA';
    final amount = (_currentTrx['total_amount'] ?? 0.0).toDouble();
    final date =
        DateTime.tryParse((_currentTrx['transaction_date'] ?? '').toString()) ??
        DateTime.now();
    final dateStr = DateFormat('dd MMM yyyy').format(date);
    final currencyFormat = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: Text(
          'NOTA DIGITAL',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.w900,
            color: const Color(0xFF1E293B),
          ),
        ),
        centerTitle: true,
        actions: [
          IconButton(
            onPressed: _fetchLatestData,
            icon: const Icon(Icons.refresh_rounded, color: Colors.blueAccent),
          ),
        ],
        leading: IconButton(
          icon: const Icon(Icons.close_rounded, color: Color(0xFF334155)),
          onPressed: () {
            di.sl<SyncBloc>().add(const SyncDataRequested());
            Navigator.of(context).popUntil((route) {
              final name = route.settings.name;
              return (name != null && name.endsWith('_dashboard')) ||
                  route.isFirst;
            });
          },
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(32),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.05),
                    blurRadius: 20,
                    offset: const Offset(0, 10),
                  ),
                ],
              ),
              child: Column(
                children: [
                  const SizedBox(height: 32),

                  // Barcode or QRIS Section
                  if ((_currentTrx['payment_method'] == 'QRIS' ||
                          _currentTrx['dp_payment_method'] == 'QRIS') &&
                      _currentTrx['midtrans_qris_url'] != null)
                    _buildQRISSection(_currentTrx['midtrans_qris_url'])
                  else
                    _buildBarcodeSection(receiptNo),

                  const SizedBox(height: 32),

                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 32),
                    child: Column(
                      children: [
                        _buildRow('Toko', storeName, isBold: true),
                        const Divider(height: 32),

                        // Item Details
                        if (_currentTrx['items'] != null &&
                            (_currentTrx['items'] as List).isNotEmpty) ...[
                          ...((_currentTrx['items'] as List).map((item) {
                            final itemMap = Map<String, dynamic>.from(
                              item as Map,
                            );
                            return Padding(
                              padding: const EdgeInsets.only(bottom: 8),
                              child: Row(
                                mainAxisAlignment:
                                    MainAxisAlignment.spaceBetween,
                                children: [
                                  Expanded(
                                    child: Text(
                                      '${itemMap['product_name'] ?? (itemMap['product'] as Map?)?['name'] ?? itemMap['name'] ?? 'Produk'}',
                                      style: GoogleFonts.outfit(
                                        fontSize: 13,
                                        color: const Color(0xFF1E293B),
                                        fontWeight: FontWeight.w600,
                                      ),
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
                          const Divider(height: 24),
                        ],

                        _buildRow('Tanggal', dateStr),
                        const SizedBox(height: 12),
                        _buildRow('Status', () {
                          final status = (_currentTrx['status'] ?? 'PENDING')
                              .toString()
                              .toUpperCase();
                          final method = (_currentTrx['payment_method'] ?? '')
                              .toString()
                              .toUpperCase();
                          final payStatus =
                              (_currentTrx['payment_status'] ?? '')
                                  .toString()
                                  .toUpperCase();
                          final total =
                              (_currentTrx['total_amount'] as num?)
                                  ?.toDouble() ??
                              0.0;
                          final paid =
                              (_currentTrx['paid_amount'] as num?)
                                  ?.toDouble() ??
                              (_currentTrx['dp_amount'] as num?)?.toDouble() ??
                              0.0;
                          bool isPaid =
                              payStatus == 'PAID' ||
                              payStatus == 'LUNAS' ||
                              status == 'LUNAS' ||
                              (total > 0 && paid >= total);

                          if (method == 'TEMPO' && paid < total) {
                            isPaid = false;
                          }
                          if (method == 'SALES_ORDER' && payStatus != 'PAID') {
                            isPaid = false;
                          }

                          if (method == 'TEMPO' && !isPaid) {
                            return 'BELUM LUNAS (TEMPO)';
                          }
                          if (method == 'SALES_ORDER' && !isPaid) {
                            return 'BELUM LUNAS';
                          }
                          return _currentTrx['status'] ?? 'PENDING';
                        }()),
                        const SizedBox(height: 12),
                        _buildRow('Metode Bayar', () {
                          String method =
                              _currentTrx['payment_method'] ?? 'CASH';
                          if (method == 'TEMPO' &&
                              _currentTrx['dp_payment_method'] != null) {
                            return 'TEMPO (DP: ${_currentTrx['dp_payment_method']})';
                          }
                          return method;
                        }()),
                        const SizedBox(height: 12),
                        _buildRow(
                          'Total',
                          currencyFormat.format(amount),
                          isPrice: true,
                        ),
                      ],
                    ),
                  ),

                  // VA Information if exists
                  if (_currentTrx['payment_method'] == 'VA' &&
                      ((_currentTrx['midtrans_va_number'] != null &&
                              _currentTrx['midtrans_va_number']
                                  .toString()
                                  .isNotEmpty) ||
                          (_currentTrx['midtrans_bill_key'] != null &&
                              _currentTrx['midtrans_bill_key']
                                  .toString()
                                  .isNotEmpty))) ...[
                    const SizedBox(height: 24),
                    Container(
                      margin: const EdgeInsets.symmetric(horizontal: 24),
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: Colors.blueAccent.withOpacity(0.05),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: Colors.blueAccent.withOpacity(0.2),
                        ),
                      ),
                      child: Column(
                        children: [
                          Text(
                            'NOMOR VIRTUAL ACCOUNT (${(_currentTrx['midtrans_bank'] ?? 'Bank').toString().toUpperCase()})',
                            style: GoogleFonts.outfit(
                              fontSize: 10,
                              fontWeight: FontWeight.w900,
                              color: Colors.blueAccent,
                              letterSpacing: 1,
                            ),
                          ),
                          const SizedBox(height: 8),
                          if (_currentTrx['midtrans_va_number'] != null &&
                                  _currentTrx['midtrans_va_number']
                                      .toString()
                                      .isNotEmpty ||
                              _currentTrx['midtrans_bill_key'] != null &&
                                  _currentTrx['midtrans_bill_key']
                                      .toString()
                                      .isNotEmpty) ...[
                            Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Text(
                                  (_currentTrx['midtrans_va_number'] != null &&
                                          _currentTrx['midtrans_va_number']
                                              .toString()
                                              .isNotEmpty)
                                      ? _currentTrx['midtrans_va_number']
                                      : _currentTrx['midtrans_bill_key'],
                                  style: GoogleFonts.outfit(
                                    fontSize: 20,
                                    fontWeight: FontWeight.w900,
                                    color: const Color(0xFF1E293B),
                                    letterSpacing: 1,
                                  ),
                                ),
                                IconButton(
                                  icon: const Icon(
                                    Icons.copy_rounded,
                                    color: Colors.blueGrey,
                                    size: 20,
                                  ),
                                  onPressed: () {
                                    final va =
                                        (_currentTrx['midtrans_va_number'] !=
                                                null &&
                                            _currentTrx['midtrans_va_number']
                                                .toString()
                                                .isNotEmpty)
                                        ? _currentTrx['midtrans_va_number']
                                        : _currentTrx['midtrans_bill_key'];
                                    Clipboard.setData(ClipboardData(text: va));
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      const SnackBar(
                                        content: Text('Nomor VA disalin!'),
                                        backgroundColor: Colors.green,
                                      ),
                                    );
                                  },
                                ),
                              ],
                            ),
                            if (_currentTrx['midtrans_biller_code'] !=
                                null) ...[
                              const SizedBox(height: 4),
                              Row(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Text(
                                    'Kode Biller: ${_currentTrx['midtrans_biller_code']}',
                                    style: GoogleFonts.outfit(
                                      fontSize: 12,
                                      fontWeight: FontWeight.w700,
                                      color: Colors.blueGrey,
                                    ),
                                  ),
                                  IconButton(
                                    icon: const Icon(
                                      Icons.copy_rounded,
                                      color: Colors.blueGrey,
                                      size: 16,
                                    ),
                                    onPressed: () {
                                      Clipboard.setData(
                                        ClipboardData(
                                          text:
                                              _currentTrx['midtrans_biller_code'],
                                        ),
                                      );
                                      ScaffoldMessenger.of(
                                        context,
                                      ).showSnackBar(
                                        const SnackBar(
                                          content: Text('Kode Biller disalin!'),
                                          backgroundColor: Colors.green,
                                        ),
                                      );
                                    },
                                    padding: EdgeInsets.zero,
                                    constraints: const BoxConstraints(),
                                  ),
                                ],
                              ),
                            ],
                          ] else ...[
                            const SizedBox(height: 8),
                            const CircularProgressIndicator(strokeWidth: 2),
                            const SizedBox(height: 8),
                            Text(
                              'Sedang menyiapkan nomor VA...',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                color: Colors.blueGrey,
                                fontStyle: FontStyle.italic,
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ],
                  if ((_currentTrx['receipt_path'] != null &&
                          _currentTrx['receipt_path'].toString().isNotEmpty) ||
                      (_currentTrx['receipt_image_url'] != null &&
                          _currentTrx['receipt_image_url']
                              .toString()
                              .isNotEmpty)) ...[
                    const SizedBox(height: 24),
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 24),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'BUKTI NOTA FISIK / TRANSAKSI',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
                              color: Colors.blueGrey,
                              letterSpacing: 1.5,
                            ),
                          ),
                          const SizedBox(height: 12),
                          GestureDetector(
                            onTap: () => _showImagePreviewDialog(
                              _currentTrx['receipt_path'] ??
                                  _currentTrx['receipt_image_url'],
                            ),
                            child: ClipRRect(
                              borderRadius: BorderRadius.circular(16),
                              child: _buildReceiptImage(
                                _currentTrx['receipt_path'] ??
                                    _currentTrx['receipt_image_url'],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],

                  const SizedBox(height: 40),

                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      color: Colors.blueAccent.withOpacity(0.05),
                      borderRadius: const BorderRadius.only(
                        bottomLeft: Radius.circular(32),
                        bottomRight: Radius.circular(32),
                      ),
                    ),
                    child: Column(
                      children: [
                        Text(
                          _currentTrx['payment_method'] == 'QRIS'
                              ? 'SCAN QRIS DI ATAS'
                              : 'TUNJUKKAN KE ADMIN',
                          style: GoogleFonts.outfit(
                            fontSize: 12,
                            fontWeight: FontWeight.w900,
                            color: Colors.blueAccent,
                            letterSpacing: 1,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _currentTrx['payment_method'] == 'QRIS'
                              ? 'Pelanggan dapat memindai kode QR di atas menggunakan aplikasi perbankan atau e-wallet.'
                              : 'Gunakan barcode di atas untuk validasi transaksi oleh Admin Nota atau Kasir.',
                          textAlign: TextAlign.center,
                          style: GoogleFonts.outfit(
                            fontSize: 12,
                            color: Colors.blueGrey,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),

            const SizedBox(height: 24),

            // Bukti Transfer Preview (if already uploaded)
            if (_currentTrx['transfer_proof_url'] != null &&
                _currentTrx['transfer_proof_url'].toString().isNotEmpty) ...[
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.indigo.shade50,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: Colors.indigo.shade200),
                ),
                child: Row(
                  children: [
                    Container(
                      width: 48,
                      height: 48,
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: Colors.indigo.shade100),
                      ),
                      clipBehavior: Clip.antiAlias,
                      child: GestureDetector(
                        onTap: () => _showImagePreviewDialog(
                          _currentTrx['transfer_proof_url'].toString(),
                        ),
                        child: _buildReceiptImage(
                          _currentTrx['transfer_proof_url'].toString(),
                          height: 48,
                          width: 48,
                          fit: BoxFit.cover,
                        ),
                      ),
                    ),
                    const SizedBox(width: 14),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Bukti Pembayaran / Transfer',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800,
                              fontSize: 13,
                              color: Colors.indigo.shade900,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            'Tersimpan & siap diverifikasi Kasir',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              color: Colors.indigo.shade600,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                    ),
                    ElevatedButton.icon(
                      onPressed: () => _showImagePreviewDialog(
                        _currentTrx['transfer_proof_url'].toString(),
                      ),
                      icon: const Icon(
                        Icons.remove_red_eye_rounded,
                        size: 14,
                        color: Colors.white,
                      ),
                      label: Text(
                        'Lihat',
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w800,
                          fontSize: 11,
                          color: Colors.white,
                        ),
                      ),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.indigo,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 8,
                        ),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(10),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
            ],

            // Terima Pembayaran / Upload Bukti Transfer Button (For all unpaid / pending invoices)
            if (_currentTrx['payment_status'] != 'PAID' &&
                _currentTrx['payment_status'] != 'SUCCESS' &&
                _currentTrx['status'] != 'REJECTED' &&
                _currentTrx['status'] != 'DELETED') ...[
              SizedBox(
                width: double.infinity,
                height: 56,
                child: ElevatedButton.icon(
                  onPressed: _showTempoPaymentDialog,
                  icon: Icon(
                    (_currentTrx['payment_method'] == 'TRANSFER' ||
                            _currentTrx['payment_method'] == 'VA')
                        ? Icons.upload_file_rounded
                        : Icons.payments_rounded,
                    color: Colors.white,
                  ),
                  label: Text(
                    (_currentTrx['payment_method'] == 'TRANSFER' ||
                            _currentTrx['payment_method'] == 'VA')
                        ? 'UPLOAD BUKTI TRANSFER'
                        : 'TERIMA PEMBAYARAN',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900,
                      color: Colors.white,
                      letterSpacing: 1,
                    ),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor:
                        (_currentTrx['payment_method'] == 'TRANSFER' ||
                            _currentTrx['payment_method'] == 'VA')
                        ? Colors.indigo
                        : Colors.blueAccent,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 16),
            ],

            // Action Buttons Row (PDF, WA, Print)
            Row(
              children: [
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () => _generateAndSharePdf(context),
                    icon: const Icon(
                      Icons.picture_as_pdf_rounded,
                      color: Colors.white,
                      size: 18,
                    ),
                    label: Text(
                      'PDF',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w800,
                        color: Colors.white,
                        fontSize: 14,
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.redAccent.shade700,
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      elevation: 0,
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () => _shareToWhatsApp(context),
                    icon: const Icon(
                      Icons.chat_rounded,
                      color: Colors.white,
                      size: 18,
                    ),
                    label: Text(
                      'WA',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w800,
                        color: Colors.white,
                        fontSize: 14,
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF25D366),
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      elevation: 0,
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () => _printReceipt(context),
                    icon: const Icon(
                      Icons.print_rounded,
                      color: Colors.white,
                      size: 18,
                    ),
                    label: Text(
                      'PRINT',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w800,
                        color: Colors.white,
                        fontSize: 14,
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF1E293B),
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      elevation: 0,
                    ),
                  ),
                ),
              ],
            ),

            const SizedBox(height: 16),

            SizedBox(
              width: double.infinity,
              height: 56,
              child: OutlinedButton.icon(
                onPressed: () {
                  di.sl<SyncBloc>().add(const SyncDataRequested());
                  Navigator.of(context).popUntil((route) {
                    final name = route.settings.name;
                    return (name != null && name.endsWith('_dashboard')) ||
                        route.isFirst;
                  });
                },
                icon: const Icon(Icons.home_rounded, color: Color(0xFF1E293B)),
                label: Text(
                  'KEMBALI KE DASHBOARD',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w900,
                    color: const Color(0xFF1E293B),
                  ),
                ),
                style: OutlinedButton.styleFrom(
                  side: const BorderSide(color: Color(0xFFE2E8F0), width: 2),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBarcodeSection(String data) {
    return Container(
      padding: const EdgeInsets.all(24),
      margin: const EdgeInsets.symmetric(horizontal: 24),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(24),
      ),
      child: Column(
        children: [
          BarcodeWidget(
            barcode: Barcode.code128(),
            data: data,
            width: double.infinity,
            height: 100,
            drawText: false,
          ),
          const SizedBox(height: 12),
          Text(
            data,
            style: GoogleFonts.outfit(
              fontSize: 16,
              fontWeight: FontWeight.w800,
              color: const Color(0xFF334155),
              letterSpacing: 2,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildQRISSection(String url) {
    return Container(
      padding: const EdgeInsets.all(24),
      margin: const EdgeInsets.symmetric(horizontal: 24),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        children: [
          if (url.startsWith('http')) ...[
            QrisSandboxWarning(
              qrisUrl: url,
              margin: const EdgeInsets.only(bottom: 16),
            ),
          ],
          url.startsWith('http')
              ? CachedNetworkImage(
                  imageUrl:
                      '${AppConstants.baseUrl}sales/qris-image?url=${Uri.encodeComponent(url)}',
                  height: 240,
                  width: 240,
                  fit: BoxFit.contain,
                  placeholder: (context, url) =>
                      const Center(child: CircularProgressIndicator()),
                  errorWidget: (context, url, error) => const Icon(Icons.error),
                  memCacheWidth: 800,
                )
              : QrImageView(data: url, version: QrVersions.auto, size: 240.0),
          const SizedBox(height: 12),
          Text(
            'SCAN QRIS UNTUK BAYAR',
            style: GoogleFonts.outfit(
              fontSize: 14,
              fontWeight: FontWeight.w900,
              color: Colors.blueAccent,
            ),
          ),
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
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: Colors.blueGrey,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                IconButton(
                  icon: const Icon(
                    Icons.copy_rounded,
                    size: 18,
                    color: Colors.blueAccent,
                  ),
                  onPressed: () {
                    Clipboard.setData(ClipboardData(text: url.toString()));
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: Text(
                          'Tautan QRIS Disalin!',
                          style: GoogleFonts.outfit(),
                        ),
                        backgroundColor: Colors.green,
                      ),
                    );
                  },
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRow(
    String label,
    String value, {
    bool isBold = false,
    bool isPrice = false,
  }) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: GoogleFonts.outfit(
            fontSize: 14,
            color: Colors.blueGrey,
            fontWeight: FontWeight.w600,
          ),
        ),
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
}
