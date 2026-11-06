package domain

import (
	"time"

	"github.com/google/uuid"
)

// MarketplaceProduct menyimpan data produk/SKU yang dijual di marketplace (TikTok)
type MarketplaceProduct struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	BranchID      *uuid.UUID `gorm:"type:uuid;index" json:"branch_id"`
	Marketplace   string     `gorm:"type:varchar(50);not null;default:'tiktok'" json:"marketplace"`
	ProductID     string     `gorm:"type:varchar(100);index" json:"product_id"`
	ProductName   string     `gorm:"type:varchar(500)" json:"product_name"`
	SkuID         string     `gorm:"type:varchar(100);index" json:"sku_id"`
	SellerSKU     string     `gorm:"type:varchar(100)" json:"seller_sku"`
	PriceAmount   float64    `gorm:"type:decimal(15,2);default:0" json:"price_amount"`
	PriceCurrency string     `gorm:"type:varchar(10);default:'IDR'" json:"price_currency"`
	Stock         int        `gorm:"default:0" json:"stock"`
	Status        string     `gorm:"type:varchar(50)" json:"status"` // ACTIVATE, DRAFT, SELLER_DEACTIVATED, FREEZE, DELETED
	Category      string     `gorm:"type:varchar(200)" json:"category"`
	SalesRegions  string     `gorm:"type:varchar(200)" json:"sales_regions"` // comma-separated, e.g. "ID,MY,PH"
	LastSyncedAt  time.Time  `json:"last_synced_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Branch        *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}
