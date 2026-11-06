import 'package:equatable/equatable.dart';

class Reimbursement extends Equatable {
  final String id;
  final String employeeId;
  final String title;
  final String? description;
  final double amount;
  final String? attachmentUrl;
  final String status;
  final String? rejectedReason;
  final DateTime createdAt;
  final String? employeeName;
  final String? category;
  final double? startKm;
  final double? endKm;
  final double? totalKm;
  final String? startKmPhotoUrl;
  final String? endKmPhotoUrl;
  final String? vehiclePlateNumber;
  final double? fuelLiters;

  const Reimbursement({
    required this.id,
    required this.employeeId,
    required this.title,
    this.description,
    required this.amount,
    this.attachmentUrl,
    required this.status,
    this.rejectedReason,
    required this.createdAt,
    this.employeeName,
    this.category,
    this.startKm,
    this.endKm,
    this.totalKm,
    this.startKmPhotoUrl,
    this.endKmPhotoUrl,
    this.vehiclePlateNumber,
    this.fuelLiters,
  });

  factory Reimbursement.fromJson(Map<String, dynamic> json) {
    return Reimbursement(
      id: json['id']?.toString() ?? json['ID']?.toString() ?? '',
      employeeId: json['employee_id']?.toString() ??
          json['EmployeeID']?.toString() ??
          '',
      title: json['title'] ?? json['Title'] ?? '',
      description: json['description'] ?? json['Description'],
      amount: (json['amount'] as num?)?.toDouble() ??
          (json['Amount'] as num?)?.toDouble() ??
          0.0,
      attachmentUrl: json['attachment_url'] ?? json['AttachmentURL'],
      status: json['status'] ?? json['Status'] ?? 'PENDING',
      rejectedReason: json['rejected_reason'] ?? json['RejectedReason'],
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at']).toLocal()
          : (json['CreatedAt'] != null
              ? DateTime.parse(json['CreatedAt']).toLocal()
              : DateTime.now()),
      employeeName: json['employee_name'] ?? json['EmployeeName'],
      category: json['category'] ?? json['Category'],
      startKm: (json['start_km'] as num?)?.toDouble() ??
          (json['StartKM'] as num?)?.toDouble(),
      endKm: (json['end_km'] as num?)?.toDouble() ??
          (json['EndKM'] as num?)?.toDouble(),
      totalKm: (json['total_km'] as num?)?.toDouble() ??
          (json['TotalKM'] as num?)?.toDouble(),
      startKmPhotoUrl:
          json['start_km_photo_url'] ?? json['StartKMPhotoURL'],
      endKmPhotoUrl: json['end_km_photo_url'] ?? json['EndKMPhotoURL'],
      vehiclePlateNumber:
          json['vehicle_plate_number'] ?? json['VehiclePlateNumber'],
      fuelLiters: (json['fuel_liters'] as num?)?.toDouble() ??
          (json['FuelLiters'] as num?)?.toDouble(),
    );
  }

  @override
  List<Object?> get props => [
        id,
        employeeId,
        title,
        description,
        amount,
        attachmentUrl,
        status,
        rejectedReason,
        createdAt,
        employeeName,
        category,
        startKm,
        endKm,
        totalKm,
        startKmPhotoUrl,
        endKmPhotoUrl,
        vehiclePlateNumber,
        fuelLiters,
      ];
}
