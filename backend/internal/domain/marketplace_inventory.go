package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MarketplaceInventoryType string

const (
	InventoryIn  MarketplaceInventoryType = "IN"
	InventoryOut MarketplaceInventoryType = "OUT"
	InventoryAdj MarketplaceInventoryType = "ADJUSTMENT"
)

type MarketplaceInventory struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID  uuid.UUID `gorm:"type:uuid;not null" json:"branch_id"`
	Name      string    `gorm:"type:varchar(200);not null" json:"name"`
	SKU       string    `gorm:"type:varchar(100)" json:"sku"`
	Unit      string    `gorm:"type:varchar(50);not null" json:"unit"`
	Quantity  int       `gorm:"type:int;default:0" json:"quantity"`
	MinStock  int       `gorm:"type:int;default:0" json:"min_stock"`
	Notes     string    `gorm:"type:text" json:"notes"`
	ImageURL  string    `gorm:"type:varchar(500)" json:"image_url"`
	CreatedBy string    `gorm:"type:varchar(100)" json:"created_by"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

type MarketplaceInventoryLog struct {
	ID                     uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MarketplaceInventoryID uuid.UUID                `gorm:"type:uuid;not null" json:"marketplace_inventory_id"`
	Type                   MarketplaceInventoryType `gorm:"type:varchar(20);not null" json:"type"`
	Quantity               int                      `gorm:"type:int;not null" json:"quantity"`
	QuantityBefore         int                      `gorm:"type:int;not null" json:"quantity_before"`
	QuantityAfter          int                      `gorm:"type:int;not null" json:"quantity_after"`
	Reference              string                   `gorm:"type:varchar(100)" json:"reference"`
	Notes                  string                   `gorm:"type:text" json:"notes"`
	CreatedBy              string                   `gorm:"type:varchar(100)" json:"created_by"`
	CreatedAt              time.Time                `gorm:"default:now()" json:"created_at"`
	MarketplaceInventory   *MarketplaceInventory    `gorm:"foreignKey:MarketplaceInventoryID" json:"marketplace_inventory,omitempty"`
}

type MarketplaceInventoryOpname struct {
	ID                     uuid.UUID             `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MarketplaceInventoryID uuid.UUID             `gorm:"type:uuid;not null" json:"marketplace_inventory_id"`
	QuantitySystem         int                   `gorm:"type:int;not null" json:"quantity_system"`
	QuantityActual         int                   `gorm:"type:int;not null" json:"quantity_actual"`
	Difference             int                   `gorm:"type:int;not null" json:"difference"`
	Notes                  string                `gorm:"type:text" json:"notes"`
	CreatedBy              string                `gorm:"type:varchar(100)" json:"created_by"`
	CreatedAt              time.Time             `gorm:"default:now()" json:"created_at"`
	MarketplaceInventory   *MarketplaceInventory `gorm:"foreignKey:MarketplaceInventoryID" json:"marketplace_inventory,omitempty"`
}

func (m *MarketplaceInventory) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

func (m *MarketplaceInventoryLog) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

func (m *MarketplaceInventoryOpname) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
