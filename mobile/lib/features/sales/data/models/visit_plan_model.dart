class VisitPlanModel {
  final String storeId;
  final String storeName;
  final String address;
  final String status; // PLANNED, CHECKED_IN, COMPLETED, EXTRA
  final bool isExtra;
  final String? visitId;
  final double? latitude;
  final double? longitude;
  final DateTime? lastVisitDate;
  final DateTime? lastTransactionDate;
  final double? lastTransactionAmount;
  final String? lastTransactionItems;
  final bool isPriority;
  final String? assignedEmployeeId;
  final String? assignedEmployeeName;
  double? distanceCache; // Optional local field for sorting

  VisitPlanModel({
    required this.storeId,
    required this.storeName,
    required this.address,
    required this.status,
    required this.isExtra,
    this.visitId,
    this.latitude,
    this.longitude,
    this.lastVisitDate,
    this.lastTransactionDate,
    this.lastTransactionAmount,
    this.lastTransactionItems,
    this.isPriority = false,
    this.assignedEmployeeId,
    this.assignedEmployeeName,
    this.distanceCache,
  });

  factory VisitPlanModel.fromJson(Map<String, dynamic> json) {
    return VisitPlanModel(
      storeId: json['store_id'] ?? '',
      storeName: json['store_name'] ?? '',
      address: json['address'] ?? '',
      status: json['status'] ?? 'PLANNED',
      isExtra: json['is_extra'] ?? false,
      visitId: json['visit_id'],
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      lastVisitDate: json['last_visit_date'] != null
          ? DateTime.tryParse(json['last_visit_date'])?.toLocal()
          : null,
      lastTransactionDate: json['last_transaction_date'] != null
          ? DateTime.tryParse(json['last_transaction_date'])?.toLocal()
          : null,
      lastTransactionAmount:
          (json['last_transaction_amount'] as num?)?.toDouble(),
      lastTransactionItems: json['last_transaction_items'],
      isPriority: json['is_priority'] ?? false,
      assignedEmployeeId: json['assigned_employee_id']?.toString(),
      assignedEmployeeName: json['assigned_employee_name']?.toString(),
    );
  }

  VisitPlanModel copyWith({
    String? storeId,
    String? storeName,
    String? address,
    String? status,
    bool? isExtra,
    String? visitId,
    double? latitude,
    double? longitude,
    DateTime? lastVisitDate,
    DateTime? lastTransactionDate,
    double? lastTransactionAmount,
    String? lastTransactionItems,
    bool? isPriority,
    String? assignedEmployeeId,
    String? assignedEmployeeName,
    double? distanceCache,
  }) {
    return VisitPlanModel(
      storeId: storeId ?? this.storeId,
      storeName: storeName ?? this.storeName,
      address: address ?? this.address,
      status: status ?? this.status,
      isExtra: isExtra ?? this.isExtra,
      visitId: visitId ?? this.visitId,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      lastVisitDate: lastVisitDate ?? this.lastVisitDate,
      lastTransactionDate: lastTransactionDate ?? this.lastTransactionDate,
      lastTransactionAmount:
          lastTransactionAmount ?? this.lastTransactionAmount,
      lastTransactionItems: lastTransactionItems ?? this.lastTransactionItems,
      isPriority: isPriority ?? this.isPriority,
      assignedEmployeeId: assignedEmployeeId ?? this.assignedEmployeeId,
      assignedEmployeeName: assignedEmployeeName ?? this.assignedEmployeeName,
      distanceCache: distanceCache ?? this.distanceCache,
    );
  }

  bool get hasCoordinates =>
      latitude != null &&
      longitude != null &&
      latitude != 0.0 &&
      longitude != 0.0;
}
