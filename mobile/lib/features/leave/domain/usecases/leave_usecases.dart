import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/features/leave/domain/entities/leave.dart';
import 'package:hris_app/features/leave/domain/entities/leave_balance.dart';
import 'package:hris_app/features/leave/domain/repositories/leave_repository.dart';

import 'dart:typed_data';

class SubmitLeaveUseCase {
  final LeaveRepository repository;

  SubmitLeaveUseCase(this.repository);

  Future<Either<Failure, void>> call(
      String leaveTypeId, String startDate, String endDate, String reason,
      {Uint8List? attachmentBytes, String? attachmentName}) {
    return repository.submitLeave(leaveTypeId, startDate, endDate, reason,
        attachmentBytes: attachmentBytes, attachmentName: attachmentName);
  }
}

class GetMyLeavesUseCase {
  final LeaveRepository repository;

  GetMyLeavesUseCase(this.repository);

  Future<Either<Failure, List<Leave>>> call({int page = 1, int limit = 10}) {
    return repository.getMyLeaves(page, limit);
  }
}

class GetAllLeavesUseCase {
  final LeaveRepository repository;

  GetAllLeavesUseCase(this.repository);

  Future<Either<Failure, List<Leave>>> call(
      {String? status, int page = 1, int limit = 10}) {
    return repository.getAllLeaves(status, page, limit);
  }
}

class ApproveLeaveUseCase {
  final LeaveRepository repository;

  ApproveLeaveUseCase(this.repository);

  Future<Either<Failure, void>> call(String leaveId, String status) {
    return repository.approveLeave(leaveId, status);
  }
}

class GetLeaveBalancesUseCase {
  final LeaveRepository repository;

  GetLeaveBalancesUseCase(this.repository);

  Future<Either<Failure, List<LeaveBalance>>> call() {
    return repository.getMyBalances();
  }
}

class UpdateLeaveUseCase {
  final LeaveRepository repository;

  UpdateLeaveUseCase(this.repository);

  Future<Either<Failure, void>> call(String leaveId, String leaveTypeId,
      String startDate, String endDate, String reason,
      {Uint8List? attachmentBytes, String? attachmentName}) {
    return repository.updateLeave(
        leaveId, leaveTypeId, startDate, endDate, reason,
        attachmentBytes: attachmentBytes, attachmentName: attachmentName);
  }
}

class DeleteLeaveUseCase {
  final LeaveRepository repository;

  DeleteLeaveUseCase(this.repository);

  Future<Either<Failure, void>> call(String leaveId) {
    return repository.deleteLeave(leaveId);
  }
}
