import 'dart:io';
import 'dart:async';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'dart:convert';
import 'package:drift/drift.dart' hide Column;
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:intl/intl.dart';
import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import './sales_history_page.dart';
import './sales_journey_history_page.dart';
import 'sales_stock_request_page.dart';
import 'sales_stock_history_page.dart';
import './manual_stock_adjustment_page.dart';
import './store_list_page.dart';
import './order_banner_page.dart';
import './banner_install_page.dart';
import './route_planning_page.dart';
import './product_catalog_page.dart';
import './visit_schedule_page.dart';
import './delivery_tracking_page.dart';
import './mt_po_list_page.dart';
import './merchandising_page.dart';
import './merchandising_history_page.dart';
import './visit_plan_management_page.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/core/services/data_integrity_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/network/network_info.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:hris_app/features/sync/data/repositories/sync_repository.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_state.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_event.dart';
import './tempo_payment_page.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:hris_app/core/database/database.dart';
import './digital_receipt_page.dart';
import './instant_order_page.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'dart:math' show cos, sqrt, asin, sin, pi;
import 'package:url_launcher/url_launcher.dart';
import 'package:shimmer/shimmer.dart';
import 'package:hris_app/features/sync/presentation/widgets/sync_status_banner.dart';
import 'package:hris_app/features/reimbursement/presentation/pages/reimbursement_list_page.dart';
import 'package:hris_app/features/reimbursement/presentation/pages/add_reimbursement_page.dart';
import 'package:shared_preferences/shared_preferences.dart';

class SalesDashboardPage extends StatefulWidget {
  /// [isMT] = true jika dashboard ini digunakan oleh Sales MT (Modern Trade).
  /// [factoryCompanyId] = UUID company pabrik/pusat yang mengelola omzet MT.
  final bool isMT;
  final String? factoryCompanyId;
  final bool forceTO;

  const SalesDashboardPage({
    super.key,
    this.isMT = false,
    this.factoryCompanyId,
    this.forceTO = false,
  });

  @override
  State<SalesDashboardPage> createState() => _SalesDashboardPageState();
}

class _SalesDashboardPageState extends State<SalesDashboardPage>
    with RouteAware, WidgetsBindingObserver {
  final _apiService = SalesApiService(apiClient: di.sl<ApiClient>());
  bool _isLoading = true;
  bool _isOfflineMode = false;

  Map<String, dynamic>? _kpiData;
  List<dynamic> _pendingTransactions = [];
  List<dynamic> _outstandingTempoTransactions = [];
  List<dynamic> _todayVisitPlans = [];
  List<Map<String, dynamic>> _carriedStock = [];
  double _todayOrderTotal = 0.0;
  double _localTodayOrderTotal = 0.0;
  late String _displayedDate;
  Timer? _dayRolloverTimer;
  int _fetchGeneration = 0;
  int _completedCheckoutCount = 0;
  final PageController _pageController = PageController();
  int _currentPage = 0;

  @override
  void initState() {
    super.initState();
    _displayedDate = DateFormat('yyyy-MM-dd').format(DateTime.now());
    WidgetsBinding.instance.addObserver(this);
    _scheduleDayRollover();
    context.read<ProfileBloc>().add(LoadProfileRequested());
    // Auto-sync on load to push offline data and get fresh state
    context.read<SyncBloc>().add(const SyncDataRequested(force: true));
    _fetchData();

    final authState = context.read<AuthBloc>().state;
    bool isSupervisorUser = false;
    bool isSalesTOUser = false;
    if (authState is Authenticated && authState.userProfile != null) {
      final profile = authState.userProfile!;
      final role = profile['role']?.toString().toLowerCase() ?? '';
      final jobTitle =
          (profile['job_position']?['title'] ?? profile['job_title'] ?? '')
              .toString()
              .toLowerCase();
      isSupervisorUser =
          role == 'supervisor_sales' ||
          role == 'supervisor_sales_mt' ||
          role == 'supervisor' ||
          role == 'spv' ||
          jobTitle.contains('supervisor') ||
          jobTitle.contains('spv');
      final isMotoris = jobTitle.contains('motoris');
      isSalesTOUser =
          !isMotoris &&
          (jobTitle.contains('task order') ||
              jobTitle.contains('take order') ||
              jobTitle.contains('sales to') ||
              jobTitle.startsWith('to ') ||
              jobTitle.endsWith(' to') ||
              jobTitle == 'to');
    }

    if (widget.forceTO || isSupervisorUser || isSalesTOUser) {
      SharedPreferences.getInstance().then((prefs) {
        prefs.setString('sales_mode', 'TO');
        prefs.remove('factory_company_id');
      });
    } else if (widget.isMT) {
      SharedPreferences.getInstance().then((prefs) {
        prefs.setString('sales_mode', 'MT');
        if (widget.factoryCompanyId != null) {
          prefs.setString('factory_company_id', widget.factoryCompanyId!);
        }
      });
    } else {
      SharedPreferences.getInstance().then((prefs) {
        prefs.remove('sales_mode');
        prefs.remove('factory_company_id');
      });
    }
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    visitScheduleRouteObserver.subscribe(this, ModalRoute.of(context)!);
  }

  @override
  void didPopNext() {
    final previousDate = _displayedDate;
    _refreshForNewDay();
    if (previousDate == _displayedDate) {
      _fetchData(showSkeleton: false);
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _refreshForNewDay();
    }
  }

  void _scheduleDayRollover() {
    _dayRolloverTimer?.cancel();
    final now = DateTime.now();
    final tomorrow = DateTime(now.year, now.month, now.day + 1);
    _dayRolloverTimer = Timer(tomorrow.difference(now), _refreshForNewDay);
  }

  void _refreshForNewDay() {
    if (!mounted) return;
    final today = DateFormat('yyyy-MM-dd').format(DateTime.now());
    if (today != _displayedDate) {
      _displayedDate = today;
      ++_fetchGeneration;
      setState(() {
        _kpiData = null;
        _todayVisitPlans = [];
        _todayOrderTotal = 0;
        _localTodayOrderTotal = 0;
        _completedCheckoutCount = 0;
      });
      _fetchData(showSkeleton: false);
    }
    _scheduleDayRollover();
  }

  Future<void> _fetchData({bool showSkeleton = true}) async {
    if (!mounted) return;

    final fetchGeneration = ++_fetchGeneration;

    final prefs = di.sl<SharedPreferences>();
    final db = di.sl<AppDatabase>();
    final now = DateTime.now();
    final startOfDay = DateTime(now.year, now.month, now.day, 0, 0, 0);
    final endOfDay = DateTime(now.year, now.month, now.day, 23, 59, 59, 999);
    final todayStr = DateFormat('yyyy-MM-dd').format(now);

    // =========================================================================
    // ⚡ STEP 1: FAST LOCAL-FIRST LOAD (< 30ms, INSTANT RENDERING, ZERO BUFFERING)
    // =========================================================================
    try {
      // 1a. Load cached KPI & Visit Plans from local storage if state is empty
      if (_kpiData == null) {
        final cachedKpiStr = prefs.getString(
          SalesApiService.performanceCacheKey(now),
        );
        if (cachedKpiStr != null) {
          try {
            _kpiData = jsonDecode(cachedKpiStr) as Map<String, dynamic>;
          } catch (_) {}
        }
      }

      if (_todayVisitPlans.isEmpty) {
        final cachedVisitsStr = prefs.getString(
          'cached_today_visit_plans_$todayStr',
        );
        if (cachedVisitsStr != null) {
          try {
            _todayVisitPlans = jsonDecode(cachedVisitsStr) as List<dynamic>;
          } catch (_) {}
        }
      }

      // 1b. Load Local SQLite Data immediately
      final localDbResults = await Future.wait([
        db.select(db.salesStock).get(),
        db.select(db.products).get(),
        (db.select(
          db.localTransactions,
        )..where((t) => t.syncStatus.equals('pending'))).get(),
        (db.select(db.localTransactions)
              ..orderBy([
                (t) => OrderingTerm(
                  expression: t.createdAt,
                  mode: OrderingMode.desc,
                ),
              ])
              ..limit(15))
            .get(),
        (db.select(db.localTransactions)..where(
              (t) =>
                  t.createdAt.isBiggerOrEqualValue(startOfDay) &
                  t.createdAt.isSmallerOrEqualValue(endOfDay),
            ))
            .get(),
      ]);

      final List<dynamic> stocks = localDbResults[0] as List<dynamic>;
      final List<dynamic> products = localDbResults[1] as List<dynamic>;
      final List<dynamic> pendingTransactionsList =
          localDbResults[2] as List<dynamic>;
      final List<dynamic> localTransactionsList =
          localDbResults[3] as List<dynamic>;
      final List<dynamic> todayTxns = localDbResults[4] as List<dynamic>;

      // Load pending deductions
      final Map<String, int> pendingDeductions = {};
      if (pendingTransactionsList.isNotEmpty) {
        final localIds = pendingTransactionsList
            .map((t) => t.localId as String)
            .toList();
        final ptItems = await (db.select(
          db.localTransactionItems,
        )..where((item) => item.transactionLocalId.isIn(localIds))).get();
        for (var item in ptItems) {
          final pid = item.productId;
          pendingDeductions[pid] =
              (pendingDeductions[pid] ?? 0) + item.quantity;
        }

        for (var tx in pendingTransactionsList) {
          final String? notes = tx.notes;
          if (notes != null && notes.isNotEmpty) {
            try {
              final parsed = jsonDecode(notes);
              if (parsed is Map && parsed.containsKey('bonus_items')) {
                final bonusList = parsed['bonus_items'] as List<dynamic>?;
                if (bonusList != null) {
                  for (var bonus in bonusList) {
                    final pid = bonus['product_id'] as String;
                    final qty = (bonus['quantity'] as num).toInt();
                    pendingDeductions[pid] =
                        (pendingDeductions[pid] ?? 0) + qty;
                  }
                }
              }
            } catch (_) {}
          }
        }
      }

      final List<Map<String, dynamic>> carriedStock = [];
      for (var s in stocks) {
        Product? matchedProd;
        for (var p in products) {
          if (p.id == s.productId) {
            matchedProd = p as Product;
            break;
          }
        }
        if (matchedProd != null) {
          final pendingDeduct = pendingDeductions[s.productId] ?? 0;
          final remainingStock = (s.quantity - pendingDeduct).clamp(0, 999999);
          if (remainingStock > 0) {
            carriedStock.add({
              'id': matchedProd.id,
              'name': matchedProd.name,
              'sku': matchedProd.sku ?? '',
              'unit': matchedProd.unit ?? 'PCS',
              'pcs_per_unit': matchedProd.pcsPerUnit,
              'stock': remainingStock,
            });
          }
        }
      }

      // Calculate today total from local transactions
      final validTodayLocalTxns = todayTxns.where(
        (t) => t.syncStatus != 'discarded',
      );
      final double localTodayTotal = validTodayLocalTxns.fold<double>(
        0.0,
        (sum, t) => sum + (t.totalAmount as num).toDouble(),
      );
      final double serverTodayTotal =
          ((_kpiData?['today_order_total'] as num?)?.toDouble() ?? 0.0);
      final double todayOrderTotal = localTodayTotal > serverTodayTotal
          ? localTodayTotal
          : serverTodayTotal;

      if (!mounted ||
          fetchGeneration != _fetchGeneration ||
          todayStr != DateFormat('yyyy-MM-dd').format(DateTime.now())) {
        return;
      }

      // Populate local pending & history list
      final retryMeta = await di.sl<DataIntegrityService>().getAllRetryMeta();
      final List<Map<String, dynamic>> localCombined = [];

      for (var t in localTransactionsList) {
        if (t.syncStatus == 'discarded') continue;
        final meta = retryMeta[t.localId] as Map<String, dynamic>?;
        final lastError = meta?['lastError'] as String?;
        final retryCount = meta?['retryCount'] as int? ?? 0;

        localCombined.add({
          'id': t.localId,
          'receipt_no':
              t.receiptNo ??
              (t.syncStatus == 'synced' ? 'INV-SYNCED' : 'PENDING (OFFLINE)'),
          'store': {'name': t.storeName},
          'total_amount': t.totalAmount,
          'status': t.syncStatus == 'synced' ? 'SYNCED' : 'PENDING',
          'is_local': true,
          'payment_method': t.paymentMethod,
          'created_at': t.createdAt,
          'last_error': lastError,
          'retry_count': retryCount,
        });
      }

      final completedCheckoutCount = todayTxns
          .where(
            (t) =>
                t.companyName != 'Instant Order' &&
                !(t.receiptNo?.contains('INSTANT') ?? false),
          )
          .map((t) => t.storeId)
          .toSet()
          .length;

      if (!mounted ||
          fetchGeneration != _fetchGeneration ||
          todayStr != DateFormat('yyyy-MM-dd').format(DateTime.now())) {
        return;
      }

      // Instantly render local state!
      if (mounted) {
        setState(() {
          _carriedStock = carriedStock;
          _todayOrderTotal = todayOrderTotal;
          _localTodayOrderTotal = localTodayTotal;
          _completedCheckoutCount = completedCheckoutCount;
          _pendingTransactions = localCombined;
          // If we already have local data or cached KPI, stop blocking spinner immediately!
          if (!showSkeleton || _carriedStock.isNotEmpty || _kpiData != null) {
            _isLoading = false;
          }
        });
      }
    } catch (localErr) {
      debugPrint('[DASHBOARD LOCAL LOAD ERROR]: $localErr');
    }

    // =========================================================================
    // 🌐 STEP 2: RESILIENT BACKGROUND REMOTE SYNC (FAST TIMEOUT & NON-BLOCKING)
    // =========================================================================
    try {
      final storeApi = StoreApiService(
        apiClient: di.sl<ApiClient>(),
        prefs: prefs,
      );

      // Launch remote calls with individual timeouts so one slow call never kills others
      final kpiFuture = _apiService
          .getMyPerformance()
          .timeout(const Duration(seconds: 7))
          .catchError((e) {
            debugPrint('[DASHBOARD] KPI remote fetch error: $e');
            return <String, dynamic>{};
          });

      final visitsFuture = _apiService
          .getVisitPlans(date: todayStr)
          .timeout(const Duration(seconds: 7))
          .catchError((e) {
            debugPrint('[DASHBOARD] Visit plans remote fetch error: $e');
            return <dynamic>[];
          });

      final storesFuture = storeApi
          .getStores()
          .timeout(const Duration(seconds: 8))
          .catchError((e) {
            debugPrint('[DASHBOARD] Stores remote fetch error: $e');
            return <StoreModel>[];
          });

      final remotePendingFuture = _apiService
          .getPendingTransactions()
          .timeout(const Duration(seconds: 6))
          .catchError((e) {
            debugPrint('[DASHBOARD] Remote pending fetch error: $e');
            return <dynamic>[];
          });

      final outstandingTempoFuture = _apiService
          .getOutstandingTempoTransactions()
          .timeout(const Duration(seconds: 6))
          .catchError((e) {
            debugPrint('[DASHBOARD] Outstanding tempo fetch error: $e');
            return <dynamic>[];
          });

      final locationFuture = Future<LatLng?>(() async {
        try {
          final permission = await Geolocator.checkPermission();
          if (permission == LocationPermission.always ||
              permission == LocationPermission.whileInUse) {
            Position? position;
            try {
              position = await Geolocator.getCurrentPosition(
                desiredAccuracy: LocationAccuracy.low,
                timeLimit: const Duration(milliseconds: 2000),
              );
            } catch (_) {
              position = await Geolocator.getLastKnownPosition();
            }
            if (position != null) {
              return LatLng(position.latitude, position.longitude);
            }
          }
        } catch (_) {}
        return null;
      }).timeout(const Duration(seconds: 3)).catchError((_) => null);

      final results = await Future.wait([
        kpiFuture,
        visitsFuture,
        storesFuture,
        remotePendingFuture,
        outstandingTempoFuture,
        locationFuture,
      ]);

      final kpi = results[0] as Map<String, dynamic>?;
      final visits = results[1] as List<dynamic>?;
      final allStores = results[2] as List<StoreModel>? ?? <StoreModel>[];
      final remotePending = results[3] as List<dynamic>?;
      final outstandingTempo = results[4] as List<dynamic>? ?? [];
      final currentPos = results[5] as LatLng?;

      if (!mounted ||
          fetchGeneration != _fetchGeneration ||
          todayStr != DateFormat('yyyy-MM-dd').format(DateTime.now())) {
        return;
      }

      // Save KPI to cache if succeeded
      if (kpi != null && kpi.isNotEmpty) {
        prefs
            .setString(
              SalesApiService.performanceCacheKey(now),
              jsonEncode(kpi),
            )
            .ignore();
      }

      // Enrich and optimize visit plans if received
      List<dynamic> optimizedVisits = _todayVisitPlans;
      if (visits != null && visits.isNotEmpty) {
        final enrichedVisits = visits.map((v) {
          final map = Map<String, dynamic>.from(v as Map);
          final storeId = map['store_id'];
          final matched = allStores.firstWhere(
            (s) => s.id == storeId,
            orElse: () =>
                StoreModel(id: '', name: '', ownerName: '', address: ''),
          );
          if (matched.id.isNotEmpty) {
            map['latitude'] = matched.latitude;
            map['longitude'] = matched.longitude;
          }
          return map;
        }).toList();

        optimizedVisits = _optimizeVisitRoute(enrichedVisits, currentPos);
        prefs
            .setString(
              'cached_today_visit_plans_$todayStr',
              jsonEncode(optimizedVisits),
            )
            .ignore();
      }

      // Merge remote pending transactions
      final List<Map<String, dynamic>> finalPending = List.from(
        _pendingTransactions,
      );
      if (remotePending != null && remotePending.isNotEmpty) {
        for (var t in remotePending) {
          final String? rLocalId = t['local_id'];
          final String? rReceiptNo = t['receipt_no'];
          final bool exists = finalPending.any(
            (c) =>
                c['id'] == rLocalId ||
                (rReceiptNo != null && c['receipt_no'] == rReceiptNo),
          );
          if (!exists) {
            finalPending.add({
              ...Map<String, dynamic>.from(t as Map),
              'is_local': false,
              'status': 'SUCCESS',
              'created_at':
                  DateTime.tryParse(t['transaction_date'] ?? '') ??
                  DateTime.now(),
            });
          }
        }
        finalPending.sort((a, b) {
          final DateTime dtA = a['created_at'] is DateTime
              ? a['created_at']
              : DateTime.now();
          final DateTime dtB = b['created_at'] is DateTime
              ? b['created_at']
              : DateTime.now();
          return dtB.compareTo(dtA);
        });
      }

      final hasRemoteSuccess =
          (kpi != null && kpi.isNotEmpty) ||
          (visits != null && visits.isNotEmpty) ||
          (remotePending != null && remotePending.isNotEmpty) ||
          allStores.isNotEmpty;

      bool isOffline = false;
      if (!hasRemoteSuccess) {
        final hasNet = await NetworkInfo().hasInternetAccess;
        isOffline = !hasNet;
      }

      if (mounted) {
        setState(() {
          if (kpi != null && kpi.isNotEmpty) {
            _kpiData = kpi;
            final double serverTodayTotal =
                ((kpi['today_order_total'] as num?)?.toDouble() ?? 0.0);
            _todayOrderTotal = serverTodayTotal > _localTodayOrderTotal
                ? serverTodayTotal
                : _localTodayOrderTotal;
          }
          if (visits != null && visits.isNotEmpty) {
            _todayVisitPlans = optimizedVisits;
          }
          _outstandingTempoTransactions = outstandingTempo;
          _pendingTransactions = finalPending;
          _isOfflineMode = isOffline;
        });
      }
    } catch (e) {
      debugPrint('[DASHBOARD REMOTE SYNC ERROR]: $e');
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  List<dynamic> _optimizeVisitRoute(List<dynamic> visits, LatLng? currentPos) {
    if (visits.isEmpty) return [];

    List<dynamic> unvisited = List.from(visits);
    List<dynamic> ordered = [];

    // Default starting point (e.g. Surabaya city center coordinate or current position)
    LatLng current = currentPos ?? const LatLng(-7.250445, 112.768845);

    while (unvisited.isNotEmpty) {
      dynamic nearest;
      double minDistance = double.infinity;

      for (var visit in unvisited) {
        final double? lat = visit['latitude'] != null
            ? (visit['latitude'] as num).toDouble()
            : null;
        final double? lng = visit['longitude'] != null
            ? (visit['longitude'] as num).toDouble()
            : null;

        if (lat == null || lng == null || (lat == 0.0 && lng == 0.0)) {
          continue;
        }

        double dist = _haversineDistance(current, LatLng(lat, lng));
        if (dist < minDistance) {
          minDistance = dist;
          nearest = visit;
        }
      }

      if (nearest == null) {
        ordered.addAll(unvisited);
        break;
      }

      ordered.add(nearest);
      final double nextLat = (nearest['latitude'] as num).toDouble();
      final double nextLng = (nearest['longitude'] as num).toDouble();
      current = LatLng(nextLat, nextLng);
      unvisited.remove(nearest);
    }

    return ordered;
  }

  double _haversineDistance(LatLng p1, LatLng p2) {
    var p = 0.017453292519943295;
    var a =
        0.5 -
        cos((p2.latitude - p1.latitude) * p) / 2 +
        cos(p1.latitude * p) *
            cos(p2.latitude * p) *
            (1 - cos((p2.longitude - p1.longitude) * p)) /
            2;
    return 12742 * asin(sqrt(a)); // 2 * R; R = 6371 km
  }

  @override
  void dispose() {
    _dayRolloverTimer?.cancel();
    WidgetsBinding.instance.removeObserver(this);
    visitScheduleRouteObserver.unsubscribe(this);
    _pageController.dispose();
    super.dispose();
  }

  Widget _buildSwipeableHeader(String jobPosition) {
    return Column(
      children: [
        SizedBox(
          height: 360,
          child: PageView(
            controller: _pageController,
            physics: const BouncingScrollPhysics(),
            pageSnapping: true,
            onPageChanged: (v) => setState(() => _currentPage = v),
            children: [
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 2),
                child: _buildVisitSummaryCard(),
              ),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 2),
                child: _buildTargetCard(jobPosition),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: List.generate(
            2,
            (index) => Container(
              margin: const EdgeInsets.symmetric(horizontal: 4),
              width: _currentPage == index ? 24 : 8,
              height: 8,
              decoration: BoxDecoration(
                color: _currentPage == index
                    ? AppColors.blueSales
                    : const Color(0xFFCBD5E1),
                borderRadius: BorderRadius.circular(4),
              ),
            ),
          ),
        ),
      ],
    );
  }

  String _formatStock(int totalPcs, int multiplier, String bulkUnit) {
    if (multiplier <= 1) return '$totalPcs Pcs';
    final int bulk = totalPcs ~/ multiplier;
    final int loose = totalPcs % multiplier;
    if (bulk > 0 && loose > 0) {
      return '$bulk $bulkUnit + $loose Pcs';
    } else if (bulk > 0) {
      return '$bulk $bulkUnit';
    } else {
      return '$loose Pcs';
    }
  }

  Widget _buildRouteRoadmap() {
    if (_todayVisitPlans.isEmpty) {
      return const SizedBox.shrink();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 8),
        Row(
          children: [
            const Icon(
              Icons.navigation_rounded,
              size: 10,
              color: AppColors.blueSales,
            ),
            const SizedBox(width: 6),
            Text(
              'URUTAN RUTE KUNJUNGAN',
              style: GoogleFonts.outfit(
                fontSize: 8,
                fontWeight: FontWeight.w900,
                color: AppColors.blueSales,
                letterSpacing: 1.2,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        SizedBox(
          height: 32,
          child: ListView.builder(
            scrollDirection: Axis.horizontal,
            physics: const BouncingScrollPhysics(),
            itemCount: _todayVisitPlans.length,
            itemBuilder: (context, index) {
              final plan = _todayVisitPlans[index];
              final storeName = plan['store_name'] ?? 'Toko';
              final isCompleted =
                  plan['status'] == 'COMPLETED' || plan['status'] == 'EXTRA';
              final isLast = index == _todayVisitPlans.length - 1;

              return Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: isCompleted
                          ? const Color(0xFFDCFCE7)
                          : const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(
                        color: isCompleted
                            ? const Color(0xFF22C55E).withOpacity(0.3)
                            : const Color(0xFFE2E8F0),
                        width: 1,
                      ),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          isCompleted
                              ? Icons.check_circle_rounded
                              : Icons.storefront_rounded,
                          size: 11,
                          color: isCompleted
                              ? const Color(0xFF15803D)
                              : AppColors.blueSales,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          storeName,
                          style: GoogleFonts.outfit(
                            fontSize: 9,
                            fontWeight: FontWeight.w800,
                            color: isCompleted
                                ? const Color(0xFF15803D)
                                : const Color(0xFF334155),
                          ),
                        ),
                      ],
                    ),
                  ),
                  if (!isLast) ...[
                    const SizedBox(width: 6),
                    const Icon(
                      Icons.arrow_forward_rounded,
                      size: 12,
                      color: Color(0xFF94A3B8),
                    ),
                    const SizedBox(width: 6),
                  ],
                ],
              );
            },
          ),
        ),
      ],
    );
  }

  Widget _buildVisitSummaryCard() {
    final currencyFormat = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    final List<dynamic> visits = _todayVisitPlans;
    final int totalVisits = visits.length;
    final int targetVisits =
        ((_kpiData?['daily_missions']?['target_visits'] as num?)?.toInt() ??
        (totalVisits > 0 ? totalVisits : 1));
    final int completedVisits = _completedCheckoutCount;
    final double progress = targetVisits > 0
        ? completedVisits / targetVisits
        : 0.0;

    return InkWell(
      onTap: () => Navigator.of(
        context,
      ).push(MaterialPageRoute(builder: (_) => const VisitSchedulePage())),
      child: Container(
        width: double.infinity,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          boxShadow: [
            BoxShadow(
              color: AppColors.blueSalesDark.withOpacity(0.08),
              blurRadius: 20,
              offset: const Offset(0, 10),
            ),
          ],
          border: Border.all(
            color: AppColors.blueSales.withOpacity(0.1),
            width: 1.5,
          ),
        ),
        child: Stack(
          children: [
            Positioned(
              right: -10,
              top: -10,
              child: Icon(
                Icons.route_rounded,
                size: 100,
                color: AppColors.blueSalesLight.withOpacity(0.04),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: AppColors.blueSalesLight.withOpacity(0.1),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: const Icon(
                              Icons.calendar_today_rounded,
                              color: AppColors.blueSales,
                              size: 20,
                            ),
                          ),
                          const SizedBox(width: 12),
                          Text(
                            widget.isMT
                                ? 'RENCANA KUNJUNGAN MITRA'
                                : 'RENCANA KUNJUNGAN',
                            style: GoogleFonts.outfit(
                              color: AppColors.blueSalesDark,
                              fontWeight: FontWeight.w800,
                              fontSize: 11,
                              letterSpacing: 1.5,
                            ),
                          ),
                        ],
                      ),
                      Text(
                        DateFormat('dd MMM').format(DateTime.now()),
                        style: GoogleFonts.outfit(
                          color: AppColors.blueSales,
                          fontWeight: FontWeight.w700,
                          fontSize: 11,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),
                  Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              '$completedVisits / $targetVisits',
                              style: GoogleFonts.outfit(
                                color: const Color(0xFF0F172A),
                                fontWeight: FontWeight.w900,
                                fontSize: 24,
                              ),
                            ),
                            Text(
                              widget.isMT
                                  ? 'Mitra dikunjungi hari ini'
                                  : 'Toko dikunjungi hari ini',
                              style: GoogleFonts.outfit(
                                color: Colors.blueGrey,
                                fontSize: 11,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                            if (totalVisits < targetVisits)
                              Container(
                                margin: const EdgeInsets.only(top: 6),
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 6,
                                  vertical: 3,
                                ),
                                decoration: BoxDecoration(
                                  color: Colors.orange.shade50,
                                  borderRadius: BorderRadius.circular(6),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Icon(
                                      Icons.warning_amber_rounded,
                                      size: 10,
                                      color: Colors.orange.shade700,
                                    ),
                                    const SizedBox(width: 4),
                                    Text(
                                      'Kurang ${targetVisits - totalVisits} ${widget.isMT ? 'mitra' : 'toko'} dari target harian',
                                      style: GoogleFonts.outfit(
                                        color: Colors.orange.shade800,
                                        fontSize: 9,
                                        fontWeight: FontWeight.w800,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                          ],
                        ),
                      ),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            currencyFormat.format(_todayOrderTotal),
                            style: GoogleFonts.outfit(
                              fontSize: 16,
                              fontWeight: FontWeight.w900,
                              color: const Color(
                                0xFF10B981,
                              ), // Emerald green for premium financial design
                            ),
                          ),
                          Text(
                            'Transaksi Hari ini',
                            style: GoogleFonts.outfit(
                              fontSize: 9,
                              fontWeight: FontWeight.w700,
                              color: Colors.blueGrey.shade400,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                  _buildRouteRoadmap(),
                  const SizedBox(height: 10),
                  const Divider(height: 1, color: Color(0xFFF1F5F9)),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      const Icon(
                        Icons.inventory_2_rounded,
                        size: 14,
                        color: AppColors.blueSales,
                      ),
                      const SizedBox(width: 6),
                      Text(
                        'STOK YANG DIBAWA SAAT INI',
                        style: GoogleFonts.outfit(
                          fontSize: 10,
                          fontWeight: FontWeight.w900,
                          color: AppColors.blueSales,
                          letterSpacing: 1.2,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  if ((_carriedStock ?? []).isEmpty)
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.symmetric(
                        vertical: 10,
                        horizontal: 12,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Row(
                        children: [
                          Icon(
                            Icons.info_outline_rounded,
                            size: 14,
                            color: Colors.orange.shade700,
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              'Belum ada barang yang dibawa.',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                color: Colors.blueGrey,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ),
                        ],
                      ),
                    )
                  else
                    SizedBox(
                      height: 52,
                      child: ListView.builder(
                        scrollDirection: Axis.horizontal,
                        itemCount: (_carriedStock ?? []).length,
                        itemBuilder: (context, index) {
                          final item = (_carriedStock ?? [])[index];
                          final name = item['name'] as String;
                          final totalPcs = item['stock'] as int;
                          final multiplier = item['pcs_per_unit'] as int;
                          final unit = item['unit'] as String;

                          final displayStock = _formatStock(
                            totalPcs,
                            multiplier,
                            unit,
                          );

                          return Container(
                            margin: const EdgeInsets.only(right: 10),
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 6,
                            ),
                            decoration: BoxDecoration(
                              color: const Color(0xFFF0FDF4),
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(
                                color: const Color(0xFFBBF7D0),
                              ),
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Container(
                                  padding: const EdgeInsets.all(4),
                                  decoration: const BoxDecoration(
                                    color: Color(0xFFDCFCE7),
                                    shape: BoxShape.circle,
                                  ),
                                  child: const Icon(
                                    Icons.circle,
                                    size: 6,
                                    color: Color(0xFF16A34A),
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Text(
                                      name.length > 16
                                          ? '${name.substring(0, 14)}..'
                                          : name,
                                      style: GoogleFonts.outfit(
                                        fontSize: 10,
                                        fontWeight: FontWeight.w800,
                                        color: const Color(0xFF1E293B),
                                      ),
                                    ),
                                    Text(
                                      displayStock,
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w900,
                                        color: const Color(0xFF15803D),
                                      ),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          );
                        },
                      ),
                    ),
                  const SizedBox(height: 16),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    decoration: BoxDecoration(
                      color: AppColors.blueSales,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Center(
                      child: Text(
                        widget.isMT
                            ? 'LIHAT RENCANA KUNJUNGAN MITRA'
                            : 'LIHAT RENCANA KUNJUNGAN',
                        style: GoogleFonts.outfit(
                          color: Colors.white,
                          fontWeight: FontWeight.w800,
                          fontSize: 10,
                          letterSpacing: 1,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSkeleton() {
    return Shimmer.fromColors(
      baseColor: Colors.grey.shade300,
      highlightColor: Colors.grey.shade100,
      child: SingleChildScrollView(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 10),
            Container(
              height: 360,
              width: double.infinity,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
              ),
            ),
            const SizedBox(height: 24),
            Row(
              children: [
                Expanded(
                  child: Container(
                    height: 100,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Container(
                    height: 100,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: Container(
                    height: 100,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Container(
                    height: 100,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 24),
            Container(
              height: 80,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
              ),
            ),
            const SizedBox(height: 12),
            Container(
              height: 80,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC), // Clean Light Ice Blue
      appBar: AppBar(
        title: BlocBuilder<ProfileBloc, ProfileState>(
          builder: (context, profileState) {
            String companyName = 'Perusahaan Demo';
            String title = 'SALES PERFORMANCE';

            Map<String, dynamic>? profile;
            if (profileState is ProfileLoaded) {
              profile = profileState.profile;
            } else {
              final authState = context.read<AuthBloc>().state;
              if (authState is Authenticated) {
                profile = authState.userProfile;
              }
            }

            if (profile != null) {
              companyName = profile['company']?['name'] ?? 'Perusahaan Demo';
              final String position = (profile['job_position']?['title'] ?? '')
                  .toString()
                  .toLowerCase();

              final role = (profile['role'] ?? '').toString().toLowerCase();
              final isSpv =
                  role.contains('supervisor') ||
                  role.contains('spv') ||
                  position.contains('supervisor') ||
                  position.contains('spv');

              if (widget.isMT) {
                title = 'SALES MT (MODERN TRADE)';
              } else if (widget.forceTO ||
                  isSpv ||
                  position.contains('task order') ||
                  position.contains('sales to') ||
                  position.split(' ').contains('to')) {
                title = isSpv ? 'SALES TO (SUPERVISOR)' : 'TASK ORDER SALES';
              } else if (position.contains('booster')) {
                title = 'SALES BOOSTER';
              } else {
                title = 'SALES MOTORIS';
              }
            }

            return Padding(
              padding: const EdgeInsets.symmetric(horizontal: 4),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  FittedBox(
                    fit: BoxFit.scaleDown,
                    alignment: Alignment.centerLeft,
                    child: Text(
                      title,
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                        fontSize: 13,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ),
                  const SizedBox(height: 1),
                  FittedBox(
                    fit: BoxFit.scaleDown,
                    alignment: Alignment.centerLeft,
                    child: Text(
                      widget.isMT
                          ? 'Perusahaan Demo'
                          : companyName.toUpperCase(),
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w700,
                        color: AppColors.blueSalesLight,
                        fontSize: 8.5,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ),
                ],
              ),
            );
          },
        ),
        backgroundColor: widget.isMT
            ? const Color(0xFF059669)
            : AppColors.blueSalesDark, // Emerald for MT
        centerTitle: false,
        elevation: 0,
        automaticallyImplyLeading: false,
        leadingWidth: 76,
        leading: Center(
          child: Padding(
            padding: const EdgeInsets.only(left: 8),
            child: InkWell(
              onTap: () => Navigator.pop(context),
              borderRadius: BorderRadius.circular(10),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(
                      Icons.arrow_back_ios_new_rounded,
                      size: 11,
                      color: Colors.white,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      'HRIS',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w800,
                        fontSize: 10,
                        color: Colors.white,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
        actions: [
          BlocBuilder<SyncBloc, SyncState>(
            builder: (context, state) {
              return IconButton(
                padding: const EdgeInsets.all(6),
                constraints: const BoxConstraints(minWidth: 34, minHeight: 34),
                visualDensity: VisualDensity.compact,
                onPressed: state is SyncInProgress
                    ? null
                    : () {
                        context.read<SyncBloc>().add(
                          const SyncDataRequested(force: true, isManual: true),
                        );
                      },
                icon: state is SyncInProgress
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(
                        Icons.sync_rounded,
                        color: Colors.white,
                        size: 20,
                      ),
                tooltip: 'Sinkronisasi Data (Upload & Download)',
              );
            },
          ),
          IconButton(
            padding: const EdgeInsets.all(6),
            constraints: const BoxConstraints(minWidth: 34, minHeight: 34),
            visualDensity: VisualDensity.compact,
            onPressed: () => _exportOfflineBackup(closeDrawer: false),
            icon: const Icon(
              Icons.backup_rounded,
              color: Colors.white,
              size: 20,
            ),
            tooltip: 'Cadangkan Transaksi Offline',
          ),
          IconButton(
            padding: const EdgeInsets.all(6),
            constraints: const BoxConstraints(minWidth: 34, minHeight: 34),
            visualDensity: VisualDensity.compact,
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => const DeliveryTrackingPage()),
            ),
            icon: const Icon(
              Icons.qr_code_scanner_rounded,
              color: Colors.white,
              size: 20,
            ),
            tooltip: 'Pelacakan Pengiriman',
          ),
          const SizedBox(width: 4),
        ],
      ),
      body: BlocListener<SyncBloc, SyncState>(
        listener: (context, state) {
          if (state is SyncSuccess) {
            _fetchData(showSkeleton: false);
            final successCount = state.result?.successCount ?? 0;
            if (successCount > 0) {
              ScaffoldMessenger.of(context).hideCurrentSnackBar();
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text(
                    '✓ $successCount data offline berhasil disinkronkan ke server!',
                  ),
                  backgroundColor: const Color(0xFF10B981),
                  behavior: SnackBarBehavior.floating,
                  showCloseIcon: true,
                  closeIconColor: Colors.white,
                  duration: const Duration(seconds: 3),
                ),
              );
            }
          } else if (state is SyncPartialSuccess) {
            _fetchData(showSkeleton: false);
            // Jangan tampilkan pop-up merah yang mengganggu jika sedang mode offline dan sync bukan manual
            if (_isOfflineMode && !state.isManual) return;

            final firstError = state.errors.isNotEmpty
                ? ': ${state.errors.first}'
                : '';
            final isSpandukOnly =
                state.errors.isNotEmpty &&
                state.errors.every((e) => e.startsWith('Order Spanduk:'));
            final labelText = isSpandukOnly
                ? '${state.failureCount} order spanduk offline gagal dikirim$firstError'
                : '${state.failureCount} data offline gagal dikirim$firstError';
            ScaffoldMessenger.of(context).hideCurrentSnackBar();
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(labelText),
                backgroundColor: const Color(0xFFE11D48),
                behavior: SnackBarBehavior.floating,
                showCloseIcon: true,
                closeIconColor: Colors.white,
                duration: const Duration(seconds: 5),
                action: SnackBarAction(
                  label: 'SYNC ULANG',
                  textColor: Colors.white,
                  onPressed: () {
                    ScaffoldMessenger.of(context).hideCurrentSnackBar();
                    context.read<SyncBloc>().add(
                      const SyncDataRequested(force: true, isManual: true),
                    );
                  },
                ),
              ),
            );
          } else if (state is SyncMasterDataSuccess) {
            _fetchData(showSkeleton: false);
          } else if (state is SyncFailure) {
            _fetchData(showSkeleton: false);
            if (_isOfflineMode && !state.isManual) return;

            ScaffoldMessenger.of(context).hideCurrentSnackBar();
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(
                  state.error.startsWith('Mode Offline aktif')
                      ? state.error
                      : 'Sync gagal (${state.error.substring(0, state.error.length.clamp(0, 60))}). Data aman di lokal.',
                ),
                backgroundColor: Colors.orange[800],
                behavior: SnackBarBehavior.floating,
                showCloseIcon: true,
                closeIconColor: Colors.white,
                duration: const Duration(seconds: 4),
                action: SnackBarAction(
                  label: 'COBA LAGI',
                  textColor: Colors.white,
                  onPressed: () {
                    ScaffoldMessenger.of(context).hideCurrentSnackBar();
                    context.read<SyncBloc>().add(
                      const SyncDataRequested(force: true, isManual: true),
                    );
                  },
                ),
              ),
            );
          }
          // SyncIntegrityWarning ditangani oleh SyncStatusBanner
        },
        child: _isLoading
            ? _buildSkeleton()
            : RefreshIndicator(
                onRefresh: () async {
                  context.read<SyncBloc>().add(
                    const SyncDataRequested(force: true, isManual: true),
                  );
                  await _fetchData(showSkeleton: false);
                },
                child: SingleChildScrollView(
                  physics: const BouncingScrollPhysics(),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 12,
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const SyncStatusBanner(),
                      if (_isOfflineMode)
                        Container(
                          margin: const EdgeInsets.only(top: 8),
                          padding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 8,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.orange.shade100,
                            border: Border.all(color: Colors.orange.shade300),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Row(
                            children: [
                              Icon(
                                Icons.wifi_off_rounded,
                                color: Colors.orange.shade800,
                                size: 20,
                              ),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  'Koneksi lambat. Menampilkan data offline/lokal terakhir apabila koneksi dirasa bagus silahkan refresh .',
                                  style: GoogleFonts.outfit(
                                    color: Colors.orange.shade900,
                                    fontSize: 12,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      BlocBuilder<ProfileBloc, ProfileState>(
                        builder: (context, state) {
                          String position = '';
                          if (state is ProfileLoaded) {
                            position =
                                state.profile['job_position']?['title'] ?? '';
                          }
                          return Column(
                            children: [
                              _buildSwipeableHeader(position),
                              const SizedBox(height: 24),
                              if (_outstandingTempoTransactions.isNotEmpty) ...[
                                _buildOutstandingTempoList(),
                                const SizedBox(height: 24),
                              ],
                              _buildQuickAccess(position),
                            ],
                          );
                        },
                      ),
                      if (_pendingTransactions.isNotEmpty) ...[
                        const SizedBox(height: 32),
                        _buildPendingTransactionsList(),
                      ],
                      const SizedBox(
                        height: 40,
                      ), // extra bottom padding for comfort
                    ],
                  ),
                ),
              ),
      ),
    );
  }

  Widget _buildTargetCard(String jobPosition) {
    final List<dynamic> items = _kpiData?['items'] ?? [];
    if (items.isEmpty) {
      return Container(
        width: double.infinity,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 20,
              offset: const Offset(0, 10),
            ),
          ],
          border: Border.all(color: const Color(0xFFE2E8F0), width: 1.5),
        ),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.amber.shade50,
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  Icons.assignment_late_rounded,
                  color: Colors.amber.shade800,
                  size: 36,
                ),
              ),
              const SizedBox(height: 16),
              Text(
                'TARGET OMZET HARI INI',
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  fontWeight: FontWeight.w800,
                  color: Colors.blueGrey,
                  letterSpacing: 1.5,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Belum Ada Target',
                style: GoogleFonts.outfit(
                  fontSize: 18,
                  fontWeight: FontWeight.w900,
                  color: const Color(0xFF1E293B),
                ),
              ),
              const SizedBox(height: 6),
              Text(
                'Target performa harian Anda belum diatur oleh supervisor di dashboard manajemen.',
                textAlign: TextAlign.center,
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  color: Colors.blueGrey,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      );
    }

    Map<String, dynamic>? profile;
    final profileState = context.read<ProfileBloc>().state;
    if (profileState is ProfileLoaded) {
      profile = profileState.profile;
    } else {
      final authState = context.read<AuthBloc>().state;
      if (authState is Authenticated) {
        profile = authState.userProfile;
      }
    }

    final String roleLower = (profile?['role'] ?? '').toString().toLowerCase();
    final bool isSpv =
        roleLower.contains('supervisor') ||
        roleLower.contains('spv') ||
        jobPosition.toLowerCase().contains('supervisor') ||
        jobPosition.toLowerCase().contains('spv');
    final bool isBooster = jobPosition.toLowerCase().contains('booster');
    final bool isTO =
        widget.forceTO ||
        isSpv ||
        jobPosition.contains('Task Order') ||
        jobPosition.contains('Sales TO') ||
        jobPosition.toLowerCase().contains('take order');

    // Title & Icon based on role
    String cardTitle = 'PERFORMA OMZET';
    IconData cardIcon = Icons.auto_graph_rounded;
    Color themeColor = AppColors.blueSales;

    if (isBooster) {
      cardTitle = 'PERFORMA BOOSTER';
      cardIcon = Icons.monetization_on_rounded;
      themeColor = const Color(0xFF10B981);
    } else if (isTO) {
      cardTitle = 'PERFORMA DISTRIBUSI';
      cardIcon = Icons.local_shipping_rounded;
      themeColor = AppColors.blueWarehouse; // Steel/Sky Blue for TO
    }

    // Determine Metrics from Items
    final String primaryLabel = items[0]['label'] ?? 'Target';
    final double targetValue = (items[0]['score'] as num?)?.toDouble() ?? 0.0;
    final double actualValue = items.length > 1
        ? ((items[1]['score'] as num?)?.toDouble() ?? 0.0)
        : 0.0;
    final double commission =
        (_kpiData?['commission'] as num?)?.toDouble() ?? 0.0;
    final double achievement =
        (_kpiData?['achievement_percentage'] as num?)?.toDouble() ??
        (targetValue > 0 ? (actualValue / targetValue * 100) : 0.0);
    final double remaining = targetValue - actualValue;

    final bool isCurrency =
        primaryLabel.toLowerCase().contains('omzet') ||
        primaryLabel.toLowerCase().contains('rupiah');
    final currencyFormatter = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    final numberFormatter = NumberFormat.decimalPattern('id');

    String formatValue(double val) => isCurrency
        ? currencyFormatter.format(val)
        : numberFormatter.format(val);

    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: themeColor.withOpacity(0.08),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
        border: Border.all(color: themeColor.withOpacity(0.1), width: 1.5),
      ),
      child: Stack(
        children: [
          Positioned(
            right: -10,
            top: -10,
            child: Icon(
              cardIcon,
              size: 100,
              color: themeColor.withOpacity(0.03),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: themeColor.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Icon(cardIcon, color: themeColor, size: 20),
                    ),
                    const SizedBox(width: 12),
                    Text(
                      cardTitle,
                      style: GoogleFonts.outfit(
                        color: themeColor,
                        fontWeight: FontWeight.w800,
                        fontSize: 11,
                        letterSpacing: 1.5,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                FittedBox(
                  fit: BoxFit.scaleDown,
                  child: Text(
                    formatValue(targetValue),
                    style: GoogleFonts.outfit(
                      color: const Color(0xFF0F172A),
                      fontWeight: FontWeight.w800,
                      fontSize: 20,
                    ),
                  ),
                ),
                Text(
                  primaryLabel,
                  style: GoogleFonts.outfit(
                    color: Colors.blueGrey,
                    fontSize: 11,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                if (commission > 0) ...[
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.amber.shade50,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.amber.shade200),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          Icons.monetization_on_rounded,
                          size: 12,
                          color: Colors.amber.shade700,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          'Potensi Komisi: ${currencyFormatter.format(commission)}',
                          style: GoogleFonts.outfit(
                            fontSize: 10,
                            fontWeight: FontWeight.w800,
                            color: Colors.amber.shade800,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
                const SizedBox(height: 20),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: _buildStatItem(
                        'Realisasi',
                        formatValue(actualValue),
                        themeColor,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: _buildStatItem(
                        'Sisa',
                        formatValue(remaining < 0 ? 0 : remaining),
                        const Color(0xFFB91C1C),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: _buildStatItem(
                        'Pencapaian',
                        '${achievement.toStringAsFixed(1)}%',
                        themeColor,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                ClipRRect(
                  borderRadius: BorderRadius.circular(20),
                  child: LinearProgressIndicator(
                    value: targetValue > 0
                        ? (actualValue / targetValue).clamp(0.0, 1.0)
                        : 0,
                    minHeight: 8,
                    backgroundColor: themeColor.withOpacity(0.05),
                    valueColor: AlwaysStoppedAnimation<Color>(themeColor),
                  ),
                ),
                const SizedBox(height: 12),
                const Divider(height: 1, color: Color(0xFFF1F5F9)),
                const SizedBox(height: 10),
                Row(
                  children: [
                    Icon(
                      Icons.assignment_late_rounded,
                      size: 12,
                      color: Colors.orange.shade700,
                    ),
                    const SizedBox(width: 6),
                    Text(
                      'ESTIMASI PIUTANG TEMPO',
                      style: GoogleFonts.outfit(
                        fontSize: 9,
                        fontWeight: FontWeight.w900,
                        color: Colors.orange.shade800,
                        letterSpacing: 1.0,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                if (_outstandingTempoTransactions.isEmpty)
                  Text(
                    'Semua tagihan tempo aman / lunas.',
                    style: GoogleFonts.outfit(
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                      color: const Color(0xFF10B981),
                    ),
                  )
                else
                  _buildTempoSummaryRow(
                    NumberFormat.currency(
                      locale: 'id_ID',
                      symbol: 'Rp',
                      decimalDigits: 0,
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatItem(String label, String value, Color color) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label.toUpperCase(),
          style: GoogleFonts.outfit(
            fontSize: 9,
            fontWeight: FontWeight.w800,
            color: Colors.blueGrey,
            letterSpacing: 1,
          ),
        ),
        const SizedBox(height: 4),
        FittedBox(
          fit: BoxFit.scaleDown,
          child: Text(
            value,
            style: GoogleFonts.outfit(
              fontSize: 14,
              fontWeight: FontWeight.w800,
              color: color,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildTempoSummaryRow(NumberFormat currencyFormatter) {
    final firstTempo = _outstandingTempoTransactions.first;
    final storeName = firstTempo['store']?['name'] ?? 'Toko';
    final outstanding =
        ((firstTempo['total_amount'] as num?)?.toDouble() ?? 0.0) -
        ((firstTempo['paid_amount'] as num?)?.toDouble() ?? 0.0);
    final dueDate = firstTempo['payment_due_date'] != null
        ? DateFormat('dd MMM yyyy').format(
            DateTime.parse(firstTempo['payment_due_date'].toString()).toLocal(),
          )
        : '-';

    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                storeName,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  color: const Color(0xFF1E293B),
                ),
              ),
              Text(
                'Estimasi Bayar: $dueDate',
                style: GoogleFonts.outfit(
                  fontSize: 9,
                  fontWeight: FontWeight.w600,
                  color: Colors.orange.shade800,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(width: 12),
        Text(
          currencyFormatter.format(outstanding),
          style: GoogleFonts.outfit(
            fontSize: 12,
            fontWeight: FontWeight.w900,
            color: const Color(0xFFB91C1C),
          ),
        ),
      ],
    );
  }

  Widget _buildQuickAccess(String jobPosition) {
    Map<String, dynamic>? profile;
    final profileState = context.read<ProfileBloc>().state;
    if (profileState is ProfileLoaded) {
      profile = profileState.profile;
    } else {
      final authState = context.read<AuthBloc>().state;
      if (authState is Authenticated) {
        profile = authState.userProfile;
      }
    }

    final String roleLower = (profile?['role'] ?? '').toString().toLowerCase();
    final String distStoreId =
        profile?['distributor_store_id']?.toString() ?? '';
    final bool isDistributor =
        profile?['is_distributor'] == true ||
        distStoreId.isNotEmpty ||
        jobPosition.toLowerCase().contains('distributor');
    final bool isSpv =
        roleLower.contains('supervisor') ||
        roleLower.contains('spv') ||
        jobPosition.toLowerCase().contains('supervisor') ||
        jobPosition.toLowerCase().contains('spv');
    final bool isSalesTO =
        widget.forceTO ||
        isSpv ||
        jobPosition.contains('Task Order') ||
        jobPosition.contains('Sales TO') ||
        jobPosition.toLowerCase().contains('take order');

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'AKSES OPERASIONAL',
              style: GoogleFonts.outfit(
                fontSize: 11,
                fontWeight: FontWeight.w900,
                color: const Color(0xFF64748B),
                letterSpacing: 1.5,
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        GridView.count(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          crossAxisCount: 3,
          mainAxisSpacing: 12,
          crossAxisSpacing: 12,
          childAspectRatio: 0.95,
          children: [
            _buildActionItem(
              Icons.flash_on_rounded,
              'Order Instan',
              const Color(0xFFF59E0B),
              'Order Cepat',
            ),
            _buildActionItem(
              Icons.storefront_rounded,
              'Kunjungan',
              const Color(0xFF0D9488),
              'Jadwal Visit',
            ),
            _buildActionItem(
              Icons.view_list_rounded,
              widget.isMT ? 'Mitra' : 'Toko',
              const Color(0xFF0891B2),
              widget.isMT ? 'Kelola Mitra' : 'Kelola Pelanggan',
            ),
            _buildActionItem(
              Icons.edit_calendar_rounded,
              'Kelola Rencana',
              const Color(0xFF8B5CF6),
              'Planning Visit',
            ),
            if (!isDistributor)
              _buildActionItem(
                Icons.menu_book_rounded,
                'Katalog',
                const Color(0xFF4F46E5),
                'Produk & Stok',
              ),
            _buildActionItem(
              Icons.history_edu_rounded,
              'Riwayat',
              const Color(0xFFB45309),
              'Nota & Order',
            ),
            _buildActionItem(
              Icons.speed_rounded,
              'Riwayat KM',
              const Color(0xFF0284C7),
              'Odometer & BBM',
            ),
            // Sales MT & TO tidak membawa stok fisik sendiri - stok dikelola pabrik
            if (!isSalesTO && !widget.isMT)
              _buildActionItem(
                Icons.move_to_inbox_rounded,
                'Ambil Barang',
                const Color(0xFF15803D),
                'Stok Dari Gudang',
              ),
            _buildActionItem(
              Icons.receipt_long_rounded,
              'Reimbursement',
              const Color(0xFFE11D48),
              'Pengajuan Biaya',
            ),
            // MT-specific: PO list with tax documents download
            if (widget.isMT)
              _buildActionItem(
                Icons.assignment_turned_in_rounded,
                'PO MT',
                const Color(0xFF10B981),
                'PO & Pajak MT',
              ),
            if (widget.isMT)
              _buildActionItem(
                Icons.layers_rounded,
                'Merchandising',
                const Color(0xFFD97706),
                'Audit & Planogram',
              ),
          ],
        ),
      ],
    );
  }

  void _showStoreManagementOptions(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
      ),
      builder: (context) => SafeArea(
        child: SingleChildScrollView(
          child: Container(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  widget.isMT ? 'Manajemen Mitra' : 'Manajemen Toko',
                  style: GoogleFonts.outfit(
                    fontSize: 18,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 24),
                _buildDialogOption(
                  context,
                  Icons.view_list_rounded,
                  widget.isMT ? 'Daftar & Edit Mitra' : 'Daftar & Edit Toko',
                  widget.isMT
                      ? 'Kelola semua database mitra modern Anda'
                      : 'Kelola semua database outlet Anda',
                  () {
                    Navigator.pop(context);
                    Navigator.of(context).push(
                      MaterialPageRoute(builder: (_) => const StoreListPage()),
                    );
                  },
                ),
                const SizedBox(height: 12),
                _buildDialogOption(
                  context,
                  Icons.art_track_rounded,
                  'Order Spanduk',
                  'Pemesanan spanduk promosi outlet',
                  () {
                    Navigator.pop(context);
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => const OrderBannerPage(),
                      ),
                    );
                  },
                ),
                const SizedBox(height: 12),
                _buildDialogOption(
                  context,
                  Icons.photo_camera_rounded,
                  'Pasang Spanduk & Reimburse',
                  'Upload bukti pasang & klaim reimbursement',
                  () {
                    Navigator.pop(context);
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => const BannerInstallPage(),
                      ),
                    );
                  },
                ),
                const SizedBox(height: 24),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void _showStockManagementOptions(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
      ),
      builder: (context) => SafeArea(
        child: SingleChildScrollView(
          child: Container(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  'Perpindahan & Stok Sales',
                  style: GoogleFonts.outfit(
                    fontSize: 18,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 24),
                _buildDialogOption(
                  context,
                  Icons.input_rounded,
                  'Ambil Barang Masuk',
                  'Request penambahan stok dari gudang ke sales',
                  () async {
                    Navigator.pop(context);
                    final result = await Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) =>
                            const SalesStockRequestPage(isReturn: false),
                      ),
                    );
                    if (result == true && mounted) {
                      _fetchData(showSkeleton: false);
                    }
                  },
                ),
                const SizedBox(height: 12),
                _buildDialogOption(
                  context,
                  Icons.output_rounded,
                  'Pengembalian Barang Bawaan ke Gudang',
                  'Pengembalian barang bawaan sales ke gudang',
                  () async {
                    Navigator.pop(context);
                    final result = await Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) =>
                            const SalesStockRequestPage(isReturn: true),
                      ),
                    );
                    if (result == true && mounted) {
                      _fetchData(showSkeleton: false);
                    }
                  },
                ),
                const SizedBox(height: 12),
                _buildDialogOption(
                  context,
                  Icons.edit_note_rounded,
                  'Penyesuaian Stok Manual',
                  'Edit jumlah stok bawaan jika ada pengurangan luar sistem',
                  () async {
                    Navigator.pop(context);
                    final result = await Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => const ManualStockAdjustmentPage(),
                      ),
                    );
                    if (result == true && mounted) {
                      _fetchData(showSkeleton: false);
                    }
                  },
                ),
                const SizedBox(height: 12),
                _buildDialogOption(
                  context,
                  Icons.history_rounded,
                  'Riwayat Ambil Barang',
                  'Pantau status request & retur barang',
                  () {
                    Navigator.pop(context);
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => const SalesStockHistoryPage(),
                      ),
                    );
                  },
                ),
                const SizedBox(height: 24),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildDialogOption(
    BuildContext context,
    IconData icon,
    String title,
    String desc,
    VoidCallback onTap,
  ) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: AppColors.primaryGreen.withOpacity(0.05),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: AppColors.primaryGreen.withOpacity(0.1)),
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: AppColors.primaryGreen.withOpacity(0.1),
                shape: BoxShape.circle,
              ),
              child: Icon(icon, color: AppColors.primaryGreen, size: 24),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w800,
                      fontSize: 15,
                    ),
                  ),
                  Text(
                    desc,
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      color: Colors.blueGrey,
                    ),
                  ),
                ],
              ),
            ),
            const Icon(Icons.chevron_right_rounded, color: Colors.blueGrey),
          ],
        ),
      ),
    );
  }

  void _showMerchandisingOptions(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
      ),
      builder: (context) => SafeArea(
        child: SingleChildScrollView(
          child: Container(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  'Laporan Merchandising',
                  style: GoogleFonts.outfit(
                    fontSize: 18,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 24),
                _buildDialogOption(
                  context,
                  Icons.add_a_photo_rounded,
                  'Buat Laporan Baru',
                  'Catat data planogram, display foto, dll.',
                  () {
                    Navigator.pop(context);
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => const MerchandisingPage(),
                      ),
                    );
                  },
                ),
                const SizedBox(height: 12),
                _buildDialogOption(
                  context,
                  Icons.history_edu_rounded,
                  'Riwayat Laporan',
                  'Lihat daftar laporan merchandising yang sudah dibuat',
                  () {
                    Navigator.pop(context);
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => const MerchandisingHistoryPage(),
                      ),
                    );
                  },
                ),
                const SizedBox(height: 24),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildActionItem(
    IconData icon,
    String label,
    Color color,
    String subLabel,
  ) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: color.withOpacity(0.08),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () {
            if (label == 'Order Instan') {
              Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const InstantOrderPage()),
              );
            } else if (label == 'Kunjungan') {
              Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const VisitSchedulePage()),
              );
            } else if (label == 'Kelola Rencana') {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => const VisitPlanManagementPage(),
                ),
              );
            } else if (label == 'Toko' || label == 'Mitra') {
              _showStoreManagementOptions(context);
            } else if (label == 'Katalog') {
              Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const ProductCatalogPage()),
              );
            } else if (label == 'Riwayat') {
              Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const SalesHistoryPage()),
              );
            } else if (label == 'Riwayat KM') {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => const SalesJourneyHistoryPage(),
                ),
              );
            } else if (label == 'Stok' || label == 'Ambil Barang') {
              _showStockManagementOptions(context);
            } else if (label == 'Reimbursement') {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => const ReimbursementListPage(),
                ),
              );
            } else if (label == 'PO MT') {
              Navigator.of(
                context,
              ).push(MaterialPageRoute(builder: (_) => const MtPoListPage()));
            } else if (label == 'Merchandising') {
              _showMerchandisingOptions(context);
            } else if (label == 'Pusat') {
              // Info dialog tentang MT invoice management
              showDialog(
                context: context,
                builder: (ctx) => AlertDialog(
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20),
                  ),
                  title: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: const BoxDecoration(
                          color: Color(0xFFECFDF5),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(
                          Icons.account_balance_rounded,
                          color: Color(0xFF10B981),
                          size: 20,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Text(
                        'Faktur MT',
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w900,
                          fontSize: 16,
                        ),
                      ),
                    ],
                  ),
                  content: Text(
                    'Anda sedang beroperasi dalam mode Sales MT (Modern Trade).\n\n'
                    'Setiap pesanan yang Anda buat akan:\n'
                    '• Dikirim sebagai Sales Order ke sistem pusat\n'
                    '• Diproses dan dikelola oleh Admin Nota pusat\n'
                    '• Omzet Anda tercatat di laporan pusat\n\n'
                    'Proses ini berjalan otomatis saat Anda sync data.',
                    style: GoogleFonts.outfit(fontSize: 13, height: 1.5),
                  ),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.pop(ctx),
                      child: Text(
                        'Mengerti',
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w800,
                          color: const Color(0xFF10B981),
                        ),
                      ),
                    ),
                  ],
                ),
              );
            }
          },
          borderRadius: BorderRadius.circular(24),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 6),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        color.withOpacity(0.15),
                        color.withOpacity(0.05),
                      ],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(icon, color: color, size: 20),
                ),
                const SizedBox(height: 6),
                FittedBox(
                  fit: BoxFit.scaleDown,
                  child: Text(
                    label,
                    textAlign: TextAlign.center,
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      color: const Color(0xFF1E293B),
                    ),
                  ),
                ),
                const SizedBox(height: 2),
                FittedBox(
                  fit: BoxFit.scaleDown,
                  child: Text(
                    subLabel,
                    textAlign: TextAlign.center,
                    style: GoogleFonts.outfit(
                      fontSize: 9,
                      fontWeight: FontWeight.w500,
                      color: const Color(0xFF94A3B8),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildOutstandingTempoList() {
    final formatter = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Row(
              children: [
                const Icon(
                  Icons.warning_amber_rounded,
                  color: Colors.redAccent,
                  size: 16,
                ),
                const SizedBox(width: 6),
                Text(
                  'PIUTANG / TEMPO JATUH TEMPO',
                  style: GoogleFonts.outfit(
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                    color: const Color(0xFFE11D48),
                    letterSpacing: 1.5,
                  ),
                ),
              ],
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: Colors.red.shade50,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                '${_outstandingTempoTransactions.length} INVOICE',
                style: GoogleFonts.outfit(
                  fontSize: 9,
                  fontWeight: FontWeight.w900,
                  color: Colors.red.shade800,
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        ListView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: _outstandingTempoTransactions.length,
          itemBuilder: (context, index) {
            final txn = _outstandingTempoTransactions[index];
            final storeName = txn['store']?['name'] ?? 'Toko';
            final storeAddress =
                txn['store']?['address'] ?? 'Alamat tidak tersedia';
            final receiptNo = txn['receipt_no'] ?? '';
            final totalAmount =
                (txn['total_amount'] as num?)?.toDouble() ?? 0.0;
            final paidAmount = (txn['paid_amount'] as num?)?.toDouble() ?? 0.0;
            final outstandingAmount = totalAmount - paidAmount;
            final dueDateStr = txn['payment_due_date'] != null
                ? DateFormat('dd MMM yyyy').format(
                    DateTime.parse(
                      txn['payment_due_date'].toString(),
                    ).toLocal(),
                  )
                : '-';

            return GestureDetector(
              onTap: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => DigitalReceiptPage(
                      transaction: txn,
                      localId: txn['local_id'],
                    ),
                  ),
                );
              },
              child: Container(
                margin: const EdgeInsets.only(bottom: 10),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: Colors.red.shade100, width: 1.5),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.red.shade900.withOpacity(0.04),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: Colors.red.shade50,
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.receipt_long_rounded,
                        color: Colors.redAccent,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
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
                                      storeName,
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w900,
                                        fontSize: 13,
                                        color: const Color(0xFF1E293B),
                                      ),
                                    ),
                                    const SizedBox(height: 2),
                                    Row(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Expanded(
                                          child: Text(
                                            storeAddress,
                                            maxLines: 2,
                                            overflow: TextOverflow.ellipsis,
                                            style: GoogleFonts.outfit(
                                              fontWeight: FontWeight.w500,
                                              fontSize: 10,
                                              color: Colors.blueGrey,
                                            ),
                                          ),
                                        ),
                                        const SizedBox(width: 6),
                                        IconButton(
                                          padding: EdgeInsets.zero,
                                          constraints: const BoxConstraints(
                                            minWidth: 24,
                                            minHeight: 24,
                                          ),
                                          onPressed: () async {
                                            final uri = Uri.parse(
                                              'https://www.google.com/maps/search/?api=1&query=${Uri.encodeComponent(storeAddress)}',
                                            );
                                            if (await canLaunchUrl(uri)) {
                                              await launchUrl(
                                                uri,
                                                mode: LaunchMode
                                                    .externalApplication,
                                              );
                                            } else {
                                              if (context.mounted) {
                                                ScaffoldMessenger.of(
                                                  context,
                                                ).showSnackBar(
                                                  const SnackBar(
                                                    content: Text(
                                                      'Gagal membuka Maps',
                                                    ),
                                                  ),
                                                );
                                              }
                                            }
                                          },
                                          icon: Container(
                                            padding: const EdgeInsets.all(4),
                                            decoration: BoxDecoration(
                                              color: Colors.blue.shade50,
                                              shape: BoxShape.circle,
                                            ),
                                            child: Icon(
                                              Icons.location_on_rounded,
                                              size: 14,
                                              color: Colors.blue.shade700,
                                            ),
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
                                  Text(
                                    'Jatuh Tempo:',
                                    style: GoogleFonts.outfit(
                                      fontSize: 8,
                                      color: Colors.blueGrey,
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                  Text(
                                    dueDateStr,
                                    style: GoogleFonts.outfit(
                                      fontSize: 10,
                                      color: Colors.red.shade700,
                                      fontWeight: FontWeight.w800,
                                    ),
                                  ),
                                  const SizedBox(height: 6),
                                  Text(
                                    'Sisa Pembayaran:',
                                    style: GoogleFonts.outfit(
                                      fontSize: 8,
                                      color: Colors.red.shade800,
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                  Text(
                                    formatter.format(outstandingAmount),
                                    style: GoogleFonts.outfit(
                                      fontSize: 12,
                                      color: Colors.red.shade700,
                                      fontWeight: FontWeight.w900,
                                    ),
                                  ),
                                ],
                              ),
                            ],
                          ),
                          const SizedBox(height: 4),
                          Text(
                            receiptNo,
                            style: GoogleFonts.outfit(
                              fontSize: 10,
                              color: Colors.blueGrey.shade400,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          const SizedBox(height: 6),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      children: [
                                        Text(
                                          'Tagihan: ',
                                          style: GoogleFonts.outfit(
                                            fontSize: 9,
                                            color: Colors.blueGrey,
                                            fontWeight: FontWeight.w500,
                                          ),
                                        ),
                                        Expanded(
                                          child: Text(
                                            formatter.format(totalAmount),
                                            style: GoogleFonts.outfit(
                                              fontSize: 9,
                                              color: const Color(0xFF1E293B),
                                              fontWeight: FontWeight.w700,
                                            ),
                                            overflow: TextOverflow.ellipsis,
                                          ),
                                        ),
                                      ],
                                    ),
                                    if (paidAmount > 0)
                                      Row(
                                        children: [
                                          Text(
                                            'Telah Dibayar (DP): ',
                                            style: GoogleFonts.outfit(
                                              fontSize: 9,
                                              color: Colors.blueGrey,
                                              fontWeight: FontWeight.w500,
                                            ),
                                          ),
                                          Expanded(
                                            child: Text(
                                              formatter.format(paidAmount),
                                              style: GoogleFonts.outfit(
                                                fontSize: 9,
                                                color: Colors.green.shade700,
                                                fontWeight: FontWeight.w700,
                                              ),
                                              overflow: TextOverflow.ellipsis,
                                            ),
                                          ),
                                        ],
                                      ),
                                  ],
                                ),
                              ),
                              const SizedBox(width: 8),
                              ElevatedButton.icon(
                                onPressed: () {
                                  Navigator.push(
                                    context,
                                    MaterialPageRoute(
                                      builder: (_) => TempoPaymentPage(
                                        transactionData: txn,
                                        onSuccess: (method, amount, signatureUrl) {
                                          Navigator.pop(
                                            context,
                                          ); // Close payment page
                                          ScaffoldMessenger.of(
                                            context,
                                          ).showSnackBar(
                                            const SnackBar(
                                              content: Text(
                                                'Pembayaran berhasil dicatat!',
                                              ),
                                              backgroundColor: Colors.green,
                                            ),
                                          );
                                          // Refresh the dashboard data
                                          _fetchData();
                                        },
                                      ),
                                    ),
                                  );
                                },
                                icon: const Icon(
                                  Icons.payments_rounded,
                                  size: 14,
                                  color: Colors.white,
                                ),
                                label: Text(
                                  'TAGIH SEKARANG',
                                  style: GoogleFonts.outfit(
                                    fontSize: 10,
                                    color: Colors.white,
                                    fontWeight: FontWeight.w900,
                                  ),
                                ),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: const Color(0xFFE11D48),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 14,
                                    vertical: 8,
                                  ),
                                  elevation: 4,
                                  shadowColor: Colors.red.shade200,
                                  minimumSize: const Size(0, 36),
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
            );
          },
        ),
      ],
    );
  }

  Widget _buildPendingTransactionsList() {
    final formatter = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    final int localCount = _pendingTransactions
        .where((t) => t['is_local'] == true && t['status'] == 'PENDING')
        .length;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'ANTRIAN & STATUS TRANSAKSI',
              style: GoogleFonts.outfit(
                fontSize: 11,
                fontWeight: FontWeight.w900,
                color: const Color(0xFF64748B),
                letterSpacing: 1.5,
              ),
            ),
            if (localCount > 0)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.orange.shade100,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  '$localCount OFFLINE',
                  style: GoogleFonts.outfit(
                    fontSize: 9,
                    fontWeight: FontWeight.w900,
                    color: Colors.orange.shade900,
                  ),
                ),
              )
            else
              Row(
                children: [
                  const Icon(
                    Icons.check_circle_rounded,
                    color: Color(0xFF10B981),
                    size: 12,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    'Terupload Di Server',
                    style: GoogleFonts.outfit(
                      fontSize: 9,
                      fontWeight: FontWeight.w900,
                      color: const Color(0xFF10B981),
                    ),
                  ),
                ],
              ),
          ],
        ),
        const SizedBox(height: 12),
        ListView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: _pendingTransactions.length > 3
              ? 3
              : _pendingTransactions.length,
          itemBuilder: (context, index) {
            final txn = _pendingTransactions[index];
            final String status = txn['status'] ?? 'PENDING';
            final String method = (txn['payment_method'] ?? '')
                .toString()
                .toUpperCase();
            final bool isMidtrans = method.contains('MIDTRANS');
            final bool isLocalPending =
                txn['is_local'] == true && status == 'PENDING';
            final bool isSynced =
                status == 'SYNCED' ||
                status == 'SUCCESS' ||
                txn['is_local'] == false;

            // Choose color theme
            Color cardBorderColor = const Color(0xFFE2E8F0);
            Color iconBgColor = const Color(0xFFF1F5F9);
            Color iconColor = const Color(0xFF64748B);
            IconData iconData = Icons.cloud_done_rounded;
            Color tagBgColor = const Color(0xFFF1F5F9);
            Color tagTextColor = const Color(0xFF64748B);
            String tagText = 'DI SERVER';

            if (isLocalPending) {
              cardBorderColor = Colors.orange.shade100;
              iconBgColor = Colors.orange.shade50;
              iconColor = Colors.orange.shade400;
              iconData = Icons.cloud_off_rounded;
              tagBgColor = Colors.orange.shade100;
              tagTextColor = Colors.orange.shade900;
              tagText = 'OFFLINE';
            } else if (isSynced) {
              if (isMidtrans) {
                cardBorderColor = Colors.blue.shade100;
                iconBgColor = Colors.blue.shade50;
                iconColor = Colors.blueAccent;
                iconData = Icons.account_balance_wallet_rounded;
                tagBgColor = Colors.blue.shade100;
                tagTextColor = Colors.blue.shade900;
                tagText = 'LUNAS (MIDTRANS)';
              } else if (method == 'VA' || method == 'TRANSFER') {
                cardBorderColor = Colors.indigo.shade100;
                iconBgColor = Colors.indigo.shade50;
                iconColor = Colors.indigo;
                iconData = Icons.account_balance_rounded;
                tagBgColor = Colors.indigo.shade100;
                tagTextColor = Colors.indigo.shade900;
                tagText = 'LUNAS (TRANSFER)';
              } else if (method == 'SALES_ORDER') {
                cardBorderColor = Colors.blue.shade100;
                iconBgColor = Colors.blue.shade50;
                iconColor = Colors.blue;
                iconData = Icons.description_rounded;
                tagBgColor = Colors.blue.shade100;
                tagTextColor = Colors.blue.shade900;
                tagText = 'BELUM LUNAS (SO)';
              } else {
                cardBorderColor = Colors.green.shade100;
                iconBgColor = Colors.green.shade50;
                iconColor = Colors.green;
                iconData = Icons.check_circle_rounded;
                tagBgColor = Colors.green.shade100;
                tagTextColor = Colors.green.shade900;
                tagText = 'MENUNGGU KASIR';
              }
            }

            return GestureDetector(
              onTap: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => DigitalReceiptPage(
                      transaction: txn,
                      localId: txn['local_id'],
                    ),
                  ),
                );
              },
              child: Container(
                margin: const EdgeInsets.only(bottom: 10),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: cardBorderColor),
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: iconBgColor,
                        shape: BoxShape.circle,
                      ),
                      child: Icon(iconData, color: iconColor, size: 20),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(
                                  txn['store']?['name'] ?? 'Toko',
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w800,
                                    fontSize: 14,
                                  ),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                              const SizedBox(width: 8),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 6,
                                  vertical: 2,
                                ),
                                decoration: BoxDecoration(
                                  color: tagBgColor,
                                  borderRadius: BorderRadius.circular(6),
                                ),
                                child: Text(
                                  tagText,
                                  style: GoogleFonts.outfit(
                                    fontSize: 8,
                                    fontWeight: FontWeight.w900,
                                    color: tagTextColor,
                                  ),
                                ),
                              ),
                            ],
                          ),
                          Text(
                            formatter.format(txn['total_amount'] ?? 0),
                            style: GoogleFonts.outfit(
                              fontSize: 12,
                              color: Colors.blueGrey,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          if (isLocalPending &&
                              txn['last_error'] != null &&
                              (txn['last_error'] as String).isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(top: 4),
                              child: Row(
                                children: [
                                  const Icon(
                                    Icons.error_outline_rounded,
                                    size: 12,
                                    color: Color(0xFFE11D48),
                                  ),
                                  const SizedBox(width: 4),
                                  Expanded(
                                    child: Text(
                                      'Gagal kirim: ${txn['last_error']}',
                                      style: GoogleFonts.outfit(
                                        fontSize: 10,
                                        color: const Color(0xFFE11D48),
                                        fontWeight: FontWeight.w600,
                                      ),
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ),
                                ],
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
        ),
        if (_pendingTransactions.length > 3)
          Center(
            child: TextButton(
              onPressed: () => Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const SalesHistoryPage()),
              ),
              child: Text(
                'LIHAT SEMUA',
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  color: Colors.blueGrey,
                ),
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildSalesDrawer() {
    return Drawer(
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.only(
          topRight: Radius.circular(32),
          bottomRight: Radius.circular(32),
        ),
      ),
      child: Column(
        children: [
          Container(
            width: double.infinity,
            padding: const EdgeInsets.fromLTRB(24, 80, 24, 40),
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                colors: [AppColors.blueSalesDark, AppColors.blueSales],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
            ),
            child: BlocBuilder<ProfileBloc, ProfileState>(
              builder: (context, state) {
                String name = 'Sales Team';
                String company = 'Perusahaan Demo';
                if (state is ProfileLoaded) {
                  name = state.profile['name'] ?? 'Sales Team';
                  company =
                      state.profile['company']?['name'] ?? 'Perusahaan Demo';
                }
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.1),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.person_pin_rounded,
                        color: Colors.white,
                        size: 32,
                      ),
                    ),
                    const SizedBox(height: 20),
                    Text(
                      'SALES OPERATION',
                      style: GoogleFonts.outfit(
                        color: AppColors.blueSalesLight,
                        fontWeight: FontWeight.w900,
                        fontSize: 11,
                        letterSpacing: 2,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      name,
                      style: GoogleFonts.outfit(
                        color: Colors.white,
                        fontWeight: FontWeight.w800,
                        fontSize: 20,
                      ),
                    ),
                    Text(
                      company,
                      style: GoogleFonts.outfit(
                        color: Colors.white.withOpacity(0.7),
                        fontWeight: FontWeight.w600,
                        fontSize: 13,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ],
                );
              },
            ),
          ),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.symmetric(vertical: 20),
              children: [
                _buildDrawerItem(Icons.storefront_rounded, 'Daftar Toko', () {
                  Navigator.pop(context); // Close drawer
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (_) => const StoreListPage()),
                  );
                }),
                _buildDrawerItem(Icons.map_rounded, 'Rute Optimal', () {
                  Navigator.pop(context);
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => const RoutePlanningPage(),
                    ),
                  );
                }),
                Builder(
                  builder: (context) {
                    Map<String, dynamic>? drawerProfile;
                    final drawerProfileState = context
                        .read<ProfileBloc>()
                        .state;
                    if (drawerProfileState is ProfileLoaded) {
                      drawerProfile = drawerProfileState.profile;
                    } else {
                      final authState = context.read<AuthBloc>().state;
                      if (authState is Authenticated) {
                        drawerProfile = authState.userProfile;
                      }
                    }
                    final String drawerDistStoreId =
                        drawerProfile?['distributor_store_id']?.toString() ??
                        '';
                    final bool isDistributorDrawer =
                        drawerProfile?['is_distributor'] == true ||
                        drawerDistStoreId.isNotEmpty ||
                        (drawerProfile?['job_position']?['title'] ?? '')
                            .toString()
                            .toLowerCase()
                            .contains('distributor');

                    if (isDistributorDrawer) return const SizedBox.shrink();

                    return _buildDrawerItem(
                      Icons.menu_book_rounded,
                      'Katalog Produk Kecap',
                      () {
                        Navigator.pop(context);
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (_) => const ProductCatalogPage(),
                          ),
                        );
                      },
                    );
                  },
                ),
                _buildDrawerItem(
                  Icons.backup_rounded,
                  'Cadangan Data Offline',
                  _exportOfflineBackup,
                ),
                const Padding(
                  padding: EdgeInsets.symmetric(horizontal: 24, vertical: 20),
                  child: Divider(color: Color(0xFFF1F5F9), thickness: 2),
                ),
                _buildDrawerItem(
                  Icons.support_agent_rounded,
                  'Bantuan Lapangan',
                  () {},
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _exportOfflineBackup({bool closeDrawer = true}) async {
    if (closeDrawer)
      Navigator.pop(context); // Close drawer when opened from menu
    try {
      final db = di.sl<AppDatabase>();
      final pending = await (db.select(
        db.localTransactions,
      )..where((t) => t.syncStatus.equals('pending'))).get();

      if (pending.isEmpty) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              '✓ Tidak ada transaksi offline pending yang perlu dicadangkan.',
            ),
            backgroundColor: Color(0xFF10B981),
          ),
        );
        return;
      }

      final List<Map<String, dynamic>> exportedTxns = [];
      for (var tr in pending) {
        final items = await (db.select(
          db.localTransactionItems,
        )..where((t) => t.transactionLocalId.equals(tr.localId))).get();

        exportedTxns.add({
          'client_id': tr.localId,
          'company_id': tr.companyId,
          'store_id': tr.storeId,
          'latitude': tr.latitude,
          'longitude': tr.longitude,
          'total_amount': tr.totalAmount,
          'paid_amount': tr.paidAmount,
          'payment_method': tr.paymentMethod,
          'dp_payment_method': tr.dpPaymentMethod,
          'payment_due_date': tr.paymentDueDate != null
              ? '${tr.paymentDueDate!.toIso8601String().substring(0, 10)}T00:00:00.000Z'
              : null,
          'tenor': tr.tenor,
          'bank': tr.paymentBank,
          'notes': tr.notes,
          'store_category': 'TOKO_LAMA',
          'transaction_date': tr.createdAt.toUtc().toIso8601String(),
          'receipt_no': tr.receiptNo,
          'receipt_image_url': tr.receiptPath,
          'signature_image_url': tr.signaturePath,
          'is_mt': tr.isMt,
          'factory_company_id': tr.factoryCompanyId,
          'company_bank_account_id': tr.companyBankAccountId,
          'visit_id': tr.visitId,
          'items': items
              .map(
                (i) => {
                  'product_id': i.productId,
                  'quantity': i.quantity,
                  'ordered_quantity': i.orderedQuantity,
                  'unit': i.unit,
                  'pieces_per_unit': i.piecesPerUnit,
                  'price': i.price,
                  'discount_percent': i.discountPercent,
                },
              )
              .toList(),
        });
      }

      final backupPayload = jsonEncode({
        'exported_at': DateTime.now().toIso8601String(),
        'transactions': exportedTxns,
      });

      if (!mounted) return;
      showDialog(
        context: context,
        builder: (ctx) => AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          title: Row(
            children: [
              const Icon(Icons.backup_rounded, color: AppColors.blueSalesDark),
              const SizedBox(width: 8),
              Text(
                'Cadangan Offline',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800,
                  fontSize: 16,
                ),
              ),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Ditemukan ${exportedTxns.length} transaksi offline pending.',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w700,
                  fontSize: 14,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Anda dapat membagikan file cadangan (.json) ini ke WhatsApp Admin/Supervisor atau menyalin teksnya langsung untuk di-import di Web HRIS jika terkendala internet.',
                style: GoogleFonts.outfit(
                  fontSize: 12,
                  color: Colors.grey[700],
                ),
              ),
            ],
          ),
          actionsAlignment: MainAxisAlignment.spaceBetween,
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: Text(
                'TUTUP',
                style: GoogleFonts.outfit(color: Colors.grey),
              ),
            ),
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                OutlinedButton.icon(
                  icon: const Icon(Icons.copy_rounded, size: 15),
                  label: Text(
                    'SALIN TEKS',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w700,
                      fontSize: 11,
                    ),
                  ),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: AppColors.blueSalesDark,
                    side: const BorderSide(color: AppColors.blueSalesDark),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 8,
                    ),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                  onPressed: () {
                    Clipboard.setData(ClipboardData(text: backupPayload));
                    Navigator.pop(ctx);
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(
                        content: Text(
                          '✓ Data cadangan berhasil disalin ke clipboard!',
                        ),
                        backgroundColor: Color(0xFF10B981),
                      ),
                    );
                  },
                ),
                const SizedBox(width: 8),
                ElevatedButton.icon(
                  icon: const Icon(Icons.share_rounded, size: 15),
                  label: Text(
                    'BAGIKAN FILE',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w800,
                      fontSize: 11,
                    ),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.blueSalesDark,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 8,
                    ),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                  onPressed: () async {
                    Navigator.pop(ctx);
                    try {
                      final tempDir = await getTemporaryDirectory();
                      final timestamp = DateFormat(
                        'yyyyMMdd_HHmmss',
                      ).format(DateTime.now());
                      final fileName = 'cadangan_offline_$timestamp.json';
                      final file = File('${tempDir.path}/$fileName');
                      await file.writeAsString(backupPayload);
                      await Share.shareXFiles(
                        [
                          XFile(
                            file.path,
                            mimeType: 'application/json',
                            name: fileName,
                          ),
                        ],
                        text:
                            'Cadangan Transaksi Offline HRIS Karsa HR ($timestamp)',
                      );
                    } catch (err) {
                      if (!mounted) return;
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text('Gagal membagikan file: $err'),
                          backgroundColor: Colors.red,
                        ),
                      );
                    }
                  },
                ),
              ],
            ),
          ],
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Gagal mengekspor data: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  Widget _buildDrawerItem(IconData icon, String title, VoidCallback onTap) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: const Color(0xFFF1F5F9),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Icon(icon, color: const Color(0xFF64748B), size: 20),
      ),
      title: Text(
        title,
        style: GoogleFonts.outfit(
          color: const Color(0xFF334155),
          fontWeight: FontWeight.w700,
          fontSize: 14,
        ),
      ),
      onTap: onTap,
      contentPadding: const EdgeInsets.symmetric(horizontal: 24, vertical: 4),
    );
  }
}
