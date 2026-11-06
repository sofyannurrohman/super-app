package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MarketplaceToken struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	BranchID     *uuid.UUID     `gorm:"type:uuid;index;uniqueIndex:idx_marketplace_branch" json:"branch_id"`
	Marketplace  string         `gorm:"type:varchar(50);not null;uniqueIndex:idx_marketplace_branch" json:"marketplace"` // 'shopee' atau 'tokopedia' atau 'tiktok'
	ShopID       string         `gorm:"type:varchar(100);not null" json:"shop_id"`
	AccessToken  string         `gorm:"type:text;not null" json:"access_token"`
	RefreshToken string         `gorm:"type:text" json:"refresh_token"`
	ExpiresAt    time.Time      `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Branch       *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

type MarketplaceConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Marketplace string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"marketplace"` // 'shopee', 'tiktok', 'tokopedia'
	AppID       string    `gorm:"type:varchar(100)" json:"app_id"`                         // Partner ID (Shopee) / App Key (TikTok)
	AppSecret   string    `gorm:"type:text" json:"app_secret,omitempty"`                   // Partner Key (Shopee) / App Secret (TikTok)
	Environment string    `gorm:"type:varchar(50);default:'production'" json:"environment"` // 'production' or 'sandbox'
	BaseURL     string    `gorm:"type:text" json:"base_url"`
	RedirectURL string    `gorm:"type:text" json:"redirect_url"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SaveShopeeConfigReq struct {
	PartnerID   string `json:"partner_id" binding:"required"`
	PartnerKey  string `json:"partner_key"`
	Environment string `json:"environment"` // 'production' or 'sandbox'
}

type ShopeeConfigResponse struct {
	PartnerID      string `json:"partner_id"`
	PartnerKeyMask string `json:"partner_key_mask"`
	HasPartnerKey  bool   `json:"has_partner_key"`
	Environment    string `json:"environment"`
	BaseURL        string `json:"base_url"`
}

