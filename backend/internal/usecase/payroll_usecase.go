package usecase

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

type PayrollUseCase interface {
	GeneratePayroll(companyID uuid.UUID, req GeneratePayrollRequest) (*domain.PayrollRun, error)
	GetMyPayslipHistory(userID uuid.UUID) ([]domain.MyPayslipResponse, error)
	ExportPayrollRunCSV(payrollRunID uuid.UUID) ([]byte, error)
	GetAllPayrollRuns(page, limit int, branchID *uuid.UUID) ([]domain.PayrollRun, error)
	DeletePayrollRun(id uuid.UUID) error

	// Salary Settings
	ListEmployeeSalarySettings(employeeID uuid.UUID) ([]domain.EmployeeSalarySetting, error)
	SaveEmployeeSalarySetting(setting *domain.EmployeeSalarySetting) error
	DeleteEmployeeSalarySetting(id uuid.UUID) error

	// Run Details
	GetPayrollRunDetails(runID uuid.UUID) ([]domain.Payslip, error)

	// HRD Approval & Adjustment Workflow
	ApprovePayrollRun(runID uuid.UUID) (*domain.PayrollRun, error)
	AdjustPayslip(slipID uuid.UUID, req AdjustPayslipRequest) (*domain.Payslip, error)

	// Bulk Transfer File & CMS Bank Excel
	GetTransferSummary(payrollRunID uuid.UUID) (*domain.PayrollTransferSummary, error)
	GenerateBankTransferFile(payrollRunID uuid.UUID, bankName string, format string) ([]byte, string, error)
	GenerateBankTransferExcel(payrollRunID uuid.UUID, bankName string, format string) ([]byte, string, error)
}

type payrollUseCase struct {
	repo            repository.PayrollRepository
	employeeRepo    repository.EmployeeRepository
	configRepo      repository.PayrollConfigRepository
	performanceRepo repository.PerformanceRepository
}

func NewPayrollUseCase(repo repository.PayrollRepository, employeeRepo repository.EmployeeRepository, configRepo repository.PayrollConfigRepository, performanceRepo repository.PerformanceRepository) PayrollUseCase {
	return &payrollUseCase{repo, employeeRepo, configRepo, performanceRepo}
}

type GeneratePayrollRequest struct {
	BranchID        *string `json:"branch_id"`                      // Optional UUID string, if empty/null => ALL branches
	PeriodStart     string  `json:"period_start" binding:"required"` // YYYY-MM-DD
	PeriodEnd       string  `json:"period_end" binding:"required"`
	PaymentSchedule string  `json:"payment_schedule" binding:"required"`
}

type AdjustPayslipRequest struct {
	Items []AdjustPayslipItem `json:"items" binding:"required"`
}

type AdjustPayslipItem struct {
	ComponentName string  `json:"component_name"`
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"` // EARNING, DEDUCTION
}

func (u *payrollUseCase) GeneratePayroll(companyID uuid.UUID, req GeneratePayrollRequest) (*domain.PayrollRun, error) {
	layout := "2006-01-02"
	start, err := time.Parse(layout, req.PeriodStart)
	if err != nil {
		return nil, errors.New("invalid period_start format, expected YYYY-MM-DD")
	}
	end, err := time.Parse(layout, req.PeriodEnd)
	if err != nil {
		return nil, errors.New("invalid period_end format, expected YYYY-MM-DD")
	}
	payment, err := time.Parse(layout, req.PaymentSchedule)
	if err != nil {
		return nil, errors.New("invalid payment_schedule format, expected YYYY-MM-DD")
	}

	if end.Before(start) {
		return nil, errors.New("period_end cannot be before period_start")
	}

	// Parse optional BranchID
	var branchUUID *uuid.UUID
	if req.BranchID != nil && *req.BranchID != "" && *req.BranchID != "ALL_BRANCHES" {
		parsed, err := uuid.Parse(*req.BranchID)
		if err == nil && parsed != uuid.Nil {
			branchUUID = &parsed
		}
	}

	payrollRunID := uuid.New()
	var grandTotalPayout float64
	var allPayslips []domain.Payslip

	employeeIDs, err := u.repo.GetActiveEmployees(branchUUID)
	if err != nil {
		return nil, errors.New("failed to retrieve active employees")
	}

	// Fetch Payroll Config for the company
	config, err := u.configRepo.GetByCompanyID(companyID)
	if err != nil || config == nil {
		// Default config if not found
		config = &domain.PayrollConfig{
			BPJSKesEmployeePercentage: 1.0,
			BPJSKesMaxWageBase:        12000000,
			JHTEmployeePercentage:     2.0,
			JPEmployeePercentage:      1.0,
			JPMaxWageBase:             10042300,
			AbsentDeduction:           100000.0,
		}
	}

	for _, empID := range employeeIDs {
		employee, err := u.employeeRepo.FindByID(empID)
		if err != nil || employee == nil {
			continue
		}

		settings, err := u.repo.GetEmployeeSalarySettings(empID)
		if err != nil {
			continue // Skip if no settings (or evaluate if we should still run just base salary)
		}

		var totalAllowance, totalDeduction, basicSalary float64
		var payslipItems []domain.PayslipItem

		// Inject the new base salary
		if employee.Salary != nil && *employee.Salary > 0 {
			basicSalary = *employee.Salary
			payslipItems = append(payslipItems, domain.PayslipItem{
				ID:            uuid.New(),
				PayslipID:     uuid.Nil,
				ComponentName: "Gaji Pokok",
				Amount:        basicSalary,
				Type:          "EARNING",
			})
		}

		for _, setting := range settings {
			if setting.Component == nil {
				continue
			}

			item := domain.PayslipItem{
				ID:            uuid.New(),
				PayslipID:     uuid.Nil, // Will be linked automatically by GORM relations or we set it manually below
				ComponentName: setting.Component.Name,
				Amount:        setting.Amount,
				Type:          setting.Component.Type,
			}

			if setting.Component.Type == "EARNING" {
				if setting.Component.Name == "Gaji Pokok" || setting.Component.Name == "Basic Salary" {
					if basicSalary == 0 {
						basicSalary = setting.Amount
					} else {
						// Skip if basicSalary is already populated from employee.Salary
						continue
					}
				} else {
					totalAllowance += setting.Amount
				}
			} else if setting.Component.Type == "DEDUCTION" {
				totalDeduction += setting.Amount
			}

			// Optional: Handle "BENEFIT" like insurance paid by company if needed, not affecting THP directly unless specified.
			payslipItems = append(payslipItems, item)
		}

		// Dynamic Deductions (BPJS & Ketenagakerjaan) based on Global Config
		if basicSalary > 0 {
			// BPJS Kesehatan
			kesBasis := basicSalary
			if config.BPJSKesMaxWageBase > 0 && kesBasis > config.BPJSKesMaxWageBase {
				kesBasis = config.BPJSKesMaxWageBase
			}
			kesDeduction := kesBasis * (config.BPJSKesEmployeePercentage / 100.0)
			if kesDeduction > 0 {
				totalDeduction += kesDeduction
				payslipItems = append(payslipItems, domain.PayslipItem{
					ID:            uuid.New(),
					PayslipID:     uuid.Nil,
					ComponentName: "BPJS Kesehatan (Karyawan)",
					Amount:        kesDeduction,
					Type:          "DEDUCTION",
				})
			}

			// JHT
			jhtDeduction := basicSalary * (config.JHTEmployeePercentage / 100.0)
			if jhtDeduction > 0 {
				totalDeduction += jhtDeduction
				payslipItems = append(payslipItems, domain.PayslipItem{
					ID:            uuid.New(),
					PayslipID:     uuid.Nil,
					ComponentName: "BPJS JHT (Karyawan)",
					Amount:        jhtDeduction,
					Type:          "DEDUCTION",
				})
			}

			// JP
			jpBasis := basicSalary
			if config.JPMaxWageBase > 0 && jpBasis > config.JPMaxWageBase {
				jpBasis = config.JPMaxWageBase
			}
			jpDeduction := jpBasis * (config.JPEmployeePercentage / 100.0)
			if jpDeduction > 0 {
				totalDeduction += jpDeduction
				payslipItems = append(payslipItems, domain.PayslipItem{
					ID:            uuid.New(),
					PayslipID:     uuid.Nil,
					ComponentName: "BPJS JP (Karyawan)",
					Amount:        jpDeduction,
					Type:          "DEDUCTION",
				})
			}
		}

		absentDays, err := u.repo.GetEmployeeAbsentDays(empID, start, end)
		if err == nil && absentDays > 0 {
			absentDeduction := float64(absentDays) * config.AbsentDeduction
			totalDeduction += absentDeduction

			payslipItems = append(payslipItems, domain.PayslipItem{
				ID:            uuid.New(),
				PayslipID:     uuid.Nil,
				ComponentName: "Potongan Mangkir",
				Amount:        absentDeduction,
				Type:          "DEDUCTION",
			})
		}

		// 8. Performance Bonus (KPI) - Only for FINALIZED regular KPIs
		periodMonth := int(start.Month())
		periodYear := start.Year()
		kpi, _ := u.performanceRepo.GetEmployeeKPIByEmployeeAndPeriod(empID, periodMonth, periodYear)
		if kpi != nil && kpi.Status == "FINALIZED" {
			// Calculation: Configurable KPI Bonus * (Score / 100)
			kpiBonus := (kpi.FinalScore / 100.0) * config.KPIBonusAmount
			if kpiBonus > 0 {
				totalAllowance += kpiBonus
				payslipItems = append(payslipItems, domain.PayslipItem{
					ID:            uuid.New(),
					PayslipID:     uuid.Nil,
					ComponentName: "Bonus Performa (KPI)",
					Amount:        kpiBonus,
					Type:          "EARNING",
				})
			}
		}

		takeHomePay := (basicSalary + totalAllowance) - totalDeduction
		grandTotalPayout += takeHomePay

		jobTitle, ptkp := u.repo.GetEmployeeSnapshotInfo(empID)

		payslipID := uuid.New()

		// Link proper ID to items manually for safety just in case
		for idx := range payslipItems {
			payslipItems[idx].PayslipID = payslipID
		}

		slip := domain.Payslip{
			ID:                 payslipID,
			PayrollRunID:       payrollRunID,
			EmployeeID:         empID,
			SnapshotJobTitle:   jobTitle,
			SnapshotPtkpStatus: ptkp,
			BasicSalary:        basicSalary,
			TotalAllowance:     totalAllowance,
			TotalDeduction:     totalDeduction,
			TakeHomePay:        takeHomePay,
			Items:              payslipItems,
		}

		allPayslips = append(allPayslips, slip)
	}

	if len(allPayslips) == 0 {
		return nil, errors.New("no employees with salary components were found to process")
	}

	payrollRun := &domain.PayrollRun{
		ID:              payrollRunID,
		BranchID:        branchUUID,
		CompanyID:       &companyID, // Assign to the admin's company requesting it
		PeriodStart:     start,
		PeriodEnd:       end,
		PaymentSchedule: &payment,
		Status:          "DRAFT", // HRD can review and adjust before approval
		TotalPayout:     &grandTotalPayout,
	}

	err = u.repo.SavePayrollBatch(payrollRun, allPayslips)
	if err != nil {
		return nil, errors.New("fatal sequence error saving payroll batch: " + err.Error())
	}

	return payrollRun, nil
}

func (u *payrollUseCase) GetMyPayslipHistory(userID uuid.UUID) ([]domain.MyPayslipResponse, error) {
	// 1. Resolve Employee from UserID
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("employee record not found for this user")
	}

	payslips, err := u.repo.GetMyPayslips(employee.ID)
	if err != nil {
		return nil, errors.New("failed to retrieve payslip data")
	}

	var response []domain.MyPayslipResponse
	if payslips == nil {
		return []domain.MyPayslipResponse{}, nil
	}

	for _, slip := range payslips {
		var earnings []domain.PayslipItemDetail
		var deductions []domain.PayslipItemDetail

		for _, item := range slip.Items {
			detail := domain.PayslipItemDetail{
				Name:   item.ComponentName,
				Amount: item.Amount,
			}

			if item.Type == "EARNING" {
				earnings = append(earnings, detail)
			} else if item.Type == "DEDUCTION" {
				deductions = append(deductions, detail)
			}
		}

		if earnings == nil {
			earnings = []domain.PayslipItemDetail{}
		}
		if deductions == nil {
			deductions = []domain.PayslipItemDetail{}
		}

		periodStr := ""
		paymentDate := ""
		if slip.PayrollRun != nil {
			periodStr = slip.PayrollRun.PeriodStart.Format("02 Jan 2006") + " - " + slip.PayrollRun.PeriodEnd.Format("02 Jan 2006")
			if slip.PayrollRun.PaymentSchedule != nil {
				paymentDate = slip.PayrollRun.PaymentSchedule.Format("2006-01-02")
			}
		}

		res := domain.MyPayslipResponse{
			ID:             slip.ID.String(),
			Period:         periodStr,
			PaymentDate:    paymentDate,
			JobTitle:       slip.SnapshotJobTitle,
			BasicSalary:    slip.BasicSalary,
			TotalAllowance: slip.TotalAllowance,
			TotalDeduction: slip.TotalDeduction,
			TakeHomePay:    slip.TakeHomePay,
			Earnings:       earnings,
			Deductions:     deductions,
		}

		response = append(response, res)
	}

	if response == nil {
		response = []domain.MyPayslipResponse{}
	}

	return response, nil
}

func (u *payrollUseCase) ExportPayrollRunCSV(payrollRunID uuid.UUID) ([]byte, error) {
	payslips, err := u.repo.GetPayslipsByRunID(payrollRunID)
	if err != nil {
		return nil, errors.New("failed to retrieve payslips for the given run ID")
	}

	if len(payslips) == 0 {
		return nil, errors.New("no payslips found for this payroll run")
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// CSV Header (MCM Mandiri / BCA standard format)
	header := []string{"No", "Account_Number", "Account_Name", "Bank_Name", "Amount", "Description"}
	if err := writer.Write(header); err != nil {
		return nil, errors.New("failed to write CSV header")
	}

	for i, slip := range payslips {
		emp := slip.Employee
		if emp == nil {
			continue // Avoid panic if relation fails to load
		}

		desc := fmt.Sprintf("Gaji %s", emp.FirstName)
		amountStr := fmt.Sprintf("%.2f", slip.TakeHomePay)

		row := []string{
			fmt.Sprintf("%d", i+1),
			emp.BankAccountNumber,
			emp.AccountHolderName,
			emp.BankName,
			amountStr,
			desc,
		}

		if err := writer.Write(row); err != nil {
			return nil, errors.New("failed to write CSV row")
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, errors.New("error flushing CSV writer")
	}

	return buf.Bytes(), nil
}

func (u *payrollUseCase) GetAllPayrollRuns(page, limit int, branchID *uuid.UUID) ([]domain.PayrollRun, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	return u.repo.GetAllPayrollRuns(limit, offset, branchID)
}

func (u *payrollUseCase) DeletePayrollRun(id uuid.UUID) error {
	return u.repo.DeletePayrollRun(id)
}

func (u *payrollUseCase) ListEmployeeSalarySettings(employeeID uuid.UUID) ([]domain.EmployeeSalarySetting, error) {
	return u.repo.GetEmployeeSalarySettings(employeeID)
}

func (u *payrollUseCase) SaveEmployeeSalarySetting(setting *domain.EmployeeSalarySetting) error {
	return u.repo.SaveEmployeeSalarySetting(setting)
}

func (u *payrollUseCase) DeleteEmployeeSalarySetting(id uuid.UUID) error {
	return u.repo.DeleteEmployeeSalarySetting(id)
}

func (u *payrollUseCase) GetPayrollRunDetails(runID uuid.UUID) ([]domain.Payslip, error) {
	return u.repo.GetPayslipsByRunID(runID)
}

// normalizeBankName menyeragamkan nama bank ke uppercase singkatan standar.
func normalizeBankName(raw string) string {
	upper := strings.ToUpper(strings.TrimSpace(raw))
	switch {
	case strings.Contains(upper, "BCA"):
		return "BCA"
	case strings.Contains(upper, "MANDIRI"):
		return "MANDIRI"
	case strings.Contains(upper, "BNI"):
		return "BNI"
	case strings.Contains(upper, "BRI"):
		return "BRI"
	case strings.Contains(upper, "CIMB"):
		return "CIMB"
	case strings.Contains(upper, "DANAMON"):
		return "DANAMON"
	case strings.Contains(upper, "PERMATA"):
		return "PERMATA"
	case strings.Contains(upper, "BTN"):
		return "BTN"
	case upper == "":
		return "TIDAK DIKETAHUI"
	default:
		return upper
	}
}

// GetTransferSummary menghasilkan ringkasan per bank dari sebuah payroll run.
func (u *payrollUseCase) GetTransferSummary(payrollRunID uuid.UUID) (*domain.PayrollTransferSummary, error) {
	payslips, err := u.repo.GetPayslipsByRunID(payrollRunID)
	if err != nil {
		return nil, errors.New("failed to retrieve payslips")
	}
	if len(payslips) == 0 {
		return nil, errors.New("no payslips found for this payroll run")
	}

	// Ambil info PayrollRun untuk field Period
	var run domain.PayrollRun
	for _, slip := range payslips {
		if slip.PayrollRun != nil {
			run = *slip.PayrollRun
			break
		}
	}

	groupMap := make(map[string]*domain.BankTransferGroup)
	var totalAmount float64
	totalRecords := 0
	missingBank := 0

	for _, slip := range payslips {
		emp := slip.Employee
		if emp == nil {
			continue
		}
		if emp.BankAccountNumber == "" {
			missingBank++
			continue
		}

		bankKey := normalizeBankName(emp.BankName)
		if _, ok := groupMap[bankKey]; !ok {
			groupMap[bankKey] = &domain.BankTransferGroup{BankName: bankKey}
		}
		groupMap[bankKey].TotalRecords++
		groupMap[bankKey].TotalAmount += slip.TakeHomePay
		totalAmount += slip.TakeHomePay
		totalRecords++
	}

	bankList := make([]domain.BankTransferGroup, 0, len(groupMap))
	for _, g := range groupMap {
		if missingBank > 0 {
			g.HasMissing = true
		}
		bankList = append(bankList, *g)
	}

	periodStr := run.PeriodStart.Format("02 Jan") + " - " + run.PeriodEnd.Format("02 Jan 2006")
	return &domain.PayrollTransferSummary{
		PayrollRunID: payrollRunID.String(),
		Period:       periodStr,
		TotalRecords: totalRecords,
		TotalAmount:  totalAmount,
		MissingBank:  missingBank,
		Banks:        bankList,
	}, nil
}

// GenerateBankTransferFile menghasilkan CSV file transfer gaji untuk bank tertentu.
// bankName: "BCA", "MANDIRI", "BNI", "ALL" (nama bank atau ALL untuk semua)
// format: "BCA", "MANDIRI", "GENERIC"
func (u *payrollUseCase) GenerateBankTransferFile(payrollRunID uuid.UUID, bankName string, format string) ([]byte, string, error) {
	payslips, err := u.repo.GetPayslipsByRunID(payrollRunID)
	if err != nil {
		return nil, "", errors.New("failed to retrieve payslips")
	}
	if len(payslips) == 0 {
		return nil, "", errors.New("no payslips found for this payroll run")
	}

	// Ambil tanggal untuk nama file
	var periodLabel string
	if len(payslips) > 0 && payslips[0].PayrollRun != nil {
		periodLabel = payslips[0].PayrollRun.PeriodStart.Format("Jan2006")
	} else {
		periodLabel = fmt.Sprintf("%d", time.Now().Unix())
	}

	normFormat := strings.ToUpper(strings.TrimSpace(format))
	normBank := strings.ToUpper(strings.TrimSpace(bankName))

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = ','

	// Tulis header sesuai format bank
	switch normFormat {
	case "BCA":
		_ = writer.Write([]string{"NO", "ACCOUNT_NUMBER", "ACCOUNT_HOLDER_NAME", "AMOUNT", "DESCRIPTION", "EMPLOYEE_ID"})
	case "MANDIRI":
		_ = writer.Write([]string{"NO_REKENING", "NOMINAL", "NAMA_PENERIMA", "KETERANGAN"})
	default: // GENERIC
		_ = writer.Write([]string{"NO", "BANK_NAME", "ACCOUNT_NUMBER", "ACCOUNT_HOLDER_NAME", "AMOUNT", "DESCRIPTION"})
	}

	rowNum := 1
	for _, slip := range payslips {
		emp := slip.Employee
		if emp == nil || emp.BankAccountNumber == "" {
			continue
		}

		// Filter berdasarkan bank jika bukan ALL
		if normBank != "ALL" && normBank != "" {
			if normalizeBankName(emp.BankName) != normBank {
				continue
			}
		}

		desc := fmt.Sprintf("GAJI %s", strings.ToUpper(periodLabel))
		amountStr := fmt.Sprintf("%.0f", slip.TakeHomePay) // Tanpa desimal, tanpa pemisah ribuan
		acctHolder := emp.AccountHolderName
		if acctHolder == "" {
			acctHolder = strings.ToUpper(emp.FirstName + " " + emp.LastName)
		}

		var row []string
		switch normFormat {
		case "BCA":
			row = []string{
				fmt.Sprintf("%d", rowNum),
				emp.BankAccountNumber,
				acctHolder,
				amountStr,
				desc,
				emp.EmployeeIDNumber,
			}
		case "MANDIRI":
			row = []string{
				emp.BankAccountNumber,
				amountStr,
				acctHolder,
				desc,
			}
		default: // GENERIC
			row = []string{
				fmt.Sprintf("%d", rowNum),
				normalizeBankName(emp.BankName),
				emp.BankAccountNumber,
				acctHolder,
				amountStr,
				desc,
			}
		}

		if err := writer.Write(row); err != nil {
			return nil, "", fmt.Errorf("failed to write row %d: %w", rowNum, err)
		}
		rowNum++
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", errors.New("error flushing CSV writer")
	}

	if rowNum == 1 {
		return nil, "", fmt.Errorf("no eligible employees found for bank '%s'", bankName)
	}

	// Generate filename: payroll_transfer_BCA_Apr2025.csv
	filename := fmt.Sprintf("payroll_transfer_%s_%s.csv", normBank, periodLabel)
	return buf.Bytes(), filename, nil
}

// ApprovePayrollRun menyetujui batch gaji dari DRAFT menjadi APPROVED
func (u *payrollUseCase) ApprovePayrollRun(runID uuid.UUID) (*domain.PayrollRun, error) {
	run, err := u.repo.GetPayrollRunByID(runID)
	if err != nil {
		return nil, errors.New("batch penggajian tidak ditemukan")
	}

	if run.Status == "APPROVED" || run.Status == "PROCESSED" {
		return nil, errors.New("batch penggajian sudah disetujui sebelumnya")
	}

	// Recalculate total payout based on any adjusted payslips
	payslips, err := u.repo.GetPayslipsByRunID(runID)
	if err == nil && len(payslips) > 0 {
		var grandTotal float64
		for _, s := range payslips {
			grandTotal += s.TakeHomePay
		}
		run.TotalPayout = &grandTotal
	}

	run.Status = "APPROVED"
	if err := u.repo.UpdatePayrollRun(run); err != nil {
		return nil, fmt.Errorf("gagal menyetujui batch: %w", err)
	}

	return run, nil
}

// AdjustPayslip memungkinkan HRD menyesuaikan komponen slip gaji (penambahan/pengurangan) sebelum approval
func (u *payrollUseCase) AdjustPayslip(slipID uuid.UUID, req AdjustPayslipRequest) (*domain.Payslip, error) {
	slip, err := u.repo.GetPayslipByID(slipID)
	if err != nil {
		return nil, errors.New("slip gaji tidak ditemukan")
	}

	run, err := u.repo.GetPayrollRunByID(slip.PayrollRunID)
	if err == nil && run.Status != "DRAFT" {
		return nil, errors.New("penyesuaian hanya bisa dilakukan pada batch berstatus DRAFT")
	}

	var newItems []domain.PayslipItem
	var totalAllowance, totalDeduction float64
	basicSalary := slip.BasicSalary

	for _, item := range req.Items {
		newItems = append(newItems, domain.PayslipItem{
			ID:            uuid.New(),
			PayslipID:     slip.ID,
			ComponentName: item.ComponentName,
			Amount:        item.Amount,
			Type:          item.Type,
		})

		if item.Type == "EARNING" {
			if item.ComponentName == "Gaji Pokok" || item.ComponentName == "Basic Salary" {
				basicSalary = item.Amount
			} else {
				totalAllowance += item.Amount
			}
		} else if item.Type == "DEDUCTION" {
			totalDeduction += item.Amount
		}
	}

	slip.BasicSalary = basicSalary
	slip.TotalAllowance = totalAllowance
	slip.TotalDeduction = totalDeduction
	slip.TakeHomePay = (basicSalary + totalAllowance) - totalDeduction

	if err := u.repo.UpdatePayslip(slip, newItems); err != nil {
		return nil, fmt.Errorf("gagal menyimpan penyesuaian slip gaji: %w", err)
	}

	// Update grand total in run
	allSlips, err := u.repo.GetPayslipsByRunID(slip.PayrollRunID)
	if err == nil {
		var grandTotal float64
		for _, s := range allSlips {
			grandTotal += s.TakeHomePay
		}
		run.TotalPayout = &grandTotal
		_ = u.repo.UpdatePayrollRun(run)
	}

	slip.Items = newItems
	return slip, nil
}

// GenerateBankTransferExcel menghasilkan file Excel (.xlsx) untuk CMS Bank Mass Transfer
func (u *payrollUseCase) GenerateBankTransferExcel(payrollRunID uuid.UUID, bankName string, format string) ([]byte, string, error) {
	payslips, err := u.repo.GetPayslipsByRunID(payrollRunID)
	if err != nil {
		return nil, "", errors.New("failed to retrieve payslips")
	}
	if len(payslips) == 0 {
		return nil, "", errors.New("no payslips found for this payroll run")
	}

	var periodLabel string
	var branchLabel string
	if len(payslips) > 0 && payslips[0].PayrollRun != nil {
		run := payslips[0].PayrollRun
		periodLabel = run.PeriodStart.Format("Jan 2006")
		if run.Branch != nil {
			branchLabel = run.Branch.Name
		} else {
			branchLabel = "SEMUA CABANG"
		}
	} else {
		periodLabel = time.Now().Format("Jan 2006")
		branchLabel = "CABANG"
	}

	normFormat := strings.ToUpper(strings.TrimSpace(format))
	normBank := strings.ToUpper(strings.TrimSpace(bankName))

	f := excelize.NewFile()
	sheetName := "Transfer CMS"
	f.SetSheetName("Sheet1", sheetName)

	// Styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 13, Color: "0F172A"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
		},
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
		},
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "1E293B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
		},
	})
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true, Size: 10, Color: "0F766E"},
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		CustomNumFmt: &[]string{"#,##0"}[0],
		Border: []excelize.Border{
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
		},
	})

	// Header Banner
	_ = f.SetCellValue(sheetName, "A1", fmt.Sprintf("BATCH PAYROLL CMS MASS TRANSFER — %s (%s)", strings.ToUpper(periodLabel), strings.ToUpper(branchLabel)))
	_ = f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	_ = f.SetRowHeight(sheetName, 1, 28)

	startRow := 3
	var headers []string

	switch normFormat {
	case "BCA":
		headers = []string{"NO", "NO_REKENING_TUJUAN", "NAMA_PENERIMA", "NOMINAL", "BERITA_TRANSFER", "ID_KARYAWAN"}
	case "MANDIRI":
		headers = []string{"NO", "NO_REKENING", "NOMINAL", "NAMA_PENERIMA", "KETERANGAN", "BANK_TUJUAN"}
	default: // GENERIC CMS
		headers = []string{"NO", "KODE_BANK", "NAMA_BANK", "NO_REKENING", "NAMA_PENERIMA", "NOMINAL_TRANSFER", "KETERANGAN", "NIK_KARYAWAN"}
	}

	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, startRow)
		_ = f.SetCellValue(sheetName, cell, h)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	_ = f.SetRowHeight(sheetName, startRow, 24)

	currentRow := startRow + 1
	seq := 1
	var totalTransfer float64

	for _, slip := range payslips {
		emp := slip.Employee
		if emp == nil || strings.TrimSpace(emp.BankAccountNumber) == "" {
			continue
		}

		if normBank != "ALL" && normBank != "" {
			if normalizeBankName(emp.BankName) != normBank {
				continue
			}
		}

		desc := fmt.Sprintf("GAJI %s %s", strings.ToUpper(periodLabel), strings.ToUpper(branchLabel))
		acctHolder := strings.TrimSpace(emp.AccountHolderName)
		if acctHolder == "" {
			acctHolder = strings.ToUpper(strings.TrimSpace(emp.FirstName + " " + emp.LastName))
		}

		totalTransfer += slip.TakeHomePay

		switch normFormat {
		case "BCA":
			_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), seq)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), emp.BankAccountNumber)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), acctHolder)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), slip.TakeHomePay)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), desc)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", currentRow), emp.EmployeeIDNumber)

			_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), centerStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("B%d", currentRow), fmt.Sprintf("B%d", currentRow), centerStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("C%d", currentRow), cellStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("D%d", currentRow), fmt.Sprintf("D%d", currentRow), currencyStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("E%d", currentRow), fmt.Sprintf("E%d", currentRow), cellStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("F%d", currentRow), fmt.Sprintf("F%d", currentRow), centerStyle)

		case "MANDIRI":
			_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), seq)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), emp.BankAccountNumber)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), slip.TakeHomePay)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), acctHolder)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), desc)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", currentRow), normalizeBankName(emp.BankName))

			_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), centerStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("B%d", currentRow), fmt.Sprintf("B%d", currentRow), centerStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("C%d", currentRow), currencyStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("D%d", currentRow), fmt.Sprintf("D%d", currentRow), cellStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("E%d", currentRow), fmt.Sprintf("E%d", currentRow), cellStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("F%d", currentRow), fmt.Sprintf("F%d", currentRow), centerStyle)

		default: // GENERIC
			_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), seq)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), normalizeBankName(emp.BankName))
			_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), emp.BankName)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), emp.BankAccountNumber)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), acctHolder)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", currentRow), slip.TakeHomePay)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("G%d", currentRow), desc)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("H%d", currentRow), emp.EmployeeIDNumber)

			_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), centerStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("B%d", currentRow), fmt.Sprintf("B%d", currentRow), centerStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("C%d", currentRow), cellStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("D%d", currentRow), fmt.Sprintf("D%d", currentRow), centerStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("E%d", currentRow), fmt.Sprintf("E%d", currentRow), cellStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("F%d", currentRow), fmt.Sprintf("F%d", currentRow), currencyStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("G%d", currentRow), fmt.Sprintf("G%d", currentRow), cellStyle)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("H%d", currentRow), fmt.Sprintf("H%d", currentRow), centerStyle)
		}

		_ = f.SetRowHeight(sheetName, currentRow, 20)
		currentRow++
		seq++
	}

	if seq == 1 {
		return nil, "", fmt.Errorf("tidak ada karyawan yang memiliki rekening bank valid untuk bank '%s'", bankName)
	}

	// Auto-fit column widths
	_ = f.SetColWidth(sheetName, "A", "A", 6)
	_ = f.SetColWidth(sheetName, "B", "B", 22)
	_ = f.SetColWidth(sheetName, "C", "C", 26)
	_ = f.SetColWidth(sheetName, "D", "D", 20)
	_ = f.SetColWidth(sheetName, "E", "E", 30)
	_ = f.SetColWidth(sheetName, "F", "F", 20)
	_ = f.SetColWidth(sheetName, "G", "G", 30)
	_ = f.SetColWidth(sheetName, "H", "H", 18)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", fmt.Errorf("gagal merender file excel: %w", err)
	}

	cleanBranch := strings.ReplaceAll(strings.TrimSpace(branchLabel), " ", "_")
	filename := fmt.Sprintf("CMS_Payroll_%s_%s_%s.xlsx", cleanBranch, normBank, strings.ReplaceAll(periodLabel, " ", "_"))
	return buf.Bytes(), filename, nil
}
