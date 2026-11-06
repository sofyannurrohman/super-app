package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductExchangeStatus string

const (
	ExchangeStatusPending   ProductExchangeStatus = "PENDING"
	ExchangeStatusApproved  ProductExchangeStatus = "APPROVED"
	ExchangeStatusCompleted ProductExchangeStatus = "COMPLETED"
	ExchangeStatusRejected  ProductExchangeStatus = "REJECTED"
)

type ProductExchange struct {
	ID           uuid.UUID             `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ExchangeNo   string                `gorm:"type:varchar(50);unique;not null" json:"exchange_no"`
	StoreID      uuid.UUID             `gorm:"type:uuid;not null" json:"store_id"`
	BranchID     uuid.UUID             `gorm:"type:uuid;not null" json:"branch_id"`
	EmployeeID   uuid.UUID             `gorm:"type:uuid;not null" json:"employee_id"` // inputter
	Status       ProductExchangeStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	DeliveryType string                `gorm:"type:varchar(20);default:'DELIVERY'" json:"delivery_type"` // DIRECT, DELIVERY
	Notes        string                `gorm:"type:text" json:"notes"`

	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	ApprovedByID *uuid.UUID `gorm:"type:uuid" json:"approved_by_id"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Store      *Store                `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Branch     *Branch               `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Employee   *Employee             `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	ApprovedBy *Employee             `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	Items      []ProductExchangeItem `gorm:"foreignKey:ProductExchangeID" json:"items,omitempty"`
}

type ProductExchangeItem struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductExchangeID uuid.UUID `gorm:"type:uuid;not null" json:"product_exchange_id"`
	ProductID         uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity          int       `gorm:"type:int;not null" json:"quantity"`
	Reason            string    `gorm:"type:varchar(255)" json:"reason"`
	Condition         string    `gorm:"type:varchar(20);default:'DAMAGED'" json:"condition"` // DAMAGED, GOOD

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (pe *ProductExchange) BeforeCreate(tx *gorm.DB) (err error) {
	if pe.ID == uuid.Nil {
		pe.ID = uuid.New()
	}
	return
}
