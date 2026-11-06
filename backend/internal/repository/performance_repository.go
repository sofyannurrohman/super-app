package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type PerformanceRepository interface {
	// Regular Employee KPIs
	GetEmployeeKPIs(month, year int) ([]domain.EmployeeKPI, error)
	GetEmployeeKPIByID(id uuid.UUID) (*domain.EmployeeKPI, error)
	GetEmployeeKPIByEmployeeAndPeriod(employeeID uuid.UUID, month, year int) (*domain.EmployeeKPI, error)
	SaveEmployeeKPI(kpi *domain.EmployeeKPI) error

	// Managerial Appraisals
	GetAppraisalsByEmployee(employeeID uuid.UUID) ([]domain.ManagerialAppraisal, error)
	SaveAppraisal(appraisal *domain.ManagerialAppraisal) error

	// Sales KPIs
	GetSalesKPIByEmployeeAndPeriod(employeeID uuid.UUID, month, year int) (*domain.SalesKPI, error)
	GetSalesKPIHistory(employeeID uuid.UUID) ([]domain.SalesKPI, error)
	GetSalesKPIReportByMonth(month, year int, branchIDs []uuid.UUID) ([]SalesKPIReport, error)
	GetSalesKPIReportByPeriod(periodType, dateStr string, month, year int, branchIDs []uuid.UUID) ([]SalesKPIReport, error)
	GetSalesKPIsByMonth(month, year int, branchIDs []uuid.UUID) ([]domain.SalesKPI, error)
	GetSalesKPIsByYear(year int, branchIDs []uuid.UUID) ([]domain.SalesKPI, error)
	GetSalesDailyKPIs(date time.Time, branchIDs []uuid.UUID) ([]domain.SalesKPI, error)
	GetSalesKPIsByDateRange(startDate, endDate time.Time, branchIDs []uuid.UUID) ([]domain.SalesKPI, error)

	SaveSalesKPI(kpi *domain.SalesKPI) error
	DeleteSalesKPI(id uuid.UUID) error

	// Single Employee KPI History
	GetEmployeeKPIHistory(employeeID uuid.UUID) ([]domain.EmployeeKPI, error)

	// Shared utilities (e.g., getting all regular employees)
	GetRegularEmployees() ([]domain.Employee, error)
	GetEmployeeByID(id uuid.UUID) (*domain.Employee, error)

	// Daily Omzet
	GetTodayOmzetByEmployee(employeeID uuid.UUID) (unvalidated float64, validated float64, err error)
	GetTodayTotalOrderValue(employeeID uuid.UUID) (float64, error)
	GetTodayVisitsCount(employeeID uuid.UUID) (int64, error)
	GetTodayNewStoresCount(employeeID uuid.UUID) (int64, error)
}

type performanceRepository struct {
	db *gorm.DB
}

func NewPerformanceRepository(db *gorm.DB) PerformanceRepository {
	return &performanceRepository{db}
}

func (r *performanceRepository) GetEmployeeKPIs(month, year int) ([]domain.EmployeeKPI, error) {
	var kpis []domain.EmployeeKPI
	err := r.db.Preload("Employee.User").Preload("Employee.JobPosition").
		Where("period_month = ? AND period_year = ?", month, year).
		Find(&kpis).Error
	return kpis, err
}

func (r *performanceRepository) GetEmployeeKPIByID(id uuid.UUID) (*domain.EmployeeKPI, error) {
	var kpi domain.EmployeeKPI
	err := r.db.Preload("Employee.User").Where("id = ?", id).First(&kpi).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &kpi, nil
}

func (r *performanceRepository) GetEmployeeKPIByEmployeeAndPeriod(employeeID uuid.UUID, month, year int) (*domain.EmployeeKPI, error) {
	var kpi domain.EmployeeKPI
	err := r.db.Where("employee_id = ? AND period_month = ? AND period_year = ?", employeeID, month, year).First(&kpi).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &kpi, nil
}

func (r *performanceRepository) SaveEmployeeKPI(kpi *domain.EmployeeKPI) error {
	return r.db.Save(kpi).Error
}

func (r *performanceRepository) GetAppraisalsByEmployee(employeeID uuid.UUID) ([]domain.ManagerialAppraisal, error) {
	var appraisals []domain.ManagerialAppraisal
	err := r.db.Preload("Manager.User").Where("employee_id = ?", employeeID).Order("review_date DESC").Find(&appraisals).Error
	return appraisals, err
}

func (r *performanceRepository) SaveAppraisal(appraisal *domain.ManagerialAppraisal) error {
	return r.db.Save(appraisal).Error
}

func (r *performanceRepository) GetRegularEmployees() ([]domain.Employee, error) {
	var employees []domain.Employee
	// Simple assumption: non-sales employees (we filter out sales roles if needed, for now get all active)
	err := r.db.Preload("User").Preload("JobPosition").
		Where("employment_status != ?", "TERMINATED").
		Find(&employees).Error
	return employees, err
}

func (r *performanceRepository) GetEmployeeByID(id uuid.UUID) (*domain.Employee, error) {
	var emp domain.Employee
	err := r.db.Preload("JobPosition").Where("id = ?", id).First(&emp).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *performanceRepository) GetSalesKPIByEmployeeAndPeriod(employeeID uuid.UUID, month, year int) (*domain.SalesKPI, error) {
	var kpi domain.SalesKPI
	err := r.db.Where("employee_id = ? AND period_month = ? AND period_year = ?", employeeID, month, year).First(&kpi).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &kpi, nil
}

func (r *performanceRepository) GetSalesKPIHistory(employeeID uuid.UUID) ([]domain.SalesKPI, error) {
	var kpis []domain.SalesKPI
	err := r.db.Where("employee_id = ?", employeeID).Order("period_year DESC, period_month DESC").Limit(12).Find(&kpis).Error
	return kpis, err
}

func (r *performanceRepository) GetEmployeeKPIHistory(employeeID uuid.UUID) ([]domain.EmployeeKPI, error) {
	var kpis []domain.EmployeeKPI
	err := r.db.Where("employee_id = ?", employeeID).Order("period_year DESC, period_month DESC").Limit(12).Find(&kpis).Error
	return kpis, err
}

func (r *performanceRepository) SaveSalesKPI(kpi *domain.SalesKPI) error {
	return r.db.Save(kpi).Error
}

func (r *performanceRepository) DeleteSalesKPI(id uuid.UUID) error {
	return r.db.Delete(&domain.SalesKPI{}, id).Error
}

type SalesKPIReport struct {
	EmployeeID    uuid.UUID `gorm:"column:employee_id"`
	TotalOmzet    float64   `gorm:"column:total_omzet"`
	OmzetLama     float64   `gorm:"column:omzet_lama"`
	OmzetBaru     float64   `gorm:"column:omzet_baru"`
	TotalTokoBaru int       `gorm:"column:total_toko_baru"`
	TotalVisits   int       `gorm:"column:total_visits"`
}

func (r *performanceRepository) GetSalesKPIReportByMonth(month, year int, branchIDs []uuid.UUID) ([]SalesKPIReport, error) {
	return r.GetSalesKPIReportByPeriod("monthly", "", month, year, branchIDs)
}

func (r *performanceRepository) GetSalesKPIReportByPeriod(periodType, dateStr string, month, year int, branchIDs []uuid.UUID) ([]SalesKPIReport, error) {
	var report []SalesKPIReport

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}

	var whereClauseTrx string
	var whereClausePayment string
	var argsTrx []interface{}
	var argsPayment []interface{}

	if periodType == "daily" || (periodType == "custom" && dateStr != "") {
		targetDate := time.Now().In(loc)
		if dateStr != "" {
			if strings.Contains(dateStr, ",") || strings.Contains(dateStr, ":") {
				// Formats like "2026-09-01,2026-09-28" or "2026-09-01:2026-09-28"
				sep := ","
				if strings.Contains(dateStr, ":") {
					sep = ":"
				}
				parts := strings.Split(dateStr, sep)
				startParsed, errS := time.ParseInLocation("2006-01-02", strings.TrimSpace(parts[0]), loc)
				endParsed, errE := time.ParseInLocation("2006-01-02", strings.TrimSpace(parts[1]), loc)
				if errS == nil && errE == nil {
					startOfDay := time.Date(startParsed.Year(), startParsed.Month(), startParsed.Day(), 0, 0, 0, 0, loc)
					endOfDay := time.Date(endParsed.Year(), endParsed.Month(), endParsed.Day(), 23, 59, 59, 999999999, loc)
					whereClauseTrx = "status = 'VERIFIED' AND transaction_date >= ? AND transaction_date <= ?"
					whereClausePayment = "sp.payment_status = 'SUCCESS' AND sp.penerimaan_status = 'VERIFIED' AND sp.payment_date >= ? AND sp.payment_date <= ?"
					argsTrx = append(argsTrx, startOfDay, endOfDay)
					argsPayment = append(argsPayment, startOfDay, endOfDay)
				}
			} else if parsed, errParse := time.ParseInLocation("2006-01-02", dateStr, loc); errParse == nil {
				targetDate = parsed
				startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc)
				endOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, loc)
				whereClauseTrx = "status = 'VERIFIED' AND transaction_date >= ? AND transaction_date <= ?"
				whereClausePayment = "sp.payment_status = 'SUCCESS' AND sp.penerimaan_status = 'VERIFIED' AND sp.payment_date >= ? AND sp.payment_date <= ?"
				argsTrx = append(argsTrx, startOfDay, endOfDay)
				argsPayment = append(argsPayment, startOfDay, endOfDay)
			}
		} else {
			startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc)
			endOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, loc)
			whereClauseTrx = "status = 'VERIFIED' AND transaction_date >= ? AND transaction_date <= ?"
			whereClausePayment = "sp.payment_status = 'SUCCESS' AND sp.penerimaan_status = 'VERIFIED' AND sp.payment_date >= ? AND sp.payment_date <= ?"
			argsTrx = append(argsTrx, startOfDay, endOfDay)
			argsPayment = append(argsPayment, startOfDay, endOfDay)
		}
	} else if periodType == "yearly" {
		whereClauseTrx = "status = 'VERIFIED' AND period_year = ?"
		whereClausePayment = "sp.payment_status = 'SUCCESS' AND sp.penerimaan_status = 'VERIFIED' AND EXTRACT(YEAR FROM sp.payment_date) = ?"
		argsTrx = append(argsTrx, year)
		argsPayment = append(argsPayment, year)
	} else {
		// default monthly
		whereClauseTrx = "status = 'VERIFIED' AND period_month = ? AND period_year = ?"
		whereClausePayment = "sp.payment_status = 'SUCCESS' AND sp.penerimaan_status = 'VERIFIED' AND EXTRACT(MONTH FROM sp.payment_date) = ? AND EXTRACT(YEAR FROM sp.payment_date) = ?"
		argsTrx = append(argsTrx, month, year)
		argsPayment = append(argsPayment, month, year)
	}

	query := fmt.Sprintf(`
		WITH regular_sales AS (
			SELECT 
				employee_id,
				store_category,
				total_amount,
				visit_id
			FROM sales_transactions
			WHERE %s
			  AND (payment_method != 'TEMPO' OR payment_method IS NULL OR NOT EXISTS (
			  	SELECT 1 FROM sales_payments sp_check 
			  	WHERE sp_check.sales_transaction_id = sales_transactions.id 
			  	  AND sp_check.penerimaan_status = 'VERIFIED'
			  ))
		),
		tempo_sales AS (
			SELECT 
				COALESCE(st.employee_id, sp.employee_id) as employee_id,
				COALESCE(st.store_category, 'TOKO_LAMA') as store_category,
				sp.amount as total_amount,
				st.visit_id
			FROM sales_payments sp
			JOIN sales_transactions st ON st.id = sp.sales_transaction_id
			WHERE %s
		),
		combined_sales AS (
			SELECT * FROM regular_sales
			UNION ALL
			SELECT * FROM tempo_sales
		)
		SELECT 
			c.employee_id,
			SUM(c.total_amount) as total_omzet,
			SUM(CASE WHEN c.store_category = 'TOKO_LAMA' THEN c.total_amount ELSE 0 END) as omzet_lama,
			SUM(CASE WHEN c.store_category = 'TOKO_BARU' THEN c.total_amount ELSE 0 END) as omzet_baru,
			COUNT(CASE WHEN c.store_category = 'TOKO_BARU' THEN 1 END) as total_toko_baru,
			COUNT(DISTINCT c.visit_id) as total_visits
		FROM combined_sales c
	`, whereClauseTrx, whereClausePayment)

	allArgs := append([]interface{}{}, argsTrx...)
	allArgs = append(allArgs, argsPayment...)

	if len(branchIDs) == 1 {
		query += ` JOIN employees e ON e.id = c.employee_id `
		query += ` WHERE COALESCE(e.operating_branch_id, e.branch_id) = ? `
		allArgs = append(allArgs, branchIDs[0])
	} else if len(branchIDs) > 1 {
		query += ` JOIN employees e ON e.id = c.employee_id `
		query += ` WHERE COALESCE(e.operating_branch_id, e.branch_id) IN (?) `
		allArgs = append(allArgs, branchIDs)
	}
	query += ` GROUP BY c.employee_id `

	err = r.db.Raw(query, allArgs...).Scan(&report).Error
	return report, err
}

func (r *performanceRepository) GetSalesKPIsByYear(year int, branchIDs []uuid.UUID) ([]domain.SalesKPI, error) {
	// Ambil semua KPI bulanan di tahun tersebut, lalu akumulasikan
	var allKpis []domain.SalesKPI
	query := r.db.Preload("Employee.User").Preload("Employee.JobPosition").Preload("Employee.Branch").Preload("Employee.OperatingBranch").
		Where("period_year = ?", year)

	if len(branchIDs) == 1 {
		query = query.Joins("JOIN employees ON employees.id = sales_kpis.employee_id").Where("COALESCE(employees.operating_branch_id, employees.branch_id) = ?", branchIDs[0])
	} else if len(branchIDs) > 1 {
		query = query.Joins("JOIN employees ON employees.id = sales_kpis.employee_id").Where("COALESCE(employees.operating_branch_id, employees.branch_id) IN (?)", branchIDs)
	}

	err := query.Find(&allKpis).Error
	if err != nil {
		return nil, err
	}

	empMap := make(map[uuid.UUID]*domain.SalesKPI)
	for _, kpi := range allKpis {
		if existing, ok := empMap[kpi.EmployeeID]; ok {
			existing.TargetOmzet += kpi.TargetOmzet
			existing.TargetNewStores += kpi.TargetNewStores
			existing.AchievedOmzet += kpi.AchievedOmzet
			existing.AchievedOmzetLama += kpi.AchievedOmzetLama
			existing.AchievedOmzetBaru += kpi.AchievedOmzetBaru
			existing.AchievedNewStores += kpi.AchievedNewStores
			existing.TotalVisits += kpi.TotalVisits
			existing.EstimatedBonus += kpi.EstimatedBonus
		} else {
			copyKpi := kpi
			empMap[kpi.EmployeeID] = &copyKpi
		}
	}

	var results []domain.SalesKPI
	for _, kpi := range empMap {
		results = append(results, *kpi)
	}

	return results, nil
}

func (r *performanceRepository) GetSalesKPIsByMonth(month, year int, branchIDs []uuid.UUID) ([]domain.SalesKPI, error) {
	var kpis []domain.SalesKPI
	query := r.db.Preload("Employee.User").Preload("Employee.JobPosition").Preload("Employee.Branch").Preload("Employee.OperatingBranch").
		Where("period_month = ? AND period_year = ?", month, year)

	if len(branchIDs) == 1 {
		query = query.Joins("JOIN employees ON employees.id = sales_kpis.employee_id").Where("COALESCE(employees.operating_branch_id, employees.branch_id) = ?", branchIDs[0])
	} else if len(branchIDs) > 1 {
		query = query.Joins("JOIN employees ON employees.id = sales_kpis.employee_id").Where("COALESCE(employees.operating_branch_id, employees.branch_id) IN (?)", branchIDs)
	}

	err := query.Find(&kpis).Error
	if err != nil {
		return nil, err
	}

	// Auto-Rollover Logic: Ensure ALL sales employees in this branch have KPIs for this month
	existingEmpKPIs := make(map[uuid.UUID]bool)
	for _, k := range kpis {
		existingEmpKPIs[k.EmployeeID] = true
	}

	// Get all active sales employees in this branch (exclude resigned/non-sales)
	var employees []domain.Employee
	empQuery := r.db.Joins("LEFT JOIN job_positions ON job_positions.id = employees.job_position_id").
		Joins("LEFT JOIN departments ON departments.id = employees.department_id").
		Where("(employees.employment_status IS NULL OR employees.employment_status NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED'))").
		Where("employees.first_name NOT ILIKE '%resign%' AND employees.last_name NOT ILIKE '%resign%' AND employees.first_name NOT ILIKE '%issued out%' AND employees.last_name NOT ILIKE '%issued out%'").
		Where("(job_positions.title ILIKE '%sales%' OR job_positions.title ILIKE '%marketing%' OR job_positions.title ILIKE '%motoris%' OR job_positions.title ILIKE '%booster%' OR job_positions.title ~* '\\mmt\\M' OR job_positions.title ~* '\\mto\\M' OR job_positions.title ILIKE '%take order%' OR departments.name ILIKE '%sales%' OR departments.name ILIKE '%marketing%')").
		Where("COALESCE(job_positions.title, '') NOT ILIKE '%operator%' AND COALESCE(job_positions.title, '') NOT ILIKE '%driver%' AND COALESCE(job_positions.title, '') NOT ILIKE '%auditor%' AND COALESCE(job_positions.title, '') NOT ILIKE '%editor%' AND COALESCE(job_positions.title, '') NOT ILIKE '%customer service%' AND COALESCE(job_positions.title, '') NOT ILIKE '%validator%' AND COALESCE(job_positions.title, '') NOT ILIKE '%gudang%' AND COALESCE(job_positions.title, '') NOT ILIKE '%security%'")
		
	if len(branchIDs) == 1 {
		empQuery = empQuery.Where("COALESCE(employees.operating_branch_id, employees.branch_id) = ?", branchIDs[0])
	} else if len(branchIDs) > 1 {
		empQuery = empQuery.Where("COALESCE(employees.operating_branch_id, employees.branch_id) IN (?)", branchIDs)
	}
	empQuery.Find(&employees)

	for _, emp := range employees {
		if existingEmpKPIs[emp.ID] {
			continue
		}

		// This employee doesn't have a KPI for this month.
		// Let's find their most recent KPI by looping backward 12 months.
		var prevKpi domain.SalesKPI
		searchMonth := month
		searchYear := year
		found := false

		for i := 0; i < 12; i++ {
			searchMonth--
			if searchMonth == 0 {
				searchMonth = 12
				searchYear--
			}
			err := r.db.Where("employee_id = ? AND period_month = ? AND period_year = ?", emp.ID, searchMonth, searchYear).First(&prevKpi).Error
			if err == nil {
				found = true
				break
			}
		}

		targetOmzet := 0.0
		targetNewStores := 0
		targetDailyVisits := 0
		targetDailyNewStores := 0
		workingTerritory := ""
		receivableLimitPercent := 0.0
		maxReceivableLimit := 0.0
		commission := 0.0
		if found {
			targetOmzet = prevKpi.TargetOmzet
			targetNewStores = prevKpi.TargetNewStores
			targetDailyVisits = prevKpi.TargetDailyVisits
			targetDailyNewStores = prevKpi.TargetDailyNewStores
			workingTerritory = prevKpi.WorkingTerritory
			receivableLimitPercent = prevKpi.ReceivableLimitPercent
			maxReceivableLimit = prevKpi.MaxReceivableLimit
			commission = prevKpi.Commission
		}

		newKpi := domain.SalesKPI{
			EmployeeID:             emp.ID,
			PeriodMonth:            month,
			PeriodYear:             year,
			TargetOmzet:            targetOmzet,
			TargetNewStores:        targetNewStores,
			TargetDailyVisits:      targetDailyVisits,
			TargetDailyNewStores:   targetDailyNewStores,
			WorkingTerritory:       workingTerritory,
			ReceivableLimitPercent: receivableLimitPercent,
			MaxReceivableLimit:     maxReceivableLimit,
			Commission:             commission,
			AchievedOmzet:          0,
			AchievedOmzetLama:      0,
			AchievedOmzetBaru:      0,
			AchievedNewStores:      0,
			TotalVisits:            0,
			EstimatedBonus:         0,
		}
		if r.db.Create(&newKpi).Error == nil {
			// To ensure all relations are loaded for the response, fetch the employee again with preloads
			var loadedEmp domain.Employee
			r.db.Preload("User").Preload("JobPosition").Preload("Branch").First(&loadedEmp, emp.ID)
			newKpi.Employee = &loadedEmp
			kpis = append(kpis, newKpi)
		}
	}

	// Dynamically calculate achievements for the month
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	for i := range kpis {
		empID := kpis[i].EmployeeID

		kpis[i].AchievedOmzetLama = 0
		kpis[i].AchievedOmzetBaru = 0
		kpis[i].AchievedOmzet = 0
		kpis[i].TotalVisits = 0
		kpis[i].AchievedNewStores = 0

		var trxTotal float64
		r.db.Table("sales_transactions").
			Select("COALESCE(SUM(total_amount), 0)").
			Where("employee_id = ? AND transaction_date >= ? AND transaction_date < ? AND status = ? AND store_category = ?", empID, startOfMonth, endOfMonth, "VERIFIED", "TOKO_LAMA").
			Scan(&trxTotal)
		kpis[i].AchievedOmzetLama = trxTotal

		var soTotal float64
		r.db.Table("sales_transactions").
			Select("COALESCE(SUM(total_amount), 0)").
			Where("employee_id = ? AND transaction_date >= ? AND transaction_date < ? AND status = ? AND store_category = ?", empID, startOfMonth, endOfMonth, "VERIFIED", "TOKO_BARU").
			Scan(&soTotal)
		kpis[i].AchievedOmzetBaru = soTotal

		kpis[i].AchievedOmzet = kpis[i].AchievedOmzetLama + kpis[i].AchievedOmzetBaru

		var visitCount int64
		r.db.Table("sales_visits").
			Select("COUNT(DISTINCT store_id || '_' || DATE_TRUNC('minute', check_in_time)::text)").
			Where("employee_id = ? AND check_in_time >= ? AND check_in_time < ?", empID, startOfMonth, endOfMonth).
			Scan(&visitCount)
		kpis[i].TotalVisits = int(visitCount)

		var newStoreCount int64
		r.db.Table("stores").
			Where("assigned_employee_id = ? AND created_at >= ? AND created_at < ?", empID, startOfMonth, endOfMonth).
			Count(&newStoreCount)
		kpis[i].AchievedNewStores = int(newStoreCount)
	}

	return kpis, nil
}

func (r *performanceRepository) GetSalesDailyKPIs(date time.Time, branchIDs []uuid.UUID) ([]domain.SalesKPI, error) {
	// First get the monthly KPIs for this date's month/year
	kpis, err := r.GetSalesKPIsByMonth(int(date.Month()), date.Year(), branchIDs)
	if err != nil {
		return nil, err
	}

	// Use Asia/Jakarta timezone so daily boundaries match WIB business hours
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.Local
	}
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	for i := range kpis {
		empID := kpis[i].EmployeeID

		// Reset achieved values to 0 for daily calculation
		kpis[i].AchievedOmzetLama = 0
		kpis[i].AchievedOmzetBaru = 0
		kpis[i].AchievedOmzet = 0
		kpis[i].TotalVisits = 0
		kpis[i].AchievedNewStores = 0

		// 1. Transactions (Toko Lama)
		var trxTotal float64
		r.db.Table("sales_transactions").
			Select("COALESCE(SUM(total_amount), 0)").
			Where("employee_id = ? AND created_at >= ? AND created_at < ? AND status = ? AND store_category = ?", empID, startOfDay, endOfDay, "VERIFIED", "TOKO_LAMA").
			Scan(&trxTotal)
		kpis[i].AchievedOmzetLama = trxTotal

		// 2. Transactions (Toko Baru)
		var soTotal float64
		r.db.Table("sales_transactions").
			Select("COALESCE(SUM(total_amount), 0)").
			Where("employee_id = ? AND created_at >= ? AND created_at < ? AND status = ? AND store_category = ?", empID, startOfDay, endOfDay, "VERIFIED", "TOKO_BARU").
			Scan(&soTotal)
		kpis[i].AchievedOmzetBaru = soTotal

		kpis[i].AchievedOmzet = kpis[i].AchievedOmzetLama + kpis[i].AchievedOmzetBaru

		// 3. Visits
		var visitCount int64
		r.db.Table("sales_visits").
			Select("COUNT(DISTINCT store_id || '_' || DATE_TRUNC('minute', check_in_time)::text)").
			Where("employee_id = ? AND check_in_time >= ? AND check_in_time < ?", empID, startOfDay, endOfDay).
			Scan(&visitCount)
		kpis[i].TotalVisits = int(visitCount)

		// 4. New Stores
		var newStoreCount int64
		r.db.Table("stores").
			Where("assigned_employee_id = ? AND created_at >= ? AND created_at < ?", empID, startOfDay, endOfDay).
			Count(&newStoreCount)
		kpis[i].AchievedNewStores = int(newStoreCount)
	}

	return kpis, nil
}

func (r *performanceRepository) GetSalesKPIsByDateRange(startDate, endDate time.Time, branchIDs []uuid.UUID) ([]domain.SalesKPI, error) {
	// First get base KPIs list (e.g., from startDate's month/year)
	kpis, err := r.GetSalesKPIsByMonth(int(startDate.Month()), startDate.Year(), branchIDs)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.Local
	}
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, loc)

	for i := range kpis {
		empID := kpis[i].EmployeeID

		kpis[i].AchievedOmzetLama = 0
		kpis[i].AchievedOmzetBaru = 0
		kpis[i].AchievedOmzet = 0
		kpis[i].TotalVisits = 0
		kpis[i].AchievedNewStores = 0

		// 1. Transactions (Toko Lama)
		var trxTotal float64
		r.db.Table("sales_transactions").
			Select("COALESCE(SUM(total_amount), 0)").
			Where("employee_id = ? AND transaction_date >= ? AND transaction_date <= ? AND status = ? AND store_category = ?", empID, start, end, "VERIFIED", "TOKO_LAMA").
			Where("(payment_method != 'TEMPO' OR payment_method IS NULL OR NOT EXISTS (SELECT 1 FROM sales_payments sp2 WHERE sp2.sales_transaction_id = sales_transactions.id AND sp2.penerimaan_status = 'VERIFIED'))").
			Scan(&trxTotal)

		var tempoTotalLama float64
		r.db.Table("sales_payments sp").
			Joins("JOIN sales_transactions st ON st.id = sp.sales_transaction_id").
			Select("COALESCE(SUM(sp.amount), 0)").
			Where("COALESCE(st.employee_id, sp.employee_id) = ? AND sp.payment_date >= ? AND sp.payment_date <= ? AND sp.payment_status = 'SUCCESS' AND sp.penerimaan_status = 'VERIFIED' AND st.store_category = 'TOKO_LAMA'", empID, start, end).
			Scan(&tempoTotalLama)
		kpis[i].AchievedOmzetLama = trxTotal + tempoTotalLama

		// 2. Transactions (Toko Baru)
		var soTotal float64
		r.db.Table("sales_transactions").
			Select("COALESCE(SUM(total_amount), 0)").
			Where("employee_id = ? AND transaction_date >= ? AND transaction_date <= ? AND status = ? AND store_category = ?", empID, start, end, "VERIFIED", "TOKO_BARU").
			Where("(payment_method != 'TEMPO' OR payment_method IS NULL OR NOT EXISTS (SELECT 1 FROM sales_payments sp2 WHERE sp2.sales_transaction_id = sales_transactions.id AND sp2.penerimaan_status = 'VERIFIED'))").
			Scan(&soTotal)

		var tempoTotalBaru float64
		r.db.Table("sales_payments sp").
			Joins("JOIN sales_transactions st ON st.id = sp.sales_transaction_id").
			Select("COALESCE(SUM(sp.amount), 0)").
			Where("COALESCE(st.employee_id, sp.employee_id) = ? AND sp.payment_date >= ? AND sp.payment_date <= ? AND sp.payment_status = 'SUCCESS' AND sp.penerimaan_status = 'VERIFIED' AND st.store_category = 'TOKO_BARU'", empID, start, end).
			Scan(&tempoTotalBaru)
		kpis[i].AchievedOmzetBaru = soTotal + tempoTotalBaru

		kpis[i].AchievedOmzet = kpis[i].AchievedOmzetLama + kpis[i].AchievedOmzetBaru

		// 3. Visits
		var visitCount int64
		r.db.Table("sales_visits").
			Select("COUNT(DISTINCT store_id || '_' || DATE_TRUNC('minute', check_in_time)::text)").
			Where("employee_id = ? AND check_in_time >= ? AND check_in_time <= ?", empID, start, end).
			Scan(&visitCount)
		kpis[i].TotalVisits = int(visitCount)

		// 4. New Stores
		var newStoreCount int64
		r.db.Table("stores").
			Where("assigned_employee_id = ? AND created_at >= ? AND created_at <= ?", empID, start, end).
			Count(&newStoreCount)
		kpis[i].AchievedNewStores = int(newStoreCount)
	}

	return kpis, nil
}

func (r *performanceRepository) GetTodayOmzetByEmployee(employeeID uuid.UUID) (float64, float64, error) {
	var unvalidated, validated float64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// 1. Validated (status = 'VERIFIED')
	r.db.Table("sales_transactions").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("employee_id = ? AND created_at >= ? AND created_at < ? AND status = ?", employeeID, startOfDay, endOfDay, "VERIFIED").
		Where("(payment_method != 'TEMPO' OR payment_method IS NULL OR NOT EXISTS (SELECT 1 FROM sales_payments sp2 WHERE sp2.sales_transaction_id = sales_transactions.id AND sp2.penerimaan_status = 'VERIFIED'))").
		Scan(&validated)

	var tempoValidated float64
	r.db.Table("sales_payments sp").
		Joins("JOIN sales_transactions st ON st.id = sp.sales_transaction_id").
		Select("COALESCE(SUM(sp.amount), 0)").
		Where("COALESCE(st.employee_id, sp.employee_id) = ? AND sp.payment_date >= ? AND sp.payment_date < ? AND sp.payment_status = 'SUCCESS' AND sp.penerimaan_status = 'VERIFIED'", employeeID, startOfDay, endOfDay).
		Scan(&tempoValidated)
	validated += tempoValidated

	// 2. Unvalidated sales transactions (PENDING or APPROVED)
	var trxUnvalidated float64
	r.db.Table("sales_transactions").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("employee_id = ? AND created_at >= ? AND created_at < ? AND status IN ?", employeeID, startOfDay, endOfDay, []string{"PENDING", "APPROVED"}).
		Scan(&trxUnvalidated)
	unvalidated += trxUnvalidated

	// 3. Unvalidated sales orders (not converted to transaction yet, and not cancelled/rejected)
	var soUnvalidated float64
	r.db.Table("sales_orders").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("employee_id = ? AND created_at >= ? AND created_at < ? AND status NOT IN ?", employeeID, startOfDay, endOfDay, []string{"CONVERTED", "REJECTED", "CANCELLED"}).
		Scan(&soUnvalidated)
	unvalidated += soUnvalidated

	return unvalidated, validated, nil
}

func (r *performanceRepository) GetTodayVisitsCount(employeeID uuid.UUID) (int64, error) {
	var count int64
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Kunjungan harian
	err = r.db.Table("sales_visits").
		Select("COUNT(DISTINCT store_id || '_' || DATE_TRUNC('minute', check_in_time)::text)").
		Where("employee_id = ? AND check_in_time >= ? AND check_in_time < ?", employeeID, startOfDay, endOfDay).
		Scan(&count).Error
	return count, err
}

func (r *performanceRepository) GetTodayNewStoresCount(employeeID uuid.UUID) (int64, error) {
	var count int64
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Toko baru yang diregistrasi oleh employee ini hari ini
	err = r.db.Table("stores").
		Where("assigned_employee_id = ? AND created_at >= ? AND created_at < ?", employeeID, startOfDay, endOfDay).
		Count(&count).Error
	return count, err
}

func (r *performanceRepository) GetTodayTotalOrderValue(employeeID uuid.UUID) (float64, error) {
	var totalTrx, totalSO float64
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Use the sale's business date, not created_at: an offline transaction can
	// reach the server the next day, and approval does not move it to that day.
	// Exclude transactions that originated from Sales Orders to prevent double counting.
	_ = r.db.Table("sales_transactions").
		Where("employee_id = ? AND (status IS NULL OR status != 'CANCELLED') AND transaction_date >= ? AND transaction_date < ?",
			employeeID, startOfDay, endOfDay).
		Where("payment_method != 'SALES_ORDER' AND NOT EXISTS (SELECT 1 FROM sales_orders so_ref WHERE so_ref.invoice_id = sales_transactions.id)").
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&totalTrx)

	// Sales Orders (Pre-Orders / TO) follow order_date even when synced later.
	_ = r.db.Table("sales_orders").
		Where("employee_id = ? AND (status IS NULL OR status NOT IN ('REJECTED', 'CANCELLED', 'DELETED')) AND order_date >= ? AND order_date < ?",
			employeeID, startOfDay, endOfDay).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&totalSO)

	return totalTrx + totalSO, nil
}

