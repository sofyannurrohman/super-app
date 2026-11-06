package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeliverySpare struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DeliveryBatchID uuid.UUID `gorm:"type:uuid;not null" json:"delivery_batch_id"`
	ProductID       uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity        int       `gorm:"type:int;default:0" json:"quantity"`
	CreatedAt       time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"default:now()" json:"updated_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (ds *DeliverySpare) BeforeCreate(tx *gorm.DB) (err error) {
	if ds.ID == uuid.Nil {
		ds.ID = uuid.New()
	}
	return
}
