import 'package:equatable/equatable.dart';
import 'package:hris_app/features/leave/domain/entities/leave.dart';
import 'package:hris_app/features/leave/domain/entities/leave_balance.dart';

enum LeaveStatus { initial, loading, success, failure }

class LeaveState extends Equatable {
  final LeaveStatus status;
  final List<LeaveBalance> balances;
  final List<Leave> leaves;
  final String? message;
  final String? actionMessage; // For separate success/failure of submissions

  const LeaveState({
    this.status = LeaveStatus.initial,
    this.balances = const [],
    this.leaves = const [],
    this.message,
    this.actionMessage,
  });

  LeaveState copyWith({
    LeaveStatus? status,
    List<LeaveBalance>? balances,
    List<Leave>? leaves,
    String? message,
    String? actionMessage,
    bool clearActionMessage = false,
    bool clearMessage = false,
  }) {
    return LeaveState(
      status: status ?? this.status,
      balances: balances ?? this.balances,
      leaves: leaves ?? this.leaves,
      message: clearMessage ? null : (message ?? this.message),
      actionMessage:
          clearActionMessage ? null : (actionMessage ?? this.actionMessage),
    );
  }

  @override
  List<Object?> get props => [status, balances, leaves, message, actionMessage];
}
