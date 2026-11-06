import 'package:equatable/equatable.dart';

class Payslip extends Equatable {
  final String id;
  final String period;
  final String paymentDate;
  final String jobTitle;
  final double basicSalary;
  final double totalAllowance;
  final double totalDeduction;
  final double takeHomePay;
  final List<PayslipItem> earnings;
  final List<PayslipItem> deductions;

  const Payslip({
    required this.id,
    required this.period,
    required this.paymentDate,
    required this.jobTitle,
    required this.basicSalary,
    required this.totalAllowance,
    required this.totalDeduction,
    required this.takeHomePay,
    required this.earnings,
    required this.deductions,
  });

  @override
  List<Object?> get props => [
        id,
        period,
        paymentDate,
        jobTitle,
        basicSalary,
        totalAllowance,
        totalDeduction,
        takeHomePay,
        earnings,
        deductions,
      ];
}

class PayslipItem extends Equatable {
  final String name;
  final double amount;

  const PayslipItem({
    required this.name,
    required this.amount,
  });

  @override
  List<Object?> get props => [name, amount];
}
