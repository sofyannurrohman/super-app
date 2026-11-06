import 'package:equatable/equatable.dart';

class EpkwtRequest extends Equatable {
  final String id;
  final String employeeId;
  final String? documentNumber;
  final String identityNumber;
  final String firstName;
  final String? lastName;
  final String? phoneNumber;
  final String? gender;
  final DateTime? birthDate;
  final String? birthPlace;
  final String? addressKtp;
  final String? addressResidential;
  final String? maritalStatus;
  final String? jobPositionId;
  final String? departmentId;
  final double salary;
  final DateTime? startDate;
  final DateTime? endDate;
  final DateTime? agreementDate;
  final String? employeeSignatureUrl;
  final String? hrSignatureUrl;
  final String status;
  final DateTime createdAt;
  final String? jobPositionTitle;
  final String? departmentName;

  const EpkwtRequest({
    required this.id,
    required this.employeeId,
    this.documentNumber,
    required this.identityNumber,
    required this.firstName,
    this.lastName,
    this.phoneNumber,
    this.gender,
    this.birthDate,
    this.birthPlace,
    this.addressKtp,
    this.addressResidential,
    this.maritalStatus,
    this.jobPositionId,
    this.departmentId,
    required this.salary,
    this.startDate,
    this.endDate,
    this.agreementDate,
    this.employeeSignatureUrl,
    this.hrSignatureUrl,
    required this.status,
    required this.createdAt,
    this.jobPositionTitle,
    this.departmentName,
  });

  String get employeeName => '$firstName ${lastName ?? ''}'.trim();

  factory EpkwtRequest.fromJson(Map<String, dynamic> json) {
    DateTime? parseDate(dynamic value) {
      if (value == null) return null;
      try {
        return DateTime.parse(value.toString());
      } catch (_) {
        return null;
      }
    }

    final jobPos = json['job_position'] as Map<String, dynamic>?;
    final dept = json['department'] as Map<String, dynamic>?;

    return EpkwtRequest(
      id: json['id'] ?? '',
      employeeId: json['employee_id'] ?? '',
      documentNumber: json['document_number'],
      identityNumber: json['identity_number'] ?? '',
      firstName: json['first_name'] ?? '',
      lastName: json['last_name'],
      phoneNumber: json['phone_number'],
      gender: json['gender'],
      birthDate: parseDate(json['birth_date']),
      birthPlace: json['birth_place'],
      addressKtp: json['address_ktp'],
      addressResidential: json['address_residential'],
      maritalStatus: json['marital_status'],
      jobPositionId: json['job_position_id'],
      departmentId: json['department_id'],
      salary: double.tryParse(json['salary']?.toString() ?? '0') ?? 0.0,
      startDate: parseDate(json['start_date']),
      endDate: parseDate(json['end_date']),
      agreementDate: parseDate(json['agreement_date']),
      employeeSignatureUrl: json['employee_signature_url'],
      hrSignatureUrl: json['hr_signature_url'],
      status: json['status'] ?? 'PENDING_HR',
      createdAt: parseDate(json['created_at']) ?? DateTime.now(),
      jobPositionTitle: jobPos?['title'],
      departmentName: dept?['name'],
    );
  }

  @override
  List<Object?> get props => [
        id,
        employeeId,
        documentNumber,
        identityNumber,
        firstName,
        lastName,
        phoneNumber,
        gender,
        birthDate,
        birthPlace,
        addressKtp,
        addressResidential,
        maritalStatus,
        jobPositionId,
        departmentId,
        salary,
        startDate,
        endDate,
        agreementDate,
        employeeSignatureUrl,
        hrSignatureUrl,
        status,
        createdAt,
        jobPositionTitle,
        departmentName,
      ];
}
