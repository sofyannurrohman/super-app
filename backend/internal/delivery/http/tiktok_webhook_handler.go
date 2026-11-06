package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type TikTokWebhookHandler struct {
	DB           *gorm.DB
	TikTokSecret string
}

func NewTikTokWebhookHandler(db *gorm.DB) *TikTokWebhookHandler {
	return &TikTokWebhookHandler{
		DB:           db,
		TikTokSecret: os.Getenv("TIKTOK_APP_SECRET"),
	}
}

func (h *TikTokWebhookHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/webhooks/tiktok", h.HandleWebhook)
}

func (h *TikTokWebhookHandler) verifySignature(signature string, body []byte) bool {
	if h.TikTokSecret == "" {
		return false // Tolak jika secret belum diset
	}
	mac := hmac.New(sha256.New, []byte(h.TikTokSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

type TikTokWebhookPayload struct {
	Type           int    `json:"type"` // 1 = ORDER_STATUS_CHANGED
	ShopID         string `json:"shop_id"`
	Timestamp      int64  `json:"timestamp"`
	NotificationID string `json:"notification_id"`
	Data           struct {
		OrderID     string `json:"order_id"`
		OrderStatus string `json:"order_status"` // e.g., AWAITING_SHIPMENT
	} `json:"data"`
}

func (h *TikTokWebhookHandler) HandleWebhook(c *gin.Context) {
	signature := c.GetHeader("x-tts-webhook-signature")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature header"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Validasi HMAC
	if !h.verifySignature(signature, bodyBytes) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Parsing payload
	var payload TikTokWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Idempotency Check: Pastikan webhook dengan NotificationID ini belum diproses
	var existingLog domain.WebhookLog
	if err := h.DB.Where("notification_id = ?", payload.NotificationID).First(&existingLog).Error; err == nil {
		// Log sudah ada, abaikan pesan ini namun balas dengan 200 OK agar TikTok tidak retrying
		c.JSON(http.StatusOK, gin.H{"message": "Already processed"})
		return
	}

	// Simpan ke tabel webhook_logs
	logEntry := domain.WebhookLog{
		NotificationID: payload.NotificationID,
		Payload:        string(bodyBytes),
	}
	h.DB.Create(&logEntry)

	// Jika ini adalah event ORDER_STATUS_CHANGED (biasanya type 1 untuk TikTok API v2)
	// dan status pesanan adalah AWAITING_SHIPMENT, maka masukkan ke antrean WMS
	if payload.Type == 1 && payload.Data.OrderStatus == "AWAITING_SHIPMENT" {

		// 1. Simpan Header Order
		newOrder := domain.Order{
			OrderNumber:     payload.Data.OrderID,
			Source:          "TIKTOK",
			ExternalOrderID: payload.Data.OrderID,
			Status:          "PENDING",
		}

		// Untuk menghindari duplikat order_number karena mungkin ada perubahan status lain yang juga men-trigger
		if err := h.DB.Where("order_number = ?", newOrder.OrderNumber).FirstOrCreate(&newOrder).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save order"})
			return
		}

		// 2. Dalam implementasi nyata, pada titik ini kita seharusnya melakukan HTTP GET ke TikTok
		// menggunakan endpoint `/api/v2/order/detail` dengan AccessToken untuk mendapatkan daftar SKU dan QTY,
		// lalu memasukkannya ke tabel `order_items`.

		// (Implementasi get_order_detail ditunda ke fase selanjutnya saat sinkronisasi penuh dibutuhkan)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook received successfully"})
}
