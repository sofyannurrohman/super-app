import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:get_it/get_it.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/widgets/app_network_image.dart';
import 'package:image_picker/image_picker.dart';
import 'package:geolocator/geolocator.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:hris_app/features/sales/data/services/banner_order_api_service.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:intl/intl.dart';

import 'package:hris_app/features/sales/presentation/pages/operator_spanduk_schedule_page.dart';
import 'package:hris_app/features/sales/presentation/pages/operator_spanduk_plan_page.dart';

class OperatorSpandukDashboardPage extends StatefulWidget {
  const OperatorSpandukDashboardPage({super.key});

  @override
  State<OperatorSpandukDashboardPage> createState() =>
      _OperatorSpandukDashboardPageState();
}

class _OperatorSpandukDashboardPageState
    extends State<OperatorSpandukDashboardPage> {
  final _apiService = GetIt.I<BannerOrderApiService>();
  bool _isLoading = true;
  bool _isFetchingDashboard = false;

  List<dynamic> _printedBanners = [];
  List<dynamic> _waitingBanners = [];
  List<dynamic> _installedBanners = [];
  int _totalSelesaiCount = 0;
  double _totalKomisi = 0.0;
  String _operatorName = '';
  String _operatorBranchName = '';

  // Deep Emerald Palette
  static const Color primaryEmerald = Color(0xFF047857); // Deep rich emerald
  static const Color darkEmerald = Color(0xFF064E3B); // Forest emerald
  static const Color lightEmerald = Color(0xFFECFDF5); // Soft emerald tint

  String _extractName(Map<String, dynamic>? profile) {
    if (profile == null) return '';
    final firstName = profile['first_name']?.toString().trim() ?? '';
    final lastName = profile['last_name']?.toString().trim() ?? '';
    if (firstName.isNotEmpty) {
      return '$firstName $lastName'.trim();
    }
    return (profile['name'] ??
            profile['full_name'] ??
            profile['user']?['name'] ??
            profile['User']?['Name'] ??
            profile['username'] ??
            '')
        .toString()
        .trim();
  }

  @override
  void initState() {
    super.initState();
    final authState = context.read<AuthBloc>().state;
    if (authState is Authenticated) {
      final profile = authState.userProfile;
      final name = _extractName(profile);
      if (name.isNotEmpty) {
        _operatorName = name;
      }
      _operatorBranchName =
          (profile?['branch_name'] ??
                  profile?['branch']?['name'] ??
                  profile?['BranchName'] ??
                  '')
              .toString();
    }
    _fetchDashboardData();
  }

  Future<void> _fetchDashboardData() async {
    if (_isFetchingDashboard) return;
    _isFetchingDashboard = true;
    setState(() => _isLoading = true);
    try {
      String currentEmpId = '';
      String currentUserId = '';
      String currentUserName = '';
      final authState = context.read<AuthBloc>().state;
      Map<String, dynamic>? profile = authState is Authenticated
          ? authState.userProfile
          : null;

      if (profile == null) {
        try {
          final res = await _apiService.apiClient.client.get(
            'employees/profile',
          );
          final data = res.data is Map ? res.data['data'] : null;
          if (data is Map) profile = Map<String, dynamic>.from(data);
        } catch (_) {}
      }

      if (profile != null) {
        currentEmpId =
            (profile['employee_id'] ??
                    profile['employee']?['id'] ??
                    profile['id'])
                ?.toString() ??
            '';
        currentUserId =
            (profile['user_id'] ?? profile['user']?['id'] ?? profile['id'])
                ?.toString() ??
            '';
        final name = _extractName(profile);
        if (name.isNotEmpty) {
          _operatorName = name;
        }
        currentUserName = _operatorName;
        _operatorBranchName =
            (profile['branch_name'] ??
                    profile['branch']?['name'] ??
                    profile['BranchName'] ??
                    '')
                .toString();
      }

      final results = await _apiService.getAllBannersChunked();
      final performance = await _apiService.getOperatorPerformance(
        currentEmpId.isNotEmpty ? currentEmpId : currentUserId,
      );

      double komisi = 0;
      int totalSelesai = 0;

      if (performance != null) {
        komisi =
            double.tryParse(performance['total_komisi']?.toString() ?? '0') ??
            0.0;
        totalSelesai =
            int.tryParse(performance['banner_terpasang']?.toString() ?? '0') ??
            0;
      }

      final printed = <dynamic>[];
      final waiting = <dynamic>[];
      final installed = <dynamic>[];

      for (var b in results) {
        final status = b['status']?.toString().toLowerCase() ?? '';

        final bannerOpId =
            (b['operator_id'] ?? b['OperatorID'])?.toString().trim() ?? '';
        final bannerOpName =
            (b['operator_name'] ?? b['OperatorName'])
                ?.toString()
                .trim()
                .toLowerCase() ??
            '';
        final myName = currentUserName.toLowerCase();

        final bool isAssignedToMe =
            (bannerOpId.isNotEmpty &&
                (bannerOpId == currentEmpId || bannerOpId == currentUserId)) ||
            (myName.isNotEmpty &&
                bannerOpName.isNotEmpty &&
                (bannerOpName == myName ||
                    bannerOpName.contains(myName) ||
                    myName.contains(bannerOpName)));

        if (status == 'ready_for_operator' ||
            status == 'siap_pasang' ||
            status == 'siap pasang' ||
            status == 'sent_to_vendor' ||
            status == 'percetakan') {
          // Hanya spanduk yang belum diambil oleh operator lain
          if (bannerOpId.isEmpty || isAssignedToMe) {
            printed.add(b);
          }
        } else if (status == 'diterima_operator') {
          // HANYA spanduk yang secara eksplisit telah diambil / ditugaskan ke operator ini
          if (isAssignedToMe) {
            waiting.add(b);
          }
        } else if (status == 'installed' || status == 'terpasang') {
          if (isAssignedToMe) {
            installed.add(b);
          }
        }
      }

      if (mounted) {
        setState(() {
          _printedBanners = printed;
          _waitingBanners = waiting;
          _installedBanners = installed;
          _totalSelesaiCount = totalSelesai > installed.length
              ? totalSelesai
              : installed.length;
          _totalKomisi = komisi;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'Gagal memuat tugas spanduk. Tarik untuk mencoba lagi.',
            ),
          ),
        );
        setState(() => _isLoading = false);
      }
    } finally {
      _isFetchingDashboard = false;
    }
  }

  Future<void> _takeBanner(dynamic banner) async {
    if (!mounted) return;
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (c) =>
          const Center(child: CircularProgressIndicator(color: primaryEmerald)),
    );
    try {
      await _apiService.takeBanner(banner['id']);
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Spanduk berhasil diambil'),
            backgroundColor: primaryEmerald,
          ),
        );
        _fetchDashboardData();
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

  Future<void> _contactStore(String? phone) async {
    if (phone == null || phone.isEmpty || phone == '-') {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Nomor toko tidak tersedia')),
      );
      return;
    }
    String formattedPhone = phone.replaceAll(RegExp(r'[^0-9]'), '');
    if (formattedPhone.startsWith('0')) {
      formattedPhone = '62${formattedPhone.substring(1)}';
    }
    final url =
        'https://wa.me/$formattedPhone?text=Halo,%20kami%20dari%20tim%20pemasangan%20spanduk%20Karsa%20HR%20ingin%20konfirmasi%20terkait%20pemasangan%20spanduk%20di%20toko%20Bapak/Ibu.';
    final uri = Uri.parse(url);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Tidak dapat membuka WhatsApp')),
        );
      }
    }
  }

  Future<void> _contactSales(String? phone, String shopName) async {
    if (phone == null || phone.isEmpty || phone == '-') {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Nomor sales tidak tersedia')),
      );
      return;
    }
    String formattedPhone = phone.replaceAll(RegExp(r'[^0-9]'), '');
    if (formattedPhone.startsWith('0')) {
      formattedPhone = '62${formattedPhone.substring(1)}';
    }
    final encoded = Uri.encodeComponent(shopName);
    final url =
        'https://wa.me/$formattedPhone?text=Halo,%20saya%20operator%20pemasangan%20spanduk.%20Ingin%20konfirmasi%20terkait%20tugas%20pemasangan%20spanduk%20di%20toko%20$encoded.';
    final uri = Uri.parse(url);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Tidak dapat membuka WhatsApp')),
        );
      }
    }
  }

  void _showStorePhotoDialog(dynamic banner) {
    final shopName = banner['shop_name']?.toString().isNotEmpty == true
        ? banner['shop_name']
        : (banner['shop']?['name'] ??
              banner['Shop']?['Name'] ??
              'Toko Tanpa Nama');
    final shopAddress = banner['shop_address']?.toString().isNotEmpty == true
        ? banner['shop_address']
        : (banner['shop']?['address'] ?? banner['Shop']?['Address'] ?? '-');
    final rawShopPhoto =
        banner['shop_photo']?.toString() ??
        banner['shop']?['photo_url']?.toString() ??
        '';
    final shopPhoto = AppConstants.buildImageUrl(rawShopPhoto);
    final rawDesignLink = banner['design_link']?.toString() ?? '';
    final designLink = AppConstants.buildImageUrl(rawDesignLink);
    final shopPhone = banner['shop_phone']?.toString();
    final salesPhone = banner['sales_phone']?.toString();
    final lat = banner['latitude'] is num
        ? (banner['latitude'] as num).toDouble()
        : double.tryParse(banner['latitude']?.toString() ?? '') ?? 0.0;
    final lng = banner['longitude'] is num
        ? (banner['longitude'] as num).toDouble()
        : double.tryParse(banner['longitude']?.toString() ?? '') ?? 0.0;

    int selectedTab = 0;

    showDialog(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            final hasDesign = designLink.isNotEmpty;
            final currentImageUrl = selectedTab == 0 ? shopPhoto : designLink;

            return Dialog(
              backgroundColor: Colors.transparent,
              insetPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 24,
              ),
              child: Container(
                constraints: const BoxConstraints(
                  maxWidth: 480,
                  maxHeight: 620,
                ),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.25),
                      blurRadius: 20,
                      offset: const Offset(0, 10),
                    ),
                  ],
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    // Header
                    Container(
                      padding: const EdgeInsets.fromLTRB(16, 14, 12, 12),
                      decoration: const BoxDecoration(
                        color: darkEmerald,
                        borderRadius: BorderRadius.only(
                          topLeft: Radius.circular(20),
                          topRight: Radius.circular(20),
                        ),
                      ),
                      child: Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(6),
                            decoration: BoxDecoration(
                              color: Colors.white.withValues(alpha: 0.2),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: const Icon(
                              Icons.storefront_rounded,
                              color: Colors.white,
                              size: 18,
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  shopName,
                                  style: GoogleFonts.outfit(
                                    color: Colors.white,
                                    fontSize: 15,
                                    fontWeight: FontWeight.bold,
                                  ),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                Text(
                                  shopAddress,
                                  style: GoogleFonts.outfit(
                                    color: Colors.white.withValues(alpha: 0.8),
                                    fontSize: 11,
                                  ),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ],
                            ),
                          ),
                          IconButton(
                            icon: const Icon(
                              Icons.close_rounded,
                              color: Colors.white,
                              size: 22,
                            ),
                            onPressed: () => Navigator.pop(context),
                            padding: EdgeInsets.zero,
                            constraints: const BoxConstraints(),
                          ),
                        ],
                      ),
                    ),

                    // Tab Selector if Design Link Available
                    if (hasDesign)
                      Container(
                        color: const Color(0xFFF1F5F9),
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                        child: Row(
                          children: [
                            Expanded(
                              child: InkWell(
                                onTap: () =>
                                    setDialogState(() => selectedTab = 0),
                                borderRadius: BorderRadius.circular(8),
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 6,
                                  ),
                                  decoration: BoxDecoration(
                                    color: selectedTab == 0
                                        ? Colors.white
                                        : Colors.transparent,
                                    borderRadius: BorderRadius.circular(8),
                                    boxShadow: selectedTab == 0
                                        ? [
                                            BoxShadow(
                                              color: Colors.black.withValues(
                                                alpha: 0.05,
                                              ),
                                              blurRadius: 4,
                                            ),
                                          ]
                                        : null,
                                  ),
                                  child: Row(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.storefront_rounded,
                                        size: 14,
                                        color: selectedTab == 0
                                            ? primaryEmerald
                                            : Colors.grey[600],
                                      ),
                                      const SizedBox(width: 6),
                                      Text(
                                        'Foto Toko',
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: selectedTab == 0
                                              ? FontWeight.bold
                                              : FontWeight.w500,
                                          color: selectedTab == 0
                                              ? primaryEmerald
                                              : Colors.grey[600],
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 6),
                            Expanded(
                              child: InkWell(
                                onTap: () =>
                                    setDialogState(() => selectedTab = 1),
                                borderRadius: BorderRadius.circular(8),
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 6,
                                  ),
                                  decoration: BoxDecoration(
                                    color: selectedTab == 1
                                        ? Colors.white
                                        : Colors.transparent,
                                    borderRadius: BorderRadius.circular(8),
                                    boxShadow: selectedTab == 1
                                        ? [
                                            BoxShadow(
                                              color: Colors.black.withValues(
                                                alpha: 0.05,
                                              ),
                                              blurRadius: 4,
                                            ),
                                          ]
                                        : null,
                                  ),
                                  child: Row(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.image_outlined,
                                        size: 14,
                                        color: selectedTab == 1
                                            ? primaryEmerald
                                            : Colors.grey[600],
                                      ),
                                      const SizedBox(width: 6),
                                      Text(
                                        'Desain Spanduk',
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: selectedTab == 1
                                              ? FontWeight.bold
                                              : FontWeight.w500,
                                          color: selectedTab == 1
                                              ? primaryEmerald
                                              : Colors.grey[600],
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),

                    // Main Image Area
                    Expanded(
                      child: Container(
                        color: const Color(0xFF0F172A),
                        width: double.infinity,
                        child: currentImageUrl.isNotEmpty
                            ? InteractiveViewer(
                                minScale: 0.8,
                                maxScale: 4.0,
                                child: AppNetworkImage(
                                  imageUrl: currentImageUrl,
                                  fit: BoxFit.contain,
                                  placeholder: const Center(
                                    child: CircularProgressIndicator(
                                      color: Colors.white,
                                    ),
                                  ),
                                  errorWidget: Center(
                                    child: Column(
                                      mainAxisAlignment:
                                          MainAxisAlignment.center,
                                      children: [
                                        Icon(
                                          Icons.broken_image_rounded,
                                          size: 48,
                                          color: Colors.grey[500],
                                        ),
                                        const SizedBox(height: 8),
                                        Text(
                                          'Gagal memuat gambar',
                                          style: GoogleFonts.outfit(
                                            color: Colors.grey[400],
                                            fontSize: 12,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  memCacheWidth: 1000,
                                ),
                              )
                            : Center(
                                child: Padding(
                                  padding: const EdgeInsets.all(24),
                                  child: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Container(
                                        padding: const EdgeInsets.all(16),
                                        decoration: BoxDecoration(
                                          color: Colors.white.withValues(
                                            alpha: 0.1,
                                          ),
                                          shape: BoxShape.circle,
                                        ),
                                        child: const Icon(
                                          Icons.add_a_photo_outlined,
                                          size: 40,
                                          color: Colors.white70,
                                        ),
                                      ),
                                      const SizedBox(height: 12),
                                      Text(
                                        selectedTab == 0
                                            ? 'Foto profil toko belum diunggah'
                                            : 'Preview desain belum tersedia',
                                        style: GoogleFonts.outfit(
                                          color: Colors.white,
                                          fontSize: 13,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                      const SizedBox(height: 4),
                                      Text(
                                        'Hubungi Sales PIC untuk konfirmasi bentuk fisik toko sebelum pemasangan.',
                                        textAlign: TextAlign.center,
                                        style: GoogleFonts.outfit(
                                          color: Colors.white70,
                                          fontSize: 11,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                      ),
                    ),

                    // Action Footer Bar
                    Container(
                      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
                      decoration: const BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.only(
                          bottomLeft: Radius.circular(20),
                          bottomRight: Radius.circular(20),
                        ),
                        border: Border(
                          top: BorderSide(color: Color(0xFFE2E8F0)),
                        ),
                      ),
                      child: Row(
                        children: [
                          if (shopPhone != null &&
                              shopPhone.isNotEmpty &&
                              shopPhone != '-') ...[
                            Expanded(
                              child: OutlinedButton.icon(
                                onPressed: () {
                                  Navigator.pop(context);
                                  _contactStore(shopPhone);
                                },
                                icon: const Icon(
                                  Icons.storefront_rounded,
                                  size: 13,
                                ),
                                label: Text(
                                  'Chat Toko',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                                style: OutlinedButton.styleFrom(
                                  foregroundColor: primaryEmerald,
                                  side: const BorderSide(color: primaryEmerald),
                                  minimumSize: const Size(0, 34),
                                  tapTargetSize:
                                      MaterialTapTargetSize.shrinkWrap,
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 6,
                                    horizontal: 8,
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 6),
                          ],
                          Expanded(
                            child: OutlinedButton.icon(
                              onPressed: () {
                                Navigator.pop(context);
                                _contactSales(salesPhone, shopName);
                              },
                              icon: const Icon(Icons.person_rounded, size: 13),
                              label: Text(
                                'Chat Sales',
                                style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              style: OutlinedButton.styleFrom(
                                foregroundColor: Colors.blue[700],
                                side: BorderSide(color: Colors.blue[600]!),
                                minimumSize: const Size(0, 34),
                                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                padding: const EdgeInsets.symmetric(
                                  vertical: 6,
                                  horizontal: 8,
                                ),
                              ),
                            ),
                          ),
                          if (lat != 0.0 && lng != 0.0) ...[
                            const SizedBox(width: 6),
                            IconButton(
                              onPressed: () {
                                final mapUrl =
                                    'https://www.google.com/maps/search/?api=1&query=$lat,$lng';
                                launchUrl(
                                  Uri.parse(mapUrl),
                                  mode: LaunchMode.externalApplication,
                                );
                              },
                              icon: const Icon(
                                Icons.map_rounded,
                                color: Colors.indigo,
                                size: 18,
                              ),
                              tooltip: 'Buka di Google Maps',
                              style: IconButton.styleFrom(
                                backgroundColor: const Color(0xFFEEF2FF),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                padding: const EdgeInsets.all(8),
                                minimumSize: const Size(34, 34),
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildSmartHeroActionCard() {
    final hasActiveTasks = _waitingBanners.isNotEmpty;
    final readyCount = _printedBanners.length;
    final activeCount = _waitingBanners.length;

    if (hasActiveTasks) {
      return Container(
        margin: const EdgeInsets.symmetric(horizontal: 16),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: const Color(0xFFFDE68A), width: 1.5),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFFD97706).withValues(alpha: 0.08),
              blurRadius: 12,
              offset: const Offset(0, 4),
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
                    color: const Color(0xFFFEF3C7),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Icon(
                    Icons.pending_actions_rounded,
                    color: Color(0xFFD97706),
                    size: 22,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '$activeCount Spanduk Menunggu Dipasang',
                        style: GoogleFonts.outfit(
                          fontSize: 15,
                          fontWeight: FontWeight.bold,
                          color: const Color(0xFF0F172A),
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        'Lanjutkan rute dan pasang spanduk di lokasi tujuan',
                        style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: const Color(0xFF64748B),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 14),
            ElevatedButton.icon(
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const OperatorSpandukSchedulePage(),
                  ),
                ).then((_) => _fetchDashboardData());
              },
              icon: const Icon(Icons.route_rounded, size: 18),
              label: Text(
                'Buka Rute & Pasang Sekarang',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.bold,
                  fontSize: 13,
                ),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: primaryEmerald,
                foregroundColor: Colors.white,
                minimumSize: const Size(double.infinity, 44),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                elevation: 0,
              ),
            ),
          ],
        ),
      );
    } else {
      return Container(
        margin: const EdgeInsets.symmetric(horizontal: 16),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: const Color(0xFFA7F3D0), width: 1.5),
          boxShadow: [
            BoxShadow(
              color: primaryEmerald.withValues(alpha: 0.08),
              blurRadius: 12,
              offset: const Offset(0, 4),
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
                    color: lightEmerald,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Icon(
                    Icons.assignment_add,
                    color: primaryEmerald,
                    size: 22,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Siap Menerima Tugas Baru',
                        style: GoogleFonts.outfit(
                          fontSize: 15,
                          fontWeight: FontWeight.bold,
                          color: const Color(0xFF0F172A),
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        readyCount > 0
                            ? 'Tersedia $readyCount spanduk siap pasang di wilayah Anda'
                            : 'Belum ada spanduk siap pasang saat ini',
                        style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: const Color(0xFF64748B),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 14),
            ElevatedButton.icon(
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const OperatorSpandukPlanPage(),
                  ),
                ).then((_) => _fetchDashboardData());
              },
              icon: const Icon(Icons.add_location_alt_rounded, size: 18),
              label: Text(
                'Pilih & Ambil Tugas Pasang',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.bold,
                  fontSize: 13,
                ),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: primaryEmerald,
                foregroundColor: Colors.white,
                minimumSize: const Size(double.infinity, 44),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                elevation: 0,
              ),
            ),
          ],
        ),
      );
    }
  }

  Widget _buildNavHubCard({
    required IconData icon,
    required Color iconColor,
    required Color iconBgColor,
    required String title,
    required String subtitle,
    required String badgeText,
    required Color badgeColor,
    required Color badgeTextColor,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFFE2E8F0)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.02),
              blurRadius: 6,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: iconBgColor,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: iconColor, size: 22),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: GoogleFonts.outfit(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: const Color(0xFF0F172A),
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    subtitle,
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: const Color(0xFF64748B),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: badgeColor,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                badgeText,
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: badgeTextColor,
                ),
              ),
            ),
            const SizedBox(width: 4),
            const Icon(
              Icons.chevron_right_rounded,
              size: 20,
              color: Color(0xFF94A3B8),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final currencyFormatter = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp',
      decimalDigits: 0,
    );

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'Dashboard Operator',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.bold,
            fontSize: 16,
            color: Colors.white,
          ),
        ),
        backgroundColor: darkEmerald,
        elevation: 0,
        centerTitle: false,
        iconTheme: const IconThemeData(color: Colors.white),
        actions: [
          IconButton(
            onPressed: _fetchDashboardData,
            icon: const Icon(Icons.refresh_rounded, size: 20),
            tooltip: 'Segarkan Data',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(color: primaryEmerald),
            )
          : RefreshIndicator(
              color: primaryEmerald,
              onRefresh: _fetchDashboardData,
              child: SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // 1. Header Profile & Earnings Card
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
                      decoration: const BoxDecoration(
                        gradient: LinearGradient(
                          colors: [darkEmerald, primaryEmerald],
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                        ),
                        borderRadius: BorderRadius.only(
                          bottomLeft: Radius.circular(24),
                          bottomRight: Radius.circular(24),
                        ),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          // Profile Header
                          Row(
                            children: [
                              Container(
                                padding: const EdgeInsets.all(8),
                                decoration: BoxDecoration(
                                  color: Colors.white.withValues(alpha: 0.15),
                                  shape: BoxShape.circle,
                                ),
                                child: const Icon(
                                  Icons.engineering_rounded,
                                  color: Colors.white,
                                  size: 20,
                                ),
                              ),
                              const SizedBox(width: 10),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      _operatorName.isNotEmpty
                                          ? _operatorName
                                          : 'Operator',
                                      style: GoogleFonts.outfit(
                                        color: Colors.white,
                                        fontSize: 15,
                                        fontWeight: FontWeight.bold,
                                      ),
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                    Text(
                                      _operatorBranchName.isNotEmpty
                                          ? 'Cabang $_operatorBranchName'
                                          : 'Operator Pemasangan Spanduk',
                                      style: GoogleFonts.outfit(
                                        color: Colors.white.withValues(
                                          alpha: 0.85,
                                        ),
                                        fontSize: 11,
                                      ),
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
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
                                  color: Colors.white.withValues(alpha: 0.18),
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Container(
                                      width: 6,
                                      height: 6,
                                      decoration: const BoxDecoration(
                                        color: Color(0xFF34D399),
                                        shape: BoxShape.circle,
                                      ),
                                    ),
                                    const SizedBox(width: 5),
                                    Text(
                                      'Siap Kerja',
                                      style: GoogleFonts.outfit(
                                        fontSize: 10,
                                        color: Colors.white,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 14),

                          // Earnings Summary Box
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 12,
                            ),
                            decoration: BoxDecoration(
                              color: Colors.white.withValues(alpha: 0.12),
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(
                                color: Colors.white.withValues(alpha: 0.2),
                              ),
                            ),
                            child: Row(
                              children: [
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        'Estimasi Komisi',
                                        style: GoogleFonts.outfit(
                                          color: Colors.white.withValues(
                                            alpha: 0.8,
                                          ),
                                          fontSize: 11,
                                        ),
                                      ),
                                      const SizedBox(height: 2),
                                      Text(
                                        currencyFormatter.format(_totalKomisi),
                                        style: GoogleFonts.outfit(
                                          color: Colors.white,
                                          fontSize: 20,
                                          fontWeight: FontWeight.w800,
                                          letterSpacing: -0.5,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                Container(
                                  height: 36,
                                  width: 1,
                                  color: Colors.white.withValues(alpha: 0.2),
                                ),
                                const SizedBox(width: 16),
                                Column(
                                  crossAxisAlignment: CrossAxisAlignment.end,
                                  children: [
                                    Text(
                                      'Total Terpasang',
                                      style: GoogleFonts.outfit(
                                        color: Colors.white.withValues(
                                          alpha: 0.8,
                                        ),
                                        fontSize: 11,
                                      ),
                                    ),
                                    const SizedBox(height: 2),
                                    Text(
                                      '$_totalSelesaiCount Spanduk',
                                      style: GoogleFonts.outfit(
                                        color: const Color(0xFF6EE7B7),
                                        fontSize: 14,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),

                    const SizedBox(height: 14),

                    // 2. Smart Hero Work Status (Dynamic Action Guide)
                    _buildSmartHeroActionCard(),

                    const SizedBox(height: 16),

                    // 3. 3 Menu Utama Alur Kerja
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'ALUR KERJA OPERATOR',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                              color: const Color(0xFF64748B),
                              letterSpacing: 1.1,
                            ),
                          ),
                          const SizedBox(height: 8),

                          // Menu 1: Ambil Tugas Baru
                          _buildNavHubCard(
                            icon: Icons.add_location_alt_rounded,
                            iconColor: Colors.blue[700]!,
                            iconBgColor: const Color(0xFFEFF6FF),
                            title: 'Pilih Tugas Spanduk',
                            subtitle:
                                'Pilih tugas siap pasang & rencanakan area',
                            badgeText: '${_printedBanners.length} Siap',
                            badgeColor: const Color(0xFFDBEAFE),
                            badgeTextColor: Colors.blue[800]!,
                            onTap: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (_) =>
                                      const OperatorSpandukPlanPage(),
                                ),
                              ).then((_) => _fetchDashboardData());
                            },
                          ),
                          const SizedBox(height: 8),

                          // Menu 2: Rute & Jadwal Hari Ini
                          _buildNavHubCard(
                            icon: Icons.route_rounded,
                            iconColor: const Color(0xFFD97706),
                            iconBgColor: const Color(0xFFFEF3C7),
                            title: 'Rute & Jadwal Hari Ini',
                            subtitle:
                                'Navigasi Maps & upload foto bukti pasang',
                            badgeText: '${_waitingBanners.length} Proses',
                            badgeColor: const Color(0xFFFDE68A),
                            badgeTextColor: const Color(0xFF92400E),
                            onTap: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (_) =>
                                      const OperatorSpandukSchedulePage(),
                                ),
                              ).then((_) => _fetchDashboardData());
                            },
                          ),
                          const SizedBox(height: 8),

                          // Menu 3: Riwayat & Komisi
                          _buildNavHubCard(
                            icon: Icons.task_alt_rounded,
                            iconColor: primaryEmerald,
                            iconBgColor: lightEmerald,
                            title: 'Riwayat & Komisi',
                            subtitle:
                                'Daftar spanduk selesai & rincian hasil kerja',
                            badgeText: '$_totalSelesaiCount Selesai',
                            badgeColor: const Color(0xFFA7F3D0),
                            badgeTextColor: darkEmerald,
                            onTap: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (_) => const OperatorSpandukListPage(
                                    initialTab: 2,
                                  ),
                                ),
                              ).then((_) => _fetchDashboardData());
                            },
                          ),
                        ],
                      ),
                    ),

                    const SizedBox(height: 18),

                    // 4. Section Tugas Terdekat / Aktif atau Riwayat Terkini
                    if (_waitingBanners.isNotEmpty) ...[
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              'TUGAS SEDANG DIKERJAKAN',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                                color: const Color(0xFF64748B),
                                letterSpacing: 1.1,
                              ),
                            ),
                            InkWell(
                              onTap: () {
                                Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (_) =>
                                        const OperatorSpandukSchedulePage(),
                                  ),
                                ).then((_) => _fetchDashboardData());
                              },
                              child: Text(
                                'Buka Jadwal',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  fontWeight: FontWeight.bold,
                                  color: primaryEmerald,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 6),
                      ListView.builder(
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        itemCount: _waitingBanners.length > 2
                            ? 2
                            : _waitingBanners.length,
                        itemBuilder: (context, index) {
                          final banner = _waitingBanners[index];
                          final shopName =
                              banner['shop_name']
                                      ?.toString()
                                      .trim()
                                      .isNotEmpty ==
                                  true
                              ? banner['shop_name']
                              : (banner['shop']?['name'] ??
                                    banner['Shop']?['Name'] ??
                                    'Toko Tanpa Nama');
                          final shopAddress =
                              banner['shop_address']
                                      ?.toString()
                                      .trim()
                                      .isNotEmpty ==
                                  true
                              ? banner['shop_address']
                              : (banner['shop']?['address'] ??
                                    banner['Shop']?['Address'] ??
                                    '-');
                          final rawShopPhoto =
                              banner['shop_photo']?.toString() ??
                              banner['shop']?['photo_url']?.toString() ??
                              '';

                          return Container(
                            margin: const EdgeInsets.only(bottom: 8),
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(14),
                              border: Border.all(
                                color: const Color(0xFFE2E8F0),
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withValues(alpha: 0.02),
                                  blurRadius: 4,
                                  offset: const Offset(0, 2),
                                ),
                              ],
                            ),
                            child: Row(
                              children: [
                                InkWell(
                                  onTap: () => _showStorePhotoDialog(banner),
                                  borderRadius: BorderRadius.circular(10),
                                  child: Container(
                                    width: 44,
                                    height: 44,
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFF1F5F9),
                                      borderRadius: BorderRadius.circular(10),
                                      border: Border.all(
                                        color: const Color(0xFFCBD5E1),
                                      ),
                                    ),
                                    clipBehavior: Clip.antiAlias,
                                    child: rawShopPhoto.isNotEmpty
                                        ? AppNetworkImage(
                                            imageUrl: rawShopPhoto,
                                            fit: BoxFit.cover,
                                            placeholder: const Center(
                                              child: Icon(
                                                Icons.storefront_rounded,
                                                size: 20,
                                                color: Color(0xFF94A3B8),
                                              ),
                                            ),
                                            errorWidget: const Center(
                                              child: Icon(
                                                Icons.storefront_rounded,
                                                size: 20,
                                                color: Color(0xFF94A3B8),
                                              ),
                                            ),
                                            memCacheWidth: 200,
                                          )
                                        : const Center(
                                            child: Icon(
                                              Icons.storefront_rounded,
                                              size: 22,
                                              color: primaryEmerald,
                                            ),
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
                                        shopName,
                                        style: GoogleFonts.outfit(
                                          fontSize: 13,
                                          fontWeight: FontWeight.bold,
                                          color: const Color(0xFF0F172A),
                                        ),
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                      const SizedBox(height: 2),
                                      Text(
                                        shopAddress,
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          color: const Color(0xFF64748B),
                                        ),
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 8),
                                ElevatedButton(
                                  onPressed: () {
                                    Navigator.push(
                                      context,
                                      MaterialPageRoute(
                                        builder: (_) =>
                                            const OperatorSpandukSchedulePage(),
                                      ),
                                    ).then((_) => _fetchDashboardData());
                                  },
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: primaryEmerald,
                                    foregroundColor: Colors.white,
                                    minimumSize: const Size(0, 34),
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 10,
                                    ),
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    elevation: 0,
                                  ),
                                  child: Text(
                                    'Pasang',
                                    style: GoogleFonts.outfit(
                                      fontSize: 11,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          );
                        },
                      ),
                    ] else if (_printedBanners.isNotEmpty) ...[
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              'SPANDUK SIAP DIAMBIL',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                                color: const Color(0xFF64748B),
                                letterSpacing: 1.1,
                              ),
                            ),
                            InkWell(
                              onTap: () {
                                Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (_) =>
                                        const OperatorSpandukPlanPage(),
                                  ),
                                ).then((_) => _fetchDashboardData());
                              },
                              child: Text(
                                'Pilih Semua',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  fontWeight: FontWeight.bold,
                                  color: primaryEmerald,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 6),
                      ListView.builder(
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        itemCount: _printedBanners.length > 2
                            ? 2
                            : _printedBanners.length,
                        itemBuilder: (context, index) {
                          final banner = _printedBanners[index];
                          final shopName =
                              banner['shop_name']
                                      ?.toString()
                                      .trim()
                                      .isNotEmpty ==
                                  true
                              ? banner['shop_name']
                              : (banner['shop']?['name'] ??
                                    banner['Shop']?['Name'] ??
                                    'Toko Tanpa Nama');
                          final shopAddress =
                              banner['shop_address']
                                      ?.toString()
                                      .trim()
                                      .isNotEmpty ==
                                  true
                              ? banner['shop_address']
                              : (banner['shop']?['address'] ??
                                    banner['Shop']?['Address'] ??
                                    '-');
                          final rawShopPhoto =
                              banner['shop_photo']?.toString() ??
                              banner['shop']?['photo_url']?.toString() ??
                              '';
                          final reward =
                              double.tryParse(
                                banner['operator_reward']?.toString() ??
                                    '15000',
                              ) ??
                              15000.0;

                          return Container(
                            margin: const EdgeInsets.only(bottom: 8),
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(14),
                              border: Border.all(
                                color: const Color(0xFFE2E8F0),
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withValues(alpha: 0.02),
                                  blurRadius: 4,
                                  offset: const Offset(0, 2),
                                ),
                              ],
                            ),
                            child: Row(
                              children: [
                                InkWell(
                                  onTap: () => _showStorePhotoDialog(banner),
                                  borderRadius: BorderRadius.circular(10),
                                  child: Container(
                                    width: 44,
                                    height: 44,
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFF1F5F9),
                                      borderRadius: BorderRadius.circular(10),
                                      border: Border.all(
                                        color: const Color(0xFFCBD5E1),
                                      ),
                                    ),
                                    clipBehavior: Clip.antiAlias,
                                    child: rawShopPhoto.isNotEmpty
                                        ? AppNetworkImage(
                                            imageUrl: rawShopPhoto,
                                            fit: BoxFit.cover,
                                            placeholder: const Center(
                                              child: Icon(
                                                Icons.storefront_rounded,
                                                size: 20,
                                                color: Color(0xFF94A3B8),
                                              ),
                                            ),
                                            errorWidget: const Center(
                                              child: Icon(
                                                Icons.storefront_rounded,
                                                size: 20,
                                                color: Color(0xFF94A3B8),
                                              ),
                                            ),
                                            memCacheWidth: 200,
                                          )
                                        : const Center(
                                            child: Icon(
                                              Icons.storefront_rounded,
                                              size: 22,
                                              color: primaryEmerald,
                                            ),
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
                                        shopName,
                                        style: GoogleFonts.outfit(
                                          fontSize: 13,
                                          fontWeight: FontWeight.bold,
                                          color: const Color(0xFF0F172A),
                                        ),
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                      const SizedBox(height: 2),
                                      Text(
                                        shopAddress,
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          color: const Color(0xFF64748B),
                                        ),
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                      const SizedBox(height: 2),
                                      Text(
                                        'Komisi: ${currencyFormatter.format(reward)}',
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          fontWeight: FontWeight.bold,
                                          color: primaryEmerald,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 8),
                                ElevatedButton(
                                  onPressed: () => _takeBanner(banner),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: primaryEmerald,
                                    foregroundColor: Colors.white,
                                    minimumSize: const Size(0, 34),
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 12,
                                    ),
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    elevation: 0,
                                  ),
                                  child: Text(
                                    'Ambil',
                                    style: GoogleFonts.outfit(
                                      fontSize: 11,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          );
                        },
                      ),
                    ] else ...[
                      Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 8,
                        ),
                        child: Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(14),
                            border: Border.all(color: const Color(0xFFE2E8F0)),
                          ),
                          child: Column(
                            children: [
                              Icon(
                                Icons.check_circle_outline_rounded,
                                size: 32,
                                color: Colors.green[400],
                              ),
                              const SizedBox(height: 6),
                              Text(
                                'Semua tugas spanduk sudah selesai!',
                                style: GoogleFonts.outfit(
                                  fontSize: 13,
                                  fontWeight: FontWeight.bold,
                                  color: const Color(0xFF334155),
                                ),
                              ),
                              const SizedBox(height: 2),
                              Text(
                                'Tidak ada tugas yang menunggu saat ini.',
                                style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  color: const Color(0xFF64748B),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ],

                    // 5. Riwayat Terpasang Terkini
                    if (_installedBanners.isNotEmpty) ...[
                      const SizedBox(height: 10),
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              'RIWAYAT TERAKHIR',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                                color: const Color(0xFF64748B),
                                letterSpacing: 1.1,
                              ),
                            ),
                            InkWell(
                              onTap: () {
                                Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (_) =>
                                        const OperatorSpandukListPage(
                                          initialTab: 2,
                                        ),
                                  ),
                                ).then((_) => _fetchDashboardData());
                              },
                              child: Text(
                                'Lihat Semua',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  fontWeight: FontWeight.bold,
                                  color: primaryEmerald,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 6),
                      ListView.builder(
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        itemCount: _installedBanners.length > 2
                            ? 2
                            : _installedBanners.length,
                        itemBuilder: (context, index) {
                          final banner = _installedBanners[index];
                          final shopName =
                              banner['shop_name']
                                      ?.toString()
                                      .trim()
                                      .isNotEmpty ==
                                  true
                              ? banner['shop_name']
                              : (banner['shop']?['name'] ??
                                    banner['Shop']?['Name'] ??
                                    'Toko Tanpa Nama');
                          final shopAddress =
                              banner['shop_address']
                                      ?.toString()
                                      .trim()
                                      .isNotEmpty ==
                                  true
                              ? banner['shop_address']
                              : (banner['shop']?['address'] ??
                                    banner['Shop']?['Address'] ??
                                    '-');
                          final rawShopPhoto =
                              banner['shop_photo']?.toString() ??
                              banner['shop']?['photo_url']?.toString() ??
                              '';

                          return Container(
                            margin: const EdgeInsets.only(bottom: 8),
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 10,
                            ),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(
                                color: const Color(0xFFE2E8F0),
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withValues(alpha: 0.02),
                                  blurRadius: 4,
                                  offset: const Offset(0, 2),
                                ),
                              ],
                            ),
                            child: Row(
                              children: [
                                InkWell(
                                  onTap: () => _showStorePhotoDialog(banner),
                                  borderRadius: BorderRadius.circular(8),
                                  child: Container(
                                    width: 38,
                                    height: 38,
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFF1F5F9),
                                      borderRadius: BorderRadius.circular(8),
                                      border: Border.all(
                                        color: const Color(0xFFCBD5E1),
                                      ),
                                    ),
                                    clipBehavior: Clip.antiAlias,
                                    child: rawShopPhoto.isNotEmpty
                                        ? AppNetworkImage(
                                            imageUrl: rawShopPhoto,
                                            fit: BoxFit.cover,
                                            placeholder: const Center(
                                              child: Icon(
                                                Icons.storefront_rounded,
                                                size: 18,
                                                color: Color(0xFF94A3B8),
                                              ),
                                            ),
                                            errorWidget: const Center(
                                              child: Icon(
                                                Icons.storefront_rounded,
                                                size: 18,
                                                color: Color(0xFF94A3B8),
                                              ),
                                            ),
                                            memCacheWidth: 200,
                                          )
                                        : const Center(
                                            child: Icon(
                                              Icons.storefront_rounded,
                                              size: 20,
                                              color: primaryEmerald,
                                            ),
                                          ),
                                  ),
                                ),
                                const SizedBox(width: 10),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        shopName,
                                        style: GoogleFonts.outfit(
                                          fontSize: 13,
                                          fontWeight: FontWeight.bold,
                                          color: const Color(0xFF0F172A),
                                        ),
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                      const SizedBox(height: 2),
                                      Text(
                                        shopAddress,
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          color: const Color(0xFF64748B),
                                        ),
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 3,
                                  ),
                                  decoration: BoxDecoration(
                                    color: lightEmerald,
                                    borderRadius: BorderRadius.circular(6),
                                  ),
                                  child: Text(
                                    'Terpasang',
                                    style: GoogleFonts.outfit(
                                      fontSize: 10,
                                      fontWeight: FontWeight.bold,
                                      color: primaryEmerald,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          );
                        },
                      ),
                    ],

                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ),
    );
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Operator Spanduk List Page (Detail Lists)
// ─────────────────────────────────────────────────────────────────────────────
class OperatorSpandukListPage extends StatefulWidget {
  final int initialTab;
  const OperatorSpandukListPage({super.key, required this.initialTab});

  @override
  State<OperatorSpandukListPage> createState() =>
      _OperatorSpandukListPageState();
}

class _OperatorSpandukListPageState extends State<OperatorSpandukListPage>
    with SingleTickerProviderStateMixin {
  final _apiService = GetIt.I<BannerOrderApiService>();
  late TabController _tabController;

  bool _isLoading = true;
  List<dynamic> _printedBanners = [];
  List<dynamic> _waitingBanners = [];
  List<dynamic> _installedBanners = [];
  String _searchQuery = '';

  static const Color primaryEmerald = Color(0xFF047857);
  static const Color darkEmerald = Color(0xFF064E3B);
  static const Color lightEmerald = Color(0xFFECFDF5);

  @override
  void initState() {
    super.initState();
    _tabController = TabController(
      length: 3,
      vsync: this,
      initialIndex: widget.initialTab,
    );
    _fetchBanners();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _fetchBanners() async {
    setState(() => _isLoading = true);
    try {
      final results = await _apiService.getBanners();
      setState(() {
        _printedBanners = results.where((b) {
          final status = b['status']?.toString().toLowerCase();
          return status == 'ready_for_operator' ||
              status == 'sent_to_vendor' ||
              status == 'percetakan';
        }).toList();

        _waitingBanners = results.where((b) {
          final status = b['status']?.toString().toLowerCase();
          return status == 'diterima_operator';
        }).toList();

        _installedBanners = results.where((b) {
          final status = b['status']?.toString().toLowerCase();
          return status == 'installed' || status == 'terpasang';
        }).toList();
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
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

  Future<void> _contactStore(String? phone) async {
    if (phone == null || phone.isEmpty || phone == '-') {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Nomor toko tidak tersedia')),
      );
      return;
    }
    String formattedPhone = phone;
    if (formattedPhone.startsWith('0')) {
      formattedPhone = '62${formattedPhone.substring(1)}';
    }
    final url =
        'https://wa.me/$formattedPhone?text=Halo,%20kami%20dari%20tim%20pemasangan%20spanduk%20Karsa%20HR%20ingin%20konfirmasi%20terkait%20pemasangan%20spanduk%20di%20toko%20Bapak/Ibu.';
    final uri = Uri.parse(url);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Tidak dapat membuka WhatsApp')),
        );
      }
    }
  }

  Future<void> _contactSales(String? phone, String shopName) async {
    if (phone == null || phone.isEmpty || phone == '-') {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Nomor sales tidak tersedia')),
      );
      return;
    }
    String formattedPhone = phone;
    if (formattedPhone.startsWith('0')) {
      formattedPhone = '62${formattedPhone.substring(1)}';
    }
    final url =
        'https://wa.me/$formattedPhone?text=Halo,%20saya%20operator%20pemasangan%20spanduk.%20Ingin%20konfirmasi%20terkait%20tugas%20pemasangan%20spanduk%20di%20toko%20$shopName.';
    final uri = Uri.parse(url);
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Tidak dapat membuka WhatsApp')),
        );
      }
    }
  }

  Future<void> _takeBanner(dynamic banner) async {
    if (!mounted) return;
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (c) =>
          const Center(child: CircularProgressIndicator(color: primaryEmerald)),
    );
    try {
      await _apiService.takeBanner(banner['id']);
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Spanduk berhasil diambil'),
            backgroundColor: primaryEmerald,
          ),
        );
        _fetchBanners();
        _tabController.animateTo(1);
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

  Future<void> _processInstall(dynamic banner) async {
    Position position;
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) throw Exception('GPS tidak aktif');
      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) {
          throw Exception('Izin lokasi ditolak');
        }
      }
      if (!mounted) return;
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (c) => const Center(
          child: CircularProgressIndicator(color: primaryEmerald),
        ),
      );
      position = await Geolocator.getCurrentPosition();
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: Colors.red),
        );
      }
      return;
    }

    final picker = ImagePicker();
    final XFile? image = await picker.pickImage(
      maxWidth: 1024,
      maxHeight: 1024,
      source: ImageSource.camera,
      imageQuality: 70,
    );
    if (image == null) return;

    if (!mounted) return;
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (c) =>
          const Center(child: CircularProgressIndicator(color: primaryEmerald)),
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
            backgroundColor: primaryEmerald,
          ),
        );
        _fetchBanners();
        _tabController.animateTo(2);
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

  void _showStorePhotoDialog(dynamic banner) {
    final shopName = banner['shop_name']?.toString().isNotEmpty == true
        ? banner['shop_name']
        : (banner['shop']?['name'] ??
              banner['Shop']?['Name'] ??
              'Toko Tanpa Nama');
    final shopAddress = banner['shop_address']?.toString().isNotEmpty == true
        ? banner['shop_address']
        : (banner['shop']?['address'] ?? banner['Shop']?['Address'] ?? '-');
    final rawShopPhoto =
        banner['shop_photo']?.toString() ??
        banner['shop']?['photo_url']?.toString() ??
        '';
    final shopPhoto = AppConstants.buildImageUrl(rawShopPhoto);
    final rawDesignLink = banner['design_link']?.toString() ?? '';
    final designLink = AppConstants.buildImageUrl(rawDesignLink);
    final shopPhone = banner['shop_phone']?.toString();
    final salesPhone = banner['sales_phone']?.toString();
    final lat = banner['latitude'] is num
        ? (banner['latitude'] as num).toDouble()
        : double.tryParse(banner['latitude']?.toString() ?? '') ?? 0.0;
    final lng = banner['longitude'] is num
        ? (banner['longitude'] as num).toDouble()
        : double.tryParse(banner['longitude']?.toString() ?? '') ?? 0.0;

    int selectedTab = 0;

    showDialog(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            final hasDesign = designLink.isNotEmpty;
            final currentImageUrl = selectedTab == 0 ? shopPhoto : designLink;

            return Dialog(
              backgroundColor: Colors.transparent,
              insetPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 24,
              ),
              child: Container(
                constraints: const BoxConstraints(
                  maxWidth: 480,
                  maxHeight: 620,
                ),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.25),
                      blurRadius: 20,
                      offset: const Offset(0, 10),
                    ),
                  ],
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    // Header
                    Container(
                      padding: const EdgeInsets.fromLTRB(16, 14, 12, 12),
                      decoration: const BoxDecoration(
                        color: darkEmerald,
                        borderRadius: BorderRadius.only(
                          topLeft: Radius.circular(20),
                          topRight: Radius.circular(20),
                        ),
                      ),
                      child: Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(6),
                            decoration: BoxDecoration(
                              color: Colors.white.withValues(alpha: 0.2),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: const Icon(
                              Icons.storefront_rounded,
                              color: Colors.white,
                              size: 18,
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  shopName,
                                  style: GoogleFonts.outfit(
                                    color: Colors.white,
                                    fontSize: 15,
                                    fontWeight: FontWeight.bold,
                                  ),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                Text(
                                  shopAddress,
                                  style: GoogleFonts.outfit(
                                    color: Colors.white.withValues(alpha: 0.8),
                                    fontSize: 11,
                                  ),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ],
                            ),
                          ),
                          IconButton(
                            icon: const Icon(
                              Icons.close_rounded,
                              color: Colors.white,
                              size: 22,
                            ),
                            onPressed: () => Navigator.pop(context),
                            padding: EdgeInsets.zero,
                            constraints: const BoxConstraints(),
                          ),
                        ],
                      ),
                    ),

                    // Tab Selector if Design Link Available
                    if (hasDesign)
                      Container(
                        color: const Color(0xFFF1F5F9),
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                        child: Row(
                          children: [
                            Expanded(
                              child: InkWell(
                                onTap: () =>
                                    setDialogState(() => selectedTab = 0),
                                borderRadius: BorderRadius.circular(8),
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 6,
                                  ),
                                  decoration: BoxDecoration(
                                    color: selectedTab == 0
                                        ? Colors.white
                                        : Colors.transparent,
                                    borderRadius: BorderRadius.circular(8),
                                    boxShadow: selectedTab == 0
                                        ? [
                                            BoxShadow(
                                              color: Colors.black.withValues(
                                                alpha: 0.05,
                                              ),
                                              blurRadius: 4,
                                            ),
                                          ]
                                        : null,
                                  ),
                                  child: Row(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.storefront_rounded,
                                        size: 14,
                                        color: selectedTab == 0
                                            ? primaryEmerald
                                            : Colors.grey[600],
                                      ),
                                      const SizedBox(width: 6),
                                      Text(
                                        'Foto Toko',
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: selectedTab == 0
                                              ? FontWeight.bold
                                              : FontWeight.w500,
                                          color: selectedTab == 0
                                              ? primaryEmerald
                                              : Colors.grey[600],
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 6),
                            Expanded(
                              child: InkWell(
                                onTap: () =>
                                    setDialogState(() => selectedTab = 1),
                                borderRadius: BorderRadius.circular(8),
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 6,
                                  ),
                                  decoration: BoxDecoration(
                                    color: selectedTab == 1
                                        ? Colors.white
                                        : Colors.transparent,
                                    borderRadius: BorderRadius.circular(8),
                                    boxShadow: selectedTab == 1
                                        ? [
                                            BoxShadow(
                                              color: Colors.black.withValues(
                                                alpha: 0.05,
                                              ),
                                              blurRadius: 4,
                                            ),
                                          ]
                                        : null,
                                  ),
                                  child: Row(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.image_outlined,
                                        size: 14,
                                        color: selectedTab == 1
                                            ? primaryEmerald
                                            : Colors.grey[600],
                                      ),
                                      const SizedBox(width: 6),
                                      Text(
                                        'Desain Spanduk',
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: selectedTab == 1
                                              ? FontWeight.bold
                                              : FontWeight.w500,
                                          color: selectedTab == 1
                                              ? primaryEmerald
                                              : Colors.grey[600],
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),

                    // Main Image Area
                    Expanded(
                      child: Container(
                        color: const Color(0xFF0F172A),
                        width: double.infinity,
                        child: currentImageUrl.isNotEmpty
                            ? InteractiveViewer(
                                minScale: 0.8,
                                maxScale: 4.0,
                                child: AppNetworkImage(
                                  imageUrl: currentImageUrl,
                                  fit: BoxFit.contain,
                                  placeholder: const Center(
                                    child: CircularProgressIndicator(
                                      color: Colors.white,
                                    ),
                                  ),
                                  errorWidget: Center(
                                    child: Column(
                                      mainAxisAlignment:
                                          MainAxisAlignment.center,
                                      children: [
                                        Icon(
                                          Icons.broken_image_rounded,
                                          size: 48,
                                          color: Colors.grey[500],
                                        ),
                                        const SizedBox(height: 8),
                                        Text(
                                          'Gagal memuat gambar',
                                          style: GoogleFonts.outfit(
                                            color: Colors.grey[400],
                                            fontSize: 12,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  memCacheWidth: 1000,
                                ),
                              )
                            : Center(
                                child: Padding(
                                  padding: const EdgeInsets.all(24),
                                  child: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Container(
                                        padding: const EdgeInsets.all(16),
                                        decoration: BoxDecoration(
                                          color: Colors.white.withValues(
                                            alpha: 0.1,
                                          ),
                                          shape: BoxShape.circle,
                                        ),
                                        child: const Icon(
                                          Icons.add_a_photo_outlined,
                                          size: 40,
                                          color: Colors.white70,
                                        ),
                                      ),
                                      const SizedBox(height: 12),
                                      Text(
                                        selectedTab == 0
                                            ? 'Foto profil toko belum diunggah'
                                            : 'Preview desain belum tersedia',
                                        style: GoogleFonts.outfit(
                                          color: Colors.white,
                                          fontSize: 13,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                      const SizedBox(height: 4),
                                      Text(
                                        'Hubungi Sales PIC untuk konfirmasi bentuk fisik toko sebelum pemasangan.',
                                        textAlign: TextAlign.center,
                                        style: GoogleFonts.outfit(
                                          color: Colors.white70,
                                          fontSize: 11,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                      ),
                    ),

                    // Action Footer Bar
                    Container(
                      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
                      decoration: const BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.only(
                          bottomLeft: Radius.circular(20),
                          bottomRight: Radius.circular(20),
                        ),
                        border: Border(
                          top: BorderSide(color: Color(0xFFE2E8F0)),
                        ),
                      ),
                      child: Row(
                        children: [
                          if (shopPhone != null &&
                              shopPhone.isNotEmpty &&
                              shopPhone != '-') ...[
                            Expanded(
                              child: OutlinedButton.icon(
                                onPressed: () {
                                  Navigator.pop(context);
                                  _contactStore(shopPhone);
                                },
                                icon: const Icon(
                                  Icons.storefront_rounded,
                                  size: 13,
                                ),
                                label: Text(
                                  'Chat Toko',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                                style: OutlinedButton.styleFrom(
                                  foregroundColor: primaryEmerald,
                                  side: const BorderSide(color: primaryEmerald),
                                  minimumSize: const Size(0, 34),
                                  tapTargetSize:
                                      MaterialTapTargetSize.shrinkWrap,
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 6,
                                    horizontal: 8,
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 6),
                          ],
                          Expanded(
                            child: OutlinedButton.icon(
                              onPressed: () {
                                Navigator.pop(context);
                                _contactSales(salesPhone, shopName);
                              },
                              icon: const Icon(Icons.person_rounded, size: 13),
                              label: Text(
                                'Chat Sales',
                                style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              style: OutlinedButton.styleFrom(
                                foregroundColor: Colors.blue[700],
                                side: BorderSide(color: Colors.blue[600]!),
                                minimumSize: const Size(0, 34),
                                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                padding: const EdgeInsets.symmetric(
                                  vertical: 6,
                                  horizontal: 8,
                                ),
                              ),
                            ),
                          ),
                          if (lat != 0.0 && lng != 0.0) ...[
                            const SizedBox(width: 6),
                            IconButton(
                              onPressed: () => _launchMaps(lat, lng),
                              icon: const Icon(
                                Icons.map_rounded,
                                color: Colors.indigo,
                                size: 18,
                              ),
                              tooltip: 'Buka di Google Maps',
                              style: IconButton.styleFrom(
                                backgroundColor: const Color(0xFFEEF2FF),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                padding: const EdgeInsets.all(8),
                                minimumSize: const Size(34, 34),
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildEmptyState(String message) {
    return ListView(
      physics: const AlwaysScrollableScrollPhysics(),
      children: [
        SizedBox(height: MediaQuery.of(context).size.height * 0.25),
        Icon(Icons.inbox_outlined, size: 64, color: Colors.grey[400]),
        const SizedBox(height: 16),
        Center(
          child: Text(
            message,
            style: GoogleFonts.outfit(
              color: Colors.grey[600],
              fontWeight: FontWeight.w500,
              fontSize: 16,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildBannerList(
    List<dynamic> banners,
    String emptyMessage,
    String listType,
  ) {
    final q = _searchQuery.trim().toLowerCase();
    final filtered = banners.where((b) {
      if (q.isEmpty) return true;
      final name =
          (b['shop_name'] ??
                  b['store_name'] ??
                  (b['shop'] is Map ? b['shop']['name'] : ''))
              .toString()
              .toLowerCase();
      final addr = (b['shop_address'] ?? b['address'] ?? b['location'] ?? '')
          .toString()
          .toLowerCase();
      final sales = (b['sales_name'] ?? '').toString().toLowerCase();
      final owner = (b['owner_name'] ?? '').toString().toLowerCase();
      final district = (b['district'] ?? '').toString().toLowerCase();
      final regency = (b['regency'] ?? '').toString().toLowerCase();
      final key = '$name $owner $sales $addr $district $regency';
      return key.contains(q);
    }).toList();

    if (filtered.isEmpty) return _buildEmptyState(emptyMessage);

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: filtered.length,
      itemBuilder: (context, index) {
        final banner = filtered[index];
        final rawShopPhoto =
            banner['shop_photo']?.toString() ??
            banner['shop']?['photo_url']?.toString() ??
            '';
        double? lat = banner['latitude'] != null
            ? double.tryParse(banner['latitude'].toString())
            : null;
        double? lng = banner['longitude'] != null
            ? double.tryParse(banner['longitude'].toString())
            : null;

        return Card(
          elevation: 2,
          shadowColor: Colors.black12,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
          ),
          margin: const EdgeInsets.only(bottom: 16),
          child: Container(
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: Colors.grey[100]!),
              color: Colors.white,
            ),
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: listType == 'printed'
                            ? Colors.blue[50]
                            : (listType == 'waiting'
                                  ? Colors.orange[50]
                                  : lightEmerald),
                        borderRadius: BorderRadius.circular(20),
                        border: Border.all(
                          color: listType == 'printed'
                              ? Colors.blue[200]!
                              : (listType == 'waiting'
                                    ? Colors.orange[200]!
                                    : primaryEmerald.withValues(alpha: 0.5)),
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            listType == 'printed'
                                ? Icons.print
                                : (listType == 'waiting'
                                      ? Icons.pending_actions
                                      : Icons.check_circle),
                            size: 14,
                            color: listType == 'printed'
                                ? Colors.blue[700]
                                : (listType == 'waiting'
                                      ? Colors.orange[700]
                                      : darkEmerald),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            listType == 'printed'
                                ? 'SUDAH DICETAK'
                                : (listType == 'waiting'
                                      ? 'MENUNGGU DIPASANG'
                                      : 'SELESAI'),
                            style: GoogleFonts.outfit(
                              fontSize: 10,
                              fontWeight: FontWeight.bold,
                              color: listType == 'printed'
                                  ? Colors.blue[700]
                                  : (listType == 'waiting'
                                        ? Colors.orange[700]
                                        : darkEmerald),
                            ),
                          ),
                        ],
                      ),
                    ),
                    if (lat != null && lng != null)
                      GestureDetector(
                        onTap: () => _launchMaps(lat, lng),
                        child: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.grey[100],
                            shape: BoxShape.circle,
                          ),
                          child: Icon(
                            Icons.directions,
                            size: 20,
                            color: Colors.grey[700],
                          ),
                        ),
                      ),
                  ],
                ),
                const SizedBox(height: 14),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    InkWell(
                      onTap: () => _showStorePhotoDialog(banner),
                      borderRadius: BorderRadius.circular(10),
                      child: Container(
                        width: 44,
                        height: 44,
                        decoration: BoxDecoration(
                          color: const Color(0xFFF1F5F9),
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(color: const Color(0xFFCBD5E1)),
                        ),
                        clipBehavior: Clip.antiAlias,
                        child: rawShopPhoto.isNotEmpty
                            ? AppNetworkImage(
                                imageUrl: rawShopPhoto,
                                fit: BoxFit.cover,
                                placeholder: const Center(
                                  child: Icon(
                                    Icons.storefront_rounded,
                                    size: 20,
                                    color: Color(0xFF94A3B8),
                                  ),
                                ),
                                errorWidget: const Center(
                                  child: Icon(
                                    Icons.storefront_rounded,
                                    size: 20,
                                    color: Color(0xFF94A3B8),
                                  ),
                                ),
                                memCacheWidth: 200,
                              )
                            : const Center(
                                child: Icon(
                                  Icons.storefront_rounded,
                                  size: 22,
                                  color: primaryEmerald,
                                ),
                              ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            banner['shop_name']?.toString().isNotEmpty == true
                                ? banner['shop_name']
                                : (banner['shop']?['name'] ??
                                      banner['Shop']?['Name'] ??
                                      'Toko Tidak Diketahui'),
                            style: GoogleFonts.outfit(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: Colors.black87,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Row(
                            children: [
                              Icon(
                                Icons.straighten,
                                size: 13,
                                color: Colors.grey[500],
                              ),
                              const SizedBox(width: 4),
                              Text(
                                '${banner['type'] ?? 'Spanduk'} • ${banner['size_length'] ?? 0}x${banner['size_width'] ?? 0} cm',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  fontWeight: FontWeight.w600,
                                  color: Colors.grey[600],
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: Colors.grey[50],
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(color: Colors.grey[200]!),
                  ),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Icon(
                        Icons.location_on,
                        size: 15,
                        color: Colors.grey[400],
                      ),
                      const SizedBox(width: 6),
                      Expanded(
                        child: Text(
                          banner['shop_address']?.toString().isNotEmpty == true
                              ? banner['shop_address']
                              : (banner['address'] ??
                                    banner['shop']?['address'] ??
                                    banner['notes'] ??
                                    '-'),
                          style: GoogleFonts.outfit(
                            fontSize: 11.5,
                            color: Colors.grey[700],
                            height: 1.4,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton.icon(
                        onPressed: () => _showStorePhotoDialog(banner),
                        icon: const Icon(Icons.photo_camera_rounded, size: 13),
                        label: Text(
                          'Foto Toko',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: const Color(0xFF334155),
                          side: const BorderSide(color: Color(0xFFCBD5E1)),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(8),
                          ),
                          padding: const EdgeInsets.symmetric(vertical: 8),
                        ),
                      ),
                    ),
                    const SizedBox(width: 6),
                    Expanded(
                      child: OutlinedButton.icon(
                        onPressed: () =>
                            _contactStore(banner['shop_phone']?.toString()),
                        icon: const Icon(Icons.storefront_rounded, size: 13),
                        label: Text(
                          'Chat Toko',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: primaryEmerald,
                          side: const BorderSide(color: primaryEmerald),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(8),
                          ),
                          padding: const EdgeInsets.symmetric(vertical: 8),
                        ),
                      ),
                    ),
                    const SizedBox(width: 6),
                    Expanded(
                      child: OutlinedButton.icon(
                        onPressed: () => _contactSales(
                          banner['sales_phone']?.toString(),
                          banner['shop_name'] ?? 'Toko',
                        ),
                        icon: const Icon(Icons.person_rounded, size: 13),
                        label: Text(
                          'Chat Sales',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.blue[700],
                          side: BorderSide(color: Colors.blue[700]!),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(8),
                          ),
                          padding: const EdgeInsets.symmetric(vertical: 8),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                if (listType == 'printed')
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: ElevatedButton.icon(
                      onPressed: () => _takeBanner(banner),
                      icon: const Icon(Icons.pan_tool, size: 20),
                      label: Text(
                        'Ambil Spanduk',
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.bold,
                          fontSize: 15,
                        ),
                      ),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.blue[600],
                        foregroundColor: Colors.white,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        elevation: 0,
                      ),
                    ),
                  ),
                if (listType == 'waiting')
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: ElevatedButton.icon(
                      onPressed: () => _processInstall(banner),
                      icon: const Icon(Icons.camera_alt, size: 20),
                      label: Text(
                        'Pasang & Upload Bukti',
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.bold,
                          fontSize: 15,
                        ),
                      ),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: primaryEmerald,
                        foregroundColor: Colors.white,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        elevation: 0,
                      ),
                    ),
                  ),
                if (listType == 'installed')
                  Container(
                    padding: const EdgeInsets.symmetric(
                      vertical: 12,
                      horizontal: 16,
                    ),
                    decoration: BoxDecoration(
                      color: lightEmerald.withValues(alpha: 0.5),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: primaryEmerald.withValues(alpha: 0.3),
                      ),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(
                          Icons.verified,
                          color: darkEmerald,
                          size: 20,
                        ),
                        const SizedBox(width: 8),
                        Text(
                          'Tugas Selesai (Settlement)',
                          style: GoogleFonts.outfit(
                            color: darkEmerald,
                            fontWeight: FontWeight.bold,
                            fontSize: 13,
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'Daftar Tugas',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.bold,
            fontSize: 18,
            color: Colors.white,
          ),
        ),
        backgroundColor: primaryEmerald,
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.white),
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: Colors.white,
          indicatorWeight: 3,
          labelColor: Colors.white,
          unselectedLabelColor: Colors.white70,
          labelStyle: GoogleFonts.outfit(
            fontWeight: FontWeight.bold,
            fontSize: 11,
          ),
          tabs: const [
            Tab(text: 'SUDAH DICETAK'),
            Tab(text: 'MENUNGGU\nDIPASANG'),
            Tab(text: 'SELESAI'),
          ],
        ),
      ),
      body: Column(
        children: [
          Container(
            color: primaryEmerald,
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
            child: Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(12),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.05),
                    blurRadius: 10,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: TextField(
                onChanged: (v) => setState(() => _searchQuery = v),
                decoration: InputDecoration(
                  hintText: 'Cari nama toko, sales, jalan, kecamatan...',
                  hintStyle: GoogleFonts.outfit(color: Colors.grey[400]),
                  prefixIcon: Icon(Icons.search, color: Colors.grey[400]),
                  suffixIcon: _searchQuery.isNotEmpty
                      ? IconButton(
                          icon: const Icon(Icons.clear_rounded, size: 18),
                          onPressed: () => setState(() => _searchQuery = ''),
                        )
                      : null,
                  border: InputBorder.none,
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 14,
                  ),
                ),
              ),
            ),
          ),
          Expanded(
            child: _isLoading
                ? const Center(
                    child: CircularProgressIndicator(color: primaryEmerald),
                  )
                : TabBarView(
                    controller: _tabController,
                    children: [
                      RefreshIndicator(
                        color: primaryEmerald,
                        onRefresh: _fetchBanners,
                        child: _buildBannerList(
                          _printedBanners,
                          'Belum ada spanduk yang dicetak',
                          'printed',
                        ),
                      ),
                      RefreshIndicator(
                        color: primaryEmerald,
                        onRefresh: _fetchBanners,
                        child: _buildBannerList(
                          _waitingBanners,
                          'Tidak ada spanduk menunggu dipasang',
                          'waiting',
                        ),
                      ),
                      RefreshIndicator(
                        color: primaryEmerald,
                        onRefresh: _fetchBanners,
                        child: _buildBannerList(
                          _installedBanners,
                          'Belum ada riwayat pemasangan',
                          'installed',
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
