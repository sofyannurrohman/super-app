import 'dart:async';
import 'package:internet_connection_checker_plus/internet_connection_checker_plus.dart';
import 'package:drift/drift.dart' as drift;
import 'package:flutter/foundation.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/network/network_info.dart';

class SyncManager {
  static final SyncManager _instance = SyncManager._internal();
  factory SyncManager() => _instance;
  
  late AppDatabase _db;
  late ApiClient _apiClient;
  bool _isInitialized = false;
  bool _isSyncing = false;
  StreamSubscription? _connectivitySubscription;

  SyncManager._internal();

  void initialize(AppDatabase db, ApiClient apiClient) {
    if (_isInitialized) return;
    _db = db;
    _apiClient = apiClient;
    _isInitialized = true;

    // Listen to network changes
    _connectivitySubscription = NetworkInfo().onStatusChange.listen((status) {
      if (status == InternetStatus.connected) {
        startSync();
      }
    });
  }

  void dispose() {
    _connectivitySubscription?.cancel();
  }

  Future<void> startSync() async {
    if (_isSyncing || !_isInitialized) return;
    
    // Check actual connection first
    final hasInternet = await NetworkInfo().hasInternetAccess;
    if (!hasInternet) return;

    _isSyncing = true;
    try {
      debugPrint('[SyncManager] Memulai sinkronisasi background...');
      // 1. Ambil semua antrean dari database yang berstatus 'pending' (berurutan/FIFO)
      final pendingTasks = await (_db.select(_db.offlineTasks)
            ..where((t) => t.status.equals('pending'))
            ..orderBy([(t) => drift.OrderingTerm.asc(t.createdAt)]))
          .get();

      if (pendingTasks.isEmpty) {
        debugPrint('[SyncManager] Tidak ada antrean. Sinkronisasi selesai.');
        _isSyncing = false;
        return;
      }

      for (final task in pendingTasks) {
        // Tandai sebagai sedang diproses
        await (_db.update(_db.offlineTasks)..where((t) => t.id.equals(task.id)))
            .write(const OfflineTasksCompanion(status: drift.Value('processing')));

        bool success = false;
        try {
          // TODO: Implementasi logika spesifik berdasarkan task.type
          // Contoh:
          // if (task.type == 'ATTENDANCE') { ... }
          // else if (task.type == 'ORDER') { ... }
          
          // Kirim task.id sebagai idempotency key di header/body
          // Tunggu respons 200 OK
          
          // Jika sukses, ubah success = true
          success = true; // Simulasi
        } catch (e) {
          debugPrint('[SyncManager] Error saat sinkronisasi task ${task.id}: $e');
        }

        if (success) {
          // Hapus task dari antrean jika berhasil (atau direspons 200 OK karena idempotency)
          await (_db.delete(_db.offlineTasks)..where((t) => t.id.equals(task.id))).go();
        } else {
          // Kembalikan status ke pending dan tambah retryCount
          await (_db.update(_db.offlineTasks)..where((t) => t.id.equals(task.id))).write(
            OfflineTasksCompanion(
              status: const drift.Value('pending'),
              retryCount: drift.Value(task.retryCount + 1),
            ),
          );
        }
      }
      
      debugPrint('[SyncManager] Sinkronisasi background selesai.');
    } catch (e) {
      debugPrint('[SyncManager] Terjadi kesalahan fatal pada sinkronisasi: $e');
    } finally {
      _isSyncing = false;
    }
  }
}
