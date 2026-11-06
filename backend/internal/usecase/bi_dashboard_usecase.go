package usecase

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---- Response Structs ----

type BIHRSummary struct {
	TotalEmployees   int     `json:"total_employees"`
	AttendanceRate   float64 `json:"attendance_rate"`   // % hadir hari ini
	PendingLeave     int     `json:"pending_leave"`     // cuti pending approval
	PendingOvertime  int     `json:"pending_overtime"`  // lembur pending
	PendingReimburse int     `json:"pending_reimburse"` // reimbursement pending
}

type BISalesSummary struct {
	TotalRevenue      float64 `json:"total_revenue"`
	TotalTransactions int     `json:"total_transactions"`
	AvgOrderValue     float64 `json:"avg_order_value"`
}

type BIPayrollSummary struct {
	TotalPayout        float64 `json:"total_payout"`
	TotalEmployeesPaid int     `json:"total_employees_paid"`
	LastRunDate        string  `json:"last_run_date"`
}

type BIMonthlyTrend struct {
	Label string  `json:"label"` // "Jan 2025"
	Value float64 `json:"value"`
}

type BISalespersonRank struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Rank   int     `json:"rank"`
}

type BIRevenueExpensePoint struct {
	Label   string  `json:"label"`
	Revenue float64 `json:"revenue"`
	Expense float64 `json:"expense"` // total gaji + reimbursement
	Profit  float64 `json:"profit"`
}

// ---- Interface ----

type BIDashboardUsecase interface {
	GetHRSummary(companyID uuid.UUID) (*BIHRSummary, error)
	GetSalesSummary(companyID uuid.UUID, month, year int) (*BISalesSummary, error)
	GetPayrollSummary(companyID uuid.UUID, month, year int) (*BIPayrollSummary, error)
	GetAttendanceTrend(companyID uuid.UUID, months int) ([]BIMonthlyTrend, error)
	GetSalesTrend(companyID uuid.UUID, months int) ([]BIMonthlyTrend, error)
	GetTopSalesperson(companyID uuid.UUID, month, year int, limit int) ([]BISalespersonRank, error)
	GetRevenueExpense(companyID uuid.UUID, months int) ([]BIRevenueExpensePoint, error)
}

// ---- Implementation ----

type biDashboardUsecase struct {
	db *gorm.DB
}

func NewBIDashboardUsecase(db *gorm.DB) BIDashboardUsecase {
	return &biDashboardUsecase{db: db}
}

// GetHRSummary — KPI box HR: total karyawan, kehadiran hari ini, pending requests
func (u *biDashboardUsecase) GetHRSummary(companyID uuid.UUID) (*BIHRSummary, error) {
	summary := &BIHRSummary{}
	today := time.Now().Format("2006-01-02")

	// Total karyawan aktif
	var totalEmp int64
	q := u.db.Table("employees").Where("employment_status NOT IN ('TERMINATED', 'RESIGNED')")
	if companyID != uuid.Nil {
		q = q.Where("company_id = ?", companyID)
	}
	q.Count(&totalEmp)
	summary.TotalEmployees = int(totalEmp)

	// Kehadiran hari ini (total hadir / total karyawan)
	var presentToday int64
	u.db.Table("attendance_logs al").
		Joins("JOIN employees e ON al.employee_id = e.id").
		Where("DATE(al.clock_in_time) = ? AND al.status IN ('ON_TIME', 'LATE')", today).
		Where(func() *gorm.DB {
			if companyID != uuid.Nil {
				return u.db.Where("e.company_id = ?", companyID)
			}
			return u.db
		}()).
		Count(&presentToday)

	if summary.TotalEmployees > 0 {
		summary.AttendanceRate = float64(presentToday) / float64(summary.TotalEmployees) * 100
	}

	// Pending leave
	var pendingLeave int64
	lq := u.db.Table("leave_requests lr").
		Joins("JOIN employees e ON lr.employee_id = e.id").
		Where("lr.status = 'PENDING'")
	if companyID != uuid.Nil {
		lq = lq.Where("e.company_id = ?", companyID)
	}
	lq.Count(&pendingLeave)
	summary.PendingLeave = int(pendingLeave)

	// Pending overtime
	var pendingOT int64
	oq := u.db.Table("overtimes o").
		Joins("JOIN employees e ON o.employee_id = e.id").
		Where("o.status = 'PENDING'")
	if companyID != uuid.Nil {
		oq = oq.Where("e.company_id = ?", companyID)
	}
	oq.Count(&pendingOT)
	summary.PendingOvertime = int(pendingOT)

	// Pending reimbursement
	var pendingReimburse int64
	rq := u.db.Table("reimbursements r").
		Joins("JOIN employees e ON r.employee_id = e.id").
		Where("r.status = 'PENDING'")
	if companyID != uuid.Nil {
		rq = rq.Where("e.company_id = ?", companyID)
	}
	rq.Count(&pendingReimburse)
	summary.PendingReimburse = int(pendingReimburse)

	return summary, nil
}

// GetSalesSummary — KPI box Sales: total omzet, jumlah transaksi, avg order value
func (u *biDashboardUsecase) GetSalesSummary(companyID uuid.UUID, month, year int) (*BISalesSummary, error) {
	type result struct {
		TotalRevenue      float64
		TotalTransactions int
	}
	var r result

	q := u.db.Table("sales_transactions st").
		Select("COALESCE(SUM(st.total_amount), 0) as total_revenue, COUNT(st.id) as total_transactions").
		Joins("JOIN stores s ON st.store_id = s.id").
		Where("EXTRACT(MONTH FROM st.transaction_date) = ? AND EXTRACT(YEAR FROM st.transaction_date) = ?", month, year).
		Where("st.status = 'VERIFIED'")
	if companyID != uuid.Nil {
		q = q.Where("s.company_id = ?", companyID)
	}

	if err := q.Scan(&r).Error; err != nil {
		return nil, err
	}

	avg := 0.0
	if r.TotalTransactions > 0 {
		avg = r.TotalRevenue / float64(r.TotalTransactions)
	}

	return &BISalesSummary{
		TotalRevenue:      r.TotalRevenue,
		TotalTransactions: r.TotalTransactions,
		AvgOrderValue:     avg,
	}, nil
}

// GetPayrollSummary — KPI box Payroll: total gaji dikeluarkan bulan ini
func (u *biDashboardUsecase) GetPayrollSummary(companyID uuid.UUID, month, year int) (*BIPayrollSummary, error) {
	type result struct {
		TotalPayout        float64
		TotalEmployeesPaid int
		LastRunDate        *time.Time
	}
	var r result

	q := u.db.Table("payroll_runs pr").
		Select(`
			COALESCE(SUM(pr.total_payout), 0) as total_payout,
			COALESCE(SUM((SELECT COUNT(p.id) FROM payslips p WHERE p.payroll_run_id = pr.id)), 0) as total_employees_paid,
			MAX(pr.created_at) as last_run_date
		`).
		Where("EXTRACT(MONTH FROM pr.period_start) = ? AND EXTRACT(YEAR FROM pr.period_start) = ?", month, year).
		Where("pr.status = 'PROCESSED'")

	if companyID != uuid.Nil {
		q = q.Where("pr.company_id = ?", companyID)
	}

	if err := q.Scan(&r).Error; err != nil {
		return nil, err
	}

	lastRunStr := "-"
	if r.LastRunDate != nil {
		lastRunStr = r.LastRunDate.Format("02 Jan 2006")
	}

	return &BIPayrollSummary{
		TotalPayout:        r.TotalPayout,
		TotalEmployeesPaid: r.TotalEmployeesPaid,
		LastRunDate:        lastRunStr,
	}, nil
}

// GetAttendanceTrend — Rate kehadiran per bulan untuk N bulan terakhir
func (u *biDashboardUsecase) GetAttendanceTrend(companyID uuid.UUID, months int) ([]BIMonthlyTrend, error) {
	if months <= 0 || months > 24 {
		months = 6
	}

	type row struct {
		MonthYear  string
		PresentCnt int
		TotalEmp   int
	}
	var rows []row

	q := u.db.Raw(`
		SELECT
			TO_CHAR(DATE_TRUNC('month', al.clock_in_time), 'Mon YYYY') AS month_year,
			COUNT(DISTINCT CASE WHEN al.status IN ('ON_TIME', 'LATE') THEN al.id END) AS present_cnt,
			COUNT(DISTINCT al.id) AS total_emp
		FROM attendance_logs al
		JOIN employees e ON al.employee_id = e.id
		WHERE al.clock_in_time >= NOW() - INTERVAL '? months'
		`+companyFilter(companyID, "e")+`
		GROUP BY DATE_TRUNC('month', al.clock_in_time)
		ORDER BY DATE_TRUNC('month', al.clock_in_time) ASC
	`, months)

	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]BIMonthlyTrend, 0, len(rows))
	for _, r := range rows {
		rate := 0.0
		if r.TotalEmp > 0 {
			rate = float64(r.PresentCnt) / float64(r.TotalEmp) * 100
		}
		result = append(result, BIMonthlyTrend{Label: r.MonthYear, Value: rate})
	}
	return result, nil
}

// GetSalesTrend — Total penjualan per bulan untuk N bulan terakhir
func (u *biDashboardUsecase) GetSalesTrend(companyID uuid.UUID, months int) ([]BIMonthlyTrend, error) {
	if months <= 0 || months > 24 {
		months = 6
	}

	type row struct {
		MonthYear   string
		TotalAmount float64
	}
	var rows []row

	q := u.db.Raw(fmt.Sprintf(`
		SELECT
			TO_CHAR(DATE_TRUNC('month', st.transaction_date), 'Mon YYYY') AS month_year,
			COALESCE(SUM(st.total_amount), 0) AS total_amount
		FROM sales_transactions st
		JOIN stores s ON st.store_id = s.id
		WHERE st.transaction_date >= NOW() - INTERVAL '%d months'
		  AND st.status = 'VERIFIED'
		  %s
		GROUP BY DATE_TRUNC('month', st.transaction_date)
		ORDER BY DATE_TRUNC('month', st.transaction_date) ASC
	`, months, companyFilter(companyID, "s")))

	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]BIMonthlyTrend, 0, len(rows))
	for _, r := range rows {
		result = append(result, BIMonthlyTrend{Label: r.MonthYear, Value: r.TotalAmount})
	}
	return result, nil
}

// GetTopSalesperson — Top N salesperson berdasarkan total penjualan bulan ini
func (u *biDashboardUsecase) GetTopSalesperson(companyID uuid.UUID, month, year int, limit int) ([]BISalespersonRank, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	type row struct {
		Name        string
		TotalAmount float64
	}
	var rows []row

	q := u.db.Raw(fmt.Sprintf(`
		SELECT
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS name,
			COALESCE(SUM(st.total_amount), 0) AS total_amount
		FROM sales_transactions st
		JOIN employees e ON st.employee_id = e.id
		JOIN stores s ON st.store_id = s.id
		WHERE EXTRACT(MONTH FROM st.transaction_date) = %d
		  AND EXTRACT(YEAR FROM st.transaction_date) = %d
		  AND st.status = 'VERIFIED'
		  %s
		GROUP BY e.id, e.first_name, e.last_name
		ORDER BY total_amount DESC
		LIMIT %d
	`, month, year, companyFilter(companyID, "s"), limit))

	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]BISalespersonRank, 0, len(rows))
	for i, r := range rows {
		result = append(result, BISalespersonRank{
			Rank:   i + 1,
			Name:   r.Name,
			Amount: r.TotalAmount,
		})
	}
	return result, nil
}

// GetRevenueExpense — Pendapatan vs Pengeluaran (gaji + reimbursement) per bulan
func (u *biDashboardUsecase) GetRevenueExpense(companyID uuid.UUID, months int) ([]BIRevenueExpensePoint, error) {
	if months <= 0 || months > 24 {
		months = 6
	}

	type salesRow struct {
		MonthYear   string
		TotalAmount float64
	}
	type expenseRow struct {
		MonthYear    string
		TotalPayroll float64
		TotalReimb   float64
	}

	var salesRows []salesRow
	var expenseRows []expenseRow

	// Sales revenue
	u.db.Raw(fmt.Sprintf(`
		SELECT
			TO_CHAR(DATE_TRUNC('month', st.transaction_date), 'Mon YYYY') AS month_year,
			COALESCE(SUM(st.total_amount), 0) AS total_amount
		FROM sales_transactions st
		JOIN stores s ON st.store_id = s.id
		WHERE st.transaction_date >= NOW() - INTERVAL '%d months'
		  AND st.status = 'VERIFIED'
		  %s
		GROUP BY DATE_TRUNC('month', st.transaction_date)
		ORDER BY DATE_TRUNC('month', st.transaction_date) ASC
	`, months, companyFilter(companyID, "s"))).Scan(&salesRows)

	// Payroll + reimbursement expense
	u.db.Raw(fmt.Sprintf(`
		SELECT
			TO_CHAR(DATE_TRUNC('month', pr.period_start), 'Mon YYYY') AS month_year,
			COALESCE(SUM(pr.total_payout), 0) AS total_payroll,
			COALESCE((
				SELECT SUM(r.amount)
				FROM reimbursements r
				JOIN employees e2 ON r.employee_id = e2.id
				WHERE DATE_TRUNC('month', r.created_at) = DATE_TRUNC('month', pr.period_start)
				  AND r.status = 'APPROVED'
				  %s
			), 0) AS total_reimb
		FROM payroll_runs pr
		WHERE pr.period_start >= NOW() - INTERVAL '%d months'
		  AND pr.status = 'PROCESSED'
		  %s
		GROUP BY DATE_TRUNC('month', pr.period_start)
		ORDER BY DATE_TRUNC('month', pr.period_start) ASC
	`, companyFilterRaw(companyID, "e2"), months, companyFilterPR(companyID))).Scan(&expenseRows)

	// Merge by month label
	salesMap := make(map[string]float64)
	for _, s := range salesRows {
		salesMap[s.MonthYear] = s.TotalAmount
	}

	result := make([]BIRevenueExpensePoint, 0, len(expenseRows))
	for _, ex := range expenseRows {
		rev := salesMap[ex.MonthYear]
		expense := ex.TotalPayroll + ex.TotalReimb
		result = append(result, BIRevenueExpensePoint{
			Label:   ex.MonthYear,
			Revenue: rev,
			Expense: expense,
			Profit:  rev - expense,
		})
	}
	return result, nil
}

// ---- Helper: Company filter snippets ----

func companyFilter(companyID uuid.UUID, alias string) string {
	if companyID == uuid.Nil {
		return ""
	}
	return fmt.Sprintf("AND %s.company_id = '%s'", alias, companyID.String())
}

func companyFilterRaw(companyID uuid.UUID, alias string) string {
	if companyID == uuid.Nil {
		return ""
	}
	return fmt.Sprintf("AND %s.company_id = '%s'", alias, companyID.String())
}

func companyFilterPR(companyID uuid.UUID) string {
	if companyID == uuid.Nil {
		return ""
	}
	return fmt.Sprintf("AND pr.company_id = '%s'", companyID.String())
}
