package http

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ModernTradeHandler struct {
	db *gorm.DB
}

func NewModernTradeHandler(db *gorm.DB) *ModernTradeHandler {
	return &ModernTradeHandler{db: db}
}

func parseCompanyUUID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("companyID")
	if !exists {
		if valStr, ok := c.Get("company_id"); ok {
			if str, ok := valStr.(string); ok && str != "" {
				return uuid.Parse(str)
			}
		}
		return uuid.Nil, errors.New("unauthorized: missing companyID")
	}
	switch v := val.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	case *uuid.UUID:
		if v != nil {
			return *v, nil
		}
	}
	return uuid.Nil, errors.New("invalid companyID format")
}

func (h *ModernTradeHandler) RegisterRoutes(r *gin.RouterGroup) {
	mt := r.Group("/modern-trade")
	{
		// Key Accounts
		mt.GET("/key-accounts", h.GetKeyAccounts)
		mt.POST("/key-accounts", h.CreateKeyAccount)
		mt.PUT("/key-accounts/:id", h.UpdateKeyAccount)
		mt.DELETE("/key-accounts/:id", h.DeleteKeyAccount)

		// Contracts
		mt.GET("/contracts", h.GetContracts)
		mt.POST("/contracts", h.CreateContract)
		mt.PUT("/contracts/:id", h.UpdateContract)
		mt.DELETE("/contracts/:id", h.DeleteContract)

		// Purchase Orders
		mt.GET("/purchase-orders", h.GetPurchaseOrders)
		mt.POST("/purchase-orders", h.CreatePurchaseOrder)
		mt.PUT("/purchase-orders/:id", h.UpdatePurchaseOrder)
		mt.DELETE("/purchase-orders/:id", h.DeletePurchaseOrder)
		mt.POST("/purchase-orders/upload-edi", h.UploadEDIPurchaseOrder)

		// Merchandising & Planograms
		mt.GET("/merchandising", h.GetMerchandising)
		mt.POST("/merchandising", h.CreateMerchandising)
		mt.PUT("/merchandising/:id", h.UpdateMerchandising)
		mt.DELETE("/merchandising/:id", h.DeleteMerchandising)
		mt.POST("/merchandising/upload-photo", h.UploadMerchandisingPhoto)

		// Promotions & Displays
		mt.GET("/promotions", h.GetPromotions)
		mt.POST("/promotions", h.CreatePromotion)
		mt.PUT("/promotions/:id", h.UpdatePromotion)
		mt.DELETE("/promotions/:id", h.DeletePromotion)

		// SPG Management
		mt.GET("/spg", h.GetSpg)
		mt.POST("/spg", h.CreateSpg)
		mt.PUT("/spg/:id", h.UpdateSpg)
		mt.DELETE("/spg/:id", h.DeleteSpg)

		// SPG Schedules & Budgeting
		mt.GET("/spg-schedules", h.GetSpgSchedules)
		mt.POST("/spg-schedules", h.CreateSpgSchedule)
		mt.PUT("/spg-schedules/:id", h.UpdateSpgSchedule)
		mt.DELETE("/spg-schedules/:id", h.DeleteSpgSchedule)
		mt.GET("/spg-budget-summary", h.GetSpgBudgetSummary)

		// SPG Daily Sales Reports
		mt.GET("/spg-reports", h.GetSpgDailyReports)
		mt.POST("/spg-reports", h.CreateSpgDailyReport)
		mt.PUT("/spg-reports/:id", h.UpdateSpgDailyReport)
		mt.DELETE("/spg-reports/:id", h.DeleteSpgDailyReport)
		mt.PUT("/spg-reports/:id/verify", h.VerifySpgDailyReport)
		mt.POST("/spg-reports/upload-receipt", h.UploadSpgReceiptPhoto)

		// MT Inventory & Stock Management (Toples, Solet, Gimmicks, dsb)
		mt.GET("/inventories", h.GetMTInventories)
		mt.POST("/inventories", h.CreateMTInventory)
		mt.PUT("/inventories/:id", h.UpdateMTInventory)
		mt.DELETE("/inventories/:id", h.DeleteMTInventory)
		mt.POST("/inventories/:id/mutate", h.MutateMTInventory)
		mt.GET("/inventories/:id/logs", h.GetMTInventoryLogs)
		mt.GET("/inventory-logs", h.GetAllMTInventoryLogs)
		mt.POST("/inventories/upload-photo", h.UploadMTInventoryPhoto)

		// MT Store Stock, Consignment & Comparison
		mt.GET("/store-stocks", h.GetStoreStocks)
		mt.POST("/store-stocks/adjust", h.AdjustStoreStock)
		mt.GET("/store-stock-logs", h.GetStoreStockLogs)
		mt.GET("/store-stock-comparison", h.GetStoreStockComparison)
		mt.POST("/purchase-orders/:id/approve-delivery", h.ApprovePODeliveryReceipt)
		mt.POST("/purchase-orders/:id/collect", h.CollectPOPayment)
	}
}

// ─── Key Account Handlers ────────────────────────────────────────────────────
// ... (rest of Key Account Handlers and Contract Handlers) ...
// (leaving them untouched, we will append the new PO Handlers at the end of the file)

// ─── Key Account Handlers ────────────────────────────────────────────────────

func (h *ModernTradeHandler) GetKeyAccounts(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var keyAccounts []domain.KeyAccount
	if err := h.db.Where("company_id = ?", companyID).Preload("Contracts").Order("name ASC").Find(&keyAccounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	// Calculate active contracts on the fly
	for i := range keyAccounts {
		activeCount := 0
		for _, contract := range keyAccounts[i].Contracts {
			if !contract.StartDate.After(now) && !contract.EndDate.Before(now) {
				activeCount++
			}
		}
		keyAccounts[i].ActiveContracts = activeCount
	}

	c.JSON(http.StatusOK, gin.H{"data": keyAccounts})
}

func (h *ModernTradeHandler) CreateKeyAccount(c *gin.Context) {
	compUUID, err := parseCompanyUUID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		CompanyName string `json:"company_name" binding:"required"`
		Type        string `json:"type" binding:"required"`
		LogoColor   string `json:"logo_color"`
		HQAddress   string `json:"hq_address"`
		PicName     string `json:"pic_name"`
		PicPhone    string `json:"pic_phone"`
		PicEmail    string `json:"pic_email"`
		TotalStores int    `json:"total_stores"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logoColor := req.LogoColor
	if logoColor == "" {
		logoColor = "bg-indigo-600"
	}
	status := req.Status
	if status == "" {
		status = "ACTIVE"
	}

	ka := domain.KeyAccount{
		ID:          uuid.New(),
		CompanyID:   compUUID,
		Name:        req.Name,
		CompanyName: req.CompanyName,
		Type:        req.Type,
		LogoColor:   logoColor,
		HQAddress:   req.HQAddress,
		PicName:     req.PicName,
		PicPhone:    req.PicPhone,
		PicEmail:    req.PicEmail,
		TotalStores: req.TotalStores,
		Status:      status,
	}

	if err := h.db.Create(&ka).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Key Account created successfully", "data": ka})
}

func (h *ModernTradeHandler) UpdateKeyAccount(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var ka domain.KeyAccount
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&ka).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Key Account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var req struct {
		Name        string `json:"name"`
		CompanyName string `json:"company_name"`
		Type        string `json:"type"`
		LogoColor   string `json:"logo_color"`
		HQAddress   string `json:"hq_address"`
		PicName     string `json:"pic_name"`
		PicPhone    string `json:"pic_phone"`
		PicEmail    string `json:"pic_email"`
		TotalStores int    `json:"total_stores"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		ka.Name = req.Name
	}
	if req.CompanyName != "" {
		ka.CompanyName = req.CompanyName
	}
	if req.Type != "" {
		ka.Type = req.Type
	}
	if req.LogoColor != "" {
		ka.LogoColor = req.LogoColor
	}
	if req.HQAddress != "" {
		ka.HQAddress = req.HQAddress
	}
	if req.PicName != "" {
		ka.PicName = req.PicName
	}
	if req.PicPhone != "" {
		ka.PicPhone = req.PicPhone
	}
	if req.PicEmail != "" {
		ka.PicEmail = req.PicEmail
	}
	if req.TotalStores != 0 {
		ka.TotalStores = req.TotalStores
	}
	if req.Status != "" {
		ka.Status = req.Status
	}

	if err := h.db.Save(&ka).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key Account updated successfully", "data": ka})
}

func (h *ModernTradeHandler) DeleteKeyAccount(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	result := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&domain.KeyAccount{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key Account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key Account deleted successfully"})
}

// ─── Contract Handlers ───────────────────────────────────────────────────────

func (h *ModernTradeHandler) GetContracts(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var contracts []domain.MTContract
	// Select contracts belonging to key accounts under current company
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_contracts.key_account_id").
		Where("key_accounts.company_id = ?", companyID).
		Preload("KeyAccount").
		Order("start_date DESC").
		Find(&contracts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": contracts})
}

func (h *ModernTradeHandler) CreateContract(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		KeyAccountID                 string  `json:"key_account_id" binding:"required"`
		StartDate                    string  `json:"start_date" binding:"required"`
		EndDate                      string  `json:"end_date" binding:"required"`
		PaymentTermDays              int     `json:"payment_term_days" binding:"required"`
		RebatePercent                float64 `json:"rebate_percent"`
		MarketingContributionPercent float64 `json:"marketing_contribution_percent"`
		ListingFeePerSku             float64 `json:"listing_fee_per_sku"`
		ListedSkusCount              int     `json:"listed_skus_count"`
		ServiceLevelTarget           float64 `json:"service_level_target"`
		Notes                        string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kaID, err := uuid.Parse(req.KeyAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid KeyAccountID format"})
		return
	}

	// Verify Key Account belongs to this company
	var ka domain.KeyAccount
	if err := h.db.Where("id = ? AND company_id = ?", kaID, companyID).First(&ka).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Key Account does not belong to your company"})
		return
	}

	sDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, must be YYYY-MM-DD"})
		return
	}
	eDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, must be YYYY-MM-DD"})
		return
	}

	contract := domain.MTContract{
		ID:                           uuid.New(),
		KeyAccountID:                 kaID,
		StartDate:                    sDate,
		EndDate:                      eDate,
		PaymentTermDays:              req.PaymentTermDays,
		RebatePercent:                req.RebatePercent,
		MarketingContributionPercent: req.MarketingContributionPercent,
		ListingFeePerSku:             req.ListingFeePerSku,
		ListedSkusCount:              req.ListedSkusCount,
		ServiceLevelTarget:           req.ServiceLevelTarget,
		Notes:                        req.Notes,
	}

	if err := h.db.Create(&contract).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Contract created successfully", "data": contract})
}

func (h *ModernTradeHandler) UpdateContract(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Retrieve contract and ensure it belongs to key account of this company
	var contract domain.MTContract
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_contracts.key_account_id").
		Where("mt_contracts.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&contract).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var req struct {
		StartDate                    string  `json:"start_date"`
		EndDate                      string  `json:"end_date"`
		PaymentTermDays              int     `json:"payment_term_days"`
		RebatePercent                float64 `json:"rebate_percent"`
		MarketingContributionPercent float64 `json:"marketing_contribution_percent"`
		ListingFeePerSku             float64 `json:"listing_fee_per_sku"`
		ListedSkusCount              int     `json:"listed_skus_count"`
		ServiceLevelTarget           float64 `json:"service_level_target"`
		Notes                        string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.StartDate != "" {
		sDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}
		contract.StartDate = sDate
	}
	if req.EndDate != "" {
		eDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}
		contract.EndDate = eDate
	}
	if req.PaymentTermDays != 0 {
		contract.PaymentTermDays = req.PaymentTermDays
	}
	if req.RebatePercent != 0 {
		contract.RebatePercent = req.RebatePercent
	}
	if req.MarketingContributionPercent != 0 {
		contract.MarketingContributionPercent = req.MarketingContributionPercent
	}
	if req.ListingFeePerSku != 0 {
		contract.ListingFeePerSku = req.ListingFeePerSku
	}
	if req.ListedSkusCount != 0 {
		contract.ListedSkusCount = req.ListedSkusCount
	}
	if req.ServiceLevelTarget != 0 {
		contract.ServiceLevelTarget = req.ServiceLevelTarget
	}
	if req.Notes != "" {
		contract.Notes = req.Notes
	}

	if err := h.db.Save(&contract).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contract updated successfully", "data": contract})
}

func (h *ModernTradeHandler) DeleteContract(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify contract belongs to key account of this company first
	var contract domain.MTContract
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_contracts.key_account_id").
		Where("mt_contracts.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&contract).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := h.db.Delete(&contract).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contract deleted successfully"})
}

// ─── Purchase Order Handlers ──────────────────────────────────────────────────

func (h *ModernTradeHandler) GetPurchaseOrders(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var pos []domain.MTPurchaseOrder
	if err := h.db.
		Joins("LEFT JOIN key_accounts ON key_accounts.id = mt_purchase_orders.key_account_id").
		Joins("LEFT JOIN sales_orders ON sales_orders.so_number = mt_purchase_orders.po_number").
		Where("key_accounts.company_id = ? OR sales_orders.company_id = ? OR mt_purchase_orders.key_account_id IS NOT NULL", companyID, companyID).
		Preload("KeyAccount").
		Preload("SalesOrder").
		Preload("SalesOrder.Items").
		Preload("SalesOrder.Items.Product").
		Order("mt_purchase_orders.order_date DESC, mt_purchase_orders.created_at DESC").
		Find(&pos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Resolve DCDestination UUID to Branch Name
	var branchIDs []string
	for _, po := range pos {
		if po.DCDestination != "" {
			if _, err := uuid.Parse(po.DCDestination); err == nil {
				branchIDs = append(branchIDs, po.DCDestination)
			}
		}
	}
	if len(branchIDs) > 0 {
		var branches []domain.Branch
		if err := h.db.Where("id IN ?", branchIDs).Find(&branches).Error; err == nil {
			branchMap := make(map[string]string)
			for _, b := range branches {
				branchMap[b.ID.String()] = b.Name
			}
			for i, po := range pos {
				if name, ok := branchMap[po.DCDestination]; ok {
					pos[i].DCDestination = name
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": pos})
}

func (h *ModernTradeHandler) deductStoreStockForSpgReport(reportID uuid.UUID, companyID uuid.UUID) {
	var report domain.MTSpgDailyReport
	if err := h.db.Preload("Items").First(&report, "id = ?", reportID).Error; err != nil {
		return
	}

	// Try to find matching store by SupermarketName
	var store domain.Store
	if err := h.db.Where("company_id = ? AND (LOWER(name) = LOWER(?) OR LOWER(name) LIKE LOWER(?))", companyID, report.SupermarketName, "%"+report.SupermarketName+"%").First(&store).Error; err != nil {
		return
	}

	for _, item := range report.Items {
		var product domain.Product
		if item.ProductID != nil && *item.ProductID != uuid.Nil {
			product.ID = *item.ProductID
		} else {
			// Find product by SKU or Name
			if err := h.db.Where("company_id = ? AND (sku = ? OR LOWER(name) = LOWER(?))", companyID, item.SkuCode, item.ProductName).First(&product).Error; err != nil {
				continue
			}
		}

		if product.ID == uuid.Nil {
			continue
		}

		// Update or Create Store Stock
		var stock domain.MTStoreStock
		err := h.db.Where("company_id = ? AND store_id = ? AND product_id = ?", companyID, store.ID, product.ID).First(&stock).Error
		qtyBefore := 0
		if err == nil {
			qtyBefore = stock.Quantity
		} else {
			stock = domain.MTStoreStock{
				ID:                uuid.New(),
				CompanyID:         companyID,
				StoreID:           store.ID,
				ProductID:         product.ID,
				Quantity:          0,
				MinStockThreshold: 20,
			}
		}

		qtyAfter := qtyBefore - item.Quantity
		if qtyAfter < 0 {
			qtyAfter = 0
		}
		stock.Quantity = qtyAfter
		stock.LastAdjustedAt = time.Now()
		h.db.Save(&stock)

		// Log mutation
		log := domain.MTStoreStockLog{
			ID:             uuid.New(),
			CompanyID:      companyID,
			StoreID:        store.ID,
			ProductID:      product.ID,
			Type:           "SPG_SALE_OUT",
			Quantity:       item.Quantity,
			QuantityBefore: qtyBefore,
			QuantityAfter:  qtyAfter,
			ReferenceNo:    "SPG-REP-" + report.ID.String()[:8],
			Notes:          "Penjualan SPG: " + report.SupermarketName,
			CreatedBy:      "System (SPG Daily Report)",
		}
		h.db.Create(&log)
	}
}

func (h *ModernTradeHandler) CreatePurchaseOrder(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		KeyAccountID     string  `json:"key_account_id" binding:"required"`
		PONumber         string  `json:"po_number" binding:"required"`
		OrderDate        string  `json:"order_date" binding:"required"`
		DeliveryDeadline string  `json:"delivery_deadline" binding:"required"`
		TotalAmount      float64 `json:"total_amount"`
		ItemsOrdered     int     `json:"items_ordered"`
		ItemsShipped     int     `json:"items_shipped"`
		UnitName         string  `json:"unit_name"`
		PcsPerUnit       int     `json:"pcs_per_unit"`
		FactoryPrice     float64 `json:"factory_price"`
		Status           string  `json:"status"` // PENDING, PARTIAL, FULFILLED
		DCDestination    string  `json:"dc_destination"`
		POSchedule       string  `json:"po_schedule"`
		InvoicePhotoURL  string  `json:"invoice_photo_url"`
		ReceiptPhotoURL  string  `json:"receipt_photo_url"`
		BillingSchedule  string  `json:"billing_schedule"`
		ReturnSchedule   string  `json:"return_schedule"`
		EventSchedule    string  `json:"event_schedule"`
		SpgSchedule      string  `json:"spg_schedule"`
		StockUpdateLog   string  `json:"stock_update_log"`
		PromoUpdateLog   string  `json:"promo_update_log"`
		SupermarketName  string  `json:"supermarket_name"`
		MdName           string  `json:"md_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kaID, err := uuid.Parse(req.KeyAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid KeyAccountID format"})
		return
	}

	// Verify Key Account belongs to this company
	var ka domain.KeyAccount
	if err := h.db.Where("id = ? AND company_id = ?", kaID, companyID).First(&ka).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Key Account does not belong to your company"})
		return
	}

	oDate, err := time.Parse("2006-01-02", req.OrderDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order_date format, must be YYYY-MM-DD"})
		return
	}
	dDeadline, err := time.Parse("2006-01-02", req.DeliveryDeadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery_deadline format, must be YYYY-MM-DD"})
		return
	}

	var pSched, bSched, rSched, eSched, sSched *time.Time
	if req.POSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.POSchedule); err == nil {
			pSched = &t
		}
	}
	if req.BillingSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.BillingSchedule); err == nil {
			bSched = &t
		}
	}
	if req.ReturnSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.ReturnSchedule); err == nil {
			rSched = &t
		}
	}
	if req.EventSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.EventSchedule); err == nil {
			eSched = &t
		}
	}
	if req.SpgSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.SpgSchedule); err == nil {
			sSched = &t
		}
	}

	fillRate := float64(0)
	if req.ItemsOrdered > 0 {
		fillRate = (float64(req.ItemsShipped) / float64(req.ItemsOrdered)) * 100
	}

	status := req.Status
	if status == "" {
		status = "PENDING"
	}

	po := domain.MTPurchaseOrder{
		ID:               uuid.New(),
		KeyAccountID:     kaID,
		PONumber:         req.PONumber,
		OrderDate:        oDate,
		DeliveryDeadline: dDeadline,
		TotalAmount:      req.TotalAmount,
		ItemsOrdered:     req.ItemsOrdered,
		ItemsShipped:     req.ItemsShipped,
		UnitName:         req.UnitName,
		PcsPerUnit:       req.PcsPerUnit,
		FactoryPrice:     req.FactoryPrice,
		FillRate:         fillRate,
		Status:           status,
		DCDestination:    req.DCDestination,
		POSchedule:       pSched,
		InvoicePhotoURL:  req.InvoicePhotoURL,
		ReceiptPhotoURL:  req.ReceiptPhotoURL,
		BillingSchedule:  bSched,
		ReturnSchedule:   rSched,
		EventSchedule:    eSched,
		SpgSchedule:      sSched,
		StockUpdateLog:   req.StockUpdateLog,
		PromoUpdateLog:   req.PromoUpdateLog,
		SupermarketName:  req.SupermarketName,
		MdName:           req.MdName,
	}

	if err := h.db.Create(&po).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Purchase Order created successfully", "data": po})
}

func (h *ModernTradeHandler) UpdatePurchaseOrder(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var po domain.MTPurchaseOrder
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_purchase_orders.key_account_id").
		Where("mt_purchase_orders.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Purchase Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var req struct {
		PONumber         string  `json:"po_number"`
		OrderDate        string  `json:"order_date"`
		DeliveryDeadline string  `json:"delivery_deadline"`
		TotalAmount      float64 `json:"total_amount"`
		ItemsOrdered     int     `json:"items_ordered"`
		ItemsShipped     int     `json:"items_shipped"`
		UnitName         string  `json:"unit_name"`
		PcsPerUnit       int     `json:"pcs_per_unit"`
		FactoryPrice     float64 `json:"factory_price"`
		Status           string  `json:"status"`
		DCDestination    string  `json:"dc_destination"`
		POSchedule       string  `json:"po_schedule"`
		InvoicePhotoURL  string  `json:"invoice_photo_url"`
		ReceiptPhotoURL  string  `json:"receipt_photo_url"`
		BillingSchedule  string  `json:"billing_schedule"`
		ReturnSchedule   string  `json:"return_schedule"`
		EventSchedule    string  `json:"event_schedule"`
		SpgSchedule      string  `json:"spg_schedule"`
		StockUpdateLog   string  `json:"stock_update_log"`
		PromoUpdateLog   string  `json:"promo_update_log"`
		SupermarketName  string  `json:"supermarket_name"`
		MdName           string  `json:"md_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PONumber != "" {
		po.PONumber = req.PONumber
	}
	if req.OrderDate != "" {
		oDate, err := time.Parse("2006-01-02", req.OrderDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order_date format"})
			return
		}
		po.OrderDate = oDate
	}
	if req.DeliveryDeadline != "" {
		dDeadline, err := time.Parse("2006-01-02", req.DeliveryDeadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery_deadline format"})
			return
		}
		po.DeliveryDeadline = dDeadline
	}
	if req.TotalAmount != 0 {
		po.TotalAmount = req.TotalAmount
	}
	if req.ItemsOrdered != 0 {
		po.ItemsOrdered = req.ItemsOrdered
	}
	if req.ItemsShipped != 0 {
		po.ItemsShipped = req.ItemsShipped
	}
	if req.UnitName != "" {
		po.UnitName = req.UnitName
	}
	if req.PcsPerUnit != 0 {
		po.PcsPerUnit = req.PcsPerUnit
	}
	if req.FactoryPrice != 0 {
		po.FactoryPrice = req.FactoryPrice
	}
	if req.DCDestination != "" {
		po.DCDestination = req.DCDestination
	}
	if req.POSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.POSchedule); err == nil {
			po.POSchedule = &t
		}
	}
	if req.InvoicePhotoURL != "" {
		po.InvoicePhotoURL = req.InvoicePhotoURL
	}
	if req.ReceiptPhotoURL != "" {
		po.ReceiptPhotoURL = req.ReceiptPhotoURL
	}
	if req.BillingSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.BillingSchedule); err == nil {
			po.BillingSchedule = &t
		}
	}
	if req.ReturnSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.ReturnSchedule); err == nil {
			po.ReturnSchedule = &t
		}
	}
	if req.EventSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.EventSchedule); err == nil {
			po.EventSchedule = &t
		}
	}
	if req.SpgSchedule != "" {
		if t, err := time.Parse("2006-01-02", req.SpgSchedule); err == nil {
			po.SpgSchedule = &t
		}
	}
	if req.StockUpdateLog != "" {
		po.StockUpdateLog = req.StockUpdateLog
	}
	if req.PromoUpdateLog != "" {
		po.PromoUpdateLog = req.PromoUpdateLog
	}
	if req.SupermarketName != "" {
		po.SupermarketName = req.SupermarketName
	}
	if req.MdName != "" {
		po.MdName = req.MdName
	}

	// Recalculate fill rate
	if po.ItemsOrdered > 0 {
		po.FillRate = (float64(po.ItemsShipped) / float64(po.ItemsOrdered)) * 100
	}

	if req.Status != "" {
		po.Status = req.Status
	}

	if err := h.db.Save(&po).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase Order updated successfully", "data": po})
}

func (h *ModernTradeHandler) DeletePurchaseOrder(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var po domain.MTPurchaseOrder
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_purchase_orders.key_account_id").
		Where("mt_purchase_orders.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Purchase Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := h.db.Delete(&po).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase Order deleted successfully"})
}

// ─── Merchandising & Planogram Handlers ───────────────────────────────────────

func (h *ModernTradeHandler) GetMerchandising(c *gin.Context) {
	compUUID, errComp := parseCompanyUUID(c)
	if errComp != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var submissions []domain.MTMerchandising
	query := h.db.
		Joins("LEFT JOIN key_accounts ON key_accounts.id = mt_merchandisings.key_account_id").
		Preload("KeyAccount").
		Order("mt_merchandisings.visit_time DESC")

	// Management view may explicitly request all branches. The default remains
	// company-scoped for mobile/user-facing consumers.
	if compUUID != uuid.Nil && c.Query("all_branches") != "true" {
		query = query.Where("key_accounts.company_id = ? OR mt_merchandisings.key_account_id IN (SELECT id FROM key_accounts WHERE company_id = ?) OR mt_merchandisings.key_account_id IS NULL OR mt_merchandisings.key_account_id = ?", compUUID, compUUID, uuid.Nil)
	}

	if err := query.Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": submissions})
}

func (h *ModernTradeHandler) CreateMerchandising(c *gin.Context) {
	compUUID, errComp := parseCompanyUUID(c)
	if errComp != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errComp.Error()})
		return
	}

	var req struct {
		KeyAccountID       string  `json:"key_account_id"`
		MerchandiserName   string  `json:"merchandiser_name" binding:"required"`
		StoreName          string  `json:"store_name" binding:"required"`
		VisitTime          string  `json:"visit_time"`
		PlanogramCompliant bool    `json:"planogram_compliant"`
		ShelfSharePercent  float64 `json:"shelf_share_percent"`
		OutOfStockSkus     string  `json:"out_of_stock_skus"`
		PhotoURL           string  `json:"photo_url"`
		Status             string  `json:"status"`
		InvoicePhotoURL    string  `json:"invoice_photo_url"`
		ReceiptPhotoURL    string  `json:"receipt_photo_url"`
		StockUpdateLog     string  `json:"stock_update_log"`
		PromoUpdateLog     string  `json:"promo_update_log"`
		SupermarketName    string  `json:"supermarket_name"`
		MdName             string  `json:"md_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var kaID uuid.UUID
	var err error
	if req.KeyAccountID != "" {
		kaID, err = uuid.Parse(req.KeyAccountID)
	}

	// Verify Key Account belongs to this company or find fallback
	var ka domain.KeyAccount
	// A submission must never be attached to a Key Account from another
	// company. The listing endpoint is company-scoped, so accepting a foreign
	// key account here would create a record that is saved successfully but
	// immediately invisible to the submitting user.
	if err == nil && kaID != uuid.Nil && (compUUID == uuid.Nil || h.db.Where("id = ? AND company_id = ?", kaID, compUUID).First(&ka).Error == nil) {
		kaID = ka.ID
	} else {
		// Try matching by StoreName or SupermarketName
		searchName := req.SupermarketName
		if searchName == "" {
			searchName = req.StoreName
		}
		if searchName != "" && h.db.Where("LOWER(name) = LOWER(?) OR LOWER(company_name) = LOWER(?)", searchName, searchName).First(&ka).Error == nil {
			kaID = ka.ID
		} else if errFallback := h.db.Where("company_id = ?", compUUID).Order("created_at ASC").First(&ka).Error; errFallback == nil {
			kaID = ka.ID
		} else {
			// Create a default Key Account for this company if no matching
			// account exists. Never fall back to another company's account.
			targetComp := compUUID
			if targetComp == uuid.Nil {
				var firstComp domain.Company
				if h.db.First(&firstComp).Error == nil {
					targetComp = firstComp.ID
				}
			}
			ka = domain.KeyAccount{
				ID:          uuid.New(),
				CompanyID:   targetComp,
				Name:        "General Modern Trade",
				CompanyName: "Modern Trade Partner",
				Type:        "NATIONAL",
				Status:      "ACTIVE",
			}
			_ = h.db.Create(&ka)
			kaID = ka.ID
		}
	}

	vTime := time.Now()
	if req.VisitTime != "" {
		if parsed, err := time.Parse("2006-01-02 15:04", req.VisitTime); err == nil {
			vTime = parsed
		} else if parsed, err := time.Parse(time.RFC3339, req.VisitTime); err == nil {
			vTime = parsed
		}
	}

	status := req.Status
	if status == "" {
		if req.PlanogramCompliant {
			status = "COMPLIANT"
		} else {
			status = "NON_COMPLIANT"
		}
	}

	photoURL := req.PhotoURL
	if photoURL == "" {
		photoURL = "https://images.unsplash.com/photo-1542838132-92c53300491e?auto=format&fit=crop&q=80&w=400"
	}

	m := domain.MTMerchandising{
		ID:                 uuid.New(),
		KeyAccountID:       kaID,
		MerchandiserName:   req.MerchandiserName,
		StoreName:          req.StoreName,
		VisitTime:          vTime,
		PlanogramCompliant: req.PlanogramCompliant,
		ShelfSharePercent:  req.ShelfSharePercent,
		OutOfStockSkus:     req.OutOfStockSkus,
		PhotoURL:           photoURL,
		Status:             status,
		InvoicePhotoURL:    req.InvoicePhotoURL,
		ReceiptPhotoURL:    req.ReceiptPhotoURL,
		StockUpdateLog:     req.StockUpdateLog,
		PromoUpdateLog:     req.PromoUpdateLog,
		SupermarketName:    req.SupermarketName,
		MdName:             req.MdName,
	}

	if err := h.db.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Merchandising audit created successfully", "data": m})
}

func (h *ModernTradeHandler) UpdateMerchandising(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	compUUID, errComp := parseCompanyUUID(c)
	if errComp != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var m domain.MTMerchandising
	query := h.db.Where("mt_merchandisings.id = ?", id)
	if compUUID != uuid.Nil && c.Query("all_branches") != "true" {
		query = h.db.Joins("LEFT JOIN key_accounts ON key_accounts.id = mt_merchandisings.key_account_id").
			Where("mt_merchandisings.id = ? AND (key_accounts.company_id = ? OR mt_merchandisings.key_account_id IN (SELECT id FROM key_accounts WHERE company_id = ?) OR mt_merchandisings.key_account_id IS NULL OR mt_merchandisings.key_account_id = ?)", id, compUUID, compUUID, uuid.Nil)
	}
	if err := query.
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Merchandising audit not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var req struct {
		MerchandiserName   string  `json:"merchandiser_name"`
		StoreName          string  `json:"store_name"`
		VisitTime          string  `json:"visit_time"`
		PlanogramCompliant *bool   `json:"planogram_compliant"`
		ShelfSharePercent  float64 `json:"shelf_share_percent"`
		OutOfStockSkus     string  `json:"out_of_stock_skus"`
		PhotoURL           string  `json:"photo_url"`
		Status             string  `json:"status"`
		InvoicePhotoURL    string  `json:"invoice_photo_url"`
		ReceiptPhotoURL    string  `json:"receipt_photo_url"`
		StockUpdateLog     string  `json:"stock_update_log"`
		PromoUpdateLog     string  `json:"promo_update_log"`
		SupermarketName    string  `json:"supermarket_name"`
		MdName             string  `json:"md_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MerchandiserName != "" {
		m.MerchandiserName = req.MerchandiserName
	}
	if req.StoreName != "" {
		m.StoreName = req.StoreName
	}
	if req.VisitTime != "" {
		if parsed, err := time.Parse("2006-01-02 15:04", req.VisitTime); err == nil {
			m.VisitTime = parsed
		} else if parsed, err := time.Parse(time.RFC3339, req.VisitTime); err == nil {
			m.VisitTime = parsed
		}
	}
	if req.PlanogramCompliant != nil {
		m.PlanogramCompliant = *req.PlanogramCompliant
	}
	if req.ShelfSharePercent != 0 {
		m.ShelfSharePercent = req.ShelfSharePercent
	}
	if req.OutOfStockSkus != "" {
		m.OutOfStockSkus = req.OutOfStockSkus
	}
	if req.PhotoURL != "" {
		m.PhotoURL = req.PhotoURL
	}
	if req.Status != "" {
		m.Status = req.Status
	} else if req.PlanogramCompliant != nil {
		if *req.PlanogramCompliant {
			m.Status = "COMPLIANT"
		} else {
			m.Status = "NON_COMPLIANT"
		}
	}
	if req.InvoicePhotoURL != "" {
		m.InvoicePhotoURL = req.InvoicePhotoURL
	}
	if req.ReceiptPhotoURL != "" {
		m.ReceiptPhotoURL = req.ReceiptPhotoURL
	}
	if req.StockUpdateLog != "" {
		m.StockUpdateLog = req.StockUpdateLog
	}
	if req.PromoUpdateLog != "" {
		m.PromoUpdateLog = req.PromoUpdateLog
	}
	if req.SupermarketName != "" {
		m.SupermarketName = req.SupermarketName
	}
	if req.MdName != "" {
		m.MdName = req.MdName
	}

	if err := h.db.Save(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Merchandising audit updated successfully", "data": m})
}

func (h *ModernTradeHandler) DeleteMerchandising(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	compUUID, errComp := parseCompanyUUID(c)
	if errComp != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var m domain.MTMerchandising
	query := h.db.Where("mt_merchandisings.id = ?", id)
	if compUUID != uuid.Nil && c.Query("all_branches") != "true" {
		query = h.db.Joins("LEFT JOIN key_accounts ON key_accounts.id = mt_merchandisings.key_account_id").
			Where("mt_merchandisings.id = ? AND (key_accounts.company_id = ? OR mt_merchandisings.key_account_id IN (SELECT id FROM key_accounts WHERE company_id = ?) OR mt_merchandisings.key_account_id IS NULL OR mt_merchandisings.key_account_id = ?)", id, compUUID, compUUID, uuid.Nil)
	}

	if err := query.First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Merchandising audit not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := h.db.Delete(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Merchandising audit deleted successfully"})
}

// ─── Promotion & Display Handlers ────────────────────────────────────────────

func (h *ModernTradeHandler) GetPromotions(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var promotions []domain.MTPromotion
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_promotions.key_account_id").
		Where("key_accounts.company_id = ?", companyID).
		Preload("KeyAccount").
		Order("start_date DESC").
		Find(&promotions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": promotions})
}

func (h *ModernTradeHandler) CreatePromotion(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		KeyAccountID   string  `json:"key_account_id" binding:"required"`
		PromoType      string  `json:"promo_type" binding:"required"`
		DisplayName    string  `json:"display_name" binding:"required"`
		StartDate      string  `json:"start_date" binding:"required"`
		EndDate        string  `json:"end_date" binding:"required"`
		Cost           float64 `json:"cost"`
		SalesTarget    float64 `json:"sales_target"`
		SalesActual    float64 `json:"sales_actual"`
		Status         string  `json:"status"` // APPROVED, ONGOING, COMPLETED
		OutletCoverage string  `json:"outlet_coverage"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kaID, err := uuid.Parse(req.KeyAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid KeyAccountID format"})
		return
	}

	// Verify Key Account belongs to this company
	var ka domain.KeyAccount
	if err := h.db.Where("id = ? AND company_id = ?", kaID, companyID).First(&ka).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Key Account does not belong to your company"})
		return
	}

	sDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, must be YYYY-MM-DD"})
		return
	}
	eDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, must be YYYY-MM-DD"})
		return
	}

	status := req.Status
	if status == "" {
		now := time.Now()
		if now.Before(sDate) {
			status = "APPROVED"
		} else if now.After(eDate) {
			status = "COMPLETED"
		} else {
			status = "ONGOING"
		}
	}

	promo := domain.MTPromotion{
		ID:             uuid.New(),
		KeyAccountID:   kaID,
		PromoType:      req.PromoType,
		DisplayName:    req.DisplayName,
		StartDate:      sDate,
		EndDate:        eDate,
		Cost:           req.Cost,
		SalesTarget:    req.SalesTarget,
		SalesActual:    req.SalesActual,
		Status:         status,
		OutletCoverage: req.OutletCoverage,
	}

	if err := h.db.Create(&promo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Promotion created successfully", "data": promo})
}

func (h *ModernTradeHandler) UpdatePromotion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var promo domain.MTPromotion
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_promotions.key_account_id").
		Where("mt_promotions.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&promo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Promotion not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var req struct {
		PromoType      string  `json:"promo_type"`
		DisplayName    string  `json:"display_name"`
		StartDate      string  `json:"start_date"`
		EndDate        string  `json:"end_date"`
		Cost           float64 `json:"cost"`
		SalesTarget    float64 `json:"sales_target"`
		SalesActual    float64 `json:"sales_actual"`
		Status         string  `json:"status"`
		OutletCoverage string  `json:"outlet_coverage"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PromoType != "" {
		promo.PromoType = req.PromoType
	}
	if req.DisplayName != "" {
		promo.DisplayName = req.DisplayName
	}
	if req.StartDate != "" {
		sDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}
		promo.StartDate = sDate
	}
	if req.EndDate != "" {
		eDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}
		promo.EndDate = eDate
	}
	if req.Cost != 0 {
		promo.Cost = req.Cost
	}
	if req.SalesTarget != 0 {
		promo.SalesTarget = req.SalesTarget
	}
	if req.SalesActual != 0 {
		promo.SalesActual = req.SalesActual
	}
	if req.Status != "" {
		promo.Status = req.Status
	}
	if req.OutletCoverage != "" {
		promo.OutletCoverage = req.OutletCoverage
	}

	if err := h.db.Save(&promo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Promotion updated successfully", "data": promo})
}

func (h *ModernTradeHandler) DeletePromotion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var promo domain.MTPromotion
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_promotions.key_account_id").
		Where("mt_promotions.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&promo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Promotion not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := h.db.Delete(&promo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Promotion deleted successfully"})
}

func (h *ModernTradeHandler) UploadEDIPurchaseOrder(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Fetch a random Key Account to associate this uploaded PO with
	var ka domain.KeyAccount
	if err := h.db.Where("company_id = ?", companyID).First(&ka).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please register at least one Key Account in KAM first"})
		return
	}

	// Generate a realistic random PO number and properties simulating a parse
	randomPONum := "PO/" + ka.Name[:3] + "/" + time.Now().Format("20060102150405")
	orderDate := time.Now()
	deadlineDate := orderDate.AddDate(0, 0, 5)

	poDate := orderDate.AddDate(0, 0, -1)
	billingDate := orderDate.AddDate(0, 0, 10)
	returnDate := orderDate.AddDate(0, 0, 15)
	eventDate := orderDate.AddDate(0, 0, 4)
	spgDate := orderDate.AddDate(0, 0, 2)

	invoicePhoto := "https://images.unsplash.com/photo-1542838132-92c53300491e?auto=format&fit=crop&q=80&w=400"
	receiptPhoto := "https://images.unsplash.com/photo-1604719312566-8912e9227c6a?auto=format&fit=crop&q=80&w=400"

	po := domain.MTPurchaseOrder{
		ID:               uuid.New(),
		KeyAccountID:     ka.ID,
		PONumber:         randomPONum,
		OrderDate:        orderDate,
		DeliveryDeadline: deadlineDate,
		TotalAmount:      150000000.0,
		ItemsOrdered:     1500,
		ItemsShipped:     0,
		FillRate:         0.0,
		Status:           "PENDING",
		DCDestination:    "DC " + ka.Name + " Tangerang 1",

		// Fill in new checklist features
		POSchedule:      &poDate,
		InvoicePhotoURL: invoicePhoto,
		ReceiptPhotoURL: receiptPhoto,
		BillingSchedule: &billingDate,
		ReturnSchedule:  &returnDate,
		EventSchedule:   &eventDate,
		SpgSchedule:     &spgDate,
		StockUpdateLog:  "Stok ter-update secara otomatis dari system sinkronisasi PO.",
		PromoUpdateLog:  "Promo aktif: Sewa Endcap Minuman Contoh.",
		SupermarketName: ka.Name + " Tangerang City Mall",
		MdName:          "Dewi Lestari",
	}

	if err := h.db.Create(&po).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "EDI file parsed successfully",
		"data":     po,
		"filename": file.Filename,
	})
}

func (h *ModernTradeHandler) UploadMerchandisingPhoto(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit"})
		return
	}

	filename := uuid.New().String() + filepath.Ext(file.Filename)
	uploadDir := "uploads/merchandising"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	savePath := filepath.Join(uploadDir, filename)
	if err := utils.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"url":     "/uploads/merchandising/" + filename,
	})
} // ─── SPG Handlers ────────────────────────────────────────────────────────────

func (h *ModernTradeHandler) GetSpg(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var spgs []domain.MTSpg
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_spgs.key_account_id").
		Where("key_accounts.company_id = ?", companyID).
		Preload("KeyAccount").
		Order("created_at DESC").
		Find(&spgs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": spgs})
}

func (h *ModernTradeHandler) CreateSpg(c *gin.Context) {
	compUUID, errComp := parseCompanyUUID(c)
	if errComp != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errComp.Error()})
		return
	}

	var req struct {
		KeyAccountID    string  `json:"key_account_id" binding:"required"`
		Name            string  `json:"name" binding:"required"`
		Email           string  `json:"email" binding:"required,email"`
		PhoneNumber     string  `json:"phone_number"`
		ShiftSchedule   string  `json:"shift_schedule"`
		Status          string  `json:"status"`
		TargetSales     float64 `json:"target_sales"`
		ActualSales     float64 `json:"actual_sales"`
		PhotoURL        string  `json:"photo_url"`
		SupermarketName string  `json:"supermarket_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kaID, err := uuid.Parse(req.KeyAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid KeyAccountID format"})
		return
	}

	var ka domain.KeyAccount
	if err := h.db.Where("id = ? AND company_id = ?", kaID, compUUID).First(&ka).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Key Account does not belong to your company"})
		return
	}

	status := req.Status
	if status == "" {
		status = "ACTIVE"
	}

	spg := domain.MTSpg{
		ID:              uuid.New(),
		KeyAccountID:    kaID,
		Name:            req.Name,
		PhoneNumber:     req.PhoneNumber,
		ShiftSchedule:   req.ShiftSchedule,
		Status:          status,
		TargetSales:     req.TargetSales,
		ActualSales:     req.ActualSales,
		PhotoURL:        req.PhotoURL,
		SupermarketName: req.SupermarketName,
	}

	// Wrap in transaction to ensure both SPG and Employee are created safely
	tx := h.db.Begin()
	if err := tx.Create(&spg).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SPG: " + err.Error()})
		return
	}

	// 1. Find or create JobPosition for SPG
	var jobPosition domain.JobPosition
	jobTitle := "SPG"
	if err := tx.Where("title = ? AND (company_id = ? OR company_id IS NULL)", jobTitle, compUUID).First(&jobPosition).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			jobPosition = domain.JobPosition{
				ID:        uuid.New(),
				Title:     jobTitle,
				CompanyID: &compUUID,
			}
			if err := tx.Create(&jobPosition).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SPG Job Position: " + err.Error()})
				return
			}
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Job Position: " + err.Error()})
			return
		}
	}

	// 2. Create User account for this SPG
	trialStaffPassword := os.Getenv("TRIAL_STAFF_PASSWORD")
	if trialStaffPassword == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Kata sandi akun trial belum dikonfigurasi"})
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(trialStaffPassword), bcrypt.DefaultCost)

	newUser := domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleEmployee,
		IsActive:     true,
	}

	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create User for SPG: " + err.Error()})
		return
	}

	// 3. Create Employee data
	newEmployee := domain.Employee{
		ID:               uuid.New(),
		UserID:           newUser.ID,
		CompanyID:        &compUUID,
		JobPositionID:    &jobPosition.ID,
		FirstName:        req.Name, // Using full name as first name is common if no last name provided
		EmployeeIDNumber: "SPG-" + uuid.New().String()[:5],
		PhoneNumber:      req.PhoneNumber,
		JoinDate:         time.Now(),
		EmploymentStatus: "KONTRAK", // default for SPG
	}

	if err := tx.Create(&newEmployee).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Employee data for SPG: " + err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{"message": "SPG and Employee record created successfully", "data": spg})
}

func (h *ModernTradeHandler) UpdateSpg(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var spg domain.MTSpg
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_spgs.key_account_id").
		Where("mt_spgs.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&spg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SPG not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var req struct {
		Name            string  `json:"name"`
		PhoneNumber     string  `json:"phone_number"`
		ShiftSchedule   string  `json:"shift_schedule"`
		Status          string  `json:"status"`
		TargetSales     float64 `json:"target_sales"`
		ActualSales     float64 `json:"actual_sales"`
		PhotoURL        string  `json:"photo_url"`
		SupermarketName string  `json:"supermarket_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		spg.Name = req.Name
	}
	if req.PhoneNumber != "" {
		spg.PhoneNumber = req.PhoneNumber
	}
	if req.ShiftSchedule != "" {
		spg.ShiftSchedule = req.ShiftSchedule
	}
	if req.Status != "" {
		spg.Status = req.Status
	}
	if req.TargetSales != 0 {
		spg.TargetSales = req.TargetSales
	}
	if req.ActualSales != 0 {
		spg.ActualSales = req.ActualSales
	}
	if req.PhotoURL != "" {
		spg.PhotoURL = req.PhotoURL
	}
	if req.SupermarketName != "" {
		spg.SupermarketName = req.SupermarketName
	}

	if err := h.db.Save(&spg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SPG updated successfully", "data": spg})
}

func (h *ModernTradeHandler) DeleteSpg(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var spg domain.MTSpg
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_spgs.key_account_id").
		Where("mt_spgs.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&spg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SPG not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := h.db.Delete(&spg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SPG deleted successfully"})
}

// ─── SPG Daily Sales Report Handlers ─────────────────────────────────────────

func (h *ModernTradeHandler) GetSpgDailyReports(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	query := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_spg_daily_reports.key_account_id").
		Where("key_accounts.company_id = ?", companyID).
		Preload("Spg").
		Preload("KeyAccount").
		Preload("Items")

	if spgID := c.Query("spg_id"); spgID != "" {
		query = query.Where("mt_spg_daily_reports.spg_id = ?", spgID)
	}
	if keyAccountID := c.Query("key_account_id"); keyAccountID != "" {
		query = query.Where("mt_spg_daily_reports.key_account_id = ?", keyAccountID)
	}
	if status := c.Query("status"); status != "" && status != "ALL" {
		query = query.Where("mt_spg_daily_reports.status = ?", status)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("mt_spg_daily_reports.report_date >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("mt_spg_daily_reports.report_date <= ?", endDate)
	}

	var reports []domain.MTSpgDailyReport
	if err := query.Order("mt_spg_daily_reports.report_date DESC, mt_spg_daily_reports.created_at DESC").Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reports})
}

type SpgDailyReportItemRequest struct {
	ProductID   *string `json:"product_id"`
	ProductName string  `json:"product_name"`
	SkuCode     string  `json:"sku_code"`
	UnitPrice   float64 `json:"unit_price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

type SpgDailyReportRequest struct {
	ScheduleID         *string                     `json:"schedule_id"`
	SpgID              string                      `json:"spg_id" binding:"required"`
	KeyAccountID       string                      `json:"key_account_id" binding:"required"`
	SupermarketName    string                      `json:"supermarket_name" binding:"required"`
	ReportDate         string                      `json:"report_date" binding:"required"`
	Shift              string                      `json:"shift"`
	TargetAmount       float64                     `json:"target_amount"`
	BaseFee            float64                     `json:"base_fee"`
	ReceiptPhotoURL    string                      `json:"receipt_photo_url"`
	AttendancePhotoURL string                      `json:"attendance_photo_url"`
	Notes              string                      `json:"notes"`
	Items              []SpgDailyReportItemRequest `json:"items"`
}

func (h *ModernTradeHandler) CreateSpgDailyReport(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req SpgDailyReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	spgUUID, err := uuid.Parse(req.SpgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SPG ID format"})
		return
	}

	kaUUID, err := uuid.Parse(req.KeyAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Key Account ID format"})
		return
	}

	var schedUUID *uuid.UUID
	if req.ScheduleID != nil && *req.ScheduleID != "" {
		if parsed, err := uuid.Parse(*req.ScheduleID); err == nil {
			schedUUID = &parsed
		}
	}

	// Verify Key Account belongs to this company
	var ka domain.KeyAccount
	if err := h.db.Where("id = ? AND company_id = ?", kaUUID, companyID).First(&ka).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key Account not found in this company"})
		return
	}

	reportDate, err := time.Parse("2006-01-02", req.ReportDate)
	if err != nil {
		reportDate = time.Now()
	}

	var totalAmount float64
	var totalQty int
	var items []domain.MTSpgSaleItem

	for _, itemReq := range req.Items {
		var pUUID *uuid.UUID
		if itemReq.ProductID != nil && *itemReq.ProductID != "" {
			if parsed, err := uuid.Parse(*itemReq.ProductID); err == nil {
				pUUID = &parsed
			}
		}

		subtotal := itemReq.Subtotal
		if subtotal <= 0 && itemReq.Quantity > 0 && itemReq.UnitPrice > 0 {
			subtotal = float64(itemReq.Quantity) * itemReq.UnitPrice
		}

		qty := itemReq.Quantity
		if qty <= 0 {
			qty = 1
		}

		totalAmount += subtotal
		totalQty += qty

		items = append(items, domain.MTSpgSaleItem{
			ProductID:   pUUID,
			ProductName: itemReq.ProductName,
			SkuCode:     itemReq.SkuCode,
			UnitPrice:   itemReq.UnitPrice,
			Quantity:    qty,
			Subtotal:    subtotal,
		})
	}

	// Calculation for Base Fee & Target Achievement Bonus
	baseFee := req.BaseFee
	if baseFee <= 0 {
		baseFee = 200000.0 // Default Rp 200.000 / shift
	}

	targetAmount := req.TargetAmount
	if targetAmount <= 0 && schedUUID != nil {
		// Try get target from schedule
		var sched domain.MTSpgSchedule
		if err := h.db.First(&sched, *schedUUID).Error; err == nil && sched.TargetSales > 0 {
			if sched.WorkingDaysCount > 0 {
				targetAmount = sched.TargetSales / float64(sched.WorkingDaysCount)
			} else {
				targetAmount = sched.TargetSales
			}
		}
	}

	isAchieved := false
	incentiveBonus := 0.0
	if targetAmount > 0 && totalAmount >= targetAmount {
		isAchieved = true
		incentiveBonus = 50000.0 // Bonus +Rp 50.000 jika capai target
	}
	totalPayout := baseFee + incentiveBonus

	report := domain.MTSpgDailyReport{
		ScheduleID:         schedUUID,
		SpgID:              spgUUID,
		KeyAccountID:       kaUUID,
		SupermarketName:    req.SupermarketName,
		ReportDate:         reportDate,
		Shift:              req.Shift,
		TargetAmount:       targetAmount,
		TotalAmount:        totalAmount,
		TotalQty:           totalQty,
		BaseFee:            baseFee,
		IncentiveBonus:     incentiveBonus,
		TotalSpgPayout:     totalPayout,
		IsTargetAchieved:   isAchieved,
		ReceiptPhotoURL:    req.ReceiptPhotoURL,
		AttendancePhotoURL: req.AttendancePhotoURL,
		Notes:              req.Notes,
		Status:             "SUBMITTED",
		Items:              items,
	}

	if err := h.db.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create daily report: " + err.Error()})
		return
	}

	// Load associations
	h.db.Preload("Schedule").Preload("Spg").Preload("KeyAccount").Preload("Items").First(&report, report.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Laporan penjualan SPG berhasil dibuat",
		"data":    report,
	})
}

func (h *ModernTradeHandler) UpdateSpgDailyReport(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var report domain.MTSpgDailyReport
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_spg_daily_reports.key_account_id").
		Where("mt_spg_daily_reports.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&report).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laporan penjualan tidak ditemukan"})
		return
	}

	var req SpgDailyReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if req.SpgID != "" {
		if u, err := uuid.Parse(req.SpgID); err == nil {
			report.SpgID = u
		}
	}
	if req.KeyAccountID != "" {
		if u, err := uuid.Parse(req.KeyAccountID); err == nil {
			report.KeyAccountID = u
		}
	}
	if req.SupermarketName != "" {
		report.SupermarketName = req.SupermarketName
	}
	if req.ReportDate != "" {
		if t, err := time.Parse("2006-01-02", req.ReportDate); err == nil {
			report.ReportDate = t
		}
	}
	if req.Shift != "" {
		report.Shift = req.Shift
	}
	if req.ReceiptPhotoURL != "" {
		report.ReceiptPhotoURL = req.ReceiptPhotoURL
	}
	if req.AttendancePhotoURL != "" {
		report.AttendancePhotoURL = req.AttendancePhotoURL
	}
	report.Notes = req.Notes

	tx := h.db.Begin()
	if len(req.Items) > 0 {
		// Replace items
		if err := tx.Where("report_id = ?", report.ID).Delete(&domain.MTSpgSaleItem{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update items: " + err.Error()})
			return
		}

		var totalAmount float64
		var totalQty int
		var newItems []domain.MTSpgSaleItem

		for _, itemReq := range req.Items {
			var pUUID *uuid.UUID
			if itemReq.ProductID != nil && *itemReq.ProductID != "" {
				if parsed, err := uuid.Parse(*itemReq.ProductID); err == nil {
					pUUID = &parsed
				}
			}

			subtotal := itemReq.Subtotal
			if subtotal <= 0 && itemReq.Quantity > 0 && itemReq.UnitPrice > 0 {
				subtotal = float64(itemReq.Quantity) * itemReq.UnitPrice
			}

			qty := itemReq.Quantity
			if qty <= 0 {
				qty = 1
			}

			totalAmount += subtotal
			totalQty += qty

			newItems = append(newItems, domain.MTSpgSaleItem{
				ReportID:    report.ID,
				ProductID:   pUUID,
				ProductName: itemReq.ProductName,
				SkuCode:     itemReq.SkuCode,
				UnitPrice:   itemReq.UnitPrice,
				Quantity:    qty,
				Subtotal:    subtotal,
			})
		}

		if len(newItems) > 0 {
			if err := tx.Create(&newItems).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert new items: " + err.Error()})
				return
			}
		}

		report.TotalAmount = totalAmount
		report.TotalQty = totalQty
	}

	if req.ScheduleID != nil && *req.ScheduleID != "" {
		if u, err := uuid.Parse(*req.ScheduleID); err == nil {
			report.ScheduleID = &u
		}
	}
	if req.TargetAmount > 0 {
		report.TargetAmount = req.TargetAmount
	}
	if req.BaseFee > 0 {
		report.BaseFee = req.BaseFee
	}

	if report.TargetAmount > 0 && report.TotalAmount >= report.TargetAmount {
		report.IsTargetAchieved = true
		report.IncentiveBonus = 50000.0
	} else {
		report.IsTargetAchieved = false
		report.IncentiveBonus = 0.0
	}
	if report.BaseFee <= 0 {
		report.BaseFee = 200000.0
	}
	report.TotalSpgPayout = report.BaseFee + report.IncentiveBonus

	if err := tx.Save(&report).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	// If already approved, recalculate SPG's actual_sales
	if report.Status == "APPROVED" {
		h.recalculateSpgActualSales(report.SpgID)
	}

	h.db.Preload("Spg").Preload("KeyAccount").Preload("Items").First(&report, report.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Laporan penjualan berhasil diperbarui", "data": report})
}

func (h *ModernTradeHandler) DeleteSpgDailyReport(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var report domain.MTSpgDailyReport
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_spg_daily_reports.key_account_id").
		Where("mt_spg_daily_reports.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&report).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laporan penjualan tidak ditemukan"})
		return
	}

	spgID := report.SpgID
	wasApproved := report.Status == "APPROVED"

	if err := h.db.Select("Items").Delete(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if wasApproved {
		h.recalculateSpgActualSales(spgID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Laporan penjualan berhasil dihapus"})
}

func (h *ModernTradeHandler) VerifySpgDailyReport(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	compUUID, errComp := parseCompanyUUID(c)
	if errComp != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errComp.Error()})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"` // APPROVED, REJECTED
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var report domain.MTSpgDailyReport
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_spg_daily_reports.key_account_id").
		Where("mt_spg_daily_reports.id = ? AND key_accounts.company_id = ?", id, compUUID).
		Preload("Items").First(&report).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Laporan penjualan tidak ditemukan"})
		return
	}

	now := time.Now()
	report.Status = req.Status
	report.VerifiedAt = &now

	if userIDVal, exists := c.Get("userID"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			report.VerifiedBy = &uid
		}
	}
	if req.Notes != "" {
		report.Notes = req.Notes
	}

	if err := h.db.Save(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Always sync SPG actual_sales based on all APPROVED reports
	h.recalculateSpgActualSales(report.SpgID)

	// If APPROVED, reduce store stocks for each item sold in report
	if req.Status == "APPROVED" {
		go h.deductStoreStockForSpgReport(report.ID, compUUID)
	}

	h.db.Preload("Spg").Preload("KeyAccount").Preload("Items").First(&report, report.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Status laporan penjualan SPG berhasil diperbarui",
		"data":    report,
	})
}

func (h *ModernTradeHandler) recalculateSpgActualSales(spgID uuid.UUID) {
	var totalApproved float64
	h.db.Model(&domain.MTSpgDailyReport{}).
		Where("spg_id = ? AND status = 'APPROVED'", spgID).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalApproved)

	h.db.Model(&domain.MTSpg{}).Where("id = ?", spgID).Update("actual_sales", totalApproved)
}

func (h *ModernTradeHandler) UploadSpgReceiptPhoto(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit"})
		return
	}

	filename := uuid.New().String() + filepath.Ext(file.Filename)
	uploadDir := "uploads/spg_reports"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	savePath := filepath.Join(uploadDir, filename)
	if err := utils.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"url":     "/uploads/spg_reports/" + filename,
	})
}

// ─── MT Inventory & Stock Handlers ──────────────────────────────────────────

func (h *ModernTradeHandler) GetMTInventories(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	branchIDStr := c.Query("branch_id")
	category := c.Query("category")
	search := c.Query("search")

	query := h.db.Model(&domain.MTInventory{}).Preload("Branch").Preload("Company")

	if companyID != nil {
		query = query.Where("company_id = ? OR company_id IS NULL", companyID)
	}

	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		if bID, err := uuid.Parse(branchIDStr); err == nil {
			query = query.Where("branch_id = ? OR branch_id IS NULL", bID)
		}
	}

	if category != "" && category != "ALL" {
		query = query.Where("category = ?", category)
	}

	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var items []domain.MTInventory
	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
	})
}

func (h *ModernTradeHandler) CreateMTInventory(c *gin.Context) {
	var req struct {
		BranchID      *string                    `json:"branch_id"`
		Name          string                     `json:"name" binding:"required"`
		Category      domain.MTInventoryCategory `json:"category"`
		Quantity      int                        `json:"quantity"`
		Unit          string                     `json:"unit" binding:"required"`
		MinStockAlert int                        `json:"min_stock_alert"`
		PhotoURL      string                     `json:"photo_url"`
		Description   string                     `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := domain.MTInventory{
		Name:          req.Name,
		Category:      req.Category,
		Quantity:      req.Quantity,
		Unit:          req.Unit,
		MinStockAlert: req.MinStockAlert,
		PhotoURL:      req.PhotoURL,
		Description:   req.Description,
	}

	if req.Category == "" {
		item.Category = domain.MTInventoryCatPromoGift
	}
	if item.MinStockAlert <= 0 {
		item.MinStockAlert = 10
	}

	if compUUID, errComp := parseCompanyUUID(c); errComp == nil && compUUID != uuid.Nil {
		item.CompanyID = &compUUID
	}

	if req.BranchID != nil && *req.BranchID != "" && *req.BranchID != "ALL_BRANCHES" {
		if bID, err := uuid.Parse(*req.BranchID); err == nil {
			item.BranchID = &bID
		}
	}

	if err := h.db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If initial quantity > 0, log it
	if item.Quantity > 0 {
		userName, _ := c.Get("userName")
		createdBy := "System"
		if userName != nil {
			createdBy = userName.(string)
		}

		log := domain.MTInventoryLog{
			MTInventoryID:  item.ID,
			Type:           domain.MTInventoryLogIN,
			Quantity:       item.Quantity,
			QuantityBefore: 0,
			QuantityAfter:  item.Quantity,
			ReferenceNo:    "INITIAL-STOCK",
			RecipientType:  "GUDANG",
			RecipientName:  "Gudang Modern Trade",
			Notes:          "Stok awal pembuatan master barang",
			CreatedBy:      createdBy,
		}
		_ = h.db.Create(&log)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Barang inventaris MT berhasil dibuat",
		"data":    item,
	})
}

func (h *ModernTradeHandler) UpdateMTInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var item domain.MTInventory
	if err := h.db.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	var req struct {
		BranchID      *string                    `json:"branch_id"`
		Name          string                     `json:"name"`
		Category      domain.MTInventoryCategory `json:"category"`
		Unit          string                     `json:"unit"`
		MinStockAlert int                        `json:"min_stock_alert"`
		PhotoURL      string                     `json:"photo_url"`
		Description   string                     `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Category != "" {
		item.Category = req.Category
	}
	if req.Unit != "" {
		item.Unit = req.Unit
	}
	if req.MinStockAlert > 0 {
		item.MinStockAlert = req.MinStockAlert
	}
	item.PhotoURL = req.PhotoURL
	item.Description = req.Description

	if req.BranchID != nil {
		if *req.BranchID == "" || *req.BranchID == "ALL_BRANCHES" {
			item.BranchID = nil
		} else if bID, err := uuid.Parse(*req.BranchID); err == nil {
			item.BranchID = &bID
		}
	}

	if err := h.db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item inventaris MT berhasil diperbarui",
		"data":    item,
	})
}

func (h *ModernTradeHandler) DeleteMTInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.db.Where("mt_inventory_id = ?", id).Delete(&domain.MTInventoryLog{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Delete(&domain.MTInventory{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item inventaris MT berhasil dihapus",
	})
}

func (h *ModernTradeHandler) MutateMTInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req struct {
		Type          domain.MTInventoryLogType `json:"type" binding:"required"` // IN, OUT, ADJUSTMENT, RETURN
		Quantity      int                       `json:"quantity" binding:"required"`
		ReferenceNo   string                    `json:"reference_no"`
		RecipientType string                    `json:"recipient_type"` // SPG, OUTLET_MT, EVENT, GUDANG, LAINNYA
		RecipientName string                    `json:"recipient_name"`
		Notes         string                    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Quantity <= 0 && req.Type != domain.MTInventoryLogADJUSTMENT {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity harus lebih besar dari 0"})
		return
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var item domain.MTInventory
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&item, "id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	qtyBefore := item.Quantity
	var qtyAfter int

	switch req.Type {
	case domain.MTInventoryLogIN, domain.MTInventoryLogRETURN:
		qtyAfter = qtyBefore + req.Quantity
	case domain.MTInventoryLogOUT:
		if qtyBefore < req.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Stok tidak mencukupi untuk pengeluaran"})
			return
		}
		qtyAfter = qtyBefore - req.Quantity
	case domain.MTInventoryLogADJUSTMENT:
		// In adjustment, quantity is the new target quantity
		qtyAfter = req.Quantity
	default:
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipe mutasi tidak valid"})
		return
	}

	item.Quantity = qtyAfter
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userName, _ := c.Get("userName")
	createdBy := "Admin MT"
	if userName != nil {
		createdBy = userName.(string)
	}

	log := domain.MTInventoryLog{
		MTInventoryID:  item.ID,
		Type:           req.Type,
		Quantity:       req.Quantity,
		QuantityBefore: qtyBefore,
		QuantityAfter:  qtyAfter,
		ReferenceNo:    req.ReferenceNo,
		RecipientType:  req.RecipientType,
		RecipientName:  req.RecipientName,
		Notes:          req.Notes,
		CreatedBy:      createdBy,
	}

	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mutasi stok inventaris MT berhasil dicatat",
		"data": gin.H{
			"inventory": item,
			"log":       log,
		},
	})
}

func (h *ModernTradeHandler) GetMTInventoryLogs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var logs []domain.MTInventoryLog
	if err := h.db.Where("mt_inventory_id = ?", id).Order("created_at DESC").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
	})
}

func (h *ModernTradeHandler) GetAllMTInventoryLogs(c *gin.Context) {
	var logs []domain.MTInventoryLog
	query := h.db.Model(&domain.MTInventoryLog{}).Preload("MTInventory").Order("created_at DESC").Limit(200)

	typeFilter := c.Query("type")
	if typeFilter != "" && typeFilter != "ALL" {
		query = query.Where("type = ?", typeFilter)
	}

	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
	})
}

func (h *ModernTradeHandler) UploadMTInventoryPhoto(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit"})
		return
	}

	filename := uuid.New().String() + filepath.Ext(file.Filename)
	uploadDir := "uploads/mt_inventories"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	savePath := filepath.Join(uploadDir, filename)
	if err := utils.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"url":     "/uploads/mt_inventories/" + filename,
	})
}

// ─── SPG Scheduling & Budgeting Handlers ──────────────────────────────────────

func calculateWorkingDays(startDate, endDate time.Time, pattern string) int {
	days := 0
	cur := startDate
	for !cur.After(endDate) {
		weekday := cur.Weekday()
		switch pattern {
		case "WEEKEND_FRI_SUN":
			if weekday == time.Friday || weekday == time.Saturday || weekday == time.Sunday {
				days++
			}
		case "DAILY_MON_SAT":
			if weekday != time.Sunday {
				days++
			}
		default: // ALL / CUSTOM
			days++
		}
		cur = cur.AddDate(0, 0, 1)
	}
	if days == 0 {
		days = 1
	}
	return days
}

func (h *ModernTradeHandler) GetSpgSchedules(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var schedules []domain.MTSpgSchedule
	query := h.db.Where("company_id = ?", companyID).
		Preload("Spg").
		Preload("KeyAccount").
		Order("start_date DESC")

	if spgID := c.Query("spg_id"); spgID != "" {
		query = query.Where("spg_id = ?", spgID)
	}
	if kaID := c.Query("key_account_id"); kaID != "" {
		query = query.Where("key_account_id = ?", kaID)
	}
	if status := c.Query("status"); status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&schedules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": schedules})
}

func (h *ModernTradeHandler) CreateSpgSchedule(c *gin.Context) {
	compUUID, errComp := parseCompanyUUID(c)
	if errComp != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errComp.Error()})
		return
	}

	var req struct {
		SpgID           string  `json:"spg_id" binding:"required"`
		KeyAccountID    string  `json:"key_account_id" binding:"required"`
		SupermarketName string  `json:"supermarket_name" binding:"required"`
		StartDate       string  `json:"start_date" binding:"required"`
		EndDate         string  `json:"end_date" binding:"required"`
		DayPattern      string  `json:"day_pattern"` // WEEKEND_FRI_SUN, DAILY_MON_SAT, CUSTOM
		DailyFee        float64 `json:"daily_fee"`
		AllowanceFee    float64 `json:"allowance_fee"`
		TargetSales     float64 `json:"target_sales"`
		Status          string  `json:"status"` // PLANNED, ACTIVE, COMPLETED, CANCELLED
		Notes           string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spgUUID, err := uuid.Parse(req.SpgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid spg_id format"})
		return
	}
	kaUUID, err := uuid.Parse(req.KeyAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key_account_id format"})
		return
	}
	cUUID := compUUID

	sDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, must be YYYY-MM-DD"})
		return
	}
	eDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, must be YYYY-MM-DD"})
		return
	}

	pattern := req.DayPattern
	if pattern == "" {
		pattern = "WEEKEND_FRI_SUN"
	}

	workingDays := calculateWorkingDays(sDate, eDate, pattern)
	totalBudget := float64(workingDays) * (req.DailyFee + req.AllowanceFee)

	status := req.Status
	if status == "" {
		status = "PLANNED"
	}

	schedule := domain.MTSpgSchedule{
		ID:               uuid.New(),
		CompanyID:        cUUID,
		SpgID:            spgUUID,
		KeyAccountID:     kaUUID,
		SupermarketName:  req.SupermarketName,
		StartDate:        sDate,
		EndDate:          eDate,
		DayPattern:       pattern,
		WorkingDaysCount: workingDays,
		DailyFee:         req.DailyFee,
		AllowanceFee:     req.AllowanceFee,
		TotalBudget:      totalBudget,
		TargetSales:      req.TargetSales,
		Status:           status,
		Notes:            req.Notes,
	}

	if err := h.db.Create(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Preload associations before returning
	h.db.Preload("Spg").Preload("KeyAccount").First(&schedule, schedule.ID)

	c.JSON(http.StatusCreated, gin.H{"message": "Jadwal & Budget SPG berhasil dibuat", "data": schedule})
}

func (h *ModernTradeHandler) UpdateSpgSchedule(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var schedule domain.MTSpgSchedule
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&schedule).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jadwal SPG tidak ditemukan"})
		return
	}

	var req struct {
		SpgID           string  `json:"spg_id"`
		KeyAccountID    string  `json:"key_account_id"`
		SupermarketName string  `json:"supermarket_name"`
		StartDate       string  `json:"start_date"`
		EndDate         string  `json:"end_date"`
		DayPattern      string  `json:"day_pattern"`
		DailyFee        float64 `json:"daily_fee"`
		AllowanceFee    float64 `json:"allowance_fee"`
		TargetSales     float64 `json:"target_sales"`
		Status          string  `json:"status"`
		Notes           string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SpgID != "" {
		if uid, err := uuid.Parse(req.SpgID); err == nil {
			schedule.SpgID = uid
		}
	}
	if req.KeyAccountID != "" {
		if uid, err := uuid.Parse(req.KeyAccountID); err == nil {
			schedule.KeyAccountID = uid
		}
	}
	if req.SupermarketName != "" {
		schedule.SupermarketName = req.SupermarketName
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			schedule.StartDate = t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			schedule.EndDate = t
		}
	}
	if req.DayPattern != "" {
		schedule.DayPattern = req.DayPattern
	}
	if req.DailyFee >= 0 {
		schedule.DailyFee = req.DailyFee
	}
	if req.AllowanceFee >= 0 {
		schedule.AllowanceFee = req.AllowanceFee
	}
	if req.TargetSales >= 0 {
		schedule.TargetSales = req.TargetSales
	}
	if req.Status != "" {
		schedule.Status = req.Status
	}
	if req.Notes != "" {
		schedule.Notes = req.Notes
	}

	// Recalculate working days and budget
	schedule.WorkingDaysCount = calculateWorkingDays(schedule.StartDate, schedule.EndDate, schedule.DayPattern)
	schedule.TotalBudget = float64(schedule.WorkingDaysCount) * (schedule.DailyFee + schedule.AllowanceFee)

	if err := h.db.Save(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.db.Preload("Spg").Preload("KeyAccount").First(&schedule, schedule.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Jadwal & Budget SPG berhasil diperbarui", "data": schedule})
}

func (h *ModernTradeHandler) DeleteSpgSchedule(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var schedule domain.MTSpgSchedule
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&schedule).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jadwal SPG tidak ditemukan"})
		return
	}

	if err := h.db.Delete(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Jadwal SPG berhasil dihapus"})
}

func (h *ModernTradeHandler) GetSpgBudgetSummary(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	type StoreBudgetStat struct {
		SupermarketName string  `json:"supermarket_name"`
		KeyAccountName  string  `json:"key_account_name"`
		TotalSchedules  int     `json:"total_schedules"`
		TotalBudget     float64 `json:"total_budget"`
		TotalTarget     float64 `json:"total_target"`
	}

	var stats []StoreBudgetStat
	err := h.db.Table("mt_spg_schedules").
		Select("mt_spg_schedules.supermarket_name, key_accounts.name as key_account_name, count(mt_spg_schedules.id) as total_schedules, sum(mt_spg_schedules.total_budget) as total_budget, sum(mt_spg_schedules.target_sales) as total_target").
		Joins("JOIN key_accounts ON key_accounts.id = mt_spg_schedules.key_account_id").
		Where("mt_spg_schedules.company_id = ?", companyID).
		Group("mt_spg_schedules.supermarket_name, key_accounts.name").
		Order("total_budget DESC").
		Scan(&stats).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var totalOverallBudget, totalOverallTarget float64
	for _, s := range stats {
		totalOverallBudget += s.TotalBudget
		totalOverallTarget += s.TotalTarget
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"store_budgets":        stats,
			"total_overall_budget": totalOverallBudget,
			"total_overall_target": totalOverallTarget,
		},
	})
}

// ─── MT Store Stock & Comparison Handlers ────────────────────────────────────

func (h *ModernTradeHandler) GetStoreStocks(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	query := h.db.Where("company_id = ?", companyID).Preload("Store").Preload("Product")

	storeID := c.Query("store_id")
	if storeID != "" {
		if sUUID, err := uuid.Parse(storeID); err == nil {
			query = query.Where("store_id = ?", sUUID)
		}
	}

	var stocks []domain.MTStoreStock
	if err := query.Order("updated_at DESC").Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stocks})
}

type StoreStockAdjustItem struct {
	ProductID         string `json:"product_id" binding:"required"`
	Quantity          int    `json:"quantity"`
	MinStockThreshold int    `json:"min_stock_threshold"`
}

func (h *ModernTradeHandler) AdjustStoreStock(c *gin.Context) {
	cUUID, errComp := parseCompanyUUID(c)
	if errComp != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errComp.Error()})
		return
	}

	var req struct {
		StoreID           string                 `json:"store_id" binding:"required"`
		ProductID         string                 `json:"product_id"`
		Quantity          int                    `json:"quantity"`
		MinStockThreshold int                    `json:"min_stock_threshold"`
		Notes             string                 `json:"notes"`
		ReferenceNo       string                 `json:"reference_no"`
		AdjustmentType    string                 `json:"adjustment_type"` // ADJUSTMENT, INITIAL, RETURN
		Items             []StoreStockAdjustItem `json:"items"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sUUID, err := uuid.Parse(req.StoreID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store_id format"})
		return
	}

	userName, _ := c.Get("userName")
	createdBy := "Admin"
	if userName != nil {
		createdBy = userName.(string)
	}

	adjType := req.AdjustmentType
	if adjType == "" {
		adjType = "ADJUSTMENT"
	}

	// Prepare items list: either from req.Items or single req.ProductID
	itemsToProcess := req.Items
	if len(itemsToProcess) == 0 && req.ProductID != "" {
		itemsToProcess = append(itemsToProcess, StoreStockAdjustItem{
			ProductID:         req.ProductID,
			Quantity:          req.Quantity,
			MinStockThreshold: req.MinStockThreshold,
		})
	}

	if len(itemsToProcess) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setidaknya 1 produk wajib diisi"})
		return
	}

	now := time.Now()
	var processedStocks []domain.MTStoreStock

	tx := h.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}
	for _, it := range itemsToProcess {
		pUUID, err := uuid.Parse(it.ProductID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product_id format", "product_id": it.ProductID})
			return
		}

		// Do not silently create stock for a store/product that is outside the
		// authenticated company's scope. A successful response must mean that
		// the opname was actually persisted.
		var store domain.Store
		if err := tx.Where("id = ? AND company_id = ?", sUUID, cUUID).First(&store).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Store tidak ditemukan pada perusahaan aktif"})
			return
		}
		var product domain.Product
		if err := tx.Where("id = ?", pUUID).First(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Produk tidak ditemukan", "product_id": it.ProductID})
			return
		}

		var stock domain.MTStoreStock
		err = tx.Where("company_id = ? AND store_id = ? AND product_id = ?", cUUID, sUUID, pUUID).First(&stock).Error

		qtyBefore := 0
		if err == nil {
			qtyBefore = stock.Quantity
		} else {
			stock = domain.MTStoreStock{
				ID:                uuid.New(),
				CompanyID:         cUUID,
				StoreID:           sUUID,
				ProductID:         pUUID,
				MinStockThreshold: 20,
			}
		}

		if it.MinStockThreshold > 0 {
			stock.MinStockThreshold = it.MinStockThreshold
		}
		stock.Quantity = it.Quantity
		stock.LastAdjustedAt = now

		if err := tx.Save(&stock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log := domain.MTStoreStockLog{
			ID:             uuid.New(),
			CompanyID:      cUUID,
			StoreID:        sUUID,
			ProductID:      pUUID,
			Type:           adjType,
			Quantity:       it.Quantity,
			QuantityBefore: qtyBefore,
			QuantityAfter:  it.Quantity,
			ReferenceNo:    req.ReferenceNo,
			Notes:          req.Notes,
			CreatedBy:      createdBy,
		}
		if err := tx.Create(&log).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		processedStocks = append(processedStocks, stock)
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sisa stok toko MT berhasil disesuaikan",
		"count":   len(processedStocks),
		"data":    processedStocks,
	})
}

func (h *ModernTradeHandler) GetStoreStockLogs(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	query := h.db.Where("company_id = ?", companyID).Preload("Store").Preload("Product")

	storeID := c.Query("store_id")
	if storeID != "" {
		if sUUID, err := uuid.Parse(storeID); err == nil {
			query = query.Where("store_id = ?", sUUID)
		}
	}

	var logs []domain.MTStoreStockLog
	if err := query.Order("created_at DESC").Limit(100).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *ModernTradeHandler) GetStoreStockComparison(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	type StoreStockSummaryItem struct {
		StoreID           uuid.UUID  `json:"store_id"`
		StoreName         string     `json:"store_name"`
		StoreAddress      string     `json:"store_address"`
		OwnerName         string     `json:"owner_name"`
		PhoneNumber       string     `json:"phone_number"`
		TotalSkus         int        `json:"total_skus"`
		TotalStockPcs     int        `json:"total_stock_pcs"`
		MinStockThreshold int        `json:"min_stock_threshold"`
		StockStatus       string     `json:"stock_status"`   // OUT_OF_STOCK, LOW_STOCK, SUFFICIENT, OVERSTOCKED
		Recommendation    string     `json:"recommendation"` // REPEAT_ORDER, SPG_BOOSTER, NORMAL
		LastAdjustedAt    *time.Time `json:"last_adjusted_at"`
	}

	var rows []struct {
		StoreID           uuid.UUID  `gorm:"column:store_id"`
		StoreName         string     `gorm:"column:store_name"`
		StoreAddress      string     `gorm:"column:store_address"`
		OwnerName         string     `gorm:"column:owner_name"`
		PhoneNumber       string     `gorm:"column:phone_number"`
		TotalSkus         int        `gorm:"column:total_skus"`
		TotalStockPcs     int        `gorm:"column:total_stock_pcs"`
		MinStockThreshold int        `gorm:"column:min_stock_threshold"`
		LastAdjustedAt    *time.Time `gorm:"column:last_adjusted_at"`
	}

	query := `
		SELECT 
			s.id AS store_id,
			s.name AS store_name,
			s.address AS store_address,
			s.owner_name AS owner_name,
			s.phone_number AS phone_number,
			COALESCE(COUNT(st.product_id), 0) AS total_skus,
			COALESCE(SUM(st.quantity), 0) AS total_stock_pcs,
			COALESCE(AVG(st.min_stock_threshold), 20) AS min_stock_threshold,
			MAX(st.last_adjusted_at) AS last_adjusted_at
		FROM stores s
		LEFT JOIN mt_store_stocks st ON st.store_id = s.id AND st.company_id = s.company_id
		WHERE s.company_id = ? 
		  AND s.is_active = true
		  AND (
			s.assigned_employee_id IN (
				SELECT e.id FROM employees e 
				JOIN job_positions jp ON jp.id = e.job_position_id 
				WHERE jp.title ILIKE '%sales mt%' 
				   OR jp.title ILIKE '%modern trade%' 
				   OR jp.title ILIKE '%kam%' 
				   OR jp.title ILIKE '%key account%' 
				   OR jp.title ILIKE '%merchandis%'
				   OR jp.title ILIKE '%spg%'
			)
			OR s.created_by_employee_id IN (
				SELECT e.id FROM employees e 
				JOIN job_positions jp ON jp.id = e.job_position_id 
				WHERE jp.title ILIKE '%sales mt%' 
				   OR jp.title ILIKE '%modern trade%' 
				   OR jp.title ILIKE '%kam%' 
				   OR jp.title ILIKE '%key account%' 
				   OR jp.title ILIKE '%merchandis%'
				   OR jp.title ILIKE '%spg%'
			)
			OR s.id IN (SELECT DISTINCT store_id FROM mt_store_stocks)
			OR s.id IN (SELECT DISTINCT so.store_id FROM sales_orders so JOIN employees e ON e.id = so.employee_id JOIN job_positions jp ON jp.id = e.job_position_id WHERE jp.title ILIKE '%sales mt%' OR jp.title ILIKE '%modern trade%')
			OR s.customer_type IN ('SUPERMARKET', 'MINIMARKET', 'HYPERMARKET')
		  )
		GROUP BY s.id, s.name, s.address, s.owner_name, s.phone_number
		ORDER BY total_stock_pcs DESC, s.name ASC
	`

	if err := h.db.Raw(query, companyID).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []StoreStockSummaryItem
	var totalOverallPcs int
	var totalOosCount int
	var totalLowStockCount int
	var totalSpgRecommendedCount int

	for _, r := range rows {
		status := "SUFFICIENT"
		rec := "NORMAL"

		if r.TotalStockPcs == 0 {
			status = "OUT_OF_STOCK"
			rec = "REPEAT_ORDER_URGENT"
			totalOosCount++
		} else if r.TotalStockPcs <= r.MinStockThreshold {
			status = "LOW_STOCK"
			rec = "REPEAT_ORDER"
			totalLowStockCount++
		} else if r.TotalStockPcs >= 100 {
			status = "OVERSTOCKED"
			rec = "SPG_BOOSTER"
			totalSpgRecommendedCount++
		} else {
			status = "SUFFICIENT"
			rec = "NORMAL"
		}

		totalOverallPcs += r.TotalStockPcs

		result = append(result, StoreStockSummaryItem{
			StoreID:           r.StoreID,
			StoreName:         r.StoreName,
			StoreAddress:      r.StoreAddress,
			OwnerName:         r.OwnerName,
			PhoneNumber:       r.PhoneNumber,
			TotalSkus:         r.TotalSkus,
			TotalStockPcs:     r.TotalStockPcs,
			MinStockThreshold: r.MinStockThreshold,
			StockStatus:       status,
			Recommendation:    rec,
			LastAdjustedAt:    r.LastAdjustedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"stores":                       result,
			"total_overall_pcs":            totalOverallPcs,
			"total_oos_stores":             totalOosCount,
			"total_low_stock_stores":       totalLowStockCount,
			"total_spg_recommended_stores": totalSpgRecommendedCount,
		},
	})
}

func (h *ModernTradeHandler) ApprovePODeliveryReceipt(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PO ID format"})
		return
	}

	compUUID, errComp := parseCompanyUUID(c)
	if errComp != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errComp.Error()})
		return
	}
	cUUID := compUUID

	var po domain.MTPurchaseOrder
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_purchase_orders.key_account_id").
		Where("mt_purchase_orders.id = ? AND key_accounts.company_id = ?", id, compUUID).
		Preload("KeyAccount").First(&po).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase Order tidak ditemukan"})
		return
	}

	// Update PO Status to DELIVERED_AND_UPLOADED and calculate due date (TOP 30 Days)
	now := time.Now()
	dueDate := po.OrderDate.AddDate(0, 0, 30)
	po.Status = "DELIVERED_AND_UPLOADED"
	po.DueDate = &dueDate
	po.BillingStatus = "PENDING_SETTLEMENT"

	if err := h.db.Save(&po).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Influx stock into mt_store_stocks based on SupermarketName or KeyAccount Name
	storeSearch := po.SupermarketName
	if storeSearch == "" && po.KeyAccount != nil {
		storeSearch = po.KeyAccount.Name
	}

	if storeSearch != "" {
		var store domain.Store
		if err := h.db.Where("company_id = ? AND (LOWER(name) = LOWER(?) OR LOWER(name) LIKE LOWER(?))", cUUID, storeSearch, "%"+storeSearch+"%").First(&store).Error; err == nil {
			// Find default/first product for this company to increment or match
			var product domain.Product
			if err := h.db.Where("company_id = ?", cUUID).First(&product).Error; err == nil {
				pcsToAdd := po.ItemsShipped
				if po.PcsPerUnit > 0 {
					pcsToAdd = po.ItemsShipped * po.PcsPerUnit
				} else if pcsToAdd == 0 {
					pcsToAdd = po.ItemsOrdered
				}

				var stock domain.MTStoreStock
				err := h.db.Where("company_id = ? AND store_id = ? AND product_id = ?", cUUID, store.ID, product.ID).First(&stock).Error
				qtyBefore := 0
				if err == nil {
					qtyBefore = stock.Quantity
				} else {
					stock = domain.MTStoreStock{
						ID:                uuid.New(),
						CompanyID:         cUUID,
						StoreID:           store.ID,
						ProductID:         product.ID,
						MinStockThreshold: 20,
					}
				}

				qtyAfter := qtyBefore + pcsToAdd
				stock.Quantity = qtyAfter
				stock.LastAdjustedAt = now
				h.db.Save(&stock)

				log := domain.MTStoreStockLog{
					ID:             uuid.New(),
					CompanyID:      cUUID,
					StoreID:        store.ID,
					ProductID:      product.ID,
					Type:           "DELIVERY_IN",
					Quantity:       pcsToAdd,
					QuantityBefore: qtyBefore,
					QuantityAfter:  qtyAfter,
					ReferenceNo:    "PO-DELIV-" + po.PONumber,
					Notes:          "Penerimaan Pengiriman PO " + po.PONumber,
					CreatedBy:      "Admin MT Delivery Approval",
				}
				h.db.Create(&log)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Penerimaan pengiriman berhasil di-approve dan stok toko MT telah diperbarui",
		"data":    po,
	})
}

func (h *ModernTradeHandler) CollectPOPayment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PO ID format"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var po domain.MTPurchaseOrder
	if err := h.db.Joins("JOIN key_accounts ON key_accounts.id = mt_purchase_orders.key_account_id").
		Where("mt_purchase_orders.id = ? AND key_accounts.company_id = ?", id, companyID).
		First(&po).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase Order tidak ditemukan"})
		return
	}

	var req struct {
		PaymentProofURL string `json:"payment_proof_url"`
		Notes           string `json:"notes"`
		BillingStatus   string `json:"billing_status"` // PAID_FULL, PENDING_SETTLEMENT
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Also allow form data
		req.PaymentProofURL = c.PostForm("payment_proof_url")
		req.Notes = c.PostForm("notes")
		req.BillingStatus = c.PostForm("billing_status")
	}

	// Handle photo upload if present in form-data
	if file, err := c.FormFile("file"); err == nil {
		uploadDir := "./uploads/payment_proofs"
		_ = os.MkdirAll(uploadDir, os.ModePerm)
		ext := filepath.Ext(file.Filename)
		filename := "po_pay_" + uuid.New().String() + ext
		filePath := filepath.Join(uploadDir, filename)
		if err := c.SaveUploadedFile(file, filePath); err == nil {
			req.PaymentProofURL = "/uploads/payment_proofs/" + filename
		}
	}

	now := time.Now()
	userName, _ := c.Get("userName")
	collectedBy := "Sales MT"
	if userName != nil {
		collectedBy = userName.(string)
	}

	po.BillingStatus = "PAID_FULL"
	if req.BillingStatus != "" {
		po.BillingStatus = req.BillingStatus
	}
	if req.PaymentProofURL != "" {
		po.PaymentProofURL = req.PaymentProofURL
	}
	po.CollectedAt = &now
	po.CollectedBy = collectedBy

	if err := h.db.Save(&po).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bukti penagihan / pelunasan TOP berhasil disimpan",
		"data":    po,
	})
}
