package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type SalesTesterRequestRepository interface {
	Create(req *domain.SalesTesterRequest) error
	Update(req *domain.SalesTesterRequest) error
	FindByID(id uuid.UUID) (*domain.SalesTesterRequest, error)
	FindAll(branchID *uuid.UUID, status string) ([]domain.SalesTesterRequest, error)
	FindByEmployee(employeeID uuid.UUID) ([]domain.SalesTesterRequest, error)
	Delete(id uuid.UUID) error
}

type salesTesterRequestRepository struct {
	db *gorm.DB
}

func NewSalesTesterRequestRepository(db *gorm.DB) SalesTesterRequestRepository {
	return &salesTesterRequestRepository{db}
}

func (r *salesTesterRequestRepository) Create(req *domain.SalesTesterRequest) error {
	return r.db.Create(req).Error
}

func (r *salesTesterRequestRepository) Update(req *domain.SalesTesterRequest) error {
	return r.db.Save(req).Error
}

func (r *salesTesterRequestRepository) FindByID(id uuid.UUID) (*domain.SalesTesterRequest, error) {
	var req domain.SalesTesterRequest
	err := r.db.
		Preload("Branch").
		Preload("Employee").
		Preload("Store").
		Preload("Product").
		Preload("DeliveryBatch").
		First(&req, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *salesTesterRequestRepository) FindAll(branchID *uuid.UUID, status string) ([]domain.SalesTesterRequest, error) {
	var requests []domain.SalesTesterRequest
	q := r.db.
		Preload("Branch").
		Preload("Employee").
		Preload("Store").
		Preload("Product").
		Preload("DeliveryBatch")

	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}

	if status != "" && status != "ALL" {
		q = q.Where("status = ?", status)
	}

	err := q.Order("created_at desc").Find(&requests).Error
	return requests, err
}

func (r *salesTesterRequestRepository) FindByEmployee(employeeID uuid.UUID) ([]domain.SalesTesterRequest, error) {
	var requests []domain.SalesTesterRequest
	err := r.db.
		Preload("Branch").
		Preload("Employee").
		Preload("Store").
		Preload("Product").
		Preload("DeliveryBatch").
		Where("employee_id = ?", employeeID).
		Order("created_at desc").
		Find(&requests).Error
	return requests, err
}

func (r *salesTesterRequestRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.SalesTesterRequest{}, "id = ?", id).Error
}
