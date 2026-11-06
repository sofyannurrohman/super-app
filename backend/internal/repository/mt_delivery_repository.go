package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type MTDeliveryRepository interface {
	// MT Delivery Order
	Create(order *domain.MTDeliveryOrder) error
	GetByID(id uuid.UUID) (*domain.MTDeliveryOrder, error)
	ListByCompany(companyID uuid.UUID) ([]domain.MTDeliveryOrder, error)
	ListByBranch(branchID uuid.UUID) ([]domain.MTDeliveryOrder, error)
	Update(order *domain.MTDeliveryOrder) error
	Delete(id uuid.UUID) error
	GenerateDONumber(companyID uuid.UUID) string
}

type mtDeliveryRepository struct {
	db *gorm.DB
}

func NewMTDeliveryRepository(db *gorm.DB) MTDeliveryRepository {
	return &mtDeliveryRepository{db}
}

func (r *mtDeliveryRepository) preload(q *gorm.DB) *gorm.DB {
	return q.
		Preload("KeyAccount").
		Preload("SourceBranch").
		Preload("SourceFactory").
		Preload("Vehicle").
		Preload("Driver").
		Preload("SalesOrder").
		Preload("SalesOrder.BonusItems").
		Preload("SalesOrder.BonusItems.BonusInventory").
		Preload("Items").
		Preload("Items.Product")
}

func (r *mtDeliveryRepository) Create(order *domain.MTDeliveryOrder) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(order).Error
	})
}

func (r *mtDeliveryRepository) GetByID(id uuid.UUID) (*domain.MTDeliveryOrder, error) {
	var order domain.MTDeliveryOrder
	err := r.preload(r.db).First(&order, "id = ?", id).Error
	return &order, err
}

func (r *mtDeliveryRepository) ListByCompany(companyID uuid.UUID) ([]domain.MTDeliveryOrder, error) {
	var orders []domain.MTDeliveryOrder
	err := r.preload(r.db).
		Joins("LEFT JOIN key_accounts ON key_accounts.id = mt_delivery_orders.key_account_id").
		Joins("LEFT JOIN sales_orders ON sales_orders.id = mt_delivery_orders.sales_order_id").
		Where("mt_delivery_orders.company_id = ? OR key_accounts.company_id = ? OR sales_orders.company_id = ? OR mt_delivery_orders.key_account_id IS NOT NULL", companyID, companyID, companyID).
		Order("mt_delivery_orders.created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *mtDeliveryRepository) ListByBranch(branchID uuid.UUID) ([]domain.MTDeliveryOrder, error) {
	var orders []domain.MTDeliveryOrder
	query := r.preload(r.db).Where("status NOT IN (?)", []string{string(domain.MTDeliveryDraft)})
	if branchID != uuid.Nil {
		query = query.Where("source_branch_id = ?", branchID)
	}
	err := query.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *mtDeliveryRepository) Update(order *domain.MTDeliveryOrder) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete and re-insert items
		if len(order.Items) > 0 {
			if err := tx.Where("mt_delivery_order_id = ?", order.ID).Delete(&domain.MTDeliveryOrderItem{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}

		// Sync status to SalesOrder if linked
		if order.SalesOrderID != nil {
			var soStatus string
			switch order.Status {
			case domain.MTDeliveryApproved, domain.MTDeliveryLoading:
				soStatus = string(domain.SOStatusApproved)
			case domain.MTDeliveryShipped:
				soStatus = string(domain.SOStatusInDelivery)
			case domain.MTDeliveryDelivered:
				soStatus = string(domain.SOStatusDelivered)
			case domain.MTDeliveryRejected, domain.MTDeliveryCancelled:
				soStatus = string(domain.SOStatusRejected)
			}
			if soStatus != "" {
				if err := tx.Model(&domain.SalesOrder{}).Where("id = ?", order.SalesOrderID).Update("status", soStatus).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *mtDeliveryRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("mt_delivery_order_id = ?", id).Delete(&domain.MTDeliveryOrderItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.MTDeliveryOrder{}, "id = ?", id).Error
	})
}

func (r *mtDeliveryRepository) GenerateDONumber(companyID uuid.UUID) string {
	now := time.Now()
	return fmt.Sprintf("MT-DO-%s-%04d", now.Format("20060102"), now.Unix()%10000)
}
