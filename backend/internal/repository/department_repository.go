package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type DepartmentRepository interface {
	Create(department *domain.Department) error
	FindAll() ([]domain.Department, error)
	FindByID(id uuid.UUID) (*domain.Department, error)
	Update(department *domain.Department) error
	Delete(id uuid.UUID) error
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db}
}

func (r *departmentRepository) Create(department *domain.Department) error {
	return r.db.Create(department).Error
}

func (r *departmentRepository) FindAll() ([]domain.Department, error) {
	var departments []domain.Department
	if err := r.db.Preload("ParentDepartment").Find(&departments).Error; err != nil {
		return nil, err
	}
	return departments, nil
}

func (r *departmentRepository) FindByID(id uuid.UUID) (*domain.Department, error) {
	var department domain.Department
	if err := r.db.Preload("ParentDepartment").Where("id = ?", id).First(&department).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &department, nil
}

func (r *departmentRepository) Update(department *domain.Department) error {
	return r.db.Save(department).Error
}

func (r *departmentRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&domain.Department{}).Error
}
