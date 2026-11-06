class DeliveryBatchModel {
  final String id;
  final String deliveryOrderNo;
  final String status;
  final String driverName;
  final String helperName;
  final String vehicleInfo;
  final DateTime? createdAt;
  final List<DeliveryItemModel> items;

  DeliveryBatchModel({
    required this.id,
    required this.deliveryOrderNo,
    required this.status,
    required this.driverName,
    required this.helperName,
    required this.vehicleInfo,
    this.createdAt,
    this.items = const [],
  });

  factory DeliveryBatchModel.fromJson(Map<String, dynamic> json) {
    var itemsList = json['items'] as List? ?? [];
    List<DeliveryItemModel> mappedItems =
        itemsList.map((i) => DeliveryItemModel.fromJson(i)).toList();

    return DeliveryBatchModel(
      id: json['id'] ?? '',
      deliveryOrderNo: json['delivery_order_no'] ?? '',
      status: json['status'] ?? '',
      driverName: (json['driver'] != null)
          ? '${json['driver']['first_name'] ?? ''} ${json['driver']['last_name'] ?? ''}'
              .trim()
          : '-',
      helperName: (json['helper'] != null)
          ? '${json['helper']['first_name'] ?? ''} ${json['helper']['last_name'] ?? ''}'
              .trim()
          : '-',
      vehicleInfo: (json['vehicle'] != null)
          ? '${json['vehicle']['license_plate'] ?? ''} (${json['vehicle']['type'] ?? ''})'
          : '-',
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'])
          : null,
      items: mappedItems,
    );
  }

  double get totalAmount {
    double total = 0;
    for (var item in items) {
      total += item.totalAmount;
    }
    return total;
  }
}

class DeliveryItemModel {
  final String id;
  final String storeName;
  final String receiptNo;
  final double totalAmount;

  DeliveryItemModel({
    required this.id,
    required this.storeName,
    required this.receiptNo,
    required this.totalAmount,
  });

  factory DeliveryItemModel.fromJson(Map<String, dynamic> json) {
    String store = 'Toko Lain';
    String receipt = '';
    double amount = 0;

    if (json['sales_order'] != null) {
      store = json['sales_order']['store']?['name'] ?? store;
      receipt = json['sales_order']['so_number'] ?? '';
      amount = (json['sales_order']['total_amount'] is num)
          ? (json['sales_order']['total_amount'] as num).toDouble()
          : 0.0;
    } else if (json['sales_transaction'] != null) {
      store = json['sales_transaction']['store']?['name'] ?? store;
      receipt = json['sales_transaction']['receipt_no'] ?? '';
      amount = (json['sales_transaction']['total_amount'] is num)
          ? (json['sales_transaction']['total_amount'] as num).toDouble()
          : 0.0;
    } else {
      receipt = (json['sales_order_id'] ?? '').toString();
    }

    return DeliveryItemModel(
      id: json['id'] ?? '',
      storeName: store,
      receiptNo: receipt,
      totalAmount: amount,
    );
  }
}
