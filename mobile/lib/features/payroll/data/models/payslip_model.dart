import '../../domain/entities/payslip.dart';

class PayslipModel extends Payslip {
  const PayslipModel({
    required super.id,
    required super.period,
    required super.paymentDate,
    required super.jobTitle,
    required super.basicSalary,
    required super.totalAllowance,
    required super.totalDeduction,
    required super.takeHomePay,
    required super.earnings,
    required super.deductions,
  });

  factory PayslipModel.fromJson(Map<String, dynamic> json) {
    return PayslipModel(
      id: json['id'] ?? '',
      period: json['period'] ?? '',
      paymentDate: json['payment_date'] ?? '',
      jobTitle: json['job_title'] ?? '',
      basicSalary: (json['basic_salary'] as num?)?.toDouble() ?? 0.0,
      totalAllowance: (json['total_allowance'] as num?)?.toDouble() ?? 0.0,
      totalDeduction: (json['total_deduction'] as num?)?.toDouble() ?? 0.0,
      takeHomePay: (json['take_home_pay'] as num?)?.toDouble() ?? 0.0,
      earnings: (json['earnings'] as List<dynamic>?)
              ?.map((e) => PayslipItemModel.fromJson(e))
              .toList() ??
          [],
      deductions: (json['deductions'] as List<dynamic>?)
              ?.map((e) => PayslipItemModel.fromJson(e))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'period': period,
      'payment_date': paymentDate,
      'job_title': jobTitle,
      'basic_salary': basicSalary,
      'total_allowance': totalAllowance,
      'total_deduction': totalDeduction,
      'take_home_pay': takeHomePay,
      'earnings':
          earnings.map((e) => (e as PayslipItemModel).toJson()).toList(),
      'deductions':
          deductions.map((e) => (e as PayslipItemModel).toJson()).toList(),
    };
  }
}

class PayslipItemModel extends PayslipItem {
  const PayslipItemModel({
    required super.name,
    required super.amount,
  });

  factory PayslipItemModel.fromJson(Map<String, dynamic> json) {
    return PayslipItemModel(
      name: json['name'] ?? '',
      amount: (json['amount'] as num?)?.toDouble() ?? 0.0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'amount': amount,
    };
  }
}
