package repository

import (
	"fmt"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type WarehouseRepository interface {
	// Stock
	GetStock(branchID, productID uuid.UUID, stockType string, batchNo string) (*domain.WarehouseStock, error)
	GetStockByID(id uuid.UUID) (*domain.WarehouseStock, error)
	DeleteStock(id uuid.UUID) error
	UpdateStock(stock *domain.WarehouseStock) error
	GetStocksByBranchID(branchID uuid.UUID) ([]domain.WarehouseStock, error)
	GetAvailableBatches(branchID, productID uuid.UUID, stockType string) ([]domain.WarehouseStock, error) // For FEFO

	// Transfers
	GetTransfersByBranch(branchID uuid.UUID) ([]domain.ProductTransfer, error)
	GetPendingTransfers(branchID uuid.UUID) ([]domain.ProductTransfer, error)
	GetTransferByID(id uuid.UUID) (*domain.ProductTransfer, error)
	GetTransferByDO(doNo string) (*domain.ProductTransfer, error)
	UpdateTransfer(transfer *domain.ProductTransfer) error

	// Logs
	CreateLog(log *domain.WarehouseLog) error
	GetLogsByBranchID(branchID uuid.UUID) ([]domain.WarehouseLog, error)

	// Reserve / Release untuk mekanisme SO
	ReserveStock(branchID, productID uuid.UUID, batchNo string, qty int) error // Tambah reserved_quantity
	ReleaseStock(branchID, productID uuid.UUID, batchNo string, qty int) error // Kurangi reserved_quantity
}

type warehouseRepository struct {
	db *gorm.DB
}

func NewWarehouseRepository(db *gorm.DB) WarehouseRepository {
	return &warehouseRepository{db}
}

func (r *warehouseRepository) GetStock(branchID, productID uuid.UUID, stockType string, batchNo string) (*domain.WarehouseStock, error) {
	var stock domain.WarehouseStock
	if stockType == "" {
		stockType = "GOOD"
	}
	if batchNo == "" {
		batchNo = "DEFAULT"
	}
	// Query matching batchNo, with fallback between 'DEFAULT' and '' to handle unnormalized legacy/opname rows
	err := r.db.Preload("Branch").Preload("Product").
		Where("branch_id = ? AND product_id = ? AND stock_type = ? AND (batch_no = ? OR (? = 'DEFAULT' AND batch_no = ''))", branchID, productID, stockType, batchNo, batchNo).
		Order("quantity DESC").
		First(&stock).Error
	if err == gorm.ErrRecordNotFound {
		return &domain.WarehouseStock{BranchID: branchID, ProductID: productID, StockType: stockType, BatchNo: batchNo, Quantity: 0}, nil
	}
	return &stock, err
}

func (r *warehouseRepository) GetStockByID(id uuid.UUID) (*domain.WarehouseStock, error) {
	var stock domain.WarehouseStock
	err := r.db.Preload("Branch").Preload("Product").First(&stock, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func (r *warehouseRepository) DeleteStock(id uuid.UUID) error {
	return r.db.Delete(&domain.WarehouseStock{}, "id = ?", id).Error
}

func (r *warehouseRepository) GetAvailableBatches(branchID, productID uuid.UUID, stockType string) ([]domain.WarehouseStock, error) {
	var stocks []domain.WarehouseStock
	if stockType == "" {
		stockType = "GOOD"
	}
	err := r.db.Where("branch_id = ? AND product_id = ? AND stock_type = ? AND quantity > 0 AND batch_no != 'BACKORDER'", branchID, productID, stockType).
		Order("expiry_date ASC NULLS LAST, created_at ASC").
		Find(&stocks).Error
	return stocks, err
}

func (r *warehouseRepository) UpdateStock(stock *domain.WarehouseStock) error {
	if stock.ID == uuid.Nil {
		return r.db.Create(stock).Error
	}
	return r.db.Save(stock).Error
}

func (r *warehouseRepository) GetStocksByBranchID(branchID uuid.UUID) ([]domain.WarehouseStock, error) {
	var stocks []domain.WarehouseStock
	q := r.db.Preload("Product").Preload("Branch")
	if branchID != uuid.Nil {
		q = q.Where("branch_id = ?", branchID)
	}
	err := q.Find(&stocks).Error
	return stocks, err
}

func (r *warehouseRepository) GetTransfersByBranch(branchID uuid.UUID) ([]domain.ProductTransfer, error) {
	var transfers []domain.ProductTransfer
	q := r.db.Preload("Product").Preload("ToBranch").Preload("FromFactory").Preload("FulfillmentBranch").Preload("Vehicle").Preload("Driver")
	if branchID != uuid.Nil {
		q = q.Where("to_branch_id = ? OR fulfillment_branch_id = ?", branchID, branchID)
	}
	err := q.Order("created_at desc").Limit(300).Find(&transfers).Error
	return transfers, err
}

func (r *warehouseRepository) GetPendingTransfers(branchID uuid.UUID) ([]domain.ProductTransfer, error) {
	fmt.Printf("DEBUG: [Warehouse] Fetching Pending for Branch: %s\n", branchID.String())
	var transfers []domain.ProductTransfer

	// Check count first
	var count int64
	var total int64
	r.db.Model(&domain.ProductTransfer{}).Count(&total)
	fmt.Printf("DEBUG: [Warehouse] Total Records in DB: %d\n", total)

	q := r.db.Model(&domain.ProductTransfer{}).Where("status IN ?", []string{"REQUESTED", "APPROVED", "LOADING", "SHIPPED", "ARRIVED"})
	if branchID != uuid.Nil {
		q = q.Where("to_branch_id = ? OR fulfillment_branch_id = ?", branchID, branchID)
	}
	q.Count(&count)
	fmt.Printf("DEBUG: [Warehouse] Rows found for Branch %s: %d\n", branchID.String(), count)

	q2 := r.db.Preload("Product").Preload("ToBranch").Preload("FromFactory").Preload("FulfillmentBranch").Preload("Vehicle").Preload("Driver").Where("status IN ?", []string{"REQUESTED", "APPROVED", "LOADING", "SHIPPED", "ARRIVED"})
	if branchID != uuid.Nil {
		q2 = q2.Where("to_branch_id = ? OR fulfillment_branch_id = ?", branchID, branchID)
	}
	err := q2.Order("created_at desc").Limit(200).Find(&transfers).Error

	if err != nil {
		fmt.Printf("DEBUG: [Warehouse] Query Error: %v\n", err)
		return nil, err
	}

	// Limit di atas memilih antrian terbaru, tetapi tidak boleh memotong
	// sebagian item dari SJ yang sama. Ambil semua item dengan nomor SJ terpilih.
	doNumbers := make([]string, 0, len(transfers))
	seenDOs := make(map[string]bool)
	seenIDs := make(map[uuid.UUID]bool, len(transfers))
	for _, transfer := range transfers {
		seenIDs[transfer.ID] = true
		if transfer.DeliveryOrderNo != "" && !seenDOs[transfer.DeliveryOrderNo] {
			doNumbers = append(doNumbers, transfer.DeliveryOrderNo)
			seenDOs[transfer.DeliveryOrderNo] = true
		}
	}
	if len(doNumbers) > 0 {
		var siblings []domain.ProductTransfer
		q3 := r.db.Preload("Product").Preload("ToBranch").Preload("FromFactory").Preload("FulfillmentBranch").Preload("Vehicle").Preload("Driver").
			Where("delivery_order_no IN ? AND status IN ?", doNumbers, []string{"REQUESTED", "APPROVED", "LOADING", "SHIPPED", "ARRIVED"})
		if branchID != uuid.Nil {
			q3 = q3.Where("to_branch_id = ? OR fulfillment_branch_id = ?", branchID, branchID)
		}
		if err := q3.Find(&siblings).Error; err != nil {
			return nil, err
		}
		for _, sibling := range siblings {
			if !seenIDs[sibling.ID] {
				transfers = append(transfers, sibling)
			}
		}
	}

	return transfers, err
}

func (r *warehouseRepository) GetTransferByID(id uuid.UUID) (*domain.ProductTransfer, error) {
	var transfer domain.ProductTransfer
	err := r.db.Preload("Product").Preload("FromFactory").Preload("FulfillmentBranch").Preload("Vehicle").Preload("Driver").First(&transfer, "id = ?", id).Error
	return &transfer, err
}

func (r *warehouseRepository) GetTransferByDO(doNo string) (*domain.ProductTransfer, error) {
	var transfer domain.ProductTransfer
	err := r.db.Preload("Product").Preload("FromFactory").Preload("FulfillmentBranch").Preload("Vehicle").Preload("Driver").Where("delivery_order_no = ?", doNo).First(&transfer).Error
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *warehouseRepository) UpdateTransfer(transfer *domain.ProductTransfer) error {
	return r.db.Save(transfer).Error
}

func (r *warehouseRepository) CreateLog(log *domain.WarehouseLog) error {
	return r.db.Create(log).Error
}

func (r *warehouseRepository) GetLogsByBranchID(branchID uuid.UUID) ([]domain.WarehouseLog, error) {
	var logs []domain.WarehouseLog
	q := r.db.Preload("Product")
	if branchID != uuid.Nil {
		q = q.Where("branch_id = ?", branchID)
	}
	err := q.Order("created_at desc").Limit(200).Find(&logs).Error
	return logs, err
}

// ReserveStock menambah reserved_quantity secara atomic pada batch tertentu.
func (r *warehouseRepository) ReserveStock(branchID, productID uuid.UUID, batchNo string, qty int) error {
	if batchNo == "" {
		batchNo = "DEFAULT"
	}
	// Ensure row exists
	var stock domain.WarehouseStock
	err := r.db.Where("branch_id = ? AND product_id = ? AND stock_type = 'GOOD' AND batch_no = ?", branchID, productID, batchNo).First(&stock).Error
	if err == gorm.ErrRecordNotFound {
		stock = domain.WarehouseStock{
			BranchID:         branchID,
			ProductID:        productID,
			StockType:        "GOOD",
			BatchNo:          batchNo,
			Quantity:         0,
			ReservedQuantity: 0,
		}
		if err := r.db.Create(&stock).Error; err != nil {
			return err
		}
	}

	return r.db.Model(&domain.WarehouseStock{}).
		Where("branch_id = ? AND product_id = ? AND stock_type = 'GOOD' AND batch_no = ?", branchID, productID, batchNo).
		UpdateColumn("reserved_quantity", gorm.Expr("reserved_quantity + ?", qty)).Error
}

// ReleaseStock mengurangi reserved_quantity secara atomic pada batch tertentu.
func (r *warehouseRepository) ReleaseStock(branchID, productID uuid.UUID, batchNo string, qty int) error {
	if batchNo == "" {
		batchNo = "DEFAULT"
	}
	return r.db.Model(&domain.WarehouseStock{}).
		Where("branch_id = ? AND product_id = ? AND stock_type = 'GOOD' AND batch_no = ?", branchID, productID, batchNo).
		UpdateColumn("reserved_quantity", gorm.Expr("GREATEST(reserved_quantity - ?, 0)", qty)).Error
}
