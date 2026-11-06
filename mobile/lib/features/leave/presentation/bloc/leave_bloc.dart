import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/leave/domain/usecases/leave_usecases.dart';
import 'leave_event.dart';
import 'leave_state.dart';

class LeaveBloc extends Bloc<LeaveEvent, LeaveState> {
  final SubmitLeaveUseCase submitLeaveUseCase;
  final GetMyLeavesUseCase getMyLeavesUseCase;
  final GetAllLeavesUseCase getAllLeavesUseCase;
  final ApproveLeaveUseCase approveLeaveUseCase;
  final GetLeaveBalancesUseCase getLeaveBalancesUseCase;
  final UpdateLeaveUseCase updateLeaveUseCase;
  final DeleteLeaveUseCase deleteLeaveUseCase;

  LeaveBloc({
    required this.submitLeaveUseCase,
    required this.getMyLeavesUseCase,
    required this.getAllLeavesUseCase,
    required this.approveLeaveUseCase,
    required this.getLeaveBalancesUseCase,
    required this.updateLeaveUseCase,
    required this.deleteLeaveUseCase,
  }) : super(const LeaveState()) {
    on<SubmitLeaveRequested>(_onSubmitLeave);
    on<FetchMyLeavesRequested>(_onFetchMyLeaves);
    on<FetchAllLeavesRequested>(_onFetchAllLeaves);
    on<ApproveLeaveRequested>(_onApproveLeave);
    on<FetchLeaveBalancesRequested>(_onFetchLeaveBalances);
    on<UpdateLeaveRequested>(_onUpdateLeave);
    on<DeleteLeaveRequested>(_onDeleteLeave);
  }

  Future<void> _onSubmitLeave(
      SubmitLeaveRequested event, Emitter<LeaveState> emit) async {
    emit(state.copyWith(status: LeaveStatus.loading, clearActionMessage: true));
    final result = await submitLeaveUseCase(
      event.leaveTypeId,
      event.startDate,
      event.endDate,
      event.reason,
      attachmentBytes: event.attachmentBytes,
      attachmentName: event.attachmentName,
    );
    result.fold(
      (failure) => emit(state.copyWith(
          status: LeaveStatus.failure, actionMessage: failure.message)),
      (_) {
        emit(state.copyWith(
            status: LeaveStatus.success,
            actionMessage: 'Pengajuan cuti berhasil dikirim!'));
        add(const FetchMyLeavesRequested());
        add(const FetchLeaveBalancesRequested());
      },
    );
  }

  Future<void> _onFetchMyLeaves(
      FetchMyLeavesRequested event, Emitter<LeaveState> emit) async {
    emit(state.copyWith(status: LeaveStatus.loading, clearActionMessage: true));
    final result =
        await getMyLeavesUseCase(page: event.page, limit: event.limit);
    result.fold(
      (failure) => emit(state.copyWith(
          status: LeaveStatus.failure, message: failure.message)),
      (leaves) =>
          emit(state.copyWith(status: LeaveStatus.success, leaves: leaves)),
    );
  }

  Future<void> _onFetchAllLeaves(
      FetchAllLeavesRequested event, Emitter<LeaveState> emit) async {
    emit(state.copyWith(status: LeaveStatus.loading));
    final result = await getAllLeavesUseCase(
        status: event.status, page: event.page, limit: event.limit);
    result.fold(
      (failure) => emit(state.copyWith(
          status: LeaveStatus.failure, message: failure.message)),
      (leaves) =>
          emit(state.copyWith(status: LeaveStatus.success, leaves: leaves)),
    );
  }

  Future<void> _onApproveLeave(
      ApproveLeaveRequested event, Emitter<LeaveState> emit) async {
    emit(state.copyWith(status: LeaveStatus.loading, clearActionMessage: true));
    final result = await approveLeaveUseCase(event.leaveId, event.status);
    result.fold(
      (failure) => emit(state.copyWith(
          status: LeaveStatus.failure, actionMessage: failure.message)),
      (_) {
        emit(state.copyWith(
            status: LeaveStatus.success,
            actionMessage: 'Status cuti diperbarui menjadi ${event.status}'));
        add(const FetchMyLeavesRequested());
        add(const FetchLeaveBalancesRequested());
      },
    );
  }

  Future<void> _onFetchLeaveBalances(
      FetchLeaveBalancesRequested event, Emitter<LeaveState> emit) async {
    emit(state.copyWith(status: LeaveStatus.loading, clearActionMessage: true));
    final result = await getLeaveBalancesUseCase();
    result.fold(
      (failure) => emit(state.copyWith(
          status: LeaveStatus.failure, message: failure.message)),
      (balances) =>
          emit(state.copyWith(status: LeaveStatus.success, balances: balances)),
    );
  }

  Future<void> _onUpdateLeave(
      UpdateLeaveRequested event, Emitter<LeaveState> emit) async {
    emit(state.copyWith(status: LeaveStatus.loading, clearActionMessage: true));
    final result = await updateLeaveUseCase(
      event.leaveId,
      event.leaveTypeId,
      event.startDate,
      event.endDate,
      event.reason,
      attachmentBytes: event.attachmentBytes,
      attachmentName: event.attachmentName,
    );
    result.fold(
      (failure) => emit(state.copyWith(
          status: LeaveStatus.failure, actionMessage: failure.message)),
      (_) {
        emit(state.copyWith(
            status: LeaveStatus.success,
            actionMessage: 'Pengajuan cuti berhasil diperbarui!'));
        add(const FetchMyLeavesRequested());
        add(const FetchLeaveBalancesRequested());
      },
    );
  }

  Future<void> _onDeleteLeave(
      DeleteLeaveRequested event, Emitter<LeaveState> emit) async {
    emit(state.copyWith(status: LeaveStatus.loading, clearActionMessage: true));
    final result = await deleteLeaveUseCase(event.leaveId);
    result.fold(
      (failure) => emit(state.copyWith(
          status: LeaveStatus.failure, actionMessage: failure.message)),
      (_) {
        emit(state.copyWith(
            status: LeaveStatus.success,
            actionMessage: 'Pengajuan cuti berhasil dibatalkan!'));
        add(const FetchMyLeavesRequested());
        add(const FetchLeaveBalancesRequested());
      },
    );
  }
}
