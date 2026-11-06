package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MarketplaceOpnameType string

const (
	OpnameStockIn    MarketplaceOpnameType = "STOCK_IN"
	OpnameStockOut   MarketplaceOpnameType = "STOCK_OUT"
	OpnameAdjustment MarketplaceOpnameType = "ADJUSTMENT"
)

type MarketplaceStock struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID    uuid.UUID  `gorm:"type:uuid;not null" json:"branch_id"`
	ProductID   *uuid.UUID `gorm:"type:uuid;index" json:"product_id"`              // FK ke products pabrik (nullable untuk kompatibilitas)
	ProductName string     `gorm:"type:varchar(200);not null" json:"product_name"` // snapshot/cache
	SKU         string     `gorm:"type:varchar(100);not null" json:"sku"`          // snapshot/cache
	ImageURL    string     `gorm:"type:text" json:"image_url"`
	Quantity    int        `gorm:"type:int;default:0" json:"quantity"` // Total
	MinStock    int        `gorm:"type:int;default:0" json:"min_stock"`
	Price       float64    `gorm:"type:decimal(15,2);default:0" json:"price"`
	Notes       string     `gorm:"type:text" json:"notes"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedBy   string     `gorm:"type:varchar(100)" json:"created_by"`
	CreatedAt   time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"default:now()" json:"updated_at"`
	Product     *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type MarketplaceOpname struct {
	ID                 uuid.UUID             `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MarketplaceStockID uuid.UUID             `gorm:"type:uuid;not null" json:"marketplace_stock_id"`
	Type               MarketplaceOpnameType `gorm:"type:varchar(20);not null" json:"type"`
	Quantity           int                   `gorm:"type:int;not null" json:"quantity"`
	QuantityBefore     int                   `gorm:"type:int;not null" json:"quantity_before"`
	QuantityAfter      int                   `gorm:"type:int;not null" json:"quantity_after"`
	Price              float64               `gorm:"type:decimal(15,2);default:0" json:"price"`
	TotalValue         float64               `gorm:"type:decimal(15,2);default:0" json:"total_value"`
	Reference          string                `gorm:"type:varchar(100)" json:"reference"`
	SourceType         string                `gorm:"type:varchar(20);index" json:"source_type"` // MANUAL, TIKTOK, SHOPEE, TOKOPEDIA
	MarketplaceOrderID string                `gorm:"type:varchar(100);index" json:"marketplace_order_id"`
	Notes              string                `gorm:"type:text" json:"notes"`
	CreatedBy          string                `gorm:"type:varchar(100)" json:"created_by"`
	CreatedAt          time.Time             `gorm:"default:now()" json:"created_at"`
	MarketplaceStock   *MarketplaceStock     `gorm:"foreignKey:MarketplaceStockID" json:"marketplace_stock,omitempty"`
}

func (m *MarketplaceStock) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

func (m *MarketplaceOpname) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
