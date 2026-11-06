package domain

import (
	"github.com/google/uuid"
)

type MarketplaceProductMapping struct {
	ID                     uint                  `gorm:"primaryKey" json:"id"`
	BranchID               *uuid.UUID            `gorm:"type:uuid;index" json:"branch_id"`
	SkuID                  string                `gorm:"type:varchar(100);index" json:"sku_id"`
	MarketplaceStockID     *uuid.UUID            `gorm:"type:uuid;column:marketplace_stock_id" json:"marketplace_stock_id"`
	MarketplaceInventoryID *uuid.UUID            `gorm:"type:uuid;column:marketplace_inventory_id" json:"marketplace_inventory_id"`
	Quantity               int                   `gorm:"type:int;default:1" json:"quantity"`
	MarketplaceStock       *MarketplaceStock     `gorm:"foreignKey:MarketplaceStockID" json:"marketplace_stock"`
	MarketplaceInventory   *MarketplaceInventory `gorm:"foreignKey:MarketplaceInventoryID" json:"marketplace_inventory"`
	Branch                 *Branch               `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

// TableName explicitly defines the table name for GORM
func (MarketplaceProductMapping) TableName() string {
	return "marketplace_product_mappings"
}
