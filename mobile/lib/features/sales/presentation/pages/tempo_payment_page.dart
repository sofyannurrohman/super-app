import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_checkout_page.dart'
    show SignatureCanvas;
import 'dart:ui' as ui;
import 'package:flutter/foundation.dart';
import 'dart:convert';
import 'package:image_picker/image_picker.dart';
import 'package:flutter/services.dart';
import 'dart:io';
import 'package:shared_preferences/shared_preferences.dart';

class TempoPaymentPage extends StatefulWidget {
  final Map<String, dynamic> transactionData;
  final Function(String, double, String?) onSuccess;

  const TempoPaymentPage({
    super.key,
    required this.transactionData,
    required this.onSuccess,
  });

  @override
  State<TempoPaymentPage> createState() => _TempoPaymentPageState();
}

class _TempoPaymentPageState extends State<TempoPaymentPage> {
  bool _isSubmitting = false;
  String _selectedPaymentMethod = 'CASH'; // 'CASH', 'TRANSFER'
  late double _totalAmount;
  late double _paidAmount;
  late double _remainingAmount;
  late TextEditingController _amountController;
  Map<String, dynamic>? _activeBankAccount;
  List<Map<String, dynamic>> _bankAccounts = [];
  Map<String, dynamic>? _selectedBankAccount;
  bool _isLoadingBanks = false;
  String _branchId = '';
  String _companyId = '';

  List<Offset> _courierSignaturePoints = [];
  List<Offset> _customerSignaturePoints = [];

  final ImagePicker _picker = ImagePicker();
  XFile? _transferProofImage;

  Future<void> _pickTransferProofPhoto() async {
    final XFile? photo = await _picker.pickImage(
      source: ImageSource.gallery,
      imageQuality: 70,
    );
    if (photo != null) {
      setState(() {
        _transferProofImage = photo;
      });
    }
  }

  Future<void> _takeTransferProofPhoto() async {
    final XFile? photo = await _picker.pickImage(
      source: ImageSource.camera,
      imageQuality: 70,
    );
    if (photo != null) {
      setState(() {
        _transferProofImage = photo;
      });
    }
  }

  final Map<String, Map<String, dynamic>> _paymentOptions = {
    'CASH': {
      'label': 'Tunai (Cash)',
      'icon': Icons.payments_rounded,
      'color': Colors.green,
    },
    'TRANSFER': {
      'label': 'Transfer Bank',
      'icon': Icons.account_balance_rounded,
      'color': Colors.indigo,
    },
  };

  @override
  void initState() {
    super.initState();
    final orderData =
        widget.transactionData['sales_transaction'] ??
        widget.transactionData['sales_order'] ??
        widget.transactionData;

    _totalAmount =
        double.tryParse(orderData['total_amount']?.toString() ?? '0') ?? 0.0;
    _paidAmount =
        double.tryParse(orderData['paid_amount']?.toString() ?? '0') ?? 0.0;

    _remainingAmount = _totalAmount - _paidAmount;
    if (_remainingAmount <= 0) _remainingAmount = _totalAmount;

    final initialMethod =
        orderData['payment_method']?.toString().toUpperCase() ?? '';
    if (initialMethod == 'TRANSFER' ||
        initialMethod == 'VA' ||
        initialMethod == 'QRIS') {
      _selectedPaymentMethod = 'TRANSFER';
    }

    _branchId =
        orderData['branch_id']?.toString() ??
        orderData['employee']?['branch_id']?.toString() ??
        '';
    _companyId = orderData['company_id']?.toString() ?? '';

    _amountController = TextEditingController(
      text: _remainingAmount.toStringAsFixed(0),
    );
    _fetchBankAccounts();
  }

  Future<void> _fetchBankAccounts() async {
    setState(() => _isLoadingBanks = true);
    try {
      if (_branchId.isEmpty || _companyId.isEmpty) {
        final prefs = await SharedPreferences.getInstance();
        if (_branchId.isEmpty) _branchId = prefs.getString('branch_id') ?? '';
        if (_companyId.isEmpty)
          _companyId = prefs.getString('company_id') ?? '';
      }
      final salesApi = SalesApiService(apiClient: di.sl<ApiClient>());
      final banks = await salesApi.getBankAccounts(branchId: _branchId);
      if (mounted) {
        setState(() {
          _bankAccounts = banks;
          _isLoadingBanks = false;
          if (_bankAccounts.isNotEmpty) {
            _selectedBankAccount = _bankAccounts.firstWhere(
              (b) => b['is_active'] == true,
              orElse: () => _bankAccounts.first,
            );
            _activeBankAccount = _selectedBankAccount;
          }
        });
      }
    } catch (_) {
      if (mounted) setState(() => _isLoadingBanks = false);
    }
  }

  Future<void> _fetchActiveBankAccount() async {
    await _fetchBankAccounts();
  }

  Future<Uint8List?> _renderSignatureToPng(List<Offset> points) async {
    if (points.isEmpty) return null;
    final recorder = ui.PictureRecorder();
    final canvas = ui.Canvas(recorder, const Rect.fromLTWH(0, 0, 300, 150));
    final bgPaint = Paint()..color = Colors.white;
    canvas.drawRect(const Rect.fromLTWH(0, 0, 300, 150), bgPaint);
    final paint = Paint()
      ..color = const Color(0xFF0F172A)
      ..strokeCap = StrokeCap.round
      ..strokeWidth = 4.0;
    for (int i = 0; i < points.length - 1; i++) {
      if (points[i] != Offset.infinite && points[i + 1] != Offset.infinite) {
        canvas.drawLine(points[i], points[i + 1], paint);
      }
    }
    final picture = recorder.endRecording();
    final img = await picture.toImage(300, 150);
    final pngBytes = await img.toByteData(format: ui.ImageByteFormat.png);

    // Dispose GPU resources to prevent CanvasKit context loss
    img.dispose();
    picture.dispose();

    // Yield to browser event loop to prevent WebGL Context Loss on Flutter Web
    await Future.delayed(const Duration(milliseconds: 100));

    return pngBytes?.buffer.asUint8List();
  }

  Future<void> _submitPayment() async {
    if (_isSubmitting) return;
    final amt = double.tryParse(_amountController.text) ?? 0;
    if (amt <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Nominal pembayaran tidak valid.')),
      );
      return;
    }

    if (_selectedPaymentMethod == 'TRANSFER' && _transferProofImage == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Harap upload bukti transfer bank.'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    final isTransferWithProof =
        _selectedPaymentMethod == 'TRANSFER' && _transferProofImage != null;

    if (!isTransferWithProof &&
        (_courierSignaturePoints.isEmpty || _customerSignaturePoints.isEmpty)) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Tanda tangan penagih dan pelanggan wajib diisi untuk tunai.',
            style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
          ),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    setState(() => _isSubmitting = true);
    try {
      final orderData =
          widget.transactionData['sales_transaction'] ??
          widget.transactionData['sales_order'] ??
          widget.transactionData;
      final trxId = orderData['id'] ?? widget.transactionData['id'];

      if (trxId == null) {
        throw Exception(
          'ID Transaksi tidak ditemukan atau faktur belum sinkron.',
        );
      }

      String? combinedSignatureBase64;
      try {
        final courierSigBytes = _courierSignaturePoints.isNotEmpty
            ? await _renderSignatureToPng(_courierSignaturePoints)
            : null;
        final customerSigBytes = _customerSignaturePoints.isNotEmpty
            ? await _renderSignatureToPng(_customerSignaturePoints)
            : null;
        if (courierSigBytes != null && customerSigBytes != null) {
          final courierB64 =
              'data:image/png;base64,${base64Encode(courierSigBytes)}';
          final customerB64 =
              'data:image/png;base64,${base64Encode(customerSigBytes)}';
          combinedSignatureBase64 = '$courierB64,$customerB64';
        } else if (courierSigBytes != null) {
          final courierB64 =
              'data:image/png;base64,${base64Encode(courierSigBytes)}';
          combinedSignatureBase64 = courierB64;
        }
      } catch (e) {
        debugPrint('Gagal render ttd: $e');
      }

      String? transferProofBase64;
      if (_selectedPaymentMethod == 'TRANSFER' && _transferProofImage != null) {
        try {
          final bytes = await _transferProofImage!.readAsBytes();
          transferProofBase64 = 'data:image/png;base64,${base64Encode(bytes)}';
        } catch (e) {
          debugPrint('Error converting transfer proof image to base64: $e');
        }
      }

      final apiService = SalesApiService(apiClient: di.sl<ApiClient>());
      await apiService.recordPayment(trxId, {
        'amount': amt,
        'payment_method': _selectedPaymentMethod,
        'notes': 'Penagihan Tempo via Mobile',
        if (_selectedPaymentMethod == 'TRANSFER' &&
            _selectedBankAccount != null)
          'company_bank_account_id': _selectedBankAccount!['id'],
        if (combinedSignatureBase64 != null)
          'signature_image_url': combinedSignatureBase64,
        if (transferProofBase64 != null)
          'transfer_proof_url': transferProofBase64,
      });

      if (mounted) {
        showDialog(
          context: context,
          barrierDismissible: false,
          builder: (dialogContext) => AlertDialog(
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(20),
            ),
            icon: const Icon(
              Icons.check_circle_rounded,
              color: Colors.green,
              size: 48,
            ),
            title: Text(
              'Pembayaran Berhasil',
              style: GoogleFonts.outfit(fontWeight: FontWeight.w800),
            ),
            content: Text(
              'Pembayaran tempo berhasil dicatat.',
              style: GoogleFonts.outfit(),
              textAlign: TextAlign.center,
            ),
            actions: [
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.green,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  onPressed: () {
                    Navigator.pop(dialogContext);
                    widget.onSuccess(
                      _selectedPaymentMethod,
                      amt,
                      combinedSignatureBase64,
                    );
                  },
                  child: Text(
                    'TUTUP',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            ],
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal mencatat pembayaran: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  Widget _row(IconData icon, String label, String value, {Color? valueColor}) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, color: Colors.blueAccent, size: 18),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  color: Colors.blueGrey,
                  fontWeight: FontWeight.w700,
                ),
              ),
              Text(
                value,
                style: GoogleFonts.outfit(
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                  color: valueColor ?? const Color(0xFF1E293B),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final currency = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    final orderData =
        widget.transactionData['sales_transaction'] ??
        widget.transactionData['sales_order'] ??
        widget.transactionData;

    final storeName = orderData['store']?['name'] ?? 'Toko Tanpa Nama';
    final receiptNo =
        orderData['invoice_no'] ??
        orderData['so_number'] ??
        orderData['receipt_no'] ??
        'INV-PENDING';

    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(
            Icons.arrow_back_ios_rounded,
            color: Color(0xFF334155),
          ),
          onPressed: _isSubmitting ? null : () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'PENAGIHAN TEMPO',
              style: GoogleFonts.outfit(
                fontSize: 10,
                color: Colors.blueAccent,
                fontWeight: FontWeight.w800,
                letterSpacing: 1,
              ),
            ),
            Text(
              'Pembayaran Faktur',
              style: GoogleFonts.outfit(
                fontSize: 17,
                fontWeight: FontWeight.w800,
                color: const Color(0xFF1E293B),
              ),
            ),
          ],
        ),
      ),
      body: Column(
        children: [
          LinearProgressIndicator(
            value: 1.0,
            minHeight: 4,
            backgroundColor: Colors.grey.shade200,
            color: Colors.blueAccent,
          ),
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: Column(
                children: [
                  // Summary card
                  Container(
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(28),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.04),
                          blurRadius: 20,
                          offset: const Offset(0, 8),
                        ),
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: Colors.blueAccent.withOpacity(0.1),
                                shape: BoxShape.circle,
                              ),
                              child: const Icon(
                                Icons.receipt_long_rounded,
                                color: Colors.blueAccent,
                                size: 28,
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'Sisa Tagihan Tempo',
                                    style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w900,
                                      fontSize: 18,
                                      color: Colors.blueAccent,
                                    ),
                                  ),
                                  Text(
                                    currency.format(_remainingAmount),
                                    style: GoogleFonts.outfit(
                                      fontSize: 15,
                                      color: const Color(0xFF1E293B),
                                      fontWeight: FontWeight.w900,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 20),
                        const Divider(),
                        const SizedBox(height: 16),
                        _row(Icons.storefront_rounded, 'Toko', storeName),
                        const SizedBox(height: 10),
                        _row(Icons.tag_rounded, 'No. Faktur', receiptNo),
                        if (orderData['installments'] != null &&
                            (orderData['installments'] as List).isNotEmpty) ...[
                          const SizedBox(height: 16),
                          const Divider(),
                          const SizedBox(height: 16),
                          Text(
                            'JADWAL CICILAN',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
                              color: Colors.blueGrey,
                              letterSpacing: 1.5,
                            ),
                          ),
                          const SizedBox(height: 8),
                          ...(orderData['installments'] as List).map((inst) {
                            final isPaid = inst['status'] == 'PAID';
                            final due =
                                DateTime.tryParse(
                                  inst['due_date']?.toString() ?? '',
                                ) ??
                                DateTime.now();
                            return Container(
                              margin: const EdgeInsets.only(bottom: 8),
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: isPaid
                                    ? Colors.green.withOpacity(0.05)
                                    : const Color(0xFFF8FAFC),
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(
                                  color: isPaid
                                      ? Colors.green.withOpacity(0.3)
                                      : Colors.grey.shade200,
                                ),
                              ),
                              child: Row(
                                children: [
                                  Container(
                                    padding: const EdgeInsets.all(8),
                                    decoration: BoxDecoration(
                                      color: isPaid
                                          ? Colors.green
                                          : Colors.orange,
                                      shape: BoxShape.circle,
                                    ),
                                    child: Text(
                                      '${inst['term_number']}',
                                      style: GoogleFonts.outfit(
                                        color: Colors.white,
                                        fontWeight: FontWeight.w900,
                                        fontSize: 12,
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          currency.format(
                                            double.tryParse(
                                              inst['amount']?.toString() ?? '0',
                                            ),
                                          ),
                                          style: GoogleFonts.outfit(
                                            fontWeight: FontWeight.w800,
                                            fontSize: 14,
                                            color: const Color(0xFF1E293B),
                                          ),
                                        ),
                                        Text(
                                          DateFormat('dd MMM yyyy').format(due),
                                          style: GoogleFonts.outfit(
                                            fontSize: 11,
                                            color: Colors.blueGrey,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 4,
                                    ),
                                    decoration: BoxDecoration(
                                      color: isPaid
                                          ? Colors.green
                                          : Colors.orange.withOpacity(0.1),
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: Text(
                                      isPaid ? 'LUNAS' : 'BELUM',
                                      style: GoogleFonts.outfit(
                                        fontSize: 10,
                                        fontWeight: FontWeight.w900,
                                        color: isPaid
                                            ? Colors.white
                                            : Colors.orange,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            );
                          }),
                        ],
                      ],
                    ),
                  ),

                  const SizedBox(height: 24),

                  // Nominal Input
                  Container(
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(28),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.04),
                          blurRadius: 20,
                          offset: const Offset(0, 8),
                        ),
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'NOMINAL DIBAYAR',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            fontWeight: FontWeight.w900,
                            color: Colors.blueGrey,
                            letterSpacing: 1.5,
                          ),
                        ),
                        const SizedBox(height: 16),
                        TextField(
                          controller: _amountController,
                          keyboardType: TextInputType.number,
                          style: GoogleFonts.outfit(
                            fontSize: 20,
                            fontWeight: FontWeight.w900,
                            color: const Color(0xFF1E293B),
                          ),
                          decoration: InputDecoration(
                            prefixText: 'Rp ',
                            prefixStyle: GoogleFonts.outfit(
                              fontSize: 20,
                              fontWeight: FontWeight.w900,
                              color: Colors.blueGrey,
                            ),
                            filled: true,
                            fillColor: Colors.grey.shade50,
                            contentPadding: const EdgeInsets.all(20),
                            enabledBorder: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(16),
                              borderSide: BorderSide(
                                color: Colors.grey.shade200,
                              ),
                            ),
                            focusedBorder: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(16),
                              borderSide: const BorderSide(
                                color: Colors.blueAccent,
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 24),

                  // Payment Method selection
                  Container(
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(28),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.04),
                          blurRadius: 20,
                          offset: const Offset(0, 8),
                        ),
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'METODE PEMBAYARAN',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            fontWeight: FontWeight.w900,
                            color: Colors.blueGrey,
                            letterSpacing: 1.5,
                          ),
                        ),
                        const SizedBox(height: 16),
                        ..._paymentOptions.entries.map((e) {
                          final isSelected = _selectedPaymentMethod == e.key;
                          return GestureDetector(
                            onTap: () => setState(() {
                              _selectedPaymentMethod = e.key;
                            }),
                            child: Container(
                              margin: const EdgeInsets.only(bottom: 12),
                              padding: const EdgeInsets.all(16),
                              decoration: BoxDecoration(
                                color: isSelected
                                    ? e.value['color'].withOpacity(0.05)
                                    : const Color(0xFFF8FAFC),
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(
                                  color: isSelected
                                      ? e.value['color']
                                      : Colors.transparent,
                                  width: 2,
                                ),
                              ),
                              child: Row(
                                children: [
                                  Icon(
                                    e.value['icon'],
                                    color: isSelected
                                        ? e.value['color']
                                        : Colors.blueGrey,
                                    size: 24,
                                  ),
                                  const SizedBox(width: 16),
                                  Expanded(
                                    child: Text(
                                      e.value['label'],
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w800,
                                        fontSize: 15,
                                        color: const Color(0xFF1E293B),
                                      ),
                                    ),
                                  ),
                                  if (isSelected)
                                    Icon(
                                      Icons.check_circle_rounded,
                                      color: e.value['color'],
                                      size: 20,
                                    ),
                                ],
                              ),
                            ),
                          );
                        }),
                        if (_selectedPaymentMethod == 'TRANSFER') ...[
                          const SizedBox(height: 16),
                          Text(
                            'PILIH REKENING TUJUAN TRANSFER',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
                              color: Colors.blueGrey,
                              letterSpacing: 1.2,
                            ),
                          ),
                          const SizedBox(height: 8),
                          if (_isLoadingBanks)
                            const Center(
                              child: Padding(
                                padding: EdgeInsets.symmetric(vertical: 12),
                                child: CircularProgressIndicator(),
                              ),
                            )
                          else if (_bankAccounts.isNotEmpty) ...[
                            Container(
                              margin: const EdgeInsets.only(bottom: 12),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 14,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(14),
                                border: Border.all(
                                  color: Colors.indigo.shade200,
                                  width: 1.5,
                                ),
                                boxShadow: [
                                  BoxShadow(
                                    color: Colors.indigo.withOpacity(0.05),
                                    blurRadius: 8,
                                    offset: const Offset(0, 2),
                                  ),
                                ],
                              ),
                              child: DropdownButtonHideUnderline(
                                child: DropdownButton<String>(
                                  isExpanded: true,
                                  value: _selectedBankAccount?['id']
                                      ?.toString(),
                                  hint: Text(
                                    'Pilih Rekening Bank',
                                    style: GoogleFonts.outfit(
                                      fontSize: 13,
                                      color: Colors.grey,
                                    ),
                                  ),
                                  icon: const Icon(
                                    Icons.keyboard_arrow_down_rounded,
                                    color: Colors.indigo,
                                  ),
                                  items: _bankAccounts.map((b) {
                                    final bName =
                                        b['bank_name']?.toString() ?? 'Bank';
                                    final accNum =
                                        b['account_number']?.toString() ?? '';
                                    final companyName =
                                        b['company']?['name']?.toString() ??
                                        b['account_holder']?.toString() ??
                                        '';
                                    final branchName = b['branch']?['name']
                                        ?.toString();
                                    final tag =
                                        branchName != null &&
                                            branchName.isNotEmpty
                                        ? ' - $branchName'
                                        : '';
                                    final displayName = companyName.isNotEmpty
                                        ? '$accNum - $companyName$tag ($bName)'
                                        : '$accNum$tag ($bName)';
                                    return DropdownMenuItem<String>(
                                      value: b['id']?.toString(),
                                      child: Row(
                                        children: [
                                          const Icon(
                                            Icons.account_balance_rounded,
                                            size: 18,
                                            color: Colors.indigo,
                                          ),
                                          const SizedBox(width: 8),
                                          Expanded(
                                            child: Text(
                                              displayName,
                                              style: GoogleFonts.outfit(
                                                fontSize: 13,
                                                fontWeight: FontWeight.w600,
                                                color: const Color(0xFF1E293B),
                                              ),
                                              overflow: TextOverflow.ellipsis,
                                            ),
                                          ),
                                        ],
                                      ),
                                    );
                                  }).toList(),
                                  onChanged: (val) {
                                    if (val != null) {
                                      setState(() {
                                        _selectedBankAccount = _bankAccounts
                                            .firstWhere(
                                              (b) => b['id']?.toString() == val,
                                              orElse: () => _bankAccounts.first,
                                            );
                                        _activeBankAccount =
                                            _selectedBankAccount;
                                      });
                                    }
                                  },
                                ),
                              ),
                            ),
                          ],
                          Container(
                            padding: const EdgeInsets.all(16),
                            decoration: BoxDecoration(
                              color: Colors.indigo.withOpacity(0.04),
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(
                                color: Colors.indigo.withOpacity(0.15),
                              ),
                            ),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  children: [
                                    _selectedBankAccount?['bank_logo_url'] !=
                                                null &&
                                            (_selectedBankAccount!['bank_logo_url']
                                                    as String)
                                                .isNotEmpty
                                        ? CachedNetworkImage(
                                            imageUrl:
                                                _selectedBankAccount!['bank_logo_url']
                                                    as String,
                                            width: 50,
                                            height: 25,
                                            placeholder: (context, url) =>
                                                const Center(
                                                  child:
                                                      CircularProgressIndicator(
                                                        strokeWidth: 2,
                                                      ),
                                                ),
                                            errorWidget:
                                                (context, url, error) =>
                                                    const Icon(
                                                      Icons.account_balance,
                                                      color: Colors.indigo,
                                                      size: 28,
                                                    ),
                                            memCacheWidth: 800,
                                          )
                                        : const Icon(
                                            Icons.account_balance_rounded,
                                            color: Colors.indigo,
                                            size: 28,
                                          ),
                                    const SizedBox(width: 12),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            _selectedBankAccount?['bank_name']
                                                    as String? ??
                                                _activeBankAccount?['bank_name']
                                                    as String? ??
                                                'Bank Central Asia (BCA)',
                                            style: GoogleFonts.outfit(
                                              fontWeight: FontWeight.bold,
                                              fontSize: 14,
                                              color: const Color(0xFF1E293B),
                                            ),
                                          ),
                                          if (_selectedBankAccount?['company'] !=
                                              null)
                                            Text(
                                              'PT: ${_selectedBankAccount!['company']['name']}',
                                              style: GoogleFonts.outfit(
                                                fontSize: 11,
                                                color: Colors.indigo.shade700,
                                                fontWeight: FontWeight.w600,
                                              ),
                                            ),
                                          if (_selectedBankAccount?['branch'] !=
                                              null)
                                            Text(
                                              'Cabang: ${_selectedBankAccount!['branch']['name']}',
                                              style: GoogleFonts.outfit(
                                                fontSize: 11,
                                                color: Colors.blueGrey.shade700,
                                                fontWeight: FontWeight.w500,
                                              ),
                                            ),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 12),
                                Row(
                                  mainAxisAlignment:
                                      MainAxisAlignment.spaceBetween,
                                  children: [
                                    Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          'NOMOR REKENING',
                                          style: GoogleFonts.outfit(
                                            fontSize: 9,
                                            color: Colors.blueGrey,
                                            fontWeight: FontWeight.w800,
                                            letterSpacing: 0.5,
                                          ),
                                        ),
                                        const SizedBox(height: 2),
                                        Text(
                                          _selectedBankAccount?['account_number']
                                                  as String? ??
                                              _activeBankAccount?['account_number']
                                                  as String? ??
                                              '817-529-3021',
                                          style: GoogleFonts.outfit(
                                            fontSize: 16,
                                            fontWeight: FontWeight.w900,
                                            color: const Color(0xFF1E293B),
                                          ),
                                        ),
                                      ],
                                    ),
                                    IconButton(
                                      icon: const Icon(
                                        Icons.copy_rounded,
                                        color: Colors.indigo,
                                      ),
                                      tooltip: 'Salin Rekening',
                                      onPressed: () {
                                        final accNo =
                                            (_selectedBankAccount?['account_number']
                                                        as String? ??
                                                    _activeBankAccount?['account_number']
                                                        as String? ??
                                                    '8175293021')
                                                .replaceAll('-', '')
                                                .replaceAll(' ', '');
                                        Clipboard.setData(
                                          ClipboardData(text: accNo),
                                        );
                                        ScaffoldMessenger.of(
                                          context,
                                        ).showSnackBar(
                                          const SnackBar(
                                            content: Text(
                                              'Nomor rekening disalin!',
                                            ),
                                            duration: Duration(seconds: 1),
                                          ),
                                        );
                                      },
                                    ),
                                  ],
                                ),
                                const Divider(height: 16),
                                Text(
                                  'NAMA PENERIMA',
                                  style: GoogleFonts.outfit(
                                    fontSize: 9,
                                    color: Colors.blueGrey,
                                    fontWeight: FontWeight.w800,
                                    letterSpacing: 0.5,
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  _selectedBankAccount?['account_holder']
                                          as String? ??
                                      _activeBankAccount?['account_holder']
                                          as String? ??
                                      'Perusahaan Demo',
                                  style: GoogleFonts.outfit(
                                    fontSize: 13,
                                    fontWeight: FontWeight.bold,
                                    color: const Color(0xFF1E293B),
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 24),
                          Text(
                            'BUKTI TRANSFER BANK',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
                              color: Colors.blueGrey,
                              letterSpacing: 1.5,
                            ),
                          ),
                          const SizedBox(height: 12),
                          GestureDetector(
                            onTap: () {
                              showModalBottomSheet(
                                context: context,
                                shape: const RoundedRectangleBorder(
                                  borderRadius: BorderRadius.vertical(
                                    top: Radius.circular(20),
                                  ),
                                ),
                                builder: (context) => SafeArea(
                                  child: Wrap(
                                    children: [
                                      ListTile(
                                        leading: const Icon(
                                          Icons.camera_alt_rounded,
                                          color: Colors.indigo,
                                        ),
                                        title: Text(
                                          'Ambil Foto Kamera',
                                          style: GoogleFonts.outfit(),
                                        ),
                                        onTap: () {
                                          Navigator.pop(context);
                                          _takeTransferProofPhoto();
                                        },
                                      ),
                                      ListTile(
                                        leading: const Icon(
                                          Icons.photo_library_rounded,
                                          color: Colors.indigo,
                                        ),
                                        title: Text(
                                          'Pilih dari Galeri / Screenshot',
                                          style: GoogleFonts.outfit(),
                                        ),
                                        onTap: () {
                                          Navigator.pop(context);
                                          _pickTransferProofPhoto();
                                        },
                                      ),
                                    ],
                                  ),
                                ),
                              );
                            },
                            child: Container(
                              height: 150,
                              width: double.infinity,
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(20),
                                border: Border.all(
                                  color: _transferProofImage != null
                                      ? Colors.indigo
                                      : Colors.grey.shade300,
                                  width: 2,
                                  style: BorderStyle.solid,
                                ),
                              ),
                              child: _transferProofImage != null
                                  ? Stack(
                                      children: [
                                        Positioned.fill(
                                          child: ClipRRect(
                                            borderRadius: BorderRadius.circular(
                                              18,
                                            ),
                                            child: kIsWeb
                                                ? CachedNetworkImage(
                                                    imageUrl:
                                                        _transferProofImage!
                                                            .path,
                                                    fit: BoxFit.cover,
                                                    placeholder:
                                                        (
                                                          context,
                                                          url,
                                                        ) => const Center(
                                                          child:
                                                              CircularProgressIndicator(),
                                                        ),
                                                    errorWidget:
                                                        (context, url, error) =>
                                                            const Icon(
                                                              Icons.error,
                                                            ),
                                                    memCacheWidth: 800,
                                                  )
                                                : Image.file(
                                                    File(
                                                      _transferProofImage!.path,
                                                    ),
                                                    fit: BoxFit.cover,
                                                  ),
                                          ),
                                        ),
                                        Positioned(
                                          top: 8,
                                          right: 8,
                                          child: CircleAvatar(
                                            backgroundColor: Colors.black
                                                .withOpacity(0.6),
                                            child: IconButton(
                                              icon: const Icon(
                                                Icons.close,
                                                color: Colors.white,
                                              ),
                                              onPressed: () => setState(
                                                () =>
                                                    _transferProofImage = null,
                                              ),
                                            ),
                                          ),
                                        ),
                                      ],
                                    )
                                  : Column(
                                      mainAxisAlignment:
                                          MainAxisAlignment.center,
                                      children: [
                                        const Icon(
                                          Icons.cloud_upload_outlined,
                                          size: 40,
                                          color: Colors.blueGrey,
                                        ),
                                        const SizedBox(height: 8),
                                        Text(
                                          'Upload Bukti Transfer',
                                          style: GoogleFonts.outfit(
                                            fontWeight: FontWeight.bold,
                                            color: const Color(0xFF1E293B),
                                          ),
                                        ),
                                        Text(
                                          'Format JPG/PNG (Galeri/Kamera)',
                                          style: GoogleFonts.outfit(
                                            fontSize: 11,
                                            color: Colors.grey,
                                          ),
                                        ),
                                      ],
                                    ),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),

                  const SizedBox(height: 24),

                  // Signature pads for collection proof
                  Container(
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(28),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.04),
                          blurRadius: 20,
                          offset: const Offset(0, 8),
                        ),
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'BUKTI TANDA TANGAN',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            fontWeight: FontWeight.w900,
                            color: Colors.blueGrey,
                            letterSpacing: 1.5,
                          ),
                        ),
                        const SizedBox(height: 16),
                        SignatureCanvas(
                          label: 'TANDA TANGAN PENAGIH',
                          points: _courierSignaturePoints,
                          onDraw: (pts) =>
                              setState(() => _courierSignaturePoints = pts),
                          onClear: () =>
                              setState(() => _courierSignaturePoints.clear()),
                        ),
                        const SizedBox(height: 20),
                        SignatureCanvas(
                          label: _selectedPaymentMethod == 'TRANSFER'
                              ? 'TANDA TANGAN PELANGGAN (OPSIONAL UNTUK TRANSFER)'
                              : 'TANDA TANGAN PELANGGAN',
                          points: _customerSignaturePoints,
                          onDraw: (pts) =>
                              setState(() => _customerSignaturePoints = pts),
                          onClear: () =>
                              setState(() => _customerSignaturePoints.clear()),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 40),
                  // Checkout button
                  SizedBox(
                    width: double.infinity,
                    height: 72,
                    child: ElevatedButton(
                      onPressed: _isSubmitting ? null : _submitPayment,
                      style: ElevatedButton.styleFrom(
                        backgroundColor:
                            _paymentOptions[_selectedPaymentMethod]!['color'],
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(24),
                        ),
                        elevation: 12,
                        shadowColor:
                            _paymentOptions[_selectedPaymentMethod]!['color']
                                .withOpacity(0.4),
                      ),
                      child: _isSubmitting
                          ? const CircularProgressIndicator(color: Colors.white)
                          : Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Text(
                                  'SIMPAN PEMBAYARAN',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w900,
                                    fontSize: 18,
                                    color: Colors.white,
                                  ),
                                ),
                                Text(
                                  'Metode: ${_paymentOptions[_selectedPaymentMethod]!['label']}',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    color: Colors.white70,
                                  ),
                                ),
                              ],
                            ),
                    ),
                  ),
                  const SizedBox(height: 40),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
