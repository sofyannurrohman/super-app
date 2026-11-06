package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type WarehouseDeliveryOrderUsecase interface {
	CreateOrGetDO(batchID uuid.UUID) (*domain.WarehouseDeliveryOrder, error)
	CreateOrGetDelegationDO(doNo string) (*domain.WarehouseDeliveryOrder, error)
	CreateMotorissDO(salesTransferIDs []uuid.UUID) (*domain.WarehouseDeliveryOrder, error)
	UpdateDO(req UpdateWarehouseDORequest, userID uuid.UUID, userName string, ip string) (*domain.WarehouseDeliveryOrder, error)
	DispatchDO(id uuid.UUID, userID uuid.UUID, userName string, ip string) error
	FinalizeDO(id uuid.UUID, userID uuid.UUID, userName string, ip string) error
	GetByBatchID(batchID uuid.UUID) (*domain.WarehouseDeliveryOrder, error)
	GetMotorisDOs(branchID *uuid.UUID) ([]domain.WarehouseDeliveryOrder, error)
	GetByID(id uuid.UUID) (*domain.WarehouseDeliveryOrder, error)
}

type warehouseDeliveryOrderUsecase struct {
	repo         repository.WarehouseDeliveryOrderRepository
	deliveryRepo repository.DeliveryRepository
	auditUsecase AuditLogUsecase
	db           *gorm.DB
}

func NewWarehouseDeliveryOrderUsecase(
	repo repository.WarehouseDeliveryOrderRepository,
	deliveryRepo repository.DeliveryRepository,
	auditUsecase AuditLogUsecase,
	db *gorm.DB,
) WarehouseDeliveryOrderUsecase {
	return &warehouseDeliveryOrderUsecase{repo, deliveryRepo, auditUsecase, db}
}

type UpdateWarehouseDORequest struct {
	ID     uuid.UUID             `json:"id" binding:"required"`
	Notes  string                `json:"notes"`
	Status string                `json:"status"`
	Items  []UpdateDOItemRequest `json:"items" binding:"required"`
}

type UpdateDOItemRequest struct {
	ProductID               *uuid.UUID `json:"product_id"`
	BonusInventoryID        *uuid.UUID `json:"bonus_inventory_id"`
	CartonsLoaded           int        `json:"cartons_loaded"`
	PcsLoaded               int        `json:"pcs_loaded"`
	CustomPcsPerUnit        int        `json:"custom_pcs_per_unit"`
	PcsReturnedToWarehouse  int        `json:"pcs_returned_to_warehouse"`
	PcsSold                 int        `json:"pcs_sold"`
	PcsSample               int        `json:"pcs_sample"`
	PcsReturnedFromCustomer int        `json:"pcs_returned_from_customer"`
}

// CreateOrGetDO — untuk SJ biasa (REGULAR DO, terhubung ke DeliveryBatch)
func (u *warehouseDeliveryOrderUsecase) CreateOrGetDO(batchID uuid.UUID) (*domain.WarehouseDeliveryOrder, error) {
	// Fetch DeliveryBatch to get products and bonus items list
	batch, err := u.deliveryRepo.GetBatchByID(batchID)
	if err != nil {
		return nil, fmt.Errorf("surat jalan tidak ditemukan: %w", err)
	}

	// Aggregate products and quantities from DeliveryBatch items
	productQtyMap := make(map[uuid.UUID]int)
	bonusQtyMap := make(map[uuid.UUID]int)
	for _, item := range batch.Items {
		if item.SalesOrder != nil {
			for _, soItem := range item.SalesOrder.Items {
				productQtyMap[soItem.ProductID] += soItem.OrderedQuantity
			}
			for _, bItem := range item.SalesOrder.BonusItems {
				bonusQtyMap[bItem.BonusInventoryID] += bItem.Quantity
			}
		}
		if item.SalesSubsequentItemID != nil {
			var subItem domain.SalesSubsequentItem
			if errDb := u.db.First(&subItem, "id = ?", *item.SalesSubsequentItemID).Error; errDb == nil {
				productQtyMap[subItem.ProductID] += subItem.Quantity
			}
		}
		if item.SalesTransferID != nil {
			var trf domain.SalesTransfer
			if errDb := u.db.First(&trf, "id = ?", *item.SalesTransferID).Error; errDb == nil {
				if trf.ProductID != nil {
					productQtyMap[*trf.ProductID] += trf.Quantity
				} else if trf.BonusInventoryID != nil {
					bonusQtyMap[*trf.BonusInventoryID] += trf.Quantity
				}
			}
		}
	}

	for _, spare := range batch.Spares {
		productQtyMap[spare.ProductID] += spare.Quantity
	}

	// Check if already exists
	existing, err := u.repo.FindByBatchID(batchID)
	if err == nil && existing != nil {
		// If existing DO is in DRAFT, check if any items/bonuses are missing and append them
		if existing.Status == "DRAFT" {
			existingItemMap := make(map[string]bool)
			for _, item := range existing.Items {
				if item.ProductID != nil {
					existingItemMap[item.ProductID.String()] = true
				} else if item.BonusInventoryID != nil {
					existingItemMap[item.BonusInventoryID.String()] = true
				}
			}

			hasNewItems := false
			for prodID, totalPcs := range productQtyMap {
				if !existingItemMap[prodID.String()] {
					var prod domain.Product
					if errDb := u.db.First(&prod, "id = ?", prodID).Error; errDb == nil {
						pcsPerUnit := prod.PcsPerUnit
						if pcsPerUnit <= 0 {
							pcsPerUnit = 1
						}
						prodIDCopy := prodID
						existing.Items = append(existing.Items, domain.WarehouseDeliveryOrderItem{
							WarehouseDeliveryOrderID: existing.ID,
							ProductID:                &prodIDCopy,
							CartonsLoaded:            totalPcs / pcsPerUnit,
							PcsLoaded:                totalPcs % pcsPerUnit,
							CustomPcsPerUnit:         pcsPerUnit,
						})
						hasNewItems = true
					}
				}
			}

			for bonusID, totalPcs := range bonusQtyMap {
				if !existingItemMap[bonusID.String()] {
					bonusIDCopy := bonusID
					existing.Items = append(existing.Items, domain.WarehouseDeliveryOrderItem{
						WarehouseDeliveryOrderID: existing.ID,
						BonusInventoryID:         &bonusIDCopy,
						CartonsLoaded:            0,
						PcsLoaded:                totalPcs,
						CustomPcsPerUnit:         1,
					})
					hasNewItems = true
				}
			}

			if hasNewItems {
				_ = u.repo.Update(existing)
				return u.repo.FindByID(existing.ID)
			}
		}
		return existing, nil
	}

	// Generate DO Number
	now := time.Now()
	count, _ := u.repo.CountByDate(now)
	doNumber := repository.GenerateWarehouseDONumber(now, count)

	wdoID := uuid.New()
	wdo := &domain.WarehouseDeliveryOrder{
		ID:              wdoID,
		DeliveryBatchID: &batchID,
		DOType:          domain.DOTypeRegular,
		DONumber:        doNumber,
	}

	// Build DO Items with auto-calculated Carton/Pcs suggestions
	for prodID, totalPcs := range productQtyMap {
		var prod domain.Product
		if errDb := u.db.First(&prod, "id = ?", prodID).Error; errDb == nil {
			pcsPerUnit := prod.PcsPerUnit
			if pcsPerUnit <= 0 {
				pcsPerUnit = 1
			}

			cartons := totalPcs / pcsPerUnit
			pcs := totalPcs % pcsPerUnit

			prodIDCopy := prodID
			wdo.Items = append(wdo.Items, domain.WarehouseDeliveryOrderItem{
				WarehouseDeliveryOrderID: wdoID,
				ProductID:                &prodIDCopy,
				CartonsLoaded:            cartons,
				PcsLoaded:                pcs,
				CustomPcsPerUnit:         pcsPerUnit,
			})
		}
	}

	for bonusID, totalPcs := range bonusQtyMap {
		var inv domain.BonusInventory
		if errDb := u.db.First(&inv, "id = ?", bonusID).Error; errDb == nil {
			bonusIDCopy := bonusID
			wdo.Items = append(wdo.Items, domain.WarehouseDeliveryOrderItem{
				WarehouseDeliveryOrderID: wdoID,
				BonusInventoryID:         &bonusIDCopy,
				CartonsLoaded:            0,
				PcsLoaded:                totalPcs,
				CustomPcsPerUnit:         1,
			})
		}
	}

	if err := u.repo.Create(wdo); err != nil {
		return nil, fmt.Errorf("gagal membuat DO gudang: %w", err)
	}

	return u.repo.FindByID(wdo.ID)
}

// CreateOrGetDelegationDO — untuk DO Delegasi antar cabang
func (u *warehouseDeliveryOrderUsecase) CreateOrGetDelegationDO(doNo string) (*domain.WarehouseDeliveryOrder, error) {
	// 1. Check if WDO already exists by DONumber
	var existing domain.WarehouseDeliveryOrder
	err := u.db.Preload("Items.Product").First(&existing, "do_number = ?", doNo).Error
	if err == nil {
		return &existing, nil
	}

	// 2. Fetch all ProductTransfer with DeliveryOrderNo == doNo
	var transfers []domain.ProductTransfer
	err = u.db.Where("delivery_order_no = ?", doNo).Find(&transfers).Error
	if err != nil || len(transfers) == 0 {
		return nil, fmt.Errorf("delegasi dengan nomor SJ %s tidak ditemukan", doNo)
	}

	// 3. Create new WDO
	wdoID := uuid.New()
	wdo := &domain.WarehouseDeliveryOrder{
		ID:       wdoID,
		DOType:   domain.DOTypeDelegation,
		DONumber: doNo,
	}

	// 4. Aggregate products and quantities
	productQtyMap := make(map[uuid.UUID]int)
	for _, trf := range transfers {
		qty := trf.Quantity
		if trf.PcsPerUnit > 0 {
			qty *= trf.PcsPerUnit
		}
		productQtyMap[trf.ProductID] += qty
	}

	// 5. Build DO Items
	for prodID, totalPcs := range productQtyMap {
		var prod domain.Product
		if errDb := u.db.First(&prod, "id = ?", prodID).Error; errDb == nil {
			pcsPerUnit := prod.PcsPerUnit
			if pcsPerUnit <= 0 {
				pcsPerUnit = 1
			}

			cartons := totalPcs / pcsPerUnit
			pcs := totalPcs % pcsPerUnit

			wdo.Items = append(wdo.Items, domain.WarehouseDeliveryOrderItem{
				WarehouseDeliveryOrderID: wdoID,
				ProductID:                &prodID,
				CartonsLoaded:            cartons,
				PcsLoaded:                pcs,
				CustomPcsPerUnit:         pcsPerUnit,
			})
		}
	}

	if err := u.repo.Create(wdo); err != nil {
		return nil, fmt.Errorf("gagal membuat DO gudang untuk delegasi: %w", err)
	}

	return u.repo.FindByID(wdo.ID)
}

// CreateMotorissDO — membuat DO Motoris standalone (tanpa DeliveryBatch/SJ)
// dipanggil ketika user klik "Buat DO Motoris" di menu Perpindahan Barang
func (u *warehouseDeliveryOrderUsecase) CreateMotorissDO(salesTransferIDs []uuid.UUID) (*domain.WarehouseDeliveryOrder, error) {
	var result *domain.WarehouseDeliveryOrder

	err := u.db.Transaction(func(tx *gorm.DB) error {
		if len(salesTransferIDs) == 0 {
			return fmt.Errorf("tidak ada transfer yang dipilih")
		}

		// Ambil data transfer pertama untuk info salesman
		var empID uuid.UUID
		for _, trfID := range salesTransferIDs {
			var employeeID uuid.UUID

			var trf domain.SalesTransfer
			if err := tx.First(&trf, "id = ?", trfID).Error; err == nil {
				employeeID = trf.EmployeeID
				if trf.Status == domain.SalesTransferCompleted || trf.Status == domain.SalesTransferCancelled {
					return fmt.Errorf("tidak dapat membuat DO untuk transfer yang sudah selesai atau dibatalkan")
				}
				var existCount int64
				if errExist := tx.Model(&domain.WarehouseDeliveryOrderTransfer{}).Where("sales_transfer_id = ?", trfID).Count(&existCount).Error; errExist != nil {
					return fmt.Errorf("gagal mengecek DO Motoris eksisting: %w", errExist)
				}
				if existCount > 0 {
					return fmt.Errorf("DO Motoris sudah pernah dibuat untuk salah satu transfer yang dipilih")
				}
			} else {
				var bonusReq domain.SalesBonusRequest
				if err := tx.First(&bonusReq, "id = ?", trfID).Error; err == nil {
					employeeID = bonusReq.EmployeeID
					if bonusReq.Status != domain.SalesBonusRequestPending {
						return fmt.Errorf("tidak dapat membuat DO untuk bonus request yang sudah tidak pending")
					}
					var existCount int64
					if errExist := tx.Model(&domain.WarehouseDeliveryOrderTransfer{}).Where("sales_bonus_request_id = ?", trfID).Count(&existCount).Error; errExist != nil {
						return fmt.Errorf("gagal mengecek DO Motoris eksisting: %w", errExist)
					}
					if existCount > 0 {
						return fmt.Errorf("DO Motoris sudah pernah dibuat untuk salah satu bonus request yang dipilih")
					}
				} else {
					return fmt.Errorf("transfer atau bonus request tidak ditemukan")
				}
			}

			if empID == uuid.Nil {
				empID = employeeID
			} else if empID != employeeID {
				return fmt.Errorf("semua transfer/bonus harus milik salesman yang sama")
			}
		}

		// Cek apakah ada DRAFT atau DISPATCHED DO Motoris untuk salesman ini hari ini
		now := time.Now()
		var existingWDO domain.WarehouseDeliveryOrder
		errDraft := tx.Where("employee_id = ? AND status IN ('DRAFT', 'DISPATCHED') AND do_type = ? AND DATE(created_at AT TIME ZONE 'Asia/Jakarta') = DATE(? AT TIME ZONE 'Asia/Jakarta')",
			empID, domain.DOTypeMotoris, now).Order("created_at DESC").First(&existingWDO).Error

		var wdoID uuid.UUID
		var wdo *domain.WarehouseDeliveryOrder
		isAppending := false

		if errDraft == nil {
			wdoID = existingWDO.ID
			wdo = &existingWDO
			isAppending = true
		} else if !errors.Is(errDraft, gorm.ErrRecordNotFound) {
			return errDraft
		} else {
			count, errCount := repository.NewWarehouseDeliveryOrderRepository(tx).CountByDate(now)
			if errCount != nil {
				return fmt.Errorf("gagal menghitung DO hari ini: %w", errCount)
			}
			doNumber := repository.GenerateWarehouseDONumber(now, count)

			wdoID = uuid.New()
			wdo = &domain.WarehouseDeliveryOrder{
				ID:         wdoID,
				DOType:     domain.DOTypeMotoris,
				EmployeeID: &empID,
				DONumber:   doNumber,
				Status:     "DRAFT",
			}
			if err := tx.Create(wdo).Error; err != nil {
				return fmt.Errorf("gagal membuat DO Motoris: %w", err)
			}
		}

		// Aggregate produk dari semua transfer yang dipilih
		loadProductQtyMap := make(map[uuid.UUID]int)
		loadBonusQtyMap := make(map[uuid.UUID]int)
		returnProductQtyMap := make(map[uuid.UUID]int)
		returnBonusQtyMap := make(map[uuid.UUID]int)

		for _, trfID := range salesTransferIDs {
			var trf domain.SalesTransfer
			if err := tx.First(&trf, "id = ?", trfID).Error; err == nil {
				if trf.Type == domain.SalesTransferIn || trf.Type == "RETURN" {
					if trf.ProductID != nil {
						returnProductQtyMap[*trf.ProductID] += trf.Quantity
					} else if trf.BonusInventoryID != nil {
						returnBonusQtyMap[*trf.BonusInventoryID] += trf.Quantity
					}
				} else {
					if trf.ProductID != nil {
						loadProductQtyMap[*trf.ProductID] += trf.Quantity
					} else if trf.BonusInventoryID != nil {
						loadBonusQtyMap[*trf.BonusInventoryID] += trf.Quantity
					}
				}
			} else {
				var bonusReq domain.SalesBonusRequest
				if err := tx.First(&bonusReq, "id = ?", trfID).Error; err == nil {
					loadBonusQtyMap[bonusReq.BonusInventoryID] += bonusReq.Quantity
				}
			}
		}

		// Helper function to build or update DO item
		updateOrBuildDOItem := func(wdoID uuid.UUID, prodID *uuid.UUID, bonusID *uuid.UUID, loadQty int, returnQty int, isApp bool) error {
			if prodID != nil {
				var prod domain.Product
				if errDb := tx.First(&prod, "id = ?", *prodID).Error; errDb == nil {
					pcsPerUnit := prod.PcsPerUnit
					if pcsPerUnit <= 0 {
						pcsPerUnit = 1
					}

					if isApp {
						var existItem domain.WarehouseDeliveryOrderItem
						errItem := tx.Where("warehouse_delivery_order_id = ? AND product_id = ?", wdoID, *prodID).First(&existItem).Error
						if errItem == nil {
							if loadQty > 0 {
								totalPcsCombined := (existItem.CartonsLoaded * existItem.CustomPcsPerUnit) + existItem.PcsLoaded + loadQty
								existItem.CartonsLoaded = totalPcsCombined / pcsPerUnit
								existItem.PcsLoaded = totalPcsCombined % pcsPerUnit
								existItem.CustomPcsPerUnit = pcsPerUnit
							}
							if returnQty > 0 {
								existItem.PcsReturnedToWarehouse += returnQty
							}
							return tx.Save(&existItem).Error
						}
					}

					prodIDPtr := *prodID
					newItem := domain.WarehouseDeliveryOrderItem{
						WarehouseDeliveryOrderID: wdoID,
						ProductID:                &prodIDPtr,
						CartonsLoaded:            loadQty / pcsPerUnit,
						PcsLoaded:                loadQty % pcsPerUnit,
						CustomPcsPerUnit:         pcsPerUnit,
						PcsReturnedToWarehouse:   returnQty,
					}
					return tx.Create(&newItem).Error
				}
			} else if bonusID != nil {
				var inv domain.BonusInventory
				if errDb := tx.First(&inv, "id = ?", *bonusID).Error; errDb == nil {
					if isApp {
						var existItem domain.WarehouseDeliveryOrderItem
						errItem := tx.Where("warehouse_delivery_order_id = ? AND bonus_inventory_id = ?", wdoID, *bonusID).First(&existItem).Error
						if errItem == nil {
							if loadQty > 0 {
								existItem.PcsLoaded += loadQty
							}
							if returnQty > 0 {
								existItem.PcsReturnedToWarehouse += returnQty
							}
							return tx.Save(&existItem).Error
						}
					}

					bonusIDPtr := *bonusID
					newItem := domain.WarehouseDeliveryOrderItem{
						WarehouseDeliveryOrderID: wdoID,
						BonusInventoryID:         &bonusIDPtr,
						CartonsLoaded:            0,
						PcsLoaded:                loadQty,
						CustomPcsPerUnit:         1,
						PcsReturnedToWarehouse:   returnQty,
					}
					return tx.Create(&newItem).Error
				}
			}
			return nil
		}

		// Build DO Items for Load (TRANSFER)
		for prodID, totalPcs := range loadProductQtyMap {
			if err := updateOrBuildDOItem(wdoID, &prodID, nil, totalPcs, 0, isAppending); err != nil {
				return err
			}
		}
		for bonusID, totalPcs := range loadBonusQtyMap {
			if err := updateOrBuildDOItem(wdoID, nil, &bonusID, totalPcs, 0, isAppending); err != nil {
				return err
			}
		}

		// Build DO Items for Return (RETURN)
		for prodID, totalPcs := range returnProductQtyMap {
			if err := updateOrBuildDOItem(wdoID, &prodID, nil, 0, totalPcs, isAppending); err != nil {
				return err
			}
		}
		for bonusID, totalPcs := range returnBonusQtyMap {
			if err := updateOrBuildDOItem(wdoID, nil, &bonusID, 0, totalPcs, isAppending); err != nil {
				return err
			}
		}

		// Simpan relasi ke SalesTransfer/SalesBonusRequest dan ubah status ke PROCESSING
		for _, trfID := range salesTransferIDs {
			var trf domain.SalesTransfer
			if err := tx.First(&trf, "id = ?", trfID).Error; err == nil {
				trfIDPtr := trfID
				rel := domain.WarehouseDeliveryOrderTransfer{
					WarehouseDeliveryOrderID: wdoID,
					SalesTransferID:          &trfIDPtr,
				}
				if err := tx.Create(&rel).Error; err != nil {
					return err
				}
				if err := tx.Model(&domain.SalesTransfer{}).Where("id = ?", trfID).Update("status", domain.SalesTransferProcessing).Error; err != nil {
					return err
				}
			} else {
				var bonusReq domain.SalesBonusRequest
				if err := tx.First(&bonusReq, "id = ?", trfID).Error; err == nil {
					bonusIDPtr := trfID
					rel := domain.WarehouseDeliveryOrderTransfer{
						WarehouseDeliveryOrderID: wdoID,
						SalesBonusRequestID:      &bonusIDPtr,
					}
					if err := tx.Create(&rel).Error; err != nil {
						return err
					}
					if err := tx.Model(&domain.SalesBonusRequest{}).Where("id = ?", trfID).Update("status", "PROCESSING").Error; err != nil {
						return err
					}
				}
			}
		}

		var fetchErr error
		result, fetchErr = repository.NewWarehouseDeliveryOrderRepository(tx).FindByID(wdoID)
		return fetchErr
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (u *warehouseDeliveryOrderUsecase) GetMotorisDOs(branchID *uuid.UUID) ([]domain.WarehouseDeliveryOrder, error) {
	return u.repo.ListMotoris(branchID)
}

func (u *warehouseDeliveryOrderUsecase) UpdateDO(req UpdateWarehouseDORequest, userID uuid.UUID, userName string, ip string) (*domain.WarehouseDeliveryOrder, error) {
	var wdo *domain.WarehouseDeliveryOrder
	err := u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewWarehouseDeliveryOrderRepository(tx)
		var err error
		wdo, err = txRepo.FindByID(req.ID)
		if err != nil {
			return err
		}

		oldNotes := wdo.Notes
		wdo.Notes = req.Notes

		if req.Status != "" {
			wdo.Status = req.Status
		}

		// Map old items for comparison in audit log
		// Map old items for comparison in audit log
		oldItemsMap := make(map[string]domain.WarehouseDeliveryOrderItem)
		for _, item := range wdo.Items {
			if item.ProductID != nil {
				oldItemsMap[item.ProductID.String()] = item
			} else if item.BonusInventoryID != nil {
				oldItemsMap[item.BonusInventoryID.String()] = item
			}
		}

		// Rebuild items list
		var newItems []domain.WarehouseDeliveryOrderItem
		var auditLogs []string

		for _, itemReq := range req.Items {
			var prodName string = "Barang"
			var itemKey string

			if itemReq.ProductID != nil && *itemReq.ProductID != uuid.Nil {
				itemKey = itemReq.ProductID.String()
				var prod domain.Product
				if errDb := tx.Select("name").First(&prod, "id = ?", *itemReq.ProductID).Error; errDb == nil {
					prodName = prod.Name
				}
			} else if itemReq.BonusInventoryID != nil && *itemReq.BonusInventoryID != uuid.Nil {
				itemKey = itemReq.BonusInventoryID.String()
				var inv domain.BonusInventory
				if errDb := tx.Select("name").First(&inv, "id = ?", *itemReq.BonusInventoryID).Error; errDb == nil {
					prodName = fmt.Sprintf("Bonus: %s", inv.Name)
				}
			}

			if itemKey != "" {
				oldItem, exists := oldItemsMap[itemKey]
				if exists {
					if oldItem.CartonsLoaded != itemReq.CartonsLoaded || oldItem.PcsLoaded != itemReq.PcsLoaded {
						auditLogs = append(auditLogs, fmt.Sprintf("- %s: %d kardus %d pcs (sebelumnya %d kardus %d pcs)",
							prodName, itemReq.CartonsLoaded, itemReq.PcsLoaded, oldItem.CartonsLoaded, oldItem.PcsLoaded))
					}
				} else {
					auditLogs = append(auditLogs, fmt.Sprintf("- %s (Baru): %d kardus %d pcs", prodName, itemReq.CartonsLoaded, itemReq.PcsLoaded))
				}
			}

			customUnits := itemReq.CustomPcsPerUnit
			if customUnits <= 0 {
				customUnits = 1
			}

			newItems = append(newItems, domain.WarehouseDeliveryOrderItem{
				WarehouseDeliveryOrderID: wdo.ID,
				ProductID:                itemReq.ProductID,
				BonusInventoryID:         itemReq.BonusInventoryID,
				CartonsLoaded:            itemReq.CartonsLoaded,
				PcsLoaded:                itemReq.PcsLoaded,
				CustomPcsPerUnit:         customUnits,
				PcsReturnedToWarehouse:   itemReq.PcsReturnedToWarehouse,
				PcsSold:                  itemReq.PcsSold,
				PcsSample:                itemReq.PcsSample,
				PcsReturnedFromCustomer:  itemReq.PcsReturnedFromCustomer,
			})
		}

		// Detect deleted items
		newItemsMap := make(map[string]bool)
		for _, item := range newItems {
			if item.ProductID != nil {
				newItemsMap[item.ProductID.String()] = true
			} else if item.BonusInventoryID != nil {
				newItemsMap[item.BonusInventoryID.String()] = true
			}
		}
		for idStr, oldItem := range oldItemsMap {
			if !newItemsMap[idStr] {
				itemName := "Barang"
				if oldItem.ProductID != nil {
					var prod domain.Product
					if errDb := tx.Select("name").First(&prod, "id = ?", *oldItem.ProductID).Error; errDb == nil {
						itemName = prod.Name
					}
				} else if oldItem.BonusInventoryID != nil {
					var inv domain.BonusInventory
					if errDb := tx.Select("name").First(&inv, "id = ?", *oldItem.BonusInventoryID).Error; errDb == nil {
						itemName = fmt.Sprintf("Bonus: %s", inv.Name)
					}
				}
				auditLogs = append(auditLogs, fmt.Sprintf("- %s (Dihapus): sebelumnya %d kardus %d pcs", itemName, oldItem.CartonsLoaded, oldItem.PcsLoaded))
			}
		}

		wdo.Items = newItems

		if err := txRepo.Update(wdo); err != nil {
			return err
		}

		// Log changes in AuditLog
		companyID := uuid.Nil
		if wdo.DeliveryBatch != nil && wdo.DeliveryBatch.CompanyID != nil {
			companyID = *wdo.DeliveryBatch.CompanyID
		}

		auditDesc := "Mengubah Surat DO (Bukti Muat Gudang)."
		if oldNotes != req.Notes {
			auditDesc += fmt.Sprintf(" Catatan diubah dari '%s' menjadi '%s'.", oldNotes, req.Notes)
		}
		if len(auditLogs) > 0 {
			auditDesc += " Perubahan rincian muat:\n" + fmt.Sprintf("%v", auditLogs)
		}

		if len(auditLogs) > 0 || oldNotes != req.Notes {
			u.auditUsecase.Log(companyID, userID, userName, "UPDATE", "WarehouseDeliveryOrder", wdo.ID.String(), auditDesc, oldItemsMap, req.Items, ip)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return u.repo.FindByID(req.ID)
}

func (u *warehouseDeliveryOrderUsecase) GetByBatchID(batchID uuid.UUID) (*domain.WarehouseDeliveryOrder, error) {
	return u.repo.FindByBatchID(batchID)
}

func (u *warehouseDeliveryOrderUsecase) GetByID(id uuid.UUID) (*domain.WarehouseDeliveryOrder, error) {
	return u.repo.FindByID(id)
}

func (u *warehouseDeliveryOrderUsecase) DispatchDO(id uuid.UUID, userID uuid.UUID, userName string, ip string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewWarehouseDeliveryOrderRepository(tx)
		wdo, err := txRepo.FindByID(id)
		if err != nil {
			return err
		}

		if wdo.Status == "DISPATCHED" || wdo.Status == "FINALIZED" {
			return fmt.Errorf("DO ini sudah di-dispatch atau difinalisasi")
		}

		wdo.Status = "DISPATCHED"
		if err := txRepo.Update(wdo); err != nil {
			return err
		}

		now := time.Now()

		// ── DO MOTORIS: selesaikan SalesTransfer + potong stok gudang + tambah SalesStock ──
		if wdo.DOType == domain.DOTypeMotoris {
			wRepo := repository.NewWarehouseRepository(tx)
			salesTransferRepo := repository.NewSalesTransferRepository(tx)
			bonusRepo := repository.NewBonusInventoryRepository(tx)
			salesBonusRepo := repository.NewSalesBonusRepository(tx)

			for _, rel := range wdo.SalesTransfers {
				if rel.SalesTransferID != nil {
					var transfer domain.SalesTransfer
					if err := tx.First(&transfer, "id = ?", *rel.SalesTransferID).Error; err != nil {
						continue
					}
					if transfer.Status == domain.SalesTransferCompleted {
						continue
					}

					if transfer.ProductID != nil {
						// 1. Potong stok gudang (FEFO)
						batches, err := wRepo.GetAvailableBatches(transfer.BranchID, *transfer.ProductID, "GOOD")
						if err != nil {
							return fmt.Errorf("gagal ambil stok gudang untuk produk %s: %w", *transfer.ProductID, err)
						}
						remaining := transfer.Quantity
						for i := range batches {
							b := &batches[i]
							available := b.Quantity
							if available <= 0 {
								continue
							}
							deduct := available
							if deduct > remaining {
								deduct = remaining
							}
							b.Quantity -= deduct
							remaining -= deduct
							if err := wRepo.UpdateStock(b); err != nil {
								return err
							}
							_ = wRepo.CreateLog(&domain.WarehouseLog{
								BranchID:  transfer.BranchID,
								ProductID: *transfer.ProductID,
								Type:      "OUT",
								Source:    fmt.Sprintf("DO_MOTORIS:%s", wdo.DONumber),
								Quantity:  deduct,
							})
							if remaining <= 0 {
								break
							}
						}
						if remaining > 0 {
							return fmt.Errorf("stok gudang tidak cukup untuk produk %s (kurang %d pcs)", *transfer.ProductID, remaining)
						}

						// 2. Tambah SalesStock Salesman
						sStock, errStock := salesTransferRepo.GetSalesStock(transfer.EmployeeID, *transfer.ProductID)
						if errStock != nil {
							sStock = &domain.SalesStock{
								EmployeeID: transfer.EmployeeID,
								ProductID:  *transfer.ProductID,
								Quantity:   transfer.Quantity,
								UpdatedAt:  now,
							}
							if err := tx.Create(sStock).Error; err != nil {
								return err
							}
						} else {
							sStock.Quantity += transfer.Quantity
							sStock.UpdatedAt = now
							if err := salesTransferRepo.UpdateSalesStock(sStock); err != nil {
								return err
							}
						}
					} else if transfer.BonusInventoryID != nil {
						// 1. Potong stok gudang bonus
						inv, err := bonusRepo.GetByID(tx.Statement.Context, *transfer.BonusInventoryID)
						if err != nil {
							return fmt.Errorf("gagal ambil stok gudang untuk barang bonus %s: %w", *transfer.BonusInventoryID, err)
						}
						if inv.Quantity >= transfer.Quantity {
							inv.Quantity -= transfer.Quantity
						} else {
							inv.Quantity = 0
						}
						if err := bonusRepo.Update(tx.Statement.Context, inv); err != nil {
							return err
						}
						_ = bonusRepo.CreateLog(tx.Statement.Context, &domain.BonusInventoryLog{
							BonusInventoryID: inv.ID,
							Type:             domain.BonusInventoryOut,
							Quantity:         transfer.Quantity,
							Notes:            fmt.Sprintf("DO_MOTORIS:%s", wdo.DONumber),
						})

						// 2. Tambah SalesBonusStock Salesman
						sStock, errStock := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, transfer.EmployeeID, *transfer.BonusInventoryID)
						if errStock != nil {
							sStock = &domain.SalesBonusStock{
								EmployeeID:       transfer.EmployeeID,
								BonusInventoryID: *transfer.BonusInventoryID,
								Quantity:         transfer.Quantity,
								UpdatedAt:        now,
							}
							if err := salesBonusRepo.CreateSalesBonusStock(tx.Statement.Context, sStock); err != nil {
								return err
							}
						} else {
							sStock.Quantity += transfer.Quantity
							sStock.UpdatedAt = now
							if err := salesBonusRepo.UpdateSalesBonusStock(tx.Statement.Context, sStock); err != nil {
								return err
							}
						}
					}

					// 3. Selesaikan SalesTransfer
					transfer.Status = domain.SalesTransferCompleted
					transfer.UpdatedAt = now
					if err := tx.Save(&transfer).Error; err != nil {
						return err
					}
				} else if rel.SalesBonusRequestID != nil {
					var bonusReq domain.SalesBonusRequest
					if err := tx.First(&bonusReq, "id = ?", *rel.SalesBonusRequestID).Error; err != nil {
						continue
					}
					if bonusReq.Status == domain.SalesBonusRequestApproved {
						continue
					}

					// 1. Potong stok gudang bonus
					inv, err := bonusRepo.GetByID(tx.Statement.Context, bonusReq.BonusInventoryID)
					if err != nil {
						return fmt.Errorf("gagal ambil stok gudang untuk barang bonus %s: %w", bonusReq.BonusInventoryID, err)
					}
					if inv.Quantity >= bonusReq.Quantity {
						inv.Quantity -= bonusReq.Quantity
					} else {
						inv.Quantity = 0
					}
					if err := bonusRepo.Update(tx.Statement.Context, inv); err != nil {
						return err
					}
					_ = bonusRepo.CreateLog(tx.Statement.Context, &domain.BonusInventoryLog{
						BonusInventoryID: inv.ID,
						Type:             domain.BonusInventoryOut,
						Quantity:         bonusReq.Quantity,
						Notes:            fmt.Sprintf("DO_MOTORIS:%s", wdo.DONumber),
					})

					// 2. Tambah SalesBonusStock Salesman
					sStock, errStock := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, bonusReq.EmployeeID, bonusReq.BonusInventoryID)
					if errStock != nil {
						sStock = &domain.SalesBonusStock{
							EmployeeID:       bonusReq.EmployeeID,
							BonusInventoryID: bonusReq.BonusInventoryID,
							Quantity:         bonusReq.Quantity,
							UpdatedAt:        now,
						}
						if err := salesBonusRepo.CreateSalesBonusStock(tx.Statement.Context, sStock); err != nil {
							return err
						}
					} else {
						sStock.Quantity += bonusReq.Quantity
						sStock.UpdatedAt = now
						if err := salesBonusRepo.UpdateSalesBonusStock(tx.Statement.Context, sStock); err != nil {
							return err
						}
					}

					// 3. Selesaikan Bonus Request (di sini kita update ke APPROVED)
					bonusReq.Status = domain.SalesBonusRequestApproved
					bonusReq.UpdatedAt = now
					if err := tx.Save(&bonusReq).Error; err != nil {
						return err
					}
				}
			}
		}

		// ── DO REGULAR: selesaikan SalesTransfer yang ada di DeliveryBatch ──
		if wdo.DOType == domain.DOTypeRegular && wdo.DeliveryBatchID != nil {
			var items []domain.DeliveryItem
			if err := tx.Where("delivery_batch_id = ?", *wdo.DeliveryBatchID).Find(&items).Error; err == nil {
				allDelivered := true
				for _, item := range items {
					if item.SalesTransferID != nil {
						if item.Status != domain.DeliveryItemDelivered {
							item.Status = domain.DeliveryItemDelivered
							if err := tx.Save(&item).Error; err != nil {
								return err
							}
						}
						var transfer domain.SalesTransfer
						if err := tx.First(&transfer, "id = ?", *item.SalesTransferID).Error; err == nil {
							if transfer.Status != domain.SalesTransferCompleted {
								transfer.Status = domain.SalesTransferCompleted
								transfer.UpdatedAt = now
								if err := tx.Save(&transfer).Error; err != nil {
									return err
								}
								salesTransferRepo := repository.NewSalesTransferRepository(tx)
								if transfer.ProductID != nil {
									sStock, errStock := salesTransferRepo.GetSalesStock(transfer.EmployeeID, *transfer.ProductID)
									if errStock != nil {
										sStock = &domain.SalesStock{
											EmployeeID: transfer.EmployeeID,
											ProductID:  *transfer.ProductID,
											Quantity:   transfer.Quantity,
											UpdatedAt:  now,
										}
										if err := tx.Create(sStock).Error; err != nil {
											return err
										}
									} else {
										sStock.Quantity += transfer.Quantity
										sStock.UpdatedAt = now
										if err := salesTransferRepo.UpdateSalesStock(sStock); err != nil {
											return err
										}
									}
								} else if transfer.BonusInventoryID != nil {
									salesBonusRepo := repository.NewSalesBonusRepository(tx)
									sStock, errStock := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, transfer.EmployeeID, *transfer.BonusInventoryID)
									if errStock != nil {
										sStock = &domain.SalesBonusStock{
											EmployeeID:       transfer.EmployeeID,
											BonusInventoryID: *transfer.BonusInventoryID,
											Quantity:         transfer.Quantity,
											UpdatedAt:        now,
										}
										if err := salesBonusRepo.CreateSalesBonusStock(tx.Statement.Context, sStock); err != nil {
											return err
										}
									} else {
										sStock.Quantity += transfer.Quantity
										sStock.UpdatedAt = now
										if err := salesBonusRepo.UpdateSalesBonusStock(tx.Statement.Context, sStock); err != nil {
											return err
										}
									}
								}
							}
						}
					} else if item.Status != domain.DeliveryItemDelivered {
						allDelivered = false
					}
				}
				if allDelivered {
					tx.Model(&domain.DeliveryBatch{}).Where("id = ?", *wdo.DeliveryBatchID).Update("status", domain.DeliveryBatchCompleted)
				}
			}
		}

		companyID := uuid.Nil
		if wdo.DeliveryBatch != nil && wdo.DeliveryBatch.CompanyID != nil {
			companyID = *wdo.DeliveryBatch.CompanyID
		}

		u.auditUsecase.Log(companyID, userID, userName, "DISPATCH", "WarehouseDeliveryOrder", wdo.ID.String(), "Dispatch Surat DO (Bukti Muat Gudang). Stok dipotong dan ditransfer ke Sales.", nil, nil, ip)

		return nil
	})
}

func (u *warehouseDeliveryOrderUsecase) FinalizeDO(id uuid.UUID, userID uuid.UUID, userName string, ip string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewWarehouseDeliveryOrderRepository(tx)
		wdo, err := txRepo.FindByID(id)
		if err != nil {
			return err
		}

		if wdo.Status == "FINALIZED" {
			return fmt.Errorf("DO ini sudah difinalisasi")
		}

		wdo.Status = "FINALIZED"
		if err := txRepo.Update(wdo); err != nil {
			return err
		}

		// Kembalikan stok "Pulang" ke gudang, dan kurangi dari SalesStock
		now := time.Now()
		salesTransferRepo := repository.NewSalesTransferRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		// For Regular DO we might not have a direct EmployeeID on the WDO, it's on the DeliveryBatch.DriverID
		// or SalesTransfer.EmployeeID. For Motoris it's wdo.EmployeeID.
		// So we loop over items and find the proper employee ID.
		var employeeID *uuid.UUID
		if wdo.EmployeeID != nil {
			employeeID = wdo.EmployeeID
		} else if wdo.DeliveryBatch != nil && wdo.DeliveryBatch.DriverID != nil {
			employeeID = wdo.DeliveryBatch.DriverID
		}

		for _, item := range wdo.Items {
			if item.PcsReturnedToWarehouse > 0 {
				// 1. Tambah stok gudang
				// Get or create a branch info. For now, assume branch from wdo.DeliveryBatch.BranchID or SalesTransfer
				// For simplicity, just find a default batch or active batch to add back to
				// A simpler way: use UpdateStock directly or create a new batch logic.
				// Since we don't have the exact batch it was taken from easily, we can add it to an existing GOOD batch.
				var branchID uuid.UUID
				if wdo.DeliveryBatch != nil && wdo.DeliveryBatch.BranchID != nil {
					branchID = *wdo.DeliveryBatch.BranchID
				} else if wdo.EmployeeID != nil {
					if len(wdo.SalesTransfers) > 0 {
						var trf domain.SalesTransfer
						if wdo.SalesTransfers[0].SalesTransferID != nil {
							if err := tx.First(&trf, "id = ?", *wdo.SalesTransfers[0].SalesTransferID).Error; err == nil {
								employeeID = &trf.EmployeeID
								branchID = trf.BranchID
							}
						} else if wdo.SalesTransfers[0].SalesBonusRequestID != nil {
							var bonusReq domain.SalesBonusRequest
							if err := tx.First(&bonusReq, "id = ?", *wdo.SalesTransfers[0].SalesBonusRequestID).Error; err == nil {
								employeeID = &bonusReq.EmployeeID
							}
						}
					}
				}

				if branchID != uuid.Nil {
					if item.ProductID != nil {
						// Find an existing active batch for this product in the branch
						var wb domain.WarehouseStock
						errWb := tx.Where("branch_id = ? AND product_id = ? AND stock_type = 'GOOD'", branchID, *item.ProductID).
							Order("expiry_date ASC").First(&wb).Error
						if errWb == nil {
							wb.Quantity += item.PcsReturnedToWarehouse
							tx.Save(&wb)
						} else {
							// Create a new generic batch
							wb = domain.WarehouseStock{
								ID:        uuid.New(),
								BranchID:  branchID,
								ProductID: *item.ProductID,
								Quantity:  item.PcsReturnedToWarehouse,
								StockType: "GOOD",
								BatchNo:   "RETURN-" + wdo.DONumber,
								CreatedAt: now,
								UpdatedAt: now,
							}
							tx.Create(&wb)
						}

						_ = wRepo.CreateLog(&domain.WarehouseLog{
							BranchID:  branchID,
							ProductID: *item.ProductID,
							Type:      "IN",
							Source:    fmt.Sprintf("DO_RETURN:%s", wdo.DONumber),
							Quantity:  item.PcsReturnedToWarehouse,
						})
					} else if item.BonusInventoryID != nil {
						bonusRepo := repository.NewBonusInventoryRepository(tx)
						inv, err := bonusRepo.GetByID(tx.Statement.Context, *item.BonusInventoryID)
						if err == nil {
							inv.Quantity += item.PcsReturnedToWarehouse
							_ = bonusRepo.Update(tx.Statement.Context, inv)
							_ = bonusRepo.CreateLog(tx.Statement.Context, &domain.BonusInventoryLog{
								BonusInventoryID: inv.ID,
								Type:             domain.BonusInventoryIn,
								Quantity:         item.PcsReturnedToWarehouse,
								Notes:            fmt.Sprintf("DO_RETURN:%s", wdo.DONumber),
							})
						}
					}
				}

				// 2. Kurangi SalesStock
				if employeeID != nil {
					if item.ProductID != nil {
						sStock, errStock := salesTransferRepo.GetSalesStock(*employeeID, *item.ProductID)
						if errStock == nil {
							sStock.Quantity -= item.PcsReturnedToWarehouse
							if sStock.Quantity < 0 {
								sStock.Quantity = 0
							}
							sStock.UpdatedAt = now
							_ = salesTransferRepo.UpdateSalesStock(sStock)
						}
					} else if item.BonusInventoryID != nil {
						salesBonusRepo := repository.NewSalesBonusRepository(tx)
						sStock, errStock := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, *employeeID, *item.BonusInventoryID)
						if errStock == nil {
							sStock.Quantity -= item.PcsReturnedToWarehouse
							if sStock.Quantity < 0 {
								sStock.Quantity = 0
							}
							sStock.UpdatedAt = now
							_ = salesBonusRepo.UpdateSalesBonusStock(tx.Statement.Context, sStock)
						}
					}
				}

				// 3. Update PENDING RETURN transfers for this employee & product to COMPLETED
				if employeeID != nil {
					if item.ProductID != nil {
						tx.Model(&domain.SalesTransfer{}).
							Where("employee_id = ? AND product_id = ? AND type = 'RETURN' AND status = 'PENDING'", *employeeID, *item.ProductID).
							Update("status", "COMPLETED")
					} else if item.BonusInventoryID != nil {
						tx.Model(&domain.SalesTransfer{}).
							Where("employee_id = ? AND bonus_inventory_id = ? AND type = 'RETURN' AND status = 'PENDING'", *employeeID, *item.BonusInventoryID).
							Update("status", "COMPLETED")
					}
				}
			}
		}

		companyID := uuid.Nil
		if wdo.DeliveryBatch != nil && wdo.DeliveryBatch.CompanyID != nil {
			companyID = *wdo.DeliveryBatch.CompanyID
		}

		u.auditUsecase.Log(companyID, userID, userName, "FINALIZE", "WarehouseDeliveryOrder", wdo.ID.String(), "Finalisasi Surat DO (Bukti Muat Gudang). Sisa barang direkonsiliasi.", nil, nil, ip)

		return nil
	})
}
