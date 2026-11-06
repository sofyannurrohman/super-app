import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:drift/drift.dart' as drift;
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

class SalesApiService {
  final ApiClient apiClient;

  SalesApiService({required this.apiClient});

  AppDatabase get _db => di.sl<AppDatabase>();
  SharedPreferences get _prefs => di.sl<SharedPreferences>();

  static String performanceCacheKey(DateTime date) =>
      'cached_sales_kpi_${DateFormat('yyyy-MM-dd').format(date)}';

  /// Sends the latest background location. The backend decides whether the
  /// employee is currently allowed to be tracked.
  Future<bool> sendLiveLocation({
    required double latitude,
    required double longitude,
    double? accuracy,
    DateTime? recordedAt,
  }) async {
    try {
      final response = await apiClient.client.post(
        'sales/live-location',
        data: {
          'latitude': latitude,
          'longitude': longitude,
          if (accuracy != null) 'accuracy': accuracy,
          'recorded_at': (recordedAt ?? DateTime.now().toUtc()).toIso8601String(),
        },
      );
      return response.statusCode != null &&
          response.statusCode! >= 200 &&
          response.statusCode! < 300;
    } on DioException catch (e) {
      // A 403 means checkout/auto-stop; callers should stop the tracker.
      if (e.response?.statusCode == 403) return false;
      if (_isNetworkOrOfflineError(e)) return false;
      rethrow;
    }
  }

  bool _isNetworkOrOfflineError(dynamic e) {
    if (e is DioException) {
      if (e.type == DioExceptionType.connectionTimeout ||
          e.type == DioExceptionType.sendTimeout ||
          e.type == DioExceptionType.receiveTimeout ||
          e.type == DioExceptionType.connectionError ||
          e.type == DioExceptionType.unknown ||
          e.response == null) {
        return true;
      }
      final errMsg = '${e.message ?? ''} ${e.error ?? ''}'.toLowerCase();
      if (errMsg.contains('socketexception') ||
          errMsg.contains('failed host lookup') ||
          errMsg.contains('connection errored') ||
          errMsg.contains('connection refused') ||
          errMsg.contains('network') ||
          errMsg.contains('timeout')) {
        return true;
      }
    }
    final s = e.toString().toLowerCase();
    return s.contains('socketexception') ||
        s.contains('network') ||
        s.contains('connection refused') ||
        s.contains('failed host lookup') ||
        s.contains('host lookup') ||
        s.contains('connection errored') ||
        s.contains('clientexception') ||
        s.contains('timeout') ||
        s.contains('xmlhttprequest error');
  }

  /// Pastikan tabel SQLite untuk alur offline sales selalu tersedia (self-healing untuk DB versi lama)
  Future<void> _ensureOfflineSalesTables() async {
    try {
      await _db.customStatement('''
        CREATE TABLE IF NOT EXISTS visit_plans (
          id TEXT NOT NULL PRIMARY KEY,
          store_id TEXT NOT NULL,
          store_name TEXT NOT NULL,
          date TEXT NOT NULL,
          status TEXT NOT NULL DEFAULT 'pending'
        )
      ''');
      await _db.customStatement('''
        CREATE TABLE IF NOT EXISTS offline_tasks (
          id TEXT NOT NULL PRIMARY KEY,
          type TEXT NOT NULL,
          payload TEXT NOT NULL,
          file_path TEXT,
          created_at INTEGER NOT NULL,
          status TEXT NOT NULL DEFAULT 'pending',
          error_message TEXT
        )
      ''');
      await _db.customStatement('''
        CREATE TABLE IF NOT EXISTS sales_daily_journeys (
          id TEXT NOT NULL PRIMARY KEY,
          user_id TEXT NOT NULL,
          date TEXT NOT NULL,
          start_km INTEGER NOT NULL,
          start_photo_url TEXT NOT NULL,
          start_location TEXT NOT NULL,
          start_time INTEGER NOT NULL,
          end_km INTEGER,
          end_photo_url TEXT,
          end_location TEXT,
          end_time INTEGER,
          total_km INTEGER,
          status TEXT NOT NULL DEFAULT 'STARTED',
          is_synced INTEGER NOT NULL DEFAULT 0
        )
      ''');
    } catch (_) {}
  }

  /// Antrekan task offline ke SQLite `offline_tasks` dengan fallback `SharedPreferences`
  Future<void> _enqueueOfflineTask({
    required String type,
    required Map<String, dynamic> payload,
  }) async {
    final taskId = const Uuid().v4();
    final payloadJson = jsonEncode(payload);
    try {
      await _ensureOfflineSalesTables();
      await _db.into(_db.offlineTasks).insert(
            OfflineTasksCompanion.insert(
              id: taskId,
              type: type,
              payload: payloadJson,
              createdAt: DateTime.now(),
              status: const drift.Value('pending'),
            ),
          );
    } catch (_) {
      // Fallback ke SharedPreferences jika tabel SQLite sedang terkunci/bermasalah
      final rawQueue = _prefs.getStringList('offline_visit_plan_queue') ?? [];
      rawQueue.add(jsonEncode({
        'id': taskId,
        'type': type,
        'payload': payload,
        'created_at': DateTime.now().toIso8601String(),
      }));
      await _prefs.setStringList('offline_visit_plan_queue', rawQueue);
    }
  }

  String _getLoggedInUserId() {
    final cached = _prefs.getString('cached_employee_profile');
    if (cached != null) {
      try {
        final decoded = jsonDecode(cached);
        return (decoded['user_id'] ?? decoded['id'] ?? 'offline_user').toString();
      } catch (_) {}
    }
    return 'offline_user';
  }

  // GET /api/v1/admin/sales/director-summary
  Future<Map<String, dynamic>> getDirectorSummary({String? date}) async {
    try {
      final queryParams = <String, dynamic>{};
      if (date != null && date.isNotEmpty) {
        queryParams['date'] = date;
      }
      final response = await apiClient.client.get(
        'admin/sales/director-summary',
        queryParameters: queryParams,
      );
      
      if (response.data is String) {
        return {}; 
      }
      if (response.data is Map) {
        final payload = Map<String, dynamic>.from(response.data as Map);
        final nested = payload['data'];
        if (nested is Map) return Map<String, dynamic>.from(nested);
        return payload;
      }
      return {};
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        return {};
      }
      throw Exception('Failed to get director summary: $e');
    }
  }

  // GET /api/v1/performance/my-performance
  Future<Map<String, dynamic>> getMyPerformance() async {
    final cacheKey = performanceCacheKey(DateTime.now());
    try {
      final response = await apiClient.client.get(
        'performance/my-performance',
        queryParameters: {'_t': DateTime.now().millisecondsSinceEpoch},
      );
      final data = (response.data['data'] as Map<String, dynamic>?) ?? {};
      if (cacheKey != performanceCacheKey(DateTime.now())) return {};
      if (data.isNotEmpty) {
        await _prefs.setString(cacheKey, jsonEncode(data));
      }
      return data;
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        final cached = _prefs.getString(cacheKey);
        if (cached != null && cached.isNotEmpty) {
          try {
            return Map<String, dynamic>.from(jsonDecode(cached) as Map);
          } catch (_) {}
        }
        return {};
      }
      throw Exception('Failed to load performance data: $e');
    }
  }

  // GET /api/v1/sales/transactions/pending
  Future<List<dynamic>> getPendingTransactions() async {
    try {
      final response = await apiClient.client.get(
        'sales/transactions/pending',
        queryParameters: {'_t': DateTime.now().millisecondsSinceEpoch},
      );
      return (response.data['data'] as List<dynamic>?) ?? [];
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        return [];
      }
      throw Exception('Failed to load pending transactions: $e');
    }
  }

  // GET /api/v1/sales/transactions/history
  Future<List<dynamic>> getHistoryTransactions() async {
    try {
      final response = await apiClient.client.get(
        'sales/transactions/history',
        queryParameters: {'_t': DateTime.now().millisecondsSinceEpoch},
      );
      return (response.data['data'] as List<dynamic>?) ?? [];
    } catch (e) {
      throw Exception('Failed to load history transactions: $e');
    }
  }

  // Hapus transaksi server milik sales yang masih PENDING tanpa pembayaran.
  Future<void> deletePendingTransaction(String id) async {
    await apiClient.client.delete('sales/transactions/$id/pending');
  }

  // POST /api/v1/sales/transactions
  Future<Map<String, dynamic>> createTransaction(
      Map<String, dynamic> data) async {
    try {
      final response =
          await apiClient.client.post('sales/transactions', data: data);
      return (response.data['data'] as Map<String, dynamic>?) ?? {};
    } catch (e) {
      throw Exception('Failed to create transaction: $e');
    }
  }

  // PATCH /api/v1/sales/transactions/:id/verify
  Future<void> verifyTransaction(String id, Map<String, dynamic> data) async {
    try {
      await apiClient.client.patch('sales/transactions/$id/verify', data: data);
    } catch (e) {
      throw Exception('Failed to verify transaction: $e');
    }
  }

  /// Helper membaca rencana kunjungan dari cache & DB lokal untuk tanggal tertentu
  Future<List<Map<String, dynamic>>> _loadLocalVisitPlans(String dateStr) async {
    final deletedSet = (_prefs.getStringList('deleted_stores_$dateStr') ?? []).toSet();
    final Map<String, Map<String, dynamic>> planMap = {};

    // 1. Baca dari JSON cache tanggal tersebut (menyimpan metadata lengkap termasuk projections)
    final cachedStr = _prefs.getString('cached_visit_plans_$dateStr') ??
        _prefs.getString('cached_today_visit_plans_$dateStr');
    if (cachedStr != null && cachedStr.isNotEmpty) {
      try {
        final List<dynamic> decoded = jsonDecode(cachedStr);
        for (final item in decoded) {
          if (item is Map) {
            final mapItem = Map<String, dynamic>.from(item);
            final sId = (mapItem['store_id'] ?? '').toString();
            if (sId.isNotEmpty && !deletedSet.contains(sId)) {
              planMap[sId] = mapItem;
            }
          }
        }
      } catch (_) {}
    }

    // 2. Baca dari tabel SQLite db.visitPlans
    try {
      final rows = await (_db.select(_db.visitPlans)
            ..where((t) => t.date.equals(dateStr)))
          .get();
      if (rows.isNotEmpty) {
        final storeRows = await _db.select(_db.stores).get();
        final storeLookup = {for (var s in storeRows) s.id: s};

        for (final r in rows) {
          if (deletedSet.contains(r.storeId)) continue;
          final st = storeLookup[r.storeId];
          if (planMap.containsKey(r.storeId)) {
            // Update status jika di DB lokal sudah COMPLETED
            if (r.status.toUpperCase() == 'COMPLETED') {
              planMap[r.storeId]!['status'] = 'COMPLETED';
            }
            if ((planMap[r.storeId]!['latitude'] == null ||
                    planMap[r.storeId]!['latitude'] == 0.0) &&
                st != null) {
              planMap[r.storeId]!['latitude'] = st.latitude ?? 0.0;
              planMap[r.storeId]!['longitude'] = st.longitude ?? 0.0;
              planMap[r.storeId]!['address'] = st.address ?? '';
            }
          } else {
            planMap[r.storeId] = {
              'visit_id': r.id,
              'store_id': r.storeId,
              'store_name': r.storeName.isNotEmpty ? r.storeName : (st?.name ?? 'Toko'),
              'address': st?.address ?? '',
              'latitude': st?.latitude ?? 0.0,
              'longitude': st?.longitude ?? 0.0,
              'status': r.status.toUpperCase(),
              'is_extra': false,
              'projected_omzet': 0.0,
              'projections': <dynamic>[],
            };
          }
        }
      }
    } catch (_) {}

    return planMap.values.toList();
  }

  /// Helper menyimpan daftar rencana kunjungan ke cache & tabel lokal db.visitPlans
  Future<void> _persistVisitPlansLocally(
      String dateStr, List<dynamic> plans) async {
    try {
      final jsonStr = jsonEncode(plans);
      await _prefs.setString('cached_visit_plans_$dateStr', jsonStr);
      final todayStr = DateFormat('yyyy-MM-dd').format(DateTime.now());
      if (dateStr == todayStr) {
        await _prefs.setString('cached_today_visit_plans_$todayStr', jsonStr);
        await _prefs.setString('cached_today_visit_plans', jsonStr);
      }

      await _ensureOfflineSalesTables();
      await _db.batch((batch) {
        for (final p in plans) {
          if (p is Map) {
            final storeId = (p['store_id'] ?? '').toString();
            if (storeId.isEmpty) continue;
            final storeName = (p['store_name'] ?? 'Toko').toString();
            final status = (p['status'] ?? 'PLANNED').toString();
            final id = (p['visit_id'] ?? '${dateStr}_$storeId').toString();
            batch.insert(
              _db.visitPlans,
              VisitPlansCompanion.insert(
                id: id,
                storeId: storeId,
                storeName: storeName,
                date: dateStr,
                status: drift.Value(status),
              ),
              mode: drift.InsertMode.insertOrReplace,
            );
          }
        }
      });
    } catch (_) {}
  }

  // GET /api/v1/sales/visit-plans
  Future<List<dynamic>> getVisitPlans({String? date}) async {
    final String targetDate =
        date ?? DateFormat('yyyy-MM-dd').format(DateTime.now());
    final deletedSet =
        (_prefs.getStringList('deleted_stores_$targetDate') ?? []).toSet();
    final pendingLocalStoreIds =
        (_prefs.getStringList('offline_plan_pending_stores_$targetDate') ?? [])
            .toSet();

    try {
      final Map<String, dynamic> query = {'date': targetDate};
      query['_t'] = DateTime.now().millisecondsSinceEpoch.toString();

      final response = await apiClient.client
          .get('sales/visit-plans', queryParameters: query);
      final serverPlans = (response.data['data'] as List<dynamic>?) ?? [];

      // Gabungkan dengan plan lokal yang masih pending (belum ter-sync ke server)
      final localPlans = await _loadLocalVisitPlans(targetDate);
      final Map<String, Map<String, dynamic>> mergedMap = {};

      for (final sp in serverPlans) {
        if (sp is Map) {
          final mapItem = Map<String, dynamic>.from(sp);
          final sId = (mapItem['store_id'] ?? '').toString();
          if (sId.isNotEmpty && !deletedSet.contains(sId)) {
            mergedMap[sId] = mapItem;
          }
        }
      }

      for (final lp in localPlans) {
        final sId = (lp['store_id'] ?? '').toString();
        if (sId.isEmpty || deletedSet.contains(sId)) continue;
        if (!mergedMap.containsKey(sId) && pendingLocalStoreIds.contains(sId)) {
          // Plan dibuat saat offline dan belum naik ke server
          mergedMap[sId] = lp;
        } else if (mergedMap.containsKey(sId)) {
          // Pertahankan proyeksi lokal jika masih pending_save
          if (pendingLocalStoreIds.contains(sId) &&
              lp['projections'] is List &&
              (lp['projections'] as List).isNotEmpty) {
            mergedMap[sId]!['projections'] = lp['projections'];
            mergedMap[sId]!['projected_omzet'] = lp['projected_omzet'];
          }
          // Pertahankan status COMPLETED lokal
          if (lp['status'] == 'COMPLETED') {
            mergedMap[sId]!['status'] = 'COMPLETED';
          }
        }
      }

      final mergedList = mergedMap.values.toList();
      await _persistVisitPlansLocally(targetDate, mergedList);
      return mergedList;
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        return await _loadLocalVisitPlans(targetDate);
      }
      throw Exception('Failed to load visit plans: $e');
    }
  }

  // GET /api/v1/sales/visits/pending
  Future<List<dynamic>> getPendingVisits() async {
    try {
      final response = await apiClient.client.get('sales/visits/pending');
      return (response.data['data'] as List<dynamic>?) ?? [];
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        return [];
      }
      throw Exception('Failed to load pending visits: $e');
    }
  }

  Future<List<Map<String, dynamic>>> _getProductsFromLocalDb() async {
    try {
      final rows = await _db.select(_db.products).get();
      return rows
          .map((p) => {
                'id': p.id,
                'company_id': p.companyId,
                'name': p.name,
                'sku': p.sku ?? '',
                'selling_price': p.sellingPrice,
                'unit': p.unit ?? 'PCS',
                'category': p.category ?? '',
                'warehouse_stock': p.warehouseStock,
                'pcs_per_unit': p.pcsPerUnit,
                'image_url': p.imageUrl,
              })
          .toList();
    } catch (_) {
      return [];
    }
  }

  // GET /api/v1/factory/products
  Future<List<dynamic>> getProducts() async {
    try {
      final localProducts = await _getProductsFromLocalDb();
      if (localProducts.isNotEmpty) {
        // Refresh in background and return local products immediately for fast offline-first UX
        apiClient.client
            .get('factory/products', queryParameters: {'company_id': 'all'})
            .then((_) {})
            .catchError((_) {});
        return localProducts;
      }

      final response = await apiClient.client
          .get('factory/products', queryParameters: {'company_id': 'all'});
      if (response.data is Map && response.data.containsKey('data')) {
        return (response.data['data'] as List<dynamic>?) ?? [];
      }
      return (response.data as List<dynamic>?) ?? [];
    } catch (e) {
      final localProducts = await _getProductsFromLocalDb();
      if (localProducts.isNotEmpty || _isNetworkOrOfflineError(e)) {
        return localProducts;
      }
      throw Exception('Failed to load products: $e');
    }
  }

  // GET /api/v1/warehouse/stock
  Future<List<dynamic>> getBranchWarehouseStock() async {
    try {
      final response = await apiClient.client.get('warehouse/stock');
      if (response.data is Map && response.data.containsKey('data')) {
        return (response.data['data'] as List<dynamic>?) ?? [];
      }
      return (response.data as List<dynamic>?) ?? [];
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        return [];
      }
      throw Exception('Failed to load branch warehouse stock: $e');
    }
  }

  // POST /api/v1/sales-transfers/adjust
  Future<void> adjustSalesStock(
      String employeeId, List<Map<String, dynamic>> items) async {
    try {
      final response =
          await apiClient.client.post('sales-transfers/adjust', data: {
        'employee_id': employeeId,
        'items': items,
      });

      if (response.statusCode != 200 && response.statusCode != 201) {
        throw Exception('Failed to adjust sales stock');
      }
    } catch (e) {
      throw Exception('Failed to adjust sales stock: $e');
    }
  }

  // GET /api/v1/sales/transactions/due-date?date=all
  Future<List<dynamic>> getOutstandingTempoTransactions() async {
    try {
      final response = await apiClient.client
          .get('sales/transactions/due-date', queryParameters: {'date': 'all'});
      final data = (response.data['data'] as List<dynamic>?) ?? [];
      await _prefs.setString('cached_outstanding_tempo', jsonEncode(data));
      return data;
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        final cached = _prefs.getString('cached_outstanding_tempo');
        if (cached != null && cached.isNotEmpty) {
          try {
            return jsonDecode(cached) as List<dynamic>;
          } catch (_) {}
        }
        return [];
      }
      throw Exception('Failed to load outstanding tempo transactions: $e');
    }
  }

  // POST /api/v1/sales/transactions/:id/payments
  Future<void> recordPayment(
      String transactionId, Map<String, dynamic> data) async {
    try {
      await apiClient.client
          .post('sales/transactions/$transactionId/payments', data: data);
    } catch (e) {
      throw Exception('Failed to record payment: $e');
    }
  }

  // POST /api/v1/sales/visit-plans/manual
  Future<void> saveManualVisitPlan({
    List<String>? storeIds,
    List<Map<String, dynamic>>? stores,
    required DateTime planDate,
  }) async {
    final dateStr = DateFormat('yyyy-MM-dd').format(planDate);
    final Map<String, dynamic> data = {
      'plan_date': '${dateStr}T00:00:00Z',
    };
    if (storeIds != null) data['store_ids'] = storeIds;
    if (stores != null) data['stores'] = stores;

    // 1. Simpan secara lokal terlebih dahulu (Offline-First)
    try {
      final existingPlans = await _loadLocalVisitPlans(dateStr);
      final Map<String, Map<String, dynamic>> planByStore = {
        for (var p in existingPlans) (p['store_id'] ?? '').toString(): p
      };

      Map<String, Store> storeLookup = {};
      try {
        final dbStores = await _db.select(_db.stores).get();
        storeLookup = {for (var s in dbStores) s.id: s};
      } catch (_) {}

      // Fallback lookup dari cached_stores_list & cached_visit_plan_builder jika belum ada di db.stores
      final Map<String, Map<String, dynamic>> cachedStoreMap = {};
      for (final key in _prefs.getKeys()) {
        if (key.startsWith('cached_stores_list') ||
            key.startsWith('cached_visit_plan_builder')) {
          final raw = _prefs.getString(key);
          if (raw != null) {
            try {
              final list = jsonDecode(raw) as List<dynamic>;
              for (final item in list) {
                if (item is Map) {
                  final id = (item['id'] ?? item['store_id'] ?? '').toString();
                  if (id.isNotEmpty) {
                    cachedStoreMap[id] = Map<String, dynamic>.from(item);
                  }
                }
              }
            } catch (_) {}
          }
        }
      }

      Map<String, Product> productLookup = {};
      try {
        final dbProducts = await _db.select(_db.products).get();
        productLookup = {for (var p in dbProducts) p.id: p};
      } catch (_) {}

      final List<Map<String, dynamic>> normalizedStores = [];
      if (stores != null && stores.isNotEmpty) {
        normalizedStores.addAll(stores);
      } else if (storeIds != null && storeIds.isNotEmpty) {
        for (final id in storeIds) {
          normalizedStores.add({'store_id': id});
        }
      }

      final pendingStoreIds =
          (_prefs.getStringList('offline_plan_pending_stores_$dateStr') ?? [])
              .toSet();
      final deletedSet =
          (_prefs.getStringList('deleted_stores_$dateStr') ?? []).toSet();

      for (final entry in normalizedStores) {
        final sId = (entry['store_id'] ?? '').toString();
        if (sId.isEmpty) continue;
        deletedSet.remove(sId);
        pendingStoreIds.add(sId);

        final st = storeLookup[sId];
        final cachedSt = cachedStoreMap[sId];
        final existing = planByStore[sId];

        final storeName = st?.name ??
            (cachedSt?['name'] ??
                    cachedSt?['store_name'] ??
                    existing?['store_name'] ??
                    'Toko')
                .toString();
        final address = st?.address ??
            (cachedSt?['address'] ?? existing?['address'] ?? '').toString();
        final lat = st?.latitude ??
            (cachedSt?['latitude'] as num?)?.toDouble() ??
            (existing?['latitude'] as num?)?.toDouble() ??
            0.0;
        final lng = st?.longitude ??
            (cachedSt?['longitude'] as num?)?.toDouble() ??
            (existing?['longitude'] as num?)?.toDouble() ??
            0.0;

        List<Map<String, dynamic>> enrichedProjections = [];
        double projectedOmzet =
            (existing?['projected_omzet'] as num?)?.toDouble() ?? 0.0;

        if (entry.containsKey('projections') && entry['projections'] is List) {
          projectedOmzet = 0.0;
          for (final proj in (entry['projections'] as List)) {
            if (proj is Map) {
              final pId = (proj['product_id'] ?? '').toString();
              final qty = (proj['quantity'] as num?)?.toInt() ?? 0;
              final prod = productLookup[pId];
              final price = prod?.sellingPrice ??
                  (proj['price'] as num?)?.toDouble() ??
                  0.0;
              final subtotal = price * qty;
              projectedOmzet += subtotal;
              enrichedProjections.add({
                'product_id': pId,
                'product_name': prod?.name ?? proj['product_name'] ?? 'Produk',
                'quantity': qty,
                'price': price,
                'subtotal': subtotal,
              });
            }
          }
        } else if (existing?['projections'] is List) {
          enrichedProjections = (existing!['projections'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }

        planByStore[sId] = {
          'visit_id': existing?['visit_id'] ?? '${dateStr}_$sId',
          'store_id': sId,
          'store_name': storeName,
          'address': address,
          'latitude': lat,
          'longitude': lng,
          'status': existing?['status'] ?? 'PLANNED',
          'is_extra': existing?['is_extra'] ?? false,
          'projected_omzet': projectedOmzet,
          'projections': enrichedProjections,
        };
      }

      await _prefs.setStringList(
          'deleted_stores_$dateStr', deletedSet.toList());
      await _prefs.setStringList(
          'offline_plan_pending_stores_$dateStr', pendingStoreIds.toList());
      await _persistVisitPlansLocally(dateStr, planByStore.values.toList());
    } catch (_) {}

    // 2. Coba kirim ke server; jika offline, antrekan ke db.offlineTasks / fallback queue
    try {
      await apiClient.client.post('sales/visit-plans/manual', data: data);
      // Berhasil dikirim ke server -> bersihkan penanda pending untuk toko-toko ini
      final pendingStoreIds =
          (_prefs.getStringList('offline_plan_pending_stores_$dateStr') ?? [])
              .toSet();
      if (stores != null) {
        for (final s in stores) {
          pendingStoreIds.remove((s['store_id'] ?? '').toString());
        }
      }
      if (storeIds != null) {
        pendingStoreIds.removeAll(storeIds);
      }
      await _prefs.setStringList(
          'offline_plan_pending_stores_$dateStr', pendingStoreIds.toList());
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        await _enqueueOfflineTask(
          type: 'VISIT_PLAN_SAVE',
          payload: data,
        );
        return;
      }
      throw Exception('Failed to save manual visit plan: $e');
    }
  }

  // GET /api/v1/sales/visit-plans/builder
  Future<List<dynamic>> getVisitPlanBuilder(
      {String? salesmanId, String? branchId}) async {
    try {
      final queryParams = <String, dynamic>{
        if (salesmanId != null && salesmanId.isNotEmpty)
          'salesman_id': salesmanId,
        if (branchId != null && branchId.isNotEmpty) 'branch_id': branchId,
      };
      final response = await apiClient.client.get(
        'sales/visit-plans/builder',
        queryParameters: queryParams,
      );
      final serverData = (response.data['data'] as List<dynamic>?) ?? [];

      // Sertakan juga toko yang dibuat secara offline di db.stores dan belum ada di serverData
      final offlineStores = await (_db.select(_db.stores)
            ..where((t) => t.isOffline.equals(true)))
          .get();
      if (offlineStores.isNotEmpty) {
        final existingIds = serverData
            .whereType<Map>()
            .map((m) => (m['store_id'] ?? m['id'] ?? '').toString())
            .toSet();
        for (final s in offlineStores) {
          if (!existingIds.contains(s.id)) {
            serverData.insert(0, {
              'store_id': s.id,
              'store_name': s.name,
              'address': s.address ?? '',
              'latitude': s.latitude ?? 0.0,
              'longitude': s.longitude ?? 0.0,
              'status': 'PLANNED',
              'is_extra': false,
              'is_priority': true,
            });
          }
        }
      }
      return serverData;
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        // Fallback ke db.stores lokal atau cached_stores_list
        final Map<String, Map<String, dynamic>> fallbackStores = {};
        final dbStores = await _db.select(_db.stores).get();
        for (final s in dbStores) {
          fallbackStores[s.id] = {
            'store_id': s.id,
            'store_name': s.name,
            'address': s.address ?? '',
            'latitude': s.latitude ?? 0.0,
            'longitude': s.longitude ?? 0.0,
            'status': 'PLANNED',
            'is_extra': false,
            'is_priority': s.isOffline,
          };
        }
        if (fallbackStores.isEmpty) {
          for (final key in _prefs.getKeys()) {
            if (key.startsWith('cached_stores_list')) {
              final raw = _prefs.getString(key);
              if (raw != null) {
                try {
                  final list = jsonDecode(raw) as List<dynamic>;
                  for (final item in list) {
                    if (item is Map) {
                      final id = (item['id'] ?? item['store_id'] ?? '').toString();
                      if (id.isNotEmpty) {
                        fallbackStores[id] = {
                          'store_id': id,
                          'store_name': item['name'] ?? item['store_name'] ?? 'Toko',
                          'address': item['address'] ?? '',
                          'latitude': (item['latitude'] as num?)?.toDouble() ?? 0.0,
                          'longitude': (item['longitude'] as num?)?.toDouble() ?? 0.0,
                          'status': 'PLANNED',
                          'is_extra': false,
                        };
                      }
                    }
                  }
                } catch (_) {}
              }
            }
          }
        }
        return fallbackStores.values.toList();
      }
      throw Exception('Failed to load visit plan builder data: $e');
    }
  }

  bool _isSalesPerson(String position, String department) {
    final pos = position.toLowerCase().trim();
    final dept = department.toLowerCase().trim();

    const blacklist = [
      'editor', 'validator', 'operator', 'auditor', 'customer', 'service',
      'support', 'designer', 'desain', 'pabrik', 'gudang', 'warehouse',
      'kasir', 'cashier', 'finance', 'keuangan', 'accounting', 'driver',
      'helper', 'security', 'satpam', 'hr', 'direktur', 'director', 'owner',
      'admin', 'kepala gudang'
    ];
    for (final b in blacklist) {
      if (pos.contains(b)) return false;
    }

    final words = pos.split(RegExp(r'[\s/_\-]+'));
    final isSalesKeyword = pos.contains('sales') || pos.contains('marketing') ||
        pos.contains('motoris') || pos.contains('booster') ||
        pos.contains('take order') || pos.contains('task order') ||
        pos.contains('modern trade') || pos.contains('kanvas') || pos.contains('canvas') ||
        dept.contains('sales') || dept.contains('marketing');

    final hasShortCode = words.contains('to') || words.contains('mt') || words.contains('spv');

    return isSalesKeyword || hasShortCode;
  }

  // GET /api/v1/employees/directory
  Future<List<Map<String, dynamic>>> getSalesmenDirectory({String? branchId}) async {
    try {
      final queryParams = <String, dynamic>{
        if (branchId != null && branchId.isNotEmpty && branchId != 'ALL_BRANCHES')
          'branch_id': branchId,
      };
      final response = await apiClient.client.get(
        'employees/directory',
        queryParameters: queryParams.isNotEmpty ? queryParams : null,
      );
      final List<dynamic> raw = (response.data['data'] as List<dynamic>?) ?? [];
      final List<Map<String, dynamic>> results = [];
      for (var item in raw) {
        if (item is Map<String, dynamic>) {
          final position = (item['position'] ?? item['job_title'] ?? '').toString();
          final department = (item['department'] ?? '').toString();
          if (_isSalesPerson(position, department)) {
            results.add(item);
          }
        }
      }
      return results;
    } catch (_) {
      return [];
    }
  }

  // DELETE /api/v1/sales/visits/:id
  Future<void> deleteVisitPlan(String id) async {
    try {
      await apiClient.client.delete('sales/visits/$id');
    } catch (e) {
      throw Exception('Failed to delete visit plan: $e');
    }
  }

  // DELETE /api/v1/sales/visit-plans/:store_id
  Future<void> deleteManualVisitPlan(String storeId, {String? date}) async {
    final String dateStr =
        date ?? DateFormat('yyyy-MM-dd').format(DateTime.now());

    // 1. Hapus dari penyimpanan lokal terlebih dahulu
    try {
      final deletedSet =
          (_prefs.getStringList('deleted_stores_$dateStr') ?? []).toSet();
      deletedSet.add(storeId);
      await _prefs.setStringList(
          'deleted_stores_$dateStr', deletedSet.toList());

      final pendingSet =
          (_prefs.getStringList('offline_plan_pending_stores_$dateStr') ?? [])
              .toSet();
      pendingSet.remove(storeId);
      await _prefs.setStringList(
          'offline_plan_pending_stores_$dateStr', pendingSet.toList());

      final existing = await _loadLocalVisitPlans(dateStr);
      existing.removeWhere((p) => (p['store_id'] ?? '').toString() == storeId);
      await _persistVisitPlansLocally(dateStr, existing);

      try {
        await _ensureOfflineSalesTables();
        await (_db.delete(_db.visitPlans)
              ..where(
                  (t) => t.date.equals(dateStr) & t.storeId.equals(storeId)))
            .go();
      } catch (_) {}
    } catch (_) {}

    // 2. Coba hapus di server; jika offline, antrekan ke db.offlineTasks / fallback queue
    try {
      await apiClient.client.delete(
        'sales/visit-plans/$storeId',
        queryParameters: {'date': dateStr},
      );
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        await _enqueueOfflineTask(
          type: 'VISIT_PLAN_DELETE',
          payload: {
            'store_id': storeId,
            'date': dateStr,
          },
        );
        return;
      }
      throw Exception('Failed to delete manual visit plan: $e');
    }
  }

  // GET /api/v1/sales/orders — ambil semua SO milik sales yang sedang login
  Future<List<dynamic>> getSalesOrders() async {
    try {
      final response = await apiClient.client.get(
        'sales/orders',
        queryParameters: {'_t': DateTime.now().millisecondsSinceEpoch},
      );
      return (response.data['data'] as List<dynamic>?) ?? [];
    } catch (e) {
      throw Exception('Failed to load sales orders: $e');
    }
  }

  // PATCH /api/v1/sales/orders/:id/edit — sales edit SO yang masih DRAFT
  Future<Map<String, dynamic>> editSalesOrder(
      String soId, Map<String, dynamic> data) async {
    try {
      final response =
          await apiClient.client.patch('sales/orders/$soId/edit', data: data);
      return (response.data['data'] as Map<String, dynamic>?) ?? {};
    } catch (e) {
      throw Exception('Failed to edit sales order: $e');
    }
  }

  // DELETE /api/v1/sales/orders/:id/draft — sales hapus SO yang masih DRAFT
  Future<void> deleteSalesOrder(String soId) async {
    try {
      await apiClient.client.delete('sales/orders/$soId/draft');
    } catch (e) {
      throw Exception('Failed to delete sales order: $e');
    }
  }

  Map<String, dynamic> _journeyRowToMap(SalesDailyJourney row) {
    return {
      'id': row.id,
      'user_id': row.userId,
      'date': row.date,
      'start_km': row.startKm,
      'start_photo_url': row.startPhotoUrl,
      'start_location': row.startLocation,
      'start_time': row.startTime.toIso8601String(),
      'end_km': row.endKm,
      'end_photo_url': row.endPhotoUrl,
      'end_location': row.endLocation,
      'end_time': row.endTime?.toIso8601String(),
      'total_km': row.totalKm,
      'status': row.status.toUpperCase(),
      'is_synced': row.isSynced,
    };
  }

  // GET /api/v1/sales/journeys/today
  Future<Map<String, dynamic>?> getTodayJourney() async {
    final todayStr = DateFormat('yyyy-MM-dd').format(DateTime.now());
    SalesDailyJourney? localRow;
    try {
      localRow = await (_db.select(_db.salesDailyJourneys)
            ..where((t) => t.date.equals(todayStr))
            ..limit(1))
          .getSingleOrNull();
    } catch (_) {}

    try {
      final response = await apiClient.client.get('sales/journeys/today');
      Map<String, dynamic>? serverMap;
      if (response.data is Map<String, dynamic>) {
        final map = response.data as Map<String, dynamic>;
        if (map['data'] != null && map['data'] is Map<String, dynamic>) {
          serverMap = map['data'] as Map<String, dynamic>;
        } else {
          serverMap = map;
        }
      }

      // Jika data lokal belum ter-sync (misal baru saja Start/End Journey saat offline), prioritaskan data lokal
      if (localRow != null && !localRow.isSynced) {
        return _journeyRowToMap(localRow);
      }

      if (serverMap != null && serverMap.isNotEmpty) {
        final sDate = (serverMap['date'] ?? todayStr).toString().substring(0, 10);
        if (sDate == todayStr) {
          await _db.into(_db.salesDailyJourneys).insertOnConflictUpdate(
                SalesDailyJourneysCompanion.insert(
                  id: (serverMap['id'] ?? 'journey_$todayStr').toString(),
                  userId: (serverMap['user_id'] ?? _getLoggedInUserId()).toString(),
                  date: todayStr,
                  startKm: (serverMap['start_km'] as num?)?.toInt() ?? 0,
                  startPhotoUrl: (serverMap['start_photo_url'] ?? '').toString(),
                  startLocation: (serverMap['start_location'] ?? '').toString(),
                  startTime: DateTime.tryParse(
                          (serverMap['start_time'] ?? '').toString()) ??
                      DateTime.now(),
                  endKm: drift.Value((serverMap['end_km'] as num?)?.toInt()),
                  endPhotoUrl:
                      drift.Value(serverMap['end_photo_url']?.toString()),
                  endLocation:
                      drift.Value(serverMap['end_location']?.toString()),
                  endTime: drift.Value(serverMap['end_time'] != null
                      ? DateTime.tryParse(serverMap['end_time'].toString())
                      : null),
                  totalKm:
                      drift.Value((serverMap['total_km'] as num?)?.toInt()),
                  status: drift.Value(
                      (serverMap['status'] ?? 'STARTED').toString().toUpperCase()),
                  isSynced: const drift.Value(true),
                ),
              );
        }
        return serverMap;
      }
      return localRow != null ? _journeyRowToMap(localRow) : null;
    } catch (e) {
      if (e.toString().contains('404')) {
        if (localRow != null && !localRow.isSynced) {
          return _journeyRowToMap(localRow);
        }
        return null;
      }
      if (_isNetworkOrOfflineError(e)) {
        return localRow != null ? _journeyRowToMap(localRow) : null;
      }
      return localRow != null ? _journeyRowToMap(localRow) : null;
    }
  }

  // POST /api/v1/sales/journeys/start
  Future<Map<String, dynamic>> startJourney(Map<String, dynamic> data) async {
    final todayStr = DateFormat('yyyy-MM-dd').format(DateTime.now());
    final startKm = (data['start_km'] as num?)?.toInt() ?? 0;
    final startPhoto = (data['start_photo_url'] ?? '').toString();
    final startLoc = (data['start_location'] ?? '').toString();
    final now = DateTime.now();

    // Simpan ke DB lokal terlebih dahulu (isSynced = false)
    final localId = 'journey_$todayStr';
    await _ensureOfflineSalesTables();
    await _db.into(_db.salesDailyJourneys).insertOnConflictUpdate(
          SalesDailyJourneysCompanion.insert(
            id: localId,
            userId: _getLoggedInUserId(),
            date: todayStr,
            startKm: startKm,
            startPhotoUrl: startPhoto,
            startLocation: startLoc,
            startTime: now,
            status: const drift.Value('STARTED'),
            isSynced: const drift.Value(false),
          ),
        );

    // Jika foto masih berupa path lokal (karena gagal upload saat offline), jangan paksa kirim ke endpoint dulu, biar SyncRepository yang upload foto & kirim nanti
    final bool isLocalPhoto = startPhoto.isNotEmpty &&
        !startPhoto.startsWith('http') &&
        !startPhoto.startsWith('/uploads') &&
        !startPhoto.startsWith('uploads') &&
        !startPhoto.startsWith('/api/v1/uploads') &&
        !startPhoto.startsWith('api/v1/uploads') &&
        !startPhoto.startsWith('/api/uploads') &&
        !startPhoto.startsWith('api/uploads');
    if (isLocalPhoto) {
      return {
        'id': localId,
        'date': todayStr,
        'start_km': startKm,
        'start_photo_url': startPhoto,
        'start_location': startLoc,
        'status': 'STARTED',
        'is_offline': true,
      };
    }

    try {
      final response =
          await apiClient.client.post('sales/journeys/start', data: data);
      Map<String, dynamic> result = {};
      if (response.data is Map<String, dynamic>) {
        final map = response.data as Map<String, dynamic>;
        if (map['data'] != null && map['data'] is Map<String, dynamic>) {
          result = map['data'] as Map<String, dynamic>;
        } else {
          result = map;
        }
      }
      await (_db.update(_db.salesDailyJourneys)
            ..where((t) => t.date.equals(todayStr)))
          .write(const SalesDailyJourneysCompanion(isSynced: drift.Value(true)));
      return result;
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        return {
          'id': localId,
          'date': todayStr,
          'start_km': startKm,
          'start_photo_url': startPhoto,
          'start_location': startLoc,
          'status': 'STARTED',
          'is_offline': true,
        };
      }
      if (e is DioException) {
        final errorMsg = e.response?.data is Map
            ? (e.response?.data['message'] ?? e.response?.data['error'])
            : null;
        if (errorMsg != null && errorMsg.toString().isNotEmpty) {
          throw Exception(errorMsg);
        }
      }
      throw Exception('Gagal mencatat KM Berangkat: $e');
    }
  }

  // POST /api/v1/sales/journeys/end
  Future<Map<String, dynamic>> endJourney(Map<String, dynamic> data) async {
    final todayStr = DateFormat('yyyy-MM-dd').format(DateTime.now());
    final endKm = (data['end_km'] as num?)?.toInt() ?? 0;
    final endPhoto = (data['end_photo_url'] ?? '').toString();
    final endLoc = (data['end_location'] ?? '').toString();
    final now = DateTime.now();

    await _ensureOfflineSalesTables();
    final existing = await (_db.select(_db.salesDailyJourneys)
          ..where((t) => t.date.equals(todayStr))
          ..limit(1))
        .getSingleOrNull();

    if (existing != null && endKm < existing.startKm) {
      throw Exception(
          'KM Pulang ($endKm) tidak boleh lebih kecil dari KM Berangkat (${existing.startKm})');
    }

    final int startKm = existing?.startKm ?? 0;
    final int totalKm = (endKm - startKm).clamp(0, 999999);
    final String localId = existing?.id ?? 'journey_$todayStr';

    await _db.into(_db.salesDailyJourneys).insertOnConflictUpdate(
          SalesDailyJourneysCompanion.insert(
            id: localId,
            userId: existing?.userId ?? _getLoggedInUserId(),
            date: todayStr,
            startKm: startKm,
            startPhotoUrl: existing?.startPhotoUrl ?? '',
            startLocation: existing?.startLocation ?? '',
            startTime: existing?.startTime ?? now,
            endKm: drift.Value(endKm),
            endPhotoUrl: drift.Value(endPhoto),
            endLocation: drift.Value(endLoc),
            endTime: drift.Value(now),
            totalKm: drift.Value(totalKm),
            status: const drift.Value('ENDED'),
            isSynced: const drift.Value(false),
          ),
        );

    final bool isLocalPhoto = endPhoto.isNotEmpty &&
        !endPhoto.startsWith('http') &&
        !endPhoto.startsWith('/uploads') &&
        !endPhoto.startsWith('uploads') &&
        !endPhoto.startsWith('/api/v1/uploads') &&
        !endPhoto.startsWith('api/v1/uploads') &&
        !endPhoto.startsWith('/api/uploads') &&
        !endPhoto.startsWith('api/uploads');
    final bool startNotSyncedYet = existing != null &&
        existing.startPhotoUrl.isNotEmpty &&
        !existing.startPhotoUrl.startsWith('http') &&
        !existing.startPhotoUrl.startsWith('/uploads') &&
        !existing.startPhotoUrl.startsWith('uploads') &&
        !existing.startPhotoUrl.startsWith('/api/v1/uploads') &&
        !existing.startPhotoUrl.startsWith('api/v1/uploads') &&
        !existing.startPhotoUrl.startsWith('/api/uploads') &&
        !existing.startPhotoUrl.startsWith('api/uploads');

    if (isLocalPhoto || startNotSyncedYet) {
      return {
        'id': localId,
        'date': todayStr,
        'start_km': startKm,
        'end_km': endKm,
        'total_km': totalKm,
        'end_photo_url': endPhoto,
        'end_location': endLoc,
        'status': 'ENDED',
        'is_offline': true,
      };
    }

    try {
      final response =
          await apiClient.client.post('sales/journeys/end', data: data);
      Map<String, dynamic> result = {};
      if (response.data is Map<String, dynamic>) {
        final map = response.data as Map<String, dynamic>;
        if (map['data'] != null && map['data'] is Map<String, dynamic>) {
          result = map['data'] as Map<String, dynamic>;
        } else {
          result = map;
        }
      }
      await (_db.update(_db.salesDailyJourneys)
            ..where((t) => t.date.equals(todayStr)))
          .write(const SalesDailyJourneysCompanion(isSynced: drift.Value(true)));
      return result;
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        return {
          'id': localId,
          'date': todayStr,
          'start_km': startKm,
          'end_km': endKm,
          'total_km': totalKm,
          'end_photo_url': endPhoto,
          'end_location': endLoc,
          'status': 'ENDED',
          'is_offline': true,
        };
      }
      if (e is DioException) {
        final errorMsg = e.response?.data is Map
            ? (e.response?.data['message'] ?? e.response?.data['error'])
            : null;
        if (errorMsg != null && errorMsg.toString().isNotEmpty) {
          throw Exception(errorMsg);
        }
      }
      throw Exception('Gagal mencatat KM Pulang: $e');
    }
  }

  // GET /api/v1/sales/journeys/history
  Future<List<dynamic>> getJourneyHistory(
      {String? startDate, String? endDate}) async {
    List<SalesDailyJourney> localRows = [];
    try {
      localRows = await (_db.select(_db.salesDailyJourneys)
            ..orderBy([
              (t) => drift.OrderingTerm(
                  expression: t.date, mode: drift.OrderingMode.desc)
            ]))
          .get();
    } catch (_) {}

    try {
      final Map<String, dynamic> queryParams = {};
      if (startDate != null && startDate.isNotEmpty) {
        queryParams['start_date'] = startDate;
      }
      if (endDate != null && endDate.isNotEmpty) {
        queryParams['end_date'] = endDate;
      }

      final response = await apiClient.client.get(
        'sales/journeys/history',
        queryParameters: queryParams.isNotEmpty ? queryParams : null,
      );

      List<dynamic> serverList = [];
      if (response.data is Map<String, dynamic>) {
        final map = response.data as Map<String, dynamic>;
        if (map['data'] is List<dynamic>) {
          serverList = List<dynamic>.from(map['data'] as List<dynamic>);
        }
      } else if (response.data is List<dynamic>) {
        serverList = List<dynamic>.from(response.data as List<dynamic>);
      }

      final existingDates = serverList
          .whereType<Map>()
          .map((m) => (m['date'] ?? '').toString().substring(0, 10))
          .toSet();

      for (final lr in localRows) {
        if (!lr.isSynced || !existingDates.contains(lr.date)) {
          serverList.removeWhere((m) =>
              m is Map &&
              (m['date'] ?? '').toString().startsWith(lr.date));
          serverList.insert(0, _journeyRowToMap(lr));
        }
      }
      return serverList;
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        return localRows.map(_journeyRowToMap).toList();
      }
      throw Exception('Failed to get journey history: $e');
    }
  }

  // PATCH /api/v1/sales/journeys/:id
  Future<Map<String, dynamic>> updateJourney(
      String journeyId, Map<String, dynamic> data) async {
    try {
      final response = await apiClient.client.patch(
        'sales/journeys/$journeyId',
        data: data,
      );
      if (response.data is Map<String, dynamic>) {
        final map = response.data as Map<String, dynamic>;
        return (map['data'] as Map<String, dynamic>?) ?? map;
      }
      return {};
    } catch (e) {
      if (e is DioException) {
        final errorMsg = e.response?.data is Map
            ? (e.response?.data['message'] ?? e.response?.data['error'])
            : null;
        if (errorMsg != null && errorMsg.toString().isNotEmpty) {
          throw Exception(errorMsg);
        }
      }
      throw Exception('Gagal memperbarui data perjalanan: $e');
    }
  }

  // ─── Modern Trade Store Stock & Collection Methods ─────────────────────────

  Future<List<Map<String, dynamic>>> getMTStoreStocks({String? storeId}) async {
    try {
      final response = await apiClient.client.get(
        'modern-trade/store-stocks',
        queryParameters: storeId != null ? {'store_id': storeId} : null,
      );
      if (response.data is Map<String, dynamic> &&
          response.data['data'] is List) {
        final list = (response.data['data'] as List)
            .map((e) => Map<String, dynamic>.from(e))
            .toList();
        if (storeId != null) {
          await _prefs.setString(
              'cached_mt_store_stocks_$storeId', jsonEncode(list));
        }
        return list;
      }
      return [];
    } catch (e) {
      if (storeId != null) {
        final cached = _prefs.getString('cached_mt_store_stocks_$storeId');
        if (cached != null) {
          try {
            return (jsonDecode(cached) as List)
                .map((e) => Map<String, dynamic>.from(e as Map))
                .toList();
          } catch (_) {}
        }
      }
      return [];
    }
  }

  Future<bool> adjustMTStoreStock({
    required String storeId,
    required String productId,
    required int quantity,
    int minStockThreshold = 20,
    String? notes,
    String? referenceNo,
    String adjustmentType = 'ADJUSTMENT',
  }) async {
    return adjustMTStoreStocksBulk(
      storeId: storeId,
      items: [
        {
          'product_id': productId,
          'quantity': quantity,
          'min_stock_threshold': minStockThreshold,
        }
      ],
      notes: notes,
      referenceNo: referenceNo,
      adjustmentType: adjustmentType,
    );
  }

  Future<bool> adjustMTStoreStocksBulk({
    required String storeId,
    required List<Map<String, dynamic>> items,
    String? notes,
    String? referenceNo,
    String adjustmentType = 'ADJUSTMENT',
  }) async {
    final payload = {
      'store_id': storeId,
      'items': items,
      'notes': notes,
      'reference_no': referenceNo,
      'adjustment_type': adjustmentType,
    };
    try {
      final response = await apiClient.client.post(
        'modern-trade/store-stocks/adjust',
        data: payload,
      );
      return response.statusCode == 200 || response.statusCode == 201;
    } catch (e) {
      if (_isNetworkOrOfflineError(e)) {
        await _enqueueOfflineTask(
          type: 'MT_STOCK_OPNAME',
          payload: payload,
        );
        return true;
      }
      return false;
    }
  }

  Future<List<Map<String, dynamic>>> getActiveProducts() async {
    final localProducts = await _getProductsFromLocalDb();
    try {
      if (localProducts.isNotEmpty) {
        return localProducts;
      }
      final response = await apiClient.client.get(
        'factory/products',
        queryParameters: {
          'scope': 'sales',
          'type': 'finished_goods',
          'limit': 200,
        },
      );
      if (response.data is Map<String, dynamic> &&
          response.data['data'] is List) {
        return (response.data['data'] as List)
            .map((e) => Map<String, dynamic>.from(e))
            .toList();
      } else if (response.data is List) {
        return (response.data as List)
            .map((e) => Map<String, dynamic>.from(e))
            .toList();
      }
      return [];
    } catch (e) {
      return localProducts;
    }
  }

  Future<List<Map<String, dynamic>>> getMTPurchaseOrders({String? status}) async {
    try {
      final response = await apiClient.client.get(
        'modern-trade/purchase-orders',
      );
      if (response.data is Map<String, dynamic> &&
          response.data['data'] is List) {
        final list = (response.data['data'] as List)
            .map((e) => Map<String, dynamic>.from(e))
            .toList();
        await _prefs.setString('cached_mt_purchase_orders', jsonEncode(list));
        return list;
      }
      return [];
    } catch (e) {
      final cached = _prefs.getString('cached_mt_purchase_orders');
      if (cached != null) {
        try {
          return (jsonDecode(cached) as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        } catch (_) {}
      }
      return [];
    }
  }

  Future<bool> collectMTPurchaseOrderPayment({
    required String poId,
    required FormData formData,
  }) async {
    try {
      final response = await apiClient.client.post(
        'modern-trade/purchase-orders/$poId/collect',
        data: formData,
      );
      return response.statusCode == 200 || response.statusCode == 201;
    } catch (e) {
      return false;
    }
  }

  Future<List<Map<String, dynamic>>> getBankAccounts({
    String? branchId,
    String? companyId,
  }) async {
    final cacheKey =
        'cached_bank_accounts_${branchId ?? "all"}_${companyId ?? "all"}';
    try {
      final queryParams = <String, dynamic>{};
      if (branchId != null &&
          branchId.isNotEmpty &&
          branchId != '00000000-0000-0000-0000-000000000000') {
        queryParams['branch_id'] = branchId;
      }
      if (companyId != null &&
          companyId.isNotEmpty &&
          companyId != '00000000-0000-0000-0000-000000000000') {
        queryParams['company_id'] = companyId;
      }

      final response = await apiClient.client.get(
        'finance/bank-accounts',
        queryParameters: queryParams,
      );
      if (response.statusCode == 200 && response.data != null) {
        List<Map<String, dynamic>> banks = [];
        if (response.data['data'] is List) {
          banks = (response.data['data'] as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        } else if (response.data is List) {
          banks = (response.data as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        }
        if (banks.isNotEmpty) {
          await _prefs.setString(cacheKey, jsonEncode(banks));
          await _prefs.setString('cached_bank_accounts_default', jsonEncode(banks));
        }
        return banks;
      }
      return [];
    } catch (e) {
      final cached = _prefs.getString(cacheKey) ??
          _prefs.getString('cached_bank_accounts_default');
      if (cached != null && cached.isNotEmpty) {
        try {
          return (jsonDecode(cached) as List)
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
        } catch (_) {}
      }
      return [];
    }
  }

  // --- DISTRIBUTOR (SECONDARY SALES / SELL-OUT) ---

  Future<List<Map<String, dynamic>>> getDistributorStores({String? search, int page = 1, int limit = 50}) async {
    try {
      final response = await apiClient.client.get('distributor/stores', queryParameters: {
        if (search != null && search.isNotEmpty) 'search': search,
        'page': page,
        'limit': limit,
      });
      if (response.statusCode == 200 && response.data != null) {
        final data = response.data['data'];
        if (data != null && data['stores'] is List) {
          return (data['stores'] as List).map((e) => Map<String, dynamic>.from(e as Map)).toList();
        }
      }
      return [];
    } catch (e) {
      return [];
    }
  }

  Future<Map<String, dynamic>?> createDistributorStore(Map<String, dynamic> storeData) async {
    final response = await apiClient.client.post('distributor/stores', data: storeData);
    if (response.statusCode == 200 || response.statusCode == 201) {
      return Map<String, dynamic>.from(response.data['data'] as Map);
    }
    return null;
  }

  Future<Map<String, dynamic>?> recordDistributorVisit({
    required String storeId,
    required double latitude,
    required double longitude,
    String? selfieUrl,
    String? notes,
  }) async {
    final response = await apiClient.client.post('distributor/visits', data: {
      'store_id': storeId,
      'latitude': latitude,
      'longitude': longitude,
      if (selfieUrl != null) 'selfie_url': selfieUrl,
      if (notes != null) 'notes': notes,
    });
    if (response.statusCode == 200 || response.statusCode == 201) {
      return Map<String, dynamic>.from(response.data['data'] as Map);
    }
    return null;
  }

  Future<Map<String, dynamic>?> createDistributorSale(Map<String, dynamic> saleData) async {
    final response = await apiClient.client.post('distributor/sales', data: saleData);
    if (response.statusCode == 200 || response.statusCode == 201) {
      return Map<String, dynamic>.from(response.data['data'] as Map);
    }
    return null;
  }
}


