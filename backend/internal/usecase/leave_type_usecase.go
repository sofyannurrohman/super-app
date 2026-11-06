package usecase

import (
	"errors"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type LeaveTypeUseCase interface {
	CreateLeaveType(req *CreateLeaveTypeRequest) error
	GetLeaveTypes() ([]domain.LeaveType, error)
	GetLeaveTypeByID(id uuid.UUID) (*domain.LeaveType, error)
	UpdateLeaveType(id uuid.UUID, req *UpdateLeaveTypeRequest) error
	DeleteLeaveType(id uuid.UUID) error
}

type leaveTypeUseCase struct {
	repo repository.LeaveTypeRepository
}

type CreateLeaveTypeRequest struct {
	Name         string     `json:"name" binding:"required"`
	IsPaid       bool       `json:"isPaid"`
	DefaultQuota int        `json:"defaultQuota"`
	CompanyID    *uuid.UUID `json:"companyId"`
}

type UpdateLeaveTypeRequest struct {
	Name         string `json:"name" binding:"required"`
	IsPaid       bool   `json:"isPaid"`
	DefaultQuota int    `json:"defaultQuota"`
}

func NewLeaveTypeUseCase(repo repository.LeaveTypeRepository) LeaveTypeUseCase {
	return &leaveTypeUseCase{repo: repo}
}

func (u *leaveTypeUseCase) CreateLeaveType(req *CreateLeaveTypeRequest) error {
	leaveType := &domain.LeaveType{
		Name:         req.Name,
		IsPaid:       req.IsPaid,
		DefaultQuota: req.DefaultQuota,
		CompanyID:    req.CompanyID,
	}

	return u.repo.Create(leaveType)
}

func (u *leaveTypeUseCase) GetLeaveTypes() ([]domain.LeaveType, error) {
	return u.repo.FindAll()
}

func (u *leaveTypeUseCase) GetLeaveTypeByID(id uuid.UUID) (*domain.LeaveType, error) {
	return u.repo.FindByID(id)
}

func (u *leaveTypeUseCase) UpdateLeaveType(id uuid.UUID, req *UpdateLeaveTypeRequest) error {
	leaveType, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if leaveType == nil {
		return errors.New("leave type not found")
	}

	leaveType.Name = req.Name
	leaveType.IsPaid = req.IsPaid
	leaveType.DefaultQuota = req.DefaultQuota

	return u.repo.Update(leaveType)
}

func (u *leaveTypeUseCase) DeleteLeaveType(id uuid.UUID) error {
	return u.repo.Delete(id)
}
