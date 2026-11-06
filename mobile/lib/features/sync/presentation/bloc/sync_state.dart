part of 'sync_bloc.dart';

abstract class SyncState extends Equatable {
  const SyncState();

  @override
  List<Object> get props => [];
}

class SyncInitial extends SyncState {}

class SyncInProgress extends SyncState {}

class SyncSuccess extends SyncState {
  final SyncResult? result;
  const SyncSuccess({this.result});

  @override
  List<Object> get props => result != null ? [result!.successCount, result!.failureCount] : [];
}

class SyncPartialSuccess extends SyncState {
  final int successCount;
  final int failureCount;
  final List<String> errors;
  final bool isManual;

  const SyncPartialSuccess({
    required this.successCount,
    required this.failureCount,
    this.errors = const [],
    this.isManual = false,
  });

  @override
  List<Object> get props => [successCount, failureCount, errors, isManual];
}

class SyncMasterDataSuccess extends SyncState {}

class SyncFailure extends SyncState {
  final String error;
  final bool isManual;
  const SyncFailure(this.error, {this.isManual = false});

  @override
  List<Object> get props => [error, isManual];
}

/// State ini di-emit saat pengecekan integritas menemukan data yang bermasalah.
/// UI harus menampilkan banner peringatan kepada user agar mereka sadar ada
/// transaksi yang belum berhasil tersinkronisasi.
class SyncIntegrityWarning extends SyncState {
  final DataIntegritySummary summary;
  const SyncIntegrityWarning(this.summary);

  @override
  List<Object> get props => [
        summary.pendingTransactions,
        summary.stuckItemCount,
        summary.stalePendingCount
      ];
}
