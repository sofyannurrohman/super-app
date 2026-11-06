import 'package:hris_app/features/supervisor/data/models/sales_transfer_model.dart';
import 'package:hris_app/features/supervisor/data/models/delivery_batch_model.dart';
import 'package:hris_app/features/supervisor/data/models/sales_kpi_model.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:dartz/dartz.dart';

abstract class SupervisorRepository {
  Future<Either<Failure, List<SalesTransferModel>>> getPendingTransfers(
      String branchId);
  Future<Either<Failure, void>> approveTransfer(
      String transferId, int approvedQuantity);
  Future<Either<Failure, void>> rejectTransfer(String transferId);

  Future<Either<Failure, List<DeliveryBatchModel>>> getPendingDeliveryBatches(
      String branchId);
  Future<Either<Failure, void>> approveDeliveryBatch(String batchId);
  Future<Either<Failure, void>> cancelDeliveryBatch(String batchId);

  Future<Either<Failure, List<SalesKpiModel>>> getSalesKPIs(
      int month, int year, String branchId);
}
