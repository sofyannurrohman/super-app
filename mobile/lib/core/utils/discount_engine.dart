class DiscountTier {
  final int minQty;
  final double discountPercent;

  DiscountTier({required this.minQty, required this.discountPercent});

  factory DiscountTier.fromJson(Map<String, dynamic> json) {
    return DiscountTier(
      minQty: json['min_qty'] ?? 0,
      discountPercent: (json['discount_percent'] ?? 0).toDouble(),
    );
  }
}

class DiscountExcludedProduct {
  final String productId;

  DiscountExcludedProduct({required this.productId});

  factory DiscountExcludedProduct.fromJson(Map<String, dynamic> json) {
    return DiscountExcludedProduct(
      productId: json['product_id'] ?? '',
    );
  }
}

class DiscountValidUnit {
  final String unitName;

  DiscountValidUnit({required this.unitName});

  factory DiscountValidUnit.fromJson(Map<String, dynamic> json) {
    return DiscountValidUnit(
      unitName: json['unit_name'] ?? '',
    );
  }
}

class DiscountRule {
  final String id;
  final String name;
  final String paymentMethodRequired;
  final bool isActive;
  final List<DiscountTier> tiers;
  final List<DiscountExcludedProduct> excludedProducts;
  final List<DiscountValidUnit> validUnits;

  DiscountRule({
    required this.id,
    required this.name,
    required this.paymentMethodRequired,
    required this.isActive,
    required this.tiers,
    required this.excludedProducts,
    required this.validUnits,
  });

  factory DiscountRule.fromJson(Map<String, dynamic> json) {
    return DiscountRule(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      paymentMethodRequired: json['payment_method_required'] ?? '',
      isActive: json['is_active'] ?? false,
      tiers: (json['Tiers'] as List?)
              ?.map((e) => DiscountTier.fromJson(e))
              .toList() ??
          [],
      excludedProducts: (json['ExcludedProducts'] as List?)
              ?.map((e) => DiscountExcludedProduct.fromJson(e))
              .toList() ??
          [],
      validUnits: (json['ValidUnits'] as List?)
              ?.map((e) => DiscountValidUnit.fromJson(e))
              .toList() ??
          [],
    );
  }
}

class OrderItemInput {
  final String productId;
  final String unit;
  final int quantity;
  final double price;

  OrderItemInput({
    required this.productId,
    required this.unit,
    required this.quantity,
    required this.price,
  });
}

class DiscountEngine {
  /// Evaluates the discount for each item and returns a map of {productId: discountPercent}
  static Map<String, double> evaluateDiscounts({
    required List<DiscountRule> activeRules,
    required String currentPaymentMethod, // e.g. "CASH"
    required List<OrderItemInput> items,
  }) {
    Map<String, double> result = {};

    for (var rule in activeRules) {
      if (!rule.isActive) continue;

      // Check payment method match (if rule has one)
      if (rule.paymentMethodRequired.isNotEmpty &&
          rule.paymentMethodRequired != currentPaymentMethod) {
        continue;
      }

      for (var item in items) {
        // 1. Check Excluded Product
        bool isExcluded =
            rule.excludedProducts.any((exp) => exp.productId == item.productId);
        if (isExcluded) continue;

        // 2. Check Valid Unit
        bool isValidUnit = true;
        if (rule.validUnits.isNotEmpty) {
          isValidUnit = rule.validUnits.any((vu) => vu.unitName == item.unit);
        }

        if (isValidUnit) {
          // 3. Find highest applicable tier for this item's quantity
          DiscountTier? applicableTier;
          for (var tier in rule.tiers) {
            if (item.quantity >= tier.minQty) {
              if (applicableTier == null ||
                  tier.minQty > applicableTier.minQty) {
                applicableTier = tier;
              }
            }
          }

          if (applicableTier != null) {
            // Apply discount (only overriding if the new discount is higher)
            final currentDisc = result[item.productId] ?? 0.0;
            if (applicableTier.discountPercent > currentDisc) {
              result[item.productId] = applicableTier.discountPercent;
            }
          }
        }
      }
    }

    return result;
  }
}
