package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DOType membedakan DO yang berasal dari Surat Jalan biasa vs DO Motoris (standalone)
type DOType string

const (
	DOTypeRegular    DOType = "REGULAR"    // Terhubung ke DeliveryBatch (SJ)
	DOTypeMotoris    DOType = "MOTORIS"    // Standalone, langsung dari SalesTransfer
	DOTypeDelegation DOType = "DELEGATION" // Standalone, untuk delegasi cabang (ProductTransfer)
)

type WarehouseDeliveryOrder struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DeliveryBatchID *uuid.UUID `gorm:"-:migration;type:uuid;column:delivery_batch_id" json:"delivery_batch_id"` // nullable: nil untuk DO Motoris
	DOType          DOType     `gorm:"type:varchar(20);default:'REGULAR'" json:"do_type"`
	EmployeeID      *uuid.UUID `gorm:"type:uuid" json:"employee_id"` // Salesman tujuan (untuk DO Motoris)
	DONumber        string     `gorm:"type:varchar(50);unique;not null" json:"do_number"`
	Notes           string     `gorm:"type:text" json:"notes"`
	Status          string     `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	CreatedAt       time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"default:now()" json:"updated_at"`

	DeliveryBatch  *DeliveryBatch                   `gorm:"foreignKey:DeliveryBatchID" json:"delivery_batch,omitempty"`
	Employee       *Employee                        `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Items          []WarehouseDeliveryOrderItem     `gorm:"foreignKey:WarehouseDeliveryOrderID" json:"items,omitempty"`
	SalesTransfers []WarehouseDeliveryOrderTransfer `gorm:"foreignKey:WarehouseDeliveryOrderID" json:"sales_transfers,omitempty"`
}

// WarehouseDeliveryOrderTransfer menghubungkan WDO ke SalesTransfer atau SalesBonusRequest (untuk DO Motoris)
type WarehouseDeliveryOrderTransfer struct {
	ID                       uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WarehouseDeliveryOrderID uuid.UUID  `gorm:"type:uuid;not null" json:"warehouse_delivery_order_id"`
	SalesTransferID          *uuid.UUID `gorm:"type:uuid" json:"sales_transfer_id"`
	SalesBonusRequestID      *uuid.UUID `gorm:"type:uuid" json:"sales_bonus_request_id"`

	SalesTransfer     *SalesTransfer     `gorm:"foreignKey:SalesTransferID" json:"sales_transfer,omitempty"`
	SalesBonusRequest *SalesBonusRequest `gorm:"foreignKey:SalesBonusRequestID" json:"sales_bonus_request,omitempty"`
}

type WarehouseDeliveryOrderItem struct {
	ID                       uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WarehouseDeliveryOrderID uuid.UUID  `gorm:"type:uuid;not null" json:"warehouse_delivery_order_id"`
	ProductID                *uuid.UUID `gorm:"type:uuid" json:"product_id"`
	BonusInventoryID         *uuid.UUID `gorm:"type:uuid" json:"bonus_inventory_id"`
	CartonsLoaded            int        `gorm:"type:int;default:0" json:"cartons_loaded"`
	PcsLoaded                int        `gorm:"type:int;default:0" json:"pcs_loaded"`
	CustomPcsPerUnit         int        `gorm:"type:int;default:1" json:"custom_pcs_per_unit"`

	// Reconciliation fields
	PcsReturnedToWarehouse  int `gorm:"type:int;default:0" json:"pcs_returned_to_warehouse"`
	PcsSold                 int `gorm:"type:int;default:0" json:"pcs_sold"`
	PcsSample               int `gorm:"type:int;default:0" json:"pcs_sample"`
	PcsReturnedFromCustomer int `gorm:"type:int;default:0" json:"pcs_returned_from_customer"`

	Product        *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	BonusInventory *BonusInventory `gorm:"foreignKey:BonusInventoryID" json:"bonus_inventory,omitempty"`
}

func (wdo *WarehouseDeliveryOrder) BeforeCreate(tx *gorm.DB) (err error) {
	if wdo.ID == uuid.Nil {
		wdo.ID = uuid.New()
	}
	return
}

func (wdoi *WarehouseDeliveryOrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	if wdoi.ID == uuid.Nil {
		wdoi.ID = uuid.New()
	}
	return
}

func (wdot *WarehouseDeliveryOrderTransfer) BeforeCreate(tx *gorm.DB) (err error) {
	if wdot.ID == uuid.Nil {
		wdot.ID = uuid.New()
	}
	return
}
