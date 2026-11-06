package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type EpkwtRepository interface {
	Create(req *domain.EpkwtRequest) error
	Update(req *domain.EpkwtRequest) error
	FindByID(id uuid.UUID) (*domain.EpkwtRequest, error)
	FindAll(employeeID *uuid.UUID) ([]domain.EpkwtRequest, error)
	CountByYear(year int) (int64, error)
	Delete(id uuid.UUID) error
}

type epkwtRepository struct {
	db *gorm.DB
}

func NewEpkwtRepository(db *gorm.DB) EpkwtRepository {
	return &epkwtRepository{db: db}
}

func (r *epkwtRepository) Create(req *domain.EpkwtRequest) error {
	return r.db.Create(req).Error
}

func (r *epkwtRepository) Update(req *domain.EpkwtRequest) error {
	return r.db.Save(req).Error
}

func (r *epkwtRepository) FindByID(id uuid.UUID) (*domain.EpkwtRequest, error) {
	var req domain.EpkwtRequest
	err := r.db.Preload("Employee").
		Preload("JobPosition").
		Preload("Department").
		First(&req, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *epkwtRepository) FindAll(employeeID *uuid.UUID) ([]domain.EpkwtRequest, error) {
	var reqs []domain.EpkwtRequest
	query := r.db.Preload("Employee").
		Preload("JobPosition").
		Preload("Department")
	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}
	err := query.Order("created_at DESC").Find(&reqs).Error
	return reqs, err
}

func (r *epkwtRepository) CountByYear(year int) (int64, error) {
	var count int64
	err := r.db.Model(&domain.EpkwtRequest{}).
		Where("EXTRACT(YEAR FROM created_at) = ?", year).
		Count(&count).Error
	return count, err
}

func (r *epkwtRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.EpkwtRequest{}, "id = ?", id).Error
}
