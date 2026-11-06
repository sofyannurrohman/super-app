import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:geolocator/geolocator.dart';
import 'package:get_it/get_it.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/widgets/app_network_image.dart';
import 'package:hris_app/features/sales/data/services/banner_order_api_service.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';

class OperatorSpandukPlanPage extends StatefulWidget {
  const OperatorSpandukPlanPage({super.key});

  @override
  State<OperatorSpandukPlanPage> createState() =>
      _OperatorSpandukPlanPageState();
}

class _OperatorSpandukPlanPageState extends State<OperatorSpandukPlanPage> {
  final _apiService = GetIt.I<BannerOrderApiService>();
  bool _isLoading = true;
  List<dynamic> _allLoadedBanners = [];
  List<dynamic> _availableBanners = [];
  final List<dynamic> _selectedBanners = [];
  String _operatorBranchId = '';
  String _operatorBranchName = '';
  String _operatorOperatingBranchId = '';
  String _operatorOperatingBranchName = '';
  Position? _currentPosition;
  String _searchQuery = '';
  bool _sortByNearest = true;
  bool _showAllBranches = false;

  // Regional Filter & Grouping States
  String _selectedDistrictFilter = 'ALL';
  bool _groupByDistrict = true;

  static const Color primaryEmerald = Color(0xFF047857);
  static const Color darkEmerald = Color(0xFF064E3B);

  @override
  void initState() {
    super.initState();
    _fetchAvailableBanners();
  }

  /// Apakah cabang operator berhasil dideteksi (ID atau nama).
  bool get _hasOperatorBranch =>
      _operatorBranchId.isNotEmpty ||
      _operatorOperatingBranchId.isNotEmpty ||
      _operatorBranchName.trim().isNotEmpty ||
      _operatorOperatingBranchName.trim().isNotEmpty;

  bool _isBannerInOperatorBranch(dynamic b) {
    // Fail-closed: jika cabang operator tidak diketahui, jangan loloskan semua banner.
    if (!_hasOperatorBranch) return false;

    final opIds = <String>{
      _operatorBranchId,
      _operatorOperatingBranchId,
    }.where((e) => e.isNotEmpty).map((e) => e.toLowerCase()).toSet();
    final bId =
        (b['branch_id'] ?? b['BranchID'])?.toString().trim().toLowerCase() ??
        '';

    // 1. Strict ID Match (branch_id atau operating_branch_id operator)
    if (bId.isNotEmpty && opIds.isNotEmpty) {
      return opIds.contains(bId);
    }

    // 2. Fallback nama cabang: harus sama persis (case-insensitive)
    final bBranchName =
        (b['_branch_name'] ??
                b['sales_branch'] ??
                b['branch_name'] ??
                (b['branch'] is Map ? b['branch']['name'] : null) ??
                (b['Branch'] is Map ? b['Branch']['Name'] : null) ??
                '')
            .toString()
            .trim()
            .toLowerCase();
    if (bBranchName.isEmpty) return false;
    final opNames = <String>{
      _operatorBranchName,
      _operatorOperatingBranchName,
    }.map((e) => e.trim().toLowerCase()).where((e) => e.isNotEmpty);
    return opNames.contains(bBranchName);
  }

  void _applyBranchFilter() {
    if (!_showAllBranches) {
      _availableBanners = _allLoadedBanners
          .where(_isBannerInOperatorBranch)
          .toList();
    } else {
      _availableBanners = List.from(_allLoadedBanners);
    }
  }

  void _toggleBranchScope(bool showAll) {
    if (_showAllBranches == showAll) return;
    setState(() {
      _showAllBranches = showAll;
      _selectedDistrictFilter = 'ALL';
      _applyBranchFilter();
    });
  }

  String _capitalize(String text) {
    if (text.trim().isEmpty) return text;
    return text
        .trim()
        .split(RegExp(r'\s+'))
        .map((word) {
          if (word.isEmpty) return word;
          return word[0].toUpperCase() + word.substring(1).toLowerCase();
        })
        .join(' ');
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

  String _getDistrict(dynamic b) {
    // 1. Cek field langsung dari API / Store
    final direct =
        (b['district'] ??
                b['District'] ??
                (b['shop'] is Map ? b['shop']['district'] : null) ??
                (b['Shop'] is Map ? b['Shop']['District'] : null))
            ?.toString()
            .trim();
    if (direct != null && direct.isNotEmpty && direct != '-') {
      final name = direct.toLowerCase().startsWith('kec')
          ? direct
          : 'Kec. ${_capitalize(direct)}';
      return name;
    }

    // 2. Fallback regex dari alamat toko
    final addr = _getShopAddress(b);
    if (addr.isNotEmpty && addr != '-') {
      final match = RegExp(
        r'(?:kec(?:amatan)?\.?\s*)([a-zA-Z0-9\s]+?)(?:,|$|\.|\s(?:kab|kabupaten|kota|desa|kel))',
        caseSensitive: false,
      ).firstMatch(addr);
      if (match != null && match.group(1) != null) {
        final extracted = match.group(1)!.trim();
        if (extracted.isNotEmpty) {
          return 'Kec. ${_capitalize(extracted)}';
        }
      }
    }

    return 'Wilayah Lainnya';
  }

  String _getRegency(dynamic b) {
    final direct =
        (b['regency'] ??
                b['Regency'] ??
                (b['shop'] is Map ? b['shop']['regency'] : null) ??
                (b['Shop'] is Map ? b['Shop']['Regency'] : null))
            ?.toString()
            .trim();
    if (direct != null && direct.isNotEmpty && direct != '-') {
      return _capitalize(direct);
    }

    final addr = _getShopAddress(b);
    if (addr.isNotEmpty && addr != '-') {
      final match = RegExp(
        r'(?:kab(?:upaten)?|kota)\.?\s*([a-zA-Z0-9\s]+?)(?:,|$|\.)',
        caseSensitive: false,
      ).firstMatch(addr);
      if (match != null && match.group(1) != null) {
        final extracted = match.group(1)!.trim();
        if (extracted.isNotEmpty) {
          return _capitalize(extracted);
        }
      }
    }
    return '';
  }

  String _getVillage(dynamic b) {
    final direct =
        (b['village'] ??
                b['Village'] ??
                (b['shop'] is Map ? b['shop']['village'] : null) ??
                (b['Shop'] is Map ? b['Shop']['Village'] : null))
            ?.toString()
            .trim();
    if (direct != null && direct.isNotEmpty && direct != '-') {
      return _capitalize(direct);
    }

    final addr = _getShopAddress(b);
    if (addr.isNotEmpty && addr != '-') {
      final match = RegExp(
        r'(?:desa|kel(?:urahan)?)\.?\s*([a-zA-Z0-9\s]+?)(?:,|$|\.|\skec)',
        caseSensitive: false,
      ).firstMatch(addr);
      if (match != null && match.group(1) != null) {
        final extracted = match.group(1)!.trim();
        if (extracted.isNotEmpty) {
          return _capitalize(extracted);
        }
      }
    }
    return '';
  }

  String _getBranchName(dynamic b) {
    final direct =
        (b['branch_name'] ??
                b['BranchName'] ??
                (b['branch'] is Map ? b['branch']['name'] : null) ??
                (b['Branch'] is Map ? b['Branch']['Name'] : null) ??
                b['sales_branch'])
            ?.toString()
            .trim();
    if (direct != null && direct.isNotEmpty && direct != '-') {
      return direct;
    }
    return '';
  }

  String _getSalesName(dynamic b) {
    final direct =
        (b['sales_name'] ??
                b['SalesName'] ??
                (b['sales'] is Map ? b['sales']['name'] : null) ??
                (b['Sales'] is Map ? b['Sales']['Name'] : null) ??
                (b['employee'] is Map ? b['employee']['name'] : null))
            ?.toString()
            .trim();
    if (direct != null && direct.isNotEmpty && direct != '-') {
      return direct;
    }
    return '';
  }

  String _getOwnerName(dynamic b) {
    final direct =
        (b['owner_name'] ??
                b['OwnerName'] ??
                (b['shop'] is Map ? b['shop']['owner_name'] : null) ??
                (b['Shop'] is Map ? b['Shop']['OwnerName'] : null))
            ?.toString()
            .trim();
    if (direct != null && direct.isNotEmpty && direct != '-') {
      return direct;
    }
    return '';
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
        timeLimit: const Duration(seconds: 6),
      );
    } catch (_) {
      return null;
    }
  }

  void _applyOperatorBranchFromProfile(Map<String, dynamic>? profile) {
    String s(dynamic v) {
      final t = v?.toString().trim() ?? '';
      return (t == 'null') ? '' : t;
    }

    final branchMap = profile?['branch'] is Map ? profile!['branch'] : null;
    final opMap = profile?['operating_branch'] is Map
        ? profile!['operating_branch']
        : null;
    _operatorBranchId = s(profile?['branch_id'] ?? profile?['BranchID']);
    _operatorBranchName = s(
      profile?['branch_name'] ?? branchMap?['name'] ?? profile?['BranchName'],
    );
    _operatorOperatingBranchId = s(
      profile?['operating_branch_id'] ?? profile?['OperatingBranchID'],
    );
    _operatorOperatingBranchName = s(opMap?['name']);
  }

  Future<void> _fetchAvailableBanners() async {
    setState(() => _isLoading = true);
    try {
      final authState = context.read<AuthBloc>().state;
      Map<String, dynamic>? profile = authState is Authenticated
          ? authState.userProfile
          : null;

      // Profil bisa masih null (AuthBloc memuat profil di latar belakang) -> ambil langsung.
      if (profile == null) {
        try {
          final res = await _apiService.apiClient.client.get(
            'employees/profile',
          );
          final data = res.data is Map ? res.data['data'] : null;
          if (data is Map) profile = Map<String, dynamic>.from(data);
        } catch (_) {}
      }
      _applyOperatorBranchFromProfile(profile);

      // 1. Ambil last known GPS (0ms) & request banners dengan status installation secara chunked paralel
      final bannersFuture = _apiService.getAllBannersChunked(
        status: 'installation',
      );
      final lastPosFuture = _getLastKnownPosition();

      final combined = await Future.wait([bannersFuture, lastPosFuture]);
      final rawResults = combined[0] as List<dynamic>;
      _currentPosition = combined[1] as Position?;

      // Deduplikasi data berdasarkan banner ID untuk memastikan integritas data
      final seenIds = <String>{};
      final results = <dynamic>[];
      for (var item in rawResults) {
        final id = (item['id'] ?? item['ID'] ?? '').toString().trim();
        if (id.isEmpty || seenIds.add(id)) {
          results.add(item);
        }
      }

      // 2. Pre-compute fields (Wilayah, Nama, Alamat, Jarak, Search Key, Cabang) sekali saja
      final allFiltered = <dynamic>[];
      for (var b in results) {
        final status = b['status']?.toString().toLowerCase() ?? '';

        // Abaikan banner yang sudah terpasang/selesai
        if (status == 'installed' ||
            status == 'terpasang' ||
            status == 'selesai' ||
            status == 'selesai_dipasang') {
          continue;
        }

        final shopName = _getShopName(b);
        final shopAddr = _getShopAddress(b);
        final district = _getDistrict(b);
        final regency = _getRegency(b);
        final village = _getVillage(b);
        final branchName = _getBranchName(b);
        final salesName = _getSalesName(b);
        final ownerName = _getOwnerName(b);
        final bannerNotes = (b['notes'] ?? b['Notes'] ?? '').toString().trim();
        final bannerType = (b['type'] ?? b['banner_type'] ?? b['Type'] ?? '')
            .toString()
            .trim();
        final bannerId = (b['id'] ?? b['ID'] ?? '').toString().trim();
        final lat = _getLatitude(b);
        final lng = _getLongitude(b);

        b['_shop_name'] = shopName;
        b['_shop_address'] = shopAddr;
        b['_district'] = district;
        b['_regency'] = regency;
        b['_village'] = village;
        b['_branch_name'] = branchName;
        b['_sales_name'] = salesName;
        b['_owner_name'] = ownerName;
        b['_lat'] = lat;
        b['_lng'] = lng;
        b['_search_key'] =
            '$shopName $ownerName $salesName $shopAddr $district $regency $village $branchName $bannerNotes $bannerType $bannerId'
                .toLowerCase();

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
        allFiltered.add(b);
      }

      if (mounted) {
        setState(() {
          _allLoadedBanners = allFiltered;
          // Jika tidak ada tugas di cabang operator (atau cabang tidak terdeteksi),
          // langsung tampilkan Semua Wilayah agar daftar tidak kosong.
          if (!_showAllBranches &&
              allFiltered.isNotEmpty &&
              !allFiltered.any(_isBannerInOperatorBranch)) {
            _showAllBranches = true;
          }
          _applyBranchFilter();
          _isLoading = false;
        });
      }

      // 3. Update GPS presisi di background secara non-blocking
      _updatePreciseLocationInBackground();
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Gagal memuat spanduk: $e')));
      }
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
                  bool updated = false;
                  for (var b in _allLoadedBanners) {
                    final lat = b['_lat'] as double?;
                    final lng = b['_lng'] as double?;
                    if (lat != null && lng != null) {
                      b['_dist_from_me'] = Geolocator.distanceBetween(
                        pos.latitude,
                        pos.longitude,
                        lat,
                        lng,
                      );
                      updated = true;
                    }
                  }
                  if (updated && mounted) {
                    setState(() {
                      _applyBranchFilter();
                    });
                  }
                })
                .catchError((_) {});
          });
        })
        .catchError((_) {});
  }

  void _toggleSelection(dynamic banner) {
    setState(() {
      final exists = _selectedBanners.any((b) => b['id'] == banner['id']);
      if (exists) {
        _selectedBanners.removeWhere((b) => b['id'] == banner['id']);
      } else {
        _selectedBanners.add(banner);
      }
    });
  }

  void _toggleSelectAll(List<dynamic> items) {
    setState(() {
      final allSelected =
          items.isNotEmpty &&
          items.every(
            (item) => _selectedBanners.any((sb) => sb['id'] == item['id']),
          );
      if (allSelected) {
        final idsToRemove = items.map((e) => e['id']).toSet();
        _selectedBanners.removeWhere((b) => idsToRemove.contains(b['id']));
      } else {
        for (var item in items) {
          if (!_selectedBanners.any((sb) => sb['id'] == item['id'])) {
            _selectedBanners.add(item);
          }
        }
      }
    });
  }

  void _toggleSelectDistrict(List<dynamic> districtBanners) {
    setState(() {
      final allSelected = districtBanners.every(
        (b) => _selectedBanners.any((sb) => sb['id'] == b['id']),
      );
      if (allSelected) {
        final idsToRemove = districtBanners.map((b) => b['id']).toSet();
        _selectedBanners.removeWhere((b) => idsToRemove.contains(b['id']));
      } else {
        for (var b in districtBanners) {
          if (!_selectedBanners.any((sb) => sb['id'] == b['id'])) {
            _selectedBanners.add(b);
          }
        }
      }
    });
  }

  Future<void> _openGoogleMapsRoute() async {
    if (_selectedBanners.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Pilih minimal 1 spanduk untuk melihat rute'),
        ),
      );
      return;
    }

    try {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (c) => const Center(
          child: CircularProgressIndicator(color: primaryEmerald),
        ),
      );

      Position? position = _currentPosition;
      position ??= await _determinePosition();
      if (mounted) Navigator.pop(context);

      final validCoords = <Map<String, double>>[];
      for (var b in _selectedBanners) {
        final lat = _getLatitude(b);
        final lng = _getLongitude(b);
        if (lat != null && lng != null) {
          validCoords.add({'lat': lat, 'lng': lng});
        }
      }

      if (validCoords.isEmpty) {
        throw Exception('Data lokasi toko tidak lengkap untuk rute ini');
      }

      String origin = '';
      if (position != null) {
        origin = '${position.latitude},${position.longitude}';
      } else {
        // Keep the first store in the list. For a single selected store,
        // removing it would leave no destination and trigger a false
        // "Toko tujuan tidak valid" error.
        origin = '${validCoords.first['lat']},${validCoords.first['lng']}';
      }

      if (validCoords.isEmpty) {
        throw Exception('Toko tujuan tidak valid');
      }

      final destination =
          '${validCoords.last['lat']},${validCoords.last['lng']}';
      final waypointList = validCoords.sublist(0, validCoords.length - 1);

      String url =
          'https://www.google.com/maps/dir/?api=1&origin=$origin&destination=$destination';
      if (waypointList.isNotEmpty) {
        final waypoints = waypointList
            .map((w) => '${w['lat']},${w['lng']}')
            .join('|');
        url += '&waypoints=$waypoints';
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

  Future<void> _savePlan() async {
    if (_selectedBanners.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Pilih minimal 1 spanduk untuk diambil')),
      );
      return;
    }

    try {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (c) => const Center(
          child: CircularProgressIndicator(color: primaryEmerald),
        ),
      );

      // 1. Prioritaskan penugasan spanduk (takeBanner) ke operator ini
      int successCount = 0;
      final List<String> errors = [];
      for (var b in _selectedBanners) {
        final bid = b['id'] ?? b['ID'];
        if (bid != null) {
          try {
            await _apiService.takeBanner(bid);
            successCount++;
          } catch (err) {
            errors.add(err.toString());
          }
        }
      }

      if (successCount == 0 && errors.isNotEmpty) {
        throw Exception(errors.first);
      }

      // 2. Opsional: Daftarkan toko ke rencana kunjungan harian jika ID toko tersedia
      final storeIds = <String>[];
      for (var b in _selectedBanners) {
        final sid =
            (b['shop_id'] ??
                    b['store_id'] ??
                    (b['shop'] is Map ? b['shop']['id'] : null) ??
                    (b['Shop'] is Map ? b['Shop']['ID'] : null) ??
                    b['ShopID'])
                ?.toString();
        if (sid != null &&
            sid.isNotEmpty &&
            sid != '-' &&
            !storeIds.contains(sid)) {
          storeIds.add(sid);
        }
      }

      if (storeIds.isNotEmpty) {
        try {
          final salesService = GetIt.I<SalesApiService>();
          await salesService.saveManualVisitPlan(
            storeIds: storeIds,
            planDate: DateTime.now(),
          );
        } catch (_) {
          // Abaikan kegagalan visit plan jika endpoint visit plan gagal,
          // karena spanduk sudah berhasil di-take.
        }
      }

      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              '$successCount Tugas pemasangan berhasil diambil dan ditugaskan kepada Anda!',
            ),
            backgroundColor: primaryEmerald,
          ),
        );
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(e.toString().replaceAll('Exception: ', '')),
            backgroundColor: Colors.red,
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
        'https://wa.me/$formattedPhone?text=Halo,%20kami%20dari%20tim%20pemasangan%20spanduk%20Karsa%20HR%20ingin%20konfirmasi%20terkait%20rencana%20pemasangan%20spanduk%20di%20toko%20$encodedShop.';
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
    final shopName = (banner['_shop_name'] as String?) ?? _getShopName(banner);
    final shopAddress =
        (banner['_shop_address'] as String?) ?? _getShopAddress(banner);
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

    int selectedTab = 0; // 0 = Foto Toko, 1 = Desain Spanduk

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

  String _formatDistance(double? meters) {
    if (meters == null || meters == double.infinity) return '';
    if (meters < 1000) {
      return '${meters.toStringAsFixed(0)} m';
    }
    return '${(meters / 1000).toStringAsFixed(1)} km';
  }

  Widget _buildStoreCard(dynamic banner) {
    final isSelected = _selectedBanners.any((b) => b['id'] == banner['id']);
    final shopName = (banner['_shop_name'] as String?) ?? _getShopName(banner);
    final shopAddress =
        (banner['_shop_address'] as String?) ?? _getShopAddress(banner);
    final distFromMe = banner['_dist_from_me'] as double?;
    final type = banner['type'] ?? 'Spanduk';
    final length = banner['size_length'] ?? 0;
    final width = banner['size_width'] ?? 0;
    final qty = banner['qty'] ?? 1;
    final salesName = banner['sales_name'] ?? '-';
    final salesBranch = banner['sales_branch'] ?? '';
    final notes = banner['notes']?.toString();
    final distStr = _formatDistance(distFromMe);
    final shopPhone = banner['shop_phone']?.toString();
    final salesPhone = banner['sales_phone']?.toString();
    final village = (banner['_village'] as String?) ?? _getVillage(banner);
    final branchName =
        (banner['_branch_name'] as String?) ?? _getBranchName(banner);
    final rawShopPhoto =
        banner['shop_photo']?.toString() ??
        banner['shop']?['photo_url']?.toString() ??
        '';

    return Card(
      elevation: isSelected ? 2 : 0,
      margin: const EdgeInsets.only(bottom: 12),
      color: isSelected ? primaryEmerald.withValues(alpha: 0.04) : Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: BorderSide(
          color: isSelected ? primaryEmerald : const Color(0xFFE2E8F0),
          width: isSelected ? 2 : 1,
        ),
      ),
      child: InkWell(
        onTap: () => _toggleSelection(banner),
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Checkbox(
                    value: isSelected,
                    onChanged: (val) => _toggleSelection(banner),
                    activeColor: primaryEmerald,
                    materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    visualDensity: VisualDensity.compact,
                  ),
                  const SizedBox(width: 4),

                  // Store Photo Thumbnail
                  InkWell(
                    onTap: () => _showStorePhotoDialog(context, banner),
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
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                shopName,
                                style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 15,
                                  color: const Color(0xFF0F172A),
                                ),
                              ),
                            ),
                            if (distStr.isNotEmpty)
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 6,
                                  vertical: 2,
                                ),
                                decoration: BoxDecoration(
                                  color: Colors.blue.withValues(alpha: 0.08),
                                  borderRadius: BorderRadius.circular(6),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Icon(
                                      Icons.near_me_rounded,
                                      size: 11,
                                      color: Colors.blue[700],
                                    ),
                                    const SizedBox(width: 2),
                                    Text(
                                      distStr,
                                      style: GoogleFonts.outfit(
                                        fontSize: 10,
                                        fontWeight: FontWeight.bold,
                                        color: Colors.blue[800],
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                          ],
                        ),
                        const SizedBox(height: 2),
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Icon(
                              Icons.location_on_outlined,
                              size: 13,
                              color: Colors.grey[500],
                            ),
                            const SizedBox(width: 4),
                            Expanded(
                              child: Text(
                                shopAddress,
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  color: const Color(0xFF64748B),
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
              Wrap(
                spacing: 6,
                runSpacing: 6,
                children: [
                  if (_showAllBranches && branchName.isNotEmpty)
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFFEFF6FF),
                        borderRadius: BorderRadius.circular(6),
                        border: Border.all(color: const Color(0xFFBFDBFE)),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.domain_rounded,
                            size: 12,
                            color: Colors.blue[800],
                          ),
                          const SizedBox(width: 4),
                          Text(
                            'Cabang $branchName',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.bold,
                              color: Colors.blue[900],
                            ),
                          ),
                        ],
                      ),
                    ),
                  if (village.isNotEmpty)
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFFFEF3C7),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.maps_home_work_rounded,
                            size: 12,
                            color: Colors.amber[900],
                          ),
                          const SizedBox(width: 4),
                          Text(
                            'Desa $village',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.bold,
                              color: Colors.amber[900],
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
                      color: const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          Icons.straighten,
                          size: 13,
                          color: Colors.grey[700],
                        ),
                        const SizedBox(width: 4),
                        Text(
                          '$type • ${length}x$width cm ($qty pcs)',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            color: const Color(0xFF334155),
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
                      color: Colors.blue.withValues(alpha: 0.08),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          Icons.person_outline,
                          size: 13,
                          color: Colors.blue[700],
                        ),
                        const SizedBox(width: 4),
                        Text(
                          salesBranch.isNotEmpty
                              ? '$salesName ($salesBranch)'
                              : salesName,
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            color: Colors.blue[800],
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              if (notes != null && notes.trim().isNotEmpty) ...[
                const SizedBox(height: 8),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF8FAFC),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: const Color(0xFFE2E8F0)),
                  ),
                  child: Text(
                    'Catatan: $notes',
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: const Color(0xFF475569),
                      fontStyle: FontStyle.italic,
                    ),
                  ),
                ),
              ],
              const SizedBox(height: 10),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: () => _showStorePhotoDialog(context, banner),
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
                      onPressed: () => _contactStore(shopPhone, shopName),
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
                      onPressed: () => _contactSales(salesPhone, shopName),
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
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                        padding: const EdgeInsets.symmetric(vertical: 8),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    // Menghitung jumlah tugas di cabang sendiri vs total semua wilayah
    final int branchCount = _allLoadedBanners
        .where(_isBannerInOperatorBranch)
        .length;
    final int allCount = _allLoadedBanners.length;

    // 1. Filter by Search Query menggunakan precomputed search key
    // Jika sedang mencari keyword, cari ke SELURUH DATA (_allLoadedBanners), jika tidak, gunakan filter wilayah aktif (_availableBanners)
    final query = _searchQuery.trim().toLowerCase();
    final sourceList = query.isNotEmpty ? _allLoadedBanners : _availableBanners;
    final searchFiltered = sourceList.where((b) {
      if (query.isEmpty) return true;
      final key = (b['_search_key'] as String?) ?? '';
      return key.contains(query);
    }).toList();

    // 2. Kumpulkan seluruh Kecamatan yang tersedia untuk Filter Chips
    final Map<String, int> districtCounts = {};
    for (var b in searchFiltered) {
      final d = (b['_district'] as String?) ?? _getDistrict(b);
      districtCounts[d] = (districtCounts[d] ?? 0) + 1;
    }
    final sortedDistrictsList = districtCounts.keys.toList()
      ..sort((a, b) {
        if (a == 'Wilayah Lainnya') return 1;
        if (b == 'Wilayah Lainnya') return -1;
        return a.compareTo(b);
      });

    // 3. Filter by Selected District Chip
    final displayedBanners = searchFiltered.where((b) {
      if (query.isNotEmpty) {
        // Jika sedang mencari kata kunci, tampilkan semua hasil pencarian lintas wilayah/kecamatan
        return true;
      }
      if (_selectedDistrictFilter == 'ALL') return true;
      final d = (b['_district'] as String?) ?? _getDistrict(b);
      return d == _selectedDistrictFilter;
    }).toList();

    // Sort by nearest if enabled
    if (_sortByNearest) {
      displayedBanners.sort((a, b) {
        final distA = (a['_dist_from_me'] as double?) ?? double.infinity;
        final distB = (b['_dist_from_me'] as double?) ?? double.infinity;
        return distA.compareTo(distB);
      });
    }

    // 4. Kelompokkan berdasarkan Kecamatan untuk Grouped View
    final Map<String, List<dynamic>> districtGroups = {};
    for (var b in displayedBanners) {
      final dist = (b['_district'] as String?) ?? _getDistrict(b);
      districtGroups.putIfAbsent(dist, () => []).add(b);
    }

    // Urutkan grup kecamatan berdasarkan jarak toko terdekat di kecamatan tersebut
    final sortedDistrictKeys = districtGroups.keys.toList()
      ..sort((a, b) {
        final minA = districtGroups[a]!.fold<double>(
          double.infinity,
          (min, banner) =>
              ((banner['_dist_from_me'] as double?) ?? double.infinity) < min
              ? (banner['_dist_from_me'] as double)
              : min,
        );
        final minB = districtGroups[b]!.fold<double>(
          double.infinity,
          (min, banner) =>
              ((banner['_dist_from_me'] as double?) ?? double.infinity) < min
              ? (banner['_dist_from_me'] as double)
              : min,
        );
        return minA.compareTo(minB);
      });

    final isAllSelected =
        displayedBanners.isNotEmpty &&
        displayedBanners.every(
          (b) => _selectedBanners.any((sb) => sb['id'] == b['id']),
        );

    // Menghitung ringkasan wilayah yang terpilih
    final selectedDistrictNames = _selectedBanners
        .map((b) => _getDistrict(b))
        .toSet()
        .toList();

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'Pilih Tugas Pemasangan',
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
          if (displayedBanners.isNotEmpty)
            TextButton(
              onPressed: () => _toggleSelectAll(displayedBanners),
              child: Text(
                isAllSelected ? 'Batal Semua' : 'Pilih Semua',
                style: GoogleFonts.outfit(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  fontSize: 12,
                ),
              ),
            ),
        ],
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(color: primaryEmerald),
            )
          : Column(
              children: [
                // Top Search & View Controller Header
                Container(
                  color: primaryEmerald,
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 14),
                  child: Column(
                    children: [
                      // 1. Search Field
                      Container(
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
                          onChanged: (v) => setState(() {
                            _searchQuery = v;
                            if (v.trim().isNotEmpty &&
                                _selectedDistrictFilter != 'ALL') {
                              _selectedDistrictFilter = 'ALL';
                            }
                          }),
                          decoration: InputDecoration(
                            hintText:
                                'Cari nama toko, sales, pemilik, jalan, kecamatan...',
                            hintStyle: GoogleFonts.outfit(
                              color: Colors.grey[400],
                            ),
                            prefixIcon: Icon(
                              Icons.search,
                              color: Colors.grey[400],
                            ),
                            suffixIcon: _searchQuery.isNotEmpty
                                ? IconButton(
                                    icon: const Icon(
                                      Icons.clear_rounded,
                                      size: 18,
                                    ),
                                    onPressed: () =>
                                        setState(() => _searchQuery = ''),
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
                      const SizedBox(height: 10),

                      // 2. Scope Selector: Cabang Saya vs Semua Wilayah
                      Container(
                        padding: const EdgeInsets.all(3),
                        decoration: BoxDecoration(
                          color: Colors.black.withValues(alpha: 0.15),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          children: [
                            // Cabang Saya
                            Expanded(
                              child: InkWell(
                                onTap: () => _toggleBranchScope(false),
                                borderRadius: BorderRadius.circular(10),
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 8,
                                  ),
                                  decoration: BoxDecoration(
                                    color: !_showAllBranches
                                        ? Colors.white
                                        : Colors.transparent,
                                    borderRadius: BorderRadius.circular(10),
                                    boxShadow: !_showAllBranches
                                        ? [
                                            BoxShadow(
                                              color: Colors.black.withValues(
                                                alpha: 0.1,
                                              ),
                                              blurRadius: 4,
                                              offset: const Offset(0, 1),
                                            ),
                                          ]
                                        : null,
                                  ),
                                  child: Row(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.domain_rounded,
                                        size: 14,
                                        color: !_showAllBranches
                                            ? darkEmerald
                                            : Colors.white.withValues(
                                                alpha: 0.8,
                                              ),
                                      ),
                                      const SizedBox(width: 6),
                                      Flexible(
                                        child: Text(
                                          !_hasOperatorBranch
                                              ? 'Cabang Saya (-)'
                                              : _operatorBranchName.isNotEmpty
                                              ? 'Cabang $_operatorBranchName ($branchCount)'
                                              : 'Cabang Saya ($branchCount)',
                                          style: GoogleFonts.outfit(
                                            fontSize: 12,
                                            fontWeight: !_showAllBranches
                                                ? FontWeight.bold
                                                : FontWeight.w500,
                                            color: !_showAllBranches
                                                ? darkEmerald
                                                : Colors.white.withValues(
                                                    alpha: 0.9,
                                                  ),
                                          ),
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 4),
                            // Semua Wilayah
                            Expanded(
                              child: InkWell(
                                onTap: () => _toggleBranchScope(true),
                                borderRadius: BorderRadius.circular(10),
                                child: Container(
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 8,
                                  ),
                                  decoration: BoxDecoration(
                                    color: _showAllBranches
                                        ? Colors.white
                                        : Colors.transparent,
                                    borderRadius: BorderRadius.circular(10),
                                    boxShadow: _showAllBranches
                                        ? [
                                            BoxShadow(
                                              color: Colors.black.withValues(
                                                alpha: 0.1,
                                              ),
                                              blurRadius: 4,
                                              offset: const Offset(0, 1),
                                            ),
                                          ]
                                        : null,
                                  ),
                                  child: Row(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.public_rounded,
                                        size: 14,
                                        color: _showAllBranches
                                            ? darkEmerald
                                            : Colors.white.withValues(
                                                alpha: 0.8,
                                              ),
                                      ),
                                      const SizedBox(width: 6),
                                      Flexible(
                                        child: Text(
                                          'Semua Wilayah ($allCount)',
                                          style: GoogleFonts.outfit(
                                            fontSize: 12,
                                            fontWeight: _showAllBranches
                                                ? FontWeight.bold
                                                : FontWeight.w500,
                                            color: _showAllBranches
                                                ? darkEmerald
                                                : Colors.white.withValues(
                                                    alpha: 0.9,
                                                  ),
                                          ),
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
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
                      const SizedBox(height: 10),

                      // 3. Mode Switcher: Wilayah vs Flat List + Sort Terdekat
                      Row(
                        children: [
                          // Toggle Group by District
                          InkWell(
                            onTap: () => setState(
                              () => _groupByDistrict = !_groupByDistrict,
                            ),
                            borderRadius: BorderRadius.circular(20),
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 10,
                                vertical: 5,
                              ),
                              decoration: BoxDecoration(
                                color: _groupByDistrict
                                    ? Colors.white
                                    : Colors.white.withValues(alpha: 0.2),
                                borderRadius: BorderRadius.circular(20),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(
                                    Icons.alt_route_rounded,
                                    size: 13,
                                    color: _groupByDistrict
                                        ? darkEmerald
                                        : Colors.white,
                                  ),
                                  const SizedBox(width: 4),
                                  Text(
                                    'Kelompok Wilayah (Sejalur)',
                                    style: GoogleFonts.outfit(
                                      fontSize: 11,
                                      fontWeight: FontWeight.bold,
                                      color: _groupByDistrict
                                          ? darkEmerald
                                          : Colors.white,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                          const Spacer(),
                          // Sort By Nearest Toggle
                          InkWell(
                            onTap: () => setState(
                              () => _sortByNearest = !_sortByNearest,
                            ),
                            borderRadius: BorderRadius.circular(20),
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 10,
                                vertical: 5,
                              ),
                              decoration: BoxDecoration(
                                color: _sortByNearest
                                    ? Colors.white
                                    : Colors.white.withValues(alpha: 0.2),
                                borderRadius: BorderRadius.circular(20),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(
                                    Icons.near_me_rounded,
                                    size: 12,
                                    color: _sortByNearest
                                        ? darkEmerald
                                        : Colors.white,
                                  ),
                                  const SizedBox(width: 4),
                                  Text(
                                    'Terdekat',
                                    style: GoogleFonts.outfit(
                                      fontSize: 11,
                                      fontWeight: FontWeight.bold,
                                      color: _sortByNearest
                                          ? darkEmerald
                                          : Colors.white,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),

                // Horizontal District Filter Chips
                if (sortedDistrictsList.isNotEmpty)
                  Container(
                    height: 48,
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    color: Colors.white,
                    child: ListView(
                      scrollDirection: Axis.horizontal,
                      padding: const EdgeInsets.symmetric(horizontal: 12),
                      children: [
                        // 'Semua Wilayah' Chip
                        Padding(
                          padding: const EdgeInsets.only(right: 6),
                          child: FilterChip(
                            label: Text(
                              'Semua (${searchFiltered.length})',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.bold,
                                color: _selectedDistrictFilter == 'ALL'
                                    ? Colors.white
                                    : const Color(0xFF334155),
                              ),
                            ),
                            selected: _selectedDistrictFilter == 'ALL',
                            selectedColor: primaryEmerald,
                            backgroundColor: const Color(0xFFF1F5F9),
                            checkmarkColor: Colors.white,
                            showCheckmark: false,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(20),
                              side: BorderSide(
                                color: _selectedDistrictFilter == 'ALL'
                                    ? primaryEmerald
                                    : const Color(0xFFE2E8F0),
                              ),
                            ),
                            padding: const EdgeInsets.symmetric(horizontal: 6),
                            onSelected: (_) =>
                                setState(() => _selectedDistrictFilter = 'ALL'),
                          ),
                        ),
                        // Dynamic District Chips
                        ...sortedDistrictsList.map((districtName) {
                          final count = districtCounts[districtName] ?? 0;
                          final isSelected =
                              _selectedDistrictFilter == districtName;
                          return Padding(
                            padding: const EdgeInsets.only(right: 6),
                            child: FilterChip(
                              label: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(
                                    Icons.location_city_rounded,
                                    size: 12,
                                    color: isSelected
                                        ? Colors.white
                                        : Colors.grey[700],
                                  ),
                                  const SizedBox(width: 4),
                                  Text(
                                    '$districtName ($count)',
                                    style: GoogleFonts.outfit(
                                      fontSize: 11,
                                      fontWeight: FontWeight.bold,
                                      color: isSelected
                                          ? Colors.white
                                          : const Color(0xFF334155),
                                    ),
                                  ),
                                ],
                              ),
                              selected: isSelected,
                              selectedColor: primaryEmerald,
                              backgroundColor: const Color(0xFFF1F5F9),
                              checkmarkColor: Colors.white,
                              showCheckmark: false,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(20),
                                side: BorderSide(
                                  color: isSelected
                                      ? primaryEmerald
                                      : const Color(0xFFE2E8F0),
                                ),
                              ),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 6,
                              ),
                              onSelected: (_) => setState(
                                () => _selectedDistrictFilter = districtName,
                              ),
                            ),
                          );
                        }),
                      ],
                    ),
                  ),

                // Main List of Banners (Grouped vs Flat)
                Expanded(
                  child: RefreshIndicator(
                    onRefresh: _fetchAvailableBanners,
                    color: primaryEmerald,
                    child: displayedBanners.isEmpty
                        ? ListView(
                            physics: const AlwaysScrollableScrollPhysics(),
                            children: [
                              SizedBox(
                                height:
                                    MediaQuery.of(context).size.height * 0.45,
                                child: Center(
                                  child: Padding(
                                    padding: const EdgeInsets.all(32),
                                    child: Column(
                                      mainAxisAlignment:
                                          MainAxisAlignment.center,
                                      children: [
                                        Icon(
                                          Icons.inventory_2_outlined,
                                          size: 48,
                                          color: Colors.grey[400],
                                        ),
                                        const SizedBox(height: 12),
                                        Text(
                                          !_showAllBranches &&
                                                  _allLoadedBanners.isNotEmpty
                                              ? 'Tidak ada spanduk siap pasang di cabang Anda.\nAda ${_allLoadedBanners.length} spanduk di wilayah lain.'
                                              : 'Tidak ada spanduk siap pasang ditemukan.',
                                          textAlign: TextAlign.center,
                                          style: GoogleFonts.outfit(
                                            color: const Color(0xFF475569),
                                            fontSize: 13,
                                            fontWeight: FontWeight.w600,
                                          ),
                                        ),
                                        if (!_showAllBranches &&
                                            _allLoadedBanners.isNotEmpty) ...[
                                          const SizedBox(height: 14),
                                          ElevatedButton.icon(
                                            onPressed: () =>
                                                _toggleBranchScope(true),
                                            icon: const Icon(
                                              Icons.public_rounded,
                                              size: 16,
                                            ),
                                            label: Text(
                                              'Tampilkan Semua Wilayah (${_allLoadedBanners.length})',
                                              style: GoogleFonts.outfit(
                                                fontSize: 12,
                                                fontWeight: FontWeight.bold,
                                              ),
                                            ),
                                            style: ElevatedButton.styleFrom(
                                              backgroundColor: primaryEmerald,
                                              foregroundColor: Colors.white,
                                              shape: RoundedRectangleBorder(
                                                borderRadius:
                                                    BorderRadius.circular(10),
                                              ),
                                            ),
                                          ),
                                        ],
                                      ],
                                    ),
                                  ),
                                ),
                              ),
                            ],
                          )
                        : _groupByDistrict
                        // GROUPED BY DISTRICT (WILAYAH / SEJALUR VIEW)
                        ? ListView.builder(
                            physics: const AlwaysScrollableScrollPhysics(),
                            padding: const EdgeInsets.all(16),
                            itemCount: sortedDistrictKeys.length,
                            itemBuilder: (context, index) {
                              final districtName = sortedDistrictKeys[index];
                              final districtBanners =
                                  districtGroups[districtName]!;
                              final regencyName = _getRegency(
                                districtBanners.first,
                              );

                              final isDistrictAllSelected = districtBanners
                                  .every(
                                    (b) => _selectedBanners.any(
                                      (sb) => sb['id'] == b['id'],
                                    ),
                                  );

                              final minDistance = districtBanners.fold<double>(
                                double.infinity,
                                (min, b) =>
                                    ((b['_dist_from_me'] as double?) ??
                                            double.infinity) <
                                        min
                                    ? (b['_dist_from_me'] as double)
                                    : min,
                              );

                              return Container(
                                margin: const EdgeInsets.only(bottom: 16),
                                decoration: BoxDecoration(
                                  color: const Color(0xFFF8FAFC),
                                  borderRadius: BorderRadius.circular(16),
                                  border: Border.all(
                                    color: const Color(0xFFCBD5E1),
                                    width: 1,
                                  ),
                                ),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    // District Header Card
                                    Container(
                                      padding: const EdgeInsets.all(12),
                                      decoration: const BoxDecoration(
                                        color: Color(0xFFE2E8F0),
                                        borderRadius: BorderRadius.only(
                                          topLeft: Radius.circular(15),
                                          topRight: Radius.circular(15),
                                        ),
                                      ),
                                      child: Row(
                                        children: [
                                          Container(
                                            padding: const EdgeInsets.all(6),
                                            decoration: BoxDecoration(
                                              color: primaryEmerald,
                                              borderRadius:
                                                  BorderRadius.circular(8),
                                            ),
                                            child: const Icon(
                                              Icons.alt_route_rounded,
                                              size: 16,
                                              color: Colors.white,
                                            ),
                                          ),
                                          const SizedBox(width: 8),
                                          Expanded(
                                            child: Column(
                                              crossAxisAlignment:
                                                  CrossAxisAlignment.start,
                                              children: [
                                                Text(
                                                  districtName,
                                                  style: GoogleFonts.outfit(
                                                    fontWeight: FontWeight.bold,
                                                    fontSize: 14,
                                                    color: const Color(
                                                      0xFF0F172A,
                                                    ),
                                                  ),
                                                ),
                                                Row(
                                                  children: [
                                                    Text(
                                                      '${districtBanners.length} Spanduk',
                                                      style: GoogleFonts.outfit(
                                                        fontSize: 11,
                                                        fontWeight:
                                                            FontWeight.w600,
                                                        color: const Color(
                                                          0xFF475569,
                                                        ),
                                                      ),
                                                    ),
                                                    if (regencyName
                                                        .isNotEmpty) ...[
                                                      Text(
                                                        ' • ',
                                                        style: TextStyle(
                                                          color:
                                                              Colors.grey[500],
                                                        ),
                                                      ),
                                                      Text(
                                                        regencyName,
                                                        style:
                                                            GoogleFonts.outfit(
                                                              fontSize: 11,
                                                              color:
                                                                  const Color(
                                                                    0xFF64748B,
                                                                  ),
                                                            ),
                                                      ),
                                                    ],
                                                    if (minDistance !=
                                                        double.infinity) ...[
                                                      Text(
                                                        ' • ',
                                                        style: TextStyle(
                                                          color:
                                                              Colors.grey[500],
                                                        ),
                                                      ),
                                                      Text(
                                                        'Mulai ${_formatDistance(minDistance)}',
                                                        style:
                                                            GoogleFonts.outfit(
                                                              fontSize: 11,
                                                              fontWeight:
                                                                  FontWeight
                                                                      .bold,
                                                              color: Colors
                                                                  .blue[800],
                                                            ),
                                                      ),
                                                    ],
                                                  ],
                                                ),
                                              ],
                                            ),
                                          ),
                                          // Tombol Pilih Sejalur
                                          ElevatedButton.icon(
                                            onPressed: () =>
                                                _toggleSelectDistrict(
                                                  districtBanners,
                                                ),
                                            icon: Icon(
                                              isDistrictAllSelected
                                                  ? Icons.check_box_outlined
                                                  : Icons
                                                        .check_box_outline_blank,
                                              size: 14,
                                            ),
                                            label: Text(
                                              isDistrictAllSelected
                                                  ? 'Batal Jalur'
                                                  : 'Pilih Sejalur',
                                              style: GoogleFonts.outfit(
                                                fontSize: 11,
                                                fontWeight: FontWeight.bold,
                                              ),
                                            ),
                                            style: ElevatedButton.styleFrom(
                                              backgroundColor:
                                                  isDistrictAllSelected
                                                  ? Colors.grey[700]
                                                  : primaryEmerald,
                                              foregroundColor: Colors.white,
                                              padding:
                                                  const EdgeInsets.symmetric(
                                                    horizontal: 10,
                                                    vertical: 6,
                                                  ),
                                              minimumSize: const Size(0, 32),
                                              elevation: 0,
                                              shape: RoundedRectangleBorder(
                                                borderRadius:
                                                    BorderRadius.circular(8),
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    // Store Cards in this district
                                    Padding(
                                      padding: const EdgeInsets.fromLTRB(
                                        10,
                                        10,
                                        10,
                                        0,
                                      ),
                                      child: Column(
                                        children: districtBanners
                                            .map((b) => _buildStoreCard(b))
                                            .toList(),
                                      ),
                                    ),
                                  ],
                                ),
                              );
                            },
                          )
                        // FLAT LIST VIEW
                        : ListView.builder(
                            physics: const AlwaysScrollableScrollPhysics(),
                            padding: const EdgeInsets.all(16),
                            itemCount: displayedBanners.length,
                            itemBuilder: (context, index) {
                              return _buildStoreCard(displayedBanners[index]);
                            },
                          ),
                  ),
                ),

                // Bottom Save & Route Action Card
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.05),
                        blurRadius: 10,
                        offset: const Offset(0, -5),
                      ),
                    ],
                  ),
                  child: SafeArea(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(
                              '${_selectedBanners.length} Spanduk Terpilih',
                              style: GoogleFonts.outfit(
                                fontWeight: FontWeight.bold,
                                fontSize: 13,
                              ),
                            ),
                            if (selectedDistrictNames.isNotEmpty) ...[
                              Text(
                                ' • ',
                                style: TextStyle(color: Colors.grey[400]),
                              ),
                              Flexible(
                                child: Text(
                                  selectedDistrictNames.length == 1
                                      ? selectedDistrictNames.first
                                      : '${selectedDistrictNames.length} Wilayah Sejalur',
                                  style: GoogleFonts.outfit(
                                    fontSize: 12,
                                    fontWeight: FontWeight.w600,
                                    color: primaryEmerald,
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                            ],
                          ],
                        ),
                        const SizedBox(height: 10),
                        Row(
                          children: [
                            Expanded(
                              child: ElevatedButton.icon(
                                onPressed: _selectedBanners.isEmpty
                                    ? null
                                    : _openGoogleMapsRoute,
                                icon: const Icon(Icons.map_rounded, size: 18),
                                label: const Text(
                                  'Rute Maps',
                                  style: TextStyle(fontWeight: FontWeight.bold),
                                ),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: Colors.blue[600],
                                  foregroundColor: Colors.white,
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 12,
                                  ),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: ElevatedButton.icon(
                                onPressed: _selectedBanners.isEmpty
                                    ? null
                                    : _savePlan,
                                icon: const Icon(
                                  Icons.assignment_turned_in_rounded,
                                  size: 18,
                                ),
                                label: Text(
                                  _selectedBanners.isEmpty
                                      ? 'Ambil Tugas'
                                      : 'Ambil Tugas (${_selectedBanners.length})',
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: primaryEmerald,
                                  foregroundColor: Colors.white,
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 12,
                                  ),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
    );
  }
}
