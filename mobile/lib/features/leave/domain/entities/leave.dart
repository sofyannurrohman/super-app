import 'package:equatable/equatable.dart';

class Leave extends Equatable {
  final String id;
  final String userId;
  final String? leaveTypeId;
  final DateTime startDate;
  final DateTime endDate;
  final String reason;
  final String status;
  final String? attachmentUrl;
  final String? leaveTypeName;
  final String? employeeName;
  final String? rejectReason;
  final String? approverName;
  final String? approverPosition;
  final DateTime? createdAt;

  const Leave({
    required this.id,
    required this.userId,
    this.leaveTypeId,
    required this.startDate,
    required this.endDate,
    required this.reason,
    required this.status,
    this.attachmentUrl,
    this.leaveTypeName,
    this.employeeName,
    this.rejectReason,
    this.approverName,
    this.approverPosition,
    this.createdAt,
  });

  factory Leave.fromJson(Map<String, dynamic> json) {
    return Leave(
      id: json['id']?.toString() ?? json['ID']?.toString() ?? '',
      userId: json['employee_id']?.toString() ??
          json['EmployeeID']?.toString() ??
          json['UserID']?.toString() ??
          '',
      leaveTypeId: json['leave_type_id']?.toString() ??
          json['LeaveTypeID']?.toString() ??
          (json['leave_type'] != null
              ? json['leave_type']['id']?.toString()
              : (json['LeaveType'] != null
                  ? json['LeaveType']['ID']?.toString()
                  : null)),
      startDate: json['start_date'] != null
          ? DateTime.parse(json['start_date']).toLocal()
          : (json['StartDate'] != null
              ? DateTime.parse(json['StartDate']).toLocal()
              : DateTime.now()),
      endDate: json['end_date'] != null
          ? DateTime.parse(json['end_date']).toLocal()
          : (json['EndDate'] != null
              ? DateTime.parse(json['EndDate']).toLocal()
              : DateTime.now()),
      reason: json['reason'] ?? json['Reason'] ?? '',
      status: json['status'] ?? json['Status'] ?? 'PENDING',
      attachmentUrl: json['attachment_url'] ?? json['AttachmentURL'],
      leaveTypeName: json['leave_type_name'] ??
          (json['leave_type'] != null
              ? json['leave_type']['name']
              : (json['LeaveType'] != null
                  ? json['LeaveType']['Name'] ?? json['LeaveType']['name']
                  : null)),
      employeeName: json['employee_name'] ??
          json['EmployeeName'] ??
          (json['employee'] != null
              ? '${json['employee']['first_name'] ?? ''} ${json['employee']['last_name'] ?? ''}'
                  .trim()
              : (json['Employee'] != null
                  ? '${json['Employee']['FirstName'] ?? json['Employee']['first_name'] ?? ''} ${json['Employee']['LastName'] ?? json['Employee']['last_name'] ?? ''}'
                      .trim()
                  : null)),
      rejectReason: json['reject_reason'] ??
          json['RejectReason'] ??
          (json['reject_reason'] is String ? json['reject_reason'] : null),
      approverName: json['approver_name'] ??
          json['ApproverName'] ??
          (json['approver'] != null
              ? '${json['approver']['first_name'] ?? ''} ${json['approver']['last_name'] ?? ''}'
                  .trim()
              : (json['Approver'] != null
                  ? '${json['Approver']['FirstName'] ?? json['Approver']['first_name'] ?? ''} ${json['Approver']['LastName'] ?? json['Approver']['last_name'] ?? ''}'
                      .trim()
                  : null)),
      approverPosition: json['approver_position'] ??
          (json['approver'] != null && json['approver']['job_position'] != null
              ? json['approver']['job_position']['title']
              : (json['Approver'] != null && json['Approver']['JobPosition'] != null
                  ? json['Approver']['JobPosition']['Title'] ?? json['Approver']['JobPosition']['title']
                  : null)),
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at']).toLocal()
          : (json['CreatedAt'] != null
              ? DateTime.parse(json['CreatedAt']).toLocal()
              : null),
    );
  }

  @override
  List<Object?> get props => [
        id,
        userId,
        leaveTypeId,
        startDate,
        endDate,
        reason,
        status,
        attachmentUrl,
        leaveTypeName,
        employeeName,
        rejectReason,
        approverName,
        approverPosition,
        createdAt
      ];
}
