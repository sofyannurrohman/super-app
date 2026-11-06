package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleComponent struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id" form:"id"`
	VehicleID      uuid.UUID `gorm:"type:uuid;not null" json:"vehicle_id" form:"vehicle_id"`
	Name           string    `gorm:"type:varchar(255);not null" json:"name" form:"name"`
	MaxMileage     int       `gorm:"type:int;not null" json:"max_mileage" form:"max_mileage"`
	CurrentMileage int       `gorm:"type:int;default:0" json:"current_mileage" form:"current_mileage"`
	IsActive       bool      `gorm:"default:true" json:"is_active" form:"is_active"`
	CreatedAt      time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"default:now()" json:"updated_at"`

	Vehicle *Vehicle `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
}

type VehicleMaintenance struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id" form:"id"`
	VehicleID          uuid.UUID `gorm:"type:uuid;not null" json:"vehicle_id" form:"vehicle_id"`
	VehicleComponentID uuid.UUID `gorm:"type:uuid;not null" json:"vehicle_component_id" form:"vehicle_component_id"`
	MaintenanceDate    time.Time `gorm:"type:date;not null" json:"maintenance_date" form:"maintenance_date"`
	VehicleMileage     int       `gorm:"type:int;not null" json:"vehicle_mileage" form:"vehicle_mileage"`
	Notes              string    `gorm:"type:text" json:"notes" form:"notes"`
	CreatedAt          time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt          time.Time `gorm:"default:now()" json:"updated_at"`

	Vehicle          *Vehicle          `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	VehicleComponent *VehicleComponent `gorm:"foreignKey:VehicleComponentID" json:"vehicle_component,omitempty"`
}

func (vc *VehicleComponent) BeforeCreate(tx *gorm.DB) (err error) {
	if vc.ID == uuid.Nil {
		vc.ID = uuid.New()
	}
	return
}

func (vm *VehicleMaintenance) BeforeCreate(tx *gorm.DB) (err error) {
	if vm.ID == uuid.Nil {
		vm.ID = uuid.New()
	}
	return
}
