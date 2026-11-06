package usecase

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type PayslipRequestUsecase interface {
	CreateRequest(userID uuid.UUID, month, year int, reason *string) error
	GetMyRequests(userID uuid.UUID, limit, offset int) ([]domain.PayslipRequest, int64, error)
	GetAllRequests(status string, limit, offset int) ([]domain.PayslipRequest, int64, error)
	LookupPayrollData(requestID uuid.UUID) (*domain.PayslipLookupResult, error)
	ApproveRequest(id uuid.UUID, nominal *float64, fileURL *string, payslipData *string, approvedBy *uuid.UUID) error
	RejectRequest(id uuid.UUID, reason string) error
}

type payslipRequestUsecase struct {
	repo         repository.PayslipRequestRepository
	employeeRepo repository.EmployeeRepository
	configRepo   repository.PayrollConfigRepository
	payrollRepo  repository.PayrollRepository
	db           *gorm.DB
}

func NewPayslipRequestUsecase(
	repo repository.PayslipRequestRepository,
	employeeRepo repository.EmployeeRepository,
	configRepo repository.PayrollConfigRepository,
	payrollRepo repository.PayrollRepository,
	db *gorm.DB,
) PayslipRequestUsecase {
	return &payslipRequestUsecase{
		repo:         repo,
		employeeRepo: employeeRepo,
		configRepo:   configRepo,
		payrollRepo:  payrollRepo,
		db:           db,
	}
}

func (u *payslipRequestUsecase) CreateRequest(userID uuid.UUID, month, year int, reason *string) error {
	req := &domain.PayslipRequest{
		UserID:      userID,
		PeriodMonth: month,
		PeriodYear:  year,
		Reason:      reason,
		Status:      domain.PayslipRequestStatusPending,
	}
	return u.repo.Create(req)
}

func (u *payslipRequestUsecase) GetMyRequests(userID uuid.UUID, limit, offset int) ([]domain.PayslipRequest, int64, error) {
	return u.repo.GetByUserID(userID, limit, offset)
}

func (u *payslipRequestUsecase) GetAllRequests(status string, limit, offset int) ([]domain.PayslipRequest, int64, error) {
	return u.repo.GetAll(status, limit, offset)
}

func getMonthNameID(month int) string {
	months := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	if month >= 1 && month <= 12 {
		return months[month]
	}
	return ""
}

func (u *payslipRequestUsecase) LookupPayrollData(requestID uuid.UUID) (*domain.PayslipLookupResult, error) {
	req, err := u.repo.GetByID(requestID)
	if err != nil || req == nil {
		return nil, errors.New("request slip gaji tidak ditemukan")
	}

	emp, err := u.employeeRepo.FindByUserID(req.UserID)
	if err != nil || emp == nil {
		return nil, errors.New("data profil karyawan untuk user ini belum terdaftar")
	}

	// 1. Period bounds
	year := req.PeriodYear
	month := time.Month(req.PeriodMonth)
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	endOfMonth = time.Date(year, month, endOfMonth.Day(), 23, 59, 59, 999999999, time.Local)

	periodLabel := fmt.Sprintf("%s %d", getMonthNameID(req.PeriodMonth), req.PeriodYear)

	// 2. Employee Info
	empName := fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)
	jobTitle := ""
	if emp.JobPosition != nil {
		jobTitle = emp.JobPosition.Title
	}
	departmentName := ""
	if emp.Department != nil {
		departmentName = emp.Department.Name
	}

	empLookup := domain.PayslipLookupEmployee{
		ID:                emp.ID.String(),
		Name:              empName,
		EmployeeIDNumber:  emp.EmployeeIDNumber,
		JobPosition:       jobTitle,
		Department:        departmentName,
		PtkpStatus:        emp.PtkpStatus,
		BankName:          emp.BankName,
		BankAccountNumber: emp.BankAccountNumber,
	}

	// 3. Basic Salary & Allowances
	var basicSalary float64
	if emp.Salary != nil && *emp.Salary > 0 {
		basicSalary = *emp.Salary
	}

	var allowances []domain.PayslipItemSummary
	var deductions []domain.PayslipItemSummary
	var totalAllowance float64
	var totalDeduction float64

	// Fetch specific salary settings if any
	settings, err := u.payrollRepo.GetEmployeeSalarySettings(emp.ID)
	if err == nil {
		for _, s := range settings {
			if s.Component == nil {
				continue
			}
			if s.Component.Type == "EARNING" {
				if s.Component.Name == "Gaji Pokok" || s.Component.Name == "Basic Salary" {
					if basicSalary == 0 {
						basicSalary = s.Amount
					}
				} else {
					allowances = append(allowances, domain.PayslipItemSummary{
						Name:   s.Component.Name,
						Amount: s.Amount,
						Type:   "EARNING",
					})
					totalAllowance += s.Amount
				}
			} else if s.Component.Type == "DEDUCTION" {
				deductions = append(deductions, domain.PayslipItemSummary{
					Name:   s.Component.Name,
					Amount: s.Amount,
					Type:   "DEDUCTION",
				})
				totalDeduction += s.Amount
			}
		}
	}

	// 4. Payroll Config (BPJS & Mangkir)
	var companyID uuid.UUID
	if emp.CompanyID != nil {
		companyID = *emp.CompanyID
	}
	config, err := u.configRepo.GetByCompanyID(companyID)
	if err != nil || config == nil {
		config = &domain.PayrollConfig{
			BPJSKesEmployeePercentage: 1.0,
			BPJSKesMaxWageBase:        12000000,
			JHTEmployeePercentage:     2.0,
			JPEmployeePercentage:      1.0,
			JPMaxWageBase:             10042300,
			AbsentDeduction:           100000.0,
		}
	}

	// 5. Attendance Lookup
	var presentCount int64
	var lateCount int64
	var absentCount int64

	u.db.Model(&domain.AttendanceLog{}).
		Where("employee_id = ? AND clock_in_time >= ? AND clock_in_time <= ? AND status != ?", emp.ID, startOfMonth, endOfMonth, "ABSENT").
		Count(&presentCount)

	u.db.Model(&domain.AttendanceLog{}).
		Where("employee_id = ? AND clock_in_time >= ? AND clock_in_time <= ? AND status = ?", emp.ID, startOfMonth, endOfMonth, "LATE").
		Count(&lateCount)

	u.db.Model(&domain.AttendanceLog{}).
		Where("employee_id = ? AND clock_in_time >= ? AND clock_in_time <= ? AND status = ?", emp.ID, startOfMonth, endOfMonth, "ABSENT").
		Count(&absentCount)

	absentDeduction := float64(absentCount) * config.AbsentDeduction
	if absentDeduction > 0 {
		deductions = append(deductions, domain.PayslipItemSummary{
			Name:   fmt.Sprintf("Potongan Mangkir (%d hari)", absentCount),
			Amount: absentDeduction,
			Type:   "DEDUCTION",
		})
		totalDeduction += absentDeduction
	}

	attendanceLookup := domain.PayslipLookupAttendance{
		PresentDays:     int(presentCount),
		LateDays:        int(lateCount),
		AbsentDays:      int(absentCount),
		AbsentDeduction: absentDeduction,
	}

	// 6. Leave Lookup
	var leaves []domain.LeaveRequest
	u.db.Preload("LeaveType").
		Where("employee_id = ? AND status = ? AND start_date <= ? AND end_date >= ?", emp.ID, "APPROVED", endOfMonth, startOfMonth).
		Find(&leaves)

	paidLeaveDays := 0
	unpaidLeaveDays := 0

	for _, lv := range leaves {
		// Calculate overlap days
		lvStart := lv.StartDate
		if lvStart.Before(startOfMonth) {
			lvStart = startOfMonth
		}
		lvEnd := lv.EndDate
		if lvEnd.After(endOfMonth) {
			lvEnd = endOfMonth
		}
		days := int(math.Max(1, lvEnd.Sub(lvStart).Hours()/24+1))

		if lv.LeaveType != nil && !lv.LeaveType.IsPaid {
			unpaidLeaveDays += days
		} else {
			paidLeaveDays += days
		}
	}

	var unpaidDeduction float64
	if unpaidLeaveDays > 0 && basicSalary > 0 {
		// Standard 25 working days per month
		unpaidDeduction = float64(unpaidLeaveDays) * (basicSalary / 25.0)
		deductions = append(deductions, domain.PayslipItemSummary{
			Name:   fmt.Sprintf("Potongan Izin Unpaid (%d hari)", unpaidLeaveDays),
			Amount: unpaidDeduction,
			Type:   "DEDUCTION",
		})
		totalDeduction += unpaidDeduction
	}

	leaveLookup := domain.PayslipLookupLeave{
		PaidLeaveDays:   paidLeaveDays,
		UnpaidLeaveDays: unpaidLeaveDays,
		UnpaidDeduction: unpaidDeduction,
	}

	// 7. BPJS Ketenagakerjaan & BPJS Kesehatan
	var jhtDeduction float64
	var jpDeduction float64
	var kesDeduction float64

	if basicSalary > 0 {
		// BPJS Ketenagakerjaan - JHT (2%)
		jhtDeduction = basicSalary * (config.JHTEmployeePercentage / 100.0)
		if jhtDeduction > 0 {
			deductions = append(deductions, domain.PayslipItemSummary{
				Name:   "BPJS Ketenagakerjaan JHT (2%)",
				Amount: jhtDeduction,
				Type:   "DEDUCTION",
			})
			totalDeduction += jhtDeduction
		}

		// BPJS Ketenagakerjaan - JP (1%) with wage cap
		jpBasis := basicSalary
		if config.JPMaxWageBase > 0 && jpBasis > config.JPMaxWageBase {
			jpBasis = config.JPMaxWageBase
		}
		jpDeduction = jpBasis * (config.JPEmployeePercentage / 100.0)
		if jpDeduction > 0 {
			deductions = append(deductions, domain.PayslipItemSummary{
				Name:   "BPJS Ketenagakerjaan JP (1%)",
				Amount: jpDeduction,
				Type:   "DEDUCTION",
			})
			totalDeduction += jpDeduction
		}

		// BPJS Kesehatan (1%) with wage cap
		kesBasis := basicSalary
		if config.BPJSKesMaxWageBase > 0 && kesBasis > config.BPJSKesMaxWageBase {
			kesBasis = config.BPJSKesMaxWageBase
		}
		kesDeduction = kesBasis * (config.BPJSKesEmployeePercentage / 100.0)
		if kesDeduction > 0 {
			deductions = append(deductions, domain.PayslipItemSummary{
				Name:   "BPJS Kesehatan (1%)",
				Amount: kesDeduction,
				Type:   "DEDUCTION",
			})
			totalDeduction += kesDeduction
		}
	}

	bpjsLookup := domain.PayslipLookupBPJS{
		JHTEmployee: jhtDeduction,
		JPEmployee:  jpDeduction,
		KesEmployee: kesDeduction,
		TotalBPJS:   jhtDeduction + jpDeduction + kesDeduction,
	}

	// 8. Calculations
	totalEarning := basicSalary + totalAllowance
	takeHomePay := totalEarning - totalDeduction
	if takeHomePay < 0 {
		takeHomePay = 0
	}

	result := &domain.PayslipLookupResult{
		RequestID:      req.ID.String(),
		PeriodMonth:    req.PeriodMonth,
		PeriodYear:     req.PeriodYear,
		PeriodLabel:    periodLabel,
		Employee:       empLookup,
		BasicSalary:    basicSalary,
		Allowances:     allowances,
		TotalAllowance: totalAllowance,
		TotalEarning:   totalEarning,
		Attendance:     attendanceLookup,
		Leaves:         leaveLookup,
		BPJS:           bpjsLookup,
		Deductions:     deductions,
		TotalDeduction: totalDeduction,
		TakeHomePay:    takeHomePay,
	}

	return result, nil
}

func (u *payslipRequestUsecase) ApproveRequest(
	id uuid.UUID,
	nominal *float64,
	fileURL *string,
	payslipData *string,
	approvedBy *uuid.UUID,
) error {
	req, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}

	if req.Status != domain.PayslipRequestStatusPending {
		return errors.New("request is not pending")
	}

	now := time.Now()
	req.Status = domain.PayslipRequestStatusApproved
	req.ApprovedAt = &now
	if approvedBy != nil {
		req.ApprovedBy = approvedBy
	}
	if nominal != nil {
		req.Nominal = nominal
	}
	if fileURL != nil {
		req.FileURL = fileURL
	}
	if payslipData != nil {
		req.PayslipData = payslipData
	}

	return u.repo.Update(req)
}

func (u *payslipRequestUsecase) RejectRequest(id uuid.UUID, reason string) error {
	req, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}

	if req.Status != domain.PayslipRequestStatusPending {
		return errors.New("request is not pending")
	}

	req.Status = domain.PayslipRequestStatusRejected
	if reason != "" {
		req.RejectReason = &reason
	}

	return u.repo.Update(req)
}
