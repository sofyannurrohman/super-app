class StoreModel {
  final String id;
  final String name;
  final String ownerName;
  final String phoneNumber;
  final String address;
  final bool isNew;
  final double latitude;
  final double longitude;
  final String photoUrl;
  final String province;
  final String regency;
  final String district;
  final String village;

  final List<int> visitDays;
  final String visitFrequency;
  final String? assignedEmployeeId;
  final String? assignedEmployeeName;
  final String? createdByEmployeeId;
  final bool isOffline;
  final String syncStatus;
  final String? localPhotoPath;
  final double? accuracy;
  final String customerType;

  StoreModel({
    required this.id,
    required this.name,
    required this.ownerName,
    this.phoneNumber = '',
    required this.address,
    this.isNew = false,
    this.latitude = 0.0,
    this.longitude = 0.0,
    this.photoUrl = '',
    this.province = '',
    this.regency = '',
    this.district = '',
    this.village = '',
    this.visitDays = const [],
    this.visitFrequency = '',
    this.assignedEmployeeId,
    this.assignedEmployeeName,
    this.createdByEmployeeId,
    this.isOffline = false,
    this.syncStatus = 'synced',
    this.localPhotoPath,
    this.accuracy,
    this.customerType = 'TOKO',
  });

  /// Returns server [photoUrl] if available, otherwise falls back to [localPhotoPath] (local file path or data:image base64 URI).
  String get effectivePhoto {
    if (photoUrl.isNotEmpty) return photoUrl;
    if (localPhotoPath != null && localPhotoPath!.isNotEmpty) {
      return localPhotoPath!;
    }
    return '';
  }

  factory StoreModel.fromJson(Map<String, dynamic> json) {
    return StoreModel(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      ownerName: json['owner_name'] ?? '',
      phoneNumber: json['phone_number'] ?? '',
      address: json['address'] ?? '',
      latitude: (json['latitude'] ?? 0.0).toDouble(),
      longitude: (json['longitude'] ?? 0.0).toDouble(),
      isNew: json['first_transaction_date'] == null,
      visitDays: List<int>.from(json['visit_days'] ?? []),
      visitFrequency: json['visit_frequency'] ?? '',
      photoUrl: json['photo_url'] ?? '',
      province: json['province'] ?? '',
      regency: json['regency'] ?? '',
      district: json['district'] ?? '',
      village: json['village'] ?? '',
      assignedEmployeeId: json['assigned_employee_id']?.toString(),
      assignedEmployeeName: json['assigned_employee'] != null
          ? ((json['assigned_employee']['first_name'] != null ||
                  json['assigned_employee']['last_name'] != null)
              ? '${json['assigned_employee']['first_name'] ?? ''} ${json['assigned_employee']['last_name'] ?? ''}'
                  .trim()
              : (json['assigned_employee']['name']?.toString() ?? ''))
          : (json['assigned_employee_name']?.toString() ?? ''),
      createdByEmployeeId: json['created_by_employee_id']?.toString(),
      isOffline: json['is_offline'] == true,
      syncStatus: json['sync_status'] ?? 'synced',
      localPhotoPath: json['local_photo_path'],
      accuracy: json['accuracy'] != null ? (json['accuracy'] as num).toDouble() : null,
      customerType: json['customer_type']?.toString() ?? 'TOKO',
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'owner_name': ownerName,
        'phone_number': phoneNumber,
        'address': address,
        'latitude': latitude,
        'longitude': longitude,
        'photo_url': photoUrl,
        'province': province,
        'regency': regency,
        'district': district,
        'village': village,
        'visit_days': visitDays,
        'visit_frequency': visitFrequency,
        if (assignedEmployeeId != null && assignedEmployeeId!.isNotEmpty)
          'assigned_employee_id': assignedEmployeeId,
        if (assignedEmployeeName != null && assignedEmployeeName!.isNotEmpty)
          'assigned_employee_name': assignedEmployeeName,
        if (createdByEmployeeId != null && createdByEmployeeId!.isNotEmpty)
          'created_by_employee_id': createdByEmployeeId,
        'is_offline': isOffline,
        'sync_status': syncStatus,
        'customer_type': customerType,
        if (localPhotoPath != null) 'local_photo_path': localPhotoPath,
        if (accuracy != null) 'accuracy': accuracy,
        // we just use isNew to set first_transaction_date for fromJson compatibility
        'first_transaction_date': isNew ? null : 'not_null',
      };
}
