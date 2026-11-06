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

type SalesTransferUsecase interface {
	CreateTransfer(branchID uuid.UUID, req CreateSalesTransferRequest) error
	CreateBatchTransfer(branchID uuid.UUID, req CreateBatchSalesTransferRequest) error
	CompleteTransfer(transferID uuid.UUID) error
	CancelTransfer(transferID uuid.UUID) error
	GetTransfersByBranch(branchID uuid.UUID) ([]domain.SalesTransfer, error)
	GetTransfersByCompany(companyID uuid.UUID) ([]domain.SalesTransfer, error)
	GetTransfersByEmployee(employeeID uuid.UUID) ([]domain.SalesTransfer, error)
	GetSalesStock(employeeID uuid.UUID) ([]domain.SalesStock, error)
	DeleteTransfer(id uuid.UUID) error
	ApproveSupervisorTransfer(transferID uuid.UUID, newQty int) error
	RejectSupervisorTransfer(transferID uuid.UUID) error
	ReconcileSalesStock(employeeID uuid.UUID, branchID uuid.UUID, req SalesStockReconciliationRequest, userID uuid.UUID) error
	UpdateTransfer(transferID uuid.UUID, req UpdateSalesTransferRequest) error
	GenerateDOForTransfers(req GenerateDORequest) error
	AdjustSalesStock(employeeID uuid.UUID, items []AdjustSalesStockItem) error
}

type AdjustSalesStockItem struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity"` // new quantity
}

type GenerateDORequest struct {
	SalesTransferIDs []uuid.UUID `json:"sales_transfer_ids" binding:"required"`
}

type UpdateSalesTransferRequest struct {
	Quantity        int    `json:"quantity"`
	ActualQuantity  int    `json:"actual_quantity"`
	DamagedQuantity int    `json:"damaged_quantity"`
	Notes           string `json:"notes"`
	Reason          string `json:"reason"`
	CorrectedBy     string `json:"corrected_by"`
}

type SalesStockReconciliationRequest struct {
	Items []ReconciliationItem `json:"items" binding:"required"`
}

type ReconciliationItem struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	ActualQty int       `json:"actual_qty" binding:"required"`
	Reason    string    `json:"reason"` // SOLD, RETURN, ADJUSTMENT
	Notes     string    `json:"notes"`
}

type salesTransferUsecase struct {
	repo          repository.SalesTransferRepository
	warehouseRepo repository.WarehouseRepository
	db            *gorm.DB
	fcmClient     FcmUsecase
	wdoUsecase    WarehouseDeliveryOrderUsecase
}

func NewSalesTransferUsecase(repo repository.SalesTransferRepository, warehouseRepo repository.WarehouseRepository, db *gorm.DB, fcmClient FcmUsecase, wdoUsecase WarehouseDeliveryOrderUsecase) SalesTransferUsecase {
	return &salesTransferUsecase{repo, warehouseRepo, db, fcmClient, wdoUsecase}
}

type CreateSalesTransferRequest struct {
	EmployeeID            uuid.UUID                  `json:"employee_id"`
	DistributorStoreID    *uuid.UUID                 `json:"distributor_store_id"`
	RequestSource         string                     `json:"request_source"`
	ProductID             *uuid.UUID                 `json:"product_id"`
	BonusInventoryID      *uuid.UUID                 `json:"bonus_inventory_id"`
	BatchNo               string                     `json:"batch_no"`
	Quantity              int                        `json:"quantity" binding:"required"`
	ActualQuantity        int                        `json:"actual_quantity"`
	DamagedQuantity       int                        `json:"damaged_quantity"`
	Unit                  string                     `json:"unit"`
	Type                  domain.SalesTransferType   `json:"type" binding:"required"`
	Notes                 string                     `json:"notes"`
	ReferenceNo           string                     `json:"reference_no"`
	Status                domain.SalesTransferStatus `json:"status"`
	IgnoreStockValidation bool                       `json:"ignore_stock_validation"`
	AutoApprove           bool                       `json:"auto_approve"`
}

type BatchTransferItem struct {
	ProductID        *uuid.UUID `json:"product_id"`
	BonusInventoryID *uuid.UUID `json:"bonus_inventory_id"`
	BatchNo          string     `json:"batch_no"`
	Quantity         int        `json:"quantity" binding:"required,min=1"`
	ActualQuantity   int        `json:"actual_quantity"`
	DamagedQuantity  int        `json:"damaged_quantity"`
	Unit             string     `json:"unit"`
	Notes            string     `json:"notes"`
}

type CreateBatchSalesTransferRequest struct {
	BranchID              uuid.UUID                  `json:"branch_id"`
	EmployeeID            uuid.UUID                  `json:"employee_id" binding:"required"`
	DistributorStoreID    *uuid.UUID                 `json:"distributor_store_id"`
	RequestSource         string                     `json:"request_source"`
	Type                  domain.SalesTransferType   `json:"type" binding:"required"`
	Status                domain.SalesTransferStatus `json:"status"`
	Notes                 string                     `json:"notes"`
	ReferenceNo           string                     `json:"reference_no"`
	IgnoreStockValidation bool                       `json:"ignore_stock_validation"`
	AutoApprove           bool                       `json:"auto_approve"`
	Items                 []BatchTransferItem        `json:"items" binding:"required,min=1"`
}

func (u *salesTransferUsecase) CreateTransfer(branchID uuid.UUID, req CreateSalesTransferRequest) error {
	var transferRecord *domain.SalesTransfer
	err := u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewSalesTransferRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)
		if req.RequestSource == "DISTRIBUTOR" {
			if req.DistributorStoreID == nil || *req.DistributorStoreID == uuid.Nil {
				return errors.New("distributor_store_id wajib diisi")
			}
			var store domain.Store
			if err := tx.Select("id", "customer_type").First(&store, "id = ?", *req.DistributorStoreID).Error; err != nil || store.CustomerType != "DISTRIBUTOR" {
				return errors.New("tujuan request bukan distributor yang valid")
			}
		}

		if branchID == uuid.Nil {
			var emp domain.Employee
			if err := tx.First(&emp, "id = ?", req.EmployeeID).Error; err != nil {
				return errors.New("salesman tidak ditemukan")
			}
			if emp.BranchID == nil {
				return errors.New("salesman tidak memiliki cabang terasosiasi")
			}
			branchID = *emp.BranchID
		}

		// 1. Validate quantity if it's a TRANSFER (Warehouse -> Salesman)
		if req.Type == domain.SalesTransferOut {
			if req.ProductID != nil {
				if req.BatchNo != "" && req.BatchNo != "AUTO" {
					stock, err := wRepo.GetStock(branchID, *req.ProductID, "GOOD", req.BatchNo)
					if err != nil {
						return err
					}
					if stock.Quantity < req.Quantity {
						return fmt.Errorf("stok gudang untuk batch %s tidak mencukupi", req.BatchNo)
					}
				} else {
					// Validate aggregate available stock
					batches, err := wRepo.GetAvailableBatches(branchID, *req.ProductID, "GOOD")
					if err != nil {
						return err
					}
					totalAvailable := 0
					for _, b := range batches {
						totalAvailable += b.Quantity
					}
					if totalAvailable < req.Quantity {
						return errors.New("stok gudang produk tidak mencukupi")
					}
				}
			} else if req.BonusInventoryID != nil {
				bonusRepo := repository.NewBonusInventoryRepository(tx)
				inv, err := bonusRepo.GetByID(tx.Statement.Context, *req.BonusInventoryID)
				if err != nil {
					return errors.New("barang bonus tidak ditemukan")
				}
				if inv.Quantity < req.Quantity {
					return errors.New("stok gudang bonus tidak mencukupi")
				}
			} else {
				return errors.New("product_id atau bonus_inventory_id harus diisi")
			}
		} else if req.Type == domain.SalesTransferIn {
			if !req.IgnoreStockValidation {
				if req.ProductID != nil {
					// 2. Validate quantity if it's a RETURN (Salesman -> Warehouse)
					sStock, err := repo.GetSalesStock(req.EmployeeID, *req.ProductID)
					if err != nil {
						return err
					}

					// Add active delivery cargo stock to validation
					var activeDeliveryCargo int
					var batches []domain.DeliveryBatch
					now := time.Now()
					startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

					err = tx.Preload("Items").
						Preload("Items.SalesOrder").
						Preload("Items.SalesOrder.Items").
						Where("driver_id = ? AND status IN ? AND delivery_date >= ?",
							req.EmployeeID,
							[]domain.DeliveryBatchStatus{domain.DeliveryBatchCompleted, domain.DeliveryBatchWaitingApproval},
							startOfDay).
						Find(&batches).Error
					if err == nil {
						for _, batch := range batches {
							for _, item := range batch.Items {
								isDelivered := item.Status == domain.DeliveryItemDelivered
								if item.SalesOrder != nil {
									for _, soItem := range item.SalesOrder.Items {
										if soItem.ProductID == *req.ProductID {
											ordered := soItem.OrderedQuantity
											actual := soItem.ActualQuantity
											currentOnTruck := 0
											if isDelivered {
												currentOnTruck = ordered - actual
											} else {
												currentOnTruck = ordered
											}
											if currentOnTruck > 0 {
												activeDeliveryCargo += currentOnTruck
											}
										}
									}
								}
							}
						}
					}

					if sStock.Quantity+activeDeliveryCargo < req.Quantity {
						return errors.New("stok sales produk tidak mencukupi untuk retur")
					}
				} else if req.BonusInventoryID != nil {
					salesBonusRepo := repository.NewSalesBonusRepository(tx)
					stock, err := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, req.EmployeeID, *req.BonusInventoryID)
					if err != nil || stock.Quantity < req.Quantity {
						return errors.New("stok sales bonus tidak mencukupi untuk retur")
					}
				} else {
					return errors.New("product_id atau bonus_inventory_id harus diisi")
				}
			}
		}

		status := req.Status
		if status == "" {
			status = domain.SalesTransferPending
		}

		transfer := &domain.SalesTransfer{
			BranchID:           branchID,
			EmployeeID:         req.EmployeeID,
			DistributorStoreID: req.DistributorStoreID,
			RequestSource:      req.RequestSource,
			ProductID:          req.ProductID,
			BonusInventoryID:   req.BonusInventoryID,
			BatchNo:            req.BatchNo,
			Quantity:           req.Quantity,
			ActualQuantity:     req.ActualQuantity,
			DamagedQuantity:    req.DamagedQuantity,
			Unit:               req.Unit,
			Type:               req.Type,
			Status:             status,
			ReferenceNo:        req.ReferenceNo,
			Notes:              req.Notes,
			TransferDate:       time.Now(),
		}

		if transfer.ReferenceNo == "" {
			prefix := "TRF"
			if transfer.Type == domain.SalesTransferIn {
				prefix = "RTN"
			}
			transfer.ReferenceNo = prefix + "-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:4]
		}

		transferRecord = transfer
		if err := repo.CreateTransfer(transfer); err != nil {
			return err
		}

		return nil
	})

	if err == nil && req.AutoApprove && transferRecord != nil {
		if errApprove := u.CompleteTransfer(transferRecord.ID); errApprove != nil {
			return fmt.Errorf("transfer berhasil dibuat tetapi gagal auto-approve: %w", errApprove)
		}
	}

	if err == nil && transferRecord != nil && transferRecord.Status == domain.SalesTransferWaitingSupervisor && u.fcmClient != nil {
		var supervisors []domain.Employee
		errSpv := u.db.Where("branch_id = ? AND job_position_id IN (SELECT id FROM job_positions WHERE title ILIKE '%Supervisor%')", branchID).Find(&supervisors).Error
		if errSpv == nil {
			for _, spv := range supervisors {
				if spv.UserID != uuid.Nil {
					_ = u.fcmClient.SendPushNotification(spv.UserID, "Permintaan Barang Bawaan", "Terdapat permintaan barang bawaan baru dari sales yang membutuhkan persetujuan Anda.")
				}
			}
		}
	}

	return err
}

func (u *salesTransferUsecase) CreateBatchTransfer(branchID uuid.UUID, req CreateBatchSalesTransferRequest) error {
	if len(req.Items) == 0 {
		return errors.New("daftar barang transfer tidak boleh kosong")
	}

	var hasWaitingSupervisor bool
	var createdTransfers []*domain.SalesTransfer
	err := u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewSalesTransferRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		if req.RequestSource == "DISTRIBUTOR" {
			if req.DistributorStoreID == nil || *req.DistributorStoreID == uuid.Nil {
				return errors.New("distributor_store_id wajib diisi")
			}
			var store domain.Store
			if err := tx.Select("id", "customer_type").First(&store, "id = ?", *req.DistributorStoreID).Error; err != nil || store.CustomerType != "DISTRIBUTOR" {
				return errors.New("tujuan request bukan distributor yang valid")
			}
		}

		if branchID == uuid.Nil {
			var emp domain.Employee
			if err := tx.First(&emp, "id = ?", req.EmployeeID).Error; err != nil {
				return errors.New("salesman tidak ditemukan")
			}
			if emp.BranchID == nil {
				return errors.New("salesman tidak memiliki cabang terasosiasi")
			}
			branchID = *emp.BranchID
		}

		status := req.Status
		if status == "" {
			status = domain.SalesTransferPending
		}
		if status == domain.SalesTransferWaitingSupervisor {
			hasWaitingSupervisor = true
		}

		now := time.Now()
		batchRef := req.ReferenceNo
		if batchRef == "" {
			prefix := "TRF"
			if req.Type == domain.SalesTransferIn {
				prefix = "RTN"
			}
			batchRef = prefix + "-" + now.Format("20060102") + "-" + uuid.New().String()[:4]
		}

		for idx, item := range req.Items {
			if item.Quantity <= 0 {
				return errors.New("jumlah barang harus lebih dari 0")
			}

			// Validate Outbound stock
			if req.Type == domain.SalesTransferOut {
				if item.ProductID != nil {
					if item.BatchNo != "" && item.BatchNo != "AUTO" {
						stock, err := wRepo.GetStock(branchID, *item.ProductID, "GOOD", item.BatchNo)
						if err != nil {
							return err
						}
						if stock.Quantity < item.Quantity {
							return fmt.Errorf("stok gudang untuk batch %s tidak mencukupi (item ke-%d)", item.BatchNo, idx+1)
						}
					} else {
						batches, err := wRepo.GetAvailableBatches(branchID, *item.ProductID, "GOOD")
						if err != nil {
							return err
						}
						totalAvailable := 0
						for _, b := range batches {
							totalAvailable += b.Quantity
						}
						if totalAvailable < item.Quantity {
							return fmt.Errorf("stok gudang tidak mencukupi untuk item ke-%d", idx+1)
						}
					}
				} else if item.BonusInventoryID != nil {
					bonusRepo := repository.NewBonusInventoryRepository(tx)
					inv, err := bonusRepo.GetByID(tx.Statement.Context, *item.BonusInventoryID)
					if err != nil {
						return errors.New("barang bonus tidak ditemukan")
					}
					if inv.Quantity < item.Quantity {
						return fmt.Errorf("stok gudang bonus tidak mencukupi untuk item ke-%d", idx+1)
					}
				} else {
					return errors.New("product_id atau bonus_inventory_id harus diisi untuk setiap item")
				}
			} else if req.Type == domain.SalesTransferIn {
				if !req.IgnoreStockValidation {
					if item.ProductID != nil {
						sStock, err := repo.GetSalesStock(req.EmployeeID, *item.ProductID)
						if err != nil {
							return err
						}
						if sStock.Quantity < item.Quantity {
							return fmt.Errorf("stok sales tidak mencukupi untuk retur item ke-%d", idx+1)
						}
					} else if item.BonusInventoryID != nil {
						salesBonusRepo := repository.NewSalesBonusRepository(tx)
						stock, err := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, req.EmployeeID, *item.BonusInventoryID)
						if err != nil || stock.Quantity < item.Quantity {
							return fmt.Errorf("stok sales bonus tidak mencukupi untuk retur item ke-%d", idx+1)
						}
					} else {
						return errors.New("product_id atau bonus_inventory_id harus diisi untuk setiap item")
					}
				}
			}

			itemNotes := item.Notes
			if itemNotes == "" {
				itemNotes = req.Notes
			}

			transfer := &domain.SalesTransfer{
				BranchID:           branchID,
				EmployeeID:         req.EmployeeID,
				DistributorStoreID: req.DistributorStoreID,
				RequestSource:      req.RequestSource,
				ProductID:          item.ProductID,
				BonusInventoryID:   item.BonusInventoryID,
				BatchNo:            item.BatchNo,
				Quantity:           item.Quantity,
				ActualQuantity:     item.ActualQuantity,
				DamagedQuantity:    item.DamagedQuantity,
				Unit:               item.Unit,
				Type:               req.Type,
				Status:             status,
				ReferenceNo:        batchRef,
				Notes:              itemNotes,
				TransferDate:       now,
			}

			if err := repo.CreateTransfer(transfer); err != nil {
				return err
			}

			createdTransfers = append(createdTransfers, transfer)
		}

		return nil
	})

	if err == nil && req.AutoApprove && len(createdTransfers) > 0 {
		for _, t := range createdTransfers {
			if errApprove := u.CompleteTransfer(t.ID); errApprove != nil {
				return fmt.Errorf("transfer berhasil dibuat tetapi gagal auto-approve item %s: %w", t.ReferenceNo, errApprove)
			}
		}
	}

	if err == nil && hasWaitingSupervisor && u.fcmClient != nil {
		var supervisors []domain.Employee
		errSpv := u.db.Where("branch_id = ? AND job_position_id IN (SELECT id FROM job_positions WHERE title ILIKE '%Supervisor%')", branchID).Find(&supervisors).Error
		if errSpv == nil {
			for _, spv := range supervisors {
				if spv.UserID != uuid.Nil {
					_ = u.fcmClient.SendPushNotification(spv.UserID, "Permintaan Barang Bawaan", fmt.Sprintf("Terdapat permintaan %d jenis barang bawaan baru dari sales yang membutuhkan persetujuan Anda.", len(req.Items)))
				}
			}
		}
	}

	return err
}

func (u *salesTransferUsecase) CompleteTransfer(transferID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewSalesTransferRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		transfer, err := repo.GetTransferByID(transferID)
		if err != nil {
			return err
		}

		if transfer.Status != domain.SalesTransferPending {
			return errors.New("transfer is not in pending status")
		}

		if transfer.Type == domain.SalesTransferOut {
			// Warehouse -> Salesman
			if transfer.ProductID != nil {
				if transfer.BatchNo != "" && transfer.BatchNo != "AUTO" {
					stock, err := wRepo.GetStock(transfer.BranchID, *transfer.ProductID, "GOOD", transfer.BatchNo)
					if err != nil {
						return err
					}
					if stock.Quantity < transfer.Quantity {
						return fmt.Errorf("stok gudang untuk batch %s tidak mencukupi", transfer.BatchNo)
					}
					stock.Quantity -= transfer.Quantity
					if err := wRepo.UpdateStock(stock); err != nil {
						return err
					}
				} else {
					// FEFO for product
					batches, err := wRepo.GetAvailableBatches(transfer.BranchID, *transfer.ProductID, "GOOD")
					if err != nil {
						return err
					}

					totalAvailable := 0
					for _, b := range batches {
						totalAvailable += b.Quantity
					}

					if totalAvailable < transfer.Quantity {
						return errors.New("stok gudang tidak mencukupi")
					}

					remainingToDeduct := transfer.Quantity
					for i := range batches {
						b := &batches[i]
						availableInBatch := b.Quantity
						if availableInBatch <= 0 {
							continue
						}

						deduct := availableInBatch
						if deduct > remainingToDeduct {
							deduct = remainingToDeduct
						}

						b.Quantity -= deduct
						remainingToDeduct -= deduct
						if err := wRepo.UpdateStock(b); err != nil {
							return err
						}

						if remainingToDeduct <= 0 {
							break
						}
					}
				}

				// Increase Salesman (Aggregate)
				sStock, err := repo.GetSalesStock(transfer.EmployeeID, *transfer.ProductID)
				if err != nil {
					return err
				}
				sStock.Quantity += transfer.Quantity
				if err := repo.UpdateSalesStock(sStock); err != nil {
					return err
				}

				// Log Warehouse OUT
				if err := wRepo.CreateLog(&domain.WarehouseLog{
					BranchID:  transfer.BranchID,
					ProductID: *transfer.ProductID,
					Type:      "OUT",
					Source:    "SALES_TRANSFER",
					Quantity:  transfer.Quantity,
				}); err != nil {
					return err
				}
			} else if transfer.BonusInventoryID != nil {
				bonusRepo := repository.NewBonusInventoryRepository(tx)
				salesBonusRepo := repository.NewSalesBonusRepository(tx)

				inv, err := bonusRepo.GetByID(tx.Statement.Context, *transfer.BonusInventoryID)
				if err != nil {
					return err
				}

				if inv.Quantity < transfer.Quantity {
					return errors.New("insufficient bonus inventory stock in warehouse")
				}

				inv.Quantity -= transfer.Quantity
				if err := bonusRepo.Update(tx.Statement.Context, inv); err != nil {
					return err
				}

				log := &domain.BonusInventoryLog{
					BonusInventoryID: inv.ID,
					Type:             domain.BonusInventoryOut,
					Quantity:         transfer.Quantity,
					Notes:            fmt.Sprintf("Transfer to Sales %s", transfer.EmployeeID),
				}
				if err := bonusRepo.CreateLog(tx.Statement.Context, log); err != nil {
					return err
				}

				stock, err := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, transfer.EmployeeID, *transfer.BonusInventoryID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						stock = &domain.SalesBonusStock{
							EmployeeID:       transfer.EmployeeID,
							BonusInventoryID: *transfer.BonusInventoryID,
							Quantity:         transfer.Quantity,
							UpdatedAt:        time.Now(),
						}
						if err := salesBonusRepo.CreateSalesBonusStock(tx.Statement.Context, stock); err != nil {
							return err
						}
					} else {
						return err
					}
				} else {
					stock.Quantity += transfer.Quantity
					stock.UpdatedAt = time.Now()
					if err := salesBonusRepo.UpdateSalesBonusStock(tx.Statement.Context, stock); err != nil {
						return err
					}
				}
			}
		} else {
			// Salesman -> Warehouse (RETURN)
			if transfer.ProductID != nil {
				// Decrease Salesman
				sStock, err := repo.GetSalesStock(transfer.EmployeeID, *transfer.ProductID)
				if err != nil {
					return err
				}
				sStock.Quantity -= transfer.Quantity
				if sStock.Quantity < 0 {
					sStock.Quantity = 0
				}
				if err := repo.UpdateSalesStock(sStock); err != nil {
					return err
				}

				actualQty := transfer.ActualQuantity
				damagedQty := transfer.DamagedQuantity
				if actualQty == 0 && damagedQty == 0 && transfer.Quantity > 0 {
					actualQty = transfer.Quantity
				}

				returnBatch := transfer.BatchNo
				if returnBatch == "" || returnBatch == "AUTO" {
					returnBatch = "DEFAULT"
				}

				// Increase Warehouse GOOD stock
				if actualQty > 0 {
					wStock, err := wRepo.GetStock(transfer.BranchID, *transfer.ProductID, "GOOD", returnBatch)
					if err != nil {
						return err
					}
					wStock.Quantity += actualQty
					if err := wRepo.UpdateStock(wStock); err != nil {
						return err
					}

					// Log Warehouse GOOD IN
					if err := wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  transfer.BranchID,
						ProductID: *transfer.ProductID,
						Type:      "IN",
						Source:    "SALES_RETURN",
						Quantity:  actualQty,
					}); err != nil {
						return err
					}
				}

				// Increase Warehouse DAMAGED stock
				if damagedQty > 0 {
					dStock, err := wRepo.GetStock(transfer.BranchID, *transfer.ProductID, "DAMAGED", returnBatch)
					if err != nil {
						return err
					}
					dStock.Quantity += damagedQty
					if err := wRepo.UpdateStock(dStock); err != nil {
						return err
					}

					// Log Warehouse DAMAGED IN
					if err := wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  transfer.BranchID,
						ProductID: *transfer.ProductID,
						Type:      "IN",
						Source:    "SALES_RETURN (DAMAGED)",
						Quantity:  damagedQty,
					}); err != nil {
						return err
					}
				}
			} else if transfer.BonusInventoryID != nil {
				salesBonusRepo := repository.NewSalesBonusRepository(tx)
				bonusRepo := repository.NewBonusInventoryRepository(tx)

				sStock, err := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, transfer.EmployeeID, *transfer.BonusInventoryID)
				if err == nil && sStock != nil {
					sStock.Quantity -= transfer.Quantity
					if sStock.Quantity < 0 {
						sStock.Quantity = 0
					}
					_ = salesBonusRepo.UpdateSalesBonusStock(tx.Statement.Context, sStock)
				}

				inv, err := bonusRepo.GetByID(tx.Statement.Context, *transfer.BonusInventoryID)
				if err == nil && inv != nil {
					inv.Quantity += transfer.Quantity
					_ = bonusRepo.Update(tx.Statement.Context, inv)

					log := &domain.BonusInventoryLog{
						BonusInventoryID: inv.ID,
						Type:             domain.BonusInventoryIn,
						Quantity:         transfer.Quantity,
						Notes:            fmt.Sprintf("Return from Sales %s", transfer.EmployeeID),
					}
					_ = bonusRepo.CreateLog(tx.Statement.Context, log)
				}
			}
		}

		transfer.Status = domain.SalesTransferCompleted
		return repo.UpdateTransfer(transfer)
	})
}

func (u *salesTransferUsecase) CancelTransfer(transferID uuid.UUID) error {
	repo := u.repo
	transfer, err := repo.GetTransferByID(transferID)
	if err != nil {
		return err
	}
	if transfer.Status != domain.SalesTransferPending {
		return errors.New("only pending transfers can be cancelled")
	}
	transfer.Status = domain.SalesTransferCancelled
	return repo.UpdateTransfer(transfer)
}

func (u *salesTransferUsecase) GetTransfersByBranch(branchID uuid.UUID) ([]domain.SalesTransfer, error) {
	transfers, err := u.repo.GetTransfersByBranch(branchID)
	if err != nil {
		return nil, err
	}

	for i := range transfers {
		if transfers[i].Product != nil {
			var stockQty int
			u.db.Model(&domain.WarehouseStock{}).
				Where("product_id = ? AND branch_id = ? AND stock_type = 'GOOD'", transfers[i].ProductID, transfers[i].BranchID).
				Select("COALESCE(SUM(quantity), 0)").
				Scan(&stockQty)
			transfers[i].Product.WarehouseStock = stockQty
		}
	}

	return transfers, nil
}

func (u *salesTransferUsecase) GetTransfersByCompany(companyID uuid.UUID) ([]domain.SalesTransfer, error) {
	transfers, err := u.repo.GetTransfersByCompany(companyID)
	if err != nil {
		return nil, err
	}

	for i := range transfers {
		if transfers[i].Product != nil {
			var stockQty int
			u.db.Model(&domain.WarehouseStock{}).
				Where("product_id = ? AND stock_type = 'GOOD'", transfers[i].ProductID).
				Select("COALESCE(SUM(quantity), 0)").
				Scan(&stockQty)
			transfers[i].Product.WarehouseStock = stockQty
		}
	}

	return transfers, nil
}

func (u *salesTransferUsecase) GetTransfersByEmployee(employeeID uuid.UUID) ([]domain.SalesTransfer, error) {
	transfers, err := u.repo.GetTransfersByEmployee(employeeID)
	if err != nil {
		return nil, err
	}

	for i := range transfers {
		if transfers[i].Product != nil {
			var stockQty int
			u.db.Model(&domain.WarehouseStock{}).
				Where("product_id = ? AND branch_id = ? AND stock_type = 'GOOD'", transfers[i].ProductID, transfers[i].BranchID).
				Select("COALESCE(SUM(quantity), 0)").
				Scan(&stockQty)
			transfers[i].Product.WarehouseStock = stockQty
		} else if transfers[i].BonusInventoryID != nil {
			if transfers[i].BonusInventory == nil {
				var bInv domain.BonusInventory
				if err := u.db.First(&bInv, "id = ?", transfers[i].BonusInventoryID).Error; err == nil {
					transfers[i].BonusInventory = &bInv
				}
			}
			if transfers[i].BonusInventory != nil {
				bID := *transfers[i].BonusInventoryID
				transfers[i].ProductID = &bID
				transfers[i].Product = &domain.Product{
					ID:   bID,
					Name: transfers[i].BonusInventory.Name + " (BONUS)",
					SKU:  "BONUS",
				}
			}
		}
	}

	// Fetch SalesBonusRequest and map to SalesTransfer
	var bonusRequests []domain.SalesBonusRequest
	if err := u.db.Preload("BonusInventory").Where("employee_id = ?", employeeID).Order("created_at DESC").Find(&bonusRequests).Error; err == nil {
		for _, br := range bonusRequests {
			status := domain.SalesTransferPending
			if br.Status == domain.SalesBonusRequestApproved {
				status = domain.SalesTransferCompleted
			} else if br.Status == domain.SalesBonusRequestRejected {
				status = domain.SalesTransferCancelled
			}

			bID := br.BonusInventoryID
			product := &domain.Product{
				ID:   bID,
				Name: "Bonus",
				SKU:  "BONUS",
			}
			if br.BonusInventory != nil {
				product.Name = br.BonusInventory.Name + " (BONUS)"
			}

			transfers = append(transfers, domain.SalesTransfer{
				ID:               br.ID,
				EmployeeID:       br.EmployeeID,
				ProductID:        &bID,
				BonusInventoryID: &bID,
				Quantity:         br.Quantity,
				Type:             domain.SalesTransferOut,
				Status:           status,
				ReferenceNo:      "BONUS-REQ",
				Notes:            br.Notes,
				TransferDate:     br.CreatedAt,
				CreatedAt:        br.CreatedAt,
				UpdatedAt:        br.UpdatedAt,
				Product:          product,
			})
		}
	}

	return transfers, nil
}

func (u *salesTransferUsecase) getCargoStockMap(employeeID uuid.UUID) map[uuid.UUID]*domain.SalesStock {
	var batches []domain.DeliveryBatch
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := u.db.Preload("Items.SalesOrder.Items.Product").
		Where("driver_id = ? AND (status NOT IN ? OR finished_at >= ?)",
			employeeID,
			[]domain.DeliveryBatchStatus{domain.DeliveryBatchCompleted, domain.DeliveryBatchWaitingApproval},
			startOfDay).
		Find(&batches).Error

	cargoMap := make(map[uuid.UUID]*domain.SalesStock)
	if err == nil {
		for _, batch := range batches {
			for _, item := range batch.Items {
				isDelivered := item.Status == domain.DeliveryItemDelivered
				if item.SalesOrder != nil {
					for _, soItem := range item.SalesOrder.Items {
						if soItem.ProductID == uuid.Nil {
							continue
						}
						ordered := soItem.OrderedQuantity
						actual := soItem.ActualQuantity
						currentOnTruck := 0
						if isDelivered {
							currentOnTruck = ordered - actual
						} else {
							currentOnTruck = ordered
						}

						if currentOnTruck > 0 {
							if existing, ok := cargoMap[soItem.ProductID]; ok {
								existing.Quantity += currentOnTruck
							} else {
								var productCopy *domain.Product
								if soItem.Product != nil {
									productCopy = soItem.Product
								}
								cargoMap[soItem.ProductID] = &domain.SalesStock{
									EmployeeID: employeeID,
									ProductID:  soItem.ProductID,
									Quantity:   currentOnTruck,
									Product:    productCopy,
								}
							}
						}
					}
				}
			}
		}
	}
	return cargoMap
}

func (u *salesTransferUsecase) GetSalesStock(employeeID uuid.UUID) ([]domain.SalesStock, error) {
	stocks, err := u.repo.GetStocksByEmployeeID(employeeID)
	if err != nil {
		return nil, err
	}

	cargoMap := u.getCargoStockMap(employeeID)

	// Merge cargo stock into database-loaded sales stock
	for i := range stocks {
		pid := stocks[i].ProductID
		if cargoStock, ok := cargoMap[pid]; ok {
			stocks[i].Quantity += cargoStock.Quantity
			delete(cargoMap, pid)
		}
	}

	// Append any cargo items not currently in salesman's database stock
	for _, cargoStock := range cargoMap {
		stocks = append(stocks, *cargoStock)
	}

	// Fetch SalesBonusStock and map to SalesStock for mobile compatibility
	var bonusStocks []domain.SalesBonusStock
	if err := u.db.Preload("BonusInventory").Where("employee_id = ? AND quantity > 0", employeeID).Find(&bonusStocks).Error; err == nil {
		for _, bs := range bonusStocks {
			product := &domain.Product{
				ID:   bs.BonusInventoryID,
				Name: "Bonus",
				SKU:  "BONUS",
			}
			if bs.BonusInventory != nil {
				product.Name = bs.BonusInventory.Name + " (BONUS)"
				product.Unit = bs.BonusInventory.Unit
			}
			stocks = append(stocks, domain.SalesStock{
				ID:         bs.ID,
				EmployeeID: bs.EmployeeID,
				ProductID:  bs.BonusInventoryID,
				Quantity:   bs.Quantity,
				UpdatedAt:  bs.UpdatedAt,
				Product:    product,
			})
		}
	}

	// Fetch employee to get branch_id
	var employee domain.Employee
	if err := u.db.First(&employee, "id = ?", employeeID).Error; err == nil && employee.BranchID != nil {
		for i := range stocks {
			if stocks[i].Product != nil && stocks[i].Product.SKU != "BONUS" {
				var stockQty int
				u.db.Model(&domain.WarehouseStock{}).
					Where("product_id = ? AND branch_id = ? AND stock_type = 'GOOD'", stocks[i].ProductID, *employee.BranchID).
					Select("COALESCE(SUM(quantity), 0)").
					Scan(&stockQty)
				stocks[i].Product.WarehouseStock = stockQty
			}
		}
	} else {
		// Fallback: sum across all branches if branch is nil
		for i := range stocks {
			if stocks[i].Product != nil && stocks[i].Product.SKU != "BONUS" {
				var stockQty int
				u.db.Model(&domain.WarehouseStock{}).
					Where("product_id = ? AND stock_type = 'GOOD'", stocks[i].ProductID).
					Select("COALESCE(SUM(quantity), 0)").
					Scan(&stockQty)
				stocks[i].Product.WarehouseStock = stockQty
			}
		}
	}

	var finalStocks []domain.SalesStock
	for _, s := range stocks {
		if s.Quantity > 0 {
			finalStocks = append(finalStocks, s)
		}
	}

	return finalStocks, nil
}

func (u *salesTransferUsecase) DeleteTransfer(id uuid.UUID) error {
	return u.repo.DeleteTransfer(id)
}

func (u *salesTransferUsecase) ApproveSupervisorTransfer(transferID uuid.UUID, newQty int) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewSalesTransferRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		transfer, err := repo.GetTransferByID(transferID)
		if err != nil {
			return err
		}

		if transfer.Status != domain.SalesTransferWaitingSupervisor {
			return errors.New("permintaan transfer tidak dalam status menunggu persetujuan supervisor")
		}

		if newQty > 0 {
			transfer.Quantity = newQty
		}

		// Validasi ulang stok gudang jika bertipe TRANSFER (Outbound ke Salesman)
		if transfer.Type == domain.SalesTransferOut {
			batches, err := wRepo.GetAvailableBatches(transfer.BranchID, *transfer.ProductID, "GOOD")
			if err != nil {
				return err
			}
			totalAvailable := 0
			for _, b := range batches {
				totalAvailable += b.Quantity
			}
			if totalAvailable < transfer.Quantity {
				return errors.New("stok gudang tidak mencukupi untuk jumlah barang yang disetujui")
			}
		}

		transfer.Status = domain.SalesTransferPending
		return repo.UpdateTransfer(transfer)
	})
}

func (u *salesTransferUsecase) RejectSupervisorTransfer(transferID uuid.UUID) error {
	repo := u.repo
	transfer, err := repo.GetTransferByID(transferID)
	if err != nil {
		return err
	}
	if transfer.Status != domain.SalesTransferWaitingSupervisor {
		return errors.New("hanya permintaan transfer berstatus menunggu persetujuan supervisor yang bisa ditolak")
	}
	transfer.Status = domain.SalesTransferCancelled
	return repo.UpdateTransfer(transfer)
}

func (u *salesTransferUsecase) ReconcileSalesStock(employeeID uuid.UUID, branchID uuid.UUID, req SalesStockReconciliationRequest, userID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		salesTransferRepo := repository.NewSalesTransferRepository(tx)

		for _, item := range req.Items {
			// Find salesman's current stock
			sStock, err := salesTransferRepo.GetSalesStock(employeeID, item.ProductID)
			currentQty := 0
			if err == nil && sStock != nil {
				currentQty = sStock.Quantity
			}

			diff := currentQty - item.ActualQty
			if diff == 0 {
				continue
			}

			if diff > 0 {
				// Stock is reduced (salesman has less stock than recorded)
				if item.Reason == "SOLD" {
					// 1. Create a canvassing sale transaction
					salesRepo := repository.NewSalesTransactionRepository(tx)
					invoiceCount, _ := salesRepo.CountByCompanyAndDate(uuid.Nil, now)

					var emp domain.Employee
					if errDb := tx.Preload("Company").First(&emp, "id = ?", employeeID).Error; errDb != nil {
						return errDb
					}

					var companyID uuid.UUID
					if emp.CompanyID != nil {
						companyID = *emp.CompanyID
					}
					invoiceNo := fmt.Sprintf("INV-REC-%s-%04d", now.Format("20060102"), invoiceCount+1)

					var prod domain.Product
					if errDb := tx.First(&prod, "id = ?", item.ProductID).Error; errDb != nil {
						return errDb
					}

					notes := fmt.Sprintf("Rekonsiliasi Stok Sales (Otomatis): %s", item.Notes)

					salesTrx := &domain.SalesTransaction{
						CompanyID:       companyID,
						EmployeeID:      employeeID,
						ReceiptNo:       invoiceNo,
						TotalAmount:     float64(diff) * prod.SellingPrice,
						TransactionDate: now,
						PeriodMonth:     int(now.Month()),
						PeriodYear:      now.Year(),
						Status:          "VERIFIED",
						PaymentStatus:   string(domain.PaymentStatusPaid),
						PaymentMethod:   "CASH",
						Notes:           &notes,
						Items: []domain.SalesItem{
							{
								ProductID:          item.ProductID,
								Quantity:           diff,
								PriceAtTransaction: prod.SellingPrice,
								Subtotal:           float64(diff) * prod.SellingPrice,
							},
						},
					}

					if errDb := salesRepo.Create(salesTrx); errDb != nil {
						return errDb
					}
				} else if item.Reason == "RETURN" {
					// 2. Create and complete a SalesTransfer RETURN
					transfer := &domain.SalesTransfer{
						BranchID:     branchID,
						EmployeeID:   employeeID,
						ProductID:    &item.ProductID,
						Quantity:     diff,
						Unit:         "pcs",
						Type:         domain.SalesTransferIn,
						Status:       domain.SalesTransferCompleted,
						ReferenceNo:  fmt.Sprintf("RTN-REC-%s-%s", now.Format("20060102"), uuid.New().String()[:4]),
						Notes:        fmt.Sprintf("Rekonsiliasi Retur Sales: %s", item.Notes),
						TransferDate: now,
					}
					if errDb := tx.Create(transfer).Error; errDb != nil {
						return errDb
					}

					// Return stock to warehouse GOOD stock
					wRepo := repository.NewWarehouseRepository(tx)
					wStock, errDb := wRepo.GetStock(branchID, item.ProductID, "GOOD", "DEFAULT")
					if errDb == nil && wStock != nil {
						wStock.Quantity += diff
						_ = wRepo.UpdateStock(wStock)
					} else {
						wStock = &domain.WarehouseStock{
							BranchID:  branchID,
							ProductID: item.ProductID,
							StockType: "GOOD",
							BatchNo:   "DEFAULT",
							Quantity:  diff,
						}
						_ = tx.Create(wStock)
					}

					wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  branchID,
						ProductID: item.ProductID,
						Type:      "IN",
						Source:    fmt.Sprintf("RETUR REKONSILISASI SALES: %s", transfer.ReferenceNo),
						Quantity:  diff,
					})
				}
			}

			// Update salesman's stock to the actual reported quantity
			if sStock != nil {
				sStock.Quantity = item.ActualQty
				sStock.UpdatedAt = now
				if errDb := salesTransferRepo.UpdateSalesStock(sStock); errDb != nil {
					return errDb
				}
			} else {
				sStock = &domain.SalesStock{
					EmployeeID: employeeID,
					ProductID:  item.ProductID,
					Quantity:   item.ActualQty,
					UpdatedAt:  now,
				}
				if errDb := tx.Create(sStock).Error; errDb != nil {
					return errDb
				}
			}
		}
		return nil
	})
}

func (u *salesTransferUsecase) UpdateTransfer(transferID uuid.UUID, req UpdateSalesTransferRequest) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewSalesTransferRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		transfer, err := repo.GetTransferByID(transferID)
		if err != nil {
			return err
		}

		if transfer.Status == domain.SalesTransferPending || transfer.Status == domain.SalesTransferWaitingSupervisor {
			// Just update quantities and notes directly since no stock has been moved
			transfer.Quantity = req.Quantity
			if req.ActualQuantity > 0 || req.DamagedQuantity > 0 {
				transfer.ActualQuantity = req.ActualQuantity
				transfer.DamagedQuantity = req.DamagedQuantity
			} else {
				transfer.ActualQuantity = req.Quantity
				transfer.DamagedQuantity = 0
			}
			transfer.Notes = req.Notes
			return repo.UpdateTransfer(transfer)
		}

		if transfer.Status != domain.SalesTransferCompleted {
			return errors.New("hanya transfer dengan status PENDING, WAITING_SUPERVISOR, atau COMPLETED yang dapat diedit")
		}

		// Completed transfer correction
		oldQuantity := transfer.Quantity
		oldActual := transfer.ActualQuantity
		oldDamaged := transfer.DamagedQuantity

		newQuantity := req.Quantity
		newActual := req.ActualQuantity
		newDamaged := req.DamagedQuantity

		if newQuantity == 0 {
			newQuantity = oldQuantity
		}
		if newActual == 0 && newDamaged == 0 {
			newActual = newQuantity
		}

		diffTotal := newQuantity - oldQuantity
		diffGood := newActual - oldActual
		diffDamaged := newDamaged - oldDamaged

		correctedBy := req.CorrectedBy
		if correctedBy == "" {
			correctedBy = "Admin Gudang"
		}
		reason := req.Reason
		if reason == "" {
			reason = "Koreksi data transfer/retur"
		}

		if transfer.Type == domain.SalesTransferOut {
			// Warehouse -> Salesman (TRANSFER)
			if diffTotal != 0 {
				if transfer.ProductID != nil {
					if diffTotal > 0 {
						batches, err := wRepo.GetAvailableBatches(transfer.BranchID, *transfer.ProductID, "GOOD")
						if err != nil {
							return err
						}
						totalAvailable := 0
						for _, b := range batches {
							totalAvailable += b.Quantity
						}
						if totalAvailable < diffTotal {
							return errors.New("stok gudang tidak mencukupi untuk penyesuaian transfer")
						}

						remaining := diffTotal
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
							if remaining <= 0 {
								break
							}
						}
					} else {
						wStock, err := wRepo.GetStock(transfer.BranchID, *transfer.ProductID, "GOOD", "DEFAULT")
						if err != nil {
							return err
						}
						wStock.Quantity += (-diffTotal)
						if err := wRepo.UpdateStock(wStock); err != nil {
							return err
						}
					}

					logType := "OUT"
					logQty := diffTotal
					if diffTotal < 0 {
						logType = "IN"
						logQty = -diffTotal
					}
					if err := wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  transfer.BranchID,
						ProductID: *transfer.ProductID,
						Type:      logType,
						Source:    "SALES_TRANSFER_CORRECTION",
						Quantity:  logQty,
					}); err != nil {
						return err
					}

					sStock, err := repo.GetSalesStock(transfer.EmployeeID, *transfer.ProductID)
					if err != nil {
						return err
					}
					sStock.Quantity += diffTotal
					if sStock.Quantity < 0 {
						return errors.New("koreksi tidak valid: stok sales tidak boleh kurang dari 0")
					}
					if err := repo.UpdateSalesStock(sStock); err != nil {
						return err
					}
				} else if transfer.BonusInventoryID != nil {
					bonusRepo := repository.NewBonusInventoryRepository(tx)
					salesBonusRepo := repository.NewSalesBonusRepository(tx)

					inv, err := bonusRepo.GetByID(tx.Statement.Context, *transfer.BonusInventoryID)
					if err != nil {
						return err
					}
					if diffTotal > 0 && inv.Quantity < diffTotal {
						return errors.New("stok gudang bonus tidak mencukupi untuk penyesuaian transfer")
					}

					inv.Quantity -= diffTotal
					if err := bonusRepo.Update(tx.Statement.Context, inv); err != nil {
						return err
					}

					logType := domain.BonusInventoryOut
					logQty := diffTotal
					if diffTotal < 0 {
						logType = domain.BonusInventoryIn
						logQty = -diffTotal
					}

					_ = bonusRepo.CreateLog(tx.Statement.Context, &domain.BonusInventoryLog{
						BonusInventoryID: inv.ID,
						Type:             logType,
						Quantity:         logQty,
						Notes:            "SALES_TRANSFER_CORRECTION",
					})

					sStock, err := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, transfer.EmployeeID, *transfer.BonusInventoryID)
					if err == nil {
						sStock.Quantity += diffTotal
						if sStock.Quantity < 0 {
							return errors.New("koreksi tidak valid: stok sales bonus tidak boleh kurang dari 0")
						}
						_ = salesBonusRepo.UpdateSalesBonusStock(tx.Statement.Context, sStock)
					}
				}
			}
		} else {
			// Salesman -> Warehouse (RETURN)
			if transfer.ProductID != nil {
				if diffTotal != 0 {
					sStock, err := repo.GetSalesStock(transfer.EmployeeID, *transfer.ProductID)
					if err != nil {
						return err
					}
					sStock.Quantity -= diffTotal
					if sStock.Quantity < 0 {
						sStock.Quantity = 0
					}
					if err := repo.UpdateSalesStock(sStock); err != nil {
						return err
					}
				}

				if diffGood != 0 {
					wStock, err := wRepo.GetStock(transfer.BranchID, *transfer.ProductID, "GOOD", "DEFAULT")
					if err != nil {
						return err
					}
					wStock.Quantity += diffGood
					if wStock.Quantity < 0 {
						return errors.New("koreksi tidak valid: stok gudang GOOD tidak boleh kurang dari 0")
					}
					if err := wRepo.UpdateStock(wStock); err != nil {
						return err
					}

					logType := "IN"
					logQty := diffGood
					if diffGood < 0 {
						logType = "OUT"
						logQty = -diffGood
					}
					if err := wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  transfer.BranchID,
						ProductID: *transfer.ProductID,
						Type:      logType,
						Source:    "SALES_RETURN_CORRECTION (GOOD)",
						Quantity:  logQty,
					}); err != nil {
						return err
					}
				}

				if diffDamaged != 0 {
					dStock, err := wRepo.GetStock(transfer.BranchID, *transfer.ProductID, "DAMAGED", "DEFAULT")
					if err != nil {
						return err
					}
					dStock.Quantity += diffDamaged
					if dStock.Quantity < 0 {
						return errors.New("koreksi tidak valid: stok gudang DAMAGED tidak boleh kurang dari 0")
					}
					if err := wRepo.UpdateStock(dStock); err != nil {
						return err
					}

					logType := "IN"
					logQty := diffDamaged
					if diffDamaged < 0 {
						logType = "OUT"
						logQty = -diffDamaged
					}
					if err := wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  transfer.BranchID,
						ProductID: *transfer.ProductID,
						Type:      logType,
						Source:    "SALES_RETURN_CORRECTION (DAMAGED)",
						Quantity:  logQty,
					}); err != nil {
						return err
					}
				}
			} else if transfer.BonusInventoryID != nil {
				if diffTotal != 0 {
					bonusRepo := repository.NewBonusInventoryRepository(tx)
					salesBonusRepo := repository.NewSalesBonusRepository(tx)

					sStock, err := salesBonusRepo.GetSalesBonusStockItem(tx.Statement.Context, transfer.EmployeeID, *transfer.BonusInventoryID)
					if err == nil {
						sStock.Quantity -= diffTotal
						if sStock.Quantity < 0 {
							sStock.Quantity = 0
						}
						_ = salesBonusRepo.UpdateSalesBonusStock(tx.Statement.Context, sStock)
					}

					inv, err := bonusRepo.GetByID(tx.Statement.Context, *transfer.BonusInventoryID)
					if err == nil {
						inv.Quantity += diffTotal
						if inv.Quantity < 0 {
							return errors.New("koreksi tidak valid: stok gudang bonus tidak boleh kurang dari 0")
						}
						_ = bonusRepo.Update(tx.Statement.Context, inv)

						logType := domain.BonusInventoryIn
						logQty := diffTotal
						if diffTotal < 0 {
							logType = domain.BonusInventoryOut
							logQty = -diffTotal
						}
						_ = bonusRepo.CreateLog(tx.Statement.Context, &domain.BonusInventoryLog{
							BonusInventoryID: inv.ID,
							Type:             logType,
							Quantity:         logQty,
							Notes:            "SALES_RETURN_CORRECTION",
						})
					}
				}
			}
		}

		correction := &domain.SalesTransferCorrection{
			SalesTransferID:  transfer.ID,
			CorrectedBy:      correctedBy,
			OldQuantity:      oldQuantity,
			NewQuantity:      newQuantity,
			OldDamagedQty:    oldDamaged,
			NewDamagedQty:    newDamaged,
			CorrectionReason: reason,
		}
		if err := tx.Create(correction).Error; err != nil {
			return err
		}

		transfer.Quantity = newQuantity
		transfer.ActualQuantity = newActual
		transfer.DamagedQuantity = newDamaged

		noteEntry := fmt.Sprintf("\n[Koreksi oleh %s]: %s (Bagus: %d, Rusak: %d)", correctedBy, reason, newActual, newDamaged)
		transfer.Notes = req.Notes + noteEntry

		return repo.UpdateTransfer(transfer)
	})
}

// GenerateDOForTransfers membuat DO Motoris standalone (tanpa SJ/DeliveryBatch)
// Mendelegasikan ke WarehouseDeliveryOrderUsecase.CreateMotorissDO
func (u *salesTransferUsecase) GenerateDOForTransfers(req GenerateDORequest) error {
	if u.wdoUsecase == nil {
		return errors.New("wdoUsecase tidak tersedia")
	}
	_, err := u.wdoUsecase.CreateMotorissDO(req.SalesTransferIDs)
	return err
}

func (u *salesTransferUsecase) AdjustSalesStock(employeeID uuid.UUID, items []AdjustSalesStockItem) error {
	cargoMap := u.getCargoStockMap(employeeID)
	return u.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			cargoQty := 0
			if cargoStock, ok := cargoMap[item.ProductID]; ok {
				cargoQty = cargoStock.Quantity
			}
			baseQty := item.Quantity - cargoQty

			var stock domain.SalesStock
			if err := tx.Where("employee_id = ? AND product_id = ?", employeeID, item.ProductID).First(&stock).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					stock = domain.SalesStock{
						EmployeeID: employeeID,
						ProductID:  item.ProductID,
						Quantity:   baseQty,
						UpdatedAt:  time.Now(),
					}
					if err := tx.Create(&stock).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				if err := tx.Model(&stock).Updates(map[string]interface{}{
					"quantity":   baseQty,
					"updated_at": time.Now(),
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
