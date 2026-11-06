package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeliveryBatchSample struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DeliveryBatchID      uuid.UUID  `gorm:"type:uuid;not null" json:"delivery_batch_id"`
	ProductID            uuid.UUID  `gorm:"type:uuid;not null" json:"product_id"`
	Quantity             int        `gorm:"type:int;default:1" json:"quantity"`
	RecipientType        string     `gorm:"type:varchar(20);default:'STORE'" json:"recipient_type"` // STORE, SALESMAN
	StoreID              *uuid.UUID `gorm:"type:uuid" json:"store_id,omitempty"`
	SalesmanID           *uuid.UUID `gorm:"type:uuid" json:"salesman_id,omitempty"`
	SalesTesterRequestID *uuid.UUID `gorm:"type:uuid" json:"sales_tester_request_id,omitempty"`
	Notes                string     `gorm:"type:text" json:"notes"`
	CreatedAt            time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"default:now()" json:"updated_at"`

	Product            *Product            `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Store              *Store              `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Salesman           *Employee           `gorm:"foreignKey:SalesmanID" json:"salesman,omitempty"`
	SalesTesterRequest *SalesTesterRequest `gorm:"foreignKey:SalesTesterRequestID" json:"sales_tester_request,omitempty"`
}

func (dbs *DeliveryBatchSample) BeforeCreate(tx *gorm.DB) (err error) {
	if dbs.ID == uuid.Nil {
		dbs.ID = uuid.New()
	}
	return
}
