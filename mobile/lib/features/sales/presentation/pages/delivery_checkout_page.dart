import 'package:hris_app/core/network/api_client.dart';
import 'package:get_it/get_it.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'dart:io' show File, Directory;
import 'dart:convert';
import 'dart:ui' as ui;
import 'package:hris_app/core/utils/constants.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/services/offline_delivery_manager.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:share_plus/share_plus.dart';
import 'package:hris_app/core/widgets/qris_sandbox_warning.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'return_selection_page.dart';
import 'dynamic_invoice_page.dart';

class DeliveryCheckoutPage extends StatefulWidget {
  final Map<String, dynamic> deliveryItem;
  final Map<String, dynamic>? batchData; // Menambahkan batchData
  final VoidCallback onSuccess;

  const DeliveryCheckoutPage({
    super.key,
    required this.deliveryItem,
    this.batchData,
    required this.onSuccess,
  });

  @override
  State<DeliveryCheckoutPage> createState() => _DeliveryCheckoutPageState();
}

class _DeliveryCheckoutPageState extends State<DeliveryCheckoutPage> {
  final ApiClient apiClient = di.sl<ApiClient>();
  final ImagePicker _picker = ImagePicker();

  bool get isMt =>
      widget.deliveryItem['sales_order']?['is_mt'] == true ||
      widget.deliveryItem['is_mt'] == true;

  bool _isLoading = false;
  bool _isReadOnly = false;
  XFile? _podImage;
  XFile? _transferProofImage;
  String? _existingPodImageUrl;
  Map<String, dynamic>? _midtransData;
  List<Offset> _courierSignaturePoints = [];
  List<Offset> _customerSignaturePoints = [];
  bool _isPaymentConfirmed = false;
  Map<String, dynamic>? _activeBankAccount;
  List<Map<String, dynamic>> _bankAccounts = [];
  Map<String, dynamic>? _selectedBankAccount;
  bool _isLoadingBanks = false;

  // Variables for Invoice
  late Map<String, dynamic> orderData;
  List<dynamic> _items = [];
  final Map<String, int> _returnedQuantities = {};
  final Map<String, String> _returnReasons = {}; // productId → alasan retur
  final Map<String, String> _returnTypes = {}; // productId → returnType
  final Map<String, String?> _replacementProductIds =
      {}; // productId → replacementProductId
  final Map<String, String?> _replacementProductNames =
      {}; // productId → replacementProductName
  final Map<String, double?> _replacementPrices =
      {}; // productId → replacementPrice
  final Map<String, String> _claimRefNos = {}; // productId → claimRefNo
  final Map<String, int> _replacementQuantities =
      {}; // productId → replacement quantity
  List<dynamic> _availableProducts = [];
  final Map<String, TextEditingController> _claimControllers = {};
  final Map<String, int> _orderRequests =
      {}; // productId → order request quantity

  List<ReturnItemResult> get _returnResults {
    final results = <ReturnItemResult>[];
    for (final item in _items) {
      final pId = (item['product_id'] ?? item['id'])?.toString() ?? '';
      final qty = _returnedQuantities[pId] ?? 0;
      if (qty <= 0) continue;

      final type = _returnTypes[pId] ?? 'POTONG_NOTA';
      final reason = _returnReasons[pId] ?? 'RUSAK';
      final replacementId = _replacementProductIds[pId];
      final replacementName = _replacementProductNames[pId];
      final replacementPrice = _replacementPrices[pId];
      final claimRefNo = _claimRefNos[pId] ?? '';
      final replQty = _replacementQuantities[pId] ?? qty;

      results.add(
        ReturnItemResult(
          productId: pId,
          productName: item['product']?['name'] ?? 'Produk',
          returnedQuantity: qty,
          returnReason: reason,
          returnType: type,
          orderedQuantity: item['ordered_quantity'] ?? 0,
          price: (item['price'] as num? ?? 0.0).toDouble(),
          unit: item['unit'] ?? 'PCS',
          piecesPerUnit: item['pieces_per_unit'] ?? 1,
          replacementProductId: replacementId,
          replacementProductName: replacementName,
          replacementQuantity: replacementId != null ? replQty : null,
          replacementPrice: replacementPrice,
          claimReferenceNo: claimRefNo.isNotEmpty ? claimRefNo : null,
        ),
      );
    }
    return results;
  }

  static const List<String> _returnReasonOptions = [
    'RUSAK',
    'TOLAK_CUSTOMER',
    'KADALUARSA',
    'LAINNYA',
  ];
  static const Map<String, String> _returnReasonLabels = {
    'RUSAK': '🔴 Barang Rusak',
    'TOLAK_CUSTOMER': '🚫 Ditolak Customer',
    'KADALUARSA': '⏰ Kadaluarsa',
    'LAINNYA': '📝 Lainnya',
  };

  static const Map<String, Map<String, dynamic>> _returnTypeConfigs = {
    'POTONG_NOTA': {
      'label': 'Potong Nota',
      'description': 'Kurangi tagihan nota saat ini.',
      'icon': Icons.money_off_rounded,
      'color': Colors.red,
    },
    'TUKAR_PRODUK': {
      'label': 'Tukar Langsung',
      'description': 'Tukar dari armada sekarang.',
      'icon': Icons.swap_horizontal_circle_rounded,
      'color': Colors.cyan,
    },
  };

  // Variables for Payment
  final double _paymentAmount = 0.0;
  String _paymentMethod = 'CASH'; // CASH, QRIS, VA, TEMPO
  String? _selectedBank; // bca, bni, bri, mandiri
  DateTime? _paymentDueDate;
  int? _selectedTempoDays = 14;
  final TextEditingController _receivedByController = TextEditingController();
  final TextEditingController _amountController = TextEditingController();
  final TextEditingController _notesController = TextEditingController();

  static const Map<String, Map<String, dynamic>> _paymentOptions = {
    'CASH': {
      'label': 'Tunai (Cash)',
      'icon': Icons.payments_rounded,
      'color': Colors.green,
    },
    'VA': {
      'label': 'Transfer Bank (Satu Pintu)',
      'icon': Icons.account_balance_rounded,
      'color': Colors.indigo,
    },
    'TEMPO': {
      'label': 'Tempo / Kredit',
      'icon': Icons.history_rounded,
      'color': Colors.orange,
    },
  };

  static const List<Map<String, String>> _bankOptions = [
    {'id': 'bca', 'name': 'BCA Virtual Account'},
    {'id': 'bni', 'name': 'BNI Virtual Account'},
    {'id': 'bri', 'name': 'BRI Virtual Account'},
    {'id': 'mandiri', 'name': 'Mandiri Bill Payment'},
  ];

  Future<void> _fetchAvailableProducts() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      String? companyId =
          orderData['company_id']?.toString() ??
          orderData['company']?['id']?.toString() ??
          orderData['employee']?['company_id']?.toString() ??
          orderData['employee']?['company']?['id']?.toString() ??
          prefs.getString('factory_company_id') ??
          prefs.getString('company_id');

      String? branchId =
          orderData['branch_id']?.toString() ??
          orderData['branch']?['id']?.toString() ??
          orderData['employee']?['branch_id']?.toString() ??
          orderData['employee']?['branch']?['id']?.toString() ??
          prefs.getString('branch_id');

      final Map<String, dynamic> queryParams = {};

      if (companyId != null &&
          companyId.isNotEmpty &&
          companyId != '00000000-0000-0000-0000-000000000000') {
        queryParams['company_id'] = companyId;
      }

      if (branchId != null &&
          branchId.isNotEmpty &&
          branchId != '00000000-0000-0000-0000-000000000000') {
        queryParams['branch_id'] = branchId;
      }

      final response = await apiClient.client.get(
        'factory/products',
        queryParameters: queryParams.isNotEmpty ? queryParams : null,
      );
      final resData = response.data is Map && response.data.containsKey('data')
          ? response.data['data']
          : response.data;
      if (resData is List && mounted) {
        setState(() {
          _availableProducts = resData;
        });
      }
    } catch (_) {}
  }

  @override
  void initState() {
    super.initState();
    orderData =
        widget.deliveryItem['sales_transaction'] ??
        widget.deliveryItem['sales_order'];
    _items = orderData['items'] ?? [];

    _initializeReadOnly();
    _fetchActiveBankAccount();
    _fetchAvailableProducts();

    // Populate initial maps if there are existing returns
    final itemsList = orderData['items'] as List? ?? [];
    for (var item in itemsList) {
      final pId = (item['product_id'] ?? item['id'])?.toString() ?? '';
      if (item['return_type'] != null) {
        _returnTypes[pId] = item['return_type'].toString();
      }
      if (item['claim_reference_no'] != null) {
        _claimRefNos[pId] = item['claim_reference_no'].toString();
      }
      if (item['replacement_product_id'] != null) {
        _replacementProductIds[pId] = item['replacement_product_id'].toString();
        _replacementProductNames[pId] = item['replacement_product']?['name']
            ?.toString();
        _replacementPrices[pId] = (item['replacement_price'] as num? ?? 0.0)
            .toDouble();
        if (item['replacement_quantity'] != null) {
          _replacementQuantities[pId] = (item['replacement_quantity'] as num)
              .toInt();
        }
      }
    }

    // Calculate and set initial amount
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _updateAmountFromTotal();
    });
  }

  final PageController _pageController = PageController();
  int _currentStep = 0;

  void _initializeReadOnly() {
    if (widget.deliveryItem['status'] == 'DELIVERED') {
      _isReadOnly = true;
      _currentStep = 3;

      // Populate existing data
      _receivedByController.text = widget.deliveryItem['received_by'] ?? '';
      _notesController.text = widget.deliveryItem['notes'] ?? '';
      _existingPodImageUrl = widget.deliveryItem['pod_image_url'];
      final serverMethod = widget.deliveryItem['payment_method']
          ?.toString()
          .toUpperCase();
      _paymentMethod = _paymentOptions.containsKey(serverMethod)
          ? serverMethod!
          : 'CASH';
      _amountController.text = (widget.deliveryItem['payment_amount'] ?? 0)
          .toString();
      if (widget.deliveryItem['payment_due_date'] != null) {
        _paymentDueDate = DateTime.tryParse(
          widget.deliveryItem['payment_due_date'].toString(),
        );
      } else if (orderData['payment_due_date'] != null) {
        _paymentDueDate = DateTime.tryParse(
          orderData['payment_due_date'].toString(),
        );
      }

      // Populate returns if any
      final items = orderData['items'] as List? ?? [];
      for (var item in items) {
        final ordered = item['ordered_quantity'] ?? 0;
        int actual = item['actual_quantity'] ?? ordered;
        // If actual_quantity is 0, it likely means the backend hasn't populated it
        // (default 0), so we should not treat everything as returned.
        if (actual == 0) {
          actual = ordered;
        }
        final pId = (item['product_id'] ?? item['id'])?.toString() ?? '';
        if (actual < ordered) {
          _returnedQuantities[pId] = ordered - actual;
        }
        if (item['return_reason'] != null) {
          _returnReasons[pId] = item['return_reason'].toString();
        }
        if (item['return_type'] != null) {
          _returnTypes[pId] = item['return_type'].toString();
        }
        if (item['claim_reference_no'] != null) {
          _claimRefNos[pId] = item['claim_reference_no'].toString();
        }
        if (item['replacement_product_id'] != null) {
          _replacementProductIds[pId] = item['replacement_product_id']
              .toString();
          _replacementProductNames[pId] = item['replacement_product']?['name']
              ?.toString();
          _replacementPrices[pId] = (item['replacement_price'] as num? ?? 0.0)
              .toDouble();
          if (item['replacement_quantity'] != null) {
            _replacementQuantities[pId] = (item['replacement_quantity'] as num)
                .toInt();
          }
        }
      }

      // Ensure PageController starts at the last step
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (_pageController.hasClients) {
          _pageController.jumpToPage(3);
        }
      });
    }
  }

  @override
  void dispose() {
    _pageController.dispose();
    _receivedByController.dispose();
    _amountController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _fetchBankAccounts() async {
    setState(() => _isLoadingBanks = true);
    try {
      final prefs = await SharedPreferences.getInstance();
      String? branchId =
          orderData['branch_id']?.toString() ??
          orderData['employee']?['branch_id']?.toString() ??
          prefs.getString('branch_id');
      String? companyId =
          orderData['company_id']?.toString() ??
          orderData['company']?['id']?.toString() ??
          orderData['employee']?['company_id']?.toString() ??
          orderData['employee']?['company']?['id']?.toString() ??
          prefs.getString('company_id');

      final salesApi = SalesApiService(apiClient: di.sl<ApiClient>());
      final banks = await salesApi.getBankAccounts(branchId: branchId);
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

  void _nextPage() {
    if (_pageController.hasClients && _currentStep < 3) {
      _pageController.nextPage(
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeInOut,
      );
    }
  }

  void _prevPage() {
    if (_pageController.hasClients && _currentStep > 0) {
      _pageController.previousPage(
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeInOut,
      );
    }
  }

  void _updateAmountFromTotal() {
    setState(() {
      _amountController.text = _calculatedTotal.toInt().toString();
    });
  }

  double get _originalTotal {
    if (orderData.containsKey('total_amount'))
      return (orderData['total_amount'] ?? 0).toDouble();
    double total = 0;
    for (var item in _items) {
      int qty = item['ordered_quantity'] ?? 0;
      int ppu = item['pieces_per_unit'] ?? 1;
      double price = (item['price'] ?? 0).toDouble();
      total += (qty * ppu) * price;
    }
    return total;
  }

  double get _calculatedTotal {
    double total = (orderData['total_amount'] ?? 0).toDouble();

    for (final r in _returnResults) {
      if (r.returnType == 'POTONG_NOTA') {
        total -= r.returnedQuantity * r.piecesPerUnit * r.price;
      }
      // TUKAR_PRODUK is removed here because it should not deduct the invoice.
    }

    return total > 0 ? total : 0;
  }

  Future<Uint8List?> _renderSignatureToPng(List<Offset> points) async {
    if (points.isEmpty) return null;
    final recorder = ui.PictureRecorder();
    final canvas = ui.Canvas(recorder, const Rect.fromLTWH(0, 0, 300, 150));

    // Draw white background
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

    return pngBytes?.buffer.asUint8List();
  }

  Future<void> _generateAndShareInvoicePdf(
    Uint8List? courierSigBytes,
    Uint8List? customerSigBytes,
  ) async {
    final pdf = pw.Document();
    final currencyFormat = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    // Fetch store/invoice info
    final storeName = orderData['store']?['name'] ?? 'Toko';
    final storeAddress = orderData['store']?['address'] ?? 'Alamat Toko';
    final receiptNo =
        orderData['invoice_no'] ??
        orderData['so_number'] ??
        'INV-${DateTime.now().millisecondsSinceEpoch}';
    final dateStr = DateFormat('dd MMMM yyyy').format(DateTime.now());
    final paymentMethodLabel =
        _paymentOptions[_paymentMethod]?['label'] ?? _paymentMethod;
    final payAmount = double.tryParse(_amountController.text) ?? 0.0;

    // Fetch dynamic company details
    final companyName =
        orderData['employee']?['company']?['name'] ??
        orderData['company']?['name'] ??
        'Perusahaan Demo';
    final companyAddress =
        orderData['company']?['address'] ??
        'Kawasan Industri & Distribusi Sentra Prima';
    final companyTaxNumber = orderData['company']?['tax_number'] ?? '';

    // Draw signatures
    pw.Widget? courierSigWidget;
    pw.Widget? customerSigWidget;

    if (courierSigBytes != null) {
      courierSigWidget = pw.Image(
        pw.MemoryImage(courierSigBytes),
        height: 35,
        width: 70,
      );
    }
    if (customerSigBytes != null) {
      customerSigWidget = pw.Image(
        pw.MemoryImage(customerSigBytes),
        height: 35,
        width: 70,
      );
    }

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
                      if (companyTaxNumber.isNotEmpty)
                        pw.Text(
                          'NPWP: $companyTaxNumber',
                          style: const pw.TextStyle(fontSize: 7),
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
              pw.Divider(thickness: 1.0, color: PdfColors.grey400, height: 10),
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
                        paymentMethodLabel.toUpperCase(),
                        style: pw.TextStyle(
                          fontSize: 8.5,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                      if (_paymentMethod == 'TEMPO' && _paymentDueDate != null)
                        pw.Text(
                          'Jatuh Tempo: ${DateFormat('dd MMM yyyy').format(_paymentDueDate!)}',
                          style: pw.TextStyle(
                            fontSize: 7,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.black,
                          ),
                        ),
                      pw.Text(
                        'Status: ${_paymentMethod == 'TEMPO' ? 'BELUM LUNAS (TEMPO)' : (_paymentMethod == 'VA' || _paymentMethod == 'QRIS' ? 'MENUNGGU PEMBAYARAN' : 'LUNAS')}',
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
              ..._items.map((item) {
                final prodName = item['product']?['name'] ?? 'Produk';
                final orderedQty = item['ordered_quantity'] ?? 0;
                final returQty =
                    _returnedQuantities[item['product_id']?.toString()] ?? 0;
                final actualQty = orderedQty - returQty;
                final price =
                    (item['price_at_transaction'] ?? item['price'] ?? 0)
                        .toDouble();
                final subtotal = actualQty * price;

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
                          '$actualQty',
                          style: const pw.TextStyle(fontSize: 7),
                          textAlign: pw.TextAlign.center,
                        ),
                      ),
                      pw.Expanded(
                        flex: 2,
                        child: pw.Text(
                          currencyFormat.format(price),
                          style: const pw.TextStyle(fontSize: 7),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                      pw.Expanded(
                        flex: 2,
                        child: pw.Text(
                          currencyFormat.format(subtotal),
                          style: const pw.TextStyle(fontSize: 7),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                );
              }),

              // Bonus Items (Hadiah)
              if (orderData['bonus_items'] != null &&
                  (orderData['bonus_items'] as List).isNotEmpty) ...[
                pw.SizedBox(height: 6),
                pw.Container(
                  padding: const pw.EdgeInsets.symmetric(
                    vertical: 2,
                    horizontal: 6,
                  ),
                  color: PdfColors.grey200,
                  child: pw.Row(
                    children: [
                      pw.Expanded(
                        flex: 3,
                        child: pw.Text(
                          'Barang Bonus / Hadiah',
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
                        flex: 4,
                        child: pw.Text(
                          'Status',
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
                ...(orderData['bonus_items'] as List).map((bonus) {
                  final prodName = bonus['bonus_inventory']?['name'] ?? 'Bonus';
                  final qty = bonus['quantity'] ?? 0;
                  final unit = bonus['bonus_inventory']?['unit'] ?? 'PCS';

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
                            '$qty $unit',
                            style: const pw.TextStyle(fontSize: 7),
                            textAlign: pw.TextAlign.center,
                          ),
                        ),
                        pw.Expanded(
                          flex: 4,
                          child: pw.Text(
                            'Diserahkan',
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
                            currencyFormat.format(_calculatedTotal),
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
                            currencyFormat.format(payAmount),
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
                      pw.SizedBox(height: 2),
                      if (customerSigWidget != null)
                        customerSigWidget
                      else
                        pw.SizedBox(height: 20),
                      pw.SizedBox(height: 2),
                      pw.Text(
                        _receivedByController.text.isNotEmpty
                            ? _receivedByController.text
                            : 'Pelanggan',
                        style: const pw.TextStyle(fontSize: 7),
                      ),
                    ],
                  ),
                  pw.Column(
                    children: [
                      pw.Text(
                        'Kurir / Driver',
                        style: pw.TextStyle(
                          fontSize: 8,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                      pw.SizedBox(height: 2),
                      if (courierSigWidget != null)
                        courierSigWidget
                      else
                        pw.SizedBox(height: 20),
                      pw.SizedBox(height: 2),
                      pw.Text(
                        'Driver Delivery',
                        style: const pw.TextStyle(fontSize: 7),
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

    // Save and Share PDF
    try {
      final pdfBytes = await pdf.save();
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
      print('Gagal menghasilkan / membagikan PDF: $e');
    }
  }

  Future<void> _checkPaymentStatus() async {
    setState(() => _isLoading = true);
    try {
      final soId = orderData['id'] ?? widget.deliveryItem['sales_order_id'];
      final response = await apiClient.client.get('admin/sales-orders/$soId');
      final resData = response.data is Map && response.data.containsKey('data')
          ? response.data['data']
          : response.data;

      final paymentStatus = resData['payment_status']?.toString().toUpperCase();
      if (paymentStatus == 'PAID' || paymentStatus == 'SUCCESS') {
        setState(() {
          _isPaymentConfirmed = true;
        });
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text(
                'Pembayaran Terverifikasi! Silakan berikan tanda tangan.',
              ),
              backgroundColor: Colors.green,
            ),
          );
        }
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text(
                'Pembayaran belum diterima. Silakan scan QRIS atau bayar via VA.',
              ),
              backgroundColor: Colors.orange,
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal memeriksa status pembayaran: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  void _simulateQrisSuccess() {
    setState(() {
      _isPaymentConfirmed = true;
    });
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('SIMULASI SUKSES: Pembayaran berhasil terverifikasi!'),
        backgroundColor: Colors.green,
      ),
    );
  }

  Future<void> _handleStoreClosed() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        title: Row(
          children: [
            const Icon(Icons.store_rounded, color: Colors.red, size: 28),
            const SizedBox(width: 12),
            Text(
              'Konfirmasi Toko Tutup',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.bold,
                fontSize: 18,
              ),
            ),
          ],
        ),
        content: Text(
          'Apakah Anda yakin toko/outlet tujuan pengiriman ini tutup?\n\nSemua barang akan otomatis ditandai sebagai retur (kembali ke truk), tagihan menjadi Rp 0, dan pengiriman ditandai gagal karena toko tutup.',
          style: GoogleFonts.outfit(
            fontSize: 14,
            color: const Color(0xFF475569),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(
              'BATAL',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.bold,
                color: Colors.blueGrey,
              ),
            ),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(
              'YA, TOKO TUTUP',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.bold,
                color: Colors.white,
              ),
            ),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    // Prompt user to take a photo of the closed store as proof
    final XFile? photo = await showDialog<XFile?>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        title: Text(
          'Ambil Bukti Toko Tutup',
          style: GoogleFonts.outfit(fontWeight: FontWeight.bold, fontSize: 16),
        ),
        content: Text(
          'Silakan ambil foto outlet yang tutup/terkunci sebagai bukti pengiriman gagal.',
          style: GoogleFonts.outfit(
            fontSize: 13,
            color: const Color(0xFF475569),
          ),
        ),
        actions: [
          ElevatedButton.icon(
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.blueAccent,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            onPressed: () async {
              final XFile? photo = await _picker.pickImage(
                maxWidth: 1024,
                maxHeight: 1024,
                source: ImageSource.camera,
                imageQuality: 70,
              );
              Navigator.pop(ctx, photo);
            },
            icon: const Icon(Icons.camera_alt_rounded, color: Colors.white),
            label: Text(
              'KAMERA',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.bold,
                color: Colors.white,
              ),
            ),
          ),
        ],
      ),
    );

    if (photo == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Foto bukti toko tutup wajib diambil sebagai bukti.'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    setState(() {
      _podImage = photo;
      _isLoading = true;
    });

    try {
      final soId =
          orderData['id'] ??
          widget.deliveryItem['sales_order_id'] ??
          widget.deliveryItem['so_id'];

      // 1. Upload the closed store photo as POD
      String uploadedPodUrl = '';
      try {
        final MultipartFile filePayload;
        if (kIsWeb) {
          final bytes = await _podImage!.readAsBytes();
          filePayload = MultipartFile.fromBytes(
            bytes,
            filename: _podImage!.name,
          );
        } else {
          filePayload = await MultipartFile.fromFile(_podImage!.path);
        }
        final formData = FormData.fromMap({'file': filePayload});
        final uploadRes = await apiClient.client.post(
          'sales/upload',
          data: formData,
        );
        if (uploadRes.data != null && uploadRes.data['url'] != null) {
          uploadedPodUrl = uploadRes.data['url'];
        }
      } catch (e) {
        bool isConnectionError = true;
        if (e is DioException && e.type == DioExceptionType.badResponse) {
          isConnectionError = false;
        }
        if (isConnectionError) {
          rethrow;
        }
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal mengunggah foto bukti: $e'),
            backgroundColor: Colors.orange,
          ),
        );
        setState(() => _isLoading = false);
        return;
      }

      // 2. Prepare return payload for all items (qty returned = qty ordered)
      List<Map<String, dynamic>> returns = [];
      for (var item in _items) {
        String pId = item['product_id'] ?? item['id'] ?? '';
        int ordered = item['ordered_quantity'] ?? 0;
        returns.add({
          'product_id': pId,
          'returned_quantity': ordered,
          'return_reason': 'TOLAK_CUSTOMER', // Or generic reject
        });
      }

      // 3. Confirm Delivery to Backend
      await apiClient.client.post(
        'sales-orders/confirm-delivery',
        data: {
          'so_id': soId,
          'received_by': 'TOKO TUTUP',
          'pod_image_url': uploadedPodUrl,
          'signature_image_url': '', // No signature since store is closed
          'notes':
              'Toko Tutup / Gagal Kirim. Seluruh barang diretur ke armada.',
          'returned_items': returns,
        },
      );

      // 4. Save Payment of 0.0 with CASH to ensure clean status (Skipped for Toko Tutup because status is CANCELLED)
      // await apiClient.client.post( ... )

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'Status Toko Tutup & Retur Barang berhasil diproses.',
            ),
            backgroundColor: Colors.green,
          ),
        );
        widget.onSuccess();
        Navigator.pop(context, true);
      }
    } catch (e) {
      print('DEBUG TOKO TUTUP: Error during closed store submission: $e');

      bool isConnectionError = true;
      String errorMsg = e.toString();

      if (e is DioException) {
        if (e.type == DioExceptionType.badResponse) {
          isConnectionError = false;
          final responseData = e.response?.data;
          if (responseData is Map) {
            errorMsg =
                responseData['message'] ??
                responseData['error'] ??
                e.message ??
                e.toString();
          } else {
            errorMsg = e.message ?? e.toString();
          }
        }
      } else {
        isConnectionError = false;
      }

      if (!isConnectionError) {
        setState(() => _isLoading = false);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Gagal: $errorMsg'),
              backgroundColor: Colors.red,
            ),
          );
        }
        return;
      }

      // Offline fallback handling
      final bool saveOffline = await _showOfflineCheckoutDialog(e);
      if (saveOffline) {
        try {
          final soId =
              orderData['id'] ??
              widget.deliveryItem['sales_order_id'] ??
              widget.deliveryItem['so_id'];
          final storeName = orderData['store']?['name'] ?? 'Toko';

          List<Map<String, dynamic>> returns = [];
          for (var item in _items) {
            String pId = item['product_id'] ?? item['id'] ?? '';
            int ordered = item['ordered_quantity'] ?? 0;
            returns.add({
              'product_id': pId,
              'returned_quantity': ordered,
              'return_reason': 'TOLAK_CUSTOMER',
            });
          }

          final offlineItem = OfflineDeliveryItem(
            localId: 'offline_${DateTime.now().millisecondsSinceEpoch}',
            soId: soId.toString(),
            storeName: storeName,
            receivedBy: 'TOKO TUTUP',
            podImagePath: _podImage?.path ?? '',
            courierSigPath: '',
            customerSigPath: '',
            notes:
                'Toko Tutup / Gagal Kirim. Seluruh barang diretur ke armada.',
            returnedItems: returns,
            amount: 0.0,
            paymentMethod: 'CASH',
            paymentBank: null,
            createdAt: DateTime.now(),
          );

          await OfflineDeliveryManager().queueDelivery(offlineItem);

          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text(
                  'Offline Fallback Sukses! Laporan Toko Tutup disimpan lokal.',
                ),
                backgroundColor: Colors.blueAccent,
                duration: Duration(seconds: 4),
              ),
            );
            widget.onSuccess();
            Navigator.pop(context, true);
          }
        } catch (innerErr) {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('Gagal menyimpan offline: $innerErr'),
                backgroundColor: Colors.red,
              ),
            );
          }
        }
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _takePhoto() async {
    final XFile? photo = await _picker.pickImage(
      maxWidth: 1024,
      maxHeight: 1024,
      source: ImageSource.camera,
      imageQuality: 70,
    );
    if (photo != null) {
      setState(() {
        _podImage = photo;
      });
    }
  }

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

  Future<void> _submitCheckout() async {
    if (_isLoading) return;
    if (_podImage == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Foto bukti pengiriman (POD) wajib diambil'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    if (_paymentMethod == 'VA' && _transferProofImage == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Harap upload bukti transfer bank.'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    if (_paymentMethod == 'TEMPO' && _paymentDueDate == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Harap pilih estimasi tanggal bayar (Jatuh Tempo).'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final payAmount = double.tryParse(_amountController.text) ?? 0.0;
      final soId =
          orderData['id'] ??
          widget.deliveryItem['sales_order_id'] ??
          widget.deliveryItem['so_id'];

      // 1. Upload POD Image First
      String uploadedPodUrl = '';
      if (_podImage != null) {
        try {
          final MultipartFile filePayload;
          if (kIsWeb) {
            final bytes = await _podImage!.readAsBytes();
            filePayload = MultipartFile.fromBytes(
              bytes,
              filename: _podImage!.name,
            );
          } else {
            filePayload = await MultipartFile.fromFile(_podImage!.path);
          }
          final formData = FormData.fromMap({'file': filePayload});
          final uploadRes = await apiClient.client.post(
            'sales/upload',
            data: formData,
          );
          if (uploadRes.data != null && uploadRes.data['url'] != null) {
            uploadedPodUrl = uploadRes.data['url'];
          }
        } catch (e) {
          bool isConnectionError = true;
          if (e is DioException && e.type == DioExceptionType.badResponse) {
            isConnectionError = false;
          }
          if (isConnectionError) {
            rethrow;
          }
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('Gagal mengunggah foto POD: $e'),
                backgroundColor: Colors.orange,
              ),
            );
          }
          setState(() => _isLoading = false);
          return;
        }
      }

      // 1b. Upload Transfer Proof Image
      String uploadedTransferProofUrl = '';
      if ((_paymentMethod == 'VA' || isMt) && _transferProofImage != null) {
        try {
          final MultipartFile filePayload;
          if (kIsWeb) {
            final bytes = await _transferProofImage!.readAsBytes();
            filePayload = MultipartFile.fromBytes(
              bytes,
              filename: _transferProofImage!.name,
            );
          } else {
            filePayload = await MultipartFile.fromFile(
              _transferProofImage!.path,
            );
          }
          final formData = FormData.fromMap({'file': filePayload});
          final uploadRes = await apiClient.client.post(
            'sales/upload',
            data: formData,
          );
          if (uploadRes.data != null && uploadRes.data['url'] != null) {
            uploadedTransferProofUrl = uploadRes.data['url'];
          }
        } catch (e) {
          bool isConnectionError = true;
          if (e is DioException && e.type == DioExceptionType.badResponse) {
            isConnectionError = false;
          }
          if (isConnectionError) {
            rethrow;
          }
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('Gagal mengunggah bukti transfer: $e'),
                backgroundColor: Colors.orange,
              ),
            );
          }
          setState(() => _isLoading = false);
          return;
        }
      }

      // 2. Upload Courier & Customer Signatures
      Uint8List? courierSigBytes;
      Uint8List? customerSigBytes;
      String uploadedCourierSignatureUrl = '';
      String uploadedCustomerSignatureUrl = '';

      if (_courierSignaturePoints.isNotEmpty) {
        try {
          courierSigBytes = await _renderSignatureToPng(
            _courierSignaturePoints,
          );
          if (courierSigBytes != null) {
            final MultipartFile filePayload = MultipartFile.fromBytes(
              courierSigBytes,
              filename:
                  'courier_sig_${DateTime.now().millisecondsSinceEpoch}.png',
            );
            final formData = FormData.fromMap({'file': filePayload});
            final uploadRes = await apiClient.client.post(
              'sales/upload',
              data: formData,
            );
            if (uploadRes.data != null && uploadRes.data['url'] != null) {
              uploadedCourierSignatureUrl = uploadRes.data['url'];
            }
          }
        } catch (e) {
          print('Gagal render/unggah ttd kurir: $e');
          bool isConnectionError = true;
          if (e is DioException && e.type == DioExceptionType.badResponse) {
            isConnectionError = false;
          }
          if (isConnectionError) {
            rethrow;
          }
        }
      }

      if (_customerSignaturePoints.isNotEmpty) {
        try {
          customerSigBytes = await _renderSignatureToPng(
            _customerSignaturePoints,
          );
          if (customerSigBytes != null) {
            final MultipartFile filePayload = MultipartFile.fromBytes(
              customerSigBytes,
              filename:
                  'customer_sig_${DateTime.now().millisecondsSinceEpoch}.png',
            );
            final formData = FormData.fromMap({'file': filePayload});
            final uploadRes = await apiClient.client.post(
              'sales/upload',
              data: formData,
            );
            if (uploadRes.data != null && uploadRes.data['url'] != null) {
              uploadedCustomerSignatureUrl = uploadRes.data['url'];
            }
          }
        } catch (e) {
          print('Gagal render/unggah ttd customer: $e');
          bool isConnectionError = true;
          if (e is DioException && e.type == DioExceptionType.badResponse) {
            isConnectionError = false;
          }
          if (isConnectionError) {
            rethrow;
          }
        }
      }

      String combinedSignatureUrl = '';
      if (uploadedCourierSignatureUrl.isNotEmpty &&
          uploadedCustomerSignatureUrl.isNotEmpty) {
        combinedSignatureUrl =
            '$uploadedCourierSignatureUrl,$uploadedCustomerSignatureUrl';
      } else if (uploadedCourierSignatureUrl.isNotEmpty) {
        combinedSignatureUrl = uploadedCourierSignatureUrl;
      } else if (uploadedCustomerSignatureUrl.isNotEmpty) {
        combinedSignatureUrl = uploadedCustomerSignatureUrl;
      }

      List<Map<String, dynamic>> returns = [];
      for (final r in _returnResults) {
        returns.add(r.toJson());
      }

      List<Map<String, dynamic>> reqItems = [];
      _orderRequests.forEach((pId, qty) {
        if (qty > 0) {
          final p = _availableProducts.firstWhere(
            (prod) => prod['id']?.toString() == pId,
            orElse: () => {},
          );
          reqItems.add({
            'product_id': pId,
            'ordered_quantity': qty,
            'pieces_per_unit': p['pieces_per_unit'] ?? 1,
            'unit': p['unit'] ?? 'PCS',
            'price': (p['price'] as num? ?? 0.0).toDouble(),
          });
        }
      });

      // 3. Confirm Delivery
      await apiClient.client.post(
        'sales-orders/confirm-delivery',
        data: {
          'so_id': soId,
          'received_by': _receivedByController.text.isNotEmpty
              ? _receivedByController.text
              : 'Toko (via App)',
          'pod_image_url': uploadedPodUrl,
          'signature_image_url': combinedSignatureUrl,
          'notes': _notesController.text,
          'returned_items': returns,
          'request_items': reqItems,
        },
      );

      // 4. Collect Payment if any amount is entered OR it is a TEMPO payment method AND not already collected via Wizard transition
      if ((payAmount > 0 || _paymentMethod == 'TEMPO') &&
          _midtransData == null) {
        await apiClient.client.post(
          'sales-orders/collect-payment',
          data: {
            'so_id': soId,
            'amount': _paymentMethod == 'TEMPO' ? 0.0 : payAmount,
            'payment_method': _paymentMethod,
            'payment_bank': _selectedBank,
            'collected_by': 'Driver Delivery',
            if ((_paymentMethod == 'VA' || isMt) &&
                uploadedTransferProofUrl.isNotEmpty)
              'transfer_proof_url': uploadedTransferProofUrl,
            if (_paymentMethod == 'TEMPO' && _paymentDueDate != null)
              'payment_due_date':
                  '${_paymentDueDate!.toIso8601String().substring(0, 10)}T00:00:00.000Z',
            if (_activeBankAccount != null && _activeBankAccount!['id'] != null)
              'company_bank_account_id': _activeBankAccount!['id'],
          },
        );
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Selesai! Pengiriman & Pembayaran tersimpan.'),
            backgroundColor: Colors.green,
          ),
        );
        widget.onSuccess();

        // Push and replace / pop to display the beautiful DynamicInvoicePage
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(
            builder: (context) => DynamicInvoicePage(
              orderData: orderData,
              returnResults: _returnResults,
              paymentMethod: _paymentMethod,
              payAmount: payAmount,
              paymentDueDate: _paymentDueDate,
              receivedBy: _receivedByController.text.isNotEmpty
                  ? _receivedByController.text
                  : 'Toko (via App)',
            ),
          ),
        );
      }
    } catch (e) {
      print('DEBUG OFFLINE: Hitting backend error during submission: $e');

      bool isConnectionError = true;
      String errorMsg = e.toString();

      if (e is DioException) {
        if (e.type == DioExceptionType.badResponse) {
          isConnectionError = false;
          final responseData = e.response?.data;
          if (responseData is Map) {
            errorMsg =
                responseData['message'] ??
                responseData['error'] ??
                e.message ??
                e.toString();
          } else {
            errorMsg = e.message ?? e.toString();
          }
        }
      } else {
        isConnectionError = false;
      }

      if (!isConnectionError) {
        setState(() => _isLoading = false);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Gagal menyelesaikan pengiriman: $errorMsg'),
              backgroundColor: Colors.red,
            ),
          );
        }
        return;
      }

      // 6. Present Offline Fallback dialog to Driver
      final bool saveOffline = await _showOfflineCheckoutDialog(e);
      if (saveOffline) {
        setState(() => _isLoading = true);
        try {
          final payAmount = double.tryParse(_amountController.text) ?? 0.0;
          final soId =
              orderData['id'] ??
              widget.deliveryItem['sales_order_id'] ??
              widget.deliveryItem['so_id'];
          final storeName = orderData['store']?['name'] ?? 'Toko';

          // A. Render local Courier signature
          String courierPath = '';
          Uint8List? courierSigBytes;
          if (_courierSignaturePoints.isNotEmpty) {
            courierSigBytes = await _renderSignatureToPng(
              _courierSignaturePoints,
            );
            if (courierSigBytes != null) {
              if (kIsWeb) {
                courierPath = 'base64:${base64Encode(courierSigBytes)}';
              } else {
                final tempDir = Directory.systemTemp;
                final courierFile = File(
                  '${tempDir.path}/courier_sig_offline_${DateTime.now().millisecondsSinceEpoch}.png',
                );
                await courierFile.writeAsBytes(courierSigBytes);
                courierPath = courierFile.path;
              }
            }
          }

          // B. Render local Customer signature
          String customerPath = '';
          Uint8List? customerSigBytes;
          if (_customerSignaturePoints.isNotEmpty) {
            customerSigBytes = await _renderSignatureToPng(
              _customerSignaturePoints,
            );
            if (customerSigBytes != null) {
              if (kIsWeb) {
                customerPath = 'base64:${base64Encode(customerSigBytes)}';
              } else {
                final tempDir = Directory.systemTemp;
                final customerFile = File(
                  '${tempDir.path}/customer_sig_offline_${DateTime.now().millisecondsSinceEpoch}.png',
                );
                await customerFile.writeAsBytes(customerSigBytes);
                customerPath = customerFile.path;
              }
            }
          }

          List<Map<String, dynamic>> returns = [];
          for (final r in _returnResults) {
            returns.add(r.toJson());
          }

          // C. Queue into persistent offline JSON storage
          final offlineItem = OfflineDeliveryItem(
            localId: 'offline_${DateTime.now().millisecondsSinceEpoch}',
            soId: soId.toString(),
            storeName: storeName,
            receivedBy: _receivedByController.text,
            podImagePath: _podImage?.path ?? '',
            courierSigPath: courierPath,
            customerSigPath: customerPath,
            notes: _notesController.text,
            returnedItems: returns,
            amount: payAmount,
            paymentMethod: _paymentMethod,
            paymentBank: _selectedBank,
            createdAt: DateTime.now(),
          );

          await OfflineDeliveryManager().queueDelivery(offlineItem);

          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text(
                  'Offline Fallback Sukses! Data disimpan secara lokal.',
                ),
                backgroundColor: Colors.blueAccent,
                duration: Duration(seconds: 4),
              ),
            );
            widget.onSuccess();

            Navigator.pushReplacement(
              context,
              MaterialPageRoute(
                builder: (context) => DynamicInvoicePage(
                  orderData: orderData,
                  returnResults: _returnResults,
                  paymentMethod: _paymentMethod,
                  payAmount: payAmount,
                  paymentDueDate: _paymentDueDate,
                  receivedBy: _receivedByController.text.isNotEmpty
                      ? _receivedByController.text
                      : 'Toko (via App)',
                ),
              ),
            );
          }
        } catch (innerErr) {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('Gagal menyimpan offline: $innerErr'),
                backgroundColor: Colors.red,
              ),
            );
          }
        } finally {
          if (mounted) setState(() => _isLoading = false);
        }
      } else {
        if (mounted) setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final currencyFormat = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.black87),
        leading: IconButton(
          icon: const Icon(Icons.close_rounded),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Penyelesaian Pengiriman',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w800,
                color: const Color(0xFF1E293B),
                fontSize: 16,
              ),
            ),
            Text(
              'LANGKAH ${_currentStep + 1} DARI 4',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w900,
                color: Colors.blueAccent,
                fontSize: 10,
                letterSpacing: 1,
              ),
            ),
          ],
        ),
      ),
      body: Stack(
        children: [
          Column(
            children: [
              LinearProgressIndicator(
                value: (_currentStep + 1) / 4,
                backgroundColor: Colors.blueAccent.withOpacity(0.1),
                color: Colors.blueAccent,
                minHeight: 4,
              ),
              Expanded(
                child: PageView(
                  controller: _pageController,
                  physics: const NeverScrollableScrollPhysics(),
                  onPageChanged: (idx) => setState(() => _currentStep = idx),
                  children: [
                    _buildStepReview(currencyFormat),
                    _buildStepDocumentation(),
                    _buildStepPayment(currencyFormat),
                    _buildStepConfirmation(currencyFormat),
                  ],
                ),
              ),
              _buildBottomNav(),
            ],
          ),
          if (_isLoading)
            Container(
              color: Colors.white.withOpacity(0.7),
              child: const Center(
                child: CircularProgressIndicator(color: Colors.blueAccent),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildStepReview(NumberFormat currencyFormat) {
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        _buildStepHeader(
          'Verifikasi Barang',
          'Periksa barang yang dikirim dan catat retur jika ada.',
        ),
        const SizedBox(height: 24),
        if (_items.isEmpty)
          _buildEmptyItemsWarning()
        else
          Container(
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: Colors.grey.shade200),
            ),
            child: Column(
              children: _items.map((item) {
                String pId = item['product_id'] ?? item['id'] ?? '';
                int ordered = item['ordered_quantity'] ?? 0;
                int ret = _returnedQuantities[pId] ?? 0;
                return IgnorePointer(
                  ignoring: _isReadOnly,
                  child: _buildItemRow(item, ordered, ret, pId),
                );
              }).toList(),
            ),
          ),

        // Tampilkan Bonus Items (Hadiah) jika ada
        if (orderData['bonus_items'] != null &&
            (orderData['bonus_items'] as List).isNotEmpty) ...[
          const SizedBox(height: 24),
          Text(
            'Barang Bonus / Hadiah',
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w900,
              fontSize: 16,
              color: const Color(0xFF1E293B),
            ),
          ),
          const SizedBox(height: 8),
          Container(
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: Colors.grey.shade200),
            ),
            child: Column(
              children: (orderData['bonus_items'] as List).map((bonus) {
                final prodName = bonus['bonus_inventory']?['name'] ?? 'Bonus';
                final qty = bonus['quantity'] ?? 0;
                final unit = bonus['bonus_inventory']?['unit'] ?? 'PCS';

                return Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    border: Border(
                      bottom: BorderSide(color: Colors.grey.shade100),
                    ),
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.amber.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: const Icon(
                          Icons.card_giftcard_rounded,
                          color: Colors.amber,
                          size: 24,
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              prodName,
                              style: GoogleFonts.outfit(
                                fontWeight: FontWeight.bold,
                                fontSize: 14,
                                color: const Color(0xFF1E293B),
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              'Hadiah Promosi',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                color: Colors.blueGrey,
                              ),
                            ),
                          ],
                        ),
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.green.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          '$qty $unit',
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.bold,
                            fontSize: 13,
                            color: Colors.green,
                          ),
                        ),
                      ),
                    ],
                  ),
                );
              }).toList(),
            ),
          ),
        ],

        const SizedBox(height: 24),
        // Request Pesanan Baru Section
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(24),
            border: Border.all(color: Colors.grey.shade200),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Request Pesanan Baru',
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                            color: const Color(0xFF1E293B),
                          ),
                        ),
                        Text(
                          'Driver dapat mengajukan request pesanan tambahan dari customer.',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            color: Colors.grey,
                          ),
                        ),
                      ],
                    ),
                  ),
                  if (!_isReadOnly)
                    ElevatedButton.icon(
                      onPressed: () {
                        // Open product search sheet for new order request
                        showModalBottomSheet(
                          context: context,
                          isScrollControlled: true,
                          backgroundColor: Colors.white,
                          shape: const RoundedRectangleBorder(
                            borderRadius: BorderRadius.vertical(
                              top: Radius.circular(24),
                            ),
                          ),
                          builder: (context) {
                            return _ProductSearchSheet(
                              products: _availableProducts,
                              title: 'Tambah Request Pesanan',
                              isArmada: false,
                              onSelected: (product) {
                                setState(() {
                                  final pId = product['id']?.toString() ?? '';
                                  _orderRequests[pId] =
                                      (_orderRequests[pId] ?? 0) + 1;
                                });
                              },
                            );
                          },
                        );
                      },
                      icon: const Icon(
                        Icons.add_shopping_cart_rounded,
                        size: 14,
                        color: Colors.white,
                      ),
                      label: Text(
                        'Tambah',
                        style: GoogleFonts.outfit(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                      ),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.blueAccent,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 8,
                        ),
                        minimumSize: const Size(0, 36),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(10),
                        ),
                      ),
                    ),
                ],
              ),
              if (_orderRequests.entries.where((e) => e.value > 0).isEmpty) ...[
                const SizedBox(height: 12),
                Center(
                  child: Text(
                    'Belum ada request pesanan baru.',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      color: Colors.grey.shade400,
                      fontStyle: FontStyle.italic,
                    ),
                  ),
                ),
              ] else ...[
                const SizedBox(height: 12),
                const Divider(height: 1),
                ..._orderRequests.entries.where((e) => e.value > 0).map((e) {
                  final pId = e.key;
                  final qty = e.value;
                  final p = _availableProducts.firstWhere(
                    (prod) => prod['id']?.toString() == pId,
                    orElse: () => {},
                  );
                  final pName = p['name']?.toString() ?? 'Produk';
                  return Padding(
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(
                            pName,
                            style: GoogleFonts.outfit(
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                        if (!_isReadOnly)
                          Container(
                            decoration: BoxDecoration(
                              color: Colors.grey.shade100,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Row(
                              children: [
                                IconButton(
                                  icon: const Icon(Icons.remove, size: 14),
                                  onPressed: () {
                                    setState(() {
                                      if (qty > 1) {
                                        _orderRequests[pId] = qty - 1;
                                      } else {
                                        _orderRequests.remove(pId);
                                      }
                                    });
                                  },
                                ),
                                Text(
                                  '$qty',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.bold,
                                    fontSize: 12,
                                  ),
                                ),
                                IconButton(
                                  icon: const Icon(Icons.add, size: 14),
                                  onPressed: () {
                                    setState(() {
                                      _orderRequests[pId] = qty + 1;
                                    });
                                  },
                                ),
                              ],
                            ),
                          )
                        else
                          Text(
                            '$qty ${p['unit'] ?? 'PCS'}',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.bold,
                              fontSize: 13,
                              color: Colors.blueGrey,
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
        _buildTagihanSummary(currencyFormat),
        if (!_isReadOnly) ...[
          const SizedBox(height: 24),
          OutlinedButton.icon(
            onPressed: _handleStoreClosed,
            icon: const Icon(Icons.store_rounded, color: Colors.red),
            label: Text(
              'TOKO TUTUP / GAGAL KIRIM',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.bold,
                color: Colors.red,
                fontSize: 13,
                letterSpacing: 0.5,
              ),
            ),
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 16),
              side: const BorderSide(color: Colors.red, width: 2),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildStepDocumentation() {
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        _buildStepHeader(
          'Dokumentasi',
          'Ambil foto sebagai bukti bahwa barang telah diterima oleh toko.',
        ),
        const SizedBox(height: 40),
        GestureDetector(
          onTap: _takePhoto,
          child: Container(
            height: 300,
            width: double.infinity,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(32),
              border: Border.all(
                color: _podImage != null
                    ? Colors.blueAccent
                    : Colors.grey.shade300,
                width: 2,
                style: BorderStyle.solid,
              ),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.05),
                  blurRadius: 20,
                  offset: const Offset(0, 10),
                ),
              ],
            ),
            child: _podImage != null
                ? ClipRRect(
                    borderRadius: BorderRadius.circular(30),
                    child: kIsWeb
                        ? CachedNetworkImage(
                            imageUrl: _podImage!.path,
                            fit: BoxFit.cover,
                            placeholder: (context, url) => const Center(
                              child: CircularProgressIndicator(),
                            ),
                            errorWidget: (context, url, error) =>
                                const Icon(Icons.error),
                            memCacheWidth: 800,
                          )
                        : Image.file(File(_podImage!.path), fit: BoxFit.cover),
                  )
                : Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Container(
                        padding: const EdgeInsets.all(24),
                        decoration: BoxDecoration(
                          color: Colors.blueAccent.withOpacity(0.1),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(
                          Icons.camera_alt_rounded,
                          size: 48,
                          color: Colors.blueAccent,
                        ),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        'Ambil Foto Bukti (POD)',
                        style: GoogleFonts.outfit(
                          color: const Color(0xFF1E293B),
                          fontWeight: FontWeight.w800,
                          fontSize: 16,
                        ),
                      ),
                      Text(
                        'Pastikan foto terlihat jelas',
                        style: GoogleFonts.outfit(
                          color: Colors.blueGrey,
                          fontSize: 13,
                        ),
                      ),
                    ],
                  ),
          ),
        ),
        const SizedBox(height: 32),
        Text(
          'Informasi Penerima',
          style: GoogleFonts.outfit(fontWeight: FontWeight.bold, fontSize: 16),
        ),
        const SizedBox(height: 12),
        TextField(
          controller: _receivedByController,
          decoration: InputDecoration(
            hintText: 'Masukkan nama penerima barang...',
            prefixIcon: const Icon(Icons.person_outline),
            filled: true,
            fillColor: Colors.white,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(16),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(16),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildStepPayment(NumberFormat currencyFormat) {
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        _buildStepHeader(
          'Pembayaran',
          'Pilih metode pembayaran dan masukkan nominal yang diterima.',
        ),
        const SizedBox(height: 24),
        _buildPaymentMethodSelector(),
        const SizedBox(height: 24),
        _buildPaymentAmountInput(currencyFormat),
      ],
    );
  }

  Widget _buildFallbackPhotoText() {
    return Row(
      children: [
        const Icon(Icons.photo_library_rounded, color: Colors.green, size: 20),
        const SizedBox(width: 12),
        Text(
          'Foto Bukti Tersedia',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.bold,
            color: Colors.green,
            fontSize: 13,
          ),
        ),
      ],
    );
  }

  Widget _buildStepConfirmation(NumberFormat currencyFormat) {
    final payAmount = double.tryParse(_amountController.text) ?? 0.0;
    final hasMidtrans = _midtransData != null;

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        _buildStepHeader(
          hasMidtrans
              ? 'Detail Pembayaran & Tanda Tangan'
              : 'Konfirmasi Akhir & Tanda Tangan',
          hasMidtrans
              ? 'Informasi pembayaran digital untuk pelanggan.'
              : 'Silakan berikan tanda tangan digital untuk menyelesaikan dokumen.',
        ),
        const SizedBox(height: 24),
        if (hasMidtrans && !_isPaymentConfirmed) ...[
          _buildMidtransInfoPreview(currencyFormat),
          const SizedBox(height: 16),
          // Beautiful verification controls
          Row(
            children: [
              Expanded(
                child: ElevatedButton.icon(
                  onPressed: _checkPaymentStatus,
                  icon: const Icon(Icons.refresh_rounded, size: 16),
                  label: Text(
                    'CEK PEMBAYARAN',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.bold,
                      fontSize: 12,
                    ),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blueAccent,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: ElevatedButton.icon(
                  onPressed: _simulateQrisSuccess,
                  icon: const Icon(Icons.science_rounded, size: 16),
                  label: Text(
                    'SIMULASI SUKSES (DEV)',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.bold,
                      fontSize: 11,
                    ),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.orange.shade800,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 24),
        ],
        if (_isPaymentConfirmed) ...[
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.green.withOpacity(0.1),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: Colors.green.withOpacity(0.3)),
            ),
            child: Row(
              children: [
                const Icon(
                  Icons.check_circle_rounded,
                  color: Colors.green,
                  size: 28,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Pembayaran Terkonfirmasi',
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.bold,
                          color: Colors.green.shade900,
                          fontSize: 14,
                        ),
                      ),
                      Text(
                        'Silakan gambar tanda tangan Kurir & Pelanggan di bawah.',
                        style: GoogleFonts.outfit(
                          color: Colors.green.shade800,
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),

          // Double signature pads
          if (!_isReadOnly) ...[
            SignatureCanvas(
              label: 'TANDA TANGAN KURIR / DRIVER PENGIRIMAN',
              points: _courierSignaturePoints,
              onDraw: (pts) => setState(() => _courierSignaturePoints = pts),
              onClear: () => setState(() => _courierSignaturePoints.clear()),
            ),
            const SizedBox(height: 20),
            SignatureCanvas(
              label: 'TANDA TANGAN CUSTOMER / PENERIMA BARANG',
              points: _customerSignaturePoints,
              onDraw: (pts) => setState(() => _customerSignaturePoints = pts),
              onClear: () => setState(() => _customerSignaturePoints.clear()),
            ),
            const SizedBox(height: 24),
          ],
        ],
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
            children: [
              _summaryRow('Toko', orderData['store']?['name'] ?? 'Toko'),
              const Divider(height: 32),
              if (_orderRequests.entries
                  .where((e) => e.value > 0)
                  .isNotEmpty) ...[
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Request Pesanan:',
                      style: GoogleFonts.outfit(
                        color: Colors.blueGrey,
                        fontWeight: FontWeight.w600,
                        fontSize: 14,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                ..._orderRequests.entries.where((e) => e.value > 0).map((e) {
                  final pId = e.key;
                  final qty = e.value;
                  final p = _availableProducts.firstWhere(
                    (prod) => prod['id']?.toString() == pId,
                    orElse: () => {},
                  );
                  final pName = p['name']?.toString() ?? 'Produk';
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 6, left: 8),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Expanded(
                          child: Text(
                            '- $pName',
                            style: GoogleFonts.outfit(
                              fontSize: 12.5,
                              color: const Color(0xFF1E293B),
                            ),
                          ),
                        ),
                        Text(
                          '$qty ${p['unit'] ?? 'PCS'}',
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.bold,
                            fontSize: 12.5,
                            color: Colors.blueAccent,
                          ),
                        ),
                      ],
                    ),
                  );
                }),
                const Divider(height: 32),
              ],
              _summaryRow(
                'Tagihan Akhir',
                currencyFormat.format(_calculatedTotal),
                isBold: true,
              ),
              _summaryRow(
                'Metode Bayar',
                _paymentOptions[_paymentMethod]?['label'] ?? _paymentMethod,
              ),
              _summaryRow(
                'Nominal Bayar',
                currencyFormat.format(payAmount),
                valueColor: Colors.blueAccent,
              ),
              if (payAmount < _calculatedTotal)
                _summaryRow(
                  'Sisa Piutang',
                  currencyFormat.format(_calculatedTotal - payAmount),
                  valueColor: Colors.orange.shade800,
                ),
              const Divider(height: 32),
              if (_existingPodImageUrl != null &&
                  _existingPodImageUrl!.isNotEmpty)
                ClipRRect(
                  borderRadius: BorderRadius.circular(16),
                  child: CachedNetworkImage(
                    imageUrl:
                        '${apiClient.client.options.baseUrl.replaceAll('/api/v1', '')}$_existingPodImageUrl',
                    height: 150,
                    width: double.infinity,
                    fit: BoxFit.cover,
                    placeholder: (context, url) =>
                        const Center(child: CircularProgressIndicator()),
                    errorWidget: (context, url, error) =>
                        const Icon(Icons.error),
                    memCacheWidth: 800,
                  ),
                )
              else if (_podImage != null)
                ClipRRect(
                  borderRadius: BorderRadius.circular(16),
                  child: kIsWeb
                      ? CachedNetworkImage(
                          imageUrl: _podImage!.path,
                          height: 150,
                          width: double.infinity,
                          fit: BoxFit.cover,
                          placeholder: (context, url) =>
                              const Center(child: CircularProgressIndicator()),
                          errorWidget: (context, url, error) =>
                              const Icon(Icons.error),
                          memCacheWidth: 800,
                        )
                      : Image.file(
                          File(_podImage!.path),
                          height: 150,
                          width: double.infinity,
                          fit: BoxFit.cover,
                        ),
                )
              else
                _buildFallbackPhotoText(),
            ],
          ),
        ),
        const SizedBox(height: 24),
        if (hasMidtrans) ...[
          ElevatedButton.icon(
            onPressed: _shareMidtransToWhatsApp,
            icon: const Icon(Icons.share_rounded, size: 18),
            label: Text(
              'BAGIKAN KE PELANGGAN',
              style: GoogleFonts.outfit(fontWeight: FontWeight.bold),
            ),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF25D366),
              foregroundColor: Colors.white,
              minimumSize: const Size(double.infinity, 54),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
            ),
          ),
          const SizedBox(height: 24),
        ],
        TextField(
          controller: _notesController,
          maxLines: 3,
          decoration: InputDecoration(
            hintText: 'Tambahkan catatan jika ada...',
            hintStyle: GoogleFonts.outfit(fontSize: 14),
            filled: true,
            fillColor: Colors.white,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(16),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(16),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildMidtransInfoPreview(NumberFormat currencyFormat) {
    final qrisUrl = _midtransData?['midtrans_qris_url'];
    final vaNumber =
        _midtransData?['midtrans_va_number'] ??
        _midtransData?['midtrans_bill_key'];
    final bank = (_midtransData?['midtrans_bank'] ?? 'Bank')
        .toString()
        .toUpperCase();

    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: Colors.blueAccent.withOpacity(0.05),
        borderRadius: BorderRadius.circular(28),
        border: Border.all(color: Colors.blueAccent.withOpacity(0.2)),
      ),
      child: Column(
        children: [
          if (_paymentMethod == 'QRIS' && qrisUrl != null) ...[
            if (qrisUrl.startsWith('http')) ...[
              QrisSandboxWarning(
                qrisUrl: qrisUrl,
                margin: const EdgeInsets.only(bottom: 16),
              ),
            ],
            qrisUrl.startsWith('http')
                ? Container(
                    color: Colors.white,
                    padding: const EdgeInsets.all(12),
                    child: CachedNetworkImage(
                      imageUrl:
                          '${AppConstants.baseUrl}sales/qris-image?url=${Uri.encodeComponent(qrisUrl)}',
                      height: 200,
                      width: 200,
                      fit: BoxFit.contain,
                      placeholder: (context, url) =>
                          const Center(child: CircularProgressIndicator()),
                      errorWidget: (context, url, error) =>
                          const Icon(Icons.error),
                      memCacheWidth: 800,
                    ),
                  )
                : QrImageView(
                    data: qrisUrl,
                    size: 200,
                    backgroundColor: Colors.white,
                    padding: const EdgeInsets.all(12),
                  ),
            const SizedBox(height: 16),
            Text(
              'SCAN QRIS UNTUK BAYAR',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w900,
                color: Colors.blueAccent,
                fontSize: 14,
              ),
            ),
            const SizedBox(height: 12),
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
                      qrisUrl.toString(),
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
                      Clipboard.setData(
                        ClipboardData(text: qrisUrl.toString()),
                      );
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
          ] else if (_paymentMethod == 'VA' && vaNumber != null) ...[
            Text(
              'NOMOR VIRTUAL ACCOUNT ($bank)',
              style: GoogleFonts.outfit(
                fontSize: 10,
                fontWeight: FontWeight.w900,
                color: Colors.blueGrey,
                letterSpacing: 1,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              vaNumber,
              style: GoogleFonts.outfit(
                fontSize: 28,
                fontWeight: FontWeight.w900,
                color: const Color(0xFF1E293B),
              ),
            ),
            if (_midtransData?['midtrans_biller_code'] != null)
              Text(
                'Biller: ${_midtransData!['midtrans_biller_code']}',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.bold,
                  color: Colors.blueGrey,
                ),
              ),
          ],
        ],
      ),
    );
  }

  Future<void> _shareMidtransToWhatsApp() async {
    final currencyFormat = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    final storeName = orderData['store']?['name'] ?? 'Toko';
    final payAmount = double.tryParse(_amountController.text) ?? 0.0;
    final vaNumber =
        _midtransData?['midtrans_va_number'] ??
        _midtransData?['midtrans_bill_key'];
    final bank = (_midtransData?['midtrans_bank'] ?? 'Bank')
        .toString()
        .toUpperCase();

    String message =
        "Halo *$storeName*,\n\nBerikut informasi pembayaran untuk pengiriman hari ini:\n";
    message += "💰 *Total Tagihan:* ${currencyFormat.format(payAmount)}\n";
    message += "💳 *Metode Pembayaran:* $_paymentMethod\n";

    if (_paymentMethod == 'VA' && vaNumber != null) {
      message += "\n🏦 *Informasi Pembayaran ($bank):*\n";
      message += "Nomor: `$vaNumber`";
      if (_midtransData?['midtrans_biller_code'] != null)
        message += "\nKode Biller: `${_midtransData!['midtrans_biller_code']}`";
    }

    message += "\n\nTerima kasih!";

    final whatsappUrl = Uri.parse(
      "https://wa.me/?text=${Uri.encodeComponent(message)}",
    );
    if (await canLaunchUrl(whatsappUrl)) {
      await launchUrl(whatsappUrl, mode: LaunchMode.externalApplication);
    }
  }

  Widget _buildStepHeader(String title, String subtitle) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.w900,
            fontSize: 24,
            color: const Color(0xFF1E293B),
          ),
        ),
        const SizedBox(height: 4),
        Text(
          subtitle,
          style: GoogleFonts.outfit(
            fontSize: 14,
            color: Colors.blueGrey,
            fontWeight: FontWeight.w500,
          ),
        ),
      ],
    );
  }

  Widget _buildBottomNav() {
    bool isFirst = _currentStep == 0;
    bool isLast = _currentStep == 3;

    return Container(
      padding: const EdgeInsets.fromLTRB(24, 16, 24, 32),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, -5),
          ),
        ],
      ),
      child: Row(
        children: [
          if (!isFirst)
            Expanded(
              flex: 1,
              child: OutlinedButton(
                onPressed: _prevPage,
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 18),
                  side: const BorderSide(color: Color(0xFFE2E8F0)),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
                child: Text(
                  'KEMBALI',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w800,
                    color: Colors.blueGrey,
                  ),
                ),
              ),
            ),
          if (!isFirst && !_isReadOnly) const SizedBox(width: 16),
          if (!_isReadOnly)
            Expanded(
              flex: 2,
              child: ElevatedButton(
                onPressed: isLast
                    ? (_isPaymentConfirmed ? _submitCheckout : null)
                    : _validateAndNext,
                style: ElevatedButton.styleFrom(
                  backgroundColor: isLast
                      ? (_isPaymentConfirmed ? Colors.green : Colors.grey)
                      : Colors.blueAccent,
                  padding: const EdgeInsets.symmetric(vertical: 18),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  elevation: 0,
                ),
                child: FittedBox(
                  fit: BoxFit.scaleDown,
                  child: Text(
                    isLast
                        ? (_isPaymentConfirmed
                              ? 'SELESAIKAN & SIMPAN'
                              : 'MENUNGGU PEMBAYARAN')
                        : 'LANJUT',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900,
                      color: Colors.white,
                      letterSpacing: 1,
                    ),
                  ),
                ),
              ),
            ),
          if (_isReadOnly)
            Expanded(
              child: ElevatedButton(
                onPressed: () => Navigator.pop(context),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF1E293B),
                  padding: const EdgeInsets.symmetric(vertical: 18),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
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

  Future<void> _validateAndNext() async {
    if (_currentStep == 1 && _podImage == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Harap ambil foto bukti pengiriman sebelum lanjut.'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }
    if (_currentStep == 2) {
      if (_paymentMethod == 'VA' || isMt) {
        if (_transferProofImage == null) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text(
                'Harap upload bukti transaksi/pembayaran terlebih dahulu.',
              ),
              backgroundColor: Colors.orange,
            ),
          );
          return;
        }
        setState(() {
          _isPaymentConfirmed = true;
        });
      } else if (_paymentMethod == 'QRIS') {
        setState(() => _isLoading = true);
        try {
          final payAmount = double.tryParse(_amountController.text) ?? 0.0;
          final soId = orderData['id'] ?? widget.deliveryItem['sales_order_id'];

          final response = await apiClient.client.post(
            'sales-orders/collect-payment',
            data: {
              'so_id': soId,
              'amount': payAmount,
              'payment_method': _paymentMethod,
              'payment_bank': _selectedBank,
              'collected_by': 'Driver Delivery',
            },
          );

          final resData =
              response.data is Map && response.data.containsKey('data')
              ? response.data['data']
              : response.data;
          setState(() {
            _midtransData = {
              'payment_method': _paymentMethod,
              'midtrans_qris_url': resData['midtrans_qris_url'],
              'midtrans_va_number': resData['midtrans_va_number'],
              'midtrans_bank': resData['midtrans_bank'] ?? _selectedBank,
              'midtrans_bill_key': resData['midtrans_bill_key'],
              'midtrans_biller_code': resData['midtrans_biller_code'],
            };
            _isPaymentConfirmed = false;
          });
        } catch (e) {
          final bool switchSuccess = await _showOfflineFallbackDialog(
            e.toString(),
          );
          if (switchSuccess) {
            setState(() {
              _paymentMethod = 'CASH';
              _isPaymentConfirmed = true;
              _midtransData = null;
            });
          } else {
            return; // Stay on payment step
          }
        } finally {
          setState(() => _isLoading = false);
        }
      } else {
        setState(() {
          _isPaymentConfirmed = true;
        });
      }
    }

    // Use addPostFrameCallback to avoid PageController error when PageView is rebuilding
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _nextPage();
    });
  }

  Future<bool> _showOfflineFallbackDialog(String errorMsg) async {
    return await showDialog<bool>(
          context: context,
          barrierDismissible: false,
          builder: (BuildContext context) {
            return AlertDialog(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(24),
              ),
              title: Row(
                children: [
                  const Icon(
                    Icons.signal_wifi_off_rounded,
                    color: Colors.orange,
                    size: 28,
                  ),
                  const SizedBox(width: 12),
                  Text(
                    'Daerah Susah Sinyal',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.bold,
                      fontSize: 18,
                    ),
                  ),
                ],
              ),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Gagal memuat pembayaran digital karena koneksi terganggu/lambat.',
                    style: GoogleFonts.outfit(
                      fontSize: 14,
                      color: const Color(0xFF1E293B),
                    ),
                  ),
                  const SizedBox(height: 12),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.orange.withOpacity(0.08),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.orange.withOpacity(0.2)),
                    ),
                    child: Text(
                      'Solusi Offline Fallback:\nAnda bisa mengalihkan metode pembayaran ke Cash atau Transfer Manual agar pengiriman barang dan tanda tangan digital tetap dapat diselesaikan saat offline.',
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: Colors.orange.shade900,
                        height: 1.4,
                      ),
                    ),
                  ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: Text(
                    'Coba Lagi',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.bold,
                      color: Colors.blueGrey,
                    ),
                  ),
                ),
                ElevatedButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.orange.shade800,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: Text(
                    'Aktifkan Mode Cash / Offline',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ),
              ],
            );
          },
        ) ??
        false;
  }

  Future<bool> _showOfflineCheckoutDialog(Object error) async {
    return await showDialog<bool>(
          context: context,
          barrierDismissible: false,
          builder: (BuildContext context) {
            return AlertDialog(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(24),
              ),
              title: Row(
                children: [
                  const Icon(
                    Icons.signal_wifi_off_rounded,
                    color: Colors.blueAccent,
                    size: 28,
                  ),
                  const SizedBox(width: 12),
                  Text(
                    'Koneksi Jaringan Gagal',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.bold,
                      fontSize: 18,
                    ),
                  ),
                ],
              ),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Gagal mengirimkan data ke server karena koneksi terganggu/susah sinyal.',
                    style: GoogleFonts.outfit(
                      fontSize: 14,
                      color: const Color(0xFF1E293B),
                    ),
                  ),
                  const SizedBox(height: 12),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.blueAccent.withOpacity(0.08),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: Colors.blueAccent.withOpacity(0.2),
                      ),
                    ),
                    child: Text(
                      'Fitur Offline Fallback Aktif:\nSeluruh data pengiriman (termasuk foto bukti POD dan tanda tangan) akan disimpan di memori HP Anda. Sistem akan menyelaraskan data ini otomatis begitu Anda mendapat sinyal kembali.',
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: Colors.blueAccent.shade700,
                        height: 1.4,
                      ),
                    ),
                  ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false),
                  child: Text(
                    'Batal',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.bold,
                      color: Colors.blueGrey,
                    ),
                  ),
                ),
                ElevatedButton(
                  onPressed: () => Navigator.of(context).pop(true),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blueAccent,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: Text(
                    'Simpan Offline',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ),
              ],
            );
          },
        ) ??
        false;
  }

  Widget _buildEmptyItemsWarning() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.orange.shade50,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.orange.shade200),
      ),
      child: Row(
        children: [
          Icon(Icons.info_outline, color: Colors.orange.shade800),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              'Rincian barang tidak tersedia. Silakan lanjutkan dengan total tagihan.',
              style: GoogleFonts.outfit(
                color: Colors.orange.shade900,
                fontSize: 13,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildItemRow(dynamic item, int ordered, int ret, String pId) {
    final bool hasReturn = ret > 0;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      item['product']?['name'] ?? 'Produk',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.bold,
                        fontSize: 14,
                      ),
                    ),
                    Text(
                      'Dikirim: $ordered ${item['unit'] ?? ''}  •  Sampai: ${ordered - ret}',
                      style: GoogleFonts.outfit(
                        color: Colors.grey,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
              // Stepper retur
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: hasReturn ? Colors.red.shade50 : Colors.grey.shade100,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    GestureDetector(
                      onTap: () {
                        if (ret > 0) {
                          setState(() {
                            _returnedQuantities[pId] = ret - 1;
                            if (_returnedQuantities[pId] == 0) {
                              _returnReasons.remove(pId);
                              _returnTypes.remove(pId);
                              _replacementProductIds.remove(pId);
                              _replacementProductNames.remove(pId);
                              _replacementPrices.remove(pId);
                              _claimRefNos.remove(pId);
                            }
                          });
                          _updateAmountFromTotal();
                        }
                      },
                      child: Icon(
                        Icons.remove_circle,
                        color: Colors.red.shade400,
                        size: 24,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      '$ret Retur',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.bold,
                        color: hasReturn
                            ? Colors.red.shade700
                            : Colors.blueGrey,
                      ),
                    ),
                    const SizedBox(width: 8),
                    GestureDetector(
                      onTap: () {
                        if (ret < ordered) {
                          if (_returnTypes[pId] == 'TUKAR_PRODUK') {
                            final replId = _replacementProductIds[pId] ?? pId;
                            final currentAvailable = _getFleetStockForProduct(
                              replId,
                            );
                            if (currentAvailable <= 0) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text(
                                    'Stok muatan tidak mencukupi untuk ditambah.',
                                  ),
                                ),
                              );
                              return;
                            }
                          }
                          setState(() {
                            _returnedQuantities[pId] = ret + 1;
                            if (ret == 0 && !_returnTypes.containsKey(pId)) {
                              _returnTypes[pId] = 'POTONG_NOTA'; // default
                            }
                          });
                          _updateAmountFromTotal();
                        }
                      },
                      child: Icon(
                        Icons.add_circle,
                        color: Colors.red.shade400,
                        size: 24,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),

        // Inline Return Details — muncul hanya saat ada retur
        if (hasReturn && !_isReadOnly)
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 4, 16, 12),
            child: Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.red.shade50.withOpacity(0.3),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: Colors.red.shade100),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // 1. Alasan Retur Selector (Pill wrap)
                  Text(
                    'ALASAN RETUR',
                    style: GoogleFonts.outfit(
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                      color: Colors.red.shade800,
                      letterSpacing: 0.5,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 6,
                    runSpacing: 6,
                    children: _returnReasonLabels.entries.map((e) {
                      final isSelected =
                          (_returnReasons[pId] ?? 'RUSAK') == e.key;
                      return GestureDetector(
                        onTap: () =>
                            setState(() => _returnReasons[pId] = e.key),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            color: isSelected
                                ? Colors.red.shade100
                                : Colors.white,
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(
                              color: isSelected
                                  ? Colors.red.shade300
                                  : Colors.grey.shade200,
                            ),
                          ),
                          child: Text(
                            e.value,
                            style: GoogleFonts.outfit(
                              fontSize: 10.5,
                              fontWeight: isSelected
                                  ? FontWeight.bold
                                  : FontWeight.normal,
                              color: isSelected
                                  ? Colors.red.shade900
                                  : Colors.black87,
                            ),
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                  const SizedBox(height: 16),

                  // 2. Mekanisme Retur (Type Selector)
                  Text(
                    'MEKANISME / TIPE RETUR',
                    style: GoogleFonts.outfit(
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                      color: Colors.red.shade800,
                      letterSpacing: 0.5,
                    ),
                  ),
                  const SizedBox(height: 8),
                  ..._returnTypeConfigs.entries.map((e) {
                    final typeKey = e.key;
                    final cfg = e.value;
                    final isSelected =
                        (_returnTypes[pId] ?? 'POTONG_NOTA') == typeKey;
                    final color = cfg['color'] as Color;
                    final isTukarLangsung = typeKey == 'TUKAR_PRODUK';
                    final availableFleetStock = _getFleetStockForProduct(
                      item['product']?['id']?.toString() ?? pId,
                    );

                    int displayStock = availableFleetStock;
                    if (isTukarLangsung &&
                        typeKey != (_returnTypes[pId] ?? 'POTONG_NOTA')) {
                      displayStock -= ret;
                    }
                    if (displayStock < 0) displayStock = 0;

                    final isDisabled =
                        isTukarLangsung && availableFleetStock <= 0;

                    return GestureDetector(
                      onTap: isDisabled
                          ? null
                          : () {
                              setState(() {
                                _returnTypes[pId] = typeKey;
                                if (isTukarLangsung) {
                                  _replacementProductIds[pId] =
                                      item['product']?['id']?.toString() ?? pId;
                                  _replacementProductNames[pId] =
                                      item['product']?['name']?.toString() ??
                                      'Produk';
                                  _replacementPrices[pId] =
                                      (item['price'] ?? 0.0) is num
                                      ? (item['price'] ?? 0.0).toDouble()
                                      : 0.0;
                                  _replacementQuantities[pId] = ret;
                                }
                                _updateAmountFromTotal();
                              });
                            },
                      child: Opacity(
                        opacity: isDisabled ? 0.5 : 1.0,
                        child: Container(
                          margin: const EdgeInsets.only(bottom: 6),
                          padding: const EdgeInsets.all(10),
                          decoration: BoxDecoration(
                            color: isSelected
                                ? color.withOpacity(0.06)
                                : Colors.white,
                            borderRadius: BorderRadius.circular(10),
                            border: Border.all(
                              color: isSelected ? color : Colors.grey.shade200,
                              width: isSelected ? 1.5 : 1,
                            ),
                          ),
                          child: Row(
                            children: [
                              Icon(
                                cfg['icon'] as IconData,
                                color: isSelected ? color : Colors.grey,
                                size: 18,
                              ),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      isTukarLangsung
                                          ? '${cfg['label']} (Stok: $displayStock)'
                                          : cfg['label'] as String,
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.bold,
                                        fontSize: 11,
                                        color: isSelected
                                            ? color
                                            : Colors.black87,
                                      ),
                                    ),
                                    Text(
                                      isDisabled && isTukarLangsung
                                          ? 'Stok pengganti di armada kosong'
                                          : cfg['description'] as String,
                                      style: GoogleFonts.outfit(
                                        fontSize: 9.5,
                                        color: Colors.grey.shade600,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              if (isSelected)
                                Icon(
                                  Icons.check_circle,
                                  color: color,
                                  size: 16,
                                ),
                            ],
                          ),
                        ),
                      ),
                    );
                  }),
                ],
              ),
            ),
          ),

        // Read-only mode display
        if (hasReturn && _isReadOnly)
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 4, 16, 12),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: Colors.red.shade50.withOpacity(0.5),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: Colors.red.shade100),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Alasan: ${_returnReasonLabels[_returnReasons[pId]] ?? _returnReasons[pId] ?? '-'}',
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: Colors.red.shade800,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    'Mekanisme: ${_returnTypeConfigs[_returnTypes[pId]]?['label'] ?? _returnTypes[pId] ?? 'Potong Nota'}',
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: Colors.red.shade800,
                    ),
                  ),
                  if (_replacementProductNames[pId] != null) ...[
                    const SizedBox(height: 4),
                    Text(
                      'Pengganti: ${_replacementProductNames[pId]} (Qty: ${_replacementQuantities[pId] ?? ret} ${item['unit'] ?? 'PCS'})',
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        color: Colors.cyan.shade800,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        const Divider(height: 1, indent: 16, endIndent: 16),
      ],
    );
  }

  Widget _buildTagihanSummary(NumberFormat currencyFormat) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF1E293B), Color(0xFF334155)],
        ),
        borderRadius: BorderRadius.circular(24),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (_originalTotal != _calculatedTotal)
            Text(
              'Awal: ${currencyFormat.format(_originalTotal)}',
              style: GoogleFonts.outfit(
                color: Colors.grey.shade400,
                decoration: TextDecoration.lineThrough,
                fontSize: 12,
              ),
            ),
          Text(
            'TOTAL SETELAH RETUR',
            style: GoogleFonts.outfit(
              color: Colors.blue.shade200,
              fontWeight: FontWeight.bold,
              fontSize: 10,
              letterSpacing: 1,
            ),
          ),
          Text(
            currencyFormat.format(_calculatedTotal),
            style: GoogleFonts.outfit(
              color: Colors.white,
              fontWeight: FontWeight.w900,
              fontSize: 28,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPaymentMethodSelector() {
    return Column(
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
          final isSelected = _paymentMethod == e.key;
          return GestureDetector(
            onTap: _isReadOnly
                ? null
                : () => setState(() {
                    _paymentMethod = e.key;
                    if (e.key != 'VA') _selectedBank = null;
                    if (e.key == 'TEMPO' && _paymentDueDate == null) {
                      _selectedTempoDays = 14;
                      _paymentDueDate = DateTime.now().add(
                        const Duration(days: 14),
                      );
                    }
                  }),
            child: Container(
              margin: const EdgeInsets.only(bottom: 12),
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: isSelected
                    ? e.value['color'].withOpacity(0.05)
                    : const Color(0xFFF8FAFC),
                borderRadius: BorderRadius.circular(20),
                border: Border.all(
                  color: isSelected ? e.value['color'] : Colors.transparent,
                  width: 2,
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    e.value['icon'],
                    color: isSelected ? e.value['color'] : Colors.blueGrey,
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
        if (_paymentMethod == 'VA') ...[
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
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 4),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(14),
                border: Border.all(color: Colors.indigo.shade200, width: 1.5),
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
                  value: _selectedBankAccount?['id']?.toString(),
                  hint: Text(
                    'Pilih Rekening Bank',
                    style: GoogleFonts.outfit(fontSize: 13, color: Colors.grey),
                  ),
                  icon: const Icon(
                    Icons.keyboard_arrow_down_rounded,
                    color: Colors.indigo,
                  ),
                  items: _bankAccounts.map((b) {
                    final bName = b['bank_name']?.toString() ?? 'Bank';
                    final accNum = b['account_number']?.toString() ?? '';
                    final companyName =
                        b['company']?['name']?.toString() ??
                        b['account_holder']?.toString() ??
                        '';
                    final branchName = b['branch']?['name']?.toString();
                    final tag = branchName != null && branchName.isNotEmpty
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
                        _selectedBankAccount = _bankAccounts.firstWhere(
                          (b) => b['id']?.toString() == val,
                          orElse: () => _bankAccounts.first,
                        );
                        _activeBankAccount = _selectedBankAccount;
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
              border: Border.all(color: Colors.indigo.withOpacity(0.15)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    _selectedBankAccount?['bank_logo_url'] != null &&
                            (_selectedBankAccount!['bank_logo_url'] as String)
                                .isNotEmpty
                        ? CachedNetworkImage(
                            imageUrl:
                                _selectedBankAccount!['bank_logo_url']
                                    as String,
                            width: 50,
                            height: 25,
                            placeholder: (context, url) => const Center(
                              child: CircularProgressIndicator(strokeWidth: 2),
                            ),
                            errorWidget: (context, url, error) => const Icon(
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
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            _selectedBankAccount?['bank_name'] as String? ??
                                _activeBankAccount?['bank_name'] as String? ??
                                'Bank Central Asia (BCA)',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.bold,
                              fontSize: 14,
                              color: const Color(0xFF1E293B),
                            ),
                          ),
                          if (_selectedBankAccount?['company'] != null)
                            Text(
                              'PT: ${_selectedBankAccount!['company']['name']}',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                color: Colors.indigo.shade700,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          if (_selectedBankAccount?['branch'] != null)
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
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
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
                          _selectedBankAccount?['account_number'] as String? ??
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
                        Clipboard.setData(ClipboardData(text: accNo));
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(
                            content: Text('Nomor rekening disalin!'),
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
                  _selectedBankAccount?['account_holder'] as String? ??
                      _activeBankAccount?['account_holder'] as String? ??
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
        ],
        if (_paymentMethod == 'VA' || isMt) ...[
          const SizedBox(height: 24),
          Text(
            isMt ? 'BUKTI TRANSAKSI / DOKUMEN MT' : 'BUKTI TRANSFER BANK',
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
                  borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
                ),
                builder: (context) => SafeArea(
                  child: Wrap(
                    children: [
                      ListTile(
                        leading: const Icon(
                          Icons.camera_alt_rounded,
                          color: Colors.blueAccent,
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
                          color: Colors.blueAccent,
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
                            borderRadius: BorderRadius.circular(18),
                            child: kIsWeb
                                ? CachedNetworkImage(
                                    imageUrl: _transferProofImage!.path,
                                    fit: BoxFit.cover,
                                    placeholder: (context, url) => const Center(
                                      child: CircularProgressIndicator(),
                                    ),
                                    errorWidget: (context, url, error) =>
                                        const Icon(Icons.error),
                                    memCacheWidth: 800,
                                  )
                                : Image.file(
                                    File(_transferProofImage!.path),
                                    fit: BoxFit.cover,
                                  ),
                          ),
                        ),
                        Positioned(
                          top: 8,
                          right: 8,
                          child: CircleAvatar(
                            backgroundColor: Colors.black.withOpacity(0.6),
                            child: IconButton(
                              icon: const Icon(
                                Icons.close,
                                color: Colors.white,
                              ),
                              onPressed: () =>
                                  setState(() => _transferProofImage = null),
                            ),
                          ),
                        ),
                      ],
                    )
                  : Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(
                          Icons.cloud_upload_outlined,
                          size: 40,
                          color: Colors.blueGrey,
                        ),
                        const SizedBox(height: 8),
                        Text(
                          isMt
                              ? 'Upload Bukti Tagihan / Dokumen MT'
                              : 'Upload Bukti Transfer',
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
        if (_paymentMethod == 'TEMPO') ...[
          const SizedBox(height: 16),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'ESTIMASI JATUH TEMPO',
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                  color: Colors.blueGrey,
                  letterSpacing: 1.5,
                ),
              ),
              if (_selectedTempoDays != null)
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 2,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.amber.shade50,
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(color: Colors.amber.shade200),
                  ),
                  child: Text(
                    '${_selectedTempoDays} Hari',
                    style: GoogleFonts.outfit(
                      fontSize: 10,
                      fontWeight: FontWeight.w800,
                      color: Colors.amber.shade900,
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: 10),
          if (!_isReadOnly) ...[
            SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: Row(
                children: [
                  for (final days in [7, 14, 30, 45]) ...[
                    Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: InkWell(
                        borderRadius: BorderRadius.circular(10),
                        onTap: () {
                          setState(() {
                            _selectedTempoDays = days;
                            _paymentDueDate = DateTime.now().add(
                              Duration(days: days),
                            );
                          });
                        },
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 14,
                            vertical: 8,
                          ),
                          decoration: BoxDecoration(
                            color: _selectedTempoDays == days
                                ? const Color(0xFFD97706)
                                : const Color(0xFFF8FAFC),
                            borderRadius: BorderRadius.circular(10),
                            border: Border.all(
                              color: _selectedTempoDays == days
                                  ? const Color(0xFFD97706)
                                  : Colors.grey.shade300,
                            ),
                          ),
                          child: Text(
                            '$days Hari',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w700,
                              fontSize: 12,
                              color: _selectedTempoDays == days
                                  ? Colors.white
                                  : const Color(0xFF1E293B),
                            ),
                          ),
                        ),
                      ),
                    ),
                  ],
                  InkWell(
                    borderRadius: BorderRadius.circular(10),
                    onTap: () async {
                      final picked = await showDatePicker(
                        context: context,
                        initialDate:
                            _paymentDueDate ??
                            DateTime.now().add(const Duration(days: 14)),
                        firstDate: DateTime.now().subtract(
                          const Duration(days: 1),
                        ),
                        lastDate: DateTime.now().add(const Duration(days: 365)),
                      );
                      if (picked != null) {
                        setState(() {
                          _paymentDueDate = picked;
                          _selectedTempoDays = null;
                        });
                      }
                    },
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 8,
                      ),
                      decoration: BoxDecoration(
                        color:
                            _selectedTempoDays == null &&
                                _paymentDueDate != null
                            ? const Color(0xFFD97706)
                            : const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(
                          color:
                              _selectedTempoDays == null &&
                                  _paymentDueDate != null
                              ? const Color(0xFFD97706)
                              : Colors.grey.shade300,
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.calendar_month_rounded,
                            size: 14,
                            color:
                                _selectedTempoDays == null &&
                                    _paymentDueDate != null
                                ? Colors.white
                                : Colors.blueGrey,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            'Kustom',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w700,
                              fontSize: 12,
                              color:
                                  _selectedTempoDays == null &&
                                      _paymentDueDate != null
                                  ? Colors.white
                                  : const Color(0xFF1E293B),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 10),
          ],
          InkWell(
            onTap: _isReadOnly
                ? null
                : () async {
                    final picked = await showDatePicker(
                      context: context,
                      initialDate:
                          _paymentDueDate ??
                          DateTime.now().add(const Duration(days: 14)),
                      firstDate: DateTime.now().subtract(
                        const Duration(days: 1),
                      ),
                      lastDate: DateTime.now().add(const Duration(days: 365)),
                    );
                    if (picked != null) {
                      setState(() {
                        _paymentDueDate = picked;
                        _selectedTempoDays = null;
                      });
                    }
                  },
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
              decoration: BoxDecoration(
                color: const Color(0xFFF8FAFC),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(
                  color: _paymentDueDate != null
                      ? Colors.orange
                      : Colors.grey.shade200,
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    Icons.calendar_month_rounded,
                    color: _paymentDueDate != null
                        ? Colors.orange
                        : Colors.blueGrey,
                    size: 22,
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Text(
                      _paymentDueDate == null
                          ? 'Pilih Tanggal Estimasi Pembayaran...'
                          : DateFormat('dd MMMM yyyy').format(_paymentDueDate!),
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w700,
                        fontSize: 14,
                        color: _paymentDueDate == null
                            ? Colors.blueGrey
                            : const Color(0xFF1E293B),
                      ),
                    ),
                  ),
                  if (_paymentDueDate != null)
                    Text(
                      '(${_paymentDueDate!.difference(DateTime(DateTime.now().year, DateTime.now().month, DateTime.now().day)).inDays} Hari)',
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: Colors.orange.shade800,
                      ),
                    ),
                  if (!_isReadOnly)
                    const Icon(
                      Icons.arrow_drop_down_rounded,
                      color: Colors.blueGrey,
                      size: 24,
                    ),
                ],
              ),
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildPaymentAmountInput(NumberFormat currencyFormat) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'NOMINAL PEMBAYARAN',
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
            fontWeight: FontWeight.w900,
            fontSize: 24,
            color: const Color(0xFF1E293B),
          ),
          decoration: InputDecoration(
            filled: true,
            fillColor: Colors.white,
            prefixText: 'Rp ',
            prefixStyle: GoogleFonts.outfit(
              fontWeight: FontWeight.w900,
              color: Colors.blueGrey,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(20),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(20),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            contentPadding: const EdgeInsets.all(24),
          ),
          readOnly: _isReadOnly,
          onChanged: (v) => setState(() {}),
        ),
        const SizedBox(height: 12),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            if (!_isReadOnly)
              TextButton.icon(
                onPressed: _updateAmountFromTotal,
                icon: const Icon(Icons.auto_awesome_rounded, size: 16),
                label: Text(
                  'Isi Sesuai Tagihan',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.bold,
                    color: Colors.blueAccent,
                  ),
                ),
              ),
            if ((double.tryParse(_amountController.text) ?? 0) <
                _calculatedTotal)
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: Colors.orange.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  'Sisa: ${currencyFormat.format(_calculatedTotal - (double.tryParse(_amountController.text) ?? 0))}',
                  style: GoogleFonts.outfit(
                    color: Colors.orange.shade800,
                    fontWeight: FontWeight.w800,
                    fontSize: 11,
                  ),
                ),
              ),
          ],
        ),
      ],
    );
  }

  Widget _summaryRow(
    String label,
    String value, {
    bool isBold = false,
    Color? valueColor,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
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
                fontSize: isBold ? 16 : 14,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionTitle(String title) {
    return Text(
      title,
      style: GoogleFonts.outfit(
        fontWeight: FontWeight.w900,
        fontSize: 16,
        color: const Color(0xFF1E293B),
      ),
    );
  }

  List<dynamic> get _loadedCargoProducts {
    final Map<String, dynamic> productMap = {};
    final Map<String, int> stockMap = {};

    // 1. Extract from batch spares (barangan cadangan dari armada)
    final spares =
        (widget.batchData != null
                ? widget.batchData!['spares']
                : widget.deliveryItem['spares'])
            as List? ??
        [];
    for (final s in spares) {
      if (s['product'] != null) {
        final p = s['product'];
        final pId = p['id']?.toString() ?? '';
        if (pId.isNotEmpty) {
          productMap[pId] = p;
          stockMap[pId] = (stockMap[pId] ?? 0) + ((s['quantity'] ?? 0) as int);
        }
      }
    }

    // 2. Extract from batch delivery items (sales orders / transfers)
    final items =
        (widget.batchData != null
                ? widget.batchData!['items']
                : widget.deliveryItem['items'])
            as List? ??
        [];
    for (final item in items) {
      if (item['status'] == 'DELIVERED') continue;

      // Check sales order items
      final so = item['sales_order'] ?? item['sales_transaction'];
      if (so != null && so['items'] != null) {
        final soItems = so['items'] as List? ?? [];
        for (final soi in soItems) {
          if (soi['product'] != null) {
            final p = soi['product'];
            final pId = p['id']?.toString() ?? '';
            if (pId.isNotEmpty) {
              productMap[pId] = p;
              stockMap[pId] =
                  (stockMap[pId] ?? 0) + ((soi['quantity'] ?? 0) as int);
            }
          }
        }
      }
      // Check sales transfer items
      final transfer = item['sales_transfer'];
      if (transfer != null && transfer['product'] != null) {
        final p = transfer['product'];
        final pId = p['id']?.toString() ?? '';
        if (pId.isNotEmpty) {
          productMap[pId] = p;
          stockMap[pId] =
              (stockMap[pId] ?? 0) + ((transfer['quantity'] ?? 0) as int);
        }
      }
    }

    return productMap.entries.map((e) {
      final copy = Map<String, dynamic>.from(e.value as Map);
      copy['warehouse_stock'] = stockMap[e.key] ?? 0;
      return copy;
    }).toList();
  }

  int _getFleetStockForProduct(String productId) {
    final cargo = _loadedCargoProducts;
    if (cargo.isEmpty) return 0;
    int total = 0;
    for (var p in cargo) {
      if (p['id']?.toString() == productId) {
        total +=
            (p['warehouse_stock'] ?? p['quantity'] ?? p['stock'] ?? 0) as int;
      }
    }

    int ordered = 0;
    for (final item in _items) {
      final pIdStr = (item['product']?['id'] ?? item['product_id'])?.toString();
      if (pIdStr == productId) {
        ordered += (item['ordered_quantity'] ?? 0) as int;
      }
    }

    int reservedForExchange = 0;
    for (final key in _returnTypes.keys) {
      if (_returnTypes[key] == 'TUKAR_PRODUK') {
        if (_replacementProductIds[key] == productId) {
          reservedForExchange += (_returnedQuantities[key] ?? 0);
        }
      }
    }

    int available = total > ordered ? total - ordered : 0;
    available -= reservedForExchange;
    return available > 0 ? available : 0;
  }

  void _showProductSearchSheet(String pId) {
    final type = _returnTypes[pId] ?? 'POTONG_NOTA';
    final isTukarLangsung = type == 'TUKAR_PRODUK';
    final productsList =
        (isTukarLangsung ? _loadedCargoProducts : _availableProducts)
            .map((p) {
              final copy = Map<String, dynamic>.from(p);
              if (isTukarLangsung) {
                copy['warehouse_stock'] = _getFleetStockForProduct(
                  p['id']?.toString() ?? '',
                );
              }
              return copy;
            })
            .where((p) {
              final stock =
                  (p['warehouse_stock'] ?? p['quantity'] ?? p['stock'] ?? 0)
                      as int;
              return stock > 0;
            })
            .toList();
    final sheetTitle = isTukarLangsung
        ? 'Pilih Produk (Stok Armada)'
        : 'Pilih Produk (Stok Gudang)';

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (context) {
        return _ProductSearchSheet(
          products: productsList,
          initialSelectedId: _replacementProductIds[pId],
          title: sheetTitle,
          isArmada: isTukarLangsung,
          onSelected: (product) {
            setState(() {
              final replacementId = product['id']?.toString();
              _replacementProductIds[pId] = replacementId;
              _replacementProductNames[pId] = product['name']?.toString();
              _replacementPrices[pId] = (product['price'] ?? 0.0) is num
                  ? (product['price'] ?? 0.0).toDouble()
                  : 0.0;
            });
          },
        );
      },
    );
  }
}

class _ProductSearchSheet extends StatefulWidget {
  final List<dynamic> products;
  final String? initialSelectedId;
  final String title;
  final bool isArmada;
  final ValueChanged<Map<String, dynamic>> onSelected;

  const _ProductSearchSheet({
    required this.products,
    this.initialSelectedId,
    required this.title,
    required this.isArmada,
    required this.onSelected,
  });

  @override
  State<_ProductSearchSheet> createState() => _ProductSearchSheetState();
}

class _ProductSearchSheetState extends State<_ProductSearchSheet> {
  String _searchQuery = '';
  late List<dynamic> _filteredProducts;

  @override
  void initState() {
    super.initState();
    _filteredProducts = widget.products;
  }

  void _filterProducts(String query) {
    setState(() {
      _searchQuery = query;
      if (query.isEmpty) {
        _filteredProducts = widget.products;
      } else {
        _filteredProducts = widget.products.where((p) {
          final name = (p['name'] ?? '').toString().toLowerCase();
          final sku = (p['sku'] ?? '').toString().toLowerCase();
          final q = query.toLowerCase();
          return name.contains(q) || sku.contains(q);
        }).toList();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      height: MediaQuery.of(context).size.height * 0.75,
      padding: const EdgeInsets.fromLTRB(16, 20, 16, 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                widget.title,
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800,
                  fontSize: 16,
                  color: const Color(0xFF1E293B),
                ),
              ),
              IconButton(
                icon: const Icon(Icons.close_rounded),
                onPressed: () => Navigator.pop(context),
              ),
            ],
          ),
          const SizedBox(height: 12),
          TextField(
            autofocus: true,
            style: GoogleFonts.outfit(fontSize: 13),
            decoration: InputDecoration(
              hintText: 'Cari nama produk atau SKU...',
              hintStyle: GoogleFonts.outfit(color: Colors.grey, fontSize: 13),
              prefixIcon: const Icon(
                Icons.search_rounded,
                color: Colors.blueGrey,
                size: 20,
              ),
              filled: true,
              fillColor: const Color(0xFFF8FAFC),
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 12,
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: BorderSide(color: Colors.grey.shade200),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: const BorderSide(color: Colors.indigo),
              ),
            ),
            onChanged: _filterProducts,
          ),
          const SizedBox(height: 16),
          Expanded(
            child: _filteredProducts.isEmpty
                ? Center(
                    child: Text(
                      'Produk tidak ditemukan',
                      style: GoogleFonts.outfit(
                        color: Colors.grey,
                        fontSize: 13,
                      ),
                    ),
                  )
                : ListView.builder(
                    itemCount: _filteredProducts.length,
                    itemBuilder: (context, index) {
                      final p = _filteredProducts[index];
                      final pId = p['id']?.toString() ?? '';
                      final isSelected = widget.initialSelectedId == pId;
                      final stock = p['warehouse_stock'] ?? 0;
                      return Container(
                        margin: const EdgeInsets.only(bottom: 8),
                        decoration: BoxDecoration(
                          color: isSelected
                              ? Colors.indigo.withOpacity(0.05)
                              : Colors.white,
                          borderRadius: BorderRadius.circular(14),
                          border: Border.all(
                            color: isSelected
                                ? Colors.indigo
                                : Colors.grey.shade200,
                            width: isSelected ? 1.5 : 1,
                          ),
                        ),
                        child: ListTile(
                          onTap: () {
                            widget.onSelected(p as Map<String, dynamic>);
                            Navigator.pop(context);
                          },
                          title: Text(
                            p['name']?.toString() ?? 'Produk',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.bold,
                              fontSize: 13,
                              color: const Color(0xFF1E293B),
                            ),
                          ),
                          subtitle: Row(
                            children: [
                              Text(
                                'SKU: ${p['sku'] ?? '-'}',
                                style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  color: Colors.grey.shade600,
                                ),
                              ),
                              const SizedBox(width: 10),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 6,
                                  vertical: 2,
                                ),
                                decoration: BoxDecoration(
                                  color: stock > 0
                                      ? Colors.green.withOpacity(0.1)
                                      : Colors.red.withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(4),
                                ),
                                child: Text(
                                  widget.isArmada
                                      ? 'Muatan: $stock pcs'
                                      : 'Stok: $stock pcs',
                                  style: GoogleFonts.outfit(
                                    fontSize: 10,
                                    fontWeight: FontWeight.bold,
                                    color: stock > 0
                                        ? Colors.green.shade700
                                        : Colors.red.shade700,
                                  ),
                                ),
                              ),
                            ],
                          ),
                          trailing: isSelected
                              ? const Icon(
                                  Icons.check_circle_rounded,
                                  color: Colors.indigo,
                                  size: 20,
                                )
                              : null,
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}

class SignatureCanvas extends StatefulWidget {
  final Function(List<Offset>) onDraw;
  final List<Offset> points;
  final VoidCallback onClear;
  final String label;

  const SignatureCanvas({
    super.key,
    required this.onDraw,
    required this.points,
    required this.onClear,
    required this.label,
  });

  @override
  _SignatureCanvasState createState() => _SignatureCanvasState();
}

class _SignatureCanvasState extends State<SignatureCanvas> {
  List<Offset> _points = [];

  @override
  void initState() {
    super.initState();
    _points = List.from(widget.points);
  }

  @override
  void didUpdateWidget(SignatureCanvas oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.points != oldWidget.points) {
      _points = List.from(widget.points);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Expanded(
              child: Text(
                widget.label,
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: const Color(0xFF64748B),
                ),
              ),
            ),
            GestureDetector(
              onTap: () {
                setState(() => _points.clear());
                widget.onClear();
              },
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                child: Text(
                  'Hapus',
                  style: GoogleFonts.outfit(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: Colors.redAccent,
                  ),
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Container(
          height: 120,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: const Color(0xFFE2E8F0), width: 1.5),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.02),
                blurRadius: 10,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(16),
            child: GestureDetector(
              behavior: HitTestBehavior.opaque,
              onVerticalDragStart: (_) {},
              onVerticalDragUpdate: (_) {},
              onHorizontalDragStart: (_) {},
              onHorizontalDragUpdate: (_) {},
              child: Listener(
                onPointerDown: (PointerDownEvent event) {
                  setState(() {
                    _points.add(event.localPosition);
                  });
                  widget.onDraw(List.from(_points));
                },
                onPointerMove: (PointerMoveEvent event) {
                  setState(() {
                    _points.add(event.localPosition);
                  });
                  widget.onDraw(List.from(_points));
                },
                onPointerUp: (PointerUpEvent event) {
                  setState(() {
                    _points.add(Offset.infinite);
                  });
                  widget.onDraw(List.from(_points));
                },
                child: CustomPaint(
                  painter: SignaturePainter(points: _points),
                  size: Size.infinite,
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class SignaturePainter extends CustomPainter {
  final List<Offset> points;

  SignaturePainter({required this.points});

  @override
  void paint(Canvas canvas, Size size) {
    Paint paint = Paint()
      ..color = const Color(0xFF0F172A)
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round
      ..strokeWidth = 4.5
      ..style = PaintingStyle.stroke;

    for (int i = 0; i < points.length; i++) {
      if (points[i] != Offset.infinite) {
        if (i == 0 || points[i - 1] == Offset.infinite) {
          // Draw a dot if it's a tap or start of a line
          canvas.drawCircle(
            points[i],
            paint.strokeWidth / 2.5,
            Paint()..color = paint.color,
          );
        } else if (points[i - 1] != Offset.infinite) {
          // Verify bounds to keep rendering within the canvas area
          if (points[i - 1].dx >= 0 &&
              points[i - 1].dx <= size.width &&
              points[i - 1].dy >= 0 &&
              points[i - 1].dy <= size.height &&
              points[i].dx >= 0 &&
              points[i].dx <= size.width &&
              points[i].dy >= 0 &&
              points[i].dy <= size.height) {
            canvas.drawLine(points[i - 1], points[i], paint);
          }
        }
      }
    }
  }

  @override
  bool shouldRepaint(SignaturePainter oldDelegate) => true;
}
