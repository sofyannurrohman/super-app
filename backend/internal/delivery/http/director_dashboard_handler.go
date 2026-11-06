package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DirectorDashboardHandler struct {
	db *gorm.DB
}

func NewDirectorDashboardHandler(db *gorm.DB) *DirectorDashboardHandler {
	return &DirectorDashboardHandler{db: db}
}

func (h *DirectorDashboardHandler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/admin/sales/director-summary")
	group.Use(DirectorRoleMiddleware(h.db))
	{
		group.GET("", h.GetDirectorSummary)
		group.GET("/chart", h.GetDirectorChart)
	}
}

func (h *DirectorDashboardHandler) GetDirectorSummary(c *gin.Context) {
	companyIDVal, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company not identified"})
		return
	}
	companyID := companyIDVal.(string)

	dateStr := c.Query("date")
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.Local
	}
	if dateStr == "" {
		dateStr = time.Now().In(loc).Format("2006-01-02")
	}

	// Parse tanggal yang diminta
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}
	yesterdayStr := t.AddDate(0, 0, -1).Format("2006-01-02")
	monthStr := t.Format("2006-01")

	// Hitung range awal & akhir bulan untuk filter sargable (bisa pakai index)
	monthStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
	monthEnd := monthStart.AddDate(0, 1, 0) // awal bulan berikutnya (exclusive)
	monthStartStr := monthStart.Format("2006-01-02")
	monthEndStr := monthEnd.Format("2006-01-02")

	var rows []struct {
		BranchID       string  `json:"branch_id"`
		BranchName     string  `json:"branch_name"`
		TodayOmzet     float64 `json:"today_omzet"`
		YesterdayOmzet float64 `json:"yesterday_omzet"`
		MonthOmzet     float64 `json:"month_omzet"`
	}

	// Gunakan range comparison (>= start AND < end) agar PostgreSQL bisa pakai btree index
	// pada kolom transaction_date, delivery_date, order_date — jauh lebih cepat dari TO_CHAR
	sqlQuery := fmt.Sprintf(`
		WITH AllSales AS (
			-- Motoris Sales: filter pakai range agar index terpakai
			SELECT
				e.branch_id,
				DATE(st.transaction_date AT TIME ZONE 'Asia/Jakarta') as sale_date,
				st.total_amount as amount
			FROM sales_transactions st
			JOIN employees e ON e.id = st.employee_id
			WHERE st.transaction_date >= '%s' AND st.transaction_date < '%s'

			UNION ALL

			-- Shipped TO / MT
			SELECT
				db.branch_id,
				DATE(db.delivery_date) as sale_date,
				(soi.price * soi.ordered_quantity) as amount
			FROM delivery_batches db
			JOIN delivery_items di ON di.delivery_batch_id = db.id
			JOIN sales_orders so ON so.id = di.sales_order_id
			JOIN sales_order_items soi ON soi.sales_order_id = so.id
			WHERE db.status != 'CANCELLED'
			AND db.delivery_date >= '%s' AND db.delivery_date < '%s'

			UNION ALL

			-- Pending TO
			SELECT
				so.branch_id,
				DATE(so.order_date AT TIME ZONE 'Asia/Jakarta') as sale_date,
				(soi.price * soi.ordered_quantity) as amount
			FROM sales_orders so
			JOIN sales_order_items soi ON soi.sales_order_id = so.id
			WHERE so.status IN ('DRAFT', 'CONFIRMED', 'APPROVED')
			AND so.is_mt = false
			AND so.order_date >= '%s' AND so.order_date < '%s'
		)
		SELECT
			b.id as branch_id,
			b.name as branch_name,
			COALESCE(SUM(CASE WHEN sale_date::text = '%s' THEN a.amount ELSE 0 END), 0) as today_omzet,
			COALESCE(SUM(CASE WHEN sale_date::text = '%s' THEN a.amount ELSE 0 END), 0) as yesterday_omzet,
			COALESCE(SUM(a.amount), 0) as month_omzet
		FROM branches b
		LEFT JOIN AllSales a ON a.branch_id = b.id
		WHERE b.company_id = ?
		GROUP BY b.id, b.name
		ORDER BY b.name
	`,
		monthStartStr, monthEndStr,
		monthStartStr, monthEndStr,
		monthStartStr, monthEndStr,
		dateStr, yesterdayStr,
	)

	if err := h.db.Raw(sqlQuery, companyID).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch director summary: " + err.Error()})
		return
	}

	grandToday := 0.0
	grandYesterday := 0.0
	grandMonth := 0.0

	for i := range rows {
		rows[i].BranchName = strings.ToUpper(rows[i].BranchName)
		grandToday += rows[i].TodayOmzet
		grandYesterday += rows[i].YesterdayOmzet
		grandMonth += rows[i].MonthOmzet
	}
	var transactionStats struct {
		Total     int64   `json:"total"`
		Completed int64   `json:"completed"`
		Pending   int64   `json:"pending"`
		Cancelled int64   `json:"cancelled"`
		Average   float64 `json:"average"`
	}
	h.db.Raw(`SELECT COUNT(*) FILTER (WHERE status NOT IN ('CANCELLED','REJECTED')) total,
		COUNT(*) FILTER (WHERE status IN ('VERIFIED','COMPLETED')) completed,
		COUNT(*) FILTER (WHERE status = 'PENDING') pending,
		COUNT(*) FILTER (WHERE status IN ('CANCELLED','REJECTED')) cancelled,
		COALESCE(AVG(total_amount) FILTER (WHERE status NOT IN ('CANCELLED','REJECTED')), 0) average
		FROM sales_transactions WHERE company_id = ? AND transaction_date >= ? AND transaction_date < ?`, companyID, monthStartStr, monthEndStr).Scan(&transactionStats)

	c.JSON(http.StatusOK, gin.H{
		"date":              dateStr,
		"month":             monthStr,
		"branches":          rows,
		"grand_today":       grandToday,
		"grand_yesterday":   grandYesterday,
		"grand_month":       grandMonth,
		"transaction_stats": transactionStats,
	})
}

func (h *DirectorDashboardHandler) GetDirectorChart(c *gin.Context) {
	companyIDVal, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company not identified"})
		return
	}
	companyID := companyIDVal.(string)

	monthStr := c.Query("month")
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.Local
	}
	if monthStr == "" {
		monthStr = time.Now().In(loc).Format("2006-01")
	}
	branchID := c.Query("branch_id")
	if branchID == "" {
		branchID = "all"
	}

	// Hitung range bulan untuk filter sargable
	t, err := time.ParseInLocation("2006-01", monthStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format, use YYYY-MM"})
		return
	}
	monthStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
	monthEnd := monthStart.AddDate(0, 1, 0)
	monthStartStr := monthStart.Format("2006-01-02")
	monthEndStr := monthEnd.Format("2006-01-02")

	var rows []struct {
		Date       string  `json:"date"`
		TotalOmzet float64 `json:"total_omzet"`
	}

	sqlQuery := fmt.Sprintf(`
		WITH AllSales AS (
			SELECT e.branch_id, DATE(st.transaction_date AT TIME ZONE 'Asia/Jakarta') as sale_date, st.total_amount as amount
			FROM sales_transactions st JOIN employees e ON e.id = st.employee_id
			WHERE st.transaction_date >= '%s' AND st.transaction_date < '%s'
			UNION ALL
			SELECT db.branch_id, DATE(db.delivery_date) as sale_date, (soi.price * soi.ordered_quantity) as amount
			FROM delivery_batches db JOIN delivery_items di ON di.delivery_batch_id = db.id
			JOIN sales_orders so ON so.id = di.sales_order_id JOIN sales_order_items soi ON soi.sales_order_id = so.id
			WHERE db.status != 'CANCELLED' AND db.delivery_date >= '%s' AND db.delivery_date < '%s'
			UNION ALL
			SELECT so.branch_id, DATE(so.order_date AT TIME ZONE 'Asia/Jakarta') as sale_date, (soi.price * soi.ordered_quantity) as amount
			FROM sales_orders so JOIN sales_order_items soi ON soi.sales_order_id = so.id
			WHERE so.status IN ('DRAFT', 'CONFIRMED', 'APPROVED') AND so.is_mt = false
			AND so.order_date >= '%s' AND so.order_date < '%s'
		)
		SELECT sale_date::text AS date, COALESCE(SUM(a.amount), 0) as total_omzet
		FROM branches b
		LEFT JOIN AllSales a ON a.branch_id = b.id
		WHERE b.company_id = ? AND (? = 'all' OR b.id::text = ?)
		AND a.sale_date IS NOT NULL
		GROUP BY a.sale_date
		ORDER BY a.sale_date
	`,
		monthStartStr, monthEndStr,
		monthStartStr, monthEndStr,
		monthStartStr, monthEndStr,
	)

	if err := h.db.Raw(sqlQuery, companyID, branchID, branchID).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rows)
}
