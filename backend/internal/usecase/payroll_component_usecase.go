package usecase

import (
	"errors"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type PayrollComponentUseCase interface {
	CreateComponent(req *CreatePayrollComponentRequest) error
	GetComponents() ([]domain.PayrollComponent, error)
	GetComponentByID(id uuid.UUID) (*domain.PayrollComponent, error)
	UpdateComponent(id uuid.UUID, req *UpdatePayrollComponentRequest) error
	DeleteComponent(id uuid.UUID) error
}

type payrollComponentUseCase struct {
	repo repository.PayrollComponentRepository
}

type CreatePayrollComponentRequest struct {
	Name      string     `json:"name" binding:"required"`
	Type      string     `json:"type" binding:"required"` // EARNING, DEDUCTION, BENEFIT
	IsTaxable bool       `json:"isTaxable"`
	CompanyID *uuid.UUID `json:"companyId"`
}

type UpdatePayrollComponentRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required"`
	IsTaxable bool   `json:"isTaxable"`
}

func NewPayrollComponentUseCase(repo repository.PayrollComponentRepository) PayrollComponentUseCase {
	return &payrollComponentUseCase{repo: repo}
}

func (u *payrollComponentUseCase) CreateComponent(req *CreatePayrollComponentRequest) error {
	component := &domain.PayrollComponent{
		Name:      req.Name,
		Type:      req.Type,
		IsTaxable: req.IsTaxable,
		CompanyID: req.CompanyID,
	}

	return u.repo.Create(component)
}

func (u *payrollComponentUseCase) GetComponents() ([]domain.PayrollComponent, error) {
	return u.repo.FindAll()
}

func (u *payrollComponentUseCase) GetComponentByID(id uuid.UUID) (*domain.PayrollComponent, error) {
	return u.repo.FindByID(id)
}

func (u *payrollComponentUseCase) UpdateComponent(id uuid.UUID, req *UpdatePayrollComponentRequest) error {
	component, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if component == nil {
		return errors.New("payroll component not found")
	}

	component.Name = req.Name
	component.Type = req.Type
	component.IsTaxable = req.IsTaxable

	return u.repo.Update(component)
}

func (u *payrollComponentUseCase) DeleteComponent(id uuid.UUID) error {
	return u.repo.Delete(id)
}
