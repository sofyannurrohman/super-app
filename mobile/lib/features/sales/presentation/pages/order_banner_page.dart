import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:image_picker/image_picker.dart';
import 'package:file_picker/file_picker.dart';
import 'package:hris_app/features/sales/data/services/banner_order_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_state.dart';
import 'package:hris_app/core/theme/app_colors.dart';

class OrderBannerPage extends StatefulWidget {
  const OrderBannerPage({super.key});

  @override
  State<OrderBannerPage> createState() => _OrderBannerPageState();
}

class _OrderBannerPageState extends State<OrderBannerPage> {
  final _formKey = GlobalKey<FormState>();
  final _lengthCtrl = TextEditingController();
  final _widthCtrl = TextEditingController();
  final _qtyCtrl = TextEditingController(text: '1');
  final _notesCtrl = TextEditingController();
  String _bannerType = 'Custom';
  String _sizeUnit = 'm'; // 'm' or 'cm'
  bool _isLoading = false;
  Uint8List? _attachmentBytes;
  String? _attachmentPath;
  String? _attachmentName;
  bool _isPdfAttachment = false;
  final _apiService = BannerOrderApiService(apiClient: di.sl<ApiClient>());
  final _storeApiService = StoreApiService(apiClient: di.sl<ApiClient>(), prefs: di.sl<SharedPreferences>());

  List<StoreModel> _stores = [];
  StoreModel? _selectedStore;
  bool _isLoadingStores = false;

  @override
  void initState() {
    super.initState();
    _lengthCtrl.addListener(_onDimensionChanged);
    _widthCtrl.addListener(_onDimensionChanged);
    _fetchStores();
  }

  void _onDimensionChanged() {
    if (mounted) setState(() {});
  }

  void _switchUnit(String newUnit) {
    if (_sizeUnit == newUnit) return;
    final lenStr = _lengthCtrl.text.trim().replaceAll(',', '.');
    final widStr = _widthCtrl.text.trim().replaceAll(',', '.');
    final lenVal = double.tryParse(lenStr);
    final widVal = double.tryParse(widStr);

    setState(() {
      _sizeUnit = newUnit;
      if (lenVal != null && lenVal > 0) {
        if (newUnit == 'cm') {
          final converted = lenVal * 100;
          _lengthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(1);
        } else {
          final converted = lenVal / 100;
          _lengthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(2);
        }
      }
      if (widVal != null && widVal > 0) {
        if (newUnit == 'cm') {
          final converted = widVal * 100;
          _widthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(1);
        } else {
          final converted = widVal / 100;
          _widthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(2);
        }
      }
    });
  }

  void _applyPreset(double lengthM, double widthM) {
    setState(() {
      if (_sizeUnit == 'm') {
        _lengthCtrl.text = lengthM % 1 == 0
            ? lengthM.toInt().toString()
            : lengthM.toString();
        _widthCtrl.text = widthM % 1 == 0
            ? widthM.toInt().toString()
            : widthM.toString();
      } else {
        final lengthCm = lengthM * 100;
        final widthCm = widthM * 100;
        _lengthCtrl.text = lengthCm.toInt().toString();
        _widthCtrl.text = widthCm.toInt().toString();
      }
    });
  }

  Future<void> _fetchStores() async {
    setState(() => _isLoadingStores = true);
    try {
      String? employeeId;
      if (mounted) {
        final profileState = context.read<ProfileBloc>().state;
        if (profileState is ProfileLoaded) {
          employeeId = profileState.profile['id']?.toString();
        }
      }

      final stores = await _storeApiService.getStores(assignedEmployeeId: employeeId);
      setState(() {
        _stores = stores;
      });
    } catch (e) {
      debugPrint('Error fetching stores for banner: $e');
    } finally {
      if (mounted) setState(() => _isLoadingStores = false);
    }
  }

  @override
  void dispose() {
    _lengthCtrl.removeListener(_onDimensionChanged);
    _widthCtrl.removeListener(_onDimensionChanged);
    _lengthCtrl.dispose();
    _widthCtrl.dispose();
    _qtyCtrl.dispose();
    _notesCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickImage(ImageSource source) async {
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
          _attachmentBytes = bytes;
          _attachmentPath = pickedFile.path;
          _attachmentName = pickedFile.name;
          _isPdfAttachment = false;
        });
      }
    } catch (e) {
      debugPrint('Error picking image: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal mengambil foto: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _pickDocument() async {
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
          _attachmentBytes = bytes;
          _attachmentPath = file.path;
          _attachmentName = name;
          _isPdfAttachment = isPdf;
        });
      }
    } catch (e) {
      debugPrint('Error picking document: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal memilih file: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _submit() async {
    if (_isLoading) return;
    if (_selectedStore == null) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
          content: Text('Silakan pilih toko terlebih dahulu'),
          backgroundColor: Colors.orange));
      return;
    }
    if (!_formKey.currentState!.validate()) return;
    setState(() => _isLoading = true);

    try {
      final inputLen = double.parse(_lengthCtrl.text.trim().replaceAll(',', '.'));
      final inputWid = double.parse(_widthCtrl.text.trim().replaceAll(',', '.'));
      final lengthCm = _sizeUnit == 'm' ? (inputLen * 100) : inputLen;
      final widthCm = _sizeUnit == 'm' ? (inputWid * 100) : inputWid;
      final qty = int.tryParse(_qtyCtrl.text.trim()) ?? 1;

      final payload = {
        'shop_id': _selectedStore!.id,
        'shop_name': _selectedStore!.name,
        'size_length': lengthCm,
        'size_width': widthCm,
        'type': _bannerType,
        'is_bihun': false,
        'qty': qty,
        'notes': _notesCtrl.text.trim(),
      };

      await _apiService.createBannerOrder(
        payload,
        attachmentPath: _attachmentPath,
        attachmentBytes: _attachmentBytes,
        attachmentName: _attachmentName,
      );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
            content: Text('Order Spanduk berhasil dikirim'),
            backgroundColor: Colors.green));
        Navigator.pop(context);
      }
    } catch (e) {
      if (mounted) {
        if (e.toString().contains('OFFLINE_SAVED')) {
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
            content: Text(
                'Koneksi offline. Request spanduk disimpan secara offline dan akan dikirim otomatis saat online.'),
            backgroundColor: Colors.orange,
          ));
          Navigator.pop(context);
        } else {
          ScaffoldMessenger.of(context).showSnackBar(SnackBar(
              content: Text(e.toString()), backgroundColor: Colors.red));
        }
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded,
              color: Color(0xFF334155)),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('REQUEST BARU',
                style: GoogleFonts.outfit(
                    fontSize: 10,
                    color: AppColors.primaryGreen,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1)),
            Text('Order Spanduk',
                style: GoogleFonts.outfit(
                    fontSize: 17,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B))),
          ],
        ),
      ),
      body: Form(
        key: _formKey,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                    color: AppColors.primaryGreen.withOpacity(0.08),
                    borderRadius: BorderRadius.circular(16),
                    border:
                        Border.all(color: AppColors.primaryGreen.withOpacity(0.2))),
                child: Row(
                  children: [
                    const Icon(Icons.info_outline_rounded,
                        color: AppColors.primaryGreen, size: 20),
                    const SizedBox(width: 12),
                    Expanded(
                        child: Text(
                            'Pastikan ukuran akurat. Tim marketing akan memverifikasi request ini sebelum dicetak.',
                            style: GoogleFonts.outfit(
                                fontSize: 12,
                                color: Colors.blueGrey,
                                fontWeight: FontWeight.w600))),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              Text('INFORMASI SPANDUK',
                  style: GoogleFonts.outfit(
                      fontSize: 11,
                      fontWeight: FontWeight.w900,
                      color: Colors.blueGrey,
                      letterSpacing: 1.5)),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(24),
                    boxShadow: [
                      BoxShadow(
                          color: Colors.black.withOpacity(0.04),
                          blurRadius: 20,
                          offset: const Offset(0, 8))
                    ]),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Pilih Toko *',
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w700,
                            color: Colors.blueGrey,
                            fontSize: 13)),
                    const SizedBox(height: 8),
                    GestureDetector(
                      onTap: _isLoadingStores ? null : _showStorePicker,
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 16, vertical: 16),
                        decoration: BoxDecoration(
                          color: const Color(0xFFF8FAFC),
                          borderRadius: BorderRadius.circular(14),
                          border: Border.all(color: Colors.grey.shade200),
                        ),
                        child: Row(
                          children: [
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    _selectedStore?.name ??
                                        (_isLoadingStores
                                            ? 'Memuat toko...'
                                            : 'Pilih Toko...'),
                                    style: GoogleFonts.outfit(
                                      fontSize: 15,
                                      fontWeight: _selectedStore == null
                                          ? FontWeight.normal
                                          : FontWeight.w600,
                                      color: _selectedStore == null
                                          ? Colors.grey.shade400
                                          : const Color(0xFF1E293B),
                                    ),
                                  ),
                                  if (_selectedStore?.address != null &&
                                      _selectedStore!.address.isNotEmpty) ...[
                                    const SizedBox(height: 4),
                                    Text(
                                      _selectedStore!.address,
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: GoogleFonts.outfit(
                                          fontSize: 12, color: Colors.blueGrey),
                                    ),
                                  ]
                                ],
                              ),
                            ),
                            const Icon(Icons.arrow_drop_down,
                                color: Colors.blueGrey),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 16),
                    Text('Tipe Spanduk *',
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w700,
                            color: Colors.blueGrey,
                            fontSize: 13)),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        Expanded(
                          child: InkWell(
                            onTap: () => setState(() => _bannerType = 'Custom'),
                            borderRadius: BorderRadius.circular(14),
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                  vertical: 12, horizontal: 16),
                              decoration: BoxDecoration(
                                color: _bannerType == 'Custom'
                                    ? AppColors.primaryGreen.withOpacity(0.12)
                                    : const Color(0xFFF8FAFC),
                                borderRadius: BorderRadius.circular(14),
                                border: Border.all(
                                  color: _bannerType == 'Custom'
                                      ? AppColors.primaryGreen
                                      : Colors.grey.shade200,
                                  width: _bannerType == 'Custom' ? 1.5 : 1,
                                ),
                              ),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(
                                    _bannerType == 'Custom'
                                        ? Icons.radio_button_checked_rounded
                                        : Icons.radio_button_off_rounded,
                                    size: 18,
                                    color: _bannerType == 'Custom'
                                        ? AppColors.primaryGreen
                                        : Colors.grey.shade400,
                                  ),
                                  const SizedBox(width: 8),
                                  Text(
                                    'Custom',
                                    style: GoogleFonts.outfit(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w800,
                                      color: _bannerType == 'Custom'
                                          ? AppColors.primaryGreen
                                          : const Color(0xFF334155),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: InkWell(
                            onTap: () =>
                                setState(() => _bannerType = 'Universal'),
                            borderRadius: BorderRadius.circular(14),
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                  vertical: 12, horizontal: 16),
                              decoration: BoxDecoration(
                                color: _bannerType == 'Universal'
                                    ? const Color(0xFF0284C7).withOpacity(0.12)
                                    : const Color(0xFFF8FAFC),
                                borderRadius: BorderRadius.circular(14),
                                border: Border.all(
                                  color: _bannerType == 'Universal'
                                      ? const Color(0xFF0284C7)
                                      : Colors.grey.shade200,
                                  width: _bannerType == 'Universal' ? 1.5 : 1,
                                ),
                              ),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(
                                    _bannerType == 'Universal'
                                        ? Icons.radio_button_checked_rounded
                                        : Icons.radio_button_off_rounded,
                                    size: 18,
                                    color: _bannerType == 'Universal'
                                        ? const Color(0xFF0284C7)
                                        : Colors.grey.shade400,
                                  ),
                                  const SizedBox(width: 8),
                                  Text(
                                    'Universal',
                                    style: GoogleFonts.outfit(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w800,
                                      color: _bannerType == 'Universal'
                                          ? const Color(0xFF0284C7)
                                          : const Color(0xFF334155),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 20),
                    // Header Dimensi & Unit Switcher
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      crossAxisAlignment: CrossAxisAlignment.center,
                      children: [
                        Text(
                          'Ukuran Banner',
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w700,
                            color: Colors.blueGrey,
                            fontSize: 13,
                          ),
                        ),
                        _buildUnitSelector(),
                      ],
                    ),
                    const SizedBox(height: 10),
                    _buildPresetChips(),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: _buildField(
                            'Panjang ($_sizeUnit) *',
                            _sizeUnit == 'm' ? 'Contoh: 3.0' : 'Contoh: 300',
                            _lengthCtrl,
                            Icons.width_normal_rounded,
                            isDecimal: true,
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: _buildField(
                            'Tinggi ($_sizeUnit) *',
                            _sizeUnit == 'm' ? 'Contoh: 1.0' : 'Contoh: 100',
                            _widthCtrl,
                            Icons.height_rounded,
                            isDecimal: true,
                          ),
                        ),
                      ],
                    ),
                    _buildDimensionHelper(),
                    const SizedBox(height: 16),
                    _buildField('Banyaknya (Qty) *', '1', _qtyCtrl,
                        Icons.format_list_numbered_rounded,
                        isNumber: true),
                    const SizedBox(height: 16),
                    _buildField(
                        'Catatan Tambahan',
                        'Tulis pesan atau konten...',
                        _notesCtrl,
                        Icons.notes_rounded,
                        required: false,
                        maxLines: 3),
                    const SizedBox(height: 20),
                    _buildAttachmentSection(),
                  ],
                ),
              ),
              const SizedBox(height: 40),
              SizedBox(
                width: double.infinity,
                height: 64,
                child: ElevatedButton(
                  onPressed: _isLoading ? null : _submit,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.primaryGreen,
                    shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(20)),
                    elevation: 8,
                    shadowColor: AppColors.primaryGreen.withOpacity(0.4),
                  ),
                  child: _isLoading
                      ? const CircularProgressIndicator(color: Colors.white)
                      : Text('SUBMIT ORDER',
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w900,
                              fontSize: 16,
                              color: Colors.white,
                              letterSpacing: 0.5)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildUnitSelector() {
    return Container(
      padding: const EdgeInsets.all(3),
      decoration: BoxDecoration(
        color: const Color(0xFFF1F5F9),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.grey.shade300),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          _buildUnitOption('m', 'Meter (m)'),
          _buildUnitOption('cm', 'Centimeter (cm)'),
        ],
      ),
    );
  }

  Widget _buildUnitOption(String value, String label) {
    final isSelected = _sizeUnit == value;
    return GestureDetector(
      onTap: () => _switchUnit(value),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
        decoration: BoxDecoration(
          color: isSelected ? AppColors.primaryGreen : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: AppColors.primaryGreen.withOpacity(0.3),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  )
                ]
              : null,
        ),
        child: Text(
          label,
          style: GoogleFonts.outfit(
            fontSize: 11,
            fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
            color: isSelected ? Colors.white : const Color(0xFF64748B),
          ),
        ),
      ),
    );
  }

  Widget _buildPresetChips() {
    final presets = [
      {'label': '2 × 1 m', 'l': 2.0, 'w': 1.0},
      {'label': '3 × 1 m', 'l': 3.0, 'w': 1.0},
      {'label': '3 × 0.8 m', 'l': 3.0, 'w': 0.8},
      {'label': '4 × 1 m', 'l': 4.0, 'w': 1.0},
      {'label': '5 × 1 m', 'l': 5.0, 'w': 1.0},
    ];

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: presets.map((p) {
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: InkWell(
              borderRadius: BorderRadius.circular(8),
              onTap: () => _applyPreset(p['l'] as double, p['w'] as double),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.grey.shade300),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.straighten_rounded,
                        size: 13, color: AppColors.primaryGreen),
                    const SizedBox(width: 4),
                    Text(
                      p['label'] as String,
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: const Color(0xFF334155),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }

  Widget _buildDimensionHelper() {
    final lenStr = _lengthCtrl.text.trim().replaceAll(',', '.');
    final widStr = _widthCtrl.text.trim().replaceAll(',', '.');
    final lenVal = double.tryParse(lenStr);
    final widVal = double.tryParse(widStr);

    if (lenVal == null || lenVal <= 0 || widVal == null || widVal <= 0) {
      return const SizedBox.shrink();
    }

    final double lenCm = _sizeUnit == 'm' ? (lenVal * 100) : lenVal;
    final double widCm = _sizeUnit == 'm' ? (widVal * 100) : widVal;
    final double lenM = lenCm / 100;
    final double widM = widCm / 100;
    final double areaM2 = lenM * widM;

    final String alternateText = _sizeUnit == 'm'
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
          const Icon(Icons.swap_horiz_rounded,
              color: AppColors.primaryGreen, size: 16),
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

  Widget _buildField(
      String label, String hint, TextEditingController ctrl, IconData icon,
      {bool required = true, bool isDecimal = false, bool isNumber = false, int maxLines = 1}) {
    return TextFormField(
      controller: ctrl,
      maxLines: maxLines,
      keyboardType: isDecimal
          ? const TextInputType.numberWithOptions(decimal: true)
          : (isNumber ? TextInputType.number : TextInputType.text),
      style: GoogleFonts.outfit(
          fontSize: 15,
          fontWeight: FontWeight.w600,
          color: const Color(0xFF1E293B)),
      decoration: InputDecoration(
        labelText: label,
        hintText: hint,
        prefixIcon: Padding(
          padding: EdgeInsets.only(bottom: maxLines > 1 ? 40 : 0),
          child: Icon(icon, color: AppColors.primaryGreen, size: 20),
        ),
        labelStyle: GoogleFonts.outfit(
            fontWeight: FontWeight.w700, color: Colors.blueGrey, fontSize: 13),
        hintStyle:
            GoogleFonts.outfit(color: Colors.grey.shade400, fontSize: 13),
        border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(14),
            borderSide: BorderSide(color: Colors.grey.shade200)),
        enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(14),
            borderSide: BorderSide(color: Colors.grey.shade200)),
        focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(14),
            borderSide: const BorderSide(color: AppColors.primaryGreen, width: 2)),
        filled: true,
        fillColor: const Color(0xFFF8FAFC),
      ),
      validator: (v) {
        if (required && (v == null || v.isEmpty))
          return 'Field ini wajib diisi';
        if (isDecimal && v != null && v.isNotEmpty) {
          if (double.tryParse(v.replaceAll(',', '.')) == null)
            return 'Harus berupa angka (contoh: 2.5)';
        }
        if (isNumber && v != null && v.isNotEmpty) {
          final number = int.tryParse(v);
          if (number == null || number < 1)
            return 'Minimal 1 spanduk';
        }
        return null;
      },
    );
  }

  Widget _buildAttachmentSection() {
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
                color: Colors.blueGrey,
                fontSize: 13,
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
        if (_attachmentBytes != null || _attachmentPath != null) ...[
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: AppColors.primaryGreen.withOpacity(0.4)),
            ),
            child: Row(
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(10),
                  child: _isPdfAttachment
                      ? Container(
                          width: 50,
                          height: 50,
                          color: const Color(0xFFFEE2E2),
                          child: const Icon(
                            Icons.picture_as_pdf_rounded,
                            color: Color(0xFFDC2626),
                            size: 28,
                          ),
                        )
                      : (_attachmentBytes != null
                          ? Image.memory(
                              _attachmentBytes!,
                              width: 50,
                              height: 50,
                              fit: BoxFit.cover,
                            )
                          : Container(
                              width: 50,
                              height: 50,
                              color: const Color(0xFFE2E8F0),
                              child: const Icon(
                                Icons.image_rounded,
                                color: Color(0xFF64748B),
                                size: 28,
                              ),
                            )),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _attachmentName ?? 'File terlampir',
                        style: GoogleFonts.outfit(
                          fontSize: 13,
                          fontWeight: FontWeight.w700,
                          color: const Color(0xFF1E293B),
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 2),
                      Text(
                        _isPdfAttachment ? 'Dokumen PDF' : 'File Foto/Gambar',
                        style: GoogleFonts.outfit(
                          fontSize: 11,
                          color: Colors.blueGrey,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close_rounded, color: Color(0xFFDC2626), size: 20),
                  onPressed: () {
                    setState(() {
                      _attachmentBytes = null;
                      _attachmentPath = null;
                      _attachmentName = null;
                      _isPdfAttachment = false;
                    });
                  },
                ),
              ],
            ),
          ),
        ] else ...[
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: Colors.grey.shade200),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Lampirkan logo toko, referensi desain, atau file materi (JPG, PNG, PDF):',
                  style: GoogleFonts.outfit(
                    fontSize: 11.5,
                    color: Colors.blueGrey,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: InkWell(
                        onTap: () => _pickImage(ImageSource.camera),
                        borderRadius: BorderRadius.circular(12),
                        child: Container(
                          padding: const EdgeInsets.symmetric(vertical: 10),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: Colors.grey.shade300),
                          ),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              const Icon(Icons.camera_alt_outlined, size: 16, color: AppColors.primaryGreen),
                              const SizedBox(width: 6),
                              Text(
                                'Kamera',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  fontWeight: FontWeight.w700,
                                  color: const Color(0xFF334155),
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
                        onTap: () => _pickImage(ImageSource.gallery),
                        borderRadius: BorderRadius.circular(12),
                        child: Container(
                          padding: const EdgeInsets.symmetric(vertical: 10),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: Colors.grey.shade300),
                          ),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              const Icon(Icons.photo_library_outlined, size: 16, color: Color(0xFF0284C7)),
                              const SizedBox(width: 6),
                              Text(
                                'Galeri',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  fontWeight: FontWeight.w700,
                                  color: const Color(0xFF334155),
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
                        onTap: _pickDocument,
                        borderRadius: BorderRadius.circular(12),
                        child: Container(
                          padding: const EdgeInsets.symmetric(vertical: 10),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: Colors.grey.shade300),
                          ),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              const Icon(Icons.attach_file_rounded, size: 16, color: Color(0xFF8B5CF6)),
                              const SizedBox(width: 6),
                              Text(
                                'File / PDF',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  fontWeight: FontWeight.w700,
                                  color: const Color(0xFF334155),
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
        ],
      ],
    );
  }

  void _showStorePicker() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
      builder: (context) => _StorePickerModal(stores: _stores),
    ).then((selected) {
      if (selected != null) {
        setState(() => _selectedStore = selected);
      }
    });
  }
}

class _StorePickerModal extends StatefulWidget {
  final List<StoreModel> stores;
  const _StorePickerModal({required this.stores});
  @override
  State<_StorePickerModal> createState() => _StorePickerModalState();
}

class _StorePickerModalState extends State<_StorePickerModal> {
  String _search = '';

  @override
  Widget build(BuildContext context) {
    final filtered = widget.stores
        .where((s) =>
            s.name.toLowerCase().contains(_search.toLowerCase()) ||
            s.address.toLowerCase().contains(_search.toLowerCase()))
        .toList();

    return DraggableScrollableSheet(
      initialChildSize: 0.7,
      minChildSize: 0.4,
      maxChildSize: 0.9,
      expand: false,
      builder: (context, controller) => Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 20, 20, 12),
            child: Row(
              children: [
                Expanded(
                  child: Text('Pilih Toko',
                      style: GoogleFonts.outfit(
                          fontSize: 18,
                          fontWeight: FontWeight.w800,
                          color: const Color(0xFF1E293B))),
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            child: TextField(
              decoration: InputDecoration(
                hintText: 'Cari nama atau alamat toko...',
                prefixIcon: const Icon(Icons.search, color: AppColors.primaryGreen),
                border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(14),
                    borderSide: BorderSide.none),
                filled: true,
                fillColor: const Color(0xFFF1F5F9),
              ),
              style: GoogleFonts.outfit(fontSize: 14),
              onChanged: (val) => setState(() => _search = val),
            ),
          ),
          const SizedBox(height: 12),
          Expanded(
            child: ListView.separated(
              controller: controller,
              itemCount: filtered.length,
              separatorBuilder: (_, __) =>
                  const Divider(height: 1, color: Color(0xFFF1F5F9)),
              itemBuilder: (context, index) {
                final store = filtered[index];
                return ListTile(
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 24, vertical: 8),
                  title: Text(store.name,
                      style: GoogleFonts.outfit(
                          fontSize: 15,
                          fontWeight: FontWeight.w700,
                          color: const Color(0xFF1E293B))),
                  subtitle: Text(
                      store.address.isNotEmpty
                          ? store.address
                          : 'Alamat tidak tersedia',
                      style: GoogleFonts.outfit(
                          fontSize: 12, color: Colors.blueGrey),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis),
                  onTap: () => Navigator.pop(context, store),
                  trailing: const Icon(Icons.chevron_right_rounded,
                      color: Color(0xFFCBD5E1), size: 20),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}
