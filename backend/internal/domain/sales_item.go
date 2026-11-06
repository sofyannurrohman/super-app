package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesItem struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesTransactionID uuid.UUID `gorm:"type:uuid;not null" json:"sales_transaction_id"`
	ProductID          uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity           int       `gorm:"type:int;not null" json:"quantity"`
	PriceAtTransaction float64   `gorm:"type:decimal(15,2);not null" json:"price_at_transaction"`
	DiscountPercent    float64   `gorm:"type:decimal(5,2);default:0" json:"discount_percent"`
	Subtotal           float64   `gorm:"type:decimal(15,2);not null" json:"subtotal"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (si *SalesItem) BeforeCreate(tx *gorm.DB) (err error) {
	if si.ID == uuid.Nil {
		si.ID = uuid.New()
	}
	return
}
