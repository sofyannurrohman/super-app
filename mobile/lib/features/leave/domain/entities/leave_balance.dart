import 'package:equatable/equatable.dart';

class LeaveBalance extends Equatable {
  final String leaveTypeId;
  final String leaveTypeName;
  final bool isPaid;
  final bool requiresQuota;
  final int total;
  final int used;
  final int remaining;

  const LeaveBalance({
    required this.leaveTypeId,
    required this.leaveTypeName,
    required this.isPaid,
    required this.requiresQuota,
    required this.total,
    required this.used,
    required this.remaining,
  });

  factory LeaveBalance.fromJson(Map<String, dynamic> json) {
    // Robust key mapping: Try both snake_case, PascalCase, and simple name/id
    final rawName = json['leave_type_name'] ??
        json['LeaveTypeName'] ??
        json['name'] ??
        json['Name'] ??
        '';
    final name = rawName.toString().toLowerCase();

    bool requiresQuota =
        json['requires_quota'] ?? json['RequiresQuota'] ?? true;

    // UI Robustness: Fallback if backend flag is missing or naming pattern matches
    if (name.contains('izin') ||
        name.contains('ijin') ||
        name.contains('sakit') ||
        name.contains('musibah') ||
        name.contains('dispensasi') ||
        name.contains('tugas') ||
        name.contains('dinas')) {
      requiresQuota = false;
    }

    return LeaveBalance(
      leaveTypeId: (json['leave_type_id'] ??
              json['LeaveTypeID'] ??
              json['ID'] ??
              json['id'] ??
              '')
          .toString(),
      leaveTypeName: rawName.toString(),
      isPaid: json['is_paid'] ?? json['IsPaid'] ?? true,
      requiresQuota: requiresQuota,
      total: json['total'] ?? json['Total'] ?? json['default_quota'] ?? 0,
      used: json['used'] ?? json['Used'] ?? 0,
      remaining:
          json['remaining'] ?? json['Remaining'] ?? json['default_quota'] ?? 0,
    );
  }

  @override
  List<Object?> get props => [
        leaveTypeId,
        leaveTypeName,
        isPaid,
        requiresQuota,
        total,
        used,
        remaining
      ];
}
