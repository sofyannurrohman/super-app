package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Company struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Code      string    `gorm:"-:migration;type:varchar(20);uniqueIndex" json:"code"`
	TaxNumber string    `gorm:"type:varchar(50)" json:"tax_number"`
	Address   string    `gorm:"type:text" json:"address"`
	LogoURL   string    `gorm:"type:text" json:"logo_url"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

func (c *Company) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}

type Branch struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID         *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	Name              string     `gorm:"type:varchar(100);not null" json:"name"`
	Address           string     `gorm:"type:text" json:"address"`
	Timezone          string     `gorm:"type:varchar(50);default:'Asia/Jakarta'" json:"timezone"`
	Latitude          float64    `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude         float64    `gorm:"type:decimal(11,8)" json:"longitude"`
	RadiusMeter       int        `gorm:"default:100" json:"radius_meter"`
	DefaultShiftID    *uuid.UUID `gorm:"type:uuid" json:"default_shift_id"`
	ManualCashBalance float64    `gorm:"type:decimal(15,2);default:0" json:"manual_cash_balance"`
	PettyCashBalance  float64    `gorm:"type:decimal(15,2);default:0" json:"petty_cash_balance"`
	CreatedAt         time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"default:now()" json:"updated_at"`

	Company      *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	DefaultShift *Shift   `gorm:"foreignKey:DefaultShiftID" json:"default_shift,omitempty"`
}

func (b *Branch) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}
