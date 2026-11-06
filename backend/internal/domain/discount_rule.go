package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscountRule struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name                  string     `gorm:"type:varchar(100);not null" json:"name"`
	PaymentMethodRequired string     `gorm:"type:varchar(50)" json:"payment_method_required"` // misal: "CASH"
	IsActive              bool       `gorm:"default:true" json:"is_active"`
	StartDate             *time.Time `json:"start_date"`
	EndDate               *time.Time `json:"end_date"`
	CreatedAt             time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"default:now()" json:"updated_at"`

	Tiers            []DiscountTier            `gorm:"foreignKey:DiscountRuleID" json:"tiers"`
	ExcludedProducts []DiscountExcludedProduct `gorm:"foreignKey:DiscountRuleID" json:"excluded_products"`
	ValidUnits       []DiscountValidUnit       `gorm:"foreignKey:DiscountRuleID" json:"valid_units"`
}

type DiscountTier struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DiscountRuleID  uuid.UUID `gorm:"type:uuid;not null;index" json:"discount_rule_id"`
	MinQty          int       `gorm:"type:int;not null" json:"min_qty"`
	DiscountPercent float64   `gorm:"type:decimal(5,2);not null" json:"discount_percent"`
	CreatedAt       time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"default:now()" json:"updated_at"`
}

type DiscountExcludedProduct struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DiscountRuleID uuid.UUID `gorm:"type:uuid;not null;index" json:"discount_rule_id"`
	ProductID      uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type DiscountValidUnit struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DiscountRuleID uuid.UUID `gorm:"type:uuid;not null;index" json:"discount_rule_id"`
	UnitName       string    `gorm:"type:varchar(50);not null" json:"unit_name"` // e.g. "DUS", "BALL", "KRAT", "PACK", "JERIGEN"
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`
}

func (dr *DiscountRule) BeforeCreate(tx *gorm.DB) (err error) {
	if dr.ID == uuid.Nil {
		dr.ID = uuid.New()
	}
	return
}

func (dt *DiscountTier) BeforeCreate(tx *gorm.DB) (err error) {
	if dt.ID == uuid.Nil {
		dt.ID = uuid.New()
	}
	return
}

func (dep *DiscountExcludedProduct) BeforeCreate(tx *gorm.DB) (err error) {
	if dep.ID == uuid.Nil {
		dep.ID = uuid.New()
	}
	return
}

func (dvu *DiscountValidUnit) BeforeCreate(tx *gorm.DB) (err error) {
	if dvu.ID == uuid.Nil {
		dvu.ID = uuid.New()
	}
	return
}
