import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:get_it/get_it.dart';
import 'package:geolocator/geolocator.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'dart:convert';
import 'dart:io';
import 'dart:ui' as ui;
import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:intl/intl.dart';
import 'package:uuid/uuid.dart';
import 'package:drift/drift.dart' hide Column;
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import './visit_success_page.dart';
import './delivery_payment_info_page.dart';
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';
import 'package:hris_app/features/sales/data/services/banner_order_api_service.dart';
import 'package:image_picker/image_picker.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/services.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/utils/image_compress_helper.dart';

class VisitCheckoutPage extends StatefulWidget {
  final StoreModel store;
  final String selfiePath;
  final Uint8List? selfieBytes;
  final String? receiptPath;
  final String? companyId;
  final String? companyName;
  final List<Map<String, dynamic>>? items;
  final List<Map<String, dynamic>>? ancillaryItems;
  final double? totalAmount;
  final String? notes;
  final String jobPositionTitle;
  final bool? isDistributor;

  const VisitCheckoutPage({
    super.key,
    required this.store,
    required this.selfiePath,
    this.selfieBytes,
    this.receiptPath,
    this.companyId,
    this.companyName,
    this.items,
    this.ancillaryItems,
    this.totalAmount,
    this.notes,
    required this.jobPositionTitle,
    this.isDistributor,
  });

  @override
  State<VisitCheckoutPage> createState() => _VisitCheckoutPageState();
}

class _VisitCheckoutPageState extends State<VisitCheckoutPage> {
  final ImagePicker _picker = ImagePicker();
  String? _selfiePath;
  Uint8List? _selfieBytes;
  String? _receiptPath;

  Future<void> _retakeReceiptPhoto() async {
    final XFile? photo = await _picker.pickImage(
      source: ImageSource.camera,
      imageQuality: 70,
    );
    if (photo != null) {
      setState(() {
        _receiptPath = photo.path;
      });
    }
  }

  Future<void> _retakeSelfiePhoto() async {
    final XFile? photo = await _picker.pickImage(
      maxWidth: 1024,
      maxHeight: 1024,
      source: ImageSource.camera,
      imageQuality: 70,
    );
    if (photo != null) {
      setState(() {
        _selfiePath = photo.path;
        _selfieBytes = null;
      });
    }
  }

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

  bool _isSubmitting = false;
  final AppDatabase db = di.sl<AppDatabase>();
  String _selectedPaymentMethod = 'CASH'; // 'CASH', 'TEMPO', 'VA'
  String? _selectedBank; // 'bca', 'bni', 'bri'
  DateTime? _paymentDueDate;
  int? _selectedTempoDays = 14;
  Map<String, dynamic>? _activeBankAccount;
  List<Map<String, dynamic>> _bankAccounts = [];
  Map<String, dynamic>? _selectedBankAccount;
  bool _isLoadingBanks = false;
  String _branchId = '';
  List<Offset> _courierSignaturePoints = [];
  List<Offset> _customerSignaturePoints = [];

  double _downPaymentAmount = 0.0;
  int _installmentCount = 1;

  String _companyName = '';
  String _companyId = '';
  // MT mode: dibaca dari SharedPreferences (ditetapkan oleh SalesDashboardPage)
  bool _isMtMode = false;
  bool _isToMode = false;
  String? _factoryCompanyId;

  bool _wantsBanner = false;
  String _bannerType = 'Custom';
  List<Map<String, dynamic>> _selectedBonuses = [];
  List<Map<String, dynamic>> _selectedReturns = [];

  double get _totalReturnAmount {
    return _selectedReturns.fold(0.0, (sum, item) {
      return sum + ((item['price'] as num) * (item['quantity'] as int));
    });
  }

  double get _finalTotalAmount {
    final baseTotal = widget.totalAmount ?? 0.0;
    return baseTotal - _totalReturnAmount;
  }

  String _bannerUnit = 'm'; // 'm' or 'cm'
  final _bannerLengthCtrl = TextEditingController();
  final _bannerWidthCtrl = TextEditingController();
  final _bannerQtyCtrl = TextEditingController(text: '1');
  final _bannerNotesCtrl = TextEditingController();
  Uint8List? _bannerAttachmentBytes;
  String? _bannerAttachmentPath;
  String? _bannerAttachmentName;
  bool _isBannerPdfAttachment = false;

  final _manualReceiptCtrl = TextEditingController();
  final _manualTotalAmountCtrl = TextEditingController();

  // --- Modern Trade (MT) ---
  final _poNumberCtrl = TextEditingController();
  String _fulfillmentSource = 'FACTORY'; // 'FACTORY' or 'BRANCH'
  String? _selectedFulfillmentBranchId;
  List<Map<String, dynamic>> _branches = [];
  bool _isLoadingBranches = false;

  final Map<String, Map<String, dynamic>> _paymentOptions = {
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
    'SALES_ORDER': {
      'label': 'Sales Order (Tagih saat Kirim)',
      'icon': Icons.description_rounded,
      'color': Colors.blueGrey,
    },
  };

  bool _isDistributorProfile = false;

  bool get _isSalesDistributor {
    if (widget.isDistributor == true) return true;
    if (_isDistributorProfile) return true;
    final title = widget.jobPositionTitle.toLowerCase();
    return title.contains('distributor') ||
        widget.store.customerType.toUpperCase() == 'DISTRIBUTOR';
  }

  // Untuk Sales Distributor: pilihan pemenuhan manual ('DIRECT' = Motoris Langsung, 'SO' = Pesanan PO Gudang)
  String _distributorFulfillmentMode = 'DIRECT';

  bool get _isTaskOrder {
    if (_isSalesDistributor) {
      return _distributorFulfillmentMode == 'SO';
    }
    final title = widget.jobPositionTitle.toLowerCase();
    return title.contains('task order') ||
        title.contains('take order') ||
        title.contains('sales to') ||
        title.contains('supervisor') ||
        title.contains('spv') ||
        _isMtMode ||
        _isToMode;
  }

  @override
  void initState() {
    super.initState();
    if (widget.isDistributor == true) {
      _isDistributorProfile = true;
    }
    _bannerLengthCtrl.addListener(_onBannerDimensionChanged);
    _bannerWidthCtrl.addListener(_onBannerDimensionChanged);
    _selfiePath = widget.selfiePath;
    _selfieBytes = widget.selfieBytes;
    _receiptPath = widget.receiptPath;
    _companyName = widget.companyName ?? 'Perusahaan Demo';
    _companyId = widget.companyId ?? '00000000-0000-0000-0000-000000000000';
    _fetchProfile();
    final jobUpper = widget.jobPositionTitle.toUpperCase();
    final jobLower = widget.jobPositionTitle.toLowerCase();
    final isSpv = jobLower.contains('supervisor') || jobLower.contains('spv');
    final isMtFromJob =
        jobUpper.contains('MT') ||
        jobUpper.contains('MODERN') ||
        jobUpper.contains('KAM') ||
        jobUpper.contains('MERCHANDIS');

    if (isMtFromJob) {
      _isMtMode = true;
      _selectedPaymentMethod = 'SALES_ORDER';
    } else if (isSpv) {
      _isToMode = true;
      if (!_isSalesDistributor) {
        _selectedPaymentMethod = 'SALES_ORDER';
      }
    }

    // Baca mode MT/TO dari SharedPreferences
    SharedPreferences.getInstance().then((prefs) {
      final salesMode = prefs.getString('sales_mode');
      if (mounted && (salesMode == 'MT' || isMtFromJob)) {
        setState(() {
          _isMtMode = true;
          _factoryCompanyId = prefs.getString('factory_company_id');
          _selectedPaymentMethod = 'SALES_ORDER';
        });
        _fetchBranches();
      } else if (mounted && (salesMode == 'TO' || isSpv)) {
        setState(() {
          _isToMode = true;
          if (!_isSalesDistributor) {
            _selectedPaymentMethod = 'SALES_ORDER';
          }
        });
      } else if (_isTaskOrder) {
        setState(() => _selectedPaymentMethod = 'SALES_ORDER');
      }
    });
  }

  void _onBannerDimensionChanged() {
    if (mounted) setState(() {});
  }

  void _switchBannerUnit(String newUnit) {
    if (_bannerUnit == newUnit) return;
    final lenStr = _bannerLengthCtrl.text.trim().replaceAll(',', '.');
    final widStr = _bannerWidthCtrl.text.trim().replaceAll(',', '.');
    final lenVal = double.tryParse(lenStr);
    final widVal = double.tryParse(widStr);

    setState(() {
      _bannerUnit = newUnit;
      if (lenVal != null && lenVal > 0) {
        if (newUnit == 'cm') {
          final converted = lenVal * 100;
          _bannerLengthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(1);
        } else {
          final converted = lenVal / 100;
          _bannerLengthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(2);
        }
      }
      if (widVal != null && widVal > 0) {
        if (newUnit == 'cm') {
          final converted = widVal * 100;
          _bannerWidthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(1);
        } else {
          final converted = widVal / 100;
          _bannerWidthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(2);
        }
      }
    });
  }

  void _applyBannerPreset(double lengthM, double widthM) {
    setState(() {
      if (_bannerUnit == 'm') {
        _bannerLengthCtrl.text = lengthM % 1 == 0
            ? lengthM.toInt().toString()
            : lengthM.toString();
        _bannerWidthCtrl.text = widthM % 1 == 0
            ? widthM.toInt().toString()
            : widthM.toString();
      } else {
        final lengthCm = lengthM * 100;
        final widthCm = widthM * 100;
        _bannerLengthCtrl.text = lengthCm.toInt().toString();
        _bannerWidthCtrl.text = widthCm.toInt().toString();
      }
    });
  }

  Future<void> _pickBannerImage(ImageSource source) async {
    try {
      final picker = ImagePicker();
      final pickedFile = await picker.pickImage(
        source: source,
        maxWidth: 1600,
        maxHeight: 1600,
        imageQuality: 80,
      );
      if (pickedFile != null) {
        final bytes = await pickedFile.readAsBytes();
        setState(() {
          _bannerAttachmentBytes = bytes;
          _bannerAttachmentPath = pickedFile.path;
          _bannerAttachmentName = pickedFile.name;
          _isBannerPdfAttachment = false;
        });
      }
    } catch (e) {
      debugPrint('Error picking banner image: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal mengambil foto: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _pickBannerDocument() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['pdf', 'jpg', 'jpeg', 'png'],
        withData: true,
      );
      if (result != null && result.files.isNotEmpty) {
        final file = result.files.first;
        final name = file.name;
        final isPdf = name.toLowerCase().endsWith('.pdf');
        Uint8List? bytes = file.bytes;

        setState(() {
          _bannerAttachmentBytes = bytes;
          _bannerAttachmentPath = file.path;
          _bannerAttachmentName = name;
          _isBannerPdfAttachment = isPdf;
        });
      }
    } catch (e) {
      debugPrint('Error picking banner document: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal memilih file: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _fetchBranches() async {
    setState(() => _isLoadingBranches = true);
    try {
      final prefs = await SharedPreferences.getInstance();
      final cached = prefs.getString('cached_branches_list');
      if (cached != null && _branches.isEmpty) {
        final decoded = (jsonDecode(cached) as List)
            .map((e) => Map<String, dynamic>.from(e))
            .toList();
        if (decoded.isNotEmpty && mounted) {
          setState(() {
            _branches = decoded;
            _isLoadingBranches = false;
            _selectedFulfillmentBranchId =
                (_branches[0]['id'] ?? _branches[0]['ID'])?.toString();
          });
        }
      }
      final result = await di.sl<AuthRepository>().getBranches();
      result.fold((l) => debugPrint("Error fetching branches: ${l.message}"), (
        branches,
      ) {
        prefs.setString('cached_branches_list', jsonEncode(branches)).ignore();
        if (mounted) {
          setState(() {
            _branches = branches;
            _isLoadingBranches = false;
            if (_branches.isNotEmpty && _selectedFulfillmentBranchId == null) {
              _selectedFulfillmentBranchId =
                  (_branches[0]['id'] ?? _branches[0]['ID'])?.toString();
            }
          });
        }
      });
    } catch (e) {
      debugPrint("Exception fetching branches: $e");
      if (mounted) {
        setState(() => _isLoadingBranches = false);
      }
    }
  }

  @override
  void dispose() {
    _bannerLengthCtrl.removeListener(_onBannerDimensionChanged);
    _bannerWidthCtrl.removeListener(_onBannerDimensionChanged);
    _bannerLengthCtrl.dispose();
    _bannerWidthCtrl.dispose();
    _bannerQtyCtrl.dispose();
    _bannerNotesCtrl.dispose();
    _manualReceiptCtrl.dispose();
    _manualTotalAmountCtrl.dispose();
    _poNumberCtrl.dispose();
    super.dispose();
  }

  Future<void> _fetchBankAccounts() async {
    setState(() => _isLoadingBanks = true);
    try {
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

  Future<void> _fetchProfile() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final savedBranchId = prefs.getString('branch_id');
      if (savedBranchId != null && savedBranchId.isNotEmpty) {
        _branchId = savedBranchId;
      }
      final cachedProfile = prefs.getString('cached_employee_profile');
      if (cachedProfile != null) {
        final profile = jsonDecode(cachedProfile) as Map<String, dynamic>;
        final cId =
            profile['company_id'] ??
            profile['companyId'] ??
            profile['CompanyID'];
        final cName =
            profile['company_name'] ??
            profile['companyName'] ??
            profile['CompanyNAME'] ??
            profile['company']?['name'];
        final bId =
            profile['branch_id'] ??
            profile['branchId'] ??
            profile['BranchID'] ??
            profile['branch']?['id'];
        final isDist =
            profile['is_distributor'] == true ||
            (profile['distributor_store_id'] != null &&
                profile['distributor_store_id'].toString().isNotEmpty);
        if (mounted) {
          setState(() {
            if (isDist) {
              _isDistributorProfile = true;
              if (_distributorFulfillmentMode == 'DIRECT' &&
                  _selectedPaymentMethod == 'SALES_ORDER') {
                _selectedPaymentMethod = 'CASH';
              }
            }
            if (widget.companyName == null && cName != null)
              _companyName = cName.toString();
            if (widget.companyId == null && cId != null)
              _companyId = cId.toString();
            if (_branchId.isEmpty && bId != null) _branchId = bId.toString();
          });
        }
      }
    } catch (_) {}

    _fetchBankAccounts();
  }

  final List<Map<String, String>> _bankOptions = [
    {'id': 'bca', 'name': 'BCA Virtual Account'},
    {'id': 'bni', 'name': 'BNI Virtual Account'},
    {'id': 'bri', 'name': 'BRI Virtual Account'},
    {'id': 'mandiri', 'name': 'Mandiri Bill Payment'},
  ];

  Future<void> _submitAndCheckout() async {
    if (_isSubmitting) return;
    final bool isPhysicalReceipt = widget.items == null && _receiptPath != null;
    final bool isNoTransaction = widget.items == null && _receiptPath == null;
    // "Input Nanti" tetap kena timer — hanya kunjungan "Tidak Ada Transaksi" murni yang bypass
    final bool isInputNanti =
        isNoTransaction && widget.notes == 'Input Transaksi Nanti';
    final bool shouldCheckTimer = !isNoTransaction || isInputNanti;

    if (shouldCheckTimer) {
      try {
        final latestCheckin =
            await (db.select(db.localCheckins)
                  ..where((t) => t.storeId.equals(widget.store.id))
                  ..orderBy([
                    (t) => OrderingTerm(
                      expression: t.createdAt,
                      mode: OrderingMode.desc,
                    ),
                  ])
                  ..limit(1))
                .getSingleOrNull();

        if (latestCheckin != null) {
          final diff = DateTime.now().difference(latestCheckin.createdAt);
          if (diff.inMinutes < 5) {
            final bool? shouldProceed = await showDialog<bool>(
              context: context,
              builder: (context) => AlertDialog(
                title: Text(
                  'Peringatan Waktu Kunjungan',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w800,
                    color: Colors.orange,
                  ),
                ),
                content: Text(
                  'Waktu interaksi Anda di toko belum mencapai 5 menit (Baru berjalan ${diff.inMinutes} menit).\n\nKunjungan yang terlalu cepat akan mempengaruhi penilaian KPI Anda.\n\nApakah Anda yakin ingin melanjutkan checkout?',
                  style: GoogleFonts.outfit(),
                ),
                actions: [
                  TextButton(
                    onPressed: () => Navigator.pop(context, false),
                    child: Text(
                      'TUNGGU',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w700,
                        color: Colors.grey,
                      ),
                    ),
                  ),
                  ElevatedButton(
                    onPressed: () => Navigator.pop(context, true),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.orange,
                    ),
                    child: Text(
                      'TETAP CHECKOUT',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w700,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ],
              ),
            );

            if (shouldProceed != true) {
              return;
            }
          }
        }
      } catch (e) {
        debugPrint('Error checking checkin time: $e');
      }
    }

    if (isPhysicalReceipt) {
      if (_manualTotalAmountCtrl.text.trim().isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Harap isi nominal transaksi dari nota fisik.',
              style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
            ),
            backgroundColor: Colors.orange,
          ),
        );
        return;
      }
    } else {
      if (_selectedPaymentMethod == 'TEMPO' && _paymentDueDate == null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'Harap tentukan tanggal jatuh tempo untuk metode TEMPO.',
            ),
          ),
        );
        return;
      }

      // Direct Sales Signature Validation
      if (widget.items != null && widget.items!.isNotEmpty && !_isTaskOrder) {
        if (_courierSignaturePoints.isEmpty ||
            _customerSignaturePoints.isEmpty) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                'Tanda tangan sales dan customer wajib diisi.',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
              ),
              backgroundColor: Colors.orange,
            ),
          );
          return;
        }
      }
    }

    if (_wantsBanner) {
      if (_bannerLengthCtrl.text.trim().isEmpty ||
          _bannerWidthCtrl.text.trim().isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Harap tentukan ukuran lebar dan tinggi spanduk.',
              style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
            ),
            backgroundColor: Colors.orange,
          ),
        );
        return;
      }
      final len = double.tryParse(
        _bannerLengthCtrl.text.trim().replaceAll(',', '.'),
      );
      final wid = double.tryParse(
        _bannerWidthCtrl.text.trim().replaceAll(',', '.'),
      );
      if (len == null || len <= 0 || wid == null || wid <= 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Ukuran spanduk harus berupa angka positif.',
              style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
            ),
            backgroundColor: Colors.orange,
          ),
        );
        return;
      }
      final qty = int.tryParse(_bannerQtyCtrl.text.trim()) ?? 0;
      if (qty <= 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Jumlah (Qty) spanduk minimal 1.',
              style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
            ),
            backgroundColor: Colors.orange,
          ),
        );
        return;
      }
    }

    if (widget.items != null && widget.items!.isNotEmpty) {
      for (final item in widget.items!) {
        final pId = item['product_id']?.toString() ?? '';
        if (pId.isEmpty || pId == '00000000-0000-0000-0000-000000000000') {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                'Terjadi kesalahan integritas data (Product ID tidak valid). Silakan hapus item dan coba lagi, atau lakukan Force Sync data master.',
                style: GoogleFonts.outfit(),
              ),
              backgroundColor: Colors.red,
              duration: const Duration(seconds: 4),
            ),
          );
          return;
        }
      }
    }

    setState(() => _isSubmitting = true);

    Position? position;
    try {
      position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
        timeLimit: const Duration(seconds: 5),
      );
    } catch (_) {
      // Ignore if permission denied or timeout
    }

    try {
      final localId = const Uuid().v4();
      final now = DateTime.now();

      // Render & save signatures to local compressed files for offline direct sales sync
      Uint8List? courierSigBytes;
      Uint8List? customerSigBytes;
      String? combinedSignaturePath;

      if (widget.items != null && widget.items!.isNotEmpty && !_isTaskOrder) {
        try {
          courierSigBytes = await _renderSignatureToPng(
            _courierSignaturePoints,
          );
          customerSigBytes = await _renderSignatureToPng(
            _customerSignaturePoints,
          );
          if (courierSigBytes != null && customerSigBytes != null) {
            final courierPath = await ImageCompressHelper.saveSignatureToFile(
              courierSigBytes,
              prefix: 'sig_courier_${localId.substring(0, 6)}',
            );
            final customerPath = await ImageCompressHelper.saveSignatureToFile(
              customerSigBytes,
              prefix: 'sig_customer_${localId.substring(0, 6)}',
            );
            combinedSignaturePath = '$courierPath,$customerPath';
          }
        } catch (e) {
          print('Gagal render ttd: $e');
        }
      }

      // Generate Invoice Receipt Number
      final String companyPrefix = _companyName
          .replaceAll(' ', '')
          .toUpperCase();
      final receiptNo = isPhysicalReceipt
          ? 'PHYS-$companyPrefix-${DateFormat('yyyyMMdd').format(now)}-${localId.substring(0, 4).toUpperCase()}'
          : 'INV-$companyPrefix-${DateFormat('yyyyMMdd').format(now)}-${localId.substring(0, 4).toUpperCase()}';

      final double totalAmountValue = isPhysicalReceipt
          ? (double.tryParse(
                  _manualTotalAmountCtrl.text.trim().replaceAll(
                    RegExp(r'[^0-9]'),
                    '',
                  ),
                ) ??
                0.0)
          : _finalTotalAmount;

      String? receiptPathValue = _selectedPaymentMethod == 'VA'
          ? _transferProofImage?.path
          : _receiptPath;

      if (kIsWeb &&
          _selectedPaymentMethod == 'VA' &&
          _transferProofImage != null) {
        try {
          final bytes = await _transferProofImage!.readAsBytes();
          receiptPathValue = 'data:image/png;base64,${base64Encode(bytes)}';
        } catch (e) {
          print('Error converting transfer proof image to base64 on web: $e');
        }
      }

      // ═══════════════════════════════════════════════════════════════
      // PENJAMINAN DATA: Semua operasi DB dibungkus dalam satu transaksi atomik.
      // Jika app crash atau terjadi error di tengah proses, SEMUA perubahan
      // akan otomatis di-rollback, mencegah data partial/corrupt.
      // ═══════════════════════════════════════════════════════════════
      await db.transaction(() async {
        // 1. Save Transaction Header to Local DB
        await db
            .into(db.localTransactions)
            .insert(
              LocalTransactionsCompanion.insert(
                localId: localId,
                companyId: _companyId,
                companyName: _companyName,
                storeId: widget.store.id,
                storeName: widget.store.name,
                totalAmount: totalAmountValue,
                receiptNo: Value(receiptNo),
                paidAmount: Value(
                  _isTaskOrder
                      ? 0.0
                      : (isPhysicalReceipt
                            ? totalAmountValue
                            : (_selectedPaymentMethod == 'TEMPO'
                                  ? _downPaymentAmount
                                  : 0.0)),
                ),
                selfiePath: Value(_selfiePath),
                receiptPath: Value(receiptPathValue),
                signaturePath: Value(combinedSignaturePath),
                paymentMethod: Value(
                  _isTaskOrder
                      ? 'SALES_ORDER'
                      : (isPhysicalReceipt ? 'CASH' : _selectedPaymentMethod),
                ),
                paymentBank: Value(_selectedBank),
                paymentDueDate: Value(_paymentDueDate),
                latitude: position != null
                    ? Value(position.latitude)
                    : const Value.absent(),
                longitude: position != null
                    ? Value(position.longitude)
                    : const Value.absent(),
                tenor: Value(
                  _selectedPaymentMethod == 'TEMPO' ? _installmentCount : 1,
                ),
                notes: Value(() {
                  // Encode ancillary_items along with notes so it can be extracted later by sync_repository
                  final notePayload = <String, dynamic>{};
                  if (_isMtMode) {
                    notePayload['po_number'] = '';
                    notePayload['fulfillment_type'] = 'FACTORY';
                    notePayload['notes'] = widget.notes ?? '';
                  } else if (widget.notes != null && widget.notes!.isNotEmpty) {
                    notePayload['notes'] = widget.notes;
                  }
                  if (widget.ancillaryItems != null &&
                      widget.ancillaryItems!.isNotEmpty) {
                    notePayload['ancillary_items'] = widget.ancillaryItems;
                  }
                  if (_selectedBonuses.isNotEmpty) {
                    notePayload['bonus_items'] = _selectedBonuses;
                  }
                  if (_selectedReturns.isNotEmpty) {
                    notePayload['return_items'] = _selectedReturns;
                  }

                  if (notePayload.isNotEmpty) {
                    return jsonEncode(notePayload);
                  }
                  return null;
                }()),
                createdAt: now,
                syncStatus: const Value('pending'),
                // MT flag: faktur dikelola pusat/pabrik
                isMt: Value(_isMtMode),
                isDistributorOrder: Value(_isSalesDistributor),
                factoryCompanyId: Value(_factoryCompanyId),
                companyBankAccountId: Value(
                  _activeBankAccount != null
                      ? _activeBankAccount!['id'] as String?
                      : null,
                ),
              ),
            );

        // 2. Save Transaction Items to Local DB
        if (widget.items != null) {
          for (final item in widget.items!) {
            await db
                .into(db.localTransactionItems)
                .insert(
                  LocalTransactionItemsCompanion.insert(
                    transactionLocalId: localId,
                    productId: item['product_id'],
                    productName: item['product_name'],
                    quantity: item['quantity'],
                    orderedQuantity: Value(item['ordered_quantity'] ?? 0),
                    unit: Value(item['unit'] ?? 'PCS'),
                    piecesPerUnit: Value(item['pieces_per_unit'] ?? 1),
                    price: (item['price'] as num).toDouble(),
                    discountPercent: Value(
                      (item['discount_percent'] as num? ?? 0.0).toDouble(),
                    ),
                  ),
                );
          }

          // 2b. Immediately deduct sold quantities from local salesStock
          // This gives instant dashboard feedback before server sync completes.
          // Only deduct local stock for direct sales (NOT for Sales Orders / PO Gudang)
          if (!_isTaskOrder) {
            for (final item in widget.items!) {
              final productId = item['product_id'] as String;
              final soldQty = (item['quantity'] as num).toInt();

              final existing = await (db.select(
                db.salesStock,
              )..where((t) => t.productId.equals(productId))).getSingleOrNull();

              if (existing != null) {
                final newQty = (existing.quantity - soldQty).clamp(0, 999999);
                await (db.update(
                  db.salesStock,
                )..where((t) => t.productId.equals(productId))).write(
                  SalesStockCompanion(
                    quantity: Value(newQty),
                    updatedAt: Value(DateTime.now()),
                  ),
                );
              }
            }
          }
        } // ← tutup if (widget.items != null)

        // 2c. Selalu potong stok bonus dari salesStock lokal,
        //     TERLEPAS apakah ada item reguler atau tidak.
        if (_selectedBonuses.isNotEmpty) {
          for (final bonus in _selectedBonuses) {
            final productId = bonus['product_id'] as String;
            final bonusQty = (bonus['quantity'] as num).toInt();

            final existing = await (db.select(
              db.salesStock,
            )..where((t) => t.productId.equals(productId))).getSingleOrNull();

            if (existing != null) {
              final newQty = (existing.quantity - bonusQty).clamp(0, 999999);
              await (db.update(
                db.salesStock,
              )..where((t) => t.productId.equals(productId))).write(
                SalesStockCompanion(
                  quantity: Value(newQty),
                  updatedAt: Value(DateTime.now()),
                ),
              );
            }
          }
        }

        if (_selectedReturns.isNotEmpty) {
          for (final retur in _selectedReturns) {
            final productId = retur['product_id'] as String;
            final returQty = (retur['quantity'] as num).toInt();

            final existing = await (db.select(
              db.salesStock,
            )..where((t) => t.productId.equals(productId))).getSingleOrNull();

            if (existing != null) {
              // For return, the stock increases in the motoris inventory
              final newQty = existing.quantity + returQty;
              await (db.update(
                db.salesStock,
              )..where((t) => t.productId.equals(productId))).write(
                SalesStockCompanion(
                  quantity: Value(newQty),
                  updatedAt: Value(DateTime.now()),
                ),
              );
            } else {
              await db
                  .into(db.salesStock)
                  .insert(
                    SalesStockCompanion.insert(
                      productId: productId,
                      quantity: returQty,
                      updatedAt: DateTime.now(),
                    ),
                  );
            }
          }
        }

        // 2d. Mark local VisitPlans status as COMPLETED for today
        await (db.update(db.visitPlans)
              ..where((t) => t.storeId.equals(widget.store.id)))
            .write(const VisitPlansCompanion(status: Value('COMPLETED')));
      }); // ← akhir blok db.transaction() — jika error terjadi sebelum ini, semua rollback

      // Update cached_visit_plans_$todayStr status to COMPLETED as well
      try {
        final prefs = await SharedPreferences.getInstance();
        final todayStr = DateFormat('yyyy-MM-dd').format(now);
        final cachedStr = prefs.getString('cached_visit_plans_$todayStr');
        if (cachedStr != null) {
          final List<dynamic> list = jsonDecode(cachedStr);
          for (var item in list) {
            if (item is Map &&
                item['store_id']?.toString() == widget.store.id) {
              item['status'] = 'COMPLETED';
            }
          }
          await prefs.setString(
            'cached_visit_plans_$todayStr',
            jsonEncode(list),
          );
        }
      } catch (_) {}

      // 2c. If banner request is checked, submit banner order to the server
      if (_wantsBanner) {
        try {
          final double length = double.parse(
            _bannerLengthCtrl.text.trim().replaceAll(',', '.'),
          );
          final double width = double.parse(
            _bannerWidthCtrl.text.trim().replaceAll(',', '.'),
          );
          final lengthCm = _bannerUnit == 'm' ? (length * 100) : length;
          final widthCm = _bannerUnit == 'm' ? (width * 100) : width;
          final qty = int.tryParse(_bannerQtyCtrl.text.trim()) ?? 1;

          final bannerPayload = {
            'shop_id': widget.store.id,
            'shop_name': widget.store.name,
            'size_length': lengthCm,
            'size_width': widthCm,
            'type': _bannerType,
            'is_bihun': false,
            'qty': qty,
            'notes': _bannerNotesCtrl.text.trim(),
          };

          final bannerApiService = BannerOrderApiService(
            apiClient: di.sl<ApiClient>(),
          );
          await bannerApiService.createBannerOrder(
            bannerPayload,
            attachmentPath: _bannerAttachmentPath,
            attachmentBytes: _bannerAttachmentBytes,
            attachmentName: _bannerAttachmentName,
          );
          debugPrint('Request spanduk berhasil dibuat online.');
        } catch (bannerErr) {
          if (bannerErr.toString().contains('OFFLINE_SAVED')) {
            debugPrint('Request spanduk disimpan offline.');
            if (mounted) {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text(
                    'Koneksi offline. Request spanduk disimpan secara offline.',
                  ),
                  backgroundColor: Colors.orange,
                  duration: Duration(seconds: 2),
                ),
              );
            }
          } else {
            debugPrint('Gagal membuat request spanduk: $bannerErr');
          }
        }
      }

      // 3. Prepare transaction object for DigitalReceiptPage / VisitSuccessPage
      final transaction = {
        'receipt_no': receiptNo,
        'store': {'name': widget.store.name, 'address': widget.store.address},
        'company': {'name': _companyName},
        'total_amount': totalAmountValue,
        'transaction_date': now.toUtc().toIso8601String(),
        'status': _isTaskOrder
            ? 'BELUM LUNAS'
            : (isPhysicalReceipt
                  ? 'LUNAS'
                  : (_selectedPaymentMethod == 'TEMPO'
                        ? 'TEMPO'
                        : (_selectedPaymentMethod == 'SALES_ORDER'
                              ? 'BELUM LUNAS'
                              : 'LUNAS'))),
        'payment_method': _isTaskOrder
            ? 'SALES_ORDER'
            : (isPhysicalReceipt ? 'CASH' : _selectedPaymentMethod),
        'bank': _selectedBank,
        'items': widget.items,
        'return_items': _selectedReturns,
        'signature_image_url': combinedSignaturePath,
        'receipt_path': receiptPathValue,
        'notes': widget.notes,
      };

      // 4. Trigger Sync in background
      di.sl<SyncBloc>().add(const SyncDataRequested(force: true));

      bool isAwaitingDigitalPayment = false;

      // 5. Generate Invoice PDF early so we can pass it to the payment info or success page
      Uint8List? invoicePdfBytes;
      if (widget.items != null && widget.items!.isNotEmpty && !_isTaskOrder) {
        try {
          invoicePdfBytes = await _generateInvoicePdf(
            courierSigBytes,
            customerSigBytes,
            receiptNo,
          );
        } catch (pdfErr) {
          print('Gagal generate PDF invoice: $pdfErr');
        }
      }

      if (isAwaitingDigitalPayment) {
        String displayMethod = _selectedPaymentMethod;
        showDialog(
          context: context,
          barrierDismissible: false,
          builder: (context) => AlertDialog(
            content: Row(
              children: [
                const CircularProgressIndicator(),
                const SizedBox(width: 16),
                Expanded(
                  child: Text(
                    'Menyiapkan $displayMethod ...',
                    style: GoogleFonts.outfit(),
                  ),
                ),
              ],
            ),
          ),
        );

        bool foundPayment = false;
        dynamic paymentData;
        int maxRetries =
            5; // Reduced from 20 to 5 (10s max) for a faster, smoother transition
        for (int i = 0; i < maxRetries; i++) {
          await Future.delayed(const Duration(seconds: 2));
          final checkTx = await (db.select(
            db.localTransactions,
          )..where((t) => t.localId.equals(localId))).getSingleOrNull();
          if (checkTx != null &&
              (checkTx.midtransQrisUrl != null ||
                  checkTx.midtransVaNumber != null ||
                  checkTx.midtransBillKey != null)) {
            foundPayment = true;
            paymentData = {
              'payment_method': displayMethod,
              'midtrans_qris_url': checkTx.midtransQrisUrl,
              'midtrans_va_number': checkTx.midtransVaNumber,
              'midtrans_bank': checkTx.midtransBank,
              'midtrans_bill_key': checkTx.midtransBillKey,
              'midtrans_biller_code': checkTx.midtransBillerCode,
            };
            break;
          }
        }

        if (mounted) {
          Navigator.pop(context); // pop loading dialog
          final double finalAmt = (widget.totalAmount ?? 0.0);
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(
              builder: (_) => DeliveryPaymentInfoPage(
                paymentData: paymentData ?? {'payment_method': displayMethod},
                storeName: widget.store.name,
                amount: finalAmt,
                transaction: transaction,
                localId: localId,
                invoicePdfBytes: invoicePdfBytes,
              ),
            ),
          );
          return;
        }
      }

      // Invoice PDF generation has been moved up

      final String successMsg = isPhysicalReceipt
          ? '✓ Nota fisik berhasil di-upload!'
          : (_isTaskOrder
                ? (_isMtMode
                      ? '✓ Purchase Order (PO) berhasil disimpan!'
                      : '✓ Sales Order berhasil disimpan!')
                : '✓ Kunjungan & Order $_selectedPaymentMethod tersimpan!');

      if (mounted) {
        Navigator.of(context).pushReplacement(
          MaterialPageRoute(
            builder: (_) => VisitSuccessPage(
              transaction: transaction,
              localId: localId,
              invoicePdfBytes: invoicePdfBytes,
            ),
          ),
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              successMsg,
              style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
            ),
            backgroundColor: Colors.green,
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(14),
            ),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Gagal menyimpan kunjungan: $e',
              style: GoogleFonts.outfit(),
            ),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  void _showBonusSelectionModal() async {
    final stocks = await db.select(db.salesStock).get();
    final stockMap = {for (var s in stocks) s.productId: s.quantity};

    final allProducts = await db.select(db.products).get();
    final products = allProducts
        .where((p) => p.sku == 'BONUS' && (stockMap[p.id] ?? 0) > 0)
        .toList();

    if (!mounted) return;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return Container(
              height: MediaQuery.of(context).size.height * 0.7,
              decoration: const BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.only(
                  topLeft: Radius.circular(32),
                  topRight: Radius.circular(32),
                ),
              ),
              child: Column(
                children: [
                  const SizedBox(height: 12),
                  Container(
                    width: 40,
                    height: 4,
                    decoration: BoxDecoration(
                      color: Colors.grey.shade300,
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                  const SizedBox(height: 20),
                  Text(
                    'Pilih Bonus Pembelian',
                    style: GoogleFonts.outfit(
                      fontSize: 18,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Stok bonus dari bronjong',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      color: Colors.blueGrey,
                    ),
                  ),
                  const SizedBox(height: 20),
                  Expanded(
                    child: ListView.builder(
                      itemCount: products.length,
                      itemBuilder: (context, index) {
                        final p = products[index];
                        final stockQty = stockMap[p.id] ?? 0;
                        final isSelected = _selectedBonuses.any(
                          (b) => b['product_id'] == p.id,
                        );
                        final bonusItem = isSelected
                            ? _selectedBonuses.firstWhere(
                                (b) => b['product_id'] == p.id,
                              )
                            : null;
                        final currentQty = isSelected
                            ? (bonusItem!['quantity'] as int)
                            : 0;

                        return ListTile(
                          leading: Icon(
                            Icons.card_giftcard_rounded,
                            color: isSelected ? Colors.purple : Colors.grey,
                          ),
                          title: Text(
                            p.name,
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          subtitle: Text(
                            'Stok Tersedia: $stockQty',
                            style: GoogleFonts.outfit(
                              color: Colors.green,
                              fontSize: 12,
                            ),
                          ),
                          trailing: isSelected
                              ? Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    IconButton(
                                      icon: const Icon(
                                        Icons.remove_circle_outline,
                                        color: Colors.purple,
                                      ),
                                      onPressed: () {
                                        setState(() {
                                          setModalState(() {
                                            if (currentQty > 1) {
                                              bonusItem!['quantity'] =
                                                  currentQty - 1;
                                            } else {
                                              _selectedBonuses.removeWhere(
                                                (b) => b['product_id'] == p.id,
                                              );
                                            }
                                          });
                                        });
                                      },
                                    ),
                                    Text(
                                      '$currentQty',
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.bold,
                                        fontSize: 16,
                                      ),
                                    ),
                                    IconButton(
                                      icon: const Icon(
                                        Icons.add_circle_outline,
                                        color: Colors.purple,
                                      ),
                                      onPressed: currentQty < stockQty
                                          ? () {
                                              setState(() {
                                                setModalState(() {
                                                  bonusItem!['quantity'] =
                                                      currentQty + 1;
                                                });
                                              });
                                            }
                                          : null,
                                    ),
                                  ],
                                )
                              : IconButton(
                                  icon: const Icon(
                                    Icons.add_circle_outline,
                                    color: Colors.grey,
                                  ),
                                  onPressed: stockQty > 0
                                      ? () {
                                          setState(() {
                                            setModalState(() {
                                              _selectedBonuses.add({
                                                'product_id': p.id,
                                                'product_name': p.name,
                                                'quantity': 1,
                                                'unit': p.unit,
                                              });
                                            });
                                          });
                                        }
                                      : null,
                                ),
                        );
                      },
                    ),
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  void _showReturnSelectionModal() async {
    final allProducts = await db.select(db.products).get();
    final products = allProducts
        .where((p) => p.category?.toLowerCase() == 'barang sampingan')
        .toList();

    if (!mounted) return;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return Container(
              height: MediaQuery.of(context).size.height * 0.7,
              decoration: const BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.only(
                  topLeft: Radius.circular(32),
                  topRight: Radius.circular(32),
                ),
              ),
              child: Column(
                children: [
                  const SizedBox(height: 12),
                  Container(
                    width: 40,
                    height: 4,
                    decoration: BoxDecoration(
                      color: Colors.grey.shade300,
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                  const SizedBox(height: 20),
                  Text(
                    'Pilih Retur Botol / Sampingan',
                    style: GoogleFonts.outfit(
                      fontSize: 18,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Barang akan memotong total nota',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      color: Colors.blueGrey,
                    ),
                  ),
                  const SizedBox(height: 20),
                  Expanded(
                    child: ListView.builder(
                      itemCount: products.length,
                      itemBuilder: (context, index) {
                        final p = products[index];
                        final isSelected = _selectedReturns.any(
                          (r) => r['product_id'] == p.id,
                        );
                        final returnItem = isSelected
                            ? _selectedReturns.firstWhere(
                                (r) => r['product_id'] == p.id,
                              )
                            : null;
                        final currentQty = isSelected
                            ? (returnItem!['quantity'] as int)
                            : 0;
                        final price = p.sellingPrice ?? 0;
                        final currency = NumberFormat.currency(
                          locale: 'id',
                          symbol: 'Rp ',
                          decimalDigits: 0,
                        );

                        return ListTile(
                          leading: Icon(
                            Icons.recycling_rounded,
                            color: isSelected ? Colors.orange : Colors.grey,
                          ),
                          title: Text(
                            p.name,
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          subtitle: Text(
                            'Harga Retur: ${currency.format(price)}',
                            style: GoogleFonts.outfit(
                              color: Colors.orange.shade700,
                              fontSize: 12,
                            ),
                          ),
                          trailing: isSelected
                              ? Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    IconButton(
                                      icon: const Icon(
                                        Icons.remove_circle_outline,
                                        color: Colors.orange,
                                      ),
                                      onPressed: () {
                                        setState(() {
                                          setModalState(() {
                                            if (currentQty > 1) {
                                              returnItem!['quantity'] =
                                                  currentQty - 1;
                                            } else {
                                              _selectedReturns.removeWhere(
                                                (r) => r['product_id'] == p.id,
                                              );
                                            }
                                          });
                                        });
                                      },
                                    ),
                                    Text(
                                      '$currentQty',
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.bold,
                                        fontSize: 16,
                                      ),
                                    ),
                                    IconButton(
                                      icon: const Icon(
                                        Icons.add_circle_outline,
                                        color: Colors.orange,
                                      ),
                                      onPressed: () {
                                        setState(() {
                                          setModalState(() {
                                            returnItem!['quantity'] =
                                                currentQty + 1;
                                          });
                                        });
                                      },
                                    ),
                                  ],
                                )
                              : IconButton(
                                  icon: const Icon(
                                    Icons.add_circle_outline,
                                    color: Colors.grey,
                                  ),
                                  onPressed: () {
                                    setState(() {
                                      setModalState(() {
                                        _selectedReturns.add({
                                          'product_id': p.id,
                                          'product_name': p.name,
                                          'quantity': 1,
                                          'price': price,
                                          'unit': p.unit,
                                        });
                                      });
                                    });
                                  },
                                ),
                        );
                      },
                    ),
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final currency = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

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
              'Ringkasan & Pembayaran',
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
            color: Colors.green,
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
                                color: Colors.green.withOpacity(0.1),
                                shape: BoxShape.circle,
                              ),
                              child: const Icon(
                                Icons.check_circle_rounded,
                                color: Colors.green,
                                size: 28,
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'Siap Checkout!',
                                    style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w900,
                                      fontSize: 18,
                                      color: Colors.green,
                                    ),
                                  ),
                                  Text(
                                    widget.items == null &&
                                            widget.receiptPath != null
                                        ? 'Nominal: ${currency.format(double.tryParse(_manualTotalAmountCtrl.text.trim().replaceAll(RegExp(r'[^0-9]'), '')) ?? 0)}'
                                        : 'Total Tagihan: ${currency.format(_finalTotalAmount)}',
                                    style: GoogleFonts.outfit(
                                      fontSize: 13,
                                      color: Colors.blueGrey,
                                      fontWeight: FontWeight.w600,
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
                        _row(
                          Icons.storefront_rounded,
                          'Pelanggan',
                          widget.store.name,
                        ),
                        const SizedBox(height: 10),
                        _row(Icons.business_rounded, 'PT', _companyName),
                        const SizedBox(height: 10),
                        if (widget.items != null &&
                            widget.items!.isNotEmpty) ...[
                          const SizedBox(height: 10),
                          Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Icon(
                                Icons.shopping_basket_rounded,
                                color: Colors.blueAccent,
                                size: 18,
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'Item Pesanan (${widget.items!.length})',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        color: Colors.blueGrey,
                                        fontWeight: FontWeight.w700,
                                      ),
                                    ),
                                    const SizedBox(height: 4),
                                    ...widget.items!.map((item) {
                                      final disc =
                                          (item['discount_percent'] as num? ??
                                                  0.0)
                                              .toDouble();
                                      final price =
                                          (item['price'] as num? ?? 0.0)
                                              .toDouble();
                                      final effectivePrice =
                                          price * (1 - disc / 100);
                                      final qty = item['quantity'] as int? ?? 0;
                                      return Padding(
                                        padding: const EdgeInsets.only(
                                          bottom: 6,
                                        ),
                                        child: Column(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            Text(
                                              '• ${item['product_name']} x ${item['breakdown'] ?? ('${item['quantity']} PCS')}',
                                              style: GoogleFonts.outfit(
                                                fontSize: 13,
                                                fontWeight: FontWeight.w700,
                                                color: const Color(0xFF1E293B),
                                              ),
                                            ),
                                            if (item['price_tier'] != null &&
                                                item['price_tier'] != 'NORMAL')
                                              Padding(
                                                padding: const EdgeInsets.only(
                                                  left: 12,
                                                  top: 2,
                                                ),
                                                child: Row(
                                                  children: [
                                                    Container(
                                                      padding:
                                                          const EdgeInsets.symmetric(
                                                            horizontal: 6,
                                                            vertical: 1,
                                                          ),
                                                      decoration: BoxDecoration(
                                                        color: Colors
                                                            .purple
                                                            .shade50,
                                                        borderRadius:
                                                            BorderRadius.circular(
                                                              4,
                                                            ),
                                                        border: Border.all(
                                                          color: Colors
                                                              .purple
                                                              .shade200,
                                                          width: 0.5,
                                                        ),
                                                      ),
                                                      child: Text(
                                                        item['price_tier'] ==
                                                                'DBP'
                                                            ? 'Harga Distributor (DBP)'
                                                            : (item['price_tier'] ==
                                                                      'WBP'
                                                                  ? 'Harga Grosir (WBP)'
                                                                  : 'Harga Custom Booster'),
                                                        style:
                                                            GoogleFonts.outfit(
                                                              fontSize: 10,
                                                              fontWeight:
                                                                  FontWeight
                                                                      .w800,
                                                              color: Colors
                                                                  .purple
                                                                  .shade800,
                                                            ),
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ),
                                            if (disc > 0)
                                              Padding(
                                                padding: const EdgeInsets.only(
                                                  left: 12,
                                                  top: 2,
                                                ),
                                                child: Row(
                                                  children: [
                                                    Container(
                                                      padding:
                                                          const EdgeInsets.symmetric(
                                                            horizontal: 6,
                                                            vertical: 1,
                                                          ),
                                                      decoration: BoxDecoration(
                                                        color: const Color(
                                                          0xFFEF4444,
                                                        ).withOpacity(0.1),
                                                        borderRadius:
                                                            BorderRadius.circular(
                                                              4,
                                                            ),
                                                      ),
                                                      child: Text(
                                                        'Diskon ${disc.toStringAsFixed(disc % 1 == 0 ? 0 : 1)}%',
                                                        style:
                                                            GoogleFonts.outfit(
                                                              fontSize: 10,
                                                              fontWeight:
                                                                  FontWeight
                                                                      .w800,
                                                              color:
                                                                  const Color(
                                                                    0xFFEF4444,
                                                                  ),
                                                            ),
                                                      ),
                                                    ),
                                                    const SizedBox(width: 6),
                                                    Text(
                                                      NumberFormat.currency(
                                                        locale: 'id',
                                                        symbol: 'Rp ',
                                                        decimalDigits: 0,
                                                      ).format(
                                                        effectivePrice * qty,
                                                      ),
                                                      style: GoogleFonts.outfit(
                                                        fontSize: 11,
                                                        fontWeight:
                                                            FontWeight.w700,
                                                        color: const Color(
                                                          0xFF10B981,
                                                        ),
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ),
                                          ],
                                        ),
                                      );
                                    }),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ],
                        if (_selectedBonuses.isNotEmpty) ...[
                          const SizedBox(height: 10),
                          Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Icon(
                                Icons.card_giftcard_rounded,
                                color: Colors.purple,
                                size: 18,
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'Bonus Pembelian (${_selectedBonuses.length})',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        color: Colors.purple,
                                        fontWeight: FontWeight.w700,
                                      ),
                                    ),
                                    const SizedBox(height: 4),
                                    ..._selectedBonuses.map((bonus) {
                                      return Padding(
                                        padding: const EdgeInsets.only(
                                          bottom: 6,
                                        ),
                                        child: Text(
                                          '• ${bonus['product_name']} x ${bonus['quantity']} ${bonus['unit']}',
                                          style: GoogleFonts.outfit(
                                            fontSize: 13,
                                            fontWeight: FontWeight.w700,
                                            color: const Color(0xFF1E293B),
                                          ),
                                        ),
                                      );
                                    }),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ],
                        if (_selectedReturns.isNotEmpty) ...[
                          const SizedBox(height: 10),
                          Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Icon(
                                Icons.recycling_rounded,
                                color: Colors.orange,
                                size: 18,
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'Retur Sampingan (${_selectedReturns.length})',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        color: Colors.orange.shade700,
                                        fontWeight: FontWeight.w700,
                                      ),
                                    ),
                                    const SizedBox(height: 4),
                                    ..._selectedReturns.map((retur) {
                                      final qty = retur['quantity'] as int;
                                      final price = retur['price'] as num;
                                      return Padding(
                                        padding: const EdgeInsets.only(
                                          bottom: 6,
                                        ),
                                        child: Row(
                                          mainAxisAlignment:
                                              MainAxisAlignment.spaceBetween,
                                          children: [
                                            Expanded(
                                              child: Text(
                                                '• ${retur['product_name']} x $qty ${retur['unit']}',
                                                style: GoogleFonts.outfit(
                                                  fontSize: 13,
                                                  fontWeight: FontWeight.w700,
                                                  color: const Color(
                                                    0xFF1E293B,
                                                  ),
                                                ),
                                              ),
                                            ),
                                            Text(
                                              '- ${currency.format(qty * price)}',
                                              style: GoogleFonts.outfit(
                                                fontSize: 13,
                                                fontWeight: FontWeight.w700,
                                                color: Colors.orange.shade700,
                                              ),
                                            ),
                                          ],
                                        ),
                                      );
                                    }),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ],
                        if (widget.items != null &&
                            widget.items!.isNotEmpty) ...[
                          const SizedBox(height: 16),
                          Center(
                            child: Column(
                              children: [
                                ElevatedButton.icon(
                                  onPressed: _showBonusSelectionModal,
                                  icon: const Icon(
                                    Icons.add_circle_outline_rounded,
                                    color: Colors.purple,
                                    size: 18,
                                  ),
                                  label: Text(
                                    'Tambah Bonus Pembelian',
                                    style: GoogleFonts.outfit(
                                      color: Colors.purple,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: Colors.purple.shade50,
                                    elevation: 0,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(12),
                                    ),
                                  ),
                                ),
                                const SizedBox(height: 8),
                                ElevatedButton.icon(
                                  onPressed: _showReturnSelectionModal,
                                  icon: const Icon(
                                    Icons.recycling_rounded,
                                    color: Colors.orange,
                                    size: 18,
                                  ),
                                  label: Text(
                                    'Tambah Retur Barang',
                                    style: GoogleFonts.outfit(
                                      color: Colors.orange.shade700,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: Colors.orange.shade50,
                                    elevation: 0,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(12),
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                        if (widget.items == null || widget.items!.isEmpty) ...[
                          if (widget.receiptPath != null) ...[
                            const SizedBox(height: 10),
                            _row(
                              Icons.shopping_basket_rounded,
                              'Transaksi',
                              'Upload Nota Fisik (Foto)',
                            ),
                            const SizedBox(height: 16),
                            Text(
                              'NOMINAL TRANSAKSI',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w900,
                                color: Colors.blueGrey,
                                letterSpacing: 1.5,
                              ),
                            ),
                            const SizedBox(height: 8),
                            TextFormField(
                              controller: _manualTotalAmountCtrl,
                              keyboardType: TextInputType.number,
                              decoration: InputDecoration(
                                prefixText: 'Rp ',
                                hintText: '0',
                                filled: true,
                                fillColor: const Color(0xFFF8FAFC),
                                contentPadding: const EdgeInsets.symmetric(
                                  horizontal: 16,
                                  vertical: 12,
                                ),
                                border: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  borderSide: BorderSide(
                                    color: Colors.grey.shade200,
                                  ),
                                ),
                                enabledBorder: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  borderSide: BorderSide(
                                    color: Colors.grey.shade200,
                                  ),
                                ),
                              ),
                              onChanged: (val) {
                                setState(
                                  () {},
                                ); // trigger rebuild to update summary
                              },
                            ),
                          ] else ...[
                            const SizedBox(height: 10),
                            _row(
                              Icons.shopping_basket_rounded,
                              'Item Pesanan',
                              widget.notes == 'Input Transaksi Nanti'
                                  ? 'INPUT TRANSAKSI NANTI'
                                  : 'TIDAK ADA TRANSAKSI',
                            ),
                          ],
                        ],
                        if (widget.notes != null &&
                            widget.notes!.isNotEmpty) ...[
                          const SizedBox(height: 10),
                          _row(
                            Icons.speaker_notes_rounded,
                            'Alasan / Catatan',
                            widget.notes!,
                            valueColor: Colors.orange.shade800,
                          ),
                        ],
                        if ((_selfieBytes != null ||
                                (_selfiePath != null &&
                                    _selfiePath!.isNotEmpty)) ||
                            _receiptPath != null) ...[
                          const SizedBox(height: 16),
                          const Divider(),
                          const SizedBox(height: 16),
                          Row(
                            children: [
                              if (_selfieBytes != null ||
                                  (_selfiePath != null &&
                                      _selfiePath!.isNotEmpty))
                                Expanded(
                                  child: _buildPhotoThumbnail(
                                    label: 'Selfie Check-In',
                                    path: _selfiePath ?? '',
                                    bytes: _selfieBytes,
                                    onRetake: _retakeSelfiePhoto,
                                  ),
                                ),
                              if ((_selfieBytes != null ||
                                      (_selfiePath != null &&
                                          _selfiePath!.isNotEmpty)) &&
                                  _receiptPath != null)
                                const SizedBox(width: 12),
                              if (_receiptPath != null)
                                Expanded(
                                  child: _buildPhotoThumbnail(
                                    label: 'Foto Nota Fisik',
                                    path: _receiptPath!,
                                    bytes: null,
                                    onRetake: _retakeReceiptPhoto,
                                  ),
                                ),
                            ],
                          ),
                        ],
                      ],
                    ),
                  ),

                  const SizedBox(height: 24),
                  // Khusus Sales Distributor: Kartu Pilihan Mode Pemenuhan (Direct Faktur Motoris vs PO Gudang / SO)
                  if (_isSalesDistributor &&
                      widget.items != null &&
                      widget.items!.isNotEmpty)
                    Container(
                      margin: const EdgeInsets.only(bottom: 24),
                      padding: const EdgeInsets.all(20),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(24),
                        border: Border.all(
                          color: Colors.indigo.shade100,
                          width: 1.5,
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.indigo.withOpacity(0.06),
                            blurRadius: 16,
                            offset: const Offset(0, 6),
                          ),
                        ],
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Container(
                                padding: const EdgeInsets.all(8),
                                decoration: BoxDecoration(
                                  color: Colors.indigo.shade50,
                                  borderRadius: BorderRadius.circular(10),
                                ),
                                child: Icon(
                                  Icons.swap_horiz_rounded,
                                  color: Colors.indigo.shade700,
                                  size: 20,
                                ),
                              ),
                              const SizedBox(width: 10),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'TIPE ORDER DISTRIBUTOR',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w900,
                                        color: Colors.indigo.shade900,
                                        letterSpacing: 1.2,
                                      ),
                                    ),
                                    Text(
                                      'Tentukan pemenuhan langsung motoris atau kirim via SO gudang',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        color: Colors.blueGrey,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 16),
                          Row(
                            children: [
                              Expanded(
                                child: GestureDetector(
                                  onTap: () {
                                    setState(() {
                                      _distributorFulfillmentMode = 'DIRECT';
                                      _selectedPaymentMethod = 'CASH';
                                    });
                                  },
                                  child: Container(
                                    padding: const EdgeInsets.all(14),
                                    decoration: BoxDecoration(
                                      color:
                                          _distributorFulfillmentMode ==
                                              'DIRECT'
                                          ? Colors.green.shade50
                                          : const Color(0xFFF8FAFC),
                                      borderRadius: BorderRadius.circular(16),
                                      border: Border.all(
                                        color:
                                            _distributorFulfillmentMode ==
                                                'DIRECT'
                                            ? AppColors.primaryGreen
                                            : Colors.grey.shade200,
                                        width:
                                            _distributorFulfillmentMode ==
                                                'DIRECT'
                                            ? 2
                                            : 1,
                                      ),
                                    ),
                                    child: Column(
                                      children: [
                                        Icon(
                                          Icons.two_wheeler_rounded,
                                          color:
                                              _distributorFulfillmentMode ==
                                                  'DIRECT'
                                              ? AppColors.primaryGreen
                                              : Colors.grey.shade600,
                                          size: 24,
                                        ),
                                        const SizedBox(height: 6),
                                        Text(
                                          'Faktur Langsung',
                                          style: GoogleFonts.outfit(
                                            fontSize: 13,
                                            fontWeight: FontWeight.w800,
                                            color:
                                                _distributorFulfillmentMode ==
                                                    'DIRECT'
                                                ? AppColors.primaryGreen
                                                : Colors.black87,
                                          ),
                                        ),
                                        const SizedBox(height: 2),
                                        Text(
                                          'Muatan Motoris',
                                          style: GoogleFonts.outfit(
                                            fontSize: 10,
                                            color: Colors.blueGrey,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: GestureDetector(
                                  onTap: () {
                                    setState(() {
                                      _distributorFulfillmentMode = 'SO';
                                      _selectedPaymentMethod = 'SALES_ORDER';
                                    });
                                  },
                                  child: Container(
                                    padding: const EdgeInsets.all(14),
                                    decoration: BoxDecoration(
                                      color: _distributorFulfillmentMode == 'SO'
                                          ? Colors.blue.shade50
                                          : const Color(0xFFF8FAFC),
                                      borderRadius: BorderRadius.circular(16),
                                      border: Border.all(
                                        color:
                                            _distributorFulfillmentMode == 'SO'
                                            ? const Color(0xFF0284C7)
                                            : Colors.grey.shade200,
                                        width:
                                            _distributorFulfillmentMode == 'SO'
                                            ? 2
                                            : 1,
                                      ),
                                    ),
                                    child: Column(
                                      children: [
                                        Icon(
                                          Icons.local_shipping_rounded,
                                          color:
                                              _distributorFulfillmentMode ==
                                                  'SO'
                                              ? const Color(0xFF0284C7)
                                              : Colors.grey.shade600,
                                          size: 24,
                                        ),
                                        const SizedBox(height: 6),
                                        Text(
                                          'Pesanan PO / SO',
                                          style: GoogleFonts.outfit(
                                            fontSize: 13,
                                            fontWeight: FontWeight.w800,
                                            color:
                                                _distributorFulfillmentMode ==
                                                    'SO'
                                                ? const Color(0xFF0284C7)
                                                : Colors.black87,
                                          ),
                                        ),
                                        const SizedBox(height: 2),
                                        Text(
                                          'Kirim Gudang/Truk',
                                          style: GoogleFonts.outfit(
                                            fontSize: 10,
                                            color: Colors.blueGrey,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),

                  // Payment Method selection (Only if there are items and NOT a Task Order)
                  if (widget.items != null &&
                      widget.items!.isNotEmpty &&
                      !_isTaskOrder)
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
                          ..._paymentOptions.entries
                              .where((e) => e.key != 'SALES_ORDER')
                              .map((e) {
                                final isSelected =
                                    _selectedPaymentMethod == e.key;
                                return GestureDetector(
                                  onTap: () => setState(() {
                                    _selectedPaymentMethod = e.key;
                                    if (e.key != 'VA') _selectedBank = null;
                                    if (e.key == 'TEMPO' &&
                                        _paymentDueDate == null) {
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
                          if (_selectedPaymentMethod == 'VA') ...[
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
                                                  color: const Color(
                                                    0xFF1E293B,
                                                  ),
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
                                                (b) =>
                                                    b['id']?.toString() == val,
                                                orElse: () =>
                                                    _bankAccounts.first,
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
                                                  color:
                                                      Colors.blueGrey.shade700,
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
                            Row(
                              children: [
                                Text(
                                  'BUKTI TRANSFER BANK',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    fontWeight: FontWeight.w900,
                                    color: Colors.blueGrey,
                                    letterSpacing: 1.5,
                                  ),
                                ),
                                const SizedBox(width: 6),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 6,
                                    vertical: 2,
                                  ),
                                  decoration: BoxDecoration(
                                    color: Colors.blue.shade50,
                                    borderRadius: BorderRadius.circular(6),
                                    border: Border.all(
                                      color: Colors.blue.shade200,
                                    ),
                                  ),
                                  child: Text(
                                    'OPSIONAL',
                                    style: GoogleFonts.outfit(
                                      fontSize: 9,
                                      fontWeight: FontWeight.w800,
                                      color: Colors.blue.shade700,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 6),
                            Text(
                              'Jika pelanggan baru transfer nanti sore, Anda dapat langsung menyelesaikan kunjungan ini dan mengunggah bukti transfer nanti melalui Riwayat Penjualan / Detail Faktur.',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                color: Colors.blueGrey.shade600,
                                height: 1.3,
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
                                              borderRadius:
                                                  BorderRadius.circular(18),
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
                                                          (
                                                            context,
                                                            url,
                                                            error,
                                                          ) => const Icon(
                                                            Icons.error,
                                                          ),
                                                      memCacheWidth: 800,
                                                    )
                                                  : Image.file(
                                                      File(
                                                        _transferProofImage!
                                                            .path,
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
                                                  () => _transferProofImage =
                                                      null,
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
                                            'Upload Bukti Transfer (Jika Ada)',
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
                          if (_selectedPaymentMethod == 'TEMPO') ...[
                            const SizedBox(height: 16),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'PILIHAN TENOR / JATUH TEMPO',
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
                                      border: Border.all(
                                        color: Colors.amber.shade200,
                                      ),
                                    ),
                                    child: Text(
                                      '$_selectedTempoDays Hari',
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
                                            _paymentDueDate = DateTime.now()
                                                .add(Duration(days: days));
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
                                            borderRadius: BorderRadius.circular(
                                              10,
                                            ),
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
                                            DateTime.now().add(
                                              const Duration(days: 14),
                                            ),
                                        firstDate: DateTime.now(),
                                        lastDate: DateTime.now().add(
                                          const Duration(days: 180),
                                        ),
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
                            InkWell(
                              onTap: () async {
                                final picked = await showDatePicker(
                                  context: context,
                                  initialDate:
                                      _paymentDueDate ??
                                      DateTime.now().add(
                                        const Duration(days: 14),
                                      ),
                                  firstDate: DateTime.now(),
                                  lastDate: DateTime.now().add(
                                    const Duration(days: 180),
                                  ),
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
                                  horizontal: 16,
                                  vertical: 12,
                                ),
                                decoration: BoxDecoration(
                                  color: const Color(0xFFF8FAFC),
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(
                                    color: Colors.grey.shade200,
                                  ),
                                ),
                                child: Row(
                                  children: [
                                    const Icon(
                                      Icons.calendar_month_rounded,
                                      color: Colors.blueAccent,
                                      size: 20,
                                    ),
                                    const SizedBox(width: 12),
                                    Expanded(
                                      child: Text(
                                        _paymentDueDate == null
                                            ? 'Pilih Tanggal Jatuh Tempo...'
                                            : DateFormat(
                                                'dd MMMM yyyy',
                                              ).format(_paymentDueDate!),
                                        style: GoogleFonts.outfit(
                                          fontWeight: FontWeight.w700,
                                          color: _paymentDueDate == null
                                              ? Colors.grey
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
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(height: 16),
                            Text(
                              'DP AWAL (BILA ADA)',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w900,
                                color: Colors.blueGrey,
                                letterSpacing: 1.5,
                              ),
                            ),
                            const SizedBox(height: 8),
                            TextFormField(
                              keyboardType: TextInputType.number,
                              decoration: InputDecoration(
                                prefixText: 'Rp ',
                                hintText: '0',
                                filled: true,
                                fillColor: const Color(0xFFF8FAFC),
                                contentPadding: const EdgeInsets.symmetric(
                                  horizontal: 16,
                                  vertical: 12,
                                ),
                                border: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  borderSide: BorderSide(
                                    color: Colors.grey.shade200,
                                  ),
                                ),
                                enabledBorder: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  borderSide: BorderSide(
                                    color: Colors.grey.shade200,
                                  ),
                                ),
                              ),
                              onChanged: (val) {
                                setState(() {
                                  _downPaymentAmount =
                                      double.tryParse(
                                        val.replaceAll(RegExp(r'[^0-9]'), ''),
                                      ) ??
                                      0.0;
                                });
                              },
                            ),
                            const SizedBox(height: 16),
                            Text(
                              'JUMLAH CICILAN',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w900,
                                color: Colors.blueGrey,
                                letterSpacing: 1.5,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 16,
                              ),
                              decoration: BoxDecoration(
                                color: const Color(0xFFF8FAFC),
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: Colors.grey.shade200),
                              ),
                              child: DropdownButtonHideUnderline(
                                child: DropdownButton<int>(
                                  value: _installmentCount,
                                  isExpanded: true,
                                  items: [1, 2, 3, 4, 6, 12]
                                      .map(
                                        (e) => DropdownMenuItem(
                                          value: e,
                                          child: Text(
                                            e == 1
                                                ? 'Bayar Penuh Sekaligus (1x)'
                                                : '$e Kali Cicilan',
                                            style: GoogleFonts.outfit(
                                              fontWeight: FontWeight.w700,
                                              color: const Color(0xFF1E293B),
                                            ),
                                          ),
                                        ),
                                      )
                                      .toList(),
                                  onChanged: (val) {
                                    if (val != null)
                                      setState(() => _installmentCount = val);
                                  },
                                ),
                              ),
                            ),
                          ],
                        ],
                      ),
                    )
                  else if (_isTaskOrder &&
                      widget.items != null &&
                      widget.items!.isNotEmpty)
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
                            'STATUS PESANAN',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
                              color: Colors.blueGrey,
                              letterSpacing: 1.5,
                            ),
                          ),
                          const SizedBox(height: 16),
                          Container(
                            padding: const EdgeInsets.all(16),
                            decoration: BoxDecoration(
                              color: Colors.blueAccent.withOpacity(0.05),
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(
                                color: Colors.blueAccent,
                                width: 2,
                              ),
                            ),
                            child: Row(
                              children: [
                                const Icon(
                                  Icons.description_rounded,
                                  color: Colors.blueAccent,
                                  size: 24,
                                ),
                                const SizedBox(width: 16),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        _isMtMode
                                            ? 'PURCHASE ORDER (PO)'
                                            : 'SALES ORDER (SO)',
                                        style: GoogleFonts.outfit(
                                          fontWeight: FontWeight.w900,
                                          fontSize: 15,
                                          color: const Color(0xFF1E293B),
                                        ),
                                      ),
                                      Text(
                                        _isMtMode
                                            ? 'Akan diverifikasi admin nota pusat.'
                                            : 'Akan diverifikasi admin nota untuk pembuatan invoice.',
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          color: Colors.blueGrey,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                const Icon(
                                  Icons.check_circle_rounded,
                                  color: Colors.blueAccent,
                                  size: 20,
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),

                  const SizedBox(height: 24),

                  // Signature pads for direct sales (Cash, QRIS, VA, Tempo)
                  if (widget.items != null &&
                      widget.items!.isNotEmpty &&
                      !_isTaskOrder) ...[
                    const SizedBox(height: 24),
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
                            'TANDA TANGAN KONFIRMASI',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
                              color: Colors.blueGrey,
                              letterSpacing: 1.5,
                            ),
                          ),
                          const SizedBox(height: 16),
                          SignatureCanvas(
                            label: 'TANDA TANGAN SALESMAN / MOTORIS',
                            points: _courierSignaturePoints,
                            onDraw: (pts) =>
                                setState(() => _courierSignaturePoints = pts),
                            onClear: () =>
                                setState(() => _courierSignaturePoints.clear()),
                          ),
                          const SizedBox(height: 20),
                          SignatureCanvas(
                            label: 'TANDA TANGAN CUSTOMER / PENERIMA BARANG',
                            points: _customerSignaturePoints,
                            onDraw: (pts) =>
                                setState(() => _customerSignaturePoints = pts),
                            onClear: () => setState(
                              () => _customerSignaturePoints.clear(),
                            ),
                          ),
                          const SizedBox(height: 24),
                          const Divider(height: 1, color: Color(0xFFE2E8F0)),
                          const SizedBox(height: 16),
                          SwitchListTile(
                            value: _wantsBanner,
                            onChanged: (val) =>
                                setState(() => _wantsBanner = val),
                            title: Text(
                              'Minta Pembuatan Spanduk / Banner',
                              style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w900,
                                fontSize: 14,
                                color: const Color(0xFF1E293B),
                              ),
                            ),
                            subtitle: Text(
                              'Aktifkan jika toko meminta spanduk promosi',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                color: Colors.blueGrey,
                              ),
                            ),
                            secondary: Container(
                              padding: const EdgeInsets.all(8),
                              decoration: BoxDecoration(
                                color: Colors.blueAccent.withOpacity(0.1),
                                borderRadius: BorderRadius.circular(10),
                              ),
                              child: const Icon(
                                Icons.art_track_rounded,
                                color: Colors.blueAccent,
                                size: 20,
                              ),
                            ),
                            activeThumbColor: Colors.blueAccent,
                            contentPadding: EdgeInsets.zero,
                          ),
                          if (_wantsBanner) ...[
                            const SizedBox(height: 16),
                            Container(
                              padding: const EdgeInsets.all(16),
                              decoration: BoxDecoration(
                                color: const Color(0xFFF8FAFC),
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(
                                  color: const Color(0xFFE2E8F0),
                                ),
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'DETAIL REQUEST SPANDUK',
                                    style: GoogleFonts.outfit(
                                      fontSize: 10,
                                      fontWeight: FontWeight.w900,
                                      color: Colors.blueGrey,
                                      letterSpacing: 1.2,
                                    ),
                                  ),
                                  const SizedBox(height: 12),
                                  // ── Tipe Spanduk (Custom vs Universal) ──
                                  Text(
                                    'Tipe Spanduk',
                                    style: GoogleFonts.outfit(
                                      fontSize: 12,
                                      fontWeight: FontWeight.w700,
                                      color: const Color(0xFF334155),
                                    ),
                                  ),
                                  const SizedBox(height: 6),
                                  Row(
                                    children: [
                                      Expanded(
                                        child: InkWell(
                                          onTap: () => setState(
                                            () => _bannerType = 'Custom',
                                          ),
                                          borderRadius: BorderRadius.circular(
                                            12,
                                          ),
                                          child: Container(
                                            padding: const EdgeInsets.symmetric(
                                              vertical: 10,
                                            ),
                                            decoration: BoxDecoration(
                                              color: _bannerType == 'Custom'
                                                  ? AppColors.primaryGreen
                                                        .withOpacity(0.12)
                                                  : Colors.white,
                                              borderRadius:
                                                  BorderRadius.circular(12),
                                              border: Border.all(
                                                color: _bannerType == 'Custom'
                                                    ? AppColors.primaryGreen
                                                    : Colors.grey.shade300,
                                                width: _bannerType == 'Custom'
                                                    ? 1.5
                                                    : 1,
                                              ),
                                            ),
                                            child: Row(
                                              mainAxisAlignment:
                                                  MainAxisAlignment.center,
                                              children: [
                                                Icon(
                                                  _bannerType == 'Custom'
                                                      ? Icons
                                                            .radio_button_checked_rounded
                                                      : Icons
                                                            .radio_button_off_rounded,
                                                  size: 16,
                                                  color: _bannerType == 'Custom'
                                                      ? AppColors.primaryGreen
                                                      : Colors.grey.shade400,
                                                ),
                                                const SizedBox(width: 6),
                                                Text(
                                                  'Custom',
                                                  style: GoogleFonts.outfit(
                                                    fontSize: 12,
                                                    fontWeight: FontWeight.w800,
                                                    color:
                                                        _bannerType == 'Custom'
                                                        ? AppColors.primaryGreen
                                                        : const Color(
                                                            0xFF475569,
                                                          ),
                                                  ),
                                                ),
                                              ],
                                            ),
                                          ),
                                        ),
                                      ),
                                      const SizedBox(width: 8),
                                      Expanded(
                                        child: InkWell(
                                          onTap: () => setState(
                                            () => _bannerType = 'Universal',
                                          ),
                                          borderRadius: BorderRadius.circular(
                                            12,
                                          ),
                                          child: Container(
                                            padding: const EdgeInsets.symmetric(
                                              vertical: 10,
                                            ),
                                            decoration: BoxDecoration(
                                              color: _bannerType == 'Universal'
                                                  ? const Color(
                                                      0xFF0284C7,
                                                    ).withOpacity(0.12)
                                                  : Colors.white,
                                              borderRadius:
                                                  BorderRadius.circular(12),
                                              border: Border.all(
                                                color:
                                                    _bannerType == 'Universal'
                                                    ? const Color(0xFF0284C7)
                                                    : Colors.grey.shade300,
                                                width:
                                                    _bannerType == 'Universal'
                                                    ? 1.5
                                                    : 1,
                                              ),
                                            ),
                                            child: Row(
                                              mainAxisAlignment:
                                                  MainAxisAlignment.center,
                                              children: [
                                                Icon(
                                                  _bannerType == 'Universal'
                                                      ? Icons
                                                            .radio_button_checked_rounded
                                                      : Icons
                                                            .radio_button_off_rounded,
                                                  size: 16,
                                                  color:
                                                      _bannerType == 'Universal'
                                                      ? const Color(0xFF0284C7)
                                                      : Colors.grey.shade400,
                                                ),
                                                const SizedBox(width: 6),
                                                Text(
                                                  'Universal',
                                                  style: GoogleFonts.outfit(
                                                    fontSize: 12,
                                                    fontWeight: FontWeight.w800,
                                                    color:
                                                        _bannerType ==
                                                            'Universal'
                                                        ? const Color(
                                                            0xFF0284C7,
                                                          )
                                                        : const Color(
                                                            0xFF475569,
                                                          ),
                                                  ),
                                                ),
                                              ],
                                            ),
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 14),
                                  // Header Dimensi & Unit Switcher
                                  Row(
                                    mainAxisAlignment:
                                        MainAxisAlignment.spaceBetween,
                                    crossAxisAlignment:
                                        CrossAxisAlignment.center,
                                    children: [
                                      Text(
                                        'Ukuran Spanduk',
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: FontWeight.w700,
                                          color: const Color(0xFF475569),
                                        ),
                                      ),
                                      Container(
                                        padding: const EdgeInsets.all(2),
                                        decoration: BoxDecoration(
                                          color: const Color(0xFFF1F5F9),
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                          border: Border.all(
                                            color: Colors.grey.shade300,
                                          ),
                                        ),
                                        child: Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: [
                                            GestureDetector(
                                              onTap: () =>
                                                  _switchBannerUnit('m'),
                                              child: Container(
                                                padding:
                                                    const EdgeInsets.symmetric(
                                                      horizontal: 8,
                                                      vertical: 4,
                                                    ),
                                                decoration: BoxDecoration(
                                                  color: _bannerUnit == 'm'
                                                      ? AppColors.primaryGreen
                                                      : Colors.transparent,
                                                  borderRadius:
                                                      BorderRadius.circular(6),
                                                ),
                                                child: Text(
                                                  'Meter (m)',
                                                  style: GoogleFonts.outfit(
                                                    fontSize: 11,
                                                    fontWeight:
                                                        _bannerUnit == 'm'
                                                        ? FontWeight.w800
                                                        : FontWeight.w600,
                                                    color: _bannerUnit == 'm'
                                                        ? Colors.white
                                                        : const Color(
                                                            0xFF64748B,
                                                          ),
                                                  ),
                                                ),
                                              ),
                                            ),
                                            GestureDetector(
                                              onTap: () =>
                                                  _switchBannerUnit('cm'),
                                              child: Container(
                                                padding:
                                                    const EdgeInsets.symmetric(
                                                      horizontal: 8,
                                                      vertical: 4,
                                                    ),
                                                decoration: BoxDecoration(
                                                  color: _bannerUnit == 'cm'
                                                      ? AppColors.primaryGreen
                                                      : Colors.transparent,
                                                  borderRadius:
                                                      BorderRadius.circular(6),
                                                ),
                                                child: Text(
                                                  'Centimeter (cm)',
                                                  style: GoogleFonts.outfit(
                                                    fontSize: 11,
                                                    fontWeight:
                                                        _bannerUnit == 'cm'
                                                        ? FontWeight.w800
                                                        : FontWeight.w600,
                                                    color: _bannerUnit == 'cm'
                                                        ? Colors.white
                                                        : const Color(
                                                            0xFF64748B,
                                                          ),
                                                  ),
                                                ),
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 8),
                                  SingleChildScrollView(
                                    scrollDirection: Axis.horizontal,
                                    child: Row(
                                      children:
                                          [
                                            {
                                              'label': '2 × 1 m',
                                              'l': 2.0,
                                              'w': 1.0,
                                            },
                                            {
                                              'label': '3 × 1 m',
                                              'l': 3.0,
                                              'w': 1.0,
                                            },
                                            {
                                              'label': '3 × 0.8 m',
                                              'l': 3.0,
                                              'w': 0.8,
                                            },
                                            {
                                              'label': '4 × 1 m',
                                              'l': 4.0,
                                              'w': 1.0,
                                            },
                                          ].map((p) {
                                            return Padding(
                                              padding: const EdgeInsets.only(
                                                right: 6,
                                              ),
                                              child: InkWell(
                                                borderRadius:
                                                    BorderRadius.circular(8),
                                                onTap: () => _applyBannerPreset(
                                                  p['l'] as double,
                                                  p['w'] as double,
                                                ),
                                                child: Container(
                                                  padding:
                                                      const EdgeInsets.symmetric(
                                                        horizontal: 8,
                                                        vertical: 4,
                                                      ),
                                                  decoration: BoxDecoration(
                                                    color: Colors.white,
                                                    borderRadius:
                                                        BorderRadius.circular(
                                                          6,
                                                        ),
                                                    border: Border.all(
                                                      color:
                                                          Colors.grey.shade300,
                                                    ),
                                                  ),
                                                  child: Row(
                                                    mainAxisSize:
                                                        MainAxisSize.min,
                                                    children: [
                                                      const Icon(
                                                        Icons
                                                            .straighten_rounded,
                                                        size: 12,
                                                        color: AppColors
                                                            .primaryGreen,
                                                      ),
                                                      const SizedBox(width: 4),
                                                      Text(
                                                        p['label'] as String,
                                                        style:
                                                            GoogleFonts.outfit(
                                                              fontSize: 10.5,
                                                              fontWeight:
                                                                  FontWeight
                                                                      .w700,
                                                              color:
                                                                  const Color(
                                                                    0xFF334155,
                                                                  ),
                                                            ),
                                                      ),
                                                    ],
                                                  ),
                                                ),
                                              ),
                                            );
                                          }).toList(),
                                    ),
                                  ),
                                  const SizedBox(height: 10),
                                  Row(
                                    children: [
                                      Expanded(
                                        child: TextFormField(
                                          controller: _bannerLengthCtrl,
                                          keyboardType:
                                              const TextInputType.numberWithOptions(
                                                decimal: true,
                                              ),
                                          decoration: InputDecoration(
                                            prefixIcon: const Icon(
                                              Icons.width_normal_rounded,
                                              size: 20,
                                              color: Color(0xFF64748B),
                                            ),
                                            labelText: 'Panjang ($_bannerUnit)',
                                            labelStyle: GoogleFonts.outfit(
                                              fontSize: 12,
                                            ),
                                            hintText: _bannerUnit == 'm'
                                                ? 'Contoh: 3.0'
                                                : 'Contoh: 300',
                                            filled: true,
                                            fillColor: Colors.white,
                                            border: OutlineInputBorder(
                                              borderRadius:
                                                  BorderRadius.circular(12),
                                              borderSide: BorderSide(
                                                color: Colors.grey.shade200,
                                              ),
                                            ),
                                            contentPadding:
                                                const EdgeInsets.symmetric(
                                                  horizontal: 12,
                                                  vertical: 12,
                                                ),
                                          ),
                                        ),
                                      ),
                                      const SizedBox(width: 12),
                                      Expanded(
                                        child: TextFormField(
                                          controller: _bannerWidthCtrl,
                                          keyboardType:
                                              const TextInputType.numberWithOptions(
                                                decimal: true,
                                              ),
                                          decoration: InputDecoration(
                                            prefixIcon: const Icon(
                                              Icons.height_rounded,
                                              size: 20,
                                              color: Color(0xFF64748B),
                                            ),
                                            labelText: 'Tinggi ($_bannerUnit)',
                                            labelStyle: GoogleFonts.outfit(
                                              fontSize: 12,
                                            ),
                                            hintText: _bannerUnit == 'm'
                                                ? 'Contoh: 1.0'
                                                : 'Contoh: 100',
                                            filled: true,
                                            fillColor: Colors.white,
                                            border: OutlineInputBorder(
                                              borderRadius:
                                                  BorderRadius.circular(12),
                                              borderSide: BorderSide(
                                                color: Colors.grey.shade200,
                                              ),
                                            ),
                                            contentPadding:
                                                const EdgeInsets.symmetric(
                                                  horizontal: 12,
                                                  vertical: 12,
                                                ),
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                  _buildBannerDimensionHelper(),
                                  const SizedBox(height: 12),
                                  TextFormField(
                                    controller: _bannerQtyCtrl,
                                    keyboardType: TextInputType.number,
                                    decoration: InputDecoration(
                                      prefixIcon: const Icon(
                                        Icons.format_list_numbered_rounded,
                                        size: 20,
                                        color: Color(0xFF64748B),
                                      ),
                                      labelText: 'Banyaknya (Qty) *',
                                      labelStyle: GoogleFonts.outfit(
                                        fontSize: 12,
                                      ),
                                      hintText: '1',
                                      filled: true,
                                      fillColor: Colors.white,
                                      border: OutlineInputBorder(
                                        borderRadius: BorderRadius.circular(12),
                                        borderSide: BorderSide(
                                          color: Colors.grey.shade200,
                                        ),
                                      ),
                                      contentPadding:
                                          const EdgeInsets.symmetric(
                                            horizontal: 12,
                                            vertical: 12,
                                          ),
                                    ),
                                  ),
                                  const SizedBox(height: 12),
                                  TextFormField(
                                    controller: _bannerNotesCtrl,
                                    maxLines: 2,
                                    decoration: InputDecoration(
                                      labelText: 'Catatan / Keterangan Desain',
                                      labelStyle: GoogleFonts.outfit(
                                        fontSize: 12,
                                      ),
                                      hintText:
                                          'Nama Toko, HP, Bihun cap Tanpa Tandingan...',
                                      filled: true,
                                      fillColor: Colors.white,
                                      border: OutlineInputBorder(
                                        borderRadius: BorderRadius.circular(12),
                                        borderSide: BorderSide(
                                          color: Colors.grey.shade200,
                                        ),
                                      ),
                                      contentPadding:
                                          const EdgeInsets.symmetric(
                                            horizontal: 12,
                                            vertical: 12,
                                          ),
                                    ),
                                  ),
                                  const SizedBox(height: 14),
                                  _buildBannerAttachmentSection(),
                                ],
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ],

                  const SizedBox(height: 40),
                  // Checkout button
                  SizedBox(
                    width: double.infinity,
                    height: 72,
                    child: ElevatedButton(
                      onPressed: _isSubmitting ? null : _submitAndCheckout,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: widget.items != null
                            ? _paymentOptions[_selectedPaymentMethod]!['color']
                            : (widget.receiptPath != null
                                  ? Colors.indigo
                                  : Colors.green),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(24),
                        ),
                        elevation: 12,
                        shadowColor:
                            (widget.items != null
                                    ? _paymentOptions[_selectedPaymentMethod]!['color']
                                    : (widget.receiptPath != null
                                          ? Colors.indigo
                                          : Colors.green))
                                .withOpacity(0.4),
                      ),
                      child: _isSubmitting
                          ? const CircularProgressIndicator(color: Colors.white)
                          : Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Text(
                                  widget.items == null
                                      ? (widget.receiptPath != null
                                            ? 'KIRIM NOTA FISIK'
                                            : 'SELESAI KUNJUNGAN')
                                      : _isTaskOrder
                                      ? (_isMtMode
                                            ? 'SIMPAN PURCHASE ORDER (PO)'
                                            : 'SIMPAN SALES ORDER')
                                      : _selectedPaymentMethod == 'QRIS'
                                      ? 'BAYAR VIA QRIS'
                                      : 'SELESAI & CHECKOUT',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w900,
                                    fontSize: 18,
                                    color: Colors.white,
                                  ),
                                ),
                                if (widget.items != null)
                                  Text(
                                    _isTaskOrder
                                        ? (_isMtMode
                                              ? 'Purchase Order (MT)'
                                              : 'Sales Task Order (Non-Tunai)')
                                        : 'Pembayaran: ${_paymentOptions[_selectedPaymentMethod]!['label']}',
                                    style: GoogleFonts.outfit(
                                      fontSize: 11,
                                      color: Colors.white70,
                                    ),
                                  )
                                else if (widget.receiptPath != null)
                                  Text(
                                    'Nota fisik akan diverifikasi admin',
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

  void _showImagePreviewDialog(
    BuildContext context,
    String title,
    String path,
    Uint8List? bytes,
  ) {
    Uint8List? imageBytes = bytes;
    if (imageBytes == null && path.startsWith('data:')) {
      try {
        imageBytes = base64Decode(path.split(',').last);
      } catch (e) {
        print('Error decoding base64 image: $e');
      }
    }

    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: Colors.transparent,
        insetPadding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                IconButton(
                  icon: const Icon(Icons.close, color: Colors.white, size: 30),
                  onPressed: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: Container(
                constraints: const BoxConstraints(maxHeight: 500),
                width: double.infinity,
                child: imageBytes != null
                    ? Image.memory(imageBytes, fit: BoxFit.contain)
                    : (path.startsWith('http') || kIsWeb)
                    ? (path.isNotEmpty
                          ? CachedNetworkImage(
                              imageUrl: path,
                              fit: BoxFit.contain,
                              placeholder: (context, url) => const Center(
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              ),
                              errorWidget: (context, url, error) => const Icon(
                                Icons.error_outline_rounded,
                                color: Colors.red,
                              ),
                              memCacheWidth: 800,
                            )
                          : const Icon(
                              Icons.broken_image_rounded,
                              color: Colors.grey,
                            ))
                    : Image.file(
                        File(path),
                        fit: BoxFit.contain,
                        errorBuilder: (context, error, stackTrace) =>
                            const Icon(
                              Icons.broken_image_rounded,
                              color: Colors.grey,
                            ),
                      ),
              ),
            ),
            const SizedBox(height: 12),
            Text(
              title,
              style: GoogleFonts.outfit(
                color: Colors.white,
                fontWeight: FontWeight.bold,
                fontSize: 16,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBannerDimensionHelper() {
    final lenStr = _bannerLengthCtrl.text.trim().replaceAll(',', '.');
    final widStr = _bannerWidthCtrl.text.trim().replaceAll(',', '.');
    final lenVal = double.tryParse(lenStr);
    final widVal = double.tryParse(widStr);

    if (lenVal == null || lenVal <= 0 || widVal == null || widVal <= 0) {
      return const SizedBox.shrink();
    }

    final double lenCm = _bannerUnit == 'm' ? (lenVal * 100) : lenVal;
    final double widCm = _bannerUnit == 'm' ? (widVal * 100) : widVal;
    final double lenM = lenCm / 100;
    final double widM = widCm / 100;
    final double areaM2 = lenM * widM;

    final String alternateText = _bannerUnit == 'm'
        ? '≈ ${lenCm % 1 == 0 ? lenCm.toInt() : lenCm.toStringAsFixed(1)} cm × ${widCm % 1 == 0 ? widCm.toInt() : widCm.toStringAsFixed(1)} cm'
        : '≈ ${lenM.toStringAsFixed(2)} m × ${widM.toStringAsFixed(2)} m';

    return Container(
      margin: const EdgeInsets.only(top: 8),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: AppColors.primaryGreen.withOpacity(0.08),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: AppColors.primaryGreen.withOpacity(0.2)),
      ),
      child: Row(
        children: [
          const Icon(
            Icons.swap_horiz_rounded,
            color: AppColors.primaryGreen,
            size: 16,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              'Konversi: $alternateText  •  Luas: ${areaM2.toStringAsFixed(2)} m²',
              style: GoogleFonts.outfit(
                fontSize: 11.5,
                fontWeight: FontWeight.w700,
                color: const Color(0xFF0F766E),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBannerAttachmentSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'Foto / Logo / Dokumen Pendukung',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w700,
                color: const Color(0xFF475569),
                fontSize: 12,
              ),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: const Color(0xFFF1F5F9),
                borderRadius: BorderRadius.circular(6),
              ),
              child: Text(
                'Opsional',
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  fontWeight: FontWeight.w700,
                  color: Colors.blueGrey,
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        if (_bannerAttachmentBytes != null ||
            _bannerAttachmentPath != null) ...[
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                color: AppColors.primaryGreen.withOpacity(0.4),
              ),
            ),
            child: Row(
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: _isBannerPdfAttachment
                      ? Container(
                          width: 44,
                          height: 44,
                          color: const Color(0xFFFEE2E2),
                          child: const Icon(
                            Icons.picture_as_pdf_rounded,
                            color: Color(0xFFDC2626),
                            size: 24,
                          ),
                        )
                      : (_bannerAttachmentBytes != null
                            ? Image.memory(
                                _bannerAttachmentBytes!,
                                width: 44,
                                height: 44,
                                fit: BoxFit.cover,
                              )
                            : Container(
                                width: 44,
                                height: 44,
                                color: Colors.grey.shade200,
                                child: const Icon(
                                  Icons.image,
                                  size: 22,
                                  color: Colors.blueGrey,
                                ),
                              )),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _bannerAttachmentName ?? 'File terlampir',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: GoogleFonts.outfit(
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                          color: const Color(0xFF1E293B),
                        ),
                      ),
                      Text(
                        _isBannerPdfAttachment ? 'Dokumen PDF' : 'Gambar/Foto',
                        style: GoogleFonts.outfit(
                          fontSize: 11,
                          color: Colors.grey.shade600,
                        ),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  icon: const Icon(
                    Icons.delete_outline_rounded,
                    color: Colors.redAccent,
                    size: 20,
                  ),
                  onPressed: () {
                    setState(() {
                      _bannerAttachmentBytes = null;
                      _bannerAttachmentPath = null;
                      _bannerAttachmentName = null;
                      _isBannerPdfAttachment = false;
                    });
                  },
                ),
              ],
            ),
          ),
        ] else ...[
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: () => _pickBannerImage(ImageSource.camera),
                  icon: const Icon(
                    Icons.camera_alt_rounded,
                    size: 16,
                    color: AppColors.primaryGreen,
                  ),
                  label: Text(
                    'Kamera',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: const Color(0xFF334155),
                    ),
                  ),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 10),
                    side: BorderSide(color: Colors.grey.shade300),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: () => _pickBannerImage(ImageSource.gallery),
                  icon: const Icon(
                    Icons.photo_library_rounded,
                    size: 16,
                    color: AppColors.primaryGreen,
                  ),
                  label: Text(
                    'Galeri',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: const Color(0xFF334155),
                    ),
                  ),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 10),
                    side: BorderSide(color: Colors.grey.shade300),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: _pickBannerDocument,
                  icon: const Icon(
                    Icons.attach_file_rounded,
                    size: 16,
                    color: AppColors.primaryGreen,
                  ),
                  label: Text(
                    'File/PDF',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: const Color(0xFF334155),
                    ),
                  ),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 10),
                    side: BorderSide(color: Colors.grey.shade300),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ],
      ],
    );
  }

  Widget _buildPhotoThumbnail({
    required String label,
    required String path,
    required Uint8List? bytes,
    required VoidCallback onRetake,
  }) {
    Uint8List? imageBytes = bytes;
    if (imageBytes == null && path.startsWith('data:')) {
      try {
        imageBytes = base64Decode(path.split(',').last);
      } catch (e) {
        print('Error decoding base64 image: $e');
      }
    }

    return Column(
      children: [
        Stack(
          children: [
            GestureDetector(
              onTap: () => _showImagePreviewDialog(context, label, path, bytes),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(20),
                child: Container(
                  height: 140,
                  width: double.infinity,
                  color: Colors.grey.shade100,
                  child: imageBytes != null
                      ? Image.memory(imageBytes, fit: BoxFit.cover)
                      : (path.startsWith('http') || kIsWeb)
                      ? (path.isNotEmpty
                            ? CachedNetworkImage(
                                imageUrl: path,
                                fit: BoxFit.cover,
                                placeholder: (context, url) => const Center(
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                  ),
                                ),
                                errorWidget: (context, url, error) =>
                                    const Icon(
                                      Icons.error_outline_rounded,
                                      color: Colors.red,
                                    ),
                                memCacheWidth: 800,
                              )
                            : const Icon(
                                Icons.broken_image_rounded,
                                color: Colors.grey,
                              ))
                      : Image.file(
                          File(path),
                          fit: BoxFit.cover,
                          errorBuilder: (context, error, stackTrace) =>
                              const Icon(
                                Icons.broken_image_rounded,
                                color: Colors.grey,
                              ),
                        ),
                ),
              ),
            ),
            Positioned(
              top: 6,
              right: 6,
              child: GestureDetector(
                onTap: onRetake,
                child: Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: Colors.black.withOpacity(0.6),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.replay_rounded,
                    color: Colors.white,
                    size: 18,
                  ),
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Text(
          label,
          style: GoogleFonts.outfit(
            fontSize: 12,
            fontWeight: FontWeight.bold,
            color: Colors.blueGrey,
          ),
        ),
      ],
    );
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

  Future<Uint8List?> _generateInvoicePdf(
    Uint8List? courierSigBytes,
    Uint8List? customerSigBytes,
    String receiptNo,
  ) async {
    final pdf = pw.Document();
    final currencyFormat = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    final storeName = widget.store.name;
    final storeAddress = widget.store.address;
    final dateStr = DateFormat('dd MMMM yyyy').format(DateTime.now());
    final paymentMethodLabel =
        _paymentOptions[_selectedPaymentMethod]?['label'] ??
        _selectedPaymentMethod;
    final baseTotal = widget.totalAmount ?? 0.0;
    final totalAmount = _finalTotalAmount;

    final companyName = _companyName;
    const companyAddress = 'Kawasan Industri & Distribusi Sentra Prima';

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
          // Calculate totals
          double totalBeforeDiscount = 0;
          double totalDiscount = 0;
          if (widget.items != null) {
            for (var item in widget.items!) {
              final qty = item['quantity'] ?? 0;
              final price = (item['price'] ?? 0).toDouble();
              final disc = (item['discount_percent'] as num? ?? 0.0).toDouble();
              totalBeforeDiscount += qty * price;
              totalDiscount += qty * price * (disc / 100);
            }
          }

          return pw.Column(
            crossAxisAlignment: pw.CrossAxisAlignment.start,
            children: [
              // Header (Black and White)
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
                      pw.Text(
                        _selectedPaymentMethod == 'TEMPO'
                            ? 'Status: TEMPO'
                            : (_selectedPaymentMethod == 'SALES_ORDER'
                                  ? 'Status: BELUM LUNAS'
                                  : 'Status: LUNAS'),
                        style: pw.TextStyle(
                          fontSize: 7.5,
                          fontWeight: pw.FontWeight.bold,
                          color: PdfColors.black,
                        ),
                      ),
                      if (_selectedPaymentMethod == 'TEMPO' &&
                          _paymentDueDate != null)
                        pw.Text(
                          'Jatuh Tempo: ${DateFormat('dd MMMM yyyy').format(_paymentDueDate!)}',
                          style: const pw.TextStyle(fontSize: 7),
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
                        'Harga Satuan',
                        style: pw.TextStyle(
                          color: PdfColors.black,
                          fontWeight: pw.FontWeight.bold,
                          fontSize: 7,
                        ),
                        textAlign: pw.TextAlign.right,
                      ),
                    ),
                    pw.Expanded(
                      flex: 1,
                      child: pw.Text(
                        'Diskon',
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
              if (widget.items != null)
                ...widget.items!.map((item) {
                  final prodName = item['product_name'] ?? 'Produk';
                  final qty = item['quantity'] ?? 0;
                  final price = (item['price'] ?? 0).toDouble();
                  final disc = (item['discount_percent'] as num? ?? 0.0)
                      .toDouble();
                  final effectivePrice = price * (1 - disc / 100);
                  final subtotal = qty * effectivePrice;

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
                            '$qty',
                            style: const pw.TextStyle(fontSize: 7),
                            textAlign: pw.TextAlign.center,
                          ),
                        ),
                        pw.Expanded(
                          flex: 2,
                          child: pw.Column(
                            crossAxisAlignment: pw.CrossAxisAlignment.end,
                            children: [
                              if (disc > 0)
                                pw.Text(
                                  currencyFormat.format(price),
                                  style: const pw.TextStyle(
                                    fontSize: 6,
                                    color: PdfColors.grey700,
                                    decoration: pw.TextDecoration.lineThrough,
                                  ),
                                  textAlign: pw.TextAlign.right,
                                ),
                              pw.Text(
                                currencyFormat.format(effectivePrice),
                                style: pw.TextStyle(
                                  fontSize: 7,
                                  fontWeight: disc > 0
                                      ? pw.FontWeight.bold
                                      : pw.FontWeight.normal,
                                ),
                                textAlign: pw.TextAlign.right,
                              ),
                            ],
                          ),
                        ),
                        pw.Expanded(
                          flex: 1,
                          child: pw.Text(
                            disc > 0
                                ? '${disc.toStringAsFixed(disc % 1 == 0 ? 0 : 1)}%'
                                : '-',
                            style: pw.TextStyle(
                              fontSize: 7,
                              color: disc > 0
                                  ? PdfColors.red
                                  : PdfColors.grey600,
                              fontWeight: disc > 0
                                  ? pw.FontWeight.bold
                                  : pw.FontWeight.normal,
                            ),
                            textAlign: pw.TextAlign.center,
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

              if (_selectedReturns.isNotEmpty) ...[
                pw.SizedBox(height: 10),
                pw.Container(
                  color: PdfColors.red50,
                  padding: const pw.EdgeInsets.symmetric(
                    vertical: 3,
                    horizontal: 6,
                  ),
                  child: pw.Text(
                    'RETUR BARANG SAMPINGAN',
                    style: pw.TextStyle(
                      color: PdfColors.red900,
                      fontWeight: pw.FontWeight.bold,
                      fontSize: 7,
                    ),
                  ),
                ),
                ..._selectedReturns.map((item) {
                  final prodName = item['product_name'] ?? 'Produk';
                  final qty = item['quantity'] ?? 0;
                  final price = (item['price'] ?? 0).toDouble();
                  final subtotal = qty * price;

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
                            style: const pw.TextStyle(
                              fontSize: 7,
                              color: PdfColors.red800,
                            ),
                          ),
                        ),
                        pw.Expanded(
                          flex: 1,
                          child: pw.Text(
                            '$qty',
                            style: const pw.TextStyle(
                              fontSize: 7,
                              color: PdfColors.red800,
                            ),
                            textAlign: pw.TextAlign.center,
                          ),
                        ),
                        pw.Expanded(
                          flex: 2,
                          child: pw.Text(
                            currencyFormat.format(price),
                            style: const pw.TextStyle(
                              fontSize: 7,
                              color: PdfColors.red800,
                            ),
                            textAlign: pw.TextAlign.right,
                          ),
                        ),
                        pw.Expanded(
                          flex: 1,
                          child: pw.Text(
                            '-',
                            style: const pw.TextStyle(
                              fontSize: 7,
                              color: PdfColors.red800,
                            ),
                            textAlign: pw.TextAlign.center,
                          ),
                        ),
                        pw.Expanded(
                          flex: 2,
                          child: pw.Text(
                            '-${currencyFormat.format(subtotal)}',
                            style: pw.TextStyle(
                              fontSize: 7,
                              color: PdfColors.red800,
                              fontWeight: pw.FontWeight.bold,
                            ),
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
                      if (totalDiscount > 0) ...[
                        pw.Row(
                          children: [
                            pw.Text(
                              'Total Sebelum Diskon: ',
                              style: const pw.TextStyle(fontSize: 8),
                            ),
                            pw.Text(
                              currencyFormat.format(totalBeforeDiscount),
                              style: const pw.TextStyle(
                                fontSize: 8,
                                color: PdfColors.grey700,
                                decoration: pw.TextDecoration.lineThrough,
                              ),
                            ),
                          ],
                        ),
                        pw.Row(
                          children: [
                            pw.Text(
                              'Total Diskon: ',
                              style: pw.TextStyle(
                                fontSize: 8,
                                color: PdfColors.red,
                              ),
                            ),
                            pw.Text(
                              '- ${currencyFormat.format(totalDiscount)}',
                              style: pw.TextStyle(
                                fontSize: 8,
                                color: PdfColors.red,
                              ),
                            ),
                          ],
                        ),
                      ],
                      pw.Row(
                        children: [
                          pw.Text(
                            totalDiscount > 0
                                ? 'Total Belanja: '
                                : 'Subtotal: ',
                            style: const pw.TextStyle(fontSize: 8),
                          ),
                          pw.Text(
                            currencyFormat.format(baseTotal),
                            style: pw.TextStyle(
                              fontSize: 8,
                              fontWeight: pw.FontWeight.bold,
                            ),
                          ),
                        ],
                      ),
                      if (_selectedReturns.isNotEmpty)
                        pw.Row(
                          children: [
                            pw.Text(
                              'Total Retur: ',
                              style: pw.TextStyle(
                                fontSize: 8,
                                color: PdfColors.red800,
                              ),
                            ),
                            pw.Text(
                              '- ${currencyFormat.format(_totalReturnAmount)}',
                              style: pw.TextStyle(
                                fontSize: 8,
                                color: PdfColors.red800,
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
                      pw.SizedBox(height: 2),
                      if (customerSigWidget != null)
                        customerSigWidget
                      else
                        pw.SizedBox(height: 20),
                      pw.SizedBox(height: 2),
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
                      pw.SizedBox(height: 2),
                      if (courierSigWidget != null)
                        courierSigWidget
                      else
                        pw.SizedBox(height: 20),
                      pw.SizedBox(height: 2),
                      pw.Text(
                        'Sales Motoris',
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

    return pdf.save();
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
      crossAxisAlignment: FlatButtonPaddingExtension.isWebAndroid
          ? CrossAxisAlignment.start
          : CrossAxisAlignment.start,
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
          canvas.drawCircle(
            points[i],
            paint.strokeWidth / 2.5,
            Paint()..color = paint.color,
          );
        } else if (points[i - 1] != Offset.infinite) {
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

class FlatButtonPaddingExtension {
  static bool get isWebAndroid =>
      kIsWeb && (defaultTargetPlatform == TargetPlatform.android);
}
