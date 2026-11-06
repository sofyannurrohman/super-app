import 'dart:convert';
import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart' show kIsWeb, debugPrint;
import 'package:path_provider/path_provider.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/services/data_integrity_service.dart';

class OfflineDeliveryItem {
  final String localId;
  final String soId;
  final String storeName;
  final String receivedBy;
  final String podImagePath;
  final String courierSigPath;
  final String customerSigPath;
  final String notes;
  final List<Map<String, dynamic>> returnedItems;
  final double amount;
  final String paymentMethod;
  final String? paymentBank;
  final DateTime createdAt;

  OfflineDeliveryItem({
    required this.localId,
    required this.soId,
    required this.storeName,
    required this.receivedBy,
    required this.podImagePath,
    required this.courierSigPath,
    required this.customerSigPath,
    required this.notes,
    required this.returnedItems,
    required this.amount,
    required this.paymentMethod,
    this.paymentBank,
    required this.createdAt,
  });

  Map<String, dynamic> toJson() => {
        'localId': localId,
        'soId': soId,
        'storeName': storeName,
        'receivedBy': receivedBy,
        'podImagePath': podImagePath,
        'courierSigPath': courierSigPath,
        'customerSigPath': customerSigPath,
        'notes': notes,
        'returnedItems': returnedItems,
        'amount': amount,
        'paymentMethod': paymentMethod,
        'paymentBank': paymentBank,
        'createdAt': createdAt.toIso8601String(),
      };

  factory OfflineDeliveryItem.fromJson(Map<String, dynamic> json) =>
      OfflineDeliveryItem(
        localId: json['localId'] ?? '',
        soId: json['soId'] ?? '',
        storeName: json['storeName'] ?? 'Toko',
        receivedBy: json['receivedBy'] ?? '',
        podImagePath: json['podImagePath'] ?? '',
        courierSigPath: json['courierSigPath'] ?? '',
        customerSigPath: json['customerSigPath'] ?? '',
        notes: json['notes'] ?? '',
        returnedItems:
            List<Map<String, dynamic>>.from(json['returnedItems'] ?? []),
        amount: (json['amount'] ?? 0.0 as num).toDouble(),
        paymentMethod: json['paymentMethod'] ?? 'CASH',
        paymentBank: json['paymentBank'],
        createdAt: DateTime.parse(
            json['createdAt'] ?? DateTime.now().toIso8601String()),
      );
}

class OfflineDeliveryManager {
  static final OfflineDeliveryManager _instance =
      OfflineDeliveryManager._internal();
  factory OfflineDeliveryManager() => _instance;
  OfflineDeliveryManager._internal();

  final List<OfflineDeliveryItem> _webQueue = [];

  /// PERBAIKAN KRITIS: Gunakan ApplicationDocumentsDirectory, bukan systemTemp.
  /// systemTemp bisa dihapus oleh OS kapan saja, sehingga data antrean bisa hilang.
  Future<File?> _getQueueFile() async {
    if (kIsWeb) return null;
    try {
      final dir = await getApplicationDocumentsDirectory();
      return File('${dir.path}/offline_delivery_queue.json');
    } catch (e) {
      debugPrint('[OfflineDelivery] Gagal mendapatkan direktori dokumen: $e');
      return null;
    }
  }

  Future<List<OfflineDeliveryItem>> getQueue() async {
    if (kIsWeb) {
      return List.from(_webQueue);
    }
    try {
      final file = await _getQueueFile();
      if (file == null || !await file.exists()) return [];
      final contents = await file.readAsString();
      final List<dynamic> decoded = jsonDecode(contents);
      return decoded.map((e) => OfflineDeliveryItem.fromJson(e)).toList();
    } catch (e) {
      debugPrint('[OfflineDelivery] Error reading offline delivery queue: $e');
      await DataIntegrityService().log(
        category: 'OFFLINE_DELIVERY',
        action: 'QUEUE_READ_ERROR',
        detail: e.toString(),
        severity: AuditSeverity.error,
      );
      return [];
    }
  }

  Future<void> saveQueue(List<OfflineDeliveryItem> queue) async {
    if (kIsWeb) {
      _webQueue.clear();
      _webQueue.addAll(queue);
      return;
    }
    try {
      final file = await _getQueueFile();
      if (file == null) return;
      final encoded = jsonEncode(queue.map((e) => e.toJson()).toList());
      // flush: true memastikan data benar-benar ditulis ke disk, bukan hanya buffer OS
      await file.writeAsString(encoded, flush: true);
      debugPrint(
          '[OfflineDelivery] Queue saved: ${queue.length} items di ${file.path}');
    } catch (e) {
      debugPrint('[OfflineDelivery] Error writing offline delivery queue: $e');
      await DataIntegrityService().log(
        category: 'OFFLINE_DELIVERY',
        action: 'QUEUE_WRITE_ERROR',
        detail: e.toString(),
        severity: AuditSeverity.error,
      );
    }
  }

  Future<void> queueDelivery(OfflineDeliveryItem item) async {
    final queue = await getQueue();
    queue.add(item);
    await saveQueue(queue);
    debugPrint(
        '[OfflineDelivery] Queued offline delivery for store: ${item.storeName}');
    await DataIntegrityService().log(
      category: 'OFFLINE_DELIVERY',
      action: 'ITEM_QUEUED',
      detail:
          'soId=${item.soId} store=${item.storeName} localId=${item.localId}',
    );
  }

  Future<int> getPendingCount() async {
    final queue = await getQueue();
    return queue.length;
  }

  Future<bool> syncPendingDeliveries(ApiClient apiClient) async {
    final queue = await getQueue();
    if (queue.isEmpty) return true;

    print('DEBUG OFFLINE: Starting sync of ${queue.length} pending deliveries');
    List<OfflineDeliveryItem> failedItems = [];

    for (final item in queue) {
      try {
        print('DEBUG OFFLINE: Syncing delivery for store: ${item.storeName}');

        // 1. Upload POD Image
        String uploadedPodUrl = '';
        if (item.podImagePath.isNotEmpty) {
          if (item.podImagePath.startsWith('base64:')) {
            final base64String = item.podImagePath.split(':').last;
            final bytes = base64Decode(base64String);
            final filePayload =
                MultipartFile.fromBytes(bytes, filename: 'pod.png');
            final formData = FormData.fromMap({'file': filePayload});
            final response =
                await apiClient.client.post('sales/upload', data: formData);
            if (response.data != null && response.data['url'] != null) {
              uploadedPodUrl = response.data['url'];
            }
          } else if (!kIsWeb && File(item.podImagePath).existsSync()) {
            final filePayload = await MultipartFile.fromFile(item.podImagePath);
            final formData = FormData.fromMap({'file': filePayload});
            final response =
                await apiClient.client.post('sales/upload', data: formData);
            if (response.data != null && response.data['url'] != null) {
              uploadedPodUrl = response.data['url'];
            }
          }
        }

        // 2. Upload Courier Signature
        String uploadedCourierUrl = '';
        if (item.courierSigPath.isNotEmpty) {
          if (item.courierSigPath.startsWith('base64:')) {
            final base64String = item.courierSigPath.split(':').last;
            final bytes = base64Decode(base64String);
            final filePayload =
                MultipartFile.fromBytes(bytes, filename: 'courier_sig.png');
            final formData = FormData.fromMap({'file': filePayload});
            final response =
                await apiClient.client.post('sales/upload', data: formData);
            if (response.data != null && response.data['url'] != null) {
              uploadedCourierUrl = response.data['url'];
            }
          } else if (!kIsWeb && File(item.courierSigPath).existsSync()) {
            final filePayload =
                await MultipartFile.fromFile(item.courierSigPath);
            final formData = FormData.fromMap({'file': filePayload});
            final response =
                await apiClient.client.post('sales/upload', data: formData);
            if (response.data != null && response.data['url'] != null) {
              uploadedCourierUrl = response.data['url'];
            }
          }
        }

        // 3. Upload Customer Signature
        String uploadedCustomerUrl = '';
        if (item.customerSigPath.isNotEmpty) {
          if (item.customerSigPath.startsWith('base64:')) {
            final base64String = item.customerSigPath.split(':').last;
            final bytes = base64Decode(base64String);
            final filePayload =
                MultipartFile.fromBytes(bytes, filename: 'customer_sig.png');
            final formData = FormData.fromMap({'file': filePayload});
            final response =
                await apiClient.client.post('sales/upload', data: formData);
            if (response.data != null && response.data['url'] != null) {
              uploadedCustomerUrl = response.data['url'];
            }
          } else if (!kIsWeb && File(item.customerSigPath).existsSync()) {
            final filePayload =
                await MultipartFile.fromFile(item.customerSigPath);
            final formData = FormData.fromMap({'file': filePayload});
            final response =
                await apiClient.client.post('sales/upload', data: formData);
            if (response.data != null && response.data['url'] != null) {
              uploadedCustomerUrl = response.data['url'];
            }
          }
        }

        // 4. Combine Signatures URL
        String combinedSignatureUrl = '';
        if (uploadedCourierUrl.isNotEmpty && uploadedCustomerUrl.isNotEmpty) {
          combinedSignatureUrl = '$uploadedCourierUrl,$uploadedCustomerUrl';
        } else if (uploadedCourierUrl.isNotEmpty) {
          combinedSignatureUrl = uploadedCourierUrl;
        } else if (uploadedCustomerUrl.isNotEmpty) {
          combinedSignatureUrl = uploadedCustomerUrl;
        }

        // 5. POST Confirm Delivery to Backend
        await apiClient.client.post(
          'sales-orders/confirm-delivery',
          data: {
            'so_id': item.soId,
            'received_by': item.receivedBy.isNotEmpty
                ? item.receivedBy
                : 'Toko (Offline Fallback)',
            'pod_image_url': uploadedPodUrl,
            'signature_image_url': combinedSignatureUrl,
            'notes': '${item.notes} [Offline Sync]',
            'returned_items': item.returnedItems,
          },
        );

        // 6. POST Collect Payment to Backend
        if (item.amount > 0 || item.paymentMethod == 'TEMPO') {
          await apiClient.client.post(
            'sales-orders/collect-payment',
            data: {
              'so_id': item.soId,
              'amount': item.paymentMethod == 'TEMPO' ? 0.0 : item.amount,
              'payment_method': item.paymentMethod,
              'payment_bank': item.paymentBank,
              'collected_by': 'Driver Delivery [Offline Sync]',
            },
          );
        }

        print(
            'DEBUG OFFLINE: Successfully synced delivery for store: ${item.storeName}');

        // Remove locally stored image files to keep disk clean
        if (!kIsWeb) {
          try {
            if (item.podImagePath.isNotEmpty &&
                !item.podImagePath.startsWith('base64:') &&
                File(item.podImagePath).existsSync())
              File(item.podImagePath).deleteSync();
            if (item.courierSigPath.isNotEmpty &&
                !item.courierSigPath.startsWith('base64:') &&
                File(item.courierSigPath).existsSync())
              File(item.courierSigPath).deleteSync();
            if (item.customerSigPath.isNotEmpty &&
                !item.customerSigPath.startsWith('base64:') &&
                File(item.customerSigPath).existsSync())
              File(item.customerSigPath).deleteSync();
          } catch (_) {}
        }
      } catch (e) {
        print(
            'DEBUG OFFLINE: Failed to sync delivery for ${item.storeName}: $e');
        failedItems.add(item); // Keep in queue for next attempt
      }
    }

    await saveQueue(failedItems);
    return failedItems.isEmpty;
  }
}
