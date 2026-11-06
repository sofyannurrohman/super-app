import 'package:equatable/equatable.dart';

class Overtime extends Equatable {
  final String id;
  final String employeeId;
  final DateTime date;
  final DateTime startTime;
  final DateTime endTime;
  final int durationMinutes;
  final String type;
  final String reason;
  final String status;
  final String? rejectReason;
  final String? attachmentUrl;

  const Overtime({
    required this.id,
    required this.employeeId,
    required this.date,
    required this.startTime,
    required this.endTime,
    required this.durationMinutes,
    required this.type,
    required this.reason,
    required this.status,
    this.rejectReason,
    this.attachmentUrl,
  });

  factory Overtime.fromJson(Map<String, dynamic> json) {
    return Overtime(
      id: json['id'] ?? '',
      employeeId: json['employee_id'] ?? '',
      date: DateTime.parse(json['date']),
      startTime: DateTime.parse(json['start_time']),
      endTime: DateTime.parse(json['end_time']),
      durationMinutes: json['duration_minutes'] ?? 0,
      type: json['type'] ?? 'working_day',
      reason: json['reason'] ?? '',
      status: json['status'] ?? 'pending',
      rejectReason: json['reject_reason'],
      attachmentUrl: json['attachment_url'],
    );
  }

  @override
  List<Object?> get props => [
        id,
        employeeId,
        date,
        startTime,
        endTime,
        durationMinutes,
        type,
        reason,
        status,
        rejectReason,
        attachmentUrl
      ];
}
