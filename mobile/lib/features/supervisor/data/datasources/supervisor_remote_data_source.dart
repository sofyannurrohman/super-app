import 'package:dio/dio.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/supervisor/data/models/sales_transfer_model.dart';
import 'package:hris_app/features/supervisor/data/models/delivery_batch_model.dart';
import 'package:hris_app/features/supervisor/data/models/sales_kpi_model.dart';

abstract class SupervisorRemoteDataSource {
  Future<List<SalesTransferModel>> getPendingTransfers(String branchId);
  Future<void> approveTransfer(String transferId, int approvedQuantity);
  Future<void> rejectTransfer(String transferId);

  Future<List<DeliveryBatchModel>> getPendingDeliveryBatches(String branchId);
  Future<void> approveDeliveryBatch(String batchId);
  Future<void> cancelDeliveryBatch(String batchId);

  Future<List<SalesKpiModel>> getSalesKPIs(
      int month, int year, String branchId);
}

class SupervisorRemoteDataSourceImpl implements SupervisorRemoteDataSource {
  final ApiClient dioClient;

  SupervisorRemoteDataSourceImpl({required this.dioClient});

  @override
  Future<List<SalesTransferModel>> getPendingTransfers(String branchId) async {
    try {
      final response = await dioClient.dio
          .get('/sales-transfers/supervisor/pending', queryParameters: {
        if (branchId.isNotEmpty &&
            branchId != 'ALL_BRANCHES' &&
            branchId != '0' &&
            branchId != 'all')
          'branch_id': branchId,
      });
      final responseData = response.data;
      final List data = (responseData is List)
          ? responseData
          : (responseData is Map ? (responseData['data'] ?? []) : []);
      return data.map((json) => SalesTransferModel.fromJson(json)).toList();
    } catch (e) {
      rethrow;
    }
  }

  @override
  Future<void> approveTransfer(String transferId, int approvedQuantity) async {
    try {
      await dioClient.dio
          .patch('/sales-transfers/$transferId/approve-supervisor', data: {
        'approved_quantity': approvedQuantity,
      });
    } catch (e) {
      rethrow;
    }
  }

  @override
  Future<void> rejectTransfer(String transferId) async {
    try {
      await dioClient.dio
          .patch('/sales-transfers/$transferId/reject-supervisor');
    } catch (e) {
      rethrow;
    }
  }

  @override
  Future<List<DeliveryBatchModel>> getPendingDeliveryBatches(
      String branchId) async {
    try {
      final response =
          await dioClient.dio.get('/delivery/batches', queryParameters: {
        if (branchId.isNotEmpty &&
            branchId != 'ALL_BRANCHES' &&
            branchId != '0' &&
            branchId != 'all')
          'branch_id': branchId,
      });
      final responseData = response.data;
      final List data = (responseData is List)
          ? responseData
          : (responseData is Map ? (responseData['data'] ?? []) : []);
      // Filter for WAITING_APPROVAL
      var batches = data
          .map((json) => DeliveryBatchModel.fromJson(json))
          .where((b) => b.status == 'WAITING_APPROVAL')
          .toList();
      return batches;
    } catch (e) {
      rethrow;
    }
  }

  @override
  Future<void> approveDeliveryBatch(String batchId) async {
    try {
      await dioClient.dio.post('/delivery/batch/$batchId/supervisor-approve');
    } catch (e) {
      rethrow;
    }
  }

  @override
  Future<void> cancelDeliveryBatch(String batchId) async {
    try {
      await dioClient.dio.delete('/delivery/batches/$batchId');
    } catch (e) {
      rethrow;
    }
  }

  @override
  Future<List<SalesKpiModel>> getSalesKPIs(
      int month, int year, String branchId) async {
    try {
      final response = await dioClient.dio
          .get('/admin/sales/reports/performance', queryParameters: {
        'month': month,
        'year': year,
        if (branchId.isNotEmpty &&
            branchId != 'ALL_BRANCHES' &&
            branchId != '0' &&
            branchId != 'all')
          'branch_id': branchId,
      });
      final responseData = response.data;
      print('DEBUG: responseData = $responseData');
      final List data = (responseData is List)
          ? responseData
          : (responseData is Map ? (responseData['data'] ?? []) : []);
      print('DEBUG: data list length = ${data.length}');
      final allKpis = data.map((json) => SalesKpiModel.fromJson(json)).toList();
      print('DEBUG: parsed allKpis length = ${allKpis.length}');

      // Filter out MT sales (Modern Trade) dan hanya tampilkan sales reguler cabang
      return allKpis.where((kpi) {
        final title = kpi.jobTitle.toLowerCase();
        final isMT = title == 'mt' ||
            title.contains('modern trade') ||
            title.contains('sales mt');
        return !isMT;
      }).toList();
    } catch (e) {
      print('DEBUG: Exception in getSalesKPIs: $e');
      rethrow;
    }
  }
}
