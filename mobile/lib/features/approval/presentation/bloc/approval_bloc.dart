import 'package:equatable/equatable.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/leave/domain/entities/leave.dart';
import 'package:hris_app/features/leave/domain/repositories/leave_repository.dart';
import 'package:hris_app/features/reimbursement/domain/entities/reimbursement.dart';
import 'package:hris_app/features/reimbursement/domain/repositories/reimbursement_repository.dart';

// Events
abstract class ApprovalEvent extends Equatable {
  const ApprovalEvent();
  @override
  List<Object?> get props => [];
}

class FetchPendingApprovalsRequested extends ApprovalEvent {}

class LoadMoreLeavesRequested extends ApprovalEvent {}

class LoadMoreReimbursementsRequested extends ApprovalEvent {}

class ApproveLeaveRequested extends ApprovalEvent {
  final String leaveId;
  final String status;
  final String? rejectReason;
  const ApproveLeaveRequested(this.leaveId, this.status, {this.rejectReason});
  @override
  List<Object?> get props => [leaveId, status, rejectReason];
}

class ApproveReimbursementRequested extends ApprovalEvent {
  final String reimbursementId;
  final String status;
  const ApproveReimbursementRequested(this.reimbursementId, this.status);
  @override
  List<Object?> get props => [reimbursementId, status];
}

// States
abstract class ApprovalState extends Equatable {
  const ApprovalState();
  @override
  List<Object?> get props => [];
}

class ApprovalInitial extends ApprovalState {}

class ApprovalLoading extends ApprovalState {}

class ApprovalActionLoading extends ApprovalState {}

class ApprovalDataLoaded extends ApprovalState {
  final List<Leave> pendingLeaves;
  final List<Reimbursement> pendingReimbursements;
  final bool hasMoreLeaves;
  final bool hasMoreReimbursements;
  final bool isLoadingMoreLeaves;
  final bool isLoadingMoreReimbursements;

  const ApprovalDataLoaded({
    required this.pendingLeaves,
    required this.pendingReimbursements,
    this.hasMoreLeaves = true,
    this.hasMoreReimbursements = true,
    this.isLoadingMoreLeaves = false,
    this.isLoadingMoreReimbursements = false,
  });

  ApprovalDataLoaded copyWith({
    List<Leave>? pendingLeaves,
    List<Reimbursement>? pendingReimbursements,
    bool? hasMoreLeaves,
    bool? hasMoreReimbursements,
    bool? isLoadingMoreLeaves,
    bool? isLoadingMoreReimbursements,
  }) {
    return ApprovalDataLoaded(
      pendingLeaves: pendingLeaves ?? this.pendingLeaves,
      pendingReimbursements:
          pendingReimbursements ?? this.pendingReimbursements,
      hasMoreLeaves: hasMoreLeaves ?? this.hasMoreLeaves,
      hasMoreReimbursements:
          hasMoreReimbursements ?? this.hasMoreReimbursements,
      isLoadingMoreLeaves: isLoadingMoreLeaves ?? this.isLoadingMoreLeaves,
      isLoadingMoreReimbursements:
          isLoadingMoreReimbursements ?? this.isLoadingMoreReimbursements,
    );
  }

  @override
  List<Object?> get props => [
        pendingLeaves,
        pendingReimbursements,
        hasMoreLeaves,
        hasMoreReimbursements,
        isLoadingMoreLeaves,
        isLoadingMoreReimbursements,
      ];
}

class ApprovalActionSuccess extends ApprovalState {
  final String message;
  const ApprovalActionSuccess(this.message);
  @override
  List<Object?> get props => [message];
}

class ApprovalFailure extends ApprovalState {
  final String message;
  const ApprovalFailure(this.message);
  @override
  List<Object?> get props => [message];
}

// BLoC
class ApprovalBloc extends Bloc<ApprovalEvent, ApprovalState> {
  final LeaveRepository leaveRepository;
  final ReimbursementRepository reimbursementRepository;

  int _leavePage = 1;
  int _reimbursementPage = 1;
  final int _limit = 15;

  ApprovalBloc(
      {required this.leaveRepository, required this.reimbursementRepository})
      : super(ApprovalInitial()) {
    on<FetchPendingApprovalsRequested>(_onFetchPending);
    on<LoadMoreLeavesRequested>(_onLoadMoreLeaves);
    on<LoadMoreReimbursementsRequested>(_onLoadMoreReimbursements);
    on<ApproveLeaveRequested>(_onApproveLeave);
    on<ApproveReimbursementRequested>(_onApproveReimbursement);
  }

  Future<void> _onFetchPending(
      FetchPendingApprovalsRequested event, Emitter<ApprovalState> emit) async {
    emit(ApprovalLoading());
    _leavePage = 1;
    _reimbursementPage = 1;

    final leaveResult =
        await leaveRepository.getAllLeaves('PENDING', _leavePage, _limit);
    final reimbursementResult = await reimbursementRepository.getAllPending(
        page: _reimbursementPage, limit: _limit);

    List<Leave> leaves = [];
    List<Reimbursement> reimbursements = [];

    leaveResult.fold(
      (f) => debugPrint("ApprovalBloc: Error fetching leaves: ${f.message}"),
      (l) => leaves = l,
    );
    reimbursementResult.fold(
      (f) => debugPrint(
          "ApprovalBloc: Error fetching reimbursements: ${f.message}"),
      (r) => reimbursements = r,
    );

    emit(ApprovalDataLoaded(
      pendingLeaves: leaves,
      pendingReimbursements: reimbursements,
      hasMoreLeaves: leaves.length == _limit,
      hasMoreReimbursements: reimbursements.length == _limit,
    ));
  }

  Future<void> _onLoadMoreLeaves(
      LoadMoreLeavesRequested event, Emitter<ApprovalState> emit) async {
    if (state is ApprovalDataLoaded) {
      final currentState = state as ApprovalDataLoaded;
      if (!currentState.hasMoreLeaves || currentState.isLoadingMoreLeaves)
        return;

      emit(currentState.copyWith(isLoadingMoreLeaves: true));
      _leavePage++;

      final leaveResult =
          await leaveRepository.getAllLeaves('PENDING', _leavePage, _limit);

      List<Leave> newLeaves = [];
      leaveResult.fold(
        (f) => debugPrint(
            "ApprovalBloc: Error fetching more leaves: ${f.message}"),
        (l) => newLeaves = l,
      );

      emit(currentState.copyWith(
        pendingLeaves: [...currentState.pendingLeaves, ...newLeaves],
        hasMoreLeaves: newLeaves.length == _limit,
        isLoadingMoreLeaves: false,
      ));
    }
  }

  Future<void> _onLoadMoreReimbursements(LoadMoreReimbursementsRequested event,
      Emitter<ApprovalState> emit) async {
    if (state is ApprovalDataLoaded) {
      final currentState = state as ApprovalDataLoaded;
      if (!currentState.hasMoreReimbursements ||
          currentState.isLoadingMoreReimbursements) return;

      emit(currentState.copyWith(isLoadingMoreReimbursements: true));
      _reimbursementPage++;

      final reimbursementResult = await reimbursementRepository.getAllPending(
          page: _reimbursementPage, limit: _limit);

      List<Reimbursement> newReimbursements = [];
      reimbursementResult.fold(
        (f) => debugPrint(
            "ApprovalBloc: Error fetching more reimbursements: ${f.message}"),
        (r) => newReimbursements = r,
      );

      emit(currentState.copyWith(
        pendingReimbursements: [
          ...currentState.pendingReimbursements,
          ...newReimbursements
        ],
        hasMoreReimbursements: newReimbursements.length == _limit,
        isLoadingMoreReimbursements: false,
      ));
    }
  }

  Future<void> _onApproveLeave(
      ApproveLeaveRequested event, Emitter<ApprovalState> emit) async {
    emit(ApprovalActionLoading());
    final result = await leaveRepository.approveLeave(
      event.leaveId,
      event.status,
      rejectReason: event.rejectReason,
    );
    result.fold(
      (failure) => emit(ApprovalFailure(failure.message)),
      (_) {
        emit(ApprovalActionSuccess('Leave request ${event.status}'));
        add(FetchPendingApprovalsRequested());
      },
    );
  }

  Future<void> _onApproveReimbursement(
      ApproveReimbursementRequested event, Emitter<ApprovalState> emit) async {
    emit(ApprovalActionLoading());
    final result = await reimbursementRepository.approveReimbursement(
        event.reimbursementId, event.status);
    result.fold(
      (failure) => emit(ApprovalFailure(failure.message)),
      (_) {
        emit(ApprovalActionSuccess('Reimbursement request ${event.status}'));
        add(FetchPendingApprovalsRequested());
      },
    );
  }
}
