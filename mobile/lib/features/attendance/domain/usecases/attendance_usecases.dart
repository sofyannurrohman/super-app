import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/features/attendance/domain/entities/attendance.dart';
import 'package:hris_app/features/attendance/data/repositories/attendance_repository.dart';

class CheckInUseCase {
  final AttendanceRepository repository;

  CheckInUseCase(this.repository);

  Future<Either<Failure, Attendance>> call(
      double lat, double lng, String? selfiePath) {
    return repository.checkIn(lat, lng, selfiePath, null);
  }
}

class CheckOutUseCase {
  final AttendanceRepository repository;

  CheckOutUseCase(this.repository);

  Future<Either<Failure, Attendance>> call(
      double lat, double lng, String? selfiePath) {
    return repository.checkOut(lat, lng, selfiePath, null);
  }
}

class GetAttendanceHistoryUseCase {
  final AttendanceRepository repository;

  GetAttendanceHistoryUseCase(this.repository);

  Future<Either<Failure, List<Attendance>>> call(
      {String? startDate, String? endDate}) {
    return repository.getHistory(startDate: startDate, endDate: endDate);
  }
}
