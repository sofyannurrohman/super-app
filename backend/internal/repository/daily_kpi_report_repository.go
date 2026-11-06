package repository

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type SubordinateActualKPI struct {
	EmployeeID       uuid.UUID `json:"employee_id"`
	EmployeeName     string    `json:"employee_name"`
	AreaRoute        string    `json:"area_route"`
	TargetCall       int       `json:"target_call"`
	RealizationCall  int       `json:"realization_call"`
	TargetEC         int       `json:"target_ec"`
	RealizationEC    int       `json:"realization_ec"`
	TargetSales      float64   `json:"target_sales"`
	RealizationSales float64   `json:"realization_sales"`
}

type DailyKpiFilter struct {
	Date               *time.Time
	StartDate          *time.Time
	EndDate            *time.Time
	ReporterEmployeeID *uuid.UUID
	ReviewerEmployeeID *uuid.UUID
	BranchID           *uuid.UUID
	ReporterRole       string
	Status             string
	Page               int
	Limit              int
}

type DailyKpiReportRepository interface {
	CreateReport(report *domain.DailyKpiReport) error
	UpdateReport(report *domain.DailyKpiReport) error
	DeleteReport(id uuid.UUID) error
	GetReportByID(id uuid.UUID) (*domain.DailyKpiReport, error)
	GetReports(filter DailyKpiFilter) ([]domain.DailyKpiReport, int64, error)
	GetSubordinateActualData(reporterID uuid.UUID, branchID *uuid.UUID, date time.Time) ([]SubordinateActualKPI, error)
}

type dailyKpiReportRepository struct {
	db *gorm.DB
}

func NewDailyKpiReportRepository(db *gorm.DB) DailyKpiReportRepository {
	return &dailyKpiReportRepository{db: db}
}

func (r *dailyKpiReportRepository) CreateReport(report *domain.DailyKpiReport) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(report).Error
	})
}

func (r *dailyKpiReportRepository) UpdateReport(report *domain.DailyKpiReport) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update header
		if err := tx.Save(report).Error; err != nil {
			return err
		}

		// Replace items
		if err := tx.Where("daily_kpi_report_id = ?", report.ID).Delete(&domain.DailyKpiReportItem{}).Error; err != nil {
			return err
		}
		if len(report.Items) > 0 {
			for i := range report.Items {
				report.Items[i].DailyKpiReportID = report.ID
			}
			if err := tx.Create(&report.Items).Error; err != nil {
				return err
			}
		}

		// Replace action plans
		if err := tx.Where("daily_kpi_report_id = ?", report.ID).Delete(&domain.DailyKpiReportActionPlan{}).Error; err != nil {
			return err
		}
		if len(report.ActionPlans) > 0 {
			for i := range report.ActionPlans {
				report.ActionPlans[i].DailyKpiReportID = report.ID
			}
			if err := tx.Create(&report.ActionPlans).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *dailyKpiReportRepository) DeleteReport(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		_ = tx.Where("daily_kpi_report_id = ?", id).Delete(&domain.DailyKpiReportItem{}).Error
		_ = tx.Where("daily_kpi_report_id = ?", id).Delete(&domain.DailyKpiReportActionPlan{}).Error
		return tx.Where("id = ?", id).Delete(&domain.DailyKpiReport{}).Error
	})
}

func (r *dailyKpiReportRepository) GetReportByID(id uuid.UUID) (*domain.DailyKpiReport, error) {
	var report domain.DailyKpiReport
	err := r.db.Preload("ReporterEmployee.User").
		Preload("ReviewerEmployee.User").
		Preload("Branch").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("employee_name ASC")
		}).
		Preload("ActionPlans", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ?", id).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *dailyKpiReportRepository) GetReports(filter DailyKpiFilter) ([]domain.DailyKpiReport, int64, error) {
	var reports []domain.DailyKpiReport
	var total int64

	query := r.db.Model(&domain.DailyKpiReport{}).
		Preload("ReporterEmployee.User").
		Preload("ReviewerEmployee.User").
		Preload("Branch")

	if filter.Date != nil {
		query = query.Where("report_date = ?", filter.Date.Format("2006-01-02"))
	} else if filter.StartDate != nil && filter.EndDate != nil {
		query = query.Where("report_date BETWEEN ? AND ?", filter.StartDate.Format("2006-01-02"), filter.EndDate.Format("2006-01-02"))
	}

	if filter.ReporterEmployeeID != nil {
		query = query.Where("reporter_employee_id = ?", *filter.ReporterEmployeeID)
	}

	if filter.ReviewerEmployeeID != nil {
		query = query.Where("reviewer_employee_id = ?", *filter.ReviewerEmployeeID)
	}

	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}

	if filter.ReporterRole != "" {
		query = query.Where("reporter_role = ?", filter.ReporterRole)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	err := query.Order("report_date DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	return reports, total, err
}

func (r *dailyKpiReportRepository) GetSubordinateActualData(reporterID uuid.UUID, branchID *uuid.UUID, date time.Time) ([]SubordinateActualKPI, error) {
	dateStr := date.Format("2006-01-02")
	startOfDay := dateStr + " 00:00:00"
	endOfDay := dateStr + " 23:59:59"

	// 1. Ambil daftar bawahan (subordinates)
	// Prioritas 1: Karyawan dengan manager_id = reporterID
	// Prioritas 2: Jika reporterID adalah SPV di cabang terkait, ambil seluruh Sales di cabang tersebut
	type empRow struct {
		ID                  uuid.UUID
		FirstName           string
		LastName            string
		WorkingTerritory    string
		BranchName          string
		OperatingBranchName string
		TargetDailyVisits   int
		TargetDailyNewStores int
		TargetOmzet         float64
	}

	var empRows []empRow
	month := int(date.Month())
	year := date.Year()

	// Query bawahan langsung via manager_id
	err := r.db.Table("employees e").
		Select(`
			e.id, e.first_name, e.last_name, 
			COALESCE(kpi.working_territory, '') as working_territory,
			COALESCE(b.name, '') as branch_name,
			COALESCE(ob.name, '') as operating_branch_name,
			COALESCE(kpi.target_daily_visits, 15) as target_daily_visits,
			COALESCE(kpi.target_daily_new_stores, 5) as target_daily_new_stores,
			COALESCE(kpi.target_omzet, 0) as target_omzet
		`).
		Joins("LEFT JOIN branches b ON b.id = e.branch_id").
		Joins("LEFT JOIN branches ob ON ob.id = e.operating_branch_id").
		Joins("LEFT JOIN sales_kpis kpi ON kpi.employee_id = e.id AND kpi.period_month = ? AND kpi.period_year = ?", month, year).
		Where("e.manager_id = ? AND e.employment_status NOT IN ('RESIGNED', 'TERMINATED')", reporterID).
		Order("e.first_name ASC").
		Scan(&empRows).Error

	if err != nil {
		return nil, err
	}

	// Jika belum ada karyawan yang di-assign secara direct manager_id ke reporter, dan ada branchID,
	// cari sales di cabang tersebut
	if len(empRows) == 0 && branchID != nil {
		_ = r.db.Table("employees e").
			Select(`
				e.id, e.first_name, e.last_name, 
				COALESCE(kpi.working_territory, '') as working_territory,
				COALESCE(b.name, '') as branch_name,
				COALESCE(ob.name, '') as operating_branch_name,
				COALESCE(kpi.target_daily_visits, 15) as target_daily_visits,
				COALESCE(kpi.target_daily_new_stores, 5) as target_daily_new_stores,
				COALESCE(kpi.target_omzet, 0) as target_omzet
			`).
			Joins("LEFT JOIN branches b ON b.id = e.branch_id").
			Joins("LEFT JOIN branches ob ON ob.id = e.operating_branch_id").
			Joins("LEFT JOIN job_positions jp ON jp.id = e.job_position_id").
			Joins("LEFT JOIN departments d ON d.id = e.department_id").
			Joins("LEFT JOIN sales_kpis kpi ON kpi.employee_id = e.id AND kpi.period_month = ? AND kpi.period_year = ?", month, year).
			Where("(e.branch_id = ? OR e.operating_branch_id = ?) AND e.employment_status NOT IN ('RESIGNED', 'TERMINATED')", *branchID, *branchID).
			Where("LOWER(jp.title) LIKE '%sales%' OR LOWER(d.name) LIKE '%sales%'").
			Where("e.id != ?", reporterID).
			Order("e.first_name ASC").
			Scan(&empRows).Error
	}

	results := make([]SubordinateActualKPI, 0, len(empRows))

	for _, emp := range empRows {
		empName := strings.TrimSpace(emp.FirstName + " " + emp.LastName)
		areaRoute := strings.TrimSpace(emp.WorkingTerritory)
		if areaRoute == "" {
			if emp.OperatingBranchName != "" {
				areaRoute = emp.OperatingBranchName
			} else {
				areaRoute = emp.BranchName
			}
		}

		// 1. Realisasi Calls (Kunjungan)
		var callCount int64
		_ = r.db.Table("sales_visits").
			Where("employee_id = ? AND check_in_time BETWEEN ? AND ?", emp.ID, startOfDay, endOfDay).
			Count(&callCount).Error

		// 2. Realisasi EC (Effective Call: outlet unik yang order pada hari tersebut) dan Realisasi Penjualan
		type ecRow struct {
			EffectiveStoreCount int64
			TotalAmount         float64
		}
		var ec ecRow
		_ = r.db.Raw(`
			SELECT 
				COUNT(DISTINCT store_id) as effective_store_count,
				COALESCE(SUM(total_amount), 0) as total_amount
			FROM (
				SELECT store_id, total_amount FROM sales_orders 
				WHERE employee_id = ? AND created_at BETWEEN ? AND ? AND status NOT IN ('CANCELLED', 'REJECTED', 'DRAFT')
				UNION ALL
				SELECT store_id, total_amount FROM sales_transactions 
				WHERE employee_id = ? AND transaction_date BETWEEN ? AND ? AND status NOT IN ('CANCELLED', 'REJECTED')
			) combined
		`, emp.ID, startOfDay, endOfDay, emp.ID, startOfDay, endOfDay).Scan(&ec).Error

		// Target Penjualan harian: Target bulanan dibagi 25 hari kerja jika ada
		dailySalesTarget := 0.0
		if emp.TargetOmzet > 0 {
			dailySalesTarget = emp.TargetOmzet / 25.0
		}

		targetCall := emp.TargetDailyVisits
		if targetCall <= 0 {
			targetCall = 15
		}
		targetEC := emp.TargetDailyNewStores
		if targetEC <= 0 {
			targetEC = 5
		}

		results = append(results, SubordinateActualKPI{
			EmployeeID:       emp.ID,
			EmployeeName:     empName,
			AreaRoute:        areaRoute,
			TargetCall:       targetCall,
			RealizationCall:  int(callCount),
			TargetEC:         targetEC,
			RealizationEC:    int(ec.EffectiveStoreCount),
			TargetSales:      dailySalesTarget,
			RealizationSales: ec.TotalAmount,
		})
	}

	return results, nil
}
