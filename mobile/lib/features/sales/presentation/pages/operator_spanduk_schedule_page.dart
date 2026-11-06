import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:geolocator/geolocator.dart';
import 'package:get_it/get_it.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/widgets/app_network_image.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:hris_app/features/sales/data/services/banner_order_api_service.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/features/sales/presentation/pages/operator_spanduk_plan_page.dart';

enum ScheduleSortMode { optimal, nearest, original }

enum ScheduleStatusFilter { all, pending, completed }

class OperatorSpandukSchedulePage extends StatefulWidget {
  const OperatorSpandukSchedulePage({super.key});

  @override
  State<OperatorSpandukSchedulePage> createState() =>
      _OperatorSpandukSchedulePageState();
}

class _OperatorSpandukSchedulePageState
    extends State<OperatorSpandukSchedulePage> {
  final _bannerApiService = GetIt.I<BannerOrderApiService>();
  final _salesApiService = GetIt.I<SalesApiService>();

  bool _isLoading = true;
  bool _isLocating = false;
  Position? _currentPosition;

  List<dynamic> _rawPlannedBanners = [];
  List<dynamic> _displayedBanners = [];
  double _totalKm = 0.0;
  ScheduleSortMode _sortMode = ScheduleSortMode.optimal;
  ScheduleStatusFilter _statusFilter = ScheduleStatusFilter.all;

  static const Color primaryEmerald = Color(0xFF047857);
  static const Color darkEmerald = Color(0xFF064E3B);

  @override
  void initState() {
    super.initState();
    _fetchSchedule();
  }

  String _getShopName(dynamic b) {
    if (b['shop_name'] != null && b['shop_name'].toString().trim().isNotEmpty) {
      return b['shop_name'].toString().trim();
    }
    if (b['store_name'] != null &&
        b['store_name'].toString().trim().isNotEmpty) {
      return b['store_name'].toString().trim();
    }
    if (b['shop'] is Map &&
        b['shop']['name'] != null &&
        b['shop']['name'].toString().trim().isNotEmpty) {
      return b['shop']['name'].toString().trim();
    }
    if (b['Shop'] is Map &&
        b['Shop']['Name'] != null &&
        b['Shop']['Name'].toString().trim().isNotEmpty) {
      return b['Shop']['Name'].toString().trim();
    }
    return 'Toko Tanpa Nama';
  }

  String _getShopAddress(dynamic b) {
    if (b['shop_address'] != null &&
        b['shop_address'].toString().trim().isNotEmpty) {
      return b['shop_address'].toString().trim();
    }
    if (b['address'] != null && b['address'].toString().trim().isNotEmpty) {
      return b['address'].toString().trim();
    }
    if (b['location'] != null && b['location'].toString().trim().isNotEmpty) {
      return b['location'].toString().trim();
    }
    if (b['shop'] is Map &&
        b['shop']['address'] != null &&
        b['shop']['address'].toString().trim().isNotEmpty) {
      return b['shop']['address'].toString().trim();
    }
    if (b['Shop'] is Map &&
        b['Shop']['Address'] != null &&
        b['Shop']['Address'].toString().trim().isNotEmpty) {
      return b['Shop']['Address'].toString().trim();
    }
    return '-';
  }

  double? _getLatitude(dynamic b) {
    final val =
        b['latitude'] ??
        b['lat'] ??
        (b['shop'] is Map ? b['shop']['latitude'] : null) ??
        (b['Shop'] is Map ? b['Shop']['Latitude'] : null);
    if (val == null) return null;
    final d = double.tryParse(val.toString());
    if (d != null && d != 0.0) return d;
    return null;
  }

  double? _getLongitude(dynamic b) {
    final val =
        b['longitude'] ??
        b['lng'] ??
        (b['shop'] is Map ? b['shop']['longitude'] : null) ??
        (b['Shop'] is Map ? b['Shop']['Longitude'] : null);
    if (val == null) return null;
    final d = double.tryParse(val.toString());
    if (d != null && d != 0.0) return d;
    return null;
  }

  Future<Position?> _getLastKnownPosition() async {
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) return null;
      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied ||
          permission == LocationPermission.deniedForever) {
        return null;
      }
      return await Geolocator.getLastKnownPosition();
    } catch (_) {
      return null;
    }
  }

  Future<Position?> _determinePosition() async {
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) return null;

      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) return null;
      }
      if (permission == LocationPermission.deniedForever) return null;

      return await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.medium,
        timeLimit: const Duration(seconds: 7),
      );
    } catch (_) {
      return null;
    }
  }

  void _updatePreciseLocationInBackground() {
    Geolocator.isLocationServiceEnabled()
        .then((enabled) {
          if (!enabled) return;
          Geolocator.checkPermission().then((perm) {
            if (perm == LocationPermission.denied ||
                perm == LocationPermission.deniedForever) {
              return;
            }
            Geolocator.getCurrentPosition(
                  desiredAccuracy: LocationAccuracy.medium,
                  timeLimit: const Duration(seconds: 8),
                )
                .then((pos) {
                  if (!mounted) return;
                  _currentPosition = pos;
                  _applySortAndCalculate();
                  if (mounted) {
                    setState(() {});
                  }
                })
                .catchError((_) {});
          });
        })
        .catchError((_) {});
  }

  Future<void> _fetchSchedule() async {
    setState(() => _isLoading = true);
    try {
      // 1. Fetch user identity & branch operator dari auth state (dengan fallback API langsung jika null)
      String currentEmpId = '';
      String currentUserId = '';
      String currentUserName = '';

      final authState = context.read<AuthBloc>().state;
      Map<String, dynamic>? profile = authState is Authenticated
          ? authState.userProfile
          : null;

      if (profile == null) {
        try {
          final res = await _bannerApiService.apiClient.client.get(
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
        final firstName = profile['first_name']?.toString().trim() ?? '';
        final lastName = profile['last_name']?.toString().trim() ?? '';
        if (firstName.isNotEmpty) {
          currentUserName = '$firstName $lastName'.trim();
        } else {
          currentUserName =
              (profile['name'] ??
                      profile['full_name'] ??
                      profile['user']?['name'] ??
                      '')
                  .toString()
                  .trim();
        }
      }

      // 2. Fetch Visit Plans, Banners, dan Last Known GPS secara paralel
      // Tanpa membatasi branchId agar spanduk yang diambil dari semua wilayah tetap muncul
      final todayStr = DateFormat('yyyy-MM-dd').format(DateTime.now());
      final visitPlansFuture = _salesApiService.getVisitPlans(date: todayStr);
      final bannersFuture = _bannerApiService.getAllBannersChunked();
      final lastPosFuture = _getLastKnownPosition();

      final combined = await Future.wait([
        visitPlansFuture,
        bannersFuture,
        lastPosFuture,
      ]);
      final visitPlans = combined[0] as List<dynamic>;
      final banners = combined[1] as List<dynamic>;
      _currentPosition = combined[2] as Position?;

      // 3. Map store IDs dari visit plans hari ini
      final Map<String, dynamic> visitPlanMap = {};
      final Set<String> plannedStoreIds = {};

      for (var vp in visitPlans) {
        final sid = vp['store_id']?.toString();
        if (sid != null && sid.isNotEmpty) {
          plannedStoreIds.add(sid);
          visitPlanMap[sid] = vp;
        }
      }

      // 4. Filter banners: Tampilkan spanduk yang ada di rencana kunjungan ATAU yang sudah diambil operator ini (termasuk yang selesai hari ini)
      final List<dynamic> matchedBanners = [];
      final Set<String> seenBannerIds = {};

      for (var b in banners) {
        final bId = (b['id'] ?? b['ID'])?.toString() ?? '';
        if (bId.isNotEmpty && seenBannerIds.contains(bId)) {
          continue;
        }

        final status = b['status']?.toString().toLowerCase() ?? '';
        final isInstalled = (status == 'installed' || status == 'terpasang');

        final storeId =
            (b['shop_id'] ??
                    b['store_id'] ??
                    (b['shop'] is Map ? b['shop']['id'] : null) ??
                    (b['Shop'] is Map ? b['Shop']['ID'] : null))
                ?.toString();

        final bannerOpId =
            (b['operator_id'] ?? b['OperatorID'])?.toString().trim() ?? '';
        final bannerOpName =
            (b['operator_name'] ?? b['OperatorName'])
                ?.toString()
                .trim()
                .toLowerCase() ??
            '';
        final myName = currentUserName.toLowerCase();

        final bool isMyOperator =
            (bannerOpId.isNotEmpty &&
                (bannerOpId == currentEmpId || bannerOpId == currentUserId)) ||
            (myName.isNotEmpty &&
                bannerOpName.isNotEmpty &&
                (bannerOpName == myName ||
                    bannerOpName.contains(myName) ||
                    myName.contains(bannerOpName)));

        final bool isTakenByMe =
            (status == 'diterima_operator') && isMyOperator;
        final bool isPlannedStore =
            storeId != null && plannedStoreIds.contains(storeId);

        final installedAt = b['installed_at']?.toString() ?? '';
        final updatedAt = b['updated_at']?.toString() ?? '';
        final bool isInstalledToday =
            installedAt.startsWith(todayStr) ||
            updatedAt.startsWith(todayStr) ||
            isPlannedStore;

        final bool isDoneByMeToday =
            isInstalled && isMyOperator && isInstalledToday;

        if (isTakenByMe || isPlannedStore || isDoneByMeToday) {
          if (bId.isNotEmpty) seenBannerIds.add(bId);

          // Salin / perkaya data jika field banner belum lengkap
          final vp = storeId != null ? visitPlanMap[storeId] : null;
          final enriched = Map<String, dynamic>.from(b);
          if ((enriched['shop_photo'] == null ||
                  enriched['shop_photo'].toString().trim().isEmpty) &&
              vp != null &&
              (vp['photo_url'] != null || vp['store_photo'] != null)) {
            enriched['shop_photo'] = vp['photo_url'] ?? vp['store_photo'];
          }
          if ((enriched['shop_name'] == null ||
                  enriched['shop_name'].toString().trim().isEmpty) &&
              vp != null &&
              vp['store_name'] != null) {
            enriched['shop_name'] = vp['store_name'];
          }
          if ((enriched['shop_address'] == null ||
                  enriched['shop_address'].toString().trim().isEmpty) &&
              vp != null &&
              vp['address'] != null) {
            enriched['shop_address'] = vp['address'];
          }
          if (enriched['latitude'] == null &&
              vp != null &&
              vp['latitude'] != null) {
            enriched['latitude'] = vp['latitude'];
          }
          if (enriched['longitude'] == null &&
              vp != null &&
              vp['longitude'] != null) {
            enriched['longitude'] = vp['longitude'];
          }
          matchedBanners.add(enriched);
        }
      }

      _rawPlannedBanners = matchedBanners;
      _applySortAndCalculate();

      if (mounted) {
        setState(() => _isLoading = false);
      }

      // 5. Update GPS presisi di background
      _updatePreciseLocationInBackground();
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Gagal memuat jadwal: $e')));
      }
    }
  }

  void _applySortAndCalculate() {
    if (_rawPlannedBanners.isEmpty) {
      _displayedBanners = [];
      _totalKm = 0.0;
      return;
    }

    // 1. Saring list berdasarkan tab filter status
    List<dynamic> filteredList;
    switch (_statusFilter) {
      case ScheduleStatusFilter.pending:
        filteredList = _rawPlannedBanners.where((b) {
          final s = b['status']?.toString().toLowerCase() ?? '';
          return s != 'installed' && s != 'terpasang';
        }).toList();
        break;
      case ScheduleStatusFilter.completed:
        filteredList = _rawPlannedBanners.where((b) {
          final s = b['status']?.toString().toLowerCase() ?? '';
          return s == 'installed' || s == 'terpasang';
        }).toList();
        break;
      case ScheduleStatusFilter.all:
        filteredList = List.from(_rawPlannedBanners);
        break;
    }

    if (filteredList.isEmpty) {
      _displayedBanners = [];
      _totalKm = 0.0;
      return;
    }

    // 2. Hitung jarak dari posisi saat ini ke tiap banner
    for (var b in filteredList) {
      final lat = _getLatitude(b);
      final lng = _getLongitude(b);
      if (_currentPosition != null && lat != null && lng != null) {
        b['_dist_from_me'] = Geolocator.distanceBetween(
          _currentPosition!.latitude,
          _currentPosition!.longitude,
          lat,
          lng,
        );
      } else {
        b['_dist_from_me'] = double.infinity;
      }
    }

    List<dynamic> list;
    switch (_sortMode) {
      case ScheduleSortMode.optimal:
        list = _calculateOptimalRoute(filteredList, _currentPosition);
        break;
      case ScheduleSortMode.nearest:
        list = List.from(filteredList)
          ..sort((a, b) {
            final distA = (a['_dist_from_me'] as double?) ?? double.infinity;
            final distB = (b['_dist_from_me'] as double?) ?? double.infinity;
            return distA.compareTo(distB);
          });
        break;
      case ScheduleSortMode.original:
        list = List.from(filteredList);
        break;
    }

    // 3. Hitung Total KM berurutan
    double totalMeters = 0.0;
    if (_currentPosition != null && list.isNotEmpty) {
      final firstLat = _getLatitude(list.first);
      final firstLng = _getLongitude(list.first);
      if (firstLat != null && firstLng != null) {
        totalMeters += Geolocator.distanceBetween(
          _currentPosition!.latitude,
          _currentPosition!.longitude,
          firstLat,
          firstLng,
        );
      }
    }

    for (int i = 0; i < list.length - 1; i++) {
      final lat1 = _getLatitude(list[i]);
      final lng1 = _getLongitude(list[i]);
      final lat2 = _getLatitude(list[i + 1]);
      final lng2 = _getLongitude(list[i + 1]);

      if (lat1 != null && lng1 != null && lat2 != null && lng2 != null) {
        totalMeters += Geolocator.distanceBetween(lat1, lng1, lat2, lng2);
      }
    }

    _displayedBanners = list;
    _totalKm = totalMeters / 1000;
  }

  List<dynamic> _calculateOptimalRoute(
    List<dynamic> items,
    Position? startPos,
  ) {
    if (items.isEmpty) return [];
    if (startPos == null) return List.from(items);

    final unvisited = List<dynamic>.from(items);
    final optimized = <dynamic>[];

    double currentLat = startPos.latitude;
    double currentLng = startPos.longitude;

    while (unvisited.isNotEmpty) {
      int bestIdx = -1;
      double minDistance = double.infinity;

      for (int i = 0; i < unvisited.length; i++) {
        final lat = _getLatitude(unvisited[i]);
        final lng = _getLongitude(unvisited[i]);

        if (lat != null && lng != null) {
          final dist = Geolocator.distanceBetween(
            currentLat,
            currentLng,
            lat,
            lng,
          );
          if (dist < minDistance) {
            minDistance = dist;
            bestIdx = i;
          }
        }
      }

      if (bestIdx == -1) {
        // Toko yang tidak punya koordinat ditaruh di akhir
        optimized.addAll(unvisited);
        break;
      }

      final nextItem = unvisited.removeAt(bestIdx);
      optimized.add(nextItem);
      final nextLat = _getLatitude(nextItem);
      final nextLng = _getLongitude(nextItem);
      if (nextLat != null && nextLng != null) {
        currentLat = nextLat;
        currentLng = nextLng;
      }
    }

    return optimized;
  }

  Future<void> _refreshLocationAndSort() async {
    setState(() => _isLocating = true);
    final pos = await _determinePosition();
    if (mounted) {
      setState(() {
        _currentPosition = pos;
        _isLocating = false;
        _applySortAndCalculate();
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            pos != null
                ? 'Lokasi berhasil diperbarui & rute dioptimalkan.'
                : 'Gagal mendapatkan sinyal GPS. Pastikan izin lokasi aktif.',
          ),
          duration: const Duration(seconds: 2),
        ),
      );
    }
  }

  Future<void> _processInstall(dynamic banner) async {
    final status = banner['status']?.toString().toLowerCase();
    final isUpdate = (status == 'installed' || status == 'terpasang');

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
      maxWidth: 1920,
      maxHeight: 1920,
      source: ImageSource.camera,
      imageQuality: 85,
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
      final bannerId = banner['id'];
      final bytes = await image.readAsBytes();
      await _bannerApiService.installBanner(
        bannerId,
        image.path,
        position.latitude,
        position.longitude,
        imageBytes: bytes,
        filename: image.name,
      );
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              isUpdate
                  ? 'Foto Bukti Pemasangan Berhasil Diperbarui'
                  : 'Pemasangan Spanduk Berhasil Disimpan',
            ),
            backgroundColor: primaryEmerald,
          ),
        );
        _fetchSchedule();
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

  Future<void> _openGoogleMapsMultiRoute() async {
    if (_displayedBanners.isEmpty) return;
    try {
      Position? position = _currentPosition;
      if (position == null) {
        showDialog(
          context: context,
          barrierDismissible: false,
          builder: (c) => const Center(
            child: CircularProgressIndicator(color: primaryEmerald),
          ),
        );
        position = await _determinePosition();
        if (mounted) Navigator.pop(context);
      }

      final validCoords = <Map<String, double>>[];
      for (var b in _displayedBanners) {
        final lat = _getLatitude(b);
        final lng = _getLongitude(b);
        if (lat != null && lng != null) {
          validCoords.add({'lat': lat, 'lng': lng});
        }
      }

      if (validCoords.isEmpty) {
        throw Exception('Data koordinat toko tidak tersedia untuk navigasi');
      }

      String origin = '';
      if (position != null) {
        origin = '${position.latitude},${position.longitude}';
      } else {
        // Keep the first store in the list. For a single store, removing it
        // would incorrectly leave no valid destination.
        origin = '${validCoords.first['lat']},${validCoords.first['lng']}';
      }

      if (validCoords.isEmpty) {
        throw Exception('Toko tujuan tidak valid');
      }

      final destinationCoord = validCoords.last;
      final destination =
          '${destinationCoord['lat']},${destinationCoord['lng']}';
      final waypointList = validCoords.sublist(0, validCoords.length - 1);

      String url =
          'https://www.google.com/maps/dir/?api=1&origin=$origin&destination=$destination';
      if (waypointList.isNotEmpty) {
        final waypointsStr = waypointList
            .map((w) => '${w['lat']},${w['lng']}')
            .join('|');
        url += '&waypoints=$waypointsStr';
      }

      final uri = Uri.parse(url);
      // `canLaunchUrl` can return false on Android when URL package
      // visibility is restricted, even though a browser can open the URL.
      // Try the external handler first, then fall back to the platform browser.
      final opened = await launchUrl(uri, mode: LaunchMode.externalApplication);
      if (!opened) {
        final browserOpened = await launchUrl(
          uri,
          mode: LaunchMode.platformDefault,
        );
        if (!browserOpened) {
          throw Exception('Tidak dapat membuka Google Maps atau browser');
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _openSingleLocation(dynamic banner) async {
    final lat = _getLatitude(banner);
    final lng = _getLongitude(banner);
    final shopName = _getShopName(banner);
    final shopAddress = _getShopAddress(banner);

    String url;
    if (lat != null && lng != null) {
      // Direct navigation to exact GPS coordinates (q=lat,lng places a pin directly, &dirflg=d triggers driving)
      url =
          'https://www.google.com/maps/dir/?api=1&destination=$lat,$lng&travelmode=driving';
    } else {
      // Fallback search by shop name & address
      final queryText = shopAddress != '-'
          ? '$shopName, $shopAddress'
          : shopName;
      final encodedQuery = Uri.encodeComponent(queryText);
      url = 'https://www.google.com/maps/search/?api=1&query=$encodedQuery';
    }

    final uri = Uri.parse(url);
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        // Fallback to standard web map url if external app intent fails
        final webUri = Uri.parse('https://maps.google.com/?q=$lat,$lng');
        await launchUrl(webUri, mode: LaunchMode.platformDefault);
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Tidak dapat membuka aplikasi Google Maps'),
          ),
        );
      }
    }
  }

  Future<void> _contactStore(String? phone, String shopName) async {
    if (phone == null || phone.isEmpty || phone == '-') {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Nomor telepon toko tidak tersedia')),
      );
      return;
    }
    String formattedPhone = phone.replaceAll(RegExp(r'[^0-9]'), '');
    if (formattedPhone.startsWith('0')) {
      formattedPhone = '62${formattedPhone.substring(1)}';
    }
    final encodedShop = Uri.encodeComponent(shopName);
    final url =
        'https://wa.me/$formattedPhone?text=Halo,%20kami%20dari%20tim%20pemasangan%20spanduk%20Karsa%20HR%20ingin%20konfirmasi%20terkait%20jadwal%20pemasangan%20spanduk%20di%20toko%20$encodedShop.';
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
        const SnackBar(content: Text('Nomor telepon sales tidak tersedia')),
      );
      return;
    }
    String formattedPhone = phone.replaceAll(RegExp(r'[^0-9]'), '');
    if (formattedPhone.startsWith('0')) {
      formattedPhone = '62${formattedPhone.substring(1)}';
    }
    final encodedShop = Uri.encodeComponent(shopName);
    final url =
        'https://wa.me/$formattedPhone?text=Halo,%20saya%20operator%20pemasangan%20spanduk.%20Ingin%20konfirmasi%20terkait%20tugas%20pemasangan%20spanduk%20di%20toko%20$encodedShop.';
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

  void _showStorePhotoDialog(BuildContext context, dynamic banner) {
    final shopName = _getShopName(banner);
    final shopAddress = _getShopAddress(banner);
    final rawShopPhoto =
        banner['shop_photo']?.toString() ??
        banner['shop']?['photo_url']?.toString() ??
        '';
    final shopPhoto = AppConstants.buildImageUrl(rawShopPhoto);
    final rawDesignLink = banner['design_link']?.toString() ?? '';
    final designLink = AppConstants.buildImageUrl(rawDesignLink);
    final rawPhotoInstalled = banner['photo_installed']?.toString() ?? '';
    final photoInstalled = AppConstants.buildImageUrl(rawPhotoInstalled);
    final shopPhone = banner['shop_phone']?.toString();
    final salesPhone = banner['sales_phone']?.toString();
    final lat = _getLatitude(banner);
    final lng = _getLongitude(banner);

    final status = banner['status']?.toString().toLowerCase() ?? '';
    final isDone = status == 'installed' || status == 'terpasang';

    // Jika sudah terpasang dan ada fotonya, default ke tab Bukti Pasang (2), selain itu 0
    int selectedTab = (isDone && photoInstalled.isNotEmpty) ? 2 : 0;

    showDialog(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            final hasDesign = designLink.isNotEmpty;
            final hasInstalled = photoInstalled.isNotEmpty;

            String currentImageUrl = shopPhoto;
            if (selectedTab == 1 && hasDesign) {
              currentImageUrl = designLink;
            } else if (selectedTab == 2 && hasInstalled) {
              currentImageUrl = photoInstalled;
            } else {
              currentImageUrl = shopPhoto;
            }

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

                    // Tab Selector
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
                                    const SizedBox(width: 4),
                                    Text(
                                      'Foto Toko',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
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
                          if (hasDesign) ...[
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
                                      const SizedBox(width: 4),
                                      Text(
                                        'Desain',
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
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
                          if (hasInstalled) ...[
                            const SizedBox(width: 6),
                            Expanded(
                              child: InkWell(
                                onTap: () =>
                                    setDialogState(() => selectedTab = 2),
                                borderRadius: BorderRadius.circular(8),
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 6,
                                  ),
                                  decoration: BoxDecoration(
                                    color: selectedTab == 2
                                        ? Colors.white
                                        : Colors.transparent,
                                    borderRadius: BorderRadius.circular(8),
                                    boxShadow: selectedTab == 2
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
                                        Icons.check_circle_rounded,
                                        size: 14,
                                        color: selectedTab == 2
                                            ? primaryEmerald
                                            : Colors.grey[600],
                                      ),
                                      const SizedBox(width: 4),
                                      Text(
                                        'Bukti Pasang',
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          fontWeight: selectedTab == 2
                                              ? FontWeight.bold
                                              : FontWeight.w500,
                                          color: selectedTab == 2
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
                                            : (selectedTab == 1
                                                  ? 'Preview desain belum tersedia'
                                                  : 'Foto bukti pemasangan belum tersedia'),
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
                                  _contactStore(shopPhone, shopName);
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
                          if (lat != null && lng != null) ...[
                            const SizedBox(width: 6),
                            IconButton(
                              onPressed: () => _openSingleLocation(banner),
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

  String _formatDistance(double? meters) {
    if (meters == null || meters == double.infinity)
      return 'Jarak tidak diketahui';
    if (meters < 1000) {
      return '${meters.toStringAsFixed(0)} m';
    }
    return '${(meters / 1000).toStringAsFixed(1)} km';
  }

  @override
  Widget build(BuildContext context) {
    final totalPlannedCount = _rawPlannedBanners.length;
    final completedCount = _rawPlannedBanners.where((b) {
      final s = b['status']?.toString().toLowerCase() ?? '';
      return s == 'installed' || s == 'terpasang';
    }).length;
    final pendingCount = totalPlannedCount - completedCount;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'Jadwal Pemasangan',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.bold,
            fontSize: 18,
            color: Colors.white,
          ),
        ),
        backgroundColor: primaryEmerald,
        elevation: 0,
        iconTheme: const IconThemeData(color: Colors.white),
        actions: [
          IconButton(
            tooltip: 'Perbarui Lokasi GPS',
            onPressed: _isLocating ? null : _refreshLocationAndSort,
            icon: _isLocating
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: Colors.white,
                    ),
                  )
                : const Icon(Icons.my_location_rounded, color: Colors.white),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const OperatorSpandukPlanPage()),
          ).then((_) => _fetchSchedule());
        },
        backgroundColor: primaryEmerald,
        icon: const Icon(Icons.add_task_rounded, color: Colors.white),
        label: Text(
          'Buat Rencana',
          style: GoogleFonts.outfit(
            color: Colors.white,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(color: primaryEmerald),
            )
          : RefreshIndicator(
              onRefresh: _fetchSchedule,
              color: primaryEmerald,
              child: ListView(
                padding: const EdgeInsets.all(16),
                physics: const AlwaysScrollableScrollPhysics(),
                children: [
                  // MISSION & ROUTE SUMMARY CARD
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(
                        color: primaryEmerald.withValues(alpha: 0.25),
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: primaryEmerald.withValues(alpha: 0.06),
                          blurRadius: 16,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: Column(
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              'TARGET PEMASANGAN HARI INI',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w900,
                                color: const Color(0xFF64748B),
                                letterSpacing: 1.5,
                              ),
                            ),
                            if (_currentPosition != null)
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 3,
                                ),
                                decoration: BoxDecoration(
                                  color: primaryEmerald.withValues(alpha: 0.12),
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    const Icon(
                                      Icons.gps_fixed_rounded,
                                      size: 11,
                                      color: darkEmerald,
                                    ),
                                    const SizedBox(width: 4),
                                    Text(
                                      'GPS Aktif',
                                      style: GoogleFonts.outfit(
                                        fontSize: 10,
                                        fontWeight: FontWeight.bold,
                                        color: darkEmerald,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                          ],
                        ),
                        const SizedBox(height: 18),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceAround,
                          children: [
                            Column(
                              children: [
                                Text(
                                  '$totalPlannedCount',
                                  style: GoogleFonts.outfit(
                                    fontSize: 28,
                                    fontWeight: FontWeight.w900,
                                    color: const Color(0xFF1E293B),
                                  ),
                                ),
                                Text(
                                  'Total Toko',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    color: Colors.blueGrey,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                              ],
                            ),
                            Container(
                              width: 1,
                              height: 36,
                              color: Colors.grey[200],
                            ),
                            Column(
                              children: [
                                Text(
                                  '$completedCount/$totalPlannedCount',
                                  style: GoogleFonts.outfit(
                                    fontSize: 28,
                                    fontWeight: FontWeight.w900,
                                    color: primaryEmerald,
                                  ),
                                ),
                                Text(
                                  'Selesai',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    color: Colors.blueGrey,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                              ],
                            ),
                            Container(
                              width: 1,
                              height: 36,
                              color: Colors.grey[200],
                            ),
                            Column(
                              children: [
                                Text(
                                  _totalKm.toStringAsFixed(1),
                                  style: GoogleFonts.outfit(
                                    fontSize: 28,
                                    fontWeight: FontWeight.w900,
                                    color: const Color(0xFF1E293B),
                                  ),
                                ),
                                Text(
                                  'Total KM',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    color: Colors.blueGrey,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                        const SizedBox(height: 20),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton.icon(
                            onPressed: _displayedBanners.isEmpty
                                ? null
                                : _openGoogleMapsMultiRoute,
                            icon: const Icon(Icons.navigation_rounded),
                            label: Text(
                              _sortMode == ScheduleSortMode.optimal
                                  ? 'Mulai Navigasi Rute Optimal (G-Maps)'
                                  : 'Mulai Perjalanan (G-Maps)',
                              style: GoogleFonts.outfit(
                                fontWeight: FontWeight.bold,
                                fontSize: 13,
                              ),
                            ),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.blue[600],
                              foregroundColor: Colors.white,
                              padding: const EdgeInsets.symmetric(vertical: 13),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(12),
                              ),
                              elevation: 0,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 18),

                  // STATUS FILTER SEGMENTED BAR (Semua, Belum Pasang, Selesai)
                  Container(
                    padding: const EdgeInsets.all(4),
                    decoration: BoxDecoration(
                      color: const Color(0xFFE2E8F0),
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: Row(
                      children: [
                        _buildStatusFilterTab(
                          label: 'Semua',
                          count: totalPlannedCount,
                          filter: ScheduleStatusFilter.all,
                          icon: Icons.list_alt_rounded,
                        ),
                        _buildStatusFilterTab(
                          label: 'Belum',
                          count: pendingCount,
                          filter: ScheduleStatusFilter.pending,
                          icon: Icons.pending_actions_rounded,
                        ),
                        _buildStatusFilterTab(
                          label: 'Selesai',
                          count: completedCount,
                          filter: ScheduleStatusFilter.completed,
                          icon: Icons.check_circle_rounded,
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 14),

                  // SORTING FILTER TABS
                  Row(
                    children: [
                      Text(
                        'URUTAN RUTE:',
                        style: GoogleFonts.outfit(
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                          color: const Color(0xFF64748B),
                          letterSpacing: 1.2,
                        ),
                      ),
                      const Spacer(),
                      _buildSortChip(
                        label: 'Optimal (TSP)',
                        mode: ScheduleSortMode.optimal,
                        icon: Icons.alt_route_rounded,
                      ),
                      const SizedBox(width: 6),
                      _buildSortChip(
                        label: 'Terdekat',
                        mode: ScheduleSortMode.nearest,
                        icon: Icons.near_me_rounded,
                      ),
                    ],
                  ),

                  const SizedBox(height: 12),

                  if (_displayedBanners.isEmpty)
                    Container(
                      padding: const EdgeInsets.symmetric(
                        vertical: 40,
                        horizontal: 20,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: Colors.grey[200]!),
                      ),
                      child: Column(
                        children: [
                          Icon(
                            _statusFilter == ScheduleStatusFilter.completed
                                ? Icons.task_alt_rounded
                                : Icons.event_note_rounded,
                            size: 48,
                            color: Colors.grey[400],
                          ),
                          const SizedBox(height: 16),
                          Text(
                            _rawPlannedBanners.isEmpty
                                ? 'Belum ada rencana pemasangan hari ini.'
                                : (_statusFilter ==
                                          ScheduleStatusFilter.completed
                                      ? 'Belum ada spanduk yang selesai dipasang hari ini.'
                                      : 'Semua tugas spanduk hari ini telah selesai dipasang!'),
                            textAlign: TextAlign.center,
                            style: GoogleFonts.outfit(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: Colors.black87,
                            ),
                          ),
                          const SizedBox(height: 6),
                          Text(
                            _rawPlannedBanners.isEmpty
                                ? 'Pilih spanduk yang siap pasang untuk mulai mengerjakan rute pemasangan hari ini.'
                                : 'Gunakan tab filter lain untuk melihat data pemasangan hari ini.',
                            textAlign: TextAlign.center,
                            style: GoogleFonts.outfit(
                              fontSize: 12,
                              color: Colors.grey[600],
                              height: 1.4,
                            ),
                          ),
                          if (_rawPlannedBanners.isEmpty) ...[
                            const SizedBox(height: 18),
                            ElevatedButton.icon(
                              onPressed: () {
                                Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (_) =>
                                        const OperatorSpandukPlanPage(),
                                  ),
                                ).then((_) => _fetchSchedule());
                              },
                              icon: const Icon(
                                Icons.add_task_rounded,
                                size: 18,
                              ),
                              label: Text(
                                'Pilih & Ambil Tugas Spanduk',
                                style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 13,
                                ),
                              ),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: primaryEmerald,
                                foregroundColor: Colors.white,
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 20,
                                  vertical: 12,
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                elevation: 0,
                              ),
                            ),
                          ],
                        ],
                      ),
                    )
                  else
                    ..._displayedBanners.asMap().entries.map((entry) {
                      final index = entry.key;
                      final banner = entry.value;
                      final status = banner['status']?.toString().toLowerCase();
                      final isDone =
                          status == 'installed' || status == 'terpasang';

                      final shopName = _getShopName(banner);
                      final shopAddress = _getShopAddress(banner);
                      final distFromMe = banner['_dist_from_me'] as double?;
                      final lat = _getLatitude(banner);
                      final lng = _getLongitude(banner);
                      final bannerType = banner['type'] ?? 'Spanduk';
                      final length = banner['size_length'] ?? 0;
                      final width = banner['size_width'] ?? 0;
                      final qty = banner['qty'] ?? 1;
                      final salesName = banner['sales_name'] ?? '-';
                      final shopPhone = banner['shop_phone']?.toString();
                      final salesPhone = banner['sales_phone']?.toString();

                      final rawShopPhoto =
                          banner['shop_photo']?.toString() ??
                          banner['shop']?['photo_url']?.toString() ??
                          '';
                      final rawPhotoInstalled =
                          banner['photo_installed']?.toString() ?? '';
                      final isPhotoInstalled =
                          isDone && rawPhotoInstalled.isNotEmpty;
                      final rawThumbnail = isPhotoInstalled
                          ? rawPhotoInstalled
                          : rawShopPhoto;
                      final displayThumbnail = AppConstants.buildImageUrl(
                        rawThumbnail,
                      );

                      String? installedTimeFormatted;
                      if (isDone && banner['installed_at'] != null) {
                        try {
                          final dt = DateTime.parse(
                            banner['installed_at'].toString(),
                          ).toLocal();
                          installedTimeFormatted =
                              '${DateFormat('HH:mm').format(dt)} WIB';
                        } catch (_) {}
                      }

                      return Card(
                        elevation: 0,
                        margin: const EdgeInsets.only(bottom: 12),
                        color: Colors.white,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                          side: BorderSide(
                            color: isDone
                                ? primaryEmerald.withValues(alpha: 0.5)
                                : const Color(0xFFE2E8F0),
                            width: isDone ? 1.5 : 1,
                          ),
                        ),
                        child: Padding(
                          padding: const EdgeInsets.all(16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              // Top Row: Stop Badge + Status
                              Row(
                                children: [
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 4,
                                    ),
                                    decoration: BoxDecoration(
                                      color: isDone
                                          ? primaryEmerald.withValues(
                                              alpha: 0.12,
                                            )
                                          : const Color(
                                              0xFF0F172A,
                                            ).withValues(alpha: 0.08),
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(
                                          _sortMode == ScheduleSortMode.optimal
                                              ? Icons.navigation_rounded
                                              : Icons.place_rounded,
                                          size: 13,
                                          color: isDone
                                              ? primaryEmerald
                                              : const Color(0xFF0F172A),
                                        ),
                                        const SizedBox(width: 4),
                                        Text(
                                          'Stop #${index + 1}',
                                          style: GoogleFonts.outfit(
                                            fontSize: 11,
                                            fontWeight: FontWeight.w800,
                                            color: isDone
                                                ? primaryEmerald
                                                : const Color(0xFF0F172A),
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  if (distFromMe != null &&
                                      distFromMe != double.infinity)
                                    InkWell(
                                      onTap: () => _openSingleLocation(banner),
                                      borderRadius: BorderRadius.circular(8),
                                      child: Container(
                                        padding: const EdgeInsets.symmetric(
                                          horizontal: 8,
                                          vertical: 4,
                                        ),
                                        decoration: BoxDecoration(
                                          color: Colors.blue.withValues(
                                            alpha: 0.1,
                                          ),
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                          border: Border.all(
                                            color: Colors.blue.withValues(
                                              alpha: 0.2,
                                            ),
                                          ),
                                        ),
                                        child: Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: [
                                            Icon(
                                              Icons.directions_rounded,
                                              size: 13,
                                              color: Colors.blue[700],
                                            ),
                                            const SizedBox(width: 4),
                                            Text(
                                              _formatDistance(distFromMe),
                                              style: GoogleFonts.outfit(
                                                fontSize: 11,
                                                fontWeight: FontWeight.bold,
                                                color: Colors.blue[800],
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ),
                                  const Spacer(),
                                  if (isDone)
                                    Container(
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 8,
                                        vertical: 4,
                                      ),
                                      decoration: BoxDecoration(
                                        color: primaryEmerald.withValues(
                                          alpha: 0.15,
                                        ),
                                        borderRadius: BorderRadius.circular(8),
                                      ),
                                      child: Row(
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          const Icon(
                                            Icons.check_circle_rounded,
                                            size: 12,
                                            color: darkEmerald,
                                          ),
                                          const SizedBox(width: 4),
                                          Text(
                                            installedTimeFormatted != null
                                                ? 'TERPASANG ($installedTimeFormatted)'
                                                : 'TERPASANG',
                                            style: GoogleFonts.outfit(
                                              fontSize: 10,
                                              fontWeight: FontWeight.w800,
                                              color: darkEmerald,
                                            ),
                                          ),
                                        ],
                                      ),
                                    )
                                  else
                                    Container(
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 8,
                                        vertical: 4,
                                      ),
                                      decoration: BoxDecoration(
                                        color: Colors.orange.withValues(
                                          alpha: 0.12,
                                        ),
                                        borderRadius: BorderRadius.circular(8),
                                      ),
                                      child: Text(
                                        'BELUM PASANG',
                                        style: GoogleFonts.outfit(
                                          fontSize: 10,
                                          fontWeight: FontWeight.w800,
                                          color: Colors.orange[800],
                                        ),
                                      ),
                                    ),
                                ],
                              ),

                              const SizedBox(height: 12),

                              // Store Photo / Installed Photo Thumbnail + Store Name & Address
                              Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  InkWell(
                                    onTap: () =>
                                        _showStorePhotoDialog(context, banner),
                                    borderRadius: BorderRadius.circular(10),
                                    child: Container(
                                      width: 48,
                                      height: 48,
                                      decoration: BoxDecoration(
                                        color: const Color(0xFFF1F5F9),
                                        borderRadius: BorderRadius.circular(10),
                                        border: Border.all(
                                          color: isDone
                                              ? primaryEmerald.withValues(
                                                  alpha: 0.5,
                                                )
                                              : const Color(0xFFCBD5E1),
                                        ),
                                      ),
                                      clipBehavior: Clip.antiAlias,
                                      child: Stack(
                                        fit: StackFit.expand,
                                        children: [
                                          displayThumbnail.isNotEmpty
                                              ? AppNetworkImage(
                                                  imageUrl: displayThumbnail,
                                                  fit: BoxFit.cover,
                                                  placeholder: const Center(
                                                    child: Icon(
                                                      Icons.storefront_rounded,
                                                      size: 22,
                                                      color: Color(0xFF94A3B8),
                                                    ),
                                                  ),
                                                  errorWidget: const Center(
                                                    child: Icon(
                                                      Icons.storefront_rounded,
                                                      size: 22,
                                                      color: Color(0xFF94A3B8),
                                                    ),
                                                  ),
                                                  memCacheWidth: 200,
                                                )
                                              : Center(
                                                  child: Icon(
                                                    isDone
                                                        ? Icons
                                                              .check_circle_rounded
                                                        : Icons
                                                              .storefront_rounded,
                                                    size: 24,
                                                    color: primaryEmerald,
                                                  ),
                                                ),
                                          if (isPhotoInstalled)
                                            Positioned(
                                              bottom: 0,
                                              right: 0,
                                              child: Container(
                                                padding: const EdgeInsets.all(
                                                  2,
                                                ),
                                                decoration: const BoxDecoration(
                                                  color: primaryEmerald,
                                                  borderRadius:
                                                      BorderRadius.only(
                                                        topLeft:
                                                            Radius.circular(6),
                                                      ),
                                                ),
                                                child: const Icon(
                                                  Icons.check_rounded,
                                                  size: 10,
                                                  color: Colors.white,
                                                ),
                                              ),
                                            ),
                                        ],
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 10),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Row(
                                          children: [
                                            Expanded(
                                              child: Text(
                                                shopName,
                                                style: GoogleFonts.outfit(
                                                  fontWeight: FontWeight.bold,
                                                  fontSize: 16,
                                                  color: const Color(
                                                    0xFF0F172A,
                                                  ),
                                                ),
                                              ),
                                            ),
                                            if (lat != null && lng != null)
                                              Container(
                                                margin: const EdgeInsets.only(
                                                  left: 4,
                                                ),
                                                padding:
                                                    const EdgeInsets.symmetric(
                                                      horizontal: 5,
                                                      vertical: 2,
                                                    ),
                                                decoration: BoxDecoration(
                                                  color: Colors.green
                                                      .withValues(alpha: 0.1),
                                                  borderRadius:
                                                      BorderRadius.circular(4),
                                                ),
                                                child: Row(
                                                  mainAxisSize:
                                                      MainAxisSize.min,
                                                  children: [
                                                    const Icon(
                                                      Icons.gps_fixed_rounded,
                                                      size: 10,
                                                      color: Colors.green,
                                                    ),
                                                    const SizedBox(width: 2),
                                                    Text(
                                                      'GPS',
                                                      style: GoogleFonts.outfit(
                                                        fontSize: 9,
                                                        fontWeight:
                                                            FontWeight.bold,
                                                        color:
                                                            Colors.green[800],
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ),
                                          ],
                                        ),
                                        const SizedBox(height: 2),
                                        Row(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            Icon(
                                              Icons.location_on_outlined,
                                              size: 13,
                                              color: Colors.grey[500],
                                            ),
                                            const SizedBox(width: 3),
                                            Expanded(
                                              child: Text(
                                                shopAddress,
                                                style: GoogleFonts.outfit(
                                                  fontSize: 12,
                                                  color: const Color(
                                                    0xFF64748B,
                                                  ),
                                                ),
                                                maxLines: 2,
                                                overflow: TextOverflow.ellipsis,
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

                              // Specs & Sales Info Chips
                              Wrap(
                                spacing: 6,
                                runSpacing: 6,
                                children: [
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 3,
                                    ),
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFF1F5F9),
                                      borderRadius: BorderRadius.circular(6),
                                    ),
                                    child: Text(
                                      '$bannerType • ${length}x$width cm ($qty pcs)',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w600,
                                        color: const Color(0xFF334155),
                                      ),
                                    ),
                                  ),
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 3,
                                    ),
                                    decoration: BoxDecoration(
                                      color: Colors.blue.withValues(
                                        alpha: 0.08,
                                      ),
                                      borderRadius: BorderRadius.circular(6),
                                    ),
                                    child: Text(
                                      'Sales: $salesName',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w600,
                                        color: Colors.blue[800],
                                      ),
                                    ),
                                  ),
                                ],
                              ),

                              const SizedBox(height: 12),

                              // 1. Communication & Store Photo Action Row
                              Row(
                                children: [
                                  Expanded(
                                    child: OutlinedButton.icon(
                                      onPressed: () => _showStorePhotoDialog(
                                        context,
                                        banner,
                                      ),
                                      icon: Icon(
                                        isDone
                                            ? Icons.collections_rounded
                                            : Icons.photo_camera_rounded,
                                        size: 13,
                                      ),
                                      label: Text(
                                        isDone ? 'Foto & Bukti' : 'Foto Toko',
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                      style: OutlinedButton.styleFrom(
                                        foregroundColor: const Color(
                                          0xFF334155,
                                        ),
                                        side: const BorderSide(
                                          color: Color(0xFFCBD5E1),
                                        ),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                        ),
                                        padding: const EdgeInsets.symmetric(
                                          vertical: 8,
                                        ),
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 6),
                                  Expanded(
                                    child: OutlinedButton.icon(
                                      onPressed: () =>
                                          _contactStore(shopPhone, shopName),
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
                                        side: const BorderSide(
                                          color: primaryEmerald,
                                        ),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                        ),
                                        padding: const EdgeInsets.symmetric(
                                          vertical: 8,
                                        ),
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 6),
                                  Expanded(
                                    child: OutlinedButton.icon(
                                      onPressed: () =>
                                          _contactSales(salesPhone, shopName),
                                      icon: const Icon(
                                        Icons.person_rounded,
                                        size: 13,
                                      ),
                                      label: Text(
                                        'Chat Sales',
                                        style: GoogleFonts.outfit(
                                          fontSize: 11,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                      style: OutlinedButton.styleFrom(
                                        foregroundColor: Colors.blue[700],
                                        side: BorderSide(
                                          color: Colors.blue[600]!,
                                        ),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                        ),
                                        padding: const EdgeInsets.symmetric(
                                          vertical: 8,
                                        ),
                                      ),
                                    ),
                                  ),
                                ],
                              ),

                              const SizedBox(height: 8),

                              // 2. Operational Action Buttons (Rute Google Maps & Pasang Spanduk)
                              Row(
                                children: [
                                  Expanded(
                                    child: OutlinedButton.icon(
                                      onPressed: () =>
                                          _openSingleLocation(banner),
                                      icon: const Icon(
                                        Icons.navigation_rounded,
                                        size: 14,
                                        color: Colors.blue,
                                      ),
                                      label: Text(
                                        'Buka Rute Maps',
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: FontWeight.bold,
                                          color: Colors.blue[800],
                                        ),
                                      ),
                                      style: OutlinedButton.styleFrom(
                                        backgroundColor: Colors.blue.withValues(
                                          alpha: 0.05,
                                        ),
                                        side: BorderSide(
                                          color: Colors.blue.withValues(
                                            alpha: 0.3,
                                          ),
                                        ),
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                        ),
                                        padding: const EdgeInsets.symmetric(
                                          vertical: 10,
                                        ),
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Expanded(
                                    child: ElevatedButton.icon(
                                      onPressed: () => _processInstall(banner),
                                      icon: Icon(
                                        isDone
                                            ? Icons.check_circle_rounded
                                            : Icons.camera_alt_rounded,
                                        size: 14,
                                      ),
                                      label: Text(
                                        isDone
                                            ? 'Update Bukti Pasang'
                                            : 'Pasang Spanduk',
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                      style: ElevatedButton.styleFrom(
                                        backgroundColor: isDone
                                            ? const Color(0xFF334155)
                                            : primaryEmerald,
                                        foregroundColor: Colors.white,
                                        shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(
                                            8,
                                          ),
                                        ),
                                        padding: const EdgeInsets.symmetric(
                                          vertical: 10,
                                        ),
                                        elevation: 0,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),
                      );
                    }),
                  const SizedBox(height: 70), // FAB padding
                ],
              ),
            ),
    );
  }

  Widget _buildStatusFilterTab({
    required String label,
    required int count,
    required ScheduleStatusFilter filter,
    required IconData icon,
  }) {
    final isSelected = _statusFilter == filter;
    return Expanded(
      child: InkWell(
        onTap: () {
          if (_statusFilter == filter) return;
          setState(() {
            _statusFilter = filter;
            _applySortAndCalculate();
          });
        },
        borderRadius: BorderRadius.circular(10),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 8),
          decoration: BoxDecoration(
            color: isSelected ? Colors.white : Colors.transparent,
            borderRadius: BorderRadius.circular(10),
            boxShadow: isSelected
                ? [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.06),
                      blurRadius: 4,
                      offset: const Offset(0, 2),
                    ),
                  ]
                : null,
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                icon,
                size: 13,
                color: isSelected ? primaryEmerald : const Color(0xFF64748B),
              ),
              const SizedBox(width: 4),
              Text(
                '$label ($count)',
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                  color: isSelected ? primaryEmerald : const Color(0xFF64748B),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSortChip({
    required String label,
    required ScheduleSortMode mode,
    required IconData icon,
  }) {
    final isSelected = _sortMode == mode;
    return InkWell(
      onTap: () {
        setState(() {
          _sortMode = mode;
          _applySortAndCalculate();
        });
      },
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
        decoration: BoxDecoration(
          color: isSelected ? primaryEmerald : Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: isSelected ? primaryEmerald : Colors.grey[300]!,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              icon,
              size: 12,
              color: isSelected ? Colors.white : const Color(0xFF64748B),
            ),
            const SizedBox(width: 4),
            Text(
              label,
              style: GoogleFonts.outfit(
                fontSize: 11,
                fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                color: isSelected ? Colors.white : const Color(0xFF64748B),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
