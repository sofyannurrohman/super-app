package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RawMaterialPRStatus string

const (
	RawMaterialPRPending   RawMaterialPRStatus = "PENDING"
	RawMaterialPRApproved  RawMaterialPRStatus = "PROCESSED"
	RawMaterialPRRejected  RawMaterialPRStatus = "REJECTED"
	RawMaterialPRCancelled RawMaterialPRStatus = "CANCELLED"
)

type RawMaterialPurchaseRequest struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID       uuid.UUID           `gorm:"type:uuid;not null" json:"company_id"`
	FactoryID       *uuid.UUID          `gorm:"type:uuid" json:"factory_id,omitempty"`
	PRNumber        string              `gorm:"type:varchar(100);not null;index" json:"pr_number"`
	RequestDate     time.Time           `gorm:"type:date;not null" json:"request_date"`
	RequiredDate    *time.Time          `gorm:"type:date" json:"required_date,omitempty"`
	Status          RawMaterialPRStatus `gorm:"type:varchar(30);default:'PENDING'" json:"status"`
	Notes           string              `gorm:"type:text" json:"notes"`
	CreatedByID     *uuid.UUID          `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedByName   string              `gorm:"type:varchar(100)" json:"created_by_name"`
	ReviewedByID    *uuid.UUID          `gorm:"type:uuid" json:"reviewed_by_id,omitempty"`
	ReviewedByName  string              `gorm:"type:varchar(100)" json:"reviewed_by_name,omitempty"`
	ReviewedAt      *time.Time          `json:"reviewed_at,omitempty"`
	ReviewNotes     string              `gorm:"type:text" json:"review_notes,omitempty"`
	RejectionReason string              `gorm:"type:text" json:"rejection_reason,omitempty"`
	PurchaseOrderID *uuid.UUID          `gorm:"type:uuid;index" json:"purchase_order_id,omitempty"`
	CreatedAt       time.Time           `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time           `gorm:"default:now()" json:"updated_at"`

	Company       *Company                         `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Factory       *Factory                         `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
	PurchaseOrder *RawMaterialPurchaseOrder        `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`
	Items         []RawMaterialPurchaseRequestItem `gorm:"foreignKey:PurchaseRequestID" json:"items,omitempty"`
}

type RawMaterialPurchaseRequestItem struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PurchaseRequestID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"purchase_request_id"`
	ProductID          *uuid.UUID `gorm:"type:uuid" json:"product_id,omitempty"`
	ItemName           string     `gorm:"type:varchar(255)" json:"item_name"`
	FactoryWarehouseID *uuid.UUID `gorm:"type:uuid" json:"factory_warehouse_id,omitempty"`
	RequestedQty       int        `gorm:"type:int;not null;default:1" json:"requested_qty"`
	Unit               string     `gorm:"type:varchar(20);default:'KG'" json:"unit"`
	Notes              string     `gorm:"type:text" json:"notes"`
	CreatedAt          time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"default:now()" json:"updated_at"`

	Product          *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	FactoryWarehouse *FactoryWarehouse `gorm:"foreignKey:FactoryWarehouseID" json:"factory_warehouse,omitempty"`
}

func (r *RawMaterialPurchaseRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

func (r *RawMaterialPurchaseRequestItem) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
