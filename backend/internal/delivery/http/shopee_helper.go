package http

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

const (
	ShopeeAPIBaseURL  = "https://partner.shopeemobile.com"
	ShopeeProdBaseURL = "https://partner.shopeemobile.com"
	ShopeeTestBaseURL = "https://partner.test-stable.shopeemobile.com"
)

type ShopeeClient struct {
	DB          *gorm.DB
	PartnerID   string
	PartnerKey  string
	Environment string // "production" or "sandbox"
	BaseURL     string
	HTTPClient  *http.Client
}

func NewShopeeClient(db *gorm.DB) *ShopeeClient {
	partnerID := strings.TrimSpace(os.Getenv("SHOPEE_PARTNER_ID"))
	partnerKey := strings.TrimSpace(os.Getenv("SHOPEE_PARTNER_KEY"))
	env := strings.ToLower(strings.TrimSpace(os.Getenv("SHOPEE_ENV")))
	if env == "" {
		env = "production"
	}

	// Try loading from database marketplace_configs table first
	if db != nil {
		var cfg domain.MarketplaceConfig
		if err := db.Where("marketplace = ? AND is_active = true", "shopee").First(&cfg).Error; err == nil {
			if strings.TrimSpace(cfg.AppID) != "" {
				partnerID = strings.TrimSpace(cfg.AppID)
			}
			if strings.TrimSpace(cfg.AppSecret) != "" {
				partnerKey = strings.TrimSpace(cfg.AppSecret)
			}
			if strings.TrimSpace(cfg.Environment) != "" {
				env = strings.ToLower(strings.TrimSpace(cfg.Environment))
			}
		}
	}

	baseURL := ShopeeProdBaseURL
	if env == "sandbox" || env == "test" {
		baseURL = ShopeeTestBaseURL
	}

	return &ShopeeClient{
		DB:          db,
		PartnerID:   partnerID,
		PartnerKey:  partnerKey,
		Environment: env,
		BaseURL:     baseURL,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GeneratePublicSign generates sign for public/auth APIs without shop_id or access_token
func (s *ShopeeClient) GeneratePublicSign(path string, timestamp int64) string {
	signBase := fmt.Sprintf("%s%s%d", s.PartnerID, path, timestamp)
	mac := hmac.New(sha256.New, []byte(s.PartnerKey))
	mac.Write([]byte(signBase))
	return hex.EncodeToString(mac.Sum(nil))
}

// GenerateShopSign generates sign for shop APIs with access_token and shop_id
func (s *ShopeeClient) GenerateShopSign(path string, timestamp int64, accessToken string, shopID string) string {
	signBase := fmt.Sprintf("%s%s%d%s%s", s.PartnerID, path, timestamp, accessToken, shopID)
	mac := hmac.New(sha256.New, []byte(s.PartnerKey))
	mac.Write([]byte(signBase))
	return hex.EncodeToString(mac.Sum(nil))
}

func maskKey(k string) string {
	if len(k) <= 4 {
		return "****"
	}
	return "****" + k[len(k)-4:]
}

// BuildAuthURL builds the Shopee shop authorization URL
func (s *ShopeeClient) BuildAuthURL(redirectURL string) (string, error) {
	if s.PartnerID == "" || s.PartnerKey == "" {
		return "", fmt.Errorf("Shopee Partner ID dan Partner Key belum dikonfigurasi. Silakan lengkapi di Pengaturan Shopee API")
	}

	path := "/api/v2/shop/auth_partner"
	timestamp := time.Now().Unix()
	sign := s.GeneratePublicSign(path, timestamp)

	baseURL := s.BaseURL
	if baseURL == "" {
		baseURL = ShopeeAPIBaseURL
	}

	log.Printf("[SHOPEE OAUTH] Generating Auth URL | Env: %s | PartnerID: %s | Redirect: %s", s.Environment, maskKey(s.PartnerID), redirectURL)

	return fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&sign=%s&redirect=%s",
		baseURL, path, s.PartnerID, timestamp, sign, url.QueryEscape(redirectURL)), nil
}

// GetValidShopeeToken retrieves and auto-refreshes the Shopee token for a given branch using configured client
func GetValidShopeeToken(db *gorm.DB, branchID *uuid.UUID) (*domain.MarketplaceToken, error) {
	client := NewShopeeClient(db)
	return getValidShopeeToken(db, client.PartnerID, client.PartnerKey, branchID)
}

// getValidShopeeToken retrieves and auto-refreshes the Shopee token for a given branch
func getValidShopeeToken(db *gorm.DB, partnerID, partnerKey string, branchID *uuid.UUID) (*domain.MarketplaceToken, error) {
	return getValidShopeeTokenScoped(db, partnerID, partnerKey, branchID, "")
}

func getValidShopeeTokenScoped(db *gorm.DB, partnerID, partnerKey string, branchID *uuid.UUID, shopID string) (*domain.MarketplaceToken, error) {
	client := NewShopeeClient(db)
	if partnerID == "" {
		partnerID = client.PartnerID
	}
	if partnerKey == "" {
		partnerKey = client.PartnerKey
	}

	if partnerID == "" || partnerKey == "" {
		return nil, fmt.Errorf("Shopee Partner ID dan Partner Key belum dikonfigurasi. Silakan lengkapi di Pengaturan Shopee API")
	}

	var token domain.MarketplaceToken
	q := db.Where("marketplace = ?", "shopee")
	if shopID != "" {
		q = q.Where("shop_id = ?", shopID)
	}
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}

	if err := q.Order("created_at desc").First(&token).Error; err != nil {
		if branchID != nil && *branchID != uuid.Nil {
			return nil, fmt.Errorf("Akun Shopee belum terhubung untuk cabang ini. Silakan hubungkan akun di menu Integrasi Marketplace")
		}
		return nil, fmt.Errorf("Shopee token tidak ditemukan: %w", err)
	}

	// If expired or expiring in less than 15 minutes, refresh token
	if token.RefreshToken != "" && time.Now().After(token.ExpiresAt.Add(-15*time.Minute)) {
		path := "/api/v2/auth/access_token/get"
		timestamp := time.Now().Unix()

		signBase := fmt.Sprintf("%s%s%d", partnerID, path, timestamp)
		mac := hmac.New(sha256.New, []byte(partnerKey))
		mac.Write([]byte(signBase))
		sign := hex.EncodeToString(mac.Sum(nil))

		reqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&sign=%s",
			client.BaseURL, path, partnerID, timestamp, sign)

		shopIDNum, _ := strconv.ParseInt(token.ShopID, 10, 64)
		partnerIDNum, _ := strconv.ParseInt(partnerID, 10, 64)

		reqBody, _ := json.Marshal(map[string]interface{}{
			"refresh_token": token.RefreshToken,
			"shop_id":       shopIDNum,
			"partner_id":    partnerIDNum,
		})

		log.Printf("[SHOPEE AUTH] Refreshing Token | Env: %s | PartnerID: %s | ShopID: %s", client.Environment, maskKey(partnerID), token.ShopID)
		resp, err := http.Post(reqURL, "application/json", bytes.NewBuffer(reqBody))
		if err == nil {
			defer resp.Body.Close()
			var result struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				ExpireIn     int    `json:"expire_in"`
				Error        string `json:"error"`
				Message      string `json:"message"`
			}
			if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr == nil {
				if result.Error == "" && result.AccessToken != "" {
					token.AccessToken = result.AccessToken
					if result.RefreshToken != "" {
						token.RefreshToken = result.RefreshToken
					}
					if result.ExpireIn > 0 {
						token.ExpiresAt = time.Now().Add(time.Duration(result.ExpireIn) * time.Second)
					} else {
						token.ExpiresAt = time.Now().Add(4 * time.Hour)
					}
					_ = db.Save(&token)

					// Sync with other branches sharing same shop_id
					_ = db.Model(&domain.MarketplaceToken{}).
						Where("marketplace = ? AND shop_id = ?", "shopee", token.ShopID).
						Updates(map[string]interface{}{
							"access_token":  token.AccessToken,
							"refresh_token": token.RefreshToken,
							"expires_at":    token.ExpiresAt,
						}).Error
					log.Printf("[SHOPEE AUTH] Token refreshed successfully | ShopID: %s | ExpiresIn: %ds", token.ShopID, result.ExpireIn)
				} else {
					log.Printf("[SHOPEE AUTH] Token refresh failed | ShopID: %s | Error: %s (%s)", token.ShopID, result.Error, result.Message)
				}
			}
		} else {
			log.Printf("[SHOPEE AUTH] Token refresh HTTP error: %v", err)
		}
	}

	return &token, nil
}

// Shopee API structs
type ShopeeItemListResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		Item []struct {
			ItemID     int64  `json:"item_id"`
			ItemStatus string `json:"item_status"`
			UpdateTime int64  `json:"update_time"`
		} `json:"item"`
		TotalCount  int  `json:"total_count"`
		HasNextPage bool `json:"has_next_page"`
		NextOffset  int  `json:"next_offset"`
	} `json:"response"`
}

type ShopeeItemBaseInfoResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ItemList []ShopeeItemDetail `json:"item_list"`
	} `json:"response"`
}

type ShopeeItemDetail struct {
	ItemID      int64  `json:"item_id"`
	ItemName    string `json:"item_name"`
	ItemSKU     string `json:"item_sku"`
	ItemStatus  string `json:"item_status"`
	CategoryID  int64  `json:"category_id"`
	Description string `json:"description"`
	PriceInfo   []struct {
		Currency      string  `json:"currency"`
		OriginalPrice float64 `json:"original_price"`
		CurrentPrice  float64 `json:"current_price"`
	} `json:"price_info"`
	StockInfoV2 struct {
		SummaryInfo struct {
			TotalAvailableStock int `json:"total_available_stock"`
		} `json:"summary_info"`
	} `json:"stock_info_v2"`
	HasModel bool `json:"has_model"`
}

type ShopeeModelListResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		Model []ShopeeModelDetail `json:"model"`
	} `json:"response"`
}

type ShopeeModelDetail struct {
	ModelID   int64  `json:"model_id"`
	ModelSKU  string `json:"model_sku"`
	ModelName string `json:"model_name"`
	PriceInfo []struct {
		Currency     string  `json:"currency"`
		CurrentPrice float64 `json:"current_price"`
	} `json:"price_info"`
	StockInfoV2 struct {
		SummaryInfo struct {
			TotalAvailableStock int `json:"total_available_stock"`
		} `json:"summary_info"`
	} `json:"stock_info_v2"`
}

type ShopeeOrderListResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		More       bool   `json:"more"`
		NextCursor string `json:"next_cursor"`
		OrderList  []struct {
			OrderSN     string `json:"order_sn"`
			OrderStatus string `json:"order_status"`
		} `json:"order_list"`
	} `json:"response"`
}

type ShopeeOrderDetailResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		OrderList []ShopeeOrderDetail `json:"order_list"`
	} `json:"response"`
}

type ShopeeOrderDetail struct {
	OrderSN         string  `json:"order_sn"`
	OrderStatus     string  `json:"order_status"`
	BuyerUsername   string  `json:"buyer_username"`
	CreateTime      int64   `json:"create_time"`
	UpdateTime      int64   `json:"update_time"`
	TotalAmount     float64 `json:"total_amount"`
	Currency        string  `json:"currency"`
	PaymentMethod   string  `json:"payment_method"`
	ShippingCarrier string  `json:"shipping_carrier"`
	TrackingNumber  string  `json:"tracking_number"`
	PackageList     []struct {
		PackageNumber   string `json:"package_number"`
		LogisticsStatus string `json:"logistics_status"`
		ShippingCarrier string `json:"shipping_carrier"`
		TrackingNumber  string `json:"tracking_number"`
	} `json:"package_list"`
	ItemList []struct {
		ItemID                 int64   `json:"item_id"`
		ItemName               string  `json:"item_name"`
		ItemSKU                string  `json:"item_sku"`
		ModelID                int64   `json:"model_id"`
		ModelName              string  `json:"model_name"`
		ModelSKU               string  `json:"model_sku"`
		ModelQuantityPurchased int     `json:"model_quantity_purchased"`
		ModelOriginalPrice     float64 `json:"model_original_price"`
		ModelDiscountedPrice   float64 `json:"model_discounted_price"`
	} `json:"item_list"`
}

type ShopeeTrackingResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		TrackingNumber string `json:"tracking_number"`
	} `json:"response"`
}

// FetchShopeeProducts fetches all products and variations from Shopee API
func (s *ShopeeClient) FetchShopeeProducts(token *domain.MarketplaceToken) ([]domain.MarketplaceProduct, error) {
	var results []domain.MarketplaceProduct

	offset := 0
	pageSize := 50

	for {
		path := "/api/v2/product/get_item_list"
		timestamp := time.Now().Unix()
		sign := s.GenerateShopSign(path, timestamp, token.AccessToken, token.ShopID)

		reqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&item_status=NORMAL&offset=%d&page_size=%d",
			s.BaseURL, path, s.PartnerID, timestamp, token.AccessToken, token.ShopID, sign, offset, pageSize)

		resp, err := s.HTTPClient.Get(reqURL)
		if err != nil {
			return results, fmt.Errorf("failed to get item list: %w", err)
		}

		var itemListResp ShopeeItemListResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&itemListResp)
		resp.Body.Close()
		if decodeErr != nil {
			return results, fmt.Errorf("failed to decode item list: %w", decodeErr)
		}

		if itemListResp.Error != "" {
			return results, fmt.Errorf("shopee api error: %s - %s", itemListResp.Error, itemListResp.Message)
		}

		if len(itemListResp.Response.Item) == 0 {
			break
		}

		// Batch item_id_list for get_item_base_info (max 50)
		var itemIDs []string
		for _, item := range itemListResp.Response.Item {
			itemIDs = append(itemIDs, strconv.FormatInt(item.ItemID, 10))
		}

		if len(itemIDs) > 0 {
			baseInfoPath := "/api/v2/product/get_item_base_info"
			baseTimestamp := time.Now().Unix()
			baseSign := s.GenerateShopSign(baseInfoPath, baseTimestamp, token.AccessToken, token.ShopID)
			baseReqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&item_id_list=%s",
				s.BaseURL, baseInfoPath, s.PartnerID, baseTimestamp, token.AccessToken, token.ShopID, baseSign, strings.Join(itemIDs, ","))

			baseResp, err := s.HTTPClient.Get(baseReqURL)
			if err == nil {
				var baseInfo ShopeeItemBaseInfoResponse
				_ = json.NewDecoder(baseResp.Body).Decode(&baseInfo)
				baseResp.Body.Close()

				for _, itemDetail := range baseInfo.Response.ItemList {
					itemIDStr := strconv.FormatInt(itemDetail.ItemID, 10)

					if itemDetail.HasModel {
						// Fetch models (variations)
						modelPath := "/api/v2/product/get_model_list"
						mTimestamp := time.Now().Unix()
						mSign := s.GenerateShopSign(modelPath, mTimestamp, token.AccessToken, token.ShopID)
						mReqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&item_id=%d",
							s.BaseURL, modelPath, s.PartnerID, mTimestamp, token.AccessToken, token.ShopID, mSign, itemDetail.ItemID)

						mResp, mErr := s.HTTPClient.Get(mReqURL)
						if mErr == nil {
							var modelList ShopeeModelListResponse
							_ = json.NewDecoder(mResp.Body).Decode(&modelList)
							mResp.Body.Close()

							for _, model := range modelList.Response.Model {
								modelIDStr := strconv.FormatInt(model.ModelID, 10)
								sellerSKU := model.ModelSKU
								if sellerSKU == "" {
									sellerSKU = itemDetail.ItemSKU
								}
								price := 0.0
								if len(model.PriceInfo) > 0 {
									price = model.PriceInfo[0].CurrentPrice
								}
								stock := model.StockInfoV2.SummaryInfo.TotalAvailableStock

								results = append(results, domain.MarketplaceProduct{
									BranchID:      token.BranchID,
									Marketplace:   "shopee",
									ProductID:     itemIDStr,
									ProductName:   fmt.Sprintf("%s - %s", itemDetail.ItemName, model.ModelName),
									SkuID:         modelIDStr,
									SellerSKU:     sellerSKU,
									PriceAmount:   price,
									PriceCurrency: "IDR",
									Stock:         stock,
									Status:        itemDetail.ItemStatus,
									SalesRegions:  "ID",
									LastSyncedAt:  time.Now(),
								})
							}
						}
					} else {
						// Single item without model
						price := 0.0
						if len(itemDetail.PriceInfo) > 0 {
							price = itemDetail.PriceInfo[0].CurrentPrice
						}
						stock := itemDetail.StockInfoV2.SummaryInfo.TotalAvailableStock

						results = append(results, domain.MarketplaceProduct{
							BranchID:      token.BranchID,
							Marketplace:   "shopee",
							ProductID:     itemIDStr,
							ProductName:   itemDetail.ItemName,
							SkuID:         itemIDStr,
							SellerSKU:     itemDetail.ItemSKU,
							PriceAmount:   price,
							PriceCurrency: "IDR",
							Stock:         stock,
							Status:        itemDetail.ItemStatus,
							SalesRegions:  "ID",
							LastSyncedAt:  time.Now(),
						})
					}
				}
			}
		}

		if !itemListResp.Response.HasNextPage {
			break
		}
		offset = itemListResp.Response.NextOffset
	}

	return results, nil
}

// FetchShopeeOrders fetches orders from Shopee API focusing on "PROCESSED" (Telah diproses / Siap Dikirim) status
func (s *ShopeeClient) FetchShopeeOrders(token *domain.MarketplaceToken, days int) ([]ShopeeOrderDetail, error) {
	var results []ShopeeOrderDetail
	seenOrderSN := make(map[string]bool)

	if days <= 0 {
		days = 15
	}

	// Shopee API restricts (time_to - time_from) <= 15 days.
	now := time.Now()
	totalSeconds := int64(days * 24 * 3600)
	windowSize := int64(14 * 24 * 3600) // 14 days per chunk to stay safely within 15-day limit

	// Focus specifically on "PROCESSED" (Telah diproses) orders as requested to reduce scope
	targetStatuses := []string{"PROCESSED"}

	for _, targetStatus := range targetStatuses {
		for offsetSec := int64(0); offsetSec < totalSeconds; offsetSec += windowSize {
			timeTo := now.Unix() - offsetSec
			timeFrom := timeTo - windowSize
			if timeFrom < (now.Unix() - totalSeconds) {
				timeFrom = now.Unix() - totalSeconds
			}

			cursor := ""
			for {
				path := "/api/v2/order/get_order_list"
				timestamp := time.Now().Unix()
				sign := s.GenerateShopSign(path, timestamp, token.AccessToken, token.ShopID)

				reqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&time_range_field=create_time&time_from=%d&time_to=%d&page_size=50&order_status=%s",
					s.BaseURL, path, s.PartnerID, timestamp, token.AccessToken, token.ShopID, sign, timeFrom, timeTo, targetStatus)

				if cursor != "" {
					reqURL += "&cursor=" + url.QueryEscape(cursor)
				}

				resp, err := s.HTTPClient.Get(reqURL)
				if err != nil {
					return results, fmt.Errorf("failed to get order list: %w", err)
				}

				var orderListResp ShopeeOrderListResponse
				decodeErr := json.NewDecoder(resp.Body).Decode(&orderListResp)
				resp.Body.Close()
				if decodeErr != nil {
					return results, fmt.Errorf("failed to decode order list: %w", decodeErr)
				}

				if orderListResp.Error != "" {
					log.Printf("[SHOPEE ORDERS] Warning on get_order_list (status: %s): %s - %s", targetStatus, orderListResp.Error, orderListResp.Message)
					break
				}

				if len(orderListResp.Response.OrderList) == 0 {
					break
				}

				// Batch order_sn_list (max 50)
				var orderSNs []string
				for _, o := range orderListResp.Response.OrderList {
					if !seenOrderSN[o.OrderSN] && o.OrderSN != "" {
						seenOrderSN[o.OrderSN] = true
						orderSNs = append(orderSNs, o.OrderSN)
					}
				}

				if len(orderSNs) > 0 {
					detailPath := "/api/v2/order/get_order_detail"
					dTimestamp := time.Now().Unix()
					dSign := s.GenerateShopSign(detailPath, dTimestamp, token.AccessToken, token.ShopID)
					dReqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&order_sn_list=%s&response_optional_fields=buyer_user,buyer_username,item_list,total_amount,payment_method,shipping_carrier,package_list",
						s.BaseURL, detailPath, s.PartnerID, dTimestamp, token.AccessToken, token.ShopID, dSign, strings.Join(orderSNs, ","))

					dResp, dErr := s.HTTPClient.Get(dReqURL)
					if dErr == nil {
						var detailResp ShopeeOrderDetailResponse
						_ = json.NewDecoder(dResp.Body).Decode(&detailResp)
						dResp.Body.Close()

						for _, order := range detailResp.Response.OrderList {
							// 1. Check from package_list first if present
							if len(order.PackageList) > 0 {
								for _, pkg := range order.PackageList {
									if order.TrackingNumber == "" && pkg.TrackingNumber != "" {
										order.TrackingNumber = pkg.TrackingNumber
									}
									if order.ShippingCarrier == "" && pkg.ShippingCarrier != "" {
										order.ShippingCarrier = pkg.ShippingCarrier
									}
								}
							}

							// 2. Fallback to /api/v2/logistics/get_tracking_number if tracking number is still empty
							if strings.TrimSpace(order.TrackingNumber) == "" {
								tPath := "/api/v2/logistics/get_tracking_number"
								tTimestamp := time.Now().Unix()
								tSign := s.GenerateShopSign(tPath, tTimestamp, token.AccessToken, token.ShopID)
								tReqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&order_sn=%s",
									s.BaseURL, tPath, s.PartnerID, tTimestamp, token.AccessToken, token.ShopID, tSign, order.OrderSN)

								tResp, tErr := s.HTTPClient.Get(tReqURL)
								if tErr == nil {
									var tData ShopeeTrackingResponse
									_ = json.NewDecoder(tResp.Body).Decode(&tData)
									tResp.Body.Close()
									if tData.Response.TrackingNumber != "" {
										order.TrackingNumber = tData.Response.TrackingNumber
									}
								}
							}

							// 3. Fallback to /api/v2/logistics/get_tracking_info
							if strings.TrimSpace(order.TrackingNumber) == "" {
								tPath := "/api/v2/logistics/get_tracking_info"
								tTimestamp := time.Now().Unix()
								tSign := s.GenerateShopSign(tPath, tTimestamp, token.AccessToken, token.ShopID)
								tReqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&order_sn=%s",
									s.BaseURL, tPath, s.PartnerID, tTimestamp, token.AccessToken, token.ShopID, tSign, order.OrderSN)

								tResp, tErr := s.HTTPClient.Get(tReqURL)
								if tErr == nil {
									var tData ShopeeTrackingResponse
									_ = json.NewDecoder(tResp.Body).Decode(&tData)
									tResp.Body.Close()
									if tData.Response.TrackingNumber != "" {
										order.TrackingNumber = tData.Response.TrackingNumber
									}
								}
							}

							results = append(results, order)
						}
					}
				}

				if !orderListResp.Response.More || orderListResp.Response.NextCursor == "" {
					break
				}
				cursor = orderListResp.Response.NextCursor
			}
		}
	}

	return results, nil
}

func (s *ShopeeClient) fetchShopeeOrdersWithStatus(token *domain.MarketplaceToken, timeFrom int64, timeTo int64) ([]ShopeeOrderDetail, error) {
	var results []ShopeeOrderDetail
	statuses := []string{"READY_TO_SHIP", "PROCESSED", "SHIPPED", "COMPLETED", "IN_CANCEL", "CANCELLED", "TO_RETURN"}

	for _, st := range statuses {
		cursor := ""
		for {
			path := "/api/v2/order/get_order_list"
			timestamp := time.Now().Unix()
			sign := s.GenerateShopSign(path, timestamp, token.AccessToken, token.ShopID)

			reqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&time_range_field=update_time&time_from=%d&time_to=%d&page_size=50&order_status=%s",
				s.BaseURL, path, s.PartnerID, timestamp, token.AccessToken, token.ShopID, sign, timeFrom, timeTo, st)

			if cursor != "" {
				reqURL += "&cursor=" + url.QueryEscape(cursor)
			}

			resp, err := s.HTTPClient.Get(reqURL)
			if err != nil {
				break
			}

			var orderListResp ShopeeOrderListResponse
			decodeErr := json.NewDecoder(resp.Body).Decode(&orderListResp)
			resp.Body.Close()
			if decodeErr != nil || orderListResp.Error != "" || len(orderListResp.Response.OrderList) == 0 {
				break
			}

			var orderSNs []string
			for _, o := range orderListResp.Response.OrderList {
				orderSNs = append(orderSNs, o.OrderSN)
			}

			if len(orderSNs) > 0 {
				detailPath := "/api/v2/order/get_order_detail"
				dTimestamp := time.Now().Unix()
				dSign := s.GenerateShopSign(detailPath, dTimestamp, token.AccessToken, token.ShopID)
				dReqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&order_sn_list=%s&response_optional_fields=buyer_user,buyer_username,item_list,total_amount,payment_method,shipping_carrier",
					s.BaseURL, detailPath, s.PartnerID, dTimestamp, token.AccessToken, token.ShopID, dSign, strings.Join(orderSNs, ","))

				dResp, dErr := s.HTTPClient.Get(dReqURL)
				if dErr == nil {
					var detailResp ShopeeOrderDetailResponse
					_ = json.NewDecoder(dResp.Body).Decode(&detailResp)
					dResp.Body.Close()
					results = append(results, detailResp.Response.OrderList...)
				}
			}

			if !orderListResp.Response.More || orderListResp.Response.NextCursor == "" {
				break
			}
			cursor = orderListResp.Response.NextCursor
		}
	}
	return results, nil
}

// FetchShopeeOrderDetail fetches a single order's details by order_sn directly from Shopee API
func (s *ShopeeClient) FetchShopeeOrderDetail(token *domain.MarketplaceToken, orderSN string) (*ShopeeOrderDetail, error) {
	detailPath := "/api/v2/order/get_order_detail"
	dTimestamp := time.Now().Unix()
	dSign := s.GenerateShopSign(detailPath, dTimestamp, token.AccessToken, token.ShopID)
	dReqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&order_sn_list=%s&response_optional_fields=buyer_user,buyer_username,item_list,total_amount,payment_method,shipping_carrier,package_list",
		s.BaseURL, detailPath, s.PartnerID, dTimestamp, token.AccessToken, token.ShopID, dSign, url.QueryEscape(orderSN))

	dResp, dErr := s.HTTPClient.Get(dReqURL)
	if dErr != nil {
		return nil, fmt.Errorf("gagal memanggil API Shopee get_order_detail: %w", dErr)
	}
	defer dResp.Body.Close()

	var detailResp ShopeeOrderDetailResponse
	if err := json.NewDecoder(dResp.Body).Decode(&detailResp); err != nil {
		return nil, fmt.Errorf("gagal decode response get_order_detail: %w", err)
	}

	if detailResp.Error != "" {
		return nil, fmt.Errorf("shopee api error: %s - %s", detailResp.Error, detailResp.Message)
	}

	if len(detailResp.Response.OrderList) == 0 {
		return nil, fmt.Errorf("order %s tidak ditemukan di Shopee", orderSN)
	}

	order := detailResp.Response.OrderList[0]

	// 1. Check from package_list first if present
	if len(order.PackageList) > 0 {
		for _, pkg := range order.PackageList {
			if order.TrackingNumber == "" && pkg.TrackingNumber != "" {
				order.TrackingNumber = pkg.TrackingNumber
			}
			if order.ShippingCarrier == "" && pkg.ShippingCarrier != "" {
				order.ShippingCarrier = pkg.ShippingCarrier
			}
		}
	}

	// 2. Fallback to /api/v2/logistics/get_tracking_number if tracking number is still empty
	if strings.TrimSpace(order.TrackingNumber) == "" {
		tPath := "/api/v2/logistics/get_tracking_number"
		tTimestamp := time.Now().Unix()
		tSign := s.GenerateShopSign(tPath, tTimestamp, token.AccessToken, token.ShopID)
		tReqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&order_sn=%s",
			s.BaseURL, tPath, s.PartnerID, tTimestamp, token.AccessToken, token.ShopID, tSign, url.QueryEscape(order.OrderSN))

		tResp, tErr := s.HTTPClient.Get(tReqURL)
		if tErr == nil {
			defer tResp.Body.Close()
			var tData ShopeeTrackingResponse
			if err := json.NewDecoder(tResp.Body).Decode(&tData); err == nil && tData.Response.TrackingNumber != "" {
				order.TrackingNumber = tData.Response.TrackingNumber
			}
		}
	}

	// 3. Fallback to /api/v2/logistics/get_tracking_info
	if strings.TrimSpace(order.TrackingNumber) == "" {
		tPath := "/api/v2/logistics/get_tracking_info"
		tTimestamp := time.Now().Unix()
		tSign := s.GenerateShopSign(tPath, tTimestamp, token.AccessToken, token.ShopID)
		tReqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&order_sn=%s",
			s.BaseURL, tPath, s.PartnerID, tTimestamp, token.AccessToken, token.ShopID, tSign, url.QueryEscape(order.OrderSN))

		tResp, tErr := s.HTTPClient.Get(tReqURL)
		if tErr == nil {
			defer tResp.Body.Close()
			var tData ShopeeTrackingResponse
			if err := json.NewDecoder(tResp.Body).Decode(&tData); err == nil && tData.Response.TrackingNumber != "" {
				order.TrackingNumber = tData.Response.TrackingNumber
			}
		}
	}

	return &order, nil
}

// Shopee Boost API structs
type ShopeeBoostedItem struct {
	ItemID         int64 `json:"item_id"`
	CoolDownSecond int   `json:"cool_down_second"`
	CooldownSecond int   `json:"cooldown_second"`
}

func (b *ShopeeBoostedItem) GetCooldownSeconds() int {
	if b.CoolDownSecond > 0 {
		return b.CoolDownSecond
	}
	return b.CooldownSecond
}

type ShopeeBoostedListResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ItemList []ShopeeBoostedItem `json:"item_list"`
	} `json:"response"`
	RequestID string `json:"request_id"`
}

type ShopeeBoostFailedItem struct {
	ItemID       int64  `json:"item_id"`
	FailedReason string `json:"failed_reason"`
}

type ShopeeBoostItemResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		FailureList   []ShopeeBoostFailedItem `json:"failure_list"`
		SuccessIDList []int64                 `json:"success_id_list"`
	} `json:"response"`
	RequestID string `json:"request_id"`
}

// GetBoostedList retrieves the currently boosted items and their remaining cooldown from Shopee
func (s *ShopeeClient) GetBoostedList(token *domain.MarketplaceToken) ([]ShopeeBoostedItem, error) {
	path := "/api/v2/product/get_boosted_list"
	timestamp := time.Now().Unix()
	sign := s.GenerateShopSign(path, timestamp, token.AccessToken, token.ShopID)

	reqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s",
		s.BaseURL, path, s.PartnerID, timestamp, token.AccessToken, token.ShopID, sign)

	resp, err := s.HTTPClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call get_boosted_list: %w", err)
	}
	defer resp.Body.Close()

	var boostedResp ShopeeBoostedListResponse
	if err := json.NewDecoder(resp.Body).Decode(&boostedResp); err != nil {
		return nil, fmt.Errorf("failed to decode boosted list response: %w", err)
	}

	if boostedResp.Error != "" {
		return nil, fmt.Errorf("shopee api error (%s): %s", boostedResp.Error, boostedResp.Message)
	}

	return boostedResp.Response.ItemList, nil
}

// BoostItem calls Shopee API to boost up to 5 items
func (s *ShopeeClient) BoostItem(token *domain.MarketplaceToken, itemIDs []int64) (*ShopeeBoostItemResponse, error) {
	if len(itemIDs) == 0 {
		return nil, fmt.Errorf("itemIDs cannot be empty")
	}
	if len(itemIDs) > 5 {
		itemIDs = itemIDs[:5]
	}

	path := "/api/v2/product/boost_item"
	timestamp := time.Now().Unix()
	sign := s.GenerateShopSign(path, timestamp, token.AccessToken, token.ShopID)

	reqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s",
		s.BaseURL, path, s.PartnerID, timestamp, token.AccessToken, token.ShopID, sign)

	bodyData, err := json.Marshal(map[string]interface{}{
		"item_id_list": itemIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal boost request: %w", err)
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, fmt.Errorf("failed to create boost request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send boost request: %w", err)
	}
	defer resp.Body.Close()

	var boostResp ShopeeBoostItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&boostResp); err != nil {
		return nil, fmt.Errorf("failed to decode boost response: %w", err)
	}

	if boostResp.Error != "" {
		return &boostResp, fmt.Errorf("shopee api error (%s): %s", boostResp.Error, boostResp.Message)
	}

	return &boostResp, nil
}
