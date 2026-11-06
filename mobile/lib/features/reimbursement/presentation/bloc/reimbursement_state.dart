import 'package:equatable/equatable.dart';
import 'package:hris_app/features/reimbursement/domain/entities/reimbursement.dart';

abstract class ReimbursementState extends Equatable {
  const ReimbursementState();

  @override
  List<Object?> get props => [];
}

class ReimbursementInitial extends ReimbursementState {}

class ReimbursementLoading extends ReimbursementState {}

class ReimbursementHistoryLoaded extends ReimbursementState {
  final List<Reimbursement> history;

  const ReimbursementHistoryLoaded(this.history);

  @override
  List<Object?> get props => [history];
}

class ReimbursementActionSuccess extends ReimbursementState {
  final String message;

  const ReimbursementActionSuccess(this.message);

  @override
  List<Object?> get props => [message];
}

class ReimbursementFailure extends ReimbursementState {
  final String message;

  const ReimbursementFailure(this.message);

  @override
  List<Object?> get props => [message];
}
