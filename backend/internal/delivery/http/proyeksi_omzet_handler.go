package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type ProyeksiOmzetHandler struct {
	db *gorm.DB
}

func NewProyeksiOmzetHandler(db *gorm.DB) *ProyeksiOmzetHandler {
	return &ProyeksiOmzetHandler{db: db}
}

func (h *ProyeksiOmzetHandler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/admin/sales/proyeksi-omzet")
	group.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleManager), string(domain.RoleAdminNota), string(domain.RoleSupervisorSales), string(domain.RoleSupervisorSalesMT)))
	{
		group.GET("", h.GetProyeksiOmzet)
	}
}

func (h *ProyeksiOmzetHandler) GetProyeksiOmzet(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		// Default to tomorrow if empty
		dateStr = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	branchID := c.Query("branch_id")

	dateFilterShipment := "DATE(db.delivery_date) = ?"
	dateFilterMotoris := "DATE(st.transaction_date AT TIME ZONE 'Asia/Jakarta') = ?"

	if len(dateStr) == 7 { // YYYY-MM
		dateFilterShipment = "TO_CHAR(db.delivery_date, 'YYYY-MM') = ?"
		dateFilterMotoris = "TO_CHAR(st.transaction_date AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM') = ?"
	}

	// 1. Calculate Omzet from TO Shipments (Pengiriman)
	var shipmentRows []struct {
		DeliveryOrderNo string  `json:"delivery_order_no"`
		VehiclePlateNo  string  `json:"vehicle_plate_no"`
		TotalOmzet      float64 `json:"total_omzet"`
	}

	shipmentSQL := `
		SELECT 
			db.delivery_order_no,
			COALESCE(v.license_plate, v.name, '') AS vehicle_plate_no,
			COALESCE(SUM(soi.price * soi.ordered_quantity), 0) AS total_omzet
		FROM delivery_batches db
		JOIN delivery_items di ON di.delivery_batch_id = db.id
		JOIN sales_orders so ON so.id = di.sales_order_id
		JOIN sales_order_items soi ON soi.sales_order_id = so.id
		LEFT JOIN vehicles v ON v.id = db.vehicle_id
		WHERE ` + dateFilterShipment + `
		AND db.status != 'CANCELLED'
	`
	argsShipment := []interface{}{dateStr}
	if branchID != "" && branchID != "ALL_BRANCHES" {
		shipmentSQL += " AND db.branch_id = ?"
		argsShipment = append(argsShipment, branchID)
	}
	shipmentSQL += " GROUP BY db.id, db.delivery_order_no, v.license_plate, v.name"

	if err := h.db.Raw(shipmentSQL, argsShipment...).Scan(&shipmentRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch TO shipments omzet"})
		return
	}

	totalShipmentOmzet := 0.0
	for _, row := range shipmentRows {
		totalShipmentOmzet += row.TotalOmzet
	}

	// 2. Calculate Omzet from Sales Motoris
	var motorisRows []struct {
		SalesmanName string  `json:"salesman_name"`
		TotalOmzet   float64 `json:"total_omzet"`
	}

	motorisSQL := `
		SELECT 
			CONCAT(e.first_name, ' ', e.last_name) AS salesman_name,
			COALESCE(SUM(st.total_amount), 0) AS total_omzet
		FROM sales_transactions st
		JOIN employees e ON e.id = st.employee_id
		JOIN job_positions jp ON jp.id = e.job_position_id
		WHERE ` + dateFilterMotoris + `
		AND (jp.title ILIKE '%motoris%' OR jp.title ILIKE '%booster%')
	`
	argsMotoris := []interface{}{dateStr}
	if branchID != "" && branchID != "ALL_BRANCHES" {
		motorisSQL += " AND e.branch_id = ?"
		argsMotoris = append(argsMotoris, branchID)
	}
	motorisSQL += " GROUP BY e.id, e.first_name, e.last_name"

	if err := h.db.Raw(motorisSQL, argsMotoris...).Scan(&motorisRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch Motoris omzet"})
		return
	}

	totalMotorisOmzet := 0.0
	for _, row := range motorisRows {
		totalMotorisOmzet += row.TotalOmzet
	}

	// 3. Calculate Omzet from Pending TO (Sales Order belum dikirim)
	var pendingTORows []struct {
		SONumber     string  `json:"so_number"`
		SalesmanName string  `json:"salesman_name"`
		TotalOmzet   float64 `json:"total_omzet"`
	}

	dateFilterPendingTO := "DATE(so.order_date AT TIME ZONE 'Asia/Jakarta') = ?"
	if len(dateStr) == 7 { // YYYY-MM
		dateFilterPendingTO = "TO_CHAR(so.order_date AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM') = ?"
	}

	pendingTOSQL := `
		SELECT 
			so.so_number,
			CONCAT(e.first_name, ' ', e.last_name) AS salesman_name,
			COALESCE(SUM(soi.price * soi.ordered_quantity), 0) AS total_omzet
		FROM sales_orders so
		JOIN sales_order_items soi ON soi.sales_order_id = so.id
		JOIN employees e ON e.id = so.employee_id
		WHERE ` + dateFilterPendingTO + `
		AND so.status IN ('DRAFT', 'CONFIRMED', 'APPROVED')
		AND so.is_mt = false
	`
	argsPendingTO := []interface{}{dateStr}
	if branchID != "" && branchID != "ALL_BRANCHES" {
		pendingTOSQL += " AND so.branch_id = ?"
		argsPendingTO = append(argsPendingTO, branchID)
	}
	pendingTOSQL += " GROUP BY so.id, so.so_number, e.first_name, e.last_name"

	if err := h.db.Raw(pendingTOSQL, argsPendingTO...).Scan(&pendingTORows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pending TO omzet"})
		return
	}

	totalPendingTOOmzet := 0.0
	for _, row := range pendingTORows {
		totalPendingTOOmzet += row.TotalOmzet
	}

	// 4. Construct response
	if shipmentRows == nil {
		shipmentRows = make([]struct {
			DeliveryOrderNo string  `json:"delivery_order_no"`
			VehiclePlateNo  string  `json:"vehicle_plate_no"`
			TotalOmzet      float64 `json:"total_omzet"`
		}, 0)
	}
	if motorisRows == nil {
		motorisRows = make([]struct {
			SalesmanName string  `json:"salesman_name"`
			TotalOmzet   float64 `json:"total_omzet"`
		}, 0)
	}

	if pendingTORows == nil {
		pendingTORows = make([]struct {
			SONumber     string  `json:"so_number"`
			SalesmanName string  `json:"salesman_name"`
			TotalOmzet   float64 `json:"total_omzet"`
		}, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"date":                   dateStr,
		"to_shipments":           shipmentRows,
		"total_to_omzet":         totalShipmentOmzet,
		"motoris_sales":          motorisRows,
		"total_motoris_omzet":    totalMotorisOmzet,
		"pending_to_sales":       pendingTORows,
		"total_pending_to_omzet": totalPendingTOOmzet,
		"grand_total_omzet":      totalShipmentOmzet + totalMotorisOmzet + totalPendingTOOmzet,
	})
}
