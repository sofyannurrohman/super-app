package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type WMSFulfillmentHandler struct {
	DB *gorm.DB
}

func NewWMSFulfillmentHandler(db *gorm.DB) *WMSFulfillmentHandler {
	return &WMSFulfillmentHandler{DB: db}
}

func (h *WMSFulfillmentHandler) RegisterRoutes(r *gin.RouterGroup) {
	wms := r.Group("/wms")
	{
		wms.GET("/picking-list", h.GetPickingList)
		wms.GET("/marketplace-orders", h.GetMarketplaceOrders)
		wms.POST("/pick/confirm", h.ConfirmPick)
		wms.POST("/pack/scan", h.ScanPack)
		wms.POST("/pack/complete", h.CompletePack)
	}
}

// 1. Picking List: Mengagregasi SKU dari order yang statusnya PENDING
func (h *WMSFulfillmentHandler) GetPickingList(c *gin.Context) {
	type PickingItem struct {
		SKU         string `json:"sku"`
		TotalQty    int    `json:"total_qty"`
		PickedQty   int    `json:"picked_qty"`
		BinLocation string `json:"bin_location"`
	}

	var results []PickingItem

	// Ambil semua item dari order PENDING dan kelompokkan berdasarkan SKU dan BinLocation
	err := h.DB.Model(&domain.OrderItem{}).
		Select("order_items.sku, order_items.bin_location, sum(order_items.qty) as total_qty, sum(order_items.picked_qty) as picked_qty").
		Joins("join orders on orders.id = order_items.order_id").
		Where("orders.status = ?", "PENDING_PICK").
		Group("order_items.sku, order_items.bin_location").
		Order("order_items.bin_location asc"). // Urutkan agar rute picker efisien
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch picking list", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

func (h *WMSFulfillmentHandler) GetMarketplaceOrders(c *gin.Context) {
	status := c.Query("status")
	limitStr := c.Query("page_size")

	query := h.DB.Model(&domain.MarketplaceOrder{})

	if status != "" {
		statuses := strings.Split(status, ",")
		query = query.Where("order_status IN ?", statuses)
	}

	limit := 1000
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get distinct order_ids ordered by max(created_at) desc
	var orderIDs []string
	if err := query.Select("order_id").Group("order_id").Order("MAX(created_at) desc").Limit(limit).Pluck("order_id", &orderIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch marketplace orders", "details": err.Error()})
		return
	}

	if len(orderIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []domain.MarketplaceOrderGroup{}})
		return
	}

	// Fetch all line items for these order_ids
	var lineItems []domain.MarketplaceOrder
	if err := h.DB.Where("order_id IN ?", orderIDs).Order("created_at desc").Find(&lineItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch line items", "details": err.Error()})
		return
	}

	// Group line items by order_id
	groupsMap := make(map[string]*domain.MarketplaceOrderGroup)
	for _, item := range lineItems {
		if _, exists := groupsMap[item.OrderID]; !exists {
			groupsMap[item.OrderID] = &domain.MarketplaceOrderGroup{
				OrderID:          item.OrderID,
				OrderStatus:      item.OrderStatus,
				WarehouseName:    item.WarehouseName,
				TrackingID:       item.TrackingID,
				Marketplace:      item.Marketplace,
				BuyerUsername:    item.BuyerUsername,
				PaymentMethod:    item.PaymentMethod,
				TotalAmount:      item.TotalAmount,
				Currency:         item.Currency,
				ShippingProvider: item.ShippingProvider,
				PickedBy:         item.PickedBy,
				PackedBy:         item.PackedBy,
				CreatedAt:        item.CreatedAt,
				Items:            []domain.MarketplaceOrder{},
			}
		} else {
			if groupsMap[item.OrderID].TrackingID == "" && item.TrackingID != "" {
				groupsMap[item.OrderID].TrackingID = item.TrackingID
			}
			if groupsMap[item.OrderID].ShippingProvider == "" && item.ShippingProvider != "" {
				groupsMap[item.OrderID].ShippingProvider = item.ShippingProvider
			}
			if groupsMap[item.OrderID].WarehouseName == "" && item.WarehouseName != "" {
				groupsMap[item.OrderID].WarehouseName = item.WarehouseName
			}
		}
		groupsMap[item.OrderID].Items = append(groupsMap[item.OrderID].Items, item)
	}

	// Reorder groups based on the original orderIDs order
	var groups []domain.MarketplaceOrderGroup
	for _, id := range orderIDs {
		if g, exists := groupsMap[id]; exists {
			groups = append(groups, *g)
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": groups})
}

// 2. Confirm Pick: Menandai bahwa barang sudah diambil dari rak
type ConfirmPickPayload struct {
	SKU         string `json:"sku" binding:"required"`
	QtyPicked   int    `json:"qty_picked" binding:"required,gt=0"`
	BinLocation string `json:"bin_location"`
}

func (h *WMSFulfillmentHandler) ConfirmPick(c *gin.Context) {
	var payload ConfirmPickPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Dalam skenario sebenarnya, picker mungkin mengambil 15 pcs dari satu SKU
	// yang mencakup 3 order berbeda. Kita distribusikan QtyPicked ke order_items terkait.
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var items []domain.OrderItem
		// Ambil semua item PENDING untuk SKU ini
		if err := tx.Joins("join orders on orders.id = order_items.order_id").
			Where("orders.status = ?", "PENDING").
			Where("order_items.sku = ? AND (order_items.bin_location = ? OR order_items.bin_location IS NULL OR order_items.bin_location = '')", payload.SKU, payload.BinLocation).
			Order("orders.created_at asc"). // Fulfill order terlama dulu (FIFO)
			Find(&items).Error; err != nil {
			return err
		}

		remainingToPick := payload.QtyPicked
		for i := range items {
			if remainingToPick <= 0 {
				break
			}
			needed := items[i].Qty - items[i].PickedQty
			if needed > 0 {
				pickAmount := needed
				if remainingToPick < needed {
					pickAmount = remainingToPick
				}
				items[i].PickedQty += pickAmount
				remainingToPick -= pickAmount

				if err := tx.Save(&items[i]).Error; err != nil {
					return err
				}

				// Ubah status order menjadi PICKING jika belum
				tx.Model(&domain.Order{}).Where("id = ?", items[i].OrderID).Update("status", "PICKING")
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm picking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Picking confirmed successfully"})
}

// 3. Scan Pack: Saat packer men-scan barcode barang untuk order tertentu
type ScanPackPayload struct {
	OrderNumber string `json:"order_number" binding:"required"`
	SKU         string `json:"sku" binding:"required"` // Didapat dari barcode scanner USB Putian U2B
}

func (h *WMSFulfillmentHandler) ScanPack(c *gin.Context) {
	var payload ScanPackPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order domain.Order
	if err := h.DB.Preload("OrderItems").Where("order_number = ?", payload.OrderNumber).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Cari item yang sesuai
	itemFound := false
	for i, item := range order.OrderItems {
		if item.SKU == payload.SKU {
			itemFound = true
			if item.PackedQty >= item.Qty {
				c.JSON(http.StatusBadRequest, gin.H{"error": "This SKU is already fully packed for this order"})
				return
			}

			// Tambah counter packing
			order.OrderItems[i].PackedQty += 1
			h.DB.Save(&order.OrderItems[i])

			// Update status order ke PACKING
			if order.Status != "PACKING" {
				h.DB.Model(&order).Update("status", "PACKING")
			}

			c.JSON(http.StatusOK, gin.H{
				"message":    "Scan success",
				"packed_qty": order.OrderItems[i].PackedQty,
				"target_qty": item.Qty,
			})
			return
		}
	}

	if !itemFound {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU not found in this order! Please check the item."}) // Akan mentrigger suara Buzzer Merah di frontend
	}
}

// 4. Complete Pack: Semua barang sudah dipacking
type CompletePackPayload struct {
	OrderNumber string `json:"order_number" binding:"required"`
}

func (h *WMSFulfillmentHandler) CompletePack(c *gin.Context) {
	var payload CompletePackPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order domain.Order
	if err := h.DB.Preload("OrderItems").Where("order_number = ?", payload.OrderNumber).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Validasi apakah semuanya sudah ter-pack
	for _, item := range order.OrderItems {
		if item.PackedQty < item.Qty {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot complete packing. Some items are missing."})
			return
		}
	}

	// Update order ke READY_TO_SHIP
	order.Status = "READY_TO_SHIP"
	h.DB.Save(&order)

	// Jika ini dari TIKTOK, kita bisa menambahkan logika integrasi "Ship Package" dan tarik Resi
	if order.Source == "TIKTOK" {
		// Logika call HTTP ke TikTok API ada di sini
		// ...
	}

	c.JSON(http.StatusOK, gin.H{"message": "Packing completed successfully", "status": order.Status})
}
