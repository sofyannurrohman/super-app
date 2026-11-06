package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/sse"
	"gorm.io/gorm"
)

type WarehouseUsecase interface {
	GetInventory(branchID uuid.UUID) ([]domain.WarehouseStock, error)
	GetPendingShipments(branchID uuid.UUID) ([]domain.ProductTransfer, error)
	ReceiveShipment(transferID uuid.UUID, actualQty, damagedQty int, notes string, processedBy string) error
	GetTransferByDO(doNo string) (*domain.ProductTransfer, error)
	ReceiveShipmentByDO(doNo string, actualQty, damagedQty int, notes string, processedBy string) error
	CorrectShipmentReceipt(transferID uuid.UUID, actualQty, damagedQty int, reason string, correctedBy string) error
	GetWarehouseLogs(branchID uuid.UUID) ([]domain.WarehouseLog, error)
	GetAllTransfers(branchID uuid.UUID) ([]domain.ProductTransfer, error)
	AdjustStock(branchID, productID uuid.UUID, quantity int, reason string, batchNo string, processedBy string, stockType string) error

	ApproveTransfer(transferID uuid.UUID, approverID *uuid.UUID, approverName, approverPosition string) error
	RejectTransfer(transferID uuid.UUID, reason string) error
	ConfirmArrival(transferID uuid.UUID) error
	ConfirmShipment(transferID uuid.UUID) error

	// Executor Workflow
	DispatchByInvoice(receiptNo string, branchID uuid.UUID) error
	SetStockLimit(branchID, productID uuid.UUID, limit int, batchNo string) error
	DeleteStock(stockID, branchID uuid.UUID, reason, processedBy string) error
}

type warehouseUsecase struct {
	repo      repository.WarehouseRepository
	notifRepo repository.NotificationRepository
	salesRepo repository.SalesTransactionRepository
	soRepo    repository.SalesOrderRepository
	db        *gorm.DB
	sseHub    *sse.Hub
}

func NewWarehouseUsecase(
	repo repository.WarehouseRepository,
	notifRepo repository.NotificationRepository,
	salesRepo repository.SalesTransactionRepository,
	soRepo repository.SalesOrderRepository,
	db *gorm.DB,
	sseHub *sse.Hub,
) WarehouseUsecase {
	return &warehouseUsecase{repo, notifRepo, salesRepo, soRepo, db, sseHub}
}

func (u *warehouseUsecase) GetInventory(branchID uuid.UUID) ([]domain.WarehouseStock, error) {
	return u.repo.GetStocksByBranchID(branchID)
}

func (u *warehouseUsecase) GetPendingShipments(branchID uuid.UUID) ([]domain.ProductTransfer, error) {
	return u.repo.GetPendingTransfers(branchID)
}

func (u *warehouseUsecase) ReceiveShipment(transferID uuid.UUID, actualQty, damagedQty int, notes string, processedBy string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewWarehouseRepository(tx)

		transfer, err := repo.GetTransferByID(transferID)
		if err != nil {
			return err
		}
		if transfer.Status != "SHIPPED" && transfer.Status != "ARRIVED" {
			return errors.New("shipment is not in SHIPPED or ARRIVED status")
		}

		// If both are 0, assume full receive (legacy or quick scan support)
		if actualQty == 0 && damagedQty == 0 {
			actualQty = transfer.Quantity
		}

		batchNo := transfer.BatchNo
		if batchNo == "" {
			batchNo = "DEFAULT"
		}

		// Update GOOD stock with actual received quantity
		if actualQty > 0 {
			stock, err := repo.GetStock(transfer.ToBranchID, transfer.ProductID, "GOOD", batchNo)
			if err != nil {
				return err
			}
			stock.Quantity += actualQty
			stock.ExpiryDate = transfer.ExpiryDate
			if err := repo.UpdateStock(stock); err != nil {
				return err
			}

			// Log GOOD movement
			log := &domain.WarehouseLog{
				BranchID:    transfer.ToBranchID,
				ProductID:   transfer.ProductID,
				Type:        "IN",
				Source:      "FACTORY_TRANSFER (GOOD)",
				Quantity:    actualQty,
				ProcessedBy: processedBy,
				ApprovedBy:  transfer.InitiatedBy, // If initiated by FACTORY, we can assume it's approved by them, or just let it be. Let's use processedBy.
			}
			if err := repo.CreateLog(log); err != nil {
				return err
			}
		}

		// Update DAMAGED stock with damaged quantity
		if damagedQty > 0 {
			damagedStock, err := repo.GetStock(transfer.ToBranchID, transfer.ProductID, "DAMAGED", batchNo)
			if err != nil {
				return err
			}
			damagedStock.Quantity += damagedQty
			damagedStock.ExpiryDate = transfer.ExpiryDate
			if err := repo.UpdateStock(damagedStock); err != nil {
				return err
			}

			// Log DAMAGED movement
			log := &domain.WarehouseLog{
				BranchID:    transfer.ToBranchID,
				ProductID:   transfer.ProductID,
				Type:        "IN",
				Source:      "FACTORY_TRANSFER (DAMAGED)",
				Quantity:    damagedQty,
				ProcessedBy: processedBy,
			}
			if err := repo.CreateLog(log); err != nil {
				return err
			}
		}

		transfer.Status = "RECEIVED"
		now := time.Now()
		transfer.ReceivedAt = &now
		transfer.ActualReceivedQuantity = actualQty
		transfer.DamagedQuantity = damagedQty
		if notes != "" {
			if transfer.Notes != "" {
				transfer.Notes += "\nCatatan Penerimaan: " + notes
			} else {
				transfer.Notes = notes
			}
		}

		if err := repo.UpdateTransfer(transfer); err != nil {
			return err
		}

		// Trigger Auto-Fulfillment only for GOOD stock
		if actualQty > 0 {
			return u.internalRunAutoFulfillment(tx, transfer.ToBranchID, transfer.ProductID)
		}
		return nil
	})
}

func (u *warehouseUsecase) internalRunAutoFulfillment(tx *gorm.DB, branchID, productID uuid.UUID) error {
	soRepo := repository.NewSalesOrderRepository(tx)
	wRepo := repository.NewWarehouseRepository(tx)

	// 1. Ambil semua batch yang punya stok tersedia (FEFO)
	availableBatches, err := wRepo.GetAvailableBatches(branchID, productID, "GOOD")
	if err != nil {
		return err
	}

	// 2. Cari semua rincian reservasi BACKORDER untuk produk ini
	var backorderReservations []domain.SalesOrderItemBatch
	if err := tx.Joins("JOIN sales_order_items ON sales_order_items.id = sales_order_item_batches.sales_order_item_id").
		Joins("JOIN sales_orders ON sales_orders.id = sales_order_items.sales_order_id").
		Where("sales_order_items.product_id = ? AND sales_order_item_batches.batch_no = 'BACKORDER' AND sales_order_item_batches.quantity > 0", productID).
		Where("sales_orders.branch_id = ?", branchID).
		Order("sales_orders.created_at ASC").
		Find(&backorderReservations).Error; err != nil {
		return err
	}

	// 3. Alokasikan stok baru ke reservasi BACKORDER (FIFO)
	for i := range backorderReservations {
		res := &backorderReservations[i]

		for _, b := range availableBatches {
			availableInBatch := b.Quantity - b.ReservedQuantity
			if availableInBatch <= 0 {
				continue
			}

			moveQty := res.Quantity
			if availableInBatch < moveQty {
				moveQty = availableInBatch
			}

			// Pindahkan reservasi:
			// a. Kurangi dari BACKORDER
			if err := wRepo.ReleaseStock(branchID, productID, "BACKORDER", moveQty); err != nil {
				return err
			}
			// b. Tambah ke Batch Riil
			if err := wRepo.ReserveStock(branchID, productID, b.BatchNo, moveQty); err != nil {
				return err
			}

			// c. Update/Split rincian batch di SO Item
			res.Quantity -= moveQty
			if res.Quantity == 0 {
				tx.Delete(res)
			} else {
				tx.Save(res)
			}

			// Tambah rincian batch baru
			newRes := &domain.SalesOrderItemBatch{
				SalesOrderItemID: res.SalesOrderItemID,
				BatchNo:          b.BatchNo,
				Quantity:         moveQty,
			}
			tx.Create(newRes)

			// Update total available in batch for next iteration
			b.ReservedQuantity += moveQty
			if res.Quantity <= 0 {
				break
			}
		}
	}

	// 4. Cek apakah ada SO yang sekarang sudah fully fulfilled
	orders, err := soRepo.FindWaitingStockByProduct(branchID, productID)
	if err != nil {
		return err
	}

	for _, so := range orders {
		allFulfilled := true
		for _, item := range so.Items {
			// Check if there's any BACKORDER batch left for this item
			var boCount int64
			tx.Model(&domain.SalesOrderItemBatch{}).Where("sales_order_item_id = ? AND batch_no = 'BACKORDER'", item.ID).Count(&boCount)
			if boCount > 0 {
				allFulfilled = false
				break
			}
		}

		if allFulfilled {
			so.Status = domain.SOStatusConfirmed
			soRepo.Update(&so)
		}
	}

	return nil
}

func (u *warehouseUsecase) GetTransferByDO(doNo string) (*domain.ProductTransfer, error) {
	return u.repo.GetTransferByDO(doNo)
}

func (u *warehouseUsecase) ReceiveShipmentByDO(doNo string, actualQty, damagedQty int, notes string, processedBy string) error {
	transfer, err := u.repo.GetTransferByDO(doNo)
	if err != nil {
		return err
	}
	if transfer == nil {
		return errors.New("surat jalan tidak ditemukan")
	}
	return u.ReceiveShipment(transfer.ID, actualQty, damagedQty, notes, processedBy)
}

func (u *warehouseUsecase) CorrectShipmentReceipt(transferID uuid.UUID, actualQty, damagedQty int, reason string, correctedBy string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewWarehouseRepository(tx)

		transfer, err := repo.GetTransferByID(transferID)
		if err != nil {
			return err
		}
		if transfer.Status != "RECEIVED" {
			return errors.New("hanya pengiriman dengan status RECEIVED yang dapat dikoreksi")
		}

		batchNo := transfer.BatchNo
		if batchNo == "" {
			batchNo = "DEFAULT"
		}

		oldActual := transfer.ActualReceivedQuantity
		oldDamaged := transfer.DamagedQuantity

		diffGood := actualQty - oldActual
		diffDamaged := damagedQty - oldDamaged

		// 1. Adjust GOOD stock
		if diffGood != 0 {
			stock, err := repo.GetStock(transfer.ToBranchID, transfer.ProductID, "GOOD", batchNo)
			if err != nil {
				return err
			}
			stock.Quantity += diffGood
			if stock.Quantity < 0 {
				return fmt.Errorf("koreksi tidak valid: stok akhir GOOD tidak boleh kurang dari 0 (stok saat ini: %d, penyesuaian: %d)", stock.Quantity-diffGood, diffGood)
			}
			if err := repo.UpdateStock(stock); err != nil {
				return err
			}

			// Log mutasi
			logType := "IN"
			logQty := diffGood
			if diffGood < 0 {
				logType = "OUT"
				logQty = -diffGood
			}
			log := &domain.WarehouseLog{
				BranchID:    transfer.ToBranchID,
				ProductID:   transfer.ProductID,
				Type:        logType,
				Source:      "FACTORY_TRANSFER_ADJUSTMENT (GOOD)",
				Quantity:    logQty,
				ProcessedBy: correctedBy,
				ApprovedBy:  correctedBy,
			}
			if err := repo.CreateLog(log); err != nil {
				return err
			}
		}

		// 2. Adjust DAMAGED stock
		if diffDamaged != 0 {
			damagedStock, err := repo.GetStock(transfer.ToBranchID, transfer.ProductID, "DAMAGED", batchNo)
			if err != nil {
				return err
			}
			damagedStock.Quantity += diffDamaged
			if damagedStock.Quantity < 0 {
				return fmt.Errorf("koreksi tidak valid: stok akhir DAMAGED tidak boleh kurang dari 0 (stok saat ini: %d, penyesuaian: %d)", damagedStock.Quantity-diffDamaged, diffDamaged)
			}
			if err := repo.UpdateStock(damagedStock); err != nil {
				return err
			}

			// Log mutasi
			logType := "IN"
			logQty := diffDamaged
			if diffDamaged < 0 {
				logType = "OUT"
				logQty = -diffDamaged
			}
			log := &domain.WarehouseLog{
				BranchID:    transfer.ToBranchID,
				ProductID:   transfer.ProductID,
				Type:        logType,
				Source:      "FACTORY_TRANSFER_ADJUSTMENT (DAMAGED)",
				Quantity:    logQty,
				ProcessedBy: correctedBy,
				ApprovedBy:  correctedBy,
			}
			if err := repo.CreateLog(log); err != nil {
				return err
			}
		}

		// 3. Simpan data koreksi ke tabel audit
		correction := &domain.ProductTransferCorrection{
			ProductTransferID:  transfer.ID,
			CorrectedBy:        correctedBy,
			OldActualQuantity:  oldActual,
			NewActualQuantity:  actualQty,
			OldDamagedQuantity: oldDamaged,
			NewDamagedQuantity: damagedQty,
			CorrectionReason:   reason,
		}
		if err := tx.Create(correction).Error; err != nil {
			return err
		}

		// 4. Update data ProductTransfer
		transfer.ActualReceivedQuantity = actualQty
		transfer.DamagedQuantity = damagedQty
		noteEntry := fmt.Sprintf("\n[Koreksi oleh %s]: %s (Bagus: %d, Rusak: %d)", correctedBy, reason, actualQty, damagedQty)
		transfer.Notes += noteEntry

		if err := repo.UpdateTransfer(transfer); err != nil {
			return err
		}

		// 5. Trigger Auto-Fulfillment jika ada tambahan GOOD stock
		if diffGood > 0 {
			return u.internalRunAutoFulfillment(tx, transfer.ToBranchID, transfer.ProductID)
		}

		return nil
	})
}

func (u *warehouseUsecase) GetWarehouseLogs(branchID uuid.UUID) ([]domain.WarehouseLog, error) {
	return u.repo.GetLogsByBranchID(branchID)
}

func (u *warehouseUsecase) GetAllTransfers(branchID uuid.UUID) ([]domain.ProductTransfer, error) {
	return u.repo.GetTransfersByBranch(branchID)
}

func (u *warehouseUsecase) AdjustStock(branchID, productID uuid.UUID, quantity int, reason string, batchNo string, processedBy string, stockType string) error {
	if batchNo == "BACKORDER" {
		return errors.New("batch BACKORDER adalah antrean virtual pesanan dan tidak dapat disesuaikan melalui stok opname fisik")
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewWarehouseRepository(tx)

		if stockType == "" {
			stockType = "GOOD"
		}

		stock, err := repo.GetStock(branchID, productID, stockType, batchNo)
		if err != nil {
			return err
		}

		diff := quantity - stock.Quantity
		stock.Quantity = quantity
		if err := repo.UpdateStock(stock); err != nil {
			return err
		}

		// Check Threshold
		if stock.Quantity < stock.MinLimit {
			// Trigger Notification
			u.triggerLowStockNotification(stock)
		}

		logType := "IN"
		if diff < 0 {
			logType = "OUT"
		}

		absDiff := diff
		if diff < 0 {
			absDiff = -diff
		}

		log := &domain.WarehouseLog{
			BranchID:    branchID,
			ProductID:   productID,
			Type:        logType,
			Source:      "OPNAME: " + reason,
			Quantity:    absDiff,
			ProcessedBy: processedBy,
			ApprovedBy:  processedBy,
		}
		if err := repo.CreateLog(log); err != nil {
			return err
		}

		// If stockType == "GOOD", synchronize with factory stock if this branch belongs to a factory (1:1 sync)
		if stockType == "GOOD" {
			fRepo := repository.NewFactoryRepository(tx)
			var factories []domain.Factory
			if err := tx.Where("branch_id = ?", branchID).Find(&factories).Error; err == nil && len(factories) > 0 {
				bNo := batchNo
				if bNo == "" {
					bNo = "DEFAULT"
				}
				for _, factory := range factories {
					fStock, err := fRepo.GetStock(factory.ID, productID, bNo)
					if err == nil && fStock != nil {
						fStock.Quantity = quantity
						if stock.ExpiryDate != nil {
							fStock.ExpiryDate = stock.ExpiryDate
						}
						_ = fRepo.UpdateStock(fStock)

						if absDiff > 0 {
							fLogSource := "OPNAME_GUDANG"
							if reason != "" {
								fLogSource = "OPNAME_GUDANG: " + reason
								if len(fLogSource) > 50 {
									fLogSource = fLogSource[:50]
								}
							}
							_ = fRepo.CreateInventoryLog(&domain.FactoryInventoryLog{
								FactoryID: factory.ID,
								ProductID: productID,
								Type:      logType,
								Source:    fLogSource,
								BatchNo:   bNo,
								Quantity:  absDiff,
							})
						}
					}
				}
			}
		}

		// Trigger Auto-Fulfillment if adjustment was positive
		if diff > 0 {
			return u.internalRunAutoFulfillment(tx, branchID, productID)
		}
		return nil
	})
}

func (u *warehouseUsecase) SetStockLimit(branchID, productID uuid.UUID, limit int, batchNo string) error {
	if branchID == uuid.Nil {
		// Update semua cabang
		if err := u.db.Model(&domain.WarehouseStock{}).
			Where("product_id = ?", productID).
			Update("min_limit", limit).Error; err != nil {
			return err
		}

		// Pastikan setiap cabang memiliki minimal 1 record stok default dengan limit ini
		var branches []domain.Branch
		if err := u.db.Find(&branches).Error; err != nil {
			return err
		}

		for _, b := range branches {
			var count int64
			u.db.Model(&domain.WarehouseStock{}).
				Where("branch_id = ? AND product_id = ?", b.ID, productID).
				Count(&count)
			if count == 0 {
				_ = u.repo.UpdateStock(&domain.WarehouseStock{
					BranchID:  b.ID,
					ProductID: productID,
					StockType: "GOOD",
					BatchNo:   "DEFAULT",
					Quantity:  0,
					MinLimit:  limit,
				})
			}
		}
		return nil
	}

	res := u.db.Model(&domain.WarehouseStock{}).
		Where("branch_id = ? AND product_id = ?", branchID, productID).
		Update("min_limit", limit)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return u.repo.UpdateStock(&domain.WarehouseStock{
			BranchID:  branchID,
			ProductID: productID,
			StockType: "GOOD",
			BatchNo:   "DEFAULT",
			Quantity:  0,
			MinLimit:  limit,
		})
	}
	return nil
}

func (u *warehouseUsecase) ApproveTransfer(transferID uuid.UUID, approverID *uuid.UUID, approverName, approverPosition string) error {
	transfer, err := u.repo.GetTransferByID(transferID)
	if err != nil {
		return err
	}
	if transfer.Status != "REQUESTED" {
		return errors.New("only REQUESTED transfers can be approved")
	}
	transfer.Status = "APPROVED"
	now := time.Now()
	transfer.ApprovedAt = &now
	if approverID != nil {
		transfer.ApprovedByID = approverID
	}
	if approverName != "" {
		transfer.ApprovedByName = approverName
	}
	if approverPosition != "" {
		transfer.ApprovedByPosition = approverPosition
	}
	err = u.repo.UpdateTransfer(transfer)
	if err == nil && u.sseHub != nil {
		go func(t *domain.ProductTransfer, approver string) {
			prodName := "Produk"
			if t.Product != nil {
				prodName = t.Product.Name
			}
			if approver == "" {
				approver = "Pabrik / Supervisor"
			}
			event := sse.NotificationEvent{
				ID:        t.ID.String() + "-approved",
				Type:      "warehouse_transfer_approved",
				Title:     "✅ Permintaan Pengiriman Disetujui",
				Message:   fmt.Sprintf("Permintaan pengiriman %s (%s, Qty: %d %s) telah disetujui oleh %s", t.DeliveryOrderNo, prodName, t.Quantity, t.Unit, approver),
				CreatedAt: time.Now().Format(time.RFC3339),
				Data: map[string]interface{}{
					"transfer_id": t.ID.String(),
					"do_number":   t.DeliveryOrderNo,
					"status":      "APPROVED",
					"approved_by": approver,
					"product_id":  t.ProductID.String(),
					"branch_id":   t.ToBranchID.String(),
				},
				TargetBranchIDs: []string{t.ToBranchID.String()},
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleAdminGudang),
					string(domain.RoleSupervisorGudang),
					string(domain.RoleManager),
				},
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}(transfer, approverName)
	}
	return err
}

func (u *warehouseUsecase) RejectTransfer(transferID uuid.UUID, reason string) error {
	transfer, err := u.repo.GetTransferByID(transferID)
	if err != nil {
		return err
	}
	transfer.Status = "REJECTED"
	transfer.RejectionReason = reason
	err = u.repo.UpdateTransfer(transfer)
	if err == nil && u.sseHub != nil {
		go func(t *domain.ProductTransfer, r string) {
			prodName := "Produk"
			if t.Product != nil {
				prodName = t.Product.Name
			}
			event := sse.NotificationEvent{
				ID:        t.ID.String() + "-rejected",
				Type:      "warehouse_transfer_rejected",
				Title:     "❌ Permintaan Pengiriman Ditolak",
				Message:   fmt.Sprintf("Permintaan pengiriman %s (%s) ditolak: %s", t.DeliveryOrderNo, prodName, r),
				CreatedAt: time.Now().Format(time.RFC3339),
				Data: map[string]interface{}{
					"transfer_id":      t.ID.String(),
					"do_number":        t.DeliveryOrderNo,
					"status":           "REJECTED",
					"rejection_reason": r,
					"branch_id":        t.ToBranchID.String(),
				},
				TargetBranchIDs: []string{t.ToBranchID.String()},
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleAdminGudang),
					string(domain.RoleSupervisorGudang),
					string(domain.RoleManager),
				},
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}(transfer, reason)
	}
	return err
}

func (u *warehouseUsecase) ConfirmArrival(transferID uuid.UUID) error {
	transfer, err := u.repo.GetTransferByID(transferID)
	if err != nil {
		return err
	}
	if transfer.Status != "SHIPPED" {
		return errors.New("hanya pengiriman berstatus SHIPPED yang bisa dikonfirmasi kedatangannya")
	}
	transfer.Status = "ARRIVED"
	return u.repo.UpdateTransfer(transfer)
}

func (u *warehouseUsecase) ConfirmShipment(transferID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewFactoryRepository(tx)

		// 1. Get Transfer
		transfer, err := repo.GetTransferByID(transferID)
		if err != nil {
			return err
		}
		if transfer.Status != "APPROVED" && transfer.Status != "LOADING" {
			return errors.New("hanya pengiriman berstatus APPROVED atau LOADING yang bisa dikonfirmasi pengirimannya")
		}

		// 2. Get Stock (Convert to total pieces)
		pcsPerUnit := transfer.PcsPerUnit
		if isBaseUnit(transfer.Unit) {
			pcsPerUnit = 1
		}
		totalPieces := transfer.Quantity * pcsPerUnit
		if transfer.BatchNo == "" {
			transfer.BatchNo = "DEFAULT"
		}

		if transfer.FulfillmentBranchID != nil {
			wRepo := repository.NewWarehouseRepository(tx)
			stock, err := wRepo.GetStock(*transfer.FulfillmentBranchID, transfer.ProductID, "GOOD", transfer.BatchNo)
			if err != nil {
				return err
			}
			if stock.Quantity < totalPieces && (transfer.BatchNo == "DEFAULT" || transfer.BatchNo == "") {
				availBatches, errBatches := wRepo.GetAvailableBatches(*transfer.FulfillmentBranchID, transfer.ProductID, "GOOD")
				if errBatches == nil && len(availBatches) > 0 {
					for _, b := range availBatches {
						if b.Quantity >= totalPieces {
							stock = &b
							transfer.BatchNo = b.BatchNo
							break
						}
					}
				}
			}
			if stock.Quantity < totalPieces {
				return fmt.Errorf("stok gudang cabang delegasi (Batch %s) tidak mencukupi (Butuh: %d, Tersedia: %d)", transfer.BatchNo, totalPieces, stock.Quantity)
			}
			stock.Quantity -= totalPieces
			if err := wRepo.UpdateStock(stock); err != nil {
				return err
			}
			// Create BranchPaymentLog for Piutang Transfer (Kompensasi Hutang)
			if transfer.TotalAmount > 0 {
				paymentLog := &domain.BranchPaymentLog{
					ProductTransferID: transfer.ID,
					BranchID:          *transfer.FulfillmentBranchID,
					AmountPaid:        transfer.TotalAmount,
					SourceOfFunds:     "DELEGATED_STOCK_TRANSFER",
					Notes:             fmt.Sprintf("Kompensasi pengalihan piutang untuk DO %s", transfer.DeliveryOrderNo),
				}
				if err := tx.Create(paymentLog).Error; err != nil {
					return err
				}
			}
			transfer.IsStockBorrowed = false

			log := &domain.WarehouseLog{
				BranchID:    *transfer.FulfillmentBranchID,
				ProductID:   transfer.ProductID,
				Type:        "OUT",
				Source:      "DELEGATED FACTORY TRANSFER",
				BatchNo:     transfer.BatchNo,
				Quantity:    totalPieces,
				ProcessedBy: "SISTEM",
			}
			if err := wRepo.CreateLog(log); err != nil {
				return err
			}
		} else {
			stock, err := repo.GetStock(transfer.FromFactoryID, transfer.ProductID, transfer.BatchNo)
			if err != nil {
				return err
			}

			// 3. Reduce Factory Stock
			stock.Quantity -= totalPieces
			if err := repo.UpdateStock(stock); err != nil {
				return err
			}

			inventoryLog := &domain.FactoryInventoryLog{
				FactoryID: transfer.FromFactoryID,
				ProductID: transfer.ProductID,
				Type:      "OUT",
				Source:    "TRANSFER TO BRANCH (CONFIRMED BY WAREHOUSE)",
				BatchNo:   transfer.BatchNo,
				Quantity:  totalPieces,
			}
			if err := repo.CreateInventoryLog(inventoryLog); err != nil {
				return err
			}
		}

		// 4. Update Status to SHIPPED
		now := time.Now()
		transfer.Status = "SHIPPED"
		transfer.ShippedAt = &now

		if err := repo.UpdateTransfer(transfer); err != nil {
			return err
		}

		return nil
	})
}

func (u *warehouseUsecase) triggerLowStockNotification(stock *domain.WarehouseStock) {
	// Re-fetch with product details for message
	fullStock, err := u.repo.GetStock(stock.BranchID, stock.ProductID, "GOOD", stock.BatchNo)
	if err != nil || fullStock.Branch == nil || fullStock.Branch.CompanyID == nil || fullStock.Product == nil {
		return
	}

	notif := &domain.Notification{
		CompanyID: *fullStock.Branch.CompanyID,
		BranchID:  &stock.BranchID,
		Title:     "Stok Rendah!",
		Message:   fmt.Sprintf("Stok produk %s di bawah limit (%d < %d)", fullStock.Product.Name, fullStock.Quantity, fullStock.MinLimit),
		Type:      "LOW_STOCK",
	}
	_ = u.notifRepo.Create(notif)

	if u.sseHub != nil {
		go func() {
			event := sse.NotificationEvent{
				ID:        uuid.New().String(),
				Type:      "low_stock",
				Title:     "Peringatan Stok Menipis",
				Message:   fmt.Sprintf("Stok produk %s di cabang %s menipis (%d < %d)", fullStock.Product.Name, fullStock.Branch.Name, fullStock.Quantity, fullStock.MinLimit),
				CompanyID: fullStock.Branch.CompanyID.String(),
				CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
				Data: map[string]interface{}{
					"product_id":   fullStock.ProductID.String(),
					"product_name": fullStock.Product.Name,
					"branch_name":  fullStock.Branch.Name,
					"quantity":     fullStock.Quantity,
					"min_limit":    fullStock.MinLimit,
				},
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}()
	}
}

func (u *warehouseUsecase) DispatchByInvoice(receiptNo string, branchID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		warehouseRepo := repository.NewWarehouseRepository(tx)
		salesRepo := repository.NewSalesTransactionRepository(tx)

		// 1. Find Transaction
		transaction, err := salesRepo.FindByReceiptNo(receiptNo)
		if err != nil {
			return err
		}
		if transaction == nil {
			return errors.New("nota tidak ditemukan")
		}

		// 2. Validate status
		if transaction.Status == "DELIVERED" {
			return errors.New("transaksi ini sudah selesai dikirim (Status: DELIVERED)")
		}

		// 3. Deduct Stock for each item using available batches (FEFO)
		for _, item := range transaction.Items {
			batches, err := warehouseRepo.GetAvailableBatches(branchID, item.ProductID, "GOOD")
			if err != nil {
				return err
			}

			remainingToDeduct := item.Quantity
			for idx := range batches {
				b := &batches[idx]
				if b.Quantity <= 0 {
					continue
				}

				deduct := b.Quantity
				if deduct > remainingToDeduct {
					deduct = remainingToDeduct
				}

				b.Quantity -= deduct
				if err := warehouseRepo.UpdateStock(b); err != nil {
					return err
				}

				remainingToDeduct -= deduct

				if b.Quantity < b.MinLimit && b.MinLimit > 0 {
					u.triggerLowStockNotification(b)
				}

				// Log movement per batch
				log := &domain.WarehouseLog{
					BranchID:  branchID,
					ProductID: item.ProductID,
					Type:      "OUT",
					Source:    "SALES_DISPATCH: " + receiptNo + " (Batch: " + b.BatchNo + ")",
					Quantity:  deduct,
					BatchNo:   b.BatchNo,
				}
				if err := warehouseRepo.CreateLog(log); err != nil {
					return err
				}

				if remainingToDeduct <= 0 {
					break
				}
			}
		}

		// 4. Update Transaction Status
		transaction.Status = "DELIVERED"
		if err := salesRepo.Update(transaction); err != nil {
			return err
		}

		return nil
	})
}

func (u *warehouseUsecase) DeleteStock(stockID, branchID uuid.UUID, reason, processedBy string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewWarehouseRepository(tx)
		stock, err := repo.GetStockByID(stockID)
		if err != nil {
			return errors.New("data stok tidak ditemukan")
		}

		if branchID != uuid.Nil && stock.BranchID != branchID {
			return errors.New("tidak memiliki akses untuk menghapus stok di cabang ini")
		}

		// Jika masih terdapat sisa kuantitas fisik > 0, catat mutasi keluar log manual delete
		if stock.Quantity > 0 {
			sourceText := "MANUAL_DELETE"
			if reason != "" {
				sourceText = fmt.Sprintf("MANUAL_DELETE: %s", reason)
			}
			log := &domain.WarehouseLog{
				BranchID:    stock.BranchID,
				ProductID:   stock.ProductID,
				Type:        "OUT",
				Source:      sourceText,
				BatchNo:     stock.BatchNo,
				Quantity:    stock.Quantity,
				ProcessedBy: processedBy,
			}
			if err := repo.CreateLog(log); err != nil {
				return err
			}
		}

		return repo.DeleteStock(stock.ID)
	})
}
