import 'package:equatable/equatable.dart';

class Attendance extends Equatable {
  final String id;
  final String userId;
  final DateTime checkIn;
  final DateTime? checkOut;
  final String status;
  final String branchName;
  final String selfiePath;
  final String clockInPhotoUrl;
  final String clockOutPhotoUrl;
  final int workDuration;
  final bool isOutOfTown;
  final String destination;
  final String notes;

  const Attendance({
    required this.id,
    required this.userId,
    required this.checkIn,
    this.checkOut,
    required this.status,
    this.selfiePath = '',
    this.clockInPhotoUrl = '',
    this.clockOutPhotoUrl = '',
    this.workDuration = 0,
    this.branchName = '',
    this.isOutOfTown = false,
    this.destination = '',
    this.notes = '',
  });

  /// Mengkonversi string waktu dari API ke DateTime dalam WIB (UTC+7).
  ///
  /// Alasan: API backend bisa mengembalikan waktu dengan berbagai format timezone
  /// (UTC "Z", atau WIB "+07:00" tergantung konfigurasi GORM/pgx).
  /// Menggunakan `.toLocal()` menyebabkan bug di device yang timezone-nya diset ke
  /// UTC+8 (WIT/WITA/Malaysia/Singapore) — waktu tampil 1 jam lebih lambat.
  ///
  /// Solusi: Selalu konversi ke UTC terlebih dahulu, lalu tampilkan dalam WIB (UTC+7)
  /// dengan membuat DateTime "naive" yang nilainya sudah berisi jam WIB.
  /// Dengan cara ini, DateFormat('HH:mm') selalu menampilkan jam WIB yang benar
  /// terlepas dari timezone sistem perangkat.
  static DateTime _parseAsWIB(String timeStr) {
    // Konversi ke UTC terlebih dahulu (handles "Z", "+07:00", "+08:00", dll.)
    final utc = DateTime.parse(timeStr).toUtc();
    // Tambahkan offset WIB (UTC+7)
    final wib = utc.add(const Duration(hours: 7));
    // Buat DateTime "naive local" dengan nilai jam/menit WIB
    // sehingga DateFormat.format() langsung membaca .hour dan .minute yang sudah WIB
    return DateTime(
        wib.year, wib.month, wib.day, wib.hour, wib.minute, wib.second);
  }

  factory Attendance.fromJson(Map<String, dynamic> json) {
    final inPhoto = json['clock_in_photo_url'] ?? json['ClockInPhotoURL'] ?? '';
    final outPhoto = json['clock_out_photo_url'] ?? json['ClockOutPhotoURL'] ?? '';
    final selfie = json['selfie_url'] ?? inPhoto;

    return Attendance(
      id: json['id'] ?? '',
      userId: json['user_id'] ?? '',
      checkIn: json['check_in'] != null
          ? _parseAsWIB(json['check_in'])
          : DateTime.now(),
      checkOut:
          json['check_out'] != null ? _parseAsWIB(json['check_out']) : null,
      status: json['status'] ?? 'UNKNOWN',
      selfiePath: selfie,
      clockInPhotoUrl: inPhoto.isNotEmpty ? inPhoto : selfie,
      clockOutPhotoUrl: outPhoto,
      workDuration: json['work_duration'] ?? 0,
      branchName: json['branch_name'] ?? json['BranchName'] ?? '',
      isOutOfTown: json['is_out_of_town'] ?? json['IsOutOfTown'] ?? false,
      destination: json['destination'] ?? json['Destination'] ?? '',
      notes: json['notes'] ?? json['Notes'] ?? '',
    );
  }

  Attendance copyWith({
    String? id,
    String? userId,
    DateTime? checkIn,
    DateTime? checkOut,
    String? status,
    String? branchName,
    String? selfiePath,
    String? clockInPhotoUrl,
    String? clockOutPhotoUrl,
    int? workDuration,
    bool? isOutOfTown,
    String? destination,
    String? notes,
  }) {
    return Attendance(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      checkIn: checkIn ?? this.checkIn,
      checkOut: checkOut ?? this.checkOut,
      status: status ?? this.status,
      branchName: branchName ?? this.branchName,
      selfiePath: selfiePath ?? this.selfiePath,
      clockInPhotoUrl: clockInPhotoUrl ?? this.clockInPhotoUrl,
      clockOutPhotoUrl: clockOutPhotoUrl ?? this.clockOutPhotoUrl,
      workDuration: workDuration ?? this.workDuration,
      isOutOfTown: isOutOfTown ?? this.isOutOfTown,
      destination: destination ?? this.destination,
      notes: notes ?? this.notes,
    );
  }

  @override
  List<Object?> get props => [
        id,
        userId,
        checkIn,
        checkOut,
        status,
        selfiePath,
        clockInPhotoUrl,
        clockOutPhotoUrl,
        workDuration,
        branchName,
        isOutOfTown,
        destination,
        notes,
      ];
}
