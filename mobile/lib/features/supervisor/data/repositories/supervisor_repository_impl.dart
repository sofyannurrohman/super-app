import 'package:dartz/dartz.dart';
import 'package:dio/dio.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/features/supervisor/data/datasources/supervisor_remote_data_source.dart';
import 'package:hris_app/features/supervisor/data/models/delivery_batch_model.dart';
import 'package:hris_app/features/supervisor/data/models/sales_kpi_model.dart';
import 'package:hris_app/features/supervisor/data/models/sales_transfer_model.dart';
import 'package:hris_app/features/supervisor/domain/repositories/supervisor_repository.dart';

class SupervisorRepositoryImpl implements SupervisorRepository {
  final SupervisorRemoteDataSource remoteDataSource;

  SupervisorRepositoryImpl({required this.remoteDataSource});

  String _getErrorMessage(DioException e) {
    if (e.response?.data is Map<String, dynamic>) {
      return e.response?.data['error'] ??
          e.response?.data['message'] ??
          e.message ??
          'Server error';
    }
    return e.message ?? 'Server error';
  }

  @override
  Future<Either<Failure, List<SalesTransferModel>>> getPendingTransfers(
      String branchId) async {
    try {
      final transfers = await remoteDataSource.getPendingTransfers(branchId);
      return Right(transfers);
    } on DioException catch (e) {
      return Left(ServerFailure(_getErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> approveTransfer(
      String transferId, int approvedQuantity) async {
    try {
      await remoteDataSource.approveTransfer(transferId, approvedQuantity);
      return const Right(null);
    } on DioException catch (e) {
      return Left(ServerFailure(_getErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> rejectTransfer(String transferId) async {
    try {
      await remoteDataSource.rejectTransfer(transferId);
      return const Right(null);
    } on DioException catch (e) {
      return Left(ServerFailure(_getErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, List<DeliveryBatchModel>>> getPendingDeliveryBatches(
      String branchId) async {
    try {
      final batches =
          await remoteDataSource.getPendingDeliveryBatches(branchId);
      return Right(batches);
    } on DioException catch (e) {
      return Left(ServerFailure(_getErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> approveDeliveryBatch(String batchId) async {
    try {
      await remoteDataSource.approveDeliveryBatch(batchId);
      return const Right(null);
    } on DioException catch (e) {
      return Left(ServerFailure(_getErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> cancelDeliveryBatch(String batchId) async {
    try {
      await remoteDataSource.cancelDeliveryBatch(batchId);
      return const Right(null);
    } on DioException catch (e) {
      return Left(ServerFailure(_getErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, List<SalesKpiModel>>> getSalesKPIs(
      int month, int year, String branchId) async {
    try {
      final kpis = await remoteDataSource.getSalesKPIs(month, year, branchId);
      return Right(kpis);
    } on DioException catch (e) {
      return Left(ServerFailure(_getErrorMessage(e)));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }
}
