package http

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"example.com/hris/backend/internal/domain"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InternalReportHandler menyediakan endpoint internal khusus untuk WA Report Service
// dan Telegram Bot Service / Local LLM (Gemma 12B) yang berjalan di Mac Mini M4.
// Diproteksi dengan INTERNAL_API_KEY (bukan JWT).
//
// Endpoint:
// - GET /api/v1/internal/report-data       (Legacy WA summary)
// - GET /api/v1/internal/system-knowledge  (Comprehensive 360 Knowledge)
// - GET /api/v1/internal/hr-knowledge      (Deep HR data)
// - GET /api/v1/internal/sales-knowledge   (Deep Sales & AR data)
// - GET /api/v1/internal/logistics-knowledge (Delivery & Stock data)
type InternalReportHandler struct {
	db *gorm.DB
}

func NewInternalReportHandler(db *gorm.DB) *InternalReportHandler {
	return &InternalReportHandler{db: db}
}

// ─────────────────────────────────────────────────────────────────────────────
// Data Models untuk System Knowledge
// ─────────────────────────────────────────────────────────────────────────────

// ComprehensiveSystemKnowledge merepresentasikan seluruh status operasional sistem secara real-time.
type ComprehensiveSystemKnowledge struct {
	Timestamp   string               `json:"timestamp"`
	PeriodDate  string               `json:"period_date"`
	HR          HRKnowledge          `json:"hr"`
	Sales       SalesKnowledge       `json:"sales"`
	Logistics   LogisticsKnowledge   `json:"logistics"`
	Finance     FinanceKnowledge     `json:"finance"`
	Marketplace MarketplaceKnowledge `json:"marketplace"`
}

// HR Knowledge
type HRKnowledge struct {
	TotalActiveEmployees int                   `json:"total_active_employees"`
	PresentToday         int                   `json:"present_today"`
	LateToday            int                   `json:"late_today"`
	OnLeaveToday         int                   `json:"on_leave_today"`
	LateEmployees        []LateEmployeeInfo    `json:"late_employees"`
	ActiveLeaves         []ActiveLeaveInfo     `json:"active_leaves"`
	PendingApprovals     PendingApprovalCounts `json:"pending_approvals"`
	BranchHeadcounts     []BranchHeadcount     `json:"branch_headcounts"`
	DepartmentHeadcounts []DepartmentHeadcount `json:"department_headcounts"`
}

type LateEmployeeInfo struct {
	EmployeeName string `json:"employee_name"`
	BranchName   string `json:"branch_name"`
	ClockInTime  string `json:"clock_in_time"`
	MinutesLate  int    `json:"minutes_late"`
}

type ActiveLeaveInfo struct {
	EmployeeName string `json:"employee_name"`
	BranchName   string `json:"branch_name"`
	LeaveType    string `json:"leave_type"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Reason       string `json:"reason"`
}

type PendingApprovalCounts struct {
	PendingLeaveCount         int     `json:"pending_leave_count"`
	PendingOvertimeCount      int     `json:"pending_overtime_count"`
	PendingReimbursementCount int     `json:"pending_reimbursement_count"`
	PendingReimbursementTotal float64 `json:"pending_reimbursement_total"`
}

type BranchHeadcount struct {
	BranchName   string `json:"branch_name"`
	PresentToday int    `json:"present_today"`
	LateToday    int    `json:"late_today"`
	OnLeaveToday int    `json:"on_leave_today"`
	TotalActive  int    `json:"total_active"`
}

type DepartmentHeadcount struct {
	DepartmentName string `json:"department_name"`
	TotalEmployees int    `json:"total_employees"`
}

// Sales Knowledge
type SalesKnowledge struct {
	TodaySalesAmount    float64           `json:"today_sales_amount"`
	TodayTransactions   int               `json:"today_transactions"`
	MTDSalesAmount      float64           `json:"mtd_sales_amount"`
	MTDTransactions     int               `json:"mtd_transactions"`
	TopSalesmenToday    []SalesmanRank    `json:"top_salesmen_today"`
	TopSalesmenMTD      []SalesmanRank    `json:"top_salesmen_mtd"`
	BranchPerformance   []BranchSalesPerf `json:"branch_performance"`
	PendingOrdersCount  int               `json:"pending_orders_count"`
	OverdueReceivables  float64           `json:"overdue_receivables"`
	OverdueInvoiceCount int               `json:"overdue_invoice_count"`
	NewStoresThisMonth  int               `json:"new_stores_this_month"`
}

type SalesmanRank struct {
	SalesmanName string  `json:"salesman_name"`
	BranchName   string  `json:"branch_name"`
	TotalAmount  float64 `json:"total_amount"`
	Transactions int     `json:"transactions"`
}

type BranchSalesPerf struct {
	BranchName   string  `json:"branch_name"`
	TodaySales   float64 `json:"today_sales"`
	MTDSales     float64 `json:"mtd_sales"`
	Transactions int     `json:"transactions"`
}

// Logistics Knowledge
type LogisticsKnowledge struct {
	CompletedDeliveriesToday int               `json:"completed_deliveries_today"`
	TotalDeliveriesToday     int               `json:"total_deliveries_today"`
	OngoingDeliveries        []OngoingDelivery `json:"ongoing_deliveries"`
	LowStockAlerts           []LowStockAlert   `json:"low_stock_alerts"`
}

type OngoingDelivery struct {
	BatchOrderNo string `json:"batch_order_no"`
	DriverName   string `json:"driver_name"`
	VehiclePlate string `json:"vehicle_plate"`
	Status       string `json:"status"`
	TotalItems   int    `json:"total_items"`
	BranchName   string `json:"branch_name"`
}

type LowStockAlert struct {
	ProductName string `json:"product_name"`
	SKU         string `json:"sku"`
	CurrentQty  int    `json:"current_qty"`
	Unit        string `json:"unit"`
	Warehouse   string `json:"warehouse"`
}

// Finance Knowledge
type FinanceKnowledge struct {
	PendingReimbursementAmount float64           `json:"pending_reimbursement_amount"`
	PendingReimbursementCount  int               `json:"pending_reimbursement_count"`
	BankAccounts               []BankAccountInfo `json:"bank_accounts"`
	KasPabrikBalance           float64           `json:"kas_pabrik_balance"`
}

type BankAccountInfo struct {
	BankName      string  `json:"bank_name"`
	AccountNumber string  `json:"account_number"`
	AccountHolder string  `json:"account_holder"`
	Balance       float64 `json:"balance"`
	BranchName    string  `json:"branch_name"`
}

// Marketplace Knowledge
type MarketplaceKnowledge struct {
	PendingPackingOrders int     `json:"pending_packing_orders"`
	ShippedTodayOrders   int     `json:"shipped_today_orders"`
	TodayGMV             float64 `json:"today_gmv"`
}

// Legacy Response (dipertahankan untuk kompatibilitas wa-report-service)
type ReportDataResponse struct {
	ReportPeriod         string             `json:"report_period"`
	TotalEmployees       int                `json:"total_employees"`
	PresentToday         int                `json:"present_today"`
	LateToday            int                `json:"late_today"`
	OnLeaveToday         int                `json:"on_leave_today"`
	AbsentToday          int                `json:"absent_today"`
	PendingLeaveReqs     int                `json:"pending_leave_reqs"`
	PendingOvertimeReqs  int                `json:"pending_overtime_reqs"`
	PendingReimbursement int                `json:"pending_reimbursement"`
	TotalSalesToday      float64            `json:"total_sales_today"`
	TotalTransactions    int                `json:"total_transactions"`
	TopSalesmanName      string             `json:"top_salesman_name"`
	TopSalesmanAmount    float64            `json:"top_salesman_amount"`
	PendingOrders        int                `json:"pending_orders"`
	CompletedDeliveries  int                `json:"completed_deliveries"`
	TotalDeliveries      int                `json:"total_deliveries"`
	BranchData           []BranchReportData `json:"branch_data"`
}

type BranchReportData struct {
	BranchName   string `json:"branch_name"`
	PresentToday int    `json:"present_today"`
	OnLeaveToday int    `json:"on_leave_today"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Route Registration
// ─────────────────────────────────────────────────────────────────────────────

func (h *InternalReportHandler) RegisterInternalRoutes(r *gin.RouterGroup) {
	internal := r.Group("/internal")
	internal.Use(h.internalKeyMiddleware())
	{
		internal.GET("/report-data", h.GetReportData)
		internal.GET("/system-knowledge", h.GetSystemKnowledge)
		internal.GET("/hr-knowledge", h.GetHRKnowledge)
		internal.GET("/sales-knowledge", h.GetSalesKnowledge)
		internal.GET("/logistics-knowledge", h.GetLogisticsKnowledge)
		internal.GET("/employee-search", h.SearchEmployees)
		internal.GET("/marketplace-summary", h.GetMarketplaceSummary)
	}
}

// SearchEmployees adalah endpoint read-only untuk assistant internal.
// Hanya mengembalikan identitas pekerjaan yang aman untuk konteks AI.
func (h *InternalReportHandler) SearchEmployees(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 || len(query) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q harus berisi 2-100 karakter"})
		return
	}
	pattern := "%" + query + "%"
	var employees []domain.Employee
	err := h.db.Model(&domain.Employee{}).
		Preload("Branch").Preload("Department").Preload("JobPosition").
		Where("employees.first_name ILIKE ? OR employees.last_name ILIKE ? OR CONCAT(employees.first_name, ' ', COALESCE(employees.last_name, '')) ILIKE ? OR employees.employee_id_number ILIKE ?", pattern, pattern, pattern, pattern).
		Where("COALESCE(employees.employment_status, '') NOT ILIKE ?", "%inactive%").
		Limit(20).Find(&employees).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search employees"})
		return
	}
	type result struct {
		ID         string `json:"id"`
		EmployeeID string `json:"employee_id"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		Branch     string `json:"branch,omitempty"`
		Department string `json:"department,omitempty"`
		Position   string `json:"position,omitempty"`
	}
	results := make([]result, 0, len(employees))
	for _, employee := range employees {
		name := strings.TrimSpace(employee.FirstName + " " + employee.LastName)
		item := result{ID: employee.ID.String(), EmployeeID: employee.EmployeeIDNumber, Name: name, Status: employee.EmploymentStatus}
		if employee.Branch != nil {
			item.Branch = employee.Branch.Name
		}
		if employee.Department != nil {
			item.Department = employee.Department.Name
		}
		if employee.JobPosition != nil {
			item.Position = employee.JobPosition.Title
		}
		results = append(results, item)
	}
	c.JSON(http.StatusOK, gin.H{"query": query, "count": len(results), "data": results})
}

// GetMarketplaceSummary returns branch-filtered marketplace metrics for AI tools.
func (h *InternalReportHandler) GetMarketplaceSummary(c *gin.Context) {
	branch := strings.TrimSpace(c.Query("branch"))
	if len(branch) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch terlalu panjang"})
		return
	}
	dateScope := strings.ToLower(strings.TrimSpace(c.DefaultQuery("date", "today")))
	if dateScope != "today" && dateScope != "all" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date harus today atau all"})
		return
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")

	type warehouseItem struct {
		Name   string `json:"name"`
		Orders int64  `json:"orders"`
	}

	type summaryScalar struct {
		Branch         string  `json:"branch"`
		TotalOrders    int64   `json:"total_orders"`
		PendingPacking int64   `json:"pending_packing"`
		Shipped        int64   `json:"shipped"`
		GMV            float64 `json:"gmv"`
	}

	var conditions []string
	var args []interface{}

	if branch != "" {
		conditions = append(conditions, "LOWER(b.name) = LOWER(?)")
		args = append(args, branch)
	}

	if dateScope == "today" {
		conditions = append(conditions, "DATE(mo.created_at) = ?")
		args = append(args, todayStr)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	summarySQL := fmt.Sprintf(`
		SELECT 
			COALESCE(MAX(b.name), ?) AS branch,
			COUNT(DISTINCT mo.order_id) AS total_orders,
			COUNT(DISTINCT CASE WHEN mo.order_status IN ('AWAITING_SHIPMENT','UNPAID','PENDING') THEN mo.order_id END) AS pending_packing,
			COUNT(DISTINCT CASE WHEN mo.order_status IN ('SHIPPED','DELIVERED','COMPLETED') THEN mo.order_id END) AS shipped,
			COALESCE(SUM(mo.total_amount), 0) AS gmv
		FROM marketplace_orders mo
		LEFT JOIN branches b ON b.id = mo.branch_id
		%s
	`, whereClause)

	var summary summaryScalar
	summaryArgs := append([]interface{}{branch}, args...)
	if err := h.db.Raw(summarySQL, summaryArgs...).Scan(&summary).Error; err != nil {
		log.Printf("[AI Marketplace] summary query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query marketplace summary"})
		return
	}

	warehouseSQL := fmt.Sprintf(`
		SELECT 
			COALESCE(NULLIF(TRIM(mo.warehouse_name), ''), 'Tanpa Warehouse') AS name,
			COUNT(DISTINCT mo.order_id) AS orders
		FROM marketplace_orders mo
		LEFT JOIN branches b ON b.id = mo.branch_id
		%s
		GROUP BY COALESCE(NULLIF(TRIM(mo.warehouse_name), ''), 'Tanpa Warehouse')
		ORDER BY orders DESC
	`, whereClause)

	var warehouses []warehouseItem
	if err := h.db.Raw(warehouseSQL, args...).Scan(&warehouses).Error; err != nil {
		log.Printf("[AI Marketplace] warehouse breakdown query error: %v", err)
		warehouses = []warehouseItem{}
	}
	if warehouses == nil {
		warehouses = []warehouseItem{}
	}

	out := struct {
		Branch         string          `json:"branch"`
		TotalOrders    int64           `json:"total_orders"`
		PendingPacking int64           `json:"pending_packing"`
		Shipped        int64           `json:"shipped"`
		GMV            float64         `json:"gmv"`
		Warehouses     []warehouseItem `json:"warehouses"`
	}{
		Branch:         summary.Branch,
		TotalOrders:    summary.TotalOrders,
		PendingPacking: summary.PendingPacking,
		Shipped:        summary.Shipped,
		GMV:            summary.GMV,
		Warehouses:     warehouses,
	}

	c.JSON(http.StatusOK, gin.H{"data": out, "branch_requested": branch, "date_scope": dateScope, "read_only": true})
}

func (h *InternalReportHandler) internalKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedKey := os.Getenv("INTERNAL_API_KEY")
		if expectedKey == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Internal API not configured"})
			c.Abort()
			return
		}

		key := c.GetHeader("X-Internal-Key")
		if key != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid internal API key"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Handler Implementations
// ─────────────────────────────────────────────────────────────────────────────

// GetSystemKnowledge mengumpulkan snapshot komprehensif dari semua pilar operasional.
func (h *InternalReportHandler) GetSystemKnowledge(c *gin.Context) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")
	monthStartStr := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).Format("2006-01-02")

	resp := ComprehensiveSystemKnowledge{
		Timestamp:  now.Format("2006-01-02 15:04:05 WIB"),
		PeriodDate: now.Format("Monday, 02 January 2006"),
	}

	// ── 1. HR KNOWLEDGE ────────────────────────────────────────────────────────
	h.db.Raw(`
		SELECT COUNT(DISTINCT al.id) FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		WHERE DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		  AND UPPER(al.status) IN ('PRESENT', 'ON_TIME')
		  AND e.employment_status != 'resigned'
	`, todayStr).Scan(&resp.HR.PresentToday)

	h.db.Raw(`
		SELECT COUNT(DISTINCT al.id) FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		WHERE DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		  AND UPPER(al.status) = 'LATE'
		  AND e.employment_status != 'resigned'
	`, todayStr).Scan(&resp.HR.LateToday)

	h.db.Raw(`
		SELECT COUNT(*) FROM leave_requests
		WHERE status = 'APPROVED'
		  AND ? BETWEEN DATE(start_date) AND DATE(end_date)
	`, todayStr).Scan(&resp.HR.OnLeaveToday)

	resp.HR.TotalActiveEmployees = resp.HR.PresentToday + resp.HR.LateToday + resp.HR.OnLeaveToday

	// Detail Karyawan Terlambat Hari Ini
	h.db.Raw(`
		SELECT 
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS employee_name,
			COALESCE(b.name, 'Pusat') AS branch_name,
			TO_CHAR(al.clock_in_time AT TIME ZONE 'Asia/Jakarta', 'HH24:MI') AS clock_in_time,
			GREATEST(0, EXTRACT(EPOCH FROM (al.clock_in_time AT TIME ZONE 'Asia/Jakarta' - DATE_TRUNC('day', al.clock_in_time AT TIME ZONE 'Asia/Jakarta') - INTERVAL '8 hours'))::int / 60) AS minutes_late
		FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		LEFT JOIN branches b ON b.id = e.branch_id
		WHERE DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		  AND UPPER(al.status) = 'LATE'
		  AND e.employment_status != 'resigned'
		ORDER BY al.clock_in_time ASC
		LIMIT 15
	`, todayStr).Scan(&resp.HR.LateEmployees)

	// Detail Karyawan Cuti Aktif Hari Ini
	h.db.Raw(`
		SELECT 
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS employee_name,
			COALESCE(b.name, 'Pusat') AS branch_name,
			COALESCE(lt.name, 'Cuti') AS leave_type,
			TO_CHAR(lr.start_date, 'YYYY-MM-DD') AS start_date,
			TO_CHAR(lr.end_date, 'YYYY-MM-DD') AS end_date,
			COALESCE(lr.reason, '-') AS reason
		FROM leave_requests lr
		JOIN employees e ON e.id = lr.employee_id
		LEFT JOIN branches b ON b.id = e.branch_id
		LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id
		WHERE lr.status = 'APPROVED'
		  AND ? BETWEEN DATE(lr.start_date) AND DATE(lr.end_date)
		LIMIT 20
	`, todayStr).Scan(&resp.HR.ActiveLeaves)

	// Pending Approvals
	h.db.Raw(`SELECT COUNT(*) FROM leave_requests WHERE status = 'PENDING'`).Scan(&resp.HR.PendingApprovals.PendingLeaveCount)
	h.db.Raw(`SELECT COUNT(*) FROM overtimes WHERE status = 'pending'`).Scan(&resp.HR.PendingApprovals.PendingOvertimeCount)
	h.db.Raw(`SELECT COUNT(*) FROM reimbursements WHERE status = 'PENDING'`).Scan(&resp.HR.PendingApprovals.PendingReimbursementCount)
	h.db.Raw(`SELECT COALESCE(SUM(amount), 0) FROM reimbursements WHERE status = 'PENDING'`).Scan(&resp.HR.PendingApprovals.PendingReimbursementTotal)

	// Branch Headcounts
	h.db.Raw(`
		SELECT 
			b.name AS branch_name,
			COUNT(DISTINCT CASE WHEN UPPER(al.status) IN ('PRESENT', 'ON_TIME') THEN al.id END) AS present_today,
			COUNT(DISTINCT CASE WHEN UPPER(al.status) = 'LATE' THEN al.id END) AS late_today,
			COUNT(DISTINCT CASE WHEN lr.status = 'APPROVED' AND ? BETWEEN DATE(lr.start_date) AND DATE(lr.end_date) THEN lr.id END) AS on_leave_today,
			COUNT(DISTINCT e.id) AS total_active
		FROM branches b
		JOIN employees e ON e.branch_id = b.id
		LEFT JOIN attendance_logs al ON al.employee_id = e.id AND DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		LEFT JOIN leave_requests lr ON lr.employee_id = e.id AND lr.status = 'APPROVED' AND ? BETWEEN DATE(lr.start_date) AND DATE(lr.end_date)
		WHERE e.employment_status != 'resigned'
		GROUP BY b.name
		ORDER BY b.name ASC
	`, todayStr, todayStr, todayStr).Scan(&resp.HR.BranchHeadcounts)

	// Department Headcounts
	h.db.Raw(`
		SELECT 
			COALESCE(d.name, 'Umum') AS department_name,
			COUNT(e.id) AS total_employees
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE e.employment_status != 'resigned'
		GROUP BY d.name
		ORDER BY total_employees DESC
	`).Scan(&resp.HR.DepartmentHeadcounts)

	// ── 2. SALES KNOWLEDGE ─────────────────────────────────────────────────────
	// Sales Hari Ini
	h.db.Raw(`
		SELECT COALESCE(SUM(total_amount), 0) FROM sales_transactions
		WHERE DATE(transaction_date) = ?
	`, todayStr).Scan(&resp.Sales.TodaySalesAmount)

	h.db.Raw(`
		SELECT COUNT(*) FROM sales_transactions
		WHERE DATE(transaction_date) = ?
	`, todayStr).Scan(&resp.Sales.TodayTransactions)

	// Sales MTD (Month to Date)
	h.db.Raw(`
		SELECT COALESCE(SUM(total_amount), 0) FROM sales_transactions
		WHERE DATE(transaction_date) >= ? AND DATE(transaction_date) <= ?
	`, monthStartStr, todayStr).Scan(&resp.Sales.MTDSalesAmount)

	h.db.Raw(`
		SELECT COUNT(*) FROM sales_transactions
		WHERE DATE(transaction_date) >= ? AND DATE(transaction_date) <= ?
	`, monthStartStr, todayStr).Scan(&resp.Sales.MTDTransactions)

	// Top Salesmen Hari Ini
	h.db.Raw(`
		SELECT 
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS salesman_name,
			COALESCE(b.name, '-') AS branch_name,
			COALESCE(SUM(st.total_amount), 0) AS total_amount,
			COUNT(st.id) AS transactions
		FROM sales_transactions st
		JOIN employees e ON e.id = st.employee_id
		LEFT JOIN branches b ON b.id = st.branch_id
		WHERE DATE(st.transaction_date) = ?
		GROUP BY e.first_name, e.last_name, b.name
		ORDER BY total_amount DESC
		LIMIT 5
	`, todayStr).Scan(&resp.Sales.TopSalesmenToday)

	// Top Salesmen MTD
	h.db.Raw(`
		SELECT 
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS salesman_name,
			COALESCE(b.name, '-') AS branch_name,
			COALESCE(SUM(st.total_amount), 0) AS total_amount,
			COUNT(st.id) AS transactions
		FROM sales_transactions st
		JOIN employees e ON e.id = st.employee_id
		LEFT JOIN branches b ON b.id = st.branch_id
		WHERE DATE(st.transaction_date) >= ? AND DATE(st.transaction_date) <= ?
		GROUP BY e.first_name, e.last_name, b.name
		ORDER BY total_amount DESC
		LIMIT 5
	`, monthStartStr, todayStr).Scan(&resp.Sales.TopSalesmenMTD)

	// Performa Cabang MTD
	h.db.Raw(`
		SELECT 
			COALESCE(b.name, 'Pusat') AS branch_name,
			COALESCE(SUM(CASE WHEN DATE(st.transaction_date) = ? THEN st.total_amount ELSE 0 END), 0) AS today_sales,
			COALESCE(SUM(st.total_amount), 0) AS mtd_sales,
			COUNT(st.id) AS transactions
		FROM sales_transactions st
		LEFT JOIN branches b ON b.id = st.branch_id
		WHERE DATE(st.transaction_date) >= ? AND DATE(st.transaction_date) <= ?
		GROUP BY b.name
		ORDER BY mtd_sales DESC
	`, todayStr, monthStartStr, todayStr).Scan(&resp.Sales.BranchPerformance)

	// Pending Orders (PO yang belum dikirim / diverifikasi)
	h.db.Raw(`SELECT COUNT(*) FROM sales_orders WHERE status IN ('DRAFT', 'CONFIRMED')`).Scan(&resp.Sales.PendingOrdersCount)

	// Piutang Tempo Jatuh Tempo
	h.db.Raw(`
		SELECT 
			COALESCE(SUM(total_amount - paid_amount), 0),
			COUNT(*)
		FROM sales_transactions
		WHERE payment_method = 'TEMPO' 
		  AND payment_status != 'PAID'
		  AND payment_due_date < ?
	`, todayStr).Row().Scan(&resp.Sales.OverdueReceivables, &resp.Sales.OverdueInvoiceCount)

	// Toko Baru (NOO) Bulan Ini
	h.db.Raw(`
		SELECT COUNT(*) FROM stores
		WHERE DATE(created_at) >= ? AND DATE(created_at) <= ?
	`, monthStartStr, todayStr).Scan(&resp.Sales.NewStoresThisMonth)

	// ── 3. LOGISTICS & WMS KNOWLEDGE ───────────────────────────────────────────
	h.db.Raw(`
		SELECT COUNT(*) FROM delivery_batches
		WHERE DATE(delivery_date) = ? AND status = 'COMPLETED'
	`, todayStr).Scan(&resp.Logistics.CompletedDeliveriesToday)

	h.db.Raw(`
		SELECT COUNT(*) FROM delivery_batches
		WHERE DATE(delivery_date) = ?
	`, todayStr).Scan(&resp.Logistics.TotalDeliveriesToday)

	// Pengiriman Aktif / Sedang Jalan Hari Ini
	h.db.Raw(`
		SELECT 
			COALESCE(db.delivery_order_no, 'BATCH-' || SUBSTRING(db.id::text, 1, 8)) AS batch_order_no,
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS driver_name,
			COALESCE(v.plate_number, '-') AS vehicle_plate,
			db.status AS status,
			(SELECT COUNT(*) FROM delivery_items di WHERE di.delivery_batch_id = db.id) AS total_items,
			COALESCE(b.name, '-') AS branch_name
		FROM delivery_batches db
		LEFT JOIN employees e ON e.id = db.driver_id
		LEFT JOIN vehicles v ON v.id = db.vehicle_id
		LEFT JOIN branches b ON b.id = db.branch_id
		WHERE db.status IN ('PICKING', 'ON_DELIVERY', 'SUPERVISOR_APPROVED', 'WAITING_APPROVAL')
		ORDER BY db.created_at DESC
		LIMIT 10
	`).Scan(&resp.Logistics.OngoingDeliveries)

	// Low Stock Alerts (Stok Pabrik / Gudang < 50 pcs)
	h.db.Raw(`
		SELECT 
			p.name AS product_name,
			p.sku AS sku,
			COALESCE(SUM(fs.quantity), 0) AS current_qty,
			p.unit AS unit,
			COALESCE(fw.name, 'Gudang Utama') AS warehouse
		FROM products p
		LEFT JOIN factory_stocks fs ON fs.product_id = p.id
		LEFT JOIN factory_warehouses fw ON fw.id = fs.factory_warehouse_id
		GROUP BY p.name, p.sku, p.unit, fw.name
		HAVING COALESCE(SUM(fs.quantity), 0) < 50
		ORDER BY current_qty ASC
		LIMIT 10
	`).Scan(&resp.Logistics.LowStockAlerts)

	// ── 4. FINANCE KNOWLEDGE ───────────────────────────────────────────────────
	resp.Finance.PendingReimbursementAmount = resp.HR.PendingApprovals.PendingReimbursementTotal
	resp.Finance.PendingReimbursementCount = resp.HR.PendingApprovals.PendingReimbursementCount

	// Rekening Bank Perusahaan & Saldo
	h.db.Raw(`
		SELECT 
			cba.bank_name,
			cba.account_number,
			cba.account_holder,
			COALESCE(cba.manual_balance, 0) AS balance,
			COALESCE(b.name, 'Pusat') AS branch_name
		FROM company_bank_accounts cba
		LEFT JOIN branches b ON b.id = cba.branch_id
		WHERE cba.is_active = true
		ORDER BY cba.manual_balance DESC
	`).Scan(&resp.Finance.BankAccounts)

	// Kas Pabrik
	h.db.Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM kas_pabriks
	`).Scan(&resp.Finance.KasPabrikBalance)

	// ── 5. MARKETPLACE KNOWLEDGE ───────────────────────────────────────────────
	h.db.Raw(`
		SELECT COUNT(*) FROM marketplace_orders
		WHERE order_status IN ('AWAITING_SHIPMENT', 'UNPAID', 'PENDING')
	`).Scan(&resp.Marketplace.PendingPackingOrders)

	h.db.Raw(`
		SELECT COUNT(*) FROM marketplace_orders
		WHERE order_status IN ('SHIPPED', 'DELIVERED', 'COMPLETED')
		  AND DATE(created_at) = ?
	`, todayStr).Scan(&resp.Marketplace.ShippedTodayOrders)

	h.db.Raw(`
		SELECT COALESCE(SUM(total_amount), 0) FROM marketplace_orders
		WHERE DATE(created_at) = ?
	`, todayStr).Scan(&resp.Marketplace.TodayGMV)

	c.JSON(http.StatusOK, resp)
}

// GetHRKnowledge mengembalikan data HR mendalam untuk LLM query spesifik HR.
func (h *InternalReportHandler) GetHRKnowledge(c *gin.Context) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")

	var hr HRKnowledge
	h.db.Raw(`
		SELECT COUNT(DISTINCT al.id) FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		WHERE DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		  AND UPPER(al.status) IN ('PRESENT', 'ON_TIME')
		  AND e.employment_status != 'resigned'
	`, todayStr).Scan(&hr.PresentToday)

	h.db.Raw(`
		SELECT COUNT(DISTINCT al.id) FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		WHERE DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		  AND UPPER(al.status) = 'LATE'
		  AND e.employment_status != 'resigned'
	`, todayStr).Scan(&hr.LateToday)

	h.db.Raw(`
		SELECT COUNT(*) FROM leave_requests
		WHERE status = 'APPROVED'
		  AND ? BETWEEN DATE(start_date) AND DATE(end_date)
	`, todayStr).Scan(&hr.OnLeaveToday)

	hr.TotalActiveEmployees = hr.PresentToday + hr.LateToday + hr.OnLeaveToday

	h.db.Raw(`
		SELECT 
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS employee_name,
			COALESCE(b.name, 'Pusat') AS branch_name,
			TO_CHAR(al.clock_in_time AT TIME ZONE 'Asia/Jakarta', 'HH24:MI') AS clock_in_time,
			GREATEST(0, EXTRACT(EPOCH FROM (al.clock_in_time AT TIME ZONE 'Asia/Jakarta' - DATE_TRUNC('day', al.clock_in_time AT TIME ZONE 'Asia/Jakarta') - INTERVAL '8 hours'))::int / 60) AS minutes_late
		FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		LEFT JOIN branches b ON b.id = e.branch_id
		WHERE DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		  AND UPPER(al.status) = 'LATE'
		  AND e.employment_status != 'resigned'
		ORDER BY al.clock_in_time ASC
	`, todayStr).Scan(&hr.LateEmployees)

	h.db.Raw(`
		SELECT 
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS employee_name,
			COALESCE(b.name, 'Pusat') AS branch_name,
			COALESCE(lt.name, 'Cuti') AS leave_type,
			TO_CHAR(lr.start_date, 'YYYY-MM-DD') AS start_date,
			TO_CHAR(lr.end_date, 'YYYY-MM-DD') AS end_date,
			COALESCE(lr.reason, '-') AS reason
		FROM leave_requests lr
		JOIN employees e ON e.id = lr.employee_id
		LEFT JOIN branches b ON b.id = e.branch_id
		LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id
		WHERE lr.status = 'APPROVED'
		  AND ? BETWEEN DATE(lr.start_date) AND DATE(lr.end_date)
	`, todayStr).Scan(&hr.ActiveLeaves)

	h.db.Raw(`SELECT COUNT(*) FROM leave_requests WHERE status = 'PENDING'`).Scan(&hr.PendingApprovals.PendingLeaveCount)
	h.db.Raw(`SELECT COUNT(*) FROM overtimes WHERE status = 'pending'`).Scan(&hr.PendingApprovals.PendingOvertimeCount)
	h.db.Raw(`SELECT COUNT(*) FROM reimbursements WHERE status = 'PENDING'`).Scan(&hr.PendingApprovals.PendingReimbursementCount)
	h.db.Raw(`SELECT COALESCE(SUM(amount), 0) FROM reimbursements WHERE status = 'PENDING'`).Scan(&hr.PendingApprovals.PendingReimbursementTotal)

	c.JSON(http.StatusOK, hr)
}

// GetSalesKnowledge mengembalikan data penjualan dan piutang mendalam.
func (h *InternalReportHandler) GetSalesKnowledge(c *gin.Context) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")
	monthStartStr := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).Format("2006-01-02")

	var sales SalesKnowledge
	h.db.Raw(`SELECT COALESCE(SUM(total_amount), 0) FROM sales_transactions WHERE DATE(transaction_date) = ?`, todayStr).Scan(&sales.TodaySalesAmount)
	h.db.Raw(`SELECT COUNT(*) FROM sales_transactions WHERE DATE(transaction_date) = ?`, todayStr).Scan(&sales.TodayTransactions)
	h.db.Raw(`SELECT COALESCE(SUM(total_amount), 0) FROM sales_transactions WHERE DATE(transaction_date) >= ? AND DATE(transaction_date) <= ?`, monthStartStr, todayStr).Scan(&sales.MTDSalesAmount)
	h.db.Raw(`SELECT COUNT(*) FROM sales_transactions WHERE DATE(transaction_date) >= ? AND DATE(transaction_date) <= ?`, monthStartStr, todayStr).Scan(&sales.MTDTransactions)

	h.db.Raw(`
		SELECT 
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS salesman_name,
			COALESCE(b.name, '-') AS branch_name,
			COALESCE(SUM(st.total_amount), 0) AS total_amount,
			COUNT(st.id) AS transactions
		FROM sales_transactions st
		JOIN employees e ON e.id = st.employee_id
		LEFT JOIN branches b ON b.id = st.branch_id
		WHERE DATE(st.transaction_date) >= ? AND DATE(st.transaction_date) <= ?
		GROUP BY e.first_name, e.last_name, b.name
		ORDER BY total_amount DESC
		LIMIT 10
	`, monthStartStr, todayStr).Scan(&sales.TopSalesmenMTD)

	c.JSON(http.StatusOK, sales)
}

// GetLogisticsKnowledge mengembalikan data logistik dan stok.
func (h *InternalReportHandler) GetLogisticsKnowledge(c *gin.Context) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")

	var logis LogisticsKnowledge
	h.db.Raw(`
		SELECT COUNT(*) FROM delivery_batches
		WHERE DATE(delivery_date) = ? AND status = 'COMPLETED'
	`, todayStr).Scan(&logis.CompletedDeliveriesToday)

	h.db.Raw(`
		SELECT COUNT(*) FROM delivery_batches
		WHERE DATE(delivery_date) = ?
	`, todayStr).Scan(&logis.TotalDeliveriesToday)

	h.db.Raw(`
		SELECT 
			COALESCE(db.delivery_order_no, 'BATCH-' || SUBSTRING(db.id::text, 1, 8)) AS batch_order_no,
			CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) AS driver_name,
			COALESCE(v.plate_number, '-') AS vehicle_plate,
			db.status AS status,
			(SELECT COUNT(*) FROM delivery_items di WHERE di.delivery_batch_id = db.id) AS total_items,
			COALESCE(b.name, '-') AS branch_name
		FROM delivery_batches db
		LEFT JOIN employees e ON e.id = db.driver_id
		LEFT JOIN vehicles v ON v.id = db.vehicle_id
		LEFT JOIN branches b ON b.id = db.branch_id
		WHERE db.status IN ('PICKING', 'ON_DELIVERY', 'SUPERVISOR_APPROVED', 'WAITING_APPROVAL')
		ORDER BY db.created_at DESC
	`).Scan(&logis.OngoingDeliveries)

	h.db.Raw(`
		SELECT 
			p.name AS product_name,
			p.sku AS sku,
			COALESCE(SUM(fs.quantity), 0) AS current_qty,
			p.unit AS unit,
			COALESCE(fw.name, 'Gudang Utama') AS warehouse
		FROM products p
		LEFT JOIN factory_stocks fs ON fs.product_id = p.id
		LEFT JOIN factory_warehouses fw ON fw.id = fs.factory_warehouse_id
		GROUP BY p.name, p.sku, p.unit, fw.name
		HAVING COALESCE(SUM(fs.quantity), 0) < 50
		ORDER BY current_qty ASC
		LIMIT 20
	`).Scan(&logis.LowStockAlerts)

	c.JSON(http.StatusOK, logis)
}

// GetReportData adalah handler legacy untuk kompatibilitas wa-report-service.
func (h *InternalReportHandler) GetReportData(c *gin.Context) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")

	data := ReportDataResponse{
		ReportPeriod: now.Format("Monday, 02 January 2006"),
	}

	h.db.Raw(`
		SELECT COUNT(DISTINCT al.id) FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		WHERE DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		  AND UPPER(al.status) IN ('PRESENT', 'ON_TIME')
		  AND e.employment_status != 'resigned'
	`, todayStr).Scan(&data.PresentToday)

	h.db.Raw(`
		SELECT COUNT(DISTINCT al.id) FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		WHERE DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		  AND UPPER(al.status) = 'LATE'
		  AND e.employment_status != 'resigned'
	`, todayStr).Scan(&data.LateToday)

	h.db.Raw(`
		SELECT COUNT(*) FROM leave_requests
		WHERE status = 'APPROVED'
		  AND ? BETWEEN DATE(start_date) AND DATE(end_date)
	`, todayStr).Scan(&data.OnLeaveToday)

	data.TotalEmployees = data.PresentToday + data.LateToday + data.OnLeaveToday
	data.AbsentToday = 0

	h.db.Raw(`SELECT COUNT(*) FROM leave_requests WHERE status = 'PENDING'`).Scan(&data.PendingLeaveReqs)
	h.db.Raw(`SELECT COUNT(*) FROM overtimes WHERE status = 'pending'`).Scan(&data.PendingOvertimeReqs)
	h.db.Raw(`SELECT COUNT(*) FROM reimbursements WHERE status = 'PENDING'`).Scan(&data.PendingReimbursement)

	h.db.Raw(`
		SELECT COALESCE(SUM(total_amount), 0) FROM sales_transactions
		WHERE DATE(transaction_date) = ?
	`, todayStr).Scan(&data.TotalSalesToday)

	h.db.Raw(`
		SELECT COUNT(*) FROM sales_transactions
		WHERE DATE(transaction_date) = ?
	`, todayStr).Scan(&data.TotalTransactions)

	row := h.db.Raw(`
		SELECT
			CONCAT(e.first_name, ' ', e.last_name) AS full_name,
			COALESCE(SUM(st.total_amount), 0)       AS total
		FROM sales_transactions st
		JOIN employees e ON e.id = st.employee_id
		WHERE DATE(st.transaction_date) = ?
		GROUP BY e.first_name, e.last_name
		ORDER BY total DESC
		LIMIT 1
	`, todayStr).Row()
	if row != nil {
		_ = row.Scan(&data.TopSalesmanName, &data.TopSalesmanAmount)
	}

	h.db.Raw(`
		SELECT COUNT(*) FROM delivery_batches
		WHERE DATE(delivery_date) = ?
		  AND status = 'COMPLETED'
	`, todayStr).Scan(&data.CompletedDeliveries)

	h.db.Raw(`
		SELECT COUNT(*) FROM delivery_batches
		WHERE DATE(delivery_date) = ?
	`, todayStr).Scan(&data.TotalDeliveries)

	h.db.Raw(`
		SELECT COUNT(*) FROM sales_orders
		WHERE status IN ('DRAFT', 'CONFIRMED')
	`, todayStr).Scan(&data.PendingOrders)

	h.db.Raw(`
		SELECT 
			b.name AS branch_name,
			COUNT(DISTINCT CASE WHEN UPPER(al.status) IN ('PRESENT', 'ON_TIME', 'LATE') THEN al.id END) AS present_today,
			COUNT(DISTINCT CASE WHEN lr.status = 'APPROVED' AND ? BETWEEN DATE(lr.start_date) AND DATE(lr.end_date) THEN lr.id END) AS on_leave_today
		FROM branches b
		JOIN employees e ON e.branch_id = b.id
		LEFT JOIN attendance_logs al ON al.employee_id = e.id AND DATE(al.clock_in_time AT TIME ZONE 'Asia/Jakarta') = ?
		LEFT JOIN leave_requests lr ON lr.employee_id = e.id AND lr.status = 'APPROVED' AND ? BETWEEN DATE(lr.start_date) AND DATE(lr.end_date)
		WHERE e.employment_status != 'resigned'
		GROUP BY b.name
	`, todayStr, todayStr, todayStr).Scan(&data.BranchData)

	c.JSON(http.StatusOK, data)
}
