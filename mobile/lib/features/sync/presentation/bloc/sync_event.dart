part of 'sync_bloc.dart';

abstract class SyncEvent extends Equatable {
  const SyncEvent();

  @override
  List<Object> get props => [];
}

class SyncStarted extends SyncEvent {}

class SyncDataRequested extends SyncEvent {
  final bool force;
  final bool isManual;
  const SyncDataRequested({this.force = false, this.isManual = false});

  @override
  List<Object> get props => [force, isManual];
}

class SyncMasterDataRequested extends SyncEvent {}

class SyncIntegrityCheckRequested extends SyncEvent {}
