package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
	"tiktokshop/open/sdk_golang/apis"
)

type MarketplaceIntegrationHandler struct {
	DB                 *gorm.DB
	ShopeePartnerID    string
	ShopeePartnerKey   string
	TokopediaAppID     string
	TokopediaAppSecret string
	TikTokAppKey       string
	TikTokAppSecret    string
	TikTokServiceID    string
}

func NewMarketplaceIntegrationHandler(db *gorm.DB) *MarketplaceIntegrationHandler {
	serviceID := os.Getenv("TIKTOK_SERVICE_ID")
	if serviceID == "" {
		serviceID = "7664790051491579655"
	}
	appKey := os.Getenv("TIKTOK_APP_KEY")
	if appKey == "" {
		appKey = "6kmomdj5ihm2p"
	}
	appSecret := os.Getenv("TIKTOK_APP_SECRET")
	if appSecret == "" {
		appSecret = "bc5e00b572660ec28c3e874e47b987c95de045a0"
	}
	return &MarketplaceIntegrationHandler{
		DB:                 db,
		ShopeePartnerID:    os.Getenv("SHOPEE_PARTNER_ID"),
		ShopeePartnerKey:   os.Getenv("SHOPEE_PARTNER_KEY"),
		TokopediaAppID:     os.Getenv("TOKOPEDIA_APP_ID"),
		TokopediaAppSecret: os.Getenv("TOKOPEDIA_APP_SECRET"),
		TikTokAppKey:       appKey,
		TikTokAppSecret:    appSecret,
		TikTokServiceID:    serviceID,
	}
}

func (h *MarketplaceIntegrationHandler) RegisterRoutes(r *gin.RouterGroup) {
	integration := r.Group("/marketplaces")
	{
		integration.GET("/shopee/callback", h.ShopeeCallback)
		integration.POST("/tokopedia/token", h.TokopediaGetToken)
		integration.GET("/tiktok/callback", h.TikTokCallback)

		integration.GET("/accounts", h.GetAccounts)
		integration.GET("/tiktok/auth-url", h.GetTikTokAuthURL)
		integration.GET("/shopee/auth-url", h.GetShopeeAuthURL)
		integration.GET("/shopee/config", h.GetShopeeConfig)
		integration.POST("/shopee/config", h.SaveShopeeConfig)
		integration.POST("/accounts", h.SaveAccount)
		integration.DELETE("/accounts/:id", h.DeleteAccount)
		integration.POST("/tiktok/test-connection", h.TestTikTokConnection)
		integration.POST("/shopee/test-connection", h.TestShopeeConnection)
	}
}

func (h *MarketplaceIntegrationHandler) ShopeeCallback(c *gin.Context) {
	code := c.Query("code")
	shopID := c.Query("shop_id")

	if code == "" || shopID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code or shop_id"})
		return
	}

	client := NewShopeeClient(h.DB)
	if client.PartnerID == "" || client.PartnerKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Shopee Partner ID dan Partner Key belum dikonfigurasi. Silakan atur di menu Pengaturan Shopee API",
		})
		return
	}

	path := "/api/v2/auth/token/get"
	timestamp := time.Now().Unix()
	sign := client.GeneratePublicSign(path, timestamp)

	apiURL := fmt.Sprintf("%s%s", client.BaseURL, path)
	fullURL := fmt.Sprintf("%s?partner_id=%s&timestamp=%d&sign=%s", apiURL, client.PartnerID, timestamp, sign)

	shopIDNum, _ := strconv.ParseInt(shopID, 10, 64)
	partnerIDNum, _ := strconv.ParseInt(client.PartnerID, 10, 64)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"code":       code,
		"shop_id":    shopIDNum,
		"partner_id": partnerIDNum,
	})

	log.Printf("[SHOPEE OAUTH] Exchanging auth code for token | Env: %s | PartnerID: %s | ShopID: %s", client.Environment, maskKey(client.PartnerID), shopID)

	resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("[SHOPEE OAUTH] HTTP request failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to request token"})
		return
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpireIn     int    `json:"expire_in"`
		Error        string `json:"error"`
		Message      string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[SHOPEE OAUTH] Failed to decode response JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode response"})
		return
	}

	if result.Error != "" {
		log.Printf("[SHOPEE OAUTH] Token exchange error response: %s - %s", result.Error, result.Message)
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error, "message": result.Message})
		return
	}

	log.Printf("[SHOPEE OAUTH] Token received successfully | ShopID: %s | ExpiresIn: %ds", shopID, result.ExpireIn)

	var branchID *uuid.UUID
	if bid := c.Query("branch_id"); bid != "" {
		if parsed, err := uuid.Parse(bid); err == nil {
			branchID = &parsed
		}
	} else if state := c.Query("state"); state != "" {
		if parsed, err := uuid.Parse(state); err == nil {
			branchID = &parsed
		}
	}

	if branchID == nil {
		var soloBranch domain.Branch
		if err := h.DB.Where("name ILIKE ?", "%Solo%").First(&soloBranch).Error; err == nil {
			branchID = &soloBranch.ID
		}
	}

	expiresAt := time.Now().Add(4 * time.Hour)
	if result.ExpireIn > 0 {
		expiresAt = time.Now().Add(time.Duration(result.ExpireIn) * time.Second)
	}

	tokenData := domain.MarketplaceToken{
		BranchID:     branchID,
		Marketplace:  "shopee",
		ShopID:       shopID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    expiresAt,
	}

	var existing domain.MarketplaceToken
	var findErr error
	if branchID != nil {
		findErr = h.DB.Where("marketplace = ? AND branch_id = ?", "shopee", *branchID).First(&existing).Error
	} else {
		findErr = h.DB.Where("marketplace = ?", "shopee").First(&existing).Error
	}

	if findErr == nil {
		existing.ShopID = shopID
		existing.AccessToken = result.AccessToken
		existing.RefreshToken = result.RefreshToken
		existing.ExpiresAt = expiresAt
		if branchID != nil {
			existing.BranchID = branchID
		}
		_ = h.DB.Save(&existing)
	} else {
		if createErr := h.DB.Create(&tokenData).Error; createErr != nil {
			h.DB.Where("marketplace = ?", "shopee").Updates(tokenData)
		}
	}

	branchName := "Cabang"
	if branchID != nil {
		var b domain.Branch
		if err := h.DB.First(&b, "id = ?", *branchID).Error; err == nil && b.Name != "" {
			branchName = b.Name
		}
	}

	branchIDStr := ""
	if branchID != nil {
		branchIDStr = branchID.String()
	}

	htmlMsg := fmt.Sprintf(`
	<html>
		<head><title>Shopee Authorization Successful</title></head>
		<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; text-align: center; padding: 50px; background-color: #f8fafc; color: #1e293b;">
			<div style="max-width: 500px; margin: 0 auto; background: white; padding: 30px; border-radius: 16px; box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);">
				<div style="width: 56px; height: 56px; background: #fff7ed; color: #ea580c; border-radius: 50%%; display: flex; align-items: center; justify-content: center; margin: 0 auto 16px auto; font-size: 28px; line-height: 56px;">✓</div>
				<h2 style="color: #0f172a; margin-bottom: 8px;">Otorisasi Shopee Berhasil!</h2>
				<p style="color: #64748b; font-size: 14px; margin-top: 0;">Akun Shopee telah terhubung ke <b>%s</b>.</p>
				<div style="background: #f1f5f9; padding: 12px; border-radius: 8px; margin: 20px 0; text-align: left; font-size: 13px;">
					<div style="color: #64748b; margin-bottom: 4px;">Shop ID:</div>
					<code style="font-weight: bold; color: #0f172a;">%s</code>
				</div>
				<p style="color: #64748b; font-size: 13px;">Anda dapat menutup halaman ini dan kembali ke aplikasi HRIS Web.</p>
			</div>
			<script>
				try {
					localStorage.setItem('shopee_auth_completed', Date.now().toString());
					if (window.opener) {
						window.opener.postMessage({ type: 'SHOPEE_AUTH_SUCCESS', branch_id: '%s' }, '*');
					}
				} catch(e) {}
			</script>
		</body>
	</html>`, branchName, shopID, branchIDStr)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlMsg))
}

func (h *MarketplaceIntegrationHandler) TokopediaGetToken(c *gin.Context) {
	shopID := c.Query("shop_id")
	if shopID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing shop_id"})
		return
	}

	apiURL := "https://accounts.tokopedia.com/token?grant_type=client_credentials"

	authStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", h.TokopediaAppID, h.TokopediaAppSecret)))

	req, _ := http.NewRequest("POST", apiURL, nil)
	req.Header.Add("Authorization", "Basic "+authStr)
	req.Header.Add("Content-Length", "0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to request token"})
		return
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode response"})
		return
	}

	if result.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
		return
	}

	tokenData := domain.MarketplaceToken{
		Marketplace: "tokopedia",
		ShopID:      shopID,
		AccessToken: result.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}

	h.DB.Where(domain.MarketplaceToken{Marketplace: "tokopedia", ShopID: shopID}).
		Assign(tokenData).
		FirstOrCreate(&tokenData)

	c.JSON(http.StatusOK, gin.H{"message": "Tokopedia authorized successfully", "token": result.AccessToken})
}

func (h *MarketplaceIntegrationHandler) TikTokCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	tokenApi := apis.NewAccessToken(h.TikTokAppKey, h.TikTokAppSecret)
	res, err := tokenApi.GetToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to request tiktok token", "details": err.Error()})
		return
	}
	if res.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tiktok api error", "message": res.Message})
		return
	}

	// Attempt to get shop cipher
	cfg := apis.NewConfiguration()
	cfg.AddAppInfo(h.TikTokAppKey, h.TikTokAppSecret)
	client := apis.NewAPIClient(cfg)

	shopsRes, _, err := client.AuthorizationV202309API.Authorization202309ShopsGet(context.Background()).
		XTtsAccessToken(res.Data.AccessToken).
		ContentType("application/json").
		Execute()

	shopCipher := "MANUAL_INPUT_REQUIRED"
	apiSuccess := false
	if err == nil && shopsRes != nil && shopsRes.Code != nil && *shopsRes.Code == 0 && len(shopsRes.Data.Shops) > 0 && shopsRes.Data.Shops[0].Cipher != nil {
		shopCipher = *shopsRes.Data.Shops[0].Cipher
		apiSuccess = true
	}

	var branchID *uuid.UUID
	if bid := c.Query("branch_id"); bid != "" {
		if parsed, err := uuid.Parse(bid); err == nil {
			branchID = &parsed
		}
	} else if state := c.Query("state"); state != "" {
		if parsed, err := uuid.Parse(state); err == nil {
			branchID = &parsed
		}
	}
	if branchID == nil {
		var soloBranch domain.Branch
		if err := h.DB.Where("name ILIKE ?", "%Solo%").First(&soloBranch).Error; err == nil {
			branchID = &soloBranch.ID
		}
	}

	var expiresAt time.Time
	if res.Data.AccessTokenExpireIn > 1000000000 {
		expiresAt = time.Unix(res.Data.AccessTokenExpireIn, 0)
	} else if res.Data.AccessTokenExpireIn > 0 {
		expiresAt = time.Now().Add(time.Duration(res.Data.AccessTokenExpireIn) * time.Second)
	} else {
		expiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	tokenData := domain.MarketplaceToken{
		BranchID:     branchID,
		Marketplace:  "tiktok",
		ShopID:       shopCipher,
		AccessToken:  res.Data.AccessToken,
		RefreshToken: res.Data.RefreshToken,
		ExpiresAt:    expiresAt,
	}

	// Simpan ke DB. Jika sudah ada untuk cabang tersebut, update.
	var existing domain.MarketplaceToken
	var findErr error
	if branchID != nil {
		findErr = h.DB.Where("marketplace = ? AND branch_id = ?", "tiktok", *branchID).First(&existing).Error
	} else {
		findErr = h.DB.Where("marketplace = ?", "tiktok").First(&existing).Error
	}

	if findErr == nil {
		existing.ShopID = shopCipher
		existing.AccessToken = res.Data.AccessToken
		existing.RefreshToken = res.Data.RefreshToken
		existing.ExpiresAt = expiresAt
		if branchID != nil {
			existing.BranchID = branchID
		}
		if saveErr := h.DB.Save(&existing).Error; saveErr != nil {
			fmt.Printf("Error saving updated token: %v\n", saveErr)
		}
	} else {
		if createErr := h.DB.Create(&tokenData).Error; createErr != nil {
			fmt.Printf("Error creating token, trying update: %v\n", createErr)
			h.DB.Where("marketplace = ?", "tiktok").Updates(tokenData)
		}
	}

	branchName := "Cabang"
	if branchID != nil {
		var b domain.Branch
		if err := h.DB.First(&b, "id = ?", *branchID).Error; err == nil && b.Name != "" {
			branchName = b.Name
		}
	}

	branchIDStr := ""
	if branchID != nil {
		branchIDStr = branchID.String()
	}

	var htmlMsg string
	if apiSuccess {
		htmlMsg = fmt.Sprintf(`
		<html>
			<head><title>TikTok Authorization Successful</title></head>
			<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; text-align: center; padding: 50px; background-color: #f8fafc; color: #1e293b;">
				<div style="max-width: 500px; margin: 0 auto; background: white; padding: 30px; border-radius: 16px; box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);">
					<div style="width: 56px; height: 56px; background: #ecfdf5; color: #059669; border-radius: 50%%; display: flex; align-items: center; justify-content: center; margin: 0 auto 16px auto; font-size: 28px; line-height: 56px;">✓</div>
					<h2 style="color: #0f172a; margin-bottom: 8px;">Otorisasi TikTok Berhasil!</h2>
					<p style="color: #64748b; font-size: 14px; margin-top: 0;">Akun TikTok Shop telah terhubung ke <b>%s</b>.</p>
					<div style="background: #f1f5f9; padding: 12px; border-radius: 8px; margin: 20px 0; text-align: left; font-size: 13px;">
						<div style="color: #64748b; margin-bottom: 4px;">Shop Cipher / Shop ID:</div>
						<code style="font-weight: bold; color: #0f172a;">%s</code>
					</div>
					<p style="color: #64748b; font-size: 13px;">Anda dapat menutup halaman ini dan kembali ke aplikasi HRIS Web.</p>
				</div>
				<script>
					try {
						localStorage.setItem('tiktok_auth_completed', Date.now().toString());
						if (window.opener) {
							window.opener.postMessage({ type: 'TIKTOK_AUTH_SUCCESS', branch_id: '%s' }, '*');
						}
					} catch(e) {}
				</script>
			</body>
		</html>`, branchName, shopCipher, branchIDStr)
	} else {
		htmlMsg = fmt.Sprintf(`
		<html>
			<head><title>TikTok Token Saved</title></head>
			<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; text-align: center; padding: 50px; background-color: #f8fafc; color: #1e293b;">
				<div style="max-width: 550px; margin: 0 auto; background: white; padding: 30px; border-radius: 16px; box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);">
					<div style="width: 56px; height: 56px; background: #ecfdf5; color: #059669; border-radius: 50%%; display: flex; align-items: center; justify-content: center; margin: 0 auto 16px auto; font-size: 28px; line-height: 56px;">✓</div>
					<h2 style="color: #0f172a; margin-bottom: 8px;">Token TikTok Tersimpan!</h2>
					<p style="color: #64748b; font-size: 14px; margin-top: 0;">Access Token untuk <b>%s</b> telah tersimpan di database.</p>
					<div style="background: #fffbeb; border: 1px solid #fef3c7; color: #92400e; padding: 12px; border-radius: 8px; margin: 20px 0; text-align: left; font-size: 13px;">
						<b>Catatan:</b> Jika Shop Cipher belum terisi otomatis, Anda dapat mengisi Shop Cipher di menu <b>Integrasi Marketplace</b> pada Web HRIS.
					</div>
					<p style="color: #64748b; font-size: 13px;">Anda dapat menutup halaman ini dan kembali ke aplikasi HRIS Web.</p>
				</div>
				<script>
					try {
						localStorage.setItem('tiktok_auth_completed', Date.now().toString());
						if (window.opener) {
							window.opener.postMessage({ type: 'TIKTOK_AUTH_SUCCESS', branch_id: '%s' }, '*');
						}
					} catch(e) {}
				</script>
			</body>
		</html>`, branchName, branchIDStr)
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlMsg))
}

type BranchMarketplaceAccountDTO struct {
	BranchID      string                        `json:"branch_id"`
	BranchName    string                        `json:"branch_name"`
	BranchAddress string                        `json:"branch_address"`
	Accounts      []MarketplaceAccountDetailDTO `json:"accounts"`
}

type MarketplaceAccountDetailDTO struct {
	ID              uint       `json:"id"`
	Marketplace     string     `json:"marketplace"`
	ShopID          string     `json:"shop_id"`
	IsConnected     bool       `json:"is_connected"`
	IsExpired       bool       `json:"is_expired"`
	ExpiresAt       *time.Time `json:"expires_at"`
	HasRefreshToken bool       `json:"has_refresh_token"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// GetAccounts returns all branches along with their connected marketplace accounts.
func (h *MarketplaceIntegrationHandler) GetAccounts(c *gin.Context) {
	var branches []domain.Branch
	if err := h.DB.Order("name ASC").Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var tokens []domain.MarketplaceToken
	h.DB.Find(&tokens)

	// Map branchID -> []tokens
	tokenMap := make(map[string][]domain.MarketplaceToken)
	for _, tok := range tokens {
		if tok.BranchID != nil {
			bidStr := tok.BranchID.String()
			tokenMap[bidStr] = append(tokenMap[bidStr], tok)
		}
	}

	now := time.Now()
	result := make([]BranchMarketplaceAccountDTO, 0, len(branches))

	for _, b := range branches {
		bidStr := b.ID.String()
		branchTokens := tokenMap[bidStr]

		accList := make([]MarketplaceAccountDetailDTO, 0)
		for _, tok := range branchTokens {
			exp := tok.ExpiresAt
			accList = append(accList, MarketplaceAccountDetailDTO{
				ID:              tok.ID,
				Marketplace:     tok.Marketplace,
				ShopID:          tok.ShopID,
				IsConnected:     true,
				IsExpired:       now.After(tok.ExpiresAt),
				ExpiresAt:       &exp,
				HasRefreshToken: tok.RefreshToken != "",
				UpdatedAt:       tok.UpdatedAt,
			})
		}

		result = append(result, BranchMarketplaceAccountDTO{
			BranchID:      bidStr,
			BranchName:    b.Name,
			BranchAddress: b.Address,
			Accounts:      accList,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetTikTokAuthURL generates the TikTok OAuth authorize URL with state set to branch_id.
func (h *MarketplaceIntegrationHandler) GetTikTokAuthURL(c *gin.Context) {
	branchID := c.Query("branch_id")
	if branchID == "" {
		branchID = parseBranchID(c).String()
	}

	appKey := h.TikTokAppKey
	if appKey == "" {
		appKey = "6kmomdj5ihm2p"
	}
	// Finance Transaction API v202501 requires this seller scope.
	authURL := fmt.Sprintf("https://auth.tiktok-shops.com/oauth/authorize?app_key=%s&state=%s&scope=seller.finance.info", appKey, branchID)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"app_key":  appKey,
		"state":    branchID,
	})
}

// SaveAccount allows creating or updating a marketplace token manually.
func (h *MarketplaceIntegrationHandler) SaveAccount(c *gin.Context) {
	var req struct {
		ID           *uint   `json:"id"`
		BranchID     string  `json:"branch_id" binding:"required"`
		Marketplace  string  `json:"marketplace" binding:"required"`
		ShopID       string  `json:"shop_id" binding:"required"`
		AccessToken  string  `json:"access_token" binding:"required"`
		RefreshToken string  `json:"refresh_token"`
		ExpiresAt    *string `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branchUUID, err := uuid.Parse(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, *req.ExpiresAt); parseErr == nil {
			expiresAt = parsed
		} else if parsedDate, parseErr2 := time.Parse("2006-01-02", *req.ExpiresAt); parseErr2 == nil {
			expiresAt = parsedDate
		}
	}

	tokenData := domain.MarketplaceToken{
		BranchID:     &branchUUID,
		Marketplace:  req.Marketplace,
		ShopID:       req.ShopID,
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		ExpiresAt:    expiresAt,
	}

	var existing domain.MarketplaceToken
	found := false
	if req.ID != nil && *req.ID > 0 {
		if err := h.DB.First(&existing, *req.ID).Error; err == nil {
			found = true
		}
	}
	if !found {
		if err := h.DB.Where("marketplace = ? AND branch_id = ?", req.Marketplace, branchUUID).First(&existing).Error; err == nil {
			found = true
		}
	}

	if found {
		existing.ShopID = req.ShopID
		if req.AccessToken != "" {
			existing.AccessToken = req.AccessToken
		}
		if req.RefreshToken != "" {
			existing.RefreshToken = req.RefreshToken
		}
		if req.ExpiresAt != nil && *req.ExpiresAt != "" {
			existing.ExpiresAt = expiresAt
		}
		existing.BranchID = &branchUUID
		if err := h.DB.Save(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tokenData = existing
	} else {
		if err := h.DB.Create(&tokenData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Akun marketplace berhasil disimpan",
		"data":    tokenData,
	})
}

// DeleteAccount removes a marketplace token record.
func (h *MarketplaceIntegrationHandler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	if err := h.DB.Unscoped().Delete(&domain.MarketplaceToken{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Akun marketplace berhasil dihapus"})
}

// TestTikTokConnection tests if the TikTok access token for a given branch can communicate with TikTok API.
func (h *MarketplaceIntegrationHandler) TestTikTokConnection(c *gin.Context) {
	var req struct {
		BranchID string `json:"branch_id"`
	}
	_ = c.ShouldBindJSON(&req)

	var branchIDPtr *uuid.UUID
	if req.BranchID != "" {
		if parsed, err := uuid.Parse(req.BranchID); err == nil {
			branchIDPtr = &parsed
		}
	} else {
		bid := parseBranchID(c)
		if bid != uuid.Nil {
			branchIDPtr = &bid
		}
	}

	tokenRecord, err := getValidTikTokToken(h.DB, h.TikTokAppKey, h.TikTokAppSecret, branchIDPtr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Gagal memuat token: %v", err),
		})
		return
	}

	// Try fetching warehouses to verify token validity
	cfg := apis.NewConfiguration()
	cfg.AddAppInfo(h.TikTokAppKey, h.TikTokAppSecret)
	client := apis.NewAPIClient(cfg)

	whResp, httpRes, errWh := client.LogisticsV202309API.Logistics202309WarehousesGet(context.Background()).
		ShopCipher(tokenRecord.ShopID).
		XTtsAccessToken(tokenRecord.AccessToken).
		ContentType("application/json").
		Execute()

	if errWh != nil {
		bodyStr := ""
		if httpRes != nil && httpRes.Body != nil {
			bodyBytes, _ := io.ReadAll(httpRes.Body)
			bodyStr = string(bodyBytes)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Gagal menghubungi TikTok API: %v (%s)", errWh, bodyStr),
			"shop_id": tokenRecord.ShopID,
		})
		return
	}

	warehouseCount := 0
	if whResp != nil && whResp.Data != nil && whResp.Data.Warehouses != nil {
		warehouseCount = len(whResp.Data.Warehouses)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Koneksi ke TikTok Shop aktif dan valid!",
		"shop_id":         tokenRecord.ShopID,
		"expires_at":      tokenRecord.ExpiresAt,
		"warehouse_count": warehouseCount,
	})
}

// GetShopeeConfig retrieves the active Shopee API configuration
func (h *MarketplaceIntegrationHandler) GetShopeeConfig(c *gin.Context) {
	client := NewShopeeClient(h.DB)

	hasSecret := client.PartnerKey != ""
	maskedSecret := ""
	if hasSecret {
		if len(client.PartnerKey) > 8 {
			maskedSecret = client.PartnerKey[:4] + strings.Repeat("*", len(client.PartnerKey)-8) + client.PartnerKey[len(client.PartnerKey)-4:]
		} else {
			maskedSecret = "********"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"partner_id":    client.PartnerID,
		"partner_key":   maskedSecret,
		"has_key":       hasSecret,
		"environment":   client.Environment,
		"base_url":      client.BaseURL,
		"is_configured": client.PartnerID != "" && client.PartnerKey != "",
	})
}

// SaveShopeeConfig saves Shopee API credentials (Partner ID, Key, Environment)
func (h *MarketplaceIntegrationHandler) SaveShopeeConfig(c *gin.Context) {
	var req domain.SaveShopeeConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	partnerID := strings.TrimSpace(req.PartnerID)
	partnerKey := strings.TrimSpace(req.PartnerKey)
	environment := strings.ToLower(strings.TrimSpace(req.Environment))
	if environment == "" {
		environment = "production"
	}

	if partnerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Partner ID (App ID) wajib diisi"})
		return
	}

	var existing domain.MarketplaceConfig
	err := h.DB.Where("marketplace = ?", "shopee").First(&existing).Error
	if err == nil {
		// Update existing
		existing.AppID = partnerID
		if partnerKey != "" {
			existing.AppSecret = partnerKey
		}
		existing.Environment = environment
		existing.IsActive = true
		existing.UpdatedAt = time.Now()
		if saveErr := h.DB.Save(&existing).Error; saveErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan konfigurasi: " + saveErr.Error()})
			return
		}
	} else {
		// Create new
		if partnerKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Partner Key (App Secret) wajib diisi untuk konfigurasi baru"})
			return
		}
		newCfg := domain.MarketplaceConfig{
			Marketplace: "shopee",
			AppID:       partnerID,
			AppSecret:   partnerKey,
			Environment: environment,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if createErr := h.DB.Create(&newCfg).Error; createErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat konfigurasi: " + createErr.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Konfigurasi Shopee Open Platform berhasil disimpan",
		"partner_id":  partnerID,
		"environment": environment,
	})
}

// GetShopeeAuthURL generates the Shopee OAuth authorize URL with branch_id in redirect.
func (h *MarketplaceIntegrationHandler) GetShopeeAuthURL(c *gin.Context) {
	branchID := c.Query("branch_id")
	if branchID == "" {
		branchID = parseBranchID(c).String()
	}

	client := NewShopeeClient(h.DB)
	if client.PartnerID == "" || client.PartnerKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Shopee Partner ID dan Partner Key belum dikonfigurasi. Silakan atur di tombol 'Pengaturan Shopee API' terlebih dahulu.",
		})
		return
	}

	baseURL := strings.TrimRight(os.Getenv("APP_BASE_URL"), "/")
	if baseURL == "" {
		host := c.Request.Host
		if host != "" && !strings.Contains(host, "localhost") && !strings.Contains(host, "127.0.0.1") {
			scheme := "https"
			if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") == "http" {
				scheme = "http"
			}
			baseURL = fmt.Sprintf("%s://%s", scheme, host)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "APP_BASE_URL harus diatur sebelum menghubungkan marketplace"})
			return
		}
	}
	redirectURL := fmt.Sprintf("%s/api/v1/marketplaces/shopee/callback?branch_id=%s", baseURL, branchID)

	authURL, err := client.BuildAuthURL(redirectURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url":    authURL,
		"partner_id":  client.PartnerID,
		"environment": client.Environment,
		"branch_id":   branchID,
	})
}

// TestShopeeConnection tests if the Shopee access token for a given branch can communicate with Shopee API.
func (h *MarketplaceIntegrationHandler) TestShopeeConnection(c *gin.Context) {
	var req struct {
		BranchID string `json:"branch_id"`
	}
	_ = c.ShouldBindJSON(&req)

	var branchIDPtr *uuid.UUID
	if req.BranchID != "" {
		if parsed, err := uuid.Parse(req.BranchID); err == nil {
			branchIDPtr = &parsed
		}
	} else {
		bid := parseBranchID(c)
		if bid != uuid.Nil {
			branchIDPtr = &bid
		}
	}

	tokenRecord, err := getValidShopeeToken(h.DB, h.ShopeePartnerID, h.ShopeePartnerKey, branchIDPtr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Gagal memuat token Shopee: %v", err),
		})
		return
	}

	client := NewShopeeClient(h.DB)
	path := "/api/v2/product/get_item_list"
	timestamp := time.Now().Unix()
	sign := client.GenerateShopSign(path, timestamp, tokenRecord.AccessToken, tokenRecord.ShopID)

	reqURL := fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&access_token=%s&shop_id=%s&sign=%s&item_status=NORMAL&offset=0&page_size=10",
		client.BaseURL, path, client.PartnerID, timestamp, tokenRecord.AccessToken, tokenRecord.ShopID, sign)

	resp, errReq := client.HTTPClient.Get(reqURL)
	if errReq != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Gagal menghubungi Shopee API: %v", errReq),
			"shop_id": tokenRecord.ShopID,
		})
		return
	}
	defer resp.Body.Close()

	var itemListResp ShopeeItemListResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemListResp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Gagal memproses response Shopee: %v", err),
			"shop_id": tokenRecord.ShopID,
		})
		return
	}

	if itemListResp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Shopee API Error: %s - %s", itemListResp.Error, itemListResp.Message),
			"shop_id": tokenRecord.ShopID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Koneksi ke Shopee aktif dan valid!",
		"shop_id":       tokenRecord.ShopID,
		"expires_at":    tokenRecord.ExpiresAt,
		"product_count": itemListResp.Response.TotalCount,
	})
}
