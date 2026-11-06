import 'dart:typed_data';
import 'package:equatable/equatable.dart';

abstract class LeaveEvent extends Equatable {
  const LeaveEvent();

  @override
  List<Object?> get props => [];
}

class SubmitLeaveRequested extends LeaveEvent {
  final String leaveTypeId;
  final String startDate;
  final String endDate;
  final String reason;
  final Uint8List? attachmentBytes;
  final String? attachmentName;

  const SubmitLeaveRequested({
    required this.leaveTypeId,
    required this.startDate,
    required this.endDate,
    required this.reason,
    this.attachmentBytes,
    this.attachmentName,
  });

  @override
  List<Object?> get props => [
        leaveTypeId,
        startDate,
        endDate,
        reason,
        attachmentBytes,
        attachmentName
      ];
}

class FetchLeaveBalancesRequested extends LeaveEvent {
  const FetchLeaveBalancesRequested();
}

class FetchMyLeavesRequested extends LeaveEvent {
  final int page;
  final int limit;

  const FetchMyLeavesRequested({this.page = 1, this.limit = 10});

  @override
  List<Object?> get props => [page, limit];
}

class FetchAllLeavesRequested extends LeaveEvent {
  final String? status;
  final int page;
  final int limit;

  const FetchAllLeavesRequested({this.status, this.page = 1, this.limit = 10});

  @override
  List<Object?> get props => [status, page, limit];
}

class ApproveLeaveRequested extends LeaveEvent {
  final String leaveId;
  final String status;

  const ApproveLeaveRequested({required this.leaveId, required this.status});

  @override
  List<Object?> get props => [leaveId, status];
}

class UpdateLeaveRequested extends LeaveEvent {
  final String leaveId;
  final String leaveTypeId;
  final String startDate;
  final String endDate;
  final String reason;
  final Uint8List? attachmentBytes;
  final String? attachmentName;

  const UpdateLeaveRequested({
    required this.leaveId,
    required this.leaveTypeId,
    required this.startDate,
    required this.endDate,
    required this.reason,
    this.attachmentBytes,
    this.attachmentName,
  });

  @override
  List<Object?> get props => [
        leaveId,
        leaveTypeId,
        startDate,
        endDate,
        reason,
        attachmentBytes,
        attachmentName
      ];
}

class DeleteLeaveRequested extends LeaveEvent {
  final String leaveId;

  const DeleteLeaveRequested(this.leaveId);

  @override
  List<Object?> get props => [leaveId];
}
