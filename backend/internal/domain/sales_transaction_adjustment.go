package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesTransactionAdjustmentLog struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesTransactionID uuid.UUID `gorm:"type:uuid;not null;index" json:"sales_transaction_id"`
	ProductID          uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	OriginalQuantity   int       `gorm:"type:int;not null" json:"original_quantity"`
	AdjustedQuantity   int       `gorm:"type:int;not null" json:"adjusted_quantity"`
	Difference         int       `gorm:"type:int;not null" json:"difference"`
	Notes              string    `gorm:"type:text" json:"notes"`
	AdminNotaID        uuid.UUID `gorm:"type:uuid" json:"admin_nota_id"`
	CreatedAt          time.Time `gorm:"default:now()" json:"created_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (al *SalesTransactionAdjustmentLog) BeforeCreate(tx *gorm.DB) (err error) {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return
}
