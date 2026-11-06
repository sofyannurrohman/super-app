package http

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
	"tiktokshop/open/sdk_golang/apis"
)

type unsettledCaptureRequest struct {
	ShopID           string `json:"shop_id" binding:"required"`
	Source           string `json:"source" binding:"required"`
	AwaitingDelivery *int64 `json:"awaiting_delivery"`
	AfterSales       *int64 `json:"after_sales"`
	Processing       *int64 `json:"processing"`
}

type unsettledJob struct {
	Status     string `json:"status"`
	SnapshotID uint   `json:"snapshot_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

func (h *MarketplaceHandler) GetUnsettledJob(c *gin.Context) {
	h.unsettledMu.Lock()
	job, found := h.unsettledJobs[c.Param("id")]
	h.unsettledMu.Unlock()
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proses tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *MarketplaceHandler) ListUnsettledSnapshots(c *gin.Context) {
	if err := h.ensureUnsettledTables(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tabel Untuk Dibayar belum tersedia: " + err.Error()})
		return
	}
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}
	shops, err := h.tikTokFinanceShopIDs(branchIDPtr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := h.db.Model(&domain.MarketplaceUnsettledSnapshot{}).Where("marketplace = ? AND shop_id IN ?", "tiktok", shops)
	if shopID := strings.TrimSpace(c.Query("shop_id")); shopID != "" {
		if !containsShopID(shops, shopID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Toko TikTok tidak tersedia pada cabang ini"})
			return
		}
		query = query.Where("shop_id = ?", shopID)
	}
	var snapshots []domain.MarketplaceUnsettledSnapshot
	if err := query.Order("captured_at DESC, id DESC").Limit(30).Find(&snapshots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"shops": shops, "snapshots": snapshots})
}

func (h *MarketplaceHandler) GetUnsettledSnapshot(c *gin.Context) {
	if err := h.ensureUnsettledTables(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tabel Untuk Dibayar belum tersedia: " + err.Error()})
		return
	}
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}
	shops, err := h.tikTokFinanceShopIDs(branchIDPtr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var snapshot domain.MarketplaceUnsettledSnapshot
	if err := h.db.Where("id = ? AND marketplace = ? AND shop_id IN ?", c.Param("id"), "tiktok", shops).First(&snapshot).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Snapshot tidak ditemukan"})
		return
	}
	var transactions []domain.MarketplaceUnsettledTransaction
	if err := h.db.Where("snapshot_id = ?", snapshot.ID).Order("id ASC").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"snapshot": snapshot, "transactions": transactions})
}

func (h *MarketplaceHandler) CaptureUnsettledSnapshot(c *gin.Context) {
	if err := h.ensureUnsettledTables(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tabel Untuk Dibayar belum tersedia: " + err.Error()})
		return
	}
	var req unsettledCaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shop_id dan source wajib diisi"})
		return
	}
	req.ShopID = strings.TrimSpace(req.ShopID)
	req.Source = strings.ToUpper(strings.TrimSpace(req.Source))
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}
	shops, err := h.tikTokFinanceShopIDs(branchIDPtr)
	if err != nil || !containsShopID(shops, req.ShopID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Toko TikTok tidak tersedia pada cabang ini"})
		return
	}
	token, err := getValidTikTokTokenByShopID(h.db, h.tikTokAppKey, h.tikTokAppSecret, req.ShopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Akun TikTok tidak aktif: " + err.Error()})
		return
	}
	snapshot := domain.MarketplaceUnsettledSnapshot{
		Marketplace: "tiktok", ShopID: req.ShopID, BranchID: token.BranchID,
		CapturedAt: time.Now(), Currency: "IDR", Source: req.Source,
	}
	var transactions []domain.MarketplaceUnsettledTransaction
	switch req.Source {
	case "SELLER_CENTER":
		if req.AwaitingDelivery == nil || req.AfterSales == nil || req.Processing == nil || *req.AwaitingDelivery < 0 || *req.AfterSales < 0 || *req.Processing < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ketiga nominal Seller Center wajib diisi dan tidak boleh negatif"})
			return
		}
		snapshot.AwaitingDelivery = *req.AwaitingDelivery
		snapshot.AfterSales = *req.AfterSales
		snapshot.Processing = *req.Processing
	case "TIKTOK_API":
		jobID := uuid.NewString()
		h.unsettledMu.Lock()
		h.unsettledJobs[jobID] = unsettledJob{Status: "processing"}
		h.unsettledMu.Unlock()
		go h.captureUnsettledInBackground(jobID, snapshot)
		c.JSON(http.StatusAccepted, gin.H{"status": "processing", "job_id": jobID})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "source harus SELLER_CENTER atau TIKTOK_API"})
		return
	}
	if err := h.saveUnsettledSnapshot(&snapshot, transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan snapshot: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"snapshot": snapshot})
}

func (h *MarketplaceHandler) captureUnsettledInBackground(jobID string, snapshot domain.MarketplaceUnsettledSnapshot) {
	transactions, err := h.fetchUnsettledTransactions(context.Background(), snapshot.ShopID)
	if err == nil {
		for _, tx := range transactions {
			switch tx.Category {
			case "AWAITING_DELIVERY":
				snapshot.AwaitingDelivery += tx.EstimatedAmount
			case "AFTER_SALES":
				snapshot.AfterSales += tx.EstimatedAmount
			case "PROCESSING":
				snapshot.Processing += tx.EstimatedAmount
			default:
				snapshot.UnclassifiedAmount += tx.EstimatedAmount
			}
		}
		snapshot.TransactionCount = len(transactions)
		snapshot.CapturedAt = time.Now()
		err = h.saveUnsettledSnapshot(&snapshot, transactions)
	}
	h.unsettledMu.Lock()
	if err != nil {
		h.unsettledJobs[jobID] = unsettledJob{Status: "failed", Error: err.Error()}
	} else {
		h.unsettledJobs[jobID] = unsettledJob{Status: "completed", SnapshotID: snapshot.ID}
	}
	h.unsettledMu.Unlock()
}

func (h *MarketplaceHandler) saveUnsettledSnapshot(snapshot *domain.MarketplaceUnsettledSnapshot, transactions []domain.MarketplaceUnsettledTransaction) error {
	return h.db.Transaction(func(db *gorm.DB) error {
		if err := db.Create(snapshot).Error; err != nil {
			return err
		}
		for i := range transactions {
			transactions[i].SnapshotID = snapshot.ID
		}
		for start := 0; start < len(transactions); start += 500 {
			end := start + 500
			if end > len(transactions) {
				end = len(transactions)
			}
			if err := db.Create(transactions[start:end]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *MarketplaceHandler) ensureUnsettledTables() error {
	h.unsettledMu.Lock()
	defer h.unsettledMu.Unlock()
	if h.unsettledReady {
		return nil
	}
	if err := h.db.AutoMigrate(&domain.MarketplaceUnsettledSnapshot{}, &domain.MarketplaceUnsettledTransaction{}); err != nil {
		return err
	}
	h.unsettledReady = true
	return nil
}

func containsShopID(shops []string, target string) bool {
	for _, shop := range shops {
		if shop == target {
			return true
		}
	}
	return false
}

func parseIDRAmount(raw string) (int64, error) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) || amount > math.MaxInt64 || amount < math.MinInt64 {
		return 0, fmt.Errorf("nominal TikTok tidak valid: %q", raw)
	}
	return int64(math.Round(amount)), nil
}

func classifyUnsettled(reason string, deliveredAt int64) string {
	r := strings.ToUpper(strings.TrimSpace(reason))
	if strings.Contains(r, "AFTER_SALE") || strings.Contains(r, "REFUND") || strings.Contains(r, "RETURN") {
		return "AFTER_SALES"
	}
	if strings.Contains(r, "DELIVERY") || strings.Contains(r, "DELIVERED") || strings.Contains(r, "IN_TRANSIT") || strings.Contains(r, "SHIPPING") || strings.Contains(r, "RECEIVED") {
		return "AWAITING_DELIVERY"
	}
	if strings.Contains(r, "PROCESS") || strings.Contains(r, "SETTLEMENT") || strings.Contains(r, "WAITING") {
		return "PROCESSING"
	}
	if deliveredAt > 0 && r == "" {
		return "PROCESSING"
	}
	return "UNCLASSIFIED"
}

func (h *MarketplaceHandler) fetchUnsettledTransactions(ctx context.Context, shopID string) ([]domain.MarketplaceUnsettledTransaction, error) {
	token, err := getValidTikTokTokenByShopID(h.db, h.tikTokAppKey, h.tikTokAppSecret, shopID)
	if err != nil {
		return nil, err
	}
	cfg := apis.NewConfiguration()
	cfg.AddAppInfo(h.tikTokAppKey, h.tikTokAppSecret)
	client := apis.NewAPIClient(cfg)
	var result []domain.MarketplaceUnsettledTransaction
	seen := make(map[string]bool)
	pageToken := ""
	for page := 0; page < 1000; page++ {
		requestCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		req := client.FinanceV202507API.Finance202507OrdersUnsettledGet(requestCtx).
			ShopCipher(token.ShopID).XTtsAccessToken(token.AccessToken).ContentType("application/json").
			SortField("order_create_time").SortOrder("DESC").PageSize(100)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		h.paceTikTokFinanceRequest()
		resp, _, requestErr := req.Execute()
		cancel()
		if requestErr != nil {
			return nil, fmt.Errorf("Gagal membaca transaksi belum cair TikTok: %w", requestErr)
		}
		if resp == nil || resp.GetCode() != 0 {
			if resp == nil {
				return nil, fmt.Errorf("Respons TikTok kosong")
			}
			return nil, fmt.Errorf("TikTok menolak transaksi belum cair (%d): %s", resp.GetCode(), resp.GetMessage())
		}
		if resp.Data == nil {
			return nil, fmt.Errorf("Respons TikTok tidak berisi data")
		}
		for _, item := range resp.Data.Transactions {
			id := item.GetId()
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			amount, amountErr := parseIDRAmount(item.GetEstSettlementAmount())
			if amountErr != nil {
				return nil, amountErr
			}
			var deliveryTime *time.Time
			if item.GetOrderDeliveryTime() > 0 {
				t := time.Unix(item.GetOrderDeliveryTime(), 0)
				deliveryTime = &t
			}
			result = append(result, domain.MarketplaceUnsettledTransaction{
				TransactionID: id, OrderID: item.GetOrderId(), AdjustmentID: item.GetAdjustmentId(),
				TransactionType: item.GetType(), UnsettledReason: item.GetUnsettledReason(),
				Category:        classifyUnsettled(item.GetUnsettledReason(), item.GetOrderDeliveryTime()),
				EstimatedAmount: amount, EstimatedSettlement: item.GetEstimatedSettlement(), OrderDeliveryTime: deliveryTime,
			})
		}
		next := resp.Data.GetNextPageToken()
		if next == "" {
			return result, nil
		}
		if next == pageToken {
			return nil, fmt.Errorf("Paginasi TikTok mengulang token yang sama")
		}
		pageToken = next
	}
	return nil, fmt.Errorf("Batas 1000 halaman TikTok tercapai; snapshot tidak disimpan agar tidak parsial")
}
