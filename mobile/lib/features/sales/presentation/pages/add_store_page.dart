import 'dart:io';
import 'package:cached_network_image/cached_network_image.dart';
import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:path_provider/path_provider.dart';
import 'package:uuid/uuid.dart';
import 'package:geolocator/geolocator.dart';
import 'package:dio/dio.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/network/network_info.dart';
import 'package:hris_app/features/sales/presentation/pages/map_picker_page.dart';
import 'package:image_picker/image_picker.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/sales/data/services/region_api_service.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';

class AddStorePage extends StatefulWidget {
  final StoreModel? existingStore;
  final bool isMT;
  const AddStorePage({super.key, this.existingStore, this.isMT = false});

  @override
  State<AddStorePage> createState() => _AddStorePageState();
}

class _AddStorePageState extends State<AddStorePage> {
  final _formKey = GlobalKey<FormState>();
  final _formScrollController = ScrollController();
  final _nameCtrl = TextEditingController();
  final _ownerCtrl = TextEditingController();
  final _phoneCtrl = TextEditingController();
  final _addressCtrl = TextEditingController();
  bool _isLoading = false;
  bool _isMtMode = false;
  String _selectedFrequency = 'F1';
  List<int> _selectedDays = [];
  double? _latitude;
  double? _longitude;
  double? _gpsAccuracy;
  bool _isLockingGps = false;
  bool _showPhotoGuide = true;

  // --- Multi-Product Initial Stock for MT ---
  List<Map<String, dynamic>> _availableProducts = [];
  bool _isLoadingProducts = false;
  final List<Map<String, dynamic>> _initialStockRows =
      []; // [{ 'product_id': '...', 'product_name': '...', 'sku': '...', 'controller': TextEditingController }]

  // --- Region (Wilayah) ---
  final _regionApiService = RegionApiService();
  List<dynamic> _provinces = [];
  List<dynamic> _regencies = [];
  List<dynamic> _districts = [];
  List<dynamic> _villages = [];

  String? _selectedProvinceId;
  String? _selectedProvinceName;
  String? _selectedRegencyId;
  String? _selectedRegencyName;
  String? _selectedDistrictId;
  String? _selectedDistrictName;
  String? _selectedVillageId;
  String? _selectedVillageName;
  bool _isLoadingRegions = false;

  // --- Hybrid Modern Trade (MT) Key Accounts ---
  List<Map<String, dynamic>> _keyAccounts = [];
  bool _isLoadingKeyAccounts = false;
  String? _selectedKeyAccountName;

  final List<String> _frequencies = ['F1', 'F2', 'F4'];
  final List<String> _dayNames = [
    'Sen',
    'Sel',
    'Rab',
    'Kam',
    'Jum',
    'Sab',
    'Min',
  ];
  final _apiService = StoreApiService(
    apiClient: di.sl<ApiClient>(),
    prefs: di.sl<SharedPreferences>(),
  );
  final _salesApiService = SalesApiService(apiClient: di.sl<ApiClient>());
  List<Map<String, dynamic>> _salesCandidates = [];
  String? _selectedPicId;
  bool _isManagement = false;

  String? _getLoggedInEmployeeId() {
    final prefs = di.sl<SharedPreferences>();
    final cached = prefs.getString('cached_employee_profile');
    if (cached != null) {
      try {
        final Map<String, dynamic> decoded = jsonDecode(cached);
        return decoded['id']?.toString();
      } catch (_) {}
    }
    return null;
  }

  final ImagePicker _picker = ImagePicker();
  XFile? _storePhoto;
  Uint8List? _storePhotoBytes;
  String? _storePhotoName;
  String? _existingPhotoUrl;

  void _pickStorePhoto(ImageSource source) async {
    try {
      // Pick image with highly efficient native resolution downsize & compression:
      // An original 6-8MB image from high-res camera will be downscaled to 1024x1024
      // and compressed to 65% jpeg quality. This results in a tiny file size (~100-150KB)
      // while preserving highly readable, sharp detail, not overloading the VPS bandwidth.
      final XFile? image = await _picker.pickImage(
        source: source,
        maxWidth: 1024,
        maxHeight: 1024,
        imageQuality: 65,
      );
      if (image != null) {
        final bytes = await image.readAsBytes();
        setState(() {
          _storePhoto = image;
          _storePhotoBytes = bytes;
          _storePhotoName = image.name;
        });
      }
    } catch (e) {
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

  void _showPhotoSourceSheet() {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(height: 16),
            Text(
              _isMtMode ? 'PILIH SUMBER FOTO MITRA' : 'PILIH SUMBER FOTO TOKO',
              style: GoogleFonts.outfit(
                fontSize: 14,
                fontWeight: FontWeight.w900,
                color: const Color(0xFF1E293B),
                letterSpacing: 0.8,
              ),
            ),
            const SizedBox(height: 16),
            ListTile(
              leading: const Icon(
                Icons.camera_alt_rounded,
                color: AppColors.primaryGreen,
              ),
              title: Text(
                'Ambil dari Kamera',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w600),
              ),
              onTap: () {
                Navigator.pop(context);
                _pickStorePhoto(ImageSource.camera);
              },
            ),
            ListTile(
              leading: const Icon(
                Icons.photo_library_rounded,
                color: AppColors.primaryGreen,
              ),
              title: Text(
                'Pilih dari Galeri',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w600),
              ),
              onTap: () {
                Navigator.pop(context);
                _pickStorePhoto(ImageSource.gallery);
              },
            ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }

  Future<void> _lockCurrentGps() async {
    setState(() => _isLockingGps = true);
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        throw Exception(
          'Layanan GPS/Lokasi HP belum aktif. Harap aktifkan GPS di pengaturan perangkat.',
        );
      }

      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) {
          throw Exception('Izin akses lokasi ditolak oleh pengguna.');
        }
      }

      if (permission == LocationPermission.deniedForever) {
        throw Exception(
          'Izin lokasi ditolak secara permanen. Harap aktifkan di menu pengaturan aplikasi.',
        );
      }

      // Ambil sinyal GPS satelit hardware murni dengan akurasi terbaik
      Position position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.best,
        timeLimit: const Duration(seconds: 25),
      );

      // Validasi Anti-Fake GPS
      if (position.isMocked) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text(
                'Peringatan: Terdeteksi Fake GPS / Mock Location! Gunakan sinyal lokasi fisik yang asli.',
              ),
              backgroundColor: Colors.red,
            ),
          );
        }
        return;
      }

      setState(() {
        _latitude = position.latitude;
        _longitude = position.longitude;
        _gpsAccuracy = position.accuracy;
      });

      // Coba reverse geocode jika ada internet, abaikan jika offline
      try {
        final response = await di.sl<ApiClient>().client.get(
          'https://nominatim.openstreetmap.org/reverse',
          queryParameters: {
            'lat': position.latitude,
            'lon': position.longitude,
            'format': 'json',
          },
          options: Options(
            headers: {'User-Agent': 'KarsaHRMobile/1.0'},
            sendTimeout: const Duration(seconds: 3),
            receiveTimeout: const Duration(seconds: 3),
          ),
        );
        if (response.statusCode == 200 && response.data != null) {
          final addr = response.data['display_name'];
          if (addr != null && _addressCtrl.text.trim().isEmpty) {
            _addressCtrl.text = addr.toString();
          }
        }
      } catch (_) {}

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Titik GPS Lokasi terkunci! Akurasi: ±${position.accuracy.toStringAsFixed(1)}m ${position.accuracy <= 20 ? "(Sangat Akurat)" : "(Cukup Akurat)"}',
            ),
            backgroundColor: position.accuracy <= 25
                ? Colors.green
                : Colors.orange,
            duration: const Duration(seconds: 3),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        // Coba gunakan last known position jika timeout
        try {
          final lastPos = await Geolocator.getLastKnownPosition();
          if (lastPos != null && !lastPos.isMocked) {
            setState(() {
              _latitude = lastPos.latitude;
              _longitude = lastPos.longitude;
              _gpsAccuracy = lastPos.accuracy;
            });
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(
                  'Menggunakan koordinat GPS tersimpan terakhir (Akurasi: ±${lastPos.accuracy.toStringAsFixed(1)}m)',
                ),
                backgroundColor: Colors.orange,
              ),
            );
            return;
          }
        } catch (_) {}

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal mengunci GPS: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isLockingGps = false);
    }
  }

  @override
  void initState() {
    super.initState();
    _isManagement = _isManagementRole();
    if (_isManagement) _loadSalesCandidates();
    _fetchProvinces();
    _isMtMode = widget.isMT;
    SharedPreferences.getInstance().then((prefs) {
      final salesMode = prefs.getString('sales_mode');
      if (mounted && salesMode == 'MT') {
        setState(() {
          _isMtMode = true;
        });
        _fetchKeyAccounts();
        _fetchProductsForMT();
      }
    });
    if (widget.isMT) {
      _fetchProductsForMT();
    }
    if (widget.existingStore != null) {
      _selectedPicId = widget.existingStore!.assignedEmployeeId;
      _nameCtrl.text = widget.existingStore!.name;
      _ownerCtrl.text = widget.existingStore!.ownerName;
      _selectedKeyAccountName = widget.existingStore!.ownerName;
      _phoneCtrl.text = widget.existingStore!.phoneNumber;
      _addressCtrl.text = widget.existingStore!.address;
      _selectedFrequency = widget.existingStore!.visitFrequency.isNotEmpty
          ? widget.existingStore!.visitFrequency
          : 'F1';
      _selectedDays = List<int>.from(widget.existingStore!.visitDays);
      _latitude = widget.existingStore!.latitude;
      _longitude = widget.existingStore!.longitude;
      _existingPhotoUrl = widget.existingStore!.effectivePhoto;
    }
  }

  @override
  void dispose() {
    _formScrollController.dispose();
    _nameCtrl.dispose();
    _ownerCtrl.dispose();
    _phoneCtrl.dispose();
    _addressCtrl.dispose();
    for (final row in _initialStockRows) {
      (row['controller'] as TextEditingController?)?.dispose();
    }
    super.dispose();
  }

  bool _isManagementRole() {
    final cached = di.sl<SharedPreferences>().getString(
      'cached_employee_profile',
    );
    if (cached == null) return false;
    try {
      final decoded = jsonDecode(cached) as Map<String, dynamic>;
      final role = (decoded['role'] ?? '').toString().toLowerCase();
      final title =
          (decoded['job_position']?['title'] ?? decoded['job_title'] ?? '')
              .toString()
              .toLowerCase();
      return [
            'direktur',
            'director',
            'owner',
            'superadmin',
            'manager',
          ].contains(role) ||
          title.contains('direktur') ||
          title.contains('director') ||
          title.contains('manager') ||
          title.contains('supervisor') ||
          title.contains('spv') ||
          title.contains('owner');
    } catch (_) {
      return false;
    }
  }

  Future<void> _fetchProductsForMT() async {
    setState(() => _isLoadingProducts = true);
    try {
      final prods = await _salesApiService.getActiveProducts();
      if (mounted) {
        setState(() {
          _availableProducts = prods;
          _isLoadingProducts = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => _isLoadingProducts = false);
    }
  }

  void _addInitialStockRow(Map<String, dynamic> product) {
    // Check if already in rows
    final existingIndex = _initialStockRows.indexWhere(
      (r) => r['product_id'] == product['id'],
    );
    if (existingIndex != -1) {
      return;
    }
    setState(() {
      _initialStockRows.add({
        'product_id': product['id'],
        'product_name': product['name'] ?? '',
        'sku': product['sku'] ?? '',
        'controller': TextEditingController(text: '0'),
      });
    });
  }

  void _removeInitialStockRow(int index) {
    setState(() {
      _initialStockRows[index]['controller']?.dispose();
      _initialStockRows.removeAt(index);
    });
  }

  void _showAddProductBottomSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) {
        String filter = '';
        return StatefulBuilder(
          builder: (ctx, setModalState) {
            final filtered = _availableProducts.where((p) {
              final name = (p['name'] ?? '').toString().toLowerCase();
              final sku = (p['sku'] ?? '').toString().toLowerCase();
              final q = filter.toLowerCase();
              return name.contains(q) || sku.contains(q);
            }).toList();

            return Container(
              height: MediaQuery.of(context).size.height * 0.7,
              decoration: const BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
              ),
              padding: const EdgeInsets.all(20),
              child: Column(
                children: [
                  Container(
                    width: 40,
                    height: 4,
                    decoration: BoxDecoration(
                      color: Colors.grey.shade300,
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'PILIH PRODUK STOK AWAL',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900,
                      fontSize: 14,
                      color: const Color(0xFF1E293B),
                    ),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    decoration: InputDecoration(
                      hintText: 'Cari nama produk / SKU...',
                      prefixIcon: const Icon(Icons.search_rounded, size: 20),
                      filled: true,
                      fillColor: const Color(0xFFF1F5F9),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(14),
                        borderSide: BorderSide.none,
                      ),
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                      ),
                    ),
                    onChanged: (val) => setModalState(() => filter = val),
                  ),
                  const SizedBox(height: 12),
                  Expanded(
                    child: _isLoadingProducts
                        ? const Center(child: CircularProgressIndicator())
                        : filtered.isEmpty
                        ? Center(
                            child: Text(
                              'Tidak ada produk ditemukan',
                              style: GoogleFonts.outfit(color: Colors.blueGrey),
                            ),
                          )
                        : ListView.separated(
                            itemCount: filtered.length,
                            separatorBuilder: (_, __) =>
                                const Divider(height: 1),
                            itemBuilder: (_, i) {
                              final p = filtered[i];
                              final isAdded = _initialStockRows.any(
                                (r) => r['product_id'] == p['id'],
                              );
                              return ListTile(
                                contentPadding: const EdgeInsets.symmetric(
                                  vertical: 4,
                                  horizontal: 8,
                                ),
                                title: Text(
                                  p['name'] ?? '',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w700,
                                    fontSize: 13,
                                  ),
                                ),
                                subtitle: Text(
                                  'SKU: ${p['sku'] ?? '-'} • Satuan: ${p['unit'] ?? 'PCS'}',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    color: Colors.blueGrey,
                                  ),
                                ),
                                trailing: isAdded
                                    ? const Icon(
                                        Icons.check_circle_rounded,
                                        color: Colors.green,
                                        size: 22,
                                      )
                                    : IconButton(
                                        icon: const Icon(
                                          Icons.add_circle_outline_rounded,
                                          color: Color(0xFF0284C7),
                                        ),
                                        onPressed: () {
                                          _addInitialStockRow(p);
                                          setModalState(() {});
                                        },
                                      ),
                                onTap: () {
                                  if (!isAdded) {
                                    _addInitialStockRow(p);
                                    setModalState(() {});
                                  }
                                },
                              );
                            },
                          ),
                  ),
                  const SizedBox(height: 12),
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF0284C7),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(14),
                        ),
                      ),
                      onPressed: () => Navigator.pop(ctx),
                      child: Text(
                        'SELESAI MEMILIH',
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                      ),
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

  Future<void> _loadSalesCandidates() async {
    final candidates = await _salesApiService.getSalesmenDirectory();
    if (mounted) setState(() => _salesCandidates = candidates);
  }

  Future<void> _fetchProvinces() async {
    setState(() => _isLoadingRegions = true);
    final data = await _regionApiService.getProvinces();
    if (mounted) {
      setState(() {
        _provinces = data;
        _isLoadingRegions = false;
      });
      // Jika existing store memiliki province, kita coba temukan ID-nya (asumsi name match)
      if (widget.existingStore != null &&
          widget.existingStore!.province.isNotEmpty) {
        try {
          final prov = _provinces.firstWhere(
            (p) => p['name'] == widget.existingStore!.province,
          );
          _selectedProvinceId = prov['id'];
          _selectedProvinceName = prov['name'];
          _fetchRegencies(_selectedProvinceId!);
        } catch (_) {}
      }
    }
  }

  Future<void> _fetchRegencies(String provinceId) async {
    setState(() => _isLoadingRegions = true);
    final data = await _regionApiService.getRegencies(provinceId);
    if (mounted) {
      setState(() {
        _regencies = data;
        _districts = [];
        _villages = [];
        _selectedRegencyId = null;
        _selectedRegencyName = null;
        _selectedDistrictId = null;
        _selectedDistrictName = null;
        _selectedVillageId = null;
        _selectedVillageName = null;
        _isLoadingRegions = false;
      });
      if (widget.existingStore != null &&
          widget.existingStore!.regency.isNotEmpty &&
          _selectedProvinceId != null) {
        try {
          final reg = _regencies.firstWhere(
            (r) => r['name'] == widget.existingStore!.regency,
          );
          _selectedRegencyId = reg['id'];
          _selectedRegencyName = reg['name'];
          _fetchDistricts(_selectedRegencyId!);
        } catch (_) {}
      }
    }
  }

  Future<void> _fetchDistricts(String regencyId) async {
    setState(() => _isLoadingRegions = true);
    final data = await _regionApiService.getDistricts(regencyId);
    if (mounted) {
      setState(() {
        _districts = data;
        _villages = [];
        _selectedDistrictId = null;
        _selectedDistrictName = null;
        _selectedVillageId = null;
        _selectedVillageName = null;
        _isLoadingRegions = false;
      });
      if (widget.existingStore != null &&
          widget.existingStore!.district.isNotEmpty &&
          _selectedRegencyId != null) {
        try {
          final dist = _districts.firstWhere(
            (d) => d['name'] == widget.existingStore!.district,
          );
          _selectedDistrictId = dist['id'];
          _selectedDistrictName = dist['name'];
          _fetchVillages(_selectedDistrictId!);
        } catch (_) {}
      }
    }
  }

  Future<void> _fetchVillages(String districtId) async {
    setState(() => _isLoadingRegions = true);
    final data = await _regionApiService.getVillages(districtId);
    if (mounted) {
      setState(() {
        _villages = data;
        _selectedVillageId = null;
        _selectedVillageName = null;
        _isLoadingRegions = false;
      });
      if (widget.existingStore != null &&
          widget.existingStore!.village.isNotEmpty &&
          _selectedDistrictId != null) {
        try {
          final vil = _villages.firstWhere(
            (v) => v['name'] == widget.existingStore!.village,
          );
          _selectedVillageId = vil['id'];
          _selectedVillageName = vil['name'];
        } catch (_) {}
      }
    }
  }

  Future<void> _fetchKeyAccounts() async {
    setState(() => _isLoadingKeyAccounts = true);
    try {
      final accounts = await _apiService.getKeyAccounts();
      if (mounted) {
        setState(() {
          _keyAccounts = accounts;
          _isLoadingKeyAccounts = false;
          // Auto-select or validate key account dropdown value
          if (_keyAccounts.isNotEmpty) {
            final hasMatch = _keyAccounts.any(
              (a) => (a['name'] ?? a['NAME']) == _selectedKeyAccountName,
            );
            if (!hasMatch && _selectedKeyAccountName == null) {
              _selectedKeyAccountName =
                  (_keyAccounts[0]['name'] ?? _keyAccounts[0]['NAME'])
                      ?.toString();
              _ownerCtrl.text = _selectedKeyAccountName!;
            }
          }
        });
      }
    } catch (e) {
      debugPrint('Error fetching key accounts: $e');
      if (mounted) {
        setState(() => _isLoadingKeyAccounts = false);
      }
    }
  }

  Future<bool?> _showDuplicateWarningSheet(List<dynamic> matches) async {
    return showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) {
        return Container(
          constraints: BoxConstraints(
            maxHeight: MediaQuery.of(ctx).size.height * 0.85,
          ),
          decoration: const BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
          child: SafeArea(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Center(
                  child: Container(
                    width: 40,
                    height: 4,
                    decoration: BoxDecoration(
                      color: Colors.grey.shade300,
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: Colors.amber.shade100,
                        shape: BoxShape.circle,
                      ),
                      child: Icon(
                        Icons.warning_amber_rounded,
                        color: Colors.amber.shade900,
                        size: 28,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Potensi Toko Duplikat',
                            style: GoogleFonts.outfit(
                              fontSize: 18,
                              fontWeight: FontWeight.w800,
                              color: const Color(0xFF1E293B),
                            ),
                          ),
                          Text(
                            'Ditemukan toko serupa yang sudah ada di sistem:',
                            style: GoogleFonts.outfit(
                              fontSize: 12,
                              color: Colors.grey.shade600,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                Flexible(
                  child: ListView.separated(
                    shrinkWrap: true,
                    itemCount: matches.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 12),
                    itemBuilder: (context, idx) {
                      final item = matches[idx] as Map<String, dynamic>;
                      final reasons =
                          (item['match_reasons'] as List<dynamic>?) ?? [];
                      final isMine = item['is_assigned_to_me'] == true;
                      final salesName =
                          item['assigned_employee_name']?.toString() ??
                          'Belum ada sales';
                      final branchName =
                          item['assigned_branch_name']?.toString() ?? '';

                      return Container(
                        padding: const EdgeInsets.all(14),
                        decoration: BoxDecoration(
                          color: const Color(0xFFF8FAFC),
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(color: Colors.grey.shade200),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                const Icon(
                                  Icons.storefront_rounded,
                                  color: AppColors.primaryGreen,
                                  size: 24,
                                ),
                                const SizedBox(width: 10),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        item['name'] ?? '-',
                                        style: GoogleFonts.outfit(
                                          fontSize: 15,
                                          fontWeight: FontWeight.w700,
                                          color: const Color(0xFF0F172A),
                                        ),
                                      ),
                                      if ((item['address'] ?? '').isNotEmpty)
                                        Text(
                                          item['address'],
                                          maxLines: 2,
                                          overflow: TextOverflow.ellipsis,
                                          style: GoogleFonts.outfit(
                                            fontSize: 12,
                                            color: Colors.grey.shade600,
                                          ),
                                        ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            Wrap(
                              spacing: 6,
                              runSpacing: 4,
                              children: [
                                for (var r in reasons)
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 3,
                                    ),
                                    decoration: BoxDecoration(
                                      color: Colors.amber.shade50,
                                      borderRadius: BorderRadius.circular(6),
                                      border: Border.all(
                                        color: Colors.amber.shade300,
                                      ),
                                    ),
                                    child: Text(
                                      r.toString(),
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w600,
                                        color: Colors.amber.shade900,
                                      ),
                                    ),
                                  ),
                              ],
                            ),
                            const SizedBox(height: 10),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 10,
                                vertical: 6,
                              ),
                              decoration: BoxDecoration(
                                color: isMine
                                    ? Colors.green.shade50
                                    : Colors.blue.shade50,
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: Row(
                                children: [
                                  Icon(
                                    isMine
                                        ? Icons.check_circle_rounded
                                        : Icons.person_rounded,
                                    size: 16,
                                    color: isMine
                                        ? Colors.green.shade700
                                        : Colors.blue.shade700,
                                  ),
                                  const SizedBox(width: 6),
                                  Expanded(
                                    child: Text(
                                      isMine
                                          ? 'Toko ini sudah ada di daftar Anda'
                                          : 'Sales: $salesName ${branchName.isNotEmpty ? "($branchName)" : ""}',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w600,
                                        color: isMine
                                            ? Colors.green.shade800
                                            : Colors.blue.shade800,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 12),
                            SizedBox(
                              width: double.infinity,
                              child: ElevatedButton.icon(
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: AppColors.primaryGreen,
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(10),
                                  ),
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 10,
                                  ),
                                ),
                                icon: const Icon(
                                  Icons.check_rounded,
                                  color: Colors.white,
                                  size: 18,
                                ),
                                label: Text(
                                  'Gunakan Toko Ini',
                                  style: GoogleFonts.outfit(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.white,
                                  ),
                                ),
                                onPressed: () async {
                                  final navigator = Navigator.of(context);
                                  final messenger = ScaffoldMessenger.of(
                                    context,
                                  );
                                  Navigator.pop(ctx, false);
                                  try {
                                    final storeId = item['store_id']
                                        ?.toString();
                                    if (storeId != null) {
                                      StoreModel store;
                                      try {
                                        store = await _apiService.getStoreByID(
                                          storeId,
                                        );
                                      } catch (_) {
                                        store = StoreModel(
                                          id: storeId,
                                          name: item['name']?.toString() ?? '',
                                          ownerName:
                                              item['owner_name']?.toString() ??
                                              '',
                                          phoneNumber:
                                              item['phone_number']
                                                  ?.toString() ??
                                              '',
                                          address:
                                              item['address']?.toString() ?? '',
                                          latitude:
                                              (item['latitude'] as num?)
                                                  ?.toDouble() ??
                                              0.0,
                                          longitude:
                                              (item['longitude'] as num?)
                                                  ?.toDouble() ??
                                              0.0,
                                        );
                                      }
                                      if (!mounted) return;
                                      navigator.pop(store);
                                    }
                                  } catch (e) {
                                    if (!mounted) return;
                                    messenger.showSnackBar(
                                      SnackBar(
                                        content: Text('Gagal memuat toko: $e'),
                                        backgroundColor: Colors.red,
                                      ),
                                    );
                                  }
                                },
                              ),
                            ),
                          ],
                        ),
                      );
                    },
                  ),
                ),
                const SizedBox(height: 16),
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton(
                    style: OutlinedButton.styleFrom(
                      side: BorderSide(color: Colors.grey.shade400),
                      padding: const EdgeInsets.symmetric(vertical: 12),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    onPressed: () => Navigator.pop(ctx, true),
                    child: Text(
                      'Bukan Toko Ini, Tetap Simpan Baru',
                      style: GoogleFonts.outfit(
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                        color: const Color(0xFF334155),
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 8),
                Center(
                  child: TextButton(
                    onPressed: () => Navigator.pop(ctx, false),
                    child: Text(
                      'Batal & Periksa Kembali Data',
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: Colors.grey.shade600,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Future<String?> _saveLocalStorePhoto(String storeId) async {
    Uint8List? bytes = _storePhotoBytes;
    if (bytes == null && _storePhoto != null) {
      try {
        bytes = await _storePhoto!.readAsBytes();
        _storePhotoBytes = bytes;
      } catch (_) {}
    }
    if (bytes == null || bytes.isEmpty) {
      return _existingPhotoUrl;
    }

    if (kIsWeb) {
      return 'data:image/jpeg;base64,${base64Encode(bytes)}';
    }

    try {
      final appDir = await getApplicationDocumentsDirectory();
      final dir = Directory('${appDir.path}/offline_stores');
      if (!dir.existsSync()) {
        dir.createSync(recursive: true);
      }
      final file = File('${dir.path}/$storeId.jpg');
      await file.writeAsBytes(bytes);
      return file.path;
    } catch (e) {
      debugPrint(
        '[OFFLINE STORE PHOTO] Gagal menyimpan file lokal, gunakan fallback base64: $e',
      );
      return 'data:image/jpeg;base64,${base64Encode(bytes)}';
    }
  }

  List<String> _collectFormValidationErrors() {
    final errors = <String>[];

    if (_nameCtrl.text.trim().isEmpty) {
      errors.add(
        _isMtMode ? 'Nama Mitra wajib diisi' : 'Nama Toko wajib diisi',
      );
    }

    if (_isMtMode) {
      if (_selectedKeyAccountName == null ||
          _selectedKeyAccountName!.trim().isEmpty) {
        errors.add('Key Account / Grup Mitra wajib dipilih');
      }
    } else {
      if (_ownerCtrl.text.trim().isEmpty) {
        errors.add('Nama Pemilik wajib diisi');
      }
    }

    final rawPhone = _phoneCtrl.text.trim();
    if (rawPhone.isEmpty) {
      errors.add(
        _isMtMode ? 'No. HP wajib diisi' : 'No. HP Pemilik wajib diisi',
      );
    } else {
      final phone = rawPhone.replaceAll(RegExp(r'[^0-9+]'), '');
      if (phone.length < 9 || phone.length > 15) {
        errors.add('No. HP tidak valid (min. 9 & maks. 15 digit)');
      } else if (!phone.startsWith('08') &&
          !phone.startsWith('628') &&
          !phone.startsWith('+628')) {
        errors.add('No. HP harus diawali 08, 628, atau +628');
      } else if (RegExp(r'(.)\1{6,}').hasMatch(phone)) {
        errors.add('No. HP terdeteksi tidak valid (angka berulang)');
      }
    }

    if (_provinces.isNotEmpty && _selectedProvinceId == null) {
      errors.add('Provinsi wajib dipilih');
    }
    if (_regencies.isNotEmpty && _selectedRegencyId == null) {
      errors.add('Kabupaten/Kota wajib dipilih');
    }
    if (_districts.isNotEmpty && _selectedDistrictId == null) {
      errors.add('Kecamatan wajib dipilih');
    }
    if (_villages.isNotEmpty && _selectedVillageId == null) {
      errors.add('Desa/Kelurahan wajib dipilih');
    }

    if (_addressCtrl.text.trim().isEmpty) {
      errors.add('Alamat Lengkap wajib diisi');
    }

    return errors;
  }

  void _showValidationErrorSnackBar(
    String message, {
    bool scrollToTop = false,
  }) {
    if (!mounted) return;
    if (scrollToTop && _formScrollController.hasClients) {
      _formScrollController.animateTo(
        0,
        duration: const Duration(milliseconds: 350),
        curve: Curves.easeOutCubic,
      );
    }
    ScaffoldMessenger.of(context).hideCurrentSnackBar();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(
              Icons.error_outline_rounded,
              color: Colors.white,
              size: 20,
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                message,
                style: GoogleFonts.outfit(
                  color: Colors.white,
                  fontWeight: FontWeight.w600,
                  fontSize: 13,
                ),
              ),
            ),
          ],
        ),
        backgroundColor: Colors.red.shade700,
        behavior: SnackBarBehavior.floating,
        showCloseIcon: true,
        closeIconColor: Colors.white,
        duration: const Duration(seconds: 5),
      ),
    );
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) {
      final fieldErrors = _collectFormValidationErrors();
      final detailMsg = fieldErrors.isNotEmpty
          ? 'Mohon perbaiki input berikut:\n• ${fieldErrors.join('\n• ')}'
          : 'Mohon periksa kembali input yang ditandai merah pada form di atas.';
      _showValidationErrorSnackBar(detailMsg, scrollToTop: true);
      return;
    }

    if (_latitude == null || _longitude == null) {
      _showValidationErrorSnackBar(
        'Titik lokasi GPS wajib dikunci ("KUNCI LOKASI GPS") atau dipilih dari peta terlebih dahulu.',
      );
      return;
    }

    final hasInternet = await NetworkInfo().hasInternetAccess;

    // Jika online dan daftar wilayah tersedia, wajibkan pilihan wilayah
    if (hasInternet && _provinces.isNotEmpty) {
      if (_selectedProvinceName == null ||
          _selectedRegencyName == null ||
          _selectedDistrictName == null ||
          _selectedVillageName == null) {
        _showValidationErrorSnackBar(
          'Provinsi, Kabupaten, Kecamatan, dan Desa wajib dipilih saat online.',
          scrollToTop: true,
        );
        return;
      }
    }

    if (_storePhotoBytes == null &&
        (_existingPhotoUrl == null || _existingPhotoUrl!.isEmpty)) {
      _showValidationErrorSnackBar(
        _isMtMode
            ? 'Foto tampak depan mitra wajib diambil/diunggah terlebih dahulu.'
            : 'Foto tampak depan toko wajib diambil/diunggah terlebih dahulu.',
      );
      return;
    }

    setState(() => _isLoading = true);

    final loggedInEmpId = _getLoggedInEmployeeId();
    final assignedEmpId =
        (_selectedPicId != null && _selectedPicId != '__SELF__')
        ? _selectedPicId
        : (widget.existingStore?.assignedEmployeeId != null &&
              widget.existingStore!.assignedEmployeeId!.isNotEmpty)
        ? widget.existingStore!.assignedEmployeeId!
        : (loggedInEmpId ?? '');

    final payload = <String, dynamic>{
      'name': _nameCtrl.text.trim(),
      'owner_name': _ownerCtrl.text.trim(),
      'phone_number': _phoneCtrl.text.trim(),
      'address': _addressCtrl.text.trim(),
      'latitude': _latitude!,
      'longitude': _longitude!,
      'is_new': widget.existingStore == null
          ? true
          : widget.existingStore!.isNew,
      'province':
          _selectedProvinceName ?? (widget.existingStore?.province ?? ''),
      'regency': _selectedRegencyName ?? (widget.existingStore?.regency ?? ''),
      'district':
          _selectedDistrictName ?? (widget.existingStore?.district ?? ''),
      'village': _selectedVillageName ?? (widget.existingStore?.village ?? ''),
      'visit_frequency': _selectedFrequency,
      'visit_days': _selectedDays,
      if (assignedEmpId != null && assignedEmpId.isNotEmpty)
        'assigned_employee_id': assignedEmpId,
    };

    // ALUR 1: MODE OFFLINE LANGSUNG (Tidak ada koneksi internet di Lokasi)
    if (!hasInternet) {
      try {
        final clientStoreId = widget.existingStore?.id ?? const Uuid().v4();
        payload['id'] = clientStoreId;
        final localPhoto = await _saveLocalStorePhoto(clientStoreId);
        if (widget.existingStore != null &&
            widget.existingStore!.photoUrl.isNotEmpty &&
            _storePhotoBytes == null) {
          payload['photo_url'] = widget.existingStore!.photoUrl;
        }

        final offlineStore = await _apiService.saveStoreOffline(
          data: payload,
          localPhotoPath: localPhoto,
          accuracy: _gpsAccuracy,
        );

        try {
          di.sl<SyncBloc>().add(const SyncDataRequested(force: true));
        } catch (_) {}

        if (mounted) {
          ScaffoldMessenger.of(context).hideCurrentSnackBar();
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                widget.existingStore != null
                    ? 'Perubahan toko berhasil disimpan OFFLINE! Akan disinkronkan saat online.'
                    : (_isMtMode
                          ? 'Mitra baru beserta foto berhasil disimpan OFFLINE! Siap untuk kunjungan.'
                          : 'Toko baru beserta foto berhasil disimpan OFFLINE! Siap untuk check-in & transaksi.'),
              ),
              backgroundColor: Colors.green.shade700,
              behavior: SnackBarBehavior.floating,
              showCloseIcon: true,
              closeIconColor: Colors.white,
              duration: const Duration(seconds: 4),
            ),
          );
          Navigator.pop(context, offlineStore);
        }
        return;
      } catch (e) {
        if (mounted) {
          _showValidationErrorSnackBar(
            'Gagal menyimpan toko secara offline: $e',
          );
        }
        setState(() => _isLoading = false);
        return;
      }
    }

    // ALUR 2: MODE ONLINE (Atau Edit Toko)
    try {
      // Pengecekan duplikasi toko secara otomatis saat tambah toko baru
      if (widget.existingStore == null && hasInternet) {
        try {
          final dupResult = await _apiService.checkDuplicateStore(
            name: _nameCtrl.text.trim(),
            phoneNumber: _phoneCtrl.text.trim(),
            latitude: _latitude,
            longitude: _longitude,
            district: _selectedDistrictName,
            regency: _selectedRegencyName,
            province: _selectedProvinceName,
          );

          if (dupResult['has_potential_duplicate'] == true &&
              dupResult['matches'] is List &&
              (dupResult['matches'] as List).isNotEmpty) {
            setState(() => _isLoading = false);
            final proceed = await _showDuplicateWarningSheet(
              dupResult['matches'] as List,
            );
            if (proceed != true) {
              return;
            }
            setState(() => _isLoading = true);
          }
        } catch (_) {}
      }

      if (widget.existingStore != null) {
        await _apiService.updateStore(
          widget.existingStore!.id,
          payload,
          photoBytes: _storePhotoBytes,
          photoName: _storePhotoName,
        );
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                _isMtMode
                    ? 'Mitra berhasil diperbarui'
                    : 'Toko berhasil diperbarui',
              ),
              backgroundColor: Colors.green,
            ),
          );
          Navigator.pop(context, true);
        }
      } else {
        // Pre-generate UUID agar client dan server memiliki ID yang identik
        final clientStoreId = const Uuid().v4();
        payload['id'] = clientStoreId;

        final store = await _apiService.createStore(
          payload,
          photoBytes: _storePhotoBytes,
          photoName: _storePhotoName,
        );

        // If MT Mode and initial stock items added, record initial store stock in bulk
        if (_isMtMode && store != null && _initialStockRows.isNotEmpty) {
          final itemsToAdjust = <Map<String, dynamic>>[];
          for (final row in _initialStockRows) {
            final ctrl = row['controller'] as TextEditingController?;
            final qty = int.tryParse(ctrl?.text.trim() ?? '') ?? 0;
            if (qty > 0) {
              itemsToAdjust.add({
                'product_id': row['product_id'],
                'quantity': qty,
                'min_stock_threshold': 20,
              });
            }
          }
          if (itemsToAdjust.isNotEmpty) {
            try {
              await _salesApiService.adjustMTStoreStocksBulk(
                storeId: store.id,
                items: itemsToAdjust,
                adjustmentType: 'INITIAL',
                notes: 'Saldo stok awal saat pendaftaran gerai MT baru',
                referenceNo: 'INIT-STORE-${store.id.substring(0, 8)}',
              );
            } catch (_) {}
          }
        }

        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                _isMtMode
                    ? 'Mitra baru & saldo awal berhasil ditambahkan'
                    : 'Toko baru berhasil ditambahkan',
              ),
              backgroundColor: Colors.green,
            ),
          );
          Navigator.pop(context, store);
        }
      }
    } catch (e) {
      // Fallback: Jika gagal karena koneksi internet putus di lapangan, alihkan ke penyimpanan offline
      final errStr = e.toString().toLowerCase();
      final isNetworkFail =
          errStr.contains('socket') ||
          errStr.contains('timeout') ||
          errStr.contains('network') ||
          errStr.contains('connection') ||
          errStr.contains('dioexception') ||
          errStr.contains('failed host lookup');

      if (isNetworkFail) {
        try {
          final clientStoreId =
              widget.existingStore?.id ?? payload['id'] ?? const Uuid().v4();
          payload['id'] = clientStoreId;
          final localPhoto = await _saveLocalStorePhoto(clientStoreId);

          final offlineStore = await _apiService.saveStoreOffline(
            data: payload,
            localPhotoPath: localPhoto,
            accuracy: _gpsAccuracy,
          );

          try {
            di.sl<SyncBloc>().add(const SyncDataRequested(force: true));
          } catch (_) {}

          if (mounted) {
            ScaffoldMessenger.of(context).hideCurrentSnackBar();
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: const Text(
                  'Koneksi internet terputus di Lokasi. Toko beserta foto otomatis disimpan OFFLINE di perangkat dan akan disinkronkan saat online.',
                ),
                backgroundColor: Colors.orange.shade800,
                behavior: SnackBarBehavior.floating,
                showCloseIcon: true,
                closeIconColor: Colors.white,
                duration: const Duration(seconds: 5),
              ),
            );
            Navigator.pop(context, offlineStore);
          }
          return;
        } catch (_) {}
      }

      if (mounted) {
        _showValidationErrorSnackBar(
          _isMtMode ? 'Gagal menyimpan mitra: $e' : 'Gagal menyimpan toko: $e',
        );
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
          icon: const Icon(
            Icons.arrow_back_ios_rounded,
            color: Color(0xFF334155),
          ),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              widget.existingStore != null
                  ? (_isMtMode ? 'EDIT MITRA' : 'EDIT TOKO')
                  : (_isMtMode ? 'TAMBAH MITRA BARU' : 'TAMBAH TOKO BARU'),
              style: GoogleFonts.outfit(
                fontSize: 10,
                color: AppColors.primaryGreen,
                fontWeight: FontWeight.w800,
                letterSpacing: 1,
              ),
            ),
            Text(
              _isMtMode ? 'Data Mitra Kunjungan' : 'Data Toko Kunjungan',
              style: GoogleFonts.outfit(
                fontSize: 17,
                fontWeight: FontWeight.w800,
                color: const Color(0xFF1E293B),
              ),
            ),
          ],
        ),
      ),
      body: Form(
        key: _formKey,
        child: SingleChildScrollView(
          controller: _formScrollController,
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildInfoBanner(),
              const SizedBox(height: 24),
              _buildSection(_isMtMode ? 'Informasi Mitra' : 'Informasi Toko', [
                _buildField(
                  _isMtMode ? 'Nama Mitra *' : 'Nama Toko *',
                  'Contoh: Toko Makmur Jaya',
                  _nameCtrl,
                  Icons.storefront_rounded,
                ),
                const SizedBox(height: 16),
                if (_isMtMode) ...[
                  _isLoadingKeyAccounts
                      ? const Center(child: CircularProgressIndicator())
                      : Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            DropdownButtonFormField<String>(
                              value: _selectedKeyAccountName,
                              decoration: InputDecoration(
                                labelText: 'Pilih Key Account / Grup Mitra *',
                                labelStyle: GoogleFonts.outfit(fontSize: 12),
                                filled: true,
                                fillColor: const Color(0xFFF8FAFC),
                                border: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  borderSide: BorderSide(
                                    color: Colors.grey.shade200,
                                  ),
                                ),
                                prefixIcon: const Icon(
                                  Icons.business_rounded,
                                  color: AppColors.primaryGreen,
                                ),
                              ),
                              items: _keyAccounts.map((a) {
                                final name =
                                    (a['name'] ?? a['NAME'])?.toString() ?? '';
                                return DropdownMenuItem<String>(
                                  value: name,
                                  child: Text(
                                    name,
                                    style: GoogleFonts.outfit(fontSize: 13),
                                  ),
                                );
                              }).toList(),
                              onChanged: (val) {
                                setState(() {
                                  _selectedKeyAccountName = val;
                                  if (val != null) {
                                    _ownerCtrl.text = val;
                                  }
                                });
                              },
                              validator: (val) => val == null || val.isEmpty
                                  ? 'Key account wajib dipilih'
                                  : null,
                            ),
                            const SizedBox(height: 6),
                            Align(
                              alignment: Alignment.centerRight,
                              child: TextButton.icon(
                                onPressed: _showAddKeyAccountDialog,
                                icon: const Icon(
                                  Icons.add_circle_outline_rounded,
                                  size: 14,
                                  color: AppColors.primaryGreen,
                                ),
                                label: Text(
                                  'Buat Grup Baru / Key Account',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    fontWeight: FontWeight.bold,
                                    color: AppColors.primaryGreen,
                                  ),
                                ),
                                style: TextButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 4,
                                  ),
                                  tapTargetSize:
                                      MaterialTapTargetSize.shrinkWrap,
                                ),
                              ),
                            ),
                          ],
                        ),
                ] else ...[
                  _buildField(
                    'Nama Pemilik *',
                    'Contoh: Pak Budi',
                    _ownerCtrl,
                    Icons.person_rounded,
                  ),
                ],
                if (_isManagement && !_isMtMode) ...[
                  const SizedBox(height: 16),
                  DropdownButtonFormField<String>(
                    value: _selectedPicId,
                    isExpanded: true,
                    decoration: InputDecoration(
                      labelText: 'PIC Toko',
                      prefixIcon: const Icon(Icons.assignment_ind_rounded),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    hint: const Text('Default: diri sendiri'),
                    items: [
                      const DropdownMenuItem<String>(
                        value: '__SELF__',
                        child: Text('Diri sendiri'),
                      ),
                      ..._salesCandidates.map((sales) {
                        final id = (sales['id'] ?? sales['employee_id'])
                            ?.toString();
                        final name =
                            (sales['name'] ??
                                    sales['full_name'] ??
                                    '${sales['first_name'] ?? ''} ${sales['last_name'] ?? ''}')
                                .toString()
                                .trim();
                        if (id == null || id.isEmpty) return null;
                        return DropdownMenuItem<String>(
                          value: id,
                          child: Text(name.isEmpty ? id : name),
                        );
                      }).whereType<DropdownMenuItem<String>>(),
                    ],
                    onChanged: (value) =>
                        setState(() => _selectedPicId = value),
                  ),
                ],
                const SizedBox(height: 16),
                _buildField(
                  _isMtMode ? 'No. HP *' : 'No. HP Pemilik *',
                  'Contoh: 0812XXXXXXXX',
                  _phoneCtrl,
                  Icons.phone_rounded,
                  required: true,
                  keyboardType: TextInputType.phone,
                  customValidator: (v) {
                    if (v == null || v.trim().isEmpty)
                      return 'No. HP wajib diisi';
                    final phone = v.trim().replaceAll(RegExp(r'[^0-9+]'), '');
                    if (phone.length < 9 || phone.length > 15) {
                      return 'No. HP tidak valid (min. 9 & maks. 15 digit)';
                    }
                    if (!phone.startsWith('08') &&
                        !phone.startsWith('628') &&
                        !phone.startsWith('+628')) {
                      return 'No. HP harus diawali 08, 628, atau +628';
                    }
                    if (RegExp(r'(.)\1{6,}').hasMatch(phone)) {
                      return 'No. HP terdeteksi tidak valid (asal-asalan)';
                    }
                    return null;
                  },
                ),
              ]),
              const SizedBox(height: 24),
              _buildSection(_isMtMode ? 'Lokasi Mitra' : 'Lokasi Toko', [
                _buildRegionDropdowns(),
                const SizedBox(height: 16),
                _buildField(
                  'Alamat Lengkap *',
                  'Jl. Nama Jalan No., Kota',
                  _addressCtrl,
                  Icons.location_on_rounded,
                  maxLines: 3,
                ),
                const SizedBox(height: 16),
                const SizedBox(height: 16),
                // Tombol Utama: Kunci Lokasi GPS Lokasi (Akurat & Offline)
                SizedBox(
                  width: double.infinity,
                  height: 52,
                  child: ElevatedButton.icon(
                    onPressed: _isLockingGps ? null : _lockCurrentGps,
                    icon: _isLockingGps
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(
                              color: Colors.white,
                              strokeWidth: 2.2,
                            ),
                          )
                        : const Icon(
                            Icons.my_location_rounded,
                            color: Colors.white,
                            size: 22,
                          ),
                    label: Text(
                      _isLockingGps
                          ? 'MENGUNCI SATELIT GPS...'
                          : 'KUNCI LOKASI GPS Lokasi (OFFLINE)',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                        fontSize: 13,
                        letterSpacing: 0.6,
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppColors.primaryGreen,
                      elevation: 2,
                      shadowColor: AppColors.primaryGreen.withOpacity(0.35),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(14),
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 10),
                // Tombol Sekunder: Pilih dari Peta Interaktif
                SizedBox(
                  width: double.infinity,
                  height: 44,
                  child: OutlinedButton.icon(
                    onPressed: () async {
                      final result = await Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => MapPickerPage(
                            initialLatitude: _latitude,
                            initialLongitude: _longitude,
                          ),
                        ),
                      );

                      if (result != null && result is Map<String, dynamic>) {
                        setState(() {
                          _latitude = result['latitude'];
                          _longitude = result['longitude'];
                          _addressCtrl.text = result['address'] ?? '';
                        });
                      }
                    },
                    icon: const Icon(
                      Icons.map_rounded,
                      color: Color(0xFF475569),
                      size: 18,
                    ),
                    label: Text(
                      'PILIH LOKASI DARI PETA',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w700,
                        color: const Color(0xFF475569),
                        fontSize: 12,
                        letterSpacing: 0.5,
                      ),
                    ),
                    style: OutlinedButton.styleFrom(
                      side: BorderSide(
                        color: Colors.blueGrey.shade300,
                        width: 1.2,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(14),
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 14),
                // Indikator Status & Akurasi GPS
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: (_latitude != null && _longitude != null)
                        ? Colors.green.withOpacity(0.06)
                        : Colors.orange.withOpacity(0.06),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: (_latitude != null && _longitude != null)
                          ? Colors.green.withOpacity(0.3)
                          : Colors.orange.withOpacity(0.3),
                      width: 1.2,
                    ),
                  ),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Icon(
                        (_latitude != null && _longitude != null)
                            ? Icons.gps_fixed_rounded
                            : Icons.gps_off_rounded,
                        color: (_latitude != null && _longitude != null)
                            ? Colors.green.shade700
                            : Colors.orange.shade800,
                        size: 24,
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Text(
                                  _isMtMode
                                      ? 'Titik GPS Mitra di Lokasi'
                                      : 'Titik GPS Toko di Lokasi',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w800,
                                    fontSize: 13,
                                    color: const Color(0xFF1E293B),
                                  ),
                                ),
                                const Spacer(),
                                if (_latitude != null && _longitude != null)
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 2,
                                    ),
                                    decoration: BoxDecoration(
                                      color: Colors.green.shade100,
                                      borderRadius: BorderRadius.circular(12),
                                    ),
                                    child: Text(
                                      'TERKUNCI',
                                      style: GoogleFonts.outfit(
                                        fontSize: 9,
                                        fontWeight: FontWeight.w900,
                                        color: Colors.green.shade900,
                                      ),
                                    ),
                                  ),
                              ],
                            ),
                            const SizedBox(height: 4),
                            Text(
                              (_latitude != null && _longitude != null)
                                  ? 'Latitude: $_latitude\nLongitude: $_longitude'
                                  : 'Belum mengunci koordinat lokasi. Tekan tombol "KUNCI LOKASI GPS Lokasi" di atas saat berada di depan toko.',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                color: Colors.blueGrey.shade700,
                                height: 1.35,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            if (_gpsAccuracy != null) ...[
                              const SizedBox(height: 6),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color: _gpsAccuracy! <= 20
                                      ? Colors.green.shade50
                                      : Colors.amber.shade50,
                                  borderRadius: BorderRadius.circular(8),
                                  border: Border.all(
                                    color: _gpsAccuracy! <= 20
                                        ? Colors.green.shade300
                                        : Colors.amber.shade400,
                                  ),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Icon(
                                      Icons.satellite_alt_rounded,
                                      size: 13,
                                      color: _gpsAccuracy! <= 20
                                          ? Colors.green.shade800
                                          : Colors.amber.shade900,
                                    ),
                                    const SizedBox(width: 5),
                                    Text(
                                      'Akurasi Satelit: ±${_gpsAccuracy!.toStringAsFixed(1)}m (${_gpsAccuracy! <= 20 ? "Sangat Akurat" : "Cukup Akurat"})',
                                      style: GoogleFonts.outfit(
                                        fontSize: 10,
                                        fontWeight: FontWeight.w700,
                                        color: _gpsAccuracy! <= 20
                                            ? Colors.green.shade800
                                            : Colors.amber.shade900,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ]),
              const SizedBox(height: 24),
              _buildSection(
                _isMtMode
                    ? 'Foto Mitra Tampak Depan (Wajib)'
                    : 'Foto Toko Tampak Depan (Wajib)',
                [
                  // ── Panel panduan foto yang baik & benar ──
                  Container(
                    margin: const EdgeInsets.only(bottom: 16),
                    decoration: BoxDecoration(
                      color: Colors.blue.shade50,
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(color: Colors.blue.shade100),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        InkWell(
                          onTap: () => setState(
                            () => _showPhotoGuide = !_showPhotoGuide,
                          ),
                          borderRadius: BorderRadius.circular(16),
                          child: Padding(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 14,
                              vertical: 12,
                            ),
                            child: Row(
                              children: [
                                Icon(
                                  Icons.tips_and_updates_rounded,
                                  color: Colors.blue.shade700,
                                  size: 18,
                                ),
                                const SizedBox(width: 8),
                                Text(
                                  'Contoh Foto yang Baik & Benar',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w800,
                                    color: Colors.blue.shade800,
                                    fontSize: 13,
                                  ),
                                ),
                                const Spacer(),
                                Text(
                                  _showPhotoGuide ? 'Tutup' : 'Lihat Contoh',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.blue.shade700,
                                  ),
                                ),
                                const SizedBox(width: 2),
                                Icon(
                                  _showPhotoGuide
                                      ? Icons.keyboard_arrow_up_rounded
                                      : Icons.keyboard_arrow_down_rounded,
                                  color: Colors.blue.shade700,
                                  size: 18,
                                ),
                              ],
                            ),
                          ),
                        ),
                        if (_showPhotoGuide) ...[
                          Container(
                            height: 140,
                            width: double.infinity,
                            color: const Color(0xFFE2E8F0),
                            padding: const EdgeInsets.all(16),
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(
                                  Icons.storefront_rounded,
                                  size: 40,
                                  color: Colors.blue.shade700,
                                ),
                                const SizedBox(height: 8),
                                Text(
                                  'Contoh foto fasad toko',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w800,
                                  ),
                                ),
                                Text(
                                  'Ambil foto lurus dari depan dengan papan nama terlihat',
                                  style: GoogleFonts.outfit(fontSize: 11),
                                ),
                              ],
                            ),
                          ),
                          // Chip tips foto
                          Padding(
                            padding: const EdgeInsets.fromLTRB(14, 4, 14, 14),
                            child: Wrap(
                              spacing: 6,
                              runSpacing: 6,
                              children: [
                                _buildTipChip('✅ Papan nama terlihat'),
                                _buildTipChip('✅ Foto dari depan'),
                                _buildTipChip('✅ Pencahayaan cukup'),
                                _buildTipChip('✅ Jarak ±3-5 meter'),
                              ],
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                  // ── Area upload foto ──
                  InkWell(
                    onTap: _showPhotoSourceSheet,
                    borderRadius: BorderRadius.circular(16),
                    child: Container(
                      height: 180,
                      width: double.infinity,
                      decoration: BoxDecoration(
                        color: const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: Colors.grey.shade200),
                      ),
                      child:
                          (_storePhotoBytes == null &&
                              (_existingPhotoUrl == null ||
                                  _existingPhotoUrl!.isEmpty))
                          ? Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                const Icon(
                                  Icons.add_a_photo_rounded,
                                  size: 40,
                                  color: AppColors.primaryGreen,
                                ),
                                const SizedBox(height: 10),
                                Text(
                                  _isMtMode
                                      ? 'Sentuh untuk Ambil Foto Tampak Depan Mitra'
                                      : 'Sentuh untuk Ambil Foto Tampak Depan Toko',
                                  style: GoogleFonts.outfit(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w700,
                                    color: const Color(0xFF334155),
                                  ),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  'Pastikan papan nama / area depan terlihat jelas',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    color: Colors.blueGrey.shade400,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ],
                            )
                          : Stack(
                              fit: StackFit.expand,
                              children: [
                                ClipRRect(
                                  borderRadius: BorderRadius.circular(16),
                                  child: _storePhotoBytes != null
                                      ? Image.memory(
                                          _storePhotoBytes!,
                                          fit: BoxFit.cover,
                                        )
                                      : _buildExistingOrLocalPhoto(
                                          _existingPhotoUrl!,
                                        ),
                                ),
                                Positioned(
                                  top: 12,
                                  right: 12,
                                  child: GestureDetector(
                                    onTap: () {
                                      setState(() {
                                        _storePhoto = null;
                                        _storePhotoBytes = null;
                                        _storePhotoName = null;
                                        _existingPhotoUrl = null;
                                      });
                                    },
                                    child: Container(
                                      padding: const EdgeInsets.all(6),
                                      decoration: const BoxDecoration(
                                        color: Colors.redAccent,
                                        shape: BoxShape.circle,
                                      ),
                                      child: const Icon(
                                        Icons.close_rounded,
                                        size: 16,
                                        color: Colors.white,
                                      ),
                                    ),
                                  ),
                                ),
                              ],
                            ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              _buildSection('Penjadwalan Kunjungan', [
                Text(
                  'Frekuensi Kunjungan',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w700,
                    fontSize: 13,
                    color: const Color(0xFF1E293B),
                  ),
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  initialValue: _selectedFrequency,
                  items: _frequencies
                      .map(
                        (f) => DropdownMenuItem(
                          value: f,
                          child: Text(
                            f == 'F1'
                                ? 'F1 (1x Seminggu)'
                                : f == 'F2'
                                ? 'F2 (2 Minggu Sekali)'
                                : 'F4 (4 Minggu Sekali)',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w600,
                              fontSize: 14,
                            ),
                          ),
                        ),
                      )
                      .toList(),
                  onChanged: (v) => setState(() => _selectedFrequency = v!),
                  decoration: InputDecoration(
                    prefixIcon: const Icon(
                      Icons.repeat_rounded,
                      color: AppColors.primaryGreen,
                      size: 20,
                    ),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(14),
                    ),
                    contentPadding: const EdgeInsets.symmetric(horizontal: 16),
                  ),
                ),
                const SizedBox(height: 8),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppColors.primaryGreen.withOpacity(0.06),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: AppColors.primaryGreen.withOpacity(0.15),
                    ),
                  ),
                  child: Row(
                    children: [
                      const Icon(
                        Icons.info_outline_rounded,
                        color: AppColors.primaryGreen,
                        size: 16,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _selectedFrequency == 'F1'
                              ? 'F1: Kunjungan rutin dilakukan 1 kali setiap minggu (Seminggu sekali).'
                              : _selectedFrequency == 'F2'
                              ? 'F2: Kunjungan rutin dilakukan 1 kali setiap 2 minggu (Dua minggu sekali).'
                              : 'F4: Kunjungan rutin dilakukan 1 kali setiap 4 minggu (Empat minggu sekali / Sebulan sekali).',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            color: const Color(0xFF334155),
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),
                Text(
                  'Hari Kunjungan',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w700,
                    fontSize: 13,
                    color: const Color(0xFF1E293B),
                  ),
                ),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: List.generate(7, (index) {
                    final dayNum = index + 1;
                    final isSelected = _selectedDays.contains(dayNum);
                    return ChoiceChip(
                      label: Text(_dayNames[index]),
                      selected: isSelected,
                      selectedColor: AppColors.primaryGreen,
                      labelStyle: GoogleFonts.outfit(
                        color: isSelected
                            ? Colors.white
                            : const Color(0xFF1E293B),
                        fontWeight: FontWeight.w700,
                        fontSize: 12,
                      ),
                      onSelected: (selected) {
                        setState(() {
                          if (selected) {
                            _selectedDays.add(dayNum);
                          } else {
                            _selectedDays.remove(dayNum);
                          }
                        });
                      },
                    );
                  }),
                ),
                if (_isMtMode) ...[
                  const SizedBox(height: 24),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          const Icon(
                            Icons.inventory_2_rounded,
                            size: 18,
                            color: Color(0xFF0284C7),
                          ),
                          const SizedBox(width: 8),
                          Text(
                            'Saldo Sisa Stok Awal Toko MT',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w700,
                              fontSize: 13,
                              color: const Color(0xFF1E293B),
                            ),
                          ),
                        ],
                      ),
                      TextButton.icon(
                        onPressed: _showAddProductBottomSheet,
                        icon: const Icon(
                          Icons.add_rounded,
                          size: 16,
                          color: Color(0xFF0284C7),
                        ),
                        label: Text(
                          'Tambah Produk',
                          style: GoogleFonts.outfit(
                            fontSize: 12,
                            fontWeight: FontWeight.w700,
                            color: const Color(0xFF0284C7),
                          ),
                        ),
                        style: TextButton.styleFrom(
                          backgroundColor: const Color(0xFFE0F2FE),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(10),
                          ),
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 6,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Text(
                    'Pilih produk dan input sisa barang fisik yang saat ini ada di toko/supermarket ini (dalam Pcs).',
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: Colors.blueGrey,
                    ),
                  ),
                  const SizedBox(height: 12),
                  if (_initialStockRows.isEmpty)
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.symmetric(
                        vertical: 20,
                        horizontal: 16,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Center(
                        child: Column(
                          children: [
                            const Icon(
                              Icons.shelves,
                              size: 32,
                              color: Colors.blueGrey,
                            ),
                            const SizedBox(height: 6),
                            Text(
                              'Belum ada produk stok awal ditambahkan.',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                color: Colors.blueGrey,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              'Klik "+ Tambah Produk" untuk memilih produk.',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                color: const Color(0xFF0284C7),
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ],
                        ),
                      ),
                    )
                  else
                    ListView.separated(
                      shrinkWrap: true,
                      physics: const NeverScrollableScrollPhysics(),
                      itemCount: _initialStockRows.length,
                      separatorBuilder: (_, __) => const SizedBox(height: 10),
                      itemBuilder: (ctx, idx) {
                        final row = _initialStockRows[idx];
                        final ctrl = row['controller'] as TextEditingController;
                        return Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: const Color(0xFFF0F9FF),
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(color: const Color(0xFFBAE6FD)),
                          ),
                          child: Row(
                            children: [
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      row['product_name'] ?? '',
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w700,
                                        fontSize: 13,
                                        color: const Color(0xFF1E293B),
                                      ),
                                    ),
                                    if ((row['sku'] ?? '')
                                        .toString()
                                        .isNotEmpty)
                                      Text(
                                        'SKU: ${row['sku']}',
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          color: Colors.blueGrey,
                                        ),
                                      ),
                                  ],
                                ),
                              ),
                              const SizedBox(width: 12),
                              SizedBox(
                                width: 90,
                                height: 44,
                                child: TextFormField(
                                  controller: ctrl,
                                  keyboardType: TextInputType.number,
                                  textAlign: TextAlign.center,
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.bold,
                                    fontSize: 13,
                                  ),
                                  decoration: InputDecoration(
                                    contentPadding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 8,
                                    ),
                                    suffixText: 'Pcs',
                                    suffixStyle: GoogleFonts.outfit(
                                      fontSize: 10,
                                      color: const Color(0xFF0284C7),
                                      fontWeight: FontWeight.bold,
                                    ),
                                    filled: true,
                                    fillColor: Colors.white,
                                    border: OutlineInputBorder(
                                      borderRadius: BorderRadius.circular(10),
                                      borderSide: const BorderSide(
                                        color: Color(0xFFBAE6FD),
                                      ),
                                    ),
                                    enabledBorder: OutlineInputBorder(
                                      borderRadius: BorderRadius.circular(10),
                                      borderSide: const BorderSide(
                                        color: Color(0xFFBAE6FD),
                                      ),
                                    ),
                                  ),
                                ),
                              ),
                              IconButton(
                                icon: const Icon(
                                  Icons.delete_outline_rounded,
                                  color: Colors.redAccent,
                                  size: 20,
                                ),
                                onPressed: () => _removeInitialStockRow(idx),
                              ),
                            ],
                          ),
                        );
                      },
                    ),
                ],
              ]),
              const SizedBox(height: 40),
              SizedBox(
                width: double.infinity,
                height: 64,
                child: ElevatedButton(
                  onPressed: _isLoading ? null : _submit,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.primaryGreen,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20),
                    ),
                    elevation: 8,
                    shadowColor: AppColors.primaryGreen.withOpacity(0.4),
                  ),
                  child: _isLoading
                      ? const CircularProgressIndicator(color: Colors.white)
                      : Text(
                          'SIMPAN & LANJUTKAN',
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w900,
                            fontSize: 16,
                            color: Colors.white,
                            letterSpacing: 0.5,
                          ),
                        ),
                ),
              ),
              const SizedBox(height: 40),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildExistingOrLocalPhoto(String photoRef) {
    if (photoRef.startsWith('data:image/')) {
      try {
        final commaIdx = photoRef.indexOf(',');
        final b64 = commaIdx != -1
            ? photoRef.substring(commaIdx + 1)
            : photoRef;
        return Image.memory(
          base64Decode(b64),
          fit: BoxFit.cover,
          errorBuilder: (_, __, ___) => const Icon(Icons.error),
        );
      } catch (_) {
        return const Icon(Icons.error);
      }
    }

    final isServerPath =
        photoRef.startsWith('http') ||
        photoRef.startsWith('/uploads/') ||
        photoRef.startsWith('uploads/');

    if (!kIsWeb && !isServerPath) {
      final file = File(photoRef);
      if (file.existsSync()) {
        return Image.file(
          file,
          fit: BoxFit.cover,
          errorBuilder: (_, __, ___) => const Icon(Icons.error),
        );
      }
    }

    final imageUrl = photoRef.startsWith('http')
        ? photoRef
        : '${Uri.parse(AppConstants.baseUrl).origin}${photoRef.startsWith('/') ? photoRef : '/$photoRef'}';
    return CachedNetworkImage(
      imageUrl: imageUrl,
      fit: BoxFit.cover,
      placeholder: (context, url) =>
          const Center(child: CircularProgressIndicator()),
      errorWidget: (context, url, error) => const Icon(Icons.error),
      memCacheWidth: 800,
    );
  }

  Widget _buildTipChip(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.blue.shade200),
      ),
      child: Text(
        label,
        style: GoogleFonts.outfit(
          fontSize: 10,
          fontWeight: FontWeight.w600,
          color: Colors.blue.shade700,
        ),
      ),
    );
  }

  Widget _buildInfoBanner() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.primaryGreen.withOpacity(0.08),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.primaryGreen.withOpacity(0.2)),
      ),
      child: Row(
        children: [
          const Icon(
            Icons.info_outline_rounded,
            color: AppColors.primaryGreen,
            size: 20,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              _isMtMode
                  ? 'Mitra baru akan langsung masuk ke daftar rayon Anda setelah disimpan.'
                  : 'Toko baru akan langsung masuk ke daftar rayon Anda setelah disimpan.',
              style: GoogleFonts.outfit(
                fontSize: 12,
                color: Colors.blueGrey,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _showAddKeyAccountDialog() {
    final nameCtrl = TextEditingController();
    final companyNameCtrl = TextEditingController();
    String selectedType = 'SUPERMARKET';
    final dialogFormKey = GlobalKey<FormState>();
    bool dialogLoading = false;

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(24),
          ),
          title: Text(
            'TAMBAH KEY ACCOUNT BARU',
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w900,
              fontSize: 16,
              color: const Color(0xFF1E293B),
            ),
          ),
          content: Form(
            key: dialogFormKey,
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  TextFormField(
                    controller: nameCtrl,
                    style: GoogleFonts.outfit(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: const Color(0xFF1E293B),
                    ),
                    decoration: InputDecoration(
                      labelText: 'Nama Key Account *',
                      labelStyle: GoogleFonts.outfit(fontSize: 12),
                      hintText: 'Contoh: Alfamart Jatim',
                      filled: true,
                      fillColor: const Color(0xFFF8FAFC),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    validator: (val) =>
                        val == null || val.isEmpty ? 'Nama wajib diisi' : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: companyNameCtrl,
                    style: GoogleFonts.outfit(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: const Color(0xFF1E293B),
                    ),
                    decoration: InputDecoration(
                      labelText: 'Nama Perusahaan *',
                      labelStyle: GoogleFonts.outfit(fontSize: 12),
                      hintText: 'Contoh: PT Sumber Alfaria Trijaya',
                      filled: true,
                      fillColor: const Color(0xFFF8FAFC),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    validator: (val) => val == null || val.isEmpty
                        ? 'Nama perusahaan wajib diisi'
                        : null,
                  ),
                  const SizedBox(height: 12),
                  DropdownButtonFormField<String>(
                    value: selectedType,
                    style: GoogleFonts.outfit(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: const Color(0xFF1E293B),
                    ),
                    decoration: InputDecoration(
                      labelText: 'Tipe Key Account *',
                      labelStyle: GoogleFonts.outfit(fontSize: 12),
                      filled: true,
                      fillColor: const Color(0xFFF8FAFC),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    items: const [
                      DropdownMenuItem(
                        value: 'SUPERMARKET',
                        child: Text('Supermarket'),
                      ),
                      DropdownMenuItem(
                        value: 'MINIMARKET',
                        child: Text('Minimarket'),
                      ),
                      DropdownMenuItem(
                        value: 'HYPERMARKET',
                        child: Text('Hypermarket'),
                      ),
                    ],
                    onChanged: (val) {
                      if (val != null) {
                        setDialogState(() => selectedType = val);
                      }
                    },
                  ),
                ],
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: dialogLoading ? null : () => Navigator.pop(ctx),
              child: Text(
                'BATAL',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w700,
                  color: Colors.grey,
                ),
              ),
            ),
            ElevatedButton(
              onPressed: dialogLoading
                  ? null
                  : () async {
                      if (!dialogFormKey.currentState!.validate()) return;
                      setDialogState(() => dialogLoading = true);
                      try {
                        final payload = {
                          'name': nameCtrl.text.trim(),
                          'company_name': companyNameCtrl.text.trim(),
                          'type': selectedType,
                          'status': 'ACTIVE',
                        };
                        final newAccount = await _apiService.createKeyAccount(
                          payload,
                        );
                        if (mounted) {
                          setState(() {
                            _keyAccounts.add(newAccount);
                            _selectedKeyAccountName =
                                (newAccount['name'] ?? newAccount['NAME'])
                                    ?.toString();
                            if (_selectedKeyAccountName != null) {
                              _ownerCtrl.text = _selectedKeyAccountName!;
                            }
                          });
                          Navigator.pop(ctx);
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text(
                                'Key Account baru berhasil ditambahkan!',
                              ),
                              backgroundColor: Colors.green,
                            ),
                          );
                        }
                      } catch (e) {
                        setDialogState(() => dialogLoading = false);
                        if (mounted) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(
                                'Gagal menambahkan Key Account: $e',
                              ),
                              backgroundColor: Colors.red,
                            ),
                          );
                        }
                      }
                    },
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primaryGreen,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              child: dialogLoading
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(
                        color: Colors.white,
                        strokeWidth: 2,
                      ),
                    )
                  : Text(
                      'TAMBAH',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                      ),
                    ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRegionDropdowns() {
    if (_isLoadingRegions) {
      return const Padding(
        padding: EdgeInsets.only(bottom: 16.0),
        child: Row(
          children: [
            SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
            SizedBox(width: 8),
            Text('Memuat data wilayah...'),
          ],
        ),
      );
    }

    if (_provinces.isEmpty) {
      return Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: Colors.blue.withOpacity(0.06),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: Colors.blue.withOpacity(0.25)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(
              Icons.info_outline_rounded,
              color: Colors.blue,
              size: 20,
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Mode Offline / Luar Jangkauan Data Wilayah',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      color: Colors.blue.shade900,
                    ),
                  ),
                  const SizedBox(height: 3),
                  Text(
                    'Wilayah (Desa, Kecamatan, Kabupaten, Provinsi) akan otomatis dideteksi dan dilengkapi oleh server berdasarkan titik koordinat GPS Lokasi yang Anda kunci saat data disinkronkan.',
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: Colors.blue.shade800,
                      height: 1.35,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        DropdownButtonFormField<String>(
          value: _selectedProvinceId,
          decoration: InputDecoration(
            labelText: 'Provinsi *',
            prefixIcon: const Icon(
              Icons.map_outlined,
              color: AppColors.primaryGreen,
              size: 20,
            ),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(14)),
            filled: true,
            fillColor: const Color(0xFFF8FAFC),
          ),
          items: _provinces
              .map(
                (p) => DropdownMenuItem<String>(
                  value: p['id'].toString(),
                  child: Text(
                    p['name'],
                    style: GoogleFonts.outfit(fontSize: 13),
                  ),
                ),
              )
              .toList(),
          onChanged: (val) {
            if (val != null) {
              setState(() {
                _selectedProvinceId = val;
                _selectedProvinceName = _provinces.firstWhere(
                  (p) => p['id'].toString() == val,
                )['name'];
              });
              _fetchRegencies(val);
            }
          },
          validator: (v) => (_provinces.isNotEmpty && v == null)
              ? 'Provinsi wajib dipilih'
              : null,
        ),
        const SizedBox(height: 16),
        DropdownButtonFormField<String>(
          value: _selectedRegencyId,
          decoration: InputDecoration(
            labelText: 'Kabupaten/Kota *',
            prefixIcon: const Icon(
              Icons.location_city_rounded,
              color: AppColors.primaryGreen,
              size: 20,
            ),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(14)),
            filled: true,
            fillColor: const Color(0xFFF8FAFC),
          ),
          items: _regencies
              .map(
                (r) => DropdownMenuItem<String>(
                  value: r['id'].toString(),
                  child: Text(
                    r['name'],
                    style: GoogleFonts.outfit(fontSize: 13),
                  ),
                ),
              )
              .toList(),
          onChanged: _regencies.isEmpty
              ? null
              : (val) {
                  if (val != null) {
                    setState(() {
                      _selectedRegencyId = val;
                      _selectedRegencyName = _regencies.firstWhere(
                        (r) => r['id'].toString() == val,
                      )['name'];
                    });
                    _fetchDistricts(val);
                  }
                },
          validator: (v) => (_regencies.isNotEmpty && v == null)
              ? 'Kabupaten/Kota wajib dipilih'
              : null,
        ),
        const SizedBox(height: 16),
        DropdownButtonFormField<String>(
          value: _selectedDistrictId,
          decoration: InputDecoration(
            labelText: 'Kecamatan *',
            prefixIcon: const Icon(
              Icons.holiday_village_rounded,
              color: AppColors.primaryGreen,
              size: 20,
            ),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(14)),
            filled: true,
            fillColor: const Color(0xFFF8FAFC),
          ),
          items: _districts
              .map(
                (d) => DropdownMenuItem<String>(
                  value: d['id'].toString(),
                  child: Text(
                    d['name'],
                    style: GoogleFonts.outfit(fontSize: 13),
                  ),
                ),
              )
              .toList(),
          onChanged: _districts.isEmpty
              ? null
              : (val) {
                  if (val != null) {
                    setState(() {
                      _selectedDistrictId = val;
                      _selectedDistrictName = _districts.firstWhere(
                        (d) => d['id'].toString() == val,
                      )['name'];
                    });
                    _fetchVillages(val);
                  }
                },
          validator: (v) => (_districts.isNotEmpty && v == null)
              ? 'Kecamatan wajib dipilih'
              : null,
        ),
        const SizedBox(height: 16),
        DropdownButtonFormField<String>(
          value: _selectedVillageId,
          decoration: InputDecoration(
            labelText: 'Desa/Kelurahan *',
            prefixIcon: const Icon(
              Icons.home_work_rounded,
              color: AppColors.primaryGreen,
              size: 20,
            ),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(14)),
            filled: true,
            fillColor: const Color(0xFFF8FAFC),
          ),
          items: _villages
              .map(
                (v) => DropdownMenuItem<String>(
                  value: v['id'].toString(),
                  child: Text(
                    v['name'],
                    style: GoogleFonts.outfit(fontSize: 13),
                  ),
                ),
              )
              .toList(),
          onChanged: _villages.isEmpty
              ? null
              : (val) {
                  if (val != null) {
                    setState(() {
                      _selectedVillageId = val;
                      _selectedVillageName = _villages.firstWhere(
                        (v) => v['id'].toString() == val,
                      )['name'];
                    });
                  }
                },
          validator: (v) => (_villages.isNotEmpty && v == null)
              ? 'Desa/Kelurahan wajib dipilih'
              : null,
        ),
      ],
    );
  }

  Widget _buildSection(String title, List<Widget> children) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title.toUpperCase(),
          style: GoogleFonts.outfit(
            fontSize: 11,
            fontWeight: FontWeight.w900,
            color: Colors.blueGrey,
            letterSpacing: 1.5,
          ),
        ),
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
                offset: const Offset(0, 8),
              ),
            ],
          ),
          child: Column(children: children),
        ),
      ],
    );
  }

  Widget _buildField(
    String label,
    String hint,
    TextEditingController ctrl,
    IconData icon, {
    bool required = true,
    int maxLines = 1,
    TextInputType keyboardType = TextInputType.text,
    String? Function(String?)? customValidator,
  }) {
    return TextFormField(
      controller: ctrl,
      maxLines: maxLines,
      keyboardType: keyboardType,
      style: GoogleFonts.outfit(
        fontSize: 15,
        fontWeight: FontWeight.w600,
        color: const Color(0xFF1E293B),
      ),
      decoration: InputDecoration(
        labelText: label,
        hintText: hint,
        prefixIcon: Icon(icon, color: AppColors.primaryGreen, size: 20),
        labelStyle: GoogleFonts.outfit(
          fontWeight: FontWeight.w700,
          color: Colors.blueGrey,
          fontSize: 13,
        ),
        hintStyle: GoogleFonts.outfit(
          color: Colors.grey.shade400,
          fontSize: 13,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: BorderSide(color: Colors.grey.shade200),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: BorderSide(color: Colors.grey.shade200),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: AppColors.primaryGreen, width: 2),
        ),
        filled: true,
        fillColor: const Color(0xFFF8FAFC),
      ),
      validator:
          customValidator ??
          (required
              ? (v) => (v == null || v.isEmpty) ? 'Field ini wajib diisi' : null
              : null),
    );
  }
}
