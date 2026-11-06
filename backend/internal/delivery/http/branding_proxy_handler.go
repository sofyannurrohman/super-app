package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/sse"
	"example.com/hris/backend/pkg/utils"
)

// BrandingProxyHandler is now 100% independent and queries PostgreSQL directly.
type BrandingProxyHandler struct {
	localDB         *gorm.DB // PostgreSQL Super App Connection
	employeeUseCase usecase.EmployeeUsecase
	gdrive          *utils.GDriveService
	hub             *sse.Hub // SSE Hub untuk realtime notification
}

func NewBrandingProxyHandler(brandingDB *gorm.DB, localDB *gorm.DB, employeeUseCase usecase.EmployeeUsecase, hub *sse.Hub) *BrandingProxyHandler {
	// Initialize GDrive Service
	var gdrive *utils.GDriveService
	if credentialsFile, folderID := os.Getenv("GDRIVE_CREDENTIALS_FILE"), os.Getenv("GDRIVE_FOLDER_ID"); credentialsFile != "" && folderID != "" {
		var err error
		gdrive, err = utils.NewGDriveService(credentialsFile, folderID)
		if err != nil {
			log.Printf("[BrandingHandler] GDrive Service gagal diinisialisasi: %v", err)
		}
	}

	// Asynchronously optimize indexes on banner_orders and stores for lightning fast queries
	go func(db *gorm.DB) {
		db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_lower_store_name ON banner_orders (LOWER(store_name))")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_status ON banner_orders (status)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_store_id ON banner_orders (store_id)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_employee_id ON banner_orders (employee_id)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_operator_id ON banner_orders (operator_id)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_banner_orders_created_at ON banner_orders (created_at)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_stores_lower_name ON stores (LOWER(name))")
		ensureBrandingAIDesignJobsTable(db)
	}(localDB)

	return &BrandingProxyHandler{
		localDB:         localDB,
		employeeUseCase: employeeUseCase,
		gdrive:          gdrive,
		hub:             hub,
	}
}

func (h *BrandingProxyHandler) RegisterRoutes(r *gin.RouterGroup) {
	branding := r.Group("/branding")
	{
		branding.GET("/shops", h.GetShops)
		branding.GET("/banners", h.GetBanners)
		branding.GET("/campaign-videos", h.GetCampaignVideos)
		campaignVideoAdmin := branding.Group("/campaign-videos")
		campaignVideoAdmin.Use(RoleMiddleware("admin_designer", "validator_desain", "validator_toko", "admin_purchasing", "manager", "direktur"))
		campaignVideoAdmin.POST("", h.CreateCampaignVideo)
		campaignVideoAdmin.PUT("/:id", h.UpdateCampaignVideo)
		campaignVideoAdmin.DELETE("/:id", h.DeleteCampaignVideo)
		campaignVideoAdmin.POST("/upload", h.UploadCampaignVideo)
		branding.POST("/banners", h.CreateBanner)
		branding.PUT("/banners/:id", h.UpdateBanner)
		branding.DELETE("/banners/:id", h.DeleteBanner)
		branding.GET("/agents", h.GetSalesAgents)
		branding.GET("/stats", h.GetStats)
		branding.GET("/operator/performance", h.GetOperatorPerformance)
		branding.POST("/operator/commission/payment", h.UpdateOperatorCommissionPayment)
		branding.GET("/editor/performance", h.GetEditorPerformance)
		branding.GET("/health", h.HealthCheck)
		internal := branding.Group("/internal")
		internal.Use(brandingInternalKeyMiddleware())
		internal.GET("/summary", h.GetInternalSummary)
		internal.GET("/orders", h.GetInternalOrders)
		internal.GET("/ai-design/jobs/next", h.ClaimNextAIDesignJob)
		internal.POST("/ai-design/jobs/:id/hermes-result", h.SubmitHermesDesignResult)
		internal.POST("/ai-design/jobs/:id/corel-result", h.SubmitCorelDesignResult)
		internal.POST("/ai-design/jobs/:id/fail", h.FailAIDesignJob)
		branding.POST("/sync", h.SyncFromBrandingApp)
		branding.POST("/recover", h.RecoverBrandingWorkflow)
		branding.POST("/migrate-vps-designs", h.MigrateVPSDesignsToGDrive)

		// Editor CRUD
		branding.GET("/editors", h.GetEditors)
		branding.POST("/editors", h.CreateEditor)
		branding.PUT("/editors/:id", h.UpdateEditor)
		branding.DELETE("/editors/:id", h.DeleteEditor)

		// Operator CRUD
		branding.GET("/operators", h.GetOperators)
		branding.POST("/operators", h.CreateOperator)
		branding.PUT("/operators/:id", h.UpdateOperator)
		branding.DELETE("/operators/:id", h.DeleteOperator)
		branding.PUT("/operators/:id/target", h.SetOperatorTarget)

		// Vendor CRUD
		branding.GET("/vendors", h.GetVendors)
		branding.POST("/vendors", h.CreateVendor)
		branding.PUT("/vendors/:id", h.UpdateVendor)
		branding.DELETE("/vendors/:id", h.DeleteVendor)

		// State Machine Endpoints
		branding.POST("/banners/send-vendor-email", h.SendVendorEmail)
		branding.POST("/banners/:id/send-vendor-email", h.SendVendorEmail)
		branding.PUT("/banners/bulk-vendor-process", h.BulkVendorProcess)
		branding.PUT("/banners/:id/design", h.UploadDesign)
		branding.PUT("/banners/:id/validate", h.ValidateBanner)
		branding.PUT("/banners/:id/vendor-process", h.VendorProcess)
		branding.POST("/banners/:id/install", h.InstallBanner)
		branding.PUT("/banners/:id/assign", h.AssignBanner)
		branding.PUT("/banners/:id/assign-operator", h.AssignOperator)
		branding.PUT("/banners/:id/claim", h.ClaimBanner)
		branding.POST("/banners/:id/ai-draft", h.CreateAIDesignDraft)
		branding.GET("/banners/:id/ai-draft", h.GetLatestAIDesignDraft)
		branding.PUT("/banners/:id/operator-take", h.OperatorTakeBanner)

		// Sales Banner Installation & Bulk Reimbursement
		branding.GET("/sales/unclaimed-installations", h.GetUnclaimedInstallations)
		branding.POST("/sales/direct-install", h.DirectInstallBanner)
		branding.POST("/sales/bulk-reimburse", h.BulkReimburseInstallations)
	}
}

type OperatorCommissionPaymentRequest struct {
	OperatorID       string `json:"operator_id" binding:"required"`
	Month            int    `json:"month" binding:"required,min=1,max=12"`
	Year             int    `json:"year" binding:"required,min=2000,max=2100"`
	Action           string `json:"action" binding:"required"`
	PaymentReference string `json:"payment_reference"`
	PaymentNotes     string `json:"payment_notes"`
}

// UpdateOperatorCommissionPayment updates payment status for one operator and period.
func (h *BrandingProxyHandler) UpdateOperatorCommissionPayment(c *gin.Context) {
	var req OperatorCommissionPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	operatorID, err := uuid.Parse(req.OperatorID)
	if err != nil || (req.Action != "mark_paid" && req.Action != "mark_unpaid") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Operator atau aksi pembayaran tidak valid"})
		return
	}
	query := h.localDB.Model(&domain.BannerOrder{}).Where(
		"operator_id = ? AND reimbursement_id IS NULL AND LOWER(status) IN ('installed', 'terpasang') AND EXTRACT(MONTH FROM installed_at) = ? AND EXTRACT(YEAR FROM installed_at) = ?",
		operatorID, req.Month, req.Year,
	)
	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Action == "mark_paid" {
		updates["is_paid"] = true
		updates["payment_status"] = "paid"
		updates["paid_at"] = time.Now()
		updates["payment_reference"] = req.PaymentReference
		updates["payment_notes"] = req.PaymentNotes
	} else {
		updates["is_paid"] = false
		updates["payment_status"] = "unpaid"
		updates["paid_at"] = nil
		updates["payment_reference"] = nil
		updates["payment_notes"] = nil
	}
	result := query.Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui status pembayaran komisi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status pembayaran komisi berhasil diperbarui", "updated": result.RowsAffected})
}

func brandingInternalKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := os.Getenv("INTERNAL_API_KEY")
		if expected == "" || c.GetHeader("X-Internal-Key") != expected {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid internal API key"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *BrandingProxyHandler) GetInternalSummary(c *gin.Context) {
	var rows []struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}
	if err := h.localDB.Table("banner_orders").Select("LOWER(COALESCE(status, 'pending')) AS status, COUNT(*) AS count").Group("LOWER(COALESCE(status, 'pending'))").Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query branding summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"by_status": rows}, "read_only": true})
}

func (h *BrandingProxyHandler) GetInternalOrders(c *gin.Context) {
	status := strings.TrimSpace(strings.ToLower(c.Query("status")))
	query := strings.TrimSpace(c.Query("q"))
	var rows []struct {
		ID         string    `json:"id"`
		StoreName  string    `json:"store_name"`
		BannerType string    `json:"banner_type"`
		Status     string    `json:"status"`
		Qty        int       `json:"qty"`
		VendorName string    `json:"vendor_name"`
		CreatedAt  time.Time `json:"created_at"`
	}
	db := h.localDB.Table("banner_orders").Select("id::text, store_name, banner_type, COALESCE(status, 'pending') AS status, COALESCE(qty, 0) AS qty, COALESCE(vendor_name, '') AS vendor_name, created_at").Order("created_at DESC").Limit(50)
	if status != "" {
		db = db.Where("LOWER(COALESCE(status, 'pending')) = ?", status)
	}
	if query != "" {
		p := "%" + query + "%"
		db = db.Where("LOWER(store_name) LIKE ? OR LOWER(COALESCE(banner_type, '')) LIKE ?", strings.ToLower(p), strings.ToLower(p))
	}
	if err := db.Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query branding orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "count": len(rows), "read_only": true})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /branding/health
// ─────────────────────────────────────────────────────────────────────────────

func (h *BrandingProxyHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "online",
		"source":  "postgres",
		"message": "Sistem Branding 100% Independen di PostgreSQL Super App",
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /branding/shops
// ─────────────────────────────────────────────────────────────────────────────

type ShopResponse struct {
	ID            string  `json:"id"`
	SalesID       *string `json:"sales_id"`
	Name          string  `json:"name"`
	OwnerName     string  `json:"owner_name"`
	Phone         string  `json:"phone"`
	Photo         *string `json:"photo"`
	AddressDetail string  `json:"address_detail"`
	Status        string  `json:"status"`
	Latitude      string  `json:"latitude"`
	Longitude     string  `json:"longitude"`
	SalesName     string  `json:"sales_name"`
	SalesEmail    string  `json:"sales_email"`
	SalesPhone    string  `json:"sales_phone"`
	SalesArea     string  `json:"sales_area"`
}

func (h *BrandingProxyHandler) GetShops(c *gin.Context) {
	search := c.Query("search")
	branchID := c.Query("branch_id")
	salesID := c.Query("sales_id")
	salesName := c.Query("sales_name")
	limitStr := c.Query("limit")
	limit := 2000
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	query := h.localDB.Table("stores").
		Select(`
			stores.id::text as id,
			stores.assigned_employee_id::text as sales_id,
			stores.name as name,
			stores.owner_name as owner_name,
			stores.phone_number as phone,
			stores.photo_url as photo,
			stores.address as address_detail,
			'warm' as status,
			stores.latitude::text as latitude,
			stores.longitude::text as longitude,
			COALESCE(NULLIF(TRIM(sales.first_name || ' ' || COALESCE(sales.last_name, '')), ''), 'Unknown Sales') as sales_name,
			COALESCE(u.email, '') as sales_email,
			COALESCE(sales.phone_number, '') as sales_phone,
			COALESCE(branches.name, '') as sales_area
		`).
		Joins("LEFT JOIN employees sales ON sales.id = stores.assigned_employee_id").
		Joins("LEFT JOIN users u ON u.id = sales.user_id").
		Joins("LEFT JOIN branches ON branches.id = sales.branch_id").
		Where("stores.is_active = true")

	if branchID != "" && branchID != "ALL_BRANCHES" {
		query = query.Where("(sales.branch_id = ? OR sales.branch_id IS NULL)", branchID)
	}

	if salesID != "" {
		query = query.Where("stores.assigned_employee_id = ?", salesID)
	}

	if salesName != "" {
		sName := "%" + strings.ToLower(salesName) + "%"
		query = query.Where("LOWER(TRIM(sales.first_name || ' ' || COALESCE(sales.last_name, ''))) LIKE ?", sName)
	}

	if search != "" {
		sQuery := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(stores.name) LIKE ? OR LOWER(stores.address) LIKE ?", sQuery, sQuery)
	}

	var shops []ShopResponse
	if err := query.Order("stores.name ASC").Limit(limit).Find(&shops).Error; err != nil {
		log.Printf("[BrandingHandler] GetShops error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data toko"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"source":  "postgres",
		"data":    shops,
		"total":   len(shops),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /branding/banners
// ─────────────────────────────────────────────────────────────────────────────

func (h *BrandingProxyHandler) GetBanners(c *gin.Context) {
	h.GetBannersPostgres(c)
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /branding/banners
// ─────────────────────────────────────────────────────────────────────────────

type BrandingBannerRequest struct {
	ID                    string   `json:"id"`
	ShopID                string   `json:"shop_id"`
	ShopName              string   `json:"shop_name"`
	StoreName             string   `json:"store_name"`
	SalesName             string   `json:"sales_name"`
	Type                  string   `json:"type"`
	IsBihun               bool     `json:"is_bihun"`
	Qty                   int      `json:"qty"`
	SizeLength            float64  `json:"size_length"`
	SizeWidth             float64  `json:"size_width"`
	Notes                 string   `json:"notes"`
	Status                string   `json:"status"`
	EditorID              *string  `json:"editor_id"`
	VendorID              *string  `json:"vendor_id"`
	VendorName            *string  `json:"vendor_name"`
	TotalPrice            *float64 `json:"total_price"`
	SPKDate               *string  `json:"spk_date"`
	OperatorID            *string  `json:"operator_id"`
	OperatorName          *string  `json:"operator_name"`
	OperatorReward        *float64 `json:"operator_reward"`
	InstalledAt           *string  `json:"installed_at"`
	DesignLink            *string  `json:"design_link"`
	PhotoInstalled        *string  `json:"photo_installed"`
	DocumentationImageURL *string  `json:"documentation_image_url"`
	IsDesignApproved      *bool    `json:"is_design_approved"`
	IsStoreApproved       *bool    `json:"is_store_approved"`
}

func (h *BrandingProxyHandler) CreateBanner(c *gin.Context) {
	var req BrandingBannerRequest
	var uploadedAttachmentURL string

	// Handle multipart form data if file uploaded
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		req.ID = c.PostForm("id")
		req.ShopID = c.PostForm("shop_id")
		req.ShopName = c.PostForm("shop_name")
		req.StoreName = c.PostForm("store_name")
		req.SalesName = c.PostForm("sales_name")
		req.Type = c.PostForm("type")
		req.IsBihun = c.PostForm("is_bihun") == "true"
		if qtyStr := c.PostForm("qty"); qtyStr != "" {
			req.Qty, _ = strconv.Atoi(qtyStr)
		}
		if lStr := c.PostForm("size_length"); lStr != "" {
			req.SizeLength, _ = strconv.ParseFloat(lStr, 64)
		}
		if wStr := c.PostForm("size_width"); wStr != "" {
			req.SizeWidth, _ = strconv.ParseFloat(wStr, 64)
		}
		req.Notes = c.PostForm("notes")
		req.Status = c.PostForm("status")
		if docURL := c.PostForm("documentation_image_url"); docURL != "" {
			req.DocumentationImageURL = &docURL
		}

		// Process attachment file (logo/foto/pdf pendukung)
		_, header, err := c.Request.FormFile("attachment")
		if err != nil && !errors.Is(err, http.ErrMissingFile) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File lampiran tidak dapat dibaca"})
			return
		}
		if header != nil {
			savedPath, err := utils.SaveUploadedImageAsWebP(header, "./uploads/banners/attachments", 1600, 80)
			if err != nil || savedPath == "" {
				log.Printf("[BrandingHandler] Gagal menyimpan lampiran banner: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file lampiran"})
				return
			}
			uploadedAttachmentURL = savedPath
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.DocumentationImageURL != nil {
			uploadedAttachmentURL = *req.DocumentationImageURL
		}
	}

	if req.ShopName == "" && req.StoreName != "" {
		req.ShopName = req.StoreName
	}
	if strings.TrimSpace(req.ShopName) == "" {
		req.ShopName = "Toko (Tanpa Nama)"
	}

	// Resolve Employee from JWT Token
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas pengguna tidak ditemukan"})
		return
	}
	userID := userIDStr.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data karyawan tidak ditemukan"})
		return
	}
	// Mobile retries use the same client ID. A lost HTTP response must not
	// create a second banner order when the offline queue is replayed.
	var bannerID uuid.UUID
	if strings.TrimSpace(req.ID) != "" {
		bannerID, err = uuid.Parse(strings.TrimSpace(req.ID))
		if err != nil || bannerID == uuid.Nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID order spanduk tidak valid"})
			return
		}
		var existing domain.BannerOrder
		if err := h.localDB.Where("id = ?", bannerID).First(&existing).Error; err == nil {
			if existing.EmployeeID != employee.ID {
				c.JSON(http.StatusConflict, gin.H{"error": "ID order spanduk sudah digunakan"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Order Spanduk sudah ada", "data": existing})
			return
		} else if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa order spanduk"})
			return
		}
	} else {
		bannerID = uuid.New()
	}

	// Safe company ID resolution hierarchy:
	// 1. Employee's direct CompanyID
	// 2. JWT Token CompanyID
	// 3. Employee's Branch CompanyID
	// 4. Employee's OperatingBranch CompanyID
	// 5. Default Company from DB
	var companyID uuid.UUID
	if employee.CompanyID != nil && *employee.CompanyID != uuid.Nil {
		companyID = *employee.CompanyID
	}
	if companyID == uuid.Nil {
		if tokenCompID, exists := c.Get("companyID"); exists {
			if cid, ok := tokenCompID.(uuid.UUID); ok && cid != uuid.Nil {
				companyID = cid
			}
		}
	}
	if companyID == uuid.Nil && employee.BranchID != nil && *employee.BranchID != uuid.Nil {
		var branch domain.Branch
		if err := h.localDB.Where("id = ?", *employee.BranchID).First(&branch).Error; err == nil && branch.CompanyID != nil && *branch.CompanyID != uuid.Nil {
			companyID = *branch.CompanyID
		}
	}
	if companyID == uuid.Nil && employee.OperatingBranchID != nil && *employee.OperatingBranchID != uuid.Nil {
		var opBranch domain.Branch
		if err := h.localDB.Where("id = ?", *employee.OperatingBranchID).First(&opBranch).Error; err == nil && opBranch.CompanyID != nil && *opBranch.CompanyID != uuid.Nil {
			companyID = *opBranch.CompanyID
		}
	}
	if companyID == uuid.Nil {
		var defaultComp domain.Company
		if err := h.localDB.Order("created_at asc").First(&defaultComp).Error; err == nil {
			companyID = defaultComp.ID
		}
	}

	// A custom/offline shop name is valid for a banner, but it must not create an
	// incomplete store record. Only link an existing store selected by UUID.
	var store domain.Store
	shopIDTrimmed := strings.TrimSpace(req.ShopID)
	if shopIDTrimmed != "" && !strings.HasPrefix(strings.ToUpper(shopIDTrimmed), "OFFLINE") && !strings.HasPrefix(strings.ToUpper(shopIDTrimmed), "TEMP") {
		if parsedUUID, parseErr := uuid.Parse(shopIDTrimmed); parseErr == nil && parsedUUID != uuid.Nil {
			if err := h.localDB.Where("id = ?", parsedUUID).First(&store).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					// Older mobile versions treat 409 as a completed sync and drop
					// the queued banner, so use a retryable error instead.
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Toko belum tersedia di server. Sinkronkan data toko terlebih dahulu"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa data toko"})
				}
				return
			}
		}
	}

	if companyID == uuid.Nil && store.CompanyID != uuid.Nil {
		companyID = store.CompanyID
	}

	size := req.SizeLength * req.SizeWidth
	if size <= 0 {
		size = 1
	}
	if req.Type == "" {
		req.Type = "Custom"
	}
	if req.Qty <= 0 {
		req.Qty = 1
	}

	salesName := employee.FirstName
	if employee.LastName != "" {
		salesName = employee.FirstName + " " + employee.LastName
	}
	if req.SalesName != "" {
		salesName = req.SalesName
	}

	now := time.Now()
	notes := req.Notes

	var storeID *uuid.UUID
	if store.ID != uuid.Nil {
		storeID = &store.ID
	}

	campaignName := "Branding Shop: " + req.ShopName
	if len(campaignName) > 250 {
		campaignName = campaignName[:250]
	}

	storeName := req.ShopName
	if len(storeName) > 250 {
		storeName = storeName[:250]
	}

	location := store.Address
	if strings.TrimSpace(location) == "" {
		location = "-"
	}
	if len(location) > 250 {
		location = location[:250]
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" || strings.Contains(status, "offline") {
		status = "request"
	}

	newBanner := domain.BannerOrder{
		ID:                    bannerID,
		CompanyID:             companyID,
		EmployeeID:            employee.ID,
		StoreID:               storeID,
		CampaignName:          campaignName,
		StoreName:             storeName,
		Location:              location,
		BannerType:            req.Type,
		Size:                  size,
		SizeLength:            req.SizeLength,
		SizeWidth:             req.SizeWidth,
		Qty:                   req.Qty,
		IsBihun:               req.IsBihun,
		Notes:                 &notes,
		DocumentationImageURL: uploadedAttachmentURL,
		Status:                status,
		MySQLSalesName:        &salesName,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := h.localDB.Create(&newBanner).Error; err != nil {
		log.Printf("[BrandingHandler] ERROR creating banner order in DB: %v (data: %+v)", err, newBanner)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat order spanduk: " + err.Error()})
		return
	}

	// Kirim notifikasi realtime ke manajemen branding (tab branding)
	if h.hub != nil {
		go func(banner domain.BannerOrder, empName string) {
			_ = h.hub.PublishNotification(context.Background(), sse.NotificationEvent{
				ID:        banner.ID.String(),
				Type:      "branding_new_order",
				Title:     "🎨 Order Branding Baru",
				Message:   fmt.Sprintf("Order spanduk baru dari %s untuk toko %s", empName, banner.StoreName),
				CompanyID: banner.CompanyID.String(),
				CreatedAt: banner.CreatedAt.Format(time.RFC3339),
				Data: map[string]interface{}{
					"banner_id":     banner.ID.String(),
					"store_name":    banner.StoreName,
					"employee_name": empName,
					"banner_type":   banner.BannerType,
					"qty":           banner.Qty,
				},
				// Dikirim hanya ke user yang sedang membuka tab branding dengan role berikut
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleValidatorDesain),
					string(domain.RoleValidatorToko),
					string(domain.RoleAdminPurchasing),
					string(domain.RoleAdminDesigner),
				},
			})
		}(newBanner, salesName)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order Spanduk berhasil dibuat",
		"data":    newBanner,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /branding/banners/:id
// ─────────────────────────────────────────────────────────────────────────────

func (h *BrandingProxyHandler) UpdateBanner(c *gin.Context) {
	id := c.Param("id")

	var req BrandingBannerRequest
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		req.ShopID = c.PostForm("shop_id")
		req.ShopName = c.PostForm("shop_name")
		req.SalesName = c.PostForm("sales_name")
		req.Type = c.PostForm("type")
		req.IsBihun = c.PostForm("is_bihun") == "true"
		if qtyStr := c.PostForm("qty"); qtyStr != "" {
			req.Qty, _ = strconv.Atoi(qtyStr)
		}
		if lStr := c.PostForm("size_length"); lStr != "" {
			req.SizeLength, _ = strconv.ParseFloat(lStr, 64)
		}
		if wStr := c.PostForm("size_width"); wStr != "" {
			req.SizeWidth, _ = strconv.ParseFloat(wStr, 64)
		}
		req.Notes = c.PostForm("notes")
		req.Status = c.PostForm("status")

		if editorID, ok := c.GetPostForm("editor_id"); ok {
			req.EditorID = &editorID
		}
		if vendorID, ok := c.GetPostForm("vendor_id"); ok {
			req.VendorID = &vendorID
		}
		if vendorName, ok := c.GetPostForm("vendor_name"); ok {
			req.VendorName = &vendorName
		}
		if tpStr := c.PostForm("total_price"); tpStr != "" {
			tp, _ := strconv.ParseFloat(tpStr, 64)
			req.TotalPrice = &tp
		}
		if spkDate, ok := c.GetPostForm("spk_date"); ok {
			req.SPKDate = &spkDate
		}
		if operatorID, ok := c.GetPostForm("operator_id"); ok {
			req.OperatorID = &operatorID
		}
		if operatorName, ok := c.GetPostForm("operator_name"); ok {
			req.OperatorName = &operatorName
		}
		if opRewardStr := c.PostForm("operator_reward"); opRewardStr != "" {
			opReward, _ := strconv.ParseFloat(opRewardStr, 64)
			req.OperatorReward = &opReward
		}
		if instAt, ok := c.GetPostForm("installed_at"); ok {
			req.InstalledAt = &instAt
		}
		if photoInst, ok := c.GetPostForm("photo_installed"); ok {
			req.PhotoInstalled = &photoInst
		}
		if designLink, ok := c.GetPostForm("design_link"); ok {
			req.DesignLink = &designLink
		}
		if isDesAppStr := c.PostForm("is_design_approved"); isDesAppStr != "" {
			isDesApp := isDesAppStr == "true"
			req.IsDesignApproved = &isDesApp
		}
		if isStoreAppStr := c.PostForm("is_store_approved"); isStoreAppStr != "" {
			isStoreApp := isStoreAppStr == "true"
			req.IsStoreApproved = &isStoreApp
		}
		if docURL, ok := c.GetPostForm("documentation_image_url"); ok {
			req.DocumentationImageURL = &docURL
		}

		// Process attachment file (logo/foto/pdf pendukung)
		_, header, err := c.Request.FormFile("attachment")
		if err != nil && !errors.Is(err, http.ErrMissingFile) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File lampiran tidak dapat dibaca"})
			return
		}
		if header != nil {
			savedPath, err := utils.SaveUploadedImageAsWebP(header, "./uploads/banners/attachments", 1600, 80)
			if err != nil || savedPath == "" {
				log.Printf("[BrandingHandler] Gagal menyimpan lampiran banner: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file lampiran"})
				return
			}
			req.DocumentationImageURL = &savedPath
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	var banner domain.BannerOrder
	query := h.localDB.Where("id = ?", id)
	if idInt, err := strconv.ParseInt(id, 10, 64); err == nil {
		query = h.localDB.Where("id = ? OR mysql_banner_id = ?", id, idInt)
	}

	if err := query.First(&banner).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Spanduk tidak ditemukan"})
		return
	}

	notes := req.Notes
	size := req.SizeLength * req.SizeWidth
	if req.Qty <= 0 {
		req.Qty = 1
	}

	updates := map[string]interface{}{
		"banner_type": req.Type,
		"is_bihun":    req.IsBihun,
		"qty":         req.Qty,
		"size_length": req.SizeLength,
		"size_width":  req.SizeWidth,
		"size":        size,
		"notes":       notes,
		"updated_at":  time.Now(),
	}

	cleanShopName := strings.TrimSpace(req.ShopName)
	if cleanShopName != "" {
		updates["store_name"] = cleanShopName
		updates["campaign_name"] = "Branding Shop: " + cleanShopName

		// Keep the existing link when only editing banner details. A different
		// custom shop name has no store link until a real store is selected.
		shopID := strings.TrimSpace(req.ShopID)
		if parsedID, err := uuid.Parse(shopID); err == nil && parsedID != uuid.Nil {
			var store domain.Store
			if err := h.localDB.Where("id = ?", parsedID).First(&store).Error; err == nil {
				updates["store_id"] = &store.ID
				if store.Address != "" && store.Address != "-" {
					updates["location"] = store.Address
				}
			} else if err != gorm.ErrRecordNotFound {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa data toko"})
				return
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Toko belum tersedia di server. Sinkronkan data toko terlebih dahulu"})
				return
			}
		} else if !strings.EqualFold(cleanShopName, strings.TrimSpace(banner.StoreName)) {
			updates["store_id"] = nil
		}
	}

	if req.SalesName != "" {
		updates["mysql_sales_name"] = req.SalesName
	}

	// Dynamic status update
	if req.Status != "" {
		updates["status"] = req.Status
		if strings.ToLower(req.Status) == "editing" && banner.EditingAt == nil {
			now := time.Now()
			updates["editing_at"] = &now
		}
	}

	// Dynamic Editor assignment
	if req.EditorID != nil {
		if *req.EditorID == "" || *req.EditorID == "null" || *req.EditorID == "0" {
			updates["editor_id"] = nil
			updates["mysql_editor_name"] = nil
		} else if editorUUID, err := uuid.Parse(*req.EditorID); err == nil {
			updates["editor_id"] = &editorUUID
			var editorEmp domain.Employee
			if err := h.localDB.Where("id = ?", editorUUID).First(&editorEmp).Error; err == nil {
				fullName := strings.TrimSpace(editorEmp.FirstName + " " + editorEmp.LastName)
				updates["mysql_editor_name"] = fullName
			}
		}
	}

	// Dynamic Vendor assignment
	if req.VendorID != nil {
		if *req.VendorID == "" || *req.VendorID == "null" || *req.VendorID == "0" {
			updates["vendor_id"] = nil
			updates["vendor_name"] = nil
		} else if vendorUUID, err := uuid.Parse(*req.VendorID); err == nil {
			updates["vendor_id"] = &vendorUUID
			var vendor domain.BrandingVendor
			if err := h.localDB.Where("id = ?", vendorUUID).First(&vendor).Error; err == nil {
				updates["vendor_name"] = vendor.Name
			}
		}
	}

	if req.VendorName != nil && *req.VendorName != "" {
		updates["vendor_name"] = *req.VendorName
	}

	if req.TotalPrice != nil {
		updates["total_price"] = *req.TotalPrice
	}

	if req.SPKDate != nil {
		updates["spk_date"] = *req.SPKDate
	}

	// Dynamic Operator assignment
	if req.OperatorID != nil {
		if *req.OperatorID == "" || *req.OperatorID == "null" || *req.OperatorID == "0" || *req.OperatorID == "unassign" {
			updates["operator_id"] = nil
			updates["operator_name"] = nil
			updates["operator_received_at"] = nil
		} else if opUUID, err := uuid.Parse(*req.OperatorID); err == nil {
			updates["operator_id"] = &opUUID
			var opEmp domain.Employee
			if err := h.localDB.Where("id = ?", opUUID).First(&opEmp).Error; err == nil {
				fullName := strings.TrimSpace(opEmp.FirstName + " " + opEmp.LastName)
				updates["operator_name"] = fullName
			} else if req.OperatorName != nil && *req.OperatorName != "" {
				updates["operator_name"] = *req.OperatorName
			}
			if banner.OperatorReceivedAt == nil {
				now := time.Now()
				updates["operator_received_at"] = &now
			}
		}
	} else if req.OperatorName != nil {
		if *req.OperatorName == "" || *req.OperatorName == "null" {
			updates["operator_name"] = nil
		} else {
			updates["operator_name"] = *req.OperatorName
		}
	}

	if req.OperatorReward != nil {
		updates["operator_reward"] = *req.OperatorReward
	}

	// Dynamic Installation Date & Photo
	if req.InstalledAt != nil && *req.InstalledAt != "" {
		if t, err := time.Parse("2006-01-02", *req.InstalledAt); err == nil {
			updates["installed_at"] = &t
		} else if t, err := time.Parse(time.RFC3339, *req.InstalledAt); err == nil {
			updates["installed_at"] = &t
		}
	}

	if req.PhotoInstalled != nil && *req.PhotoInstalled != "" {
		updates["photo_installed"] = *req.PhotoInstalled
	}

	// Dynamic Design Link
	if req.DesignLink != nil && *req.DesignLink != "" {
		updates["design_link"] = *req.DesignLink
		updates["design_url"] = *req.DesignLink
	}

	if req.DocumentationImageURL != nil {
		updates["documentation_image_url"] = *req.DocumentationImageURL
	}

	// Status-aware Approval Flag Synchronization
	stLower := strings.ToLower(req.Status)
	if req.IsDesignApproved != nil {
		updates["is_design_approved"] = *req.IsDesignApproved
	} else if stLower == "validation" || stLower == "editing" || stLower == "request" {
		// Ketika status dikembalikan secara manual ke validasi/editing/request tanpa flag approval,
		// reset persetujuan desain agar tidak memicu auto-promotion ke vendor
		updates["is_design_approved"] = false
		if stLower == "editing" || stLower == "request" {
			updates["design_validator_id"] = nil
			updates["design_validator_name"] = nil
			updates["design_validator_position"] = nil
			updates["design_approved_at"] = nil
		}
	}

	if req.IsStoreApproved != nil {
		updates["is_store_approved"] = *req.IsStoreApproved
	} else if stLower == "validation" || stLower == "editing" || stLower == "request" {
		// Ketika status dikembalikan secara manual ke validasi/editing/request tanpa flag approval,
		// reset persetujuan toko agar tidak memicu auto-promotion ke vendor
		updates["is_store_approved"] = false
		if stLower == "editing" || stLower == "request" {
			updates["store_validator_id"] = nil
			updates["store_validator_name"] = nil
			updates["store_validator_position"] = nil
			updates["store_approved_at"] = nil
		}
	}

	if err := h.localDB.Model(&banner).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui spanduk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Spanduk berhasil diperbarui"})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /branding/banners/:id
// ─────────────────────────────────────────────────────────────────────────────

func (h *BrandingProxyHandler) DeleteBanner(c *gin.Context) {
	id := c.Param("id")

	query := h.localDB.Where("id = ?", id)
	if idInt, err := strconv.ParseInt(id, 10, 64); err == nil {
		query = h.localDB.Where("id = ? OR mysql_banner_id = ?", id, idInt)
	}

	var banner domain.BannerOrder
	if err := query.First(&banner).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Spanduk tidak ditemukan"})
		return
	}

	if err := h.localDB.Delete(&banner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus spanduk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Spanduk berhasil dihapus"})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /branding/agents
// ─────────────────────────────────────────────────────────────────────────────

type SalesAgentResponse struct {
	ID                       string  `json:"id"`
	Name                     string  `json:"name"`
	Email                    string  `json:"email"`
	Phone                    *string `json:"phone"`
	Role                     string  `json:"role"`
	AreaCode                 *string `json:"area"`
	IsActive                 bool    `json:"is_active"`
	TargetDailyInstallations int     `json:"target_daily_installations"`
}

func (h *BrandingProxyHandler) GetSalesAgents(c *gin.Context) {
	var agents []SalesAgentResponse
	branchID := c.Query("branch_id")

	query := h.localDB.Table("employees").
		Select(`
			employees.id::text as id,
			COALESCE(employees.first_name || ' ' || COALESCE(employees.last_name, ''), employees.first_name) as name,
			COALESCE(u.email, '') as email,
			employees.phone_number as phone,
			COALESCE(jp.title, u.role) as role,
			branches.name as area_code,
			COALESCE(u.is_active, true) as is_active,
			COALESCE(employees.target_daily_installations, 0) as target_daily_installations
		`).
		Joins("JOIN users u ON u.id = employees.user_id").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Joins("LEFT JOIN job_positions jp ON jp.id = employees.job_position_id").
		Joins("LEFT JOIN departments d ON d.id = employees.department_id").
		Where("COALESCE(u.is_active, true) = true").
		Where("COALESCE(employees.employment_status, '') NOT IN ('Inactive', 'INACTIVE', 'Resigned', 'RESIGNED', 'Terminated', 'TERMINATED')").
		Where("COALESCE(employees.employment_status, '') NOT ILIKE '%harian%'").
		Where("COALESCE(jp.title, '') NOT ILIKE '%operator%' AND COALESCE(jp.title, '') NOT ILIKE '%driver%' AND COALESCE(jp.title, '') NOT ILIKE '%sopir%' AND COALESCE(jp.title, '') NOT ILIKE '%kuli%' AND COALESCE(jp.title, '') NOT ILIKE '%gudang%' AND COALESCE(jp.title, '') NOT ILIKE '%auditor%' AND COALESCE(jp.title, '') NOT ILIKE '%validator%' AND COALESCE(jp.title, '') NOT ILIKE '%editor%' AND COALESCE(jp.title, '') NOT ILIKE '%security%' AND COALESCE(jp.title, '') NOT ILIKE '%cleaning%' AND COALESCE(jp.title, '') NOT ILIKE '%kasir%' AND COALESCE(jp.title, '') NOT ILIKE '%cashier%'").
		Where("(jp.title ILIKE '%sales%' OR jp.title ILIKE '%motoris%' OR jp.title ILIKE '%booster%' OR jp.title ILIKE '%marketing%' OR jp.title ILIKE '%modern trade%' OR jp.title ILIKE '%take order%' OR jp.title ILIKE '%task order%' OR jp.title ILIKE '%kam%' OR jp.title ~* '\\mmt\\M' OR jp.title ~* '\\mto\\M' OR d.name ILIKE '%sales%' OR d.name ILIKE '%marketing%')")

	if branchID != "" && branchID != "ALL_BRANCHES" {
		query = query.Where("employees.branch_id = ?", branchID)
	}

	err := query.Find(&agents).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data sales agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"source":  "postgres",
		"data":    agents,
		"total":   len(agents),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /branding/stats
// ─────────────────────────────────────────────────────────────────────────────

func (h *BrandingProxyHandler) GetStats(c *gin.Context) {
	branchID := c.Query("branch_id")
	var stats struct {
		TotalBanners     int64   `gorm:"column:total_banners"`
		InstalledBanners int64   `gorm:"column:installed_banners"`
		ReadyBanners     int64   `gorm:"column:ready_banners"`
		PendingBanners   int64   `gorm:"column:pending_banners"`
		TotalRevenue     float64 `gorm:"column:total_revenue"`
	}

	var totalShops int64

	if branchID != "" && branchID != "ALL_BRANCHES" {
		h.localDB.Raw(`
			SELECT
				COUNT(banner_orders.id) AS total_banners,
				SUM(CASE WHEN LOWER(banner_orders.status) IN ('installed', 'terpasang') THEN 1 ELSE 0 END) AS installed_banners,
				SUM(CASE WHEN LOWER(banner_orders.status) IN ('ready_for_operator', 'diterima_operator') THEN 1 ELSE 0 END) AS ready_banners,
				SUM(CASE WHEN LOWER(banner_orders.status) IN ('ready_for_operator', 'diterima_operator') THEN 1 ELSE 0 END) AS pending_banners,
				COALESCE(SUM(CASE WHEN LOWER(banner_orders.status) IN ('installed', 'terpasang') THEN banner_orders.total_price ELSE 0 END), 0) AS total_revenue
			FROM banner_orders
			LEFT JOIN stores ON stores.id = banner_orders.store_id
			LEFT JOIN employees sales ON sales.id = banner_orders.employee_id
			LEFT JOIN employees store_sales ON store_sales.id = stores.assigned_employee_id
			WHERE COALESCE(sales.branch_id::text, store_sales.branch_id::text, '') = ?
		`, branchID).Scan(&stats)

		h.localDB.Table("stores").
			Joins("LEFT JOIN employees sales ON sales.id = stores.assigned_employee_id").
			Where("stores.is_active = true AND sales.branch_id = ?", branchID).
			Count(&totalShops)
	} else {
		h.localDB.Raw(`
			SELECT
				COUNT(*) AS total_banners,
				SUM(CASE WHEN LOWER(status) IN ('installed', 'terpasang') THEN 1 ELSE 0 END) AS installed_banners,
				SUM(CASE WHEN LOWER(status) IN ('ready_for_operator', 'diterima_operator') THEN 1 ELSE 0 END) AS ready_banners,
				SUM(CASE WHEN LOWER(status) IN ('ready_for_operator', 'diterima_operator') THEN 1 ELSE 0 END) AS pending_banners,
				COALESCE(SUM(CASE WHEN LOWER(status) IN ('installed', 'terpasang') THEN total_price ELSE 0 END), 0) AS total_revenue
			FROM banner_orders
		`).Scan(&stats)

		h.localDB.Table("stores").Where("is_active = true").Count(&totalShops)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"total_banners":        stats.TotalBanners,
			"installed_banners":    stats.InstalledBanners,
			"ready_banners":        stats.ReadyBanners,
			"pending_banners":      stats.ReadyBanners,
			"total_revenue_banner": stats.TotalRevenue,
			"total_expense_banner": stats.TotalRevenue,
			"total_shops":          totalShops,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /branding/operator/performance
// ─────────────────────────────────────────────────────────────────────────────

type OperatorPerformanceStat struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	Area             string  `json:"area"`
	BannerTerpasang  int     `json:"banner_terpasang"`
	TotalKomisi      float64 `json:"total_komisi"`
	KomisiDibayar    float64 `json:"komisi_dibayar"`
	KomisiBelumBayar float64 `json:"komisi_belum_dibayar"`
}

func (h *BrandingProxyHandler) GetOperatorPerformance(c *gin.Context) {
	var stats []OperatorPerformanceStat
	month, year := c.Query("month"), c.Query("year")
	periodFilter := ""
	args := []interface{}{}
	if date := c.Query("date"); date != "" {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tanggal komisi tidak valid"})
			return
		}
		periodFilter = " AND DATE(b.installed_at) = ?"
		args = append(args, date)
	}
	if periodFilter == "" && month != "" && year != "" {
		monthNum, monthErr := strconv.Atoi(month)
		yearNum, yearErr := strconv.Atoi(year)
		if monthErr != nil || yearErr != nil || monthNum < 1 || monthNum > 12 || yearNum < 2000 || yearNum > 2100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bulan atau tahun komisi tidak valid"})
			return
		}
		periodFilter = " AND EXTRACT(MONTH FROM b.installed_at) = ? AND EXTRACT(YEAR FROM b.installed_at) = ?"
		args = append(args, monthNum, yearNum)
	}

	query := `
		SELECT 
			e.id::text as id,
			COALESCE(e.first_name || ' ' || COALESCE(e.last_name, ''), e.first_name) as name,
			COALESCE(u.email, '') as email,
			COUNT(b.id) as banner_terpasang,
			COALESCE(SUM(b.operator_reward), 0) as total_komisi,
			COALESCE(SUM(b.operator_reward) FILTER (WHERE b.is_paid = true), 0) as komisi_dibayar,
			COALESCE(SUM(b.operator_reward) FILTER (WHERE b.is_paid = false OR b.is_paid IS NULL), 0) as komisi_belum_dibayar,
			COALESCE(
				NULLIF((
					SELECT STRING_AGG(DISTINCT NULLIF(s.regency, ''), ', ')
					FROM banner_orders b2
					JOIN stores s ON s.id = b2.store_id
					WHERE b2.operator_id = e.id AND LOWER(b2.status) IN ('installed', 'terpasang')
				), ''),
				branches.name,
				'Jawa Timur & Sekitarnya'
			) as area
		FROM employees e
		JOIN users u ON u.id = e.user_id
		LEFT JOIN branches ON branches.id = e.branch_id
		LEFT JOIN job_positions jp ON jp.id = e.job_position_id
		LEFT JOIN banner_orders b ON b.operator_id = e.id AND LOWER(b.status) IN ('installed', 'terpasang')` + periodFilter + `
		WHERE u.role IN ('operator')
		   OR LOWER(COALESCE(u.role::text, '')) = 'operator'
		   OR LOWER(COALESCE(jp.title, '')) LIKE '%operator%'
		   OR LOWER(COALESCE(jp.title, '')) LIKE '%pasang%'
		   OR e.target_daily_installations > 0
		   OR e.id IN (SELECT DISTINCT operator_id FROM banner_orders WHERE operator_id IS NOT NULL)
		GROUP BY e.id, e.first_name, e.last_name, u.email, branches.name
		ORDER BY banner_terpasang DESC, name ASC
	`
	err := h.localDB.Raw(query, args...).Scan(&stats).Error

	if err != nil {
		log.Printf("[BrandingHandler] GetOperatorPerformance error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data performa operator"})
		return
	}

	for i := range stats {
		if stats[i].Area == "" {
			stats[i].Area = "Semua Area"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    stats,
		"total":   len(stats),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /branding/editor/performance
// ─────────────────────────────────────────────────────────────────────────────

type EditorPerformanceStat struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Bulanan      int    `json:"bulanan"`
	HarianBaru   int    `json:"harian_baru"`
	HarianRevisi int    `json:"harian_revisi"`
	IsActive     bool   `json:"is_active"`
}

func (h *BrandingProxyHandler) GetEditorPerformance(c *gin.Context) {
	var stats []EditorPerformanceStat
	targetDate := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	if _, err := time.Parse("2006-01-02", targetDate); err != nil {
		targetDate = time.Now().Format("2006-01-02")
	}

	err := h.localDB.Raw(`
		SELECT 
			e.id::text as id,
			COALESCE(e.first_name || ' ' || COALESCE(e.last_name, ''), e.first_name) as name,
			COALESCE(u.email, '') as email,
			COUNT(b.id) as bulanan,
			COUNT(CASE WHEN TO_CHAR(COALESCE(b.editing_finished_at, b.editing_at, b.created_at), 'YYYY-MM-DD') = ? AND (b.revision_note IS NULL OR b.revision_note = '') THEN 1 END) as harian_baru,
			COUNT(CASE WHEN TO_CHAR(COALESCE(b.editing_finished_at, b.editing_at, b.created_at), 'YYYY-MM-DD') = ? AND (b.revision_note IS NOT NULL AND b.revision_note != '') THEN 1 END) as harian_revisi,
			COALESCE(u.is_active, true) as is_active
		FROM employees e
		JOIN users u ON u.id = e.user_id
		LEFT JOIN job_positions jp ON jp.id = e.job_position_id
		LEFT JOIN banner_orders b ON b.editor_id = e.id
		WHERE (u.role IN ('admin_designer', 'editor')
		   OR LOWER(COALESCE(u.role::text, '')) IN ('admin_designer', 'editor')
		   OR LOWER(COALESCE(jp.title, '')) LIKE '%editor%'
		   OR (LOWER(COALESCE(jp.title, '')) LIKE '%desain%' AND LOWER(COALESCE(u.role::text, '')) NOT IN ('validator_desain'))
		   OR (e.id IN (SELECT DISTINCT editor_id FROM banner_orders WHERE editor_id IS NOT NULL) AND LOWER(COALESCE(u.role::text, '')) NOT IN ('validator_desain')))
		   AND LOWER(COALESCE(u.role::text, '')) NOT IN ('validator_desain')
		GROUP BY e.id, e.first_name, e.last_name, u.email, u.is_active
		ORDER BY bulanan DESC, name ASC
	`, targetDate, targetDate).Scan(&stats).Error

	if err != nil {
		log.Printf("[BrandingHandler] GetEditorPerformance error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data performa editor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"date":    targetDate,
		"data":    stats,
		"total":   len(stats),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// CRUD Editors (PostgreSQL Users & Employees with role admin_designer)
// ─────────────────────────────────────────────────────────────────────────────

type EditorDTO struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Phone    *string `json:"phone"`
	AreaCode *string `json:"area"`
	IsActive bool    `json:"is_active"`
}

func (h *BrandingProxyHandler) GetEditors(c *gin.Context) {
	var editors []EditorDTO

	err := h.localDB.Table("employees").
		Select(`
			employees.id::text as id,
			COALESCE(employees.first_name || ' ' || COALESCE(employees.last_name, ''), employees.first_name) as name,
			COALESCE(u.email, '') as email,
			employees.phone_number as phone,
			branches.name as area_code,
			u.is_active as is_active
		`).
		Joins("JOIN users u ON u.id = employees.user_id").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Joins("LEFT JOIN job_positions jp ON jp.id = employees.job_position_id").
		Where(`
			(u.role IN ('admin_designer', 'editor')
			OR LOWER(COALESCE(u.role::text, '')) IN ('admin_designer', 'editor')
			OR LOWER(COALESCE(jp.title, '')) LIKE '%editor%'
			OR (LOWER(COALESCE(jp.title, '')) LIKE '%desain%' AND LOWER(COALESCE(u.role::text, '')) NOT IN ('validator_desain'))
			OR (employees.id IN (SELECT DISTINCT editor_id FROM banner_orders WHERE editor_id IS NOT NULL) AND LOWER(COALESCE(u.role::text, '')) NOT IN ('validator_desain')))
			AND LOWER(COALESCE(u.role::text, '')) NOT IN ('validator_desain')
		`).
		Order("name ASC").
		Find(&editors).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data editor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "data": editors, "total": len(editors)})
}

func (h *BrandingProxyHandler) CreateEditor(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		AreaCode string `json:"area"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		safeName := strings.ToLower(strings.ReplaceAll(req.Name, " ", "."))
		email = fmt.Sprintf("%s.%d@demo.local", safeName, time.Now().Unix())
	}

	// Check if user already exists
	var existingUser domain.User
	if err := h.localDB.Where("LOWER(email) = ?", email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email sudah terdaftar"})
		return
	}

	var defaultCompany domain.Company
	if err := h.localDB.Where("name = ?", os.Getenv("TRIAL_COMPANY_NAME")).First(&defaultCompany).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Perusahaan trial tidak ditemukan"})
		return
	}

	trialStaffPassword := os.Getenv("TRIAL_STAFF_PASSWORD")
	if trialStaffPassword == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Kata sandi akun trial belum dikonfigurasi"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(trialStaffPassword), bcrypt.DefaultCost)
	newUser := domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleAdminDesigner,
		IsActive:     true,
		CompanyID:    &defaultCompany.ID,
	}

	if err := h.localDB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat akun editor"})
		return
	}

	firstName := req.Name
	lastName := ""
	parts := strings.Fields(req.Name)
	if len(parts) > 1 {
		firstName = parts[0]
		lastName = strings.Join(parts[1:], " ")
	}

	var branchID *uuid.UUID
	if req.AreaCode != "" {
		var branch domain.Branch
		if err := h.localDB.Where("LOWER(name) = ? OR code = ?", strings.ToLower(strings.TrimSpace(req.AreaCode)), req.AreaCode).First(&branch).Error; err == nil {
			branchID = &branch.ID
		}
	}

	newEmp := domain.Employee{
		ID:               uuid.New(),
		UserID:           newUser.ID,
		CompanyID:        &defaultCompany.ID,
		BranchID:         branchID,
		EmployeeIDNumber: fmt.Sprintf("ED-%d", time.Now().Unix()%100000),
		FirstName:        firstName,
		LastName:         lastName,
		PhoneNumber:      req.Phone,
		JoinDate:         time.Now(),
		EmploymentStatus: "PERMANENT",
	}

	if err := h.localDB.Create(&newEmp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat profil editor"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Editor berhasil ditambahkan",
		"data": EditorDTO{
			ID:       newEmp.ID.String(),
			Name:     req.Name,
			Email:    email,
			Phone:    &req.Phone,
			AreaCode: &req.AreaCode,
			IsActive: true,
		},
	})
}

func (h *BrandingProxyHandler) UpdateEditor(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		AreaCode string `json:"area"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var emp domain.Employee
	if err := h.localDB.Where("id = ?", id).First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Editor tidak ditemukan"})
		return
	}

	if req.Name != "" {
		parts := strings.Fields(req.Name)
		emp.FirstName = parts[0]
		if len(parts) > 1 {
			emp.LastName = strings.Join(parts[1:], " ")
		} else {
			emp.LastName = ""
		}
	}
	if req.Phone != "" {
		emp.PhoneNumber = req.Phone
	}
	if req.AreaCode != "" {
		var branch domain.Branch
		if err := h.localDB.Where("LOWER(name) = ? OR code = ?", strings.ToLower(strings.TrimSpace(req.AreaCode)), req.AreaCode).First(&branch).Error; err == nil {
			emp.BranchID = &branch.ID
		}
	}
	h.localDB.Save(&emp)

	if req.Email != "" || req.IsActive != emp.User.IsActive {
		h.localDB.Model(&domain.User{}).Where("id = ?", emp.UserID).Updates(map[string]interface{}{
			"email":     req.Email,
			"is_active": req.IsActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Editor berhasil diupdate"})
}

func (h *BrandingProxyHandler) DeleteEditor(c *gin.Context) {
	id := c.Param("id")

	var emp domain.Employee
	if err := h.localDB.Where("id = ?", id).First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Editor tidak ditemukan"})
		return
	}

	h.localDB.Model(&domain.User{}).Where("id = ?", emp.UserID).Update("is_active", false)
	c.JSON(http.StatusOK, gin.H{"message": "Editor berhasil dinonaktifkan"})
}

// ─────────────────────────────────────────────────────────────────────────────
// CRUD Operators (PostgreSQL Users & Employees with role operator)
// ─────────────────────────────────────────────────────────────────────────────

type OperatorDTO struct {
	ID                       string  `json:"id"`
	Name                     string  `json:"name"`
	Email                    string  `json:"email"`
	Phone                    *string `json:"phone"`
	AreaCode                 *string `json:"area"`
	IsActive                 bool    `json:"is_active"`
	TargetDailyInstallations int     `json:"target_daily_installations"`
}

func (h *BrandingProxyHandler) GetOperators(c *gin.Context) {
	var operators []OperatorDTO

	err := h.localDB.Table("employees").
		Select(`
			employees.id::text as id,
			COALESCE(employees.first_name || ' ' || COALESCE(employees.last_name, ''), employees.first_name) as name,
			COALESCE(u.email, '') as email,
			employees.phone_number as phone,
			branches.name as area_code,
			u.is_active as is_active,
			COALESCE(employees.target_daily_installations, 0) as target_daily_installations
		`).
		Joins("JOIN users u ON u.id = employees.user_id").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Joins("LEFT JOIN job_positions jp ON jp.id = employees.job_position_id").
		Where(`
			u.role IN ('operator')
			OR LOWER(COALESCE(u.role::text, '')) = 'operator'
			OR LOWER(COALESCE(jp.title, '')) LIKE '%operator%'
			OR LOWER(COALESCE(jp.title, '')) LIKE '%pasang%'
			OR employees.target_daily_installations > 0
			OR employees.id IN (SELECT DISTINCT operator_id FROM banner_orders WHERE operator_id IS NOT NULL)
		`).
		Order("name ASC").
		Find(&operators).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data operator"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "data": operators, "total": len(operators)})
}

func (h *BrandingProxyHandler) CreateOperator(c *gin.Context) {
	var req struct {
		Name                     string `json:"name" binding:"required"`
		Email                    string `json:"email"`
		Phone                    string `json:"phone"`
		AreaCode                 string `json:"area"`
		TargetDailyInstallations int    `json:"target_daily_installations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		safeName := strings.ToLower(strings.ReplaceAll(req.Name, " ", "."))
		email = fmt.Sprintf("%s.%d@demo.local", safeName, time.Now().Unix())
	}

	var defaultCompany domain.Company
	if err := h.localDB.Where("name = ?", os.Getenv("TRIAL_COMPANY_NAME")).First(&defaultCompany).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Perusahaan trial tidak ditemukan"})
		return
	}

	trialStaffPassword := os.Getenv("TRIAL_STAFF_PASSWORD")
	if trialStaffPassword == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Kata sandi akun trial belum dikonfigurasi"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(trialStaffPassword), bcrypt.DefaultCost)
	newUser := domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleOperator,
		IsActive:     true,
		CompanyID:    &defaultCompany.ID,
	}

	if err := h.localDB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat akun operator"})
		return
	}

	firstName := req.Name
	lastName := ""
	parts := strings.Fields(req.Name)
	if len(parts) > 1 {
		firstName = parts[0]
		lastName = strings.Join(parts[1:], " ")
	}

	var branchID *uuid.UUID
	if req.AreaCode != "" {
		var branch domain.Branch
		if err := h.localDB.Where("LOWER(name) = ? OR code = ?", strings.ToLower(strings.TrimSpace(req.AreaCode)), req.AreaCode).First(&branch).Error; err == nil {
			branchID = &branch.ID
		}
	}

	newEmp := domain.Employee{
		ID:                       uuid.New(),
		UserID:                   newUser.ID,
		CompanyID:                &defaultCompany.ID,
		BranchID:                 branchID,
		EmployeeIDNumber:         fmt.Sprintf("OP-%d", time.Now().Unix()%100000),
		FirstName:                firstName,
		LastName:                 lastName,
		PhoneNumber:              req.Phone,
		JoinDate:                 time.Now(),
		EmploymentStatus:         "PERMANENT",
		TargetDailyInstallations: req.TargetDailyInstallations,
	}

	if err := h.localDB.Create(&newEmp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat profil operator"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Operator berhasil ditambahkan",
		"data": OperatorDTO{
			ID:                       newEmp.ID.String(),
			Name:                     req.Name,
			Email:                    email,
			Phone:                    &req.Phone,
			AreaCode:                 &req.AreaCode,
			IsActive:                 true,
			TargetDailyInstallations: req.TargetDailyInstallations,
		},
	})
}

func (h *BrandingProxyHandler) UpdateOperator(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name                     string `json:"name"`
		Email                    string `json:"email"`
		Phone                    string `json:"phone"`
		AreaCode                 string `json:"area"`
		IsActive                 bool   `json:"is_active"`
		TargetDailyInstallations int    `json:"target_daily_installations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var emp domain.Employee
	if err := h.localDB.Where("id = ?", id).First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Operator tidak ditemukan"})
		return
	}

	if req.Name != "" {
		parts := strings.Fields(req.Name)
		emp.FirstName = parts[0]
		if len(parts) > 1 {
			emp.LastName = strings.Join(parts[1:], " ")
		} else {
			emp.LastName = ""
		}
	}
	if req.Phone != "" {
		emp.PhoneNumber = req.Phone
	}
	if req.AreaCode != "" {
		var branch domain.Branch
		if err := h.localDB.Where("LOWER(name) = ? OR code = ?", strings.ToLower(strings.TrimSpace(req.AreaCode)), req.AreaCode).First(&branch).Error; err == nil {
			emp.BranchID = &branch.ID
		}
	}
	emp.TargetDailyInstallations = req.TargetDailyInstallations
	h.localDB.Save(&emp)

	if req.Email != "" || req.IsActive != emp.User.IsActive {
		h.localDB.Model(&domain.User{}).Where("id = ?", emp.UserID).Updates(map[string]interface{}{
			"email":     req.Email,
			"is_active": req.IsActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operator berhasil diupdate"})
}

func (h *BrandingProxyHandler) DeleteOperator(c *gin.Context) {
	id := c.Param("id")

	var emp domain.Employee
	if err := h.localDB.Where("id = ?", id).First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Operator tidak ditemukan"})
		return
	}

	h.localDB.Model(&domain.User{}).Where("id = ?", emp.UserID).Update("is_active", false)
	c.JSON(http.StatusOK, gin.H{"message": "Operator berhasil dinonaktifkan"})
}

func (h *BrandingProxyHandler) SetOperatorTarget(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		TargetDailyInstallations int `json:"target_daily_installations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.localDB.Model(&domain.Employee{}).Where("id = ?", id).Update("target_daily_installations", req.TargetDailyInstallations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate target harian operator"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Target harian operator berhasil diupdate", "target": req.TargetDailyInstallations})
}

// ─────────────────────────────────────────────────────────────────────────────
// State Machine Endpoints
// ─────────────────────────────────────────────────────────────────────────────

func (h *BrandingProxyHandler) UploadDesign(c *gin.Context) {
	bannerID := c.Param("id")

	file, header, err := c.Request.FormFile("design_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca file desain"})
		return
	}
	defer file.Close()

	// Validasi ekstensi file yang diizinkan
	originalFileName := header.Filename
	ext := strings.ToLower(originalFileName)
	allowedExts := []string{".png", ".jpg", ".jpeg", ".pdf"}
	isAllowed := false
	for _, allowed := range allowedExts {
		if strings.HasSuffix(ext, allowed) {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung. Gunakan PNG, JPG, JPEG, atau PDF"})
		return
	}

	userIDStr, _ := c.Get("userID")
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userIDStr.(uuid.UUID))
	if err != nil || employee == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas tidak ditemukan"})
		return
	}

	// Pastikan GDrive service tersedia
	if h.gdrive == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Layanan penyimpanan Google Drive tidak tersedia"})
		return
	}

	// Buat nama file GDrive: [bannerID]_namaasli.ext
	// Prefix bannerID mencegah collision dan memudahkan pencarian
	gdriveName := fmt.Sprintf("%s_%s", bannerID, originalFileName)

	designLink, err := h.gdrive.UploadFileWithName(file, gdriveName)
	if err != nil {
		log.Printf("[BrandingHandler] Gagal upload desain '%s' ke GDrive: %v", originalFileName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengunggah file desain ke Google Drive: " + err.Error()})
		return
	}

	log.Printf("[BrandingHandler] ✅ Desain '%s' berhasil diupload ke GDrive sebagai '%s': %s", originalFileName, gdriveName, designLink)

	now := time.Now()
	updates := map[string]interface{}{
		"design_url":          designLink,
		"design_link":         designLink,
		"design_file_name":    originalFileName, // nama asli file dipertahankan untuk validasi
		"status":              "validation",
		"editor_id":           employee.ID,
		"editing_at":          now,
		"editing_finished_at": now,
		"updated_at":          now,
	}

	query := h.localDB.Model(&domain.BannerOrder{}).Where("id = ?", bannerID)
	if idInt, err := strconv.ParseInt(bannerID, 10, 64); err == nil {
		query = h.localDB.Model(&domain.BannerOrder{}).Where("id = ? OR mysql_banner_id = ?", bannerID, idInt)
	}

	if err := query.Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update status banner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Desain berhasil diunggah dan status menjadi validation",
		"design_link":      designLink,
		"design_file_name": originalFileName,
	})
}

func (h *BrandingProxyHandler) AssignBanner(c *gin.Context) {
	bannerID := c.Param("id")

	var req struct {
		EditorID string `json:"editor_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Editor ID diperlukan"})
		return
	}

	editorUUID, err := uuid.Parse(req.EditorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Editor ID format tidak valid"})
		return
	}

	var editorEmp domain.Employee
	var editorFullName string
	if err := h.localDB.Where("id = ?", editorUUID).First(&editorEmp).Error; err == nil {
		editorFullName = strings.TrimSpace(editorEmp.FirstName + " " + editorEmp.LastName)
		if editorFullName == "" {
			editorFullName = editorEmp.FirstName
		}
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":            "editing",
		"editor_id":         editorUUID,
		"mysql_editor_name": editorFullName,
		"editing_at":        now,
		"updated_at":        now,
	}

	query := h.localDB.Model(&domain.BannerOrder{}).Where("id = ?", bannerID)
	if idInt, err := strconv.ParseInt(bannerID, 10, 64); err == nil {
		query = h.localDB.Model(&domain.BannerOrder{}).Where("id = ? OR mysql_banner_id = ?", bannerID, idInt)
	}

	if err := query.Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menugaskan desainer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Tugas berhasil diberikan kepada desainer",
		"editor_name": editorFullName,
	})
}

func (h *BrandingProxyHandler) AssignOperator(c *gin.Context) {
	bannerID := c.Param("id")

	var req struct {
		OperatorID *string `json:"operator_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid"})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"updated_at": now,
	}

	// Kasus 1: Kosongkan penugasan operator (Unassign)
	if req.OperatorID == nil || *req.OperatorID == "" || *req.OperatorID == "null" || *req.OperatorID == "unassign" || *req.OperatorID == "0" {
		updates["status"] = "ready_for_operator"
		updates["operator_id"] = nil
		updates["operator_name"] = nil
		updates["operator_received_at"] = nil

		query := h.localDB.Model(&domain.BannerOrder{}).Where("id = ?", bannerID)
		if idInt, err := strconv.ParseInt(bannerID, 10, 64); err == nil {
			query = h.localDB.Model(&domain.BannerOrder{}).Where("id = ? OR mysql_banner_id = ?", bannerID, idInt)
		}

		if err := query.Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengosongkan penugasan operator"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Penugasan operator berhasil dikosongkan (Siap Dipasang)", "operator_name": nil})
		return
	}

	// Kasus 2: Assign ke operator tertentu
	operatorUUID, err := uuid.Parse(*req.OperatorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Operator ID format tidak valid"})
		return
	}

	var operatorEmp domain.Employee
	var operatorName string
	if err := h.localDB.Where("id = ?", operatorUUID).First(&operatorEmp).Error; err == nil {
		operatorName = strings.TrimSpace(operatorEmp.FirstName + " " + operatorEmp.LastName)
		if operatorName == "" {
			operatorName = operatorEmp.FirstName
		}
	}

	updates["status"] = "diterima_operator"
	updates["operator_id"] = operatorUUID
	updates["operator_name"] = operatorName
	updates["operator_received_at"] = &now

	query := h.localDB.Model(&domain.BannerOrder{}).Where("id = ?", bannerID)
	if idInt, err := strconv.ParseInt(bannerID, 10, 64); err == nil {
		query = h.localDB.Model(&domain.BannerOrder{}).Where("id = ? OR mysql_banner_id = ?", bannerID, idInt)
	}

	if err := query.Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menugaskan operator"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Spanduk berhasil ditugaskan kepada operator", "operator_name": operatorName})
}

func (h *BrandingProxyHandler) ClaimBanner(c *gin.Context) {
	bannerID := c.Param("id")

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identitas tidak ditemukan"})
		return
	}
	userID := userIDStr.(uuid.UUID)
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
	if err != nil || employee == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data karyawan tidak ditemukan"})
		return
	}

	editorFullName := strings.TrimSpace(employee.FirstName + " " + employee.LastName)
	if editorFullName == "" {
		editorFullName = employee.FirstName
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":            "editing",
		"editor_id":         employee.ID,
		"mysql_editor_name": editorFullName,
		"editing_at":        now,
		"updated_at":        now,
	}

	query := h.localDB.Model(&domain.BannerOrder{}).Where("id = ?", bannerID)
	if idInt, err := strconv.ParseInt(bannerID, 10, 64); err == nil {
		query = h.localDB.Model(&domain.BannerOrder{}).Where("id = ? OR mysql_banner_id = ?", bannerID, idInt)
	}

	if err := query.Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil tugas desainer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tugas berhasil diambil", "editor_name": editorFullName})
}

type ValidateBannerRequest struct {
	Status       string `json:"status"`      // 'approved' or 'rejected'
	IsApproved   *bool  `json:"is_approved"` // true or false (alternative to status)
	Type         string `json:"type"`        // 'design', 'store', or 'both'
	RevisionNote string `json:"revision_note"`
	IsBihun      *bool  `json:"is_bihun"` // true or false if banner contains bihun
}

func (h *BrandingProxyHandler) ValidateBanner(c *gin.Context) {
	bannerID := c.Param("id")

	var req ValidateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var banner domain.BannerOrder
	query := h.localDB.Where("id = ?", bannerID)
	if idInt, err := strconv.ParseInt(bannerID, 10, 64); err == nil {
		query = h.localDB.Where("id = ? OR mysql_banner_id = ?", bannerID, idInt)
	}

	if err := query.First(&banner).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Banner tidak ditemukan"})
		return
	}

	// Role-based validation permission check
	callerRole, _ := c.Get("role")
	roleStr, _ := callerRole.(string)
	if roleStr == string(domain.RoleValidatorDesain) {
		if req.Type != "design" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Validator Desain hanya memiliki wewenang untuk validasi dan revisi desain visual"})
			return
		}
	} else if roleStr == string(domain.RoleValidatorToko) {
		if req.Type != "store" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Validator Toko hanya memiliki wewenang untuk validasi dan revisi persetujuan toko"})
			return
		}
	}

	now := time.Now()
	updates := map[string]interface{}{
		"updated_at":   now,
		"validated_at": now,
	}

	if req.IsBihun != nil {
		updates["is_bihun"] = *req.IsBihun
	}

	var empID *uuid.UUID
	approverInfo := GetApproverInfo(c, nil)
	if approverInfo != nil {
		if approverInfo.EmployeeID != nil {
			empID = approverInfo.EmployeeID
		} else if approverInfo.UserID != uuid.Nil {
			if emp, err := h.employeeUseCase.GetEmployeeByUserID(approverInfo.UserID); err == nil && emp != nil {
				empID = &emp.ID
			}
		}
	}

	isApproved := req.Status == "approved" || (req.IsApproved != nil && *req.IsApproved)
	isRejected := req.Status == "rejected" || (req.IsApproved != nil && !*req.IsApproved)

	if isApproved {
		if req.Type == "design" {
			updates["is_design_approved"] = true
			updates["design_approved_at"] = now
			if approverInfo != nil {
				if empID != nil {
					updates["design_validator_id"] = empID
				}
				if approverInfo.FullName != "" {
					updates["design_validator_name"] = approverInfo.FullName
				}
				if approverInfo.JobPosition != "" {
					updates["design_validator_position"] = approverInfo.JobPosition
				}
			}
			if banner.ValidatorName == nil || *banner.ValidatorName == "" {
				if empID != nil {
					updates["validator_id"] = empID
				}
				if approverInfo != nil && approverInfo.FullName != "" {
					updates["validator_name"] = approverInfo.FullName
				}
				if approverInfo != nil && approverInfo.JobPosition != "" {
					updates["validator_position"] = approverInfo.JobPosition
				}
			}

			if banner.Status == "val_store" || banner.IsStoreApproved {
				updates["status"] = "purchasing_queue"
			} else {
				updates["status"] = "val_design"
			}
		} else if req.Type == "store" {
			updates["is_store_approved"] = true
			updates["store_approved_at"] = now
			if approverInfo != nil {
				if empID != nil {
					updates["store_validator_id"] = empID
				}
				if approverInfo.FullName != "" {
					updates["store_validator_name"] = approverInfo.FullName
				}
				if approverInfo.JobPosition != "" {
					updates["store_validator_position"] = approverInfo.JobPosition
				}
			}
			if banner.ValidatorName == nil || *banner.ValidatorName == "" {
				if empID != nil {
					updates["validator_id"] = empID
				}
				if approverInfo != nil && approverInfo.FullName != "" {
					updates["validator_name"] = approverInfo.FullName
				}
				if approverInfo != nil && approverInfo.JobPosition != "" {
					updates["validator_position"] = approverInfo.JobPosition
				}
			}

			if banner.Status == "val_design" || banner.IsDesignApproved {
				updates["status"] = "purchasing_queue"
			} else {
				updates["status"] = "val_store"
			}
		} else {
			updates["status"] = "purchasing_queue"
			updates["is_design_approved"] = true
			updates["is_store_approved"] = true
			updates["design_approved_at"] = now
			updates["store_approved_at"] = now
			if approverInfo != nil {
				if empID != nil {
					updates["design_validator_id"] = empID
					updates["store_validator_id"] = empID
					updates["validator_id"] = empID
				}
				if approverInfo.FullName != "" {
					updates["design_validator_name"] = approverInfo.FullName
					updates["store_validator_name"] = approverInfo.FullName
					updates["validator_name"] = approverInfo.FullName
				}
				if approverInfo.JobPosition != "" {
					updates["design_validator_position"] = approverInfo.JobPosition
					updates["store_validator_position"] = approverInfo.JobPosition
					updates["validator_position"] = approverInfo.JobPosition
				}
			}
		}
	} else if isRejected {
		// Ketika revisi diberikan oleh validator, kembalikan status ke 'editing' agar editor dapat langsung merevisi
		updates["status"] = "editing"
		updates["editing_finished_at"] = nil
		updates["is_design_approved"] = false
		updates["is_store_approved"] = false

		if req.Type == "design" {
			updates["design_validator_id"] = nil
			updates["design_validator_name"] = nil
			updates["design_validator_position"] = nil
			updates["design_approved_at"] = nil
		} else if req.Type == "store" {
			updates["store_validator_id"] = nil
			updates["store_validator_name"] = nil
			updates["store_validator_position"] = nil
			updates["store_approved_at"] = nil
		} else {
			updates["design_validator_id"] = nil
			updates["design_validator_name"] = nil
			updates["design_validator_position"] = nil
			updates["design_approved_at"] = nil
			updates["store_validator_id"] = nil
			updates["store_validator_name"] = nil
			updates["store_validator_position"] = nil
			updates["store_approved_at"] = nil
		}

		newNote := req.RevisionNote
		if banner.RevisionNote != nil && *banner.RevisionNote != "" && req.RevisionNote != "" {
			newNote = *banner.RevisionNote + "\n\n" + req.RevisionNote
		}
		updates["revision_note"] = newNote
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status harus 'approved' atau 'rejected'"})
		return
	}

	if err := h.localDB.Model(&banner).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal validasi banner"})
		return
	}

	// Kirim notifikasi revisi ke editor yang sedang mengerjakan banner ini
	if isRejected && h.hub != nil && banner.EditorID != nil {
		revisionNote := req.RevisionNote
		if revisionNote == "" {
			revisionNote = "Mohon periksa kembali desain spanduk."
		}
		go func(b domain.BannerOrder, note string) {
			_ = h.hub.PublishNotification(context.Background(), sse.NotificationEvent{
				ID:        b.ID.String() + "-revision",
				Type:      "branding_revision",
				Title:     "✏️ Revisi Desain Diperlukan",
				Message:   fmt.Sprintf("Spanduk '%s' perlu direvisi: %s", b.StoreName, note),
				CompanyID: b.CompanyID.String(),
				CreatedAt: time.Now().Format(time.RFC3339),
				Data: map[string]interface{}{
					"banner_id":     b.ID.String(),
					"store_name":    b.StoreName,
					"revision_note": note,
				},
				// Hanya dikirim ke editor yang sedang mengerjakan banner ini
				TargetEmployeeIDs: []string{b.EditorID.String()},
			})
		}(banner, revisionNote)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Validasi berhasil disimpan", "status": updates["status"]})
}

type BulkVendorItem struct {
	BannerID   string  `json:"banner_id" binding:"required"`
	TotalPrice float64 `json:"total_price"`
}

type BulkVendorProcessRequest struct {
	Action           string           `json:"action" binding:"required"` // 'send_to_vendor', 'ready_for_operator', 'mark_paid', 'mark_unpaid'
	VendorID         *string          `json:"vendor_id"`
	VendorName       string           `json:"vendor_name"`
	PaymentReference *string          `json:"payment_reference"`
	PaymentNotes     *string          `json:"payment_notes"`
	BannerIDs        []string         `json:"banner_ids"`
	Items            []BulkVendorItem `json:"items"`
}

func (h *BrandingProxyHandler) BulkVendorProcess(c *gin.Context) {
	var req BulkVendorProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validActions := map[string]bool{
		"send_to_vendor":     true,
		"ready_for_operator": true,
		"finish_printing":    true,
		"mark_paid":          true,
		"mark_unpaid":        true,
		"update_pricing":     true,
		"update_vendor":      true,
		"mark_email_sent":    true,
		"unmark_email_sent":  true,
	}

	if !validActions[req.Action] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action tidak valid"})
		return
	}

	// Normalize items
	items := req.Items
	if len(items) == 0 && len(req.BannerIDs) > 0 {
		for _, bID := range req.BannerIDs {
			items = append(items, BulkVendorItem{
				BannerID: bID,
			})
		}
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Daftar banner tidak boleh kosong"})
		return
	}

	purchasingInfo := GetApproverInfo(c, nil)
	now := time.Now()

	tx := h.localDB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memulai transaksi database"})
		return
	}

	var vendorUUID *uuid.UUID
	if req.VendorID != nil && *req.VendorID != "" {
		if vUUID, err := uuid.Parse(*req.VendorID); err == nil {
			vendorUUID = &vUUID
		}
	}

	for _, item := range items {
		if item.BannerID == "" {
			continue
		}

		updates := map[string]interface{}{
			"updated_at": now,
		}

		if purchasingInfo != nil {
			if purchasingInfo.EmployeeID != nil {
				updates["purchasing_id"] = purchasingInfo.EmployeeID
			} else if purchasingInfo.UserID != uuid.Nil {
				if emp, err := h.employeeUseCase.GetEmployeeByUserID(purchasingInfo.UserID); err == nil && emp != nil {
					updates["purchasing_id"] = emp.ID
				}
			}
			if purchasingInfo.FullName != "" {
				updates["purchasing_name"] = purchasingInfo.FullName
			}
			if purchasingInfo.JobPosition != "" {
				updates["purchasing_position"] = purchasingInfo.JobPosition
			}
		}

		switch req.Action {
		case "send_to_vendor":
			updates["status"] = "sent_to_vendor"
			if item.TotalPrice > 0 {
				updates["total_price"] = item.TotalPrice
			}
			if vendorUUID != nil {
				updates["vendor_id"] = *vendorUUID
			}
			if req.VendorName != "" {
				updates["vendor_name"] = req.VendorName
			}
		case "update_pricing", "update_vendor":
			if item.TotalPrice > 0 {
				updates["total_price"] = item.TotalPrice
			}
			if vendorUUID != nil {
				updates["vendor_id"] = *vendorUUID
			}
			if req.VendorName != "" {
				updates["vendor_name"] = req.VendorName
			}
		case "ready_for_operator", "finish_printing":
			updates["status"] = "ready_for_operator"
		case "mark_paid":
			updates["is_paid"] = true
			updates["payment_status"] = "paid"
			updates["paid_at"] = &now
			if req.PaymentReference != nil && *req.PaymentReference != "" {
				updates["payment_reference"] = *req.PaymentReference
			}
			if req.PaymentNotes != nil && *req.PaymentNotes != "" {
				updates["payment_notes"] = *req.PaymentNotes
			}
		case "mark_unpaid":
			updates["is_paid"] = false
			updates["payment_status"] = "unpaid"
			updates["paid_at"] = nil
			updates["payment_reference"] = nil
			updates["payment_notes"] = nil
		case "mark_email_sent":
			updates["is_vendor_email_sent"] = true
			updates["vendor_email_sent_at"] = &now
		case "unmark_email_sent":
			updates["is_vendor_email_sent"] = false
			updates["vendor_email_sent_at"] = nil
		}

		query := tx.Model(&domain.BannerOrder{}).Where("id = ?", item.BannerID)
		if idInt, err := strconv.ParseInt(item.BannerID, 10, 64); err == nil {
			query = tx.Model(&domain.BannerOrder{}).Where("id = ? OR mysql_banner_id = ?", item.BannerID, idInt)
		}

		if err := query.Updates(updates).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal memproses banner %s: %v", item.BannerID, err)})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan perubahan status bulk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Berhasil memproses %d banner", len(items)),
		"total":   len(items),
	})
}

type VendorProcessRequest struct {
	Action           string  `json:"action" binding:"required"` // 'send_to_vendor', 'ready_for_operator', 'mark_paid', 'mark_unpaid'
	VendorID         *string `json:"vendor_id"`
	VendorName       string  `json:"vendor_name"`
	TotalPrice       float64 `json:"total_price"`
	PaymentReference *string `json:"payment_reference"`
	PaymentNotes     *string `json:"payment_notes"`
}

func (h *BrandingProxyHandler) VendorProcess(c *gin.Context) {
	bannerID := c.Param("id")

	var req VendorProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"updated_at": now,
	}

	purchasingInfo := GetApproverInfo(c, nil)
	if purchasingInfo != nil {
		if purchasingInfo.EmployeeID != nil {
			updates["purchasing_id"] = purchasingInfo.EmployeeID
		} else if purchasingInfo.UserID != uuid.Nil {
			if emp, err := h.employeeUseCase.GetEmployeeByUserID(purchasingInfo.UserID); err == nil && emp != nil {
				updates["purchasing_id"] = emp.ID
			}
		}
		if purchasingInfo.FullName != "" {
			updates["purchasing_name"] = purchasingInfo.FullName
		}
		if purchasingInfo.JobPosition != "" {
			updates["purchasing_position"] = purchasingInfo.JobPosition
		}
	}

	switch req.Action {
	case "send_to_vendor":
		updates["status"] = "sent_to_vendor"
		if req.TotalPrice > 0 {
			updates["total_price"] = req.TotalPrice
		}
		if req.VendorID != nil && *req.VendorID != "" {
			if vUUID, err := uuid.Parse(*req.VendorID); err == nil {
				updates["vendor_id"] = vUUID
			}
		}
		if req.VendorName != "" {
			updates["vendor_name"] = req.VendorName
		}
	case "update_pricing", "update_vendor":
		if req.TotalPrice > 0 {
			updates["total_price"] = req.TotalPrice
		}
		if req.VendorID != nil && *req.VendorID != "" {
			if vUUID, err := uuid.Parse(*req.VendorID); err == nil {
				updates["vendor_id"] = vUUID
			}
		}
		if req.VendorName != "" {
			updates["vendor_name"] = req.VendorName
		}
	case "ready_for_operator", "finish_printing":
		updates["status"] = "ready_for_operator"
	case "mark_paid":
		updates["is_paid"] = true
		updates["payment_status"] = "paid"
		updates["paid_at"] = &now
		if req.PaymentReference != nil && *req.PaymentReference != "" {
			updates["payment_reference"] = *req.PaymentReference
		}
		if req.PaymentNotes != nil && *req.PaymentNotes != "" {
			updates["payment_notes"] = *req.PaymentNotes
		}
	case "mark_unpaid":
		updates["is_paid"] = false
		updates["payment_status"] = "unpaid"
		updates["paid_at"] = nil
		updates["payment_reference"] = nil
		updates["payment_notes"] = nil
	case "mark_email_sent":
		updates["is_vendor_email_sent"] = true
		updates["vendor_email_sent_at"] = &now
	case "unmark_email_sent":
		updates["is_vendor_email_sent"] = false
		updates["vendor_email_sent_at"] = nil
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action tidak valid"})
		return
	}

	query := h.localDB.Model(&domain.BannerOrder{}).Where("id = ?", bannerID)
	if idInt, err := strconv.ParseInt(bannerID, 10, 64); err == nil {
		query = h.localDB.Model(&domain.BannerOrder{}).Where("id = ? OR mysql_banner_id = ?", bannerID, idInt)
	}

	if err := query.Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses status vendor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status vendor berhasil diperbarui"})
}

func (h *BrandingProxyHandler) InstallBanner(c *gin.Context) {
	bannerID := c.Param("id")

	lat := c.PostForm("latitude")
	lng := c.PostForm("longitude")

	if lat == "" || lng == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat GPS diperlukan untuk verifikasi pemasangan"})
		return
	}

	_, header, err := c.Request.FormFile("photo_installed")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca foto instalasi"})
		return
	}

	userIDStr, _ := c.Get("userID")
	employee, _ := h.employeeUseCase.GetEmployeeByUserID(userIDStr.(uuid.UUID))

	// Simpan ke VPS langsung dengan kompresi WebP (lebar max 1600px, quality 80) agar hemat ruang & sangat jernih
	photoLink, err := utils.SaveUploadedImageAsWebP(header, "./uploads/banners/installed", 1600, 80)
	if err != nil {
		log.Printf("[BrandingHandler] Gagal menyimpan foto bukti pasang di VPS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan foto instalasi di server: " + err.Error()})
		return
	}

	installedAt := time.Now()
	if instAtStr := strings.TrimSpace(c.PostForm("installed_at")); instAtStr != "" {
		formats := []string{
			"2006-01-02T15:04",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"2006-01-02",
			time.RFC3339,
		}
		for _, layout := range formats {
			if parsed, err := time.ParseInLocation(layout, instAtStr, time.Local); err == nil {
				installedAt = parsed
				break
			}
		}
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":          "installed",
		"photo_installed": photoLink,
		"installed_at":    installedAt,
		"updated_at":      now,
	}

	if employee != nil {
		updates["operator_id"] = employee.ID
		updates["installer_id"] = employee.ID
	}

	query := h.localDB.Model(&domain.BannerOrder{}).Where("id = ?", bannerID)
	if idInt, err := strconv.ParseInt(bannerID, 10, 64); err == nil {
		query = h.localDB.Model(&domain.BannerOrder{}).Where("id = ? OR mysql_banner_id = ?", bannerID, idInt)
	}

	if err := query.Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyelesaikan instalasi banner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Banner berhasil diinstal", "photo_installed": photoLink})
}

func (h *BrandingProxyHandler) OperatorTakeBanner(c *gin.Context) {
	bannerID := c.Param("id")

	userIDStr, _ := c.Get("userID")
	employee, _ := h.employeeUseCase.GetEmployeeByUserID(userIDStr.(uuid.UUID))

	if employee == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Identitas karyawan tidak valid"})
		return
	}

	var banner domain.BannerOrder
	query := h.localDB.Where("id = ?", bannerID)
	if idInt, err := strconv.ParseInt(bannerID, 10, 64); err == nil {
		query = h.localDB.Where("id = ? OR mysql_banner_id = ?", bannerID, idInt)
	}

	if err := query.First(&banner).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data spanduk tidak ditemukan"})
		return
	}

	// Proteksi anti-rebut: jika spanduk sudah berstatus diterima_operator atau terpasang oleh operator lain
	if banner.OperatorID != nil && *banner.OperatorID != employee.ID && (banner.Status == "diterima_operator" || banner.Status == "installed" || banner.Status == "terpasang") {
		var opName string = "operator lain"
		if banner.OperatorName != nil && *banner.OperatorName != "" {
			opName = *banner.OperatorName
		} else {
			var existingOp domain.Employee
			if err := h.localDB.Where("id = ?", *banner.OperatorID).First(&existingOp).Error; err == nil {
				opName = strings.TrimSpace(existingOp.FirstName + " " + existingOp.LastName)
			}
		}
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("Spanduk ini sudah diambil oleh %s. Penugasan hanya dapat diubah oleh Admin melalui Web Manajemen.", opName),
		})
		return
	}

	operatorName := strings.TrimSpace(employee.FirstName + " " + employee.LastName)
	now := time.Now()
	updates := map[string]interface{}{
		"status":               "diterima_operator",
		"operator_id":          employee.ID,
		"operator_name":        operatorName,
		"operator_received_at": now,
		"updated_at":           now,
	}

	if err := h.localDB.Model(&banner).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil spanduk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Spanduk berhasil diambil, siap untuk dipasang"})
}

// ─────────────────────────────────────────────────────────────────────────────
// VENDOR CRUD ENDPOINTS
// ─────────────────────────────────────────────────────────────────────────────

type CreateVendorRequest struct {
	Name          string  `json:"name" binding:"required"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	Address       string  `json:"address"`
	PricePerMeter float64 `json:"price_per_meter"`
	IsActive      *bool   `json:"is_active"`
}

type UpdateVendorRequest struct {
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	Address       string  `json:"address"`
	PricePerMeter float64 `json:"price_per_meter"`
	IsActive      *bool   `json:"is_active"`
}

func (h *BrandingProxyHandler) GetVendors(c *gin.Context) {
	search := c.Query("search")
	isActiveParam := c.Query("is_active")

	query := h.localDB.Model(&domain.BrandingVendor{})

	if search != "" {
		s := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(phone) LIKE ? OR LOWER(address) LIKE ?", s, s, s)
	}

	if isActiveParam != "" {
		if isActiveParam == "true" || isActiveParam == "1" {
			query = query.Where("is_active = true")
		} else if isActiveParam == "false" || isActiveParam == "0" {
			query = query.Where("is_active = false")
		}
	}

	var vendors []domain.BrandingVendor
	if err := query.Order("name ASC").Find(&vendors).Error; err != nil {
		log.Printf("[BrandingHandler] GetVendors error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data vendor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    vendors,
		"total":   len(vendors),
	})
}

func (h *BrandingProxyHandler) CreateVendor(c *gin.Context) {
	var req CreateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama vendor wajib diisi: " + err.Error()})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	vendor := domain.BrandingVendor{
		ID:            uuid.New(),
		Name:          strings.TrimSpace(req.Name),
		Email:         strings.TrimSpace(req.Email),
		Phone:         strings.TrimSpace(req.Phone),
		Address:       strings.TrimSpace(req.Address),
		PricePerMeter: req.PricePerMeter,
		IsActive:      isActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.localDB.Create(&vendor).Error; err != nil {
		log.Printf("[BrandingHandler] CreateVendor error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan vendor baru"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vendor berhasil ditambahkan",
		"data":    vendor,
	})
}

func (h *BrandingProxyHandler) UpdateVendor(c *gin.Context) {
	vendorID := c.Param("id")

	var vendor domain.BrandingVendor
	if err := h.localDB.Where("id = ?", vendorID).First(&vendor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vendor tidak ditemukan"})
		return
	}

	var req UpdateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	updates["email"] = strings.TrimSpace(req.Email)
	updates["phone"] = strings.TrimSpace(req.Phone)
	updates["address"] = strings.TrimSpace(req.Address)
	if req.PricePerMeter >= 0 {
		updates["price_per_meter"] = req.PricePerMeter
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.localDB.Model(&vendor).Updates(updates).Error; err != nil {
		log.Printf("[BrandingHandler] UpdateVendor error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data vendor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data vendor berhasil diperbarui",
		"data":    vendor,
	})
}

func (h *BrandingProxyHandler) DeleteVendor(c *gin.Context) {
	vendorID := c.Param("id")

	var vendor domain.BrandingVendor
	if err := h.localDB.Where("id = ?", vendorID).First(&vendor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vendor tidak ditemukan"})
		return
	}

	if err := h.localDB.Delete(&vendor).Error; err != nil {
		log.Printf("[BrandingHandler] DeleteVendor error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus vendor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vendor berhasil dihapus"})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /branding/migrate-vps-designs
// Migrates legacy VPS local banner designs to Google Drive and cleans VPS disk
// ─────────────────────────────────────────────────────────────────────────────

func (h *BrandingProxyHandler) MigrateVPSDesignsToGDrive(c *gin.Context) {
	if h.gdrive == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Layanan Google Drive tidak tersedia"})
		return
	}

	type BannerRecord struct {
		ID             string  `gorm:"column:id"`
		StoreName      string  `gorm:"column:store_name"`
		DesignURL      string  `gorm:"column:design_url"`
		DesignLink     *string `gorm:"column:design_link"`
		DesignFileName *string `gorm:"column:design_file_name"`
	}

	var records []BannerRecord
	query := h.localDB.Table("banner_orders").
		Select("id, store_name, design_url, design_link, design_file_name").
		Where("(design_url IS NOT NULL AND design_url != '' AND design_url NOT LIKE '%drive.google.com%') OR (design_link IS NOT NULL AND design_link != '' AND design_link NOT LIKE '%drive.google.com%')")

	if err := query.Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal query data banner: " + err.Error()})
		return
	}

	if len(records) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":          "Semua file desain sudah tersimpan di Google Drive",
			"total_processed":  0,
			"migrated_success": 0,
		})
		return
	}

	go func(items []BannerRecord) {
		successCount := 0
		var totalBytesFreed int64 = 0

		for _, record := range items {
			rawPath := record.DesignURL
			if rawPath == "" && record.DesignLink != nil {
				rawPath = *record.DesignLink
			}

			cleanPath := strings.TrimPrefix(rawPath, "/")
			cleanPath = strings.TrimPrefix(cleanPath, "api/v1/")
			cleanPath = strings.TrimPrefix(cleanPath, "api/")

			possiblePaths := []string{
				cleanPath,
				"./" + cleanPath,
				"uploads/banners/designs/" + strings.TrimPrefix(cleanPath, "uploads/banners/designs/"),
				"./uploads/banners/designs/" + strings.TrimPrefix(cleanPath, "uploads/banners/designs/"),
			}

			var targetFilePath string
			for _, p := range possiblePaths {
				if fileInfo, err := os.Stat(p); err == nil && !fileInfo.IsDir() {
					targetFilePath = p
					totalBytesFreed += fileInfo.Size()
					break
				}
			}

			if targetFilePath == "" {
				continue
			}

			file, err := os.Open(targetFilePath)
			if err != nil {
				continue
			}

			ext := ".webp"
			if strings.Contains(targetFilePath, ".") {
				ext = targetFilePath[strings.LastIndex(targetFilePath, "."):]
			}
			safeStoreName := strings.ReplaceAll(record.StoreName, "/", "-")
			gdriveFileName := fmt.Sprintf("%s_%s_desain%s", record.ID, safeStoreName, ext)

			gdriveURL, err := h.gdrive.UploadFileWithName(file, gdriveFileName)
			file.Close()

			if err != nil {
				log.Printf("[BrandingMigrator] Gagal upload ke GDrive: %v", err)
				continue
			}

			updates := map[string]interface{}{
				"design_url":  gdriveURL,
				"design_link": gdriveURL,
			}
			if record.DesignFileName == nil || *record.DesignFileName == "" {
				updates["design_file_name"] = fmt.Sprintf("%s_desain%s", safeStoreName, ext)
			}

			if err := h.localDB.Table("banner_orders").Where("id = ?", record.ID).Updates(updates).Error; err == nil {
				_ = os.Remove(targetFilePath) // Hapus file lokal untuk membebaskan disk
				successCount++
			}
		}
		log.Printf("[BrandingMigrator] Migrasi selesai! %d file berhasil dipindahkan ke GDrive (%.2f MB dibebaskan)", successCount, float64(totalBytesFreed)/(1024*1024))
	}(records)

	c.JSON(http.StatusOK, gin.H{
		"message":          "Proses migrasi file desain ke Google Drive sedang berjalan di latar belakang",
		"total_to_migrate": len(records),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /branding/banners/send-vendor-email & /branding/banners/:id/send-vendor-email
// Kirim SPK Cetak langsung ke Vendor Percetakan via Server SMTP
// ─────────────────────────────────────────────────────────────────────────────

type SendVendorEmailRequest struct {
	BannerIDs   []string `json:"banner_ids"`
	BannerID    string   `json:"banner_id"`
	VendorID    *string  `json:"vendor_id"`
	VendorName  string   `json:"vendor_name"`
	VendorEmail string   `json:"vendor_email" binding:"required"`
	Subject     string   `json:"subject"`
	CustomNotes string   `json:"custom_notes"`
	HtmlBody    string   `json:"html_body"`
}

func (h *BrandingProxyHandler) SendVendorEmail(c *gin.Context) {
	var req SendVendorEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email vendor tujuan wajib diisi: " + err.Error()})
		return
	}

	// Ambil konfigurasi SMTP dari ENV
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	if smtpPort == "" {
		smtpPort = "587"
	}
	if smtpFrom == "" {
		smtpFrom = smtpUser
	}

	if smtpUser == "" || smtpPass == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Konfigurasi SMTP Server (SMTP_USER / SMTP_PASS) belum disetel di server",
		})
		return
	}

	// Tentukan Subject & Body Email
	subject := strings.TrimSpace(req.Subject)
	if subject == "" {
		subject = "[ORDER CETAK SPANDUK] Karsa HR"
	}

	htmlBody := req.HtmlBody
	if strings.TrimSpace(htmlBody) == "" {
		// Default HTML Body jika tidak dikirim dari frontend
		notesHtml := ""
		if strings.TrimSpace(req.CustomNotes) != "" {
			notesHtml = fmt.Sprintf("<p style='margin-top: 12px; color: #475569;'><strong>Catatan Tambahan:</strong><br>%s</p>", strings.ReplaceAll(req.CustomNotes, "\n", "<br>"))
		}

		htmlBody = fmt.Sprintf(`
			<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e2e8f0; border-radius: 12px; background-color: #ffffff;">
				<h2 style="color: #1e293b; border-bottom: 2px solid #3b82f6; padding-bottom: 8px;">Order Cetak Spanduk (SPK)</h2>
				<p>Halo Tim Percetakan <strong>%s</strong>,</p>
				<p>Berikut adalah Surat Perintah Kerja (SPK) cetak spanduk / banner dari Karsa HR.</p>
				%s
				<hr style="border: 0; border-top: 1px solid #e2e8f0; margin: 20px 0;">
				<p style="font-size: 13px; color: #64748b;">Email ini dikirim otomatis melalui Sistem HRIS & Operasional Karsa HR.</p>
			</div>
		`, req.VendorName, notesHtml)
	}

	// Kirim Email menggunakan utils.NewEmailSender
	sender := utils.NewEmailSender(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)
	err := sender.SendEmail(strings.TrimSpace(req.VendorEmail), subject, htmlBody)
	if err != nil {
		log.Printf("[BrandingHandler] Gagal kirim email SPK ke %s: %v", req.VendorEmail, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Gagal mengirim email via SMTP Server: %v", err),
		})
		return
	}

	log.Printf("[BrandingHandler] ✅ Sukses mengirim email SPK via SMTP ke: %s (Vendor: %s)", req.VendorEmail, req.VendorName)

	// Update tracking status pengiriman email ke database
	now := time.Now()
	var bannerIDs []string
	if len(req.BannerIDs) > 0 {
		bannerIDs = append(bannerIDs, req.BannerIDs...)
	}
	if req.BannerID != "" {
		bannerIDs = append(bannerIDs, req.BannerID)
	}
	paramID := c.Param("id")
	if paramID != "" {
		bannerIDs = append(bannerIDs, paramID)
	}

	if len(bannerIDs) > 0 {
		var intIDs []int64
		var validUUIDs []uuid.UUID
		for _, bID := range bannerIDs {
			if idInt, err := strconv.ParseInt(bID, 10, 64); err == nil {
				intIDs = append(intIDs, idInt)
			}
			if u, err := uuid.Parse(bID); err == nil {
				validUUIDs = append(validUUIDs, u)
			}
		}

		var dbQuery *gorm.DB
		if len(validUUIDs) > 0 && len(intIDs) > 0 {
			dbQuery = h.localDB.Model(&domain.BannerOrder{}).Where("id IN ? OR mysql_banner_id IN ?", validUUIDs, intIDs)
		} else if len(validUUIDs) > 0 {
			dbQuery = h.localDB.Model(&domain.BannerOrder{}).Where("id IN ?", validUUIDs)
		} else if len(intIDs) > 0 {
			dbQuery = h.localDB.Model(&domain.BannerOrder{}).Where("mysql_banner_id IN ?", intIDs)
		}

		if dbQuery != nil {
			if err := dbQuery.Updates(map[string]interface{}{
				"is_vendor_email_sent": true,
				"vendor_email_sent_at": &now,
				"updated_at":           now,
			}).Error; err != nil {
				log.Printf("[BrandingHandler] Warning: Gagal update status email vendor untuk banner %v: %v", bannerIDs, err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Email SPK berhasil dikirim ke %s via Server SMTP", req.VendorEmail),
		"recipient": req.VendorEmail,
		"subject":   subject,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// SALES BANNER INSTALLATION & REIMBURSEMENT
// ─────────────────────────────────────────────────────────────────────────────

// GET /branding/sales/unclaimed-installations
func (h *BrandingProxyHandler) GetUnclaimedInstallations(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userRole, _ := c.Get("userRole")
	roleStr, _ := userRole.(string)
	isAdmin := roleStr == "superadmin" || roleStr == "hr_admin" || roleStr == "manager" || roleStr == "admin" || roleStr == "cashier" || roleStr == "admin_kasir" || roleStr == "admin_nota"

	empParam := c.Query("employee_id")
	branchParam := c.Query("branch_id")
	dateFilter := c.Query("date") // Format: YYYY-MM-DD (optional)

	query := h.localDB.Model(&domain.BannerOrder{}).
		Preload("Employee").
		Preload("Installer").
		Preload("Operator").
		Where("status IN ('installed', 'terpasang')").
		Where("reimbursement_id IS NULL").
		Where("is_paid = false")

	if empParam != "" {
		if parsedEmpID, err := uuid.Parse(empParam); err == nil {
			query = query.Where("(installer_id = ? OR operator_id = ? OR employee_id = ?)", parsedEmpID, parsedEmpID, parsedEmpID)
		}
	} else if !isAdmin {
		employee, err := h.employeeUseCase.GetEmployeeByUserID(userIDStr.(uuid.UUID))
		if err != nil || employee == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data profil sales tidak ditemukan"})
			return
		}
		query = query.Where("(installer_id = ? OR operator_id = ? OR employee_id = ?)", employee.ID, employee.ID, employee.ID)
	}

	if branchParam != "" && branchParam != "ALL_BRANCHES" && branchParam != "ALL" {
		if parsedBranchID, err := uuid.Parse(branchParam); err == nil {
			query = query.Joins("LEFT JOIN employees ON employees.id = banner_orders.employee_id").
				Where("employees.branch_id = ?", parsedBranchID)
		}
	}

	if dateFilter != "" {
		query = query.Where("DATE(installed_at) = ? OR (installed_at IS NULL AND DATE(updated_at) = ?)", dateFilter, dateFilter)
	}

	var banners []domain.BannerOrder
	if err := query.Order("installed_at DESC, created_at DESC").Find(&banners).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat data spanduk terpasang: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  banners,
		"total": len(banners),
	})
}

// POST /branding/sales/direct-install
func (h *BrandingProxyHandler) DirectInstallBanner(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userIDStr.(uuid.UUID))
	if err != nil || employee == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data profil sales tidak ditemukan"})
		return
	}

	shopName := strings.TrimSpace(c.PostForm("shop_name"))
	if shopName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama toko/lokasi pemasangan wajib diisi"})
		return
	}

	lat := c.PostForm("latitude")
	lng := c.PostForm("longitude")
	if lat == "" || lng == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat GPS lokasi pemasangan diperlukan"})
		return
	}

	_, header, err := c.Request.FormFile("photo_installed")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Foto bukti pemasangan banner wajib diunggah"})
		return
	}

	// Simpan foto bukti pasang dengan kompresi WebP
	photoLink, err := utils.SaveUploadedImageAsWebP(header, "./uploads/banners/installed", 1600, 80)
	if err != nil {
		log.Printf("[BrandingHandler] Gagal menyimpan foto pasang: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan foto bukti pasang: " + err.Error()})
		return
	}

	bannerType := c.DefaultPostForm("banner_type", "Standard")
	notes := c.PostForm("notes")
	qty, _ := strconv.Atoi(c.DefaultPostForm("qty", "1"))
	if qty <= 0 {
		qty = 1
	}
	sizeLength, _ := strconv.ParseFloat(c.DefaultPostForm("size_length", "3"), 64)
	sizeWidth, _ := strconv.ParseFloat(c.DefaultPostForm("size_width", "1"), 64)

	// Normalisasi ke meter jika diinput dalam centimeter (misal 300 x 100 cm)
	lenM := sizeLength
	if lenM >= 10 {
		lenM = lenM / 100.0
	}
	widM := sizeWidth
	if widM >= 10 {
		widM = widM / 100.0
	}
	size := lenM * widM
	if size <= 0 {
		size = 3.0
	}

	// Cari toko jika ada
	var store domain.Store
	var storeID *uuid.UUID
	if err := h.localDB.Where("LOWER(TRIM(name)) = LOWER(TRIM(?))", shopName).First(&store).Error; err == nil {
		storeID = &store.ID
	}

	companyID := uuid.Nil
	if employee.CompanyID != nil {
		companyID = *employee.CompanyID
	}

	salesName := employee.FirstName
	if employee.LastName != "" {
		salesName += " " + employee.LastName
	}

	now := time.Now()
	newBanner := domain.BannerOrder{
		ID:                    uuid.New(),
		CompanyID:             companyID,
		EmployeeID:            employee.ID,
		InstallerID:           &employee.ID,
		OperatorID:            &employee.ID,
		StoreID:               storeID,
		StoreName:             shopName,
		Location:              fmt.Sprintf("GPS: %s, %s", lat, lng),
		CampaignName:          "Pemasangan Spanduk: " + shopName,
		BannerType:            bannerType,
		Size:                  size,
		SizeLength:            sizeLength,
		SizeWidth:             sizeWidth,
		Qty:                   qty,
		Notes:                 &notes,
		PhotoInstalled:        &photoLink,
		DocumentationImageURL: photoLink,
		Status:                "installed",
		InstalledAt:           &now,
		MySQLSalesName:        &salesName,
		OperatorName:          &salesName,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := h.localDB.Create(&newBanner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan laporan pemasangan banner: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pemasangan banner berhasil dilaporkan",
		"data":    newBanner,
	})
}

// POST /branding/sales/bulk-reimburse
func (h *BrandingProxyHandler) BulkReimburseInstallations(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	employee, err := h.employeeUseCase.GetEmployeeByUserID(userIDStr.(uuid.UUID))
	if err != nil || employee == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data profil sales tidak ditemukan"})
		return
	}

	var req struct {
		BannerIDs []string `json:"banner_ids"`
		Notes     string   `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	query := h.localDB.Model(&domain.BannerOrder{}).
		Where("(installer_id = ? OR operator_id = ? OR employee_id = ?)", employee.ID, employee.ID, employee.ID).
		Where("status IN ('installed', 'terpasang')").
		Where("reimbursement_id IS NULL").
		Where("is_paid = false")

	if len(req.BannerIDs) > 0 {
		query = query.Where("id IN (?)", req.BannerIDs)
	} else {
		// Default to today's installed banners if none specified
		today := time.Now().Format("2006-01-02")
		query = query.Where("DATE(installed_at) = ? OR (installed_at IS NULL AND DATE(updated_at) = ?)", today, today)
	}

	var banners []domain.BannerOrder
	if err := query.Find(&banners).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat data spanduk yang akan diklaim: " + err.Error()})
		return
	}

	if len(banners) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada spanduk terpasang baru yang dapat diajukan reimbursement"})
		return
	}

	// Buat ringkasan deskripsi toko dan hitung total luas m2 (dinormalisasi dari cm/meter)
	var shopNames []string
	firstPhoto := ""
	var bannerIDs []uuid.UUID
	totalArea := 0.0
	for _, b := range banners {
		bannerIDs = append(bannerIDs, b.ID)
		if b.StoreName != "" {
			shopNames = append(shopNames, b.StoreName)
		}
		if firstPhoto == "" && b.PhotoInstalled != nil && *b.PhotoInstalled != "" {
			firstPhoto = *b.PhotoInstalled
		}

		// Normalisasi cm vs meter
		lenM := b.SizeLength
		if lenM >= 10 {
			lenM = lenM / 100.0
		}
		widM := b.SizeWidth
		if widM >= 10 {
			widM = widM / 100.0
		}
		area := lenM * widM
		if area <= 0 {
			area = b.Size
			if area > 100 { // jika size tersimpan dalam cm²
				area = area / 10000.0
			}
		}
		if area <= 0 {
			area = 3.0 // default banner 3x1 m
		}
		totalArea += area
	}

	ratePerM2 := 1500.0
	estimatedAmount := totalArea * ratePerM2

	desc := req.Notes
	if desc == "" {
		desc = fmt.Sprintf("Klaim pemasangan %d spanduk (Total %.1f m² @ Rp %.0f/m²) di toko: %s", len(banners), totalArea, ratePerM2, strings.Join(shopNames, ", "))
	}

	now := time.Now()
	dateStr := now.Format("02/01/2006")
	reimbursement := domain.Reimbursement{
		ID:                 uuid.New(),
		EmployeeID:         employee.ID,
		Title:              fmt.Sprintf("Reimbursement Pasang Spanduk (%d Banner - %.1f m²) - %s", len(banners), totalArea, dateStr),
		Description:        desc,
		Amount:             estimatedAmount,
		UnitRate:           ratePerM2,
		BannerCount:        len(banners),
		Category:           "BANNER_INSTALLATION",
		AttachmentURL:      firstPhoto,
		Status:             domain.ReimbursementPending,
		DisbursementStatus: "UNPAID",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Simpan dalam transaksi DB
	tx := h.localDB.Begin()
	if err := tx.Create(&reimbursement).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat pengajuan reimbursement: " + err.Error()})
		return
	}

	// Link banner_orders ke reimbursement_id ini
	if err := tx.Model(&domain.BannerOrder{}).
		Where("id IN (?)", bannerIDs).
		Updates(map[string]interface{}{
			"reimbursement_id": reimbursement.ID,
			"updated_at":       now,
		}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghubungkan data spanduk ke reimbursement: " + err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyelesaikan pengajuan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Berhasil mengajukan reimbursement untuk %d pemasangan spanduk", len(banners)),
		"reimbursement": reimbursement,
		"banner_count":  len(banners),
	})
}
