import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'dart:convert';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/features/sales/data/models/visit_plan_model.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:hris_app/features/sales/presentation/pages/optimal_route_map_page.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_plan_builder_page.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_checkin_page.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_transaction_page.dart';
import 'package:hris_app/features/sales/presentation/pages/instant_order_page.dart';
import 'package:hris_app/features/sales/presentation/pages/add_store_page.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:drift/drift.dart' hide Column;
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:intl/intl.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:shimmer/shimmer.dart';
import 'dart:math' show cos, sqrt, asin, sin, pi;

import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/sales/presentation/pages/start_journey_page.dart';
import 'package:hris_app/features/sales/presentation/pages/end_journey_page.dart';
import 'package:hris_app/features/sales/presentation/pages/sales_journey_history_page.dart';
import 'package:hris_app/features/reimbursement/presentation/pages/add_reimbursement_page.dart';

// Global RouteObserver that must be set on MaterialApp's navigatorObservers
final RouteObserver<ModalRoute<void>> visitScheduleRouteObserver =
    RouteObserver<ModalRoute<void>>();

class VisitSchedulePage extends StatefulWidget {
  const VisitSchedulePage({super.key});

  @override
  State<VisitSchedulePage> createState() => _VisitSchedulePageState();
}

class _VisitSchedulePageState extends State<VisitSchedulePage> with RouteAware {
  final _salesApi = SalesApiService(apiClient: di.sl<ApiClient>());
  final _storeApi = StoreApiService(apiClient: di.sl<ApiClient>(), prefs: di.sl<SharedPreferences>());
  bool _isLoading = true;
  List<VisitPlanModel> _plans = [];
  Map<String, dynamic>? _myPerformance;
  bool _isMtMode = false;
  int _completedCheckoutCount = 0;
  Set<String> _pendingTransactionStoreIds = {};
  Map<String, String> _noTransactionReasons = {};
  Set<String> _deletedStoreIds = {};
  // Toko yang sudah check-in (selfie tersimpan) tapi belum ada transaksi lokal hari ini
  Set<String> _checkedInStoreIds = {};
  Map<String, dynamic>? _todayJourney;
  String _journeyStatus = 'Belum Berangkat';
  String _selectedFilter = 'all'; // 'all', 'pending', 'completed'

  @override
  void initState() {
    super.initState();
    _loadDeletedStores();
    _checkMtMode();
    _fetchData();
  }

  Future<void> _loadDeletedStores() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final today = DateFormat('yyyy-MM-dd').format(DateTime.now());
      final key = 'deleted_stores_$today';
      final deletedList = prefs.getStringList(key) ?? [];
      setState(() {
        _deletedStoreIds = deletedList.toSet();
      });
    } catch (e) {
      debugPrint('Error loading deleted stores: $e');
    }
  }

  Future<void> _saveDeletedStore(String storeId) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final today = DateFormat('yyyy-MM-dd').format(DateTime.now());
      final key = 'deleted_stores_$today';
      
      _deletedStoreIds.add(storeId);
      await prefs.setStringList(key, _deletedStoreIds.toList());
    } catch (e) {
      debugPrint('Error saving deleted store: $e');
    }
  }

  void _checkMtMode() {
    SharedPreferences.getInstance().then((prefs) {
      final salesMode = prefs.getString('sales_mode');
      if (mounted && salesMode == 'MT') {
        setState(() {
          _isMtMode = true;
        });
      }
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    visitScheduleRouteObserver.subscribe(this, ModalRoute.of(context)!);
  }

  @override
  void dispose() {
    visitScheduleRouteObserver.unsubscribe(this);
    super.dispose();
  }

  /// Auto-refresh when user navigates back to this page.
  @override
  void didPopNext() {
    _fetchData(silent: true);
  }

  Future<void> _fetchData({bool silent = false}) async {
    if (!silent) setState(() => _isLoading = true);
    try {
      // 1. Prepare independent futures to run in parallel
      final plansFuture = () async {
        try {
          return await _salesApi
              .getVisitPlans()
              .timeout(const Duration(seconds: 7));
        } catch (e) {
          debugPrint('Error fetching visit plans: $e');
          return <dynamic>[];
        }
      }();

      final perfFuture = () async {
        try {
          return await _salesApi
              .getMyPerformance()
              .timeout(const Duration(seconds: 6));
        } catch (e) {
          debugPrint('Error fetching performance: $e');
          return null;
        }
      }();

      final storesFuture = () async {
        try {
          return await _storeApi
              .getStores()
              .timeout(const Duration(seconds: 7));
        } catch (e) {
          debugPrint('Error fetching stores: $e');
          return <StoreModel>[];
        }
      }();

      final locationFuture = Future<LatLng?>(() async {
        try {
          final permission = await Geolocator.checkPermission();
          if (permission == LocationPermission.always ||
              permission == LocationPermission.whileInUse) {
            Position? position;
            try {
              position = await Geolocator.getCurrentPosition(
                desiredAccuracy: LocationAccuracy.low,
                timeLimit: const Duration(milliseconds: 1500),
              );
            } catch (e) {
              debugPrint(
                  'Error getting current position, trying last known: $e');
              position = await Geolocator.getLastKnownPosition();
            }
            if (position != null) {
              return LatLng(position.latitude, position.longitude);
            }
          }
        } catch (_) {}
        return null;
      }).timeout(const Duration(seconds: 2)).catchError((_) => null);

      // 2. Wait for all futures in parallel
      final db = di.sl<AppDatabase>();
      final now = DateTime.now();
      final startOfDay = DateTime(now.year, now.month, now.day);
      final endOfDay = DateTime(now.year, now.month, now.day, 23, 59, 59);

      final localTxnsFuture = (db.select(db.localTransactions)
            ..where((t) =>
                t.createdAt.isBiggerOrEqualValue(startOfDay) &
                t.createdAt.isSmallerOrEqualValue(endOfDay)))
          .get();

      // Query check-in lokal hari ini (selfie sudah diambil & dikonfirmasi)
      final localCheckinsFuture = (db.select(db.localCheckins)
            ..where((t) => t.createdAt.isBiggerOrEqualValue(startOfDay)))
          .get();

      final journeyFuture = () async {
        try {
          return await _salesApi
              .getTodayJourney()
              .timeout(const Duration(seconds: 5));
        } catch (e) {
          debugPrint('Error fetching today journey: $e');
          return null;
        }
      }();

      final results = await Future.wait([
        plansFuture,
        perfFuture,
        storesFuture,
        locationFuture,
        localTxnsFuture,
        localCheckinsFuture,
        journeyFuture,
      ]);

      final List<dynamic> rawPlans = results[0] as List<dynamic>;
      final Map<String, dynamic>? perfData =
          results[1] as Map<String, dynamic>?;
      final List<StoreModel> allStores = results[2] as List<StoreModel>;
      final LatLng? currentPos = results[3] as LatLng?;
      final List<dynamic> todayTxns = results[4] as List<dynamic>;
      final List<dynamic> todayCheckins = results[5] as List<dynamic>;
      final Map<String, dynamic>? todayJourney = results[6] as Map<String, dynamic>?;

      final Set<String> completedStoreIds = {};
      final Set<String> pendingTransactionStoreIds = {};
      final Set<String> instantOrderStoreIds = {};
      final Map<String, String> noTransactionReasons = {};

      for (final tx in todayTxns) {
        final isInstant = tx.companyName == 'Instant Order' ||
            (tx.receiptNo?.contains('INSTANT') ?? false);
        if (isInstant) {
          instantOrderStoreIds.add(tx.storeId as String);
        } else {
          completedStoreIds.add(tx.storeId as String);
          String notesStr = tx.notes ?? '';
          if (notesStr.contains('Input Transaksi Nanti')) {
            pendingTransactionStoreIds.add(tx.storeId as String);
          } else if (tx.totalAmount == 0) {
            String reason = 'Toko Tutup / Lainnya';
            if (notesStr.isNotEmpty) {
              try {
                final decoded = jsonDecode(notesStr);
                if (decoded is Map && decoded.containsKey('notes') && decoded['notes'] != null && decoded['notes'].toString().isNotEmpty) {
                  reason = decoded['notes'].toString();
                } else {
                  reason = notesStr;
                }
              } catch (e) {
                reason = notesStr;
              }
            }
            noTransactionReasons[tx.storeId as String] = reason;
          }
        }
      }

      pendingTransactionStoreIds.removeAll(instantOrderStoreIds);
      final int completedCheckoutCount = completedStoreIds.length;

      // Identifikasi toko yang sudah check-in (selfie ada) tapi belum checkout hari ini
      final Set<String> checkedInStoreIds = {};
      for (final ci in todayCheckins) {
        final storeId = ci.storeId as String;
        // Hanya tandai jika belum ada transaksi lokal (belum checkout)
        if (!completedStoreIds.contains(storeId) &&
            !pendingTransactionStoreIds.contains(storeId)) {
          checkedInStoreIds.add(storeId);
        }
      }

      // 3. Map raw plans to models
      List<VisitPlanModel> plans =
          rawPlans.map((e) => VisitPlanModel.fromJson(e)).toList();

      // Filter out stores deleted in the current session
      plans = plans.where((p) => !_deletedStoreIds.contains(p.storeId)).toList();

      // 4. Enrich plans with store coordinates
      final storeMap = {for (var s in allStores) s.id: s};
      plans = plans.map((p) {
        final store = storeMap[p.storeId];
        
        // Override status locally if checked out today but server still says pending
        String currentStatus = p.status;
        if (currentStatus != 'COMPLETED' && currentStatus != 'EXTRA' && completedStoreIds.contains(p.storeId)) {
          currentStatus = 'COMPLETED';
        }

        if (store != null && store.latitude != 0.0 && store.longitude != 0.0) {
          return p.copyWith(
              latitude: store.latitude, longitude: store.longitude, status: currentStatus);
        }
        if (currentStatus != p.status) {
          return p.copyWith(status: currentStatus);
        }
        return p;
      }).toList();

      // 5. Sort Visit Plans Optimally
      plans = _optimizeVisitRoute(plans, currentPos);

      if (mounted) {
        Map<String, dynamic>? effectiveJourney = todayJourney;
        String rawJourneyStatus =
            (effectiveJourney?['status'] ?? '').toString().toUpperCase();

        // Client-side date check: Ensure the journey record actually belongs to today (local device date)
        if (effectiveJourney != null) {
          final todayStr = DateFormat('yyyy-MM-dd').format(DateTime.now());
          final journeyDateStr = (effectiveJourney['date'] ?? '').toString();
          if (journeyDateStr.isNotEmpty && !journeyDateStr.startsWith(todayStr)) {
            effectiveJourney = null;
            rawJourneyStatus = 'NOT_STARTED';
          }
        }

        setState(() {
          _plans = plans;
          if (perfData != null) _myPerformance = perfData;
          _completedCheckoutCount = completedCheckoutCount;
          _pendingTransactionStoreIds = pendingTransactionStoreIds;
          _noTransactionReasons = noTransactionReasons;
          _checkedInStoreIds = checkedInStoreIds;
          _todayJourney = effectiveJourney;
          _journeyStatus =
              rawJourneyStatus.isNotEmpty ? rawJourneyStatus : 'NOT_STARTED';
        });
      }
    } catch (e) {
      debugPrint('VisitSchedulePage error: $e');
    } finally {
      if (mounted && !silent) setState(() => _isLoading = false);
    }
  }

  List<VisitPlanModel> _optimizeVisitRoute(
      List<VisitPlanModel> visits, LatLng? currentPos) {
    if (visits.isEmpty) return [];

    LatLng current = currentPos ?? const LatLng(-7.250445, 112.768845);

    List<VisitPlanModel> sorted = List.from(visits);
    for (int i = 0; i < sorted.length; i++) {
      if (sorted[i].hasCoordinates) {
        sorted[i].distanceCache = _haversineDistance(current,
            LatLng(sorted[i].latitude ?? 0.0, sorted[i].longitude ?? 0.0));
      } else {
        sorted[i].distanceCache = double.infinity;
      }
    }

    sorted.sort((a, b) {
      // Toko yang sudah selesai (COMPLETED/EXTRA) dipindahkan ke bawah
      final bool aIsDone = a.status == 'COMPLETED' || a.status == 'EXTRA';
      final bool bIsDone = b.status == 'COMPLETED' || b.status == 'EXTRA';

      if (aIsDone != bIsDone) {
        return aIsDone ? 1 : -1; // yang belum selesai naik ke atas
      }

      // Dalam grup yang sama (belum selesai vs sudah selesai),
      // urutkan berdasarkan jarak terdekat
      return (a.distanceCache ?? double.infinity)
          .compareTo(b.distanceCache ?? double.infinity);
    });
    return sorted;
  }

  double _haversineDistance(LatLng p1, LatLng p2) {
    var p = 0.017453292519943295;
    var a = 0.5 -
        cos((p2.latitude - p1.latitude) * p) / 2 +
        cos(p1.latitude * p) *
            cos(p2.latitude * p) *
            (1 - cos((p2.longitude - p1.longitude) * p)) /
            2;
    return 12742 * asin(sqrt(a)); // 2 * R; R = 6371 km
  }

  int get _completedCount => _completedCheckoutCount;

  Future<void> _openGoogleMapsNav(VisitPlanModel plan) async {
    if (!plan.hasCoordinates) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(_isMtMode
              ? 'Koordinat mitra tidak tersedia.'
              : 'Koordinat toko tidak tersedia.'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }
    final uri = Uri.parse(
      'https://www.google.com/maps/dir/?api=1&destination=${plan.latitude},${plan.longitude}&travelmode=driving',
    );
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } else {
      final fallback = Uri.parse(
          'geo:${plan.latitude},${plan.longitude}?q=${plan.latitude},${plan.longitude}(${plan.storeName})');
      await launchUrl(fallback, mode: LaunchMode.externalApplication);
    }
  }

  void _openOptimalRoute() {
    final plansWithCoords = _plans.where((p) => p.hasCoordinates).toList();
    if (plansWithCoords.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            _isMtMode
                ? 'Tidak ada mitra dengan koordinat yang tersedia.'
                : 'Tidak ada toko dengan koordinat yang tersedia.',
            style: GoogleFonts.outfit(),
          ),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => OptimalRouteMapPage(plans: _plans),
      ),
    );
  }

  Widget _buildSkeleton() {
    return Shimmer.fromColors(
      baseColor: Colors.grey.shade300,
      highlightColor: Colors.grey.shade100,
      child: CustomScrollView(
        slivers: [
          SliverToBoxAdapter(
            child: Container(
              height: 250,
              decoration: const BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.only(
                  bottomLeft: Radius.circular(32),
                  bottomRight: Radius.circular(32),
                ),
              ),
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 24)),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            sliver: SliverList(
              delegate: SliverChildBuilderDelegate(
                (context, index) => Container(
                  height: 100,
                  margin: const EdgeInsets.only(bottom: 14),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(20),
                  ),
                ),
                childCount: 4,
              ),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final now = DateTime.now();
    final dateStr = DateFormat('EEEE, dd MMMM yyyy', 'id_ID').format(now);

    final filteredPlans = _plans.where((p) {
      final isCompleted = p.status == 'COMPLETED' || p.status == 'EXTRA';
      if (_selectedFilter == 'pending') return !isCompleted;
      if (_selectedFilter == 'completed') return isCompleted;
      return true;
    }).toList();

    return BlocListener<SyncBloc, SyncState>(
      listener: (context, state) {
        if (state is SyncSuccess || state is SyncMasterDataSuccess) {
          _fetchData(silent: true);
        }
      },
      child: Scaffold(
        backgroundColor: const Color(0xFFF8FAFC),
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
              Text(
                'JADWAL KUNJUNGAN HARI INI',
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  color: const Color(0xFF10B981),
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1.5,
                ),
              ),
              Text(
                dateStr,
                style: GoogleFonts.outfit(
                  fontSize: 16,
                  fontWeight: FontWeight.w900,
                  color: const Color(0xFF1E293B),
                ),
              ),
            ],
          ),
          actions: [
            IconButton(
              onPressed: _fetchData,
              icon: const Icon(Icons.refresh_rounded, color: Color(0xFF10B981)),
            ),
          ],
        ),
        body: _isLoading
            ? _buildSkeleton()
            : RefreshIndicator(
                onRefresh: _fetchData,
                color: const Color(0xFF10B981),
                child: CustomScrollView(
                  slivers: [
                    SliverToBoxAdapter(child: _buildHeader()),
                    SliverToBoxAdapter(
                      child: Padding(
                        padding: const EdgeInsets.fromLTRB(24, 8, 24, 12),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              _isMtMode ? 'DAFTAR MITRA' : 'DAFTAR TOKO',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                fontWeight: FontWeight.w900,
                                color: Colors.blueGrey,
                                letterSpacing: 1,
                              ),
                            ),
                            Text(
                              _isMtMode
                                  ? '${_plans.length} Mitra'
                                  : '${_plans.length} Toko',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                fontWeight: FontWeight.w700,
                                color: const Color(0xFF10B981),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    SliverToBoxAdapter(child: _buildFilterChips()),
                    filteredPlans.isEmpty
                        ? SliverFillRemaining(child: _buildEmptyState())
                        : SliverPadding(
                            padding: const EdgeInsets.symmetric(horizontal: 24),
                            sliver: SliverList(
                              delegate: SliverChildBuilderDelegate(
                                (ctx, i) => _buildVisitCard(filteredPlans[i]),
                                childCount: filteredPlans.length,
                              ),
                            ),
                          ),
                    const SliverToBoxAdapter(child: SizedBox(height: 100)),
                  ],
                ),
              ),
        floatingActionButton: Column(
          mainAxisAlignment: MainAxisAlignment.end,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            FloatingActionButton.extended(
              heroTag: 'new_store',
              onPressed: () async {
                final result = await Navigator.push(
                  context,
                  MaterialPageRoute(
                      builder: (_) => AddStorePage(isMT: _isMtMode)),
                );
                if (result != null && result is StoreModel) {
                  setState(() => _isLoading = true);
                  try {
                    await _salesApi
                        .saveManualVisitPlan(storeIds: [result.id], planDate: DateTime.now());
                    if (mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                            content: Text(_isMtMode
                                ? 'Kunjungan mitra baru berhasil ditambahkan!'
                                : 'Kunjungan toko baru berhasil ditambahkan!'),
                            backgroundColor: Colors.green),
                      );
                      _fetchData();
                    }
                  } catch (e) {
                    if (mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                            content: Text('Gagal menyimpan rencana: $e'),
                            backgroundColor: Colors.red),
                      );
                    }
                    setState(() => _isLoading = false);
                  }
                }
              },
              backgroundColor: Colors.orange,
              elevation: 4,
              icon: const Icon(Icons.add_business_rounded,
                  color: Colors.white, size: 20),
              label: Text(
                _isMtMode ? 'Mitra Baru' : 'Toko Baru',
                style: GoogleFonts.outfit(
                    color: Colors.white,
                    fontWeight: FontWeight.w900,
                    fontSize: 13),
              ),
            ),
            const SizedBox(height: 12),
            FloatingActionButton.extended(
              heroTag: 'visit_plan',
              onPressed: () async {
                final result = await Navigator.push(
                  context,
                  MaterialPageRoute(
                      builder: (_) => const VisitPlanBuilderPage()),
                );
                if (result == true) {
                  _fetchData();
                }
              },
              backgroundColor: const Color(0xFF10B981),
              elevation: 8,
              icon: const Icon(Icons.add_location_alt_rounded,
                  color: Colors.white),
              label: Text(
                'Buat Rencana',
                style: GoogleFonts.outfit(
                    color: Colors.white, fontWeight: FontWeight.w900),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader() {
    final targetVisits =
        ((_myPerformance?['daily_missions']?['target_visits'] as num?)
                ?.toInt() ??
            (_plans.isEmpty ? 1 : _plans.length));
    final double progress =
        targetVisits > 0 ? _completedCount / targetVisits : 0.0;

    final bool isStarted = _journeyStatus == 'STARTED';
    final bool isEnded = _journeyStatus == 'ENDED';
    final bool isNotStarted = !isStarted && !isEnded;

    final startKm = _todayJourney?['start_km'] ?? '-';
    final endKm = _todayJourney?['end_km'] ?? '-';
    final totalKm = _todayJourney?['total_km'] ?? '-';

    return Container(
      padding: const EdgeInsets.all(24),
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.only(
          bottomLeft: Radius.circular(32),
          bottomRight: Radius.circular(32),
        ),
        boxShadow: [
          BoxShadow(
              color: Color(0x08000000), blurRadius: 20, offset: Offset(0, 10))
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _isMtMode
                        ? 'Progress Kunjungan Mitra'
                        : 'Progress Kunjungan',
                    style: GoogleFonts.outfit(
                        fontSize: 13,
                        color: Colors.blueGrey,
                        fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 4),
                  RichText(
                    text: TextSpan(
                      children: [
                        TextSpan(
                          text: '$_completedCount',
                          style: GoogleFonts.outfit(
                              fontSize: 32,
                              fontWeight: FontWeight.w900,
                              color: const Color(0xFF1E293B)),
                        ),
                        TextSpan(
                          text: ' / $targetVisits',
                          style: GoogleFonts.outfit(
                              fontSize: 18,
                              fontWeight: FontWeight.w700,
                              color: Colors.blueGrey),
                        ),
                        TextSpan(
                          text: _isMtMode ? ' Mitra' : ' Toko',
                          style: GoogleFonts.outfit(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: Colors.blueGrey),
                        ),
                      ],
                    ),
                  ),
                  if (_plans.length < targetVisits)
                    Container(
                      margin: const EdgeInsets.only(top: 6),
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 3),
                      decoration: BoxDecoration(
                          color: Colors.orange.shade50,
                          borderRadius: BorderRadius.circular(6)),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.warning_amber_rounded,
                              size: 10, color: Colors.orange.shade700),
                          const SizedBox(width: 4),
                          Text(
                              'Rencana kurang ${targetVisits - _plans.length} ${_isMtMode ? 'mitra' : 'toko'}',
                              style: GoogleFonts.outfit(
                                  color: Colors.orange.shade800,
                                  fontSize: 9,
                                  fontWeight: FontWeight.w800)),
                        ],
                      ),
                    ),
                ],
              ),
              // Route button in header as well
              GestureDetector(
                onTap: _openOptimalRoute,
                child: Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFF059669), Color(0xFF10B981)],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    borderRadius: BorderRadius.circular(16),
                    boxShadow: [
                      BoxShadow(
                          color: const Color(0xFF10B981).withOpacity(0.3),
                          blurRadius: 12,
                          offset: const Offset(0, 4)),
                    ],
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.route_rounded,
                          color: Colors.white, size: 18),
                      const SizedBox(width: 6),
                      Text(
                        'Rute\nOptimal',
                        textAlign: TextAlign.center,
                        style: GoogleFonts.outfit(
                            color: Colors.white,
                            fontWeight: FontWeight.w900,
                            fontSize: 12),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: LinearProgressIndicator(
              value: progress,
              minHeight: 8,
              backgroundColor: const Color(0xFF10B981).withOpacity(0.1),
              valueColor:
                  const AlwaysStoppedAnimation<Color>(Color(0xFF10B981)),
            ),
          ),
          const SizedBox(height: 14),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFFF1F5F9),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: const Color(0xFFE2E8F0)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Icon(Icons.flag_rounded,
                        color: Color(0xFFF59E0B), size: 16),
                    const SizedBox(width: 6),
                    Text('Misi Target Harian',
                        style: GoogleFonts.outfit(
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                            color: const Color(0xFF1E293B))),
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    _buildTargetBadge(
                        Icons.storefront_rounded,
                        '$_completedCheckoutCount',
                        '${_myPerformance?['daily_missions']?['target_visits'] ?? 0}',
                        'Kunjungan'),
                    const SizedBox(width: 8),
                    _buildTargetBadge(
                        Icons.add_business_rounded,
                        '${_myPerformance?['daily_missions']?['achieved_new_stores'] ?? 0}',
                        '${_myPerformance?['daily_missions']?['target_new_stores'] ?? 0}',
                        _isMtMode ? 'Mitra Baru' : 'Toko Baru'),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 10),
          // Compact Journey Strip below Target Harian
          if (isNotStarted)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              decoration: BoxDecoration(
                color: const Color(0xFFFFFBEB),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: const Color(0xFFFDE68A)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.warning_amber_rounded, color: Color(0xFFD97706), size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'KM Berangkat belum dicatat',
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: const Color(0xFF92400E),
                      ),
                    ),
                  ),
                  InkWell(
                    onTap: () {
                      Navigator.of(context).push(
                          MaterialPageRoute(builder: (_) => const SalesJourneyHistoryPage()));
                    },
                    borderRadius: BorderRadius.circular(6),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(6),
                        border: Border.all(color: const Color(0xFFFDE68A)),
                      ),
                      child: Text(
                        'Riwayat',
                        style: GoogleFonts.outfit(
                          fontSize: 11,
                          fontWeight: FontWeight.w800,
                          color: const Color(0xFF92400E),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 6),
                  InkWell(
                    onTap: () {
                      Navigator.of(context)
                          .push(MaterialPageRoute(builder: (_) => const StartJourneyPage()))
                          .then((result) {
                            if (result == true) {
                              _fetchData(silent: false);
                            } else {
                              _fetchData(silent: true);
                            }
                          });
                    },
                    borderRadius: BorderRadius.circular(6),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                      decoration: BoxDecoration(
                        color: const Color(0xFFD97706),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Text(
                        'Catat KM',
                        style: GoogleFonts.outfit(
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                          color: Colors.white,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            )
          else if (isStarted)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              decoration: BoxDecoration(
                color: const Color(0xFFEFF6FF),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: const Color(0xFFBFDBFE)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.directions_car_filled_rounded, color: Color(0xFF2563EB), size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Sedang di Lapangan • KM $startKm',
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        fontWeight: FontWeight.w800,
                        color: const Color(0xFF1E40AF),
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  InkWell(
                    onTap: () {
                      Navigator.of(context).push(
                          MaterialPageRoute(builder: (_) => const SalesJourneyHistoryPage()));
                    },
                    borderRadius: BorderRadius.circular(6),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 5),
                      child: const Icon(Icons.history_rounded, size: 18, color: Color(0xFF2563EB)),
                    ),
                  ),
                  const SizedBox(width: 2),
                  InkWell(
                    onTap: () {
                      final title = 'BBM Sales - ${DateFormat('yyyy-MM-dd').format(DateTime.now())}';
                      Navigator.of(context).push(MaterialPageRoute(
                          builder: (_) => AddReimbursementPage(initialTitle: title)));
                    },
                    borderRadius: BorderRadius.circular(6),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(6),
                        border: Border.all(color: const Color(0xFF93C5FD)),
                      ),
                      child: Text(
                        '+ BBM',
                        style: GoogleFonts.outfit(
                          fontSize: 10.5,
                          fontWeight: FontWeight.bold,
                          color: const Color(0xFF2563EB),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 6),
                  InkWell(
                    onTap: () {
                      Navigator.of(context)
                          .push(MaterialPageRoute(builder: (_) => const EndJourneyPage()))
                          .then((result) {
                            if (result == true) {
                              _fetchData(silent: false);
                            } else {
                              _fetchData(silent: true);
                            }
                          });
                    },
                    borderRadius: BorderRadius.circular(6),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
                      decoration: BoxDecoration(
                        color: const Color(0xFF10B981),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Text(
                        'KM Pulang',
                        style: GoogleFonts.outfit(
                          fontSize: 10.5,
                          fontWeight: FontWeight.w900,
                          color: Colors.white,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            )
          else
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
              decoration: BoxDecoration(
                color: const Color(0xFFECFDF5),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: const Color(0xFFA7F3D0)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.check_circle_rounded, color: Color(0xFF059669), size: 16),
                  const SizedBox(width: 6),
                  Expanded(
                    child: Text(
                      'Selesai ($startKm → $endKm • Total $totalKm KM)',
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: const Color(0xFF065F46),
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  InkWell(
                    onTap: () {
                      Navigator.of(context).push(
                          MaterialPageRoute(builder: (_) => const SalesJourneyHistoryPage()));
                    },
                    borderRadius: BorderRadius.circular(6),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(6),
                        border: Border.all(color: const Color(0xFFA7F3D0)),
                      ),
                      child: Text(
                        'Riwayat',
                        style: GoogleFonts.outfit(
                          fontSize: 10.5,
                          fontWeight: FontWeight.w800,
                          color: const Color(0xFF065F46),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildFilterChips() {
    final int totalCount = _plans.length;
    final int completed = _completedCount;
    final int pending = (totalCount - completed).clamp(0, totalCount);

    return Padding(
      padding: const EdgeInsets.fromLTRB(24, 0, 24, 14),
      child: Row(
        children: [
          _buildChipItem('all', 'Semua ($totalCount)'),
          const SizedBox(width: 8),
          _buildChipItem('pending', 'Belum ($pending)'),
          const SizedBox(width: 8),
          _buildChipItem('completed', 'Selesai ($completed)'),
        ],
      ),
    );
  }

  Widget _buildChipItem(String key, String label) {
    final bool isSelected = _selectedFilter == key;
    return Expanded(
      child: InkWell(
        onTap: () => setState(() => _selectedFilter = key),
        borderRadius: BorderRadius.circular(10),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 150),
          padding: const EdgeInsets.symmetric(vertical: 8),
          decoration: BoxDecoration(
            color: isSelected ? const Color(0xFF1E293B) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: isSelected ? const Color(0xFF1E293B) : const Color(0xFFE2E8F0),
            ),
            boxShadow: isSelected
                ? [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.06),
                      blurRadius: 6,
                      offset: const Offset(0, 2),
                    ),
                  ]
                : [],
          ),
          child: Text(
            label,
            textAlign: TextAlign.center,
            style: GoogleFonts.outfit(
              fontSize: 11,
              fontWeight: isSelected ? FontWeight.w900 : FontWeight.w600,
              color: isSelected ? Colors.white : const Color(0xFF64748B),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildVisitCard(VisitPlanModel plan) {
    final isCompleted = plan.status == 'COMPLETED' || plan.status == 'EXTRA';
    final isExtra = plan.isExtra || plan.status == 'EXTRA';
    final hasCoords = plan.hasCoordinates;
    final needsTransaction = _pendingTransactionStoreIds.contains(plan.storeId);
    final String? noTxReason = _noTransactionReasons[plan.storeId];
    // Sudah selfie check-in hari ini tapi belum ada transaksi/checkout
    final bool isCheckedIn = !isCompleted &&
        !needsTransaction &&
        _checkedInStoreIds.contains(plan.storeId);

    return Container(
      margin: const EdgeInsets.only(bottom: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: plan.isPriority
              ? Colors.redAccent.withOpacity(0.5)
              : (isCheckedIn
                  ? Colors.blue.withOpacity(0.4)
                  : (needsTransaction
                      ? Colors.orange.withOpacity(0.5)
                      : (isCompleted
                          ? Colors.green.withOpacity(0.15)
                          : Colors.transparent))),
          width: plan.isPriority || needsTransaction || isCheckedIn ? 2.0 : 1.5,
        ),
        boxShadow: [
          if (plan.isPriority)
            BoxShadow(
                color: Colors.redAccent.withOpacity(0.15),
                blurRadius: 16,
                offset: const Offset(0, 8))
          else if (isCheckedIn)
            BoxShadow(
                color: Colors.blue.withOpacity(0.12),
                blurRadius: 16,
                offset: const Offset(0, 8))
          else
            BoxShadow(
                color: Colors.black.withOpacity(0.04),
                blurRadius: 12,
                offset: const Offset(0, 6)),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: needsTransaction
                ? () => _navigateToInstantOrder(plan)
                : (isCompleted
                    ? () => _showRevisitDialog(plan)
                    : () => _navigateToCheckin(plan)),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
              child: Row(
                children: [
                  _buildStatusIcon(plan.status, isCheckedIn: isCheckedIn),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                plan.storeName,
                                style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w900,
                                  fontSize: 15,
                                  color: isCompleted
                                      ? const Color(0xFF64748B)
                                      : const Color(0xFF1E293B),
                                ),
                              ),
                            ),
                            if (isCheckedIn)
                              Container(
                                margin: const EdgeInsets.only(left: 6),
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 7, vertical: 2),
                                decoration: BoxDecoration(
                                  color: Colors.blue.withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(6),
                                  border: Border.all(
                                      color: Colors.blue.withOpacity(0.3)),
                                ),
                                child: Text(
                                  '▶ LANJUTKAN',
                                  style: GoogleFonts.outfit(
                                      color: Colors.blue,
                                      fontSize: 9,
                                      fontWeight: FontWeight.w900),
                                ),
                              )
                            else if (plan.isPriority)
                              Container(
                                margin: const EdgeInsets.only(left: 6),
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 7, vertical: 2),
                                decoration: BoxDecoration(
                                  color: Colors.red.withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(6),
                                  border: Border.all(
                                      color: Colors.red.withOpacity(0.3)),
                                ),
                                child: Text(
                                  'PRIORITAS',
                                  style: GoogleFonts.outfit(
                                      color: Colors.red,
                                      fontSize: 9,
                                      fontWeight: FontWeight.w900),
                                ),
                              )
                            else if (isExtra)
                              Container(
                                margin: const EdgeInsets.only(left: 6),
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 7, vertical: 2),
                                decoration: BoxDecoration(
                                  color: Colors.orange.withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(6),
                                ),
                                child: Text(
                                  'EKSTRA',
                                  style: GoogleFonts.outfit(
                                      color: Colors.orange,
                                      fontSize: 9,
                                      fontWeight: FontWeight.w900),
                                ),
                              ),
                          ],
                        ),
                        const SizedBox(height: 3),
                        Text(
                          plan.address,
                          style: GoogleFonts.outfit(
                              fontSize: 12, color: Colors.blueGrey),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        if (plan.lastTransactionAmount != null ||
                            plan.lastVisitDate != null) ...[
                          const SizedBox(height: 4),
                          Text(
                            'Trx Terakhir: ${plan.lastTransactionAmount != null ? NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0).format(plan.lastTransactionAmount) : "Belum Ada"} '
                            '${plan.lastVisitDate != null ? "(${DateTime.now().difference(plan.lastVisitDate!).inDays} Hari Lalu)" : "(Belum Pernah)"}',
                            style: GoogleFonts.outfit(
                                fontSize: 10,
                                color: Colors.blueGrey,
                                fontStyle: FontStyle.italic),
                          ),
                        ],
                        if (plan.distanceCache != null &&
                            plan.distanceCache != double.infinity) ...[
                          const SizedBox(height: 4),
                          Row(
                            children: [
                              const Icon(Icons.location_on,
                                  size: 12, color: Colors.redAccent),
                              const SizedBox(width: 4),
                              Text(
                                plan.distanceCache! < 1.0
                                    ? '${(plan.distanceCache! * 1000).toStringAsFixed(0)} m'
                                    : '${plan.distanceCache!.toStringAsFixed(1)} km',
                                style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    color: Colors.blueGrey,
                                    fontWeight: FontWeight.w600),
                              ),
                            ],
                          ),
                        ],
                        if (needsTransaction) ...[
                          const SizedBox(height: 4),
                          Row(
                            children: [
                              const Icon(Icons.warning_amber_rounded,
                                  color: Colors.orange, size: 12),
                              const SizedBox(width: 4),
                              Text(
                                'Input Transaksi Nanti',
                                style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  color: Colors.orange,
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ],
                          ),
                        ] else if (isCompleted) ...[
                          const SizedBox(height: 4),
                          Row(
                            children: [
                              const Icon(Icons.check_circle_rounded,
                                  color: Colors.green, size: 12),
                              const SizedBox(width: 4),
                              Text(
                                'Kunjungan selesai',
                                style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  color: Colors.green,
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ],
                          ),
                          if (noTxReason != null && noTxReason.isNotEmpty) ...[
                            const SizedBox(height: 6),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                              decoration: BoxDecoration(
                                color: Colors.red.shade50,
                                borderRadius: BorderRadius.circular(4),
                                border: Border.all(color: Colors.red.shade200),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(Icons.info_outline_rounded, size: 10, color: Colors.red.shade700),
                                  const SizedBox(width: 4),
                                  Flexible(
                                    child: Text(
                                      'Tanpa Transaksi: $noTxReason',
                                      style: GoogleFonts.outfit(
                                        fontSize: 9,
                                        color: Colors.red.shade700,
                                        fontWeight: FontWeight.w700,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ],
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  // Google Maps navigation button
                  if (hasCoords)
                    Tooltip(
                      message: 'Navigasi ke toko',
                      child: InkWell(
                        borderRadius: BorderRadius.circular(12),
                        onTap: () => _openGoogleMapsNav(plan),
                        child: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: const Color(0xFF4285F4).withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: const Icon(
                            Icons.navigation_rounded,
                            color: Color(0xFF4285F4),
                            size: 20,
                          ),
                        ),
                      ),
                    ),
                  if (plan.status == 'PLANNED')
                    Tooltip(
                      message: 'Hapus Rencana',
                      child: InkWell(
                        borderRadius: BorderRadius.circular(12),
                        onTap: () => _deletePlan(plan),
                        child: Container(
                          padding: const EdgeInsets.all(8),
                          margin: const EdgeInsets.only(left: 6),
                          decoration: BoxDecoration(
                            color: Colors.red.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: const Icon(
                            Icons.delete_outline_rounded,
                            color: Colors.red,
                            size: 20,
                          ),
                        ),
                      ),
                    ),
                  if (!isCompleted) ...[
                    const SizedBox(width: 8),
                    const Icon(Icons.arrow_forward_ios_rounded,
                        color: Color(0xFF10B981), size: 15),
                  ],
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }


  Widget _buildStatusIcon(String status, {bool isCheckedIn = false}) {
    // Status lokal: sudah check-in tapi belum checkout — tampilkan ikon biru
    if (isCheckedIn && status != 'COMPLETED' && status != 'EXTRA') {
      return Container(
        width: 44,
        height: 44,
        decoration: BoxDecoration(
            color: Colors.blue.withOpacity(0.1), shape: BoxShape.circle),
        child: const Icon(Icons.play_circle_outline_rounded,
            color: Colors.blue, size: 22),
      );
    }

    Color color;
    IconData icon;

    switch (status) {
      case 'COMPLETED':
      case 'EXTRA':
        color = Colors.green;
        icon = Icons.check_circle_rounded;
        break;
      case 'CHECKED_IN':
        color = Colors.orange;
        icon = Icons.login_rounded;
        break;
      case 'SKIPPED':
        color = Colors.red;
        icon = Icons.cancel_rounded;
        break;
      default:
        color = const Color(0xFF10B981);
        icon = Icons.schedule_rounded;
    }

    return Container(
      width: 44,
      height: 44,
      decoration:
          BoxDecoration(color: color.withOpacity(0.1), shape: BoxShape.circle),
      child: Icon(icon, color: color, size: 22),
    );
  }

  void _navigateToCheckin(VisitPlanModel plan) async {
    try {
      final journey = await _salesApi.getTodayJourney();
      final status = (journey?['status'] ?? '').toString().toUpperCase();
      final todayStr = DateFormat('yyyy-MM-dd').format(DateTime.now());
      final journeyDateStr = (journey?['date'] ?? '').toString();
      final isToday = journeyDateStr.isEmpty || journeyDateStr.startsWith(todayStr);

      if (journey == null || !isToday || (status != 'STARTED' && status != 'ENDED')) {
        if (!mounted) return;
        await showDialog<void>(
          context: context,
          builder: (dialogContext) => AlertDialog(
            icon: const Icon(Icons.speed_rounded, color: Colors.orange, size: 36),
            title: const Text('Catat KM terlebih dahulu'),
            content: const Text(
              'Sebelum melakukan kunjungan, Anda harus mengisi catat KM Berangkat pada menu Mulai Bekerja.',
            ),
            actions: [
              FilledButton(
                onPressed: () => Navigator.pop(dialogContext),
                child: const Text('Mengerti'),
              ),
            ],
          ),
        );
        return;
      }

      final store = await _resolveStoreForPlan(plan);

      // Cek apakah sudah ada check-in hari ini untuk toko ini (selfie sudah dikonfirmasi)
      final db = di.sl<AppDatabase>();
      final now = DateTime.now();
      final startOfDay = DateTime(now.year, now.month, now.day);

      final existingCheckin = await (db.select(db.localCheckins)
            ..where((t) =>
                t.storeId.equals(plan.storeId) &
                t.createdAt.isBiggerOrEqualValue(startOfDay))
            ..orderBy([(t) => OrderingTerm(
                expression: t.createdAt, mode: OrderingMode.desc)])
            ..limit(1))
          .getSingleOrNull();

      if (!mounted) return;

      if (existingCheckin != null) {
        // ✅ Sudah check-in sebelumnya — langsung resume ke halaman transaksi
        // tanpa perlu selfie ulang
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => VisitTransactionPage(
              store: store,
              selfiePath: existingCheckin.selfiePath,
              selfieBytes: null,
            ),
          ),
        );
      } else {
        // Belum pernah check-in hari ini — buka kamera selfie seperti biasa
        Navigator.push(
          context,
          MaterialPageRoute(builder: (_) => VisitCheckinPage(store: store)),
        );
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e')),
      );
    }
  }

  Future<StoreModel> _resolveStoreForPlan(VisitPlanModel plan) async {
    try {
      return await _storeApi.getStoreByID(plan.storeId);
    } catch (_) {
      return StoreModel(
        id: plan.storeId,
        name: plan.storeName,
        ownerName: '',
        address: plan.address,
        latitude: plan.latitude ?? 0.0,
        longitude: plan.longitude ?? 0.0,
        assignedEmployeeId: plan.assignedEmployeeId,
        assignedEmployeeName: plan.assignedEmployeeName,
      );
    }
  }

  void _navigateToInstantOrder(VisitPlanModel plan) async {
    try {
      final store = await _resolveStoreForPlan(plan);
      if (mounted) {
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => InstantOrderPage(
              initialStore: store,
              initialVisitId: plan.visitId,
            ),
          ),
        );
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e')),
      );
    }
  }

  void _showRevisitDialog(VisitPlanModel plan) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('Kunjungan Ulang',
            style: GoogleFonts.outfit(fontWeight: FontWeight.w800)),
        content: Text(
          'Toko ini sudah dikunjungi hari ini. Apakah Anda ingin melakukan kunjungan ulang (misalnya jika sebelumnya tutup dan sekarang buka, atau ada pesanan susulan)?',
          style: GoogleFonts.outfit(),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text('Batal', style: GoogleFonts.outfit()),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(ctx);
              _navigateToCheckin(plan);
            },
            style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF10B981)),
            child: Text('Ya, Kunjungi Lagi',
                style: GoogleFonts.outfit(
                    color: Colors.white, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  Future<void> _deletePlan(VisitPlanModel plan) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('Hapus Rencana?',
            style: GoogleFonts.outfit(fontWeight: FontWeight.w800)),
        content: Text(
            'Anda yakin ingin menghapus ${plan.storeName} dari daftar kunjungan?',
            style: GoogleFonts.outfit()),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: Text('Batal', style: GoogleFonts.outfit())),
          TextButton(
              onPressed: () => Navigator.pop(ctx, true),
              child: Text('Hapus',
                  style: GoogleFonts.outfit(
                      color: Colors.red, fontWeight: FontWeight.bold))),
        ],
      ),
    );

    if (confirm != true) return;

    setState(() => _isLoading = true);
    try {
      // PENTING: Hanya hapus rencana dari jadwal (deleteManualVisitPlan).
      // Jangan memanggil `deleteVisitPlan(visitId)` karena akan menghapus 
      // data transaksi check-in aktual.
      final now = DateTime.now();
      final dateStr =
          "${now.year}-${now.month.toString().padLeft(2, '0')}-${now.day.toString().padLeft(2, '0')}";
      await _salesApi.deleteManualVisitPlan(plan.storeId, date: dateStr);
      
      await _saveDeletedStore(plan.storeId);
      setState(() {
        _plans.removeWhere((p) => p.storeId == plan.storeId);
      });
      await _fetchData(silent: true);
      
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Rencana kunjungan dihapus')));
      }
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Gagal: $e')));
      }
    }
  }

  Widget _buildTargetBadge(
      IconData icon, String achieved, String target, String label) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          children: [
            Icon(icon, size: 14, color: const Color(0xFF10B981)),
            const SizedBox(width: 4),
            Text('$achieved/$target',
                style: GoogleFonts.outfit(
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                    color: const Color(0xFF1E293B))),
            const SizedBox(width: 4),
            Expanded(
                child: Text(label,
                    style: GoogleFonts.outfit(
                        fontSize: 10, color: Colors.blueGrey),
                    overflow: TextOverflow.ellipsis)),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.event_available_rounded,
              size: 80, color: Colors.grey.shade300),
          const SizedBox(height: 24),
          Text(
            'Tidak Ada Jadwal',
            style: GoogleFonts.outfit(
                fontSize: 20,
                fontWeight: FontWeight.w800,
                color: const Color(0xFF334155)),
          ),
          const SizedBox(height: 8),
          Text(
            'Anda tidak memiliki rencana\nkunjungan hari ini.',
            textAlign: TextAlign.center,
            style: GoogleFonts.outfit(fontSize: 14, color: Colors.blueGrey),
          ),
          const SizedBox(height: 32),
          ElevatedButton.icon(
            onPressed: () async {
              final result = await Navigator.push(
                context,
                MaterialPageRoute(builder: (_) => const VisitPlanBuilderPage()),
              );
              if (result == true) {
                _fetchData();
              }
            },
            icon:
                const Icon(Icons.add_location_alt_rounded, color: Colors.white),
            label: Text(
              'Buat Rencana Kunjungan',
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800, color: Colors.white),
            ),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF10B981),
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16)),
            ),
          ),
        ],
      ),
    );
  }
}
