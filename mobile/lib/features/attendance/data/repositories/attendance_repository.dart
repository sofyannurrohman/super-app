import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:camera/camera.dart' show XFile;
import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';

import 'package:dio/dio.dart';
import 'package:geolocator/geolocator.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/attendance/domain/entities/attendance.dart';
import 'package:hris_app/core/services/local_database_service.dart';

class AttendanceStats {
  final double totalHours;
  final double overtimeHours;
  final int attendanceDays;
  final int leaveDays;

  AttendanceStats({
    required this.totalHours,
    required this.overtimeHours,
    required this.attendanceDays,
    required this.leaveDays,
  });

  factory AttendanceStats.fromJson(Map<String, dynamic> json) {
    return AttendanceStats(
      totalHours: (json['total_hours'] as num?)?.toDouble() ?? 0.0,
      overtimeHours: (json['overtime_hours'] as num?)?.toDouble() ?? 0.0,
      attendanceDays: json['attendance_days'] ?? 0,
      leaveDays: json['leave_days'] ?? 0,
    );
  }
}

class AttendanceRepository {
  final ApiClient apiClient;
  final LocalDatabaseService localDb;

  AttendanceRepository({required this.apiClient, required this.localDb});

  Future<Position> _getCurrentLocation() async {
    bool serviceEnabled;
    LocationPermission permission;

    serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) {
      throw Exception('Location services are disabled.');
    }

    permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) {
        throw Exception('Location permissions are denied');
      }
    }

    if (permission == LocationPermission.deniedForever) {
      throw Exception(
          'Location permissions are permanently denied, we cannot request permissions.');
    }

    return await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high);
  }

  String _extractDioErrorMessage(DioException e) {
    final data = e.response?.data;
    if (data is Map) {
      final msg = data['error'] ?? data['message'];
      if (msg != null && msg.toString().trim().isNotEmpty) {
        return msg.toString();
      }
    } else if (data is String && data.trim().isNotEmpty) {
      try {
        final decoded = jsonDecode(data);
        if (decoded is Map) {
          final msg = decoded['error'] ?? decoded['message'];
          if (msg != null && msg.toString().trim().isNotEmpty) {
            return msg.toString();
          }
        }
      } catch (_) {}
    }

    final statusCode = e.response?.statusCode ?? 0;
    if (statusCode == 502 || statusCode == 503 || statusCode == 504) {
      return 'Server sedang mengalami gangguan atau pemeliharaan (HTTP $statusCode). Silakan coba beberapa saat lagi.';
    }
    if (statusCode == 413) {
      return 'Ukuran foto presensi terlalu besar untuk dikirim ke server.';
    }
    if (statusCode >= 500) {
      return 'Terjadi kesalahan pada server (HTTP $statusCode). Silakan coba lagi.';
    }

    if (e.type == DioExceptionType.connectionTimeout ||
        e.type == DioExceptionType.sendTimeout ||
        e.type == DioExceptionType.receiveTimeout) {
      return 'Waktu koneksi habis saat mengirim data presensi. Pastikan sinyal internet Anda stabil lalu coba lagi.';
    }
    if (e.type == DioExceptionType.connectionError) {
      return 'Tidak dapat terhubung ke server. Periksa koneksi internet Anda.';
    }

    return 'Terjadi kesalahan jaringan';
  }

  Future<Either<Failure, Attendance>> checkIn(double lat, double lng,
      String? photoPath, List<double>? faceEmbedding,
      {bool isOutOfTown = false, String destination = ''}) async {
    try {
      String? base64Photo;
      if (photoPath != null) {
        final xFile = XFile(photoPath);
        final bytes = await xFile.readAsBytes();
        base64Photo = base64Encode(bytes);
      }

      final data = {
        "latitude": lat,
        "longitude": lng,
        "selfie": base64Photo,
        "face_embedding": faceEmbedding ?? [],
        "is_out_of_town": isOutOfTown,
        "destination": destination,
      };

      final response = await apiClient.client.post(
        'attendance/checkin',
        data: data,
        options: Options(
          sendTimeout: const Duration(seconds: 25),
          receiveTimeout: const Duration(seconds: 25),
        ),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        final attendanceData = response.data['data'];
        if (attendanceData != null) {
          return Right(Attendance.fromJson(attendanceData));
        }
      }
      return const Left(
          ServerFailure('Gagal memproses data kehadiran dari server'));
    } on DioException catch (e) {
      return Left(ServerFailure(_extractDioErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  Future<Either<Failure, Attendance>> checkOut(double lat, double lng,
      String? photoPath, List<double>? faceEmbedding,
      {bool isOutOfTown = false, String destination = ''}) async {
    try {
      String? base64Photo;
      if (photoPath != null) {
        final xFile = XFile(photoPath);
        final bytes = await xFile.readAsBytes();
        base64Photo = base64Encode(bytes);
      }

      final data = {
        "latitude": lat,
        "longitude": lng,
        "selfie": base64Photo,
        "face_embedding": faceEmbedding ?? [],
        "is_out_of_town": isOutOfTown,
        "destination": destination,
      };

      final response = await apiClient.client.post(
        'attendance/checkout',
        data: data,
        options: Options(
          sendTimeout: const Duration(seconds: 25),
          receiveTimeout: const Duration(seconds: 25),
        ),
      );

      if (response.statusCode == 200) {
        final attendanceData = response.data['data'];
        if (attendanceData != null) {
          return Right(Attendance.fromJson(attendanceData));
        }
      }

      return const Left(
          ServerFailure('Gagal memproses data kehadiran dari server'));
    } on DioException catch (e) {
      return Left(ServerFailure(_extractDioErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  Future<Either<Failure, void>> updateAttendanceNotes(
      String attendanceId, String notes) async {
    try {
      final response = await apiClient.client.put(
        'attendance/my-notes/$attendanceId',
        data: {'notes': notes},
      );

      if (response.statusCode == 200) {
        return const Right(null);
      }
      return const Left(ServerFailure('Gagal menyimpan catatan keterlambatan'));
    } on DioException catch (e) {
      return Left(ServerFailure(_extractDioErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  Future<Either<Failure, List<Attendance>>> getHistory(
      {String? startDate, String? endDate}) async {
    try {
      final response =
          await apiClient.client.get('attendance/history', queryParameters: {
        if (startDate != null) 'start_date': startDate,
        if (endDate != null) 'end_date': endDate,
      });

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        if (startDate == null && endDate == null) {
          try {
            final prefs = await SharedPreferences.getInstance();
            await prefs.setString(
                'cached_attendance_history', jsonEncode(data));
          } catch (_) {}
        }
        return Right(data.map((e) => Attendance.fromJson(e)).toList());
      }
      if (startDate == null && endDate == null) {
        final prefs = await SharedPreferences.getInstance();
        final cached = prefs.getString('cached_attendance_history');
        if (cached != null) {
          try {
            final List decoded = jsonDecode(cached);
            return Right(decoded.map((e) => Attendance.fromJson(e)).toList());
          } catch (_) {}
        }
      }
      return const Right([]);
    } catch (e) {
      try {
        final prefs = await SharedPreferences.getInstance();
        final cached = prefs.getString('cached_attendance_history');
        if (cached != null) {
          final List decoded = jsonDecode(cached);
          return Right(decoded.map((e) => Attendance.fromJson(e)).toList());
        }
      } catch (_) {}
      return Left(ServerFailure(e.toString()));
    }
  }

  Future<Either<Failure, AttendanceStats>> fetchStats() async {
    try {
      final response = await apiClient.client.get('attendance/stats');
      if (response.statusCode == 200) {
        final stats = AttendanceStats.fromJson(response.data['data']);
        try {
          final prefs = await SharedPreferences.getInstance();
          await prefs.setString(
              'cached_attendance_stats', jsonEncode(response.data['data']));
        } catch (_) {}
        return Right(stats);
      }
      final prefs = await SharedPreferences.getInstance();
      final cached = prefs.getString('cached_attendance_stats');
      if (cached != null) {
        try {
          return Right(AttendanceStats.fromJson(jsonDecode(cached)));
        } catch (_) {}
      }
      return Right(AttendanceStats(
          totalHours: 0, overtimeHours: 0, attendanceDays: 0, leaveDays: 0));
    } catch (e) {
      try {
        final prefs = await SharedPreferences.getInstance();
        final cached = prefs.getString('cached_attendance_stats');
        if (cached != null) {
          return Right(AttendanceStats.fromJson(jsonDecode(cached)));
        }
        return Right(AttendanceStats(
            totalHours: 0, overtimeHours: 0, attendanceDays: 0, leaveDays: 0));
      } catch (_) {}
      return Right(AttendanceStats(
          totalHours: 0, overtimeHours: 0, attendanceDays: 0, leaveDays: 0));
    }
  }

  Future<int> getQueueCount() async {
    return await localDb.getQueueCount();
  }

  Future<void> saveOffline(String type, String photoPath, double lat,
      double lng, List<double> embedding) async {
    await localDb.insertAttendance({
      'type': type,
      'photoPath': photoPath,
      'latitude': lat,
      'longitude': lng,
      'embedding': jsonEncode(embedding),
      'timestamp': DateTime.now().toIso8601String(),
    });
  }

  Future<Map<String, int>> syncQueue() async {
    final queue = await localDb.getAttendanceQueue();
    int success = 0;
    int fail = 0;

    for (var item in queue) {
      try {
        final type = item['type']?.toString().toLowerCase();
        final photoPath = item['photoPath']?.toString();
        final lat = (item['latitude'] as num?)?.toDouble() ?? 0.0;
        final lng = (item['longitude'] as num?)?.toDouble() ?? 0.0;
        List<double>? embedding;
        if (item['embedding'] != null) {
          try {
            final List raw = jsonDecode(item['embedding']);
            embedding = raw.map((e) => (e as num).toDouble()).toList();
          } catch (_) {}
        }

        Either<Failure, Attendance> result;
        if (type == 'check_in' || type == 'clock_in') {
          result = await checkIn(lat, lng, photoPath, embedding);
        } else {
          result = await checkOut(lat, lng, photoPath, embedding);
        }

        if (result.isRight()) {
          success++;
          await localDb.deleteAttendance(item['id']);
        } else {
          fail++;
        }
      } catch (e) {
        fail++;
      }
    }
    return {'success': success, 'fail': fail};
  }
}
