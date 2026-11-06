package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type WarehouseDeliveryOrderRepository interface {
	Create(wdo *domain.WarehouseDeliveryOrder) error
	Update(wdo *domain.WarehouseDeliveryOrder) error
	FindByID(id uuid.UUID) (*domain.WarehouseDeliveryOrder, error)
	FindByBatchID(batchID uuid.UUID) (*domain.WarehouseDeliveryOrder, error)
	CountByDate(date time.Time) (int64, error)
	ListMotoris(branchID *uuid.UUID) ([]domain.WarehouseDeliveryOrder, error)
}

type warehouseDeliveryOrderRepository struct {
	db *gorm.DB
}

func NewWarehouseDeliveryOrderRepository(db *gorm.DB) WarehouseDeliveryOrderRepository {
	return &warehouseDeliveryOrderRepository{db}
}

func (r *warehouseDeliveryOrderRepository) Create(wdo *domain.WarehouseDeliveryOrder) error {
	return r.db.Create(wdo).Error
}

func (r *warehouseDeliveryOrderRepository) Update(wdo *domain.WarehouseDeliveryOrder) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("warehouse_delivery_order_id = ?", wdo.ID).Delete(&domain.WarehouseDeliveryOrderItem{}).Error; err != nil {
			return err
		}
		return tx.Save(wdo).Error
	})
}

func (r *warehouseDeliveryOrderRepository) FindByID(id uuid.UUID) (*domain.WarehouseDeliveryOrder, error) {
	var wdo domain.WarehouseDeliveryOrder
	err := r.db.
		Preload("Items.Product").
		Preload("Items.BonusInventory").
		Preload("DeliveryBatch").
		Preload("Employee").
		Preload("SalesTransfers.SalesTransfer.Product").
		Preload("SalesTransfers.SalesBonusRequest.BonusInventory").
		First(&wdo, "id = ?", id).Error
	return &wdo, err
}

func (r *warehouseDeliveryOrderRepository) FindByBatchID(batchID uuid.UUID) (*domain.WarehouseDeliveryOrder, error) {
	var wdo domain.WarehouseDeliveryOrder
	err := r.db.
		Preload("Items.Product").
		Preload("Items.BonusInventory").
		Preload("DeliveryBatch").
		Preload("Employee").
		Preload("SalesTransfers.SalesTransfer.Product").
		Preload("SalesTransfers.SalesBonusRequest.BonusInventory").
		First(&wdo, "delivery_batch_id = ?", batchID).Error
	return &wdo, err
}

func (r *warehouseDeliveryOrderRepository) CountByDate(date time.Time) (int64, error) {
	var count int64
	dateStr := date.Format("2006-01-02")
	err := r.db.Model(&domain.WarehouseDeliveryOrder{}).
		Where("created_at::date = ?", dateStr).
		Count(&count).Error
	return count, err
}

// ListMotoris mengembalikan semua DO Motoris (standalone, tanpa SJ) untuk halaman Kelola DO Spares
func (r *warehouseDeliveryOrderRepository) ListMotoris(branchID *uuid.UUID) ([]domain.WarehouseDeliveryOrder, error) {
	var wdos []domain.WarehouseDeliveryOrder
	db := r.db.
		Preload("Items.Product").
		Preload("Items.BonusInventory").
		Preload("Employee").
		Preload("SalesTransfers.SalesTransfer.Product").
		Preload("SalesTransfers.SalesBonusRequest.BonusInventory").
		Where("do_type = ?", domain.DOTypeMotoris)

	if branchID != nil && *branchID != uuid.Nil {
		db = db.Where("EXISTS (SELECT 1 FROM warehouse_delivery_order_transfers wdot JOIN sales_transfers st ON st.id = wdot.sales_transfer_id WHERE wdot.warehouse_delivery_order_id = warehouse_delivery_orders.id AND st.branch_id = ?)", *branchID)
	}

	err := db.Order("created_at DESC").Find(&wdos).Error
	return wdos, err
}

func GenerateWarehouseDONumber(date time.Time, count int64) string {
	return fmt.Sprintf("WDO-%s-%04d", date.Format("20060102"), count+1)
}
