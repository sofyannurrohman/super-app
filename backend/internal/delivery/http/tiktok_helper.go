package http

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
	"tiktokshop/open/sdk_golang/apis"
)

// getValidTikTokToken retrieves the TikTok token from DB for a given branch.
// If branchID is not specified or not found, it falls back to the first available token (e.g. Solo branch).
// If token is expired or expires within 5 minutes, it refreshes it.
func getValidTikTokToken(db *gorm.DB, appKey, appSecret string, branchID *uuid.UUID) (*domain.MarketplaceToken, error) {
	var token domain.MarketplaceToken
	q := db.Where("marketplace = ?", "tiktok")
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}

	if err := q.Order("created_at desc").First(&token).Error; err != nil {
		if branchID != nil && *branchID != uuid.Nil {
			return nil, fmt.Errorf("Akun TikTok Shop belum terhubung untuk cabang ini. Silakan hubungkan akun di menu Integrasi Marketplace")
		}
		return nil, fmt.Errorf("TikTok token tidak ditemukan: %w", err)
	}

	return refreshTikTokToken(db, appKey, appSecret, &token)
}

func getValidTikTokTokenByShopID(db *gorm.DB, appKey, appSecret, shopID string) (*domain.MarketplaceToken, error) {
	var token domain.MarketplaceToken
	if err := db.Where("marketplace = ? AND shop_id = ?", "tiktok", shopID).
		Order("expires_at desc").First(&token).Error; err != nil {
		return nil, fmt.Errorf("TikTok token untuk shop %s tidak ditemukan: %w", shopID, err)
	}
	return refreshTikTokToken(db, appKey, appSecret, &token)
}

func refreshTikTokToken(db *gorm.DB, appKey, appSecret string, tokenRecord *domain.MarketplaceToken) (*domain.MarketplaceToken, error) {
	token := *tokenRecord
	// If this shop is shared across branches, check if a peer branch token has a newer active token
	if token.ShopID != "" {
		var peerToken domain.MarketplaceToken
		if err := db.Where("marketplace = ? AND shop_id = ?", "tiktok", token.ShopID).
			Order("expires_at desc").First(&peerToken).Error; err == nil {
			if peerToken.ExpiresAt.After(token.ExpiresAt) && peerToken.AccessToken != "" {
				token.AccessToken = peerToken.AccessToken
				token.RefreshToken = peerToken.RefreshToken
				token.ExpiresAt = peerToken.ExpiresAt
				_ = db.Save(&token)
			}
		}
	}

	// Refresh if expiring in less than 5 minutes
	if time.Now().After(token.ExpiresAt.Add(-5 * time.Minute)) {
		tokenApi := apis.NewAccessToken(appKey, appSecret)
		res, err := tokenApi.RefreshTokenFull(token.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh tiktok token: %w", err)
		}
		if res.Code != 0 {
			return nil, fmt.Errorf("failed to refresh tiktok token, api error: %s", res.Message)
		}

		token.AccessToken = res.Data.AccessToken
		token.RefreshToken = res.Data.RefreshToken
		token.ExpiresAt = time.Unix(res.Data.AccessTokenExpireIn, 0)

		if err := db.Save(&token).Error; err != nil {
			return nil, fmt.Errorf("failed to save refreshed token: %w", err)
		}

		// Also update any other branch tokens sharing the same shop_id to keep them in sync
		if token.ShopID != "" {
			_ = db.Model(&domain.MarketplaceToken{}).
				Where("marketplace = ? AND shop_id = ?", token.Marketplace, token.ShopID).
				Updates(map[string]interface{}{
					"access_token":  token.AccessToken,
					"refresh_token": token.RefreshToken,
					"expires_at":    token.ExpiresAt,
				}).Error
		}
	}

	return &token, nil
}
