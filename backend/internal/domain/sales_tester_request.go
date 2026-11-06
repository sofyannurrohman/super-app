package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TesterRequestStatus string

const (
	TesterStatusPending    TesterRequestStatus = "PENDING"
	TesterStatusApproved   TesterRequestStatus = "APPROVED"
	TesterStatusInDelivery TesterRequestStatus = "IN_DELIVERY"
	TesterStatusDelivered  TesterRequestStatus = "DELIVERED"
	TesterStatusRejected   TesterRequestStatus = "REJECTED"
)

type SalesTesterRequest struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID        uuid.UUID           `gorm:"type:uuid;not null" json:"branch_id"`
	EmployeeID      uuid.UUID           `gorm:"type:uuid;not null" json:"employee_id"` // Sales TO who requested
	StoreID         *uuid.UUID          `gorm:"type:uuid" json:"store_id,omitempty"`
	ProductID       uuid.UUID           `gorm:"type:uuid;not null" json:"product_id"`
	Quantity        int                 `gorm:"type:int;default:1" json:"quantity"`
	Notes           string              `gorm:"type:text" json:"notes"`
	Status          TesterRequestStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	DeliveryBatchID *uuid.UUID          `gorm:"type:uuid" json:"delivery_batch_id,omitempty"`
	CreatedAt       time.Time           `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time           `gorm:"default:now()" json:"updated_at"`

	Branch        *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Employee      *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Store         *Store         `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Product       *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	DeliveryBatch *DeliveryBatch `gorm:"foreignKey:DeliveryBatchID" json:"delivery_batch,omitempty"`
}

func (str *SalesTesterRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if str.ID == uuid.Nil {
		str.ID = uuid.New()
	}
	return
}
