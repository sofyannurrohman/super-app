package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type PayslipRequestRepository interface {
	Create(req *domain.PayslipRequest) error
	GetByUserID(userID uuid.UUID, limit, offset int) ([]domain.PayslipRequest, int64, error)
	GetAll(status string, limit, offset int) ([]domain.PayslipRequest, int64, error)
	GetByID(id uuid.UUID) (*domain.PayslipRequest, error)
	Update(req *domain.PayslipRequest) error
}

type payslipRequestRepository struct {
	db *gorm.DB
}

func NewPayslipRequestRepository(db *gorm.DB) PayslipRequestRepository {
	return &payslipRequestRepository{db}
}

func (r *payslipRequestRepository) Create(req *domain.PayslipRequest) error {
	return r.db.Create(req).Error
}

func (r *payslipRequestRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]domain.PayslipRequest, int64, error) {
	var reqs []domain.PayslipRequest
	var total int64

	query := r.db.Model(&domain.PayslipRequest{}).Where("user_id = ?", userID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at desc").Limit(limit).Offset(offset).Find(&reqs).Error
	return reqs, total, err
}

func (r *payslipRequestRepository) GetAll(status string, limit, offset int) ([]domain.PayslipRequest, int64, error) {
	var reqs []domain.PayslipRequest
	var total int64

	query := r.db.Model(&domain.PayslipRequest{}).
		Preload("User").
		Preload("User.Employee").
		Preload("User.Employee.JobPosition").
		Preload("User.Employee.Department").
		Preload("Approver")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at desc").Limit(limit).Offset(offset).Find(&reqs).Error
	return reqs, total, err
}

func (r *payslipRequestRepository) GetByID(id uuid.UUID) (*domain.PayslipRequest, error) {
	var req domain.PayslipRequest
	err := r.db.
		Preload("User").
		Preload("User.Employee").
		Preload("User.Employee.JobPosition").
		Preload("User.Employee.Department").
		Preload("Approver").
		First(&req, "id = ?", id).Error
	return &req, err
}

func (r *payslipRequestRepository) Update(req *domain.PayslipRequest) error {
	return r.db.Save(req).Error
}
