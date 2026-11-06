import 'package:dartz/dartz.dart';
import 'package:dio/dio.dart';
import '../../../../core/error/failures.dart';
import '../../../../core/network/api_client.dart';
import '../../domain/entities/payslip.dart';
import '../../domain/repositories/payroll_repository.dart';
import '../models/payslip_model.dart';

class PayrollRepositoryImpl implements PayrollRepository {
  final ApiClient apiClient;

  PayrollRepositoryImpl({required this.apiClient});

  @override
  Future<Either<Failure, List<Payslip>>> getMyPayslipHistory() async {
    try {
      final response = await apiClient.client.get('payroll/my-payslip');

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        final List<Payslip> payslips =
            data.map((e) => PayslipModel.fromJson(e)).toList();
        return Right(payslips);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to load payslips')
            : 'Failed to load payslips'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to load payslips'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }
}
