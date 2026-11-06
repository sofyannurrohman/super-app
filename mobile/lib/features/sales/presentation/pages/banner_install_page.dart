import 'dart:async';
import 'package:flutter/material.dart';
import 'package:get_it/get_it.dart';
import 'package:image_picker/image_picker.dart';
import 'package:geolocator/geolocator.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/features/sales/data/services/banner_order_api_service.dart';

class BannerInstallPage extends StatefulWidget {
  const BannerInstallPage({super.key});

  @override
  State<BannerInstallPage> createState() => _BannerInstallPageState();
}

class _BannerInstallPageState extends State<BannerInstallPage>
    with SingleTickerProviderStateMixin {
  final _apiService = GetIt.I<BannerOrderApiService>();
  late TabController _tabController;
  late ScrollController _scrollController;
  late final TextEditingController _searchController;

  bool _isInitialLoading = true;
  bool _isBannersLoading = false;
  bool _isLoadingMore = false;
  final List<dynamic> _banners = [];
  List<dynamic> _unclaimedInstallations = [];
  String _searchQuery = '';
  int _currentPage = 1;
  static const int _pageSize = 20;
  bool _hasMoreData = true;
  int _totalCount = 0;
  Timer? _debounce;

  @override
  void initState() {
    super.initState();
    _searchController = TextEditingController();
    _tabController = TabController(length: 2, vsync: this);
    _scrollController = ScrollController();
    _scrollController.addListener(_onScroll);
    _fetchAllData(isInitial: true);
  }

  @override
  void dispose() {
    _searchController.dispose();
    _tabController.dispose();
    _scrollController.dispose();
    _debounce?.cancel();
    super.dispose();
  }

  Future<void> _fetchAllData({bool isInitial = false}) async {
    if (isInitial && mounted) {
      setState(() => _isInitialLoading = true);
    }
    await Future.wait([
      _fetchBanners(reset: true),
      _fetchUnclaimedInstallations(),
    ]);
    if (isInitial && mounted) {
      setState(() => _isInitialLoading = false);
    }
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - 300) {
      _loadMoreBanners();
    }
  }

  void _onSearchChanged(String value) {
    _debounce?.cancel();
    if (mounted) setState(() => _searchQuery = value);
    if (value.isEmpty) {
      _fetchBanners(reset: true);
    } else {
      _debounce = Timer(const Duration(milliseconds: 400), () {
        _fetchBanners(reset: true);
      });
    }
  }

  Future<void> _fetchBanners({bool reset = true}) async {
    if (reset) {
      if (mounted) {
        setState(() {
          _currentPage = 1;
          _banners.clear();
          _hasMoreData = true;
          _isBannersLoading = true;
        });
      }
    }
    try {
      final result = await _apiService.getBannersPaginated(
        status: 'ready_for_operator',
        page: _currentPage,
        limit: _pageSize,
        search: _searchQuery.isEmpty ? null : _searchQuery,
      );
      final List<dynamic> data = List<dynamic>.from(result['data'] ?? []);
      final int total = (result['total'] as num?)?.toInt() ?? 0;
      if (mounted) {
        setState(() {
          _banners.addAll(data);
          _totalCount = total;
          _hasMoreData = _banners.length < total;
          _isBannersLoading = false;
        });
      }
    } catch (e) {
      debugPrint('Error fetching banners: $e');
      if (mounted) setState(() => _isBannersLoading = false);
    }
  }

  Future<void> _loadMoreBanners() async {
    if (_isLoadingMore || !_hasMoreData) return;
    setState(() {
      _isLoadingMore = true;
      _currentPage++;
    });
    try {
      final result = await _apiService.getBannersPaginated(
        status: 'ready_for_operator',
        page: _currentPage,
        limit: _pageSize,
        search: _searchQuery.isEmpty ? null : _searchQuery,
      );
      final List<dynamic> data = List<dynamic>.from(result['data'] ?? []);
      if (mounted) {
        setState(() {
          _banners.addAll(data);
          _hasMoreData = _banners.length < _totalCount;
          _isLoadingMore = false;
        });
      }
    } catch (e) {
      debugPrint('Error loading more banners: $e');
      if (mounted) {
        setState(() {
          _currentPage--;
          _isLoadingMore = false;
        });
      }
    }
  }

  Future<void> _fetchUnclaimedInstallations() async {
    try {
      final results = await _apiService.getUnclaimedInstallations();
      if (mounted) {
        setState(() {
          _unclaimedInstallations = results;
        });
      }
    } catch (e) {
      debugPrint('Error fetching unclaimed installations: $e');
    }
  }

  Future<void> _launchMaps(double lat, double lng) async {
    final url = 'https://www.google.com/maps/search/?api=1&query=$lat,$lng';
    final uri = Uri.parse(url);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Tidak dapat membuka Google Maps')),
        );
      }
    }
  }

  Future<Position?> _determinePosition() async {
    bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) {
      throw Exception('GPS tidak aktif. Mohon aktifkan lokasi di perangkat Anda.');
    }

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) {
        throw Exception('Izin akses lokasi ditolak');
      }
    }

    if (permission == LocationPermission.deniedForever) {
      throw Exception('Izin akses lokasi ditolak permanen. Buka pengaturan untuk mengaktifkan.');
    }

    return await Geolocator.getCurrentPosition();
  }

  Future<void> _processInstall(dynamic banner) async {
    Position? position;
    try {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (c) => const Center(child: CircularProgressIndicator()),
      );
      position = await _determinePosition();
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: Colors.red),
        );
      }
      return;
    }

    if (position == null) return;

    final picker = ImagePicker();
    final XFile? image = await picker.pickImage(
      maxWidth: 1600,
      maxHeight: 1600,
      source: ImageSource.camera,
      imageQuality: 80,
    );

    if (image == null) return;

    if (!mounted) return;
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (c) => const Center(child: CircularProgressIndicator()),
    );

    try {
      final bytes = await image.readAsBytes();
      await _apiService.installBanner(
        banner['id'],
        image.path,
        position.latitude,
        position.longitude,
        imageBytes: bytes,
        filename: image.name,
      );

      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Pemasangan Spanduk Berhasil Disimpan'),
            backgroundColor: Colors.green,
          ),
        );
        _fetchAllData();
      }
    } catch (e) {
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: Colors.red),
        );
      }
    }
  }

  static const double _ratePerM2 = 1500.0;

  double _getBannerArea(dynamic item) {
    double l = double.tryParse(item['size_length']?.toString() ?? '') ?? 0.0;
    double w = double.tryParse(item['size_width']?.toString() ?? '') ?? 0.0;
    if (l > 0 && w > 0) {
      double lInM = l >= 10 ? l / 100.0 : l;
      double wInM = w >= 10 ? w / 100.0 : w;
      return lInM * wInM;
    }
    double s = double.tryParse(item['size']?.toString() ?? '') ?? 0.0;
    if (s > 0) {
      if (s > 100) return s / 10000.0;
      return s;
    }
    return 3.0; // default 3x1
  }

  String _formatBannerDimension(dynamic lengthVal, dynamic widthVal) {
    double l = double.tryParse(lengthVal?.toString() ?? '') ?? 0.0;
    double w = double.tryParse(widthVal?.toString() ?? '') ?? 0.0;

    if (l <= 0 && w <= 0) return 'Standard (3.0 m²)';

    bool isCm = l >= 10 || w >= 10;
    double lInM = isCm ? l / 100.0 : l;
    double wInM = isCm ? w / 100.0 : w;
    double area = lInM * wInM;

    String formatNum(double n) {
      if (n == n.roundToDouble()) return n.toInt().toString();
      return n.toStringAsFixed(1);
    }

    final unit = isCm ? 'cm' : 'm';
    return '${formatNum(l)} × ${formatNum(w)} $unit (${area.toStringAsFixed(1)} m²)';
  }

  double _getBannerEarnings(dynamic item) {
    return _getBannerArea(item) * _ratePerM2;
  }

  double get _totalUnclaimedArea {
    double total = 0.0;
    for (var item in _unclaimedInstallations) {
      total += _getBannerArea(item);
    }
    return total;
  }

  double get _totalEstimatedEarnings {
    return _totalUnclaimedArea * _ratePerM2;
  }

  void _showImagePreview(String imageUrl, String title) {
    showDialog(
      context: context,
      builder: (ctx) => Dialog(
        backgroundColor: Colors.transparent,
        insetPadding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Expanded(
                  child: Text(
                    title,
                    style: const TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close, color: Colors.white),
                  onPressed: () => Navigator.pop(ctx),
                ),
              ],
            ),
            const SizedBox(height: 8),
            ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: Image.network(
                imageUrl,
                fit: BoxFit.contain,
                loadingBuilder: (c, child, progress) {
                  if (progress == null) return child;
                  return const Center(
                    child: Padding(
                      padding: EdgeInsets.all(32),
                      child: CircularProgressIndicator(color: Colors.white),
                    ),
                  );
                },
                errorBuilder: (_, __, ___) => Container(
                  height: 200,
                  color: Colors.black54,
                  alignment: Alignment.center,
                  child: const Text(
                    'Gagal memuat gambar',
                    style: TextStyle(color: Colors.white70),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _openBulkReimburseDialog() {
    if (_unclaimedInstallations.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Tidak ada spanduk terpasang yang dapat diajukan reimbursement'),
        ),
      );
      return;
    }

    final currencyFmt = NumberFormat('#,###', 'id_ID');
    final totalRpStr = 'Rp ${currencyFmt.format(_totalEstimatedEarnings)}';
    final notesCtrl = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xFF047857).withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(Icons.receipt_long, color: Color(0xFF047857)),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                'Ajukan Reimbursement',
                style: GoogleFonts.outfit(fontWeight: FontWeight.bold, fontSize: 18),
              ),
            ),
          ],
        ),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: const Color(0xFFECFDF5),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: const Color(0xFFA7F3D0)),
                ),
                child: Column(
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Total Spanduk:',
                          style: TextStyle(fontSize: 13, color: Color(0xFF065F46)),
                        ),
                        Text(
                          '${_unclaimedInstallations.length} Spanduk',
                          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Color(0xFF065F46)),
                        ),
                      ],
                    ),
                    const SizedBox(height: 6),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Total Luas:',
                          style: TextStyle(fontSize: 13, color: Color(0xFF065F46)),
                        ),
                        Text(
                          '${_totalUnclaimedArea.toStringAsFixed(1)} m²',
                          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Color(0xFF065F46)),
                        ),
                      ],
                    ),
                    const SizedBox(height: 6),
                    const Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Tarif Pemasangan:',
                          style: TextStyle(fontSize: 13, color: Color(0xFF065F46)),
                        ),
                        Text(
                          'Rp 1.500 / m²',
                          style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Color(0xFF047857)),
                        ),
                      ],
                    ),
                    const Divider(height: 16, color: Color(0xFFA7F3D0)),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Estimasi Pendapatan:',
                          style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF065F46)),
                        ),
                        Text(
                          totalRpStr,
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.bold,
                            fontSize: 18,
                            color: const Color(0xFF047857),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 12),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F5F9),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Text(
                  'Catatan: Nominal ini akan diteruskan ke Admin Kasir untuk verifikasi foto dan pencairan dana reimbursement.',
                  style: TextStyle(fontSize: 11, color: Color(0xFF475569), fontStyle: FontStyle.italic),
                ),
              ),
              const SizedBox(height: 14),
              TextField(
                controller: notesCtrl,
                maxLines: 2,
                decoration: InputDecoration(
                  labelText: 'Catatan Pengajuan (Opsional)',
                  hintText: 'Contoh: Pemasangan spanduk area Sleman & Bantul',
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                ),
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Batal', style: TextStyle(color: Colors.grey)),
          ),
          ElevatedButton(
            onPressed: () async {
              Navigator.pop(ctx);
              showDialog(
                context: context,
                barrierDismissible: false,
                builder: (c) => const Center(child: CircularProgressIndicator()),
              );

              try {
                final bannerIds = _unclaimedInstallations
                    .map((b) => b['id'].toString())
                    .toList();

                final res = await _apiService.bulkReimburseInstallations(
                  bannerIds: bannerIds,
                  notes: notesCtrl.text.trim(),
                );

                if (mounted) {
                  Navigator.pop(context);
                  showDialog(
                    context: context,
                    builder: (c) => AlertDialog(
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
                      title: const Row(
                        children: [
                          Icon(Icons.check_circle, color: Colors.green, size: 28),
                          SizedBox(width: 8),
                          Text('Pengajuan Terkirim'),
                        ],
                      ),
                      content: Text(
                        res['message'] ??
                            'Reimbursement ${_unclaimedInstallations.length} spanduk ($totalRpStr) berhasil diajukan ke Admin Kasir.',
                      ),
                      actions: [
                        ElevatedButton(
                          onPressed: () {
                            Navigator.pop(c);
                            _fetchAllData();
                          },
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFF047857),
                            foregroundColor: Colors.white,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(12),
                            ),
                          ),
                          child: const Text('OK'),
                        ),
                      ],
                    ),
                  );
                }
              } catch (e) {
                if (mounted) {
                  Navigator.pop(context);
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text(e.toString()),
                      backgroundColor: Colors.red,
                    ),
                  );
                }
              }
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF047857),
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            ),
            child: Text('Kirim Pengajuan ($totalRpStr)'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'Pasang Spanduk & Reimburse',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.bold,
            fontSize: 18,
            color: Colors.black87,
          ),
        ),
        backgroundColor: Colors.white,
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.black87),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _fetchAllData,
          )
        ],
        bottom: TabBar(
          controller: _tabController,
          labelColor: const Color(0xFF047857),
          unselectedLabelColor: Colors.grey,
          indicatorColor: const Color(0xFF047857),
          indicatorWeight: 3,
          tabs: [
            const Tab(
              icon: Icon(Icons.assignment),
              text: 'Tugas Pasang',
            ),
            Tab(
              icon: Badge(
                isLabelVisible: _unclaimedInstallations.isNotEmpty,
                label: Text('${_unclaimedInstallations.length}'),
                child: const Icon(Icons.receipt_long),
              ),
              text: 'Klaim Reimburse',
            ),
          ],
        ),
      ),
      body: _isInitialLoading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tabController,
              children: [
                // TAB 1: TUGAS PASANG
                RefreshIndicator(
                  onRefresh: () => _fetchBanners(reset: true),
                  child: Column(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(16),
                        color: Colors.white,
                        child: TextField(
                          controller: _searchController,
                          onChanged: _onSearchChanged,
                          decoration: InputDecoration(
                            hintText: 'Cari nama toko...',
                            prefixIcon: const Icon(Icons.search, color: Colors.grey),
                            suffixIcon: _searchQuery.isNotEmpty
                                ? IconButton(
                                    icon: const Icon(Icons.clear, color: Colors.grey),
                                    onPressed: () {
                                      _searchController.clear();
                                      _onSearchChanged('');
                                    },
                                  )
                                : null,
                            filled: true,
                            fillColor: Colors.grey[100],
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(12),
                              borderSide: BorderSide.none,
                            ),
                          ),
                        ),
                      ),
                      if (_isBannersLoading)
                        const LinearProgressIndicator(
                          minHeight: 2,
                          backgroundColor: Colors.transparent,
                          valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF047857)),
                        ),
                      if (_totalCount > 0)
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
                          color: Colors.white,
                          child: Text(
                            'Menampilkan ${_banners.length} dari $_totalCount spanduk siap pasang',
                            style: TextStyle(fontSize: 12, color: Colors.grey[500]),
                          ),
                        ),
                      Expanded(
                        child: _isBannersLoading && _banners.isEmpty
                            ? const Center(child: CircularProgressIndicator())
                            : _banners.isEmpty
                                ? ListView(
                                    physics: const AlwaysScrollableScrollPhysics(),
                                    children: [
                                      SizedBox(height: MediaQuery.of(context).size.height * 0.25),
                                      Center(
                                        child: Column(
                                          children: [
                                            Icon(
                                              _searchQuery.isNotEmpty
                                                  ? Icons.search_off_rounded
                                                  : Icons.inventory_2_outlined,
                                              size: 48,
                                              color: Colors.grey[400],
                                            ),
                                            const SizedBox(height: 12),
                                            Text(
                                              _searchQuery.isNotEmpty
                                                  ? 'Tidak ada spanduk yang sesuai dengan "$_searchQuery"'
                                                  : 'Tidak ada tugas spanduk yang perlu dipasang.\nSemua spanduk sudah terpasang.',
                                              textAlign: TextAlign.center,
                                              style: const TextStyle(
                                                color: Colors.grey,
                                                fontWeight: FontWeight.bold,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  )
                                : ListView.builder(
                                controller: _scrollController,
                                padding: const EdgeInsets.all(16),
                                itemCount: _banners.length + (_isLoadingMore ? 1 : 0),
                                itemBuilder: (context, index) {
                                  if (index == _banners.length) {
                                    return const Center(
                                      child: Padding(
                                        padding: EdgeInsets.symmetric(vertical: 24),
                                        child: CircularProgressIndicator(strokeWidth: 2),
                                      ),
                                    );
                                  }
                                  final banner = _banners[index];

                                  double? lat = banner['latitude'] != null
                                      ? double.tryParse(banner['latitude'].toString())
                                      : null;
                                  double? lng = banner['longitude'] != null
                                      ? double.tryParse(banner['longitude'].toString())
                                      : null;

                                  final shopName =
                                      banner['shop_name'] ?? banner['store_name'] ?? 'Toko';

                                  return Card(
                                    elevation: 0,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(16),
                                      side: BorderSide(color: Colors.grey.shade200),
                                    ),
                                    margin: const EdgeInsets.only(bottom: 16),
                                    child: Padding(
                                      padding: const EdgeInsets.all(16),
                                      child: Column(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Row(
                                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                            children: [
                                              Container(
                                                padding: const EdgeInsets.symmetric(
                                                    horizontal: 10, vertical: 4),
                                                decoration: BoxDecoration(
                                                  color: Colors.blue[50],
                                                  borderRadius: BorderRadius.circular(8),
                                                ),
                                                child: Text(
                                                  'SIAP DIPASANG',
                                                  style: TextStyle(
                                                    fontSize: 10,
                                                    fontWeight: FontWeight.bold,
                                                    color: Colors.blue[700],
                                                  ),
                                                ),
                                              ),
                                              if (lat != null && lng != null)
                                                GestureDetector(
                                                  onTap: () => _launchMaps(lat, lng),
                                                  child: Container(
                                                    padding: const EdgeInsets.all(6),
                                                    decoration: BoxDecoration(
                                                      color: Colors.orange[50],
                                                      borderRadius: BorderRadius.circular(8),
                                                    ),
                                                    child: Icon(Icons.directions,
                                                        size: 18, color: Colors.orange[700]),
                                                  ),
                                                ),
                                            ],
                                          ),
                                          const SizedBox(height: 12),
                                          Text(
                                            shopName,
                                            style: const TextStyle(
                                                fontSize: 16,
                                                fontWeight: FontWeight.bold,
                                                color: Colors.black87),
                                          ),
                                          const SizedBox(height: 4),
                                          Text(
                                            '${banner['type'] ?? banner['banner_type'] ?? 'Standard'} • ${_formatBannerDimension(banner['size_length'], banner['size_width'])}',
                                            style: TextStyle(
                                                fontSize: 13,
                                                fontWeight: FontWeight.w600,
                                                color: Colors.grey[600]),
                                          ),
                                          const SizedBox(height: 8),
                                          if (banner['address'] != null || banner['notes'] != null)
                                            Container(
                                              width: double.infinity,
                                              padding: const EdgeInsets.all(10),
                                              decoration: BoxDecoration(
                                                color: Colors.grey[50],
                                                borderRadius: BorderRadius.circular(10),
                                              ),
                                              child: Text(
                                                banner['address'] ?? banner['notes'] ?? '-',
                                                style: TextStyle(
                                                    fontSize: 12, color: Colors.grey[700]),
                                              ),
                                            ),
                                          const SizedBox(height: 14),
                                          SizedBox(
                                            width: double.infinity,
                                            child: ElevatedButton.icon(
                                              onPressed: () => _processInstall(banner),
                                              icon: const Icon(Icons.camera_alt, size: 18),
                                              label: const Text('Foto Bukti Pasang',
                                                  style: TextStyle(fontWeight: FontWeight.bold)),
                                              style: ElevatedButton.styleFrom(
                                                backgroundColor: const Color(0xFF047857),
                                                foregroundColor: Colors.white,
                                                padding: const EdgeInsets.symmetric(vertical: 12),
                                                shape: RoundedRectangleBorder(
                                                    borderRadius: BorderRadius.circular(12)),
                                                elevation: 0,
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                  );
                                },
                              ),
                      ),
                    ],
                  ),
                ),

                // TAB 2: KLAIM REIMBURSEMENT
                RefreshIndicator(
                  onRefresh: _fetchUnclaimedInstallations,
                  child: Column(
                    children: [
                      Container(
                        width: double.infinity,
                        margin: const EdgeInsets.all(16),
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          gradient: const LinearGradient(
                            colors: [Color(0xFF047857), Color(0xFF065F46)],
                            begin: Alignment.topLeft,
                            end: Alignment.bottomRight,
                          ),
                          borderRadius: BorderRadius.circular(24),
                          boxShadow: [
                            BoxShadow(
                              color: const Color(0xFF047857).withOpacity(0.25),
                              blurRadius: 12,
                              offset: const Offset(0, 4),
                            )
                          ],
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: Colors.white.withOpacity(0.2),
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: const Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Icon(Icons.sell_outlined, size: 12, color: Colors.white),
                                      SizedBox(width: 4),
                                      Text(
                                        'TARIF: Rp 1.500 / m²',
                                        style: TextStyle(
                                          color: Colors.white,
                                          fontSize: 11,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: Colors.amber[400],
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: Text(
                                    '${_unclaimedInstallations.length} Siap Klaim',
                                    style: const TextStyle(
                                      color: Colors.black87,
                                      fontSize: 11,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 14),
                            const Text(
                              'Total Estimasi Pendapatan Pemasangan',
                              style: TextStyle(
                                color: Colors.white70,
                                fontSize: 13,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              'Rp ${NumberFormat('#,###', 'id_ID').format(_totalEstimatedEarnings)}',
                              style: GoogleFonts.outfit(
                                color: Colors.white,
                                fontSize: 30,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(height: 12),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                              decoration: BoxDecoration(
                                color: Colors.black.withOpacity(0.15),
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Row(
                                children: [
                                  const Icon(Icons.square_foot, size: 16, color: Colors.white70),
                                  const SizedBox(width: 6),
                                  Text(
                                    'Total Luas: ${_totalUnclaimedArea.toStringAsFixed(1)} m²',
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 12,
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                  const Spacer(),
                                  Text(
                                    '${_unclaimedInstallations.length} Banner',
                                    style: const TextStyle(
                                      color: Colors.white70,
                                      fontSize: 12,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            if (_unclaimedInstallations.isNotEmpty) ...[
                              const SizedBox(height: 16),
                              SizedBox(
                                width: double.infinity,
                                child: ElevatedButton.icon(
                                  onPressed: _openBulkReimburseDialog,
                                  icon: const Icon(Icons.send_rounded, size: 18),
                                  label: Text(
                                    'Ajukan Klaim Reimburse (Rp ${NumberFormat('#,###', 'id_ID').format(_totalEstimatedEarnings)})',
                                    style: const TextStyle(
                                      fontWeight: FontWeight.bold,
                                      fontSize: 14,
                                    ),
                                  ),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: Colors.white,
                                    foregroundColor: const Color(0xFF047857),
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(14),
                                    ),
                                    padding: const EdgeInsets.symmetric(vertical: 14),
                                    elevation: 0,
                                  ),
                                ),
                              ),
                            ],
                          ],
                        ),
                      ),
                      Expanded(
                        child: _unclaimedInstallations.isEmpty
                            ? ListView(
                                physics: const AlwaysScrollableScrollPhysics(),
                                children: [
                                  SizedBox(height: MediaQuery.of(context).size.height * 0.12),
                                  Center(
                                    child: Padding(
                                      padding: const EdgeInsets.symmetric(horizontal: 32),
                                      child: Column(
                                        children: [
                                          Container(
                                            padding: const EdgeInsets.all(20),
                                            decoration: BoxDecoration(
                                              color: const Color(0xFF047857).withOpacity(0.08),
                                              shape: BoxShape.circle,
                                            ),
                                            child: const Icon(
                                              Icons.assignment_turned_in_outlined,
                                              size: 56,
                                              color: Color(0xFF047857),
                                            ),
                                          ),
                                          const SizedBox(height: 16),
                                          Text(
                                            'Belum Ada Spanduk Terpasang',
                                            style: GoogleFonts.outfit(
                                              fontSize: 17,
                                              fontWeight: FontWeight.bold,
                                              color: Colors.black87,
                                            ),
                                          ),
                                          const SizedBox(height: 8),
                                          Text(
                                            'Selesaikan tugas pemasangan spanduk di tab "Tugas Pasang".\nSetiap spanduk terpasang bernilai Rp 1.500 per m².',
                                            textAlign: TextAlign.center,
                                            style: TextStyle(
                                              fontSize: 13,
                                              color: Colors.grey[600],
                                              height: 1.4,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                  ),
                                ],
                              )
                            : ListView.builder(
                                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                                itemCount: _unclaimedInstallations.length,
                                itemBuilder: (context, index) {
                                  final item = _unclaimedInstallations[index];
                                  final shopName =
                                      item['store_name'] ?? item['shop_name'] ?? 'Toko';
                                  final installedAt = item['installed_at'] != null
                                      ? DateTime.tryParse(item['installed_at'].toString())
                                      : null;
                                  final dateStr = installedAt != null
                                      ? DateFormat('dd MMM yyyy, HH:mm').format(installedAt)
                                      : 'Hari ini';

                                  final earnings = _getBannerEarnings(item);
                                  final photoRaw = (item['photo_installed'] ?? item['documentation_image_url'])?.toString() ?? '';
                                  final photoUrl = AppConstants.buildImageUrl(photoRaw);

                                  return Card(
                                    elevation: 0,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(16),
                                      side: BorderSide(color: Colors.grey.shade200),
                                    ),
                                    margin: const EdgeInsets.only(bottom: 12),
                                    child: Padding(
                                      padding: const EdgeInsets.all(12),
                                      child: Row(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          GestureDetector(
                                            onTap: () {
                                              if (photoUrl.isNotEmpty) {
                                                _showImagePreview(photoUrl, 'Foto Bukti: $shopName');
                                              }
                                            },
                                            child: ClipRRect(
                                              borderRadius: BorderRadius.circular(12),
                                              child: photoUrl.isNotEmpty
                                                  ? Image.network(
                                                      photoUrl,
                                                      width: 68,
                                                      height: 68,
                                                      fit: BoxFit.cover,
                                                      errorBuilder: (_, __, ___) => Container(
                                                        width: 68,
                                                        height: 68,
                                                        color: Colors.grey[100],
                                                        child: const Icon(
                                                          Icons.image_not_supported_outlined,
                                                          size: 24,
                                                          color: Colors.grey,
                                                        ),
                                                      ),
                                                    )
                                                  : Container(
                                                      width: 68,
                                                      height: 68,
                                                      color: Colors.grey[100],
                                                      child: const Icon(
                                                        Icons.photo_outlined,
                                                        size: 24,
                                                        color: Colors.grey,
                                                      ),
                                                    ),
                                            ),
                                          ),
                                          const SizedBox(width: 12),
                                          Expanded(
                                            child: Column(
                                              crossAxisAlignment: CrossAxisAlignment.start,
                                              children: [
                                                Text(
                                                  shopName,
                                                  style: const TextStyle(
                                                    fontWeight: FontWeight.bold,
                                                    fontSize: 15,
                                                    color: Colors.black87,
                                                  ),
                                                  maxLines: 1,
                                                  overflow: TextOverflow.ellipsis,
                                                ),
                                                const SizedBox(height: 4),
                                                Text(
                                                  '${item['banner_type'] ?? 'Standard'} • ${_formatBannerDimension(item['size_length'], item['size_width'])}',
                                                  style: TextStyle(
                                                    fontSize: 12,
                                                    fontWeight: FontWeight.w600,
                                                    color: Colors.grey[600],
                                                  ),
                                                ),
                                                const SizedBox(height: 4),
                                                Row(
                                                  children: [
                                                    Icon(Icons.access_time, size: 12, color: Colors.grey[500]),
                                                    const SizedBox(width: 4),
                                                    Text(
                                                      dateStr,
                                                      style: TextStyle(
                                                        fontSize: 11,
                                                        color: Colors.grey[600],
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ],
                                            ),
                                          ),
                                          const SizedBox(width: 8),
                                          Column(
                                            crossAxisAlignment: CrossAxisAlignment.end,
                                            children: [
                                              Container(
                                                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                                decoration: BoxDecoration(
                                                  color: const Color(0xFFECFDF5),
                                                  borderRadius: BorderRadius.circular(8),
                                                  border: Border.all(color: const Color(0xFFA7F3D0)),
                                                ),
                                                child: Text(
                                                  '+Rp ${NumberFormat('#,###', 'id_ID').format(earnings)}',
                                                  style: const TextStyle(
                                                    fontSize: 12,
                                                    fontWeight: FontWeight.bold,
                                                    color: Color(0xFF047857),
                                                  ),
                                                ),
                                              ),
                                              const SizedBox(height: 6),
                                              Container(
                                                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                                decoration: BoxDecoration(
                                                  color: Colors.amber[50],
                                                  borderRadius: BorderRadius.circular(6),
                                                ),
                                                child: Text(
                                                  'Siap Klaim',
                                                  style: TextStyle(
                                                    fontSize: 10,
                                                    fontWeight: FontWeight.bold,
                                                    color: Colors.amber[800],
                                                  ),
                                                ),
                                              ),
                                            ],
                                          ),
                                        ],
                                      ),
                                    ),
                                  );
                                },
                              ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
    );
  }
}
