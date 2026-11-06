package usecase

import (
	"time"

	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type DiscountEngine struct {
	DB *gorm.DB
}

func NewDiscountEngine(db *gorm.DB) *DiscountEngine {
	return &DiscountEngine{DB: db}
}

// EvaluateCart evaluates a given cart/sales order items and calculates applicable discounts based on rules.
func (e *DiscountEngine) EvaluateCart(paymentMethod string, items []domain.SalesOrderItem) (float64, error) {
	var rules []domain.DiscountRule
	// Load active rules
	now := time.Now()
	err := e.DB.Preload("Tiers").
		Preload("ExcludedProducts").
		Preload("ValidUnits").
		Where("is_active = ?", true).
		Where("payment_method_required = ? OR payment_method_required = ''", paymentMethod).
		Where("start_date IS NULL OR start_date <= ?", now).
		Where("end_date IS NULL OR end_date >= ?", now).
		Find(&rules).Error

	if err != nil {
		return 0, err
	}

	var totalDiscountAmount float64 = 0

	for _, rule := range rules {
		// Evaluate rule for EACH item independently
		for _, item := range items {
			// Check Excluded Products
			isExcluded := false
			for _, exp := range rule.ExcludedProducts {
				if item.ProductID == exp.ProductID {
					isExcluded = true
					break
				}
			}
			if isExcluded {
				continue
			}

			// Check Valid Unit
			isValidUnit := false
			if len(rule.ValidUnits) == 0 {
				isValidUnit = true // If no specific units, all units valid
			} else {
				for _, vu := range rule.ValidUnits {
					if item.Unit == vu.UnitName {
						isValidUnit = true
						break
					}
				}
			}

			if isValidUnit {
				// The quantity to evaluate is only THIS specific item's quantity
				itemQty := item.OrderedQuantity

				// Find highest applicable tier for THIS item
				var applicableTier *domain.DiscountTier
				for j := range rule.Tiers {
					if itemQty >= rule.Tiers[j].MinQty {
						if applicableTier == nil || rule.Tiers[j].MinQty > applicableTier.MinQty {
							applicableTier = &rule.Tiers[j]
						}
					}
				}

				// Apply discount to this item if a tier is found
				if applicableTier != nil {
					// Assuming price is unit price
					itemBasePrice := item.Price * float64(item.OrderedQuantity)
					discountAmount := (itemBasePrice * applicableTier.DiscountPercent) / 100
					totalDiscountAmount += discountAmount

					// Optional: If you want to update the item's DiscountPercent directly in the struct:
					// items[i].DiscountPercent = applicableTier.DiscountPercent
					// Tapi perlu return items juga atau di-update reference-nya oleh caller.
				}
			}
		}
	}

	return totalDiscountAmount, nil
}
