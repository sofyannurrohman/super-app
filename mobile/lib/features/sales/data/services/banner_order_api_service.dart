import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';
import 'package:hris_app/core/network/api_client.dart';

class BannerOrderApiService {
  final ApiClient apiClient;

  BannerOrderApiService({required this.apiClient});

  // POST /api/v1/branding/banners
  Future<void> createBannerOrder(
    Map<String, dynamic> data, {
    String? attachmentPath,
    Uint8List? attachmentBytes,
    String? attachmentName,
  }) async {
    // Keep one ID across the first request and every offline retry.
    final orderData = Map<String, dynamic>.from(data);
    orderData['id'] ??= const Uuid().v4();
    try {
      if (attachmentBytes != null && attachmentBytes.isNotEmpty) {
        final map = Map<String, dynamic>.from(orderData);
        map['attachment'] = MultipartFile.fromBytes(
          attachmentBytes,
          filename: attachmentName ?? 'attachment.jpg',
        );
        final formData = FormData.fromMap(map);
        await apiClient.client.post('branding/banners', data: formData);
      } else if (attachmentPath != null && attachmentPath.isNotEmpty && !kIsWeb) {
        final map = Map<String, dynamic>.from(orderData);
        map['attachment'] = await MultipartFile.fromFile(attachmentPath);
        final formData = FormData.fromMap(map);
        await apiClient.client.post('branding/banners', data: formData);
      } else {
        await apiClient.client.post('branding/banners', data: orderData);
      }
    } catch (e) {
      bool isNetworkError = false;
      if (e is DioException) {
        final type = e.type;
        if (type == DioExceptionType.connectionTimeout ||
            type == DioExceptionType.sendTimeout ||
            type == DioExceptionType.receiveTimeout ||
            type == DioExceptionType.connectionError ||
            (e.error != null &&
                e.error.toString().toLowerCase().contains('socketexception'))) {
          isNetworkError = true;
        }
      } else {
        final errStr = e.toString().toLowerCase();
        if (errStr.contains('socketexception') ||
            errStr.contains('network') ||
            errStr.contains('failed to connect')) {
          isNetworkError = true;
        }
      }

      if (isNetworkError) {
        try {
          final prefs = await SharedPreferences.getInstance();
          final String? existingJson = prefs.getString('offline_banner_orders');
          List<dynamic> list = [];
          if (existingJson != null) {
            list = jsonDecode(existingJson);
          }
          if ((attachmentBytes != null && attachmentBytes.isNotEmpty) ||
              (attachmentPath != null && attachmentPath.isNotEmpty)) {
            orderData['offline_attachment_name'] = attachmentName ?? 'attachment.jpg';
            if (kIsWeb) {
              if (attachmentBytes == null || attachmentBytes.isEmpty) {
                throw Exception('Lampiran offline tidak tersedia');
              }
              orderData['offline_attachment_base64'] = base64Encode(attachmentBytes);
            } else {
              final directory = await getApplicationDocumentsDirectory();
              final savedFile = File('${directory.path}/banner_${orderData['id']}_attachment');
              if (attachmentBytes != null && attachmentBytes.isNotEmpty) {
                await savedFile.writeAsBytes(attachmentBytes, flush: true);
              } else {
                await File(attachmentPath!).copy(savedFile.path);
              }
              orderData['offline_attachment_path'] = savedFile.path;
            }
          }
          list.add(orderData);
          await prefs.setString('offline_banner_orders', jsonEncode(list));
          print('DEBUG OFFLINE: Banner order saved offline (network error)');
          throw Exception(
              'OFFLINE_SAVED: Request spanduk disimpan secara offline karena tidak ada koneksi internet');
        } catch (saveErr) {
          print('DEBUG OFFLINE: Failed to save banner order offline: $saveErr');
          if (saveErr.toString().contains('OFFLINE_SAVED')) {
            rethrow;
          }
        }
      }

      if (e is DioException) {
        final resData = e.response?.data;
        if (resData is Map && resData.containsKey('error')) {
          throw Exception(resData['error'].toString());
        } else if (resData is Map && resData.containsKey('message')) {
          throw Exception(resData['message'].toString());
        } else if (e.response?.statusCode != null) {
          throw Exception('Gagal membuat order spanduk (Status ${e.response?.statusCode})');
        } else if (e.message != null && e.message!.isNotEmpty) {
          throw Exception(e.message!);
        }
        throw Exception('Gagal membuat order spanduk');
      }
      throw Exception('Gagal membuat order spanduk: $e');
    }
  }

  // GET /api/v1/branding/banners
  Future<List<dynamic>> getBanners(
      {String? status, String? branchId, int? limit = 99999}) async {
    try {
      final Map<String, dynamic> query = {};
      if (status != null && status.isNotEmpty) query['status'] = status;
      if (branchId != null &&
          branchId.isNotEmpty &&
          branchId != 'ALL_BRANCHES') {
        query['branch_id'] = branchId;
      }
      if (limit != null) {
        query['limit'] = limit.toString();
      }
      final response = await apiClient.client
          .get('branding/banners', queryParameters: query);
      return response.data['data'] ?? [];
    } catch (e) {
      if (e is DioException) {
        final resData = e.response?.data;
        if (resData is Map && resData.containsKey('error')) {
          throw Exception(resData['error'].toString());
        } else if (resData is Map && resData.containsKey('message')) {
          throw Exception(resData['message'].toString());
        } else if (e.response?.statusCode != null) {
          throw Exception('Gagal memuat data banner (Status ${e.response?.statusCode})');
        }
        throw Exception(e.message ?? 'Gagal memuat data banner');
      }
      throw Exception('Gagal memuat data banner: $e');
    }
  }

  // GET /api/v1/branding/banners (paginated version for infinite scroll)
  // Returns {data: List, total: int, page: int, limit: int}
  Future<Map<String, dynamic>> getBannersPaginated({
    String? status,
    String? branchId,
    int page = 1,
    int limit = 20,
    String? search,
  }) async {
    try {
      final Map<String, dynamic> query = {
        'page': page,
        'limit': limit,
      };
      if (status != null && status.isNotEmpty) query['status'] = status;
      if (branchId != null &&
          branchId.isNotEmpty &&
          branchId != 'ALL_BRANCHES') {
        query['branch_id'] = branchId;
      }
      if (search != null && search.isNotEmpty) query['search'] = search;

      final response = await apiClient.client
          .get('branding/banners', queryParameters: query);
      return {
        'data': response.data['data'] ?? [],
        'total': (response.data['total'] as num?)?.toInt() ?? 0,
        'page': (response.data['page'] as num?)?.toInt() ?? page,
        'limit': (response.data['limit'] as num?)?.toInt() ?? limit,
      };
    } catch (e) {
      if (e is DioException) {
        final resData = e.response?.data;
        if (resData is Map && resData.containsKey('error')) {
          throw Exception(resData['error'].toString());
        } else if (resData is Map && resData.containsKey('message')) {
          throw Exception(resData['message'].toString());
        } else if (e.response?.statusCode != null) {
          throw Exception('Gagal memuat data banner (Status ${e.response?.statusCode})');
        }
        throw Exception(e.message ?? 'Gagal memuat data banner');
      }
      throw Exception('Gagal memuat data banner: $e');
    }
  }

  // GET all banners in bounded pages without overloading the database.
  Future<List<dynamic>> getAllBannersChunked({
    String? status,
    String? branchId,
    int pageSize = 500,
  }) async {
    final firstPage = await getBannersPaginated(
      status: status,
      branchId: branchId,
      page: 1,
      limit: pageSize,
    );
    final total = (firstPage['total'] as int?) ?? 0;
    final List<dynamic> allData = List.from(firstPage['data'] as List? ?? []);

    if (allData.length >= total || total <= pageSize) {
      return allData;
    }

    final totalPages = (total / pageSize).ceil();
    for (int p = 2; p <= totalPages; p++) {
      final pageRes = await getBannersPaginated(
        status: status,
        branchId: branchId,
        page: p,
        limit: pageSize,
      );
      final list = pageRes['data'] as List? ?? [];
      allData.addAll(list);
    }

    return allData;
  }

  // GET /api/v1/branding/operator/performance
  Future<Map<String, dynamic>?> getOperatorPerformance(dynamic operatorId) async {
    try {
      final response =
          await apiClient.client.get('branding/operator/performance');
      final data = response.data['data'] as List;
      for (var item in data) {
        if (item['id']?.toString() == operatorId?.toString()) {
          return item as Map<String, dynamic>;
        }
      }
      return null;
    } catch (e) {
      if (e is DioException) {
        throw Exception(
            e.response?.data['error'] ?? 'Gagal memuat performa operator');
      }
      throw Exception('Gagal memuat performa operator: $e');
    }
  }

  // POST /api/v1/branding/banners/:id/install
  Future<void> installBanner(
      dynamic bannerId, String imagePath, double lat, double lng,
      {Uint8List? imageBytes, String? filename}) async {
    try {
      MultipartFile file;
      if (imageBytes != null && imageBytes.isNotEmpty) {
        file = MultipartFile.fromBytes(
          imageBytes,
          filename: filename ?? 'photo_installed.jpg',
        );
      } else if (!kIsWeb && imagePath.isNotEmpty) {
        file = await MultipartFile.fromFile(imagePath, filename: filename);
      } else {
        throw Exception('Data gambar pemasangan tidak valid');
      }

      FormData formData = FormData.fromMap({
        'latitude': lat.toString(),
        'longitude': lng.toString(),
        'photo_installed': file,
      });
      await apiClient.client
          .post('branding/banners/$bannerId/install', data: formData);
    } catch (e) {
      if (e is DioException) {
        throw Exception(e.response?.data['error'] ??
            'Gagal menyelesaikan pemasangan banner');
      }
      throw Exception('Gagal menyelesaikan pemasangan banner: $e');
    }
  }

  // PUT /api/v1/branding/banners/:id
  Future<void> updateBannerOrder(
      dynamic bannerId, Map<String, dynamic> data) async {
    final String idStr = bannerId.toString();
    if (idStr.startsWith('OFFLINE-')) {
      final prefs = await SharedPreferences.getInstance();
      final String? existingJson = prefs.getString('offline_banner_orders');
      if (existingJson != null) {
        List<dynamic> list = jsonDecode(existingJson);
        for (int i = 0; i < list.length; i++) {
          if (list[i]['id'] == idStr) {
            list[i].addAll(data);
            break;
          }
        }
        await prefs.setString('offline_banner_orders', jsonEncode(list));
      }
      return;
    }

    try {
      await apiClient.client.put('branding/banners/$bannerId', data: data);
    } catch (e) {
      if (e is DioException) {
        throw Exception(
            e.response?.data['error'] ?? 'Gagal memperbarui order spanduk');
      }
      throw Exception('Gagal memperbarui order spanduk: $e');
    }
  }

  // DELETE /api/v1/branding/banners/:id
  Future<void> deleteBannerOrder(dynamic bannerId) async {
    final String idStr = bannerId.toString();
    if (idStr.startsWith('OFFLINE-')) {
      final prefs = await SharedPreferences.getInstance();
      final String? existingJson = prefs.getString('offline_banner_orders');
      if (existingJson != null) {
        List<dynamic> list = jsonDecode(existingJson);
        list.removeWhere((item) => item['id'] == idStr);
        await prefs.setString('offline_banner_orders', jsonEncode(list));
      }
      return;
    }

    try {
      await apiClient.client.delete('branding/banners/$bannerId');
    } catch (e) {
      if (e is DioException) {
        throw Exception(
            e.response?.data['error'] ?? 'Gagal menghapus order spanduk');
      }
      throw Exception('Gagal menghapus order spanduk: $e');
    }
  }

  // PUT /api/v1/branding/banners/:id/operator-take
  Future<void> takeBanner(dynamic bannerId) async {
    try {
      await apiClient.client.put('branding/banners/$bannerId/operator-take');
    } catch (e) {
      if (e is DioException) {
        throw Exception(e.response?.data['error'] ?? 'Gagal mengambil spanduk');
      }
      throw Exception('Gagal mengambil spanduk: $e');
    }
  }

  // GET /api/v1/branding/sales/unclaimed-installations
  Future<List<dynamic>> getUnclaimedInstallations({String? date}) async {
    try {
      final Map<String, dynamic> query = {};
      if (date != null && date.isNotEmpty) {
        query['date'] = date;
      }
      final response = await apiClient.client
          .get('branding/sales/unclaimed-installations', queryParameters: query);
      return response.data['data'] ?? [];
    } catch (e) {
      if (e is DioException) {
        throw Exception(e.response?.data['error'] ??
            'Gagal memuat daftar spanduk belum direimburse');
      }
      throw Exception('Gagal memuat daftar spanduk: $e');
    }
  }

  // POST /api/v1/branding/sales/direct-install
  Future<Map<String, dynamic>> directInstallBanner({
    required String shopName,
    required double lat,
    required double lng,
    required String imagePath,
    Uint8List? imageBytes,
    String? filename,
    String bannerType = 'Standard',
    String? notes,
    int qty = 1,
    double sizeLength = 3.0,
    double sizeWidth = 1.0,
  }) async {
    try {
      MultipartFile file;
      if (imageBytes != null && imageBytes.isNotEmpty) {
        file = MultipartFile.fromBytes(
          imageBytes,
          filename: filename ?? 'photo_installed.jpg',
        );
      } else if (!kIsWeb && imagePath.isNotEmpty) {
        file = await MultipartFile.fromFile(imagePath, filename: filename);
      } else {
        throw Exception('Foto bukti pemasangan tidak valid');
      }

      FormData formData = FormData.fromMap({
        'shop_name': shopName,
        'latitude': lat.toString(),
        'longitude': lng.toString(),
        'banner_type': bannerType,
        'notes': notes ?? '',
        'qty': qty.toString(),
        'size_length': sizeLength.toString(),
        'size_width': sizeWidth.toString(),
        'photo_installed': file,
      });

      final response = await apiClient.client
          .post('branding/sales/direct-install', data: formData);
      return response.data;
    } catch (e) {
      if (e is DioException) {
        throw Exception(e.response?.data['error'] ??
            'Gagal menyimpan laporan pasang spanduk');
      }
      throw Exception('Gagal menyimpan laporan pasang spanduk: $e');
    }
  }

  // POST /api/v1/branding/sales/bulk-reimburse
  Future<Map<String, dynamic>> bulkReimburseInstallations({
    List<String>? bannerIds,
    String? notes,
  }) async {
    try {
      final response = await apiClient.client.post(
        'branding/sales/bulk-reimburse',
        data: {
          'banner_ids': bannerIds ?? [],
          'notes': notes ?? '',
        },
      );
      return response.data;
    } catch (e) {
      if (e is DioException) {
        throw Exception(e.response?.data['error'] ??
            'Gagal mengajukan reimbursement spanduk');
      }
      throw Exception('Gagal mengajukan reimbursement spanduk: $e');
    }
  }
}
