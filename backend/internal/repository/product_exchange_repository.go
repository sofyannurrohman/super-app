package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type ProductExchangeRepository interface {
	Create(pe *domain.ProductExchange) error
	Update(pe *domain.ProductExchange) error
	FindByID(id uuid.UUID) (*domain.ProductExchange, error)
	FindByBranch(branchID uuid.UUID, status string) ([]domain.ProductExchange, error)
	CountByBranchAndDate(branchID uuid.UUID, date time.Time) (int64, error)
	GenerateNextExchangeNumber(date time.Time) (string, error)
}

type productExchangeRepository struct {
	db *gorm.DB
}

func NewProductExchangeRepository(db *gorm.DB) ProductExchangeRepository {
	return &productExchangeRepository{db}
}

func (r *productExchangeRepository) Create(pe *domain.ProductExchange) error {
	return r.db.Create(pe).Error
}

func (r *productExchangeRepository) Update(pe *domain.ProductExchange) error {
	return r.db.Save(pe).Error
}

func (r *productExchangeRepository) FindByID(id uuid.UUID) (*domain.ProductExchange, error) {
	var pe domain.ProductExchange
	err := r.db.
		Preload("Store").
		Preload("Branch").
		Preload("Employee").
		Preload("Employee.JobPosition").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Preload("Items.Product").
		First(&pe, "id = ?", id).Error
	return &pe, err
}

func (r *productExchangeRepository) FindByBranch(branchID uuid.UUID, status string) ([]domain.ProductExchange, error) {
	var exchanges []domain.ProductExchange
	q := r.db.
		Preload("Store").
		Preload("Employee").
		Preload("Employee.JobPosition").
		Preload("ApprovedBy").
		Preload("ApprovedBy.JobPosition").
		Preload("Items.Product").
		Where("branch_id = ?", branchID)
	if status != "" && status != "ALL" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at desc").Find(&exchanges).Error
	return exchanges, err
}

func (r *productExchangeRepository) CountByBranchAndDate(branchID uuid.UUID, date time.Time) (int64, error) {
	var count int64
	dateStr := date.Format("2006-01-02")
	err := r.db.Model(&domain.ProductExchange{}).
		Where("branch_id = ? AND created_at::date = ?", branchID, dateStr).
		Count(&count).Error
	return count, err
}

func (r *productExchangeRepository) GenerateNextExchangeNumber(date time.Time) (string, error) {
	prefix := fmt.Sprintf("EXC-%s-", date.Format("20060102"))
	var lastNo string
	err := r.db.Model(&domain.ProductExchange{}).
		Where("exchange_no LIKE ?", prefix+"%").
		Order("exchange_no DESC").
		Limit(1).
		Pluck("exchange_no", &lastNo).Error

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

func GenerateExchangeNumber(branchID uuid.UUID, date time.Time, count int64) string {
	return fmt.Sprintf("EXC-%s-%04d", date.Format("20060102"), count+1)
}
