import 'package:dartz/dartz.dart';
import '../../../../core/error/failures.dart';
import '../entities/payslip.dart';

abstract class PayrollRepository {
  Future<Either<Failure, List<Payslip>>> getMyPayslipHistory();
}
