import 'package:equatable/equatable.dart';

class EmployeeDirectory extends Equatable {
  final String id;
  final String name;
  final String position;
  final String department;
  final String phoneNumber;
  final String? profileUrl;
  final String? branchId;
  final String? branchName;

  const EmployeeDirectory({
    required this.id,
    required this.name,
    required this.position,
    required this.department,
    required this.phoneNumber,
    this.profileUrl,
    this.branchId,
    this.branchName,
  });

  @override
  List<Object?> get props =>
      [id, name, position, department, phoneNumber, profileUrl, branchId, branchName];

  factory EmployeeDirectory.fromJson(Map<String, dynamic> json) {
    return EmployeeDirectory(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      position: json['position'] ?? '',
      department: json['department'] ?? '',
      phoneNumber: json['phone_number'] ?? '',
      profileUrl: json['profile_url'],
      branchId: json['branch_id'],
      branchName: json['branch_name'],
    );
  }
}
