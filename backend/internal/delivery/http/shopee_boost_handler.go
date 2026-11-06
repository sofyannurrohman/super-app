package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	gosync "sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

var autoBoostMutex gosync.Mutex

type ShopeeBoostHandler struct {
	DB     *gorm.DB
	Client *ShopeeClient
}

func NewShopeeBoostHandler(db *gorm.DB) *ShopeeBoostHandler {
	return &ShopeeBoostHandler{
		DB:     db,
		Client: NewShopeeClient(db),
	}
}

func (h *ShopeeBoostHandler) RegisterRoutes(rg *gin.RouterGroup) {
	boost := rg.Group("/marketplaces/shopee/boost")
	{
		boost.GET("/overview", h.GetOverview)
		boost.POST("/toggle", h.ToggleAutoBoost)
		boost.GET("/queue", h.GetQueue)
		boost.POST("/queue", h.AddToQueue)
		boost.DELETE("/queue/:id", h.DeleteFromQueue)
		boost.PATCH("/queue/:id/toggle", h.ToggleItemActive)
		boost.PATCH("/queue/reorder", h.ReorderQueue)
		boost.POST("/trigger-now", h.TriggerNow)
		boost.GET("/logs", h.GetLogs)
	}
}

// resolveBranchAndShop finds the target Shopee token for given branchID or first available
func (h *ShopeeBoostHandler) resolveToken(branchIDStr string) (*domain.MarketplaceToken, error) {
	var branchID *uuid.UUID
	if branchIDStr != "" && branchIDStr != "all" {
		if id, err := uuid.Parse(branchIDStr); err == nil && id != uuid.Nil {
			branchID = &id
		}
	}
	return GetValidShopeeToken(h.DB, branchID)
}

// GetOverview handles GET /api/v1/marketplaces/shopee/boost/overview
func (h *ShopeeBoostHandler) GetOverview(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	token, err := h.resolveToken(branchIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"data": gin.H{
				"is_connected": false,
				"message":      "Shopee belum terhubung atau token tidak valid: " + err.Error(),
			},
		})
		return
	}

	// 1. Get Setting
	var setting domain.ShopeeAutoBoostSetting
	if err := h.DB.Where("shop_id = ?", token.ShopID).First(&setting).Error; err != nil {
		setting = domain.ShopeeAutoBoostSetting{
			BranchID:     token.BranchID,
			ShopID:       token.ShopID,
			IsEnabled:    false,
			RotationMode: "ROUND_ROBIN",
		}
		_ = h.DB.Create(&setting)
	}

	// 2. Fetch live boosted list from Shopee
	boostedList, apiErr := h.Client.GetBoostedList(token)
	if apiErr != nil {
		log.Printf("[ShopeeBoost] Warning: Failed to fetch live boosted list for shop %s: %v", token.ShopID, apiErr)
	}

	// 3. Sync live boosted items with DB state
	now := time.Now()
	boostedItemIDs := make([]int64, 0, len(boostedList))
	slotInfos := make([]domain.ShopeeBoostSlotInfo, 5)

	// Inisialisasi 5 slot kosong
	for i := 0; i < 5; i++ {
		slotInfos[i] = domain.ShopeeBoostSlotInfo{
			SlotIndex:  i + 1,
			IsOccupied: false,
		}
	}

	// Isi slot dari boostedList
	for i, b := range boostedList {
		boostedItemIDs = append(boostedItemIDs, b.ItemID)
		if i < 5 {
			cooldownSec := b.GetCooldownSeconds()
			cooldownUntil := now.Add(time.Duration(cooldownSec) * time.Second)
			slotInfos[i].ItemID = b.ItemID
			slotInfos[i].CooldownSeconds = cooldownSec
			slotInfos[i].CooldownUntil = cooldownUntil
			slotInfos[i].IsOccupied = true

			// Cari nama & gambar dari DB shopee_auto_boost_items atau marketplace_products
			var item domain.ShopeeAutoBoostItem
			if err := h.DB.Where("shop_id = ? AND item_id = ?", token.ShopID, b.ItemID).First(&item).Error; err == nil {
				slotInfos[i].ItemName = item.ItemName
				slotInfos[i].ImageURL = item.ImageURL

				// Perbarui status dan cooldown di DB jika beda
				h.DB.Model(&item).Updates(map[string]interface{}{
					"status":         "BOOSTED",
					"cooldown_until": cooldownUntil,
				})
			} else {
				var mp domain.MarketplaceProduct
				if err := h.DB.Where("product_id = ? AND marketplace = 'shopee'", strconv.FormatInt(b.ItemID, 10)).First(&mp).Error; err == nil {
					slotInfos[i].ItemName = mp.ProductName
				} else {
					slotInfos[i].ItemName = fmt.Sprintf("Produk Shopee #%d", b.ItemID)
				}
			}
		}
	}

	// 4. Update status produk di DB yang tadinya BOOSTED tapi sudah tidak ada di boostedList
	if len(boostedItemIDs) > 0 {
		h.DB.Model(&domain.ShopeeAutoBoostItem{}).
			Where("shop_id = ? AND status = 'BOOSTED' AND item_id NOT IN (?)", token.ShopID, boostedItemIDs).
			Updates(map[string]interface{}{
				"status": "IDLE",
			})
	} else if apiErr == nil {
		// Jika Shopee bilang list kosong, reset semua status BOOSTED di toko ini
		h.DB.Model(&domain.ShopeeAutoBoostItem{}).
			Where("shop_id = ? AND status = 'BOOSTED'", token.ShopID).
			Updates(map[string]interface{}{
				"status": "IDLE",
			})
	}

	// 5. Total dan Active di antrean
	var totalInQueue int64
	var activeInQueue int64
	h.DB.Model(&domain.ShopeeAutoBoostItem{}).Where("shop_id = ?", token.ShopID).Count(&totalInQueue)
	h.DB.Model(&domain.ShopeeAutoBoostItem{}).Where("shop_id = ? AND is_active = true", token.ShopID).Count(&activeInQueue)

	// 6. Recent logs
	var logs []domain.ShopeeAutoBoostLog
	h.DB.Where("shop_id = ?", token.ShopID).Order("boosted_at desc").Limit(10).Find(&logs)

	occupiedSlots := len(boostedList)
	if occupiedSlots > 5 {
		occupiedSlots = 5
	}
	availableSlots := 5 - occupiedSlots

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": domain.ShopeeBoostOverviewResponse{
			ShopID:         token.ShopID,
			IsEnabled:      setting.IsEnabled,
			MaxSlots:       5,
			OccupiedSlots:  occupiedSlots,
			AvailableSlots: availableSlots,
			TotalInQueue:   totalInQueue,
			ActiveInQueue:  activeInQueue,
			Slots:          slotInfos,
			Settings:       setting,
			RecentLogs:     logs,
		},
	})
}

// ToggleAutoBoost handles POST /api/v1/marketplaces/shopee/boost/toggle
func (h *ShopeeBoostHandler) ToggleAutoBoost(c *gin.Context) {
	var req struct {
		BranchID  string `json:"branch_id"`
		IsEnabled *bool  `json:"is_enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	token, err := h.resolveToken(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shopee token not found: " + err.Error()})
		return
	}

	var setting domain.ShopeeAutoBoostSetting
	if err := h.DB.Where("shop_id = ?", token.ShopID).First(&setting).Error; err != nil {
		setting = domain.ShopeeAutoBoostSetting{
			BranchID:     token.BranchID,
			ShopID:       token.ShopID,
			IsEnabled:    *req.IsEnabled,
			RotationMode: "ROUND_ROBIN",
		}
		if err := h.DB.Create(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting: " + err.Error()})
			return
		}
	} else {
		setting.IsEnabled = *req.IsEnabled
		if err := h.DB.Save(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Auto Boost berhasil %s", func() string {
			if *req.IsEnabled {
				return "diaktifkan"
			}
			return "dinonaktifkan"
		}()),
		"data": setting,
	})
}

// GetQueue handles GET /api/v1/marketplaces/shopee/boost/queue
func (h *ShopeeBoostHandler) GetQueue(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	token, err := h.resolveToken(branchIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []domain.ShopeeAutoBoostItem{},
			"total":   0,
		})
		return
	}

	search := strings.TrimSpace(c.Query("search"))
	q := h.DB.Model(&domain.ShopeeAutoBoostItem{}).Where("shop_id = ?", token.ShopID)
	if search != "" {
		q = q.Where("item_name ILIKE ? OR item_sku ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	q.Count(&total)

	var items []domain.ShopeeAutoBoostItem
	if err := q.Order("sequence asc, id asc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queue: " + err.Error()})
		return
	}

	// Isi current_stock dan product_status terkini dari marketplace_products
	for i := range items {
		itemIDStr := strconv.FormatInt(items[i].ItemID, 10)
		var totalStock int
		h.DB.Model(&domain.MarketplaceProduct{}).
			Where("product_id = ? AND marketplace = 'shopee'", itemIDStr).
			Select("COALESCE(SUM(stock), 0)").
			Scan(&totalStock)
		items[i].CurrentStock = totalStock

		var status string
		h.DB.Model(&domain.MarketplaceProduct{}).
			Where("product_id = ? AND marketplace = 'shopee'", itemIDStr).
			Select("status").
			Limit(1).
			Scan(&status)
		items[i].ProductStatus = status
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"total":   total,
	})
}

// AddToQueue handles POST /api/v1/marketplaces/shopee/boost/queue
func (h *ShopeeBoostHandler) AddToQueue(c *gin.Context) {
	var req struct {
		BranchID string  `json:"branch_id"`
		ItemIDs  []int64 `json:"item_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if len(req.ItemIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_ids cannot be empty"})
		return
	}

	token, err := h.resolveToken(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shopee token not found: " + err.Error()})
		return
	}

	// Temukan sequence tertinggi yang sudah ada
	var maxSeq int
	h.DB.Model(&domain.ShopeeAutoBoostItem{}).Where("shop_id = ?", token.ShopID).Select("COALESCE(MAX(sequence), 0)").Scan(&maxSeq)

	addedCount := 0
	for _, itemID := range req.ItemIDs {
		// Cek apakah sudah ada di antrean
		var count int64
		h.DB.Model(&domain.ShopeeAutoBoostItem{}).Where("shop_id = ? AND item_id = ?", token.ShopID, itemID).Count(&count)
		if count > 0 {
			continue
		}

		// Cari informasi produk dari marketplace_products
		var mp domain.MarketplaceProduct
		itemName := fmt.Sprintf("Produk Shopee #%d", itemID)
		itemSKU := ""
		price := 0.0
		qProd := h.DB.Where("product_id = ? AND marketplace = 'shopee'", strconv.FormatInt(itemID, 10))
		if token.BranchID != nil && *token.BranchID != uuid.Nil {
			qProd = qProd.Where("branch_id = ?", *token.BranchID)
		}
		if err := qProd.First(&mp).Error; err != nil {
			h.DB.Where("product_id = ? AND marketplace = 'shopee'", strconv.FormatInt(itemID, 10)).First(&mp)
		}
		if mp.ID > 0 {
			itemName = mp.ProductName
			itemSKU = mp.SellerSKU
			price = mp.PriceAmount
		}

		var totalStock int
		qStock := h.DB.Model(&domain.MarketplaceProduct{}).
			Where("product_id = ? AND marketplace = 'shopee'", strconv.FormatInt(itemID, 10))
		if token.BranchID != nil && *token.BranchID != uuid.Nil {
			qStock = qStock.Where("branch_id = ?", *token.BranchID)
		}
		qStock.Select("COALESCE(SUM(stock), 0)").Scan(&totalStock)

		status := "IDLE"
		lastError := ""
		if totalStock <= 0 {
			status = "FAILED"
			lastError = "Stok produk 0 di Shopee (wajib > 0 untuk boost)"
		}

		maxSeq++
		newItem := domain.ShopeeAutoBoostItem{
			BranchID:    token.BranchID,
			ShopID:      token.ShopID,
			ItemID:      itemID,
			ItemName:    itemName,
			ItemSKU:     itemSKU,
			PriceAmount: price,
			Sequence:    maxSeq,
			IsActive:    true,
			Status:      status,
			LastError:   lastError,
		}
		if err := h.DB.Create(&newItem).Error; err == nil {
			addedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     fmt.Sprintf("%d produk berhasil ditambahkan ke antrean", addedCount),
		"added_count": addedCount,
	})
}

// DeleteFromQueue handles DELETE /api/v1/marketplaces/shopee/boost/queue/:id
func (h *ShopeeBoostHandler) DeleteFromQueue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	if err := h.DB.Delete(&domain.ShopeeAutoBoostItem{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Produk berhasil dihapus dari antrean boost",
	})
}

// ToggleItemActive handles PATCH /api/v1/marketplaces/shopee/boost/queue/:id/toggle
func (h *ShopeeBoostHandler) ToggleItemActive(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req struct {
		IsActive *bool `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if err := h.DB.Model(&domain.ShopeeAutoBoostItem{}).Where("id = ?", id).Update("is_active", *req.IsActive).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status produk berhasil diperbarui",
	})
}

// ReorderQueue handles PATCH /api/v1/marketplaces/shopee/boost/queue/reorder
func (h *ShopeeBoostHandler) ReorderQueue(c *gin.Context) {
	var req struct {
		ItemIDs []uint `json:"item_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	tx := h.DB.Begin()
	for seq, id := range req.ItemIDs {
		if err := tx.Model(&domain.ShopeeAutoBoostItem{}).Where("id = ?", id).Update("sequence", seq+1).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder: " + err.Error()})
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Urutan antrean berhasil disimpan",
	})
}

// TriggerNow handles POST /api/v1/marketplaces/shopee/boost/trigger-now (manual boost trigger)
func (h *ShopeeBoostHandler) TriggerNow(c *gin.Context) {
	var req struct {
		BranchID string `json:"branch_id"`
	}
	_ = c.ShouldBindJSON(&req)

	token, err := h.resolveToken(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shopee token not found: " + err.Error()})
		return
	}

	boostedCount, msg, runErr := h.ExecuteAutoBoostForShop(token, "MANUAL")
	if runErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   runErr.Error(),
			"message": msg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"boosted_count": boostedCount,
		"message":       msg,
	})
}

// GetLogs handles GET /api/v1/marketplaces/shopee/boost/logs
func (h *ShopeeBoostHandler) GetLogs(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	token, err := h.resolveToken(branchIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []domain.ShopeeAutoBoostLog{},
		})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var logs []domain.ShopeeAutoBoostLog
	h.DB.Where("shop_id = ?", token.ShopID).Order("boosted_at desc").Limit(limit).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// ExecuteAutoBoostForShop executes the round-robin boost for a single shop
func (h *ShopeeBoostHandler) ExecuteAutoBoostForShop(token *domain.MarketplaceToken, triggeredBy string) (int, string, error) {
	autoBoostMutex.Lock()
	defer autoBoostMutex.Unlock()

	now := time.Now()

	// 1. Get live boosted list from Shopee
	boostedList, err := h.Client.GetBoostedList(token)
	if err != nil {
		return 0, fmt.Sprintf("Gagal memeriksa slot boost dari Shopee: %v", err), err
	}

	occupiedCount := len(boostedList)
	availableSlots := 5 - occupiedCount
	if availableSlots <= 0 {
		return 0, "Semua 5 slot boost Shopee saat ini sedang penuh. Antrean akan berjalan otomatis saat durasi countdown slot selesai.", nil
	}

	// 2. Kumpulkan item_id yang saat ini sedang aktif diboost agar tidak dipilih ganda
	activeBoostedIDs := make([]int64, 0, len(boostedList))
	for _, b := range boostedList {
		activeBoostedIDs = append(activeBoostedIDs, b.ItemID)
	}

	// 3. Query candidate produk dari tabel shopee_auto_boost_items
	// Kriteria:
	// - shop_id sama
	// - is_active = true
	// - tidak ada di activeBoostedIDs
	// - cooldown_until NULL atau <= now
	// Urutan Smart Round-Robin:
	// - last_boosted_at ASC NULLS FIRST (yang paling lama belum diboost atau belum pernah diboost)
	// - sequence ASC (prioritas urutan)
	q := h.DB.Model(&domain.ShopeeAutoBoostItem{}).
		Where("shop_id = ? AND is_active = true", token.ShopID).
		Where("cooldown_until IS NULL OR cooldown_until <= ?", now)

	if len(activeBoostedIDs) > 0 {
		q = q.Where("item_id NOT IN (?)", activeBoostedIDs)
	}

	var candidates []domain.ShopeeAutoBoostItem
	if err := q.Order("last_boosted_at asc nulls first, sequence asc, id asc").
		Limit(25).
		Find(&candidates).Error; err != nil {
		return 0, "Gagal mengambil antrean produk dari database", err
	}

	if len(candidates) == 0 {
		return 0, "Tidak ada produk di antrean yang siap dinaikkan (semua produk masih dalam cooldown atau antrean kosong).", nil
	}

	cooldownUntil := now.Add(4 * time.Hour)
	successCount := 0
	skippedCount := 0
	failedCount := 0
	var lastErrMsg string

	for _, c := range candidates {
		if availableSlots <= 0 {
			break
		}

		itemIDStr := strconv.FormatInt(c.ItemID, 10)

		// 4. Validasi stok dan status produk di database marketplace_products
		var totalStock int
		var statuses []string
		h.DB.Model(&domain.MarketplaceProduct{}).
			Where("product_id = ? AND marketplace = 'shopee'", itemIDStr).
			Select("COALESCE(SUM(stock), 0)").
			Scan(&totalStock)

		h.DB.Model(&domain.MarketplaceProduct{}).
			Where("product_id = ? AND marketplace = 'shopee'", itemIDStr).
			Pluck("DISTINCT status", &statuses)

		isStatusInvalid := false
		if len(statuses) > 0 {
			hasNormal := false
			for _, st := range statuses {
				upper := strings.ToUpper(st)
				if upper == "NORMAL" || upper == "ACTIVATE" {
					hasNormal = true
					break
				}
			}
			if !hasNormal {
				isStatusInvalid = true
			}
		}

		if totalStock <= 0 || isStatusInvalid {
			reason := "Stok produk 0 di Shopee (wajib > 0 untuk boost)"
			if isStatusInvalid {
				reason = "Status produk tidak aktif/normal di Shopee"
			}

			// Tandai item FAILED dengan cooldown 1 jam agar tidak loop terus-menerus
			h.DB.Model(&domain.ShopeeAutoBoostItem{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
				"status":         "FAILED",
				"last_error":     reason,
				"cooldown_until": now.Add(1 * time.Hour),
			})

			h.DB.Create(&domain.ShopeeAutoBoostLog{
				ShopID:      token.ShopID,
				ItemID:      c.ItemID,
				ItemName:    c.ItemName,
				TriggeredBy: triggeredBy,
				Status:      "FAILED",
				Message:     fmt.Sprintf("Dilewati: %s", reason),
				BoostedAt:   now,
			})

			skippedCount++
			continue
		}

		// 5. Panggil Shopee API BoostItem secara terisolasi per produk
		boostResp, boostErr := h.Client.BoostItem(token, []int64{c.ItemID})
		if boostErr != nil || (boostResp != nil && len(boostResp.Response.FailureList) > 0) {
			failReason := "Gagal menaikkan produk ke Shopee"
			if boostErr != nil {
				failReason = boostErr.Error()
			}
			if boostResp != nil && len(boostResp.Response.FailureList) > 0 {
				failReason = boostResp.Response.FailureList[0].FailedReason
			}
			lastErrMsg = failReason

			h.DB.Model(&domain.ShopeeAutoBoostItem{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
				"status":         "FAILED",
				"last_error":     failReason,
				"cooldown_until": now.Add(1 * time.Hour),
			})

			h.DB.Create(&domain.ShopeeAutoBoostLog{
				ShopID:      token.ShopID,
				ItemID:      c.ItemID,
				ItemName:    c.ItemName,
				TriggeredBy: triggeredBy,
				Status:      "FAILED",
				Message:     failReason,
				BoostedAt:   now,
			})

			failedCount++
			continue
		}

		// Berhasil diboost
		successCount++
		availableSlots--
		h.DB.Model(&domain.ShopeeAutoBoostItem{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
			"status":          "BOOSTED",
			"last_boosted_at": now,
			"cooldown_until":  cooldownUntil,
			"last_error":      "",
		})

		h.DB.Create(&domain.ShopeeAutoBoostLog{
			ShopID:      token.ShopID,
			ItemID:      c.ItemID,
			ItemName:    c.ItemName,
			TriggeredBy: triggeredBy,
			Status:      "SUCCESS",
			Message:     "Berhasil dinaikkan (Boost 4 jam)",
			BoostedAt:   now,
		})
	}

	if successCount == 0 && (skippedCount > 0 || failedCount > 0) {
		msg := fmt.Sprintf("Tidak ada produk yang berhasil dinaikkan. %d produk dilewati (stok 0 / nonaktif), %d gagal dari Shopee (%s).", skippedCount, failedCount, lastErrMsg)
		return 0, msg, fmt.Errorf("%s", msg)
	}

	msg := fmt.Sprintf("Berhasil menaikkan %d produk ke Shopee.", successCount)
	if skippedCount > 0 {
		msg += fmt.Sprintf(" (%d produk dilewati karena stok 0 / nonaktif)", skippedCount)
	}
	if failedCount > 0 {
		msg += fmt.Sprintf(" (%d produk gagal dari Shopee)", failedCount)
	}

	return successCount, msg, nil
}

// RunAutoBoostForAllShops dipanggil oleh Background Cron setiap 5 menit
func (h *ShopeeBoostHandler) RunAutoBoostForAllShops() {
	var activeSettings []domain.ShopeeAutoBoostSetting
	if err := h.DB.Where("is_enabled = true").Find(&activeSettings).Error; err != nil {
		log.Printf("[ShopeeAutoBoostWorker] Gagal mengambil pengaturan aktif: %v", err)
		return
	}

	if len(activeSettings) == 0 {
		return
	}

	log.Printf("[ShopeeAutoBoostWorker] Menjalankan pengecekan auto boost untuk %d toko aktif...", len(activeSettings))

	for _, setting := range activeSettings {
		token, err := GetValidShopeeToken(h.DB, setting.BranchID)
		if err != nil {
			log.Printf("[ShopeeAutoBoostWorker] Token Shopee untuk Shop %s tidak valid: %v", setting.ShopID, err)
			continue
		}

		now := time.Now()
		count, msg, runErr := h.ExecuteAutoBoostForShop(token, "WORKER")

		status := "SUCCESS"
		if runErr != nil {
			status = "ERROR"
		}

		// Update last run info di setting
		h.DB.Model(&setting).Updates(map[string]interface{}{
			"last_run_at":      now,
			"last_run_status":  status,
			"last_run_message": fmt.Sprintf("[%d item boosted] %s", count, msg),
		})
	}
}
