import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/features/leave/domain/entities/leave.dart';
import 'package:hris_app/features/leave/domain/entities/leave_balance.dart';

import 'dart:typed_data';

abstract class LeaveRepository {
  Future<Either<Failure, void>> submitLeave(
      String leaveTypeId, String startDate, String endDate, String reason,
      {Uint8List? attachmentBytes, String? attachmentName});
  Future<Either<Failure, List<Leave>>> getMyLeaves(int page, int limit);
  Future<Either<Failure, List<Leave>>> getAllLeaves(
      String? status, int page, int limit);
  Future<Either<Failure, void>> approveLeave(String leaveId, String status,
      {String? rejectReason});
  Future<Either<Failure, List<LeaveBalance>>> getMyBalances();
  Future<Either<Failure, void>> updateLeave(String leaveId, String leaveTypeId,
      String startDate, String endDate, String reason,
      {Uint8List? attachmentBytes, String? attachmentName});
  Future<Either<Failure, void>> deleteLeave(String leaveId);
}
