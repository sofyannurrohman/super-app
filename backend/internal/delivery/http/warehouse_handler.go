package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"gorm.io/gorm"
)

type WarehouseHandler struct {
	usecase usecase.WarehouseUsecase
	db      *gorm.DB
}

func NewWarehouseHandler(u usecase.WarehouseUsecase, db *gorm.DB) *WarehouseHandler {
	return &WarehouseHandler{usecase: u, db: db}
}

func parseBranchID(c *gin.Context) uuid.UUID {
	bid := c.Query("branch_id")
	if bid == "ALL_BRANCHES" {
		return uuid.Nil
	}
	if bid != "" {
		if id, err := uuid.Parse(bid); err == nil {
			return id
		}
	}
	if val, ok := c.Get("branch_id"); ok && val != "" {
		if id, err := uuid.Parse(val.(string)); err == nil {
			return id
		}
	}
	return uuid.Nil
}

func (h *WarehouseHandler) RegisterRoutes(r *gin.RouterGroup) {
	warehouse := r.Group("/warehouse")
	{
		// GET /warehouse/stock and dashboard overview accessible by authenticated employees
		warehouse.GET("/stock", h.GetInventory)
		warehouse.GET("/dashboard/overview", h.GetDashboardOverview)

		// Restricted warehouse actions (Gudang Admin, HR, Manager, etc.)
		adminGroup := warehouse.Group("")
		adminGroup.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager), string(domain.RoleAdminGudang)))
		{
			adminGroup.GET("/logs", h.GetLogs)
			adminGroup.GET("/transfers/pending", h.GetPendingShipments)
			adminGroup.GET("/transfers/all", h.GetAllTransfers)
			adminGroup.POST("/transfers/:id/receive", h.ReceiveShipment)
			adminGroup.POST("/transfers/:id/correct", h.CorrectShipmentReceipt)
			adminGroup.POST("/transfers/:id/arrive", h.ConfirmArrival)
			adminGroup.GET("/transfers/do/:do_no", h.GetByDO)
			adminGroup.POST("/transfers/do/:do_no/receive", h.ReceiveByDO)
			adminGroup.POST("/transfers/:id/approve", h.ApproveShipment)
			adminGroup.POST("/transfers/:id/reject", h.RejectShipment)
			adminGroup.POST("/transfers/:id/confirm-shipment", h.ConfirmShipment)
			adminGroup.POST("/adjust", h.AdjustStock)
			adminGroup.DELETE("/stock/:id", h.DeleteStock)
			adminGroup.POST("/stock-limit", h.SetStockLimit)
			adminGroup.POST("/dispatch/invoice/:receipt_no", h.DispatchByInvoice)
		}
	}
}

func (h *WarehouseHandler) GetInventory(c *gin.Context) {
	branchID := parseBranchID(c)

	inventory, err := h.usecase.GetInventory(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func (h *WarehouseHandler) GetLogs(c *gin.Context) {
	branchID := parseBranchID(c)

	logs, err := h.usecase.GetWarehouseLogs(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *WarehouseHandler) GetPendingShipments(c *gin.Context) {
	branchID := parseBranchID(c)

	fmt.Printf("[API] Hit GetPendingShipments for branch: %s\n", branchID)
	transfers, err := h.usecase.GetPendingShipments(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, transfers)
}

func (h *WarehouseHandler) GetAllTransfers(c *gin.Context) {
	branchID := parseBranchID(c)

	transfers, err := h.usecase.GetAllTransfers(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, transfers)
}

func (h *WarehouseHandler) ReceiveShipment(c *gin.Context) {
	transferID := uuid.MustParse(c.Param("id"))
	var req struct {
		ActualReceivedQuantity int    `json:"actual_received_quantity"`
		DamagedQuantity        int    `json:"damaged_quantity"`
		Notes                  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	processedBy := "Admin Gudang"
	if val, ok := c.Get("username"); ok && val != "" {
		processedBy = val.(string)
	} else if val, ok := c.Get("name"); ok && val != "" {
		processedBy = val.(string)
	}

	if err := h.usecase.ReceiveShipment(transferID, req.ActualReceivedQuantity, req.DamagedQuantity, req.Notes, processedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Shipment received and stock updated"})
}

func (h *WarehouseHandler) ApproveShipment(c *gin.Context) {
	transferID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pengiriman tidak valid"})
		return
	}

	var approverID *uuid.UUID
	if val, ok := c.Get("employeeID"); ok && val != nil {
		if empID, ok := val.(uuid.UUID); ok && empID != uuid.Nil {
			approverID = &empID
		}
	}
	if approverID == nil {
		if val, ok := c.Get("userID"); ok && val != nil {
			if uid, ok := val.(uuid.UUID); ok && uid != uuid.Nil {
				approverID = &uid
			}
		}
	}

	approverName := ""
	if val, ok := c.Get("name"); ok && val != nil {
		approverName = val.(string)
	}

	approverPosition := "Admin Gudang"
	if val, ok := c.Get("job_position"); ok && val != nil && val.(string) != "" {
		approverPosition = val.(string)
	} else if rVal, ok := c.Get("role"); ok && rVal != nil {
		approverPosition = strings.ToUpper(strings.ReplaceAll(rVal.(string), "_", " "))
	}

	if err := h.usecase.ApproveTransfer(transferID, approverID, approverName, approverPosition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Shipment approved by branch"})
}

func (h *WarehouseHandler) RejectShipment(c *gin.Context) {
	transferID := uuid.MustParse(c.Param("id"))
	var req struct {
		RejectionReason string `json:"rejection_reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alasan penolakan wajib diisi"})
		return
	}

	if err := h.usecase.RejectTransfer(transferID, req.RejectionReason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Shipment rejected by branch"})
}

func (h *WarehouseHandler) ConfirmShipment(c *gin.Context) {
	transferID := uuid.MustParse(c.Param("id"))
	if err := h.usecase.ConfirmShipment(transferID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Surat Jalan telah disetujui dan pengiriman sedang diproses"})
}

func (h *WarehouseHandler) ConfirmArrival(c *gin.Context) {
	transferID := uuid.MustParse(c.Param("id"))
	if err := h.usecase.ConfirmArrival(transferID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status pengiriman diperbarui menjadi ARRIVED (Siap Bongkar)"})
}

func (h *WarehouseHandler) SetStockLimit(c *gin.Context) {
	var req struct {
		ProductID uuid.UUID `json:"product_id" binding:"required"`
		Limit     int       `json:"limit"`
		BatchNo   string    `json:"batch_no"`
		BranchID  string    `json:"branch_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var branchID uuid.UUID
	// 1. Prioritaskan branch_id dari payload request jika spesifik
	if req.BranchID != "" && req.BranchID != "ALL_BRANCHES" && req.BranchID != "ALL" {
		if parsed, err := uuid.Parse(req.BranchID); err == nil {
			branchID = parsed
		}
	}

	// 2. Jika tidak ada di payload, fallback ke query param atau session context
	if branchID == uuid.Nil && req.BranchID != "ALL_BRANCHES" && req.BranchID != "ALL" {
		branchID = parseBranchID(c)
	}

	// 3. Set stock limit via usecase (branchID == uuid.Nil akan mengupdate seluruh cabang)
	if err := h.usecase.SetStockLimit(branchID, req.ProductID, req.Limit, req.BatchNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock limit updated successfully"})
}

func (h *WarehouseHandler) AdjustStock(c *gin.Context) {
	var req struct {
		ProductID uuid.UUID `json:"product_id" binding:"required"`
		Quantity  *int      `json:"quantity" binding:"required"`
		Reason    string    `json:"reason"`
		BatchNo   string    `json:"batch_no"`
		StockType string    `json:"stock_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branchID := parseBranchID(c)

	if branchID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id not found in session or query"})
		return
	}

	processedBy := "Admin Gudang"
	if val, ok := c.Get("username"); ok && val != "" {
		processedBy = val.(string)
	} else if val, ok := c.Get("name"); ok && val != "" {
		processedBy = val.(string)
	}

	if err := h.usecase.AdjustStock(branchID, req.ProductID, *req.Quantity, req.Reason, req.BatchNo, processedBy, req.StockType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock adjusted successfully"})
}

func (h *WarehouseHandler) DeleteStock(c *gin.Context) {
	idParam := c.Param("id")
	stockID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID stok tidak valid"})
		return
	}

	var req struct {
		Reason   string `json:"reason"`
		BranchID string `json:"branch_id"`
	}
	_ = c.ShouldBindJSON(&req)

	branchID := parseBranchID(c)
	if branchID == uuid.Nil && req.BranchID != "" && req.BranchID != "ALL_BRANCHES" {
		if parsed, err := uuid.Parse(req.BranchID); err == nil {
			branchID = parsed
		}
	}

	processedBy := "Admin Gudang"
	if val, ok := c.Get("username"); ok && val != "" {
		processedBy = val.(string)
	} else if val, ok := c.Get("name"); ok && val != "" {
		processedBy = val.(string)
	}

	if err := h.usecase.DeleteStock(stockID, branchID, req.Reason, processedBy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Data stok berhasil dihapus"})
}

func (h *WarehouseHandler) GetByDO(c *gin.Context) {
	doNo := c.Param("do_no")
	transfer, err := h.usecase.GetTransferByDO(doNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Surat jalan tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, transfer)
}

func (h *WarehouseHandler) ReceiveByDO(c *gin.Context) {
	doNo := c.Param("do_no")
	var req struct {
		ActualReceivedQuantity int    `json:"actual_received_quantity"`
		DamagedQuantity        int    `json:"damaged_quantity"`
		Notes                  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	processedBy := "Admin Gudang"
	if val, ok := c.Get("username"); ok && val != "" {
		processedBy = val.(string)
	} else if val, ok := c.Get("name"); ok && val != "" {
		processedBy = val.(string)
	}

	if err := h.usecase.ReceiveShipmentByDO(doNo, req.ActualReceivedQuantity, req.DamagedQuantity, req.Notes, processedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Barang telah diterima dan stok gudang diperbarui"})
}

func (h *WarehouseHandler) DispatchByInvoice(c *gin.Context) {
	receiptNo := c.Param("receipt_no")

	branchID := parseBranchID(c)

	if branchID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id not found in session or query"})
		return
	}

	if err := h.usecase.DispatchByInvoice(receiptNo, branchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Barang telah dikeluarkan dari gudang dan status nota diperbarui"})
}

func (h *WarehouseHandler) CorrectShipmentReceipt(c *gin.Context) {
	transferID := uuid.MustParse(c.Param("id"))
	var req struct {
		ActualReceivedQuantity int    `json:"actual_received_quantity"`
		DamagedQuantity        int    `json:"damaged_quantity"`
		Reason                 string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alasan koreksi wajib diisi"})
		return
	}

	// Dapatkan username / nama admin dari context JWT
	correctedBy := "Admin Gudang"
	if val, ok := c.Get("username"); ok && val != "" {
		correctedBy = val.(string)
	} else if val, ok := c.Get("name"); ok && val != "" {
		correctedBy = val.(string)
	}

	if err := h.usecase.CorrectShipmentReceipt(transferID, req.ActualReceivedQuantity, req.DamagedQuantity, req.Reason, correctedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Penerimaan barang berhasil dikoreksi dan stok diperbarui"})
}

// Response structs for Warehouse Dashboard Overview
type WarehouseDashboardOverviewResponse struct {
	Summary                 WarehouseDashboardSummary      `json:"summary"`
	MutationTrend           []WarehouseMutationTrendPoint  `json:"mutation_trend"`
	TopProducts             []WarehouseTopStockProduct     `json:"top_products"`
	StockHealth             WarehouseStockHealthSummary    `json:"stock_health"`
	LowStockItems           []WarehouseLowStockItem        `json:"low_stock_items"`
	PendingInboundShipments []WarehousePendingShipmentItem `json:"pending_inbound_shipments"`
	RecentLogs              []WarehouseRecentLogItem       `json:"recent_logs"`
}

type WarehouseDashboardSummary struct {
	TotalStockPhysical         int `json:"total_stock_physical"`
	TotalSKUCount              int `json:"total_sku_count"`
	LowStockCount              int `json:"low_stock_count"`
	OutOfStockCount            int `json:"out_of_stock_count"`
	PendingInboundCount        int `json:"pending_inbound_count"`
	PendingSalesTransfersCount int `json:"pending_sales_transfers_count"`
	TodayInQuantity            int `json:"today_in_quantity"`
	TodayOutQuantity           int `json:"today_out_quantity"`
}

type WarehouseMutationTrendPoint struct {
	Date      string `json:"date"`
	Label     string `json:"label"`
	InQty     int    `json:"in_qty"`
	OutQty    int    `json:"out_qty"`
	NetChange int    `json:"net_change"`
}

type WarehouseTopStockProduct struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	SKU        string  `json:"sku"`
	Unit       string  `json:"unit"`
	Quantity   int     `json:"quantity"`
	Percentage float64 `json:"percentage"`
}

type WarehouseStockHealthSummary struct {
	SafeCount       int `json:"safe_count"`         // qty >= min_limit
	WarningCount    int `json:"warning_count"`      // 0.5 * min_limit <= qty < min_limit
	CriticalCount   int `json:"critical_count"`     // 0 < qty < 0.5 * min_limit
	OutOfStockCount int `json:"out_of_stock_count"` // qty == 0
}

type WarehouseLowStockItem struct {
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	SKU          string  `json:"sku"`
	Unit         string  `json:"unit"`
	CurrentStock int     `json:"current_stock"`
	MinLimit     int     `json:"min_limit"`
	Deficit      int     `json:"deficit"`
	StockPct     float64 `json:"stock_pct"`
}

type WarehousePendingShipmentItem struct {
	ID              string `json:"id"`
	DeliveryOrderNo string `json:"delivery_order_no"`
	FromFactory     string `json:"from_factory"`
	ToBranch        string `json:"to_branch"`
	ProductName     string `json:"product_name"`
	Quantity        int    `json:"quantity"`
	Unit            string `json:"unit"`
	Status          string `json:"status"`
	VehiclePlate    string `json:"vehicle_plate"`
	DriverName      string `json:"driver_name"`
	CreatedAt       string `json:"created_at"`
	TargetDate      string `json:"target_date"`
}

type WarehouseRecentLogItem struct {
	ID          string `json:"id"`
	ProductName string `json:"product_name"`
	SKU         string `json:"sku"`
	Type        string `json:"type"` // "IN" or "OUT"
	Quantity    int    `json:"quantity"`
	Source      string `json:"source"`
	ProcessedBy string `json:"processed_by"`
	ApprovedBy  string `json:"approved_by"`
	CreatedAt   string `json:"created_at"`
	TimeAgo     string `json:"time_ago"`
}

func (h *WarehouseHandler) GetDashboardOverview(c *gin.Context) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")

	branchID := parseBranchID(c)
	rangeParam := c.DefaultQuery("range", "7d") // 7d, 30d, year

	resp := WarehouseDashboardOverviewResponse{
		MutationTrend:           make([]WarehouseMutationTrendPoint, 0),
		TopProducts:             make([]WarehouseTopStockProduct, 0),
		LowStockItems:           make([]WarehouseLowStockItem, 0),
		PendingInboundShipments: make([]WarehousePendingShipmentItem, 0),
		RecentLogs:              make([]WarehouseRecentLogItem, 0),
	}

	// 1. Physical Stocks & Grouped Product Aggregations
	type productStockRow struct {
		ProductID string
		Name      string
		SKU       string
		Unit      string
		TotalQty  int
		MinLimit  int
	}
	var stockRows []productStockRow

	sq := h.db.Table("warehouse_stocks ws").
		Select(`
			ws.product_id,
			p.name,
			p.sku,
			COALESCE(p.unit, 'PCS') as unit,
			SUM(ws.quantity) as total_qty,
			MAX(ws.min_limit) as min_limit
		`).
		Joins("JOIN products p ON ws.product_id = p.id")

	if branchID != uuid.Nil {
		sq = sq.Where("ws.branch_id = ?", branchID)
	}
	sq.Group("ws.product_id, p.name, p.sku, p.unit").Scan(&stockRows)

	totalPhysical := 0
	totalSKUs := len(stockRows)
	lowCount := 0
	outCount := 0
	safeCount := 0
	warningCount := 0
	criticalCount := 0

	for _, sr := range stockRows {
		totalPhysical += sr.TotalQty

		if sr.TotalQty == 0 {
			outCount++
		}

		if sr.MinLimit > 0 && sr.TotalQty < sr.MinLimit {
			lowCount++
			deficit := sr.MinLimit - sr.TotalQty
			pct := 0.0
			if sr.MinLimit > 0 {
				pct = (float64(sr.TotalQty) / float64(sr.MinLimit)) * 100
			}
			resp.LowStockItems = append(resp.LowStockItems, WarehouseLowStockItem{
				ProductID:    sr.ProductID,
				ProductName:  sr.Name,
				SKU:          sr.SKU,
				Unit:         sr.Unit,
				CurrentStock: sr.TotalQty,
				MinLimit:     sr.MinLimit,
				Deficit:      deficit,
				StockPct:     pct,
			})
		}

		// Health categorization
		if sr.TotalQty == 0 {
			// handled
		} else if sr.MinLimit > 0 && float64(sr.TotalQty) < float64(sr.MinLimit)*0.5 {
			criticalCount++
		} else if sr.MinLimit > 0 && sr.TotalQty < sr.MinLimit {
			warningCount++
		} else {
			safeCount++
		}
	}

	resp.Summary.TotalStockPhysical = totalPhysical
	resp.Summary.TotalSKUCount = totalSKUs
	resp.Summary.LowStockCount = lowCount
	resp.Summary.OutOfStockCount = outCount

	resp.StockHealth = WarehouseStockHealthSummary{
		SafeCount:       safeCount,
		WarningCount:    warningCount,
		CriticalCount:   criticalCount,
		OutOfStockCount: outCount,
	}

	// Sort Top Products by Quantity (Top 7)
	// Sort by quantity desc
	for i := 0; i < len(stockRows); i++ {
		for j := i + 1; j < len(stockRows); j++ {
			if stockRows[j].TotalQty > stockRows[i].TotalQty {
				stockRows[i], stockRows[j] = stockRows[j], stockRows[i]
			}
		}
	}

	limitTop := 7
	if len(stockRows) < limitTop {
		limitTop = len(stockRows)
	}
	for i := 0; i < limitTop; i++ {
		sr := stockRows[i]
		pct := 0.0
		if totalPhysical > 0 {
			pct = (float64(sr.TotalQty) / float64(totalPhysical)) * 100
		}
		resp.TopProducts = append(resp.TopProducts, WarehouseTopStockProduct{
			ID:         sr.ProductID,
			Name:       sr.Name,
			SKU:        sr.SKU,
			Unit:       sr.Unit,
			Quantity:   sr.TotalQty,
			Percentage: pct,
		})
	}

	// 2. Pending Inbound Shipments from Factory
	var pendingInboundCount int64
	inbQ := h.db.Table("product_transfers pt").
		Where("pt.status IN ('LOADING', 'SHIPPED', 'ARRIVED', 'REQUESTED')")
	if branchID != uuid.Nil {
		inbQ = inbQ.Where("pt.to_branch_id = ?", branchID)
	}
	inbQ.Count(&pendingInboundCount)
	resp.Summary.PendingInboundCount = int(pendingInboundCount)

	type inbRow struct {
		ID              string
		DeliveryOrderNo string
		FromFactoryName string
		ToBranchName    string
		ProductName     string
		Quantity        int
		Unit            string
		Status          string
		CreatedAt       time.Time
		TargetDate      *time.Time
	}
	var inbRows []inbRow
	h.db.Table("product_transfers pt").
		Select(`
			pt.id,
			COALESCE(pt.delivery_order_no, '-') as delivery_order_no,
			COALESCE(f.name, 'Pabrik') as from_factory_name,
			COALESCE(b.name, '-') as to_branch_name,
			COALESCE(p.name, 'Produk') as product_name,
			pt.quantity,
			COALESCE(pt.unit, 'PCS') as unit,
			pt.status,
			pt.created_at,
			pt.target_shipment_date as target_date
		`).
		Joins("LEFT JOIN factories f ON pt.from_factory_id = f.id").
		Joins("LEFT JOIN branches b ON pt.to_branch_id = b.id").
		Joins("LEFT JOIN products p ON pt.product_id = p.id").
		Where("pt.status IN ('LOADING', 'SHIPPED', 'ARRIVED', 'REQUESTED')").
		Where(func() *gorm.DB {
			if branchID != uuid.Nil {
				return h.db.Where("pt.to_branch_id = ?", branchID)
			}
			return h.db
		}()).
		Order("pt.created_at DESC").
		Limit(10).
		Scan(&inbRows)

	for _, ir := range inbRows {
		targetStr := "-"
		if ir.TargetDate != nil {
			targetStr = ir.TargetDate.Format("02 Jan 2006")
		}
		resp.PendingInboundShipments = append(resp.PendingInboundShipments, WarehousePendingShipmentItem{
			ID:              ir.ID,
			DeliveryOrderNo: ir.DeliveryOrderNo,
			FromFactory:     ir.FromFactoryName,
			ToBranch:        ir.ToBranchName,
			ProductName:     ir.ProductName,
			Quantity:        ir.Quantity,
			Unit:            ir.Unit,
			Status:          ir.Status,
			CreatedAt:       ir.CreatedAt.Format("02 Jan 2006 15:04"),
			TargetDate:      targetStr,
		})
	}

	// 3. Pending Sales Transfers Count
	var pendingSalesTransfersCount int64
	stQ := h.db.Table("sales_transfers st").
		Where("st.status = 'PENDING'")
	if branchID != uuid.Nil {
		stQ = stQ.Where("st.branch_id = ?", branchID)
	}
	stQ.Count(&pendingSalesTransfersCount)
	resp.Summary.PendingSalesTransfersCount = int(pendingSalesTransfersCount)

	// 4. Today's IN and OUT Mutation Sum
	type todayLogStat struct {
		InQty  int
		OutQty int
	}
	var tlStat todayLogStat
	todayLogQ := h.db.Table("warehouse_logs wl").
		Select(`
			COALESCE(SUM(CASE WHEN wl.type = 'IN' THEN wl.quantity ELSE 0 END), 0) as in_qty,
			COALESCE(SUM(CASE WHEN wl.type = 'OUT' THEN wl.quantity ELSE 0 END), 0) as out_qty
		`).
		Where("DATE(wl.created_at AT TIME ZONE 'Asia/Jakarta') = ?", todayStr)
	if branchID != uuid.Nil {
		todayLogQ = todayLogQ.Where("wl.branch_id = ?", branchID)
	}
	todayLogQ.Scan(&tlStat)
	resp.Summary.TodayInQuantity = tlStat.InQty
	resp.Summary.TodayOutQuantity = tlStat.OutQty

	// 5. Mutation Trend Calculation (Inbound vs Outbound)
	daysToFetch := 7
	if rangeParam == "30d" || rangeParam == "month" {
		daysToFetch = 30
	} else if rangeParam == "year" {
		daysToFetch = 365
	}

	if rangeParam == "year" {
		// Monthly aggregation for 12 months
		type monthlyLogPoint struct {
			MonthYear string
			MonthKey  string
			InQty     int
			OutQty    int
		}
		var mLogs []monthlyLogPoint
		h.db.Raw(`
			SELECT 
				TO_CHAR(DATE_TRUNC('month', wl.created_at AT TIME ZONE 'Asia/Jakarta'), 'Mon YYYY') AS month_year,
				TO_CHAR(DATE_TRUNC('month', wl.created_at AT TIME ZONE 'Asia/Jakarta'), 'YYYY-MM') AS month_key,
				COALESCE(SUM(CASE WHEN wl.type = 'IN' THEN wl.quantity ELSE 0 END), 0) AS in_qty,
				COALESCE(SUM(CASE WHEN wl.type = 'OUT' THEN wl.quantity ELSE 0 END), 0) AS out_qty
			FROM warehouse_logs wl
			WHERE wl.created_at >= (NOW() AT TIME ZONE 'Asia/Jakarta') - INTERVAL '12 months'
			  `+warehouseBranchFilter(branchID)+`
			GROUP BY DATE_TRUNC('month', wl.created_at AT TIME ZONE 'Asia/Jakarta')
			ORDER BY DATE_TRUNC('month', wl.created_at AT TIME ZONE 'Asia/Jakarta') ASC
		`).Scan(&mLogs)

		for _, ml := range mLogs {
			resp.MutationTrend = append(resp.MutationTrend, WarehouseMutationTrendPoint{
				Date:      ml.MonthKey,
				Label:     ml.MonthYear,
				InQty:     ml.InQty,
				OutQty:    ml.OutQty,
				NetChange: ml.InQty - ml.OutQty,
			})
		}
	} else {
		// Daily aggregation
		startDate := now.AddDate(0, 0, -(daysToFetch - 1))
		startDateStr := startDate.Format("2006-01-02")

		type dailyLogPoint struct {
			LogDate string
			InQty   int
			OutQty  int
		}
		var dLogs []dailyLogPoint
		h.db.Raw(`
			SELECT 
				TO_CHAR(wl.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS log_date,
				COALESCE(SUM(CASE WHEN wl.type = 'IN' THEN wl.quantity ELSE 0 END), 0) AS in_qty,
				COALESCE(SUM(CASE WHEN wl.type = 'OUT' THEN wl.quantity ELSE 0 END), 0) AS out_qty
			FROM warehouse_logs wl
			WHERE DATE(wl.created_at AT TIME ZONE 'Asia/Jakarta') >= ?
			  `+warehouseBranchFilter(branchID)+`
			GROUP BY TO_CHAR(wl.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD')
			ORDER BY log_date ASC
		`, startDateStr).Scan(&dLogs)

		dailyMap := make(map[string]dailyLogPoint)
		for _, dl := range dLogs {
			dailyMap[dl.LogDate] = dl
		}

		dayNamesIndo := map[time.Weekday]string{
			time.Sunday:    "Min",
			time.Monday:    "Sen",
			time.Tuesday:   "Sel",
			time.Wednesday: "Rab",
			time.Thursday:  "Kam",
			time.Friday:    "Jum",
			time.Saturday:  "Sab",
		}
		monthNamesIndo := map[time.Month]string{
			time.January:   "Jan",
			time.February:  "Feb",
			time.March:     "Mar",
			time.April:     "Apr",
			time.May:       "Mei",
			time.June:      "Jun",
			time.July:      "Jul",
			time.August:    "Agu",
			time.September: "Sep",
			time.October:   "Okt",
			time.November:  "Nov",
			time.December:  "Des",
		}

		for i := 0; i < daysToFetch; i++ {
			curD := startDate.AddDate(0, 0, i)
			dKey := curD.Format("2006-01-02")
			label := fmt.Sprintf("%s, %02d %s", dayNamesIndo[curD.Weekday()], curD.Day(), monthNamesIndo[curD.Month()])

			dl := dailyMap[dKey]
			resp.MutationTrend = append(resp.MutationTrend, WarehouseMutationTrendPoint{
				Date:      dKey,
				Label:     label,
				InQty:     dl.InQty,
				OutQty:    dl.OutQty,
				NetChange: dl.InQty - dl.OutQty,
			})
		}
	}

	// 6. Recent Logs Stream (Top 10)
	type recLogRow struct {
		ID          string
		ProductName string
		SKU         string
		Type        string
		Quantity    int
		Source      string
		ProcessedBy string
		ApprovedBy  string
		CreatedAt   time.Time
	}
	var recRows []recLogRow
	h.db.Table("warehouse_logs wl").
		Select(`
			wl.id,
			COALESCE(p.name, '-') as product_name,
			COALESCE(p.sku, '-') as sku,
			wl.type,
			wl.quantity,
			COALESCE(wl.source, '-') as source,
			COALESCE(wl.processed_by, '-') as processed_by,
			COALESCE(wl.approved_by, '-') as approved_by,
			wl.created_at
		`).
		Joins("LEFT JOIN products p ON wl.product_id = p.id").
		Where(func() *gorm.DB {
			if branchID != uuid.Nil {
				return h.db.Where("wl.branch_id = ?", branchID)
			}
			return h.db
		}()).
		Order("wl.created_at DESC").
		Limit(10).
		Scan(&recRows)

	for _, rr := range recRows {
		diffH := int(now.Sub(rr.CreatedAt).Hours())
		timeAgo := "Baru saja"
		if diffH >= 24 {
			timeAgo = fmt.Sprintf("%d hari lalu", diffH/24)
		} else if diffH >= 1 {
			timeAgo = fmt.Sprintf("%dj lalu", diffH)
		}

		resp.RecentLogs = append(resp.RecentLogs, WarehouseRecentLogItem{
			ID:          rr.ID,
			ProductName: rr.ProductName,
			SKU:         rr.SKU,
			Type:        rr.Type,
			Quantity:    rr.Quantity,
			Source:      rr.Source,
			ProcessedBy: rr.ProcessedBy,
			ApprovedBy:  rr.ApprovedBy,
			CreatedAt:   rr.CreatedAt.In(loc).Format("02 Jan 2006 15:04"),
			TimeAgo:     timeAgo,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}

func warehouseBranchFilter(branchID uuid.UUID) string {
	if branchID != uuid.Nil {
		return fmt.Sprintf("AND wl.branch_id = '%s'", branchID.String())
	}
	return ""
}

