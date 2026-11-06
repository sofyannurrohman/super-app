package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type RawMaterialPORepository interface {
	// Purchase Order
	CreatePO(po *domain.RawMaterialPurchaseOrder) error
	GetPOByID(id uuid.UUID) (*domain.RawMaterialPurchaseOrder, error)
	ListPOs(companyID, factoryID uuid.UUID, status string) ([]domain.RawMaterialPurchaseOrder, error)
	UpdatePO(po *domain.RawMaterialPurchaseOrder) error
	DeletePO(id uuid.UUID) error
	GeneratePONumber(companyID uuid.UUID) string

	// Purchase Request
	CreatePR(pr *domain.RawMaterialPurchaseRequest) error
	GetPRByID(id uuid.UUID) (*domain.RawMaterialPurchaseRequest, error)
	ListPRs(companyID, factoryID uuid.UUID, status string) ([]domain.RawMaterialPurchaseRequest, error)
	UpdatePR(pr *domain.RawMaterialPurchaseRequest) error
	DeletePR(id uuid.UUID) error
	GeneratePRNumber(companyID uuid.UUID, factoryRequest bool) string
	UpdatePRStatus(id uuid.UUID, status domain.RawMaterialPRStatus, poID *uuid.UUID, reviewedByID *uuid.UUID, reviewedByName string, reviewNotes string, rejectionReason string) error

	// Receipts
	CreateReceipt(receipt *domain.RawMaterialReceipt) error
	GetReceiptByID(id uuid.UUID) (*domain.RawMaterialReceipt, error)
	ListReceipts(companyID, factoryID uuid.UUID) ([]domain.RawMaterialReceipt, error)
	GenerateReceiptNumber(companyID uuid.UUID) string

	// Stock Operations
	UpsertRawMaterialStock(factoryID, productID uuid.UUID, factoryWarehouseID *uuid.UUID, batchNo string, expiryDate *time.Time, deltaQty int) error
	CreateInventoryLog(log *domain.FactoryInventoryLog) error
}

type rawMaterialPORepository struct {
	db *gorm.DB
}

func NewRawMaterialPORepository(db *gorm.DB) RawMaterialPORepository {
	return &rawMaterialPORepository{db}
}

func (r *rawMaterialPORepository) CreatePO(po *domain.RawMaterialPurchaseOrder) error {
	return r.db.Create(po).Error
}

func (r *rawMaterialPORepository) GetPOByID(id uuid.UUID) (*domain.RawMaterialPurchaseOrder, error) {
	var po domain.RawMaterialPurchaseOrder
	err := r.db.
		Preload("Company").
		Preload("Factory").
		Preload("Supplier").
		Preload("Items.Product").
		Preload("Items.FactoryWarehouse").
		Preload("Receipts.Items.Product").
		Preload("Receipts.Items.FactoryWarehouse").
		First(&po, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &po, nil
}

func (r *rawMaterialPORepository) ListPOs(companyID, factoryID uuid.UUID, status string) ([]domain.RawMaterialPurchaseOrder, error) {
	var pos []domain.RawMaterialPurchaseOrder
	query := r.db.
		Preload("Factory").
		Preload("Supplier").
		Preload("Items.Product").
		Preload("Items.FactoryWarehouse").
		Order("created_at DESC")

	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}
	if factoryID != uuid.Nil {
		query = query.Where("factory_id = ?", factoryID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&pos).Error
	return pos, err
}

func (r *rawMaterialPORepository) UpdatePO(po *domain.RawMaterialPurchaseOrder) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(po).Error
}

func (r *rawMaterialPORepository) DeletePO(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var receiptIDs []uuid.UUID
		if err := tx.Model(&domain.RawMaterialReceipt{}).
			Where("purchase_order_id = ?", id).Pluck("id", &receiptIDs).Error; err != nil {
			return err
		}
		if len(receiptIDs) > 0 {
			if err := tx.Where("receipt_id IN ?", receiptIDs).Delete(&domain.RawMaterialReceiptItem{}).Error; err != nil {
				return err
			}
			if err := tx.Where("purchase_order_id = ?", id).Delete(&domain.RawMaterialReceipt{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("purchase_order_id = ?", id).Delete(&domain.RawMaterialPurchaseOrderItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.RawMaterialPurchaseOrder{}, "id = ?", id).Error
	})
}

func (r *rawMaterialPORepository) GeneratePONumber(companyID uuid.UUID) string {
	now := time.Now()
	prefix := fmt.Sprintf("PO-BB-%s-", now.Format("20060102"))
	var count int64
	r.db.Model(&domain.RawMaterialPurchaseOrder{}).
		Where("po_number LIKE ?", prefix+"%").
		Count(&count)
	return fmt.Sprintf("%s%03d", prefix, count+1)
}

func (r *rawMaterialPORepository) CreatePR(pr *domain.RawMaterialPurchaseRequest) error {
	return r.db.Create(pr).Error
}

func (r *rawMaterialPORepository) GetPRByID(id uuid.UUID) (*domain.RawMaterialPurchaseRequest, error) {
	var pr domain.RawMaterialPurchaseRequest
	err := r.db.
		Preload("Company").
		Preload("Factory").
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Supplier").
		Preload("Items.Product").
		Preload("Items.FactoryWarehouse").
		First(&pr, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *rawMaterialPORepository) ListPRs(companyID, factoryID uuid.UUID, status string) ([]domain.RawMaterialPurchaseRequest, error) {
	var prs []domain.RawMaterialPurchaseRequest
	query := r.db.
		Preload("Factory").
		Preload("PurchaseOrder").
		Preload("Items.Product").
		Preload("Items.FactoryWarehouse").
		Order("created_at DESC")

	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}
	if factoryID != uuid.Nil {
		query = query.Where("factory_id = ?", factoryID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&prs).Error
	return prs, err
}

func (r *rawMaterialPORepository) UpdatePR(pr *domain.RawMaterialPurchaseRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("purchase_request_id = ?", pr.ID).Delete(&domain.RawMaterialPurchaseRequestItem{}).Error; err != nil {
			return err
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(pr).Error
	})
}

func (r *rawMaterialPORepository) DeletePR(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("purchase_request_id = ?", id).Delete(&domain.RawMaterialPurchaseRequestItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.RawMaterialPurchaseRequest{}, "id = ?", id).Error
	})
}

func (r *rawMaterialPORepository) GeneratePRNumber(companyID uuid.UUID, factoryRequest bool) string {
	now := time.Now()
	prefix := fmt.Sprintf("PR-%s-", now.Format("20060102"))
	if factoryRequest {
		prefix = fmt.Sprintf("PR-BB-%s-", now.Format("20060102"))
	}
	var count int64
	r.db.Model(&domain.RawMaterialPurchaseRequest{}).
		Where("pr_number LIKE ?", prefix+"%").
		Count(&count)
	return fmt.Sprintf("%s%03d", prefix, count+1)
}

func (r *rawMaterialPORepository) UpdatePRStatus(id uuid.UUID, status domain.RawMaterialPRStatus, poID *uuid.UUID, reviewedByID *uuid.UUID, reviewedByName string, reviewNotes string, rejectionReason string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if poID != nil {
		updates["purchase_order_id"] = poID
	}
	if reviewedByID != nil {
		updates["reviewed_by_id"] = reviewedByID
		updates["reviewed_by_name"] = reviewedByName
		now := time.Now()
		updates["reviewed_at"] = &now
	}
	if reviewNotes != "" {
		updates["review_notes"] = reviewNotes
	}
	if rejectionReason != "" {
		updates["rejection_reason"] = rejectionReason
	}
	return r.db.Model(&domain.RawMaterialPurchaseRequest{}).Where("id = ?", id).Updates(updates).Error
}

func (r *rawMaterialPORepository) CreateReceipt(receipt *domain.RawMaterialReceipt) error {
	return r.db.Create(receipt).Error
}

func (r *rawMaterialPORepository) GetReceiptByID(id uuid.UUID) (*domain.RawMaterialReceipt, error) {
	var receipt domain.RawMaterialReceipt
	err := r.db.
		Preload("Company").
		Preload("Factory").
		Preload("Supplier").
		Preload("PurchaseOrder").
		Preload("KasPabrik").
		Preload("Items.Product").
		Preload("Items.TechnicianItem").
		Preload("Items.FactoryWarehouse").
		First(&receipt, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *rawMaterialPORepository) ListReceipts(companyID, factoryID uuid.UUID) ([]domain.RawMaterialReceipt, error) {
	var receipts []domain.RawMaterialReceipt
	query := r.db.
		Preload("Factory").
		Preload("Supplier").
		Preload("PurchaseOrder").
		Preload("KasPabrik").
		Preload("Items.Product").
		Preload("Items.TechnicianItem").
		Preload("Items.FactoryWarehouse").
		Order("received_date DESC, created_at DESC")

	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}
	if factoryID != uuid.Nil {
		query = query.Where("factory_id = ?", factoryID)
	}

	err := query.Find(&receipts).Error
	return receipts, err
}

func (r *rawMaterialPORepository) GenerateReceiptNumber(companyID uuid.UUID) string {
	now := time.Now()
	prefix := fmt.Sprintf("RCV-BB-%s-", now.Format("20060102"))
	var count int64
	r.db.Model(&domain.RawMaterialReceipt{}).
		Where("receipt_no LIKE ?", prefix+"%").
		Count(&count)
	return fmt.Sprintf("%s%03d", prefix, count+1)
}

func (r *rawMaterialPORepository) UpsertRawMaterialStock(factoryID, productID uuid.UUID, factoryWarehouseID *uuid.UUID, batchNo string, expiryDate *time.Time, deltaQty int) error {
	if batchNo == "" {
		batchNo = "DEFAULT"
	}

	var stock domain.FactoryStock
	query := r.db.Where("factory_id = ? AND product_id = ? AND batch_no = ?", factoryID, productID, batchNo)
	if factoryWarehouseID != nil && *factoryWarehouseID != uuid.Nil {
		query = query.Where("factory_warehouse_id = ?", factoryWarehouseID)
	}

	err := query.First(&stock).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			newStock := domain.FactoryStock{
				FactoryID:          factoryID,
				ProductID:          productID,
				FactoryWarehouseID: factoryWarehouseID,
				BatchNo:            batchNo,
				ExpiryDate:         expiryDate,
				Quantity:           deltaQty,
			}
			return r.db.Create(&newStock).Error
		}
		return err
	}

	stock.Quantity += deltaQty
	if expiryDate != nil {
		stock.ExpiryDate = expiryDate
	}
	if factoryWarehouseID != nil && *factoryWarehouseID != uuid.Nil {
		stock.FactoryWarehouseID = factoryWarehouseID
	}
	return r.db.Save(&stock).Error
}

func (r *rawMaterialPORepository) CreateInventoryLog(log *domain.FactoryInventoryLog) error {
	return r.db.Create(log).Error
}
