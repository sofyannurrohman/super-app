import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/supervisor/data/models/delivery_batch_model.dart';
import 'package:hris_app/features/supervisor/data/models/sales_kpi_model.dart';
import 'package:hris_app/features/supervisor/data/models/sales_transfer_model.dart';
import 'package:hris_app/features/supervisor/domain/usecases/supervisor_usecases.dart';

abstract class SupervisorEvent extends Equatable {
  const SupervisorEvent();

  @override
  List<Object?> get props => [];
}

class LoadPendingTransfers extends SupervisorEvent {
  final String branchId;
  const LoadPendingTransfers(this.branchId);
}

class DoApproveTransfer extends SupervisorEvent {
  final String transferId;
  final int approvedQuantity;
  const DoApproveTransfer(this.transferId, this.approvedQuantity);
}

class DoRejectTransfer extends SupervisorEvent {
  final String transferId;
  const DoRejectTransfer(this.transferId);
}

class LoadPendingDeliveryBatches extends SupervisorEvent {
  final String branchId;
  const LoadPendingDeliveryBatches(this.branchId);
}

class DoApproveDeliveryBatch extends SupervisorEvent {
  final String batchId;
  const DoApproveDeliveryBatch(this.batchId);
}

class DoCancelDeliveryBatch extends SupervisorEvent {
  final String batchId;
  const DoCancelDeliveryBatch(this.batchId);
}

class LoadSalesKPIs extends SupervisorEvent {
  final int month;
  final int year;
  final String branchId;
  const LoadSalesKPIs(this.month, this.year, this.branchId);
}

class SupervisorState extends Equatable {
  final bool isLoadingTransfers;
  final bool isLoadingBatches;
  final bool isLoadingKPIs;
  final bool isActionLoading;

  final List<SalesTransferModel> pendingTransfers;
  final List<DeliveryBatchModel> pendingBatches;
  final List<SalesKpiModel> salesKPIs;

  final String? errorMessage;
  final String? successMessage;

  const SupervisorState({
    this.isLoadingTransfers = false,
    this.isLoadingBatches = false,
    this.isLoadingKPIs = false,
    this.isActionLoading = false,
    this.pendingTransfers = const [],
    this.pendingBatches = const [],
    this.salesKPIs = const [],
    this.errorMessage,
    this.successMessage,
  });

  SupervisorState copyWith({
    bool? isLoadingTransfers,
    bool? isLoadingBatches,
    bool? isLoadingKPIs,
    bool? isActionLoading,
    List<SalesTransferModel>? pendingTransfers,
    List<DeliveryBatchModel>? pendingBatches,
    List<SalesKpiModel>? salesKPIs,
    String? errorMessage,
    String? successMessage,
    bool clearMessages = false,
  }) {
    return SupervisorState(
      isLoadingTransfers: isLoadingTransfers ?? this.isLoadingTransfers,
      isLoadingBatches: isLoadingBatches ?? this.isLoadingBatches,
      isLoadingKPIs: isLoadingKPIs ?? this.isLoadingKPIs,
      isActionLoading: isActionLoading ?? this.isActionLoading,
      pendingTransfers: pendingTransfers ?? this.pendingTransfers,
      pendingBatches: pendingBatches ?? this.pendingBatches,
      salesKPIs: salesKPIs ?? this.salesKPIs,
      errorMessage: clearMessages ? null : (errorMessage ?? this.errorMessage),
      successMessage:
          clearMessages ? null : (successMessage ?? this.successMessage),
    );
  }

  @override
  List<Object?> get props => [
        isLoadingTransfers,
        isLoadingBatches,
        isLoadingKPIs,
        isActionLoading,
        pendingTransfers,
        pendingBatches,
        salesKPIs,
        errorMessage,
        successMessage,
      ];
}

class SupervisorBloc extends Bloc<SupervisorEvent, SupervisorState> {
  final GetPendingTransfers getPendingTransfers;
  final ApproveTransfer approveTransfer;
  final RejectTransfer rejectTransfer;
  final GetPendingDeliveryBatches getPendingDeliveryBatches;
  final ApproveDeliveryBatch approveDeliveryBatch;
  final CancelDeliveryBatch cancelDeliveryBatch;
  final GetSalesKPIs getSalesKPIs;

  SupervisorBloc({
    required this.getPendingTransfers,
    required this.approveTransfer,
    required this.rejectTransfer,
    required this.getPendingDeliveryBatches,
    required this.approveDeliveryBatch,
    required this.cancelDeliveryBatch,
    required this.getSalesKPIs,
  }) : super(const SupervisorState()) {
    on<LoadPendingTransfers>(_onLoadPendingTransfers);
    on<DoApproveTransfer>(_onDoApproveTransfer);
    on<DoRejectTransfer>(_onDoRejectTransfer);
    on<LoadPendingDeliveryBatches>(_onLoadPendingDeliveryBatches);
    on<DoApproveDeliveryBatch>(_onDoApproveDeliveryBatch);
    on<DoCancelDeliveryBatch>(_onDoCancelDeliveryBatch);
    on<LoadSalesKPIs>(_onLoadSalesKPIs);
  }

  Future<void> _onLoadPendingTransfers(
      LoadPendingTransfers event, Emitter<SupervisorState> emit) async {
    emit(state.copyWith(isLoadingTransfers: true, clearMessages: true));
    final result = await getPendingTransfers(event.branchId);
    result.fold(
      (failure) => emit(state.copyWith(
          isLoadingTransfers: false, errorMessage: failure.message)),
      (data) => emit(
          state.copyWith(isLoadingTransfers: false, pendingTransfers: data)),
    );
  }

  Future<void> _onDoApproveTransfer(
      DoApproveTransfer event, Emitter<SupervisorState> emit) async {
    emit(state.copyWith(isActionLoading: true, clearMessages: true));
    final result =
        await approveTransfer(event.transferId, event.approvedQuantity);
    result.fold(
      (failure) => emit(state.copyWith(
          isActionLoading: false, errorMessage: failure.message)),
      (_) {
        final newTransfers = state.pendingTransfers
            .where((t) => t.id != event.transferId)
            .toList();
        emit(state.copyWith(
            isActionLoading: false,
            pendingTransfers: newTransfers,
            successMessage: 'Barang berhasil disetujui'));
      },
    );
  }

  Future<void> _onDoRejectTransfer(
      DoRejectTransfer event, Emitter<SupervisorState> emit) async {
    emit(state.copyWith(isActionLoading: true, clearMessages: true));
    final result = await rejectTransfer(event.transferId);
    result.fold(
      (failure) => emit(state.copyWith(
          isActionLoading: false, errorMessage: failure.message)),
      (_) {
        final newTransfers = state.pendingTransfers
            .where((t) => t.id != event.transferId)
            .toList();
        emit(state.copyWith(
            isActionLoading: false,
            pendingTransfers: newTransfers,
            successMessage: 'Barang berhasil ditolak'));
      },
    );
  }

  Future<void> _onLoadPendingDeliveryBatches(
      LoadPendingDeliveryBatches event, Emitter<SupervisorState> emit) async {
    emit(state.copyWith(isLoadingBatches: true, clearMessages: true));
    final result = await getPendingDeliveryBatches(event.branchId);
    result.fold(
      (failure) => emit(state.copyWith(
          isLoadingBatches: false, errorMessage: failure.message)),
      (data) =>
          emit(state.copyWith(isLoadingBatches: false, pendingBatches: data)),
    );
  }

  Future<void> _onDoApproveDeliveryBatch(
      DoApproveDeliveryBatch event, Emitter<SupervisorState> emit) async {
    emit(state.copyWith(isActionLoading: true, clearMessages: true));
    final result = await approveDeliveryBatch(event.batchId);
    result.fold(
      (failure) => emit(state.copyWith(
          isActionLoading: false, errorMessage: failure.message)),
      (_) {
        final newBatches =
            state.pendingBatches.where((b) => b.id != event.batchId).toList();
        emit(state.copyWith(
            isActionLoading: false,
            pendingBatches: newBatches,
            successMessage: 'Surat jalan berhasil diterbitkan'));
      },
    );
  }

  Future<void> _onDoCancelDeliveryBatch(
      DoCancelDeliveryBatch event, Emitter<SupervisorState> emit) async {
    emit(state.copyWith(isActionLoading: true, clearMessages: true));
    final result = await cancelDeliveryBatch(event.batchId);
    result.fold(
      (failure) => emit(state.copyWith(
          isActionLoading: false, errorMessage: failure.message)),
      (_) {
        final newBatches =
            state.pendingBatches.where((b) => b.id != event.batchId).toList();
        emit(state.copyWith(
            isActionLoading: false,
            pendingBatches: newBatches,
            successMessage: 'Surat jalan berhasil dibatalkan'));
      },
    );
  }

  Future<void> _onLoadSalesKPIs(
      LoadSalesKPIs event, Emitter<SupervisorState> emit) async {
    emit(state.copyWith(isLoadingKPIs: true, clearMessages: true));
    final result = await getSalesKPIs(event.month, event.year, event.branchId);
    result.fold(
      (failure) => emit(
          state.copyWith(isLoadingKPIs: false, errorMessage: failure.message)),
      (data) => emit(state.copyWith(isLoadingKPIs: false, salesKPIs: data)),
    );
  }
}
