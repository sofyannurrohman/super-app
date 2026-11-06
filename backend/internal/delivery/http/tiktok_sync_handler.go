package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"

	"tiktokshop/open/sdk_golang/apis"
	order_v202309 "tiktokshop/open/sdk_golang/models/order/v202309"
)

type TikTokSyncHandler struct {
	DB        *gorm.DB
	AppKey    string
	AppSecret string
}

func NewTikTokSyncHandler(db *gorm.DB) *TikTokSyncHandler {
	return &TikTokSyncHandler{
		DB:        db,
		AppKey:    os.Getenv("TIKTOK_APP_KEY"),
		AppSecret: os.Getenv("TIKTOK_APP_SECRET"),
	}
}

func (h *TikTokSyncHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/wms/tiktok/sync", h.SyncOrders)
}

func (h *TikTokSyncHandler) SyncOrders(c *gin.Context) {
	branchID := parseBranchID(c)
	var branchIDPtr *uuid.UUID
	if branchID != uuid.Nil {
		branchIDPtr = &branchID
	}

	// 1. Get valid TikTok access token from DB (auto-refresh if expired)
	tokenRecord, err := getValidTikTokToken(h.DB, h.AppKey, h.AppSecret, branchIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Setup SDK Configuration
	cfg := apis.NewConfiguration()
	cfg.AddAppInfo(h.AppKey, h.AppSecret)
	client := apis.NewAPIClient(cfg)

	// 3. Fetch only IN_TRANSIT ("Dalam Pengiriman") orders updated in the last 7 days
	updateTimeGe := time.Now().AddDate(0, 0, -7).Unix()
	statusInTransit := "IN_TRANSIT"

	var allOrders []order_v202309.Order202309GetOrderListResponseDataOrders
	var pageCursor string

	maxPages := 100
	pageCount := 0

	for {
		pageCount++
		if pageCount > maxPages {
			break
		}

		reqBody := order_v202309.Order202309GetOrderListRequestBody{
			OrderStatus:  &statusInTransit,
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch orders from TikTok: %v, Body: %s", err, bodyStr)})
			return
		}

		if resp != nil && resp.Data != nil && resp.Data.Orders != nil {
			allOrders = append(allOrders, resp.Data.Orders...)
		}

		if resp == nil || resp.Data == nil || resp.Data.GetNextPageToken() == "" {
			break
		}
		pageCursor = resp.Data.GetNextPageToken()
	}

	if len(allOrders) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No new orders to sync", "synced_count": 0})
		return
	}

	syncedCount := 0

	// 4. Ingest into WMS Database
	for _, tkOrder := range allOrders {
		orderId := tkOrder.GetId()
		if orderId == "" {
			continue
		}

		// Check if order exists
		var count int64
		h.DB.Model(&domain.Order{}).Where("order_number = ?", orderId).Count(&count)
		if count > 0 {
			continue // Already synced
		}

		// Extract payment info
		var totalAmount float64
		currency := "IDR"
		if tkOrder.HasPayment() {
			payment := tkOrder.GetPayment()
			totalAmountStr := payment.GetTotalAmount()
			totalAmount, _ = strconv.ParseFloat(totalAmountStr, 64)
			currency = payment.GetCurrency()
		}
		paymentMethod := tkOrder.GetPaymentMethodName()

		// Create new WMS order
		newOrder := domain.Order{
			OrderNumber: orderId,
			Source:      "TIKTOK",
			Status:      "PENDING_PICK", // Ready for picker
		}

		err = h.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&newOrder).Error; err != nil {
				return err
			}

			// Map order items
			if tkOrder.LineItems != nil {
				itemCounts := make(map[string]int)
				firstItems := make(map[string]order_v202309.Order202309GetOrderListResponseDataOrdersLineItems)

				for _, tkItem := range tkOrder.LineItems {
					skuID := tkItem.GetSkuId()
					if skuID == "" {
						skuID = tkItem.GetProductId()
					}
					itemCounts[skuID]++
					if _, exists := firstItems[skuID]; !exists {
						firstItems[skuID] = tkItem
					}
				}

				for skuID, count := range itemCounts {
					tkItem := firstItems[skuID]
					item := domain.OrderItem{
						OrderID:     newOrder.ID,
						SKU:         tkItem.GetSellerSku(),
						Qty:         count,
						PickedQty:   0,
						PackedQty:   0,
						BinLocation: "Rak Umum",
					}
					if err := tx.Create(&item).Error; err != nil {
						return err
					}

					orderStatus := tkOrder.GetStatus()
					if orderStatus == "AWAITING_COLLECTION" {
						orderStatus = "VERIFIED"
					}

					skuID := tkItem.GetSkuId()
					if skuID == "" {
						skuID = tkItem.GetProductId()
					}

					// Add to marketplace_orders table
					mpOrder := domain.MarketplaceOrder{
						BranchID:         tokenRecord.BranchID,
						OrderID:          orderId,
						OrderStatus:      orderStatus,
						SkuID:            skuID,
						SellerSKU:        tkItem.GetSellerSku(),
						ProductName:      tkItem.GetProductName(),
						Marketplace:      "tiktok",
						BuyerUsername:    tkOrder.GetBuyerNickname(),
						TotalAmount:      totalAmount,
						Currency:         currency,
						PaymentMethod:    paymentMethod,
						TrackingID:       tkOrder.GetTrackingNumber(),
						ShippingProvider: tkOrder.GetShippingProvider(),
						CreateTimeRaw:    tkOrder.GetCreateTime(),
						ItemCount:        count,
					}

					if err := tx.Create(&mpOrder).Error; err != nil {
						return err
					}
				}
			}
			return nil
		})

		if err == nil {
			syncedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Sync completed",
		"synced_count": syncedCount,
	})
}
