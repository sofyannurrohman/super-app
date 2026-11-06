package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type PayrollComponentRepository interface {
	Create(component *domain.PayrollComponent) error
	FindAll() ([]domain.PayrollComponent, error)
	FindByID(id uuid.UUID) (*domain.PayrollComponent, error)
	Update(component *domain.PayrollComponent) error
	Delete(id uuid.UUID) error
}

type payrollComponentRepository struct {
	db *gorm.DB
}

func NewPayrollComponentRepository(db *gorm.DB) PayrollComponentRepository {
	return &payrollComponentRepository{db}
}

func (r *payrollComponentRepository) Create(component *domain.PayrollComponent) error {
	return r.db.Create(component).Error
}

func (r *payrollComponentRepository) FindAll() ([]domain.PayrollComponent, error) {
	var components []domain.PayrollComponent
	if err := r.db.Find(&components).Error; err != nil {
		return nil, err
	}
	return components, nil
}

func (r *payrollComponentRepository) FindByID(id uuid.UUID) (*domain.PayrollComponent, error) {
	var component domain.PayrollComponent
	if err := r.db.Where("id = ?", id).First(&component).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &component, nil
}

func (r *payrollComponentRepository) Update(component *domain.PayrollComponent) error {
	return r.db.Save(component).Error
}

func (r *payrollComponentRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&domain.PayrollComponent{}).Error
}
