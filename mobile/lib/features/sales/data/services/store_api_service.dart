import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:drift/drift.dart' as drift;
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/core/network/network_info.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:uuid/uuid.dart';

class StoreApiService {
  final ApiClient apiClient;
  final SharedPreferences prefs; // Injected singleton — no async needed

  // Global lock to prevent multiple concurrent background fetches
  static bool _isFetchingPaginated = false;
  // Track last refresh time in memory
  static DateTime? _lastRefreshTime;

  StoreApiService({required this.apiClient, required this.prefs});

  String _cacheKey(String status, String? assignedEmployeeId) {
    final branchId = _getLoggedInBranchId();
    return 'cached_stores_list_${branchId ?? "all"}_${status}_${assignedEmployeeId ?? "all"}';
  }

  String? _getLoggedInEmployeeId() {
    final cached = prefs.getString('cached_employee_profile');
    if (cached != null) {
      try {
        final Map<String, dynamic> decoded = jsonDecode(cached);
        return decoded['id']?.toString();
      } catch (_) {}
    }
    return null;
  }

  String? _getLoggedInBranchId() {
    final cached = prefs.getString('cached_employee_profile');
    if (cached != null) {
      try {
        final Map<String, dynamic> decoded = jsonDecode(cached);
        final branchId = decoded['branch_id'] ??
            decoded['BranchID'] ??
            decoded['branch']?['id'] ??
            decoded['Branch']?['ID'] ??
            decoded['employee']?['branch_id'];
        if (branchId != null) return branchId.toString();
      } catch (_) {}
    }
    return null;
  }

  bool _isManagementRole() {
    final cached = prefs.getString('cached_employee_profile');
    if (cached != null) {
      try {
        final Map<String, dynamic> decoded = jsonDecode(cached);
        final jobTitle = (decoded['job_position']?['title'] ?? decoded['job_title'] ?? '').toString().toLowerCase();
        final role = (decoded['role'] ?? '').toString().toLowerCase();
        return role == 'direktur' || role == 'director' || role == 'owner' || 
               role == 'superadmin' || role == 'manager' || role == 'supervisor' || role == 'spv' ||
               role == 'supervisor_sales' || role == 'supervisor_sales_mt' ||
               jobTitle.contains('direktur') || jobTitle.contains('director') || 
               jobTitle.contains('manager') || jobTitle.contains('owner') ||
               jobTitle.contains('spv') || jobTitle.contains('supervisor');
      } catch (_) {}
    }
    return false;
  }

  String? _resolveAssignedEmployeeId(String? assignedEmployeeId) {
    if (assignedEmployeeId == 'ALL') return null;
    final loggedInId = _getLoggedInEmployeeId();
    if (_isManagementRole()) {
      // If assignedEmployeeId is omitted or matches logged-in supervisor's own ID,
      // unlock all branch stores (return null).
      // Only pass assignedEmployeeId if an explicit different sales ID was selected for filtering.
      if (assignedEmployeeId == null || assignedEmployeeId.isEmpty || assignedEmployeeId == loggedInId) {
        return null;
      }
      return assignedEmployeeId;
    }
    if (assignedEmployeeId != null && assignedEmployeeId.isNotEmpty) return assignedEmployeeId;
    return loggedInId;
  }

  /// SYNCHRONOUS — reads from in-memory SharedPreferences instantly (0ms).
  /// Returns null if no cache exists yet (first install).
  List<StoreModel>? getStoresCached({String status = 'ACTIVE', String? assignedEmployeeId}) {
    assignedEmployeeId = _resolveAssignedEmployeeId(assignedEmployeeId);

    final key = _cacheKey(status, assignedEmployeeId);
    final cachedString = prefs.getString(key) ?? prefs.getString('cached_stores_list');

    debugPrint('[STORE CACHE] getStoresCached() → key=$key, found=${cachedString != null}, len=${cachedString?.length ?? 0}');

    if (cachedString == null) return null;
    try {
      final List<dynamic> decoded = jsonDecode(cachedString);
      debugPrint('[STORE CACHE] Decoded ${decoded.length} stores from cache ✓');
      return decoded.map((json) => StoreModel.fromJson(json)).toList();
    } catch (e) {
      debugPrint('[STORE CACHE] jsonDecode FAILED: $e');
      return null;
    }
  }

  // GET /api/v1/stores — async network fetch, also updates cache
  Future<List<StoreModel>> getStores({
    String? sortBy,
    String status = 'ACTIVE',
    String? assignedEmployeeId,
    bool forceRefresh = false,
  }) async {
    assignedEmployeeId = _resolveAssignedEmployeeId(assignedEmployeeId);

    try {
      final cacheKey = _cacheKey(status, assignedEmployeeId);

      // Fast path: return cached data immediately and refresh in background
      // using paginated fetch (safe on slow 3G — no single big request)
      if (!forceRefresh) {
        final cachedString = prefs.getString(cacheKey) ?? prefs.getString('cached_stores_list');
        if (cachedString != null) {
          try {
            final List<dynamic> decoded = jsonDecode(cachedString);
            final cachedStores = decoded.map((json) => StoreModel.fromJson(json)).toList();
            getStoresPaginated(
              sortBy: sortBy,
              status: status,
              assignedEmployeeId: assignedEmployeeId,
            ).ignore();
            return cachedStores;
          } catch (_) {}
        }
      }

      // Slow path: cache empty or forceRefresh requested
      return await getStoresPaginated(
        sortBy: sortBy,
        status: status,
        assignedEmployeeId: assignedEmployeeId,
        forceRefresh: forceRefresh,
      );
    } catch (e) {
      final localDbStores = await getStoresFromLocalDb();
      if (localDbStores.isNotEmpty) return localDbStores;
      final cached = getStoresCached(status: status, assignedEmployeeId: assignedEmployeeId);
      if (cached != null && cached.isNotEmpty) return cached;
      throw Exception('Failed to load stores: $e');
    }
  }

  Future<List<StoreModel>> getStoresFromLocalDb() async {
    try {
      final db = di.sl<AppDatabase>();
      final rows = await db.select(db.stores).get();
      if (rows.isEmpty) return [];
      return rows.map((r) => StoreModel(
        id: r.id,
        name: r.name,
        address: r.address ?? '',
        latitude: r.latitude ?? 0.0,
        longitude: r.longitude ?? 0.0,
        ownerName: r.ownerName ?? '',
        phoneNumber: r.phoneNumber ?? '',
        photoUrl: r.photoUrl ?? '',
        localPhotoPath: r.localPhotoPath,
        province: r.province ?? '',
        regency: r.regency ?? '',
        district: r.district ?? '',
        village: r.village ?? '',
        visitFrequency: r.visitFrequency ?? '',
        visitDays: (r.visitDays != null && r.visitDays!.isNotEmpty)
            ? r.visitDays!.split(',').map((e) => int.tryParse(e.trim()) ?? 1).toList()
            : const [],
        assignedEmployeeId: r.assignedEmployeeId,
        isOffline: r.isOffline,
        syncStatus: r.syncStatus,
        accuracy: r.accuracy,
      )).toList();
    } catch (e) {
      debugPrint('[STORE DB] getStoresFromLocalDb error: $e');
      return [];
    }
  }

  /// Fetch stores page-by-page (150 per request to reduce roundtrips).
  /// Calls [onPartialLoad] after each page so the UI can update incrementally
  /// even on a slow 3G connection. All data is saved to cache and local SQLite when done.
  Future<List<StoreModel>> getStoresPaginated({
    String? sortBy,
    String status = 'ACTIVE',
    String? assignedEmployeeId,
    int pageSize = 150,
    bool forceRefresh = false,
    Function(List<StoreModel>)? onPartialLoad,
  }) async {
    assignedEmployeeId = _resolveAssignedEmployeeId(assignedEmployeeId);

    // 1. If currently fetching, skip so we don't cause a network storm
    if (_isFetchingPaginated && !forceRefresh) {
      debugPrint('[STORE CACHE] Fetch already in progress, skipping...');
      final cached = getStoresCached(status: status, assignedEmployeeId: assignedEmployeeId);
      if (cached != null && cached.isNotEmpty) return cached;
      return await getStoresFromLocalDb();
    }

    // 2. If we refreshed within the last 5 minutes AND cache actually has data, skip (unless forceRefresh)
    if (!forceRefresh && _lastRefreshTime != null && DateTime.now().difference(_lastRefreshTime!).inMinutes < 5) {
      final cached = getStoresCached(status: status, assignedEmployeeId: assignedEmployeeId);
      if (cached != null && cached.isNotEmpty) {
        debugPrint('[STORE CACHE] Cache is fresh and has data (${cached.length} stores), skipping fetch...');
        return cached;
      }
      debugPrint('[STORE CACHE] Cache is fresh by time but data missing for this key, forcing fetch...');
    }

    _isFetchingPaginated = true;
    final cacheKey = _cacheKey(status, assignedEmployeeId);
    final allStores = <StoreModel>[];
    
    try {
      int page = 1;
      bool hasMore = true;

      final branchId = _getLoggedInBranchId();
      while (hasMore) {
        final params = <String, dynamic>{
          'status': status,
          'page': page,
          'limit': pageSize,
          if (sortBy != null) 'sort_by': sortBy,
          if (branchId != null && branchId.isNotEmpty && branchId != 'ALL_BRANCHES')
            'branch_id': branchId,
          if (assignedEmployeeId != null && assignedEmployeeId.isNotEmpty)
            'assigned_employee_id': assignedEmployeeId,
        };

        final response = await apiClient.client.get('stores', queryParameters: params);
        final raw = (response.data['data'] as List<dynamic>?) ?? [];
        final pageStores = raw.map((j) => StoreModel.fromJson(j)).toList();

        allStores.addAll(pageStores);

        // Write cache after EVERY page so it's available immediately
        final jsonStr = jsonEncode(allStores.map((s) => s.toJson()).toList());
        final writeOk = await prefs.setString(cacheKey, jsonStr);
        await prefs.setString('cached_stores_list', jsonStr);
        debugPrint('[STORE CACHE] setString page=$page, stores=${allStores.length}, success=$writeOk');

        // Notify caller so UI can render partial data immediately
        onPartialLoad?.call(List.unmodifiable(allStores));

        if (pageStores.length < pageSize) {
          hasMore = false;
        } else {
          page++;
        }
      }

      // Record successful refresh time
      _lastRefreshTime = DateTime.now();

      // Merge any pending offline stores from Drift so they don't disappear before syncing
      try {
        final db = di.sl<AppDatabase>();
        final pendingOfflineRows = await (db.select(db.stores)
              ..where((t) =>
                  t.isOffline.equals(true) | t.syncStatus.equals('pending')))
            .get();
        if (pendingOfflineRows.isNotEmpty) {
          final existingIds = allStores.map((s) => s.id).toSet();
          for (final r in pendingOfflineRows.reversed) {
            if (!existingIds.contains(r.id)) {
              allStores.insert(
                0,
                StoreModel(
                  id: r.id,
                  name: r.name,
                  address: r.address ?? '',
                  latitude: r.latitude ?? 0.0,
                  longitude: r.longitude ?? 0.0,
                  ownerName: r.ownerName ?? '',
                  phoneNumber: r.phoneNumber ?? '',
                  photoUrl: r.photoUrl ?? '',
                  localPhotoPath: r.localPhotoPath,
                  province: r.province ?? '',
                  regency: r.regency ?? '',
                  district: r.district ?? '',
                  village: r.village ?? '',
                  visitFrequency: r.visitFrequency ?? '',
                  visitDays: (r.visitDays != null && r.visitDays!.isNotEmpty)
                      ? r.visitDays!
                          .split(',')
                          .map((e) => int.tryParse(e.trim()) ?? 1)
                          .toList()
                      : const [],
                  assignedEmployeeId: r.assignedEmployeeId,
                  isOffline: r.isOffline,
                  syncStatus: r.syncStatus,
                  accuracy: r.accuracy,
                  isNew: true,
                ),
              );
            }
          }
        }
      } catch (_) {}

      // Final persistence to prefs
      final jsonStr = jsonEncode(allStores.map((s) => s.toJson()).toList());
      await prefs.setString(cacheKey, jsonStr);
      await prefs.setString('cached_stores_list', jsonStr);

      // Async write to Drift database in background (skip overwriting pending offline stores)
      try {
        final db = di.sl<AppDatabase>();
        db.batch((batch) {
          for (final s in allStores) {
            if (s.isOffline || s.syncStatus == 'pending') continue;
            batch.insert(
              db.stores,
              StoresCompanion(
                id: drift.Value(s.id),
                name: drift.Value(s.name),
                address: drift.Value(s.address),
                latitude: drift.Value(s.latitude),
                longitude: drift.Value(s.longitude),
                ownerName: drift.Value(s.ownerName),
                phoneNumber: drift.Value(s.phoneNumber),
                photoUrl: drift.Value(s.photoUrl),
                province: drift.Value(s.province),
                regency: drift.Value(s.regency),
                district: drift.Value(s.district),
                village: drift.Value(s.village),
                visitFrequency: drift.Value(s.visitFrequency),
                visitDays: drift.Value(s.visitDays.join(',')),
                assignedEmployeeId: drift.Value(s.assignedEmployeeId),
                isOffline: const drift.Value(false),
                syncStatus: const drift.Value('synced'),
              ),
              mode: drift.InsertMode.insertOrReplace,
            );
          }
        }).ignore();
      } catch (_) {}

      return allStores;
    } catch (e) {
      if (allStores.isNotEmpty) return allStores;
      final cached = getStoresCached(status: status, assignedEmployeeId: assignedEmployeeId);
      if (cached != null && cached.isNotEmpty) {
        onPartialLoad?.call(List.unmodifiable(cached));
        return cached;
      }
      final localDb = await getStoresFromLocalDb();
      if (localDb.isNotEmpty) {
        onPartialLoad?.call(List.unmodifiable(localDb));
        return localDb;
      }
      rethrow;
    } finally {
      // ALWAYS release the lock, even if network timeouts
      _isFetchingPaginated = false;
    }
  }

  Future<List<StoreModel>> _fetchAndCacheStores(
      String? sortBy, String status, String? assignedEmployeeId, String cacheKey) async {
    final queryParams = <String, dynamic>{};
    if (status.isNotEmpty) queryParams['status'] = status;
    if (sortBy != null) queryParams['sort_by'] = sortBy;
    if (assignedEmployeeId != null && assignedEmployeeId.isNotEmpty) {
      queryParams['assigned_employee_id'] = assignedEmployeeId;
    }

    final response = await apiClient.client.get('stores', queryParameters: queryParams);
    final data = (response.data['data'] as List<dynamic>?) ?? [];
    final storesList = data.map((json) => StoreModel.fromJson(json)).toList();

    final storesJson = jsonEncode(storesList.map((s) => s.toJson()).toList());
    await prefs.setString(cacheKey, storesJson);

    return storesList;
  }

  void _invalidateCacheSync() {
    // Non-destructive invalidation: reset timestamp and lock to allow immediate background re-fetch
    // without wiping out cached data from SharedPreferences (avoids 0ms -> slow spinner regression)
    _lastRefreshTime = null;
    _isFetchingPaginated = false;
  }

  // POST /api/v1/stores
  Future<StoreModel> createStore(Map<String, dynamic> data,
      {dynamic photoBytes, String? photoName}) async {
    try {
      dynamic payload = data;
      if (photoBytes != null && photoName != null) {
        final formData = FormData.fromMap({
          ...data,
          'photo': MultipartFile.fromBytes(photoBytes, filename: photoName),
        });
        if (data['visit_days'] is List) {
          formData.fields.removeWhere((item) => item.key == 'visit_days');
          for (final day in data['visit_days'] as List) {
            formData.fields.add(MapEntry('visit_days[]', day.toString()));
          }
        }
        payload = formData;
      }
      final response = await apiClient.client.post('stores', data: payload);
      final json = response.data['data'];
      _invalidateCacheSync();
      final store = StoreModel.fromJson(json);

      // Also persist to Drift database
      try {
        final db = di.sl<AppDatabase>();
        await db.into(db.stores).insertOnConflictUpdate(
          StoresCompanion(
            id: drift.Value(store.id),
            name: drift.Value(store.name),
            address: drift.Value(store.address),
            latitude: drift.Value(store.latitude),
            longitude: drift.Value(store.longitude),
            ownerName: drift.Value(store.ownerName),
            phoneNumber: drift.Value(store.phoneNumber),
            photoUrl: drift.Value(store.photoUrl),
            province: drift.Value(store.province),
            regency: drift.Value(store.regency),
            district: drift.Value(store.district),
            village: drift.Value(store.village),
            visitFrequency: drift.Value(store.visitFrequency),
            visitDays: drift.Value(store.visitDays.join(',')),
            assignedEmployeeId: drift.Value(store.assignedEmployeeId),
            isOffline: const drift.Value(false),
            syncStatus: const drift.Value('synced'),
          ),
        );
      } catch (_) {}

      return store;
    } catch (e) {
      throw Exception('Failed to create store: $e');
    }
  }

  /// Simpan toko baru ke database lokal (Drift) secara offline
  Future<StoreModel> saveStoreOffline({
    required Map<String, dynamic> data,
    String? localPhotoPath,
    double? accuracy,
  }) async {
    final storeId = (data['id'] != null && data['id'].toString().isNotEmpty)
        ? data['id'].toString()
        : const Uuid().v4();
    final now = DateTime.now();

    final visitDaysList = data['visit_days'];
    String? visitDaysStr;
    if (visitDaysList is List) {
      visitDaysStr = visitDaysList.join(',');
    }

    final storeCompanion = StoresCompanion(
      id: drift.Value(storeId),
      name: drift.Value(data['name'] ?? ''),
      address: drift.Value(data['address']),
      latitude: drift.Value(data['latitude'] != null ? (data['latitude'] as num).toDouble() : null),
      longitude: drift.Value(data['longitude'] != null ? (data['longitude'] as num).toDouble() : null),
      ownerName: drift.Value(data['owner_name']),
      phoneNumber: drift.Value(data['phone_number']),
      photoUrl: drift.Value(data['photo_url']),
      localPhotoPath: drift.Value(localPhotoPath),
      province: drift.Value(data['province']),
      regency: drift.Value(data['regency']),
      district: drift.Value(data['district']),
      village: drift.Value(data['village']),
      visitFrequency: drift.Value(data['visit_frequency']),
      visitDays: drift.Value(visitDaysStr),
      assignedEmployeeId: drift.Value(data['assigned_employee_id']),
      companyId: drift.Value(data['company_id']),
      isOffline: const drift.Value(true),
      syncStatus: const drift.Value('pending'),
      createdAt: drift.Value(now),
      accuracy: drift.Value(accuracy),
    );

    final db = di.sl<AppDatabase>();
    await db.into(db.stores).insertOnConflictUpdate(storeCompanion);

    final storeModel = StoreModel(
      id: storeId,
      name: data['name'] ?? '',
      ownerName: data['owner_name'] ?? '',
      phoneNumber: data['phone_number'] ?? '',
      address: data['address'] ?? '',
      latitude: (data['latitude'] ?? 0.0).toDouble(),
      longitude: (data['longitude'] ?? 0.0).toDouble(),
      isNew: true,
      photoUrl: data['photo_url'] ?? '',
      province: data['province'] ?? '',
      regency: data['regency'] ?? '',
      district: data['district'] ?? '',
      village: data['village'] ?? '',
      visitDays: (visitDaysList is List) ? List<int>.from(visitDaysList) : const [],
      visitFrequency: data['visit_frequency'] ?? '',
      assignedEmployeeId: data['assigned_employee_id'],
      isOffline: true,
      syncStatus: 'pending',
      localPhotoPath: localPhotoPath,
      accuracy: accuracy,
    );

    // Update in-memory SharedPreferences cache agar halaman SelectStorePage & StoreListPage langsung melihat toko ini
    try {
      final keys = prefs.getKeys().where((k) => k.startsWith('cached_stores_list')).toList();
      if (keys.isEmpty) {
        final raw = prefs.getString('cached_stores_list');
        final list = raw != null ? (jsonDecode(raw) as List) : [];
        list.removeWhere((item) => item is Map && item['id'] == storeId);
        list.insert(0, storeModel.toJson());
        await prefs.setString('cached_stores_list', jsonEncode(list));
      } else {
        for (final k in keys) {
          final raw = prefs.getString(k);
          if (raw != null) {
            try {
              final list = jsonDecode(raw) as List;
              list.removeWhere((item) => item is Map && item['id'] == storeId);
              list.insert(0, storeModel.toJson());
              await prefs.setString(k, jsonEncode(list));
            } catch (_) {}
          }
        }
      }
      _invalidateCacheSync();
    } catch (e) {
      debugPrint('[STORE CACHE] Error inserting offline store to cache: $e');
    }

    return storeModel;
  }

  // PUT /api/v1/stores/:id
  Future<StoreModel> updateStore(String id, Map<String, dynamic> data,
      {dynamic photoBytes, String? photoName}) async {
    try {
      dynamic payload = data;
      if (photoBytes != null && photoName != null) {
        final formData = FormData.fromMap({
          ...data,
          'photo': MultipartFile.fromBytes(photoBytes, filename: photoName),
        });
        if (data['visit_days'] is List) {
          formData.fields.removeWhere((item) => item.key == 'visit_days');
          for (final day in data['visit_days'] as List) {
            formData.fields.add(MapEntry('visit_days[]', day.toString()));
          }
        }
        payload = formData;
      }
      final response = await apiClient.client.put('stores/$id', data: payload);
      final json = response.data['data'];
      _invalidateCacheSync();
      return StoreModel.fromJson(json);
    } catch (e) {
      throw Exception('Failed to update store: $e');
    }
  }

  // DELETE /api/v1/stores/:id
  Future<void> deleteStore(String id) async {
    try {
      await apiClient.client.delete('stores/$id');
      _invalidateCacheSync();
    } on DioException catch (e) {
      if (e.response?.data != null && e.response?.data['error'] != null) {
        throw e.response!.data['error'].toString();
      }
      throw 'Failed to delete store: ${e.message}';
    } catch (e) {
      throw 'Failed to delete store: $e';
    }
  }

  // PUT /api/v1/stores/:id/status
  Future<StoreModel> updateStoreStatus(String id, String status) async {
    try {
      final response =
          await apiClient.client.put('stores/$id/status', data: {'status': status});
      final json = response.data['data'];
      _invalidateCacheSync();
      return StoreModel.fromJson(json);
    } on DioException catch (e) {
      throw Exception('Failed to update store status: $e');
    }
  }

  // GET /api/v1/sales/stores/:store_id/orders
  Future<Map<String, dynamic>> getStoreOrders(String storeId, {int page = 1, int limit = 10}) async {
    try {
      final queryParams = {'page': page, 'limit': limit};
      final response = await apiClient.client.get(
        'sales/stores/$storeId/orders',
        queryParameters: queryParams,
      );
      return response.data;
    } on DioException catch (e) {
      throw Exception('Failed to get store orders: $e');
    }
  }

  /// Find store by ID in local SQLite (`db.stores`) or SharedPreferences caches.
  Future<StoreModel?> findStoreLocally(String id) async {
    try {
      final db = di.sl<AppDatabase>();
      final localRow = await (db.select(db.stores)
            ..where((t) => t.id.equals(id))
            ..limit(1))
          .getSingleOrNull();
      if (localRow != null) {
        return StoreModel(
          id: localRow.id,
          name: localRow.name,
          ownerName: localRow.ownerName ?? '',
          phoneNumber: localRow.phoneNumber ?? '',
          address: localRow.address ?? '',
          latitude: localRow.latitude ?? 0.0,
          longitude: localRow.longitude ?? 0.0,
          photoUrl: localRow.photoUrl ?? '',
          localPhotoPath: localRow.localPhotoPath,
          province: localRow.province ?? '',
          regency: localRow.regency ?? '',
          district: localRow.district ?? '',
          village: localRow.village ?? '',
          visitFrequency: localRow.visitFrequency ?? '',
          visitDays: (localRow.visitDays != null && localRow.visitDays!.isNotEmpty)
              ? localRow.visitDays!.split(',').map((e) => int.tryParse(e.trim()) ?? 1).toList()
              : const [],
          assignedEmployeeId: localRow.assignedEmployeeId,
          isOffline: localRow.isOffline,
          syncStatus: localRow.syncStatus,
          accuracy: localRow.accuracy,
        );
      }
    } catch (_) {}

    try {
      final cachedStores = getStoresCached();
      if (cachedStores != null) {
        for (final s in cachedStores) {
          if (s.id == id) return s;
        }
      }

      for (final key in prefs.getKeys()) {
        if (key.startsWith('cached_stores_list')) {
          final raw = prefs.getString(key);
          if (raw == null) continue;
          final decoded = jsonDecode(raw) as List<dynamic>;
          for (final item in decoded) {
            if (item is Map<String, dynamic> && item['id']?.toString() == id) {
              return StoreModel.fromJson(item);
            }
          }
        }
      }
    } catch (_) {}

    return null;
  }

  // GET /api/v1/stores/:id
  Future<StoreModel> getStoreByID(String id) async {
    bool isOnline = true;
    try {
      isOnline = await NetworkInfo().hasInternetAccess;
    } catch (_) {}

    if (!isOnline) {
      final local = await findStoreLocally(id);
      if (local != null) return local;
      throw Exception('Offline: Store $id not found in local storage');
    }

    // Even when online, if this is a newly created offline store not yet synced, return local immediately
    final localPrecheck = await findStoreLocally(id);
    if (localPrecheck != null && localPrecheck.isOffline) {
      return localPrecheck;
    }

    try {
      final response = await apiClient.client.get('stores/$id');
      final json = response.data['data'];
      return StoreModel.fromJson(json);
    } catch (e) {
      if (localPrecheck != null) return localPrecheck;
      final local = await findStoreLocally(id);
      if (local != null) return local;
      throw Exception('Failed to get store details: $e');
    }
  }

  // GET /api/v1/modern-trade/key-accounts
  Future<List<Map<String, dynamic>>> getKeyAccounts() async {
    try {
      final response = await apiClient.client.get('modern-trade/key-accounts');
      final data = (response.data['data'] as List<dynamic>?) ?? [];
      return data.map((item) => item as Map<String, dynamic>).toList();
    } catch (e) {
      throw Exception('Failed to load key accounts: $e');
    }
  }

  // POST /api/v1/modern-trade/key-accounts
  Future<Map<String, dynamic>> createKeyAccount(Map<String, dynamic> data) async {
    try {
      final response =
          await apiClient.client.post('modern-trade/key-accounts', data: data);
      return response.data['data'] as Map<String, dynamic>;
    } catch (e) {
      throw Exception('Failed to create key account: $e');
    }
  }

  // POST /api/v1/stores/check-duplicate
  Future<Map<String, dynamic>> checkDuplicateStore({
    required String name,
    required String phoneNumber,
    double? latitude,
    double? longitude,
    String? district,
    String? regency,
    String? province,
    String? excludeId,
  }) async {
    try {
      final payload = {
        'name': name,
        'phone_number': phoneNumber,
        if (latitude != null) 'latitude': latitude,
        if (longitude != null) 'longitude': longitude,
        if (district != null) 'district': district,
        if (regency != null) 'regency': regency,
        if (province != null) 'province': province,
        if (excludeId != null) 'exclude_id': excludeId,
      };
      final response = await apiClient.client.post('stores/check-duplicate', data: payload);
      return (response.data['data'] as Map<String, dynamic>?) ?? {'has_potential_duplicate': false, 'matches': []};
    } catch (e) {
      debugPrint('[STORE DUP CHECK] Error: $e');
      return {'has_potential_duplicate': false, 'matches': []};
    }
  }
}
