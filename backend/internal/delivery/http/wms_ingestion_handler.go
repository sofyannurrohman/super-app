package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type WMSIngestionHandler struct {
	DB *gorm.DB
}

func NewWMSIngestionHandler(db *gorm.DB) *WMSIngestionHandler {
	return &WMSIngestionHandler{DB: db}
}

func (h *WMSIngestionHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Endpoint ini mungkin diproteksi oleh middleware JWT Admin/Mobile di dalam main.go
	r.POST("/internal/orders", h.CreateInternalOrder)
}

type InternalOrderPayload struct {
	OrderNumber string `json:"order_number" binding:"required"`
	Items       []struct {
		SKU         string `json:"sku" binding:"required"`
		Qty         int    `json:"qty" binding:"required,gt=0"`
		BinLocation string `json:"bin_location"`
	} `json:"items" binding:"required,dive"`
}

func (h *WMSIngestionHandler) CreateInternalOrder(c *gin.Context) {
	var payload InternalOrderPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Gunakan transaksi untuk menjamin konsistensi
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		order := domain.Order{
			OrderNumber: payload.OrderNumber,
			Source:      "SUPER_APP",
			Status:      "PENDING",
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		var items []domain.OrderItem
		for _, itemPayload := range payload.Items {
			items = append(items, domain.OrderItem{
				OrderID:     order.ID,
				SKU:         itemPayload.SKU,
				Qty:         itemPayload.Qty,
				BinLocation: itemPayload.BinLocation,
			})
		}

		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order in WMS"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order successfully ingested to WMS"})
}
