import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:image_picker/image_picker.dart';
import 'package:geolocator/geolocator.dart';
import 'package:hris_app/features/sales/data/services/banner_order_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';

class BannerInstallPage extends StatefulWidget {
  const BannerInstallPage({super.key});

  @override
  State<BannerInstallPage> createState() => _BannerInstallPageState();
}

class _BannerInstallPageState extends State<BannerInstallPage> {
  final _apiService = BannerOrderApiService(apiClient: di.sl<ApiClient>());
  bool _isLoading = true;
  List<dynamic> _banners = [];

  @override
  void initState() {
    super.initState();
    _fetchBanners();
  }

  Future<void> _fetchBanners() async {
    setState(() => _isLoading = true);
    try {
      final data = await _apiService.getBanners(status: 'ready_for_operator');
      setState(() {
        _banners = data;
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(e.toString()), backgroundColor: Colors.red));
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _installBanner(dynamic banner) async {
    // Show dialog to pick image and get GPS
    final picker = ImagePicker();
    final pickedFile =
        await picker.pickImage(maxWidth: 1024, maxHeight: 1024, source: ImageSource.camera, imageQuality: 70);
    if (pickedFile == null) return;

    // Get GPS
    bool serviceEnabled;
    LocationPermission permission;

    serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) {
      if (mounted)
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
            content: Text('GPS tidak aktif. Nyalakan GPS terlebih dahulu.')));
      return;
    }

    permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) {
        if (mounted)
          ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Akses lokasi ditolak.')));
        return;
      }
    }

    if (permission == LocationPermission.deniedForever) {
      if (mounted)
        ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Akses lokasi ditolak permanen.')));
      return;
    }

    if (mounted) {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (c) => const Center(child: CircularProgressIndicator()),
      );
    }

    try {
      Position position = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.high);
      final bytes = await pickedFile.readAsBytes();
      await _apiService.installBanner(
          banner['id'], pickedFile.path, position.latitude, position.longitude,
          imageBytes: bytes, filename: pickedFile.name);

      if (mounted) {
        Navigator.pop(context); // close loading
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
            content: Text('Instalasi Banner Berhasil!'),
            backgroundColor: Colors.green));
        _fetchBanners();
      }
    } catch (e) {
      if (mounted) {
        Navigator.pop(context); // close loading
        ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(e.toString()), backgroundColor: Colors.red));
      }
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
            Text('OPERATOR',
                style: GoogleFonts.outfit(
                    fontSize: 10,
                    color: Colors.indigo,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1)),
            Text('Pasang Spanduk',
                style: GoogleFonts.outfit(
                    fontSize: 17,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B))),
          ],
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _banners.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.check_circle_outline,
                          size: 64, color: Colors.grey),
                      const SizedBox(height: 16),
                      Text('Tidak ada spanduk untuk dipasang',
                          style: GoogleFonts.outfit(
                              fontSize: 16,
                              color: Colors.blueGrey,
                              fontWeight: FontWeight.w600)),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _fetchBanners,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _banners.length,
                    itemBuilder: (context, index) {
                      final banner = _banners[index];
                      return Card(
                        elevation: 2,
                        margin: const EdgeInsets.only(bottom: 16),
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16)),
                        child: Padding(
                          padding: const EdgeInsets.all(16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                        horizontal: 10, vertical: 4),
                                    decoration: BoxDecoration(
                                      color: Colors.blue.shade50,
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: Text('Siap Pasang',
                                        style: GoogleFonts.outfit(
                                            fontSize: 10,
                                            fontWeight: FontWeight.bold,
                                            color: Colors.blue.shade700)),
                                  ),
                                  const Spacer(),
                                  Text('Qty: ${banner['qty']}',
                                      style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: FontWeight.bold,
                                          color: Colors.blueGrey)),
                                ],
                              ),
                              const SizedBox(height: 12),
                              Text(
                                  banner['shop_name'] ??
                                      'Shop #${banner['shop_id']}',
                                  style: GoogleFonts.outfit(
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                      color: const Color(0xFF1E293B))),
                              const SizedBox(height: 4),
                              Text(
                                  'Ukuran: ${banner['size_length']}m x ${banner['size_width']}m',
                                  style: GoogleFonts.outfit(
                                      fontSize: 13, color: Colors.blueGrey)),
                              if (banner['notes'] != null &&
                                  banner['notes'].toString().isNotEmpty) ...[
                                const SizedBox(height: 8),
                                Container(
                                  padding: const EdgeInsets.all(8),
                                  decoration: BoxDecoration(
                                      color: Colors.amber.shade50,
                                      borderRadius: BorderRadius.circular(8)),
                                  child: Text('Catatan: ${banner['notes']}',
                                      style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          color: Colors.amber.shade900)),
                                ),
                              ],
                              const SizedBox(height: 16),
                              SizedBox(
                                width: double.infinity,
                                child: ElevatedButton.icon(
                                  onPressed: () => _installBanner(banner),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: Colors.indigo,
                                    padding: const EdgeInsets.symmetric(
                                        vertical: 12),
                                    shape: RoundedRectangleBorder(
                                        borderRadius:
                                            BorderRadius.circular(12)),
                                  ),
                                  icon: const Icon(Icons.camera_alt_rounded,
                                      color: Colors.white, size: 18),
                                  label: Text('AMBIL FOTO PEMASANGAN',
                                      style: GoogleFonts.outfit(
                                          fontWeight: FontWeight.bold,
                                          color: Colors.white)),
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
    );
  }
}
