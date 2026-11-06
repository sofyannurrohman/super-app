import 'package:equatable/equatable.dart';
import '../../domain/entities/payslip.dart';

abstract class PayrollState extends Equatable {
  const PayrollState();

  @override
  List<Object?> get props => [];
}

class PayrollInitial extends PayrollState {}

class PayrollLoading extends PayrollState {}

class PayrollLoaded extends PayrollState {
  final List<Payslip> payslips;

  const PayrollLoaded(this.payslips);

  @override
  List<Object?> get props => [payslips];
}

class PayrollFailure extends PayrollState {
  final String message;

  const PayrollFailure(this.message);

  @override
  List<Object?> get props => [message];
}
