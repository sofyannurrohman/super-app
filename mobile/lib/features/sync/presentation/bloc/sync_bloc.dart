import 'dart:async';
import 'package:bloc/bloc.dart';
import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter/foundation.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/core/network/network_info.dart';
import 'package:hris_app/core/services/data_integrity_service.dart';
import 'package:hris_app/injection.dart' as di;
import '../../data/repositories/sync_repository.dart';

part 'sync_event.dart';
part 'sync_state.dart';

class SyncBloc extends Bloc<SyncEvent, SyncState> {
  final SyncRepository syncRepository;
  late StreamSubscription _connectivitySubscription;
  Timer? _periodicSyncTimer;
  Timer? _integrityCheckTimer;

  /// Interval sync periodik latar belakang.
  /// Bahkan saat koneksi tidak berubah (misal 3G stabil), tetap sync setiap interval ini.
  static const Duration periodicSyncInterval = Duration(minutes: 5);

  /// Interval pengecekan integritas data (deteksi transaksi yang terlalu lama pending).
  static const Duration integrityCheckInterval = Duration(minutes: 30);

  SyncBloc({required this.syncRepository}) : super(SyncInitial()) {
    on<SyncStarted>(_onSyncStarted);
    on<SyncDataRequested>(_onSyncDataRequested);
    on<SyncMasterDataRequested>(_onSyncMasterDataRequested);
    on<SyncIntegrityCheckRequested>(_onSyncIntegrityCheckRequested);

    // Trigger sync otomatis saat koneksi pulih
    _connectivitySubscription =
        Connectivity().onConnectivityChanged.listen((results) {
      if (results.any((r) => r != ConnectivityResult.none)) {
        debugPrint('[SyncBloc] Koneksi pulih — memicu sync otomatis');
        add(const SyncDataRequested(force: true));
      }
    });

    // Periodic sync latar belakang setiap 5 menit
    // Ini menjamin sync tetap berjalan meski koneksi tidak pernah berubah (misal 3G stabil tapi lambat)
    _periodicSyncTimer = Timer.periodic(periodicSyncInterval, (_) {
      debugPrint('[SyncBloc] Periodic sync triggered');
      add(const SyncDataRequested());
    });

    // Pengecekan integritas data setiap 30 menit
    _integrityCheckTimer = Timer.periodic(integrityCheckInterval, (_) {
      debugPrint('[SyncBloc] Integrity check triggered');
      add(SyncIntegrityCheckRequested());
    });
  }

  Future<void> _onSyncStarted(
      SyncStarted event, Emitter<SyncState> emit) async {
    add(const SyncDataRequested());
  }

  Future<void> _onSyncDataRequested(
      SyncDataRequested event, Emitter<SyncState> emit) async {
    // Jangan jalankan sync baru jika sync sebelumnya masih berjalan
    if (state is SyncInProgress) return;

    // Cek ketersediaan koneksi internet terlebih dahulu
    bool isOnline = true;
    try {
      isOnline = await NetworkInfo().hasInternetAccess;
    } catch (_) {}

    if (!isOnline) {
      debugPrint('[SyncBloc] Perangkat sedang offline — menunda sync hingga sinyal pulih.');
      if (event.isManual) {
        emit(const SyncFailure(
          'Mode Offline aktif. Data tersimpan aman di perangkat dan akan disinkronkan otomatis saat online.',
          isManual: true,
        ));
      } else {
        emit(SyncInitial());
      }
      return;
    }

    emit(SyncInProgress());
    try {
      // 1. Push local data ke server (dengan retry-aware logic)
      final result = await syncRepository.syncAll(force: event.force);
      // 2. Pull fresh data dari server
      await syncRepository.pullMasterData();

      if (result.hasFailures) {
        emit(SyncPartialSuccess(
          successCount: result.successCount,
          failureCount: result.failureCount,
          errors: result.errors,
          isManual: event.isManual,
        ));
      } else {
        emit(SyncSuccess(result: result));
      }
    } catch (e) {
      debugPrint('[SyncBloc] Sync error: $e');
      emit(SyncFailure(e.toString(), isManual: event.isManual));
    }
  }

  Future<void> _onSyncMasterDataRequested(
      SyncMasterDataRequested event, Emitter<SyncState> emit) async {
    emit(SyncInProgress());
    try {
      await syncRepository.pullMasterData();
      emit(SyncMasterDataSuccess());
    } catch (e) {
      emit(SyncFailure(e.toString()));
    }
  }

  /// Handler untuk pengecekan integritas data.
  /// Mendeteksi transaksi yang sudah terlalu lama pending dan memperbarui state
  /// agar UI bisa menampilkan peringatan kepada user.
  Future<void> _onSyncIntegrityCheckRequested(
    SyncIntegrityCheckRequested event,
    Emitter<SyncState> emit,
  ) async {
    try {
      final db = di.sl<AppDatabase>();
      final integrity = DataIntegrityService();
      final summary = await integrity.getSummary(db);

      debugPrint('[SyncBloc] Integrity summary: $summary');

      if (summary.hasWarnings) {
        emit(SyncIntegrityWarning(summary));
      }
    } catch (e) {
      debugPrint('[SyncBloc] Integrity check error: $e');
    }
  }

  @override
  Future<void> close() {
    _connectivitySubscription.cancel();
    _periodicSyncTimer?.cancel();
    _integrityCheckTimer?.cancel();
    return super.close();
  }
}
