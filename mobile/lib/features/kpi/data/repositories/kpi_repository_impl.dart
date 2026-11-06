import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/kpi/data/models/kpi_model.dart';
import 'package:hris_app/features/kpi/domain/repositories/kpi_repository.dart';

class KPIRepositoryImpl implements KPIRepository {
  final ApiClient apiClient;

  KPIRepositoryImpl({required this.apiClient});

  @override
  Future<List<KPIModel>> getKPIHistory() async {
    try {
      final response = await apiClient.client.get('performance/my-history');
      if (response.statusCode == 200) {
        final List data = response.data['data'] ?? [];
        return data.map((json) => KPIModel.fromJson(json)).toList();
      }
      throw Exception('Gagal mengambil riwayat KPI');
    } catch (e) {
      throw Exception('Gagal mengambil riwayat KPI: $e');
    }
  }

  @override
  Future<KPIModel> getKPIByPeriod(String yearMonth) async {
    try {
      // Parse yearMonth (format: YYYY-MM)
      final parts = yearMonth.split('-');
      final year = parts[0];
      final month = parts[1];

      final response = await apiClient.client.get(
        'performance/my-performance',
        queryParameters: {
          'month': month,
          'year': year,
        },
      );

      if (response.statusCode == 200) {
        return KPIModel.fromJson(response.data['data']);
      }
      throw Exception('Gagal mengambil data KPI');
    } catch (e) {
      throw Exception('Gagal mengambil data KPI: $e');
    }
  }
}
