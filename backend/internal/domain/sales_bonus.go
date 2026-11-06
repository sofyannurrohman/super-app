package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesBonusRequestStatus string

const (
	SalesBonusRequestPending  SalesBonusRequestStatus = "PENDING"
	SalesBonusRequestApproved SalesBonusRequestStatus = "APPROVED"
	SalesBonusRequestRejected SalesBonusRequestStatus = "REJECTED"
)

type SalesBonusRequest struct {
	ID               uuid.UUID               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID       uuid.UUID               `gorm:"type:uuid;not null" json:"employee_id"`
	BonusInventoryID uuid.UUID               `gorm:"type:uuid;not null" json:"bonus_inventory_id"`
	Quantity         int                     `gorm:"type:int;not null" json:"quantity"`
	Status           SalesBonusRequestStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	Notes            string                  `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time               `gorm:"default:now()" json:"created_at"`
	UpdatedAt        time.Time               `gorm:"default:now()" json:"updated_at"`

	Employee       *Employee       `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	BonusInventory *BonusInventory `gorm:"foreignKey:BonusInventoryID" json:"bonus_inventory,omitempty"`
}

type SalesBonusStock struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID       uuid.UUID `gorm:"type:uuid;not null;index" json:"employee_id"`
	BonusInventoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"bonus_inventory_id"`
	Quantity         int       `gorm:"type:int;default:0" json:"quantity"`
	UpdatedAt        time.Time `gorm:"default:now()" json:"updated_at"`

	Employee       *Employee       `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	BonusInventory *BonusInventory `gorm:"foreignKey:BonusInventoryID" json:"bonus_inventory,omitempty"`
}

func (s *SalesBonusRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func (s *SalesBonusStock) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
