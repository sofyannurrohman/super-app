import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:drift/drift.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/network/network_info.dart';
import 'package:hris_app/core/services/offline_delivery_manager.dart';
import 'package:hris_app/core/services/data_integrity_service.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

class SyncResult {
  final int totalCount;
  final int successCount;
  final int failureCount;
  final List<String> errors;

  const SyncResult({
    this.totalCount = 0,
    this.successCount = 0,
    this.failureCount = 0,
    this.errors = const [],
  });

  bool get hasFailures => failureCount > 0;
  bool get isFullSuccess => totalCount > 0 && failureCount == 0;
}

class SyncRepository {
  final AppDatabase db;
  final ApiClient api;
  final DataIntegrityService _integrity = DataIntegrityService();

  SyncRepository({required this.db, required this.api});

  bool _isAlreadyServerUrl(String url) {
    final trimmed = url.trim();
    return trimmed.startsWith('http://') ||
        trimmed.startsWith('https://') ||
        trimmed.startsWith('/uploads/') ||
        trimmed.startsWith('uploads/') ||
        trimmed.startsWith('/api/v1/uploads/') ||
        trimmed.startsWith('api/v1/uploads/') ||
        trimmed.startsWith('/api/uploads/') ||
        trimmed.startsWith('api/uploads/');
  }

  Future<SyncResult> syncAll({bool force = false}) async {
    bool isOnline = true;
    try {
      isOnline = await NetworkInfo().hasInternetAccess;
    } catch (_) {}
    if (!isOnline) {
      print('DEBUG SYNC: syncAll() skipped because device is offline');
      return const SyncResult();
    }

    print('DEBUG SYNC: syncAll() started');
    int totalCount = 0;
    int successCount = 0;
    int failureCount = 0;
    final List<String> errors = [];

    // 1. Sync Offline Deliveries (Poor-signal fallback)
    try {
      print('DEBUG SYNC: syncOfflineDeliveries started');
      await OfflineDeliveryManager().syncPendingDeliveries(api);
    } catch (e) {
      print('DEBUG SYNC: syncOfflineDeliveries fatal error: $e');
      errors.add('Delivery sync: $e');
    }

    // 2. Sync Offline Stores first so subsequent plans, checkins & orders have valid store on backend
    try {
      final storeResult = await syncOfflineStores(force: force);
      totalCount += storeResult.totalCount;
      successCount += storeResult.successCount;
      failureCount += storeResult.failureCount;
      errors.addAll(storeResult.errors);
    } catch (e) {
      print('DEBUG SYNC: syncOfflineStores fatal error: $e');
      errors.add('Store sync: $e');
    }

    // 3. Sync Offline Tasks (Visit Plan Save/Delete, MT Stock Opname, Merchandising)
    try {
      final taskResult = await syncOfflineTasks(force: force);
      totalCount += taskResult.totalCount;
      successCount += taskResult.successCount;
      failureCount += taskResult.failureCount;
      errors.addAll(taskResult.errors);
    } catch (e) {
      print('DEBUG SYNC: syncOfflineTasks fatal error: $e');
      errors.add('Task sync: $e');
    }

    // 4. Sync Daily Journeys (Catat KM Berangkat & Pulang)
    try {
      final journeyResult = await syncDailyJourneys(force: force);
      totalCount += journeyResult.totalCount;
      successCount += journeyResult.successCount;
      failureCount += journeyResult.failureCount;
      errors.addAll(journeyResult.errors);
    } catch (e) {
      print('DEBUG SYNC: syncDailyJourneys fatal error: $e');
      errors.add('Journey sync: $e');
    }

    // 5. Sync Checkins
    try {
      final checkinResult = await syncCheckins(force: force);
      totalCount += checkinResult.totalCount;
      successCount += checkinResult.successCount;
      failureCount += checkinResult.failureCount;
      errors.addAll(checkinResult.errors);
    } catch (e) {
      print('DEBUG SYNC: syncCheckins fatal error: $e');
      errors.add('Checkin sync: $e');
    }

    try {
      final txResult = await syncTransactions(force: force);
      totalCount += txResult.totalCount;
      successCount += txResult.successCount;
      failureCount += txResult.failureCount;
      errors.addAll(txResult.errors);
    } catch (e) {
      print('DEBUG SYNC: syncTransactions fatal error: $e');
      errors.add('Transaction sync: $e');
    }

    try {
      final bannerResult = await syncBannerOrders(force: force);
      totalCount += bannerResult.totalCount;
      successCount += bannerResult.successCount;
      failureCount += bannerResult.failureCount;
      errors.addAll(bannerResult.errors);
    } catch (e) {
      print('DEBUG SYNC: syncBannerOrders fatal error: $e');
      errors.add('Banner order sync: $e');
    }

    return SyncResult(
      totalCount: totalCount,
      successCount: successCount,
      failureCount: failureCount,
      errors: errors,
    );
  }

  Future<SyncResult> syncCheckins({bool force = false}) async {
    print('DEBUG SYNC: syncCheckins started');
    final pending = await (db.select(db.localCheckins)
          ..where((t) => t.syncStatus.equals('pending')))
        .get();

    print('DEBUG SYNC: Found ${pending.length} pending checkins');
    if (pending.isEmpty) return const SyncResult();

    int successCount = 0;
    int failureCount = 0;
    final List<String> errors = [];

    for (final item in pending) {
      final itemKey = 'checkin_${item.id}';

      // Periksa apakah item ini sudah melewati window backoff-nya
      final ready = await _integrity.isReadyForRetry(itemKey, force: force);
      if (!ready) {
        print(
            'DEBUG SYNC: Skipping checkin ${item.id} - masih dalam backoff window');
        continue;
      }

      print(
          'DEBUG SYNC: Processing checkin for store ${item.storeId} at ${item.createdAt}');
      try {
        String selfieUrl = item.selfiePath;

        if (selfieUrl.isNotEmpty && !_isAlreadyServerUrl(selfieUrl)) {
          final uploadedUrl = await _uploadFile(selfieUrl);
          if (uploadedUrl != null) {
            selfieUrl = uploadedUrl;
            await (db.update(db.localCheckins)
                  ..where((t) => t.id.equals(item.id)))
                .write(LocalCheckinsCompanion(selfiePath: Value(selfieUrl)));
          } else {
            throw Exception('Gagal mengupload foto selfie');
          }
        }

        final clientId = 'checkin_${item.id}_${item.createdAt.millisecondsSinceEpoch}';
        final response = await api.client.post('sales/attendance', data: {
          'client_id': clientId,
          'store_id': item.storeId,
          'latitude': item.latitude,
          'longitude': item.longitude,
          'selfie_url': selfieUrl,
          'type': 'CHECKIN',
          'check_time': item.createdAt.toUtc().toIso8601String(),
        });

        if (response.statusCode == 200 || response.statusCode == 201) {
          await (db.update(db.localCheckins)
                ..where((t) => t.id.equals(item.id)))
              .write(const LocalCheckinsCompanion(syncStatus: Value('synced')));
          await _integrity.recordSyncSuccess(itemKey);
          successCount++;
          print('DEBUG SYNC: Checkin synced successfully');
        } else {
          final msg = 'Server checkin return status ${response.statusCode}';
          errors.add('Checkin Toko: $msg');
          failureCount++;
          await _integrity.recordSyncFailure(itemKey, reason: msg);
        }
      } catch (e) {
        print('DEBUG SYNC: Failed to upload check-in: $e');
        String errMsg = e.toString();
        if (e is DioException) {
          final resData = e.response?.data;
          print('DEBUG SYNC ERROR DATA: $resData');
          if (resData is Map && resData.containsKey('error')) {
            errMsg = resData['error'].toString();
          } else if (resData is Map && resData.containsKey('message')) {
            errMsg = resData['message'].toString();
          } else if (e.message != null && e.message!.isNotEmpty) {
            errMsg = e.message!;
          }
        }
        errors.add('Checkin Toko: $errMsg');
        failureCount++;
        await _integrity.recordSyncFailure(itemKey, reason: errMsg);
      }
    }

    return SyncResult(
      totalCount: pending.length,
      successCount: successCount,
      failureCount: failureCount,
      errors: errors,
    );
  }

  Future<SyncResult> syncTransactions({bool force = false}) async {
    print('DEBUG SYNC: syncTransactions started');

    final allCount = await db.select(db.localTransactions).get();
    print('DEBUG SYNC: Total transactions in local DB: ${allCount.length}');

    final pending = await (db.select(db.localTransactions)
          ..where((t) => t.syncStatus.equals('pending')))
        .get();

    print('DEBUG SYNC: Found ${pending.length} pending transactions');
    if (pending.isEmpty) return const SyncResult();

    int successCount = 0;
    int failureCount = 0;
    final List<String> errors = [];

    for (final tr in pending) {
      // Periksa apakah item ini sudah melewati window backoff-nya
      final ready = await _integrity.isReadyForRetry(tr.localId, force: force);
      if (!ready) {
        print(
            'DEBUG SYNC: Skipping ${tr.localId.substring(0, 8)} - masih dalam backoff window');
        failureCount++;
        continue;
      }

      print(
          'DEBUG SYNC: Processing transaction ${tr.localId} for ${tr.storeName}');
      try {
        final items = await (db.select(db.localTransactionItems)
              ..where((t) => t.transactionLocalId.equals(tr.localId)))
            .get();

        final validItems = items
            .where((i) =>
                i.productId.isNotEmpty &&
                i.productId != '00000000-0000-0000-0000-000000000000')
            .map((i) => {
                  'product_id': i.productId,
                  'quantity': i.quantity,
                  'ordered_quantity': i.orderedQuantity,
                  'unit': i.unit,
                  'pieces_per_unit': i.piecesPerUnit,
                  'price': i.price,
                  'discount_percent': i.discountPercent,
                })
            .toList();

        // Coba ekstrak ancillary_items, bonus_items, dan notes dari field notes (karena di-inject JSON dari visit_checkout_page)
        String finalNotes = tr.notes ?? '';
        List<dynamic> ancillaryItems = [];
        List<dynamic> bonusItems = [];
        if (finalNotes.startsWith('{')) {
          try {
            final parsedNotes = jsonDecode(finalNotes) as Map<String, dynamic>;
            if (parsedNotes.containsKey('ancillary_items')) {
              ancillaryItems = parsedNotes['ancillary_items'] as List<dynamic>;
              parsedNotes.remove('ancillary_items');
            }
            if (parsedNotes.containsKey('bonus_items')) {
              bonusItems = parsedNotes['bonus_items'] as List<dynamic>;
              parsedNotes.remove('bonus_items');
            }
            if (parsedNotes.containsKey('return_items')) {
              parsedNotes.remove('return_items'); // not used in payload directly here
            }

            String textNotes = '';
            if (parsedNotes.containsKey('notes')) {
              textNotes = parsedNotes['notes'].toString();
              parsedNotes.remove('notes');
            }

            if (parsedNotes.isNotEmpty) {
              finalNotes = textNotes.isNotEmpty 
                  ? '$textNotes | ${jsonEncode(parsedNotes)}'
                  : jsonEncode(parsedNotes);
            } else {
              finalNotes = textNotes;
            }
          } catch (_) {
            // Abaikan jika bukan JSON
          }
        }

        // If items are empty AND there is no physical receipt, it's just a checkout log (Tidak Ada Transaksi)
        if (items.isEmpty &&
            (tr.receiptPath == null || tr.receiptPath!.isEmpty)) {
          try {
            final response = await api.client.post('sales/attendance', data: {
              'client_id': tr.localId,
              'store_id': tr.storeId,
              'latitude': tr.latitude ?? 0.0,
              'longitude': tr.longitude ?? 0.0,
              'selfie_url': '',
              'type': 'CHECKOUT',
              'check_time': tr.createdAt.toUtc().toIso8601String(),
              'notes': finalNotes,
            });
            if (response.statusCode == 200 || response.statusCode == 201) {
              await (db.update(db.localTransactions)
                    ..where((t) => t.localId.equals(tr.localId)))
                  .write(const LocalTransactionsCompanion(
                      syncStatus: Value('synced')));
              await _integrity.recordSyncSuccess(tr.localId);
              successCount++;
            }
          } catch (e) {
            await _integrity.recordSyncFailure(tr.localId, reason: e.toString());
            failureCount++;
            print('DEBUG SYNC: Failed to upload checkout log: $e');
          }
          continue;
        }

        final payload = {
          'client_id': tr.localId,
          if (tr.companyId.isNotEmpty) 'company_id': tr.companyId,
          'store_id': tr.storeId,
          'latitude': tr.latitude,
          'longitude': tr.longitude,
          'total_amount': tr.totalAmount,
          'paid_amount': (tr.paymentMethod == 'QRIS' ||
                  tr.paymentMethod == 'VA')
              ? 0.0
              : (tr.paymentMethod == 'TEMPO' ? tr.paidAmount : tr.totalAmount),
          'payment_method': tr.paymentMethod,
          'dp_payment_method': tr.dpPaymentMethod,
          'payment_due_date': tr.paymentDueDate != null
              ? '${tr.paymentDueDate!.toIso8601String().substring(0, 10)}T00:00:00.000Z'
              : null,
          'tenor': tr.tenor,
          'bank': tr.paymentBank,
          'notes': finalNotes,
          'store_category': 'TOKO_LAMA',
          'transaction_date': tr.createdAt.toUtc().toIso8601String(),
          'items': validItems,
          if (ancillaryItems.isNotEmpty) 'ancillary_items': ancillaryItems,
          if (bonusItems.isNotEmpty)
            'bonus_items': bonusItems
                .map((b) => {
                      'bonus_inventory_id': b['product_id'],
                      'quantity': b['quantity'],
                      'notes': b['notes'] ?? '',
                    })
                .toList(),
          // MT flag: jika true, SO/faktur dikelola oleh pusat/pabrik
          'is_mt': tr.isMt,
          'is_distributor_order': tr.isDistributorOrder,
          if (tr.factoryCompanyId != null && tr.factoryCompanyId!.isNotEmpty)
            'factory_company_id': tr.factoryCompanyId,
          if (tr.companyBankAccountId != null &&
              tr.companyBankAccountId!.isNotEmpty)
            'company_bank_account_id': tr.companyBankAccountId,
          if (tr.visitId != null && tr.visitId!.isNotEmpty)
            'visit_id': tr.visitId,
        };

        if (tr.receiptNo != null && tr.receiptNo!.isNotEmpty) {
          payload['receipt_no'] = tr.receiptNo!;
        }

        if (tr.receiptPath != null &&
            tr.receiptPath!.isNotEmpty &&
            !tr.receiptPath!.startsWith('http')) {
          final imageUrl = await _uploadFile(tr.receiptPath!);
          if (imageUrl != null) {
            payload['receipt_image_url'] = imageUrl;
            // Update local DB to prevent re-uploading on next retry
            await (db.update(db.localTransactions)
                  ..where((t) => t.localId.equals(tr.localId)))
                .write(
                    LocalTransactionsCompanion(receiptPath: Value(imageUrl)));
          } else {
            print(
                'DEBUG SYNC WARNING: Foto nota fisik (${tr.receiptPath}) tidak dapat diupload, tetap melanjutkan sync transaksi ke server.');
          }
        } else if (tr.receiptPath != null &&
            tr.receiptPath!.startsWith('http')) {
          payload['receipt_image_url'] = tr.receiptPath!;
        }

        if (tr.signaturePath != null && tr.signaturePath!.isNotEmpty) {
          List<String> signatureParts = [];
          if (tr.signaturePath!.contains('data:image/')) {
            final rawParts = tr.signaturePath!.split('data:image/');
            for (var p in rawParts) {
              if (p.trim().isNotEmpty) {
                var clean = p.trim();
                if (clean.endsWith(','))
                  clean = clean.substring(0, clean.length - 1);
                signatureParts.add('data:image/$clean');
              }
            }
          } else {
            signatureParts = tr.signaturePath!.split(',');
          }

          List<String> uploadedUrls = [];
          for (final part in signatureParts) {
            if (part.isNotEmpty) {
              if (part.startsWith('http')) {
                uploadedUrls.add(part);
              } else {
                final url = await _uploadFile(part);
                if (url != null) {
                  uploadedUrls.add(url);
                } else {
                  print(
                      'DEBUG SYNC WARNING: Gagal mengupload foto tanda tangan, tetap melanjutkan.');
                }
              }
            }
          }
          if (uploadedUrls.isNotEmpty) {
            final joinedUrls = uploadedUrls.join(',');
            payload['signature_image_url'] = joinedUrls;
            // Update local DB to prevent re-uploading on next retry
            await (db.update(db.localTransactions)
                  ..where((t) => t.localId.equals(tr.localId)))
                .write(LocalTransactionsCompanion(
                    signaturePath: Value(joinedUrls)));
          }
        }

        // Mark transaction as currently syncing
        await (db.update(db.localTransactions)
              ..where((t) => t.localId.equals(tr.localId)))
            .write(LocalTransactionsCompanion(
          syncStatus: const Value('syncing'),
          lastSyncAttempt: Value(DateTime.now()),
        ));

        final response =
            await api.client.post('sales/transactions', data: payload);

        if (response.statusCode != null &&
            response.statusCode! >= 200 &&
            response.statusCode! < 300) {
          // Jangan menganggap sync gagal hanya karena endpoint mengembalikan
          // 204 atau response tanpa envelope `data`. Ini penting untuk retry:
          // request bisa sudah committed di server, tetapi parsing response
          // yang terlalu ketat membuat transaksi tetap pending di device.
          final responseBody = response.data;
          final dynamic data = responseBody is Map<String, dynamic> &&
                  responseBody['data'] is Map<String, dynamic>
              ? responseBody['data']
              : responseBody is Map<String, dynamic>
                  ? responseBody
                  : <String, dynamic>{};
          final officialNo = data['receipt_no']?.toString();

          await (db.update(db.localTransactions)
                ..where((t) => t.localId.equals(tr.localId)))
              .write(LocalTransactionsCompanion(
            syncStatus: const Value('synced'),
            syncErrorMessage: const Value(null),
            lastSyncAttempt: Value(DateTime.now()),
            receiptNo: officialNo != null && officialNo.isNotEmpty
                ? Value(officialNo)
                : const Value.absent(),
            paymentMethod: data['payment_method'] != null
                ? Value(data['payment_method'].toString())
                : const Value.absent(),
            midtransId: Value(data['midtrans_id']?.toString()),
            midtransQrisUrl: Value(data['midtrans_qris_url']?.toString()),
            midtransVaNumber: Value(data['midtrans_va_number']?.toString()),
            midtransBank: Value(data['midtrans_bank']?.toString()),
            midtransBillKey: Value(data['midtrans_bill_key']?.toString()),
            midtransBillerCode: Value(data['midtrans_biller_code']?.toString()),
          ));

          // Catat keberhasilan sync ke audit trail
          await _integrity.recordSyncSuccess(tr.localId);
          successCount++;
          print('DEBUG SYNC: Successfully synced ${tr.localId} -> $officialNo');
        } else {
          final msg = 'Server returned status ${response.statusCode}';
          errors.add('${tr.storeName}: $msg');
          failureCount++;
          await (db.update(db.localTransactions)
                ..where((t) => t.localId.equals(tr.localId)))
              .write(LocalTransactionsCompanion(
            syncStatus: const Value('failed'),
            syncErrorMessage: Value(msg),
            lastSyncAttempt: Value(DateTime.now()),
          ));
          await _integrity.recordSyncFailure(tr.localId, reason: msg);
          print('DEBUG SYNC: $msg: ${response.data}');
        }
      } catch (e) {
        print('DEBUG SYNC: Failed to sync transaction ${tr.localId}: $e');
        String errMsg = e.toString();

        if (e is DioException) {
          final resData = e.response?.data;
          final statusCode = e.response?.statusCode;
          print('DEBUG SYNC ERROR DATA: $resData');

          // Handle 409 Conflict (DUPLICATE_TRANSACTION / already exists on server)
          if (statusCode == 409 || (resData is Map && resData['code'] == 'DUPLICATE_TRANSACTION')) {
            print('DEBUG SYNC: Transaksi ${tr.localId} sudah tercatat di server (409 Conflict). Menandai sebagai synced.');
            await (db.update(db.localTransactions)
                  ..where((t) => t.localId.equals(tr.localId)))
                .write(LocalTransactionsCompanion(
              syncStatus: const Value('synced'),
              syncErrorMessage: const Value(null),
              lastSyncAttempt: Value(DateTime.now()),
            ));
            await _integrity.recordSyncSuccess(tr.localId);
            successCount++;
            continue;
          }

          if (resData is Map && resData.containsKey('error')) {
            errMsg = resData['error'].toString();
          } else if (resData is Map && resData.containsKey('message')) {
            errMsg = resData['message'].toString();
          } else if (e.message != null && e.message!.isNotEmpty) {
            errMsg = e.message!;
          }
        }

        errors.add('${tr.storeName}: $errMsg');
        failureCount++;

        await (db.update(db.localTransactions)
              ..where((t) => t.localId.equals(tr.localId)))
            .write(LocalTransactionsCompanion(
          syncStatus: const Value('failed'),
          syncErrorMessage: Value(errMsg),
          lastSyncAttempt: Value(DateTime.now()),
        ));

        // Catat kegagalan untuk dijadwalkan retry berikutnya beserta alasannya
        await _integrity.recordSyncFailure(tr.localId, reason: errMsg);
      }
    }

    if (successCount > 0) {
      try {
        print('DEBUG SYNC: Pulling updated sales stock after $successCount synced transactions');
        await pullSalesStock();
      } catch (e) {
        print('DEBUG SYNC: pullSalesStock after transactions batch failed: $e');
      }
    }

    return SyncResult(
      totalCount: pending.length,
      successCount: successCount,
      failureCount: failureCount,
      errors: errors,
    );
  }

  Future<void> pullMasterData() async {
    // 1. Pull Companies
    try {
      final compRes = await api.client.get('companies');
      List? companies;
      if (compRes.data is Map && compRes.data.containsKey('data')) {
        companies = compRes.data['data'] as List?;
      } else if (compRes.data is List) {
        companies = compRes.data as List?;
      }

      if (companies != null && companies.isNotEmpty) {
        await db.batch((batch) {
          batch.insertAll(
              db.companies,
              companies!.map((c) => CompaniesCompanion.insert(
                    id: c['id'] ?? c['ID'] ?? '',
                    name: c['name'] ?? c['Name'] ?? 'Unknown',
                    code: Value(c['code'] ?? c['Code']),
                  )),
              mode: InsertMode.insertOrReplace);
        });
        print('DEBUG: Synced ${companies.length} companies');
      }
    } catch (e) {
      print('DEBUG: pull companies error: $e');
    }

    // 2. Pull Master Stores so local SQLite db.stores is ready for full offline visits
    try {
      final prefs = await SharedPreferences.getInstance();
      final storeApi = StoreApiService(apiClient: api, prefs: prefs);
      final stores = await storeApi.getStoresPaginated(forceRefresh: true);
      print('DEBUG: Synced ${stores.length} stores to local DB');
    } catch (e) {
      print('DEBUG: pull stores error: $e');
    }

    // 3. Pull Master Products (All products from factory to ensure we have the catalog)
    try {
      print('DEBUG: Pulling all master products from factory');
      // Pass company_id=all, scope=sales, type=finished_goods to fetch only sellable goods for mobile
      final prodRes = await api.client.get('factory/products', queryParameters: {
        'company_id': 'all',
        'scope': 'sales',
        'type': 'finished_goods',
      });
      if (prodRes.statusCode == 200) {
        List? products;
        if (prodRes.data is Map && prodRes.data.containsKey('data')) {
          products = prodRes.data['data'] as List?;
        } else if (prodRes.data is List) {
          products = prodRes.data as List?;
        }

        if (products != null && products.isNotEmpty) {
          await db.batch((batch) {
            for (var p in products!) {
              final cat = (p['category'] ?? p['Category'] ?? '').toString().toUpperCase();
              final name = (p['name'] ?? p['Name'] ?? '').toString().toUpperCase();
              
              // Ancillary goods (Barang Sampingan) and Gifts/Merchandise are sellable
              final isAncillaryOrGift = cat.contains('SAMPINGAN') ||
                  cat.contains('HADIAH') ||
                  cat.contains('MERCHANDISE') ||
                  cat.contains('PROMOSI') ||
                  cat.contains('MANGKOK') ||
                  cat.contains('PIRING') ||
                  cat.contains('GELAS') ||
                  name.contains('MANGKOK') ||
                  name.contains('PIRING') ||
                  name.contains('GELAS') ||
                  name.contains('SENDOK') ||
                  name.contains('GARPU');

              // Skip raw materials / packaging items locally just in case (unless they are ancillary goods / gifts)
              if (!isAncillaryOrGift &&
                  (cat.contains('BAHAN') ||
                   cat.contains('RAW') ||
                   cat.contains('KEMASAN') ||
                   cat.contains('WADAH') ||
                   cat.contains('BOTOL') ||
                   cat.contains('BUMBU') ||
                   name.contains('KEDELAI') ||
                   name.contains('GULA') ||
                   name.contains('GARAM') ||
                   name.contains('TEPUNG') ||
                   name.contains('BOTOL KOSONG'))) {
                continue;
              }

              batch.insert(
                  db.products,
                  ProductsCompanion.insert(
                    id: p['id'] ?? p['ID'] ?? '',
                    companyId: p['company_id'] ??
                        p['companyId'] ??
                        p['CompanyID'] ??
                        '',
                    name: p['name'] ?? p['Name'] ?? 'Unknown',
                    sellingPrice: (p['selling_price'] ??
                            p['sellingPrice'] ??
                            p['SellingPrice'] ??
                            0 as num)
                        .toDouble(),
                    dbp: Value(((p['dbp'] ?? p['DBP'] ?? 0) as num).toDouble()),
                    wbp: Value(((p['wbp'] ?? p['WBP'] ?? 0) as num).toDouble()),
                    rbp: Value(((p['rbp'] ?? p['RBP'] ?? 0) as num).toDouble()),
                    costPrice: Value(((p['cost_price'] ?? p['costPrice'] ?? 0) as num).toDouble()),
                    sku: Value(p['sku'] ?? p['SKU']),
                    unit: Value(p['unit'] ?? p['Unit']),
                    pcsPerUnit: Value(p['pcs_per_unit'] ??
                        p['pcsPerUnit'] ??
                        p['PcsPerUnit'] ??
                        1),
                    category: Value(p['category'] ?? p['Category']),
                    // We'll update stock in the next step
                    warehouseStock: const Value(0),
                    imageUrl: Value(p['image_url'] ?? p['imageUrl']),
                  ),
                  mode: InsertMode.insertOrReplace);
            }
          });
          print('DEBUG: Synced ${products.length} master products');
        }
      }
    } catch (e) {
      print('DEBUG: Error pulling master products: $e');
    }

    // 4. Pull Products from Warehouse Stock (To update stock quantities)
    try {
      print('DEBUG: Pulling products from warehouse/stock');
      final whRes = await api.client.get('warehouse/stock');
      List? whStocks;
      if (whRes.data is Map && whRes.data.containsKey('data')) {
        whStocks = whRes.data['data'] as List?;
      } else if (whRes.data is List) {
        whStocks = whRes.data as List?;
      }

      if (whStocks != null && whStocks.isNotEmpty) {
        // Aggregate stock by product ID in case there are multiple batches
        final Map<String, int> aggregatedStock = {};
        final Map<String, dynamic> productDetails = {};
        for (var item in whStocks) {
          final p = item['product'] ?? item['Product'];
          if (p == null) continue;
          final id = p['id'] ?? p['ID'] ?? '';
          if (id.isEmpty) continue;

          final int qty =
              (item['quantity'] ?? item['Quantity'] ?? 0 as num).toInt();
          final int currentStock = aggregatedStock[id] ?? 0;
          aggregatedStock[id] = currentStock + qty;

          if (!productDetails.containsKey(id)) {
            productDetails[id] = p;
          }
        }

        // Update stock for existing products and insert new products from warehouse stock
        await db.batch((batch) {
          for (final entry in aggregatedStock.entries) {
            final p = productDetails[entry.key];
            if (p != null) {
              batch.insert(
                  db.products,
                  ProductsCompanion.insert(
                    id: p['id'] ?? p['ID'] ?? '',
                    companyId: p['company_id'] ??
                        p['companyId'] ??
                        p['CompanyID'] ??
                        '',
                    name: p['name'] ?? p['Name'] ?? 'Unknown',
                    sellingPrice: (p['selling_price'] ??
                            p['sellingPrice'] ??
                            p['SellingPrice'] ??
                            0 as num)
                        .toDouble(),
                    dbp: Value(((p['dbp'] ?? p['DBP'] ?? 0) as num).toDouble()),
                    wbp: Value(((p['wbp'] ?? p['WBP'] ?? 0) as num).toDouble()),
                    rbp: Value(((p['rbp'] ?? p['RBP'] ?? 0) as num).toDouble()),
                    costPrice: Value(((p['cost_price'] ?? p['costPrice'] ?? 0) as num).toDouble()),
                    sku: Value(p['sku'] ?? p['SKU']),
                    unit: Value(p['unit'] ?? p['Unit']),
                    pcsPerUnit: Value(p['pcs_per_unit'] ??
                        p['pcsPerUnit'] ??
                        p['PcsPerUnit'] ??
                        1),
                    category: Value(p['category'] ?? p['Category']),
                    warehouseStock: Value(entry.value),
                    imageUrl: Value(p['image_url'] ?? p['imageUrl']),
                  ),
                  mode: InsertMode.insertOrReplace);
            } else {
              batch.update(
                db.products,
                ProductsCompanion(warehouseStock: Value(entry.value)),
                where: (t) => t.id.equals(entry.key),
              );
            }
          }
        });

        print('DEBUG: Synced ${whStocks.length} total stock rows');
        print(
            'DEBUG: Found ${aggregatedStock.length} UNIQUE products in warehouse/stock!');
      }
    } catch (e) {
      print('DEBUG: Error pulling from warehouse/stock: $e');
    }

    // 5. Pull Sales Stock (Bronjong)
    await pullSalesStock();

    // 6. Warm Bank Accounts & Today's Visit Plans cache for offline readiness
    try {
      final bankRes = await api.client.get('bank-accounts');
      final rawBanks = bankRes.data['data'] ?? bankRes.data ?? [];
      if (rawBanks is List) {
        final prefs = await SharedPreferences.getInstance();
        await prefs.setString('cached_bank_accounts', jsonEncode(rawBanks));
      }
    } catch (_) {}

    try {
      final todayStr = DateTime.now().toIso8601String().split('T')[0];
      final planRes = await api.client.get('sales/visit-plans', queryParameters: {'date': todayStr});
      final rawPlans = planRes.data['data'] ?? planRes.data ?? [];
      if (rawPlans is List && rawPlans.isNotEmpty) {
        final prefs = await SharedPreferences.getInstance();
        await prefs.setString('cached_visit_plans_$todayStr', jsonEncode(rawPlans));
      }
    } catch (_) {}
  }

  Future<void> pullSalesStock() async {
    try {
      // 1. Get Profile to find employee ID
      final profileRes = await api.client.get('employees/profile');
      final profileData =
          profileRes.data is Map && profileRes.data.containsKey('data')
              ? profileRes.data['data']
              : profileRes.data;

      if (profileData == null) {
        print('DEBUG: Profile data is null, skipping sales stock sync');
        return;
      }

      final employeeId = profileData['id'] ?? profileData['ID'];
      if (employeeId == null) {
        print('DEBUG: Employee ID not found in profile');
        return;
      }

      // 2. Pull Sales Stock (Bronjong)
      final stockRes =
          await api.client.get('sales-transfers/stock/$employeeId');
      List? stocks;
      if (stockRes.data is Map && stockRes.data.containsKey('data')) {
        stocks = stockRes.data['data'] as List?;
      } else if (stockRes.data is List) {
        stocks = stockRes.data as List?;
      }

      if (stocks != null && stocks.isNotEmpty) {
        await db.batch((batch) {
          for (var s in stocks!) {
            // Insert stock
            batch.insert(
                db.salesStock,
                SalesStockCompanion.insert(
                  productId:
                      s['product_id'] ?? s['productId'] ?? s['ProductID'] ?? '',
                  quantity: s['quantity'] ?? s['Quantity'] ?? 0,
                  updatedAt: DateTime.now(),
                ),
                mode: InsertMode.insertOrReplace);

            final p = s['product'] ?? s['Product'];
            if (p != null) {
              batch.insert(
                  db.products,
                  ProductsCompanion.insert(
                    id: p['id'] ?? p['ID'] ?? '',
                    companyId: p['company_id'] ??
                        p['companyId'] ??
                        p['CompanyID'] ??
                        '',
                    name: p['name'] ?? p['Name'] ?? 'Unknown',
                    sellingPrice: (p['selling_price'] ??
                            p['sellingPrice'] ??
                            p['SellingPrice'] ??
                            0 as num)
                        .toDouble(),
                    dbp: Value(((p['dbp'] ?? p['DBP'] ?? 0) as num).toDouble()),
                    wbp: Value(((p['wbp'] ?? p['WBP'] ?? 0) as num).toDouble()),
                    rbp: Value(((p['rbp'] ?? p['RBP'] ?? 0) as num).toDouble()),
                    costPrice: Value(((p['cost_price'] ?? p['costPrice'] ?? 0) as num).toDouble()),
                    sku: Value(p['sku'] ?? p['SKU']),
                    unit: Value(p['unit'] ?? p['Unit']),
                    category: Value(p['category'] ?? p['Category']),
                    imageUrl: Value(p['image_url'] ?? p['imageUrl']),
                    warehouseStock: Value(p['warehouse_stock'] ??
                        p['warehouseStock'] ??
                        p['WarehouseStock'] ??
                        0),
                  ),
                  mode: InsertMode.insertOrReplace);
            }
          }
        });
        print('DEBUG: Synced ${stocks.length} sales stocks and their products');
      } else {
        print('DEBUG: No sales stocks found for employee $employeeId');
      }
    } catch (e) {
      print('DEBUG: pullSalesStock error: $e');
    }
  }

  Future<String?> _uploadFile(String filePathOrBase64) async {
    if (_isAlreadyServerUrl(filePathOrBase64)) {
      return filePathOrBase64.trim();
    }
    try {
      final isJpg = filePathOrBase64.startsWith('data:image/jpeg') ||
          filePathOrBase64.toLowerCase().endsWith('.jpg') ||
          filePathOrBase64.toLowerCase().endsWith('.jpeg');
      final ext = isJpg ? 'jpg' : 'png';
      final fileName = 'upload_${DateTime.now().millisecondsSinceEpoch}.$ext';

      late MultipartFile multipartFile;
      if (filePathOrBase64.startsWith('data:')) {
        if (!filePathOrBase64.contains(',')) {
          print(
              'DEBUG SYNC: Skipping corrupt Base64 data with missing payload: $filePathOrBase64');
          return null;
        }
        print('DEBUG SYNC: Uploading BASE64 image...');
        final base64String = filePathOrBase64.split(',').last;
        final bytes = base64Decode(base64String);
        multipartFile = MultipartFile.fromBytes(bytes, filename: fileName);
      } else {
        if (kIsWeb) {
          print(
              'DEBUG SYNC ERROR: Cannot use fromFile on Web for path: $filePathOrBase64');
          return null;
        }
        final file = File(filePathOrBase64);
        if (!file.existsSync()) {
          print(
              'DEBUG SYNC WARNING: File lokal tidak ditemukan di path: $filePathOrBase64');
          return null;
        }
        print('DEBUG SYNC: Uploading FILE image from path: $filePathOrBase64');
        multipartFile =
            await MultipartFile.fromFile(filePathOrBase64, filename: fileName);
      }

      final formData = FormData.fromMap({
        'file': multipartFile,
      });

      // Buat instance Dio khusus untuk upload dengan timeout yang jauh lebih besar (5 menit)
      // Ini mencegah timeout di koneksi 3G yang lambat saat mengupload gambar besar
      final uploadDio = Dio(api.client.options.copyWith(
        connectTimeout: const Duration(minutes: 5),
        receiveTimeout: const Duration(minutes: 5),
        sendTimeout: const Duration(minutes: 5),
      ));
      // Sertakan interceptors (auth token, dll)
      uploadDio.interceptors.addAll(api.client.interceptors);

      final response = await uploadDio.post('sales/upload', data: formData);
      if (response.statusCode == 200 || response.statusCode == 201) {
        final url = response.data['url']?.toString();
        print('DEBUG SYNC: Upload successful! URL: $url');
        return url;
      }
      print('DEBUG SYNC: Upload failed with status ${response.statusCode}');
      return null;
    } catch (e) {
      print('DEBUG SYNC: _uploadFile error: $e');
      return null;
    }
  }

  Future<SyncResult> syncOfflineStores({bool force = false}) async {
    print('DEBUG SYNC: syncOfflineStores started');
    final pending = await (db.select(db.stores)
          ..where((t) => t.isOffline.equals(true) | t.syncStatus.equals('pending')))
        .get();

    print('DEBUG SYNC: Found ${pending.length} pending offline stores');
    if (pending.isEmpty) return const SyncResult();

    int successCount = 0;
    int failureCount = 0;
    final List<String> errors = [];

    for (final item in pending) {
      final itemKey = 'store_${item.id}';
      final ready = await _integrity.isReadyForRetry(itemKey, force: force);
      if (!ready) {
        print('DEBUG SYNC: Skipping store ${item.id} - masih dalam backoff window');
        continue;
      }

      print('DEBUG SYNC: Processing store ${item.name} (${item.id})');
      try {
        String? photoUrl = item.photoUrl;
        List<int>? localPhotoBytes;

        if (item.localPhotoPath != null && item.localPhotoPath!.isNotEmpty) {
          final path = item.localPhotoPath!;
          if (path.startsWith('data:image/')) {
            try {
              final commaIdx = path.indexOf(',');
              final b64 = commaIdx != -1 ? path.substring(commaIdx + 1) : path;
              localPhotoBytes = base64Decode(b64);
            } catch (_) {}
          } else if (!kIsWeb) {
            try {
              final file = File(path);
              if (file.existsSync()) {
                localPhotoBytes = await file.readAsBytes();
              }
            } catch (_) {}
          }

          // Also try _uploadFile if photoUrl is still empty
          if (photoUrl == null || photoUrl.isEmpty) {
            final uploaded = await _uploadFile(path);
            if (uploaded != null && uploaded.isNotEmpty) {
              photoUrl = uploaded;
            }
          }
        }

        List<int> visitDays = [];
        if (item.visitDays != null && item.visitDays!.isNotEmpty) {
          try {
            visitDays = item.visitDays!
                .split(',')
                .map((s) => int.tryParse(s.trim()))
                .whereType<int>()
                .toList();
          } catch (_) {}
        }

        final payloadMap = <String, dynamic>{
          'id': item.id,
          'name': item.name,
          if (item.ownerName != null) 'owner_name': item.ownerName,
          if (item.phoneNumber != null) 'phone_number': item.phoneNumber,
          if (item.address != null) 'address': item.address,
          if (item.latitude != null) 'latitude': item.latitude,
          if (item.longitude != null) 'longitude': item.longitude,
          if (photoUrl != null && photoUrl.isNotEmpty) 'photo_url': photoUrl,
          if (item.province != null && item.province!.isNotEmpty)
            'province': item.province,
          if (item.regency != null && item.regency!.isNotEmpty)
            'regency': item.regency,
          if (item.district != null && item.district!.isNotEmpty)
            'district': item.district,
          if (item.village != null && item.village!.isNotEmpty)
            'village': item.village,
          if (item.visitFrequency != null && item.visitFrequency!.isNotEmpty)
            'visit_frequency': item.visitFrequency,
          if (visitDays.isNotEmpty) 'visit_days': visitDays,
          if (item.assignedEmployeeId != null &&
              item.assignedEmployeeId!.isNotEmpty)
            'assigned_employee_id': item.assignedEmployeeId,
          if (item.companyId != null && item.companyId!.isNotEmpty)
            'company_id': item.companyId,
        };

        dynamic requestData = payloadMap;
        if (localPhotoBytes != null && localPhotoBytes.isNotEmpty) {
          final formData = FormData.fromMap({
            ...payloadMap,
            'photo': MultipartFile.fromBytes(
              localPhotoBytes,
              filename: 'store_${item.id}.jpg',
            ),
          });
          if (visitDays.isNotEmpty) {
            formData.fields.removeWhere((f) => f.key == 'visit_days');
            for (final day in visitDays) {
              formData.fields.add(MapEntry('visit_days[]', day.toString()));
            }
          }
          requestData = formData;
        }

        Response response;
        try {
          response = await api.client.post('stores', data: requestData);
        } on DioException catch (dioErr) {
          // If store already exists and we are updating an existing store edited offline, try PUT
          if (dioErr.response?.statusCode == 409) {
            response = await api.client.put('stores/${item.id}', data: requestData);
          } else {
            rethrow;
          }
        }

        if (response.statusCode == 200 || response.statusCode == 201) {
          // If backend returned a canonical photo_url (e.g. from multipart upload or existing store), use it
          try {
            final respData = response.data;
            if (respData is Map && respData['data'] is Map) {
              final serverPhoto = respData['data']['photo_url']?.toString();
              if (serverPhoto != null && serverPhoto.isNotEmpty) {
                photoUrl = serverPhoto;
              }
            }
          } catch (_) {}

          // If store already existed idempotently on server (200 OK) and we still have localPhotoBytes that weren't saved on POST, update via PUT
          if (response.statusCode == 200 &&
              (photoUrl == null || photoUrl.isEmpty) &&
              localPhotoBytes != null &&
              localPhotoBytes.isNotEmpty) {
            try {
              final putRes =
                  await api.client.put('stores/${item.id}', data: requestData);
              if (putRes.data is Map && putRes.data['data'] is Map) {
                final serverPhoto = putRes.data['data']['photo_url']?.toString();
                if (serverPhoto != null && serverPhoto.isNotEmpty) {
                  photoUrl = serverPhoto;
                }
              }
            } catch (_) {}
          }

          await (db.update(db.stores)..where((t) => t.id.equals(item.id))).write(
            StoresCompanion(
              isOffline: const Value(false),
              syncStatus: const Value('synced'),
              photoUrl: Value(photoUrl),
              localPhotoPath: const Value(null),
              syncErrorMessage: const Value(null),
            ),
          );

          // Update SharedPreferences cache so the synced status is reflected
          try {
            final prefs = await SharedPreferences.getInstance();
            final keys = prefs
                .getKeys()
                .where((k) => k.startsWith('cached_stores_list'))
                .toList();
            for (final key in keys) {
              final cachedStr = prefs.getString(key);
              if (cachedStr != null) {
                final List<dynamic> decoded = jsonDecode(cachedStr);
                final updatedList = decoded.map((j) {
                  if (j is Map<String, dynamic> && j['id'] == item.id) {
                    j['is_offline'] = false;
                    j['sync_status'] = 'synced';
                    if (photoUrl != null && photoUrl.isNotEmpty) {
                      j['photo_url'] = photoUrl;
                    }
                    j.remove('local_photo_path');
                  }
                  return j;
                }).toList();
                await prefs.setString(key, jsonEncode(updatedList));
              }
            }
          } catch (_) {}

          // Delete temporary local photo file if synced and file exists
          if (item.localPhotoPath != null &&
              !item.localPhotoPath!.startsWith('data:image/') &&
              !kIsWeb) {
            try {
              final file = File(item.localPhotoPath!);
              if (file.existsSync()) {
                await file.delete();
              }
            } catch (_) {}
          }

          await _integrity.recordSyncSuccess(itemKey);
          successCount++;
          print(
              'DEBUG SYNC: Store ${item.name} (${item.id}) synced successfully');
        } else {
          throw Exception(
              'Server mengembalikan kode status ${response.statusCode}');
        }
      } catch (e) {
        print('DEBUG SYNC: Gagal menyinkronkan toko ${item.name}: $e');
        failureCount++;
        errors.add('Toko ${item.name}: $e');
        await _integrity.recordSyncFailure(itemKey);
        await (db.update(db.stores)..where((t) => t.id.equals(item.id))).write(
          StoresCompanion(
            syncStatus: const Value('failed'),
            syncErrorMessage: Value(e.toString()),
          ),
        );
      }
    }

    return SyncResult(
      totalCount: pending.length,
      successCount: successCount,
      failureCount: failureCount,
      errors: errors,
    );
  }

  Future<SyncResult> syncBannerOrders({bool force = false}) async {
    print('DEBUG SYNC: syncBannerOrders started');
    int totalCount = 0;
    int successCount = 0;
    int failureCount = 0;
    final List<String> errors = [];

    try {
      final prefs = await SharedPreferences.getInstance();
      final String? existingJson = prefs.getString('offline_banner_orders');
      if (existingJson == null || existingJson.isEmpty) {
        print('DEBUG SYNC: No pending offline banner orders');
        return const SyncResult();
      }

      List<dynamic> list = jsonDecode(existingJson);
      totalCount = list.length;
      print('DEBUG SYNC: Found ${list.length} pending banner orders');

      // Persist a server UUID for legacy queued orders before sending.
      for (final item in list) {
        if (item is Map &&
            (item['id'] == null ||
                item['id'].toString().startsWith('OFFLINE-')) &&
            item['sync_id'] == null) {
          item['sync_id'] = const Uuid().v4();
        }
      }
      await prefs.setString('offline_banner_orders', jsonEncode(list));

      List<dynamic> remaining = [];
      for (var item in list) {
        final bannerKey = 'banner_${item['sync_id'] ?? item['id'] ?? item['shop_id'] ?? item['created_at'] ?? item.hashCode}';

        // Check if this item is ready for retry
        final ready = await _integrity.isReadyForRetry(bannerKey, force: force);
        if (!ready) {
          print('DEBUG SYNC: Skipping $bannerKey - masih dalam backoff window');
          failureCount++;
          remaining.add(item);
          continue;
        }

        try {
          final payload = Map<String, dynamic>.from(item);
          if (payload['sync_id'] != null) {
            payload['id'] = payload.remove('sync_id');
          }
          final attachmentPath = payload.remove('offline_attachment_path')?.toString();
          final attachmentBase64 = payload.remove('offline_attachment_base64')?.toString();
          final attachmentName = payload.remove('offline_attachment_name')?.toString() ?? 'attachment.jpg';
          if (payload['status'] != null && payload['status'].toString().contains('OFFLINE')) {
            payload.remove('status');
          }
          dynamic requestData = payload;
          if (attachmentPath != null || attachmentBase64 != null) {
            final bytes = attachmentBase64 != null
                ? base64Decode(attachmentBase64)
                : await File(attachmentPath!).readAsBytes();
            requestData = FormData.fromMap({
              ...payload,
              'attachment': MultipartFile.fromBytes(bytes, filename: attachmentName),
            });
          }
          await api.client.post('branding/banners', data: requestData);
          print('DEBUG SYNC: Offline banner order synced successfully');
          if (attachmentPath != null && !kIsWeb) {
            try {
              await File(attachmentPath).delete();
            } catch (_) {}
          }
          await _integrity.recordSyncSuccess(bannerKey);
          successCount++;
        } catch (e) {
          print('DEBUG SYNC: Failed to sync offline banner order: $e');
          String errMsg = 'Gagal sinkronisasi order spanduk';

          if (e is DioException) {
            final resData = e.response?.data;
            final statusCode = e.response?.statusCode;

            if (resData is Map && resData.containsKey('error')) {
              errMsg = resData['error'].toString();
            } else if (resData is Map && resData.containsKey('message')) {
              errMsg = resData['message'].toString();
            } else if (statusCode != null) {
              errMsg = 'Server mengembalikan status $statusCode';
            } else if (e.message != null && e.message!.isNotEmpty) {
              errMsg = e.message!;
            }
          } else {
            errMsg = e.toString().replaceFirst('Exception: ', '');
          }

          failureCount++;
          errors.add('Order Spanduk: $errMsg');
          await _integrity.recordSyncFailure(bannerKey, reason: errMsg);
          remaining.add(item); // Keep in queue to retry later
        }
      }

      if (remaining.isEmpty) {
        await prefs.remove('offline_banner_orders');
      } else {
        await prefs.setString('offline_banner_orders', jsonEncode(remaining));
      }
    } catch (err) {
      print('DEBUG SYNC: error in syncBannerOrders: $err');
      errors.add('Banner order sync: $err');
    }

    return SyncResult(
      totalCount: totalCount,
      successCount: successCount,
      failureCount: failureCount,
      errors: errors,
    );
  }

  Future<SyncResult> syncOfflineTasks({bool force = false}) async {
    print('DEBUG SYNC: syncOfflineTasks started');
    try {
      final prefs = await SharedPreferences.getInstance();
      final fallbackQueue = prefs.getStringList('offline_visit_plan_queue') ?? [];
      if (fallbackQueue.isNotEmpty) {
        final List<String> remaining = [];
        for (final itemJson in fallbackQueue) {
          try {
            final item = jsonDecode(itemJson) as Map<String, dynamic>;
            await db.into(db.offlineTasks).insertOnConflictUpdate(
                  OfflineTasksCompanion.insert(
                    id: (item['id'] ?? DateTime.now().millisecondsSinceEpoch.toString()).toString(),
                    type: (item['type'] ?? 'VISIT_PLAN_SAVE').toString(),
                    payload: jsonEncode(item['payload'] ?? {}),
                    createdAt: DateTime.tryParse((item['created_at'] ?? '').toString()) ?? DateTime.now(),
                    status: const Value('pending'),
                  ),
                );
          } catch (_) {
            remaining.add(itemJson);
          }
        }
        await prefs.setStringList('offline_visit_plan_queue', remaining);
      }
    } catch (_) {}

    final pending = await (db.select(db.offlineTasks)
          ..where((t) =>
              t.status.equals('pending') | t.status.equals('failed'))
          ..orderBy([(t) => OrderingTerm(expression: t.createdAt, mode: OrderingMode.asc)]))
        .get();

    if (pending.isEmpty) return const SyncResult();
    print('DEBUG SYNC: Found ${pending.length} pending offline tasks');

    int successCount = 0;
    int failureCount = 0;
    final List<String> errors = [];

    for (final task in pending) {
      final itemKey = 'task_${task.id}';
      final ready = await _integrity.isReadyForRetry(itemKey, force: force);
      if (!ready) continue;

      try {
        final Map<String, dynamic> payload = jsonDecode(task.payload);

        if (task.type == 'VISIT_PLAN_SAVE') {
          final response = await api.client.post('sales/visit-plans/manual', data: payload);
          if (response.statusCode == 200 || response.statusCode == 201) {
            final rawDate = (payload['plan_date'] ?? payload['date'] ?? '').toString();
            final planDate = rawDate.split('T').first;
            if (planDate.isNotEmpty) {
              final prefs = await SharedPreferences.getInstance();
              await prefs.remove('offline_plan_pending_stores_$planDate');
            }
            await (db.delete(db.offlineTasks)..where((t) => t.id.equals(task.id))).go();
            await _integrity.recordSyncSuccess(itemKey);
            successCount++;
          } else {
            throw Exception('Server status ${response.statusCode}');
          }
        } else if (task.type == 'VISIT_PLAN_DELETE') {
          final storeId = (payload['store_id'] ?? '').toString();
          final response = await api.client.delete(
            'sales/visit-plans/$storeId',
            queryParameters: {
              'date': payload['date'],
            },
          );
          if (response.statusCode == 200 || response.statusCode == 204) {
            await (db.delete(db.offlineTasks)..where((t) => t.id.equals(task.id))).go();
            await _integrity.recordSyncSuccess(itemKey);
            successCount++;
          } else {
            throw Exception('Server status ${response.statusCode}');
          }
        } else if (task.type == 'MT_STOCK_OPNAME') {
          final response = await api.client.post('modern-trade/store-stocks/adjust', data: payload);
          if (response.statusCode == 200 || response.statusCode == 201) {
            await (db.delete(db.offlineTasks)..where((t) => t.id.equals(task.id))).go();
            await _integrity.recordSyncSuccess(itemKey);
            successCount++;
          } else {
            throw Exception('Server status ${response.statusCode}');
          }
        } else if (task.type == 'MERCHANDISING') {
          String? photoUrl = payload['photo_url']?.toString() ?? payload['display_photo_url']?.toString();
          if ((photoUrl == null || photoUrl.isEmpty || !photoUrl.startsWith('http') && !photoUrl.startsWith('/')) &&
              task.filePath != null &&
              task.filePath!.isNotEmpty) {
            final file = File(task.filePath!);
            if (file.existsSync()) {
              final formData = FormData.fromMap({
                'file': await MultipartFile.fromFile(task.filePath!, filename: 'display_photo.jpg'),
              });
              final uploadRes = await api.client.post('modern-trade/merchandising/upload-photo', data: formData);
              if (uploadRes.statusCode == 200 || uploadRes.statusCode == 201) {
                photoUrl = (uploadRes.data['url'] ?? '').toString().trim();
                payload['photo_url'] = photoUrl;
              } else {
                throw Exception('Gagal mengupload foto display merchandising');
              }
            }
          }
          final response = await api.client.post('modern-trade/merchandising', data: payload);
          if (response.statusCode == 200 || response.statusCode == 201) {
            await (db.delete(db.offlineTasks)..where((t) => t.id.equals(task.id))).go();
            await _integrity.recordSyncSuccess(itemKey);
            successCount++;
          } else {
            throw Exception('Server status ${response.statusCode}');
          }
        } else {
          // Unknown task type - remove to avoid infinite loop
          await (db.delete(db.offlineTasks)..where((t) => t.id.equals(task.id))).go();
        }
      } catch (e) {
        print('DEBUG SYNC: Failed offline task ${task.id} (${task.type}): $e');
        failureCount++;
        errors.add('Task ${task.type}: $e');
        await _integrity.recordSyncFailure(itemKey, reason: e.toString());
        await (db.update(db.offlineTasks)..where((t) => t.id.equals(task.id))).write(
          OfflineTasksCompanion(
            status: const Value('failed'),
            retryCount: Value(task.retryCount + 1),
          ),
        );
      }
    }

    return SyncResult(
      totalCount: pending.length,
      successCount: successCount,
      failureCount: failureCount,
      errors: errors,
    );
  }

  Future<SyncResult> syncDailyJourneys({bool force = false}) async {
    print('DEBUG SYNC: syncDailyJourneys started');
    final pending = await (db.select(db.salesDailyJourneys)
          ..where((t) => t.isSynced.equals(false)))
        .get();

    if (pending.isEmpty) return const SyncResult();
    print('DEBUG SYNC: Found ${pending.length} pending daily journeys');

    int successCount = 0;
    int failureCount = 0;
    final List<String> errors = [];

    for (final j in pending) {
      final itemKey = 'journey_${j.id}';
      final ready = await _integrity.isReadyForRetry(itemKey, force: force);
      if (!ready) continue;

      try {
        String startPhotoUrl = j.startPhotoUrl;
        if (startPhotoUrl.isNotEmpty && !_isAlreadyServerUrl(startPhotoUrl)) {
          final bool localFileMissing = !kIsWeb &&
              !startPhotoUrl.startsWith('data:') &&
              !File(startPhotoUrl).existsSync();
          if (localFileMissing) {
            // File fisik sudah tidak ada di disk (misal dibersihkan OS), jangan blokir pengiriman KM
            startPhotoUrl = '';
          } else {
            final uploaded = await _uploadFile(startPhotoUrl);
            if (uploaded != null) {
              startPhotoUrl = uploaded;
              await (db.update(db.salesDailyJourneys)
                    ..where((t) => t.id.equals(j.id)))
                  .write(SalesDailyJourneysCompanion(startPhotoUrl: Value(startPhotoUrl)));
            } else {
              throw Exception('Gagal mengupload foto odometer berangkat');
            }
          }
        }

        // Sync Start Journey (ignore 400/409 if already started on server)
        try {
          await api.client.post('sales/journeys/start', data: {
            'start_km': j.startKm,
            'start_photo_url': startPhotoUrl,
            'start_location': j.startLocation,
          });
        } on DioException catch (e) {
          final code = e.response?.statusCode;
          if (code != 400 && code != 409) {
            rethrow;
          }
        }

        // If journey also ENDED locally, sync End Journey
        if (j.status == 'ENDED' && j.endKm != null) {
          String endPhotoUrl = j.endPhotoUrl ?? '';
          if (endPhotoUrl.isNotEmpty && !_isAlreadyServerUrl(endPhotoUrl)) {
            final bool localFileMissing = !kIsWeb &&
                !endPhotoUrl.startsWith('data:') &&
                !File(endPhotoUrl).existsSync();
            if (localFileMissing) {
              endPhotoUrl = '';
            } else {
              final uploaded = await _uploadFile(endPhotoUrl);
              if (uploaded != null) {
                endPhotoUrl = uploaded;
                await (db.update(db.salesDailyJourneys)
                      ..where((t) => t.id.equals(j.id)))
                    .write(SalesDailyJourneysCompanion(endPhotoUrl: Value(endPhotoUrl)));
              } else {
                throw Exception('Gagal mengupload foto odometer pulang');
              }
            }
          }

          try {
            await api.client.post('sales/journeys/end', data: {
              'end_km': j.endKm,
              'end_photo_url': endPhotoUrl,
              'end_location': j.endLocation ?? '',
            });
          } on DioException catch (e) {
            final code = e.response?.statusCode;
            if (code != 400 && code != 409) {
              rethrow;
            }
          }
        }

        await (db.update(db.salesDailyJourneys)
              ..where((t) => t.id.equals(j.id)))
            .write(const SalesDailyJourneysCompanion(isSynced: Value(true)));
        await _integrity.recordSyncSuccess(itemKey);
        successCount++;
      } catch (e) {
        print('DEBUG SYNC: Failed daily journey ${j.id}: $e');
        failureCount++;
        final cleanErr = e.toString().replaceAll('Exception: ', '');
        errors.add('Perjalanan KM: $cleanErr');
        await _integrity.recordSyncFailure(itemKey, reason: cleanErr);
      }
    }

    return SyncResult(
      totalCount: pending.length,
      successCount: successCount,
      failureCount: failureCount,
      errors: errors,
    );
  }
}
