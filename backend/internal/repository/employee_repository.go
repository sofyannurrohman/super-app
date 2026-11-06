package repository

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type EmployeeRepository interface {
	Create(employee *domain.Employee) error
	FindByID(id uuid.UUID) (*domain.Employee, error)
	FindByUserID(userID uuid.UUID) (*domain.Employee, error)
	Update(employee *domain.Employee) error
	Delete(id uuid.UUID) error
	FindBirthdaysToday() ([]domain.Employee, error)
	FindAll(limit int, branchIDs []uuid.UUID, jobPositionID *uuid.UUID, branchMatch ...string) ([]domain.Employee, error)
	FindPaginated(page, limit int, branchIDs []uuid.UUID, search string, isSales bool, salesType string, employmentType string, status string, jobPositionID *uuid.UUID, branchMatch ...string) ([]domain.Employee, int64, error)
	FindByFullNameExact(fullName string) (*domain.Employee, error)
	FindByPin(pin string, branchID *uuid.UUID) (*domain.Employee, error)
}

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db}
}

func (r *employeeRepository) Create(employee *domain.Employee) error {
	return r.db.Create(employee).Error
}

func (r *employeeRepository) FindAll(limit int, branchIDs []uuid.UUID, jobPositionID *uuid.UUID, branchMatch ...string) ([]domain.Employee, error) {
	var employees []domain.Employee
	today := time.Now().Format("2006-01-02")
	query := r.db.Preload("User").Preload("Branch").Preload("OperatingBranch").Preload("Department").Preload("JobPosition").Preload("Company").
		Preload("EmployeeShifts", "date = ?", today).
		Preload("EmployeeShifts.Shift").
		Order("first_name ASC, last_name ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if len(branchIDs) > 0 {
		bm := ""
		if len(branchMatch) > 0 {
			bm = branchMatch[0]
		}
		if bm == "primary" {
			query = query.Where("branch_id IN (?)", branchIDs)
		} else if bm == "both" {
			query = query.Where("(branch_id IN (?) OR operating_branch_id IN (?))", branchIDs, branchIDs)
		} else {
			query = query.Where("COALESCE(operating_branch_id, branch_id) IN (?)", branchIDs)
		}
	}
	if jobPositionID != nil {
		query = query.Where("job_position_id = ?", jobPositionID)
	}
	if err := query.Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *employeeRepository) FindPaginated(page, limit int, branchIDs []uuid.UUID, search string, isSales bool, salesType string, employmentType string, status string, jobPositionID *uuid.UUID, branchMatch ...string) ([]domain.Employee, int64, error) {
	var employees []domain.Employee
	var total int64
	today := time.Now().Format("2006-01-02")

	query := r.db.Model(&domain.Employee{})
	if status != "" {
		if strings.EqualFold(status, "active") {
			query = query.Where("COALESCE(employees.employment_status, 'Active') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED')")
		} else if strings.EqualFold(status, "inactive") {
			query = query.Where("employees.employment_status IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED')")
		} else {
			query = query.Where("employees.employment_status = ?", status)
		}
	}
	if len(branchIDs) > 0 {
		bm := ""
		if len(branchMatch) > 0 {
			bm = branchMatch[0]
		}
		if bm == "primary" {
			query = query.Where("employees.branch_id IN (?)", branchIDs)
		} else if bm == "both" {
			query = query.Where("(employees.branch_id IN (?) OR employees.operating_branch_id IN (?))", branchIDs, branchIDs)
		} else {
			query = query.Where("COALESCE(employees.operating_branch_id, employees.branch_id) IN (?)", branchIDs)
		}
	}
	if jobPositionID != nil {
		query = query.Where("employees.job_position_id = ?", jobPositionID)
	}

	if employmentType != "" {
		if strings.EqualFold(employmentType, "reguler") || strings.EqualFold(employmentType, "regular") {
			query = query.Where("COALESCE(employees.employment_status, '') NOT ILIKE '%harian%'")
		} else if strings.EqualFold(employmentType, "harian") {
			query = query.Where("employees.employment_status ILIKE '%harian%'")
		} else {
			query = query.Where("employees.employment_status ILIKE ?", "%"+employmentType+"%")
		}
	}

	if isSales {
		query = query.Joins("LEFT JOIN job_positions ON job_positions.id = employees.job_position_id").
			Joins("LEFT JOIN departments ON departments.id = employees.department_id")

		query = query.Where("COALESCE(job_positions.title, '') NOT ILIKE '%operator%' AND COALESCE(job_positions.title, '') NOT ILIKE '%driver%' AND COALESCE(job_positions.title, '') NOT ILIKE '%auditor%' AND COALESCE(job_positions.title, '') NOT ILIKE '%editor%' AND COALESCE(job_positions.title, '') NOT ILIKE '%customer service%' AND COALESCE(job_positions.title, '') NOT ILIKE '%validator%' AND COALESCE(job_positions.title, '') NOT ILIKE '%gudang%' AND COALESCE(job_positions.title, '') NOT ILIKE '%security%'")

		if salesType == "sales_mt" {
			query = query.Where("(job_positions.title ILIKE '%sales mt%' OR job_positions.title ILIKE '%modern trade%' OR job_positions.title ILIKE '%kam%' OR job_positions.title ILIKE '%key account%' OR job_positions.title ~* '\\mmt\\M') OR (departments.name ILIKE '%modern trade%' OR departments.name ~* '\\mmt\\M')")
		} else if salesType == "sales_to" {
			query = query.Where("job_positions.title ILIKE '%take order%' OR job_positions.title ILIKE '%task order%' OR job_positions.title ILIKE '%sales to%' OR job_positions.title ~* '\\mto\\M' OR departments.name ILIKE '%take order%' OR departments.name ~* '\\mto\\M'")
		} else if salesType == "sales_motoris_booster" {
			query = query.Where("job_positions.title ILIKE '%motoris%' OR job_positions.title ILIKE '%booster%' OR departments.name ILIKE '%motoris%' OR departments.name ILIKE '%booster%'")
		} else {
			query = query.Where("job_positions.title ILIKE '%sales%' OR job_positions.title ILIKE '%marketing%' OR job_positions.title ILIKE '%motoris%' OR job_positions.title ILIKE '%booster%' OR job_positions.title ~* '\\mto\\M' OR job_positions.title ~* '\\mmt\\M' OR departments.name ILIKE '%sales%' OR departments.name ILIKE '%marketing%'")
		}
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Joins("LEFT JOIN users ON users.id = employees.user_id").
			Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR CONCAT(employees.first_name, ' ', COALESCE(employees.last_name, '')) ILIKE ? OR employees.employee_id_number ILIKE ? OR users.email ILIKE ?", searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	query = query.Preload("User").Preload("Branch").Preload("OperatingBranch").Preload("Department").Preload("JobPosition").Preload("Company").
		Preload("EmployeeShifts", "date = ?", today).
		Preload("EmployeeShifts.Shift")

	if err := query.Find(&employees).Error; err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

func (r *employeeRepository) FindByID(id uuid.UUID) (*domain.Employee, error) {
	var employee domain.Employee
	if err := r.db.Preload("User").Preload("Branch").Preload("OperatingBranch").Preload("Department").Preload("JobPosition").Preload("Company").Preload("DistributorStore").Where("id = ?", id).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) FindByUserID(userID uuid.UUID) (*domain.Employee, error) {
	var employee domain.Employee
	today := time.Now().Format("2006-01-02")
	if err := r.db.Preload("Branch").
		Preload("Branch.DefaultShift").
		Preload("User").
		Preload("Department").
		Preload("JobPosition").
		Preload("Company").
		Preload("DistributorStore").
		Preload("DefaultShift").
		Preload("EmployeeShifts", "date = ?", today).
		Preload("EmployeeShifts.Shift").
		Where("user_id = ?", userID).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) FindByFullNameExact(fullName string) (*domain.Employee, error) {
	var employee domain.Employee
	// Simple assumption: If fullName contains space, split into first_name and last_name, or just query concat.
	// To be safe, we query by concat(first_name, ' ', last_name) or just check first_name for now since we might save the whole name in first_name.
	// A safer approach: where trim(concat(first_name, ' ', COALESCE(last_name, ''))) ILIKE ?
	if err := r.db.Where("TRIM(CONCAT(first_name, ' ', COALESCE(last_name, ''))) ILIKE ?", fullName).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) FindByPin(pin string, branchID *uuid.UUID) (*domain.Employee, error) {
	var employee domain.Employee
	query := r.db.Where("pin = ?", pin)
	if branchID != nil {
		query = query.Where("branch_id = ?", branchID)
	}
	if err := query.Preload("User").Preload("Branch").First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) Update(employee *domain.Employee) error {
	return r.db.Omit("User", "Branch", "Department", "JobPosition", "Manager", "EmployeeShifts", "Company", "DefaultShift").Save(employee).Error
}

func (r *employeeRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Handle Payslip Items (via Payslips)
		// We need to delete payslip items before payslips
		if err := tx.Exec("DELETE FROM payslip_items WHERE payslip_id IN (SELECT id FROM payslips WHERE employee_id = ?)", id).Error; err != nil {
			return err
		}

		// 2. Delete dependent records using GORM's Where logic for safety
		dependentModels := []interface{}{
			&domain.Payslip{},
			&domain.EmployeeSalarySetting{},
			&domain.LeaveBalance{},
			&domain.LeaveRequest{},
			&domain.AttendanceLog{},
			&domain.EmployeeShift{},
			&domain.Reimbursement{},
			&domain.Overtime{},
			&domain.SalesKPI{},
			&domain.EmployeeKPI{},
			&domain.ManagerialAppraisal{},
		}

		for _, model := range dependentModels {
			if err := tx.Where("employee_id = ?", id).Delete(model).Error; err != nil {
				return err
			}
		}

		// 3. Handle specific FKs that might point to this employee as an approver/manager
		// Set ApprovedBy to NULL where this employee was the approver
		if err := tx.Model(&domain.LeaveRequest{}).Where("approved_by = ?", id).Update("approved_by", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.Reimbursement{}).Where("approved_by = ?", id).Update("approved_by", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.Overtime{}).Where("approved_by = ?", id).Update("approved_by", nil).Error; err != nil {
			return err
		}

		// Handle ManagerialAppraisal where this employee was the manager
		// Since ManagerID is NOT NULL, we must delete these appraisals
		if err := tx.Where("manager_id = ?", id).Delete(&domain.ManagerialAppraisal{}).Error; err != nil {
			return err
		}

		// 4. Update subordinates (set manager_id to NULL)
		if err := tx.Model(&domain.Employee{}).Where("manager_id = ?", id).Update("manager_id", nil).Error; err != nil {
			return err
		}

		// Keep transfer history, but remove references to the deleted driver.
		// Existing databases may still have restrictive GORM-created foreign keys.
		if err := tx.Table("product_transfers").Where("driver_id = ?", id).Update("driver_id", nil).Error; err != nil {
			return err
		}

		// 5. Finally delete the employee record
		if err := tx.Where("id = ?", id).Delete(&domain.Employee{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *employeeRepository) FindBirthdaysToday() ([]domain.Employee, error) {
	var employees []domain.Employee
	// Postgres syntax for extracting month and day
	if err := r.db.Where("EXTRACT(MONTH FROM birth_date) = EXTRACT(MONTH FROM CURRENT_DATE) AND EXTRACT(DAY FROM birth_date) = EXTRACT(DAY FROM CURRENT_DATE)").Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}
