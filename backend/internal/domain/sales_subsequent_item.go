package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesSubsequentItemStatus string

const (
	SubsequentStatusPending    SalesSubsequentItemStatus = "PENDING"
	SubsequentStatusInDelivery SalesSubsequentItemStatus = "IN_DELIVERY"
	SubsequentStatusShipped    SalesSubsequentItemStatus = "SHIPPED"
	SubsequentStatusDelivered  SalesSubsequentItemStatus = "DELIVERED"
)

type SalesSubsequentItem struct {
	ID                uuid.UUID                 `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesOrderID      *uuid.UUID                `gorm:"type:uuid" json:"sales_order_id,omitempty"`
	ProductExchangeID *uuid.UUID                `gorm:"type:uuid" json:"product_exchange_id,omitempty"`
	SalesReturnID     *uuid.UUID                `gorm:"type:uuid" json:"sales_return_id,omitempty"`
	ProductID         uuid.UUID                 `gorm:"type:uuid;not null" json:"product_id"`
	Quantity          int                       `gorm:"type:int;not null" json:"quantity"`
	Status            SalesSubsequentItemStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	DeliveryBatchID   *uuid.UUID                `gorm:"type:uuid" json:"delivery_batch_id,omitempty"`
	StoreID           *uuid.UUID                `gorm:"type:uuid" json:"store_id,omitempty"`
	SalesmanID        *uuid.UUID                `gorm:"type:uuid" json:"salesman_id,omitempty"`
	Notes             string                    `gorm:"type:text" json:"notes,omitempty"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	SalesOrder      *SalesOrder      `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`
	ProductExchange *ProductExchange `gorm:"foreignKey:ProductExchangeID" json:"product_exchange,omitempty"`
	SalesReturn     *SalesReturn     `gorm:"foreignKey:SalesReturnID" json:"sales_return,omitempty"`
	Product         *Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	DeliveryBatch   *DeliveryBatch   `gorm:"foreignKey:DeliveryBatchID" json:"delivery_batch,omitempty"`
	Store           *Store           `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Salesman        *Employee        `gorm:"foreignKey:SalesmanID" json:"salesman,omitempty"`
}

func (ssi *SalesSubsequentItem) BeforeCreate(tx *gorm.DB) (err error) {
	if ssi.ID == uuid.Nil {
		ssi.ID = uuid.New()
	}
	return
}
