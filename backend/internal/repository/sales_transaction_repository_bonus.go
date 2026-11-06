package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
)

func (r *salesTransactionRepository) injectBonusProducts(transactions []*domain.SalesTransaction) {
	missingProductIDs := make([]uuid.UUID, 0)
	for _, t := range transactions {
		if t == nil {
			continue
		}
		for _, item := range t.Items {
			if item.Product == nil {
				missingProductIDs = append(missingProductIDs, item.ProductID)
			}
		}
	}

	if len(missingProductIDs) > 0 {
		var bonuses []domain.BonusInventory
		if err := r.db.Where("id IN ?", missingProductIDs).Find(&bonuses).Error; err == nil {
			bonusMap := make(map[uuid.UUID]domain.BonusInventory)
			for _, b := range bonuses {
				bonusMap[b.ID] = b
			}

			for i := range transactions {
				if transactions[i] == nil {
					continue
				}
				for j := range transactions[i].Items {
					if transactions[i].Items[j].Product == nil {
						if b, ok := bonusMap[transactions[i].Items[j].ProductID]; ok {
							transactions[i].Items[j].Product = &domain.Product{
								ID:   b.ID,
								Name: b.Name + " (BONUS)",
								SKU:  "BONUS",
								Unit: b.Unit,
							}
						}
					}
				}
			}
		}
	}
}
