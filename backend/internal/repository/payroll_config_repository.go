package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PayrollConfigRepository interface {
	GetByCompanyID(companyID uuid.UUID) (*domain.PayrollConfig, error)
	Upsert(config *domain.PayrollConfig) error
}

type payrollConfigRepository struct {
	db *gorm.DB
}

func NewPayrollConfigRepository(db *gorm.DB) PayrollConfigRepository {
	return &payrollConfigRepository{db}
}

func (r *payrollConfigRepository) GetByCompanyID(companyID uuid.UUID) (*domain.PayrollConfig, error) {
	var config domain.PayrollConfig
	err := r.db.Where("company_id = ?", companyID).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *payrollConfigRepository) Upsert(config *domain.PayrollConfig) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "company_id"}},
		UpdateAll: true,
	}).Create(config).Error
}
