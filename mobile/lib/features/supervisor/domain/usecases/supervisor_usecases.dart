import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/features/supervisor/data/models/delivery_batch_model.dart';
import 'package:hris_app/features/supervisor/data/models/sales_kpi_model.dart';
import 'package:hris_app/features/supervisor/data/models/sales_transfer_model.dart';
import 'package:hris_app/features/supervisor/domain/repositories/supervisor_repository.dart';

class GetPendingTransfers {
  final SupervisorRepository repository;

  GetPendingTransfers(this.repository);

  Future<Either<Failure, List<SalesTransferModel>>> call(String branchId) {
    return repository.getPendingTransfers(branchId);
  }
}

class ApproveTransfer {
  final SupervisorRepository repository;

  ApproveTransfer(this.repository);

  Future<Either<Failure, void>> call(String transferId, int approvedQuantity) {
    return repository.approveTransfer(transferId, approvedQuantity);
  }
}

class RejectTransfer {
  final SupervisorRepository repository;

  RejectTransfer(this.repository);

  Future<Either<Failure, void>> call(String transferId) {
    return repository.rejectTransfer(transferId);
  }
}

class GetPendingDeliveryBatches {
  final SupervisorRepository repository;

  GetPendingDeliveryBatches(this.repository);

  Future<Either<Failure, List<DeliveryBatchModel>>> call(String branchId) {
    return repository.getPendingDeliveryBatches(branchId);
  }
}

class ApproveDeliveryBatch {
  final SupervisorRepository repository;

  ApproveDeliveryBatch(this.repository);

  Future<Either<Failure, void>> call(String batchId) {
    return repository.approveDeliveryBatch(batchId);
  }
}

class CancelDeliveryBatch {
  final SupervisorRepository repository;

  CancelDeliveryBatch(this.repository);

  Future<Either<Failure, void>> call(String batchId) {
    return repository.cancelDeliveryBatch(batchId);
  }
}

class GetSalesKPIs {
  final SupervisorRepository repository;

  GetSalesKPIs(this.repository);

  Future<Either<Failure, List<SalesKpiModel>>> call(
      int month, int year, String branchId) {
    return repository.getSalesKPIs(month, year, branchId);
  }
}
