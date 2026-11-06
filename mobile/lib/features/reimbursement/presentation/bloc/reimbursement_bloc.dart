import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/reimbursement/domain/repositories/reimbursement_repository.dart';
import 'reimbursement_event.dart';
import 'reimbursement_state.dart';

class ReimbursementBloc extends Bloc<ReimbursementEvent, ReimbursementState> {
  final ReimbursementRepository repository;

  ReimbursementBloc({required this.repository})
      : super(ReimbursementInitial()) {
    on<SubmitReimbursementRequested>(_onSubmitReimbursement);
    on<FetchReimbursementHistoryRequested>(_onFetchHistory);
    on<UpdateReimbursementRequested>(_onUpdateReimbursement);
    on<DeleteReimbursementRequested>(_onDeleteReimbursement);
  }

  Future<void> _onSubmitReimbursement(SubmitReimbursementRequested event,
      Emitter<ReimbursementState> emit) async {
    emit(ReimbursementLoading());
    final result = await repository.submitReimbursement(
      title: event.title,
      description: event.description,
      amount: event.amount,
      attachmentBytes: event.attachmentBytes,
      attachmentName: event.attachmentName,
    );
    result.fold(
      (failure) => emit(ReimbursementFailure(failure.message)),
      (_) => emit(const ReimbursementActionSuccess(
          'Reimbursement submitted successfully!')),
    );
  }

  Future<void> _onFetchHistory(FetchReimbursementHistoryRequested event,
      Emitter<ReimbursementState> emit) async {
    emit(ReimbursementLoading());
    final result =
        await repository.getMyHistory(page: event.page, limit: event.limit);
    result.fold(
      (failure) => emit(ReimbursementFailure(failure.message)),
      (history) => emit(ReimbursementHistoryLoaded(history)),
    );
  }

  Future<void> _onUpdateReimbursement(UpdateReimbursementRequested event,
      Emitter<ReimbursementState> emit) async {
    emit(ReimbursementLoading());
    final result = await repository.updateReimbursement(
      id: event.id,
      title: event.title,
      description: event.description,
      amount: event.amount,
      attachmentBytes: event.attachmentBytes,
      attachmentName: event.attachmentName,
    );
    result.fold(
      (failure) => emit(ReimbursementFailure(failure.message)),
      (_) => emit(const ReimbursementActionSuccess(
          'Reimbursement updated successfully!')),
    );
  }

  Future<void> _onDeleteReimbursement(DeleteReimbursementRequested event,
      Emitter<ReimbursementState> emit) async {
    emit(ReimbursementLoading());
    final result = await repository.deleteReimbursement(event.id);
    result.fold(
      (failure) => emit(ReimbursementFailure(failure.message)),
      (_) => emit(const ReimbursementActionSuccess(
          'Reimbursement deleted successfully!')),
    );
  }
}
