package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"tiktokshop/open/sdk_golang/apis"
	finance_v202309 "tiktokshop/open/sdk_golang/models/finance/v202309"
	finance_v202501 "tiktokshop/open/sdk_golang/models/finance/v202501"
	order_v202309 "tiktokshop/open/sdk_golang/models/order/v202309"
	product_v202309 "tiktokshop/open/sdk_golang/models/product/v202309"
)

// Global broker for marketplace real-time sync
var MarketplaceBroker = NewSSEBroker()

type MarketplaceHandler struct {
	usecase          usecase.MarketplaceUsecase
	db               *gorm.DB
	tikTokAppKey     string
	tikTokAppSecret  string
	shopeePartnerID  string
	shopeePartnerKey string
	financeSyncMu    sync.RWMutex
	financeSyncJobs  map[string]financeSyncStatus
	financeRequestMu sync.Mutex
	financeNextCall  time.Time
	unsettledMu      sync.Mutex
	unsettledReady   bool
	unsettledJobs    map[string]unsettledJob
}

type financeSyncStatus struct {
	Status                 string     `json:"status"`
	SyncedCount            int        `json:"synced_count"`
	StatementsCount        int        `json:"statements_count"`
	PaidStatementsCount    int        `json:"paid_statements_count"`
	SettledStatementsCount int        `json:"settled_statements_count"`
	LatestPaymentAt        *time.Time `json:"latest_payment_at,omitempty"`
	ProcessedCount         int        `json:"processed_count"`
	Phase                  string     `json:"phase,omitempty"`
	Platform               string     `json:"platform,omitempty"`
	Error                  string     `json:"error,omitempty"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func NewMarketplaceHandler(u usecase.MarketplaceUsecase, db *gorm.DB) *MarketplaceHandler {
	appKey := os.Getenv("TIKTOK_APP_KEY")
	if appKey == "" {
		appKey = "6kmomdj5ihm2p"
	}
	appSecret := os.Getenv("TIKTOK_APP_SECRET")
	if appSecret == "" {
		appSecret = "bc5e00b572660ec28c3e874e47b987c95de045a0"
	}
	return &MarketplaceHandler{
		usecase:          u,
		db:               db,
		tikTokAppKey:     appKey,
		tikTokAppSecret:  appSecret,
		shopeePartnerID:  os.Getenv("SHOPEE_PARTNER_ID"),
		shopeePartnerKey: os.Getenv("SHOPEE_PARTNER_KEY"),
		financeSyncJobs:  make(map[string]financeSyncStatus),
		unsettledJobs:    make(map[string]unsettledJob),
	}
}

func (h *MarketplaceHandler) RegisterRoutes(r *gin.RouterGroup) {
	marketplace := r.Group("/marketplace")
	marketplace.Use(RoleMiddleware("superadmin", "manager", "hr_admin", "admin_gudang", "supervisor_sales", "admin_marketplace"))
	{
		marketplace.GET("/dashboard", h.GetDashboard)
		marketplace.GET("/profitability", h.GetProfitability)
		marketplace.POST("/profitability/expenses", h.CreateProfitabilityExpense)
		marketplace.DELETE("/profitability/expenses/:id", h.DeleteProfitabilityExpense)
		marketplace.GET("/factory-products", h.GetFactoryProducts)
		marketplace.GET("/stocks", h.ListStocks)
		marketplace.GET("/stocks/:id", h.GetStock)
		marketplace.POST("/stocks", h.CreateStock)
		marketplace.PUT("/stocks/:id", h.UpdateStock)
		marketplace.DELETE("/stocks/:id", h.DeleteStock)
		marketplace.POST("/stocks/:id/opname", h.DoOpname)
		marketplace.GET("/opnames", h.ListOpnames)
		marketplace.PUT("/opnames/:id", h.UpdateOpname)
		marketplace.GET("/audit-logs", h.GetAuditLogs)
		// Marketplace order & product management
		marketplace.GET("/orders", h.ListMarketplaceOrders)
		marketplace.GET("/orders/filters", h.GetOrderFilters)
		marketplace.POST("/sync-orders", h.SyncMarketplaceOrders)
		marketplace.POST("/orders/fetch-single", h.FetchSingleOrder)
		marketplace.GET("/stream", h.StreamOrderUpdates)
		marketplace.GET("/mp-products", h.ListMarketplaceProducts)
		marketplace.POST("/sync-products", h.SyncMarketplaceProducts)
		marketplace.POST("/pack/:order_id", h.PackOrder)
		marketplace.POST("/verify/:order_id", h.VerifyOrder)
		marketplace.GET("/scan-history", h.GetScanHistory)
		marketplace.GET("/scan-history/warehouses", h.GetScanHistoryWarehouses)
		marketplace.GET("/mappings", h.GetProductMappings)
		marketplace.POST("/mappings", h.SaveProductMappings)
		// Marketplace settlement & finance
		marketplace.GET("/settlements", h.ListSettlements)
		marketplace.GET("/settlements/summary", h.GetSettlementSummary)
		marketplace.GET("/settlements/unsettled-snapshots", h.ListUnsettledSnapshots)
		marketplace.GET("/settlements/unsettled-snapshots/:id", h.GetUnsettledSnapshot)
		marketplace.POST("/settlements/unsettled-snapshots", h.CaptureUnsettledSnapshot)
		marketplace.GET("/settlements/unsettled-jobs/:id", h.GetUnsettledJob)
		marketplace.POST("/settlements/restock", h.RestockOrder)
		marketplace.POST("/settlements/import", h.ImportSettlements)
		marketplace.PUT("/settlements/:order_id", h.UpdateSettlementOrder)
		marketplace.POST("/settlements/auto-classify", h.AutoClassifySettlements)
		marketplace.POST("/settlements/sync-finance", h.SyncTikTokFinance)
		marketplace.POST("/settlements/sync-shopee-finance", h.SyncShopeeFinance)
		marketplace.GET("/settlements/sync-finance/:sync_id", h.GetTikTokFinanceSyncStatus)
		marketplace.POST("/settlements/sync-order-finance/:order_id", h.SyncTikTokOrderFinance)
		marketplace.POST("/settlements/sync-shopee-order-finance/:order_id", h.SyncShopeeOrderFinance)
	}
}

func (h *MarketplaceHandler) GetFactoryProducts(c *gin.Context) {
	companyID := uuid.Nil
	if val, ok := c.Get("companyID"); ok {
		companyID = val.(uuid.UUID)
	}

	products, err := h.usecase.GetFactoryProducts(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

func getActorInfo(c *gin.Context) (uuid.UUID, string) {
	var actorID uuid.UUID
	if val, ok := c.Get("user_id"); ok {
		actorID, _ = uuid.Parse(val.(string))
	}
	actorName := "Admin"
	if val, ok := c.Get("username"); ok && val != "" {
		actorName = val.(string)
	} else if val, ok := c.Get("name"); ok && val != "" {
		actorName = val.(string)
	}
	return actorID, actorName
}

func (h *MarketplaceHandler) GetDashboard(c *gin.Context) {
	branchID := parseBranchID(c)
	period := c.Query("period")
	if period == "" {
		period = "daily"
	}
	dateFilter := c.Query("date")
	stats, err := h.usecase.GetDashboardStats(branchID, period, dateFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *MarketplaceHandler) ListStocks(c *gin.Context) {
	branchID := parseBranchID(c)
	stocks, err := h.usecase.GetAllStocks(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stocks)
}

func (h *MarketplaceHandler) GetStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	stock, err := h.usecase.GetStockByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, stock)
}

func (h *MarketplaceHandler) handleImageUpload(c *gin.Context) (string, error) {
	file, err := c.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}
	// Limit 5MB
	if file.Size > 5*1024*1024 {
		return "", fmt.Errorf("ukuran file maksimal 5MB")
	}
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	uploadDir := "uploads/marketplace"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}
	savePath := filepath.Join(uploadDir, filename)
	if err := utils.CompressAndSaveImage(file, savePath, 1000); err != nil {
		return "", err
	}
	return "/uploads/marketplace/" + filename, nil
}

func (h *MarketplaceHandler) CreateStock(c *gin.Context) {
	actorID, actorName := getActorInfo(c)
	stock := &domain.MarketplaceStock{
		BranchID:  parseBranchID(c),
		CreatedBy: actorName,
	}

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		stock.ProductName = c.PostForm("product_name")
		stock.SKU = c.PostForm("sku")
		if q, err := strconv.Atoi(c.PostForm("quantity")); err == nil {
			stock.Quantity = q
		}
		if ms, err := strconv.Atoi(c.PostForm("min_stock")); err == nil {
			stock.MinStock = ms
		}
		if p, err := strconv.ParseFloat(c.PostForm("price"), 64); err == nil {
			stock.Price = p
		}
		stock.Notes = c.PostForm("notes")

		pidStr := c.PostForm("product_id")
		if pidStr != "" {
			pid, err := uuid.Parse(pidStr)
			if err == nil {
				stock.ProductID = &pid
			}
		}

		imgUrl, err := h.handleImageUpload(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		stock.ImageURL = imgUrl
	} else {
		var req struct {
			ProductID   *string `json:"product_id"` // optional: link ke produk pabrik
			ProductName string  `json:"product_name"`
			SKU         string  `json:"sku"`
			Quantity    int     `json:"quantity"`
			MinStock    int     `json:"min_stock"`
			Price       float64 `json:"price"`
			Notes       string  `json:"notes"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		stock.ProductName = req.ProductName
		stock.SKU = req.SKU
		stock.Quantity = req.Quantity
		stock.MinStock = req.MinStock
		stock.Price = req.Price
		stock.Notes = req.Notes
		if req.ProductID != nil && *req.ProductID != "" {
			pid, err := uuid.Parse(*req.ProductID)
			if err == nil {
				stock.ProductID = &pid
			}
		}
	}

	if err := h.usecase.CreateStock(stock, actorID, actorName, c.ClientIP()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, stock)
}

func (h *MarketplaceHandler) UpdateStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var stock domain.MarketplaceStock
	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		stock.ProductName = c.PostForm("product_name")
		stock.SKU = c.PostForm("sku")
		if ms, err := strconv.Atoi(c.PostForm("min_stock")); err == nil {
			stock.MinStock = ms
		}
		if p, err := strconv.ParseFloat(c.PostForm("price"), 64); err == nil {
			stock.Price = p
		}
		stock.Notes = c.PostForm("notes")

		pidStr := c.PostForm("product_id")
		if pidStr != "" {
			pid, err := uuid.Parse(pidStr)
			if err == nil {
				stock.ProductID = &pid
			}
		}

		imgUrl, err := h.handleImageUpload(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if imgUrl != "" {
			stock.ImageURL = imgUrl
		}
	} else {
		var req struct {
			ProductID   *string `json:"product_id"`
			ProductName string  `json:"product_name"`
			SKU         string  `json:"sku"`
			MinStock    int     `json:"min_stock"`
			Price       float64 `json:"price"`
			Notes       string  `json:"notes"`
			ImageURL    string  `json:"image_url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		stock.ProductName = req.ProductName
		stock.SKU = req.SKU
		stock.MinStock = req.MinStock
		stock.Price = req.Price
		stock.Notes = req.Notes
		stock.ImageURL = req.ImageURL
		if req.ProductID != nil && *req.ProductID != "" {
			pid, err := uuid.Parse(*req.ProductID)
			if err == nil {
				stock.ProductID = &pid
			}
		}
	}

	actorID, actorName := getActorInfo(c)
	if err := h.usecase.UpdateStock(id, &stock, actorID, actorName, c.ClientIP()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil diperbarui"})
}

func (h *MarketplaceHandler) DeleteStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	actorID, actorName := getActorInfo(c)
	if err := h.usecase.DeleteStock(id, actorID, actorName, c.ClientIP()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil dihapus"})
}

func (h *MarketplaceHandler) DoOpname(c *gin.Context) {
	stockID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	var req struct {
		Type       domain.MarketplaceOpnameType `json:"type" binding:"required"`
		Quantity   *int                         `json:"quantity" binding:"required"`
		Reference  string                       `json:"reference"`
		Notes      string                       `json:"notes"`
		Date       *string                      `json:"date"`
		OpnameDate *string                      `json:"opname_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var opnameDate *time.Time
	dateStr := ""
	if req.Date != nil && strings.TrimSpace(*req.Date) != "" {
		dateStr = strings.TrimSpace(*req.Date)
	} else if req.OpnameDate != nil && strings.TrimSpace(*req.OpnameDate) != "" {
		dateStr = strings.TrimSpace(*req.OpnameDate)
	}

	if dateStr != "" {
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04",
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"2006-01-02",
		}
		for _, f := range formats {
			if parsed, err := time.ParseInLocation(f, dateStr, time.Local); err == nil {
				opnameDate = &parsed
				break
			}
		}
	}

	actorID, actorName := getActorInfo(c)
	opname, err := h.usecase.DoOpname(stockID, req.Type, *req.Quantity, req.Reference, req.Notes, opnameDate, actorID, actorName, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, opname)
}

func (h *MarketplaceHandler) UpdateOpname(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pergerakan stok tidak valid"})
		return
	}

	var req struct {
		Date      *string `json:"date"`
		Quantity  *int    `json:"quantity"`
		Reference *string `json:"reference"`
		Notes     *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID, actorName := getActorInfo(c)
	payload := usecase.UpdateOpnamePayload{
		Date:      req.Date,
		Quantity:  req.Quantity,
		Reference: req.Reference,
		Notes:     req.Notes,
	}

	opname, err := h.usecase.UpdateOpname(id, payload, actorID, actorName, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data pergerakan stok berhasil diperbarui",
		"data":    opname,
	})
}

func (h *MarketplaceHandler) ListOpnames(c *gin.Context) {
	branchID := parseBranchID(c)
	opnameType := c.Query("type")

	var stockID *uuid.UUID
	if sid := c.Query("stock_id"); sid != "" {
		if parsed, err := uuid.Parse(sid); err == nil {
			stockID = &parsed
		}
	}

	var from, to *time.Time
	if f := c.Query("from"); f != "" {
		t, err := time.Parse("2006-01-02", f)
		if err == nil {
			from = &t
		}
	}
	if t := c.Query("to"); t != "" {
		if parsed, err := time.Parse("2006-01-02", t); err == nil {
			endOfDay := parsed.Add(24 * time.Hour).Add(-time.Second)
			to = &endOfDay
		}
	}

	search := c.Query("search")
	sourceType := strings.ToUpper(strings.TrimSpace(c.Query("source_type")))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	opnames, total, summary, err := h.usecase.GetOpnames(branchID, stockID, opnameType, from, to, search, limit, offset, sourceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    opnames,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"summary": summary,
	})
}

func (h *MarketplaceHandler) GetAuditLogs(c *gin.Context) {
	entity := c.Query("entity")
	if entity == "" {
		entity = "MarketplaceStock"
	}
	logs, err := h.usecase.GetAuditLogs(entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *MarketplaceHandler) GetOrderFilters(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	filters, err := h.usecase.GetOrderFilters(branchIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, filters)
}

// ListMarketplaceOrders returns paginated marketplace orders with optional filters.
func (h *MarketplaceHandler) ListMarketplaceOrders(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	filter := repository.MarketplaceOrderFilter{
		BranchID:         branchIDPtr,
		Marketplace:      c.Query("marketplace"),
		Status:           c.Query("status"),
		Keyword:          c.Query("keyword"),
		ProductName:      c.Query("product_name"),
		WarehouseName:    c.Query("warehouse_name"),
		ShippingProvider: c.Query("shipping_provider"),
		Page:             1,
		PageSize:         20,
	}

	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			filter.Page = n
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			filter.PageSize = n
		}
	}
	if f := c.Query("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			filter.From = &t
		}
	}
	if t := c.Query("to"); t != "" {
		if parsed, err := time.Parse("2006-01-02", t); err == nil {
			endOfDay := parsed.Add(24 * time.Hour).Add(-time.Second)
			filter.To = &endOfDay
		}
	}

	orders, total, totalOmzet, totalTiktok, totalShopee, err := h.usecase.GetMarketplaceOrders(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stats, err := h.usecase.GetMarketplaceOrderStatusStats(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":         orders,
		"total":        total,
		"total_omzet":  totalOmzet,
		"total_tiktok": totalTiktok,
		"total_shopee": totalShopee,
		"status_stats": stats,
		"page":         filter.Page,
		"page_size":    filter.PageSize,
	})
}

// resolveBranchIDByWarehouse maps a marketplace warehouse to a branch in this installation.
func resolveBranchIDByWarehouse(db *gorm.DB, whName string, fallbackBranchID *uuid.UUID) *uuid.UUID {
	whLower := strings.ToLower(strings.TrimSpace(whName))
	// An authenticated marketplace account's branch takes precedence.
	if fallbackBranchID != nil && *fallbackBranchID != uuid.Nil {
		var tokenBranch domain.Branch
		if err := db.First(&tokenBranch, "id = ?", *fallbackBranchID).Error; err == nil {
			return fallbackBranchID
		}
	}
	if whLower == "" {
		return nil
	}
	// A name match must be exact to avoid assigning another client's warehouse.
	var branches []domain.Branch
	if err := db.Find(&branches).Error; err == nil {
		for _, b := range branches {
			if whLower == strings.ToLower(strings.TrimSpace(b.Name)) {
				return &b.ID
			}
		}
	}

	return fallbackBranchID
}

// cleanShippingProvider normalizes shipping provider strings, stripping generic delivery tiers
// and inferring real carriers from tracking numbers if necessary.
func cleanShippingProvider(rawProvider string, trackingID string) string {
	raw := strings.TrimSpace(rawProvider)
	rawLower := strings.ToLower(raw)
	track := strings.ToUpper(strings.TrimSpace(trackingID))

	// If raw provider already explicitly specifies known carriers
	if strings.Contains(rawLower, "cargo") || strings.Contains(rawLower, "kargo") {
		return "J&T Cargo"
	}
	if strings.Contains(rawLower, "j&t") || strings.Contains(rawLower, "jnt") || strings.Contains(rawLower, "express") {
		return "J&T Express"
	}
	if strings.Contains(rawLower, "ninja") {
		return "Ninja Van"
	}
	if strings.Contains(rawLower, "sicepat") {
		return "SiCepat"
	}
	if strings.Contains(rawLower, "jne") {
		return "JNE"
	}
	if strings.Contains(rawLower, "spx") || strings.Contains(rawLower, "shopee express") {
		return "SPX Express"
	}
	if strings.Contains(rawLower, "anteraja") {
		return "Anteraja"
	}
	if strings.Contains(rawLower, "wahana") {
		return "Wahana"
	}
	if strings.Contains(rawLower, "pos") {
		return "POS Indonesia"
	}

	// Check if raw matches generic buyer delivery tier noise
	isGenericTier := rawLower == "" ||
		rawLower == "-" ||
		strings.Contains(rawLower, "standard") ||
		strings.Contains(rawLower, "economical") ||
		strings.Contains(rawLower, "hemat") ||
		strings.Contains(rawLower, "nextday") ||
		strings.Contains(rawLower, "next day") ||
		strings.Contains(rawLower, "same day") ||
		strings.Contains(rawLower, "instant")

	if isGenericTier {
		// Infer from tracking ID prefix if available
		if strings.HasPrefix(track, "JX") || strings.HasPrefix(track, "JP") || strings.HasPrefix(track, "JD") || strings.HasPrefix(track, "JT") || strings.HasPrefix(track, "JY") {
			return "J&T Express"
		}
		if strings.HasPrefix(track, "JC") || (len(track) > 0 && track[0] == '3') {
			return "J&T Cargo"
		}
		if strings.HasPrefix(track, "SPX") || strings.HasPrefix(track, "ID") {
			return "SPX Express"
		}
		return "-"
	}

	return raw
}

// backfillMarketplaceOrderBranches updates existing orders whose branch_id is mismatched with its warehouse_name
func backfillMarketplaceOrderBranches(db *gorm.DB) {
	// Backfill only unassigned orders whose warehouse exactly matches a branch name.
	db.Exec(`UPDATE marketplace_orders AS mo SET branch_id = b.id
		FROM branches AS b WHERE mo.branch_id IS NULL
		AND LOWER(TRIM(mo.warehouse_name)) = LOWER(TRIM(b.name))`)

	// Backfill true creation timestamps for all historical marketplace orders
	db.Exec("UPDATE marketplace_orders SET created_at = to_timestamp(create_time_raw) WHERE create_time_raw > 0 AND created_at IS DISTINCT FROM to_timestamp(create_time_raw)")

	// Clean generic delivery tiers from shipping_provider
	db.Exec(`UPDATE marketplace_orders 
		SET shipping_provider = CASE 
			WHEN tracking_id ILIKE 'JX%' OR tracking_id ILIKE 'JP%' OR tracking_id ILIKE 'JD%' OR tracking_id ILIKE 'JT%' THEN 'J&T Express'
			WHEN tracking_id ILIKE 'JC%' OR tracking_id LIKE '3%' THEN 'J&T Cargo'
			WHEN tracking_id ILIKE 'SPX%' THEN 'SPX Express'
			ELSE '-'
		END
		WHERE LOWER(TRIM(shipping_provider)) IN ('standard shipping', 'standard delivery', 'economical', 'nextday delivery', 'standard', 'same day', 'instant', 'hemat', 'next day')
		   OR shipping_provider IS NULL
		   OR shipping_provider = ''`)

	// Update active Shopee orders to VERIFIED
	db.Exec("UPDATE marketplace_orders SET order_status = 'VERIFIED' WHERE marketplace = 'shopee' AND order_status IN ('AWAITING_COLLECTION', 'AWAITING_SHIPMENT', 'PROCESSED', 'READY_TO_SHIP')")
}

// SyncMarketplaceOrders fetches orders from TikTok and Shopee and upserts them into the DB.
func (h *MarketplaceHandler) SyncMarketplaceOrders(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	targetMp := strings.ToLower(c.Query("marketplace"))
	totalSynced := 0
	var syncErrors []string
	var mu sync.Mutex

	// Auto-correct any legacy orders with mismatched branches based on warehouse_name
	backfillMarketplaceOrderBranches(h.db)

	var wg sync.WaitGroup

	// Sync TikTok
	if targetMp == "" || targetMp == "all" || targetMp == "tiktok" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count, err := h.syncTikTokOrders(branchIDPtr)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("TikTok: %v", err))
			} else {
				totalSynced += count
			}
		}()
	}

	// Sync Shopee
	if targetMp == "" || targetMp == "all" || targetMp == "shopee" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count, err := h.syncShopeeOrders(branchIDPtr)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("Shopee: %v", err))
			} else {
				totalSynced += count
			}
		}()
	}

	wg.Wait()

	// Clean up noisy unverified cancelled orders and restore all packed orders that have scan histories
	_ = h.usecase.CleanNoiseAndRestoreScannedOrders(branchIDPtr)

	// STRICT ISOLATION: Penarikan order WMS gudang berdiri sendiri.
	// Auto-klasifikasi settlement dilepas dari sync order agar status keuangan tidak ditebak secara sepihak saat staf menarik pesanan baru.

	if len(syncErrors) > 0 && (targetMp == "tiktok" || targetMp == "shopee") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        strings.Join(syncErrors, "; "),
			"synced_count": totalSynced,
		})
		return
	}

	msg := fmt.Sprintf("Sync pesanan selesai (%d pesanan tersinkron)", totalSynced)
	if len(syncErrors) > 0 {
		msg += fmt.Sprintf(" (Catatan: %s)", strings.Join(syncErrors, ", "))
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      msg,
		"synced_count": totalSynced,
	})
}

func (h *MarketplaceHandler) syncTikTokOrders(branchIDPtr *uuid.UUID) (int, error) {
	tokenRecord, err := getValidTikTokToken(h.db, h.tikTokAppKey, h.tikTokAppSecret, branchIDPtr)
	if err != nil {
		return 0, err
	}

	cfg := apis.NewConfiguration()
	cfg.AddAppInfo(h.tikTokAppKey, h.tikTokAppSecret)
	client := apis.NewAPIClient(cfg)

	warehouseMap := make(map[string]string)
	whResp, _, errWh := client.LogisticsV202309API.Logistics202309WarehousesGet(context.Background()).
		ShopCipher(tokenRecord.ShopID).
		XTtsAccessToken(tokenRecord.AccessToken).
		ContentType("application/json").
		Execute()

	if errWh == nil && whResp != nil && whResp.Data != nil && whResp.Data.Warehouses != nil {
		for _, wh := range whResp.Data.Warehouses {
			if id, ok := wh.GetIdOk(); ok && id != nil {
				if name, ok := wh.GetNameOk(); ok && name != nil {
					warehouseMap[*id] = *name
				}
			}
		}
	}

	updateTimeGe := time.Now().AddDate(0, 0, -30).Unix()
	var allOrders []order_v202309.Order202309GetOrderListResponseDataOrders
	statusesToFetch := []string{"AWAITING_COLLECTION"}

	for _, targetStatus := range statusesToFetch {
		var pageCursor string
		maxPages := 100
		pageCount := 0

		for {
			pageCount++
			if pageCount > maxPages {
				break
			}

			currentStatus := targetStatus
			reqBody := order_v202309.Order202309GetOrderListRequestBody{
				OrderStatus:  &currentStatus,
				UpdateTimeGe: &updateTimeGe,
			}

			reqBuilder := client.OrderV202309API.Order202309OrdersSearchPost(context.Background()).
				ShopCipher(tokenRecord.ShopID).
				XTtsAccessToken(tokenRecord.AccessToken).
				PageSize(100).
				SortField("update_time").
				SortOrder("DESC").
				ContentType("application/json")

			if pageCursor != "" {
				reqBuilder = reqBuilder.PageToken(pageCursor)
			}

			resp, httpRes, err := reqBuilder.Order202309GetOrderListRequestBody(reqBody).Execute()
			if err != nil {
				bodyStr := ""
				if httpRes != nil && httpRes.Body != nil {
					bodyBytes, _ := io.ReadAll(httpRes.Body)
					bodyStr = string(bodyBytes)
				}
				return 0, fmt.Errorf("tiktok api error: %v (%s)", err, bodyStr)
			}

			if resp != nil && resp.Data != nil && resp.Data.Orders != nil {
				allOrders = append(allOrders, resp.Data.Orders...)
			}

			if resp == nil || resp.Data == nil || resp.Data.GetNextPageToken() == "" {
				break
			}
			pageCursor = resp.Data.GetNextPageToken()
		}
	}

	// Deduplicate orders by order ID, keeping the latest update_time
	uniqueOrders := make(map[string]order_v202309.Order202309GetOrderListResponseDataOrders)
	for _, o := range allOrders {
		oid := o.GetId()
		if oid == "" {
			continue
		}
		if existing, ok := uniqueOrders[oid]; ok {
			if o.GetUpdateTime() >= existing.GetUpdateTime() {
				uniqueOrders[oid] = o
			}
		} else {
			uniqueOrders[oid] = o
		}
	}

	syncedCount := 0
	for _, tkOrder := range uniqueOrders {
		orderID := tkOrder.GetId()
		if orderID == "" {
			continue
		}

		orderStatus := tkOrder.GetStatus()
		if orderStatus != "AWAITING_COLLECTION" {
			continue
		}
		orderStatus = "VERIFIED"

		var totalAmount float64
		currency := "IDR"
		paymentMethod := ""
		if tkOrder.HasPayment() {
			payment := tkOrder.GetPayment()
			totalAmountStr := payment.GetTotalAmount()
			totalAmount, _ = strconv.ParseFloat(totalAmountStr, 64)
			currency = payment.GetCurrency()
		}
		paymentMethod = tkOrder.GetPaymentMethodName()

		whID := tkOrder.GetWarehouseId()
		whName := whID
		if name, exists := warehouseMap[whID]; exists {
			whName = name
		}

		assignedBranchID := resolveBranchIDByWarehouse(h.db, whName, tokenRecord.BranchID)

		trackingID := tkOrder.GetTrackingNumber()
		shippingProvider := cleanShippingProvider(tkOrder.GetShippingProvider(), trackingID)
		if shippingProvider == "" || shippingProvider == "-" {
			shippingProvider = cleanShippingProvider(tkOrder.GetDeliveryOptionName(), trackingID)
		}

		var orderCreatedAt time.Time
		if createTime := tkOrder.GetCreateTime(); createTime > 0 {
			orderCreatedAt = time.Unix(createTime, 0)
		} else {
			orderCreatedAt = time.Now()
		}

		if tkOrder.LineItems != nil {
			itemCounts := make(map[string]int)
			firstItems := make(map[string]order_v202309.Order202309GetOrderListResponseDataOrdersLineItems)

			for _, tkItem := range tkOrder.LineItems {
				skuID := tkItem.GetSkuId()
				itemCounts[skuID]++
				if _, exists := firstItems[skuID]; !exists {
					firstItems[skuID] = tkItem
				}
			}

			now := time.Now()
			for skuID, count := range itemCounts {
				tkItem := firstItems[skuID]
				mpOrder := &domain.MarketplaceOrder{
					BranchID:         assignedBranchID,
					OrderID:          orderID,
					OrderStatus:      orderStatus,
					SkuID:            tkItem.GetSkuId(),
					LineItemID:       tkItem.GetId(),
					SellerSKU:        tkItem.GetSellerSku(),
					ProductName:      tkItem.GetProductName(),
					WarehouseName:    whName,
					Marketplace:      "tiktok",
					BuyerUsername:    tkOrder.GetBuyerNickname(),
					TotalAmount:      totalAmount,
					Currency:         currency,
					PaymentMethod:    paymentMethod,
					TrackingID:       trackingID,
					ShippingProvider: shippingProvider,
					CreateTimeRaw:    tkOrder.GetCreateTime(),
					CreatedAt:        orderCreatedAt,
					SyncedAt:         &now,
					ItemCount:        count,
				}
				if saveErr := h.usecase.SaveMarketplaceOrder(mpOrder); saveErr != nil {
					return syncedCount, fmt.Errorf("gagal menyimpan order TikTok %s: %w", orderID, saveErr)
				}
				syncedCount++
			}
		} else {
			now := time.Now()
			mpOrder := &domain.MarketplaceOrder{
				BranchID:         assignedBranchID,
				OrderID:          orderID,
				OrderStatus:      orderStatus,
				Marketplace:      "tiktok",
				WarehouseName:    whName,
				BuyerUsername:    tkOrder.GetBuyerNickname(),
				TotalAmount:      totalAmount,
				Currency:         currency,
				PaymentMethod:    paymentMethod,
				TrackingID:       trackingID,
				ShippingProvider: shippingProvider,
				CreateTimeRaw:    tkOrder.GetCreateTime(),
				CreatedAt:        orderCreatedAt,
				SyncedAt:         &now,
				ItemCount:        0,
			}
			if saveErr := h.usecase.SaveMarketplaceOrder(mpOrder); saveErr != nil {
				return syncedCount, fmt.Errorf("gagal menyimpan order TikTok %s: %w", orderID, saveErr)
			}
			syncedCount++
		}
	}

	// Recovery: isi ulang tracking_id untuk order di database yang resinya masih kosong
	h.recoverMissingTrackingIDs(client, tokenRecord, branchIDPtr)

	return syncedCount, nil
}

// recoverMissingTrackingIDs mengambil ulang nomor resi dari TikTok (Get Order Detail)
// untuk order di database yang tracking_id-nya kosong, lalu meng-update kolom tracking_id & shipping_provider saja.
func (h *MarketplaceHandler) recoverMissingTrackingIDs(client *apis.APIClient, tokenRecord *domain.MarketplaceToken, branchIDPtr *uuid.UUID) {
	if client == nil || tokenRecord == nil {
		return
	}

	var missingOrderIDs []string
	q := h.db.Model(&domain.MarketplaceOrder{}).
		Where("marketplace = 'tiktok'").
		Where("tracking_id IS NULL OR tracking_id = '' OR tracking_id = '-'").
		Where("order_status NOT IN ('CANCELED', 'CANCELLED')").
		Where("created_at >= ?", time.Now().AddDate(0, 0, -30))
	if branchIDPtr != nil && *branchIDPtr != uuid.Nil {
		q = q.Where("branch_id = ?", *branchIDPtr)
	}
	_ = q.Distinct("order_id").Limit(500).Pluck("order_id", &missingOrderIDs).Error

	if len(missingOrderIDs) == 0 {
		return
	}

	recovered := 0
	for i := 0; i < len(missingOrderIDs); i += 50 {
		end := i + 50
		if end > len(missingOrderIDs) {
			end = len(missingOrderIDs)
		}
		batch := missingOrderIDs[i:end]

		resp, _, err := client.OrderV202309API.Order202309OrdersGet(context.Background()).
			ShopCipher(tokenRecord.ShopID).
			XTtsAccessToken(tokenRecord.AccessToken).
			Ids(batch).
			ContentType("application/json").
			Execute()
		if err != nil || resp == nil || resp.Data == nil {
			log.Printf("[TIKTOK TRACKING RECOVERY] batch error: %v", err)
			continue
		}

		for _, ord := range resp.Data.Orders {
			trackNo := strings.TrimSpace(ord.GetTrackingNumber())
			if trackNo == "" || trackNo == "-" {
				continue
			}
			updates := map[string]interface{}{"tracking_id": trackNo}
			if provider := cleanShippingProvider(ord.GetShippingProvider(), trackNo); provider != "" && provider != "-" {
				updates["shipping_provider"] = provider
			}
			res := h.db.Model(&domain.MarketplaceOrder{}).
				Where("order_id = ?", ord.GetId()).
				Where("tracking_id IS NULL OR tracking_id = '' OR tracking_id = '-'").
				Updates(updates)
			if res.Error == nil && res.RowsAffected > 0 {
				recovered++
			}
		}
	}

	if recovered > 0 {
		log.Printf("[TIKTOK TRACKING RECOVERY] %d order berhasil dipulihkan tracking_id-nya", recovered)
	}
}

func (h *MarketplaceHandler) syncTikTokSettlements(client *apis.APIClient, tokenRecord *domain.MarketplaceToken, branchIDPtr *uuid.UUID) int {
	if client == nil || tokenRecord == nil {
		return 0
	}

	statementTimeGe := time.Now().AddDate(0, 0, -30).Unix()
	var sortField = "statement_time"
	var sortOrder = "DESC"
	var pageSize interface{} = 10

	syncedSettlements := 0
	processedOrderIDs := make(map[string]bool)

	// 1. Fetch recent TikTok Statements & Statement Transactions (v202501)
	stResp, httpRes, err := client.FinanceV202309API.Finance202309StatementsGet(context.Background()).
		ShopCipher(tokenRecord.ShopID).
		XTtsAccessToken(tokenRecord.AccessToken).
		SortField(sortField).
		SortOrder(sortOrder).
		StatementTimeGe(statementTimeGe).
		PageSize(pageSize).
		ContentType("application/json").
		Execute()

	if err == nil && stResp != nil && stResp.Data != nil && len(stResp.Data.Statements) > 0 {
		statementsToCheck := stResp.Data.Statements
		if len(statementsToCheck) > 5 {
			statementsToCheck = statementsToCheck[:5]
		}

		for _, stmt := range statementsToCheck {
			stmtID := stmt.GetId()
			if stmtID == "" {
				continue
			}

			var settleTime *time.Time
			if stmtTime := stmt.GetStatementTime(); stmtTime > 0 {
				t := time.Unix(stmtTime, 0)
				settleTime = &t
			}

			var stmtSortField = "order_create_time"
			var stmtPageSize interface{} = 100
			txResp, _, txErr := client.FinanceV202501API.Finance202501StatementsStatementIdStatementTransactionsGet(context.Background(), stmtID).
				ShopCipher(tokenRecord.ShopID).
				XTtsAccessToken(tokenRecord.AccessToken).
				SortField(stmtSortField).
				PageSize(stmtPageSize).
				ContentType("application/json").
				Execute()

			if txErr != nil || txResp == nil || txResp.Data == nil || len(txResp.Data.Transactions) == 0 {
				continue
			}

			for _, tx := range txResp.Data.Transactions {
				orderID := tx.GetOrderId()
				if orderID == "" {
					orderID = tx.GetAssociatedOrderId()
				}
				if orderID == "" {
					orderID = tx.GetAdjustmentOrderId()
				}
				if orderID == "" {
					continue
				}

				settlementAmountStr := tx.GetSettlementAmount()
				settlementAmount, _ := strconv.ParseFloat(settlementAmountStr, 64)

				feeTaxAmountStr := tx.GetFeeTaxAmount()
				feeTaxAmount, _ := strconv.ParseFloat(feeTaxAmountStr, 64)
				if feeTaxAmount < 0 {
					feeTaxAmount = -feeTaxAmount
				}

				revenueAmountStr := tx.GetRevenueAmount()
				revenueAmount, _ := strconv.ParseFloat(revenueAmountStr, 64)

				status := "SETTLED"
				updates := map[string]interface{}{
					"settlement_status": status,
					"settlement_amount": settlementAmount,
					"platform_fee":      feeTaxAmount,
				}
				if revenueAmount > 0 {
					updates["total_amount"] = revenueAmount
				}
				if settleTime != nil {
					updates["settlement_date"] = *settleTime
				}

				res := h.db.Model(&domain.MarketplaceOrder{}).Where("order_id = ?", orderID).Updates(updates)
				if res.Error == nil && res.RowsAffected > 0 {
					syncedSettlements++
					processedOrderIDs[orderID] = true
				}
			}
		}
	} else if err != nil {
		bodyStr := ""
		if httpRes != nil && httpRes.Body != nil {
			bodyBytes, _ := io.ReadAll(httpRes.Body)
			bodyStr = string(bodyBytes)
		}
		log.Printf("[TIKTOK FINANCE SYNC] Notice fetching statements: %v (%s)", err, bodyStr)
	}

	// 2. Direct Order-Level Transactions Lookup (v202501) for orders still missing fees / settlements
	// Offload to background goroutine so the synchronous HTTP response returns immediately without timing out (504)
	var pendingOrderIDs []string
	qPending := h.db.Model(&domain.MarketplaceOrder{}).
		Where("marketplace = 'tiktok'").
		Where("order_status IN ('COMPLETED', 'DELIVERED', 'VERIFIED', 'PACKED')").
		Where("platform_fee = 0 OR settlement_amount = 0").
		Where("created_at >= ?", time.Now().AddDate(0, 0, -30))
	if branchIDPtr != nil && *branchIDPtr != uuid.Nil {
		qPending = qPending.Where("branch_id = ?", *branchIDPtr)
	}
	_ = qPending.Distinct("order_id").Limit(50).Pluck("order_id", &pendingOrderIDs).Error

	if len(pendingOrderIDs) > 0 {
		go h.syncPendingTikTokOrderTransactions(client, tokenRecord, pendingOrderIDs, processedOrderIDs)
	}

	if syncedSettlements > 0 {
		log.Printf("[TIKTOK FINANCE SYNC] Successfully synchronized %d order finance figures via Finance API v202501", syncedSettlements)
	}
	return syncedSettlements
}

func (h *MarketplaceHandler) syncPendingTikTokOrderTransactions(client *apis.APIClient, tokenRecord *domain.MarketplaceToken, orderIDs []string, alreadyProcessed map[string]bool) {
	if client == nil || tokenRecord == nil || len(orderIDs) == 0 {
		return
	}

	for _, oID := range orderIDs {
		if alreadyProcessed[oID] {
			continue
		}

		ordResp, _, ordErr := client.FinanceV202501API.Finance202501OrdersOrderIdStatementTransactionsGet(context.Background(), oID).
			ShopCipher(tokenRecord.ShopID).
			XTtsAccessToken(tokenRecord.AccessToken).
			ContentType("application/json").
			Execute()

		if ordErr == nil && ordResp != nil && ordResp.Data != nil {
			data := ordResp.Data
			settleStr := data.GetSettlementAmount()
			settleAmt, _ := strconv.ParseFloat(settleStr, 64)

			feeStr := data.GetFeeAndTaxAmount()
			feeAmt, _ := strconv.ParseFloat(feeStr, 64)
			if feeAmt < 0 {
				feeAmt = -feeAmt
			}

			revStr := data.GetRevenueAmount()
			revAmt, _ := strconv.ParseFloat(revStr, 64)

			if settleAmt > 0 || feeAmt > 0 || revAmt > 0 {
				updates := map[string]interface{}{
					"settlement_status": "SETTLED",
					"settlement_amount": settleAmt,
					"platform_fee":      feeAmt,
				}
				if revAmt > 0 {
					updates["total_amount"] = revAmt
				}
				if data.GetOrderCreateTime() > 0 {
					t := time.Unix(data.GetOrderCreateTime(), 0)
					updates["settlement_date"] = t
				}

				h.db.Model(&domain.MarketplaceOrder{}).Where("order_id = ?", oID).Updates(updates)
			}
		}
	}
}

func (h *MarketplaceHandler) syncShopeeOrders(branchIDPtr *uuid.UUID) (int, error) {
	client := NewShopeeClient(h.db)
	tokenRecord, err := getValidShopeeToken(h.db, client.PartnerID, client.PartnerKey, branchIDPtr)
	if err != nil {
		return 0, err
	}

	shopeeOrders, err := client.FetchShopeeOrders(tokenRecord, 30)
	if err != nil {
		return 0, fmt.Errorf("gagal mengambil pesanan dari Shopee: %w", err)
	}

	whName := "Cabang"
	if tokenRecord.BranchID != nil {
		var b domain.Branch
		if err := h.db.First(&b, "id = ?", *tokenRecord.BranchID).Error; err == nil && b.Name != "" {
			whName = b.Name
		}
	}

	syncedCount := 0
	for _, spOrder := range shopeeOrders {
		orderID := spOrder.OrderSN
		if orderID == "" {
			continue
		}

		spStatus := strings.ToUpper(strings.TrimSpace(spOrder.OrderStatus))
		orderStatus := "VERIFIED"
		if spStatus == "SHIPPED" {
			orderStatus = "IN_TRANSIT"
		} else if spStatus == "COMPLETED" {
			orderStatus = "COMPLETED"
		} else if spStatus == "CANCELLED" || spStatus == "IN_CANCEL" {
			orderStatus = "CANCELLED"
		} else if spStatus == "TO_RETURN" {
			orderStatus = "RETURN_SUCCESS"
		} else if spStatus == "UNPAID" {
			orderStatus = "UNPAID"
		} else {
			// PROCESSED, READY_TO_SHIP -> VERIFIED
			orderStatus = "VERIFIED"
		}

		trackingID := strings.TrimSpace(spOrder.TrackingNumber)
		if trackingID == "" || trackingID == "-" {
			log.Printf("[SHOPEE SYNC] Melewati order %s karena belum memiliki ID Tracking/Resi", orderID)
			continue
		}

		buyerUsername := spOrder.BuyerUsername
		paymentMethod := spOrder.PaymentMethod
		shippingProvider := cleanShippingProvider(spOrder.ShippingCarrier, trackingID)
		totalAmount := spOrder.TotalAmount
		currency := spOrder.Currency
		if currency == "" {
			currency = "IDR"
		}
		createTime := spOrder.CreateTime

		var orderCreatedAt time.Time
		if createTime > 0 {
			orderCreatedAt = time.Unix(createTime, 0)
		} else {
			orderCreatedAt = time.Now()
		}

		if len(spOrder.ItemList) > 0 {
			for _, item := range spOrder.ItemList {
				skuID := strconv.FormatInt(item.ModelID, 10)
				if item.ModelID == 0 {
					skuID = strconv.FormatInt(item.ItemID, 10)
				}
				sellerSKU := item.ModelSKU
				if sellerSKU == "" {
					sellerSKU = item.ItemSKU
				}
				productName := item.ItemName
				if item.ModelName != "" {
					productName = fmt.Sprintf("%s - %s", item.ItemName, item.ModelName)
				}

				now := time.Now()
				mpOrder := &domain.MarketplaceOrder{
					BranchID:         tokenRecord.BranchID,
					OrderID:          orderID,
					OrderStatus:      orderStatus,
					SkuID:            skuID,
					LineItemID:       fmt.Sprintf("%s_%s", orderID, skuID),
					SellerSKU:        sellerSKU,
					ProductName:      productName,
					WarehouseName:    whName,
					Marketplace:      "shopee",
					BuyerUsername:    buyerUsername,
					TotalAmount:      totalAmount,
					Currency:         currency,
					PaymentMethod:    paymentMethod,
					TrackingID:       trackingID,
					ShippingProvider: shippingProvider,
					CreateTimeRaw:    createTime,
					CreatedAt:        orderCreatedAt,
					SyncedAt:         &now,
					ItemCount:        item.ModelQuantityPurchased,
				}
				if saveErr := h.usecase.SaveMarketplaceOrder(mpOrder); saveErr != nil {
					return syncedCount, fmt.Errorf("gagal menyimpan order Shopee %s: %w", orderID, saveErr)
				}
				syncedCount++
			}
		} else {
			now := time.Now()
			mpOrder := &domain.MarketplaceOrder{
				BranchID:         tokenRecord.BranchID,
				OrderID:          orderID,
				OrderStatus:      orderStatus,
				Marketplace:      "shopee",
				WarehouseName:    whName,
				BuyerUsername:    buyerUsername,
				TotalAmount:      totalAmount,
				Currency:         currency,
				PaymentMethod:    paymentMethod,
				TrackingID:       trackingID,
				ShippingProvider: shippingProvider,
				CreateTimeRaw:    createTime,
				CreatedAt:        orderCreatedAt,
				SyncedAt:         &now,
				ItemCount:        0,
			}
			if saveErr := h.usecase.SaveMarketplaceOrder(mpOrder); saveErr != nil {
				return syncedCount, fmt.Errorf("gagal menyimpan order Shopee %s: %w", orderID, saveErr)
			}
			syncedCount++
		}
	}
	return syncedCount, nil
}

type FetchSingleOrderReq struct {
	OrderID       string `json:"order_id" binding:"required"`
	Marketplace   string `json:"marketplace"`
	AsPendingPack bool   `json:"as_pending_pack"`
}

func (h *MarketplaceHandler) FetchSingleOrder(c *gin.Context) {
	var req FetchSingleOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID wajib diisi"})
		return
	}

	orderID := strings.TrimSpace(req.OrderID)
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID tidak boleh kosong"})
		return
	}

	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	mp := strings.ToLower(strings.TrimSpace(req.Marketplace))
	if mp == "" || mp == "auto" {
		hasLetter := false
		for _, r := range orderID {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				hasLetter = true
				break
			}
		}
		if hasLetter {
			mp = "shopee"
		} else {
			mp = "tiktok"
		}
	}

	if mp == "shopee" {
		count, err := h.fetchSingleShopeeOrder(branchIDPtr, orderID, req.AsPendingPack)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":      fmt.Sprintf("Berhasil menarik order Shopee %s (%d item tersimpan)", orderID, count),
			"synced_count": count,
			"order_id":     orderID,
			"marketplace":  "shopee",
		})
		return
	} else if mp == "tiktok" {
		count, err := h.fetchSingleTikTokOrder(branchIDPtr, orderID, req.AsPendingPack)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":      fmt.Sprintf("Berhasil menarik order TikTok %s (%d item tersimpan)", orderID, count),
			"synced_count": count,
			"order_id":     orderID,
			"marketplace":  "tiktok",
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Platform marketplace tidak didukung"})
}

func (h *MarketplaceHandler) fetchSingleShopeeOrder(branchIDPtr *uuid.UUID, orderID string, asPendingPack bool) (int, error) {
	client := NewShopeeClient(h.db)
	tokenRecord, err := getValidShopeeToken(h.db, client.PartnerID, client.PartnerKey, branchIDPtr)
	if err != nil {
		return 0, err
	}

	spOrder, err := client.FetchShopeeOrderDetail(tokenRecord, orderID)
	if err != nil {
		return 0, err
	}

	whName := "Cabang"
	if tokenRecord.BranchID != nil {
		var b domain.Branch
		if err := h.db.First(&b, "id = ?", *tokenRecord.BranchID).Error; err == nil && b.Name != "" {
			whName = b.Name
		}
	}

	spStatus := strings.ToUpper(strings.TrimSpace(spOrder.OrderStatus))
	var orderStatus string
	if asPendingPack {
		orderStatus = "VERIFIED"
	} else {
		if spStatus == "SHIPPED" {
			orderStatus = "IN_TRANSIT"
		} else if spStatus == "COMPLETED" {
			orderStatus = "COMPLETED"
		} else if spStatus == "CANCELLED" || spStatus == "IN_CANCEL" {
			orderStatus = "CANCELLED"
		} else if spStatus == "TO_RETURN" {
			orderStatus = "RETURN_SUCCESS"
		} else if spStatus == "UNPAID" {
			orderStatus = "UNPAID"
		} else {
			orderStatus = "VERIFIED"
		}
	}

	trackingID := strings.TrimSpace(spOrder.TrackingNumber)
	if trackingID == "" {
		trackingID = "-"
	}

	buyerUsername := spOrder.BuyerUsername
	paymentMethod := spOrder.PaymentMethod
	shippingProvider := cleanShippingProvider(spOrder.ShippingCarrier, trackingID)
	totalAmount := spOrder.TotalAmount
	currency := spOrder.Currency
	if currency == "" {
		currency = "IDR"
	}
	createTime := spOrder.CreateTime

	var orderCreatedAt time.Time
	if createTime > 0 {
		orderCreatedAt = time.Unix(createTime, 0)
	} else {
		orderCreatedAt = time.Now()
	}

	syncedCount := 0
	if len(spOrder.ItemList) > 0 {
		for _, item := range spOrder.ItemList {
			skuID := strconv.FormatInt(item.ModelID, 10)
			if item.ModelID == 0 {
				skuID = strconv.FormatInt(item.ItemID, 10)
			}
			sellerSKU := item.ModelSKU
			if sellerSKU == "" {
				sellerSKU = item.ItemSKU
			}
			productName := item.ItemName
			if item.ModelName != "" {
				productName = fmt.Sprintf("%s - %s", item.ItemName, item.ModelName)
			}

			now := time.Now()
			mpOrder := &domain.MarketplaceOrder{
				BranchID:         tokenRecord.BranchID,
				OrderID:          orderID,
				OrderStatus:      orderStatus,
				SkuID:            skuID,
				LineItemID:       fmt.Sprintf("%s_%s", orderID, skuID),
				SellerSKU:        sellerSKU,
				ProductName:      productName,
				WarehouseName:    whName,
				Marketplace:      "shopee",
				BuyerUsername:    buyerUsername,
				TotalAmount:      totalAmount,
				Currency:         currency,
				PaymentMethod:    paymentMethod,
				TrackingID:       trackingID,
				ShippingProvider: shippingProvider,
				CreateTimeRaw:    createTime,
				CreatedAt:        orderCreatedAt,
				SyncedAt:         &now,
				ItemCount:        item.ModelQuantityPurchased,
			}
			if saveErr := h.usecase.SaveMarketplaceOrder(mpOrder); saveErr == nil {
				syncedCount++
			} else {
				return 0, fmt.Errorf("gagal menyimpan pesanan: %w", saveErr)
			}
		}
	} else {
		now := time.Now()
		mpOrder := &domain.MarketplaceOrder{
			BranchID:         tokenRecord.BranchID,
			OrderID:          orderID,
			OrderStatus:      orderStatus,
			Marketplace:      "shopee",
			WarehouseName:    whName,
			BuyerUsername:    buyerUsername,
			TotalAmount:      totalAmount,
			Currency:         currency,
			PaymentMethod:    paymentMethod,
			TrackingID:       trackingID,
			ShippingProvider: shippingProvider,
			CreateTimeRaw:    createTime,
			CreatedAt:        orderCreatedAt,
			SyncedAt:         &now,
			ItemCount:        0,
		}
		if saveErr := h.usecase.SaveMarketplaceOrder(mpOrder); saveErr == nil {
			syncedCount++
		} else {
			return 0, fmt.Errorf("gagal menyimpan pesanan: %w", saveErr)
		}
	}

	return syncedCount, nil
}

func (h *MarketplaceHandler) fetchSingleTikTokOrder(branchIDPtr *uuid.UUID, orderID string, asPendingPack bool) (int, error) {
	tokenRecord, err := getValidTikTokToken(h.db, h.tikTokAppKey, h.tikTokAppSecret, branchIDPtr)
	if err != nil {
		return 0, err
	}

	cfg := apis.NewConfiguration()
	cfg.AddAppInfo(h.tikTokAppKey, h.tikTokAppSecret)
	client := apis.NewAPIClient(cfg)

	warehouseMap := make(map[string]string)
	whResp, _, errWh := client.LogisticsV202309API.Logistics202309WarehousesGet(context.Background()).
		ShopCipher(tokenRecord.ShopID).
		XTtsAccessToken(tokenRecord.AccessToken).
		ContentType("application/json").
		Execute()

	if errWh == nil && whResp != nil && whResp.Data != nil && whResp.Data.Warehouses != nil {
		for _, wh := range whResp.Data.Warehouses {
			if id, ok := wh.GetIdOk(); ok && id != nil {
				if name, ok := wh.GetNameOk(); ok && name != nil {
					warehouseMap[*id] = *name
				}
			}
		}
	}

	resp, httpRes, err := client.OrderV202309API.Order202309OrdersGet(context.Background()).
		ShopCipher(tokenRecord.ShopID).
		XTtsAccessToken(tokenRecord.AccessToken).
		Ids([]string{orderID}).
		ContentType("application/json").
		Execute()
	if err != nil {
		bodyStr := ""
		if httpRes != nil && httpRes.Body != nil {
			bodyBytes, _ := io.ReadAll(httpRes.Body)
			bodyStr = string(bodyBytes)
		}
		return 0, fmt.Errorf("tiktok api error: %v (%s)", err, bodyStr)
	}

	if resp == nil || resp.Data == nil || len(resp.Data.GetOrders()) == 0 {
		return 0, fmt.Errorf("order %s tidak ditemukan di TikTok", orderID)
	}

	tkOrder := resp.Data.GetOrders()[0]
	rawStatus := tkOrder.GetStatus()
	var orderStatus string
	if asPendingPack {
		orderStatus = "VERIFIED"
	} else {
		if rawStatus == "AWAITING_SHIPMENT" || rawStatus == "AWAITING_COLLECTION" {
			orderStatus = "VERIFIED"
		} else if rawStatus == "IN_TRANSIT" || rawStatus == "SHIPPED" {
			orderStatus = "IN_TRANSIT"
		} else if rawStatus == "DELIVERED" || rawStatus == "COMPLETED" {
			orderStatus = "COMPLETED"
		} else if rawStatus == "CANCELLED" {
			orderStatus = "CANCELLED"
		} else {
			orderStatus = rawStatus
		}
	}

	var totalAmount float64
	currency := "IDR"
	paymentMethod := ""
	if tkOrder.HasPayment() {
		payment := tkOrder.GetPayment()
		totalAmountStr := payment.GetTotalAmount()
		totalAmount, _ = strconv.ParseFloat(totalAmountStr, 64)
		currency = payment.GetCurrency()
	}
	paymentMethod = tkOrder.GetPaymentMethodName()

	whID := tkOrder.GetWarehouseId()
	whName := whID
	if name, exists := warehouseMap[whID]; exists {
		whName = name
	}

	assignedBranchID := resolveBranchIDByWarehouse(h.db, whName, tokenRecord.BranchID)

	trackingID := tkOrder.GetTrackingNumber()
	shippingProvider := cleanShippingProvider(tkOrder.GetShippingProvider(), trackingID)
	if shippingProvider == "" || shippingProvider == "-" {
		shippingProvider = cleanShippingProvider(tkOrder.GetDeliveryOptionName(), trackingID)
	}

	var orderCreatedAt time.Time
	if createTime := tkOrder.GetCreateTime(); createTime > 0 {
		orderCreatedAt = time.Unix(createTime, 0)
	} else {
		orderCreatedAt = time.Now()
	}

	syncedCount := 0
	if tkOrder.LineItems != nil {
		itemCounts := make(map[string]int)
		firstItems := make(map[string]order_v202309.Order202309GetOrderDetailResponseDataOrdersLineItems)

		for _, tkItem := range tkOrder.LineItems {
			skuID := tkItem.GetSkuId()
			itemCounts[skuID]++
			if _, exists := firstItems[skuID]; !exists {
				firstItems[skuID] = tkItem
			}
		}

		now := time.Now()
		for skuID, count := range itemCounts {
			tkItem := firstItems[skuID]
			mpOrder := &domain.MarketplaceOrder{
				BranchID:         assignedBranchID,
				OrderID:          orderID,
				OrderStatus:      orderStatus,
				SkuID:            skuID,
				LineItemID:       fmt.Sprintf("%s_%s", orderID, skuID),
				SellerSKU:        tkItem.GetSellerSku(),
				ProductName:      tkItem.GetProductName(),
				WarehouseName:    whName,
				Marketplace:      "tiktok",
				BuyerUsername:    tkOrder.GetBuyerEmail(),
				TotalAmount:      totalAmount,
				Currency:         currency,
				PaymentMethod:    paymentMethod,
				TrackingID:       trackingID,
				ShippingProvider: shippingProvider,
				CreateTimeRaw:    tkOrder.GetCreateTime(),
				CreatedAt:        orderCreatedAt,
				SyncedAt:         &now,
				ItemCount:        count,
			}
			if saveErr := h.usecase.SaveMarketplaceOrder(mpOrder); saveErr == nil {
				syncedCount++
			} else {
				return 0, fmt.Errorf("gagal menyimpan order: %w", saveErr)
			}
		}
	} else {
		now := time.Now()
		mpOrder := &domain.MarketplaceOrder{
			BranchID:         assignedBranchID,
			OrderID:          orderID,
			OrderStatus:      orderStatus,
			Marketplace:      "tiktok",
			WarehouseName:    whName,
			BuyerUsername:    tkOrder.GetBuyerEmail(),
			TotalAmount:      totalAmount,
			Currency:         currency,
			PaymentMethod:    paymentMethod,
			TrackingID:       trackingID,
			ShippingProvider: shippingProvider,
			CreateTimeRaw:    tkOrder.GetCreateTime(),
			CreatedAt:        orderCreatedAt,
			SyncedAt:         &now,
			ItemCount:        0,
		}
		if saveErr := h.usecase.SaveMarketplaceOrder(mpOrder); saveErr == nil {
			syncedCount++
		} else {
			return 0, fmt.Errorf("gagal menyimpan order: %w", saveErr)
		}
	}

	return syncedCount, nil
}

// ListMarketplaceProducts returns all marketplace products with optional in-memory filters.
func (h *MarketplaceHandler) ListMarketplaceProducts(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	products, err := h.usecase.GetMarketplaceProducts(branchIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Optional in-memory filters
	statusFilter := c.Query("status")
	keyword := strings.ToLower(c.Query("keyword"))
	marketplaceFilter := strings.ToLower(c.Query("marketplace"))

	if statusFilter == "" && keyword == "" && marketplaceFilter == "" {
		c.JSON(http.StatusOK, products)
		return
	}

	filtered := make([]domain.MarketplaceProduct, 0, len(products))
	for _, p := range products {
		if statusFilter != "" && p.Status != statusFilter {
			continue
		}
		if marketplaceFilter != "" && strings.ToLower(p.Marketplace) != marketplaceFilter {
			continue
		}
		if keyword != "" {
			if !strings.Contains(strings.ToLower(p.ProductName), keyword) &&
				!strings.Contains(strings.ToLower(p.SellerSKU), keyword) &&
				!strings.Contains(strings.ToLower(p.SkuID), keyword) &&
				!strings.Contains(strings.ToLower(p.ProductID), keyword) {
				continue
			}
		}
		filtered = append(filtered, p)
	}
	c.JSON(http.StatusOK, filtered)
}

// SyncMarketplaceProducts fetches products from TikTok and Shopee and upserts SKUs into the DB.
func (h *MarketplaceHandler) SyncMarketplaceProducts(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	targetMp := strings.ToLower(c.Query("marketplace"))
	totalSynced := 0
	var syncErrors []string
	var mu sync.Mutex

	var wg sync.WaitGroup

	// Sync TikTok
	if targetMp == "" || targetMp == "all" || targetMp == "tiktok" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count, err := h.syncTikTokProducts(branchIDPtr)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("TikTok: %v", err))
			} else {
				totalSynced += count
			}
		}()
	}

	// Sync Shopee
	if targetMp == "" || targetMp == "all" || targetMp == "shopee" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count, err := h.syncShopeeProducts(branchIDPtr)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("Shopee: %v", err))
			} else {
				totalSynced += count
			}
		}()
	}

	wg.Wait()

	if totalSynced == 0 && len(syncErrors) > 0 && (targetMp == "tiktok" || targetMp == "shopee") {
		c.JSON(http.StatusBadRequest, gin.H{"error": strings.Join(syncErrors, "; ")})
		return
	}

	msg := fmt.Sprintf("Sync produk selesai (%d SKU produk tersinkron)", totalSynced)
	if len(syncErrors) > 0 {
		msg += fmt.Sprintf(" (Catatan: %s)", strings.Join(syncErrors, ", "))
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      msg,
		"synced_count": totalSynced,
	})
}

func (h *MarketplaceHandler) syncTikTokProducts(branchIDPtr *uuid.UUID) (int, error) {
	tokenRecord, err := getValidTikTokToken(h.db, h.tikTokAppKey, h.tikTokAppSecret, branchIDPtr)
	if err != nil {
		return 0, err
	}

	cfg := apis.NewConfiguration()
	cfg.AddAppInfo(h.tikTokAppKey, h.tikTokAppSecret)
	client := apis.NewAPIClient(cfg)

	syncedCount := 0
	nextPageToken := ""
	for {
		req := client.ProductV202309API.Product202309ProductsSearchPost(context.Background()).
			ShopCipher(tokenRecord.ShopID).
			XTtsAccessToken(tokenRecord.AccessToken).
			PageSize(50).
			ContentType("application/json").
			Product202309SearchProductsRequestBody(product_v202309.Product202309SearchProductsRequestBody{})

		if nextPageToken != "" {
			req = req.PageToken(nextPageToken)
		}

		resp, httpRes, err := req.Execute()
		if err != nil {
			bodyStr := ""
			if httpRes != nil && httpRes.Body != nil {
				bodyBytes, _ := io.ReadAll(httpRes.Body)
				bodyStr = string(bodyBytes)
			}
			return 0, fmt.Errorf("tiktok product error: %v (%s)", err, bodyStr)
		}

		if resp == nil || resp.Data == nil || len(resp.Data.Products) == 0 {
			break
		}

		for _, tkProduct := range resp.Data.Products {
			productID := tkProduct.GetId()
			productName := tkProduct.GetTitle()
			productStatus := tkProduct.GetStatus()
			salesRegions := strings.Join(tkProduct.GetSalesRegions(), ",")

			for _, sku := range tkProduct.GetSkus() {
				skuID := sku.GetId()
				if skuID == "" {
					skuID = productID
				}

				var priceAmount float64
				price := sku.GetPrice()
				if price.SalePrice != nil {
					priceAmount, _ = strconv.ParseFloat(*price.SalePrice, 64)
				}

				var totalStock int
				for _, inv := range sku.GetInventory() {
					totalStock += int(inv.GetQuantity())
				}

				targetBranchID := branchIDPtr
				if targetBranchID == nil {
					targetBranchID = tokenRecord.BranchID
				}

				mpProduct := &domain.MarketplaceProduct{
					BranchID:     targetBranchID,
					Marketplace:  "tiktok",
					ProductID:    productID,
					ProductName:  productName,
					Status:       productStatus,
					SkuID:        skuID,
					SellerSKU:    sku.GetSellerSku(),
					PriceAmount:  priceAmount,
					Stock:        totalStock,
					SalesRegions: salesRegions,
					LastSyncedAt: time.Now(),
				}
				if saveErr := h.usecase.SaveMarketplaceProduct(mpProduct); saveErr == nil {
					syncedCount++
				}
			}
		}

		nextPageToken = resp.Data.GetNextPageToken()
		if nextPageToken == "" {
			break
		}
	}

	return syncedCount, nil
}

func (h *MarketplaceHandler) syncShopeeProducts(branchIDPtr *uuid.UUID) (int, error) {
	client := NewShopeeClient(h.db)
	tokenRecord, err := getValidShopeeToken(h.db, client.PartnerID, client.PartnerKey, branchIDPtr)
	if err != nil {
		return 0, err
	}

	products, err := client.FetchShopeeProducts(tokenRecord)
	if err != nil {
		return 0, fmt.Errorf("shopee product error: %w", err)
	}

	syncedCount := 0
	for _, p := range products {
		prod := p
		if branchIDPtr != nil {
			prod.BranchID = branchIDPtr
		}
		if saveErr := h.usecase.SaveMarketplaceProduct(&prod); saveErr == nil {
			syncedCount++
		}
	}
	return syncedCount, nil
}

func (h *MarketplaceHandler) PackOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	_, adminUsername := getActorInfo(c)
	if adminUsername == "" {
		adminUsername = "admin"
	}

	// Body opsional — backward compatible dengan single-item flow
	var req struct {
		BoxStockID      *string                    `json:"box_stock_id"`
		BoxQty          int                        `json:"box_qty"`
		AdditionalItems []domain.AdditionalItemReq `json:"additional_items"`
		SkipDefaultBox  bool                       `json:"skip_default_box"`
	}
	_ = c.ShouldBindJSON(&req)

	var boxStockUUID *uuid.UUID
	if req.BoxStockID != nil && *req.BoxStockID != "" {
		if parsed, err := uuid.Parse(*req.BoxStockID); err == nil {
			boxStockUUID = &parsed
		}
	}

	packReq := domain.PackOrderReq{
		BoxStockID:      boxStockUUID,
		BoxQty:          req.BoxQty,
		AdditionalItems: req.AdditionalItems,
		SkipDefaultBox:  req.SkipDefaultBox,
	}

	if err := h.usecase.PackOrder(orderID, adminUsername, packReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast event via SSE
	payload := fmt.Sprintf(`{"order_id":"%s", "status":"PACKED", "actor":"%s"}`, orderID, adminUsername)
	MarketplaceBroker.Broadcast(payload)

	c.JSON(http.StatusOK, gin.H{"message": "Order packed successfully"})
}

func (h *MarketplaceHandler) VerifyOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	_, adminUsername := getActorInfo(c)
	if adminUsername == "" {
		adminUsername = "admin"
	}
	if err := h.usecase.VerifyOrder(orderID, adminUsername); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast event via SSE
	payload := fmt.Sprintf(`{"order_id":"%s", "status":"VERIFIED", "actor":"%s"}`, orderID, adminUsername)
	MarketplaceBroker.Broadcast(payload)

	c.JSON(http.StatusOK, gin.H{"message": "Order verified successfully"})
}

func (h *MarketplaceHandler) StreamOrderUpdates(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// Optional: add CORS headers if not handled by middleware
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan string, 100)
	MarketplaceBroker.AddClient(clientChan)
	defer MarketplaceBroker.RemoveClient(clientChan)

	// Keep alive heartbeat to prevent connection timeout
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				MarketplaceBroker.Broadcast(": keepalive\n\n")
			case <-c.Request.Context().Done():
				return // Exit goroutine when client disconnects
			}
		}
	}()

	for {
		select {
		case msg := <-clientChan:
			// If it's a heartbeat, it already has the required format
			if strings.HasPrefix(msg, ":") {
				c.Writer.Write([]byte(msg))
			} else {
				c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", msg)))
			}
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func (h *MarketplaceHandler) GetProductMappings(c *gin.Context) {
	skuID := c.Query("sku_id")
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	mappings, err := h.usecase.GetProductMappings(skuID, branchIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mappings)
}

func (h *MarketplaceHandler) SaveProductMappings(c *gin.Context) {
	skuID := c.Query("sku_id")
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	var req struct {
		Mappings []domain.MarketplaceProductMapping `json:"mappings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure sku_id and branch_id matches
	for i := range req.Mappings {
		req.Mappings[i].SkuID = skuID
		if branchIDPtr != nil {
			req.Mappings[i].BranchID = branchIDPtr
		}
	}

	if err := h.usecase.SaveProductMappings(skuID, branchIDPtr, req.Mappings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mappings saved successfully"})
}

func (h *MarketplaceHandler) GetScanHistory(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	filter := repository.ScanHistoryFilter{
		BranchID: branchIDPtr,
	}

	if f := c.Query("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			filter.From = &t
		}
	}
	if t := c.Query("to"); t != "" {
		if parsed, err := time.Parse("2006-01-02", t); err == nil {
			// Add 23:59:59 to the "to" date to include the whole day
			endOfDay := parsed.Add(24 * time.Hour).Add(-time.Second)
			filter.To = &endOfDay
		}
	}
	filter.Action = c.Query("action")
	filter.WarehouseName = c.Query("warehouse")

	histories, err := h.usecase.GetScanHistories(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": histories})
}

func (h *MarketplaceHandler) GetScanHistoryWarehouses(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	warehouses, err := h.usecase.GetScanHistoryWarehouses(branchIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": warehouses})
}

// ListSettlements returns paginated settlement orders with financial stats.
func (h *MarketplaceHandler) ListSettlements(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	filter := repository.MarketplaceSettlementFilter{
		BranchID:         branchIDPtr,
		Marketplace:      c.Query("marketplace"),
		Tab:              c.Query("tab"),
		SettlementStatus: c.Query("settlement_status"),
		ReturnStatus:     c.Query("return_status"),
		PackingStatus:    c.Query("packing_status"),
		Keyword:          c.Query("keyword"),
		ProductName:      c.Query("product_name"),
		WarehouseName:    c.Query("warehouse_name"),
		ShippingProvider: c.Query("shipping_provider"),
		DateType:         c.Query("date_type"),
		Page:             1,
		PageSize:         20,
	}

	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			filter.Page = n
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			filter.PageSize = n
		}
	}
	if f := c.Query("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			filter.From = &t
		}
	}
	if t := c.Query("to"); t != "" {
		if parsed, err := time.Parse("2006-01-02", t); err == nil {
			endOfDay := parsed.Add(24 * time.Hour).Add(-time.Second)
			filter.To = &endOfDay
		}
	}

	orders, total, totalSettled, totalFee, totalRefund, err := h.usecase.GetSettlementOrders(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":                 orders,
		"total":                total,
		"total_settled_amount": totalSettled,
		"total_platform_fee":   totalFee,
		"total_refund_amount":  totalRefund,
	})
}

// GetSettlementSummary calculates aggregated metrics across all settlement tabs.
func (h *MarketplaceHandler) GetSettlementSummary(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	var fromPtr, toPtr *time.Time
	if f := c.Query("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			fromPtr = &t
		}
	}
	if t := c.Query("to"); t != "" {
		if parsed, err := time.Parse("2006-01-02", t); err == nil {
			endOfDay := parsed.Add(24 * time.Hour).Add(-time.Second)
			toPtr = &endOfDay
		}
	}

	marketplace := c.Query("marketplace")
	dateType := c.Query("date_type")
	summary, err := h.usecase.GetSettlementSummary(branchIDPtr, fromPtr, toPtr, marketplace, dateType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// RestockOrder processes restock of returned/cancelled orders back into inventory.
func (h *MarketplaceHandler) RestockOrder(c *gin.Context) {
	var req domain.RestockOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID wajib diisi"})
		return
	}

	_, actorName := getActorInfo(c)
	if err := h.usecase.RestockOrder(req.OrderID, actorName, req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Pesanan %s berhasil di-restock ke inventori gudang", req.OrderID),
		"order_id": req.OrderID,
	})
}

// ImportSettlements imports batch settlement updates from Excel/CSV parsed records.
func (h *MarketplaceHandler) ImportSettlements(c *gin.Context) {
	var req domain.BatchSettlementImportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data impor tidak valid: " + err.Error()})
		return
	}

	if len(req.Records) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada data rekonsiliasi yang dikirim"})
		return
	}

	_, actorName := getActorInfo(c)
	count, err := h.usecase.BatchImportSettlements(req.Records, actorName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengimpor rekonsiliasi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Berhasil memproses & merekonsiliasi %d pesanan", count),
		"updated_count": count,
	})
}

// UpdateSettlementOrder manually updates financial / settlement data for a single order.
func (h *MarketplaceHandler) UpdateSettlementOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id diperlukan"})
		return
	}

	var req domain.UpdateSettlementOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid"})
		return
	}

	if err := h.usecase.UpdateSettlementOrder(orderID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data pencairan pesanan berhasil diperbarui"})
}

// AutoClassifySettlements batches automatic settlement classification for completed/delivered & cancelled orders.
func (h *MarketplaceHandler) AutoClassifySettlements(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	count, err := h.usecase.AutoClassifySettlements(branchIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menjalankan auto-rekonsiliasi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Auto-rekonsiliasi selesai! %d pesanan berhasil diklasifikasikan status pencairannya.", count),
		"updated_count": count,
	})
}

// SyncTikTokFinance performs bulk synchronization using Finance Transaction API v202501.
// Strictly isolated: only financial columns are updated, never touching operational order status.
func (h *MarketplaceHandler) setFinanceSyncStatus(syncID string, status financeSyncStatus) {
	h.financeSyncMu.Lock()
	defer h.financeSyncMu.Unlock()
	if old, ok := h.financeSyncJobs[syncID]; ok {
		if status.Platform == "" {
			status.Platform = old.Platform
		}
		if status.ProcessedCount == 0 && status.Status != "processing" {
			status.ProcessedCount = old.ProcessedCount
		}
		if status.LatestPaymentAt == nil {
			status.LatestPaymentAt = old.LatestPaymentAt
		}
	}
	status.UpdatedAt = time.Now()
	for id, old := range h.financeSyncJobs {
		if time.Since(old.UpdatedAt) > time.Hour {
			delete(h.financeSyncJobs, id)
		}
	}
	h.financeSyncJobs[syncID] = status
}

func (h *MarketplaceHandler) GetTikTokFinanceSyncStatus(c *gin.Context) {
	h.financeSyncMu.RLock()
	status, ok := h.financeSyncJobs[c.Param("sync_id")]
	h.financeSyncMu.RUnlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status sinkronisasi tidak ditemukan atau sudah kedaluwarsa"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *MarketplaceHandler) beginTikTokFinanceSync() (string, bool) {
	h.financeSyncMu.Lock()
	defer h.financeSyncMu.Unlock()
	for id, job := range h.financeSyncJobs {
		if job.Platform == "tiktok" && job.Status == "processing" {
			return id, false
		}
	}
	id := uuid.NewString()
	h.financeSyncJobs[id] = financeSyncStatus{Status: "processing", Platform: "tiktok", Phase: "Menyiapkan akun TikTok", UpdatedAt: time.Now()}
	return id, true
}

func (h *MarketplaceHandler) paceTikTokFinanceRequest() {
	h.financeRequestMu.Lock()
	now := time.Now()
	callAt := now
	if h.financeNextCall.After(now) {
		callAt = h.financeNextCall
	}
	h.financeNextCall = callAt.Add(time.Second)
	h.financeRequestMu.Unlock()
	if delay := time.Until(callAt); delay > 0 {
		time.Sleep(delay)
	}
}

// Retry the same page after a rate limit. Every attempt builds its own request
// context because the SDK request holds the context passed at construction.
func retryTikTokFinanceRequest[T any](pace func(), wait func(time.Duration), onRetry func(time.Duration), execute func() (T, *http.Response, error), getCode func(T) int) (T, *http.Response, error) {
	const maxRetries = 6
	for attempt := 0; ; attempt++ {
		pace()
		result, response, err := execute()
		code := getCode(result)
		rateLimited := code == 36009002 || (response != nil && response.StatusCode == http.StatusTooManyRequests)
		if !rateLimited && err != nil {
			rateLimited = strings.Contains(err.Error(), "429 Too Many Requests")
		}
		if !rateLimited {
			return result, response, err
		}
		if attempt == maxRetries {
			return result, response, fmt.Errorf("batas laju API TikTok masih tercapai setelah %d percobaan; data yang sudah diproses tersimpan, coba sinkronisasi lagi beberapa menit kemudian", attempt+1)
		}
		delay := tikTokFinanceRetryDelay(attempt, response, time.Now())
		log.Printf("[TIKTOK FINANCE SYNC] API rate limit, retrying same page in %s (attempt %d/%d)", delay, attempt+1, maxRetries)
		if onRetry != nil {
			onRetry(delay)
		}
		wait(delay)
	}
}

func tikTokFinanceRetryDelay(attempt int, response *http.Response, now time.Time) time.Duration {
	delay := time.Duration(3<<min(attempt, 4)) * time.Second
	if response != nil {
		if raw := strings.TrimSpace(response.Header.Get("Retry-After")); raw != "" {
			if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			} else if at, err := http.ParseTime(raw); err == nil && at.After(now) {
				delay = at.Sub(now)
			}
		}
	}
	if delay > 2*time.Minute {
		return 2 * time.Minute
	}
	return delay
}

func (h *MarketplaceHandler) tikTokFinanceShopIDs(branchIDPtr *uuid.UUID) ([]string, error) {
	if branchIDPtr != nil {
		token, err := getValidTikTokToken(h.db, h.tikTokAppKey, h.tikTokAppSecret, branchIDPtr)
		if err != nil {
			return nil, err
		}
		return []string{token.ShopID}, nil
	}
	var shopIDs []string
	if err := h.db.Model(&domain.MarketplaceToken{}).Where("marketplace = ? AND shop_id != ''", "tiktok").
		Distinct("shop_id").Pluck("shop_id", &shopIDs).Error; err != nil {
		return nil, err
	}
	if len(shopIDs) == 0 {
		return nil, fmt.Errorf("belum ada akun TikTok yang terhubung")
	}
	return shopIDs, nil
}

func (h *MarketplaceHandler) SyncTikTokFinance(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	days := 30
	var reqBody struct {
		BranchID *string `json:"branch_id"`
		Days     *int    `json:"days"`
		Async    bool    `json:"async"`
	}
	if err := c.ShouldBindJSON(&reqBody); err == nil {
		if reqBody.Days != nil && *reqBody.Days >= 0 {
			days = *reqBody.Days
		}
		if reqBody.BranchID != nil && *reqBody.BranchID != "" {
			if bid, bErr := uuid.Parse(*reqBody.BranchID); bErr == nil && bid != uuid.Nil {
				branchIDPtr = &bid
			}
		}
	}
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d >= 0 {
			days = d
		}
	}

	shopIDs, err := h.tikTokFinanceShopIDs(branchIDPtr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca akun TikTok: " + err.Error()})
		return
	}

	cfg := apis.NewConfiguration()
	cfg.AddAppInfo(h.tikTokAppKey, h.tikTokAppSecret)
	client := apis.NewAPIClient(cfg)
	syncID, started := h.beginTikTokFinanceSync()
	if !started {
		c.JSON(http.StatusOK, gin.H{"status": "processing", "in_background": true, "sync_id": syncID, "message": "Sinkronisasi finansial TikTok sudah berjalan. Menampilkan progres proses yang sama."})
		return
	}

	// Run in background goroutine to eliminate 504 Gateway Timeout
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("[TIKTOK FINANCE SYNC] Panic: %v", recovered)
				h.setFinanceSyncStatus(syncID, financeSyncStatus{Status: "failed", Error: "Proses sinkronisasi TikTok terhenti secara tidak terduga"})
			}
		}()
		totalSynced, totalStatements, totalPaid, totalSettled, totalProcessed := 0, 0, 0, 0, 0
		var latestPayment *time.Time
		var syncErrors []string
		for _, shopID := range shopIDs {
			h.setFinanceSyncStatus(syncID, financeSyncStatus{Status: "processing", SyncedCount: totalSynced, StatementsCount: totalStatements, PaidStatementsCount: totalPaid, SettledStatementsCount: totalSettled, LatestPaymentAt: latestPayment, ProcessedCount: totalProcessed, Phase: "Membaca statement toko " + shopID})
			tokenRecord, err := getValidTikTokTokenByShopID(h.db, h.tikTokAppKey, h.tikTokAppSecret, shopID)
			if err != nil {
				log.Printf("[TIKTOK FINANCE SYNC] Token error for shop %s: %v", shopID, err)
				syncErrors = append(syncErrors, fmt.Sprintf("shop %s: %v", shopID, err))
				continue
			}
			shopProcessed := 0
			syncedCount, statementsCount, paidCount, settledCount, shopLatest, syncErr := h.executeTikTokFinanceSync(client, tokenRecord, days, func(synced, statements, paid, settled, processed int, currentLatest *time.Time, phase string) {
				totalProcessed += processed - shopProcessed
				shopProcessed = processed
				if currentLatest != nil && (latestPayment == nil || currentLatest.After(*latestPayment)) {
					latestPayment = currentLatest
				}
				h.setFinanceSyncStatus(syncID, financeSyncStatus{Status: "processing", SyncedCount: totalSynced + synced, StatementsCount: totalStatements + statements, PaidStatementsCount: totalPaid + paid, SettledStatementsCount: totalSettled + settled, LatestPaymentAt: latestPayment, ProcessedCount: totalProcessed, Phase: phase})
			})
			totalSynced += syncedCount
			totalStatements += statementsCount
			totalPaid += paidCount
			totalSettled += settledCount
			if shopLatest != nil && (latestPayment == nil || shopLatest.After(*latestPayment)) {
				latestPayment = shopLatest
			}
			if syncErr != nil {
				log.Printf("[TIKTOK FINANCE SYNC] Shop %s failed: %v", shopID, syncErr)
				syncErrors = append(syncErrors, fmt.Sprintf("shop %s: %v", shopID, syncErr))
				continue
			}
		}
		if len(syncErrors) > 0 {
			errorText := strings.Join(syncErrors, "; ")
			h.setFinanceSyncStatus(syncID, financeSyncStatus{Status: "failed", SyncedCount: totalSynced, StatementsCount: totalStatements, PaidStatementsCount: totalPaid, SettledStatementsCount: totalSettled, LatestPaymentAt: latestPayment, ProcessedCount: totalProcessed, Error: errorText})
			MarketplaceBroker.Broadcast(fmt.Sprintf(`{"type":"FINANCE_SYNC_FAILED","error":%q,"synced_count":%d}`, errorText, totalSynced))
			return
		}
		log.Printf("[TIKTOK FINANCE SYNC] Complete: %d orders from %d statements across %d shops.", totalSynced, totalStatements, len(shopIDs))
		h.setFinanceSyncStatus(syncID, financeSyncStatus{Status: "completed", SyncedCount: totalSynced, StatementsCount: totalStatements, PaidStatementsCount: totalPaid, SettledStatementsCount: totalSettled, LatestPaymentAt: latestPayment, ProcessedCount: totalProcessed, Phase: "Selesai"})
		MarketplaceBroker.Broadcast(fmt.Sprintf(`{"type":"FINANCE_SYNC_COMPLETED","synced_count":%d,"statements_count":%d}`, totalSynced, totalStatements))
	}()

	c.JSON(http.StatusOK, gin.H{
		"status":        "processing",
		"in_background": true,
		"sync_id":       syncID,
		"message":       "Sinkronisasi Finansial TikTok sedang berjalan di latar belakang. Data pencairan dan fee akan diperbarui secara bertahap.",
	})
}

func (h *MarketplaceHandler) executeTikTokFinanceSync(client *apis.APIClient, tokenRecord *domain.MarketplaceToken, days int, onProgress func(int, int, int, int, int, *time.Time, string)) (int, int, int, int, *time.Time, error) {
	syncedCount, processedCount := 0, 0
	matchedOrders := make(map[string]struct{})
	statementsCount := 0
	paidStatementsCount := 0
	settledStatementsCount := 0
	var latestPayment *time.Time
	statementCursor := ""

	cutoff := int64(0)
	if days > 0 {
		cutoff = time.Now().AddDate(0, 0, -days).Unix()
	}

	for {
		// Apply older statements first so the latest transaction wins on repeated syncs.
		statementResp, httpRes, err := retryTikTokFinanceRequest(h.paceTikTokFinanceRequest, time.Sleep, func(delay time.Duration) {
			if onProgress != nil {
				onProgress(syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, processedCount, latestPayment, fmt.Sprintf("TikTok membatasi request; menunggu %s sebelum melanjutkan statement", delay.Round(time.Second)))
			}
		}, func() (*finance_v202309.Finance202309GetStatementsResponse, *http.Response, error) {
			statementCtx, cancelStatement := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancelStatement()
			statementReq := client.FinanceV202309API.Finance202309StatementsGet(statementCtx).
				ShopCipher(tokenRecord.ShopID).
				XTtsAccessToken(tokenRecord.AccessToken).
				ContentType("application/json").
				SortField("statement_time").
				SortOrder("ASC").
				PageSize(50)
			if cutoff > 0 {
				statementReq = statementReq.StatementTimeGe(cutoff)
			}
			if statementCursor != "" {
				statementReq = statementReq.PageToken(statementCursor)
			}
			return statementReq.Execute()
		}, func(resp *finance_v202309.Finance202309GetStatementsResponse) int {
			if resp == nil {
				return 0
			}
			return int(resp.GetCode())
		})
		if err != nil {
			bodyStr := ""
			if httpRes != nil && httpRes.Body != nil {
				bodyBytes, _ := io.ReadAll(httpRes.Body)
				bodyStr = string(bodyBytes)
			}
			return syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, latestPayment, fmt.Errorf("Gagal memanggil TikTok Statement API: %v (%s)", err, bodyStr)
		}
		if statementResp != nil && statementResp.GetCode() != 0 {
			return syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, latestPayment, fmt.Errorf("TikTok Statement API menolak permintaan (%d): %s", statementResp.GetCode(), statementResp.GetMessage())
		}

		if statementResp == nil || statementResp.Data == nil || len(statementResp.Data.Statements) == 0 {
			break
		}

		for _, stmt := range statementResp.Data.Statements {
			stmtID := stmt.GetId()
			if stmtID == "" {
				continue
			}
			if cutoff > 0 && stmt.GetStatementTime() < cutoff {
				continue
			}
			statementsCount++
			paymentStatus := strings.ToUpper(strings.TrimSpace(stmt.GetPaymentStatus()))

			var paymentTime *time.Time
			var statementTime *time.Time
			if stmt.GetPaymentTime() > 0 {
				t := time.Unix(stmt.GetPaymentTime(), 0)
				paymentTime = &t
			}
			if stmt.GetStatementTime() > 0 {
				t := time.Unix(stmt.GetStatementTime(), 0)
				statementTime = &t
			}
			if paymentStatus == "PAID" {
				paidStatementsCount++
				if paymentTime != nil && (latestPayment == nil || paymentTime.After(*latestPayment)) {
					latestPayment = paymentTime
				}
			} else if paymentStatus == "SETTLED" {
				settledStatementsCount++
			}
			if onProgress != nil {
				onProgress(syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, processedCount, latestPayment, "Membaca transaksi statement "+stmtID)
			}

			pageCursor := ""
			for {
				txResp, txHTTPRes, txErr := retryTikTokFinanceRequest(h.paceTikTokFinanceRequest, time.Sleep, func(delay time.Duration) {
					if onProgress != nil {
						onProgress(syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, processedCount, latestPayment, fmt.Sprintf("TikTok membatasi request; menunggu %s lalu mengulang statement %s", delay.Round(time.Second), stmtID))
					}
				}, func() (*finance_v202501.Finance202501GetTransactionsbyStatementResponse, *http.Response, error) {
					txCtx, cancelTx := context.WithTimeout(context.Background(), 45*time.Second)
					defer cancelTx()
					txReq := client.FinanceV202501API.Finance202501StatementsStatementIdStatementTransactionsGet(txCtx, stmtID).
						ShopCipher(tokenRecord.ShopID).
						XTtsAccessToken(tokenRecord.AccessToken).
						ContentType("application/json").
						SortField("order_create_time").
						SortOrder("DESC").
						PageSize(100)
					if pageCursor != "" {
						txReq = txReq.PageToken(pageCursor)
					}
					return txReq.Execute()
				}, func(resp *finance_v202501.Finance202501GetTransactionsbyStatementResponse) int {
					if resp == nil {
						return 0
					}
					return int(resp.GetCode())
				})
				if txErr != nil {
					bodyStr := ""
					if txHTTPRes != nil && txHTTPRes.Body != nil {
						bodyBytes, _ := io.ReadAll(txHTTPRes.Body)
						bodyStr = string(bodyBytes)
					}
					return syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, latestPayment, fmt.Errorf("Gagal mengambil transaksi statement TikTok %s: %v (%s)", stmtID, txErr, bodyStr)
				}
				if txResp != nil && txResp.GetCode() != 0 {
					return syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, latestPayment, fmt.Errorf("TikTok menolak transaksi statement %s (%d): %s", stmtID, txResp.GetCode(), txResp.GetMessage())
				}

				pageRecords := make([]domain.FinanceTransactionRecord, 0, 100)
				if txResp != nil && txResp.Data != nil && len(txResp.Data.Transactions) > 0 {
					for _, tx := range txResp.Data.Transactions {
						processedCount++
						// Reserve and adjustment records are not the order settlement itself.
						// Applying them to the order row overwrites its fee and payout.
						if strings.ToUpper(tx.GetType()) != "ORDER" {
							continue
						}
						oid := tx.GetOrderId()
						if oid == "" {
							continue
						}

						settlementAmt, settlementProvided := parseOptionalTikTokAmount(tx.GetSettlementAmount())
						feeTaxAmt, feeTaxProvided := parseOptionalTikTokFeeAmount(tx.GetFeeTaxAmount())
						shippingCostAmt, _ := strconv.ParseFloat(tx.GetShippingCostAmount(), 64)
						adjAmt, _ := strconv.ParseFloat(tx.GetAdjustmentAmount(), 64)

						breakdownJSON := ""
						if tx.FeeTaxBreakdown != nil {
							if bBytes, bErr := json.Marshal(tx.FeeTaxBreakdown); bErr == nil {
								breakdownJSON = string(bBytes)
							}
						}

						status, settledAt, holdReason := classifyTikTokStatementPayment(paymentStatus, paymentTime, statementTime)

						pageRecords = append(pageRecords, domain.FinanceTransactionRecord{
							OrderID:            oid,
							Marketplace:        "tiktok",
							BranchID:           tokenRecord.BranchID,
							SettlementAmount:   settlementAmt,
							SettlementProvided: settlementProvided,
							FeeAndTaxAmount:    feeTaxAmt,
							FeeAndTaxProvided:  feeTaxProvided,
							ShippingCostAmount: shippingCostAmt,
							AdjustmentAmount:   adjAmt,
							FeeTaxBreakdown:    breakdownJSON,
							SettlementStatus:   status,
							SettlementDate:     settledAt,
							StatementID:        stmtID,
							HoldReason:         holdReason,
						})
					}
				}
				if len(pageRecords) > 0 {
					writtenIDs, saveErr := h.usecase.SaveFinanceTransactions(pageRecords)
					if saveErr != nil {
						return syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, latestPayment, fmt.Errorf("Gagal menyimpan transaksi finansial statement %s: %w", stmtID, saveErr)
					}
					for _, orderID := range writtenIDs {
						matchedOrders[orderID] = struct{}{}
					}
					syncedCount = len(matchedOrders)
				}
				if onProgress != nil {
					onProgress(syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, processedCount, latestPayment, "Memproses transaksi statement "+stmtID)
				}

				if txResp == nil || txResp.Data == nil || txResp.Data.GetNextPageToken() == "" {
					break
				}
				pageCursor = txResp.Data.GetNextPageToken()
			}
		}

		if statementResp.Data.GetNextPageToken() == "" {
			break
		}
		statementCursor = statementResp.Data.GetNextPageToken()
	}

	return syncedCount, statementsCount, paidStatementsCount, settledStatementsCount, latestPayment, nil
}

func classifyTikTokStatementPayment(paymentStatus string, paymentTime, statementTime *time.Time) (string, *time.Time, string) {
	status := strings.ToUpper(strings.TrimSpace(paymentStatus))
	if status == "PAID" {
		return "SETTLED", paymentTime, ""
	}
	if status == "SETTLED" {
		return "SETTLED", statementTime, ""
	}
	reason := "Menunggu pembayaran statement TikTok"
	if status != "" {
		reason += " (" + status + ")"
	}
	return "HOLD", nil, reason
}

// TikTok can report deductions with a negative sign; the UI stores fee as a positive deduction.
func parseTikTokFeeAmount(raw string) float64 {
	amount, _ := parseOptionalTikTokFeeAmount(raw)
	return amount
}

func parseOptionalTikTokFeeAmount(raw string) (float64, bool) {
	amount, provided := parseOptionalTikTokAmount(raw)
	return math.Abs(amount), provided
}

func parseOptionalTikTokAmount(raw string) (float64, bool) {
	if strings.TrimSpace(raw) == "" {
		return 0, false
	}
	amount, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, false
	}
	return amount, true
}

// SyncTikTokOrderFinance performs drill-down / single order retry using Get Transactions by Order API v202501.
// Strictly isolated: only financial columns are updated, never touching operational order status.
func (h *MarketplaceHandler) SyncTikTokOrderFinance(c *gin.Context) {
	orderID := strings.TrimSpace(c.Param("order_id"))
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id harus disertakan"})
		return
	}

	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	shopIDs, err := h.tikTokFinanceShopIDs(branchIDPtr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca akun TikTok: " + err.Error()})
		return
	}

	cfg := apis.NewConfiguration()
	cfg.AddAppInfo(h.tikTokAppKey, h.tikTokAppSecret)
	client := apis.NewAPIClient(cfg)

	var resp *finance_v202501.Finance202501GetTransactionsbyOrderResponse
	var matchedBranchID *uuid.UUID
	var lookupErr error
	for _, shopID := range shopIDs {
		tokenRecord, tokenErr := getValidTikTokTokenByShopID(h.db, h.tikTokAppKey, h.tikTokAppSecret, shopID)
		if tokenErr != nil {
			lookupErr = tokenErr
			continue
		}
		candidate, _, apiErr := client.FinanceV202501API.Finance202501OrdersOrderIdStatementTransactionsGet(context.Background(), orderID).
			ShopCipher(tokenRecord.ShopID).
			XTtsAccessToken(tokenRecord.AccessToken).
			ContentType("application/json").
			Execute()
		if apiErr != nil {
			lookupErr = apiErr
			continue
		}
		if candidate != nil && candidate.GetCode() == 0 && candidate.Data != nil && candidate.Data.GetOrderId() == orderID {
			resp = candidate
			matchedBranchID = tokenRecord.BranchID
			break
		}
		if candidate != nil && candidate.GetCode() != 0 {
			lookupErr = fmt.Errorf("TikTok (%d): %s", candidate.GetCode(), candidate.GetMessage())
		}
	}

	if resp == nil || resp.Data == nil {
		message := "Data transaksi finansial untuk pesanan ini belum tersedia pada akun TikTok yang terhubung"
		if lookupErr != nil {
			message += ": " + lookupErr.Error()
		}
		c.JSON(http.StatusNotFound, gin.H{"error": message})
		return
	}

	settlementAmt, settlementProvided := parseOptionalTikTokAmount(resp.Data.GetSettlementAmount())
	feeAndTaxAmt, feeAndTaxProvided := parseOptionalTikTokFeeAmount(resp.Data.GetFeeAndTaxAmount())
	shippingCostAmt, _ := strconv.ParseFloat(resp.Data.GetShippingCostAmount(), 64)
	if settlementAmt == 0 && feeAndTaxAmt == 0 && len(resp.Data.SkuTransactions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "TikTok belum menyediakan rincian finansial untuk pesanan ini"})
		return
	}

	breakdownJSON := ""
	if resp.Data.SkuTransactions != nil && len(resp.Data.SkuTransactions) > 0 {
		if bBytes, bErr := json.Marshal(resp.Data.SkuTransactions); bErr == nil {
			breakdownJSON = string(bBytes)
		}
	}

	record := domain.FinanceTransactionRecord{
		OrderID:            orderID,
		Marketplace:        "tiktok",
		BranchID:           matchedBranchID,
		SettlementAmount:   settlementAmt,
		SettlementProvided: settlementProvided,
		FeeAndTaxAmount:    feeAndTaxAmt,
		FeeAndTaxProvided:  feeAndTaxProvided,
		ShippingCostAmount: shippingCostAmt,
		AdjustmentAmount:   0,
		FeeTaxBreakdown:    breakdownJSON,
		// Order detail has no statement payment status; leave payout status unchanged.
	}

	updatedIDs, saveErr := h.usecase.SaveFinanceTransactions([]domain.FinanceTransactionRecord{record})
	if saveErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan rincian finansial order: " + saveErr.Error()})
		return
	}
	if len(updatedIDs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan ini tidak ditemukan pada data marketplace lokal untuk akun TikTok yang dipilih"})
		return
	}
	var savedOrder domain.MarketplaceOrder
	q := h.db.Model(&domain.MarketplaceOrder{}).Where("marketplace = ? AND order_id = ?", "tiktok", orderID)
	if matchedBranchID != nil {
		q = q.Where("branch_id = ?", *matchedBranchID)
	}
	_ = q.Select("settlement_status").First(&savedOrder).Error

	message := fmt.Sprintf("Rincian fee pesanan %s diperbarui. Status dana cair mengikuti statement pembayaran.", orderID)
	if !settlementProvided {
		message = fmt.Sprintf("TikTok tidak mengirim nominal pencairan untuk pesanan %s. Nominal lama dipertahankan; periksa statement di Seller Center.", orderID)
	} else if !feeAndTaxProvided {
		message = fmt.Sprintf("TikTok tidak mengirim nominal fee untuk pesanan %s. Fee lama dipertahankan; periksa rincian di Seller Center.", orderID)
	} else if feeAndTaxAmt == 0 {
		message = fmt.Sprintf("TikTok mengirim fee Rp0 untuk pesanan %s. Cocokkan dengan rincian di Seller Center jika terasa janggal.", orderID)
	}
	c.JSON(http.StatusOK, gin.H{
		"message":                    message,
		"order_id":                   orderID,
		"settlement_amount":          settlementAmt,
		"settlement_amount_provided": settlementProvided,
		"fee_and_tax_amount":         feeAndTaxAmt,
		"fee_and_tax_provided":       feeAndTaxProvided,
		"shipping_cost_amount":       shippingCostAmt,
		"settlement_status":          savedOrder.SettlementStatus,
		"fee_tax_breakdown":          breakdownJSON,
	})
}
