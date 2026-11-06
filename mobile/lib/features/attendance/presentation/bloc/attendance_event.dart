import 'package:equatable/equatable.dart';

abstract class AttendanceEvent extends Equatable {
  const AttendanceEvent();

  @override
  List<Object?> get props => [];
}

class ClockInRequested extends AttendanceEvent {
  final bool isClockIn;
  final String? imagePath;
  final double? latitude;
  final double? longitude;
  final List<double>? faceEmbedding;
  final bool isOutOfTown;
  final String destination;

  const ClockInRequested({
    required this.isClockIn,
    this.imagePath,
    this.latitude,
    this.longitude,
    this.faceEmbedding,
    this.isOutOfTown = false,
    this.destination = '',
  });

  @override
  List<Object?> get props =>
      [isClockIn, imagePath, latitude, longitude, faceEmbedding, isOutOfTown, destination];
}

class RegisterFaceRequested extends AttendanceEvent {
  final String imagePath;
  final List<double>? faceEmbedding;

  const RegisterFaceRequested(this.imagePath, {this.faceEmbedding});

  @override
  List<Object?> get props => [imagePath, faceEmbedding];
}

class FetchHistoryRequested extends AttendanceEvent {
  final String? startDate;
  final String? endDate;

  const FetchHistoryRequested({this.startDate, this.endDate});

  @override
  List<Object?> get props => [startDate, endDate];
}

class FetchHomeDataRequested extends AttendanceEvent {}

class SaveOfflineAttendanceEvent extends AttendanceEvent {
  final String type;
  final String photoPath;
  final double latitude;
  final double longitude;

  const SaveOfflineAttendanceEvent({
    required this.type,
    required this.photoPath,
    required this.latitude,
    required this.longitude,
  });

  @override
  List<Object?> get props => [type, photoPath, latitude, longitude];
}

class SyncAttendanceQueueRequested extends AttendanceEvent {}

class GetQueueCountRequested extends AttendanceEvent {}

class SubmitLateReasonRequested extends AttendanceEvent {
  final String attendanceId;
  final String reason;

  const SubmitLateReasonRequested({
    required this.attendanceId,
    required this.reason,
  });

  @override
  List<Object?> get props => [attendanceId, reason];
}
