package repository

import (
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeaveRepository interface {
	Create(leave *domain.LeaveRequest) error
	Update(leave *domain.LeaveRequest) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*domain.LeaveRequest, error)
	FindByUserID(userID uuid.UUID, limit, offset int) ([]domain.LeaveRequest, error)
	FindAll(status, search string, limit, offset int, branchIDs []uuid.UUID) ([]domain.LeaveRequest, int64, error)
	GetLeaveTypeByID(id uuid.UUID) (*domain.LeaveType, error)
	GetLeaveTypesByCompany(companyID uuid.UUID) ([]domain.LeaveType, error)
	GetLeaveBalance(employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (*domain.LeaveBalance, error)
	GetAllBalancesByEmployee(employeeID uuid.UUID, year int) ([]*domain.LeaveBalance, error)
	GetAllBalances() ([]*domain.LeaveBalance, error)
	CreateLeaveRequestWithBalance(req *domain.LeaveRequest, balance *domain.LeaveBalance) error
	UpdateLeaveRequestStatus(req *domain.LeaveRequest, balanceToRefund *domain.LeaveBalance) error
	SaveBalance(balance *domain.LeaveBalance) error
	DeleteLeaveBalance(employeeID, leaveTypeID uuid.UUID, year int) error
	CountMonthlyLeaveDays(employeeID, leaveTypeID uuid.UUID, month, year int, excludeID uuid.UUID) (int, error)
	FindApprovedByEmployeeAndDateRange(employeeID uuid.UUID, start, end time.Time) ([]domain.LeaveRequest, error)
}

type leaveRepository struct {
	db *gorm.DB
}

func NewLeaveRepository(db *gorm.DB) LeaveRepository {
	return &leaveRepository{db}
}

func (r *leaveRepository) Create(leave *domain.LeaveRequest) error {
	return r.db.Create(leave).Error
}

func (r *leaveRepository) Update(leave *domain.LeaveRequest) error {
	return r.db.Omit(clause.Associations).Save(leave).Error
}

func (r *leaveRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.LeaveRequest{}, "id = ?", id).Error
}

func (r *leaveRepository) FindByID(id uuid.UUID) (*domain.LeaveRequest, error) {
	var leave domain.LeaveRequest
	err := r.db.Preload("Employee").
		Preload("Employee.Branch").
		Preload("Employee.JobPosition").
		Preload("LeaveType").
		Preload("Approver").
		Preload("Approver.JobPosition").
		Where("id = ?", id).First(&leave).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &leave, nil
}

func (r *leaveRepository) FindByUserID(employeeID uuid.UUID, limit, offset int) ([]domain.LeaveRequest, error) {
	var leaves []domain.LeaveRequest
	err := r.db.Where("employee_id = ?", employeeID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("LeaveType").
		Preload("Approver").
		Preload("Approver.JobPosition").
		Find(&leaves).Error
	return leaves, err
}

func (r *leaveRepository) FindAll(status, search string, limit, offset int, branchIDs []uuid.UUID) ([]domain.LeaveRequest, int64, error) {
	var leaves []domain.LeaveRequest
	var total int64
	query := r.db.Model(&domain.LeaveRequest{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if len(branchIDs) > 0 {
		query = query.Where("employee_id IN (SELECT id FROM employees WHERE branch_id IN (?))", branchIDs)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"employee_id IN (SELECT id FROM employees WHERE first_name ILIKE ? OR last_name ILIKE ?) OR reason ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Employee").
		Preload("Employee.Branch").
		Preload("Employee.JobPosition").
		Preload("LeaveType").
		Preload("Approver").
		Preload("Approver.JobPosition").
		Find(&leaves).Error

	return leaves, total, err
}

func (r *leaveRepository) GetLeaveTypeByID(id uuid.UUID) (*domain.LeaveType, error) {
	var leaveType domain.LeaveType
	err := r.db.Where("id = ?", id).First(&leaveType).Error
	return &leaveType, err
}

func (r *leaveRepository) GetLeaveTypesByCompany(companyID uuid.UUID) ([]domain.LeaveType, error) {
	var leaveTypes []domain.LeaveType
	err := r.db.Where("company_id = ?", companyID).Find(&leaveTypes).Error
	return leaveTypes, err
}

func (r *leaveRepository) GetLeaveBalance(employeeID uuid.UUID, leaveTypeID uuid.UUID, year int) (*domain.LeaveBalance, error) {
	var balance domain.LeaveBalance
	err := r.db.Where("employee_id = ? AND leave_type_id = ? AND year = ?", employeeID, leaveTypeID, year).First(&balance).Error
	return &balance, err
}

func (r *leaveRepository) GetAllBalancesByEmployee(employeeID uuid.UUID, year int) ([]*domain.LeaveBalance, error) {
	var balances []*domain.LeaveBalance
	err := r.db.Preload("LeaveType").Where("employee_id = ? AND year = ?", employeeID, year).Find(&balances).Error
	return balances, err
}

func (r *leaveRepository) GetAllBalances() ([]*domain.LeaveBalance, error) {
	var balances []*domain.LeaveBalance
	err := r.db.Preload("LeaveType").Preload("Employee").Order("year desc").Find(&balances).Error
	return balances, err
}

func (r *leaveRepository) CreateLeaveRequestWithBalance(req *domain.LeaveRequest, balance *domain.LeaveBalance) error {
	tx := r.db.Begin()

	if err := tx.Create(req).Error; err != nil {
		tx.Rollback()
		return err
	}

	if balance != nil {
		if err := tx.Save(balance).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *leaveRepository) UpdateLeaveRequestStatus(req *domain.LeaveRequest, balanceToRefund *domain.LeaveBalance) error {
	tx := r.db.Begin()

	if err := tx.Omit(clause.Associations).Save(req).Error; err != nil {
		tx.Rollback()
		return err
	}

	if balanceToRefund != nil {
		if err := tx.Save(balanceToRefund).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *leaveRepository) SaveBalance(balance *domain.LeaveBalance) error {
	return r.db.Save(balance).Error
}

func (r *leaveRepository) DeleteLeaveBalance(employeeID, leaveTypeID uuid.UUID, year int) error {
	return r.db.Delete(&domain.LeaveBalance{}, "employee_id = ? AND leave_type_id = ? AND year = ?", employeeID, leaveTypeID, year).Error
}

func (r *leaveRepository) CountMonthlyLeaveDays(employeeID, leaveTypeID uuid.UUID, month, year int, excludeID uuid.UUID) (int, error) {
	var requests []domain.LeaveRequest
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, -1)

	// Fetch all requests that overlap with this month for this employee and leave type
	query := r.db.Where("employee_id = ? AND leave_type_id = ? AND status NOT IN ('REJECTED', 'CANCELLED')", employeeID, leaveTypeID)
	if excludeID != uuid.Nil {
		query = query.Where("id != ?", excludeID)
	}

	err := query.Where("start_date <= ? AND end_date >= ?", monthEnd, monthStart).
		Find(&requests).Error
	if err != nil {
		return 0, err
	}

	totalDays := 0
	for _, req := range requests {
		start := req.StartDate
		if start.Before(monthStart) {
			start = monthStart
		}
		end := req.EndDate
		if end.After(monthEnd) {
			end = monthEnd
		}

		// Calculate overlap days
		days := int(end.Sub(start).Hours()/24) + 1
		totalDays += days
	}

	return totalDays, nil
}
func (r *leaveRepository) FindApprovedByEmployeeAndDateRange(employeeID uuid.UUID, start, end time.Time) ([]domain.LeaveRequest, error) {
	var results []domain.LeaveRequest
	err := r.db.Preload("LeaveType").Where("employee_id = ? AND status = 'APPROVED'", employeeID).
		Where("((start_date <= ? AND end_date >= ?) OR (start_date >= ? AND start_date <= ?))", end, start, start, end).
		Find(&results).Error
	return results, err
}
