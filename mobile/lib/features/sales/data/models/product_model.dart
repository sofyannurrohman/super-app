import 'package:hris_app/core/utils/constants.dart';

class ProductModel {
  final String id;
  final String name;
  final String sku;
  final String unit;
  final String category;
  final String brand;
  final double sellingPrice;
  final String description;
  final String imageUrl;
  final int warehouseStock;

  ProductModel({
    required this.id,
    required this.name,
    required this.sku,
    required this.unit,
    required this.category,
    required this.brand,
    required this.sellingPrice,
    required this.description,
    required this.imageUrl,
    this.warehouseStock = 0,
  });

  factory ProductModel.fromJson(Map<String, dynamic> json) {
    String rawImageUrl = json['image_url'] ?? '';
    String fullImageUrl = '';

    if (rawImageUrl.isNotEmpty) {
      if (rawImageUrl.startsWith('http')) {
        fullImageUrl = rawImageUrl;
      } else {
        // Assume relative path from backend
        fullImageUrl =
            '${AppConstants.baseUrl.replaceAll('/api/v1/', '')}$rawImageUrl';
      }
    }

    return ProductModel(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      sku: json['sku'] ?? '',
      unit: json['unit'] ?? '',
      category: json['category'] ?? '',
      brand: json['brand'] ?? '',
      sellingPrice: (json['selling_price'] is int)
          ? (json['selling_price'] as int).toDouble()
          : (json['selling_price'] as num?)?.toDouble() ?? 0.0,
      description: json['description'] ?? '',
      imageUrl: fullImageUrl,
      warehouseStock: json['warehouse_stock'] ?? 0,
    );
  }
}
