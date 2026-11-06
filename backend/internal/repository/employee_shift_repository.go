package repository

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type EmployeeShiftRepository interface {
	Create(es *domain.EmployeeShift) error
	FindAll() ([]domain.EmployeeShift, error)
	FindByEmployeeID(employeeID uuid.UUID) ([]domain.EmployeeShift, error)
	FindByEmployeeIDAndDateRange(employeeID uuid.UUID, startDate, endDate time.Time) ([]domain.EmployeeShift, error)
	FindByID(id uuid.UUID) (*domain.EmployeeShift, error)
	Update(es *domain.EmployeeShift) error
	Delete(id uuid.UUID) error
}

type employeeShiftRepository struct {
	db *gorm.DB
}

func NewEmployeeShiftRepository(db *gorm.DB) EmployeeShiftRepository {
	return &employeeShiftRepository{db}
}

func (r *employeeShiftRepository) Create(es *domain.EmployeeShift) error {
	return r.db.Create(es).Error
}

func (r *employeeShiftRepository) FindAll() ([]domain.EmployeeShift, error) {
	var list []domain.EmployeeShift
	if err := r.db.Preload("Employee").Preload("Shift").Order("date desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *employeeShiftRepository) FindByEmployeeID(employeeID uuid.UUID) ([]domain.EmployeeShift, error) {
	var list []domain.EmployeeShift
	if err := r.db.Preload("Shift").Where("employee_id = ?", employeeID).Order("date desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *employeeShiftRepository) FindByEmployeeIDAndDateRange(employeeID uuid.UUID, startDate, endDate time.Time) ([]domain.EmployeeShift, error) {
	var list []domain.EmployeeShift
	if err := r.db.Preload("Shift").
		Where("employee_id = ? AND date >= ? AND date <= ?", employeeID, startDate, endDate).
		Order("date asc").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *employeeShiftRepository) FindByID(id uuid.UUID) (*domain.EmployeeShift, error) {
	var es domain.EmployeeShift
	if err := r.db.Preload("Employee").Preload("Shift").Where("id = ?", id).First(&es).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &es, nil
}

func (r *employeeShiftRepository) Update(es *domain.EmployeeShift) error {
	return r.db.Save(es).Error
}

func (r *employeeShiftRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&domain.EmployeeShift{}).Error
}
