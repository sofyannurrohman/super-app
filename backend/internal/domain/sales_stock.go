package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesStock struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_employee_product" json:"employee_id"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_employee_product" json:"product_id"`
	Quantity   int       `gorm:"type:int;default:0" json:"quantity"`
	UpdatedAt  time.Time `gorm:"default:now()" json:"updated_at"`

	Employee *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Product  *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type SalesTransferType string

const (
	SalesTransferOut SalesTransferType = "TRANSFER" // Warehouse -> Salesman
	SalesTransferIn  SalesTransferType = "RETURN"   // Salesman -> Warehouse
)

type SalesTransferStatus string

const (
	SalesTransferWaitingSupervisor SalesTransferStatus = "WAITING_SUPERVISOR"
	SalesTransferPending           SalesTransferStatus = "PENDING"
	SalesTransferProcessing        SalesTransferStatus = "PROCESSING"
	SalesTransferCompleted         SalesTransferStatus = "COMPLETED"
	SalesTransferCancelled         SalesTransferStatus = "CANCELLED"
)

type SalesTransfer struct {
	ID                 uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID           uuid.UUID           `gorm:"type:uuid;not null" json:"branch_id"`
	EmployeeID         uuid.UUID           `gorm:"type:uuid;not null" json:"employee_id"`
	DistributorStoreID *uuid.UUID          `gorm:"type:uuid;index" json:"distributor_store_id,omitempty"`
	RequestSource      string              `gorm:"type:varchar(30);default:'SALES'" json:"request_source"`
	SalesTransactionID *uuid.UUID          `gorm:"type:uuid;index" json:"sales_transaction_id,omitempty"`
	ProductID          *uuid.UUID          `gorm:"type:uuid" json:"product_id"`
	BonusInventoryID   *uuid.UUID          `gorm:"type:uuid" json:"bonus_inventory_id"`
	BatchNo            string              `gorm:"type:varchar(50);default:''" json:"batch_no"`
	Quantity           int                 `gorm:"type:int;not null" json:"quantity"`
	Unit               string              `gorm:"type:varchar(20);default:'pcs'" json:"unit"`
	Type               SalesTransferType   `gorm:"type:varchar(20);not null" json:"type"`
	Status             SalesTransferStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	ReferenceNo        string              `gorm:"type:varchar(50)" json:"reference_no"`
	ActualQuantity     int                 `gorm:"type:int;default:0" json:"actual_quantity"`
	DamagedQuantity    int                 `gorm:"type:int;default:0" json:"damaged_quantity"`
	Notes              string              `gorm:"type:text" json:"notes"`
	TransferDate       time.Time           `gorm:"default:now()" json:"transfer_date"`
	CreatedAt          time.Time           `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time           `gorm:"default:now()" json:"updated_at"`

	Branch           *Branch         `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Employee         *Employee       `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	DistributorStore *Store          `gorm:"foreignKey:DistributorStoreID" json:"distributor_store,omitempty"`
	Product          *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	BonusInventory   *BonusInventory `gorm:"foreignKey:BonusInventoryID" json:"bonus_inventory,omitempty"`

	Corrections []SalesTransferCorrection `gorm:"foreignKey:SalesTransferID" json:"corrections,omitempty"`
}

type SalesTransferCorrection struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesTransferID  uuid.UUID `gorm:"type:uuid;not null" json:"sales_transfer_id"`
	CorrectedBy      string    `gorm:"type:varchar(100);not null" json:"corrected_by"`
	OldQuantity      int       `gorm:"type:int;not null" json:"old_quantity"`
	NewQuantity      int       `gorm:"type:int;not null" json:"new_quantity"`
	OldDamagedQty    int       `gorm:"type:int;not null" json:"old_damaged_qty"`
	NewDamagedQty    int       `gorm:"type:int;not null" json:"new_damaged_qty"`
	CorrectionReason string    `gorm:"type:text;not null" json:"correction_reason"`
	CreatedAt        time.Time `gorm:"default:now()" json:"created_at"`

	SalesTransfer *SalesTransfer `gorm:"foreignKey:SalesTransferID" json:"sales_transfer,omitempty"`
}

func (s *SalesStock) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func (s *SalesTransfer) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.ActualQuantity == 0 && s.Quantity > 0 {
		s.ActualQuantity = s.Quantity
	}
	return
}

func (c *SalesTransferCorrection) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
