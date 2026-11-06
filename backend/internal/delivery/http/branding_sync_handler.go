package http

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"example.com/hris/backend/internal/domain"
)

type MySQLBannerSyncRow struct {
	ID                 int64      `gorm:"column:id"`
	SalesID            *int64     `gorm:"column:sales_id"`
	ShopID             int64      `gorm:"column:shop_id"`
	EditorID           *int64     `gorm:"column:editor_id"`
	Type               string     `gorm:"column:type"`
	IsBihun            bool       `gorm:"column:is_bihun"`
	Qty                int        `gorm:"column:qty"`
	SizeLength         float64    `gorm:"column:size_length"`
	SizeWidth          float64    `gorm:"column:size_width"`
	Notes              *string    `gorm:"column:notes"`
	RevisionNote       *string    `gorm:"column:revision_note"`
	IsDesignApproved   bool       `gorm:"column:is_design_approved"`
	IsStoreApproved    bool       `gorm:"column:is_store_approved"`
	Status             string     `gorm:"column:status"`
	DesignLink         *string    `gorm:"column:design_link"`
	TotalPrice         float64    `gorm:"column:total_price"`
	OperatorReward     float64    `gorm:"column:operator_reward"`
	VendorID           *int64     `gorm:"column:vendor_id"`
	OperatorID         *int64     `gorm:"column:operator_id"`
	SPKDate            *string    `gorm:"column:spk_date"`
	PhotoInstalled     *string    `gorm:"column:photo_installed"`
	InstalledAt        *time.Time `gorm:"column:installed_at"`
	CreatedAt          *time.Time `gorm:"column:created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
	EditingAt          *time.Time `gorm:"column:editing_at"`
	EditingFinishedAt  *time.Time `gorm:"column:editing_finished_at"`
	ValidatedAt        *time.Time `gorm:"column:validated_at"`
	OperatorReceivedAt *time.Time `gorm:"column:operator_received_at"`
	IsMarkedCustomer   bool       `gorm:"column:is_marked_customer"`
	// Joined fields
	ShopName     string  `gorm:"column:shop_name"`
	ShopAddress  string  `gorm:"column:shop_address"`
	SalesName    *string `gorm:"column:sales_name"`
	EditorName   *string `gorm:"column:editor_name"`
	VendorName   *string `gorm:"column:vendor_name"`
	OperatorName *string `gorm:"column:operator_name"`
}

// connectRemoteMySQL attempts to connect to MySQL Branding with safe timeout
func connectRemoteMySQL() (*gorm.DB, error) {
	dbHost := os.Getenv("BRANDING_DB_HOST")
	dbPort := os.Getenv("BRANDING_DB_PORT")
	dbUser := os.Getenv("BRANDING_DB_USER")
	dbPass := os.Getenv("BRANDING_DB_PASSWORD")
	dbName := os.Getenv("BRANDING_DB_NAME")

	if dbHost == "" || dbUser == "" || dbPass == "" || dbName == "" {
		return nil, fmt.Errorf("integrasi branding belum dikonfigurasi untuk trial")
	}
	if dbPort == "" {
		dbPort = "3306"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	type connResult struct {
		db  *gorm.DB
		err error
	}
	ch := make(chan connResult, 1)

	go func() {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		ch <- connResult{db: db, err: err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return nil, res.err
		}
		return res.db, nil
	case <-time.After(12 * time.Second):
		return nil, fmt.Errorf("koneksi ke MySQL Branding timeout (12s). Pastikan remote MySQL diizinkan.")
	}
}

// getStatusRank returns numeric weight representing workflow progression
func getStatusRank(status string) int {
	st := strings.ToLower(strings.TrimSpace(status))
	switch st {
	case "installed", "terpasang":
		return 5
	case "ready_for_operator", "diterima_operator", "siap_pasang", "siap pasang":
		return 4
	case "sent_to_vendor", "purchasing_queue", "percetakan":
		return 3
	case "validation", "val_design", "val_store", "admin_review":
		return 2
	case "editing", "in_design":
		return 1
	default: // "request", "pending", ""
		return 0
	}
}

// SyncFromBrandingApp handles POST /branding/sync with Workflow Forward-Only Protection
func (h *BrandingProxyHandler) SyncFromBrandingApp(c *gin.Context) {
	mysqlDB, err := connectRemoteMySQL()
	if err != nil {
		log.Printf("[BrandingSync] Gagal menghubungkan ke MySQL Branding: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": fmt.Sprintf("Gagal terhubung ke MySQL Branding: %v", err),
		})
		return
	}

	// 1. Resolve Company
	var company domain.Company
	if err := h.localDB.Where("LOWER(name) LIKE ?", "%").First(&company).Error; err != nil {
		if err := h.localDB.First(&company).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Perusahaan tidak ditemukan di PostgreSQL Super App",
			})
			return
		}
	}

	// 2. Fetch all banners from MySQL
	var myBanners []MySQLBannerSyncRow
	query := `
		SELECT b.*,
		       s.name as shop_name,
		       s.address_detail as shop_address,
		       sales.name as sales_name,
		       editors.name as editor_name,
		       vendors.name as vendor_name,
		       operators.name as operator_name
		FROM banners b
		LEFT JOIN shops s ON s.id = b.shop_id
		LEFT JOIN users sales ON sales.id = b.sales_id
		LEFT JOIN users editors ON editors.id = b.editor_id
		LEFT JOIN vendors ON vendors.id = b.vendor_id
		LEFT JOIN users operators ON operators.id = b.operator_id
	`
	if err := mysqlDB.Raw(query).Scan(&myBanners).Error; err != nil {
		log.Printf("[BrandingSync] Gagal membaca data banners dari MySQL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Gagal query banner MySQL: %v", err),
		})
		return
	}

	log.Printf("[BrandingSync] Menemukan %d banner di MySQL Branding...", len(myBanners))

	// Preload stores mapping by branding_shop_id
	type StoreMapRow struct {
		ID             uuid.UUID
		BrandingShopID *int
	}
	var stores []StoreMapRow
	h.localDB.Table("stores").Select("id, branding_shop_id").Where("branding_shop_id IS NOT NULL").Scan(&stores)
	storeMap := make(map[int64]uuid.UUID)
	for _, s := range stores {
		if s.BrandingShopID != nil {
			storeMap[int64(*s.BrandingShopID)] = s.ID
		}
	}

	// Preload default employee
	var defaultEmp domain.Employee
	h.localDB.Where("company_id = ?", company.ID).First(&defaultEmp)

	updatedCount := 0
	createdCount := 0
	completedCount := 0

	for _, mb := range myBanners {
		size := mb.SizeLength * mb.SizeWidth
		bannerIDInt := mb.ID

		var storeID *uuid.UUID
		if sid, ok := storeMap[mb.ShopID]; ok {
			storeID = &sid
		}

		var photoInstalled *string
		if mb.PhotoInstalled != nil && *mb.PhotoInstalled != "" {
			photoUrl := strings.TrimSpace(*mb.PhotoInstalled)
			if !strings.HasPrefix(photoUrl, "http") {
				photoUrl = "/uploads/" + strings.TrimPrefix(photoUrl, "/")
			}
			photoInstalled = &photoUrl
		}

		designURL := ""
		if mb.DesignLink != nil {
			designURL = *mb.DesignLink
		}

		shopName := mb.ShopName
		if shopName == "" {
			shopName = fmt.Sprintf("Shop #%d", mb.ShopID)
		}

		// MySQL Status normalization
		mysqlStatus := strings.ToLower(strings.TrimSpace(mb.Status))
		if mysqlStatus == "" {
			mysqlStatus = "request"
		}
		if mysqlStatus == "terpasang" || mysqlStatus == "selesai" || mysqlStatus == "selesai_dipasang" || photoInstalled != nil {
			mysqlStatus = "installed"
		}

		if mysqlStatus == "installed" || photoInstalled != nil {
			completedCount++
		}

		createdAt := time.Now()
		if mb.CreatedAt != nil {
			createdAt = *mb.CreatedAt
		}
		updatedAt := time.Now()
		if mb.UpdatedAt != nil {
			updatedAt = *mb.UpdatedAt
		}

		var existing domain.BannerOrder
		err := h.localDB.Where("mysql_banner_id = ?", bannerIDInt).First(&existing).Error

		if err == nil {
			// ── STRICT WORKFLOW PROTECTION ──
			// PostgreSQL Super App is the single source of truth for workflow progress.
			// MySQL Branding is only allowed to update status if it has real installation proof (photo_installed)
			// or if the PostgreSQL record has NO workflow progress yet (status == request/pending).
			existingRank := getStatusRank(existing.Status)
			mysqlRank := getStatusRank(mysqlStatus)

			targetStatus := existing.Status
			if photoInstalled != nil || mysqlStatus == "installed" {
				targetStatus = "installed"
			} else if existingRank == 0 && mysqlRank > 0 {
				targetStatus = mysqlStatus
			}

			// Preserve Super App design files & evidence
			finalDesignURL := existing.DesignURL
			if finalDesignURL == "" && designURL != "" {
				finalDesignURL = designURL
			}

			finalDesignLink := existing.DesignLink
			if finalDesignLink == nil || *finalDesignLink == "" {
				if mb.DesignLink != nil && *mb.DesignLink != "" {
					finalDesignLink = mb.DesignLink
				}
			}

			finalPhotoInstalled := existing.PhotoInstalled
			if photoInstalled != nil {
				finalPhotoInstalled = photoInstalled
			}

			finalInstalledAt := existing.InstalledAt
			if mb.InstalledAt != nil && finalInstalledAt == nil {
				finalInstalledAt = mb.InstalledAt
			}

			updates := map[string]interface{}{
				"mysql_banner_id":    bannerIDInt,
				"status":             targetStatus,
				"design_url":         finalDesignURL,
				"design_link":        finalDesignLink,
				"photo_installed":    finalPhotoInstalled,
				"installed_at":       finalInstalledAt,
				"spk_date":           mb.SPKDate,
				"total_price":        mb.TotalPrice,
				"operator_reward":    mb.OperatorReward,
				"is_bihun":           mb.IsBihun,
				"banner_type":        mb.Type,
				"qty":                mb.Qty,
				"size_length":        mb.SizeLength,
				"size_width":         mb.SizeWidth,
				"size":               size,
				"mysql_sales_name":   mb.SalesName,
				"mysql_editor_name":  mb.EditorName,
				"mysql_updated_at":   mb.UpdatedAt,
				"is_marked_customer": mb.IsMarkedCustomer,
			}

			// Only update notes / revision note if MySQL has value and existing is empty
			if mb.Notes != nil && *mb.Notes != "" && (existing.Notes == nil || *existing.Notes == "") {
				updates["notes"] = mb.Notes
			}
			if mb.RevisionNote != nil && *mb.RevisionNote != "" && (existing.RevisionNote == nil || *existing.RevisionNote == "") {
				updates["revision_note"] = mb.RevisionNote
			}
			if mb.EditingAt != nil && existing.EditingAt == nil {
				updates["editing_at"] = mb.EditingAt
			}
			if mb.EditingFinishedAt != nil && existing.EditingFinishedAt == nil {
				updates["editing_finished_at"] = mb.EditingFinishedAt
			}
			if mb.ValidatedAt != nil && existing.ValidatedAt == nil {
				updates["validated_at"] = mb.ValidatedAt
			}
			if mb.OperatorReceivedAt != nil && existing.OperatorReceivedAt == nil {
				updates["operator_received_at"] = mb.OperatorReceivedAt
			}
			if mb.VendorName != nil && *mb.VendorName != "" && (existing.VendorName == nil || *existing.VendorName == "") {
				updates["vendor_name"] = mb.VendorName
			}
			if mb.OperatorName != nil && *mb.OperatorName != "" && (existing.OperatorName == nil || *existing.OperatorName == "") {
				updates["operator_name"] = mb.OperatorName
			}
			// The Branding shop ID is authoritative. Older records may have been
			// linked to another store with the same name.
			if storeID != nil && (existing.StoreID == nil || *existing.StoreID != *storeID) {
				updates["store_id"] = storeID
			}
			if strings.TrimSpace(mb.ShopAddress) != "" {
				updates["location"] = mb.ShopAddress
			}

			if err := h.localDB.Model(&existing).Updates(updates).Error; err == nil {
				updatedCount++
			}
		} else {
			// Create new record in PostgreSQL
			empID := defaultEmp.ID
			newBO := domain.BannerOrder{
				ID:                 uuid.New(),
				CompanyID:          company.ID,
				EmployeeID:         empID,
				StoreID:            storeID,
				CampaignName:       "Branding Shop: " + shopName,
				StoreName:          shopName,
				Location:           mb.ShopAddress,
				BannerType:         mb.Type,
				Size:               size,
				Notes:              mb.Notes,
				RevisionNote:       mb.RevisionNote,
				IsDesignApproved:   mb.IsDesignApproved,
				IsStoreApproved:    mb.IsStoreApproved,
				Status:             mysqlStatus,
				DesignURL:          designURL,
				DesignLink:         mb.DesignLink,
				TotalPrice:         mb.TotalPrice,
				OperatorReward:     mb.OperatorReward,
				SPKDate:            mb.SPKDate,
				PhotoInstalled:     photoInstalled,
				InstalledAt:        mb.InstalledAt,
				CreatedAt:          createdAt,
				UpdatedAt:          updatedAt,
				EditingAt:          mb.EditingAt,
				EditingFinishedAt:  mb.EditingFinishedAt,
				ValidatedAt:        mb.ValidatedAt,
				OperatorReceivedAt: mb.OperatorReceivedAt,
				IsMarkedCustomer:   mb.IsMarkedCustomer,
				MySQLUpdatedAt:     mb.UpdatedAt,
				MySQLBannerID:      &bannerIDInt,
				MySQLSalesName:     mb.SalesName,
				MySQLEditorName:    mb.EditorName,
				VendorName:         mb.VendorName,
				OperatorName:       mb.OperatorName,
				Qty:                mb.Qty,
				SizeLength:         mb.SizeLength,
				SizeWidth:          mb.SizeWidth,
				IsBihun:            mb.IsBihun,
			}

			if err := h.localDB.Create(&newBO).Error; err == nil {
				createdCount++
			}
		}
	}

	// 3. Run Auto-Recovery on PostgreSQL records to ensure all active workflows are restored
	recoveredCount := h.executeAutoRecovery()

	log.Printf("[BrandingSync] ✅ Sinkronisasi selesai: %d total diproses, %d diupdate, %d baru, %d dipulihkan",
		len(myBanners), updatedCount, createdCount, recoveredCount)

	c.JSON(http.StatusOK, gin.H{
		"message":         "Sinkronisasi berhasil",
		"total_banners":   len(myBanners),
		"updated_count":   updatedCount,
		"created_count":   createdCount,
		"completed_count": completedCount,
		"recovered_count": recoveredCount,
	})
}

// executeAutoRecovery reconstructs correct status for all PostgreSQL banners based on their evidence
func (h *BrandingProxyHandler) executeAutoRecovery() int {
	var banners []domain.BannerOrder
	if err := h.localDB.Find(&banners).Error; err != nil {
		return 0
	}

	recovered := 0
	for _, b := range banners {
		curStatus := strings.ToLower(strings.TrimSpace(b.Status))
		targetStatus := curStatus

		hasRealDesignVal := (b.DesignValidatorName != nil && strings.TrimSpace(*b.DesignValidatorName) != "") || b.DesignValidatorID != nil
		hasRealStoreVal := (b.StoreValidatorName != nil && strings.TrimSpace(*b.StoreValidatorName) != "") || b.StoreValidatorID != nil
		hasGenericVal := (b.ValidatorName != nil && strings.TrimSpace(*b.ValidatorName) != "") || b.ValidatorID != nil
		isLegitDualValidated := b.IsDesignApproved && b.IsStoreApproved && (hasRealDesignVal || hasRealStoreVal || hasGenericVal)

		updatesToApply := map[string]interface{}{}

		// Priority 1: Installed (Photo or InstalledAt exists)
		if (b.PhotoInstalled != nil && *b.PhotoInstalled != "") || b.InstalledAt != nil {
			targetStatus = "installed"
		} else if b.OperatorID != nil || (b.OperatorName != nil && *b.OperatorName != "") || b.OperatorReceivedAt != nil {
			// Priority 2: Operator (Assigned to Operator)
			if curStatus != "installed" && curStatus != "terpasang" {
				if b.OperatorReceivedAt != nil {
					targetStatus = "diterima_operator"
				} else {
					targetStatus = "ready_for_operator"
				}
			}
		} else if b.VendorID != nil || (b.VendorName != nil && *b.VendorName != "") {
			// Priority 3: Vendor / Percetakan (Explicitly assigned to a Vendor)
			if curStatus != "installed" && curStatus != "terpasang" && curStatus != "ready_for_operator" && curStatus != "diterima_operator" {
				targetStatus = "sent_to_vendor"
			}
		} else if isLegitDualValidated {
			// Priority 4: Fully Validated by Dual Validators (Desain & Toko) -> Ready for Purchasing / Vendor Queue
			if curStatus != "installed" && curStatus != "terpasang" && curStatus != "ready_for_operator" && curStatus != "diterima_operator" && curStatus != "sent_to_vendor" {
				targetStatus = "purchasing_queue"
			}
		} else if curStatus == "editing" || curStatus == "in_design" {
			// Priority 4b: Banner sedang aktif dikerjakan/direvisi oleh editor — pertahankan status editing
			targetStatus = "editing"
		} else if (b.RevisionNote != nil && strings.TrimSpace(*b.RevisionNote) != "") && (!b.IsDesignApproved || !b.IsStoreApproved) && curStatus != "purchasing_queue" && curStatus != "sent_to_vendor" && curStatus != "ready_for_operator" && curStatus != "diterima_operator" && curStatus != "installed" && curStatus != "terpasang" {
			// Priority 4c: Banner memiliki catatan revisi dan belum di-ACC penuh -> kembalikan ke pengerjaan editor
			targetStatus = "editing"
		} else if b.IsDesignApproved && !b.IsStoreApproved {
			// Hanya Desain yang disetujui, masih menunggu Validator Toko
			targetStatus = "val_design"
		} else if !b.IsDesignApproved && b.IsStoreApproved {
			// Hanya Toko yang disetujui, masih menunggu Validator Desain
			targetStatus = "val_store"
		} else if (b.DesignURL != "" || (b.DesignLink != nil && *b.DesignLink != "")) || b.EditingFinishedAt != nil {
			// Priority 5: Design uploaded -> Validation
			targetStatus = "validation"
			if !isLegitDualValidated && (curStatus == "purchasing_queue" || curStatus == "sent_to_vendor") {
				// Kembalikan banner yang belum validasi resmi dari antrean vendor ke antrean validasi
				updatesToApply["is_design_approved"] = false
				updatesToApply["is_store_approved"] = false
			}
		} else if b.EditorID != nil || b.EditingAt != nil {
			// Priority 6: Claimed / Assigned to Editor
			if curStatus == "request" || curStatus == "pending" {
				targetStatus = "editing"
			}
		}

		if targetStatus != curStatus || len(updatesToApply) > 0 {
			updatesToApply["status"] = targetStatus
			if err := h.localDB.Model(&b).Updates(updatesToApply).Error; err == nil {
				recovered++
			}
		}
	}

	return recovered
}

// RecoverBrandingWorkflow handles POST /branding/recover
func (h *BrandingProxyHandler) RecoverBrandingWorkflow(c *gin.Context) {
	recovered := h.executeAutoRecovery()
	c.JSON(http.StatusOK, gin.H{
		"message":         fmt.Sprintf("Berhasil merekonstruksi dan memulihkan status %d banner", recovered),
		"recovered_count": recovered,
	})
}
