package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type SalesReturnRepository interface {
	Create(ret *domain.SalesReturn) error
	Update(ret *domain.SalesReturn) error
	FindByID(id uuid.UUID) (*domain.SalesReturn, error)
	FindByBranch(branchID uuid.UUID, status string) ([]domain.SalesReturn, error)
	CountByBranchAndDate(branchID uuid.UUID, date time.Time) (int64, error)
	GenerateNextReturnNumber(date time.Time) (string, error)
	Delete(id uuid.UUID) error
}

type salesReturnRepository struct {
	db *gorm.DB
}

func NewSalesReturnRepository(db *gorm.DB) SalesReturnRepository {
	return &salesReturnRepository{db}
}

func (r *salesReturnRepository) Create(ret *domain.SalesReturn) error {
	return r.db.Create(ret).Error
}

func (r *salesReturnRepository) Update(ret *domain.SalesReturn) error {
	return r.db.Save(ret).Error
}

func (r *salesReturnRepository) FindByID(id uuid.UUID) (*domain.SalesReturn, error) {
	var ret domain.SalesReturn
	err := r.db.
		Preload("Items.Product").
		Preload("Items.ReplacementProduct").
		Preload("SalesOrder").
		Preload("SalesOrder.Store").
		Preload("SalesOrder.DeliveryBatch").
		Preload("SalesOrder.DeliveryBatch.Driver").
		Preload("SalesOrder.DeliveryBatch.DailyDriver").
		Preload("SalesOrder.DeliveryBatch.Vehicle").
		Preload("Transaction").
		Preload("Transaction.Store").
		Preload("Employee").
		Preload("Employee.JobPosition").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		First(&ret, "id = ?", id).Error
	return &ret, err
}

func (r *salesReturnRepository) FindByBranch(branchID uuid.UUID, status string) ([]domain.SalesReturn, error) {
	var returns []domain.SalesReturn
	q := r.db.
		Preload("Items.Product").
		Preload("Items.ReplacementProduct").
		Preload("SalesOrder").
		Preload("SalesOrder.Store").
		Preload("SalesOrder.DeliveryBatch").
		Preload("SalesOrder.DeliveryBatch.Driver").
		Preload("SalesOrder.DeliveryBatch.DailyDriver").
		Preload("SalesOrder.DeliveryBatch.Vehicle").
		Preload("Transaction").
		Preload("Transaction.Store").
		Preload("Employee").
		Preload("Employee.JobPosition").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Where("branch_id = ?", branchID)
	if status != "" && status != "ALL" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at desc").Find(&returns).Error
	return returns, err
}

func (r *salesReturnRepository) CountByBranchAndDate(branchID uuid.UUID, date time.Time) (int64, error) {
	var count int64
	dateStr := date.Format("2006-01-02")
	err := r.db.Model(&domain.SalesReturn{}).
		Where("branch_id = ? AND created_at::date = ?", branchID, dateStr).
		Count(&count).Error
	return count, err
}

func (r *salesReturnRepository) GenerateNextReturnNumber(date time.Time) (string, error) {
	prefix := fmt.Sprintf("RET-%s-", date.Format("20060102"))
	var lastNo string
	err := r.db.Model(&domain.SalesReturn{}).
		Where("return_no LIKE ?", prefix+"%").
		Order("return_no DESC").
		Limit(1).
		Pluck("return_no", &lastNo).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}

	seq := 1
	if lastNo != "" {
		var prevSeq int
		if _, scanErr := fmt.Sscanf(lastNo, prefix+"%d", &prevSeq); scanErr == nil {
			seq = prevSeq + 1
		}
	}
	return fmt.Sprintf("%s%04d", prefix, seq), nil
}

func (r *salesReturnRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.SalesReturn{}, "id = ?", id).Error
}

func GenerateReturnNumber(branchID uuid.UUID, date time.Time, count int64) string {
	return fmt.Sprintf("RET-%s-%04d", date.Format("20060102"), count+1)
}
