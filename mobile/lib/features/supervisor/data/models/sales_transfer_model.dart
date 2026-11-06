import 'package:hris_app/features/sales/data/models/product_model.dart';

class SalesTransferModel {
  final String id;
  final String employeeId;
  final String employeeName;
  final String referenceNo;
  final String productId;
  final int quantity;
  final String status;
  final ProductModel? product;
  final DateTime? transferDate;

  SalesTransferModel({
    required this.id,
    required this.employeeId,
    required this.employeeName,
    required this.referenceNo,
    required this.productId,
    required this.quantity,
    required this.status,
    this.product,
    this.transferDate,
  });

  factory SalesTransferModel.fromJson(Map<String, dynamic> json) {
    return SalesTransferModel(
      id: json['id'] ?? '',
      employeeId: json['employee_id'] ?? json['employee']?['id'] ?? '',
      employeeName: (json['employee'] != null)
          ? '${json['employee']['first_name'] ?? ''} ${json['employee']['last_name'] ?? ''}'
              .trim()
          : '',
      referenceNo: json['reference_no'] ?? '',
      productId: json['product_id'] ?? '',
      quantity: json['quantity'] ?? 0,
      status: json['status'] ?? '',
      product: json['product'] != null
          ? ProductModel.fromJson(json['product'])
          : null,
      transferDate: json['transfer_date'] != null
          ? DateTime.tryParse(json['transfer_date'])
          : null,
    );
  }
}
