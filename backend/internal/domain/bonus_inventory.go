package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BonusInventory struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID    uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`
	Quantity    int       `gorm:"type:int;default:0" json:"quantity"`
	Unit        string    `gorm:"type:varchar(50);not null" json:"unit"` // e.g., PCS, LUSIN
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`

	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

type BonusInventoryType string

const (
	BonusInventoryIn  BonusInventoryType = "IN"
	BonusInventoryOut BonusInventoryType = "OUT"
	BonusInventoryAdj BonusInventoryType = "ADJUSTMENT"
)

type BonusInventoryLog struct {
	ID               uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BonusInventoryID uuid.UUID          `gorm:"type:uuid;not null;index" json:"bonus_inventory_id"`
	Type             BonusInventoryType `gorm:"type:varchar(20);not null" json:"type"`
	Quantity         int                `gorm:"type:int;not null" json:"quantity"`
	QuantityBefore   int                `gorm:"type:int;not null" json:"quantity_before"`
	QuantityAfter    int                `gorm:"type:int;not null" json:"quantity_after"`
	Reference        string             `gorm:"type:varchar(100)" json:"reference"`
	Notes            string             `gorm:"type:text" json:"notes"`
	CreatedBy        string             `gorm:"type:varchar(100)" json:"created_by"`
	CreatedAt        time.Time          `gorm:"default:now()" json:"created_at"`

	BonusInventory *BonusInventory `gorm:"foreignKey:BonusInventoryID" json:"bonus_inventory,omitempty"`
}

func (m *BonusInventory) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

func (m *BonusInventoryLog) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
