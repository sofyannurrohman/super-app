package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"example.com/hris/backend/pkg/ai"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIOperationalHandler struct {
	db *gorm.DB
}

func NewAIOperationalHandler(db *gorm.DB) *AIOperationalHandler {
	return &AIOperationalHandler{db: db}
}

type aiCustomerResult struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	City    string    `json:"city,omitempty"`
	Address string    `json:"address,omitempty"`
}

type aiCustomerSalesSummary struct {
	CustomerID       uuid.UUID `json:"customer_id"`
	CustomerName     string    `json:"customer_name"`
	StartDate        string    `json:"start_date"`
	EndDate          string    `json:"end_date"`
	TotalRevenue     float64   `json:"total_revenue"`
	TransactionCount int64     `json:"transaction_count"`
	LastTransaction  *string   `json:"last_transaction_date,omitempty"`
}

type aiSalesBranchRow struct {
	BranchName       string  `json:"branch_name"`
	TotalRevenue     float64 `json:"total_revenue"`
	TransactionCount int64   `json:"transaction_count"`
}

type aiSalesChannelRow struct {
	Channel          string  `json:"channel"`
	TotalRevenue     float64 `json:"total_revenue"`
	TransactionCount int64   `json:"transaction_count"`
}

type aiSalesTopProductRow struct {
	ProductName  string  `json:"product_name"`
	SKU          string  `json:"sku"`
	TotalQtySold int64   `json:"total_qty_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

type aiSalesSummary struct {
	StartDate          string                 `json:"start_date"`
	EndDate            string                 `json:"end_date"`
	FilterBranch       string                 `json:"filter_branch,omitempty"`
	FilterSalesType    string                 `json:"filter_sales_type,omitempty"`
	TotalRevenue       float64                `json:"total_revenue"`
	TransactionCount   int64                  `json:"transaction_count"`
	BranchBreakdown    []aiSalesBranchRow     `json:"branch_breakdown,omitempty"`
	ChannelBreakdown   []aiSalesChannelRow    `json:"channel_breakdown,omitempty"`
	TopSalesmen        []aiTopSalesmanRow     `json:"top_salesmen,omitempty"`
	TopSellingProducts []aiSalesTopProductRow `json:"top_selling_products,omitempty"`
}

type aiTopSalesmanRow struct {
	SalesmanName string  `json:"salesman_name"`
	TotalAmount  float64 `json:"total_amount"`
	Transactions int64   `json:"transactions"`
}

type aiInventoryStockRow struct {
	ProductName   string  `json:"product_name"`
	SKU           string  `json:"sku"`
	WarehouseName string  `json:"warehouse_name"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	IsLowStock    bool    `json:"is_low_stock"`
}

type aiPresentEmpRow struct {
	EmployeeName string `json:"employee_name"`
	ClockInTime  string `json:"clock_in_time"`
	ClockOutTime string `json:"clock_out_time,omitempty"`
	Status       string `json:"status"`
	Department   string `json:"department"`
	Branch       string `json:"branch"`
}

type aiAttendanceSummary struct {
	Date             string             `json:"date"`
	TotalPresent     int64              `json:"total_present"`
	TotalLate        int64              `json:"total_late"`
	TotalOnLeave     int64              `json:"total_on_leave"`
	PresentEmployees []aiPresentEmpRow  `json:"present_employees,omitempty"`
	LateEmployees    []aiLateEmpRow     `json:"late_employees,omitempty"`
	ActiveLeaves     []aiActiveLeaveRow `json:"active_leaves,omitempty"`
}

type aiLateEmpRow struct {
	EmployeeName string `json:"employee_name"`
	ClockInTime  string `json:"clock_in_time"`
	Department   string `json:"department"`
}

type aiActiveLeaveRow struct {
	EmployeeName string `json:"employee_name"`
	LeaveType    string `json:"leave_type"`
	Reason       string `json:"reason"`
}

type aiEmployeeSearchRow struct {
	EmployeeIDNumber string `json:"employee_id_number"`
	FullName         string `json:"full_name"`
	JobPosition      string `json:"job_position"`
	Department       string `json:"department"`
	Branch           string `json:"branch"`
	PhoneNumber      string `json:"phone_number,omitempty"`
	Gender           string `json:"gender,omitempty"`
	Status           string `json:"status"`
}

type AIWidget struct {
	Type      string         `json:"type"` // "chart", "approval_list", "attendance_stats"
	Title     string         `json:"title,omitempty"`
	ChartType string         `json:"chart_type,omitempty"` // "bar", "pie", "line"
	Data      map[string]any `json:"data,omitempty"`
	Summary   map[string]any `json:"summary,omitempty"`
	Items     any            `json:"items,omitempty"`
}

type aiPendingItemDetailRow struct {
	Type         string  `json:"type"` // "Reimbursement (Approval)", "Reimbursement (Pencairan Kasir)", "Cuti", "Lembur", "Memo"
	EmployeeName string  `json:"employee_name"`
	BranchName   string  `json:"branch_name"`
	Title        string  `json:"title"`
	Amount       float64 `json:"amount,omitempty"`
	Date         string  `json:"date"`
	Status       string  `json:"status"`
	RoutePath    string  `json:"route_path,omitempty"`
}

type aiPendingApprovalsSummary struct {
	PendingLeaves                 int64                    `json:"pending_leaves"`
	PendingOvertimes              int64                    `json:"pending_overtimes"`
	PendingReimbursementsApproval int64                    `json:"pending_reimbursements_approval"`
	TotalReimbursementApprovalRp  float64                  `json:"total_reimbursement_approval_rp"`
	PendingReimbursementsDisburse int64                    `json:"pending_reimbursements_disbursement"`
	TotalReimbursementDisburseRp  float64                  `json:"total_reimbursement_disbursement_rp"`
	PendingMemos                  int64                    `json:"pending_memos"`
	PendingEpkwt                  int64                    `json:"pending_epkwt"`
	RecentPendingItems            []aiPendingItemDetailRow `json:"recent_pending_items,omitempty"`
}

type aiBranchRow struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Address string `json:"address,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

type aiDepartmentRow struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type aiDepartmentsAndBranches struct {
	Branches    []aiBranchRow     `json:"branches"`
	Departments []aiDepartmentRow `json:"departments"`
}

type aiDeliveryOrderRow struct {
	DONumber     string `json:"delivery_order_no"`
	DriverName   string `json:"driver_name"`
	VehiclePlate string `json:"vehicle_plate"`
	BranchName   string `json:"branch_name"`
	Status       string `json:"status"`
	DeliveryDate string `json:"delivery_date,omitempty"`
}

type aiSalesOrderRow struct {
	SONumber     string  `json:"so_number"`
	StoreName    string  `json:"store_name"`
	SalesmanName string  `json:"salesman_name"`
	TotalAmount  float64 `json:"total_amount"`
	Status       string  `json:"status"`
	OrderDate    string  `json:"order_date"`
}

type aiVehicleRow struct {
	Name         string `json:"name"`
	LicensePlate string `json:"license_plate"`
	Model        string `json:"model"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	BranchName   string `json:"branch_name"`
	Mileage      int    `json:"mileage"`
}

type aiRawMaterialStockRow struct {
	Name     string  `json:"name"`
	SKU      string  `json:"sku"`
	Category string  `json:"category"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type aiKasPabrikRow struct {
	KasName     string  `json:"kas_name"`
	BranchName  string  `json:"branch_name"`
	Balance     float64 `json:"balance"`
	AccountCode string  `json:"account_code,omitempty"`
}

type aiKasPabrikSummary struct {
	TotalBalance float64          `json:"total_balance"`
	Accounts     []aiKasPabrikRow `json:"accounts"`
}

type aiMarketplaceBranchRow struct {
	BranchName   string  `json:"branch_name"`
	TotalOrders  int64   `json:"total_orders"`
	TotalRevenue float64 `json:"total_revenue"`
}

type aiMarketplacePlatformRow struct {
	Platform     string  `json:"platform"`
	TotalOrders  int64   `json:"total_orders"`
	TotalRevenue float64 `json:"total_revenue"`
}

type aiMarketplaceTopProductRow struct {
	ProductName  string  `json:"product_name"`
	SellerSKU    string  `json:"seller_sku"`
	TotalQtySold int64   `json:"total_qty_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

type aiMarketplaceSummary struct {
	Period             string                       `json:"period,omitempty"`
	StartDate          string                       `json:"start_date,omitempty"`
	EndDate            string                       `json:"end_date,omitempty"`
	FilterBranch       string                       `json:"filter_branch,omitempty"`
	FilterPlatform     string                       `json:"filter_platform,omitempty"`
	TotalOrders        int64                        `json:"total_orders"`
	TotalRevenue       float64                      `json:"total_revenue"`
	AverageOrderValue  float64                      `json:"average_order_value"`
	BranchBreakdown    []aiMarketplaceBranchRow     `json:"branch_breakdown,omitempty"`
	PlatformBreakdown  []aiMarketplacePlatformRow   `json:"platform_breakdown,omitempty"`
	TopSellingProducts []aiMarketplaceTopProductRow `json:"top_selling_products,omitempty"`
}

func (h *AIOperationalHandler) RegisterRoutes(r *gin.RouterGroup) {
	aiRoutes := r.Group("/ai/tools")
	aiRoutes.POST("/search-customer", h.SearchCustomer)
	aiRoutes.POST("/customer-sales-summary", h.CustomerSalesSummary)
	aiRoutes.POST("/execute", h.Execute)
}

// ExecuteDirect mengeksekusi tool secara programatis untuk keperluan orchestrator AI
func (h *AIOperationalHandler) ExecuteDirect(ctx context.Context, companyID uuid.UUID, tool string, args map[string]any) (any, *AIWidget, error) {
	if args == nil {
		args = make(map[string]any)
	}
	if err := ai.ValidateToolCall(tool, args); err != nil {
		return nil, nil, err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	switch tool {
	case "search_customer":
		query, _ := args["query"].(string)
		like := "%" + strings.TrimSpace(query) + "%"
		var rows []aiCustomerResult
		dbq := h.db.WithContext(ctx).Table("stores s").
			Select("s.id, s.name, COALESCE(NULLIF(s.distribution_city, ''), NULLIF(s.regency, ''), '') AS city, s.address").
			Where("s.is_active = TRUE AND (s.name ILIKE ? OR s.owner_name ILIKE ? OR s.address ILIKE ? OR s.distribution_city ILIKE ? OR s.regency ILIKE ?)", like, like, like, like, like)
		if companyID != uuid.Nil {
			dbq = dbq.Where("s.company_id = ?", companyID)
		}
		if err := dbq.Order("s.name ASC").Limit(10).Scan(&rows).Error; err != nil {
			return nil, nil, fmt.Errorf("customer search gagal: %w", err)
		}
		return rows, nil, nil

	case "get_customer_sales_summary":
		customerIDStr, _ := args["customer_id"].(string)
		customerID, err := uuid.Parse(customerIDStr)
		if err != nil {
			return nil, nil, fmt.Errorf("customer_id tidak valid: %w", err)
		}
		startStr, _ := args["start_date"].(string)
		endStr, _ := args["end_date"].(string)
		periodStr, _ := args["period"].(string)
		start, end, err := parseAIRange(startStr, endStr, periodStr)
		if err != nil {
			return nil, nil, err
		}
		var result aiCustomerSalesSummary
		result.CustomerID = customerID
		result.StartDate = start.Format("2006-01-02")
		result.EndDate = end.Add(-24 * time.Hour).Format("2006-01-02")

		var customer struct{ Name string }
		cq := h.db.WithContext(ctx).Table("stores").Select("name").Where("id = ? AND is_active = TRUE", customerID)
		if companyID != uuid.Nil {
			cq = cq.Where("company_id = ?", companyID)
		}
		if err := cq.Scan(&customer).Error; err != nil || customer.Name == "" {
			return nil, nil, fmt.Errorf("customer tidak ditemukan")
		}
		result.CustomerName = customer.Name

		var aggregate struct {
			TotalRevenue     float64
			TransactionCount int64
			LastTransaction  *string
		}
		query := h.db.WithContext(ctx).Table("sales_transactions st").
			Select("COALESCE(SUM(st.total_amount), 0) AS total_revenue, COUNT(st.id) AS transaction_count, MAX(st.transaction_date)::text AS last_transaction").
			Where("st.store_id = ? AND st.transaction_date >= ? AND st.transaction_date < ? AND st.status NOT IN ('CANCELLED', 'REJECTED')", customerID, start, end)
		if companyID != uuid.Nil {
			query = query.Where("st.company_id = ?", companyID)
		}
		if err := query.Scan(&aggregate).Error; err != nil {
			return nil, nil, fmt.Errorf("sales summary gagal: %w", err)
		}
		result.TotalRevenue = aggregate.TotalRevenue
		result.TransactionCount = aggregate.TransactionCount
		result.LastTransaction = aggregate.LastTransaction
		return result, nil, nil

	case "get_sales_summary":
		startStr, _ := args["start_date"].(string)
		endStr, _ := args["end_date"].(string)
		periodStr, _ := args["period"].(string)
		branchName, _ := args["branch_name"].(string)
		branchIDStr, _ := args["branch_id"].(string)
		salesType, _ := args["sales_type"].(string)

		start, end, err := parseAIRange(startStr, endStr, periodStr)
		if err != nil {
			return nil, nil, err
		}

		var summary aiSalesSummary
		summary.StartDate = start.Format("2006-01-02")
		summary.EndDate = end.Add(-24 * time.Hour).Format("2006-01-02")
		summary.FilterBranch = branchName
		summary.FilterSalesType = salesType
		summary.BranchBreakdown = make([]aiSalesBranchRow, 0)
		summary.ChannelBreakdown = make([]aiSalesChannelRow, 0)
		summary.TopSalesmen = make([]aiTopSalesmanRow, 0)
		summary.TopSellingProducts = make([]aiSalesTopProductRow, 0)

		// Resolusi Branch ID jika branch_name disediakan
		var targetBranchID *uuid.UUID
		if branchIDStr != "" {
			if bID, err := uuid.Parse(branchIDStr); err == nil {
				targetBranchID = &bID
			}
		} else if branchName != "" {
			var b struct{ ID uuid.UUID }
			_ = h.db.WithContext(ctx).Table("branches").Select("id").Where("name ILIKE ?", "%"+strings.TrimSpace(branchName)+"%").Limit(1).Scan(&b).Error
			if b.ID != uuid.Nil {
				targetBranchID = &b.ID
			}
		}

		// 1. Total Penjualan & Jumlah Transaksi
		txQ := h.db.WithContext(ctx).Table("sales_transactions st").
			Select("COALESCE(SUM(st.total_amount), 0) AS total_revenue, COUNT(st.id) AS transaction_count").
			Where("st.transaction_date >= ? AND st.transaction_date < ? AND st.status NOT IN ('CANCELLED', 'REJECTED')", start, end)
		if companyID != uuid.Nil {
			txQ = txQ.Where("st.company_id = ?", companyID)
		}
		if targetBranchID != nil {
			txQ = txQ.Joins("JOIN employees e ON st.employee_id = e.id").Where("e.branch_id = ?", *targetBranchID)
		}
		if salesType != "" && !strings.EqualFold(salesType, "all") {
			if strings.EqualFold(salesType, "GT") || strings.EqualFold(salesType, "canvas") {
				txQ = txQ.Where("(st.channel IN ('CANVAS', 'MOTORIS') OR st.channel IS NULL)")
			} else if strings.EqualFold(salesType, "MT") || strings.EqualFold(salesType, "modern_trade") {
				txQ = txQ.Where("st.channel ILIKE '%modern%' OR st.channel = 'MT'")
			} else {
				txQ = txQ.Where("st.channel ILIKE ?", "%"+strings.TrimSpace(salesType)+"%")
			}
		}

		var agg struct {
			TotalRevenue     float64
			TransactionCount int64
		}
		if err := txQ.Scan(&agg).Error; err != nil {
			return nil, nil, fmt.Errorf("gagal query total sales: %w", err)
		}
		summary.TotalRevenue = agg.TotalRevenue
		summary.TransactionCount = agg.TransactionCount

		// 2. Per-Branch Breakdown (Perbandingan Antar Cabang)
		branchQ := h.db.WithContext(ctx).Table("sales_transactions st").
			Select("COALESCE(b.name, 'Gudang/Pusat') AS branch_name, COALESCE(SUM(st.total_amount), 0) AS total_revenue, COUNT(st.id) AS transaction_count").
			Joins("JOIN employees e ON st.employee_id = e.id").
			Joins("LEFT JOIN branches b ON e.branch_id = b.id").
			Where("st.transaction_date >= ? AND st.transaction_date < ? AND st.status NOT IN ('CANCELLED', 'REJECTED')", start, end)
		if companyID != uuid.Nil {
			branchQ = branchQ.Where("st.company_id = ?", companyID)
		}
		if targetBranchID != nil {
			branchQ = branchQ.Where("e.branch_id = ?", *targetBranchID)
		}
		_ = branchQ.Group("b.name").Order("total_revenue DESC").Limit(10).Scan(&summary.BranchBreakdown).Error

		// 3. Channel Breakdown (Kanvas, Motoris, Modern Trade, Direct)
		channelQ := h.db.WithContext(ctx).Table("sales_transactions st").
			Select("COALESCE(NULLIF(TRIM(st.channel), ''), 'CANVAS / REGULER') AS channel, COALESCE(SUM(st.total_amount), 0) AS total_revenue, COUNT(st.id) AS transaction_count").
			Where("st.transaction_date >= ? AND st.transaction_date < ? AND st.status NOT IN ('CANCELLED', 'REJECTED')", start, end)
		if companyID != uuid.Nil {
			channelQ = channelQ.Where("st.company_id = ?", companyID)
		}
		if targetBranchID != nil {
			channelQ = channelQ.Joins("JOIN employees e ON st.employee_id = e.id").Where("e.branch_id = ?", *targetBranchID)
		}
		_ = channelQ.Group("st.channel").Order("total_revenue DESC").Scan(&summary.ChannelBreakdown).Error

		// 4. Top 5 Salesmen
		topQ := h.db.WithContext(ctx).Table("sales_transactions st").
			Select("COALESCE(e.first_name || ' ' || COALESCE(e.last_name, ''), 'Unknown') AS salesman_name, SUM(st.total_amount) AS total_amount, COUNT(st.id) AS transactions").
			Joins("JOIN employees e ON st.employee_id = e.id").
			Where("st.transaction_date >= ? AND st.transaction_date < ? AND st.status NOT IN ('CANCELLED', 'REJECTED')", start, end)
		if companyID != uuid.Nil {
			topQ = topQ.Where("st.company_id = ?", companyID)
		}
		if targetBranchID != nil {
			topQ = topQ.Where("e.branch_id = ?", *targetBranchID)
		}
		_ = topQ.Group("e.first_name, e.last_name").Order("SUM(st.total_amount) DESC").Limit(5).Scan(&summary.TopSalesmen).Error

		// 5. Top 5 Produk Terlaris
		prodQ := h.db.WithContext(ctx).Table("sales_items si").
			Select("p.name AS product_name, p.sku, COALESCE(SUM(si.quantity), 0) AS total_qty_sold, COALESCE(SUM(si.subtotal), 0) AS total_revenue").
			Joins("JOIN sales_transactions st ON si.sales_transaction_id = st.id").
			Joins("JOIN products p ON si.product_id = p.id").
			Where("st.transaction_date >= ? AND st.transaction_date < ? AND st.status NOT IN ('CANCELLED', 'REJECTED')", start, end)
		if companyID != uuid.Nil {
			prodQ = prodQ.Where("st.company_id = ?", companyID)
		}
		if targetBranchID != nil {
			prodQ = prodQ.Joins("JOIN employees e ON st.employee_id = e.id").Where("e.branch_id = ?", *targetBranchID)
		}
		_ = prodQ.Group("p.name, p.sku").Order("total_qty_sold DESC").Limit(5).Scan(&summary.TopSellingProducts).Error

		var widget *AIWidget
		if len(summary.BranchBreakdown) > 0 {
			var categories []string
			var seriesData []float64
			for _, b := range summary.BranchBreakdown {
				categories = append(categories, b.BranchName)
				seriesData = append(seriesData, b.TotalRevenue)
			}
			widget = &AIWidget{
				Type:      "chart",
				Title:     "Performa Omzet per Cabang",
				ChartType: "bar",
				Data: map[string]any{
					"categories":         categories,
					"series":             []map[string]any{{"name": "Omzet (Rp)", "data": seriesData}},
					"total_revenue":      summary.TotalRevenue,
					"total_transactions": summary.TransactionCount,
				},
			}
		} else if len(summary.ChannelBreakdown) > 0 {
			var categories []string
			var seriesData []float64
			for _, c := range summary.ChannelBreakdown {
				categories = append(categories, c.Channel)
				seriesData = append(seriesData, c.TotalRevenue)
			}
			widget = &AIWidget{
				Type:      "chart",
				Title:     "Performa Omzet per Channel",
				ChartType: "bar",
				Data: map[string]any{
					"categories":         categories,
					"series":             []map[string]any{{"name": "Omzet (Rp)", "data": seriesData}},
					"total_revenue":      summary.TotalRevenue,
					"total_transactions": summary.TransactionCount,
				},
			}
		}

		return summary, widget, nil

	case "get_employee_attendance":
		dateStr, _ := args["date"].(string)
		branchName, _ := args["branch_name"].(string)
		branchIDStr, _ := args["branch_id"].(string)
		if dateStr == "" {
			dateStr = now.Format("2006-01-02")
		}
		targetDate, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			return nil, nil, fmt.Errorf("format date harus YYYY-MM-DD: %w", err)
		}

		var att aiAttendanceSummary
		att.Date = targetDate.Format("2006-01-02")
		att.PresentEmployees = make([]aiPresentEmpRow, 0)
		att.LateEmployees = make([]aiLateEmpRow, 0)
		att.ActiveLeaves = make([]aiActiveLeaveRow, 0)

		startWIB := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc)
		endWIB := startWIB.Add(24 * time.Hour)
		dateKey := targetDate.Format("2006-01-02")

		// 1. Ambil Daftar Karyawan yang Hadir / Clock-In
		presentQ := h.db.WithContext(ctx).Table("attendance_logs al").
			Select(`
				COALESCE(NULLIF(TRIM(COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')), ''), u.email, 'Karyawan') AS employee_name,
				TO_CHAR(COALESCE(al.clock_in_time, al.created_at) AT TIME ZONE 'Asia/Jakarta', 'HH24:MI') AS clock_in_time,
				CASE WHEN al.clock_out_time IS NOT NULL THEN TO_CHAR(al.clock_out_time AT TIME ZONE 'Asia/Jakarta', 'HH24:MI') ELSE '-' END AS clock_out_time,
				COALESCE(al.status, 'ON_TIME') AS status,
				COALESCE(d.name, '-') AS department,
				COALESCE(b.name, '-') AS branch
			`).
			Joins("LEFT JOIN employees e ON (al.employee_id = e.id OR al.employee_id = e.user_id)").
			Joins("LEFT JOIN users u ON (al.employee_id = u.id OR e.user_id = u.id)").
			Joins("LEFT JOIN departments d ON e.department_id = d.id").
			Joins("LEFT JOIN branches b ON e.branch_id = b.id").
			Where("(DATE(COALESCE(al.clock_in_time, al.created_at) AT TIME ZONE 'Asia/Jakarta') = ? OR (al.clock_in_time >= ? AND al.clock_in_time < ?))", dateKey, startWIB.UTC(), endWIB.UTC())

		if branchName != "" {
			presentQ = presentQ.Where("b.name ILIKE ?", "%"+strings.TrimSpace(branchName)+"%")
		}
		if branchIDStr != "" {
			if bID, err := uuid.Parse(branchIDStr); err == nil {
				presentQ = presentQ.Where("e.branch_id = ?", bID)
			}
		}

		// Hitung Total Hadir Aktual
		var countPresent int64
		_ = h.db.WithContext(ctx).Table("attendance_logs al").
			Joins("LEFT JOIN employees e ON (al.employee_id = e.id OR al.employee_id = e.user_id)").
			Where("(DATE(COALESCE(al.clock_in_time, al.created_at) AT TIME ZONE 'Asia/Jakarta') = ? OR (al.clock_in_time >= ? AND al.clock_in_time < ?))", dateKey, startWIB.UTC(), endWIB.UTC()).
			Count(&countPresent).Error

		_ = presentQ.Order("COALESCE(al.clock_in_time, al.created_at) ASC").Limit(30).Scan(&att.PresentEmployees).Error
		att.TotalPresent = countPresent
		if att.TotalPresent < int64(len(att.PresentEmployees)) {
			att.TotalPresent = int64(len(att.PresentEmployees))
		}

		for _, p := range att.PresentEmployees {
			if strings.EqualFold(p.Status, "LATE") {
				att.TotalLate++
				att.LateEmployees = append(att.LateEmployees, aiLateEmpRow{
					EmployeeName: p.EmployeeName,
					ClockInTime:  p.ClockInTime,
					Department:   p.Department,
				})
			}
		}

		// 2. Count & List Active Leaves
		type leaveRow struct {
			EmployeeName string `gorm:"column:employee_name"`
			LeaveType    string `gorm:"column:leave_type"`
			Reason       string `gorm:"column:reason"`
		}
		var leaves []leaveRow
		lq := h.db.WithContext(ctx).Table("leave_requests lr").
			Select(`
				e.first_name || ' ' || COALESCE(e.last_name, '') AS employee_name,
				COALESCE(lt.name, 'Cuti') AS leave_type,
				COALESCE(lr.reason, '-') AS reason
			`).
			Joins("JOIN employees e ON lr.employee_id = e.id").
			Joins("LEFT JOIN leave_types lt ON lr.leave_type_id = lt.id").
			Joins("LEFT JOIN branches b ON e.branch_id = b.id").
			Where("lr.start_date <= ? AND lr.end_date >= ? AND UPPER(TRIM(lr.status)) = 'APPROVED' AND (e.deleted_at IS NULL OR e.id IS NULL)", targetDate, targetDate)
		if companyID != uuid.Nil {
			lq = lq.Where("(e.company_id = ? OR e.company_id IS NULL)", companyID)
		}
		if branchName != "" {
			lq = lq.Where("b.name ILIKE ?", "%"+strings.TrimSpace(branchName)+"%")
		}
		if branchIDStr != "" {
			if bID, err := uuid.Parse(branchIDStr); err == nil {
				lq = lq.Where("e.branch_id = ?", bID)
			}
		}
		_ = lq.Limit(15).Scan(&leaves).Error
		att.TotalOnLeave = int64(len(leaves))
		for _, l := range leaves {
			att.ActiveLeaves = append(att.ActiveLeaves, aiActiveLeaveRow{
				EmployeeName: l.EmployeeName,
				LeaveType:    l.LeaveType,
				Reason:       l.Reason,
			})
		}

		widget := &AIWidget{
			Type:      "attendance_stats",
			Title:     "Statistik Presensi Hari Ini",
			ChartType: "pie",
			Data: map[string]any{
				"total_present":     att.TotalPresent,
				"total_late":        att.TotalLate,
				"total_on_leave":    att.TotalOnLeave,
				"present_employees": att.PresentEmployees,
			},
		}

		return att, widget, nil

	case "get_inventory_stock":
		productQuery, _ := args["product_query"].(string)
		warehouseName, _ := args["warehouse_name"].(string)
		branchIDStr, _ := args["branch_id"].(string)

		var rows []aiInventoryStockRow
		sq := h.db.WithContext(ctx).Table("warehouse_stocks ws").
			Select("p.name AS product_name, p.sku, COALESCE(b.name, 'Gudang Utama') AS warehouse_name, SUM(ws.quantity) AS quantity, COALESCE(p.unit, 'PCS') AS unit, (SUM(ws.quantity) <= MAX(COALESCE(ws.min_limit, 10))) AS is_low_stock").
			Joins("JOIN products p ON ws.product_id = p.id").
			Joins("JOIN branches b ON ws.branch_id = b.id").
			Where("ws.stock_type = 'GOOD'")
		if companyID != uuid.Nil {
			sq = sq.Where("b.company_id = ?", companyID)
		}
		if productQuery != "" {
			like := "%" + strings.TrimSpace(productQuery) + "%"
			sq = sq.Where("p.name ILIKE ? OR p.sku ILIKE ?", like, like)
		}
		if warehouseName != "" {
			likeW := "%" + strings.TrimSpace(warehouseName) + "%"
			sq = sq.Where("b.name ILIKE ? OR b.address ILIKE ?", likeW, likeW)
		}
		if branchIDStr != "" {
			if bID, err := uuid.Parse(branchIDStr); err == nil {
				sq = sq.Where("ws.branch_id = ?", bID)
			}
		}
		sq = sq.Group("p.name, p.sku, b.name, p.unit")
		if err := sq.Order("quantity ASC").Limit(30).Scan(&rows).Error; err != nil {
			return nil, nil, fmt.Errorf("query inventory stock gagal: %w", err)
		}

		// Match the branch name exactly when the marketplace does not provide a branch ID.
		if len(rows) == 0 && warehouseName != "" {
			trimmedBranch := strings.TrimSpace(warehouseName)
			trimmedBranch = strings.ReplaceAll(trimmedBranch, "Cabang", "")
			trimmedBranch = strings.ReplaceAll(trimmedBranch, "cabang", "")
			trimmedBranch = strings.ReplaceAll(trimmedBranch, "Gudang", "")
			trimmedBranch = strings.ReplaceAll(trimmedBranch, "gudang", "")
			trimmedBranch = strings.TrimSpace(trimmedBranch)
			if len(trimmedBranch) >= 2 {
				retryQ := h.db.WithContext(ctx).Table("warehouse_stocks ws").
					Select("p.name AS product_name, p.sku, COALESCE(b.name, 'Gudang Utama') AS warehouse_name, SUM(ws.quantity) AS quantity, COALESCE(p.unit, 'PCS') AS unit, (SUM(ws.quantity) <= MAX(COALESCE(ws.min_limit, 10))) AS is_low_stock").
					Joins("JOIN products p ON ws.product_id = p.id").
					Joins("JOIN branches b ON ws.branch_id = b.id").
					Where("ws.stock_type = 'GOOD' AND (b.name ILIKE ? OR b.address ILIKE ?)", "%"+trimmedBranch+"%", "%"+trimmedBranch+"%")
				if companyID != uuid.Nil {
					retryQ = retryQ.Where("b.company_id = ?", companyID)
				}
				if productQuery != "" {
					like := "%" + strings.TrimSpace(productQuery) + "%"
					retryQ = retryQ.Where("p.name ILIKE ? OR p.sku ILIKE ?", like, like)
				}
				retryQ = retryQ.Group("p.name, p.sku, b.name, p.unit")
				_ = retryQ.Order("quantity ASC").Limit(30).Scan(&rows).Error
			}
		}

		return rows, nil, nil

	case "search_employee":
		keyword, _ := args["keyword"].(string)
		cleanKeyword := cleanSearchKeyword(keyword)
		like := "%" + cleanKeyword + "%"
		var rows []aiEmployeeSearchRow
		eq := h.db.WithContext(ctx).Table("employees e").
			Select(`
				COALESCE(e.employee_id_number, '-') AS employee_id_number,
				COALESCE(NULLIF(TRIM(COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')), ''), u.email, 'Karyawan') AS full_name,
				COALESCE(jp.name, '-') AS job_position,
				COALESCE(d.name, '-') AS department,
				COALESCE(b.name, '-') AS branch,
				COALESCE(e.phone_number, '-') AS phone_number,
				COALESCE(e.gender, '-') AS gender,
				COALESCE(e.employment_status, 'ACTIVE') AS status
			`).
			Joins("LEFT JOIN users u ON e.user_id = u.id").
			Joins("LEFT JOIN job_positions jp ON e.job_position_id = jp.id").
			Joins("LEFT JOIN departments d ON e.department_id = d.id").
			Joins("LEFT JOIN branches b ON e.branch_id = b.id").
			Where("(e.first_name ILIKE ? OR e.last_name ILIKE ? OR e.employee_id_number ILIKE ? OR CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) ILIKE ? OR u.email ILIKE ? OR e.phone_number ILIKE ?)", like, like, like, like, like, like)
		if companyID != uuid.Nil {
			eq = eq.Where("(e.company_id = ? OR e.company_id IS NULL)", companyID)
		}
		if err := eq.Order("e.first_name ASC").Limit(15).Scan(&rows).Error; err != nil {
			return nil, nil, fmt.Errorf("pencarian karyawan gagal: %w", err)
		}

		tokens := strings.Fields(cleanKeyword)
		// Fallback token match if multi-word keyword yielded 0 results
		if len(rows) == 0 && len(tokens) > 1 {
			tokenQ := h.db.WithContext(ctx).Table("employees e").
				Select(`
					COALESCE(e.employee_id_number, '-') AS employee_id_number,
					COALESCE(NULLIF(TRIM(COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')), ''), u.email, 'Karyawan') AS full_name,
					COALESCE(jp.name, '-') AS job_position,
					COALESCE(d.name, '-') AS department,
					COALESCE(b.name, '-') AS branch,
					COALESCE(e.phone_number, '-') AS phone_number,
					COALESCE(e.gender, '-') AS gender,
					COALESCE(e.employment_status, 'ACTIVE') AS status
				`).
				Joins("LEFT JOIN users u ON e.user_id = u.id").
				Joins("LEFT JOIN job_positions jp ON e.job_position_id = jp.id").
				Joins("LEFT JOIN departments d ON e.department_id = d.id").
				Joins("LEFT JOIN branches b ON e.branch_id = b.id")
			for _, tok := range tokens {
				tLike := "%" + tok + "%"
				tokenQ = tokenQ.Or("(e.first_name ILIKE ? OR e.last_name ILIKE ? OR CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')) ILIKE ? OR u.email ILIKE ?)", tLike, tLike, tLike, tLike)
			}
			if companyID != uuid.Nil {
				tokenQ = tokenQ.Where("(e.company_id = ? OR e.company_id IS NULL)", companyID)
			}
			_ = tokenQ.Order("e.first_name ASC").Limit(15).Scan(&rows).Error
		}

		// Fallback direct check to users table (e.g. superadmin / admin / management accounts without employee profile)
		if len(rows) == 0 {
			var uRows []struct {
				Email    string `gorm:"column:email"`
				Role     string `gorm:"column:role"`
				IsActive bool   `gorm:"column:is_active"`
			}
			uq := h.db.WithContext(ctx).Table("users u").
				Select("u.email, u.role, u.is_active").
				Where("(u.email ILIKE ? OR u.role ILIKE ?)", like, like)
			if len(tokens) > 1 {
				for _, tok := range tokens {
					tLike := "%" + tok + "%"
					uq = uq.Or("(u.email ILIKE ? OR u.role ILIKE ?)", tLike, tLike)
				}
			}
			if companyID != uuid.Nil {
				uq = uq.Where("(u.company_id = ? OR u.company_id IS NULL)", companyID)
			}
			_ = uq.Limit(10).Scan(&uRows).Error
			for _, u := range uRows {
				statusStr := "ACTIVE"
				if !u.IsActive {
					statusStr = "INACTIVE"
				}
				emailPrefix := strings.Split(u.Email, "@")[0]
				displayName := strings.Title(strings.ReplaceAll(emailPrefix, ".", " "))
				rows = append(rows, aiEmployeeSearchRow{
					EmployeeIDNumber: "-",
					FullName:         fmt.Sprintf("%s (%s)", displayName, u.Email),
					JobPosition:      strings.ToUpper(strings.ReplaceAll(u.Role, "_", " ")),
					Department:       "Akun Sistem / Manajemen",
					Branch:           "-",
					PhoneNumber:      "-",
					Gender:           "-",
					Status:           statusStr,
				})
			}
		}

		return rows, nil, nil

	case "get_pending_approvals":
		var summary aiPendingApprovalsSummary
		summary.RecentPendingItems = make([]aiPendingItemDetailRow, 0)

		// 1. Cuti (Leaves)
		_ = h.db.WithContext(ctx).Table("leave_requests lr").
			Joins("LEFT JOIN employees e ON (lr.employee_id = e.id OR lr.employee_id = e.user_id)").
			Where("(UPPER(TRIM(lr.status)) IN ('PENDING', 'SUBMITTED', 'WAITING_APPROVAL', 'WAITING') OR lr.status ILIKE '%pending%') AND (e.deleted_at IS NULL OR e.id IS NULL)").
			Count(&summary.PendingLeaves).Error

		// 2. Lembur (Overtimes)
		_ = h.db.WithContext(ctx).Table("overtimes o").
			Joins("LEFT JOIN employees e ON (o.employee_id = e.id OR o.employee_id = e.user_id)").
			Where("(UPPER(TRIM(o.status)) IN ('PENDING', 'SUBMITTED', 'WAITING_APPROVAL', 'WAITING') OR o.status ILIKE '%pending%') AND (e.deleted_at IS NULL OR e.id IS NULL)").
			Count(&summary.PendingOvertimes).Error

		// 3. Reimbursement Menunggu Persetujuan (Approval)
		var reimbApprAgg struct {
			Count int64
			Total float64
		}
		_ = h.db.WithContext(ctx).Table("reimbursements r").
			Select("COUNT(r.id) AS count, COALESCE(SUM(r.amount), 0) AS total").
			Joins("LEFT JOIN employees e ON (r.employee_id = e.id OR r.employee_id = e.user_id)").
			Where("(UPPER(TRIM(r.status)) IN ('PENDING', 'SUBMITTED', 'WAITING_APPROVAL', 'WAITING') OR r.status ILIKE '%pending%') AND (e.deleted_at IS NULL OR e.id IS NULL)").
			Scan(&reimbApprAgg).Error
		summary.PendingReimbursementsApproval = reimbApprAgg.Count
		summary.TotalReimbursementApprovalRp = reimbApprAgg.Total

		// 4. Reimbursement Menunggu Pencairan Kasir (Disbursement)
		var reimbDisbAgg struct {
			Count int64
			Total float64
		}
		_ = h.db.WithContext(ctx).Table("reimbursements r").
			Select("COUNT(r.id) AS count, COALESCE(SUM(r.amount), 0) AS total").
			Joins("LEFT JOIN employees e ON (r.employee_id = e.id OR r.employee_id = e.user_id)").
			Where("UPPER(TRIM(r.status)) = 'APPROVED' AND (UPPER(TRIM(r.disbursement_status)) IN ('UNPAID', 'PENDING', 'WAITING', '') OR r.disbursement_status IS NULL) AND (e.deleted_at IS NULL OR e.id IS NULL)").
			Scan(&reimbDisbAgg).Error
		summary.PendingReimbursementsDisburse = reimbDisbAgg.Count
		summary.TotalReimbursementDisburseRp = reimbDisbAgg.Total

		// 5. Memo Requests
		_ = h.db.WithContext(ctx).Table("memo_requests mr").
			Joins("LEFT JOIN employees e ON (mr.employee_id = e.id OR mr.employee_id = e.user_id)").
			Where("(UPPER(TRIM(mr.status)) IN ('PENDING', 'SUBMITTED', 'WAITING_APPROVAL', 'WAITING') OR mr.status ILIKE '%pending%') AND (e.deleted_at IS NULL OR e.id IS NULL)").
			Count(&summary.PendingMemos).Error

		// 6. EPKWT Requests
		_ = h.db.WithContext(ctx).Table("epkwt_requests er").
			Where("(UPPER(TRIM(er.status)) IN ('PENDING', 'PENDING_HR', 'PENDING_EMPLOYEE', 'PENDING_DIRECTOR', 'DRAFT') OR er.status ILIKE '%pending%')").
			Count(&summary.PendingEpkwt).Error

		// 7. Ambil Detail Pengajuan Pending Terbaru
		type itemRow struct {
			EmployeeName string    `gorm:"column:employee_name"`
			BranchName   string    `gorm:"column:branch_name"`
			Title        string    `gorm:"column:title"`
			Amount       float64   `gorm:"column:amount"`
			CreatedAt    time.Time `gorm:"column:created_at"`
		}

		// Detail Reimbursement Menunggu Persetujuan
		var rApprItems []itemRow
		_ = h.db.WithContext(ctx).Table("reimbursements r").
			Select("COALESCE(NULLIF(TRIM(COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')), ''), u.email, 'Karyawan') AS employee_name, COALESCE(b.name, '-') AS branch_name, r.title, r.amount, r.created_at").
			Joins("LEFT JOIN employees e ON (r.employee_id = e.id OR r.employee_id = e.user_id)").
			Joins("LEFT JOIN users u ON (r.employee_id = u.id OR e.user_id = u.id)").
			Joins("LEFT JOIN branches b ON e.branch_id = b.id").
			Where("(UPPER(TRIM(r.status)) IN ('PENDING', 'SUBMITTED', 'WAITING_APPROVAL', 'WAITING') OR r.status ILIKE '%pending%') AND (e.deleted_at IS NULL OR e.id IS NULL)").
			Order("r.created_at DESC").Limit(5).Scan(&rApprItems).Error
		for _, item := range rApprItems {
			summary.RecentPendingItems = append(summary.RecentPendingItems, aiPendingItemDetailRow{
				Type:         "Reimbursement (Approval)",
				EmployeeName: item.EmployeeName,
				BranchName:   item.BranchName,
				Title:        item.Title,
				Amount:       item.Amount,
				Date:         item.CreatedAt.In(loc).Format("2006-01-02"),
				Status:       "Menunggu Approval",
				RoutePath:    "/reimbursements",
			})
		}

		// Detail Reimbursement Menunggu Pencairan Kasir
		var rDisbItems []itemRow
		_ = h.db.WithContext(ctx).Table("reimbursements r").
			Select("COALESCE(NULLIF(TRIM(COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')), ''), u.email, 'Karyawan') AS employee_name, COALESCE(b.name, '-') AS branch_name, r.title, r.amount, r.created_at").
			Joins("LEFT JOIN employees e ON (r.employee_id = e.id OR r.employee_id = e.user_id)").
			Joins("LEFT JOIN users u ON (r.employee_id = u.id OR e.user_id = u.id)").
			Joins("LEFT JOIN branches b ON e.branch_id = b.id").
			Where("UPPER(TRIM(r.status)) = 'APPROVED' AND (UPPER(TRIM(r.disbursement_status)) IN ('UNPAID', 'PENDING', 'WAITING', '') OR r.disbursement_status IS NULL) AND (e.deleted_at IS NULL OR e.id IS NULL)").
			Order("r.created_at DESC").Limit(5).Scan(&rDisbItems).Error
		for _, item := range rDisbItems {
			summary.RecentPendingItems = append(summary.RecentPendingItems, aiPendingItemDetailRow{
				Type:         "Reimbursement (Pencairan Kasir)",
				EmployeeName: item.EmployeeName,
				BranchName:   item.BranchName,
				Title:        item.Title,
				Amount:       item.Amount,
				Date:         item.CreatedAt.In(loc).Format("2006-01-02"),
				Status:       "Menunggu Pencairan",
				RoutePath:    "/finance/approval",
			})
		}

		// Detail Cuti Pending
		var lItems []itemRow
		_ = h.db.WithContext(ctx).Table("leave_requests lr").
			Select("COALESCE(NULLIF(TRIM(COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')), ''), u.email, 'Karyawan') AS employee_name, COALESCE(b.name, '-') AS branch_name, COALESCE(lr.reason, 'Pengajuan Cuti') AS title, 0 AS amount, lr.created_at").
			Joins("LEFT JOIN employees e ON (lr.employee_id = e.id OR lr.employee_id = e.user_id)").
			Joins("LEFT JOIN users u ON (lr.employee_id = u.id OR e.user_id = u.id)").
			Joins("LEFT JOIN branches b ON e.branch_id = b.id").
			Where("(UPPER(TRIM(lr.status)) IN ('PENDING', 'SUBMITTED', 'WAITING_APPROVAL', 'WAITING') OR lr.status ILIKE '%pending%') AND (e.deleted_at IS NULL OR e.id IS NULL)").
			Order("lr.created_at DESC").Limit(5).Scan(&lItems).Error
		for _, item := range lItems {
			summary.RecentPendingItems = append(summary.RecentPendingItems, aiPendingItemDetailRow{
				Type:         "Cuti",
				EmployeeName: item.EmployeeName,
				BranchName:   item.BranchName,
				Title:        item.Title,
				Date:         item.CreatedAt.In(loc).Format("2006-01-02"),
				Status:       "Menunggu Approval",
				RoutePath:    "/leaves",
			})
		}

		// Detail Lembur Pending
		var oItems []itemRow
		_ = h.db.WithContext(ctx).Table("overtimes o").
			Select("COALESCE(NULLIF(TRIM(COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')), ''), u.email, 'Karyawan') AS employee_name, COALESCE(b.name, '-') AS branch_name, COALESCE(o.reason, 'Pengajuan Lembur') AS title, 0 AS amount, o.created_at").
			Joins("LEFT JOIN employees e ON (o.employee_id = e.id OR o.employee_id = e.user_id)").
			Joins("LEFT JOIN users u ON (o.employee_id = u.id OR e.user_id = u.id)").
			Joins("LEFT JOIN branches b ON e.branch_id = b.id").
			Where("(UPPER(TRIM(o.status)) IN ('PENDING', 'SUBMITTED', 'WAITING_APPROVAL', 'WAITING') OR o.status ILIKE '%pending%') AND (e.deleted_at IS NULL OR e.id IS NULL)").
			Order("o.created_at DESC").Limit(5).Scan(&oItems).Error
		for _, item := range oItems {
			summary.RecentPendingItems = append(summary.RecentPendingItems, aiPendingItemDetailRow{
				Type:         "Lembur",
				EmployeeName: item.EmployeeName,
				BranchName:   item.BranchName,
				Title:        item.Title,
				Date:         item.CreatedAt.In(loc).Format("2006-01-02"),
				Status:       "Menunggu Approval",
				RoutePath:    "/overtimes",
			})
		}

		widget := &AIWidget{
			Type:  "approval_list",
			Title: "Antrean Approval Pending",
			Summary: map[string]any{
				"pending_leaves":                  summary.PendingLeaves,
				"pending_overtimes":               summary.PendingOvertimes,
				"pending_reimbursements_approval": summary.PendingReimbursementsApproval,
				"total_reimbursement_approval_rp": summary.TotalReimbursementApprovalRp,
				"pending_reimbursements_disburse": summary.PendingReimbursementsDisburse,
				"total_reimbursement_disburse_rp": summary.TotalReimbursementDisburseRp,
				"pending_memos":                   summary.PendingMemos,
				"pending_epkwt":                   summary.PendingEpkwt,
			},
			Items: summary.RecentPendingItems,
		}

		return summary, widget, nil

	case "get_departments_and_branches":
		var result aiDepartmentsAndBranches
		result.Branches = make([]aiBranchRow, 0)
		result.Departments = make([]aiDepartmentRow, 0)

		bq := h.db.WithContext(ctx).Table("branches b").
			Select("b.name, COALESCE(b.code, '-') AS code, COALESCE(b.address, '-') AS address, COALESCE(b.phone, '-') AS phone").
			Where("b.is_active = TRUE")
		if companyID != uuid.Nil {
			bq = bq.Where("(b.company_id = ? OR b.company_id IS NULL)", companyID)
		}
		_ = bq.Order("b.name ASC").Scan(&result.Branches).Error

		dq := h.db.WithContext(ctx).Table("departments d").
			Select("d.name, COALESCE(d.code, '-') AS code").
			Where("d.deleted_at IS NULL")
		_ = dq.Order("d.name ASC").Scan(&result.Departments).Error

		return result, nil, nil

	case "get_delivery_orders":
		statusFilter, _ := args["status"].(string)
		dateFilter, _ := args["date"].(string)
		searchFilter, _ := args["search"].(string)

		var rows []aiDeliveryOrderRow
		doQ := h.db.WithContext(ctx).Table("delivery_batches db").
			Select(`
				COALESCE(db.delivery_order_no, 'DO-DRAFT') AS delivery_order_no,
				COALESCE(NULLIF(TRIM(COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')), ''), 'Driver Belum Ditentukan') AS driver_name,
				COALESCE(v.license_plate, '-') AS vehicle_plate,
				COALESCE(b.name, '-') AS branch_name,
				db.status,
				TO_CHAR(COALESCE(db.delivery_date, db.created_at) AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS delivery_date
			`).
			Joins("LEFT JOIN employees e ON db.driver_id = e.id").
			Joins("LEFT JOIN vehicles v ON db.vehicle_id = v.id").
			Joins("LEFT JOIN branches b ON db.branch_id = b.id")

		if companyID != uuid.Nil {
			doQ = doQ.Where("(db.company_id = ? OR db.company_id IS NULL)", companyID)
		}
		if statusFilter != "" {
			doQ = doQ.Where("db.status ILIKE ?", "%"+strings.TrimSpace(statusFilter)+"%")
		}
		if dateFilter != "" {
			doQ = doQ.Where("DATE(COALESCE(db.delivery_date, db.created_at) AT TIME ZONE 'Asia/Jakarta') = ?", dateFilter)
		}
		if searchFilter != "" {
			like := "%" + strings.TrimSpace(searchFilter) + "%"
			doQ = doQ.Where("(db.delivery_order_no ILIKE ? OR e.first_name ILIKE ? OR v.license_plate ILIKE ?)", like, like, like)
		}
		if err := doQ.Order("db.created_at DESC").Limit(20).Scan(&rows).Error; err != nil {
			return nil, nil, fmt.Errorf("gagal query surat jalan: %w", err)
		}
		return rows, nil, nil

	case "get_sales_orders":
		statusFilter, _ := args["status"].(string)
		searchFilter, _ := args["search"].(string)

		var rows []aiSalesOrderRow
		soQ := h.db.WithContext(ctx).Table("sales_orders so").
			Select(`
				so.so_number,
				COALESCE(s.name, 'Pelanggan Umum') AS store_name,
				COALESCE(NULLIF(TRIM(COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')), ''), 'Salesman') AS salesman_name,
				so.total_amount,
				so.status,
				TO_CHAR(so.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS order_date
			`).
			Joins("LEFT JOIN stores s ON so.store_id = s.id").
			Joins("LEFT JOIN employees e ON so.employee_id = e.id")

		if companyID != uuid.Nil {
			soQ = soQ.Where("(so.company_id = ? OR so.company_id IS NULL)", companyID)
		}
		if statusFilter != "" {
			soQ = soQ.Where("so.status ILIKE ?", "%"+strings.TrimSpace(statusFilter)+"%")
		}
		if searchFilter != "" {
			like := "%" + strings.TrimSpace(searchFilter) + "%"
			soQ = soQ.Where("(so.so_number ILIKE ? OR s.name ILIKE ? OR e.first_name ILIKE ?)", like, like, like)
		}
		if err := soQ.Order("so.created_at DESC").Limit(20).Scan(&rows).Error; err != nil {
			return nil, nil, fmt.Errorf("gagal query sales order: %w", err)
		}
		return rows, nil, nil

	case "get_vehicles_status":
		searchFilter, _ := args["search"].(string)

		var rows []aiVehicleRow
		vq := h.db.WithContext(ctx).Table("vehicles v").
			Select("v.name, v.license_plate, COALESCE(v.model, '-') AS model, COALESCE(v.type, 'TRUCK') AS type, v.status, COALESCE(b.name, '-') AS branch_name, v.mileage").
			Joins("LEFT JOIN branches b ON v.branch_id = b.id")

		if companyID != uuid.Nil {
			vq = vq.Where("(v.company_id = ? OR v.company_id IS NULL)", companyID)
		}
		if searchFilter != "" {
			like := "%" + strings.TrimSpace(searchFilter) + "%"
			vq = vq.Where("(v.name ILIKE ? OR v.license_plate ILIKE ? OR v.model ILIKE ?)", like, like, like)
		}
		if err := vq.Order("v.license_plate ASC").Limit(30).Scan(&rows).Error; err != nil {
			return nil, nil, fmt.Errorf("gagal query status kendaraan: %w", err)
		}
		return rows, nil, nil

	case "get_raw_material_stock":
		queryFilter, _ := args["query"].(string)

		var rows []aiRawMaterialStockRow
		rq := h.db.WithContext(ctx).Table("products p").
			Select("p.name, p.sku, COALESCE(p.category, 'Raw Material') AS category, COALESCE(SUM(ws.quantity), 0) AS quantity, COALESCE(p.unit, 'KG') AS unit").
			Joins("LEFT JOIN warehouse_stocks ws ON p.id = ws.product_id").
			Where("p.deleted_at IS NULL AND (p.category ILIKE '%raw%' OR p.category ILIKE '%bahan%' OR p.category ILIKE '%pack%' OR p.category ILIKE '%kemasan%')")

		if companyID != uuid.Nil {
			rq = rq.Where("(p.company_id = ? OR p.company_id IS NULL)", companyID)
		}
		if queryFilter != "" {
			like := "%" + strings.TrimSpace(queryFilter) + "%"
			rq = rq.Where("(p.name ILIKE ? OR p.sku ILIKE ?)", like, like)
		}
		rq = rq.Group("p.name, p.sku, p.category, p.unit")
		if err := rq.Order("p.name ASC").Limit(30).Scan(&rows).Error; err != nil {
			return nil, nil, fmt.Errorf("gagal query stok bahan baku: %w", err)
		}
		return rows, nil, nil

	case "get_kas_pabrik_summary":
		var summary aiKasPabrikSummary
		summary.Accounts = make([]aiKasPabrikRow, 0)

		kq := h.db.WithContext(ctx).Table("kas_pabrik kp").
			Select("kp.name AS kas_name, COALESCE(b.name, 'Pabrik Pusat') AS branch_name, kp.current_balance AS balance, COALESCE(kp.account_code, '-') AS account_code").
			Joins("LEFT JOIN branches b ON kp.branch_id = b.id").
			Where("kp.is_active = TRUE")

		if companyID != uuid.Nil {
			kq = kq.Where("(kp.company_id = ? OR kp.company_id IS NULL)", companyID)
		}
		_ = kq.Order("kp.name ASC").Scan(&summary.Accounts).Error
		for _, acc := range summary.Accounts {
			summary.TotalBalance += acc.Balance
		}
		return summary, nil, nil

	case "get_marketplace_summary":
		startStr, _ := args["start_date"].(string)
		endStr, _ := args["end_date"].(string)
		periodStr, _ := args["period"].(string)
		branchName, _ := args["branch_name"].(string)
		branchIDStr, _ := args["branch_id"].(string)
		platformFilter, _ := args["marketplace"].(string)

		start, end, err := parseAIRange(startStr, endStr, periodStr)
		if err != nil {
			return nil, nil, err
		}

		var summary aiMarketplaceSummary
		summary.Period = periodStr
		summary.StartDate = start.Format("2006-01-02")
		summary.EndDate = end.Add(-24 * time.Hour).Format("2006-01-02")
		summary.FilterBranch = branchName
		summary.FilterPlatform = platformFilter
		summary.BranchBreakdown = make([]aiMarketplaceBranchRow, 0)
		summary.PlatformBreakdown = make([]aiMarketplacePlatformRow, 0)
		summary.TopSellingProducts = make([]aiMarketplaceTopProductRow, 0)

		// Resolusi Branch ID jika branch_name disediakan
		var targetBranchID *uuid.UUID
		if branchIDStr != "" {
			if bID, err := uuid.Parse(branchIDStr); err == nil {
				targetBranchID = &bID
			}
		} else if branchName != "" {
			var b struct{ ID uuid.UUID }
			_ = h.db.WithContext(ctx).Table("branches").Select("id").Where("name ILIKE ?", "%"+strings.TrimSpace(branchName)+"%").Limit(1).Scan(&b).Error
			if b.ID != uuid.Nil {
				targetBranchID = &b.ID
			}
		}

		// 1. Total Order & Total Omzet Marketplace
		moQ := h.db.WithContext(ctx).Table("marketplace_orders mo").
			Select("COUNT(mo.id) AS total_orders, COALESCE(SUM(mo.total_amount), 0) AS total_revenue").
			Where("mo.created_at >= ? AND mo.created_at < ? AND UPPER(TRIM(mo.order_status)) NOT IN ('CANCELLED', 'CANCELED', 'RETURNED')", start, end)

		if targetBranchID != nil {
			moQ = moQ.Where("mo.branch_id = ?", *targetBranchID)
		} else if branchName != "" {
			moQ = moQ.Where("mo.warehouse_name ILIKE ?", "%"+strings.TrimSpace(branchName)+"%")
		}

		if platformFilter != "" && !strings.EqualFold(platformFilter, "all") {
			moQ = moQ.Where("mo.marketplace ILIKE ?", "%"+strings.TrimSpace(platformFilter)+"%")
		}

		var agg struct {
			TotalOrders  int64
			TotalRevenue float64
		}
		if err := moQ.Scan(&agg).Error; err != nil {
			return nil, nil, fmt.Errorf("gagal query total marketplace: %w", err)
		}

		summary.TotalOrders = agg.TotalOrders
		summary.TotalRevenue = agg.TotalRevenue
		if summary.TotalOrders > 0 {
			summary.AverageOrderValue = summary.TotalRevenue / float64(summary.TotalOrders)
		}

		// 2. Branch / Warehouse Breakdown
		branchQ := h.db.WithContext(ctx).Table("marketplace_orders mo").
			Select("COALESCE(b.name, NULLIF(mo.warehouse_name, ''), 'Gudang Marketplace') AS branch_name, COUNT(mo.id) AS total_orders, COALESCE(SUM(mo.total_amount), 0) AS total_revenue").
			Joins("LEFT JOIN branches b ON mo.branch_id = b.id").
			Where("mo.created_at >= ? AND mo.created_at < ? AND UPPER(TRIM(mo.order_status)) NOT IN ('CANCELLED', 'CANCELED', 'RETURNED')", start, end)

		if targetBranchID != nil {
			branchQ = branchQ.Where("mo.branch_id = ?", *targetBranchID)
		} else if branchName != "" {
			branchQ = branchQ.Where("mo.warehouse_name ILIKE ?", "%"+strings.TrimSpace(branchName)+"%")
		}
		if platformFilter != "" && !strings.EqualFold(platformFilter, "all") {
			branchQ = branchQ.Where("mo.marketplace ILIKE ?", "%"+strings.TrimSpace(platformFilter)+"%")
		}
		_ = branchQ.Group("COALESCE(b.name, NULLIF(mo.warehouse_name, ''), 'Gudang Marketplace')").Order("total_revenue DESC").Limit(10).Scan(&summary.BranchBreakdown).Error

		// 3. Platform Breakdown (TikTok Shop, Shopee, Tokopedia, Lazada)
		platQ := h.db.WithContext(ctx).Table("marketplace_orders mo").
			Select("COALESCE(NULLIF(TRIM(mo.marketplace), ''), 'TikTok Shop') AS platform, COUNT(mo.id) AS total_orders, COALESCE(SUM(mo.total_amount), 0) AS total_revenue").
			Where("mo.created_at >= ? AND mo.created_at < ? AND UPPER(TRIM(mo.order_status)) NOT IN ('CANCELLED', 'CANCELED', 'RETURNED')", start, end)

		if targetBranchID != nil {
			platQ = platQ.Where("mo.branch_id = ?", *targetBranchID)
		} else if branchName != "" {
			platQ = platQ.Where("mo.warehouse_name ILIKE ?", "%"+strings.TrimSpace(branchName)+"%")
		}
		_ = platQ.Group("mo.marketplace").Order("total_revenue DESC").Scan(&summary.PlatformBreakdown).Error

		// 4. Top 5 Produk Terlaris di Marketplace
		prodQ := h.db.WithContext(ctx).Table("marketplace_orders mo").
			Select("COALESCE(mo.product_name, 'Produk') AS product_name, COALESCE(mo.seller_sku, '-') AS seller_sku, COALESCE(SUM(mo.item_count), 0) AS total_qty_sold, COALESCE(SUM(mo.total_amount), 0) AS total_revenue").
			Where("mo.created_at >= ? AND mo.created_at < ? AND UPPER(TRIM(mo.order_status)) NOT IN ('CANCELLED', 'CANCELED', 'RETURNED')", start, end)

		if targetBranchID != nil {
			prodQ = prodQ.Where("mo.branch_id = ?", *targetBranchID)
		} else if branchName != "" {
			prodQ = prodQ.Where("mo.warehouse_name ILIKE ?", "%"+strings.TrimSpace(branchName)+"%")
		}
		if platformFilter != "" && !strings.EqualFold(platformFilter, "all") {
			prodQ = prodQ.Where("mo.marketplace ILIKE ?", "%"+strings.TrimSpace(platformFilter)+"%")
		}
		_ = prodQ.Group("mo.product_name, mo.seller_sku").Order("total_qty_sold DESC").Limit(5).Scan(&summary.TopSellingProducts).Error

		var widget *AIWidget
		if len(summary.PlatformBreakdown) > 0 {
			var pieData []map[string]any
			for _, p := range summary.PlatformBreakdown {
				pieData = append(pieData, map[string]any{
					"name":  p.Platform,
					"value": p.TotalRevenue,
				})
			}
			widget = &AIWidget{
				Type:      "chart",
				Title:     "Pangsa Platform Marketplace",
				ChartType: "pie",
				Data: map[string]any{
					"pie_data":            pieData,
					"total_revenue":       summary.TotalRevenue,
					"total_orders":        summary.TotalOrders,
					"average_order_value": summary.AverageOrderValue,
				},
			}
		}

		return summary, widget, nil

	case "get_products_catalog":
		category, _ := args["category"].(string)
		query, _ := args["query"].(string)
		type productRow struct {
			Name         string  `json:"name"`
			SKU          string  `json:"sku"`
			Category     string  `json:"category"`
			Brand        string  `json:"brand"`
			Unit         string  `json:"unit"`
			Weight       float64 `json:"weight"`
			WeightUnit   string  `json:"weight_unit"`
			SellingPrice float64 `json:"selling_price,omitempty"`
			Description  string  `json:"description,omitempty"`
			Specs        string  `json:"specs,omitempty"`
		}
		var products []productRow
		pq := h.db.WithContext(ctx).Table("products p").
			Select("p.name, p.sku, COALESCE(p.category, 'Finished Goods') AS category, COALESCE(p.brand, '') AS brand, COALESCE(p.unit, 'PCS') AS unit, COALESCE(p.weight, 0) AS weight, COALESCE(p.weight_unit, 'KG') AS weight_unit, COALESCE(p.selling_price, 0) AS selling_price, COALESCE(p.description, '') AS description, COALESCE(p.specs, '') AS specs").
			Where("p.deleted_at IS NULL")
		if companyID != uuid.Nil {
			pq = pq.Where("(p.company_id = ? OR p.company_id IS NULL)", companyID)
		}
		if category != "" && !strings.EqualFold(category, "all") {
			pq = pq.Where("p.category ILIKE ?", "%"+strings.TrimSpace(category)+"%")
		}
		if query != "" {
			like := "%" + strings.TrimSpace(query) + "%"
			pq = pq.Where("p.name ILIKE ? OR p.sku ILIKE ? OR p.brand ILIKE ?", like, like, like)
		}
		if err := pq.Order("p.category ASC, p.name ASC").Limit(30).Scan(&products).Error; err != nil {
			return nil, nil, fmt.Errorf("gagal mengambil katalog produk: %w", err)
		}
		return products, nil, nil

	case "get_pending_visits":
		dateStr, _ := args["date"].(string)
		if dateStr == "" {
			dateStr = now.Format("2006-01-02")
		}
		type visitRow struct {
			SalesmanName string `json:"salesman_name"`
			StoreName    string `json:"store_name"`
			City         string `json:"city"`
			Status       string `json:"status"`
		}
		var rows []visitRow
		vq := h.db.WithContext(ctx).Table("sales_visit_plans svp").
			Select("COALESCE(e.first_name || ' ' || COALESCE(e.last_name, ''), 'Unknown') AS salesman_name, s.name AS store_name, COALESCE(s.distribution_city, s.regency, '') AS city, svp.status").
			Joins("JOIN employees e ON svp.employee_id = e.id").
			Joins("JOIN stores s ON svp.store_id = s.id").
			Where("svp.plan_date = ? AND svp.status IN ('PLANNED', 'PENDING')", dateStr)
		if companyID != uuid.Nil {
			vq = vq.Where("svp.company_id = ?", companyID)
		}
		_ = vq.Limit(20).Scan(&rows).Error
		return rows, nil, nil

	default:
		return nil, nil, fmt.Errorf("tool belum diimplementasikan: %s", tool)
	}
}

func (h *AIOperationalHandler) Execute(c *gin.Context) {
	var req struct {
		Tool string         `json:"tool"`
		Args map[string]any `json:"arguments"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Tool == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tool dan arguments wajib diisi"})
		return
	}
	companyID, _ := aiCompanyID(c)
	data, widget, err := h.ExecuteDirect(c.Request.Context(), companyID, req.Tool, req.Args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tool": req.Tool, "read_only": true, "data": data, "widget": widget})
}

func (h *AIOperationalHandler) SearchCustomer(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if c.ShouldBindJSON(&req) != nil || len(strings.TrimSpace(req.Query)) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query minimal 2 karakter"})
		return
	}
	companyID, _ := aiCompanyID(c)
	data, _, err := h.ExecuteDirect(c.Request.Context(), companyID, "search_customer", map[string]any{"query": req.Query})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tool": "search_customer", "read_only": true, "data": data})
}

func (h *AIOperationalHandler) CustomerSalesSummary(c *gin.Context) {
	var req struct {
		CustomerID string `json:"customer_id"`
		StartDate  string `json:"start_date"`
		EndDate    string `json:"end_date"`
	}
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request tidak valid"})
		return
	}
	companyID, _ := aiCompanyID(c)
	data, _, err := h.ExecuteDirect(c.Request.Context(), companyID, "get_customer_sales_summary", map[string]any{
		"customer_id": req.CustomerID,
		"start_date":  req.StartDate,
		"end_date":    req.EndDate,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tool": "get_customer_sales_summary", "read_only": true, "data": data})
}

func aiCompanyID(c *gin.Context) (uuid.UUID, bool) {
	if roleVal, exists := c.Get("role"); exists {
		if roleStr, ok := roleVal.(string); ok && strings.EqualFold(roleStr, "superadmin") {
			return uuid.Nil, false
		}
	}
	v, ok := c.Get("companyID")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok && id != uuid.Nil
}

func parseAIRange(startDate, endDate, period string) (time.Time, time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	period = strings.ToLower(strings.TrimSpace(period))
	period = strings.ReplaceAll(period, " ", "_")

	switch period {
	case "hari_ini", "today", "hariini":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 1)
		return start, end, nil
	case "kemarin", "yesterday":
		end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		start := end.AddDate(0, 0, -1)
		return start, end, nil
	case "minggu_ini", "this_week", "mingguini":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Minggu
		}
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -(weekday - 1))
		end := start.AddDate(0, 0, 7)
		return start, end, nil
	case "bulan_ini", "this_month", "bulanini":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0)
		return start, end, nil
	case "bulan_lalu", "last_month", "bulanlalu":
		start := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, loc)
		end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		return start, end, nil
	case "tahun_ini", "this_year", "tahunini":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		end := start.AddDate(1, 0, 0)
		return start, end, nil
	case "all_time", "semua":
		start := time.Date(2020, 1, 1, 0, 0, 0, 0, loc)
		end := time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, loc)
		return start, end, nil
	}

	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)

	if len(startDate) == 7 && strings.Count(startDate, "-") == 1 { // e.g. "2026-10"
		parsedMonth, err := time.ParseInLocation("2006-01", startDate, loc)
		if err == nil {
			start := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, loc)
			end := start.AddDate(0, 1, 0)
			return start, end, nil
		}
	}

	if startDate == "" {
		startDate = now.Format("2006-01-02")[:8] + "01"
	}
	start, err := time.ParseInLocation("2006-01-02", startDate, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("format start_date harus YYYY-MM-DD atau YYYY-MM")
	}

	var end time.Time
	if endDate == "" {
		firstOfNextMonth := time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, loc)
		end = firstOfNextMonth
	} else {
		parsedEnd, err := time.ParseInLocation("2006-01-02", endDate, loc)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("format end_date harus YYYY-MM-DD")
		}
		if parsedEnd.Before(start) {
			return time.Time{}, time.Time{}, fmt.Errorf("end_date tidak boleh sebelum start_date")
		}
		end = parsedEnd.AddDate(0, 0, 1)
	}

	if end.Sub(start) > 730*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("periode maksimal 2 tahun")
	}
	return start, end, nil
}

func cleanSearchKeyword(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune("?!.,\"'()[]{}:;-_", r) {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")

	stopWords := map[string]bool{
		"siapa": true, "itu": true, "si": true, "sih": true, "adalah": true,
		"karyawan": true, "karyawannya": true, "staf": true, "staff": true, "pegawai": true,
		"pak": true, "bapak": true, "bu": true, "ibu": true, "mas": true, "mba": true, "mbak": true,
		"profil": true, "biodata": true, "kontak": true, "tentang": true, "kenal": true, "dengan": true,
		"apakah": true, "ada": true, "tau": true, "tahu": true, "tolong": true, "cari": true,
		"carikan": true, "info": true, "informasi": true, "data": true, "orang": true, "nama": true,
		"dong": true, "ya": true, "kah": true,
	}

	words := strings.Fields(s)
	var filtered []string
	for _, w := range words {
		if !stopWords[strings.ToLower(w)] {
			filtered = append(filtered, w)
		}
	}

	if len(filtered) > 0 {
		return strings.Join(filtered, " ")
	}
	return strings.TrimSpace(raw)
}
