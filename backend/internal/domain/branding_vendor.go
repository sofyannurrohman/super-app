package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BrandingVendor struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MySQLVendorID *int64    `gorm:"column:mysql_vendor_id;index" json:"mysql_vendor_id,omitempty"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Email         string    `gorm:"type:varchar(255)" json:"email"`
	Phone         string    `gorm:"type:varchar(50)" json:"phone"`
	Address       string    `gorm:"type:text" json:"address"`
	PricePerMeter float64   `gorm:"type:numeric(15,2);default:0" json:"price_per_meter"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time `gorm:"default:now()" json:"updated_at"`
}

func (v *BrandingVendor) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return
}
