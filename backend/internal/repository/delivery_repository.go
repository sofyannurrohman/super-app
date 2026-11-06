package repository

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type DeliveryRepository interface {
	CreateBatch(batch *domain.DeliveryBatch) error
	GetBatchByDO(doNo string) (*domain.DeliveryBatch, error)
	GetBatchByID(id uuid.UUID) (*domain.DeliveryBatch, error)
	UpdateBatch(batch *domain.DeliveryBatch) error
	UpdateItem(item *domain.DeliveryItem) error
	GetBatchesByDriver(driverID uuid.UUID) ([]domain.DeliveryBatch, error)
	GetBatchesHistoryByDriver(driverID uuid.UUID) ([]domain.DeliveryBatch, error)
	ListBatches(branchIDs []uuid.UUID, salesType string) ([]domain.DeliveryBatch, error)
	GetItemByReceiptNo(receiptNo string) (*domain.DeliveryItem, error)
	GetItemBySONumber(soNumber string) (*domain.DeliveryItem, error)
	DeleteBatch(id uuid.UUID) error
	// Gudang-specific
	GetWarehouseBatches(branchIDs []uuid.UUID) ([]domain.DeliveryBatch, error)
	GetBatchByBarcode(barcodeData string) (*domain.DeliveryBatch, error)
}

type deliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) DeliveryRepository {
	return &deliveryRepository{db}
}

// preloadBatch adalah helper untuk memuat semua relasi yang dibutuhkan secara konsisten.
func (r *deliveryRepository) preloadBatch(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Driver").
		Preload("DailyDriver").
		Preload("Helper").
		Preload("DailyHelper").
		Preload("Vehicle").
		Preload("AdminNota").
		Preload("AdminNota.JobPosition").
		Preload("Supervisor").
		Preload("Supervisor.JobPosition").
		// Alur baru: muat SalesOrder beserta relasinya
		Preload("Items.SalesOrder").
		Preload("Items.SalesOrder.Store").
		Preload("Items.SalesOrder.Employee").
		Preload("Items.SalesOrder.Employee.Company").
		Preload("Items.SalesOrder.Items").
		Preload("Items.SalesOrder.Items.Product").
		Preload("Items.SalesOrder.Returns").
		Preload("Items.SalesOrder.Returns.Items.Product").
		Preload("Items.SalesOrder.Payments").
		// PENTING: muat alokasi batch per item SO agar StartDelivery bisa potong stok per-batch
		Preload("Items.SalesOrder.Items.Batches").
		Preload("Items.SalesOrder.BonusItems").
		Preload("Items.SalesOrder.BonusItems.BonusInventory").
		// Muat barang susulan (subsequent delivery items) beserta relasinya
		Preload("Items.SalesSubsequentItem").
		Preload("Items.SalesSubsequentItem.Product").
		Preload("Items.SalesSubsequentItem.SalesOrder").
		Preload("Items.SalesSubsequentItem.SalesOrder.Store").
		Preload("Items.SalesSubsequentItem.ProductExchange").
		Preload("Items.SalesSubsequentItem.ProductExchange.Store").
		Preload("Items.SalesTransfer").
		Preload("Items.SalesTransfer.Product").
		// Legacy: muat SalesTransaction untuk data lama
		Preload("Items.SalesTransaction").
		Preload("Items.SalesTransaction.Store").
		Preload("Items.SalesTransaction.Payments").
		Preload("Items.SalesTransaction.Employee").
		Preload("Items.SalesTransaction.Employee.Company").
		Preload("Items.SalesTransaction.Items.Product").
		Preload("Spares").
		Preload("Spares.Product").
		Preload("Samples").
		Preload("Samples.Product").
		Preload("Samples.Store").
		Preload("Samples.Salesman").
		Preload("Samples.SalesTesterRequest")
}

func (r *deliveryRepository) CreateBatch(batch *domain.DeliveryBatch) error {
	return r.db.Create(batch).Error
}

func (r *deliveryRepository) GetBatchByDO(doNo string) (*domain.DeliveryBatch, error) {
	var batch domain.DeliveryBatch
	err := r.preloadBatch(r.db).
		Where("delivery_order_no = ?", doNo).First(&batch).Error
	return &batch, err
}

func (r *deliveryRepository) GetBatchByID(id uuid.UUID) (*domain.DeliveryBatch, error) {
	var batch domain.DeliveryBatch
	err := r.preloadBatch(r.db).First(&batch, id).Error
	return &batch, err
}

func (r *deliveryRepository) UpdateBatch(batch *domain.DeliveryBatch) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// If Items are provided, we need to replace them
		// We manually delete to avoid NOT NULL constraint errors with Association.Replace
		if batch.Items != nil {
			if err := tx.Where("delivery_batch_id = ?", batch.ID).Delete(&domain.DeliveryItem{}).Error; err != nil {
				return err
			}
		}
		if batch.Spares != nil {
			if err := tx.Where("delivery_batch_id = ?", batch.ID).Delete(&domain.DeliverySpare{}).Error; err != nil {
				return err
			}
		}
		if batch.Samples != nil {
			if err := tx.Where("delivery_batch_id = ?", batch.ID).Delete(&domain.DeliveryBatchSample{}).Error; err != nil {
				return err
			}
		}

		// Save the batch and its associations
		return tx.Save(batch).Error
	})
}

func (r *deliveryRepository) UpdateItem(item *domain.DeliveryItem) error {
	return r.db.Save(item).Error
}

func (r *deliveryRepository) GetBatchesByDriver(driverID uuid.UUID) ([]domain.DeliveryBatch, error) {
	var batches []domain.DeliveryBatch
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := r.preloadBatch(r.db).
		Where("(driver_id = ? OR helper_id = ?) AND (status NOT IN ? OR finished_at >= ?)",
			driverID, driverID,
			[]domain.DeliveryBatchStatus{domain.DeliveryBatchCompleted, domain.DeliveryBatchWaitingApproval},
			startOfDay).
		Order("created_at DESC").Find(&batches).Error
	return batches, err
}

func (r *deliveryRepository) GetBatchesHistoryByDriver(driverID uuid.UUID) ([]domain.DeliveryBatch, error) {
	var batches []domain.DeliveryBatch
	err := r.db.
		Preload("Items.SalesOrder.Store").
		Preload("Items.SalesOrder.Items.Product").
		Preload("Items.SalesOrder.BonusItems").
		Preload("Items.SalesOrder.BonusItems.BonusInventory").
		Preload("Items.SalesTransaction.Store").
		Preload("Items.SalesTransaction.Items.Product").
		Where("(driver_id = ? OR helper_id = ?) AND status = ?", driverID, driverID, domain.DeliveryBatchCompleted).
		Order("finished_at DESC").Find(&batches).Error
	return batches, err
}

func (r *deliveryRepository) ListBatches(branchIDs []uuid.UUID, salesType string) ([]domain.DeliveryBatch, error) {
	var batches []domain.DeliveryBatch
	db := r.preloadBatch(r.db)
	if len(branchIDs) > 0 {
		db = db.Where("delivery_batches.branch_id IN (?) OR delivery_batches.id IN (SELECT di.delivery_batch_id FROM delivery_items di LEFT JOIN sales_orders so ON so.id = di.sales_order_id LEFT JOIN sales_transactions st ON st.id = di.sales_transaction_id WHERE so.branch_id IN (?) OR st.branch_id IN (?))", branchIDs, branchIDs, branchIDs)
	}

	if salesType == "sales_mt" {
		db = db.Joins("JOIN delivery_items di ON di.delivery_batch_id = delivery_batches.id").
			Joins("JOIN sales_orders so ON so.id = di.sales_order_id").
			Where("so.is_mt = ?", true).Distinct()
	} else if salesType == "sales_to" {
		db = db.Joins("JOIN delivery_items di ON di.delivery_batch_id = delivery_batches.id").
			Joins("JOIN sales_orders so ON so.id = di.sales_order_id").
			Joins("JOIN employees e ON e.id = so.employee_id").
			Joins("JOIN job_positions jp ON jp.id = e.job_position_id").
			Where("so.is_mt = ? AND (jp.title ILIKE '%to%' OR jp.title ILIKE '%take order%')", false).Distinct()
	} else if salesType == "sales_motoris_booster" {
		db = db.Joins("JOIN delivery_items di ON di.delivery_batch_id = delivery_batches.id").
			Joins("JOIN sales_orders so ON so.id = di.sales_order_id").
			Joins("JOIN employees e ON e.id = so.employee_id").
			Joins("JOIN job_positions jp ON jp.id = e.job_position_id").
			Where("so.is_mt = ? AND (jp.title ILIKE '%motoris%' OR jp.title ILIKE '%booster%')", false).Distinct()
	}

	err := db.Order("delivery_batches.created_at DESC").Find(&batches).Error
	return batches, err
}

func (r *deliveryRepository) GetItemByReceiptNo(receiptNo string) (*domain.DeliveryItem, error) {
	var item domain.DeliveryItem
	err := r.db.Preload("SalesTransaction").
		Joins("JOIN sales_transactions ON sales_transactions.id = delivery_items.sales_transaction_id").
		Where("sales_transactions.receipt_no = ?", receiptNo).
		First(&item).Error
	return &item, err
}

func (r *deliveryRepository) GetItemBySONumber(soNumber string) (*domain.DeliveryItem, error) {
	var item domain.DeliveryItem
	err := r.db.Preload("SalesOrder").
		Preload("SalesOrder.Store").
		Preload("SalesOrder.Items.Product").
		Joins("JOIN sales_orders ON sales_orders.id = delivery_items.sales_order_id").
		Where("sales_orders.so_number = ?", soNumber).
		First(&item).Error
	return &item, err
}

func (r *deliveryRepository) DeleteBatch(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Reset foreign keys on nullable related tables
		tx.Model(&domain.SalesOrder{}).Where("delivery_batch_id = ?", id).Update("delivery_batch_id", nil)
		tx.Model(&domain.SalesSubsequentItem{}).Where("delivery_batch_id = ?", id).Update("delivery_batch_id", nil)
		tx.Model(&domain.SalesTesterRequest{}).Where("delivery_batch_id = ?", id).Update("delivery_batch_id", nil)

		// Delete hard-linked child tables
		tx.Where("delivery_batch_id = ?", id).Delete(&domain.DeliveryItem{})
		tx.Where("delivery_batch_id = ?", id).Delete(&domain.DeliverySpare{})
		tx.Where("delivery_batch_id = ?", id).Delete(&domain.DeliveryBatchSample{})

		// Clean up warehouse delivery orders safely
		var wdo domain.WarehouseDeliveryOrder
		if err := tx.Where("delivery_batch_id = ?", id).First(&wdo).Error; err == nil {
			tx.Where("warehouse_delivery_order_id = ?", wdo.ID).Delete(&domain.WarehouseDeliveryOrderItem{})
			tx.Where("warehouse_delivery_order_id = ?", wdo.ID).Delete(&domain.WarehouseDeliveryOrderTransfer{})
			tx.Delete(&wdo)
		}

		// Delete the batch itself
		return tx.Delete(&domain.DeliveryBatch{}, id).Error
	})
}

// GetWarehouseBatches mengembalikan batch dengan status yang relevan untuk menu gudang:
// SUPERVISOR_APPROVED (perlu disiapkan), PICKING (sedang disiapkan),
// ON_DELIVERY (sedang dikirim), COMPLETED (selesai).
func (r *deliveryRepository) GetWarehouseBatches(branchIDs []uuid.UUID) ([]domain.DeliveryBatch, error) {
	var batches []domain.DeliveryBatch
	db := r.preloadBatch(r.db).
		Where("status IN ?", []domain.DeliveryBatchStatus{
			domain.DeliveryBatchSupervisorApproved,
			domain.DeliveryBatchPending,
			domain.DeliveryBatchPicking,
			domain.DeliveryBatchOnDelivery,
			domain.DeliveryBatchCompleted,
		})
	if len(branchIDs) > 0 {
		db = db.Where("branch_id IN (?)", branchIDs)
	}
	err := db.Order("created_at DESC").Find(&batches).Error
	return batches, err
}

// GetBatchByBarcode mencari batch via barcode_data field.
func (r *deliveryRepository) GetBatchByBarcode(barcodeData string) (*domain.DeliveryBatch, error) {
	var batch domain.DeliveryBatch
	err := r.preloadBatch(r.db).
		Where("barcode_data = ?", barcodeData).First(&batch).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}
