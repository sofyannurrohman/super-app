package sync

import (
	"fmt"
	"log"
	"os"
	"strings"
	gosync "sync"
	"time"

	"example.com/hris/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	bannerSyncMutex gosync.Mutex
	lastBannerSync  time.Time
)

func SyncBannersFromMySQL(localDB *gorm.DB, mysqlDB *gorm.DB, companyID uuid.UUID) error {
	if mysqlDB == nil {
		return nil // Nothing to sync
	}

	// Prevent concurrent duplicate syncs
	bannerSyncMutex.Lock()
	defer bannerSyncMutex.Unlock()

	log.Println("[BrandingSync] Starting banner sync from MySQL to PostgreSQL...")

	// Automatically add the new mysql_sales_name and mysql_editor_name columns
	localDB.AutoMigrate(&domain.BannerOrder{})

	// 1. Get the latest mysql_banner_id in postgres
	var lastSynced int64
	localDB.Table("banner_orders").Select("COALESCE(MAX(mysql_banner_id), 0)").Scan(&lastSynced)

	// 1b. Get the last sync time
	maxUpdatedAt := lastBannerSync
	if maxUpdatedAt.IsZero() {
		// Fallback for the first run after restart to catch up any missed updates
		maxUpdatedAt = time.Now().AddDate(0, -6, 0) // 6 months
	}

	// 2. Fetch new AND updated banners from MySQL
	type MySQLBanner struct {
		ID             int64      `gorm:"column:id"`
		ShopName       string     `gorm:"column:shop_name"`
		ShopAddress    *string    `gorm:"column:shop_address"`
		Type           string     `gorm:"column:type"`
		Qty            int        `gorm:"column:qty"`
		SizeLength     float64    `gorm:"column:size_length"`
		SizeWidth      float64    `gorm:"column:size_width"`
		Notes          *string    `gorm:"column:notes"`
		Status         string     `gorm:"column:status"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
		UpdatedAt      time.Time  `gorm:"column:updated_at"`
		SalesName      *string    `gorm:"column:sales_name"`
		EditorName     *string    `gorm:"column:editor_name"`
		OperatorID     *int64     `gorm:"column:operator_id"`
		OperatorName   *string    `gorm:"column:operator_name"`
		OperatorReward float64    `gorm:"column:operator_reward"`
		PhotoInstalled *string    `gorm:"column:photo_installed"`
		InstalledAt    *time.Time `gorm:"column:installed_at"`
	}

	var newBanners []MySQLBanner
	err := mysqlDB.Table("banners").
		Select("banners.*, shops.name as shop_name, shops.address as shop_address, sales.name as sales_name, editors.name as editor_name").
		Joins("LEFT JOIN shops ON shops.id = banners.shop_id").
		Joins("LEFT JOIN users sales ON sales.id = banners.sales_id").
		Joins("LEFT JOIN users editors ON editors.id = banners.editor_id").
		Joins("LEFT JOIN users operators ON operators.id = banners.operator_id").
		Where("banners.id > ? OR banners.updated_at > ?", lastSynced, maxUpdatedAt).
		Find(&newBanners).Error

	if err != nil {
		log.Printf("[BrandingSync] Failed to fetch banners from MySQL: %v", err)
		return err
	}

	if len(newBanners) == 0 {
		return nil
	}

	log.Printf("[BrandingSync] Found %d new banners to sync", len(newBanners))

	// Find the superadmin or default employee to assign as the creator for synced data
	var defaultEmployee domain.Employee
	localDB.Where("company_id = ?", companyID).First(&defaultEmployee)

	// Pre-load all employees in memory to map by name quickly
	var allEmployees []domain.Employee
	localDB.Where("company_id = ?", companyID).Find(&allEmployees)

	employeeNameMap := make(map[string]uuid.UUID)
	for _, emp := range allEmployees {
		name := strings.ToLower(strings.TrimSpace(emp.FirstName))
		employeeNameMap[name] = emp.ID
		if emp.LastName != "" {
			fullName := strings.ToLower(strings.TrimSpace(emp.FirstName + " " + emp.LastName))
			employeeNameMap[fullName] = emp.ID
		}
	}

	for _, mb := range newBanners {
		// Legacy branding records can contain zero even though a banner order
		// always represents at least one physical banner.
		qty := mb.Qty
		if qty <= 0 {
			qty = 1
		}
		// Calculate size
		size := mb.SizeLength * mb.SizeWidth

		status := mb.Status
		if status == "" {
			status = "request"
		}

		empID := defaultEmployee.ID
		if mb.SalesName != nil {
			if id, ok := employeeNameMap[strings.ToLower(strings.TrimSpace(*mb.SalesName))]; ok {
				empID = id
			}
		}

		var editorID *uuid.UUID
		if mb.EditorName != nil {
			if id, ok := employeeNameMap[strings.ToLower(strings.TrimSpace(*mb.EditorName))]; ok {
				editorID = &id
			}
		}

		location := "N/A"
		if mb.ShopAddress != nil && *mb.ShopAddress != "" {
			location = *mb.ShopAddress
		}

		// Format photo URL if it exists
		var finalPhotoInstalled *string
		if mb.PhotoInstalled != nil && *mb.PhotoInstalled != "" {
			photoUrl := *mb.PhotoInstalled
			photoUrl = externalBrandingPhotoURL(photoUrl)
			finalPhotoInstalled = &photoUrl
		}

		// Check if this banner was already imported
		var existingBanner domain.BannerOrder
		err := localDB.Where("mysql_banner_id = ?", mb.ID).First(&existingBanner).Error

		if err == nil {
			// PostgreSQL Super App adalah Source of Truth.
			// Jangan timpa status yang sudah dikelola oleh operator Super App.
			// Namun tetap update field non-status (qty, size, notes) dari MySQL
			// agar data operasional tetap fresh.
			currentStatus := strings.ToLower(existingBanner.Status)

			updates := map[string]interface{}{
				"mysql_updated_at":  mb.UpdatedAt,
				"mysql_sales_name":  mb.SalesName,
				"mysql_editor_name": mb.EditorName,
				"qty":               qty,
				"size_length":       mb.SizeLength,
				"size_width":        mb.SizeWidth,
				"size":              size,
				"notes":             mb.Notes,
				"banner_type":       mb.Type,
			}

			// Hanya update status jika Super App belum mengambil alih banner ini.
			// Status terminal diatur oleh operator lapangan dan tidak boleh ditimpa MySQL.
			// Status editing/revisi di Super App juga dilindungi agar tidak tertimpa status MySQL lama.
			statusLower := strings.ToLower(strings.TrimSpace(status))
			if currentStatus != "diterima_operator" && currentStatus != "installed" &&
				currentStatus != "terpasang" && currentStatus != "selesai_dipasang" {
				if (currentStatus != "editing" && currentStatus != "in_design") || statusLower == "installed" || statusLower == "terpasang" {
					updates["status"] = status
				}
			}

			if editorID != nil {
				updates["editor_id"] = editorID
			}
			if finalPhotoInstalled != nil {
				updates["photo_installed"] = finalPhotoInstalled
			}
			if mb.InstalledAt != nil {
				updates["installed_at"] = mb.InstalledAt
			}
			updates["operator_reward"] = mb.OperatorReward
			if mb.OperatorID != nil {
				var operator domain.Employee
				if err := localDB.Where("employee_id_number = ?", fmt.Sprintf("BRANDING-OPERATOR-%d", *mb.OperatorID)).First(&operator).Error; err == nil {
					updates["operator_id"] = operator.ID
				}
			}
			if mb.OperatorName != nil && strings.TrimSpace(*mb.OperatorName) != "" {
				updates["operator_name"] = strings.TrimSpace(*mb.OperatorName)
			}
			localDB.Model(&existingBanner).Updates(updates)
			continue
		}

		bo := domain.BannerOrder{
			ID:              uuid.New(),
			CompanyID:       companyID,
			EmployeeID:      empID,
			CampaignName:    "Branding Shop: " + mb.ShopName,
			StoreName:       mb.ShopName,
			Location:        location,
			BannerType:      mb.Type,
			Size:            size,
			Notes:           mb.Notes,
			Status:          status,
			CreatedAt:       mb.CreatedAt,
			UpdatedAt:       mb.UpdatedAt,
			MySQLUpdatedAt:  &mb.UpdatedAt,
			MySQLBannerID:   &mb.ID,
			MySQLSalesName:  mb.SalesName,
			MySQLEditorName: mb.EditorName,
			Qty:             qty,
			SizeLength:      mb.SizeLength,
			SizeWidth:       mb.SizeWidth,
			EditorID:        editorID,
			PhotoInstalled:  finalPhotoInstalled,
			InstalledAt:     mb.InstalledAt,
			OperatorReward:  mb.OperatorReward,
		}
		if mb.OperatorID != nil {
			var operator domain.Employee
			if err := localDB.Where("employee_id_number = ?", fmt.Sprintf("BRANDING-OPERATOR-%d", *mb.OperatorID)).First(&operator).Error; err == nil {
				bo.OperatorID = &operator.ID
				bo.OperatorName = mb.OperatorName
			}
		}

		if err := localDB.Create(&bo).Error; err != nil {
			log.Printf("[BrandingSync] Failed to sync banner %d: %v", mb.ID, err)
		}
	}

	lastBannerSync = time.Now()
	log.Println("[BrandingSync] Banner sync completed.")
	return nil
}

func externalBrandingPhotoURL(photoURL string) string {
	if strings.HasPrefix(photoURL, "http://") || strings.HasPrefix(photoURL, "https://") {
		return photoURL
	}
	base := strings.TrimRight(os.Getenv("BRANDING_STORAGE_BASE_URL"), "/")
	if base == "" {
		return "/uploads/" + strings.TrimLeft(photoURL, "/")
	}
	return base + "/" + strings.TrimLeft(photoURL, "/")
}
