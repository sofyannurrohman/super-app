package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type EmployeeShiftUseCase interface {
	AssignShift(req *AssignShiftRequest) error
	GetAllAssignments() ([]domain.EmployeeShift, error)
	GetAssignmentsByEmployee(employeeID uuid.UUID) ([]domain.EmployeeShift, error)
	UpdateAssignment(id uuid.UUID, req *UpdateShiftAssignmentRequest) error
	DeleteAssignment(id uuid.UUID) error
}

type employeeShiftUseCase struct {
	repo         repository.EmployeeShiftRepository
	employeeRepo repository.EmployeeRepository
	shiftRepo    repository.ShiftRepository
}

type AssignShiftRequest struct {
	EmployeeID uuid.UUID `json:"employeeId" binding:"required"`
	ShiftID    uuid.UUID `json:"shiftId" binding:"required"`
	Date       time.Time `json:"date" binding:"required"`
	IsOffDay   bool      `json:"isOffDay"`
}

type UpdateShiftAssignmentRequest struct {
	ShiftID  uuid.UUID `json:"shiftId" binding:"required"`
	Date     time.Time `json:"date" binding:"required"`
	IsOffDay bool      `json:"isOffDay"`
}

func NewEmployeeShiftUseCase(repo repository.EmployeeShiftRepository, employeeRepo repository.EmployeeRepository, shiftRepo repository.ShiftRepository) EmployeeShiftUseCase {
	return &employeeShiftUseCase{repo: repo, employeeRepo: employeeRepo, shiftRepo: shiftRepo}
}

func (u *employeeShiftUseCase) AssignShift(req *AssignShiftRequest) error {
	// Pre-validation to avoid FK constraint violations
	emp, _ := u.employeeRepo.FindByID(req.EmployeeID)
	if emp == nil {
		return errors.New("employee record not found")
	}

	shift, _ := u.shiftRepo.FindByID(req.ShiftID)
	if shift == nil {
		return errors.New("shift record not found")
	}

	es := &domain.EmployeeShift{
		EmployeeID: req.EmployeeID,
		ShiftID:    req.ShiftID,
		Date:       req.Date,
		IsOffDay:   req.IsOffDay,
	}
	return u.repo.Create(es)
}

func (u *employeeShiftUseCase) GetAllAssignments() ([]domain.EmployeeShift, error) {
	return u.repo.FindAll()
}

func (u *employeeShiftUseCase) GetAssignmentsByEmployee(employeeID uuid.UUID) ([]domain.EmployeeShift, error) {
	return u.repo.FindByEmployeeID(employeeID)
}

func (u *employeeShiftUseCase) UpdateAssignment(id uuid.UUID, req *UpdateShiftAssignmentRequest) error {
	es, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if es == nil {
		return errors.New("shift assignment not found")
	}

	es.ShiftID = req.ShiftID
	es.Date = req.Date
	es.IsOffDay = req.IsOffDay

	return u.repo.Update(es)
}

func (u *employeeShiftUseCase) DeleteAssignment(id uuid.UUID) error {
	return u.repo.Delete(id)
}
