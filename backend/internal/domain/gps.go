package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GpsDevice struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id" form:"id"`
	VehicleID uuid.UUID `gorm:"type:uuid;not null" json:"vehicle_id" form:"vehicle_id"`
	IMEI      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"imei" form:"imei"`
	SimNumber string    `gorm:"type:varchar(20)" json:"sim_number" form:"sim_number"`
	Status    string    `gorm:"type:varchar(20);default:'ACTIVE'" json:"status" form:"status"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Vehicle *Vehicle `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
}

func (g *GpsDevice) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return
}

type VehicleLocation struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VehicleID uuid.UUID `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	Latitude  float64   `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude float64   `gorm:"type:decimal(11,8)" json:"longitude"`
	Speed     float64   `gorm:"type:decimal(8,2)" json:"speed"`
	Course    float64   `gorm:"type:decimal(5,2)" json:"course"`
	IsAccOn   bool      `gorm:"type:boolean;default:false" json:"is_acc_on"`
	Timestamp time.Time `gorm:"not null;index" json:"timestamp"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`

	Vehicle *Vehicle `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
}

func (v *VehicleLocation) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return
}
