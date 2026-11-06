package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/sse"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeliveryUsecase interface {
	CreateBatch(req CreateBatchRequest) (*domain.DeliveryBatch, error)
	SupervisorApproveBatch(batchID uuid.UUID, supervisorID uuid.UUID) error
	ApproveBatch(batchID uuid.UUID, adminID uuid.UUID) error
	AssignArmada(batchID uuid.UUID, driverID uuid.UUID, helperID *uuid.UUID, vehicleID uuid.UUID, supervisorID uuid.UUID) error
	StartDelivery(doNo string) error
	ConfirmItemDelivery(itemID uuid.UUID, status domain.DeliveryItemStatus, notes string) error
	GetDriverLoad(driverID uuid.UUID) ([]map[string]interface{}, error)
	GetDriverTasks(driverID uuid.UUID) ([]domain.DeliveryBatch, error)
	GetDriverHistory(driverID uuid.UUID) ([]domain.DeliveryBatch, error)
	GetBatchDetail(doNo string) (*domain.DeliveryBatch, error)
	GetBatchByID(id uuid.UUID) (*domain.DeliveryBatch, error)
	UpdateBatchItems(batchID uuid.UUID, req CreateBatchRequest) error
	ConfirmItemByReceipt(receiptNo string, notes string) error
	UpdateBatchCash(batchID uuid.UUID, amount float64) error
	UpdateBatchSpares(batchID uuid.UUID, spares []SpareItemRequest) error
	GetPendingSubsequentItems(branchID *uuid.UUID) ([]domain.SalesSubsequentItem, error)
	CreateManualSubsequentItem(req CreateManualSubsequentItemRequest) (*domain.SalesSubsequentItem, error)
	SlipSubsequentItem(batchID uuid.UUID, subItemID uuid.UUID) error
	ListBatches(branchIDs []uuid.UUID, salesType string) ([]domain.DeliveryBatch, error)
	DeleteBatch(id uuid.UUID) error

	// Gudang: melihat surat jalan yang perlu disiapkan + scan barcode
	GetWarehouseBatches(branchIDs []uuid.UUID) ([]domain.DeliveryBatch, error)
	ScanBatch(code string) (*domain.DeliveryBatch, error)
	WarehouseConfirmReady(batchID uuid.UUID) error

	// Public Tracking
	GetPublicStores() ([]domain.Store, error)
	TrackShipment(storeID uuid.UUID) (*PublicTrackingResponse, error)
	AdminReportBatch(batchID uuid.UUID, items []AdminReportItemRequest, departureMileage *int, returnMileage *int) error
	ReviseOrderQuantityBeforeDelivery(batchID uuid.UUID, req ReviseOrderQuantityRequest) error

	// ForceInvoiceBatch: Membuat/memperbaiki faktur penjualan dari SJ COMPLETED
	// yang menggunakan daily driver (tidak bisa melalui alur normal driver tetap).
	// Admin nota menginputkan payment method & amount per item dari laporan kertas harian.
	ForceInvoiceBatch(batchID uuid.UUID, companyID uuid.UUID, items []ForceInvoiceItemRequest) error

	// Susulan Nota / Pesanan SO ke Surat Jalan yang sudah terbuat
	GetEligibleOrdersForBatch(batchID uuid.UUID, search string) ([]domain.SalesOrder, error)
	AppendOrdersToBatch(batchID uuid.UUID, req AppendOrdersRequest) error
}

type AppendOrdersRequest struct {
	SalesOrderIDs []uuid.UUID `json:"sales_order_ids" binding:"required,min=1"`
	Notes         string      `json:"notes"`
}

type AdminReportItemRequest struct {
	ItemID        uuid.UUID                 `json:"item_id" binding:"required"`
	Status        domain.DeliveryItemStatus `json:"status" binding:"required"`
	ReturnedItems []AdminReportReturnedItem `json:"returned_items"`
	Notes         string                    `json:"notes"`
	PaymentMethod string                    `json:"payment_method"`
	TotalAmount   float64                   `json:"total_amount"`
	PaidAmount    float64                   `json:"paid_amount"`
	InvoiceNotes  string                    `json:"invoice_notes"`
}

type ReviseOrderQuantityRequest struct {
	SalesOrderID uuid.UUID                `json:"sales_order_id" binding:"required"`
	Items        []ReviseOrderItemRequest `json:"items" binding:"required"`
}

type ReviseOrderItemRequest struct {
	SalesOrderItemID uuid.UUID `json:"sales_order_item_id" binding:"required"`
	NewQuantity      int       `json:"new_quantity" binding:"min=0"`
}

type AdminReportReturnedItem struct {
	ProductID            uuid.UUID  `json:"product_id"`
	ReturTukarQty        int        `json:"retur_tukar_qty"`
	ReturPotongQty       int        `json:"retur_potong_qty"`
	Reason               string     `json:"reason"`
	Condition            string     `json:"condition"`              // "GOOD" atau "DAMAGED"
	ReplacementProductID *uuid.UUID `json:"replacement_product_id"` // Produk pengganti jika tukar produk
	ReplacementQty       int        `json:"replacement_quantity"`   // Qty produk pengganti
	ReplacementPrice     float64    `json:"replacement_price"`      // Harga per unit produk pengganti
}

// ForceInvoiceItemRequest digunakan admin nota untuk menginputkan
// detail pembayaran per item dari laporan kertas harian driver harian.
type ForceInvoiceItemRequest struct {
	// SalesOrderID adalah ID dari SalesOrder (nota) yang ada dalam batch
	SalesOrderID uuid.UUID `json:"sales_order_id" binding:"required"`
	// PaymentMethod: CASH, TRANSFER, TEMPO
	PaymentMethod string `json:"payment_method" binding:"required"`
	// TotalAmount: total tagihan aktual setelah admin nota verifikasi dengan laporan driver
	// Jika 0, sistem akan menggunakan TotalAmount dari SalesOrder
	TotalAmount float64 `json:"total_amount"`
	// PaidAmount: jumlah yang SUDAH dibayarkan saat pengiriman (penting untuk TEMPO).
	// CASH/TRANSFER: biasanya sama dengan TotalAmount.
	// TEMPO: bisa 0 (bayar nanti) atau sebagian (uang muka).
	// Jika 0 dan payment method bukan TEMPO, sistem gunakan TotalAmount.
	PaidAmount float64 `json:"paid_amount"`
	Notes      string  `json:"notes"`
}

type CreateManualSubsequentItemRequest struct {
	StoreID    uuid.UUID `json:"store_id" binding:"required"`
	SalesmanID uuid.UUID `json:"salesman_id" binding:"required"`
	ProductID  uuid.UUID `json:"product_id" binding:"required"`
	Quantity   int       `json:"quantity" binding:"required,min=1"`
	Notes      string    `json:"notes"`
}

type CreateBatchRequest struct {
	CompanyID              *uuid.UUID                  `json:"company_id"`
	BranchID               *uuid.UUID                  `json:"branch_id"`
	DriverID               *uuid.UUID                  `json:"driver_id"`
	DailyDriverName        string                      `json:"daily_driver_name,omitempty"`
	HelperID               *uuid.UUID                  `json:"helper_id"`
	DailyHelperName        string                      `json:"daily_helper_name,omitempty"`
	VehicleID              *uuid.UUID                  `json:"vehicle_id"`
	Status                 *domain.DeliveryBatchStatus `json:"status,omitempty"`
	SalesOrderIDs          []uuid.UUID                 `json:"sales_order_ids"`
	CollectionOnlySOIDs    []uuid.UUID                 `json:"collection_only_so_ids"`
	CollectionOnlyTrxIDs   []uuid.UUID                 `json:"collection_only_trx_ids"`
	SalesSubsequentItemIDs []uuid.UUID                 `json:"sales_subsequent_item_ids"`
	SalesTransferIDs       []uuid.UUID                 `json:"sales_transfer_ids"`
	Spares                 []SpareItemRequest          `json:"spares"`
	Samples                []DeliverySampleRequest     `json:"samples"`
	DeliveryDate           *time.Time                  `json:"delivery_date,omitempty"`
	DeliveryFee            float64                     `json:"delivery_fee"`

	// Legacy — dipertahankan agar tidak break frontend lama
	SalesTrxIDs []uuid.UUID `json:"sales_transaction_ids"`
}

type SpareItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

type DeliverySampleRequest struct {
	ProductID            uuid.UUID  `json:"product_id" binding:"required"`
	Quantity             int        `json:"quantity" binding:"required,min=1"`
	RecipientType        string     `json:"recipient_type" binding:"required"` // STORE, SALESMAN
	StoreID              *uuid.UUID `json:"store_id"`
	SalesmanID           *uuid.UUID `json:"salesman_id"`
	SalesTesterRequestID *uuid.UUID `json:"sales_tester_request_id"`
	Notes                string     `json:"notes"`
}

type deliveryUsecase struct {
	repo          repository.DeliveryRepository
	salesRepo     repository.SalesTransactionRepository
	soRepo        repository.SalesOrderRepository
	branchRepo    repository.BranchRepository
	warehouseRepo repository.WarehouseRepository
	db            *gorm.DB
	sseHub        *sse.Hub
	fcmClient     FcmUsecase
}

func NewDeliveryUsecase(
	repo repository.DeliveryRepository,
	salesRepo repository.SalesTransactionRepository,
	soRepo repository.SalesOrderRepository,
	branchRepo repository.BranchRepository,
	warehouseRepo repository.WarehouseRepository,
	db *gorm.DB,
	sseHub *sse.Hub,
	fcmClient FcmUsecase,
) DeliveryUsecase {
	return &deliveryUsecase{repo, salesRepo, soRepo, branchRepo, warehouseRepo, db, sseHub, fcmClient}
}

func (u *deliveryUsecase) CreateBatch(req CreateBatchRequest) (*domain.DeliveryBatch, error) {
	now := time.Now()
	// Generate nomor Surat Jalan saat batch dibuat (bukan saat di-approve)
	doNo := fmt.Sprintf("SJ-%s-%s", now.Format("20060102"), uuid.New().String()[:6])

	var dailyDriverID *uuid.UUID
	if req.DailyDriverName != "" {
		var worker domain.DailyWorker
		err := u.db.Where("name = ? AND type = ?", req.DailyDriverName, "Driver").First(&worker).Error
		if err != nil {
			worker = domain.DailyWorker{Name: req.DailyDriverName, Type: "Driver"}
			u.db.Create(&worker)
		}
		dailyDriverID = &worker.ID
	}

	var dailyHelperID *uuid.UUID
	if req.DailyHelperName != "" {
		var worker domain.DailyWorker
		err := u.db.Where("name = ? AND type = ?", req.DailyHelperName, "Helper").First(&worker).Error
		if err != nil {
			worker = domain.DailyWorker{Name: req.DailyHelperName, Type: "Helper"}
			u.db.Create(&worker)
		}
		dailyHelperID = &worker.ID
	}

	batch := &domain.DeliveryBatch{
		CompanyID:       req.CompanyID, // nullable: nil untuk batch lintas company
		BranchID:        req.BranchID,
		DriverID:        req.DriverID,
		DailyDriverID:   dailyDriverID,
		HelperID:        req.HelperID,
		DailyHelperID:   dailyHelperID,
		VehicleID:       req.VehicleID,
		DeliveryOrderNo: &doNo,
		BarcodeData:     doNo, // QR akan di-generate dari field ini di frontend
		Status:          domain.DeliveryBatchWaitingApproval,
		DeliveryDate:    req.DeliveryDate,
		DeliveryFee:     req.DeliveryFee,
	}

	// 1. Resolve Branch Coordinates for Optimization
	startLat := 0.0
	startLon := 0.0
	if req.BranchID != nil {
		branch, _ := u.branchRepo.FindByID(*req.BranchID)
		if branch != nil {
			startLat = branch.Latitude
			startLon = branch.Longitude
		}
	}

	// 2. Fetch and Optimize Sales Orders
	var ordersToOptimize []domain.SalesOrder
	for _, soID := range req.SalesOrderIDs {
		so, err := u.soRepo.FindByID(soID)
		if err == nil && so != nil {
			ordersToOptimize = append(ordersToOptimize, *so)
		}
	}

	optimizedOrders := u.optimizeRoute(startLat, startLon, ordersToOptimize)

	// 3. Tambahkan SO ke dalam batch (alur baru - optimized)
	for i, so := range optimizedOrders {
		soIDCopy := so.ID
		batch.Items = append(batch.Items, domain.DeliveryItem{
			SalesOrderID:       &soIDCopy,
			SalesTransactionID: so.InvoiceID,
			BarcodeData:        so.ID.String(),
			Sequence:           i + 1,
			Status:             domain.DeliveryItemPending,
		})
	}

	// 3b. Tambahkan Collection Only SOs ke dalam batch
	for i, soID := range req.CollectionOnlySOIDs {
		soIDCopy := soID
		batch.Items = append(batch.Items, domain.DeliveryItem{
			SalesOrderID:     &soIDCopy,
			IsCollectionOnly: true,
			BarcodeData:      soIDCopy.String(),
			Sequence:         len(req.SalesOrderIDs) + i + 1,
			Status:           domain.DeliveryItemPending,
		})
	}

	// 3c. Tambahkan Collection Only Trxs ke dalam batch
	for i, trxID := range req.CollectionOnlyTrxIDs {
		trxIDCopy := trxID
		batch.Items = append(batch.Items, domain.DeliveryItem{
			SalesTransactionID: &trxIDCopy,
			IsCollectionOnly:   true,
			BarcodeData:        trxIDCopy.String(),
			Sequence:           len(req.SalesOrderIDs) + len(req.CollectionOnlySOIDs) + i + 1,
			Status:             domain.DeliveryItemPending,
		})
	}

	// Legacy: tambahkan SalesTransaction ke batch (backward-compatibility)
	for i, trxID := range req.SalesTrxIDs {
		trxIDCopy := trxID
		batch.Items = append(batch.Items, domain.DeliveryItem{
			SalesTransactionID: &trxIDCopy,
			Sequence:           len(req.SalesOrderIDs) + len(req.CollectionOnlySOIDs) + len(req.CollectionOnlyTrxIDs) + i + 1,
			Status:             domain.DeliveryItemPending,
		})
	}

	// Tambahkan SalesSubsequentItem ke dalam batch
	for i, subID := range req.SalesSubsequentItemIDs {
		subIDCopy := subID
		batch.Items = append(batch.Items, domain.DeliveryItem{
			SalesSubsequentItemID: &subIDCopy,
			Sequence:              len(req.SalesOrderIDs) + len(req.CollectionOnlySOIDs) + len(req.CollectionOnlyTrxIDs) + len(req.SalesTrxIDs) + i + 1,
			Status:                domain.DeliveryItemPending,
		})
	}

	// Tambahkan SalesTransfer ke dalam batch
	for i, trfID := range req.SalesTransferIDs {
		trfIDCopy := trfID
		batch.Items = append(batch.Items, domain.DeliveryItem{
			SalesTransferID: &trfIDCopy,
			Sequence:        len(req.SalesOrderIDs) + len(req.CollectionOnlySOIDs) + len(req.CollectionOnlyTrxIDs) + len(req.SalesTrxIDs) + len(req.SalesSubsequentItemIDs) + i + 1,
			Status:          domain.DeliveryItemPending,
		})
	}

	// Tambahkan Spares ke dalam batch
	for _, spareReq := range req.Spares {
		batch.Spares = append(batch.Spares, domain.DeliverySpare{
			ProductID: spareReq.ProductID,
			Quantity:  spareReq.Quantity,
		})
	}

	// Tambahkan Samples ke dalam batch
	for _, sampleReq := range req.Samples {
		batch.Samples = append(batch.Samples, domain.DeliveryBatchSample{
			ProductID:            sampleReq.ProductID,
			Quantity:             sampleReq.Quantity,
			RecipientType:        sampleReq.RecipientType,
			StoreID:              sampleReq.StoreID,
			SalesmanID:           sampleReq.SalesmanID,
			SalesTesterRequestID: sampleReq.SalesTesterRequestID,
			Notes:                sampleReq.Notes,
		})
	}

	if err := u.repo.CreateBatch(batch); err != nil {
		return nil, err
	}

	// Update status tester requests yang masuk ke SJ ini
	for _, sample := range batch.Samples {
		if sample.SalesTesterRequestID != nil {
			_ = u.db.Model(&domain.SalesTesterRequest{}).
				Where("id = ?", *sample.SalesTesterRequestID).
				Updates(map[string]interface{}{
					"status":            domain.TesterStatusInDelivery,
					"delivery_batch_id": batch.ID,
				})
		}
	}

	// Push Firebase Notification to Assigned Driver
	if batch.DriverID != nil {
		var emp domain.Employee
		if err := u.db.Select("user_id").Where("id = ?", *batch.DriverID).First(&emp).Error; err == nil {
			_ = u.fcmClient.SendPushNotification(emp.UserID, "Tugas Pengiriman Baru", "Anda mendapatkan penugasan pengiriman baru: "+doNo)
		}
	}

	// Push Firebase Notification to Supervisors to approve the batch
	if u.fcmClient != nil {
		var supervisors []domain.Employee
		errSpv := u.db.Where("branch_id = ? AND job_position_id IN (SELECT id FROM job_positions WHERE title ILIKE '%Supervisor%')", batch.BranchID).Find(&supervisors).Error
		if errSpv == nil {
			for _, spv := range supervisors {
				if spv.UserID != uuid.Nil {
					_ = u.fcmClient.SendPushNotification(spv.UserID, "Persetujuan Surat Jalan", "Admin Nota telah membuat draft Surat Jalan baru yang membutuhkan persetujuan Anda: "+doNo)
				}
			}
		}
	}

	// Kirim notifikasi realtime SSE ke Admin Gudang & Supervisor Gudang di cabang yang bersangkutan
	if u.sseHub != nil && batch.BranchID != nil {
		go func(b *domain.DeliveryBatch, sjNumber string, itemCount int) {
			driverName := "Driver"
			if b.DriverID != nil {
				var drv domain.Employee
				if err := u.db.Select("first_name, last_name").Where("id = ?", *b.DriverID).First(&drv).Error; err == nil {
					driverName = strings.TrimSpace(drv.FirstName + " " + drv.LastName)
				}
			} else if b.DailyDriverID != nil {
				var worker domain.DailyWorker
				if err := u.db.Select("name").Where("id = ?", *b.DailyDriverID).First(&worker).Error; err == nil {
					driverName = worker.Name
				}
			}

			event := sse.NotificationEvent{
				ID:        b.ID.String(),
				Type:      "warehouse_new_sj",
				Title:     "🚚 Surat Jalan Baru",
				Message:   fmt.Sprintf("Surat Jalan %s dibuat untuk %d pengiriman (Driver: %s)", sjNumber, itemCount, driverName),
				CreatedAt: time.Now().Format(time.RFC3339),
				Data: map[string]interface{}{
					"batch_id":    b.ID.String(),
					"do_number":   sjNumber,
					"driver_name": driverName,
					"item_count":  itemCount,
					"branch_id":   b.BranchID.String(),
					"status":      string(b.Status),
				},
				TargetBranchIDs: []string{b.BranchID.String()},
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleAdminGudang),
					string(domain.RoleSupervisorGudang),
				},
			}
			if b.CompanyID != nil {
				event.CompanyID = b.CompanyID.String()
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}(batch, doNo, len(batch.Items))
	}

	// Update status setiap SO menjadi IN_DELIVERY
	for _, soID := range req.SalesOrderIDs {
		_ = u.db.Model(&domain.SalesOrder{}).
			Where("id = ?", soID).
			Updates(map[string]interface{}{
				"status":            domain.SOStatusInDelivery,
				"delivery_batch_id": batch.ID,
				"delivery_order_no": doNo,
			})
	}
	// A redelivered SO retains its invoice number. Date that invoice to the new
	// shipment so it appears with the current delivery once the SJ is created.
	invoiceDate := now
	if req.DeliveryDate != nil && !req.DeliveryDate.IsZero() {
		invoiceDate = *req.DeliveryDate
	}
	for _, so := range optimizedOrders {
		if so.InvoiceID != nil {
			if err := u.db.Model(&domain.SalesTransaction{}).Where("id = ?", *so.InvoiceID).
				Updates(map[string]interface{}{
					"transaction_date": invoiceDate,
					"period_month":     int(invoiceDate.Month()),
					"period_year":      invoiceDate.Year(),
				}).Error; err != nil {
				return nil, fmt.Errorf("gagal memperbarui tanggal faktur pengiriman ulang: %w", err)
			}
		}
	}

	// Update delivery_batch_id untuk CollectionOnlySOIDs
	for _, soID := range req.CollectionOnlySOIDs {
		_ = u.db.Model(&domain.SalesOrder{}).
			Where("id = ?", soID).
			Updates(map[string]interface{}{
				"delivery_batch_id": batch.ID,
				"delivery_order_no": doNo,
			})
	}

	// Update status setiap barang susulan menjadi IN_DELIVERY
	for _, subID := range req.SalesSubsequentItemIDs {
		_ = u.db.Model(&domain.SalesSubsequentItem{}).
			Where("id = ?", subID).
			Updates(map[string]interface{}{
				"status":            domain.SubsequentStatusInDelivery,
				"delivery_batch_id": batch.ID,
			})
	}

	return batch, nil
}

func (u *deliveryUsecase) SupervisorApproveBatch(batchID uuid.UUID, supervisorID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewDeliveryRepository(tx)
		batch, err := repo.GetBatchByID(batchID)
		if err != nil {
			return err
		}

		if batch.Status != domain.DeliveryBatchWaitingApproval {
			return errors.New("batch tidak dalam status menunggu persetujuan supervisor")
		}

		now := time.Now()
		batch.Status = domain.DeliveryBatchSupervisorApproved
		batch.SupervisorID = &supervisorID
		batch.AssignedAt = &now

		if batch.VehicleID != nil {
			if err := tx.Model(&domain.Vehicle{}).Where("id = ?", *batch.VehicleID).Update("status", "IN_USE").Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&domain.DeliveryBatch{}).Where("id = ?", batch.ID).Updates(map[string]interface{}{
			"status":        domain.DeliveryBatchSupervisorApproved,
			"supervisor_id": supervisorID,
			"assigned_at":   now,
			"updated_at":    now,
		}).Error; err != nil {
			return err
		}

		// Tentukan tanggal pengiriman faktur: prioritaskan batch.DeliveryDate (jika diisi), fallback ke now (waktu approval)
		deliveryDate := now
		if batch.DeliveryDate != nil && !batch.DeliveryDate.IsZero() {
			deliveryDate = *batch.DeliveryDate
		}

		// Buat Faktur Penjualan (SalesTransaction) untuk setiap SO
		for _, di := range batch.Items {
			if di.SalesOrderID != nil {
				var so domain.SalesOrder
				if err := tx.Preload("Items").Preload("Store").First(&so, "id = ?", *di.SalesOrderID).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						continue
					}
					return fmt.Errorf("gagal memuat SO untuk pengiriman: %w", err)
				}

				var existingTrxCount int64
				if so.InvoiceID != nil {
					if err := tx.Model(&domain.SalesTransaction{}).Where("id = ?", *so.InvoiceID).Count(&existingTrxCount).Error; err != nil {
						return fmt.Errorf("gagal mengecek faktur SO %s: %w", so.SONumber, err)
					}
				} else {
					if err := tx.Model(&domain.SalesTransaction{}).Where("receipt_no = ?", so.SONumber).Count(&existingTrxCount).Error; err != nil {
						return fmt.Errorf("gagal mengecek faktur SO %s: %w", so.SONumber, err)
					}
				}

				if existingTrxCount == 0 {
					// Generate nomor invoice
					var comp domain.Company
					compCode := "INV"
					if err := tx.First(&comp, "id = ?", so.CompanyID).Error; err == nil && comp.Code != "" {
						compCode = comp.Code
					}
					var invoiceCount int64
					if err := tx.Unscoped().Model(&domain.SalesTransaction{}).
						Where("company_id = ? AND DATE(transaction_date) = ?", so.CompanyID, deliveryDate.Format("2006-01-02")).
						Count(&invoiceCount).Error; err != nil {
						return fmt.Errorf("gagal menghitung urutan faktur: %w", err)
					}
					invoiceNo := fmt.Sprintf("INV-%s-%s-%04d", compCode, deliveryDate.Format("20060102"), invoiceCount+1)
					for attempt := 1; attempt <= 50; attempt++ {
						var dupCount int64
						if err := tx.Unscoped().Model(&domain.SalesTransaction{}).Where("receipt_no = ?", invoiceNo).Count(&dupCount).Error; err != nil {
							return fmt.Errorf("gagal memvalidasi nomor faktur: %w", err)
						}
						if dupCount == 0 {
							break
						}
						invoiceNo = fmt.Sprintf("INV-%s-%s-%04d", compCode, deliveryDate.Format("20060102"), invoiceCount+1+int64(attempt))
					}

					var trxItems []domain.SalesItem
					for _, soItem := range so.Items {
						qty := soItem.OrderedQuantity
						if qty <= 0 {
							qty = soItem.Quantity
						}
						if qty > 0 {
							trxItems = append(trxItems, domain.SalesItem{
								ProductID:          soItem.ProductID,
								Quantity:           qty,
								PriceAtTransaction: soItem.Price,
								DiscountPercent:    soItem.DiscountPercent,
								Subtotal:           float64(qty) * soItem.Price * (1 - soItem.DiscountPercent/100),
							})
						}
					}

					var trxBranchID *uuid.UUID
					if so.BranchID != uuid.Nil {
						trxBranchID = &so.BranchID
					} else if batch.BranchID != nil && *batch.BranchID != uuid.Nil {
						trxBranchID = batch.BranchID
					}

					trx := &domain.SalesTransaction{
						CompanyID:       so.CompanyID,
						BranchID:        trxBranchID,
						StoreID:         &so.StoreID,
						EmployeeID:      so.EmployeeID,
						ReceiptNo:       invoiceNo,
						TotalAmount:     so.TotalAmount,
						DiscountAmount:  so.DiscountAmount,
						StoreCategory:   so.StoreCategory,
						TransactionDate: deliveryDate,
						PeriodMonth:     int(deliveryDate.Month()),
						PeriodYear:      deliveryDate.Year(),
						Status:          "PENDING",
						PaymentStatus:   string(domain.PaymentStatusUnpaid),
						PaymentMethod:   so.PaymentMethod,
						PaymentDueDate:  so.PaymentDueDate,
						Items:           trxItems,
					}

					if err := tx.Create(trx).Error; err != nil {
						return fmt.Errorf("gagal membuat faktur penjualan untuk SO %s: %w", so.SONumber, err)
					}

					// Update SO dengan invoice_id
					if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", so.ID).Updates(map[string]interface{}{
						"invoice_id": trx.ID,
					}).Error; err != nil {
						return fmt.Errorf("gagal mengaitkan faktur ke SO %s: %w", so.SONumber, err)
					}

					// Buat antrean SalesPayment berstatus PENDING agar langsung muncul di Persetujuan Kasir
					targetTrxID := trx.ID
					payMethod := so.PaymentMethod
					if payMethod == "" {
						payMethod = "CASH"
					}
					payMethodForPayment := payMethod
					if payMethod == "TEMPO" {
						payMethodForPayment = "CASH"
					}

					doNoStr := "N/A"
					if batch.DeliveryOrderNo != nil {
						doNoStr = *batch.DeliveryOrderNo
					}

					var existingPayment domain.SalesPayment
					errPay := tx.Where("sales_order_id = ? OR sales_transaction_id = ?", so.ID, targetTrxID).First(&existingPayment).Error
					if errPay != nil {
						if !errors.Is(errPay, gorm.ErrRecordNotFound) {
							return fmt.Errorf("gagal mengecek pembayaran SO %s: %w", so.SONumber, errPay)
						}
						payment := &domain.SalesPayment{
							SalesOrderID:       &so.ID,
							SalesTransactionID: &targetTrxID,
							EmployeeID:         so.EmployeeID,
							Amount:             so.TotalAmount,
							PaymentMethod:      payMethodForPayment,
							PaymentStatus:      "PENDING",
							PaymentDate:        deliveryDate,
							Notes:              fmt.Sprintf("Pengiriman Surat Jalan %s - %s", doNoStr, so.SONumber),
						}
						if err := tx.Create(payment).Error; err != nil {
							return fmt.Errorf("gagal membuat antrean pembayaran untuk SO %s: %w", so.SONumber, err)
						}
					} else if existingPayment.PaymentStatus == "PENDING" {
						if err := tx.Model(&existingPayment).Updates(map[string]interface{}{
							"sales_transaction_id": &targetTrxID,
							"amount":               so.TotalAmount,
							"payment_method":       payMethodForPayment,
							"payment_date":         deliveryDate,
							"notes":                fmt.Sprintf("Pengiriman Surat Jalan %s - %s", doNoStr, so.SONumber),
						}).Error; err != nil {
							return fmt.Errorf("gagal memperbarui antrean pembayaran untuk SO %s: %w", so.SONumber, err)
						}
					}
				} else {
					var existingTrx domain.SalesTransaction
					var errTrx error
					if so.InvoiceID != nil {
						errTrx = tx.First(&existingTrx, "id = ?", *so.InvoiceID).Error
					} else {
						errTrx = tx.Where("receipt_no = ?", so.SONumber).First(&existingTrx).Error
					}

					if errTrx != nil && !errors.Is(errTrx, gorm.ErrRecordNotFound) {
						return fmt.Errorf("gagal mengambil faktur SO %s: %w", so.SONumber, errTrx)
					}

					if errTrx == nil {
						// Update existing transaction
						if err := tx.Model(&existingTrx).Updates(map[string]interface{}{
							"total_amount":     so.TotalAmount,
							"transaction_date": deliveryDate,
							"period_month":     int(deliveryDate.Month()),
							"period_year":      deliveryDate.Year(),
							"updated_at":       time.Now(),
						}).Error; err != nil {
							return fmt.Errorf("gagal memperbarui faktur SO %s: %w", so.SONumber, err)
						}

						// Link SO ke invoice_id yang sudah ada
						if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", so.ID).Update("invoice_id", existingTrx.ID).Error; err != nil {
							return fmt.Errorf("gagal menautkan faktur ke SO %s: %w", so.SONumber, err)
						}

						targetTrxID := existingTrx.ID
						payMethod := so.PaymentMethod
						if payMethod == "" {
							payMethod = "CASH"
						}
						payMethodForPayment := payMethod
						if payMethod == "TEMPO" {
							payMethodForPayment = "CASH"
						}

						doNoStr := "N/A"
						if batch.DeliveryOrderNo != nil {
							doNoStr = *batch.DeliveryOrderNo
						}

						var existingPayment domain.SalesPayment
						errPay := tx.Where("sales_order_id = ? OR sales_transaction_id = ?", so.ID, targetTrxID).First(&existingPayment).Error
						if errPay != nil {
							if !errors.Is(errPay, gorm.ErrRecordNotFound) {
								return fmt.Errorf("gagal mengecek pembayaran SO %s: %w", so.SONumber, errPay)
							}
							payment := &domain.SalesPayment{
								SalesOrderID:       &so.ID,
								SalesTransactionID: &targetTrxID,
								EmployeeID:         so.EmployeeID,
								Amount:             so.TotalAmount,
								PaymentMethod:      payMethodForPayment,
								PaymentStatus:      "PENDING",
								PaymentDate:        deliveryDate,
								Notes:              fmt.Sprintf("Pengiriman Surat Jalan %s - %s", doNoStr, so.SONumber),
							}
							if err := tx.Create(payment).Error; err != nil {
								return fmt.Errorf("gagal membuat antrean pembayaran untuk SO %s: %w", so.SONumber, err)
							}
						} else if existingPayment.PaymentStatus == "PENDING" {
							if err := tx.Model(&existingPayment).Updates(map[string]interface{}{
								"sales_transaction_id": &targetTrxID,
								"amount":               so.TotalAmount,
								"payment_method":       payMethodForPayment,
								"payment_date":         deliveryDate,
							}).Error; err != nil {
								return fmt.Errorf("gagal memperbarui antrean pembayaran untuk SO %s: %w", so.SONumber, err)
							}
						}
					}
				}
			}
		}

		if batch.DriverID != nil && u.fcmClient != nil {
			var emp domain.Employee
			if errDb := tx.First(&emp, "id = ?", *batch.DriverID).Error; errDb == nil {
				doNo := "N/A"
				if batch.DeliveryOrderNo != nil {
					doNo = *batch.DeliveryOrderNo
				}
				go func() {
					_ = u.fcmClient.SendPushNotification(emp.UserID, "Tugas Pengiriman Disetujui", "Tugas pengiriman "+doNo+" telah disetujui oleh Supervisor. Silakan lakukan persiapan di gudang.")
				}()
			}
		}

		return nil
	})
}

func (u *deliveryUsecase) ApproveBatch(batchID uuid.UUID, adminID uuid.UUID) error {
	batch, err := u.repo.GetBatchByID(batchID)
	if err != nil {
		return err
	}

	if batch.Status != domain.DeliveryBatchSupervisorApproved {
		return errors.New("surat jalan belum disetujui oleh supervisor")
	}

	now := time.Now()
	batch.Status = domain.DeliveryBatchPending

	// Generate DO Number
	doNo := fmt.Sprintf("SJ-%s-%s", now.Format("20060102"), uuid.New().String()[:6])
	batch.DeliveryOrderNo = &doNo

	batch.AdminNotaID = &adminID
	batch.ApprovedAt = &now

	return u.repo.UpdateBatch(batch)
}

func (u *deliveryUsecase) AssignArmada(batchID uuid.UUID, driverID uuid.UUID, helperID *uuid.UUID, vehicleID uuid.UUID, supervisorID uuid.UUID) error {
	batch, err := u.repo.GetBatchByID(batchID)
	if err != nil {
		return err
	}

	// Allow revision if status is WAITING_ASSIGNMENT or PENDING
	if batch.Status != domain.DeliveryBatchWaitingAssignment && batch.Status != domain.DeliveryBatchPending {
		return errors.New("surat jalan tidak dalam status yang bisa ditugaskan atau direvisi armada")
	}

	now := time.Now()

	// Only generate DO number and change status if it's the first assignment
	if batch.Status == domain.DeliveryBatchWaitingAssignment {
		doNo := fmt.Sprintf("SJ-%s-%s", now.Format("20060102"), uuid.New().String()[:6])
		batch.Status = domain.DeliveryBatchPending
		batch.DeliveryOrderNo = &doNo
	}

	if batch.VehicleID != nil && *batch.VehicleID != vehicleID {
		_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", *batch.VehicleID).Update("status", "AVAILABLE")
	}

	batch.DriverID = &driverID
	batch.HelperID = helperID
	batch.VehicleID = &vehicleID
	batch.SupervisorID = &supervisorID
	batch.AssignedAt = &now

	_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", vehicleID).Update("status", "IN_USE")

	// Send Push Firebase Notification to Assigned Driver
	var emp domain.Employee
	if err := u.db.Select("user_id").Where("id = ?", driverID).First(&emp).Error; err == nil {
		doNoText := ""
		if batch.DeliveryOrderNo != nil {
			doNoText = *batch.DeliveryOrderNo
		}
		_ = u.fcmClient.SendPushNotification(emp.UserID, "Penugasan Pengiriman Baru", "Anda ditugaskan untuk pengiriman baru: "+doNoText)
	}

	return u.repo.UpdateBatch(batch)
}

func (u *deliveryUsecase) StartDelivery(doNo string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewDeliveryRepository(tx)
		wRepo := repository.NewWarehouseRepository(tx)

		batch, err := repo.GetBatchByDO(doNo)
		if err != nil {
			return err
		}

		// Terima dari SUPERVISOR_APPROVED (alur langsung dipotong Admin Nota),
		// PICKING (alur gudang konfirmasi → driver ambil barang),
		// maupun PENDING (alur langsung berangkat)
		if batch.Status != domain.DeliveryBatchPicking && batch.Status != domain.DeliveryBatchPending && batch.Status != domain.DeliveryBatchSupervisorApproved {
			return errors.New("pengiriman hanya bisa dimulai dari status SUPERVISOR_APPROVED, PICKING, atau PENDING")
		}

		now := time.Now()
		batch.Status = domain.DeliveryBatchOnDelivery
		batch.StartedAt = &now

		// 1. Update status batch ke ON_DELIVERY
		if err := repo.UpdateBatch(batch); err != nil {
			return err
		}

		// 1.b Auto-create PENDING payments for Kasir
		for _, di := range batch.Items {
			var amount float64
			var soID *uuid.UUID
			var trxID *uuid.UUID
			var empID uuid.UUID
			var pm string

			if di.SalesTransactionID != nil {
				var st domain.SalesTransaction
				if err := tx.First(&st, "id = ?", *di.SalesTransactionID).Error; err == nil {
					amount = st.TotalAmount
					trxID = &st.ID
					empID = st.EmployeeID
					pm = st.PaymentMethod
					var so domain.SalesOrder
					if err := tx.Where("invoice_id = ?", st.ID).First(&so).Error; err == nil {
						soID = &so.ID
					}
				}
			} else if di.SalesOrderID != nil {
				var so domain.SalesOrder
				if err := tx.First(&so, "id = ?", *di.SalesOrderID).Error; err == nil {
					amount = so.TotalAmount
					soID = &so.ID
					empID = so.EmployeeID
					pm = so.PaymentMethod
				}
			}

			if amount > 0 && (pm == "CASH" || pm == "SALES_ORDER" || pm == "TUNAI" || pm == "") {
				var existingCount int64
				if trxID != nil {
					tx.Model(&domain.SalesPayment{}).Where("sales_transaction_id = ?", *trxID).Count(&existingCount)
				} else if soID != nil {
					tx.Model(&domain.SalesPayment{}).Where("sales_order_id = ?", *soID).Count(&existingCount)
				}

				if existingCount == 0 {
					payment := &domain.SalesPayment{
						SalesOrderID:       soID,
						SalesTransactionID: trxID,
						EmployeeID:         empID,
						Amount:             amount,
						PaymentMethod:      "CASH",
						PaymentStatus:      "PENDING",
						PaymentDate:        now,
						Notes:              "Setoran langsung Kasir - tagihan otomatis",
					}
					if err := tx.Create(payment).Error; err != nil {
						return fmt.Errorf("gagal membuat antrean pembayaran otomatis: %w", err)
					}
				}
			}
		}

		// 2. Potong Stok Gudang secara otomatis
		if batch.BranchID == nil {
			return errors.New("branch_id tidak ditemukan dalam batch")
		}

		for _, di := range batch.Items {
			if di.SalesSubsequentItemID != nil {
				var subItem domain.SalesSubsequentItem
				if err := tx.Preload("Product").First(&subItem, "id = ?", *di.SalesSubsequentItemID).Error; err == nil {
					stock, err := wRepo.GetStock(*batch.BranchID, subItem.ProductID, "GOOD", "DEFAULT")
					if err == nil && stock != nil {
						physicalDeduct := subItem.Quantity
						if stock.Quantity < physicalDeduct {
							physicalDeduct = stock.Quantity
						}
						if physicalDeduct > 0 {
							stock.Quantity -= physicalDeduct
							_ = wRepo.UpdateStock(stock)

							if batch.DriverID != nil {
								salesTransferRepo := repository.NewSalesTransferRepository(tx)
								sStock, errStock := salesTransferRepo.GetSalesStock(*batch.DriverID, subItem.ProductID)
								if errStock != nil {
									sStock = &domain.SalesStock{
										EmployeeID: *batch.DriverID,
										ProductID:  subItem.ProductID,
										Quantity:   physicalDeduct,
										UpdatedAt:  now,
									}
									_ = tx.Create(sStock)
								} else {
									sStock.Quantity += physicalDeduct
									sStock.UpdatedAt = now
									_ = salesTransferRepo.UpdateSalesStock(sStock)
								}
							}

							wRepo.CreateLog(&domain.WarehouseLog{
								BranchID:  *batch.BranchID,
								ProductID: subItem.ProductID,
								Type:      "OUT",
								Source:    fmt.Sprintf("KIRIM SUSULAN: %s", doNo),
								Quantity:  physicalDeduct,
							})
						}
					}
				}
				continue
			}

			if di.SalesTransferID != nil {
				var transfer domain.SalesTransfer
				if err := tx.Preload("Product").First(&transfer, "id = ?", *di.SalesTransferID).Error; err == nil {
					if transfer.ProductID != nil {
						stock, err := wRepo.GetStock(*batch.BranchID, *transfer.ProductID, "GOOD", "DEFAULT")
						if err == nil && stock != nil {
							deduct := transfer.Quantity
							if stock.Quantity < deduct {
								deduct = stock.Quantity
							}

							if deduct > 0 {
								stock.Quantity -= deduct
								_ = wRepo.UpdateStock(stock)

								wRepo.CreateLog(&domain.WarehouseLog{
									BranchID:  *batch.BranchID,
									ProductID: *transfer.ProductID,
									Type:      "OUT",
									Source:    fmt.Sprintf("KIRIM MUTASI: %s", doNo),
									Quantity:  deduct,
								})

								if batch.DriverID != nil {
									salesTransferRepo := repository.NewSalesTransferRepository(tx)
									sStock, errStock := salesTransferRepo.GetSalesStock(*batch.DriverID, *transfer.ProductID)
									if errStock != nil {
										sStock = &domain.SalesStock{
											EmployeeID: *batch.DriverID,
											ProductID:  *transfer.ProductID,
											Quantity:   deduct,
											UpdatedAt:  now,
										}
										_ = tx.Create(sStock)
									} else {
										sStock.Quantity += deduct
										sStock.UpdatedAt = now
										_ = salesTransferRepo.UpdateSalesStock(sStock)
									}
								}
							}
						}
					}
				}
				continue
			}

			if di.SalesOrder == nil {
				continue
			}

			for _, soi := range di.SalesOrder.Items {
				if soi.Quantity <= 0 || batch.BranchID == nil {
					continue
				}

				// Ambil batch GOOD yang tersedia di gudang cabang (FEFO)
				batches, err := wRepo.GetAvailableBatches(*batch.BranchID, soi.ProductID, "GOOD")
				if err != nil {
					return fmt.Errorf("gagal ambil stok gudang untuk produk %s: %w", soi.ProductID, err)
				}

				remainingToDeduct := soi.Quantity
				totalDeducted := 0

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
					if err := wRepo.UpdateStock(b); err != nil {
						return err
					}

					remainingToDeduct -= deduct
					totalDeducted += deduct

					_ = wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  *batch.BranchID,
						ProductID: soi.ProductID,
						Type:      "OUT",
						Source:    fmt.Sprintf("KIRIM: %s", doNo),
						Quantity:  deduct,
						BatchNo:   b.BatchNo,
					})

					if remainingToDeduct <= 0 {
						break
					}
				}

				// Increase Driver's SalesStock (Gudang Berjalan)
				if batch.DriverID != nil && totalDeducted > 0 {
					salesTransferRepo := repository.NewSalesTransferRepository(tx)
					sStock, err := salesTransferRepo.GetSalesStock(*batch.DriverID, soi.ProductID)
					if err != nil {
						sStock = &domain.SalesStock{
							EmployeeID: *batch.DriverID,
							ProductID:  soi.ProductID,
							Quantity:   totalDeducted,
							UpdatedAt:  now,
						}
						_ = tx.Create(sStock)
					} else {
						sStock.Quantity += totalDeducted
						sStock.UpdatedAt = now
						_ = salesTransferRepo.UpdateSalesStock(sStock)
					}
				}
			}

			// Potong stok BonusInventory (Barang Hadiah / Bonus)
			bonusRepo := repository.NewBonusInventoryRepository(tx)
			salesBonusRepo := repository.NewSalesBonusRepository(tx)
			for _, bonus := range di.SalesOrder.BonusItems {
				if bonus.Quantity > 0 {
					if di.SalesOrder.EmployeeID != uuid.Nil {
						stock, err := salesBonusRepo.GetSalesBonusStockItem(context.Background(), di.SalesOrder.EmployeeID, bonus.BonusInventoryID)
						if err == nil && stock != nil {
							if stock.Quantity >= bonus.Quantity {
								stock.Quantity -= bonus.Quantity
							} else {
								stock.Quantity = 0
							}
							_ = salesBonusRepo.UpdateSalesBonusStock(context.Background(), stock)
						}
					} else {
						inv, err := bonusRepo.GetByID(context.Background(), bonus.BonusInventoryID)
						if err == nil && inv != nil {
							qtyBefore := inv.Quantity
							if inv.Quantity >= bonus.Quantity {
								inv.Quantity -= bonus.Quantity
							} else {
								inv.Quantity = 0
							}
							_ = bonusRepo.Update(context.Background(), inv)

							_ = bonusRepo.CreateLog(context.Background(), &domain.BonusInventoryLog{
								BonusInventoryID: inv.ID,
								Type:             domain.BonusInventoryOut,
								Quantity:         bonus.Quantity,
								QuantityBefore:   qtyBefore,
								QuantityAfter:    inv.Quantity,
								Reference:        fmt.Sprintf("KIRIM DO: %s", doNo),
								Notes:            fmt.Sprintf("Bonus SO: %s", di.SalesOrder.SONumber),
								CreatedBy:        "SISTEM",
							})
						}
					}
				}
			}
		}

		// 3. Pindahkan Barang Cadangan (Spares) dari Gudang ke Gudang Berjalan Supir
		// (Stok Gudang sudah dipotong saat Admin menyimpan cadangan di UpdateBatchSpares)
		for _, spare := range batch.Spares {
			deduct := spare.Quantity
			if deduct > 0 && batch.DriverID != nil {
				salesTransferRepo := repository.NewSalesTransferRepository(tx)
				sStock, errStock := salesTransferRepo.GetSalesStock(*batch.DriverID, spare.ProductID)
				if errStock != nil {
					sStock = &domain.SalesStock{
						EmployeeID: *batch.DriverID,
						ProductID:  spare.ProductID,
						Quantity:   deduct,
						UpdatedAt:  now,
					}
					_ = tx.Create(sStock)
				} else {
					sStock.Quantity += deduct
					sStock.UpdatedAt = now
					_ = salesTransferRepo.UpdateSalesStock(sStock)
				}
			}
		}

		// 4. Potong Stok Gudang untuk Barang Sample / Tester (Samples)
		for _, sample := range batch.Samples {
			stock, err := wRepo.GetStock(*batch.BranchID, sample.ProductID, "GOOD", "DEFAULT")
			if err == nil && stock != nil {
				deduct := sample.Quantity
				if stock.Quantity < deduct {
					deduct = stock.Quantity
				}

				if deduct > 0 {
					stock.Quantity -= deduct
					_ = wRepo.UpdateStock(stock)

					wRepo.CreateLog(&domain.WarehouseLog{
						BranchID:  *batch.BranchID,
						ProductID: sample.ProductID,
						Type:      "OUT",
						Source:    fmt.Sprintf("KIRIM SAMPLE: %s", doNo),
						Quantity:  deduct,
					})

					if batch.DriverID != nil {
						salesTransferRepo := repository.NewSalesTransferRepository(tx)
						sStock, errStock := salesTransferRepo.GetSalesStock(*batch.DriverID, sample.ProductID)
						if errStock != nil {
							sStock = &domain.SalesStock{
								EmployeeID: *batch.DriverID,
								ProductID:  sample.ProductID,
								Quantity:   deduct,
								UpdatedAt:  now,
							}
							_ = tx.Create(sStock)
						} else {
							sStock.Quantity += deduct
							sStock.UpdatedAt = now
							_ = salesTransferRepo.UpdateSalesStock(sStock)
						}
					}
				}
			}

			// Update status of related SalesTesterRequest to IN_DELIVERY
			if sample.SalesTesterRequestID != nil {
				_ = tx.Model(&domain.SalesTesterRequest{}).
					Where("id = ?", *sample.SalesTesterRequestID).
					Updates(map[string]interface{}{
						"status":            domain.TesterStatusInDelivery,
						"delivery_batch_id": batch.ID,
					})
			}
		}

		return nil
	})
}

func (u *deliveryUsecase) ConfirmItemDelivery(itemID uuid.UUID, status domain.DeliveryItemStatus, notes string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var fullItem domain.DeliveryItem
		if err := tx.Preload("DeliveryBatch").First(&fullItem, "id = ?", itemID).Error; err != nil {
			return err
		}

		fullItem.Status = status
		fullItem.Notes = notes
		fullItem.DeliveredAt = &now

		if err := tx.Save(&fullItem).Error; err != nil {
			return err
		}

		if fullItem.Status == domain.DeliveryItemDelivered && fullItem.SalesSubsequentItemID != nil {
			if err := tx.Model(&domain.SalesSubsequentItem{}).
				Where("id = ?", *fullItem.SalesSubsequentItemID).
				Update("status", domain.SubsequentStatusDelivered).Error; err != nil {
				return err
			}

			var subItem domain.SalesSubsequentItem
			if errSub := tx.Select("product_exchange_id").First(&subItem, "id = ?", *fullItem.SalesSubsequentItemID).Error; errSub == nil && subItem.ProductExchangeID != nil {
				var count int64
				tx.Model(&domain.SalesSubsequentItem{}).
					Where("product_exchange_id = ? AND status != ?", *subItem.ProductExchangeID, domain.SubsequentStatusDelivered).
					Count(&count)
				if count == 0 {
					if err := tx.Model(&domain.ProductExchange{}).
						Where("id = ?", *subItem.ProductExchangeID).
						Update("status", domain.ExchangeStatusCompleted).Error; err != nil {
						return err
					}
				}
			}
		}

		if fullItem.Status == domain.DeliveryItemDelivered && fullItem.SalesTransferID != nil {
			var transfer domain.SalesTransfer
			if err := tx.First(&transfer, "id = ?", *fullItem.SalesTransferID).Error; err == nil {
				transfer.Status = domain.SalesTransferCompleted
				transfer.UpdatedAt = now
				_ = tx.Save(&transfer)

				if fullItem.DeliveryBatch != nil && fullItem.DeliveryBatch.DriverID != nil {
					if transfer.ProductID != nil {
						salesTransferRepo := repository.NewSalesTransferRepository(tx)
						dStock, errStock := salesTransferRepo.GetSalesStock(*fullItem.DeliveryBatch.DriverID, *transfer.ProductID)
						if errStock == nil && dStock != nil {
							if dStock.Quantity >= transfer.Quantity {
								dStock.Quantity -= transfer.Quantity
							} else {
								dStock.Quantity = 0
							}
							dStock.UpdatedAt = now
							_ = salesTransferRepo.UpdateSalesStock(dStock)
						}
					}
				}

				if transfer.ProductID != nil {
					salesTransferRepo := repository.NewSalesTransferRepository(tx)
					sStock, errStock := salesTransferRepo.GetSalesStock(transfer.EmployeeID, *transfer.ProductID)
					if errStock != nil {
						sStock = &domain.SalesStock{
							EmployeeID: transfer.EmployeeID,
							ProductID:  *transfer.ProductID,
							Quantity:   transfer.Quantity,
							UpdatedAt:  now,
						}
						_ = tx.Create(sStock)
					} else {
						sStock.Quantity += transfer.Quantity
						sStock.UpdatedAt = now
						_ = salesTransferRepo.UpdateSalesStock(sStock)
					}
				}
			}
		}

		// Check for batch completion
		if fullItem.DeliveryBatchID != uuid.Nil {
			var batch domain.DeliveryBatch
			if err := tx.Preload("Items").Preload("Samples").First(&batch, fullItem.DeliveryBatchID).Error; err == nil {
				allCompleted := true
				for _, bItem := range batch.Items {
					if bItem.Status == domain.DeliveryItemPending {
						allCompleted = false
						break
					}
				}
				if allCompleted && batch.Status != domain.DeliveryBatchCompleted {
					batch.Status = domain.DeliveryBatchCompleted
					batch.FinishedAt = &now
					if batch.VehicleID != nil {
						tx.Model(&domain.Vehicle{}).Where("id = ?", *batch.VehicleID).Update("status", "AVAILABLE")
					}

					// Process samples when batch is completed!
					if err := completeSamplesInTransaction(tx, &batch, now); err != nil {
						return err
					}

					if err := tx.Save(&batch).Error; err != nil {
						return err
					}

					// Insert Daily Worker Salary if using Daily Driver
					if batch.DailyDriverID != nil {
						salary := domain.DailyWorkerSalary{
							DailyWorkerID:   *batch.DailyDriverID,
							BranchID:        batch.BranchID,
							Date:            now,
							TotalPaid:       0, // Akan diinput Kasir
							DeliveryBatchID: &batch.ID,
							Notes:           "Gaji Driver Surat Jalan " + *batch.DeliveryOrderNo,
							Status:          "UNPAID",
						}
						if err := tx.Create(&salary).Error; err != nil {
							return err
						}
					}

					// Insert Daily Worker Salary if using Daily Helper
					if batch.DailyHelperID != nil {
						salary := domain.DailyWorkerSalary{
							DailyWorkerID:   *batch.DailyHelperID,
							BranchID:        batch.BranchID,
							Date:            now,
							TotalPaid:       0, // Akan diinput Kasir
							DeliveryBatchID: &batch.ID,
							Notes:           "Gaji Helper Surat Jalan " + *batch.DeliveryOrderNo,
							Status:          "UNPAID",
						}
						if err := tx.Create(&salary).Error; err != nil {
							return err
						}
					}
				}
			}
		}

		u.publishPODNotification(&fullItem)
		return nil
	})
}

func (u *deliveryUsecase) ConfirmItemByReceipt(receiptNo string, notes string) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var item domain.DeliveryItem
		if err := tx.Preload("DeliveryBatch").Where("receipt_no = ?", receiptNo).First(&item).Error; err != nil {
			return err
		}

		if item.Status == domain.DeliveryItemDelivered {
			return errors.New("barang sudah ditandai sebagai terkirim")
		}

		if item.DeliveryBatch != nil && item.DeliveryBatch.Status == domain.DeliveryBatchWaitingApproval {
			return errors.New("tidak bisa konfirmasi item pada batch yang belum disetujui")
		}

		item.Status = domain.DeliveryItemDelivered
		item.Notes = notes
		item.DeliveredAt = &now

		if err := tx.Save(&item).Error; err != nil {
			return err
		}

		if item.SalesSubsequentItemID != nil {
			if err := tx.Model(&domain.SalesSubsequentItem{}).
				Where("id = ?", *item.SalesSubsequentItemID).
				Update("status", domain.SubsequentStatusDelivered).Error; err != nil {
				return err
			}

			var subItem domain.SalesSubsequentItem
			if errSub := tx.Select("product_exchange_id").First(&subItem, "id = ?", *item.SalesSubsequentItemID).Error; errSub == nil && subItem.ProductExchangeID != nil {
				var count int64
				tx.Model(&domain.SalesSubsequentItem{}).
					Where("product_exchange_id = ? AND status != ?", *subItem.ProductExchangeID, domain.SubsequentStatusDelivered).
					Count(&count)
				if count == 0 {
					if err := tx.Model(&domain.ProductExchange{}).
						Where("id = ?", *subItem.ProductExchangeID).
						Update("status", domain.ExchangeStatusCompleted).Error; err != nil {
						return err
					}
				}
			}
		}

		if item.Status == domain.DeliveryItemDelivered && item.SalesTransferID != nil {
			var transfer domain.SalesTransfer
			if err := tx.First(&transfer, "id = ?", *item.SalesTransferID).Error; err == nil {
				transfer.Status = domain.SalesTransferCompleted
				transfer.UpdatedAt = now
				_ = tx.Save(&transfer)

				if item.DeliveryBatch != nil && item.DeliveryBatch.DriverID != nil {
					if transfer.ProductID != nil {
						salesTransferRepo := repository.NewSalesTransferRepository(tx)
						dStock, errStock := salesTransferRepo.GetSalesStock(*item.DeliveryBatch.DriverID, *transfer.ProductID)
						if errStock == nil && dStock != nil {
							if dStock.Quantity >= transfer.Quantity {
								dStock.Quantity -= transfer.Quantity
							} else {
								dStock.Quantity = 0
							}
							dStock.UpdatedAt = now
							_ = salesTransferRepo.UpdateSalesStock(dStock)
						}
					}
				}

				if transfer.ProductID != nil {
					salesTransferRepo := repository.NewSalesTransferRepository(tx)
					sStock, errStock := salesTransferRepo.GetSalesStock(transfer.EmployeeID, *transfer.ProductID)
					if errStock != nil {
						sStock = &domain.SalesStock{
							EmployeeID: transfer.EmployeeID,
							ProductID:  *transfer.ProductID,
							Quantity:   transfer.Quantity,
							UpdatedAt:  now,
						}
						_ = tx.Create(sStock)
					} else {
						sStock.Quantity += transfer.Quantity
						sStock.UpdatedAt = now
						_ = salesTransferRepo.UpdateSalesStock(sStock)
					}
				}
			}
		}

		// Check for batch completion
		if item.DeliveryBatchID != uuid.Nil {
			var batch domain.DeliveryBatch
			if err := tx.Preload("Items").Preload("Samples").First(&batch, item.DeliveryBatchID).Error; err == nil {
				allCompleted := true
				for _, bItem := range batch.Items {
					if bItem.Status == domain.DeliveryItemPending {
						allCompleted = false
						break
					}
				}
				if allCompleted && batch.Status != domain.DeliveryBatchCompleted {
					batch.Status = domain.DeliveryBatchCompleted
					batch.FinishedAt = &now
					if batch.VehicleID != nil {
						tx.Model(&domain.Vehicle{}).Where("id = ?", *batch.VehicleID).Update("status", "AVAILABLE")
					}

					// Process samples when batch is completed!
					if err := completeSamplesInTransaction(tx, &batch, now); err != nil {
						return err
					}

					if err := tx.Save(&batch).Error; err != nil {
						return err
					}

					// Insert Daily Worker Salary if using Daily Driver
					if batch.DailyDriverID != nil {
						salary := domain.DailyWorkerSalary{
							DailyWorkerID:   *batch.DailyDriverID,
							BranchID:        batch.BranchID,
							Date:            now,
							TotalPaid:       0, // Akan diinput Kasir
							DeliveryBatchID: &batch.ID,
							Notes:           "Gaji Driver Surat Jalan " + *batch.DeliveryOrderNo,
							Status:          "UNPAID",
						}
						if err := tx.Create(&salary).Error; err != nil {
							return err
						}
					}

					// Insert Daily Worker Salary if using Daily Helper
					if batch.DailyHelperID != nil {
						salary := domain.DailyWorkerSalary{
							DailyWorkerID:   *batch.DailyHelperID,
							BranchID:        batch.BranchID,
							Date:            now,
							TotalPaid:       0, // Akan diinput Kasir
							DeliveryBatchID: &batch.ID,
							Notes:           "Gaji Helper Surat Jalan " + *batch.DeliveryOrderNo,
							Status:          "UNPAID",
						}
						if err := tx.Create(&salary).Error; err != nil {
							return err
						}
					}
				}
			}
		}

		u.publishPODNotification(&item)
		return nil
	})
}

func (u *deliveryUsecase) UpdateBatchCash(batchID uuid.UUID, amount float64) error {
	batch, err := u.repo.GetBatchByID(batchID)
	if err != nil {
		return err
	}

	batch.TotalCashCollected = amount
	return u.repo.UpdateBatch(batch)
}

func (u *deliveryUsecase) UpdateBatchSpares(batchID uuid.UUID, sparesReq []SpareItemRequest) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repo := repository.NewDeliveryRepository(tx)
		batch, err := repo.GetBatchByID(batchID)
		if err != nil {
			return err
		}

		// Allow editing spares at any time to support evening adjustments by Warehouse Admin.

		// Get old spares first to compute difference
		var oldSpares []domain.DeliverySpare
		if err := tx.Where("delivery_batch_id = ?", batchID).Find(&oldSpares).Error; err != nil {
			return fmt.Errorf("gagal mengambil barang cadangan lama: %w", err)
		}

		// Delete old spares
		if err := tx.Where("delivery_batch_id = ?", batchID).Delete(&domain.DeliverySpare{}).Error; err != nil {
			return fmt.Errorf("gagal menghapus barang cadangan lama: %w", err)
		}

		// Insert new spares
		var spares []domain.DeliverySpare
		for _, s := range sparesReq {
			spares = append(spares, domain.DeliverySpare{
				DeliveryBatchID: batchID,
				ProductID:       s.ProductID,
				Quantity:        s.Quantity,
			})
		}

		if len(spares) > 0 {
			if err := tx.Create(&spares).Error; err != nil {
				return fmt.Errorf("gagal menyimpan barang cadangan baru: %w", err)
			}
		}

		// Adjust stock dynamically based on difference
		if batch.BranchID != nil {
			wRepo := repository.NewWarehouseRepository(tx)
			var doNo string
			if batch.DeliveryOrderNo != nil {
				doNo = *batch.DeliveryOrderNo
			} else {
				doNo = batch.ID.String()
			}

			newQtyMap := make(map[uuid.UUID]int)
			for _, s := range sparesReq {
				newQtyMap[s.ProductID] = s.Quantity
			}

			oldQtyMap := make(map[uuid.UUID]int)
			for _, s := range oldSpares {
				oldQtyMap[s.ProductID] = s.Quantity
			}

			allProductIDs := make(map[uuid.UUID]bool)
			for id := range newQtyMap {
				allProductIDs[id] = true
			}
			for id := range oldQtyMap {
				allProductIDs[id] = true
			}

			for pID := range allProductIDs {
				newQty := newQtyMap[pID]
				oldQty := oldQtyMap[pID]
				diff := newQty - oldQty
				if diff == 0 {
					continue
				}

				stock, err := wRepo.GetStock(*batch.BranchID, pID, "GOOD", "DEFAULT")
				if err != nil {
					return fmt.Errorf("gagal mendapatkan stok gudang untuk produk %s: %w", pID, err)
				}

				stock.Quantity -= diff
				if err := wRepo.UpdateStock(stock); err != nil {
					return fmt.Errorf("gagal update stok gudang: %w", err)
				}

				logType := "OUT"
				logQty := diff
				logSource := ""
				if diff < 0 {
					logType = "IN"
					logQty = -diff
					logSource = fmt.Sprintf("KEMBALI CADANGAN: %s", doNo)
				} else {
					logSource = fmt.Sprintf("TAMBAH CADANGAN: %s", doNo)
				}

				_ = wRepo.CreateLog(&domain.WarehouseLog{
					BranchID:  *batch.BranchID,
					ProductID: pID,
					Type:      logType,
					Source:    logSource,
					Quantity:  logQty,
					BatchNo:   "DEFAULT",
				})
			}
		}

		return nil
	})
}

func (u *deliveryUsecase) GetPendingSubsequentItems(branchID *uuid.UUID) ([]domain.SalesSubsequentItem, error) {
	var items []domain.SalesSubsequentItem
	query := u.db.Preload("Product").
		Preload("SalesOrder").
		Preload("SalesOrder.Store").
		Preload("ProductExchange").
		Preload("ProductExchange.Store").
		Preload("Store").
		Joins("LEFT JOIN sales_orders ON sales_orders.id = sales_subsequent_items.sales_order_id").
		Joins("LEFT JOIN product_exchanges ON product_exchanges.id = sales_subsequent_items.product_exchange_id").
		Where("sales_subsequent_items.status = ? AND sales_subsequent_items.delivery_batch_id IS NULL", domain.SubsequentStatusPending).
		Where("sales_orders.id IS NULL OR sales_orders.status = ?", domain.SOStatusConfirmed)

	if branchID != nil && *branchID != uuid.Nil {
		query = query.Where("(sales_orders.branch_id = ? OR product_exchanges.branch_id = ?)", *branchID, *branchID)
	}

	err := query.Order("sales_subsequent_items.created_at ASC").Find(&items).Error
	return items, err
}

func (u *deliveryUsecase) SlipSubsequentItem(batchID uuid.UUID, subItemID uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var batch domain.DeliveryBatch
		if err := tx.First(&batch, "id = ?", batchID).Error; err != nil {
			return err
		}

		if batch.Status == domain.DeliveryBatchOnDelivery || batch.Status == domain.DeliveryBatchCompleted {
			return errors.New("tidak bisa menyelipkan barang ke pengiriman yang sudah berjalan atau selesai")
		}

		var subItem domain.SalesSubsequentItem
		if err := tx.First(&subItem, "id = ?", subItemID).Error; err != nil {
			return err
		}

		if subItem.Status != domain.SubsequentStatusPending || subItem.DeliveryBatchID != nil {
			return errors.New("barang susulan sudah tidak berstatus PENDING atau sudah memiliki batch")
		}

		var count int64
		tx.Model(&domain.DeliveryItem{}).Where("delivery_batch_id = ?", batchID).Count(&count)

		di := domain.DeliveryItem{
			DeliveryBatchID:       batchID,
			SalesSubsequentItemID: &subItemID,
			Sequence:              int(count) + 1,
			Status:                domain.DeliveryItemPending,
		}
		if err := tx.Create(&di).Error; err != nil {
			return err
		}

		subItem.Status = domain.SubsequentStatusInDelivery
		subItem.DeliveryBatchID = &batchID
		if err := tx.Save(&subItem).Error; err != nil {
			return err
		}

		return nil
	})
}

func (u *deliveryUsecase) GetDriverTasks(driverID uuid.UUID) ([]domain.DeliveryBatch, error) {
	return u.repo.GetBatchesByDriver(driverID)
}

func (u *deliveryUsecase) GetDriverHistory(driverID uuid.UUID) ([]domain.DeliveryBatch, error) {
	return u.repo.GetBatchesHistoryByDriver(driverID)
}

func (u *deliveryUsecase) GetBatchDetail(doNo string) (*domain.DeliveryBatch, error) {
	return u.repo.GetBatchByDO(doNo)
}

func (u *deliveryUsecase) GetBatchByID(id uuid.UUID) (*domain.DeliveryBatch, error) {
	return u.repo.GetBatchByID(id)
}

func (u *deliveryUsecase) ListBatches(branchIDs []uuid.UUID, salesType string) ([]domain.DeliveryBatch, error) {
	return u.repo.ListBatches(branchIDs, salesType)
}

func (u *deliveryUsecase) UpdateBatchItems(batchID uuid.UUID, req CreateBatchRequest) error {
	batch, err := u.repo.GetBatchByID(batchID)
	if err != nil {
		return err
	}

	if batch.Status == domain.DeliveryBatchPicking || batch.Status == domain.DeliveryBatchOnDelivery || batch.Status == domain.DeliveryBatchCompleted {
		return errors.New("tidak bisa mengedit batch yang sudah mulai diproses gudang, dikirim, atau selesai")
	}

	// 1. Reset status SO lama menjadi CONFIRMED agar kembali ke antrean
	for _, oldItem := range batch.Items {
		if oldItem.SalesOrderID != nil {
			_ = u.db.Model(&domain.SalesOrder{}).
				Where("id = ?", *oldItem.SalesOrderID).
				Updates(map[string]interface{}{
					"status":            domain.SOStatusConfirmed,
					"delivery_batch_id": nil,
					"delivery_order_no": nil,
				})
		}
		if oldItem.SalesSubsequentItemID != nil {
			_ = u.db.Model(&domain.SalesSubsequentItem{}).
				Where("id = ?", *oldItem.SalesSubsequentItemID).
				Updates(map[string]interface{}{
					"status":            domain.SubsequentStatusPending,
					"delivery_batch_id": nil,
				})
		}
	}

	// Reset status old SalesTesterRequests back to PENDING and clear delivery_batch_id
	for _, oldSample := range batch.Samples {
		if oldSample.SalesTesterRequestID != nil {
			_ = u.db.Model(&domain.SalesTesterRequest{}).
				Where("id = ?", *oldSample.SalesTesterRequestID).
				Updates(map[string]interface{}{
					"status":            domain.TesterStatusPending,
					"delivery_batch_id": nil,
				})
		}
	}

	// Release old vehicle if changed
	if batch.Status == domain.DeliveryBatchSupervisorApproved && batch.VehicleID != nil && (req.VehicleID == nil || *batch.VehicleID != *req.VehicleID) {
		_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", *batch.VehicleID).Update("status", "AVAILABLE")
	}

	if req.Status != nil && *req.Status != "" {
		batch.Status = *req.Status
	}

	batch.DriverID = req.DriverID
	batch.Driver = nil // Reset association agar GORM mengupdate kolom FK
	batch.VehicleID = req.VehicleID
	batch.Vehicle = nil // Reset association agar GORM mengupdate kolom FK
	batch.DeliveryDate = req.DeliveryDate

	// Set new vehicle to IN_USE if supervisor approved
	if batch.Status == domain.DeliveryBatchSupervisorApproved && req.VehicleID != nil {
		_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", *req.VehicleID).Update("status", "IN_USE")
	}

	// 3. Resolve Branch Coordinates for Optimization
	startLat := 0.0
	startLon := 0.0
	if batch.BranchID != nil {
		branch, _ := u.branchRepo.FindByID(*batch.BranchID)
		if branch != nil {
			startLat = branch.Latitude
			startLon = branch.Longitude
		}
	}

	// 4. Fetch and Optimize Sales Orders
	var ordersToOptimize []domain.SalesOrder
	for _, soID := range req.SalesOrderIDs {
		so, err := u.soRepo.FindByID(soID)
		if err == nil && so != nil {
			ordersToOptimize = append(ordersToOptimize, *so)
		}
	}

	optimizedOrders := u.optimizeRoute(startLat, startLon, ordersToOptimize)

	// Update items: clear existing and add new
	items := make([]domain.DeliveryItem, 0)

	// Tambahkan SO ke dalam batch (alur baru - optimized)
	for i, so := range optimizedOrders {
		soIDCopy := so.ID
		items = append(items, domain.DeliveryItem{
			DeliveryBatchID: batchID,
			SalesOrderID:    &soIDCopy,
			Sequence:        i + 1,
			Status:          domain.DeliveryItemPending,
		})

		// Update status SO menjadi IN_DELIVERY kembali
		_ = u.db.Model(&domain.SalesOrder{}).
			Where("id = ?", so.ID).
			Updates(map[string]interface{}{
				"status":            domain.SOStatusInDelivery,
				"delivery_batch_id": batchID,
				"delivery_order_no": batch.DeliveryOrderNo,
			})
	}

	// Tambahkan Collection Only SOs ke dalam batch
	for _, soID := range req.CollectionOnlySOIDs {
		soIDCopy := soID
		items = append(items, domain.DeliveryItem{
			DeliveryBatchID:  batchID,
			SalesOrderID:     &soIDCopy,
			IsCollectionOnly: true,
			BarcodeData:      soIDCopy.String(),
			Sequence:         len(items) + 1,
			Status:           domain.DeliveryItemPending,
		})
	}

	// Tambahkan Collection Only Trxs ke dalam batch
	for _, trxID := range req.CollectionOnlyTrxIDs {
		trxIDCopy := trxID
		items = append(items, domain.DeliveryItem{
			DeliveryBatchID:    batchID,
			SalesTransactionID: &trxIDCopy,
			IsCollectionOnly:   true,
			BarcodeData:        trxIDCopy.String(),
			Sequence:           len(items) + 1,
			Status:             domain.DeliveryItemPending,
		})
	}

	// Legacy: tambahkan SalesTransaction ke batch (backward-compatibility)
	for _, trxID := range req.SalesTrxIDs {
		trxIDCopy := trxID
		items = append(items, domain.DeliveryItem{
			DeliveryBatchID:    batchID,
			SalesTransactionID: &trxIDCopy,
			Sequence:           len(items) + 1,
			Status:             domain.DeliveryItemPending,
		})
	}

	// Tambahkan SalesSubsequentItem ke dalam batch
	for _, subID := range req.SalesSubsequentItemIDs {
		subIDCopy := subID
		items = append(items, domain.DeliveryItem{
			DeliveryBatchID:       batchID,
			SalesSubsequentItemID: &subIDCopy,
			Sequence:              len(items) + 1,
			Status:                domain.DeliveryItemPending,
		})
	}

	// Tambahkan SalesTransfer ke dalam batch
	for _, trfID := range req.SalesTransferIDs {
		trfIDCopy := trfID
		items = append(items, domain.DeliveryItem{
			DeliveryBatchID: batchID,
			SalesTransferID: &trfIDCopy,
			Sequence:        len(items) + 1,
			Status:          domain.DeliveryItemPending,
		})
	}
	batch.Items = items

	var spares []domain.DeliverySpare
	for _, spareReq := range req.Spares {
		spares = append(spares, domain.DeliverySpare{
			DeliveryBatchID: batchID,
			ProductID:       spareReq.ProductID,
			Quantity:        spareReq.Quantity,
		})
	}
	batch.Spares = spares

	var samples []domain.DeliveryBatchSample
	for _, sampleReq := range req.Samples {
		samples = append(samples, domain.DeliveryBatchSample{
			DeliveryBatchID:      batchID,
			ProductID:            sampleReq.ProductID,
			Quantity:             sampleReq.Quantity,
			RecipientType:        sampleReq.RecipientType,
			StoreID:              sampleReq.StoreID,
			SalesmanID:           sampleReq.SalesmanID,
			SalesTesterRequestID: sampleReq.SalesTesterRequestID,
			Notes:                sampleReq.Notes,
		})
	}
	batch.Samples = samples

	if err := u.repo.UpdateBatch(batch); err != nil {
		return err
	}

	// Update status setiap barang susulan baru menjadi IN_DELIVERY
	for _, subID := range req.SalesSubsequentItemIDs {
		_ = u.db.Model(&domain.SalesSubsequentItem{}).
			Where("id = ?", subID).
			Updates(map[string]interface{}{
				"status":            domain.SubsequentStatusInDelivery,
				"delivery_batch_id": batchID,
			})
	}

	// Update status tester requests yang masuk ke SJ ini
	for _, sample := range req.Samples {
		if sample.SalesTesterRequestID != nil {
			_ = u.db.Model(&domain.SalesTesterRequest{}).
				Where("id = ?", *sample.SalesTesterRequestID).
				Updates(map[string]interface{}{
					"status":            domain.TesterStatusInDelivery,
					"delivery_batch_id": batchID,
				})
		}
	}

	// Send Push Firebase Notification to Assigned Driver if changed or assigned
	if req.DriverID != nil && (batch.DriverID == nil || *batch.DriverID != *req.DriverID) {
		var emp domain.Employee
		if err := u.db.Select("user_id").Where("id = ?", *req.DriverID).First(&emp).Error; err == nil {
			doNoText := ""
			if batch.DeliveryOrderNo != nil {
				doNoText = *batch.DeliveryOrderNo
			}
			_ = u.fcmClient.SendPushNotification(emp.UserID, "Penugasan Pengiriman Diperbarui", "Ada pembaruan tugas pengiriman untuk Anda: "+doNoText)
		}
	}

	// 2. Update status SO baru menjadi IN_DELIVERY
	for _, soID := range req.SalesOrderIDs {
		_ = u.db.Model(&domain.SalesOrder{}).
			Where("id = ?", soID).
			Updates(map[string]interface{}{
				"status":            domain.SOStatusInDelivery,
				"delivery_batch_id": batch.ID,
				"delivery_order_no": batch.DeliveryOrderNo,
			})
	}

	return nil
}

func (u *deliveryUsecase) DeleteBatch(id uuid.UUID) error {
	batch, err := u.repo.GetBatchByID(id)
	if err != nil {
		return err
	}

	if batch.Status == domain.DeliveryBatchPicking || batch.Status == domain.DeliveryBatchOnDelivery || batch.Status == domain.DeliveryBatchCompleted {
		return errors.New("tidak bisa menghapus batch yang sudah mulai diproses gudang, dikirim, atau selesai")
	}

	// Reset status SO menjadi CONFIRMED agar bisa dibatch lagi
	for _, item := range batch.Items {
		if item.SalesOrderID != nil {
			_ = u.db.Model(&domain.SalesOrder{}).
				Where("id = ?", *item.SalesOrderID).
				Updates(map[string]interface{}{
					"status":            domain.SOStatusConfirmed,
					"delivery_batch_id": nil,
					"delivery_order_no": nil,
				})
		}
		if item.SalesSubsequentItemID != nil {
			_ = u.db.Model(&domain.SalesSubsequentItem{}).
				Where("id = ?", *item.SalesSubsequentItemID).
				Updates(map[string]interface{}{
					"status":            domain.SubsequentStatusPending,
					"delivery_batch_id": nil,
				})
		}
	}

	// Reset status SalesTesterRequests back to PENDING and clear delivery_batch_id
	for _, sample := range batch.Samples {
		if sample.SalesTesterRequestID != nil {
			_ = u.db.Model(&domain.SalesTesterRequest{}).
				Where("id = ?", *sample.SalesTesterRequestID).
				Updates(map[string]interface{}{
					"status":            domain.TesterStatusPending,
					"delivery_batch_id": nil,
				})
		}
	}

	if batch.VehicleID != nil {
		_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", *batch.VehicleID).Update("status", "AVAILABLE")
	}

	return u.repo.DeleteBatch(id)
}

// optimizeRoute implements a basic Nearest Neighbor algorithm for route optimization
func (u *deliveryUsecase) optimizeRoute(startLat, startLon float64, orders []domain.SalesOrder) []domain.SalesOrder {
	if len(orders) <= 1 {
		return orders
	}

	result := make([]domain.SalesOrder, 0, len(orders))
	visited := make([]bool, len(orders))
	currentLat, currentLon := startLat, startLon

	for len(result) < len(orders) {
		nearestIdx := -1
		minDist := math.MaxFloat64

		for i, order := range orders {
			if visited[i] {
				continue
			}

			targetLat, targetLon := currentLat, currentLon
			if order.Store != nil {
				targetLat = order.Store.Latitude
				targetLon = order.Store.Longitude
			}

			dist := haversine(currentLat, currentLon, targetLat, targetLon)
			if dist < minDist {
				minDist = dist
				nearestIdx = i
			}
		}

		if nearestIdx != -1 {
			visited[nearestIdx] = true
			result = append(result, orders[nearestIdx])
			if orders[nearestIdx].Store != nil {
				currentLat = orders[nearestIdx].Store.Latitude
				currentLon = orders[nearestIdx].Store.Longitude
			}
		} else {
			break
		}
	}

	return result
}

// Haversine distance for route optimization helper
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// =============================================================================
// Gudang: Surat Jalan & Barcode Scan
// =============================================================================

// GetWarehouseBatches mengembalikan batch yang relevan untuk menu gudang:
// SUPERVISOR_APPROVED (perlu disiapkan), PICKING (sedang disiapkan),
// ON_DELIVERY (sedang dikirim), COMPLETED (selesai).
func (u *deliveryUsecase) GetWarehouseBatches(branchIDs []uuid.UUID) ([]domain.DeliveryBatch, error) {
	return u.repo.GetWarehouseBatches(branchIDs)
}

// ScanBatch mencari batch berdasarkan nomor SJ atau barcode_data.
func (u *deliveryUsecase) ScanBatch(code string) (*domain.DeliveryBatch, error) {
	// Coba cari via delivery_order_no dulu
	batch, err := u.repo.GetBatchByDO(code)
	if err == nil && batch != nil {
		return batch, nil
	}
	// Fallback: cari via barcode_data
	return u.repo.GetBatchByBarcode(code)
}

// WarehouseConfirmReady: Gudang konfirmasi barang sudah siap diambil driver.
// Status SUPERVISOR_APPROVED atau PENDING → PICKING.
func (u *deliveryUsecase) WarehouseConfirmReady(batchID uuid.UUID) error {
	batch, err := u.repo.GetBatchByID(batchID)
	if err != nil {
		return err
	}
	if batch.Status != domain.DeliveryBatchSupervisorApproved && batch.Status != domain.DeliveryBatchPending {
		return errors.New("surat jalan harus berstatus SUPERVISOR_APPROVED atau PENDING agar bisa dikonfirmasi siap oleh gudang")
	}
	now := time.Now()
	batch.Status = domain.DeliveryBatchPicking
	batch.PickingStartedAt = &now
	return u.repo.UpdateBatch(batch)
}

type WarehouseDetail struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type StoreSequenceDetail struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Address     string     `json:"address"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	Sequence    int        `json:"sequence"`
	Status      string     `json:"status"` // PENDING, DELIVERED, FAILED
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	IsTarget    bool       `json:"is_target"`
}

type TruckLocationDetail struct {
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	StatusText    string    `json:"status_text"`
}

type PublicTrackingResponse struct {
	DeliveryBatchID string                `json:"delivery_batch_id"`
	DeliveryOrderNo string                `json:"delivery_order_no"`
	Status          string                `json:"status"`
	DriverName      string                `json:"driver_name"`
	DriverPhone     string                `json:"driver_phone"`
	VehiclePlate    string                `json:"vehicle_plate"`
	VehicleModel    string                `json:"vehicle_model"`
	Warehouse       *WarehouseDetail      `json:"warehouse"`
	Stores          []StoreSequenceDetail `json:"stores"`
	TruckLocation   *TruckLocationDetail  `json:"truck_location"`
}

func (u *deliveryUsecase) GetPublicStores() ([]domain.Store, error) {
	var stores []domain.Store
	err := u.db.Model(&domain.Store{}).
		Select("id, name, address, latitude, longitude, assigned_employee_id").
		Preload("AssignedEmployee").
		Where("is_active = ? AND latitude IS NOT NULL AND longitude IS NOT NULL AND latitude != 0 AND longitude != 0", true).
		Order("name ASC").
		Find(&stores).Error
	return stores, err
}

func (u *deliveryUsecase) TrackShipment(storeID uuid.UUID) (*PublicTrackingResponse, error) {
	var batch domain.DeliveryBatch

	// Subquery to find batch IDs that contain items for this store
	subQuery := u.db.Table("delivery_items").
		Select("delivery_items.delivery_batch_id").
		Joins("LEFT JOIN sales_orders ON sales_orders.id = delivery_items.sales_order_id").
		Joins("LEFT JOIN sales_transactions ON sales_transactions.id = delivery_items.sales_transaction_id").
		Where("sales_orders.store_id = ? OR sales_transactions.store_id = ?", storeID, storeID)

	err := u.db.Model(&domain.DeliveryBatch{}).
		Preload("Driver").
		Preload("Vehicle").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Preload("Items.SalesOrder").
		Preload("Items.SalesOrder.Store").
		Preload("Items.SalesTransaction").
		Preload("Items.SalesTransaction.Store").
		Where("id IN (?)", subQuery).
		Order(`
			CASE status
				WHEN 'ON_DELIVERY' THEN 1
				WHEN 'PICKING' THEN 2
				WHEN 'SUPERVISOR_APPROVED' THEN 3
				WHEN 'COMPLETED' THEN 4
				ELSE 5
			END ASC,
			created_at DESC
		`).
		First(&batch).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tidak ada pengiriman aktif untuk toko ini")
		}
		return nil, err
	}

	var branch domain.Branch
	if batch.BranchID != nil {
		if err := u.db.First(&branch, *batch.BranchID).Error; err != nil {
			branch.Name = "Gudang Utama"
			branch.Address = "Gudang Utama"
		}
	} else {
		branch.Name = "Gudang Utama"
		branch.Address = "Gudang Utama"
	}

	warehouseDetail := &WarehouseDetail{
		Name:      branch.Name,
		Address:   branch.Address,
		Latitude:  branch.Latitude,
		Longitude: branch.Longitude,
	}

	var stores []StoreSequenceDetail
	var lastCompletedStore *domain.Store
	var firstPendingStore *domain.Store
	var firstStore *domain.Store

	for _, item := range batch.Items {
		var store *domain.Store
		if item.SalesOrder != nil && item.SalesOrder.Store != nil {
			store = item.SalesOrder.Store
		} else if item.SalesTransaction != nil && item.SalesTransaction.Store != nil {
			store = item.SalesTransaction.Store
		}

		if store == nil {
			continue
		}

		if firstStore == nil {
			firstStore = store
		}

		isTarget := store.ID == storeID

		storeDetail := StoreSequenceDetail{
			ID:          store.ID.String(),
			Name:        store.Name,
			Address:     store.Address,
			Latitude:    store.Latitude,
			Longitude:   store.Longitude,
			Sequence:    item.Sequence,
			Status:      string(item.Status),
			DeliveredAt: item.DeliveredAt,
			IsTarget:    isTarget,
		}
		stores = append(stores, storeDetail)

		if item.Status == domain.DeliveryItemDelivered || item.Status == domain.DeliveryItemFailed {
			lastCompletedStore = store
		} else if item.Status == domain.DeliveryItemPending && firstPendingStore == nil {
			firstPendingStore = store
		}
	}

	driverName := "Belum Ditugaskan"
	driverPhone := "-"
	if batch.Driver != nil {
		driverName = batch.Driver.FirstName + " " + batch.Driver.LastName
		driverPhone = batch.Driver.PhoneNumber
	}

	vehiclePlate := "-"
	vehicleModel := "-"
	if batch.Vehicle != nil {
		vehiclePlate = batch.Vehicle.LicensePlate
		vehicleModel = batch.Vehicle.Model
	}

	truckLat := branch.Latitude
	truckLon := branch.Longitude
	statusText := "Pesanan sedang diproses di gudang."
	lastUpdated := batch.UpdatedAt

	if batch.Status == domain.DeliveryBatchOnDelivery {
		if lastCompletedStore != nil {
			truckLat = lastCompletedStore.Latitude
			truckLon = lastCompletedStore.Longitude
			if firstPendingStore != nil {
				statusText = fmt.Sprintf("Selesai mengirim ke %s, sedang menuju ke %s", lastCompletedStore.Name, firstPendingStore.Name)
			} else {
				statusText = fmt.Sprintf("Selesai mengirim ke toko terakhir: %s", lastCompletedStore.Name)
			}
		} else {
			if firstPendingStore != nil {
				statusText = fmt.Sprintf("Driver sedang dalam perjalanan menuju %s", firstPendingStore.Name)
			} else {
				statusText = "Driver sedang dalam perjalanan."
			}
		}
	} else if batch.Status == domain.DeliveryBatchCompleted {
		if lastCompletedStore != nil {
			truckLat = lastCompletedStore.Latitude
			truckLon = lastCompletedStore.Longitude
		}
		statusText = "Semua pengiriman pesanan dalam batch ini telah selesai."
	} else if batch.Status == domain.DeliveryBatchPicking {
		statusText = "Driver sedang memuat barang di gudang."
	}

	truckLocation := &TruckLocationDetail{
		Latitude:      truckLat,
		Longitude:     truckLon,
		LastUpdatedAt: lastUpdated,
		StatusText:    statusText,
	}

	doNo := "DRAFT"
	if batch.DeliveryOrderNo != nil {
		doNo = *batch.DeliveryOrderNo
	}

	return &PublicTrackingResponse{
		DeliveryBatchID: batch.ID.String(),
		DeliveryOrderNo: doNo,
		Status:          string(batch.Status),
		DriverName:      driverName,
		DriverPhone:     driverPhone,
		VehiclePlate:    vehiclePlate,
		VehicleModel:    vehicleModel,
		Warehouse:       warehouseDetail,
		Stores:          stores,
		TruckLocation:   truckLocation,
	}, nil
}

func (u *deliveryUsecase) publishPODNotification(item *domain.DeliveryItem) {
	if u.sseHub == nil {
		return
	}
	go func() {
		// Load details
		batch, err := u.repo.GetBatchByID(item.DeliveryBatchID)
		if err != nil || batch == nil {
			return
		}

		driverName := "Driver"
		if batch.Driver != nil {
			driverName = fmt.Sprintf("%s %s", batch.Driver.FirstName, batch.Driver.LastName)
		}

		storeName := "Toko"
		receiptNo := ""
		var employeeID *uuid.UUID

		if item.SalesOrder != nil {
			receiptNo = item.SalesOrder.SONumber
			employeeID = &item.SalesOrder.EmployeeID
			if item.SalesOrder.Store != nil {
				storeName = item.SalesOrder.Store.Name
			}
		} else if item.SalesTransaction != nil {
			receiptNo = item.SalesTransaction.ReceiptNo
			employeeID = &item.SalesTransaction.EmployeeID
			if item.SalesTransaction.Store != nil {
				storeName = item.SalesTransaction.Store.Name
			}
		} else {
			// Try manual preload to get Store Name and Receipt/SO Number and Salesman EmployeeID
			var fullItem domain.DeliveryItem
			if err := u.db.Preload("SalesOrder.Store").Preload("SalesTransaction.Store").First(&fullItem, "id = ?", item.ID).Error; err == nil {
				if fullItem.SalesOrder != nil {
					receiptNo = fullItem.SalesOrder.SONumber
					employeeID = &fullItem.SalesOrder.EmployeeID
					if fullItem.SalesOrder.Store != nil {
						storeName = fullItem.SalesOrder.Store.Name
					}
				} else if fullItem.SalesTransaction != nil {
					receiptNo = fullItem.SalesTransaction.ReceiptNo
					employeeID = &fullItem.SalesTransaction.EmployeeID
					if fullItem.SalesTransaction.Store != nil {
						storeName = fullItem.SalesTransaction.Store.Name
					}
				}
			}
		}

		if receiptNo == "" {
			receiptNo = "N/A"
		}

		event := sse.NotificationEvent{
			ID:      item.ID.String(),
			Type:    "pod",
			Title:   "Pengiriman Selesai (POD)",
			Message: fmt.Sprintf("Driver %s telah menyelesaikan pengiriman nota %s ke toko %s", driverName, receiptNo, storeName),
			CompanyID: func() string {
				if batch.CompanyID != nil {
					return batch.CompanyID.String()
				}
				return ""
			}(),
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			Data: map[string]interface{}{
				"item_id":     item.ID.String(),
				"receipt_no":  receiptNo,
				"store_name":  storeName,
				"driver_name": driverName,
				"status":      item.Status,
			},
		}
		_ = u.sseHub.PublishNotification(context.Background(), event)

		// Send FCM Push Notification to Salesman
		if employeeID != nil && u.fcmClient != nil {
			var salesman domain.Employee
			if errDb := u.db.First(&salesman, "id = ?", *employeeID).Error; errDb == nil {
				title := "Pengiriman Selesai (POD)"
				body := fmt.Sprintf("Pesanan %s untuk toko %s telah selesai dikirim oleh driver %s.", receiptNo, storeName, driverName)
				if item.Status == "FAILED" || item.Status == "CANCELLED" || item.Status == "RETURNED" {
					title = "Pengiriman Gagal / Retur"
					body = fmt.Sprintf("Pengiriman pesanan %s untuk toko %s berstatus: %s (Driver: %s).", receiptNo, storeName, item.Status, driverName)
				}
				_ = u.fcmClient.SendPushNotification(salesman.UserID, title, body)
			}
		}
	}()
}

func completeSamplesInTransaction(tx *gorm.DB, batch *domain.DeliveryBatch, now time.Time) error {
	salesTransferRepo := repository.NewSalesTransferRepository(tx)
	for _, sample := range batch.Samples {
		if sample.SalesTesterRequestID != nil {
			if err := tx.Model(&domain.SalesTesterRequest{}).
				Where("id = ?", *sample.SalesTesterRequestID).
				Update("status", domain.TesterStatusDelivered).Error; err != nil {
				return err
			}
		}

		if batch.DriverID != nil {
			sStock, err := salesTransferRepo.GetSalesStock(*batch.DriverID, sample.ProductID)
			if err == nil && sStock != nil {
				if sStock.Quantity >= sample.Quantity {
					sStock.Quantity -= sample.Quantity
				} else {
					sStock.Quantity = 0
				}
				sStock.UpdatedAt = now
				if err := salesTransferRepo.UpdateSalesStock(sStock); err != nil {
					return err
				}
			}
		}

		if sample.RecipientType == "SALESMAN" && sample.SalesmanID != nil {
			sStock, err := salesTransferRepo.GetSalesStock(*sample.SalesmanID, sample.ProductID)
			if err != nil {
				sStock = &domain.SalesStock{
					EmployeeID: *sample.SalesmanID,
					ProductID:  sample.ProductID,
					Quantity:   sample.Quantity,
					UpdatedAt:  now,
				}
				if err := tx.Create(sStock).Error; err != nil {
					return err
				}
			} else {
				sStock.Quantity += sample.Quantity
				sStock.UpdatedAt = now
				if err := salesTransferRepo.UpdateSalesStock(sStock); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (u *deliveryUsecase) GetDriverLoad(driverID uuid.UUID) ([]map[string]interface{}, error) {
	actualDriverID := driverID
	var activeBatch domain.DeliveryBatch
	if err := u.db.Where("helper_id = ? AND status NOT IN (?, ?)", driverID, domain.DeliveryBatchCompleted, domain.DeliveryBatchWaitingApproval).
		Order("created_at desc").First(&activeBatch).Error; err == nil && activeBatch.DriverID != nil {
		actualDriverID = *activeBatch.DriverID
	}

	var stocks []domain.SalesStock
	err := u.db.Preload("Product").
		Where("employee_id = ? AND quantity > 0", actualDriverID).
		Find(&stocks).Error
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for _, stock := range stocks {
		price := 0.0
		name := "Unknown Product"
		if stock.Product != nil {
			price = stock.Product.SellingPrice
			name = stock.Product.Name
		}

		itemMap := map[string]interface{}{
			"id":             stock.ProductID,
			"name":           name,
			"price":          price,
			"stock":          stock.Quantity,
			"dropdown_label": fmt.Sprintf("%s (Sisa: %d)", name, stock.Quantity),
		}
		results = append(results, itemMap)
	}

	return results, nil
}

func (u *deliveryUsecase) CreateManualSubsequentItem(req CreateManualSubsequentItemRequest) (*domain.SalesSubsequentItem, error) {
	item := &domain.SalesSubsequentItem{
		StoreID:    &req.StoreID,
		SalesmanID: &req.SalesmanID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		Notes:      req.Notes,
		Status:     domain.SubsequentStatusPending,
	}

	if err := u.db.Create(item).Error; err != nil {
		return nil, err
	}

	return item, nil
}

func (u *deliveryUsecase) AdminReportBatch(batchID uuid.UUID, items []AdminReportItemRequest, departureMileage *int, returnMileage *int) error {
	// 0. Jika batch belum dimulai pengirimannya (masih SUPERVISOR_APPROVED atau PICKING),
	// otomatis jalankan StartDelivery agar stok gudang dipotong dan status berubah menjadi ON_DELIVERY.
	var initialBatch domain.DeliveryBatch
	if err := u.db.First(&initialBatch, "id = ?", batchID).Error; err == nil {
		if initialBatch.Status == domain.DeliveryBatchSupervisorApproved || initialBatch.Status == domain.DeliveryBatchPicking || initialBatch.Status == domain.DeliveryBatchPending {
			if initialBatch.DeliveryOrderNo != nil && *initialBatch.DeliveryOrderNo != "" {
				if errStart := u.StartDelivery(*initialBatch.DeliveryOrderNo); errStart != nil {
					return fmt.Errorf("gagal memotong stok pengiriman: %w", errStart)
				}
			}
		}
	}

	for _, reqItem := range items {
		// 1. Update DeliveryItem status
		if err := u.ConfirmItemDelivery(reqItem.ItemID, reqItem.Status, reqItem.Notes); err != nil {
			return err
		}

		// 2. Process DeliveryItem logic
		if reqItem.Status == domain.DeliveryItemFailed {
			var item domain.DeliveryItem
			if err := u.db.Preload("DeliveryBatch").First(&item, "id = ?", reqItem.ItemID).Error; err == nil && item.SalesOrderID != nil {
				var so domain.SalesOrder
				if err := u.db.Preload("Items").First(&so, "id = ?", *item.SalesOrderID).Error; err == nil {
					var existingTrx domain.SalesTransaction
					var errTrx error
					if so.InvoiceID != nil {
						errTrx = u.db.First(&existingTrx, "id = ?", *so.InvoiceID).Error
					} else {
						errTrx = u.db.Where("receipt_no = ?", so.SONumber).First(&existingTrx).Error
					}
					if errTrx == nil {
						_ = u.db.Model(&item).Update("sales_transaction_id", existingTrx.ID)
					}
					// Update SO status to CANCELLED for failed delivery
					_ = u.db.Model(&so).Update("status", domain.SOStatusCancelled)

					// Buat SalesTransfer untuk pengembalian muatan
					now := time.Now()
					var employeeID uuid.UUID
					if item.DeliveryBatch != nil && item.DeliveryBatch.DriverID != nil {
						employeeID = *item.DeliveryBatch.DriverID
					} else {
						employeeID = so.EmployeeID
					}

					for _, soItem := range so.Items {
						prodID := soItem.ProductID
						transfer := domain.SalesTransfer{
							ReferenceNo:  fmt.Sprintf("RTN-MTN-%d", time.Now().UnixNano()),
							Type:         "RETURN",
							Status:       "PENDING",
							BranchID:     so.BranchID,
							EmployeeID:   employeeID,
							ProductID:    &prodID,
							Quantity:     soItem.Quantity,
							Notes:        "Pengembalian Muatan (Gagal Kirim): " + reqItem.Notes,
							TransferDate: now,
						}
						_ = u.db.Create(&transfer)
					}
				}
			}
		}

		if reqItem.Status == domain.DeliveryItemDelivered {
			var item domain.DeliveryItem
			if err := u.db.Preload("DeliveryBatch").Preload("DeliveryBatch.DailyDriver").Preload("DeliveryBatch.Driver").First(&item, "id = ?", reqItem.ItemID).Error; err == nil {

				if item.SalesOrderID != nil {
					var so domain.SalesOrder
					if err := u.db.Preload("Items").Preload("Items.Product").First(&so, "id = ?", *item.SalesOrderID).Error; err == nil {

						totalDeduction := float64(0)
						totalAddition := float64(0)
						totalReturnQty := 0
						hasTukarProduk := false

						for _, retItem := range reqItem.ReturnedItems {
							totalReturnQty += (retItem.ReturTukarQty + retItem.ReturPotongQty)

							// Find the price of the returned product
							for _, soItem := range so.Items {
								if soItem.ProductID == retItem.ProductID {
									totalDeduction += (float64(retItem.ReturPotongQty) * soItem.Price)
									break
								}
							}

							if retItem.ReplacementProductID != nil && retItem.ReplacementQty > 0 {
								hasTukarProduk = true
								totalAddition += (float64(retItem.ReplacementQty) * retItem.ReplacementPrice)
							}
						}

						// Update DeliveryItem return quantities
						if totalReturnQty > 0 {
							_ = u.db.Model(&domain.DeliveryItem{}).Where("id = ?", reqItem.ItemID).Updates(map[string]interface{}{
								"return_quantity": totalReturnQty,
								"return_reason":   reqItem.Notes,
								"has_return":      true,
							}).Error

							// Buat SalesReturn
							returnRepo := repository.NewSalesReturnRepository(u.db)
							returnNo, _ := returnRepo.GenerateNextReturnNumber(time.Now())

							var employeeID uuid.UUID
							if item.DeliveryBatch != nil && item.DeliveryBatch.DriverID != nil {
								employeeID = *item.DeliveryBatch.DriverID
							} else {
								employeeID = so.EmployeeID
							}

							returnType := string(domain.ReturnTypePotongNota)
							if hasTukarProduk {
								returnType = string(domain.ReturnTypeTukarProduk)
							}

							deliveryReturnNotes := reqItem.Notes
							if !strings.Contains(deliveryReturnNotes, "[DELIVERY_RETURN]") {
								if deliveryReturnNotes != "" {
									deliveryReturnNotes = "[DELIVERY_RETURN] " + deliveryReturnNotes
								} else {
									deliveryReturnNotes = "[DELIVERY_RETURN]"
								}
							}

							sr := domain.SalesReturn{
								ReturnNo:        returnNo,
								SalesOrderID:    &so.ID,
								BranchID:        so.BranchID,
								EmployeeID:      employeeID,
								ReturnType:      returnType,
								Status:          domain.ReturnStatusPendingWarehouse, // Langsung siap diterima fisik oleh Admin Gudang
								TotalAmount:     totalDeduction,
								PriceDifference: totalAddition - totalDeduction,
								Notes:           deliveryReturnNotes,
							}

							for _, retItem := range reqItem.ReturnedItems {
								if retItem.ReturPotongQty > 0 || retItem.ReturTukarQty > 0 {
									qty := retItem.ReturPotongQty + retItem.ReturTukarQty
									price := float64(0)
									for _, soItem := range so.Items {
										if soItem.ProductID == retItem.ProductID {
											price = soItem.Price
											break
										}
									}
									condition := "DAMAGED"
									if strings.ToUpper(retItem.Condition) == "GOOD" {
										condition = "GOOD"
									}

									sri := domain.SalesReturnItem{
										ProductID:            retItem.ProductID,
										Quantity:             qty,
										Price:                price,
										Reason:               retItem.Reason,
										Condition:            condition,
										ReplacementProductID: retItem.ReplacementProductID,
										ReplacementQuantity:  retItem.ReplacementQty,
										ReplacementPrice:     retItem.ReplacementPrice,
									}
									sr.Items = append(sr.Items, sri)
								}
							}
							_ = u.db.Create(&sr)

							// Notifikasi SSE real-time ke Gudang
							if u.sseHub != nil {
								go func(createdSR domain.SalesReturn, storeName string) {
									event := sse.NotificationEvent{
										ID:        createdSR.ID.String(),
										Type:      "warehouse_return",
										Title:     "Retur Fisik Pengiriman Masuk",
										Message:   fmt.Sprintf("Barang retur %s dari toko %s siap diterima di gudang", createdSR.ReturnNo, storeName),
										CompanyID: so.BranchID.String(),
										CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
										Data: map[string]interface{}{
											"return_id":  createdSR.ID.String(),
											"return_no":  createdSR.ReturnNo,
											"store_name": storeName,
											"status":     "PENDING_WAREHOUSE",
										},
									}
									_ = u.sseHub.PublishNotification(context.Background(), event)
								}(sr, so.Store.Name)
							}
						}

						// Create or Update SalesTransaction (Faktur Penjualan)
						finalAmount := so.TotalAmount - totalDeduction + totalAddition
						if reqItem.TotalAmount > 0 {
							finalAmount = reqItem.TotalAmount
						}
						if finalAmount < 0 {
							finalAmount = 0
						}

						paymentMethod := so.PaymentMethod
						if paymentMethod == "" {
							paymentMethod = "CASH"
						}
						if reqItem.PaymentMethod != "" {
							paymentMethod = reqItem.PaymentMethod
						}

						paidNow := finalAmount
						if reqItem.PaidAmount >= 0 && (reqItem.PaidAmount < finalAmount || paymentMethod == "TEMPO") {
							paidNow = reqItem.PaidAmount
						}
						if paymentMethod == "TEMPO" && paidNow == 0 {
							paidNow = finalAmount
						}

						var existingTrx domain.SalesTransaction
						var errTrx error
						if so.InvoiceID != nil {
							errTrx = u.db.Preload("Items").First(&existingTrx, "id = ?", *so.InvoiceID).Error
						} else {
							errTrx = u.db.Preload("Items").Where("receipt_no = ?", so.SONumber).First(&existingTrx).Error
						}

						var currentTrxID uuid.UUID

						if errTrx == nil {
							currentTrxID = existingTrx.ID
							// Update existing transaction items
							for _, retItem := range reqItem.ReturnedItems {
								if retItem.ReturPotongQty > 0 {
									for i, trxItem := range existingTrx.Items {
										if trxItem.ProductID == retItem.ProductID {
											existingTrx.Items[i].Quantity -= retItem.ReturPotongQty
											if existingTrx.Items[i].Quantity < 0 {
												existingTrx.Items[i].Quantity = 0
											}
											existingTrx.Items[i].Subtotal = float64(existingTrx.Items[i].Quantity) * existingTrx.Items[i].PriceAtTransaction
											_ = u.db.Save(&existingTrx.Items[i])
											break
										}
									}
								}
								// Jika ada replacement product, tambahkan item baru jika belum ada
								if retItem.ReplacementProductID != nil && retItem.ReplacementQty > 0 {
									repItem := domain.SalesItem{
										SalesTransactionID: existingTrx.ID,
										ProductID:          *retItem.ReplacementProductID,
										Quantity:           retItem.ReplacementQty,
										PriceAtTransaction: retItem.ReplacementPrice,
										Subtotal:           float64(retItem.ReplacementQty) * retItem.ReplacementPrice,
									}
									_ = u.db.Create(&repItem)
								}
							}
							existingTrx.TotalAmount = finalAmount
							existingTrx.PaymentMethod = paymentMethod
							existingTrx.Status = "PENDING"
							if reqItem.InvoiceNotes != "" {
								n := reqItem.InvoiceNotes
								existingTrx.Notes = &n
							}
							_ = u.db.Omit(clause.Associations).Save(&existingTrx)

							// Update the SO to indicate it has been converted and link the invoice
							u.db.Model(&domain.SalesOrder{}).Where("id = ?", so.ID).Updates(map[string]interface{}{
								"status":     string(domain.SOStatusConverted),
								"invoice_id": existingTrx.ID,
							})
						} else {
							now := time.Now()
							deliveryDate := now
							var batch domain.DeliveryBatch
							if errDb := u.db.First(&batch, "id = ?", batchID).Error; errDb == nil {
								if batch.DeliveryDate != nil && !batch.DeliveryDate.IsZero() {
									deliveryDate = *batch.DeliveryDate
								}
							}

							// Generate nomor invoice — format: INV-{compCode}-{YYYYMMDD}-{XXXX}
							var comp domain.Company
							compCode := "INV"
							if err := u.db.First(&comp, "id = ?", so.CompanyID).Error; err == nil && comp.Code != "" {
								compCode = comp.Code
							}
							var invoiceCount int64
							u.db.Model(&domain.SalesTransaction{}).
								Where("company_id = ? AND DATE(transaction_date) = ?", so.CompanyID, deliveryDate.Format("2006-01-02")).
								Count(&invoiceCount)
							invoiceNo := fmt.Sprintf("INV-%s-%s-%04d", compCode, deliveryDate.Format("20060102"), invoiceCount+1)

							// Copy items to SalesTransaction items
							var trxItems []domain.SalesItem
							for _, soItem := range so.Items {
								trxQty := soItem.OrderedQuantity
								if trxQty <= 0 {
									trxQty = soItem.Quantity
								}
								// Reduce quantity based on ReturPotongQty for this specific product
								for _, retItem := range reqItem.ReturnedItems {
									if soItem.ProductID == retItem.ProductID {
										trxQty -= retItem.ReturPotongQty
									}
								}
								if trxQty > 0 {
									trxItems = append(trxItems, domain.SalesItem{
										ProductID:          soItem.ProductID,
										Quantity:           trxQty,
										PriceAtTransaction: soItem.Price,
										DiscountPercent:    soItem.DiscountPercent,
										Subtotal:           float64(trxQty) * soItem.Price * (1 - soItem.DiscountPercent/100),
									})
								}
							}

							// Tambahkan replacement items ke daftar item faktur
							for _, retItem := range reqItem.ReturnedItems {
								if retItem.ReplacementProductID != nil && retItem.ReplacementQty > 0 {
									trxItems = append(trxItems, domain.SalesItem{
										ProductID:          *retItem.ReplacementProductID,
										Quantity:           retItem.ReplacementQty,
										PriceAtTransaction: retItem.ReplacementPrice,
										Subtotal:           float64(retItem.ReplacementQty) * retItem.ReplacementPrice,
									})
								}
							}

							var trxBranchID *uuid.UUID
							if so.BranchID != uuid.Nil {
								trxBranchID = &so.BranchID
							} else if item.DeliveryBatch != nil && item.DeliveryBatch.BranchID != nil && *item.DeliveryBatch.BranchID != uuid.Nil {
								trxBranchID = item.DeliveryBatch.BranchID
							}

							var notes *string
							if reqItem.InvoiceNotes != "" {
								n := reqItem.InvoiceNotes
								notes = &n
							}

							trx := &domain.SalesTransaction{
								CompanyID:       so.CompanyID,
								BranchID:        trxBranchID,
								StoreID:         &so.StoreID,
								EmployeeID:      so.EmployeeID,
								ReceiptNo:       invoiceNo, // Nomor faktur resmi INV-{compCode}-{YYYYMMDD}-{XXXX}
								TotalAmount:     finalAmount,
								DiscountAmount:  so.DiscountAmount,
								TransactionDate: deliveryDate,
								Status:          "PENDING",
								PaymentStatus:   string(domain.PaymentStatusUnpaid),
								PaymentMethod:   paymentMethod,
								PaymentDueDate:  so.PaymentDueDate,
								StoreCategory:   so.StoreCategory,
								PeriodMonth:     int(deliveryDate.Month()),
								PeriodYear:      deliveryDate.Year(),
								Notes:           notes,
								Items:           trxItems,
							}

							_ = u.db.Create(trx)
							currentTrxID = trx.ID

							// Update the SO to indicate it has been converted and link the invoice
							u.db.Model(&domain.SalesOrder{}).Where("id = ?", so.ID).Updates(map[string]interface{}{
								"status":     string(domain.SOStatusConverted),
								"invoice_id": trx.ID,
							})
						}

						// Handle SalesPayment untuk Kasir
						if currentTrxID != uuid.Nil {
							var existingPayment domain.SalesPayment
							errPay := u.db.Where("sales_transaction_id = ? OR sales_order_id = ?", currentTrxID, so.ID).Where("payment_status = ?", "PENDING").First(&existingPayment).Error

							payMethodForPayment := paymentMethod
							if paymentMethod == "TEMPO" {
								payMethodForPayment = "CASH"
							}

							if errPay == nil {
								newNotes := existingPayment.Notes
								if strings.Contains(newNotes, "[SELISIH") && !strings.Contains(newNotes, "[ADJUSTED]") {
									newNotes += " [ADJUSTED]"
								}
								u.db.Model(&existingPayment).Updates(map[string]interface{}{
									"amount":         paidNow,
									"payment_method": payMethodForPayment,
									"notes":          newNotes,
								})
							} else if paidNow > 0 {
								paymentStatus := "PENDING"
								if item.DeliveryBatch != nil && item.DeliveryBatch.CashSettledAt != nil && item.DeliveryBatch.Status != domain.DeliveryBatchWaitingCrosscheck {
									paymentStatus = "SUCCESS"
								}

								payment := &domain.SalesPayment{
									SalesOrderID:       &so.ID,
									SalesTransactionID: &currentTrxID,
									EmployeeID:         so.EmployeeID,
									Amount:             paidNow,
									PaymentMethod:      payMethodForPayment,
									PaymentStatus:      paymentStatus,
									PaymentDate:        time.Now(),
									Notes:              "Setoran driver harian - input admin nota",
								}
								if errCreatePay := u.db.Create(payment).Error; errCreatePay == nil && paymentStatus == "SUCCESS" {
									_ = u.db.Model(&domain.SalesTransaction{}).Where("id = ?", currentTrxID).UpdateColumn("paid_amount", gorm.Expr("paid_amount + ?", paidNow)).Error
								}
							}
						}
					}
				}
			}
		}
	}

	// 3. Pastikan status batch selesai (COMPLETED) jika seluruh item sudah tidak ada yang PENDING
	var pendingCount int64
	u.db.Model(&domain.DeliveryItem{}).Where("delivery_batch_id = ? AND status = ?", batchID, domain.DeliveryItemPending).Count(&pendingCount)

	now := time.Now()
	if pendingCount == 0 {
		var finalBatch domain.DeliveryBatch
		if err := u.db.Preload("Samples").First(&finalBatch, "id = ?", batchID).Error; err == nil {
			if finalBatch.Status != domain.DeliveryBatchCompleted {
				finalBatch.Status = domain.DeliveryBatchCompleted
				finalBatch.FinishedAt = &now
				if finalBatch.VehicleID != nil {
					_ = u.db.Model(&domain.Vehicle{}).Where("id = ?", *finalBatch.VehicleID).Update("status", "AVAILABLE").Error
				}

				// Selesaikan tester/samples
				_ = completeSamplesInTransaction(u.db, &finalBatch, now)

				// Simpan status COMPLETED
				_ = u.db.Model(&domain.DeliveryBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
					"status":      domain.DeliveryBatchCompleted,
					"finished_at": &now,
				}).Error

				// Insert Daily Worker Salary if using Daily Driver
				if finalBatch.DailyDriverID != nil {
					var countSalary int64
					u.db.Model(&domain.DailyWorkerSalary{}).Where("delivery_batch_id = ? AND daily_worker_id = ?", batchID, *finalBatch.DailyDriverID).Count(&countSalary)
					if countSalary == 0 {
						salary := domain.DailyWorkerSalary{
							DailyWorkerID:   *finalBatch.DailyDriverID,
							BranchID:        finalBatch.BranchID,
							Date:            now,
							TotalPaid:       0,
							DeliveryBatchID: &finalBatch.ID,
							Notes:           "Gaji Driver Surat Jalan " + *finalBatch.DeliveryOrderNo,
							Status:          "UNPAID",
						}
						_ = u.db.Create(&salary)
					}
				}

				// Insert Daily Worker Salary if using Daily Helper
				if finalBatch.DailyHelperID != nil {
					var countSalary int64
					u.db.Model(&domain.DailyWorkerSalary{}).Where("delivery_batch_id = ? AND daily_worker_id = ?", batchID, *finalBatch.DailyHelperID).Count(&countSalary)
					if countSalary == 0 {
						salary := domain.DailyWorkerSalary{
							DailyWorkerID:   *finalBatch.DailyHelperID,
							BranchID:        finalBatch.BranchID,
							Date:            now,
							TotalPaid:       0,
							DeliveryBatchID: &finalBatch.ID,
							Notes:           "Gaji Helper Surat Jalan " + *finalBatch.DeliveryOrderNo,
							Status:          "UNPAID",
						}
						_ = u.db.Create(&salary)
					}
				}
			}
		}
	}

	// 4. Update mileage info in batch (hanya update kolom spesifik agar tidak menimpa status batch)
	mileageUpdates := map[string]interface{}{}
	if departureMileage != nil {
		mileageUpdates["departure_mileage"] = *departureMileage
	}
	if returnMileage != nil {
		mileageUpdates["return_mileage"] = *returnMileage
	}
	if len(mileageUpdates) > 0 {
		_ = u.db.Model(&domain.DeliveryBatch{}).Where("id = ?", batchID).Updates(mileageUpdates).Error
	}

	// Akumulasi Jarak Tempuh ke Kendaraan dan Komponennya
	var b domain.DeliveryBatch
	if err := u.db.First(&b, "id = ?", batchID).Error; err == nil {
		if b.DepartureMileage != nil && b.ReturnMileage != nil && b.VehicleID != nil {
			distance := *b.ReturnMileage - *b.DepartureMileage
			if distance > 0 {
				vehicleRepo := repository.NewVehicleRepository(u.db)
				_ = vehicleRepo.AddMileage(*b.VehicleID, distance)
			}
		}
	}

	return nil
}

// ForceInvoiceBatch membuat faktur penjualan (SalesTransaction) dari SJ
// berstatus COMPLETED yang menggunakan driver harian.
// Admin nota menginputkan payment method, total tagihan, dan jumlah yang sudah
// dibayarkan per nota berdasarkan laporan kertas yang dibawa driver harian.
func (u *deliveryUsecase) ForceInvoiceBatch(batchID uuid.UUID, companyID uuid.UUID, items []ForceInvoiceItemRequest) error {
	var batch domain.DeliveryBatch
	if err := u.db.First(&batch, "id = ?", batchID).Error; err != nil {
		return fmt.Errorf("surat jalan tidak ditemukan: %w", err)
	}
	if batch.Status != domain.DeliveryBatchCompleted && batch.Status != domain.DeliveryBatchWaitingCrosscheck {
		return errors.New("hanya Surat Jalan berstatus SELESAI atau WAITING_CROSSCHECK yang dapat dibuatkan/dilanjutkan fakturnya")
	}

	// Build map SO → request agar lookup O(1)
	itemMap := make(map[uuid.UUID]ForceInvoiceItemRequest, len(items))
	for _, it := range items {
		itemMap[it.SalesOrderID] = it
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		// Map untuk caching company code berdasarkan so.CompanyID
		compCodeMap := make(map[uuid.UUID]string)

		// Ambil semua DeliveryItem dalam batch ini
		var deliveryItems []domain.DeliveryItem
		if err := tx.Preload("SalesOrder").Preload("SalesOrder.Items").Preload("SalesOrder.Items.Product").
			Where("delivery_batch_id = ?", batchID).
			Find(&deliveryItems).Error; err != nil {
			return err
		}

		now := time.Now()

		for _, di := range deliveryItems {
			if di.SalesOrderID == nil || di.SalesOrder == nil {
				continue
			}
			so := di.SalesOrder

			// Dapatkan input dari admin nota untuk nota (SO) ini
			inputReq, hasInput := itemMap[so.ID]

			// Tentukan nilai payment method & amount dari input admin nota
			paymentMethod := so.PaymentMethod
			if paymentMethod == "" {
				paymentMethod = "CASH"
			}
			if hasInput && inputReq.PaymentMethod != "" {
				paymentMethod = inputReq.PaymentMethod
			}
			finalAmount := so.TotalAmount
			if hasInput && inputReq.TotalAmount > 0 {
				finalAmount = inputReq.TotalAmount
			}
			// Jumlah yang dibayarkan sekarang:
			// CASH/TRANSFER → biasanya sama dengan finalAmount
			// TEMPO → bisa 0 (bayar nanti) atau sebagian (uang muka)
			paidNow := finalAmount
			if hasInput && inputReq.PaidAmount >= 0 && (inputReq.PaidAmount < finalAmount || paymentMethod == "TEMPO") {
				paidNow = inputReq.PaidAmount
			}
			// Untuk TEMPO dengan paid=0, kasir tetap perlu tahu ada tagihan → gunakan 1 sen dummy? Tidak.
			// Kita tetap buat SalesPayment dengan amount=0 supaya record ada, tapi hanya tampil jika > 0.
			// Solusi: jika paidNow == 0 dan TEMPO, skip payment creation (bayar nanti via mekanisme lain)
			// → tapi user meminta masuk kasir, jadi kita set paidNow minimal = finalAmount untuk ditampilkan.
			if paymentMethod == "TEMPO" && paidNow == 0 {
				// Tidak ada DP sama sekali: kasir tetap perlu approve tagihan, gunakan total sebagai amount tagihan
				paidNow = finalAmount
			}

			// --- Cek apakah sudah ada SalesTransaction untuk SO ini ---
			var existingTrx domain.SalesTransaction
			var existingErr error
			if so.InvoiceID != nil {
				existingErr = tx.First(&existingTrx, "id = ?", *so.InvoiceID).Error
			} else {
				existingErr = tx.Where("receipt_no = ?", so.SONumber).First(&existingTrx).Error
			}

			if existingErr == nil {
				// Sudah ada (dari AdminReportBatch sebelumnya) — upgrade status & update data
				updates := map[string]interface{}{"status": "PENDING"}
				if hasInput && inputReq.PaymentMethod != "" {
					updates["payment_method"] = paymentMethod
				}
				if hasInput && inputReq.TotalAmount > 0 {
					updates["total_amount"] = finalAmount
				}
				if err := tx.Model(&domain.SalesTransaction{}).Where("id = ?", existingTrx.ID).Updates(updates).Error; err != nil {
					return fmt.Errorf("gagal update faktur untuk SO %s: %w", so.SONumber, err)
				}
				// Pastikan SO sudah ter-link ke invoice ini
				tx.Model(&domain.SalesOrder{}).Where("id = ?", so.ID).Updates(map[string]interface{}{
					"status":     string(domain.SOStatusConverted),
					"invoice_id": existingTrx.ID,
				})
				
				var existingPayment domain.SalesPayment
				errPay := tx.Where("sales_transaction_id = ? OR sales_order_id = ?", existingTrx.ID, so.ID).Where("payment_status = ?", "PENDING").First(&existingPayment).Error
				
				if errPay == nil {
					// Secara dinamis update setoran yang sudah ada di meja Kasir sesuai dengan hasil retur/adjust
					payMethodForPayment := paymentMethod
					if paymentMethod == "TEMPO" {
						payMethodForPayment = "CASH"
					}
					newNotes := existingPayment.Notes
					if strings.Contains(newNotes, "[SELISIH") && !strings.Contains(newNotes, "[ADJUSTED]") {
						newNotes += " [ADJUSTED]"
					}
					tx.Model(&existingPayment).Updates(map[string]interface{}{
						"amount":         paidNow,
						"payment_method": payMethodForPayment,
						"notes":          newNotes,
					})
				} else if errPay != nil && paidNow > 0 {
						trxID := existingTrx.ID
						payMethodForPayment := paymentMethod
						if paymentMethod == "TEMPO" {
							payMethodForPayment = "CASH"
						}

						paymentStatus := "PENDING"
						if batch.CashSettledAt != nil && batch.Status != domain.DeliveryBatchWaitingCrosscheck {
							paymentStatus = "SUCCESS"
						}

						payment := &domain.SalesPayment{
							SalesOrderID:       &so.ID,
							SalesTransactionID: &trxID,
							EmployeeID:         so.EmployeeID,
							Amount:             paidNow,
							PaymentMethod:      payMethodForPayment,
							PaymentStatus:      paymentStatus,
							PaymentDate:        now,
							Notes:              "Setoran driver harian - input admin nota",
						}
						if err := tx.Create(payment).Error; err != nil {
							return fmt.Errorf("gagal membuat setoran untuk SO %s: %w", so.SONumber, err)
						}
						if paymentStatus == "SUCCESS" {
							if err := tx.Model(&domain.SalesTransaction{}).Where("id = ?", trxID).UpdateColumn("paid_amount", gorm.Expr("paid_amount + ?", paidNow)).Error; err != nil {
								return fmt.Errorf("gagal update nominal bayar faktur SO %s: %w", so.SONumber, err)
							}
						}
					}
				continue
			}

			// --- Belum ada: buat SalesTransaction baru ---
			// Dapatkan company code dari so.CompanyID
			compCode, ok := compCodeMap[so.CompanyID]
			if !ok {
				var comp domain.Company
				if err := tx.Select("id, code").First(&comp, "id = ?", so.CompanyID).Error; err == nil && comp.Code != "" {
					compCode = comp.Code
				} else {
					compCode = "WOW" // fallback
				}
				compCodeMap[so.CompanyID] = compCode
			}

			deliveryDate := now
			if batch.DeliveryDate != nil && !batch.DeliveryDate.IsZero() {
				deliveryDate = *batch.DeliveryDate
			}

			// Generate nomor invoice — format: INV-{compCode}-{YYYYMMDD}-{XXXX}
			var invoiceCount int64
			tx.Model(&domain.SalesTransaction{}).Where("company_id = ? AND DATE(transaction_date) = ?", so.CompanyID, deliveryDate.Format("2006-01-02")).Count(&invoiceCount)
			invoiceNo := fmt.Sprintf("INV-%s-%s-%04d", compCode, deliveryDate.Format("20060102"), invoiceCount+1)

			// Salin item dari SO ke SalesTransaction
			var trxItems []domain.SalesItem
			for _, soItem := range so.Items {
				qty := soItem.OrderedQuantity
				if qty <= 0 {
					qty = soItem.Quantity
				}
				if qty > 0 {
					trxItems = append(trxItems, domain.SalesItem{
						ProductID:          soItem.ProductID,
						Quantity:           qty,
						PriceAtTransaction: soItem.Price,
						DiscountPercent:    soItem.DiscountPercent,
						Subtotal:           float64(qty) * soItem.Price * (1 - soItem.DiscountPercent/100),
					})
				}
			}

			var notes *string
			if hasInput && inputReq.Notes != "" {
				n := inputReq.Notes
				notes = &n
			}

			var trxBranchID *uuid.UUID
			if so.BranchID != uuid.Nil {
				trxBranchID = &so.BranchID
			} else if batch.BranchID != nil && *batch.BranchID != uuid.Nil {
				trxBranchID = batch.BranchID
			}

			trx := &domain.SalesTransaction{
				CompanyID:       so.CompanyID,
				BranchID:        trxBranchID,
				StoreID:         &so.StoreID,
				EmployeeID:      so.EmployeeID,
				ReceiptNo:       invoiceNo,
				TotalAmount:     finalAmount,
				DiscountAmount:  so.DiscountAmount,
				StoreCategory:   so.StoreCategory,
				TransactionDate: deliveryDate,
				PeriodMonth:     int(deliveryDate.Month()),
				PeriodYear:      deliveryDate.Year(),
				Status:          "PENDING",
				PaymentStatus:   string(domain.PaymentStatusUnpaid),
				PaymentMethod:   paymentMethod,
				PaymentDueDate:  so.PaymentDueDate,
				Notes:           notes,
				Items:           trxItems,
			}

			if err := tx.Create(trx).Error; err != nil {
				return fmt.Errorf("gagal membuat faktur untuk SO %s: %w", so.SONumber, err)
			}

			// Update SO: link ke invoice dan ubah status ke CONVERTED
			invoiceID := trx.ID
			if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", so.ID).Updates(map[string]interface{}{
				"status":     string(domain.SOStatusConverted),
				"invoice_id": invoiceID,
			}).Error; err != nil {
				return fmt.Errorf("gagal update SO %s: %w", so.SONumber, err)
			}

			// Buat SalesPayment PENDING agar muncul di Persetujuan Kasir.
			// CASH / TRANSFER: paidNow = finalAmount (lunas saat pengiriman)
			// TEMPO: paidNow = jumlah DP yang dibayar driver (bisa sebagian atau full tagihan)
			if paidNow > 0 {
				trxID := trx.ID
				// Untuk TEMPO, pembayaran yang diterima driver masuk sebagai CASH ke kasir
				payMethodForPayment := paymentMethod
				if paymentMethod == "TEMPO" {
					payMethodForPayment = "CASH"
				}

				paymentStatus := "PENDING"
				if batch.CashSettledAt != nil && batch.Status != domain.DeliveryBatchWaitingCrosscheck {
					paymentStatus = "SUCCESS"
				}

				payment := &domain.SalesPayment{
					
						SalesOrderID:       &so.ID,
						SalesTransactionID: &trxID,
					EmployeeID:         so.EmployeeID,
					Amount:             paidNow,
					PaymentMethod:      payMethodForPayment,
					PaymentStatus:      paymentStatus,
					PaymentDate:        now,
					Notes:              "Setoran driver harian - input admin nota",
				}
				if err := tx.Create(payment).Error; err != nil {
					return fmt.Errorf("gagal membuat setoran untuk SO %s: %w", so.SONumber, err)
				}

				// Update invoice status directly if SUCCESS
				if paymentStatus == "SUCCESS" {
					if paidNow >= trx.TotalAmount {
						tx.Model(&domain.SalesTransaction{}).Where("id = ?", trxID).Updates(map[string]interface{}{
							"paid_amount":    gorm.Expr("paid_amount + ?", paidNow),
							"payment_status": "PAID",
						})
					} else {
						tx.Model(&domain.SalesTransaction{}).Where("id = ?", trxID).Updates(map[string]interface{}{
							"paid_amount":    gorm.Expr("paid_amount + ?", paidNow),
							"payment_status": "PARTIAL",
						})
					}
				}
			}
		}

		return nil
	})
}

func (u *deliveryUsecase) ReviseOrderQuantityBeforeDelivery(batchID uuid.UUID, req ReviseOrderQuantityRequest) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var batch domain.DeliveryBatch
		if err := tx.First(&batch, "id = ?", batchID).Error; err != nil {
			return errors.New("surat jalan tidak ditemukan")
		}

		if batch.Status != domain.DeliveryBatchPending && batch.Status != domain.DeliveryBatchSupervisorApproved {
			return errors.New("hanya surat jalan dengan status PENDING atau SUPERVISOR_APPROVED yang bisa direvisi kuantitasnya")
		}

		var so domain.SalesOrder
		if err := tx.Preload("Items").First(&so, "id = ?", req.SalesOrderID).Error; err != nil {
			return errors.New("sales order tidak ditemukan")
		}

		var newTotal float64
		// Buat map reqItem untuk akses cepat
		reqItemsMap := make(map[uuid.UUID]int)
		for _, ri := range req.Items {
			reqItemsMap[ri.SalesOrderItemID] = ri.NewQuantity
		}

		for _, soItem := range so.Items {
			if newQty, exists := reqItemsMap[soItem.ID]; exists {
				soItem.OrderedQuantity = newQty
				soItem.Quantity = newQty
				soItem.Subtotal = float64(newQty) * soItem.Price
				if err := tx.Save(&soItem).Error; err != nil {
					return err
				}
			}
			newTotal += soItem.Subtotal
		}

		so.TotalAmount = newTotal
		if err := tx.Save(&so).Error; err != nil {
			return err
		}

		if so.InvoiceID != nil {
			var invoice domain.SalesTransaction
			if err := tx.Preload("Items").First(&invoice, "id = ?", *so.InvoiceID).Error; err == nil {
				var invNewTotal float64
				for _, invItem := range invoice.Items {
					// Cari soItem yang memiliki ProductID sama
					for _, soItem := range so.Items {
						if invItem.ProductID == soItem.ProductID {
							if newQty, exists := reqItemsMap[soItem.ID]; exists {
								invItem.Quantity = newQty
								invItem.Subtotal = float64(newQty) * invItem.PriceAtTransaction
								tx.Save(&invItem)
							}
							break
						}
					}
					invNewTotal += invItem.Subtotal
				}
				invoice.TotalAmount = invNewTotal
				tx.Omit(clause.Associations).Save(&invoice)
			}
		}

		// Sinkronisasi antrean Persetujuan Kasir (SalesPayment PENDING)
		targetAmount := newTotal
		if so.InvoiceID != nil {
			var inv domain.SalesTransaction
			if err := tx.Select("total_amount").First(&inv, "id = ?", *so.InvoiceID).Error; err == nil {
				targetAmount = inv.TotalAmount
			}
		}

		var pendingPayments []domain.SalesPayment
		pQuery := tx.Where("payment_status = ?", "PENDING")
		if so.InvoiceID != nil {
			pQuery = pQuery.Where("sales_order_id = ? OR sales_transaction_id = ?", so.ID, *so.InvoiceID)
		} else {
			pQuery = pQuery.Where("sales_order_id = ?", so.ID)
		}
		if err := pQuery.Find(&pendingPayments).Error; err == nil {
			for _, p := range pendingPayments {
				updatedNote := p.Notes
				if !strings.Contains(updatedNote, "[ADJUSTED]") {
					if updatedNote != "" {
						updatedNote = "[ADJUSTED] " + updatedNote
					} else {
						updatedNote = "[ADJUSTED] Revisi Surat Jalan oleh Admin Nota"
					}
				}
				tx.Model(&domain.SalesPayment{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
					"amount": targetAmount,
					"notes":  updatedNote,
				})
			}
		}

		return nil
	})
}

func (u *deliveryUsecase) GetEligibleOrdersForBatch(batchID uuid.UUID, search string) ([]domain.SalesOrder, error) {
	var batch domain.DeliveryBatch
	if err := u.db.First(&batch, "id = ?", batchID).Error; err != nil {
		return nil, errors.New("surat jalan / batch tidak ditemukan")
	}

	if batch.Status == domain.DeliveryBatchCompleted {
		return nil, errors.New("tidak dapat menyusulkan nota ke surat jalan yang sudah selesai (COMPLETED)")
	}

	query := u.db.Model(&domain.SalesOrder{}).
		Preload("Store").
		Preload("Employee").
		Preload("Employee.Company").
		Preload("Items").
		Preload("Items.Product").
		Preload("Payments").
		Where("sales_orders.status = ?", domain.SOStatusConfirmed).
		Where("sales_orders.delivery_batch_id IS NULL").
		Where("sales_orders.id NOT IN (SELECT di.sales_order_id FROM delivery_items di JOIN delivery_batches db ON db.id = di.delivery_batch_id WHERE di.sales_order_id IS NOT NULL AND db.status NOT IN ('CANCELLED', 'REJECTED'))")

	if batch.BranchID != nil {
		query = query.Where("sales_orders.branch_id = ?", *batch.BranchID)
	}

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query = query.Joins("LEFT JOIN stores ON stores.id = sales_orders.store_id").
			Joins("LEFT JOIN employees ON employees.id = sales_orders.employee_id").
			Where("LOWER(sales_orders.so_number) LIKE ? OR LOWER(stores.name) LIKE ? OR LOWER(employees.name) LIKE ?", s, s, s)
	}

	var orders []domain.SalesOrder
	if err := query.Order("sales_orders.created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (u *deliveryUsecase) AppendOrdersToBatch(batchID uuid.UUID, req AppendOrdersRequest) error {
	if len(req.SalesOrderIDs) == 0 {
		return errors.New("tidak ada pesanan yang dipilih untuk disusulkan")
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		var batch domain.DeliveryBatch
		if err := tx.First(&batch, "id = ?", batchID).Error; err != nil {
			return errors.New("surat jalan / batch tidak ditemukan")
		}

		if batch.Status == domain.DeliveryBatchCompleted {
			return errors.New("tidak dapat menyusulkan nota ke surat jalan yang sudah berstatus SELESAI (COMPLETED)")
		}

		var count int64
		if err := tx.Model(&domain.DeliveryItem{}).Where("delivery_batch_id = ?", batchID).Count(&count).Error; err != nil {
			return err
		}

		lastSeq := int(count)

		for i, soID := range req.SalesOrderIDs {
			var so domain.SalesOrder
			if err := tx.First(&so, "id = ?", soID).Error; err != nil {
				return fmt.Errorf("pesanan dengan ID %s tidak ditemukan", soID.String())
			}

			if so.DeliveryBatchID != nil {
				return fmt.Errorf("pesanan %s sudah dimasukkan ke surat jalan lain", so.SONumber)
			}

			if so.Status != domain.SOStatusConfirmed {
				return fmt.Errorf("pesanan %s tidak berstatus CONFIRMED (status saat ini: %s)", so.SONumber, so.Status)
			}

			var existingCount int64
			if err := tx.Model(&domain.DeliveryItem{}).
				Joins("JOIN delivery_batches db ON db.id = delivery_items.delivery_batch_id").
				Where("delivery_items.sales_order_id = ? AND db.status NOT IN ('CANCELLED', 'REJECTED')", so.ID).
				Count(&existingCount).Error; err == nil && existingCount > 0 {
				return fmt.Errorf("pesanan %s sudah terdaftar dalam Surat Jalan", so.SONumber)
			}

			soIDCopy := so.ID
			itemNotes := req.Notes
			if itemNotes == "" {
				itemNotes = "Nota Susulan"
			} else if !strings.Contains(itemNotes, "Susulan") {
				itemNotes = "Susulan: " + itemNotes
			}

			di := domain.DeliveryItem{
				DeliveryBatchID: batchID,
				SalesOrderID:    &soIDCopy,
				IsSubsequent:    true,
				BarcodeData:     so.ID.String(),
				Sequence:        lastSeq + i + 1,
				Status:          domain.DeliveryItemPending,
				Notes:           itemNotes,
			}

			if err := tx.Create(&di).Error; err != nil {
				return err
			}

			// Update SalesOrder status and delivery_batch_id
			updates := map[string]interface{}{
				"delivery_batch_id": batchID,
				"status":            domain.SOStatusInDelivery,
			}
			if batch.Status == domain.DeliveryBatchOnDelivery {
				now := time.Now()
				updates["shipped_at"] = &now
			}

			if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", so.ID).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

