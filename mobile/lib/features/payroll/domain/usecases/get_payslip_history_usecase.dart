import 'package:dartz/dartz.dart';
import '../../../../core/error/failures.dart';
import '../entities/payslip.dart';
import '../repositories/payroll_repository.dart';

class GetMyPayslipHistoryUseCase {
  final PayrollRepository repository;

  GetMyPayslipHistoryUseCase(this.repository);

  Future<Either<Failure, List<Payslip>>> call() {
    return repository.getMyPayslipHistory();
  }
}
