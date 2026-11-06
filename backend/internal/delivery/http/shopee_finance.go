package http

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
)

type shopeeEscrowDetail struct {
	OrderSN string `json:"order_sn"`
	Income  struct {
		EscrowAmount                float64  `json:"escrow_amount"`
		EscrowAmountAfterAdjustment *float64 `json:"escrow_amount_after_adjustment"`
		ActualShippingFee           float64  `json:"actual_shipping_fee"`
		TotalAdjustmentAmount       float64  `json:"total_adjustment_amount"`
		CommissionFee               float64  `json:"commission_fee"`
		ServiceFee                  float64  `json:"service_fee"`
		SellerTransactionFee        float64  `json:"seller_transaction_fee"`
		CampaignFee                 float64  `json:"campaign_fee"`
		OrderAMSCommissionFee       float64  `json:"order_ams_commission_fee"`
		EscrowTax                   float64  `json:"escrow_tax"`
		WithholdingTax              float64  `json:"withholding_tax"`
		WithholdingVATTax           float64  `json:"withholding_vat_tax"`
		WithholdingPITTax           float64  `json:"withholding_pit_tax"`
		WithholdingCITTax           float64  `json:"withholding_cit_tax"`
		ShippingProtectionFee       float64  `json:"shipping_seller_protection_fee_amount"`
		DeliveryProtectionPremium   float64  `json:"delivery_seller_protection_fee_premium_amount"`
		SellerOrderProcessingFee    float64  `json:"seller_order_processing_fee"`
		AdsEscrowFee                float64  `json:"ads_escrow_top_up_fee_or_technical_support_fee"`
		FBSFee                      float64  `json:"fbs_fee"`
	} `json:"order_income"`
}

type shopeeEscrowRelease struct {
	OrderSN           string  `json:"order_sn"`
	EscrowReleaseTime int64   `json:"escrow_release_time"`
	EscrowAmount      float64 `json:"escrow_amount"`
}

func (s *ShopeeClient) paymentGET(token *domain.MarketplaceToken, path string, extra url.Values, dst any) error {
	if token == nil || token.AccessToken == "" || token.ShopID == "" {
		return fmt.Errorf("token Shopee tidak tersedia")
	}
	if s.PartnerID == "" || s.PartnerKey == "" {
		return fmt.Errorf("konfigurasi Shopee Partner ID/Key belum lengkap")
	}
	ts := time.Now().Unix()
	q := url.Values{}
	q.Set("partner_id", s.PartnerID)
	q.Set("timestamp", strconv.FormatInt(ts, 10))
	q.Set("access_token", token.AccessToken)
	q.Set("shop_id", token.ShopID)
	q.Set("sign", s.GenerateShopSign(path, ts, token.AccessToken, token.ShopID))
	for key, values := range extra {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	base := s.BaseURL
	if base == "" {
		base = ShopeeAPIBaseURL
	}
	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Get(base + path + "?" + q.Encode())
	if err != nil {
		return fmt.Errorf("Shopee Payment API: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Shopee Payment API HTTP %d: %s", resp.StatusCode, string(data))
	}
	var envelope struct {
		Error   string          `json:"error"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("respons Shopee Payment API tidak valid: %w", err)
	}
	if envelope.Error != "" {
		return fmt.Errorf("Shopee Payment API %s: %s (periksa izin Payment dan otorisasi toko)", envelope.Error, envelope.Message)
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return fmt.Errorf("Shopee Payment API tidak mengembalikan data")
	}
	return json.Unmarshal(envelope.Data, dst)
}

func (s *ShopeeClient) fetchEscrowDetail(token *domain.MarketplaceToken, orderID string) (shopeeEscrowDetail, error) {
	var detail shopeeEscrowDetail
	q := url.Values{"order_sn": {orderID}}
	err := s.paymentGET(token, "/api/v2/payment/get_escrow_detail", q, &detail)
	if err == nil && detail.OrderSN != orderID {
		err = fmt.Errorf("Shopee mengembalikan order %q untuk permintaan %q", detail.OrderSN, orderID)
	}
	if err == nil && shopeeNetAmount(detail) == 0 && shopeeFeeAmount(detail) == 0 {
		err = fmt.Errorf("Shopee belum menyediakan rincian escrow untuk order %s", orderID)
	}
	return detail, err
}

func (s *ShopeeClient) fetchEscrowDetailWithRetry(token *domain.MarketplaceToken, orderID string) (shopeeEscrowDetail, error) {
	var detail shopeeEscrowDetail
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		detail, err = s.fetchEscrowDetail(token, orderID)
		if err == nil || (!strings.Contains(err.Error(), "error_rate_limit") && !strings.Contains(err.Error(), "HTTP 429")) {
			return detail, err
		}
		time.Sleep(time.Duration(1<<attempt) * time.Second)
	}
	return detail, err
}

func (s *ShopeeClient) fetchEscrowListPage(token *domain.MarketplaceToken, from, to int64, page int) ([]shopeeEscrowRelease, bool, error) {
	q := url.Values{}
	q.Set("release_time_from", strconv.FormatInt(from, 10))
	q.Set("release_time_to", strconv.FormatInt(to, 10))
	q.Set("page_size", "100")
	q.Set("page_no", strconv.Itoa(page))
	var result struct {
		EscrowList json.RawMessage `json:"escrow_list"`
		More       bool            `json:"more"`
		HasMore    bool            `json:"has_more"`
	}
	if err := s.paymentGET(token, "/api/v2/payment/get_escrow_list", q, &result); err != nil {
		return nil, false, err
	}
	if len(result.EscrowList) == 0 {
		return nil, false, fmt.Errorf("respons daftar escrow Shopee tidak berisi escrow_list")
	}
	var items []shopeeEscrowRelease
	if err := json.Unmarshal(result.EscrowList, &items); err != nil {
		return nil, false, fmt.Errorf("format escrow_list Shopee tidak valid: %w", err)
	}
	return items, result.More || result.HasMore || len(items) == 100, nil
}

func shopeeFeeAmount(d shopeeEscrowDetail) float64 {
	i := d.Income
	return math.Max(0, i.CommissionFee+i.ServiceFee+i.SellerTransactionFee+
		i.CampaignFee+i.OrderAMSCommissionFee+i.EscrowTax+
		i.WithholdingTax+i.WithholdingVATTax+i.WithholdingPITTax+i.WithholdingCITTax+
		i.ShippingProtectionFee+i.DeliveryProtectionPremium+i.SellerOrderProcessingFee+
		i.AdsEscrowFee+i.FBSFee)
}

func shopeeNetAmount(d shopeeEscrowDetail) float64 {
	if d.Income.EscrowAmountAfterAdjustment != nil {
		return *d.Income.EscrowAmountAfterAdjustment
	}
	return d.Income.EscrowAmount
}

// Only finance columns are written; marketplace and branch predicates protect other shops.
func (h *MarketplaceHandler) saveShopeeFinance(detail shopeeEscrowDetail, token *domain.MarketplaceToken, releaseTime int64) (int64, error) {
	if detail.OrderSN == "" {
		return 0, fmt.Errorf("order_sn kosong pada rincian Shopee")
	}
	breakdown, err := json.Marshal(detail.Income)
	if err != nil {
		return 0, err
	}
	updates := map[string]any{
		"settlement_amount":    shopeeNetAmount(detail),
		"fee_and_tax_amount":   shopeeFeeAmount(detail),
		"platform_fee":         shopeeFeeAmount(detail),
		"commission_fee":       math.Max(0, detail.Income.CommissionFee),
		"shipping_cost_amount": math.Abs(detail.Income.ActualShippingFee),
		"adjustment_amount":    detail.Income.TotalAdjustmentAmount,
		"fee_tax_breakdown":    string(breakdown),
		"finance_synced_at":    time.Now(),
	}
	if releaseTime > 0 {
		updates["settlement_status"] = "SETTLED"
		updates["settlement_date"] = time.Unix(releaseTime, 0)
		updates["hold_reason"] = ""
	} else {
		// Escrow detail alone is an expected amount, not proof of release.
		updates["settlement_status"] = "HOLD"
		updates["settlement_date"] = nil
		updates["hold_reason"] = "Menunggu konfirmasi pelepasan escrow Shopee"
	}
	q := h.db.Model(&domain.MarketplaceOrder{}).Where("marketplace = ? AND order_id = ?", "shopee", detail.OrderSN)
	if token.BranchID != nil {
		q = q.Where("branch_id = ?", *token.BranchID)
	}
	res := q.Updates(updates)
	return res.RowsAffected, res.Error
}

func (h *MarketplaceHandler) shopeeFinanceShops(branchID *uuid.UUID) ([]domain.MarketplaceToken, error) {
	var tokens []domain.MarketplaceToken
	q := h.db.Where("marketplace = ? AND shop_id <> ''", "shopee")
	if branchID != nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	if err := q.Order("created_at desc").Find(&tokens).Error; err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("belum ada akun Shopee terhubung")
	}
	unique := make([]domain.MarketplaceToken, 0, len(tokens))
	seen := make(map[string]bool)
	for _, token := range tokens {
		key := token.ShopID
		if token.BranchID != nil {
			key += ":" + token.BranchID.String()
		}
		if !seen[key] {
			seen[key] = true
			unique = append(unique, token)
		}
	}
	return unique, nil
}

func (h *MarketplaceHandler) SyncShopeeFinance(c *gin.Context) {
	branch := parseBranchID(c)
	var branchID *uuid.UUID
	if branch != uuid.Nil {
		branchID = &branch
	}
	days := 30
	if value := c.Query("days"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter days tidak valid"})
			return
		}
		days = parsed
	}
	shops, err := h.shopeeFinanceShops(branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.NewString()
	h.setFinanceSyncStatus(id, financeSyncStatus{Status: "processing", Platform: "shopee", Phase: "Menyiapkan akun Shopee"})
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				h.setFinanceSyncStatus(id, financeSyncStatus{Status: "failed", Error: "Proses sinkronisasi Shopee terhenti secara tidak terduga"})
			}
		}()
		client := NewShopeeClient(h.db)
		total, releases, processed := 0, 0, 0
		var failures []string
		for _, shop := range shops {
			h.setFinanceSyncStatus(id, financeSyncStatus{Status: "processing", SyncedCount: total, StatementsCount: releases, ProcessedCount: processed, Phase: "Membaca escrow toko " + shop.ShopID})
			token, err := getValidShopeeTokenScoped(h.db, client.PartnerID, client.PartnerKey, shop.BranchID, shop.ShopID)
			if err != nil {
				failures = append(failures, fmt.Sprintf("toko %s: %v", shop.ShopID, err))
				continue
			}
			count, releaseCount, processedCount, err := h.executeShopeeFinanceSync(client, token, days, func(synced, released, checked int, phase string) {
				h.setFinanceSyncStatus(id, financeSyncStatus{Status: "processing", SyncedCount: total + synced, StatementsCount: releases + released, ProcessedCount: processed + checked, Phase: phase})
			})
			total += count
			releases += releaseCount
			processed += processedCount
			if err != nil {
				failures = append(failures, fmt.Sprintf("toko %s: %v", shop.ShopID, err))
			}
		}
		status := financeSyncStatus{Status: "completed", SyncedCount: total, StatementsCount: releases, ProcessedCount: processed, Phase: "Selesai"}
		if len(failures) > 0 {
			status.Status = "failed"
			status.Error = strings.Join(failures, "; ")
		}
		h.setFinanceSyncStatus(id, status)
	}()
	c.JSON(http.StatusOK, gin.H{"status": "processing", "in_background": true, "sync_id": id,
		"message": "Sinkronisasi escrow Shopee berjalan di latar belakang."})
}

func (h *MarketplaceHandler) executeShopeeFinanceSync(client *ShopeeClient, token *domain.MarketplaceToken, days int, onProgress func(int, int, int, string)) (int, int, int, error) {
	to := time.Now().Unix()
	from := time.Now().AddDate(0, 0, -days).Unix()
	if days == 0 {
		var firstUnix int64
		q := h.db.Model(&domain.MarketplaceOrder{}).Where("marketplace = ?", "shopee")
		if token.BranchID != nil {
			q = q.Where("branch_id = ?", *token.BranchID)
		}
		if err := q.Select("COALESCE(MIN(NULLIF(create_time_raw, 0)), 0)").Scan(&firstUnix).Error; err != nil {
			return 0, 0, 0, err
		}
		if firstUnix == 0 {
			return 0, 0, 0, nil
		}
		from = firstUnix
	}
	updated, released, processed := 0, 0, 0
	for start := from; start <= to; start += 7 * 86400 {
		if onProgress != nil {
			onProgress(updated, released, processed, "Memeriksa pencairan Shopee "+time.Unix(start, 0).Format("02 Jan 2006"))
		}
		refreshed, err := getValidShopeeTokenScoped(h.db, client.PartnerID, client.PartnerKey, token.BranchID, token.ShopID)
		if err != nil {
			return updated, released, processed, err
		}
		token = refreshed
		end := start + 7*86400 - 1
		if end > to {
			end = to
		}
		for page := 1; page <= 1000; page++ {
			items, more, err := client.fetchEscrowListPage(token, start, end, page)
			if err != nil {
				return updated, released, processed, err
			}
			for _, item := range items {
				processed++
				if item.OrderSN == "" || item.EscrowReleaseTime <= 0 {
					continue
				}
				released++
				var exists int64
				q := h.db.Model(&domain.MarketplaceOrder{}).Where("marketplace = ? AND order_id = ?", "shopee", item.OrderSN)
				if token.BranchID != nil {
					q = q.Where("branch_id = ?", *token.BranchID)
				}
				if err := q.Count(&exists).Error; err != nil {
					return updated, released, processed, err
				}
				if exists == 0 {
					continue
				}
				detail, err := client.fetchEscrowDetailWithRetry(token, item.OrderSN)
				if err != nil {
					return updated, released, processed, fmt.Errorf("order %s: %w", item.OrderSN, err)
				}
				count, err := h.saveShopeeFinance(detail, token, item.EscrowReleaseTime)
				if err != nil {
					return updated, released, processed, err
				}
				updated += int(count)
				if onProgress != nil && processed%20 == 0 {
					onProgress(updated, released, processed, "Mengisi fee Shopee hingga order "+item.OrderSN)
				}
			}
			if onProgress != nil {
				onProgress(updated, released, processed, "Memeriksa pencairan Shopee "+time.Unix(start, 0).Format("02 Jan 2006"))
			}
			if !more {
				break
			}
			if page == 1000 {
				return updated, released, processed, fmt.Errorf("pagination Shopee melebihi 1000 halaman untuk satu periode")
			}
		}
	}
	return updated, released, processed, nil
}

func (h *MarketplaceHandler) SyncShopeeOrderFinance(c *gin.Context) {
	orderID := strings.TrimSpace(c.Param("order_id"))
	branch := parseBranchID(c)
	q := h.db.Model(&domain.MarketplaceOrder{}).Where("marketplace = ? AND order_id = ?", "shopee", orderID)
	if branch != uuid.Nil {
		q = q.Where("branch_id = ?", branch)
	}
	var order domain.MarketplaceOrder
	if err := q.First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan Shopee tidak ditemukan"})
		return
	}
	client := NewShopeeClient(h.db)
	token, err := getValidShopeeTokenScoped(h.db, client.PartnerID, client.PartnerKey, order.BranchID, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	detail, err := client.fetchEscrowDetail(token, orderID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	releaseTime := int64(0)
	if order.SettlementDate != nil && order.FinanceSyncedAt != nil && order.SettlementStatus == "SETTLED" {
		releaseTime = order.SettlementDate.Unix()
	}
	count, err := h.saveShopeeFinance(detail, token, releaseTime)
	if err != nil || count == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal menyimpan finansial Shopee: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rincian escrow Shopee diperbarui. Status cair memerlukan konfirmasi daftar pencairan.",
		"order_id": orderID, "settlement_amount": shopeeNetAmount(detail), "fee_and_tax_amount": shopeeFeeAmount(detail),
		"shipping_cost_amount": detail.Income.ActualShippingFee, "settlement_status": map[bool]string{true: "SETTLED", false: "HOLD"}[releaseTime > 0],
		"fee_tax_breakdown": stringMustJSON(detail.Income)})
}

func stringMustJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
