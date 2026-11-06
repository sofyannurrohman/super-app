package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type LeaveTypeRepository interface {
	Create(leaveType *domain.LeaveType) error
	FindAll() ([]domain.LeaveType, error)
	FindByID(id uuid.UUID) (*domain.LeaveType, error)
	Update(leaveType *domain.LeaveType) error
	Delete(id uuid.UUID) error
}

type leaveTypeRepository struct {
	db *gorm.DB
}

func NewLeaveTypeRepository(db *gorm.DB) LeaveTypeRepository {
	return &leaveTypeRepository{db}
}

func (r *leaveTypeRepository) Create(leaveType *domain.LeaveType) error {
	return r.db.Create(leaveType).Error
}

func (r *leaveTypeRepository) FindAll() ([]domain.LeaveType, error) {
	var leaveTypes []domain.LeaveType
	if err := r.db.Find(&leaveTypes).Error; err != nil {
		return nil, err
	}
	return leaveTypes, nil
}

func (r *leaveTypeRepository) FindByID(id uuid.UUID) (*domain.LeaveType, error) {
	var leaveType domain.LeaveType
	if err := r.db.Where("id = ?", id).First(&leaveType).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &leaveType, nil
}

func (r *leaveTypeRepository) Update(leaveType *domain.LeaveType) error {
	return r.db.Save(leaveType).Error
}

func (r *leaveTypeRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete related leave balances
		if err := tx.Where("leave_type_id = ?", id).Delete(&domain.LeaveBalance{}).Error; err != nil {
			return err
		}

		// Delete related leave requests
		if err := tx.Where("leave_type_id = ?", id).Delete(&domain.LeaveRequest{}).Error; err != nil {
			return err
		}

		// Finally delete the leave type
		if err := tx.Where("id = ?", id).Delete(&domain.LeaveType{}).Error; err != nil {
			return err
		}

		return nil
	})
}
