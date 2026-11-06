package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type SupplierRepository interface {
	// Suppliers
	CreateSupplier(supplier *domain.Supplier) error
	BulkCreateSuppliers(suppliers []domain.Supplier) error
	GetSupplierByID(id uuid.UUID) (*domain.Supplier, error)
	ListSuppliers(companyID uuid.UUID) ([]domain.Supplier, error)
	UpdateSupplier(supplier *domain.Supplier) error
	DeleteSupplier(id uuid.UUID) error

	// Supplier Receipts
	CreateSupplierReceipt(receipt *domain.SupplierReceipt) error
	GetSupplierReceiptByID(id uuid.UUID) (*domain.SupplierReceipt, error)
	ListSupplierReceipts(branchID uuid.UUID) ([]domain.SupplierReceipt, error)
	UpdateSupplierReceipt(receipt *domain.SupplierReceipt) error
	DeleteSupplierReceipt(id uuid.UUID) error
	CreateSupplierPayment(payment *domain.SupplierPayment) error

	// Supplier Stock
	UpsertSupplierStock(branchID, supplierID, productID uuid.UUID, delta int) error
	GetSupplierStock(branchID uuid.UUID) ([]domain.SupplierStock, error)
	GenerateReceiptNo(branchID uuid.UUID) string
}

type supplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) SupplierRepository {
	return &supplierRepository{db}
}

func (r *supplierRepository) CreateSupplier(supplier *domain.Supplier) error {
	return r.db.Create(supplier).Error
}

func (r *supplierRepository) BulkCreateSuppliers(suppliers []domain.Supplier) error {
	if len(suppliers) == 0 {
		return nil
	}
	return r.db.Create(&suppliers).Error
}

func (r *supplierRepository) GetSupplierByID(id uuid.UUID) (*domain.Supplier, error) {
	var supplier domain.Supplier
	err := r.db.First(&supplier, "id = ?", id).Error
	return &supplier, err
}

func (r *supplierRepository) ListSuppliers(companyID uuid.UUID) ([]domain.Supplier, error) {
	var suppliers []domain.Supplier
	query := r.db.Order("name ASC")
	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Find(&suppliers).Error
	return suppliers, err
}

func (r *supplierRepository) UpdateSupplier(supplier *domain.Supplier) error {
	return r.db.Save(supplier).Error
}

func (r *supplierRepository) DeleteSupplier(id uuid.UUID) error {
	var purchaseOrderCount int64
	if err := r.db.Table("raw_material_purchase_orders").Where("supplier_id = ?", id).Count(&purchaseOrderCount).Error; err != nil {
		return err
	}
	if purchaseOrderCount > 0 {
		return domain.ErrSupplierInUse
	}

	var rawReceiptCount int64
	if err := r.db.Table("raw_material_receipts").Where("supplier_id = ?", id).Count(&rawReceiptCount).Error; err != nil {
		return err
	}
	if rawReceiptCount > 0 {
		return domain.ErrSupplierInUse
	}

	var supplierReceiptCount int64
	if err := r.db.Table("supplier_receipts").Where("supplier_id = ?", id).Count(&supplierReceiptCount).Error; err != nil {
		return err
	}
	if supplierReceiptCount > 0 {
		return domain.ErrSupplierInUse
	}

	return r.db.Delete(&domain.Supplier{}, "id = ?", id).Error
}

func (r *supplierRepository) CreateSupplierReceipt(receipt *domain.SupplierReceipt) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(receipt).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *supplierRepository) GetSupplierReceiptByID(id uuid.UUID) (*domain.SupplierReceipt, error) {
	var receipt domain.SupplierReceipt
	err := r.db.
		Preload("Supplier").
		Preload("Branch").
		Preload("Items").
		Preload("Items.Product").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("payment_date DESC, created_at DESC")
		}).
		First(&receipt, "id = ?", id).Error
	return &receipt, err
}

func (r *supplierRepository) ListSupplierReceipts(branchID uuid.UUID) ([]domain.SupplierReceipt, error) {
	var receipts []domain.SupplierReceipt
	query := r.db.
		Preload("Supplier").
		Preload("Branch").
		Preload("Items").
		Preload("Items.Product").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("payment_date DESC, created_at DESC")
		}).
		Order("created_at DESC")
	if branchID != uuid.Nil {
		query = query.Where("branch_id = ?", branchID)
	}
	err := query.Find(&receipts).Error
	return receipts, err
}

func (r *supplierRepository) UpdateSupplierReceipt(receipt *domain.SupplierReceipt) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing items and re-insert
		if err := tx.Where("supplier_receipt_id = ?", receipt.ID).Delete(&domain.SupplierReceiptItem{}).Error; err != nil {
			return err
		}
		if err := tx.Save(receipt).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *supplierRepository) DeleteSupplierReceipt(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("supplier_receipt_id = ?", id).Delete(&domain.SupplierReceiptItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.SupplierReceipt{}, "id = ?", id).Error
	})
}

func (r *supplierRepository) CreateSupplierPayment(payment *domain.SupplierPayment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var receipt domain.SupplierReceipt
		if err := tx.First(&receipt, "id = ?", payment.SupplierReceiptID).Error; err != nil {
			return err
		}
		if payment.Amount <= 0 {
			return fmt.Errorf("jumlah pembayaran harus lebih dari nol")
		}
		if receipt.InvoiceTotal <= 0 {
			return fmt.Errorf("total faktur belum tersedia")
		}
		if receipt.PaidAmount+payment.Amount > receipt.InvoiceTotal {
			return fmt.Errorf("pembayaran melebihi sisa hutang")
		}
		if err := tx.Create(payment).Error; err != nil {
			return err
		}
		receipt.PaidAmount += payment.Amount
		receipt.PaymentStatus = domain.SupplierInvoicePartial
		if receipt.PaidAmount >= receipt.InvoiceTotal {
			receipt.PaymentStatus = domain.SupplierInvoicePaid
		}
		return tx.Save(&receipt).Error
	})
}

func (r *supplierRepository) UpsertSupplierStock(branchID, supplierID, productID uuid.UUID, delta int) error {
	var stock domain.SupplierStock
	err := r.db.Where("branch_id = ? AND supplier_id = ? AND product_id = ?", branchID, supplierID, productID).First(&stock).Error
	if err == gorm.ErrRecordNotFound {
		// Create new
		newQty := delta
		if newQty < 0 {
			newQty = 0
		}
		stock = domain.SupplierStock{
			BranchID:   branchID,
			SupplierID: supplierID,
			ProductID:  productID,
			Quantity:   newQty,
		}
		return r.db.Create(&stock).Error
	} else if err != nil {
		return err
	}
	// Update existing
	newQty := stock.Quantity + delta
	if newQty < 0 {
		newQty = 0
	}
	stock.Quantity = newQty
	stock.UpdatedAt = time.Now()
	return r.db.Save(&stock).Error
}

func (r *supplierRepository) GetSupplierStock(branchID uuid.UUID) ([]domain.SupplierStock, error) {
	var stocks []domain.SupplierStock
	query := r.db.
		Preload("Branch").
		Preload("Supplier").
		Preload("Product").
		Where("quantity > 0")
	if branchID != uuid.Nil {
		query = query.Where("branch_id = ?", branchID)
	}
	err := query.Find(&stocks).Error
	return stocks, err
}

func (r *supplierRepository) GenerateReceiptNo(branchID uuid.UUID) string {
	now := time.Now()
	shortID := branchID.String()[:4]
	return fmt.Sprintf("SR-%s-%s-%04d",
		shortID,
		now.Format("20060102"),
		now.Unix()%10000,
	)
}
