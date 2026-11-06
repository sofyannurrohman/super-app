package usecase

import (
	"math"
	"time"

	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetFactoryLinkedBranchIDs returns the factory's own branch. Additional warehouse
// links require explicit configuration rather than matching branch names.
func GetFactoryLinkedBranchIDs(tx *gorm.DB, factoryID uuid.UUID) ([]uuid.UUID, error) {
	var factory domain.Factory
	if err := tx.First(&factory, "id = ?", factoryID).Error; err != nil {
		return nil, err
	}

	branchIDMap := make(map[uuid.UUID]bool)
	if factory.BranchID != uuid.Nil {
		branchIDMap[factory.BranchID] = true
	}

	result := make([]uuid.UUID, 0, len(branchIDMap))
	for id := range branchIDMap {
		result = append(result, id)
	}

	return result, nil
}

// syncFactoryStockDeltaToLinkedWarehouses applies a quantity change (diff) to all linked branch warehouse stocks 1:1
func syncFactoryStockDeltaToLinkedWarehouses(
	tx *gorm.DB,
	factoryID uuid.UUID,
	productID uuid.UUID,
	batchNo string,
	expiryDate *time.Time,
	diff int,
	logSource string,
	processedBy string,
) error {
	if diff == 0 {
		return nil
	}
	if batchNo == "" {
		batchNo = "DEFAULT"
	}
	if processedBy == "" {
		processedBy = "SYSTEM"
	}

	branchIDs, err := GetFactoryLinkedBranchIDs(tx, factoryID)
	if err != nil {
		return err
	}

	wRepo := repository.NewWarehouseRepository(tx)
	logType := "IN"
	absDiff := int(math.Abs(float64(diff)))
	if diff < 0 {
		logType = "OUT"
	}

	for _, branchID := range branchIDs {
		whStock, err := wRepo.GetStock(branchID, productID, "GOOD", batchNo)
		if err != nil && err != gorm.ErrRecordNotFound {
			continue
		}
		if whStock == nil || whStock.ID == uuid.Nil {
			whStock = &domain.WarehouseStock{
				ID:        uuid.New(),
				BranchID:  branchID,
				ProductID: productID,
				BatchNo:   batchNo,
				StockType: "GOOD",
				Quantity:  0,
			}
		}

		whStock.Quantity += diff
		if whStock.Quantity < 0 {
			whStock.Quantity = 0
		}
		if expiryDate != nil {
			whStock.ExpiryDate = expiryDate
		}

		if err := wRepo.UpdateStock(whStock); err != nil {
			return err
		}

		_ = wRepo.CreateLog(&domain.WarehouseLog{
			BranchID:    branchID,
			ProductID:   productID,
			Type:        logType,
			Source:      logSource,
			BatchNo:     batchNo,
			Quantity:    absDiff,
			ProcessedBy: processedBy,
		})
	}

	return nil
}

// syncFactoryStockAbsoluteToLinkedWarehouses sets the exact quantity to all linked branch warehouse stocks 1:1
func syncFactoryStockAbsoluteToLinkedWarehouses(
	tx *gorm.DB,
	factoryID uuid.UUID,
	productID uuid.UUID,
	batchNo string,
	expiryDate *time.Time,
	absoluteQty int,
	logSource string,
	processedBy string,
) error {
	if batchNo == "" {
		batchNo = "DEFAULT"
	}
	if processedBy == "" {
		processedBy = "SYSTEM"
	}

	branchIDs, err := GetFactoryLinkedBranchIDs(tx, factoryID)
	if err != nil {
		return err
	}

	wRepo := repository.NewWarehouseRepository(tx)

	for _, branchID := range branchIDs {
		whStock, err := wRepo.GetStock(branchID, productID, "GOOD", batchNo)
		if err != nil && err != gorm.ErrRecordNotFound {
			continue
		}
		if whStock == nil || whStock.ID == uuid.Nil {
			whStock = &domain.WarehouseStock{
				ID:        uuid.New(),
				BranchID:  branchID,
				ProductID: productID,
				BatchNo:   batchNo,
				StockType: "GOOD",
				Quantity:  0,
			}
		}

		diff := absoluteQty - whStock.Quantity
		whStock.Quantity = absoluteQty
		if expiryDate != nil {
			whStock.ExpiryDate = expiryDate
		}

		if err := wRepo.UpdateStock(whStock); err != nil {
			return err
		}

		if diff != 0 {
			logType := "IN"
			absDiff := int(math.Abs(float64(diff)))
			if diff < 0 {
				logType = "OUT"
			}
			_ = wRepo.CreateLog(&domain.WarehouseLog{
				BranchID:    branchID,
				ProductID:   productID,
				Type:        logType,
				Source:      logSource,
				BatchNo:     batchNo,
				Quantity:    absDiff,
				ProcessedBy: processedBy,
			})
		}
	}

	return nil
}
