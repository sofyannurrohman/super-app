import 'package:hris_app/core/network/api_client.dart';

class FinanceApiService {
  final ApiClient apiClient;

  FinanceApiService({required this.apiClient});

  Future<List<dynamic>> getPendingPayments() async {
    try {
      final response = await apiClient.client.get('finance/payments/pending');
      return (response.data['data'] as List<dynamic>?) ?? [];
    } catch (e) {
      throw Exception('Failed to load pending payments: $e');
    }
  }

  Future<void> approvePayment(String id) async {
    try {
      await apiClient.client.post('finance/payments/$id/approve');
    } catch (e) {
      throw Exception('Failed to approve payment: $e');
    }
  }
}
